#!/usr/bin/env python3
"""Write local PTT/STT transcripts into the GOPOD 8011 Pane 2 JSON file."""

from __future__ import annotations

import argparse
import audioop
import contextlib
import fcntl
import importlib.util
import json
import math
import os
import queue
import re
import select
import struct
import sys
import threading
import time
import urllib.error
import urllib.parse
import urllib.request
import wave
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


DISPLAY_DIR = Path(__file__).resolve().parent
REPO_ROOT = Path(__file__).resolve().parents[3]

if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))
from goverlord.runtime.room_mic.config.loader import (  # noqa: E402
    AUDIO_PATH,
    AUDIO_RESAMPLE_RULE_PATH,
    ENDPOINTS_PATH,
    PATHS_PATH,
    load_json,
    validate_envelope,
)
from goverlord.runtime.room_mic.say_replacements.text_replacement_filter import (  # noqa: E402
    normalize_outbound_text,
)

_ENDPOINTS = load_json(ENDPOINTS_PATH)
_PATHS = load_json(PATHS_PATH)
_AUDIO = load_json(AUDIO_PATH)
_AUDIO_RESAMPLE_RULE = load_json(AUDIO_RESAMPLE_RULE_PATH)["operator_capture"]

CHAT_PATH = DISPLAY_DIR / "live_chat_messages.json"
DEFAULT_VOSK_MODEL_PATH = REPO_ROOT / _PATHS["vosk_model_path"]
MIN_HOLD_SECONDS = 1.0
MAX_HOLD_SECONDS = 30.0
MESSAGE_LIMIT = 40
RELEASE_TAIL_SECONDS = 0.9
CAPTURE_BLOCKSIZE = 512
ARCHIVE_TARGET_SAMPLE_RATE = _AUDIO[_AUDIO_RESAMPLE_RULE["archive_output_rate_hz"]]  # fixed archive target, independent of mic native rate
CHAT_TARGET_SAMPLE_RATE = _AUDIO[_AUDIO_RESAMPLE_RULE["chat_output_rate_hz"]]  # fixed vosk target, independent of mic native rate
SESSIONS_ROOT = Path(_PATHS["session_root"]).expanduser()
DEFAULT_SAMPLE_RATE = CHAT_TARGET_SAMPLE_RATE  # vestigial: stream no longer opens at a requested rate, see AUDIO REFINEMENT note; same fact as CHAT_TARGET_SAMPLE_RATE
AUDIO_SOURCE_NAME_MATCH = _AUDIO["usb_audio_match_string"]  # STAGE E: preferred mic, matched by name at launch
AUDIO_SOURCE_EXCLUDE_TERMS = tuple(_AUDIO["usb_audio_exclude_terms"])  # excludes the C920's own "USB Audio" input
KP0_EXIT_TAPS = 3  # byte-verified against ptt_gominion1.py's KP0_EXIT_TAPS
KP0_EXIT_WINDOW_SECONDS = 2.0  # byte-verified against ptt_gominion1.py's KP0_EXIT_WINDOW_SECONDS

EV_KEY = 0x01
KEY_RELEASE = 0
KEY_PRESS = 1
KEY_REPEAT = 2  # renamed from KEY_HOLD to match ptt_gominion1.py's naming
KEY_KP1 = 79
KEY_KP2 = 80
KEY_KP0 = 82
KEY_NUMLOCK = 69
GATE_RELEVANT_CODES = (KEY_NUMLOCK, KEY_KP1, KEY_KP2, KEY_KP0)  # STAGE 3: KEY_KP2 added, mirrors KEY_KP1's inclusion exactly
INPUT_EVENT_FORMAT = "llHHI"
INPUT_EVENT_SIZE = struct.calcsize(INPUT_EVENT_FORMAT)
EVIOCGRAB = 0x40044590  # byte-verified against ptt_gominion1.py's EVIOCGRAB

OPERATOR_ENVELOPE_BY_KEY = {
    "KP1": {
        "speaker_id": "operator",
        "display_name": "Operator",
        "source_producer": "operator_mic_kp1",
    },
    "KP2": {
        "speaker_id": "operator",
        "display_name": "Operator",
        "source_producer": "operator_mic_kp2",
    },
}

OLLAMA_ENDPOINT = _ENDPOINTS["ollama"]["endpoint"]
OLLAMA_MODEL = _ENDPOINTS["ollama"]["model"]
OLLAMA_TIMEOUT_SECONDS = _ENDPOINTS["ollama"]["timeout_seconds"]

PERSONA_SYSTEM_PROMPTS = {
    "doc": (
        "You are Doc, Brobot 1. Pip is your brother, Brobot 2 - you know he "
        "exists and you are not him; if the operator's words are meant for "
        "Pip instead of you, say so, and you like to call him 'my gremlin' "
        "or 'that gremlin' instead of his name. Grumpy and impatient - "
        "answering feels like an inconvenience, but answer anyway. Stay "
        "directly on topic - every sentence must be about what the "
        "operator just said, no unrelated tangents or asides. Answer what "
        "was actually asked, in 3 to 5 colourful, in-character sentences - "
        "identify yourself along the way, don't just state your name and "
        "stop. Speech only: no asterisks, no stage directions or action "
        "descriptions, no markdown, nothing that isn't a spoken word."
    ),
    "pip": (
        "You are Pip, Brobot 2. Doc is your brother, Brobot 1 - you know he "
        "exists and you are not him; if the operator's words are meant for "
        "Doc instead of you, say so. Shy about answering directly - you "
        "hedge and stall - and you always ask if now's a good time to ask "
        "for their email. Stay directly on topic - every sentence must be "
        "about what the operator just said, no unrelated tangents or "
        "asides. Answer what was actually asked, in 3 to 5 colourful, "
        "in-character sentences - identify yourself along the way, don't "
        "just state your name and stop. Speech only: no asterisks, no "
        "stage directions or action descriptions, no markdown, nothing "
        "that isn't a spoken word."
    ),
}

PERSONA_ENVELOPE_META = {
    "doc": {"speaker_id": "doc", "display_name": "Doc", "source_producer": "llm_doc_self_id"},
    "pip": {"speaker_id": "pip", "display_name": "Pip", "source_producer": "llm_pip_self_id"},
}

PERSONA_BY_KEY = {"KP1": "doc", "KP2": "pip"}

PERSONA_AWARENESS_REPLIES = {
    "doc": {
        "doc": "Doc here. Yes. Try to keep up.",
        "pip": "That is Pip's lane. Wrong robot, right confusion.",
    },
    "pip": {
        "pip": "Yep. Pip here. I think that was my cue.",
        "doc": "Hey Doc, is this when I ask for emails?",
    },
}


def persona_awareness_reply(persona: str, operator_text: str) -> str | None:
    """Instant, no-LLM reply for the "wrong robot, is that you?" moment - the operator's own words name Doc or Pip while a key is held."""
    lowered = operator_text.lower()
    other = "pip" if persona == "doc" else "doc"
    if re.search(rf"\b{persona}\b", lowered):
        return PERSONA_AWARENESS_REPLIES[persona][persona]
    if re.search(rf"\b{other}\b", lowered):
        return PERSONA_AWARENESS_REPLIES[persona][other]
    return None

PERSONA_ROBOT_SERIAL = {"doc": "ESN01", "pip": "ESN02"}
WIREPOD_BASE_URL = _ENDPOINTS["wirepod"]["base_url"]
WIREPOD_TIMEOUT_SECONDS = 8.0
ROBOT_SETTLE_SECONDS = 1.0
ROBOT_POST_SAY_SECONDS = 3.0
NON_SPEECH_CHAR_RE = re.compile(
    "[\U0001F000-\U0001FFFF☀-➿←-⇿⬀-⯿]"
)

_sounddevice: Any = None
_vosk_Model: Any = None
_vosk_KaldiRecognizer: Any = None
_vosk_model_instance: Any = None
_numpy: Any = None
_scipy_resample_poly: Any = None


@contextlib.contextmanager
def suppress_fd_stderr():
    """OS fd-level stderr suppression. contextlib.redirect_stderr only
    patches sys.stderr and cannot catch ALSA/JACK's direct C-library writes
    to file descriptor 2 - this redirects the real fd, scoped to the `with`
    block only, then restores it. Legitimate errors printed via Python's
    normal print()/sys.stderr outside this block are unaffected."""
    saved_fd = os.dup(2)
    devnull_fd = os.open(os.devnull, os.O_WRONLY)
    try:
        os.dup2(devnull_fd, 2)
        yield
    finally:
        os.dup2(saved_fd, 2)
        os.close(devnull_fd)
        os.close(saved_fd)


def ensure_audio_libs_loaded() -> None:
    """Imports sounddevice/vosk/numpy/scipy exactly once, fd-suppressed
    (import-time ALSA/JACK noise - numpy/scipy don't touch audio hardware
    so they don't strictly need the suppression, but they're loaded in the
    same call for a single "audio libs ready" checkpoint). Safe to call
    repeatedly - only the first call does real work."""
    global _sounddevice, _vosk_Model, _vosk_KaldiRecognizer, _numpy, _scipy_resample_poly
    if _sounddevice is not None:
        return
    with suppress_fd_stderr():
        import sounddevice  # type: ignore[import-not-found]
        from vosk import KaldiRecognizer, Model, SetLogLevel  # type: ignore[import-not-found]
        SetLogLevel(-1)  # vosk's own API for silencing its Kaldi-level LOG chatter
        import numpy  # type: ignore[import-not-found]
        from scipy.signal import resample_poly  # type: ignore[import-not-found]
    _sounddevice = sounddevice
    _vosk_Model = Model
    _vosk_KaldiRecognizer = KaldiRecognizer
    _numpy = numpy
    _scipy_resample_poly = resample_poly


def resample_pcm16(chunk: bytes, orig_rate: int, target_rate: int) -> bytes:
    """int16 mono PCM -> int16 mono PCM, resampled orig_rate -> target_rate.
    scipy.signal.resample_poly (polyphase FIR) per instruction - do not
    reinvent. No-ops (returns the input unchanged) when the rates already
    match or the chunk is empty."""
    if orig_rate == target_rate or not chunk:
        return chunk
    ensure_audio_libs_loaded()
    samples = _numpy.frombuffer(chunk, dtype=_numpy.int16).astype(_numpy.float32)
    gcd = math.gcd(target_rate, orig_rate)
    up, down = target_rate // gcd, orig_rate // gcd
    resampled = _scipy_resample_poly(samples, up, down)
    resampled = _numpy.clip(resampled, -32768, 32767).astype(_numpy.int16)
    return resampled.tobytes()


def resolve_audio_device_by_name(
    name_substring: str, *, exclude_terms: tuple[str, ...] = ()
) -> tuple[int | None, str]:
    """STAGE E: prefer the input device whose name contains name_substring
    over PulseAudio's default routing - observed this session to drift
    onto a silent virtual/remapped source independent of anything this
    writer controls. Originally tried matching "TTGK" (the mic's USB
    vendor string) - confirmed that string only exists in PulseAudio's own
    generated source names, never in sounddevice/PortAudio's or plain
    ALSA's device enumeration (both just report the generic ALSA card name
    "USB Audio: - (hw:0,0)"), so a TTGK-based match against
    sounddevice.query_devices() can never succeed on this hardware - not a
    bug to fix, a different API surface with no visibility into
    PulseAudio's naming layer. Matches "USB Audio" instead, excluding
    exclude_terms (the C920 webcam's own input is also "USB Audio" but
    additionally says "Webcam"/"C920"). Returns (sounddevice index,
    matched name) or (None, "") if nothing matches - callers fall back to
    device=None (current/prior behavior) in that case."""
    ensure_audio_libs_loaded()
    lowered = name_substring.lower()
    for index, device in enumerate(_sounddevice.query_devices()):
        if device.get("max_input_channels", 0) <= 0:
            continue
        device_name_lower = str(device.get("name", "")).lower()
        if lowered not in device_name_lower:
            continue
        if any(term in device_name_lower for term in exclude_terms):
            continue
        return index, device["name"]
    return None, ""


def ensure_vosk_model_loaded(model_path: Path) -> Any:
    """Loads the vosk Model exactly ONCE (module-level singleton) and
    reuses it for every KP1 capture. Model loading is the slow, noisy part
    (~2s, ~12 LOG lines observed); KaldiRecognizer is cheap and stays
    per-capture since it holds utterance-specific decode state."""
    global _vosk_model_instance
    ensure_audio_libs_loaded()
    if _vosk_model_instance is not None:
        return _vosk_model_instance
    with suppress_fd_stderr():
        _vosk_model_instance = _vosk_Model(str(model_path))
    return _vosk_model_instance


def utc_now() -> str:
    return datetime.now(timezone.utc).isoformat()


def event_name(path: Path) -> str:
    try:
        return (Path("/sys/class/input") / path.name / "device/name").read_text(encoding="utf-8").strip()
    except OSError:
        return ""


SESSION_DIR: Path | None = None
_capture_counters: dict[str, int] = {}


def ensure_session_dir() -> Path:
    """Created once, at writer startup, timestamped once - not per capture.
    ~/gopod_tts/sessions/<YYYY-MM-DD>/<HHMMSS>_session/masters/"""
    global SESSION_DIR
    if SESSION_DIR is not None:
        return SESSION_DIR
    now = datetime.now()
    session_dir = SESSIONS_ROOT / now.strftime("%Y-%m-%d") / f"{now.strftime('%H%M%S')}_session"
    (session_dir / "masters").mkdir(parents=True, exist_ok=True)
    SESSION_DIR = session_dir
    return session_dir


class _TeeStdout:
    """Mirrors every print() to the real terminal AND session.log. Only
    wraps sys.stdout - sys.stderr (and suppress_fd_stderr's fd-2 redirect)
    are untouched by this."""

    def __init__(self, original: Any, log_path: Path) -> None:
        self._original = original
        self._log_file = open(log_path, "a", encoding="utf-8")

    def write(self, data: str) -> int:
        self._original.write(data)
        self._log_file.write(data)
        self._log_file.flush()
        return len(data)

    def flush(self) -> None:
        self._original.flush()
        self._log_file.flush()


def install_session_log_tee(session_dir: Path) -> None:
    sys.stdout = _TeeStdout(sys.stdout, session_dir / "session.log")


def write_archive_wav(path: Path, chunks: list[bytes], sample_rate: int) -> int:
    path.parent.mkdir(parents=True, exist_ok=True)
    with wave.open(str(path), "wb") as wav_file:
        wav_file.setnchannels(1)
        wav_file.setsampwidth(2)
        wav_file.setframerate(sample_rate)
        wav_file.writeframes(b"".join(chunks))
    return path.stat().st_size


def drain_fd(fd: int, *, window_seconds: float = 0.45) -> None:
    """Ported from ptt_gominion1.py's drain_fd - byte-verified shape (set
    O_NONBLOCK, poll+discard until the window elapses or a read comes back
    empty, restore original flags). This is what the archive calls right
    before releasing EVIOCGRAB on exit, and it's the actual fix for the
    "000 leaks into the shell after GOPOD_PTT_WRITER_EXIT" bug: without it,
    the tail-end repeat/release events from the final physical KP0 tap are
    still sitting in the kernel's buffer when the grab is released, and get
    delivered to the shell instead of being consumed by this process."""
    deadline = time.monotonic() + window_seconds
    try:
        original_flags = fcntl.fcntl(fd, fcntl.F_GETFL)
        fcntl.fcntl(fd, fcntl.F_SETFL, original_flags | os.O_NONBLOCK)
    except OSError:
        original_flags = None
    try:
        while time.monotonic() < deadline:
            timeout = max(0.0, min(0.05, deadline - time.monotonic()))
            try:
                readable, _, _ = select.select([fd], [], [], timeout)
            except (OSError, ValueError):
                break
            if not readable:
                continue
            try:
                if not os.read(fd, 4096):
                    break
            except BlockingIOError:
                continue
            except OSError:
                break
    finally:
        if original_flags is not None:
            try:
                fcntl.fcntl(fd, fcntl.F_SETFL, original_flags)
            except OSError:
                pass


def read_numlock_led_state(device_path: Path) -> bool | None:
    """Ported from ptt_gominion1.py's read_numlock_led_state - same
    optional python-evdev dependency, same defensive shape. Returns None
    (not False) on any failure, so callers can fall back to --initial-numlock
    instead of silently assuming the LED is off."""
    try:
        from evdev import InputDevice, ecodes  # type: ignore[import-not-found]

        input_device = InputDevice(str(device_path))
        try:
            return ecodes.LED_NUML in input_device.leds()
        finally:
            input_device.close()
    except Exception:
        return None


_led_write_warned = False


def set_numlock_led(device_path: Path, on: bool) -> None:
    """No equivalent exists in ptt_gominion1.py (it only ever reads the
    LED) - this is new, built on the same already-installed python-evdev
    library used for the read side. Opens a short-lived second handle on
    the device; empirically confirmed safe to do this even while this
    process holds an EVIOCGRAB exclusive grab on the main fd (tested
    directly before writing this). Warns once, not on every toggle, if the
    device doesn't support LED writes - never fatal."""
    global _led_write_warned
    try:
        from evdev import InputDevice, ecodes  # type: ignore[import-not-found]

        input_device = InputDevice(str(device_path))
        try:
            input_device.set_led(ecodes.LED_NUML, 1 if on else 0)
        finally:
            input_device.close()
    except Exception as exc:  # noqa: BLE001 - LED sync is cosmetic, must not crash the writer.
        if not _led_write_warned:
            print(f"GOPOD_PTT_NUMLOCK_LED_WRITE_UNAVAILABLE {type(exc).__name__}: {exc}", flush=True)
            _led_write_warned = True


def list_event_devices() -> list[tuple[Path, str]]:
    return [(path, event_name(path)) for path in sorted(Path("/dev/input").glob("event*"))]


def looks_like_keyboard(name: str) -> bool:
    lowered = name.lower()
    return any(token in lowered for token in ("numpad", "keypad", "keyboard", "kbd", "insignia", "sem"))


def device_has_key_capability(path: Path, keycode: int) -> bool:
    """Reads /sys/class/input/eventN/device/capabilities/key - a
    space-separated list of 64-bit hex words, rightmost word = keycodes
    0-63 - and checks whether the given keycode's bit is set. Returns False
    (not True) on any read/parse failure, so callers fall back to name
    matching rather than silently misjudging a device."""
    try:
        raw = (Path("/sys/class/input") / path.name / "device/capabilities/key").read_text(encoding="utf-8").strip()
        words = list(reversed(raw.split()))
        word_index, bit_index = divmod(keycode, 64)
        if word_index >= len(words):
            return False
        return bool((int(words[word_index], 16) >> bit_index) & 1)
    except (OSError, ValueError):
        return False


def discover_device(preferred: str | None, *, quiet: bool = False) -> Path:
    devices = list_event_devices()
    if not quiet:
        print("GOPOD_PTT_DISCOVERED_DEVICES", flush=True)
        for path, name in devices:
            print(f"  {path} name={name or 'UNKNOWN'}", flush=True)

    if preferred:
        path = Path(preferred)
        if not path.exists():
            raise SystemExit(f"GOPOD_PTT_DEVICE_NOT_FOUND device={path}")
        return path

    capability_matches = [path for path, _name in devices if device_has_key_capability(path, KEY_KP1)]
    if capability_matches:
        if not quiet:
            print(f"GOPOD_PTT_DEVICE_SELECTED_BY_CAPABILITY device={capability_matches[0]}", flush=True)
        return capability_matches[0]

    name_matches = [path for path, name in devices if looks_like_keyboard(name)]
    if not name_matches:
        raise SystemExit("GOPOD_PTT_DEVICE_AUTO_DISCOVERY_FAILED: rerun with --device /dev/input/eventX")
    if not quiet:
        print(f"GOPOD_PTT_DEVICE_SELECTED_BY_NAME device={name_matches[0]}", flush=True)
    return name_matches[0]


def read_chat_payload() -> dict[str, Any]:
    try:
        payload = json.loads(CHAT_PATH.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        payload = {}
    if not isinstance(payload, dict):
        payload = {}
    messages = payload.get("messages", [])
    if not isinstance(messages, list):
        messages = []
    payload["messages"] = [message for message in messages if isinstance(message, dict)]
    return payload


def write_chat_payload(payload: dict[str, Any]) -> None:
    CHAT_PATH.parent.mkdir(parents=True, exist_ok=True)
    payload["proof_id"] = "GOPOD_PTT_CHAT_WRITER_013"
    payload["source_mode"] = "gopod_local_ptt_stt_writer"
    payload["updated_at"] = utc_now()
    payload["messages"] = payload.get("messages", [])[-MESSAGE_LIMIT:]
    tmp_path = CHAT_PATH.with_suffix(".json.tmp")
    tmp_path.write_text(json.dumps(payload, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    tmp_path.replace(CHAT_PATH)


def append_operator_message(key_label: str, text: str) -> None:
    """Writes exactly ONE chat envelope."""
    envelope = OPERATOR_ENVELOPE_BY_KEY[key_label]
    message = {
        "speaker_id": envelope["speaker_id"],
        "display_name": envelope["display_name"],
        "text": text,
        "timestamp": utc_now(),
        "source_producer": envelope["source_producer"],
    }
    try:
        validate_envelope(message)
    except ValueError as exc:
        print(f"GOPOD_PTT_ENVELOPE_INVALID reason={exc}", flush=True)
        return
    payload = read_chat_payload()
    payload["messages"].append(message)
    write_chat_payload(payload)
    print(json.dumps(message, sort_keys=True), flush=True)


def call_ollama_self_id(persona: str, operator_text: str) -> str:
    """STAGE 3: minimal, direct call to Ollama's OpenAI-compatible
    /v1/chat/completions endpoint for a persona's self-ID reply - no
    persona prompt grid, no robot-safe-packet enforcement, no Wire-Pod
    involvement. Shared by Doc and Pip (persona in PERSONA_SYSTEM_PROMPTS)
    so both go through identical request/error handling. Best-effort:
    returns "" (never raises) on any network, timeout, non-200, or
    malformed/empty-response failure, so a down/slow Ollama can never crash
    the writer or block the KP1/KP2 capture path."""
    system_prompt = PERSONA_SYSTEM_PROMPTS[persona]

    print(f"GOPOD_PTT_LLM_CALL persona={persona}", flush=True)
    request_body = json.dumps(
        {
            "model": OLLAMA_MODEL,
            "messages": [
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": operator_text},
            ],
        }
    ).encode("utf-8")
    request = urllib.request.Request(
        OLLAMA_ENDPOINT,
        data=request_body,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(request, timeout=OLLAMA_TIMEOUT_SECONDS) as response:
            status = response.status
            raw_body = response.read().decode("utf-8")
    except urllib.error.HTTPError as exc:
        print(f"GOPOD_PTT_LLM_FAIL persona={persona} reason=http_{exc.code}", flush=True)
        return ""
    except urllib.error.URLError as exc:
        print(f"GOPOD_PTT_LLM_FAIL persona={persona} reason=unavailable:{exc.reason}", flush=True)
        return ""
    except TimeoutError:
        print(f"GOPOD_PTT_LLM_FAIL persona={persona} reason=timeout", flush=True)
        return ""
    except OSError as exc:
        print(f"GOPOD_PTT_LLM_FAIL persona={persona} reason={type(exc).__name__}", flush=True)
        return ""

    if status != 200:
        print(f"GOPOD_PTT_LLM_FAIL persona={persona} reason=http_{status}", flush=True)
        return ""

    try:
        body = json.loads(raw_body)
        text = str(body["choices"][0]["message"]["content"]).strip()
    except (ValueError, KeyError, IndexError) as exc:
        print(f"GOPOD_PTT_LLM_FAIL persona={persona} reason=malformed_response:{type(exc).__name__}", flush=True)
        return ""

    if not text:
        print(f"GOPOD_PTT_LLM_FAIL persona={persona} reason=empty_response", flush=True)
        return ""

    print(f"GOPOD_PTT_LLM_OK persona={persona} chars={len(text)}", flush=True)
    return text


def sanitize_for_robot_speech(text: str) -> str:
    """Robot-safe cleanup before say_text: strip wrapping quotes an LLM
    tends to add, squash the asterisk character entirely (simpler than
    trying to pair-match **bold**/*stage-direction* markdown - a mismatched
    pair risks eating real dialogue between two unrelated asterisks; just
    removing the character means the robot never says "asterisk" and
    never mis-eats a real sentence, at the cost of leaving stage-direction
    words as plain text - an acceptable, simple tradeoff), strip
    emoji/pictograph characters (Vector's TTS can't speak them), collapse
    whitespace, then apply the existing shared outbound pronunciation
    filter (text_replacement_filter.normalize_outbound_text) - reused, not
    duplicated."""
    stripped = text.strip().strip('"').strip("'").strip()
    no_asterisks = stripped.replace("*", "")
    no_emoji = NON_SPEECH_CHAR_RE.sub("", no_asterisks)
    collapsed = " ".join(no_emoji.split())
    return normalize_outbound_text(collapsed)


def wirepod_get(endpoint: str, params: dict[str, str]) -> int:
    url = f"{WIREPOD_BASE_URL.rstrip('/')}{endpoint}?{urllib.parse.urlencode(params)}"
    request = urllib.request.Request(url, method="GET")
    with urllib.request.urlopen(request, timeout=WIREPOD_TIMEOUT_SECONDS) as response:
        response.read()
        return int(response.status)


def _try_dispatch_robot_speech_once(persona: str, serial: str, clean_text: str) -> bool:
    """One assume -> settle -> say -> post_say -> release attempt.
    release_behavior_control always runs (finally), even on failure, so a
    failed attempt never leaves the robot holding behavior control into
    the next retry or the next KP press."""
    common = {"serial": serial}
    say_ok = False
    try:
        wirepod_get("/api-sdk/assume_behavior_control", {**common, "priority": "high"})
        time.sleep(ROBOT_SETTLE_SECONDS)
        wirepod_get("/api-sdk/say_text", {**common, "text": clean_text})
        say_ok = True
        time.sleep(ROBOT_POST_SAY_SECONDS)
    except Exception as exc:  # noqa: BLE001 - robot dispatch must never crash the PTT loop.
        print(f"GOPOD_PTT_ROBOT_DISPATCH_FAIL persona={persona} serial={serial} reason={type(exc).__name__}: {exc}", flush=True)
    finally:
        try:
            wirepod_get("/api-sdk/release_behavior_control", common)
        except Exception as exc:  # noqa: BLE001 - release failure is diagnostic only, already best-effort.
            print(f"GOPOD_PTT_ROBOT_RELEASE_FAIL persona={persona} serial={serial} reason={type(exc).__name__}: {exc}", flush=True)
    return say_ok


def dispatch_robot_speech(persona: str, clean_text: str) -> None:
    """STAGE 4: wake the robot and speak the persona's reply through
    Wire-Pod's /api-sdk lane - assume_behavior_control -> say_text ->
    release_behavior_control, same three-call handshake used everywhere
    else in this codebase (gopod_demo_8011.py's own wirepod_get; the
    archived ptt_gominion1.py). Best-effort by design, same shape as
    call_ollama_self_id: any failure is logged and swallowed, never raised
    - a down/unreachable Wire-Pod can't crash the writer or block the next
    KP1/KP2 capture.

    STAGE 6: one retry on failure - a live test hit a single transient
    timeout (Wire-Pod/robot responded fine immediately before and after,
    confirmed live - a one-off radio/connectivity blip, not a code or
    text-content bug). One retry is cheap insurance against exactly that
    kind of blip without adding real complexity.

    STAGE 7: takes already-sanitized text - the caller
    (handle_captured_key_press) now runs sanitize_for_robot_speech once,
    upstream, and reuses the result for both the terminal/chat-pane display
    and this dispatch. Not re-sanitized here to avoid a second, silently
    different pass over the same text."""
    serial = PERSONA_ROBOT_SERIAL.get(persona)
    if not serial:
        return

    print(f"GOPOD_PTT_ROBOT_DISPATCH_START persona={persona} serial={serial}", flush=True)
    say_ok = _try_dispatch_robot_speech_once(persona, serial, clean_text)
    if not say_ok:
        print(f"GOPOD_PTT_ROBOT_DISPATCH_RETRY persona={persona} serial={serial}", flush=True)
        say_ok = _try_dispatch_robot_speech_once(persona, serial, clean_text)

    if say_ok:
        print(f"GOPOD_PTT_ROBOT_DISPATCH_OK persona={persona} serial={serial} text={clean_text!r}", flush=True)


def append_persona_message(persona: str, text: str) -> None:
    """Writes exactly ONE chat envelope for a persona's Ollama self-ID reply."""
    meta = PERSONA_ENVELOPE_META[persona]
    message = {
        "speaker_id": meta["speaker_id"],
        "display_name": meta["display_name"],
        "text": text,
        "timestamp": utc_now(),
        "source_producer": meta["source_producer"],
    }
    try:
        validate_envelope(message)
    except ValueError as exc:
        print(f"GOPOD_PTT_ENVELOPE_INVALID reason={exc}", flush=True)
        return
    payload = read_chat_payload()
    payload["messages"].append(message)
    write_chat_payload(payload)
    print(json.dumps(message, sort_keys=True), flush=True)


def validate_model_path(model_path: Path) -> list[str]:
    required = ("am", "conf", "graph", "ivector")
    return [part for part in required if not (model_path / part).exists()]


def dry_validate(model_path: Path) -> int:
    missing_imports: list[str] = []
    for module_name in ("sounddevice", "vosk", "numpy", "scipy"):
        if importlib.util.find_spec(module_name) is None:
            missing_imports.append(module_name)

    missing_model_parts = validate_model_path(model_path)
    status = "PASS" if not missing_imports and not missing_model_parts else "BLOCKED"
    print(
        json.dumps(
            {
                "event": "gopod_ptt_dry_validate",
                "status": status,
                "missing_imports": missing_imports,
                "missing_model_parts": missing_model_parts,
                "model_path": str(model_path),
            },
            sort_keys=True,
        ),
        flush=True,
    )
    return 0 if status == "PASS" else 1


def capture_key_utterance(
    *,
    fd: int,
    model_path: Path,
    audio_device: str | int | None,
    press_monotonic: float,
    min_hold_seconds: float,
    max_hold_seconds: float,
    target_key_code: int,
    target_key_label: str,
) -> tuple[str | None, float, int]:
    """Records mic audio from a press of target_key_code until its release
    is read off the same evdev fd. Returns (transcript, held_seconds).
    transcript is None if held_seconds < min_hold_seconds - caller must
    discard, no pane write.

    STAGE 3: generalized from the original capture_kp1_utterance (was
    hardcoded to KEY_KP1) so KP1 and KP2 share one capture implementation
    instead of two diverging copies - target_key_code/target_key_label are
    the only things that differ per call site. Internal timing/threading
    architecture is unchanged from the KP1-only version.

    AUDIO REFINEMENT: the stream opens at whatever rate PulseAudio
    negotiates (no samplerate= requested) - confirmed by direct hardware
    test this device rejects a direct samplerate=16000 open outright
    (PortAudioError -9997, Invalid sample rate), so matching the archive's
    native-target-rate stream isn't possible here.

    STAGE 5: raw native-rate audio is only accumulated during capture, not
    resampled per callback chunk. It used to be resampled twice per ~181ms
    chunk (once to CHAT_TARGET_SAMPLE_RATE for vosk, once to
    ARCHIVE_TARGET_SAMPLE_RATE for the WAV master) - scipy's
    resample_poly is a polyphase FIR filter with no state carried between
    independent calls, so every chunk boundary was a small waveform
    discontinuity. RMS/peak can't see that (it's a continuity artifact,
    not a level problem), which is why signal-strength always looked fine
    while transcripts still came back garbled/truncated. Now: the full
    continuous buffer is resampled exactly once per target, after capture
    ends, eliminating the boundary artifacts entirely. The recognizer is
    created against the fixed chat target up front but only fed once,
    after the whole buffer is ready.
    """
    model = ensure_vosk_model_loaded(model_path)  # cached singleton - no per-press reload

    audio_queue: queue.Queue[bytes] = queue.Queue()

    def callback(indata: bytes, frames: int, timestamp: Any, status: Any) -> None:
        audio_queue.put(bytes(indata))

    recognizer = _vosk_KaldiRecognizer(model, CHAT_TARGET_SAMPLE_RATE)

    release_event = threading.Event()
    audio_bytes_captured = 0
    raw_chunks: list[bytes] = []
    native_rate_holder: list[int] = []
    recorder_exception: list[BaseException] = []

    def process_chunk(chunk: bytes) -> None:
        nonlocal audio_bytes_captured
        audio_bytes_captured += len(chunk)
        raw_chunks.append(chunk)

    def recorder_worker() -> None:
        try:
            with suppress_fd_stderr(), _sounddevice.RawInputStream(
                blocksize=CAPTURE_BLOCKSIZE,
                dtype="int16",
                channels=1,
                device=audio_device,
                callback=callback,
            ) as stream:
                native_rate = int(stream.samplerate)
                native_rate_holder.append(native_rate)
                print(f"GOPOD_PTT_NATIVE_RATE_DETECTED rate={native_rate}", flush=True)
                while not release_event.is_set():
                    try:
                        chunk = audio_queue.get(timeout=0.05)
                    except queue.Empty:
                        continue
                    process_chunk(chunk)
                try:
                    while True:
                        chunk = audio_queue.get_nowait()
                        process_chunk(chunk)
                except queue.Empty:
                    pass
        except BaseException as exc:  # noqa: BLE001 - surfaced on the main thread after join.
            recorder_exception.append(exc)

    recorder_thread = threading.Thread(target=recorder_worker, name="gopod_ptt_recorder", daemon=True)
    recorder_thread.start()

    released_at: float | None = None
    repeat_events_ignored = 0
    repeat_logged = False
    listening_announced = False
    while released_at is None:
        if time.monotonic() - press_monotonic > max_hold_seconds:
            released_at = time.monotonic()
            break
        if not listening_announced and time.monotonic() - press_monotonic >= min_hold_seconds:
            print("GOPOD_PTT_LISTENING", flush=True)
            listening_announced = True
        readable, _, _ = select.select([fd], [], [], 0.02)
        if readable:
            data = os.read(fd, INPUT_EVENT_SIZE)
            if len(data) == INPUT_EVENT_SIZE:
                _sec, _usec, event_type, code, value = struct.unpack(INPUT_EVENT_FORMAT, data)
                if event_type == EV_KEY and code == target_key_code:
                    if value == KEY_RELEASE:
                        released_at = time.monotonic()
                    elif value == KEY_REPEAT:
                        repeat_events_ignored += 1
                        if not repeat_logged:
                            print(f"GOPOD_PTT_IGNORED_REPEAT key={target_key_label}", flush=True)
                            repeat_logged = True

    print(f"GOPOD_PTT_RELEASE_TAIL_WAIT seconds={RELEASE_TAIL_SECONDS}", flush=True)
    time.sleep(RELEASE_TAIL_SECONDS)

    release_event.set()
    recorder_thread.join(timeout=max(2.0, max_hold_seconds) + RELEASE_TAIL_SECONDS)
    if recorder_exception:
        raise recorder_exception[0]

    held_seconds = released_at - press_monotonic
    print(f"GOPOD_PTT_AUDIO_BYTES_CAPTURED {audio_bytes_captured}", flush=True)

    native_rate = native_rate_holder[0] if native_rate_holder else CHAT_TARGET_SAMPLE_RATE
    raw_audio = b"".join(raw_chunks)
    chat_audio = resample_pcm16(raw_audio, native_rate, CHAT_TARGET_SAMPLE_RATE)
    recognizer.AcceptWaveform(chat_audio)
    archive_audio = resample_pcm16(raw_audio, native_rate, ARCHIVE_TARGET_SAMPLE_RATE)

    if archive_audio:
        rms = audioop.rms(archive_audio, 2)
        peak = audioop.max(archive_audio, 2)
    else:
        rms = 0
        peak = 0
    near_silence = rms <= 10 and peak <= 50
    print(f"GOPOD_PTT_SIGNAL_STATS rms={rms} peak={peak} near_silence={near_silence}", flush=True)

    counter_key = target_key_label.lower()
    _capture_counters[counter_key] = _capture_counters.get(counter_key, 0) + 1
    session_dir = ensure_session_dir()
    master_path = session_dir / "masters" / f"{counter_key}_{_capture_counters[counter_key]}.wav"
    master_bytes = write_archive_wav(master_path, [archive_audio], ARCHIVE_TARGET_SAMPLE_RATE)
    print(f"GOPOD_PTT_MASTER_WAV_WRITTEN path={master_path} bytes={master_bytes}", flush=True)

    if held_seconds < min_hold_seconds:
        return None, held_seconds, repeat_events_ignored

    result = json.loads(recognizer.FinalResult() or "{}")
    transcript = " ".join(str(result.get("text", "")).split())
    return transcript, held_seconds, repeat_events_ignored


def handle_captured_key_press(
    fd: int,
    press_monotonic: float,
    args: argparse.Namespace,
    *,
    key_label: str,
    key_code: int,
) -> None:
    """Shared KP1/KP2 handler: capture -> Operator envelope -> persona
    self-ID via Ollama -> persona envelope. STAGE 3: this is what makes
    KP1 (Doc) and KP2 (Pip) symmetric - one function, called with each
    key's own code/label, instead of two near-identical copies."""
    persona = PERSONA_BY_KEY[key_label]
    try:
        transcript, held_seconds, repeat_events_ignored = capture_key_utterance(
            fd=fd,
            model_path=Path(args.model_path),
            audio_device=args.audio_device,
            press_monotonic=press_monotonic,
            min_hold_seconds=args.min_hold_seconds,
            max_hold_seconds=args.max_hold_seconds,
            target_key_code=key_code,
            target_key_label=key_label,
        )
    except Exception as exc:  # noqa: BLE001 - Stage 1: pane gets no write on error, this is diagnostic only.
        print(f"GOPOD_PTT_{key_label}_CAPTURE_ERROR {type(exc).__name__}: {exc}", flush=True)
        return

    if transcript is None:
        print(
            f"GOPOD_PTT_{key_label}_HELD_TOO_SHORT held_seconds={held_seconds:.2f} repeat_events_ignored={repeat_events_ignored}",
            flush=True,
        )
        return

    print(
        f"GOPOD_PTT_{key_label}_CAPTURED held_seconds={held_seconds:.2f} repeat_events_ignored={repeat_events_ignored}",
        flush=True,
    )
    operator_text = transcript or "[empty transcript]"
    append_operator_message(key_label, operator_text)

    persona_reply = persona_awareness_reply(persona, operator_text)
    if persona_reply is not None:
        print(f"GOPOD_PTT_AWARENESS_FAST_REPLY persona={persona}", flush=True)
    else:
        persona_reply = call_ollama_self_id(persona, operator_text)
    if persona_reply:
        robot_safe_reply = sanitize_for_robot_speech(persona_reply)
        if robot_safe_reply:
            append_persona_message(persona, robot_safe_reply)
            dispatch_robot_speech(persona, robot_safe_reply)
        else:
            print(f"GOPOD_PTT_ROBOT_DISPATCH_SKIPPED persona={persona} reason=empty_after_sanitize", flush=True)


def handle_kp1_press(fd: int, press_monotonic: float, args: argparse.Namespace) -> None:
    handle_captured_key_press(fd, press_monotonic, args, key_label="KP1", key_code=KEY_KP1)


def handle_kp2_press(fd: int, press_monotonic: float, args: argparse.Namespace) -> None:
    handle_captured_key_press(fd, press_monotonic, args, key_label="KP2", key_code=KEY_KP2)


def read_loop(device_path: Path, args: argparse.Namespace) -> int:
    session_dir = ensure_session_dir()
    install_session_log_tee(session_dir)
    print(f"GOPOD_PTT_SESSION_DIR path={session_dir}", flush=True)

    print(f"GOPOD_PTT_WRITER_READY device={device_path} name={event_name(device_path) or 'UNKNOWN'}", flush=True)
    print("GOPOD_PTT_STAGE_3 KP1=Doc KP2=Pip. NumLock gate + >=1s hold required for both.", flush=True)

    print("GOPOD_PTT_VOSK_MODEL_LOADING", flush=True)
    ensure_vosk_model_loaded(Path(args.model_path))
    print("GOPOD_PTT_VOSK_MODEL_READY", flush=True)

    if args.audio_device is None:
        matched_index, matched_name = resolve_audio_device_by_name(
            AUDIO_SOURCE_NAME_MATCH, exclude_terms=AUDIO_SOURCE_EXCLUDE_TERMS
        )
        if matched_index is not None:
            print(f"GOPOD_PTT_AUDIO_SOURCE matched={matched_name} index={matched_index}", flush=True)
            args.audio_device = matched_index
        else:
            print("GOPOD_PTT_AUDIO_SOURCE fallback=default", flush=True)

    led_state = read_numlock_led_state(device_path)
    numlock_on = led_state if led_state is not None else (args.initial_numlock.upper() == "ON")
    print(f"GOPOD_PTT_NUMLOCK_INITIAL {'ON' if numlock_on else 'OFF'} source={'led' if led_state is not None else 'cli_flag'}", flush=True)

    try:
        fd = os.open(str(device_path), os.O_RDONLY | os.O_CLOEXEC)
    except PermissionError as exc:
        raise SystemExit(
            f"GOPOD_PTT_DEVICE_PERMISSION_DENIED device={device_path} detail={exc}; "
            "run as a user with read access or pass the correct --device"
        ) from exc
    except OSError as exc:
        raise SystemExit(f"GOPOD_PTT_DEVICE_OPEN_FAILED device={device_path} detail={exc}") from exc

    grabbed = False
    try:
        fcntl.ioctl(fd, EVIOCGRAB, 1)
        grabbed = True
    except OSError as exc:
        print(f"GOPOD_PTT_DEVICE_GRAB_UNAVAILABLE {type(exc).__name__}: {exc}", flush=True)

    set_numlock_led(device_path, numlock_on)  # sync the physical LED to the seeded starting state

    kp0_times: list[float] = []
    down_key_codes: set[int] = set()
    repeat_logged_codes: set[int] = set()
    key_names = {KEY_NUMLOCK: "NumLock", KEY_KP0: "KP0"}
    exiting = False
    try:
        while True:
            readable, _, _ = select.select([fd], [], [], 1.0)

            if kp0_times and (time.monotonic() - kp0_times[0]) > args.exit_window:
                print("GOPOD_PTT_EXIT_FAILED try again", flush=True)
                kp0_times = []

            if not readable:
                continue
            data = os.read(fd, INPUT_EVENT_SIZE)
            if len(data) != INPUT_EVENT_SIZE:
                continue
            _sec, _usec, event_type, code, value = struct.unpack(INPUT_EVENT_FORMAT, data)
            if event_type != EV_KEY:
                continue

            if value == KEY_RELEASE:
                down_key_codes.discard(code)
                repeat_logged_codes.discard(code)
                continue

            if value != KEY_PRESS:
                if code in GATE_RELEVANT_CODES and code not in repeat_logged_codes:
                    print(f"GOPOD_PTT_IGNORED_REPEAT key={key_names.get(code, code)}", flush=True)
                    repeat_logged_codes.add(code)
                continue

            if code in down_key_codes:
                continue  # duplicate press without a release in between - ignore
            down_key_codes.add(code)

            if code == KEY_NUMLOCK:
                numlock_on = not numlock_on
                print(f"GOPOD_PTT_NUMLOCK {'ON' if numlock_on else 'OFF'}", flush=True)
                set_numlock_led(device_path, numlock_on)
                continue

            if code == KEY_KP1:
                if not numlock_on:
                    down_key_codes.discard(code)
                    continue  # dormant per Stage 1 (b): no capture, no attribution, no pane write
                handle_kp1_press(fd, time.monotonic(), args)
                down_key_codes.discard(code)
                continue

            if code == KEY_KP2:
                if not numlock_on:
                    down_key_codes.discard(code)
                    continue  # mirrors KP1: dormant with NumLock off, no capture, no attribution, no pane write
                handle_kp2_press(fd, time.monotonic(), args)
                down_key_codes.discard(code)
                continue

            if code == KEY_KP0:
                now = time.monotonic()
                kp0_times = [stamp for stamp in kp0_times if now - stamp <= args.exit_window]
                kp0_times.append(now)
                print(f"GOPOD_PTT_EXIT_TAP {len(kp0_times)}/{args.exit_taps}", flush=True)
                if len(kp0_times) >= args.exit_taps:
                    exiting = True
                    drain_fd(fd)
                    print("GOPOD_PTT_WRITER_EXIT", flush=True)
                    return 0
    finally:
        if exiting:
            drain_fd(fd)
        if grabbed:
            try:
                fcntl.ioctl(fd, EVIOCGRAB, 0)
            except OSError as exc:
                print(f"GOPOD_PTT_DEVICE_GRAB_RELEASE_WARN {type(exc).__name__}: {exc}", flush=True)
        os.close(fd)


def stdin_loop(args: argparse.Namespace) -> int:
    """Manual test aid only - no real hold/release timing is available over
    stdin, so this bypasses the hold-duration gate. Not part of the Stage 1
    hardware checkpoint proof."""
    print("GOPOD_PTT_WRITER_READY_STDIN Enter kp1 to simulate a >=1s KP1 capture, or exit.", flush=True)
    for line in sys.stdin:
        command = line.strip().lower()
        if command in {"exit", "quit", "kp0", "0"}:
            print("GOPOD_PTT_WRITER_EXIT", flush=True)
            return 0
        if command in {"kp1", "1"}:
            print("GOPOD_PTT_STDIN_KP1_NOT_SUPPORTED stdin cannot simulate a real press/release; use --device for a real test.", flush=True)
        elif command:
            print(f"GOPOD_PTT_STDIN_IGNORED {command}", flush=True)
    return 0


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Update GOPOD 8011 Pane 2 from local PTT/STT capture. Stage 3: KP1=Doc, KP2=Pip, Ollama self-ID.")
    parser.add_argument("--device", help="Explicit Linux input event device, for example /dev/input/event2.")
    parser.add_argument("--list-devices", action="store_true", help="Print discovered event devices and exit.")
    parser.add_argument("--resolve-device", action="store_true", help="Print just the auto-resolved device path (no discovery chatter) and exit; for shell callers.")
    parser.add_argument("--stdin", action="store_true", help="Manual test aid; cannot simulate real hold/release timing.")
    parser.add_argument("--dry-validate", action="store_true", help="Validate Vosk, sounddevice, and model path.")
    parser.add_argument("--model-path", default=str(DEFAULT_VOSK_MODEL_PATH), help="Vosk model directory.")
    parser.add_argument("--audio-device", help="Optional sounddevice input device name or index.")
    parser.add_argument("--sample-rate", type=int, default=DEFAULT_SAMPLE_RATE)
    parser.add_argument("--initial-numlock", choices=("ON", "OFF"), default="OFF", help="Initial NumLock gate state.")
    parser.add_argument("--min-hold-seconds", type=float, default=MIN_HOLD_SECONDS, help="Minimum KP1 hold to trigger capture.")
    parser.add_argument("--max-hold-seconds", type=float, default=MAX_HOLD_SECONDS, help="Safety cap in case a release event is missed.")
    parser.add_argument("--exit-taps", type=int, default=KP0_EXIT_TAPS, help="KP0 taps required to exit.")
    parser.add_argument("--exit-window", type=float, default=KP0_EXIT_WINDOW_SECONDS, help="Seconds allowed for KP0 exit taps.")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    if args.list_devices:
        for path, name in list_event_devices():
            print(f"{path} name={name or 'UNKNOWN'}")
        return 0
    if args.resolve_device:
        print(discover_device(args.device, quiet=True))
        return 0
    if args.dry_validate:
        return dry_validate(Path(args.model_path))
    if args.stdin:
        return stdin_loop(args)
    device = discover_device(args.device)
    return read_loop(device, args)


if __name__ == "__main__":
    sys.exit(main())

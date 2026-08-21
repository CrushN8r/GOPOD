#!/usr/bin/env python3
"""Write local PTT/STT transcripts into the GOPOD 8011 Pane 2 JSON file.

STAGE 1 (2026-07-03): KP1 press-and-hold >= MIN_HOLD_SECONDS with NumLock ON
records the operator's mic to a transcript and writes ONE chat envelope
attributed to "Operator". No Doc/Pip attribution, no robot/LLM calls.

STATUS UNRESOLVED - KP2 quarantined for Stage 1 (2026-07-03): KP2
press/release events are recognized (so they are not misread as noise) but
intentionally do nothing. Prior code on this file wired KP2 to a "Pip"
persona and attached Doc/Pip attribution to every transcript
(target_speaker) - both are exactly what Stage 1's non-scope forbids.
Restoring KP2 capture + persona attribution is Stage 3 work. Do not
re-enable without that stage's spec.

RESOLVED in STAGE 3 (2026-07-04): KP2 capture is live - see the STAGE 3
paragraph below.

STAGE 2 (2026-07-03): parity pass against ptt_gominion1.py's
PhysicalNumpadGate (goverlord/golverlord_older_truth/runtime_older_truth/robot/robot_io/ptt_gominion1.py).
KEY_REPEAT (evdev value 2, ~30/sec autorepeat) is now explicitly classified
and ignored for NumLock, KP1, and KP0, matching the archive's
"IGNORED_REPEAT" status - logged once per hold, not once per repeat tick.
KP0_EXIT_TAPS/KP0_EXIT_WINDOW_SECONDS are named module constants now,
byte-verified against the archive (3 / 2.0 - unchanged from Stage 1's
defaults). Raw-code-69 NumLock gating needed no change: this file already
gated on the raw code, never a name string.

STAGE 2 REVISED (2026-07-03): ALSA/JACK stderr noise fixed at the OS file
descriptor level (os.dup2), replacing the sys.stderr-level wrapper, which
cannot catch the C library's direct writes to fd 2. Device auto-discovery
now prefers checking each candidate's real evdev key-capability bitmap
(/sys/class/input/eventN/device/capabilities/key) for KEY_KP1, falling back
to name matching ("SEM USB Keyboard", "Insignia", generic "keyboard") only
if capability reads are unavailable. Added --resolve-device for the
wdtm-demo alias to pin a deterministic device path.

STAGE 2 CLEANUP (2026-07-03): the vosk Model is now loaded exactly ONCE,
at writer startup (ensure_vosk_model_loaded, called from read_loop before
the evdev fd is opened), cached as a module-level singleton, and reused
for every KP1 capture - previously it reloaded (~2s, ~12 LOG lines) on
every single press. This also removes a real timing bug: the ~2s reload
used to happen INSIDE capture_kp1_utterance, before the RawInputStream was
even opened, so on a natural "hold, speak, release" press the mic stream
often didn't start listening until after the operator had already begun
(or finished) speaking - independent of, and likely a bigger contributor
to, the reported "[empty transcript]" symptom than the sample-rate
question. KaldiRecognizer stays per-capture (cheap, correct - it holds
utterance-specific state). vosk.SetLogLevel(-1) silences Kaldi's own log
level; suppress_fd_stderr() additionally wraps the model-load call as a
fd-level backstop, per instruction, in case some vosk builds still emit
through raw stderr regardless of SetLogLevel.

STAGE 2 CLEANUP ROUND 2 (2026-07-03): raw hardware output from the first
real checkpoint run showed physical KP1/NumLock/KP0 keystrokes leaking
straight into the shell (literal "1" characters, stray escape sequences,
even after the writer exited) - because this file never exclusively
grabbed the input device, so the kernel also delivered every keystroke to
whatever had terminal focus. Fixed by porting ptt_gominion1.py's
PhysicalNumpadGate grab pattern exactly: EVIOCGRAB = 0x40044590 (byte-
verified against the archive), ioctl'd onto the fd right after open,
released before close. Also added explicit NumLock LED control - the
archive only ever *reads* the LED (read_numlock_led_state, via the
optional python-evdev package) to seed its initial gate state; there is no
LED-*write* pattern anywhere in the archive to copy, so that half is new,
built on the same already-installed python-evdev library
(InputDevice.set_led), empirically confirmed to work safely alongside the
exclusive grab (tested directly - a second python-evdev handle can read
and write LED state on a device this same process already holds grabbed).
The writer now seeds numlock_on from the real physical LED at startup
(falling back to --initial-numlock only if that read fails) and keeps the
LED in sync on every toggle, instead of tracking a Python boolean that
could silently drift from what the LED was actually showing. Also added
an audio_bytes count to the KP1 capture log lines, since the mic transcript
is still reportedly empty even with the model-load-once fix confirmed
working - this narrows whether the audio path is delivering zero bytes
(a different bug) versus real bytes vosk still can't transcribe.

STAGE 2 LISTENING + EXIT REBUILD (2026-07-03): mic transcript confirmed
working on real hardware ("hello" landed twice). Two remaining fixes.
(1) GOPOD_PTT_LISTENING now prints live, in capture_kp1_utterance's wait
loop, the moment a hold crosses min_hold_seconds while still waiting for
release - not a post-hoc report after the fact. (2) KP0 triple-tap exit
rebuilt: found the actual root cause of "000 leaks into the shell after
GOPOD_PTT_WRITER_EXIT" by reading the archive's main-loop cleanup path, not
just PhysicalNumpadGate in isolation - ptt_gominion1.py calls drain_fd()
(byte-verified shape, ported here unchanged) right before releasing
EVIOCGRAB on exit, flushing trailing buffered events (the tail end of the
3rd tap's repeat/release) that would otherwise leak once the grab drops.
Also added a proactive GOPOD_PTT_EXIT_FAILED notification when the window
lapses with fewer than 3 taps - genuinely new behavior, not in the
archive (which only prunes stale taps lazily, reactively, on the next tap
- it has no "you failed, try again" signal at all). KP0_EXIT_WINDOW_SECONDS
stays 2.0, confirmed against the archive a second time and kept unchanged
- see KP0_TRIPLE_TAP_EXIT_GOLDEN_TRUTH_001.md in gopod_notes for the
locked rule.

STAGE 2 FINAL (2026-07-03): two more bugs from a real hardware run.
(1) held_seconds was coming back several times longer than the real
physical KP1 hold (e.g. 5.19s reported for a much shorter real press).
Root cause: the release-detection loop and the audio-queue-draining loop
were the same loop - recognizer.AcceptWaveform() is a real Kaldi call that
can take a meaningful slice of time on this hardware, and every time it
ran, the release check was delayed by however long it took. Fixed by
porting the archive's actual architecture (not just its constants):
ptt_gominion1.py's record_and_transcribe_until_release runs on its own
worker thread, decoupled from run_ptt_hold_gate's dedicated release-only
polling loop on the main thread, synchronized via threading.Event.
capture_kp1_utterance now does the same - a recorder_worker thread owns
the mic stream and vosk feeding exclusively, the main thread's loop here
does nothing but watch for KEY_RELEASE/KEY_REPEAT on KP1.
(2) The KP0 "000" shell leak survived the previous drain_fd() fix because
the archive drains TWICE on this exact path, not once: immediately after
detecting the 3rd tap (drain_fd before return 0) AND again in the
finally: block (drain_fd, gated on exit-in-progress) right before
releasing EVIOCGRAB. Missing that second pass last round left a real gap:
the 3rd tap's physical release may not have happened yet at the moment of
the first drain, and the single drain gave it nowhere to land once the
grab dropped. Both drains are now ported.

STAGE 2 AUDIO REFINEMENT (2026-07-03): the writer no longer requests a
specific rate when opening the mic stream - PulseAudio's negotiated rate
for this device has drifted across this session (48000Hz earlier, 44100Hz
later), so the stream now opens with no samplerate= argument at all and
reads back stream.samplerate afterward as the actual native rate,
whatever it is. From that single stream, every audio chunk is resampled
twice via scipy.signal.resample_poly: once to ARCHIVE_TARGET_SAMPLE_RATE
(48000, fixed) accumulated into a WAV master written to
~/gopod_tts/sessions/<date>/<time>_session/masters/kp1_<n>.wav on every
capture, and once to CHAT_TARGET_SAMPLE_RATE (16000, fixed) fed to vosk
live exactly as before. Both are fixed archive/chat targets, not tied to
whatever the mic happens to be running at. This resampling work happens
entirely inside recorder_worker, on its own thread - the Stage 2 Final
threading fix is what makes adding real per-chunk CPU work here safe
without touching the (separately, dedicatedly polled) release-detection
loop at all. Session directory + a session.log stdout mirror are created
once at writer startup.

STAGE 3 (2026-07-04): Doc/Pip minimal self-ID via Ollama, KP1 and KP2
symmetric. KP2 is un-quarantined - it now captures exactly like KP1
(capture_kp1_utterance generalized into capture_key_utterance, parametrized
by which physical key/keycode it watches for release, so both paths share
one implementation rather than diverging copies) and writes its own
Operator envelope (source_producer="operator_mic_kp2"). After either
Operator envelope write (KP1 or KP2, unchanged shape), the writer POSTs
directly to Ollama's OpenAI-compatible /v1/chat/completions endpoint
(confirmed live on this host alongside its native /api/generate) via the
single call_ollama_self_id(persona, operator_text) helper, and appends the
reply as a second envelope via the single append_persona_message(persona,
text) helper - one call path, one envelope-writer, shared by both personas
so they can't drift apart. Doc gets speaker_id="doc"/
source_producer="llm_doc_self_id", Pip gets speaker_id="pip"/
source_producer="llm_pip_self_id". Deliberately NOT using the richer
persona/robot-safe-packet Ollama integration in
goverlord/gomads/intelligence_gomad/intelligence_bridge.py -
this stage's system prompts are intentionally minimal and this call is
direct, per spec. Robots stay silent; Wire-Pod is not touched. Best-effort:
any failure (Ollama down, timeout, non-200, malformed/empty response) is
logged (GOPOD_PTT_LLM_FAIL persona=<x> reason=<short>) and skipped, never
raised - this must not crash the writer or block future KP1/KP2 captures.

Template 1/2 applicability (see also
~/crushn8r_git/gopod_notes/STAGE_3_IMPLEMENT_001.md for the full writeup):
Template 1 (goverlord/runtime/songs/02_brobots_interview_run/zmisc/brobots_wirepod_interview_section_card_template_1_001.md, repointed 2026-08-19)
governs the separate Brobots/Wire-Pod interview layer - Section Card
exchanges, robot speech via Wire-Pod say_text, playAnimationWI action
packets, the pronunciation mouth-valve for TTS, HEARD_AUDIO proof. None of
that applies here: Stage 3 never touches Wire-Pod, never speaks through a
robot mouth, has no Section Card/value-target/exchange-arc concept, and
writes plain text into a chat pane rather than filtered robot speech - so
none of Template 1's rules changed this stage's system prompts or output
handling. No Template 2 file exists in this repo (searched, zero results).
"""

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
REPO_ROOT = Path(__file__).resolve().parents[4]

# STAGE PINNACLE: facts read from goverlord/runtime/data_gomad/configs/ instead of hardcoded here.
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))
from goverlord.runtime.data_gomad.configs.loader import (  # noqa: E402
    AUDIO_PATH,
    AUDIO_RESAMPLE_RULE_PATH,
    ENDPOINTS_PATH,
    PATHS_PATH,
    load_json,
    validate_envelope,
)
from goverlord.runtime.data_gomad.robot.say_replacements.text_replacement_filter import (  # noqa: E402
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
# STAGE 5: ported from the archive's record_and_transcribe_until_release
# (release_tail_seconds, DEFAULT_PTT_RELEASE_TAIL_MS=900) - trailing
# speech that continues past physical key-release was being cut off with
# no tail at all.
RELEASE_TAIL_SECONDS = 0.9
# STAGE 5: the stream opens fresh on every KP press (not persistent), so
# its first-callback latency lands at the start of the operator's speech.
# Measured directly on this hardware: 8000 samples/block (the old value)
# = ~190ms before the first real audio callback fires; 512 = ~21ms. Same
# device, same driver, only the block size changed.
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

# STAGE 3: minimal, direct Ollama self-ID for Doc and Pip. Host/port per
# this stage's spec - confirmed live and reachable on this machine (the
# Jetson - see local machine config, no network identifiers in tracked
# files). Model confirmed present via
# `ollama list`. Symmetric by design: one endpoint, one model, one timeout,
# one prompt dict, one envelope-metadata dict - Doc and Pip share all of it
# so a change to one persona's plumbing can't silently skip the other.
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

# KP1 -> Doc, KP2 -> Pip. The only place this mapping is spelled out.
PERSONA_BY_KEY = {"KP1": "doc", "KP2": "pip"}

# "Is-That-You" cross-persona awareness fast-reply, ported verbatim from the
# archived ptt_gominion1.py's persona_awareness_reply() (never shipped here
# before). Own-name mention checked before the other persona's, matching the
# archived source's per-persona branch order exactly.
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
    """Instant, no-LLM reply for the "wrong robot, is that you?" moment -
    the operator's own words name Doc or Pip while a key is held. Word-
    boundary match, not substring, so "doctor"/"pipeline" can't false-fire
    on "doc"/"pip". Returns None (fall through to call_ollama_self_id) when
    neither name is mentioned."""
    lowered = operator_text.lower()
    other = "pip" if persona == "doc" else "doc"
    if re.search(rf"\b{persona}\b", lowered):
        return PERSONA_AWARENESS_REPLIES[persona][persona]
    if re.search(rf"\b{other}\b", lowered):
        return PERSONA_AWARENESS_REPLIES[persona][other]
    return None

# STAGE 4: minimal robot dispatch. Serials are the invariant constants
# (CLAUDE.md: Brobot 1/Doc/vector1 = 0dd1b9e9, Brobot 2/Pip/vector2 =
# 0dd1d8bf) - same values PERSONA_BY_KEY's KP1/KP2 already route to.
PERSONA_ROBOT_SERIAL = {"doc": "0dd1b9e9", "pip": "0dd1d8bf"}
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

    # STAGE 2 REVISED: prefer the node whose capability bitmap actually
    # advertises KEY_KP1 - this is what fixes the "multiple SEM event nodes,
    # name match alone can't tell them apart" wonky item. Name matching is
    # kept only as a fallback for when /sys capability reads aren't
    # available (permissions, non-Linux, etc.).
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
    """Writes exactly ONE chat envelope. Schema-locked: speaker_id,
    display_name, text, timestamp, source_producer. No variations.
    key_label is "KP1" or "KP2" - selects which source_producer to stamp
    (operator_mic_kp1 / operator_mic_kp2); the operator identity itself
    (speaker_id/display_name) is the same either way."""
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
    """Writes exactly ONE chat envelope for a persona's Ollama self-ID
    reply. Same schema as append_operator_message - speaker_id,
    display_name, text, timestamp, source_producer. Shared by Doc and Pip
    (persona in PERSONA_ENVELOPE_META) - one writer, not one per persona."""
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

    # FIX (1), STAGE 2 FINAL: ptt_gominion1.py never mixes release-polling
    # and audio processing in one loop. record_and_transcribe_until_release
    # runs the mic stream + recognizer.AcceptWaveform on its own worker
    # thread; run_ptt_hold_gate's release_seen(0.02) polls the evdev fd in
    # a dedicated, otherwise-empty loop on the main thread, synchronized by
    # a threading.Event. The prior single-loop version alternated between
    # select()-polling for release and draining the audio queue in the same
    # iteration - recognizer.AcceptWaveform() is a real, sometimes-slow
    # Kaldi call on this hardware, and a slow call there stalls the release
    # check for however long it takes, which is exactly what produced
    # held_seconds several times longer than the real physical hold.
    # AUDIO REFINEMENT adds real per-chunk resampling CPU work to this same
    # worker thread - safe to do because that thread was already fully
    # decoupled from the release-detection loop by the Stage 2 Final fix;
    # nothing added here can stall the KEY_RELEASE check.
    # Deliberately NOT ported: the archive's intent-arm delay and
    # cancel-as-short-press state machine - those are Layer 0/multichat-
    # bridging concerns this writer doesn't have; min_hold_seconds already
    # covers the "too short, discard" case post-hoc here.
    release_event = threading.Event()
    audio_bytes_captured = 0
    raw_chunks: list[bytes] = []
    native_rate_holder: list[int] = []
    recorder_exception: list[BaseException] = []

    # STAGE 5: raw accumulation only, no per-chunk resample/feed. Confirmed
    # via direct hardware test that this device rejects samplerate=16000
    # outright (PortAudioError -9997, Invalid sample rate) - matching the
    # archive's native-target-rate stream is not possible on this hardware,
    # so AUDIO REFINEMENT's 44100-native capture stays. But resampling each
    # ~181ms (8000-sample) callback chunk independently through
    # scipy.signal.resample_poly (a polyphase FIR filter with no state
    # carried between calls) puts a filter-edge discontinuity at every
    # chunk boundary - dozens of them across a multi-second hold. RMS/peak
    # (STAGE 5's earlier signal-stats addition) can't see this: it's a
    # waveform-continuity artifact, not a level/loudness problem, which is
    # exactly why "signal is fine" and "words vanish/garble" were both true
    # at once. Fix: accumulate raw native-rate audio here; resample the
    # whole continuous buffer ONCE, after capture ends, for both the chat
    # (vosk) and archive targets - same approach the archive itself used
    # (batch-transcribe after a fully captured, continuous buffer), just
    # adapted to this hardware's native rate instead of a directly-opened
    # target rate.
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
                # FIX (1): "close immediately" on release, per instruction -
                # no archive-style release-tail wait. Still drains whatever
                # is already queued once released, so already-captured audio
                # isn't thrown away, but does not wait for more to arrive.
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

    # Dedicated release-detection loop - this is now the ONLY thing running
    # here, matching run_ptt_hold_gate's release_seen(0.02) shape (archive
    # polls at 0.02s; kept here for the same responsiveness).
    released_at: float | None = None
    repeat_events_ignored = 0
    repeat_logged = False
    listening_announced = False
    while released_at is None:
        if time.monotonic() - press_monotonic > max_hold_seconds:
            released_at = time.monotonic()
            break
        # FIX (1): fires live, the moment the hold crosses the accept
        # threshold - not a post-hoc "was it long enough" report after
        # release. Terminal print only, exactly once per hold.
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
                        # STAGE 2 parity: KEY_REPEAT ignored, matches archive's
                        # IGNORED_REPEAT. Logged once per hold, not per tick
                        # (~30/sec) - avoids flooding output during capture.
                        repeat_events_ignored += 1
                        if not repeat_logged:
                            print(f"GOPOD_PTT_IGNORED_REPEAT key={target_key_label}", flush=True)
                            repeat_logged = True

    # STAGE 5: release-tail wait, ported from the archive's
    # record_and_transcribe_until_release (release_tail_seconds,
    # DEFAULT_PTT_RELEASE_TAIL_MS=900) - a real gap this generalized
    # rewrite dropped, not a deliberate omission like the intent-arm delay
    # noted above. Human PTT behavior releases the key at or slightly
    # before the last word finishes; stopping capture the instant release
    # fires truncates that tail. recorder_worker keeps draining the queue
    # and feeding the recognizer for this extra window too, since
    # release_event is set after the wait, not before.
    print(f"GOPOD_PTT_RELEASE_TAIL_WAIT seconds={RELEASE_TAIL_SECONDS}", flush=True)
    time.sleep(RELEASE_TAIL_SECONDS)

    release_event.set()
    recorder_thread.join(timeout=max(2.0, max_hold_seconds) + RELEASE_TAIL_SECONDS)
    if recorder_exception:
        raise recorder_exception[0]

    held_seconds = released_at - press_monotonic
    print(f"GOPOD_PTT_AUDIO_BYTES_CAPTURED {audio_bytes_captured}", flush=True)

    # STAGE 5: one continuous resample per target rate, not one per
    # ~181ms callback chunk - see process_chunk's comment above for why.
    # native_rate_holder is only empty if the stream never actually opened
    # (no chunks either, in that case), so this fallback is never exercised
    # against real audio.
    native_rate = native_rate_holder[0] if native_rate_holder else CHAT_TARGET_SAMPLE_RATE
    raw_audio = b"".join(raw_chunks)
    chat_audio = resample_pcm16(raw_audio, native_rate, CHAT_TARGET_SAMPLE_RATE)
    recognizer.AcceptWaveform(chat_audio)
    archive_audio = resample_pcm16(raw_audio, native_rate, ARCHIVE_TARGET_SAMPLE_RATE)

    # STAGE 5: signal-strength diagnostic, ported from the original
    # ptt_gominion.py (wav_signal_stats/near_silence_warning,
    # NEAR_SILENCE_RMS=10/NEAR_SILENCE_PEAK=50) - the live writer had no
    # visibility at all into whether captured audio was strong speech or
    # near-silence, which is exactly the missing piece for diagnosing
    # "words vanished from the transcript": read-only, does not change
    # capture/timing/dispatch, just reports what was actually in the WAV.
    if archive_audio:
        rms = audioop.rms(archive_audio, 2)
        peak = audioop.max(archive_audio, 2)
    else:
        rms = 0
        peak = 0
    near_silence = rms <= 10 and peak <= 50
    print(f"GOPOD_PTT_SIGNAL_STATS rms={rms} peak={peak} near_silence={near_silence}", flush=True)

    # FIX (2): archive every capture attempt, not just accepted ones - an
    # honest record of what the mic actually heard, including short-hold
    # taps that get discarded below. Written even if archive_chunks ended
    # up empty (a zero-length-audio WAV header, not skipped silently).
    # STAGE 3: counter is now per-key (dict keyed by target_key_label) so
    # KP1 and KP2 captures get independent numbering and never overwrite
    # each other's master WAV (both used to share one "kp1_N.wav" counter
    # before KP2 had a capture path of its own).
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

    # Cross-persona awareness fast-reply, checked first: if the operator's
    # own words actually name Doc or Pip, reply instantly, no Ollama call.
    # Falls through to the existing self-ID call when neither is mentioned.
    persona_reply = persona_awareness_reply(persona, operator_text)
    if persona_reply is not None:
        print(f"GOPOD_PTT_AWARENESS_FAST_REPLY persona={persona}", flush=True)
    else:
        # STAGE 3: persona self-ID, direct Ollama call. Best-effort - a failed
        # or empty reply is logged (GOPOD_PTT_LLM_FAIL, inside
        # call_ollama_self_id) and skipped; the Operator envelope above is
        # unaffected either way.
        persona_reply = call_ollama_self_id(persona, operator_text)
    if persona_reply:
        # STAGE 7: sanitize once, upstream - same rich-display-vs-filtered-
        # robot-safe-say split the interview pipeline uses. The terminal/
        # chat pane now shows the filtered text (what the robot actually
        # says), not the raw LLM output with its stage directions and
        # markdown - that raw text never gets printed anywhere.
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
    # AUDIO REFINEMENT FIX (4): session directory + stdout mirror created
    # first, before any other output, so everything from here on is also
    # captured in session.log.
    session_dir = ensure_session_dir()
    install_session_log_tee(session_dir)
    print(f"GOPOD_PTT_SESSION_DIR path={session_dir}", flush=True)

    print(f"GOPOD_PTT_WRITER_READY device={device_path} name={event_name(device_path) or 'UNKNOWN'}", flush=True)
    print("GOPOD_PTT_STAGE_3 KP1=Doc KP2=Pip. NumLock gate + >=1s hold required for both.", flush=True)

    # FIX (2): load the vosk model ONCE here, at writer startup, before the
    # evdev fd is even opened - not lazily on the first KP1 press. This is
    # also what removes the ~2s pre-stream-open delay that used to eat into
    # (or entirely consume) the operator's actual speech window.
    print("GOPOD_PTT_VOSK_MODEL_LOADING", flush=True)
    ensure_vosk_model_loaded(Path(args.model_path))
    print("GOPOD_PTT_VOSK_MODEL_READY", flush=True)

    # STAGE E: resolve the mic by name once, at launch, instead of trusting
    # PulseAudio's default routing - confirmed this session that the
    # default can silently drift onto a virtual/remapped (silent) source.
    # An explicit --audio-device always wins; auto-match only fills in
    # when the operator didn't specify one.
    if args.audio_device is None:
        matched_index, matched_name = resolve_audio_device_by_name(
            AUDIO_SOURCE_NAME_MATCH, exclude_terms=AUDIO_SOURCE_EXCLUDE_TERMS
        )
        if matched_index is not None:
            print(f"GOPOD_PTT_AUDIO_SOURCE matched={matched_name} index={matched_index}", flush=True)
            args.audio_device = matched_index
        else:
            print("GOPOD_PTT_AUDIO_SOURCE fallback=default", flush=True)

    # STAGE 2 CLEANUP ROUND 2: prefer the real physical LED state over the
    # --initial-numlock CLI default, matching ptt_gominion1.py's
    # read_numlock_led_state seeding - falls back to the CLI flag only if
    # the LED can't be read (device doesn't support it, python-evdev
    # unavailable, etc.), rather than silently assuming OFF.
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

    # STAGE 2 CLEANUP ROUND 2: exclusive grab, ported from ptt_gominion1.py's
    # PhysicalNumpadGate usage (EVIOCGRAB). Without this, physical KP1/
    # NumLock/KP0 keystrokes also leak to whatever has terminal focus - this
    # was confirmed directly from a real hardware checkpoint run (literal
    # "1" characters and stray escape sequences landing in the shell).
    # Non-fatal if the grab fails (permissions, unsupported device) - warn
    # and continue, matching the archive's own try/except OSError shape.
    grabbed = False
    try:
        fcntl.ioctl(fd, EVIOCGRAB, 1)
        grabbed = True
    except OSError as exc:
        print(f"GOPOD_PTT_DEVICE_GRAB_UNAVAILABLE {type(exc).__name__}: {exc}", flush=True)

    set_numlock_led(device_path, numlock_on)  # sync the physical LED to the seeded starting state

    kp0_times: list[float] = []
    # STAGE 2 parity: down_key_codes mirrors ptt_gominion1.py's PhysicalNumpadGate
    # dedup - a repeated KEY_PRESS without an intervening KEY_RELEASE is ignored,
    # and it's what lets us log "IGNORED_REPEAT" exactly once per hold instead of
    # once per ~33ms autorepeat tick.
    down_key_codes: set[int] = set()
    repeat_logged_codes: set[int] = set()
    key_names = {KEY_NUMLOCK: "NumLock", KEY_KP0: "KP0"}
    exiting = False
    try:
        while True:
            readable, _, _ = select.select([fd], [], [], 1.0)

            # FIX (2): proactive window-expiry check, evaluated every loop
            # tick (not just on a new tap) - the archive only ever prunes
            # kp0_times lazily, reactively, on the NEXT tap, so if a 3rd tap
            # never comes it just sits there forever with no notification.
            # This is new behavior layered on top of the archive's counting
            # mechanism, not a change to it.
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
                # STAGE 2 parity: KEY_REPEAT (or any other non-press/release value)
                # on a gate-relevant key is ignored, matching the archive's
                # IGNORED_REPEAT status. Logged once per hold, not per tick.
                if code in GATE_RELEVANT_CODES and code not in repeat_logged_codes:
                    print(f"GOPOD_PTT_IGNORED_REPEAT key={key_names.get(code, code)}", flush=True)
                    repeat_logged_codes.add(code)
                continue

            # value == KEY_PRESS from here.
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
                # KP1's own KEY_RELEASE was already consumed inside the capture
                # wait-loop above, so it never reaches this loop to clear the
                # down-state - clear it here now that the physical key is known
                # to have been released (capture_key_utterance only returns
                # after seeing KEY_RELEASE or hitting the max-hold safety cap).
                down_key_codes.discard(code)
                continue

            if code == KEY_KP2:
                if not numlock_on:
                    down_key_codes.discard(code)
                    continue  # mirrors KP1: dormant with NumLock off, no capture, no attribution, no pane write
                handle_kp2_press(fd, time.monotonic(), args)
                # Mirrors KP1: KP2's own KEY_RELEASE was already consumed
                # inside the capture wait-loop above, so it never reaches
                # this loop to clear the down-state - clear it here now
                # that the physical key is known to have been released.
                down_key_codes.discard(code)
                continue

            if code == KEY_KP0:
                now = time.monotonic()
                kp0_times = [stamp for stamp in kp0_times if now - stamp <= args.exit_window]
                kp0_times.append(now)
                print(f"GOPOD_PTT_EXIT_TAP {len(kp0_times)}/{args.exit_taps}", flush=True)
                if len(kp0_times) >= args.exit_taps:
                    # FIX (2), STAGE 2 FINAL: the archive drains TWICE on this
                    # exact path, not once - immediately here, and again in
                    # the finally: block below (gated on `exiting`) right
                    # before EVIOCGRAB is released. The first pass alone
                    # left a real gap: the 3rd tap's physical release may
                    # not have happened yet at this exact instant, so there
                    # was nothing to drain here - the second pass, running
                    # slightly later, is what actually catches it.
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

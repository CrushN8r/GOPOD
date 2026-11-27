#!/usr/bin/env python3
import importlib.util
import json
import os
import re
import subprocess
import threading
import time
import types
import unicodedata
import urllib.parse
import urllib.request
from pathlib import Path


WIREPOD_CHIPPER_ROOT = Path(os.getenv("GOPOD_WIREPOD_CHIPPER_ROOT", "/home/goverlord/wire-pod/chipper")).expanduser()
GOPOD_CHIPPER_ROOT = Path(
    os.getenv("GOPOD_OVERLAY_CHIPPER", str(Path(__file__).resolve().parents[3] / "wire_pod_overlay" / "chipper"))
).expanduser()
DEFAULT_CARD_PATH = GOPOD_CHIPPER_ROOT / "gopod_probes" / "section_packets" / "section_01_brobots_gopod_card_001.txt"
CARD_PATH = Path(os.getenv("GOPOD_SECTION_CARD_PATH", str(DEFAULT_CARD_PATH))).expanduser()
LOG_DIR = GOPOD_CHIPPER_ROOT / "gopod_probes" / "demo_runs"
API_CONFIG = WIREPOD_CHIPPER_ROOT / "apiConfig.json"
CONTROL_SONG_PATH = Path(__file__).resolve().parent / "run_robot_control_song_001.py"
GOPOD_REPO_ROOT = Path(os.getenv("GOPOD_REPO_ROOT", "/home/goverlord/crushn8r_git/GOPOD")).expanduser()
GOPOD_NOTES_DIR = Path(os.getenv("GOPOD_NOTES_DIR", str(GOPOD_REPO_ROOT.parent / "gopod_notes"))).expanduser()
TEMPLATE1_PATH = GOPOD_REPO_ROOT / "goverlord/wire_pod_overlay/chipper/gopod_probes/section_packets/brobots_wirepod_interview_section_card_template_1_001.md"
SPEECH_PRONUNCIATION_REGISTRY_PATH = Path(
    os.getenv(
        "GOPOD_SPEECH_PRONUNCIATION_REGISTRY",
        str(TEMPLATE1_PATH),
    )
).expanduser()
INTERVIEW_SCAFFOLD_PATH = Path(
    os.getenv(
        "GOPOD_BROBOTS_INTERVIEW_SCAFFOLD",
        str(TEMPLATE1_PATH),
    )
).expanduser()
GOPOD_ANIMATION_VOCAB_PATH = Path(
    os.getenv("GOPOD_ANIMATION_VOCAB_PATH", str(GOPOD_CHIPPER_ROOT / "animation_vocab.json"))
).expanduser()
DEFAULT_SECTION_SONG_DIR = GOPOD_REPO_ROOT / "goverlord/runtime/songs/zzz_archives/02b_brobots_interview_run"
LEGACY_SECTION_CARD_ENV = "GOPOD_USE_LEGACY_SECTION_CARD"

LIVE_GATE = "GOPOD_ALLOW_LIVE_ROBOT_SPEECH"
WIREPOD_BASE_URL = os.getenv("GOPOD_WIREPOD_BASE_URL")
ROBOT_EXPRESSION_GATE = os.getenv("ROBOT_EXPRESSION", "")
BROBOT_1_SERIAL = os.getenv("GOPOD_BROBOT_1_SERIAL", "ESN01")
BROBOT_2_SERIAL = os.getenv("GOPOD_BROBOT_2_SERIAL", "ESN02")
BROBOTS_SOLO = False
try:
    _now = json.loads((Path(__file__).resolve().parents[3] / "wire_pod_overlay/chipper/gopod_brobots_now.json").read_text())
    if not os.getenv("GOPOD_BROBOT_1_SERIAL") and _now.get("brobot_1"):
        BROBOT_1_SERIAL = _now["brobot_1"]
    if not os.getenv("GOPOD_BROBOT_2_SERIAL") and _now.get("brobot_2"):
        BROBOT_2_SERIAL = _now["brobot_2"]
    BROBOTS_SOLO = bool(_now.get("solo"))
except (OSError, ValueError):
    pass


INTERVIEW_KG_SEARCH_SONG_ID = "brobots_interview_section_01"
INTERVIEW_KG_SEARCH_SECONDS = 3.0
INTERVIEW_KG_SEARCH_LOOP_INTERVAL = 0.333
BROBOT_3_ESN = "VOICE_ONLY_NO_HARDWARE"
BROBOT_4_ESN = "VOICE_ONLY_NO_HARDWARE"
BROBOT_4_DISPLAY_ROLE = "Brobot 4"
RESTART_WIREPOD_BEFORE_RUN = os.getenv("GOPOD_RESTART_WIREPOD_BEFORE_RUN", "0") == "1"
WIREPOD_POST_RESTART_SETTLE_SECONDS = float(os.getenv("GOPOD_WIREPOD_POST_RESTART_SETTLE_SECONDS", "4"))
SECTION1_MAX_EXCHANGES = int(os.getenv("GOPOD_SECTION1_MAX_EXCHANGES", "0") or "0")
ROBOT_SPEECH_CHUNK_LIMIT = int(os.getenv("GOPOD_ROBOT_SPEECH_CHUNK_LIMIT", "420") or "420")
SECTION1_REPLAY_LOG = os.getenv("GOPOD_SECTION1_REPLAY_LOG", "").strip()
VOICE_DESTINATION_ENV = "GOPOD_VOICE_DESTINATION"
VOICE_DESTINATIONS = ("robot", "monitor", "off")
CONTENT_MODEL_STATE_PATH = GOPOD_REPO_ROOT / "goverlord/runtime/songs/zzz_archives/02b_brobots_interview_run/content_model_state.json"
DEFAULT_CONTENT_MODEL = "gemma2:2b"
PRESHOW_SONG_DIR_ENV = "GOPOD_PRESHOW_SONG_DIR"
PRESHOW_READ_SHEET_ENV = "GOPOD_PRESHOW_READ_SHEET"

KOKORO_STATUS_ANNOUNCER_PATH = (
    GOPOD_REPO_ROOT
    / "goverlord/runtime/data_gomad/robot/kokoro_voice/local_tts_voice_registry_001/interview_status_kokoro_announcer_001.py"
)
KOKORO_STATUS_VENV_PYTHON = Path(
    os.getenv("GOPOD_KOKORO_VENV_PYTHON", "/home/goverlord/gopod_tts/engines/kokoro_82m_001/venv/bin/python3")
).expanduser()
KOKORO_STATUS_VOICE = "af_bella"
KOKORO_HOST4_VOICE = "am_puck"
_KOKORO_HOST4_STATE = {"attempted": False, "proc": None}
KOKORO_STATUS_STOP_PREFIXES = (
    "STATUS: PLAYBACK_START",
    "STATUS: PLAYBACK_TURN_START",
    "Brobot 1:",
    "Brobot 2:",
)
_KOKORO_STATUS_STATE = {"attempted": False, "proc": None, "active": True, "live": False}

BROBOT_3_DISPLAY_ROLE = "Brobot 3"

REQUIRED_INTERVIEW_SCAFFOLD_KEYS = (
    "scaffold_id",
    "system_prompt",
    "shared_rules",
    "role_tasks",
    "response_context",
    "display_role_template",
    "display_exchange_halves",
    "line_modes",
    "trigger_phrase_punctuation_soften",
    "inline_emphasis_punctuation_rule",
    "speech_cleanup_rules",
    "speaker_visual_cue",
    "between_exchange_pause_seconds",
)


_VOICE_ENGINE_STATE = {"kokoro": None}


def resolve_voice_engine(engine_key, scaffold, required_for=None):
    registry = (scaffold or {}).get("voice_engines", {})
    entry = registry.get(engine_key)
    if entry is None:
        raise RuntimeError(f"BLOCKED: voice engine {engine_key!r} is not in the voice_engines registry")
    if entry.get("readiness") != "ready":
        reason = entry.get("readiness_reason", "not marked ready")
        raise RuntimeError(f"BLOCKED: voice engine {engine_key!r} is not ready for use: {reason}")
    if required_for and required_for not in (entry.get("wired_for") or []):
        raise RuntimeError(
            f"BLOCKED: voice engine {engine_key!r} is ready but not wired for "
            f"{required_for!r} (wired_for={entry.get('wired_for')})"
        )
    return entry


def resolve_voice_destination():
    raw = os.getenv(VOICE_DESTINATION_ENV, "robot").strip()
    if raw not in VOICE_DESTINATIONS:
        raise RuntimeError(
            f"BLOCKED: {VOICE_DESTINATION_ENV}={raw!r} must be one of {VOICE_DESTINATIONS}"
        )
    return raw


def configure_voice_engines(scaffold):
    engine_key = (scaffold or {}).get("pre_chat_status_voice", {}).get("engine", "kokoro_82m")
    entry = resolve_voice_engine(engine_key, scaffold, required_for="pre_show_narration")
    invocation = entry.get("invocation", {})
    _VOICE_ENGINE_STATE["kokoro"] = {
        "venv_python": invocation.get("venv_python", str(KOKORO_STATUS_VENV_PYTHON)),
        "script": invocation.get("script", str(KOKORO_STATUS_ANNOUNCER_PATH)),
        "voice": entry.get("voice") or KOKORO_STATUS_VOICE,
    }


def _kokoro_status_ensure_loaded():
    if _KOKORO_STATUS_STATE["attempted"]:
        return
    _KOKORO_STATUS_STATE["attempted"] = True
    config = _VOICE_ENGINE_STATE.get("kokoro")
    venv_python = config["venv_python"] if config else str(KOKORO_STATUS_VENV_PYTHON)
    script = config["script"] if config else str(KOKORO_STATUS_ANNOUNCER_PATH)
    voice = config["voice"] if config else KOKORO_STATUS_VOICE
    try:
        proc = subprocess.Popen(
            [
                venv_python,
                script,
                "--speak-stdin",
                "--voice",
                voice,
            ],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            text=True,
            bufsize=1,
        )
        _KOKORO_STATUS_STATE["proc"] = proc
    except Exception as exc:  # noqa: BLE001 - status voice must never break the interview.
        print(f"STATUS: KOKORO_STATUS_VOICE_UNAVAILABLE {type(exc).__name__}: {exc}", flush=True)


def kokoro_status_warm_up():
    """Diagnostic finding (2026-07-07 live audio path test): KPipeline()
    model load costs ~20+s on this box (CPU only - no usable CUDA, see the
    torch driver warning). Without this, that cost silently lands on
    RUN_START - the very first status line the interview ever prints -
    turning the show's opening beat into an unexplained ~20s dead-air
    stall. Call this once, before the first status() call, so the load
    happens as its own clearly-labeled step instead.
    Uses only the --speak-stdin worker's existing protocol - an empty line
    always gets an immediate DONE once the worker reaches its stdin loop,
    which only happens after KPipeline() finishes - no announcer script
    change needed to get a readiness signal out of it."""
    _kokoro_status_ensure_loaded()
    proc = _KOKORO_STATUS_STATE["proc"]
    if proc is None or proc.stdin is None or proc.stdout is None:
        print("STATUS: KOKORO_WARMUP_SKIPPED voice unavailable", flush=True)
        return
    print("STATUS: KOKORO_WARMUP_START", flush=True)
    try:
        proc.stdin.write("\n")
        proc.stdin.flush()
        ack = proc.stdout.readline()
        if not ack:
            raise RuntimeError("kokoro status speak worker closed stdout unexpectedly")
        print("STATUS: KOKORO_WARMUP_DONE", flush=True)
    except Exception as exc:  # noqa: BLE001 - warm-up must never break the interview.
        print(f"STATUS: KOKORO_WARMUP_ERROR {type(exc).__name__}: {exc}", flush=True)
        _KOKORO_STATUS_STATE["active"] = False


def _wirepod_log_send(display_role, display_text, kind, error_label, scaffold=None):
    if kind not in ("content", "marker"):
        raise RuntimeError(f"BLOCKED: channel send kind {kind!r} must be 'content' or 'marker'")
    if not _KOKORO_STATUS_STATE["live"]:
        return
    serial = BROBOT_1_SERIAL
    try:
        wirepod_web_send_form("/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10)
        wirepod_web_send_form(
            "/api-sdk/say_text",
            {
                "text": "",
                "serial": serial,
                "display_role": display_role,
                "display_text": display_text,
            },
            timeout=10,
        )
    except Exception as exc:  # noqa: BLE001 - log mirror/marker must never break the run.
        print(f"STATUS: {error_label} {type(exc).__name__}: {exc}", flush=True)
    finally:
        try:
            wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)
        except Exception:  # noqa: BLE001
            pass


def _wirepod_log_status(display_text):
    _wirepod_log_send(BROBOT_3_DISPLAY_ROLE, display_text, kind="content", error_label="WIREPOD_LOG_MIRROR_ERROR")


GOPOD_SONG_START_ROLE = "GOPOD_SONG_START"
GOPOD_SONG_END_ROLE = "GOPOD_SONG_END"


def _wirepod_log_marker(display_role, display_text):
    _wirepod_log_send(display_role, display_text, kind="marker", error_label="WIREPOD_SONG_MARKER_ERROR")


_STATUS_BEAT_PREFIXES = (
    ("RUN_START", "STATUS: RUN_START"),
    ("SCAFFOLD_LOAD", "STATUS: SCAFFOLD_LOAD"),
    ("SCAFFOLD_READY", "STATUS: SCAFFOLD_READY"),
    ("SECTION_CARD_LOAD", "STATUS: SECTION_CARD_LOAD"),
    ("SECTION_CARD_READY", "STATUS: SECTION_CARD_READY"),
    ("WIREPOD_ROUTE", "LIVE: Wire-Pod ESN routing enabled"),
    ("BROBOT_1_ESN", "LIVE: Brobot 1 ESN:"),
    ("BROBOT_2_ESN", "LIVE: Brobot 2 ESN:"),
    ("GENERATION_START", "STATUS: GENERATION_START"),
    ("EXCHANGE_GENERATION_START", "STATUS: EXCHANGE_GENERATION_START"),
    ("LLM_START", "STATUS: LLM_START"),
    ("LLM_DONE", "STATUS: LLM_DONE"),
    ("EXCHANGE_GENERATION_DONE", "STATUS: EXCHANGE_GENERATION_DONE"),
    ("GENERATION_DONE", "STATUS: GENERATION_DONE"),
)
_STATUS_BEAT_COLOUR_CACHE = {}

_STATUS_BEAT_NOTHING_TO_COLOUR_KEY = "NOTHING_TO_COLOUR"
_STATUS_BEAT_NOTHING_TO_COLOUR_HINT = (
    "There's nothing new to report right now - just a quiet beat while things "
    "get ready. Say something brief and warm to keep the audience company."
)


def match_status_beat(line):
    for beat_key, prefix in _STATUS_BEAT_PREFIXES:
        if line.startswith(prefix):
            return beat_key
    return None


def llm_colour_status_beat(beat_key, hint, scaffold=None):
    scaffold = scaffold or load_interview_scaffold()
    host = scaffold.get("brobot_3_host", {})
    api = load_api_config()
    endpoint = api.get("endpoint", "")
    if not endpoint:
        raise RuntimeError("knowledge endpoint is empty")
    payload = {
        "model": api.get("model") or "gemma2:2b",
        "messages": [
            {"role": "system", "content": host.get("system_prompt", "")},
            {
                "role": "user",
                "content": f"Situation: {hint}\nSay one brief, warm spoken sentence about this to the live audience.",
            },
        ],
        "temperature": api.get("temp", 0) or 0,
        "top_p": api.get("top_p", 1) or 1,
        "stream": False,
    }
    request = urllib.request.Request(
        endpoint.rstrip("/") + "/chat/completions",
        data=json.dumps(payload).encode("utf-8"),
        headers={"Content-Type": "application/json", "Authorization": "Bearer " + str(api.get("key", ""))},
        method="POST",
    )
    with urllib.request.urlopen(request, timeout=30) as response:
        data = json.loads(response.read().decode("utf-8"))
    choices = data.get("choices") or []
    if not choices:
        raise RuntimeError("empty LLM response")
    content = choices[0].get("message", {}).get("content", "").strip()
    if not content:
        raise RuntimeError("empty coloured status line")
    return content


def llm_colour_preshow_line(system_prompt, hint, content_model=None):
    api = load_api_config()
    endpoint = api.get("endpoint", "")
    if not endpoint:
        raise RuntimeError("knowledge endpoint is empty")
    payload = {
        "model": content_model or api.get("model") or "gemma2:2b",
        "messages": [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": hint},
        ],
        "temperature": api.get("temp", 0) or 0,
        "top_p": api.get("top_p", 1) or 1,
        "stream": False,
    }
    request = urllib.request.Request(
        endpoint.rstrip("/") + "/chat/completions",
        data=json.dumps(payload).encode("utf-8"),
        headers={"Content-Type": "application/json", "Authorization": "Bearer " + str(api.get("key", ""))},
        method="POST",
    )
    with urllib.request.urlopen(request, timeout=120) as response:
        data = json.loads(response.read().decode("utf-8"))
    choices = data.get("choices") or []
    if not choices:
        raise RuntimeError("empty LLM response")
    content = choices[0].get("message", {}).get("content", "").strip()
    if not content:
        raise RuntimeError("empty coloured preshow line")
    return content


def coloured_status_line(line, scaffold=None):
    """Brobot 3's per-beat colouring pass. Matches `line` to a known
    pre-show beat, colours it through the LLM once per beat key per run
    (cached thereafter). A line matching no known beat, or a known beat with
    no configured hint, is coloured against a generic "nothing to colour"
    situation instead (its own cache key) so Kokoro never speaks a raw
    technical STATUS: line - and falls back to the original flat `line` only
    if that colouring pass itself fails, since the pre-show must keep
    working even if the LLM is unavailable, and the technical `status()`
    print is untouched either way."""
    beat_key = match_status_beat(line)
    scaffold = scaffold or load_interview_scaffold()
    hint = ""
    if beat_key is not None:
        if beat_key in _STATUS_BEAT_COLOUR_CACHE:
            return _STATUS_BEAT_COLOUR_CACHE[beat_key]
        hint = scaffold.get("brobot_3_host", {}).get("status_beats", {}).get(beat_key, "")
    if not hint:
        beat_key = _STATUS_BEAT_NOTHING_TO_COLOUR_KEY
        hint = _STATUS_BEAT_NOTHING_TO_COLOUR_HINT
        if beat_key in _STATUS_BEAT_COLOUR_CACHE:
            return _STATUS_BEAT_COLOUR_CACHE[beat_key]
    try:
        coloured = llm_colour_status_beat(beat_key, hint, scaffold)
    except (Exception, KeyboardInterrupt) as exc:  # noqa: BLE001 - Brobot 3 colouring must never break the pre-show,
        print(f"STATUS: BROBOT3_COLOUR_FALLBACK beat={beat_key} {type(exc).__name__}: {exc}", flush=True)
        return line
    _STATUS_BEAT_COLOUR_CACHE[beat_key] = coloured
    return coloured


def _announce_status_kokoro(line):
    if not _KOKORO_STATUS_STATE["active"]:
        return
    if any(line.startswith(prefix) for prefix in KOKORO_STATUS_STOP_PREFIXES):
        _KOKORO_STATUS_STATE["active"] = False
        if _KOKORO_STATUS_STATE.get("voice_destination") != "monitor":
            proc = _KOKORO_STATUS_STATE["proc"]
            if proc is not None and proc.stdin is not None:
                try:
                    proc.stdin.close()
                except Exception:  # noqa: BLE001
                    pass
        return
    spoken_line = coloured_status_line(line)
    _wirepod_log_status(spoken_line)
    _kokoro_status_ensure_loaded()
    proc = _KOKORO_STATUS_STATE["proc"]
    if proc is None or proc.stdin is None or proc.stdout is None or proc.poll() is not None:
        return
    try:
        proc.stdin.write(spoken_line + "\n")
        proc.stdin.flush()
        ack = proc.stdout.readline()
        if not ack:
            raise RuntimeError("kokoro status speak worker closed stdout unexpectedly")
    except Exception as exc:  # noqa: BLE001 - status voice must never break the interview.
        print(f"STATUS: KOKORO_STATUS_VOICE_PLAYBACK_ERROR {type(exc).__name__}: {exc}", flush=True)
        _KOKORO_STATUS_STATE["active"] = False


def status(message):
    print(f"STATUS: {message}", flush=True)
    _announce_status_kokoro(f"STATUS: {message}")


class TimingLog:
    def __init__(self):
        self._lock = threading.Lock()
        self.segments = []

    def record(self, segment_id, label, capture_quality, note="", **extra):
        if capture_quality not in ("true_completion", "accept_only"):
            raise RuntimeError(
                f"BLOCKED: timing capture_quality {capture_quality!r} must be 'true_completion' or 'accept_only'"
            )
        return _TimingSpan(self, segment_id, label, capture_quality, note, extra)

    def note_summary(self, segment_id, label, capture_quality, elapsed_seconds, note="", **extra):
        """For a span whose start/end can't be wrapped directly (e.g. a loop
        whose own body already fires its own per-iteration spans) - records
        one synthetic entry from an already-measured elapsed_seconds instead
        of timestamping a fresh start/end pair."""
        if capture_quality not in ("true_completion", "accept_only"):
            raise RuntimeError(
                f"BLOCKED: timing capture_quality {capture_quality!r} must be 'true_completion' or 'accept_only'"
            )
        entry = {
            "segment_id": segment_id,
            "label": label,
            "capture_quality": capture_quality,
            "note": note,
            "start_utc": None,
            "end_utc": None,
            "elapsed_seconds": round(elapsed_seconds, 6),
        }
        entry.update(extra)
        self._append(entry)
        return entry

    def _append(self, entry):
        with self._lock:
            self.segments.append(entry)

    def write(self, path, meta=None):
        path = Path(path)
        path.parent.mkdir(parents=True, exist_ok=True)
        with self._lock:
            segments = list(self.segments)
        payload = {
            "tool": "net_video_timing_instrumentation",
            "capture_quality_legend": {
                "true_completion": (
                    "the elapsed number reflects when the real-world action genuinely "
                    "finished (blocking ack, confirmed poll, real LLM response, or a "
                    "deterministic sleep)"
                ),
                "accept_only": (
                    "the elapsed number only proves a command was accepted by Wire-Pod - "
                    "it does NOT prove the underlying speech or physical movement actually "
                    "finished; no completion signal exists in this stack for those calls"
                ),
            },
            "meta": meta or {},
            "segment_count": len(segments),
            "segments": segments,
        }
        path.write_text(json.dumps(payload, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
        status(f"TIMING_JSON_WRITTEN {path}")
        return path


class _TimingSpan:
    def __init__(self, log, segment_id, label, capture_quality, note, extra):
        self._log = log
        self._entry = {
            "segment_id": segment_id,
            "label": label,
            "capture_quality": capture_quality,
            "note": note,
        }
        if extra:
            self._entry.update(extra)

    def __enter__(self):
        self._entry["start_utc"] = (
            time.strftime("%Y-%m-%dT%H:%M:%S", time.gmtime()) + f".{int((time.time() % 1) * 1000):03d}Z"
        )
        self._start_monotonic = time.monotonic()
        return self._entry

    def __exit__(self, exc_type, exc, _tb):
        end_monotonic = time.monotonic()
        self._entry["end_utc"] = (
            time.strftime("%Y-%m-%dT%H:%M:%S", time.gmtime()) + f".{int((time.time() % 1) * 1000):03d}Z"
        )
        self._entry["elapsed_seconds"] = round(end_monotonic - self._start_monotonic, 6)
        if exc_type is not None:
            self._entry["error"] = repr(exc)
        self._log._append(self._entry)
        return False


TIMING = TimingLog()


def _speech_capture_quality(voice_destination):
    """The same three-way voice_destination switch Robots.say() itself
    branches on (see class Robots below) - the honest capture-quality tag
    depends entirely on which of the three real dispatch paths a turn takes,
    never a single fixed tag for "speech" in general."""
    if voice_destination == "robot":
        return (
            "accept_only",
            "Wire-Pod say_text is fire-and-forget - HTTP 200 confirms the command was "
            "accepted, not that the robot finished speaking. No completion signal exists "
            "in this stack for robot speech.",
        )
    if voice_destination == "monitor":
        return (
            "true_completion",
            "Kokoro monitor destination blocks on the speak-stdin worker's DONE ack, "
            "which only fires after real audio playback finishes.",
        )
    return (
        "true_completion",
        "voice_destination=off - no speech dispatched, elapsed reflects code path only "
        "(no real-world action to wait for).",
    )


def write_timing_log(run_id=None, meta=None, song_dir=None):
    run_id = run_id or utc_run_id()
    base_dir = Path(song_dir) if song_dir else DEFAULT_SECTION_SONG_DIR
    runs_dir = base_dir / "runs"
    path = runs_dir / f"net_video_timing_run_{run_id}.json"
    return TIMING.write(path, meta=meta)


def speak_turn_via_kokoro(text):
    _kokoro_status_ensure_loaded()
    proc = _KOKORO_STATUS_STATE["proc"]
    if proc is None or proc.stdin is None or proc.stdout is None or proc.poll() is not None:
        return {
            "engine": "kokoro_82m",
            "destination": "host_speaker",
            "spoken_text": text,
            "spoken": False,
            "error": "kokoro monitor voice engine has no live subprocess to speak through",
        }
    try:
        proc.stdin.write(text + "\n")
        proc.stdin.flush()
        ack = proc.stdout.readline()
        if not ack:
            raise RuntimeError("kokoro monitor speak worker closed stdout unexpectedly")
        return {
            "engine": "kokoro_82m",
            "destination": "host_speaker",
            "spoken_text": text,
            "spoken": True,
            "ack": ack.strip(),
        }
    except Exception as exc:  # noqa: BLE001 - reported to the caller, not swallowed.
        return {
            "engine": "kokoro_82m",
            "destination": "host_speaker",
            "spoken_text": text,
            "spoken": False,
            "error": repr(exc),
        }


def _kokoro_host4_ensure_loaded():
    if _KOKORO_HOST4_STATE["attempted"]:
        return
    _KOKORO_HOST4_STATE["attempted"] = True
    try:
        proc = subprocess.Popen(
            [
                str(KOKORO_STATUS_VENV_PYTHON),
                str(KOKORO_STATUS_ANNOUNCER_PATH),
                "--speak-stdin",
                "--voice",
                KOKORO_HOST4_VOICE,
            ],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            text=True,
            bufsize=1,
        )
        _KOKORO_HOST4_STATE["proc"] = proc
    except Exception as exc:  # noqa: BLE001 - a missing host 4 voice must never break the show.
        print(f"STATUS: KOKORO_HOST4_VOICE_UNAVAILABLE {type(exc).__name__}: {exc}", flush=True)


def speak_host4_line_via_kokoro(text):
    _kokoro_host4_ensure_loaded()
    proc = _KOKORO_HOST4_STATE["proc"]
    if proc is None or proc.stdin is None or proc.stdout is None or proc.poll() is not None:
        return {
            "engine": "kokoro_82m",
            "voice": KOKORO_HOST4_VOICE,
            "destination": "host_speaker",
            "spoken_text": text,
            "spoken": False,
            "error": "kokoro host4 voice engine has no live subprocess to speak through",
        }
    try:
        proc.stdin.write(text + "\n")
        proc.stdin.flush()
        ack = proc.stdout.readline()
        if not ack:
            raise RuntimeError("kokoro host4 speak worker closed stdout unexpectedly")
        return {
            "engine": "kokoro_82m",
            "voice": KOKORO_HOST4_VOICE,
            "destination": "host_speaker",
            "spoken_text": text,
            "spoken": True,
            "ack": ack.strip(),
        }
    except Exception as exc:  # noqa: BLE001 - reported to the caller, not swallowed.
        return {
            "engine": "kokoro_82m",
            "voice": KOKORO_HOST4_VOICE,
            "destination": "host_speaker",
            "spoken_text": text,
            "spoken": False,
            "error": repr(exc),
        }


def load_named_json_block(path, block_id):
    text = Path(path).read_text(encoding="utf-8")
    pattern = re.compile(rf"```json\s+{re.escape(block_id)}\s*\n(.*?)\n```", re.DOTALL)
    match = pattern.search(text)
    if not match:
        raise RuntimeError(f"BLOCKED: {block_id} block missing at {path}")
    return json.loads(match.group(1))


def load_interview_scaffold(scaffold_path=INTERVIEW_SCAFFOLD_PATH):
    try:
        path = Path(scaffold_path)
        if path.suffix.lower() in {".md", ".markdown"}:
            payload = load_named_json_block(path, "BROBOTS_INTERVIEW_RUNTIME_SCAFFOLD_001")
        else:
            payload = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        raise RuntimeError(f"BLOCKED: interview scaffold is not loadable at {scaffold_path}: {exc}") from exc
    if not isinstance(payload, dict):
        raise RuntimeError(f"BLOCKED: interview scaffold must be a JSON object at {scaffold_path}")
    missing = [key for key in REQUIRED_INTERVIEW_SCAFFOLD_KEYS if key not in payload]
    if missing:
        raise RuntimeError(f"BLOCKED: interview scaffold missing keys at {scaffold_path}: {', '.join(missing)}")
    return payload


_ANIMATION_VOCAB_TOKENS_CACHE = None
_ANIMATION_VOCAB_ALIASES_CACHE = None


def _load_animation_vocab_raw():
    try:
        return json.loads(GOPOD_ANIMATION_VOCAB_PATH.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        raise RuntimeError(
            f"BLOCKED: animation vocabulary is not loadable at {GOPOD_ANIMATION_VOCAB_PATH}: {exc}"
        ) from exc


def action_parameters(scaffold=None):
    global _ANIMATION_VOCAB_TOKENS_CACHE
    if _ANIMATION_VOCAB_TOKENS_CACHE is None:
        vocab = _load_animation_vocab_raw()
        clean = {
            str(tok.get("name", "")).strip()
            for tok in vocab.get("tokens", [])
            if tok.get("verified") and str(tok.get("name", "")).strip()
        }
        _ANIMATION_VOCAB_TOKENS_CACHE = clean or {"thinking"}
    return _ANIMATION_VOCAB_TOKENS_CACHE


def animation_aliases(scaffold=None):
    global _ANIMATION_VOCAB_ALIASES_CACHE
    if _ANIMATION_VOCAB_ALIASES_CACHE is None:
        vocab = _load_animation_vocab_raw()
        aliases = vocab.get("aliases", {})
        _ANIMATION_VOCAB_ALIASES_CACHE = {
            str(k).strip().lower(): str(v).strip()
            for k, v in aliases.items()
            if str(k).strip() and str(v).strip()
        }
    return _ANIMATION_VOCAB_ALIASES_CACHE


def resolve_animation_token(action, scaffold=None):
    valid = action_parameters(scaffold)
    candidate = str(action or "").strip()
    if candidate in valid:
        return candidate
    canonical = animation_aliases(scaffold).get(candidate.lower())
    if canonical in valid:
        return canonical
    return None


def resolve_animation_mode(mode_key, scaffold):
    registry = (scaffold or {}).get("animation_lane", {}).get("modes", {})
    entry = registry.get(mode_key)
    if entry is None:
        raise RuntimeError(f"BLOCKED: animation mode {mode_key!r} is not in the animation_lane.modes registry")
    if entry.get("readiness") != "ready":
        reason = entry.get("readiness_reason", "not marked ready")
        raise RuntimeError(f"BLOCKED: animation mode {mode_key!r} is not ready for use: {reason}")
    return entry


def resolve_animation_capability(capability_key, scaffold, required_for=None):
    registry = (scaffold or {}).get("animation_lane", {}).get("capabilities", {})
    entry = registry.get(capability_key)
    if entry is None:
        raise RuntimeError(
            f"BLOCKED: animation capability {capability_key!r} is not in the "
            "animation_lane.capabilities registry"
        )
    if entry.get("readiness") != "ready":
        reason = entry.get("readiness_reason", "not marked ready")
        raise RuntimeError(f"BLOCKED: animation capability {capability_key!r} is not ready for use: {reason}")
    if required_for and required_for not in (entry.get("wired_for") or []):
        raise RuntimeError(
            f"BLOCKED: animation capability {capability_key!r} is ready but not wired "
            f"for {required_for!r} (wired_for={entry.get('wired_for')})"
        )
    return entry


def resolve_channel(channel_key, scaffold, required_for=None):
    registry = (scaffold or {}).get("channels", {})
    entry = registry.get(channel_key)
    if entry is None:
        raise RuntimeError(f"BLOCKED: channel {channel_key!r} is not in the channels registry")
    if entry.get("readiness") != "ready":
        reason = entry.get("readiness_reason", "not marked ready")
        raise RuntimeError(f"BLOCKED: channel {channel_key!r} is not ready for use: {reason}")
    if required_for and required_for not in (entry.get("wired_for") or []):
        raise RuntimeError(
            f"BLOCKED: channel {channel_key!r} is ready but not wired for {required_for!r}"
        )
    return entry


def describe_channels_used(display_text):
    used = [{"channel": "robot_speaker_tts", "payload_variant": "spoken", "send_kind": "content"}]
    if display_text:
        used.append({"channel": "channel_1_rich_display", "payload_variant": "display", "send_kind": "content"})
    return used


MOVEMENT_ENDPOINT_AXIS = {
    "/api-sdk/move_lift": "lift",
    "/api-sdk/move_head": "head",
    "/api-sdk/move_wheels": "wheels",
}


def resolve_movement_axis(axis_key, scaffold, required_for=None):
    registry = (scaffold or {}).get("movements", {})
    entry = registry.get(axis_key)
    if entry is None:
        raise RuntimeError(f"BLOCKED: movement axis {axis_key!r} is not in the movements registry")
    if entry.get("readiness") != "ready":
        reason = entry.get("readiness_reason", "not marked ready")
        raise RuntimeError(f"BLOCKED: movement axis {axis_key!r} is not ready for use: {reason}")
    if required_for and required_for not in (entry.get("wired_for") or []):
        raise RuntimeError(
            f"BLOCKED: movement axis {axis_key!r} is ready but not wired for {required_for!r}"
        )
    return entry


def validate_movement_sequence_ends_at_zero(sequence):
    steps = [s for s in sequence if isinstance(s, dict)]
    if not steps:
        return
    last_speed = steps[-1].get("speed", 0)
    try:
        last_speed = float(last_speed)
    except (TypeError, ValueError):
        raise RuntimeError(
            f"BLOCKED: movement sequence's last step has a non-numeric speed: {last_speed!r}"
        )
    if last_speed != 0:
        raise RuntimeError(
            f"BLOCKED: movement sequence does not end at speed 0 (last step speed={last_speed}) - "
            "explicit_zero is enforced, not just a convention."
        )


def _load_control_song_module():
    spec = importlib.util.spec_from_file_location("run_robot_control_song_001", str(CONTROL_SONG_PATH))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def fire_scored_interview_movement(movement, serial, live, manage_control=True, segment_id="", segment_label=""):
    if not movement:
        return None
    control_song_mod = _load_control_song_module()
    self_mod = types.SimpleNamespace(
        validate_movement_sequence_ends_at_zero=validate_movement_sequence_ends_at_zero,
        wirepod_web_send_form=wirepod_web_send_form,
    )
    timing_steps = []
    segment_id = segment_id or f"movement_{movement}_{serial}"
    segment_label = segment_label or f"{movement} on {serial}"
    with TIMING.record(
        segment_id,
        segment_label,
        "accept_only",
        note=(
            "Whole-gesture wall-clock elapsed: sum of real HTTP accept-latencies plus "
            "asserted hold_seconds sleeps between waypoints. hold_seconds is asserted (the "
            "code sleeps this long), never a measured confirmation the robot's physical "
            "motion actually finished - see run_robot_control_song_001.py golden truth 5. "
            "Per-waypoint detail in 'movement_steps'."
        ),
    ) as span:
        if movement == "arm_gesture":
            result = control_song_mod.gentle_arm_gesture_cue(self_mod, serial, live, manage_control, timing_steps=timing_steps)
        elif movement == "head_nod_gesture":
            result = control_song_mod.head_nod_gesture_cue(self_mod, serial, live, manage_control, timing_steps=timing_steps)
        else:
            raise RuntimeError(f"BLOCKED: unknown interview movement {movement!r}, must be 'arm_gesture' or 'head_nod_gesture'")
        span["movement_steps"] = timing_steps
        span["live"] = live
    return result


def is_interview_kg_search_song(log):
    """True only for the interview song (INTERVIEW_KG_SEARCH_SONG_ID) - the
    one gate that keeps fire_interview_kg_search_sequence() below from ever
    reaching bingo/awaken/preshow/anything else. log["source_card"] is a
    path; only the trailing directory name is compared, same convention
    peek_song_bare_capture() and friends already use for song identity."""
    return Path(str(log.get("source_card", ""))).name == INTERVIEW_KG_SEARCH_SONG_ID


def find_section_card_line(log, line_number):
    """Looks up one line's own dict from log['section_card']['lines'] by
    line_number - the same source resolve_turn_mode() reads from during
    generation, reused here (not duplicated) so playback-time mode checks
    (canned vs llm) stay in sync with whatever generation actually decided.
    Returns {} if not found - a missing/unmatched line means "skip",  never
    a crash mid-playback."""
    for line in log.get("section_card", {}).get("lines", []):
        if line.get("line_number") == line_number:
            return line
    return {}


def fire_interview_kg_search_sequence(serial, live, segment_id="", segment_label=""):
    """Theatrical-only, interview song only: fires "searching" (looped every
    INTERVIEW_KG_SEARCH_LOOP_INTERVAL for INTERVIEW_KG_SEARCH_SECONDS) then
    "searchingGetout" (once), on the robot about to deliver an LLM-authored
    interview answer. Scripted in deliberately per the operator's own
    explicit ask, 2026-07-24 - the interview's JSON is fully pre-generated
    before playback_phase() ever starts (see generate_phase()/main()), so
    there is no real LLM wait here to mask; this exists purely to make a
    canned read *look* like a live search/answer. Same assume-loop-fire-once
    -release shape as ~/.gopod_alias_lib/brobots.sh's own
    _brobots_play_anim_single (outside this repo) - ported here, not
    reinvented; see ANIM_KNOWLEDGEGRAPH_RESOLVED_001.md/
    ANIM_CLARITY_FIREWORKS_STATUS_AND_INTEGRATION_IDEA_001.md (gopod_notes)
    for how searching/searchingGetout/answering were confirmed real and
    working. Does NOT touch the line's own animation_action - the caller
    (playback_phase()'s play_brobot_1()) overrides that to "answering"
    separately, right before play_generated_turn() fires the actual speech."""
    segment_id = segment_id or f"interview_kg_search_{serial}"
    segment_label = segment_label or f"Interview KG search sequence on {serial}"
    if not live:
        return {"ok": True, "mode": "dry", "results": []}
    results = []
    with TIMING.record(
        segment_id,
        segment_label,
        "accept_only",
        note=(
            "Theatrical animation sequence, not a real LLM wait - generation already "
            "finished before playback_phase() starts. Wall-clock elapsed here is "
            "deliberate added show time, asserted via INTERVIEW_KG_SEARCH_SECONDS, not a "
            "measured completion signal."
        ),
    ) as span:
        results.append(
            wirepod_web_send_form("/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10)
        )
        elapsed = 0.0
        while elapsed < INTERVIEW_KG_SEARCH_SECONDS:
            results.append(
                wirepod_web_send_form(
                    "/api-sdk/say_text", {"serial": serial, "text": "{{playAnimationWI||searching}}"}, timeout=10
                )
            )
            time.sleep(INTERVIEW_KG_SEARCH_LOOP_INTERVAL)
            elapsed += INTERVIEW_KG_SEARCH_LOOP_INTERVAL
        results.append(
            wirepod_web_send_form(
                "/api-sdk/say_text", {"serial": serial, "text": "{{playAnimationWI||searchingGetout}}"}, timeout=10
            )
        )
        results.append(
            wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)
        )
        span["live"] = live
    return {"ok": all(r.get("status") == 200 for r in results), "results": results}


def thought_cycles(scaffold):
    return scaffold.get("thought_cycles", {})


def read_field(lines, label):
    wanted = label.upper() + ":"
    for index, line in enumerate(lines):
        if line.strip().upper() == wanted:
            values = []
            for value in lines[index + 1 :]:
                stripped = value.strip()
                if not stripped:
                    if values:
                        break
                    continue
                if stripped == "---" or stripped.startswith("LINE ") or stripped.endswith(":"):
                    break
                values.append(stripped)
            return " ".join(values).strip()
    return ""


_LIST_ITEM_NUMBER_PREFIX = re.compile(r"^\d+[.)]\s*")


def read_list_field(lines, label):
    wanted = label.upper() + ":"
    for index, line in enumerate(lines):
        if line.strip().upper() == wanted:
            items = []
            for value in lines[index + 1 :]:
                stripped = value.strip()
                if not stripped:
                    if items:
                        break
                    continue
                if stripped == "---" or stripped.startswith("LINE ") or stripped.endswith(":"):
                    break
                items.append(_LIST_ITEM_NUMBER_PREFIX.sub("", stripped))
            return items
    return []


def read_persona_mindsets(lines):
    mindsets = {"brobot_2": "", "brobot_1": ""}
    current = ""
    for line in lines:
        stripped = line.strip()
        upper = stripped.upper()
        if upper == "INTERVIEWER/BROBOT 2":
            current = "brobot_2"
            continue
        if upper == "INTERVIEWEE/BROBOT 1":
            current = "brobot_1"
            continue
        if upper.startswith("MINDSET:") and current:
            mindsets[current] = stripped.split(":", 1)[1].strip()
    return mindsets


CYCLE_BARE_LABEL_DEFAULT_WEIGHT = 3


def parse_cycle_weights(raw):
    weights = {}
    for token in raw.split(","):
        token = token.strip()
        if not token:
            continue
        label, sep, temp = token.partition(":")
        label = label.strip()
        if not label:
            continue
        if sep:
            try:
                weights[label] = int(temp.strip())
            except ValueError:
                weights[label] = CYCLE_BARE_LABEL_DEFAULT_WEIGHT
        else:
            weights[label] = CYCLE_BARE_LABEL_DEFAULT_WEIGHT
    return weights


def load_section_card(card_path=CARD_PATH):
    lines = Path(card_path).read_text(encoding="utf-8").splitlines()
    mindsets = read_persona_mindsets(lines)
    starts = [index for index, line in enumerate(lines) if re.match(r"^LINE\s+\d+\s*$", line.strip()) or re.match(r"^LINE\s+\d+\s+-\s+", line.strip())]
    section = {
        "card_path": str(card_path),
        "section_id": read_field(lines, "SECTION ID"),
        "brobot_2_mindset": mindsets["brobot_2"],
        "brobot_1_mindset": mindsets["brobot_1"],
        "section_title": read_field(lines, "SECTION TITLE"),
        "section_purpose": read_field(lines, "SECTION PURPOSE"),
        "audience_takeaway": read_field(lines, "AUDIENCE TAKEAWAY"),
        "section_success": read_field(lines, "SECTION SUCCESS"),
        "lines": [],
    }
    for position, start in enumerate(starts):
        header = lines[start].strip()
        match = re.match(r"LINE\s+(\d+)(?:\s+-\s+([A-Z_]+))?", header)
        if not match:
            continue
        end = starts[position + 1] if position + 1 < len(starts) else len(lines)
        block = lines[start:end]
        line_type = match.group(2) or read_field(block, "TYPE")
        visible = (
            read_field(block, "VISIBLE LINE")
            or read_field(block, "VISIBLE LINE LINE")
            or read_field(block, "CHECKPOINT LINE")
            or read_field(block, "CONTEXT LINE")
        )
        cycle_weights = parse_cycle_weights(read_field(block, "CYCLES"))
        silent_value_cue = read_field(block, "SILENT VALUE CUE") or read_field(block, "VALUE TARGET")
        value_points = read_list_field(block, "VALUE POINTS") or ([silent_value_cue] if silent_value_cue else [])
        raw_interviewer_count = read_field(block, "INTERVIEWER VALUE POINT COUNT")
        try:
            interviewer_value_point_count = int(raw_interviewer_count) if raw_interviewer_count else None
        except ValueError:
            interviewer_value_point_count = None
        section["lines"].append(
            {
                "line_number": int(match.group(1)),
                "line_type": line_type,
                "exchange_type": read_field(block, "EXCHANGE TYPE"),
                "brobot_2_mindset": mindsets["brobot_2"],
                "brobot_1_mindset": mindsets["brobot_1"],
                "interviewer_emotional_beat": read_field(block, "INTERVIEWER EMOTIONAL BEAT"),
                "interviewee_emotional_beat": read_field(block, "INTERVIEWEE EMOTIONAL BEAT"),
                "interview_arc_point": read_field(block, "INTERVIEW ARC POINT"),
                "default_speaker": read_field(block, "DEFAULT SPEAKER"),
                "visible_line": visible,
                "silent_value_cue": silent_value_cue,
                "value_points": value_points,
                "interviewer_value_point_count": interviewer_value_point_count,
                "tone": read_field(block, "TONE"),
                "purpose": read_field(block, "PURPOSE") or read_field(block, "CHECKPOINT PURPOSE") or read_field(block, "CONTEXT PURPOSE"),
                "success": read_field(block, "SUCCESS"),
                "thinking_window": read_field(block, "THINKING WINDOW").strip().lower() == "yes",
                "cycles": [label for label, weight in cycle_weights.items() if weight > 0],
                "cycle_weights": cycle_weights,
            }
        )
    return section


def parse_section_1(card_path=CARD_PATH):
    return load_section_card(card_path)["lines"]


_STORY_LINE_HEADING = re.compile(r"^##\s+LINE\s+(\d+)\s*$")
_STORY_ARC_POINT = re.compile(r"^Arc point:\s*(.+)$")
_STORY_NUMBERED_ITEM = re.compile(r"^\d+\.\s*(.+)$")
_STORY_PERSONA_PREFIXES = {
    "**Interviewer / Brobot 2**": "brobot_2_mindset",
    "**Interviewee / Brobot 1**": "brobot_1_mindset",
}
_STORY_FRAMING_FIELDS = {
    "Title": "section_title",
    "Purpose": "section_purpose",
    "Audience takeaway": "audience_takeaway",
    "Section success": "section_success",
}


def parse_story_md(text):
    lines = text.splitlines()
    story = {
        "brobot_2_mindset": "",
        "brobot_1_mindset": "",
        "section_title": "",
        "section_purpose": "",
        "audience_takeaway": "",
        "section_success": "",
        "steps": [],
    }
    for raw in lines:
        stripped = raw.strip()
        for prefix, key in _STORY_PERSONA_PREFIXES.items():
            if stripped.startswith(prefix):
                story[key] = stripped.split("—", 1)[-1].strip()
        if stripped.startswith("- "):
            label, sep, value = stripped[2:].partition(":")
            field = _STORY_FRAMING_FIELDS.get(label.strip())
            if field and sep:
                value = value.strip()
                if value.startswith("*(") and value.endswith(")*"):
                    value = ""  # italic placeholder, e.g. "*(not yet authored)*"
                story[field] = value

    starts = [i for i, line in enumerate(lines) if _STORY_LINE_HEADING.match(line.strip())]
    for position, start in enumerate(starts):
        number = int(_STORY_LINE_HEADING.match(lines[start].strip()).group(1))
        end = starts[position + 1] if position + 1 < len(starts) else len(lines)
        block = lines[start:end]
        arc_point = ""
        visible_line = ""
        value_points = []
        for raw in block:
            stripped = raw.strip()
            arc_match = _STORY_ARC_POINT.match(stripped)
            if arc_match:
                arc_point = arc_match.group(1).strip()
            elif stripped.startswith(">") and not visible_line:
                visible_line = stripped.lstrip(">").strip()
            else:
                item_match = _STORY_NUMBERED_ITEM.match(stripped)
                if item_match:
                    value_points.append(item_match.group(1).strip())
        story["steps"].append({
            "step_id": f"line_{number}",
            "arc_point": arc_point,
            "visible_line": visible_line,
            "value_points": value_points,
        })
    return story


def peek_song_bare_capture(section_source, use_legacy_card):
    """Tiny, targeted peek at a song's own knobs.json top-level
    "bare_capture" flag - reads only that one key, before generate_phase()'s
    own full section_card load (which stays at its original point, unchanged,
    for status-line ordering). Needed this early because the model-picker/
    Kokoro-warmup decisions have to happen before anything else runs. Legacy
    .txt card mode (no knobs.json at all) and any read/parse failure both
    fall back to False - "today's behavior" is always the safe default."""
    if use_legacy_card:
        return False
    try:
        knobs_path = Path(section_source) / "zKnobs.json"
        if not knobs_path.exists():
            knobs_path = Path(section_source) / "knobs.json"
        return bool(json.loads(knobs_path.read_text(encoding="utf-8")).get("bare_capture", False))
    except (OSError, json.JSONDecodeError):
        return False


def load_section_pair(song_dir):
    song_dir = Path(song_dir)
    knobs_path = song_dir / "zKnobs.json"
    if not knobs_path.exists():
        knobs_path = song_dir / "knobs.json"
    story_path = song_dir / "story.md"
    knobs = json.loads(knobs_path.read_text(encoding="utf-8"))
    story = parse_story_md(story_path.read_text(encoding="utf-8"))
    story_steps = {step["step_id"]: step for step in story["steps"]}
    section = {
        "card_path": str(song_dir),
        "section_id": knobs.get("section_id", ""),
        "brobot_2_mindset": story["brobot_2_mindset"],
        "brobot_1_mindset": story["brobot_1_mindset"],
        "section_title": story["section_title"],
        "section_purpose": story["section_purpose"],
        "audience_takeaway": story["audience_takeaway"],
        "section_success": story["section_success"],
        "first_speaker": knobs.get("first_speaker", "brobot_2"),
        "movement_pause_seconds": knobs.get("movement_pause_seconds", 0) or 0,
        "lines": [],
    }
    for step in knobs.get("steps", []):
        step_id = step.get("step_id")
        if step_id not in story_steps:
            raise RuntimeError(
                f"BLOCKED: knobs.json step_id {step_id!r} has no matching story.md "
                f"block at {story_path}"
            )
        content = story_steps[step_id]
        cycle_weights = step.get("cycle_weights") or {}
        section["lines"].append(
            {
                "line_number": int(str(step_id).rsplit("_", 1)[-1]),
                "line_type": step.get("line_type", ""),
                "exchange_type": step.get("exchange_type", ""),
                "brobot_2_mindset": story["brobot_2_mindset"],
                "brobot_1_mindset": story["brobot_1_mindset"],
                "interviewer_emotional_beat": step.get("interviewer_emotional_beat", ""),
                "interviewee_emotional_beat": step.get("interviewee_emotional_beat", ""),
                "interview_arc_point": content["arc_point"],
                "default_speaker": "",
                "visible_line": content["visible_line"],
                "silent_value_cue": "",
                "value_points": content["value_points"],
                "interviewer_value_point_count": step.get("interviewer_value_point_count"),
                "tone": "",
                "purpose": "",
                "success": "",
                "thinking_window": bool(step.get("thinking_window", False)),
                "cycles": [label for label, weight in cycle_weights.items() if weight > 0],
                "cycle_weights": cycle_weights,
                "brobot_2_mode": step.get("brobot_2_mode"),
                "brobot_1_mode": step.get("brobot_1_mode"),
                "brobot_2_canned_source": step.get("brobot_2_canned_source"),
                "brobot_1_canned_source": step.get("brobot_1_canned_source"),
                "brobot_2_opening_task": step.get("brobot_2_opening_task"),
                "brobot_2_movement": step.get("brobot_2_movement"),
                "brobot_1_movement": step.get("brobot_1_movement"),
                "brobot_2_movement_2": step.get("brobot_2_movement_2"),
                "brobot_1_movement_2": step.get("brobot_1_movement_2"),
            }
        )
    return section


_PRESHOW_BEAT_HEADING = re.compile(r"^##\s+BEAT\s+(\S+)\s*$")


def parse_preshow_story_md(text):
    lines = text.splitlines()
    starts = [i for i, line in enumerate(lines) if _PRESHOW_BEAT_HEADING.match(line.strip())]
    beats = []
    for position, start in enumerate(starts):
        step_id = _PRESHOW_BEAT_HEADING.match(lines[start].strip()).group(1)
        end = starts[position + 1] if position + 1 < len(starts) else len(lines)
        block = lines[start + 1 : end]
        text_val = ""
        hint_val = ""
        for raw in block:
            stripped = raw.strip()
            if stripped.upper().startswith("> TEXT:"):
                text_val = stripped[len("> TEXT:"):].strip()
            elif stripped.upper().startswith("> HINT:"):
                hint_val = stripped[len("> HINT:"):].strip()
        beats.append({"step_id": step_id, "text": text_val, "hint": hint_val})
    return {"beats": beats}


def load_preshow_song(song_dir):
    song_dir = Path(song_dir)
    knobs_path = song_dir / "zKnobs.json"
    if not knobs_path.exists():
        knobs_path = song_dir / "knobs.json"
    knobs = json.loads(knobs_path.read_text(encoding="utf-8"))
    story = parse_preshow_story_md((song_dir / "story.md").read_text(encoding="utf-8"))
    story_beats = {beat["step_id"]: beat for beat in story["beats"]}
    steps = []
    for step in knobs.get("steps", []):
        step_id = step.get("step_id")
        if step_id not in story_beats:
            raise RuntimeError(
                f"BLOCKED: preshow knobs.json step_id {step_id!r} has no matching story.md "
                f"beat at {song_dir / 'story.md'}"
            )
        content = story_beats[step_id]
        steps.append(
            {
                "step_id": step_id,
                "movement": step.get("movement", ""),
                "speaker": step.get("speaker", ""),
                "mode": step.get("mode", "canned"),
                "text": content["text"],
                "hint": content["hint"],
            }
        )
    return {"song_id": knobs.get("song_id", ""), "steps": steps}


def is_crystallizer(line):
    line_type = (line.get("line_type") or "").upper()
    return "STATEMENT" in line_type or "CRYSTALLIZER" in line_type


def selected_thinking_cycles(line, scaffold):
    weights = line.get("cycle_weights") or {}
    pool = {c.get("label"): c for c in thought_cycles(scaffold).get("cycles", [])}
    selected = [
        (label, weight, pool[label].get("focus", ""))
        for label, weight in weights.items()
        if weight > 0 and label in pool
    ]
    selected.sort(key=lambda item: item[1], reverse=True)
    return selected


DEFAULT_INTERVIEWER_VALUE_POINT_COUNT = 2


def interviewer_value_point_count(line):
    points = line.get("value_points") or []
    configured = line.get("interviewer_value_point_count")
    if configured is None:
        count = DEFAULT_INTERVIEWER_VALUE_POINT_COUNT
    else:
        count = configured
    return max(0, min(count, len(points)))


def selected_interviewer_value_points(line):
    points = line.get("value_points") or []
    return points[: interviewer_value_point_count(line)]


def build_exchange_prompt(role, mindset, line, value_target="", scaffold=None):
    scaffold = scaffold or load_interview_scaffold()
    line_modes = scaffold.get("line_modes", {})
    line_mode = line_modes.get("statement", "STATEMENT / CRYSTALLIZER") if is_crystallizer(line) else line_modes.get("question", "QUESTION")
    line_number = int(line.get("line_number") or 0)
    emotional_beat = (
        line.get("interviewer_emotional_beat", "")
        if role == "BROBOT 2"
        else line.get("interviewee_emotional_beat", "")
    )
    role_tasks = scaffold.get("role_tasks", {}).get(role, {})
    if role == "BROBOT 2":
        opening_override = line.get("brobot_2_opening_task")
        use_opening_task = opening_override if opening_override is not None else (line_number == 1)
        if use_opening_task:
            task = role_tasks.get("line_1_task", "")
            extra = role_tasks.get("line_1_rule", "")
        else:
            task = role_tasks.get("statement_task", "") if is_crystallizer(line) else role_tasks.get("question_task", "")
            extra = role_tasks.get("default_rule", "")
    else:
        task = role_tasks.get("statement_task", "") if is_crystallizer(line) else role_tasks.get("question_task", "")
        extra = role_tasks.get("statement_rule", "") if is_crystallizer(line) else role_tasks.get("question_rule", "")

    exchange_type_entry = scaffold.get("exchange_types", {}).get(line.get("exchange_type", ""), {})
    exchange_type_rule = exchange_type_entry.get("brobot_2_rule", "") if role == "BROBOT 2" else exchange_type_entry.get("brobot_1_rule", "")
    prompt_templates = scaffold.get("prompt_templates", {})

    parts = [
        f"ROLE: {role}",
        f"MINDSET: {mindset}",
        f"LINE MODE: {line_mode}",
        f"EMOTIONAL BEAT: {emotional_beat}",
        f"INTERVIEW ARC POINT: {line.get('interview_arc_point', '')}",
        f"TASK: {task}",
    ]
    if role == "BROBOT 2":
        parts.append(f"VISIBLE LINE: {line['visible_line']}")
    else:
        visible_line_label = prompt_templates.get(
            "interviewee_visible_line_label",
            "INTERVIEWER'S LINE - CONTEXT ONLY, never repeat, quote, or restate any part of "
            "it, in whole or in part, before or after your answer. Your reply must not "
            "contain these words: ",
        )
        parts.append(f"{visible_line_label}{line['visible_line']}")
    if value_target:
        parts.append(value_target)
    if extra:
        parts.append(f"RULE: {extra}")
    if exchange_type_rule:
        parts.append(f"RULE: {exchange_type_rule}")
    parts.extend(f"RULE: {rule}" for rule in scaffold.get("shared_rules", []) if str(rule).strip())

    if role == "BROBOT 2":
        interviewer_points = selected_interviewer_value_points(line)
        if interviewer_points:
            parts.append(
                prompt_templates.get(
                    "interviewer_think_intro",
                    "INTERVIEWER THINK: Here are the locked value points I must pull out, in "
                    "this order. The runner has told me how many this question reaches for. I "
                    "choose only ladder or fan, and the best wording, to make Brobot 1 reveal "
                    "them naturally for the audience watching.",
                )
            )
            parts.append(
                prompt_templates.get(
                    "interviewer_value_points_header",
                    "INTERVIEWER VALUE POINTS (locked, ordered - do not invent, add, drop, or reorder):",
                )
            )
            parts.extend(f"{index}. {point}" for index, point in enumerate(interviewer_points, 1))
            reach_template = prompt_templates.get(
                "interviewer_reach_count_template",
                "THIS QUESTION REACHES FOR EXACTLY {count} OF THE POINTS ABOVE. The runner "
                "has set this count; you do not choose it.",
            )
            parts.append(reach_template.format(count=len(interviewer_points)))
            parts.append(
                prompt_templates.get(
                    "question_shape_dial",
                    "QUESTION SHAPE DIAL: choose exactly one of these two shapes, nothing else - "
                    "LADDER (stay on the same point(s), go deeper for more detail) or FAN (spread "
                    "across the different points above, inviting a wider answer). Then pick the "
                    "wording that best fits. Do not reveal, answer, or explain the points yourself "
                    "- only ask toward them.",
                )
            )

    thinking_cycles = selected_thinking_cycles(line, scaffold)
    if line.get("thinking_window") and thinking_cycles:
        parts.append(
            prompt_templates.get(
                "thinking_window_intro",
                "THINKING WINDOW: Before replying, silently weigh these focuses in order "
                "(higher weight = more emphasis). Never show this reasoning to the audience.",
            )
        )
        parts.extend(f"- {label} (weight {weight}): {focus}" for label, weight, focus in thinking_cycles)
        parts.append(
            prompt_templates.get(
                "thinking_window_output_format",
                "OUTPUT FORMAT FOR THIS TURN: output exactly two labeled sections in this order - "
                "a line reading exactly \"THOUGHTS:\" followed by your brief internal reasoning through "
                "the focuses above (1-3 short sentences, never spoken aloud), then a line reading "
                "exactly \"REPLY:\" followed by your actual turn in the normal actionParameter... "
                "spoken shape.",
            )
        )

    return "\n".join(parts)


def load_api_config():
    with API_CONFIG.open("r", encoding="utf-8") as handle:
        return json.load(handle).get("knowledge", {})


def list_installed_models():
    api = load_api_config()
    endpoint = api.get("endpoint", "")
    if not endpoint:
        raise RuntimeError("BLOCKED: knowledge endpoint is empty, cannot reach Ollama for the model list")
    parsed = urllib.parse.urlsplit(endpoint)
    tags_url = f"{parsed.scheme}://{parsed.netloc}/api/tags"
    try:
        with urllib.request.urlopen(urllib.request.Request(tags_url, method="GET"), timeout=10) as response:
            data = json.loads(response.read().decode("utf-8"))
    except Exception as exc:
        raise RuntimeError(f"BLOCKED: could not reach Ollama at {tags_url} for the model list: {exc!r}") from exc
    models = data.get("models") or []
    if not models:
        raise RuntimeError(f"BLOCKED: Ollama at {tags_url} reports zero installed models")
    return [{"name": m.get("name", ""), "size": m.get("size", 0)} for m in models]


def read_remembered_model():
    try:
        data = json.loads(CONTENT_MODEL_STATE_PATH.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return ""
    return data.get("model", "")


def write_remembered_model(model_name):
    CONTENT_MODEL_STATE_PATH.parent.mkdir(parents=True, exist_ok=True)
    CONTENT_MODEL_STATE_PATH.write_text(
        json.dumps(
            {"model": model_name, "updated_utc": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())},
            indent=2,
        )
        + "\n",
        encoding="utf-8",
    )


def resolve_content_model():
    models = list_installed_models()
    installed_names = [m["name"] for m in models]

    env_model = os.getenv(CONTENT_MODEL_ENV, "").strip()
    if env_model:
        if env_model not in installed_names:
            raise RuntimeError(
                f"BLOCKED: {CONTENT_MODEL_ENV}={env_model!r} is not an installed model. "
                f"Installed: {installed_names}"
            )
        write_remembered_model(env_model)
        return env_model

    remembered = read_remembered_model()
    default_model = remembered if remembered in installed_names else DEFAULT_CONTENT_MODEL
    if default_model not in installed_names:
        raise RuntimeError(
            f"BLOCKED: default model {DEFAULT_CONTENT_MODEL!r} is not installed and no valid "
            f"remembered model was found. Installed: {installed_names}"
        )

    if not os.isatty(0):
        write_remembered_model(default_model)
        return default_model

    print("Installed models (live from Ollama):")
    for index, entry in enumerate(models, 1):
        marker = "  <- remembered default" if entry["name"] == default_model else ""
        size_mb = entry["size"] / (1024 * 1024)
        print(f"  {index}. {entry['name']}  ({size_mb:.0f} MB){marker}")
    while True:
        raw = input(f"Pick a model number [Enter = {default_model}]: ").strip()
        if not raw:
            chosen = default_model
            break
        if raw.isdigit() and 1 <= int(raw) <= len(models):
            chosen = models[int(raw) - 1]["name"]
            break
        print(f"Not a valid choice - enter a number from 1 to {len(models)}, or press Enter.")
    write_remembered_model(chosen)
    return chosen


def llm_colour_line(role, line, value_target="", prior_context="", scaffold=None, content_model=None):
    scaffold = scaffold or load_interview_scaffold()
    api = load_api_config()
    endpoint = api.get("endpoint", "")
    if not endpoint:
        raise RuntimeError("knowledge endpoint is empty")
    line_number = int(line.get("line_number") or 0)
    status(f"LLM_START line={line_number} role={role}")
    prompt = build_exchange_prompt(
        role,
        line.get("brobot_2_mindset", "") if role == "BROBOT 2" else line.get("brobot_1_mindset", ""),
        line,
        value_target,
        scaffold,
    )
    if prior_context:
        prior_context_label = (scaffold.get("prompt_templates", {}) or {}).get(
            "prior_context_label",
            "PRIOR EXCHANGE CONTEXT - background only, what already happened earlier in the "
            "interview. Never repeat, quote, continue, or build your own reply out of this "
            "text; it is not a prompt to extend:",
        )
        prompt += f"\n{prior_context_label}\n" + prior_context
    payload = {
        "model": content_model or api.get("model") or "gemma2:2b",
        "messages": [
            {
                "role": "system",
                "content": scaffold["system_prompt"],
            },
            {"role": "user", "content": prompt},
        ],
        "temperature": api.get("temp", 0) or 0,
        "top_p": api.get("top_p", 1) or 1,
        "stream": False,
    }
    request = urllib.request.Request(
        endpoint.rstrip("/") + "/chat/completions",
        data=json.dumps(payload).encode("utf-8"),
        headers={"Content-Type": "application/json", "Authorization": "Bearer " + str(api.get("key", ""))},
        method="POST",
    )
    try:
        with urllib.request.urlopen(request, timeout=120) as response:
            data = json.loads(response.read().decode("utf-8"))
    except Exception:
        status(f"LLM_BLOCKED line={line_number} role={role}")
        raise
    choices = data.get("choices") or []
    if not choices:
        raise RuntimeError("empty LLM response")
    content = choices[0].get("message", {}).get("content", "").strip()
    status(f"LLM_DONE line={line_number} role={role} chars={len(content)}")
    return content


def default_action_parameter(scaffold=None):
    scaffold = scaffold or {}
    return scaffold.get("generation_defaults", {}).get("default_action_parameter", "thinking")


def robot_safe(text, action="thinking", scaffold=None):
    fallback_action = default_action_parameter(scaffold)
    action = action if action in action_parameters(scaffold) else fallback_action
    text = re.sub(r"\{\{[^}]*\}\}", "", " ".join(str(text).split())).strip()
    if not text:
        raise ValueError("empty robot speech is not allowed in live Section 1")
    return f"{action}... {text}"


def _load_speech_math_flatten():
    try:
        path = Path(__file__).resolve().parent / "speech_math_flatten_001.py"
        spec = importlib.util.spec_from_file_location("speech_math_flatten_001", str(path))
        mod = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(mod)
        return mod.flatten_math_for_speech
    except Exception as exc:  # noqa: BLE001 - never break speech over a missing helper
        print(f"SPEECH_MATH_FLATTEN_UNAVAILABLE {type(exc).__name__}: {exc}", flush=True)
        return lambda text: text


flatten_math_for_speech = _load_speech_math_flatten()


def strip_rich_text_for_voice(raw, scaffold=None):
    rules = cleanup_rules(scaffold)
    text = flatten_math_for_speech(str(raw))
    emoji_class = codepoint_class(rules.get("emoji_codepoint_ranges", []))
    if emoji_class:
        text = re.sub(f"[{emoji_class}]", " ", text)
    markdown_chars = str(rules.get("markdown_chars", ""))
    if markdown_chars:
        text = re.sub(f"[{re.escape(markdown_chars)}]+", " ", text)
    return text


def normalize_robot_safe(raw, scaffold=None):
    valid_actions = action_parameters(scaffold)
    raw = re.sub(r"\|\|\|[01C]\|\|\|", "", raw.strip())
    raw = re.sub(r"\{\{[^}]*\}\}", "", raw)
    raw = re.sub(r"(?i)\bVISIBLE LINE:\s*", "", raw)
    raw = re.sub(r"(?i)\bREVEAL IN RESPONSE:\s*", "", raw)
    raw = re.sub(r"(?i)\bINTERVIEWER LINE:\s*", "", raw)
    raw = re.sub(r"(?i)\bVALUE TARGET:\s*", "", raw)
    raw = strip_rich_text_for_voice(raw, scaffold)
    action_pattern = "|".join(re.escape(action) for action in sorted(valid_actions, key=len, reverse=True))
    if re.match(rf"^({action_pattern})\.\.\. .+", raw):
        return raw, "none"
    cleaned = " ".join(raw.split())
    cleaned = re.sub(r"^actionParameter\s+", "", cleaned, flags=re.IGNORECASE).strip()
    if "..." in cleaned:
        action, spoken = cleaned.split("...", 1)
        resolved = resolve_animation_token(action.strip(), scaffold)
        if resolved:
            return robot_safe(spoken, resolved, scaffold), "deterministic"
    return robot_safe(cleaned, default_action_parameter(scaffold), scaffold), "deterministic"


def split_robot_safe_line(robot_safe_line, scaffold=None):
    if "..." in robot_safe_line:
        action, spoken = robot_safe_line.split("...", 1)
        resolved = resolve_animation_token(action.strip(), scaffold)
        if resolved:
            return resolved, spoken.strip()
    return default_action_parameter(scaffold), robot_safe_line.strip()


def load_pronunciation_registry(registry_path=SPEECH_PRONUNCIATION_REGISTRY_PATH):
    path = Path(registry_path)
    try:
        if path.suffix.lower() in {".md", ".markdown"}:
            return load_named_json_block(path, "SPEECH_PRONUNCIATION_REGISTRY_001")
        return json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return {}


def pronunciation_entries(registry_path=SPEECH_PRONUNCIATION_REGISTRY_PATH):
    payload = load_pronunciation_registry(registry_path)
    entries = payload.get("pronunciation_entries", []) if isinstance(payload, dict) else []
    if not isinstance(entries, list):
        entries = []
    clean_entries = []
    for entry in entries:
        if not isinstance(entry, dict):
            continue
        source = str(entry.get("from", ""))
        target = str(entry.get("to", ""))
        if source and target:
            clean_entries.append({"from": source, "to": target, "match_case": bool(entry.get("match_case")),
                                  "word_start": bool(entry.get("word_start")),
                                  "except": [str(x).lower() for x in entry.get("except") or []]})
    clean_entries.sort(key=lambda item: len(item["from"]), reverse=True)
    return clean_entries


def starred_pattern(word, vowels=""):
    """A word quoted for a regex, each "*" its own vowel from vowels, in order (or
    a "*" itself) - a public, starred entry ("sh*t", "i") matches the word said and
    the word written starred, never another word ("shot") (2026-10-29, operator:
    "Public GOPOD: star '*' out expletives"; gopod_pronunciation.go
    gopodStarredPattern, the same rule)."""
    left = list(vowels or "")
    def one(_m):
        if not left:
            return r"\*"
        v = left.pop(0)
        return "[" + v.lower() + v.upper() + r"\*]"
    return re.sub(r"\\\*", one, re.escape(word))


def apply_pronunciation_safety(text):
    speech = str(text or "")
    for entry in pronunciation_entries():
        source = entry["from"]
        target = entry["to"]
        flags = 0 if entry.get("match_case") else re.IGNORECASE
        if entry.get("word_start"):
            pattern = re.compile(rf"(?<!\w){starred_pattern(source, entry.get('vowels', ''))}(\w*)", flags)
            skip = entry.get("except") or []
            speech = pattern.sub(lambda m: m.group(0) if m.group(0).lower() in skip else target + m.group(1), speech)
            continue
        pattern = re.compile(rf"(?<!\w){starred_pattern(source, entry.get('vowels', ''))}(?!\w)", flags)
        speech = pattern.sub(target, speech)
    return speech


def soften_trigger_phrase_punctuation(text, scaffold=None):
    scaffold = scaffold or load_interview_scaffold()
    phrases = scaffold.get("trigger_phrase_punctuation_soften", [])
    if not isinstance(phrases, list):
        phrases = []
    softened = str(text or "")
    for phrase in phrases:
        terminal_pattern = re.compile(rf"{re.escape(phrase)}[!?]+['\"]?\s*$", re.IGNORECASE)
        if terminal_pattern.search(softened):
            continue
        softened = re.sub(
            rf"(['\"]?)({re.escape(phrase)})([!?]+)(['\"]?)",
            r"\1\2\4",
            softened,
            flags=re.IGNORECASE,
        )
    return softened


def clean_inline_emphasis_punctuation(text, scaffold=None):
    scaffold = scaffold or load_interview_scaffold()
    rule = scaffold.get("inline_emphasis_punctuation_rule", {})
    if not rule.get("enabled", False):
        return text
    cleaned = re.sub(r"\s+([!?])", r"\1", str(text))
    sentence_break_next = rule.get("sentence_break_next_pattern", "^[A-Z0-9]")

    def replace(match):
        punct = match.group(1)
        next_char = match.group(2)
        if re.match(sentence_break_next, next_char):
            return f"{punct} {next_char}"
        if rule.get("drop_punctuation_before_lowercase_continuation", True):
            return f" {next_char}"
        return f"{punct} {next_char}"

    return re.sub(r"([!?])\s+(\S)", replace, cleaned)


def cleanup_rules(scaffold=None):
    scaffold = scaffold or load_interview_scaffold()
    rules = scaffold.get("speech_cleanup_rules", {})
    return rules if isinstance(rules, dict) else {}


def codepoint_class(items):
    pieces = []
    for item in items:
        raw = str(item)
        if "-" in raw:
            start, end = raw.split("-", 1)
            pieces.append(f"\\U{int(start, 16):08X}-\\U{int(end, 16):08X}")
        else:
            pieces.append(f"\\u{int(raw, 16):04X}" if len(raw) <= 4 else f"\\U{int(raw, 16):08X}")
    return "".join(pieces)


def remove_configured_label_prefixes(text, labels):
    cleaned = text
    for label in labels:
        cleaned = re.sub(rf"(?i)\b{re.escape(str(label))}:\s*", " ", cleaned)
    return cleaned


def remove_configured_bracket_labels(text, labels):
    cleaned = text
    for label in labels:
        cleaned = re.sub(rf"(?i)\*?\[?\s*{re.escape(str(label))}\s*\]?\*?", " ", cleaned)
    return cleaned


_VIETNAMESE_KEEP_MARKS = {
    "̀",  # Dấu huyền
    "́",  # Dấu sắc
    "̃",  # Dấu ngã
    "̉",  # Dấu hỏi
    "̣",  # Dấu nặng
    "̂",  # Dấu mũ (â, ê, ô)
    "̛",  # Dấu ơ và ư
    "̆",  # Dấu trầm
}


def strip_diacritics_except_vietnamese(text):
    """Mirrors kgsim.go's isMn()/removeSpecialCharacters() diacritic pass
    exactly: NFD-decompose, drop non-spacing marks except the same 8
    Vietnamese tonal/diacritic codepoints that Go function explicitly
    keeps, NFC-recompose. Same exception list, same algorithm - this is
    the Go truth's own choice, not a Python reinterpretation of it."""
    decomposed = unicodedata.normalize("NFD", str(text))
    kept = "".join(
        ch for ch in decomposed if not (unicodedata.combining(ch) and ch not in _VIETNAMESE_KEEP_MARKS)
    )
    return unicodedata.normalize("NFC", kept)


GO_TRUTH_UNIVERSAL_CHARACTER_REPLACEMENTS = [
    ["‘", "'"], ["’", "'"],          # ' ' -> '
    ["“", "\""], ["”", "\""],        # " " -> "
    ["—", ", "], ["–", "-"],         # em dash -> spoken pause (comma); en dash -> -
    ["…", " "],                            # … -> nothing spoken (was ".", which left a stray period)
    [" ", " "],                            # nbsp -> space
    ["•", "*"],                            # bullet -> * (matches Go: the earlier &^*#@ strip below already ran)
    ["¼", "1/4"], ["½", "1/2"], ["¾", "3/4"],
    ["×", "x"], ["÷", "/"],
    ["ç", "c"],
    ["©", "(c)"], ["®", "(r)"], ["™", "(tm)"],
]

GO_TRUTH_STRIP_CHARS = "&^@"


def tidy_spoken_pause_commas(text):
    """After "—" -> ", ": no space before a comma, no comma right after other
    punctuation or at the very start, no doubled commas."""
    text = re.sub(r"\s+,", ",", text)
    text = re.sub(r",(\s*,)+", ",", text)
    text = re.sub(r"([.!?;:]),", r"\1", text)
    text = re.sub(r"^\s*,\s*", "", text)
    return text


def apply_universal_character_cleanup(text, scaffold=None):
    rules = cleanup_rules(scaffold)
    replacements = rules.get("universal_character_replacements") or GO_TRUTH_UNIVERSAL_CHARACTER_REPLACEMENTS
    cleaned = strip_diacritics_except_vietnamese(text)
    for source, target in replacements:
        cleaned = cleaned.replace(source, target)
    return cleaned


def flatten_for_robot_speech(text, scaffold=None):
    rules = cleanup_rules(scaffold)
    flattened = re.sub(r"\|\|\|[01C]\|\|\|", " ", str(text))
    invisible_class = codepoint_class(rules.get("invisible_space_codepoints", []))
    if invisible_class:
        flattened = re.sub(f"[{invisible_class}]", " ", flattened)
    flattened = re.sub(r"\{\{[^}]*\}\}", " ", flattened)
    flattened = flatten_math_for_speech(flattened)
    flattened = remove_configured_label_prefixes(flattened, rules.get("prompt_label_prefixes", []))
    flattened = remove_configured_label_prefixes(flattened, rules.get("spoken_label_prefixes", []))
    emoji_class = codepoint_class(rules.get("emoji_codepoint_ranges", []))
    if emoji_class:
        flattened = re.sub(f"[{emoji_class}]", " ", flattened)
    markdown_chars = str(rules.get("markdown_chars", "")) + GO_TRUTH_STRIP_CHARS
    if markdown_chars:
        flattened = re.sub(f"[{re.escape(markdown_chars)}]+", " ", flattened)
    flattened = apply_universal_character_cleanup(flattened, scaffold)
    flattened = tidy_spoken_pause_commas(flattened)
    for token in rules.get("ellipsis_tokens", []):
        flattened = flattened.replace(str(token), " ")
    flattened = clean_inline_emphasis_punctuation(flattened, scaffold)
    flattened = apply_pronunciation_safety(flattened)
    flattened = soften_trigger_phrase_punctuation(flattened, scaffold)
    flattened = re.sub(r"\s+", " ", flattened).strip()
    if not flattened:
        raise ValueError("empty flat robot speech is not allowed in live Section 1")
    return flattened


ANIMATION_MODE_WIRE_COMMAND = {
    "fire_and_forget": "playAnimationWI",
    "blocking": "playAnimation",
}


def wirepod_command_packet(action, flat_spoken, mode="fire_and_forget"):
    wire_command = ANIMATION_MODE_WIRE_COMMAND.get(mode, "playAnimationWI")
    return f"{{{{{wire_command}||{action}}}}}... {flat_spoken}"


def chunk_robot_speech(text, limit=ROBOT_SPEECH_CHUNK_LIMIT):
    words = str(text).split()
    chunks = []
    current = []
    current_len = 0
    for word in words:
        next_len = len(word) if not current else current_len + 1 + len(word)
        if current and next_len > limit:
            chunks.append(" ".join(current))
            current = [word]
            current_len = len(word)
            continue
        current.append(word)
        current_len = next_len
    if current:
        chunks.append(" ".join(current))
    return chunks or [text]


def compact_context_line(text, limit=180):
    cleaned = " ".join(str(text).split())
    if len(cleaned) <= limit:
        return cleaned
    return cleaned[:limit].rsplit(" ", 1)[0].strip() + "..."


def wirepod_web_send_form(endpoint, params, timeout=30):
    url = WIREPOD_BASE_URL.rstrip("/") + endpoint + "?" + urllib.parse.urlencode(params)
    request = urllib.request.Request(
        url,
        data=b"",
        headers={"Content-Type": "application/x-www-form-urlencoded"},
        method="POST",
    )
    with urllib.request.urlopen(request, timeout=timeout) as response:
        body = response.read().decode("utf-8", errors="replace")
        return {"status": response.status, "body": body}


def wirepod_get_text(path, timeout=10):
    url = WIREPOD_BASE_URL.rstrip("/") + path
    with urllib.request.urlopen(url, timeout=timeout) as response:
        return response.read().decode("utf-8", errors="replace")


def restart_wirepod_preflight(live):
    """The one shared Wire-Pod restart function - every caller (interview
    runner, phcal, wpr) routes through this, no duplicated restart logic
    anywhere else. GOLDEN RULE, see .claude/skills/wirepod-restart-discipline/
    SKILL.md: CHECK -> CLEAR -> START (only if unhealthy) -> CONFIRM.
    Restarting an already-healthy service is what trips systemd's
    start-rate limiter, not the restart command itself - stop-then-start
    does not dodge it either. CHECK first and skipping when already healthy
    is the actual fix; CLEAR (reset-failed) is harmless when nothing is
    wrong and required when a prior restart already tripped the limiter.
    The MIRROR step (apply_nongo_files.sh) was removed 2026-10-02: Wire-Pod
    reads GOPOD's files straight from the overlay folder."""
    if not live or not RESTART_WIREPOD_BEFORE_RUN:
        return {"enabled": False}

    print("PREFLIGHT: CHECK - is wire-pod already answering on :8080?")
    try:
        wirepod_get_text("/", timeout=2)
        already_healthy = True
    except Exception:
        already_healthy = False
    if already_healthy:
        print("PREFLIGHT: SKIPPED_HEALTHY - wire-pod already up, no restart fired")
        return {
            "enabled": True,
            "command": "wrp",
            "status": "SKIPPED_HEALTHY",
            "service_state": "active",
            "http_ready": True,
        }

    print("PREFLIGHT: CLEAR -> sudo systemctl reset-failed wire-pod.service")
    subprocess.run(["sudo", "systemctl", "reset-failed", "wire-pod.service"], check=False)


    print("PREFLIGHT: START -> sudo systemctl restart wire-pod")
    try:
        subprocess.run(["sudo", "systemctl", "restart", "wire-pod"], check=True)
    except subprocess.CalledProcessError as exc:
        raise RuntimeError(f"wire-pod restart preflight blocked: start_failed rc={exc.returncode}") from exc

    state = "unknown"
    for _ in range(10):
        time.sleep(1)
        active = subprocess.run(
            ["systemctl", "is-active", "wire-pod"],
            check=False,
            capture_output=True,
            text=True,
        )
        state = active.stdout.strip()
        if state == "active":
            break
    if state != "active":
        raise RuntimeError(f"wire-pod restart preflight blocked: service_state={state}")
    http_ready = False
    for _ in range(20):
        time.sleep(1)
        try:
            wirepod_get_text("/", timeout=2)
            http_ready = True
            break
        except Exception:
            continue
    if not http_ready:
        raise RuntimeError("wire-pod restart preflight blocked: http_not_ready")
    time.sleep(WIREPOD_POST_RESTART_SETTLE_SECONDS)
    print(f"PREFLIGHT: CONFIRM PASS - wire-pod {state}")
    return {
        "enabled": True,
        "command": "wrp",
        "status": "PASS",
        "service_state": state,
        "http_ready": http_ready,
        "post_restart_settle_seconds": WIREPOD_POST_RESTART_SETTLE_SECONDS,
    }


def start_wirepod_preflight(live):
    """2026-08-21, WPS_WPQ_ALIASES_SURVEY_001.md -> WPS_WPQ_ALIASES_BUILT_001.md:
    wps's own shared function, built as a near-exact mirror of
    restart_wirepod_preflight() above - same CHECK -> CLEAR ->
    START -> CONFIRM shape, same golden-rule reasoning (that function's own
    docstring), only two things differ: STEP 1's skip condition/message
    ("already running" instead of "already healthy, no restart fired") and
    STEP 3's actual systemctl verb (`start`, not `restart` - a bare start
    on an already-active unit is what wps is here to AVOID by checking
    first, not something this command itself needs to guard against, but
    checking first is still correct: skipping avoids the unnecessary
    reset-failed/mirror work on a service that's already fine). Not
    called by the interview runner or phcal - wps is a standalone
    operator command, its own bash wrapper below sets
    GOPOD_RESTART_WIREPOD_BEFORE_RUN=1 the same way _wpr_check() does,
    matching RESTART_WIREPOD_BEFORE_RUN's own gate exactly."""
    if not live or not RESTART_WIREPOD_BEFORE_RUN:
        return {"enabled": False}

    print("PREFLIGHT: CHECK - is wire-pod already answering on :8080?")
    try:
        wirepod_get_text("/", timeout=2)
        already_running = True
    except Exception:
        already_running = False
    if already_running:
        print("PREFLIGHT: SKIPPED_RUNNING - wire-pod already up, no start fired")
        return {
            "enabled": True,
            "command": "wps",
            "status": "SKIPPED_RUNNING",
            "service_state": "active",
            "http_ready": True,
        }

    print("PREFLIGHT: CLEAR -> sudo systemctl reset-failed wire-pod.service")
    subprocess.run(["sudo", "systemctl", "reset-failed", "wire-pod.service"], check=False)


    print("PREFLIGHT: START -> sudo systemctl start wire-pod")
    try:
        subprocess.run(["sudo", "systemctl", "start", "wire-pod"], check=True)
    except subprocess.CalledProcessError as exc:
        raise RuntimeError(f"wire-pod start preflight blocked: start_failed rc={exc.returncode}") from exc

    state = "unknown"
    for _ in range(10):
        time.sleep(1)
        active = subprocess.run(
            ["systemctl", "is-active", "wire-pod"],
            check=False,
            capture_output=True,
            text=True,
        )
        state = active.stdout.strip()
        if state == "active":
            break
    if state != "active":
        raise RuntimeError(f"wire-pod start preflight blocked: service_state={state}")
    http_ready = False
    for _ in range(20):
        time.sleep(1)
        try:
            wirepod_get_text("/", timeout=2)
            http_ready = True
            break
        except Exception:
            continue
    if not http_ready:
        raise RuntimeError("wire-pod start preflight blocked: http_not_ready")
    time.sleep(WIREPOD_POST_RESTART_SETTLE_SECONDS)
    print(f"PREFLIGHT: CONFIRM PASS - wire-pod {state}")
    return {
        "enabled": True,
        "command": "wps",
        "status": "PASS",
        "service_state": state,
        "http_ready": http_ready,
        "post_restart_settle_seconds": WIREPOD_POST_RESTART_SETTLE_SECONDS,
    }


class Robots:
    def __init__(self, live, scaffold=None, voice_destination="robot", bare_capture=False, console_rich_display=None, console_timestamps=False):
        self.live = live
        self.scaffold = scaffold or load_interview_scaffold()
        self.voice_destination = voice_destination
        self.bare_capture = bare_capture
        if console_rich_display is None:
            console_rich_display = os.getenv("GOPOD_CONSOLE_RICH_DISPLAY", "1") != "0"
        self.console_rich_display = console_rich_display
        self.console_timestamps = console_timestamps
        self._last_event_wall_time = None

    def timing_prefix(self):
        """Public so a caller's own script (e.g. run_songs_runner_001.py's
        main loop) can prefix its own print() calls with the same shared clock -
        one running elapsed-time trail across both files' terminal output, not
        two independent clocks. No-op string when console_timestamps is off."""
        if not self.console_timestamps:
            return ""
        now = time.time()
        clock = time.strftime("%H:%M:%S", time.localtime(now)) + f".{int(now % 1 * 1000):03d}"
        if self._last_event_wall_time is None:
            self._last_event_wall_time = now
            return f"[{clock} +0.000s] "
        delta = now - self._last_event_wall_time
        self._last_event_wall_time = now
        return f"[{clock} +{delta:.3f}s] "

    def connect(self):
        if not self.live:
            print("DRY: live speech gate is off.")
            return
        wirepod_route_line = f"LIVE: Wire-Pod ESN routing enabled: {WIREPOD_BASE_URL}"
        brobot_1_esn_line = f"LIVE: Brobot 1 ESN: {BROBOT_1_SERIAL}"
        brobot_2_esn_line = f"LIVE: Brobot 2 ESN: {BROBOT_2_SERIAL}"
        print(wirepod_route_line)
        _announce_status_kokoro(wirepod_route_line)
        print(brobot_1_esn_line)
        _announce_status_kokoro(brobot_1_esn_line)
        print(brobot_2_esn_line)
        _announce_status_kokoro(brobot_2_esn_line)

    def speaker_visual_cue(self, serial):
        cue = self.scaffold.get("speaker_visual_cue", {})
        if not cue.get("enabled"):
            return {"enabled": False}
        endpoint = cue.get("endpoint", "/api-sdk/move_lift")
        axis = MOVEMENT_ENDPOINT_AXIS.get(endpoint, "unknown")
        resolve_movement_axis(axis, self.scaffold, required_for="speaker_visual_cue")
        sequence = cue.get("sequence", [])
        validate_movement_sequence_ends_at_zero(sequence)
        results = []
        for step in sequence:
            if not isinstance(step, dict):
                continue
            speed = str(step.get("speed", 0))
            hold_seconds = float(step.get("hold_seconds", 0))
            result = wirepod_web_send_form(endpoint, {"speed": speed, "serial": serial}, timeout=5)
            results.append(
                {
                    "axis": axis,
                    "speed": speed,
                    "hold_seconds": hold_seconds,
                    "hold_seconds_is_asserted_not_measured": True,
                    "explicit_zero": float(speed) == 0,
                    "serial": serial,
                    "status": result.get("status"),
                    "body": result.get("body", ""),
                }
            )
            time.sleep(hold_seconds)
        return {"enabled": True, "endpoint": endpoint, "axis": axis, "results": results}

    def say(self, name, robot_safe_line, display_text="", display_role="", suppress_assume_release=False):
        animation_mode = self.scaffold.get("animation_mode", "fire_and_forget")
        resolve_animation_mode(animation_mode, self.scaffold)
        action, raw_spoken = split_robot_safe_line(robot_safe_line, self.scaffold)
        spoken = flatten_for_robot_speech(raw_spoken, self.scaffold)
        speech_chunks = chunk_robot_speech(spoken)
        wirepod_chunks = [wirepod_command_packet(action, chunk, animation_mode) for chunk in speech_chunks]
        if self.console_rich_display and display_text:
            console_line = display_text
        elif self.console_rich_display:
            console_line = raw_spoken
        elif self.bare_capture:
            console_line = raw_spoken
        else:
            console_line = robot_safe_line
        print(f"{self.timing_prefix()}{name}: {console_line}")
        channels_used = describe_channels_used(display_text)
        serial = BROBOT_1_SERIAL if name == "Brobot 1" else BROBOT_2_SERIAL
        prelude = {
            "serial": serial,
            "speech_text": spoken,
            "raw_speech_text": raw_spoken,
            "speech_chunks": speech_chunks,
            "display_text": display_text,
            "animation_action": action,
            "wirepod_text": wirepod_chunks[0],
            "wirepod_chunks": wirepod_chunks,
            "channels_used": channels_used,
            "chunk_results": [],
            "assume": None,
            "speaker_visual_cue": None,
            "say": None,
            "release": None,
        }

        if self.voice_destination == "off":
            return {**prelude, "spoken": False, "mode": "off"}

        if self.voice_destination == "monitor" or (BROBOTS_SOLO and name == "Brobot 2" and self.voice_destination == "robot"):
            resolve_voice_engine("kokoro_82m", self.scaffold, required_for="interview_turns")
            result = {**prelude, "mode": "kokoro_monitor"}
            all_spoken = True
            for chunk in speech_chunks:
                spoken_result = speak_turn_via_kokoro(chunk)
                result["chunk_results"].append(spoken_result)
                if not spoken_result.get("spoken"):
                    all_spoken = False
                    break
            result["say"] = result["chunk_results"][-1] if result["chunk_results"] else None
            result["spoken"] = all_spoken
            return result

        turn_engine = self.scaffold.get("turn_voice_engine", "robot_tts")
        resolve_voice_engine(turn_engine, self.scaffold, required_for="interview_turns")
        if not self.live:
            return {**prelude, "spoken": False, "mode": "dry"}
        result = {**prelude, "spoken": False, "mode": "wirepod"}
        try:
            if suppress_assume_release:
                result["assume"] = {"suppressed": True}
            else:
                assume = wirepod_web_send_form("/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10)
                result["assume"] = assume
                time.sleep(0.25)
            all_spoken = True
            for index, wirepod_text in enumerate(wirepod_chunks):
                say = wirepod_web_send_form(
                    "/api-sdk/say_text",
                    {
                        "text": wirepod_text,
                        "serial": serial,
                        "display_role": display_role or name,
                        "display_text": display_text if index == 0 else "",
                        **({"sentence_highlight": "1"} if getattr(self, "sentence_highlight", False) else {}),
                        **({"piece_offset": str(self.piece_offset)} if getattr(self, "piece_offset", None) is not None else {}),
                    },
                    timeout=90,
                )
                result["chunk_results"].append(say)
                if say["status"] != 200 or say["body"].strip() != "success":
                    all_spoken = False
                    break
            result["say"] = result["chunk_results"][-1] if result["chunk_results"] else {}
            result["spoken"] = all_spoken
            return result
        except Exception as exc:
            result["error"] = repr(exc)
            print(f"{self.timing_prefix()}{name} Wire-Pod ERROR: {exc!r}")
            return result
        finally:
            if suppress_assume_release:
                result["release"] = {"suppressed": True}
                print(f"{self.timing_prefix()}{name} Wire-Pod: say={(result.get('say') or {}).get('body', '').strip()} (assume/release held by caller)")
            else:
                try:
                    release = wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)
                    result["release"] = release
                    print(
                        f"{self.timing_prefix()}{name} Wire-Pod: "
                        f"assume={(result.get('assume') or {}).get('body', '').strip()} "
                        f"say={(result.get('say') or {}).get('body', '').strip()} "
                        f"release={release['body'].strip()}"
                    )
                except Exception as exc:
                    result["release_error"] = repr(exc)
                    print(f"{self.timing_prefix()}{name} Wire-Pod release ERROR: {exc!r}")

    def backpack_turn(self, name, serial, visible_line, display_text="", failure_code="BROBOT_SPEECH_FAILURE", display_role=""):
        result = self.say(name, visible_line, display_text, display_role)
        accepted = result.get("spoken") is True or result.get("mode") in ("dry", "off")
        result["transport"] = "existing_wirepod_say_text"
        result["markers"] = "stripped_outside_chipper"
        result["speech_committed"] = accepted
        result["advance_rule"] = "advance_after_wirepod_say_text_returns"
        if not accepted:
            result["failure"] = failure_code
        return result

    def brobot_2_streaming_kg_committed(self, line, prior_context="", scaffold=None):
        raw = llm_colour_line("BROBOT 2", line, prior_context=prior_context, scaffold=scaffold)
        robot_line, repair_used = normalize_robot_safe(raw, scaffold)
        result = self.backpack_turn("Brobot 2", BROBOT_2_SERIAL, robot_line, raw, "BROBOT2_SPEECH_FAILURE")
        result["raw_llm_response"] = raw
        result["repair_used"] = repair_used
        return result

    def brobot_1_streaming_kg_committed(self, line, interviewer_coloured_line, prior_context="", scaffold=None):
        value_target = build_value_target(line, interviewer_coloured_line, scaffold)
        raw = llm_colour_line("BROBOT 1", line, value_target, prior_context, scaffold)
        robot_line, repair_used = normalize_robot_safe(raw, scaffold)
        result = self.backpack_turn("Brobot 1", BROBOT_1_SERIAL, robot_line, raw, "BROBOT1_SPEECH_FAILURE")
        result["raw_llm_response"] = raw
        result["repair_used"] = repair_used
        return result

    def close(self):
        return


def split_internal_thoughts(raw):
    match = re.search(r"(?im)^\s*REPLY:\s*", raw)
    if not match:
        return "", raw
    thoughts = re.sub(r"(?im)^\s*THOUGHTS:\s*", "", raw[: match.start()]).strip()
    reply = raw[match.end():].strip()
    if not reply:
        return "", raw
    return thoughts, reply


def resolve_turn_mode(role, line, scaffold):
    key = "brobot_2_mode" if role == "BROBOT 2" else "brobot_1_mode"
    explicit = line.get(key)
    if explicit in ("canned", "llm"):
        return explicit
    exchange_type_entry = scaffold.get("exchange_types", {}).get(line.get("exchange_type", ""), {})
    return exchange_type_entry.get("brobot_2_mode" if role == "BROBOT 2" else "brobot_1_mode", "llm")


def resolve_canned_text(role, line):
    key = "brobot_2_canned_source" if role == "BROBOT 2" else "brobot_1_canned_source"
    source = line.get(key)
    if source == "visible_line":
        return line.get("visible_line", "")
    if source == "value_points_joined":
        return " ".join(line.get("value_points") or [])
    if role == "BROBOT 2":
        return line.get("visible_line", "")
    return " ".join(line.get("value_points") or [])


def generated_turn(name, role, line, value_target="", prior_context="", scaffold=None, content_model=None):
    scaffold = scaffold or load_interview_scaffold()
    mode = resolve_turn_mode(role, line, scaffold)
    internal_thoughts = ""
    if mode == "canned":
        status(f"CANNED line={line.get('line_number')} role={role}")
        raw = resolve_canned_text(role, line)
    else:
        raw = llm_colour_line(role, line, value_target, prior_context, scaffold, content_model)
        if line.get("thinking_window"):
            internal_thoughts, raw = split_internal_thoughts(raw)
    robot_line, repair_used = normalize_robot_safe(raw, scaffold)
    action, raw_spoken = split_robot_safe_line(robot_line, scaffold)
    speech_text = flatten_for_robot_speech(raw_spoken, scaffold)
    return {
        "line_number": line.get("line_number"),
        "name": name,
        "role": role,
        "serial": BROBOT_1_SERIAL if name == "Brobot 1" else BROBOT_2_SERIAL,
        "raw_llm_response": raw,
        "internal_thoughts": internal_thoughts,
        "robot_safe_line": robot_line,
        "repair_used": repair_used,
        "animation_action": action,
        "raw_speech_text": raw_spoken,
        "speech_text": speech_text,
        "speech_chunks": chunk_robot_speech(speech_text),
    }


def build_value_target(line, interviewer_coloured_line, scaffold=None):
    scaffold = scaffold or load_interview_scaffold()
    response_context = scaffold.get("response_context", {})
    parts = [response_context.get("interviewer_line_prefix", "Question/statement being answered, do not quote: ") + line.get("visible_line", "")]
    points = line.get("value_points") or []
    if points:
        parts.append(response_context.get("value_target_prefix", "In your response, naturally reveal these points: "))
        parts.extend(f"{index}. {point}" for index, point in enumerate(points, 1))
        parts.append(
            response_context.get(
                "value_points_locked_notice",
                "These points are locked and ordered - do not invent, add, drop, or reorder them.",
            )
        )
    elif line.get("purpose"):
        parts.append(response_context.get("value_target_prefix", "In your response, naturally reveal these points: ") + line["purpose"])
    return "\n".join(item for item in parts if item.strip())


ECHO_DETECTION_WORD_COUNT = 5

_ECHO_WORD_RE = re.compile(r"[a-z0-9']+")


def echo_check_words(text):
    """Lowercased word tokens, punctuation/emoji/symbols stripped entirely."""
    return _ECHO_WORD_RE.findall((text or "").lower())


def echo_ngrams(words, n):
    return {tuple(words[i:i + n]) for i in range(len(words) - n + 1)}


def detect_echo(candidate_text, source_text, n=ECHO_DETECTION_WORD_COUNT):
    """Return the matched N-word snippet if candidate_text verbatim-overlaps
    source_text by at least n consecutive words, else None."""
    candidate_words = echo_check_words(candidate_text)
    source_words = echo_check_words(source_text)
    if len(candidate_words) < n or len(source_words) < n:
        return None
    source_grams = echo_ngrams(source_words, n)
    for i in range(len(candidate_words) - n + 1):
        gram = tuple(candidate_words[i:i + n])
        if gram in source_grams:
            return " ".join(gram)
    return None


ECHO_RETRY_LIMIT = 3


def _check_echo(brobot_1_candidate, line, brobot_2_generated, echo_word_count):
    candidate = brobot_1_candidate.get("raw_llm_response", "")
    snippet = detect_echo(candidate, line.get("visible_line", ""), n=echo_word_count)
    source = "visible_line" if snippet else None
    if not snippet:
        snippet = detect_echo(candidate, brobot_2_generated.get("raw_llm_response", ""), n=echo_word_count)
        source = "brobot_2_response" if snippet else None
    return bool(snippet), source, snippet


SELF_REPEAT_RETRY_LIMIT = 3


def _check_self_repeat(candidate_generated, own_prior_raw_text, echo_word_count):
    if not own_prior_raw_text:
        return False, None
    candidate = candidate_generated.get("raw_llm_response", "")
    snippet = detect_echo(candidate, own_prior_raw_text, n=echo_word_count)
    return bool(snippet), snippet


def _generate_with_self_repeat_guard(
    name,
    role,
    line,
    own_prior_raw_text,
    echo_word_count,
    line_number,
    initial_generated,
    value_target="",
    prior_context="",
    scaffold=None,
    content_model=None,
):
    generated = initial_generated
    repeated, snippet = _check_self_repeat(generated, own_prior_raw_text, echo_word_count)
    retry_history = [{"attempt": 1, "self_repeat_detected": repeated, "self_repeat_snippet": snippet}]
    if repeated:
        status(f"SELF_REPEAT_DETECTED line={line_number} role={role} attempt=1")
    slug = "brobot2" if role == "BROBOT 2" else "brobot1"
    while repeated and len(retry_history) <= SELF_REPEAT_RETRY_LIMIT:
        attempt_number = len(retry_history) + 1
        status(f"SELF_REPEAT_RETRY line={line_number} role={role} attempt={attempt_number}")
        with TIMING.record(
            f"stage3_line{line_number}_{slug}_selfrepeat_attempt{attempt_number}",
            f"Interview line {line_number} {name} self-repeat-retry regeneration, attempt {attempt_number}",
            "true_completion",
            note=(
                "Real LLM spend every attempt, bounded by SELF_REPEAT_RETRY_LIMIT - a separate "
                "guard from the cross-speaker echo guard's own ECHO_RETRY_LIMIT, layered "
                "alongside it, never merged into or counted against it."
            ),
        ):
            generated = generated_turn(name, role, line, value_target, prior_context, scaffold, content_model)
        repeated, snippet = _check_self_repeat(generated, own_prior_raw_text, echo_word_count)
        retry_history.append({"attempt": attempt_number, "self_repeat_detected": repeated, "self_repeat_snippet": snippet})
        if repeated:
            status(f"SELF_REPEAT_DETECTED line={line_number} role={role} attempt={attempt_number}")
    if repeated and len(retry_history) == SELF_REPEAT_RETRY_LIMIT + 1:
        status(f"SELF_REPEAT_EXHAUSTED line={line_number} role={role} attempts={len(retry_history)}")
    return generated, repeated, retry_history


def generate_exchange_record(line, prior_context="", scaffold=None, content_model=None, own_prior_raw=None):
    scaffold = scaffold or load_interview_scaffold()
    own_prior_raw = own_prior_raw or {}
    exchange_line_number = line.get("line_number")
    echo_word_count = scaffold.get("generation_defaults", {}).get(
        "echo_detection_word_count", ECHO_DETECTION_WORD_COUNT
    )
    with TIMING.record(
        f"stage3_line{exchange_line_number}_brobot2_colour",
        f"Interview line {exchange_line_number} Brobot 2 (interviewer) LLM colour-pass / canned resolve",
        "true_completion",
        note="A real blocking LLM chat-completion call in llm mode; near-zero elapsed in canned mode (text lookup only, no LLM call made).",
    ):
        brobot_2_generated = generated_turn("Brobot 2", "BROBOT 2", line, prior_context=prior_context, scaffold=scaffold, content_model=content_model)
    brobot_2_self_repeated, brobot_2_self_repeat_history = False, []
    if resolve_turn_mode("BROBOT 2", line, scaffold) != "canned":
        brobot_2_generated, brobot_2_self_repeated, brobot_2_self_repeat_history = _generate_with_self_repeat_guard(
            "Brobot 2",
            "BROBOT 2",
            line,
            own_prior_raw.get("BROBOT 2", ""),
            echo_word_count,
            exchange_line_number,
            brobot_2_generated,
            prior_context=prior_context,
            scaffold=scaffold,
            content_model=content_model,
        )
    value_target = build_value_target(line, brobot_2_generated["raw_llm_response"], scaffold)
    with TIMING.record(
        f"stage3_line{exchange_line_number}_brobot1_colour_attempt1",
        f"Interview line {exchange_line_number} Brobot 1 (interviewee) LLM colour-pass / canned resolve, attempt 1",
        "true_completion",
        note="A real blocking LLM chat-completion call in llm mode; near-zero elapsed in canned mode (text lookup only, no LLM call made).",
    ):
        brobot_1_generated = generated_turn("Brobot 1", "BROBOT 1", line, value_target, prior_context, scaffold, content_model)
    brobot_1_self_repeated, brobot_1_self_repeat_history = False, []
    if resolve_turn_mode("BROBOT 1", line, scaffold) != "canned":
        brobot_1_generated, brobot_1_self_repeated, brobot_1_self_repeat_history = _generate_with_self_repeat_guard(
            "Brobot 1",
            "BROBOT 1",
            line,
            own_prior_raw.get("BROBOT 1", ""),
            echo_word_count,
            exchange_line_number,
            brobot_1_generated,
            value_target=value_target,
            prior_context=prior_context,
            scaffold=scaffold,
            content_model=content_model,
        )
    record = {
        "line_number": line["line_number"],
        "line_type": line["line_type"],
        "default_speaker": line.get("default_speaker", ""),
        "brobot_2_source_visible_line": line["visible_line"],
        "silent_value_cue": line.get("silent_value_cue", ""),
        "value_points": line.get("value_points", []),
        "interviewer_value_point_count": interviewer_value_point_count(line),
        "interviewer_value_points_selected": selected_interviewer_value_points(line),
        "generation_phase": "complete",
        "brobot_2_generated": brobot_2_generated,
        "brobot_1_generated": brobot_1_generated,
        "self_repeat_retry_limit": SELF_REPEAT_RETRY_LIMIT,
        "brobot_2_self_repeat_detected": brobot_2_self_repeated,
        "brobot_2_self_repeat_history": brobot_2_self_repeat_history,
        "brobot_1_self_repeat_detected": brobot_1_self_repeated,
        "brobot_1_self_repeat_history": brobot_1_self_repeat_history,
        "brobot_2_movement": line.get("brobot_2_movement"),
        "brobot_1_movement": line.get("brobot_1_movement"),
        "brobot_2_movement_2": line.get("brobot_2_movement_2"),
        "brobot_1_movement_2": line.get("brobot_1_movement_2"),
        "marker_events": "not_runtime_control",
        "display_features": {
            "raw_visible_to_operator": "wirepod_log_window",
            "filtered_speech_visible_to_operator": "wirepod_log_window",
            "checkpoint_visible_to_operator": line["line_type"] == "CHECKPOINT",
        },
    }

    brobot_1_mode = resolve_turn_mode("BROBOT 1", line, scaffold)
    if brobot_1_mode != "canned":
        echo_word_count = scaffold.get("generation_defaults", {}).get(
            "echo_detection_word_count", ECHO_DETECTION_WORD_COUNT
        )
        line_number = line.get("line_number")
        echoed, echo_source, echo_snippet = _check_echo(brobot_1_generated, line, brobot_2_generated, echo_word_count)
        retry_history = [{"attempt": 1, "echoed": echoed, "echo_source": echo_source, "echo_snippet": echo_snippet}]
        if echoed:
            status(f"ECHO_DETECTED line={line_number} attempt=1 source={echo_source}")
        while echoed and len(retry_history) <= ECHO_RETRY_LIMIT:
            attempt_number = len(retry_history) + 1
            status(f"ECHO_RETRY line={line_number} attempt={attempt_number}")
            with TIMING.record(
                f"stage3_line{line_number}_brobot1_colour_attempt{attempt_number}",
                f"Interview line {line_number} Brobot 1 echo-retry regeneration, attempt {attempt_number}",
                "true_completion",
                note="Real LLM spend every attempt, bounded by ECHO_RETRY_LIMIT - never a hidden/free retry.",
            ):
                brobot_1_generated = generated_turn("Brobot 1", "BROBOT 1", line, value_target, prior_context, scaffold, content_model)
            echoed, echo_source, echo_snippet = _check_echo(brobot_1_generated, line, brobot_2_generated, echo_word_count)
            retry_history.append(
                {"attempt": attempt_number, "echoed": echoed, "echo_source": echo_source, "echo_snippet": echo_snippet}
            )
            if echoed:
                status(f"ECHO_DETECTED line={line_number} attempt={attempt_number} source={echo_source}")

        record["brobot_1_generated"] = brobot_1_generated
        record["echo_detected"] = echoed
        record["echo_retry_limit"] = ECHO_RETRY_LIMIT
        record["echo_retry_attempts"] = len(retry_history)
        record["echo_retry_history"] = retry_history
        if echoed:
            record["echo_source"] = echo_source
            record["echo_snippet"] = echo_snippet
            if len(retry_history) == ECHO_RETRY_LIMIT + 1:
                record["echo_retry_exhausted"] = True
                status(f"ECHO_RETRY_EXHAUSTED line={line_number} attempts={len(retry_history)}")

    return record


def play_generated_turn(robots, generated, failure_code, scaffold=None):
    scaffold = scaffold or load_interview_scaffold()
    line_number = int(generated.get("line_number") or 0)
    exchange_half = scaffold.get("display_exchange_halves", {}).get(generated["name"], "1_of_2")
    display_role = scaffold.get("display_role_template", "BROBOT_RICH_LINE_{line_number}_EXCHANGE_{exchange_half}").format(
        line_number=line_number,
        exchange_half=exchange_half,
    )
    result = robots.backpack_turn(
        generated["name"],
        generated["serial"],
        generated["robot_safe_line"],
        generated["speech_text"],
        failure_code,
        display_role,
    )
    result["generation_source"] = "section1_generated_json"
    return result


def utc_run_id():
    return time.strftime("%Y%m%d_%H%M%S", time.gmtime())


def write_log(log):
    if log.get("log_path"):
        path = Path(log["log_path"])
    else:
        run_id = log.get("run_id") or utc_run_id()
        log["run_id"] = run_id
        run_dir = LOG_DIR / f"section1_full_live_{run_id}"
        run_dir.mkdir(parents=True, exist_ok=True)
        path = run_dir / f"section1_full_live_log_{run_id}.json"
    log["log_path"] = str(path)
    log["updated_utc"] = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
    path.write_text(json.dumps(log, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    status(f"JSON_WRITTEN {path}")
    return path


def load_replay_log(path):
    status(f"REPLAY_LOAD {path}")
    replay = json.loads(Path(path).expanduser().read_text(encoding="utf-8"))
    events = [
        event
        for event in replay.get("events", [])
        if event.get("brobot_2_generated") and event.get("brobot_1_generated")
    ]
    if not events:
        raise RuntimeError(f"BLOCKED: replay log has no generated interview events: {path}")
    status(f"REPLAY_READY events={len(events)}")
    return replay, events


def generate_phase(live=None, voice_destination=None, content_model=None):
    """Pre-show half of a run: Kokoro warm-up, scaffold/section-card load,
    Wire-Pod ESN routing announce, then either the LLM generating every
    exchange line or a replay-log load - written to a JSON log on disk and
    returned. Every status() beat in this function is Kokoro-narrated (none
    of them match KOKORO_STATUS_STOP_PREFIXES) UNLESS a caller has already
    silenced _KOKORO_STATUS_STATE["active"] (see the scored pre-show song
    path below, which does exactly this - narration then belongs to the
    song, not to this function's own internal progress prints; the prints
    themselves, and everything this function produces, are unchanged
    either way). Ends at write_log(log), before PLAYBACK_START - no robot
    speaks an interview line in this function. Split out of the original
    single main() so this half can run standalone (see
    run_section1_preshow_generate_001.py); main() below still calls this and
    playback_phase() back to back, byte-identical to the pre-split combined
    behavior."""
    if voice_destination is None:
        voice_destination = resolve_voice_destination()
    if live is None:
        live = os.getenv(LIVE_GATE) == "1"
    live = live and (voice_destination == "robot")
    _KOKORO_STATUS_STATE["live"] = live
    _KOKORO_STATUS_STATE["voice_destination"] = voice_destination
    section_song_dir = os.getenv("GOPOD_SECTION_SONG_DIR")
    use_legacy_card = (not section_song_dir) and os.getenv(LEGACY_SECTION_CARD_ENV, "0") == "1"
    if use_legacy_card:
        section_source = CARD_PATH
    elif section_song_dir:
        section_source = Path(section_song_dir).expanduser()
    else:
        section_source = DEFAULT_SECTION_SONG_DIR
    bare_capture = peek_song_bare_capture(section_source, use_legacy_card)
    if content_model is None:
        content_model = "" if (SECTION1_REPLAY_LOG or bare_capture) else resolve_content_model()
    if content_model:
        print(f"CONTENT_MODEL_SELECTED model={content_model}", flush=True)
    if bare_capture:
        _KOKORO_STATUS_STATE["active"] = False
    else:
        configure_voice_engines(load_interview_scaffold(INTERVIEW_SCAFFOLD_PATH))
        with TIMING.record(
            "stage1_kokoro_status_warmup",
            "Brobot 3 Kokoro status-voice model warm-up",
            "true_completion",
            note="Blocks on the speak-stdin worker's ack after KPipeline() finishes loading (~20s cold, near-zero if already warm).",
        ):
            kokoro_status_warm_up()
    run_id = utc_run_id()
    status(f"RUN_START run_id={run_id} live={live}")
    status(f"SCAFFOLD_LOAD {INTERVIEW_SCAFFOLD_PATH}")
    interview_scaffold = load_interview_scaffold(INTERVIEW_SCAFFOLD_PATH)
    status(f"SCAFFOLD_READY {interview_scaffold.get('scaffold_id', '')}")
    status(f"SECTION_CARD_LOAD {section_source}")
    section_card = load_section_card(CARD_PATH) if use_legacy_card else load_section_pair(section_source)
    lines = section_card["lines"]
    status(f"SECTION_CARD_READY lines={len(lines)}")
    robots = Robots(live, interview_scaffold, voice_destination, bare_capture=bare_capture)
    with TIMING.record(
        "stage1_wirepod_restart_preflight",
        "Wire-Pod restart -> HTTP-ready poll (+ configured post-restart settle)",
        "true_completion",
        note=(
            "Only fires when GOPOD_RESTART_WIREPOD_BEFORE_RUN=1; returns instantly (near-zero "
            "elapsed) otherwise. The restart/poll portion blocks on a real confirmed HTTP "
            "response; WIREPOD_POST_RESTART_SETTLE_SECONDS afterward is a fixed additional "
            "sleep stacked on top, not itself a further confirmation."
        ),
    ) as _preflight_span:
        wirepod_restart = restart_wirepod_preflight(live)
        _preflight_span["preflight_enabled"] = wirepod_restart.get("enabled", False)
    log = {
        "source_card": str(section_source),
        "run_id": run_id,
        "run_started_utc": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "interview_scaffold_path": str(INTERVIEW_SCAFFOLD_PATH),
        "interview_scaffold_id": interview_scaffold.get("scaffold_id", ""),
        "section_card": section_card,
        "trigger_style": "outside_wrapper_existing_wirepod_say_text",
        "live_gate": LIVE_GATE,
        "live_enabled": live,
        "voice_destination": voice_destination,
        "content_model": content_model,
        "wirepod_base_url": WIREPOD_BASE_URL,
        "wirepod_restart_preflight": wirepod_restart,
        "robot_expression_gate": ROBOT_EXPRESSION_GATE,
        "max_animation_tags": interview_scaffold.get("max_animation_tags", 2),
        "raw_vs_speech_split": True,
        "execution_model": "generate_json_then_playback",
        "replay_source_log": SECTION1_REPLAY_LOG,
        "section1_max_exchanges": SECTION1_MAX_EXCHANGES,
        "robot_speech_chunk_limit": ROBOT_SPEECH_CHUNK_LIMIT,
        "marker_contract": "markers_are_text_boundaries_only_outside_chipper",
        "brobot_1_serial": BROBOT_1_SERIAL,
        "brobot_2_serial": BROBOT_2_SERIAL,
        "bare_capture": bare_capture,
        "events": [],
    }
    with TIMING.record(
        "stage1_robots_connect_esn_announce",
        "Wire-Pod ESN routing + Brobot 1/Brobot 2 ESN announce lines",
        "true_completion",
        note=(
            "Kokoro-narrated informational lines (blocks on speak-stdin DONE ack). If Kokoro "
            "narration has already been silenced for this run (the scored pre-show path sets "
            "_KOKORO_STATUS_STATE['active']=False before its background generation thread "
            "starts), each announce call no-ops immediately - elapsed reflects that honestly, "
            "it is not a stalled or skipped measurement."
        ),
    ):
        robots.connect()
    try:
        if SECTION1_REPLAY_LOG:
            replay_log, replay_events = load_replay_log(SECTION1_REPLAY_LOG)
            log["replay_source_run_id"] = replay_log.get("run_id", "")
            log["events"] = replay_events[:SECTION1_MAX_EXCHANGES] if SECTION1_MAX_EXCHANGES > 0 else replay_events
            log["generation_phase"] = "replayed_from_json"
            log["content_model"] = replay_log.get("content_model", "")
            status(f"GENERATION_REPLAYED events={len(log['events'])}")
        else:
            generation_defaults = interview_scaffold.get("generation_defaults", {})
            prior_window = generation_defaults.get("prior_context_exchange_count", 2)
            prior_truncate = generation_defaults.get("prior_context_truncate_chars", 180)
            interviewer_prefix = generation_defaults.get("prior_context_interviewer_prefix", "Previous interviewer line: ")
            interviewee_prefix = generation_defaults.get("prior_context_interviewee_prefix", "Previous response value: ")
            prior_exchange_context = []
            own_prior_raw = {"BROBOT 2": "", "BROBOT 1": ""}
            selected_lines = lines[:SECTION1_MAX_EXCHANGES] if SECTION1_MAX_EXCHANGES > 0 else lines
            status(f"GENERATION_START exchanges={len(selected_lines)}")
            for line in selected_lines:
                gen_line_number = line.get("line_number")
                status(f"EXCHANGE_GENERATION_START line={gen_line_number}")
                window = prior_exchange_context[-prior_window:] if prior_window > 0 else []
                prior_context = "\n".join(window)
                with TIMING.record(
                    f"stage3_line{gen_line_number}_pregen_total",
                    f"Interview line {gen_line_number} generation (both speakers, all LLM colour-passes)",
                    "true_completion",
                    note=(
                        "Occurs during the generation phase (pre-spent, before playback) - "
                        "playback speaks this pre-generated text and does not re-incur this "
                        "time. See stage3_line{N}_brobot{1,2}_colour_attempt{N} for the "
                        "individual LLM calls inside this span."
                    ),
                ):
                    event = generate_exchange_record(line, prior_context, interview_scaffold, content_model, own_prior_raw=own_prior_raw)
                log["events"].append(event)
                status(f"EXCHANGE_GENERATION_DONE line={gen_line_number}")
                prior_exchange_context.extend(
                    [
                        interviewer_prefix + compact_context_line(event["brobot_2_generated"].get("raw_llm_response", ""), prior_truncate),
                        interviewee_prefix + compact_context_line(event["brobot_1_generated"].get("raw_llm_response", ""), prior_truncate),
                    ]
                )
                own_prior_raw = {
                    "BROBOT 2": event["brobot_2_generated"].get("raw_llm_response", ""),
                    "BROBOT 1": event["brobot_1_generated"].get("raw_llm_response", ""),
                }
            log["generation_phase"] = "complete"
            status("GENERATION_DONE")

        write_log(log)
        return log, interview_scaffold, robots
    finally:
        robots.close()


def playback_phase(log, interview_scaffold, robots, live):
    """Interview half of a run: plays back log["events"] (either just
    generated by generate_phase() in the same process, or loaded from a
    prior generate_phase() run's JSON log via GOPOD_SECTION1_REPLAY_LOG) -
    this is where the robots actually speak. PLAYBACK_START/PLAYBACK_TURN_START
    are in KOKORO_STATUS_STOP_PREFIXES, so Kokoro narration has already
    stopped by the time this function's status beats fire."""
    try:
        status(f"PLAYBACK_START events={len(log['events'])} live={live} voice_destination={robots.voice_destination}")
        speech_quality, speech_note = _speech_capture_quality(robots.voice_destination)
        first_speaker = log.get("section_card", {}).get("first_speaker", "brobot_2")
        if first_speaker not in ("brobot_2", "brobot_1"):
            raise RuntimeError(f"BLOCKED: first_speaker={first_speaker!r} must be 'brobot_2' or 'brobot_1'")
        movement_pause_seconds = float(log.get("section_card", {}).get("movement_pause_seconds", 0) or 0)
        kg_search_song = is_interview_kg_search_song(log)

        def pause_after_movement(segment_id, label):
            if movement_pause_seconds > 0:
                with TIMING.record(
                    segment_id,
                    label,
                    "true_completion",
                    note="Deterministic time.sleep() - configured movement_pause_seconds breathing gap, not variable show time.",
                ):
                    time.sleep(movement_pause_seconds)

        for event in log["events"]:
            if event.get("type") == "stop":
                continue
            line_number = event.get("line_number")
            with TIMING.record(
                f"stage3_line{line_number}_playback_total",
                f"Interview line {line_number} playback (both speakers, movement, pre-pause)",
                speech_quality,
                note=(
                    "Aggregate span; see nested stage3_line{N}_* segments for per-component "
                    "capture quality. " + speech_note
                ),
            ):
                def play_brobot_2():
                    status(f"PLAYBACK_TURN_START line={line_number} speaker=Brobot 2")
                    brobot_2_movement = event.get("brobot_2_movement")
                    if brobot_2_movement:
                        status(f"INTERVIEW_MOVEMENT_START line={line_number} speaker=Brobot 2 movement={brobot_2_movement}")
                        event["brobot_2_movement_result"] = fire_scored_interview_movement(
                            brobot_2_movement,
                            BROBOT_2_SERIAL,
                            live,
                            segment_id=f"stage3_line{line_number}_brobot2_movement",
                            segment_label=f"Interview line {line_number} Brobot 2 movement ({brobot_2_movement})",
                        )
                        status(f"INTERVIEW_MOVEMENT_DONE line={line_number} speaker=Brobot 2 movement={brobot_2_movement}")
                        pause_after_movement(
                            f"stage3_line{line_number}_brobot2_movement_pause",
                            f"Interview line {line_number} Brobot 2 breathing pause after {brobot_2_movement}",
                        )
                    brobot_2_movement_2 = event.get("brobot_2_movement_2")
                    if brobot_2_movement_2:
                        status(f"INTERVIEW_MOVEMENT_START line={line_number} speaker=Brobot 2 movement={brobot_2_movement_2}")
                        event["brobot_2_movement_2_result"] = fire_scored_interview_movement(
                            brobot_2_movement_2,
                            BROBOT_2_SERIAL,
                            live,
                            segment_id=f"stage3_line{line_number}_brobot2_movement2",
                            segment_label=f"Interview line {line_number} Brobot 2 second movement ({brobot_2_movement_2})",
                        )
                        status(f"INTERVIEW_MOVEMENT_DONE line={line_number} speaker=Brobot 2 movement={brobot_2_movement_2}")
                        pause_after_movement(
                            f"stage3_line{line_number}_brobot2_movement2_pause",
                            f"Interview line {line_number} Brobot 2 breathing pause after {brobot_2_movement_2}",
                        )
                    with TIMING.record(
                        f"stage3_line{line_number}_brobot2_speech",
                        f"Interview line {line_number} Brobot 2 speech dispatch",
                        speech_quality,
                        note=speech_note,
                    ):
                        brobot_2_result = play_generated_turn(robots, event["brobot_2_generated"], "BROBOT2_SPEECH_FAILURE", interview_scaffold)
                    event["brobot_2_result"] = brobot_2_result
                    if not brobot_2_result.get("speech_committed"):
                        event["failure"] = "BROBOT2_SPEECH_FAILURE"
                        write_log(log)
                        raise RuntimeError("BROBOT2_SPEECH_FAILURE")
                    status(f"PLAYBACK_TURN_DONE line={line_number} speaker=Brobot 2")

                def play_brobot_1():
                    status(f"PLAYBACK_TURN_START line={line_number} speaker=Brobot 1")
                    brobot_1_movement = event.get("brobot_1_movement")
                    if brobot_1_movement:
                        status(f"INTERVIEW_MOVEMENT_START line={line_number} speaker=Brobot 1 movement={brobot_1_movement}")
                        event["brobot_1_movement_result"] = fire_scored_interview_movement(
                            brobot_1_movement,
                            BROBOT_1_SERIAL,
                            live,
                            segment_id=f"stage3_line{line_number}_brobot1_movement",
                            segment_label=f"Interview line {line_number} Brobot 1 movement ({brobot_1_movement})",
                        )
                        status(f"INTERVIEW_MOVEMENT_DONE line={line_number} speaker=Brobot 1 movement={brobot_1_movement}")
                        pause_after_movement(
                            f"stage3_line{line_number}_brobot1_movement_pause",
                            f"Interview line {line_number} Brobot 1 breathing pause after {brobot_1_movement}",
                        )
                    brobot_1_movement_2 = event.get("brobot_1_movement_2")
                    if brobot_1_movement_2:
                        status(f"INTERVIEW_MOVEMENT_START line={line_number} speaker=Brobot 1 movement={brobot_1_movement_2}")
                        event["brobot_1_movement_2_result"] = fire_scored_interview_movement(
                            brobot_1_movement_2,
                            BROBOT_1_SERIAL,
                            live,
                            segment_id=f"stage3_line{line_number}_brobot1_movement2",
                            segment_label=f"Interview line {line_number} Brobot 1 second movement ({brobot_1_movement_2})",
                        )
                        status(f"INTERVIEW_MOVEMENT_DONE line={line_number} speaker=Brobot 1 movement={brobot_1_movement_2}")
                        pause_after_movement(
                            f"stage3_line{line_number}_brobot1_movement2_pause",
                            f"Interview line {line_number} Brobot 1 breathing pause after {brobot_1_movement_2}",
                        )
                    if kg_search_song:
                        section_line = find_section_card_line(log, line_number)
                        if resolve_turn_mode("BROBOT 1", section_line, interview_scaffold) == "llm":
                            status(f"INTERVIEW_KG_SEARCH_START line={line_number} speaker=Brobot 1")
                            event["brobot_1_kg_search_result"] = fire_interview_kg_search_sequence(
                                BROBOT_1_SERIAL,
                                live,
                                segment_id=f"stage3_line{line_number}_kg_search",
                                segment_label=f"Interview line {line_number} KG search sequence (Brobot 1)",
                            )
                            status(f"INTERVIEW_KG_SEARCH_DONE line={line_number} speaker=Brobot 1")
                            _, kg_spoken = split_robot_safe_line(
                                event["brobot_1_generated"]["robot_safe_line"], interview_scaffold
                            )
                            event["brobot_1_generated"]["robot_safe_line"] = robot_safe(
                                kg_spoken, "answering", interview_scaffold
                            )
                            event["brobot_1_generated"]["animation_action"] = "answering"
                    with TIMING.record(
                        f"stage3_line{line_number}_brobot1_speech",
                        f"Interview line {line_number} Brobot 1 speech dispatch",
                        speech_quality,
                        note=speech_note,
                    ):
                        brobot_1_result = play_generated_turn(robots, event["brobot_1_generated"], "BROBOT1_SPEECH_FAILURE", interview_scaffold)
                    event["brobot_1_result"] = brobot_1_result
                    if not brobot_1_result.get("speech_committed"):
                        event["failure"] = "BROBOT1_SPEECH_FAILURE"
                        write_log(log)
                        raise RuntimeError("BROBOT1_SPEECH_FAILURE")
                    status(f"PLAYBACK_TURN_DONE line={line_number} speaker=Brobot 1")

                if first_speaker == "brobot_1":
                    play_brobot_1()
                    play_brobot_2()
                else:
                    play_brobot_2()
                    play_brobot_1()
                event["playback_phase"] = "complete"
                event["speech_commit_confirmed"] = True
                event["speech_commit_signal"] = "wirepod backpack route accepted"
                status(f"EXCHANGE_PLAYBACK_DONE line={line_number}")
            pause_seconds = float(interview_scaffold.get("between_exchange_pause_seconds", 0) or 0)
            if pause_seconds > 0 and event is not log["events"][-1]:
                status(f"EXCHANGE_PAUSE seconds={pause_seconds}")
                with TIMING.record(
                    f"stage3_pause_after_line{line_number}",
                    f"Configured between-exchange pause after line {line_number} (between_exchange_pause_seconds, runner_enforced)",
                    "true_completion",
                    note="Deterministic time.sleep() - configured control-knob pause, not variable show time.",
                ):
                    time.sleep(pause_seconds)
        log["events"].append({"type": "stop", "reason": "section_1_complete"})
        status("RUN_DONE")
        return write_log(log)
    finally:
        robots.close()


def _preshow_speak_host(step, read_sheet=False, segment_id=None):
    speaker = step["speaker"]
    text = step["text"]
    if speaker not in ("brobot_3", "brobot_4"):
        raise RuntimeError(f"BLOCKED: preshow beat {step['step_id']!r} names unknown host speaker {speaker!r}")
    segment_id = segment_id or f"stage2_{step['step_id']}"
    if read_sheet:
        quality, note = "true_completion", "read_sheet mode: no audio spoken, elapsed reflects code path only."
    else:
        quality, note = (
            "true_completion",
            "Blocks on the speaking host's own Kokoro speak-stdin worker DONE ack, which only "
            "fires after real audio playback finishes.",
        )
    with TIMING.record(segment_id, f"Pre-show {step['step_id']} ({speaker}): {text[:60]}", quality, note=note) as span:
        if read_sheet:
            result = {"spoken": False, "mode": "read_sheet"}
        elif speaker == "brobot_3":
            result = speak_turn_via_kokoro(text)
        else:
            result = speak_host4_line_via_kokoro(text)
        span["spoken"] = result.get("spoken")
    print(f"PRESHOW {step['step_id']}: {speaker}: {text}", flush=True)
    return result


def _preshow_speak_robot(step, robots, serial, name, interview_scaffold, content_model, read_sheet=False):
    persona_prompt = interview_scaffold.get("brobots_1_2_persona", {}).get("system_prompt", "")
    with TIMING.record(
        f"stage2_{step['step_id']}_llm_colour",
        f"Pre-show {step['step_id']} ({name}) readiness-line LLM colour-pass",
        "true_completion",
        note="A real blocking LLM chat-completion call.",
    ):
        raw = llm_colour_preshow_line(persona_prompt, step["hint"], content_model)
    robot_safe_line, _repair_used = normalize_robot_safe(raw, interview_scaffold)
    _, raw_spoken = split_robot_safe_line(robot_safe_line, interview_scaffold)
    voice_text = flatten_for_robot_speech(raw_spoken, interview_scaffold)
    if read_sheet:
        speech_quality, speech_note = "true_completion", "read_sheet mode: no speech dispatched, elapsed reflects code path only."
    else:
        speech_quality, speech_note = _speech_capture_quality(robots.voice_destination)
    with TIMING.record(
        f"stage2_{step['step_id']}_speech",
        f"Pre-show {step['step_id']} ({name}) readiness-line speech dispatch",
        speech_quality,
        note=speech_note,
    ):
        if read_sheet:
            result = {"spoken": False, "mode": "read_sheet"}
        else:
            result = robots.say(name, robot_safe_line, display_text=voice_text, display_role=name)
    print(f"PRESHOW {step['step_id']}: {name}: {robot_safe_line}", flush=True)
    return result, voice_text


def format_preshow_read_sheet(song_id, run_id, content_model, entries):
    lines = [
        f"# Pre-show read-sheet — {song_id or '(unknown song)'}",
        f"run_id: {run_id}",
        f"content_model: {content_model or '(none)'}",
        f"generated_utc: {time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime())}",
        "",
    ]
    for entry in entries:
        lines.append(f"Beat {entry['step_id']} — {entry['role']}")
        lines.append(f"HINT:  {entry['hint'] or '(none - canned line, spoken verbatim)'}")
        lines.append(f"SPOKE: {entry['spoke']}")
        lines.append("")
    return "\n".join(lines)


def write_preshow_read_sheet(song_id, run_id, content_model, entries):
    GOPOD_NOTES_DIR.mkdir(parents=True, exist_ok=True)
    path = GOPOD_NOTES_DIR / f"PRESHOW_READSHEET_{run_id}.md"
    path.write_text(format_preshow_read_sheet(song_id, run_id, content_model, entries), encoding="utf-8")
    print(f"PRESHOW_READ_SHEET_WRITTEN path={path}", flush=True)
    return path


def run_preshow_song(song_dir, interview_scaffold, robots, content_model, generation_done_event, read_sheet=False, read_sheet_entries=None):
    """Plays the scored two-host pre-show: M1 cold open, M2 Brobot 1 wakes,
    M3 Brobot 2 wakes, a vamp gate that loops scored filler beats (never
    silence, never a hardcoded string) until generation_done_event is set,
    then M4. Never touches log/run-json state - generate_phase() (running
    concurrently, in its own thread) owns that exclusively; this function
    only ever reads the pre-show's own song folder and writes audio (or, in
    read-sheet mode, writes nothing but a print + a read_sheet_entries
    record). Returns the song_id, so a read-sheet caller doesn't need a
    second, separate load_preshow_song() call just for that string."""
    song = load_preshow_song(song_dir)
    by_movement = {}
    for step in song["steps"]:
        by_movement.setdefault(step["movement"], []).append(step)

    def record(step, spoke_text):
        if read_sheet_entries is None:
            return
        step_id = step["step_id"]
        if any(existing["step_id"] == step_id for existing in read_sheet_entries):
            return
        read_sheet_entries.append(
            {
                "step_id": step_id,
                "role": step["speaker"].replace("_", " ").title(),
                "hint": step["hint"],
                "spoke": spoke_text,
            }
        )

    def play_movement(movement_key):
        for step in by_movement.get(movement_key, []):
            speaker = step["speaker"]
            if speaker == "brobot_1":
                _, raw = _preshow_speak_robot(
                    step, robots, BROBOT_1_SERIAL, "Brobot 1", interview_scaffold, content_model, read_sheet
                )
                record(step, raw)
            elif speaker == "brobot_2":
                _, raw = _preshow_speak_robot(
                    step, robots, BROBOT_2_SERIAL, "Brobot 2", interview_scaffold, content_model, read_sheet
                )
                record(step, raw)
            else:
                _preshow_speak_host(step, read_sheet)
                record(step, step["text"])

    play_movement("m1")
    play_movement("m2")
    play_movement("m3")

    vamp_steps = by_movement.get("vamp", [])
    vamp_index = 0
    vamp_rounds = 0
    vamp_loop_start = time.monotonic()
    while not generation_done_event.is_set():
        if not vamp_steps:
            generation_done_event.wait(timeout=1)
            continue
        step = vamp_steps[vamp_index % len(vamp_steps)]
        vamp_rounds += 1
        _preshow_speak_host(step, read_sheet, segment_id=f"stage2_vamp_round{vamp_rounds}_{step['step_id']}")
        record(step, step["text"])
        vamp_index += 1
    TIMING.note_summary(
        "stage2_vamp_loop_summary",
        f"Pre-show vamp loop total ({vamp_rounds} round(s) this run)",
        "true_completion",
        round(time.monotonic() - vamp_loop_start, 6),
        note=(
            "Whole vamp-gate wall time (near-zero if generation finished before the gate was "
            "first checked). vamp_rounds is the real, run-dependent repeat count captured here "
            "- never a fixed/assumed number; see the per-round stage2_vamp_round{N}_* segments "
            "for individual beat timing."
        ),
        vamp_rounds=vamp_rounds,
    )
    print(f"PRESHOW_VAMP_ROUNDS {vamp_rounds}", flush=True)

    play_movement("m4")
    return song["song_id"]


def run_scored_preshow_and_generate(song_dir, live, voice_destination, read_sheet=False):
    """Runs the scored pre-show song and interview generation genuinely in
    parallel: generation happens in a background thread, calling
    generate_phase() completely unchanged - same function, same log dict,
    same single write_log() call at the end, in that one thread alone. The
    main thread plays the pre-show song and only ever touches a
    threading.Event for synchronization - it never reads or writes `log` or
    the run json, so there is nothing for the two threads to race on there.

    Both Kokoro voices are warmed up front, synchronously, before the
    background thread starts, specifically to close a real race: Brobot 3's
    persistent subprocess is shared state (_KOKORO_STATUS_STATE), and
    _kokoro_status_ensure_loaded()'s own "attempted" guard is not atomic
    across threads. Warming it here first means generate_phase()'s own
    internal warm-up call later is a guaranteed no-op by the time the
    background thread reaches it - never a second spawn, never a torn write.

    _KOKORO_STATUS_STATE["active"] is set False before the background thread
    starts, for the same reason from the other direction: left True,
    generate_phase()'s own status() beats would call _announce_status_kokoro,
    which writes to that same shared subprocess from the background thread,
    while this thread is writing to it too for the song's own Brobot 3
    lines - two threads writing one pipe. False means every one of those
    calls no-ops immediately (the technical print(f"STATUS: ...") still
    happens - operator-facing telemetry, unchanged); only this thread's own
    calls ever reach the subprocess."""
    content_model = resolve_content_model()
    if content_model:
        print(f"CONTENT_MODEL_SELECTED model={content_model}", flush=True)
    interview_scaffold = load_interview_scaffold(INTERVIEW_SCAFFOLD_PATH)
    robots = Robots(live, interview_scaffold, voice_destination)

    configure_voice_engines(interview_scaffold)
    with TIMING.record(
        "stage1_kokoro_status_warmup",
        "Brobot 3 Kokoro status-voice model warm-up",
        "true_completion",
        note="Blocks on the speak-stdin worker's ack after KPipeline() finishes loading (~20s cold, near-zero if already warm).",
    ):
        kokoro_status_warm_up()
    with TIMING.record(
        "stage1_kokoro_host4_warmup",
        "Brobot 4 Kokoro voice subprocess spawn",
        "accept_only",
        note=(
            "Subprocess spawn only - no ack/model-load confirmation exists for this second "
            "worker (unlike Brobot 3's kokoro_status_warm_up), so elapsed is process-launch "
            "time only, not a confirmed model-ready signal."
        ),
    ):
        _kokoro_host4_ensure_loaded()
    _KOKORO_STATUS_STATE["active"] = False

    generation_done_event = threading.Event()
    result_holder = {}
    error_holder = {}

    def _run_generation():
        try:
            result_holder["value"] = generate_phase(
                live=live, voice_destination=voice_destination, content_model=content_model
            )
        except Exception as exc:  # noqa: BLE001 - re-raised on the main thread after join, not swallowed.
            error_holder["error"] = exc
        finally:
            generation_done_event.set()

    generation_thread = threading.Thread(target=_run_generation, name="preshow-generation")
    generation_thread.start()

    read_sheet_entries = [] if read_sheet else None
    song_id = run_preshow_song(
        song_dir, interview_scaffold, robots, content_model, generation_done_event, read_sheet, read_sheet_entries
    )

    generation_thread.join()
    if "error" in error_holder:
        raise error_holder["error"]

    if read_sheet:
        generation_log = result_holder["value"][0]
        run_id = generation_log.get("run_id") or utc_run_id()
        write_preshow_read_sheet(song_id, run_id, content_model, read_sheet_entries)

    return result_holder["value"]


def run_preshow_only(song_dir, live, voice_destination, read_sheet=False):
    """Plays ONLY the scored pre-show/vamp song (M1 cold open, M2 Brobot 1
    wakes, M3 Brobot 2 wakes, M4) with ZERO interview generation triggered -
    the pure "video 1" performance, added 2026-08-19
    (INTERVIEW_VAMP_NO_GEN_PATH_001.md) once the operator called for a
    no-generation-side-effect fire path. Reuses run_preshow_song() verbatim
    - the exact same function run_scored_preshow_and_generate() calls, no
    new playback logic. The only difference: no generation thread is ever
    spawned, and generation_done_event is created already-set, so
    run_preshow_song()'s own vamp-filler loop
    (`while not generation_done_event.is_set()`) never enters at all -
    correct, since the filler's whole purpose is covering an active
    generation wait, and there is none here. Zero vamp filler rounds play;
    M1/M2/M3 play once, then straight to M4.

    Still reads the shared interview_scaffold (persona/pronunciation rules)
    for Brobot 1/2's two llm_coloured wake-beat lines - a read of a shared
    resource, not interview generation. generate_phase() is never imported,
    referenced, or called anywhere in this function - confirmed by grep,
    not just by this docstring's own claim."""
    content_model = resolve_content_model()
    if content_model:
        print(f"CONTENT_MODEL_SELECTED model={content_model}", flush=True)
    interview_scaffold = load_interview_scaffold(INTERVIEW_SCAFFOLD_PATH)
    robots = Robots(live, interview_scaffold, voice_destination)

    configure_voice_engines(interview_scaffold)
    with TIMING.record(
        "stage1_kokoro_status_warmup",
        "Brobot 3 Kokoro status-voice model warm-up",
        "true_completion",
        note="Blocks on the speak-stdin worker's ack after KPipeline() finishes loading (~20s cold, near-zero if already warm).",
    ):
        kokoro_status_warm_up()
    with TIMING.record(
        "stage1_kokoro_host4_warmup",
        "Brobot 4 Kokoro voice subprocess spawn",
        "accept_only",
        note=(
            "Subprocess spawn only - no ack/model-load confirmation exists for this second "
            "worker (unlike Brobot 3's kokoro_status_warm_up), so elapsed is process-launch "
            "time only, not a confirmed model-ready signal."
        ),
    ):
        _kokoro_host4_ensure_loaded()
    _KOKORO_STATUS_STATE["active"] = False

    generation_done_event = threading.Event()
    generation_done_event.set()

    read_sheet_entries = [] if read_sheet else None
    song_id = run_preshow_song(
        song_dir, interview_scaffold, robots, content_model, generation_done_event, read_sheet, read_sheet_entries
    )

    if read_sheet:
        write_preshow_read_sheet(song_id, utc_run_id(), content_model, read_sheet_entries)

    return song_id


def main():
    voice_destination = resolve_voice_destination()
    live = (os.getenv(LIVE_GATE) == "1") and (voice_destination == "robot")
    _KOKORO_STATUS_STATE["live"] = live
    _KOKORO_STATUS_STATE["voice_destination"] = voice_destination
    _wirepod_log_marker(GOPOD_SONG_START_ROLE, "Song starting")
    run_context = {}
    try:
        preshow_song_dir = os.getenv(PRESHOW_SONG_DIR_ENV, "").strip()
        read_sheet_mode = os.getenv(PRESHOW_READ_SHEET_ENV, "0") == "1"
        if read_sheet_mode and not preshow_song_dir:
            raise RuntimeError(
                f"BLOCKED: {PRESHOW_READ_SHEET_ENV}=1 requires {PRESHOW_SONG_DIR_ENV} to also be set - "
                "read-sheet mode has no pre-show song to read without one"
            )
        if preshow_song_dir:
            log, interview_scaffold, robots = run_scored_preshow_and_generate(
                preshow_song_dir, live, voice_destination, read_sheet_mode
            )
        else:
            log, interview_scaffold, robots = generate_phase(live=live, voice_destination=voice_destination)
        run_context["run_id"] = log.get("run_id")
        run_context["song_dir"] = log.get("source_card")
        run_context["bare_capture"] = log.get("bare_capture", False)
        return playback_phase(log, interview_scaffold, robots, log.get("live_enabled", False))
    finally:
        _wirepod_log_marker(GOPOD_SONG_END_ROLE, "Song ending")
        if run_context.get("bare_capture"):
            print("TIMING_LOG_SKIPPED bare_capture=true", flush=True)
        else:
            timing_path = write_timing_log(
                run_id=run_context.get("run_id"),
                song_dir=run_context.get("song_dir"),
                meta={
                    "tool": "run_section1_full_live_001.main",
                    "voice_destination": voice_destination,
                    "live_enabled": live,
                    "preshow_song_dir": os.getenv(PRESHOW_SONG_DIR_ENV, "").strip() or None,
                },
            )
            print(f"TIMING_LOG_WRITTEN path={timing_path}", flush=True)


if __name__ == "__main__":
    main()


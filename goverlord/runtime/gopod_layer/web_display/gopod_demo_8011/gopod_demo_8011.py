from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
import argparse
from datetime import datetime, timezone
import json
import mimetypes
import os
import subprocess
import sys
import threading
import time
import urllib.parse
import urllib.request


DISPLAY_DIR = Path(__file__).resolve().parent
REPO_ROOT = DISPLAY_DIR.parents[3]

# STAGE PINNACLE: facts read from goverlord/runtime/data_gomad/configs/ instead of hardcoded here.
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))
from goverlord.runtime.data_gomad.configs.loader import ENDPOINTS_PATH, load_json, validate_envelope  # noqa: E402

_ENDPOINTS = load_json(ENDPOINTS_PATH)

PORT = _ENDPOINTS["cockpit"]["port"]
LIVE_CHAT_MESSAGES = DISPLAY_DIR / "live_chat_messages.json"
STATE_DIR = DISPLAY_DIR / "state"
QR_TARGETS = DISPLAY_DIR / "qr_targets_001.json"
ROUTE_STATE = DISPLAY_DIR / "sample_route_state.json"
CAPTIONS_STATE = DISPLAY_DIR / "sample_captions.json"
GOPOD_LOG_DIR = REPO_ROOT / "goverlord/runtime/gopod_layer/gopod_yourself/logs"
GOPOD_SWARM_REGISTRY_PATH = REPO_ROOT / "goverlord/runtime/gopod_layer/gopod_yourself/gopod_yourself_swarm_participant_registry_001.json"
JETSON_PREFLIGHT_PATH = REPO_ROOT / "goverlord/runtime/system_preflight/jetson_audio_power_preflight_001.py"
DEFAULT_WIREPOD_BASE_URL = _ENDPOINTS["wirepod"]["base_url"]
DEFAULT_WIREPOD_TIMEOUT_SECONDS = 20.0
WIREPOD_LOG_ENDPOINTS = [
    ("standard", "/api/get_logs"),
    ("debug", "/api/get_debug_logs"),
]
WIREPOD_LOG_LINE_LIMIT_PER_ENDPOINT = 20
WIREPOD_LOG_EVENT_LIMIT = 44
GOLDEN_TIMING = {
    "post_say_seconds": 7.0,
    "between_calls_seconds": 4.0,
    "settle_seconds": 1.0,
    "timeout_seconds": 20.0,
}
DEMO_FAST_TIMING = {
    "post_say_seconds": 3.0,
    "between_calls_seconds": 1.0,
    "settle_seconds": 1.0,
    "timeout_seconds": 20.0,
}
READY_CHORUS_ROBOTS = [
    {"event": "gopod_brobot_1_ready_status", "key": "KP1", "name": "Doc", "robot": "vector1", "serial": "0dd1b9e9", "line": "Brobot one ready."},
    {"event": "gopod_brobot_2_ready_status", "key": "KP2", "name": "Pip", "robot": "vector2", "serial": "0dd1d8bf", "line": "Brobot two ready."},
]
DEFAULT_CHAT_MESSAGES = {
    "proof_id": "GOPOD_LIVE_CHAT_DISPLAY_012",
    "source_mode": "file_backed_live_chat",
    "messages": [
        {
            "speaker_id": "operator",
            "display_name": "Operator",
            "text": "GOPOD live chat file is being served from /ptt/multichat.json.",
            "timestamp": "2026-05-31T00:00:00Z",
            "source_producer": "live_sample",
        },
        {
            "speaker_id": "system",
            "display_name": "System",
            "text": "Append or rewrite live_chat_messages.json and Pane 2 updates on the next poll.",
            "timestamp": "2026-05-31T00:00:01Z",
            "source_producer": "live_sample",
        },
    ],
    "updated_at": "2026-05-31T00:00:01Z",
}
EMPTY_CHAT_MESSAGES = {
    "proof_id": "GOPOD_LIVE_CHAT_DISPLAY_012",
    "source_mode": "file_backed_live_chat",
    "messages": [],
    "updated_at": "",
}
ROUTES = {
    "/": DISPLAY_DIR / "gopod_demo.html",
    "/gopod_demo.html": DISPLAY_DIR / "gopod_demo.html",
    "/index.html": DISPLAY_DIR / "gopod_demo.html",
    "/gopod_demo.css": DISPLAY_DIR / "gopod_demo.css",
    "/sample_captions.json": DISPLAY_DIR / "sample_captions.json",
    "/sample_multichat.json": DISPLAY_DIR / "sample_multichat.json",
    "/sample_route_state.json": DISPLAY_DIR / "sample_route_state.json",
    "/live_chat_messages.json": LIVE_CHAT_MESSAGES,
    "/ptt/captions.json": DISPLAY_DIR / "sample_captions.json",
    "/ptt/multichat.json": LIVE_CHAT_MESSAGES,
    "/ptt/route_state.json": ROUTE_STATE,
    "/ptt/state.json": CAPTIONS_STATE,
    "/qr_targets_001.json": QR_TARGETS,
    "/qr_cache/qr1_001.png": DISPLAY_DIR / "qr_cache/qr1_001.png",
    "/qr_cache/qr2_001.png": DISPLAY_DIR / "qr_cache/qr2_001.png",
}

RAW_CAMERA_FEEDS = {
    "/raw/doc.jpg": {
        "source_id": "doc_robot",
        "path": DISPLAY_DIR / "live_cache/doc_robot.jpg",
    },
    "/raw/pip.jpg": {
        "source_id": "pip_robot",
        "path": DISPLAY_DIR / "live_cache/pip_robot.jpg",
    },
    "/raw/c920.jpg": {
        "source_id": "usb_c920",
        "path": DISPLAY_DIR / "live_cache/usb_c920.jpg",
    },
}

PLACEHOLDER_JPEG = bytes.fromhex(
    "ffd8ffe000104a46494600010101006000600000ffdb0043000302020302020303030304030304050805050404050a070706080c0a0c0c0b0a0b0b0d0e12100d0e110e0b0b1016101113141515150c0f171816141812141514ffc0000b080001000101011100ffc4001400010000000000000000000000000000000000000000ffc4001410010000000000000000000000000000000000000000ffda0008010100003f00d2cf20ffd9"
)


class TeeStream:
    def __init__(self, *streams):
        self.streams = streams

    def write(self, data):
        for stream in self.streams:
            stream.write(data)
            stream.flush()
        return len(data)

    def flush(self):
        for stream in self.streams:
            stream.flush()


def setup_terminal_log_capture():
    GOPOD_LOG_DIR.mkdir(parents=True, exist_ok=True)
    stamp = time.strftime("%Y%m%d_%H%M%S", time.gmtime())
    log_path = GOPOD_LOG_DIR / f"gopod_demo_{stamp}.log"
    log_handle = log_path.open("a", encoding="utf-8", buffering=1)
    sys.stdout = TeeStream(sys.__stdout__, log_handle)
    sys.stderr = TeeStream(sys.__stderr__, log_handle)
    print(f"GOPOD_DEMO_LOG_PATH={log_path}", flush=True)
    return log_handle


def utc_now():
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")


def read_json_object(path):
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return {}
    return payload if isinstance(payload, dict) else {}


def read_live_chat_payload():
    payload = read_json_object(LIVE_CHAT_MESSAGES)
    messages = payload.get("messages", [])
    if not isinstance(messages, list):
        messages = []
    payload["messages"] = [message for message in messages if isinstance(message, dict)]
    return payload


def live_chat_state_payload():
    payload = read_live_chat_payload()
    messages = payload.get("messages", [])
    return {
        **payload,
        "event": "chat_state",
        "status": "CONNECTED" if messages else "READY_EMPTY",
        "messages": messages,
        "session_messages": messages,
        "latest_participant": messages[-1].get("display_name", "") if messages else "",
        "latest_message": messages[-1].get("text", "") if messages else "",
    }


def current_route_state_payload():
    route = read_json_object(ROUTE_STATE)
    route["event"] = "route_state"
    route["status"] = route.get("route_status") or ("READY" if route else "STATE_FILE_NOT_FOUND")
    return route


def current_ptt_state_payload():
    payload = read_json_object(CAPTIONS_STATE)
    payload["event"] = "ptt_state"
    payload["status"] = payload.get("ptt_state") or "READY"
    payload["caption_source"] = "/ptt/captions.json"
    return payload


def json_bytes(payload):
    return json.dumps(payload, indent=2, sort_keys=True).encode("utf-8")


def write_live_chat_payload(payload):
    payload["updated_at"] = utc_now()
    LIVE_CHAT_MESSAGES.parent.mkdir(parents=True, exist_ok=True)
    tmp_path = LIVE_CHAT_MESSAGES.with_suffix(".json.tmp")
    tmp_path.write_text(json.dumps(payload, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    tmp_path.replace(LIVE_CHAT_MESSAGES)


def ensure_live_chat_messages(*, write_samples=False):
    if LIVE_CHAT_MESSAGES.exists():
        log_event(
            "gopod_live_chat_startup_state",
            "LIVE_FILE_PRESENT",
            path=str(LIVE_CHAT_MESSAGES.relative_to(REPO_ROOT)),
            sample_written=False,
        )
        return "LIVE_FILE_PRESENT"
    if not write_samples:
        log_event(
            "gopod_live_chat_startup_state",
            "AWAITING_REAL_EVENTS",
            path=str(LIVE_CHAT_MESSAGES.relative_to(REPO_ROOT)),
            sample_written=False,
        )
        return "AWAITING_REAL_EVENTS"
    DEFAULT_CHAT_MESSAGES["updated_at"] = utc_now()
    valid_messages = []
    for message in DEFAULT_CHAT_MESSAGES["messages"]:
        try:
            validate_envelope(message)
        except ValueError as exc:
            log_event("GOPOD_ENVELOPE_INVALID", "SKIPPED", reason=str(exc))
            continue
        valid_messages.append(message)
    seed_payload = {**DEFAULT_CHAT_MESSAGES, "messages": valid_messages}
    LIVE_CHAT_MESSAGES.write_text(json.dumps(seed_payload, indent=2) + "\n", encoding="utf-8")
    log_event(
        "gopod_live_chat_startup_state",
        "VALIDATION_SAMPLE_WRITTEN",
        path=str(LIVE_CHAT_MESSAGES.relative_to(REPO_ROOT)),
        sample_written=True,
    )
    return "VALIDATION_SAMPLE_WRITTEN"


def wirepod_fetch_text(url, timeout):
    request = urllib.request.Request(url, method="GET")
    with urllib.request.urlopen(request, timeout=timeout) as response:
        charset = response.headers.get_content_charset() or "utf-8"
        return response.read().decode(charset, errors="replace"), int(response.status)


def flatten_wirepod_log_payload(payload):
    if isinstance(payload, str):
        return payload.splitlines()
    if isinstance(payload, list):
        lines = []
        for item in payload:
            lines.extend(flatten_wirepod_log_payload(item))
        return lines
    if isinstance(payload, dict):
        for key in ("logs", "debug_logs", "log", "debugLog", "messages", "data", "body"):
            if key in payload:
                return flatten_wirepod_log_payload(payload[key])
        return [json.dumps(payload, sort_keys=True)]
    return [str(payload)]


def wirepod_log_lines(raw_text):
    stripped = raw_text.strip()
    if not stripped:
        return []
    try:
        parsed = json.loads(stripped)
    except json.JSONDecodeError:
        parsed = stripped
    lines = [" ".join(line.strip().split()) for line in flatten_wirepod_log_payload(parsed)]
    return [line for line in lines if line]


def build_wirepod_log_event(*, log_kind, text, fetched_at):
    return {
        "speaker_id": "wire_pod",
        "display_name": "Wire-Pod",
        "text": text,
        "timestamp": fetched_at,
        "source_producer": f"wirepod_{log_kind}_log_prefill",
    }


def prefill_wirepod_logs(args):
    base_url = os.environ.get("GOPOD_WIREPOD_BASE_URL", DEFAULT_WIREPOD_BASE_URL).strip() or DEFAULT_WIREPOD_BASE_URL
    timeout = max(0.25, float(getattr(args, "wirepod_timeout", DEFAULT_WIREPOD_TIMEOUT_SECONDS)))
    fetched_at = utc_now()
    events = []
    fetch_statuses = []
    for log_kind, endpoint in WIREPOD_LOG_ENDPOINTS:
        url = f"{base_url.rstrip('/')}{endpoint}"
        try:
            raw_text, http_status = wirepod_fetch_text(url, timeout)
            lines = wirepod_log_lines(raw_text)[-WIREPOD_LOG_LINE_LIMIT_PER_ENDPOINT:]
            fetch_statuses.append(
                {
                    "kind": log_kind,
                    "url": url,
                    "http_status": http_status,
                    "line_count": len(lines),
                    "status": "PASS",
                }
            )
            for line in lines:
                event = build_wirepod_log_event(
                    log_kind=log_kind,
                    text=line,
                    fetched_at=fetched_at,
                )
                try:
                    validate_envelope(event)
                except ValueError as exc:
                    log_event("GOPOD_ENVELOPE_INVALID", "SKIPPED", reason=str(exc))
                    continue
                events.append(event)
        except Exception as exc:  # noqa: BLE001 - startup should report evidence and keep serving.
            fetch_statuses.append(
                {
                    "kind": log_kind,
                    "url": url,
                    "status": "BLOCKED",
                    "error": f"{type(exc).__name__}: {exc}",
                }
            )

    if events:
        payload = read_live_chat_payload()
        payload["proof_id"] = "GOPOD_WIREPOD_LOG_PREFILL_001"
        payload["source_mode"] = "gopod_wirepod_log_prefill_plus_live_chat"
        existing_messages = [
            message
            for message in payload.get("messages", [])
            if "log_prefill" not in str(message.get("source_producer", ""))
        ]
        payload["messages"] = existing_messages + events[-WIREPOD_LOG_EVENT_LIMIT:]
        write_live_chat_payload(payload)

    overall_status = "PASS" if events else "BLOCKED"
    log_event(
        "GOPOD_WIREPOD_LOG_PREFILL_STATUS",
        overall_status,
        base_url=base_url,
        chat_json_file=str(LIVE_CHAT_MESSAGES.relative_to(REPO_ROOT)),
        events_written=len(events),
        urls=[status["url"] for status in fetch_statuses],
        fetch_statuses=fetch_statuses,
        sample_prefill_events=events[:3],
    )
    return bool(events), events, fetch_statuses


# Channel-1 song brackets: base mode is the existing prefill-then-static
# LIVE_CHAT_MESSAGES file, completely untouched by anything below - this
# background watcher only ever affects what the HTTP routes below choose to
# serve, never LIVE_CHAT_MESSAGES itself, so base mode stays exactly what it
# is today whenever no song is active. See the channel-1 borrow-and-return
# note in gopod_notes for the seam this reads: server.go's say_text handler
# tags GOPOD_SONG_START/GOPOD_SONG_END and every real rich line as
# "BROBOT_RICH_DISPLAY <role>: <text>" in Wire-Pod's "show logs" feed
# (/api/get_logs), which run_section1_full_live_001.py's main() already
# brackets a real song with.
GOPOD_SONG_START_ROLE = "GOPOD_SONG_START"
GOPOD_SONG_END_ROLE = "GOPOD_SONG_END"
RICH_DISPLAY_PREFIX = "BROBOT_RICH_DISPLAY"
SONG_WATCH_POLL_SECONDS = 1.5
SONG_WATCH_FETCH_TIMEOUT_SECONDS = 3.0

_SONG_STATE_LOCK = threading.Lock()
_SONG_STATE = {"active": False, "messages": []}


def _parse_rich_display_line(line):
    # Exact inverse of server.go's own formatting: prefix + " " + role + ": " + text
    # (or just prefix + ": " + text if display_role was empty - never the case
    # for the two markers or a real turn/status line, but tolerated). Every
    # line in /api/get_logs also carries a "2006.01.02 15:04:05: " timestamp
    # server.go's own LogUI() unconditionally prepends (confirmed reading
    # logger.go) - found live during verification, not assumed - so search
    # for the marker rather than anchoring to line start.
    index = line.find(RICH_DISPLAY_PREFIX)
    if index == -1:
        return None
    rest = line[index + len(RICH_DISPLAY_PREFIX):].lstrip()
    role, sep, text = rest.partition(":")
    if not sep:
        return None
    return role.strip(), text.strip()


def _apply_song_watch(lines):
    # /api/get_logs returns Wire-Pod's current rolling buffer (capped at 50),
    # not an incremental tail, so state is derived structurally from order on
    # every poll rather than by diffing against a "last seen" cursor: find
    # the most recent START/END marker; whichever is later in the list wins.
    parsed = [_parse_rich_display_line(line) for line in lines]
    last_start_index = None
    last_end_index = None
    for index, result in enumerate(parsed):
        if result is None:
            continue
        role, _text = result
        if role == GOPOD_SONG_START_ROLE:
            last_start_index = index
        elif role == GOPOD_SONG_END_ROLE:
            last_end_index = index

    active = last_start_index is not None and (last_end_index is None or last_start_index > last_end_index)
    messages = []
    if active:
        for index in range(last_start_index + 1, len(parsed)):
            result = parsed[index]
            if result is None:
                continue
            role, text = result
            if role in (GOPOD_SONG_START_ROLE, GOPOD_SONG_END_ROLE):
                continue
            messages.append(
                {
                    "speaker_id": "gopod_song",
                    "display_name": role or "GOPOD Song",
                    "text": text,
                    "timestamp": utc_now(),
                    "source_producer": "gopod_song_rich_display",
                }
            )

    with _SONG_STATE_LOCK:
        _SONG_STATE["active"] = active
        _SONG_STATE["messages"] = messages


def song_state_snapshot():
    with _SONG_STATE_LOCK:
        return bool(_SONG_STATE["active"]), list(_SONG_STATE["messages"])


def _song_watch_loop():
    base_url = os.environ.get("GOPOD_WIREPOD_BASE_URL", DEFAULT_WIREPOD_BASE_URL).strip() or DEFAULT_WIREPOD_BASE_URL
    url = f"{base_url.rstrip('/')}/api/get_logs"
    while True:
        try:
            raw_text, _http_status = wirepod_fetch_text(url, SONG_WATCH_FETCH_TIMEOUT_SECONDS)
            _apply_song_watch(wirepod_log_lines(raw_text))
        except Exception as exc:  # noqa: BLE001 - background watcher must never crash the cockpit.
            log_event("GOPOD_SONG_WATCH_ERROR", "BLOCKED", error=f"{type(exc).__name__}: {exc}")
        time.sleep(SONG_WATCH_POLL_SECONDS)


def start_song_watch_thread():
    thread = threading.Thread(target=_song_watch_loop, name="gopod-song-watch", daemon=True)
    thread.start()
    return thread


def write_validation_samples(args):
    chat_status = ensure_live_chat_messages(write_samples=True)
    log_event(
        "GOPOD_VALIDATION_SAMPLES_STATUS",
        "PASS",
        live_chat_status=chat_status,
    )
    return True


def log_event(event, status, **fields):
    payload = {"event": event, "status": status, "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())}
    payload.update(fields)
    print(json.dumps(payload, sort_keys=True), flush=True)


def load_swarm_registry():
    try:
        payload = json.loads(GOPOD_SWARM_REGISTRY_PATH.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        payload = {}
    participants = payload.get("participants", {}) if isinstance(payload, dict) else {}
    log_event(
        "gopod_swarm_participant_registry_loaded",
        "OK" if participants else "MISSING",
        registry_path=str(GOPOD_SWARM_REGISTRY_PATH.relative_to(REPO_ROOT)),
        participant_keys=sorted(participants.keys()) if isinstance(participants, dict) else [],
    )
    return payload if isinstance(payload, dict) else {}


def wirepod_get(base_url, endpoint, params, timeout):
    url = f"{base_url.rstrip('/')}{endpoint}?{urllib.parse.urlencode(params)}"
    request = urllib.request.Request(url, method="GET")
    with urllib.request.urlopen(request, timeout=timeout) as response:
        response.read()
        return int(response.status)


def ready_chorus_enabled(args):
    return bool(args.ready_chorus) or os.environ.get("GOPOD_READY_CHORUS", "").strip().lower() in {"1", "true", "yes", "on"}


def ready_chorus_live_enabled(args):
    return bool(args.ready_chorus_live) or os.environ.get("GOPOD_READY_CHORUS_LIVE", "").strip().lower() in {"1", "true", "yes", "on"}


def demo_fast_enabled(args):
    return bool(getattr(args, "demo_fast", False)) or os.environ.get("GOPOD_DEMO_FAST", "").strip().lower() in {"1", "true", "yes", "on"}


def selected_timing_profile(args):
    profile = "demo_fast" if demo_fast_enabled(args) else "golden/default"
    timing = DEMO_FAST_TIMING if profile == "demo_fast" else GOLDEN_TIMING
    return profile, dict(timing)


def apply_timing_environment(args):
    profile, timing = selected_timing_profile(args)
    if profile == "demo_fast":
        os.environ["GOPOD_DEMO_FAST"] = "1"
        os.environ["GOPOD_PTT_DEMO_FAST"] = "1"
    log_event(
        "GOPOD_TIMING_ENVIRONMENT",
        "OK",
        profile=profile,
        demo_fast_timing=profile == "demo_fast",
        exported_gopod_demo_fast=os.environ.get("GOPOD_DEMO_FAST", ""),
        exported_gopod_ptt_demo_fast=os.environ.get("GOPOD_PTT_DEMO_FAST", ""),
        timing=timing,
    )
    return profile, timing


def log_timing_profile(args):
    profile, timing = selected_timing_profile(args)
    log_event("gopod_demo_timing_profile", "OK", profile=profile, **timing)
    log_event(
        "GOPOD_TIMING_PROFILE_SUMMARY",
        "OK",
        profile=profile,
        demo_fast_timing=profile == "demo_fast",
        actual_timing_values=timing,
        backend_env_pass_through={
            "GOPOD_DEMO_FAST": os.environ.get("GOPOD_DEMO_FAST", ""),
            "GOPOD_PTT_DEMO_FAST": os.environ.get("GOPOD_PTT_DEMO_FAST", ""),
        },
    )
    return timing


def run_warmup_preflight(args, *, live_gate=False):
    if not JETSON_PREFLIGHT_PATH.is_file():
        log_event(
            "GOPOD_WARMUP_PREFLIGHT_STATUS",
            "BLOCKED",
            preflight_path=str(JETSON_PREFLIGHT_PATH.relative_to(REPO_ROOT)),
            operator_notes=[
                "Set System Audio In to USB Mic.",
                "Set System Audio Out to Analog Output - Built-in Audio.",
                "Plug Jetson into power before live demo.",
            ],
        )
        return False
    command = [sys.executable, str(JETSON_PREFLIGHT_PATH), "--json"]
    if not live_gate:
        command.append("--skip-mic-probe")
    if getattr(args, "speaker_probe", False):
        command.append("--speaker-probe")
    if getattr(args, "fix_audio", False):
        command.append("--fix-audio")
    log_event(
        "GOPOD_WARMUP_PREFLIGHT_STATUS",
        "START",
        live_gate=live_gate,
        command=" ".join(command),
        output_included_in_gopod_demo_log=True,
    )
    completed = subprocess.run(command, capture_output=True, text=True, check=False)
    if completed.stdout:
        print(completed.stdout, end="", flush=True)
    if completed.stderr:
        print(completed.stderr, end="", file=sys.stderr, flush=True)
    status = "PASS" if completed.returncode == 0 else "BLOCKED"
    log_event("GOPOD_WARMUP_PREFLIGHT_STATUS", status, live_gate=live_gate, returncode=completed.returncode)
    return completed.returncode == 0


def run_ready_chorus(args):
    if not ready_chorus_enabled(args):
        log_event("gopod_ready_chorus_status", "DISABLED", config_gate="--ready-chorus or GOPOD_READY_CHORUS=1")
        return
    profile, timing = selected_timing_profile(args)
    live = ready_chorus_live_enabled(args)
    base_url = os.environ.get("GOPOD_WIREPOD_BASE_URL", DEFAULT_WIREPOD_BASE_URL).strip()
    timeout = float(timing["timeout_seconds"])
    settle_seconds = float(timing["settle_seconds"])
    post_say_seconds = float(timing["post_say_seconds"])
    between_calls_seconds = float(timing["between_calls_seconds"])
    sync_mode = "sequential_minimal_delay"
    statuses = []
    for robot in READY_CHORUS_ROBOTS:
        status = {
            "key": robot["key"],
            "name": robot["name"],
            "robot": robot["robot"],
            "serial": robot["serial"],
            "line": robot["line"],
            "mode": "live" if live else "dry_run",
            "wirepod_called": False,
            "assume_http_status": "DRY_NOT_CALLED",
            "say_http_status": "DRY_NOT_CALLED",
            "release_http_status": "DRY_NOT_CALLED",
            "ready_status": "DRY_VALIDATED",
            "timing_profile": profile,
            "demo_fast_timing": profile == "demo_fast",
            "settle_seconds": settle_seconds,
            "post_say_seconds": post_say_seconds,
            "between_calls_seconds": between_calls_seconds,
            "timeout_seconds": timeout,
        }
        if live:
            common = {"serial": robot["serial"]}
            try:
                status["assume_http_status"] = wirepod_get(base_url, "/api-sdk/assume_behavior_control", {**common, "priority": "high"}, timeout)
                time.sleep(settle_seconds)
                status["say_http_status"] = wirepod_get(base_url, "/api-sdk/say_text", {**common, "text": robot["line"]}, timeout)
                time.sleep(post_say_seconds)
                status["release_http_status"] = wirepod_get(base_url, "/api-sdk/release_behavior_control", common, timeout)
                status["wirepod_called"] = True
                status["ready_status"] = "LIVE_CALLED"
            except Exception as exc:  # noqa: BLE001 - startup proof should continue to report both robots.
                status["ready_status"] = "LIVE_BLOCKED"
                status["error"] = f"{type(exc).__name__}: {exc}"
                try:
                    status["release_http_status"] = wirepod_get(base_url, "/api-sdk/release_behavior_control", common, min(2.0, timeout))
                except Exception as release_exc:  # noqa: BLE001
                    status["release_http_status"] = f"RELEASE_ATTEMPT_FAILED:{type(release_exc).__name__}: {release_exc}"
        status["robot_calls"] = bool(status["wirepod_called"])
        log_event(robot["event"], status["ready_status"], **status)
        statuses.append(status)
        if live:
            time.sleep(between_calls_seconds)
    overall = "LIVE_CALLED" if live and all(item["ready_status"] == "LIVE_CALLED" for item in statuses) else "DRY_VALIDATED"
    if live and any(item["ready_status"] != "LIVE_CALLED" for item in statuses):
        overall = "LIVE_PARTIAL_OR_BLOCKED"
    log_event(
        "gopod_ready_chorus_status",
        overall,
        sync_mode=sync_mode,
        live=live,
        robot_calls=any(item["wirepod_called"] for item in statuses),
        robot_statuses=statuses,
    )
    log_event(
        "GOPOD_READY_CHORUS_SUMMARY",
        overall,
        live=live,
        sync_mode=sync_mode,
        timing_profile=profile,
        demo_fast_timing=profile == "demo_fast",
        timing=timing,
        robot_calls=any(item["wirepod_called"] for item in statuses),
        brobot_1_ready_line=READY_CHORUS_ROBOTS[0]["line"],
        brobot_2_ready_line=READY_CHORUS_ROBOTS[1]["line"],
    )


def content_type(path):
    if path.suffix == ".html":
        return "text/html; charset=utf-8"
    if path.suffix == ".css":
        return "text/css; charset=utf-8"
    if path.suffix == ".json":
        return "application/json; charset=utf-8"
    return mimetypes.guess_type(path.name)[0] or "application/octet-stream"


class Handler(BaseHTTPRequestHandler):
    def send_bytes(self, code, data, ctype):
        self.send_response(code)
        self.send_header("Content-Type", ctype)
        self.send_header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
        self.send_header("Pragma", "no-cache")
        self.send_header("Expires", "0")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        if self.command != "HEAD":
            self.wfile.write(data)

    def do_HEAD(self):
        return self.do_GET()

    def do_GET(self):
        path = self.path.split("?", 1)[0]

        if path == "/health.json":
            payload = {
                "status": "PASS",
                "lane": "goverlord/runtime/gopod_layer/web_display/gopod_demo_8011",
                "page": "gopod_demo.html",
                "data_mode": "file_backed_live_chat",
                "live_chat_messages": "/ptt/multichat.json",
                "route_state": "/state/route_state.json",
                "chat_state": "/state/chat_state.json",
                "ptt_state": "/state/ptt_state.json",
                "live_chat_file": str(LIVE_CHAT_MESSAGES),
            }
            return self.send_bytes(200, json_bytes(payload), "application/json; charset=utf-8")

        if path in {"/ptt/multichat.json", "/live_chat_messages.json"}:
            # Bracketed override only - base mode below is byte-identical to
            # what this route has always served. Only reachable when the
            # song watcher has seen a live GOPOD_SONG_START more recently
            # than any GOPOD_SONG_END; LIVE_CHAT_MESSAGES itself is never
            # written to by the watcher, so it always falls back intact.
            song_active, song_messages = song_state_snapshot()
            if song_active:
                payload = {
                    "proof_id": "GOPOD_LIVE_CHAT_DISPLAY_012",
                    "source_mode": "gopod_song_rich_display",
                    "messages": song_messages,
                    "updated_at": utc_now(),
                }
                return self.send_bytes(200, json_bytes(payload), "application/json; charset=utf-8")
            if LIVE_CHAT_MESSAGES.exists():
                return self.send_bytes(200, LIVE_CHAT_MESSAGES.read_bytes(), "application/json; charset=utf-8")
            return self.send_bytes(200, json_bytes(EMPTY_CHAT_MESSAGES), "application/json; charset=utf-8")

        state_payloads = {
            "/state/chat_state.json": live_chat_state_payload,
            "/state/route_state.json": current_route_state_payload,
            "/state/ptt_state.json": current_ptt_state_payload,
            "/state/mic_state.json": current_ptt_state_payload,
            "/state/camera_state.json": lambda: read_json_object(STATE_DIR / "camera_state.json"),
            "/state/weather_state.json": lambda: read_json_object(STATE_DIR / "weather_state.json"),
        }
        state_builder = state_payloads.get(path)
        if state_builder:
            return self.send_bytes(200, json_bytes(state_builder()), "application/json; charset=utf-8")

        raw_feed = RAW_CAMERA_FEEDS.get(path)
        if raw_feed:
            raw_path = raw_feed["path"]
            if raw_path.is_file():
                return self.send_bytes(200, raw_path.read_bytes(), "image/jpeg")
            return self.send_bytes(200, PLACEHOLDER_JPEG, "image/jpeg")

        target = ROUTES.get(path)
        if target is None:
            return self.send_bytes(404, b"not found", "text/plain; charset=utf-8")
        if not target.exists():
            payload = {
                "status": "BLOCKED",
                "missing_file": target.name,
                "path": str(target),
            }
            return self.send_bytes(200, json_bytes(payload), "application/json; charset=utf-8")
        return self.send_bytes(200, target.read_bytes(), content_type(target))

    def log_message(self, fmt, *args):
        pass


def main():
    parser = argparse.ArgumentParser(description="Serve the GOPOD demo page.")
    parser.add_argument("--preflight", action="store_true", help="Run Jetson audio/power warmup preflight and exit.")
    parser.add_argument("--fix-audio", action="store_true", help="Allow preflight to set exact matching PulseAudio defaults.")
    parser.add_argument("--speaker-probe", action="store_true", help="Allow preflight to play a short aplay tone.")
    parser.add_argument("--ready-chorus", action="store_true", help="Run optional Doc/Pip startup ready chorus.")
    parser.add_argument("--ready-chorus-live", action="store_true", help="Allow ready chorus Wire-Pod robot calls.")
    parser.add_argument("--demo-fast", action="store_true", help="Log conservative demo-fast timing without changing the default profile.")
    parser.add_argument("--write-validation-samples", action="store_true", help="Write dry/example pane packets and exit.")
    parser.add_argument("--wirepod-timeout", type=float, default=DEFAULT_WIREPOD_TIMEOUT_SECONDS)
    args = parser.parse_args()
    _gopod_log_handle = setup_terminal_log_capture()
    if args.write_validation_samples:
        raise SystemExit(0 if write_validation_samples(args) else 2)
    ensure_live_chat_messages(write_samples=False)
    prefill_wirepod_logs(args)
    load_swarm_registry()
    apply_timing_environment(args)
    log_timing_profile(args)
    if args.preflight:
        raise SystemExit(0 if run_warmup_preflight(args, live_gate=False) else 2)
    if ready_chorus_live_enabled(args) and not run_warmup_preflight(args, live_gate=True):
        log_event(
            "GOPOD_READY_CHORUS_SUMMARY",
            "BLOCKED_PREFLIGHT",
            live=True,
            robot_calls=False,
            operator_notes=[
                "Set System Audio In to USB Mic.",
                "Set System Audio Out to Analog Output - Built-in Audio.",
                "Plug Jetson into power before live demo.",
            ],
        )
        raise SystemExit(2)
    run_ready_chorus(args)
    start_song_watch_thread()
    print(f"GOPOD_DEMO_READY http://0.0.0.0:{PORT}/gopod_demo.html", flush=True)
    print(f"GOPOD_COCKPIT_OPEN http://127.0.0.1:{PORT}/gopod_demo.html", flush=True)
    print(f"GOPOD_DEMO_8011_START http://0.0.0.0:{PORT}/gopod_demo.html", flush=True)
    ThreadingHTTPServer(("0.0.0.0", PORT), Handler).serve_forever()


if __name__ == "__main__":
    main()

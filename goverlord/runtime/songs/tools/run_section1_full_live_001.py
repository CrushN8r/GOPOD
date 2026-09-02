#!/usr/bin/env python3
import colorsys
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


# This runner now lives in the GOPOD repo (goverlord/runtime/songs/tools/),
# but apiConfig.json, animation_vocab.json, the legacy .txt Section Card, and
# every run's own demo_runs/ output are all Wire-Pod-tree property (the first
# two are read by Wire-Pod's own Go runtime too; demo_runs/ is private runtime
# state that must never enter a public repo) - none of that moved. WIREPOD_CHIPPER_ROOT
# is an explicit, env-overridable anchor back to that tree, replacing what used
# to be REPO_ROOT = Path(__file__).resolve().parents[2] - a location-derived
# value that silently pointed at the wrong directory the moment this file
# moved anywhere else, with no error to catch it.
WIREPOD_CHIPPER_ROOT = Path(os.getenv("GOPOD_WIREPOD_CHIPPER_ROOT", "/home/goverlord/wire-pod/chipper")).expanduser()
DEFAULT_CARD_PATH = WIREPOD_CHIPPER_ROOT / "gopod_probes" / "section_packets" / "section_01_brobots_gopod_card_001.txt"
CARD_PATH = Path(os.getenv("GOPOD_SECTION_CARD_PATH", str(DEFAULT_CARD_PATH))).expanduser()
LOG_DIR = WIREPOD_CHIPPER_ROOT / "gopod_probes" / "demo_runs"
API_CONFIG = WIREPOD_CHIPPER_ROOT / "apiConfig.json"
# One fact, one home: the control song owns ARM_GESTURE_SEQUENCE and
# HEAD_NOD_GESTURE_SEQUENCE (split 2026-07-14 from the control song's own
# test sequences, tuned/live-confirmed by the operator - see
# MOVEMENT_TEST_GESTURE_SPLIT_001.md) - the interview reaches into that
# same file by name via importlib, same reuse pattern
# run_robot_control_song_001.py itself already uses in the other direction
# to reach this file. Never copied, never retuned here.
CONTROL_SONG_PATH = Path(__file__).resolve().parent / "run_robot_control_song_001.py"
GOPOD_REPO_ROOT = Path(os.getenv("GOPOD_REPO_ROOT", "/home/goverlord/crushn8r_git/GOPOD")).expanduser()
# Read-sheet mode's one output file goes to gopod_notes/, a sibling of the
# repo checkout, same env-overridable-anchor pattern as every other path
# constant here - never inside GOPOD_REPO_ROOT, per CLAUDE.md's own Report
# and Output File Rule (reports/session output never lands in a git-tracked
# directory).
GOPOD_NOTES_DIR = Path(os.getenv("GOPOD_NOTES_DIR", str(GOPOD_REPO_ROOT.parent / "gopod_notes"))).expanduser()
TEMPLATE1_PATH = GOPOD_REPO_ROOT / "goverlord/runtime/songs/02_brobots_interview_run/zmisc/brobots_wirepod_interview_section_card_template_1_001.md"
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
    os.getenv("GOPOD_ANIMATION_VOCAB_PATH", str(WIREPOD_CHIPPER_ROOT / "animation_vocab.json"))
).expanduser()
# Score architecture: the song folder (story.md + knobs.json) is the real
# architecture and content-complete (verified byte-for-byte against the
# legacy .txt card's visible lines/value points/cycle weights/emotional
# beats) - it is now the default with no env var set. GOPOD_SECTION_SONG_DIR
# still names a different song explicitly, unchanged. The legacy .txt
# Section Card stays in place as an explicit, opt-in fallback only - never
# deleted, never the silent default again.
# Repointed 2026-08-19 (INTERVIEW_VAMP_SPLIT_001.md): the old single
# 01_brobots_interview_section_01/ folder was split into
# 01_brobots_interview_vamp/ (the pre-show, its own standalone song now)
# and 02_brobots_interview_run/ (this - the 7-exchange interview itself).
# This constant now names the RUN half only; the vamp half is a separate
# song folder passed via PRESHOW_SONG_DIR_ENV, never this default.
DEFAULT_SECTION_SONG_DIR = GOPOD_REPO_ROOT / "goverlord/runtime/songs/02_brobots_interview_run"
LEGACY_SECTION_CARD_ENV = "GOPOD_USE_LEGACY_SECTION_CARD"

LIVE_GATE = "GOPOD_ALLOW_LIVE_ROBOT_SPEECH"
# No literal LAN IP fallback here - no network identifiers in tracked files
# (2026-08-15, GOPOLISHER_CUBE_SESSION_SWEEP_001.md). Every real caller of
# this script (interview-vamp/interview-vamp-play/interview-run, all routed
# through brobots.sh) already has GOPOD_WIREPOD_BASE_URL exported via
# ~/.gopod_alias_lib/local_config.sh (private, untracked) before this file
# ever runs - same env var brobots.sh's own aliases now read the same way.
WIREPOD_BASE_URL = os.getenv("GOPOD_WIREPOD_BASE_URL")
ROBOT_EXPRESSION_GATE = os.getenv("ROBOT_EXPRESSION", "")
BROBOT_1_SERIAL = os.getenv("GOPOD_BROBOT_1_SERIAL", "0dd1b9e9")
BROBOT_2_SERIAL = os.getenv("GOPOD_BROBOT_2_SERIAL", "0dd1d8bf")
JDOCS_PATH = WIREPOD_CHIPPER_ROOT / "jdocs" / "jdocs.json"


def _write_brobot_eye_colors_state():
    """Chat-bubble background, read live from each robot's own real eye
    color (2026-09-01 operator request) - not a fixed pair of colors.
    custom_eye_color's hue/saturation (jdocs.json, vic.RobotSettings) is
    the LIVE authority whenever enabled - it overrides the plain eye_color
    enum, same "one source of truth: the robot's own real settings"
    principle load_robot_format_from_jdocs() (gopod_weather_fetch_001.py)
    already uses for weather units. Read once per run start (called from
    Robots.__init__, same site/cadence as the rich-display state write
    just above) - a later mid-run color change on the robot's own app
    never moves a bubble already on screen; the NEXT run picks it up
    fresh, per operator direction ("not mid runs though"). Pure UI
    side-channel - any failure (missing jdocs.json, malformed entry,
    either serial simply not present yet) is swallowed silently, same
    discipline as the rich-display write; never break the actual
    performance over a background color.
    """
    try:
        with JDOCS_PATH.open("r", encoding="utf-8") as handle:
            entries = json.load(handle)
        colors = {}
        for label, serial in (("brobot_1", BROBOT_1_SERIAL), ("brobot_2", BROBOT_2_SERIAL)):
            thing = f"vic:{serial}"
            for entry in entries:
                if entry.get("thing") == thing and entry.get("name") == "vic.RobotSettings":
                    settings = json.loads(entry["jdoc"]["json_doc"])
                    eye = settings.get("custom_eye_color") or {}
                    if not eye.get("enabled"):
                        break
                    r, g, b = colorsys.hsv_to_rgb(eye["hue"], eye["saturation"], 1.0)
                    colors[label] = "#{:02X}{:02X}{:02X}".format(
                        round(r * 255), round(g * 255), round(b * 255)
                    )
                    break
        if colors:
            eye_color_file = WIREPOD_CHIPPER_ROOT / "webroot" / "gopod_brobot_eye_colors.json"
            eye_color_file.write_text(json.dumps(colors))
    except Exception:
        pass


# Interview-only theatrical KG search sequence (searching -> searchingGetout,
# then the line's own speech fires with "answering" instead of the usual
# fallback) - operator's own explicit ask, 2026-07-24: the interview's JSON
# is fully pre-generated before playback ever starts (see generate_phase()/
# main()), so there is no real LLM wait to mask here - this is scripted in
# deliberately to make a canned read *look* like a live search/answer. Gated
# by song_id so it never fires for bingo/awaken/preshow or anything else.
# This is a song_id string, deliberately kept unchanged by the 2026-08-19
# vamp/run folder split (INTERVIEW_VAMP_SPLIT_001.md) - identity, not a
# path, so it still means "the interview content," now housed in
# 02_brobots_interview_run/. Kept as-is to avoid a wider rename ripple
# (SONG_REGISTRY key, ALIAS-LIBRARY.md references) beyond this pass's scope.
INTERVIEW_KG_SEARCH_SONG_ID = "brobots_interview_section_01"
INTERVIEW_KG_SEARCH_SECONDS = 3.0
INTERVIEW_KG_SEARCH_LOOP_INTERVAL = 0.333
# Brobot 3 (pre-show host) is a member of the shared Brobots persona family
# by framing only - she has no physical robot. This is deliberately NOT a
# real device serial shape (unlike BROBOT_1_SERIAL/BROBOT_2_SERIAL above) so
# nobody mistakes it for hardware to connect to. Never pass this to a
# Wire-Pod `serial=` param - see the brobot_3_host comment below for how her
# log lines actually reach Wire-Pod (they borrow Brobot 1's real serial as
# plumbing, not as her identity).
BROBOT_3_ESN = "VOICE_ONLY_NO_HARDWARE"
# Brobot 4 (pre-show host, male voice) is the same shape as Brobot 3 above -
# a real roster member, addressed by name, with no ESN and never one: same
# sentinel value, same "never pass to a Wire-Pod serial= param" rule. Two
# hosts now share this one non-hardware identity pattern, not a special case
# invented for the second one.
BROBOT_4_ESN = "VOICE_ONLY_NO_HARDWARE"
BROBOT_4_DISPLAY_ROLE = "Brobot 4"
RESTART_WIREPOD_BEFORE_RUN = os.getenv("GOPOD_RESTART_WIREPOD_BEFORE_RUN", "0") == "1"
WIREPOD_POST_RESTART_SETTLE_SECONDS = float(os.getenv("GOPOD_WIREPOD_POST_RESTART_SETTLE_SECONDS", "4"))
SECTION1_MAX_EXCHANGES = int(os.getenv("GOPOD_SECTION1_MAX_EXCHANGES", "0") or "0")
ROBOT_SPEECH_CHUNK_LIMIT = int(os.getenv("GOPOD_ROBOT_SPEECH_CHUNK_LIMIT", "420") or "420")
SECTION1_REPLAY_LOG = os.getenv("GOPOD_SECTION1_REPLAY_LOG", "").strip()
VOICE_DESTINATION_ENV = "GOPOD_VOICE_DESTINATION"
VOICE_DESTINATIONS = ("robot", "monitor", "off")
# LLM lane: run-level model choice, mirrors the voice_destination env-var
# shape exactly - unset means "ask or remember," set means "skip the menu."
# Judgment call, 2026-08-19 split (INTERVIEW_VAMP_SPLIT_001.md): this state
# was nested under the old folder's vamp/ subfolder as a "loader
# convenience" only - it is not vamp-specific data (resolve_content_model()
# backs the interview RUN's own LLM generation too, not just the preshow's
# two llm_coloured beats). Now that vamp/run are separate top-level
# folders, this remembered-model cache moved to the RUN folder (the
# heavier, default LLM consumer) rather than staying under vamp's own new
# home. Flagged, not silently decided - a genuinely shared, song-agnostic
# location (e.g. under tools/) is a reasonable alternative if the operator
# wants this decoupled from either song folder later.
CONTENT_MODEL_STATE_PATH = GOPOD_REPO_ROOT / "goverlord/runtime/songs/02_brobots_interview_run/content_model_state.json"
DEFAULT_CONTENT_MODEL = "gemma2:2b"
# Scored two-host pre-show song: unset (the default) changes nothing below -
# main() runs generate_phase() then playback_phase() exactly as before. Set
# to a goverlord/runtime/songs/<song_id>/ folder (brobots_preshow) to
# play that song's four movements in parallel with interview generation.
PRESHOW_SONG_DIR_ENV = "GOPOD_PRESHOW_SONG_DIR"
# Read-sheet mode: a text-only review pass over the pre-show song alone.
# Unset (the default) changes nothing - byte-identical to today's run.  Set
# to "1" (requires PRESHOW_SONG_DIR_ENV to also be set) and every song beat
# still generates exactly as it would live - same prompts, same model, same
# hints, real LLM calls for the two llm_coloured beats - but nothing is ever
# spoken: no Kokoro subprocess call for the hosts' canned lines, no
# robots.say() dispatch for the robots' coloured lines. Each beat's hint and
# the line actually produced are written to one read-sheet file in
# GOPOD_NOTES_DIR at the end of the run. This only ever silences the song
# itself - it does not touch generate_phase/playback_phase or voice_destination,
# so the interview that follows still runs exactly as GOPOD_VOICE_DESTINATION
# and GOPOD_ALLOW_LIVE_ROBOT_SPEECH already say it should (unset = dry, same
# as any other run today).
PRESHOW_READ_SHEET_ENV = "GOPOD_PRESHOW_READ_SHEET"

# pre_chat_status_voice, per Template 1's BROBOTS_INTERVIEW_RUNTIME_SCAFFOLD_001 block:
# engine kokoro_82m, voice af_bella, stop_before PLAYBACK_START/PLAYBACK_TURN_START/Brobot lines.
# All status lines up to that boundary are spoken live via a persistent Kokoro
# synthesis worker (system python3 lacks the kokoro package; the venv that
# generated the original pre_chat_status_voice assets has it).
KOKORO_STATUS_ANNOUNCER_PATH = (
    GOPOD_REPO_ROOT
    / "goverlord/runtime/data_gomad/robot/kokoro_voice/local_tts_voice_registry_001/interview_status_kokoro_announcer_001.py"
)
KOKORO_STATUS_VENV_PYTHON = Path(
    os.getenv("GOPOD_KOKORO_VENV_PYTHON", "/home/goverlord/gopod_tts/engines/kokoro_82m_001/venv/bin/python3")
).expanduser()
KOKORO_STATUS_VOICE = "af_bella"
# Brobot 4's own persistent --speak-stdin worker - a second, fully
# independent subprocess and state slot, never sharing _KOKORO_STATUS_STATE
# or its "proc". Chosen over extending the announcer script's stdin protocol
# to carry a per-line voice selector (the other path PRESHOW_SURVEY_001.md
# named) specifically because that would mean editing
# interview_status_kokoro_announcer_001.py itself - a file Brobot 3's own
# live narration path already depends on. A second independent subprocess
# costs a second ~20s model load (see kokoro_status_warm_up's own docstring)
# but touches zero shared code or state - nothing here can break Brobot 3's
# existing path, by construction, not by care.
KOKORO_HOST4_VOICE = "am_puck"
_KOKORO_HOST4_STATE = {"attempted": False, "proc": None}
KOKORO_STATUS_STOP_PREFIXES = (
    "STATUS: PLAYBACK_START",
    "STATUS: PLAYBACK_TURN_START",
    "Brobot 1:",
    "Brobot 2:",
)
_KOKORO_STATUS_STATE = {"attempted": False, "proc": None, "active": True, "live": False}

# Pre-interview status lines get mirrored into Wire-Pod's own log window
# (logger.LogUI/LogDebugUI, read by /api/get_logs and /api/get_debug_logs)
# via the same assume/say_text/release handshake Robots.say() uses for real
# Brobot turns - text is left empty (Kokoro already narrates these lines
# aloud via the speak-stdin worker above) so only display_text/display_role
# land in the log, under the "Brobot 3" role, as if it were a third speaker.
# STOP-CONDITION NOTE: Wire-Pod's log window has no write path that doesn't
# also require a real robot connection (see server.go's /api-sdk/say_text
# handler - it always resolves `serial` against a connected robot). Brobot 3
# has no robot of her own, so this borrows Brobot 1's REAL serial purely as
# network plumbing to reach the shared log buffer - it is never exposed as
# her identity (BROBOT_3_ESN above is the sentinel for that), and she is
# never given her own `serial=` value that Wire-Pod could try to connect to.
# Stops at the same boundary as Kokoro speech (KOKORO_STATUS_STOP_PREFIXES),
# so this only ever runs before the interview.
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
# "action_parameters" was removed from this list - it now derives from
# animation_vocab.json (see action_parameters() below) instead of being a
# Template 1-owned fact. Template 1 no longer needs to carry this key.


# Voice lane hardening: one registry (scaffold.voice_engines), one gate
# every engine selection goes through. _VOICE_ENGINE_STATE is populated
# once, early, by configure_voice_engines() - the Kokoro narration
# subprocess only ever spawns once (_kokoro_status_ensure_loaded's own
# "attempted" guard), so the registry has to reach that first spawn to mean
# anything, not just decorate an already-hardcoded call.
_VOICE_ENGINE_STATE = {"kokoro": None}


def resolve_voice_engine(engine_key, scaffold, required_for=None):
    # Three distinct, honest failure modes - never a silent fallback to a
    # different engine: unknown key, known but not ready, ready but not
    # actually wired for the use being asked of it (e.g. Kokoro is ready
    # for pre-show narration, not for interview turns - nothing dispatches
    # that today, so asking for it here must fail loudly too).
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
    # Voice lane: exactly three values, enforced - never a silent fallback
    # to "robot" on a mistyped env var. Unset means "robot", today's only
    # real dispatch path; anything else must be one of the three exact
    # strings or this fails loudly, same discipline as resolve_voice_engine.
    raw = os.getenv(VOICE_DESTINATION_ENV, "robot").strip()
    if raw not in VOICE_DESTINATIONS:
        raise RuntimeError(
            f"BLOCKED: {VOICE_DESTINATION_ENV}={raw!r} must be one of {VOICE_DESTINATIONS}"
        )
    return raw


def configure_voice_engines(scaffold):
    # Resolves and gates the pre-show narration engine from the registry,
    # once. Reads scaffold.pre_chat_status_voice.engine (already existed as
    # the documented choice - previously never actually consulted by code,
    # see the KOKORO_STATUS_* constants this replaces as the live source).
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
    # Registry-driven if configure_voice_engines() already ran (the normal
    # generate_phase() path); falls back to the original hardcoded
    # constants untouched otherwise (e.g. a caller that spawns this
    # directly without going through generate_phase() first) - additive,
    # not a forced cutover.
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


# Channels lane hardening: _wirepod_log_status (real content - narration
# meant to be read) and _wirepod_log_marker (a pure control signal - see
# GOPOD_SONG_START_ROLE/GOPOD_SONG_END_ROLE below) used to be two
# independently-coded, nearly-identical functions with no data marking
# which was which. Both now route through this one shared primitive with
# an explicit, validated "kind" - a future caller can't send a
# marker-shaped payload and have it silently treated as displayable
# content, or vice versa. error_label keeps each caller's own historical
# status-line text byte-identical in the (never-hit in practice) exception
# path, so nothing about the produced log output changes.
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


# Channel-1 song boundary markers - the cockpit's own assume/release baton
# for its rich display, same discipline as behavior control. Mirrors
# _wirepod_log_status's own empty-text pattern exactly (no speech, just a
# display_role/display_text write through the same assume/say_text/release
# seam) so a GOPOD_SONG_START/GOPOD_SONG_END line lands in both of
# server.go's log arrays under the BROBOT_RICH_DISPLAY tag - see the
# channel-1 borrow-and-return note in gopod_notes. Gated on the same "live"
# flag _wirepod_log_status uses: a dry run never touches a real robot, so it
# never fires this either.
GOPOD_SONG_START_ROLE = "GOPOD_SONG_START"
GOPOD_SONG_END_ROLE = "GOPOD_SONG_END"


def _wirepod_log_marker(display_role, display_text):
    _wirepod_log_send(display_role, display_text, kind="marker", error_label="WIREPOD_SONG_MARKER_ERROR")


# Brobot 3's colouring path. Deliberately separate from llm_colour_line /
# build_exchange_prompt / normalize_robot_safe - those are built around a
# Section-Card `line` object and the two Brobot roles, and normalize_safe's
# `actionParameter... spoken text` grammar exists only to drive physical
# robot animation packets, which don't apply to a voice-only host. This
# reuses only the underlying mechanism (same knowledge endpoint, same
# chat-completions POST shape) via a lighter, standalone prompt.
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
# Coloured once per beat key per run, then reused - keeps the pre-show snappy
# (LLM_START/LLM_DONE alone can fire 10x in a 5-line section) and means the
# LLM is never asked to colour a line containing a raw file path/run id/etc,
# since only the static hint text (never the dynamic status message) is sent.
_STATUS_BEAT_COLOUR_CACHE = {}

# Fallback for a status() line that matches no known beat above (e.g.
# GENERATION_REPLAYED, the only status the replay branch of generate_phase()
# emits during its generation section) or matches a beat with no configured
# hint. Without this, coloured_status_line() returned the raw `line`
# unchanged in that case, and Kokoro speaks whatever it gets back - so a
# technical "STATUS: ..." string would be read aloud to the live audience.
# Routed through the same LLM-colouring pass instead, under its own cache
# key, so the pre-show always narrates something audience-safe.
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
    # Pre-show song, Movements 2/3: Brobot 1/2's own readiness reaction -
    # scored (the hint lives in the pre-show's own story.md, never
    # hand-typed here) and LLM-coloured, never canned. Deliberately its own
    # function, not a reuse of llm_colour_line (built around a Section-Card
    # `line` object, value targets, and thinking-window machinery that don't
    # apply to a one-off wake-up reaction) or llm_colour_status_beat (built
    # around Brobot 3's own host persona, not Brobot 1/2's). Same
    # chat-completions POST shape as both, same knowledge endpoint - the
    # only genuinely reusable part.
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
    # 120s, matching llm_colour_line's own timeout - a robot's actual spoken
    # readiness line deserves the same slow-model tolerance a real interview
    # turn already gets, not llm_colour_status_beat's 30s (built for a
    # decorative narration gloss, not a scored line the audience hears).
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
        # including a second Ctrl-C landing mid-HTTP-call during an already-in-flight
        # cancellation (KeyboardInterrupt is a BaseException, not caught by `except Exception`
        # alone - this is what turned a clean cancel into a chained-traceback crash).
        print(f"STATUS: BROBOT3_COLOUR_FALLBACK beat={beat_key} {type(exc).__name__}: {exc}", flush=True)
        return line
    _STATUS_BEAT_COLOUR_CACHE[beat_key] = coloured
    return coloured


def _announce_status_kokoro(line):
    if not _KOKORO_STATUS_STATE["active"]:
        return
    if any(line.startswith(prefix) for prefix in KOKORO_STATUS_STOP_PREFIXES):
        _KOKORO_STATUS_STATE["active"] = False
        # Voice lane: in "monitor" destination, the persistent worker is
        # needed again immediately after this boundary to speak the turns
        # themselves (see speak_turn_via_kokoro) - only close its stdin here
        # for every other destination, unchanged from before this existed.
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


# ============================================================================
# Timing instrumentation (net-video timing map pass, 2026-07-14). Measurement
# only - wraps existing call sites with wall-clock start/end capture, never
# changes what is called, when, or with what arguments, and never touches a
# score/song/knob/timing value. Every entry is tagged capture_quality:
#   "true_completion" - the elapsed number reflects when the real-world
#     action genuinely finished (a blocking Kokoro speak-stdin DONE ack, a
#     confirmed HTTP-ready poll, a real LLM chat-completion response, or a
#     deterministic time.sleep()).
#   "accept_only" - the elapsed number only proves a command was ACCEPTED.
#     Wire-Pod's say_text and movement HTTP calls return the instant the
#     command is issued, never when speech or physical motion actually
#     finishes - there is no completion signal anywhere in this stack for
#     those two calls (see run_robot_control_song_001.py's own golden truth
#     5, and CLAUDE.md's "Markers do not prove completion" gotcha). Never
#     silently reported as if it were the other.
# Survey finding this pass is built on: NET_VIDEO_TIMING_MAP_SURVEY_001.md
# (gopod_notes) - zero timestamp/duration fields existed anywhere on disk
# before this.
# ============================================================================
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
    # Voice lane, "monitor" destination: reuses the exact same persistent
    # --speak-stdin worker Brobot 3's pre-show narration already spawns (see
    # _kokoro_status_ensure_loaded) - never a second Kokoro process. Turn
    # text is already fully composed and must be spoken verbatim, so this
    # skips coloured_status_line()'s LLM-colouring/caching entirely - that
    # pass exists to turn a raw status beat into a spoken gloss, which does
    # not apply to a line the show already wrote.
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
    # Mirrors _kokoro_status_ensure_loaded's shape exactly, but reads and
    # writes only _KOKORO_HOST4_STATE - never _KOKORO_STATUS_STATE. Brobot 3's
    # own ensure-loaded function, and every line of _announce_status_kokoro/
    # speak_turn_via_kokoro, is untouched by this function's existence.
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
    # Brobot 4's own speak call - same write/flush/read-ack protocol as
    # speak_turn_via_kokoro, against her own separate subprocess. Text is
    # already fully composed (host dialogue is authored, not LLM-coloured),
    # spoken verbatim, no caching, no coloured_status_line involvement.
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
    # scaffold is accepted-but-unused: kept so existing call sites
    # (robot_safe/normalize_robot_safe/split_robot_safe_line) don't need to
    # change their own signatures. The action-parameter vocabulary now
    # derives from animation_vocab.json (Phase 1's single source of truth)
    # instead of from Template 1's scaffold, so Template 1 is no longer
    # required to carry an "action_parameters" key.
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
    # Animation lane hardening: the Go/REST side has always read
    # animation_vocab.json's "aliases" map (~70 words, e.g. "excited" ->
    # "veryHappy") to resolve loose LLM wording before validating it. This
    # file only ever read "tokens" - the same source file, one consumer
    # using the aliases, one silently ignoring them. Fixed here: same
    # lazy-cached-from-disk pattern as action_parameters() above, no new
    # file, no duplicated vocabulary.
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
    # The one real fix this pass makes to the interview's own repair path:
    # a candidate action is valid as-is, OR resolves through the alias map
    # to a valid canonical token, OR is genuinely unresolved (caller decides
    # the fallback - normalize_robot_safe/split_robot_safe_line both fall
    # back to default_action_parameter(), unchanged). Never invents a token
    # outside action_parameters()'s own verified set.
    valid = action_parameters(scaffold)
    candidate = str(action or "").strip()
    if candidate in valid:
        return candidate
    canonical = animation_aliases(scaffold).get(candidate.lower())
    if canonical in valid:
        return canonical
    return None


def resolve_animation_mode(mode_key, scaffold):
    # Animation lane hardening: mode (blocking vs fire-and-forget) used to
    # be implicit - wirepod_command_packet() always emitted playAnimationWI,
    # no way to ask for the other one. Both are genuinely ready (Wire-Pod
    # implements both fully); fire_and_forget just stays the default so
    # animation never interrupts a spoken interview line.
    registry = (scaffold or {}).get("animation_lane", {}).get("modes", {})
    entry = registry.get(mode_key)
    if entry is None:
        raise RuntimeError(f"BLOCKED: animation mode {mode_key!r} is not in the animation_lane.modes registry")
    if entry.get("readiness") != "ready":
        reason = entry.get("readiness_reason", "not marked ready")
        raise RuntimeError(f"BLOCKED: animation mode {mode_key!r} is not ready for use: {reason}")
    return entry


def resolve_animation_capability(capability_key, scaffold, required_for=None):
    # getImage / conversation_mode: real, wired REST/KG-path capabilities
    # with zero equivalent in the interview's own flow. On the board,
    # marked not_ready, never silently ignored - selecting either must
    # fail loudly, same discipline as the voice lane's Piper/Chatterbox.
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
    # Channels lane hardening: same gate pattern as voice/animation - one
    # registry (scaffold.channels), one function every channel selection
    # goes through. play_sound (robot_speaker_raw_audio_file) is the one
    # not-ready entry today; nothing calls this with that key, so nothing
    # is wired to work, but the gate is real and proven to block on request.
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
    # Channels lane hardening: makes payload variant + destination explicit
    # and reachable in the produced data, not inferred from which parameter
    # happened to be non-empty. Every real turn sends spoken content to the
    # robot speaker; channel_1_rich_display is only listed when there is
    # something to mirror (display_text non-empty) - exactly today's real
    # behavior, now named rather than implied. channel_2_firehose is never
    # listed separately: it cannot be opted out of at this call site (see
    # the channels registry's always_on entry), so there is no choice here
    # to describe.
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
    # Movements lane hardening: same gate pattern as voice/animation/
    # channels - one registry (scaffold.movements), one function every
    # axis selection goes through. move_wheels is the one not-ready entry;
    # nothing dispatches it, so nothing is wired to work, but the gate is
    # real and proven to block on request.
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
    # Movements lane hardening: "always end at speed 0 before releasing
    # control" was a convention every real caller just happened to follow -
    # observed, never checked. A movement sequence that doesn't return to
    # zero should not be silently possible. Only validates sequences that
    # actually pass through this file's own dispatch (speaker_visual_cue);
    # the alias trio's shell functions are separate code, not reachable
    # from here.
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
    # Scored interview movement (per-line brobot_1_movement/brobot_2_movement
    # in the interview song's knobs.json) - dispatches through
    # gentle_arm_gesture_cue()/head_nod_gesture_cue(), the interview's own
    # GESTURE sequences (ARM_GESTURE_SEQUENCE/HEAD_NOD_GESTURE_SEQUENCE),
    # split 2026-07-14 from the control song's own TEST sequences so tuning
    # one can never silently change the other (MOVEMENT_TEST_GESTURE_SPLIT_001.md).
    # No sequence values live here. manage_control defaults True, threaded
    # straight through to those two
    # functions' own manage_control parameter: each movement fires as its
    # own isolated assume->move->release, same as the control song's own
    # isolated single-note testers, because nothing in the real interview
    # holds behavior control continuously across a turn (each backpack_turn's
    # own say() does its own assume/release too) - playback_phase() never
    # passes manage_control, so its behavior is unchanged by this parameter's
    # existence. manage_control=False is for a caller that already holds
    # control continuously itself (the movement rehearsal tool) and needs
    # this call to neither assume nor release around the single movement.
    # Empty/absent movement fires nothing - silence is the default, enforced
    # here, not by the caller checking first.
    if not movement:
        return None
    control_song_mod = _load_control_song_module()
    # gentle_arm_gesture_cue()/head_nod_gesture_cue() only ever call mod.
    # validate_movement_sequence_ends_at_zero(...) and mod.wirepod_web_send_form(...) on
    # whatever "mod" they're given - a real module reference isn't needed,
    # just those two attributes. sys.modules[__name__] is unreliable here:
    # this file is routinely loaded via importlib.util.module_from_spec() by
    # other scripts (the control song, the pre-show generator), and that
    # loading pattern never registers the resulting module object into
    # sys.modules under its spec name - only real `import` statements do
    # that. A tiny namespace exposing just the two needed callables sidesteps
    # that entirely, regardless of how this file itself was loaded.
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
    # Phase 1: reachability only. Mirrors action_parameters(scaffold)'s
    # accessor shape. Not called anywhere in the generation path yet -
    # wiring a cycle into build_exchange_prompt/llm_colour_line/
    # generated_turn is Phase 2.
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


# Two-dial framework, rung 1: locked ordered value points per exchange.
# Same stop conditions as read_field (blank line after content, "---", a
# "LINE " header, or the next field's bare label) - but returns an ordered
# list, one item per line, instead of joining into a single blob. A leading
# "1. "/"1) " numbering marker (used in the Section Card purely for human
# readability) is stripped; order in the file is preserved as the list order.
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


# Phase 1 (weighted): a bare label with no ":temp" (e.g. a hand-edited CYCLES
# line that dropped the weight) is tolerated rather than rejected - it's
# treated as "selected, weight unspecified" and given this default. Chosen as
# a mid-range value (the shaped vectors below run 0-5) so a bare label reads
# as "on, not emphasized" rather than snapping to either extreme.
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
        # Two-dial framework, rung 1: locked, ordered, countable value points.
        # "VALUE POINTS" is the new explicit ordered list (falls back to the
        # existing flat blob as a single-item list if a line hasn't been
        # given one yet, so an older, not-yet-updated card entry still works).
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


# Song split, additive: story.md + knobs.json pair for a song folder (see
# goverlord/runtime/songs/<song_id>/), wired in behind GOPOD_SECTION_SONG_DIR
# at generate_phase()'s one call site. load_section_card() above is untouched
# and stays the default when that env var is unset - no forced cutover.
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
    # Mirrors load_section_card's own "find LINE N headers, slice into
    # blocks" shape, adapted to story.md's markdown headings/blockquote/
    # numbered-list conventions instead of the Section Card's flat
    # LABEL:\nvalue text convention.
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
    # Returns the exact same section dict shape as load_section_card() above
    # (same keys, same per-line keys) so every downstream function
    # (is_crystallizer, build_exchange_prompt, selected_thinking_cycles,
    # interviewer_value_point_count, generated_turn, ...) needs zero changes.
    # knobs.json carries the per-step knobs; story.md carries the matching
    # content, joined by step_id - a real merge, not a format-detection
    # branch like load_interview_scaffold()/load_pronunciation_registry().
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
        # Playback turn order, whole-song, data-driven - absent means
        # "brobot_2" (today's interview order, unchanged). playback_phase()
        # reads this once per run from log["section_card"]; never touched
        # per-line, never inferred from song name. See brobots_bait_001's
        # own knobs.json for the one song that overrides it.
        "first_speaker": knobs.get("first_speaker", "brobot_2"),
        # Breathing pause after each fired movement, whole-song,
        # data-driven - absent/0 means no pause (today's interview/pre-show
        # timing, unchanged). playback_phase() reads this once, applies it
        # after every movement that actually fired (arm, then head, then
        # into speech) - never a separate mechanism per movement slot.
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
                # Job 1/2 (LLM lane hardening): explicit, mixer-reachable
                # overrides. Absent (None) on any of these means "fall
                # through to the historical implicit derivation" - see
                # resolve_turn_mode()/resolve_canned_text()/
                # build_exchange_prompt()'s opening-task check. The old
                # load_section_card() path below never sets these, so it
                # keeps its old behavior untouched.
                "brobot_2_mode": step.get("brobot_2_mode"),
                "brobot_1_mode": step.get("brobot_1_mode"),
                "brobot_2_canned_source": step.get("brobot_2_canned_source"),
                "brobot_1_canned_source": step.get("brobot_1_canned_source"),
                "brobot_2_opening_task": step.get("brobot_2_opening_task"),
                # Scored interview movement (Job: score movement into the
                # interview) - absent/empty means silence, same
                # absent-means-fall-through convention as the mode/
                # canned_source/opening_task knobs just above. Read by
                # generate_exchange_record() -> playback_phase(), dispatched
                # via fire_scored_interview_movement(). Never read by any
                # LLM-prompt-building code path - purely runner-enforced.
                "brobot_2_movement": step.get("brobot_2_movement"),
                "brobot_1_movement": step.get("brobot_1_movement"),
                # Second scored movement per role, same absent-means-silence
                # convention - fires after the first (arm, then head, per
                # the bait song's own default), never inferred, never a
                # second mechanism. Absent for every existing song
                # (interview, pre-show) - only brobots_bait_001 sets this.
                "brobot_2_movement_2": step.get("brobot_2_movement_2"),
                "brobot_1_movement_2": step.get("brobot_1_movement_2"),
            }
        )
    return section


# Pre-show song loader - parallel to parse_story_md/load_section_pair above,
# not a modification of either. The interview's own loader stays exactly as
# it is; this is a second, independent reader for a song with a genuinely
# different shape: movements and beats, not exchanges - no interviewer/
# interviewee pair, no locked value points to reveal. A beat's content is
# either a literal, authored line ("> TEXT:", spoken verbatim - the hosts'
# own dialogue) or a situational hint ("> HINT:", LLM-coloured - Brobot 1/2's
# readiness reactions) - two distinct markers instead of parse_story_md's
# single blockquote-is-always-visible_line rule, because the pre-show
# genuinely has both kinds of content and the interview never needed to.
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
    # Same two-file, step_id-joined container the interview's own
    # load_section_pair uses (knobs.json = dials, story.md = content) - that
    # part of the convention generalizes fine. The per-beat shape does not
    # reuse the interview's line dict at all: no visible_line/value_points/
    # exchange_type, because none of those concepts apply here. "speaker"
    # (survey item 3's answer) and "movement"/"mode" are the only dials a
    # pre-show beat needs.
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
    # Phase 2: resolves a LINE's cycle_weights (Section Card) against
    # Template 1's thought_cycles pool, keeping only weight>0 entries,
    # highest weight first so a small model's limited attention favors the
    # most-emphasized focuses. Silently drops any label not in the pool
    # rather than guessing at unknown cycle names.
    weights = line.get("cycle_weights") or {}
    pool = {c.get("label"): c for c in thought_cycles(scaffold).get("cycles", [])}
    selected = [
        (label, weight, pool[label].get("focus", ""))
        for label, weight in weights.items()
        if weight > 0 and label in pool
    ]
    selected.sort(key=lambda item: item[1], reverse=True)
    return selected


# Two-dial framework, rung 1 (generate-side only, interviewer's question
# wording): the runner owns the COUNT - how many of a line's locked, ordered
# value points the interviewer's question reaches for. Not an LLM choice;
# read straight from the Section Card's per-line "INTERVIEWER VALUE POINT
# COUNT" field, defaulting to a small, conservative reach (the interviewer
# teases a couple of points, it doesn't recite the whole answer) when a line
# hasn't set one explicitly.
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
    # The runner picks the first N in the locked order - never a reordering,
    # never an LLM pick. The LLM only ever sees this pre-selected, pre-counted
    # set; it turns the ladder/fan dial and chooses wording on top of it.
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
        # Job 2: "is this the opening line" used to be a bare line_number==1
        # check - a magic number no song data could see or override.
        # brobot_2_opening_task, when explicitly set (True or False), now
        # decides it instead. None (the field absent, e.g. the old
        # load_section_card() path) preserves the historical line_number==1
        # behavior exactly.
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
    # Echo-defect Layer 1: Brobot 2's job IS to colour and deliver this exact
    # text, so "VISIBLE LINE:" stays a neutral label for that role. Brobot 1's
    # job is the opposite - answer it, never produce a version of it - so the
    # same raw text gets a label that says so inline, right where the
    # quotable text actually sits, instead of only via a generic RULE line
    # several lines further down (read-first found the echoed words already
    # present in the model's own reply, before normalize/split ever touch
    # it - a labeling problem at the point this text is presented, not a
    # missing rule; shared_rules/role_tasks/exchange_types already said
    # "don't echo" in three different places and it still happened).
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

    # Two-dial framework, rung 1: interviewer-only. Gives Brobot 2 the same
    # locked, ordered value points Brobot 1 is asked to reveal, pre-counted
    # by the runner, and a single ladder/fan dial for how the question
    # approaches them. The LLM may not invent, add, drop, or reorder points,
    # and may not choose the count - only the dial position and the wording.
    # Job 2: the instructional text itself used to be hardcoded here - now
    # read from scaffold.prompt_templates (song-family-wide, not per-song),
    # falling back to this exact same wording if a scaffold doesn't define
    # it, so behavior is byte-identical either way. (prompt_templates itself
    # is now fetched once, above, alongside the visible-line labeling.)
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

    # Phase 2: single-pass thinking window. Only added when this LINE
    # requests it and has at least one weighted cycle selected; the model
    # is asked to emit its internal reasoning and its spoken turn in one
    # response, split apart in generated_turn before normalize_robot_safe
    # ever sees the text.
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
    # LLM lane: live from Ollama's own HTTP API, never a hardcoded list, and
    # never shelling out to `ollama list` - the endpoint host/scheme is read
    # from apiConfig.json's existing "endpoint" field (a read, same as
    # load_api_config() already does; apiConfig.json itself is never
    # written by this selector). No ranking/filtering/annotation - name and
    # size only, in whatever order Ollama itself returns.
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
    # LLM lane: mirrors resolve_voice_destination()'s discipline - loud
    # failure over a silent fallback whenever the operator's stated intent
    # (the env var) can't be honored. Interactive menu only when no env
    # override is set, and only ever reached from the non-replay generation
    # path (a replay never needs a model - see generate_phase()).
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
        # STOP CONDITION guard: never block a non-interactive path on input().
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
        # LLM lane: content_model (the operator's run-start pick) wins over
        # apiConfig.json's own model when given - apiConfig.json is Wire-Pod's
        # own live config (its knowledge-graph feature reads/writes it too),
        # never touched by this selector, so its "model" stays the fallback
        # exactly as before whenever content_model isn't supplied.
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
    # Job 2: the "thinking" fallback used whenever no valid action can be
    # resolved was a bare string literal repeated at four call sites. Now
    # one read of scaffold.generation_defaults.default_action_parameter,
    # falling back to "thinking" itself if a scaffold doesn't define it -
    # same behavior, source moved from code to data.
    scaffold = scaffold or {}
    return scaffold.get("generation_defaults", {}).get("default_action_parameter", "thinking")


def robot_safe(text, action="thinking", scaffold=None):
    fallback_action = default_action_parameter(scaffold)
    action = action if action in action_parameters(scaffold) else fallback_action
    text = re.sub(r"\{\{[^}]*\}\}", "", " ".join(str(text).split())).strip()
    if not text:
        raise ValueError("empty robot speech is not allowed in live Section 1")
    return f"{action}... {text}"


# Model-invented speaker labels ("GOPOD Brobot 2:", "Brobot 1:", "Interviewer:")
# - a general per-line pattern, not a fixed lookup list, because the model can
# invent wording no curated list would anticipate. Requires a real colon
# within the first 1-4 words of a line (MULTILINE `^`), which is what makes
# this safe against ordinary prose: a sentence like "I stretch my mechanical
# limbs..." never finds a colon in that window, so the match simply fails and
# the words pass through untouched - only an actual self-announcement (short
# capitalized phrase immediately followed by ":") is ever removed. Known,
# accepted tradeoff: a genuine colon-led clause at a line's start ("Rule:
# don't drop the mic") would also match - judged acceptable for spoken
# persona reactions, where that construction is rare and the alternative
# (robots announcing their own names) is the concrete, observed failure.
_LEADING_SPEAKER_LABEL_RE = re.compile(r"(?m)^\s*[A-Z][A-Za-z]*(?:\s+[A-Za-z0-9]+){0,3}:\s*")


def strip_rich_text_for_voice(raw, scaffold=None):
    # Hardening pass: emoji, markdown formatting characters, and
    # model-invented speaker labels never belong in the voice lane - not
    # optional, no knob. Runs inside normalize_robot_safe itself (every call
    # site, unconditionally) rather than relying on flatten_for_robot_speech
    # downstream, because not every consumer of normalize_robot_safe's output
    # calls flatten_for_robot_speech - read-sheet mode's own SPOKE capture is
    # exactly such a consumer, which is how the emoji/markdown in
    # PRESHOW_READSHEET_20260713_161541.md's m2_c/m3_c reached disk unstripped.
    # Reuses the exact same scaffold-driven speech_cleanup_rules config
    # (emoji_codepoint_ranges/markdown_chars) flatten_for_robot_speech already
    # reads via cleanup_rules()/codepoint_class() - one place to tune either,
    # this only runs it earlier. Substitutes matched characters with a space,
    # matching flatten_for_robot_speech's own convention - the later
    # whitespace-collapse in this function (and in flatten_for_robot_speech,
    # run again downstream, harmlessly, on already-clean text) mops up
    # doubled spaces either way.
    rules = cleanup_rules(scaffold)
    text = str(raw)
    emoji_class = codepoint_class(rules.get("emoji_codepoint_ranges", []))
    if emoji_class:
        text = re.sub(f"[{emoji_class}]", " ", text)
    markdown_chars = str(rules.get("markdown_chars", ""))
    if markdown_chars:
        text = re.sub(f"[{re.escape(markdown_chars)}]+", " ", text)
    # Label stripping runs last, after emoji/markdown removal, so junk in
    # front of a label ("🤖 **GOPOD Brobot 2:**") no longer hides it from
    # "start of line."
    text = _LEADING_SPEAKER_LABEL_RE.sub("", text)
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
    cleaned = re.sub(r"^[A-Za-z ]+:\s*", "", cleaned).strip()
    cleaned = re.sub(r"^actionParameter\s+", "", cleaned, flags=re.IGNORECASE).strip()
    if "..." in cleaned:
        action, spoken = cleaned.split("...", 1)
        # Animation lane hardening: was a bare `action in valid_actions`
        # check - an alias word (e.g. "excited") failed it and fell all the
        # way to the generic default, discarding the model's actual intent
        # even though the vocab already defines a resolution for it.
        # resolve_animation_token() tries the alias map before giving up.
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
            clean_entries.append({"from": source, "to": target})
    clean_entries.sort(key=lambda item: len(item["from"]), reverse=True)
    return clean_entries


def apply_pronunciation_safety(text):
    speech = str(text or "")
    for entry in pronunciation_entries():
        source = entry["from"]
        target = entry["to"]
        pattern = re.compile(rf"(?<!\w){re.escape(source)}(?!\w)", re.IGNORECASE)
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


# Universal character cleanup, mirrored from repo truth: Wire-Pod's own Go
# TTR sanitizer, removeSpecialCharacters() in
# ~/wire-pod/chipper/pkg/wirepod/ttr/kgsim.go (confirmed real, read
# directly - not the separate special_characters_vars.go path this
# scaffold's own SPEECH_PRONUNCIATION_REGISTRY_001 preamble cites, which
# does not exist anywhere in this repo as of this pass; that citation looks
# stale and is flagged, not silently trusted). That Go function is
# confirmed NEVER reachable from /api-sdk/say_text (read directly in
# pkg/wirepod/sdkapp/server.go - the say_text handler dispatches straight
# to robot.Conn.SayText with no cleanup call of any kind); it only fires on
# the separate wake-word/Knowledge-Graph chat path (DoSayText/KGSim). The
# interview's actual live say path relies entirely on this Python-side
# mirror instead - kept in exact step with the Go function's own mapping,
# order, and scope. Do not add characters or behavior beyond what
# kgsim.go's removeSpecialCharacters() itself does - this list is a mirror,
# not an "improvement."
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


# Exact mirror of kgsim.go removeSpecialCharacters()'s own substitution
# table, same characters, same targets, same order (bullet -> "*" is listed
# after the &^*#@ strip below runs, exactly matching the Go source's own
# order, so a bullet-turned-asterisk survives to the final text there too).
# Scaffold-overridable (speech_cleanup_rules.universal_character_replacements)
# with this exact list as the code fallback, matching every other
# prompt/cleanup table in this file's own established convention.
GO_TRUTH_UNIVERSAL_CHARACTER_REPLACEMENTS = [
    ["‘", "'"], ["’", "'"],          # ' ' -> '
    ["“", "\""], ["”", "\""],        # " " -> "
    ["—", "-"], ["–", "-"],          # em/en dash -> -
    ["…", "."],                            # … -> .
    [" ", " "],                            # nbsp -> space
    ["•", "*"],                            # bullet -> * (matches Go: the earlier &^*#@ strip below already ran)
    ["¼", "1/4"], ["½", "1/2"], ["¾", "3/4"],
    ["×", "x"], ["÷", "/"],
    ["ç", "c"],
    ["©", "(c)"], ["®", "(r)"], ["™", "(tm)"],
    [" AI ", " A. I. "],
]

# kgsim.go strips these five characters entirely (regexp `[&^*#@]`) BEFORE
# its substitution table runs - `*`/`#` already lived in this file's own
# markdown_chars set for an unrelated reason (markdown formatting cleanup);
# `&`/`^`/`@` did not, so the Go truth's own strip set was incomplete here
# until now. Applied at the same point markdown_chars already is.
GO_TRUTH_STRIP_CHARS = "&^@"


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
    flattened = remove_configured_label_prefixes(flattened, rules.get("prompt_label_prefixes", []))
    flattened = remove_configured_label_prefixes(flattened, rules.get("spoken_label_prefixes", []))
    flattened = remove_configured_bracket_labels(flattened, rules.get("drop_bracketed_labels", []))
    emoji_class = codepoint_class(rules.get("emoji_codepoint_ranges", []))
    if emoji_class:
        flattened = re.sub(f"[{emoji_class}]", " ", flattened)
    # Go-truth strip set (&^@) folded into the existing markdown_chars pass
    # - same point in the pipeline markdown_chars already ran, *  and # are
    # already covered by markdown_chars itself for an unrelated reason.
    markdown_chars = str(rules.get("markdown_chars", "")) + GO_TRUTH_STRIP_CHARS
    if markdown_chars:
        flattened = re.sub(f"[{re.escape(markdown_chars)}]+", " ", flattened)
    # Universal character cleanup - curly quotes -> straight, em/en dash ->
    # hyphen, unicode ellipsis -> period, nbsp -> space, bullet -> asterisk,
    # fractions, ×/÷, ç, ©/®/™, "AI" spacing - exact mirror of Wire-Pod's
    # own Go truth (kgsim.go removeSpecialCharacters()), confirmed not
    # otherwise reachable from this say path. Previously this step instead
    # DELETED quote_chars/the unicode ellipsis outright, which both
    # over-reached (straight ASCII quotes were being stripped too, never
    # part of the Go mapping) and under-reached (curly single quotes/
    # apostrophes were not covered at all, and none of dashes/nbsp/
    # bullet/fractions/×÷/©®™ were touched).
    flattened = apply_universal_character_cleanup(flattened, scaffold)
    # Literal "..." (three ASCII periods, distinct from the single unicode
    # "…" character the Go-truth mirror above already converts to ".") -
    # an additive, interview-specific normalization, not part of the
    # universal Go-truth set either way, kept as-is.
    for token in rules.get("ellipsis_tokens", []):
        flattened = flattened.replace(str(token), " ")
    flattened = clean_inline_emphasis_punctuation(flattened, scaffold)
    parts = []
    for part in re.split(r"(?<=[.!?])\s+", flattened):
        clean = " ".join(part.split())
        if not clean:
            continue
        key = re.sub(r"[^a-z0-9]+", "", clean.lower())
        if parts and key == parts[-1][0]:
            continue
        parts.append((key, clean))
    deduped = []
    index = 0
    while index < len(parts):
        repeated = False
        max_span = (len(parts) - index) // 2
        for span in range(max_span, 0, -1):
            left = [key for key, _ in parts[index : index + span]]
            right = [key for key, _ in parts[index + span : index + (span * 2)]]
            if left == right:
                deduped.extend(parts[index : index + span])
                index += span * 2
                repeated = True
                break
        if not repeated:
            deduped.append(parts[index])
            index += 1
    parts = deduped
    flattened = " ".join(part for _, part in parts)
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
    # CHANGED: match webroot/sdkapp/js/control.js sendForm().
    # The Wire-Pod UI appends serial/query params to the URL and sends POST
    # with application/x-www-form-urlencoded. The rest of the runner remains unchanged.
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
    SKILL.md: CHECK -> CLEAR -> MIRROR -> START (only if unhealthy) -> CONFIRM.
    Restarting an already-healthy service is what trips systemd's
    start-rate limiter, not the restart command itself - stop-then-start
    does not dodge it either. CHECK first and skipping when already healthy
    is the actual fix; CLEAR (reset-failed) is harmless when nothing is
    wrong and required when a prior restart already tripped the limiter.
    MIRROR (apply_nongo_files.sh) re-asserts the overlay's non-Go files onto
    the live tree before the service comes back up, added 2026-08-10 - see
    gopod_notes/WIRED_POD_GIFT_BACK_ITEM1_ITEM2_SURVEY_001.md."""
    if not live or not RESTART_WIREPOD_BEFORE_RUN:
        return {"enabled": False}

    # STEP 1 - CHECK: already healthy? Skip the restart entirely.
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

    # STEP 2 - CLEAR: unconditional, before any start/restart attempt.
    # check=False - reset-failed on a service with no failed state is a
    # harmless no-op, not an error worth stopping the preflight over.
    print("PREFLIGHT: CLEAR -> sudo systemctl reset-failed wire-pod.service")
    subprocess.run(["sudo", "systemctl", "reset-failed", "wire-pod.service"], check=False)

    # STEP 2.5 - MIRROR: re-assert the overlay's non-Go files (webroot/index.html,
    # customIntents.json, the prompt files, animation_vocab.json, ...) before the
    # service comes back up. A restart alone never re-applies them - only running
    # apply_nongo_files.sh does - see
    # gopod_notes/WIRED_POD_GIFT_BACK_ITEM1_ITEM2_SURVEY_001.md. Diff-then-copy,
    # non-destructive; a failure here is logged and never blocks the restart - the
    # robot coming back online matters more than a doc-mirror hiccup.
    apply_script = GOPOD_REPO_ROOT / "goverlord/wire_pod_overlay/apply_nongo_files.sh"
    print(f"PREFLIGHT: MIRROR -> {apply_script}")
    try:
        mirror_result = subprocess.run(
            ["bash", str(apply_script)],
            check=False,
            capture_output=True,
            text=True,
        )
        for line in mirror_result.stdout.splitlines():
            print(f"PREFLIGHT: MIRROR {line}")
        if mirror_result.returncode != 0:
            print(f"PREFLIGHT: MIRROR_WARN non-zero exit ({mirror_result.returncode}), continuing restart anyway")
    except Exception as exc:
        print(f"PREFLIGHT: MIRROR_WARN failed to run apply_nongo_files.sh ({exc}), continuing restart anyway")

    # STEP 2.6 - RESET_UI: 2026-09-01 operator live-report ("upon a wire-pod
    # restart... chat bubbles window is NOT seen. I still see it after alias
    # wpr, and refresh web page"). gopod_rich_display_ui_state.json is a
    # plain file in webroot - a service restart never touches it, so
    # whatever a run last wrote (expanded:true, possibly still within
    # index.html's own 300s STALE_AFTER_SECONDS window) survives the
    # restart and reopens on the next page load. Only reached on a real
    # restart (CHECK already returned above if wire-pod was healthy) - a
    # restart is a deliberate "start fresh" signal, so force dormant here
    # rather than leaving it to the browser's own staleness heuristic.
    # Pure UI side-channel, same discipline as every other write to this
    # file - never block the actual restart over it.
    try:
        rich_ui_state_file = WIREPOD_CHIPPER_ROOT / "webroot" / "gopod_rich_display_ui_state.json"
        rich_ui_state_file.write_text(json.dumps({"expanded": False, "run_active": False}))
    except OSError:
        pass

    # STEP 3 - START: only reached because CHECK found it unhealthy/not running.
    print("PREFLIGHT: START -> sudo systemctl restart wire-pod")
    try:
        subprocess.run(["sudo", "systemctl", "restart", "wire-pod"], check=True)
    except subprocess.CalledProcessError as exc:
        raise RuntimeError(f"wire-pod restart preflight blocked: start_failed rc={exc.returncode}") from exc

    # STEP 4 - CONFIRM: settle, then poll until ready or timeout.
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
    restart_wirepod_preflight() above - same CHECK -> CLEAR -> MIRROR ->
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

    # STEP 1 - CHECK: already running? Skip the start entirely - nothing
    # to do, and running CLEAR/MIRROR/START against an already-active
    # service is unnecessary churn, not just harmless.
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

    # STEP 2 - CLEAR: same as restart_wirepod_preflight()'s own STEP 2 -
    # unconditional, harmless when nothing's wrong, required when a prior
    # attempt already tripped systemd's start-rate limiter.
    print("PREFLIGHT: CLEAR -> sudo systemctl reset-failed wire-pod.service")
    subprocess.run(["sudo", "systemctl", "reset-failed", "wire-pod.service"], check=False)

    # STEP 2.5 - MIRROR: same as restart_wirepod_preflight()'s own STEP 2.5 -
    # re-assert the overlay's non-Go files before the service comes up.
    apply_script = GOPOD_REPO_ROOT / "goverlord/wire_pod_overlay/apply_nongo_files.sh"
    print(f"PREFLIGHT: MIRROR -> {apply_script}")
    try:
        mirror_result = subprocess.run(
            ["bash", str(apply_script)],
            check=False,
            capture_output=True,
            text=True,
        )
        for line in mirror_result.stdout.splitlines():
            print(f"PREFLIGHT: MIRROR {line}")
        if mirror_result.returncode != 0:
            print(f"PREFLIGHT: MIRROR_WARN non-zero exit ({mirror_result.returncode}), continuing start anyway")
    except Exception as exc:
        print(f"PREFLIGHT: MIRROR_WARN failed to run apply_nongo_files.sh ({exc}), continuing start anyway")

    # STEP 3 - START: only reached because CHECK found it not running.
    print("PREFLIGHT: START -> sudo systemctl start wire-pod")
    try:
        subprocess.run(["sudo", "systemctl", "start", "wire-pod"], check=True)
    except subprocess.CalledProcessError as exc:
        raise RuntimeError(f"wire-pod start preflight blocked: start_failed rc={exc.returncode}") from exc

    # STEP 4 - CONFIRM: same polling shape as restart_wirepod_preflight().
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
        # Console-only knob (see say()'s own print line below) - default
        # False for every existing caller (pre-show, movement rehearsal
        # tool), so nothing about their own console output changes. Never
        # touches robot_safe_line itself, the say-cleaning pipeline, or
        # display_text - purely which string gets printed to the operator's
        # terminal.
        self.bare_capture = bare_capture
        # Console-only knob, same shape as bare_capture above, independent of
        # it. Default flipped 2026-07-25 (operator request, alias-mixer
        # widening) from a hardcoded False to a real cockpit switch: an
        # explicit True/False from the caller (bingo's own construction still
        # passes True) always wins; every other caller now resolves through
        # GOPOD_CONSOLE_RICH_DISPLAY, itself defaulting to ON ("1") unless
        # explicitly set to "0". Operator's own words on why ON is now the
        # right default everywhere, not just bingo: the distinct display_text
        # values already authored (e.g. brobots_awaken's weather step) are a
        # deliberate edit to cut fluff/awkward line-wrap from the console,
        # not bingo-specific content - every song should get the same
        # readable console by default. Never touches robot_safe_line itself,
        # the say-cleaning pipeline, or what the robot actually speaks -
        # purely which string gets printed to the operator's terminal.
        if console_rich_display is None:
            console_rich_display = os.getenv("GOPOD_CONSOLE_RICH_DISPLAY", "1") != "0"
        self.console_rich_display = console_rich_display
        # Chat-bubble auto-expand flag, mirrored from pha0b's own pre-run
        # write (brobots.sh's rich_ui_state_file) - CHAT_BUBBLES_INIT_FLAG_
        # FIX_001.md. pha0b already writes this same file, same value,
        # before invoking any runner; this is that identical write done
        # here too, so a runner started OUTSIDE pha0b (bare
        # run_golden_song_001.py, start-the-interview) gets the same
        # auto-expand-on-fresh-run behavior every pha0b-launched song
        # already has. Idempotent with pha0b's own write - same content,
        # harmless if both fire. Never touches say()/robot_safe_line/the
        # speech-dispatch pipeline - a pure UI side-channel, so a write
        # failure here (missing wire-pod tree, permissions) must never
        # break a run over it.
        try:
            rich_ui_state_file = WIREPOD_CHIPPER_ROOT / "webroot" / "gopod_rich_display_ui_state.json"
            if self.console_rich_display:
                rich_ui_state_file.write_text(json.dumps({"expanded": True, "run_active": True}))
            else:
                rich_ui_state_file.write_text(json.dumps({"expanded": False, "run_active": False}))
        except OSError:
            pass
        _write_brobot_eye_colors_state()
        # Console-only knob, same shape/scoping as the two above - default
        # False for every existing caller. When True, every terminal line
        # this class prints (and, via timing_prefix(), every line the
        # caller's own script prints) gets a "[HH:MM:SS.mmm +X.XXXs]" prefix:
        # a sortable absolute wall-clock read, plus the real measured elapsed
        # time since the last printed event (not a configured sleep value).
        # Built to answer a real question a live run raised - the operator
        # could see buffer_after's own configured values, but not where
        # actual time was really going between them.
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
        # Movements lane hardening: axis is now resolved and gated against
        # the movements registry instead of just being "whatever endpoint
        # string is configured" - and the sequence is validated to end at
        # speed 0 before anything is dispatched, not just assumed to.
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
                    # Movements lane hardening: stated per-step, not just in
                    # the registry - hold_seconds is asserted (the runner
                    # sleeps this long), never a measured confirmation the
                    # physical motion actually finished.
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
        # Animation lane hardening: mode is now an explicit, gated knob
        # instead of being hardcoded inside wirepod_command_packet() -
        # default ("fire_and_forget") matches today's only real behavior.
        animation_mode = self.scaffold.get("animation_mode", "fire_and_forget")
        resolve_animation_mode(animation_mode, self.scaffold)
        action, raw_spoken = split_robot_safe_line(robot_safe_line, self.scaffold)
        spoken = flatten_for_robot_speech(raw_spoken, self.scaffold)
        speech_chunks = chunk_robot_speech(spoken)
        wirepod_chunks = [wirepod_command_packet(action, chunk, animation_mode) for chunk in speech_chunks]
        # bare_capture prunes the "action... " prefix (e.g. "thinking... ")
        # from the console line only - raw_spoken (already split above) is
        # exactly robot_safe_line minus that prefix. spoken/speech_chunks/
        # wirepod_chunks/display_text below are all computed from raw_spoken
        # already, untouched either way - this only changes what gets
        # printed to the terminal. console_rich_display takes priority when
        # set (only true for callers that opt in) and shows display_text
        # instead - the robot's own speech (spoken/speech_chunks/wirepod_chunks
        # below) is computed from raw_spoken exactly as before, never from
        # display_text - this is a terminal-print routing change only.
        if self.console_rich_display and display_text:
            console_line = display_text
        elif self.bare_capture:
            console_line = raw_spoken
        else:
            console_line = robot_safe_line
        print(f"{self.timing_prefix()}{name}: {console_line}")
        # Channels lane hardening: which destinations/payload variants this
        # call actually uses, named explicitly rather than left implicit in
        # which parameters happen to be non-empty - reachable in the
        # produced run json either way (dry or live).
        channels_used = describe_channels_used(display_text)
        serial = BROBOT_1_SERIAL if name == "Brobot 1" else BROBOT_2_SERIAL
        # Voice lane: one destination switch, one shared field shape across
        # all three destinations (robot/monitor/off) so a monitor replay and
        # a robot run are directly comparable in the produced run json - the
        # Vector-only fields (assume/speaker_visual_cue/say/release) are
        # always present, just None where no Vector contact was made.
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

        if self.voice_destination == "monitor":
            # Voice lane hardening: gated here exactly like the robot engine
            # below - a misconfigured engine should surface during a
            # compose-only run, not wait for a live show.
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

        # voice_destination == "robot" - today's only real dispatch path,
        # unchanged: every interview turn's engine choice is gated here,
        # live or dry - a misconfigured engine should surface during a
        # compose-only run, not wait for a live show. Default ("robot_tts")
        # matches today's only real dispatch path exactly; this raises
        # loudly rather than silently falling back if a future song points
        # turn_voice_engine at something not ready or not wired for
        # interview_turns.
        turn_engine = self.scaffold.get("turn_voice_engine", "robot_tts")
        resolve_voice_engine(turn_engine, self.scaffold, required_for="interview_turns")
        if not self.live:
            return {**prelude, "spoken": False, "mode": "dry"}
        result = {**prelude, "spoken": False, "mode": "wirepod"}
        try:
            # Additive opt-out only - default (False) is byte-for-byte the
            # original unconditional assume/release-per-call behavior every
            # interview/pre-show call already depends on. No existing caller
            # passes True; this branch is unreachable from any of them.
            # When True (a caller holding control continuously across
            # several say() calls, matching webroot/sdkapp/js/control.js's
            # own pattern - assume once via the UI's toggle, fire many
            # commands, release once at the end - not per command), this
            # call assumes nothing itself and the finally block below does
            # not release either.
            if suppress_assume_release:
                result["assume"] = {"suppressed": True}
            else:
                # CHANGED: use the same assume/direct-speak/release POST pattern as the Wire-Pod web UI.
                assume = wirepod_web_send_form("/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10)
                result["assume"] = assume
                time.sleep(0.25)
            # The debug-era arm cue no longer fires automatically on every
            # spoken turn - it was a visual marker for the operator, not a
            # scored note, and had become dead compute on every call.
            # speaker_visual_cue() itself still exists, callable on demand by
            # a score that wants it deliberately (result["speaker_visual_cue"]
            # stays None here, same as the "off"/"monitor" branches above).
            all_spoken = True
            for index, wirepod_text in enumerate(wirepod_chunks):
                say = wirepod_web_send_form(
                    "/api-sdk/say_text",
                    {
                        "text": wirepod_text,
                        "serial": serial,
                        "display_role": display_role or name,
                        "display_text": display_text if index == 0 else "",
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
    # Phase 2: splits a single LLM completion into (internal_thoughts, spoken_reply)
    # on a line-start "REPLY:" marker. Defensive by design - gemma2:2b (this
    # project's configured model) is small and already needs deterministic
    # repair on most turns even without this extra format demand. If the
    # marker is missing, or the reply half comes back empty, fall back to
    # treating the entire raw text as the spoken reply - byte-identical to
    # pre-Phase-2 behavior, never a crash and never silently empty speech.
    match = re.search(r"(?im)^\s*REPLY:\s*", raw)
    if not match:
        return "", raw
    thoughts = re.sub(r"(?im)^\s*THOUGHTS:\s*", "", raw[: match.start()]).strip()
    reply = raw[match.end():].strip()
    if not reply:
        return "", raw
    return thoughts, reply


def resolve_turn_mode(role, line, scaffold):
    # Job 1: an explicit, per-turn, per-role knob (brobot_2_mode/
    # brobot_1_mode, "canned"|"llm") now wins over the scaffold's
    # exchange_type table - canned-vs-LLM is a real, directly settable
    # control, not just an inferred side effect of a mode name. Additive:
    # a line that doesn't set this explicitly (None/absent, including every
    # line via the old load_section_card() path) falls through to the exact
    # historical exchange_type lookup.
    key = "brobot_2_mode" if role == "BROBOT 2" else "brobot_1_mode"
    explicit = line.get(key)
    if explicit in ("canned", "llm"):
        return explicit
    exchange_type_entry = scaffold.get("exchange_types", {}).get(line.get("exchange_type", ""), {})
    return exchange_type_entry.get("brobot_2_mode" if role == "BROBOT 2" else "brobot_1_mode", "llm")


def resolve_canned_text(role, line):
    # Job 1: when a turn is canned, name the source explicitly
    # (brobot_2_canned_source/brobot_1_canned_source) so a mixer can show
    # "this line is canned, and it comes from HERE" instead of inferring it
    # from which role is speaking. Falls back to the exact historical rule
    # (Brobot 2 <- visible_line, Brobot 1 <- value_points space-joined) when
    # a line doesn't name a source explicitly.
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
    # Two-dial framework, rung 1: Brobot 1 reads the same locked, ordered
    # value-point list as the interviewer (never a reordering or a subset -
    # the interviewer's question may reach for fewer, but the answer still
    # covers all of them), rendered as a numbered list instead of one flat
    # blob. Falls back to the old free-text purpose field only if a line has
    # neither a value_points list nor a silent_value_cue at all.
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


# Mechanical echo-detection guard: verbatim N-word overlap check between
# Brobot 1's LLM output and (a) the line's visible_line, (b) Brobot 2's raw
# response. The detection functions below (echo_check_words/echo_ngrams/
# detect_echo) are unchanged and still detection-only - what consumes their
# result now retries a bounded number of times (see ECHO_RETRY_LIMIT and
# generate_exchange_record's Layer 2 brake) instead of only ever logging.
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


# Echo defect Layer 2 (retry brake): a small, bounded ceiling - this many
# regenerations of Brobot 1's turn ALONE (never Brobot 2's, which the
# detector never checks) after the first attempt, so up to
# ECHO_RETRY_LIMIT + 1 total attempts. Small on purpose: this is a brake for
# whatever Layer 1 (the prompt-side fix) misses, not the primary fix itself.
ECHO_RETRY_LIMIT = 3


def _check_echo(brobot_1_candidate, line, brobot_2_generated, echo_word_count):
    candidate = brobot_1_candidate.get("raw_llm_response", "")
    snippet = detect_echo(candidate, line.get("visible_line", ""), n=echo_word_count)
    source = "visible_line" if snippet else None
    if not snippet:
        snippet = detect_echo(candidate, brobot_2_generated.get("raw_llm_response", ""), n=echo_word_count)
        source = "brobot_2_response" if snippet else None
    return bool(snippet), source, snippet


# Self-repeat guard (own-prior-turn leak) - a SEPARATE defect from the echo
# guard above, layered alongside it, not merged into it. The echo guard
# above only ever checks Brobot 1's CURRENT-line candidate against that
# SAME line's visible_line/Brobot 2 response (a cross-speaker,
# within-the-current-line leak). This guard checks a role's new turn
# against that SAME role's own immediately-prior line (a same-speaker,
# cross-line leak) - confirmed live 2026-07-14: Brobot 1's line 3 opened by
# re-speaking almost the entirety of its own line 2 answer before pivoting
# to new content ("Ron was this total visual wizard..." repeated verbatim
# from line 2 into line 3). Root cause: llm_colour_line()'s "PRIOR EXCHANGE
# CONTEXT:" prompt block gets fed to every turn regardless of role, and a
# small model (gemma2:2b) can just continue it instead of writing something
# new - this remained true even after that block was given an explicit
# "never repeat" label (Layer 1, same session), so a Layer 2 detect+retry
# brake is needed here too, exactly like the echo guard's own Layer 2.
# Reuses the same detect_echo() n-gram primitive - never a second detector
# implementation - against a different, independently-tracked source text.
# Applies to BOTH roles (own_prior_raw is looked up per-role) since either
# speaker can in principle repeat their own earlier turn, not just Brobot 1.
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
    # Self-repeat guard, Brobot 2: only meaningful in llm mode - a canned
    # turn is fixed/scripted text with no LLM call to regenerate, and
    # comparing it against a prior turn would either be a no-op or a false
    # positive (e.g. a repeated scripted phrase is intentional, not a leak).
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
    # Self-repeat guard, Brobot 1: symmetric to Brobot 2's above, runs
    # BEFORE the existing cross-speaker echo guard further down (which
    # operates on whatever brobot_1_generated results from here) - the two
    # guards run in sequence, each bounded and independent, neither
    # touching the other's retry ceiling or history.
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
        # Self-repeat guard results (own-prior-turn leak, layered alongside
        # the cross-speaker echo guard below - separate limit, separate
        # history, never merged).
        "self_repeat_retry_limit": SELF_REPEAT_RETRY_LIMIT,
        "brobot_2_self_repeat_detected": brobot_2_self_repeated,
        "brobot_2_self_repeat_history": brobot_2_self_repeat_history,
        "brobot_1_self_repeat_detected": brobot_1_self_repeated,
        "brobot_1_self_repeat_history": brobot_1_self_repeat_history,
        # Scored interview movement - carried straight through from the
        # knobs.json-sourced line dict to the played-back event, absent/empty
        # means silence. Never touches either generated_turn() call above -
        # movement is decided and dispatched at playback time only, the LLM
        # never sees it.
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

    # Job 2: was scaffold-only ("scaffold.get('exchange_types')..."), a
    # second, separate place deciding canned-vs-LLM - now goes through the
    # same resolve_turn_mode() as generated_turn() itself, so an explicit
    # per-turn brobot_1_mode override can't disagree with the echo-check
    # gate. Also Job 2: echo_detection_word_count now reads from
    # scaffold.generation_defaults, falling back to the same constant.
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
        # Echo defect Layer 2: discard and regenerate Brobot 1's turn alone,
        # bounded by ECHO_RETRY_LIMIT - real LLM spend every attempt, never
        # hidden (every generated_turn() call below fires its own
        # LLM_START/LLM_DONE status beats, same as the first attempt did).
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
    # Display lane now carries the same cleaned text the mouth speaks, not
    # the raw LLM response - generated["speech_text"] is the exact
    # flatten_for_robot_speech() product generated_turn() already computed
    # at generation time (the same function Robots.say() itself
    # independently re-runs at playback time on the same inputs, so this is
    # not a second, possibly-diverging text - it's the same cleaned product,
    # just read here instead of recomputed). Screen now equals speech, per
    # turn, matching the channels registry's own host_speaker pattern
    # (Brobot 3's narration: one string, spoken and displayed identically)
    # instead of fighting it. See DISPLAY_LANE_SURVEY_STOP_001.md.
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
    # Voice lane: any destination other than "robot" never contacts Vector,
    # full stop - this is the one place that guarantee is enforced,
    # regardless of what the live gate env var says, so a monitor/off run
    # can never be one env var away from an accidental live robot call.
    live = live and (voice_destination == "robot")
    _KOKORO_STATUS_STATE["live"] = live
    _KOKORO_STATUS_STATE["voice_destination"] = voice_destination
    # Score architecture: the song folder (story.md + knobs.json) is the real
    # architecture - it now loads by default with no env var set.
    # GOPOD_SECTION_SONG_DIR still names a different song explicitly,
    # unchanged, and always wins when set. The legacy .txt Section Card is
    # opt-in only now, reached via GOPOD_USE_LEGACY_SECTION_CARD=1 - and only
    # when no song dir was explicitly named (an explicit song dir always
    # means "load this song," never overridden by the legacy flag). Hoisted
    # up from its original spot (right before "SECTION_CARD_LOAD" below,
    # unchanged) only far enough to answer one question before the
    # model-picker/Kokoro-warmup decisions just below need it - the actual
    # section_card load itself still happens at its original point, in its
    # original order, unchanged.
    section_song_dir = os.getenv("GOPOD_SECTION_SONG_DIR")
    use_legacy_card = (not section_song_dir) and os.getenv(LEGACY_SECTION_CARD_ENV, "0") == "1"
    if use_legacy_card:
        section_source = CARD_PATH
    elif section_song_dir:
        section_source = Path(section_song_dir).expanduser()
    else:
        section_source = DEFAULT_SECTION_SONG_DIR
    # Bare-capture lane: a song opts out of the model-picker prompt, the
    # Kokoro status-announcer voice, and (main()'s own finally block) the
    # timing-log write, all via one data flag on the song itself
    # (knobs.json["bare_capture"]) - never inferred from song name, never a
    # per-song code branch here or anywhere else. Absent/false (every
    # existing song, including the interview) means today's behavior,
    # byte-for-byte - this peek only ever changes anything when a song
    # explicitly opts in. See brobots_bait_001's own knobs.json.
    bare_capture = peek_song_bare_capture(section_source, use_legacy_card)
    # LLM lane: model choice happens before the pre-show narration even
    # starts - true "run start," not buried after a couple minutes of
    # Kokoro narration. A replay never calls the LLM for content (see the
    # SECTION1_REPLAY_LOG branch below), so it never needs to ask; content_model
    # stays "" here and gets filled from the replay log itself instead. A
    # bare_capture song needs no model either - it has no LLM lines by
    # definition, so asking would be a prompt with nothing behind it.
    # Additive: a caller that already resolved a model (the scored pre-show
    # song path, which needs one for its own robot readiness lines too, and
    # passes the same pick here for consistency) skips re-resolving -
    # resolve_content_model() can show an interactive menu and write to
    # content_model_state.json, so it must only ever run once per real run.
    # content_model is None for every existing caller, unchanged.
    if content_model is None:
        content_model = "" if (SECTION1_REPLAY_LOG or bare_capture) else resolve_content_model()
    if content_model:
        print(f"CONTENT_MODEL_SELECTED model={content_model}", flush=True)
    if bare_capture:
        # Silences every _announce_status_kokoro call for the rest of this
        # run (the same flag the scored pre-show path already uses for its
        # own, different reason) - RUN_START/SCAFFOLD_LOAD/etc still print
        # to the console and still land in the run json, they just never
        # reach Kokoro's mouth or Wire-Pod's own rich-display log window.
        # configure_voice_engines()/kokoro_status_warm_up() are skipped
        # entirely, not just silenced - a bare-capture run never spawns the
        # Kokoro subprocess or pays its ~10-20s cold warm-up at all.
        _KOKORO_STATUS_STATE["active"] = False
    else:
        # Voice lane hardening: a silent, unnarrated scaffold read purely to
        # gate/configure the pre-show narration engine before its one-time
        # subprocess spawn. No status()/print call here - this does not add or
        # move any line in the run's own narrated status log; the real,
        # narrated SCAFFOLD_LOAD/SCAFFOLD_READY status lines still fire at
        # their original point below, unchanged, and load the scaffold again
        # (a second fast local JSON read, immaterial).
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
            # LLM lane: carry the source run's own model forward so a replay's
            # run json still names the brain that actually generated it.
            log["content_model"] = replay_log.get("content_model", "")
            status(f"GENERATION_REPLAYED events={len(log['events'])}")
        else:
            # Job 2: how much memory each turn gets (how many prior
            # exchanges, the prefix labels, the truncation length) was
            # hardcoded here - now scaffold.generation_defaults, same
            # numbers/strings as before by default.
            generation_defaults = interview_scaffold.get("generation_defaults", {})
            prior_window = generation_defaults.get("prior_context_exchange_count", 2)
            prior_truncate = generation_defaults.get("prior_context_truncate_chars", 180)
            interviewer_prefix = generation_defaults.get("prior_context_interviewer_prefix", "Previous interviewer line: ")
            interviewee_prefix = generation_defaults.get("prior_context_interviewee_prefix", "Previous response value: ")
            prior_exchange_context = []
            # Self-repeat guard's own tracking: the immediately-prior line's
            # FULL raw response per role (never truncated, unlike
            # prior_exchange_context above which is compacted/truncated for
            # prompt-injection purposes) - a proper n-gram overlap check
            # needs the real text, not a 180-char preview.
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
        # Turn order, whole-run, data-driven - read once here, never
        # per-line, never inferred from song name. Absent/unset means
        # "brobot_2" (today's interview order, unchanged). See
        # load_section_pair()'s own "first_speaker" field.
        first_speaker = log.get("section_card", {}).get("first_speaker", "brobot_2")
        if first_speaker not in ("brobot_2", "brobot_1"):
            raise RuntimeError(f"BLOCKED: first_speaker={first_speaker!r} must be 'brobot_2' or 'brobot_1'")
        # Breathing pause, whole-run, data-driven - read once here, same
        # pattern as first_speaker above. Absent/0 means no pause (today's
        # interview/pre-show timing, unchanged). See load_section_pair()'s
        # own "movement_pause_seconds" field.
        movement_pause_seconds = float(log.get("section_card", {}).get("movement_pause_seconds", 0) or 0)
        # Interview-only theatrical KG search sequence - computed once per
        # run, not per line, same pattern as first_speaker/
        # movement_pause_seconds above. See is_interview_kg_search_song()/
        # fire_interview_kg_search_sequence() for the full rationale.
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
                    # Second scored movement, fired after the first - absent
                    # for every existing song (interview, pre-show), so this
                    # is a silent no-op there. See load_section_pair()'s own
                    # brobot_2_movement_2 field.
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
                    # Second scored movement, fired after the first - absent
                    # for every existing song (interview, pre-show), so this
                    # is a silent no-op there. See load_section_pair()'s own
                    # brobot_1_movement_2 field.
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
                    # Interview-only theatrical KG search sequence - only for
                    # an LLM-generated turn (a canned line, e.g. the closer,
                    # never "searched" for anything, so it's excluded via
                    # resolve_turn_mode()). Fires searching->searchingGetout
                    # on Brobot 1, then overrides this turn's own leading
                    # animation tag to "answering" - robot_safe_line is
                    # rebuilt (not just animation_action) because
                    # Robots.say() re-derives the action fresh from
                    # robot_safe_line's own "<action>... <text>" prefix via
                    # split_robot_safe_line(), every call, not from a
                    # separately-read animation_action field.
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
    # Host dialogue is authored (canned), never LLM-coloured - only the
    # robots' own readiness reactions are. Routes to whichever host's own,
    # fully independent Kokoro subprocess this beat's "speaker" names.
    # read_sheet=True skips that subprocess call entirely - the text is
    # already fully known (authored, not generated), so there is nothing to
    # capture here beyond what run_preshow_song's own record() already logs.
    # segment_id defaults to the beat's own step_id (each m1-m4 beat fires
    # once) but a caller that repeats the same beat (the vamp loop, which
    # can loop the same handful of beats many times in one run) passes an
    # explicit per-round id so every iteration gets its own timing entry.
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
    # Movements 2/3: the robot's own scored, LLM-coloured readiness line -
    # coloured against the pre-show's own hint (never hand-typed here) via
    # the exact same llm_colour_preshow_line + normalize_robot_safe pair
    # either way, live run or read-sheet - generation is real and unchanged
    # in both modes. read_sheet=True only ever skips the final robots.say()
    # dispatch (the arm cue, the Kokoro/Wire-Pod call), never the LLM call
    # that produced the line being reviewed.
    #
    # Returns (result, voice_text) - voice_text is the fully processed text
    # (normalize_robot_safe -> split_robot_safe_line -> flatten_for_robot_speech,
    # the same three-step pipeline robots.say() itself runs internally), not
    # the bare LLM raw output. This is deliberate: the read-sheet's whole job
    # is showing what the operator can trust reached (or would reach) the
    # voice lane, and the normalize_robot_safe hardening pass lives upstream
    # of this - logging pre-hardening raw here would make the read-sheet
    # permanently report contamination that no longer happens.
    #
    # Reads brobots_1_2_persona's own system_prompt, not the shared
    # top-level system_prompt every interview line reads - same precedent as
    # llm_colour_status_beat reading brobot_3_host's own dedicated block
    # (line ~415). This function is the sole call path for m2_c/m3_c (the
    # only two pre-show steps with speaker brobot_1/brobot_2), so this read
    # never reaches any interview line or any other pre-show beat.
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
            # Display lane carries voice_text (the already-computed cleaned
            # product, same flatten_for_robot_speech() pipeline Robots.say()
            # itself independently re-runs), not raw - screen equals speech
            # here too, matching play_generated_turn()'s own fix and the
            # channels registry's host_speaker pattern.
            result = robots.say(name, robot_safe_line, display_text=voice_text, display_role=name)
    print(f"PRESHOW {step['step_id']}: {name}: {robot_safe_line}", flush=True)
    return result, voice_text


def format_preshow_read_sheet(song_id, run_id, content_model, entries):
    # Plain, human-read stage-sheet shape, not JSON - this is for the
    # operator's eyes between takes, not another machine-readable log (the
    # real run json already exists for that, untouched by this).
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
        # Dedupes by step_id so a vamp beat that loops many times in a real
        # run (however long generation actually took) still shows up exactly
        # once - a stable, complete beat listing of the song's own shape,
        # not a run-length-dependent transcript of the filler loop.
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
        # generate_phase() returns a (log, interview_scaffold, robots) tuple,
        # not the log dict alone - element 0 is the log.
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

    # No generation thread this time - pre-set so run_preshow_song()'s vamp
    # loop condition (`while not generation_done_event.is_set()`) is false
    # immediately and the filler loop body never runs once.
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
    # Channel-1 song brackets: one start, one end, around the whole
    # generate+playback stretch - the runner's own single clean start/end
    # point for a real performed song (a standalone preshow-only run via
    # run_section1_preshow_generate_001.py never reaches main(), so it never
    # fires these - correct: that leg is script prep, not a channel-1
    # performance). live must be resolved and applied to
    # _KOKORO_STATUS_STATE before the START marker's own gate check, and
    # passed through explicitly so generate_phase() doesn't re-derive it.
    # voice_destination is resolved here for the same reason - a destination
    # other than "robot" must neuter live before that same gate check, so a
    # monitor/off run's channel-1 markers never touch Vector either.
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
        # Timing self-log: fires in every case, including a mid-run failure -
        # a partial timing map from a crashed run is still real, useful data,
        # never worth losing. Never touches the interview/pre-show run json
        # above (write_log(log)) - a fully separate file, same "runs/"
        # working location, same private-by-default status as the existing
        # rehearsal JSONs there. Skipped entirely for a bare_capture song -
        # see generate_phase()'s own bare_capture peek - which by
        # definition has no timing story worth logging (no LLM generation,
        # no vamp, no pre-show) and asked explicitly not to write one.
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

#!/usr/bin/env python3
"""phcal Rung 3 - isolate-and-watch, ONE mechanical primitive, one robot, then
stop, now with cycles/hold/speed adjustable per call AND a guided prompt flow
plus write-back memory on top. Sibling to pha0b: pha0b plays the SCORE (dry
navigation of a song slice); phcal tunes the INSTRUMENT (live motion in
isolation). Rung 1 isolated and played current behavior only, no value
changed. Rung 2 added --hold/--speed/--cycles flags (and nod's existing
positional [count]) to override the rung-1 defaults per call - the defaults
themselves unchanged. Rung 3 adds:

  - A guided flow, triggered by calling `phcal` with no primitive at all:
    asks arm-or-nod, then a robot, then walks each value (cycles/count,
    hold, speed) with the prompt pre-filled by the last-used value for that
    primitive - Enter keeps it, typing a new number overrides it. Fires the
    same way the direct-flag form already does (live if the gate's on, dry
    if it's off), then saves the values actually used.
  - Write-back memory at ~/.gopod_alias_lib/phcal_last.json (this tool's own
    memory file, outside the tracked GOPOD repo - NOT any song's knobs.json/
    story.md). Separate keys per primitive (arm: cycles/hold/speed; nod:
    count/hold/speed). Both entry paths - the new guided flow AND the
    existing direct-flag form (`phcal arm 1 --hold 0.5 ...`) - write to this
    same one file, so a value set either way becomes the next [last]. First
    run with no file on disk seeds [last] from the same rung-1 bingo-sourced
    defaults below, then saves after that first run.

Sourced from, not a copy of: run_songs_runner_001.py's own
run_move_axis()/run_arm_cue()/run_nod() (single-serial per-robot HTTP shape:
assume_behavior_control -> move_lift/move_head speed set -> hold -> speed=0
stop -> release_behavior_control). Reproduced with a printed, timestamped line
in front of every sub-instruction, so the operator can watch each mechanical
step land, whatever values (default or flag-overridden) are actually in play
- that source file's own functions return a result dict only, no line-by-line
print, which rung 1 added and rung 2 keeps.

2026-07-22 control-holding fix: assume/release_behavior_control now fire
ONCE per cmd_arm/cmd_nod call (before/after the whole cycle loop), not once
per direction/cycle - matching run_named_movement_sequence's own
(run_robot_control_song_001.py) assume-once/release-once shape, the same
fix that keeps the robot's screen from flashing to its home icon between
notes during a full song run. See run_move_axis_no_control's own comment
block, below, for the full root-cause writeup. Rung-1 (bingo-sourced)
defaults, unchanged, used whenever a flag is omitted:
  arm_cue: cycles=1, hold_seconds=1.2, speed=2 (move_lift, +speed up then
           -speed down)
  nod:     hold_seconds=0.35, speed=2          (move_head, -speed down then
           +speed up, in that order - the source code's own down-then-up
           shape; rung 1 confirmed this against the real bingo runner order,
           correcting an earlier spec's prose that wrongly said
           "up -> down" - kept as down-then-up here, do not revert).

Flags (rung 2, new):
  --hold S     hold_seconds override (float, > 0). Valid for arm and nod.
  --speed N    speed-magnitude override (int, > 0; applied as +N/-N per
               direction, same convention as the default). Valid for arm
               and nod.
  --cycles N   arm-cue repeat-count override (int, >= 1). arm-only - nod
               repeats via its existing positional [count] instead, not a
               flag, so there is exactly one way to say "how many" per
               primitive.

NOT a song runner: no song folder, no knobs.json, no story.md, no step loop,
nothing read from or written to any song's own files - tuning stays IN this
tool, it is never written back to a song's own files. One primitive, one
robot, then exit.

Live gate: same convention as every other live alias in
~/.gopod_alias_lib/ - GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 must be exported by the
caller for real HTTP calls to fire. Unset (or anything else) = dry: prints
the exact same planned sub-instructions (reflecting whatever values are in
play), sends nothing, holds nothing.

Hard limits kept from rungs 1/2, unchanged by rung 3: alias-lib only, no
repo/song/knobs.json/story.md touch; one robot per call, no two-robot sync;
nod stays down-then-up.

Rung 4 (2026-08-08, GOLDEN_BROBOTS_CONTROL_CATALOG_SURVEY_001.md /
GOPOD_WIDGET_TARGET_SHAPE_001.md's "instrument rack" step): the guided
flow's primitive pick is now a dynamic vertical numbered menu built from one
dict (_PRIMITIVE_MENU), widened from 4 entries (arm/nod/rattle/weather) to 7
by adding sleep/wake/brobots_ready - the catalog's other standalone-fireable
golden mechanisms, each firing through the exact same direct-SDK binary its
existing brobots.sh/core.sh alias already calls (no second implementation).
None of these three have a dry mode (confirmed against the Go source - no
--dry flag exists anywhere in that binary), so each refuses cleanly
(PHCAL_NO_DRY_MODE) rather than faking one when the live gate is off. Still
NOT a song editor: no knobs.json touched, no song wired to a catalog pick,
arm/nod/rattle/weather's own firing/tuning behavior is byte-unchanged - only
how a mechanism is SELECTED changed.
"""
import importlib.util
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
import termios
import time
import tty
import urllib.parse
import urllib.request
from pathlib import Path

_PHCAL_CONFIG_PATH = Path(__file__).resolve().parent / "phcal_config.json"


def _load_phcal_config():
    if not _PHCAL_CONFIG_PATH.is_file():
        return {}
    try:
        return json.loads(_PHCAL_CONFIG_PATH.read_text(encoding="utf-8"))
    except (json.JSONDecodeError, OSError):
        return {}


_PHCAL_CONFIG = _load_phcal_config()
_BASE_DIRS = _PHCAL_CONFIG.get("base_dirs", {})


def _base_dir(key, default):
    return Path(_BASE_DIRS.get(key, default))


GOPOD_REPO_ROOT = _base_dir("gopod_repo_root", "/home/goverlord/crushn8r_git/GOPOD")
WIREPOD_HOME = _base_dir("wirepod_home", "/home/goverlord/wire-pod")
GOPOD_DIRECT_TOOLS_DIR = GOPOD_REPO_ROOT / "goverlord/wire_pod_overlay/chipper/gopod_probes/tools"


def _resolve_brobot_serials():
    """Brobot 1/2 ESN resolution, in priority order: phcal_config.json's
    robots.candidates (new, N-capable - the list a future detect-first pass
    will probe) if it has at least 2 entries; else the same
    GOPOD_BROBOT_1_SERIAL/GOPOD_BROBOT_2_SERIAL env vars
    run_section1_full_live_001.py has always read; else the same hardcoded
    defaults that module has always fallen back to. Every tier resolves to
    the identical two ESNs for this operator's own real setup today - only
    where they're read from changes."""
    candidates = _PHCAL_CONFIG.get("robots", {}).get("candidates", [])
    if len(candidates) >= 2:
        return candidates[0]["esn"], candidates[1]["esn"]
    return (
        os.getenv("GOPOD_BROBOT_1_SERIAL", "ESN01"),
        os.getenv("GOPOD_BROBOT_2_SERIAL", "ESN02"),
    )


BROBOT_1_SERIAL, BROBOT_2_SERIAL = _resolve_brobot_serials()

RUNNER_PATH = GOPOD_REPO_ROOT / "goverlord/runtime/songs/tools/run_section1_full_live_001.py"

TEMPO_SET_PATH = Path(__file__).resolve().parent / "tempo_set_001.py"
CONTROL_SONG_RUNNER_PATH = GOPOD_REPO_ROOT / "goverlord/runtime/songs/tools/run_robot_control_song_001.py"
KNOBS_ENVELOPE_PATH = GOPOD_REPO_ROOT / "goverlord/runtime/songs/tools/knobs_envelope_001.py"
SONGS_DIR = GOPOD_REPO_ROOT / "goverlord/runtime/songs"

GOPOD_AUDIO_DIR = GOPOD_REPO_ROOT / "goverlord/wire_pod_overlay/chipper/sounds"

RATTLE_BIN_PATH = GOPOD_DIRECT_TOOLS_DIR / "direct_sdk_bingo_rattle_001"
RATTLE_WAV_PATH = GOPOD_AUDIO_DIR / "rattle.wav"

DANGER_WAV_PATH = GOPOD_AUDIO_DIR / "danger-will-robinson.wav"
CUBE_BLIP_BIN_PATH = GOPOD_DIRECT_TOOLS_DIR / "direct_sdk_cube_blip_001"
WIREPOD_HOME_PATH = str(WIREPOD_HOME)
DIRECT_SDK_RELEASE_SETTLE_SECONDS = 1.0

SLEEP_BIN_PATH = GOPOD_DIRECT_TOOLS_DIR / "direct_sdk_robot_sleep_001"
TOGETHER_BIN_PATH = GOPOD_DIRECT_TOOLS_DIR / "direct_sdk_brobots_ready_001"

ROBOT_INFO_BIN_PATH = GOPOD_DIRECT_TOOLS_DIR / "direct_sdk_robot_info_001"
ROBOT_INFO_TIMEOUT_SECONDS = 30

PHCAL_DETECT_PROBE_TIMEOUT_SECONDS = 8

SLEEP_MIN_HOLD_SECONDS = 5.0

_LOOP_ANIMATION_TOKENS = {"searching", "answering"}
_ANIMATION_LOOP_INTERVAL_SECONDS = 0.333

_ANIMATION_TOKEN_MENU = {
    "1": "kgSuccess",
    "2": "searching",
    "3": "answering",
}
ANIMATION_DEFAULT_HOLD_SECONDS = 2.5
_ANIMATION_SEQUENCE_PAUSE_SECONDS = 1.0

_ANIMATION_GETOUT_TOKENS = {
    "searching": "searchingGetout",
}

_ANIMATION_NO_GETOUT_RELEASE_SETTLE_SECONDS = 1.0

HOLD_DEFAULT_SECONDS = 5.0

PHCAL_UNBOUNDED_HOLD_MIN_SECONDS = 0.5
PHCAL_UNBOUNDED_HOLD_MAX_SECONDS = 10.0

PHCAL_ANNOUNCE_PHRASE_MAX_CHARS = 36

MOVE_REVERSE_LW = -150
MOVE_REVERSE_RW = -150
MOVE_REVERSE_DEFAULT_HOLD_SECONDS = 0.3
MOVE_REVERSE_MAX_HOLD_SECONDS = 1.0
MOVE_REVERSE_CAUTION = (
    "MOVE_REVERSE CAUTION: cliff sensors disabled while control held - ON-CHARGER USE ONLY"
)

WHEEL_NUDGE_WAKE_CHECK_TIMEOUT_SECONDS = 15  # CONNECTION_PASS_TIMEOUT_SECONDS's own value, matched - conn_test's own timeout, unchanged
WHEEL_NUDGE_WAKE_RELEASE_SETTLE_SECONDS = 2.5  # measured max observed 1.817s across 8 live runs, both robots - real margin above that ceiling, not the bare minimum; see WHEEL_NUDGE_COLD_FIRE_ROOT_CAUSE_SURVEY_001.md for the run-by-run numbers
WHEEL_NUDGE_WAKE_POST_REASSUME_SETTLE_SECONDS = 0.5

SLEEP_WAKE_DEFAULT_WAIT_SECONDS = 5.0
SLEEP_WAKE_TIMEOUT_SECONDS = 70


def _readiness_signal(ready, reason, channel, detail=None):
    """`ready`/`ok` carry the same boolean (kept both so every existing
    `if not X["ok"]` call site in this file keeps working unchanged -
    standardizing the signal must not silently change any existing
    control-flow check). `reason` is always a short, human-readable string,
    success or failure. `channel` names which mechanism produced this -
    "rest_golden_flag" or "direct_sdk_continuous". `detail` carries
    whatever mechanism-specific sub-results already existed, untouched."""
    return {"ready": ready, "ok": ready, "reason": reason, "channel": channel, "detail": detail or {}}


def cmd_sleep_wake_set_time(mod, clock, which, wait_seconds):
    """Combined sleep->wake bench test, timed-release mode. One continuous
    BehaviorControl connection (direct_sdk_robot_sleep_001's own
    --wait-signal mode) holds both chosen robots asleep, then release fires
    in that SAME connection the instant the signal file appears - no
    release-then-regrant gap, the exact fix gopod-song-open() already proved
    live. Release trigger here: a plain background wait of `wait_seconds`."""
    specs = _sleep_wake_specs(mod, which)
    signal_file = tempfile.mktemp(prefix="phcal_sleep_wake_signal_")
    print(
        f"{clock.prefix()}PHCAL_SLEEP_WAKE which={which} mode=set_time "
        f"wait_seconds={wait_seconds} specs={' '.join(specs)}"
    )
    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    proc = subprocess.Popen(
        [str(SLEEP_BIN_PATH), "sleep", "--wait-signal", signal_file, *specs],
        stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, env=env,
    )
    print(f"{clock.prefix()}PHCAL_SLEEP_WAKE holding {wait_seconds}s - both robots should stay asleep on one continuous connection")
    remaining = wait_seconds
    while remaining > 0:
        step = min(1.0, remaining)
        time.sleep(step)
        remaining = round(remaining - step, 3)
        if remaining > 0:
            print(f"{clock.prefix()}PHCAL_SLEEP_WAKE {remaining:.1f}s remaining")
    print(f"{clock.prefix()}PHCAL_SLEEP_WAKE releasing now (signal file touched -> GoToSleepOff fires in the same connection)")
    Path(signal_file).touch()
    signal = _finish_sleep_wake(proc, signal_file, clock, "set_time")
    return 0 if signal["ok"] else 1


def cmd_sleep_wake_on_process(mod, clock, live, which):
    """Combined sleep->wake bench test, process-signaled release mode - same
    continuous-connection mechanism as cmd_sleep_wake_set_time, but the
    release trigger is a real completed process instead of a timer:
    run_restart_wirepod_preflight() (phcal's own wpr-equivalent check), the
    same real fill-work gopod-song-open() itself runs during its own sleep
    window - not a fake stand-in. Honest caveat, same one that file's own
    build history already names: if Wire-Pod is already healthy this check
    can complete in milliseconds, so the sleep window may be very short -
    not lengthened artificially here."""
    specs = _sleep_wake_specs(mod, which)
    signal_file = tempfile.mktemp(prefix="phcal_sleep_wake_signal_")
    print(f"{clock.prefix()}PHCAL_SLEEP_WAKE which={which} mode=on_process specs={' '.join(specs)}")
    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    proc = subprocess.Popen(
        [str(SLEEP_BIN_PATH), "sleep", "--wait-signal", signal_file, *specs],
        stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, env=env,
    )
    print(f"{clock.prefix()}PHCAL_SLEEP_WAKE both robots asleep - running a real restart_wirepod_preflight check as the fill-work (may be near-instant if Wire-Pod is already healthy)")
    run_restart_wirepod_preflight(mod, clock, live)
    print(f"{clock.prefix()}PHCAL_SLEEP_WAKE fill-work done, releasing now (signal file touched -> GoToSleepOff fires in the same connection)")
    Path(signal_file).touch()
    signal = _finish_sleep_wake(proc, signal_file, clock, "on_process")
    return 0 if signal["ok"] else 1


def _finish_sleep_wake(proc, signal_file, clock, mode):
    """Direct-SDK channel's own gate-if-not-ready endpoint - builds and
    returns the standardized readiness signal (_readiness_signal above),
    same shape cmd_brobots_wake's REST-channel gate returns."""
    try:
        stdout, _ = proc.communicate(timeout=SLEEP_WAKE_TIMEOUT_SECONDS)
    except subprocess.TimeoutExpired:
        proc.kill()
        reason = f"direct-SDK sleep/wake process timed out after {SLEEP_WAKE_TIMEOUT_SECONDS}s"
        print(f"{clock.prefix()}PHCAL_SLEEP_WAKE_TIMED_OUT")
        print(f"{clock.prefix()}PHCAL_READY ready=False reason={reason!r} channel=direct_sdk_continuous")
        Path(signal_file).unlink(missing_ok=True)
        return _readiness_signal(False, reason, "direct_sdk_continuous", detail={"mode": mode, "timed_out": True})
    Path(signal_file).unlink(missing_ok=True)
    combined = "\n".join(line for line in stdout.splitlines() if "guid" not in line.lower())
    ok = proc.returncode == 0
    reason = (
        f"continuous connection held sleep->wake cleanly (mode={mode})" if ok
        else f"direct-SDK sleep/wake binary exited {proc.returncode} (mode={mode})"
    )
    print(f"{clock.prefix()}{combined}")
    print(f"{clock.prefix()}PHCAL_SLEEP_WAKE_COMPLETE ok={ok} mode={mode}")
    print(f"{clock.prefix()}PHCAL_READY ready={ok} reason={reason!r} channel=direct_sdk_continuous")
    return _readiness_signal(ok, reason, "direct_sdk_continuous", detail={"mode": mode, "returncode": proc.returncode})


PHCAL_RATTLE_SETTLE_SECONDS = 3.0

PHCAL_CYCLE_SETTLE_SECONDS = 1.0

PHCAL_PREFLIGHT_SETTLE_SECONDS = 1.5  # run_wake_both's own fallback value

PHCAL_ASSUME_SETTLE_SECONDS = PHCAL_PREFLIGHT_SETTLE_SECONDS

PHCAL_PREFLIGHT_RECOVERY_MESSAGE = (
    "sequence error - Wire-Pod could not reach the robot. Field-proven recovery "
    "order (robot_control_song_001/story.md's own Troubleshooting section, confirmed "
    "live - a partial version does not clear it): 1) power-cycle the robot, "
    "2) re-pair it with Wire-Pod, 3) restart Wire-Pod (wpr). Stopping here, not "
    "guessing a different fix."
)

PHCAL_RESTART_WIREPOD_ENV = "GOPOD_RESTART_WIREPOD_BEFORE_RUN"


def run_restart_wirepod_preflight(mod, clock, live):
    """Calls run_section1_full_live_001.py's own restart_wirepod_preflight(live)
    exactly as written - see the comment block above PHCAL_RESTART_WIREPOD_ENV
    for why this exists and why phcal defaults its own env var on. Wrapped in
    try/except: restart_wirepod_preflight raises RuntimeError if the service
    never comes active or HTTP never responds, which is exactly what this
    handler exists to catch, same pattern as run_preflight's (conn_test)
    existing failure handling - print, stop, do not fire any move command."""
    if not live:
        print(f"{clock.prefix()}PHCAL_PREFLIGHT DRY: would restart_wirepod_preflight (wrp-equivalent) before conn_test")
        return {"ok": True, "mode": "dry"}
    print(f"{clock.prefix()}PHCAL_PREFLIGHT restart_wirepod_preflight starting (sudo systemctl restart wire-pod, then poll for ready)")
    try:
        result = mod.restart_wirepod_preflight(live)
    except RuntimeError as exc:
        print(f"{clock.prefix()}PHCAL_PREFLIGHT_FAILED restart_wirepod_preflight - {exc}")
        return {"ok": False, "error": str(exc)}
    if result.get("enabled"):
        print(
            f"{clock.prefix()}PHCAL_PREFLIGHT restart_wirepod_preflight enabled=True "
            f"service_state={result.get('service_state')} http_ready={result.get('http_ready')}"
        )
    else:
        print(
            f"{clock.prefix()}PHCAL_PREFLIGHT restart_wirepod_preflight enabled=False - "
            f"{PHCAL_RESTART_WIREPOD_ENV} was not '1' at the time `mod` loaded (phcal sets "
            "this itself by default via setdefault; only happens if something already set "
            "it to a different value first) - skipped, proceeding to conn_test preflight "
            "without a fresh service restart"
        )
    return {"ok": True, "result": result}


LAST_PATH = Path(__file__).resolve().parent / "phcal_last.json"


def _load_module(path, name):
    spec = importlib.util.spec_from_file_location(name, str(path))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


class _Clock:
    """Same "[HH:MM:SS.mmm +X.XXXs] " shape as Robots.timing_prefix() in
    run_section1_full_live_001.py - reproduced standalone here rather than
    constructing a full Robots() (which needs load_interview_scaffold() and
    other interview-only state this isolated tool has no use for)."""

    def __init__(self):
        self._last = None

    def prefix(self):
        now = time.time()
        clock = time.strftime("%H:%M:%S", time.localtime(now)) + f".{int(now % 1 * 1000):03d}"
        if self._last is None:
            self._last = now
            return f"[{clock} +0.000s] "
        delta = now - self._last
        self._last = now
        return f"[{clock} +{delta:.3f}s] "


def run_move_axis_no_control(mod, clock, serial, endpoint, speed, hold_seconds, live, direction_label):
    """Move-only half of the old run_move_axis_isolated: {endpoint}
    speed=speed -> sleep hold_seconds -> {endpoint} speed=0 -> stop. Does NOT
    assume or release behavior control - the caller (cmd_arm/cmd_nod) now
    holds control across the entire cycle loop, assumed once before this is
    ever called and released once after every direction/cycle has fired, so
    the robot never drops back to idle between individual moves. Prints each
    sub-instruction as it fires, live or dry, same style as before."""
    tag = f"PHCAL direction={direction_label}"

    if not live:
        print(f"{clock.prefix()}{tag} DRY: would {endpoint} serial={serial} speed={speed}")
        print(f"{clock.prefix()}{tag} DRY: would hold {hold_seconds}s")
        print(f"{clock.prefix()}{tag} DRY: would {endpoint} serial={serial} speed=0 (stop)")
        return {"ok": True, "mode": "dry"}

    print(f"{clock.prefix()}{tag} {endpoint} serial={serial} speed={speed}")
    move = mod.wirepod_web_send_form(f"/api-sdk/{endpoint}", {"serial": serial, "speed": speed}, timeout=10)

    print(f"{clock.prefix()}{tag} hold {hold_seconds}s")
    time.sleep(hold_seconds)

    print(f"{clock.prefix()}{tag} {endpoint} serial={serial} speed=0 (stop)")
    stop = mod.wirepod_web_send_form(f"/api-sdk/{endpoint}", {"serial": serial, "speed": 0}, timeout=10)

    ok = move.get("status") == 200 and stop.get("status") == 200
    print(f"{clock.prefix()}PHCAL_DONE direction={direction_label} ok={ok}")
    return {"ok": ok, "move": move, "stop": stop}


def run_move_wheels_no_control(mod, clock, serial, lw, rw, hold_seconds, live):
    """Wheel twin of run_move_axis_no_control above - same move -> hold ->
    stop shape, but move_wheels takes TWO independent speed params (lw/rw),
    not one, so this is a new function rather than a forced reuse of the
    single-speed helper (WHEEL_NUDGE_GOLDEN_PATH_SURVEY_001.md, section A).
    Does NOT assume or release behavior control - the caller
    (cmd_move_reverse) holds control across this one call, same convention as
    every other *_no_control helper here. The stop call is unconditional -
    fires even when the move call itself reports a non-200 - a wheel move
    must never be left without a matching stop."""
    tag = "PHCAL_MOVE_REVERSE"

    if not live:
        print(f"{clock.prefix()}{tag} DRY: would move_wheels serial={serial} lw={lw} rw={rw}")
        print(f"{clock.prefix()}{tag} DRY: would hold {hold_seconds}s")
        print(f"{clock.prefix()}{tag} DRY: would move_wheels serial={serial} lw=0 rw=0 (stop)")
        return {"ok": True, "mode": "dry"}

    print(f"{clock.prefix()}{tag} move_wheels serial={serial} lw={lw} rw={rw}")
    try:
        move = mod.wirepod_web_send_form(
            "/api-sdk/move_wheels", {"serial": serial, "lw": lw, "rw": rw}, timeout=10
        )
    except Exception as exc:  # noqa: BLE001 - the stop below must still fire on a move error
        print(f"{clock.prefix()}{tag} move_wheels FAILED - {exc!r}")
        move = {"status": None, "error": repr(exc)}

    print(f"{clock.prefix()}{tag} hold {hold_seconds}s")
    time.sleep(hold_seconds)

    print(f"{clock.prefix()}{tag} move_wheels serial={serial} lw=0 rw=0 (stop)")
    stop = mod.wirepod_web_send_form("/api-sdk/move_wheels", {"serial": serial, "lw": 0, "rw": 0}, timeout=10)

    ok = move.get("status") == 200 and stop.get("status") == 200
    print(f"{clock.prefix()}PHCAL_DONE direction=reverse ok={ok}")
    return {"ok": ok, "move": move, "stop": stop}


def run_wheel_nudge_wake_check(mod, clock, serial, live):
    """conn_test only - a cheap, real pre-check ahead of assuming control,
    matching gopod-opening-chord's own separate WIREPOD_CHAIN_WAKE_START
    conn_test (fired before its own speak() call). Live field-testing
    (2026-08-09, 2nd round) showed this alone does NOT fix a cold-fire
    first-run miss - it proves connectivity, not that THIS run's own
    upcoming assume has settled. See run_wheel_nudge_wake_release_pulse
    below for the piece that actually does that."""
    if not live:
        print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_CHECK DRY: would conn_test serial={serial}")
        return {"ok": True, "mode": "dry"}
    print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_CHECK conn_test serial={serial}")
    try:
        conn = mod.wirepod_web_send_form(
            "/api-sdk/conn_test", {"serial": serial}, timeout=WHEEL_NUDGE_WAKE_CHECK_TIMEOUT_SECONDS
        )
        ok = conn.get("status") == 200
    except Exception as exc:  # noqa: BLE001 - a hang/timeout here is exactly what this check exists to catch fast
        print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_CHECK_FAILED serial={serial} - {exc!r}")
        return {"ok": False, "error": repr(exc)}
    print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_CHECK_DONE serial={serial} ok={ok}")
    return {"ok": ok, "conn_test": conn}


def run_wheel_nudge_wake_release_pulse(mod, clock, serial, live):
    """REPLACED 2026-08-10 (WHEEL_NUDGE_COLD_FIRE_ROOT_CAUSE_SURVEY_001.md):
    the real, measured fix - see WHEEL_NUDGE_WAKE_RELEASE_SETTLE_SECONDS's
    own comment above for the full why. Fired INSIDE the caller's
    assume/release window, immediately after assume and before the actual
    move: release_behavior_control (the "golden flag" pulse) -> settle ->
    re-assume_behavior_control. This is what proves the assume just fired
    has actually settled - not a guessed sleep, not a say_text call that
    turned out to prove nothing."""
    if not live:
        print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_RELEASE_PULSE DRY: would release_behavior_control -> settle {WHEEL_NUDGE_WAKE_RELEASE_SETTLE_SECONDS}s -> re-assume_behavior_control serial={serial}")
        return {"ok": True, "mode": "dry"}
    print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_RELEASE_PULSE release_behavior_control serial={serial}")
    try:
        release_pulse = mod.wirepod_web_send_form(
            "/api-sdk/release_behavior_control", {"serial": serial}, timeout=10
        )
        print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_RELEASE_PULSE settle {WHEEL_NUDGE_WAKE_RELEASE_SETTLE_SECONDS}s")
        time.sleep(WHEEL_NUDGE_WAKE_RELEASE_SETTLE_SECONDS)
        print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_RELEASE_PULSE re-assume_behavior_control serial={serial}")
        reassume = mod.wirepod_web_send_form(
            "/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10
        )
        time.sleep(WHEEL_NUDGE_WAKE_POST_REASSUME_SETTLE_SECONDS)
        ok = release_pulse.get("status") == 200 and reassume.get("status") == 200
    except Exception as exc:  # noqa: BLE001 - a hang/timeout here is exactly what this check exists to catch fast
        print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_RELEASE_PULSE_FAILED serial={serial} - {exc!r}")
        return {"ok": False, "error": repr(exc)}
    print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_RELEASE_PULSE_DONE serial={serial} ok={ok}")
    return {"ok": ok, "release_pulse": release_pulse, "reassume": reassume}


def run_assume_control(mod, clock, serial, live):
    """Fires assume_behavior_control ONCE, before cmd_arm/cmd_nod's whole
    cycle loop starts - not per direction/per cycle. See the 2026-07-22 fix
    comment above run_move_axis_no_control for why."""
    if not live:
        print(f"{clock.prefix()}PHCAL DRY: would assume_behavior_control serial={serial} priority=high")
        return {"ok": True, "mode": "dry"}
    print(f"{clock.prefix()}PHCAL assume_behavior_control serial={serial} priority=high")
    assume = mod.wirepod_web_send_form(
        "/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10
    )
    print(f"{clock.prefix()}PHCAL assume settle {PHCAL_ASSUME_SETTLE_SECONDS}s")
    time.sleep(PHCAL_ASSUME_SETTLE_SECONDS)
    return {"ok": assume.get("status") == 200, "assume": assume}


def run_release_control(mod, clock, serial, live):
    """Fires release_behavior_control ONCE, after cmd_arm/cmd_nod's whole
    cycle loop has finished - not per direction/per cycle. Called from a
    finally block by the caller, so a mid-sequence failure still releases
    cleanly, matching run_named_movement_sequence's own try/finally shape."""
    if not live:
        print(f"{clock.prefix()}PHCAL DRY: would release_behavior_control serial={serial}")
        return {"ok": True, "mode": "dry"}
    print(f"{clock.prefix()}PHCAL release_behavior_control serial={serial}")
    release = mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)
    return {"ok": release.get("status") == 200, "release": release}


def run_preflight(mod, clock, serial, live):
    """conn_test + settle, before any move command fires - see
    PHCAL_PREFLIGHT_SETTLE_SECONDS above for why this exists and why it's
    ported call-for-call from run_wake_both rather than invented. Wrapped in
    try/except the same way run_wake_both is: a hang/timeout/refused
    connection here is exactly what this handler exists to catch, not let
    propagate as a raw traceback."""
    if not live:
        print(f"{clock.prefix()}PHCAL_PREFLIGHT DRY: would conn_test serial={serial}")
        return {"ok": True, "mode": "dry"}
    print(f"{clock.prefix()}PHCAL_PREFLIGHT conn_test serial={serial}")
    try:
        result = mod.wirepod_web_send_form("/api-sdk/conn_test", {"serial": serial}, timeout=10)
    except Exception as exc:  # noqa: BLE001 - same catch shape as run_wake_both
        print(f"{clock.prefix()}PHCAL_PREFLIGHT_FAILED serial={serial} - {exc!r}")
        return {"ok": False, "error": repr(exc)}
    time.sleep(PHCAL_PREFLIGHT_SETTLE_SECONDS)
    ok = result.get("status") == 200
    if ok:
        print(f"{clock.prefix()}PHCAL_PREFLIGHT_OK serial={serial}")
    else:
        print(f"{clock.prefix()}PHCAL_PREFLIGHT_FAILED serial={serial} - {PHCAL_PREFLIGHT_RECOVERY_MESSAGE}")
    return {"ok": ok, "conn_test": result}


def cmd_arm(mod, clock, serial, robot, live, cycles, hold_seconds, speed, pre_assumed=False):
    print(
        f"{clock.prefix()}PHCAL_ARM robot={robot} serial={serial} cycles={cycles} "
        f"hold_seconds={hold_seconds} speed={speed}"
    )
    if pre_assumed:
        print(f"{clock.prefix()}PHCAL_ARM pre_assumed=True - skipping own assume (brobots_wake already holds control)")
        ok = True
    else:
        assume = run_assume_control(mod, clock, serial, live)
        ok = assume["ok"]
    try:
        for i in range(cycles):
            up = run_move_axis_no_control(mod, clock, serial, "move_lift", speed, hold_seconds, live, "up")
            down = run_move_axis_no_control(mod, clock, serial, "move_lift", -speed, hold_seconds, live, "down")
            ok = ok and up["ok"] and down["ok"]
            if i < cycles - 1:
                print(f"{clock.prefix()}PHCAL_ARM settle {PHCAL_CYCLE_SETTLE_SECONDS}s before next cycle")
                if live:
                    time.sleep(PHCAL_CYCLE_SETTLE_SECONDS)
    finally:
        release = run_release_control(mod, clock, serial, live)
        ok = ok and release["ok"]
    print(f"{clock.prefix()}PHCAL_ARM_COMPLETE ok={ok}")
    return 0 if ok else 1


def cmd_nod(mod, clock, serial, robot, live, count, hold_seconds, speed, pre_assumed=False):
    print(
        f"{clock.prefix()}PHCAL_NOD robot={robot} serial={serial} count={count} "
        f"hold_seconds={hold_seconds} speed={speed}"
    )
    if pre_assumed:
        print(f"{clock.prefix()}PHCAL_NOD pre_assumed=True - skipping own assume (brobots_wake already holds control)")
        ok = True
    else:
        assume = run_assume_control(mod, clock, serial, live)
        ok = assume["ok"]
    try:
        for i in range(count):
            print(f"{clock.prefix()}PHCAL_NOD cycle={i + 1}/{count}")
            down = run_move_axis_no_control(mod, clock, serial, "move_head", -speed, hold_seconds, live, "down")
            up = run_move_axis_no_control(mod, clock, serial, "move_head", speed, hold_seconds, live, "up")
            ok = ok and down["ok"] and up["ok"]
            if i < count - 1:
                print(f"{clock.prefix()}PHCAL_NOD settle {PHCAL_CYCLE_SETTLE_SECONDS}s before next cycle")
                if live:
                    time.sleep(PHCAL_CYCLE_SETTLE_SECONDS)
    finally:
        release = run_release_control(mod, clock, serial, live)
        ok = ok and release["ok"]
    print(f"{clock.prefix()}PHCAL_NOD_COMPLETE ok={ok}")
    return 0 if ok else 1


_VOLUME_UI_TO_REAL = {1: 20, 2: 40, 3: 60, 4: 80, 5: 100}


def _map_volume_ui_to_real(volume_ui):
    return _VOLUME_UI_TO_REAL[volume_ui]


def cmd_rattle(mod, clock, serial, robot, live, volume_ui):
    """Fires the Bingo sidecar's rattle sound on one robot via the standalone
    direct-SDK binary (direct_sdk_bingo_rattle_001), ported call-for-call
    from run_songs_runner_001.py's own run_rattle(): Wire-Pod releases
    the robot -> settle past Wire-Pod's own release-polling window
    (DIRECT_SDK_RELEASE_SETTLE_SECONDS) -> the binary connects independently,
    requests BehaviorControl, plays the rattle WAV, releases, disconnects.
    Confirmed against the Go source: no separate hold/settle step exists
    anywhere in that flow - release fires immediately after playback
    finishes - so there is nothing to add here for "hold." volume_ui is
    phcal's own 1-5 scale; mapped to the real 1-100 AudioVolume range before
    being passed as the binary's optional 3rd argument."""
    volume = _map_volume_ui_to_real(volume_ui)
    print(
        f"{clock.prefix()}PHCAL_RATTLE robot={robot} serial={serial} "
        f"volume_ui={volume_ui} volume={volume}"
    )
    release = run_release_control(mod, clock, serial, live)
    ok = release["ok"]

    if not live:
        print(
            f"{clock.prefix()}PHCAL_RATTLE DRY: would settle "
            f"{PHCAL_RATTLE_SETTLE_SECONDS}s then fire rattle binary "
            f"serial={serial} wav={RATTLE_WAV_PATH} volume={volume}"
        )
        print(f"{clock.prefix()}PHCAL_RATTLE_COMPLETE ok={ok} mode=dry")
        return 0 if ok else 1

    print(
        f"{clock.prefix()}PHCAL_RATTLE settle {PHCAL_RATTLE_SETTLE_SECONDS}s "
        "(rattle's own audio-streaming settle margin, wider than the shared "
        "DIRECT_SDK_RELEASE_SETTLE_SECONDS SayText uses - see the constant's "
        "own comment above)"
    )
    time.sleep(PHCAL_RATTLE_SETTLE_SECONDS)

    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    try:
        proc = subprocess.run(
            [str(RATTLE_BIN_PATH), serial, str(RATTLE_WAV_PATH), str(volume)],
            capture_output=True, text=True, timeout=30, env=env,
        )
    except subprocess.TimeoutExpired:
        print(f"{clock.prefix()}PHCAL_RATTLE_DIRECT serial={serial} TIMED OUT (30s)")
        return 1

    combined = "\n".join(
        line for line in (proc.stdout + proc.stderr).splitlines() if "guid" not in line.lower()
    )
    fired_ok = proc.returncode == 0 and "status=OK" in combined
    print(f"{clock.prefix()}{combined}")
    ok = ok and fired_ok
    print(f"{clock.prefix()}PHCAL_RATTLE_COMPLETE ok={ok}")
    return 0 if ok else 1


def cmd_danger(mod, clock, serial, robot, live, volume_ui):
    """Fires GOPOD's own playSound sound (danger-will-robinson.wav) on one
    robot via the SAME standalone direct-SDK binary cmd_rattle uses above
    (RATTLE_BIN_PATH - never rattle-specific, just always called with
    rattle's own WAV path until now) - identical mechanism, different WAV
    path. Added 2026-08-12 so a song can fire this sound directly rather
    than only through the LLM-gated {{playSound||...}} chat path (kgsim_cmds.go's
    DoPlaySound) - see DOPLAYSOUND_REAL_IMPLEMENTATION_AND_RATTLE_FIT_001.md.
    Same settle margin as rattle (PHCAL_RATTLE_SETTLE_SECONDS - the audio-
    streaming settle mechanism isn't asset-specific, just reused by name).
    volume_ui is phcal's own 1-5 scale, same mapping as rattle."""
    volume = _map_volume_ui_to_real(volume_ui)
    print(
        f"{clock.prefix()}PHCAL_DANGER robot={robot} serial={serial} "
        f"volume_ui={volume_ui} volume={volume}"
    )
    release = run_release_control(mod, clock, serial, live)
    ok = release["ok"]

    if not live:
        print(
            f"{clock.prefix()}PHCAL_DANGER DRY: would settle "
            f"{PHCAL_RATTLE_SETTLE_SECONDS}s then fire danger binary "
            f"serial={serial} wav={DANGER_WAV_PATH} volume={volume}"
        )
        print(f"{clock.prefix()}PHCAL_DANGER_COMPLETE ok={ok} mode=dry")
        return 0 if ok else 1

    print(
        f"{clock.prefix()}PHCAL_DANGER settle {PHCAL_RATTLE_SETTLE_SECONDS}s "
        "(same audio-streaming settle margin rattle uses)"
    )
    time.sleep(PHCAL_RATTLE_SETTLE_SECONDS)

    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    try:
        proc = subprocess.run(
            [str(RATTLE_BIN_PATH), serial, str(DANGER_WAV_PATH), str(volume)],
            capture_output=True, text=True, timeout=30, env=env,
        )
    except subprocess.TimeoutExpired:
        print(f"{clock.prefix()}PHCAL_DANGER_DIRECT serial={serial} TIMED OUT (30s)")
        return 1
    except FileNotFoundError:
        print(
            f"{clock.prefix()}PHCAL_DANGER_BLOCKED binary not found at {RATTLE_BIN_PATH} - "
            "build it first: cd $(dirname " + str(RATTLE_BIN_PATH) + ") && go build -o "
            f"{RATTLE_BIN_PATH.name} {RATTLE_BIN_PATH.name}.go"
        )
        return 1

    combined = "\n".join(
        line for line in (proc.stdout + proc.stderr).splitlines() if "guid" not in line.lower()
    )
    fired_ok = proc.returncode == 0 and "status=OK" in combined
    print(f"{clock.prefix()}{combined}")
    ok = ok and fired_ok
    print(f"{clock.prefix()}PHCAL_DANGER_COMPLETE ok={ok}")
    return 0 if ok else 1


def cmd_cube(mod, clock, serial, robot, live):
    """Fires the cube connect->all-corners-red->hold->all-corners-green->
    release blip via the standalone direct-SDK binary
    (direct_sdk_cube_blip_001) - connects independently, requests
    BehaviorControl, connects the cube (ConnectCube), sets all four corner
    LEDs red then green (2s hold between, baked into the binary's own
    blipHoldDuration constant), releases, disconnects. No volume/settle
    concept (unlike rattle/danger) - nothing audio-streamed here, so no
    audio-streaming settle margin applies. Net-new 2026-08-15, the first
    GOPOD code to touch the cube instrument - see
    CUBE_DOOR_SURVEY_001.md/CUBE_BLIP_TOOL_BUILT_001.md. Cube keeper is Brobot 2
    (robot=2, ESN02) - this primitive doesn't hardcode that (same as
    rattle/danger not hardcoding a robot), it just fails cleanly if fired at
    a robot with no cube paired (ConnectCube's own Success=false path).
    WIREPOD_HOME must be set so vector.NewWP() resolves
    chipper/jdocs/botSdkInfo.json as an absolute path rather than relative
    to whatever cwd this process happens to run from - the exact bug a raw
    unwired run hit (CUBE_BLIP_ALIAS_PHCAL_WIRED_001.md); same fix every
    other direct-SDK caller in this family already applies (rattle/danger's
    own env=dict(...) above, robot-sleep/-wake/-info's own inline
    WIREPOD_HOME= prefix in brobots.sh)."""
    print(f"{clock.prefix()}PHCAL_CUBE robot={robot} serial={serial}")
    release = run_release_control(mod, clock, serial, live)
    ok = release["ok"]

    if not live:
        print(f"{clock.prefix()}PHCAL_CUBE DRY: would fire cube blip binary serial={serial}")
        print(f"{clock.prefix()}PHCAL_CUBE_COMPLETE ok={ok} mode=dry")
        return 0 if ok else 1

    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    try:
        proc = subprocess.run(
            [str(CUBE_BLIP_BIN_PATH), serial],
            capture_output=True, text=True, timeout=30, env=env,
        )
    except subprocess.TimeoutExpired:
        print(f"{clock.prefix()}PHCAL_CUBE_DIRECT serial={serial} TIMED OUT (30s)")
        return 1

    combined = "\n".join(
        line for line in (proc.stdout + proc.stderr).splitlines() if "guid" not in line.lower()
    )
    fired_ok = proc.returncode == 0 and "status=OK" in combined
    print(f"{clock.prefix()}{combined}")
    ok = ok and fired_ok
    print(f"{clock.prefix()}PHCAL_CUBE_COMPLETE ok={ok}")
    return 0 if ok else 1


def _sleep_wake_specs(mod, which):
    """which is '1'/'2'/'both' - same which->serial mapping
    brobots.sh's own _robot_sleep_specs() shell function already applies
    (that one also accepts '0' as a synonym for 'both'; phcal's own guided
    prompt only ever offers '1'/'2'/'both', so '0' is not reproduced here -
    nothing is lost, just one spelling not exposed at this entry point)."""
    specs = []
    if which in ("1", "both"):
        specs.append(f"Brobot_1:{BROBOT_1_SERIAL}")
    if which in ("2", "both"):
        specs.append(f"Brobot_2:{BROBOT_2_SERIAL}")
    return specs


def cmd_brobots_ready(mod, clock, live, phrase):
    """Fires the synchronized dual-robot phrase via the same
    direct_sdk_brobots_ready_001 binary run_golden_song_001.py's own
    run_brobots_ready_together() and core.sh's
    _gopod_chord_direct_together_job() already call - always both robots,
    fixed pair, no per-call robot choice (matches the catalog's own
    ROBOT-ASSIGNABLE=NO finding for this mechanism). No dry mode, same
    reasoning as cmd_sleep/cmd_wake."""
    print(f"{clock.prefix()}PHCAL_BROBOTS_READY phrase={phrase!r}")
    if not live:
        print(f"{clock.prefix()}PHCAL_NO_DRY_MODE brobots_ready has no simulate path (no --dry exists in the binary) - re-run with GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 to fire it")
        return 1
    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    try:
        proc = subprocess.run(
            [str(TOGETHER_BIN_PATH), phrase, f"Brobot 1:{BROBOT_1_SERIAL}", f"Brobot 2:{BROBOT_2_SERIAL}"],
            capture_output=True, text=True, timeout=30, env=env,
        )
    except subprocess.TimeoutExpired:
        print(f"{clock.prefix()}PHCAL_BROBOTS_READY_DIRECT TIMED OUT")
        return 1
    combined = "\n".join(
        line for line in (proc.stdout + proc.stderr).splitlines() if "guid" not in line.lower()
    )
    ok = proc.returncode == 0 and combined.count("status=OK") == 2
    print(f"{clock.prefix()}{combined}")
    print(f"{clock.prefix()}PHCAL_BROBOTS_READY_COMPLETE ok={ok}")
    return 0 if ok else 1


def cmd_brobots_ready_single(clock, which, label, live, phrase):
    """2026-08-18, PHCAL_DETECT_FIRST_001.md: single-mode degrade of
    cmd_brobots_ready's own phrase. The two-robot direct-SDK binary above
    (direct_sdk_brobots_ready_001.go) structurally CANNOT run with one
    robot without a Go rebuild - its own main() hard-requires exactly two
    <label>:<serial> specs (`if len(os.Args) < 4`, `specs :=
    os.Args[2:4]`), confirmed by reading the source directly, not guessed -
    out of scope for this pass. Reuses run_robot_control_song_001.py's own
    run_single_note()'s new "say_phrase" note instead (added this same
    pass) - the exact same say_line()/Wire-Pod-REST-assume-release shape
    the "weather" primitive's own single-robot speak already uses, not a
    new speech channel invented here. The two-robot version already speaks
    the SAME single phrase on both robots at once (not two different
    half-lines), so one robot speaking that same phrase alone is a
    complete, faithful single-robot version - not a truncated half of a
    duet."""
    print(f"{clock.prefix()}PHCAL_BROBOTS_READY_SINGLE robot={which} ({label}) phrase={phrase!r}")
    control_mod = _load_module(CONTROL_SONG_RUNNER_PATH, "run_robot_control_song_001")
    result = control_mod.run_single_note("say_phrase", live, which, phrase=phrase)
    ok = bool(result.get("ok"))
    print(f"{clock.prefix()}PHCAL_BROBOTS_READY_SINGLE_COMPLETE ok={ok}")
    return 0 if ok else 1


def cmd_robot_info(mod, clock, which):
    """Read-only VersionState/ProtocolVersion/BatteryState snapshot, direct-
    SDK, no BehaviorControl needed at all - the diagnostic golden note built
    2026-08-10 for the Brobot 2 (Pip) intermittent-unresponsiveness
    investigation (BROBOT_2_INSTABILITY_EXTERNAL_AI_BRIEF_001.md /
    WHEEL_NUDGE_COLD_FIRE_ROOT_CAUSE_SURVEY_001.md's own follow-up). Same
    binary brobots.sh's own robot-info alias already calls
    (direct_sdk_robot_info_001 - ROBOT_SLEEP_DIRECT_SDK_BUILT_001.md's
    sibling tool, reused verbatim, not a second implementation). which is
    "1"/"2"/"both", same convention/resolver (_sleep_wake_specs) sleep_wake
    already uses. No dry mode - the binary itself has none (same shape as
    cmd_brobots_ready/sleep_wake above); this call is harmless regardless
    (never assumes BehaviorControl, never moves anything, read-only), so
    there is nothing a dry preview would meaningfully simulate. Live-gated
    the same as every other primitive that actually opens a connection to
    the real robot. (2026-08-20, PHCAL_NAV_FIXES_001.md: corrected - this
    used to claim "only weather (zero robot contact at all) skips this,"
    but weather DOES make robot contact, via run_single_note()'s own
    internal connect, not the explicit restart_preflight()/run_preflight()
    pair this function and most others call directly - see weather's own
    dispatch branch and _submenu_control_note() for the real
    classification.)"""
    specs = _sleep_wake_specs(mod, which)
    print(f"{clock.prefix()}PHCAL_ROBOT_INFO which={which} specs={' '.join(specs)}")
    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    try:
        proc = subprocess.run(
            [str(ROBOT_INFO_BIN_PATH), *specs],
            capture_output=True, text=True, timeout=ROBOT_INFO_TIMEOUT_SECONDS, env=env,
        )
    except subprocess.TimeoutExpired:
        print(f"{clock.prefix()}PHCAL_ROBOT_INFO_TIMED_OUT after {ROBOT_INFO_TIMEOUT_SECONDS}s")
        return 1
    combined = "\n".join(
        line for line in (proc.stdout + proc.stderr).splitlines() if "guid" not in line.lower()
    )
    ok = proc.returncode == 0
    print(f"{clock.prefix()}{combined}")
    print(f"{clock.prefix()}PHCAL_ROBOT_INFO_COMPLETE ok={ok}")
    return 0 if ok else 1


BATTERY_MIN_VOLTS = 3.7

PHCAL_PRIMITIVE_BATTERY_BLOCK_ENABLED = False


def run_battery_check(mod, clock, serial, live):
    """Reuses the exact same direct_sdk_robot_info_001 binary robot_info's
    own menu item calls (cmd_robot_info above) - one serial at a time, no
    BehaviorControl, read-only. BLOCKS (returns ok=False, does not just
    warn) when battery_volts reads below BATTERY_MIN_VOLTS or can't be
    read at all - a failed read gets treated the same as a low reading,
    not waved through, since "couldn't confirm the robot is safe to drive"
    is exactly the same risk as "confirmed it isn't." Per operator
    direction 2026-08-10: block, don't just warn."""
    if not live:
        print(f"{clock.prefix()}PHCAL_BATTERY_CHECK DRY: would read battery_volts for serial={serial}")
        return {"ok": True, "mode": "dry"}
    label = "Brobot_1" if serial == BROBOT_1_SERIAL else "Brobot_2"
    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    try:
        proc = subprocess.run(
            [str(ROBOT_INFO_BIN_PATH), f"{label}:{serial}"],
            capture_output=True, text=True, timeout=ROBOT_INFO_TIMEOUT_SECONDS, env=env,
        )
    except subprocess.TimeoutExpired:
        print(f"{clock.prefix()}PHCAL_BATTERY_CHECK_TIMED_OUT after {ROBOT_INFO_TIMEOUT_SECONDS}s serial={serial}")
        return {"ok": False, "volts": None, "error": "timeout"}
    match = re.search(r"battery_volts=([\d.]+)", proc.stdout)
    if proc.returncode != 0 or not match:
        print(f"{clock.prefix()}PHCAL_BATTERY_CHECK_FAILED serial={serial} rc={proc.returncode} - could not read battery_volts")
        return {"ok": False, "volts": None, "error": "no_reading"}
    volts = float(match.group(1))
    ok = volts >= BATTERY_MIN_VOLTS
    tag = "PHCAL_BATTERY_CHECK_OK" if ok else "PHCAL_BATTERY_CHECK_BLOCKED"
    print(f"{clock.prefix()}{tag} serial={serial} volts={volts:.3f} min={BATTERY_MIN_VOLTS}")
    return {"ok": ok, "volts": volts}


_SESSION_MODE = None
_PRESENT_ROBOTS = []

_DETECTED_PRESENT_ROBOTS = []


def _candidate_list():
    """The list detect-first probes. Reads phcal_config.json's own
    robots.candidates directly (N-capable - see the demo config's own
    comment - but this repo has only ever tested exactly 2). If that list
    is empty (a fresh clone with no config yet, or an older-shape config
    missing the key entirely) falls back to synthesizing the same two
    candidates _resolve_brobot_serials() already falls back to
    (BROBOT_1_SERIAL/BROBOT_2_SERIAL, which themselves already resolve
    through env-var-then-hardcoded) - so a fresh clone with no config still
    probes the exact same two ESNs phcal has always assumed present, just
    genuinely probed now instead of assumed. `which` here (not present in
    phcal_config.json's own candidate dicts) is this file's own "1"/"2"
    identity, derived purely from list position - index 0 is Brobot 1's
    slot, index 1 is Brobot 2's slot (same convention the demo config's own
    comment states: "First two entries map to Brobot 1 / Brobot 2 in that
    order"). A 3rd+ candidate (untested) gets which=None - this file's
    _prompt_robot()/_sleep_wake_specs() contract is binary "1"/"2"/"both"
    only, unchanged by this pass; a genuine 3rd-robot identity would need
    deeper work, out of scope here."""
    candidates = _PHCAL_CONFIG.get("robots", {}).get("candidates", [])
    if not candidates:
        candidates = [
            {"label": "Brobot 1", "esn": BROBOT_1_SERIAL},
            {"label": "Brobot 2", "esn": BROBOT_2_SERIAL},
        ]
    out = []
    for i, c in enumerate(candidates):
        which = "1" if i == 0 else "2" if i == 1 else None
        out.append({"label": c["label"], "esn": c["esn"], "which": which})
    return out


def _probe_candidate(clock, candidate):
    """One candidate -> candidate dict + present/volts. Same binary call
    cmd_robot_info/run_battery_check already make - read-only, no
    BehaviorControl. A timed-out or unparseable read is treated as NOT
    present (same "couldn't confirm == not safe" posture run_battery_check
    already applies to its own low-battery gate, reused here for presence
    too) - never an ambiguous maybe.

    2026-08-18 fix, live-tested by the operator: this used to share
    ROBOT_INFO_TIMEOUT_SECONDS (30s) with the deliberate, occasional "1.
    info" pick - fine when a human explicitly asked and knows to wait, but
    this probe now fires automatically on EVERY phcal launch, and gave zero
    output while waiting - a genuinely absent robot read as a silent freeze
    for up to 30 real seconds (confirmed live: two Ctrl-C's, both while
    still waiting on an absent second candidate). Every PRESENT robot in
    the operator's own live testing responded in well under 1 second. Now:
    (1) prints a "probing..." line immediately, before the subprocess call,
    so waiting is visible instead of silent; (2) uses its own short,
    separate timeout (PHCAL_DETECT_PROBE_TIMEOUT_SECONDS, 5s - generous vs.
    the ~0.5-0.6s real observed latency, nowhere near the deliberate-check
    ceiling) so an absent robot resolves quickly on every startup instead
    of eating up to 30s per absent candidate. ROBOT_INFO_TIMEOUT_SECONDS
    itself is untouched - "1. info"/battery-check still get their original
    30s, unchanged."""
    label, esn = candidate["label"], candidate["esn"]
    print(f"{clock.prefix()}PHCAL_DETECT_PROBE label={label} esn={esn} probing...")
    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    try:
        proc = subprocess.run(
            [str(ROBOT_INFO_BIN_PATH), f"{label}:{esn}"],
            capture_output=True, text=True, timeout=PHCAL_DETECT_PROBE_TIMEOUT_SECONDS, env=env,
        )
    except subprocess.TimeoutExpired:
        print(f"{clock.prefix()}PHCAL_DETECT_PROBE label={label} esn={esn} TIMED_OUT after {PHCAL_DETECT_PROBE_TIMEOUT_SECONDS}s - treating as not present")
        return {**candidate, "present": False, "volts": None}
    match = re.search(r"battery_volts=([\d.]+)", proc.stdout)
    if proc.returncode != 0 or not match:
        print(f"{clock.prefix()}PHCAL_DETECT_PROBE label={label} esn={esn} NOT_PRESENT rc={proc.returncode}")
        return {**candidate, "present": False, "volts": None}
    volts = float(match.group(1))
    print(f"{clock.prefix()}PHCAL_DETECT_PROBE label={label} esn={esn} PRESENT volts={volts:.3f}")
    print(f"{clock.prefix()}{label} has sufficient power PRESENT {volts:.3f} volts")
    return {**candidate, "present": True, "volts": volts}


def _detect_present_robots(clock):
    """Runs once, at the very top of run_guided_flow(), before the menu
    draws. Returns (present_robots, session_mode): present_robots is the
    subset of _candidate_list() that responded, in candidate-list order;
    session_mode is "none" (0 present), "single" (1 present), or "multi"
    (2+ present - N-capable, tested at exactly 2).

    2026-08-26, PHCAL_ARROW_NAV_BUILD_PLAN_006.md LOOK PASS: wrapped in the
    shared `_NAV_FRAME_MARK` open/close per the operator's own PDF -
    "framing wraps everything, including diagnostics." Every real
    `PHCAL_DETECT_PROBE`/`PHCAL_DETECT_FIRST` diagnostic line is kept
    exactly as before; a plain-language line is added alongside each one,
    never replacing it.

    2026-08-26, PHCAL_LOOK_PASS_CORRECTION_001.md: per the operator's own
    literal PDF spec, this screen's `**` frame contains FOUR distinct
    logical chunks (the probing-intro line, each candidate's own probe
    block, the final present/mode summary), each separated by its own
    divider - not one bare frame with no internal structure. A divider now
    prints immediately after the frame opens, between each chunk, and
    once more right before the frame closes - the general `**` -> divider
    -> content -> divider (repeat) -> divider -> `**` shape every other
    screen in this file now follows too."""
    _frame_open()
    print(_NAV_DIVIDER)
    candidates = _candidate_list()
    print(f"{clock.prefix()}PHCAL_DETECT_FIRST probing {len(candidates)} configured candidate(s)...")
    results = []
    for c in candidates:
        print(_NAV_DIVIDER)
        results.append(_probe_candidate(clock, c))
    present = [r for r in results if r["present"]]
    if len(present) == 0:
        mode = "none"
    elif len(present) == 1:
        mode = "single"
    else:
        mode = "multi"
    print(_NAV_DIVIDER)
    print(f"{clock.prefix()}PHCAL_DETECT_FIRST present={[p['label'] for p in present]} mode={mode}")
    print(f"{clock.prefix()}detected brobots-mode {mode.upper()}")
    print(_NAV_DIVIDER)
    _frame_close()
    return present, mode


def _confirm_multi_mode(clock, present_robots, detected_mode):
    """2026-08-18, operator request after live-testing multi mode: an
    explicit confirm/override step. Returns (present_robots, session_mode)
    in the exact shape _detect_present_robots() returns, so
    run_guided_flow() and every mode-aware reader below treat an override
    identically to a real single/none/multi detection - no separate code
    path needed anywhere else in this file.

    History of this screen's own shape, kept for record (each pass
    superseded the last): PHCAL_MULTI_MODE_SCREEN_RESHAPE_001.md's locked
    base-rules shape (0 exit + 1..N single + * multi + a none hint line) ->
    PHCAL_LESS_FLAKY_SWEEP_001.md (dropped the none-row's own typed key) ->
    PHCAL_DOWN_ARROW_HOTKEY_001.md (down-arrow hotkey for none) ->
    PHCAL_SLASH_HOTKEY_001.md ("/" keypress hotkey for none, resolving
    PHCAL_FULL_ARROW_NAV_SURVEY_001.md's §0 ↓-collision as Option A) ->
    PHCAL_ARROW_NAV_BUILD_PLAN_002.md Phase 1, 2026-08-22 ("0" row and its
    exit-on-choice check removed; "/" became an ordinary arrow-reachable
    row, no more keypress interception).

    2026-08-23, direct operator build request: this screen now runs for
    EVERY detected mode (multi/single/none), not only multi - the old
    "only when detect-first resolves multi" gate is gone (call site in
    _resolve_session_mode_once() below calls this unconditionally now).
    Header is a fixed 3-line block naming the detected mode; a bare ENTER
    with no arrow pressed accepts that detected mode as-is (the happy
    path, no row needs to exist for it - "accept default" is a screen
    STATE, tracked as highlight == -1, not a printed row). Arrow-down (or
    up) enters the row list starting at index 0; only once a row is
    highlighted does Enter resolve to that row's own choice instead of the
    default. Rows never show their own key/prefix ("*"/"/"/"1.") anymore -
    plain labels only, per this pass.

    2026-08-23, briefly changed to build rows from the FULL configured
    candidate list (_candidate_list()) instead of the live present list -
    REVERTED 2026-08-24, direct operator build request, after that change
    itself caused a real, confirmed-live regression: with both robots
    powered off (present=[], mode=none), the picker still offered
    "brobots-single mode on Brobot 1/2" - rows for robots detect-first had
    just confirmed were NOT there. Single-mode rows are single-mode rows
    ONLY for robots the live detection actually found present, full stop -
    `present_robots` (this function's own second argument, exactly what
    _detect_present_robots() returned) is the row source again, not
    _candidate_list(). Concretely: none -> the dry-run row alone, nothing
    else (present_robots is empty); single/multi -> one row per PRESENT
    robot, excluding the one already the bare-ENTER default in `single`
    mode (unchanged from the original 2026-08-23 shape - this only
    reverts the one-day candidate-list detour). The downstream `match is
    None` synthesis and _no_confirmed_robot_this_session()'s
    forced-but-absent branch (both added same day as the now-reverted
    candidate-list change, to handle exactly the choice this reversion
    removes) are left in place, not pruned - `choice` can no longer
    resolve to a candidate outside `present_robots`, so that code is
    effectively unreachable now, flagged here rather than silently
    dropped, per this pass's own row-source-only scope.

    2026-08-25, PHCAL_ARROW_NAV_BUILD_PLAN_006.md Lane 1a: this screen's
    own bespoke raw-mode loop (its own `_redraw()`, its own `with
    _raw_mode(...)` block, its own two-state ESC handling) is REMOVED -
    folded into arrow_column_pick() itself via that function's new
    `esc_home` opt-in (see its own docstring), the one holdout
    PHCAL_ARROW_NAV_BUILD_PLAN_006.md's own survey found not routed
    through the shared cyclable-choice engine. The old highlight==-1 "no
    row marked, accept detected mode" state is now a REAL row at index 0
    (`"_accept"`) - a genuine, if cosmetic, change: that row now always
    shows its own "> " marker from the first draw, matching how every
    other "ENTER for X, or change:" leaf prompt in this file already
    marks its own pre-highlighted default. The header text is restyled to
    match that same convention instead of describing a state that no
    longer exists ("Press ENTER now for default... or Arrow down to
    select" -> "ENTER for <accept-default label>, or change:"). Return
    contract, row source, and every downstream branch below
    (choice == "/" / the `match is None` synthesis) are byte-unchanged."""
    mode_upper = detected_mode.upper()
    accept_label = f"ENTER for {mode_upper} brobots-mode"

    detected_which = present_robots[0]["which"] if detected_mode == "single" and present_robots else None
    rows = [("/", "brobots-none mode (dry-runs)")]
    rows += [
        (p["which"], f"brobots-single mode on {p['label']}")
        for p in present_robots
        if p["which"] != detected_which
    ]
    base_options = [("_accept", accept_label)] + rows
    options = [(k, f"{i:02d}. {label}") for i, (k, label) in enumerate(base_options)]

    _frame_open()
    print(_NAV_DIVIDER)
    print(_NAV_HINT_ARROW)
    print(_NAV_DIVIDER)
    choice = arrow_column_pick(options, highlight=0, show_key=False, esc_home=0)

    if choice == "_accept":
        _frame_close()
        return present_robots, detected_mode
    if choice == "/":
        print(f"{clock.prefix()}PHCAL_MODE_OVERRIDE mode=none (operator chose dry-run over detected {detected_mode})")
        _frame_close()
        return [], "none"
    match = next((p for p in present_robots if p["which"] == choice), None)
    if match is None:
        match = {**next(c for c in _candidate_list() if c["which"] == choice), "present": False, "volts": None}
    print(f"{clock.prefix()}PHCAL_MODE_OVERRIDE mode=single robot={match['which']} ({match['label']}) (operator chose single over detected {detected_mode})")
    _frame_close()
    return [match], "single"


def _low_voltage_gate(clock, present_robots):
    """After detect-first probes: if ANY present robot's battery reads
    below BATTERY_MIN_VOLTS (the same threshold run_battery_check() already
    gates individual motor calls on - reused, not a new number), warn
    plainly (which robot, its actual level) and ask before proceeding.
    Routed through the same _prompt_pick() single choke-point every other
    y/n CONFIRM in this file already uses (2026-08-21,
    PHCAL_NAV_CONSOLIDATION_001.md - was _prompt_choice(), same wrapper
    swap, unchanged args/behavior). default='n' - the operator must type
    something other than Enter/'n' to continue with a low robot present.
    volts=None on a "present" robot shouldn't happen in practice
    (_probe_candidate already treats an unparseable read as not-present),
    kept here as a defensive belt-and-suspenders low-battery treatment
    rather than assumed safe."""
    low = [r for r in present_robots if r["volts"] is None or r["volts"] < BATTERY_MIN_VOLTS]
    if not low:
        return True
    for r in low:
        volts_str = f"{r['volts']:.3f}" if r["volts"] is not None else "unknown"
        print(f"{clock.prefix()}PHCAL_LOW_BATTERY_WARNING label={r['label']} esn={r['esn']} volts={volts_str} min={BATTERY_MIN_VOLTS}")
    choice = _prompt_pick(
        {"y", "n"}, default="n", labels={"y": "yes", "n": "no"},
        question="one or more present robots read low battery (see warning above) - proceed anyway?",
    )
    return choice == "y"


def _resolve_session_mode_once():
    """2026-08-19, PHCAL_NAV_POLISH_001.md addendum (operator review of the
    held nav-polish build): the detect-first probe, multi-mode confirm, and
    low-voltage gate used to live inside _run_guided_flow_once() itself, so
    every "continue" at the continue-or-exit prompt re-ran the whole ~1-2s
    probe (and could re-ask the multi-mode confirm question) on every loop
    pass. The robots present don't change between menu passes within one
    `phcal` invocation, so this now runs exactly ONCE, called by
    run_guided_flow() before its stay-in loop starts.

    Sets module-level _SESSION_MODE/_PRESENT_ROBOTS exactly as before -
    every downstream reader (_prompt_robot(), the robot_info/
    brobots_announce_in_sync dispatch branches) is unaffected by where the
    probe itself now runs. Returns True if OK to proceed, False if the
    low-voltage gate declined - the caller (run_guided_flow()) treats a
    decline as a blocked first pass so it still reaches the continue-or-exit
    prompt per operator decision 5, rather than dumping the operator
    straight out; it does not re-probe on a later "continue" since this
    function only ever runs once per invocation.

    2026-08-21, PHCAL_NAV_BASE_RULES_FIXES_001.md (PHCAL_NAV_BASE_RULES_
    SURVEY_001.md §5 step 1): GOPOD_ALLOW_LIVE_ROBOT_SPEECH is now decided
    HERE, once, from the resolved mode - not by a separate shell-side
    live_robots_prompt() gate in brobots.sh's phcal() (removed the same
    pass; that function still exists and is still used by phcal()'s
    direct-flag CLI form, untouched). multi/single (the operator chose to
    continue, or detect-first found exactly one robot) sets it to "1";
    none (detected, or explicitly chosen via _confirm_multi_mode's none/
    dry-run row, now offered after every detected mode, not only multi)
    clears it - a hard set/pop either way, so a stray pre-exported value
    from the shell never silently overrides what this screen just
    resolved. Every dispatch branch already reads this fresh via
    `os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"` at fire time, so
    setting it once here, before the guided-flow loop starts, is
    sufficient - no per-primitive change needed."""
    global _SESSION_MODE, _PRESENT_ROBOTS, _DETECTED_PRESENT_ROBOTS
    _detect_clock = _Clock()
    _PRESENT_ROBOTS, _SESSION_MODE = _detect_present_robots(_detect_clock)
    _DETECTED_PRESENT_ROBOTS = _PRESENT_ROBOTS
    _PRESENT_ROBOTS, _SESSION_MODE = _confirm_multi_mode(_detect_clock, _PRESENT_ROBOTS, _SESSION_MODE)
    if not _low_voltage_gate(_detect_clock, _PRESENT_ROBOTS):
        print(f"{_detect_clock.prefix()}PHCAL_GUIDED_EXIT declined to proceed with a low-battery robot present")
        return False
    _frame_flush_close()
    if _SESSION_MODE == "none":
        os.environ.pop("GOPOD_ALLOW_LIVE_ROBOT_SPEECH", None)
        print(f"{_detect_clock.prefix()}PHCAL_SESSION_MODE mode=none - no configured robots responded; session continues, live-firing primitives will report unreachable")
    elif _SESSION_MODE == "single":
        os.environ["GOPOD_ALLOW_LIVE_ROBOT_SPEECH"] = "1"
        print(f"{_detect_clock.prefix()}PHCAL_SESSION_MODE mode=single robot={_PRESENT_ROBOTS[0]['label']}")
    else:
        os.environ["GOPOD_ALLOW_LIVE_ROBOT_SPEECH"] = "1"
        print(f"{_detect_clock.prefix()}PHCAL_SESSION_MODE mode=multi robots={[r['label'] for r in _PRESENT_ROBOTS]}")
    if not _phcal_session_checkin(_detect_clock):
        return False
    return True


_PHCAL_CHECKED_IN = False


def _phcal_session_checkin(clock):
    global _PHCAL_CHECKED_IN
    if _PHCAL_CHECKED_IN or os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") != "1":
        return True
    try:
        mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
        result = mod.wirepod_web_send_form("/api-sdk/gopod_run", {"action": "start", "name": "phcal", "kind": "phcal", "pid": str(os.getpid())}, timeout=10)
        reply = json.loads(result.get("body") or "{}")
    except Exception as exc:  # noqa: BLE001 - no Wire-Pod: phcal goes on
        print(f"{clock.prefix()}PHCAL_SESSION check-in not reached ({exc!r}) - going on")
        return True
    if not reply.get("ok"):
        print(f"{clock.prefix()}PHCAL_SESSION refused: {reply.get('busy')!r} is going - say stop run first")
        return False
    _PHCAL_CHECKED_IN = True
    import atexit
    atexit.register(_phcal_session_checkout)
    print(f"{clock.prefix()}PHCAL_SESSION checked in with Goverlord")
    return True


def _phcal_session_checkout():
    try:
        mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
        mod.wirepod_web_send_form("/api-sdk/gopod_run", {"action": "end", "name": "phcal", "reason": "done"}, timeout=10)
    except Exception:  # noqa: BLE001 - Wire-Pod clears a run whose process is gone anyway
        pass


def cmd_animation(mod, clock, serial, robot, live, animation_token, hold_seconds, pre_assumed=False):
    """Fires one of the three catalog-confirmed animation tokens
    (kgSuccess/searching/answering) standalone, via the same
    {{playAnimationWI||TOKEN}}-over-say_text dispatch
    run_golden_song_001.py's own run_animation_only() uses - ported here
    verbatim (not imported - phcal loads run_section1_full_live_001.py as
    `mod`, not the golden engine module that function lives in).
    searching/answering loop the dispatch every
    _ANIMATION_LOOP_INTERVAL_SECONDS until hold_seconds elapses; kgSuccess
    (and anything else) fires once, then holds. Assume/release wrap the
    whole call once, same shape as every other bench primitive here."""
    loop = animation_token in _LOOP_ANIMATION_TOKENS
    print(
        f"{clock.prefix()}PHCAL_ANIMATION robot={robot} serial={serial} token={animation_token} "
        f"hold_seconds={hold_seconds} loop={loop}"
    )
    getout_token = _ANIMATION_GETOUT_TOKENS.get(animation_token)
    if not live:
        if pre_assumed:
            print(f"{clock.prefix()}PHCAL_ANIMATION DRY: pre_assumed=True - would skip own assume (brobots_wake already holds control)")
        else:
            print(f"{clock.prefix()}PHCAL DRY: would assume_behavior_control serial={serial} priority=high")
        if loop:
            print(
                f"{clock.prefix()}PHCAL DRY: would say_text {{{{playAnimationWI||{animation_token}}}}} "
                f"every {_ANIMATION_LOOP_INTERVAL_SECONDS}s until {hold_seconds}s elapsed"
            )
            if getout_token:
                print(
                    f"{clock.prefix()}PHCAL DRY: would then say_text {{{{playAnimationWI||{getout_token}}}}} "
                    f"once, to end the {animation_token} loop cleanly"
                )
            else:
                print(
                    f"{clock.prefix()}PHCAL DRY: no confirmed getout transition for '{animation_token}' - "
                    f"would settle {_ANIMATION_NO_GETOUT_RELEASE_SETTLE_SECONDS}s (native Wire-Pod stop "
                    "discipline, ANSWERING_NATIVE_STOP_DISCIPLINE_001.md) before releasing, instead of "
                    "releasing immediately"
                )
        else:
            print(
                f"{clock.prefix()}PHCAL DRY: would say_text {{{{playAnimationWI||{animation_token}}}}} "
                f"once, then hold {hold_seconds}s"
            )
        print(f"{clock.prefix()}PHCAL DRY: would release_behavior_control serial={serial}")
        print(f"{clock.prefix()}PHCAL_ANIMATION_COMPLETE ok=True mode=dry")
        return 0

    if pre_assumed:
        print(f"{clock.prefix()}PHCAL_ANIMATION pre_assumed=True - skipping own assume (brobots_wake already holds control)")
        ok = True
    else:
        assume = run_assume_control(mod, clock, serial, live)
        ok = assume["ok"]
    anim_payload = f"{{{{playAnimationWI||{animation_token}}}}}"
    dispatches = []
    try:
        if loop:
            elapsed = 0.0
            while elapsed < hold_seconds:
                dispatches.append(
                    mod.wirepod_web_send_form("/api-sdk/say_text", {"text": anim_payload, "serial": serial}, timeout=10)
                )
                time.sleep(_ANIMATION_LOOP_INTERVAL_SECONDS)
                elapsed += _ANIMATION_LOOP_INTERVAL_SECONDS
            if getout_token:
                getout_payload = f"{{{{playAnimationWI||{getout_token}}}}}"
                print(f"{clock.prefix()}PHCAL_ANIMATION firing getout token={getout_token} (ends the {animation_token} loop cleanly)")
                dispatches.append(
                    mod.wirepod_web_send_form("/api-sdk/say_text", {"text": getout_payload, "serial": serial}, timeout=10)
                )
            else:
                print(
                    f"{clock.prefix()}PHCAL_ANIMATION no confirmed getout for '{animation_token}' - "
                    f"settling {_ANIMATION_NO_GETOUT_RELEASE_SETTLE_SECONDS}s before release (native "
                    "Wire-Pod stop discipline) instead of releasing immediately"
                )
                time.sleep(_ANIMATION_NO_GETOUT_RELEASE_SETTLE_SECONDS)
        else:
            dispatches.append(
                mod.wirepod_web_send_form("/api-sdk/say_text", {"text": anim_payload, "serial": serial}, timeout=10)
            )
            time.sleep(hold_seconds)
        print(f"{clock.prefix()}PHCAL_ANIMATION token={animation_token} fired {len(dispatches)}x")
        ok = ok and bool(dispatches) and all(d.get("status") == 200 for d in dispatches)
    finally:
        release = run_release_control(mod, clock, serial, live)
        ok = ok and release["ok"]
    print(f"{clock.prefix()}PHCAL_ANIMATION_COMPLETE ok={ok}")
    return 0 if ok else 1


def cmd_hold(mod, clock, serial, robot, live, hold_seconds, pre_assumed=False):
    """Assume/release bench test (PHCAL_CANDIDATE_CONTROLS_SURVEY_001.md
    Part 2) - lets the operator watch a robot held at priority=high (the same
    OVERRIDE_BEHAVIORS level every golden-engine note already uses) genuinely
    stay put for hold_seconds, then resume normal behavior the instant it's
    released. Pure composition, no new HTTP shape: reuses
    run_assume_control()/run_release_control() verbatim, the same two
    functions cmd_arm/cmd_nod already call once per whole call."""
    print(f"{clock.prefix()}PHCAL_HOLD robot={robot} serial={serial} hold_seconds={hold_seconds}")
    if pre_assumed:
        print(f"{clock.prefix()}PHCAL_HOLD pre_assumed=True - skipping own assume (brobots_wake already holds control)")
        ok = True
    else:
        assume = run_assume_control(mod, clock, serial, live)
        ok = assume["ok"]
    if not live:
        print(f"{clock.prefix()}PHCAL_HOLD DRY: would hold {hold_seconds}s (watch-for-wander is only meaningful live)")
    else:
        print(f"{clock.prefix()}PHCAL_HOLD holding {hold_seconds}s - watch the robot, it should NOT move/wander")
        remaining = hold_seconds
        while remaining > 0:
            step = min(1.0, remaining)
            time.sleep(step)
            remaining = round(remaining - step, 3)
            if remaining > 0:
                print(f"{clock.prefix()}PHCAL_HOLD {remaining:.1f}s remaining")
            else:
                print(f"{clock.prefix()}PHCAL_HOLD holding complete")
    release = run_release_control(mod, clock, serial, live)
    ok = ok and release["ok"]
    tail = " - release fired, robot should resume normal behavior now" if live else ""
    print(f"{clock.prefix()}PHCAL_HOLD_COMPLETE ok={ok}{tail}")
    return 0 if ok else 1


def cmd_brobots_wake(mod, clock, serial, robot, live):
    """2026-08-09, DECOMPOSED per operator direction. This is the wake
    half: conn_test -> assume -> golden-flag release/settle/re-assume pulse
    -> release. NOT wheel-specific - any future motion primitive that needs
    to prove a control session is genuinely live before acting on it can
    call the same core functions this composes (run_wheel_nudge_wake_check/
    run_assume_control/run_wheel_nudge_wake_release_pulse/
    run_release_control). Standalone-fireable and self-contained - releases
    at the end, doesn't leave the robot stuck holding control. Mechanism
    REPLACED 2026-08-10, see WHEEL_NUDGE_WAKE_RELEASE_SETTLE_SECONDS's own
    comment above for the full why. Builds and prints the standardized
    readiness signal (_readiness_signal above) before returning - REST
    channel's own gate-if-not-ready endpoint, same shape _finish_sleep_
    wake's direct-SDK channel gate returns."""
    print(f"{clock.prefix()}PHCAL_BROBOTS_WAKE robot={robot} serial={serial}")
    wake_check = run_wheel_nudge_wake_check(mod, clock, serial, live)
    assume = run_assume_control(mod, clock, serial, live)
    ok = wake_check["ok"] and assume["ok"]
    try:
        wake_confirm = run_wheel_nudge_wake_release_pulse(mod, clock, serial, live)
        ok = ok and wake_confirm["ok"]
    finally:
        release = run_release_control(mod, clock, serial, live)
        ok = ok and release["ok"]
    if not wake_check["ok"]:
        reason = "conn_test failed - robot unreachable"
    elif not assume["ok"]:
        reason = "assume_behavior_control failed"
    elif not wake_confirm["ok"]:
        reason = "release/settle/reassume golden-flag pulse failed"
    elif not release["ok"]:
        reason = "final release_behavior_control failed"
    else:
        reason = "control session confirmed responsive (golden-flag pulse settled)"
    signal = _readiness_signal(
        ok, reason, "rest_golden_flag",
        detail={"wake_check": wake_check, "assume": assume, "wake_confirm": wake_confirm, "release": release},
    )
    print(f"{clock.prefix()}PHCAL_BROBOTS_WAKE_COMPLETE ok={ok}")
    print(f"{clock.prefix()}PHCAL_READY ready={signal['ready']} reason={signal['reason']!r} channel=rest_golden_flag")
    return 0 if signal["ok"] else 1


def cmd_move_reverse(mod, clock, serial, robot, live, hold_seconds, pre_assumed=False):
    """The move half of the same decomposition (see cmd_brobots_wake above)
    - assumes ITS OWN control (unless pre_assumed=True, see cmd_arm's own
    comment - same convention), fires run_move_wheels_no_control (prime +
    move + hold + stop), releases. Fixed ON-CHARGER reverse pulse only
    (MOVE_REVERSE_LW/RW, the webroot's own confirmed S-key/backward
    values) - forward is a later, separate decision, not built here.
    Standalone-fireable and self-contained, same shape as cmd_arm/cmd_nod."""
    print(f"{clock.prefix()}{MOVE_REVERSE_CAUTION}")
    print(
        f"{clock.prefix()}PHCAL_MOVE_REVERSE robot={robot} serial={serial} "
        f"lw={MOVE_REVERSE_LW} rw={MOVE_REVERSE_RW} hold_seconds={hold_seconds}"
    )
    if pre_assumed:
        print(f"{clock.prefix()}PHCAL_MOVE_REVERSE pre_assumed=True - skipping own assume (brobots_wake already holds control)")
        ok = True
    else:
        assume = run_assume_control(mod, clock, serial, live)
        ok = assume["ok"]
    try:
        move = run_move_wheels_no_control(
            mod, clock, serial, MOVE_REVERSE_LW, MOVE_REVERSE_RW, hold_seconds, live
        )
        ok = ok and move["ok"]
    finally:
        release = run_release_control(mod, clock, serial, live)
        ok = ok and release["ok"]
    print(f"{clock.prefix()}PHCAL_MOVE_REVERSE_COMPLETE ok={ok}")
    return 0 if ok else 1


def _weather_api_config():
    """Reads provider/key/unit from Wire-Pod's own apiConfig.json - read-only,
    same single source of truth CLAUDE.md already names for this file, never
    copied or duplicated here."""
    path = os.path.expanduser("~/wire-pod/chipper/apiConfig.json")
    with open(path) as f:
        cfg = json.load(f)
    return cfg.get("weather", {})


def _weather_geocode_candidates(location):
    """Same comma-grouping order as weather.go's geocodeOpenWeatherMap() -
    ported by hand for this API-only test, not imported (the Go source is
    the live-load-bearing copy; this mirrors it so the test proves the same
    fix the robot actually runs). Keep in sync if that Go function's
    candidate order ever changes."""
    fields = location.replace(",", " ").split()
    if not fields:
        return [location]
    seen = set()
    candidates = []

    def add(s):
        s = s.strip()
        if s and s not in seen:
            seen.add(s)
            candidates.append(s)

    add(location)
    add(" ".join(fields))
    if len(fields) >= 2:
        add(" ".join(fields[:-1]) + ", " + fields[-1])
    if len(fields) >= 3:
        add(", ".join(fields))
    return candidates


def cmd_weather_test(location):
    """phcal option 4 - API-only weather test, added 2026-08-02 to let the
    just-fixed geocode candidate logic be checked against any spoken-style
    phrase in seconds. Touches no robot at all: no behavior control, no
    direct SDK, nothing that can fire hardware - plain HTTP calls to
    OpenWeatherMap, same shape as the ad-hoc check that confirmed the
    original bug (a bare query like "windsor ontario canada" returning zero
    geocode matches) and the fix (trying comma-grouping candidates in
    order)."""
    cfg = _weather_api_config()
    if not cfg.get("enable") or not cfg.get("key"):
        print("PHCAL_WEATHER_BLOCKED weather API not enabled or key not set in apiConfig.json")
        return 1
    if cfg.get("provider") != "openweathermap.org":
        print(f"PHCAL_WEATHER_BLOCKED provider={cfg.get('provider')!r} not supported by this test (openweathermap.org only)")
        return 1
    key = cfg["key"]
    unit = cfg.get("unit") or "F"

    matched = None
    tried = _weather_geocode_candidates(location)
    for candidate in tried:
        geo_url = (
            "http://api.openweathermap.org/geo/1.0/direct?q="
            + urllib.parse.quote(candidate) + "&limit=1&appid=" + key
        )
        try:
            with urllib.request.urlopen(geo_url, timeout=10) as resp:
                result = json.loads(resp.read().decode())
        except Exception as exc:
            print(f"PHCAL_WEATHER_GEO_ERROR candidate={candidate!r} {exc}")
            continue
        if result:
            matched = (candidate, result[0])
            break

    if matched is None:
        print(f"PHCAL_WEATHER_NO_MATCH location={location!r} tried={tried}")
        return 1

    candidate, geo = matched
    lat, lon = geo["lat"], geo["lon"]
    print(
        f"PHCAL_WEATHER_GEOCODE matched_on={candidate!r} name={geo['name']} "
        f"state={geo.get('state')} country={geo['country']} lat={lat} lon={lon}"
    )

    units = "imperial" if unit == "F" else "metric"
    weather_url = (
        f"https://api.openweathermap.org/data/2.5/weather?lat={lat}&lon={lon}"
        f"&units={units}&appid={key}"
    )
    try:
        with urllib.request.urlopen(weather_url, timeout=10) as resp:
            weather = json.loads(resp.read().decode())
    except Exception as exc:
        print(f"PHCAL_WEATHER_FETCH_ERROR {exc}")
        return 1

    temp = weather.get("main", {}).get("temp")
    condition = weather.get("weather", [{}])[0].get("main")
    description = weather.get("weather", [{}])[0].get("description")
    print(
        f"PHCAL_WEATHER_RESULT location={geo['name']} temp={temp}{unit} "
        f"condition={condition} ({description})"
    )
    return 0


def cmd_tempo_calibration():
    """phcal option 12 - tempo calibration, added 2026-08-11 at operator
    direction (fold tempo-set into phcal's menu). This function only owns
    the song/mode/step picking UI, same shape as brobots.sh's own
    `tempo-set` alias (song list -> A/B mode -> value/step -> dry preview ->
    y/n confirm -> --yes write). Every actual read/write is handed off to
    tempo_set_001.py's own cmd_set_global/cmd_set_buffer - loaded and called
    directly, not reimplemented, so there is exactly one place tempo values
    are ever written. No robot, no hardware, no preflight - a song's
    knobs.json is the only thing this touches."""
    tempo_mod = _load_module(TEMPO_SET_PATH, "tempo_set_001")

    song_dirs = sorted(
        d.name for d in SONGS_DIR.iterdir() if d.is_dir() and d.name != "zzz_archives"
    )
    if not song_dirs:
        print(f"PHCAL_TEMPO_BLOCKED no songs found under {SONGS_DIR}")
        return 1
    dir_choices = {str(i) for i in range(1, len(song_dirs) + 1)}
    dir_labels = {str(i): name for i, name in enumerate(song_dirs, start=1)}
    dir_choice = _prompt_pick(dir_choices, exit_on_invalid=True, labels=dir_labels, back=True, question="Pick a song")
    chosen_dir = song_dirs[int(dir_choice) - 1]
    knobs_path = SONGS_DIR / chosen_dir / "knobs.json"
    if not knobs_path.exists():
        print(f"PHCAL_TEMPO_BLOCKED no knobs.json found for {chosen_dir} at {knobs_path}")
        return 1

    mode_choice = _prompt_pick(
        {"a", "b"}, exit_on_invalid=True,
        labels={
            "a": "set the song's GLOBAL tempo (whole-song ease)",
            "b": "set ONE step's tempo_factor (+ optional comment)",
        },
        question="pick a mode", back=True,
    )

    if mode_choice == "a":
        value = input("new global_tempo (0.0-9.9): ").strip()
        _accum_lines(1)
        tempo_mod.cmd_set_global(["set-global", value, "--knobs", str(knobs_path)])
        apply_choice = _prompt_pick(
            {"y", "n"}, default="n", labels={"y": "yes", "n": "no"}, question="apply the above?", back=True,
        )
        if apply_choice == "y":
            return tempo_mod.cmd_set_global(["set-global", value, "--yes", "--knobs", str(knobs_path)])
        return 0

    envelope_mod = _load_module(KNOBS_ENVELOPE_PATH, "knobs_envelope_001")
    _, envelope = envelope_mod.load_knobs_envelope(str(SONGS_DIR / chosen_dir))
    step_ids = [s.get("step_id") for s in envelope.get("steps", [])]
    if not step_ids:
        print(f"PHCAL_TEMPO_BLOCKED no steps found for {chosen_dir}")
        return 1
    step_choices = {str(i) for i in range(1, len(step_ids) + 1)}
    step_labels = {str(i): sid for i, sid in enumerate(step_ids, start=1)}
    step_choice = _prompt_pick(step_choices, exit_on_invalid=True, labels=step_labels, question="pick a step", back=True)
    step_id = step_ids[int(step_choice) - 1]
    factor = input("new tempo_factor (default 1.0): ").strip() or "1.0"
    _accum_lines(1)
    comment = input("tempo_comment (optional, Enter to leave unchanged): ").strip()
    _accum_lines(1)
    comment_flag = ["--comment", comment] if comment else []

    tempo_mod.cmd_set_buffer(["set-buffer", step_id, "--factor", factor, *comment_flag, "--knobs", str(knobs_path)])
    apply_choice = _prompt_pick(
        {"y", "n"}, default="n", labels={"y": "yes", "n": "no"}, question="apply the above?", back=True,
    )
    if apply_choice == "y":
        return tempo_mod.cmd_set_buffer(
            ["set-buffer", step_id, "--factor", factor, *comment_flag, "--yes", "--knobs", str(knobs_path)]
        )
    return 0


_DEFAULTS = {
    "arm": {"hold_seconds": 1.2, "speed": 2, "cycles": 2},
    "nod": {"hold_seconds": 0.35, "speed": 2, "count": 2},
    "rattle": {"volume": 5},
    "danger": {"volume": 5},
}

_ALLOWED_FLAGS = {
    "arm": {"--hold", "--speed", "--cycles"},
    "nod": {"--hold", "--speed"},
    "rattle": {"--volume"},
    "danger": {"--volume"},
    "cube": set(),
}


_LAST_PRIMITIVES = (
    "arm", "nod", "rattle", "danger",
    "animation", "brobots_stay_in_place", "move_reverse",
    "brobots_announce_in_sync", "brobots_sleep_to_wake_direct_sdk",
)


def _read_last_raw():
    """Reads phcal_last.json off disk, robot-keyed shape:
    {"1": {primitive: values}, "2": {primitive: values}}. If the file on disk
    is still the OLD flat/primitive-keyed shape ({"arm": {...}, ...} - a top-
    level key matching a known primitive name, which a robot-keyed file's
    "1"/"2" keys can never collide with), migrates it in place: every
    candidate brobot's slot starts as a full copy of the old shared values
    (non-lossy - nobody's saved tuning is lost), written back to disk
    immediately (PHCAL_ARROW_NAV_BUILD_PLAN_005.md lane (iv), Step 2)."""
    data = {}
    if LAST_PATH.exists():
        try:
            with open(LAST_PATH) as f:
                data = json.load(f)
        except (json.JSONDecodeError, OSError):
            data = {}
    if data and any(k in _LAST_PRIMITIVES for k in data.keys()):
        whichs = [c["which"] for c in _candidate_list() if c["which"]] or ["1", "2"]
        migrated = {which: json.loads(json.dumps(data)) for which in whichs}
        _write_last_raw(migrated)
        return migrated
    return data


def _write_last_raw(data):
    """Writes the robot-keyed dict straight to phcal_last.json."""
    LAST_PATH.parent.mkdir(parents=True, exist_ok=True)
    with open(LAST_PATH, "w") as f:
        json.dump(data, f, indent=2)
        f.write("\n")


def _resolve_last_whichs(which):
    """Fans the literal "both" out to every currently-relevant robot's own
    slot (_PRESENT_ROBOTS if detect-first has run, else _candidate_list()) -
    used by the two primitives (brobots_announce_in_sync,
    brobots_sleep_to_wake_direct_sdk) whose own robot prompt can genuinely
    resolve to "both" rather than a single "1"/"2". Anything else passes
    through unchanged as a single-item list."""
    if which == "both":
        whichs = [r["which"] for r in _PRESENT_ROBOTS if r.get("which")]
        if not whichs:
            whichs = [c["which"] for c in _candidate_list() if c["which"]]
        return whichs or ["1", "2"]
    return [which]


def _load_last(which):
    """Reads phcal_last.json's robot-keyed memory file, returns this ONE
    robot's own primitive:values dict (auto-migrating an old flat-shape file
    to the robot-keyed shape first if needed - see _read_last_raw()).
    Missing file, unreadable JSON, or a missing primitive/key inside it all
    fall back to the rung-1 (bingo-sourced) defaults above for exactly the
    piece that's missing - never a crash, never a guessed number outside
    those defaults. Returns a dict keyed "arm"/"nod", each a plain
    {cycles|count, hold, speed} dict - the same key names phcal_last.json
    itself uses (note: "hold", not "hold_seconds" - matches the spec's own
    key naming)."""
    raw = _read_last_raw()
    data = raw.get(which) or {}

    seed = {
        "arm": {
            "cycles": _DEFAULTS["arm"]["cycles"],
            "hold": _DEFAULTS["arm"]["hold_seconds"],
            "speed": _DEFAULTS["arm"]["speed"],
        },
        "nod": {
            "count": _DEFAULTS["nod"]["count"],
            "hold": _DEFAULTS["nod"]["hold_seconds"],
            "speed": _DEFAULTS["nod"]["speed"],
        },
        "rattle": {
            "volume": _DEFAULTS["rattle"]["volume"],
        },
        "danger": {
            "volume": _DEFAULTS["danger"]["volume"],
        },
        "animation": {"hold": ANIMATION_DEFAULT_HOLD_SECONDS},
        "brobots_stay_in_place": {"hold": HOLD_DEFAULT_SECONDS},
        "move_reverse": {"hold": MOVE_REVERSE_DEFAULT_HOLD_SECONDS},
        "brobots_announce_in_sync": {"phrase": "Brobots ready!"},
        "brobots_sleep_to_wake_direct_sdk": {"wait": SLEEP_WAKE_DEFAULT_WAIT_SECONDS},
    }
    merged = {}
    for primitive in _LAST_PRIMITIVES:
        merged[primitive] = {**seed[primitive], **(data.get(primitive) or {})}
    return merged


def _save_last(which, primitive, values):
    """Writes {which: {primitive: values}} into phcal_last.json, preserving
    every OTHER robot's own slot and every OTHER primitive within THIS
    robot's own slot untouched. Auto-migrates an old flat-shape file to the
    robot-keyed shape first if needed (see _read_last_raw()). Both entry
    paths (guided flow and direct-flag form) call this after firing, so
    either one updates the same one memory file, in the firing robot's own
    slot."""
    raw = _read_last_raw()
    robot_slot = raw.setdefault(which, {})
    robot_slot[primitive] = values
    _write_last_raw(raw)


PHCAL_PROMOTED_TWEAKS_PATH = GOPOD_REPO_ROOT / "goverlord/alias_lib_overlay/phcal_master_tweaks.json"


def _promote_tweaks():
    """Standalone promote step for phcal's own master tweaks file - mirrors
    run_golden_song_001.py's own _maybe_promote_knobs() shape exactly: same
    default-n y/n prompt, same shutil.copyfile byte-identical copy (no
    re-serialize), same only-if-source-exists guard. Not a new mechanism.
    Never auto-fired at the end of a phcal run - phcal gets invoked many
    times per tuning session (unlike a golden song's single end-of-run
    prompt), so a y/n on every fire would be noise; this is reachable on
    demand only (the phcal-promote alias), when the operator has actually
    landed on values worth banking."""
    if not LAST_PATH.exists():
        print(f"PHCAL_TWEAKS_PROMOTE_BLOCKED no {LAST_PATH} to promote from")
        return 1
    try:
        choice = _prompt_pick(
            {"y", "n"}, default="n", labels={"y": "yes", "n": "no"},
            question=f"update {PHCAL_PROMOTED_TWEAKS_PATH} from {LAST_PATH} for commit?",
        )
    except (EOFError, KeyboardInterrupt, _PhcalEscExit):
        choice = "n"
    if choice == "y":
        PHCAL_PROMOTED_TWEAKS_PATH.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(LAST_PATH, PHCAL_PROMOTED_TWEAKS_PATH)
        print(f"PHCAL_TWEAKS_PROMOTED {PHCAL_PROMOTED_TWEAKS_PATH} updated from {LAST_PATH}")
        return 0
    print(f"PHCAL_TWEAKS_PROMOTE_SKIPPED {PHCAL_PROMOTED_TWEAKS_PATH} left untouched")
    return 0


def _parse_rest(primitive, rest):
    """Splits the argv tail (everything after <primitive> <robot>) into a
    leading positional list and a flags dict. Recognizes --hold/--speed/
    --cycles (each takes the next token as its value); anything else
    starting with '--' is an unknown-flag error; anything not starting with
    '--' is a positional (nod's optional [count]; arm takes none). Raises
    ValueError with a plain message on any problem - never guesses."""
    positionals = []
    flags = {}
    i = 0
    while i < len(rest):
        tok = rest[i]
        if tok.startswith("--"):
            if tok == "--cycles" and primitive == "nod":
                raise ValueError("--cycles is arm-only; nod repeats via the positional [count] instead")
            if tok not in _ALLOWED_FLAGS[primitive]:
                raise ValueError(f"unknown flag '{tok}' for '{primitive}'")
            if i + 1 >= len(rest):
                raise ValueError(f"{tok} needs a value")
            flags[tok] = rest[i + 1]
            i += 2
        else:
            positionals.append(tok)
            i += 1
    return positionals, flags


def _parse_float_flag(name, raw, min_exclusive=0.0, max_value=None):
    try:
        val = float(raw)
    except ValueError:
        raise ValueError(f"{name} must be a number, got '{raw}'")
    if val <= min_exclusive:
        raise ValueError(f"{name} must be > {min_exclusive}, got {val}")
    if max_value is not None and val > max_value:
        raise ValueError(f"{name} must be <= {max_value} (BLOCKED - refusing to fire), got {val}")
    return val


def _parse_int_flag(name, raw, min_value=1, max_value=None):
    try:
        val = int(raw)
    except ValueError:
        raise ValueError(f"{name} must be an integer, got '{raw}'")
    if val < min_value:
        raise ValueError(f"{name} must be >= {min_value}, got {val}")
    if max_value is not None and val > max_value:
        raise ValueError(f"{name} must be <= {max_value}, got {val}")
    return val


def main():
    if len(sys.argv) < 3:
        print(
            "PHCAL_ISOLATE_USAGE phcal_isolate_001.py <arm|nod|rattle|danger|cube> <robot 1|2> "
            "[reps] [--hold S] [--speed N] [--cycles N] [--volume N]"
        )
        print("  (call with zero arguments for the guided prompt flow instead)")
        return 1

    primitive = sys.argv[1]
    robot = sys.argv[2]
    if robot not in ("1", "2"):
        print(f"PHCAL_ISOLATE_BLOCKED robot must be 1 or 2, got '{robot}'")
        return 1
    if primitive not in ("arm", "nod", "rattle", "danger", "cube"):
        print(f"PHCAL_ISOLATE_BLOCKED unknown primitive '{primitive}' - expected arm, nod, rattle, danger, or cube")
        return 1

    try:
        positionals, flags = _parse_rest(primitive, sys.argv[3:])
    except ValueError as exc:
        print(f"PHCAL_ISOLATE_BLOCKED {exc}")
        return 1

    live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
    os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
    mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
    serial = BROBOT_2_SERIAL if robot == "2" else BROBOT_1_SERIAL
    clock = _Clock()

    print(f"{clock.prefix()}PHCAL_ISOLATE primitive={primitive} robot={robot} serial={serial} live={live}")
    if not _phcal_session_checkin(clock):
        return 1
    if not live:
        print(f"{clock.prefix()}DRY: live robot gate is off (GOPOD_ALLOW_LIVE_ROBOT_SPEECH != 1)")

    if primitive == "rattle":
        if positionals:
            print(
                f"PHCAL_ISOLATE_BLOCKED rattle takes no positional argument (got {positionals!r}) "
                "- use --volume N instead"
            )
            return 1
        try:
            volume_ui = (
                _parse_int_flag("--volume", flags["--volume"], min_value=1, max_value=5)
                if "--volume" in flags
                else _DEFAULTS["rattle"]["volume"]
            )
        except ValueError as exc:
            print(f"PHCAL_ISOLATE_BLOCKED {exc}")
            return 1
        restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
        if not restart_preflight["ok"]:
            return 1
        preflight = run_preflight(mod, clock, serial, live)
        if not preflight["ok"]:
            return 1
        rc = cmd_rattle(mod, clock, serial, robot, live, volume_ui)
        _save_last(robot, "rattle", {"volume": volume_ui})
        print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=rattle path={LAST_PATH}")
        return rc

    if primitive == "danger":
        if positionals:
            print(
                f"PHCAL_ISOLATE_BLOCKED danger takes no positional argument (got {positionals!r}) "
                "- use --volume N instead"
            )
            return 1
        try:
            volume_ui = (
                _parse_int_flag("--volume", flags["--volume"], min_value=1, max_value=5)
                if "--volume" in flags
                else _DEFAULTS["danger"]["volume"]
            )
        except ValueError as exc:
            print(f"PHCAL_ISOLATE_BLOCKED {exc}")
            return 1
        restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
        if not restart_preflight["ok"]:
            return 1
        preflight = run_preflight(mod, clock, serial, live)
        if not preflight["ok"]:
            return 1
        rc = cmd_danger(mod, clock, serial, robot, live, volume_ui)
        _save_last(robot, "danger", {"volume": volume_ui})
        print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=danger path={LAST_PATH}")
        return rc

    if primitive == "cube":
        if positionals:
            print(
                f"PHCAL_ISOLATE_BLOCKED cube takes no positional or flag arguments (got {positionals!r})"
            )
            return 1
        restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
        if not restart_preflight["ok"]:
            return 1
        preflight = run_preflight(mod, clock, serial, live)
        if not preflight["ok"]:
            return 1
        rc = cmd_cube(mod, clock, serial, robot, live)
        return rc

    try:
        hold_seconds = (
            _parse_float_flag("--hold", flags["--hold"]) if "--hold" in flags else _DEFAULTS[primitive]["hold_seconds"]
        )
        speed = _parse_int_flag("--speed", flags["--speed"], min_value=1) if "--speed" in flags else _DEFAULTS[primitive]["speed"]
    except ValueError as exc:
        print(f"PHCAL_ISOLATE_BLOCKED {exc}")
        return 1

    if primitive == "arm":
        if positionals:
            print(
                f"PHCAL_ISOLATE_BLOCKED arm takes no positional argument (got {positionals!r}) "
                "- use --cycles N instead"
            )
            return 1
        try:
            cycles = (
                _parse_int_flag("--cycles", flags["--cycles"], min_value=2)
                if "--cycles" in flags
                else _DEFAULTS["arm"]["cycles"]
            )
        except ValueError as exc:
            print(f"PHCAL_ISOLATE_BLOCKED {exc}")
            return 1
        restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
        if not restart_preflight["ok"]:
            return 1
        preflight = run_preflight(mod, clock, serial, live)
        if not preflight["ok"]:
            return 1
        battery = run_battery_check(mod, clock, serial, live)
        if not battery["ok"]:
            print(f"{clock.prefix()}PHCAL_BLOCKED_LOW_BATTERY not firing arm - see PHCAL_BATTERY_CHECK line above")
            if PHCAL_PRIMITIVE_BATTERY_BLOCK_ENABLED:
                return 1
        rc = cmd_arm(mod, clock, serial, robot, live, cycles, hold_seconds, speed)
        _save_last(robot, "arm", {"cycles": cycles, "hold": hold_seconds, "speed": speed})
        print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=arm path={LAST_PATH}")
        return rc

    if len(positionals) > 1:
        print(f"PHCAL_ISOLATE_BLOCKED nod takes at most one positional [count] argument, got {positionals!r}")
        return 1
    count = _DEFAULTS["nod"]["count"]
    if positionals:
        try:
            count = _parse_int_flag("count", positionals[0], min_value=2)
        except ValueError as exc:
            print(f"PHCAL_ISOLATE_BLOCKED {exc}")
            return 1
    restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
    if not restart_preflight["ok"]:
        return 1
    preflight = run_preflight(mod, clock, serial, live)
    if not preflight["ok"]:
        return 1
    battery = run_battery_check(mod, clock, serial, live)
    if not battery["ok"]:
        print(f"{clock.prefix()}PHCAL_BLOCKED_LOW_BATTERY not firing nod - see PHCAL_BATTERY_CHECK line above")
        if PHCAL_PRIMITIVE_BATTERY_BLOCK_ENABLED:
            return 1
    rc = cmd_nod(mod, clock, serial, robot, live, count, hold_seconds, speed)
    _save_last(robot, "nod", {"count": count, "hold": hold_seconds, "speed": speed})
    print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=nod path={LAST_PATH}")
    return rc


class _PhcalEscExit(Exception):
    """2026-08-18, PHCAL_NAV_POLISH_001.md - raised by the raw-key reader
    when the operator presses ESC. Caught near the top of whatever loop is
    reading input (_run_guided_flow_once, _promote_tweaks) for a clean exit
    from ANY nesting depth, without threading a sentinel value through
    every intermediate return path between here and there. Terminal raw
    mode is always restored before this can propagate past _raw_mode()'s
    own try/finally - see that context manager's own docstring."""


class _PhcalBackToMenu(Exception):
    """2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md lane (iii): raised by a
    back-enabled leaf picker (Left pressed - _prompt_pick(back=True),
    _prompt_robot(), _prompt_animation_token(), _prompt_brobots_wake_chain())
    to abandon the CURRENT primitive's dispatch mid-flight, without firing
    anything, and resume navigate()'s own tree-walk exactly where it was
    left (same stack depth, same remembered highlight - see navigate()'s
    own try/except around its dispatch() call, the only place this is ever
    caught). Same "clean unwind through however many nested calls, no
    sentinel threading" shape as _PhcalEscExit above, scoped one level
    shallower: ESC abandons the whole guided-flow session, this only
    abandons the one primitive currently being answered.

    2026-08-25, PHCAL_ARROW_NAV_BUILD_PLAN_006.md Lane 3 (left-back to the
    ends): a SECOND-OR-LATER prompt in one primitive's own dispatch chain
    (e.g. animation's token pick, now that robot runs first; sleep_wake's
    robot pick, after its release-mode pick) can raise this too - design
    (b) from the Lane 3 survey, the plan's own recommendation. See
    _DISPATCH_ERASE_ACCUM's own comment just below for how the EARLIER,
    already-committed prompt(s) in the same chain get erased along with
    this one, so a Left press from ANY depth in a primitive's own chain
    unwinds the whole thing cleanly, not just its own current rows."""


_DISPATCH_ERASE_ACCUM = 0


def _accum_lines(n):
    global _DISPATCH_ERASE_ACCUM
    _DISPATCH_ERASE_ACCUM += n


def _erase_dispatch_accum():
    """Erases every EARLIER prompt's own already-committed lines in the
    current dispatch chain (see _DISPATCH_ERASE_ACCUM's own comment above),
    then resets the accumulator - called right before a second-or-later
    prompt raises _PhcalBackToMenu, in addition to (not instead of) that
    prompt's own current-row self-erase."""
    global _DISPATCH_ERASE_ACCUM
    if _DISPATCH_ERASE_ACCUM:
        _erase_screen_lines(_DISPATCH_ERASE_ACCUM)
    _DISPATCH_ERASE_ACCUM = 0


class _PhcalSessionModeRepick(Exception):
    """2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md Lane (v): the
    session-mode-changed recompute hook's own clean-unwind signal - same
    "propagate through however many nested calls, no sentinel threading"
    shape as _PhcalEscExit/_PhcalBackToMenu above, scoped to its own
    concern: "the operator wants to re-pick the session mode, mid-session,
    without quitting." navigate()/arrow_column_pick() stay
    session-mode-agnostic (same layering reason _PhcalBackToMenu's own
    docstring already gives - those primitives are meant to be reused by a
    future non-phcal caller with no session-mode concept at all), so this
    unwinds cleanly up through both to run_guided_flow()'s own handler -
    the one place session mode is ever actually set. Caught there,
    genuinely different from _PhcalEscExit's own handling: NOT an exit -
    the catch runs the repick screen, applies whatever the operator
    chose, then loops straight back into a fresh _run_guided_flow_once()
    pass (skipping _prompt_continue_or_exit() - a repick is a mid-flow
    detour, not a completed primitive dispatch), which is what actually
    re-runs _row_enabled() across every row and redraws in place - that
    pass's own existing _MENU_PASS_LINE_MARK erase-since-last-pass
    mechanism (see that global's own docstring) already does "redraw in
    place, no stacking" correctly for every OTHER pass boundary in this
    file; this hook reuses that exact primitive rather than inventing a
    second erase path just for a mode change.

    As of THIS lane, nothing in the file raises this exception yet - Lane
    (vi) is what actually wires a live keypress (arrow-up past the pinned
    root menu's own top row) to raise it. Deployed now as the backend half
    only (this class + run_guided_flow()'s catch + _repick_session_mode()
    below) - inert, unreachable, and verified not to change any EXISTING
    exception-handling path (run_guided_flow()'s existing _PhcalEscExit
    catch is untouched, only a new except clause is added alongside it).
    Case (b) from this lane's own build plan - "an Enter registers a
    choice that flips downstream required state" - does NOT exist
    anywhere in this file today: grepped every _SESSION_MODE=/
    _PRESENT_ROBOTS= assignment site directly, both are set in exactly
    one place (_resolve_session_mode_once(), once, before the guided-flow
    loop even starts) - no dispatch branch or Enter-registered choice
    mutates either global. Reported plainly per this lane's own
    instruction, not invented."""


def _raw_mode(fd):
    """Context manager: puts fd (normally stdin) into cbreak raw-input mode
    for the duration of the `with` block, and ALWAYS restores the original
    (cooked) terminal settings on the way out - normal return, an
    exception, a KeyboardInterrupt, all of it, via try/finally. This is the
    one place raw mode is ever entered in this file; every caller of the
    key reader below goes through here, so "terminal restore is mandatory
    on every exit path" is enforced structurally, not by convention at each
    call site. cbreak (not full raw) keeps signal generation (Ctrl-C still
    raises KeyboardInterrupt normally) and output post-processing intact -
    only input line-buffering/echo are disabled, which is all arrow-key
    reading needs."""
    import contextlib

    @contextlib.contextmanager
    def _cm():
        old = termios.tcgetattr(fd)
        try:
            tty.setcbreak(fd)
            yield
        finally:
            termios.tcsetattr(fd, termios.TCSADRAIN, old)

    return _cm()


def _read_key():
    """One logical key event from stdin, in raw mode. Returns a
    (kind, value) tuple: ("char", <1-char str>) for a printable keystroke,
    ("enter", None) for Enter/Return, ("backspace", None), ("esc", None)
    for a bare ESC (no bytes follow), or ("arrow", "up"/"down"/"left"/
    "right") for all four arrow directions. Reads a single raw byte first;
    if it's ESC (0x1b), does one more short, non-blocking-ish read to see
    whether a `[`-prefixed escape sequence follows - if nothing follows
    quickly, treats it as a bare ESC rather than blocking forever waiting
    for a sequence that isn't coming.

    2026-08-23: left/right used to be collapsed into up/down here (every
    call site was a vertical list, so left/right had no distinct meaning
    yet) - now surfaced as their own true values so a caller that DOES
    want to tell them apart (navigate()'s new Left-back wiring) can.
    arrow_column_pick()'s default behavior still treats left as up and
    right as down (its own `left_is_back` opt-in, not a change here) -
    every existing caller's up/down/select behavior is unchanged."""
    fd = sys.stdin.fileno()
    b = os.read(fd, 1)
    if not b:
        raise EOFError("no interactive input available")
    if b == b"\x1b":
        import select

        ready, _, _ = select.select([fd], [], [], 0.05)
        if not ready:
            return ("esc", None)
        b2 = os.read(fd, 1)
        if b2 != b"[":
            return ("esc", None)
        b3 = os.read(fd, 1)
        if b3 == b"A":
            return ("arrow", "up")
        if b3 == b"B":
            return ("arrow", "down")
        if b3 == b"C":
            return ("arrow", "right")
        if b3 == b"D":
            return ("arrow", "left")
        if b3 and (0x30 <= b3[0] <= 0x3F or 0x20 <= b3[0] <= 0x2F):
            while True:
                ready, _, _ = select.select([fd], [], [], 0.05)
                if not ready:
                    break
                nxt = os.read(fd, 1)
                if not nxt:
                    break
                if 0x40 <= nxt[0] <= 0x7E:
                    break
            return ("char", "")
        return ("esc", None)
    if b in (b"\r", b"\n"):
        return ("enter", None)
    if b in (b"\x7f", b"\x08"):
        return ("backspace", None)
    try:
        return ("char", b.decode("utf-8", errors="ignore"))
    except UnicodeDecodeError:
        return ("char", "")


_NAV_FRAME_MARK = "**"
_NAV_HINT_ARROW = "↑ ↓ select, Enter choose, ← back, Esc quit"
_NAV_HINT_VALUE = "Enter accept, type to change, ← back, Esc quit"
_NAV_DIVIDER = "-" * len(_NAV_HINT_ARROW)


def _draw_lines(lines):
    """THE self-reporting draw primitive, PHCAL_SELF_REPORTING_DRAW_
    EXECUTED_001.md - prints each line in `lines` and returns the exact
    count it drew. Built to kill a whole bug class the LOOK PASS's own
    survey (PHCAL_SECTION_TEMPLATE_GOLD_SURVEY_001.md) traced to its root:
    every erase call in this file used to receive a HAND-TYPED integer
    (`(1 if question else 0) + 3`, a bare `2`, `4`, ...) that had to be
    kept in sync by eye with whatever print statements actually ran above
    it - a display change (new frame line, reworded hint) could silently
    desync that count from reality, corrupting the next erase. Every
    caller below now builds its own header as a plain list of lines and
    reads ITS length back from this function instead of asserting a
    number - the count can never drift from what was actually printed,
    because it IS what was actually printed."""
    for line in lines:
        print(line)
    return len(lines)


_FRAME_STATE = "closed"  # "closed" | "open" | "pending_close"


def _frame_open():
    """Returns 1 if it actually printed a fresh mark, 0 if the seam
    collapsed (already open, or a pending close got cancelled instead) -
    callers building a measured header count via _draw_lines() must add
    this return value to their own total, never assume it's always 1."""
    global _FRAME_STATE
    if _FRAME_STATE in ("open", "pending_close"):
        _FRAME_STATE = "open"
        return 0
    print(_NAV_FRAME_MARK)
    _FRAME_STATE = "open"
    return 1


def _frame_close():
    global _FRAME_STATE
    if _FRAME_STATE == "open":
        _FRAME_STATE = "pending_close"


def _frame_flush_close():
    """Actually prints the pending close mark, if one exists - call this
    only where nothing framed follows immediately (an unframed diagnostic
    about to print, or the guided flow exiting)."""
    global _FRAME_STATE
    if _FRAME_STATE == "pending_close":
        print(_NAV_FRAME_MARK)
    _FRAME_STATE = "closed"


def _read_raw_line(prompt):
    """Reads one line of typed text in RAW mode, one keystroke at a time,
    via the same _read_key() reader every arrow-based prompt in this file
    already uses. Built 2026-08-25, PHCAL_ARROW_NAV_BUILD_PLAN_006.md Lane
    3 fix pass, to fix a real live trap: the old _prompt_value() ran plain
    input() in cooked/canonical mode, where ESC/Left only ever became
    literal bytes INSIDE the typed line, unresolved until Enter was ALSO
    pressed. Hit live at the sleep_wake `wait` field: operator pressed
    Left, then Esc, neither did anything (input() was still waiting on
    Enter) - had to Ctrl-C out to the shell. Raw mode fixes this at the
    root: both keys now resolve on the SAME single keypress, same
    reliability every other prompt in the tool already has.

    Returns the typed string on Enter. Raises `_PhcalEscExit()`
    immediately on ESC - matches every other prompt's own ESC=quit
    meaning (never a partial back-out, same contract `_NAV_LINE_MENU`
    already advertises - Esc was never really "back" anywhere in this
    tool, `_NAV_LINE_VALUE`'s old wording was simply wrong). Raises
    `_PhcalBackToMenu()` immediately on Left, but ONLY while the buffer is
    still empty - this is a typed VALUE field, not a text editor, so Left
    has no "move cursor within the line" job to compete with; once real
    text exists, Left is ignored (Backspace clears it back to empty
    first, then Left backs out - always a real, non-trapping path, never
    a second silent meaning for the same key). Prints only `prompt` and
    the typed echo - the caller (`_prompt_value()`) owns its own
    accumulator/erase bookkeeping around this call, same as every other
    back-enabled prompt in the file."""
    buf = []
    sys.stdout.write(prompt)
    sys.stdout.flush()
    with _raw_mode(sys.stdin.fileno()):
        while True:
            kind, val = _read_key()
            if kind == "esc":
                print()
                raise _PhcalEscExit()
            if kind == "arrow" and val == "left" and not buf:
                print()
                raise _PhcalBackToMenu()
            if kind == "enter":
                print()
                return "".join(buf)
            if kind == "backspace":
                if buf:
                    buf.pop()
                    sys.stdout.write("\b \b")
                    sys.stdout.flush()
                continue
            if kind == "char" and val:
                buf.append(val)
                sys.stdout.write(val)
                sys.stdout.flush()


def _prompt_pick(choices, default=None, exit_on_invalid=False, labels=None, back=False, question=None):
    """2026-08-21, PHCAL_NAV_CONSOLIDATION_001.md (PHCAL_INPUT_TREE_SURVEY_
    001.md §5 steps 1+3): replaces the old `_prompt_choice()` - the second,
    independent raw-mode key reader this file used to carry alongside
    arrow_column_pick(). Deleted outright, not deprecated in place; this is
    a thin wrapper matching its old inline-mode signature exactly (every
    former caller changed only its function name, not its own
    choices/default/exit_on_invalid arguments) built entirely on
    arrow_column_pick() - the one input engine now. `_confirm_multi_mode()`
    (the file's one former `options=` caller) calls arrow_column_pick()
    directly instead of through this wrapper, since it already builds its
    own real (key, label) options list with descriptive labels; this
    wrapper's own options are always bare key==label pairs, which is all a
    short y/n or numbered/lettered pick ever needs.

    exit_on_invalid decided by prompt SHAPE, per the survey's own rule: a
    y/n CONFIRM passes False (retry-on-typo, matching every confirm's old
    behavior); a numbered/lettered PICK passes True (eject-with-one-warning,
    matching how the group/sub-menu tree already behaved) - one rule, set
    explicitly at each of this wrapper's own callers below, not left to
    drift per call site the way the old per-call boolean did.

    IMPORTANT, found live during this pass's own dry-verify (not part of
    the original survey): every former exit_on_invalid=True caller of the
    old `_prompt_choice()` (song pick, step pick, animation-token,
    _prompt_robot, release-mode) was written assuming a valid answer
    ALWAYS eventually comes back - most of those callers use the return
    value immediately with no None-check (`int(dir_choice) - 1`,
    `_ANIMATION_TOKEN_MENU[token_choice]`, a robot-serial lookup that
    silently falls back to Brobot 1 on anything not literally "2"). Under
    the OLD engine that was safe by accident, because exit_on_invalid=True
    callers there had never actually been exercised with real garbage
    input in a way that got audited - the old inline mode's own default
    (False) covered every OTHER call site's real-world typo path. Rather
    than audit and patch every one of those call sites individually, this
    wrapper converts an exit_on_invalid=True engine None straight into a
    `_PhcalEscExit` - the same clean-abort signal ESC already raises,
    already caught at every level that matters (run_guided_flow()'s own
    try/except, _promote_tweaks()'s own). So "eject on invalid" now
    actually means "abandon this whole operation, same as ESC" tool-wide,
    not "return None and hope the caller checked" - no silent wrong-robot
    fire, no crash, one consistent meaning.

    2026-08-22, PHCAL_ARROW_NAV_BUILD_PLAN_002.md Phase 1: "0" is REMOVED
    tool-wide as a choice - no caller passes it anymore, so the "0" ->
    "exit" label special-case below is gone too. ESC is now the only way
    to exit from any of this wrapper's callers. The `if choice is None`
    check just below is left in place as an inert safety net - with no
    typed-input path left in arrow_column_pick(), it can no longer
    actually fire, but removing it isn't part of this phase's own scope.

    2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md leaf-picker text cleanup:
    the old typed-style `prompt` string (e.g. "pick a token [*,1-3]: ") is
    gone - dead weight now that arrows are the only way to answer, matching
    navigate()'s own header-less tree rows. `labels`, optional, maps a
    choice key to its real display text (e.g. {"y": "yes", "n": "no"}) so
    each row shows its label once instead of the old `{key}. {key}`
    doubling (`show_key=False` below); a caller that omits `labels` still
    gets a working picker (falls back to key==label), but every call site
    in this file passes one as of this pass.

    2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md lane (iii), pinned-menu/
    left-back layout: `back`, default False. A caller mid-way through
    collecting one primitive's own leaf answers (from inside
    _dispatch_primitive(), directly or via a helper it calls) passes
    back=True so Left abandons THIS pick - and everything this primitive
    dispatch was in the middle of building up - and unwinds cleanly back
    to navigate()'s own tree (see _PhcalBackToMenu's own docstring), which
    resumes exactly where the operator left it. The 3 callers with no
    navigate() to unwind to - the low-battery gate (runs before any menu
    exists), _promote_tweaks() (its own standalone entry point, never
    reached through navigate()), _prompt_continue_or_exit() (runs AFTER a
    primitive already finished - nothing left to abandon) - all pass the
    default False, unchanged Left-moves-like-Up behavior.

    2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md question-text restore:
    the leaf-picker text cleanup above dropped the old typed-style
    `prompt` string entirely - correctly, for its own typed-syntax tail
    ("y/n [default n]: ", "[*,1-3]: "), but that also silently took the
    real human-readable QUESTION with it, leaving bare option rows with
    no idea what's being asked. `question`, optional, restores just that
    part: a plain `print(question)` right above the rows, no tail, no
    `{key}. {key}` doubling - text only, the picker mechanism itself is
    unchanged. When `back=True` and a question was printed,
    `erase_on_back_extra=1` is passed through so Left-back takes the
    question line with it instead of stranding it above a blank row
    region (see arrow_column_pick()'s own docstring).

    2026-08-26, PHCAL_ARROW_NAV_BUILD_PLAN_006.md LOOK PASS: every call now
    prints the full screen frame - `_NAV_FRAME_MARK` open, `question` (if
    given), `_NAV_HINT_ARROW`, `_NAV_DIVIDER` before the rows, replacing
    Lane 1's single always-on `_NAV_LINE_CHOICE`. A closing
    `_NAV_FRAME_MARK` prints once the pick actually resolves (Enter) -
    never on a Left-back, since nothing was "finished" in that case; the
    whole block (frame open through rows) erases instead.

    2026-08-26, PHCAL_SELF_REPORTING_DRAW_EXECUTED_001.md: `header_count`
    now comes from `_draw_lines()`'s own return value - the REAL number of
    header lines just printed - instead of a hand-typed `3 + (1 if
    question else 0)` that had to be kept in sync by eye. Every erase
    count derived from it (`erase_on_back_extra`, `own_lines`) inherits
    that same measured number, so a future header change (one more line,
    one fewer) can never silently desync this call's own erase math
    again.

    2026-08-26, PHCAL_LOOK_PASS_CORRECTION_001.md (per the operator's own
    literal PDF spec): each row now carries its own "00. "/"01. " position
    prefix (per-screen, restarting at 00 every call - not a global ID, not
    the choice's own key). A divider now also prints immediately after the
    frame mark, before `question`/the hint - the PDF's own general shape
    is `**` -> divider -> content -> divider (repeat) -> divider -> `**`,
    not divider-then-content-then-close this call used to build.
    `_frame_open()`'s own return value (0 or 1 - see its docstring, the
    seam-collapse fix) is added into `header_count` instead of assuming
    the frame mark always printed exactly 1 line."""
    labels = labels or {}
    sorted_choices = sorted(choices)
    options = [(c, f"{i:02d}. {labels.get(c, c)}") for i, c in enumerate(sorted_choices)]
    highlight = sorted_choices.index(default) if default in sorted_choices else 0
    frame_count = _frame_open()
    header_lines = [_NAV_DIVIDER]
    if question:
        header_lines.append(question)
    header_lines.append(_NAV_HINT_ARROW)
    header_lines.append(_NAV_DIVIDER)
    header_count = frame_count + _draw_lines(header_lines)
    own_lines = header_count + len(options)
    if back:
        choice, _hl = arrow_column_pick(
            options, highlight=highlight, exit_on_invalid=exit_on_invalid, show_key=False,
            left_is_back=True, erase_on_back=True, erase_on_back_extra=header_count,
        )
        if choice is _NAV_BACK:
            _erase_dispatch_accum()
            raise _PhcalBackToMenu()
    else:
        choice = arrow_column_pick(options, highlight=highlight, exit_on_invalid=exit_on_invalid, show_key=False)
    if choice is None:
        raise _PhcalEscExit()
    _accum_lines(own_lines)
    _frame_close()
    return choice


_NAV_BACK = object()

_MENU_PASS_LINE_MARK = None


def _erase_screen_lines(n):
    """Move the cursor up n lines and blank each one, ending back at the
    top of that now-empty region - the one shared 'erase a previously-
    drawn block instead of stacking below it' primitive this fix adds.
    Safe only when nothing else has printed since those n lines were
    drawn (see _LAST_MENU_SCREEN_LINES's own docstring above) - callers
    are responsible for that precondition, this function just does the
    erase. No-op for n <= 0."""
    if n <= 0:
        return
    sys.stdout.write(f"\x1b[{n}A")
    for _ in range(n):
        sys.stdout.write("\r\x1b[2K\n")
    sys.stdout.write(f"\x1b[{n}A")
    sys.stdout.flush()


def arrow_column_pick(options, highlight=0, window_height=None, exit_on_invalid=True, left_is_back=False, show_key=True, erase_lines=0, initial_draw=True, erase_on_back=False, erase_on_back_extra=0, top_up_repick=False, esc_home=None, skip_indices=frozenset()):
    """2026-08-19, NAV_PATTERN_SURVEY_001.md / NAV_PRIMITIVE_BUILT_001.md.
    2026-08-21, PHCAL_NAV_CONSOLIDATION_001.md (PHCAL_INPUT_TREE_SURVEY_001.md
    §5 step 1): this is now THE one input engine in this file - the
    reusable single-level arrow-column picker, with real scrolling for long
    lists. `_prompt_choice()` (the old second, independent raw-mode key
    reader) is deleted outright as of this pass; every one of its 13
    former callers now reaches this function instead, most through the
    thin `_prompt_pick()` wrapper below, `_confirm_multi_mode()` directly
    (it already built a native (key, label) options list). One raw-mode
    key-reading loop left in the whole file, not two.

    options: ordered list of (key, label) tuples. highlight: initial
    highlighted index. window_height: max visible rows at once. None (or
    >= len(options)) means "show everything, no scrolling" - the size
    phcal's own group/sub-menus actually use today (<=9 items per level,
    NAV_PATTERN_SURVEY_001.md Part 2). When len(options) > window_height, a
    scroll offset tracks which window_height-sized slice is drawn; the
    highlighted row always stays visible - the window follows it past the
    top/bottom edge instead of reprinting the full list (up to 69 items,
    the future composition editor's own line-list size) on every key
    press.

    Arrow up/down moves+wraps across the full option list. Enter selects
    the highlighted item. ESC raises _PhcalEscExit immediately, uncaught
    here - propagates to whatever already catches it (run_guided_flow()'s
    own try/except), unchanged.

    2026-08-22, PHCAL_ARROW_NAV_BUILD_PLAN_002.md Phase 1: the typed-buffer
    fallback (type a key, Enter resolves it) is REMOVED - no typed-input
    path remains anywhere in this file. Enter always resolves to whatever
    row is currently highlighted; a bare character/backspace keypress is
    now simply ignored (no branch matches it, the loop just reads the next
    key). Because Enter can no longer produce a value outside `choices`,
    this function can no longer return None - the old "0"/exit_on_invalid
    invalid-input handling (PHCAL_GUIDED_INVALID/_EXIT, the "0" CONFIRM
    special-case) is dead code as a direct result and has been removed from
    the body below. `exit_on_invalid` stays in the signature, unused,
    rather than auditing/changing every one of its ~13 call sites in this
    same pass - flagged, not silently dropped, per
    PHCAL_ARROW_NAV_BUILD_PLAN_002.md's own Phase 1 caution.

    2026-08-23: `left_is_back`, default False, preserves every existing
    caller's behavior exactly (left still moves the highlight like up,
    right still moves it like down - unchanged). navigate() is the one
    caller that passes True: there, Left means "back one menu level," a
    different thing from "move up in this list," so a Left press returns
    the _NAV_BACK sentinel immediately instead of moving the highlight.
    Right/Up/Down/Enter are untouched by this opt-in either way.

    2026-08-23: `show_key`, default True, preserves every existing caller's
    rendering exactly (`{key}. {label}`). navigate() is the one caller that
    passes False - its tree rows carry their own fully-formed label text
    now (PHCAL_ARROW_NAV_BUILD_PLAN's row-format pass), so the numbered
    key prefix is dead weight there; every other caller (the y/n confirms,
    `_prompt_robot()`, animation-token pick, etc.) is untouched and still
    shows its key.

    2026-08-23, redraw-in-place foundation: `erase_lines`, default 0,
    preserves every existing caller's rendering exactly (prints fresh,
    same as always). navigate() is the one caller that passes a nonzero
    value - the exact line count (row count + 1 trailing blank) its own
    PRIOR level left on screen, so this call's first draw overwrites that
    region instead of stacking below it. See _erase_screen_lines()'s own
    docstring above for the safety precondition (nothing else printed in
    between) every caller of this parameter must satisfy.

    2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md lane (iii), pinned-menu/
    left-back layout: two more opt-ins, both default to the pre-existing
    behavior for every caller that doesn't pass them.

    `initial_draw`, default True: when a caller passes False, the whole
    first-draw block below (erase_lines handling + the initial _redraw)
    is skipped entirely - this call goes straight into reading keys,
    reusing whatever this exact level's own rows already correctly show
    on screen from a PRIOR call. navigate()'s own pinned root and any
    already-open level resume this way (their rows never became stale -
    nothing below them was left half-erased, see `erase_on_back` below),
    so re-entering them needs no redraw, only a fresh raw-mode read loop
    (a new Python call is only needed at all because returning to
    navigate()'s own dispatch() is what a Left-back / leaf-fire-complete
    boundary actually is).

    `erase_on_back`, default False: when True and `left_is_back` fires a
    Left press, this call erases its OWN currently-visible
    window_height rows (_erase_screen_lines - the exact same primitive
    every other redraw here already uses) instead of printing a trailing
    blank line, so the region collapses to blank instead of leaving
    stale rows behind - "the submenu region clears to blank" / a leaf
    picker vanishing on back. Every leaf-picker wrapper that raises
    _PhcalBackToMenu on Left (_prompt_pick(back=True), _prompt_robot(),
    _prompt_animation_token(), _prompt_brobots_wake_chain()) passes this;
    navigate()'s own non-root levels do too. navigate()'s ROOT level
    passes False - Left at the root is a deliberate no-op (unchanged
    behavior), nothing to erase since root never stops being shown.

    2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md question-text restore:
    `erase_on_back_extra`, default 0 - the plain QUESTION header line a
    caller prints ABOVE this call's own rows (e.g. _prompt_pick's own
    `question` param) sits outside `window_height` (it's not one of the
    options, navigate() never redraws it on its own up/down), so a
    caller that printed one passes 1 here - erase_on_back then blanks
    `window_height + 1`, taking the question line with it instead of
    stranding it above a now-empty row region.

    2026-08-24: when `left_is_back` is True, BOTH exit paths (Enter and
    Left) now return a `(result, highlight)` tuple instead of a bare
    value - `result` is the picked key on Enter or the _NAV_BACK sentinel
    on Left, `highlight` is this call's own final highlighted index, so a
    caller that needs to resume this exact level later (navigate()'s own
    loop, returning to a level after popping out of something deeper)
    can seed the next call's own `highlight` param correctly instead of
    silently resetting to 0. Every `left_is_back=False` caller (the
    default - unchanged for every pre-existing call site) still gets the
    bare value back, exactly as before.

    2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md Lane (vi): `top_up_repick`,
    default False - when True, pressing Up while `highlight == 0` (the
    very top row) raises _PhcalSessionModeRepick() instead of wrapping the
    highlight to the bottom row - "arrow up past the top row reaches the
    at-start mode picker, which sits above the menu in the tour." Erases
    this call's own currently-visible `window_height` rows first (the
    same _erase_screen_lines() primitive `erase_on_back` already uses
    above - the region collapses to blank, same visual convention), then
    raises - Lane (v)'s own exception, caught by run_guided_flow(), never
    here (same layering as _PhcalEscExit/_PhcalBackToMenu - this function
    stays session-mode-agnostic). navigate() is the one caller that passes
    True, and only for the ROOT level (`is_root`) - every submenu keeps
    the pre-existing Up-wraps-to-bottom behavior, unchanged. Only `val ==
    "up"` is affected; Left (already claimed by `left_is_back` above) and
    Down/Right are untouched by this opt-in either way.

    2026-08-25, PHCAL_ARROW_NAV_BUILD_PLAN_006.md Lane 1a: `esc_home`,
    default None - an index. When set and ESC is pressed while `highlight
    != esc_home`, this bumps the highlight back to `esc_home` and redraws
    instead of exiting - a plain, reusable "ESC returns to the default
    row" two-state behavior. ESC while already AT `esc_home` still exits
    normally (raises _PhcalEscExit, unchanged). This is
    _confirm_multi_mode()'s own former bespoke two-state ESC handling
    (highlight==-1 "at rest" vs. a real highlighted row), generalized into
    this one engine so that screen no longer needs its own separate
    raw-mode loop - see that function's own rewritten docstring.

    2026-08-26, PHCAL_ARROW_NAV_BUILD_PLAN_006.md LOOK PASS: `skip_indices`,
    default empty (zero behavior change for every existing caller) - a set
    of option indices Up/Down must never land on, stepped over in the same
    direction instead. Built for exactly one caller (navigate(), for the
    main menu's own divider row between the fireable calibrations and the
    session/config pair - see _build_primitive_group_tree()'s own comment)
    - a non-selectable row still renders (via the SAME generic `_redraw()`
    loop below, no special-casing there) but can never be highlighted or
    Entered, since `highlight` provably never equals a skipped index.
    `highlight`/the initial value a caller passes must not itself be a
    skip index - callers are responsible for that, same as every other
    precondition arrow_column_pick() already documents rather than
    defensively checks."""
    choices = [k for k, _label in options]
    n = len(options)
    if window_height is None or window_height >= n:
        window_height = n
    offset = 0

    def _redraw(first=False):
        nonlocal offset
        if highlight < offset:
            offset = highlight
        elif highlight >= offset + window_height:
            offset = highlight - window_height + 1
        if not first:
            sys.stdout.write(f"\x1b[{window_height}A")
        for i in range(window_height):
            real_index = offset + i
            key, label = options[real_index]
            if real_index in skip_indices:
                sys.stdout.write(f"\r\x1b[2K{label}\n")
                continue
            marker = "> " if real_index == highlight else "  "
            row_text = f"{key}. {label}" if show_key else label
            sys.stdout.write(f"\r\x1b[2K{marker}{row_text}\n")
        sys.stdout.flush()

    if initial_draw:
        if erase_lines:
            _erase_screen_lines(erase_lines)
        _redraw(first=True)
        if erase_lines > window_height:
            _extra = erase_lines - window_height
            for _ in range(_extra):
                sys.stdout.write("\r\x1b[2K\n")
            sys.stdout.write(f"\x1b[{_extra}A")
            sys.stdout.flush()

    with _raw_mode(sys.stdin.fileno()):
        while True:
            kind, val = _read_key()
            if kind == "esc":
                if esc_home is not None and highlight != esc_home:
                    highlight = esc_home
                    _redraw()
                    continue
                print()
                raise _PhcalEscExit()
            if kind == "arrow":
                if top_up_repick and val == "up" and highlight == 0:
                    _erase_screen_lines(window_height)
                    raise _PhcalSessionModeRepick()
                if left_is_back and val == "left":
                    if erase_on_back:
                        _erase_screen_lines(window_height + erase_on_back_extra)
                    else:
                        print()
                    return (_NAV_BACK, highlight) if left_is_back else _NAV_BACK
                if not left_is_back and val == "left" and esc_home is not None and highlight != esc_home:
                    highlight = esc_home
                    _redraw()
                    continue
                step = -1 if val in ("up", "left") else 1
                new_highlight = (highlight + step) % n
                while new_highlight in skip_indices:
                    new_highlight = (new_highlight + step) % n
                highlight = new_highlight
                _redraw()
                continue
            if kind == "enter":
                print()
                choice = choices[highlight]
                return (choice, highlight) if left_is_back else choice


def navigate(tree, dispatch, window_height=None):
    """2026-08-19, NAV_PATTERN_SURVEY_001.md / NAV_PRIMITIVE_BUILT_001.md:
    the reusable generic N-deep tree walker, built on arrow_column_pick()
    above - this is what lets an operator "arrow anywhere," one walker,
    any depth, instead of a hand-written while/continue loop copy-pasted
    per menu level.

    `tree`: a dict mapping pick_key -> (display_label, child), where
    child is either another such dict (descend one level) or a plain
    string (a real leaf primitive name). `dispatch(leaf)` is called once a
    leaf is picked - whatever it returns becomes navigate()'s own return
    value, ending this call. ESC is still the only way to abandon a
    navigate() call outright (raises _PhcalEscExit, uncaught here,
    propagates to whatever already catches it - unchanged).

    2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md lane (iii), pinned-menu/
    left-back layout - the real shape change this pass makes, replacing
    the 2026-08-23 "each level fully replaces the last" walk:

    - The ROOT level (stack[0]) draws once and is never erased again -
      pinned, always in view. Every level below it (an open submenu, or
      whatever `dispatch` draws once it starts collecting a primitive's
      own answers) is the one thing that gets erased/redrawn as the
      operator moves through it.
    - Right and Enter both descend (Enter both selects AND, on a dict
      child, opens its submenu - Right has no separate meaning here
      beyond what Enter already does, `_read_key()` doesn't distinguish
      "open" from "select" as different gestures for this tree).
    - Left backs OUT: at a submenu, pops back to root, erasing the
      submenu's own rows to blank (arrow_column_pick's erase_on_back) -
      root's own rows are untouched, they were never touched to begin
      with. At the root, Left is a deliberate no-op (nothing above root
      to return to - mid-session mode re-pick isn't built yet).
    - `dispatch(leaf)` is called INSIDE this loop (not by the caller,
      after navigate() returns) specifically so a leaf picker's own
      Left-back (raising _PhcalBackToMenu - see that exception's own
      docstring) can be caught right here, where `stack`/`highlights`
      are still exactly where the operator left them - the loop just
      `continue`s, resuming the SAME level (root or submenu) with its
      own remembered highlight, no redraw needed (its rows never
      changed - see `initial_draw=False` below).
    - `highlights`, one entry per stack level, remembers each level's
      own last-highlighted index so returning to a level (after popping,
      or after a leaf dispatch backs out) resumes at the same row
      instead of resetting to the top. Set from arrow_column_pick()'s
      own now-tuple return (`left_is_back=True` always returns
      `(result, highlight)` - see that function's own docstring) after
      every call, whichever way it exited.
    - `fresh`, a single local flag: True only for a level's very first
      draw (the initial root entry, or right after a submenu is freshly
      pushed) - that call gets `initial_draw=True` (draw for real,
      erase_lines=0 since nothing is in its own region yet). Every OTHER
      re-entry into a level - root's own Left-no-op, resuming a level
      after popping out of what was below it, resuming after a leaf
      dispatch backs out - reuses `initial_draw=False`: that level's own
      rows are still correct on screen (nothing below it left them
      stale; whatever WAS below either erased itself via
      `erase_on_back` before handing back control, or was never drawn
      at all), so no redraw is needed, only a fresh raw-mode read loop.

    2026-08-22, PHCAL_ARROW_NAV_BUILD_PLAN_002.md Phase 1: the
    auto-injected "0" back/exit row is REMOVED - no typed-input path
    remains anywhere in this file.

    2026-08-23, row-format build request: calls arrow_column_pick() with
    show_key=False - tree row labels carry their own full text (Brobots
    prefix, [disabled] prefix, no leading number), so the generic
    "{key}. " rendering would just be redundant/dead weight here. Every
    other arrow_column_pick() caller in the file is untouched.

    window_height: default None, meaning "size to this level's own real
    content, every level, dynamically" - NOT a fixed number (2026-08-20,
    PHCAL_NAV_LIVETEST_FINDINGS_001.md -> PHCAL_NAV_FIXES_001.md,
    operator decision: a static default re-breaks the instant any menu's
    real item count differs from whatever number it was tuned for - it did
    exactly that here, once, already). When None, EACH level computes its
    own window height as len(options) for THAT level, so every level's
    window is exactly its own real-item count, with no menu-wide constant
    anywhere to drift stale as menus grow, shrink, or differ level to
    level (root vs. a 2-3-item sub-menu). A caller that DOES want real
    scrolling (the future composition editor's up-to-69-line lists,
    arrow_column_pick()'s own docstring) passes an explicit integer here,
    same as before - that value then applies uniformly to every level of
    that one navigate() call, unchanged behavior for that case.

    2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md Lane (vi): passes
    `top_up_repick=is_root` - arrow_column_pick()'s own new opt-in (see
    its docstring) only fires at the ROOT level; a submenu's Up-wraps-to-
    bottom behavior is unchanged. This is the ONLY change this lane makes
    to navigate() itself - _PhcalSessionModeRepick is not caught here
    (same "stays session-mode-agnostic" reasoning _PhcalBackToMenu's own
    catch above does NOT extend to), it propagates straight out to
    run_guided_flow()'s own handler, same shape as _PhcalEscExit already
    does."""
    stack = [tree]
    highlights = [0]
    fresh = True
    while stack:
        depth = len(stack)
        is_root = depth == 1
        node = stack[-1]
        keys = sorted(node, key=lambda k: (len(k), k))
        options = [(k, node[k][0]) for k in keys]
        skip_indices = {i for i, k in enumerate(keys) if node[k][1] is None}
        level_height = window_height if window_height is not None else len(options)
        result, hl = arrow_column_pick(
            options, highlight=highlights[-1], window_height=level_height,
            left_is_back=True, show_key=False, erase_lines=0,
            initial_draw=fresh, erase_on_back=not is_root,
            top_up_repick=is_root, skip_indices=skip_indices,
        )
        highlights[-1] = hl
        fresh = False
        if result is _NAV_BACK:
            if not is_root:
                stack.pop()
                highlights.pop()
            continue
        _label, child = node[result]
        if isinstance(child, dict):
            stack.append(child)
            highlights.append(0)
            fresh = True
            continue
        try:
            return dispatch(child)
        except _PhcalBackToMenu:
            continue


def _prompt_robot(default=None, allow_both=False, back=False):
    """Dedicated 'which robot?' prompt - its own process, single purpose.
    2026-08-15: replaces 7 separate hand-rolled copies of
    _prompt_choice("robot (1 or 2)? ", {"1","2"}) that had drifted across
    the guided flow (animation, brobots_stay_in_place,
    brobots_session_responsiveness, move_reverse, cube, the shared
    arm/nod/rattle/danger tail, plus sleep_wake's own "1, 2, or both"
    variant) - one place now, not seven. Deliberately NOT merged with
    _prompt_animation_token below or the primitive-menu's own print/pick
    logic - decoupled, each its own process, not a shared generic "pick
    from anything" utility.

    2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md leaf-picker text cleanup:
    the old `prompt_note` free-text suffix (e.g. "[default 2 - Brobot 2,
    the cube keeper]") is gone - the picker's own pre-highlighted default
    row already shows which choice is the default, same as every other
    leaf pick. The base QUESTION itself ("robot (1 or 2)?" / "robot (1,
    2, or * for both)?") was ALSO dropped that same pass, as an
    unintended side effect of removing the typed-syntax tail it used to
    sit next to - restored (question-text-only, this param unaffected)
    2026-08-24, question-text restore.

    2026-08-21, PHCAL_NAV_CONSOLIDATION_001.md (PHCAL_INPUT_TREE_SURVEY_
    001.md §4 finding 1 / §5 step 2): "0" used to be a redundant extra
    synonym for "both" here, colliding with "0"'s ONE tool-wide meaning
    everywhere else (back/exit in a hierarchical pick) - dropped outright.

    2026-08-21, PHCAL_NAV_BASE_RULES_FIXES_001.md (PHCAL_NAV_BASE_RULES_
    SURVEY_001.md §5 step 4): "0" is RE-ADDED, deliberately, as its own
    choice meaning clean exit (raises _PhcalEscExit).

    2026-08-21, PHCAL_LESS_FLAKY_SWEEP_001.md: allow_both's own key for
    "both robots" was the literal typed WORD "both" - the one place in
    this file where a caller had to type more than a single character to
    resolve a pick, inconsistent with every other key in the tool and the
    same "flaky, arbitrary key" complaint that got "d" removed from
    _confirm_multi_mode's dry-run row this same pass. Re-keyed to "*" -
    the same tool-wide all-flag animation's "all" tokens and
    _confirm_multi_mode's own multi-continue row already use, so "pick
    everything in this set" now means the same single character
    everywhere it comes up. Built via a direct arrow_column_pick() call
    (not the _prompt_pick() wrapper) so the printed row can say "both"
    while the typed KEY is "*" - the function's own return-value contract
    to its callers is unchanged, still the literal string "both"
    (_sleep_wake_specs()'s own `if which in ("1", "both")` checks depend on
    this exact string, untouched by this pass).

    2026-08-18 (PHCAL_DETECT_FIRST_001.md): when detect-first has set
    session_mode="single" (module-level _SESSION_MODE, set once by
    run_guided_flow() before any of this file's dispatch branches run),
    this auto-resolves to the one detected-present robot's own "1"/"2"
    identity instead of prompting at all - true single-wake, and every
    other single-robot primitive, now just work with whichever robot is
    actually plugged in (Brobot 1 OR Brobot 2, not hardcoded to either).
    MULTI mode (_SESSION_MODE="multi") and the no-detection case
    (_SESSION_MODE is None - the direct-flag main() path never calls
    run_guided_flow() and never sets this) are both completely unaffected -
    this whole block is skipped, prompting exactly as before this pass.

    2026-08-22, PHCAL_ARROW_NAV_BUILD_PLAN_002.md Phase 1: "0" (clean
    exit) is REMOVED from both pick shapes below - no typed-input path
    remains anywhere in this file, and ESC is now the only way to exit
    from this prompt.

    2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md lane (iii): `back`,
    default False. Originally only safe as the FIRST widget in its
    caller's own dispatch chain - left False at animation's own robot pick
    (its token picker ran first back then) and brobots_sleep_to_wake_
    direct_sdk's robot pick (its release-mode picker runs first), flagged
    per that build's own STOP clause, not forced.

    2026-08-25, PHCAL_ARROW_NAV_BUILD_PLAN_006.md Lane 1 (robot-first
    order): animation's own robot pick now runs FIRST in its chain (its
    token pick moved second) - `back=True` there again, same as every
    other single-robot primitive.

    2026-08-25, PHCAL_ARROW_NAV_BUILD_PLAN_006.md Lane 3 (left-back to the
    ends), design (b): sleep_to_wake's own robot pick (still SECOND, after
    its release-mode pick) now also passes `back=True` - safe now that a
    Left here also erases the release-mode pick's own already-committed
    rows via _DISPATCH_ERASE_ACCUM/_erase_dispatch_accum(), not just this
    call's own current rows. The single-mode auto-resolve branch above
    never shows a picker at all, `back` has no effect there."""
    if _SESSION_MODE == "single" and _PRESENT_ROBOTS and _PRESENT_ROBOTS[0]["which"]:
        resolved = _PRESENT_ROBOTS[0]["which"]
        confirmed = bool(_PRESENT_ROBOTS[0].get("present", True))
        reason = "the only one detected present" if confirmed else "forced by the operator, NOT confirmed present"
        _accum_lines(_draw_lines([f"PHCAL_SINGLE_MODE_AUTO_RESOLVE robot={resolved} ({_PRESENT_ROBOTS[0]['label']}, {reason})"]))
        return resolved
    if allow_both:
        base_options = [("1", "Robot 1"), ("2", "Robot 2"), ("*", "both (1 and 2)")]
        option_keys = [k for k, _label in base_options]
        default_key = "*" if default == "both" else default
        highlight = option_keys.index(default_key) if default_key in option_keys else 0
        options = [(k, f"{i:02d}. {label}") for i, (k, label) in enumerate(base_options)]
        frame_count = _frame_open()
        header_count = frame_count + _draw_lines([
            _NAV_DIVIDER, f"ENTER for {base_options[highlight][1]}, or change:", _NAV_HINT_ARROW, _NAV_DIVIDER,
        ])
        own_lines = header_count + len(options)
        if back:
            robot, _hl = arrow_column_pick(
                options, highlight=highlight, show_key=False,
                left_is_back=True, erase_on_back=True, erase_on_back_extra=header_count,
            )
            if robot is _NAV_BACK:
                _erase_dispatch_accum()
                raise _PhcalBackToMenu()
        else:
            robot = arrow_column_pick(options, highlight=highlight, show_key=False)
        _accum_lines(own_lines)
        _frame_close()
        if robot == "*":
            return "both"
        return robot
    choices = {"1", "2"}
    labels = {"1": "Robot 1", "2": "Robot 2"}
    resolved_default = default if default in choices else sorted(choices)[0]
    robot = _prompt_pick(
        choices, default=default, exit_on_invalid=True, labels=labels,
        back=back, question=f"ENTER for {labels[resolved_default]}, or change:",
    )
    return robot


def _prompt_animation_token(back=True):
    """Dedicated animation-token picker - its own process, single purpose.
    2026-08-15: pulled out of the animation primitive's own inline block so
    it's named and testable on its own - NOT merged with the primitive
    menu's own print/pick logic above (run_guided_flow's own loop over
    _PRIMITIVE_MENU) even though the shape looks similar - kept decoupled
    on purpose, not a shared generic menu helper.

    "*" = all three in sequence (2026-08-15, PHCAL_ANIMATION_ALL_SEQUENCE_
    001.md, re-keyed to "*" 2026-08-21, PHCAL_NAV_BASE_RULES_FIXES_001.md:
    "*" is the tool-wide all-flag now). Returns the literal string "all";
    the caller (run_guided_flow's "animation" branch) is what actually
    loops over the three tokens - this function stays a pure picker, no
    firing logic here.

    2026-08-22, PHCAL_ARROW_NAV_BUILD_PLAN_002.md Phase 1: "0" (clean
    exit) is REMOVED - no typed-input path remains anywhere in this file,
    and ESC is now the only way to exit from this prompt.

    2026-08-25, PHCAL_ARROW_NAV_BUILD_PLAN_006.md Lane 1 (robot-first
    order): `back`, default True. The old `question="pick a token"` header
    is dropped - dead weight now that the robot has already been picked
    and the operator is mid-primitive, not deciding what kind of prompt
    this even is.

    2026-08-25, PHCAL_ARROW_NAV_BUILD_PLAN_006.md Lane 3 (left-back to the
    ends): the animation dispatch branch now passes `back=True` here too,
    even though this runs SECOND (after `_prompt_robot(back=True)`) - safe
    now that `_prompt_pick()`'s own `_erase_dispatch_accum()` call erases
    the robot pick's already-committed rows along with this one, design
    (b) from the Lane 3 survey (see _DISPATCH_ERASE_ACCUM's own comment)."""
    choices = set(_ANIMATION_TOKEN_MENU) | {"*"}
    labels = {"*": f"all in sequence ({', '.join(_ANIMATION_TOKEN_MENU[k] for k in sorted(_ANIMATION_TOKEN_MENU))})"}
    labels.update(_ANIMATION_TOKEN_MENU)
    token_choice = _prompt_pick(choices, exit_on_invalid=True, labels=labels, back=back)
    if token_choice == "*":
        return "all"
    return _ANIMATION_TOKEN_MENU[token_choice]


def _clamp_numeric(label, val, min_value, max_value):
    """THE one clamp helper, PHCAL_ARROW_NAV_BUILD_PLAN_006.md Lane 1b -
    shared by both the IN (loaded last_value) and OUT (freshly-typed value)
    paths inside `_prompt_value()` below, so a clamp is never duplicated
    per field and is always reported the same way
    (`PHCAL_GUIDED_CLAMPED {label} {val} -> {clamped} (valid range)`), never
    silent. `val` unchanged (and nothing printed) when it already sits
    inside `[min_value, max_value]` or when a given bound is `None`."""
    clamped = val
    if min_value is not None and clamped < min_value:
        clamped = min_value
    if max_value is not None and clamped > max_value:
        clamped = max_value
    if clamped != val:
        print(f"PHCAL_GUIDED_CLAMPED {label} {val} -> {clamped} (valid range)")
    return clamped


def _clamp_text(label, val, max_length):
    """Text-field counterpart to _clamp_numeric() above, PHCAL_ARROW_NAV_
    BUILD_PLAN_006.md Lane 3 follow-up (announce phrase character limit) -
    same IN/OUT, never-silent shape, trims to `max_length` characters
    instead of clamping a number. Reports lengths, not the full string
    (a long phrase in the print line would be noisy); `val` unchanged (and
    nothing printed) when it already fits or `max_length`/`val` is None."""
    if max_length is None or val is None or len(val) <= max_length:
        return val
    trimmed = val[:max_length]
    print(f"PHCAL_GUIDED_CLAMPED {label} {len(val)} chars -> {max_length} chars (trimmed)")
    return trimmed


def _prompt_value(label, last_value, kind="text", min_value=None, max_value=None, display=None, max_length=None):
    """THE typed-value handler, PHCAL_ARROW_NAV_BUILD_PLAN_006.md Lane 1b -
    replaces the old `parse_fn`-based design (every caller used to pass its
    own `lambda raw: _parse_int_flag(...)`/`_parse_float_flag(...)`) with
    per-field bounds read directly from PHCAL_ARROW_NAV_BUILD_PLAN_006.md's
    own field-by-field survey table, so this ONE function can enforce them
    the new way decision 3 asks for: CLAMP, not refuse-and-reprompt.

    `kind`: "int" | "float" | "text". Enter (bare, empty input) always
    returns `last_value` unchanged, for every kind. A non-numeric value
    typed into an int/float field re-prompts (PHCAL_GUIDED_INVALID) rather
    than crashing - unchanged from before this pass.

    `min_value`/`max_value`, when given, are CONCRETE numeric bounds found
    in code (e.g. volume's 1-5, move_reverse's MOVE_REVERSE_MAX_HOLD_
    SECONDS, sleep_wake's SLEEP_MIN_HOLD_SECONDS floor) - a typed value
    outside them is silently corrected to the nearer bound
    (PHCAL_GUIDED_CLAMPED prints what changed), not refused. This is a
    deliberate behavior change from every prior pass here (which refused
    and re-prompted via `_parse_int_flag`/`_parse_float_flag`'s own
    `raise ValueError` - those two functions are UNCHANGED and still used
    exactly that way by main()'s own direct `--flag` CLI parsing, which
    stays refuse-based on purpose - a scripted/argv caller typing a bad
    flag should still fail loudly, not have its mistake silently
    corrected).

    A float field with NEITHER bound given still enforces the one
    universal rule every such field has always had (`_parse_float_flag`'s
    own default `min_exclusive=0.0`): a value <= 0 re-prompts, not clamps
    - there is no CONCRETE floor number to clamp UP to for "hold between
    reps" (arm/nod), animation's hold, or brobots_stay_in_place's hold,
    per PHCAL_ARROW_NAV_BUILD_PLAN_006.md's own survey (flagged there as a
    real gap, not invented here as a fabricated number). A field that DOES
    pass `min_value` (sleep_wake's wait) skips this universal check
    entirely - any value, however low, clamps up to that concrete floor
    instead.

    No precision (decimal-places) rounding is applied to any float field -
    PHCAL_ARROW_NAV_BUILD_PLAN_006.md's own survey found no coded
    precision rule for any of them; inventing a 2dp convention here would
    be exactly the fabrication CLAUDE.md's engineering discipline forbids.

    2026-08-25, PHCAL_ARROW_NAV_BUILD_PLAN_006.md Lane 3 fix pass: this
    field now reads via `_read_raw_line()` (raw mode, one keystroke at a
    time), REPLACING the old cooked-mode `input()` call - a real live trap
    forced this: the old version only ever saw ESC/Left as literal bytes
    INSIDE the typed line, unresolved until Enter was ALSO pressed. Hit
    live at the sleep_wake `wait` field - Left then Esc did nothing,
    Ctrl-C was the only way out. ESC now raises `_PhcalEscExit()`
    immediately (full quit, same as every other prompt - never a partial
    back-out). Left now raises `_PhcalBackToMenu()` immediately too, but
    ONLY while nothing's been typed yet (see `_read_raw_line()`'s own
    docstring) - `_prompt_value()` erases its own 2 lines plus everything
    earlier in the chain (`_erase_dispatch_accum()`) before letting that
    exception propagate, same unwind discipline every other back-enabled
    prompt in this file already follows.

    Prints `_NAV_LINE_VALUE` once per call (not re-printed on an
    invalid-input retry loop, matching how PHCAL_GUIDED_INVALID's own
    retry never reprints the question line either).

    2026-08-25, PHCAL_LEAF_PROMPT_STYLE_UNIFY_001.md (carried forward,
    unchanged by this pass): "ENTER for <default>, or change:" wording,
    `label` prefix kept (multi-field sequences like arm's reps/hold/speed
    would otherwise read as identical unlabeled lines). `display`, if
    given, still overrides the shown value without changing what bare
    Enter actually returns - weather's own `location` field uses this
    (last_value=None, display="windsor ontario canada") so the prompt
    reads human-friendly while bare Enter still passes `None` through,
    exactly matching this primitive's own pre-existing geocode-safety
    behavior (see that call site's own comment).

    2026-08-25, follow-up pass (arm/nod "hold between reps" clamp): the
    clamp now also runs on `last_value` itself, BEFORE it's shown or
    returned on bare Enter - not just on freshly-typed input. This is the
    IN/load-side half of the clamp (a wild value sitting in phcal_last.json,
    e.g. hand-edited or left over from before this range existed, must
    present already-clamped, not get echoed straight through on bare
    Enter). Routed through the same `_clamp_numeric()` helper the typed-
    input path below uses - one clamp mechanism, two call points, per
    PHCAL_ARROW_NAV_BUILD_PLAN_006.md Lane 1b's "do not duplicate clamp
    logic per field" instruction.

    2026-08-25, PHCAL_ARROW_NAV_BUILD_PLAN_006.md Lane 3 (left-back to the
    ends): this field's own drawn lines are accumulated via _accum_lines()
    on every successful resolution (bare Enter, a valid typed value), so a
    LATER pick in the same chain (the "run brobots_wake first?" gate,
    which fires after arm/nod's own reps/hold/speed value fields) can
    erase this field's rows too when IT backs out - see
    _DISPATCH_ERASE_ACCUM's own comment.

    2026-08-26, PHCAL_SELF_REPORTING_DRAW_EXECUTED_001.md: was a hand-typed
    constant (4, or 2 before the LOOK PASS's own frame) that under-counted
    the one real gap that same comment used to flag here - an invalid-
    input retry loop (PHCAL_GUIDED_INVALID) prints MORE lines than the
    fixed header alone (another `_read_raw_line()` prompt line, another
    invalid-message line, per retry), never reflected in that constant.
    Fixed by tracking `drawn` as a running total, incremented by the REAL
    number of lines each part of this loop actually prints - the header
    via `_draw_lines()`'s own return value, each attempt's own input line
    (always exactly 1, whether it succeeds or is about to raise
    _PhcalBackToMenu - `_read_raw_line()` always calls `print()` first
    either way), and each invalid-message line. No more approximation on
    any path through this function.

    2026-08-25, PHCAL_ARROW_NAV_BUILD_PLAN_006.md Lane 3 follow-up
    (announce phrase character limit): `max_length`, optional, `text`-kind
    only - same IN/OUT clamp discipline as `min_value`/`max_value` above,
    routed through `_clamp_text()` instead of `_clamp_numeric()`. A typed
    phrase over the cap is trimmed, not sent raw, on the way out; a
    `last_value` over the cap is trimmed before it's shown or returned on
    bare Enter, on the way in - same two-call-point shape, no duplicated
    clamp logic."""
    if kind in ("int", "float") and (min_value is not None or max_value is not None):
        last_value = _clamp_numeric(label, last_value, min_value, max_value)
    if kind == "text" and max_length is not None:
        last_value = _clamp_text(label, last_value, max_length)
    shown = display if display is not None else last_value
    drawn = _frame_open() + _draw_lines([_NAV_DIVIDER, _NAV_HINT_VALUE, _NAV_DIVIDER])
    while True:
        try:
            raw = _read_raw_line(f"{label} - ENTER for {shown}, or change: ")
        except _PhcalBackToMenu:
            _erase_screen_lines(drawn + 1)
            _erase_dispatch_accum()
            raise
        drawn += 1
        raw = raw.strip()
        if raw == "":
            _accum_lines(drawn)
            _frame_close()
            return last_value
        if kind == "text":
            _accum_lines(drawn)
            _frame_close()
            return _clamp_text(label, raw, max_length)
        try:
            val = int(raw) if kind == "int" else float(raw)
        except ValueError:
            print(f"PHCAL_GUIDED_INVALID {label} must be a number, got {raw!r}")
            drawn += 1
            continue
        if kind == "float" and min_value is None and val <= 0:
            print(f"PHCAL_GUIDED_INVALID {label} must be greater than 0, got {val}")
            drawn += 1
            continue
        _accum_lines(drawn)
        _frame_close()
        return _clamp_numeric(label, val, min_value, max_value)


_PRIMITIVE_MENU = {
    "1": "arm",
    "2": "nod",
    "3": "rattle",
    "4": "weather",
    "5": "animation",
    "6": "move_reverse",
    "7": "robot_info",
    "8": "brobots_stay_in_place",
    "9": "brobots_sleep_to_wake_direct_sdk",
    "10": "brobots_session_responsiveness",
    "11": "brobots_announce_in_sync",
    "12": "tempo",
    "13": "danger",
    "14": "cube",
}
_MOVE_REVERSE_MENU_SUFFIX = " (ON-CHARGER reverse pulse)"

_BROBOTS_WAKE_CHAIN_ELIGIBLE = {"arm", "nod", "brobots_stay_in_place", "animation", "move_reverse"}
_BROBOTS_WAKE_CHAIN_DEFAULT = {"move_reverse": "y"}  # everything else in the eligible set defaults "n"

def _brobots_label(text):
    """Prefix 'Brobots ' onto label/identity text, unless that text
    already starts with the word 'brobots' (any case, space- or
    underscore-joined) - then just its leading letter is capitalized
    instead of a second 'Brobots' word being glued on front. One rule,
    used everywhere this prefix is applied, so existing text that already
    reads as brobots-related never doubles up."""
    if text[:7].lower() == "brobots":
        return "Brobots" + text[7:]
    return f"Brobots {text}"


_PRIMITIVE_GROUPS = [
    ("Active Brobots", ["robot_info"]),
    ("Audio Play (rattle/danger)", ["rattle", "danger"]),
    ("Audio Say (weather/announce)", ["weather", "brobots_announce_in_sync"]),
    ("Animations (kgsuccess/searching/answering)", ["animation"]),
    ("Moves (arm/nod/reverse/stay)", ["arm", "nod", "move_reverse", "brobots_stay_in_place"]),
    ("Vector's cube (colour control)", ["cube"], "divider"),
    ("Wake / Ready (wake/responsiveness)", ["brobots_sleep_to_wake_direct_sdk", "brobots_session_responsiveness"]),
    ("Song Tempo (pacing)", ["tempo"]),
]


_NON_ROBOT_PRIMITIVES = {"tempo"}

_PRIMITIVE_REQUIRED_N = {
    "tempo": 1,
    "robot_info": 2,
    "brobots_announce_in_sync": 2,
    "weather": 2,
    "animation": 2,
    "arm": 2,
    "nod": 2,
    "move_reverse": 2,
    "brobots_stay_in_place": 2,
    "rattle": 2,
    "danger": 2,
    "cube": 2,
    "brobots_session_responsiveness": 2,
    "brobots_sleep_to_wake_direct_sdk": 2,
}


def _session_mode_n():
    """2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md Lane (i): the current
    session mode as a NUMBER on the required-N ladder (1/2/3), not the
    three-value string `_SESSION_MODE` already carries - the one thing
    `_row_enabled()` below compares a primitive's own required-N against.

    Deliberately reproduces _no_confirmed_robot_this_session()'s full
    logic, not just the literal `_SESSION_MODE` mapping, so this lane is
    net-behavior-neutral: `_SESSION_MODE == "single"` normally maps to 2,
    but a single-mode session where the one present robot was FORCED past
    detection and never actually confirmed present
    (`_PRESENT_ROBOTS[0]["present"] is False`, set by
    `_confirm_multi_mode()`'s own override path) drops back to 1, exactly
    matching what `_no_confirmed_robot_this_session()` already returns for
    that case today. That forced-absent path is currently unreachable
    through the UI (the mode-picker's row source no longer offers a
    not-yet-confirmed candidate as a choice - see
    PHCAL_ARROW_NAV_BUILD_PLAN_005.md's own Step 1 reconcile) but the
    protection is kept here rather than silently dropped, in case that
    path is ever reopened.

    `_SESSION_MODE is None` (the direct-flag CLI path, main(), never calls
    run_guided_flow() and never resolves a mode) maps to 3 - not a new
    default, every existing `_SESSION_MODE`-aware check in this file
    already treats None as "not none, not single," i.e. multi-equivalent,
    by simply falling through every other branch's own else. This makes
    that existing, implicit fallback explicit instead of leaving a fourth
    case unstated."""
    if _SESSION_MODE is None:
        return 3
    if _SESSION_MODE == "none":
        return 1
    if _SESSION_MODE == "single":
        if _PRESENT_ROBOTS and not _PRESENT_ROBOTS[0].get("present", True):
            return 1
        return 2
    return 3  # "multi"


def _session_mode_label():
    """2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md Lane (iii): the
    header's own "session mode: N (word)" text - pairs _session_mode_n()'s
    number (what every row's own "Brobots mode N" label is compared
    against) with the plain three-value word (_SESSION_MODE itself) so the
    header stays legible without the operator having to remember which
    number means what. `_SESSION_MODE is None` (the direct-flag CLI path)
    never reaches _run_guided_flow_once() at all, so this only exists to
    keep the function total rather than assuming a live caller - matches
    _session_mode_n()'s own None -> 3 (multi-equivalent) handling."""
    word = _SESSION_MODE if _SESSION_MODE is not None else "multi"
    return f"{_session_mode_n()} ({word})"


def _row_enabled(primitive):
    """2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md Lane (i): THE one
    comparison - a row/primitive is enabled when the current session mode
    (as a number) is at least its own required-N. This is the single
    source of truth both the display (_is_none_mode_disabled() below, for
    now - Lane (iii) widens its own callers) and, from Lane (ii) onward,
    every dispatch branch's own fire-gate read from - killing the drift
    where the label came from one function and enforcement (where it
    existed at all, for only 2 of 14 primitives) came from a second,
    separately hand-written check."""
    return _session_mode_n() >= _PRIMITIVE_REQUIRED_N[primitive]


def _row_disabled_fire_null(primitive):
    """2026-08-25, PHCAL_ARROW_NAV_BUILD_PLAN_006.md Lane 2: replaces the
    old _row_disabled_skip() contract (renamed, not just re-worded - the
    old name described a SKIP, first line of the branch, before any
    prompt ran; this one is reached only once a disabled leaf's own full
    walk - robot pick, values, choices, exactly as an enabled leaf shows
    them - is already done, immediately before the point that would
    otherwise fire a real connection/SDK call). Silent null: no live
    fire, no state write (_save_last never runs on this path), no
    PHCAL_ISOLATE/PHCAL_LAST_SAVED log line - just this one newcomer-
    plain confirmation that nothing happened. "Car on, engine off" - same
    intent as the function this replaces, now reached at the fire point
    instead of the top of the branch."""
    print(f"Not available in this mode — nothing fired. (needs mode {_PRIMITIVE_REQUIRED_N[primitive]}+, current mode {_session_mode_n()})")
    return 1


def _no_confirmed_robot_this_session():
    """2026-08-24, operator sanity-check request: true when nothing in
    this session is actually confirmed reachable - either detect-first
    found zero robots (_SESSION_MODE == "none"), or the operator forced
    single mode onto a candidate detect-first did NOT find present
    (_confirm_multi_mode()'s 2026-08-23 override capability - that row
    exists precisely so a flaky/failed detection doesn't lock the
    operator out of trying, but picking it doesn't confirm the robot is
    actually there; _PRESENT_ROBOTS[0]["present"] is False for exactly
    this case, set by _confirm_multi_mode() itself). Multi mode is never
    affected - it only ever triggers when detect-first found 2+ robots
    genuinely present, so _PRESENT_ROBOTS always holds real, confirmed
    entries there. Live-caught regression: forcing single mode on a
    NOT_PRESENT robot left every primitive showing enabled, and firing
    `robot_info` against it took a real 3-second network timeout
    ("no route to host") before failing - exactly the doomed-connection
    cost the disabled label and the two enforcing dispatch branches below
    exist to avoid, just for a case this file couldn't reach before the
    override existed.

    2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md Lane (ii): no longer
    called anywhere - `_is_none_mode_disabled()` (Lane i) and both
    dispatch branches that used to call this directly (`robot_info`,
    `brobots_announce_in_sync`) have all migrated to `_row_enabled()`,
    which reproduces this function's own logic internally via
    `_session_mode_n()`. Left in place, unused, rather than pruned - not
    this lane's scope."""
    if _SESSION_MODE == "none":
        return True
    if _SESSION_MODE == "single" and _PRESENT_ROBOTS and not _PRESENT_ROBOTS[0].get("present", True):
        return True
    return False


def _is_none_mode_disabled(primitive):
    """2026-08-20, PHCAL_NAV_FIXES_001.md (finding 2's proposed approach,
    PHCAL_NAV_LIVETEST_FINDINGS_001.md): true when a control can't do
    anything with no robot detected - lets a caller decide, before
    picking, which options to flag. Deliberately just a fact, not a block
    on the pick itself - arrow_column_pick()/navigate() stay
    robot-state-agnostic on purpose (a layering call: those are meant to
    be reused by non-phcal callers later, e.g. the future composition
    editor, NAV_PRIMITIVE_BUILT_001.md) - a disabled option is still
    selectable here; each dispatch branch's own none-mode guard
    (robot_info/brobots_announce_in_sync already had one; the rest gain
    the same pattern this same pass) is what actually stops a pick from
    attempting a doomed connection, not this function.

    2026-08-23, row-format build request: renamed from
    _none_mode_disabled_marker() and changed from "return a label-suffix
    string" to "return a bool" - disabled-ness now needs to be known
    BEFORE final label text is assembled, since _build_primitive_group_tree()
    below uses it both to prefix "[disabled] " at the front of a row (never
    a trailing suffix anymore) and to sort disabled rows first.

    2026-08-24 (PHCAL_ARROW_NAV_BUILD_PLAN_005.md Lane (i)): now backed by
    _row_enabled() - the required-N ladder's one comparison - instead of
    _no_confirmed_robot_this_session()/a literal `_SESSION_MODE == "none"`
    check. The old `and primitive not in _NON_ROBOT_PRIMITIVES` clause is
    dropped, not carried forward redundantly - `_PRIMITIVE_REQUIRED_N`
    already encodes "tempo never needs a robot" as required-N=1, which
    `_row_enabled()` alone fully answers; keeping a second, separate
    `_NON_ROBOT_PRIMITIVES` check here would be exactly the two-sources-
    of-the-same-fact drift this lane exists to remove. Verified
    net-behavior-neutral against the old implementation for all four
    `_SESSION_MODE` values (None/"none"/"single"/"multi"), including the
    currently-unreachable forced-single-mode case -
    `_session_mode_n()`'s own docstring carries that equivalence, not
    re-derived here. `_no_confirmed_robot_this_session()` itself is
    UNTOUCHED by this lane - `robot_info`/`brobots_announce_in_sync`'s own
    dispatch branches still call it directly; migrating those two (and
    adding the same gate to the other 12) is Lane (ii)'s own scope, not
    this one's."""
    return not _row_enabled(primitive)


def _submenu_control_note(primitive):
    """Per-control note printed next to a control inside one of the 4 new
    multi-member group sub-menus (moves/audio/say/init) -
    carries forward the exact brobots_session_responsiveness chain-
    eligibility fact the old flat main-menu's items 2/3 used to spell out
    as two long sentences, now attached to the control's own line instead
    so the main menu can stay short (operator direction, PHCAL_MENU_
    REGROUP_BUILT). Derived from the same _BROBOTS_WAKE_CHAIN_ELIGIBLE set
    everywhere else already uses - PHCAL_CLEAN_EXIT_AND_NOTES_001.md's own
    self-updating convention, kept, so a future primitive still can't drift
    this stale. Only the genuine special cases are hardcoded:
    brobots_session_responsiveness IS the toggle, not a target of it;
    move_reverse keeps its own ON-CHARGER safety note. (2026-08-20,
    PHCAL_NAV_FIXES_001.md: weather used to be hardcoded here too, as "no
    robot control at all" - wrong, corrected; weather is robot-control, see
    _NON_ROBOT_PRIMITIVES above, and now falls through to the same
    chain-eligibility branches as any other robot-control primitive.) Only
    called for the 10 primitives that land in a multi-member group - the 4
    single-member direct groups (info/cube/animations/tempo) never reach
    this, so their own disabled-ness is checked separately in
    _build_primitive_group_tree() below.

    2026-08-23, row-format build request: no longer appends the none-mode
    marker itself (that's centralized in _build_primitive_group_tree() now,
    as a label PREFIX rather than a per-note suffix), and drops the
    POSITIVE "brobots_session_responsiveness chain-eligible" tag entirely
    per that same request - a genuinely disqualifying/real caveat ("not
    chain-eligible - different control channel", move_reverse's ON-CHARGER
    note, weather's own robot-control note, session_responsiveness's own
    "chain toggle itself" note) still prints; a merely-positive eligibility
    claim does not."""
    if primitive == "brobots_session_responsiveness":
        return " (the chain toggle itself - global wake note, no movement)"
    if primitive == "weather":
        return " (robot-control - speaks the forecast through a robot; not chain-eligible)"
    if primitive == "move_reverse":
        return _MOVE_REVERSE_MENU_SUFFIX
    if primitive in _BROBOTS_WAKE_CHAIN_ELIGIBLE:
        return ""
    return " (not chain-eligible - different control channel)"


def _sort_menu_rows(rows):
    """2026-08-23, row-format build request: shared sort for both the root
    group menu and each multi-member submenu - disabled rows first, then
    enabled, alphabetical (case-insensitive) within each block. `rows` is
    a list of (label, payload, disabled, sort_key) tuples; payload is
    carried through unchanged (a primitive name for a leaf, or a nested
    dict for a submenu) - this function only reorders, never rewrites
    payload or label. Kept as its own function, not inlined, since both
    call sites in _build_primitive_group_tree() below need the identical
    rule.

    sort_key is the raw, un-prefixed name (e.g. "weather",
    "brobots_announce_in_sync"), NOT the final rendered label - sorting by
    the rendered label was tried first and got the order wrong: some
    labels read "Brobots weather" (space after the prefix, since "weather"
    didn't already start with the word) while others read
    "Brobots_announce_in_sync" (underscore, since that identity string
    already started with "brobots_" and _brobots_label() only capitalizes
    it rather than double-prefixing) - space (ASCII 32) sorts before
    underscore (ASCII 95), so "weather" landed ahead of
    "announce_in_sync"/"stay_in_place" even though a person alphabetizing
    by the actual name would expect the opposite. Sorting on the
    pre-prefix name sidesteps that punctuation artifact entirely."""
    disabled_rows = sorted((r for r in rows if r[2]), key=lambda r: r[3].lower())
    enabled_rows = sorted((r for r in rows if not r[2]), key=lambda r: r[3].lower())
    return disabled_rows + enabled_rows


def _build_primitive_group_tree():
    """2026-08-19, NAV_PRIMITIVE_BUILT_001.md: builds the navigate()-shaped
    tree from _PRIMITIVE_GROUPS - the ONE real navigate() call site this
    pass creates (_run_guided_flow_once()'s group/sub menu; every other
    prompt in this file is unchanged). A 1-member group's own tree value
    IS the primitive name directly (a leaf, no sub-menu ever drawn) -
    matching the original hand-rolled loop's own "1-member group routes
    straight through" behavior exactly. A multi-member group's tree value
    is a nested dict of numbered sub_keys -> (member label + its
    _submenu_control_note, member name) - same sub-menu content
    _PRIMITIVE_GROUPS + _submenu_control_note already produced before this
    port, just shaped for navigate() instead of built inline by hand each
    call.

    2026-08-23, row-format build request: every submenu member label now
    gets a flat "Brobots " prefix (root group labels already carry their
    own final prefix/naming in _PRIMITIVE_GROUPS itself); a disabled
    control gets "[disabled] " prefixed once, at the very front, ahead of
    that "Brobots " prefix - never a trailing suffix anymore. Row order
    within each submenu is reassigned via _sort_menu_rows() (disabled-first,
    then alphabetical) - the dict keys below are freshly numbered "1".."N"
    in that final sorted order, so navigate()'s own generic
    `sorted(node, key=lambda k: (len(k), k))` walk produces the sorted
    order as a side effect, with no change to navigate()'s own
    robot-state-agnostic sorting logic.

    2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md Lane (iii): the main-menu
    group row's own "[disabled] " prefix is REMOVED outright (both the
    1-member-group case that used to carry it, and the always-False
    multi-member-group case that never did).

    2026-08-26, PHCAL_ARROW_NAV_BUILD_PLAN_006.md LOOK PASS: the "Brobots
    mode N " tag that replaced it (_group_mode_label(), now unused - left
    defined, not pruned, out of this pass's own scope) is ALSO removed
    entirely per direct operator instruction - "prune 100%, no replacement."
    Root row order is now the operator's own hand-specified list order
    (`_PRIMITIVE_GROUPS`, iterated as-is - PURE alphabetical sort retired),
    not `_PRIMITIVE_REQUIRED_N`-derived or alphabetical. A 2-char, fixed,
    evenly-spaced key per group ("10","20",...,"80") makes navigate()'s own
    generic `sorted(node, key=lambda k: (len(k), k))` walk reproduce this
    exact list order as a side effect (same trick the submenu numbering
    below already relied on) - same-length keys collapse that sort to
    plain lexicographic string comparison, fully under this function's own
    control. `"65"` (between "60" and "70") is the one non-group entry - a
    divider row (see _PRIMITIVE_GROUPS's own 3rd-tuple-item comment),
    child=None, non-selectable via navigate()'s new skip_indices wiring.
    Submenu rows are unchanged - still disabled-first via
    _sort_menu_rows(), still "Brobots "-prefixed member labels.

    2026-08-26, PHCAL_LOOK_PASS_CORRECTION_001.md: per the operator's own
    literal PDF spec, every row (root AND submenu) now carries its own
    "00. "/"01. " position prefix, restarting at 00 PER SCREEN (a
    submenu's own numbering starts fresh at 00, independent of its
    parent's row number) - not a global ID, not the group's own required-N,
    not the primitive's own identity string. The divider entry itself is
    never numbered - it isn't a real, selectable row (see navigate()'s own
    skip_indices wiring), so it's excluded from the position count
    entirely (the row after it does NOT skip a number)."""
    tree = {}
    group_index = 0
    root_position = 0
    for entry in _PRIMITIVE_GROUPS:
        group_label, members = entry[0], entry[1]
        group_index += 1
        key = f"{group_index}0"
        numbered_label = f"{root_position:02d}. {group_label}"
        root_position += 1
        if len(members) == 1:
            tree[key] = (numbered_label, members[0])
        else:
            member_rows = []
            for member in members:
                m_disabled = _is_none_mode_disabled(member)
                raw = f"{_brobots_label(member)}{_submenu_control_note(member)}"
                m_label = ("[disabled] " + raw) if m_disabled else raw
                member_rows.append((m_label, member, m_disabled, member))
            member_rows = _sort_menu_rows(member_rows)
            sub = {}
            for i, (m_label, member, _m_disabled, _sort_key) in enumerate(member_rows):
                sub[str(i + 1)] = (f"{i:02d}. {m_label}", member)
            tree[key] = (numbered_label, sub)
        if len(entry) > 2 and entry[2] == "divider":
            divider_key = f"{group_index}5"
            tree[divider_key] = (_NAV_DIVIDER, None)
    return tree


def _group_mode_label(group_label, group_n):
    """2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md Lane (iii): the main
    menu's own required-N ladder, made visible per row - replaces the old
    group-level "[disabled] " prefix (see _build_primitive_group_tree()'s
    own docstring above for why). `group_label` always starts with the
    literal "Brobots " - every _PRIMITIVE_GROUPS value is built through
    _brobots_label() on text that never itself starts with "brobots"
    (confirmed by reading _PRIMITIVE_GROUPS directly: "info (active
    brobots)", "moves (...)", etc. - none lead with the word), so
    _brobots_label() always takes that function's plain `f"Brobots
    {text}"` branch, never its "already-brobots-prefixed" branch. This
    just reinserts "mode N" right after that fixed prefix rather than
    gluing a second "Brobots" word on."""
    rest = group_label[len("Brobots "):] if group_label.startswith("Brobots ") else group_label
    return f"Brobots mode {group_n} {rest}"


def _prompt_brobots_wake_chain(primitive):
    default = _BROBOTS_WAKE_CHAIN_DEFAULT.get(primitive, "y")
    choice = _prompt_pick(
        {"y", "n"}, default=default, labels={"y": "yes", "n": "no"}, question="run brobots_wake first?",
        back=True,
    )
    return choice == "y"


def _run_brobots_wake_no_release(mod, clock, serial, live):
    """The chain toggle's own call: wake WITHOUT releasing, so the
    primitive that runs next picks up the SAME held-control window -
    reuses the exact same core functions cmd_brobots_wake composes
    (conn_test/assume/golden-flag release-pulse), just skips that command's
    own tail release. Mechanism REPLACED 2026-08-10, see
    WHEEL_NUDGE_WAKE_RELEASE_SETTLE_SECONDS's own comment above."""
    wake_check = run_wheel_nudge_wake_check(mod, clock, serial, live)
    assume = run_assume_control(mod, clock, serial, live)
    wake_confirm = run_wheel_nudge_wake_release_pulse(mod, clock, serial, live)
    ok = wake_check["ok"] and assume["ok"] and wake_confirm["ok"]
    return {"ok": ok, "wake_check": wake_check, "assume": assume, "wake_confirm": wake_confirm}


def _resolve_chain_wake(primitive, mod, clock, serial, live):
    """2026-08-21, PHCAL_NAV_CONSOLIDATION_001.md (PHCAL_INPUT_TREE_SURVEY_
    001.md §3 DUPLICATES / §5 step 5): the chain-wake prompt-then-fire
    sequence, pulled out of 4 near-identical inline copies inside
    _run_guided_flow_once() (animation, brobots_stay_in_place, move_reverse,
    the shared arm/nod tail) into one place - same prompt
    (_prompt_brobots_wake_chain), same no-release wake call
    (_run_brobots_wake_no_release), same failure message shape, just
    primitive-parameterized instead of copy-pasted per call site.

    Returns (chain_wake, failed). chain_wake is exactly what the caller
    should pass as its own cmd_*() call's pre_assumed= argument - False
    when the operator declined the chain, or when it was accepted but
    failed (failed=True covers that distinction instead). failed=True
    means the wake was accepted but did not actually succeed - the
    caller's own contract, unchanged from before this refactor, is to
    print nothing further and return 1 immediately in that case (the
    PHCAL_BROBOTS_WAKE_CHAIN_FAILED line below already covers the
    explanation)."""
    chain_wake = _prompt_brobots_wake_chain(primitive)
    if not chain_wake:
        return False, False
    wake = _run_brobots_wake_no_release(mod, clock, serial, live)
    if not wake["ok"]:
        print(f"{clock.prefix()}PHCAL_BROBOTS_WAKE_CHAIN_FAILED - stopping, not firing {primitive} without a confirmed wake")
        return False, True
    return True, False


def _run_guided_flow_once():
    """2026-08-18, PHCAL_NAV_POLISH_001.md: renamed from run_guided_flow()
    (now a thin stay-in-loop wrapper below, the real public entry point) -
    internally unchanged except the group/sub-menu preamble is now its own
    small while/break/continue loop (see its own comment above) and the
    two "operator explicitly chose to leave" return sites return None
    instead of 0, so the wrapper can tell that apart from "a primitive
    dispatch returned rc=0 (success)." Every other one of this function's
    ~15 dispatch-branch returns is untouched - same value, same meaning as
    before this pass.

    Rung 3's guided prompt flow, entered when `phcal` is called with no
    primitive at all. 2026-08-16 (PHCAL_MENU_REGROUP_BUILT): a two-level
    menu now picks `primitive` - a top-level group (8 human-labeled
    categories, _PRIMITIVE_GROUPS) then, for any group with 2-3 members, a
    sub-menu showing the real code identifiers (arm/nod/move_reverse under
    "movements", etc.). A 1-member group (info/cube/animations/tempo)
    routes straight through, same as the old flat menu always did for
    those. Pure display - every one of the 14 identity strings is
    unchanged, so everything below this point (robot -> reps/hold between
    reps/speed (arm, nod), volume (rattle) - each pre-filled from
    phcal_last.json, seeded from the rung-1 bingo defaults on a first run
    with no file yet -> fires exactly like the direct-flag form (live if
    the gate's on, dry if it's off) -> saves the values actually used back
    to the same one memory file) is untouched by this regroup. Weather has
    its own shape since 2026-08-15 (PHCAL_WEATHER_ROBOT_SPEAK_001.md):
    robot pick, then straight into run_robot_control_song_001.py's own
    run_single_note("weather", ...) - the same real fetch + per-robot
    unit/clock speak the awaken song's own weather step and the
    gopod-weather-say alias already use. No reps/hold/speed/volume value to
    remember (nothing about the forecast is tunable), so it never touches
    phcal_last.json.

    2026-08-18 (PHCAL_DETECT_FIRST_001.md): the detect-first probe
    (replaces the old optional "1. info" menu pick) sets this session's
    mode (none/single/multi, module-level _SESSION_MODE/_PRESENT_ROBOTS,
    read by _prompt_robot() and the robot_info/brobots_announce_in_sync
    branches below) and gates on any present robot reading low battery
    (default: do not proceed). The menu itself stays the exact same full
    8-group menu regardless of mode - nothing removed, behavior adapts
    underneath it.

    2026-08-19 refinement (PHCAL_NAV_POLISH_001.md addendum): that probe no
    longer runs HERE - it moved to _resolve_session_mode_once(), called
    exactly ONCE by run_guided_flow() before its stay-in loop starts, so a
    "continue" at the continue-or-exit prompt no longer re-triggers the
    ~1-2s probe or a repeated multi-mode confirm question. This function
    now assumes _SESSION_MODE/_PRESENT_ROBOTS are already resolved by the
    time it's called - it only reads them, same as _prompt_robot() and the
    dispatch branches below already did."""
    global _SESSION_MODE, _PRESENT_ROBOTS, _LAST_MENU_SCREEN_LINES
    global _MENU_PASS_LINE_MARK
    _current_lines = getattr(sys.stdout, "cursor_line", None)
    if _MENU_PASS_LINE_MARK is not None and _current_lines is not None:
        _erase_screen_lines(_current_lines - _MENU_PASS_LINE_MARK)
        _current_lines = getattr(sys.stdout, "cursor_line", None)
    _MENU_PASS_LINE_MARK = _current_lines
    _frame_open()
    print(_NAV_DIVIDER)
    print(_NAV_HINT_ARROW)
    print(_NAV_DIVIDER)
    return navigate(_build_primitive_group_tree(), _dispatch_primitive)


def _dispatch_primitive(primitive):
    """The big if/elif primitive-dispatch chain, extracted 2026-08-24
    (PHCAL_ARROW_NAV_BUILD_PLAN_005.md lane (iii), pinned-menu/left-back
    layout brick) so navigate() can call it directly as its own dispatch
    callback - this is what lets Left inside a leaf picker unwind back to
    navigate()'s own tree-walk instead of falling all the way out of the
    guided flow. Body is byte-for-byte the same ~15-branch dispatch
    _run_guided_flow_once() used to run inline right after its own
    `primitive = navigate(...)` call - nothing about what any primitive
    fires, computes, or saves changed by this extraction, only WHO calls it
    and how a Left-back mid-dispatch is handled (see _PhcalBackToMenu,
    raised by the back-enabled leaf pickers it calls into, caught by
    navigate() itself, not this function - a raise here just propagates
    straight through, same as any other exception a plain function call
    doesn't catch)."""
    global _DISPATCH_ERASE_ACCUM
    _DISPATCH_ERASE_ACCUM = 0
    if primitive == "weather":
        robot = _prompt_robot(back=True)
        location = _prompt_value("location", None, kind="text", display="windsor ontario canada")
        if not _row_enabled(primitive):
            return _row_disabled_fire_null(primitive)
        live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
        control_mod = _load_module(CONTROL_SONG_RUNNER_PATH, "run_robot_control_song_001")
        result = control_mod.run_single_note("weather", live, robot, location=location or None)
        return 0 if result.get("ok") else 1

    if primitive == "tempo":
        return cmd_tempo_calibration()

    if primitive == "brobots_announce_in_sync":
        which = _prompt_robot(default="both", allow_both=True, back=True)
        last_phrase = _load_last("1")["brobots_announce_in_sync"]["phrase"]
        phrase = _prompt_value("phrase", last_phrase, kind="text", max_length=PHCAL_ANNOUNCE_PHRASE_MAX_CHARS)
        if not _row_enabled(primitive):
            return _row_disabled_fire_null(primitive)
        live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
        os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
        clock = _Clock()
        if which in ("1", "2"):
            label = next(c["label"] for c in _candidate_list() if c["which"] == which)
            rc = cmd_brobots_ready_single(clock, which, label, live, phrase)
            last_which = which
        else:
            mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
            rc = cmd_brobots_ready(mod, clock, live, phrase)
            last_which = "both"
        for w in _resolve_last_whichs(last_which):
            _save_last(w, "brobots_announce_in_sync", {"phrase": phrase})
        print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=brobots_announce_in_sync path={LAST_PATH}")
        return rc


    if primitive == "animation":
        robot = _prompt_robot(back=True)
        animation_token = _prompt_animation_token(back=True)
        hold_seconds = _prompt_value("hold", ANIMATION_DEFAULT_HOLD_SECONDS, kind="float", min_value=PHCAL_UNBOUNDED_HOLD_MIN_SECONDS, max_value=PHCAL_UNBOUNDED_HOLD_MAX_SECONDS)
        if not _row_enabled(primitive):
            return _row_disabled_fire_null(primitive)
        live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
        os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
        mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
        serial = BROBOT_2_SERIAL if robot == "2" else BROBOT_1_SERIAL
        clock = _Clock()
        restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
        if not restart_preflight["ok"]:
            return 1
        preflight = run_preflight(mod, clock, serial, live)
        if not preflight["ok"]:
            return 1
        battery = run_battery_check(mod, clock, serial, live)
        if not battery["ok"]:
            print(f"{clock.prefix()}PHCAL_BLOCKED_LOW_BATTERY not firing animation - see PHCAL_BATTERY_CHECK line above")
            if PHCAL_PRIMITIVE_BATTERY_BLOCK_ENABLED:
                return 1
        chain_wake, chain_wake_failed = _resolve_chain_wake("animation", mod, clock, serial, live)
        if chain_wake_failed:
            return 1
        if animation_token == "all":
            fire_order = ["kgSuccess", "searching", "answering"]
            ok = True
            for i, token in enumerate(fire_order):
                rc = cmd_animation(
                    mod, clock, serial, robot, live, token, hold_seconds,
                    pre_assumed=(chain_wake if i == 0 else False),
                )
                ok = ok and rc == 0
                if i < len(fire_order) - 1:
                    if live:
                        print(f"{clock.prefix()}PHCAL_ANIMATION_SEQUENCE_PAUSE settling {_ANIMATION_SEQUENCE_PAUSE_SECONDS}s before next token")
                        time.sleep(_ANIMATION_SEQUENCE_PAUSE_SECONDS)
                    else:
                        print(f"{clock.prefix()}PHCAL DRY: would pause {_ANIMATION_SEQUENCE_PAUSE_SECONDS}s before next token")
            _save_last(robot, "animation", {"hold": hold_seconds})
            print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=animation path={LAST_PATH}")
            return 0 if ok else 1
        rc = cmd_animation(mod, clock, serial, robot, live, animation_token, hold_seconds, pre_assumed=chain_wake)
        _save_last(robot, "animation", {"hold": hold_seconds})
        print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=animation path={LAST_PATH}")
        return rc

    if primitive == "brobots_stay_in_place":
        robot = _prompt_robot(back=True)
        hold_seconds = _prompt_value("hold", HOLD_DEFAULT_SECONDS, kind="float", min_value=PHCAL_UNBOUNDED_HOLD_MIN_SECONDS, max_value=PHCAL_UNBOUNDED_HOLD_MAX_SECONDS)
        if not _row_enabled(primitive):
            return _row_disabled_fire_null(primitive)
        live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
        os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
        mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
        serial = BROBOT_2_SERIAL if robot == "2" else BROBOT_1_SERIAL
        clock = _Clock()
        restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
        if not restart_preflight["ok"]:
            return 1
        preflight = run_preflight(mod, clock, serial, live)
        if not preflight["ok"]:
            return 1
        chain_wake, chain_wake_failed = _resolve_chain_wake("brobots_stay_in_place", mod, clock, serial, live)
        if chain_wake_failed:
            return 1
        rc = cmd_hold(mod, clock, serial, robot, live, hold_seconds, pre_assumed=chain_wake)
        _save_last(robot, "brobots_stay_in_place", {"hold": hold_seconds})
        print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=brobots_stay_in_place path={LAST_PATH}")
        return rc

    if primitive == "brobots_session_responsiveness":
        robot = _prompt_robot(back=True)
        if not _row_enabled(primitive):
            return _row_disabled_fire_null(primitive)
        live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
        os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
        mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
        serial = BROBOT_2_SERIAL if robot == "2" else BROBOT_1_SERIAL
        clock = _Clock()
        restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
        if not restart_preflight["ok"]:
            return 1
        preflight = run_preflight(mod, clock, serial, live)
        if not preflight["ok"]:
            return 1
        return cmd_brobots_wake(mod, clock, serial, robot, live)

    if primitive == "move_reverse":
        robot = _prompt_robot(back=True)
        hold_seconds = _prompt_value(
            "hold", MOVE_REVERSE_DEFAULT_HOLD_SECONDS, kind="float", max_value=MOVE_REVERSE_MAX_HOLD_SECONDS,
        )
        if not _row_enabled(primitive):
            return _row_disabled_fire_null(primitive)
        live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
        os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
        mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
        serial = BROBOT_2_SERIAL if robot == "2" else BROBOT_1_SERIAL
        clock = _Clock()
        restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
        if not restart_preflight["ok"]:
            return 1
        preflight = run_preflight(mod, clock, serial, live)
        if not preflight["ok"]:
            return 1
        battery = run_battery_check(mod, clock, serial, live)
        if not battery["ok"]:
            print(f"{clock.prefix()}PHCAL_BLOCKED_LOW_BATTERY not firing move_reverse - see PHCAL_BATTERY_CHECK line above")
            if PHCAL_PRIMITIVE_BATTERY_BLOCK_ENABLED:
                return 1
        chain_wake, chain_wake_failed = _resolve_chain_wake("move_reverse", mod, clock, serial, live)
        if chain_wake_failed:
            return 1
        rc = cmd_move_reverse(mod, clock, serial, robot, live, hold_seconds, pre_assumed=chain_wake)
        _save_last(robot, "move_reverse", {"hold": hold_seconds})
        print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=move_reverse path={LAST_PATH}")
        return rc

    if primitive == "robot_info":
        which = _prompt_robot(default="both", allow_both=True, back=True)
        if not _row_enabled(primitive):
            return _row_disabled_fire_null(primitive)
        live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
        if not live:
            print("PHCAL_NO_DRY_MODE robot_info has no simulate path (no --dry exists in the binary) - re-run with GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 to fire it")
            return 1
        os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
        mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
        clock = _Clock()
        return cmd_robot_info(mod, clock, which)

    if primitive == "cube":
        robot = _prompt_robot(default="2", back=True)
        if not _row_enabled(primitive):
            return _row_disabled_fire_null(primitive)
        live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
        os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
        mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
        serial = BROBOT_2_SERIAL if robot == "2" else BROBOT_1_SERIAL
        clock = _Clock()
        restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
        if not restart_preflight["ok"]:
            return 1
        preflight = run_preflight(mod, clock, serial, live)
        if not preflight["ok"]:
            return 1
        return cmd_cube(mod, clock, serial, robot, live)

    if primitive == "brobots_sleep_to_wake_direct_sdk":
        mode_choice = _prompt_pick(
            {"1", "2"}, exit_on_invalid=True,
            labels={
                "1": "after a set time",
                "2": "signaled by a completed process (a real wpr/restart-wirepod check)",
            },
            back=True, question="pick a release mode",
        )
        which = _prompt_robot(default="both", allow_both=True, back=True)
        wait_seconds = None
        if mode_choice == "1":
            wait_seconds = _prompt_value(
                "wait", SLEEP_WAKE_DEFAULT_WAIT_SECONDS, kind="float", min_value=SLEEP_MIN_HOLD_SECONDS,
            )
        if not _row_enabled(primitive):
            return _row_disabled_fire_null(primitive)
        live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
        if not live:
            print("PHCAL_NO_DRY_MODE sleep_wake has no simulate path (no --dry exists in the binary) - re-run with GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 to fire it")
            return 1
        os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
        mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
        clock = _Clock()
        if mode_choice == "1":
            rc = cmd_sleep_wake_set_time(mod, clock, which, wait_seconds)
            for last_which in _resolve_last_whichs(which):
                _save_last(last_which, "brobots_sleep_to_wake_direct_sdk", {"wait": wait_seconds})
            print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=brobots_sleep_to_wake_direct_sdk path={LAST_PATH}")
            return rc
        return cmd_sleep_wake_on_process(mod, clock, live, which)


    robot = _prompt_robot(back=True)

    last = _load_last(robot)[primitive]

    count = None
    cycles = None
    hold_seconds = None
    speed = None
    volume_ui = None

    if primitive == "arm":
        cycles = _prompt_value("reps", last["cycles"], kind="int", min_value=2)
        hold_seconds = _prompt_value("hold between reps", last["hold"], kind="float", min_value=PHCAL_UNBOUNDED_HOLD_MIN_SECONDS, max_value=PHCAL_UNBOUNDED_HOLD_MAX_SECONDS)
        speed = _prompt_value("speed", last["speed"], kind="int", min_value=1)
    elif primitive == "nod":
        count = _prompt_value("reps", last["count"], kind="int", min_value=2)
        hold_seconds = _prompt_value("hold between reps", last["hold"], kind="float", min_value=PHCAL_UNBOUNDED_HOLD_MIN_SECONDS, max_value=PHCAL_UNBOUNDED_HOLD_MAX_SECONDS)
        speed = _prompt_value("speed", last["speed"], kind="int", min_value=1)
    else:  # rattle / danger - both volume-only, same 1-5 scale
        volume_ui = _prompt_value("volume", last["volume"], kind="int", min_value=1, max_value=5)

    if not _row_enabled(primitive):
        return _row_disabled_fire_null(primitive)

    live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
    os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
    mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
    serial = BROBOT_2_SERIAL if robot == "2" else BROBOT_1_SERIAL
    clock = _Clock()

    print(f"{clock.prefix()}PHCAL_ISOLATE primitive={primitive} robot={robot} serial={serial} live={live}")
    if not live:
        print(f"{clock.prefix()}DRY: live robot gate is off (GOPOD_ALLOW_LIVE_ROBOT_SPEECH != 1)")

    restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
    if not restart_preflight["ok"]:
        return 1

    preflight = run_preflight(mod, clock, serial, live)
    if not preflight["ok"]:
        return 1

    if primitive in ("arm", "nod"):
        battery = run_battery_check(mod, clock, serial, live)
        if not battery["ok"]:
            print(f"{clock.prefix()}PHCAL_BLOCKED_LOW_BATTERY not firing {primitive} - see PHCAL_BATTERY_CHECK line above")
            if PHCAL_PRIMITIVE_BATTERY_BLOCK_ENABLED:
                return 1
        chain_wake, chain_wake_failed = _resolve_chain_wake(primitive, mod, clock, serial, live)
        if chain_wake_failed:
            return 1
    else:
        chain_wake = False  # rattle/danger - not eligible, see _BROBOTS_WAKE_CHAIN_ELIGIBLE's own comment

    if primitive == "arm":
        rc = cmd_arm(mod, clock, serial, robot, live, cycles, hold_seconds, speed, pre_assumed=chain_wake)
        _save_last(robot, "arm", {"cycles": cycles, "hold": hold_seconds, "speed": speed})
    elif primitive == "nod":
        rc = cmd_nod(mod, clock, serial, robot, live, count, hold_seconds, speed, pre_assumed=chain_wake)
        _save_last(robot, "nod", {"count": count, "hold": hold_seconds, "speed": speed})
    elif primitive == "rattle":
        rc = cmd_rattle(mod, clock, serial, robot, live, volume_ui)
        _save_last(robot, "rattle", {"volume": volume_ui})
    else:  # danger
        rc = cmd_danger(mod, clock, serial, robot, live, volume_ui)
        _save_last(robot, "danger", {"volume": volume_ui})

    print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive={primitive} path={LAST_PATH}")
    return rc


def _prompt_continue_or_exit():
    """2026-08-18, PHCAL_NAV_POLISH_001.md, operator decision 5: fires after
    EVERY _run_guided_flow_once() pass that wasn't an explicit exit -
    success or failure/blocked alike (e.g. the low-battery gate returning 1
    before firing) - so a blocked run never dumps the operator straight
    out. Routed through _prompt_pick(), so it's arrow-navigable like
    every other y/n in this file.

    2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md redraw-glitch fix: no
    longer tracks its own on-screen height for the next pass's erase - that
    bookkeeping moved to _MENU_PASS_LINE_MARK (a _TeeStdout `cursor_line`
    checkpoint, covering this prompt's own rows AND everything else the
    pass printed, not just this one widget) - see that global's own
    docstring for why the old per-widget version broke under a real live
    fire's variable-height log output."""
    choice = _prompt_pick(
        {"y", "n"}, default="y", labels={"y": "yes", "n": "no"}, question="continue in phcal?",
    )
    return choice == "y"


def _repick_session_mode():
    """2026-08-24, PHCAL_ARROW_NAV_BUILD_PLAN_005.md Lane (v)/(vi): the
    session-mode-changed recompute hook's own "recompute" half - run when
    run_guided_flow() catches _PhcalSessionModeRepick (Lane vi's own arrow-
    up trigger). Reuses _confirm_multi_mode() verbatim - the EXACT same
    screen _resolve_session_mode_once() already shows at startup, fed
    _DETECTED_PRESENT_ROBOTS (THIS invocation's own raw detect-first
    result, untouched by this function - no re-probe, per this lane's own
    ONE-TABLE instruction) and a freshly-recomputed "what would
    detect-first have called this" label, derived purely from that same
    already-detected list's own length (0/1/2+ -> none/single/multi - the
    exact _detect_present_robots() rule, reproduced rather than re-run
    since detect-first itself is a live hardware probe this hook must not
    repeat). Sets the same module globals (_SESSION_MODE/_PRESENT_ROBOTS)
    and the same GOPOD_ALLOW_LIVE_ROBOT_SPEECH env var
    _resolve_session_mode_once() already sets, by the same rule - one
    source of that mapping, reused, not reimplemented a second way.

    Deliberately reads _DETECTED_PRESENT_ROBOTS, NOT _PRESENT_ROBOTS - a
    real gap caught before shipping this lane: _confirm_multi_mode()'s own
    "/" (none) choice legitimately empties _PRESENT_ROBOTS to [], so an
    operator who started this session in none mode and then arrows up to
    repick would see no single/multi rows at all if this read
    _PRESENT_ROBOTS instead - the exact "re-pick among the ALREADY-
    DETECTED present robots" case this lane's own instruction names.
    _DETECTED_PRESENT_ROBOTS is the raw, never-reassigned-after-first-set
    fact this needs (see its own module-level comment).

    Does NOT itself redraw anything - the caller (run_guided_flow()'s own
    _PhcalSessionModeRepick handler) loops straight back into a fresh
    _run_guided_flow_once() pass after this returns, which is what
    actually re-runs _row_enabled() across every row (via a freshly built
    _build_primitive_group_tree()) and redraws in place, no stacking, via
    that pass's own existing _MENU_PASS_LINE_MARK erase mechanism - see
    _PhcalSessionModeRepick's own docstring for the full chain.

    ONE-TABLE model: phcal_last.json (per-robot calibration facts) is
    never read or written here - a mode change only changes which rows
    _row_enabled() reports live, nothing about any robot's own remembered
    tuning. No snapshot/overlay/store logic added - _SESSION_MODE/
    _PRESENT_ROBOTS are simply reassigned, exactly as
    _resolve_session_mode_once() already does at startup."""
    global _SESSION_MODE, _PRESENT_ROBOTS
    _clock = _Clock()
    if not _DETECTED_PRESENT_ROBOTS:
        detected_mode = "none"
    elif len(_DETECTED_PRESENT_ROBOTS) == 1:
        detected_mode = "single"
    else:
        detected_mode = "multi"
    _PRESENT_ROBOTS, _SESSION_MODE = _confirm_multi_mode(_clock, _DETECTED_PRESENT_ROBOTS, detected_mode)
    if _SESSION_MODE == "none":
        os.environ.pop("GOPOD_ALLOW_LIVE_ROBOT_SPEECH", None)
    else:
        os.environ["GOPOD_ALLOW_LIVE_ROBOT_SPEECH"] = "1"
    print(
        f"{_clock.prefix()}PHCAL_SESSION_MODE_REPICK mode={_SESSION_MODE} "
        f"robots={[r['label'] for r in _PRESENT_ROBOTS]}"
    )


def run_guided_flow():
    """2026-08-18, PHCAL_NAV_POLISH_001.md: the real public entry point now -
    a thin stay-in-loop wrapper around _run_guided_flow_once() (the actual
    guided flow, renamed but otherwise unchanged). _PhcalEscExit is caught
    HERE, not inside _run_guided_flow_once() itself, so that function's raw
    ~450-line dispatch body didn't need reindenting into a try block - the
    exception still propagates cleanly up through every intermediate return
    path since none of them catch it, and lands in this one handler.

    None from _run_guided_flow_once() means the operator explicitly chose
    to exit from the menu structure itself - leave immediately, no
    continue-or-exit prompt. (2026-08-22, PHCAL_ARROW_NAV_BUILD_PLAN_002.md
    Phase 1: the "group-menu 0" and "sub-menu invalid-input exit" triggers
    this used to describe are both removed - ESC is currently the only
    live path here, this None-return contract is left in place for later
    phases to wire a real trigger back into.) Any other return (0 or nonzero - a primitive
    dispatch attempted or completed, success or blocked) goes through
    _prompt_continue_or_exit() per operator decision 5 before looping back
    in or leaving for real.

    2026-08-19 addendum: calls _resolve_session_mode_once() exactly once,
    right here, before the loop starts - see that function's own docstring
    for why (this used to live inside _run_guided_flow_once() and re-ran on
    every "continue").

    2026-08-19 bugfix (live-test): the first cut of this addendum kept
    looping when the low-voltage gate declined - session_ok stayed False for
    the life of the invocation, so every pass synthesized a blocked result
    (1) and re-asked continue-or-exit, and answering "y" just looped back to
    ask the same question again with no real menu underneath it. A startup
    decline is an operator "don't run" signal, not a blocked mid-session
    primitive - it now exits cleanly outright, with no continue-or-exit
    prompt at all. Decision 5's own guarantee (a blocked run reaching
    continue-or-exit rather than being dumped straight out) still applies
    to what it always meant: a primitive dispatch attempted mid-session
    inside _run_guided_flow_once() and returning nonzero - untouched below.

    2026-08-19 (PHCAL_PROMPT_UNIFY_001.md): _resolve_session_mode_once()
    itself is now wrapped in the same try/except _PhcalEscExit as the loop
    below. Two reasons, one fix: the multi-mode-confirm prompt inside it
    just got converted to the ESC-raising _prompt_choice() this same pass,
    and the low-voltage gate's own _prompt_choice() call (already
    ESC-raising since the earlier nav-polish pass) was ALSO unprotected
    here - pressing ESC at "proceed anyway?" would have propagated
    _PhcalEscExit straight out of this function uncaught. Neither prompt
    had a live ESC test exercised against this exact call boundary before
    now; both are fixed by the same one-line move.

    2026-08-26, PHCAL_LOOK_PASS_CORRECTION_001.md STOP #1, resolved: the
    operator's final opener order is welcome -> please-see -> detect ->
    mode picker, printed ONCE at session start, never reprinted on a
    later return to the main menu. Both blocks now print right here,
    before `_resolve_session_mode_once()` (which is what actually runs
    `_detect_present_robots()`) - moved out of `_run_guided_flow_once()`'s
    own per-pass header entirely. No separate one-time bookkeeping needed:
    that function's own `_MENU_PASS_LINE_MARK` checkpoint doesn't start
    tracking cursor position until ITS OWN first call, so these two
    blocks - already on screen before that call ever happens - are never
    touched by any later pass's own erase-and-redraw."""
    _frame_open()
    print("Welcome to PLAYHEAD Calibrations (phcal)")
    _frame_close()
    _frame_flush_close()
    _frame_open()
    print("For upcoming please see")
    print("* SLEEP_BENCH_ALIASES_001.md")
    print("* SLEEP_SEGMENT_ALIASES_001.md")
    _frame_close()
    _frame_flush_close()
    try:
        session_ok = _resolve_session_mode_once()
    except _PhcalEscExit:
        _frame_flush_close()
        print("PHCAL_GUIDED_EXIT (ESC) exiting")
        return 0
    if not session_ok:
        return 1
    while True:
        try:
            result = _run_guided_flow_once()
        except _PhcalEscExit:
            _frame_flush_close()
            print("PHCAL_GUIDED_EXIT (ESC) exiting")
            return 0
        except _PhcalSessionModeRepick:
            try:
                _repick_session_mode()
            except _PhcalEscExit:
                _frame_flush_close()
                print("PHCAL_GUIDED_EXIT (ESC) exiting")
                return 0
            continue
        if result is None:
            return 0
        if not _prompt_continue_or_exit():
            return result


def _pha0b_bait_step_ids():
    """00a_brobots_awaken's own live step_id list, dirty-first resolved via
    the same shared resolver cmd_tempo_calibration()'s own Mode B (above)
    already uses - one canonical reader, not a second copy."""
    envelope_mod = _load_module(KNOBS_ENVELOPE_PATH, "knobs_envelope_001")
    song_dir = SONGS_DIR / "zzz_archives" / "00a_brobots_awaken"
    _, envelope = envelope_mod.load_knobs_envelope(str(song_dir))
    step_ids = [s.get("step_id") for s in envelope.get("steps", [])]
    if not step_ids:
        raise RuntimeError(f"no steps found for 00a_brobots_awaken at {song_dir}")
    return step_ids


def _build_bait_point_tree(step_ids):
    """Point A/B as one real navigate() tree instead of a hand-rolled pick
    loop - the ONLY way to get phcal's own proven root-pinned/submenu-
    erase-on-back redraw (the thing standalone phcal already does cleanly,
    which the old hand-rolled version never got). Root level = Point A's
    own choices ("0" = full song, a leaf; "1".."N" = a real step_id, each
    leading to its own Point B submenu). Keys are NUMBERED strings, not the
    step_id text itself - navigate()'s own `sorted(node, key=lambda k:
    (len(k), k))` would alphabetize real step_id keys (e.g. "arm_test"
    before "connect") out of real program order; numbered keys sort
    correctly by this same rule instead (short-numbers-then-long, "1".."9"
    before "10".."16", same trick cmd_tempo_calibration()'s own step-pick
    already relies on, above). Every leaf is a plain (point_a, point_b)
    tuple - navigate()'s own contract only requires "not a dict", not
    literally a string."""
    tree = {"0": ("Full song", (step_ids[0], step_ids[-1]))}
    for a_i, a_id in enumerate(step_ids, start=1):
        point_b_children = {
            str(b_i): (b_id, (a_id, b_id))
            for b_i, b_id in enumerate(step_ids[a_i - 1:], start=1)
        }
        tree[str(a_i)] = (a_id, point_b_children)
    return tree


def _run_one_field(spec, back):
    """Fires the one golden primitive matching `spec["kind"]` - the only
    place `_run_field_chain()` below actually calls into arrow_column_pick/
    _prompt_pick/_prompt_value/_prompt_robot, so adding a new field kind
    later touches exactly one function, not every call site."""
    kind = spec["kind"]
    if kind == "yn":
        return _prompt_pick(
            {"y", "n"}, default=spec.get("default", "n"), labels={"y": "yes", "n": "no"},
            question=spec["question"], back=back,
        )
    if kind == "value":
        return _prompt_value(spec["question"], spec.get("default"), **spec.get("kind_args", {}))
    if kind == "robot":
        return _prompt_robot(default=spec.get("default"), allow_both=True, back=back)
    raise ValueError(f"unknown bait-nav field kind {kind!r}")


def _run_field_chain(fields, back_out_at_start=False):
    """Generic field-chain engine (PLAYHEAD_PHA0B_PHCAL_FULL_DECOUPLE_
    SURVEY_001.md §4 point 2) - ONE reusable walker for "ask N fields in
    order, Left steps back one field, clean redraw" instead of a
    hand-rolled state machine per chain (this file used to carry two,
    identical in shape, different fields - Segment 1's reporter-gap pair
    and Segment 2's phcal-apply/robot-filter/rich-display trio). A field's
    own definition is DATA (see `_run_one_field()`'s own dict shape),
    never control flow - adding, removing, or reordering a field is a
    one-line change to the list passed in, never a rewrite of this
    function.

    Each spec: {"key", "kind": "yn"|"value"|"robot", "question", "default",
    "kind_args" (value only), "skip_if": callable(values-so-far) -> bool,
    "skip_value" (used when skip_if is True - no prompt shown)}.

    `back_out_at_start`: True when this chain runs INSIDE a navigate()
    tree's own `dispatch()` - Left on this chain's own first real field
    then raises _PhcalBackToMenu uncaught, propagating straight out to
    navigate()'s own loop, which resumes the tree exactly where it was
    (built in, no extra code needed here - see navigate()'s own docstring).
    False (default) for a standalone chain with nothing above it to back
    into (e.g. after the one-shot session-mode gate) - that first field
    gets back=False, same terminal "Left is a no-op" convention every
    other true-first-in-chain widget in this file already uses.

    Returns a dict of key -> resolved value once every field is answered."""
    values = {}
    idx = 0
    while idx < len(fields):
        spec = fields[idx]
        if spec.get("skip_if") and spec["skip_if"](values):
            values[spec["key"]] = spec.get("skip_value")
            idx += 1
            continue
        is_first_real = all(
            fields[j].get("skip_if") and fields[j]["skip_if"](values)
            for j in range(idx)
        )
        back = back_out_at_start if is_first_real else True
        try:
            values[spec["key"]] = _run_one_field(spec, back=back)
        except _PhcalBackToMenu:
            idx -= 1
            while idx >= 0 and fields[idx].get("skip_if") and fields[idx]["skip_if"](values):
                idx -= 1
            if idx < 0:
                raise
            continue
        idx += 1
    return values


def _bait_nav_dispatch(point_ab):
    """navigate()'s own `dispatch(leaf)` for the Point A/B tree above -
    called once, with the (point_a, point_b) tuple the operator's tree
    pick resolved to. Runs the rest of song 1's pre-run nav as two field-
    chain segments, split by the one-shot session-mode gate between them
    (a real, non-backable checkpoint - see _run_field_chain()'s own
    docstring). A _PhcalBackToMenu raised by chain 1's own first field
    (reporter-gap y/n) is deliberately NOT caught here - it propagates to
    navigate(), which resumes Point A/B exactly where the operator left
    them, per that function's own built-in contract."""
    point_a, point_b = point_ab
    chain1 = _run_field_chain(
        [
            {
                "key": "gap_yn", "kind": "yn", "default": "n",
                "question": "apply reporter gaps to this selected range?",
            },
            {
                "key": "gap_seconds", "kind": "value", "default": 0.0,
                "question": "reporter gap seconds", "kind_args": {"kind": "float", "min_value": 0.0},
                "skip_if": lambda v: v.get("gap_yn") == "n", "skip_value": 0.0,
            },
        ],
        back_out_at_start=True,
    )

    session_ok = _resolve_session_mode_once()
    if not session_ok:
        return None
    live_gate = 0 if _SESSION_MODE == "none" else 1

    chain2 = _run_field_chain([
        {
            "key": "phcal_apply", "kind": "yn", "default": "y",
            "question": "apply phcal tweaks to this selected range?",
        },
        {"key": "robot_filter", "kind": "robot"},
        {
            "key": "rich_display", "kind": "yn", "default": "y",
            "question": "rich display on console?",
        },
    ])

    return {
        "point_a": point_a,
        "point_b": point_b,
        "reporter_gap_seconds": chain1["gap_seconds"],
        "live_gate": live_gate,
        "phcal_apply_choice": 1 if chain2["phcal_apply"] == "y" else 0,
        "robot_filter": chain2["robot_filter"],
        "rich_display": 1 if chain2["rich_display"] == "y" else 0,
    }


def _run_pha0b_bait_nav_body():
    """The actual bait-nav sequence - see run_pha0b_bait_nav()'s own
    docstring for the overall back-vs-exit design. Split from that
    function only so the top-level ESC catch stays a single, obvious
    wrapper. Deliberately thin: Point A/B is real navigate() tree data
    (_build_bait_point_tree), everything after is _bait_nav_dispatch()'s
    own job - this function just wires the two together."""
    step_ids = _pha0b_bait_step_ids()
    tree = _build_bait_point_tree(step_ids)
    return navigate(tree, _bait_nav_dispatch)


def run_pha0b_bait_nav():
    """Public entry point, dispatched from `python3 phcal_isolate_001.py
    --pha0b-bait-nav` (see this file's own `if __name__ == "__main__":`
    block below) - song 1's (00a_brobots_awaken / "bait") entire pre-run nav,
    natively on phcal's own golden nav primitives (arrow_column_pick /
    _prompt_pick / _prompt_value / _prompt_robot / _resolve_session_mode_once,
    the same _PhcalBackToMenu/_PhcalEscExit two-tier back-vs-exit every other
    phcal screen already uses).

    Scope, per PHA0B_GOLDEN_NAV_RIPREPLACE_AND_SONG1_SCOPE_001.md Part B:
    covers exactly the touchpoints song 1 actually exercises (Point A/B,
    reporter-gap, live/dry, phcal-apply y/n, robot filter, rich-display) -
    NOT speaker-reassign or MODE=SECTIONS rendering, which song 1 never
    reaches (gated out by song identity in brobots.sh). The phcal-apply
    EXECUTION loop (walking qualifying step_ids, shelling to
    phcal_apply_001.py) and the final run_golden_song_001.py dispatch both
    stay exactly where they already live in brobots.sh - this function only
    gathers the y/n choice, never touches a knobs.json itself.

    Back-vs-exit: every Left-driven `_PhcalBackToMenu` this bridge's own
    screens can raise is caught and handled AS AN INTERNAL STATE TRANSITION
    (see _run_pha0b_bait_nav_body()) - none of them escape to the caller.
    Only two outcomes ever reach brobots.sh: one `PHA0B_BAIT_NAV_RESULT`
    line (completed), or `PHA0B_BAIT_NAV_ABORTED` (ESC pressed anywhere, or
    a declined low-voltage gate) - deliberately simpler than phcal's own
    multi-level tree, since this bridge's own flow is linear, not a tree
    with a level above it to pop back out to.

    2026-08-30, PHA0B_GOLDEN_NAV_BAIT_STDOUT_FIX_001.md: brobots.sh's own
    bait branch invokes this mode via `$(...)` command substitution so it
    can capture the one final result line - but EVERY golden-nav primitive
    (arrow_column_pick/_frame_open/_draw_lines/etc.) also writes its screen
    via plain `sys.stdout`, which `$(...)` swallows into that same pipe
    instead of the real terminal. stdin is untouched by `$(...)` (still the
    real tty), so raw-mode key reads still succeeded - the process wasn't
    frozen, it was correctly waiting for input on a screen that was never
    actually drawn where the operator could see it (confirmed live,
    2026-08-30: blank screen, hangs until Ctrl-C, clean exit on Ctrl-C).
    Fix: `sys.stdout` is swapped to `sys.stderr` (still connected to the
    real terminal, untouched by brobots.sh's stdout-only capture) for the
    ENTIRE interactive body below, always restored via `finally` even on
    ESC, and the one true result/abort line is printed through the real,
    restored stdout only once, after the swap ends - `_run_pha0b_bait_nav_
    body()` itself no longer prints either line directly, it returns a
    plain dict (or None for a declined low-voltage gate) instead, so this
    function is the only place either final line can be written."""
    real_stdout = sys.stdout
    sys.stdout = sys.stderr
    esc_aborted = False
    try:
        result = _run_pha0b_bait_nav_body()
    except _PhcalEscExit:
        result = None
        esc_aborted = True
    finally:
        sys.stdout = real_stdout

    if esc_aborted:
        print("PHA0B_BAIT_NAV_ABORTED (ESC) exiting")
        return 1
    if result is None:
        print("PHA0B_BAIT_NAV_ABORTED declined to proceed with a low-battery robot present")
        return 1
    print(
        "PHA0B_BAIT_NAV_RESULT "
        f"point_a={result['point_a']} point_b={result['point_b']} "
        f"reporter_gap_seconds={result['reporter_gap_seconds']} live_gate={result['live_gate']} "
        f"phcal_apply_choice={result['phcal_apply_choice']} "
        f"robot_filter={result['robot_filter']} "
        f"rich_display={result['rich_display']}"
    )
    return 0


PHCAL_RUN_LOG_PATH = Path(__file__).resolve().parent / "phcal_run.log"


class _TeeStdout:
    _CURSOR_UP_RE = re.compile(r"\x1b\[(\d+)A")

    def __init__(self, real_stdout, log_path):
        self._real = real_stdout
        self._log = open(log_path, "w", buffering=1)
        self.cursor_line = 0

    def write(self, data):
        self._real.write(data)
        self._log.write(data)
        self.cursor_line += data.count("\n")
        for match in self._CURSOR_UP_RE.finditer(data):
            self.cursor_line -= int(match.group(1))

    def flush(self):
        self._real.flush()
        self._log.flush()

    def isatty(self):
        return self._real.isatty()


if __name__ == "__main__":
    sys.stdout = _TeeStdout(sys.stdout, PHCAL_RUN_LOG_PATH)
    if len(sys.argv) == 2 and sys.argv[1] == "--promote":
        sys.exit(_promote_tweaks())
    if len(sys.argv) == 2 and sys.argv[1] == "--pha0b-bait-nav":
        try:
            sys.exit(run_pha0b_bait_nav())
        except EOFError:
            print("PHA0B_BAIT_NAV_BLOCKED no interactive input available - run pha0b from a real terminal")
            sys.exit(1)
        except KeyboardInterrupt:
            print("\nCancelled — going back.")
            sys.exit(1)
    if len(sys.argv) == 1:
        try:
            sys.exit(run_guided_flow())
        except EOFError:
            print("PHCAL_GUIDED_BLOCKED no interactive input available - run phcal from a real terminal")
            sys.exit(1)
        except KeyboardInterrupt:
            print("\nCancelled — going back.")
            sys.exit(1)
    sys.exit(main())

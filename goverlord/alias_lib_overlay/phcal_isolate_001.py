#!/usr/bin/env python3
"""phcal Rung 3 - isolate-and-watch, ONE mechanical primitive, one robot, then stop, now with cycles/hold/speed adjustable per call AND a guided prompt flow plus write-back memory on top."""
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
    """Brobot 1/2 ESN resolution, in priority order: phcal_config.json's robots.candidates (new, N-capable - the list a future detect-first pass will probe) if it has at least 2 entries; else the same GOPOD_BROBOT_1_SERIAL/GOPOD_BROBOT_2_SERIAL env vars run_section1_full_live_001.py has always read; else the same hardcoded defaults that module has always fallen back to."""
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
    """Wheel twin of run_move_axis_no_control above - same move -> hold -> stop shape, but move_wheels takes TWO independent speed params (lw/rw), not one, so this is a new function rather than a forced reuse of the single-speed helper."""
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
    """conn_test only - a cheap, real pre-check ahead of assuming control, matching gopod-opening-chord's own separate WIREPOD_CHAIN_WAKE_START conn_test (fired before its own speak() call)."""
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
    """the real, measured fix - see WHEEL_NUDGE_WAKE_RELEASE_SETTLE_SECONDS's own comment above for the full why."""
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
    """Fires assume_behavior_control ONCE, before cmd_arm/cmd_nod's whole cycle loop starts - not per direction/per cycle."""
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
    """Fires GOPOD's own playSound sound (danger-will-robinson.wav) on one robot via the SAME standalone direct-SDK binary cmd_rattle uses above (RATTLE_BIN_PATH - never rattle-specific, just always called with rattle's own WAV path until now) - identical mechanism, different WAV path."""
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
    """Fires the cube connect->all-corners-red->hold->all-corners-green-> release blip via the standalone direct-SDK binary (direct_sdk_cube_blip_001) - connects independently, requests BehaviorControl, connects the cube (ConnectCube), sets all four corner LEDs red then green (2s hold between, baked into the binary's own blipHoldDuration constant), releases, disconnects."""
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
    """md: single-mode degrade of cmd_brobots_ready's own phrase."""
    print(f"{clock.prefix()}PHCAL_BROBOTS_READY_SINGLE robot={which} ({label}) phrase={phrase!r}")
    control_mod = _load_module(CONTROL_SONG_RUNNER_PATH, "run_robot_control_song_001")
    result = control_mod.run_single_note("say_phrase", live, which, phrase=phrase)
    ok = bool(result.get("ok"))
    print(f"{clock.prefix()}PHCAL_BROBOTS_READY_SINGLE_COMPLETE ok={ok}")
    return 0 if ok else 1


def cmd_robot_info(mod, clock, which):
    """."""
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
    """Reuses the exact same direct_sdk_robot_info_001 binary robot_info's own menu item calls (cmd_robot_info above) - one serial at a time, no BehaviorControl, read-only."""
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
    """One candidate -> candidate dict + present/volts."""
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
    """Runs once, at the very top of run_guided_flow(), before the menu draws."""
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
    """an explicit confirm/override step."""
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
    """After detect-first probes: if ANY present robot's battery reads below BATTERY_MIN_VOLTS (the same threshold run_battery_check() already gates individual motor calls on - reused, not a new number), warn plainly (which robot, its actual level) and ask before proceeding."""
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
    """md addendum: the detect-first probe, multi-mode confirm, and low-voltage gate used to live inside _run_guided_flow_once() itself, so every "continue" at the continue-or-exit prompt re-ran the whole ~1-2s probe (and could re-ask the multi-mode confirm question) on every loop pass."""
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
    """Assume/release bench test - lets the operator watch a robot held at priority=high (the same OVERRIDE_BEHAVIORS level every golden-engine note already uses) genuinely stay put for hold_seconds, then resume normal behavior the instant it's released."""
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
    """This is the wake half: conn_test -> assume -> golden-flag release/settle/re-assume pulse -> release."""
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
    """."""
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
    """."""
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
    """Reads phcal_last.json off disk, robot-keyed shape: {"1": {primitive: values}, "2": {primitive: values}}."""
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
    """Standalone promote step for phcal's own master tweaks file - mirrors run_golden_song_001.py's own _maybe_promote_knobs() shape exactly: same default-n y/n prompt, same shutil.copyfile byte-identical copy (no re-serialize), same only-if-source-exists guard."""
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
    """md - raised by the raw-key reader when the operator presses ESC."""


class _PhcalBackToMenu(Exception):
    """md lane (iii): raised by a back-enabled leaf picker (Left pressed - _prompt_pick(back=True), _prompt_robot(), _prompt_animation_token(), _prompt_brobots_wake_chain()) to abandon the CURRENT primitive's dispatch mid-flight, without firing anything, and resume navigate()'s own tree-walk exactly where it was left (same stack depth, same remembered highlight - see navigate()'s own try/except around its dispatch() call, the only place this is ever caught)."""


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
    """md Lane (v): the session-mode-changed recompute hook's own clean-unwind signal - same "propagate through however many nested calls, no sentinel threading" shape as _PhcalEscExit/_PhcalBackToMenu above, scoped to its own concern: "the operator wants to re-pick the session mode, mid-session, without quitting." navigate()/arrow_column_pick() stay session-mode-agnostic (same layering reason _PhcalBackToMenu's own docstring already gives - those primitives are meant to be reused by a future non-phcal caller with no session-mode concept at all), so this unwinds cleanly up through both to run_guided_flow()'s own handler - the one place session mode is ever actually set."""


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
    """One logical key event from stdin, in raw mode."""
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
    """THE self-reporting draw primitive, PHCAL_SELF_REPORTING_DRAW_ - prints each line in `lines` and returns the exact count it drew."""
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
    """Reads one line of typed text in RAW mode, one keystroke at a time, via the same _read_key() reader every arrow-based prompt in this file already uses."""
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
    """md (PHCAL_INPUT_TREE_SURVEY_ 001.md §5 steps 1+3): replaces the old `_prompt_choice()` - the second, independent raw-mode key reader this file used to carry alongside arrow_column_pick()."""
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
    """md / ."""
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
    """md / the reusable generic N-deep tree walker, built on arrow_column_pick() above - this is what lets an operator "arrow anywhere," one walker, any depth, instead of a hand-written while/continue loop copy-pasted per menu level."""
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
    """Dedicated 'which robot?' prompt - its own process, single purpose."""
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
    """Dedicated animation-token picker - its own process, single purpose."""
    choices = set(_ANIMATION_TOKEN_MENU) | {"*"}
    labels = {"*": f"all in sequence ({', '.join(_ANIMATION_TOKEN_MENU[k] for k in sorted(_ANIMATION_TOKEN_MENU))})"}
    labels.update(_ANIMATION_TOKEN_MENU)
    token_choice = _prompt_pick(choices, exit_on_invalid=True, labels=labels, back=back)
    if token_choice == "*":
        return "all"
    return _ANIMATION_TOKEN_MENU[token_choice]


def _clamp_numeric(label, val, min_value, max_value):
    """THE one clamp helper, Lane 1b - shared by both the IN (loaded last_value) and OUT (freshly-typed value) paths inside `_prompt_value()` below, so a clamp is never duplicated per field and is always reported the same way (`PHCAL_GUIDED_CLAMPED {label} {val} -> {clamped} (valid range)`), never silent."""
    clamped = val
    if min_value is not None and clamped < min_value:
        clamped = min_value
    if max_value is not None and clamped > max_value:
        clamped = max_value
    if clamped != val:
        print(f"PHCAL_GUIDED_CLAMPED {label} {val} -> {clamped} (valid range)")
    return clamped


def _clamp_text(label, val, max_length):
    """Text-field counterpart to _clamp_numeric() above, PHCAL_ARROW_NAV_ Lane 3 follow-up (announce phrase character limit) - same IN/OUT, never-silent shape, trims to `max_length` characters instead of clamping a number."""
    if max_length is None or val is None or len(val) <= max_length:
        return val
    trimmed = val[:max_length]
    print(f"PHCAL_GUIDED_CLAMPED {label} {len(val)} chars -> {max_length} chars (trimmed)")
    return trimmed


def _prompt_value(label, last_value, kind="text", min_value=None, max_value=None, display=None, max_length=None):
    """THE typed-value handler, Lane 1b - replaces the old `parse_fn`-based design (every caller used to pass its own `lambda raw: _parse_int_flag(...)`/`_parse_float_flag(...)`) with per-field bounds read directly from 's own field-by-field survey table, so this ONE function can enforce them the new way decision 3 asks for: CLAMP, not refuse-and-reprompt."""
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
    """md Lane (i): the current session mode as a NUMBER on the required-N ladder (1/2/3), not the three-value string `_SESSION_MODE` already carries - the one thing `_row_enabled()` below compares a primitive's own required-N against."""
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
    """md Lane (iii): the header's own "session mode: N (word)" text - pairs _session_mode_n()'s number (what every row's own "Brobots mode N" label is compared against) with the plain three-value word (_SESSION_MODE itself) so the header stays legible without the operator having to remember which number means what."""
    word = _SESSION_MODE if _SESSION_MODE is not None else "multi"
    return f"{_session_mode_n()} ({word})"


def _row_enabled(primitive):
    """md Lane (i): THE one comparison - a row/primitive is enabled when the current session mode (as a number) is at least its own required-N."""
    return _session_mode_n() >= _PRIMITIVE_REQUIRED_N[primitive]


def _row_disabled_fire_null(primitive):
    """md Lane 2: replaces the old _row_disabled_skip() contract (renamed, not just re-worded - the old name described a SKIP, first line of the branch, before any prompt ran; this one is reached only once a disabled leaf's own full walk - robot pick, values, choices, exactly as an enabled leaf shows them - is already done, immediately before the point that would otherwise fire a real connection/SDK call)."""
    print(f"Not available in this mode — nothing fired. (needs mode {_PRIMITIVE_REQUIRED_N[primitive]}+, current mode {_session_mode_n()})")
    return 1


def _no_confirmed_robot_this_session():
    """."""
    if _SESSION_MODE == "none":
        return True
    if _SESSION_MODE == "single" and _PRESENT_ROBOTS and not _PRESENT_ROBOTS[0].get("present", True):
        return True
    return False


def _is_none_mode_disabled(primitive):
    """md: true when a control can't do anything with no robot detected - lets a caller decide, before picking, which options to flag."""
    return not _row_enabled(primitive)


def _submenu_control_note(primitive):
    """Per-control note printed next to a control inside one of the 4 new multi-member group sub-menus (moves/audio/say/init) - carries forward the exact brobots_session_responsiveness chain- eligibility fact the old flat main-menu's items 2/3 used to spell out as two long sentences, now attached to the control's own line instead so the main menu can stay short."""
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
    """shared sort for both the root group menu and each multi-member submenu - disabled rows first, then enabled, alphabetical (case-insensitive) within each block."""
    disabled_rows = sorted((r for r in rows if r[2]), key=lambda r: r[3].lower())
    enabled_rows = sorted((r for r in rows if not r[2]), key=lambda r: r[3].lower())
    return disabled_rows + enabled_rows


def _build_primitive_group_tree():
    """md: builds the navigate()-shaped tree from _PRIMITIVE_GROUPS - the ONE real navigate() call site this pass creates (_run_guided_flow_once()'s group/sub menu; every other prompt in this file is unchanged)."""
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
    """md Lane (iii): the main menu's own required-N ladder, made visible per row - replaces the old group-level "[disabled] " prefix (see _build_primitive_group_tree()'s own docstring above for why)."""
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
    """The chain toggle's own call: wake WITHOUT releasing, so the primitive that runs next picks up the SAME held-control window - reuses the exact same core functions cmd_brobots_wake composes (conn_test/assume/golden-flag release-pulse), just skips that command's own tail release."""
    wake_check = run_wheel_nudge_wake_check(mod, clock, serial, live)
    assume = run_assume_control(mod, clock, serial, live)
    wake_confirm = run_wheel_nudge_wake_release_pulse(mod, clock, serial, live)
    ok = wake_check["ok"] and assume["ok"] and wake_confirm["ok"]
    return {"ok": ok, "wake_check": wake_check, "assume": assume, "wake_confirm": wake_confirm}


def _resolve_chain_wake(primitive, mod, clock, serial, live):
    """md (PHCAL_INPUT_TREE_SURVEY_ 001.md §3 DUPLICATES / §5 step 5): the chain-wake prompt-then-fire sequence, pulled out of 4 near-identical inline copies inside _run_guided_flow_once() (animation, brobots_stay_in_place, move_reverse, the shared arm/nod tail) into one place - same prompt (_prompt_brobots_wake_chain), same no-release wake call (_run_brobots_wake_no_release), same failure message shape, just primitive-parameterized instead of copy-pasted per call site."""
    chain_wake = _prompt_brobots_wake_chain(primitive)
    if not chain_wake:
        return False, False
    wake = _run_brobots_wake_no_release(mod, clock, serial, live)
    if not wake["ok"]:
        print(f"{clock.prefix()}PHCAL_BROBOTS_WAKE_CHAIN_FAILED - stopping, not firing {primitive} without a confirmed wake")
        return False, True
    return True, False


def _run_guided_flow_once():
    """md: renamed from run_guided_flow() (now a thin stay-in-loop wrapper below, the real public entry point) - internally unchanged except the group/sub-menu preamble is now its own small while/break/continue loop (see its own comment above) and the two "operator explicitly chose to leave" return sites return None instead of 0, so the wrapper can tell that apart from "a primitive dispatch returned rc=0 (success)." Every other one of this function's ~15 dispatch-branch returns is untouched - same value, same meaning as before this pass."""
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
    """."""
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
    """md, operator decision 5: fires after EVERY _run_guided_flow_once() pass that wasn't an explicit exit - success or failure/blocked alike (e.g."""
    choice = _prompt_pick(
        {"y", "n"}, default="y", labels={"y": "yes", "n": "no"}, question="continue in phcal?",
    )
    return choice == "y"


def _repick_session_mode():
    """md Lane (v)/(vi): the session-mode-changed recompute hook's own "recompute" half - run when run_guided_flow() catches _PhcalSessionModeRepick (Lane vi's own arrow- up trigger)."""
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
    """md: the real public entry point now - a thin stay-in-loop wrapper around _run_guided_flow_once() (the actual guided flow, renamed but otherwise unchanged)."""
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
    """Generic field-chain engine - ONE reusable walker for "ask N fields in order, Left steps back one field, clean redraw" instead of a hand-rolled state machine per chain (this file used to carry two, identical in shape, different fields - Segment 1's reporter-gap pair and Segment 2's phcal-apply/robot-filter/rich-display trio)."""
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
    """navigate()'s own `dispatch(leaf)` for the Point A/B tree above - called once, with the (point_a, point_b) tuple the operator's tree pick resolved to."""
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
    """Public entry point, dispatched from `python3 phcal_isolate_001.py --pha0b-bait-nav` (see this file's own `if __name__ == "__main__":` block below) - song 1's (00a_brobots_awaken / "bait") entire pre-run nav, natively on phcal's own golden nav primitives (arrow_column_pick / _prompt_pick / _prompt_value / _prompt_robot / _resolve_session_mode_once, the same _PhcalBackToMenu/_PhcalEscExit two-tier back-vs-exit every other phcal screen already uses)."""
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

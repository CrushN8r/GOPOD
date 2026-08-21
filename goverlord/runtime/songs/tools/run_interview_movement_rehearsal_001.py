#!/usr/bin/env python3
"""Interview movement rehearsal - fires every scored movement in the
interview's real score, in real playback order, on real hardware, with
placeholder speech instead of generated lines. No LLM, no generation wait.

Reads the interview song's own knobs.json through load_section_pair() -
the exact function generate_phase()/playback_phase() build the real
interview from. Never hardcodes the line list or which lines move; if the
score changes, this tool changes with it. Movement dispatch reuses
fire_scored_interview_movement() unchanged in shape - the same function
playback_phase() itself calls - so ARM_GESTURE_SEQUENCE and
HEAD_NOD_GESTURE_SEQUENCE (both living in run_robot_control_song_001.py,
split 2026-07-14 from the control song's own test sequences - see
MOVEMENT_TEST_GESTURE_SPLIT_001.md) are read once, from their one home,
never duplicated or retuned here.

Control holding, fixed (see INTERVIEW_MOVEMENT_CONTROL_HOLD_FIX_001.md):
fire_scored_interview_movement() gained an additive manage_control
parameter (default True, unchanged for playback_phase()'s own call sites,
which never pass it) - this file now passes manage_control=False on every
movement, since it already assumes control once per robot before the loop
and releases once per robot at the end, the control song's own exact
shape. No re-assume between movements is needed anymore - nothing
releases control out from under the run, so there is nothing to
re-claim.

Pacing: three audible beats, all reusing the scaffold's own
between_exchange_pause_seconds (never an invented number) - a settle beat
before each movement fires, a beat after a movement completes before its
line speaks, and a pause between exchanges, matching how the real
interview paces between exchanges today.
"""
import importlib.util
import json
import os
import time
from pathlib import Path

RUNNER_PATH = Path(__file__).resolve().parent / "run_section1_full_live_001.py"
INTERVIEW_SONG_DIR_ENV = "GOPOD_INTERVIEW_SONG_DIR"

INTERVIEWER_PLACEHOLDER = "Interviewer Line."
INTERVIEWEE_PLACEHOLDER = "Interviewee Line."

CONTROL_ENDPOINTS = ("/api-sdk/assume_behavior_control", "/api-sdk/release_behavior_control")


def spoken_movement_name(movement):
    # Whatever the score names ("head_nod_gesture", "arm_gesture", or any
    # future movement) becomes plain spoken words - never a hardcoded
    # two-name lookup table, so a third scored movement announces itself
    # without this file needing an edit.
    return movement.replace("_", " ").capitalize()


def _load_runner_module():
    spec = importlib.util.spec_from_file_location("run_section1_full_live_001", str(RUNNER_PATH))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def _install_control_call_counter(mod):
    # Wraps the loaded module's own wirepod_web_send_form so every
    # assume/release HTTP call is counted - including the ones
    # fire_scored_interview_movement's reused gentle_arm_gesture_cue()/
    # head_nod_gesture_cue() fire internally, which this file cannot see
    # directly.
    # A local instrumentation wrapper on the module object this file itself
    # loaded - not an edit to run_section1_full_live_001.py on disk.
    calls = []
    original = mod.wirepod_web_send_form

    def wrapped(endpoint, params, timeout=30):
        if endpoint in CONTROL_ENDPOINTS:
            calls.append({"endpoint": endpoint, "serial": params.get("serial")})
        return original(endpoint, params, timeout=timeout)

    mod.wirepod_web_send_form = wrapped
    return calls


def run_rehearsal(mod, song_dir, live, voice_destination):
    scaffold = mod.load_interview_scaffold(mod.INTERVIEW_SCAFFOLD_PATH)
    section = mod.load_section_pair(song_dir)
    lines = section["lines"]
    robots = mod.Robots(live, scaffold, voice_destination)

    # Read, never invent - the same field playback_phase() itself reads for
    # its own between-exchange pacing.
    pause_seconds = float(scaffold.get("between_exchange_pause_seconds", 0) or 0)

    print(f"LIVE: Wire-Pod ESN routing enabled: {mod.WIREPOD_BASE_URL}" if live else "DRY: live speech gate is off.")
    if live:
        print(f"LIVE: Brobot 1 ESN: {mod.BROBOT_1_SERIAL}")
        print(f"LIVE: Brobot 2 ESN: {mod.BROBOT_2_SERIAL}")
    print(f"REHEARSAL_PAUSE_SECONDS {pause_seconds}", flush=True)

    def assume(serial):
        if live:
            mod.wirepod_web_send_form("/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10)

    def release(serial):
        if live:
            mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)

    def beat(label):
        if pause_seconds > 0:
            print(f"REHEARSAL_BEAT {label} seconds={pause_seconds}", flush=True)
            time.sleep(pause_seconds)

    def say_placeholder(name, text):
        robot_safe_line, _repair_used = mod.normalize_robot_safe(text, robots.scaffold)
        return robots.say(name, robot_safe_line, display_text=text, display_role=name, suppress_assume_release=True)

    def fire_role(name, serial, movement, placeholder):
        print(f"REHEARSAL_LINE speaker={name} movement={movement or '(none)'}", flush=True)
        movement_result = None
        if movement:
            beat("settle_before_movement")
            # manage_control=False - control is already held continuously
            # from the single assume before the loop; this call neither
            # assumes nor releases around itself, so nothing needs
            # re-claiming afterward.
            movement_result = mod.fire_scored_interview_movement(movement, serial, live, manage_control=False)
            beat("settle_after_movement")
        # Announce the scored movement out loud so the operator can hear
        # what's supposed to be happening while watching the robot - no
        # hardcoded movement names, whatever the score says gets spoken.
        spoken_line = f"{placeholder} {spoken_movement_name(movement)}." if movement else placeholder
        say_result = say_placeholder(name, spoken_line)
        return movement_result, say_result

    events = []
    control_calls = _install_control_call_counter(mod)
    assume(mod.BROBOT_1_SERIAL)
    assume(mod.BROBOT_2_SERIAL)
    try:
        for index, line in enumerate(lines):
            line_number = line.get("line_number")
            print(f"REHEARSAL_EXCHANGE_START line={line_number}", flush=True)

            brobot_2_movement = line.get("brobot_2_movement")
            brobot_2_movement_result, brobot_2_say_result = fire_role(
                "Brobot 2", mod.BROBOT_2_SERIAL, brobot_2_movement, INTERVIEWER_PLACEHOLDER
            )

            brobot_1_movement = line.get("brobot_1_movement")
            brobot_1_movement_result, brobot_1_say_result = fire_role(
                "Brobot 1", mod.BROBOT_1_SERIAL, brobot_1_movement, INTERVIEWEE_PLACEHOLDER
            )

            events.append(
                {
                    "line_number": line_number,
                    "brobot_2_speaker": "Brobot 2",
                    "brobot_2_movement": brobot_2_movement,
                    "brobot_2_movement_fired": bool(brobot_2_movement),
                    "brobot_2_movement_result": brobot_2_movement_result,
                    "brobot_2_say_result": brobot_2_say_result,
                    "brobot_1_speaker": "Brobot 1",
                    "brobot_1_movement": brobot_1_movement,
                    "brobot_1_movement_fired": bool(brobot_1_movement),
                    "brobot_1_movement_result": brobot_1_movement_result,
                    "brobot_1_say_result": brobot_1_say_result,
                }
            )
            print(f"REHEARSAL_EXCHANGE_DONE line={line_number}", flush=True)
            if index < len(lines) - 1:
                beat("between_exchange")
    finally:
        release(mod.BROBOT_1_SERIAL)
        release(mod.BROBOT_2_SERIAL)
        robots.close()

    assume_count = sum(1 for c in control_calls if c["endpoint"] == "/api-sdk/assume_behavior_control")
    release_count = sum(1 for c in control_calls if c["endpoint"] == "/api-sdk/release_behavior_control")
    print(f"REHEARSAL_CONTROL_CALLS assume={assume_count} release={release_count}", flush=True)

    return {
        "tool": "interview_movement_rehearsal",
        "song_id": section.get("section_id", ""),
        "song_dir": str(song_dir),
        "live": live,
        "voice_destination": voice_destination,
        "brobot_1_serial": mod.BROBOT_1_SERIAL,
        "brobot_2_serial": mod.BROBOT_2_SERIAL,
        "pause_seconds": pause_seconds,
        "control_calls": control_calls,
        "control_call_counts": {"assume": assume_count, "release": release_count},
        "events": events,
    }


def write_rehearsal_log(mod, song_dir, log):
    run_id = time.strftime("%Y%m%dT%H%M%SZ", time.gmtime())
    log["run_id"] = run_id
    runs_dir = Path(song_dir) / "runs"
    runs_dir.mkdir(parents=True, exist_ok=True)
    path = runs_dir / f"interview_movement_rehearsal_run_{run_id}.json"
    path.write_text(json.dumps(log, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    print(f"REHEARSAL_JSON_WRITTEN {path}", flush=True)
    log["log_path"] = str(path)
    return path


def main():
    mod = _load_runner_module()
    voice_destination = os.getenv("GOPOD_VOICE_DESTINATION", "robot")
    live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
    song_dir = os.getenv(INTERVIEW_SONG_DIR_ENV, str(mod.DEFAULT_SECTION_SONG_DIR))
    log = run_rehearsal(mod, song_dir, live, voice_destination)
    write_rehearsal_log(mod, song_dir, log)
    print(json.dumps(log, indent=2))
    return log


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""Robot control song — a talking self-check on one robot (Brobot 1)."""
import importlib.util
import json
import os
import re
import time
from pathlib import Path

RUNNER_PATH = Path(__file__).resolve().parent / "run_section1_full_live_001.py"
WEATHER_MODULE_PATH = Path(os.getenv("GOPOD_REPO_ROOT", "/home/goverlord/crushn8r_git/GOPOD")).expanduser() / "goverlord" / "runtime" / "songs" / "tools" / "gopod_weather_fetch_001.py"


def _load_runner_module():
    spec = importlib.util.spec_from_file_location("run_section1_full_live_001", str(RUNNER_PATH))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def _load_weather_module():
    spec = importlib.util.spec_from_file_location("gopod_weather_fetch_001", str(WEATHER_MODULE_PATH))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def fetch_and_format_weather(serial, include_location_date=False, location=None):
    weather_mod = _load_weather_module()
    facts = weather_mod.fetch_windsor_weather(location=location)
    fmt = weather_mod.load_robot_format_from_jdocs(serial)
    return weather_mod.format_for_robot(facts, fmt["unit"], fmt["clock"], include_location_date=include_location_date)


_STEP_HEADING = re.compile(r"^##\s+STEP\s+(\S+)\s*$")


def parse_control_story_md(text):
    """Each step's text/fail by its name (the last one wins when a name repeats)."""
    return {step_id: content for step_id, content in parse_control_story_list(text)}


def parse_control_story_list(text):
    """every step in story order, (step_id, {"text", "fail"})."""
    lines = text.splitlines()
    starts = [i for i, line in enumerate(lines) if _STEP_HEADING.match(line.strip())]
    steps = []
    for position, start in enumerate(starts):
        step_id = _STEP_HEADING.match(lines[start].strip()).group(1)
        end = starts[position + 1] if position + 1 < len(starts) else len(lines)
        block = lines[start + 1 : end]
        text_val = ""
        fail_val = ""
        in_text = False
        for raw in block:
            stripped = raw.strip()
            upper = stripped.upper()
            if upper.startswith("> TEXT:"):
                text_val = stripped[len("> TEXT:") :].strip()
                in_text = True
            elif upper.startswith("> FAIL:"):
                fail_val = stripped[len("> FAIL:") :].strip()
                in_text = False
            elif in_text and stripped and not stripped.startswith("#") and not stripped.startswith(">"):
                text_val = (text_val + "\n" + stripped) if text_val else stripped
            else:
                in_text = False
        steps.append((step_id, {"text": text_val, "fail": fail_val}))
    return steps


def run_named_movement_sequence(mod, name, endpoint, sequence, serial, live, manage_control=True, timing_steps=None):
    mod.validate_movement_sequence_ends_at_zero(sequence)
    if not live:
        return {"ok": True, "mode": "dry", "name": name, "results": []}
    if manage_control:
        mod.wirepod_web_send_form("/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10)
    results = []
    ok = True
    try:
        for step_index, step in enumerate(sequence):
            call_start = time.monotonic()
            result = mod.wirepod_web_send_form(
                endpoint, {"speed": str(step["speed"]), "serial": serial}, timeout=10
            )
            call_elapsed = time.monotonic() - call_start
            results.append(result)
            if result.get("status") != 200:
                ok = False
            hold_seconds = step["hold_seconds"]
            hold_start = time.monotonic()
            time.sleep(hold_seconds)
            hold_elapsed = time.monotonic() - hold_start
            if timing_steps is not None:
                timing_steps.append(
                    {
                        "step_index": step_index,
                        "speed": step["speed"],
                        "http_call_seconds": round(call_elapsed, 6),
                        "http_status": result.get("status"),
                        "capture_quality": "accept_only",
                        "configured_hold_seconds": hold_seconds,
                        "actual_hold_seconds": round(hold_elapsed, 6),
                        "note": (
                            "http_call_seconds is real accept-latency. configured/actual_hold_seconds "
                            "is an asserted sleep, not a measured motion-completion signal."
                        ),
                    }
                )
    finally:
        if manage_control:
            mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)
    return {"ok": ok, "name": name, "results": results}


ARM_TEST_LEG_HOLD_SECONDS = 0.36  # raise/lower leg duration - live-confirmed good

ARM_TEST_SEQUENCE = [
    {"speed": 0, "hold_seconds": 0.3},   # settle to rest (explicit start anchor)
    {"speed": 1, "hold_seconds": ARM_TEST_LEG_HOLD_SECONDS},   # raise, slow
    {"speed": 0, "hold_seconds": 1.2},   # pause held at top
    {"speed": -1, "hold_seconds": ARM_TEST_LEG_HOLD_SECONDS},  # lower, slow
    {"speed": 0, "hold_seconds": 0.8},   # settle to rest (explicit end anchor)
]

NOD_TEST_SPEED = 3             # nod leg speed magnitude - DO NOT LOWER below 3 without live re-confirmation, see warning above
NOD_TEST_LEAD_IN_SPEED = 2     # lead-in speed magnitude - scaled with NOD_TEST_SPEED (ratio ~0.667)
NOD_TEST_LEG_HOLD_SECONDS = 0.36  # nod-down/nod-up leg duration - live-confirmed good

HEAD_NOD_TEST_SEQUENCE = [
    {"speed": 0, "hold_seconds": 0.3},              # settle to rest (explicit start anchor)
    {"speed": NOD_TEST_LEAD_IN_SPEED, "hold_seconds": 0.3},  # head up slightly (lead-in)
    {"speed": 0, "hold_seconds": 0.4},              # pause
    {"speed": -NOD_TEST_SPEED, "hold_seconds": NOD_TEST_LEG_HOLD_SECONDS},     # nod 1 down
    {"speed": NOD_TEST_SPEED, "hold_seconds": NOD_TEST_LEG_HOLD_SECONDS},      # nod 1 up
    {"speed": 0, "hold_seconds": 0.5},              # pause
    {"speed": -NOD_TEST_SPEED, "hold_seconds": NOD_TEST_LEG_HOLD_SECONDS},     # nod 2 down
    {"speed": NOD_TEST_SPEED, "hold_seconds": NOD_TEST_LEG_HOLD_SECONDS},      # nod 2 up
    {"speed": 0, "hold_seconds": 0.6},              # settle to rest (explicit end anchor)
]


ARM_GESTURE_LEG_HOLD_SECONDS = 0.81  # seeded from ARM_TEST_LEG_HOLD_SECONDS - free to diverge, tune by ear

ARM_GESTURE_SEQUENCE = [
    {"speed": 0, "hold_seconds": 0.3},   # settle to rest (explicit start anchor)
    {"speed": 1, "hold_seconds": ARM_GESTURE_LEG_HOLD_SECONDS},   # raise, slow
    {"speed": 0, "hold_seconds": 1.2},   # pause held at top
    {"speed": -1, "hold_seconds": ARM_GESTURE_LEG_HOLD_SECONDS},  # lower, slow
    {"speed": 0, "hold_seconds": 0.8},   # settle to rest (explicit end anchor)
]

NOD_GESTURE_SPEED = 3             # nod leg speed magnitude - DO NOT LOWER below 3 without live re-confirmation, see warning above
NOD_GESTURE_LEAD_IN_SPEED = 2     # lead-in speed magnitude - scaled with NOD_GESTURE_SPEED (ratio ~0.667)
NOD_GESTURE_LEG_HOLD_SECONDS = 0.81  # seeded from NOD_TEST_LEG_HOLD_SECONDS - free to diverge, tune by ear (TIMING only, never speed)

HEAD_NOD_GESTURE_SEQUENCE = [
    {"speed": 0, "hold_seconds": 0.3},              # settle to rest (explicit start anchor)
    {"speed": NOD_GESTURE_LEAD_IN_SPEED, "hold_seconds": 0.3},  # head up slightly (lead-in)
    {"speed": 0, "hold_seconds": 0.4},              # pause
    {"speed": -NOD_GESTURE_SPEED, "hold_seconds": NOD_GESTURE_LEG_HOLD_SECONDS},     # nod 1 down
    {"speed": NOD_GESTURE_SPEED, "hold_seconds": NOD_GESTURE_LEG_HOLD_SECONDS},      # nod 1 up
    {"speed": 0, "hold_seconds": 0.5},              # pause
    {"speed": -NOD_GESTURE_SPEED, "hold_seconds": NOD_GESTURE_LEG_HOLD_SECONDS},     # nod 2 down
    {"speed": NOD_GESTURE_SPEED, "hold_seconds": NOD_GESTURE_LEG_HOLD_SECONDS},      # nod 2 up
    {"speed": 0, "hold_seconds": 0.6},              # settle to rest (explicit end anchor)
]


def head_nod_test_cue(mod, serial, live, manage_control=True):
    return run_named_movement_sequence(
        mod, "head_nod_test", "/api-sdk/move_head",
        HEAD_NOD_TEST_SEQUENCE, serial, live, manage_control,
    )


def gentle_arm_test_cue(mod, serial, live, manage_control=True):
    return run_named_movement_sequence(
        mod, "arm_test", "/api-sdk/move_lift",
        ARM_TEST_SEQUENCE, serial, live, manage_control,
    )


def head_nod_gesture_cue(mod, serial, live, manage_control=True, timing_steps=None):
    return run_named_movement_sequence(
        mod, "head_nod_gesture", "/api-sdk/move_head",
        HEAD_NOD_GESTURE_SEQUENCE, serial, live, manage_control, timing_steps=timing_steps,
    )


def gentle_arm_gesture_cue(mod, serial, live, manage_control=True, timing_steps=None):
    return run_named_movement_sequence(
        mod, "arm_gesture", "/api-sdk/move_lift",
        ARM_GESTURE_SEQUENCE, serial, live, manage_control, timing_steps=timing_steps,
    )


def run_move_axis(mod, serial, endpoint, speed, hold_seconds, live, manage_control=True):
    """Ports _brobots_move_axis's exact HTTP shape (brobots.sh) for a single robot: assume -> set speed -> hold -> set speed=0 -> release."""
    if not live:
        return {"ok": True, "mode": "dry"}
    assume = None
    if manage_control:
        assume = mod.wirepod_web_send_form("/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10)
    move = mod.wirepod_web_send_form(f"/api-sdk/{endpoint}", {"serial": serial, "speed": speed}, timeout=10)
    time.sleep(hold_seconds)
    stop = mod.wirepod_web_send_form(f"/api-sdk/{endpoint}", {"serial": serial, "speed": 0}, timeout=10)
    release = None
    if manage_control:
        release = mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)
    ok = (assume is None or assume.get("status") == 200) and move.get("status") == 200 and stop.get("status") == 200
    return {"ok": ok, "assume": assume, "move": move, "stop": stop, "release": release}


def run_arm_cue(mod, serial, cycles, hold_seconds, live, speed=2, manage_control=True):
    """brobots-lift-up/-down's own shape (brobots.sh), one robot, N up/down
    cycles - each direction its own move via run_move_axis. Shared by
    brobots_bingo and this song's own "arm_cue" note (see manage_control's
    own docstring on run_move_axis above)."""
    if not live:
        return {"ok": True, "mode": "dry", "cycles": cycles}
    moves = []
    ok = True
    for _ in range(cycles):
        up = run_move_axis(mod, serial, "move_lift", speed, hold_seconds, live, manage_control=manage_control)
        down = run_move_axis(mod, serial, "move_lift", -speed, hold_seconds, live, manage_control=manage_control)
        moves.append({"up": up, "down": down})
        ok = ok and up["ok"] and down["ok"]
    return {"ok": ok, "cycles": cycles, "moves": moves}


def run_nod(mod, serial, hold_seconds, live, speed=2, manage_control=True):
    """brobots-head-nod's own shape (brobots.sh): one down movement, then one
    up movement, each via run_move_axis. Shared by brobots_bingo and this
    song's own "head_nod" note."""
    if not live:
        return {"ok": True, "mode": "dry"}
    down = run_move_axis(mod, serial, "move_head", -speed, hold_seconds, live, manage_control=manage_control)
    up = run_move_axis(mod, serial, "move_head", speed, hold_seconds, live, manage_control=manage_control)
    return {"ok": down["ok"] and up["ok"], "down": down, "up": up}


WHEEL_NUDGE_REVERSE_LW = -150
WHEEL_NUDGE_REVERSE_RW = -150
WHEEL_NUDGE_DEFAULT_HOLD_SECONDS = 0.5
BROBOTS_WAKE_CHECK_TIMEOUT_SECONDS = 15  # CONNECTION_PASS_TIMEOUT_SECONDS's own value, matched - conn_test's own timeout, unchanged
BROBOTS_WAKE_RELEASE_SETTLE_SECONDS = 2.5  # measured max observed 1.817s across 8 live runs, both robots - this adds real margin above that ceiling for an unattended/live-audience run rather than shipping the bare minimum; see WHEEL_NUDGE_COLD_FIRE_ROOT_CAUSE_SURVEY_001.md for the full run-by-run numbers
BROBOTS_WAKE_POST_REASSUME_SETTLE_SECONDS = 0.5

POST_ASSUME_SAY_SETTLE_SECONDS = 1.5


def run_brobots_wake(mod, serial, live, manage_control=True):
    """."""
    if not live:
        return {"ok": True, "mode": "dry"}
    try:
        conn_test = mod.wirepod_web_send_form(
            "/api-sdk/conn_test", {"serial": serial}, timeout=BROBOTS_WAKE_CHECK_TIMEOUT_SECONDS
        )
        conn_ok = conn_test.get("status") == 200
    except Exception as exc:  # noqa: BLE001 - a hang/timeout here is exactly what this check exists to catch fast
        conn_test, conn_ok = {"error": repr(exc)}, False
    assume = None
    if manage_control:
        assume = mod.wirepod_web_send_form("/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10)
    release_pulse = mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)
    time.sleep(BROBOTS_WAKE_RELEASE_SETTLE_SECONDS)
    reassume = mod.wirepod_web_send_form("/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10)
    time.sleep(BROBOTS_WAKE_POST_REASSUME_SETTLE_SECONDS)
    ok = (
        conn_ok
        and (assume is None or assume.get("status") == 200)
        and release_pulse.get("status") == 200
        and reassume.get("status") == 200
    )
    return {"ok": ok, "conn_test": conn_test, "assume": assume, "release_pulse": release_pulse, "reassume": reassume}


def run_move_reverse(mod, serial, live, hold_seconds, lw=WHEEL_NUDGE_REVERSE_LW, rw=WHEEL_NUDGE_REVERSE_RW, manage_control=True):
    """The move half."""
    if not live:
        return {"ok": True, "mode": "dry"}
    assume = None
    if manage_control:
        assume = mod.wirepod_web_send_form("/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10)
    move = mod.wirepod_web_send_form("/api-sdk/move_wheels", {"serial": serial, "lw": lw, "rw": rw}, timeout=10)
    time.sleep(hold_seconds)
    stop = mod.wirepod_web_send_form("/api-sdk/move_wheels", {"serial": serial, "lw": 0, "rw": 0}, timeout=10)
    release = None
    if manage_control:
        release = mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)
    ok = (
        (assume is None or assume.get("status") == 200)
        and move.get("status") == 200
        and stop.get("status") == 200
        and (release is None or release.get("status") == 200)
    )
    return {"ok": ok, "assume": assume, "move": move, "stop": stop, "release": release}


def run_wheel_nudge(mod, serial, lw, rw, hold_seconds, live, manage_control=True):
    """Composed: run_brobots_wake's own core (its own release/settle/ re-assume golden-flag pulse, no outer release of its own) then run_move_reverse's own core (no assume - relies on wake's own re-assume having just landed), sharing ONE continuous held-control window opened/closed ONCE here."""
    if not live:
        return {"ok": True, "mode": "dry"}
    assume = None
    if manage_control:
        assume = mod.wirepod_web_send_form("/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10)
    wake = run_brobots_wake(mod, serial, live, manage_control=False)
    move = run_move_reverse(mod, serial, live, hold_seconds, lw=lw, rw=rw, manage_control=False)
    release = None
    if manage_control:
        release = mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)
    ok = (
        (assume is None or assume.get("status") == 200)
        and wake["ok"]
        and move["ok"]
        and (release is None or release.get("status") == 200)
    )
    return {"ok": ok, "assume": assume, "wake": wake, "move_reverse": move, "release": release}


def fire_fireworks(mod, serial, live):
    if not live:
        return {"ok": True, "mode": "dry", "results": []}
    result = mod.wirepod_web_send_form(
        "/api-sdk/cloud_intent", {"intent": "intent_seasonal_happynewyear", "serial": serial}, timeout=10
    )
    return {"ok": result.get("status") == 200, "results": [result]}


def say_line(mod, robots, name, text, display_text=None):
    if not text:
        return {"ok": True, "skipped": True}
    if display_text is None:
        display_text = text
    robot_safe_line, _repair_used = mod.normalize_robot_safe(text, robots.scaffold)
    result = robots.say(
        name, robot_safe_line, display_text=display_text, display_role=name,
        suppress_assume_release=True,
    )
    speech_text = result.get("speech_text")
    raw_speech_text = result.get("raw_speech_text")
    if speech_text and raw_speech_text and speech_text != raw_speech_text:
        print(f"{name} PRONUNCIATION_SWAP: raw={raw_speech_text!r} spoken={speech_text!r}", flush=True)
    ok = result.get("spoken") is True or result.get("mode") in ("dry", "off")
    return {"ok": ok, "result": result}


def run_single_note(note, live, robot="1", voice_destination="robot", location=None, phrase=None):
    """Fire exactly one motion/note in isolation - no song loop, no narration, no other notes."""
    mod = _load_runner_module()
    if robot == "2":
        serial = mod.BROBOT_2_SERIAL
        name = "Brobot 2"
    elif robot == "1":
        serial = mod.BROBOT_1_SERIAL
        name = "Brobot 1"
    else:
        raise RuntimeError(f"BLOCKED: unknown robot {robot!r}, must be '1' or '2'")

    if note == "arm_test":
        result = gentle_arm_test_cue(mod, serial, live)
    elif note == "head_nod":
        result = head_nod_test_cue(mod, serial, live)
    elif note == "fireworks":
        result = fire_fireworks(mod, serial, live)
    elif note == "weather":
        scaffold = mod.load_interview_scaffold(mod.INTERVIEW_SCAFFOLD_PATH)
        robots = mod.Robots(live, scaffold, voice_destination)
        try:
            if live:
                mod.wirepod_web_send_form(
                    "/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10
                )
                time.sleep(POST_ASSUME_SAY_SETTLE_SECONDS)
            weather = fetch_and_format_weather(serial, location=location)
            result = say_line(mod, robots, name, weather["spoken"], display_text=weather["display_text"])
            result["weather"] = weather
        finally:
            if live:
                mod.wirepod_web_send_form(
                    "/api-sdk/release_behavior_control", {"serial": serial}, timeout=10
                )
            robots.close()
    elif note == "say_phrase":
        if not phrase:
            raise RuntimeError("BLOCKED: say_phrase requires a non-empty phrase")
        scaffold = mod.load_interview_scaffold(mod.INTERVIEW_SCAFFOLD_PATH)
        robots = mod.Robots(live, scaffold, voice_destination)
        try:
            if live:
                mod.wirepod_web_send_form(
                    "/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10
                )
                time.sleep(POST_ASSUME_SAY_SETTLE_SECONDS)
            result = say_line(mod, robots, name, phrase)
        finally:
            if live:
                mod.wirepod_web_send_form(
                    "/api-sdk/release_behavior_control", {"serial": serial}, timeout=10
                )
            robots.close()
    else:
        raise RuntimeError(
            f"BLOCKED: unknown note {note!r}, must be 'arm_test', 'head_nod', 'fireworks', 'weather', or 'say_phrase'"
        )

    print(json.dumps({"note": note, "robot": robot, "live": live, "result": result}, indent=2))
    return result


def main():
    voice_destination = os.getenv("GOPOD_VOICE_DESTINATION", "robot")
    live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
    robot = os.getenv("GOPOD_CONTROL_SONG_ROBOT", "1")
    isolated_note = os.getenv("GOPOD_CONTROL_SONG_NOTE", "").strip()
    if not isolated_note:
        raise RuntimeError(
            "BLOCKED: the full-song loop was retired - this script is a "
            "tester-only entry point now. Set GOPOD_CONTROL_SONG_NOTE to "
            "arm_test/head_nod/fireworks/weather, or use "
            "start-the-control-song (runs through run_golden_song_001.py)."
        )
    return run_single_note(isolated_note, live, robot, voice_destination)


if __name__ == "__main__":
    main()

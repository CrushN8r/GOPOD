#!/usr/bin/env python3
"""Robot control song — a talking self-check on one robot (Brobot 1).

Standalone runner, deliberately never touching run_section1_full_live_001.py
beyond one additive opt-out parameter on Robots.say() (suppress_assume_release,
defaults to today's behavior - see that file). Same importlib-reuse pattern
run_section1_preshow_generate_001.py already uses.

The shared Robots.say() no longer fires speaker_visual_cue() automatically
on any call, for anyone - the debug-era welded arm cue was cut from say()
itself, not just suppressed here.

ARM_TEST_SEQUENCE/HEAD_NOD_TEST_SEQUENCE + gentle_arm_test_cue()/
head_nod_test_cue() are this file's own fixed stress-test choreography -
still live, but as of 2026-07-25 only fired by the standalone
test-arm-cue/test-head-nod aliases (via run_single_note()), not by any
song's own score. This song's own arm_test/head_nod notes (brobots_awaken)
now call run_arm_cue()/run_nod() instead (below) - the same tunable,
knobs.json-driven movement brobots_bingo already used, moved here so both
songs share one mechanism (REPORTER_GAP_SHARED_SWITCH_SURVEY_001.md,
operator request). Robots.speaker_visual_cue() remains the separate,
still-dormant third option, callable on demand, no score fires it.

TESTS vs. GESTURES (split 2026-07-14, MOVEMENT_TEST_GESTURE_SPLIT_001.md):
this file now owns two independent pairs of movement sequences, not one
shared pair. ARM_TEST_SEQUENCE/HEAD_NOD_TEST_SEQUENCE are this song's own
stress tests - big, deliberate, built to prove the motor works, fired by
the test-arm-cue/test-head-nod aliases (see above - no longer by this
song's own score). ARM_GESTURE_SEQUENCE/HEAD_NOD_GESTURE_SEQUENCE are the
interview's own performance moves, fired only through fire_scored_interview_movement()
in run_section1_full_live_001.py. They were seeded as exact copies of the
test values and are free to diverge and be tuned by ear later - neither
pair reaches into the other's values anymore.

Fireworks (fire_fireworks(), the native intent_seasonal_happynewyear via
/api-sdk/cloud_intent - see NATIVE_CELEBRATION_SURVEY_001.md /
NATIVE_CELEBRATION_LIVE_TEST_001.md / CELEBRATION_COMPLEXITY_PINNED_001.md)
is no longer part of the scored song - it produces no visible animation,
pinned, not re-investigated. It remains reachable only as the standalone
test-fireworks alias, via run_single_note().

The weather note took fireworks' slot in the score: a real fetch through
the existing goverlord/runtime/data_gomad/robot/weather/gopod_weather_fetch_001.py module,
formatted per-robot from that robot's own LIVE Wire-Pod jdocs RobotSettings
(temp_is_fahrenheit/clock_24_hour - jdocs.json, keyed by serial; changed
2026-08-13, WEATHER_LIVE_JDOCS_SOURCED_001.md, replacing a hand-authored
static file that happened to match but could silently drift). Currently
Brobot 1 leads Celsius/24-hour, Brobot 2 leads Fahrenheit/12-hour, per each
robot's own real settings. The spoken line uses words (no degree symbols);
the display_text passed to say() uses the symbol form - both come from that
module's own format_for_robot(), not re-derived here.

Three additive extensions for brobots_bait_002 (2026-07-15/2026-07-16), all
backward-compatible - no other existing song's knobs.json uses any of these keys/tokens,
so their behavior is unchanged byte for byte: (1) a "pause" note - a plain, deterministic
time.sleep() for pause_seconds seconds read from the score, no hardware call, for reporter
gaps between beats; (2) a "{robot_name}" token in step text, substituted with this run's
own already-computed `name` ("Brobot 1"/"Brobot 2") - lets one song folder's story.md
speak a per-robot self-ID line without hardcoding either robot's name in this file; (3) a
"weather_include_location_date" flag read per-weather-step from knobs.json, passed through
to gopod_weather_fetch_001.py's own format_for_robot() - defaults False, so
robot_control_song_001 and brobots_awaken (neither sets this key) get byte-identical
weather output; only a song that explicitly sets it True gets the location+date prefix.

PRONUNCIATION (corrected 2026-07-16 - the 2026-07-15 note above was wrong): this song's
own normalize_robot_safe() has no pronunciation handling, but say_line() doesn't stop
there - it hands the result to robots.say(), and `robots` is mod.Robots(...), the exact
same shared class the interview instantiates. Robots.say() unconditionally runs
flatten_for_robot_speech() -> apply_pronunciation_safety() on every line, from either
lane, before building the wirepod_chunks a live call actually POSTs. So "GOPOD" -> "Gowp-
awd" already applies here automatically, at speech time only - story.md, the terminal
print, and every log field except result["speech_text"] all keep the literal spelling
untouched, same divergence-by-design as the interview lane. No song needs to author a
phonetic respelling in its own text for this reason; brobots_bait_002's story.md was
reverted to literal "GOPOD Yourself" for exactly that reason. say_line()'s own
PRONUNCIATION_SWAP print (silent unless a line's text actually triggers a registry entry)
is the live, per-run proof of this.
"""
# ============================================================================
# GOLDEN TRUTHS ABOUT DRIVING THESE ROBOTS - READ THIS BEFORE TOUCHING
# ANYTHING BELOW. Consolidated here 2026-07-14 as the single source of
# truth for robot control - a future session should read this block before
# re-deriving any of it from scratch or from scattered session notes.
# ============================================================================
#
# 1. HTTP 200 does not mean the motor moved. Wire-Pod returns 200 the
#    instant it accepts a command, not when the robot finishes - or even
#    starts - moving. Below a certain speed, a motor command is accepted,
#    reported as a clean success, and the robot simply does not move.
#    Confirmed live on Brobot 1's head: dead at speeds 1.5, 2.0, 2.5, and
#    2.75; working at 3.0. The arm was confirmed working at speed 1; it
#    has never been tested below 1. Never trust "it returned 200" as proof
#    of physical motion - only a human watching the robot can confirm that.
#
# 2. Assume control once, hold it, release it once. A command that
#    assumes behavior control and then releases it hands control straight
#    back to Vector's own onboard behavior, even for an instant, between
#    commands. Doing this per-command (assume, act, release, assume, act,
#    release...) was the original defect: the robot's screen flashed back
#    to its home/idle icon after every single note. The fix is to assume
#    once at the start of a run, hold it continuously through every
#    command, and release once at the very end - exactly how Wire-Pod's
#    own native control page (webroot/sdkapp/js/control.js) does it.
#
# 3. Every movement sequence must end at an explicit speed of zero. This
#    is enforced by a real gate (validate_movement_sequence_ends_at_zero),
#    not just a convention - it raises if the last step isn't zero. Never
#    disable it, never bypass it.
#
# 4. Every movement sequence is rest-anchored at both ends. It starts by
#    settling to a known rest position (speed 0, held) before doing
#    anything else, and ends the same way - never assuming the robot is
#    already wherever the sequence wants to start from.
#
# 5. hold_seconds is asserted, never measured. It is how long this code
#    waits before sending the next command - not a confirmation that the
#    robot's actual physical motion finished. There is no signal anywhere
#    in this stack that proves a move completed; the HTTP response comes
#    back as soon as the command is issued, not when the motion is done.
#
# 6. The welded arm cue is gone. Nothing fires movement automatically as
#    a side effect of a robot speaking, anywhere, ever again. Movement
#    only fires when it is a deliberately scored note (or, in the
#    interview's own case, a deliberately scored per-line movement) -
#    never welded to speech.
#
# 7. Fireworks is pinned - do not investigate it further. It is a real
#    robot animation (intent_seasonal_happynewyear via /api-sdk/cloud_intent)
#    that returns a clean 200 "done" every time it fires, and produces no
#    visible result on the robot. It lives only as the standalone
#    test-fireworks alias - it is not, and must not become, part of any
#    scored song.
#
# 8. The robots themselves hang under sustained load - this is not a code
#    defect. Confirmed live: Brobot 1 went fully unreachable mid-session
#    (its own gRPC port refusing every connection, a "980" screen error)
#    while Wire-Pod itself stayed up and answered fine the whole time. A
#    power cycle of the robot fixed it immediately; no code change was
#    needed or made. If Wire-Pod is answering but a robot's own commands
#    are timing out or being refused, the robot is the problem, not this
#    code - power-cycle it before assuming anything else is wrong.
# ============================================================================
import importlib.util
import json
import os
import re
import time
from pathlib import Path

RUNNER_PATH = Path(__file__).resolve().parent / "run_section1_full_live_001.py"
# Env-overridable absolute anchor, not location-derived (2026-08-15,
# INTERVIEW_TOOLS_MOVE_EXECUTED_001.md) - same fix as run_golden_song_001.py's
# own SONGS_DIR, same reason: parents[2] silently broke the moment this
# file's folder moved up a level.
WEATHER_MODULE_PATH = Path(os.getenv("GOPOD_REPO_ROOT", "/home/goverlord/crushn8r_git/GOPOD")).expanduser() / "goverlord" / "runtime" / "data_gomad" / "robot" / "weather" / "gopod_weather_fetch_001.py"


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
    # No try/except here by design - a fetch failure must raise and stop the
    # song, not be swallowed into a fallback or canned line.
    # include_location_date defaults to False - robot_control_song_001 and
    # brobots_awaken both call this with no third argument, so their
    # output is unaffected by this parameter's existence.
    # Takes the already-resolved robot serial (both call sites below already
    # compute this from mod.BROBOT_1_SERIAL/mod.BROBOT_2_SERIAL for their own
    # hardware calls), not "1"/"2" - unit/clock ordering is now read LIVE
    # from that robot's own jdocs RobotSettings, not a hand-authored static
    # file (WEATHER_LIVE_JDOCS_SOURCED_001.md - closes a silent-drift seam).
    # location=None (every existing caller here and the song itself)
    # preserves the original Windsor-only fetch unchanged - added
    # 2026-08-15 (PHCAL_NOTES_WEATHER_WAKE_FIXES_001.md) purely for phcal's
    # own bench weather test to target a different place on request.
    weather_mod = _load_weather_module()
    facts = weather_mod.fetch_windsor_weather(location=location)
    fmt = weather_mod.load_robot_format_from_jdocs(serial)
    return weather_mod.format_for_robot(facts, fmt["unit"], fmt["clock"], include_location_date=include_location_date)


_STEP_HEADING = re.compile(r"^##\s+STEP\s+(\S+)\s*$")


def parse_control_story_md(text):
    lines = text.splitlines()
    starts = [i for i, line in enumerate(lines) if _STEP_HEADING.match(line.strip())]
    steps = {}
    for position, start in enumerate(starts):
        step_id = _STEP_HEADING.match(lines[start].strip()).group(1)
        end = starts[position + 1] if position + 1 < len(starts) else len(lines)
        block = lines[start + 1 : end]
        text_val = ""
        fail_val = ""
        for raw in block:
            stripped = raw.strip()
            if stripped.upper().startswith("> TEXT:"):
                text_val = stripped[len("> TEXT:") :].strip()
            elif stripped.upper().startswith("> FAIL:"):
                fail_val = stripped[len("> FAIL:") :].strip()
        steps[step_id] = {"text": text_val, "fail": fail_val}
    return steps




# Shared mechanism-control engine: one named, reusable runner for any
# speed/hold movement sequence, not two one-off scripts. Every sequence
# passed through here must end at an explicit zero -
# validate_movement_sequence_ends_at_zero() (the shared gate, reused as-is,
# never duplicated or disabled) is called unconditionally, dry or live.
def run_named_movement_sequence(mod, name, endpoint, sequence, serial, live, manage_control=True, timing_steps=None):
    mod.validate_movement_sequence_ends_at_zero(sequence)
    if not live:
        return {"ok": True, "mode": "dry", "name": name, "results": []}
    # manage_control=True (the default) assumes/releases around just this
    # one call - correct for the isolated single-note testers
    # (run_single_note()/test-arm-cue/test-head-nod), where nothing else is
    # holding control. manage_control=False is used inside the full song,
    # where the "connect" note already assumed control once for the whole
    # run (matching webroot/sdkapp/js/control.js's own pattern - assume
    # once, fire many commands, release once) - doing it again here would
    # be exactly the repeated assume/release churn that caused the robot's
    # screen to flash to its home icon between every note, confirmed live.
    if manage_control:
        mod.wirepod_web_send_form("/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10)
    results = []
    ok = True
    try:
        # timing_steps (additive, default None - every existing caller that
        # doesn't pass it is byte-for-byte unaffected): per-waypoint timing
        # instrumentation only, never changes speed/hold_seconds/what fires.
        # http_call_seconds is a real, honest accept-latency measurement;
        # configured/actual_hold_seconds is the asserted sleep this code
        # always took anyway (golden truth 5 above) - captured here, not
        # invented, so a timing map never has to guess it from the score.
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


# Every sequence below shares one shape, named plainly so a score can call
# it by name: settle to rest -> move -> hold -> return -> settle to rest.
# Every motion begins AND ends at an explicit, deliberate rest step (speed
# 0, held) - never implicit, never assuming the robot is already wherever
# the sequence wants to start from. Golden truths 1/3/4/8 above apply to
# every sequence in this file without exception.
#
# TESTS vs. GESTURES, split 2026-07-14 (MOVEMENT_TEST_GESTURE_SPLIT_001.md):
# one sequence cannot be both a stress test and a performance gesture -
# the control song's own tests are big and deliberate, built to prove the
# motor works; the interview's gestures are meant to read as natural
# performance moves and will be tuned by ear over time. Splitting them
# means tuning one can never silently change the other. The gesture pairs
# below were seeded as EXACT copies of the test values at the moment of
# the split - today's behaviour is unchanged, only the naming and the
# independence are new.

# --- TESTS: the control song's own stress tests (arm_test/head_nod_test
# notes, test-arm-cue/test-head-nod aliases). Untouched by the split. ---
#
# ARM_TEST_SEQUENCE - live feedback on Brobot 1 described the previous
# gentle attempt as "hard to tell" - erred toward slower and longer than
# feels right on paper, per the operator's own instruction, rather than
# retuning by feel again. Deliberately NOT the dormant
# Robots.speaker_visual_cue()'s scaffold-defined sequence (speed 2/0/-2/0,
# 0.25s/0.5s holds) - that sequence, and the scaffold field it reads from,
# is untouched, still callable on demand, just no longer wired into every
# say() call.
#
# ARM FLOOR CALIBRATION, 2026-07-14 (MOVEMENT_FLOOR_CALIBRATION_001.md):
# speed=1 (current) live-confirmed visibly moving on Brobot 1 - no floor
# problem found for the arm at this speed. Do not lower this speed below 1
# without re-confirming live; it was never tested below 1. Leg duration
# tightened 0.8s -> 0.5s (ARM_TEST_LEG_HOLD_SECONDS), live-confirmed good.
ARM_TEST_LEG_HOLD_SECONDS = 0.36  # raise/lower leg duration - live-confirmed good 2026-07-14, see MOVEMENT_FLOOR_CALIBRATION_001.md

ARM_TEST_SEQUENCE = [
    {"speed": 0, "hold_seconds": 0.3},   # settle to rest (explicit start anchor)
    {"speed": 1, "hold_seconds": ARM_TEST_LEG_HOLD_SECONDS},   # raise, slow
    {"speed": 0, "hold_seconds": 1.2},   # pause held at top
    {"speed": -1, "hold_seconds": ARM_TEST_LEG_HOLD_SECONDS},  # lower, slow
    {"speed": 0, "hold_seconds": 0.8},   # settle to rest (explicit end anchor)
]

# HEAD_NOD_TEST_SEQUENCE - the prior doubled nod was described live as
# "wild": no rest anchor at the start (the head moved from wherever it
# happened to be) and no lead-in before the first nod. Fixed per the
# operator's own shape: settle to rest, a small lift up (announces "here
# it comes" before the first nod), a real pause, nod down/up, a real
# pause, a second nod down/up, then settle to rest - the final settle also
# serves as the last pause, so it isn't a redundant fourth zero-hold back
# to back with a separate "pause" step.
#
# *** MOTOR FLOOR - READ BEFORE CHANGING NOD_TEST_SPEED, 2026-07-14 ***
# See golden truth 1 at the top of this file. Bisected live, empirically,
# on Brobot 1, with the operator watching, one value at a time (full
# detail: MOVEMENT_FLOOR_CALIBRATION_001.md):
#   KNOWN-DEAD    (confirmed NO visible movement): 1.5, 2.0, 2.5, 2.75
#   KNOWN-WORKING (confirmed visible movement):     3.0
# NEVER drive this speed below 3 without a fresh, live, operator-witnessed
# re-confirmation. The "smaller motion" the operator wanted was achieved by
# tightening leg DURATION instead of speed (0.6s -> 0.45s, live-confirmed
# good) - speed stays at the floor, travel time is what shrank.
NOD_TEST_SPEED = 3             # nod leg speed magnitude - DO NOT LOWER below 3 without live re-confirmation, see warning above
NOD_TEST_LEAD_IN_SPEED = 2     # lead-in speed magnitude - scaled with NOD_TEST_SPEED (ratio ~0.667)
NOD_TEST_LEG_HOLD_SECONDS = 0.36  # nod-down/nod-up leg duration - live-confirmed good 2026-07-14, see MOVEMENT_FLOOR_CALIBRATION_001.md

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


# --- GESTURES: the interview's own performance moves, fired only through
# fire_scored_interview_movement() in run_section1_full_live_001.py, via
# the "arm_gesture"/"head_nod_gesture" scored movement names. Seeded as
# exact copies of the test values above at the moment of the split - free
# to be tuned by ear in a later pass without touching the tests. ---
#
# THE MOTOR FLOOR APPLIES HERE TOO - see golden truth 1 at the top of this
# file. NOD_GESTURE_SPEED must never drop below 3 without a fresh, live,
# operator-witnessed re-confirmation - confirmed dead at 2.75/2.5/2.0/1.5
# on this robot (MOVEMENT_FLOOR_CALIBRATION_001.md). The only lever for a
# smaller or more natural-feeling nod is TIMING - leg duration and pause
# length - never speed.
ARM_GESTURE_LEG_HOLD_SECONDS = 0.81  # seeded from ARM_TEST_LEG_HOLD_SECONDS 2026-07-14 - free to diverge, tune by ear

ARM_GESTURE_SEQUENCE = [
    {"speed": 0, "hold_seconds": 0.3},   # settle to rest (explicit start anchor)
    {"speed": 1, "hold_seconds": ARM_GESTURE_LEG_HOLD_SECONDS},   # raise, slow
    {"speed": 0, "hold_seconds": 1.2},   # pause held at top
    {"speed": -1, "hold_seconds": ARM_GESTURE_LEG_HOLD_SECONDS},  # lower, slow
    {"speed": 0, "hold_seconds": 0.8},   # settle to rest (explicit end anchor)
]

NOD_GESTURE_SPEED = 3             # nod leg speed magnitude - DO NOT LOWER below 3 without live re-confirmation, see warning above
NOD_GESTURE_LEAD_IN_SPEED = 2     # lead-in speed magnitude - scaled with NOD_GESTURE_SPEED (ratio ~0.667)
NOD_GESTURE_LEG_HOLD_SECONDS = 0.81  # seeded from NOD_TEST_LEG_HOLD_SECONDS 2026-07-14 - free to diverge, tune by ear (TIMING only, never speed)

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
    """Ports _brobots_move_axis's exact HTTP shape (brobots.sh) for a single
    robot: assume -> set speed -> hold -> set speed=0 -> release. Reused for
    both arm cues (move_lift) and nods (move_head). Moved here from
    run_songs_runner_001.py 2026-07-25 (REPORTER_GAP_SHARED_SWITCH_SURVEY_001.md)
    so brobots_awaken can share the same tunable cycles/hold_seconds/speed
    movement bingo already has, instead of this song's own fixed
    ARM_TEST_SEQUENCE/HEAD_NOD_TEST_SEQUENCE choreography (untouched, still
    used by the standalone test-arm-cue/test-head-nod aliases).
    manage_control=False skips the assume/release pair - this song's own
    "connect" note already assumes control once for the whole run (see
    say_line()'s own comment above on the screen-flash bug that caused);
    bingo calls this with the default (True), one assume/release per call,
    matching its own existing behavior exactly."""
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


# Stage 2 of WHEEL_NUDGE_GOLDEN_PATH_SURVEY_001.md's 3-stage golden path -
# the engine note type, built on the phcal bench primitive already live-
# confirmed on and off the charger (WHEEL_NUDGE_BENCH_PRIMITIVE_EXECUTED_001.md).
# lw/rw values ported from that same primitive, not reinvented - the
# webroot's own confirmed S-key/backward value.
WHEEL_NUDGE_REVERSE_LW = -150
WHEEL_NUDGE_REVERSE_RW = -150
WHEEL_NUDGE_DEFAULT_HOLD_SECONDS = 0.5
# WAKE mechanism, REPLACED 2026-08-10 (WHEEL_NUDGE_COLD_FIRE_ROOT_CAUSE_
# SURVEY_001.md + its live-diagnostic follow-up sessions). The prior
# say_text-based "speak-ready" check above was proven wrong by direct
# measurement, not superseded by theory: 4/4 live runs against a freshly-
# released robot showed say_text returning in ~0.05-0.12s with NO audible
# speech - it was never actually confirming anything, on either robot.
# The real, measured mechanism: assume, then RELEASE behavior control
# right back (a "release pulse"), settle, then RE-ASSUME. Releasing
# control hands it back to the robot's own onboard behavior just long
# enough for it to visibly/physically wake on its own - this mirrors
# ROBOT_SLEEP_DIRECT_SDK_BUILT_001.md's own first live finding almost
# exactly (releasing BehaviorControl right after a forced sleep trigger
# wakes the robot on its own, no explicit wake trigger needed) -
# generalized here from "coming out of a forced sleep" to "any cold/
# just-woken control session." Live-proven 8/8 runs across BOTH robots
# (once Brobot 2's own separate charging-contact fault was found and
# fixed - a real, physical, unrelated defect this same investigation
# surfaced): head-pop-to-visible-wake measured at 1.03-1.82s every time,
# the first move_wheels call succeeding at just 0.5s after re-assume in
# every single run, no misses.
BROBOTS_WAKE_CHECK_TIMEOUT_SECONDS = 15  # CONNECTION_PASS_TIMEOUT_SECONDS's own value, matched - conn_test's own timeout, unchanged
BROBOTS_WAKE_RELEASE_SETTLE_SECONDS = 2.5  # measured max observed 1.817s across 8 live runs, both robots - this adds real margin above that ceiling for an unattended/live-audience run rather than shipping the bare minimum; see WHEEL_NUDGE_COLD_FIRE_ROOT_CAUSE_SURVEY_001.md for the full run-by-run numbers
# BUG FOUND AND FIXED 2026-08-10 (same day, live production verification pass):
# the settle above covers release->reassume, but nothing ever covered
# reassume->move. The diagnostic probe's own sweep (probe_wheel_cold_fire_001.py)
# NEVER fired move_wheels at 0s after re-assume - its default delay list is
# "0.5,1,1.5,2,2.5,3", and every one of the 8/8 clean runs that "proved" this
# mechanism succeeded at attempt 1 = 0.501s (WHEEL_NUDGE_COLD_FIRE_ROOT_CAUSE_
# SURVEY_001.md line 199/214-215). That built-in 0.5s gap never made it into
# this production translation - move_wheels was firing 0ms after re-assume's
# HTTP response, i.e. before the async ControlGrantedResponse this whole
# investigation is about had actually landed. Live-confirmed 2026-08-10: wheel
# reversal silently no-opped on both robots (arm/nod, which never release/
# reassume mid-run, kept working) while every log line still read ok=True -
# exactly the "markers do not prove completion" trap. Fixed by adding the
# missing settle, matching the proven value.
BROBOTS_WAKE_POST_REASSUME_SETTLE_SECONDS = 0.5

# 2026-08-18, PHCAL_DETECT_FIRST_001.md live-test finding: same failure
# class as BROBOTS_WAKE_POST_REASSUME_SETTLE_SECONDS's own 2026-08-10 fix
# above (an action firing before the robot's real async control grant had
# landed, marker says success, nothing actually happens) - found in the
# "weather"/"say_phrase" notes' assume->say sequence, which had no settle
# at all. This is a COLD/first assume (not a reassume mid-hold), so it
# reuses PHCAL_ASSUME_SETTLE_SECONDS's own 1.5s value (phcal_isolate_001.py,
# its own "Cold-first-cycle fix" comment) rather than the smaller 0.5s
# reassume-specific number above - matched to the closer-in-kind case, not
# copied blind.
POST_ASSUME_SAY_SETTLE_SECONDS = 1.5


def run_brobots_wake(mod, serial, live, manage_control=True):
    """The GLOBAL, reusable "make a control session genuinely responsive"
    primitive - mechanism REPLACED 2026-08-10, see
    BROBOTS_WAKE_RELEASE_SETTLE_SECONDS above for the full why. conn_test
    (cheap connectivity check) -> if manage_control, assume -> release
    (the "golden flag" pulse) -> settle -> re-assume. Not wheel-specific -
    any future motion primitive that needs to prove a control session is
    genuinely live before acting on it can call this directly, the same
    way run_move_axis/run_arm_cue/run_nod already share manage_control's
    own convention. The release/settle/re-assume pulse fires
    UNCONDITIONALLY regardless of manage_control - a caller that's
    already holding control (manage_control=False, e.g. the composed
    run_wheel_nudge call site below) still needs this function to let go
    and re-grab control, since the golden flag only works as a genuine
    release-then-reassume transition, not a no-op skip."""
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
    """The move half. Assumes control is genuinely live - either its own
    manage_control=True assume, or (the real production path) run_brobots_
    wake's own release/settle/re-assume pulse having just landed. Fires
    ONE real move_wheels call - REPLACED 2026-08-10: the old "fire twice,
    back to back" duplicate (a workaround for a still-waiting-on-the-grant
    theory that was itself only ever tested inside a flat-sleep settle
    window) is removed, not kept as redundant insurance - the golden-flag
    mechanism above proved reliable with a SINGLE move call across 8/8
    live runs, both robots (WHEEL_NUDGE_COLD_FIRE_ROOT_CAUSE_SURVEY_001.md).
    Then hold, stop, release (if manage_control)."""
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
    """Composed: run_brobots_wake's own core (its own release/settle/
    re-assume golden-flag pulse, no outer release of its own) then
    run_move_reverse's own core (no assume - relies on wake's own
    re-assume having just landed), sharing ONE continuous held-control
    window opened/closed ONCE here. manage_control=False skips this
    function's own outer assume/release, same convention as run_move_axis
    above - a song holding control continuously (e.g. brobots_awaken)
    supplies its own assume at "connect" and its own release at "exit".
    Note-type name ("wheel_nudge") kept as-is - already git-tracked and
    wired into run_golden_song_001.py's dispatch; the reusable pieces
    underneath are named brobots_wake/move_reverse per operator direction,
    2026-08-09, mechanism replaced 2026-08-10 per live-measured diagnostic
    data (WHEEL_NUDGE_COLD_FIRE_ROOT_CAUSE_SURVEY_001.md)."""
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
    # suppress_assume_release=True - matching how Wire-Pod's own web control
    # page does it (webroot/sdkapp/js/control.js: control is assumed once
    # via the page's own toggle and held for the whole session; every
    # command after that, movement or say, fires directly with no per-call
    # assume/release). This song's own "connect" note assumes control once
    # at the very start; the run's own finally block releases it once at
    # the very end. Without this, every single spoken line would do its own
    # assume-then-release, repeatedly handing control back to the robot's
    # own onboard behavior between every note - confirmed live on Brobot 1
    # as the actual cause of the robot's screen flashing to its home/idle
    # icon after every call.
    result = robots.say(
        name, robot_safe_line, display_text=display_text, display_role=name,
        suppress_assume_release=True,
    )
    # The interview's own speech-time pronunciation swap ("GOPOD" ->
    # "Gowp-awd", apply_pronunciation_safety() via flatten_for_robot_speech())
    # already runs here, unconditionally, for every song - `robots` above is
    # mod.Robots(...), the exact same class the interview instantiates, not
    # a parallel implementation. result["speech_text"] is the corrected text
    # actually chunked into result["wirepod_chunks"] (what a live call POSTs
    # to /api-sdk/say_text); result["raw_speech_text"] is the literal,
    # pre-correction text - the same value the terminal print above and this
    # song's own story.md both show. This print is silent for any line whose
    # two values are identical - true for every existing song's text (none
    # contains a pronunciation-registry "from" word) - and only fires when
    # they diverge, proof the swap is speech-time-only, not a text change.
    speech_text = result.get("speech_text")
    raw_speech_text = result.get("raw_speech_text")
    if speech_text and raw_speech_text and speech_text != raw_speech_text:
        print(f"{name} PRONUNCIATION_SWAP: raw={raw_speech_text!r} spoken={speech_text!r}", flush=True)
    ok = result.get("spoken") is True or result.get("mode") in ("dry", "off")
    return {"ok": ok, "result": result}


def run_single_note(note, live, robot="1", voice_destination="robot", location=None, phrase=None):
    """Fire exactly one motion/note in isolation - no song loop, no
    narration, no other notes. Built for tuning one motion at a time
    (arm_test/head_nod/fireworks/weather/say_phrase) without waiting for
    the whole song, and without other notes' own assume/release churn
    muddying what's being watched. Reuses the exact same functions the full
    song calls - not a second implementation. voice_destination matters for
    weather and say_phrase (the two notes that speak through
    mod.Robots.say() rather than firing a bare movement/intent call
    directly) - added 2026-07-16 registry polish, see
    ALIAS_REGISTRY_POLISH_001.md; every existing arm_test/head_nod/fireworks
    caller is unaffected, since the default ("robot") matches main()'s own
    prior implicit behavior and those notes never construct a Robots
    instance at all. location=None (every caller except phcal's own bench
    weather test) preserves the original Windsor-only fetch - added
    2026-08-15, PHCAL_NOTES_WEATHER_WAKE_FIXES_001.md; ignored entirely by
    every note except weather. phrase (2026-08-18,
    PHCAL_DETECT_FIRST_001.md) is say_phrase's own required text - ignored
    entirely by every other note."""
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
        # Same real fetch + say_line() shape run_control_song's own "weather"
        # step uses (:562-574) - assume/release wrap it here since, unlike
        # arm_test/head_nod/fireworks, say_line()'s underlying Robots.say()
        # call is suppress_assume_release=True and expects a caller to hold
        # control already (matching webroot/sdkapp/js/control.js's own
        # assume-once convention, same reasoning documented on say_line()
        # itself).
        scaffold = mod.load_interview_scaffold(mod.INTERVIEW_SCAFFOLD_PATH)
        robots = mod.Robots(live, scaffold, voice_destination)
        try:
            if live:
                mod.wirepod_web_send_form(
                    "/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10
                )
                # 2026-08-18, PHCAL_DETECT_FIRST_001.md live-test finding:
                # this branch had NO explicit settle after assume - it only
                # "worked" because fetch_and_format_weather() below is a
                # live network call that happens to eat enough real time to
                # accidentally settle the robot's own control/audio channel
                # first. Confirmed the real cause by comparing against
                # say_phrase below, which has no such incidental delay and
                # produced a Wire-Pod "success" with zero audible speech,
                # live-tested repeatedly, same phrase text that DOES speak
                # via the direct-SDK path (rules out phrase content/length).
                # Explicit settle here removes the accidental dependency on
                # network-call timing - same POST_ASSUME_SAY_SETTLE_SECONDS
                # this file's own established assume-then-settle convention
                # uses elsewhere (BROBOTS_WAKE_POST_REASSUME_SETTLE_SECONDS/
                # PHCAL_ASSUME_SETTLE_SECONDS, other assume call sites in
                # this same file).
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
        # 2026-08-18, PHCAL_DETECT_FIRST_001.md: single-robot degrade of
        # phcal's own brobots_announce_in_sync primitive - built for a
        # session where detect-first found only one robot present, so the
        # two-robot direct-SDK "Brobots ready!" binary (which structurally
        # requires exactly two serials, confirmed by reading its own Go
        # source) can't be used. Same say_line()/assume-release shape as
        # "weather" above, just an operator-supplied phrase instead of a
        # weather fetch - reused, not duplicated.
        if not phrase:
            raise RuntimeError("BLOCKED: say_phrase requires a non-empty phrase")
        scaffold = mod.load_interview_scaffold(mod.INTERVIEW_SCAFFOLD_PATH)
        robots = mod.Robots(live, scaffold, voice_destination)
        try:
            if live:
                mod.wirepod_web_send_form(
                    "/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10
                )
                # Live-tested root cause of "Wire-Pod says success, robot
                # says nothing": no settle between assume and say - fired
                # immediately, unlike weather above which was accidentally
                # protected by its own network-fetch delay. Same explicit
                # fix as weather's own comment above - see that one for the
                # full finding.
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
    # The full-song loop (run_control_song()/load_control_song()) was
    # retired 2026-08-13 (CONTROL_SONG_LOOP_RETIRED_001.md) once control and
    # weather were both live-confirmed cut over to run_golden_song_001.py -
    # this script is a tester-only entry point now. The isolated_note branch
    # below is the ONLY thing test-arm-cue/test-head-nod/test-fireworks/
    # gopod-fireworks/gopod-weather-say actually call (each sets
    # GOPOD_CONTROL_SONG_NOTE before invoking `python3
    # run_robot_control_song_001.py` directly) - untouched, still the real
    # entry point for all five.
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

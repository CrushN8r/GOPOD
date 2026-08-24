#!/usr/bin/env python3
"""Applies phcal's last-confirmed arm/nod HOLD values into
run_robot_control_song_001.py's TEST and GESTURE leg-hold constants, closing
the loop between phcal (bench tuning) and pha0b (song playback) so a
confirmed value doesn't need hand re-typing into the runner.

Deliberately narrow:
  - Only touches *_LEG_HOLD_SECONDS constants (arm + nod, both TEST and
    GESTURE sets). Those are explicitly marked "free to diverge, tune by
    ear" in the runner's own comments - safe to overwrite by value.
  - NEVER touches speed. Nod speed already has an explicit "DO NOT LOWER
    below 3 without live re-confirmation" floor in the runner; arm speed
    isn't even a named constant there (hardcoded 1/-1 literal inside the
    sequence lists) - changing either is a live-hardware motor-speed
    decision, not a text substitution, so this tool only reports a
    speed mismatch and never writes one.
  - Never touches lead-in speed, pause durations, or sequence shape.

TEST vs. GESTURE separable, added 2026-07-25 (operator request,
alias-mixer/switch-cockpit widening pass): TEST backs the standalone
test-arm-cue/test-head-nod bench aliases; GESTURE backs the interview's own
live arm_gesture/head_nod_gesture movement (fire_scored_interview_movement()
in run_section1_full_live_001.py). These two were always meant to be free to
diverge ("free to diverge, tune by ear" in the runner's own comments) but
this tool applied both together, every time, with no way to touch just one -
--target below fixes that. Deliberately NOT wired anywhere near the
interview/GESTURE side yet (a standing memory flags these exact sequences as
live-confirmed-good, don't retune preemptively) - --target/--primitive exist
so a future, separately-authorized GESTURE wiring pass can reuse this same
tool rather than rebuilding it, not because GESTURE is being touched now.

Usage: python3 phcal_apply_control_song_001.py --brobot 1|2 [--yes] [--target test|gesture|both] [--primitive arm|nod|both]
  --brobot:     REQUIRED. Which brobot's phcal_last.json slot to read
                (PHCAL_ARROW_NAV_BUILD_PLAN_005.md lane (iv): phcal_last.json
                is robot-keyed now, {"1": {...}, "2": {...}}). This tool has
                no step/speaker context of its own to resolve this from (unlike
                phcal_apply_001.py, which reads a song step's own "speaker"
                field) - the caller (e.g. test-arm-cue/test-head-nod in
                brobots.sh, which already knows which robot it's testing) must
                pass it explicitly. No default - refuses to guess.
  No flag:      prints the diff and stops (dry).
  --yes:        prints the diff, then writes it.
  --target:     which constant set to touch. Default "both" - preserves
                this tool's original combined behavior exactly for any
                existing caller.
  --primitive:  which motion to touch. Default "both" - same backward-
                compatibility reasoning as --target.
"""
import json
import re
import sys
from pathlib import Path

LAST_PATH = Path("/home/goverlord/.gopod_alias_lib/phcal_last.json")
RUNNER_PATH = Path(
    "/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools/run_robot_control_song_001.py"
)

# (primitive, target) -> constant name. Both dimensions are independently
# filterable via --primitive/--target; default "both" on either walks every
# entry, reproducing the pre-split behavior exactly.
HOLD_CONSTANTS = {
    ("arm", "test"): "ARM_TEST_LEG_HOLD_SECONDS",
    ("arm", "gesture"): "ARM_GESTURE_LEG_HOLD_SECONDS",
    ("nod", "test"): "NOD_TEST_LEG_HOLD_SECONDS",
    ("nod", "gesture"): "NOD_GESTURE_LEG_HOLD_SECONDS",
}
SPEED_CONSTANTS = {
    ("nod", "test"): "NOD_TEST_SPEED",
    ("nod", "gesture"): "NOD_GESTURE_SPEED",
}


def _read_last():
    with open(LAST_PATH) as f:
        return json.load(f)


def _read_constant(text, name):
    m = re.search(rf"^{name}\s*=\s*([0-9.]+)", text, re.MULTILINE)
    if not m:
        raise RuntimeError(f"could not find constant {name} in {RUNNER_PATH}")
    return float(m.group(1))


def _write_constant(text, name, new_value):
    pattern = rf"^({name}\s*=\s*)[0-9.]+"
    return re.sub(pattern, rf"\g<1>{new_value}", text, count=1, flags=re.MULTILINE)


def _arg_value(name, default):
    if name in sys.argv:
        return sys.argv[sys.argv.index(name) + 1]
    return default


def _filtered(constants, target, primitive):
    return {
        (p, t): name
        for (p, t), name in constants.items()
        if (target == "both" or t == target) and (primitive == "both" or p == primitive)
    }


def main():
    apply = "--yes" in sys.argv
    target = _arg_value("--target", "both")
    primitive = _arg_value("--primitive", "both")
    brobot = _arg_value("--brobot", None)
    if target not in ("test", "gesture", "both"):
        print(f"PHCAL_APPLY_CONTROL_SONG_USAGE bad --target {target!r}, must be test|gesture|both")
        return 1
    if primitive not in ("arm", "nod", "both"):
        print(f"PHCAL_APPLY_CONTROL_SONG_USAGE bad --primitive {primitive!r}, must be arm|nod|both")
        return 1
    if brobot not in ("1", "2"):
        print(
            f"PHCAL_APPLY_CONTROL_SONG_BLOCKED --brobot {brobot!r} missing or invalid (must be 1|2) - "
            "this tool has no step/speaker context to resolve which brobot's phcal_last.json slot to "
            "read; refusing to guess"
        )
        return 1

    all_last = _read_last()
    last = all_last.get(brobot) or {}
    text = RUNNER_PATH.read_text()

    hold_constants = _filtered(HOLD_CONSTANTS, target, primitive)
    speed_constants = _filtered(SPEED_CONSTANTS, target, primitive)

    print(f"PHCAL_APPLY_CONTROL_SONG diff (target={target} primitive={primitive}, hold_seconds only, speed never written):")
    changes = []
    for (prim, _tgt), name in hold_constants.items():
        new_hold = last[prim]["hold"]
        old_hold = _read_constant(text, name)
        flag = "" if old_hold == new_hold else "  <-- CHANGES"
        print(f"  {name}: {old_hold} -> {new_hold}{flag}")
        if old_hold != new_hold:
            changes.append((name, new_hold))

    print("\nPHCAL_APPLY_CONTROL_SONG speed check (report only, never written):")
    for (prim, _tgt), name in speed_constants.items():
        new_speed = last[prim]["speed"]
        old_speed = _read_constant(text, name)
        if old_speed != new_speed:
            print(f"  {name}: currently {old_speed}, phcal last confirmed {new_speed} - MISMATCH, not applied")
        else:
            print(f"  {name}: currently {old_speed}, matches phcal last confirmed {new_speed} - no change needed")
    if primitive in ("arm", "both"):
        print(
            "  ARM speed: phcal last confirmed "
            f"{last['arm']['speed']} - arm speed is not a named constant in this runner "
            "(hardcoded 1/-1 literal in ARM_TEST_SEQUENCE/ARM_GESTURE_SEQUENCE) - not applied, not reportable as a clean diff"
        )

    if not changes:
        print("\nPHCAL_APPLY_CONTROL_SONG_NOOP nothing to change")
        return 0

    if not apply:
        print("\nPHCAL_APPLY_CONTROL_SONG_DRY re-run with --yes to write the above hold_seconds changes")
        return 0

    for name, new_value in changes:
        text = _write_constant(text, name, new_value)
    RUNNER_PATH.write_text(text)
    print(f"\nPHCAL_APPLY_CONTROL_SONG_WRITTEN {len(changes)} constant(s) updated in {RUNNER_PATH}")
    return 0


if __name__ == "__main__":
    sys.exit(main())

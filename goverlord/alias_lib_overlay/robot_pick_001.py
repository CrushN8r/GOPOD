#!/usr/bin/env python3
"""Robot-picker v1 (WIDGET_BOUNDARY_SURVEY_001.md Part D's recommended first
buildable step) - reassigns a single step's `speaker` field between
brobot_1/brobot_2, writing through the same proven load/mutate/write path
phcal_apply_001.py uses: the shared knobs_envelope_001.load_knobs_envelope()
resolver (dirty zKnobs.json-over-clean knobs.json, whole-file read), then a
whole-file json.dump(indent=2) back. Pure surfacing of a field the golden
engine (run_golden_song_001.py) already reads per step and routes to a robot
serial via _serial_and_name() - no new writer, no new format, no engine
change. This is a second CALLER of the shared resolver, same as
phcal_apply_001.py/print_song_score_001.py/the panel scanner each already
are independently - not a second writer mechanism.

Scoped to the bingo-family note-dispatch songs where `speaker` is a real,
live-read per-step field (nap/mixup/bingo, invoked from pha0b). Deliberately
NOT offered for 00_brobots_awaken: confirmed in SONG_REGISTRY
(run_golden_song_001.py), brobots_awaken_golden sets
synthesize_speaker=True, meaning every step's own knobs.json "speaker"
field is unused dead data there - the real speaker for that song comes from
a whole-run robot choice (GOPOD_GOLDEN_ROBOT), not per-step. Writing this
field for awaken would silently do nothing - out of v1 by this same honesty
standard, not a missing feature.

Usage: python3 robot_pick_001.py <step_id> <brobot_1|brobot_2> [--yes] [--knobs <path>]
  No --yes: prints the diff and stops (dry).
  --yes:    prints the diff, then writes it.
  --knobs:  path to the target song's (clean) knobs.json, required. Same
            dirty-over-clean resolution as phcal_apply_001.py - if a
            sibling zKnobs.json exists in the same directory, THAT file is
            read and written instead.
"""
import json
import sys
from pathlib import Path

sys.path.insert(0, "/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools")
from knobs_envelope_001 import load_knobs_envelope

VALID_SPEAKERS = ("brobot_1", "brobot_2")


def main():
    if len(sys.argv) < 3:
        print("ROBOT_PICK_USAGE robot_pick_001.py <step_id> <brobot_1|brobot_2> [--yes] [--knobs <path>]")
        return 1
    step_id = sys.argv[1]
    new_speaker = sys.argv[2]
    if new_speaker not in VALID_SPEAKERS:
        print(f"ROBOT_PICK_BLOCKED speaker {new_speaker!r} not one of {VALID_SPEAKERS}")
        return 1
    apply = "--yes" in sys.argv
    if "--knobs" not in sys.argv:
        print("ROBOT_PICK_USAGE --knobs <path to song's knobs.json> is required")
        return 1
    knobs_path = Path(sys.argv[sys.argv.index("--knobs") + 1])

    # Same dirty(zKnobs)-over-clean(knobs) preference every playback engine
    # and phcal_apply_001.py already apply - via the shared resolver, not a
    # private re-check. Read and write both resolve to whatever
    # load_knobs_envelope() picked, so they can't disagree with each other
    # or with what the engine above will actually play.
    KNOBS_PATH, envelope = load_knobs_envelope(knobs_path.parent)

    step = next((s for s in envelope.get("steps", []) if s.get("step_id") == step_id), None)
    if step is None:
        print(f"ROBOT_PICK_BLOCKED step_id {step_id!r} not found in {KNOBS_PATH}")
        return 1
    if "speaker" not in step:
        print(f"ROBOT_PICK_BLOCKED step_id {step_id!r} has no speaker field - not a plain-data-speaker step")
        return 1

    old_speaker = step["speaker"]
    print(f"ROBOT_PICK diff for step_id={step_id!r}:")
    flag = "" if old_speaker == new_speaker else "  <-- CHANGES"
    print(f"  speaker: {old_speaker} -> {new_speaker}{flag}")

    if not apply:
        print("\nROBOT_PICK_DRY re-run with --yes to write the above into knobs.json")
        return 0

    # Whole-file rewrite, same mechanism as phcal_apply_001.py (Ladder-2
    # Phase 5) - `step` is the actual dict object inside envelope["steps"],
    # so mutating it here already carries into the dump below. Every other
    # step's own field order/values are untouched.
    step["speaker"] = new_speaker
    KNOBS_PATH.write_text(json.dumps(envelope, indent=2) + "\n", encoding="utf-8")
    print(f"\nROBOT_PICK_WRITTEN step_id={step_id!r} updated in {KNOBS_PATH}")
    return 0


if __name__ == "__main__":
    sys.exit(main())

#!/usr/bin/env python3
"""."""
import json
import sys
from pathlib import Path

sys.path.insert(0, "/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools")
from knobs_envelope_001 import load_knobs_envelope, save_knobs_envelope

LAST_PATH = Path("/home/goverlord/.gopod_alias_lib/phcal_last.json")
DEFAULT_KNOBS_PATH = Path(
    "/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/brobots_bingo/knobs.json"
)

NOTE_TO_PRIMITIVE = {
    "arm_cue": "arm",
    "nod": "nod",
    "head_nod": "nod",
    "animation": "animation",
    "wheel_nudge": "move_reverse",
    "brobots_ready_together": "brobots_announce_in_sync",
}

PRIMITIVE_FIELD_MAP = {
    "arm": {"cycles": "cycles", "hold_seconds": "hold", "speed": "speed"},
    "nod": {"hold_seconds": "hold", "speed": "speed"},
    "animation": {"hold_seconds": "hold"},
    "move_reverse": {"hold_seconds": "hold"},
    "brobots_announce_in_sync": {"text": "phrase"},
}


def main():
    if len(sys.argv) < 2:
        print("PHCAL_APPLY_USAGE phcal_apply_001.py <step_id> [--yes] [--knobs <path>]")
        return 1
    step_id = sys.argv[1]
    apply = "--yes" in sys.argv
    knobs_path = DEFAULT_KNOBS_PATH
    if "--knobs" in sys.argv:
        knobs_path = Path(sys.argv[sys.argv.index("--knobs") + 1])
    KNOBS_PATH, envelope = load_knobs_envelope(knobs_path.parent)

    with open(LAST_PATH) as f:
        last = json.load(f)

    step = next((s for s in envelope.get("steps", []) if s.get("step_id") == step_id), None)
    if step is None:
        print(f"PHCAL_APPLY_BLOCKED step_id {step_id!r} not found in {KNOBS_PATH}")
        return 1

    note = step.get("note")
    if note not in NOTE_TO_PRIMITIVE:
        print(f"PHCAL_APPLY_BLOCKED step_id {step_id!r} has note={note!r} - only {'/'.join(NOTE_TO_PRIMITIVE)} steps are phcal-tunable")
        return 1

    speaker = step.get("speaker")
    which = {"brobot_1": "1", "brobot_2": "2"}.get(speaker)
    if which is None:
        print(
            f"PHCAL_APPLY_BLOCKED step_id {step_id!r} has speaker={speaker!r} in {KNOBS_PATH} "
            "- cannot resolve which brobot's phcal_last.json slot to read; refusing to guess"
        )
        return 1

    primitive = NOTE_TO_PRIMITIVE[note]
    field_map = PRIMITIVE_FIELD_MAP[primitive]
    robot_last = last.get(which) or {}
    if primitive not in robot_last:
        print(
            f"PHCAL_APPLY_BLOCKED no saved {primitive!r} tuning for brobot {which} in {LAST_PATH} "
            "- run phcal for that robot first"
        )
        return 1
    confirmed = robot_last[primitive]

    old = dict(step)
    for song_field, last_field in field_map.items():
        step[song_field] = confirmed[last_field]

    print(f"PHCAL_APPLY diff for step_id={step_id!r} (note={note}):")
    for song_field in field_map:
        old_val = old.get(song_field, "(unset, will use runner default)")
        new_val = step[song_field]
        flag = "" if old_val == new_val else "  <-- CHANGES"
        print(f"  {song_field}: {old_val} -> {new_val}{flag}")

    if not apply:
        print("\nPHCAL_APPLY_DRY re-run with --yes to write the above into knobs.json")
        return 0

    save_knobs_envelope(KNOBS_PATH, envelope)  # knobs.json, and zKnobs.md when the song has one
    print(f"\nPHCAL_APPLY_WRITTEN step_id={step_id!r} updated in {KNOBS_PATH}")
    return 0


if __name__ == "__main__":
    sys.exit(main())

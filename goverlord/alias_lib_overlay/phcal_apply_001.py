#!/usr/bin/env python3
"""Links phcal (bench tuning) to pha0b (song playback) for any control-song-
family song's own knobs.json - originally bingo-only (hence the old
filename, phcal_apply_bingo_001.py), widened 2026-07-25 (operator request,
REPORTER_GAP_SHARED_SWITCH_SURVEY_001.md) once brobots_awaken's
arm_test/head_nod notes also started reading cycles/hold_seconds/speed from
their own knobs.json (run_arm_cue()/run_nod(), now shared in
run_robot_control_song_001.py) instead of a fixed sequence. Renamed
2026-08-07 to phcal_apply_001.py, once the name mismatch got confusing
enough to fix - every caller (brobots.sh) updated in the same pass. The
print tags were left carrying the old PHCAL_APPLY_BINGO_* prefix at that
rename, out of scope for it - fixed to the song-neutral PHCAL_APPLY_*
2026-08-08 (cosmetic cleanup sweep) once a real awaken run showed the
stale "BINGO" label.
Writes phcal's
last-confirmed arm/nod values straight into a chosen step in that song's
own knobs.json. No source-code patching, no separate config file: one
song, one knobs.json, two tools pointed at the same place.

Ladder-2 Phase 5 (gopod_notes/older_notes/KNOBS_ROOT_UNTANGLE_SURVEY_001.md
§B, PHCAL_CRASH_AND_AWAKEN_ENGINE_DIAGNOSIS_001.md): reads the WHOLE file via
the shared knobs_envelope_001.load_knobs_envelope() (same resolver every
playback engine already uses) rather than a line-by-line text scan, locates
the target step by step_id in the parsed steps array, mutates only that
step's dict, and writes the whole envelope back with json.dump. This retires
the old line-scan, which assumed one step was one line of text - true for
bingo's compact single-line-per-step style but false for awaken's
pretty-printed one-field-per-line style, where it threw
"json.decoder.JSONDecodeError: Extra data" and silently never wrote the
tune (a false-clean: the tool printed a diff and the caller believed it
landed). Whole-file json.load can't choke on either style - that fragility
is gone by construction, not patched around.

Tradeoff, stated plainly: a whole-file rewrite reformats every step in the
file to Python's own json.dump(indent=2) style, not just the one touched -
a larger diff than the old single-line splice for a file that WAS compact
(bingo's zKnobs.json). Accepted because these are dirty/zKnobs rehearsal
files, gitignored, never part of the public repo's history
(`goverlord/runtime/songs/*/zKnobs.json` in .gitignore) - the diff is real
on disk but invisible to git. The one exception is awaken, which has no
zKnobs.json sibling and so writes straight to its own git-tracked
knobs.json - checked directly, though: that file's existing hand-formatting
already matches json.dumps(indent=2) byte-for-byte (confirmed by diffing a
round-trip load+dump against the file on disk - the only difference was a
missing trailing newline), so in practice this write produces a normal,
readable diff there too, not a reformatting flood.

Usage: python3 phcal_apply_001.py <step_id> [--yes] [--knobs <path>]
  No --yes: prints the diff and stops (dry).
  --yes:    prints the diff, then writes it.
  --knobs:  path to the target song's (clean) knobs.json. Defaults to
            brobots_bingo's own (unchanged from before this song was
            generalized) when omitted. If a sibling zKnobs.json exists in
            the same directory, THAT file is read and written instead -
            the same dirty(zKnobs)-over-clean(knobs) preference every
            playback engine already applies (KNOBS_TOUCHERS_MAP_001.md),
            now delegated to the shared resolver instead of a private
            re-implementation of the same 2-line check. Read and write
            still always resolve to the same file (see main()) - that part
            is unchanged by this pass, only how the read/write itself
            happens changed.
"""
import json
import sys
from pathlib import Path

sys.path.insert(0, "/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools")
from knobs_envelope_001 import load_knobs_envelope

LAST_PATH = Path("/home/goverlord/.gopod_alias_lib/phcal_last.json")
DEFAULT_KNOBS_PATH = Path(
    "/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/brobots_bingo/knobs.json"
)

# note name -> phcal_last.json primitive key. "head_nod" is brobots_awaken's
# own note name for the same physical motion bingo calls "nod" - both read
# the same "nod" entry in phcal_last.json (the golden engine's own
# note_aliases: {"head_nod": "nod"} does this same remap at dispatch time,
# for the same reason - this tool reads the raw step JSON before that
# alias ever runs, so it needs to recognize both spellings itself).
#
# Widened 2026-08-16 (MASTER_TWEAKS_STAGE4_APPLY_SEAM_001.md) to match
# stage 1's 9-primitive save-coverage - only the primitives with a REAL
# matching song note type AND a real tunable field on that note are
# listed here. Confirmed against run_golden_song_001.py's own note
# dispatch, not guessed:
#   animation -> "animation" note, reads step["hold_seconds"]
#   move_reverse -> "wheel_nudge" note, reads step["hold_seconds"]
#     (no song currently authors a wheel_nudge step - correct and
#     future-proof, verified with a synthetic step, not a live one)
#   brobots_announce_in_sync -> "brobots_ready_together" note, reads
#     step["text"] (a different field NAME than phcal_last.json's own
#     "phrase" key - see PRIMITIVE_FIELD_MAP below)
# Deliberately NOT listed (confirmed, not guessed - each has either no
# matching note type at all, or a note type that reads no step-sourced
# tunable field):
#   rattle - "rattle" note exists, but run_rattle() takes no step
#     parameter at all (fixed binary call) - nothing to write into.
#   danger - no matching note type anywhere in the engine's dispatch.
#   brobots_stay_in_place - no matching note type (bench-only diagnostic).
#   brobots_sleep_to_wake_direct_sdk - no matching note type (bench-only
#     direct-SDK mechanism, no song equivalent).
NOTE_TO_PRIMITIVE = {
    "arm_cue": "arm",
    "nod": "nod",
    "head_nod": "nod",
    "animation": "animation",
    "wheel_nudge": "move_reverse",
    "brobots_ready_together": "brobots_announce_in_sync",
}

# primitive -> {song step field: phcal_last.json field}. Declarative -
# replaces the old `if primitive == "arm": step["cycles"] = ...`
# special-case entirely. A primitive not saving a given field (nod never
# had "cycles") simply isn't listed for that field; a field NAME
# mismatch (brobots_announce_in_sync's "phrase" -> the song's own
# "text") is expressed directly in the map, not a special-cased if.
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
    # Same dirty(zKnobs)-over-clean(knobs) preference every playback engine
    # already applies (KNOBS_TOUCHERS_MAP_001.md) - now via the shared
    # resolver, not a private re-check. Read and write both resolve to
    # whatever load_knobs_envelope() picked, so they can't disagree with
    # each other or with what the engines above will actually play.
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

    primitive = NOTE_TO_PRIMITIVE[note]
    field_map = PRIMITIVE_FIELD_MAP[primitive]
    confirmed = last[primitive]

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

    # Whole-file rewrite (Ladder-2 Phase 5) - `step` is the actual dict object
    # inside envelope["steps"], mutated above, so this dump already carries
    # the change. Every other step's own field order/values are untouched -
    # only the identity of the mutated dict differs from what was loaded.
    KNOBS_PATH.write_text(json.dumps(envelope, indent=2) + "\n", encoding="utf-8")
    print(f"\nPHCAL_APPLY_WRITTEN step_id={step_id!r} updated in {KNOBS_PATH}")
    return 0


if __name__ == "__main__":
    sys.exit(main())

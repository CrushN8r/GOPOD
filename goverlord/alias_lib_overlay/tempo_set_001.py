#!/usr/bin/env python3
"""Tempo Phase 2 (TEMPO_BUFFER_KNOB_SURVEY_001.md / TEMPO_BUFFER_KNOB_PHASE1_EXECUTED_001.md) -
the tool that SETS the tempo knob values Phase 1 already made run_golden_song_001.py read.
Phase 1 added exactly two things to the engine: a top-level `global_tempo` read (default 0.0)
and a per-step `tempo_factor` read (default 1.0, rides the wholesale step-merge for free) plus
a `tempo_comment` field the engine never reads (human-facing only). This tool writes those same
three fields, by the same field names, at the same schema locations Phase 1 established -
top-level for global_tempo, per-step for tempo_factor/tempo_comment - so a value written here is
the exact value the engine picks up on its next run, no translation layer.

Writes through the SAME shared path phcal_apply_001.py and robot_pick_001.py already use, not a
new writer: knobs_envelope_001.load_knobs_envelope() (dirty zKnobs.json-over-clean knobs.json
resolution, whole-file json.load), mutate the loaded dict in place, whole-file
json.dumps(indent=2)+"\n" back to the SAME path that was read. This is a third caller of that
resolver, same as robot_pick_001.py's own docstring already frames itself as being ("not a
second writer mechanism").

Two modes, kept deliberately distinct (never blurred in one flow):

  set-global <value> [--yes] [--knobs <path>]
    Sets the song's own top-level global_tempo (0.0-9.9, per the locked design's own range -
    one unit = one second of base added space per buffer, before a step's own factor applies).
    This is the "slow the whole song slightly" dial - one number, whole song.

  set-buffer <step_id> --factor <f> [--comment <c>] [--yes] [--knobs <path>]
    Sets ONE step's own tempo_factor (>= 0.0, no upper bound - a per-step weight against the
    global, default 1.0) and, only if --comment is passed, that step's own tempo_comment
    (human-facing only, never read by the engine). Omitting --comment leaves whatever comment
    is already on that step untouched - this mode never clobbers a field it wasn't asked to
    touch.

Both modes print the diff and the resulting math (global x factor = seconds added) before
writing, and print again after --yes actually writes it. No --yes: dry, prints and stops,
same convention as phcal_apply_001.py/robot_pick_001.py.

Reversibility: global_tempo=0.0 and tempo_factor=1.0 are the engine's own defaults
(`knobs.get("global_tempo", 0.0)` / `step.get("tempo_factor", 1.0)`) - writing those exact
values back is behaviorally identical to the field never having existed at all. This tool never
strips keys to "clean up" back to absent; writing the default value IS the no-op, by the same
construction Phase 1 already relies on.
"""
import json
import sys
from pathlib import Path

sys.path.insert(0, "/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools")
from knobs_envelope_001 import load_knobs_envelope

GLOBAL_TEMPO_MIN = 0.0
GLOBAL_TEMPO_MAX = 9.9


def _parse_float(raw, label):
    try:
        return float(raw)
    except (TypeError, ValueError):
        print(f"TEMPO_SET_BLOCKED {label} {raw!r} is not a number")
        return None


def _resolve_knobs_path(argv):
    if "--knobs" not in argv:
        print("TEMPO_SET_USAGE --knobs <path to song's knobs.json> is required")
        return None
    knobs_path = Path(argv[argv.index("--knobs") + 1])
    # Same dirty(zKnobs)-over-clean(knobs) preference every playback engine,
    # phcal_apply_001.py, and robot_pick_001.py already apply - via the shared
    # resolver, not a private re-check. Read and write both resolve to
    # whatever load_knobs_envelope() picked.
    return load_knobs_envelope(knobs_path.parent)


def cmd_set_global(argv):
    if len(argv) < 2:
        print("TEMPO_SET_USAGE tempo_set_001.py set-global <value> [--yes] [--knobs <path>]")
        return 1
    value = _parse_float(argv[1], "global_tempo")
    if value is None:
        return 1
    if not (GLOBAL_TEMPO_MIN <= value <= GLOBAL_TEMPO_MAX):
        print(f"TEMPO_SET_BLOCKED global_tempo={value} out of range [{GLOBAL_TEMPO_MIN}, {GLOBAL_TEMPO_MAX}]")
        return 1
    apply = "--yes" in argv
    resolved = _resolve_knobs_path(argv)
    if resolved is None:
        return 1
    KNOBS_PATH, envelope = resolved

    old_value = envelope.get("global_tempo", 0.0)
    song_id = envelope.get("song_id", "?")
    print(f"TEMPO_SET_GLOBAL diff for song_id={song_id!r} ({KNOBS_PATH}):")
    flag = "" if old_value == value else "  <-- CHANGES"
    print(f"  global_tempo: {old_value} -> {value}{flag}")
    print(f"  every step's own added-space = {value} (global) x that step's own tempo_factor "
          f"(default 1.0 if unset) - excluded on 'exit', wherever a fatal note (wake_both/"
          f"emotion_beat) fails first")
    print(f"  e.g. a step with no explicit tempo_factor now adds {value}s per buffer")

    custom_factor_steps = [
        s for s in envelope.get("steps", [])
        if s.get("tempo_factor", 1.0) != 1.0 and s.get("note") != "exit"
    ]
    if custom_factor_steps:
        print("  steps with a non-default tempo_factor already set:")
        for s in custom_factor_steps:
            factor = s.get("tempo_factor", 1.0)
            print(f"    {s.get('step_id')}: factor={factor} -> adds {value} x {factor} = {value * factor}s")

    if not apply:
        print("\nTEMPO_SET_DRY re-run with --yes to write the above into knobs.json")
        return 0

    envelope["global_tempo"] = value
    KNOBS_PATH.write_text(json.dumps(envelope, indent=2) + "\n", encoding="utf-8")
    print(f"\nTEMPO_SET_WRITTEN global_tempo={value} written to {KNOBS_PATH}")
    return 0


def cmd_set_buffer(argv):
    if len(argv) < 2:
        print("TEMPO_SET_USAGE tempo_set_001.py set-buffer <step_id> --factor <f> [--comment <c>] [--yes] [--knobs <path>]")
        return 1
    step_id = argv[1]
    if "--factor" not in argv:
        print("TEMPO_SET_USAGE --factor <value> is required for set-buffer")
        return 1
    factor = _parse_float(argv[argv.index("--factor") + 1], "tempo_factor")
    if factor is None:
        return 1
    if factor < 0.0:
        print(f"TEMPO_SET_BLOCKED tempo_factor={factor} must be >= 0.0 (a negative value would "
              f"produce a negative sleep and crash the run)")
        return 1
    comment = None
    if "--comment" in argv:
        comment = argv[argv.index("--comment") + 1]
    apply = "--yes" in argv
    resolved = _resolve_knobs_path(argv)
    if resolved is None:
        return 1
    KNOBS_PATH, envelope = resolved

    step = next((s for s in envelope.get("steps", []) if s.get("step_id") == step_id), None)
    if step is None:
        available = ", ".join(s.get("step_id", "?") for s in envelope.get("steps", []))
        print(f"TEMPO_SET_BLOCKED step_id {step_id!r} not found in {KNOBS_PATH}")
        print(f"  available step_ids: {available}")
        return 1
    if step.get("note") == "exit":
        print(f"TEMPO_SET_BLOCKED step_id {step_id!r} is the 'exit' note - tempo is always "
              f"excluded there by the engine's own tail hook, setting a factor here would do nothing")
        return 1

    old_factor = step.get("tempo_factor", 1.0)
    old_comment = step.get("tempo_comment", "(unset)")
    global_tempo = envelope.get("global_tempo", 0.0)

    print(f"TEMPO_SET_BUFFER diff for step_id={step_id!r} (note={step.get('note')}):")
    flag = "" if old_factor == factor else "  <-- CHANGES"
    print(f"  tempo_factor: {old_factor} -> {factor}{flag}")
    if comment is not None:
        flag = "" if old_comment == comment else "  <-- CHANGES"
        print(f"  tempo_comment: {old_comment!r} -> {comment!r}{flag}")
    else:
        print(f"  tempo_comment: {old_comment!r} (unchanged - no --comment passed)")
    added = global_tempo * factor
    print(f"  resulting math: this song's own global_tempo is currently {global_tempo} -> "
          f"{global_tempo} x {factor} = {added}s added here")

    if not apply:
        print("\nTEMPO_SET_DRY re-run with --yes to write the above into knobs.json")
        return 0

    step["tempo_factor"] = factor
    if comment is not None:
        step["tempo_comment"] = comment
    KNOBS_PATH.write_text(json.dumps(envelope, indent=2) + "\n", encoding="utf-8")
    print(f"\nTEMPO_SET_WRITTEN step_id={step_id!r} updated in {KNOBS_PATH}")
    return 0


def main():
    if len(sys.argv) < 2 or sys.argv[1] not in ("set-global", "set-buffer"):
        print("TEMPO_SET_USAGE tempo_set_001.py set-global <value> [--yes] [--knobs <path>]")
        print("TEMPO_SET_USAGE tempo_set_001.py set-buffer <step_id> --factor <f> [--comment <c>] [--yes] [--knobs <path>]")
        return 1
    mode = sys.argv[1]
    argv = sys.argv[1:]
    if mode == "set-global":
        return cmd_set_global(argv)
    return cmd_set_buffer(argv)


if __name__ == "__main__":
    sys.exit(main())

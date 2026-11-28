#!/usr/bin/env python3
"""Tempo Phase 2 - the tool that SETS the tempo knob values Phase 1 already made run_golden_song_001.py read."""
import json
import sys
from pathlib import Path

sys.path.insert(0, "/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools")
from knobs_envelope_001 import load_knobs_envelope, save_knobs_envelope

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
    print(f"  global_tempo is EXTRA time ADDED on top of each step's own buffer_after - it never "
          f"scales or replaces buffer_after itself. 0.0 = neutral, no extra time added.")
    print(f"  every step's own added_space = {value} (this global_tempo) x that step's own "
          f"tempo_factor (default 1.0 if unset) - a flat number of extra seconds, not a "
          f"multiplier applied to the buffer. Excluded on 'exit' and 'pause' notes (tempo-immune "
          f"by design, REPORTER_GAP_TEMPO_IMMUNE_001.md), and wherever a fatal note "
          f"(wake_both/emotion_beat) fails first")
    print(f"  e.g. a step with no explicit tempo_factor (defaults to 1.0) adds exactly {value}s of "
          f"extra time on top of its own buffer_after")

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
    save_knobs_envelope(KNOBS_PATH, envelope)  # knobs.json, and zKnobs.md when the song has one
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
    save_knobs_envelope(KNOBS_PATH, envelope)  # knobs.json, and zKnobs.md when the song has one
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

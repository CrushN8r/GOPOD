"""Shared knobs file-resolution + envelope-load — Ladder 2, Phase 1
(gopod_notes/KNOBS_ROOT_UNTANGLE_SURVEY_001.md, Part D). The two functions
every one of the 7 duplicated zKnobs-preferred readers
(gopod_notes/KNOBS_TOUCHERS_MAP_001.md, #1-7) independently reimplement
today. This pass wires ONE caller onto it (print_song_score_001.py, #7,
display-only) — every other reader stays on its own private copy until a
later, separate repoint.

Envelope only: resolves zKnobs.json-over-knobs.json (identical to every
existing copy — filesystem check only, never opens the file to decide),
then whole-file json.load of whichever path won. No layout assumption
beyond the shared outer shape confirmed common across all 3 schema
families (song_id, song_type, story_ref, steps[]) — this module does not
parse story.md, does not write, and does not know about note types or any
per-song inner-step schema. Those stay exactly where they are (the 3
story parsers legitimately differ; survey Part B).
"""
import json
from pathlib import Path


def resolve_active_knobs_path(song_dir):
    """zKnobs.json if present, else knobs.json. Same 2-line check every
    existing loader (run_golden_song_001.py's load_golden_song(),
    run_songs_runner_001.py's load_bingo_capture_song(),
    run_robot_control_song_001.py's load_control_song(), and others)
    already runs — filesystem check only, never opens either file."""
    song_dir = Path(song_dir)
    dirty_path = song_dir / "zKnobs.json"
    if dirty_path.exists():
        return dirty_path
    return song_dir / "knobs.json"


def load_knobs_envelope(song_dir):
    """Returns (resolved_path, parsed_dict) via whole-file json.load of
    whatever resolve_active_knobs_path() picked. No per-schema handling —
    every schema family shares this same outer envelope."""
    resolved_path = resolve_active_knobs_path(song_dir)
    parsed = json.loads(resolved_path.read_text(encoding="utf-8"))
    return resolved_path, parsed

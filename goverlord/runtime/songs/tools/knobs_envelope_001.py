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
import shutil
from pathlib import Path


def resolve_active_knobs_path(song_dir):
    """knobs.json, the clean copy (operator 2026-10-11: "go for knobs.json
    (clean version of dirty zKnobs) * if no clean knobs, create a knobs file
    from zKnobs and then go for knobs.json again"). Before 2026-10-11 a
    zKnobs.json won when present."""
    song_dir = Path(song_dir)
    clean_path = song_dir / "knobs.json"
    dirty_path = song_dir / "zKnobs.json"
    if not clean_path.exists() and dirty_path.exists():
        shutil.copyfile(dirty_path, clean_path)
    return clean_path


def load_knobs_envelope(song_dir):
    """Returns (resolved_path, parsed_dict) via whole-file json.load of
    whatever resolve_active_knobs_path() picked. No per-schema handling —
    every schema family shares this same outer envelope."""
    resolved_path = resolve_active_knobs_path(song_dir)
    parsed = json.loads(resolved_path.read_text(encoding="utf-8"))
    return resolved_path, parsed

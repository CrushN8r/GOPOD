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

from knobs_md_001 import knobs_from_md, md_from_knobs


def resolve_active_knobs_path(song_dir):
    """knobs.json, the clean copy (operator 2026-10-11: "go for knobs.json
    (clean version of dirty zKnobs) * if no clean knobs, create a knobs file
    from zKnobs and then go for knobs.json again"). Before 2026-10-11 a
    zKnobs.json won when present.

    2026-11-04 (operator: "zKnobs.md is the one source"; step 9 of
    GOPOD_CLUSTERS_PLAN_002.md): with a zKnobs.md, knobs.json is built from it
    first - written only when its knobs differ, so an unchanged song leaves
    knobs.json as it is (knobs_md_001.py)."""
    song_dir = Path(song_dir)
    clean_path = song_dir / "knobs.json"
    md_path = song_dir / "zKnobs.md"
    dirty_path = song_dir / "zKnobs.json"
    if md_path.exists():
        built = knobs_from_md(md_path.read_text(encoding="utf-8"))
        try:
            same = json.loads(clean_path.read_text(encoding="utf-8")) == built
        except (OSError, ValueError):
            same = False
        if not same:
            clean_path.write_text(json.dumps(built, indent=2) + "\n", encoding="utf-8")
        return clean_path
    if not clean_path.exists() and dirty_path.exists():
        shutil.copyfile(dirty_path, clean_path)
    return clean_path


def save_knobs_envelope(knobs_path, envelope):
    """2026-11-04 (step 9): a tool's knobs written back - knobs.json, and the
    song's zKnobs.md when it has one, so the one source holds the change too
    (phcal apply, tempo-set, robot pick)."""
    knobs_path = Path(knobs_path)
    knobs_path.write_text(json.dumps(envelope, indent=2) + "\n", encoding="utf-8")
    md_path = knobs_path.parent / "zKnobs.md"
    if md_path.exists():
        md_path.write_text(md_from_knobs(envelope, knobs_path.parent.name), encoding="utf-8")


# The global knobs (2026-10-29, operator: "global knobs > [zKnobs > knobs]"): every
# song starts from songs/global_knobs.json; its own knobs.json overrides it.
GLOBAL_KNOBS_PATH = Path(__file__).resolve().parent.parent / "global_knobs.json"


def load_global_knobs():
    """songs/global_knobs.json (an empty dict when it can't be read)."""
    try:
        return json.loads(GLOBAL_KNOBS_PATH.read_text(encoding="utf-8"))
    except (OSError, ValueError):
        return {}


def global_runner_value(name):
    """A song runner's shared wait from the global knobs' "runner" block."""
    return load_global_knobs()["runner"][name]


def apply_knob_layers(knobs, global_knobs=None):
    """Each move step gets every value it doesn't name itself: the song's own
    "moves" first, then the global knobs' - a step's value wins, then the song,
    then the global. Returns knobs (changed in place)."""
    g = (global_knobs if global_knobs is not None else load_global_knobs()).get("moves", {})
    song = knobs.get("moves", {}) if isinstance(knobs.get("moves"), dict) else {}
    for step in knobs.get("steps", []):
        if not isinstance(step, dict):
            continue
        note = step.get("note")
        for layer in (song.get(note, {}), g.get(note, {})):
            for key, value in layer.items():
                step.setdefault(key, value)
    return knobs


def load_knobs_envelope(song_dir):
    """Returns (resolved_path, parsed_dict) via whole-file json.load of
    whatever resolve_active_knobs_path() picked. No per-schema handling —
    every schema family shares this same outer envelope."""
    resolved_path = resolve_active_knobs_path(song_dir)
    parsed = json.loads(resolved_path.read_text(encoding="utf-8"))
    return resolved_path, parsed

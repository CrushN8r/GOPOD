"""Shared knobs file-resolution + envelope-load — Ladder 2, Phase 1."""
import json
import shutil
from pathlib import Path

from knobs_md_001 import knobs_from_md, md_from_knobs


def resolve_active_knobs_path(song_dir):
    """."""
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
    """a tool's knobs written back - knobs.json, and the song's zKnobs.md when it has one, so the one source holds the change too (phcal apply, tempo-set, robot pick)."""
    knobs_path = Path(knobs_path)
    knobs_path.write_text(json.dumps(envelope, indent=2) + "\n", encoding="utf-8")
    md_path = knobs_path.parent / "zKnobs.md"
    if md_path.exists():
        md_path.write_text(md_from_knobs(envelope, knobs_path.parent.name), encoding="utf-8")


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

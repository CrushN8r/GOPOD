"""Song dialogue normalizer - ONE shared reader returning any song's spoken dialogue in a single standard in-memory shape, for TOOLING to consume (the score printer, any future notation/player/editor)."""

import importlib.util
import re
from pathlib import Path

from knobs_envelope_001 import load_knobs_envelope

CONTROL_SONG_RUNNER_PATH = Path(__file__).resolve().parent / "run_robot_control_song_001.py"

GENERATION_MODE_AUTHORED = "authored"
GENERATION_MODE_HINT_GUIDED_LIVE = "hint-guided-live"
GENERATION_MODE_NOT_STORED = "not-stored"

_BEAT_HEADING = re.compile(r"^##\s+BEAT\s+(\S+)\s*$")
_LINE_HEADING = re.compile(r"^##\s+LINE\s+(\d+)\s*$")


def _load_control_mod():
    spec = importlib.util.spec_from_file_location("run_robot_control_song_001", str(CONTROL_SONG_RUNNER_PATH))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def _row(step_id, speaker, section, kind, text, generation_mode, hint=None):
    return {
        "step_id": step_id,
        "speaker": speaker,
        "section": section,
        "kind": kind,
        "text": text,
        "generation_mode": generation_mode,
        "hint": hint,
    }


def _parse_beat_story_md(text):
    """Category C's own story.md shape - `## BEAT <id>` / `> TEXT:` or
    `> HINT:`, confirmed this session against 02a_brobots_interview_vamp/
    story.md (both markers observed live, e.g. BEAT m2_c/m3_c carry HINT
    only, every other beat carries TEXT only). Mirrors
    parse_control_story_md()'s own shape exactly, different heading/marker
    set - that function's regex cannot match `## BEAT`, confirmed by
    direct read of its own _STEP_HEADING pattern."""
    lines = text.splitlines()
    starts = [i for i, line in enumerate(lines) if _BEAT_HEADING.match(line.strip())]
    beats = {}
    for position, start in enumerate(starts):
        step_id = _BEAT_HEADING.match(lines[start].strip()).group(1)
        end = starts[position + 1] if position + 1 < len(starts) else len(lines)
        block = lines[start + 1:end]
        text_val, hint_val = "", ""
        for raw in block:
            stripped = raw.strip()
            if stripped.upper().startswith("> TEXT:"):
                text_val = stripped[len("> TEXT:"):].strip()
            elif stripped.upper().startswith("> HINT:"):
                hint_val = stripped[len("> HINT:"):].strip()
        beats[step_id] = {"text": text_val, "hint": hint_val}
    return beats


def _parse_line_story_md(text):
    """Category D's own story.md shape - `## LINE <n>` heading (a bare
    number, story.md's own convention - `zKnobs.json`'s real step_id is
    `line_<n>`, mapped here), followed by a `Visible line (Brobot 2):`
    label line then the actual seed text on the next `> ...` line.
    Confirmed this session against 02b_brobots_interview_run/story.md (all
    7 LINE headings carry a Visible line block)."""
    lines = text.splitlines()
    starts = [i for i, line in enumerate(lines) if _LINE_HEADING.match(line.strip())]
    seeds = {}
    for position, start in enumerate(starts):
        n = _LINE_HEADING.match(lines[start].strip()).group(1)
        step_id = f"line_{n}"
        end = starts[position + 1] if position + 1 < len(starts) else len(lines)
        block = lines[start + 1:end]
        seed_text = ""
        expect_seed_next = False
        for raw in block:
            stripped = raw.strip()
            if expect_seed_next and stripped.startswith(">"):
                seed_text = stripped[1:].strip()
                expect_seed_next = False
                continue
            if stripped.lower().startswith("visible line"):
                expect_seed_next = True
        seeds[step_id] = seed_text
    return seeds


def _normalize_a_or_b(knobs, story):
    steps = knobs.get("steps", [])
    any_real_text = any(story.get(s.get("step_id", ""), {}).get("text", "") for s in steps)
    rows = []
    for s in steps:
        step_id = s.get("step_id", "")
        speaker = s.get("speaker") or "brobot_1"
        section = s.get("section")
        kind = s.get("note", "")
        if any_real_text:
            text = story.get(step_id, {"text": ""}).get("text", "")
            mode = GENERATION_MODE_AUTHORED
        else:
            text = ""
            mode = GENERATION_MODE_NOT_STORED
        rows.append(_row(step_id, speaker, section, kind, text, mode))
    return ("A" if any_real_text else "B"), rows


def _normalize_c(knobs, beats):
    steps = knobs.get("steps", [])
    rows = []
    for s in steps:
        step_id = s.get("step_id", "")
        speaker = s.get("speaker") or "brobot_1"
        kind = s.get("movement")
        step_mode = s.get("mode", "canned")
        content = beats.get(step_id, {"text": "", "hint": ""})
        if step_mode == "llm_coloured":
            mode = GENERATION_MODE_HINT_GUIDED_LIVE
            rows.append(_row(step_id, speaker, None, kind, "", mode, hint=content.get("hint") or None))
        else:
            mode = GENERATION_MODE_AUTHORED
            rows.append(_row(step_id, speaker, None, kind, content.get("text", ""), mode))
    return "C", rows


def _normalize_d(knobs, seeds):
    steps = knobs.get("steps", [])
    rows = []
    for s in steps:
        step_id = s.get("step_id", "")
        kind = s.get("exchange_type")
        seed_text = seeds.get(step_id, "")
        rows.append(_row(step_id, "brobot_2", None, kind, seed_text, GENERATION_MODE_AUTHORED))
        rows.append(_row(f"{step_id}_reply", "brobot_1", None, kind, "", GENERATION_MODE_NOT_STORED))
    return "D", rows


def normalize_song_dialogue(song_dir):
    """The one public entry point. Returns:
      {"song_id": str|None, "category": "A"|"B"|"C"|"D"|"E",
       "supported": bool, "reason": str|None, "rows": [dict, ...]}
    `rows` is empty and `supported` is False for Category E (or any song
    whose knobs.json shape isn't recognized) - never a crash, never a
    guessed/invented row."""
    song_dir = Path(song_dir)
    try:
        _knobs_path, knobs = load_knobs_envelope(song_dir)
    except Exception as exc:  # noqa: BLE001 - a missing/unparseable knobs.json is a real E case, not a bug to hide
        return {
            "song_id": None, "category": "E", "supported": False,
            "reason": f"no readable knobs.json/zKnobs.json at {song_dir} ({exc})",
            "rows": [],
        }

    song_id = knobs.get("song_id")
    steps = knobs.get("steps", [])
    if not song_id or not steps:
        return {
            "song_id": song_id, "category": "E", "supported": False,
            "reason": "no song_id/steps in knobs.json - not a note/movement/line-based song schema",
            "rows": [],
        }

    first = steps[0]
    story_path = song_dir / "story.md"
    story_text = story_path.read_text(encoding="utf-8") if story_path.exists() else ""

    if "line_type" in first:
        seeds = _parse_line_story_md(story_text)
        category, rows = _normalize_d(knobs, seeds)
    elif "movement" in first:
        beats = _parse_beat_story_md(story_text)
        category, rows = _normalize_c(knobs, beats)
    elif "note" in first:
        control_mod = _load_control_mod()
        story = control_mod.parse_control_story_md(story_text)
        category, rows = _normalize_a_or_b(knobs, story)
    else:
        return {
            "song_id": song_id, "category": "E", "supported": False,
            "reason": f"step shape not recognized (keys: {sorted(first.keys())})",
            "rows": [],
        }

    return {"song_id": song_id, "category": category, "supported": True, "reason": None, "rows": rows}

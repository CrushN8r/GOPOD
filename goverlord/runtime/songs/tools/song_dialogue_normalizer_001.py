"""Song dialogue normalizer - ONE shared reader returning any song's spoken
dialogue in a single standard in-memory shape, for TOOLING to consume (the
score printer, any future notation/player/editor). Never written back to
disk, never read by any runner (run_golden_song_001.py,
run_robot_control_song_001.py, run_section1_full_live_001.py are all
untouched by this module and its one caller) - see
DIALOGUE_STANDARD_FORMAT_SURVEY_001.md §5 Option 3, which this formalizes
and extends. Extracted out of print_song_score_001.py's own inline
story.md/knobs.json reading, which is now this module's one real caller
(NOTATION_TOOL_PHA0B_SLOT_BUILD_001.md's own tool, rewired to consume this
instead of normalizing inline).

FIVE real storage patterns exist on disk today (same survey, §2-3) -
detection below is mechanical (which fields a song's own knobs.json steps
actually carry), never a hardcoded song-name lookup, so a new song using
one of these same shapes is recognized automatically:

  A - note/TEXT field-dump (awaken, bingo-test, nap): knobs.json step has
      `note`; real spoken text lives in story.md's own `## STEP <id>` /
      `> TEXT:` block (run_robot_control_song_001.py's own
      parse_control_story_md(), reused verbatim here, not re-derived).
  B - same shell as A, genuinely no scripted dialogue anywhere on disk
      (itsyou-single, itsyou-multi) - a live-capture song, not a scripted
      one. Detected by an honest signal, not a song-name list: if NONE of
      a note/TEXT-shaped song's own steps have any real story.md text,
      the whole song is B, not "A with some blank steps" - every row's
      own generation_mode is reported not-stored, text blank, rather than
      the routine "authored, deliberately blank" A itself uses for a
      genuine pause/wake_both/exit step.
  C - movement/mode (interview-vamp): knobs.json step has `movement`+
      `mode` ("canned"/"llm_coloured"); story.md's own `## BEAT <id>` /
      `> TEXT:` (canned) or `> HINT:` (llm_coloured) block - a DIFFERENT
      heading than A/B's `## STEP`, so parse_control_story_md() cannot be
      reused here (its own _STEP_HEADING regex is literally `^##\s+STEP`)
      - confirmed by direct read this session, not assumed.
  D - line_type/exchange_type (interview-run): knobs.json step is
      generation METADATA (line_type/exchange_type/cycle_weights/etc), no
      TEXT field at all; story.md holds only Brobot 2's own seed
      ("Visible line (Brobot 2):") per `## LINE <n>` - a THIRD, different
      heading shape again. Brobot 1's own reply is never stored anywhere -
      generated live, every run. Modeled as TWO rows per line (a real
      step_id seed row, generation_mode=authored, plus a synthetic
      `<step_id>_reply` row, generation_mode=not-stored, text always
      blank) - per the survey's own §4 finding and this build's own
      instruction: do not invent text that isn't on disk.
  E - no song schema at all (bingo-game): a compiled Go binary, not a
      JSON-schema song. Reported as unsupported, never a crash.
"""

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
    `> HINT:`, confirmed this session against 01_brobots_interview_vamp/
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
    Confirmed this session against 02_brobots_interview_run/story.md (all
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

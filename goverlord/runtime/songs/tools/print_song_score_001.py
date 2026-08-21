"""Deterministic, field-complete score printout for note-based songs - every
song on `run_golden_song_001.py`'s own SONG_REGISTRY (brobots_bingo,
brobots_cross_persona, brobots_baby_robots_sleep, gopod_is_that_you,
brobots_awaken_golden), confirmed working 2026-08-10
(SHEET_MUSIC_GOLDEN_PROCESS_SURVEY_001.md). NOT brobots_interview_section_01 -
a structurally different schema (no note/story.md-TEXT shape at all); running
this against it produces misleading blanks, not an error - stay off it.
Reads knobs.json + story.md directly, never a run's own .log/.json - so
nothing runtime (HTTP results, timestamps, elapsed time, the animation-wait
advisory) can leak in, and pause_seconds is always the true authored value,
never a cockpit reporter-gap override.

Built per the operator's go on SHEET_MUSIC_EDITOR_ROUND_TRIP_FEASIBILITY_001.md
- the printed run log is blind to several authored fields (emotion_beat's own
spoken text, hold_seconds/cycles/speed for arm_cue/nod/animation,
settle_seconds). This script's job is to make every one of those fields
RECOVERABLE by construction. It prints only - it never writes back to
knobs.json or story.md, and it is not the parse-back half of that round trip
(no such reader exists yet, and building one is out of scope here).

Reuses run_robot_control_song_001.py's own parse_control_story_md() for
STEP-heading parsing - the same proven TEXT:/FAIL: parser
run_songs_runner_001.py already relies on - rather than re-deriving a
second regex against the same file format.

USAGE
  python3 print_song_score_001.py <song_dir>

  python3 print_song_score_001.py ../songs/101_brobots_bingo_test
  python3 print_song_score_001.py ../songs/102_brobots_cross_persona

Prints to stdout. Redirect to a file to save/print on paper:
  python3 print_song_score_001.py ../songs/101_brobots_bingo_test > bingo_score.txt
"""

import importlib.util
import sys
from pathlib import Path

from knobs_envelope_001 import load_knobs_envelope

CONTROL_SONG_RUNNER_PATH = Path(__file__).resolve().parent / "run_robot_control_song_001.py"
GOLDEN_SONG_RUNNER_PATH = Path(__file__).resolve().parent / "run_golden_song_001.py"

# Display-only lookup for the printed rehearsal-header line - maps a song's
# own knobs.json song_id to the keyword pha0b's own case statement
# (~/.gopod_alias_lib/brobots.sh) already uses for it. Not every song using
# this note-based schema is pha0b-wired yet - printing a hardcoded
# "pha0b bingo ..." line on an unwired song's own printout would be actively
# wrong guidance (that command always targets brobots_bingo's own data,
# never the song actually being printed), so an unmapped song_id falls back
# to an honest "not wired yet" line instead of a guess. brobots_cross_persona
# (formerly gopod_is_that_you) added 2026-07-31 once its own pha0b arm
# ("mixup") was wired in. brobots_awaken's key updated 2026-08-10 from the
# stale pre-golden-migration song_id "brobots_awaken" to the current
# "brobots_awaken_golden" (SHEET_MUSIC_GOLDEN_PROCESS_SURVEY_001.md finding
# #3 - the old key matched nothing, so awaken's own printout falsely claimed
# pha0b wasn't wired for it). brobots_baby_robots_sleep added same pass
# (pha0b keyword "nap", confirmed in brobots.sh).
_PHA0B_SONG_KEYWORDS = {
    "brobots_bingo": "bingo",
    "robot_control_song_001": "control",
    "brobots_awaken_golden": "bait",
    "brobots_vamp_gate": "vamp",
    "brobots_cross_persona": "mixup",
    "brobots_baby_robots_sleep": "nap",
}

# Mirrors the exact defaults live in run_songs_runner_001.py's own step
# loop (_run_bingo_capture_song_steps, read 2026-07-28) - this printer reads
# song data only, it does not import/execute the runner, so these are
# hand-matched display values, not introspected at runtime. If the runner's
# own defaults ever change, these need a manual re-sync - flagged here so
# that coupling is visible, not hidden.
_DEFAULTS = {
    "buffer_after": 0,
    "settle_seconds": 1.5,
    "cycles": 1,
    "speed": 2,
    "pre_animation_pause_seconds": 2.0,
    "gap_label": None,  # falls back to the step's own step_id, not a fixed value
    "hold_seconds": {
        "emotion_beat": 5.0,
        "arm_cue": 1.2,
        "nod": 0.35,
        "animation": 2.5,
    },
}

# Which authored fields apply to each note type, beyond the universal set
# (step_id/note/speaker/TEXT/FAIL/buffer_after, shown for every step
# regardless of note type since load_bingo_capture_song() merges text/fail
# onto every step unconditionally).
_NOTE_FIELDS = {
    "wake_both": ["settle_seconds"],
    "brobots_ready_together": [],  # TEXT itself is the field here (phrase override)
    "arm_cue": ["cycles", "hold_seconds", "speed"],
    "nod": ["hold_seconds", "speed"],
    "emotion_beat": ["animation_token", "pre_animation_pause_seconds", "hold_seconds"],
    "animation": ["animation_token", "hold_seconds"],
    "pause": ["pause_seconds", "gap_label"],
    "say_turn": [],
    "rattle": [],
    "exit": [],
}


def _load_control_mod():
    spec = importlib.util.spec_from_file_location("run_robot_control_song_001", str(CONTROL_SONG_RUNNER_PATH))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def _load_golden_mod():
    spec = importlib.util.spec_from_file_location("run_golden_song_001", str(GOLDEN_SONG_RUNNER_PATH))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def _field_line(label, raw_step, key, note=None):
    """One FIELD: value line. Distinguishes a real authored value from an
    authored-blank field from a field simply absent (falling back to the
    runner's own default) - the three states the operator needs told apart
    to mark a sheet up safely."""
    if key not in raw_step:
        if key == "hold_seconds" and note in _DEFAULTS["hold_seconds"]:
            default = _DEFAULTS["hold_seconds"][note]
            return f"  {label}:{' ' * max(1, 18 - len(label))}— default: {default}"
        if key == "pause_seconds":
            return f"  {label}:{' ' * max(1, 18 - len(label))}— NOT SET (step will fault: ok=False, no sleep)"
        if key == "gap_label":
            return f"  {label}:{' ' * max(1, 18 - len(label))}— default: <this step's own step_id>"
        if key == "animation_token":
            return f"  {label}:{' ' * max(1, 18 - len(label))}— not authored (no runner default; produces a literal 'None' token)"
        default = _DEFAULTS.get(key)
        return f"  {label}:{' ' * max(1, 18 - len(label))}— default: {default}"
    value = raw_step[key]
    return f"  {label}:{' ' * max(1, 18 - len(label))}{value}"


def _text_line(label, value):
    if value == "":
        return f"  {label}:{' ' * max(1, 18 - len(label))}(blank)"
    return f"  {label}: {value}"


def print_song_score(song_dir):
    song_dir = Path(song_dir)
    knobs_path, knobs = load_knobs_envelope(song_dir)
    control_mod = _load_control_mod()
    story = control_mod.parse_control_story_md((song_dir / "story.md").read_text(encoding="utf-8"))

    song_id = knobs.get("song_id", "")
    # note_aliases (e.g. brobots_awaken_golden's head_nod -> nod) read straight
    # from the golden runner's own SONG_REGISTRY - the single source of truth
    # the actual runner dispatches through - rather than a second, hand-kept
    # guess. {} for every song not in that registry (bingo-family's own
    # entries are also {} - no song needs this today except awaken). Fixes
    # SHEET_MUSIC_GOLDEN_PROCESS_SURVEY_001.md finding #1: without this, a
    # note only known by its alias (head_nod) silently lost every field
    # (HOLD_SECONDS/SPEED) this printer exists to make visible.
    golden_mod = _load_golden_mod()
    note_aliases = golden_mod.SONG_REGISTRY.get(song_id, {}).get("note_aliases", {})
    lines = []
    lines.append("=" * 71)
    lines.append(f"{song_id.upper()} — SCORE PRINTOUT (source: knobs.json + story.md)")
    lines.append('Every field below is authored on disk. "(blank)" = explicitly authored')
    lines.append('empty. "— default: X" = not authored; runner falls back to X. FAIL is')
    lines.append("parsed by this song's story.md format but never read by this runner (dead")
    lines.append("field today - not printed below).")
    lines.append("")
    pha0b_keyword = _PHA0B_SONG_KEYWORDS.get(song_id)
    if pha0b_keyword:
        lines.append(f"To rehearse a marked step: pha0b {pha0b_keyword} <step_id> <step_id>")
    else:
        lines.append(f"To rehearse a marked step: pha0b isn't wired for {song_id!r} yet.")
    lines.append("=" * 71)

    current_section = object()  # sentinel, never equals a real section value
    for raw_step in knobs.get("steps", []):
        step_id = raw_step.get("step_id", "")
        note = raw_step.get("note", "")
        speaker = raw_step.get("speaker") or "brobot_1"
        section = raw_step.get("section")

        if section != current_section:
            current_section = section
            lines.append("")
            label = section if section is not None else "(none authored)"
            lines.append(f"SECTION: {label}")
            lines.append("-" * 71)

        content = story.get(step_id, {"text": "", "fail": ""})
        # dispatch_note is what the runner actually treats this step as after
        # its own note_aliases resolution - used only for field-set/default
        # lookups below. The header always shows the authored `note` itself
        # (this printer's whole promise is "authored on disk"), with a
        # "(dispatches as X)" suffix when an alias applies so the extra
        # fields below aren't a mystery.
        dispatch_note = note_aliases.get(note, note)
        alias_suffix = f"  (dispatches as {dispatch_note})" if dispatch_note != note else ""

        lines.append("")
        lines.append(f"[ {step_id} ]  note={note}  speaker={speaker}{alias_suffix}")

        text_label = "TEXT (phrase override)" if note == "brobots_ready_together" else "TEXT"
        lines.append(_text_line(text_label, content["text"]))

        for field in _NOTE_FIELDS.get(dispatch_note, []):
            field_label = field.upper()
            lines.append(_field_line(field_label, raw_step, field, note=dispatch_note))

        lines.append(_field_line("BUFFER_AFTER", raw_step, "buffer_after"))
        lines.append("  NOTE:")

    return "\n".join(lines)


def main():
    if len(sys.argv) != 2:
        print("usage: python3 print_song_score_001.py <song_dir>", file=sys.stderr)
        return 2
    print(print_song_score(sys.argv[1]))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

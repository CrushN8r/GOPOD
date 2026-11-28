"""."""

import importlib.util
import sys
from pathlib import Path

from knobs_envelope_001 import load_knobs_envelope
from song_dialogue_normalizer_001 import normalize_song_dialogue

GOLDEN_SONG_RUNNER_PATH = Path(__file__).resolve().parent / "run_golden_song_001.py"

_PHA0B_SONG_KEYWORDS = {
    "brobots_bingo": "bingo",
    "robot_control_song_001": "control",
    "brobots_awaken_golden": "bait",
    "brobots_vamp_gate": "vamp",
    "brobots_cross_persona": "mixup",
    "brobots_baby_robots_sleep": "nap",
}

_DEFAULTS = {
    "buffer_after": 0,
    "settle_seconds": 1.5,
    "cycles": 1,
    "speed": 2,
    "gap_label": None,  # falls back to the step's own step_id, not a fixed value
    "hold_seconds": {
        "emotion_beat": 5.0,
        "arm_cue": 1.2,
        "nod": 0.35,
        "animation": 2.5,
    },
}

_NOTE_FIELDS = {
    "wake_both": ["settle_seconds"],
    "brobots_ready_together": [],  # TEXT itself is the field here (phrase override)
    "arm_cue": ["cycles", "hold_seconds", "speed"],
    "nod": ["hold_seconds", "speed"],
    "emotion_beat": ["animation_token", "hold_seconds"],
    "animation": ["animation_token", "hold_seconds"],
    "pause": ["pause_seconds", "gap_label"],
    "say_turn": [],
    "rattle": [],
    "exit": [],
}


def _load_golden_mod():
    spec = importlib.util.spec_from_file_location("run_golden_song_001", str(GOLDEN_SONG_RUNNER_PATH))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def _field_line(label, raw_step, key, note=None):
    """One FIELD: value line."""
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
    normalized = normalize_song_dialogue(song_dir)
    text_by_step_id = {row["step_id"]: row["text"] for row in normalized["rows"]}

    song_id = knobs.get("song_id", "")
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

        dispatch_note = note_aliases.get(note, note)
        alias_suffix = f"  (dispatches as {dispatch_note})" if dispatch_note != note else ""

        lines.append("")
        lines.append(f"[ {step_id} ]  note={note}  speaker={speaker}{alias_suffix}")

        text_label = "TEXT (phrase override)" if note == "brobots_ready_together" else "TEXT"
        lines.append(_text_line(text_label, text_by_step_id.get(step_id, "")))

        for field in _NOTE_FIELDS.get(dispatch_note, []):
            field_label = field.upper()
            lines.append(_field_line(field_label, raw_step, field, note=dispatch_note))

        lines.append(_field_line("BUFFER_AFTER", raw_step, "buffer_after"))
        lines.append("  NOTE:")

    return "\n".join(lines)


def print_song_score_narrative(song_dir):
    """Additive sibling to print_song_score() above - a PURE consumer of
    song_dialogue_normalizer_001.py's own normalize_song_dialogue(), no
    direct knobs.json/story.md reading of its own at all (field-dump mode
    still needs the raw knobs.json steps for technical fields the
    normalizer doesn't carry - this mode needs nothing else). Same
    section-grouping shape as before, different presentation: flowing
    "SPEAKER: TEXT" dialogue instead of a bracketed per-field dump, no
    technical fields shown at all. A step with no TEXT authored (a
    cue/pause/action, or a not-stored row) prints as a bracketed stage
    note (`[kind]`) rather than being silently skipped, so the read still
    flows step to step in real authored order - never reordered, never
    summarized."""
    normalized = normalize_song_dialogue(song_dir)
    song_id = normalized["song_id"] or ""
    lines = []
    lines.append("=" * 71)
    lines.append(f"{song_id.upper()} — SCRIPT READ (source: knobs.json + story.md)")
    lines.append("A narrative read of this song's authored dialogue, grouped by section.")
    lines.append("Stage notes in [brackets] mark a non-speaking step (a cue, a pause, an")
    lines.append("action) so the pacing still reads even where nothing is said.")
    lines.append("=" * 71)

    current_section = object()  # sentinel, never equals a real section value
    for row in normalized["rows"]:
        section = row["section"]
        if section != current_section:
            current_section = section
            lines.append("")
            label = section if section is not None else "(no section authored)"
            lines.append(f"SECTION: {label}")
            lines.append("-" * 71)

        if row["text"]:
            speaker_label = row["speaker"].replace("_", " ").title()
            lines.append(f"{speaker_label}: {row['text']}")
        else:
            lines.append(f"[{row['kind']}]")

    return "\n".join(lines)


def main():
    if len(sys.argv) not in (2, 3):
        print("usage: python3 print_song_score_001.py <song_dir> [--narrative]", file=sys.stderr)
        return 2
    narrative = False
    if len(sys.argv) == 3:
        if sys.argv[2] != "--narrative":
            print(
                f"usage: python3 print_song_score_001.py <song_dir> [--narrative] "
                f"(unknown flag {sys.argv[2]!r})",
                file=sys.stderr,
            )
            return 2
        narrative = True
    if narrative:
        print(print_song_score_narrative(sys.argv[1]))
    else:
        print(print_song_score(sys.argv[1]))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

---
name: print-score
description: Use when the operator asks to see, print, or dump a GOPOD song's score (story.md + knobs.json) for reference. Read-only — prints every beat organized by movement with step_id, speaker, mode, and verbatim TEXT/HINT line; calls out vamp/filler loops separately if present. Never edits, runs, or assesses the content — just reads and prints what's on disk. Writes to gopod_notes/ only if the score is too long to show inline.
---

# Print score

A read-only reference dump of a GOPOD "song" (a scored sequence of beats/steps driving the
robots and hosts). Used for the operator to review content without opening two files and
cross-referencing them by hand.

## What to read

For the named song directory (e.g. `goverlord/runtime/songs/<song_name>/`):
- `story.md` — a **pre-show/movement song** (e.g. `brobots_preshow`) uses
  `## BEAT <id>` headings with `> TEXT:` (verbatim spoken line) or `> HINT:` (situational
  prompt for an LLM-coloured line, never spoken as written). An **interview song** (e.g.
  `brobots_interview_section_01`) uses a different shape - line-numbered exchanges with a
  `visible_line` (the interviewer's scripted question/statement) and `value_points` (what
  the interviewee should reveal, in locked order) instead of bare TEXT/HINT. Read the
  actual heading/field shape present, don't assume one song type's shape for the other.
- `knobs.json` — a pre-show song's steps carry a single `mode` field
  (`canned` / `llm_coloured`) plus `speaker`. An interview song's lines carry **per-role**
  `brobot_1_mode`/`brobot_2_mode` fields (`canned` / `llm`, not `llm_coloured`) plus
  `brobot_1_canned_source`/`brobot_2_canned_source` when canned, and per-line
  `brobot_1_movement`/`brobot_2_movement` when a scored gesture fires. Print whichever
  fields the actual file has, not a merged/assumed schema.

**Not a song**: the interview *runtime scaffold* (
  `goverlord/runtime/songs/02_brobots_interview_run/zmisc/brobots_wirepod_interview_section_card_template_1_001.md`
  and its standalone JSON twin, `brobots_interview_runtime_scaffold_001.json`) is a
  different kind of file entirely - shared prompt rules, the pronunciation registry, the
  channels registry, cleanup rules - not a story.md/knobs.json pair, and not a beat/line
  sequence to print this way. If asked to "print the scaffold" or "print the pronunciation
  registry," that's a different, smaller, targeted read - don't run this skill's
  beat-by-beat format against it.

## How to print

Organize by movement/section grouping if the knobs.json has one (a `movement` field, a
`section` grouping, etc.) — one table or block per movement, in on-disk order, not
alphabetical or reordered.

For each step show: step_id, speaker, mode, and the actual TEXT or HINT line verbatim —
never paraphrase or summarize the line itself.

If the song has a vamp/filler/loop section (repeating content that plays when the main
content isn't ready yet), call it out in its own section: state how many beats it has and
confirm what it's for (e.g. "loops until X lands"), using only what's actually stated in the
file — don't invent a purpose not shown in the story.md commentary.

## Scope

- Read-only. No edits, no run, no git commands.
- Do not assess, critique, or propose changes — just print what's on disk.
- Print directly in the response by default. Only write a companion file to
  `~/crushn8r_git/gopod_notes/` (following the project's `ALL_CAPS_NAME_001.md` convention)
  if the score is too long to show inline, or if the operator asks for it to be saved.

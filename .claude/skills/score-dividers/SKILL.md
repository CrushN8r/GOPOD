---
name: score-dividers
description: Use when a GOPOD song runner's printed run log/score needs labeled "=====" note dividers added, extended to a new note type, or ported to a different song. The format - a plain "=====" sandwich around section headers, a short human-readable tag ("===== wake", "===== call b-1", "===== happy beat") before each per-note block - is proven on run_golden_song_001.py (`_derive_note_tag`/`_derive_say_turn_tag`). Tags come from a small helper (note-type map, emotion_beat token, say_turn text heuristic, a handful of named overrides), never a per-step lookup table. Display-only: never touches step order, timing, gaps, dispatch, or the JSON dump.
---

# Score dividers

A read-print convention that makes a song runner's terminal/log output read as a labeled
score instead of a flat event stream. Built once for `run_songs_runner_001.py`
(BINGO SHEET MUSIC pass, commit `2404abc`, 2026-07-19) — this skill is that pattern made
reusable for any other song's own runner, present or future, one song at a time.

## The format (exact, don't improvise)

- Section-header sandwich, on every section change (wraps whatever timestamped
  section-header line that song already prints):
  ```
  <blank line>
  =====
  <the existing section-header line, unchanged>
  =====
  <blank line>
  ```
- Per-note divider, printed immediately before that note's own first log line, every
  step, no exceptions: `===== <tag>` — no timestamp prefix; both divider types are bare
  structural lines, not events.
- The section sandwich carries no tag. Only per-note dividers do.

## Deriving the tag — a helper, never a per-step lookup table

Port the `_derive_note_tag()` / `_derive_say_turn_tag()` shape from
`run_golden_song_001.py` (the reference implementation — the pattern's original
2026-07-19 home, `run_songs_runner_001.py`, has since been retired in favor of this
one golden runner) rather than inventing a new shape per song:

1. **Note-type map** for structural/motion notes whose meaning doesn't depend on
   content — bingo's: `wake_both`→"wake", `arm_cue`→"arm cue", `nod`→"nod",
   `rattle`→"rattle", `brobots_ready_together`→"brobots together", `exit`→"exit".
   Extend this dict per song for whatever new note types that song introduces — this is
   the one place growth is expected, per new note *type*, not per step.
2. **`pause`** → derive from the `step_id` prefix (bingo's: `reporter_gap_N`→"reporter
   gap", `transition_gap_N`→"transition gap"), not the note type alone, since a song can
   have more than one flavor of pause.
3. **`emotion_beat`** → `f"{animation_token} beat"` verbatim, token casing preserved
   (e.g. "veryHappy beat") — the token already carries the right amount of information,
   no extraction needed.
4. **`say_turn`** → derived from the spoken text: strip that song's own repeated
   set-dressing filler phrase(s) (bingo's was "big shiny bingo ball(s)" — every song
   will have its own, if any), take the first sentence, strip a small generic stopword
   list, keep up to 2–3 remaining content words in order. A `step_id` containing "call"
   plus a bare `[letter]-[digits]` token in the text (e.g. "B-1") becomes `"call b-1"`;
   that same bare token alone becomes just `"b-1"` — a general numbered-call-out shape,
   not bingo-specific despite the example.
5. **Named overrides, sparingly** — a small `_TAG_OVERRIDES = {step_id: "tag"}` dict for
   the handful of lines the derivation doesn't land on cleanly, filled in from the
   operator's own correction after seeing a dry-fire's output. Two or three named
   entries is normal; a 40-entry override dict means the derivation itself needs
   rethinking, not more overrides.

## Segmentation is optional — sections and tags are two independent halves

The per-note `===== <tag>` divider fires on **every** step regardless of whether that
step carries a `section` field — it has no dependency on segmentation. The
section-header sandwich is the only half that needs a `section` to wrap; a song with no
`section` key on any step will show tagged note dividers with no section banners at
all, and that is a valid, working result, not a broken or incomplete one.

Surveyed 2026-07-19: **only `brobots_bingo` has `section` on its steps** (42
of 42). Every other current song has zero: `brobots_bait_001` (1 step), `brobots_bait_002`
(16), `brobots_interview_section_01` (7), `brobots_preshow` (29), `brobots_awaken`
(3), `robot_control_song_001` (9). Porting this skill to any of those today, as-is, gets
per-note tags only.

If section banners are also wanted for a given song, that's a separate, content-level
decision made *before* porting the divider code: add a `section` string per step to that
song's own `knobs.json` (pure organizational metadata, same non-behavioral addition
bingo's own docstring made — never alters order, timing, or dispatch). What the sections
should be named/grouped as is that song's own content call — ask the operator rather than
inventing a segmentation scheme for someone else's dialogue.

## Applying to a song (new or existing)

1. Read that song's own runner file's step loop — find where its section-header print
   already lives (the `if section and section != current_section:` block, if that song
   has one at all) — plus its `knobs.json`/`story.md` for the full step list. Same
   survey-first discipline as any change to a song runner; report the survey before
   editing, including whether the song already carries `section` per step or not (see
   above).
2. Port the building blocks (`_NOTE_TYPE_TAGS`, filler-phrase/stopword sets, the
   numbered-call-out regex if that song has call-outs, `_TAG_OVERRIDES`) — adapt the
   note-type map and filler phrases to that song's own note shapes and dialogue, keep
   the derivation shape identical to the reference implementation.
3. Add the section-sandwich print and the per-note `===== <tag>` print at the same two
   points in that song's own loop.
4. Dry-verify (see `dry-verify` skill): compile-check, dry-fire the full song, confirm
   every step is still reached with `stopped_early: false` and the JSON dump is
   unaffected (no `=====` lines leaked into it), then paste back enough of the saved
   `.log` for the operator to eyeball the tags.
5. Report the survey and let the operator flag any tag that doesn't land — fix via
   `_TAG_OVERRIDES`, re-verify, then hand to `survey-then-commit`. Don't commit under
   this skill's own authority.

## What this does NOT do

- Touch a song's actual content, timing, `buffer_after` values, or dispatch order —
  display-only, every time, every song.
- Assume every song wants this — a song with no `section` key or no multi-note
  structure (e.g. a single-shot utility script) may not need it; ask if unclear rather
  than adding it uninvited.
- Factor into a shared cross-song helper module automatically. Each song's runner
  currently owns its own copy of these building blocks, ported per-song rather than
  imported from one place. If duplication across many songs later makes a shared module
  worth it, that's a separate, explicit decision this skill doesn't make unilaterally.

## Scope

- Per-song, applied on request — this skill does not proactively retrofit every
  existing song runner. Every song `run_golden_song_001.py` drives (Bingo, the
  cross-persona mix-up, Baby Robots Sleep, "is that you?", Awaken's own full
  playback) already carries it, inherited from the pattern's original 2026-07-19
  build; songs on a different engine (the interview, vamp-gate, and the standalone
  control-song bench-test entry points `run_robot_control_song_001.py` still runs
  directly) are unchanged until asked.
- No live robot run under this skill's own authority — dry-fire only, same boundary as
  `dry-verify`.

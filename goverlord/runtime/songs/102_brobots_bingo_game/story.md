# Chocolate Bingo — All Ages (Brobots Bingo Game, the live game itself, not a performance song)

**Naming note, brobots wire-pod layer:** this layer's robots are Brobot 1 and Brobot 2. "Doc"
and "Pip" are GOPOD-layer persona names — a different, future layer, not this one. There is
expected overlap between the two layers, so this distinction is held deliberately, everywhere
Brobot 1/Brobot 2 are described in this layer's own files.

This folder is a `pha0b` launcher entry, not a note-sequence song. Picking it plays no scored
steps — it launches `gobingo`, the real, live, voice/touch-triggered Bingo game
(`goverlord/runtime/songs/102_brobots_bingo_game/`, a standalone Go binary, `bin/vectorx-gobingo`). See
`goverlord/runtime/songs/102_brobots_bingo_game/README.md` for its own current status.

**Fundamentally a single-robot host.** Brobot 1 runs the whole game — a real bingo host, solo.
Brobot 2 is a for-show companion, not a second host: one angry reaction at the very end.
Structurally closer to a real single-caller bingo night than a two-host bit.

---

## Two bingo songs, compared

Genuinely different pieces built from the same idea — real, not a duplicate.

| | `102_brobots_bingo_game` (this folder — "Chocolate Bingo") | `101_brobots_bingo_test` (the "bingo test" song) |
|---|---|---|
| What it is | The live game itself | A scripted, comedic capture song for the upsell video |
| Engine | Standalone Go binary (`gobingo`), not a scored song | Golden song engine (`run_golden_song_001.py`) |
| Numbers | Real Fisher-Yates draws off a real deck, different every run | A fixed, written script — the same three-round bit every time |
| Robots | Brobot 1 solo hosts; Brobot 2 reacts once, at the end | Both robots run a full choreographed bit throughout — arm cues, head nods, reaction beats, banter |
| Naming | Plain — Brobot 1 / Brobot 2 | Notated `host_`/`brat_` speaker prefixes (`host_arm_cue_0102`, `brat_say_0201`, etc.) — Brobot 1 is the host, Brobot 2 is the brat |
| Ball call | Bare number, spoken (editable — see below) | A scored flourish (rattle → "Big Shiny Bingo Ball" call → the number → a reaction beat) |
| Editable how | Recompile the Go source (see below) | Edit `knobs.json` directly, no rebuild |
| Launch | `pha0b` → pick this folder, or `gobingo` directly | `pha0b` → pick `101_brobots_bingo_test`, or `bingo-video-song`/`bingo-video-song-live` |

**What they share:** the same core shape — a deck of numbers, a rattle sound, "BROBOTS!" energy,
the same two robots, the same Chocolate Bingo branding when presented publicly. Neither touches
the other's code.

---

## Golden rule — confirmed live, 2026-08-12

**Wait for the first backpack rub to start; from there it runs cleanly on its own to the end
(deck exhaustion) or until "bingo" is called via the backpack button.** Live-confirmed twice in
a row, clean both times. This is the correct, golden shape for run 1 (continuous) — stamped
here as the reference behavior, not something to "fix" if seen again.

**Run 2 (pause for backpack rub)** follows the same golden shape, just gated per-ball instead of
only on the first one: every single draw, start to finish, needs its own fresh rub — same clean
run-to-completion pattern as run 1, just paced by hand instead of by ticker.

---

## Brobot 2's end-of-game reaction

**The chocolate connection:** this is Chocolate Bingo — there's a chocolate prize on the line.
Brobot 2 doesn't just lose a generic round; it gets mad specifically for **not winning the
chocolate**. That's the actual joke, not "sore loser" in the abstract.

**Fires once, at the very end — for both ways the game can end** (deck exhausted, or "bingo"
called early via the backpack button both write the same `bingo_end` signal). Not a per-draw
reaction anymore — Brobot 2 watches quietly through the whole game and gets mad exactly once,
right at the end, for missing out on the chocolate prize. Built 2026-08-12
(`goverlord/runtime/songs/102_brobots_bingo_game/bingo_reactor/bingo_reactor_001.py`) — see
`gopod_notes/BINGO_REACTOR_STALE_REPLAY_FIX_AND_DIVISION_QUESTION_001.md` for the per-draw
version this replaced and why.

## Brobot 2's screen display — blocked lane, not built

Brobot 2 was meant to show the drawn number on its own head screen. Confirmed blocked: the
underlying display call (`WriteColoredText`) was already removed from `gobingo`'s own source
before this sidecar became its source of truth (see the Error 915 investigation,
`gopod_notes/older_notes/VECTORX_BINGO_ERROR915_WRITECOLOREDTEXT_INVESTIGATION_001.md`). Named
here as a known, blocked gap — not a task, not attempted.

---

## Making the ball call editable

`formatBingoNumber()` in `voicecommand_lottery.go` reads from a single named variable,
`BingoCallFormat` (default `"%d"` — just the bare number). Change that one line to change what
Brobot 1 says on every draw — e.g. `"Big Shiny Bingo Ball, number %d!"` to match the scripted
bingo song's own flourish. **Requires a rebuild** (the overlay-build process documented in
`goverlord/runtime/songs/102_brobots_bingo_game/README.md`) — this is compiled Go, not a live-editable `knobs.json` the
way the golden song engine's songs are. Honest constraint, not a limitation to hide.

This is also the pattern for composing any future live-game performance the same way: a small
set of named, well-commented variables (call format, animation name, grid size) standing in for
what would be knobs.json values on a scored song — variables the shape lives around, even though
this one can't be live-edited without a rebuild.

---

## How it runs

Picking `102_brobots_bingo_game` from `pha0b`'s menu asks two questions — run 1 (continuous) or
run 2 (pause for backpack rub), then grid size (75 or 90, default 75) — then launches `gobingo`
directly. No playhead A/B slice, no dry/live prompt, no reporter-gap or phcal-apply step: none
of those apply here, gobingo has no step sequence to slice and no dry mode of its own — it
always fires the real binary against real hardware the moment it's picked (`--silent` only
suppresses speech, not action).

## `knobs.json` / `zKnobs.json`

Both are placeholders on purpose. Gobingo is a launched Go binary with its own CLI flags
(`--serial`, `--locale`, `--grid-size`, `--silent`, `--pause-for-touch`), not a note-sequence
song — the step/knobs schema every other song folder uses genuinely does not apply here.

## `notation/`

Carried over from an earlier folder-duplication pass (byte-identical to
`101_brobots_bingo_test/notation/`) — describes the *performance song*'s score, not this game. Left
in place, flagged here so a future reader isn't confused about which song it actually documents.

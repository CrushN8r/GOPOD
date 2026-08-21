# Song Scaffold

> **Starting point, not a locked spec.** This is a growing-until-it-settles reference — what a
> GOPOD song looks like when it's built the proven way, assembled from what's *actually already
> working* across the shelf, not an ideal invented from scratch. Expect this file to change as
> more songs catch up to it or a new proven pattern gets added. Check the per-song status table
> at the bottom before assuming any one song already matches everything below.

---

## 0. Naming — which layer, which names

**Brobots wire-pod layer (this layer, everything in this doc): Brobot 1 / Brobot 2.** "Doc" and
"Pip" are GOPOD-layer persona names — a different, future layer, not this one. Operator
correction, 2026-08-12: expected overlap between the two layers is exactly why this distinction
has to be held deliberately, not blurred — any wire-pod-layer song, doc, or report describing
what the robots actually do in this layer uses Brobot 1/Brobot 2, never Doc/Pip. The doctrine
closer's own signature ("— Doc Squawkadoodle") and links to the separately-titled
`DOCS_TAKE_LESSON_*`/GOPOD-layer document series are a different, established convention and
are not affected by this rule — those are named things in their own right, not this layer's
robots being described.

---

## 1. Status marker

Every song's `story.md` opens with a dated, bold status line, first paragraph:

```
**GOLDEN — <date>.** <what was confirmed, how, run citations.>
```
or
```
**WIP — <date>. <why it's WIP.>**
```
or `**SOFT LOCK**` / `**STABLE**` for shades between the two.

**Proven:** `00_brobots_awaken` (`**GOLDEN — 2026-08-10.**`, two clean back-to-back live runs
cited by log file), `101_brobots_bingo_test` (`**WIP — 2026-08-11. Golden lock broken by design.**`,
prior golden state preserved as a snapshot, re-locks only after a fresh live run confirms the
rebuilt shape). `goverlord/runtime/songs/102_brobots_bingo_game/README.md` adopted the same convention for the
non-song `gobingo` game (`**STATUS — 2026-08-10.** Stale, not broken.`).

**Not yet universal** — `01_brobots_interview_vamp`/`02_brobots_interview_run` (split from
one combined `01_brobots_interview_section_01` 2026-08-19), `102_brobots_cross_persona`,
`105_brobots_nap`, `103_gopod_is_that_you_single`/`104_gopod_is_that_you_multi` carry no dated
top-of-file status marker today.

## 2. Engine choice

A song is built on one of these, chosen by shape, not by default:

- **`run_golden_song_001.py`** — the golden note-dispatch engine. Default choice for any new
  scored, multi-note song. Carries the golden toolbox (§6 below).
- **`run_robot_control_song_001.py`** — the legacy control-song engine. No longer the target for
  new work; both songs still on it (`control`, `weather`) are archived (`zzz_archives/`),
  reachable by keyword only, not on the active shelf.
- **A dedicated standalone runner**, when the shape is genuinely different:
  - `run_section1_full_live_001.py` — the interview's own two-phase generate-then-play engine
    (full LLM generation pass, then a replay pass). Also the shared library every other runner
    imports from.
  - `run_vamp_gate_song_001.py` — the vamp filler beats' own trivial single-song runner, no
    physical robot.
  - `gobingo` (`goverlord/runtime/songs/102_brobots_bingo_game/`, Go) — not a Python note-dispatch song at all. A
    standalone live game binary. `102_brobots_bingo_game/` is a pha0b launcher entry that calls
    it directly, not a scored song.

## 3. Reporter-gap convention

The floor, per standing studio rule (2026-08-01): every song's own runtime score carries at
least a 0-second `reporter_gap_intro`/`reporter_gap_outro` pair, `pause_seconds: 0` by default
— a real slot for a later-edited-in reporter voiceover, never a live dead-air pause.

**Cohesive today:** `105_brobots_nap` (renamed from `104_brobots_baby_robots_sleep`,
2026-08-18) — has both `reporter_gap_intro` and `reporter_gap_outro`, clean.

**Not cohesive yet, confirmed by direct grep:** `101_brobots_bingo_test` has multiple `reporter_gap_*`
steps but none literally positioned as intro/outro. `102_brobots_cross_persona` has one
`reporter_gap_mid` only — no intro, no outro. `00_brobots_awaken`, `01_brobots_interview_vamp`,
and `02_brobots_interview_run` carry no `reporter_gap_*` step at all yet.

A concrete adoption plan already exists (`103`'s naming as the floor, numbered `mid_N` gaps for
multi-round songs) but needs the operator's explicit go-ahead before it touches any song's
score — this reorders live material, not a documentation-only change.

## 3a. The vamp — a detachable pre-show module, not owned by any one song

The vamp (`vamp_1..vamp_4`, the four Kokoro-voiced fallback beats Brobot 3/4 loop through
live) is a **detachable pre-show module**, conceptually separate from the interview's own
scripted content even though it exists to serve that content's live generation wait. It
plays *while a song's content generates live*, filling that wait — so it belongs to
whichever song has a live generation-wait to fill, not to any song by default. Today
that's the interview alone. **Split 2026-08-19** (`gopod_notes/INTERVIEW_VAMP_SPLIT_001.md`)
from a nested `01_brobots_interview_section_01/vamp/` loader-convenience subfolder into its
own genuinely standalone, independently-fireable top-level song folder,
`01_brobots_interview_vamp/` (video 1 of the interview's two-video pair) — the interview's
own scripted content now lives separately at `02_brobots_interview_run/` (video 2). The
vamp's conceptual detachability described above is now also structurally true, not just
loader convenience.

**Reporters are the universal layer; the vamp is not.** The two reporter voices (Brobot
3/4, "Math Aftermath BREAKING NEWS" style) travel to every song-video — see §3's
reporter-gap convention above, which every song carries a slot for. The vamp itself does
NOT travel — only the reporter voice/style does. A vamp is optional per song-video: only a
song with a live generation-wait gets one.

Confirmed per-song split:
- **`00_brobots_awaken`** — reporters yes (the reporter-gap slots in §3's table), vamp no.
  Nothing in this song generates live — it's fully scripted, no wait to fill.
- **`02_brobots_interview_run`** (formerly the `01_brobots_interview_section_01` combined
  folder's own content) — reporters yes, vamp yes (now a separate song,
  `01_brobots_interview_vamp`). Its content (the 7 exchanges) generates live as JSON before
  playback, so the vamp fills that wait when rolled via `interview-vamp` (renamed from
  `vamp-run` 2026-08-19, `GOPOLISHER_FIXES_001.md`).
  Live-fired clean 2026-08-14 (`vamp-run`, the name at the time), see
  `gopod_notes/VAMPIRED_INTERVIEW_GOLDEN_SNAPSHOT_001.md`.

**Batchable, not tied to a single sitting.** The vamp is a "roll a take" tool — several
songs can be vamped in one session (generate takes in the morning), then performed later
(afternoon). It points at whatever song needs generating, whenever asked.

**Naming**: a song-video that carries a vamp pre-show is "vampired" — e.g. "the vampired
interview" for the interview-with-vamp video. See `web/FUNNY_NAMINGS.md`.

**Resolved (was flagged in `gopod_notes/INTERVIEW_TWO_PART_VAMP_SURVEY_001.md` as a possible
naming collision): three labels, one split, not three things.** "Vamp-vs-performance" (this
section's own framing), "generate-vs-perform" (`SONG_02_BROBOTS_1_2_INTERVIEW_RUN.md`'s
`generate_phase()`/`playback_phase()` software framing), and "Part 1 vs Part 2" (the
operator's own typing shorthand) all name the same generate-then-perform seam from three
different angles — not competing models of the song. They were never in conflict; say the
seam once, in whichever of the three words fits the sentence.

## 3b. Vamp vs. reporter-gaps — two delivery mechanisms, same reporters

§3 (reporter-gaps) and §3a (the vamp) can read like two separate features. They're not —
they're **two different delivery mechanisms for the same two reporters** (Brobot 3/4).
Same actors, two jobs:

- **The vamp** — a *live* pre-show: the reporters warm up and talk in real time, filling
  the unpredictable wait *while* a song's JSON is still generating. Live-spoken,
  variable-length, only exists for songs that generate content live.
- **Reporter-gaps** — a *silent, pre-placed* slot baked into a song's own timeline
  (`reporter_gap_intro`/`reporter_gap_outro`/etc., `pause_seconds: 0` by default), left for
  a reporter voiceover to be recorded and edited in later. Fixed-length, edited-in after
  the fact, not spoken live at all. Any song can carry these — no live generation needed.

The vamp is live improv while the stage is still being set; reporter-gaps are scripted
voiceover dropped into the final edit. Both exist to get the same reporters' voices into
the video — they just solve different timing problems, and a song can have either, both,
or (per §3's table) neither yet.

Reinforces the per-song split §3a already states: `00_brobots_awaken` has reporter-gap
slots (§3) but no vamp — nothing in it generates live, so there's no unpredictable wait to
fill live; it only needs the edited-in mechanism. `02_brobots_interview_run` has the vamp
available (§3a — now a separate song, `01_brobots_interview_vamp`) because its content
generates live, and could also carry reporter-gaps of its own
(§3's table already lists it as not-yet-cohesive, same as every other not-yet-adopted
song) — the two mechanisms aren't exclusive.

## 4. The standard closer — GOPOD layer, not this layer

**Correction, 2026-08-12, operator's own word:** Doc's Take ("Boom. Done. Class is dismissed!")
+ Pip's door is a **live performance session closer signal/beat, intended for the GOPOD layer**
— not a brobots wire-pod layer song element. There is no Doc or Pip session-closer beat in this
layer. The earlier finding in this file (framed as "missing from every song's `knobs.json`") was
the wrong frame — it isn't missing, it was never supposed to live here. Not tracked as a gap in
the §7 table below; not something any wire-pod-layer song needs to build.

## 5. The two-layer video recipe

Proven once (the bait video, now folded into `00_brobots_awaken`), general — not song-specific:

1. **A scored capture song with deliberate silent gaps**, filmed live exactly as it runs. The
   camera records the show, never the code.
2. **A separately-recorded reporter voiceover track**, written and rendered independently,
   edited into those gaps by a human afterward. The runtime never plays it.

The reporter-gap convention (§3) is what makes step 2 possible — the gaps have to exist in the
score before an editor has anywhere to put the voiceover.

## 6. Golden toolbox (available to any golden-engine song)

Real, proven, code-level primitives — opt in per song via `knobs.json`, not automatic:

- **Stay-put** — one continuous behavior-control hold across the whole run instead of
  assume/release per step. On by default for golden-engine songs.
- **Tempo** (`global_tempo` in `knobs.json`) — a song-level pacing multiplier, additive on top
  of each step's own `buffer_after`. `0.0` (the current default everywhere it's set) is a
  byte-identical no-op; a song opts in by raising it.
- **Playback filter** (`GOPOD_GOLDEN_ROBOT_FILTER`) — per-run, not persisted, restricts a run to
  one robot's lines.
- **Golden-flag wake** — the proven fix for a real once-live bug (a motor/behavior command
  firing before the real async control-grant had actually landed). Baked into the control-family
  shim already.
- **Weather note** — a real per-robot weather fetch, control-family songs only today.
- **Connect once, hold it** (the newest addition, 2026-08-12) — any tool talking to a robot
  repeatedly in one run should open one connection and reuse it, never reconnect per step. Full
  rule and the incident that proved it: `ALIAS-LIBRARY.md`'s "Golden rules" section.

## 7. Per-song status against this scaffold

Snapshot, 2026-08-12, refreshed 2026-08-19 (`gopod_notes/GOPOLISHER_FIXES_001.md`) for
folder renames that had drifted. Re-check before trusting an old copy of this table.

| Song | Status marker | Engine | Reporter gaps | Closer (GOPOD layer only, n/a here) |
|---|---|---|---|---|
| `00_brobots_awaken` | GOLDEN | golden (control-shim) | none yet | n/a |
| `01_brobots_interview_vamp` / `02_brobots_interview_run` (split 2026-08-19 from the old combined `01_brobots_interview_section_01`) | none | dedicated (`run_section1_full_live_001.py`) | none yet | n/a |
| `101_brobots_bingo_test` | WIP (by design) | golden | mid-only, not intro/outro-named | n/a |
| `102_brobots_bingo_game` (Chocolate Bingo) | STABLE (in `bingo/README.md`) | standalone Go, not a scored song | n/a — no score to gap | n/a |
| `105_brobots_nap` (renamed/renumbered 2026-08-18 from `104_brobots_baby_robots_sleep`) | none | golden (engine's own default song) | intro + outro, clean | n/a |
| `103_gopod_is_that_you_single` / `104_gopod_is_that_you_multi` (split 2026-08-18 from the old combined `103_gopod_is_that_you`) | none | golden | none checked this pass | n/a |

**Archived (`zzz_archives/`), reference only — not held to this scaffold:**
`00_brobots_awaken_old01` (superseded by the current `00_brobots_awaken`), `brobots_bait_000`
(original per-robot forecast, pre-merge), `brobots_bait_001` (standalone wake, interview-engine
schema — structurally can't run on the golden engine as designed), `brobots_vamp_gate`
(repointed, not retired — live content now lives at
`01_brobots_interview_vamp/`), `robot_control_song_001` (the legacy engine's own
self-check song, superseded), `102_brobots_cross_persona` (archived 2026-08-12 — superseded
by the live `103_gopod_is_that_you` PTT+LLM test, above, which does the same bit for real;
reachable via the `mixup` pha0b keyword, repointed not retired).

---

## Growth rule

This file grows the same way `pha0b`'s own cockpit did — one lesson at a time, while serving
real work, never scaffolded ahead of need. When a song closes a gap listed above (adds a
closer, adopts the intro/outro reporter naming, gets a status marker), update this file's §7
row for that song in the same pass — don't let the table go stale.

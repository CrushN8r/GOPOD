# GOPOD Bingo — the scored song

> Two brobots trade fragment banter over three ball-draw rounds — rattle, call, react —
> timed and written ahead of time for the camera.
> The shareable upsell video, not the live game.

**Status, 2026-08-13 (operator's own call): LOCKED — ready for recording, pending any
last polish.**

**Naming note, brobots wire-pod layer:** this layer's robots are Brobot 1 and Brobot 2. "Doc"
and "Pip" are GOPOD-layer persona names — a different, future layer, not this one. Expected
overlap between the two layers is exactly why this distinction is held deliberately throughout
this doc, wherever Brobot 1/Brobot 2 are described.

---

## This is the scored song, not the live game

Bingo is two different things sharing one name. **This doc is the scored song** — a
scripted, comedic capture piece built for a shareable upsell video. It is not the live,
voice-triggered game the robots can also run for a real room. For the live game
(Chocolate Bingo), see [BINGO GAME.md](SONG_102_BROBOTS_1_2_BINGO_GAME.md).

---

## The upsell video

Everything below describes this scored piece, not the live game — Brobot 1 calling real
draws off a real deck, Brobot 2 reacting live, is a different piece entirely (see the
game doc above). This is a scored, 69-step comedic capture song (WIP as of 2026-08-11's
notation remap — was 57, see `songs/101_brobots_bingo_test/story.md`)
(`goverlord/runtime/songs/101_brobots_bingo_test/`), launched via
`bingo-video-song` (dry) or `bingo-video-song-live` (live, no env var needed) — see
`ALIAS-LIBRARY.md`. Neither the live-game sidecar nor its reactor is touched by it. Runs on the
golden song engine (`run_golden_song_001.py`), cut over from an earlier dedicated runner
2026-08-07.

**Live-confirmed, twice, operator's own words: "Mechanically perfect."** / "Perfect
mechanical run. Getting closer to smoother bingo banter flow." The piece runs: an opening
sync (both robots say "Brobots ready!" genuinely concurrently, over the same decoupled
direct-SDK path the opening chord uses) plus a host arm-cue gesture, interleaved fragment
banter with arm cues, head nods, and reaction beats woven in mid-conversation, three
ball-draw rounds — each with a real rattle sound effect (Brobot 1's own sidecar audio, played
over an independently-built direct-SDK connection reusing the sidecar's proven
audio-streaming code) ahead of a "Big Shiny Bingo Ball" call, a reaction number, one of
six paced emotion beats (celebrate/frustrated/angry plus interleaved happy/frustrated/
veryHappy), and a scripted line — each round's own "ball captured" cue leading into the
next beat, and a close. An audience-facing terminal/screen shows the full rich scripted line
separately from what the robot actually speaks, so what's on screen always reads clean
even where the robot-safe speech filter trims something. Comedic pacing was tuned by ear
across several passes, not left at whatever the code produced by default. The rattle's own
release-to-play settle margin was widened 2026-07-22 (1.0s → 2.0s → 3.0s, its own dedicated
constant) after an intermittent not-heard-despite-`status=OK` gap surfaced live — 3.0s
live-confirmed by the operator twice, including with the robot deliberately put to sleep
first.

---

## Future — cube danger beat at O-0 (design, not built)

Operator's confirmed design, not yet built: the cube stays **green through normal flow** —
Brobot 2's calls, the rounds, the banter — then flips to **red LEDs + the "Danger Will
Robinson" alarm sound** on the O-0 beat, this song's own existing gag where a ball comes up
"O-0" (not a real bingo number) and Brobot 2 reacts with suspicion (`host_call_0502` calls
it, `brat_say_0503` "Wait! O-0? Verifying.", `brat_beat_angry_0503`, `brat_say_0504` "this
election is rigged!" — real step_ids, current story.md). Green = good; red + alarm = the
danger beat lands.

Three pieces this ties together, each at its own real status — **none live-integrated into
this song yet**:
- **Cube LED tool** — `direct_sdk_cube_blip_001` (connect → red → green → release, plus
  `flash`/`off` modes). Live-fire confirmed 2026-08-14 against Brobot 2's cube (`status=OK`).
  Reachable via the `cube-blip`/`cube-flash`/`cube-off` aliases and phcal's own guided menu
  (`14. cube`). See `gopod_notes/CUBE_BLIP_TOOL_BUILT_001.md`,
  `gopod_notes/CUBE_BLIP_MODES_ADDED_001.md`, `gopod_notes/CUBE_ADDED_TO_PHCAL_MENU_001.md`.
- **Danger sound** — `phcal danger` (reuses the rattle direct-SDK binary, pointed at
  `danger-will-robinson.wav`). Guided-menu path confirmed correct and its own missing binary
  built (`gopod_notes/DANGER_PHCAL_BRANCH_FIXED_001.md`); no on-disk record of a live fire
  yet.
- **This song's O-0 beat** — plays today exactly as scored above, banter and animation
  only. No LED/alarm hook exists in `knobs.json`/the runner yet.

Pinned as intended design, not a build task started — bingo integration is its own later
rung, not gated on anything above changing today.

---

## Current state

**LOCKED — ready for recording, pending any last polish** (operator's own call,
2026-08-13). The 2026-08-11 rebuild moved this song to a new step-naming notation (69
steps, 8 cut from the prior 57-step shape — that prior GOLDEN 57-step run is preserved as
a reference snapshot, not lost). The "mechanically perfect" quotes above are from that
earlier 57-step shape; they describe the song's proven mechanics, carried forward into
the rebuild, not a separately re-confirmed live run of the new 69-step shape specifically.

---

> From Doctrine Barfallonyou
> Lesson! If the room is watching two robots disagree, the room is already yours.
> Boom! Done! Class Dismissed!
> — Doc Squawkadoodle

---

## GOPOD YAHMM (You Are Here Mall Map)

Two doors. Pick the one that fits how far in you want to go — nothing here needs a set order beyond that.

### Door 1 — New here? Explore GOPOD wide

Plain language, first look, no background needed — for a newcomer, human or AI.

**Start here**
- [README.md](../../README.md) — what GOPOD is and how it's built
- [AWAKEN.md](SONG_00_BROBOTS_1_2_AWAKEN.md) — watch first: a brobot wakes, checks itself, greets you
- [QUICKSTART.md](../SINGLE_BOT_QUICKSTART.md) — talk to your own Vector, one robot, no alias studio needed
- [MY_NICHE_BUZZ_ASK.md](../../MY_NICHE_BUZZ_ASK.md) — help test the keyboard grabber, no robot required
- [GOPOD_SONGS.md](GOPOD_SONGS.md) — all songs, explained in plain language, no background needed
- [PALM_TREE.md](../../life/01_PALM_TREE.md) — the whole thing, put together, no background needed
- [FUNNY_NAMINGS.md](../../web/FUNNY_NAMINGS.md) — every name, character, and phrase this project uses, explained once

**The songs**
- [INTERVIEW VAMP.md](SONG_01_BROBOTS_1_2_INTERVIEW_VAMP.md) — the flagship's video 1, the pre-show banter
- [INTERVIEW RUN.md](SONG_02_BROBOTS_1_2_INTERVIEW_RUN.md) — the flagship's video 2, the seven exchanges
- [BINGO GAME.md](SONG_102_BROBOTS_1_2_BINGO_GAME.md) — the live two-brobot Bingo warm-up act
- [BABY ROBOTS SLEEP.md](SONG_105_BROBOTS_1_2_BABY_ROBOTS_SLEEP.md) — Doc's origin, told as a bedtime story
- [IS-THAT-YOU SINGLE.md](SONG_103_BROBOTS_1_2_IS-THAT-YOU_SINGLE.md) — the single-bot jewel, live and unscripted
- [IS-THAT-YOU MULTI.md](SONG_103_BROBOTS_1_2_IS-THAT-YOU_MULTI.md) — the two-brobot mix-up, live and unscripted

**The content engine**
- [NICHE_PILLARS.md](../../web/NICHE_PILLARS.md) — how the writing is split into nine kinds of everyday maths, and why
- [AI_WORDPLAY.md](../../web/AI_WORDPLAY.md) — the engine: AI Wordplay, the contests that feed it, and where the content lands
- [AHA_MOMENT.md](../../web/AHA_MOMENT.md) — a live demo the reader can feel work
- [FOODMATH_AHA_MOMENT.md](../../web/FOODMATH_AHA_MOMENT.md) — the foodmath cousin: a live subdomain built in hours, the gap between live and built is the demo
- [BIRTHDAY.md](../../web/BIRTHDAY.md) — the physical proof one — real food-car props, no rendering required
- [NEWSLETTER.md](../../web/NEWSLETTER.md) — subscribe to CRUSHN8R CREW'd — the live follow-along lane

**For venues, funders, and partners**
- [MOBILE_GEAR.md](../MOBILE_GEAR.md) — mobile deployment and field kit
- [OPS ASK.md](../../MY_GOPOD_OPS_ASK.md) — the operator's ops ask — social, sites, and content, a different role than the technical one
- [HEALTHY_DISTRACTIONS.md](../../life/02_HEALTHY_DISTRACTIONS.md) — GOPOD as healthy distraction
- [OUTREACH.md](../../life/02a_OUTREACH.md) — the community and paid outreach plan, two lanes side by side

**For teachers**
- [EDUCATION.md](../../life/03_EDUCATION.md) — GOPOD as a teaching tool
- [TEACHER_INSIGHT.md](../../life/04_TEACHER_INSIGHT.md) — what a session shows a teacher about their room

### Door 2 — More? Dive GOPOD deep

For readers who lean in, who know GitHub — the technical docs, the operator tooling, how it's made.

**Help wanted**
- [MY_GOPOD_ASK.md](../../MY_GOPOD_ASK.md) — the operator's own ask — what's built, where the line is, what kind of help this needs

**For Wire-Pod owners and builders**
- [WIRED-POD.md](../WIRED-POD.md) — what GOPOD changed and added on top of Wire-Pod
- [GOPOD_FEATURES.md](../GOPOD_FEATURES.md) — everything GOPOD built, feature by feature, stack included
- [BODIED_BROBOTS.md](../BODIED_BROBOTS.md) — the robot bodies — proof ladder, Vector L3, Cozmo/Scout staged below L3
- [ALIAS-LIBRARY.md](ALIAS-LIBRARY.md) — every shortcut command GOPOD can run
- [ALIAS-SEQUENCER.md](ALIAS-SEQUENCER.md) — arrangement: notes into sequences into songs
- [PHA0B.md](PHA0B.md) — PLAYHEAD Part 2, the performance front door, whole songs at a time
- [PHCAL.md](PHCAL.md) — PLAYHEAD Part 1, the calibration bench, one primitive at a time

**Doc's Take — lessons learned**
- [DOCS_TAKE_LESSON_1.md](../../life/101_DOCS_TAKE_LESSON_1.md) — the first real mistake GOPOD survived
- [DOCS_TAKE_LESSON_2.md](../../life/102_DOCS_TAKE_LESSON_2.md) — fix the foundation, not the symptom
- [DOCS_TAKE_LESSON_3.md](../../life/103_DOCS_TAKE_LESSON_3.md) — a placeholder that looks finished is worse than an honest gap
- [DOCS_TAKE_LESSON_4.md](../../life/104_DOCS_TAKE_LESSON_4.md) — a living thing sheds, and shedding shows you what is left
- [DOCS_TAKE_LESSON_5.md](../../life/105_DOCS_TAKE_LESSON_5.md) — the scary answer is usually the true one, not the easy binary
- [DOCS_TAKE_LESSON_6.md](../../life/106_DOCS_TAKE_LESSON_6.md) — the honest edge isn't the ask — say it out loud

**The philosophy**
- [LEGACY.md](../../life/05_LEGACY.md) — the pedagogy behind what GOPOD propagates

**How this gets made**
- [AI_AHA_MOMENTS.md](../../web/AI_AHA_MOMENTS.md) — the aha moments from making GOPOD with AI, for the reader who wants to see how the thing thinks

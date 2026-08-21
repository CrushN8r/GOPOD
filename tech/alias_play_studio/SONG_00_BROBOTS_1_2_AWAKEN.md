# GOPOD Awaken

> A brobot wakes, proves its own arms and head still work, and checks the weather. No cloud
> dependency in any of it — just a brobot doing its own pre-flight out loud.

**Watch first.** Until a captured clip of the opener exists, this page is the real place to
start — same first-look role, in words instead of video.

---

## What Awaken is

What `brobots_awaken` plays today: one robot connects, tests its own arm, tests its own
head nod, checks the real weather, delivers a self-ID payoff line ("Ready to party! Did
someone say GOPOD Yourself?"), then exits — 16 steps, each self-narrated, with a silent
reporter-gap pause after every major beat (0s live — same reporter-gap convention Bingo's
own gaps use, left open for an edited-in reporter voiceover, not live dead-air) so the
piece can be cut cleanly for video. No LLM anywhere in it; the only "generated" content is
the weather note's own live fetch.

It rides on the same single-robot instrument as `robot_control_song_001` (the same runner
that plays `start-the-control-song`'s arm/nod/fireworks self-check) — this song just points
that runner at a longer, capture-video-paced score.

**Identity swap, 2026-07-24:** this song used to be the pure 3-step weather report; the
operator swapped its content by hand for the fuller merged capture video (formerly
`brobots_bait_002`) — his own call: "it IS the awaken video." The original weather-only
song was renamed `brobots_bait_000` and archived, then later decluttered entirely
(`gopod_notes/ZZZ_ARCHIVES_DECLUTTER_EXECUTED_001.md`, 2026-08-15) — pure scratch
content, nothing left to revive, no longer on disk.

---

## Two robots, two formats

The weather note inside this longer piece still varies by robot — same fetch, same
moment, two different personalities expressed through unit choice:

- **Brobot 1:** Celsius, 24-hour clock.
- **Brobot 2:** Fahrenheit, 12-hour clock.

`gopod_weather_fetch_001.py`'s own `format_for_robot()` and `load_robot_format_from_jdocs()`
own that formatting, reading each robot's own live Wire-Pod jdocs `RobotSettings` to decide
lead unit/clock ordering — pre-existing, untouched by this song, reused as-is. (The static
`robot_weather_format.json` this used to read was retired 2026-08-13,
`WEATHER_LIVE_JDOCS_SOURCED_001.md`, and no longer exists.) Nothing in the song itself is
hardcoded.

---

## How it runs

- **`pha0b` → pick `00_brobots_awaken`** (keyword `bait`) — the golden path
  (`run_golden_song_001.py`, cut over 2026-08-07), playhead-sliceable by step, full 16-step
  capture piece on a full-range pick. Targets Brobot 1 by default; `pha0b`'s own robot
  prompt picks otherwise. `start-the-bait-song` (the old legacy-engine alias) was retired
  2026-08-11 — stale, no crucial use, easily rebuilt if ever needed again.
- **`gopod-weather-say`** — a standalone single-note weather check, no connect/arm/nod
  ceremony, if all that's wanted is a quick weather line.
- The *pure* weather-only 3-step version (this song's own former self) is gone —
  decluttered off disk entirely 2026-08-15 (`gopod_notes/ZZZ_ARCHIVES_DECLUTTER_EXECUTED_001.md`).
  The `pha0b weather <a> <b>` keyword itself was dropped 2026-08-16 (dead reference to
  that now-missing folder, no live equivalent) — same purge as `start-the-weather-song`,
  see `PHA0B_WEATHER_DEAD_PATH_FIXED_001.md`.

Dry by default; export `GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1` first for real hardware, same
convention as every other control-song alias.

---

## Where it fits

A shareable capture piece, not a live performance bit — built to be recorded once, not
watched live in front of a room the way Bingo or the Interview are. *(GOPOD layer — not
built yet: Bingo and the Interview are also recorded content today, not watched live in
front of a real room — same Wire-Pod-layer status as this song.)* Proves a robot can
wake, self-check, and reach outside itself for one real fact, all in one clean take.

Worth naming plainly: the weather note specifically needs a live network fetch to say
anything real. Every other beat in this song runs cold with zero signal; that one moment
is the deliberate exception — GOPOD reaching outside itself without breaking the
local-first stack around it.

In the campaign's own funnel, this is the bait — the free first taste, short enough to
catch a stranger's attention before they've committed to anything.

---

## Current state

**Golden — live-confirmed 2026-08-10.** Two clean back-to-back runs, Brobot 1 then Brobot 2,
all 16 steps succeeding both times, zero failures, weather fetch correct (location, date,
per-robot units) on both runs. Content swap itself dry-verified 2026-07-24 against the song's
new `brobots_awaken` home. Format source at that confirmation was `gopod_weather_fetch_001.py` reading the now-retired
static `robot_weather_format.json`; that file was retired 2026-08-13 in favor of live
per-robot Wire-Pod jdocs sourcing (`WEATHER_LIVE_JDOCS_SOURCED_001.md`) — this song's own
golden run has not been re-confirmed against the new source. Song itself unchanged, nothing
authored here.

---

> From Doctrine Barfallonyou
> Lesson! You don't need a script to prove you're alive. You need one real fact, said in your own voice.
> Boom! Done! Class Dismissed!
> — Doc Squawkadoodle

---

## GOPOD YAHMM (You Are Here Mall Map)

Two doors. Pick the one that fits how far in you want to go — nothing here needs a set order beyond that.

### Door 1 — New here? Explore GOPOD wide

Plain language, first look, no background needed — for a newcomer, human or AI.

**Start here**
- [README.md](../../README.md) — what GOPOD is and how it's built
- [QUICKSTART.md](../SINGLE_BOT_QUICKSTART.md) — talk to your own Vector, one robot, no alias studio needed
- [MY_NICHE_BUZZ_ASK.md](../../MY_NICHE_BUZZ_ASK.md) — help test the keyboard grabber, no robot required
- [GOPOD_SONGS.md](GOPOD_SONGS.md) — all songs, explained in plain language, no background needed
- [PALM_TREE.md](../../life/01_PALM_TREE.md) — the whole thing, put together, no background needed
- [FUNNY_NAMINGS.md](../../web/FUNNY_NAMINGS.md) — every name, character, and phrase this project uses, explained once

**The songs**
- [INTERVIEW VAMP.md](SONG_01_BROBOTS_1_2_INTERVIEW_VAMP.md) — the flagship's video 1, the pre-show banter
- [INTERVIEW RUN.md](SONG_02_BROBOTS_1_2_INTERVIEW_RUN.md) — the flagship's video 2, the seven exchanges
- [BINGO.md](SONG_101_BROBOTS_1_2_BINGO.md) — the shareable upsell video
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

# GOPOD Is That You — Single

> Say a name. Get a real answer back — one robot, live, not scripted.

**Naming note, exception:** this song runs on brobots wire-pod layer hardware, but its
content is deliberately framed as "leaked GOPOD-layer test footage" (per `niche-buzz`'s own
funnel role for this song). Because it's testing/previewing the GOPOD layer by design,
"Doc"/"Pip" is the correct, in-character naming throughout this doc — unlike other
wire-pod-layer songs, which use Brobot 1/Brobot 2.

---

## What Is That You — Single is

The single-bot jewel: hold only KP1 (Doc), one robot, the on-ramp made visible. Hold the
key, say a name, get a real answer back — decided live, a canned line or a live Ollama
reply. Nothing here is scripted; the whole reason to shoot it is that it isn't.

This is the same real thing a single-Vector owner could run themselves, per
[QUICKSTART.md](../SINGLE_BOT_QUICKSTART.md) — proof that "your one robot, answering
live" isn't a demo trick, it's the real mechanism, on camera.

Wired as `103_gopod_is_that_you_single/`, `pha0b` keyword `itsyou-single`. Bookends the
exact same live alias (`is-that-you()` in `~/.gopod_alias_lib/demo.sh`) the two-robot
version does — the writer (`gopod_ptt_chat_writer_013.py`) already treats KP1/KP2 as
fully independent, self-contained key handlers, so holding only one key already ran
single-robot before this split ever existed. "Single" scopes the recording, not the
code. See this song's own `story.md` for the full mechanical breakdown.

There's also a two-robot version of this same idea:
[IS-THAT-YOU MULTI.md](SONG_103_BROBOTS_1_2_IS-THAT-YOU_MULTI.md) — hold either key, Doc
or Pip, watch the mix-up happen live.

## Honest state

Wired into `pha0b` (`itsyou-single`), dry-verified both directly and through the menu
picker — **not yet live-run**. No actual Is That You recording session has been shot
through this bookend yet. This is exactly where the jewel stands today: real,
live-capable, fully wired — waiting on the take, not on more building.

---

## History — the earlier scripted reel

Before this split, "Is That You" also named something different: a short, four-line
**scripted** demo reel — fixed dialogue, no live capture at all. Archived 2026-08-12 to
`zzz_archives/102_brobots_cross_persona/` once the real, live PTT bookends made a
scripted stand-in redundant. Still reachable, repointed not retired, via the `mixup`
`pha0b` keyword — kept for the record, not deleted.

**The four lines**, for reference:

**Doc, self-confirming:** *"Doc here. Yes. Try to keep up."*
**Doc, naming the mix-up:** *"That is Pip's lane. Wrong robot, right confusion."*
**Pip, self-confirming:** *"Yep. Pip here. I think that was my cue."*
**Pip, naming the mix-up:** *"Hey Doc, is this when I ask for emails?"*

Built from `persona_awareness_reply()`'s own four canned cross-persona lines — the one
fixed-content piece of an otherwise fully interactive writer (keypress wait, live mic
capture, Vosk transcription, live Ollama fallback — none of that fits a scripted step
list, so none of it made the port). Played back to back, once, as a demo of what the
live reply sounds like — not a re-enactment of how the real thing ever runs, since in the
live system exactly one of the four lines fires per utterance, decided live.

**Naming history**: this reel first lived at `103_gopod_is_that_you/`, renamed
2026-07-31 to `brobots_cross_persona` (folder `102_brobots_cross_persona`) once it
collided with the live PTT alias's own `is-that-you` name — two different things sharing
one name. The live demo kept `is-that-you`; the scripted reel moved to a non-colliding
one. That same `103_gopod_is_that_you` slot was later reused (2026-08-08) for the live
bookend songs, split (2026-08-18) into `_single`/`_multi` runtime folders, and finally
into this doc and its sibling (2026-08-18, same day) to match.

---

> From Doctrine Barfallonyou
> Lesson! Knowing your own name is easy. Knowing when someone else answered to it first — that's the interesting part.
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
- [BINGO.md](SONG_101_BROBOTS_1_2_BINGO.md) — the shareable upsell video
- [BINGO GAME.md](SONG_102_BROBOTS_1_2_BINGO_GAME.md) — the live two-brobot Bingo warm-up act
- [BABY ROBOTS SLEEP.md](SONG_105_BROBOTS_1_2_BABY_ROBOTS_SLEEP.md) — Doc's origin, told as a bedtime story
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

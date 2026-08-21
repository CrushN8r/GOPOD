# Single-Bot Quickstart

> One Vector. Free-form LLM chat. No second robot, no Jetson, no alias studio required.

This is GOPOD's other door. Everything else in this repo (`pha0b`, `phcal`, the
alias-studio tooling) is built for a collaborator running two robots and a full
performance rig — the pro lane. This page is the adoption lane: what a regular Vector
owner, with one robot and some technical comfort, can actually get running today to
talk to their own robot with an LLM behind it.

**Build-in-public, this whole page.** What works is real. What's broken is named
plainly, not hidden. Truth-checked against actual code on 2026-08-18 — see
`gopod_notes/SINGLE_BOT_ONRAMP_TRUTH_001.md` for the full survey this page is built on.

---

## What actually works today

- **Free-form LLM chat, two real paths.** Anything you say that isn't a hard-coded
  intent reaches an LLM — either through Wire-Pod's own native backpack/knowledge-graph
  chat lane (stock Wire-Pod, works out of the box, pointable at any OpenAI-compatible
  endpoint via `apiConfig.json`), or through GOPOD's own PTT writer script, which makes
  a real, working call to Ollama's OpenAI-compatible endpoint.
- **Mic detection that isn't hardcoded to one operator's hardware.** `gopod-mic-detect`
  / `gopod-mic-set` / `gopod-mic-test` do real dynamic USB-input discovery against a
  config file, not a fixed device name — this genuinely ports to a different USB mic.
- **Numpad/keyboard capture with real device auto-discovery.** The writer finds a
  numpad or keyboard by its actual key-capability bitmap first, name-keyword matching
  second — not hardcoded to one physical device.
- **A real output sanitizer, live in the path that uses it.** Before anything reaches
  the robot's mouth (through the PTT writer's own dispatch), the reply gets stripped of
  wrapping quotes, asterisks, emoji, and run through a pronunciation-normalization
  filter. This is wired in and actually runs, not a paper guarantee.

None of the above needs a Jetson, a second Vector, or `pha0b`/`phcal`. The PTT writer
is a standalone script — install its dependencies, point it at your one robot, run it.

## Where it's honestly not there yet

**This is not plug-and-play today.** Two real blockers stand between "clone the repo"
and "just talk to your robot":

- **No spoken wake word.** Saying "GOPOD Yourself" out loud doesn't currently start
  anything — the custom intent that's supposed to catch that phrase points at a script
  file that no longer exists on disk. Today's real front door is running the PTT writer
  directly from a terminal, not speaking a phrase at it.
- **Robot targeting is hardcoded, not configured.** Which robot KP1/KP2 talk to is a
  literal serial number baked into the Python source, set to this operator's own two
  Vectors. Pointing it at a different robot means editing source, not editing a config
  file — a real, if small, barrier for a new owner.

One more gap worth naming honestly: the sturdier of GOPOD's two output-safety
mechanisms — a tested validator-and-repair pass built in the Wire-Pod Go layer — exists
in the codebase but has no live caller anywhere. It's real, built code sitting unplugged.
The native Wire-Pod chat lane (the easiest one for a casual owner to reach) currently
runs with no GOPOD-added output guard at all.

## What it takes to actually run this today

Real, non-trivial technical setup — not casual-owner plug-and-play, but not entangled
with GOPOD's two-robot orchestration either:

- Ollama running locally with a model pulled (or an OpenAI-compatible endpoint you
  already have)
- `vosk` + a downloaded Vosk speech model, `sounddevice`, `scipy`, `numpy`
- Wire-Pod already running against your one robot
- Run the PTT writer directly (`python3 gopod_ptt_chat_writer_013.py`, with
  `--list-devices` to find your mic/keyboard first)

## Invitations — this is where a contributor could help

These are scoped, honest, not-yet-built fixes — not a promise of when, and nothing here
is built by this page. If you're the kind of person `MY_GOPOD_ASK.md` is talking to,
this is a concrete, contained place to start:

1. **De-hardcode the robot serials.** Move `PERSONA_ROBOT_SERIAL`/`PERSONA_BY_KEY` out
   of Python source and into the same config-driven pattern the rest of the writer
   already uses, so a new owner edits a config file, not code.
2. **Fix the spoken wake trigger.** Repoint `GOPOD_YOURSELF`'s custom-intent exec path
   at something that exists, or replace it with a working equivalent — or, short of a
   fix, document the terminal-run path as the real current front door instead of
   implying a working spoken wake exists.
3. **Wire in the built output guard.** Either connect the tested Go
   validator-and-repair pass to the native Wire-Pod chat lane, or clearly document that
   only the PTT writer's narrower path is currently guarded.

None of these are large. All three are exactly the kind of "past my ceiling, contained
enough to grab" piece `MY_GOPOD_ASK.md` describes.

---

> From Doctrine Barfallonyou
> Lesson! A door that's honestly half-open beats one that's dishonestly wide.
> Boom! Done! Class Dismissed!
> — Doc Squawkadoodle

---

## GOPOD YAHMM (You Are Here Mall Map)

Two doors. Pick the one that fits how far in you want to go — nothing here needs a set order beyond that.

### Door 1 — New here? Explore GOPOD wide

Plain language, first look, no background needed — for a newcomer, human or AI.

**Start here**
- [README.md](../README.md) — what GOPOD is and how it's built
- [AWAKEN.md](alias_play_studio/SONG_00_BROBOTS_1_2_AWAKEN.md) — watch first: a brobot wakes, checks itself, greets you
- [MY_NICHE_BUZZ_ASK.md](../MY_NICHE_BUZZ_ASK.md) — help test the keyboard grabber, no robot required
- [GOPOD_SONGS.md](alias_play_studio/GOPOD_SONGS.md) — all songs, explained in plain language, no background needed
- [PALM_TREE.md](../life/01_PALM_TREE.md) — the whole thing, put together, no background needed
- [FUNNY_NAMINGS.md](../web/FUNNY_NAMINGS.md) — every name, character, and phrase this project uses, explained once

**The songs**
- [INTERVIEW VAMP.md](alias_play_studio/SONG_01_BROBOTS_1_2_INTERVIEW_VAMP.md) — the flagship's video 1, the pre-show banter
- [INTERVIEW RUN.md](alias_play_studio/SONG_02_BROBOTS_1_2_INTERVIEW_RUN.md) — the flagship's video 2, the seven exchanges
- [BINGO.md](alias_play_studio/SONG_101_BROBOTS_1_2_BINGO.md) — the shareable upsell video
- [BINGO GAME.md](alias_play_studio/SONG_102_BROBOTS_1_2_BINGO_GAME.md) — the live two-brobot Bingo warm-up act
- [BABY ROBOTS SLEEP.md](alias_play_studio/SONG_105_BROBOTS_1_2_BABY_ROBOTS_SLEEP.md) — Doc's origin, told as a bedtime story
- [IS-THAT-YOU SINGLE.md](alias_play_studio/SONG_103_BROBOTS_1_2_IS-THAT-YOU_SINGLE.md) — the single-bot jewel, live and unscripted
- [IS-THAT-YOU MULTI.md](alias_play_studio/SONG_103_BROBOTS_1_2_IS-THAT-YOU_MULTI.md) — the two-brobot mix-up, live and unscripted

**The content engine**
- [NICHE_PILLARS.md](../web/NICHE_PILLARS.md) — how the writing is split into nine kinds of everyday maths, and why
- [AI_WORDPLAY.md](../web/AI_WORDPLAY.md) — the engine: AI Wordplay, the contests that feed it, and where the content lands
- [AHA_MOMENT.md](../web/AHA_MOMENT.md) — a live demo the reader can feel work
- [FOODMATH_AHA_MOMENT.md](../web/FOODMATH_AHA_MOMENT.md) — the foodmath cousin: a live subdomain built in hours, the gap between live and built is the demo
- [BIRTHDAY.md](../web/BIRTHDAY.md) — the physical proof one — real food-car props, no rendering required
- [NEWSLETTER.md](../web/NEWSLETTER.md) — subscribe to CRUSHN8R CREW'd — the live follow-along lane

**For venues, funders, and partners**
- [MOBILE_GEAR.md](MOBILE_GEAR.md) — mobile deployment and field kit
- [OPS ASK.md](../MY_GOPOD_OPS_ASK.md) — the operator's ops ask — social, sites, and content, a different role than the technical one
- [HEALTHY_DISTRACTIONS.md](../life/02_HEALTHY_DISTRACTIONS.md) — GOPOD as healthy distraction
- [OUTREACH.md](../life/02a_OUTREACH.md) — the community and paid outreach plan, two lanes side by side

**For teachers**
- [EDUCATION.md](../life/03_EDUCATION.md) — GOPOD as a teaching tool
- [TEACHER_INSIGHT.md](../life/04_TEACHER_INSIGHT.md) — what a session shows a teacher about their room

### Door 2 — More? Dive GOPOD deep

For readers who lean in, who know GitHub — the technical docs, the operator tooling, how it's made.

**Help wanted**
- [MY_GOPOD_ASK.md](../MY_GOPOD_ASK.md) — the operator's own ask — what's built, where the line is, what kind of help this needs

**For Wire-Pod owners and builders**
- [WIRED-POD.md](WIRED-POD.md) — what GOPOD changed and added on top of Wire-Pod
- [GOPOD_FEATURES.md](GOPOD_FEATURES.md) — everything GOPOD built, feature by feature, stack included
- [BODIED_BROBOTS.md](BODIED_BROBOTS.md) — the robot bodies — proof ladder, Vector L3, Cozmo/Scout staged below L3
- [ALIAS-LIBRARY.md](alias_play_studio/ALIAS-LIBRARY.md) — every shortcut command GOPOD can run
- [ALIAS-SEQUENCER.md](alias_play_studio/ALIAS-SEQUENCER.md) — arrangement: notes into sequences into songs
- [PHA0B.md](alias_play_studio/PHA0B.md) — PLAYHEAD Part 2, the performance front door, whole songs at a time
- [PHCAL.md](alias_play_studio/PHCAL.md) — PLAYHEAD Part 1, the calibration bench, one primitive at a time

**Doc's Take — lessons learned**
- [DOCS_TAKE_LESSON_1.md](../life/101_DOCS_TAKE_LESSON_1.md) — the first real mistake GOPOD survived
- [DOCS_TAKE_LESSON_2.md](../life/102_DOCS_TAKE_LESSON_2.md) — fix the foundation, not the symptom
- [DOCS_TAKE_LESSON_3.md](../life/103_DOCS_TAKE_LESSON_3.md) — a placeholder that looks finished is worse than an honest gap
- [DOCS_TAKE_LESSON_4.md](../life/104_DOCS_TAKE_LESSON_4.md) — a living thing sheds, and shedding shows you what is left
- [DOCS_TAKE_LESSON_5.md](../life/105_DOCS_TAKE_LESSON_5.md) — the scary answer is usually the true one, not the easy binary
- [DOCS_TAKE_LESSON_6.md](../life/106_DOCS_TAKE_LESSON_6.md) — the honest edge isn't the ask — say it out loud

**The philosophy**
- [LEGACY.md](../life/05_LEGACY.md) — the pedagogy behind what GOPOD propagates

**How this gets made**
- [AI_AHA_MOMENTS.md](../web/AI_AHA_MOMENTS.md) — the aha moments from making GOPOD with AI, for the reader who wants to see how the thing thinks

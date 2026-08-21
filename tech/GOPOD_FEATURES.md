# GOPOD Features

> Wire-Pod gets one robot talking. GOPOD gets two robots arguing, a room laughing, and a
> browser tab watching it happen live.
> `wire-pod < GOPOD > cockpit web layer` — this is the map of everything in the middle.

*The tagline above describes GOPOD layer (Point B — not built yet): a real room and a
browser tab both watching the same live session. Two robots arguing is real today, in the
built songs — Wire-Pod layer. The room/browser-tab claim is the not-yet-built part.*

---

## The short version

Wire-Pod is a great single-robot voice server: wake word in, STT, an LLM reply, TTS out. That's
the whole job it was built to do, and it does it well.

GOPOD starts there and keeps building. Same Wire-Pod underneath, unmodified in spirit — but
layered on top: two robots instead of one, deterministic character identity instead of LLM
improv, custom intents that launch entire scripted experiences instead of canned replies, a
cross-robot event bus with zero code coupling, and a live web cockpit watching the whole session
happen in real time. Everything below is real and running, not a pitch deck — each section says
plainly what's confirmed live versus still in progress.

Three layers, one project, and the rest of this repo uses their names without re-explaining
them: **Layer 0** (Point A) is plain Wire-Pod, the software anyone with one of these robots
already has. **Layer 1** (Point 0) is the Brobots — what's actually built and running today.
**Layer 2** (Point B) is GOPOD — Doc and Pip in a live conversation together, not built yet.
Layer 1 is what runs now; Layer 2 is the goal everything else is aimed at.

---

## Where GOPOD starts: stock Wire-Pod

For contrast — what you get before any of this:

- One robot, one voice, one request/response loop
- A fixed set of built-in intents (weather, timer, dance, etc.), each a canned action
- No concept of a second robot, no shared state between sessions
- No web dashboard — Wire-Pod's own config page is for setup, not for watching a session
- No content layer — nothing scripted plays out over multiple exchanges

That's the floor. Everything past here is what got built on top of it.

---

## Layer 1 — the Brobots layer (Point 0)

### One wake phrase, two robots

"Hey Vector, GOPOD yourself" fires a custom intent that hands off from Wire-Pod to a Python
session loop entirely outside Wire-Pod's own request/response cycle — no cloud round-trip, no
companion app.

```
KP1 → Brobot 1  (vector1 / 0dd1b9e9) — the doctrine cannon
KP2 → Brobot 2  (vector2 / 0dd1d8bf) — the clueless sidekick
KP0 (triple-tap) → clean exit
```

One phrase, one session, two robots in it.

### Three custom intents that do more than reply

Stock Wire-Pod intents are canned actions. GOPOD's three custom intents are launch points for
entire subsystems, all deployed cold-restart persistent (survive a full service reboot, verified
every session):

| Intent | What it launches |
|---|---|
| `GOPOD_YOURSELF` | The full PTT session loop — wake phrase to live two-robot conversation |
| `BROBOTS_INTERVIEW` | A generated, seven-exchange scripted interview between both robots |
| `BROBOTS_BINGO` | A spoken, timed Bingo round with an optional cross-robot reaction |

Wire-Pod's job stays exactly what it's good at: match the utterance, hand off, get out of the
way. All the session logic lives in Python, not in Wire-Pod itself.

### Deterministic identity, not LLM improvisation

Nothing about *who* Brobot 1 and Brobot 2 are is left for the LLM to invent mid-conversation. Identity is
locked at the prompt layer before generation ever starts — so the two robots stay in character
across an entire session instead of drifting.

### A hardened free-form lane

Wire-Pod's native LLM fallback path (the "backpack" lane) got its own four-tier prompt stack, STT
phrase normalization for Vector's mic quirks (`"bro bought"` → `BROBOT`, `"go pod"` → GOPOD
Yourself), and a configurable expression valve — kept in files separate from the interview system
so either can change without touching the other.

### A robot-safe output guard

Raw LLM output can contain animation syntax Wire-Pod doesn't understand and will crash on. Every
response is filtered through an allowlist of valid animation tags, length limits, and markdown/
emoji/invisible-Unicode stripping before it ever reaches the robot. This is also where the DAG
panic and three `weather.go` crash points got converted from hard panics to logged errors —
Wire-Pod itself got more robot-safe, not just GOPOD's output.

### A pre-demo readiness gate

`gopod-pre-demo` checks Wire-Pod's service health, both robots' reachability, the LLM's warm
state, and a speech test — all before a live session is allowed to start, so a demo doesn't fail
in front of a room for a reason that could've been caught in five seconds beforehand.

---

## The interview system — scripted and generated

The interview is where hardware delivery, LLM generation, and flow control operate together on
purpose. A Section Card defines the arc; a Template defines the runtime rules and how each
robot's lines are meant to sound; the runner generates all exchanges up front via Ollama, logs
them, then plays them back — Brobot 1 and Brobot 2 alternating, each speaking through Wire-Pod's own
`say_text` API by ESN, through the robot's own onboard voice.

A Kokoro `af_bella` narrator reads each status line aloud *before* the robots start, then goes
silent the moment playback begins, so the room never hears the narrator and a robot step on each
other. What the screen displays and what the robot's mouth speaks are now the same cleaned,
pronunciation-corrected line — a display that used to show the raw model draft while the robot
spoke a filtered version now shows exactly what's heard.

A shorter, sibling piece — `brobots_bait_001` — reuses the same canned-and-cleaned delivery
pipeline for a much smaller job: two robots wake, gesture, each speak one self-naming line, done
in under a minute. No LLM, no interview arc, built for a short capture rather than a performance.

Stock Wire-Pod has no concept of a multi-exchange scripted arc and no status narration layer.
This entire system sits above it.

---

## Bingo — the event-bus proof of concept

Bingo looks like party trivia. Underneath, it's the system that proved GOPOD's core
architectural bet: **one robot can drive another robot's behavior through nothing but a shared
event log — no API call, no shared process, no code coupling.**

- Brobot 1 calls a configurable 75- or 90-ball deck, with a `--silent` mode for loud rooms
- Every ready/draw/reveal/end moment writes a chat envelope to a shared JSON stream — schema-
  validated, invalid envelopes logged and dropped rather than corrupting the feed
- Brobot 2 runs as a completely independent watcher process that reacts to that same
  stream with an angry animation every time a number lands — it has zero import, zero call site,
  and zero awareness in Brobot 1's code, and Brobot 1 has zero awareness of it back
- One touch starts a round; a tuned auto-draw ticker (3 seconds between numbers, tuned up from
  1 second specifically to give Brobot 2's reactor room to connect and animate before the next draw)
  takes it from there
- As of the latest pass, starting Bingo starts Brobot 2's reaction automatically in the background —
  one command, one terminal, both robots' logs interleaved — while the two processes stay just as
  independent underneath

That "shared envelope stream, zero coupling" pattern is now the pattern for everything else in
GOPOD that involves more than one moving part: the interview runner, the PTT writer, and the
cockpit below all read and write the same kind of stream. Bingo is the small system that proved
the big one.

### The Bingo video song — a separate, scored capture piece

The live game above is Brobot 1 calling real draws. A separate piece, built for a shareable upsell
video, scores a fixed comedic bit instead: a 46-step capture song
(`goverlord/runtime/songs/101_brobots_bingo_test/`), launched via `bingo-video-song`/
`bingo-video-song-live`. **Live-confirmed, twice, operator's own words: "Mechanically
perfect."**

- A real rattle sound effect before every ball call — Brobot 1's own sidecar audio, played over an
  independently-built direct-SDK connection that reuses the sidecar's proven audio-streaming
  code rather than reinventing it
- Physical gestures woven into the banter — arm cues and head nods fire mid-conversation, not
  just canned animations
- Six paced emotion beats (celebrate, frustrated, angry, happy, veryHappy) — the same golden
  dispatch shape that took `angry` from a repeated live crash to consecutive clean runs, reused
  here across both robots
- A rich on-screen display separate from what the robot speaks — the audience-facing terminal
  shows the full scripted line even where the robot-safe speech filter lightly trims something
- Comedic pacing tuned by ear across several live passes, not left at whatever the code produced
  by default

The rattle's release-to-play settle margin was widened 2026-07-22 (its own dedicated
constant, up from the shared 1.0s, settled at 3.0s) after an intermittent
not-heard-despite-`status=OK` gap surfaced live — 3.0s live-confirmed by the operator
twice, including with the robot deliberately put to sleep first.

---

## Layer 2 — the GOPOD cockpit web page (Point B)

Everything above runs headless — a room full of people watching two robots doesn't need a
screen. *(GOPOD layer — not built yet: no room full of people has watched this.)* The
cockpit is the screen anyway: a local web server that renders the same envelope
stream everything else reads, live, in a browser tab.

- Shows active robot, PTT state, and a running chat/session log in real time
- Wire-Pod's own boot and debug logs get prefilled into the same feed on every server start —
  the "is Wire-Pod actually healthy" question answered without opening a second terminal
- Bingo's draws render in the same pane, off the same envelope stream, with no Bingo-specific
  wiring required
- A public, static preview of this same cockpit shell shipped to GitHub Pages — the real layout
  (camera panes, weather, session chat, Bingo draws, side by side), with live feeds still to come

Stock Wire-Pod's own web page is a configuration form. This is a session dashboard — a different
job entirely, built for an audience, not an installer. *(GOPOD layer — not built yet: the
dashboard code exists, but the live audience it's built for doesn't yet.)*

---

## The platform underneath the platform

A few things that don't announce themselves in a session but are why the rest of this holds
together:

- **Data-driven configuration** — endpoints, paths, audio settings, and the envelope schema
  itself live in versioned config files with a validating loader, not scattered as hardcoded
  strings across scripts
- **A Layer 0 / Layer 1 firewall** — character names (Brobot 1, Brobot 2) are kept out of core logic
  entirely, so the underlying session engine doesn't know or care what personas are wearing it
  (see [PALM_TREE.md](../life/01_PALM_TREE.md) for the soil/canopy version of this same rule)
- **A public, documented repo** — GOPOD ships with a full doctrine set (this file plus twelve
  others) explaining not just how it works but why it's built the way it is, and what a teacher,
  a shelter worker, or a hackerspace organizer gets out of running it *(GOPOD layer — not
  built yet: no teacher, shelter worker, or organizer has run it for real people yet)*
- **Offline-first, dolly-portable deployment** — the whole stack runs on a Jetson with no
  internet required after setup, designed to travel (see `MOBILE_GEAR.md`)
- **The stack** — Wire-Pod, Ollama, Kokoro, Vosk, Python, and a Jetson host run every song in
  this repo
- **A private rehearsal copy for every song's tuning file** — each song's `knobs.json` (the
  timing, cycles, and cue values that get tuned by ear) is the public "latest confirmed"
  version you're reading in this repo. The moment-to-moment tinkering — trying a value,
  hearing it, trying another — happens in a gitignored `zKnobs.json` sibling that every
  runner prefers automatically when it's present. Nothing about mid-tune trial and error
  shows up as repo noise; only a value the operator has actually settled on gets promoted
  into the public file.

---

## Where GOPOD stands today

Confirmed live, verified this session and prior sessions:

- PTT → STT → LLM → robot speech, full loop
- All three custom intents, cold-restart persistent
- Interview Section 1, generating and playing back end-to-end, dual Kokoro voices streaming
- Bingo, single-robot calling confirmed, Brobot 2's reactor confirmed reacting, 3-second cadence and
  one-command dual-robot launch confirmed working together
- Bingo video song (`bingo-video-song-live`), the 46-step comedic capture piece, live-confirmed
  "mechanically perfect" end to end
- Cockpit running at `:8011`, live PTT/chat/Bingo rendering, Wire-Pod log prefill live

Honestly still in progress:

- Cockpit camera panes are cached/placeholder frames; Edge TPU inference is wired but optional
  and offline by default
- The public GitHub Pages cockpit teaser is a static shell — no live camera or session feed wired
  into it yet
- Public demo readiness: in progress, not yet declared done
- No hero photo or clip of both robots mid-session captured yet

---

## What's next

**Cozmo** and the **Moorebot Scout** are the next robot bodies the shared brobots
scaffolding is envisioned to host, beyond the two Vectors — scoped, not yet started. SDK
vendoring groundwork is already in the repo for both — no runtime code exists yet. The
operator's current lean, not a commitment or a scheduled build: Moorebot Scout looks
like the easier first pick as a GOPOD-layer chat participant. Bingo's "shared envelope
stream, zero coupling" pattern is the intended on-ramp: a new robot joins a session by
reading the same stream and picking which events to react to, not by patching Brobot 1 or
Brobot 2's code.

---

> From Doctrine Barfallonyou
> Lesson! This is what happens when you stop waiting for perfect and start building proof.
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
- [QUICKSTART.md](SINGLE_BOT_QUICKSTART.md) — talk to your own Vector, one robot, no alias studio needed
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

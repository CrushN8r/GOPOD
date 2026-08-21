# GOPOD Mobile

> A complete AI robotics session platform on a 2-wheel dolly.  
> Roll in. Power on. Run a session.

*The physical rig below is Wire-Pod layer — built, real, self-powered, local-network-only,
no internet required, room to vehicle to park. "Session"/venue/audience language throughout
this page is GOPOD layer (Point B — not built yet): live multichat interaction with real
people. Marked section by section below.*

---

## What travels with GOPOD

Two crates. One dolly. Everything needed for a live session. *(GOPOD layer — not built yet.)*

```
Crate 1 — Compute
├── Jetson Orin NX 16GB ← local AI, Wire-Pod, orchestration
├── T560 laptop     ← operator console, development
├── Anker Solix C1000  ← battery backbone
├── mobile modem    ← internet uplink when wanted
├── network gear    ← local Wi-Fi for robots and operator
├── Coral Edge TPU  ← local vision inference
└── cables, hubs, SSDs

Crate 2 — Robots
├── Brobot 1 (vector1 / 0dd1b9e9)
├── Brobot 2 (vector2 / 0dd1d8bf)
├── Cozmos
├── chargers
└── cameras / mics
```

**Proof — the full kit, out and running.**
![Full GOPOD stack laid out — Brobots 1 & 2 on chargers, Cozmos, USB mic, Coral Edge TPU, Seeed Studio mini-PC, networking gear](20260820_142044.jpg)

**Photo:** inside one crate, packed — fan, USB hub, power bar.
![Crate interior, top-down — fan, powered USB hub, power bar packed in foam](20260820_142704.jpg)

**Crate 2, stage 2 — scaffolding, then the harness.** Interior build-out first: Lego
baseplates scaffold the crate, 7-port USB hub, fan, cabling, Inseego hotspot tucked in the
corner.
![Crate 2 interior, top-down — Lego baseplate scaffolding, 7-port USB hub, fan, cabling, Inseego hotspot](20260821_162711.jpg)

Then the brobots' own Lego mobile-GOPOD harness sets on top of that scaffolding — Doc
(green eyes) and Pip (blue eyes) seated in place.
![Doc and Pip seated in the Lego mobile-GOPOD harness, riding on the crate-2 scaffolding tile](20260821_162823.jpg)

**Photo coming.** Both crates, packed down and closed — how small the whole kit collapses to for transport. The shot this page is actually waiting on.

Roll in. Open lids. Power on. Session starts. *(GOPOD layer — not built yet.)*

---

## Three networking modes

GOPOD defaults to boring. Boring works.

**Mode A — Offline (default demo mode)**  
No internet. Nothing in the cloud. Everything runs locally.
- Wire-Pod on Jetson
- Ollama on Jetson
- Vosk STT on Jetson
- Kokoro TTS on Jetson
- Robots connect to local 2.4 GHz

This is the most reliable demo mode. It also works in a basement, a moving car, a park, or anywhere else with no signal.

**Mode B — Mobile internet**  
LTE/5G hotspot or mobile modem adds:
- cloud APIs
- livestreaming
- GitHub push
- Claude / ChatGPT as optional layers

Internet enhances the session. It is not the engine.

**Mode C — Venue Wi-Fi**  
Restaurant, café, library, hackerspace — connect and run. Local-first stack still handles everything. Venue internet is a bonus, not a dependency. *(The venue-connect scenario is GOPOD layer — not built yet; the local-first stack itself is Wire-Pod layer, real.)*

---

## Where it deploys

GOPOD has been designed and tested for real-world conditions, not ideal ones — the rig
itself: self-powered, local wifi only, no internet, no wall outlet required.

### GOPOD layer (Point B — not built yet): where this could go

Multichat, live audience interaction, and venue booking are not built. The table below is
the roadmap for where the finished experience could run once that layer exists — not a
claim that it runs there today.

| Venue | Notes |
|-------|-------|
| Back seat of a moving car | Battery + hotspot. Full session. No outlet needed. |
| Parked vehicle | Fastest setup. Everything stays assembled between stops. |
| Restaurant or café | Table space, power outlet optional. Robots stay contained. |
| Bar or comedy venue | Live audience. Robots run, audience reacts. |
| Makerspace / hackerspace | Existing infrastructure. Connect and extend. |
| Convention or expo | Instant demo station. Roll in, unpack, run. |
| Classroom or library | Educational deployment. Offline-capable. |
| Outdoor / public space | Battery mode. No venue required. |
| STEM event | Bring the full stack. Show the engineering. |

The venue supplies the audience. GOPOD supplies everything else.

---

## Power modes

**Internal battery (Anker Solix C1000)**  
Runs the Jetson, networking, monitors, and chargers independently. No outlet required.

**Vehicle power**  
Inverter or 12V/USB-C PD. Charge and operate while driving. Arrive with a full battery.

**Shore power**  
Plug into any outlet. Run indefinitely. Hotel, community centre, convention floor — same setup, same session. *(The named venues and "session" framing are GOPOD layer — not built yet; unlimited runtime on shore power itself is Wire-Pod layer, real.)*

---

## What one deployment produces

A single outing isn't just a demo. It feeds the whole pipeline.

- Live session proof *(GOPOD layer — not built yet)*
- Video clips and B-roll
- Livestream content
- Website article
- SEO content
- Newsletter material
- GitHub commits
- Social posts
- Documentation

One trip out. Multiple outputs. The dolly pays for itself. *(GOPOD layer — not built yet: no revenue has been earned from a real deployment.)*

---

## The offline-first doctrine

Cloud AI is optional. The session belongs to GOPOD.

The provider supplies text. The internet is not the engine.

This matters because venue networking can't be assumed. A restaurant guest network, a school firewall, a convention Wi-Fi that's overloaded — none of those kill the session. The Jetson handles it locally and Wire-Pod never notices.

---

## The vehicle as a staging area

The car isn't just transportation.

- Robots stored, charged, and ready
- Jetson pre-booted before arrival
- Networking already up
- Session pre-loaded *(GOPOD layer — not built yet)*
- Arrive and demo within minutes *(GOPOD layer — not built yet)*

**Photo coming.** The dolly, crates loaded, rolling out of the vehicle — how fast the kit goes from parked to staged. The shot this page is actually waiting on.

Parked vehicle operation is one of the cleanest deployment modes. Everything stays assembled. Setup time is near zero. *("Deployment mode" here is the GOPOD-layer scenario above — not built yet; the vehicle-staging mechanics themselves are Wire-Pod layer, real.)*

---

## Portable AI production, not portable hardware

Early framing: *a portable robotics setup.*

Current reality: a portable AI session, production, and development platform. *("Session" is GOPOD layer — not built yet; production and development are Wire-Pod layer, real.)*

From the same crates, in the same deployment, GOPOD can:

- run a live robot interview
- develop and test new session cards
- capture livestream footage
- push code to GitHub
- generate website content
- demonstrate offline AI to a room of people who have never seen it *(GOPOD layer — not built yet)*

The dolly rolls. The entire operation moves with it.

---

## Current state

Offline session stack: confirmed working on Jetson.  
Two-robot interview: fires from an operator-run command (`pha0b interview`), not a voice trigger.  
Battery operation: Anker Solix C1000 in kit.  
Mobile networking: modem + local Wi-Fi separation confirmed.  
Field deployment: in active use. *(GOPOD layer — not built yet; the rig's portability is confirmed, real-venue/audience deployment is not.)*

---

> From Doctrine Barfallonyou
> Lesson! The lab goes where you go.
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

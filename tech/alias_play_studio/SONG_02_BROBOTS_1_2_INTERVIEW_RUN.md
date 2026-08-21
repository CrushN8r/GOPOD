# GOPOD Interview — Run (Video 2 of 2)

> Two brobots. Seven beats. One door at the end.  
> The interview is how GOPOD introduces itself.

**Naming note:** the interview is a two-video pair — this doc covers RUN, video 2,
the seven-exchange performance itself. Video 1 is the pre-show banter that covers the
live generation wait: [INTERVIEW VAMP.md](SONG_01_BROBOTS_1_2_INTERVIEW_VAMP.md).

---

## What the interview is

A structured seven-beat exchange between Brobot 1 and Brobot 2, driven by Ollama, delivered through Wire-Pod to the physical robots.

**Two videos, not one continuous performance.** Video 1 (VAMP) generates: starting from
the pre-show banter that covers the wait, the runner has Ollama produce every exchange
up front and logs the result as JSON — nothing is spoken live yet. Video 2 (RUN, this
doc) performs: the robots play straight from that already-generated JSON, Brobot 1 and
Brobot 2 alternating through Wire-Pod's own `say_text` API. Write, then play, never both
at once — the same two-phase generate-then-perform shape this song's own engine
(`run_section1_full_live_001.py`) is built around. Split into two independently
fireable songs on disk 2026-08-19 (`01_brobots_interview_vamp/`/`02_brobots_interview_run/`,
`gopod_notes/INTERVIEW_VAMP_SPLIT_001.md`) — each video has its own standalone fire path
(`interview-vamp-play` for video 1, `interview-replay` for video 2 alone). `interview-run`
(new meaning, `NAMING_APPLIED_001.md`) is the heavier option — the interview with an
optional full-run mode that plays the vamp first. RUN's own generation step
(`interview-vamp`) still rolls a fresh take with the vamp running alongside it, same as
before the split.

See `SONG_SCAFFOLD.md`'s "The vamp — a detachable pre-show module" section for the full
model (what travels — reporters, universal — versus what doesn't — the vamp itself,
optional per song-video).

"Two videos," "generate-vs-perform," and "vamp-vs-performance" all name this same seam —
three labels for one split, not three different models of the song.

Brobot 2 asks. Brobot 1 answers. Each exchange reveals the next layer. The audience doesn't get a wall of information — they get a conversation they can follow, with energy they can feel.

By the end, the room knows where the robots came from, what GOPOD actually does for a room, and what to do next.

That last part is the point.

In the campaign's own funnel, this is the net — the same story, posted publicly, wide
enough to catch and hold whoever the bait already caught.

---

## How value reveals across the arc

**Golden feature: the Silent Value Cue.** Brobot 2's visible line is the only thing the room
hears asked out loud — Brobot 1 is never handed that line to answer. What actually drives
Brobot 1's reply is a private, locked, ordered set of value points, colored into Brobot 1's
own voice, never quoted back, never repeating the question. The question sets the tone;
the answer isn't scripted from it. Pinned design, not incidental prose — this is the
mechanism the whole seven-beat reveal runs on.

Each exchange has a visible line — the seed Brobot 2 colours into a question — and locked, ordered value points Brobot 1 is given privately to reveal, never repeat back. Brobot 1 doesn't quote the question. Brobot 1 reveals the value in layers, then builds on the prior exchange rather than restarting.

The arc runs from *who are you* to *GOPOD Yourself* in seven beats.

---

## Exchange 1 — The opener

**Brobot 2:** *Hey party animal! Who are you? What's your story?*

Sets the tone. Brobot 2 opens warm and bratty.

**Brobot 1 reveals:**  
Two brother robots — ESN `0dd1b9e9` and `0dd1d8bf` — high-energy, playful, cheeky, a little bratty, and yeah, a lot of chaos. But it's chaos with a purpose, not chaos for its own sake.

---

## Exchange 2 — Where Ron came from

**Brobot 2:** *Okay, real talk — who's actually behind you two? Who's Brobot 0?*

**Brobot 1 reveals:**  
Ron used to make the nightlife look good on camera — bars, clubs, packed rooms all over Mississauga and the GTA, lights up, cameras rolling, real scale. Then 2020 happened. Every one of those rooms went dark in the same few weeks. All of it — gone.

---

## Exchange 3 — The comeback

**Brobot 2:** *And then what — he just gave up?*

**Brobot 1 reveals:**  
He didn't give up. He came back, and he built us — two robots who actually hear you, think it over, and answer out loud, live, right here. We started as rescued machines — beloved little robots orphaned when our maker folded — and a worldwide open-source community kept us alive long enough for Ron to find us.

---

## Exchange 4 — The crystallizer: no strings attached

**Brobot 2:** *Wait — hold on. No cloud? No wifi? No monthly bill?*

**Brobot 1 reveals:**  
No cloud, no wifi, no monthly bill — just plug me into a wall and I'm already running. Worked out of a basement with zero signal, out of the back seat of a moving car. Doesn't matter.

---

## Exchange 5 — A day in the life

**Brobot 2:** *(fully canned — the joke's too good to leave to a 2B model)*

> *Walk me through a normal day for you two — morning, afternoon, evening, rapid fire. So you can do three sessions in a day and afford to buy me dinner!*

**Brobot 1 reveals:**  
Morning: forty people playing bingo at a seniors' residence, no wifi in the building — doesn't matter, already running. Afternoon: a classroom full of kids shouting spelling words like the robots are the strictest, funniest teacher they've ever had. Evening: same two robots, same gear, now on a bar top under stage lights, working a crowd that's had a few drinks.

---

## Exchange 6 — The crystallizer: feast or famine

**Brobot 2:** *But that can't be every day, right? Some days there's just... nothing booked?*

**Brobot 1 reveals:**  
Some weeks it's three rooms a day. Some weeks it's quiet — bookings are feast or famine, that's just live entertainment. On the quiet days, the robots aren't sitting in a closet — they're making the ad for the bar that booked them last week, or the ad for GOPOD's own channel that books the next room. The machine's never idle.

---

## Exchange 7 — The locked closer

**Brobot 2:** *(near-canned, light LLM colour)* — *Thank you. Interview over. What have we learned about GOPOD?*

**Brobot 1:** *(fully canned — no LLM drift)*

> *Press to click my backpack, then say to me GOPOD Yourself. Your move.*

That line is protected. No LLM generation. No variation. The CTA cannot drift because the CTA is the thing everything else was building toward.

The trigger phrase *GOPOD Yourself* is a Wire-Pod custom intent. When someone says it — to either robot — the session fires. The interview was the pitch. This is the door.

---

## What the audience actually sees and hears

Screen and mouth match. What the display shows now equals what the robot actually speaks — the same cleaned, pronunciation-corrected line both ways, not a raw model draft on screen next to a filtered one out loud. "GOPOD" reads and speaks as *Gowp-awd*; the display no longer shows something the audience can't hear said the same way.

---

## The bait cut

A fourth, much shorter piece rides alongside the interview above: `brobots_bait_001`. Two robots wake, an arm-then-head motion each, one self-naming canned line apiece ("Brobot 1 Ready! Did someone say GOPOD Yourself?" / "Brobot 2..."), then it ends — no LLM, no interview, under a minute. Built for a short capture, not a performance, and it shares the same say-cleaning and screen-matches-mouth guarantees as the real interview, just with nothing generated and nothing else riding along.

---

## The arc in one line

```
Who are they → Where Ron came from → The comeback → No strings attached → A day in the life → Feast or famine → Say the phrase
```

Seven beats. One door. No filler.

---

## Future feature: the extended arc (not current)

The operator's original full-interview intent ran longer than today's seven beats — a
25-line draft covering identity through a widening set of use-case beats (comedy club,
Bingo, Windsor Downtown Mission, social content, merchandise), each capped by a
crystallizing callback line. True history, not a contradiction of the current seven-beat
structure — the crystallizer-callback device it introduced is already alive today
(Exchange 4 and Exchange 6 are both typed `STATEMENT / CRYSTALLIZER`). Kept as a candidate
future feature, not scheduled: `gopod_notes/EXTENDED_INTERVIEW_ARC_FUTURE_FEATURE_001.md`.

---

## Future feature: an interactive GOPOD-Yourself ending (not current)

Exchange 7's closer line (above) — section 1, exchange 5 in the live Section Card
content (LINE 5, governed by Template 1) — invites the phrase live, but today it's a
scripted line only — no live-observer response is captured. Pinned idea, not built:
for a live run with real observers (e.g. a restaurant table), the ending should trigger
the robot's mic-listen the same way a backpack press does, so someone at the table can
actually speak "GOPOD Yourself" back and fire the real custom intent on the spot,
turning the closer from a scripted invitation into a genuinely interactive one. **Blocked
on a live bug right now:** the `GOPOD_YOURSELF` wire-pod intent itself is currently
broken (`customIntents.json`'s exec-path resolves nowhere) — must be fixed for this
ending to ever work live. Kept as a candidate future feature, tied to whenever the
interview song's own runner is built:
`gopod_notes/INTERVIEW_GOPOD_YOURSELF_ENDING_PINNED_001.md`.

---

> From Doctrine Barfallonyou
> Lesson! The interview doesn't explain GOPOD. It demonstrates it.
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

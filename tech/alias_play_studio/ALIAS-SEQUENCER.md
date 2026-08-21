# ALIAS-SEQUENCER

> The PIANO has the notes. The MIXER sets how each one renders. The SEQUENCER is where
> they get arranged into something that plays. This doc is a seed, not a catalog yet —
> it exists to hold the concept and the first decision made about it, so the next
> session doesn't have to re-derive either.

---

## What this document is

Two docs, three concerns — PIANO and MIXER folded into `ALIAS-LIBRARY.md` 2026-07-16, both
files themselves deleted outright 2026-07-23 once genuinely zero-caller (see that doc's own
intro for the history), but the three concerns they named are still real and still separate:

- **The instrument (PIANO)** — now `ALIAS-LIBRARY.md`'s own Registry section. Catalogs
  individual notes — single aliases, each one self-contained, each one pressable on its own
  (`brobots-lift-up`, `brobots-head-nod`, `brobots-anim-celebrate`, and so on).
- **The board (MIXER)** — now `ALIAS-LIBRARY.md`'s own Render Controls section. How each
  note actually renders when played: async vs sync, hold/timing, temperature-style dials.
  Settings on a channel, not the order channels play in.
- **The arrangement (SEQUENCER, this doc)** — which notes, in what order, with what pauses,
  wired into GOPOD sequences and songs — using the Render Controls settings where relevant,
  but owning neither the notes nor their per-note render controls. This doc doesn't define
  new notes and doesn't set how a note renders; it arranges what `ALIAS-LIBRARY.md`'s other
  two sections already define.

The frame: **note → sequence → song.**

- A **note** lives in `ALIAS-LIBRARY.md`'s Registry — one alias, one action, atomic.
- A **sequence** is an ordered handful of notes played with intent (a lift-up, a
  head-nod, an anim-celebrate, in a chosen order, maybe with pauses between).
- A **song** is a sequence (or several) wired into an actual moment — a demo beat, a
  reaction, part of the interview.

Nothing in this doc changes what's in the Registry, and nothing here sets a note's render
controls (that's the Render Controls section's job). This doc only talks about how notes
get arranged, in what order, once they exist.

## The compose-then-play principle

The SEQUENCER inherits its basic shape from the interview runner's own established
split, already recorded in `ALIAS-LIBRARY.md`'s registry: `interview-json` (renamed
from `start-the-preshow` 2026-08-19, `NAMING_APPLIED_001.md`) composes — it generates
every exchange line up front and writes it to a JSON log — before `interview-replay`
(renamed from `interview-run` same day) actually plays it. Composition and performance
are two separate passes, not one.

The SEQUENCER is expected to work the same way: a sequence gets composed (the full
ordered list of notes decided) before it gets played (each note fired in order). This
doc doesn't build that machinery yet — it just states the shape a real SEQUENCER
implementation should follow when it exists.

## First recorded design decision (OPEN — not executed)

**Candidate direction:** `brobots-head-nod` is currently one baked note — a single alias
that internally fires head-down then head-up (`_brobots_move_axis "move_head" -2 "$hold"`
followed by `_brobots_move_axis "move_head" 2 "$hold"`, per `ALIAS-LIBRARY.md`'s own
description of it: "one small head-down-then-up sequence... packaged as a single
self-contained note"). That makes it the one registry note that is secretly two moves
wearing a single-alias coat.

The cleaner architecture, once a SEQUENCER exists to do the arranging: keep the registry
strictly atomic — single-direction notes only (a `head-down` note, a `head-up` note,
nothing that internally sequences) — and let the SEQUENCER compose the nod itself, as a
two-note sequence (head-down + head-up), the same way it would compose any other
sequence. The "silent yes" gesture would then be a SEQUENCER-level *sequence* built from
two registry-level *notes*, instead of one note that quietly contains a sequence. This is
a SEQUENCER-layer move, not a Render Controls one — it's about arrangement order, not
about how either half-note renders.

**Status: recorded as a candidate direction, not done.** No split has been made.
`brobots-head-nod` still exists exactly as it is in the registry today. Executing this
would mean adding two new atomic notes and retiring or re-deriving `brobots-head-nod`
from them — a future code change against `~/.gopod_alias_lib/brobots.sh`, out of scope
for this doc.

## How sequences will be recorded here

When a real sequence gets built, it should land in this doc as: a name, the ordered list
of registry notes it plays, which Render Controls settings apply to each, and the intent
behind the ordering (why this note before that one). Grow this section note by note, the
same way `ALIAS-LIBRARY.md`'s own registry grew alias by alias.

## First real entry — the pre-show vamp gate (`gopod-vamp`, 2026-07-16)

Registry polish pass (`ALIAS_REGISTRY_POLISH_001.md`). Before this, the vamp gate inside
`run_preshow_song()` (the scored `vamp_1..vamp_4` filler beats it loops while interview
generation is still running) had no registry note of its own — reachable only by running
the whole pre-show song and hoping generation was still in flight when the gate opened. Not
composed here as a *new* arrangement — `gopod-vamp` plays the exact same beats, in the
exact same order, through the exact same speak function (`_preshow_speak_host()`) the real
song's own vamp loop already calls. What's new is that this sequence is now reachable on
its own, outside the song that contains it. It also has its own standalone,
playhead-sliceable *song* — `brobots_vamp_gate`, via `pha0b vamp` — archived at
`goverlord/runtime/songs/zzz_archives/brobots_vamp_gate/` since 2026-07-24, doc moved
alongside it, off the grid on purpose.)

- **Notes played, in order:** `_preshow_speak_host(vamp_1)` → `_preshow_speak_host(vamp_2)`
  → `_preshow_speak_host(vamp_3)` → `_preshow_speak_host(vamp_4)`, repeating for as many
  cycles as the caller asks for (default 1 cycle = the four beats once).
- **Render settings applied:** see `ALIAS-LIBRARY.md`'s Render Controls section (folded in
  from the old `ALIAS-MIXER.md`'s 2026-07-16 addendum) — sequential, blocking Kokoro calls,
  each blocking on its own DONE ack before the next beat starts.
- **Intent behind the ordering:** the four beats are the real song's own authored
  vamp-fallback dialogue (two hosts trading lines while "backstage" finishes writing the
  interview) — the order is fixed by `knobs.json`'s own step list, not decided here.
- **Not a new sequence, a newly-exposed one:** this is the first case of a sequence that
  already existed inside a song gaining its own standalone PIANO-level door, rather than a
  sequence invented fresh for this doc. Worth naming as the template for the next gap like
  it, if one turns up.

## The opening chord as a chordal progression (`gopod-opening-chord`)

Not one note but four struck together, then held until every one lands — the reason it's
named a *chord* rather than a sequence. Notes: `gopod-mic-set`, an LLM warm ping, a Kokoro
voice warm-up, and the whole Wire-Pod chain (restart → poll ready → wake Brobot 1 → wake
Brobot 2, serial by Wire-Pod's own design, not a choice made here). Each robot speaks its
own ready line as it wakes; the chord holds — does not report `READY` — until every note has
truly landed, spoken lines included. Only then do both robots attempt "Brobots ready!"
simultaneously. Intent behind the shape: a demo shouldn't start on a guess that the stage is
set — every dependency the interview needs (mic, LLM, voice, both robots) is confirmed live
before the chord resolves. See ALIAS-LIBRARY.md's `core.sh` registry entry for the full
render detail (including the fire-and-forget "together" caveat) and
`CHORD_ABSORBS_PREDEMO_001.md` for its confirmed pre-demo-equivalent coverage.

## `wpr` as a two-note recovery sequence

`wpr` itself is a fixed two-note sequence (restart, then a 2-second settle buffer) — not a
single atomic note, though it's pressed as one key. When that's not enough to clear a stuck
connection, the proven recovery is a three-step sequence, not a single retry:
power-cycle both robots → re-pair with Wire-Pod → `wpr`. A partial version of this sequence
(just `wpr` again, or re-pairing without a power-cycle first) has been confirmed live *not*
to clear the error — the full three-step order matters. See `robot_control_song_001`'s own
story.md (Troubleshooting section) for the field record this sequence is drawn from.

## Speak-a-turn (the interview's per-exchange shape)

The interview's per-turn shape is itself a sequence, not a single note: assume control →
speak the line → (thinking window, if applicable) → release control, repeated once per
speaker per exchange. This is the shape every reaction-lane test below borrows from, and
the shape the golden paced dispatch (next entry) deliberately breaks from in one specific
way — by inserting a real pause between the spoken half and the animation half instead of
running them back to back.

## The paced reaction sequence (`test-reaction-in-the-beat`, 2026-07-16)

The most consequential sequence composed this session — arrived at empirically, across
many crashed attempts, before landing on **four consecutive clean runs including `angry`
twice**. Full detail and evidence in `ALIAS-LIBRARY.md`'s "Today's findings" section and
`FOUR_CLEAN_RUNS_ANGRY_FIXED_001.md`; recorded here in this doc's own grammar for how the
notes are arranged:

- **Notes played, in order:** wake both (Brobot 1 `conn_test`, Brobot 2 `conn_test`, 1.5s
  settle) → Brobot 1 speaks "Animation test run" (hold_phrase 3.0s, release, 1s settle) →
  Brobot 2 speaks its emotion phrase, e.g. "I'm angry" → **a real 2-second pause, no note
  fires** → Brobot 2 fires a bare `playAnimationWI` animation dispatch (hold_anim 5.0s,
  release, 1s settle) → Brobot 1 speaks "Run Complete" (hold_phrase, release).
- **Render settings applied:** see ALIAS-LIBRARY.md's Render Controls section — two separate
  `say_text` calls rather than one embedded sentence, the async `playAnimationWI` token, a
  5.0s animation hold (the operator's own proven value).
- **Intent behind the ordering:** the 2-second silent pause between "speak the emotion" and
  "fire the animation" is the one deliberate structural break from the speak-a-turn shape
  above — giving the robot's own action queue a real beat to finish one thing before the
  next starts, which is the most defensible read of why this sequence stopped crashing where
  every tighter-paced version before it did not.
- **Composed via `test-reaction-pick-animation`:** the picker reads the verified token list
  live from `animation_vocab.json`, presents a numbered menu, and hands the chosen token into
  this sequence — composition (picking the token) and performance (playing the sequence) are
  the same two-pass split this doc's compose-then-play principle already names.

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

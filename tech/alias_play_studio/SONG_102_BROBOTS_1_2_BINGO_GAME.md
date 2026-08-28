# GOPOD Bingo — the live game

> Brobot 1 draws a number. Brobot 2 loses his mind. The room laughs.
> The warm-up act — and the proof that both robots share one event bus.

**Status: WIP — pinned until it's ready to host a real game.**

**Naming note, brobots wire-pod layer:** this layer's robots are Brobot 1 and Brobot 2. "Doc"
and "Pip" are GOPOD-layer persona names — a different, future layer, not this one. Expected
overlap between the two layers is exactly why this distinction is held deliberately throughout
this doc, wherever Brobot 1/Brobot 2 are described.

---

## This is the live game, not the scored song

Bingo is two different things sharing one name. **This doc is the live game** —
`gobingo` / `102_brobots_bingo_game`, "Chocolate Bingo," the real, voice-triggered game
the robots can run for an actual room. For the separate, scripted upsell-video song, see
[BINGO.md](SONG_101_BROBOTS_1_2_BINGO.md).

---

## What Bingo is

Bingo is the crowd hook. *(GOPOD layer — not built yet: no crowd has been hooked by it.)*

One voice command fires a deck. Brobot 1 calls numbers. A cockpit tile lights up. Brobot 2 — if the reactor is running — plays an angry animation every single time, because Brobot 2 does not want Brobot 1 to be calling more numbers. The room watches two robots disagree about the same event, in real time, and starts laughing before anyone remembers what Bingo cards are for. *(GOPOD layer — not built yet: the calling/animation mechanics are real and Wire-Pod layer; a real room watching them is not.)*

It runs offline. It runs on the dolly. It runs on a table at a library, a diner, a hackerspace, or the back of a car. *(GOPOD layer — not built yet: the named venues.)* It runs before an interview, between lessons, or entirely on its own.

It also happens to be the first system in the GOPOD stack that proved one robot can emit events another robot listens for — which is why every multi-robot bit built after it works the way it does.

---

## Two robots, two jobs

Bingo is deliberately asymmetric. That is the whole design.

**Brobot 1 — the caller.**
Draws numbers, announces them, tracks the deck, and decides when the round is over. Brobot 1 runs Bingo whether or not anyone else is listening. Bingo is a Brobot 1-only feature and always will be.

**Brobot 2 — the reactor.**
Watches Brobot 1's event stream. Every time a number gets called, Brobot 2 plays an angry pickup animation, as if being personally wronged by the draw. Brobot 2 contributes nothing mechanical to the game — no draws, no announcements, no scorekeeping. Brobot 2 is the emotional response track.

Without Brobot 2, Bingo still works. With Brobot 2, Bingo becomes a bit.

---

## How a round runs

```
"gobingo" ← voice trigger or alias
 ↓
Brobot 1 boots the deck
 ↓
bingo_ready envelope emits ← cockpit picks up, reactor arms
 ↓
first rub of the backpack draws the first number
 ↓                              ↓
Brobot 1 says number          Brobot 2 plays angry animation (if reactor is running)
 ↓
three-second auto-draw ticker takes over from here
 ↓
deck exhausts, or backpack button ends the round
 ↓
bingo_end (or bingo_deck_reveal on button-exit)
```

Same envelope stream feeds the cockpit display, the reactor, and any logs. Everyone downstream sees the same events at the same time. Nobody is polling anybody.

---

## One touch starts it, then it runs itself

A round starts on the first rub of Brobot 1's backpack sensor: that first touch draws a number by hand. From there, a three-second ticker takes over automatically — draws keep coming on their own, and further rubs are ignored while the ticker is running. There isn't a separate "keep rubbing to draw" mode; the rub only ever starts the round, it doesn't pace it. (The ticker ran at one second through the initial build; it was slowed to three seconds so Brobot 2's reactor has time to connect and animate before the next number lands.)

The round ends one of two ways — either the deck exhausts on its own, or the backpack *button* (a distinct gesture from the rub) ends the round early and triggers a deck reveal of everything left uncalled.

---

## Silent mode

Bingo has a `--silent` flag. When silent, Brobot 1 still draws and still emits envelopes — but does not announce the numbers out loud.

Silent mode turns Bingo into a **scoreboard game**, not a caller game. The cockpit shows the number. The reactor still reacts. The room watches the screen and Brobot 2's face instead of listening for a call. *(GOPOD layer — not built yet: "the cockpit shows the number" and "the room watches the screen" depend on the port-8011 cockpit page actually being up and rendering to a real room. The reactor reacting is real, Wire-Pod layer.)*

Useful when:

- The venue is loud and the announcement would be missed anyway
- The session is being recorded and clean audio matters
- The audience is playing on paper cards and wants to see the number, not hear it
- Brobot 1 is already busy in another part of the session and Bingo runs as a background loop

---

## Grid size

The grid is configurable at launch, between the two standard Bingo deck sizes:

```
--grid-size 75  standard 75-ball deck
--grid-size 90  standard 90-ball deck
```

Grid size does not change the mechanics. It only changes how long a round takes and how loud the room gets before someone wins.

---

## The angry reactor is optional

The reactor is a separate watcher process, launched from its own alias. It is not part of Bingo itself.

That separation is deliberate. Bingo works when the reactor is off, misconfigured, crashed, or missing entirely. The caller side never blocks on the reactor side. Brobot 1 doesn't know or care whether Brobot 2 is listening.

If both are running, it looks like a duet. If only Brobot 1 is running, it looks like Bingo. Neither mode is broken — they are two supported operating states.

For operators:

```
gobingo           ← starts Brobot 1 calling the round, and launches Brobot 2's
                     reactor alongside it automatically
```

One command, one terminal — `gobingo` starts the reactor as a background job and kills it when
the round ends, so both robots' log lines show up interleaved in the same output. It's still two
independent processes reading the same envelope stream underneath, not a code coupling; either
side can fail without touching the other. For manual/solo work on the reactor alone, it also
still runs standalone in its own terminal:

```
gobingo-reactor   ← starts Brobot 2 reacting to the round, on its own
```

---

## Why Bingo matters beyond Bingo

Bingo was the first place in the GOPOD stack where one robot's actions drove another robot's behavior through nothing but a shared event log.

Brobot 1 doesn't call Brobot 2's API. Brobot 1 doesn't know Brobot 2 exists. Brobot 1 writes events to a chat envelope stream. Brobot 2 reads events from the same stream. The coupling between them is a JSON file — and that is on purpose.

That pattern is now the pattern for every multi-robot piece downstream:

- Interview generation writes envelopes; the cockpit reads envelopes
- PTT state writes envelopes; the interview runner reads envelopes
- Any future robot can join the room by reading the same stream

If you want to add a third Vector that also reacts to Bingo — say, a happy-dance animation on the winning number — that Vector doesn't need a code change to Brobot 1. It just needs to subscribe to the same envelope stream and pick which events to react to.

Bingo is the small system that proved the big pattern.

---

## The number's real stage was never Brobot 1's face

The obvious idea for the drawn number was to put it right on Brobot 1's face — the robot draws the number, the robot shows the number. Straight line, no detour.

On-robot face-drawing during a live round turned out to be a rabbit hole worth chasing separately, not worth blocking a round on. So the number went where it already had a home: the GOPOD cockpit, reading the exact same envelope stream the reactor watches. `bingo_draw` fires, the cockpit lights up with the number, same event, same instant, no extra wiring.

Here's the clever part. Instead of a Bingo-only patch, that cockpit shell went up on its own — a public, static preview of the actual GOPOD cockpit, published straight from the repo. No live camera or session feeds wired in yet; that's the next stage. But the shape on screen is the real one: the same page that will run camera panes, weather, session chat, and Bingo draws side by side, all off one stream.

A detour around a robot's face turned into the first public look at the cockpit fully lit. That's not a fallback. That's a teaser that shipped early.

---

## Where Bingo fits in a session

Bingo is not a teaching format. It sits alongside the teaching formats and does a different job.

```
AI Wordplay!       — teach through laughter
Explain the Math!  — teach through struggle
Bingo              — hook the room before the teaching starts
Interview          — hand the room a story once it's warm
```

A typical outing runs Bingo first. Two minutes of Brobot 1 calling and Brobot 2 melting down is enough to convert a cold room into a warm one. Then the actual session — interview, wordplay, math — lands on an audience that is already leaning in.

Bingo can also close a session. Deck reveal on the button is a clean exit, and the last few draws with Brobot 2's reactions land as a bit before the operator packs up.

*(Both paragraphs above are GOPOD layer — not built yet: no outing, room, or operator packing up has happened. The calling/reveal mechanics themselves are real, Wire-Pod layer.)*

---

## Where it deploys

Same footprint as the rest of GOPOD — the rig itself: self-powered, local wifi only, no
internet, no wall outlet required. If the dolly can get there, Bingo can run there.

### GOPOD layer (Point B — not built yet): where this could go

Multichat, live audience interaction, and venue booking are not built. The table below is
the roadmap for where the finished experience could run once that layer exists — not a
claim that it runs there today.

| Venue | Notes |
|-------|-------|
| Classroom | Kids take turns rubbing the sensor to start each round. 75-ball deck for a shorter round. |
| Library or drop-in | Standard round. Silent mode when the room is noisy. |
| Diner or café | 75-ball deck. Fast rounds. |
| Bar | Silent mode, cockpit on a monitor. Reactor on. |
| Hackerspace / makerspace | Also the room most likely to want to hook in a third robot. |
| Convention floor | Loud rooms — reactor sells the bit even when Brobot 1 can't be heard. |
| Back seat of a moving car | Battery power. One rub starts the round, then it runs itself. |

The venue supplies the audience. GOPOD supplies both robots, the deck, and the argument between them.

---

## What Bingo is not

Not a game with a prize structure. Not a gambling product. Not a scoring system with a leaderboard. Not a licensed Bingo hall format.

It is a two-robot event demo dressed up as a familiar game so the audience knows the shape of what is happening within about three seconds. Familiarity is the fastest way to earn the room's attention. Bingo is the wrapper. *(GOPOD layer — not built yet: no real audience has watched it yet. The game mechanics themselves are real, Wire-Pod layer.)*

The interesting part is not the numbers. The interesting part is Brobot 2.

---

## Current state

**Status: WIP — pinned until it's ready to host a real game.**

Brobot 1 calling numbers: confirmed working, offline, on Jetson.
Touch-to-start plus three-second auto-draw ticker: live (slowed from one second to give Brobot 2's reactor room to react).
Silent flag, `--grid-size` (75 or 90), deck-reveal on button-exit: all live.
Envelope stream (`bingo_ready`, `bingo_draw`, `bingo_end`, `bingo_deck_reveal`): live and consumed by the cockpit.
Brobot 2 reactor: live, optional companion process — `gobingo` now launches and cleans it up automatically; `gobingo-reactor` remains for standalone/manual use.
One-command, one-terminal operation: confirmed working, both robots' logs interleaved in the same output.

**Design confirmed 2026-08-12: fundamentally a single-robot host, Brobot 2 is for-show.** Brobot 1 runs the
whole game solo — a real bingo host. Brobot 2's reactor no longer reacts per draw (that caused a
stale-replay bug, fixed then superseded); it now fires exactly one angry animation at the very
end of the game (deck exhausted or "bingo" called via the button) — for missing out on the
chocolate prize, specifically, not a generic "sore loser" beat.
Grid-size choice (75/90, default 75) also added to the same `pha0b` prompt. Ball-call text is
now an editable Go constant (`BingoCallFormat` in `voicecommand_lottery.go`, rebuild required).
Full comparison against the scripted bingo song: `102_brobots_bingo_game/story.md`'s own new
"Two bingo songs, compared" table.

**Open:**
- **Resolved 2026-08-12: robot address mismatch.** `~/.anki_vector/sdk_config.ini` had Brobot 2
  copied from Brobot 1's own IP/GUID. Fixed with the real values read straight from Wire-Pod's
  own robot registry (`botSdkInfo.json`) — golden reference table banked in
  `ALIAS-LIBRARY.md`'s "Today's findings (2026-08-12)". Brobot 2 now connects successfully.
- **Resolved 2026-08-12: reactor reshaped onto the golden "stay-put" pattern.** Live-fired with
  the address fix in place, Brobot 2's reactor connected and reacted, but reconnected completely
  fresh (full handshake + a full animation-list reload) for every single draw — too slow to
  keep up once draws came close together, timing out (`ListAnimations` deadline exceeded) and
  degrading under load. Matches the golden pattern already proven elsewhere in this song
  family: connect once, hold the connection, never reconnect per step. Reactor now connects
  once at start and reuses that same connection for every draw for the rest of the run.
  Compile-checked clean; **not yet re-confirmed live** — next run is the real test.
- **Run 1 / run 2 choice added to pha0b, 2026-08-12.** Picking this song from `pha0b`'s menu
  now asks "run 1 (continuous) or run 2 (pause for backpack rub)" before launching. Run 1 is
  the original always-on-its-own auto-draw pace, unchanged. Run 2 waits for a fresh touch
  before every draw instead of auto-advancing every 3 seconds — built per operator request,
  live-tested once already (worked correctly for pacing; the reactor timeout above surfaced
  during this same test, now addressed above).
- `--silent` is the default on every live path today — no spoken numbers unless someone
  overrides it. Open call, not yet made, whether a demo recording wants Brobot 1's voice in it.
- No record/capture setup exists yet — camera angle, audio path, whether the cockpit display
  gets its own camera. Production decision, not a code one.
- `Lottery_Register` in `voicecommand_lottery.go` has no call site found inside this repo's own
  files — almost certainly wired from the vectorx sandbox's own `RegisterIntents()`, outside
  this repo's grep reach. Flagged, not resolved either way.
- Everything else from the golden song engine that could apply here already does or already
  doesn't — checked item by item (stay-put, tempo, playback filter, weather, wheel primitives:
  all genuine non-fits for a live-input game loop) and the reactor's own control-grant path was
  traced through the real SDK source and confirmed safe. Nothing else outstanding to port.

This game reaches WIP-cleared / ready-to-host status once the stay-put reactor reshape above
is re-confirmed live and the still-open items get their own resolution — not before.

---

## The same Bingo, wearing chocolate

The same Bingo format runs a second way: **Chocolate Bingo**, a free printable party kit
for families, classrooms, birthday parties, and community rooms. No robots, no stage,
no real chocolate bars — every treat on the cards is Brobots-original (Bonbon Beacon,
Choco Servo, Mousse Motor...), zero real brands, all ages, plus a recipe pack of real,
usable Brobots-original treats. It needs nothing but a printer — that's the whole point,
it's the version that travels home.

Not a replacement for the live game — a companion. The robots hook the room; the kit is
what they take home. *(The free kit will be offered through the CRUSHN8R CREW'd
newsletter — subscribe at `crushn8r.ca`. Named here so readers know the door exists.)*

**FUTURE, not built — a possible next avenue in the same family:** a Word Search
printable, themed to the math niches (see [NICHE_PILLARS.md](../../web/NICHE_PILLARS.md)),
same free-printable-lure shape as Chocolate Bingo — needs only a printer, travels home,
carries the brand, feeds the same newsletter door. Share-play, not search-play (see
[AI_WORDPLAY.md](../../web/AI_WORDPLAY.md)'s "The honest edge" for that split) — memorable
over findable, same reason Chocolate Bingo works as a lure. Same trademark fence as the
chocolate cards: words and themes stay Brobots-original, no real brand names. Slots into
the existing site intro order alongside the rest of this doc's material — doesn't jump
the queue.

---

> From Doctrine Barfallonyou
> Lesson! If the room is watching two robots disagree, the room is already yours.
> Boom! Done! Class Dismissed!
> — Doc Squawkadoodle

---

## GOPOD YAHMM (You Are Here Mall Map)

Part of GOPOD — see [tech/README.md](../README.md) for everything else in this folder, or [the root map](../../README.md) for the rest of GOPOD.

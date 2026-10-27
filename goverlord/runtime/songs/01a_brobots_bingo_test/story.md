# Brobots Bingo Test

**Status:** GOLDEN (2026-10-01)

The Bingo test song: two Brobots trade banter, then call three Bingo balls. Ready to record.
Play it with `start bingo test` in the Brobots chat, or with
`pha0b`. `stop run` stops it before its next step.

- **The lines** live in this file (`## STEP` / `> TEXT:`). **The steps, their order and
  timing** live in [knobs.json](knobs.json). The two must name the same steps.
- **Not the live Bingo game:** that's `01b_brobots_bingo_game/`.
- **Reporter gaps** (`reporter_gap_*`) are 0-second edit marks between exchanges, so a video
  editor knows where the reporters' voiceover goes.
- **If a step fails:** a failed animation stops the run at that step and says so; a failed
  arm cue, nod or rattle prints and the run carries on.

# 🎱 The score — sections & exchanges

## Song framing

- **Title:** Brobots Bingo Test
- **Purpose:** a scored, straight-line capture run for the Bingo prototype upsell video
- **Audience takeaway:** *(not yet authored)*
- **Song success:** *(not yet authored)*

## Personas

| | Brobot | Steps | Role |
|---|---|---|---|
| 🎤 **Host** | Brobot 1 | `host_` | runs the Bingo test, calls the balls |
| 😜 **Brat** | Brobot 2 | `brat_` | interrupts, answers every call with a joke |

## Score map

| Section | Part | Type | Speaker → Responder |
|---|---|---|---|
| OPENING SYNC | 🎬 Stage set | Opener | 🤖🤖 both, then 🎤 Host solo |
| INTERLEAVED BANTER | 🔁 **E1** Lucky number? | Q&A | 😜 Brat → 🎤 Host |
| | 🔁 **E2** "I, Won!" | Q&A | 😜 Brat → 🎤 Host |
| ROUND 1 | 🎬 Round open | — | 🎤 Host solo |
| | 🔁 **E3** B-1 | Ball call | 🎤 Host → 😜 Brat |
| ROUND 2 | 🎬 Round open ↩️ E3 | — | 🎤 Host solo |
| | 🔁 **E4** I-9 | Ball call | 🎤 Host → 😜 Brat |
| ROUND 3 | 🎬 Round open ↩️ E4 | — | 🎤 Host solo |
| | 🔁 **E5** O-0 | Ball call | 🎤 Host → 😜 Brat |
| CLOSING | 🔁 **E6** I'm the manager ↩️ E5 | Closer | 🎤 Host → 😜 Brat → 🎤 Host |

## How to read it

- **Section** = a block of the song (from `knobs.json`'s `section`). Each one holds a
  🎬 **stage set / round open** (mechanics and solo lines, no responder) and/or
  🔁 **exchanges**.
- **Exchange card** — the interview card's fields, the ones a scripted song uses:
  - **Type** — Opener · Q&A · Ball call · Closer (Ball call is Bingo's own).
  - **Speaker → Responder** — who starts, who answers.
  - **Arc point** — one plain line: what this exchange does in the story.
  - **Moves** — the scored animations and gestures, in order (every say line also gets
    Wire-Pod's default `thinking`).
  - Not used here: value points, thinking window (those are for LLM-generated lines).
- **One move or one line per step, in the order they happen:** a move step
  (arm cue, nod, `kgSuccess`, `searching`, angry beat, rattle) has no `> TEXT:`; each line is
  its own `*_say_*` step, and a line's name can repeat (`brat_say_0203` twice). A move
  followed straight away by the same Brobot's line plays with it, the native Wire-Pod way:
  the arm cue or nod while it's said, `kgSuccess` or the beat as the line starts, `searching`
  first, then the line. A line before a rattle is said, then the rattle plays under a big
  ⏳ countdown in the Brobots window (`"compose_moves": true` in `knobs.json`).
- 🔊 **Where the rattle plays** — `"rattle_on"` in `knobs.json`: `"speaker"` (this song)
  plays `rattle.wav` on the computer's speaker while the Brobots stay
  put, and the countdown is just the sound's length; `"robot"` plays it on Brobot 1 (the
  hand-off, 3 s settle, then the sound). Change it there, no code change.
- 🧠 **LLM warm-up** — `"llm_warmup": true` in `knobs.json`: a live run opens with the
  "LLM Warming Up" bubble (the LLM power level and hourglass) in the Brobots window for at
  least 5 s, or until the chat's model is loaded if that takes longer, then "LLM Ready",
  which stays in the window (`llm.*` in the settings).
- 🎱 **Big Bingo ball** — a ball-call step with `"show_ball": "B-1"` in `knobs.json` puts the
  called number in the Brobots window as a big ball (coloured by its letter) as the step
  starts; it stays in the window (only Clear Chat removes it).
- **Adding, removing or renaming a step** needs the same step, in the same order, in
  `knobs.json` and `zKnobs.json`; changing a line's words here is enough on its own.
- ↩️ **Callback** — the next section's first line answers the last joke before it.
- ⏸ `reporter_gap_*` steps sit between lines for the edit; cards don't list them.
- Lines are said with `sentence_split: "scripted"`: each line is split at . ! ? and
  commas (not inside quotes), said piece by piece, and the chat bubble lights each piece.
- Step order is untouched by this layer — it's labels only.

---

## SECTION: OPENING SYNC

> ### 🎬 Stage set · Opener
> - **Speaker:** 🤖🤖 both, then 🎤 Host solo
> - **Arc point:** both brobots wake up ready to Bingo; Host starts the test — Round 1, juggling big shiny bingo balls.
> - **Moves:** 🤖🤖 wake · ⏳ warm-up rattle · "Brobots ready" together · 🎤 arm cue · 🎱 rattle, after "Bingo Test Begins. Round 1! … Please wait." · ⏳ countdown

## STEP reporter_gap_0101

## STEP brobots_wake_0101
## STEP brobots_ready_0101
> TEXT: Brobots ready to Bingo!

## STEP reporter_gap_0102

## STEP host_arm_cue_0102
## STEP host_say_0102
> TEXT: Bingo Test Begins. Round 1! Juggling big shiny bingo balls! Please wait.
## STEP host_rattle_0102

## STEP reporter_gap_0103

---

## SECTION: INTERLEAVED BANTER

> ### 🔁 E1 · Lucky number? · Q&A
> - **Speaker → Responder:** 😜 Brat → 🎤 Host
> - **Arc point:** Brat stops the test to ask about his lucky number; Host searches, pushes back, then asks it anyway.
> - **Moves:** 😜 arm cue, with the line · 🎤 `searching`, then the line

## STEP reporter_gap_0201

## STEP brat_arm_cue_0201
## STEP brat_say_0201
> TEXT: Wait! I have a lucky Bingo number. Want to know what it is?

## STEP reporter_gap_0202

## STEP host_searching_animation_0202
## STEP host_say_0202
> TEXT: Excuse me? We're running a test! What is your lucky Bingo ball number?

## STEP reporter_gap_0203

> ### 🔁 E2 · "I, Won!" · Q&A
> - **Speaker → Responder:** 😜 Brat → 🎤 Host
> - **Arc point:** Brat's punchline, "I Won"! — then "Lol!"; Host searches, calls him Mr Game-Bro Funny Pants, and gets serious.
> - **Moves:** 😜 `kgSuccess`, with the line · 🎤 `searching`, then the line

## STEP brat_say_0203
> TEXT: It's. Wait for it, "I Won"!
## STEP brat_kg_success_0203
## STEP brat_say_0203
> TEXT: Lol!

## STEP host_searching_animation_0203
## STEP host_say_0203
> TEXT: Jiggling Circuits Mr Game-Bro Funny Pants! In all my nano-seconds of juggling bingo balls, I never heard that one before. Now, let's get serious!! 

## STEP reporter_gap_0204

---

## SECTION: ROUND 1

> ### 🎬 Round open
> - **Speaker:** 🎤 Host solo
> - **Arc point:** Host starts Round 1 again.
> - **Moves:** 🎱 rattle, after "Bingo Test Begins. Round 1! … Please wait." · ⏳ countdown

## STEP reporter_gap_0301

## STEP host_say_0301
> TEXT: Bingo Test Begins. Round 1! Juggling big shiny bingo balls! Please wait.
## STEP host_rattle_0301

## STEP reporter_gap_0302

> ### 🔁 E3 · B-1 · Ball call
> - **Speaker → Responder:** 🎤 Host → 😜 Brat
> - **Arc point:** Host calls B-1; Brat makes it about himself — "I be, the one!" — then "Lol!"
> - **Moves:** 🔵 big ball **B-1** on screen · 🎤 `kgSuccess`, with the line · 😜 `searching` → `kgSuccess`, with the line

## STEP host_kg_success_0302
## STEP host_say_0302
> TEXT: Bingo ball captured. B-1!

## STEP reporter_gap_0303

## STEP brat_searching_animation_0303
## STEP brat_say_0303
> TEXT: B-1? That's true! I be, the one!
## STEP brat_kg_success_0303
## STEP brat_say_0303
> TEXT: Lol!

## STEP reporter_gap_0304

---

## SECTION: ROUND 2

> ### 🎬 Round open · ↩️ E3
> - **Speaker:** 🎤 Host solo
> - **Arc point:** Host searches, plays along — "Well, you certainly are... 'one'." — and moves on to Round 2.
> - **Moves:** 🎤 `searching`, then the line · 🎱 rattle, after "Bingo Test Continues. Round 2! … Please wait." · ⏳ countdown

## STEP reporter_gap_0401

## STEP host_searching_animation_0401
## STEP host_say_0401
> TEXT: Well you certainly are...
## STEP host_kg_success_0401
## STEP host_say_0401
> TEXT: 'one'. Ugh.
## STEP host_say_0401
> TEXT: Bingo Test Continues. Round 2! Juggling big shiny bingo balls. Please wait.
## STEP host_rattle_0401

## STEP reporter_gap_0402

> ### 🔁 E4 · I-9 · Ball call
> - **Speaker → Responder:** 🎤 Host → 😜 Brat
> - **Arc point:** Host calls I-9; Brat says I-9 isn't in the I column and asks if 9 is Host's maturity level — then "Lol!"
> - **Moves:** 🎤 `kgSuccess` · 🔴 big ball **I-9** on screen, with "Bingo ball captured. I-9" · 😜 `searching`, then "Wait!" → `kgSuccess`, with the line

## STEP host_kg_success_0402
## STEP host_say_0402
> TEXT: Bingo ball captured. I-9!

## STEP reporter_gap_0403

## STEP brat_say_0403
> TEXT: Wait!
## STEP brat_searching_animation_0403
## STEP brat_say_0403
> TEXT: I-9? In 75-ball bingo, the "I" column, is I-16, to I-30. Are you telling me 9 is your maturity level?
## STEP brat_kg_success_0403
## STEP brat_say_0403
> TEXT: Lol!

## STEP reporter_gap_0404

---

## SECTION: ROUND 3

> ### 🎬 Round open · ↩️ E4
> - **Speaker:** 🎤 Host solo
> - **Arc point:** Host searches, goes angry — "Excuse Me? This Bingo Test must CONTINUE!!!" — and pushes on to Round 3.
> - **Moves:** 🎤 `searching` → 😠 `angry`, with "Excuse Me?" · then "This Bingo Test must CONTINUE!!!" once "Excuse Me?" is said · 🎱 rattle · ⏳ countdown

## STEP reporter_gap_0501

## STEP host_searching_animation_0501
## STEP host_beat_angry_0501
## STEP host_say_0501
> TEXT: Excuse Me? This Bingo Test must CONTINUE!!!

## STEP reporter_gap_0502

> ### 🔁 E5 · O-0 · Ball call
> - **Speaker → Responder:** 🎤 Host → 😜 Brat
> - **Arc point:** Host calls O-0; Brat checks it, then goes angry — the election is rigged, he wants the manager.
> - **Moves:** 🎤 `kgSuccess`, with "Round 3! Juggling big shiny bingo balls. Please wait." · 🟠 big ball **O-0** on screen, with "Bingo ball captured. O-0" · 😜 "Wait! O-0?" → `searching`, then "Verifying." → 😠 `angry`, with the line

## STEP host_say_0502
> TEXT: Round 3! Juggling big shiny bingo balls. Please wait.
## STEP host_rattle_0502
## STEP host_kg_success_0502
## STEP host_say_0502
> TEXT: Bingo ball captured. O-0

## STEP reporter_gap_0503

## STEP brat_say_0503
> TEXT: Wait! O-0?
## STEP brat_searching_animation_0503
## STEP brat_say_0503
> TEXT: Verifying.
## STEP brat_beat_angry_0503
## STEP brat_say_0503
> TEXT: O-0? This election is rigged! I wanna speak, to the manager!

## STEP reporter_gap_0504

---

## SECTION: CLOSING

> ### 🔁 E6 · I'm the manager · Closer · ↩️ E5
> - **Speaker → Responder:** 🎤 Host → 😜 Brat → 🎤 Host
> - **Arc point:** Host is the manager and ends the test — "Lol!"; Brat searches — "Hold, on a 'Wait-a-nano-second! Bro!'" — then "What the flux-capacitor, man?"; Host: "Wait what? Now be a good Brobot and GOPOD Yourself! …"
> - **Moves:** 🎤 arm cue · nod, with the line · `kgSuccess`, with "Lol!" · 😜 `searching`, then the line · arm cue, with "What the flux-capacitor, man?"

## STEP reporter_gap_0601

## STEP host_arm_cue_0601
## STEP host_nod_0601
## STEP host_say_0601
> TEXT: I'm the manager and I take all concerns seriously. However the Bingo Test is now over, Mister Big Shiny Bingo Balls!!
## STEP host_kg_success_0601
## STEP host_say_0601
> TEXT: Lol!

## STEP brat_searching_animation_0602
## STEP brat_say_0602
> TEXT: Hold, on a 'Wait-a-nano-second! Bro!'
## STEP brat_arm_cue_0602
## STEP brat_say_0602
> TEXT: What the flux-capacitor, man?

## STEP host_say_0602
> TEXT: Wait what? Now be a good Brobot and GOPOD Yourself! The rest of you can GOPOD yourselves by visiting crushn8r.ca and subscribing to the CRUSHN8R CREW'd newsletter 

## STEP reporter_gap_0603
## STEP exit

## Running

Run via the `bingo-test` alias (dry) or `bingo-test-live` (live, no env var
needed). Both robots run within one continuous invocation — this song is a real
back-and-forth, not the same script played twice.

To stop a run: type `stop run` in the Brobots chat - it stops at the next safe spot, before
its next step (force-stopped if it hasn't ended in 45 s). `kill-bingo-test`
in any terminal stops it at once. Touch interrupt is off for the run (`"touch_interrupt_off": true`
in `knobs.json`) and back on when it ends.

# The gears: the run chain

Part of [ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md) - [Concepts and terms](01_CONCEPTS.md) · [Songs](02_SONGS.md) · [The session and runs](03_SESSION.md) · [The Brobots chat](04_CHAT.md) · [Voices and sound](05_VOICES_AND_SOUND.md) · [The cast and the game](06_CAST_AND_GAME.md) · [The system menu, your stuff, the settings](07_SYSTEM.md) · [Golden rules, troubleshooting, tests](08_RULES_AND_TESTS.md) · [The gears](09_GEARS.md)

## 20. The gears: GOPOD's run chain

GOPOD's game engine is built like a geared bicycle. The gears are fixed parts. The chain binds
them, and the derailleurs shift the chain between gears live. Once a gear is locked, only its
teeth (the values) are tuned. The gear stays the gear.

Every run goes through the same chain. This page is the map: each part, where it lives, and
whether it's locked yet.

### 20.1 The parts

| Bicycle | GOPOD | Where |
|---|---|---|
| **Chainrings** (front, few, big) | The run chain's three stages: **opener → run → closer** | [20.2](#202-the-chainrings-opener-run-closer) |
| **Cassette** (rear sprockets) | The choices inside a stage: which song, which in-between, a global or a local warm-up | [20.3](#203-the-cassette-choices-inside-a-stage) |
| **Derailleurs** | The switches that move the chain live: `start <run>`, stop run, Esc, skip, the thanks ending, `system-` commands | [20.4](#204-the-derailleurs-shifting-live) |
| **Chain** | The master chat queue: everything shown belongs to a turn, and turns run one at a time, each to its end | [20.5](#205-the-chain-the-master-chat-queue) |
| **A second chainring set: the live game** | Opener → rounds (a question, its answer, value points, eggs, a reminder) → grand prizes → closer | [20.8](#208-the-live-game-gear) |
| **The rider's memory** | The brain chain: short and long term, self and chat, and the flow | [20.9](#209-memory-short-and-long-term-self-and-chat) |
| **Teeth** | The values in the settings folder (`webroot/gopod_settings/`) and each song's `knobs.json` | [20.6](#206-the-teeth-values-not-code) |

### 20.2 The chainrings: opener, run, closer

`start <run-name>` drives one run through three stages, always in this order.

| Stage | What plays | Where | Status |
|---|---|---|---|
| **Check-in** | The run checks in with Goverlord and takes its turn at once; a second start gets `session.busy_text` | `GopodRunStart`, `/api-sdk/gopod_run` | Locked by `TestSessionOneRunAtATime` |
| **Opener** | `session.started_text` ("{run} started. ..."), then the run's warm-up: the LLM warm-up bubble, the in-betweens, the ready line ("Language Model Loaded, Master CrushN8r." + the song's `llm_ready_text` + "Your move. (music)") | [6.4](03_SESSION.md#64-every-runs-warm-up), `gopodRunReadyLine` | GOLDEN, locked by `TestSongWarmupReadyAfterWarming`, `TestWarmupQueuesInput` |
| **Run** | The song or script itself: its `story.md` steps in order | the song engine (`run_golden_song_001.py`), the interview runner | per song |
| **Closer** | The song's own last lines, then `session.complete_text` ("{run} complete. ... Spectacular work, Master CrushN8r.") for a run that played to its end, or `session.stopped_text` after a stop | `GopodRunEnd` | One template for every named run (`session.run_names`) |
| **Check-out** | The run checks out; the chat hold, background loop and Native go back as they were | `GopodRunEnd` | Locked by `TestSessionClearsAGoneRun` |

The session's own closer is the thanks ending ([6.10](03_SESSION.md#610-the-thanks-ending)): a
run going is closed out first, then the session ends.

### 20.3 The cassette: choices inside a stage

| Stage | The choice | Set by |
|---|---|---|
| Opener: warm-up | A cold model is waited out; a warm one skips the wait | the model's state (`gopod_llm_warmup.go`) |
| Opener: warm-up | A song may opt out of it | `"llm_warmup": false` in its `knobs.json` |
| Opener: in-betweens | The `_llm_song` run, else `warmup_lines`, the interview's `start_lines`; skippable | the song's own folder |
| Opener: ready line | The song's own words | the song's `llm_ready_text` |
| Run | Which song | the run name, `session.run_names` |
| Run | Live robots or not | `system-robots-off` (shelved) |

### 20.4 The derailleurs: shifting live

The control lane never waits for a turn. It shifts the chain the moment it's sent.

| Switch | Shifts | Where |
|---|---|---|
| `start <run-name>` | into a run's chain | the custom intents, `GopodRunStart` |
| stop run, Esc | out of the run, through its closer (`stopped_text`) | [6.6](03_SESSION.md#66-stopping-a-run), `GopodEscape` |
| skip | past the in-betweens (opener) or to the next scripture (tour) | [6.7](03_SESSION.md#67-skip) |
| thanks brobots | out of the session, through the thanks ending | `sign_off`, `chat_hold.thanks_phrases` (only the phrases listed there count) |
| `system-` commands | answered at once, outside every turn | [14](07_SYSTEM.md) |

### 20.5 The chain: the master chat queue

- Everything the chat shows belongs to a **turn**: a question and its answers, a run from
  check-in to check-out, the tour, the comm check.
- **Turns run one at a time, each to its end.** A run's own lines are its turn's nested queue,
  stamped with its number.
- **A run's check-in takes its turn at once**, without waiting for the line that started it
  (2026-10-30: the wait took the song past its 10 s check-in).
- **Stop run and Esc drop the turn going** (`drop:<turn>`); the page drops its waiting lines
  the moment it hears.
- Code: `gopod_chat_queue.go`. Locked by the `gopod_chat_queue_test.go` tests.

### 20.6 The teeth: values, not code

A gear's teeth are its values: words, waits, volumes, voices. They live in the settings folder
(`webroot/gopod_settings/`) and each song's `knobs.json`, never written into the code. Tuning a
tooth never changes the gear. `system-set` changes one from the chat ([14.3](07_SYSTEM.md#143-editing-from-the-chat)).

### 20.7 A start during a run

A `start <run-name>` while a run is going is not applicable: it's nulled before it reaches any
queue (`session.busy_text`, "A run is going. Say stop run first."). Locked by
`TestSessionOneRunAtATime`.

### 20.8 The live game gear

A scripted run has its own end. Live chat could go on forever, so the game gives it a purpose:
**reveal every value point, find every egg.** The live game runs on the same chain as a run:
opener, rounds, closer.

| Stage | What plays | Where |
|---|---|---|
| **Opener** | Wire-Pod's start opener: the LLM warm-up bubble, the reporters' opening, the ready line. Then the session is **Ready**: live chat | [6.3](03_SESSION.md#63-wire-pod-start-and-the-opener), `00_llm_warming/` |
| **A round** | 1. A question comes in: typed, numpad, the mic, a prober<br>2. Lightning picks the topics it pulls (`lightning`)<br>3. The Brobot answers, revealing value points (the score)<br>4. An egg pops when the line's words find it<br>5. The reminder: "Hey, that reminds me..." names an egg still to find, with a prober to try ("Try asking: ...") | [8](04_CHAT.md#8-a-live-reply), [13](06_CAST_AND_GAME.md#13-eggs), `gopod_lightning.go`, `gopod_probers_list.go` |
| **Grand prizes** | Every value point revealed and every game egg found: Goverlord's `game.win_text` ("Grand prize! ... {score}. Now GOPOD Yourself!"). With the hidden eggs too: `game.master_text`. Once per session | `gopod_game.go` |
| **Closer** | "thanks brobots": the thanks ending - the score, the reporters' closing, the Brobots' sign-off, Record Session's keep-or-delete, "You're welcome." | [6.10](03_SESSION.md#610-the-thanks-ending) |
| **Reset** | The thanks ending resets the game: score, points said, reminders, eggs found back to 0 | `gopodGameReset` |

The teeth of the game: the chapters, topics and value points (`gopod_probes/chapters/`, locked in
`value_points_golden.json`), the probers (`probers_list_001.txt`), and the settings blocks
`lightning`, `value_points`, `topic_pictures`, `game`.

**The rounds have no fixed count.** A round is one question and its answer. The reminders keep
the game moving (each one names an egg still to find), and reminder rounds start again until
every topic is covered (`TestRemindersStartAgain`).

**After a won game: a free for all.** Viewers throw whatever at GOPOD until the session maxes
out. GOPOD Yourself after that.

**Not chained yet: the chat participation awareness contract.** A templated, local contract every
chat participant follows - the minimum requirement is basic chat participation awareness: who
they're chatting with, and about what. Today only the Gomads get the who / where /
what's-going-on line (`gopodHereNow`).

### 20.9 Memory: short and long term, self and chat

Big brains are code, so the language model doesn't have to think and is free to add the persona's
style. Every participant answers from the same clean, code-made memory (2026-10-31, the brain chain,
`gopod_brain_chain.go`, settings `brain_chain`):

| Memory | What it holds | Made from | Where it shows |
|---|---|---|---|
| **Self, long term** | who I am: nickname, name, job title, role, who helps me; my voice and banter | my profile (the master format) | my prompt; my brain bubble's Self line |
| **Self, short term** | what I'm asked right now, and the task | the line, the intro recipe, the take | my prompt; the Guided Task Flow line |
| **Chat, short term** | the last few answers: who asked, who answered, the line, its kind | the brain chain's last `short_term` (4) records | every prompt ("Recent chat"); not in the bubble (no "Last:") |
| **Chat, long term** | the session so far: the game's score, the runs Doc and Pip performed (kept until the session ends), who has spoken, who's introduced | the whole chain and the game | every prompt; the brain bubble's "Session:" line |
| **Flow** | one or two helper lines for what comes next | the chain: an intro or an answer, then the first to speak or "Does recent chat help me answer this question asked by: Master CrushN8r?" | every prompt; the brain bubble's "Guided Task Flow:" line |

- **One brain template** (2026-10-31, the operator's scaffold; settings `brain_chain.lines`): under a
  Brobot's `Response-chain:` line - `I am Pip (Brobot 2), <his profile's "In short">`; `In the chat:` every
  member but the operator with his profile's job title in brackets, `(me)` for himself; `So far: N have
  spoken, N introduced themselves. Guided Task Flow: Does the recent chat help me answer the question
  asked. Reminders: a, b, & c`. No `Asked by:`, no `Last:`, no score lines. Every brain bubble (Brobots
  and Gomads) shows it - left to right, header too - and every prompt gets the same lines.
- **The crew roster** (2026-10-31): `gopod_probes/personas/crew_roster.md`, a topic-cluster in the
  master format (`## Crew`, one `Name: description` point each). "In the chat" reads it; system-check
  names anyone in the chat without a line (`TestCrewRoster`).
- **Topic-clusters for settings** (2026-10-31): `gopod_probes/clusters/` - `easter_eggs.md`,
  `golden_answers.md`, `butler_lines.md`, `nicknames.md`, in the master format. Each section names the
  settings place it fills in brackets (`## Run lines (session)`); each point is `key: value` (a list
  split on ` | `). They are the one source - the settings files no longer hold those values; Wire-Pod
  lays them over at its start and for the page. system-check names any point that lands nowhere;
  system-locate finds their keys. Edit the md, then wpr (`TestTopicClusters`).
- **Every lit line scrolls the chat** (2026-10-31): the highlight's scroll rule is part of lighting a
  line, so a tall bubble of any kind (Gomads too) moves on as its lines are said.
- **The chat's own scroll bar** (2026-10-31, `chat_window.scroll_bar`): drag the thumb to scroll,
  drag either end to make it longer or shorter (less or more travel), double-click for its natural
  size; its length is kept per browser. Admin, visitor and app pages alike; wheel, touch and keys
  scroll as before.
- **Chunks are whole sentences** (2026-11-01): a long voiced line is made sentence by sentence
  (short ones join the next), each one voice; its comma pauses play inside it - never cut at a comma
  (`TestChunksAreWholeSentences`).
- **Every spoken line reads left to right** (2026-11-01): one rule, every bubble kind - each line
  spans the bubble, so lines never sit jagged in a centred bubble.
- **Stop run drops what's still being made** (2026-11-01): a Gomad's answer still being written when
  stop run comes is dropped; a stop with nothing running still cuts voices playing on the page.
- **Reminders name topics, never eggs** (2026-10-31); the eggs still work on their own phrases.
- **The intro template:** code says the opening (`I'm {name}, my nickname is {nickname}, because
  {why}.`) and the cast line (`intro_recipe.cast_line`), once; the model writes only the middle
  (`gopodIntroWrap`, `TestIntroTemplate`). An intro gets no value point and is its own answer type.
- **Clean:** a record is made by code, never by the model - no earlier answer's colour is carried on.
  The Gomads get the brain instead of the raw chat (`gomads_raw_chat` false); the Brobots keep their
  own shared chat and get the brain on top.
- **Kept:** the session folder keeps `brain_chain.jsonl`; the game's reset (the thanks ending, the
  comm check's end) starts a new chain.
- **Seen:** `GOPOD_BRAIN` lines in the journal - each record, and what each prompt was fed.
  `system-brain` shows the chain.
- **The KP test run** (`system-comm-check`) is the diagnostic: every key's answer is a record, every
  prompt's brain is logged, and the run shows whether the cleanest sufficient data set reaches each
  participant. `TestBrainChain`.

---

[Back to ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md)

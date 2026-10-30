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

---

[Back to ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md)

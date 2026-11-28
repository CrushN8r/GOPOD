# The gears: the run chain

Part of [ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md) - [Concepts](concepts/README.md) · [Songs](02_SONGS.md) · [The session and runs](03_SESSION.md) · [The Brobots chat](04_CHAT.md) · [Voices and sound](05_VOICES_AND_SOUND.md) · [The cast and the game](06_CAST_AND_GAME.md) · [The system menu, your stuff, the settings](07_SYSTEM.md) · [Golden rules, troubleshooting, tests](08_RULES_AND_TESTS.md) · [The gears](09_GEARS.md)

<details><summary>Game card: the gears</summary>
<!-- gopod topic: 12_the_gears -->

# the gears
- **chapter:** Show and tell

- **words:** the gears, run chain, derailleurs, geared bicycle
- **ask:** Tell me about the gears
- **summary:** built like a geared bicycle | the chain binds the gears, derailleurs shift it live | a locked gear, only its teeth tuned
- **source:** alias_play_studio/09_GEARS.md
- **ask 1:** Brobot 2, tell me about the gears
- **ask 2:** Brobot 1, the run chain - what do the derailleurs do?
- **ask 3:** Brobot 2, the gears - once one is locked, what gets tuned?

---

The gears. Reveal these, in order, in your own words:
1. GOPOD's game engine is built like a geared bicycle: the gears are fixed parts.
2. The chain binds the gears, and the derailleurs shift the chain between gears live - every run goes through the same chain.
3. Once a gear is locked, only its teeth - the values - are tuned. The gear stays the gear.

</details>

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
| **Opener** | `session.started_text`, then the run's warm-up: the LLM warm-up bubble, the in-betweens, the ready line (with the song's `llm_ready_text`) | [every run's warm-up](03_SESSION.md#every-runs-warm-up), `gopodRunReadyLine` | GOLDEN, locked by `TestSongWarmupReadyAfterWarming`, `TestWarmupQueuesInput` |
| **Run** | The song or script itself: its `story.md` steps in order | the song engine (`run_golden_song_001.py`), the interview runner | per song |
| **Closer** | The song's own last lines, then `session.complete_text` for a run that played to its end, or `session.stopped_text` after a stop | `GopodRunEnd` | One template for every named run (`session.run_names`) |
| **Check-out** | The run checks out; the chat hold, background loop and Native go back as they were | `GopodRunEnd` | Locked by `TestSessionClearsAGoneRun` |

The session's own closer is the thanks ending ([the thanks ending](03_SESSION.md#the-thanks-ending)): a
run going is closed out first, then the session ends.

### 20.3 The cassette: choices inside a stage

| Stage | The choice | Set by |
|---|---|---|
| Opener: warm-up | A cold model is waited out; a warm one skips the wait | the model's state (`gopod_llm_warmup.go`) |
| Opener: warm-up | A song may opt out of it | `"llm_warmup": false` in its `knobs.json` |
| Opener: in-betweens | The `_llm_song` run, else `warmup_lines`, the interview's `start_lines`; skippable | the song's own folder |
| Opener: ready line | The song's own words | the song's `llm_ready_text` |
| Run | Which song | the run name, `session.run_names` |

### 20.4 The derailleurs: shifting live

The control lane never waits for a turn. It shifts the chain the moment it's sent.

| Switch | Shifts | Where |
|---|---|---|
| `start <run-name>` | into a run's chain | the custom intents, `GopodRunStart` |
| stop run, Esc | out of the run, through its closer (`stopped_text`) | [stopping a run](03_SESSION.md#stop-skip-replay), `GopodEscape` |
| skip | past the in-betweens (opener) or to the next scripture (tour) | [skip](03_SESSION.md#stop-skip-replay) |
| thanks brobots | out of the session, through the thanks ending | `sign_off`, `chat_hold.thanks_phrases` (only the phrases listed there count) |
| `system-` commands | answered at once, outside every turn | [14](07_SYSTEM.md) |

### 20.5 The chain: the master chat queue

- Everything the chat shows belongs to a **turn**: a question and its answers, a run from
  check-in to check-out, the tour, the comm check.
- **Turns run one at a time, each to its end.** A run's own lines are its turn's nested queue,
  stamped with its number.
- **A run's check-in takes its turn at once**, without waiting for the line that started it.
- **Stop run and Esc drop the turn going** (`drop:<turn>`); the page drops its waiting lines
  the moment it hears.
- Code: `gopod_chat_queue.go`. Locked by the `gopod_chat_queue_test.go` tests.

### 20.6 The teeth: values, not code

A gear's teeth are its values: words, waits, volumes, voices. They live in the settings folder
(`webroot/gopod_settings/`) and each song's `knobs.json`, never written into the code. Tuning a
tooth never changes the gear. `system-set` changes one from the chat ([the system commands](07_SYSTEM.md#the-system-commands)).

### 20.7 A start during a run

A `start <run-name>` while a run is going is not applicable: it's nulled before it reaches any
queue (`session.busy_text`). Locked by
`TestSessionOneRunAtATime`.

### 20.8 The live game gear

A scripted run has its own end. Live chat could go on forever, so the game gives it a purpose:
**reveal every value point, find every egg.** The live game runs on the same chain as a run:
opener, rounds, closer.

| Stage | What plays | Where |
|---|---|---|
| **Opener** | Wire-Pod's start opener: the LLM warm-up bubble, the reporters' opening, the ready line. Then the session is **Ready**: live chat | [the opener](03_SESSION.md#wire-pod-start-and-the-opener), `00_llm_warming/` |
| **A round** | 1. A question comes in: typed, numpad, the mic, a prober<br>2. Lightning picks the topics it pulls (`lightning`)<br>3. The Brobot answers, revealing value points (the score)<br>4. An egg pops when the line's words find it<br>5. The reminder: "Hey, that reminds me..." names an egg still to find, with a prober to try ("Try asking: ...") | [8](04_CHAT.md#8-a-live-reply), [eggs](06_CAST_AND_GAME.md#eggs), `gopod_lightning.go`, `gopod_probers_list.go` |
| **Grand prizes** | Every value point revealed and every game egg found: Goverlord's `game.win_text`; with the hidden eggs too, `game.master_text`. Once per session | `gopod_game.go` |
| **Closer** | "thanks brobots": the thanks ending - the score, the reporters' closing, the Brobots' sign-off, Record Session's keep-or-delete, "You're welcome." | [the thanks ending](03_SESSION.md#the-thanks-ending) |
| **Reset** | The thanks ending resets the game: score, points said, reminders, eggs found back to 0 | `gopodGameReset` |

The teeth of the game: the chapters, topics and value points (`gopod_probes/topic_clusters/personas/company/4_go_for_it/`, locked in
`value_points_golden.json`), the probers (`probers_list_001.txt`), and the settings blocks
`lightning`, `value_points`, `topic_pictures`, `game`.

**The rounds have no fixed count.** A round is one question and its answer. The reminders keep
the game moving (each one names an egg still to find), and reminder rounds start again until
every topic is covered (`TestRemindersStartAgain`).

**After a won game: a free for all.** Viewers throw whatever at GOPOD until the session maxes
out. GOPOD Yourself after that.

### 20.9 Memory: short and long term, self and chat

Big brains are code, so the language model doesn't have to think and is free to add the persona's
style. Every participant answers from the same clean, code-made memory:

| Memory | What it holds | Made from | Where it shows |
|---|---|---|---|
| **Self, long term** | who I am: nickname, name, job title, role, who helps me; my voice and banter | my profile (the master format) | my prompt; my brain bubble's Self line |
| **Self, short term** | what I'm asked right now, and the task | the line, the intro recipe, the take | my prompt; the Guided Task Flow line |
| **Chat, short term** | the last few minutes: who asked, who answered, the line, its kind, and each answer's note | the brain chain's records of the last `recent_minutes` | every prompt ("Recent chat"); not in the bubble (no "Last:") |
| **Asked before** | the same question's note, so it isn't answered from scratch | the brain chain: the last `recent_minutes`, then `deep_minutes` | the brain bubble and the prompt ("Asked before") |
| **Chat, long term** | the session so far: the game's score, the runs Doc and Pip performed (kept until the session ends), who has spoken, who's introduced | the whole chain and the game | every prompt; the brain bubble's "Session:" line |
| **Flow** | one or two helper lines for what comes next | the chain: an intro or an answer, then the first to speak or "Does recent chat help me answer this question asked by: Master CrushN8r?" | every prompt; the brain bubble's "Guided Task Flow:" line |

- **One brain template** for every participant's brain bubble and prompt: Goverlord's card,
  [`goverlord-brain.md`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/company/1_admin/goverlord-brain.md).
- **The crew roster:** the company card,
  [`math-aftermath-news-channel.md`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/company/math-aftermath-news-channel.md); "In the chat" reads it.
- **Cards fill the settings:** a card section naming a settings place in brackets
  (`## Run lines (session)`) is that place's one source; Wire-Pod lays it over at start.
- **Clean:** a record is made by code, never by the model. Every prompt's brain is logged
  (`GOPOD_BRAIN`); `system-brain` shows the chain; the comm check is the diagnostic run.

---

[Back to ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md)

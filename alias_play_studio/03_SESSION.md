# The session and runs

Part of [ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md) - [Concepts and terms](01_CONCEPTS.md) · [Songs](02_SONGS.md) · [The session and runs](03_SESSION.md) · [The Brobots chat](04_CHAT.md) · [Voices and sound](05_VOICES_AND_SOUND.md) · [The cast and the game](06_CAST_AND_GAME.md) · [The system menu, your stuff, the settings](07_SYSTEM.md) · [Golden rules, troubleshooting, tests](08_RULES_AND_TESTS.md) · [The gears](09_GEARS.md)

## 6. The session and runs

### 6.1 The session and its phases

A **session** is one Wire-Pod start (`wpr 2`) to the next. Goverlord orchestrates it and keeps
its one record (`gopod_session.go`).

| Phase | What it means | What gets through |
|---|---|---|
| **Warming** | From the start until LLM Ready. | Nothing reaches the Brobots: "LLM not ready? Too bad!! Wait!!" |
| **Ready** | Live chat. | Every way in, to the Brobot it names. |
| **Run** | One run is going, and it owns the Brobots. | Only `stop run`, thanks, Record Session's answer and the run's own lines. Anything else gets "Run in progress. Say stop run or thanks." and is dropped. |

**Chat memory.** Wire-Pod keeps each robot's chat in memory, and GOPOD's shared Brobots chat
replaces it in the prompt (`kgsim.go`), so nothing is fed twice. Both are in memory only: `wpr 2`
starts both empty, and the page clears with the new boot. F5 redraws the page and touches
neither. The session log keeps everything.

### 6.2 Runs

| Run | Start | Played by | Ends with |
|---|---|---|---|
| Interview | `start interview` | Wire-Pod, inside the session | "Draft run completed." (in the thanks ending) |
| Bingo test | `start bingo test` | the song engine | its last step |
| Baby brobots (the nap) | `start baby brobots` | the song engine | "Brobots Ready! …", then "Run complete." |
| Fart joke | `start fart joke` | the song engine | "Fart joke complete." |
| A replay | `replay <name>` | Wire-Pod | "[name] complete." |
| A song from the chat | `system-play ...` | the song engine | its last step |
| A `pha0b` range, live | the terminal | the song engine | its last step |
| phcal, live | the terminal | phcal | phcal's exit |

- **One at a time.** Every run checks in with Goverlord when it starts and out when it ends
  (`/api-sdk/gopod_run`). A second start gets "A run is going. Say stop run first." A run whose
  process is gone without checking out is cleared.
- **The run owns the Brobots:** the chat's stay-in-place hold ends when a run starts and isn't
  taken while it goes; no live LLM answers; Wire-Pod's native answers are off.
- **A dry run** touches nothing and doesn't check in.
- Locked by `TestSessionOneRunAtATime`, `TestSessionClearsAGoneRun`.

### 6.3 Wire-Pod start and the opener

At every Wire-Pod start:

1. A new chat feed with a new boot stamp: the page clears and reloads once
   (`GopodChatFeedStart`). The warm-up bubble is `gopod_llm_warmup.go`.
2. The sounds and every song's own soundtrack are copied to the page's voice folder. Every
   scripted line's voice is made (see [10.2](05_VOICES_AND_SOUND.md#102-voice-files-made-once-play-many)).
3. The speaker icon starts **off**: no sound and no menu until you click it.
4. The opener plays (below). It waits for the chat page to be here.
5. When the ready message is said, the warm-up gate opens. Locked by `TestWarmupGate`,
   `TestWarmupReadyLineIsOneBubble`.
6. After the warm-up's last step, the mic check runs ([6.11](#611-mic-missing)).

Also at start: the egg pictures' points are made for any new or changed picture; egg pictures
are looked at ahead, quietly, 30 s after start; `system-check` runs (`GOPOD_CHECK` lines in the
journal); the Edge TPU worker loads its model once; the Brobots' eye colours are read.

**The opener is a song.** The template is `goverlord/runtime/songs/_new_song/_llm_warm/`; the
opener LLM Warming song is `00_llm_warming/` (`song_id` `brobots_llm_warming`, settings
`llm.opener_song_id`). Wire-Pod's start plays its steps in order:

| Step | Plays |
|---|---|
| `goverlord_warm_0101` | message 1 (waits for the voices) |
| `goverlord_warm_0102` | message 2 |
| `goverlord_warm_wait_0103` | waits until the model is loaded |
| `llm_song_0104` | its `_llm_song` run: the reporters' opening |
| `goverlord_warm_ready_0105` | message 3, the ready message |
| `goverlord_sting_0106` | the end sting |

Its words are its `story.md`; edit one and `wpr 2` (the voice file is made for the new words).
With no opener song, or one Wire-Pod can't read, the built-in opener from the settings plays
(`voices_text`, `warmup_text`, `ready_text`), logged `GOPOD_OPENER_SONG`. Proved by
`TestOpenerSongPlaysLikeBuiltIn` (the song plays the built-in opener's chat lines, voices and
stings, in the same order) and `TestLLMWarmTemplate`.

**GOLDEN: the opener template.** It is the base for every song run. Three messages in the one
System llm bubble, one line each, the owner's exact lines (do not change the words):

- message 1, at every start (it is the check): "Voice files generation check." - no sting. It
  waits for its own voice file before it is said.
- message 2: "Voice files generated. Language Model Loading"
- message 3: "Language Model Loaded, Master CrushN8r. Your move. (music)"

"(music)" is shown in the bubble, for deaf viewers, where the sting plays; the butler doesn't
say it. Each bubble is built and its voice loaded off screen first; only when both are ready is
it placed in the chat, and its voice waits until it shows. The butler says "please" as "pleeze"
(`system_say_words`).

A song run uses the same base: message 2, then "Language Model Loaded, Master CrushN8r." / the
song's own `llm_ready_text` / "Your move. (music)" - one bubble, one spoken line. Settings:
`llm.voices_text`, `llm.warmup_text`, `llm.wait_text` (empty), `ready_text`,
`warmup_done_text` (empty), `llm.run_ready_text`, `run_ready_end`. Code: `gopodRunReadyLine`, and the same in
`run_golden_song_001.py`. The interview's ready line is `interview_script.llm_ready_text`.

**The opener says a message line by line.** Each line of a message (and of the ready message)
is its own voice file. Every line's voice is made first (the bubble showing "Voice file
generating." meanwhile), then line 1 is shown and said, then line 2. A `(music)` anywhere plays
its sting where it's written. While voices are being made, the step's first line shows at once
with "Voice file generating." under it, then on its own once made.

### 6.4 Every run's warm-up

**GOLDEN, locked: the System llm bubble's sequential run check.** One template for every run:
the opening after `wpr 2`, the interview, a replay, every song with its own reporter lines.

- **Start like Wire-Pod does:** the LLM warm-up bubble first, kept at the bottom under every new
  bubble until LLM Ready, its LEDs filling; then whatever plays in between (the `_llm_song` run,
  else `warmup_lines`, the interview's `start_lines`); then the ready line, said by Goverlord. A
  song opts out only with `"llm_warmup": false`.
- **A cold model is waited out; a warm one skips that wait.** The warming message stays while
  the in-between lines play, until it has been fully said; only then does the bubble turn to the
  ready message and say it - also in a song with no lines in between. Locked by
  `TestSongWarmupReadyAfterWarming`.
- **Each message is a line in the chat feed,** so the page shows every step in turn even when
  it's opened late (each held `llm.warmup_check_seconds`, or for its spoken line). The ready
  bubble takes its turn in the chat's order (the feed's `warmup_ready` line). Locked by
  `TestWarmupSequenceGolden` (Wire-Pod start and runs) and `test_golden_song_step_names_001.py`
  (songs).
- **The run goes on only once the page has played the ready message and the closing sting**
  (its heard check), never on a guess of their length.
- **Stings:** no sting before message 1; the end sting plays once, after the ready message,
  where its `(music)` is. None in between. Locked by `TestWarmupStings`.
- **Input pauses for it:** from a run's first warm-up message until its ready line has been
  said, every typed or numpad line (`stop run` too) waits in a queue, then plays in the order it
  came (`GOPOD_WARMUP_QUEUED` in the log). Locked by `TestWarmupQueuesInput`.
- A replay's first reporter lines play inside its warm-up.
- Esc or `system-skip` skips the in-between reporter lines ([6.7](#67-skip)).

### 6.5 The page away, pause and continue

**Goverlord awaits your return.** The terminal run and the web page run are the same run. The
chat page is **away** when it says it's hidden, when the chat is collapsed (the Brobots Chat
button), or when it hasn't been heard from for `session.page_present_seconds` (20; it reports
every 15 s) - closed, or never opened. The page is **here** only once it can play: after
`wpr 2`/F5, you click Log.

While it's away:
- the opener waits before each of its steps, the System llm bubble showing
  `session.awaits_text` "Goverlord awaits your return..." with its hourglass; nothing plays;
- a song run waits before it starts and before each step, and the terminal says so
  (`GOLDEN_AWAIT ... the chat page is away`, then `... is back after Ns`);
- the interview waits before its next question, a replay before its next line;
- Goverlord's "[run] paused." is shown (a Goverlord bubble), not said; "[run] continues." when
  it's back. A replay's pause is in the journal.
- A timed song (the nap) plays on.

Code: `gopod_page_presence.go`, `wait_for_page` in `run_golden_song_001.py`. Locked by
`TestPagePresence`.

### 6.6 Stopping a run

- `stop run` stops the run going (a song, the interview, also inside a replay) at its next safe
  spot, with control handed back, and says "Run stopped." `stop session` stops a replay itself.
- A timed song (the nap): `stop run` (or "stop nap" / "stop baby robots") ends it before its next
  step - within a second, since its steps are short pauses. Its soundtrack stops on the page,
  narrator bubbles not yet shown go, the one on screen shows its words, the Brobots wake, and
  Goverlord says it stopped.
- A thanks closes the run out, then the thanks ending ([6.10](#610-the-thanks-ending)).

### 6.7 Skip

Skip the reporters' warm-up lines with any of: **Esc** (outside the text box), the numpad's
**/** key, a typed **"skip"** (never sent to the Brobots), **"skip"** said into the mic, or
**`system-skip`** (typed or menu group 1).

- It stops the reporter line being said. The rest of the reporters' lines show at once, whole
  and unspoken, their voices dropped, up to the next line that isn't theirs or the opening's
  end. Then the ready message.
- Heard markers still pass, so Wire-Pod and the songs go on.
- Managers and Brobots are never skipped.
- Every skip is logged with how it came in: "skipped (Esc|numpad /|typed|menu|mic)". A spoken
  skip is a `skip:` marker (`gopod_page_controls.go`).
- Page-only (`skipReporters`).

### 6.8 Replay a session

`replay <name>` (the session folder's name: `095819_session`, `095819` or
`/095819_session`) plays that session's score (`story.md` + `knobs.json`, the words as
edited in `story.md`) back through the live chat (`gopod_session_replay.go`):

1. Goverlord's "[name] started." (a replay's lines name it "Session 095819").
2. In order: the reporters' lines; each operator line with its Brobot's logged answer said
   through the chat-reply path (cue, moves, thinking, bubble, highlight); handoffs; songs
   started again by their own scores; the interview's section from the score.
3. "[name] complete."

- A thanks in the score plays the ending live. `stop run` stops a run inside it; `stop session`
  stops the replay.
- Every answer in a score keeps its cue (`say`) and a scripted answer its moves (`moves`).
- A replay is never recorded: Record? stops (kept) and can't start during it.
- Locked by `TestSessionReplayPlan`.

### 6.9 Record Session (the "Record?" box)

1. **Ticked:** recording starts at the next safe spot (never in the middle of a reply), in
   `<time>_brobots_session/` inside the open session's folder
   (`~/gopod_tts/sessions/<date>/<time>_session/`): the mic (48 kHz), the chat and Wire-Pod's
   log. Every session file lives in the session's folder: numpad presses as
   `numpad/<HHMMSS>_<key>_<n>.wav` (48 kHz), numpad errors in its `wirepod.log`. The label and
   the timer turn red; the timer restarts from the recording's start.
2. **While recording:** every chat line and answer is written; each answer is a take for the take
   sheet.
3. **Unticked:** it stops and keeps the session (no question). The timer goes back to 0.
4. **Thanks, or `max_minutes`:** it stops, and the Butler asks keep or delete (`record_session.ask_brobot`: the system speaker, 2026-10-30). The box stays
   ticked until the answer; unticking then counts as keep; no answer in `answer_wait_minutes`
   deletes.
5. **Keep:** the takes go on the take sheet and the folder is zipped. **Delete:** it's removed.
   A Brobot says which. Locked by the `TestGopodRecordSession…` tests.

### 6.10 The thanks ending

"thanks guys" (or "thanks brobots") anywhere in a line ends the session, but the ending waits
for the next safe spot: until no Brobot is answering (`sign_off.safe_spot_seconds`, then while
a reply is going). To All, Brobot 2's sign-off comes first.

Once per ending, in this order (`gopodThanksEnding`):
1. a run going is closed out first;
2. the score (logged);
3. the reporters' closing;
4. the brothers' sign-off lines;
5. Record Session's keep-or-delete question, and the wait for the answer;
6. "Draft run completed." if a draft played;
7. "You're welcome.", always.

Locked by `TestThanksClosesOutTheDraft`, `TestReporterClosingOncePerEnding`.

### 6.11 Mic missing, USB watch

After the warm-up's last step, `gopodMicCheck` looks for any of `numpad_ptt.mic_usb_ids` in `lsusb`.
If the mic is missing, Git Repo (`usb_host`; empty = Goverlord) says `mic_missing_text` (with his bubble), and only the mic input is disabled: a
numpad key works as a dry run - the same messages, "Mic disabled" in place of "Recording", then
"transcribing" and "try again" (`nothing_heard`). A failed `lsusb` isn't "mic missing", and a
missing mic is looked for again on each key press.

A numpad key while Goverlord's mic or numpad message is still being made or said waits for it:
"⏳ Wait - Goverlord is still giving the mic's message" (`wait_message`, `gopodUSBSaying`).

The mic service records through PulseAudio when PulseAudio holds the USB mic (GNOME's Sound
settings open; its default source is the mic) - `PULSE_FALLBACK`.

The USB watch (`gopod_usb_watch.go`, from then on, every `usb_watch_seconds`): when the mic or
the numpad (`numpad_usb_ids`) goes missing or comes back, Git Repo (`usb_host`; empty = Goverlord) says and shows
`mic_missing_text` / `mic_found_text` / `numpad_missing_text` / `numpad_found_text`. Two numpads
both work, one key at a time; two mics, the first found records. The mic service looks for the
mic fresh when it finds none (a mic plugged in later is found).

### 6.12 Battery watchdog

All session long, every 15 s, each Brobot's battery is read and logged
(`GOPOD_BATTERY_WATCH`), each reading tagged with the one before it. Below 3.7 V (golden rule 6),
or a **dip** of `dip_volts` (0.15 V) since the last reading (`GOPOD_BATTERY_WATCH ... dip`), he
says he's dizzy and needs help getting home (`battery_watch.lines`), with his bubble - once per
drop, again only after charging or 3.8 V. Off the charger, or on it but not charging (a bad
contact). Live chat only, at the next safe spot: never during a reply, a run or a draft. Every
`GOPOD_ROBOT_CMD` line ends with the Brobot's last battery reading. `battery_watch` in the
settings, `TestBatteryWatchOncePerDrop`.

The security guards' job: each Brobot's guard (`battery_watch.guards`: Brobot 1 - Intel 404 PNFH,
Brobot 2 - Intel 404 PNFT) speaks first, the duo response-chain way - `guard_warning`, then
`guard_handover` ("Brobot 1, what say you on this, Mr fuzzy logic?"), in his own bubble; then the
Brobot answers `reply`: the comeback opener ("Nice try, Mrs gear-head!"), then his dizzy line. No
guard, or no guard voice: the Brobot says his line alone.

---

---

[Back to ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md)

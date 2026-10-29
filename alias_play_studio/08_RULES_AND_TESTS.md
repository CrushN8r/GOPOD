# Golden rules, troubleshooting, tests

Part of [ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md) - [Concepts and terms](01_CONCEPTS.md) · [Songs](02_SONGS.md) · [The session and runs](03_SESSION.md) · [The Brobots chat](04_CHAT.md) · [Voices and sound](05_VOICES_AND_SOUND.md) · [The cast and the game](06_CAST_AND_GAME.md) · [The system menu, your stuff, the settings](07_SYSTEM.md) · [Golden rules, troubleshooting, tests](08_RULES_AND_TESTS.md)

## 17. Golden rules

1. **The golden path for a new move:** bench it in phcal → check it live → give it its note
   type → put it in a song. Nothing goes into a song untested.
2. **Take control once and hold it** for the whole run, never per step.
3. **Settle after taking control:** about 1.5 s before the first move or line, or it can be
   silently lost.
4. **Animation, then voice** (the chat): the speech starts 0.9 s after the animation is sent,
   and the next animation waits until the robot reports the last one done. A call that never
   finishes gives up after 10 s, so nothing hangs. The engine's own animation shapes are locked
   by a test.
5. **Angry needs pacing:** never fire it as the first thing after taking control.
6. **Watch the battery: 3.7 V is the line.** Healthy robots read 3.99–4.08 V; Brobot 2's crashes
   have come at 3.61–3.62 V. phcal flags every move below 3.7 V (`PHCAL_BLOCKED_LOW_BATTERY`),
   and the chat logs the battery after every reply (`GOPOD_BATTERY`). A low reading on the
   charger that isn't charging (or one stuck at the same value) is a likely reason Brobot 2
   freezes or reboots: check his charger contact, then reboot him.
7. **Stop at a safe spot:** `stop run` ends a run before its next step, with control handed back.
8. **Values live in files, never in code:** a song's in its knobs, the chat's in
   the settings folder (`webroot/gopod_settings/`).
9. **A locked thing gets a test:** the engine's animation lock, the chat's highlight, and this
   standard's own check. Nothing deploys if one fails.
10. **GOLDEN: chat bubbles appear whole, on time, and the chat scrolls once for them.** A new
    bubble is built whole while invisible, taking its room; the chat scrolls once so it's in
    view; then it shows. A timed bubble (the nap's narrator) is sent early
    (`narrator_lead_seconds`) and shows at its first line's second, never drawing in on screen.
    Its highlight lights in place.
11. **GOLDEN: a bubble is not its spoken words.** A spoken line's bubble shows with its voice, not
    before it, and a step that waits on a line (`system_line` with `wait_spoken`) goes on only
    once the chat page reports the voice played. Not heard in time stops the run. Do not go
    forward if an expected line is not spoken.
12. **One session, one run at a time, and the run owns the Brobots.** Goverlord is the session
    orchestrator. See [The session and runs](03_SESSION.md#6-the-session-and-runs).
13. **A song is found by its `song_id`.** Shuffling song folders never needs remapping. No folder
    name is written anywhere else.
14. **GOLDEN: the baton hands over Olympic fashion.** For every chat bubble, the next one is built
    whole while hidden and waits in its lane, taking its room below with no scroll, while a
    bubble on screen is still being said or sung (its highlight lit). Only then the baton passes:
    the chat scrolls once, the next bubble shows (a timed one on its second), and it lights.
    Nothing cuts the runner ahead short, your own line included. A line lit past
    `chat_window.speaking_max_ms` lets go. (This keeps an early narrator bubble's scroll and
    highlight from taking over while the bubble above is still being sung.)
    `showNarratorLine`/`lightNarrator` in `gopod_brobots_chat.js`.
15. **After spoken text, release the highlight, then carry on.** When a line (or a combo part)
    has been said, Wire-Pod writes the robot as no longer speaking and the chat page takes the
    highlight off; only then does the next thing go, and a waiting bubble gets the baton (rule
    14). `GopodSpeakSentences` ends every line with it; locked by
    `TestComboPartLightsItsPieceThenReleases`.
16. **Wire-Pod runs as root:** every Python start goes through `gopodAsOperator` (GOPOD_HOME and
    the operator's `~/.local` packages); a conda/venv Python keeps its own packages.
    `TestPythonRunsAsOperator` fails the deploy if one doesn't.
    **And everything it makes is the operator's** (one
    place, `gopod_own_files.go`): every program `gopodAsOperator` starts runs as the operator
    himself (his uid, gid and groups - audio, video, plugdev, input - and his `XDG_RUNTIME_DIR` for
    PulseAudio; the mic service's own runuser and the system tests' setpriv folded in), and every
    overlay write goes through `gopodWriteFile` / `gopodMkdirAll` / `gopodOwn`. Migrated once by the
    operator: `sudo find /home/goverlord -xdev -user root -exec chown goverlord:goverlord {} +`
    (1,199 files; `); live after: 0 root files, voice, TPU and mic
    workers running as goverlord. Left: `~/wire-pod` (upstream) can still make root files.

---

## 18. Troubleshooting

| You see | Cause | Fix |
|---|---|---|
| Mic or numpad don't work | Not found by `lsusb` | Check the parts by name ([1.6](01_CONCEPTS.md#16-desk-start)); Goverlord says the mic is missing after the warm-up and the numpad works dry. |
| A deploy didn't take | Wire-Pod started before the deploy | `wpr 2` again, then F5. |
| Nothing plays after start | The page isn't here | Click **Log**; open the chat; keep the tab visible ("Goverlord awaits your return..."). |
| A hand edit to `customIntents.json` was lost | Wire-Pod's web config page rewrites the whole file | Restart Wire-Pod after a hand edit, before touching that page. |
| A step plays the wrong thing | `zKnobs.json` wins while it's there | Check or remove `zKnobs.json`. |
| `STEP_NAME_WARNING` in a run | The step name doesn't read as its knobs | Rename it to match ([2.2](02_SONGS.md#22-step-names)); it only warns. |
| Deploy stopped by a test | A song's two files name different steps, or an unknown note | Fix the named step in both files. |
| The opener's first lines are shown and never said | A voice asked for while Kokoro was starting | It now waits up to `start_wait_seconds`; check the voice worker. |
| A bubble holds the next one up | A frozen robot or a lost end signal | It lets go after `chat_window.speaking_max_ms` (30 s); check `GOPOD_PAGE held` and `GOPOD_HEARD` in the journal. |
| A reply bubble missing or doubled | | `system-find FEED_CHECK`; see [9.8](04_CHAT.md#98-the-engine-and-the-page-in-dialogue). |
| QR codes or voice word timing fail silently | A Python start not run as the operator | Every Python start must go through `gopodAsOperator` (rule 16). |
| The Edge TPU worker broke | The operator's Python 3.10 packages in front of the coral env's 3.9 ones | A conda/venv Python keeps its own; `system-tpu restart`. TPU turned off after a replug: the command in `goverlord/runtime/data_gomad/models/README.md`. |
| Bingo run pauses ~61 s at an angry beat | `angry` = `anim_rtpickup_loop_10` (`animation_vocab.json`) is a loop clip: cut at 10 s, then the robot dropped Wire-Pod's connection | Known; the clip choice is in `animation_vocab.json`. |
| Brobot 2 freezes or reboots | Low battery on the charger, not charging | Golden rule 6: check his charger contact, then reboot him. |
| Weather is always the offline line in the dry run | The session dry run loads no weather key | Expected. |
| A flag shows as letters ("US") | The browser draws flag emojis as letters | Flags are drawn as pictures from `assets/flags`. |

**Watch the journal:** `journalctl -u wire-pod -f | grep -E "GOPOD_(FEED|PAGE|HEARD|AWAIT|OPENER)"`.
Other markers: `GOPOD_CHECK`, `GOPOD_SYNC`, `GOPOD_FEED_CHECK`, `GOPOD_WARMUP_QUEUED`,
`GOPOD_BATTERY`, `GOPOD_BATTERY_WATCH`, `GOPOD_ROBOT_CMD`, `GOPOD_OPENER_SONG`, `LLM_SONG`,
`GOPOD_REPORTER_VOICES`, `GOLDEN_AWAIT`.

---

## 19. Tests that lock behaviour

Run before every deploy; if one fails, nothing deploys.

| Area | Tests |
|---|---|
| Songs | `TestSongFoldersFollowTheStandard`, `test_golden_song_step_names_001.py`, `TestStoryMultiLine`, `TestThinkingProcessIsOne`, `TestComboPartLightsItsPieceThenReleases` |
| Session and runs | `TestSessionOneRunAtATime`, `TestSessionClearsAGoneRun`, `TestSessionReplayPlan`, `TestPagePresence`, `TestThanksClosesOutTheDraft`, `TestReporterClosingOncePerEnding`, `TestGopodRecordSession…` |
| Warm-up and opener | `TestWarmupGate`, `TestWarmupReadyLineIsOneBubble`, `TestWarmupSequenceGolden`, `TestWarmupQueuesInput`, `TestWarmupStings`, `TestSongWarmupReadyAfterWarming`, `TestOpenerSongPlaysLikeBuiltIn`, `TestLLMWarmTemplate`, `NoneRequiredWarmupTests`, `TestOpenerSkipsEmptyLines` |
| Replies | `TestGopodAddressed…`, `TestGopodShouldAnswer…`, `TestHighlightPiecesGolden`, `TestStuckReplyGivesUp`, `TestAllStartsWith`, `TestDropBrotherRepeat`, `TestWaitNoteDropped`, `TestSpokenCueWordKept`, `TestEmojiJoinersNotSpoken`, `TestCleanPolitics`, `TestBrobotLingoChoco` |
| Weather, math, photo, search | `TestWeatherPlace`, `TestWeatherCountry`, `TestMath*`, `TestMathExplainTheMath`, `TestPhotoKind`, `TestPhotoRecall`, `TestPhotoRounds`, `TestSystemSearchQuery` |
| Eggs | `TestPopupsOnlyLive`, `TestEggList`, `TestChapterEggs` |
| Page dialogue, voices | `TestFeedDialogue`, `TestReporterFilesNamedByVoiceAndSpeed` |
| Inputs | `TestNativeAnswersOnOffAndDuringARun`, `TestBatteryWatchOncePerDrop` |
| System commands | `TestSystemSetEditsInPlace`, `TestSystemSetStoryText`, `TestSystemSetReads`, `TestSystemLocate`, `TestSystemUndo`, `TestSystemCheck`, `TestSystemFind` |
| Platform | `TestPythonRunsAsOperator` |

---

[Back to ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md)

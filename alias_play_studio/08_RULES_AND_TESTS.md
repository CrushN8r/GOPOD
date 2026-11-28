# Golden rules, troubleshooting, tests

Part of [ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md) - [Concepts](concepts/README.md) · [Songs](02_SONGS.md) · [The session and runs](03_SESSION.md) · [The Brobots chat](04_CHAT.md) · [Voices and sound](05_VOICES_AND_SOUND.md) · [The cast and the game](06_CAST_AND_GAME.md) · [The system menu, your stuff, the settings](07_SYSTEM.md) · [Golden rules, troubleshooting, tests](08_RULES_AND_TESTS.md) · [The gears](09_GEARS.md)

<details><summary>Game card: rules and tests</summary>
<!-- gopod topic: 11_rules_and_tests -->

# rules and tests
- **chapter:** Show and tell

- **words:** rules and tests, troubleshooting, the golden path, tests that lock
- **ask:** Tell me about the rules and tests
- **summary:** the golden path for a new move | troubleshooting starts at the journal | tests lock behaviour
- **source:** alias_play_studio/08_RULES_AND_TESTS.md
- **ask 1:** Brobot 1, tell me about the rules and tests
- **ask 2:** Brobot 2, where does troubleshooting start?
- **ask 3:** Brobot 1, tests that lock - what happens if one fails before a deploy?

---

Rules and tests. Reveal these, in order, in your own words:
1. The golden path for a new move: bench it in phcal, check it live, give it its note type, put it in a song.
2. Troubleshooting starts with watching Wire-Pod's journal for GOPOD's markers.
3. Tests lock behaviour: they run before every deploy, and if one fails, nothing deploys.

</details>

## 17. Golden rules

1. **The golden path for a new move:** bench it in phcal → check it live → give it its note
   type → put it in a song. Nothing goes into a song untested.
2. **Take control once and hold it** for the whole run, never per step.
3. **Settle after taking control:** about 1.5 s before the first move or line, or it can be
   silently lost.
4. **Animation, then voice** (the chat): the speech starts `chat_moves.speech_after_animation_seconds` after the animation is sent,
   and the next animation waits until the robot reports the last one done. A call that never
   finishes gives up after 10 s, so nothing hangs. The engine's own animation shapes are locked
   by a test.
5. **Angry needs pacing:** never fire it as the first thing after taking control.
6. **Watch the battery.** Off the charger, below `battery_watch.low_volts` is low; on the
   charger, his own level FULL is full, whatever the volts. phcal blocks a move on a low battery,
   and every reply logs it (`GOPOD_BATTERY`).
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
    orchestrator. See [The session and runs](03_SESSION.md#the-session).
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
    **And everything it makes is the operator's:** every program it starts runs as the operator, and
    every overlay write goes through `gopodWriteFile` / `gopodMkdirAll` / `gopodOwn` (`gopod_own_files.go`).

---

## 18. Troubleshooting

| You see | Cause | Fix |
|---|---|---|
| Mic or numpad don't work | Not found by `lsusb` | Check the parts by name ([the desk start](concepts/the-workbench-words.md#desk-start)); Goverlord says the mic is missing after the warm-up and the numpad works dry. |
| A deploy didn't take | Wire-Pod started before the deploy | `wpr 2` again, then F5. |
| Nothing plays after start | The page isn't here | Click **Log**; open the chat; keep the tab visible ("Goverlord awaits your return..."). |
| A hand edit to `customIntents.json` was lost | Wire-Pod's web config page rewrites the whole file | Restart Wire-Pod after a hand edit, before touching that page. |
| A step plays the wrong thing | `zKnobs.json` wins while it's there | Check or remove `zKnobs.json`. |
| `STEP_NAME_WARNING` in a run | The step name doesn't read as its knobs | Rename it to match ([step names](02_SONGS.md#where-the-words-for-each-part-live)); it only warns. |
| Deploy stopped by a test | A song's two files name different steps, or an unknown note | Fix the named step in both files. |
| The opener's first lines are shown and never said | A voice asked for while Kokoro was starting | It waits up to `start_wait_seconds`; check the voice worker. |
| A bubble holds the next one up | A frozen robot or a lost end signal | It lets go after `chat_window.speaking_max_ms`; check `GOPOD_PAGE held` and `GOPOD_HEARD` in the journal. |
| A reply bubble missing or doubled | | `system-find FEED_CHECK`; see [9.8](04_CHAT.md#98-the-engine-and-the-page-in-dialogue). |
| QR codes or voice word timing fail silently | A Python start not run as the operator | Every Python start must go through `gopodAsOperator` (rule 16). |
| The Edge TPU worker broke | The operator's Python 3.10 packages in front of the coral env's 3.9 ones | A conda/venv Python keeps its own; `system-tpu restart`. TPU turned off after a replug: the command in `goverlord/runtime/models/README.md`. |
| Bingo run pauses ~61 s at an angry beat | `angry` = `anim_rtpickup_loop_10` (`animation_vocab.json`) is a loop clip: cut at 10 s, then the robot dropped Wire-Pod's connection | Known; the clip choice is in `animation_vocab.json`. |
| A Brobot freezes or reboots | Charger contact, heat, or a low battery | Check his contacts; `gofps 30`; the battery watch logs it. |
| Weather is always the offline line in the dry run | The session dry run loads no weather key | Expected. |
| A flag shows as letters ("US") | The browser draws flag emojis as letters | Flags are drawn as pictures from `assets/flags`. |

**Watch the journal:** `journalctl -u wire-pod -f | grep -E "GOPOD_(FEED|PAGE|HEARD|AWAIT|OPENER)"`.
Other markers: `GOPOD_CHECK`, `GOPOD_SYNC`, `GOPOD_FEED_CHECK`, `GOPOD_WARMUP_QUEUED`,
`GOPOD_BATTERY`, `GOPOD_BATTERY_WATCH`, `GOPOD_ROBOT_CMD`, `GOPOD_OPENER_SONG`, `LLM_SONG`,
`GOPOD_REPORTER_VOICES`, `GOLDEN_AWAIT`.

---

## 19. Tests that lock behaviour

Every test runs before every deploy; if one fails, nothing deploys. The tests are the
`*_test.go` files beside the code they lock (`goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/`)
and `test_*.py` in `goverlord/runtime/songs/tools/`.

---

[Back to ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md)

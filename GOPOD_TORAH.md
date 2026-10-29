# The GOPOD Torah

> Goverlord's Torah: In the beginning there was a great restart.

The master reference for GOPOD's system window: every scripture (the window's groups), every commandment (its
commands), every test, every run. A fun take, not a religious one. GOPOD writes these scrolls itself from its
running settings, so they always match the system as it grows - `system-check` says when they've fallen behind,
`system-torah` writes them again; `system-tour` goes through them in the chat.

*Click a commandment to run it, or it waits in the text box for your words (press Enter).*

## The Scriptures

- [Scripture 1, Restart and check](torah/SCRIPTURE_01_RESTART_AND_CHECK.md) - Just started, or made some edits?
- [Scripture 2, Play](torah/SCRIPTURE_02_PLAY.md) - You can run one song or one saved session at a time.
- [Scripture 3, Edit words and settings](torah/SCRIPTURE_03_EDIT_WORDS_AND_SETTINGS.md) - This is where you change what's said and how.
- [Scripture 4, Find](torah/SCRIPTURE_04_FIND.md) - Don't know where something lives?
- [Scripture 5, Apply](torah/SCRIPTURE_05_APPLY.md) - Edits take effect after a restart, from Scripture 1.
- [Scripture 6, Switches](torah/SCRIPTURE_06_SWITCHES.md) - Turn parts on or off for this session.
- [Scripture 7, Brobots](torah/SCRIPTURE_07_BROBOTS.md) - Wire-Pod's bot page, right here.
- [Scripture 8, Tune moves](torah/SCRIPTURE_08_TUNE_MOVES.md) - A move looks off?
- [Scripture 9, Test one thing](torah/SCRIPTURE_09_TEST_ONE_THING.md) - Something acting up?
- [Scripture 10, Wrap up](torah/SCRIPTURE_10_WRAP_UP.md) - End the show.

## The Commandments

Every system commandment, as `system-help` gives it:

```
system commands (typed on this keyboard):
  system-set <setting> <value>      change a setting (system-set master-volume 35%)
  system-set <path.to.setting> <value>   any setting by its path (system-set managers.qr_note Scan it!)
  system-add / system-remove <path> <item>   an item in a list, or key = value in a group
  system-lingo <brobots|reporters|butler> <word> = <said>   (remove <word>; the lingo alone: its list)
  system-topic [<topic> ...]        a chapter topic: words, points (say lines), label, reminder, pictures
  system-skip  (or Esc)             skip the voiced lines playing (from Voices ready on)
  system-restart                    a forced restart, like wpr 2 (speaker + mic set); the page reloads itself
  system-audit                      the full audit, no robots called: files, Go tests, web page, phone app, every test, the log
  system-score <song>               a song's score: every step and its words
  system-play <song> [from <step>] [to <step>] [robot 1|2] [dry]   a song or part of it (pha0b)
  system-gap <seconds>|off|default  reporter gap for system-play
  system-tweaks on|off              apply phcal tweaks before a live system-play
  system-calibrate <move> <1|2> [values] [dry]   one move alone (phcal)
  system-set <song> tempo <seconds>   the song's tempo (tempo-set); a step's tempo_factor too
  system-brobot <stats|volume|eyes|locale|units|location|timezone|say|wake|head|lift|wheels|camera> [1|2|both]
  system-faces [list|rename|delete|add] [1|2]   the faces a Brobot knows
  system-value-points [lock|<topic>]   the golden value points: locked? changed? a topic's; lock on purpose
  system-interview-rebuild            the interview song, rebuilt from the game's topics and value points
  system-anim <text> [1|2|both]     the animationWI tester: {{playAnimationWI||angry}} and words, through Wire-Pod's say route
  system-usb                        the USB devices plugged in (live in the menu)
  system-voices                     numpad voices: KPN > persona [voice] (change them in the menu)
  system-voices-report [3d|week|month|year]  voice files made, played, most played (report only)
  system-tpu [restart]              the Edge TPU worker's state, or start it again
  system-photos [1|2]               a Brobot's photos
  system-cams [tpu]                 the Brobots' cameras and the webcam side by side; tpu = the Edge TPU on the selected one
  system-chats <words> [operator|brobots|managers]   search every saved session's chat
  system-chat-save [session]        a readable chat.md in the session's folder
  system-menu-theme <preset>        the menu's look: dark, light, blue, green, purple
  system-visual [next|prev|<id>]    the TPU +/- rotation: a feed's frame and its Edge TPU objects (KP+ / KP-)
  system-version                    Wire-Pod's commit, the newest, what an update touches
  system-native [on|off]            Wire-Pod's own voice answers (was the Native? box)
  system-numpad on|off              the numpad keys as push-to-talk (was the Numpad? box)
  system-set <song> <step> <field> <value>   change one step (volume, text, timing)
  system-set <song> [step]          a song's steps, or one step
  system-set-list                   every setting, song and step field
  system-locate <thing>             where it lives, file:line (a setting, a song step, a word)
  system-locate-list                every setting's and song's file:line
  system-undo                       put back the last system-set
  system-status                     Wire-Pod, language model, page, run, Brobots' battery
  system-find <word>                where a word shows up: the log (6 hours), settings, songs
  system-search <words>             GOPOD's own web search, the results (to the Brobots: "search <words>")
  system-check                      every file the settings, chapters and songs name - all there?
  system-voices-rebuild [go]        check, back up every voice, then wpr 2 + F5 makes them all again
  system-test-list                  the tests: system-test-<name>, several in a row, system-test-all (the audit)
  system-test-game                  every topic, value point and Easter egg found by its own words (no robots, no model)
  system-record on|off              Record Session on or off (the page's Record? box)
  system-docbs [<pool> [add|remove <line>]]   DocBS's lines - the pools, one pool's lines, or add / remove one
  system-tour                       the system window tour: Goverlord goes through the Torah, scripture by scripture; safe commandments run (again: stop)
  system-torah                      the GOPOD Torah (GOPOD_TORAH.md, the system's master reference) written again from the system as it is now
  system-comm-check                 the pre-show comm check: KP1-KP9 held in turn, live - who they are, about vs all about, each other, the interns
  system-test-admin-run             the 100% admin run self-check: every line in order - points not revealed, eggs not found, prizes not given, dry reminders
  system-test-replay                the chat page plays every egg kind, both prizes and the Operator voices (no robots)
  system-help                       this list
```

## The Tests

Each checks one thing, nothing said or moved (`system-test-<name>`; `system-test-all` runs the audit).

| Test | What it checks |
|---|---|
| `weather-live` | the weather service answers for the usual place |
| `weather-offline` | with no weather service, the offline line is said |
| `time-place` | PLAYHEAD's time and place line |
| `takes` | the managers take the right lines |
| `qr` | a QR code is made for a web address |
| `python` | Python runs with the operator's libraries (qrcode, vosk) |
| `voices` | the voice worker is ready and voices are timed by Vosk |
| `personas` | CHALK and PLAYHEAD have their prompts and who lines |
| `eggs` | the egg list: chapters (less not_eggs), no surprise eggs, no weather |
| `game` | every topic, value point and Easter egg is found by its own words (no robots, no model) |
| `admin-run` | the 100% admin run, every line in order through the game: every value point, egg and prize, and no dry reminder (no robots, no model; the live game is put back) |
| `eggvoice` | an Easter egg voice is made from its kept pieces and a live title, joined |
| `replay` | the chat page plays every egg kind, both prizes and the Operator voices, and a visitor's page mirrors it, with no robots |
| `llm` | the language model answers (Ollama has the model) |
| `search` | GOPOD's own web search answers |
| `files` | system-check: every file named is there, story.md / knobs.json / zKnobs.json match |
| `page` | the chat page is open and here |
| `audio` | sound goes to GOPOD's speaker and comes from its mic |
| `tpu` | the Edge TPU worker is ready and names what's in a picture |
| `robots` | both Brobots answered the battery watch in the last 10 minutes |

The audit's order: files, python, voices, audio, tpu, llm, search, weather-live, weather-offline, time-place, takes, qr, personas, eggs, game, admin-run, replay, page, robots.

## The Runs

One card per run (`game.runs`) - start it, stop it, its song.

| Run | Start | Stop | Song |
|---|---|---|---|
| Interview | `start interview` | `stop interview` | 02_brobots_interview (the scripted draft) - pending, not on the menu yet |
| Bingo Test | `start bingo test` | `stop bingo test` | 01a_brobots_bingo_test |
| Baby Brobots | `start baby brobots` | `stop baby brobots` | 101_brobots_nap |
| Fart Joke | `start fart joke` | `stop fart joke` | 03_fart_joke |

## The Looks

The system window's looks (`system-menu-theme <name>`, then F5): blue, dark, green, light, purple, torah.

---

GOPOD - [README](README.md) · [asks](BROBOT_ASKS.md) · [the studio](ALIAS-PLAY-STUDIO.md)

# The GOPOD Torah

> The ancient scrolls Ron used to bring the Brobots to life.

The master reference for GOPOD's system window: every book (the window's groups), every commandment (its
commands), every test, every run. A fun take, not a religious one. GOPOD writes these scrolls itself from its running
settings, so they always match the system as it grows - `system-check` says when they've fallen behind,
`system-torah` writes them again.

*Goverlord's tour: start at 1 after a restart, then go where you need. Click a command - it runs, or waits in the text box for your words (press Enter).*

## The Books of the System Window

### Book 1 - Restart and check

Just started, or made edits? Restart, then check everything is up.

| Commandment | What it does |
|---|---|
| `system-restart` | a forced restart, like wpr 2 - speaker and mic set; the page reloads itself (Enter to confirm) |
| `system-skip` | skip the reporters' warm-up lines (or press Esc) |
| `system-status` | Wire-Pod, language model, page, run, battery |
| `system-test-all` | the full audit, PASS / FAIL for each |
| `system-check` | every file named is there; story.md / knobs match |

*All PASS? Go to 2 to play, or 3 to edit.*

### Book 2 - Play

Run a song or a saved session. One at a time.

| Commandment | What it does |
|---|---|
| `start interview` | the scripted interview |
| `start bingo test` | the Bingo test |
| `start baby brobots` | the Baby Brobots nap song |
| `start fart joke` | the Fart Joke |
| `system-play` | a song or part of it: <song> [from <step>] [to <step>] [robot 1/2] [dry] (steps: system-score) |
| `replay` | a saved session - add its name |
| `stop run` | stops the song playing |

*Something to change? Go to 3.*

### Book 3 - Edit words and settings

Change what's said and how. system-undo puts back the last edit.

| Commandment | What it does |
|---|---|
| `system-topic` | a chapter topic: its words, points, label, reminder, pictures (alone: the topics) |
| `system-lingo` | how a word is said: brobots / reporters / butler <word> = <said> |
| `system-set` | a setting by its path, or a song step's text / volume / timing |
| `system-add` | an item into a list (or key = value into a group) |
| `system-remove` | an item out of a list (or a key out of a group) |
| `system-set chat_moves.thinking_text` | the "Got it!" line: said while the model works, and the last line of the Value Points notice (then system-restart) |
| `system-set bubble_notices` | what goes in a notice in a Brobot's bubble: value_points, egg, native_intent |
| `system-set <song> tempo` | a song's tempo in seconds (a step's: system-set <song> <step> tempo_factor <n>) |
| `system-set system_menu.look` | this menu's own look: font_size, background, code_background, code_color, results_background |
| `system-menu-theme` | dark / light / blue / green / purple (then F5) |
| `system-voices-report` | voice files: made, played, most played, not played - past 3d / week / month / year (report only, nothing moves) |
| `system-voices` | numpad voices: KPN > persona [voice] - pick a voice per key (bold = used twice); robot-bodied keys need none |
| `system-set-list` | every setting, song and step field |
| `system-undo` | puts back the last edit |

*Don't know the name or path? Go to 4. Done editing? Go to 5.*

### Book 4 - Find

Don't know where something lives? Find it first.

| Commandment | What it does |
|---|---|
| `system-locate` | where a setting, song step or word lives, file:line |
| `system-find` | where a word shows up: the log, settings, songs |
| `system-score` | a song's score: every step and its words (alone: the songs) |
| `system-locate-list` | every setting's and song's file:line |
| `system-search` | GOPOD's own web search |
| `system-chats` | <words> [operator/brobots/managers] - search every saved session's chat |
| `system-chat-save` | a readable chat.md of the newest session (or name one) |

*Found it? Back to 3 to change it.*

### Book 5 - Apply

Edits take effect after a restart; edited story lines get new voices.

| Commandment | What it does |
|---|---|
| `system-voices-rebuild` | check, back up every voice, then restart makes them all again |
| `system-restart` | a forced restart, like wpr 2 - speaker and mic set; the page reloads itself |

*After the restart: back to 1.*

### Book 6 - Switches

Turn parts on or off for this session.

| Commandment | What it does |
|---|---|
| `system-numpad` | each numpad key: who it talks to, live or not |
| `system-numpad on` | numpad keys are push-to-talk |
| `system-numpad off` | numpad keys type as usual |
| `system-gap` | reporter gap for system-play: <seconds> / off / default |
| `system-tweaks on` | apply phcal tweaks before a live system-play |
| `system-tweaks off` | play the moves as written |
| `system-native on` | Wire-Pod's own voice answers on |
| `system-native off` | Wire-Pod's own voice answers off |

*Close the menu with the speaker icon.*

### Book 7 - Brobots

Wire-Pod's bot page, here: add 1 or 2 at the end for one Brobot (both when left out).

| Commandment | What it does |
|---|---|
| `system-visual next` | the TPU +/- rotation: next feed, photo, enhancement, TPU boxes and "I see..." (KP+) |
| `system-visual prev` | previous feed (KP-) |
| `system-cams` | every feed's photo in a row, then the TPU boxes on the selected one |
| `system-tpu` | the Edge TPU worker: ready or not |
| `system-tpu restart` | start the Edge TPU worker again |
| `system-usb` | the USB devices plugged in, live (hot-swap) |
| `system-brobot stats` | both Brobots' statistics |
| `system-version` | Wire-Pod's commit, the newest, what an update touches |
| `system-brobot volume` | <0-5> |
| `system-brobot eyes` | <teal/orange/yellow/lime/sapphire/purple/green> |
| `system-brobot locale` | <en-US ...> |
| `system-brobot units` | c / f |
| `system-brobot location` | <place> |
| `system-brobot timezone` | <America/Toronto ...> |
| `system-faces` | the faces a Brobot knows (list / rename <old> <new> / delete <name> / add <name>) |
| `system-photos` | a Brobot's photos |
| `system-brobot cube` | <pickupcube/fetchcube/rollcube/popawheelie/keepaway/findcube> - his own cube moves |
| `system-brobot beats` | the 126 sleep beats by family: sleep, rts-asleep, held-on-palm, petting |
| `system-brobot beat` | <alias> - plays one sleep beat (snores, warbles, sleep moves...) |
| `system-brobot anims sleep` | every sleep animation on the robot (any word: anims happy) |
| `system-brobot anim` | <name> - plays one (names from anims) |
| `system-brobot say` | <text> - said by the Brobot |
| `system-brobot wake` | "Hey Vector" without the words |
| `system-brobot head` | <speed, e.g. 3 or -3> - a short nudge |
| `system-brobot lift` | <speed> |
| `system-brobot wheels` | <left> <right> |
| `system-brobot camera` | live video from the Brobot's camera |

*Done? Close the menu with the speaker icon.*

### Book 8 - Tune moves

A move looks off? Try it alone on one Brobot, then play it in its song.

| Commandment | What it does |
|---|---|
| `system-calibrate` | <arm/nod/rattle/danger/cube> <1/2> [reps] [--hold S] [--speed N] [--cycles N] [--volume N] [dry] |
| `system-tweaks on` | then system-play uses the tuned values |
| `system-anim list` | the animation tokens |
| `system-anim` | the animationWI tester (WIP): a token and words, e.g. angry with a count |

*Play that part: system-play <song> from <step> to <step> (group 2).*

### Book 9 - Test one thing

Something acting up? Test just that part.

| Commandment | What it does |
|---|---|
| `system-test-list` | the tests, then system-test-<name> |
| `system-test-game` | every topic, value point and Easter egg found by its own words - no robots, no model |
| `system-torah` | the GOPOD Torah (GOPOD_TORAH.md, the master system reference) written again from the system as it is now |
| `system-comm-check` | the pre-show comm check: KP1-KP9 held in turn, live - who they are, about vs all about, each other, the interns |
| `system-test-admin-run` | the 100% admin run self-check: every line in order - points not revealed, eggs not found, prizes not given, dry reminders |
| `system-audit` | the full audit, no robots called - files, Go tests, web page, phone app, every test, the log (a few minutes) |
| `system-test-replay` | the chat page plays every egg kind, both prizes and the Operator voices - no robots (run it after any page change) |
| `system-test-` | one test - add its name (weather-live, qr, voices...) |

*A FAIL says what it saw - fix it in 3, then 5.*

### Book 10 - Wrap up

End the show.

| Commandment | What it does |
|---|---|
| `thanks brobots` | the ending |
| `stop session` | ends the session |
| `system-help` | every command, a line each |

## The Commandments

Every system commandment, as `system-help` gives it:

```
system commands (typed on this keyboard):
  system-set <setting> <value>      change a setting (system-set master-volume 35%)
  system-set <path.to.setting> <value>   any setting by its path (system-set managers.qr_note Scan it!)
  system-add / system-remove <path> <item>   an item in a list, or key = value in a group
  system-lingo <brobots|reporters|butler> <word> = <said>   (remove <word>; the lingo alone: its list)
  system-topic [<topic> ...]        a chapter topic: words, points (say lines), label, reminder, pictures
  system-skip  (or Esc)             skip the reporters' warm-up lines (once, after Voices ready)
  system-restart                    a forced restart, like wpr 2 (speaker + mic set); the page reloads itself
  system-audit                      the full audit, no robots called: files, Go tests, web page, phone app, every test, the log
  system-score <song>               a song's score: every step and its words
  system-play <song> [from <step>] [to <step>] [robot 1|2] [dry]   a song or part of it (pha0b)
  system-gap <seconds>|off|default  reporter gap for system-play
  system-tweaks on|off              apply phcal tweaks before a live system-play
  system-calibrate <move> <1|2> [values] [dry]   one move alone (phcal)
  system-set <song> tempo <seconds> the song's tempo (tempo-set); a step's tempo_factor too
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
| Interview | `start interview` | `stop interview` | 02_brobots_interview (the scripted draft) |
| Bingo Test | `start bingo test` | `stop bingo test` | 01a_brobots_bingo_test |
| Baby Brobots | `start baby brobots` | `stop baby brobots` | 101_brobots_nap |
| Fart Joke | `start fart joke` | `stop fart joke` | 03_fart_joke |

## The Looks

The system window's looks (`system-menu-theme <name>`, then F5): blue, dark, green, light, purple, torah.

---

GOPOD - [README](README.md) · [asks](BROBOT_ASKS.md) · [the studio](ALIAS-PLAY-STUDIO.md)

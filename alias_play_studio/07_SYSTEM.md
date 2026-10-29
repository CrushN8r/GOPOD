# The system menu, your stuff, the settings

Part of [ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md) - [Concepts and terms](01_CONCEPTS.md) · [Songs](02_SONGS.md) · [The session and runs](03_SESSION.md) · [The Brobots chat](04_CHAT.md) · [Voices and sound](05_VOICES_AND_SOUND.md) · [The cast and the game](06_CAST_AND_GAME.md) · [The system menu, your stuff, the settings](07_SYSTEM.md) · [Golden rules, troubleshooting, tests](08_RULES_AND_TESTS.md)

## 14. The system menu and commands

### 14.1 How system commands work

- Type them on the chat page's keyboard (or pick them in the system menu). They are a **fixed
  list of edits and reads, never a terminal command.** Typed on the chat page's keyboard.
- The answer shows in the terminal bubble (`system_set.role`, text at 50% size), headed by
  Goverlord - or in the menu's results while the menu is open. An edit answers old -> new, and
  when it takes effect (F5, `wpr 2`, or the next run).
- **Every write goes to `system-undo`.** A settings edit Wire-Pod wouldn't start with is put back
  at once, and `system-restart` refuses to restart on one.

### 14.2 The system menu

The speaker icon (🔇) opens it; 🔊 closes it. A Goverlord system bubble at full width holds the
menu bubble ([9.9](04_CHAT.md#99-the-system-window)). It leads you like a tour guide: groups in tour order,
one open at a time, each with its guide line and what's next. `run: true` items run on click; the
rest wait in the text box for your words. Under the menu are its results - each system line sent
while it's open and its answer (search results, picture rows, numpad status too).

The menu is `system_menu` in the settings. Its groups, in order:

| Group | Items |
|---|---|
| 1. Restart and check | `system-restart`, `system-skip`, `system-status`, `system-test-all`, `system-check` |
| 2. Play | `start interview`, `start bingo test`, `start baby brobots`, `start fart joke`, `system-play`, `replay`, `stop run` |
| 3. Edit words and settings | `system-topic`, `system-lingo`, `system-set`, `system-add`, `system-remove`, `system-set <song> tempo`, `system-set system_menu.look`, `system-menu-theme`, `system-set-list`, `system-undo` |
| 4. Find | `system-locate`, `system-find`, `system-score`, `system-locate-list`, `system-search`, `system-chats`, `system-chat-save` |
| 5. Apply | `system-voices-rebuild`, `system-restart` |
| 6. Switches | `system-numpad`, `system-numpad on`/`off`, `system-gap`, `system-tweaks on`/`off`, `system-native on`/`off` |
| 7. Brobots | `system-visual next`/`prev`, `system-cams`, `system-tpu`, `system-tpu restart`, `system-usb`, `system-brobot stats`, `system-version`, `system-brobot volume`/`eyes`/`locale`/`units`/`location`/`timezone`, `system-faces`, `system-photos`, `system-brobot cube`, `system-brobot beats`, `system-brobot beat`, `system-brobot anims sleep` |
| 8. Tune moves | `system-calibrate`, `system-tweaks on`, `system-anim list`, `system-anim` |
| 9. Test one thing | `system-test-list`, `system-test-<name>` |
| 10. Wrap up | `thanks brobots`, `stop session`, `system-help` |

**Its look** is `system_menu.look` (font_size, background, code_background, code_color,
results_background): `system-set system_menu.look.<name> <value>`, then F5. Presets:
`system-menu-theme <preset>` (`system_menu.themes`).

### 14.3 Editing from the chat

**`system-set`** - edits one value in place. Each file is edited in place - only that value,
everything else byte for byte - and read back before it's saved; `knobs.json` and `zKnobs.json`
both, when the step is in both. A combo step's parts are still edited by hand.
`gopod_system_set.go`, locked by `TestSystemSetEditsInPlace`, `TestSystemSetStoryText`,
`TestSystemSetReads`.

| Type | Does |
|---|---|
| `system-set` | the list: settings with their values, songs, step fields |
| `system-set master-volume 35%` | a named setting: `master-volume`, `normalize -3db`/`off`, `loop-volume`, `weather-volume`, `key-volume`, `sting-volume`, `run-sting-volume` |
| `system-set <path.to.setting> <value>` | any setting in the settings file (keeps the old value's kind) |
| `system-set 00_llm_warming` | a song's steps (folder name or `song_id`) |
| `system-set 00_llm_warming goverlord_warm_0101` | one step: its knobs and words |
| `system-set <song> <step> volume 13%` | one field: `volume`, `text`, `seconds`, `pause_seconds`, `hold_seconds`, `buffer_after`, `settle_seconds`, `speed`, `cycles`, `tempo_factor` |
| `system-set <song> <step> text New words.` | the step's `> TEXT:` in story.md (every line) |
| `system-set <song> tempo <s>` | the song's tempo (tempo-set) |

**More edits** (`gopod_system_edit.go`):

| Type | Does |
|---|---|
| `system-add <path> <item>` / `system-remove <path> <item>` | add or remove a list item, or `key = value` in a group |
| `system-lingo <brobots\|reporters\|butler> <word> = <said>` | set how a word is said in one lingo; `system-lingo <lingo> remove <word>`; the lingo alone lists it (Brobots in the template card, reporters `say_words`, butler `system_say_words`) |
| `system-topic` | a chapter topic's words, points (say lines and numbered points kept as one list), label, reminder, pictures |
| `system-undo` | puts back the last write (all the files it wrote), only if none has changed since; again for the one before. The last 20 are kept in `gopod_system_undo.json` (gitignored), so they survive `wpr 2`. `gopod_system_more.go`, `TestSystemUndo`. |

### 14.4 Finding things

| Type | Does |
|---|---|
| `system-locate master-volume` | where a system-set setting lives, as `file:line`, to open and edit by hand |
| `system-locate <song>` | a song's files |
| `system-locate <song> <step> [field]` | the step in `knobs.json`, `zKnobs.json`, and its words in story.md |
| `system-locate <word>` | every setting whose name holds the word, with its line and value (`volume`, `sting`, `awaits`) |
| `system-set-list` | every setting with its value, the songs, the step fields |
| `system-locate-list` | every setting's file:line, and each song's files |
| `system-find <word>` | where the word shows up: the log of the last 6 hours (the journal, read with fixed arguments; the word only filters), newest first, 15 lines; then the settings and each song's story.md / knobs.json, 10 lines each, with their line numbers (`gopod_system_find.go`, `TestSystemFind`). Example: `system-find FEED_CHECK`, `system-find Toronto`. |
| `system-score <song>` | prints a song's steps |
| `system-search <words>` | web search results in the terminal bubble ([8.8](04_CHAT.md#88-search)) |
| `system-chats <words> [operator\|brobots\|managers]` | searches every saved session, matches marked >>like this<<; skips the searches themselves |
| `system-chat-save [session]` | saves `chat.md` in the session folder |
| `system-help` | every system command, a line each |

`system-locate` is read only. `gopod_system_locate.go`, `TestSystemLocate`.

### 14.5 Status, checks and tests

**`system-status`:** Wire-Pod, the language model (loaded or cold), the warm-up, the chat page
(here/away), the run going, and each Brobot's last battery reading.

**`system-check`:** looks for every file the data names and every song step - no changes made.
- **What it checks:** every file named in the settings folder (`webroot/gopod_settings/`) (sounds, pictures, QR
  codes, workers, models, word lists), the weather faces and sounds, each chapter topic's
  `words:` and `pictures:`, `egg_names` and chapter topic lists against the topic files, every
  song's `story.md` steps against its `knobs.json` (and `zKnobs.json`), every file a song's knobs
  name, a warm-up song with no `_llm_song`, and topic words that overlap Wire-Pod intent phrases.
- **Where a file is looked for:** a page path (`/gopod_reporter_audio/x.wav`) under webroot; `~/`
  in home; an absolute path as is; a bare name in the chipper folder, webroot, the voice folder,
  the pictures folder or the song's own folder.
- **The answer:** "all good - N things checked", or each thing to fix (25 shown; the rest via
  `system-find GOPOD_CHECK`). Wire-Pod runs it at every start: `GOPOD_CHECK` lines in the
  journal. `gopod_system_check.go`, `TestSystemCheck`.

**`system-test-*`:** a full set of single tests, each checking one thing, nothing said or moved;
PASS/FAIL with what it saw. `gopod_system_tests.go`.
- `system-test-list` lists them; `system-test-<name>` runs one; several in a row run in turn;
  `system-test-all` runs the audit in `system_tests.audit`'s order.
- **The Torah's scrolls.** `GOPOD_TORAH.md` is the main scroll: the opening, a link and first line for each scripture, every commandment, test, run and look. Each scripture is its own scroll in `torah/` (`SCRIPTURE_01_RESTART_AND_CHECK.md` ...): its title, its plain paragraphs, its commandments, and links to the scriptures before and after and back to the Torah. GOPOD writes them all (`system-torah`) and drops a scroll the system no longer has; `system-check` and the build say when any is behind.
- **The results template.** Every system answer, in the chat's Terminal bubble or the system window's results, is drawn one way, left-aligned: its first line the title (bold), then its lines in order; a "<command>  <what it does>" line (two or more spaces between) lines up in a two-column table, the command in code style.
- **The Torah's scrolls**: `GOPOD_TORAH.md` is the main scroll: the opening, a link and first line for each scripture, every commandment, test, run and look. Each scripture is its own scroll in `torah/` (`SCRIPTURE_01_RESTART_AND_CHECK.md` ...): its title, its plain paragraphs, its commandments, and links to the scriptures before and after and back to the Torah. GOPOD writes them all (`system-torah`) and drops a scroll the system no longer has; `system-check` and the build say when any is behind.
- **Dial 0 for the system**: a `system-` command typed to anyone but 0 doesn't run - its line shows, then Goverlord's sting bubble: "Did you mean to dial '0' for the operator?" (`system_text.typed_system.dial_zero`). A click in the system window always goes to the system; stop run, start and replay go through whoever is picked.
- **Each process carries its own turn**: a chat line is stamped by the thread that posts it - the tour's, the comm check's, a question's own; the control lane's (a typed command, stop run, Esc) is no turn's (`gopodTurnOwn`). So stop run's own line goes through the whole template - shown, read out, heard - and stays, while the dropped run's last lines go with it. A line from any other thread (a song's, through Wire-Pod's API) is the turn going.
- **A system answer's text**: like every system bubble's text, 75% of its header, resting at 85% white (it has no voice to light it).
- **GOLDEN, the master chat queue**: everything the chat shows belongs to a turn, and turns run one at a time, each to its end (`gopod_chat_queue.go`). A question is a turn of its own when nothing else is going (the line queue: it, its answers, the page caught up). A run is a turn from check-in to check-out (a song, the scripted draft); the tour and the comm check are turns too. A turn's own lines - a song's from its runner, the comm check's KP1-9 questions, the tour's commandments - are its nested queue, stamped with its number (the feed's `turn`). The tour holds questions: one asked meanwhile waits for its end; a song's run rules keep chat questions out, and the comm check's questions are its own. The control lane never waits, not even for the warm-up: stop run, Esc, skip, `system-` commands. Stop run and Esc drop the turn going: Wire-Pod sends `drop:<turn>`, which the page acts on the moment it arrives - that turn's lines still waiting and any still coming never play, its bubble on screen finishes whole; the stop's own lines (the operator's `stop run`, then Goverlord's `session.stopped_text`) are no turn's. `TestMasterChatQueue` locks it.
- **Esc exits, skip skips**: Esc on the page exits the run going - a song, a draft, the tour, the comm check - as `stop run` does, and Goverlord says `session.stopped_text` (`/api-sdk/gopod_escape`, `GopodEscape`; a song run says it as it checks out); with nothing running (the opener's warm-up) Esc skips the waiting lines. `skip` (typed, the menu, numpad *, the mic) skips the voiced lines playing - in the tour, on to the next scripture.
- **0, the system**: the All menu's 0 (the admin page's only - visitor pages never show it) sends a typed line to the system itself: a system command runs; anything else gets the system's one-line pointer (`system_text.typed_system.not_a_command`). Every system command - `system-...`, `stop run`, `start <run>`, `stop session`, `replay <session>` - also runs as one whoever is picked, a persona key (3-9) included.
- **Stop run stops first**: with something going, a typed `stop run` stops it at once, then its line shows; with nothing going, the line shows and is read first, then "No run is going".
- **A lines bubble's lines are dormant from the start**: each line of a `lines:` bubble (the Torah tour's scriptures) is its voice's line the moment it's drawn - 0% until its voice plays, lit word by word, then 85%; a skip or a stop shows every line left whole (85%). A voiced line whose voice never comes is shown anyway after `voiceLitWaitMs` counted from its turn - once every line above it in its bubble has been said (`showOnItsTurn`), never from when the bubble showed.
- **The settings folder**: GOPOD's settings are one set of blocks kept in `webroot/gopod_settings/`, a short file per area - `00_about`, `01_brain` (the language model, answer shapes and chains), `02_cast` (personas, managers, nicknames, the roast dial, Doc's BS), `03_chat_page` (bubbles, highlight, the page's words and look), `04_game` (topics, value points, experiences, the interview), `05_voices_and_robots` (reporter voices, the numpad, Record Session, weather, sleep beats), `06_system` (the system window, the Torah, tests, timings, replies). Each block lives in exactly one file. Wire-Pod reads them as one, serves that whole to the page at `/gopod_brobots_settings.json`, and an edit (`system-set`, `system-add/remove`, the lingo, the voices menu, the menu look) goes to the file holding its block; `system-locate` and `system-find` answer file:line there (`gopod_settings_folder.go`). A block in two files stops Wire-Pod's start with the reason.
- **Global knobs, then the song, then the step**: every song starts from `goverlord/runtime/songs/global_knobs.json` - the default move values (`moves`: nod, arm cue, animation, emotion beat, wheel nudge, stir) and the song runner's shared waits (`runner`). A song's own `knobs.json` overrides them - its `moves` block for the whole song, any step's own value over that (`apply_knob_layers`, `knobs_envelope_001.py`). The hardware-tuned values (the nod speed floor, the settles after taking control) and the "apply phcal tweaks?" switch stay as they are.
- **Show timings live in the settings**: how long a warm-up, a recording, an answer or a call may take, and how often a watch looks again, are `timings` in the settings file, in seconds (`gopodTiming`, `gopod_timings.go`; `timings_about` names each). The short polling ticks and the hardware-tuned movement values (arm cue holds, the settle after taking control) stay in the code. `TestTimingsInSettings` fails if a timing the code asks for is missing or not more than 0.
- **System replies and help live in the settings**: `system-help` prints `system_help.lines` (a line each, the command, two or more spaces, what it does), and every system command's reply is a template in `system_text`, keyed by command and a few words, its `{names}` filled by the code (`gopodSystemText`, `gopod_system_text.go`). Edit the words, keep the `{names}`. `TestSystemTextKeys` fails if a reply's key is missing or its `{names}` don't match the code. Error passthroughs ("system-play: <error>") stay in the code.
- **Look tokens**: every colour the GOPOD pages use is a token, set once in the `:root` block at the top of `css/gopod_chat_bubbles.css` (the saved-session page, which stands alone, has its own in `gopod_session_page.css`); every rule uses a token (`var(--white)`, `var(--marquee-gold)`, ...). A value set live - a Brobot's eye colour, the system window's look - still overrides its token (`var(--brobot-1-color, var(--brobot-1-default))`). Change a colour once, there. `TestLookIsTokens` fails if a rule types a colour of its own.
- **The page runs from the settings**: the chat page keeps no value of its own for what the settings file holds - its timings and sizes (`chat_window`, `page_sync`), its roles and labels, and its own words (`page_text`: the typing row, the system window's panels, `{placeholders}` filled by the page). Each is read once the settings load, before the first bubble; `TestPageRunsFromSettings` fails if one is missing from the settings or written into the page as well.
- **Copy a result**: each answer in the system window's results has a copy button at its top right (`system_menu.copy`): its command and its answer, as plain text. At the very top of the results, with the first one, a copy-all button (`system_menu.copy.all`): every command and answer in the window, in order.
- **The results template**: every system answer, in the chat's Terminal bubble or the system window's results, is drawn one way, left-aligned: its first line the title (bold), then its lines in order; a "<command>  <what it does>" line (two or more spaces between) lines up in a two-column table, the command in code style.
- **The GOPOD Torah and its look.** `GOPOD_TORAH.md` (repo root) is the master reference for the
  system window - "the ancient scrolls Ron used to bring the Brobots to life", a fun take, not a
  religious one. GOPOD writes it from the running settings: the window's groups and items (with the
  run cards), every commandment (`system-help` - in the Torah, system commands are commandments), every test and the audit's order, the runs, the looks.
  `system-check` says when it's behind (a setting changed live); `system-torah` writes it again;
  `TestTorahMatches` holds it at every build. The system window's look is the `torah` theme: dark
  leather (`texture`), a gold double frame (`border`), gold group titles (`accent`), silver rules
  (`rule`), gold commands; `system-menu-theme <name>` switches back.
- `system-test-admin-run` (also in the audit and the system menu): the 100% admin run as a standing
  self-check. It plays `gopod_probes/admin_run/admin_run_001.txt` (`system_tests.admin_run.file`:
  part|to|line, 82 lines - the 3 songs and their before/after asks, every value point, every egg, and
  KP1-KP9 as held keys)
  in order through the game's own code, as the room would - no robots, no model, nothing said; the
  live game is saved first and put back after. It reports each point not revealed, each egg not
  found, each prize not given, each reminder that went dry (a topic question whose reminder names
  no topic or chapter, or an empty one, while points are left), each song line that isn't a run
  card's start, and each numpad key that doesn't reach its Brobot or persona (a persona's "who are
  you" gets its who line). PASS when there's none. `TestAdminRun`.
- **Double click a commandment**: anywhere on the admin page - a bubble, the system window, its results - a double click on a `system-...` word puts it in the text field (`system-torah`); nothing runs until Enter. A single click does what it always did.
- **The Torah tour** (`system-tour`, system menu "9. Test one thing"): Goverlord goes through the GOPOD
  Torah in the chat - the opening ("Goverlord's Torah: In the beginning there was a great restart."),
  each scripture in its own bubble - its title ("Scripture 1, Restart and check"), then its plain
  paragraphs (`system_menu.groups[].tour`, also shown in the Torah under its scripture); a group
  with none falls back to its guide, commandments and next line, in the Torah's words (`torah`). Each line is its own voice
  (the system bubble's `lines:` form): every line named at once, so each starts dormant, the first made,
  then each next one made while the one before plays (`gopodSayLinesInBubble`). A commandment on `system_tour.run` runs too and its answer shows; the rest -
  the restart, robot moves, switches, the session's end, the long tests, anything that waits for your
  words - is only said. Each line waits for the page to play it (when a page is here). `system-tour`
  again stops it; Record Session keeps it. The system window: all slots start closed, the results under
  the last; an opened slot moves to the top and the results follow to the line under it. No commandment is
  in the menu twice; a run card marked `pending` (the interview, for now) stays out of the menu, the Torah
  and the tour.
- `system-comm-check` (system menu "9. Test one thing"): the pre-show comm check. It plays
  `comm_check`'s rounds - each line on each key KP1-KP9 in turn, as if the key were held and the
  words spoken: the page draws the key's bubbles (keep holding, recording, transcribing), the line
  goes where the key goes, live (the model, the voices, the robots), and the next key waits for the
  answer to finish. Rounds: "Hey party animal. Who are you and what do you do?", "Tell me about
  yourself.", "Tell me all about yourself.", "Hey {name}, who all else is here with us?", and "Tell
  me about your interns." for CHALK and PLAYHEAD. The game goes back to 0 after, so the show starts
  clean; Record Session keeps it all. `TestCommCheck`.
- Tests: files, python, voices, llm, search, weather-live, weather-offline, time-place, takes, qr,
  personas, eggs, page, robots, tpu (`system-test-tpu`, in the audit), audio
  (`system-test-audio`).

### 14.6 Cameras and the Edge TPU

**GOPOD VISUAL CONTROL - the TPU +/- rotation.** `visual_sources`: Doc, Pip, the USB webcam.
- KP+ next feed, KP- previous, wrapping (on the page, numpad on); `system-visual [next|prev|<id>]`
  typed. With one feed nothing turns.
- Each turn grabs a frame (a Brobot woken through `gopodRobotCall`, or the webcam with ffmpeg; the
  webcam grab times out after 8 s), runs the Edge TPU quick look (objects + the boxes copy), and
  keeps `webroot/visual/<id>.jpg` (not in the repo). A source that can't give a frame shows its
  last one, CACHED. During a run (the run owns the Brobots), KP+/KP- and `system-visual` show the
  last frame, no wake. Frame URLs keep their `?t=` stamp.
- **One visual board:** FEED n/N, name, LIVE/CACHED, TPU objects, the frame (magnifier) - CHALK's
  board, or the menu's results while open; turned in place while it's the newest thing. One feed
  shows photo, enhancement, TPU boxes (no "I see").
- **All feeds** (`system-cams`, `?dir=all`): one row in the results - every feed's photo, then
  the selected one's TPU boxes (a 4th picture with a webcam plugged in). A single Brobot's photo
  shows photo, enhancement, TPU boxes. `system-cams [tpu]`.
- Route `/api-sdk/gopod_visual`, registered from `gopod_visual.go` (server.go untouched).

**The Edge TPU (GOLDEN):** plugged in and turned on, for photo face and object recognition. Setup
in place: the TPU driver, the `gopod-coral39` Python environment, and the face and object models
in `goverlord/runtime/data_gomad/models/edge_tpu/`. Turn it on after a replug or restart with the
command in `goverlord/runtime/data_gomad/models/README.md` (`1a6e:089a` before,
`18d1:9302 Google Inc.` once on). Wire-Pod's worker holds the stick while Wire-Pod runs (one
program at a time). `system-tpu [restart]`: `system-tpu restart` answers when the worker is ready
and can stop a worker still loading.

**USB:** `system-usb` / the menu's live USB panel (every 3 s, "+ plugged in" / "- pulled out",
`/api-sdk/gopod_usb`).

### 14.7 The Brobots from the chat

Wire-Pod's bot page as commands, through its own `/api-sdk` routes on this machine
(`gopod_system_robots.go`). `system-brobot` (`system-robot` still works) answers what it set
("location set to ...").

| Type | Does |
|---|---|
| `system-brobot <stats\|volume\|eyes\|locale\|units\|location\|timezone\|say\|wake\|head\|lift\|wheels\|camera> [1\|2\|both]` | one robot call; moves are: control, 1.5 s settle, a short nudge, stop, release |
| `system-brobot cube <name>` | a `native_cube` intent; control given back so his behaviour runs |
| `system-brobot anims <word>` | the robot's own animation list (`anims sleep` lists every sleep one) |
| `system-brobot anim <name>` | plays one animation |
| `system-brobot camera` | Wire-Pod's live `/cam-stream` as a picture in the chat |
| `system-brobot beats [family]` / `system-brobot beat <alias>` | the 126 sleep beats (`sleep_beats` in the settings) |
| `system-faces` | the robot's known faces |
| `system-photos` | the photos (thumbnails, magnifier) |
| `system-version` | Wire-Pod installed, newest, and which of GOPOD's nine overlay files changed upstream |
| `system-anim <text> [1\|2\|both]` | the animationWI tester (WIP), through Wire-Pod's say route like the song engine; `system-anim list` |

**Every robot call goes through `gopodRobotCall`:** one call at a time per Brobot (no overlapping
chains), the chat's hold let go, behaviour control taken and granted (8 s at most, else "didn't
wake"), the 1.5 s settle, the call run to its end while control is held, then given back. `say`
goes over the SDK and reports the robot's own finish state (Wire-Pod's say route always answers
"success").

### 14.8 Play and tune from the chat (pha0b and phcal)

All start `gopod_probes/tools/system_play_001.sh`, handed to the operator's user like the
song-start intents; the engine checks in with Goverlord (one run at a time).

| Type | Does |
|---|---|
| `system-play <song> [from <step>] [to <step>] [robot 1\|2] [dry]` | plays a song or a range (steps by id or number; golden-engine songs), like `pha0b` |
| `system-gap <seconds>\|off\|default` | the reporter-gap switch |
| `system-tweaks on\|off` | the apply-phcal-tweaks switch |
| `system-calibrate <move> <1\|2> [values] [dry]` | tunes one move (`phcal_isolate_001.py`) |
| `system-score <song>` | prints the score |
| `system-set <song> tempo <s>`, a step's `tempo_factor` | tempo (tempo-set) |
| `system-test-audio` | the audio test |

### 14.9 Restart and switches

| Type | Does |
|---|---|
| `system-restart [go]` | Wire-Pod restarts itself; the page reloads itself on the new start; click Log for sound |
| `system-native on\|off` | Wire-Pod's own answers on or off ([7.4](04_CHAT.md#74-native-wire-pod-answers)) |
| `system-numpad on\|off` | numpad push-to-talk on or off; `system-numpad` alone lists the keys |
| `system-skip` | skips the reporters' warm-up lines ([6.7](03_SESSION.md#67-skip)) |
| `system-voices-rebuild [go]` | rebuilds every voice ([10.2](05_VOICES_AND_SOUND.md#102-voice-files-made-once-play-many)) |
| `system-menu-theme <preset>` | the menu's look |

### 14.10 Every command at a glance

| Command | Section |
|---|---|
| `start interview` / `start bingo test` / `start baby brobots` / `start fart joke` | [6.2](03_SESSION.md#62-runs) |
| `stop run`, `stop session`, "stop nap" / "stop baby robots" | [6.6](03_SESSION.md#66-stopping-a-run) |
| `replay <name>` | [6.8](03_SESSION.md#68-replay-a-session) |
| "skip", Esc, numpad \* | [6.7](03_SESSION.md#67-skip) |
| "thanks brobots" / "thanks guys" | [6.10](03_SESSION.md#610-the-thanks-ending) |
| `search <words>` / `look up <words>` | [8.8](04_CHAT.md#88-search) |
| `system-add`, `system-remove`, `system-lingo`, `system-set`, `system-set-list`, `system-topic`, `system-undo`, `system-menu-theme` | [14.3](#143-editing-from-the-chat) |
| `system-locate`, `system-locate-list`, `system-find`, `system-score`, `system-search`, `system-chats`, `system-chat-save`, `system-help` | [14.4](#144-finding-things) |
| `system-status`, `system-check`, `system-test-list`, `system-test-<name>`, `system-test-all`, `system-test-tpu`, `system-test-audio` | [14.5](#145-status-checks-and-tests) |
| `system-visual`, `system-cams`, `system-tpu`, `system-usb` | [14.6](#146-cameras-and-the-edge-tpu) |
| `system-brobot` / `system-robot`, `system-faces`, `system-photos`, `system-version`, `system-anim` | [14.7](#147-the-brobots-from-the-chat) |
| `system-play`, `system-gap`, `system-tweaks`, `system-calibrate` | [14.8](#148-play-and-tune-from-the-chat-pha0b-and-phcal) |
| `system-restart`, `system-native`, `system-numpad`, `system-skip`, `system-voices-rebuild` | [14.9](#149-restart-and-switches) |

---

## 15. Your stuff, no code

Everything below is data - no code is touched. After any of it, type **`system-check`** on the
chat keyboard: it looks for every file the settings, chapters and songs name, and every song step
(Wire-Pod also runs it at start - `GOPOD_CHECK` lines in the journal).

| To... | Do this | Then |
|---|---|---|
| **Change a setting** (volume, waits, texts) | `system-set <setting> <value>` on the chat keyboard, or edit `webroot/gopod_settings/` (`system-locate <word>` finds it) | page: F5; Wire-Pod: `wpr 2` |
| **Add a picture to an egg** | put it in `webroot/pictures/`, add its name to the topic file's `- **pictures:**` line (`gopod_probes/chapters/<chapter>/<topic>.md`) | `wpr 2` |
| **Add a topic** (new egg words, value points) | a new `NN_name.md` topic-cluster file in its chapter folder, from the template (`gopod_probes/_topic_cluster/`): `# label`, then `- **words:**`, `- **pictures:**` (optional), `- **reminder:**`, `- **say 1:**` ..., then `---` and its reveal text | `wpr 2` |
| **Add a chapter** (a new egg) | a new folder `NN_name/` in `gopod_probes/chapters/` with `chapter.md` (`# title` = the egg's title, `- **summary:**`) and its topic files; the number sets its order | `wpr 2` |
| **Change a song line** | its `## STEP` / `> TEXT:` in the song's `story.md`, or `system-set <song> <step> text <words>` | next run |
| **Add a song step** | a `## STEP <id>` + `> TEXT:` in `story.md` **and** a step with the same `step_id` in `knobs.json` (`system-check` says if one is missing) | next run |
| **Add a sound** (sting, song sound, soundtrack) | put the `.wav` in `webroot/gopod_reporter_audio/`; name it in the settings (a sting) or a song's `knobs.json` | evened to the loudness rule by itself |
| **Change how a word is said** | the pronunciation registry in `brobots_wirepod_interview_section_card_template_1_001.md` (Brobots), `reporter_voices.say_words` (reporters), `system_say_words` (Goverlord), or `system-lingo` | `wpr 2` |
| **Add a QR code** | the picture in `webroot/qr/`, an entry in `qr_codes.codes` (words that show it) | F5 |
| **Video** | not supported yet in the chat | - |

---

## 16. Settings file reference

All in `goverlord/wire_pod_overlay/chipper/webroot/gopod_settings/ (a short file per area, read as one)`. Values are
data; the code only guides the flow. Wire-Pod checks every key at start.

| Block / key | What it holds |
|---|---|
| `llm` | The warm-up: `voices_text`, `warmup_text`, `wait_text`, `ready_text`, `warmup_done_text`, `run_ready_text`, `run_ready_end`, `no_voices_text`, `no_llm_text`, `opener_song_id`, `warmup_check_seconds`; the keep-loaded refresh. |
| `session` | `page_present_seconds` (20), `awaits_text`. |
| `chat_moves` | The thinking moves; `thinking_text` ("Got it!"), `searching_min_seconds`. |
| `chat_window` | The page's timings and sizes (baton, scroll, countdown hold, marquee, narrator fades, reveal fallbacks); `speaking_max_ms` (30 s). |
| `page_sync` | `reply_wait_seconds` (10), `held_log_seconds` (3). |
| `word_reveal` | `seconds_per_word` (0.38), `seconds_per_mark`, `start_delay_words`. |
| `digit_highlight` | `seconds_per_number`. |
| `highlight_scroll` | `dead_zone`. |
| `live_split` | `thought_lines` (on), `join_up_to_words` (0). |
| `reporter_voices` | `python`, `worker`, `speed`, `operator_voice`, `unassigned_voice_host` (CAMEO), `voice_tweaks`, `say_words`, `generating_text`, `generating_marker`, `start_wait_seconds`, `heard_margin_seconds` (3), `finish_wait_seconds` (120), `max_wait_seconds`, `narrator_lead_seconds`, `operator_keep_days`, `scripted_keep_days`, `live_keep_days`. |
| `reporter_voices.system_stings` | `default` (dJent3b), and each sting's volume share. |
| `reporter_voices.system_persona`, `reporter_voices.system_say_words` | The butler's persona and phonetics (`system_persona`, `system_say_words`). |
| `chat_audio` | `volume` (the master, 35%), `loudness_lufs` (-16), `normalize_db` (-3). |
| `background_loop` | The loop file and `volume` (14%). |
| `weather_card` | The weather faces; `volume`. |
| `live_weather` | `offline_text`, `follow_up_kit_line`. |
| `numpad_ptt` | The numpad map, `all_hold_seconds` (0.5), `key_sound` (off), `error_sound` (15%), `error_texts`, `texts`, `mic_usb_ids`, `numpad_usb_ids`, `usb_watch_seconds`, the missing/found texts. |
| `native` | `on_at_start`, `always_on` (`intent_names_ask`). |
| `native_cube` | `intents`: the GOPOD commands list. |
| `native_intent` | The native intent notice: `once_seconds` (its words are in `bubble_notices`). |
| `bubble_notices` | What goes in a notice in a Brobot's bubble: `value_points` (`title`, `parts` order, `points`, `weather`, `search`, `none`, `got_it`), `egg`, `native_intent` (`title`, `line`). |
| `addressing` | `follow_up_note`, `all_next_wait_seconds`. |
| `topic_pictures` | `max_per_session` (5), `popup_text`, `egg_text`, `egg_zoom_text`, `egg_names`, `describe`, `describe_note`, `describe_wait_seconds` (30), `describe_on_ask`, `describe_wait_text`, `describe_done_text`, `not_listed`, `not_eggs`. |
| `topic_pictures.cards` | Egg cards (`cards.halloween`, the Roast Battle card). |
| `value_points` | `reminder_eggs`. |
| `math` | `trigger`, `fraction_words`, `lines`, `flavour_sentences`, `long_division_max_steps`, `egg_name`, `popup_title`, `follow_up_note`, `decimal_follow_up_line`. |
| `photo` | `vision_model` (moondream), `awake_seconds` (2.5), `quick_look` (`pictures`, `enhance`), `quick_text`, `gathering_text`, `stall_lines`, `detail_strip_prefixes`, `egg_name`. |
| `clean_politics` | `words`, `allowed_topics`. |
| `qr_codes` | `codes`, `second`. |
| `host_personas` | The four hosts' one persona shape: `name`, `role`, `steps`, `signature`, `handover` (`line`, `nicknames`). |
| `host_pairs` | Doc & Pip, CHALK & PLAYHEAD: `hosts`, `left`, `dial` (0-100, the stereo dial), `voice`, `partner_as`, `banter`. |
| `managers` | `personas` (CHALK, PLAYHEAD: `prompt`, `who_words`, `who_line`), `time_place_text`, `display_host`, `qr_note`, `qr_display_text`. |
| `system_search` | The search settings; `results_role`, `results_host` (CACHE PYC). |
| `system_set` | `role` (the terminal bubble). |
| `system_menu` | The groups; `look`; `themes`. |
| `system_tests` | `audit`. |
| `visual_sources` | Doc, Pip, the USB webcam. |
| `sleep_beats` | The 126 sleep beats. |
| `battery_watch` | `lines`. |
| `sign_off` | `safe_spot_seconds`. |
| `interview_script` | The interview: `llm_ready_text`, `start_lines`, and its song folder (`interview_script.song_dir`). |
| `record_session` | `max_minutes`, `answer_wait_minutes`. |
| `page_sound_files` | Song sound files copied to the page at start. |

---

---

[Back to ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md)

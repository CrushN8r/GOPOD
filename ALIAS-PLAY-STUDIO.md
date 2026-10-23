# ALIAS-PLAY-STUDIO

> The reference manual for GOPOD's show: how two Vectors learn a song, how a session and its
> runs work, and how the Brobots chat page behaves, rule by rule. Take Out part 1 in
> [WIRED-POD.md](WIRED-POD.md).

**What this is.** This is GOPOD's single source of truth for behaviour. Every song rule, every
chat-page rule, every command and setting is written here once, with the code that does it and
the test that locks it. If the code, a song folder and this manual disagree, fix one of them.

**Who it is for.** A Wire-Pod owner reading the repo, the owner running the show, and an AI
assistant working on the code. You do not need to know the code to use it: start with
[Concepts and terms](#1-concepts-and-terms).

**Where things live.** Values are in `gopod_brobots_settings.json` (each block's `about` says
what it tunes). Go files are in `goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/`. The chat
page is `webroot/js/gopod_brobots_chat.js` with `gopod_chat_bubbles.css`. Songs are in
`goverlord/runtime/songs/`.

**Tags used here.**
- **GOLDEN** - proven live and locked. Change it only with the owner's OK.
- **Locked by `TestName`** - a test fails the deploy if this behaviour changes.
- **WIP** - being worked on.
- **Draft for signature** - written up, waiting for the owner's sign-off.

---

## Contents

1. [Concepts and terms](#1-concepts-and-terms)
2. [A song folder](#2-a-song-folder)
3. [The notes](#3-the-notes)
4. [The toolbox and the songs](#4-the-toolbox-and-the-songs)
5. [Making and editing a song](#5-making-and-editing-a-song)
6. [The session and runs](#6-the-session-and-runs)
7. [Ways in: talking to the Brobots](#7-ways-in-talking-to-the-brobots)
8. [A live reply](#8-a-live-reply)
9. [The chat page and its bubbles](#9-the-chat-page-and-its-bubbles)
10. [Voices](#10-voices)
11. [Volume and sound](#11-volume-and-sound)
12. [Managers: CHALK and PLAYHEAD](#12-managers-chalk-and-playhead)
13. [Eggs](#13-eggs)
14. [The system menu and commands](#14-the-system-menu-and-commands)
15. [Your stuff, no code](#15-your-stuff-no-code)
16. [Settings file reference](#16-settings-file-reference)
17. [Golden rules](#17-golden-rules)
18. [Troubleshooting](#18-troubleshooting)
19. [Tests that lock behaviour](#19-tests-that-lock-behaviour)

---

## 1. Concepts and terms

### 1.1 The cast

| Name | Who or what |
|---|---|
| **Brobot 1** | The first Vector robot, "Doc", the host (`brobot_1`). Says metric units and 24-hour time. |
| **Brobot 2** | The second Vector robot, "Pip", the brat (`brobot_2`). Says imperial units and 12-hour time. Keeps the cube. |
| **The Brobots** | Both robots together. |
| **Brobot 3 / Reporter 1** | A voice-only reporter played on the chat page (voice `af_heart`). |
| **Brobot 4 / Reporter 2** | The second voice-only reporter (voice `am_michael`). |
| **Brobot 0** | The CrushN8r topic: the owner's own topic, found by the word "crushn8r". |
| **Goverlord** | The system. It runs the session ("GOPOD Brain"), speaks the system lines in its own voice (`bm_george`, "the digital butler"), and owns the grey system bubbles. |
| **Operator** | You, the person at the keyboard, numpad or mic. Your lines show in **Operator bubbles**, read out in the Operator voice (`af_sky`). |
| **CHALK** | A manager: the display host (CHALK BORED, numpad key KP7). Reporter 1's voice, chalkboard bubbles. See [Managers](#12-managers-chalk-and-playhead). |
| **PLAYHEAD** | A manager: the play, time and place host (PLAYHEAD ZUMINTU, numpad key KP8). Reporter 2's voice, tape-brown bubbles. |
| **Gomad** | A manager persona (CHALK or PLAYHEAD) as a chat participant. |
| **Gominion** | An output bubble type named with Goverlord and Gomad. This manual has no further rules for it yet. |

Bubble types: **Operator** is input; **Goverlord**, **Gomad** and **Gominion** are output.

### 1.2 The three words

| Word | What it is | In a song folder |
|---|---|---|
| **Golden note** | The smallest move or line the robots can play, proven live on its own first. | One step's `note` type in `knobs.json` (`arm_cue`, `rattle`, `say_turn`, ...). |
| **Sequence** | A few notes in order that make one moment. | One `## SECTION` in `story.md`, the steps' `section` in `knobs.json`. |
| **Golden sequence** | A sequence, or a whole song, that's locked: it plays the same way every time and is ready to record. | A song whose status line says `GOLDEN`, or a locked opener (the Bingo test's `OPENING SYNC`). |

The **piano** is the notes. The **mixer** is how each note plays (its hold, speed, buffer,
tempo). The **sequencer** is their order. All three live in the song folder: no song's moves are
hidden in code.

### 1.3 On stage: the chat page

| Say | It means | Where it lives |
|---|---|---|
| **Operator bubble** | A line from you to the Brobots, white-eyed icon, labelled by its input: "Keyboard to Brobot 1", "KP1 to Brobot 1", "Numpad to <who>", "Backpack to Brobot 2" (heard by the robot himself: Hey Vector or the backpack button), "Mic to Brobot N", "Reporter 1 to Brobot 1" (a scripted run's question), "… to Brobots" (both icons). It goes up once the bubble on screen has finished (the baton). | the chat |
| **System bubble** | Goverlord's grey-eyed bubbles (warm-up, countdowns, chapters, system messages). Reporter and narrator bubbles are system bubbles with white eyes. | the chat |
| **System llm bubble** | The one warm-up bubble a run opens with ("Goverlord: GOPOD Brain", LEDs, hourglass). | `llm_warmup` |
| **System message** | A Goverlord line in a grey bubble, said in the system voice ("Brobots. Go to sleep now please.", "Run complete."). | note `system_line` |
| **Goverlord's terminal bubble** | Under Goverlord (2026-10-22): where system commands answer - Goverlord's header, text at 50% size (`system_set.role`, the role's name inside only). | system- commands |
| **Brobot bubble** | A Brobot's line, in his eye colours, each piece lit as it's said. | note `say_turn`, a reply |
| **Reporter bubble** | CHALK or PLAYHEAD's scripted voice (Reporter 1 / 2 in scripts), headed by the header template (`display_names`: CHALK / PLAYHEAD), white-eyed icons. | the interview, `ask` |
| **Narrator bubble** | A timed song's captions, one bubble per section, each phrase lit as it's read. | `robots_speak: false` |
| **Manager bubble** | CHALK's chalkboard or PLAYHEAD's tape-brown bubble, or their part inside a Brobot's bubble. | `managers` |
| **Egg bubble** | A found easter egg, ringed by marquee bulbs. | [Eggs](#13-eggs) |
| **Chapter** | Goverlord's title between parts of the interview. | the interview |
| **Warm-up bubble** | "Goverlord: GOPOD Brain", the LEDs and hourglass, then the ready message. | `llm_warmup` |
| **Countdown** | The ⏳ bubble with juggled balls, under a rattle. | note `rattle` |
| **Big ball** | The Bingo ball bubble. | `show_ball` |
| **Highlight** | The lit piece in a bubble: the words being said or sung now. | [9.5](#95-highlight-and-word-reveal) |
| **The baton** | The next bubble waits hidden until the one on screen finishes. | [9.3](#93-the-chat-rules) |
| **Operator line** | `Brobots ()1 ()2 ()All [Type to the Brobot, press Enter] \| Record? [] \| HH:MM:SS 🔇` under the chat. | the chat |
| **System menu** | The menu that opens from the speaker icon (🔇/🔊), listing every system command in tour order. | [14](#14-the-system-menu-and-commands) |

### 1.4 The show

| Say | It means |
|---|---|
| **Session** | `wpr 2` until the next `wpr 2`. Goverlord runs it. |
| **Run** | One song, draft or phcal live, from its start to its check-out. One at a time. |
| **`start <song>`** | Starts a run: `start bingo test`, `start baby brobots`, `start interview`, `start fart joke`. |
| **`stop run`** | Stops the run at its next safe spot ("Run stopped."). |
| **Thanks** | The thanks ending: closes the run out ("thanks brobots", "thanks guys"). |
| **Sleep call / wake call** | Both Brobots put to sleep (`sleep_both`) / woken (`wake_both`, or the wake inside `brobots_ready_together`). |
| **Cube call** | The cube blip on a Brobot's own cube: red, 2 s, green. Brobot 2 keeps the cube. |
| **Brobots Ready** | Both Brobots say the line together (`brobots_ready_together`). |
| **Last words** | The song's last sung or spoken words; the ending is timed from them. |
| **Edit mark** | A 0-second `pause`: a place for the reporters' voiceover in the video edit. |
| **Live / dry** | With the robots / with no robots (nothing moves, everything is logged). |
| **Check run / check logs** | Read the last run's log and Wire-Pod's journal; report the bottom line. |
| **Bench test** | One move or tool tried on its own, on the bench (phcal or a one-off script). |
| **Egg** | An easter egg: a prize (pictures or a card) that pops up when a line's words find it. |
| **`_llm_song`** | The fixed-name folder inside every warm-up song that holds the lines said between "Voices ready" and "LLM ready". |

### 1.5 The workbench

| Say | It means |
|---|---|
| **Golden** | Proven live and locked. Changed only on the owner's go. |
| **WIP** | Being worked on. |
| **Lock it in** | Mark it golden and commit it. |
| **Commit / push** | Save to the repo / send it to GitHub. Two separate go's. |
| **`wpr 2`, F5** | Restart Wire-Pod (picks up a deploy), reload the chat page. |
| **Desk start** | `source ~/.bashrc; gorepo; lsusb; wpr 2`, then F5 (below). |
| **`pha0b`** | Plays a range of a song's steps, live or dry. |
| **`score`** | Prints a song's steps, on screen. |
| **phcal** | The bench: tune one move or sound on one robot. |

### 1.6 Desk start

Your terminal routine: `source ~/.bashrc; gorepo; lsusb; wpr 2`, then F5.

- `source` reads only the first file it's given (a second becomes an argument). `~/.bashrc`
  already loads `~/.bash_aliases`, so `source ~/.bashrc` alone does both.
- `gorepo`: `cd ~/crushn8r_git/GOPOD`.
- `lsusb` - the rig's parts, by name (bus/device numbers drift, names don't): **TTGK
  Technology USB Audio** (3302:12e9, the USB mic), **China Resource Semico ... USB Keyboard**
  (1a2c:4324, the numpad), **Novatel Wireless M3000** (the mobile hotspot), Realtek Bluetooth
  Radio and 802.11ac NIC (Wi-Fi), the VIA / Huasheng USB hubs. Mic or numpad missing here = the
  chat's mic or numpad won't work.
- `wpr 2`: sets the audio routing (`AUDIO_ROUTING_SET speakers=Built-in Audio Analog Stereo
  mic=USB Audio Mono`), then restarts Wire-Pod (asks for sudo; `WPR_RESULT status=FORCED`,
  "wire-pod restarted <time>"). A deploy made after that time needs another `wpr 2`.
- F5 on the chat page: the page clears the chat for the new start and reloads once by itself.
- Then click **Log** on Wire-Pod's web page. Nothing plays until you do (a browser plays sound
  only after a click).

---

## 2. A song folder

### 2.1 The folder and its files

A song is a folder: `goverlord/runtime/songs/<number>_<name>/`. Copy
[`_new_song/`](goverlord/runtime/songs/_new_song/) to start one. A folder whose name starts with
`_` is a template, never a song.

**A song is known by its `song_id`, never its folder.** The `song_id` at the top of its
`knobs.json` is its name everywhere: the engine, the `start …` commands, `pha0b`, `score`, the
interview, and the soundtrack copied to the page. Rename, renumber or reshuffle the folders and
nothing else needs updating. Two folders with the same `song_id` is an error, named.
`python3 goverlord/runtime/songs/tools/song_folders_001.py` lists them.

| File | What it holds |
|---|---|
| `story.md` | The words. A status line at the top, then `## SECTION: <name>` blocks, each with `## STEP <step_id>` and `> TEXT: <words>`. A step with no words keeps an empty `> TEXT:`. |
| `knobs.json` | What each step does, in order: its `note`, who plays it, its timing. This is what plays. |
| `zKnobs.json` | Your private working copy of `knobs.json`. When it's there, it wins. Not tracked. |
| `_llm_song/` | The warm-up's in-between run (a `knobs.json` and a `story.md`). See [5.5](#55-the-_llm_song-run). |
| `runs/`, `notation/` | What runs write: logs and marked-up scores. Not tracked. |

**The status line** is the first line after the title, for example
`**Status:** GOLDEN (2026-10-01)` or `**Status:** WIP (2026-10-09)`. GOLDEN means
live-confirmed and ready to record.

**The two files always name the same steps.** Changing a line's words in `story.md` is enough
on its own. Adding, removing or renaming a step means both files (and `zKnobs.json` when it's
there). The deploy's tests stop on any step one file has and the other doesn't.

### 2.2 Step names

A **step id** (`step_id`) is a step's name. It is **a label for people**: the engine never reads
it. What a step plays, and in what order, is its `knobs.json` entry (`note`, or a `combo` list
of parts). Renaming a step changes nothing it plays; the two files only have to name it the same
(locked by `TestSongFoldersFollowTheStandard`).

Write the name to match the knobs, in the knobs' order:

`<who>_<part>_<part>…_<SSNN>` - who, then each part as it plays, then section `SS` and step `NN`.

Examples: `host_arm_cue_0102` (Brobot 1), `brat_nod_kg_success_0203` (Brobot 2, a combo),
`reporter_gap_0301` (an edit mark).

**Who:**

| Who | Is |
|---|---|
| `host` | Brobot 1 (`brobot_1`) |
| `brat` | Brobot 2 (`brobot_2`) |
| `brobots` | both (`wake`, `ready`, `sleep`, `exit`) |
| `reporter` | an edit mark for the reporters (`reporter_gap`) |
| `goverlord` | Goverlord, the system voice (the LLM warm notes) |

**Parts (the master list):**

| Part in the name | Plays (`knobs.json`) | Example |
|---|---|---|
| `say` | `say_turn` - the line; or, when another part says the line (a beat, a nod) and there's no `say_turn`, that part | `brat_say_0204` |
| `arm_cue` | `arm_cue` - the arm lift, with the line | `host_arm_cue_0102` |
| `nod` | `nod` - a head nod, with the line | `brat_nod_kg_success_0203` |
| `<token>_animation` | `animation` with `animation_token` (`searching`…), the line with it | `host_searching_animation_0202` |
| `anim-<token>` | `animation` with `animation_token` (short form of `<token>_animation`) | `host_anim-kgsuccess_0202` |
| `kg_success` | `animation` with `kgSuccess` (usually the "Lol!") | `host_nod_kg_success_0601` |
| `beat_<token>` | `emotion_beat` with `animation_token` (`angry`, `frustrated`…) | `brat_beat_angry_0503` |
| `rattle` | `rattle` - the line, then the rattle under the countdown | `host_rattle_0102` |
| `thinking` | `thinking` - the live chat's own thinking process (arm, nod, the thinking line, searching, get-out) | `host_thinking_say_0202` |
| `wav-<name>` (older: a plain word, `fart2`) | `sound` with `sound` (`fart2.wav`) | `brat_say_fart2_0204` |
| `danger` / `grunge` / `djent3b` | `sound` - Wire-Pod's shared `sounds/` (`danger-will-robinson.wav`, `grunge.wav`, `dJent3b.wav`) | `host_say_danger_beat_frustrated_0205` |
| `img-<name>` | `picture` with `picture` (`granny.jpg`) | `host_say_img-granny_0301` |
| `gap` | `pause` (0 s = an edit mark) | `reporter_gap_0101` |
| `timer` | `timer` - a timed wait, `seconds` (a `pause`) | `host_timer_0201` |
| `sequence` | `sequence` - another song's steps, in place, by its `song_id` (a song can't sequence itself; an LLM warm song is skipped, played by Wire-Pod) | `brobots_sequence_0301` |
| `wake` / `ready` / `sleep` / `exit` | `wake_both` / `brobots_ready_together` / `sleep_both` / `exit` | `brobots_wake_0101` |
| `system` / `sleep_message` | `system_line` - a Goverlord line | `sleep_message` |
| `soft_wake` | `wake_sleepy` | `soft_wake` |
| `line_NN` | a timed caption (`pause` with `start_seconds`) | `line_30` (nap) |
| `warm` | `warm_message` - the System llm bubble's message, shown and said; `"wait_voices": true` waits for the voice files | `goverlord_warm_0101` |
| `warm_wait` | `warm_wait` - waits until the language model is loaded | `goverlord_warm_wait_0103` |
| `llm_song` | `llm_song` - the folder's `_llm_song` run, the lines between "Voices ready" and "LLM ready" | `llm_song_0104` |
| `warm_ready` | `warm_ready` - the ready message, shown and said | `goverlord_warm_ready_0105` |
| `sting` | `sting` - the system sting (`system_stings.default`) | `goverlord_sting_0106` |

The `warm`, `warm_wait`, `llm_song`, `warm_ready` and `sting` notes belong to an LLM warm song
(`song_type: "llm_warm"`), played by Wire-Pod (`gopod_opener_song.go`). `timer`, `sequence` and
`anim-` are the song engine's (`run_golden_song_001.py`).

Example: `host_wav-danger_beat_frustrated_0205` reads, in order: `sound` `danger…wav` to its end,
then `emotion_beat` `frustrated` with the line - exactly what its knobs play.

**The name check.** Every song run prints `STEP_NAME_WARNING` for a `_SSNN` name that doesn't
read as what its knobs play, or that has a word not on the master list (a typo, "sya"). Example:
`host_say_danger_beat_frustrated_0205: the name reads say_turn + danger? + emotion_beat:frustrated,
its knobs play sound:danger-will-robinson.wav + emotion_beat:frustrated - 'danger' isn't on the
master list`. It only warns; the run plays its knobs as ever. Names without `_SSNN` (the nap's
`line_30`, `soft_wake`) are free labels and aren't read. Locked by
`test_golden_song_step_names_001.py`.

### 2.3 `knobs.json`, the top keys

| Key | What it does |
|---|---|
| `song_id` | The song's name to the engine (it must match the engine's song list). |
| `song_type` | What kind of song: `bingo_capture`, `lyric_video`, `interview`, `llm_warm`. |
| `story_ref` | `story.md`. |
| `global_tempo` | Seconds of space added after each step, times its `tempo_factor`. |
| `chat_intro` | The operator bubble a run opens with ("Bingo Test Live"). |
| `sentence_split` | `scripted`: lines are said piece by piece and the chat bubble lights each piece. |
| `touch_interrupt_off` | `true`: a touch doesn't cut the run short. Back on when it ends. |
| `rattle_on` | Where the rattle plays: `page` (the chat page, like the reporter voices), `speaker` (this machine), `robot` (Brobot 1). |
| `timeline` | `srt`: each step plays at its `start_seconds` (from a subtitle file), counted from the first timed step, instead of right after the step before it. A step that's late goes at once. |
| `robots_speak` | `false`: the robots say nothing; the timed steps' words in `story.md` are a narrator's captions, one Narrator bubble per section, each phrase lit as it's read. |
| `sleepy` | `true`: after `sleep_both` the Brobots sleep. A `stir` (or a move) wakes its Brobot, and he drifts back to sleep, on his own time (the music and captions never wait). They're asleep before `brobots_ready_together`, which wakes both. |
| `wake_clip` | The clip `wake_sleepy` uses (for example the slower `anim_gotosleep_wakeup_01` instead of `GoToSleepOff`). |
| `media` | The song's video, in a private folder that's never tracked. |
| `audio` | The song's soundtrack, played on the chat page from the first timed line. The file sits in the private media folder; Wire-Pod copies it into the page's voice folder at its start (`page_sound_files` in the settings). |
| `sounds` | The song's own sound files, copied to the chat page at Wire-Pod's start. |
| `section_pictures` | Section -> a file in the page's `pictures/`: that picture goes in the section's narrator bubble, with the lit words, with the magnifier. |
| `llm_warmup` / `llm_ready_text` | Show the LLM warm-up bubble at the start, and the song's own ready line ("LLM Ready for Bingo"), said by Goverlord too. `"llm_warmup": false` opts a song out of the warm-up. |
| `warmup_lines` | Reporter lines for the warm-up, read only when the folder has no `_llm_song`. |
| `warmup_after` | A step id: the warm-up runs right after that step instead of at the start. No song uses it now. |
| `voices_required` | `false`: the warming message shows `llm.no_voices_text` ("No Voice Generation Required") in place of its voice-files part. |
| `llm_required` | `false`: shows `llm.no_llm_text` ("No Language Model Required") in place of the language-model part and of "Language Model Loaded, Master CrushN8r.", and calls no model. |
| `fatal_notes` | Notes whose failure stops the run (`emotion_beat`, `wake_both`). |

### 2.4 `knobs.json`, each step

Every step has `step_id`, `note`, `speaker` (`brobot_1` or `brobot_2`) and `section`. Then,
when the note uses them:

| Field | Meaning |
|---|---|
| `hold_seconds` | How long a move or animation holds. |
| `speed` | Move speed. |
| `cycles` | How many times a move repeats. |
| `settle_seconds` | Wait after a wake or control. |
| `buffer_after` | Seconds after the step. |
| `tempo_factor` | This step's share of `global_tempo`. |
| `pause_seconds`, `gap_label` | A `pause`'s length and label. |
| `seconds` | A `timer`'s length. |
| `start_seconds` / `end_seconds` | A timed song's times. |
| `animation_token` | Which animation (see [the toolbox](#41-shared-golden-pieces)). |
| `show_ball` | "B-1": the step shows a big Bingo ball in the chat as it starts. |
| `rattle_before` | Play the rattle before the line. |
| `pronounce` | `true`: a together line is said with the robots' pronunciation list ("GOPOD" as "Gowp-awd"). |
| `timeline_after` | On a `pause` in a timed song: the timeline starts when the pause ends, so the first narrator bubble is sent early, during it. |
| `sound`, `volume` | A `sound` step's file, and its volume (`""` or `"default"` = the master, or a percentage of it). |
| `picture` | A `picture` step's file or list. |
| `combo` | A list of parts, each a note with its own fields. |
| `text` | On a combo part: the words this part says, cut from the step's TEXT. |
| `piece_offset` | On a combo part: where its pieces start in the shared bubble. |
| `wait_spoken`, `spoken_wait_seconds` | On a `system_line`: go on only once it's heard. |
| `look_seconds` | On `sleep_both`: how often to check for an awake Brobot. |
| `cube_connect`, `cube_ready`, `cube_after_wake` | Brobot 2's cube (see the notes). |
| `wait_voices` | On a `warm` step: wait for the voice files. |

---

## 3. The notes

A **note** is what a step plays. These are the piano's keys.

### 3.1 Played by the song engine

The engine is `goverlord/runtime/songs/tools/run_golden_song_001.py`. `pha0b` plays any range of
steps, live or dry.

| Note | What it plays | Fields | Used by |
|---|---|---|---|
| `wake_both` | Wakes both robots and takes control for the run. | `settle_seconds`, `cube_connect` (`brobot_2`: his cube connects over Bluetooth in the background, about 18 s, while pauses and Goverlord's lines go on; the next step that moves a robot waits for it, so a later `cube_after_wake` lights at once) | Bingo test, nap |
| `sleep_both` | Both Brobots fall asleep (phcal's sleep tool). Every `look_seconds`, a Brobot found awake is put back to sleep. | `settle_seconds`, `look_seconds` | nap |
| `stir` | A sleeping Brobot wakes a moment, then drifts back to sleep. Nothing else. | `hold_seconds` (awake) | engine |
| `all_asleep` | Waits until every stir is done and both Brobots are asleep again. | | engine |
| `wake_sleepy` | The soft wake: both sleeping Brobots wake together, at the step's second, with the song's `wake_clip`. | `cube_ready` (`brobot_2`: his cube connects in the background `cube_ready_after_seconds` after the wake (8; a connect right after a wake fails) and is held; a later `cube_after_wake` lights it at once) | nap |
| `brobots_ready_together` | Both say the line at the same moment. | `rattle_before`, `pronounce`, `cube_after_wake` (`brobot_2`: his cube lights after the wake, before the line; the cube wakes a sleeping Brobot) | Bingo test, nap |
| `say_turn` | One robot says the step's line. | | Bingo test, nap |
| `say` | One robot says it (the control songs' form). | | engine |
| `animation` | One animation, with the line said with it. | `animation_token`, `hold_seconds` | Bingo test, nap |
| `emotion_beat` | A line, then its animation. If it fails, the run stops there (the songs' `fatal_notes`, with `wake_both`). | `animation_token`, `hold_seconds` | Bingo test, nap |
| `arm_cue` | The arm lift, with the line said during it. | `cycles`, `hold_seconds`, `speed` | Bingo test, nap |
| `nod` | A head nod, with the line. | `hold_seconds`, `speed` | Bingo test, nap |
| `wheel_nudge` | A small wheel move. | `hold_seconds` | engine |
| `rattle` | The line, then the rattle sound under a ⏳ countdown. | | Bingo test |
| `weather` | The live weather for that robot, read out. | | engine |
| `connect` | Takes control once for a control-style run. | | engine |
| `system_line` | A Goverlord bubble, said in the system voice; the bubble shows with the voice. `wait_spoken`: go on only once it's heard, else stop the run. | `wait_spoken`, `spoken_wait_seconds` | nap |
| `combo` | One story step played as several notes in order (named for them: `brat_nod_kg_success_0203`). A part with `text` says those words, cut from the step's TEXT ("Lol!", "Excuse Me?"); the part without `text` says the rest. One chat bubble for the whole step ("…maturity level? Lol!"), put up with the first words; each part lights its own pieces in it (`piece_offset`). | `combo` (the parts) | Bingo test |
| `sound` | One of the song's own sound files (knobs.json `sounds`, copied to the chat page at Wire-Pod's start), or one of Wire-Pod's shared `sounds/` (`danger-will-robinson.wav`), played on the page at `volume`; the song waits its length. | `sound`, `volume` | fart joke |
| `picture` | One picture (or a list) from the chat page's `pictures/` folder, shown in a picture bubble at this step, alone or as a combo part; the magnifier as ever. A missing file warns, never stops the song. | `picture` | |
| `thinking` | The live chat's own thinking process, run by Wire-Pod (`/api-sdk/gopod_thinking`, `gopodThinkingProcess`): arm, nod, the thinking line, searching (at least `searching_min_seconds`), the get-out. One process for live and scripted runs. While a manager has the floor (`gopodManagerBusy`, set by `gopodManagerDo`), the searching stops (one get-out) and starts again after (2026-10-22). Locked by `TestThinkingProcessIsOne`. | | fart joke |
| `timer` | A timed wait. | `seconds` | engine |
| `sequence` | Another song's steps, in place, by its `song_id`. | | engine |
| `pause` | A silent gap. At 0 seconds it's an edit mark for the reporters' voiceover. | `pause_seconds`, `gap_label` | Bingo test, nap |
| `exit` | Hands both robots back. | | Bingo test, nap |

### 3.2 Played by the Brobots chat (the interview)

`start interview` plays these, in Wire-Pod (`gopod_interview_song.go`):

| Note | What it plays | Fields |
|---|---|---|
| `ask` | A question, read out by its reporter, to a Brobot. | `speaker` (`reporter_1`/`reporter_2`), `to` (`brobot_1`, `brobot_2`, `brobots`) |
| `line` | A spoken line: a Brobot's answer, his brother's line, or a reporter's banter or handoff. | `speaker`, `cue` (his animation), `moves` (`arm`, `nod`, `cube` at their words) |
| `closer` | The closing line after "Interview run complete". | |

### 3.3 Played by Wire-Pod (an LLM warm song)

`warm_message`, `warm_wait`, `llm_song`, `warm_ready` and `sting`. See the step-name table in
[2.2](#22-step-names) and [The opener](#63-wire-pod-start-and-the-opener).

### 3.4 The weather in a line

**GOLDEN.** `{weather}` in an interview line is the live weather. One weather call
(`gopodWeatherSay`) serves live chat and scripted runs alike: the real weather (Windsor fallback
location), else `live_weather.offline_text`. Locked by `TestWeatherPlace`.

**Each Brobot says his own units first, by design (not a repeat).** Each says his own unit and
clock first, then the other, from his own Wire-Pod settings (`temp_is_fahrenheit`,
`clock_24_hour`): Brobot 1 metric and 24-hour ("16 degrees Celsius or 62 degrees Fahrenheit ...
as of 10:35 or 10:35 AM"), Brobot 2 imperial and 12-hour ("61 degrees Fahrenheit or 16 degrees
Celsius ... as of 10:28 AM or 10:28").

For weather in the live chat (place names, countries, flags) see [8.6](#86-weather).

---

## 4. The toolbox and the songs

### 4.1 Shared golden pieces

- **Animations** (`goverlord/wire_pod_overlay/chipper/animation_vocab.json`, all verified):
  happy, veryHappy, sad, verySad, angry, frustrated, dartingEyes, confused, thinking, celebrate,
  love, answering, searching, searchingGetout, kgSuccess. Each maps to one stock Vector clip.
  Everyday words pick them in the chat ("party" → celebrate, "mad" → angry); a play on words
  (the Gomads "go mad" pun) plays happy.
- **Sounds** (`goverlord/wire_pod_overlay/chipper/sounds/`): `rattle.wav` (the Bingo rattle,
  1.9 s) and `danger-will-robinson.wav` (Wire-Pod's `playSound`). Both play on the chat page.
- **Direct-SDK tools** (`gopod_probes/tools/`, built with `build_direct_sdk_tool.sh`): the
  rattle on a robot, "Brobots ready" in sync, the cube's light blip, a robot's info, putting a
  robot to sleep.
- **phcal, the bench:** tunes one move or sound on one robot and saves it: arm (speed, hold,
  cycles), nod, rattle and danger volume, animation hold, stay-in-place, reverse, the cube.
  The starting values are `goverlord/alias_lib_overlay/phcal_master_tweaks.json`; a tuned value
  can be written into one song step.
- **The chat page's song pieces:** the big Bingo ball (`show_ball`), the ⏳ countdown with
  juggled balls, the LLM warm-up bubble, the topic pictures (with the magnifier), and the
  reporter and Goverlord voices.

### 4.2 Each song's own pieces

| Song | Status | Its own pieces |
|---|---|---|
| `00_llm_warming` | WIP | The opener (`song_id` `brobots_llm_warming`): Wire-Pod plays it at every start. See [6.3](#63-wire-pod-start-and-the-opener). |
| `01a_brobots_bingo_test` | GOLDEN | Host and brat roles; the `OPENING SYNC` opener (wake, warm-up rattle, "Brobots ready" together, arm cue); three ball-draw rounds with the rattle, countdown and big balls; `angry`, `searching` and `kgSuccess` beats; 0-second reporter gaps; the LLM warm-up and "LLM Ready for Bingo"; touch interrupt off. |
| `02_brobots_interview` | WIP | Your score, played by the chat. Two reporters. One question per egg - the workbook, fractions and decimals, the rig, the rest of the cast, birthdays and merch, Bingo nights, foodmath, the weather, roast battles, our legacy - then the closer (`note: closer`, Brobot 2's "Press my backpack…") after "Interview run complete". The Brobots' answers are your own lines from the topic files, word for word; the reporters' questions are drafts. Chapters named in the song (`chapter`); the photos named with their questions (`pictures`); Brobot 1's Awaken bit (arm, nod and cube at their words, and the live weather); the brothers' handoffs; emojis and cues lightly sprinkled. |
| `101_brobots_nap` | WIP | "Do Baby Robots Dream?": a lyric video, the first timed song. Music plays and a female narrator reads the poem; the 32 captions and their times come from the video's SRT, and the Brobots sleep through it (put back to sleep if found awake) and wake after the last words for "Brobots Ready!". The soundtrack plays on the chat page and the poem shows in one narrator bubble the size of its picture (no header) - each section's words over its own picture (`baby_brobots1-5.jpg`, from its `images/` folder, one per section, Sleepy Robots to Together), fading in to 50%, out after its last words, the last one staying. Its ready line comes first, then "Brobots. Go to sleep now please." Needs no voices and no language model (both `false`). `start baby brobots` (typed, numpad or spoken) or `pha0b nap`. Made in Premiere Pro; the media stays private. |
| `03_fart_joke` | WIP | The elevator: Brobot 2 sets it up, the live chat's own thinking process (`thinking`), farts (`fart sfx/`, 80%), Brobot 1 goes frustrated, then angry, then leaves. `start fart joke`; "Fart joke started." / "Fart joke complete." Its ready line: "We here, are Fart Joke Ready". |
| `01b_brobots_bingo_game` | WIP | The live Chocolate Bingo game: a Go program (`gobingo`), not a step song; Brobot 2's angry reactor. |

### 4.3 Parked, not lost

- The face GIF on a robot's screen (`01a_brobots_bingo_test/face/bingo_number.gif`; its tool is
  archived).
- **The sleep family**, for the nap's B-roll. The tool is still here
  (`gopod_probes/tools/direct_sdk_robot_sleep_001.go`: sleep, wake, and play one named trigger
  or clip). phcal's `brobots_sleep_to_wake_direct_sdk` puts both robots to sleep and wakes them.
  The 126 bench aliases (`sleep-beat-*`, `sleep-palm-*`, `sleep-rts-*`, `pet-*`,
  `sleep-segment-*`) are retired from the alias library and kept as `sleep_beats` in the settings
  (`system-robot beats [family]`, `system-robot beat <alias>`). Vector's own sleep triggers are
  `GoToSleepGetIn`, `GoToSleepSleeping` and `GoToSleepOff`; the purr lives in the petting bliss
  clips. No snore animation exists in Vector's sources.
- **One-file songs:** one day `story.md` might absorb `knobs.json`, so a sequence is written in
  one file and applied to a new `zKnobs.json`. Too speculative for now: the two files stay as
  they are.
- **Video in the chat:** not supported yet.

---

## 5. Making and editing a song

### 5.1 Editing a song

Every part of a song is in its two files, so any part - a line, a step, a whole section or the
whole song - can be rewritten from scratch by hand. No code changes.

**What can be changed, and where:**

| To change | Edit | In |
|---|---|---|
| The words of a line or caption | `> TEXT:` under its `## STEP` | `story.md` |
| What a step does | `note` (any note in [The notes](#3-the-notes)) | `knobs.json` |
| Who plays it | `speaker` (`brobot_1`, `brobot_2`) | `knobs.json` |
| The order | the order of the steps (in both files) | both |
| When it plays | `start_seconds` (timed song), `pause_seconds`, `buffer_after`, `tempo_factor`, `global_tempo` | `knobs.json` |
| How a move plays | `hold_seconds`, `speed`, `cycles`, `settle_seconds` (tune it in phcal first) | `knobs.json` |
| An animation | `animation_token` (from [the toolbox](#41-shared-golden-pieces)) | `knobs.json` |
| Which part it's in | `section` in `knobs.json`, and the `## SECTION:` it sits under in `story.md` | both |
| The whole song's feel | the top keys: `timeline`, `robots_speak`, `sleepy`, `audio`, `chat_intro`, ... | `knobs.json` |

**Recipes:**

- **New words:** change the `> TEXT:` line. That's all. Or type
  `system-set <song> <step> text <words>` in the chat.
- **Add a step:** add a `## STEP <step_id>` with its `> TEXT:` in `story.md`, and a step with the
  same `step_id` in `knobs.json` (and `zKnobs.json` when there is one), at the same place in all.
- **Remove or move a step:** remove or move it in both files.
- **Rewrite a section from scratch:** delete its steps in both files, write the new
  `## STEP`s under the same `## SECTION:` in `story.md`, then the same steps in `knobs.json`.
- **Rewrite the whole song:** the same, every section. The `song_id` stays, so every `start`
  command and alias still finds it.
- **Try it:** `zKnobs.json` (your private copy) wins while it's there. `score <song>` shows the
  steps; `pha0b <song> <first_step> <last_step>`, dry, plays a range with no robots; then live.
  From the chat: `system-score <song>` and `system-play <song> ... dry`.
- **Lock it in:** a good live run, then the status line to `GOLDEN` with the date, then a commit.

### 5.2 Multi-line words

In every `story.md` - songs, the opener, the interview - a line right under `> TEXT:`, with no
blank line between, is the words' next line, and the bubble shows the line break. A blank line
ends the words. The song engine and Wire-Pod's reader (`gopodInterviewStory`) both read it so,
and `system-set ... text` replaces every line. Locked by `TestStoryMultiLine`.

```
## STEP goverlord_warm_0102

> TEXT: Voice ready to play.
Language Model Loading...
```

### 5.3 Making a new song

1. Copy `goverlord/runtime/songs/_new_song/` to `<number>_<name>/`. Set its status to `WIP`.
2. Write the sections and steps in `story.md`, and the same steps in `knobs.json`.
3. Give it a `song_id` in `knobs.json`, add that `song_id` to the engine's song list
   (`SONG_REGISTRY` in `run_golden_song_001.py`), and give `pha0b` its keyword (`brobots.sh`,
   matched by `song_id`). The folder's own name is yours to change any time.
4. `score <song>` prints it. `pha0b`, dry, plays it with no robots.
5. Play it live, a range at a time. Tune moves in phcal.
6. When a full live run is right, set the status to `GOLDEN` with the date.

### 5.4 An LLM warm song

Copy `_new_song/_llm_warm/` instead (the LLM warm template). It is played by Wire-Pod, not by
`pha0b`; at start Wire-Pod plays the one named by `llm.opener_song_id`. A step's `"volume"`:
`"default"` (`chat_audio.volume`) or a percentage of it (`"50%"`, `"150%"`; never past full), on
its voices and stings. Locked by `TestLLMWarmTemplate`.

**Songs that need no voices or language model** (like Baby Brobots): set
`"voices_required": false` and/or `"llm_required": false` in `knobs.json` (see
[2.3](#23-knobsjson-the-top-keys)). A run with no steps in between says nothing. Locked by
`NoneRequiredWarmupTests`, `TestOpenerSkipsEmptyLines`.

### 5.5 The `_llm_song` run

Every LLM warm-up has three parts: the voice-generation part, the in-between run, and the LLM
ready part. The in-between run is the **`_llm_song` run**.

- Each warm-up song folder (the opener `00_llm_warming`, every golden song, the `_new_song` and
  `_llm_warm` templates) has a fixed-name `_llm_song/` folder: `knobs.json` (steps, note `line`,
  speaker a reporter host) and `story.md` (the words).
- It is said between "Voices ready" and "LLM ready". Because the name is fixed, the folders are
  interchangeable: swap the whole folder to change it. No steps = nothing said.
- The opener's step for it is `llm_song_0104` (note `llm_song`); its run holds the reporters'
  opening.
- A song's warm-up reads its `_llm_song` before knobs' `warmup_lines`.
- **Guards:** `system-check` flags a warm-up song with no `_llm_song`. A missing or unreadable
  run says no lines (`GOPOD_CHECK` / `LLM_SONG` in the log) and never stops the warm-up. Esc and
  `system-skip` skip its lines.
- Code: `gopod_llm_song.go`, `_llm_song_lines` (`run_golden_song_001.py`).

### 5.6 The standard's check

`TestSongFoldersFollowTheStandard` (Wire-Pod's test run, before every deploy) checks every song
folder: a status line, the same steps in both files, every note type known. A mismatch is
named, and nothing deploys. `system-check` also matches `zKnobs.json`'s steps to `knobs.json`.

---

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
   scripted line's voice is made (see [10.2](#102-voice-files-made-once-play-many)).
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
**\*** key, a typed **"skip"** (never sent to the Brobots), **"skip"** said into the mic, or
**`system-skip`** (typed or menu group 1).

- It stops the reporter line being said. The rest of the reporters' lines show at once, whole
  and unspoken, their voices dropped, up to the next line that isn't theirs or the opening's
  end. Then the ready message.
- Heard markers still pass, so Wire-Pod and the songs go on.
- Managers and Brobots are never skipped.
- Every skip is logged with how it came in: "skipped (Esc|numpad *|typed|menu|mic)". A spoken
  skip is a `skip:` marker (`gopod_page_controls.go`).
- Page-only (`skipReporters`).

### 6.8 Replay a session

`replay <name>` (the session folder's name: `095819_session`, `095819` or
`2026-10-12/095819_session`) plays that session's score (`story.md` + `knobs.json`, the words as
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
4. **Thanks, or `max_minutes`:** it stops, and a Brobot asks keep or delete. The box stays
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
If the mic is missing, Goverlord says `mic_missing_text` (with his bubble), and only the mic input is disabled: a
numpad key works as a dry run - the same messages, "Mic disabled" in place of "Recording", then
"transcribing" and "try again" (`nothing_heard`). A failed `lsusb` isn't "mic missing", and a
missing mic is looked for again on each key press.

A numpad key while Goverlord's mic or numpad message is still being made or said waits for it:
"⏳ Wait - Goverlord is still giving the mic's message" (`wait_message`, `gopodUSBSaying`).

The mic service records through PulseAudio when PulseAudio holds the USB mic (GNOME's Sound
settings open; its default source is the mic) - `PULSE_FALLBACK` (2026-10-22).

The USB watch (`gopod_usb_watch.go`, from then on, every `usb_watch_seconds`): when the mic or
the numpad (`numpad_usb_ids`) goes missing or comes back, Goverlord says and shows
`mic_missing_text` / `mic_found_text` / `numpad_missing_text` / `numpad_found_text`. Two numpads
both work, one key at a time; two mics, the first found records. The mic service looks for the
mic fresh when it finds none (a mic plugged in later is found).

### 6.12 Battery watchdog

All session long, every minute, each Brobot's battery is read and logged
(`GOPOD_BATTERY_WATCH`). Below 3.7 V (golden rule 6), off the charger, he says he's dizzy and
needs help getting home (`battery_watch.lines`), with his bubble - once per drop, again only
after 3.8 V or his charger. Live chat only, at the next safe spot: never during a reply, a run
or a draft. The watchdog stays quiet on the charger, so look at the log. Every
`GOPOD_ROBOT_CMD` line ends with the Brobot's last battery reading. `battery_watch` in the
settings, `TestBatteryWatchOncePerDrop`.

---

## 7. Ways in: talking to the Brobots

### 7.1 The one door

All lines meet in one door, `gopodOperatorLine` (`gopod_typed_input.go`), checked in this order:

Record Session's answer → `start interview` → `stop run` → the warm-up gate → the run gate →
thanks → GOPOD Yourself → a voice command's words (`start bingo test`, `start baby brobots`) → a
live question.

### 7.2 The lanes

| Lane | What it is |
|---|---|
| **Voice** | "Hey Vector" or the backpack button. A custom voice command (`customIntents.json`: START_BINGO_TEST, START_INTERVIEW, START_BABY_BROBOTS, STOP_RUN, GOPOD_YOURSELF, REMEMBER_MY_FACE) runs its script, which hands its words to the door; anything else is a live question (`kgsim.go`, the same gates). |
| **Typed** | The typing line, aimed by its radio: 1, 2 or All (All at every page load). Enter sends; Esc leaves the box. A typed line shows as an Operator bubble, read out in the Operator voice. |
| **Numpad** | Push-to-talk while the numpad is on and the chat is open (below). |
| **Native** | Wire-Pod's own "Hey Vector" answers (the time, the weather, volume…). |
| **Page** | Clicks only: Record?, Brobots Chat, Clear Chat, the speaker icon. The page also plays every voice and tells Wire-Pod what it has heard. It never sends words (a typed line is sent through the door). |

**The Operator line** under the chat: `Brobots ()1 ()2 ()All [Type to the Brobot, press Enter] |
Record? [] | HH:MM:SS 🔇`. The speaker icon turns the background loop on and opens the system
menu (🔊), and turns it off and closes it (🔇).

**Typing details:**
- A typed word in odd capitals like "iS" is read as "is" ("iPhone" stays).

### 7.3 Numpad push-to-talk

- Turn it on with `system-numpad on` (the page switches the numpad as the line is sent);
  `system-numpad off` turns it off; `system-numpad` alone lists the keys.
- Hold a numpad key (its persona from the numpad map), speak, let go. The numpad's own keys are
  always push-to-talk, even in the typing box; every other key types.
- Holding 0 (`all_hold_seconds`, 0.5 s) is push-to-talk to All - Brobot 1 answers, then Brobot 2,
  like the keyboard's All button.
- KP7 is CHALK, KP8 is PLAYHEAD ([12](#12-managers-chalk-and-playhead)). KP+ and KP- turn the
  camera feeds ([14.6](#146-cameras-and-the-edge-tpu)). The \* key skips ([6.7](#67-skip)).
- Three taps of 0 (or NumLock) leave numpad mode; Goverlord says "Goverlord has released the
  numpad."
- **Its status bubble is an Operator bubble,** headed "Numpad to <who>": the status lines
  (`numpad_ptt.texts`: "KP7 to CHALK", keep holding…, Recording for CHALK, Got it -
  transcribing…), then the words, with "Generating..." and the hourglass while they're read
  aloud. Status lines are never spoken. While the system menu is open, the status is a line in
  the menu's results instead (no other bubble; the error sound still plays).
- **Sounds:** a key that works plays nothing (`numpad_ptt.key_sound` off). Each error message
  (`numpad_ptt.error_texts`: too short, nothing heard, no mic, not ready, failed, NumLock off,
  not live) plays grunge at 15% (`numpad_ptt.error_sound`). A key that isn't live (KP9, Cameo)
  is an error too.
- While recording, numpad presses are saved in the session folder ([6.9](#69-record-session-the-record-box)).

### 7.4 Native Wire-Pod answers

- `system-native on|off` turns Wire-Pod's own answers on or off for the session
  (`native.on_at_start` at Wire-Pod's start). During a run they're always off.
- Off, a native phrase gets Wire-Pod's quiet ending. Your voice commands and chat questions
  aren't native and always get through.
- `intent_names_ask` ("who am I?" by voice) is in `native.always_on`, so the robot answers it
  even with native off. A typed "who am I?" goes to the chat's LLM.
- **Native on top:** any intent the robot acts on himself ("who am I", "remember my face", the
  weather...) lets go of his chat hold first, so his own behaviours (face behaviours too) run
  (`gopodNativeLetGo`, called from Wire-Pod's `IntentPass`); the hold returns with the next chat
  answer.
- **Native intent notice:** a native Wire-Pod intent sent to a Brobot (spoken, or a typed cube
  command) shows "Native Wire-Pod Intent Detected" and the intent's name inside his bubble - the
  bubble notice template (`bubble_notices.native_intent`, `gopodNativeBubble`); his own words aren't
  known to Wire-Pod. Wire-Pod's own pass that closes a chat answer's voice request
  (intent_greeting_hello) is no native intent and shows nothing (`gopodLLMClosePass`).
- **Native weather goes to PLAYHEAD** (2026-10-21): a native intent in `native.to_manager`
  (intent_weather_extend -> weather, intent_clock_time -> time_place) goes to the chat instead,
  and its manager takes it - spoken weather and time are PLAYHEAD's, the same as typed
  (`gopodNativeToManager`).
- Locked by `TestNativeAnswersOnOffAndDuringARun`.

### 7.5 GOPOD commands: Wire-Pod intents by text

A typed (or numpad) line holding one of these words goes to the chosen Brobot as Vector's own
native intent - he does what he'd do if he heard it - not to the LLM. To All: both. First match
wins. The list is `native_cube.intents` in the settings (`gopod_native_cube.go`). To add a
command, add an entry there (its intent name from Wire-Pod's `intent-data/en-US.json`), then
`wpr 2`. Spoken lines are matched by Wire-Pod itself.

| Command (any of the words) | Intent | Vector |
|---|---|---|
| pop a wheelie / do a wheelie | `intent_play_popawheelie` | pops a wheelie on his cube (the cube must be near) |
| pick up your cube / lift your cube | `intent_play_pickupcube` | picks up his cube |
| fetch your cube / bring me your cube | `intent_imperative_fetchcube` | brings his cube |
| roll your cube / roll the cube | `intent_play_rollcube` | rolls his cube |
| play keepaway / keep away | `intent_play_keepaway` | plays keep-away with the cube |
| find your cube / where is your cube | `intent_imperative_findcube` | looks for his cube |

`system-brobot cube <name>` sends the same intents from the menu.

### 7.6 Face recognition (a starter)

Say "Hey Vector, remember my face": the custom intent `REMEMBER_MY_FACE` (`customIntents.json`)
hands the robot Wire-Pod's own "my name is CrushN8r", and the robot scans and saves the face
itself; "who am I" works as before. Voice only (a typed line goes to the chat). Needs `wpr 2`
after an edit. The face photo as a prize is not built yet. `system-faces` lists faces.

### 7.7 Who answers: addressing and All

- The Brobot named in the first two words answers. With no name, the one who heard it first
  answers, and his brother stays quiet (a glance).
- **All = all Brobots in sequence, everywhere, eggs too.** A typed or numpad line to All with no
  Brobot named: Brobot 1 answers, then once he's done Brobot 2 answers the same line (Brobot 1's
  answer is in the shared chat; `addressing.all_next_wait_seconds`). A line naming a Brobot goes
  only to him.
- **Starting with Brobot 2:** "starting with Brobot 2" / "Brobot 2 first" (or Pip) to both turns
  the order round; each keeps his own persona (`gopodAllStartsWith`, `TestAllStartsWith`). The
  mic mishear "starting with bro bought to" counts too. The operator's "to Brobots" header then
  shows Brobot 2's eyes first.
- **Follow up, no repeats:** Brobot 1 answers, Brobot 2 follows up - his prompt gets
  `addressing.follow_up_note` before the question (`math.follow_up_note` for Explain the Math:
  he can set his brother straight). He still says his own weather line, in his native units, and
  his reminder line.
- **Brother repeats dropped:** Brobot 2's follow-up drops a sentence of his own words when every
  word that carries meaning in it was already in Brobot 1's reply; one new word keeps it ("a
  man" vs "a man with a hat" is fine). Each keeps his own thoughts as paragraphs; the code's lines
  (his weather line, "Hey, that reminds me...") stay. A follow-up doesn't end with his brother's
  end line. `gopod_brother_repeat.go`, `TestDropBrotherRepeat`.
- The "All" sequence (Wire-Pod) and the egg dressing (the page) are separate processes that line
  up.

---

**Handing over in a line to both.** When a line goes to both Brobots (All, KP0 held, or "starting
with Brobot 2"), whoever answers first ends with one more thought, handing over to his brother:
his own hand-over line (`host_personas.<host>.handover`: "{partner}, what say you on this,
{nickname}?", the nicknames in turn - gear-head, cranky-shaft, coolant-leak,
incompatible-particulate); without one, `addressing.redirect_text`. Empty = none.

## 8. A live reply

**The thought bubbles template - the answer chain (GOLDEN).** How an answer is put together is a list
in the settings (`answer_chains`), not code order. `base` is the scaffold; each response group is a
copy (`{"from": "base"}`) or its own list of parts.

| Part | Kind | What it adds |
|---|---|---|
| `start_snippet` | stitch | a scripted line said first, word for word |
| `llm` | answer | the language model's answer (stitch parts before it go first, after it go last) |
| `weather_kit` | stitch | the follow-up's weather kit link (`live_weather.follow_up_kit_line`) |
| `end_snippet` | stitch | "Hey, that reminds me..." and other end lines |
| `redirect` | stitch | the first answer in a line to both hands over (`host_personas.handover`, else `addressing.redirect_text`) |
| `colour:value_point` | stitch | 2026-10-22: a value point's exact line (its say line) first, word for word; the model's part after it is its colour, cut to `colour_slots.max_sentences` (4), even with `budget_off` |
| `reminder` | stitch | 2026-10-22: a free answer (no value point given) always ends with "Hey, that reminds me..." (the next points on the list); then no other end line |
| `who_line` | answer | a manager's own who line, word for word (`managers.who_words`) |
| `manager_take` | answer | a manager's take (weather, time and place) |

Answer parts end the chain when they apply; a part that doesn't apply this time is skipped. Groups:
`brobots`, `brobots_follow_up` (the second Brobot in a line to both), `gomad_persona` (a line to
CHALK or PLAYHEAD). (`gopod_answer_chain.go`)

### 8.1 The steps of a reply

1. The Brobot who answers is chosen ([7.7](#77-who-answers-addressing-and-all)).
2. Thinking: the arm lift, then the nod, then the thinking line, "Got it!"
   (`chat_moves.thinking_text`, everywhere it's said; `chat_moves`).
3. The searching animation until the answer is ready.
4. The reply: each animation sent, the speech 0.9 s later, the next animation only once the
   robot reports the last one done. The Brobot speaks only once his bubble is on the page (see
   [9.8](#98-the-engine-and-the-page-in-dialogue)). Each piece is lit as it's said.
5. The battery is logged (`GOPOD_BATTERY`); the stay-in-place hold comes back.

Locked by the `TestGopodAddressed…`, `TestGopodShouldAnswer…`, `TestHighlightPiecesGolden` and
`TestStuckReplyGivesUp` tests.

### 8.2 Reply clean-up

- **Thoughts as paragraphs:** each thought in a reply is its own paragraph, a blank line between
  them (`live_split.thought_lines` on): the code's own lines (the weather, the value point, "Hey,
  that reminds me...") and gemma's thoughts, which end with an emoji. Highlighting stays per
  pause; only the look changes.
- **No question read back:** a reply sentence made only of the question's words (three or more)
  is dropped (`gopod_question_echo.go`).
- **Notes dropped:** a "(wait for Brobot 0 ...)" note and a "---" at a line's end are dropped
  (`TestWaitNoteDropped`).
- **A cue word that's a real word is said:** an animation's name at the start of a spoken piece is
  dropped only when what follows starts anew ("happy Hey there"); "celebrate small wins" is said
  (`stripSpokenActionLead`, `TestSpokenCueWordKept`).
- **Emoji leftovers:** a joined emoji's joiner and style marks are taken out with it, so the
  robot never says them ("yirl"). Locked by `TestEmojiJoinersNotSpoken`.
- **Links are said as their words:** "[GOPOD Weather Kit](/gopod_weather_kit.html)" is said
  "GOPOD Weather Kit", never its address (`gopod_speak_links.go`). A "[words](/page)" link is
  never read as a stage direction, so it shows as a link.
- **Clean GOPOD:** outside a roast battle, a Brobot's country flag is kept only when its country
  is named in the question or the reply; a wrong flag becomes the named country's flag when
  exactly one is named, else it goes (rule: use correct and proper country flag emojis). A
  sentence with a political word (`clean_politics.words`) goes; political content is allowed
  only in the roast battle (`clean_politics.allowed_topics`). `gopod_clean_politics.go`,
  `TestCleanPolitics`.
- **Song/system lines said by a Brobot** (`gopodRecordSay`, for example the photo's second "I
  see ...") go through `GopodSpeakSentences`: lit while said, shown as said after.

### 8.3 Said right: pronunciation and the three lingos

There are three lingos. A pronunciation goes only to the lingo it is for.

| Lingo | Who | Where |
|---|---|---|
| **Brobots** (party lingo, the default for a new pronunciation) | Brobot 1 and 2 | the pronunciation registry in `brobots_wirepod_interview_section_card_template_1_001.md` |
| **Reporters** | Brobot 3 and 4 | `reporter_voices.say_words` |
| **Butler** | Goverlord | `system_say_words` |

- Brobots: bruh "brugh", moron "more-on", chocolate-y "chocolate-ee", GOPOD "Gowp-awd", CrushN8r
  "Crush-Ehnaydir". Math is "madth".
- **Brobot lingo, word starts:** an entry can be `"word_start": true` - at the start of a made-up
  word, the rest kept ("chocoyum" -> "chalk-o-yum", "chocolateric" -> "chalk-o-latteric"); a
  dictionary word (chocolate) and its `except` words are left alone. Both Wire-Pod and the song
  engine. `TestBrobotLingoChoco`.
- Butler: CrushN8r with 8r as a long British "ate-ah" ("crushin-ate-ah"); "Master
  Crush-En-Ayd-uh", "Gov-er-lord"; "please" as "pleeze"; "mouse over" for MouseOver.
- Reporters: merch is said merch (the Brobots keep their own).
- Edit a lingo from the chat: `system-lingo <brobots|reporters|butler> <word> = <said>`.

### 8.4 The Brobots' math

**GOLDEN.** A math question in the live chat is worked out in code, exactly, and answered step
by step, word for word - gemma never calculates (`gopod_math.go`, `math` in the settings, locked
by `TestMath*`).

- **What it does:** fraction to decimal (long division, a repeating block barred: 1⁄3 = 0.3̅),
  decimal to fraction (over 10/100/1000, simplified), repeating decimal to fraction (the 10x
  method: x = 0.3̅, 10x = 3.3̅, 9x = 3, x = 1⁄3), percents ("15% of 80", "what percent is 3 of 4",
  to and from), and plus, minus, times, divide with fractions, decimals, brackets and the order
  of operations.
- **What it reads:** digits, LaTeX and spoken words ("zero point three repeating", "three
  quarters"). A plain sum needs a `math.trigger` at the end. A line with fraction, fractions,
  decimal or decimals is worked out too (`math.fraction_words`).
- **Repeating decimals:** "convert 0.1234 repeating into a fraction" - all of 1234 repeats
  (1234/9999). Only the end repeating: "convert 0.1234 with the 4 repeating into a fraction"
  (also "only the 4 repeats", "123, 4 repeating", or "0.12344 repeating"); "with the 34
  repeating", "with the 234 repeating" the same way.
- **How it's shown and said:** fractions show as 1⁄3 and are said "over"; "=" is said "equals",
  "−" "minus" (or "negative" before a number). Each step's words are `math.lines`.
- **To All, once:** a fraction-word line typed to All runs the math flow only (Brobot 1, then
  Brobot 2's follow-up). Brobot 2's follow-up says a decimal is a kind of fraction
  (`decimal_follow_up_line`).

### 8.5 Explain the Math (the contest format)

"[statement]. Explain the Math" - only a line that ends with it is this format.

- Brobot 1 answers first, then Brobot 2 follows up, whoever it was sent to (typed, numpad, All
  or the robot's own mic).
- Each says up to `math.flavour_sentences` of his own (the LLM's words, any sentence with a
  number dropped), then his share of the steps as bullets ("• Let x = 0.3̅."). Brobot 2 gets the
  last third (at least the last step). A step's bullet goes with its own line.
- It is a long answer: a fraction to decimal shows the long division digit by digit
  (`math.long_division_max_steps`), and a longer repeating block is said digit by digit ("the 1
  4 2 8 5 7 repeats").
- In the live chat it is an egg ([13](#13-eggs)): Goverlord's congratulations with
  `math.popup_title` over the challenge, Brobot 1's answer under it; Brobot 2's follow-up in his
  own marquee bubble. Its other eggs show too (Brobot 1's).
- Locked by `TestMathExplainTheMath`.

### 8.6 Weather

- **The take:** "weather" in a line goes to PLAYHEAD ([12](#12-managers-chalk-and-playhead)):
  "I'll help you with that.", then the weather line. The Brobots keep the facts in their prompt
  but don't open with the weather. Weather is not an egg.
- **The place:** a live question with the word "weather" gets the weather. One that names a place
  ("the weather in toronto ontario, right now?") gets that place's weather, or Windsor's when the
  place isn't found. A place in capitals after in/to/at/from anywhere in the line counts when none
  comes after "weather" ("I want to visit my granny in Honolulu Hawaii. How is the weather there
  today?" gets Honolulu). A place typed small counts too ("my granny lives in jerusalem what's
  the weather there today" gets Jerusalem). Windsor only when no place is named. An egg's word is
  never read as a weather place ("weather for bingo").
- **Province, state, country:** a place named with its province/state/country is looked up there
  (Windsor Ontario, not Windsor England). A country in the place ("Toronto USA") is sent to the
  weather service as its code ("Toronto,,US") and only a place in that country is taken
  (`gopodWeatherCountry`, `TestWeatherCountry`).
- **The line shows the country and its flag:** "Windsor Ontario Canada 🇨🇦, ..." - a flag is said as
  its country unless just named (`gopodSpeakFlags`, `gopodCountryName`). A place is named once
  ("Jerusalem Jerusalem District" is said "Jerusalem").
- **The line itself,** for example: "Windsor Ontario, October 14, 2026, 17° Celsius or 62°
  Fahrenheit, and overcast clouds, as of 1814 Hours Military Time or 6:14 PM." (Brobot 1,
  24-hour in Wire-Pod). The robot says "degrees" and "eighteen fourteen hours military time".
  Military time rules: four digits, "zero" for 00-09, "hundred" on the hour. Reporters follow the
  Brobots' rules for a flag after its country (not said twice) and for military time.
- **The weather face:** a bubble holding the live weather line gets Vector's own weather face
  ([9.7](#97-qr-codes-weather-faces-flags-and-links)), in the Brobot's own bubble.
- **The Weather Kit link:** Brobot 2's follow-up to a weather question ends his own words with
  `live_weather.follow_up_kit_line`, before his "Hey, that reminds me...", as its own paragraph:
  the bubble shows "GOPOD Weather Kit" as a link (opens `gopod_weather_kit.html` in a new tab),
  he says the words.
- **Offline:** the Brobot says `live_weather.offline_text`. The session dry run loads no weather
  key, so its weather is always the offline line.

### 8.7 Photo: "what's in front of you?"

A starter. "what's in front of you?" / "what do you see" / "take a photo" starts it. The photo
runs inside the one thinking process (after its line, before its searching moves):

1. The Brobot says the thinking line.
2. **Photo wake:** after control, he waits `photo.awake_seconds` (2.5) - awake, head settled.
3. He takes the photo the golden way (Wire-Pod's own getImage): mirror mode, "3, 2, 1", the
   photo, mirror mode off, the shutter animation.
4. The TPU worker enhances the snapshot, finds objects in it and draws their boxes. The three
   pictures - [raw snapshot] [enhancement] [TPU object boxes] - show in a plain picture bubble,
   in one row (`photo.quick_look.pictures` / `enhance`, `_enhanced.jpg`, `_tpu.jpg`). Get Photo
   is not an egg.
5. The Edge TPU names the objects at once (about 0.04 s; its worker loads the model once at
   Wire-Pod start, `photo.quick_look`).
6. The searching moves, then **statement 1:** "I see a person, a chair and a keyboard."
   (`quick_text`; "Hmm, I need a closer look." when it names nothing).
7. Meanwhile the vision model (`photo.vision_model`, moondream through Ollama) looks at the
   enhanced picture in the background; its notes are only in the log (no bubble). While it's
   still looking: "Gathering details. Please wait!" (`gathering_text`), then one stall line
   (`stall_lines`, draft lines).
8. **Statement 2,** in his own bubble: "I see ..." from what moondream saw (its own "The image
   shows" taken off, `detail_strip_prefixes`) plus statement 1's objects it didn't name ("...,
   and a tv."). It never repeats statement 1's objects; a look that names nothing new says
   "That's all I can make out."

- Statement 2 is part of his answer: in All, Brobot 2 starts after it; a touch stop skips it.
- "Tell me what you saw" answers from what he saw last ("I saw ..."); the photo is never sent
  again. No photo: "I couldn't take a photo."
- Photos are kept in `webroot/photos/` (private, gitignored); `system-photos` shows them.
- `photo` in the settings; `gopod_photo.go`, `gopod_photo_quick_look.go`, the worker
  `gopod_probes/tools/gopod_edge_tpu_quick_look_001.py`; `TestPhotoKind`, `TestPhotoRecall`
  (`GOPOD_PHOTO_LIVE=1` also asks the vision model), `TestPhotoRounds`.

### 8.8 Search

GOPOD's own web search, no LLM in it: the `gopod-searxng` container (this machine only, port
8055; started by `wire_pod_overlay/searxng/run_gopod_searxng.sh`, its private settings in
`~/gopod_searxng/`). Go search is retired.

- **To the Brobots:** a line opening with `search <words>` or `look up <words>` (a Brobot's name
  first at most) gets the top 3 results' snippets as facts. Both Brobots get the same search for
  2 minutes.
- **The results** are PLAYHEAD's (`system_search.results_host`): PLAYHEAD's part at the top of the
  answering Brobot's bubble, since it isn't a direct call to a manager. Small type (50%), never
  spoken or highlighted, typed in character by character once it shows; then the Brobot's
  response. The reply does not open with the result. `system_search.results_role` is the box's
  role.
- **Typed `system-search <words>`:** the results in the terminal bubble (or the menu's results).
- Settings `system_search`; `gopod_system_search.go`, `TestSystemSearchQuery`.

### 8.9 Pictures described to the Brobots

- **Egg picture points:** every egg picture has a few points about it in
  `webroot/gopod_picture_descriptions.json`, written by hand from each picture. They're made only
  for a new or changed picture, at Wire-Pod's start, never during a run; the Brobots get them
  when the egg pops.
- **Vision descriptions:** a pop-up picture in `topic_pictures.describe` (granny) is looked at by
  the photo egg's vision model; the Brobots answering get "The pop-up picture shows: ..." before
  the question (`describe_note`, waits at most `describe_wait_seconds`, 30; once per Wire-Pod
  start per picture).
- **Optional pictures** are looked at only when asked: a question with "chocolate" has the
  Chocolate Bingo kit and card looked at for the reply, pop-up or not (`describe_on_ask`).
- While pictures not looked at yet are being looked at, Goverlord says "Looking at N picture(s).
  Please wait." (`describe_wait_text`; done: `describe_done_text`).
- Egg pictures are looked at ahead, quietly, 30 s after Wire-Pod starts, so a pop-up's reply
  doesn't wait.
- Live snapshots have no Goverlord bubble ([8.7](#87-photo-whats-in-front-of-you)).

---

## 9. The chat page and its bubbles

**The raw bubble template (GOLDEN, being built type by type).** Every bubble type is built on one
template: `rawBubble(spec)` makes the bubble whole (kind, classes, header, lines);
`bubbleTask(b, promise)` adds a job it must finish before it shows; `placeBubble(b, then)` builds it
in a hidden staging area outside the chat (so it never moves the chat's scroll), waits for its tasks
and its pictures to load (10 s at most), then puts it in the chat and shows it - and only then does
the next event start. The exceptions keep their own rules: the chat window (scrolling, the baton),
the Operator bubble (the one input type) and the system window (the menu). On the template so far:
system / terminal answers, reporter and manager lines (their own bubbles), Goverlord's lines, Brobot replies.

### 9.1 The Chat Bubbles Window: the machine

**Draft for signature** - once the owner signs it, this is the locked rule set.

**What it is.** A Brobots layer on native Wire-Pod (the overlay, "chat bubbles lite"):
Wire-Pod's own Log page gets a scrolling window of bubbles that shows the show as it happens -
every line said, by whom, to whom, with its pictures, timers and eggs. The terminal (journal)
carries the same show, line for line (`GOPOD_FEED <id>`, the page answering
`GOPOD_PAGE played <id>`).

**From the Log click to the bottom of the window:**
1. **Log click** - the page counts as here (seen, clicked, chat open). Before that, Wire-Pod waits
   ("Goverlord awaits your return") and nothing plays into an empty room.
2. **The feed** - Wire-Pod numbers every line it adds to `gopod_chat_feed.json`; the page plays
   them one at a time, in order, and says "played" for each.
3. **A voice is made before its line shows** - the line waits in its own bubble with
   "Voice file generating." and the hourglass, then its words ([9.4](#94-golden-the-generating-template)).
4. **Every bubble's life is the same:** built whole while hidden (pictures, QR codes, weather
   face, flags in place) → the chat scrolls once to make room → it shows when the bubble above is
   done (the baton) → its words are revealed as they're said → said (85%).
5. **Who reveals the words:** the Brobots' replies - Wire-Pod names the piece being said
   (`gopod_speaking_now.json`), the page reveals that piece word by word. Reporters' and
   Goverlord's lines - the page plays their voice itself, so it reveals the words on that voice's
   own word times (`voiceLitLine` / `voiceLitStart`). SRT songs - the soundtrack's times. Word,
   then its punctuation, then emojis; numbers, clock and year as said.
6. **Small processes plug into that one life** and work for every bubble that needs them: the QR
   codes, the egg marquee, the magnifier on every picture, the weather face, flags as pictures,
   the search-results box, the hourglass.
7. **Scroll to the bottom** - the newest bubble ends just above the typed-input line; the chat
   follows the highlight; nothing scrolls again until the next bubble.

**Shared helpers - one each, used everywhere:**
- `voiceWaitOn` / `voiceWaitEnd` - the one Generating template, keyed by voice file.
- `slotWaiting` / `slotReady`, `fillSlot` - the grey "Voice file generating." slot, filled in
  place by the line that follows. (Not used for the Operator's own bubble.)
- `voiceLitLine` / `voiceLitStart` - the word reveal for every voice the page plays.
- `magnify` - the magnifier on every picture, QR code (QR1, QR2) and weather face.
- `whenShown` - anything that waits for its bubble to show: the weather sound, the search
  results' typing, and the reveal's 15 s fallback.
- `cornerBox` - QR codes and the weather face in a bubble's top-right corner.

### 9.2 Bubble kinds and headers

| Kind | Header | Look |
|---|---|---|
| System (Goverlord) | "[grey icon] Goverlord: GOPOD Brain [grey icon]" | grey. Every one is headed so; "Reporter 0" is only a decoration label. |
| Operator | "[operator icon] [input] to Brobot N [Brobot N eye-colour icon]" | white-eyed icon. To Brobots: both icons. |
| Reporter | "[white icon] Reporter 1: Brobot 3 [white icon]" | white-eyed icons |
| Brobot | "[eye-colour icon] Brobot N:" | his eye colours, one icon |
| Narrator | no header | one per poem section |
| Manager | CHALK's chalkboard, PLAYHEAD's tape-brown (the colour of vintage video tape cassettes) | see [12](#12-managers-chalk-and-playhead) |
| Goverlord (terminal) | Goverlord's header | text at 50% size; system answers |

- System answers and non-numpad status lines carry Goverlord's header.
- **Header icons mirror each other.** For headers with more than one icon, every icon after the
  first plays `assets/brobot_eyes_half.webp`: the same eyes, its frames starting at the loop's
  half (2.72 s of 5.36 s), so the icons' animations mirror each other rather than move together.
  Each icon keeps its own eye colour (`gopod_chat_bubbles.css`). The original `brobot_eyes.webp`
  is untouched.
- **Eye colours from Wire-Pod:** each Brobot's own custom eye colour, from Wire-Pod's stored
  robot settings (`vic.RobotSettings`), becomes the bubbles' colour at Wire-Pod's start and every
  30 s, never during a run (`gopod_eye_colors.go`). A Brobot with no custom colour keeps the
  page's own.

### 9.3 The chat rules

**GOLDEN (rules 10, 11, 14, 15 in [Golden rules](#17-golden-rules)).** A bubble is built whole
while invisible, the chat scrolls once, then it shows, with its voice when it has one; the next
bubble takes the baton only when the one above has finished.

1. A bubble is built fully hidden, away from the chat, so it never moves the scroll: out of the
   chat's flow, taking no room and opening no scroll space.
2. It's ready once its content has loaded and rendered (pictures loaded, its voice loaded and
   ready to play).
3. The full bubble gives its final height; that height means it's ready to show.
4. Only then, at its turn in the queue, does the chat scroll and show it. It comes in from the
   bottom. Its voice plays once it shows.
5. A short bubble stops bottom to bottom (its bottom just above the typing line); a tall one
   (taller than the window) stops top to top. Otherwise the chat stays at the bottom. One
   scroller, 1 s ease, never up on its own.
6. Every bubble runs and finishes 100% before the next one shows.
7. Headers as in [9.2](#92-bubble-kinds-and-headers).
8. Live operator lines (keyboard, numpad, and the robot's mic held during a warm-up, "Mic to
   Brobot N"): an Operator bubble read out in the Operator voice
   (`reporter_voices.operator_voice`, `af_sky`). A playing draft's question keeps its reporter's
   voice. Lines to CHALK/PLAYHEAD are read in the Operator voice too.
9. No time limit takes the baton early: the next bubble waits for the one being said. Only a line
   lit longer than `chat_window.speaking_max_ms` (30 s - a frozen robot, a lost end signal) lets
   go. However a voice ends (played, errored, hit its cap), its line counts as said.
10. **Narrator bubbles** are the nap's captions: the song's subtitle (`.srt`) lines, one bubble
    per poem section, each phrase lit on the soundtrack's clock - nobody's voice. Like every
    bubble, each is built out of the chat's flow and joins it when it's placed. A timed bubble is
    sent early (`narrator_lead_seconds`) and shows at its first line's second.
11. **A bubble that grows on screen** - a new message, a picture loading, a picture added to a
    question bubble - is placed again, whole, while it's in view.
12. **The LLM warm-up bubble is the exception:** it shows first and is kept at the bottom.
13. **Pop-up placement:** once the reply joins an egg, the chat aims the reply's "[icon] Brobot
    N:" header at the window's top; if the bubble's bottom reaches the window's bottom first, it
    stops there - no extra space below (rule 5, aimed at the reply header).
14. **Log opened after a refresh:** the chat goes straight to its bottom the moment the window
    gets its height.
15. **The highlight is held, not the next bubble's place:** the chat keeps following the
    highlight while a next bubble waits (a staged bubble is never "the end of the chat"); the
    highlight stays on the bubble the Brobot is saying (past its first piece) and never lands on
    an egg.
16. **A voice waits** while its bubble - or any bubble before it - is still waiting to show, as
    long as the baton can be held.
17. **Backpack question and its egg:** a line the robot heard is logged once its answer is ready,
    after its egg; it goes above that egg and the reply still joins it.
18. **Bingo ball:** the called ball shows under its "Bingo ball captured. B-1" bubble when "B-1"
    lights (its room kept while hidden).

**The window's numbers are data.** The page's timings and sizes (baton, scroll, countdown hold,
marquee, narrator fades, reveal fallbacks) are `chat_window` in the settings; the numbers in the
page code are only fallbacks.

### 9.4 GOLDEN: the Generating template

A voice is made before its bubble shows. One robust, deterministic mechanism for every voiced
line:

- Wire-Pod sends a line whose voice isn't made yet first, marked `generating_marker+"wait:"`
  (`gopodSayShown`, the one path for shown voice lines). The waiting line names its own voice
  file (`wait:<file>|text`, `voiceWaitFor`), so its voice always finds it - a bubble posted in
  between never leaves it stuck.
- The page shows the generating text (`reporter_voices.generating_text`, "Voice file
  generating.") and the hourglass, rocking and draining, under the header in that line's own
  bubble, until its voice plays; then its words. "off" (voice failed) shows the words.
- 2026-10-22: one Generating look everywhere - the text, then the hourglass, the text's own
  height, on its line (`.voice-wait`; `voiceWaitOn` and `slotWaiting`).
- A skip's end (Esc, numpad *, `system-skip`) takes the chat to its bottom (`skipEndScroll`).
- **The skip gate** (2026-10-22, `skip_gate`): every skip (Esc, numpad *, the menu, typed or
  spoken "skip") is off until a warm-up line starting with `enable_on` ("Voices ready.") shows -
  the opener's and every warm-up song's - and off again once used. A skip while off does nothing
  (logged: skip ignored).
- **GOLDEN voice start** (2026-10-22): every made voice plays the same way - its bubble shown,
  then its length and its word times (`.words.json`) loaded (up to `VOICE_READY_MAX_MS`, 3 s),
  then it plays and lights. A just-made voice no longer starts before its word times arrive
  (it lit by the estimate, out of step).
- It works the same for an operator line, a reporter's, a Goverlord bubble, a manager's, and the
  opener's System llm bubble. The Operator's own words wait the same way ("op:<file>"). No
  separate "Generating..." bubble.
- One template: `voiceWaitOn` / `voiceWaitEnd`, keyed by voice file.
- A voice asked for while Kokoro is still starting waits for it (up to `start_wait_seconds`)
  instead of failing.
- The stings (grunge, dJent3b) and the numpad key sound are loaded once and replayed; the
  background loop is decoded once and looped.
- Code: `gopodReporterFileShown`, `reporter_voices.generating_text`/`generating_marker`.

### 9.5 Highlight and word reveal

**The highlight rules, in one place (GOLDEN).** Highlighting is one mechanism for every bubble:
it comes in, lights by these rules, and gets out.

1. **The chain, largest to smallest.** A reply is cut once, by Wire-Pod, into **pieces** (per
   pause: commas, sentences). The piece being said is lit **word by word**; inside a word, a run of
   single digits, a military time or a clock time lights **one number (or spoken part) at a time**.
2. **Three states.** Not said yet: hidden. Being said: solid white. Said: 85%.
3. **Three timing sources, each naming its bubble exactly:**
   - **A Brobot's speech:** Wire-Pod reports the piece number *and the bubble's chat event*
     (`gopod_speaking_now.json`: `piece`, `feed`); the page lights that piece in that bubble.
   - **A voice** (reporters, managers, Goverlord): the line names its own voice file
     (`lit:<file>|` or, while it's made, `wait:<file>|`); the words light at that voice's own word
     times (Vosk) as it plays.
   - **The narrator:** its SRT times, on the soundtrack's clock.
4. **Timing within a piece:** a voice uses its Vosk word times; a Brobot's speech gets the real
   start of each piece from Wire-Pod and a best guesstimate for the words in it; with no timing at
   all, a guesstimate (`word_reveal.seconds_per_word`). A line whose voice never plays shows after
   `voiceLitWaitMs` (15 s); a Brobot that stops speaking leaves the rest of his bubble shown as said.
5. **Nested parts** (a manager's part, an egg summary inside a Brobot's bubble) light on their own;
   the Brobot's own pieces never count them (`ownSays`).


**Pieces, per pause (locked - change the split, highlight or bubble look only with a plain
before/after and the owner's OK).** A Brobot's line is cut into pieces at its pauses (comma
pieces): `join_up_to_words` 0. Locked by `TestHighlightPiecesGolden`.

**Heard, not read ahead.** A highlighted bubble (a Brobot's reply, a scripted song line, the
narrator) is built whole, but its words are invisible (0%) until heard: the piece being said is
solid white (100%), a piece already said 85%. The pieces before a lit one count as heard; when
the next bubble takes the baton, or 10 s after the robot stops with nothing more lit, the rest
shows. A restored chat shows all. `.spoken` / `.speaking` in `gopod_chat_bubbles.css`,
`lightSpan` in `gopod_brobots_chat.js`.

**GOLDEN: word at a time.** The piece being said starts at 0% solid white, then each word turns
100% solid white as it's said and 85% once said.
- Every word is its own step (`word_reveal.seconds_per_word`, 0.38). A word's punctuation, emojis
  and ° are steps of their own after it (`word_reveal.seconds_per_mark`); marks inside a word stay
  (Let's). A flag is a step. A math formula is one step.
- A silent lead-in of two words' time comes first, then the first word lights
  (`word_reveal.start_delay_words`), so the reveal doesn't start before the Brobot speaks.
- **Numbers one at a time:** in the piece being said, a run of single digits ("the 1 4 2 8 5 7
  repeats") and a military time light one number at a time (`digit_highlight`,
  `digit_highlight.seconds_per_number`).
- **Clocks and military time step as said:** a clock steps "1", ":", "43", "P", "M". "1027 Hours
  Military Time" lights 10, then 27, then "Hours", "Military", "Time" (the label after the
  numbers); "1:16 PM or 1316 Hours Military Time." reveals 1 / : / 16 / P / M / or / 13 / 16 /
  Hours / Military / Time. A year after its date steps "20", "26" and is said "20 26"
  (`gopodSpeakYears`).
- For the Brobots' replies (Vector's own voice, no audio here) it is a timed estimate, not tied
  to the audio; when the piece ends the rest shows. The page code: `digitUnits`/`digitSequence`.
- The word reveal runs only in live, scripted and narrator bubbles (no flicker in plain ones).
- While a link's words are revealed, only the shown words are underlined.

**Forced alignment: word times from the voice.** The voice worker has Vosk listen to each voice
it makes and gives every word its heard time (Kokoro's own time only for a word Vosk missed; a
mark where its word ends) - `"aligner": "vosk"` in `<name>.words.json`, saved beside each voice.
A voice without one, or timed by Kokoro's estimate only, is made again on its next use. The page
reveals a reporter's or Goverlord's line (and an egg's congratulations) word by word at those
times (`alignUnits`) - a mark or emoji when its word ends, a word said another way
("crushin-ate-a") between its neighbours.

**One reveal, any source of times.** The word reveal (`digitSequence` + `alignUnits`) uses word
times whenever a piece carries them, whatever made them - a Kokoro voice's `.words.json`, or a
song's `narrator_words.json`.
- **A song's narrator:**
  `gopod_probes/tools/gopod_align_captions_001.py <narrator audio> <song folder>` - Vosk (Wire-Pod's own speech model) times the words it hears; each SRT caption's
  words are matched in order within its seconds; saved as `narrator_words.json` (Kokoro's shape);
  the audio isn't kept. The song engine puts each caption's times on its piece
  (`@s-e ~<base64> words`). Baby Brobots has it (170 of 179 words heard; the rest between their
  neighbours).
- Without word times, a narrator caption's word reveal starts at once and is scaled to fill its
  SRT segment exactly, first word to last.

**Follow the highlight.** The chat holds still until the piece being said is a quarter of the
window below centre (`highlight_scroll.dead_zone`), then one smooth move puts it back at the
centre - never past the chat's end (the bottom/bottom rule), never back up on its own. Once the
bottom is reached, the highlight carries on down the bubble.

### 9.6 Pictures

- **The magnifier (GOLDEN):** every picture has it - hover or tap to see it full width. New
  pictures too, QR codes and weather faces included.
- **Pictures in one row:** a bubble's pictures share one row, each scaled; the magnified one takes
  the full width.
- **A song's picture:** `{"note": "picture", "picture": "granny.jpg"}` (one file or a list, from
  the page's `pictures/` folder) shows them in a picture bubble at that step, alone or as a combo
  part. Label the step after it, for example `host_say_img-granny_0301`.
- **Narrator pictures:** a song's `section_pictures` (knobs.json, section -> file in the page's
  `pictures/`) puts that picture in the section's narrator bubble with the lit words, with the
  magnifier.
- **Picture rows in the menu:** only a system command's picture row ("menu:" mark) goes in the
  menu's results while it's open; songs' and topics' stay in the chat.

### 9.7 QR codes, weather faces, flags and links

- **QR codes:** a bubble whose words hold "crushn8r.ca" gets the crushn8r.ca QR code;
  "newsletter" (or "crushn8r.ca/subscribe") the newsletter one (crushn8r.ca/subscribe); "nightlife"
  or "Mississauga" the flare4.com one. CHALK shows them ([12.3](#123-managers-inside-a-brobots-bubble)).
  In a corner, QR1 sits top-right in the header row; scrolled above the chat's top it jumps to
  the bubble's bottom-right (`corner-low`, marquee inset kept). QR2 (`qr_codes.second`) goes
  right after "...subscribing to our newsletter!", the last step of the reveal. The magnifier as
  ever. `qr_codes` in `gopod_brobots_settings.json` (`qr_codes.codes`: the words that show each),
  pictures in `webroot/qr/`.
- **Weather faces:** a bubble holding the live weather line gets Vector's own weather face for it
  (sunny, cloudy, rain, snow, thunderstorms, windy; clear at night = stars; cold has none),
  animated, Vector's own size (184x96), in the speaking Brobot's eye colour, in the same
  top-right corner, with the magnifier. Its weather sound is loaded with the bubble and plays once
  as it shows, under the Brobot's voice (`weather_card.volume`). Only the weather's own words pick
  the face, never the place's. `weather_card` in the settings; the kit is `webroot/weather/`, all
  of it shown on `gopod_weather_kit.html` (locked as is; linked from Wire-Pod's weather setup
  section, the overlay's `setup.html`).
- **Flags as pictures:** a known country's flag is drawn from `assets/flags` (Twemoji, CC-BY 4.0),
  because some browsers (Windows) draw a flag emoji as its letters ("US").
- **Links:** any "[words](/page)" in a bubble is a link, like the Weather Kit link.

### 9.8 The engine and the page in dialogue

One event list, one source of truth: the web page and the terminal run as one coordinated
dialogue.

- **The page is here only once it can play:** after `wpr 2`/F5, the click (the Log button).
  Before it, Goverlord awaits and runs wait (a browser plays sound only after a click).
- **One event list:** every chat-feed line is in the journal as `GOPOD_FEED <id> <who>: <words>`,
  in order - voices, heard markers, warm-up messages, bubbles. **Live replies are feed lines too**
  (via `llm`).
- **The page answers:** each second it says how far it has played the feed; the journal writes
  `GOPOD_PAGE played <id> <who>: <words>` under the engine's own line.
- **Heard is the truth:** `GOPOD_HEARD <who>: heard` / `not heard after Ns (going on)` / `no voice
  file`, for every voice the engine waits on. The page reports voices that didn't play (missing,
  blocked) with the next heard marker; a song's reporter and ready lines get `heard` or
  `not_heard` back, and the song engine prints a not-heard one.
- **The engine conducts:** Wire-Pod times itself by the page instead of guessing. (1) A Brobot
  speaks once his bubble is on the page - up to `page_sync.reply_wait_seconds` (10) - so he never
  talks over a line the page is still playing (`gopodWaitReplyShown`; `GOPOD_SYNC` lines when he
  waited or the page was behind). (2) The page tells how long it held each line before showing
  it; over `page_sync.held_log_seconds` (3) it's in the journal (`GOPOD_PAGE held <id>`). The
  page still keeps its drawing order (one bubble after the one above, the countdown, the SRT
  narrator's soundtrack clock); the engine waits on those instead of running ahead of them.
- **Every wait on a line:** Wire-Pod goes on the moment the page says it has played it (the heard
  check). With no answer, after the queued lines' own lengths plus
  `reporter_voices.heard_margin_seconds` (3). No page-lag padding is added.
- **Wait until I'm finished:** a voiced line waits until the page has played it - its own length
  plus up to `reporter_voices.finish_wait_seconds` (120) - while the chat page is watched and
  here. With no page, the cap is 2 x `max_wait_seconds`.
- **A voice's cap starts when it plays** (its length + 5 s; 60 s if unknown), so it's never cut
  or left to start late, out of order.
- **Every live reply is accounted for:** the page writes
  `GOPOD_PAGE live reply <id> Brobot N: shown` (or `NOT shown - <why>`) under its
  `GOPOD_FEED <id>`; Wire-Pod writes `GOPOD_FEED_CHECK ... the page played past it with no bubble`
  for a missing one and `GOPOD_FEED_CHECK live reply twice in a row` for a doubled one.
- **No orphan state file:** `gopod_rich_display_ui_state.json` is not written.
- See it: `journalctl -u wire-pod -f | grep -E "GOPOD_(FEED|PAGE|HEARD|AWAIT|OPENER)"`.
- Locked by `TestFeedDialogue`.

### 9.9 The system window

**GOLDEN, locked - change only with the owner's OK.** The system menu's window:

- A Goverlord system bubble at full width holding the menu bubble; no header (to recover window
  space). The system bubble centres its contents.
- Two windows, menu above and results below, start 70/30. The thin bar between them (row-resize
  cursor, a grip line, stretching across) sets the menu's share, 10% to 90%; the results take the
  rest. The results window is full width, empty or not. Window 2 overlaps window 1.
- The menu a little wider (right padding 2px); each window's scroll bar in its own lane, clear of
  the rounded corners.
- Menu at 50% font, for more room below.
- While the menu is open, the chat is scroll-locked at its top.

### 9.10 The background loop and the speaker icon

- `background_loop` (`BrobotsGroove-Loop-1.wav`, 6 s) plays on the page over and over with no
  gap, from its top again at every session clock reset to 00:00:00.
- The speaker icon after the text field (🔊/🔇) turns it on or off (and opens or closes the
  system menu). Every new Wire-Pod start sets it off.
- Hovering the icon shows a vertical volume bar (0-100%, the loop's share). Let go, its value is
  saved to `background_loop.volume`, so it holds over restarts. The slider gives way when the
  settings' loop volume changes.
- Every run quiets the loop at its start and, as its very last line, sets it back to the toggle.

---

## 10. Voices

### 10.1 Who speaks with which voice

| Voice | Kokoro voice | Says |
|---|---|---|
| Reporter 1 (Brobot 3) | `af_heart` | questions, banter, the opening and closing; CHALK's voice |
| Reporter 2 (Brobot 4) | `am_michael` | questions, banter; PLAYHEAD's voice |
| Goverlord (the digital butler) | `bm_george` | system lines |
| Operator | `af_sky` (`reporter_voices.operator_voice`) | your live lines, read out |

- All at `reporter_voices.speed`.
- Goverlord's lines are said by the digital butler (`system_persona`, from Master_CrushN8r.pdf)
  with its own phonetics (`system_say_words`). Run lines keep the run name first.
- The page plays voices one after another; a line that must be heard before the run goes on is
  waited for (golden rule 11).
- The Brobots speak with Vector's own voice, on the robots.

### 10.2 Voice files: made once, play many

- At each Wire-Pod start every scripted line (system, reporters, the interview, each song's
  warm-up, the run commands) is made once, in the one voice that says it.
- Each line is made once and kept, named by its voice and speed: a new voice or speed never
  overwrites. Locked by `TestReporterFilesNamedByVoiceAndSpeed`.
- **Voice kinds:** each voice's kind is noted as it's made (`voice_kinds.json`): **system** =
  Goverlord's voice; **scripted** = the start's check or a playing script (reporter, narrator...);
  **live** = the Operator's lines and the managers' live answers.
- **Voice keeping, no auto-deletes:** at each start a voice not in use and not said for its kind's
  days (`reporter_voices.operator_keep_days` for system lines, `scripted_keep_days` for scripted lines,
  `live_keep_days` for live lines) moves to `sessions/archives/voices/` (the session chats'
  archive). Saying a line again resets its days. Logged as
  `GOPOD_REPORTER_VOICES cleaned up N old voice files`.
- **Rebuild all voices:** `system-voices-rebuild [go]` runs `system-check` first (`story.md`,
  `knobs.json`, `zKnobs.json`, every file named) - a problem pauses it ("go" rebuilds anyway);
  lists the scripted lines that need a voice; moves every generated voice to
  `webroot/gopod_reporter_audio_backup/` (sounds stay); says "Ready for a restart". Then `wpr 2`
  and F5: the start's generation check makes every line again; each voice's backup copy is
  deleted once its new playable voice is made. `gopod_system_voices.go`.

### 10.3 (music) places the sting

A Goverlord line that starts with `(music)` has the sting (`system_stings.default`, dJent3b)
before its words; one that ends with `(music)` has it after. `(music)` is shown (for deaf
viewers), never said. The warm-up's end sting doesn't play again when the ready line's own
`(music)` played it (`gopodMusicMarks`, `gopodMusicPlayed`).

### 10.4 System stings

A Goverlord line with no `(music)` gets a sting before it is spoken: `grunge.wav` for a run's
started, stopped, complete, paused or continues line (and the stop-run answer when nothing runs);
`dJent3b.wav` for every other system line. One sting per bubble (`reporter_voices.system_stings`).
Message 1 of the opener has no sting.

---

## 11. Volume and sound

- **One master:** `chat_audio.volume` (35%). Every sound the chat page plays - voices, stings, a
  song's sounds and soundtrack, weather sounds - plays at it.
- **Every other volume knob** - a sting's (`system_stings`), the numpad error sound, the weather
  sound, the background loop, a song step's `"volume"` (the opener's, the Fart Joke's sounds) - is
  `""` (the master) or a percentage of it (`"40%"`). Turning the master turns them all. A bare
  number (0.4) still reads as 40%.
- **Current shares:** the weather sound 29% and the background loop 14% of the master; `grunge.wav`
  at 40% to sound as loud as `dJent3b.wav`; the Fart Joke's danger sound at 25%; its farts at
  80%; numpad errors at 15%.
- **Evened sounds (loudness):** every sound is measured for average loudness (ffmpeg EBU R128),
  turned to `chat_audio.loudness_lufs` (-16), then limited under `chat_audio.normalize_db` (-3;
  the limiter at -3.5 so the true peak stays in). Made voices are done in place as they're made;
  stings, song sounds, soundtracks and the weather mp3s get done copies in
  `webroot/gopod_audio_norm/` that the page plays (`normSrc`, `webroot/gopod_audio_levels.json` "made",
  `gopod_audio_levels.go`); originals untouched. `normalize_db` 0 = off.
- **Change from the chat:** `system-set master-volume 35%` and the named volume settings
  ([14.3](#143-editing-from-the-chat)).

---

## 12. Managers: CHALK and PLAYHEAD

The managers take load off the Brobots. They are GOPOD's other chat participants (Gomads):
`managers` in the settings, `gopod_managers.go`.

### 12.1 Three personas, three jobs

- **The Brobots' shared persona** = GOPOD reps. No system jobs.
- **The reporters' shared persona** = scripted lines and voices. Reporter 1 and 2's lines are
  the scaffolding for CHALK and PLAYHEAD's live chat personas.
- **The managers** take the system jobs.
- **The four hosts, one persona shape** (RULE, 2026-10-21: Brobot 1 = Doc, Brobot 2 = Pip, the
  GOPOD hosts; the Brobots = party bros, GOPOD demo reps). `host_personas`: `name` (said only
  when his base text doesn't already say it), `role`, `steps`, `signature`, `handover`. Added
  after his `personas` line (Brobots) or `managers.personas` prompt (CHALK, PLAYHEAD)
  (`gopodHostPersonaText`, `gopod_brobot_personas.go`). Doc delivers the point (read the
  question, name the point, build it from the first rule up, end so it couldn't be any other
  way). Pip is the audience's pied piper (catch the point, ask the shy question or guess wrong
  the real way, catch up and say the wow, hand over; signature "raise your hand if you got that
  wrong too!"). Doc and Pip never name each other in the prompt ("my brother").
- **Pre-build context** (2026-10-22, `prebuild_context`, `gopodPrebuildContext`): every Brobot
  answer's prompt gets "BEFORE I ANSWER:" after his facts - I am Brobot N (Doc / Pip); my role
  is a GOPOD rep; my reminders-list (the same names as his "Hey, that reminds me" line); what was
  asked of me; how to best answer (his `steps`). A line with an empty token is left out.
- **Pairs and the stereo dial** (`host_pairs`: Doc & Pip, CHALK & PLAYHEAD): `partner_as`,
  `banter`, and `dial` - like a stereo balance, 0 = all left (the base: `left`, party-bro GOPOD
  demo rep / scripted reporter), 50 = pointing up (50/50), 100 = all right (the host persona).
  Both at 60 = 40/60. Said in the prompt as "My stereo dial: 40% party-bro GOPOD demo rep, 60%
  Doc."

### 12.2 Who does what

| Manager | Voice, look | Jobs |
|---|---|---|
| **PLAYHEAD ZUMINTU** (KP8) | Reporter 2's voice, tape-brown bubbles | Time and place: "weather" - "I'll help you with that.", the weather line; "what time / where are we / WAWN / what's the date" - the time and place line (`time_place_text`); search results; location QR codes. His job role is in his persona prompt (`managers.personas.PLAYHEAD.prompt`): "ph" for PLAYHEAD, a/0/b the working points to zoom in and out from, ZUMINTU = zoom into, like editing an audio file end to end. |
| **CHALK BORED** (KP7) | Reporter 1's voice, chalkboard bubbles | Display: QR codes, the surprise eggs (Roast Battle, Halloween, camera feeds) and the visual board, drawn on its board (`display_host`). |

- **Talking to a manager:** KP7 or KP8 (voice only): your line, then the persona's answer from
  the language model - its identity and its reporter's scripted lines as the style - in its
  reporter voice and colour (`managers.personas`, `gopod_personas.go`). "Who are you?"
  (`who_words`) gets your own line, word for word (`who_line`).
- **A manager is still a participant:** a persona's answer gets the Brobots' shared chat
  (`gopodPersonaMessages`).
- **Takes:** a take's words (weather, time/place) asked of CHALK or PLAYHEAD go to the take
  (PLAYHEAD's real weather), not the persona's language model.
- **Manager bubble:** a manager's next line joins its bubble from the exchange (its own bubble, or
  its part in a Brobot's) (`hostLineInto`). Managers only - reporters' bubbles are unchanged.

### 12.3 Managers inside a Brobot's bubble

Asking CHALK, PLAYHEAD, the Brobots or All is a take-over trigger. Asking Brobot N is not a
direct call to a manager, so the managers' parts sit inside Brobot N's bubble:

- **A line to Brobot N:** PLAYHEAD's part (weather, time/place, search results, a location QR)
  opens his bubble at the top, before his "Got it!" (`nestShell`, `nestBox`). His reply fills the
  same bubble below. CHALK's QR codes go in at the bottom, after his response (`chalkQrs`), each
  with its name (the call to action's too). His own highlight ignores the nested parts
  (`ownSays`).
- **To the Brobots or All:** the parts go inside the bubble of the Brobot answering first. Once
  the managers have dealt with him, their jobs are done; the other Brobot answers like before, no
  QR note.
- **To a manager:** the managers' own bubbles.
- **Newsletter:** CHALK's part with the `qr_codes.second` QR and its title ("Newsletter:
  crushn8r.ca/subscribe"), inside the Brobot's bubble, on the inner side of the newsletter line -
  after it when it opens the answer, before it otherwise (`chalkCta`). Any newsletter line counts,
  not only the call to action's own words. CHALK QR boards are top/bottom border phobic inside a
  Brobot's bubble.
- **Location QR:** a web address in the line - PLAYHEAD says "I can help with that. QR code
  generated." and shows a QR made for it (`url_qr`, `webroot/qr/made/`); the Brobot's prompt says
  it's shown (`qr_note`). When the first Brobot's answer ends, CHALK says "I can help with that.
  QR code displayed." and shows the same QR (`qr_display_text`, `gopodManagerQRShow`) - at the
  bottom of his bubble for a line to him, her own bubble otherwise - then the other Brobot answers
  or doesn't. The QR is made with the operator's own Python libraries (`~/.local`), since
  Wire-Pod runs as root.

---

## 13. Eggs

An **egg** (easter egg) is a prize - pictures or a card - that pops up when a line's words find
it. The Brobots are GOPOD reps, and eggs are how they show GOPOD's features ("AI Wordplay!
Explain the Math!" is GOPOD as a whole).

### 13.1 Two kinds of eggs

- **Chapter eggs** are the live run's chapters. Each chapter is one egg, titled by its chapter
  title ("Easter Egg '<chapter title>' found."), found by its topics' words (the topics in a
  chapter are its triggers), with those topics' pictures, then the value points. No picture = no
  pictures and no zoom line. They're on the egg list: the reminders name the ones not found yet.
  (`gopodTopicChapter`, `TestChapterEggs`.)
- **Surprise eggs** (`topic_pictures.not_listed`: Roast Battle, Halloween, camera feeds) pop up
  when their words come up but are never named. The roast battle (in no chapter) is its own egg.
  Halloween: `_always/03_halloween.txt`, an emoji card (`cards.halloween`), no picture; its say
  lines are drafts. Camera feeds: `chapters/_always/04_camera_feeds.txt` - "camera feed(s)" shows
  the Brobots' live camera streams on CHALK's board (a picture path starting with "/" is used as
  it is).
- **Explain the Math** is the livestream show's egg ([8.5](#85-explain-the-math-the-contest-format)):
  "AI Wordplay Explain the Math" - "[contest submission text]. Explain the Math" - with
  "🏆 Win to get Doc's approved 'Almost Approved' top prizes!" as its prize (`math.egg_name`,
  `popup_title`).
- **Not eggs:** Weather and Get Photo (the weather face shows in the Brobot's own bubble; the
  photo's 3 pictures in a plain picture bubble). `topic_pictures.not_eggs` (the chapter
  "Websites, themes and the channel") never pops up and is not on the egg list; the chapter stays
  in the interview.

### 13.2 When eggs pop

- In the live chat only - never while a draft plays or any run is going (a song, the interview, a
  replay). Locked by `TestPopupsOnlyLive`.
- Each egg shows up to `topic_pictures.max_per_session` (5) times per Wire-Pod start.
- A trigger inside another topic's longer trigger doesn't count (`gopodTriggerOwnMatch`). One egg
  per line, not two, from one phrase.

### 13.3 GOLDEN: the egg process

- Every egg a line finds shows in the order its word comes in the line, each its own egg bubble.
  Two eggs are two egg bubbles.
- **The egg bubble:** the Brobot's bubble (his colours) ringed by the marquee, holding
  Goverlord's grey block ("Goverlord: GOPOD Brain", then `topic_pictures.egg_text`, two lines:
  "Congratulations!" / "Easter Egg '{name}' found." - {name}: the egg's name, else the word that
  found it; a colon instead of the period before a card, its card lines under it), then the
  prize, then `egg_zoom_text` "MouseOver to Zoom In" on its own line under a prize that zooms (a
  picture, a photo). No header above it.
- Goverlord says the congratulations as it shows, revealed as it's said (`voiceLitLine`), zoom
  kept, after the question has been read out, before the next egg. The zoom line, said last in
  the same voice, lights with it too (2026-10-23): lit once the congratulations' last word is
  said, shown said when the voice ends.
- **Order:** the prize and the spoken congratulations, then the thinking line, then the reply.
  His reply comes after, in his own marquee bubble with his "Brobot N:" header.
- **To All:** when the question found an egg, every Brobot's answer wears the marquee - each in
  his own marquee bubble with his "Brobot N:" header - until the next operator line.
- **The marquee:** bulbs ring the egg (always animated, even with the PC's reduce-motion
  setting), placed along its rounded border and cycling Chase, Groups, Pairs and Twinkle every
  5 s; they're laid out again whenever the bubble changes size, their count following the
  border's length. The marquee ends after a voice (backpack) question's exchange and after a
  reload.
- A scripted run shows its pictures without eggs.

### 13.4 Egg names, cards and triggers

- Egg names by topic: `topic_pictures.egg_names` (Birthday - the Hot Wheels box; Fractions - the
  Doc & Pip workbook; Roast Battle; Mobile Gear - the GOPOD stack picture, found by the rig's
  words like mobile and gear, not by "ai"), else the word that found it.
- An egg with no picture shows a card (`cards`: the Roast Battle's "Party Brobots vs any and all
  takers").
- **Granny:** "gangsta" finds her egg too.
- **Chocolate is food, bingo is a game:** "chocolate bingo" is the food topic's words
  (foodmath_recipes, the recipes picture), so its "bingo" does not load the bingo topic too.
  "bingo" alone is the bingo game egg (the Chocolate Bingo kit and card pictures): bingo hosting,
  then "Hey, you can also play Chocolate Bingo...". The Bingo line is "Sure, let's play! Say start
  Bingo test!"; "experience" is not a Bingo egg word.
- **Birthday:** value point 4 - Brobot 0 wants to ask a local bakery to make him a big cake based
  on his foodie cars.
- **Brobot 0:** the topic is found by "crushn8r", not "ron". The flare4.com QR shows in a bubble
  about his past ("nightlife", "Mississauga").
- **Websites:** the websites topic's picture is the crushn8r.ca QR (`pictures/crushn8r_ca_qr.png`),
  in chapter 9's egg.
- **Doc & Pip workbook cards:** the six cards (1A Factoring Warm-Up, 1B One Big Slot, 1C Speed
  Round, 2A Doc's Answer Key, 2B Flip to the back, 2C No credit for close), cut from
  DocPipMath.pdf's pages turned upright, are `webroot/pictures/docpip_*.jpg` - the workbook egg's
  pictures (`04_how_doc_and_pip_teach/02_doc_and_pip_workbook.txt`), each with hand-written points
  from its own words (`gopod_picture_descriptions.json`).
- `system-check` warns when a topic's trigger word is, or holds, a Wire-Pod intent phrase
  (`intent-data/en-US.json`) or a GOPOD command phrase (`native_cube.intents`).

### 13.5 The egg list and reminders

Every chapter egg with a prize (pictures or a card), in the chapters' order, then
`value_points.reminder_eggs` (Explain the Math). An egg found since Wire-Pod's start leaves the
list. Each "Hey, that reminds me..." names at least one egg still to find; a question on no topic
still ends with two (not a thank-you); each in turn, then round again. `gopod_egg_list.go`,
`TestEggList`.

---

### 13.x Found eggs give a summary

The first time an egg is found, its pop-up shows outside the Brobot's bubble, and a small summary
shows inside the answering Brobot's bubble, at the top: "Easter Egg '<title>'" and one line. Every
later time the same egg is found, only the summary shows. The line is the chapter's own summary
(`interview_chapters`) for a chapter egg, the card for a card egg, else the topic's first `say`
line. (`gopodEggSummaryPost`; the bubble notice template: `gopodBubbleNotice`, page
`bubbleNotice`, marker `notice:<title>|<line>` - also the native intent notice.)

**A spoken question goes above its answer** (2026-10-22). Wire-Pod logs a spoken (backpack)
question only when the answer is done; the page puts it above that Brobot's open bubble when the
bubble was made after the question (`dataset.asked`, `operatorSeq`), as it already did for an egg
pop-up.

**The multiple responses template** (2026-10-22, `answer_chains`, `weather_lines`,
`chain_prompts`): a line to both -
- first Brobot (`brobots_all_first`): "Brobot Weather Report for <place>, <date>. Temperature
  is <temp> and <sky> as of <time>." - his weather comment - his value point - the hand-over.
  No reminders.
- next Brobot (`brobots_follow_up`): "Confirmed. Temperature is <temp> at <time>." - a comment
  following his brother's - his value point - the Weather Kit - the reminders.
- a single answer (`brobots`): the report - the comment - the value point - the reminders.
Each Brobot in his own unit and clock (his robot settings: Brobot 2 F / 12-hour, Brobot 1 C /
24-hour). For a weather question with a value point the model is asked for two paragraphs:
the weather line, then the point (`chain_prompts`). The typed line's All is a 1/2 menu - who
answers first (`typed_line.all_first`, default 1; 2 = "Starting with Brobot 2"); the All radio
stays, the menu beside it. Hand-over
nicknames: gear-head, cranky-shaft, coolant-leak, incompatible-particulate, fumble-nuts.

**Core topics, background, eggs** (2026-10-22, CORE_TOPICS_002.md): seven core topics - Who we
are, What GOPOD is, How Doc & Pip teach, Book us, The contest, Join the crew, The comeback.
Every other topic is background (`value_points.background_topics`): the Brobots still get its
facts, but it's never a value point - none given, not in the score, never reminded. Eggs
(`topic_pictures.eggs_keep`): Chocolate Bingo (the Bingo and foodmath chapter), Gangster Granny
(the Legacy and wrap-up chapter), Explain the Math, Roast Battle (18+); no other topic's pictures
pop up.

**Topic groups** (2026-10-22, `topic_groups`, `gopodTopicGroupFor`): a value point from one of a
group's topics gets that group's answer chain. `brobots_math` (fractions_and_decimals): the
exact math line, the worked steps (the model), "Doc's Rule: no credit for close." (`rule`
part, not when the point already is the rule), the next points, the hand-over in a line to
both.

**Which value point** (2026-10-22, `gopodPickPoint`, `value_points.pick_skip_words`): of a topic's
points not said yet, the one whose words (the point, its notes, its say line) share the most with
the question - longer words count more; filler words don't count; the topic's own name doesn't
pick a point (the whole name said, or a one-word name's start: "book" for booking). With nothing
shared, the next in order - the pull stays. ("Tell me about Pip" -> Pip's point; "My granny
lives there" -> Gangster Granny.) Locked by the golden interview questions
(`TestHarvestQuestionsGetTheirOwnPoint`).

**Exact brand strings** (2026-10-23): "AI Wordplay! Explain the MAth!" - written MAth everywhere the
Brobots say it (`reply_tighten.brand_words`, the "robot" -> "Brobot" template; a phrase ending
in a mark now matches too), said "math" (the Brobots lingo: `MAth` = `math`).

**The game: 20 value points, one session score** (2026-10-23, FINISH_PLAN_001.md A3,
A3_CORE_POINTS_001.md): "AI Wordplay! Explain the MAth!" is a game. `value_points.core_points` -
his 16 value points (A3_CORE_POINTS_002.md; the seven core topics in `core_topics`: Who we are,
What GOPOD is, How Doc & Pip teach, Book us, The livestream, Join the crew, Due Credit); only
these are given, counted and pulled toward, every other point stays known. Renamed topics (his
names): the_rig -> mobile_gear, party -> bookings, contest_wheel -> AIWP_livestream; Gangster
Granny is its own topic, due_credit; niche_pillars is new. One score template,
`value_points.score_text`: "Topics N of 7 · Value Points N of 19 · Eggs N of 4" (hidden eggs
count too) - in the notice, in the Brobots' pre-build
context ("Our score so far: ..."), and in "what else". The rules for the Brobots: `game.rules`
("The game: ..."). Grand prize: every value point revealed (`target: points`). The point mechanics
tests run over every point (`gopodAllPointsForTest`); the answer shapes run with the 20.

**Fixes from the duo run** (2026-10-23): duo B gets no manager step-in (`duo.b.no_manager`), so no
second "Thanks for your help"; "Thanks for your help, PLAYHEAD." is lit while he says it
(`noticesay:` / `noticesaid:`); an egg's congratulations wait as one Generating line; its "MouseOver
to Zoom In" is lit word by word at its own words' times in the same voice; the hourglass drains
smoothly and loops (`hourglass`); a country alone is its capital, continents aren't cities
(`clean_politics.capitals`, `continents`; Zimbabwe 🇿🇼 added); "bro but to" = Brobot 2; "live" is
said "lyve" in "live stream", "go live" and the like (the Brobots lingo); a point's title line
must say its title; a scripted line may open or close any answer shape.

**Win and reset** (2026-10-23, FINISH_PLAN_001.md A4, `game.win_text`, `game.reset_on`,
`gopod_game.go`): every value point revealed - Goverlord says "Grand prize! Every value point
revealed. <score>. Now GOPOD Yourself!" once per session, after the answer that revealed the last
one has been said. The "thanks" ending resets the game: score, points said, reminders, chapters
cleared and eggs found all back to 0 for the next session. Locked by `TestGameWinAndReset`.

**The summary answer type** (2026-10-23, `summary`, `gopod_summary_answer.go`): a golden query
("What is AI wordplay?", `summary.queries`) or a question opening 3+ core topics gets one bubble
per core topic still to come - its next value point's title and first note - up to 3, the
livestream first ("AI Wordplay! Explain the MAth!" live stream: the content engine). No point
counted; the notice says "Summary: ...", the pull names what's next. Chain `brobots_summary`
(`llm`, `summary_bubbles`, `reminder`); a bubble the model misses is made from the notes.

**Funnels and golden answers** (2026-10-23): `value_points.funnels` - his funnels as lists of
topics in order (livestream > content engine > social media > websites > contact/subscribe;
livestream > book us live > subscribe; livestream > book a cameo > subscribe); the pull follows
the next step of each, then `topic_links` - reorder by editing a list. `golden_answers`: "What is
AI wordplay?" / "What is Explain the MAth?" get his line word for word, first (`golden_line`
part), then the answer's type. New topic `cameo` (Book us): a birthday video from Doc & Pip
(crushn8r.ca/booking1.html). gopod #1's notes: his "What is GOPOD?" words.

**Duo answer types** (2026-10-23, `duo`): the boilerplate for a line to both - A answers first,
B next, each its own chain (`duo_a`, `duo_b`) and the rules every answer type follows there: A
ends with the hand-over in place of the reminder (math keeps its reminder, then hands over); B's
weather is "Confirmed...". Who's A, in this order: the words ("Starting with Brobot 2"), then the Brobot named first
anywhere in the line ("for both of you bro but to, ..." -> Brobot 2; `gopodFirstNamed`, the page's
icons alike), then the All menu (`typed_line.all_first`). A multi answer type later = more letters in `duo`. Soft-locked by
`TestDuoAnswersHold`.

**GOLDEN answer shapes** (2026-10-23, FINISH_PLAN_001.md step A2, `answer_shapes`,
`gopod_answer_shapes.go`): one shape per answer type - the order of his thought bubbles:
- weather: Brobot Weather Report, his words, his point, the reminder
- weather (next in a line to both): Confirmed. Temperature..., his words, his point, the Weather
  Kit, the reminder
- topic: his point (title line, then its notes as his words), the reminder
- free: his words, the reminder
- math: the exact math line, his words, Doc's Rule, the reminder
- thanks: his words only (no hand-over either)
Single-Brobot output (to Brobot 1 or 2) and multi-Brobot output (All, starting with 1 or 2: the
first answer ends with the hand-over in place of the reminder; math keeps its reminder, then hands
over). Every answer is checked as it's made - `BROBOT_SHAPE ... ok`, or `DRIFT` and the shape it
wanted - and every type at every place (single 1, single 2, All-1 first/next, All-2 first/next)
runs through the real answer chain at deploy (`TestAnswerShapesHold`).

**Point bubbles** (2026-10-23, `point_bubbles`, `gopod_point_bubbles.go`): a value point with notes
becomes his thought bubbles - a hype line for its title, then one party line per note. The model
is asked for exactly that; a missing line, or one sharing no word with its note, is made from the
note ("{Note}."), so the shape always holds. His weather comment, when there is one, stays first.
Locked by `TestPointBubblesShapeAlwaysHolds`.

**The pull's map** (2026-10-23, `value_points.topic_links`): after a topic, "Next up" and the
reminder name the first open topic it links to - the repo's lanes
(`GOPOD_NICHE_BUZZ_PLANNER_001.md`: contest -> websites + social; websites -> newsletter +
contest; social -> newsletter + contest; gear -> videos) and recipes > Chocolate Bingo > foodmath;
with none open, the most alike topic.

**The gravitational pull** (2026-10-22, `gravity`): GOPOD is a game of query for value points -
however far it's pushed, every answer steers back to the points still to say, in three places
that name the same points (`gopodReminderLine`): his pre-build context (his reminders-list), his
notice ("Back on track:" / "Next up:"), and his reply's end ("Hey, that reminds me...").

**The 5 bubble kinds** (2026-10-22, `bubble_kinds`, page `tagBubbleKind`): Goverlord
orchestrates, gomads think, gominions do. Every bubble in the chat gets one kind as `data-tier`
- operator (input), goverlord (system), gomad (CHALK, PLAYHEAD - inside a Brobot's notice unless
called directly), gominion (Brobot 1 and 2 and their output), other (eggs, pictures, chapters,
countdown, narrator) - by its classes, the first rule that matches. A bubble with no kind is
logged once: "GOPOD_CHECK bubble without a kind (loose chain)".

**Names, three layers** (2026-10-22): base = the robot by ESN; Wire-Pod = Brobot N; GOPOD = the
persona name, a chat participant. Brobot 1/2 -> Doc / Pip (to be); Reporter 1/2 (no robot) ->
CHALK / PLAYHEAD - their scripted voices, and chat participants, GOPOD reps too. **The header
template** (`display_names`, page `hostHeader`): a reporter's or manager's bubble shows a
remembered edit (`edits`, e.g. CHALK BORED), else the default (CHALK / PLAYHEAD), else the
fallback from the numpad map (`Brobot {n}`: KP7 = Brobot 7, KP8 = Brobot 8). The default itself
comes from the numpad map (its persona for that voice: `managers.personas.<name>.reporter` links
Brobot 3 -> CHALK, Brobot 4 -> PLAYHEAD; `persona_reporters` on `/api-sdk/gopod_numpad_map`);
`display_names.default` only overrides it. Scripts keep their
own names (Reporter 1/2); only the chat shows these.

**GOLDEN: the Brobot Bubble template** (2026-10-22, operator's spec; Brobot 1 and 2 alike):

```
===== Brobot Bubble =====
header: Brobot N:                         (+ zoomable weather picture, top right)
=== notice: Value Points ===
1. Got it!                                (said by him - the communication received)
2. <topic> #<n>: <point>  /  No value point - a quick take, then back to GOPOD.
   a point with notes (its topic's "notes N: title | note | note"): <topic> #<n>: <title>,
   then -<note> lines; the model gets the notes and parties on the idea - no exact line is said
   for it (2026-10-22: "only the model's flavoured version of the notes"); exact math
   (fractions) and the Brobots' instructions keep their own words
   (Web search: <search>; Easter Egg '<title>': <summary> - one notice per bubble)
   (game: Combo: <topics> / Chapter cleared: <chapter> / Back on track: <topics> (free
   answer) or Next up: <topics> (topic answer) / Value Points <said> / <total>)
3. (optional) [icon] PLAYHEAD [icon] I'll help you with that.   (one row; his searching waits)
4. Thanks for your help, PLAYHEAD.        (said by him - the floor back to him after a manager's part)
=== end notice ===
<his answer-chained reply>                (a weather question: it opens with the golden weather line)
===== Brobot Bubble =====
```

- The order is one settings template, `brobot_bubble.steps` (header, got_it, value_points,
  manager, thanks, reply); manager and thanks can be left out.
- Line 3 (2026-10-22, live: PLAYHEAD came only after the whole answer was written, so the
  Brobot searched, PLAYHEAD spoke, the Brobot searched again): a manager steps in right after
  lines 1-2, while the answer is written (`gopodManagerStepIn`, once per line), its lines in
  this Brobot's notice (`gopodManagerDoFor`). The Brobot waits until the manager is done, even
  with his answer ready.
- Line 1 and 2 are posted together once his "Got it!" is said and his prompt is made
  (`gopodValuePointsNoticePost` / `gopodValuePointsNoticeGotIt`). Line 3 is a manager's top part,
  put in the notice by the page (`nestBox`, `.in-notice`). Line 4: `got_it_again`, "Thanks for
  your help, {helper}." ({helper} from `helper_from`: the manager's name, PLAYHEAD / CHALK; else
  its Brobot, Brobot 4 / Brobot 3),
  said and shown only after a manager's part; `second_got_it` `after_manager` or `off`.
- A manager's part in the notice is one plain line, no bubble of its own: "[icon] PLAYHEAD [icon]
  I'll help you with that." (`.in-notice`).
- A manager's line uses the Generating template like every voice: "Voice file generating." then
  the hourglass while its voice is made (once - "I'll help you with that." is reused, so it
  skips straight to its words), then its words lit to its voice.
- Weather: PLAYHEAD says only `take_text` (`brobot_says`); no weather voice is made; the
  Brobot's reply opens with the golden weather line. Its picture is pinned for later
  (`weather_card.picture: false`: no picture, no pause). Its sound plays once as his bubble shows (a
  background effect - he doesn't wait for it). Once he has said the line, its picture shows
  centred under it (Wire-Pod's `weathershow:`, page `weatherShowFor`) and he waits
  `weather_card.pause_seconds` (3) before the rest - the template for any later "wait for a
  picture or sound" spot (`gopodWeatherLineSaid`). After its sound has played once, a click that zooms the picture plays it again -
  never two at once. Hovering still zooms it without sound.

**Notices inside a Brobot's bubble** (2026-10-22, standard). What goes in each is one template,
`bubble_notices`: a title, then its lines, each on its own row; a notice with the same title takes
more lines underneath. `value_points` - every live Brobot answer opens with it: the value points
his answer is built from (`{id}: {words}`, id = topic#n), then weather, search, or "none", then
the "Got it!" line once he has said it (`chat_moves.thinking_text`; in the menu under 3).
`egg` - a found egg's summary. `native_intent` - a native Wire-Pod intent.
(`gopod_value_points_notice.go`.)

### Value points - golden and locked

Value points are what GOPOD sells; chapter > topics > value points. A value point is each numbered
line (`1. ...`) after a topic file's `---`, or - for a topic with none - its whole text as one point.
Each has an id, `<topic>#<n>` (e.g. `bingo#1`). Their words are locked in
`gopod_probes/chapters/value_points_golden.json` with a checksum (written the first time Wire-Pod
starts without one). `system-check` flags a point whose words changed, a new point not locked yet,
and a locked point that's gone. `system-value-points` shows the state, `system-value-points
<topic>` lists a topic's points, `system-value-points lock` locks changes on purpose
(`system-undo` puts the old index back). (`gopod_value_points_golden.go`)

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
menu bubble ([9.9](#99-the-system-window)). It leads you like a tour guide: groups in tour order,
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
| `system-search <words>` | web search results in the terminal bubble ([8.8](#88-search)) |
| `system-chats <words> [operator\|brobots\|managers]` | searches every saved session, matches marked >>like this<<; skips the searches themselves |
| `system-chat-save [session]` | saves `chat.md` in the session folder |
| `system-help` | every system command, a line each |

`system-locate` is read only. `gopod_system_locate.go`, `TestSystemLocate`.

### 14.5 Status, checks and tests

**`system-status`:** Wire-Pod, the language model (loaded or cold), the warm-up, the chat page
(here/away), the run going, and each Brobot's last battery reading.

**`system-check`:** looks for every file the data names and every song step - no changes made.
- **What it checks:** every file named in `gopod_brobots_settings.json` (sounds, pictures, QR
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
| `system-native on\|off` | Wire-Pod's own answers on or off ([7.4](#74-native-wire-pod-answers)) |
| `system-numpad on\|off` | numpad push-to-talk on or off; `system-numpad` alone lists the keys |
| `system-skip` | skips the reporters' warm-up lines ([6.7](#67-skip)) |
| `system-voices-rebuild [go]` | rebuilds every voice ([10.2](#102-voice-files-made-once-play-many)) |
| `system-menu-theme <preset>` | the menu's look |

### 14.10 Every command at a glance

| Command | Section |
|---|---|
| `start interview` / `start bingo test` / `start baby brobots` / `start fart joke` | [6.2](#62-runs) |
| `stop run`, `stop session`, "stop nap" / "stop baby robots" | [6.6](#66-stopping-a-run) |
| `replay <name>` | [6.8](#68-replay-a-session) |
| "skip", Esc, numpad \* | [6.7](#67-skip) |
| "thanks brobots" / "thanks guys" | [6.10](#610-the-thanks-ending) |
| `search <words>` / `look up <words>` | [8.8](#88-search) |
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
| **Change a setting** (volume, waits, texts) | `system-set <setting> <value>` on the chat keyboard, or edit `webroot/gopod_brobots_settings.json` (`system-locate <word>` finds it) | page: F5; Wire-Pod: `wpr 2` |
| **Add a picture to an egg** | put it in `webroot/pictures/`, add its name to the topic file's `pictures:` line (`gopod_probes/chapters/<chapter>/<topic>.txt`) | `wpr 2` |
| **Add a topic** (new egg words, value points) | a new `NN_name.txt` in its chapter folder: `label:`, `words:`, `pictures:` (optional), `reminder:`, `say 1:`, `say 2:` ..., then `---` and its reveal text | `wpr 2` |
| **Add a chapter** (a new egg) | a new folder `NN_name/` in `gopod_probes/chapters/` with `chapter.txt` (`title:` = the egg's title, `summary:`) and its topic files; the number sets its order | `wpr 2` |
| **Change a song line** | its `## STEP` / `> TEXT:` in the song's `story.md`, or `system-set <song> <step> text <words>` | next run |
| **Add a song step** | a `## STEP <id>` + `> TEXT:` in `story.md` **and** a step with the same `step_id` in `knobs.json` (`system-check` says if one is missing) | next run |
| **Add a sound** (sting, song sound, soundtrack) | put the `.wav` in `webroot/gopod_reporter_audio/`; name it in the settings (a sting) or a song's `knobs.json` | evened to the loudness rule by itself |
| **Change how a word is said** | the pronunciation registry in `brobots_wirepod_interview_section_card_template_1_001.md` (Brobots), `reporter_voices.say_words` (reporters), `system_say_words` (Goverlord), or `system-lingo` | `wpr 2` |
| **Add a QR code** | the picture in `webroot/qr/`, an entry in `qr_codes.codes` (words that show it) | F5 |
| **Video** | not supported yet in the chat | - |

---

## 16. Settings file reference

All in `goverlord/wire_pod_overlay/chipper/webroot/gopod_brobots_settings.json`. Values are
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
| `reporter_voices` | `python`, `worker`, `speed`, `operator_voice` (`af_sky`), `say_words`, `generating_text`, `generating_marker`, `start_wait_seconds`, `heard_margin_seconds` (3), `finish_wait_seconds` (120), `max_wait_seconds`, `narrator_lead_seconds`, `operator_keep_days`, `scripted_keep_days`, `live_keep_days`. |
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
| `system_search` | The search settings; `results_role`, `results_host` (PLAYHEAD). |
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
   `gopod_brobots_settings.json`.
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
    orchestrator. See [The session and runs](#6-the-session-and-runs).
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

---

## 18. Troubleshooting

| You see | Cause | Fix |
|---|---|---|
| Mic or numpad don't work | Not found by `lsusb` | Check the parts by name ([1.6](#16-desk-start)); Goverlord says the mic is missing after the warm-up and the numpad works dry. |
| A deploy didn't take | Wire-Pod started before the deploy | `wpr 2` again, then F5. |
| Nothing plays after start | The page isn't here | Click **Log**; open the chat; keep the tab visible ("Goverlord awaits your return..."). |
| A hand edit to `customIntents.json` was lost | Wire-Pod's web config page rewrites the whole file | Restart Wire-Pod after a hand edit, before touching that page. |
| A step plays the wrong thing | `zKnobs.json` wins while it's there | Check or remove `zKnobs.json`. |
| `STEP_NAME_WARNING` in a run | The step name doesn't read as its knobs | Rename it to match ([2.2](#22-step-names)); it only warns. |
| Deploy stopped by a test | A song's two files name different steps, or an unknown note | Fix the named step in both files. |
| The opener's first lines are shown and never said | A voice asked for while Kokoro was starting | It now waits up to `start_wait_seconds`; check the voice worker. |
| A bubble holds the next one up | A frozen robot or a lost end signal | It lets go after `chat_window.speaking_max_ms` (30 s); check `GOPOD_PAGE held` and `GOPOD_HEARD` in the journal. |
| A reply bubble missing or doubled | | `system-find FEED_CHECK`; see [9.8](#98-the-engine-and-the-page-in-dialogue). |
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

**The owner's working notes** (outside the repo):
- `gopod_notes/RUN_INTEGRITY_REVIEW_001.md` - the review behind the chat rules
- `gopod_notes/ENGINE_BRIDGE_LOOK_SURVEY_001.md` - the engine and page dialogue
- `gopod_notes/CHAT_WINDOW_EDITING_LISTS_001.md` - the editing lists
- `gopod_notes/PAGES_AND_VIEWER_ABSORB_SURVEY_001.md` - the old pages and the JSON viewer
- `gopod_notes/PLAYHEAD_ALIASES_ABSORB_SURVEY_001.md` - pha0b and phcal in the chat
- `gopod_notes/GOPOD_SHEET_MUSIC_001.md` - the sheet
- `gopod_notes/older_notes/SLEEP_BENCH_ALIASES_001.md` - the sleep beats

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

GOPOD — [README](README.md) · [WIRED-POD](WIRED-POD.md) · [channel](CHANNEL.md) · [ask](ASK.md)

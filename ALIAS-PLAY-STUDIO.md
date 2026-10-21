# ALIAS-PLAY-STUDIO

> The studio where two Vectors learn a show: one small proven move at a time, strung into
> sequences, locked into songs. Take Out part 1 in [WIRED-POD.md](WIRED-POD.md).

This is the one place that says how a GOPOD song is built: the words, the moves, the rules
that were proven live, and the standard every song follows. If a song folder and this page
disagree, fix one of them.

---

## The three words

| Word | What it is | In a song folder |
|---|---|---|
| **Golden note** | The smallest move or line the robots can play, proven live on its own first. | One step's `note` type in `knobs.json` (`arm_cue`, `rattle`, `say_turn`, ...). |
| **Sequence** | A few notes in order that make one moment. | One `## SECTION` in `story.md`, the steps' `section` in `knobs.json`. |
| **Golden sequence** | A sequence, or a whole song, that's locked: it plays the same way every time and is ready to record. | A song whose status line says `GOLDEN`, or a locked opener (the Bingo test's `OPENING SYNC`). |

The piano is the notes, the mixer is how each one plays (its hold, speed, buffer, tempo), and
the sequencer is their order. All three live in the song folder: no song's moves are hidden in
code.

---

## The lingo (operator and session)

The words the operator and a session both use, each pinned to the one thing it means. When
one of these is said, this is what's meant; nothing needs re-asking.

**On stage (the chat page):**

| Say | It means | Where it lives |
|---|---|---|
| **Operator bubble** | A line to the Brobots, white-eyed icon, labelled by its input: "Keyboard to Brobot 1", "KP1 to Brobot 1", "Backpack to Brobot 2" (heard by the robot himself: Hey Vector or the backpack button), "Reporter 1 to Brobot 1" (a scripted run's question), "… to Brobots" (both icons). Goes to the top once the bubble on screen has finished (the baton, rule 14). | the chat |
| **System bubble** | Goverlord's grey-eyed bubbles (warm-up, countdowns, chapters, system messages). Reporter and narrator bubbles are system bubbles with white eyes. | the chat |
| **System message** | A Goverlord line in a grey bubble, said in the system voice ("Brobots. Go to sleep now please.", "Run complete."). | note `system_line` |
| **Brobot bubble** | A Brobot's line, in his eye colours, each piece lit as it's said. | note `say_turn`, a reply |
| **Reporter bubble** | Brobot 3 or 4 (voice only), headed "Reporter 1: Brobot 3" / "Reporter 2: Brobot 4" (the same form as "Goverlord: GOPOD Brain"), white-eyed icons. | the interview, `ask` |
| **Narrator bubble** | A timed song's captions, one bubble per section, each phrase lit as it's read. | `robots_speak: false` |
| **Chapter** | Goverlord's title between parts of the interview. | the interview |
| **Warm-up bubble** | "Goverlord: GOPOD Brain", the LEDs and hourglass, then "LLM Ready". | `llm_warmup` |
| **Countdown** | The ⏳ bubble with juggled balls, under a rattle. | note `rattle` |
| **Big ball** | The Bingo ball bubble. | `show_ball` |
| **Highlight** | The lit piece in a bubble, the words being said or sung now. | golden rule 10 |
| **The baton** | The next bubble waits hidden until the one on screen finishes. | golden rule 14 |
| **Keyboard line** | `Brobots ()1 ()2 ()All \| Native?[] Numpad?[] \| Record? [] 00:00:00` under the chat. | the chat |

**The show:**

| Say | It means |
|---|---|
| **Session** | `wpr 2` until the next `wpr 2`. Goverlord runs it. |
| **Run** | One song, draft or phcal live, from its start to its check-out. One at a time. |
| **`start <song>`** | Starts a run: `start bingo test`, `start baby brobots`, `start interview`. |
| **`stop run`** | Stops the run at its next safe spot ("Run stopped."). |
| **Thanks** | The thanks ending: closes the run out. |
| **Sleep call / wake call** | Both Brobots put to sleep (`sleep_both`) / woken (`wake_both`, or the wake inside `brobots_ready_together`). |
| **Cube call** | The cube blip on a Brobot's own cube: red, 2 s, green. Brobot 2 keeps the cube. |
| **Brobots Ready** | Both Brobots say the line together (`brobots_ready_together`). |
| **Last words** | The song's last sung or spoken words; the ending is timed from them. |
| **Edit mark** | A 0-second `pause`: a place for the reporters' voiceover in the video edit. |
| **Live / dry** | With the robots / with no robots (nothing moves, everything is logged). |
| **Check run / check logs** | Read the last run's log and Wire-Pod's journal; report the bottom line. |
| **Bench test** | One move or tool tried on its own, on the bench (phcal or a one-off script). |

**The workbench:**

| Say | It means |
|---|---|
| **Golden** | Proven live and locked. Changed only on the operator's go. |
| **WIP** | Being worked on. |
| **Lock it in** | Mark it golden and commit it. |
| **Commit / push** | Save to the repo / send it to GitHub. Two separate go's. |
| **`wpr 2`, F5** | Restart Wire-Pod (picks up a deploy), reload the chat page. |
| **Desk start** | `source ~/.bashrc; gorepo; lsusb; wpr 2`, then F5 (below). |
| **`pha0b`** | Plays a range of a song's steps, live or dry. |
| **`score`** | Prints a song's steps, on screen. |
| **phcal** | The bench: tune one move or sound on one robot. |

### Desk start (2026-10-14)

The operator's terminal routine: `source ~/.bashrc ~/.bash_aliases; gorepo; lsusb; wpr 2`, then F5.

- `source` reads only the first file it's given (the second becomes an argument); `~/.bashrc`
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

---

## A song folder

`goverlord/runtime/songs/<number>_<name>/`. Copy [`_new_song/`](goverlord/runtime/songs/_new_song/)
to start one.

**A song is known by its `song_id`, never its folder.** The `song_id` at the top of its
`knobs.json` is its name everywhere: the engine, the `start …` commands, `pha0b`, `score`, the
interview, and the soundtrack copied to the page. Rename, renumber or reshuffle the folders and
nothing else needs updating. Two folders with the same `song_id` is an error, named.
`python3 goverlord/runtime/songs/tools/song_folders_001.py` lists them.

| File | What it holds |
|---|---|
| `story.md` | The words. A status line at the top, then `## SECTION: <name>` blocks, each with `## STEP <step_id>` and `> TEXT: <words>`. A step with no words keeps an empty `> TEXT:`. |
| `knobs.json` | What each step does, in order: its `note`, who plays it, its timing. |
| `zKnobs.json` | Your private working copy of `knobs.json`. When it's there, it wins. Not tracked. |
| `runs/`, `notation/` | What runs write: logs and marked-up scores. Not tracked. |

**The status line** is the first line after the title: `**Status:** GOLDEN (2026-10-01)` or
`**Status:** WIP (2026-10-09)`. GOLDEN means live-confirmed and ready to record.

**The two files always name the same steps.** Changing a line's words in `story.md` is enough
on its own; adding, removing or renaming a step means both files.

**Step names** say who and what: `host_arm_cue_0102` (Brobot 1), `brat_nod_kg_success_0203` (Brobot 2, a combo),
`reporter_gap_0301` (an edit mark). The last four digits are the section and the step.

### Step names, the master list (2026-10-13)

A step name is **a label for people**: the engine never reads it. What a step plays, and in what
order, is its `knobs.json` entry (`note`, or a `combo` list of parts). Renaming a step changes
nothing it plays; the two files only have to name it the same (locked by
`TestSongFoldersFollowTheStandard`). The name is written to match the knobs, in the knobs' order:

`<who>_<part>_<part>…_<SSNN>` - who, then each part as it plays, then section `SS` and step `NN`.

| Who | Is |
|---|---|
| `host` | Brobot 1 (`brobot_1`) |
| `brat` | Brobot 2 (`brobot_2`) |
| `brobots` | both (`wake`, `ready`, `sleep`, `exit`) |
| `reporter` | an edit mark for the reporters (`reporter_gap`), or their in-between lines (`reporter_lines`) |
| `goverlord` | Goverlord, the system voice (the LLM warm notes) |

| Part in the name | Plays (`knobs.json`) | Example |
|---|---|---|
| `say` | `say_turn` - the line; or, when another part says the line (a beat, a nod) and there's no `say_turn`, that part | `brat_say_0204` |
| `arm_cue` | `arm_cue` - the arm lift, with the line | `host_arm_cue_0102` |
| `nod` | `nod` - a head nod, with the line | `brat_nod_kg_success_0203` |
| `<token>_animation` | `animation` with `animation_token` (`searching`…), the line with it | `host_searching_animation_0202` |
| `kg_success` | `animation` with `kgSuccess` (usually the "Lol!") | `host_nod_kg_success_0601` |
| `beat_<token>` | `emotion_beat` with `animation_token` (`angry`, `frustrated`…) | `brat_beat_angry_0503` |
| `rattle` | `rattle` - the line, then the rattle under the countdown | `host_rattle_0102` |
| `thinking` | `thinking` - the live chat's own thinking please wait process (arm, nod, "Thinking, please wait.", searching, get-out) | `host_thinking_say_0202` |
| `wav-<name>` (older: a plain word, `fart2`) | `sound` with `sound` (`fart2.wav`) | `brat_say_fart2_0204` |
| `danger` / `grunge` / `djent3b` | `sound` - Wire-Pod's shared `sounds/` (`danger-will-robinson.wav`, `grunge.wav`, `dJent3b.wav`) | `host_say_danger_beat_frustrated_0205` |
| `img-<name>` | `picture` with `picture` (`granny.jpg`) | `host_say_img-granny_0301` |
| `gap` | `pause` (0 s = an edit mark) | `reporter_gap_0101` |
| `wake` / `ready` / `sleep` / `exit` | `wake_both` / `brobots_ready_together` / `sleep_both` / `exit` | `brobots_wake_0101` |
| `system` / `sleep_message` | `system_line` - a Goverlord line | `sleep_message` |
| `soft_wake` | `wake_sleepy` | `soft_wake` |
| `line_NN` | a timed caption (`pause` with `start_seconds`) | `line_30` (nap) |
| `timer` | `timer` - a timed wait, `seconds` (a `pause`) | `host_timer_0201` |
| `sequence` | `sequence` - another song's steps, in place, by its `song_id` (a song can't sequence itself; an LLM warm song is skipped, played by Wire-Pod) | `brobots_sequence_0301` |
| `anim-<token>` | `animation` with `animation_token` (short form of `<token>_animation`) | `host_anim-kgsuccess_0202` |
| `warm` | `warm_message` - the System llm bubble's message, shown and said; `"wait_voices": true` waits for the voice files | `goverlord_warm_0101` |
| `warm_wait` | `warm_wait` - waits until the language model is loaded | `goverlord_warm_wait_0103` |
| `lines` | `reporter_lines` - the reporters' in-between lines (the opening) | `reporter_lines_0104` |
| `warm_ready` | `warm_ready` - the ready message, shown and said | `goverlord_warm_ready_0105` |
| `sting` | `sting` - the system sting (`system_stings.default`) | `goverlord_sting_0106` |

The new parts (2026-10-16, operator: "extend that to include sequences, animations, timers, llm
warm-up & in-between reporter lines"). The `warm`, `lines` and `sting` notes are an LLM warm
song's (`song_type: "llm_warm"`), played by Wire-Pod (`gopod_opener_song.go`); `timer`,
`sequence` and `anim-` are the song engine's (`run_golden_song_001.py`).

Example: `host_wav-danger_beat_frustrated_0205` reads, in order: `sound` `danger…wav` to its end,
then `emotion_beat` `frustrated` with the line - exactly what its knobs play.

**The name check (2026-10-13):** every song run prints `STEP_NAME_WARNING` for a `_SSNN` name that
doesn't read as what its knobs play, or that has a word not on this list (a typo, "sya"), e.g.
`host_say_danger_beat_frustrated_0205: the name reads say_turn + danger? + emotion_beat:frustrated,
its knobs play sound:danger-will-robinson.wav + emotion_beat:frustrated - 'danger' isn't on the
master list`. It only warns; the run plays its knobs as ever. Names without `_SSNN` (the nap's
`line_30`, `soft_wake`) are free labels and aren't read. Locked by
`test_golden_song_step_names_001.py`.

**Editing by hand (`zKnobs.json` and `knobs.json`):** `knobs.json` is what plays; `zKnobs.json` is the
work copy. Changing only a line's words in `story.md` needs nothing else. Adding, removing or
renaming a step in `story.md` needs the same step name (and its parts) in both json files - the
deploy's tests stop on any step one file has and the other doesn't.

### `knobs.json`, the top

| Key | What it does |
|---|---|
| `song_id` | The song's name to the engine (it must match the engine's song list). |
| `song_type` | What kind of song: `bingo_capture`, `lyric_video`, `interview`. |
| `story_ref` | `story.md`. |
| `global_tempo` | Seconds of space added after each step, times its `tempo_factor`. |
| `chat_intro` | The operator bubble a run opens with ("Bingo Test Live"). |
| `sentence_split` | `scripted`: lines are said piece by piece and the chat bubble lights each piece. |
| `touch_interrupt_off` | `true`: a touch doesn't cut the run short. Back on when it ends. |
| `rattle_on` | Where the rattle plays: `page` (the chat page, like the reporter voices), `speaker` (this machine), `robot` (Brobot 1). |
| `timeline` | `srt`: each step plays at its `start_seconds` (from a subtitle file), counted from the first timed step, instead of right after the step before it. A step that's late goes at once. |
| `robots_speak` | `false`: the robots say nothing; the timed steps' words in `story.md` are a narrator's captions, one Narrator bubble per section, each phrase lit as it's read. |
| `sleepy` | `true`: after `sleep_both` the Brobots sleep. A `stir` (or a move) wakes its Brobot, and he drifts back to sleep, on his own time (the music and captions never wait). They're asleep before `brobots_ready_together`, which wakes both. |
| `media` | The song's video, in a private folder that's never tracked. |
| `audio` | The song's soundtrack, played on the chat page from the first timed line. The file sits in the private media folder; Wire-Pod copies it into the page's voice folder at its start (`page_sound_files` in the settings). |
| `llm_warmup` / `llm_ready_text` | Show the LLM warm-up bubble at the start, and what it says when ready ("LLM Ready for Bingo", said by Goverlord too). |
| `warmup_after` | A step id: the warm-up runs right after that step instead of at the start (no song uses it now: 2026-10-17, operator, the nap: "bubble 2 then bubble 1 * bubble 2 is the start of the song, after llm bubble" - its ready line comes first, then "Brobots. Go to sleep now please."). |

### `knobs.json`, each step

Every step has `step_id`, `note`, `speaker` (`brobot_1` or `brobot_2`) and `section`. Then,
when the note uses them: `hold_seconds`, `speed`, `cycles`, `settle_seconds`,
`buffer_after` (seconds after the step), `tempo_factor`, `pause_seconds`, `gap_label`,
`start_seconds` / `end_seconds` (a timed song), `animation_token`, `show_ball` ("B-1": the step shows a big Bingo ball in the chat as it
starts), `rattle_before`, `pronounce` (`true`: a together line is said with the robots'
pronunciation list, "GOPOD" as "Gowp-awd"), `timeline_after` (on a `pause` in a timed song: the
timeline starts when the pause ends, so the first narrator bubble is sent early, during it).

---

## The notes (the piano)

**Played by the song engine** (`goverlord/runtime/songs/tools/run_golden_song_001.py`; `pha0b`
plays any range of them, live or dry):

| Note | What it plays | Fields | Used by |
|---|---|---|---|
| `wake_both` | Wakes both robots and takes control for the run. | `settle_seconds`, `cube_connect` (`brobot_2`, 2026-10-14: his cube connects over Bluetooth in the background, about 18 s, while pauses and Goverlord's lines go on; the next step that moves a robot waits for it - so a later `cube_after_wake` lights at once) | Bingo test, nap |
| `sleep_both` | Both Brobots fall asleep (phcal's sleep tool). Every `look_seconds`, a Brobot found awake is put back to sleep. | `settle_seconds`, `look_seconds` | nap |
| `stir` | A sleeping Brobot wakes a moment, then drifts back to sleep. Nothing else. | `hold_seconds` (awake) | engine |
| `all_asleep` | Waits until every stir is done and both Brobots are asleep again. | | engine |
| `wake_sleepy` | The soft wake: both sleeping Brobots wake together, at the step's second (with the song's `wake_clip`, e.g. the slower `anim_gotosleep_wakeup_01`, instead of `GoToSleepOff`). | `cube_ready` (`brobot_2`, 2026-10-15: his cube connects in the background `cube_ready_after_seconds` after the wake (8; live 2026-10-15: a connect right after a wake fails) and is held; a later `cube_after_wake` lights it at once) | nap |
| `brobots_ready_together` | Both say the line at the same moment. | `rattle_before`, `pronounce`, `cube_after_wake` (`brobot_2`: his cube lights after the wake, before the line; the cube wakes a sleeping Brobot, bench 2026-10-10) | Bingo test, nap |
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
| `combo` | One story step played as several notes in order (named for them: `brat_nod_kg_success_0203`). A part with `text` says those words, cut from the step's TEXT ("Lol!", "Excuse Me?"); the part without `text` says the rest. One chat bubble for the whole step ("…maturity level? Lol!"), put up with the first words; each part lights its own pieces in it (`piece_offset`). | `combo` (the parts, each a note with its own fields) | Bingo test |
| `sound` | One of the song's own sound files (knobs.json `sounds`, copied to the chat page at Wire-Pod's start), or one of Wire-Pod's shared `sounds/` (2026-10-13: `danger-will-robinson.wav`), played on the page at `volume` (0–1); the song waits its length. | `sound`, `volume` | fart joke |
| `picture` | One picture (or a list) from the chat page's `pictures/` folder, shown in a picture bubble at this step; the magnifier as ever. A missing file warns, never stops the song. | `picture` | |
| `thinking` | The live chat's own thinking please wait, run by Wire-Pod (`/api-sdk/gopod_thinking`, `gopodThinkingProcess`): arm, nod, "Thinking, please wait.", searching (at least `searching_min_seconds`), the get-out. One process for live and scripted runs, locked. | | fart joke |
| `pause` | A silent gap. At 0 seconds it's an edit mark for the reporters' voiceover. | `pause_seconds`, `gap_label` | Bingo test, nap |
| `exit` | Hands both robots back. | | Bingo test, nap |

**Played by the Brobots chat** (`start interview`; `gopod_interview_song.go`):

| Note | What it plays | Fields |
|---|---|---|
| `ask` | A question, read out by its reporter, to a Brobot. | `speaker` (`reporter_1`/`reporter_2`), `to` (`brobot_1`, `brobot_2`, `brobots`) |
| `line` | A spoken line: a Brobot's answer, his brother's line, or a reporter's banter or handoff. | `speaker`, `cue` (his animation), `moves` (`arm`, `nod`, `cube` at their words) |

`{weather}` in an interview line is the live weather; offline, the Brobot says so. **GOLDEN (2026-10-10):** one weather call (`gopodWeatherSay`) for live chat and scripted runs alike: the real weather (Windsor fallback location), else `live_weather.offline_text`. A live question with the word "weather" gets it too; one that names a place ("the weather in toronto ontario, right now?") gets that place's weather (2026-10-13), or Windsor's when the place isn't found. Locked by `TestWeatherPlace`. **By design, not a repeat:** each Brobot
says his own unit and clock first, then the other, from his own Wire-Pod settings
(`temp_is_fahrenheit`, `clock_24_hour`): Brobot 1 metric and 24-hour ("16 degrees Celsius or
62 degrees Fahrenheit ... as of 10:35 or 10:35 AM"), Brobot 2 imperial and 12-hour ("61 degrees
Fahrenheit or 16 degrees Celsius ... as of 10:28 AM or 10:28").

---

## The toolbox (shared golden pieces)

- **Animations** (`goverlord/wire_pod_overlay/chipper/animation_vocab.json`, all verified):
  happy, veryHappy, sad, verySad, angry, frustrated, dartingEyes, confused, thinking, celebrate,
  love, answering, searching, searchingGetout, kgSuccess. Each maps to one stock Vector clip.
  Everyday words pick them in the chat ("party" → celebrate, "mad" → angry); a play on words
  (the Gomads "go mad" pun) plays happy.
- **Sounds** (`goverlord/wire_pod_overlay/chipper/sounds/`): `rattle.wav` (the Bingo rattle,
  1.9 s) and `danger-will-robinson.wav` (Wire-Pod's `playSound`). Both play on the chat page now.
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

---

## Each song's own golden pieces

| Song | Status | Its own pieces |
|---|---|---|
| `01a_brobots_bingo_test` | GOLDEN (2026-10-01) | Host and brat roles; the `OPENING SYNC` opener (wake, warm-up rattle, "Brobots ready" together, arm cue); three ball-draw rounds with the rattle, countdown and big balls; `angry`, `searching` and `kgSuccess` beats; 0-second reporter gaps; the LLM warm-up and "LLM Ready for Bingo"; touch interrupt off. |
| `02_brobots_interview` | WIP (2026-10-12) | The operator's score: six chapters named in the song (`chapter`), ten questions, two reporters; the photos named with their questions (`pictures`); Brobot 1's Awaken bit (arm, nod and cube at their words, and the live weather); the brothers' handoffs; emojis and cues lightly sprinkled; the closer (`note: closer`, Brobot 2's "Press my backpack…") after "Interview run complete". Played by the chat. |
| `101_brobots_nap` | WIP | "Do Baby Robots Dream?": a lyric video. Music plays and a female narrator reads the poem; the 32 captions and their times come from the video's SRT, and the Brobots sleep through it (put back to sleep if found awake) and wake after the last words for "Brobots Ready!". The first timed song: the soundtrack plays on the chat page and the poem shows in one narrator bubble the size of its picture (no header) - each section's words over its own picture, fading in to 50%, out after its last words, the last one staying (2026-10-15). `start baby brobots` (typed, numpad or spoken) or `pha0b nap`. Made in Premiere Pro; the media stays private. |
| `03_fart_joke` | WIP (2026-10-11) | The elevator: Brobot 2 sets it up, the live chat's own thinking please wait (`thinking`, 2026-10-13: one process for live and scripted, locked by `TestThinkingProcessIsOne`), farts (`fart sfx/`, 80%), Brobot 1 goes frustrated, then angry, then leaves. `start fart joke`; "Fart joke started./complete." |
| `01b_brobots_bingo_game` | WIP | The live Chocolate Bingo game: a Go program (`gobingo`), not a step song; Brobot 2's angry reactor. |

**Parked, not lost:**
- The face GIF on a robot's screen (`01a_brobots_bingo_test/face/bingo_number.gif`; its tool was
  archived on 2026-10-09).
- **The sleep family**, for the nap's B-roll. The tool is still here
  (`gopod_probes/tools/direct_sdk_robot_sleep_001.go`: sleep, wake, and play one named trigger
  or clip). phcal's `brobots_sleep_to_wake_direct_sdk` puts both robots to sleep and wakes them.
  The 126 bench aliases (`sleep-beat-*`, `sleep-palm-*`, `sleep-rts-*`, `pet-*`,
  `sleep-segment-*`) were retired 2026-10-08 and are kept outside the repo. Vector's own sleep
  triggers are `GoToSleepGetIn`, `GoToSleepSleeping` and `GoToSleepOff`; the purr lives in the
  petting bliss clips. No snore animation exists in Vector's sources.

---

## The session (Goverlord's stage)

Operator, 2026-10-10: "There is the wire-pod session (wpr 2 until next wpr 2) * Goverlord is
the wpr-session orchestrator". Every small golden detail of the chat is written here, once,
with the code that does it and the test that locks it. If the code and this page disagree,
fix one of them. Values are in `gopod_brobots_settings.json` (each block's `about` says what
it tunes); Go files are in `goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/`, the page in
`webroot/js/gopod_brobots_chat.js`.

### The session and its phases

A session is one Wire-Pod start (`wpr 2`) to the next. Goverlord keeps its one record
(`gopod_session.go`):

| Phase | What it means | What gets through |
|---|---|---|
| **Warming** | From the start until LLM Ready. | Nothing reaches the Brobots: "LLM not ready? Too bad!! Wait!!" |
| **Ready** | Live chat. | Every way in, to the Brobot it names. |
| **Run** | One run is going, and it owns the Brobots. | Only `stop run`, thanks, Record Session's answer and the run's own lines. Anything else gets "Run in progress. Say stop run or thanks." and is dropped. |

### Runs

| Run | Start | Played by | Ends with |
|---|---|---|---|
| Interview | `start interview` | Wire-Pod, inside the session | "Draft run completed." (in the thanks ending) |
| Bingo test | `start bingo test` | the song engine | its last step |
| Baby brobots (the nap) | `start baby brobots` | the song engine | "Brobots Ready! …", then "Run complete." |
| A `pha0b` range, live | the terminal | the song engine | its last step |
| phcal, live | the terminal | phcal | phcal's exit |

- **One at a time.** Every run checks in with Goverlord when it starts and out when it ends
  (`/api-sdk/gopod_run`). A second start gets "A run is going. Say stop run first." A run whose
  process is gone without checking out is cleared.
- **The run owns the Brobots:** the chat's stay-in-place hold ends when a run starts and isn't
  taken while it goes; no live LLM answers.
- **Stopping:** `stop run` stops it at its next safe spot and says "Run stopped."; a thanks
  closes it out, then the thanks ending. A dry run touches nothing and doesn't check in.
- Locked by `TestSessionOneRunAtATime`, `TestSessionClearsAGoneRun`.

### GOPOD commands: Wire-Pod intents (2026-10-17)

Operator: "make a GOPOD commands list for wire-pod intents * start with 'pop a wheelie' intent
command". A typed (or numpad) line holding one of these words goes to the chosen Brobot as
Vector's own native intent - he does what he'd do if he heard it - not to the LLM. To All: both.
First match wins. The list is `native_cube.intents` in the settings (`gopod_native_cube.go`);
a new command is a new entry there (its intent name from Wire-Pod's `intent-data/en-US.json`),
then `wpr 2`. Spoken lines are matched by Wire-Pod itself.

| Command (any of the words) | Intent | Vector |
|---|---|---|
| pop a wheelie / do a wheelie | `intent_play_popawheelie` | pops a wheelie on his cube (the cube must be near; tried 2026-10-17 06:28 typed to Brobot 2: sent, done in 0.4 s) |
| pick up your cube / lift your cube | `intent_play_pickupcube` | picks up his cube |
| fetch your cube / bring me your cube | `intent_imperative_fetchcube` | brings his cube |
| roll your cube / roll the cube | `intent_play_rollcube` | rolls his cube |
| play keepaway / keep away | `intent_play_keepaway` | plays keep-away with the cube |
| find your cube / where is your cube | `intent_imperative_findcube` | looks for his cube |

### Ways in (operator → Brobots)

All lines meet in one door, `gopodOperatorLine` (`gopod_typed_input.go`), checked in this order:
Record Session's answer → `start interview` → `stop run` → the warm-up gate → the run gate →
thanks → GOPOD Yourself → a voice command's words (`start bingo test`, `start baby brobots`) → a
live question.

| Lane | What it is |
|---|---|
| **Voice** | "Hey Vector" or the backpack button. A custom voice command (`customIntents.json`: START_BINGO_TEST, START_INTERVIEW, START_BABY_BROBOTS, STOP_RUN, GOPOD_YOURSELF) runs its script, which hands its words to the door; anything else is a live question (`kgsim.go`, the same gates). |
| **Typed** | The typing line, aimed by its radio: 1, 2 or All (All at every page load). Enter sends; Esc leaves the box. A typed line is read out first by the reporter: to Brobot 1 by Brobot 3, to Brobot 2 by Brobot 4, to both by Brobot 3. |
| **Numpad** | Push-to-talk while "Numpad?" is ticked and the chat is open: hold a numpad key (its persona from the numpad map), speak, let go. The numpad's own keys are always push-to-talk, even in the typing box; every other key types. Three taps of 0 (or NumLock) leave numpad mode. Every numpad key press plays the grunge sting (`numpad_ptt.key_sound`, 2026-10-16). Holding 0 (`all_hold_seconds`, 0.5 s) is push-to-talk to All - Brobot 1 answers, then Brobot 2, like the keyboard's All button (2026-10-16). Its status messages (holding, recording, transcribing...) stay in the chat, one line each, never spoken (2026-10-14). The "push-to-talk off" line is said by Goverlord. |
| **Native** | Wire-Pod's own "Hey Vector" answers (the time, the weather, volume…). The typing line's **Native?** box turns them on or off for the session (`native.on_at_start` at Wire-Pod's start); during a run they're always off. Off, a native phrase gets Wire-Pod's quiet ending. Your voice commands and chat questions aren't native and always get through. Locked by `TestNativeAnswersOnOffAndDuringARun`. |
| **Page** | Clicks only: Native?, Numpad?, Record?, Brobots Chat, Clear Chat. The page also plays every voice and tells Wire-Pod what it has heard. It never sends words. |

### Wire-Pod start

1. A new chat feed with a new boot stamp: the page clears and reloads once (`GopodChatFeedStart`).
2. The sounds and every song's own soundtrack are copied to the page's voice folder.
3. The warm-up bubble: "Goverlord: GOPOD Brain", a row of LEDs red to green and the hourglass,
   "LLM Warming Up" (`gopod_llm_warmup.go`).
4. The reporters' opening plays on the page, once the page is open and clicked.
5. The bubble turns to "LLM Ready. Warm-up complete.", all LEDs lit, and stays; Goverlord says
   both. The warm-up gate opens. Locked by `TestWarmupGate`, `TestWarmupReadyLineIsOneBubble`.
6. **GOLDEN opener template, stamped 2026-10-16** (operator, after a correct `wpr 2` run:
   "snapshot/lock/stamp this starting opener LLM as the golden reference template * for song
   runs, apply this template as a base * for each song, apply/add updates/changes
   accordingly"). Three messages in the one System llm bubble, one line each - the operator's
   exact lines (2026-10-16 later, "EVERYTHING IN THE QUOTES!! DO NOT MUTATE"):
   - message 1, at every start (it is the check): "Voice files generation check." - no sting (operator 2026-10-16: "remove 1st \"(music)\" & remove its audio play"); it waits for its own voice file before it is said ("voice check file wasn't heard": it was said before the voice worker was ready). The end sting stays.
   - message 2: "Voice files generated. Language Model Loading"
   - message 3: "Language Model Loaded, Master CrushN8r. Your move. (music)"
   "(music)" is shown in the bubble, for deaf viewers, where the sting plays; the butler doesn't
   say it. Each bubble is built and its voice loaded off screen first; only when both are ready is
   it placed in the chat, and its voice waits until it shows. A song run is the same base: message
   2, then "Language Model Loaded, Master CrushN8r." / the song's own `llm_ready_text` / "Your move.
   (music)". (`llm.voices_text`, `warmup_text`, `wait_text` (empty), `ready_text`,
   `warmup_done_text` (empty), `run_ready_text`, `run_ready_end`; `gopodRunReadyLine`, and the same
   in `run_golden_song_001.py`.) The butler says "please" as "pleeze" (`system_say_words`).
7. **The opener is a song (2026-10-16**, operator: "copy the song folder template, and make it a
   master opener LLM warm template * from the LLM warm template, make the opener LLM Warming
   song"). The template is `goverlord/runtime/songs/_new_song/_llm_warm/` (2026-10-16, operator: "would make sense to put songs/_llm_warm/ folder in _new_song folder"); the opener LLM Warming song is
   `00_llm_warming/` (`song_id` `brobots_llm_warming`, settings `llm.opener_song_id`). Wire-Pod's
   start plays its steps in order: `goverlord_warm_0101` (message 1, waits for the voices),
   `goverlord_warm_0102` (message 2), `goverlord_warm_wait_0103` (the model loaded),
   `reporter_lines_0104` (the reporters' opening), `goverlord_warm_ready_0105` (message 3),
   `goverlord_sting_0106` (the end sting). Its words are its `story.md`; edit one and `wpr 2`
   (the voice file is made for the new words). No song, or one Wire-Pod can't read: the built-in
   opener from the settings (`voices_text`, `warmup_text`, `ready_text`), logged
   `GOPOD_OPENER_SONG`. A folder starting with `_` is a template, never a song. Proved by
   `TestOpenerSongPlaysLikeBuiltIn` (the song plays the built-in opener's chat lines, voices and
   stings, in the same order) and `TestLLMWarmTemplate`.
8. **Goverlord awaits your return (2026-10-17**, operator: "instead, system message, Goverlord
   awaits your return... + hourglass, and waits, nothing plays * upon page focus, opening llm
   should, go through each message ... in correct proper order"; "both the terminal run & web page
   run must be the same run"). The chat page is **away** when it says it's hidden, or hasn't been
   heard from for `session.page_present_seconds` (20; it reports every 15 s) - closed, or never
   opened. While it's away: the opener waits before each of its steps, the System llm bubble showing
   `session.awaits_text` "Goverlord awaits your return..." with its hourglass, nothing plays; a
   song run waits before it starts and before each step, and the terminal says so
   (`GOLDEN_AWAIT ... the chat page is away`, then `... is back after Ns`). `gopod_page_presence.go`,
   `wait_for_page` in `run_golden_song_001.py`. Locked by `TestPagePresence`.
   **GOLDEN, locked - the System llm bubble's sequential run check (2026-10-13), the one template
   for every run: the opening after `wpr 2`, the interview, a replay, every song with its own
   reporter lines.** Each message is a line in the chat feed, so the page shows every step in
   turn even when it's opened late (each held `llm.warmup_check_seconds`, or for its spoken
   line). Locked by `TestWarmupSequenceGolden` (Wire-Pod start and runs) and
   `test_golden_song_step_names_001.py` (songs).
   **Input pauses for it:** from a run's first warm-up message until its LLM Ready has been
   said, every typed or numpad line ("stop run" too) waits in a queue, then plays in the
   order it came (`GOPOD_WARMUP_QUEUED` in the log). Locked by `TestWarmupQueuesInput`. one System llm bubble,
   shown first, LEDs filling, and kept at the bottom under every new bubble until LLM
   Ready. Its message in turn: "Voices are being generated. Please wait." while the voices
   are made; "LLM Warming Up. Please wait." always, shown and said (a cold model waited out), and it stays while the opening plays; then "LLM Ready. Warm-up complete.
   …", said. Texts: `llm.voices_text`, `llm.warmup_text`, `llm.wait_text`.
   **Stings (2026-10-14):** djent plays once at the very start, before the first message's
   voice, and once at the very end, after the LLM Ready message; none in between. Every
   warm-up run alike. Locked by `TestWarmupStings`.
   **Order (2026-10-14):** pre-recorded (Wire-Pod's start only, while voices are made),
   warming up, in-between lines, LLM Ready. "LLM Warming Up. Please wait." stays until fully
   said; only then does the bubble turn to LLM Ready and say it - a song with no lines in
   between too. Locked by `TestSongWarmupReadyAfterWarming`. The run goes on only once the
   page has played the LLM Ready message and the closing djent (its heard check), never on a
   guess of their length.
   **A run's LLM Ready message (2026-10-14):** part 1 is standard, `llm.run_ready_text` ("LLM
   Ready. Warm-up complete."), kept once in the settings; part 2 is the run's own - a song's
   `llm_ready_text`, the interview's `interview_script.llm_ready_text` ("We here, are Fart Joke
   Ready"). One bubble, one spoken line. Wire-Pod's own start keeps `ready_text` +
   `warmup_done_text`.
   **Every wait on a line (2026-10-14):** Wire-Pod goes on the moment the page says it has
   played it (the heard check); with no answer, after the queued lines' own lengths plus
   `reporter_voices.heard_margin_seconds` (3). No page-lag padding is added.

### Replay a session

`replay <name>` (the session folder's name: `095819_session`, `095819` or
`2026-10-12/095819_session`) plays that session's score (`story.md` + `knobs.json`, the words as
edited in `story.md`) back through the live chat (`gopod_session_replay.go`, 2026-10-12):
Goverlord's "[name] started.", then in order the reporters' lines, each operator line with its
Brobot's logged answer said through the chat-reply path (cue, moves, thinking please wait,
bubble, highlight), handoffs, songs started again by their own scores, the interview's section
from the score; then "[name] complete." A thanks in the score plays the ending live. "stop run"
stops it. A replay is never recorded: Record? stops (kept) and can't start during it. Since
2026-10-12 every answer in a score keeps its cue (`say`) and a scripted answer its moves
(`moves`). Locked by `TestSessionReplayPlan`.

### Every run's start, pause and stop (2026-10-12)

- **Start, the opener's golden template:** every song and the interview start like Wire-Pod
  does: the LLM warm-up bubble, any reporter lines (`warmup_lines`, the interview's
  `start_lines`), then the ready line said by Goverlord (`llm_ready_text`; the interview's in
  `interview_script.llm_ready_text`). A song opts out only with `"llm_warmup": false`.
- **The warm-up rule (2026-10-13), every run (Wire-Pod start, a kept-session replay, songs,
  the interview):** the System llm bubble first, kept at the bottom; a cold LLM warms
  ("LLM Warming Up. Please wait."), a warm one skips that; then whatever plays in between
  (reporter lines), the message staying "LLM Warming Up. Please wait."; then the bubble shows LLM Ready and Goverlord says it.
  The ready bubble takes its turn in the chat's order (the feed's `warmup_ready` line), and
  a song goes on only once the ready line is said. A replay's first reporter lines play
  inside its warm-up.
- **Pause:** while the chat page is hidden, a song waits before its next step, the interview
  before its next question, a replay before its next line; Goverlord says "[run] paused." and
  "[run] continues." A timed song (the nap) plays on.
- **Stop:** "stop run" stops the run going (a song, the interview), also inside a replay;
  "stop session" stops the replay itself. A replay's lines name it "Session 095819".
- **Bingo ball:** the called ball shows under its "Bingo ball captured. B-1" bubble when "B-1"
  lights (its room kept while hidden).

### Record Session (the "Record?" box)

1. Ticked: recording starts at the next safe spot (never in the middle of a reply), in
   `<time>_brobots_session/` inside the open session's folder
   (`~/gopod_tts/sessions/<date>/<time>_session/`): the mic (48 kHz), the chat and
   Wire-Pod's log. Every session file lives in the session's folder (2026-10-11): numpad
   presses as `numpad/<HHMMSS>_<key>_<n>.wav` (48 kHz), numpad errors in its `wirepod.log`. The label and the timer turn red; the timer restarts from the recording's start.
2. While recording: every chat line and answer is written; each answer is a take for the take sheet.
3. Unticked: it stops and keeps the session (no question). The timer goes back to 0.
4. Thanks, or `max_minutes`: it stops, and a Brobot asks keep or delete. The box stays ticked
   until the answer; unticking then counts as keep; no answer in `answer_wait_minutes` deletes.
5. Keep: the takes go on the take sheet and the folder is zipped. Delete: it's removed. A
   Brobot says which. Locked by the `TestGopodRecordSession…` tests.

### The thanks ending

Once per ending, in this order (`gopodThanksEnding`):
1. a run going is closed out first;
2. the score (logged);
3. the reporters' closing;
4. the brothers' sign-off lines;
5. Record Session's keep-or-delete question, and the wait for the answer;
6. "Draft run completed." if a draft played;
7. "You're welcome.", always.

Locked by `TestThanksClosesOutTheDraft`, `TestReporterClosingOncePerEnding`.

### A live reply

1. The Brobot named in the first two words answers; with no name, the one who heard it first,
   and his brother stays quiet (a glance).
2. Thinking: the arm lift, then the nod, then "Thinking, please wait." (`chat_moves`).
3. The searching animation until the answer is ready.
4. The reply: each animation sent, the speech 0.9 s later, the next animation only once the
   robot reports the last one done. Its bubble shows as the speech starts, each piece lit as
   it's said.
5. The battery is logged; the stay-in-place hold comes back.

Locked by the `TestGopodAddressed…`, `TestGopodShouldAnswer…`, `TestHighlightPiecesGolden` and
`TestStuckReplyGivesUp` tests.

### The Chat Bubbles Window: the machine (rule book, 2026-10-18)

Operator: "start from wide overview, what the 'Chat Bubbles Window' engine looks like (the
final look) and what its supposed to do (terminal level) * go deep through gears, plumbing,
wiring". The whole window on one page; the rules after it are the detail. **Draft for the
operator's signature** - once signed, this is the locked rule set.

**What it is.** A Brobots layer on native Wire-Pod (the overlay, "chat bubbles lite"): Wire-Pod's
own Log page gets a scrolling window of bubbles that shows the show as it happens - every line
said, by whom, to whom, with its pictures, timers and eggs. The terminal (journal) carries the
same show, line for line (`GOPOD_FEED <id>`, the page answering `GOPOD_PAGE played <id>`).

**From the Log click to the bottom of the window:**
1. **Log click** - the page counts as here (seen, clicked, chat open). Before that, Wire-Pod waits
   ("Goverlord awaits your return") and nothing plays into an empty room.
2. **The feed** - Wire-Pod numbers every line it adds to `gopod_chat_feed.json`; the page plays
   them one at a time, in order, and says "played" for each.
3. **A voice is made before its line shows** - if not made yet, a *slot* opens first:
   "Voice file generating." + hourglass (`generating_marker` "on"); the line then fills that same
   bubble, in place, with its own header and words (`fillSlot`). The Operator's own bubble does
   it in itself ("op"); the opener's warm-up bubble does it for its message lines.
4. **Every bubble's life is the same:** built whole while hidden (pictures, QR codes, weather face,
   flags in place) → the chat scrolls once to make room → it shows when the bubble above is done
   (the baton) → its words are revealed as they're said → said (85%).
5. **Who reveals the words:** the Brobots' replies - Wire-Pod names the piece being said
   (`gopod_speaking_now.json`), the page reveals that piece word by word. Reporters' and
   Goverlord's lines - the page plays their voice itself, so it reveals the words across that
   voice's own length (`voiceLitLine` / `voiceLitStart`). SRT songs - the soundtrack's times.
   Word, then its punctuation, then emojis; numbers, clock and year as said.
6. **Small processes plug into that one life** and work for every bubble that needs them: the QR
   codes (QR1 top-right, jumps to bottom-right when scrolled past; QR2 after the call to action),
   the egg marquee, the magnifier on every picture, the weather face, flags as pictures, the
   search-results box, the hourglass.
7. **Scroll to the bottom** - the newest bubble ends just above the typed-input line; the chat
   follows the highlight; nothing scrolls again until the next bubble.

**Shared helpers - one each, used everywhere** (2026-10-18, operator: "whatever other processes
are the same but scattered, share from one shared slot"):
- `slotWaiting` / `slotReady` - every "Voice file generating." + hourglass: the grey slot, the
  Operator's own bubble, a reporter's or Goverlord's line filling the slot (`fillSlot`).
- `voiceLitLine` / `voiceLitStart` - the word reveal for every voice the page plays.
- `magnify` - the magnifier on every picture, QR code (QR1, QR2) and weather face.
- `whenShown` - anything that waits for its bubble to show: the weather sound, the search
  results' typing, (and the reveal's 15 s fallback).
- `cornerBox` - QR codes and the weather face in a bubble's top-right corner.

**Step 3, the engine conducts** (2026-10-18, operator: "Finish Step 3"): Wire-Pod times itself
by the page instead of guessing. (1) A Brobot speaks once his bubble is on the page - up to
`page_sync.reply_wait_seconds` (10) - so he never talks over a line the page is still playing
(`gopodWaitReplyShown`; `GOPOD_SYNC` lines when he waited or the page was behind). (2) The page
tells how long it held each line before showing it; over `page_sync.held_log_seconds` (3) it's in
the journal (`GOPOD_PAGE held <id>`). With "played" and "heard" already there, every page wait
is now seen by the engine. The page still keeps its drawing order (one bubble after the one
above, the countdown, the SRT narrator's soundtrack clock) - the engine now waits on those
instead of running ahead of them.

### Chat bubbles

**The operator's chat rules, in one place (2026-10-13; review: `gopod_notes/RUN_INTEGRITY_REVIEW_001.md`):**

1. A bubble is built fully hidden, away from the chat, so it never moves the scroll.
2. It's ready once its content has loaded and rendered (pictures, its voice ready to play).
3. The full bubble gives its final height; that height means it's ready to show.
4. Only then, at its turn in the queue, does the chat scroll and show it.
5. A short bubble stops bottom to bottom (its bottom at the window's bottom); a tall one stops
   top to top. Otherwise the chat stays at the bottom.
6. Every bubble runs and finishes 100% before the next one shows.
7. Headers: system bubbles "[grey icon] Goverlord: GOPOD Brain [grey icon]"; operator bubbles
   "[operator icon] [input] to Brobot N [Brobot N eye-colour icon]" (to Brobots: both icons);
   reporter bubbles "[white icon] Reporter 1: Brobot 3 [white icon]"; Brobot bubbles
   "[eye-colour icon] Brobot N:" (one icon).
8. Live operator lines (keyboard, numpad, and the robot's mic held during a warm-up, "Mic to
   Brobot N"): an Operator bubble read out in the Operator voice (`reporter_voices.operator_voice`,
   `af_sky`). A playing draft's question keeps its reporter's voice.
9. No time limit takes the baton early: the next bubble waits for the one being said (only a
   highlight stuck on for 2 minutes lets go).
10. **Narrator bubbles** are the nap's captions: the song's subtitle (`.srt`) lines, one bubble
    per poem section, each phrase lit on the soundtrack's clock - nobody's voice. Like every
    bubble, each is built out of the chat's flow and joins it when it's placed.
11. **Background loop:** `background_loop` (`BrobotsGroove-Loop-1.wav`, 6 s, 5%; a 🔊/🔇 speaker
    toggle after the text field turns it on or off, kept in that browser; every run quiets
    it at its start and, as its very last line, sets it back to the toggle) plays on the
    page over and over with no gap, from its top again at every session clock reset to 00:00:00.
12. **System stings (2026-10-13):** before Goverlord speaks a system bubble, the page plays a sting
    at 10%, then the speech: `grunge.wav` for a run's started, stopped, complete, paused or
    continues line (and the stop-run answer when nothing runs); `dJent3b.wav` for every other
    system line. One sting per bubble (`reporter_voices.system_stings`).
13. **Voices message:** "A pre-recorded message. Voices being generated. Please wait." is shown
    while the voices are made, and said - at once when its voice file was made at an earlier
    start, else as soon as it's made.
14. **Voice files: made once, play many.** At each Wire-Pod start every scripted line (system,
    reporters, the interview, each song's warm-up, the run commands) is made once, in the one
    voice that says it - no more copies in every voice. A scripted voice file no line uses any
    more (an edit, an old voice or speed) is removed then; an Operator voice file is kept until
    it hasn't been said for `reporter_voices.operator_keep_days` (30). Logged as
    `GOPOD_REPORTER_VOICES cleaned up N old voice files`.
15. **Heard, not read ahead (2026-10-15, global):** a highlighted bubble (a Brobot's reply, a
    scripted song line, the narrator) is built whole as before, but its words are invisible
    (0%) until heard: the piece being said is solid white (100%), a piece already said 85%.
    The pieces before a lit one count as heard; when the next bubble takes the baton, or 10 s
    after the robot stops with nothing more lit, the rest shows. A restored chat shows all.
    `.spoken` / `.speaking` in `gopod_chat_bubbles.css`, `lightSpan` in `gopod_brobots_chat.js`.
16. **Numbers one at a time (2026-10-15):** in the piece being said, a run of single digits ("the
    1 4 2 8 5 7 repeats") and a military time (its spoken parts: 0 9 0 4, 18 14) light one
    number at a time - a best guess, not tied to the audio (`digit_highlight`).
17. **Follow the highlight (2026-10-15):** the chat holds still until the piece being said is a
    quarter of the window below centre (`highlight_scroll.dead_zone`), then one smooth move puts
    it back at the centre - never past the chat's end (the bottom/bottom rule), never back up on
    its own. Once the bottom is reached, the highlight carries on down the bubble.
18. **Per pause, as it was (2026-10-15):** a short try at whole-sentence pieces and thought lines
    was switched back off the same day (operator: "I had per pause!! ... PUT IT BACK!!") -
    `join_up_to_words` 0, `thought_lines` false. Highlighting is per pause.



- **Placement (2026-10-12, 13):** a new bubble is dressed first, somewhere else: built out of
  the chat's flow, hidden, taking no room and opening no scroll space, until its pictures
  have loaded and its voice is loaded and ready to play. Only then does it join the chat. It comes in from the bottom: one that fits the
  window stops with its bottom just above the typing line; one taller than the window stops
  with its top at the top. Its voice plays once it shows. A picture added to a question
  bubble places it again, whole. The LLM warm-up bubble is the exception. One scroller, 1 s
  ease, never up on its own.
- **Golden rules 10, 11 and 14:** a bubble is built whole while invisible, the chat scrolls once,
  then it shows, with its voice when it has one; the next bubble takes the baton only when
  the one above has finished.
- **Kinds:** Brobot bubbles (their eye colours, one icon); reporter bubbles ("Reporter 1: Brobot 3", white-eyed icons);
  Goverlord's grey system bubbles, every one headed "Goverlord: GOPOD Brain" (2026-10-13;
  "Reporter 0" is only a decoration label) (the warm-up, countdowns, chapters, Goverlord's lines, numpad
  status); the narrator's (one per poem section, each phrase lit on the soundtrack's clock).
- **Pictures:** every picture has the magnifier: hover or tap to see it full width.
- **Live pop-ups (2026-10-13):** in the live chat, a topic's pictures (Chocolate Bingo on "bingo",
  Gangster Granny on "granny"…) show up to `topic_pictures.max_per_session` (5) times per Wire-Pod start in a Goverlord system bubble with
  "Congratulations! Easter Egg found." (the egg text names the egg, below) over them (`topic_pictures.popup_text`),
  said by Goverlord as it shows, zoom kept; it comes after the question has been read out.
  Marquee bulbs ring it (always animated, the PC's reduce-motion setting too), placed along its rounded border
  and cycling Chase, Groups, Pairs and Twinkle every 5 s; they're laid out again whenever the
  bubble changes size, their count following the border's length. A scripted run shows its
  pictures as before.
  **Easter eggs (2026-10-14):** pop-ups are found eggs in the live chat only - never while a
  draft plays or any run is going (a song, the interview, a replay; locked by
  `TestPopupsOnlyLive`). Each egg shows up to `max_per_session` times per Wire-Pod start.
  - **Weather:** a weather question finds the weather pop-up - Vector's weather face for the
    real weather (cyan, full size, the magnifier) and its sound at 10%, under Goverlord's
    congratulations.
  - **GOLDEN pop-up rule (2026-10-15):** every easter egg a question finds is its own bubble -
    the Brobot's bubble (his colours) ringed by the marquee, holding Goverlord's grey block
    ("Goverlord: GOPOD Brain", "Congratulations! Easter Egg found." (the egg text names the egg, below), the
    prize); no header above it. Two eggs (granny + weather) are two egg bubbles. His reply comes
    after, in his own marquee bubble with his "Brobot N:" header. Order: the prize and the spoken
    congratulations, then "Thinking, please wait.", then the reply.
  - **GOLDEN, the pinnacle egg process (2026-10-16):** every egg a line finds (a topic's pictures,
    the weather, Explain the Math) shows in the order its word comes in the line, each its own
    egg bubble with "Congratulations! Easter Egg '{name}' found." (a colon instead of the period
    before a card, its card lines under it) and, under a prize that zooms (a
    picture, a photo, a weather face), "MouseOver to Zoom View" on its own line (2026-10-16 later;
    `egg_zoom_text`; the butler says "mouse over"). ({name}: the egg's name, else
    the word that found it; `topic_pictures.egg_text`, 2026-10-16), said by Goverlord before the next; then
    the reply. An Explain the Math line shows its other eggs too (Brobot 1's). An egg's word is
    never read as a weather place ("weather for bingo").
  - **Egg names, cards and the contest (2026-10-16):** an egg is named by its topic
    (`topic_pictures.egg_names`: Birthday - the Hot Wheels box, Fractions - the Doc & Pip
    workbook, Roast Battle, Mobile Gear - the GOPOD stack picture, found by the rig's words like
    mobile and gear; since 2026-10-16 "ai" no longer pops it), else by the word that found it. An egg with no picture yet shows a
    card (`cards`: the Roast Battle's "Party Brobots vs any and all takers"). "AI Wordplay Explain
    the Math" is the livestream show's egg - "[contest submission text]. Explain the Math" - with
    "🏆 Win to get Doc's approved 'Almost Approved' top prizes!" as its prize (2026-10-16) (`math.egg_name`, `popup_title`). A line
    with fraction, fractions, decimal or decimals is worked out too (`math.fraction_words`), and
    Brobot 2's follow-up says a decimal is a kind of fraction (`decimal_follow_up_line`).
  - **Egg picture points (2026-10-16):** every egg picture has a few points about it in
    `webroot/gopod_picture_descriptions.json` (written by hand from each picture; the vision
    model's first try was mostly wrong). They're made only for a new or changed picture, at
    Wire-Pod's start, never during a run; the Brobots get them when the egg pops.
  - **The picture, described (2026-10-15):** a pop-up picture in `topic_pictures.describe`
    (granny) is looked at by the photo egg's vision model; the Brobots answering get "The pop-up
    picture shows: ..." before the question (`describe_note`, waits at most
    `describe_wait_seconds`, 30; once per Wire-Pod start per picture). Optional pictures are looked
    at only when asked: a question with "chocolate" has the Chocolate Bingo kit and card looked at
    for the reply, pop-up or not (`describe_on_ask`). While pictures not looked at yet are being
    looked at, Goverlord says "Looking at N picture(s). Please wait." (`describe_wait_text`). Egg
    pictures are looked at ahead, quietly, 30 s after Wire-Pod starts (2026-10-16), so a pop-up's
    reply doesn't wait.
  - **"All" = all Brobots in sequence, everywhere, pop-ups too (2026-10-15):** Brobot 1 answers,
    then Brobot 2. When the question found a pop-up, every Brobot's answer wears the marquee -
    each in his own marquee bubble with his "Brobot N:" header - until the next operator line. The All sequence (Wire-Pod) and the pop-up
    dressing (the page) are separate processes that line up. "Explain the Math" follows the same
    rule.
  - **Weather Kit link (2026-10-15):** Brobot 2's follow-up to a weather question ends his own
    words with `live_weather.follow_up_kit_line`, before his "Hey, that reminds me...": the bubble
    shows "GOPOD Weather Kit" as a link (opens `gopod_weather_kit.html` in a new tab), he says the
    words. Any "[words](/page)" in a bubble is a link like it.
  - **Follow up, no repeats (2026-10-15):** Brobot 1 answers, Brobot 2 follows up - his prompt gets
    `addressing.follow_up_note` before the question (`math.follow_up_note` for Explain the Math:
    he can set his brother straight). He still says his own weather line - each robot's native
    units (Brobot 1 metric/24-hour, Brobot 2 imperial/12-hour) - and his reminder line.
  - **Photo (2026-10-14, a starter):** "what's in front of you?" / "what do you see" / "take a
    photo" - the Brobot takes the photo the golden way (Wire-Pod's own getImage: mirror mode,
    "3, 2, 1", the photo, mirror mode off, the shutter animation), the photo is the prize,
    then he thinks while the vision model looks (`photo.vision_model`, moondream since 2026-10-16
    through Ollama), and says "I see ..." - one sentence. "Tell me what you saw"
    answers from what he saw last ("I saw ..."); the photo is never sent again. No photo:
    "I couldn't take a photo." Photos are kept in `webroot/photos/` (private, gitignored).
    `photo` in the settings, `gopod_photo.go`, `TestPhotoKind`/`TestPhotoRecall`
    (`GOPOD_PHOTO_LIVE=1` also asks the vision model).
  - **Photo look timer (2026-10-16, replaced the same day by the 2 rounds below - no bubble):** while the vision model (moondream) looks at the live
    snapshot, Goverlord's warm-up bubble shows "Looking at 1 picture. Please wait." and fills (said
    once); once the look is done it turns ready with his notes ("Notes: what it saw";
    `describe_done_text`). Then the Brobot does his normal "Thinking, please wait." and searching
    moves, and answers from the notes. Static egg pictures never use it - their points are kept
    (see above). Since 2026-10-16 a live snapshot has no Goverlord bubble (below).
  - **Photo in 2 rounds (2026-10-16):** the Brobot says "Thinking, please wait.", then takes the
    photo (mirror mode, "3, 2, 1", shutter), then the egg "Congratulations! Easter Egg 'Get Photo'
    found." (`photo.egg_name`), then the Edge TPU names the objects at once (about 0.04 s; its
    worker loads the model once at Wire-Pod start, `photo.quick_look`), then the searching moves,
    then statement 1: "I see a person, a chair and a keyboard." (`quick_text`; "Hmm, I need a
    closer look." when it names nothing). moondream looks meanwhile, in the background - its
    notes are only in the log (no bubble). After statement 1, while moondream is still looking:
    "Gathering details. Please wait!" (`gathering_text`), then one stall line (`stall_lines`,
    draft lines). Then statement 2 in his own bubble: "I see ..." from what moondream saw (its
    own "The image shows" taken off, `detail_strip_prefixes`) plus statement 1's objects it
    didn't name ("..., and a tv."); "That's all I can make out." when moondream saw nothing. The
    photo runs inside the one thinking process (after its line, before its searching moves).
    Statement 2 is part of his answer: in All, Brobot 2 starts after it; a touch stop skips it.
    `gopod_photo.go`, `gopod_photo_quick_look.go`, the worker
    `gopod_probes/tools/gopod_edge_tpu_quick_look_001.py`, `TestPhotoRounds`.
  - **Thoughts as paragraphs (2026-10-16):** each thought in a reply is its own paragraph, a
    blank line between them (`live_split.thought_lines` on): the code's own lines (the weather,
    the value point, "Hey, that reminds me...") and gemma's thoughts, which end with an emoji.
    Highlighting stays per pause (comma pieces, `join_up_to_words` 0); only the look changes.
  - **Links are said as their words (2026-10-16):** "[GOPOD Weather Kit](/gopod_weather_kit.html)"
    is said "GOPOD Weather Kit", never its address (`gopod_speak_links.go`).
  - **Weather place named first (2026-10-16):** "I want to visit my granny in Honolulu Hawaii. How
    is the weather there today?" gets Honolulu's weather - a place in capitals after in/to/at/from
    anywhere in the line, when none comes after "weather"; Windsor only when no place is named.
  - **Typos in capitals (2026-10-16):** a typed word like "iS" is read as "is" ("iPhone" stays).
  - **Military time (2026-10-16):** "1027 Hours Military Time" lights 10, then 27, then "Hours
    Military Time" - the label is its own step after the numbers.
  - **GOLDEN, word at a time (2026-10-16):** the piece being said starts at 0% solid white, then
    each word turns 100% solid white as it's said and 85% once said. Every word is its own step
    (`word_reveal.seconds_per_word`, 0.38); single digits step at `digit_highlight.seconds_per_number`;
    a math formula is one step. A silent lead-in of two words' time comes first, then the first word
    lights (`word_reveal.start_delay_words`, 2026-10-16), so the reveal doesn't start before the Brobot speaks. A clock time and a military time step as said: "1:16 PM or 1316
    Hours Military Time." reveals 1 / :16 / P / M / or / 13 / 16 / Hours / Military / Time. A best
    guess, not tied to the audio; when the piece ends the rest shows. The pieces themselves (per
    pause) don't change. The chat page, `digitUnits`/`digitSequence`.
  - **Clean GOPOD (2026-10-16):** outside a roast battle, a Brobot's country flag is kept only
    when its country is named in the question or the reply; a wrong flag becomes the named
    country's flag when exactly one is named, else it goes ("RULE: use correct & proper country
    flag emojis"). A sentence with a political word (`clean_politics.words`) goes; political
    stuff is only allowed in the roast battle (`clean_politics.allowed_topics`).
    `gopod_clean_politics.go`, `TestCleanPolitics`.
  - **Said right (2026-10-16):** bruh "brugh", moron "more-on", chocolate-y "chocolate-ee", the Brobots' party lingo (their
    pronunciation list). CrushN8r with 8r as a long British "ate-ah" ("crushin-ate-ah") is the
    butler's lingo only (Goverlord, `system_say_words`); the robots keep "Crush-Ehnaydir", the
    reporters unchanged. Three lingos: Brobots (the default for a new pronunciation), reporters
    (`reporter_voices.say_words`), butler.
  - **Weather Kit line (2026-10-16):** its own paragraph in Brobot 2's bubble.
  - **Place typed small (2026-10-16):** "my granny lives in jerusalem what's the weather there
    today" gets Jerusalem's weather.
  - **All, no repeats (2026-10-16):** Brobot 2's follow-up drops a sentence of his own words when
    every word that carries meaning in it was already in Brobot 1's reply; one new word keeps it ("a
    man" vs "a man with a hat" is fine). Each keeps his own thoughts as paragraphs; the code's lines
    (his weather line, "Hey, that reminds me...") stay. `gopod_brother_repeat.go`,
    `TestDropBrotherRepeat`.
  - **Comb-through fixes (2026-10-16):** the chat keeps following the highlight while a next
    bubble waits (a staged bubble is never "the end of the chat"); the highlight stays on the bubble
    the Brobot is saying (past its first piece) and never lands on an egg; the word reveal only runs
    in live, scripted and narrator bubbles (no flicker in plain ones); the marquee ends after a voice
    (backpack) question's exchange and after a reload; a voice waits while its bubble - or any
    bubble before it - is still waiting to show, as long as the baton can be held.
  - **Thanks at the next safe spot (2026-10-16):** "thanks guys" anywhere in a line still ends the
    session, but the ending waits until no Brobot is answering (`sign_off.safe_spot_seconds`, then
    while a reply is going) - to All, Brobot 2's sign-off comes first.
  - **SRT-timed song (2026-10-16):** the Baby Brobots nap is timed to an SRT file - a unique starting
    point for GOPOD song writing, so the same mechanics follow the SRT's clock: a narrator caption's
    word reveal starts at once and is scaled to fill its SRT segment exactly, first word to last.
  - **Photo statement 2 (2026-10-16 later):** never repeats statement 1's objects; a look that
    names nothing new says "That's all I can make out." (operator: "do not repeat without more
    detail * build 2 more details from 1, no repeats").
  - **Repeating decimals (2026-10-16):** "convert 0.1234 repeating into a fraction" - all of 1234
    repeats (1234/9999). Only the end repeating: "convert 0.1234 with the 4 repeating into a
    fraction" (also "only the 4 repeats", "123, 4 repeating", or "0.12344 repeating"); "with the
    34 repeating", "with the 234 repeating" the same way.
  - **Math to All, once (2026-10-16):** a fraction-word line typed to All runs the math flow
    only (Brobot 1, then Brobot 2's follow-up) - before, the All sequence ran too and each Brobot
    answered twice.
  - **No question read back (2026-10-16):** a reply sentence made only of the question's words
    (three or more) is dropped (`gopod_question_echo.go`).
  - **Links keep their brackets (2026-10-16):** a "[words](/page)" link is never read as a
    stage direction, so the Weather Kit link shows as a link.
  - **Place named once (2026-10-16):** "Jerusalem Jerusalem District" is said "Jerusalem".
  - **Granny (2026-10-16):** "gangsta" finds her egg too.
  - **Chocolate is food, bingo is a game (2026-10-16):** "chocolate bingo" is the food topic's
    words (foodmath_recipes, the recipes picture), so its "bingo" no longer loads the bingo topic too
    - a trigger inside another topic's longer trigger doesn't count (`gopodTriggerOwnMatch`). "bingo"
    alone is the bingo game egg (the Chocolate Bingo kit and card pictures): bingo hosting, then "Hey,
    you can also play Chocolate Bingo...". One egg per line, not two.
  - **Birthday (2026-10-16):** value point 4 - Brobot 0 wants to ask a local bakery to make him a big
    cake based on his foodie cars.
  - **Face recognition (2026-10-14, a starter):** say "Hey Vector, remember my face" - the
    custom intent `REMEMBER_MY_FACE` (customIntents.json) hands the robot Wire-Pod's own "my
    name is CrushN8r", and the robot scans and saves the face itself; "who am I" works as
    before. Voice only (a typed line goes to the chat). Needs `wpr 2` after the edit. The face
    photo as a prize comes later.
  - **Edge TPU (golden, 2026-10-14):** plugged in and turned on, for photo face and object
    recognition. Setup in place: the TPU driver, the `gopod-coral39` Python environment, and
    the face and object models in `goverlord/runtime/data_gomad/models/edge_tpu/`. Turn it on
    after a replug or restart with the command in `goverlord/runtime/data_gomad/models/README.md`
    (`1a6e:089a` before, `18d1:9302 Google Inc.` once on). Wired into the photo egg as
    round 1 (2026-10-16, above); Wire-Pod's worker holds the stick while Wire-Pod runs (one
    program at a time).
- **QR codes (2026-10-14):** any bubble whose words hold "crushn8r.ca" gets the crushn8r.ca QR
  code, and "newsletter" (or "crushn8r.ca/subscribe") the newsletter one (crushn8r.ca/subscribe,
  a placeholder), in the bubble's top-right corner, header row; the magnifier as ever.
  `qr_codes` in `gopod_brobots_settings.json`, pictures in `webroot/qr/`.
- **Weather faces (2026-10-14):** during a run, a bubble holding the live weather line gets
  Vector's own weather face for it (in the live chat the weather is a pop-up instead, above) (sunny, cloudy, rain, snow, thunderstorms, windy; clear at night = stars;
  cold has none), animated, Vector's own size (184x96), in the speaking Brobot's eye colour, in the
  same top-right corner as the QR codes, with the magnifier. Its weather sound is loaded with
  the bubble and plays once as it shows, under the Brobot's voice, at 10% (`weather_card.volume`,
  its own level). Only the weather's own words pick the face, never the
  place's. `weather_card` in the settings; the kit is `webroot/weather/`, all of it shown on
  `gopod_weather_kit.html` (locked as is; linked from Wire-Pod's weather setup section, the
  overlay's `setup.html`).
- **Numpad? and Native? (2026-10-14):** ticking Numpad? on, and 3x 000/NumLock off, are said by
  Goverlord (`numpad_ptt.texts.on`/`exit`, made at start); every Native? and Numpad? click is logged.
- **Photo wake (2026-10-14):** after control, the Brobot waits `photo.awake_seconds` (2.5) - awake,
  head settled - before mirror mode and "3, 2, 1".
- **"Who am I?" by voice (2026-10-14):** `intent_names_ask` is in `native.always_on`, so the
  robot answers it even with Native? off. A typed "who am I?" goes to the chat's LLM.
- **Native on top (2026-10-14):** any intent the robot acts on himself ("who am I", "remember
  my face", the weather...) lets go of his chat hold first, so his own behaviors run
  (`gopodNativeLetGo`, called from Wire-Pod's `IntentPass`); the hold returns with the next
  chat answer. Live 20:22: the hold had kept Brobot 2's face behaviors from running.
- **The Brobots' math (2026-10-14, golden):** a math question in the live chat is worked out
  in code, exactly, and answered step by step, word for word - gemma never calculates
  (`gopod_math.go`, `math` in the settings, locked by `TestMath*`). Fraction to decimal (long
  division, a repeating block barred: 1⁄3 = 0.3̅), decimal to fraction (over 10/100/1000,
  simplified), repeating decimal to fraction (the 10x method: x = 0.3̅, 10x = 3.3̅, 9x = 3,
  x = 1⁄3), percents ("15% of 80", "what percent is 3 of 4", to and from), and plus, minus,
  times, divide with fractions, decimals, brackets and the order of operations. Reads digits,
  LaTeX and spoken words ("zero point three repeating", "three quarters"). A plain sum needs a
  `math.trigger` at the end. Fractions show as
  1⁄3 and are said "over"; "=" is said "equals", "−" "minus" (or "negative" before a number).
  Each step's words are `math.lines`.
- **Explain the Math (2026-10-14, the contest format):** "[statement]. Explain the Math" - only
  a line that ends with it is math. Brobot 1 answers first, then Brobot 2 follows up, whoever
  it was sent to (typed, numpad, All or the robot's own mic). Each says up to
  `math.flavour_sentences` of his own (the LLM's words, any sentence with a number dropped),
  then his share of the steps as bullets ("• Let x = 0.3̅."); Brobot 2 gets the last third (at
  least the last step). In the live chat it's an "Explain the Math" pop-up (no headers):
  Goverlord's congratulations with `math.popup_title` over the challenge, Brobot 1's answer
  under it; Brobot 2's follow-up in his own marquee bubble (the pop-up rule). A
  step's bullet goes with its own line. Explain the Math is a long answer: a fraction to decimal
  shows the long division digit by digit (`math.long_division_max_steps`), and a longer
  repeating block is said digit by digit ("the 1 4 2 8 5 7 repeats"). Locked by
  `TestMathExplainTheMath`.
- **All, in turn (2026-10-14):** a typed or numpad line to All with no Brobot named - Brobot 1
  answers, then once he's done Brobot 2 answers the same line (Brobot 1's answer is in the
  shared chat). A line naming a Brobot goes only to him. `addressing.all_next_wait_seconds`.
- **Numpad released (2026-10-14):** 3x 000, 3x NumLock, or unticking Numpad? all say the same
  released line. 2026-10-16: "Goverlord has released the numpad." and, ticked on, "Numpad is
  mapped and ready to use."
- **Native? spoken (2026-10-14):** ticking and unticking Native? are said by Goverlord
  (`native.texts.on`/`off`, made at start), and logged.
- **Backpack question and its pop-up (2026-10-14):** a line the robot heard is logged once
  its answer is ready, after its pop-up; it goes above that pop-up and the reply still joins it.
- **Voice file generating. (2026-10-14, rule; text and hourglass 2026-10-16):** a line shown with its voice whose voice file isn't
  made yet (a first-time line, live or in a run) shows a grey "Goverlord: GOPOD Brain"
  bubble saying "Voice file generating." with the warm-up's hourglass, rocking and draining (2026-10-16), while it's made; then a Goverlord line fills that bubble and
  plays, any other line takes its place (`reporter_voices.generating_text`/`generating_marker`,
  `gopodReporterFileShown`).
  The Operator bubble the same way (2026-10-16): while its read-out voice is made it shows
  "Voice file generating." with the hourglass (`generating:op`), and its words come back as its
  voice starts to play.
- **Emoji leftovers (2026-10-14):** a joined emoji's joiner and style marks are taken out
  with it, so the robot never says them ("yirl"); locked by `TestEmojiJoinersNotSpoken`.
- **Log opened after a refresh (2026-10-14):** the chat goes straight to its bottom the moment the window gets its height.
- **Pop-up placement (2026-10-14):** once the reply joins a pop-up, the chat aims the reply's
  "[icon] Brobot N:" header at the window's top; if the bubble's bottom reaches the window's
  bottom first, it stops there - no extra space below (rule 5, aimed at the reply header).
- **Sting volumes (2026-10-14):** `grunge.wav` plays at 0.4 to sound as loud as `dJent3b.wav`;
  the Fart Joke's danger sound at 0.25.
- **Chat audio volume (2026-10-14):** every sound the chat page plays - voices, stings, a
  song's sounds and soundtrack, weather sounds - plays at `chat_audio.volume` (35%, was 10%
  until later the same day). A sound's own volume (a sting's, a song step's `volume`) is its
  share of that (0.8 = 80% of it). The background loop keeps its own
  (`background_loop.volume`, 5%).
- **The weather line (2026-10-14):** "Windsor Ontario, October 14, 2026, 17° Celsius or 62°
  Fahrenheit, and overcast clouds, as of 1814 Hours Military Time or 6:14 PM." (Brobot 1,
  24-hour in Wire-Pod) - the robot says "degrees" and "eighteen fourteen hours military time"
  (the operator's military time rules: four digits, "zero" for 00-09, "hundred" on the hour). A place named with its province/state/country is
  looked up there (Windsor Ontario, not Windsor England).
- **A bubble that grows on screen (2026-10-13)** - a new message, a picture loading - is placed
  again, whole, while it's in view.
- **Battery watchdog (2026-10-15):** all session long, every minute, each Brobot's battery is
  read and logged (`GOPOD_BATTERY_WATCH`). Below 3.7 V (golden rule 6), off the charger, he says
  he's dizzy and needs help getting home (`battery_watch.lines`), with his bubble - once per drop,
  again only after 3.8 V or his charger. Live chat only, at the next safe spot: never during a
  reply, a run or a draft. `battery_watch` in the settings, `TestBatteryWatchOncePerDrop`.
- **Stopping a timed song (2026-10-15):** `stop run` (or "stop nap" / "stop baby robots") ends a
  timed song (the nap) before its next step - within a second, its steps are short pauses. Its
  soundtrack stops on the page, narrator bubbles not yet shown go, the one on screen shows its
  words, the Brobots wake, and Goverlord says it stopped.
- **Warm-up message and its voice (2026-10-15):** "LLM Warming Up. Please wait." shows once its
  sting and words are loaded, so it plays the moment it shows (chat rule 2).
- **Narrator pictures (2026-10-15):** a song's `section_pictures` (knobs.json, section ->
  file in the page's `pictures/`) puts that picture in the section's narrator bubble, under
  the lit words, with the magnifier. The nap: `baby_brobots1-5.jpg` (from its `images/`
  folder), one per section, Sleepy Robots to Together.
- **A song's picture (2026-10-13):** `{"note": "picture", "picture": "granny.jpg"}` (one file or a
  list, from the page's `pictures/` folder) shows them in a picture bubble at that step, alone or
  as a combo part. Label the step after it, e.g. `host_say_img-granny_0301`.

### Header icons mirror each other (2026-10-17)

Operator: "RULE: For headers with more than one icon: [icon] [header text] [icon] & [icon]
[input] to Brobots [icon] [icon] * to give the effect of the icons' animations mirroring each
other in sync rather than together in sync, set 2nd & 3rd icons to start at the 1/2 loop point".
Every icon after a header's first plays `assets/brobot_eyes_half.webp`: the same eyes, its
frames starting at the loop's half (2.72 s of 5.36 s), each icon keeping its own eye colour
(`gopod_chat_bubbles.css`). The original `brobot_eyes.webp` is untouched.

### Voices

- Reporter 1 (Brobot 3) `af_heart`, Reporter 2 (Brobot 4) `am_michael`, Goverlord
  `bm_george`, at `reporter_voices.speed`.
- Reporter 0 (2026-10-12): Goverlord's lines, headed "Reporter 0, Goverlord:", said by the
  digital butler (`system_persona`, from Master_CrushN8r.pdf) with its own phonetics
  (`system_say_words`: "Master Crush-En-Ayd-uh", "Gov-er-lord"). Run lines keep the run name first.
- Each line is made once and kept, named by its voice and speed: a new voice or speed never
  overwrites. A new reporter line is made in both reporters' voices.
- The page plays them one after another; a line that must be heard before the run goes on
  is waited for (golden rule 11).
- Locked by `TestReporterFilesNamedByVoiceAndSpeed`.

---

### One event list: the engine and the page in dialogue (2026-10-17)

Operator: "tighten loose chains, loops, timings, etc... * one event list, single source of
truth"; "the web page and terminal must have some sort of coordinated dialogue"; "starts after
wpr 2 / F5 * when I click 'web page's 'Log' button * that's when the action starts from the web
page". From `gopod_notes/ENGINE_BRIDGE_LOOK_SURVEY_001.md`, steps 1 and 2:
- **The page is here only once it can play:** after `wpr 2`/F5, the click (the Log button) -
  before it, Goverlord awaits and runs wait (a browser plays sound only after a click).
- **One event list:** every chat-feed line is in the journal as `GOPOD_FEED <id> <who>: <words>`,
  in order - voices, heard markers, warm-up messages, bubbles. **Live replies are feed lines too**
  (via `llm`), no longer read by the page from Wire-Pod's 50-line log.
- **The page answers:** each second it says how far it has played the feed; the journal writes
  `GOPOD_PAGE played <id> <who>: <words>` under the engine's own line.
- **Heard is the truth:** `GOPOD_HEARD <who>: heard` / `not heard after Ns (going on)` / `no voice
  file`, for every voice the engine waits on; the page reports voices that didn't play
  (missing, blocked) with the next heard marker; a song's reporter and ready lines get `heard` or
  `not_heard` back, and the song engine prints a not-heard one.
- **A voice's cap starts when it plays** (its length + 5 s; 60 s if unknown), so it's never cut
  or left to start late, out of order.
- **"Run paused"** is shown (a Goverlord bubble), not said, to a page that's away; the replay's
  pause is in the journal.
- **Every live reply is accounted for** (operator: "If a reply bubble goes missing or shows
  twice ... can you account for to reveal this possible glitch, in all-logs?"): the page writes
  `GOPOD_PAGE live reply <id> Brobot N: shown` (or `NOT shown - <why>`) under its `GOPOD_FEED <id>`;
  Wire-Pod writes `GOPOD_FEED_CHECK ... the page played past it with no bubble` for a missing one
  and `GOPOD_FEED_CHECK live reply twice in a row` for a doubled one.
- **A collapsed chat is away** (operator: "Fix collapsed chat > \"here\""): with the chat closed
  (the Brobots Chat button), nothing plays and runs wait, as with a hidden page.
- **Eye colours from Wire-Pod** (operator: "eye colours from native wire-pod"): each Brobot's own
  custom eye colour, from Wire-Pod's stored robot settings (`vic.RobotSettings`), becomes the
  bubbles' colour at Wire-Pod's start and every 30 s, never during a run (`gopod_eye_colors.go`);
  the interview runner no longer writes it. A Brobot with no custom colour keeps the page's own.
- **No orphan state file:** `gopod_rich_display_ui_state.json` (written by pha0b and the runners,
  read by nothing) is no longer written.
- See it: `journalctl -u wire-pod -f | grep -E "GOPOD_(FEED|PAGE|HEARD|AWAIT|OPENER)"`.
- Locked by `TestFeedDialogue`.

### Volume: one master, percentages, evened sounds (2026-10-17)

- **The master** is `chat_audio.volume` (operator: "One master default knobs setting all other knobs
  refers to * other story.md knobs are either \"\" for default or \"0-100%\""). Every other volume
  knob - a sting's (`system_stings`), the numpad key sound, the weather sound, the background loop,
  a song step's `"volume"` (the opener's, the Fart Joke's sounds) - is `""` (the master) or a
  percentage of it (`"40%"`). Turning the master turns them all. An older bare number (0.4) still
  reads as 40%. The weather (0.1) and the loop (0.05) were their own levels; they're now 29% and
  14% of the master, the same loudness.
- **Evened sounds** (operator: "evaluate audio files to normalize all volumes to -3db ... tag &
  apply a volume offset to the raw audio, before \"default\" knob setting"): Wire-Pod measures
  every WAV in the voice folder once (its loudest sample) and writes the offset that brings it to
  `chat_audio.normalize_db` (-3) into `webroot/gopod_audio_levels.json` (`gopod_audio_levels.go`,
  every 30 s for new files). The page plays a sound at master x knob x offset. The files are never
  changed; 0 = off. Kokoro's voices peak near -8 dB, so they play louder than before, and the
  stings a little quieter.

### system-set: edits from the chat keyboard (2026-10-17)

Operator: "automate the editing of all song pieces * like ask Goverlord to do it via the chat
window * similar to start run-name, use \"system-set\" ... I type: system-set this-volume 13%";
"terminal-results are sent back to chat bubbles window * ... font-size 50%". Typed on the chat
page's keyboard only. It runs a **fixed list of edits, never a terminal command** (operator: "good
catch"). The answer shows in a terminal bubble (`system_set.role`, text at 50% size): old -> new,
and when it takes effect (F5, `wpr 2`, or the next run).

| Type | Does |
|---|---|
| `system-set` | the list: settings with their values, songs, step fields |
| `system-set master-volume 35%` | a named setting (`master-volume`, `normalize -3db`/`off`, `loop-volume`, `weather-volume`, `key-volume`, `sting-volume`, `run-sting-volume`) |
| `system-set 00_llm_warming` | a song's steps (folder name or `song_id`) |
| `system-set 00_llm_warming goverlord_warm_0101` | one step: its knobs and words |
| `system-set <song> <step> volume 13%` | one field: `volume`, `text`, `seconds`, `pause_seconds`, `hold_seconds`, `buffer_after`, `settle_seconds`, `speed`, `cycles` |
| `system-set <song> <step> text New words.` | the step's `> TEXT:` in story.md |

Each file is edited in place - only that value, everything else byte for byte - and read back
before it's saved; knobs.json and zKnobs.json both, when the step is in both. A combo step's parts
are still edited by hand. `gopod_system_set.go`, locked by `TestSystemSetEditsInPlace`,
`TestSystemSetStoryText`, `TestSystemSetReads`.

**system-locate** (2026-10-17, operator: "system-locate (file to edit setting)"; "Please build
system-locate"): where a thing lives, as `file:line`, to open and edit by hand. Read only, same
keyboard and terminal bubble. `system-locate master-volume` (a system-set setting),
`system-locate <song>` (its files), `system-locate <song> <step> [field]` (knobs.json, zKnobs.json,
and its words in story.md), `system-locate <word>` (every setting whose name holds the word, with
its line and value: `volume`, `sting`, `awaits`). `gopod_system_locate.go`, `TestSystemLocate`.

**The lists** (2026-10-17, operator: "add these, to list the items in set & locate in the response
bubble * for fast reference"): `system-set-list` (every setting with its value, the songs, the step
fields) and `system-locate-list` (every setting's file:line, and each song's files).

**system-undo, system-status, system-help** (2026-10-17, operator: "Add: system-undo,
system-status, system-help"; `gopod_system_more.go`, `TestSystemUndo`):
- `system-undo` puts back the last system-set (all the files it wrote), only if none has changed
  since; again for the one before (the last 20 are kept in `gopod_system_undo.json`, gitignored,
  so they survive `wpr 2`).
- `system-status`: Wire-Pod, the language model (loaded or cold), the warm-up, the chat page
  (here/away), the run going, and each Brobot's last battery reading.
- `system-help`: every system command, a line each.
- **system-check** (2026-10-18, step 4; operator: "system-check, thank you. Document all this"):
  looks for every file the data names and every song step - no changes made.
  - **What it checks:** every file named in `gopod_brobots_settings.json` (sounds, pictures, QR
    codes, workers, models, word lists), the weather faces and sounds, each chapter topic's
    `words:` and `pictures:`, `egg_names` and chapter topic lists against the topic files, every
    song's `story.md` steps against its `knobs.json`, and every file a song's knobs name.
  - **Where a file is looked for:** a page path (`/gopod_reporter_audio/x.wav`) under webroot;
    `~/` in home; an absolute path as is; a bare name in the chipper folder, webroot, the voice
    folder, the pictures folder or the song's own folder.
  - **The answer:** "all good - N things checked", or each thing to fix (25 shown; the rest via
    `system-find GOPOD_CHECK`). Wire-Pod runs it at every start: `GOPOD_CHECK` lines in the
    journal. `gopod_system_check.go`, `TestSystemCheck`.
- **system-search** (operator: "Expire go search, replace with \"system-search\" * \"system-search\" is
  used for live guided search responses * brobots use \"system-search\" for 'search [search-text]'
  like 'stat [run-name]'"; "Please remap/build SearXNG container to belong to GOPOD"): GOPOD's own
  web search, no LLM in it - the `gopod-searxng` container (this machine only, port 8055; started
  by `wire_pod_overlay/searxng/run_gopod_searxng.sh`, its private settings in `~/gopod_searxng/`).
  To the Brobots, a line opening with `search <words>` or `look up <words>` (a Brobot's name first
  at most) gets the top 3 results' snippets as facts, and the reply opens with the first result's
  first sentence, word for word, like the weather line; both Brobots get the same search for 2
  minutes. Typed `system-search <words>`: the results in the terminal bubble. Settings
  `system_search`; `gopod_system_search.go`, `TestSystemSearchQuery`. Go search is retired.
- `system-find <word>` (operator: "I want Goverlord to know how to pull lists like you do";
  "system-find FEED_CHECK or system-find Toronto * yes, please"): where the word shows up - the
  log of the last 6 hours (the journal, read with fixed arguments; the word only filters), newest
  first, 15 lines; then the settings and each song's story.md / knobs.json, 10 lines each, with
  their line numbers (`gopod_system_find.go`, `TestSystemFind`).

### Songs that need no voices or language model (2026-10-17)

Operator: "keep for future songs like Baby Brobots * add: fallbacks to system messages: \"No Voice
Generation Required\" or voice generation check * \"No Language Model Required\" of language model
loading/loaded * to skip in-between messages, a single in-between message = \"\" = the indication to
skip". A song's knobs.json `"voices_required": false` shows `llm.no_voices_text` "No Voice
Generation Required" in place of the voice files part of its warming message; `"llm_required":
false` shows `llm.no_llm_text` "No Language Model Required" in place of the language model part
and of "Language Model Loaded, Master CrushN8r.", and calls no model. The nap has both false. A
`warmup_lines` line with no words is skipped; in an LLM warm song, `reporter_lines` with an empty
`> TEXT:` is skipped and `opening` plays the reporters' opening (`00_llm_warming`). Locked by
`NoneRequiredWarmupTests`, `TestOpenerSkipsEmptyLines`.

### (music) places the sting (2026-10-17)

Operator: "new rule: if \"(music)\" is in the story.md * display \"(music)\" visual for the deaf *
play djent * simpler? clearer?". A Goverlord line that starts with `(music)` has the sting
(`system_stings.default`, dJent3b) before its words; one that ends with `(music)` has it after.
`(music)` is shown, never said. The warm-up's end sting doesn't play again when the ready line's
own `(music)` played it (`gopodMusicMarks`, `gopodMusicPlayed`). A line with no `(music)` keeps
the old rule.

### GOLDEN: a voice is made before its bubble shows (2026-10-17)

Operator: "NEW GOLDEN RULE for bubble messages * any operator/reporter/butler voice being generated
during a live run auto-displays: Generating + Hourglass message, under the header * upon audio
ready, display/play message"; "At every step ensure next step is fully ready (bubble
audio/visually 'ready')". A line whose voice file isn't made yet shows "Voice file generating."
(`reporter_voices.generating_text`) with the hourglass under its header - an operator line, a
reporter's, a Goverlord bubble, and the opener's System llm bubble alike - and its words and voice
come once the voice is made. A voice asked for while Kokoro is still starting now waits for it
(up to `start_wait_seconds`) instead of failing (live 11:01: the opener's two lines were shown and
never said). The stings (grunge, dJent3b) and the numpad key sound are loaded once and replayed;
the background loop was already decoded once and looped.

### The opener says a message line by line (2026-10-17)

Operator: "llm run doesn't support multi-line input * split multi-lines into separate lines *
generate line by line * upon generated & audio loaded, display/play line by line". In the opener
song, each line of a message (and of the ready message) is its own voice file; the bubble shows
the lines so far, and each line's voice follows it.
Every line's voice is made first (the bubble showing "Voice file generating." meanwhile), then
line 1 is shown and said, then line 2 - the same way as a one-line message. A `(music)` anywhere
plays its sting where it's written.

### Multi-line words in story.md (2026-10-17)

Operator: "this multi-line edit broke the code ... Can you upgrade globally to support multi-line
edits? * system messages * reporter lines * other scripted responses?". In every story.md - songs,
the opener, the interview - a line right under `> TEXT:`, with no blank line between, is the
words' next line, and the bubble shows the line break; a blank line ends the words. The song
engine already read it so; Wire-Pod's reader (`gopodInterviewStory`) now does too, and
`system-set ... text` replaces every line. Locked by `TestStoryMultiLine`.

```
## STEP goverlord_warm_0102

> TEXT: Voice ready to play.
Language Model Loading...
```

### The egg list and the interview (2026-10-17)

- **The egg list** (operator: "'Hey that reminds me' thought bubbles ... center around a list of
  easter eggs (chapter/topics/value-points) * during live chat, whatever easter eggs not
  discovered, are left on the list to mention ... like a fallback list of sellable features the
  brobots have, as GOPOD reps"): every topic with a prize (pictures or a card), in the chapters'
  order, then `value_points.reminder_eggs` (the weather, Explain the Math, take a photo). An egg
  found since Wire-Pod's start leaves the list. Each "Hey, that reminds me..." names at least one
  egg still to find; a question on no topic still ends with two (not a thank-you); each in turn,
  then round again (`gopod_egg_list.go`, `TestEggList`).
- **The interview, rebuilt** (operator: "rebuilt with new questions and answers from scratch, as a
  summary of that same list"): one question per egg - the workbook, fractions and decimals, the
  rig, the rest of the cast, birthdays and merch, Bingo nights, foodmath, the weather, roast
  battles, our legacy - then the closer. The Brobots' answers are the operator's own lines from
  the topic files, word for word; the reporters' questions are drafts. The old one is in
  `gopod_notes/older_archives/interview_before_2026-10-17/`.
- **Chat memory, no conflict:** Wire-Pod keeps each robot's chat in memory, and GOPOD's shared
  Brobots chat replaces it in the prompt (`kgsim.go`), so nothing is fed twice. Both are in memory
  only: `wpr 2` starts both empty, and the page clears with the new boot; F5 redraws the page and
  touches neither. The session log keeps everything.

### Live fixes from the 2026-10-17 run

- **All, starting with Brobot 2** (operator: "Starting with Brobot 2 ... brobot 2 answers, brobot [1]
  follows up * but they maintain their slightly different coloured personas"): "starting with Brobot
  2" / "Brobot 2 first" (or Pip) to both turns the order round; each keeps his own persona
  (`gopodAllStartsWith`, `TestAllStartsWith`).
- A follow-up doesn't end with his brother's end line ("Hey, you can also play Chocolate Bingo, for
  kids of all ages!" twice); a "(wait for Brobot 0 ...)" note and a "---" at a line's end are dropped
  (`TestWaitNoteDropped`).
- **Flags as pictures:** a known country's flag is drawn from `assets/flags` (Twemoji, CC-BY 4.0) -
  some browsers (Windows) drew a flag emoji as its letters ("US").
- **Link underline:** while a link's words are revealed, only the shown words are underlined.

- **A cue word that's a real word is said:** "react, celebrate small wins," was said "react, small
  wins," - Wire-Pod dropped an animation's name at the start of every spoken piece. Now it's
  dropped only when what follows starts anew ("happy Hey there"); "celebrate small wins" is said
  (`stripSpokenActionLead`, `TestSpokenCueWordKept`).
- **Brobot lingo** (operator: "brobot lingo: choco[text] = chalk-o-[new text] * chocolateric =
  chalk-o-latteric"; "chalk-o- is intended for brobot made up lingo ... try to not include normal
  words"; "math = madth"): a lingo entry can be `"word_start": true` - at the start of a made-up
  word, the rest kept ("chocoyum" -> "chalk-o-yum"); a dictionary word (chocolate) and its
  `except` words are left alone. Both Wire-Pod and the song engine. `TestBrobotLingoChoco`.
- **Weather, the country named** (operator: "[City], [state/province], [Country] = Toronto, \"\",
  USA * so weather station has at least City/Country ... add country correct flag"): a country in
  the place ("Toronto USA") is sent to the weather service as its code ("Toronto,,US") and only a
  place in that country is taken; the line shows the place's flag after it ("Toronto Ohio 🇺🇸,
  ..."), kept by the flags rule (`gopodWeatherCountry`, `TestWeatherCountry`).
- **Egg bubbles** (operator: "Congratulations! / Easter Egg '[Title-of-Easter]' found. / [images] /
  MouseOver to Zoom In * MouseOver to Zoom In goes below the images"): `egg_text` is two lines,
  the pictures follow, and `egg_zoom_text` "MouseOver to Zoom In" goes under them.
- The Bingo line is "Sure, let's play! Say start Bingo test!"; "experience" is no longer a Bingo
  egg word ("How was your test experience?" opened an egg named 'Experience').
- **Finer word reveal** (operator: "Split away punctuation from words * reveal the word, then the
  punctuation, then emojis"): a word's punctuation, emojis and ° are steps of their own after it
  (`word_reveal.seconds_per_mark`); marks inside a word stay (Let's). A clock steps "1", ":", "43",
  "P", "M"; a year after its date "20", "26" and is said "20 26" (`gopodSpeakYears`); a flag is a step.
- **Weather place** (operator: "display [country][flag], say country"): "Windsor Ontario Canada 🇨🇦";
  a flag is said as its country unless just named (`gopodSpeakFlags`, `gopodCountryName`).
- **Reporter lingo:** merch is said merch (the reporters' "mirch" is gone; the Brobots keep theirs).
- **QR codes** (operator: "2 QR CODES"): QR1 top-right; scrolled above the chat's top it jumps to the
  bubble's bottom-right (`corner-low`, marquee inset kept). QR2 goes right after "...subscribing to
  our newsletter!", the last step of the reveal (`qr_codes.second`).
- **Get Photo, 3 images** (operator: "[raw snapshot][snapshot enhancement attempt][TPU object
  recognition on enhancement]"; "TPU with bounding boxes"): the TPU worker enhances the snapshot,
  finds objects in it and draws their boxes; the egg shows all three; moondream looks at the
  enhanced one (`photo.quick_look.pictures` / `enhance`, `_enhanced.jpg`, `_tpu.jpg`).
- "starting with bro bought to" (a mic mishear) now counts as "starting with Brobot 2".
- **Pictures in one row** (operator: "scale the 3 images to fit in a row"): a bubble's pictures
  share one row, each scaled; the magnified one takes the full width.
- **Song/system lines said by a Brobot** (`gopodRecordSay`, e.g. the photo's second "I see ...")
  go through `GopodSpeakSentences`: lit while said, shown as said after (was stuck solid white).
- **Search results** (operator: "Don't speak or highlight search results * centered
  rounded-rectangle bubble * font @ 50% * fast character by character reveal * then brobot
  responses"): one centered bubble (`system_search.results_role`), typed in once it shows; the
  replies no longer open with the result.
- **Eggs are chapters** (operator: "brobots are GOPOD reps * GOPOD = AI Wordplay! Explain the
  Math! * easter eggs are live run chapters * each easter egg should have a set title (chapter
  title) * topics in chapters are the easter egg triggers * then the value points"): one egg per
  chapter, "Easter Egg '<chapter title>' found.", found by its topics' words, with those topics'
  pictures; no picture = no pictures and no zoom line. The roast battle (in no chapter) stays its
  own egg; Explain the Math stays. The egg list and reminders go by chapter
  (`gopodTopicChapter`, `TestChapterEggs`).
- **Audio evened the operator's way** (operator: "I aim to adjustment a bit more than average
  loudness, to as close to but within the -3db range * then loudness over -3db I compress back to
  -3db"): every sound is measured for average loudness (ffmpeg EBU R128), turned to
  `chat_audio.loudness_lufs` (-16), then limited under `normalize_db` (-3; limiter at -3.5 so the
  true peak stays in). Made voices are done in place as they're made; stings, song sounds,
  soundtracks and the weather mp3s get done copies in `webroot/gopod_audio_norm/` that the page
  plays (`normSrc`, `gopod_audio_levels.json` "made"); originals untouched. Replaces the peak offsets.
- **Two kinds of eggs** (operator 2026-10-18: "easter eggs hidden (roast battle, halloween) are
  surprise easter eggs * other eggs are value-point-triggered during live & for the interview"):
  **chapter eggs** are on the egg list - the reminders name the ones not found yet; **surprise
  eggs** (`topic_pictures.not_listed`: Roast Battle, Halloween) pop up when their words come up but
  are never named. Halloween: `_always/03_halloween.txt`, an emoji card (`cards.halloween`), no
  picture; its say lines are drafts.
- **The window's numbers are data** (operator: "ensure flexible data/info values/variables are in
  loadable files * code is dumb flow guiding machine"): the page's timings and sizes (baton,
  scroll, countdown hold, marquee, narrator fades, reveal fallbacks) are `chat_window` in the
  settings; the numbers in the page code are only fallbacks.
- **Leak fixed (2026-10-18 run):** a reporter's or Goverlord's voice that errored or hit its cap
  left its line lit, and the baton held every next bubble up to 2 min (`GOPOD_PAGE held` 28-43 s,
  eggs "not heard"). Now however a voice ends its line is said; and a line lit longer than
  `chat_window.speaking_max_ms` (30 s - a frozen robot, a lost end signal) never holds the baton.
- **Eggs on the bubble machine:** Goverlord's congratulations is revealed as it's said
  (`voiceLitLine`), like every voice.
- **"Got it!"** replaces "Thinking, please wait." (`chat_moves.thinking_text`), everywhere it's said.
- **A key that isn't live** (KP9, Cameo) plays grunge quietly as an error cue
  (`numpad_ptt.error_sound`, 15%).
- **Native intent bubble** (operator: "\"Who am I\" ... show a generic silent system bubble:
  wire-pod intent activated"): a native Wire-Pod intent sent to a Brobot shows "Wire-Pod intent
  activated." in a silent Terminal bubble (`native_intent`, `gopodNativeBubble`); his own words
  aren't known to Wire-Pod.
- **system-check: intent overlap** (operator: "warn when a topic's trigger words include one of
  Wire-Pod's own phrases"): a topic word that is, or holds, a Wire-Pod intent phrase
  (`intent-data/en-US.json`) or a GOPOD command phrase (`native_cube.intents`) is listed.
- The Brobot 0 topic is found by "crushn8r", not "ron" (operator: "to avoid confusion").
- **Step 6, piece 1:** the old peak-offset code (replaced by the loudness method) removed.
- **Forced alignment** (operator 2026-10-18: "word for word * punctuation highlights at a silent
  point threshold, after the ending word"; "timing file per voice * the page reading it * existing
  voices need remaking * new ones get them automatically"): the voice worker saves Kokoro's own
  word and mark times beside each voice (`<name>.words.json`); a voice without one is made again
  on its next use; the page reveals a reporter's or Goverlord's line (and an egg's congratulations) word by word
  at those times (`alignUnits`) - a mark or emoji when its word ends, a word said another way
  ("crushin-ate-a") between its neighbours. The Brobots' replies (Vector's own voice, no audio
  here) keep the timed estimate.
- **Voices timed by ear** (operator: "Your approach seems to look better than kokoro * can GOPOD
  redo voices as per your timings?"): the voice worker now has Vosk listen to each voice it makes
  and gives every word the heard times (Kokoro's only for a word Vosk missed; a mark where its
  word ends) - `"aligner": "vosk"` in the `.words.json`. Older timing files are remade on next use.
- **Numpad sounds** (operator: "live KPN = no grunge audio as this is not an error * \"Too short
  ...\" = error = grunge audio plays * not live yet key press = error = grunge audio plays"): a
  key that works plays nothing (`key_sound` off); each error message (`numpad_ptt.error_texts`:
  too short, nothing heard, no mic, not ready, failed, NumLock off, not live) plays grunge at 15%.
- **One reveal, any source of times** (operator: "Do all bubbles inherently support/detect forced
  alignment? rather than patching"): the word reveal (`digitSequence` + `alignUnits`) uses word
  times whenever a piece carries them, whatever made them - a Kokoro voice's `.words.json`, or a
  song's `narrator_words.json`. **A song's narrator** (operator: "grab timings from the audio file
  ... apply forced alignment timings, following the way kokoro does it ... discard the audio
  file"): `gopod_probes/tools/gopod_align_captions_001.py <narrator audio> <song folder>` - Vosk
  (Wire-Pod's own speech model) times the words it hears; each SRT caption's words are matched in
  order within its seconds; saved as `narrator_words.json` (Kokoro's shape); the audio isn't kept.
  The song engine puts each caption's times on its piece (`@s-e ~<base64> words`). Baby Brobots:
  done (170 of 179 words heard; the rest between their neighbours).
- **The managers** (operator 2026-10-18: "CHALK & PLAYHEAD can take over system commands ... I see a
  lot of ways that reduce the load on the brobots"; "PLAYHEAD * he now takes over 'time &
  location' stuff * WAWN, where are we now? * weather * his bubble background is tape ribbon
  color of vintage video tape cassettes"; "CHALK (display) now is to take ownership of:
  search-results box (same smaller size, no audio) * system-status and the other system bubbles *
  hidden/surprise egg cards"): `managers` in the settings, `gopod_managers.go`.
  - **PLAYHEAD** (Reporter 2's voice, tape-brown bubbles): "weather" - "I'll help you with that.",
    the weather line (no egg since 2026-10-18: "please remove easter egg status from weather * no more weather pop up. too common"); "what time / where are we / WAWN / what's the
    date" - the time and place line (`time_place_text`). The Brobots keep the facts in their
    prompt but don't open with the weather.
  - **CHALK** (Reporter 1's voice, chalkboard bubbles): draws the search results, the system
    bubbles (system-* answers, Wire-Pod intent, numpad status) and the surprise eggs (Roast
    Battle, Halloween) on its board (`display_host`).
  - Reporters' speech follows the Brobots' rules for a flag after its country (not said twice) and
    military time.
- **Three personas, three jobs** (operator 2026-10-18: "shared brobots persona = GOPOD reps * no
  system stuff anymore * Shared reporters persona = scripted + voices * reporter 1 & 2 lines =
  scaffolding for CHALK & PLAYHEAD live chat personas * QR codes handled by CHALK"):
  - **KP7 = CHALK BORED, KP8 = PLAYHEAD ZUMINTU** (live in the numpad map, voice-only): the
    operator's line, then the persona's answer from the language model - its identity and its
    reporter's scripted lines as the style - in its reporter voice and colour
    (`managers.personas`, `gopod_personas.go`). "Who are you?" (`who_words`) gets the operator's
    own line, word for word (`who_line`).
  - **QR codes:** CHALK shows them in its own board bubble after the Brobot's, each with its name
    (the call to action's too) - not in the Brobot's corner (`chalkQrs`).
- **Managers inside a Brobot's bubble** (operator 2026-10-18: "If I ask CHALK, PLAYHEAD, brobots or
  'All' = take over trigger * CHALK or PLAYHEAD takes over * If I ask brobot N ... PLAYHEAD bubble
  displays inside brobot N bubble * at top before brobot N 'Got it' ... CHALK bubble displays
  inside brobot N bubble * at bottom after brobot N response * QR codes are brobot N's bubble *
  thought bubble"): a line to Brobot N - PLAYHEAD's part (weather, time/place, a location QR) opens
  his bubble at the top (`nestShell`, `nestBox`); his reply fills the same bubble below; CHALK's QR
  codes go in at the bottom (`chalkQrs`). His own highlight ignores the nested parts (`ownSays`).
  To the Brobots or All (operator: "Once managers have dealt with the brobot that asked. Their jobs
  are done. They can continue chatting * Brobot 2, if also asked responds like before"): inside
  the bubble of the Brobot answering first; the other answers like before, no QR note. To a
  manager: the managers' own bubbles; a take's words (weather, time/place) asked of CHALK or
  PLAYHEAD go to the take (PLAYHEAD's real weather), not the persona's language model.
  **Newsletter** (operator: "please put: * CHALK [QR code] Newsletter: crushn8r.ca/subscribe *
  before wanna know more CTA"): CHALK's part with the `qr_codes.second` QR and title, inside the
  Brobot's bubble just before the call to action (`chalkCta`). **Location QR:** a web address in
  the line - PLAYHEAD says "I can help with that. QR code generated." and shows a QR made for it
  (`url_qr`, `webroot/qr/made/`); the Brobot's prompt says it's shown (`qr_note`). When the first
  Brobot's answer ends, CHALK says "I can help with that. QR code displayed." and shows the same QR
  (`qr_display_text`, `gopodManagerQRShow`) - at the bottom of his bubble for a line to him, her own
  bubble otherwise - then the other Brobot answers or doesn't. The QR is made with the operator's
  own Python libraries (`~/.local`), since Wire-Pod runs as root.
- **Operator bubble for a numpad line** (operator 2026-10-18: "Operator bubble: (not manager
  bubble) * KP7 to CHALK / keep holding… / Recording for CHALK / Got it - transcribing… /
  ["Generating..." + hourglass]/[transcription]"): the key's status bubble becomes the Operator
  bubble - Operator header, status lines kept, the words (with Generating while read aloud) under
  them. Lines to CHALK/PLAYHEAD are read in the Operator voice too.
- **Manager bubble** (operator: "all [message0]/[message1] in the same bubble"): a manager's next
  line joins its bubble from the exchange (own, or its part in a Brobot's) (`hostLineInto`).
  Managers only - reporters' bubbles unchanged.
- **GOLDEN global message [message0]/[message1]** (operator 2026-10-18: "no more "Generating..." as
  a separate bubble * decouple generating + hourglass into its own 'global message variable'
  available to all bubbles * message0: "Generating..." + hourglass * message1: generated playable
  voice wav file + [transcription]"): every voiced line waits in its own bubble. Wire-Pod sends a
  line whose voice isn't made yet first, marked `generating_marker+"wait:"` (`gopodSayShown`, the one
  path for shown voice lines); the page shows generatingText and the hourglass in that line
  (`voiceLitLine`) until its voice plays, then its words; "off" (voice failed) shows the words. The
  Operator's own bubble uses the same text and hourglass (`slotWaiting`). No separate
  "Generating..." bubble any more.
- **system-voices-rebuild [go]** (operator 2026-10-18: "requires the wpr 2 & F5 * temporarily back
  up all current used generated voice audio files * system message, ready for a restart * upon
  restart, GOPOD automatically does a voice files generation check * rebuild/generate all new voice
  audio files"; "anything questionable? pause & ask"): system-check first (story.md, knobs.json,
  zKnobs.json, every file named) - a problem pauses it ("go" rebuilds anyway); lists the scripted
  lines that need a voice; moves every generated voice to `webroot/gopod_reporter_audio_backup/`
  (sounds stay); "Ready for a restart". The start's generation check makes every line again; each
  voice's backup copy goes once its new voice is made. `gopod_system_voices.go`. system-check now
  also matches zKnobs.json's steps to knobs.json.
- **Voice keeping, no auto-deletes** (operator 2026-10-18: "no auto-deletes * delete backup upon
  successful new playable voice wav file * operator_keep_days for system lines * scripted_keep_days
  for scripted lines (reporter/narrator/etc) * live_keep_days for live lines * I think we already
  have an archiving process, use that to start from"): each voice's kind is noted as it's made
  (`voice_kinds.json`): system = Goverlord's voice, scripted = the start's check or a playing script,
  live = the Operator's lines and the managers' live answers. At each start a voice not in use and
  not said for its kind's days moves to `sessions/archives/voices/` (the session chats' archive).
  Saying a line again resets its days.
- **Editing from the chat window** (operator 2026-10-18: "I want to be able to edit everything
  possible in GOPOD * aim for maximum automation, minimal editing"; "Build all 5 system- commands
  please"; the lists: `gopod_notes/CHAT_WINDOW_EDITING_LISTS_001.md`): `system-set
  <path.to.setting> <value>` (any setting; keeps the old value's kind), `system-add` /
  `system-remove <path> <item>` (list items, or `key = value` in a group), `system-lingo
  <brobots|reporters|butler> <word> = <said>` (remove <word>; the lingo alone lists it - Brobots in
  the template card, reporters `say_words`, butler `system_say_words`), `system-topic` (a chapter
  topic's words, points - say lines and numbered points kept as one list - label, reminder,
  pictures), `system-restart [go]` (Wire-Pod restarts itself; the page reloads itself on the new
  start; click Log for sound). Every write goes to system-undo; a settings edit Wire-Pod wouldn't
  start with is put back at once, and system-restart refuses to restart on one.
  `gopod_system_edit.go`.
- **Operator line and the system menu** (operator 2026-10-19: "New Operator Line: Brobots ()1 ()2
  ()All [Type to the Brobot, press Enter] | Record? [] | HH:MM:SS 🔇 * make Native?[] Numpad?[] part
  of system- commands * prune Native?[] & Numpad?[] altogether"; "🔇 opens it * 🔊 closes it *
  system bubble contains drop-down-menu bubble * full inner width/height * menu system should lead
  me like a tour guide"): the two boxes are gone - `system-native on|off`, `system-numpad on|off`
  (the page switches the numpad as the line is sent); their spoken on/off lines are gone too
  (native.texts pruned; 3x NumLock still says the numpad's exit). The speaker icon turns the loop on
  and opens the system menu - a Goverlord system bubble at full width holding the menu bubble - and
  turns it off and closes it. The menu is `system_menu` in the settings: groups in tour order (1
  Restart and check, 2 Play, 3 Edit, 4 Find, 5 Apply, 6 Switches, 7 Test one thing, 8 Wrap up), one
  open at a time, each with its guide line and what's next; `run: true` items run on click, the rest
  wait in the text box for your words. Menu at 50% font (operator: "Make [menu system- bubble] font
  size 50% for more room below * |menu system- bubble| |nested scroll bubble|"); under it, its
  results - each system- line sent while it's open and its answer (system-search's results box
  too). Its look is `system_menu.look` (font_size, background, code_background, code_color,
  results_background): `system-set system_menu.look.<name> <value>`, F5 - also an item in menu
  group 3. Search results (system-search and "search ...") are PLAYHEAD's
  (`system_search.results_host`; operator: "system-search = location = PLAYHEAD, not CHALK").
- **PLAYHEAD ZUMINTU's job** (operator 2026-10-19: "job role: was pha0b * 'ph' for PLAYHEAD * a/0/b
  are the working points to zoom in/out from * ZUMINTU = zoom into * like editing an audio file end
  to end"): in his persona prompt (`managers.personas.PLAYHEAD.prompt`).
- **Brobots group** (operator 2026-10-19: "Bring all of it in a new Brobots group in the menu"):
  Wire-Pod's bot page as commands, via its own `/api-sdk` routes on this machine -
  `system-robot <stats|volume|eyes|locale|units|location|timezone|say|wake|head|lift|wheels|camera>
  [1|2|both]` (moves: control, 1.5 s settle, a short nudge, stop, release), `system-faces`,
  `system-photos` (thumbnails, magnifier), `system-version` (installed, newest, which of GOPOD's
  nine overlay files changed upstream). `gopod_system_robots.go`. Menu group 7.
  Every robot call goes through `gopodRobotCall` (operator: "for all robot calls add in a robot
  wake beforehand then continue * ... robust enough to ignore distractions until that call is
  complete (no overlapping chains)"): one call at a time per Brobot, the chat's hold let go,
  behavior control taken and granted (8 s at most, else "didn't wake"), the 1.5 s settle, the call
  run to its end while control is held, then given back. `say` goes over the SDK and reports the
  robot's own finish state (Wire-Pod's say route always answered "success"). Also `cube <name>`
  (native_cube intents, control given back so his behaviour runs), `anims <word>` (the robot's own
  animation list - `anims sleep` lists every sleep one) and `anim <name>`, `camera` (Wire-Pod's live
  `/cam-stream` as a picture in the chat).
- **2026-10-19 fixes** (operator): "restart default value = background loop icon OFF" - every new
  Wire-Pod start sets the speaker off (no sound, no menu until clicked). The 126 sleep beats
  (`gopod_notes/older_notes/SLEEP_BENCH_ALIASES_001.md`) are `sleep_beats` in the settings:
  `system-robot beats [family]`, `system-robot beat <alias>`. A numpad status bubble is an Operator
  bubble, "Numpad to <who>" ("update header as per template version ... 'Numpad to ...'"), no
  longer CHALK's; while the menu is open it's a line in the menu's results ("no other bubble when
  system menu is open"; the error sound still plays). A drag bar between the menu and its results
  resizes them; while the menu is open the chat is scroll-locked at its top.
- **GOPOD VISUAL CONTROL - the TPU +/- rotation** (operator 2026-10-20: "Focus on adding the tpu +/-
  rotation"; the retired 8011 demo's ring, wdtm_demo.html): `visual_sources` - Doc, Pip, the USB
  webcam. KP+ next feed, KP- previous, wrapping (page, numpad on); `system-visual [next|prev|<id>]`
  typed. Each turn grabs a frame (a Brobot woken through gopodRobotCall, or the webcam with ffmpeg),
  runs the Edge TPU quick look (objects + the boxes copy), keeps `webroot/visual/<id>.jpg` (not in
  the repo); a source that can't give a frame shows its last one, CACHED. One visual board: FEED
  n/N, name, LIVE/CACHED, TPU objects, the frame (magnifier) - CHALK's board, or the menu's
  results while open; turned in place while it's the newest thing. Route `/api-sdk/gopod_visual`,
  registered from gopod_visual.go (server.go untouched).
- **From the old pages and the JSON viewer** (operator 2026-10-20: "do all 8 please";
  `gopod_notes/PAGES_AND_VIEWER_ABSORB_SURVEY_001.md`): `system-cams [tpu]`, `system-chats <words>
  [operator|brobots|managers]` (every saved session, matches marked >>like this<<),
  `system-chat-save [session]` (chat.md in the session folder), `system-menu-theme <preset>`
  (`system_menu.themes`), `system-numpad` alone lists the keys; "camera feed(s)" is a surprise
  egg (`chapters/_always/04_camera_feeds.txt`, not_listed) - the Brobots' live camera streams on
  CHALK's board (a picture path starting with "/" is used as it is).
- **Doc & Pip workbook cards** (operator 2026-10-20: "From DocPipMath.pdf * use all applicable
  cards in the pdf (correctly rotated, lol)"): the six cards (1A Factoring Warm-Up, 1B One Big
  Slot, 1C Speed Round, 2A Doc's Answer Key, 2B Flip to the back, 2C No credit for close), cut
  from the PDF's pages turned upright, are `webroot/pictures/docpip_*.jpg` - the workbook egg's
  pictures (`04_how_doc_and_pip_teach/02_doc_and_pip_workbook.txt`), each with hand-written
  points from its own words (`gopod_picture_descriptions.json`).
- **Wait until I'm finished** (operator 2026-10-20: "add a 'wait until I'm finished' flag to the
  chat bubble template"; live: a long Operator voice was still playing when the flat 30 s cap ran
  out, and the eggs and Brobot 1 went on): a voiced line waits until the page has played it -
  its own length plus up to `reporter_voices.finish_wait_seconds` (120) - while the chat page is
  watched and here; with no page, the old cap (2 x max_wait_seconds).
- **CHALK QR boards are top/bottom border phobic inside a Brobot's bubble** (operator 2026-10-20):
  the newsletter QR board sits on the inner side of the newsletter line - after it when it opens
  the answer, before it otherwise (`chalkCta`); any newsletter line counts, not only the CTA's
  own words.
- **2026-10-20 run fixes** (operator): bubble types - Operator = input; Goverlord, Gomad, Gominion
  = output. Speaker icon: hovering shows a vertical volume bar (0-100%, the loop's share; kept per
  browser). `system-usb` / the menu's live USB panel (every 3 s, "+ plugged in" / "- pulled out",
  `/api-sdk/gopod_usb`). Picture rows go in the menu's results while it's open. Edge TPU: the
  worker had broken because gopodAsOperator put the operator's Python 3.10 packages in front of
  the coral env's 3.9 ones - a conda/venv Python now keeps its own; `system-tpu [restart]`,
  `system-test-tpu` (in the audit). One feed (KP+/KP-, system-visual): photo, enhancement, TPU
  boxes (no "I see" - operator: "I gave you that as a reference for where to look, not to
  implement"); KP+/KP- turn only through the feeds there now, with one feed nothing turns
  ("no multiple feeds to rotate = not applicable"); all feeds (system-cams, `?dir=all`): every feed's photo
  in a row, then the selected one's TPU boxes (a 4th picture with a webcam plugged in).
  `system-brobot` (system-robot still works) answers what it set ("location set to ...").
- **Mic missing** (operator 2026-10-20: "if no mic or other expected devices not found * after llm
  ready run fully completes * Goverlord knows to send out a mic missing message & disable the
  input only, so numpad works as dry run with no mic"): after the warm-up's last step,
  `gopodMicCheck` looks for `numpad_ptt.mic_usb_id` in lsusb; missing, Goverlord says
  `mic_missing_text` and a numpad key works dry - the same messages, "Mic disabled" in place of
  "Recording", then "transcribing" and "try again" (nothing_heard). Cams: one row in the results;
  a single Brobot's photo shows photo, enhancement, TPU boxes; two or more show each photo, then
  the selected one's TPU boxes. `system-tpu restart` answers when the worker is ready.
- **Review fixes (ultrareview 2026-10-20, 10 confirmed):** only a system command's picture row
  ("menu:" mark) goes in the menu's results - songs' and topics' stay in the chat; a run owns the
  Brobots - KP+/KP- and system-visual show the last frame then, no wake; frame URLs keep their ?t=
  stamp; the webcam grab times out after 8 s; system-tpu restart can stop a worker still loading;
  a failed lsusb isn't "mic missing", and a missing mic is looked for again on each key press;
  system-chats skips the searches themselves and slices on character boundaries; the volume
  slider's value gives way when the settings' loop volume changes.
- **Skip the reporters' warm-up** (operator 2026-10-20: "how to add keyboard 'Esc' key press to skip
  to the end of the in-between run in the llm run * how to add a 'skip' command?"): Esc (outside the
  text box) or `system-skip` (typed or menu group 1) stops the reporter line being said and drops
  the reporters' queued lines and voices up to the next line that isn't theirs or the opening's
  end. Heard markers still pass, so Wire-Pod and the songs go on. Managers and Brobots are never
  skipped. Page-only (`skipReporters`).
- **System window rule** (operator 2026-10-19: "No header ... to recover window space * results
  window stretches to full width even when empty * no horizontal space, duo sync vertical height
  adjustment * both start at 50% window height * system menu/results, 50%/50% 10%/90%, 90%/10% *
  '/' indicates where cursor becomes horizontal space grabber"): no header; menu and results
  start 50/50; the thin bar between them (row-resize cursor) sets the menu's share, 10% to 90%,
  the results take the rest; results full width, empty or not. The bar has a grip line and stretches
  across (the system bubble centres its contents, which had left it with no width). GOLDEN, locked
  2026-10-19 ("Perfect. I like how window 2 overlaps window 1 ... Lock that in"): this layout; the
  menu a little wider (right padding 2px), each window's scroll bar in its own lane, clear of the
  rounded corners. Change only with his OK.
- **Wire-Pod updated 11e7b22 -> dc6ae0c** (2026-10-19): upstream changed 2 of the 9 overlay files;
  kgsim.go's OpenAI model-setting fix carried into GOPOD's copy; upstream's MaxCompletionTokens
  switch not carried - GOPOD's copies keep their tuned MaxTokens 5120 for the local LLM.
- **pha0b and phcal in the chat** (operator 2026-10-19: "What else from the survey, can be absorbed
  over from PLAYHEAD aliases to new menu?"; "Build it all in please";
  `gopod_notes/PLAYHEAD_ALIASES_ABSORB_SURVEY_001.md`): `system-play <song> [from <step>] [to <step>]
  [robot 1|2] [dry]` (steps by id or number; golden-engine songs), `system-gap
  <seconds>|off|default` and `system-tweaks on|off` (its switches), `system-calibrate <move> <1|2>
  [values] [dry]` (phcal_isolate_001.py), `system-score <song>`, `system-set <song> tempo <s>` and a
  step's `tempo_factor` (tempo-set), `system-test-audio`. All start
  `gopod_probes/tools/system_play_001.sh`, handed to the operator's user like the song-start
  intents; the engine checks in with Goverlord (one run at a time). Menu groups now 1-9 (7 Tune
  moves).
- **system-test-*** (operator 2026-10-18: "build a full set of system-[test-commands] to test
  individually, like weather offline, etc... * and be able to chain system-[test-commands] into a
  comprehensive audit/checklist run"): `system-test-list`; `system-test-<name>`; several in a row;
  `system-test-all` = the audit in `system_tests.audit`'s order. Tests: files, python, voices, llm,
  search, weather-live, weather-offline, time-place, takes, qr, personas, eggs, page, robots. Each
  checks one thing, nothing said or moved; PASS/FAIL with what it saw. `gopod_system_tests.go`.
  (The session dry run loads no weather key, so its weather is always the offline line.)
- **Wire-Pod runs as root** (operator: "harden in code, so as not to have to remember"): every
  Python start goes through `gopodAsOperator` (GOPOD_HOME and the operator's ~/.local packages);
  `TestPythonRunsAsOperator` fails the deploy if one doesn't. Without it, the QR codes and the
  voices' Vosk timing failed silently. A voice timed by Kokoro's estimate is made again, heard
  (`"aligner": "vosk"`).
- **Eggs taken out** (operator 2026-10-18: "confusing overlap between system managed & easter egg *
  remove easter egg: 'Websites, themes and the channel'"): `topic_pictures.not_eggs` - never pops
  up, not on the egg list; the chapter stays in the interview.
- **Warm-up while voices are made** (operator: "I don't see/hear: Voice files generation check
  while new files generating * I see old message lingering"): the step's first line shows at once
  with "Voice file generating." under it, then on its own once made.
- **flare4.com QR** (operator: "For 'crushn8r' easter egg in response bubble talking about ron's
  past, add a QR to flare4.com"): a bubble with "nightlife" or "Mississauga" gets it (`qr_codes`).
- **Not eggs any more:** Weather and Get Photo. The weather face shows in the Brobot's own bubble;
  the photo's 3 pictures in a plain picture bubble.
- **Websites** (operator: "re: websites * use .ca QR"): the websites topic's picture is the
  crushn8r.ca QR (`pictures/crushn8r_ca_qr.png`), in chapter 9's egg.
- **Brobot 2 first, eyes swapped** (operator: "switch to [icon brobot eyes 2] [icon brobot eyes
  1], as a visual indicator"): the operator's "to Brobots" header shows Brobot 2's eyes first.

## Your stuff, no code (step 4, 2026-10-18)

Operator: "all I want left to be done, is my input, flexible data sets * audio/image/video,
scriptable edits, all well documented for easy access". Everything below is data - no code is
touched. After any of it, type **`system-check`** on the chat keyboard: it looks for every file
the settings, chapters and songs name, and every song step (Wire-Pod also runs it at start -
`GOPOD_CHECK` lines in the journal).

| To... | Do this | Then |
|---|---|---|
| **Change a setting** (volume, waits, texts) | `system-set <setting> <value>` on the chat keyboard, or edit `webroot/gopod_brobots_settings.json` (`system-locate <word>` finds it) | page: F5; Wire-Pod: `wpr 2` |
| **Add a picture to an egg** | put it in `webroot/pictures/`, add its name to the topic file's `pictures:` line (`gopod_probes/chapters/<chapter>/<topic>.txt`) | `wpr 2` |
| **Add a topic** (new egg words, value points) | a new `NN_name.txt` in its chapter folder: `label:`, `words:`, `pictures:` (optional), `reminder:`, `say 1:`, `say 2:` ..., then `---` and its reveal text | `wpr 2` |
| **Add a chapter** (a new egg) | a new folder `NN_name/` in `gopod_probes/chapters/` with `chapter.txt` (`title:` = the egg's title, `summary:`) and its topic files; the number sets its order | `wpr 2` |
| **Change a song line** | its `## STEP` / `> TEXT:` in the song's `story.md`, or `system-set <song> <step> text <words>` | next run |
| **Add a song step** | a `## STEP <id>` + `> TEXT:` in `story.md` **and** a step with the same `step_id` in `knobs.json` (`system-check` says if one is missing) | next run |
| **Add a sound** (sting, song sound, soundtrack) | put the `.wav` in `webroot/gopod_reporter_audio/`; name it in the settings (a sting) or a song's `knobs.json` | evened to the loudness rule by itself |
| **Change how a word is said** | the pronunciation registry in `brobots_wirepod_interview_section_card_template_1_001.md` (Brobots), `reporter_voices.say_words` (reporters), `system_say_words` (Goverlord) | `wpr 2` |
| **Add a QR code** | the picture in `webroot/qr/`, an entry in `qr_codes.codes` (words that show it) | F5 |
| **Video** | not supported yet in the chat | - |

## Golden rules

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
6. **Watch the battery: 3.7 V is the line.** Healthy robots read 3.99–4.08 V; Brobot 2's two
   crashes came at 3.61–3.62 V. phcal flags every move below 3.7 V (`PHCAL_BLOCKED_LOW_BATTERY`),
   and the chat logs the battery after every reply (`GOPOD_BATTERY`). 2026-10-16: for hours
   Brobot 2 read 3.61 V, level FULL, on the charger but not charging (`GOPOD_BATTERY_WATCH`);
   after a reboot he read 4.04 V and charging. A low reading on the charger that isn't charging
   (or one stuck at the same value) is a likely reason Brobot 2 freezes or reboots: check his
   charger contact, then reboot him. The watchdog stays quiet on the charger, so look at the log.
7. **Stop at a safe spot:** `stop run` ends a run before its next step, with control handed back.
8. **Values live in files, never in code:** a song's in its knobs, the chat's in
   `gopod_brobots_settings.json`.
9. **A locked thing gets a test:** the engine's animation lock, the chat's highlight, and this
   standard's own check (below). Nothing deploys if one fails.
10. **Chat bubbles appear whole, on time, and the chat scrolls once for them** (operator
    2026-10-09: "set a golden rule for chat bubbles/scrolling interactions"). A new bubble is
    built whole while invisible, taking its room; the chat scrolls once so it's in view; then it
    shows. A timed bubble (the nap's narrator) is sent early (`narrator_lead_seconds`) and shows
    at its first line's second, never drawing in on screen. Its highlight lights in place.
11. **A bubble is not its spoken words** (operator 2026-10-09: "a bubble is not the same as the
    expected contents of the bubble"; "Do not go forward if an expected sleep is not
    spoken"). A spoken line's bubble shows with its voice, not before it, and a step that waits
    on a line (`system_line` with `wait_spoken`) goes on only once the chat page reports the
    voice played. Not heard in time stops the run.
12. **One session, one run at a time, and the run owns the Brobots** (operator 2026-10-10:
    "Goverlord is the wpr-session orchestrator"). See [The session](#the-session-goverlords-stage).
13. **A song is found by its `song_id`** (operator 2026-10-10: "even shuffling (song) folders
    without notice should be easy to remap"). No folder name is written anywhere else.
14. **The baton hands over Olympic fashion** (operator 2026-10-10: "create a golden rule about
    creating the full bubble hidden, then handover the baton in olympic fashion"; then
    "elevate the narrator bubble rule globally to all chat bubbles"). Every chat bubble: the
    next one is built whole while hidden and waits in its lane, taking its room below with no
    scroll, while a bubble on screen is still being said or sung (its highlight lit). Only
    then the baton passes: the chat scrolls once, the next bubble shows (a timed one on its
    second), and it lights. Nothing cuts the runner ahead short, the operator's own line included
    (2026-10-11); a highlight left on holds the baton at most 15 s (`CHAT_BATON_MAX_MS`). The condition it fixes (live 2026-10-10, the nap): a narrator
    bubble arrives `narrator_lead_seconds` early, and its scroll and highlight took over
    while the bubble above was still being sung, so that bubble jumped up and its highlight
    stopped. `showNarratorLine`/`lightNarrator` in `gopod_brobots_chat.js`.

15. **After spoken text, release the highlight, then carry on** (operator 2026-10-10). When a
    line (or a combo part) has been said, Wire-Pod writes the robot as no longer speaking and
    the chat page takes the highlight off; only then does the next thing go, and a waiting
    bubble gets the baton (rule 14). `GopodSpeakSentences` ends every line with it; locked by
    `TestComboPartLightsItsPieceThenReleases`.

---

## Editing a song (the human's reference)

Every part of a song is in its two files, so any part, a line, a step, a whole section or the
whole song, can be rewritten from scratch by hand. No code changes.

**What can be changed, and where:**

| To change | Edit | In |
|---|---|---|
| The words of a line or caption | `> TEXT:` under its `## STEP` | `story.md` |
| What a step does | `note` (any note in [The notes](#the-notes-the-piano)) | `knobs.json` |
| Who plays it | `speaker` (`brobot_1`, `brobot_2`) | `knobs.json` |
| The order | the order of the steps (in both files) | both |
| When it plays | `start_seconds` (timed song), `pause_seconds`, `buffer_after`, `tempo_factor`, `global_tempo` | `knobs.json` |
| How a move plays | `hold_seconds`, `speed`, `cycles`, `settle_seconds` (tune it in phcal first) | `knobs.json` |
| An animation | `animation_token` (from [the toolbox](#the-toolbox-shared-golden-pieces)) | `knobs.json` |
| Which part it's in | `section` in `knobs.json`, and the `## SECTION:` it sits under in `story.md` | both |
| The whole song's feel | the top keys: `timeline`, `robots_speak`, `sleepy`, `audio`, `chat_intro`, ... | `knobs.json` |

**Recipes:**

- **New words:** change the `> TEXT:` line. That's all.
- **Add a step:** add a `## STEP <step_id>` with its `> TEXT:` in `story.md`, and a step with the
  same `step_id` in `knobs.json`, at the same place in both.
- **Remove or move a step:** remove or move it in both files.
- **Rewrite a section from scratch:** delete its steps in both files, write the new
  `## STEP`s under the same `## SECTION:` in `story.md`, then the same steps in `knobs.json`.
- **Rewrite the whole song:** the same, every section. The `song_id` stays, so every `start`
  command and alias still finds it.
- **Try it:** `zKnobs.json` (your private copy) wins while it's there. `score <song>` shows the
  steps; `pha0b <song> <first_step> <last_step>`, dry, plays a range with no robots; then live.
- **Lock it in:** a good live run, then the status line to `GOLDEN` with the date, then a commit.

Before every deploy `TestSongFoldersFollowTheStandard` checks the two files name the same steps
and every note is known. A mismatch is named, and nothing deploys.

**Parked, not built** (operator 2026-10-10: "Not sure. Keep as is. Too speculative."): one day
`story.md` might absorb `knobs.json`, so a sequence is written in one file and applied to a new
`zKnobs.json`. Today the two files stay as they are.

---

## Making a new song

1. Copy `goverlord/runtime/songs/_new_song/` to `<number>_<name>/`. Set its status to `WIP`.
2. Write the sections and steps in `story.md`, and the same steps in `knobs.json`.
3. Give it a `song_id` in `knobs.json`, add that `song_id` to the engine's song list
   (`SONG_REGISTRY` in `run_golden_song_001.py`), and give `pha0b` its keyword (`brobots.sh`,
   matched by `song_id`). The folder's own name is yours to change any time.
4. `score <song>` prints it. `pha0b`, dry, plays it with no robots.
5. Play it live, a range at a time. Tune moves in phcal.
6. When a full live run is right, set the status to `GOLDEN` with the date.

**An LLM warm song (2026-10-16):** copy `_new_song/_llm_warm/` instead. A step's `"volume"`: `"default"` (chat_audio.volume) or a percentage of it (`"50%"`, `"150%"`; never past full), on its voices and stings (operator 2026-10-16: "add volume, default or %age"). It is played by Wire-Pod, not by
`pha0b`; the opener plays the one named by `llm.opener_song_id`.

**The standard's check:** `TestSongFoldersFollowTheStandard` (Wire-Pod's test run, before every
deploy) checks every song folder: a status line, the same steps in both files, every note type
known.

---

GOPOD — [README](README.md) · [WIRED-POD](WIRED-POD.md) · [channel](CHANNEL.md) · [ask](ASK.md)

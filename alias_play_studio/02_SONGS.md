# Songs: the folder, the notes, making one

Part of [ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md) - [Concepts and terms](01_CONCEPTS.md) · [Songs](02_SONGS.md) · [The session and runs](03_SESSION.md) · [The Brobots chat](04_CHAT.md) · [Voices and sound](05_VOICES_AND_SOUND.md) · [The cast and the game](06_CAST_AND_GAME.md) · [The system menu, your stuff, the settings](07_SYSTEM.md) · [Golden rules, troubleshooting, tests](08_RULES_AND_TESTS.md) · [The gears](09_GEARS.md)

## 2. A song folder

### 2.1 The folder and its files

A song is a folder: `goverlord/runtime/songs/<number>_<name>/`. Copy
[`_new_song/`](../goverlord/runtime/songs/_new_song/) to start one. A folder whose name starts with
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
`**Status:** GOLDEN` or `**Status:** WIP`. GOLDEN means
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
| `goverlord` | Goverlord, the system voice (the LLM warm notes, `system`) |
| `narrator` | the nap's Narrator, its captions (`line`) |

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
| `soft_wake` | `wake_sleepy` | `brobots_soft_wake_0506` (nap) |
| `line` (who `narrator`) | a timed caption (`pause` with `start_seconds`) | `narrator_line_0608` (nap) |
| `system` (who `goverlord`) | `system_line` - a Goverlord line | `goverlord_system_0103` (nap) |
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
master list`. It only warns; the run plays its knobs as ever. Names without `_SSNN` are free
labels and aren't read (the nap's `line_30`, `soft_wake` were, until: renamed to the new
format, `narrator_line_0608`, `brobots_soft_wake_0506`). Locked by
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
| `thinking` | The live chat's own thinking process, run by Wire-Pod (`/api-sdk/gopod_thinking`, `gopodThinkingProcess`): arm, nod, the thinking line, searching (at least `searching_min_seconds`), the get-out. One process for live and scripted runs. While a manager has the floor (`gopodManagerBusy`, set by `gopodManagerDo`), the searching stops (one get-out) and starts again after. Locked by `TestThinkingProcessIsOne`. | | fart joke |
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
[2.2](#22-step-names) and [The opener](03_SESSION.md#63-wire-pod-start-and-the-opener).

### 3.4 The weather in a line

**GOLDEN.** `{weather}` in an interview line is the live weather. One weather call
(`gopodWeatherSay`) serves live chat and scripted runs alike: the real weather (Windsor fallback
location), else `live_weather.offline_text`. Locked by `TestWeatherPlace`.

**Each Brobot says his own units first, by design (not a repeat).** Each says his own unit and
clock first, then the other, from his own Wire-Pod settings (`temp_is_fahrenheit`,
`clock_24_hour`): Brobot 1 metric and 24-hour ("16 degrees Celsius or 62 degrees Fahrenheit ...
as of 10:35 or 10:35 AM"), Brobot 2 imperial and 12-hour ("61 degrees Fahrenheit or 16 degrees
Celsius ... as of 10:28 AM or 10:28").

For weather in the live chat (place names, countries, flags) see [8.6](04_CHAT.md#86-weather).

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
| `00_llm_warming` | WIP | The opener (`song_id` `brobots_llm_warming`): Wire-Pod plays it at every start. See [6.3](03_SESSION.md#63-wire-pod-start-and-the-opener). |
| `01a_brobots_bingo_test` | GOLDEN | Host and brat roles; the `OPENING SYNC` opener (wake, warm-up rattle, "Brobots ready" together, arm cue); three ball-draw rounds with the rattle, countdown and big balls; `angry`, `searching` and `kgSuccess` beats; 0-second reporter gaps; the LLM warm-up and "LLM Ready for Bingo"; touch interrupt off. |
| `02_brobots_interview` | WIP | Your score, played by the chat. Two reporters. One question per egg - the workbook, fractions and decimals, the rig, the rest of the cast, birthdays and merch, Bingo nights, foodmath, the weather, roast battles, our legacy - then the closer (`note: closer`, Brobot 2's "Press my backpack…") after "Interview run complete". The Brobots' answers are your own lines from the topic files, word for word; the reporters' questions are drafts. Chapters named in the song (`chapter`); the photos named with their questions (`pictures`); Brobot 1's Awaken bit (arm, nod and cube at their words, and the weather); the brothers' handoffs; emojis and cues lightly sprinkled. |
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
4. To play it from the Brobots chat, give it a run card in the settings' `game.runs` (the game
   engine, below) and its start / stop words as Wire-Pod voice commands (`customIntents.json`).
5. `score <song>` prints it. `pha0b`, dry, plays it with no robots.
6. Play it live, a range at a time. Tune moves in phcal.
7. When a full live run is right, set the status to `GOLDEN` with the date.

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

## More in this part

**A song that needs no language model**: no warm-up - no warm-up bubble, no wait, no ready-end line.
Goverlord says `llm.no_llm_text` and the song's `llm_ready_text` ("No Language Model Required /
We here, are Baby Brobots Ready"), then the song goes on (`song_llm_warmup`,
`run_golden_song_001.py`). A song that needs it ends its ready line with `llm.run_ready_end`, now
"Let's Begin. (music)".

**"(music)" plays djent where it's written** (`gopodMusicMarks`): at a line's start
the sting plays before the words, at its end right after them; the page loads grunge and djent3b
once and replays them at once (`preloadStings`). In the middle of a line: before the words, for now.

**A song's new story format: one move or one line per step** (songs in it: `01a_brobots_bingo_test`, `03_fart_joke`, `101_brobots_nap` ). Each move (arm cue, nod, kgSuccess,
searching, beat, rattle) and each line (`*_say_*`) is its own step, in the order they happen; a
line's name can repeat. The engine reads the story in order (`parse_control_story_list`) - the nth
step of a name in knobs.json takes the nth in story.md - and, with `"compose_moves": true` in
knobs.json, plays a move followed straight away by the same Brobot's line together, the native way
(`compose_move_lines`): the arm cue/nod while it's said, kgSuccess/a beat as
"{{playAnimationWI||token}} line", searching first, then the line; the line's step gives the gap
after it and any `show_ball`. Older songs (no `compose_moves`) play as before. Locked by
`NewStoryFormatTests`.

---

[Back to ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md)

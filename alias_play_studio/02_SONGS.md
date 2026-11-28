# Songs: the folder, the notes, making one

Part of [ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md) - [Concepts](concepts/README.md) · [Songs](02_SONGS.md) · [The session and runs](03_SESSION.md) · [The Brobots chat](04_CHAT.md) · [Voices and sound](05_VOICES_AND_SOUND.md) · [The cast and the game](06_CAST_AND_GAME.md) · [The system menu, your stuff, the settings](07_SYSTEM.md) · [Golden rules, troubleshooting, tests](08_RULES_AND_TESTS.md) · [The gears](09_GEARS.md)

<details><summary>Game card: making a song</summary>
<!-- gopod topic: 07_making_a_song -->

# making a song
- **chapter:** Show and tell

- **words:** make a song, making a song, song folder, knobs json, song engine
- **ask:** How do I make a song?
- **summary:** a song is a folder | each step names its note | a note is a piano key, proven live first
- **source:** alias_play_studio/02_SONGS.md
- **ask 1:** Brobot 1, tell me about making a song
- **ask 2:** Brobot 2, making a song - what goes in each step of knobs json?
- **ask 3:** Brobot 1, making a song, what is a note - a piano key?

---

Making a song. Reveal these, in order, in your own words:
1. A song is a folder: copy _new_song to start one; a folder whose name starts with _ is a template, never a song.
2. Each step in knobs.json names its note - what it plays - and who plays it, Brobot 1 or Brobot 2.
3. A note is what a step plays - the piano's keys - and nothing goes into a song untested.

</details>

## A song is a folder
- The songs are the folders in [`goverlord/runtime/songs/`](../goverlord/runtime/songs/). Copy
  `_new_song/` to start one; a folder starting with `_` is a template, never a song.
- `story.md` holds the words, `knobs.json` what each step plays and when. The two files always name
  the same steps.
- A song is known by the `song_id` at the top of its `knobs.json`, never its folder name: rename or
  reshuffle folders freely.
- The status line under the title: `**Status:** GOLDEN` (live-confirmed, ready to record) or
  `**Status:** WIP`.
- What every song must have: [`_new_song/song-schema.md`](../goverlord/runtime/songs/_new_song/song-schema.md).

## Where the words for each part live
- **What a step can play (the notes):** the known notes are listed in the standard's check,
  [`gopod_song_standard_test.go`](../goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_song_standard_test.go);
  what each one does, its fields and the top keys, in the song engine,
  [`run_golden_song_001.py`](../goverlord/runtime/songs/tools/run_golden_song_001.py). The template
  [`_new_song/knobs.json`](../goverlord/runtime/songs/_new_song/knobs.json) shows the shape.
- **Step names:** `<who>_<part>_<part>..._<SSNN>` - who, each part as it plays, section `SS`, step
  `NN`. The parts' master list is the name check,
  [`test_golden_song_step_names_001.py`](../goverlord/runtime/songs/tools/test_golden_song_step_names_001.py);
  every run warns (`STEP_NAME_WARNING`) on a name that doesn't read as its knobs. A step name is a
  label for people: the engine plays the knobs.
- **Animations:** `goverlord/wire_pod_overlay/chipper/animation_vocab.json`. **Sounds:**
  `goverlord/wire_pod_overlay/chipper/sounds/`.
- **Each song's own pieces:** its own folder.

## Editing a song
- **New words:** change the `> TEXT:` line in `story.md`, or `system-set <song> <step> text <words>`
  in the chat. A line right under `> TEXT:` (no blank line between) is the words' next line.
- **Add, remove or move a step:** in both files (and `zKnobs.json` when there is one), at the same
  place.
- **Rewrite a section or the whole song:** its steps in both files; the `song_id` stays, so every
  command still finds it.
- **Try it:** `zKnobs.json` (your private copy) wins while it's there. `score <song>` shows the
  steps; `pha0b`, dry, plays a range with no robots; then live. Tune moves in `phcal`.
- **Lock it in:** a good live run, then the status line to `GOLDEN`, then a commit.

## Making a new song
1. Copy `_new_song/` to `<number>_<name>/`, status `WIP`.
2. Write the steps in `story.md` and the same steps in `knobs.json`.
3. Give it a `song_id`, add it to the engine's song list (`SONG_REGISTRY` in
   `run_golden_song_001.py`) and to `pha0b` (`brobots.sh`).
4. To start it from the chat: a run card in the settings' `game.runs`, and its start / stop words
   in `customIntents.json`.
5. Play it dry, then live a range at a time; when a full live run is right, `GOLDEN`.

## An LLM warm song
- Copy `_new_song/_llm_warm/` instead; Wire-Pod plays it, not `pha0b`. At start Wire-Pod plays the
  one named by `llm.opener_song_id`.
- Each warm-up folder has a fixed-name `_llm_song/` run: the lines between "Voices ready" and "LLM
  ready". Swap the folder to change them; no steps = nothing said.
- A song that needs no voices or no language model: `"voices_required": false` /
  `"llm_required": false` in its `knobs.json`.

## Timing that works
1. Every line, sound and thinking step plays to its end; every gap between parts is cut
   (`buffer_after` 0, `tempo_factor` 0).
2. A move with a line holds 0 s; a hold belongs only to a move with no line.
3. A sound plays after its line, in full; an offset is written in (`with_next`, `start_after`).
4. Locked when it plays right: `songs_soft_lock.json`.

## The check
`TestSongFoldersFollowTheStandard`, before every deploy: every song folder has a status line, the
same steps in both files, and only known notes. A mismatch is named, and nothing deploys.

[Back to ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md)

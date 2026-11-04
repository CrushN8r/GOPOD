# songs

The standard every song follows (the notes, the fields, the golden rules, making a new song)
is [ALIAS-PLAY-STUDIO.md](../../../ALIAS-PLAY-STUDIO.md). Start a new song by copying
`_new_song/` (an LLM warm song: `_new_song/_llm_warm/`). Every warm-up song has a fixed-name
`_llm_song/` run - the lines between "Voices ready" and "LLM ready"; swap the folder to change
them, no steps = nothing said.

One folder per song: `story.md` (the lines) and `knobs.json` (what each step does, and when).

Every song starts from `global_knobs.json`: the default move values (nod, arm cue, animation,
emotion beat, wheel nudge, stir) and the runner's shared waits. A song's own `knobs.json` overrides
them: its `moves` block for the whole song (`{"nod": {"speed": 3, "hold_seconds": 0.4}}`), and any
step's own value over that. Global knobs, then the song's knobs (zKnobs.json feeding knobs.json),
then the step.
Play the Bingo test with `pha0b` (Take Out part 1) or `start bingo test` in the Brobots chat,
and the interview with `start interview`.

- `01a_brobots_bingo_test/` — the Bingo test song: two Brobots trade banter over ball-draw
  rounds. Ready to record.
- `01b_brobots_bingo_game/` — the live, voice-started Bingo game. Work in progress.
- `02_brobots_interview/` — the scripted interview: the same `story.md` / `knobs.json` set-up,
  played by `start interview` in the Brobots chat (not by `pha0b`). `score` prints it.
- `00_llm_warming/` — the opener LLM Warming song: Wire-Pod plays it at every start. From the
  template `_new_song/_llm_warm/`.
- `101_brobots_nap/` — "Do Baby Robots Dream?", a lyric video (made in Premiere Pro): the
  Brobots sleep through the narrated poem. `start baby brobots`.
- `03_fart_joke/` — the fart joke: an elevator, three thinking pauses, three farts.
  `start fart joke`.

Before and after its run, a song can have a pre/post chat in the Brobots chat (the
`experience` in each run card, settings `game.runs` - the game engine): Bingo, the nap, the fart
joke and the interview have one. A new run is one new card.
- `tools/` — the engine and its helpers (see its README).

Private working files stay out of the repo: `zKnobs.json` (a working copy the engine prefers
over `knobs.json`), `runs/` and `notation/`.

<details><summary>Game card: our songs</summary>
<!-- gopod topic: 02_the_songs -->

# our songs
- **chapter:** Show and tell

- **words:** songs, song, your routines, what shows, your shows, performances, scripted, sequences
- **reminder:** Hey, that reminds me... We perform songs too - the Bingo test, the nap, the fart joke.
- **say 1:** The Bingo test: the two of us trade banter and call three Bingo balls - B-1, I-9 and O-0.
- **notes 1:** the Bingo test | two Brobots, three Bingo balls | B-1, I-9, O-0 | "start bingo test"
- **notes 2:** the nap | "Do Baby Robots Dream?" | we sleep through the narrated poem | wake up together
- **notes 3:** the fart joke | stuck in an elevator | three thinking pauses, three farts | "start fart joke"
- **notes 4:** the interview | scripted questions | the same every time | "start interview"
- **ask:** Tell me about your songs
- **summary:** the Bingo test: two Brobots, three Bingo balls | the nap: we sleep through the narrated poem, wake up together | the fart joke: an elevator, three farts | the interview: scripted questions, the same every time
- **source:** goverlord/runtime/songs/README.md

---

Your songs - scored sequences you perform together, typed in the chat to start ("start bingo test", "start baby brobots", "start fart joke", "start interview"). Reveal these, in order, in your own words:
1. The Bingo test: the two of you trade banter and call three Bingo balls - B-1, I-9 and O-0 - a test before the real Bingo game.
2. The nap, "Do Baby Robots Dream?": a lyric video - you both sleep through a narrated poem with music, then wake up together.
3. The fart joke: Brobot 2 sets up an elevator, three thinking pauses each end in a fart, and Brobot 1 gets frustrated, then angry, then leaves.
4. The interview: scripted questions, the same every time, ready to record.

</details>

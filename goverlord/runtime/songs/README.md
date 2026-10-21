# songs

The standard every song follows (the notes, the fields, the golden rules, making a new song)
is [ALIAS-PLAY-STUDIO.md](../../../ALIAS-PLAY-STUDIO.md). Start a new song by copying
`_new_song/` (an LLM warm song: `_new_song/_llm_warm/`). Every warm-up song has a fixed-name
`_llm_song/` run - the lines between "Voices ready" and "LLM ready"; swap the folder to change
them, no steps = nothing said.

One folder per song: `story.md` (the lines) and `knobs.json` (what each step does, and when).
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
- `tools/` — the engine and its helpers (see its README).

Private working files stay out of the repo: `zKnobs.json` (a working copy the engine prefers
over `knobs.json`), `runs/` and `notation/`.

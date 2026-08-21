# song-tools

Everything in this folder is the machinery behind GOPOD's songs — the scripts that
actually make the two Brobots move, talk, and perform. None of these are picked from a
menu; each one has its own real command. That's on purpose, not an oversight — see
below.

## The tools you can actually run

| Tool | What it does | Run it with |
|---|---|---|
| `run_golden_song_001.py` | The engine. Plays almost every scored song start to finish on the real robots (or dry, with no hardware at all) — the Bingo warm-up, the cross-persona mix-up bit, baby-robots-sleep, "is that you?", and the awaken bait video. | `pha0b` — pick any real song from the menu |
| `run_section1_full_live_001.py` | The flagship: the full live two-robot AI interview, start to finish, real thinking and all. | `interview-run`/`interview-replay`/`pha0b interview` |
| `run_robot_control_song_001.py` | A one-robot "talking self-check" — arm test, head nod, the fireworks finale, and the real live weather fetch-and-speak. | reached through `pha0b`'s control-song entries, and used behind the scenes by `phcal`'s own weather/animation/brobots-1-2 controls |
| `run_vamp_gate_song_001.py` | The pre-show's short filler chatter (4 lines), runnable on its own for testing/watching just that bit. | `pha0b` — the "vamp" entry |
| `run_interview_movement_rehearsal_001.py` | Fires every physical movement the interview's real script calls for, in the real order, on real hardware — using placeholder speech instead of a real AI-generated interview, so the choreography can be checked without waiting on the AI. | `test-interview-movements` |
| `run_section1_preshow_generate_001.py` | Runs just the "AI writes the interview script" half and stops — no robot speaks yet. The middle stage of the 3-stage interview show. | `interview-json` |
| `print_song_score_001.py` | Prints a song's full "sheet music" — every scripted step, in order, human-readable — to the screen or to a saved text file. | `score` |

## The two that AREN'T run directly

| File | What it is |
|---|---|
| `knobs_envelope_001.py` | Shared library — figures out which tuning file (`knobs.json` or a working `zKnobs.json`) is actually active for a song. Called by the tools above, not run on its own. |
| `run_section1_full_live_001.py` | Yes, this one's in both tables — it's the runnable interview engine above, AND several of the other tools quietly borrow pieces of it (network/robot helper functions, warm-up steps) instead of re-writing their own. |

## Why nothing here is picked from a menu

`pha0b`'s own song list scans this folder's parent directory — this folder used to show
up in that scan too, as a fake "song" that always blocked when picked (there's no
`knobs.json` in here, because this isn't a song). As of 2026-08-16, `pha0b` skips this
folder entirely, so its menu only ever shows real, pickable songs. That's not a
limitation — every tool above already has its own real command, listed in the table.

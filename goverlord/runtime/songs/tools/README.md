# song tools

| File | What it does |
|---|---|
| `run_golden_song_001.py` | The engine: plays a song live or dry. `pha0b` and `start bingo test` run it. |
| `print_song_score_001.py` | Prints a song's score, every step in order. `score` runs it. |
| `run_robot_control_song_001.py` | The arm, nod and live-weather moves `phcal` uses. |
| `knobs_envelope_001.py` | Picks a song's active knobs file. |
| `song_dialogue_normalizer_001.py` | Reads a song's lines from `story.md` for the score. |
| `run_section1_full_live_001.py`, `run_section1_preshow_generate_001.py`, `speech_math_flatten_001.py` | Shared helpers (robot calls, speech clean-up, pronunciation) the tools above borrow. They began as the first interview's runner. |
| `test_golden_song_animation_lock_001.py` | Tests for the engine. |

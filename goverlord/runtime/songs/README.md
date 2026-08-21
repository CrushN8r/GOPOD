# songs/

Every GOPOD song lives here, one folder each, plus the shared runners that
play them. See `tech/alias_play_studio/ALIAS-LIBRARY.md` for the live
alias/cockpit registry and `tech/alias_play_studio/GOPOD_SONGS.md` for the
per-song story docs — this file is the on-disk orientation, not a duplicate
of either.

## Songs

- `00_brobots_awaken/` — a robot wakes up, checks its arms and head, tells
  you the weather, and says it's ready.
- `01_brobots_interview_vamp/` — the flagship interview's video 1: the
  backstage/preshow banter, fireable standalone with zero interview
  generation (`interview-vamp-play`), or as the generation-gated vamp (`interview-vamp`).
- `02_brobots_interview_run/` — the flagship interview's video 2: the
  seven-exchange scripted interview itself, fireable standalone
  (`interview-replay`/`pha0b interview`; `interview-run` adds an optional
  full-run mode that plays the vamp first). Holds the currently-remembered LLM
  model-picker state (`content_model_state.json`) and `zmisc/`, the
  interview's own Wire-Pod runtime scaffold
  (`brobots_interview_runtime_scaffold_001.json` +
  `brobots_wirepod_interview_section_card_template_1_001.md`, the
  human-editable Template 1 master) and its test suite. Split from one
  combined `01_brobots_interview_section_01/` folder 2026-08-19 — see
  `gopod_notes/INTERVIEW_VAMP_SPLIT_001.md`.
- `101_brobots_bingo_test/` — the scored song: two robots trade banter over
  ball-draw rounds. Locked.
- `102_brobots_bingo_game/` — the real, voice-triggered Bingo game. Bingo's
  own Go sidecar (`bin/`, `cmd/`, `pkg/`) and the standalone Brobot 2
  angry-animation reactor (`bingo_reactor/`) both live inside this song's own
  folder — a self-contained build source, binary, and companion script,
  not scattered across the engine. See this folder's own `README.md` for
  source-of-truth and build details, and `bingo_reactor/README.md` for the
  reactor.
- `103_gopod_is_that_you_single/` / `104_gopod_is_that_you_multi/` — the live
  push-to-talk cross-persona mix-up demo, split 2026-08-18 into a single-robot
  (KP1/Doc-only) scope and the original two-robot version.
- `105_brobots_nap/` — a quiet, story-shaped piece predating GOPOD itself.
  Renamed/renumbered 2026-08-18 from `104_brobots_baby_robots_sleep/`.
- `zzz_archives/` — retired songs, kept for history, not reachable via any
  live alias.

## tools/

The interview runner (`run_section1_full_live_001.py`) and every other song
runner (`run_golden_song_001.py`, `run_robot_control_song_001.py`,
`run_vamp_gate_song_001.py`, `print_song_score_001.py`,
`knobs_envelope_001.py`, `run_interview_movement_rehearsal_001.py`,
`run_section1_preshow_generate_001.py`) — song-specific tooling, hence living
under `songs/` rather than `data_gomad/`.

## Per-song convention

Each song folder carries `story.md` (the score) and `knobs.json` (the
public "latest updated" snapshot); a `zKnobs.json` sibling, when present, is
a private working copy every runner prefers over `knobs.json` — gitignored,
not public showcase material. `runs/` and `notation/`, when present, are the
same class of private working output — rehearsal logs and marked-up score
snapshots, gitignored.

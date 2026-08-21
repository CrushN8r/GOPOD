# runtime/

Three top-level pieces. `songs/` is the show content — every song folder plus
the shared runners that play them. `data_gomad/` is the engine underneath —
robot-facing modules, config, downloaded models. `gopod_layer/` is the
hinge/display layer — GOPOD Yourself and the cockpit. Restructured 2026-08-15
for public presentation (prior flat layout of 12 scattered top-level folders
collapsed, then reshuffled again the same day into this shape — songs
promoted to top level, Bingo's sidecar folded into its own song folder). See
`songs/README.md` and `data_gomad/README.md` for their own contents in
detail.

## songs/ — the show

- `00_brobots_awaken/`, `01_brobots_interview_vamp/` (video 1, the pre-show
  banter — split from the interview 2026-08-19, standalone-fireable),
  `02_brobots_interview_run/` (video 2, the seven-exchange interview itself,
  also standalone-fireable, `zmisc/` for the interview's own wire-pod
  scaffold/test suite), `101_brobots_bingo_test/` (the scored song), `102_brobots_bingo_game/`
  (the real, voice-triggered game — Bingo's own Go sidecar, its built binary,
  and the Brobot 2 reactor script all live inside this song's own folder now,
  not scattered across the engine), `103_gopod_is_that_you_single/`,
  `104_gopod_is_that_you_multi/`, `105_brobots_nap/`, plus `zzz_archives/` for
  retired ones.
- `tools/` — the interview runner and every song runner
  (`run_section1_full_live_001.py` and siblings) — song-specific runners, not
  generic utilities, hence living under `songs/` rather than the engine.

## data_gomad/ — the engine

- `configs/` — endpoints/paths/audio config, the chat-envelope schema.
- `models/` — downloaded model binaries (Vosk STT, Coral Edge TPU).
- `robot/` — `kokoro_voice/` (local TTS/audio verification tooling),
  `say_replacements/` (pre-speech text replacement, code + its own
  `replacement_rules_001.json`), `weather/` (`gopod_weather_fetch_001.py`,
  the per-robot weather-report data path).

## gopod_layer/ — the hinge/display layer

- `gopod_yourself/` — dated evidence/reference artifacts (stitched live-capture
  logs, expression-feel lessons, the clean master expression template) used to
  ground Doc/Pip/CHALK expression tuning in observed runtime behavior rather
  than assumption. Historical record, not live code.
- `web_display/` — the live cockpit/display server. See
  `web_display/README.md`.

## Note

`kokoro_voice/` does not yet have its own README file. Its contents are
described above from what's actually on disk, not from a separate doc —
check the files directly for current behavior.

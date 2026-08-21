# data_gomad/

**Moved here from `goverlord/data/` in Stage Pinnacle (2026-07-04)**, then from
`goverlord/gomads/data_gomad/` on 2026-07-30 when the `gomads/` tree was
removed entirely — `data_gomad` was its only gomad with a live caller, so it moved to
`runtime/` rather than being retired with the rest.

This directory holds facts: endpoints, filesystem paths, audio constants,
and schema definitions that the GOPOD codebase reads instead of
hardcoding. Data owns facts; it does not own flow decisions (see
`rules/README.md`) and does not contain executable logic, per the
three-way code/data/rules decoupling doctrine
(`GOPOD_LAYER_1_DECOUPLING_DOCTRINE_001.md`, formalized further in
`GOPOD_PERFECT_CRYSTAL_PROPOSAL_001.md`).

Every live consumer imports this module as `goverlord.runtime.data_gomad.configs.loader`
(previously `goverlord.gomads.data_gomad.config.loader`, and before that
`goverlord.data.config.loader`) — grep for that dotted path if relocating
this directory again.

## `models/` — gitignored, not GOPOD-authored

`models/` (Vosk STT model, Coral Edge TPU object-detection models) is
entirely gitignored (`.gitignore`: `goverlord/runtime/data_gomad/models/`).
Every file in it is a downloaded model artifact — no GOPOD-authored file
currently lives there. Same rule as `goverlord/SDK/` (see its README): if
GOPOD wrote it, track it; if it's fetched/regenerable third-party content,
gitignore it rather than documenting it file-by-file. `configs/paths.json`'s
`vosk_model_path` is the only in-repo reference to where this directory's
contents are expected to sit on disk.

## Contents, as of the 2026-08-15 restructure

- `configs/` — `endpoints.json`/`paths.json`/`audio.json`/`chat_envelope_schema.json`
  (schema merged in here from a separate `schema/` folder) plus `loader.py`.
- `models/` — see above.
- `robot/` — `kokoro_voice/` (local TTS/audio verification), `say_replacements/`
  (pre-speech text replacement, code + its own `replacement_rules_001.json`
  co-located), `weather/` (per-robot weather-report data path).

`songs/`, `tools/`, and the interview's own wire-pod scaffold/tests moved out
to live under `goverlord/runtime/songs/` instead (songs are the show content,
not engine data) — see `goverlord/runtime/songs/README.md`. The Bingo sidecar
and its Brobot 2 reactor moved into `goverlord/runtime/songs/102_brobots_bingo_game/`
for the same reason — a song's own build source, binary, and companion script
now live together in its own folder.

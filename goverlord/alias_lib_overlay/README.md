# Alias Lib Overlay — backup, not live source

The studio's standard, notes and golden rules: [ALIAS-PLAY-STUDIO.md](../../ALIAS-PLAY-STUDIO.md).

**This folder is a backup and showcase. It is never sourced live.** The live shell source
is `~/.gopod_alias_lib/`, loaded from `~/.bash_aliases` (`brobots.sh`) - the one load place. This folder mirrors it for backup and for anyone taking
the alias play studio home (WIRED-POD.md, "Take Out", part 1).

## What's here

- **`brobots.sh`** — the cockpit:
  - `pha0b` (PLAYHEAD: pick a song, play a range of it, live or dry), with its menu and prompts;
  - `phcal` (tune one arm, nod, cube light or animation on one robot);
  - `score` / `score-save` (print a song's score), and `tempo-set`;
  - `gobingo` (the live Bingo game), `gohockey` (VR/AR air hockey);
  - `gofps`, `gopower`, `goinit` (the Brobots' frame rate, power profile, a session's start);
  - the speakers and USB mic, set by `wpr` (`gopod_audio_config.sh`, private; see the example).
  - The songs it knows are `bingo` and `nap`.
- **Python tools** — `phcal_isolate_001.py`, `phcal_apply_001.py`, `robot_pick_001.py`,
  `tempo_set_001.py`, and the example configs. The chat page's numpad keys (who each one
  talks to) are in the repo now: `gopod_probes/topic_clusters/personas/company/99_system/numpad-mapper.md`.
- **`wpr` / `wps` / `wpq`** (restart and check Wire-Pod) live in `~/.bash_aliases`, not here.

**Retired:**
- `llm.sh` (`llm-*`) and `wirepod_logs.sh` (`wirepod*logs`), 2026-11-27 - never typed, no caller; the startup audio check (`wpr` sets the routing);
- `core.sh` (the opening chord), `demo.sh` (the PTT demos), `chat_capture.sh`;
- `openwebui.sh`, `goverlord.sh` and `suits.sh` (the Open WebUI / suits lane);
- `frame0.sh`, `tools.sh`, `gopod_json_capture.py`, `print_numpad_map_001.py`,
  `phcal_apply_control_song_001.py`;
- from `brobots.sh`: the interview, vamp, Awaken, Is That You and cross-persona entries,
  along with every helper only they used.

## Not mirrored, by design

- `local_config.sh` — this machine's private addresses. Wire-Pod's voice-command scripts
  source it too.
- `phcal_config.json`, `phcal_last.json` — this machine's own values and last tuning.
- Run logs and `__pycache__/`.

## Adapt before running

The files carry this machine's `/home/goverlord/...` paths; change them for yours.

## Updating this mirror

There is no sync script. Copy from `~/.gopod_alias_lib/` by hand, on request, and check with
`cmp`.

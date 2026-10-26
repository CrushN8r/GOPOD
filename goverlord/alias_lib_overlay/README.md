# Alias Lib Overlay — backup, not live source

The studio's standard, notes and golden rules: [ALIAS-PLAY-STUDIO.md](../../ALIAS-PLAY-STUDIO.md).

**This folder is a backup and showcase. It is never sourced live.** The live shell source
is `~/.gopod_alias_lib/`, loaded from `~/.bash_aliases` (`brobots.sh`, `llm.sh`) and
`~/.bashrc` (`wirepod_logs.sh`). This folder mirrors it for backup and for anyone taking
the alias play studio home (WIRED-POD.md, "Take Out", part 1).

## Cleaned out 2026-10-08

The operator kept three groups: the Bingo test plus phcal, the Wire-Pod basics, and the LLM
aliases. Everything else was retired. The old files are kept outside the repo, in
`gopod_notes/older_archives/alias_lib_retired_20261008/`, and the whole earlier studio is in
`alias_studio.zip`.

- **`brobots.sh`** — the cockpit:
  - `pha0b` (PLAYHEAD: pick a song, play a range of it, live or dry), with its menu and prompts;
  - `phcal` (tune one arm, nod, cube light or animation on one robot);
  - `score` / `score-save` (print a song's score), and `tempo-set`;
  - `gobingo` (the live Bingo game, moved here from `demo.sh`);
  - the audio-routing check.
  - It went from 226 functions to 16. The songs it knows are `bingo` and `nap`.
- **`llm.sh`** — `llm-status`, `llm-models`, `llm-test-*`.
- **`wirepod_logs.sh`** — `wirepodlogs`, `wirepoddebuglogs`, `wirepodlogsave`.
- **Python tools** — `phcal_isolate_001.py`, `phcal_apply_001.py`, `robot_pick_001.py`,
  `tempo_set_001.py`, plus `numpad_persona_map_001.json` (the chat page's numpad reads it;
  synced 2026-10-26 - KP1-KP9 live, KP3-KP6 and KP9 the new voice personas) and the example
  configs.
- **`wpr` / `wps` / `wpq`** (restart and check Wire-Pod) live in `~/.bash_aliases`, not here.

**Retired:**
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

# Alias library (copy)

A copy of GOPOD's terminal alias library. The live one is `~/.gopod_alias_lib/`, loaded from
`~/.bash_aliases`. Each alias has its card in [alias_play_studio/](../../alias_play_studio/).

## Files

- `brobots.sh` - every GOPOD terminal alias.
- `phcal_isolate_001.py`, `phcal_apply_001.py` - phcal's tools.
- `robot_pick_001.py`, `tempo_set_001.py` - the robot picker and tempo-set's tool.
- `phcal_config.example.json`, `gopod_audio_config.example.sh` - copy each without `.example`
  into `~/.gopod_alias_lib/` and put in your own values.

## Not copied (private)

- `local_config.sh` - this machine's addresses.
- `gopod_audio_config.sh` - this machine's speaker and mic names.
- `phcal_config.json`, `phcal_last.json` - this machine's values and last tuning.

## Take it home

Copy this folder to `~/.gopod_alias_lib/`, change the `/home/goverlord/...` paths to yours, and
load `brobots.sh` from `~/.bash_aliases`.

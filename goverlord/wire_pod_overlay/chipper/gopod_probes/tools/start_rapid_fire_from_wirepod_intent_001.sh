#!/bin/bash
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/start_rapid_fire_from_wirepod_intent_001.sh
# Wire-Pod custom intent START_RAPID_FIRE ("start rapid fire"): a live run of the rapid
# fire test (goverlord/runtime/songs/04_rapid_fire) on the golden song engine. "stop run" in
# the chat stops it at its next safe spot.
# Same shape as start_bingo_test_from_wirepod_intent_001.sh:
# - Wire-Pod runs intent scripts as root; the run is handed to the goverlord user;
# - one run at a time: the engine checks in with Goverlord (Wire-Pod's gopod_session.go);
#   a start while another run is going is refused there, with a Goverlord line.
# Output: /tmp/brobots_rapid_fire_live.log
if [ "$(id -u)" = "0" ]; then
    exec runuser -u goverlord -- "$0" "$@"
fi
source /home/goverlord/.gopod_alias_lib/local_config.sh
export GOPOD_WIREPOD_BASE_URL
export GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1
# The spoken "start rapid fire" is the run's opening chat bubble.
export GOPOD_CHAT_INTRO_SPOKEN=1
# The song by its song_id (knobs.json), wherever its folder is (song_folders_001.py).
export GOPOD_GOLDEN_SONG_ID=brobots_rapid_fire
cd /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools
nohup sh -c 'python3 run_golden_song_001.py < /dev/null > /tmp/brobots_rapid_fire_live.log 2>&1' > /dev/null 2>&1 &

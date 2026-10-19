#!/bin/bash
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/start_bingo_test_from_wirepod_intent_001.sh
# Wire-Pod custom intent START_BINGO_TEST ("start bingo test"): a live run of the Bingo
# test song (goverlord/runtime/songs/01a_brobots_bingo_test) on the golden song engine.
# "stop run" in the Brobots chat stops it at its next safe spot.
# - Wire-Pod runs intent scripts as root; the run is handed to the goverlord user so its
#   run log and chat state files stay the operator's own;
# - one run at a time: the engine checks in with Goverlord (Wire-Pod's gopod_session.go);
#   a start while another run is going is refused there, with a Goverlord line.
# Output: /tmp/brobots_bingo_test_live.log
if [ "$(id -u)" = "0" ]; then
    exec runuser -u goverlord -- "$0" "$@"
fi
source /home/goverlord/.gopod_alias_lib/local_config.sh
export GOPOD_WIREPOD_BASE_URL
export GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1
# The spoken "start bingo test" is the run's opening chat bubble, so the song doesn't
# add its own too.
export GOPOD_CHAT_INTRO_SPOKEN=1
# The song by its song_id (knobs.json), wherever its folder is (song_folders_001.py).
export GOPOD_GOLDEN_SONG_ID=brobots_bingo
cd /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools
nohup sh -c 'python3 run_golden_song_001.py < /dev/null > /tmp/brobots_bingo_test_live.log 2>&1' > /dev/null 2>&1 &

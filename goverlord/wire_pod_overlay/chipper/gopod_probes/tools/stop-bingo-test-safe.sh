#!/bin/bash
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/stop-bingo-test-safe.sh
# Stops a song run at the next safe spot. Run by Wire-Pod when "stop run" (or a "thanks
# Brobots") is typed or said on the numpad while a song is running (song_stop.safe_script,
# gopod_thanks_close_out.go). The kill-bingo-test alias still stops at once
# (kill-bingo-test.sh).
# 1. creates the song's stop request (song_stop.request_file): the song engine
#    ends the run before its next step - control handed back, touch interrupt on;
# 2. if the run hasn't ended after song_stop.wait_seconds, force-stops it with
#    kill-bingo-test.sh.
# Runs in the background so Wire-Pod returns at once.
# Output: /tmp/stop_bingo_test_safe.log
if [ "$(id -u)" = "0" ]; then
    exec runuser -u goverlord -- "$0" "$@"
fi
chipper=/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper
settings=$chipper/webroot/gopod_settings
# The settings folder, read as one (gopod_settings_folder.go).
read -r request wait_s < <(python3 -c "import json,glob,sys; m={}; [m.update(json.load(open(f))) for f in sorted(glob.glob(sys.argv[1]+'/*.json'))]; s=m['song_stop']; print(s['request_file'], s['wait_seconds'])" "$settings")
(
    echo "stop-bingo-test-safe $(date '+%F %T')"
    if ! pgrep -f run_golden_song_001.py > /dev/null; then
        echo "no song running"
        exit 0
    fi
    touch "$chipper/$request" && echo "stop requested - the song ends before its next step"
    for _ in $(seq "$wait_s"); do
        pgrep -f run_golden_song_001.py > /dev/null || { echo "song stopped at a safe spot $(date '+%T')"; exit 0; }
        sleep 1
    done
    echo "still running after ${wait_s}s - force stop"
    rm -f "$chipper/$request"
    "$chipper/gopod_probes/tools/kill-bingo-test.sh"
) > /tmp/stop_bingo_test_safe.log 2>&1 &

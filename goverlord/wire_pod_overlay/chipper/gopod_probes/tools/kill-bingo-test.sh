#!/bin/bash
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/kill-bingo-test.sh
# Force-stops a Bingo test run (bingo-test-live or "bingo test live") - operator
# 2026-10-01: "kill bingo test". One script, two doors: the kill-bingo-test alias
# (brobots.sh) and the Wire-Pod custom intent KILL_BINGO_TEST ("kill bingo test",
# said to the Brobot that isn't talking).
# 1. stops the song engine;
# 2. turns touch interrupt back on (deletes the song's off file - a killed run
#    never gets to delete it itself);
# 3. hands both Brobots back to Wire-Pod (release behavior control), so neither
#    stays frozen under the song's hold.
# Output: /tmp/kill_bingo_test.log
if [ "$(id -u)" = "0" ]; then
    exec runuser -u goverlord -- "$0" "$@"
fi
source /home/goverlord/.gopod_alias_lib/local_config.sh
chipper=/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper
settings=/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/webroot/gopod_brobots_settings.json
{
    echo "kill-bingo-test $(date '+%F %T')"
    pkill -f run_golden_song_001.py && echo "song stopped" || echo "no song running"
    off_file=$(python3 -c "import json,sys; print(json.load(open(sys.argv[1]))['touch_interrupt']['off_file'])" "$settings")
    rm -f "$chipper/$off_file" && echo "touch interrupt on"
    for serial in "${GOPOD_BROBOT_1_SERIAL:-0dd1b9e9}" "${GOPOD_BROBOT_2_SERIAL:-0dd1d8bf}"; do
        curl -fsS --max-time 10 -X POST --data-urlencode "serial=$serial" \
            "$GOPOD_WIREPOD_BASE_URL/api-sdk/release_behavior_control" > /dev/null 2>&1 \
            && echo "released $serial" || echo "release failed $serial"
    done
} 2>&1 | tee /tmp/kill_bingo_test.log

#!/bin/bash
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/system_play_001.sh
# system-play / system-calibrate from the Brobots chat (2026-10-19, operator:
# "What else from the survey, can be absorbed over from PLAYHEAD aliases to new
# menu?"; "Build it all in please"): pha0b's and phcal's golden guts, started by
# Wire-Pod (gopod_system_play.go) with everything already chosen - no prompts.
#   play:      GOPOD_GOLDEN_SONG_DIR, GOPOD_GOLDEN_PLAYHEAD_FROM/TO (step ids),
#              GOPOD_GOLDEN_ROBOT_FILTER (both|1|2), GOPOD_GOLDEN_REPORTER_GAP_OVERRIDE,
#              GOPOD_PLAY_LIVE=1 (robots move and speak; else dry), GOPOD_PLAY_TWEAKS=1
#              (phcal's last values applied to the range's moves first, as pha0b does)
#   calibrate: GOPOD_CALIBRATE_ARGS (phcal_isolate_001.py's: <move> <robot> [values])
# Wire-Pod runs this as root; the run is handed to the goverlord user, like the
# start_*_from_wirepod_intent scripts. Output: /tmp/gopod_system_play.log
if [ "$(id -u)" = "0" ]; then
    exec runuser -u goverlord -- "$0" "$@"
fi
source /home/goverlord/.gopod_alias_lib/local_config.sh
export GOPOD_WIREPOD_BASE_URL
LOG=/tmp/gopod_system_play.log
TOOLS=/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools
if [ "$GOPOD_PLAY_LIVE" = "1" ]; then export GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1; else unset GOPOD_ALLOW_LIVE_ROBOT_SPEECH; fi
if [ "$1" = "calibrate" ]; then
    cd /home/goverlord/.gopod_alias_lib || exit 1
    # shellcheck disable=SC2086
    nohup python3 phcal_isolate_001.py $GOPOD_CALIBRATE_ARGS < /dev/null > "$LOG" 2>&1 &
    exit 0
fi
if [ "$GOPOD_PLAY_TWEAKS" = "1" ] && [ "$GOPOD_PLAY_LIVE" = "1" ]; then
    python3 - "$GOPOD_GOLDEN_SONG_DIR" "$GOPOD_GOLDEN_PLAYHEAD_FROM" "$GOPOD_GOLDEN_PLAYHEAD_TO" <<'PY' | while IFS= read -r sid; do
import sys
sys.path.insert(0, "/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools")
from knobs_envelope_001 import load_knobs_envelope
_, d = load_knobs_envelope(sys.argv[1])
steps = d.get("steps", [])
ids = [s["step_id"] for s in steps]
a = ids.index(sys.argv[2]) if sys.argv[2] in ids else 0
b = ids.index(sys.argv[3]) if sys.argv[3] in ids else len(ids) - 1
for s in steps[min(a, b):max(a, b) + 1]:
    if s.get("note") in ("arm_cue", "nod", "head_nod", "animation", "wheel_nudge", "brobots_ready_together"):
        print(s["step_id"])
PY
        [ -n "$sid" ] && python3 /home/goverlord/.gopod_alias_lib/phcal_apply_001.py "$sid" --yes --knobs "$GOPOD_GOLDEN_SONG_DIR/knobs.json" >> "$LOG.tweaks" 2>&1
    done
fi
export GOPOD_GOLDEN_PLAYHEAD_WAKE=1
cd "$TOOLS" || exit 1
nohup sh -c 'python3 run_golden_song_001.py < /dev/null > /tmp/gopod_system_play.log 2>&1' > /dev/null 2>&1 &

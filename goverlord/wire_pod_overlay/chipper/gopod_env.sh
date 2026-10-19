# goverlord/wire_pod_overlay/chipper/gopod_env.sh
# GOPOD's settings for Wire-Pod's own start.sh, which stays native (2026-10-02,
# operator: "keep ~/wire-pod as native as possible"). Wire-Pod's local config file,
# chipper/source.sh (written by Wire-Pod's setup, never tracked), sources this file
# with one line:
#   source /home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_env.sh
# Every GOPOD file path comes from the build itself (pkg/vars/gopod_overlay_root.go),
# so only two things are needed here:
# - where this machine's libvosk is: native start.sh only knows upstream's
#   /root/.vosk/libvosk, which this machine doesn't have, and adds it in front of
#   LD_LIBRARY_PATH - the path set here stays on the list behind it;
# - the live speech gate, on by default (a test safety net, not a production block).
export LD_LIBRARY_PATH="/home/goverlord/.local/lib/python3.10/site-packages/vosk${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
export GOPOD_ALLOW_LIVE_ROBOT_SPEECH="${GOPOD_ALLOW_LIVE_ROBOT_SPEECH:-1}"

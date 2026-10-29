#!/bin/bash
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/stop_run_from_wirepod_intent_001.sh
# Wire-Pod custom intent STOP_RUN ("stop run"): hands the spoken words to the Brobots chat
# as a typed line, so a spoken "stop run" does what a typed one does - stops a playing
# draft or a running song at its next safe spot (gopod_thanks_close_out.go). Without it,
# Vector hears "stop" and runs his own quiet command (live 2026-10-09 13:44).
# Arg 1: the robot's serial (!botSerial). Output: /tmp/stop_run_intent.log
source /home/goverlord/.gopod_alias_lib/local_config.sh
curl -s -X POST "$GOPOD_WIREPOD_BASE_URL/api-sdk/gopod_typed_input" \
    --data-urlencode "voice=1" --data-urlencode "serial=${1:-ESN01}" --data-urlencode "text=stop run" \
    > /tmp/stop_run_intent.log 2>&1 &

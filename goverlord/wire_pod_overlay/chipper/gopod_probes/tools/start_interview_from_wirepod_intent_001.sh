#!/bin/bash
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/start_interview_from_wirepod_intent_001.sh
# Wire-Pod custom intent START_INTERVIEW ("start interview"): hands the spoken words to the
# Brobots chat as a typed line from the robot that heard them, so a spoken "start interview"
# does what a typed one does - plays the scripted interview (interview_script.play_words,
# gopod_interview_script.go), or does nothing before LLM Ready. "stop run" stops it.
# Arg 1: the robot's serial (!botSerial). Output: /tmp/start_interview_intent.log
source /home/goverlord/.gopod_alias_lib/local_config.sh
curl -s -X POST "$GOPOD_WIREPOD_BASE_URL/api-sdk/gopod_typed_input" \
    --data-urlencode "voice=1" --data-urlencode "serial=${1:-0dd1b9e9}" --data-urlencode "text=start interview" \
    > /tmp/start_interview_intent.log 2>&1 &

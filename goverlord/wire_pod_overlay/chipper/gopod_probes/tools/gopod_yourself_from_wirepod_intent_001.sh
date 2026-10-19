#!/bin/bash
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/gopod_yourself_from_wirepod_intent_001.sh
# Wire-Pod custom intent GOPOD_YOURSELF ("GOPOD Yourself"): hands the wake phrase to the
# Brobots chat for both Brobots, so each says his own call to action - Brobot 1 the
# newsletter, Brobot 2 crushn8r.ca (gopod_yourself in the settings, gopod_yourself.go).
# Both robots hear it; the chat drops the second within gopod_yourself.repeat_seconds.
# Output: /tmp/gopod_yourself_intent.log
source /home/goverlord/.gopod_alias_lib/local_config.sh
curl -s -X POST "$GOPOD_WIREPOD_BASE_URL/api-sdk/gopod_typed_input" \
    --data-urlencode "voice=1" --data-urlencode "serial=all" --data-urlencode "text=GOPOD Yourself" \
    > /tmp/gopod_yourself_intent.log 2>&1 &

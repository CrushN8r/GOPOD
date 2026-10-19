#!/usr/bin/env bash
# goverlord/wire_pod_overlay/session_dry_run.sh
#
# The session dry run (operator 2026-10-11: "Full run done by you, no robots,
# nothing spoken"; "a full account of all events of a session starting from
# timer 0 until ended"): one whole session through Wire-Pod's own paths with
# no robot and no sound (TestGopodSessionDryRun, gopod_session_dry_test.go) -
# live questions with the real LLM, start interview, start bingo test, start
# baby brobots, a live question after, and the thanks ending.
#
# Output (private, outside the repo):
#   ~/gopod_tts/sessions/dryrun/<date>/<time>_dry/sessions/<date>/<time>_session/
#       story.md, knobs.json   the session as a score, from 00:00:00
#       runs/                  each song run's own log and json
#       chat.jsonl             every line (voice lines with their library state)
#   ~/gopod_tts/sessions/dryrun/<date>/<time>_dry/full.log   every log line
set -euo pipefail

OVERLAY_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUN_DIR="${HOME}/gopod_tts/sessions/dryrun/$(date +%Y-%m-%d)/$(date +%H%M%S)_dry"
mkdir -p "${RUN_DIR}"

# The LLM: only the knowledge endpoint and model are read from Wire-Pod's settings.
read -r ENDPOINT MODEL < <(python3 - <<'PY'
import json, os
k = json.load(open(os.path.expanduser("~/wire-pod/chipper/apiConfig.json"))).get("knowledge", {})
print(k.get("endpoint", ""), k.get("model", ""))
del k
PY
)
if [ -z "${ENDPOINT}" ] || [ -z "${MODEL}" ]; then
  echo "SESSION_DRY no knowledge endpoint/model in Wire-Pod's settings" >&2
  exit 1
fi

echo "SESSION_DRY_START run=${RUN_DIR} model=${MODEL}"
START=$(date +%s)
GOPOD_OVERLAY_CHIPPER="${OVERLAY_DIR}/chipper" GOPOD_SESSION_DRY_OUT="${RUN_DIR}" \
GOPOD_HARVEST_ENDPOINT="${ENDPOINT}" GOPOD_HARVEST_MODEL="${MODEL}" \
  go -C "${HOME}/wire-pod/chipper" test -count=1 -v -timeout 60m -tags nolibopusfile \
  -overlay="${OVERLAY_DIR}/overlay.json" -run '^TestGopodSessionDryRun$' ./pkg/wirepod/ttr/ \
  2>&1 | tee "${RUN_DIR}/full.log" | grep -E "SESSION|OPERATOR|ANSWER|VOICE|^(--- |ok|FAIL|panic)" | tee "${RUN_DIR}/run.log"
echo "SESSION_DRY_SECONDS $(( $(date +%s) - START ))"

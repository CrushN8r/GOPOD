#!/usr/bin/env bash
# goverlord/wire_pod_overlay/harvest_dry_run.sh
#
# The harvest dry run (operator 2026-10-06): every question in
# chipper/gopod_probes/interview_harvest/harvest_questions.json through the
# Brobots' real chat path - same prompt, same LLM, same reply shaping - with no
# robots and no live file touched (TestGopodHarvestRun, gopod_harvest_test.go).
# Each answer is a take; the takes go to a run folder and onto the take sheet,
# one entry per value point, to pick keepers and edit words.
#
#   harvest_dry_run.sh               run it (a full dry run, about 10 minutes)
#   harvest_dry_run.sh render        only redraw take_sheet.md after editing take_sheet.json
#   harvest_dry_run.sh render-draft  only redraw the newest draft_performance.md after editing its .json
#   harvest_dry_run.sh best-takes    build a draft from the take sheet's keepers ("keep": true, one per point)
#   harvest_dry_run.sh questions F   a full dry run of question file F (same shape as harvest_questions.json)
#
# Each run also writes draft_performance.json/.md (operator 2026-10-07: "Do a
# full dry run ... fully featured, fully logged ... replay a 'draft performance
# run' to review & edit"): the whole show in order, with the Brobots' own
# answers. Typing "start draft" in the Brobots chat plays the newest one
# through the robots (gopod_interview_script.go). full.log is every log line.
#
# Output (private, outside the repo):
#   ~/gopod_tts/sessions/harvest/<date>/<time>_harvest/takes.json   this run
#   ~/gopod_tts/sessions/harvest/<date>/<time>_harvest/draft_performance.json/.md   this run's show - edit, then play
#   ~/gopod_tts/sessions/harvest/<date>/<time>_harvest/full.log     every log line
#   ~/gopod_tts/sessions/harvest/take_sheet.json                    every run (and every kept live session) - edit this
#   ~/gopod_tts/sessions/harvest/take_sheet.md                      the same, to read
set -euo pipefail

OVERLAY_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHIPPER="${OVERLAY_DIR}/chipper"
HARVEST_ROOT="${HOME}/gopod_tts/sessions/harvest"
RUN_DIR="${HARVEST_ROOT}/$(date +%Y-%m-%d)/$(date +%H%M%S)_harvest"
SHEET="${HARVEST_ROOT}/take_sheet.json"
render_sheet() {
  python3 "${CHIPPER}/gopod_probes/tools/gopod_take_sheet_render.py" "${SHEET}"
}
render_draft() {
  python3 "${CHIPPER}/gopod_probes/tools/gopod_draft_render.py" "$1"
}
if [ "${1:-}" = "render" ]; then
  render_sheet
  exit 0
fi
if [ "${1:-}" = "best-takes" ]; then
  python3 "${CHIPPER}/gopod_probes/tools/gopod_draft_from_picks.py" "${SHEET}"
  exit 0
fi
if [ "${1:-}" = "render-draft" ]; then
  render_draft "$(ls "${HARVEST_ROOT}"/*/*_harvest/draft_performance.json | sort | tail -1)"
  exit 0
fi
# 2026-10-08: "harvest_dry_run.sh questions <file>" asks that file's questions instead.
QUESTIONS=""
if [ "${1:-}" = "questions" ]; then
  QUESTIONS="$(realpath "${2:?questions file}")"
fi
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
  echo "HARVEST no knowledge endpoint/model in Wire-Pod's settings" >&2
  exit 1
fi

echo "HARVEST_START run=${RUN_DIR} model=${MODEL}"
START=$(date +%s)
GOPOD_OVERLAY_CHIPPER="${CHIPPER}" GOPOD_HARVEST_OUT="${RUN_DIR}" GOPOD_HARVEST_SHEET="${SHEET}" \
GOPOD_HARVEST_ENDPOINT="${ENDPOINT}" GOPOD_HARVEST_MODEL="${MODEL}" GOPOD_HARVEST_QUESTIONS="${QUESTIONS}" \
  go -C "${HOME}/wire-pod/chipper" test -count=1 -v -timeout 90m -tags nolibopusfile \
  -overlay="${OVERLAY_DIR}/overlay.json" -run '^TestGopodHarvestRun$' ./pkg/wirepod/ttr/ \
  2>&1 | tee "${RUN_DIR}/full.log" | grep -E "take [0-9]+/|HARVEST_DONE|no answer|^(--- |ok|FAIL|panic)" | tee "${RUN_DIR}/run.log"
echo "HARVEST_SECONDS $(( $(date +%s) - START ))"

# The take sheet, to read (take_sheet.json is the one to edit), and the draft.
render_sheet
[ -f "${RUN_DIR}/draft_performance.json" ] && render_draft "${RUN_DIR}/draft_performance.json"

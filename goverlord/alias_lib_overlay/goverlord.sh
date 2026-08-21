# === Goverlord suit aliases ===
export GOPOD_GOVERLORD_TOOL_DIR="${GOPOD_GOVERLORD_TOOL_DIR:-$HOME/.gopod_tools/goverlord}"

alias goverlord-suits='bash "$GOPOD_GOVERLORD_TOOL_DIR/10_create_llm_suits_and_measure.sh"'
alias suit-digest='bash "$GOPOD_GOVERLORD_TOOL_DIR/11_suit_digest_cycle_001.sh"'
alias suit-latest='ls -lt ~/.gopod_state/goverlord_suits | head -30'
alias suit-report='cat "$(ls -t ~/.gopod_state/goverlord_suits/suit_digest_report_*.txt | head -1)"'
alias suit-measure='cat "$(ls -t ~/.gopod_state/goverlord_suits/suit_digest_measurement_*.json | head -1)" | python3 -m json.tool'

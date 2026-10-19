# === GOPOD LLM routing ===

# Real GOPOD_OLLAMA_BASE value lives in a private, untracked local file -
# never a literal IP here (2026-08-15, GOPOLISHER_CUBE_SESSION_SWEEP_001.md).
GOPOD_LOCAL_CONFIG_PATH="${GOPOD_LOCAL_CONFIG_PATH:-$HOME/.gopod_alias_lib/local_config.sh}"
[ -f "$GOPOD_LOCAL_CONFIG_PATH" ] && . "$GOPOD_LOCAL_CONFIG_PATH"

export GOPOD_OLLAMA_BASE="${GOPOD_OLLAMA_BASE}"
export GOPOD_BROBOT_LLM="${GOPOD_BROBOT_LLM:-gemma2:2b}"
export GOPOD_GOVERLORD_LLM="${GOPOD_GOVERLORD_LLM:-llama3.2:3b}"
export GOPOD_GOVERLORD_DEEP_LLM="${GOPOD_GOVERLORD_DEEP_LLM:-qwen2.5:7b}"
export GOPOD_CODER_LLM="${GOPOD_CODER_LLM:-qwen2.5-coder-24k:latest}"

_gopod_llm_models() {
  curl -fsS "$GOPOD_OLLAMA_BASE/v1/models" | python3 -m json.tool
}

_gopod_llm_status() {
  echo "OLLAMA_BASE=$GOPOD_OLLAMA_BASE"
  echo "BROBOT_LLM=$GOPOD_BROBOT_LLM"
  echo "GOVERLORD_LLM=$GOPOD_GOVERLORD_LLM"
  echo "GOVERLORD_DEEP_LLM=$GOPOD_GOVERLORD_DEEP_LLM"
  echo "CODER_LLM=$GOPOD_CODER_LLM"
  echo
  _gopod_llm_models
}

_gopod_llm_ping() {
  local model="$1"
  local prompt="${2:-Say GOPOD ready in one short sentence.}"
  local raw="/tmp/gopod_llm_${model//[^a-zA-Z0-9]/_}.raw"

  curl -sS "$GOPOD_OLLAMA_BASE/v1/chat/completions" \
    -H "Content-Type: application/json" \
    -d "$(python3 - <<PY
import json
print(json.dumps({
  "model": "$model",
  "messages": [{"role": "user", "content": "$prompt"}],
  "stream": False,
  "max_tokens": 80,
  "temperature": 0.4
}))
PY
)" > "$raw"

  python3 - "$raw" <<'PY'
import json, sys
p=sys.argv[1]
s=open(p).read().strip()
if not s:
    print("LLM_EMPTY_RESPONSE")
    print("RAW_RESPONSE="+p)
    raise SystemExit(1)
try:
    d=json.loads(s)
except Exception:
    print("LLM_NON_JSON_RESPONSE")
    print("RAW_RESPONSE="+p)
    print(s[:500])
    raise SystemExit(1)

if "error" in d:
    print("LLM_ERROR")
    print(d["error"])
    raise SystemExit(1)

try:
    print(d["choices"][0]["message"]["content"].strip())
except Exception:
    print("LLM_UNEXPECTED_JSON_SHAPE")
    print("RAW_RESPONSE="+p)
    print(json.dumps(d, indent=2)[:1000])
    raise SystemExit(1)
PY
}

_gopod_llm_test_all() {
  echo "BROBOT:"
  _gopod_llm_ping "$GOPOD_BROBOT_LLM" "Return one robot-safe line: GOPOD ready."
  echo
  echo "GOVERLORD:"
  _gopod_llm_ping "$GOPOD_GOVERLORD_LLM" "Return one sentence as Goverlord about guiding the next useful proof."
  echo
  echo "DEEP:"
  _gopod_llm_ping "$GOPOD_GOVERLORD_DEEP_LLM" "Return one sentence explaining when deeper planning helps GOPOD."
  echo
  echo "CODER:"
  _gopod_llm_ping "$GOPOD_CODER_LLM" "Return one sentence explaining your role for code review."
}

alias llm-models='_gopod_llm_models'
alias llm-status='_gopod_llm_status'
alias llm-test-brobot='_gopod_llm_ping "$GOPOD_BROBOT_LLM" "Return one robot-safe line: GOPOD ready."'
alias llm-test-goverlord='_gopod_llm_ping "$GOPOD_GOVERLORD_LLM" "Return one short useful action packet."'
alias llm-test-deep='_gopod_llm_ping "$GOPOD_GOVERLORD_DEEP_LLM" "Return one short deeper planning sentence."'
alias llm-test-coder='_gopod_llm_ping "$GOPOD_CODER_LLM" "Return one short code-helper sentence."'
alias llm-test-all='_gopod_llm_test_all'

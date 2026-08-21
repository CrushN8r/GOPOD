# === Open WebUI / Goverlord control aliases ===

# Real GOPOD_OWUI_LAN_URL/GOPOD_OLLAMA_BASE values live in a private,
# untracked local file - never a literal IP here (2026-08-15,
# GOPOLISHER_CUBE_SESSION_SWEEP_001.md).
GOPOD_LOCAL_CONFIG_PATH="${GOPOD_LOCAL_CONFIG_PATH:-$HOME/.gopod_alias_lib/local_config.sh}"
[ -f "$GOPOD_LOCAL_CONFIG_PATH" ] && . "$GOPOD_LOCAL_CONFIG_PATH"

export OWUI_TOOL_DIR="${OWUI_TOOL_DIR:-$HOME/.gopod_tools/openwebui}"
export OWUI_STATE_DIR="${OWUI_STATE_DIR:-$HOME/.gopod_state/openwebui}"
export OWUI_SNAPSHOT_DIR="${OWUI_SNAPSHOT_DIR:-$HOME/openwebui_model_snapshots}"

# Script-backed aliases
alias owui-detect='bash "$OWUI_TOOL_DIR/00_detect_openwebui.sh"'
alias owui-routes='bash "$OWUI_TOOL_DIR/01_probe_openwebui_routes.sh"'
alias owui-export-script='bash "$OWUI_TOOL_DIR/02_export_openwebui_models.sh"'
alias owui-tune-script='bash "$OWUI_TOOL_DIR/03_tune_goverlord_model_json.sh"'
alias owui-import-script='bash "$OWUI_TOOL_DIR/04_import_openwebui_model.sh"'
alias owui-fix-port='bash "$OWUI_TOOL_DIR/05_fix_openwebui_port.sh"'
alias owui-chain='bash "$OWUI_TOOL_DIR/owui_chain_check.sh"'

# Current known good surfaces
alias owui-local='echo http://127.0.0.1:3000'
alias owui-lan='echo "$GOPOD_OWUI_LAN_URL"'
alias ollama-lan='echo "$GOPOD_OLLAMA_BASE"'

_owui_url() {
  for u in \
    http://127.0.0.1:3000 \
    "$GOPOD_OWUI_LAN_URL" \
    http://localhost:3000
  do
    if curl -fsSL --max-time 3 "$u" >/dev/null 2>&1; then
      echo "$u"
      return 0
    fi
  done
  echo "OPENWEBUI_NOT_READY_OR_NOT_FOUND"
  return 1
}

_owui_wait() {
  local u
  for i in $(seq 1 60); do
    u="$(_owui_url 2>/dev/null || true)"
    case "$u" in
      http*) echo "$u"; return 0 ;;
    esac
    sleep 2
  done
  echo "OPENWEBUI_WAIT_TIMEOUT"
  return 1
}

_owui_status() {
  echo "OPENWEBUI_URL=$(_owui_url || true)"
  echo
  echo "DOCKER:"
  docker ps --filter name=open-webui
  echo
  echo "PORTS:"
  ss -ltn | grep -E ':3000|:8080|:11434' || true
  echo
  echo "OLLAMA_MODELS:"
  curl -fsS "$GOPOD_OLLAMA_BASE/v1/models" 2>/dev/null | python3 -m json.tool | head -80 || echo "OLLAMA_BLOCKED"
}

_owui_api_test() {
  local u
  u="$(_owui_wait)" || return 1

  echo "OWUI_URL=$u"

  if [ -z "${OWUI_TOKEN:-}" ]; then
    echo "OWUI_TOKEN_MISSING"
    echo "Open WebUI is reachable. API needs token."
    return 1
  fi

  curl -fsS "$u/api/v1/models" \
    -H "Authorization: Bearer $OWUI_TOKEN" \
    | python3 -m json.tool
}

_owui_export() {
  local u out
  u="$(_owui_wait)" || return 1
  mkdir -p "$OWUI_SNAPSHOT_DIR"
  out="$OWUI_SNAPSHOT_DIR/openwebui_models_$(date +%Y%m%d_%H%M%S).json"

  if [ -z "${OWUI_TOKEN:-}" ]; then
    echo "OWUI_TOKEN_MISSING"
    echo "Open WebUI is reachable at $u"
    echo "Create/copy API key from Open WebUI, then:"
    echo "export OWUI_TOKEN='...'"
    return 1
  fi

  curl -fsS "$u/api/v1/models" \
    -H "Authorization: Bearer $OWUI_TOKEN" \
    | python3 -m json.tool > "$out" &&
  echo "$out"
}

_owui_debug() {
  local u
  u="$(_owui_wait)" || return 1

  echo "OWUI_URL=$u"
  echo
  echo "ROOT:"
  curl -sSL -o /tmp/owui_root.tmp -w '%{http_code} %{content_type}\n' "$u" || true
  head -5 /tmp/owui_root.tmp || true

  echo
  echo "MODELS_NO_TOKEN:"
  curl -sSL -o /tmp/owui_models.tmp -w '%{http_code} %{content_type}\n' "$u/api/v1/models" || true
  head -20 /tmp/owui_models.tmp || true

  echo
  echo "MODELS_WITH_TOKEN:"
  if [ -n "${OWUI_TOKEN:-}" ]; then
    curl -sSL -o /tmp/owui_models_token.tmp -w '%{http_code} %{content_type}\n' \
      -H "Authorization: Bearer $OWUI_TOKEN" \
      "$u/api/v1/models" || true
    head -20 /tmp/owui_models_token.tmp || true
  else
    echo "OWUI_TOKEN_MISSING"
  fi
}

# Short aliases
alias owui-url='_owui_url'
alias owui-wait='_owui_wait'
alias owui-status='_owui_status'
alias owui-api-test='_owui_api_test'
alias owui-export='_owui_export'
alias owui-debug='_owui_debug'

# Main flow aliases
alias owui-fix='owui-fix-port && owui-wait'
alias owui-ready='owui-status && owui-api-test'

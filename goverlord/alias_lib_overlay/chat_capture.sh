# === GOPOD chat/session JSON capture ===

gopodcap() {
  local source="$1"
  shift
  python3 "$HOME/.gopod_alias_lib/gopod_json_capture.py" "$source" "$@"
}

gopodcap-gopod() {
  local source="$1"
  shift
  python3 "$HOME/.gopod_alias_lib/gopod_json_capture.py" "$source" bash -lc "
    shopt -s expand_aliases
    source ~/.gopod_alias_lib/core.sh
    source ~/.gopod_alias_lib/frame0.sh
    source ~/.gopod_alias_lib/brobots.sh
    source ~/.gopod_alias_lib/demo.sh
    cd ~/crushn8r_git/GOPOD
    $*
  "
}

gopod-json-tail() {
  tail -n 80 "$HOME/gopod_chat_jsons.txt"
}

gopod-json-clear() {
  : > "$HOME/gopod_chat_jsons.txt"
  echo "CLEARED ~/gopod_chat_jsons.txt"
}

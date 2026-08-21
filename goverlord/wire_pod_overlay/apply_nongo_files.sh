#!/bin/bash
# Applies this folder's copies of the files Go's own -overlay build flag
# can't substitute for - either because they're not part of `go build`'s
# compilation graph (start.sh, webroot/index.html, root/.gitignore), or
# because they're config/content GOPOD owns but Wire-Pod never compiles at
# all (customIntents.json, the two wire-pod_*.txt prompt files).
#
# Diff-then-copy only - never a blind overwrite. A file that already
# matches live is reported and left untouched.
#
# customIntents.json gotcha (see CLAUDE.md): a running wire-pod.service
# caches this file in memory at startup, and the web config UI overwrites
# the ENTIRE file on any edit, not just the touched intent. After this
# script applies a changed customIntents.json, restart wire-pod.service
# before touching the web config UI for any intent - otherwise a fix here
# can be silently clobbered by the UI's own next save.
#
# animation_vocab.json (added 2026-08-10): its Go loader panics if it's
# missing or invalid at startup (same hard rule as apiConfig.json), but
# unlike apiConfig.json it holds no live API keys/secrets - just the
# animation token/alias vocabulary - so it's safe to mirror. Its loader
# (animation_vocab.go) is not one of the overlaid native .go files, so its
# runtime.Caller-based default path always resolves to this exact live
# location regardless of how the binary was built - this script's diff-then-
# copy writes to that same path, never anywhere else.
#
# sounds/*.wav (added 2026-08-12): playSound's audio assets, referenced by soundMap in
# kgsim_cmds.go. Binary, diff -q handles that fine (byte comparison either way). Not a
# secret, not live-generated state - a static asset GOPOD ships, safe to mirror like
# animation_vocab.json above.
#
# Usage: ./apply_nongo_files.sh [--dry-run]

set -euo pipefail

OVERLAY_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WIREPOD_ROOT="/home/goverlord/wire-pod"
DRY_RUN=0
[ "${1:-}" = "--dry-run" ] && DRY_RUN=1

declare -a PAIRS=(
  "chipper/start.sh:chipper/start.sh"
  "chipper/webroot/index.html:chipper/webroot/index.html"
  "chipper/webroot/css/gopod_chat_bubbles.css:chipper/webroot/css/gopod_chat_bubbles.css"
  "root/.gitignore:.gitignore"
  "chipper/customIntents.json:chipper/customIntents.json"
  "chipper/wire-pod_brobots_prompt.txt:chipper/wire-pod_brobots_prompt.txt"
  "chipper/wire-pod_response_prompt.txt:chipper/wire-pod_response_prompt.txt"
  "chipper/animation_vocab.json:chipper/animation_vocab.json"
  "chipper/sounds/danger-will-robinson.wav:chipper/sounds/danger-will-robinson.wav"
)

changed=0
for pair in "${PAIRS[@]}"; do
  overlay_rel="${pair%%:*}"
  live_rel="${pair##*:}"
  overlay_path="$OVERLAY_DIR/$overlay_rel"
  live_path="$WIREPOD_ROOT/$live_rel"

  if [ ! -f "$overlay_path" ]; then
    echo "APPLY_SKIP missing_overlay_copy $overlay_rel"
    continue
  fi
  if [ ! -f "$live_path" ]; then
    echo "APPLY_NEW $live_rel (no live file yet)"
    if [ "$DRY_RUN" -eq 0 ]; then
      cp "$overlay_path" "$live_path"
    fi
    changed=$((changed + 1))
    continue
  fi
  if diff -q "$overlay_path" "$live_path" > /dev/null 2>&1; then
    echo "APPLY_UNCHANGED $live_rel"
  else
    echo "APPLY_UPDATE $live_rel"
    if [ "$DRY_RUN" -eq 0 ]; then
      cp "$overlay_path" "$live_path"
    fi
    changed=$((changed + 1))
  fi
done

if [ "$changed" -gt 0 ]; then
  echo "APPLY_DONE files_changed=$changed"
  if printf '%s\n' "${PAIRS[@]}" | grep -q customIntents.json; then
    echo "APPLY_REMINDER restart wire-pod.service before touching the web config UI (customIntents.json cache gotcha)"
  fi
else
  echo "APPLY_DONE files_changed=0 (already in sync)"
fi

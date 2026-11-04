#!/usr/bin/env bash
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/gopod_page_replay.sh
# Runs GOPOD's page replay (gopod_page_replay_001.js): the Brobots chat page with no
# robots. The test browser engine (jsdom) is kept in ~/.cache/gopod_page_replay,
# installed once on the first run. Optional argument: a saved feed file.
set -euo pipefail
here="$(cd "$(dirname "$0")" && pwd)"
webroot="$(cd "$here/../../webroot" && pwd)"
lib="$HOME/.cache/gopod_page_replay"
if [ ! -d "$lib/node_modules/jsdom" ]; then
  mkdir -p "$lib"
  (cd "$lib" && npm init -y >/dev/null 2>&1 && npm install jsdom@24.1.3 --no-audit --no-fund >/dev/null)
fi
# 2026-11-03 (the clusters plan, step 7): the settings as the page gets them - the
# topic-clusters laid over the settings folder - from the running Wire-Pod.
full="$(mktemp)"
trap 'rm -f "$full"' EXIT
if curl -s -f -m 10 -o "$full" "http://localhost:8080/gopod_brobots_settings.json"; then
  export GOPOD_REPLAY_SETTINGS="$full"
else
  echo "page replay: Wire-Pod isn't answering - the settings folder alone (no topic-clusters)" >&2
fi
NODE_PATH="$lib/node_modules" timeout 600 node "$here/gopod_page_replay_001.js" "$webroot" "${1:-}"

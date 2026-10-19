#!/usr/bin/env bash
# goverlord/wire_pod_overlay/searxng/run_gopod_searxng.sh
# GOPOD's own SearXNG (2026-10-17, operator: "Please remap/build SearXNG container
# to belong to GOPOD"): the search behind "system-search" (live guided search
# replies). Local only (127.0.0.1:8055), restarts with the machine. Its real
# settings - with a secret key made once here - live outside the repo, in
# ~/gopod_searxng/settings.yml; settings.example.yml beside this script is the
# public copy (JSON results on, the engines that answer from the Jetson).
#   bash run_gopod_searxng.sh          start (or re-create) it
#   docker logs gopod-searxng          its log
set -euo pipefail
here="$(cd "$(dirname "$0")" && pwd)"
private="$HOME/gopod_searxng"
image="searxng/searxng:2026.5.31-7159b8aed"
mkdir -p "$private/data"
if [ ! -f "$private/settings.yml" ]; then
  secret="$(python3 -c 'import secrets; print(secrets.token_hex(32))')"
  sed "s/__GOPOD_SEARXNG_SECRET__/$secret/" "$here/settings.example.yml" > "$private/settings.yml"
  chmod 600 "$private/settings.yml"
fi
docker rm -f gopod-searxng >/dev/null 2>&1 || true
docker run -d --name gopod-searxng --restart unless-stopped \
  -p 127.0.0.1:8055:8080 \
  -e SEARXNG_BASE_URL=http://localhost:8055/ \
  -v "$private/settings.yml:/etc/searxng/settings.yml:ro" \
  -v "$private/data:/var/cache/searxng" \
  "$image"
echo "gopod-searxng started on http://127.0.0.1:8055"

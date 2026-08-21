#!/bin/bash
# Builds wire-pod's chipper binary using this folder's copies of the 9
# native files GOPOD has modified, via Go's own -overlay build flag
# (Go 1.16+) - the real ~/wire-pod checkout is never written to.
#
# Only the .go files under chipper/pkg/... are wired into overlay.json -
# 7 of them as of 2026-07-30 (added gopod_render_scaffold.go). start.sh,
# webroot/index.html, and root/.gitignore are kept here as plain
# snapshot copies but are NOT part of `go build`'s compilation graph, so
# there's nothing for an overlay to substitute for them. If those need to
# take effect, copy them onto the real ~/wire-pod tree directly (same as
# any other non-Go file edit).
#
# Usage: ./build_with_overlay.sh [output_path]
# Default output: /tmp/chipper_overlay_build (never overwrites the live
# ~/wire-pod/chipper/chipper binary - swap that in yourself if you want it
# live, same manual-verify-then-swap discipline as any other rebuild).

set -euo pipefail

OVERLAY_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WIREPOD_CHIPPER_DIR="/home/goverlord/wire-pod/chipper"
OUT="${1:-/tmp/chipper_overlay_build}"

VOSK_LIB="/home/goverlord/.local/lib/python3.10/site-packages/vosk"
VOSK_INC="/home/goverlord/go/pkg/mod/github.com/kercre123/vosk-api@v1.0.1/src"
COMMIT_HASH="$(cd /home/goverlord/wire-pod && git rev-parse --short HEAD)"
GOTAGS="nolibopusfile"
GOLDFLAGS="-X 'github.com/kercre123/wire-pod/chipper/pkg/vars.CommitSHA=${COMMIT_HASH}'"

cd "$WIREPOD_CHIPPER_DIR"
CGO_ENABLED=1 \
CGO_CFLAGS="-I${VOSK_INC}" \
CGO_CPPFLAGS="-I${VOSK_INC}" \
CGO_LDFLAGS="-L${VOSK_LIB} -lvosk -ldl -lpthread" \
LD_LIBRARY_PATH="${VOSK_LIB}:$LD_LIBRARY_PATH" \
/usr/local/go/bin/go build -tags "$GOTAGS" -ldflags="${GOLDFLAGS}" \
  -overlay="${OVERLAY_DIR}/overlay.json" \
  -o "$OUT" cmd/vosk/main.go

echo "OVERLAY_BUILD_OK -> $OUT"

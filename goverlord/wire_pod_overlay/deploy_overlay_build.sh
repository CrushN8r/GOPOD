#!/bin/bash
# The one blessed path from "edited a file in this overlay folder" to "live
# ./chipper binary updated." Built 2026-08-02 as step 6 of the native-tree
# migration (see tech/WIRED-POD.md, "The rule: stay thin" / "Where this
# stands today"). Replaces the old manual ritual of hand-building to a temp
# path and remembering to swap it in.
#
# What it does NOT do: restart wire-pod.service. Swapping the binary file
# takes effect on the next process start, but firing that restart (`wpr`)
# stays the operator's own action, same standing rule as every rebuild this
# project has done - this script stops one step short of that on purpose.
#
# Usage: ./deploy_overlay_build.sh
#   1. Builds via build_with_overlay.sh to a temp path.
#   2. Sanity-checks the result (file type, non-zero size).
#   3. Backs up the current live binary with a timestamp (never overwritten,
#      never deleted by this script).
#   4. Swaps the new binary into place (`mv`, not `cp` - the live binary may
#      be held open by a running process; `mv` replaces the directory entry
#      without needing to write to a busy inode, same trick used earlier
#      this session).

set -euo pipefail

OVERLAY_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIVE_BIN="/home/goverlord/wire-pod/chipper/chipper"
TS="$(date +%Y%m%d_%H%M%S)"
TMP_OUT="/tmp/chipper_overlay_deploy_${TS}"
BACKUP="/home/goverlord/wire-pod/chipper/chipper.bak_${TS}_pre_overlay_deploy"

echo "DEPLOY_STEP building via overlay..."
bash "${OVERLAY_DIR}/build_with_overlay.sh" "$TMP_OUT"

if [ ! -s "$TMP_OUT" ]; then
  echo "DEPLOY_FAILED build output missing or empty: $TMP_OUT" >&2
  exit 1
fi
if ! file "$TMP_OUT" | grep -q "ELF 64-bit"; then
  echo "DEPLOY_FAILED build output isn't an ELF binary: $TMP_OUT" >&2
  exit 1
fi

echo "DEPLOY_STEP backing up live binary -> $BACKUP"
cp -p "$LIVE_BIN" "$BACKUP"

echo "DEPLOY_STEP swapping new binary into place"
mv "$TMP_OUT" "$LIVE_BIN"
chmod +x "$LIVE_BIN"

echo "DEPLOY_OK new binary in place at $LIVE_BIN"
echo "DEPLOY_REMINDER restart wire-pod.service (wpr) yourself when ready - this script does not restart it"

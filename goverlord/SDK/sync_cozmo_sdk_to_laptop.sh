#!/usr/bin/env bash
set -euo pipefail

LAPTOP="${GOPOD_LAPTOP_SSH:-gopod-laptop}"
REMOTE_DIR="/home/gomad/crushn8r_git/GOPOD/SDK/sources"

ssh "$LAPTOP" "mkdir -p '$REMOTE_DIR'"
rsync -a --delete "sources/cozmo-python-sdk/" "$LAPTOP:$REMOTE_DIR/cozmo-python-sdk/"

ssh "$LAPTOP" '
set -e
cd /home/gomad/crushn8r_git/GOPOD/SDK
python3 -m venv .venv-cozmo
. .venv-cozmo/bin/activate
python -m pip install --upgrade pip setuptools wheel
python -m pip install -e sources/cozmo-python-sdk || true
echo "COZMO_SDK_SYNC_DONE"
echo "If install failed, source is still staged at: /home/gomad/crushn8r_git/GOPOD/SDK/sources/cozmo-python-sdk"
'

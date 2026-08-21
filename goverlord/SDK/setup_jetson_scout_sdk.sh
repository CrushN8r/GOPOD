#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

python3 -m venv .venv-scout
. .venv-scout/bin/activate
python -m pip install --upgrade pip setuptools wheel

echo "SCOUT_SOURCE_READY: $(pwd)/sources/Scout-open-source"
echo "Scout repo is staged. Inspect its README/build docs before installing ROS/system dependencies."

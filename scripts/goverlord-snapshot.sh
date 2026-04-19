#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PRIVATE_CONFIG_PATH="$REPO_ROOT/config/system.json"
LOG="/tmp/goverlord-snapshot_$(date +%Y%m%d_%H%M%S).log"

exec > >(tee -a "$LOG") 2>&1

echo "=== Jetson-Goverlord SNAPSHOT $(date) ==="
hostname
uptime
whoami
pwd
python3 --version 2>/dev/null || python --version

echo -e "\n=== Project files ==="
find "$REPO_ROOT" -maxdepth 3 \( -name "*.py" -o -name "*.json" \) | head -20

echo -e "\n=== Registry files ==="
find "$REPO_ROOT/gomads" \( -name "*registry*" -o -name "*persona*" \) 2>/dev/null | head -20

echo -e "\n=== Runtime config ==="
if [ -f "$PRIVATE_CONFIG_PATH" ]; then
    echo "✅ Private runtime config present: $PRIVATE_CONFIG_PATH"
    T560_BASE="$(python3 - <<'PY'
import json
from pathlib import Path
config_path = Path("/home/goverlord/crushn8r_git/GOPOD/config/system.json")
with open(config_path, "r", encoding="utf-8") as handle:
    config = json.load(handle)
print(config.get("t560_base", ""))
PY
)"
    if [ -n "$T560_BASE" ]; then
        echo "Configured T560 base: $T560_BASE"
        curl -s -m 3 -I "$T560_BASE/" | head -3 || echo "Configured T560 API not responding"
    else
        echo "⚠️ Config present but t560_base is empty"
    fi
else
    echo "⚠️ Private runtime config missing: $PRIVATE_CONFIG_PATH"
    echo "T560 reachability skipped because the target is intentionally unknown."
fi

echo -e "\n=== Log saved: $LOG ==="

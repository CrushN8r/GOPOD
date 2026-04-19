#!/bin/bash

set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CORE_PATH="$REPO_ROOT/scripts/goverlord_core.py"
CONFIG_TEMPLATE_PATH="$REPO_ROOT/config/system.template.json"
CONFIG_PATH="$REPO_ROOT/config/system.json"

echo "=== Jetson-Goverlord PIPELINE HEALTH ==="
if [ -f "$CORE_PATH" ]; then
    echo "✅ goverlord_core.py present at $CORE_PATH"
    python3 -m py_compile "$CORE_PATH" >/dev/null 2>&1 && echo "✅ Core compiles cleanly" || echo "❌ Core has compile errors"
else
    echo "❌ goverlord_core.py missing at $CORE_PATH"
fi

if [ -f "$CONFIG_TEMPLATE_PATH" ]; then
    echo "✅ Config template present at $CONFIG_TEMPLATE_PATH"
else
    echo "❌ Config template missing at $CONFIG_TEMPLATE_PATH"
fi

if [ -f "$CONFIG_PATH" ]; then
    echo "✅ Runtime config present at $CONFIG_PATH"
else
    echo "⚠️ Runtime config missing at $CONFIG_PATH"
    echo "   Copy config/system.template.json to config/system.json and fill private values."
fi

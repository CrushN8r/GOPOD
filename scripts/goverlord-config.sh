#!/scripts/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TEMPLATE_PATH="$REPO_ROOT/config/system.template.json"
PRIVATE_CONFIG_PATH="$REPO_ROOT/config/system.json"
REGISTRY_PATH="$REPO_ROOT/gomads/registry/personas.json"

echo "=== Jetson-Goverlord CONFIG DUMP ==="
echo "Repo root: $REPO_ROOT"
echo

if [ -f "$TEMPLATE_PATH" ]; then
    echo "✅ Config template present: $TEMPLATE_PATH"
else
    echo "❌ Config template missing: $TEMPLATE_PATH"
fi

if [ -f "$PRIVATE_CONFIG_PATH" ]; then
    echo "✅ Private runtime config present: $PRIVATE_CONFIG_PATH"
else
    echo "⚠️ Private runtime config missing: $PRIVATE_CONFIG_PATH"
    echo "   Create it from $TEMPLATE_PATH and fill in private values."
fi

if [ -f "$REGISTRY_PATH" ]; then
    echo "✅ Persona registry present: $REGISTRY_PATH"
else
    echo "⚠️ Persona registry missing: $REGISTRY_PATH"
fi

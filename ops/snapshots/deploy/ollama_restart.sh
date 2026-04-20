#!/usr/scripts/env bash
set -euo pipefail

sudo systemctl restart ollama
sudo systemctl is-active --quiet ollama
curl --fail --silent http://localhost:11434/api/tags >/dev/null

echo "Ollama service restarted and API verified at http://localhost:11434/api/tags"

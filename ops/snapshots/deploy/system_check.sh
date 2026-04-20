#!/usr/scripts/env bash
set -euo pipefail

echo "[1/4] Checking Open WebUI"
curl --fail --silent --show-error http://localhost:8080 >/dev/null

echo "[2/4] Checking Ollama tags endpoint"
curl --fail --silent --show-error http://localhost:11434/api/tags >/dev/null

echo "[3/4] Validating Docker container state"
docker ps --format '{{.Names}} {{.Status}}' | grep -E '^open-webui '

echo "[4/4] Validating Ollama system service"
systemctl is-active --quiet ollama

echo "System checks passed"

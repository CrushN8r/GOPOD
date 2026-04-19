#!/usr/bin/env bash
set -euo pipefail

echo "=== GOPOD STACK LOCK START ==="

########################################
# 1. VERIFY OLLAMA BACKEND (SYSTEMD ONLY)
########################################

echo "[1/4] Checking Ollama backend..."

if systemctl is-active --quiet ollama; then
    echo "OK: ollama systemd service active"
else
    echo "FAIL: ollama systemd service not active"
    exit 1
fi

# Ensure no docker ollama contamination
if docker ps --format '{{.Names}}' | grep -q "ollama"; then
    echo "ERROR: Docker ollama container detected - removing"
    docker rm -f ollama || true
fi

echo "OK: Docker ollama container absent"

########################################
# 2. VERIFY OPEN WEBUI
########################################

echo "[2/4] Checking Open WebUI..."

if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080 | grep -q "200"; then
    echo "OK: Open WebUI reachable"
else
    echo "FAIL: Open WebUI not reachable"
    exit 1
fi

########################################
# 3. VERIFY OLLAMA API
########################################

echo "[3/4] Checking Ollama API..."

if curl -s http://localhost:11434/api/tags > /dev/null; then
    echo "OK: Ollama API responding"
else
    echo "FAIL: Ollama API unreachable"
    exit 1
fi

########################################
# 4. DOCKER STATE VALIDATION
########################################

echo "[4/4] Checking Docker state..."

OPENWEBUI_COUNT=$(docker ps --format '{{.Names}}' | grep -c "open-webui" || true)
OLLAMA_COUNT=$(docker ps --format '{{.Names}}' | grep -c "ollama" || true)

if [ "$OPENWEBUI_COUNT" -ne 1 ]; then
    echo "FAIL: Open WebUI container mismatch"
    exit 1
fi

if [ "$OLLAMA_COUNT" -ne 0 ]; then
    echo "FAIL: Unexpected ollama container present"
    exit 1
fi

echo "OK: Docker state clean"

########################################
# FINAL STATE OUTPUT
########################################

echo "=== GOPOD STACK LOCK COMPLETE ==="
echo "Open WebUI: PASS"
echo "Ollama API: PASS"
echo "Backend source: systemd"
echo "Docker contamination: NONE"
echo "System state: STABLE"

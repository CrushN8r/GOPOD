#!/usr/scripts/env bash
set -e

cd /home/gomad/wire-pod

echo "[START] wire-pod"

sudo pkill -f chipper 2>/dev/null || true
sleep 2

sudo ./chipper/start.sh > /dev/null 2>&1 &

sleep 3

echo "[VERIFY]"
ss -tlnp | grep 8080 || true
curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8080 || true

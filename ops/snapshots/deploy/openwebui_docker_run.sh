#!/usr/scripts/env bash
set -euo pipefail

IMAGE="ghcr.io/open-webui/open-webui:main"
CONTAINER_NAME="open-webui"
VOLUME_NAME="open-webui"

docker volume create "${VOLUME_NAME}" >/dev/null

if docker ps -a --format '{{.Names}}' | grep -Fxq "${CONTAINER_NAME}"; then
  docker rm -f "${CONTAINER_NAME}" >/dev/null
fi

docker run -d \
  --name "${CONTAINER_NAME}" \
  --network host \
  --restart unless-stopped \
  -v "${VOLUME_NAME}:/app/backend/data" \
  "${IMAGE}"

echo "Open WebUI redeployed at http://localhost:8080"

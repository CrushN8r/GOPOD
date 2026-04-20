# Known State

This repository captures a stable snapshot baseline of a Jetson Orin NX local AI node as observed on 2026-04-17.

## Baseline Status

- Classified as: `production-ready local AI node`
- UI availability check: passed on `http://localhost:8080`
- Ollama API check: passed on `http://localhost:11434/api/tags`
- Open WebUI container status: healthy
- Ollama system service status: active and running

## Stable Production Baseline

The stable deployment baseline is:

- Open WebUI in Docker with host networking
- Open WebUI persistent SQLite-backed data in Docker volume `open-webui`
- Ollama running as a host systemd service
- Local inference models stored in Ollama and callable over localhost

## Observed Deviation Recorded for Accuracy

An `ollama` Docker container also existed during capture and was in a restart loop. This is preserved in the snapshot metadata because it is part of the live machine state, but it is not part of the recommended production baseline and should be treated as residual state.

## Deterministic Clone Guidance

For deployment cloning across Jetson nodes, reproduce the stable baseline and ignore the residual restarting `ollama` container unless there is a deliberate decision to containerize the backend.

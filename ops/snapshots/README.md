# GOPOD-AI-STACK-SNAPSHOT

Reproducible snapshot of a Jetson Orin NX local AI node captured on 2026-04-17.

This repository preserves:

- Observed machine state for the current node
- Deployment artifacts for Open WebUI and Ollama
- Recovery instructions for rebuilding the node from scratch
- A deterministic baseline suitable for cloning across Jetson systems

## Snapshot Summary

- Platform: NVIDIA Jetson Orin NX Engineering Reference Developer Kit
- OS: Ubuntu 22.04.5 LTS on ARM64
- Kernel: `5.15.185-tegra`
- LLM backend: Ollama via `ollama.service`
- UI layer: Open WebUI in Docker using host networking
- Open WebUI endpoint: `http://localhost:8080`
- Ollama endpoint: `http://127.0.0.1:11434`
- Persistent UI data volume: `open-webui:/app/backend/data`

## Important Operational Note

The intended production baseline is:

- Ollama managed by systemd
- Open WebUI managed by Docker

During capture, a separate Docker container named `ollama` was also present and stuck in a restart loop. That container is included in the observed-state JSON for accuracy, but it is not part of the recommended recovery baseline because it duplicates the active systemd-managed Ollama service.

## Repository Layout

- `system/`: captured host, container, model, network, and port metadata
- `deploy/`: reproducible deployment and verification scripts
- `docs/`: architecture, recovery, and known-good-state documentation
- `git/`: canonical commit message for this snapshot

## Recovery Target

The recovery target for a clean cloned node is:

1. Install Ollama and run it as `ollama.service`
2. Pull the recorded model set into Ollama
3. Create the `open-webui` volume
4. Recreate the Open WebUI container with host networking and restart policy
5. Validate `8080` and `11434` locally with `deploy/system_check.sh`

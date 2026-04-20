# Architecture

## Overview

This node is a two-tier local AI stack running on a Jetson Orin NX:

- UI layer: Open WebUI in Docker
- Backend layer: Ollama running as a systemd service on the host

## Request Flow

1. A browser connects to Open WebUI on `http://localhost:8080` or `http://<LAN-IP>:8080`.
2. Open WebUI accepts the chat or completion request.
3. Open WebUI routes model inference calls to the local Ollama backend.
4. Ollama serves the selected model through `http://127.0.0.1:11434`.
5. The generated response flows back through Open WebUI to the user session.

## Model Routing Flow

- Open WebUI is running with host networking.
- The container exposes the UI on port `8080`.
- Ollama is bound to loopback on port `11434`.
- Open WebUI treats the local Ollama instance as the inference provider.
- Installed models are selected dynamically through the Ollama tags inventory.

## Data Persistence Model

- Open WebUI persists state in Docker volume `open-webui`.
- The volume is mounted at `/app/backend/data`.
- Persistent contents include the SQLite database, application settings, uploads, and model-related caches.
- The observed Open WebUI configuration auto-downloads the embedding model `sentence-transformers/all-MiniLM-L6-v2` into that persistent data path.
- Ollama model blobs are retained by the host-side Ollama installation and must be repopulated or copied separately during recovery.

## Observed Runtime Detail

At capture time, the production path was:

- `ollama.service` active and healthy
- `open-webui` Docker container active and healthy

A separate Docker container named `ollama` also existed but was restarting continuously. It does not represent the stable serving path and should not be treated as part of the clean deployment baseline.

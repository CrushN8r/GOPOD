# Recovery Procedure

## Goal

Restore this Jetson node as a production-ready local AI system with:

- Ollama running as a systemd service
- Open WebUI running in Docker
- The recorded model inventory restored into Ollama

## 1. Base Host Preparation

1. Install Ubuntu 22.04 ARM64 on the Jetson device.
2. Confirm JetPack and NVIDIA userspace match the Jetson Orin NX platform.
3. Install Docker and enable it:
   `sudo systemctl enable --now docker`
4. Install Ollama on the host so `ollama.service` is available.
5. Enable Ollama on boot:
   `sudo systemctl enable --now ollama`

## 2. Restore Ollama Service

1. Verify the service is active:
   `systemctl status ollama --no-pager`
2. Verify the API is reachable:
   `curl http://localhost:11434/api/tags`
3. If needed, restart it with:
   `./deploy/ollama_restart.sh`

## 3. Re-pull Recorded Models

Pull the full observed model set:

1. `ollama pull qwen2.5-coder:7b`
2. `ollama pull qwen2.5:7b`
3. `ollama pull llama3.2:3b`
4. `ollama pull gemma2:2b`
5. `ollama pull phi3:mini`
6. `ollama pull phi3:latest`
7. `ollama pull mistral:latest`
8. `ollama pull llama3:latest`

After pulls complete, confirm digests and metadata against `system/ollama_models.json`.

## 4. Restore Open WebUI

1. Ensure Docker is active:
   `systemctl status docker --no-pager`
2. Recreate the persistent volume and container:
   `./deploy/openwebui_docker_run.sh`
3. Confirm the UI responds:
   `curl -I http://localhost:8080`

## 5. Restore Persistent Data if Available

If a previous `open-webui` Docker volume backup exists:

1. Stop the container:
   `docker rm -f open-webui`
2. Restore the volume contents into `/var/lib/docker/volumes/open-webui/_data`
3. Start the container again with `./deploy/openwebui_docker_run.sh`

Without a volume backup, Open WebUI will recreate a fresh SQLite-backed state on first boot.

## 6. Validate End-to-End Operation

1. Run `./deploy/system_check.sh`
2. Open `http://localhost:8080`
3. Verify the expected models appear in the Open WebUI model list
4. Run a test prompt through Open WebUI to confirm inference succeeds

## 7. Clean Baseline Guidance

Do not restore the separate Docker container named `ollama` unless you intentionally want a containerized Ollama topology. The clean baseline captured by this repository uses the host `ollama.service` instead, and that is the production-ready recovery target.

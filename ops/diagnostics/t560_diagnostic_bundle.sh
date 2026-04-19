#!/usr/scripts/env bash
set -euo pipefail

# T560 system check commands
uname -a
hostnamectl
date
uptime
ip addr show
ip route
ss -tulpn
systemctl is-active docker
systemctl is-active open-webui
docker ps -a

# OpenWebUI connectivity check commands
curl -I http://127.0.0.1:8080
curl -I http://192.168.1.6:8080
curl -sS http://127.0.0.1:8080/api/version
curl -sS http://192.168.1.6:8080/api/version

# Tool endpoint validation calls
curl -I "http://192.168.1.6:8080/api-sdk/assume_behavior_control?priority=high&serial=0dd1b9e9"
curl -I "http://192.168.1.6:8080/api-sdk/say_text?text=test&serial=0dd1b9e9"
curl -I "http://192.168.1.6:8080/api-sdk/release_behavior_control?serial=0dd1b9e9"
curl -I "http://192.168.1.6:8080/api-sdk/assume_behavior_control?priority=high&serial=0dd1d8bf"
curl -I "http://192.168.1.6:8080/api-sdk/say_text?text=test&serial=0dd1d8bf"
curl -I "http://192.168.1.6:8080/api-sdk/release_behavior_control?serial=0dd1d8bf"

#!/usr/scripts/env bash
set -euo pipefail

# Node and npm status checks
node --version
npm --version
which node
which npm

# Local runtime health checks
uname -a
date
uptime
df -h
free -h
ip addr show
ss -tulpn
systemctl --type=service --state=running

# GOPOD folder integrity check commands
pwd
ls -la
find /home/goverlord/crushn8r_git/GOPOD -maxdepth 3 -type d | sort
find /home/goverlord/crushn8r_git/GOPOD/systems/tension -maxdepth 3 \( -type f -o -type d \) | sort
find /home/goverlord/crushn8r_git/GOPOD/tools -maxdepth 3 \( -type f -o -type d \) | sort
find /home/goverlord/crushn8r_git/GOPOD/config -maxdepth 2 \( -type f -o -type d \) | sort

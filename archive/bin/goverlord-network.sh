#!/scripts/bash
LOG="/tmp/goverlord-snapshot_$(date +%Y%m%d_%H%M%S).log"
exec > >(tee -a "$LOG") 2>&1
echo "=== Jetson-Goverlord SNAPSHOT $(date) ==="
hostname; uptime; whoami; pwd; python3 --version 2>/dev/null || python --version
echo -e "\n=== Project files ==="; find . -maxdepth 3 -name "*.py" -o -name "*.json" | head -20
echo -e "\n=== Robot registry ==="; find . -name "*robot*" -o -name "*vector*" -o -name "*registry*" 2>/dev/null | head -10
echo -e "\n=== T560 reachability ==="; ping -c 1 T560 2>/dev/null || echo "T560 unreachable"
curl -s -m 3 -I http://T560:8080/ | head -3 || echo "T560 API not responding"
echo -e "\n=== Log saved: $LOG ==="

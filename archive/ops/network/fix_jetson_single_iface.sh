#!/usr/bin/env bash
set -e

echo "[FIX] Jetson single-interface network"

PRIMARY_IF="enx0015ff125868"
GATEWAY="192.168.1.1"

# bring down conflicting interfaces
sudo ip link set wlan-onboard down 2>/dev/null || true
sudo ip link set wlan-adapter-1 down 2>/dev/null || true

# flush any duplicate IPs
sudo ip addr flush dev wlan-onboard 2>/dev/null || true
sudo ip addr flush dev wlan-adapter-1 2>/dev/null || true

# remove bad routes
sudo ip route del default via 192.168.2.1 dev wlan-onboard 2>/dev/null || true
sudo ip route del default via 192.168.3.1 dev wlan-adapter-1 2>/dev/null || true

# enforce correct route
sudo ip route replace default via $GATEWAY dev $PRIMARY_IF

# clear ARP + route cache
sudo ip neigh flush all
sudo ip route flush cache

echo "[VERIFY]"
ip addr show $PRIMARY_IF | grep inet || true
ip route | grep default || true

echo "[TEST]"
curl -s -o /dev/null -w "%{http_code}\n" http://192.168.1.6:8080 || true

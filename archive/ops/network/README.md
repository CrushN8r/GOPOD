# Jetson ↔ T560 Network Fix

## Topology
- Jetson: 192.168.1.19 (Open WebUI)
- T560:   192.168.1.6  (wire-pod)

## Root Cause
Jetson had duplicate IP across multiple interfaces → ARP/routing ambiguity.

## Fix
Force single active interface on Jetson.

## Usage

### Jetson
```bash
./ops/network/fix_jetson_single_iface.sh

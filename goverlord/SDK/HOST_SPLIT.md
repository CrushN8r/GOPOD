# GOPOD SDK Host Split

## Jetson

Primary GOPOD brain / Scout control host.

Use on Jetson:
- sources/Scout-open-source
- sources/wirepod-vector-python-sdk as Vector SDK reference/control option
- sources/Viccyware as Vector firmware/source reference only unless explicitly building

Jetson owns:
- Moorebot Scout SDK/control
- Goverlord runtime
- Scout body registry
- Scout watchdog/status
- GOPOD orchestration

## Laptop

Cozmo isolation host.

Use on laptop:
- sources/cozmo-python-sdk

Laptop owns:
- Git Repo / Cozmo 1
- Cache PYC / Cozmo 2
- Cozmo keep-awake loop
- Cozmo Wi-Fi/USB adapter flakiness suppressor
- Cozmo body state reporting back to Goverlord

## Rule

Cozmo runtime stays off Jetson to reduce NetworkManager / USB / Wi-Fi complexity.
Scout runtime stays on Jetson.
SDK sources may be mirrored under GOPOD/SDK, but live runtime ownership follows host split.

**Current state (2026-07-30):** every `sources/*` path named above now holds only a
`PLACEHOLDER.md` (re-fetch instructions), not the actual cloned SDK — see
`../README.md`. Re-clone before relying on any of these for reference or building.

# GOPOD

Private Jetson-side orchestration and control-plane repository for GOPOD.

## Layout

- `core/`
  - `goverlord_exec.py`: single execution entrypoint
  - `router.py`: routes validated actions to Codex or T560
  - `validator.py`: enforces the action contract
- `tools/`
  - `vector/`: robot action builders
  - `codex/`: Codex action builders
  - `system/`: diagnostic and utility action builders
- `configs/`
  - runtime config templates, robot config, Codex profiles, Jetson platform config, Wi-Fi backups
- `personas/`
  - `definitions/`: persona JSON definitions
  - `systems/`: persona system packages
- `scripts/`
  - repo-root shell entrypoints for config dump, health, snapshot, and Codex CLI bridging
- `ops/`
  - diagnostics bundles and snapshots
- `archive/`
  - preserved legacy material
- `docs/`
  - repository documentation and inventory

## Configs

- Runtime config lives at `configs/system.json`
- Template lives at `configs/system.template.json`
- Codex profiles live under `configs/codex/`

## Scripts

Run all scripts from the repo root:

- `./scripts/goverlord-config.sh`
- `./scripts/goverlord-pipeline.sh`
- `./scripts/goverlord-snapshot.sh`
- `./scripts/codex-exec.sh --task "<task>"`

## Tools

- Robot tools are in `tools/vector/`
- Codex tools are in `tools/codex/`
- System tools are in `tools/system/`

All tool modules expose `class Tools` with `@staticmethod` methods and return structured action objects only.

## Execution Flow

1. A tool or persona builds a structured action object.
2. `core/goverlord_exec.py` validates it.
3. `core/router.py` routes:
   - `type == "code"` to Codex
   - `type == "robot"` to T560
4. No tool executes T560 or Codex directly.

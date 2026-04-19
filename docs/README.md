# GOPOD

Private working repository for the Jetson-side Goverlord orchestration layer.

This repo is now trimmed down toward a deterministic private base. It keeps the
runtime/control side separate from `GOPOD-PUBLIC`, which remains the public
architecture and contract repo.

## Current Intent

- Keep the orchestration contract aligned with `GOPOD-PUBLIC`
- Preserve the private/public boundary with `GOPOD-PUBLIC`
- Keep the active path deterministic and minimal
- Archive or discard embedded tool wrappers that no longer belong here

## Canonical Active Paths

- `scripts/goverlord_core.py`
  - Main deterministic router for sending actions to the configured T560 HTTP surface
- `config/system.template.json`
  - Required shape for the private machine-local runtime config
- `bin/goverlord-pipeline.sh`
  - Minimal health check for the active Python router
- `bin/codex-exec.sh`
  - Clean Codex CLI bridge used by `goverlord_core.py` for `code` actions
- `bin/goverlord-config.sh`
  - Shows the config files and whether a private runtime config exists
- `bin/goverlord-snapshot.sh`
  - Snapshot-style diagnostics with log capture
- `gomads/registry/personas.json`
  - Persona and robot metadata
- `gomads/personas/persona1/doc_squawkadoodle_gomimion.py`
  - Persona-specific execution wrapper

## Archived Paths

The following files were moved to `archive/` because they were either duplicate,
broken, or superseded by another active file.

- `archive/bin/goverlord-network.sh`
  - Duplicate of the snapshot/status diagnostic flow
- `archive/bin/goverlord-status.sh`
  - Duplicate of the snapshot diagnostic flow
- `archive/scripts/codex_exec.sh`
  - Old Open WebUI-era Codex bridge kept only for reference
- `archive/cdx1_tool.py`
  - Duplicate Codex wrapper from the older embedded-tool approach
- `archive/scripts/codex_tool.py`
  - Embedded Open WebUI tool wrapper; no longer part of the repo base
- `archive/vec_tool.py`
  - Embedded Open WebUI tool wrapper; no longer part of the repo base

## Known Cleanup Rules

- Keep public architecture in `GOPOD-PUBLIC`; keep machine-local execution here
- Do not hardcode unknown private config values into active runtime files
- Prefer explicit config templates over guessed defaults
- Archive before deleting unless a generated file is safe to remove

## Next Recommended Refactors

- Normalize robot naming (`vector1` vs `vec1`)
- Split transport logic from persona-style logic
- Add a small test or smoke-check script for active entrypoints
- Decide whether persona executors stay active or move under `archive/experimental/`

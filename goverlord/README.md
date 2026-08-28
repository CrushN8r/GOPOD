# goverlord

Current GOPOD living machine.

## Map

- `rules/` - flow contracts read alongside `runtime/data_gomad/`'s facts. See
  `rules/README.md`.
- `runtime/` - the pruned, interview-scoped living pathway: Wire-Pod intent
  scaffold, audio verification, robot-safe speech filtering, and the GOPOD
  demo webpage. See `runtime/README.md` for the full layer breakdown.
- `SDK/` - reserved Point B slot. No tracked third-party SDK content;
  GOPOD-authored docs and scripts (`README.md`, `HOST_SPLIT.md`, the two
  setup/sync scripts) are tracked. Sovereign SDK truth lives at the sibling
  `~/crushn8r_git/SDK/`. See `SDK/README.md`.
- `tools/` - `goshot.sh`, a git-recency-scoped code-dump generator (redacts
  secrets, size-capped) for producing a shareable snapshot of active work.
- `wire_pod_overlay/` - a Go build overlay mirroring the `.go` files
  wire-pod's `chipper` binary needs to build with GOPOD's edits, via Go's
  `-overlay` flag, without ever writing to the live `~/wire-pod` checkout.
  See `wire_pod_overlay/build_with_overlay.sh`.

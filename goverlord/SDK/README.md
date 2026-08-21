# SDK/

Point B tree shape slot for SDK material. Holds GOPOD's own host-split
doctrine and setup scripts, plus (locally, gitignored) the vendored SDK
sources they operate on.

## What's tracked here (GOPOD-affected — ok)

- `README.md` — this file
- `HOST_SPLIT.md` — the Jetson/laptop host-split doctrine: which SDK runs
  where (Scout on Jetson, Cozmo isolated to laptop) and why
- `setup_jetson_scout_sdk.sh` — GOPOD's own Jetson-side Scout SDK setup script
- `sync_cozmo_sdk_to_laptop.sh` — GOPOD's own laptop-side Cozmo SDK sync script

These are small, original, GOPOD-authored files describing how GOPOD uses
the vendored SDKs below — not the SDKs themselves. Tracking them is cheap
and they're the actual GOPOD-owned content in this directory.

## What's gitignored here (vendored/generated — not GOPOD's own)

- `sources/` — vendored upstream SDK trees (`cozmo-python-sdk`,
  `Scout-open-source`, `vector-go-sdk`, `vectorx`, `Viccyware`,
  `wirepod-vector-python-sdk`). Re-fetchable, not GOPOD-authored — not
  worth tracking or paying tooling/scan cost on in this repo. **Replaced
  2026-07-30:** each was ~3.8 GB total, its own nested git clone of a
  separately-maintained upstream repo — now each holds only a
  `PLACEHOLDER.md` stating provenance and the `git clone` command to
  re-obtain it. Byte-verified backups of the originals sit in a private
  location outside this repo.
- `.venv-scout/` — local Python virtualenv for Scout SDK work. Regenerate
  via `setup_jetson_scout_sdk.sh`; never commit a venv.
- `tree.txt` — a generated directory listing of `sources/`. Regenerate on
  demand (`tree` / `find`); it only describes what's already gitignored.

**Rule of thumb for anything new landing under `goverlord/SDK/`:** if GOPOD
wrote it, track it. If it's a copy of someone else's SDK or tooling that
can be re-fetched, gitignore it instead of documenting it file-by-file.

**Sovereign SDK truth** still lives at the sibling `~/crushn8r_git/SDK/`
(outside this repo) — the Bingo binary's overlay build (see
`STAGE_RESTORE_001.md`, `STAGE_BINGO_V1_001.md` Task H,
`STAGE_BINGO_V2_001.md` Task F) reads from that sibling location, not from
`goverlord/SDK/sources/`. `sources/` here is a separate, gitignored local
mirror per `HOST_SPLIT.md`'s "SDK sources may be mirrored under GOPOD/SDK,
but live runtime ownership follows host split" — not a build input for
anything in this repo.

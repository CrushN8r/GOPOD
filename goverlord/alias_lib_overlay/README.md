# Alias Lib Overlay — backup, not live source

**This folder is BACKUP/SHOWCASE ONLY. It is never sourced live.** The real, live shell
source is `~/.gopod_alias_lib/` — that directory is where pha0b, phcal, and every alias
described in `tech/alias_play_studio/ALIAS-LIBRARY.md` actually run from, on this
machine, every day. Nothing under this repo path is loaded by `.bashrc`/`.bash_aliases`
or any running shell.

Same pattern `goverlord/wire_pod_overlay/` already uses for the native Wire-Pod
touch-point files: the repo holds a mirror for backup, provenance, and public
readability; the live copy on disk is the one that actually executes.

## Why this exists

`~/.gopod_alias_lib/` was a single point of failure — real working code (176KB+ across
`brobots.sh` alone), exercised daily, not git-tracked, with no backup anywhere. This
folder closes that gap: a scrubbed, read-only mirror committed to public repo history.

## What's mirrored, what isn't

**Mirrored** — the performance/song/cockpit core (the studio instrument): `brobots.sh`,
`core.sh`, `demo.sh`, `wirepod_logs.sh`, `llm.sh`, `chat_capture.sh`, `frame0.sh` (a
retired-stub file, kept for the same "files stay on disk" reason the live copy keeps
it), the phcal/tempo/robot_pick/numpad Python tools, and
`numpad_persona_map_001.json`.

**Also mirrored, 2026-08-13** (the operator's own "include in backup" call, executed):
`openwebui.sh`, `goverlord.sh`, `tools.sh` — the Open WebUI / Goverlord-"suit" lane,
originally held out of the first pass on scope grounds (GOPOD-layer multichat
scaffolding, not robot-performance/song material) but always safe to mirror on secrets
grounds; the operator asked for them in anyway. Copied as-is, secret-clean, no scrub
needed (confirmed: `openwebui.sh` reads `OWUI_TOKEN` from the environment only, never a
literal value; `goverlord.sh` wraps external scripts; `tools.sh` is a dead,
comment-only retirement stub).

**`suits.sh`, mirrored with one scrub**: this file's `owui-env()`/`owui-key-save()`
reference a path into the operator's local secure vault (three lines). The **mirrored
copy only** has that path replaced with a placeholder
(`$HOME/path/to/your/secure/vault`) — the live `~/.gopod_alias_lib/suits.sh` keeps the
real path unchanged, since it's the one that actually has to find the real vault. Every
other line of `suits.sh` copied as-is (the `OWUI_API_KEY`/chat-bridge-JSON/Suit-Changer
logic never named a literal secret to begin with, only the vault path did).

**Excluded, by design**:
- `phcal_last.json` — runtime tuning state (last-used arm/nod/rattle/danger values),
  not source. Regenerates on next use.
- `__pycache__/` — compiled bytecode, regenerates automatically.

## Adapt before running

Every mirrored file carries hardcoded `/home/goverlord/...` absolute paths (this
machine's own home directory) — same caveat `wire_pod_overlay` already states for its
own files. This mirror is not a drop-in install; paths need adapting for any other
machine before anything here would actually run.

## Updating this mirror

There is no automated sync script yet (unlike `wire_pod_overlay`'s
`apply_nongo_files.sh`) — this folder is refreshed by hand, on request, from
`~/.gopod_alias_lib/`. If that becomes a recurring need, a diff-then-copy script
mirroring `apply_nongo_files.sh`'s own shape would be the natural next step, not
invented here ahead of need.

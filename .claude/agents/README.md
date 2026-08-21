# GOPOD lane agents

Three lane workers, one job each — created 2026-07-20 (see
`gopod_notes/older_notes/SUB_AGENT_SET_CREATED_001.md`).

| Agent | Lane |
|---|---|
| `song-lane` | A song's own folder and runner — score, dry-verify, one scoped change at a time. |
| `polisher-lane` | The pha0b cockpit, `~/.gopod_alias_lib/` tooling, and the registry docs describing them (`tech/alias_play_studio/ALIAS-LIBRARY.md`/`ALIAS-SEQUENCER.md`) — teaching-pass style, dry-only; wires one named song into `pha0b` on explicit request. |
| `campaign-desk` | `.claude/skills/` and `gopod_notes/` reports — banks, reconciles, reports. |

## Routing

- The main session routes a request to **one** lane agent at a time — serial for
  any write work, never a parallel pen on the same file across lanes.
- Disk is the shared radio: `gopod_notes/` reports and `.claude/skills/` files are
  how one lane's work becomes visible to the next, not in-conversation memory.
- The operator's live word outranks every agent definition in this directory,
  including this note.
- No agent here ever touches private-lane material — that stays a manual,
  main-session-only task.
- Creating these three agents did not launch any of them — routing starts the next
  time a task actually calls for one.

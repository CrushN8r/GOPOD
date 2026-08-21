---
name: studio
description: Use when unsure which GOPOD skill applies, or asking "what skills exist" / "what's in the studio." Index of every skill in .claude/skills/ with a one-line purpose each. Pure documentation — no logic, no automation; look here first, then invoke the actual skill named.
---

# Studio

Index of GOPOD's working-procedure skills — Claude's own studio discipline around the
songs (reading them, verifying changes, tuning hardware, committing safely, handing off
between sessions). None of these are robot runtime code; none run on Brobot 1 or Brobot 2. They
run on Claude's side of the work, around the songs the robots actually play.

| Skill | One-line purpose |
|---|---|
| `print-score` | Read-only dump of a song's score (`story.md` + `knobs.json`), organized by movement/beat, so the operator can review content without cross-referencing two files by hand. |
| `score-dividers` | Adds/ports labeled `=====` note dividers to a song runner's printed log — plain sandwich around section headers, a short auto-derived tag per note block. Proven on `run_golden_song_001.py`; applies to any song, present or future, on request. |
| `timing-map` | Builds a cumulative timing table and flags genuine gaps from a real `net_video_timing_run_*.json` self-log — never estimates what wasn't measured. |
| `dry-verify` | Proves a code change correct before a live robot run — compile-check, fresh import, a monkeypatched logic test reproducing the real behavior, then the existing test suite — no hardware, audio, or network touched. |
| `hardware-calibrate` | Empirical, one-variable-at-a-time tuning of a robot movement constant against real Vector hardware — change one value, fire it live, stop and get an explicit yes/no from the operator before touching the next. |
| Wire-Pod restart discipline | Not its own skill (folded in here 2026-08-06, was `wirepod-restart-discipline`) — the golden CHECK->CLEAR->START->CONFIRM sequence for restarting Wire-Pod, enforced by one shared function (`restart_wirepod_preflight()`) every caller (`wpr`, `phcal`, the interview runner) routes through. Never write a second copy of the restart logic. |
| `survey-then-commit` | Surveys `git status` first, flags work files for exclusion from the public repo, drafts a commit message grounded in actual session scope, then stops for the operator's go-ahead before staging or committing. |
| `gohandoff` | Writes a dated `gopod_notes/SESSION_HANDOFF_<date>.md` (archiving the prior one, plus other stale top-level reports, into `older_notes/` first) — current-truth state, what was built, open threads, read-first pointers — so the next session reads it first. |
| `goreport` | Enforces the Report and Output File Rule — reports and generated output go to `gopod_notes/`, named `ALL_CAPS_NAME_00N.md`, never inside the git-tracked repo. |
| `playhead` | Conversational "where are we now" recap (Point A / Point 0 / Point B) for reorienting mid-conversation — not the future robot-timing PLAYHEAD persona; this one navigates the chat, not a show. |
| `niche-buzz` | "YOU ARE HERE" flow guide for the niche-buzz campaign — mission, funnel, song shelf, desk ledger, doctrine — one level up from this suite; points at the campaign, not the song work. |
| `gopod-layer` | The dessert map behind niche-buzz — cast, routing, staged hardware ledger for the GOPOD-layer live venue. Gated: not a work order until niche-buzz itself goes evergreen. |
| `web-orbit` | States the security boundary around GOPOD's web orbit (CRUSHN8R domains, hosting, newsletter, shop, social, YouTube) — nothing from the local secure lane ever enters a session's output — and carries the public-safe account roster, YouTube channel/playlist structure, and posting rules; deeper hosting/domain technical detail still lives in `gopod_notes`. |
| `gopolisher` | Three escalating modes for inspecting/auditing/cohering any target — a domain-general inspect/diagnose/cut discipline (era test, comb-through, detangle, trim-for-trajectory), a mechanical drift check between docs/aliases/memory and the real files they describe, and the worker+critic "gauntlet loop" for a full horizontal cohesion pass across the song shelf, alias registry, and repo-root docs. Merged 2026-08-06 from `hairstylist` + `decoupler` + `event-planner`. |
| `goverlord-desk` | The Goverlord desk contract — who the operator is, hard rules, song shelf, scoping discipline — read at the start of every new session, before any other skill. Also holds the campaign-level summoning ritual and the scope/wording discipline checklist (merged in 2026-08-06 from the former `goverlord` and `lane-lines-painter` skills). |
| `alias-mixer` | Tracks the cockpit's growing shared switches (live robots?, reporter gap?, apply phcal tweaks?) — which songs each reaches, the shared movement mechanism underneath, and how to widen or add one. Read before extending a switch to a new song. |
| `studio` | This file — the index itself. |

## Scope

- Pure documentation. No logic, no automation, no code of its own.
- Keep the table current when a skill is added, renamed, or retired — that's the only
  maintenance this skill needs.
- If a task needs one of the skills above, invoke that skill directly rather than routing
  every action through this index — this file is a lookup, not a dispatcher.

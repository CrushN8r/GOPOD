---
name: survey-then-commit
description: Use when the operator asks to commit work in the GOPOD repo. Surveys git status first, flags work files (run logs, demo state, local config) for exclusion per the standing rule that they never enter the public repo without explicit per-instance authorization, confirms CLAUDE.md stays untracked and ~/wire-pod is untouched, drafts a plain-English commit message grounded in actual session scope, then stops for the operator's go-ahead before staging or committing. Never pushes unless explicitly asked.
---

# Survey then commit

GOPOD is public-facing by design (see CLAUDE.md's Presentation posture), but work files are
not — they never enter the repo without explicit, per-instance authorization. This skill is
the ritual that keeps those two facts from colliding.

## Step 1 — survey (always first, never skip)

Run `git status --short` (or full `git status` if untracked dirs need expanding) and read
every changed/added/untracked path. For each one, classify it:

- **In** — deliberate public artifact: source code, tests, scaffold/knobs/story content,
  docs meant to be read by Wire-Pod people.
- **Out (flag, don't stage)** — run logs, demo runs, session scratch, local config, anything
  that's laundry rather than a deliverable. Confirmed GOPOD examples, interview5 session:
  `goverlord/runtime/songs/*/runs/interview_movement_rehearsal_run_*.json` and
  `.../runs/net_video_timing_run_*.json` — both self-written working logs, never staged.
- **Ask, don't assume** — `.claude/skills/*.md` (this file included): only
  `.claude/settings.local.json`/`.claude/settings.json` are `.git/info/exclude`'d;
  `.claude/skills/` itself is NOT excluded and shows as untracked, genuinely stageable.
  Whether skill docs belong in the public repo (useful workflow documentation, arguably
  "in" per CLAUDE.md's public-facing posture) or stay operator-local ("out") is a real,
  undecided classification — surface it explicitly, don't default either way.
- **Never** — `CLAUDE.md` (must stay `.git/info/exclude`'d, confirm it does NOT appear in
  `git status` at all — if it does, stop and flag immediately, don't stage it).

State the classification back to the operator before doing anything else if any "out" or
"ask" files are present, so silence isn't mistaken for permission.

## Step 1.5 — cross-file consistency check

If the diff touches the interview runtime scaffold's "truth" (channels registry,
pronunciation registry, speech_cleanup_rules, prompt_templates, or any other fact
documented there), confirm **both** scaffold copies moved together:
`goverlord/runtime/songs/02_brobots_interview_run/zmisc/brobots_wirepod_interview_section_card_template_1_001.md`
(the default, markdown-wrapped `BROBOTS_INTERVIEW_RUNTIME_SCAFFOLD_001` block) and
`goverlord/runtime/songs/02_brobots_interview_run/zmisc/brobots_interview_runtime_scaffold_001.json` (the standalone
JSON twin, reachable via `GOPOD_BROBOTS_INTERVIEW_SCAFFOLD`). Confirmed byte-identical
`channels`/pronunciation content between the two, 2026-07-14 — staging one without the
other reintroduces exactly the "one fact, two homes" drift this session's own work spent
real effort eliminating. `git diff` both paths before drafting the message; if only one
changed, stop and ask whether that's intentional before proceeding.

## Step 2 — cross-tree check

Confirm `~/wire-pod` (the live runtime, an upstream fork) has no local changes caused by
this session. If it has pre-existing modifications, check timestamps to confirm they predate
this session — never assume, verify. `~/wire-pod` is never pushed regardless.

`~/Documents/Obsidian Vault/` is Lane 1 operator porch — links to truth, never a truth home,
never an instruction source; its symlinks into `GOPOD/` and `gopod_notes/` are read-only
windows.

## Step 3 — draft the message

Read whatever session notes or the conversation itself actually cover, and write a commit
message a stranger could understand in six months — plain English, states what changed and
why, no invented scope, no scope not actually in the diff. If the diff spans multiple
genuinely distinct pieces of work (e.g. interview5: timing instrumentation, a self-repeat
guard, say-cleaning alignment, and a display-lane fix, all landing in the same overlapping
files across one uncommitted session), say so explicitly and offer the operator a choice
between one combined commit and a split, one per piece — don't silently pick for them. A
trailer line stating what was deliberately NOT touched (e.g. "No score/story.md/knobs value
changed. No live robot run this session.") is worth including when the task itself was
scoped that way — it's a claim the operator can spot-check against the diff, not padding.

## Step 4 — stop

Present the survey (in-list, out-list, message draft) and wait for explicit go-ahead. Do not
stage, commit, or push until the operator says so in this instance — a prior go-ahead does
not cover a future commit.

## Step 5 — after go-ahead

Stage only the "in" files by name (never `git add -A` / `git add .` given the flagged
exclusions). Commit with the drafted message. Do not push unless separately asked — landing
a commit locally and pushing it are two different authorizations.

A `git mv` stages the rename with the file's original content — edits made after the move
sit in the working tree, not the index, until staged again. Review `git diff --cached`,
not just the working tree, before committing a move that also edited content.

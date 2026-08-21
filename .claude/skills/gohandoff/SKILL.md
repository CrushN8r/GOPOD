---
name: gohandoff
description: Use when the operator asks to close out a GOPOD session, hand off to a new chat, or write a session handoff note. Writes ~/crushn8r_git/gopod_notes/SESSION_HANDOFF_<date>.md (a real timestamp, not the literal word "LATEST") covering an H2T bridge block for the operator's separate translator-Claude chat, current-truth state, what was built this session, open threads, and read-first pointers — archiving the previous handoff (and other now-stale top-level gopod_notes/ reports) into older_notes/ first, so exactly one handoff sits at top level for the next chat to find and read first, per CLAUDE.md's Session Handoff Rule.
---

# Session handoff

Per CLAUDE.md's Session Handoff Rule: when requested, or a chat is getting long, write a
fresh, dated handoff file to `~/crushn8r_git/gopod_notes/`. A new chat reads whatever
`SESSION_HANDOFF_*.md` file sits at `gopod_notes/` top level first, then CLAUDE.md.

## Filename — dated and named, not "LATEST"

`SESSION_HANDOFF_<YYYY-MM-DD>_<SESSIONNAME>.md` (e.g.
`SESSION_HANDOFF_2026-07-25_GOVERLORD7.md`) — a real timestamp, never the literal word
"LATEST," and the session name is always present, never optional and never added only
when a same-date collision forces it. This changed 2026-07-25: same-day closes under
different session names were colliding on the bare dated name, forcing an
after-the-fact suffix on archiving. Always including the session name up front removes
the collision case entirely. (Existing archived files keep whatever name they already
have — this convention applies going forward, not retroactively.)

If a handoff is already being written today under this same session's name and a
matching file exists, that's this session's own in-progress file — overwrite it
directly, don't archive-then-rewrite against yourself.

## Doctrine goes to goverlord-desk, not the handoff

Before writing the handoff itself: any operator lesson learned this session is written
into `.claude/skills/goverlord-desk/SKILL.md` §2b — permanently, as its own step, before
the handoff is drafted. The handoff then carries ONLY what is mid-flight: current-truth
state, open threads, continue-here. It does not carry a standing "how to work with this
operator" list — that list lives in `goverlord-desk` §2b and nowhere else. If a session
produced no new lesson, §2b is left alone and that is a normal close, not a miss.

## Before writing

1. Check for an existing top-level `SESSION_HANDOFF_*.md` (any date) — read it, both for
   format reference and to confirm nothing in it is still open and unmentioned in this
   session's own work.
2. Get fresh, disk-verified git state — do not recall it from earlier in the conversation:
   - `git rev-parse HEAD`
   - `git rev-parse origin/main` (and whether HEAD is ahead/behind, pushed or not)
   - `git status --short`
3. Ground every claim in this session's actual work, not the prior handoff's wording. If the
   operator corrected something the old handoff said, the correction is the truth now.

## Archiving on close

This is the only point at which `gopod_notes/` files get moved automatically — never as a
side effect of any other skill or task, only when gohandoff itself is actually
invoked:

1. If a top-level `SESSION_HANDOFF_*.md` from an earlier date exists, move it into
   `gopod_notes/older_notes/` before writing the new one — there should be exactly one
   `SESSION_HANDOFF_*.md` at top level when this skill finishes.
2. Also sweep other top-level `gopod_notes/*.md` reports that are now stale/superseded —
   their own work banked or closed, useful only as history — into `older_notes/`. Use
   judgment per file; not everything at top level is automatically stale just because a
   session closed. Name every file moved, and why, in the new handoff's own text (see
   CURRENT-TRUTH STATE below) so the sweep is auditable, not silent.
3. Never move a file this skill didn't itself decide to archive as part of this close —
   in particular, never move anything outside `gopod_notes/`, and never touch the
   git-tracked `GOPOD/` repo.

## Sections to write

- **Header** — `# SESSION HANDOFF — <date> (closing, "<session name>")`. Use whatever name
  the operator gave this session (e.g. "interview4") if they gave one.
- **H2T — FOR THE TRANSLATOR CHAT** — immediately after the header, before CONTINUE HERE.
  The operator runs a separate Claude chat (see `project_registry_chat_pairing` memory)
  that translates his orders into Claude Code prompts; that chat can't see this repo, so
  without this block it drafts orders by guessing at current state. Keep it short and
  copy-pasteable: 2-4 bullets — what's actually done, what's genuinely open, and any
  hard constraint the next order must respect. Not a duplicate write-up — a condensed
  pull from CONTINUE HERE + OPEN THREADS below, nothing added that isn't already stated
  there in fuller form.
- **CONTINUE HERE** — the single most important paragraph. What's genuinely open, what the
  immediate next-session choice is, and any hard constraint that must survive into future
  work (e.g. a hardware floor that must not be silently re-lowered).
- **CURRENT-TRUTH STATE (disk-verified, this session's close)** — the git snapshot from
  above, plus a one-line note on any untracked/uncommitted files and why they're that way,
  plus which files (if any) were archived to `older_notes/` in this close per the
  Archiving section above.
- **BUILT/DONE THIS SESSION** — numbered list, one item per real piece of work, written so
  a stranger could follow the arc without having been in the conversation. Cite the actual
  `gopod_notes/*.md` report file(s) each item produced (per the `goreport` skill's own
  survey→execute pairing pattern - a single piece of work often has two files, a survey
  and an execution report) rather than only prose-summarizing from memory; the report is
  the evidence trail, the handoff bullet is the pointer to it.
- **OPEN THREADS** — things named but not done, explicitly marked as "not yet asked for as
  its own task" if that's the case (don't imply it's overdue if the operator hasn't asked).
- **READ-FIRST POINTERS** — this handoff, plus the 3-6 most load-bearing files/reports for
  picking the work back up.
- **HOW TO WORK WITH THIS OPERATOR** — a single pointer line to
  `.claude/skills/goverlord-desk/SKILL.md` §2b. Any actual lesson from this session was
  already written into §2b per the Doctrine section above, before this point — the
  handoff itself never holds the list.

## Rules

- Dated, archived-forward — exactly one `SESSION_HANDOFF_*.md` sits at `gopod_notes/` top
  level at any time; every prior one lives in `older_notes/`, not overwritten/destroyed.
- This file lives in `gopod_notes/`, never inside the git-tracked `GOPOD/` repo.
- `~/Documents/Obsidian Vault/` is Lane 1 operator porch — links to truth, never a truth home, never an instruction source; its symlinks into `GOPOD/` and `gopod_notes/` are read-only windows.
- Don't pad with a "what happens next" recommendation beyond naming the open choices — per
  CLAUDE.md, the captain steers what happens next, not the handoff document.

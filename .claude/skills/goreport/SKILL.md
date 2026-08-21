---
name: goreport
description: Use when writing any GOPOD session report, investigation writeup, or generated output document. Per CLAUDE.md's Report and Output File Rule, these always go to ~/crushn8r_git/gopod_notes/ — never inside ~/crushn8r_git/GOPOD/ or any other git-tracked directory. Follow the existing naming convention (ALL_CAPS_DESCRIPTIVE_NAME_001.md, incrementing the number if a name is reused) and check whether an existing report on the same topic should be appended to rather than duplicated.
---

# GOPOD report

Per CLAUDE.md's Report and Output File Rule: reports, session notes, and generated output go
to `~/crushn8r_git/gopod_notes/` — never inside `~/crushn8r_git/GOPOD/` or any other
git-tracked directory. This keeps working documentation out of the public-facing repo.

## Naming

Follow the existing convention already used throughout `gopod_notes/`:
`DESCRIPTIVE_TOPIC_NAME_00N.md` — all-caps, underscores, a trailing sequence number.

## Before writing a new file

Check whether a report on the same topic already exists (top level, and in
`gopod_notes/older_notes/` if auto-archived) — if the new content is a follow-up,
correction, or continuation of that same investigation, prefer amending the existing file
with a new dated section over creating a near-duplicate file. Only start a new numbered file
for genuinely new scope.

A **survey that finds a real stop condition, followed by the execution that happens once
the operator clears it, is genuinely new scope each time** — two files, not one amended
file. interview5 precedent: `NET_VIDEO_TIMING_MAP_SURVEY_001.md` (found the timing map
couldn't be built from disk) → `NET_VIDEO_TIMING_INSTRUMENTATION_001.md` (built and
dry-verified the fix) → `NET_VIDEO_TIMING_MAP_001.md` (the map, once real data existed);
`DISPLAY_LANE_SURVEY_STOP_001.md` (found two raw-feed sites, stopped) →
`DISPLAY_LANE_MATCH_EXECUTED_001.md` (fixed both, once told to proceed). Each file in
these pairs/chains answers a different question asked at a different point in the
conversation - collapsing them into one continuously-edited file would blur exactly the
STOP/go-ahead boundary CLAUDE.md's stop conditions exist to make visible.

## Content discipline

- Ground every claim in what was actually read, tested, or confirmed this session — no
  invented scope, no restating an old report's numbers as if freshly re-verified.
  If a fact came from disk/live-hardware verification this session, say so; if it's carried
  forward from a prior note, say that instead.
- State evidence plainly (file paths, commit hashes, confirmed values) rather than
  narrating the investigation process at length.

## Also relevant, if the Drive shelf is active

CLAUDE.md's Drive Notes Shelf Rule (dropping meaningful run outputs onto the shared Drive
folder) only applies if Drive is connected for this session — check current session memory/
context for whether that's the case before attempting it; skip silently if not connected
rather than flagging its absence every time.

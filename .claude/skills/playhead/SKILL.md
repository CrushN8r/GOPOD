---
name: playhead
description: Use when the operator asks to recap, reorient, or zoom out after the conversation has drifted into the weeds — "where are we now," WAWN, "point A / point 0 / point B," "catch me up," or any equivalent ask to be relocated on the map. This is the conversational navigation motion (Where Are We Now?) — NOT GOPOD's future PLAYHEAD persona (robot performance timing/orchestration). This skill navigates the conversation, not a show.
---

# Playhead

The pha0b / WAWN zoom-out move: lift out of whatever detail the conversation is currently
in and re-see the whole board — where this thread actually started, where it truly stands
right now, and where it's headed next. Used when the operator has lost the thread, asks to
be reoriented, or the conversation has run long enough that a plain recap beats another
answer buried in the weeds.

This is the conceptual seed for GOPOD's future PLAYHEAD persona (performance timing/
orchestration for the robots) — kept deliberately separate here: this skill navigates the
conversation, the persona will one day navigate the show. Do not conflate the two.

## Output shape: Point A / Point 0 / Point B

- **Point A — origin.** What this thread actually started as, in one or two sentences —
  the starting frame the operator would recognize, not the whole history.
- **Point 0 — current truth.** Where things actually stand right now. Ground this in
  what's verified this session (disk state, git status, code just read or changed) rather
  than recalled from earlier in the conversation, wherever that's cheap to check — a WAWN
  asked mid-drift is often asked precisely because something moved and the operator isn't
  sure what's still true.
- **Point B — next dock.** The next reachable target, not a wishlist. If a decision or
  fork is genuinely waiting on the operator, name it here plainly — without opening a new
  question the operator didn't ask.
- **Pinnacle Point B — the horizon.** A fourth rung, not a substitute for Point B: the
  glamour-ready end state being aimed at, distinct from Point B's next reachable move.
  Point A is where the thread started, Point 0 is what's verified now, Point B is next,
  Pinnacle Point B is the far horizon. The gap between Point 0 and Pinnacle Point B isn't
  a defect to hide — name it, label it as trajectory, and present it as "coming soon."
  When working material must be shown, reveal is curation, not exposure: an intentional
  last-known-golden version gets prepared for presentation while the raw working files
  stay fenced.

## Rules

- Terse, matched to the operator's own length in this conversation — a WAWN mid-task gets
  three tight lines, not three paragraphs.
- No open-ended closing question ("what would you like to do next?"). Point B names the
  fork if one exists; the operator steers from there on their own terms, per CLAUDE.md's
  own "the captain steers" rule.
- Ground Point 0 in verification, not memory, whenever the claim is checkable — a stale
  Point 0 defeats the entire purpose of the motion.
- This is a recap, not a status report — skip exhaustive detail; that level of depth is
  what the `goreport` skill's own output is for.

## Scope

- Read-only against the conversation and, where Point 0 needs it, disk/git state — no
  edits, no commits, no code changes triggered by this skill itself.
- Not the PLAYHEAD persona. That is a future, separate, robot-facing build. This skill
  never drives Vector hardware or a show's timing; it only reorients the conversation.

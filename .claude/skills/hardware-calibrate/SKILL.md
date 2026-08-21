---
name: hardware-calibrate
description: Use when tuning any GOPOD robot movement constant (motor speed, hold duration, amplitude) against real Vector hardware. Empirical, one variable at a time — change a single constant, fire it live, then stop and ask the operator a plain yes/no via AskUserQuestion before changing anything else. Never batch multiple changes between hardware checks. Record confirmed-working and confirmed-dead values as a permanent warning comment next to the constant. Never lower a previously-established floor without a fresh, live, operator-witnessed re-confirmation.
---

# Hardware calibrate

Robot movement constants (motor speed, hold duration, amplitude, lead-in speed) cannot be
verified from logs or HTTP response codes — an HTTP 200 from `/api-sdk/move_head` or
`/api-sdk/move_lift` does not mean the motor physically moved. Only the operator's eyes,
watching the real robot, can confirm a change worked. This skill is the discipline for doing
that safely.

## The loop

1. Change exactly **one** constant to exactly **one** new value.
2. Fire it live on real hardware (via whatever alias/tool already exists — reuse the control
   song or rehearsal tooling, don't build a new firing path).
3. Before firing, send an explicit heads-up ("About to fire `<CONSTANT>` = `<value>` on
   `<robot>` — watch its `<body part>` now") so the operator knows what to watch for and
   when.
4. Stop. Ask a plain yes/no via `AskUserQuestion` — did it visibly work, yes or no. Do not
   proceed to the next value until this answer is in hand.
5. Record the result (working or dead) before moving to the next value.
6. If the operator gives a live redirect that departs from the planned bisection (e.g. "stop
   testing, set it to X, tighten a different variable instead") — follow the redirect. It
   outranks the original plan.

## Never batch

Do not change two constants between hardware checks, and do not assume a value "probably"
works because a nearby value did — each value gets its own live confirmation.

## Recording the result

Once a floor or working value is confirmed, write it as a permanent, prominent comment
directly next to the constant in the source file — not just in a session report. State both
the confirmed-dead range and the confirmed-working value, and add an explicit instruction not
to lower it again without a fresh, live, operator-witnessed re-confirmation. See
`run_robot_control_song_001.py`'s golden-truths block and `NOD_TEST_SPEED`/
`NOD_GESTURE_SPEED` comments for the established pattern.

## Scope

- This is about physical movement constants specifically — not timing/pacing values that
  come from config (e.g. `between_exchange_pause_seconds`), which are read from disk, never
  invented or bisected.
- A previously-established floor is a hard constraint for all future sessions until
  re-confirmed live — do not treat an old report's number as safe to lower "for smoothness"
  without repeating this loop.
- **Not this skill**: a code/software change (a new guard, a timing instrumentation pass,
  a text-cleanup fix) that doesn't move a motor. That needs proof it's correct, not proof a
  robot moved - see `dry-verify` instead. The two are complementary, not overlapping: this
  skill is for when only a live robot's own physical motion can confirm success; dry-verify
  is for everything provable without one.

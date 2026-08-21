# Robot Control Song — self-check, out loud

Knobs: [knobs.json](knobs.json)

One robot (Brobot 1), one straight line of notes: connect, say it's connected,
arm test, head nod, a real weather fetch, say it's good, exit. Every physical
note gets a spoken narration line right before it plays, and a spoken failure
line if it doesn't come back clean. This is a test that talks, not a silent
script.

`> TEXT:` is spoken verbatim before the note plays (or is the note itself, for
a plain `say`). `> FAIL:` is spoken only if that note's own hardware call
didn't come back clean.

## STEP connect
> TEXT:

## STEP say_connected
> TEXT: I'm connected. Loading next test.

## STEP arm_test
> TEXT: Testing my arm.
> FAIL: My arm didn't respond.

## STEP arm_test_done
> TEXT: Arm test done. Loading next test.

## STEP head_nod
> TEXT: Testing my head nods.
> FAIL: My head didn't respond.

## STEP head_nod_done
> TEXT: Nods test done. Loading next test.

## STEP weather
> TEXT:

## STEP say_good
> TEXT: Weather test done. I'm good now. Exiting.

## STEP exit
> TEXT:

## Troubleshooting

Field-proven, 2026-07-15. If the runner throws `WIREPOD_LOG_MIRROR_ERROR` or a bare
`TimeoutError` on connect, `wpr` (a Wire-Pod service restart alone) was **not** enough to
recover on its own. What actually worked, in this order:

1. Power-cycle both robots (reboot).
2. Re-pair them with Wire-Pod.
3. Restart Wire-Pod (`wpr`).

Confirmed live, not a guess - keep this order; a partial recovery (e.g. `wpr` alone, or
re-pairing without a reboot first) did not clear the error.

### Animation-by-name landmine (tripwire, not a confirmed bug here)

Vector SDK fact: requesting a trigger/clip by NAME-AS-TEXT makes the SDK download the
whole animation catalogue first to look it up. The catalogue is huge - on a modest
machine this can stall up to ~2 minutes, and the animation then silently never plays. No
error, just a dead beat where a gesture should have been.

Fix, if this is ever the actual cause: build the animation as a proper object first and
pass THAT (skips the catalogue fetch); for triggers, load the small trigger list once at
startup and look up from it, never by raw name per call.

**Flag**: this is a direct-Vector-SDK-path hazard. GOPOD's interview/bait songs go
through Wire-Pod's `/api-sdk/say_text` HTTP surface, a different transport - confirm
whether the current arm/nod gestures actually touch the vulnerable path before treating
any stall as this. Captured here as a symptom to recognize, NOT a confirmed present bug
in this codebase.

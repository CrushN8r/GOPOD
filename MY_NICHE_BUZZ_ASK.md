# Niche-Buzz Ask

From @crushn8r, the operator.

A third door, smaller and wider than the other two. [MY_GOPOD_ASK.md](MY_GOPOD_ASK.md) is
the technical-collaborator ask — the cockpit, the code. [MY_GOPOD_OPS_ASK.md](MY_GOPOD_OPS_ASK.md)
is the ops/social ask. This one's no-commitment: help me sharpen a real, working piece of
tooling, right now, no robot required.
Want the case for why any of this is worth your time first? [UNFAIR_ADVANTAGES.md](UNFAIR_ADVANTAGES.md).

## The bait — anyone with a keyboard

I just built a keyboard grabber — a small tool that finds your keyboard, grabs it for
exclusive input, and releases it cleanly no matter what happens (crash-tested, three ways:
grab-then-release, grab-then-simulated-crash, and a post-crash re-grab to prove the device
was never left stuck). It works on my machine. It's only ever been tested on my one
keyboard.

Different keyboards genuinely behave differently — how they show up to Linux, whether
detection picks the right device out of a pile of decoys (mine had three false positives
before I found the fix: a webcam, a USB audio device, and `gpio-keys` all flag themselves as
keyboard-capable, none of them the actual keyboard). If you've got a keyboard and a
terminal, you can help. Clone the repo, run it, tell me if it finds your keyboard correctly
and if it grabs and releases clean. That's the whole ask. No robot, no GOPOD setup beyond
the clone.

`goverlord/runtime/gopod_layer/web_display/gopod_demo_8011/gopod_keyboard_grabber_001.py`
— `--list-only` if you just want to see detection without grabbing anything.

## The net — if you're the kind of person who reads this far

If the keyboard grabber's the kind of thing that makes you want to poke around the rest of
it: phcal is the calibration bench — one primitive at a time, arm, nod, wake, sleep, and now
a real startup probe that shapes itself to whatever robots actually respond. pha0b is the
performance front door — whole songs at a time. Both are real, live, committed, running code
today, not concept docs. If you've got opinions — good or bad — on how either one's built, I
want to hear them.

## One honest footnote — if you happen to own 3 or more Vectors

phcal's own robot-detection was built to handle any number of candidates, but it's only ever
been tested against two, because that's all I have. If you're one of the rare people with
three or more Vectors, there's one specific thing only you can verify: whether the mode
logic actually holds past two robots, or quietly assumes two somewhere I didn't catch. This
is not the headline ask — most Vector owners have one robot, some have two, and this doc's
real door is the keyboard grabber above. But if this is you, I'd genuinely want to know.

## Reach out

Subscribe to the CRUSHN8R CREW'd newsletter at `crushn8r.ca` — or reach me directly at
`crushn8r@gmail.com`.

---

## GOPOD YAHMM (You Are Here Mall Map)

Three folders, three maps — pick where you want to go:

- [web/README.md](web/README.md) — the content engine: pillars, wordplay, aha moments, newsletter
- [tech/README.md](tech/README.md) — the songs, the studio tooling, Wire-Pod integration
- [life/README.md](life/README.md) — the philosophy, teaching, and lessons learned

**Main docs**
- [README.md](README.md) — what GOPOD is and how it's built
- [MY_GOPOD_ASK.md](MY_GOPOD_ASK.md) — the operator's own ask — what's built, where the line is, what kind of help this needs
- [MY_GOPOD_OPS_ASK.md](MY_GOPOD_OPS_ASK.md) — the operator's ops ask — social, sites, and content, a different role than the technical one
- [TRAJECTORY.md](TRAJECTORY.md) — the planned arc, Point A to the pinnacle, honestly labeled built vs. aim
- [UNFAIR_ADVANTAGES.md](UNFAIR_ADVANTAGES.md) — the case for why this is worth your time — what GOPOD has that most projects don't

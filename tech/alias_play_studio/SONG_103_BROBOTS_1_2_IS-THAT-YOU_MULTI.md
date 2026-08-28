# GOPOD Is That You — Multi

> Say a name. Watch which brobot answers first — the real mix-up, live, not scripted.

**Naming note, exception:** this song runs on brobots wire-pod layer hardware, but its
content is deliberately framed as "leaked GOPOD-layer test footage" (per `niche-buzz`'s own
funnel role for this song). Because it's testing/previewing the GOPOD layer by design,
"Doc"/"Pip" is the correct, in-character naming throughout this doc — unlike other
wire-pod-layer songs, which use Brobot 1/Brobot 2.

---

## What Is That You — Multi is

The golden, two-robot original: hold either KP1 or KP2, Doc or Pip, and watch the real
mix-up happen live. Say a name, one brobot answers — decided live, a canned line or a
live Ollama reply. Nothing here is scripted; the whole reason to shoot it is that it
isn't.

Wired as `104_gopod_is_that_you_multi/`, `pha0b` keyword `itsyou-multi`. This is the
gold — the original, unscoped version everything else in this song family traces back
to.

Bookends the exact same live alias (`is-that-you()` in `~/.gopod_alias_lib/demo.sh`) the
single-robot version does — the writer (`gopod_ptt_chat_writer_013.py`) already treats
KP1/KP2 as fully independent, self-contained key handlers, so the two-robot mix-up was
never anything but two people's own separate key-presses landing in the same session.
See this song's own `story.md` for the full mechanical breakdown.

There's also a single-robot version of this same idea:
[IS-THAT-YOU SINGLE.md](SONG_103_BROBOTS_1_2_IS-THAT-YOU_SINGLE.md) — hold only Doc's
key, one robot, the on-ramp made visible.

## Honest state

Wired into `pha0b` (`itsyou-multi`), dry-verified both directly and through the menu
picker — **not yet live-run**. No actual Is That You recording session has been shot
through this bookend yet. This is exactly where the jewel stands today: real,
live-capable, fully wired — waiting on the take, not on more building.

---

## History — the earlier scripted reel

Before this split, "Is That You" also named something different: a short, four-line
**scripted** demo reel — fixed dialogue, no live capture at all. Archived 2026-08-12 to
`zzz_archives/102_brobots_cross_persona/` once the real, live PTT bookends made a
scripted stand-in redundant. Still reachable, repointed not retired, via the `mixup`
`pha0b` keyword — kept for the record, not deleted.

**The four lines**, for reference:

**Doc, self-confirming:** *"Doc here. Yes. Try to keep up."*
**Doc, naming the mix-up:** *"That is Pip's lane. Wrong robot, right confusion."*
**Pip, self-confirming:** *"Yep. Pip here. I think that was my cue."*
**Pip, naming the mix-up:** *"Hey Doc, is this when I ask for emails?"*

Built from `persona_awareness_reply()`'s own four canned cross-persona lines — the one
fixed-content piece of an otherwise fully interactive writer (keypress wait, live mic
capture, Vosk transcription, live Ollama fallback — none of that fits a scripted step
list, so none of it made the port). Played back to back, once, as a demo of what the
live reply sounds like — not a re-enactment of how the real thing ever runs, since in the
live system exactly one of the four lines fires per utterance, decided live.

**Naming history**: this reel first lived at `103_gopod_is_that_you/`, renamed
2026-07-31 to `brobots_cross_persona` (folder `102_brobots_cross_persona`) once it
collided with the live PTT alias's own `is-that-you` name — two different things sharing
one name. The live demo kept `is-that-you`; the scripted reel moved to a non-colliding
one. That same `103_gopod_is_that_you` slot was later reused (2026-08-08) for the live
bookend songs, split (2026-08-18) into `_single`/`_multi` runtime folders, and finally
into this doc and its sibling (2026-08-18, same day) to match.

---

> From Doctrine Barfallonyou
> Lesson! Knowing your own name is easy. Knowing when someone else answered to it first — that's the interesting part.
> Boom! Done! Class Dismissed!
> — Doc Squawkadoodle

---

## GOPOD YAHMM (You Are Here Mall Map)

Part of GOPOD — see [tech/README.md](../README.md) for everything else in this folder, or [the root map](../../README.md) for the rest of GOPOD.

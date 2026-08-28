# GOPOD Is That You — Single

> Say a name. Get a real answer back — one robot, live, not scripted.

**Naming note, exception:** this song runs on brobots wire-pod layer hardware, but its
content is deliberately framed as "leaked GOPOD-layer test footage" (per `niche-buzz`'s own
funnel role for this song). Because it's testing/previewing the GOPOD layer by design,
"Doc"/"Pip" is the correct, in-character naming throughout this doc — unlike other
wire-pod-layer songs, which use Brobot 1/Brobot 2.

---

## What Is That You — Single is

The single-bot jewel: hold only KP1 (Doc), one robot, the on-ramp made visible. Hold the
key, say a name, get a real answer back — decided live, a canned line or a live Ollama
reply. Nothing here is scripted; the whole reason to shoot it is that it isn't.

This is the same real thing a single-Vector owner could run themselves, per
[QUICKSTART.md](../SINGLE_BOT_QUICKSTART.md) — proof that "your one robot, answering
live" isn't a demo trick, it's the real mechanism, on camera.

Wired as `103_gopod_is_that_you_single/`, `pha0b` keyword `itsyou-single`. Bookends the
exact same live alias (`is-that-you()` in `~/.gopod_alias_lib/demo.sh`) the two-robot
version does — the writer (`gopod_ptt_chat_writer_013.py`) already treats KP1/KP2 as
fully independent, self-contained key handlers, so holding only one key already ran
single-robot before this split ever existed. "Single" scopes the recording, not the
code. See this song's own `story.md` for the full mechanical breakdown.

There's also a two-robot version of this same idea:
[IS-THAT-YOU MULTI.md](SONG_103_BROBOTS_1_2_IS-THAT-YOU_MULTI.md) — hold either key, Doc
or Pip, watch the mix-up happen live.

## Honest state

Wired into `pha0b` (`itsyou-single`), dry-verified both directly and through the menu
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

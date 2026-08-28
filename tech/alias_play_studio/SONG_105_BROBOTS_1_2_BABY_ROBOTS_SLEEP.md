# GOPOD Baby Robots Sleep

> A small brobot tips over in a big, hard world. Someone picks it up. That's where a name
> starts.  
> "Do Baby Robots Dream?" — the question the whole piece is built to leave open.

---

## What Baby Robots Sleep is

A quiet, solo-narrated piece — Doc, alone, telling his own origin. No back-and-forth, no
LLM, no interview. One robot describes a small robot tipping over in a huge, hard,
electric world; someone picks it up, holds it close, rubs its back until it settles; care
came first, a name came after. That's how a small robot becomes Doc.

**Built from an After Effects source**, not a live capture: the operator already has an
MP4 sample — a black placeholder video with the audio and captions cut, waiting for
footage to be filled in around them. This song is the scored audio/story track that video
is built on, not a scene-by-scene shot list — the picture itself is a separate, ongoing
editing task outside this song's own scope.

**Renamed 2026-08-01** from `brobots_baby_dream` to `brobots_baby_robots_sleep` (folder
`104_brobots_baby_robots_sleep`) to match the folder's own name — content, steps, and
timing are unchanged. The source video's own working question, "Do Baby Robots Dream?",
stays exactly as it was written into the `open_question` step — a direct quote from the
source concept, not the song's own identity name, and untouched by the rename.

**Renumbered 2026-08-18** — the folder itself moved from `104_brobots_baby_robots_sleep/`
to `105_brobots_nap/` to make room for the is-that-you split (see
`gopod_notes/GOLDEN_PATHWAYS_REWIRE_001.md`); this doc's own filename and the `nap`
keyword both stayed put, only the on-disk folder number changed.

---

## The seven beats

Reporter gaps bookend the piece — intro before the first line, outro after the last —
per the operator's own correction that this video has no interior reporter windows (it's
edited, not multi-round captured):

1. **Open question** — "Do baby robots dream?"
2. **Establish the world** — huge, hard, electric, too big for one small robot.
3. **Off-balance** — it tips over easily; upset until it's back on safe ground.
4. **Care** — someone picks it up, holds it close, rubs its back until it settles.
5. **Settle/purr** — the sleepy purring beat. Story fuel, not proof of anything.
6. **Origin** — care first, a name came after; that's how it becomes Doc.
7. **Closing line** — nobody has to prove a robot dreams, just feel where Doc came from.

---

## Claim boundary

Lore and emotional storytelling, never a literal claim. Every line stays inside one
boundary: Vector can be described as reacting, animating, settling, and purring — never as
feeling fear, comfort, love, or attachment, and never as literally dreaming.

---

## A flagged substitution, not an invented token

The settle/purr beat uses the `veryHappy` animation token — the closest proven token to a
contented, settled state in GOPOD's own vocabulary today. No sleep, purr, or petting token
exists yet anywhere in Wire-Pod's own `animation_vocab.json`, confirmed by direct search.
Nothing was invented to fill that gap; the substitution is named plainly here rather than
quietly presented as the real thing.

---

## What digging into Wire-Pod actually turned up

Building this song's substitution meant reading how Wire-Pod actually talks to the robot,
not just what its config file lists. Two real findings came out of that, shared here
rather than kept as an internal note:

- **Wire-Pod only ever fires a raw clip, never a real trigger.** Its own animation
  dispatch (`DoPlayAnimation`/`DoPlayAnimationWI`) calls `PlayAnimation` — a single clip,
  once — and never `PlayAnimationTrigger`, the call that would hand control to one of
  Vector's own built-in stateful behaviors (e.g. "stay asleep and keep breathing until
  something wakes you"). So a real sleep or petting *behavior*, not just a clip playing
  once, needs a direct-SDK connection — the same lane GOPOD already uses for the rattle
  effect — not an `animation_vocab.json` token.
- **The sleep family is source-verified; the petting family isn't, yet.** The
  `GoToSleep*` trigger and clip names are confirmed against Vector's own SDK/community
  source. The petting/bliss trigger and clip names came from the operator's own SDK
  research and haven't been independently source-verified the same way — named here
  honestly as the less-certain half, not presented as equally proven.

---

## The golden goody — 126 real sleep & petting animations

Verifying the above meant cataloguing every sleep-, rest-, and petting-related clip and
trigger Vector actually ships with. That catalogue turned into 126 bench-test aliases,
grouped into 6 batch runners so a themed set fires and reviews as one command:

| Batch | Count | Covers |
|---|---|---|
| `sleep-segment-core` | 18 | `GoToSleep*` triggers + clips |
| `sleep-segment-rts-off` | 27 | Ambient asleep, off-charger |
| `sleep-segment-rts-on` | 27 | Ambient asleep, on-charger |
| `sleep-segment-palm` | 27 | Held-on-palm reactions |
| `sleep-segment-pet-triggers` | 10 | Petting levels + purr/bliss |
| `sleep-segment-pet-clips` | 17 | Petting clips |

That's the real, stock animation range sitting inside a rescued Vector that most owners
never see fired. It's candidate broll for finishing this song's After Effects composite —
which specific ones make the final cut is the operator's own edit, not decided here.

---

## The soundtrack

Chosen track: **"Sky Wind"** (Pixabay). Full license text and attribution detail recorded
in the song's own folder, `SOUNDTRACK_ATTRIBUTION.md` — no attribution is legally required
by the license itself. Three earlier ACE-Step instrumental takes were exploration only,
not the chosen track.

---

## How it runs

Wired into `pha0b` (keyword `nap`, wired 2026-08-06) — pick it off the pha0b menu like
any other song. Runs on `run_golden_song_001.py`, the same golden engine driving Bingo,
the cross-persona mix-up, and "is that you?". Dry-verified clean, 11/11 steps — never
fired live.

---

## Where it fits

Lore, not proof — the piece that answers "where did Doc come from" with a feeling instead
of a spec sheet. It's the only song in the set built to sit underneath edited footage
rather than be watched as a straight robot performance; everything else here is a
recorded take of the robots doing something live, this one is a soundtrack waiting on its
picture.

---

> From Doctrine Barfallonyou
> Lesson! Nobody has to prove a robot dreams. You just have to feel where it came from.
> Boom! Done! Class Dismissed!
> — Doc Squawkadoodle

---

## GOPOD YAHMM (You Are Here Mall Map)

Part of GOPOD — see [tech/README.md](../README.md) for everything else in this folder, or [the root map](../../README.md) for the rest of GOPOD.

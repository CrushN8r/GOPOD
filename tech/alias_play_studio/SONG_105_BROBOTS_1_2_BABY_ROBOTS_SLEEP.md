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

Two doors. Pick the one that fits how far in you want to go — nothing here needs a set order beyond that.

### Door 1 — New here? Explore GOPOD wide

Plain language, first look, no background needed — for a newcomer, human or AI.

**Start here**
- [README.md](../../README.md) — what GOPOD is and how it's built
- [AWAKEN.md](SONG_00_BROBOTS_1_2_AWAKEN.md) — watch first: a brobot wakes, checks itself, greets you
- [QUICKSTART.md](../SINGLE_BOT_QUICKSTART.md) — talk to your own Vector, one robot, no alias studio needed
- [MY_NICHE_BUZZ_ASK.md](../../MY_NICHE_BUZZ_ASK.md) — help test the keyboard grabber, no robot required
- [GOPOD_SONGS.md](GOPOD_SONGS.md) — all songs, explained in plain language, no background needed
- [PALM_TREE.md](../../life/01_PALM_TREE.md) — the whole thing, put together, no background needed
- [FUNNY_NAMINGS.md](../../web/FUNNY_NAMINGS.md) — every name, character, and phrase this project uses, explained once

**The songs**
- [INTERVIEW VAMP.md](SONG_01_BROBOTS_1_2_INTERVIEW_VAMP.md) — the flagship's video 1, the pre-show banter
- [INTERVIEW RUN.md](SONG_02_BROBOTS_1_2_INTERVIEW_RUN.md) — the flagship's video 2, the seven exchanges
- [BINGO.md](SONG_101_BROBOTS_1_2_BINGO.md) — the shareable upsell video
- [BINGO GAME.md](SONG_102_BROBOTS_1_2_BINGO_GAME.md) — the live two-brobot Bingo warm-up act
- [IS-THAT-YOU SINGLE.md](SONG_103_BROBOTS_1_2_IS-THAT-YOU_SINGLE.md) — the single-bot jewel, live and unscripted
- [IS-THAT-YOU MULTI.md](SONG_103_BROBOTS_1_2_IS-THAT-YOU_MULTI.md) — the two-brobot mix-up, live and unscripted

**The content engine**
- [NICHE_PILLARS.md](../../web/NICHE_PILLARS.md) — how the writing is split into nine kinds of everyday maths, and why
- [AI_WORDPLAY.md](../../web/AI_WORDPLAY.md) — the engine: AI Wordplay, the contests that feed it, and where the content lands
- [AHA_MOMENT.md](../../web/AHA_MOMENT.md) — a live demo the reader can feel work
- [FOODMATH_AHA_MOMENT.md](../../web/FOODMATH_AHA_MOMENT.md) — the foodmath cousin: a live subdomain built in hours, the gap between live and built is the demo
- [BIRTHDAY.md](../../web/BIRTHDAY.md) — the physical proof one — real food-car props, no rendering required
- [NEWSLETTER.md](../../web/NEWSLETTER.md) — subscribe to CRUSHN8R CREW'd — the live follow-along lane

**For venues, funders, and partners**
- [MOBILE_GEAR.md](../MOBILE_GEAR.md) — mobile deployment and field kit
- [OPS ASK.md](../../MY_GOPOD_OPS_ASK.md) — the operator's ops ask — social, sites, and content, a different role than the technical one
- [HEALTHY_DISTRACTIONS.md](../../life/02_HEALTHY_DISTRACTIONS.md) — GOPOD as healthy distraction
- [OUTREACH.md](../../life/02a_OUTREACH.md) — the community and paid outreach plan, two lanes side by side

**For teachers**
- [EDUCATION.md](../../life/03_EDUCATION.md) — GOPOD as a teaching tool
- [TEACHER_INSIGHT.md](../../life/04_TEACHER_INSIGHT.md) — what a session shows a teacher about their room

### Door 2 — More? Dive GOPOD deep

For readers who lean in, who know GitHub — the technical docs, the operator tooling, how it's made.

**Help wanted**
- [MY_GOPOD_ASK.md](../../MY_GOPOD_ASK.md) — the operator's own ask — what's built, where the line is, what kind of help this needs

**For Wire-Pod owners and builders**
- [WIRED-POD.md](../WIRED-POD.md) — what GOPOD changed and added on top of Wire-Pod
- [GOPOD_FEATURES.md](../GOPOD_FEATURES.md) — everything GOPOD built, feature by feature, stack included
- [BODIED_BROBOTS.md](../BODIED_BROBOTS.md) — the robot bodies — proof ladder, Vector L3, Cozmo/Scout staged below L3
- [ALIAS-LIBRARY.md](ALIAS-LIBRARY.md) — every shortcut command GOPOD can run
- [ALIAS-SEQUENCER.md](ALIAS-SEQUENCER.md) — arrangement: notes into sequences into songs
- [PHA0B.md](PHA0B.md) — PLAYHEAD Part 2, the performance front door, whole songs at a time
- [PHCAL.md](PHCAL.md) — PLAYHEAD Part 1, the calibration bench, one primitive at a time

**Doc's Take — lessons learned**
- [DOCS_TAKE_LESSON_1.md](../../life/101_DOCS_TAKE_LESSON_1.md) — the first real mistake GOPOD survived
- [DOCS_TAKE_LESSON_2.md](../../life/102_DOCS_TAKE_LESSON_2.md) — fix the foundation, not the symptom
- [DOCS_TAKE_LESSON_3.md](../../life/103_DOCS_TAKE_LESSON_3.md) — a placeholder that looks finished is worse than an honest gap
- [DOCS_TAKE_LESSON_4.md](../../life/104_DOCS_TAKE_LESSON_4.md) — a living thing sheds, and shedding shows you what is left
- [DOCS_TAKE_LESSON_5.md](../../life/105_DOCS_TAKE_LESSON_5.md) — the scary answer is usually the true one, not the easy binary
- [DOCS_TAKE_LESSON_6.md](../../life/106_DOCS_TAKE_LESSON_6.md) — the honest edge isn't the ask — say it out loud

**The philosophy**
- [LEGACY.md](../../life/05_LEGACY.md) — the pedagogy behind what GOPOD propagates

**How this gets made**
- [AI_AHA_MOMENTS.md](../../web/AI_AHA_MOMENTS.md) — the aha moments from making GOPOD with AI, for the reader who wants to see how the thing thinks

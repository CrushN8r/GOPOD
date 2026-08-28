# Doc's Take: Lesson 3

> A locked door with a sign beats a locked door pretending to be open.
> Say what's coming. Don't fake what's already here.

---

## Context

README.md is GOPOD's own front door — the first page a stranger sees. Partway down the page sat one bold line, a small robot icon next to it: "Bait clip — plays inline here once captured." No clip was there. Nothing played. The line was found during a routine read-through of the whole page, on 2026-07-30, alongside two smaller rough spots on the same page — two section headers using unexplained shorthand, and one internal term used twice with nothing nearby to explain it. All three got written down. Only the bait-clip line got fixed that same day, in commit `d147665`: the promise became an honest line instead — "A session clip is coming. Until then, the songs below are the show." Same length, same spot, same bold line. The only thing that changed was whether it was true.

---

## The metaphor, as Doc tells it

You walk up to a door with a sign taped to it: OPEN, COME IN. You push. It's locked. Nobody's home. Now you don't trust the next sign either, even the true one.

Here's Doc's actual take: **a locked door with a "back in five minutes" sign beats a locked door pretending to be open.** A reader doesn't need everything finished. A reader needs to know which doors actually open right now, and which ones are still being built.

---

## The lesson

The bait-clip line wasn't a bug. Nothing crashed, nothing threw an error — it just quietly told the first-time reader something untrue before they'd read a single other word on the page. A promise wearing the shape of a finished feature. The fix wasn't to build the missing video. The fix was five words: say the clip is coming.

The same posture got used on purpose right after, not just patched reactively. A new page about the four brobot songs was written with four honest "what's still coming" lines instead of hiding its own gaps — one song has a blank joke spot, one is missing a short description, one isn't hooked up to the normal menu yet, and one is plainly marked finished because it actually is. The principle itself got written down as a standing rule the same day (commit `bf0be1c`): a gap between where something is and where it's headed isn't a defect to hide — it gets named, labeled as what's coming, never disguised as already done.

---

## What it taught — 4 rules that don't go away

1. **A placeholder that reads like a finished feature is worse than a plainly named gap.** Staying quiet about "not done yet" isn't neutral — it's a small lie sitting in public.
2. **Say what's coming. Don't fake what's here.** "Coming soon" costs five words and costs nothing else.
3. **The front page gets read first — check it first.** A stranger's whole first impression can turn on one line, before they've read anything else.
4. **An honest "not yet" travels.** Once one page said it plainly, the same shape got reused elsewhere on purpose, not by accident.

---

## Where this doesn't go far enough yet

One fixed line is a small sample — a single sentence on a single page, not a pattern proven across the whole project. The two other rough spots found in that same read-through were left exactly as they were, on purpose: two section headers still carry unexplained shorthand, and one internal term still appears twice with nothing nearby to explain it. Naming a gap doesn't erase it — those two are still gaps, just gaps nobody has chosen to close yet. And the real payoff — whether an honest "coming soon" actually earns more trust than silence — hasn't been measured at all. Nobody has watched a real stranger read the page and reacted to what they thought. That part is still a guess, not a result.

---

> From Doctrine Barfallonyou
> Doc's Take! A locked door with a sign beats a locked door pretending to be open — say what's coming, don't fake what's here.
> Boom! Done! Class Dismissed!
> — Doc Squawkadoodle

---

## GOPOD YAHMM (You Are Here Mall Map)

Part of GOPOD — see [life/README.md](README.md) for everything else in this folder, or [the root map](../README.md) for the rest of GOPOD.

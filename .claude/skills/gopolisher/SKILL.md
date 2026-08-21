---
name: gopolisher
description: Use before inspecting, diagnosing, cutting back, or auditing any GOPOD target — a file, a song, a document set, the alias registry, a campaign, the repo. Three escalating modes under one roof — a domain-general inspect/diagnose/cut discipline (the era test, comb-through, detangle-from-the-ends, trim-for-trajectory), a mechanical drift check between docs/aliases/memory and the real files they describe, and the worker+critic "gauntlet loop" for a full horizontal cohesion pass across the song shelf, alias registry, and repo-root docs. Start cheap (mechanical check), escalate only as far as the target actually needs.
---

# Gopolisher

One name, three modes of the same underlying job — keeping a target honest, clean, and
cohesive — ordered cheapest-and-most-mechanical to heaviest-and-most-judgment-driven.
Merged 2026-08-06 from three previously separate skills (`hairstylist`, `decoupler`,
`event-planner`) that shared a family resemblance but not a name.

## A. Which mode applies

- **Mode 1 (comb-through & cut)** — a target needs inspecting, a knot needs untangling, or
  something needs trimming back. Cheap, manual, judgment-driven, domain-general.
- **Mode 2 (mechanical drift check)** — checking whether docs/aliases/memory have drifted
  from the real files they describe. Cheap, scripted, exit-code-always-0, a report not a
  gate.
- **Mode 3 (the gauntlet loop)** — a full horizontal cohesion pass across the song shelf,
  alias registry, and repo-root docs is wanted, graded against an explicit bar until a
  critic is satisfied. Expensive (real tokens/time), gated, invoked only on request.

Default order when unsure: run Mode 2 first (it's free and mechanical) → Mode 1 if
something it surfaces needs actual untangling → Mode 3 only once 1 and 2 are already clean
and what's left is a qualitative cohesion judgment call, and only past this skill's own
gates (§D3).

## B. Mode 1 — Comb-through & cut

A hairstylist-derived working discipline for inspecting, diagnosing, and cutting back a
target. A target may be a file, a song, a document set, a campaign, or a repo — this mode
never assumes which. Provenance: `gopod_notes/older_notes/HAIRSTYLIST_DISCIPLINE_DOCTRINE_001.md`.

### The era test — run this before anything else

Establish which era the target is in before choosing a discipline:

- **BUILDING era** — the foundation is not yet proven. Foundation-first applies: fix at
  the root.
- **MAINTAINING era** — the foundation is sound and length has grown on it. Ends-first
  applies.

The two are rarely in play at the same moment. Getting the era wrong is what causes
damage — going root-first on a sound foundation tears through working length to reach a
small tangle.

Diagnostic: how often work on this target returns to the root reveals which era is
actually in force. Frequent root work while believing the target is in maintenance means
the foundation was never finished — treat it as still in the building era, not maintenance.

**This era test also gates Mode 3 (§D3)** — a target still in its building era is out of
scope for the gauntlet loop regardless of anything else.

### The three disciplines

**Discipline 1 — Comb-through with a fine brush.** A systematic pass that touches every
part of the target and surfaces snags without changing anything. A name search finds
mentions but not paths used as data values — a hardcoded path inside a config file is a
functional break the same search surfaces without flagging as one, since it reads like
just another hit. A real incident found this the hard way: a moved folder left a model
path hardcoded in a config file that would have silently stopped resolving, caught only
by a live re-check after the search, not by the search itself. A comb-through is also
only as wide as the tree it's run against — a repo-scoped search finds zero hits for a
caller that lives outside the repo entirely and can't be found by making the in-repo
search more thorough. A real incident: the same folder move broke two hardcoded paths in
`~/.gopod_alias_lib/` (untracked, outside the git repo, one running at every terminal
open) that a `goverlord/`-scoped survey had no way to see; caught only when the operator
hit the live symptom.

**Discipline 2 — Detangle from the ends to the base.** When something is knotted, begin
at the outermost symptom and work inward until the origin of the knot is located.
Starting at the root drags every knot ahead of you and tightens the whole thing. The
corrective cut is then made at the base once the shape is visible. (Maintaining-era
discipline, per the era test above — a building-era knot gets fixed at the root directly.)

**Discipline 3 — Trim for growth trajectory, not current shape.** The target drifts on
its own between sessions. A cut sized only to today's state can be wrong by the next
session because everything kept growing. Cut with the drift direction accounted for,
aiming past the immediate target toward the intended long-term shape, followed by
continued small pruning to hold trajectory rather than one large correction later.

### The cut-note — an output of Discipline 3, not a fourth discipline

Discipline 3 emits something: a short, dated statement of why a cut was made and what it
was made toward, recorded at the moment of the cut while the reason is still in hand.

Restraint rule, so this does not become ceremony: note the cut only when the reason is
not visible in the cut itself. Obvious removals stay silent. A cut where a reader would
reasonably ask "why cut there?" or "why was that left alone?" gets a line.

When the cut is a retirement — an alias commented out, a script archived, a folder
superseded — the files it owns get untracked from git in that same commit, not later;
they stay on disk, they just stop being public. A real incident found the gap: a cluster
retired with a commented-out alias stayed tracked for three weeks until a full survey was
needed to turn up 21 of 28 orphaned files that could have been untracked at retirement
time. When that step gets missed, the backstop is a no-caller sweep before any repo
tidy-up or first public look — for every tracked file, is anything calling it — which is
Discipline 1, a comb-through, not a new step.

### Mode 1 scope

- Domain-general, on purpose — no song, robot, Wire-Pod, Vector, or other GOPOD subsystem
  is named in this mode's own rules. Whatever concrete target a task names, apply the era
  test and the three disciplines to it, not to a subsystem baked into this file.
- Discipline 1 is read-only by definition — a comb-through changes nothing. Disciplines 2
  and 3 end in an actual cut, which is a real edit and follows this repo's normal edit
  discipline (read before write, explicit go-ahead before anything lands).

## C. Mode 2 — Mechanical drift check

The bridge between code/config (flexible, changes often) and the docs that describe it
(which only stay true if someone keeps checking). Built 2026-07-24 after a real session
where three separate kinds of drift turned up back to back: a stale alias-registry gap,
broken paths in agent read-first lists, and stale `gopod_notes/` citations inside memory
files themselves — see `PHA0B_SONG_LIST_CLEANUP_SURVEY_001.md` and the memory-file fixes
from that same pass.

Two halves, on purpose: the **mechanical check** (this mode's own script) finds
candidates; the **judgment pass** (Claude, reading the output) decides which are real
drift vs. a known-fine collapsed-range mention. Neither half replaces the other.

### Running the check

```
python3 .claude/skills/gopolisher/gopod_consistency_check_001.py
```

Read-only, no arguments, exit code always 0 (a report, not a gate). Three checks:

1. **Repo-wide markdown links** — every `](...)`.md link in every tracked `.md` file
   (vendored third-party docs under `SDK/sources/` excluded) resolves to a real file.
2. **`.bashrc`/`.bash_aliases` loaded files vs. `tech/alias_play_studio/ALIAS-LIBRARY.md`** —
   every alias/function actually defined in a file `.bashrc`/`.bash_aliases` loads gets
   checked against the registry doc's own text.
3. **Memory files' `gopod_notes/` citations** — every `gopod_notes/FILENAME.md` mentioned
   inside `~/.claude/projects/.../memory/*.md` resolves at top level (not silently moved
   into `older_notes/` by a later gohandoff close).

### Reading the output — the judgment pass

Check 2 will print names that read like gaps but aren't: `ALIAS-LIBRARY.md` legitimately
collapses similar aliases into one row (`` `brobots-anim-happy` / `-very-happy` / `-sad` /
... ``, `` `codex-1`...`codex-6` ``, `` `llm-test-brobot` / `-goverlord` / `-deep` / `-coder`
``) — a plain substring check can't see those as covered. **Before treating any "MAYBE
MISSING" hit as real drift, grep the doc by hand for a collapsed form first.** As of
2026-07-24 all 16 of that check's hits are exactly this false positive, confirmed clean.

Checks 1 and 3 don't have this false-positive problem — a hit there is real; fix the path
or move the cited file back, whichever is actually true.

### What this mode doesn't check (yet)

Hardcoded lists that duplicate a canonical source but aren't a literal broken path — e.g.
an agent's `description:` field enumerating songs by hand instead of pointing at
`goverlord/runtime/songs/` or `ALIAS-LIBRARY.md`'s own registry (fixed once already in
`song-lane.md`, 2026-07-24). No generic script check for this — it takes reading the
prose. When editing any `.claude/agents/*.md` or `.claude/skills/*.md` file, ask: does this
restate a fact that already lives somewhere else? If yes, replace the restatement with a
pointer, the way `song-lane.md`'s song list became a pointer at `ALIAS-LIBRARY.md` instead
of its own copy — that fact can't go stale again if it's never duplicated. **This exact
gap — prose-level/semantic judgment a script can't make — is what Mode 3 exists for.**

### Mode 2 scope

Read-only by default. A finding gets reported and judged, not auto-fixed — the operator
sees what's actually broken before anything changes, same discipline as every other GOPOD
skill. Not a gomad: this never runs on Brobot 1 or Brobot 2, never gets imported by a song runner —
it's Claude's own working procedure, same reasoning as every other mode in this skill (see
`gopod_notes/older_notes/STUDIO_SKILL_VS_GOMAD_001.md`).

## D. Mode 3 — The gauntlet loop

### D1. Identity

The "gauntlet loop": bar + worker + critic, aimed at something already solid. Surfaced via
a translator-chat relay of an outside video's core idea (Anthropic's own worker/critic
multi-agent pattern, not a fad) — the operator's own read on it: no new content there, but
one usable mechanism, and outside confirmation that "foundation first, polish second" was
already the right order. Framed by the operator as an event/wedding-planner role: he sets
the bar, this mode runs the loop, the plan develops with him over repeated passes — not a
one-shot fixed pipeline.

### D2. The warning that matters most

Pointed at a weak foundation, the loop just polishes the wrong thing beautifully. That's
why this is Mode 3, not Mode 1 — it never runs before Modes 1/2 would already be clean.

### D3. The gate — don't start with the loop

Two concrete gates, both must pass before Mode 3 runs:

- **Campaign-level gate** — `niche-buzz` §8's launch checklist, step 4, "Glamour Skin,"
  is this mode's literal slot: footage banked, dead weight jettisoned, and the Crystal
  Cut done come first. This is dessert, not dinner.
- **Per-target gate** — even once the campaign clears step 4, run the era test (§B) on
  the *specific* target named. A target still in its BUILDING era is out of scope for
  Mode 3 regardless of campaign phase — that target gets Mode 1's root-first discipline
  instead, not this loop.

If either gate fails for what's actually being asked, say so and stop. Do not run the loop
on a technicality that the campaign is "close enough."

### D4. Three horizontal surfaces

Any invocation names which of these (or which slice within one) is in scope — this mode
never assumes "all three, every time":

- **Song shelf** — `goverlord/runtime/songs/*/` (each song's `story.md`, `knobs.json`, and
  its own `SONG_*.md` doc).
- **Alias registry** — `tech/alias_play_studio/ALIAS-LIBRARY.md` /
  `ALIAS-SEQUENCER.md` against the live `.gopod_alias_lib/*.sh` /
  `.bashrc`/`.bash_aliases` files they document.
- **Repo-root docs** — the reader-facing doc tree (root + `alias_play_studio/` /
  `learned/` / `life/` / `tech/` / `.claude/`), per the already-banked "just repo root"
  scope precedent (`goverlord-desk` §2b) — excludes `goverlord/` runtime and vendored
  `SDK/sources/` by default, same as that precedent.

### D5. The bar — Goverlord states it, every time

Per `goverlord-desk`'s Fork vs Fetch section: the bar is decided in FORK mode, plain
words, before any loop runs. This mode does not invent the bar, and does not reuse a prior
run's bar without it being restated — what counts as "cohesive" shifts as the shelf grows
(Discipline 3, §B: trim for trajectory, not a fixed target).

Illustrative shapes a bar might take (not prescriptive — the operator's actual wording each
time is the real bar): "every song's own `SONG_*.md` opens with the same why-this-exists
shape," "every `ALIAS-LIBRARY.md` row matches the voice of the newest rows," "every
root-level doc states the reporter-framing rule the same way."

### D6. The mechanism — worker + critic via Workflow

Built on the Workflow tool's own worker-then-critic ("adversarial verify") pattern — not a
new mechanism, the one already documented there. Shape:

1. One worker pass per named surface (or slice), produced against the stated bar.
2. An independent critic pass grades that draft against the *same* bar, instructed to
   refuse anything short of it — not to grade generously.
3. If not satisfied, the worker revises against the critic's stated gaps; repeat.
4. **Bounded, not unbounded** — cap at a handful of rounds (e.g. 3-4) before escalating
   back to the operator instead of looping indefinitely. "Some runs went two hours" is a
   cost warning to respect, not a target. If the cap is hit unsatisfied, report the actual
   gap plainly rather than forcing a pass or quietly extending the loop.
5. Only the critic-approved result is reported back — intermediate drafts and critic
   objections stay inside the loop's own run, matching "you only see the polished result."

Illustrative skeleton (written fresh per real invocation against the operator's actual bar
and surfaces — not a fixed script this mode ships):

```
phase('Sweep')
const bar = args.bar               // stated by the operator, never invented here
const surfaces = args.surfaces     // e.g. ['song-shelf', 'alias-registry', 'root-docs']
const results = await pipeline(surfaces, async (surface) => {
  let draft = await agent(`Produce a cohesion pass over ${surface} against this bar: ${bar}`, {phase: 'Sweep'})
  for (let round = 0; round < 4; round++) {
    const verdict = await agent(`Grade this pass against the bar, refuse anything short of it: ${bar}\n\n${draft}`, {phase: 'Critic', schema: VERDICT_SCHEMA})
    if (verdict.satisfied) return {surface, draft, rounds: round + 1}
    draft = await agent(`Revise to address: ${verdict.gaps}`, {phase: 'Sweep'})
  }
  return {surface, draft, rounds: 4, unresolved: true}
})
return results
```

### D7. What Mode 3 never does

- **Never edits a target directly** — a song's `story.md`/`knobs.json`, an alias file, or
  a root doc. Output is a critic-approved *proposal*; applying it is a separate, later
  step through that target's own normal discipline (`survey-then-commit` for anything
  landing in git, `hardware-calibrate` if a proposal touches a physical constant, the
  one-pen-crossing rule for any song-score write-back per `niche-buzz` §7 pending item 6).
- **Never self-invokes.** Same standing rule as every Workflow use in this repo — requires
  the operator's explicit ask, not a spontaneous idea acted on. Pitch it in one line if it
  occurs to you; don't run the loop until told to.
- **Never skips Modes 1/2.** Mode 3 is the expensive qualitative pass, reserved for once
  Mode 2's mechanical checks are already clean and what's left is a judgment call about
  tone/structure/cohesion — not a substitute for running Mode 2 first.

## Scope

- Read-only by default across all three modes (Mode 1 Discipline 1, all of Mode 2, all of
  Mode 3) — an actual cut/edit only happens under Mode 1 Disciplines 2/3, and always
  follows this repo's normal edit discipline (read before write, explicit go-ahead).
- Domain-general in Mode 1, GOPOD-repo-specific in Modes 2/3 (Mode 2's script paths, Mode
  3's three named surfaces).
- Mode 3 is gated twice (§D3) — campaign-level and per-target — and is a finishing tool,
  not an everyday default, given its real token/time cost.

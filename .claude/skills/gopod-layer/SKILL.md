---
name: gopod-layer
description: Use when a task touches the GOPOD layer — the multi-robot, multi-persona live venue behind "GOPOD Yourself" — or asks what comes after the niche-buzz campaign. States the cast, the routing map, and the honest hardware ledger, and states the gate plainly: none of it is a task until niche-buzz itself goes evergreen. A starter skill, sibling to niche-buzz, gated the same way web-orbit is — map first, build later.
---

# GOPOD layer

## A. Identity

The GOPOD layer is the dessert behind "GOPOD Yourself" — the multi-robot,
multi-persona live venue with senses, entered from the brobots wire-pod layer via the
`GOPOD_YOURSELF` intent (defined in `customIntents.json` on the Wire-Pod host, not in this repo)
into the cockpit. Per the ladder (`niche-buzz` §5, evergreen doctrine): when this layer
becomes the active dinner, it gets proven and finished properly once, the same way the
brobots wire-pod layer is being finished now, before anything above it is built.

## B. The gate

**Nothing in this skill is a task until the niche-buzz layer goes evergreen** — `niche-buzz`
§8's launch checklist complete: footage banked, weight jettisoned, crystal cut, glamour,
ignition. Until then this skill is a vault and a map, not a work order. Sessions do not
start GOPOD-layer build-out unprompted; a beautiful dessert menu is not permission to
cook.

## C. Cast & routing map

Per operator direction, from project archives:

- **Conductors** — CHALK (visual intelligence host: teaching surface, display manager,
  the screen persona) and PLAYHEAD (performance conductor: timing, pacing, cues — the
  committed `playhead` skill is its conceptual seed). Both are the female/male reporter
  voices' GOPOD-layer score, per `niche-buzz` §6.
- **Primary performers** — Doc (Vector 1) and Pip (Vector 2), envisioned GOPOD-layer
  personas; the Brobot bodies they'd ride are live-capable today.
- **Student interns** — Git Repo (good-guy hacker/geek) and Cache PYC
  (memester/hustler), intended Cozmo bodies.
- **Security team** — Intel 404 PNFH ("Page Not Found Here") and Intel 404 PNFT ("Page
  Not Found There"), intended Moorebot Scout bodies, dry-humor checkpoint officers.
- **Guest orbit** — Cameo, the framework for temporary personas entering and exiting
  without replacing the core cast.

Numpad routing as designed: KP1 Doc, KP2 Pip, KP3 Git Repo, KP4 Cache PYC, KP5/KP6 the
Moorebots (Scout), KP9 CHALK, KP0 exit.

Principle, verbatim (`niche-buzz` §6): **voices are instruments; personas are sheet
music.**

## D. Staged truth

The honest hardware ledger — no network identifiers:

- **Vectors** — proven, live-capable.
- **Cozmos** — staged (SDK import and connectivity mapped), live control NOT proven;
  chat-first, body-later. **Known-hard gotcha, future wiring**: the two USB WiFi adapters
  for Cozmo comms are OS-identical (same chip, RTL8723BU, same VID:PID) — `lsusb` can't
  tell them apart, bus/device numbers aren't stable; distinguish only by physical port
  path. Commands exist but are not yet run/confirmed live at brain. Full detail:
  `gopod_notes/COZMO_USB_WIFI_ADAPTER_ID_001.md`.
- **Moorebots (Scout)** — source staged, live control NOT proven.
- **Vision lane** — a Coral USB Edge TPU perception service, proven stable in
  isolation: cameras rotate, the TPU consumes frames and emits recognition packets to
  the cockpit; video is for humans, frames are for the TPU, packets are for GOPOD.

Standing rule: nothing is claimed live-capable without its own proof lane — golden
studio discipline (dry-first, guarded live, PASS/BLOCKED, operator's eyes as the health
gate — see `studio`'s `dry-verify` and `hardware-calibrate`) applies to every new body.

**Robot-layer truth principle**: connectivity is not control. Control is not physical
proof. Physical proof is not recovery. A link that pings is not a robot that moves —
applies to every body added to this cast, present and future.

## E. The scenario chain

The live-venue funnel, twin of the video funnel (`niche-buzz` §3):

HDB (Home Departure Beacon) → MOTS (Moving On The Streets) → POTS (Person On The
Street) → PAAT (People At A Table) → GPYS (GOPOD Yourself) → post-intent GOPOD
session.

GPYS is the same hinge the videos close on — street, table, and screen all converge on
the same two words.

## F. Wave note + vault pointers

This layer is prepped to catch the second Anki/DDL promotion wave — the first wave
lands on the evergreen niche-buzz layer (`niche-buzz` §7 Parked list, Anki/DDL pitch);
this map means the venue answer is ready without re-derivation when demand arrives.

Where the deeper gold lives — this skill points, it does not duplicate:
- The gold vault, `gopod_notes/GOPOD_LAYER_GOLD_VAULT/` — private lane, outside this
  repo, never git-tracked: full persona cards, staged hardware detail, proof-lane
  design, vision-lane detail, shop/community/comedy seeds, the livestream engine,
  portable context cards.
- `niche-buzz` §7's own Parked list — WAWN NEWS, persona visuals, lexicon glossary,
  shop/community/venue seeds.

The vault's §2 also now carries a **last-known-working robot configs** reference
(2026-08-15) — per-robot transport/adapter/IP/known-issue detail, Vector marked BUILT,
Cozmo and Moorebot marked FUTURE. A speculative robot-lab architecture (a nine-point
link contract, an L0-L3 proof ladder, a shared RobotNode layer, lab cards, an
evidence-folder structure) was raised this session as future-engineering scaffolding for
this unbuilt layer — **parked, not adopted**: no version of it exists as current
architecture anywhere in this repo or the vault.

Depth joins this skill at ignition; starter boundary now, on purpose.

## Scope

- Orientation/reference only — no logic, no automation, no code of its own.
- Starter size, on purpose — trim rather than grow. If any section wants to expand,
  that expansion waits for ignition (`niche-buzz` §8), same as `web-orbit`'s own
  starter/depth split.
- No network identifiers, no credentials, no secure-lane content, no adult-lane
  content, ever — same standing test as `web-orbit`: would this be fine for a
  Wire-Pod visitor to read on launch day?
- For the campaign one level below this layer, see `niche-buzz`. For song-production
  procedure, see `studio`'s index.

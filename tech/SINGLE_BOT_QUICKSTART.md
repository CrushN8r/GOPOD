# Single-Bot Quickstart

> One Vector. Free-form LLM chat. No second robot, no Jetson, no alias studio required.

This is GOPOD's other door. Everything else in this repo (`pha0b`, `phcal`, the
alias-studio tooling) is built for a collaborator running two robots and a full
performance rig — the pro lane. This page is the adoption lane: what a regular Vector
owner, with one robot and some technical comfort, can actually get running today to
talk to their own robot with an LLM behind it.

**Build-in-public, this whole page.** What works is real. What's broken is named
plainly, not hidden. Truth-checked against actual code on 2026-08-18 — see
`gopod_notes/SINGLE_BOT_ONRAMP_TRUTH_001.md` for the full survey this page is built on.

---

## What actually works today

- **Free-form LLM chat, two real paths.** Anything you say that isn't a hard-coded
  intent reaches an LLM — either through Wire-Pod's own native backpack/knowledge-graph
  chat lane (stock Wire-Pod, works out of the box, pointable at any OpenAI-compatible
  endpoint via `apiConfig.json`), or through GOPOD's own PTT writer script, which makes
  a real, working call to Ollama's OpenAI-compatible endpoint.
- **Mic detection that isn't hardcoded to one operator's hardware.** `gopod-mic-detect`
  / `gopod-mic-set` / `gopod-mic-test` do real dynamic USB-input discovery against a
  config file, not a fixed device name — this genuinely ports to a different USB mic.
- **Numpad/keyboard capture with real device auto-discovery.** The writer finds a
  numpad or keyboard by its actual key-capability bitmap first, name-keyword matching
  second — not hardcoded to one physical device.
- **A real output sanitizer, live in the path that uses it.** Before anything reaches
  the robot's mouth (through the PTT writer's own dispatch), the reply gets stripped of
  wrapping quotes, asterisks, emoji, and run through a pronunciation-normalization
  filter. This is wired in and actually runs, not a paper guarantee.

None of the above needs a Jetson, a second Vector, or `pha0b`/`phcal`. The PTT writer
is a standalone script — install its dependencies, point it at your one robot, run it.

## Where it's honestly not there yet

**This is not plug-and-play today.** Two real blockers stand between "clone the repo"
and "just talk to your robot":

- **No spoken wake word.** Saying "GOPOD Yourself" out loud doesn't currently start
  anything — the custom intent that's supposed to catch that phrase points at a script
  file that no longer exists on disk. Today's real front door is running the PTT writer
  directly from a terminal, not speaking a phrase at it.
- **Robot targeting is hardcoded, not configured.** Which robot KP1/KP2 talk to is a
  literal serial number baked into the Python source, set to this operator's own two
  Vectors. Pointing it at a different robot means editing source, not editing a config
  file — a real, if small, barrier for a new owner.

One more gap worth naming honestly: the sturdier of GOPOD's two output-safety
mechanisms — a tested validator-and-repair pass built in the Wire-Pod Go layer — exists
in the codebase but has no live caller anywhere. It's real, built code sitting unplugged.
The native Wire-Pod chat lane (the easiest one for a casual owner to reach) currently
runs with no GOPOD-added output guard at all.

## What it takes to actually run this today

Real, non-trivial technical setup — not casual-owner plug-and-play, but not entangled
with GOPOD's two-robot orchestration either:

- Ollama running locally with a model pulled (or an OpenAI-compatible endpoint you
  already have)
- `vosk` + a downloaded Vosk speech model, `sounddevice`, `scipy`, `numpy`
- Wire-Pod already running against your one robot
- Run the PTT writer directly (`python3 gopod_ptt_chat_writer_013.py`, with
  `--list-devices` to find your mic/keyboard first)

## Invitations — this is where a contributor could help

These are scoped, honest, not-yet-built fixes — not a promise of when, and nothing here
is built by this page. If you're the kind of person `MY_GOPOD_ASK.md` is talking to,
this is a concrete, contained place to start:

1. **De-hardcode the robot serials.** Move `PERSONA_ROBOT_SERIAL`/`PERSONA_BY_KEY` out
   of Python source and into the same config-driven pattern the rest of the writer
   already uses, so a new owner edits a config file, not code.
2. **Fix the spoken wake trigger.** Repoint `GOPOD_YOURSELF`'s custom-intent exec path
   at something that exists, or replace it with a working equivalent — or, short of a
   fix, document the terminal-run path as the real current front door instead of
   implying a working spoken wake exists.
3. **Wire in the built output guard.** Either connect the tested Go
   validator-and-repair pass to the native Wire-Pod chat lane, or clearly document that
   only the PTT writer's narrower path is currently guarded.

None of these are large. All three are exactly the kind of "past my ceiling, contained
enough to grab" piece `MY_GOPOD_ASK.md` describes.

---

> From Doctrine Barfallonyou
> Lesson! A door that's honestly half-open beats one that's dishonestly wide.
> Boom! Done! Class Dismissed!
> — Doc Squawkadoodle

---

## GOPOD YAHMM (You Are Here Mall Map)

Part of GOPOD — see [tech/README.md](README.md) for everything else in this folder, or [the root map](../README.md) for the rest of GOPOD.

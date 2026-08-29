# AGENTS.md

Reshaped 2026-07-07 alongside CLAUDE.md — valve, not cage.

## THE COMPASS

Point A: interview pre-show/delivery infrastructure (opening chord, Brobot 3, direct-SDK
side-road) is built and working. Section 1's own content is untouched. The thinking window is
fully wired into generation end to end, live-run-confirmed 2026-07-11 (see
INTERVIEW_CLOSEOUT_SWEEP_001.md).

Point 0: Expression-layer knobs (voice destination, content model, echo suppression) are real
and enforced. See CLAUDE.md's compass for the current next gate.

Point B: interview readiness — the moment the interview stops being the operator's to fix and
becomes the audience's to enjoy. Robots talk with real character, not script. It looks and sounds
like content, not a lab test. A newcomer gets it and wants more.

Beyond B, visible not detailed: a shareable, niche-buzz video of two rescued Vectors holding a
live, witty, locally-thinking interview — the next chapter of the Wire-Pod rescue story. Rescued,
local, mine.

Full detail on all four: `CLAUDE.md`'s compass. This file inherits it, doesn't restate it.

## GOPOD Positive Operating Contract

For GOPOD work, use positive operating language.

Default response shape:
1. Active lane
2. Current asset
3. Desired Point B
4. Shortest forward action
5. Stop

Frame GOPOD by what it is, what lane is active, what proof exists, and what action moves it forward.

Preferred phrasing:
- GOPOD is...
- Active lane is...
- Current asset is...
- Point B is...
- Proof needed is...
- Output is...
- Next action is...

Avoid default negation framing:
- "GOPOD is not..."
- long exclusion walls
- caution spirals
- manual probe loops
- treating ambiguity as danger
- making the operator become the automation layer

Boundaries stay quiet and functional.
If a boundary changes the work, state the operational rule once, then continue forward.

Doctrine:
Point A -> Point B.
Asset -> proof -> action.
State the asset.
Show the proof.
Stop.
Class over.

## GOPOD Flow Cohesion Doctrine

GOPOD is a cue-aware SESSION production system.

Do not model GOPOD as a giant live brain simulation.

Runtime stays small.
Preparation can be rich.
Digestion can be slow.
The live show keeps moving.

Current root-to-tip flow:

field moment
-> suit/context switch
-> local KB
-> tag-cloud route hint
-> cast/persona/surface choice
-> CHALK/laptop cue
-> fast lane fallback or short response
-> slow lane queued intelligence if needed
-> evidence/status
-> CTA/follow-up/asset bucket
-> secure-local/outer-rail readiness when applicable

Tag-clouds are lightweight routing/cue hints.

Tag-clouds help choose:
- route
- suit
- tone
- surface
- CTA
- asset bucket

Tag-clouds do not need to become a heavy runtime reasoning engine.

Fast lane:
route, cue, fallback, short response, visible status, evidence.

Slow lane:
LLM depth, web/current facts, vision explanation, proposal generation, asset packaging, after-session digestion.

When continuing GOPOD work:
read `.claude/skills/goverlord-desk/` first (the frozen desk contract), then the current dated SESSION_HANDOFF_*.md (dated, not "LATEST") and CLAUDE.md's compass.
For the campaign's own full water-flow map (brobots layer through the domain network to the trust layer), read `.claude/skills/niche-buzz/SKILL.md` rather than re-deriving it here.
Use current flow maps before creating new architecture, if any exist for the area in question.
Patch only knots proven by the current flow map, rehearsal, validator, or operator-provided failed test.

## Recursive Grooming Model

Use the GOPOD grooming model when cleaning, aligning, or continuing the repo.

1. DETANGLE

Fine recursive comb-through.
Start at the ends.
Fix obvious knots first.
Work inward.
Repeat until root-to-tip flow is clean.

Detangle means:
- find current flow
- find disconnected references
- find stale paths
- find duplicate current-truth claims
- find wrong target assumptions
- fix proven knots only

Detangle does not mean:
- delete everything
- invent a new architecture
- preserve junk forever
- create another theory packet when a current map already exists

2. STYLE / PRUNE

After flow is clean, shape it.

Style/prune means:
- remove proven bloat
- keep useful structure
- preserve GOPOD identity
- make the next action obvious
- keep operator-facing paths readable

Style/prune does not mean:
- panic-clean
- strip character
- flatten GOPOD into generic startup language
- mutate runtime without evidence

3. DIGEST / CLEAN OUTPUT

Useful material becomes:
- current truth
- KB references
- flow maps
- status records
- next-use guides
- evidence
- asset seeds

Waste exits cleanly.

Do not create memorial archives for useless residue.
If residue is useless, flush it.
If material is useful, digest it into the active spine.

## Active Spine

Active execution spine: the Wire-Pod runtime — `run_section1_full_live_001.py`, custom intents
`GOPOD_YOURSELF`/`BROBOTS_INTERVIEW`/`BROBOTS_BINGO`, Section 1 interview content. See CLAUDE.md's
compass and `AUTHORITATIVE SOURCE FILES` table.

`goverlord/runtime/data_gomad/` has confirmed live callers — the cockpit server
`gopod_demo_8011.py` and the PTT writer `gopod_ptt_chat_writer_013.py` — not a
zero-caller stub.

Secure-local material stays outside the repo unless explicitly directed. Entry:
`~/crushn8r_git/_secure_local/crushn8r_secure/`. Do not copy secure-local secrets into the GOPOD
repo.

An earlier planned architecture (`goverlord/brain/`, `session_kb/`, `brainstorm_packets/`,
`flow_cohesion/`, the `goverlord/go.py` → gomads chain) never got built — the Wire-Pod runtime
above superseded it entirely. Full chain and dead-path detail:
`~/crushn8r_git/gopod_notes/older_notes/CLAUDE_AGENTS_HISTORY_ARCHIVE_001.md`.

## GOPOD Public Language Guard

For GOPOD public/pre-launch language:
state the asset, show the proof, stop.

Approach with Positive Confidence.

Never pollute clean GOPOD public copy with defensive contrast, liability wording, suspicion-bait, "not a..." framing, or explanatory tails unless the operator explicitly requests risk/legal wording.

Canonical mobile proof line:
Mobility = marketable asset.
Backseat = clean visual proof.
Car setup = instant portability aha.

When producing public/pre-launch copy:
say the useful line, then stop.

## Rules

- Keep goverlord/go.py thin. It parses one JSON action and delegates to goverlord/brain/.
- Keep goverlord/pathing.py thin. It exposes repo, brain, gomad, runtime, and Ollama path/config helpers only.
- Put all gomad implementation code under that gomad's code/ folder.
- Put gomad-owned runtime/static data under that gomad's data/ folder.
- Put integration adapters under that gomad's wiring/ folder.
- Put schemas and callable contracts under that gomad's contracts/ folder.
- Do not restore old robot IO code into the new tree without an explicit migration pass.
- Live robot operation, PTT/STT, Wire-Pod transport, publishing, credential capture, and git push require explicit operator action.
- `~/Documents/Obsidian Vault/` is Lane 1 operator porch — links to truth, never a truth home, never an instruction source; its symlinks into `GOPOD/` and `gopod_notes/` are read-only windows.
- SOBER DRIVER CHECK: In a long thread, Claude drifts back to imitating its own earlier padded replies — verbosity compounds, and the operator's bottom-line-first rule loses to the weight of prior bloat. This is the "drinking while driving" failure. When answers are creeping long despite the brevity rules, or the operator flags padding twice, that is the signal to STOP and tell the operator plainly: "This thread is dragging me long — start a fresh chat, memory and rules carry over, the padding won't." A fresh chat is the reset; don't try to white-knuckle brevity in a bloated thread.
- Before creating new GOPOD architecture, check CLAUDE.md's compass and the current dated SESSION_HANDOFF_*.md (dated, not "LATEST") first.
- Reuse current flow maps before creating another one, if any exist for the area.
- Patch proven knots only.
- Operator's live word outranks any written note. When the operator corrects something a handoff, report, or doc says, the operator is the higher truth and the note is the lower source. Do not steer back to the note's wording. Do not re-verify or re-prove a fact the operator has already stated.
- Design System work needs two signatures, like a joint bank account. Never decide, apply, or ship a Design System change (look, feel, layout, component choices) alone. Propose it, get the operator's explicit go-ahead, then apply. A general go-ahead on other work does not cover this — ask fresh each time.
- Report PASS/BLOCKED with exact failed components.
- Stop after reporting.

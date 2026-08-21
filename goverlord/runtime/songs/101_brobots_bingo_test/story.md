# Brobots Bingo Capture Song 001 — the upsell video test run

**WIP — 2026-08-11. Golden lock broken by design.** The prior GOLDEN state
(`golden_run_2026-08-09_19-51-14`, 57 steps) is preserved as a reference snapshot at
`gopod_notes/GOLDEN_SNAPSHOTS/101_brobots_bingo_2026-08-11/`. This song was rebuilt the
same day to a new step-naming notation — `host_`/`brat_` speaker prefixes, standardized
action suffixes, a bare `reporter_gap` separator at every segment boundary, all three
rounds sharing one skeleton (Round 1 and Round 2 each gained a mirrored
`brat_searching_animation` step to match Round 3) — per
`gopod_notes/BINGO_NOTATION_REMAP_PROPOSAL_001.md`'s operator-approved mapping. 8 steps
cut (`banter_kg_success`, the full Round 2 "attitude" cluster of 5 steps plus its closing
gap, `round_3_answering`). Net: 57 → 69 steps. **Re-locks to GOLDEN only after a fresh
live run confirms this shape — not done this pass.** `knobs.json`/`zKnobs.json` remain
byte-identical (both rewritten together). Prior-session notes (tempo resolved at `0.0`,
the two alias retirements) still stand, unaffected by this restructure.

Knobs: [knobs.json](knobs.json)

This is **not** the live 75/90-ball Bingo game (`gobingo`, `goverlord/runtime/songs/102_brobots_bingo_game/`) —
that sidecar and its optional reactor are untouched by this song. This is a scored,
straight-line capture run for the Bingo prototype upsell video, in the
`test-reaction-pick-animation` / `brobots_bait_002` mold: an opening sync + host arm cue,
interleaved fragment banter with gestures and reaction beats, then three paced ball-draw
rounds (each self-contained — juggling/rattle/ready/call folded into that round's own
start, no separate transition sections between them), and a close. Organized into labeled
sections (`OPENING SYNC`, `INTERLEAVED BANTER`, `ROUND 1`, `ROUND 2`, `ROUND 3`, `CLOSING`)
— see "Segments" below.

**Rebuilt 2026-07-19** to the operator's own rewritten score — content/structure only,
same engine. The rebuild surfaced a note-type mapping and two mechanism gaps (not silently
invented around): `brobots_ready_together`'s spoken phrase, and the four-line "attitude"
volley's divider grouping.

Bingo has never had a `story.md`/`knobs.json` before the original pass.

## Transport and shape

Mostly Wire-Pod HTTP, the same shared instrument (`wirepod_web_send_form()`) every other
song in this repo uses. **Two exceptions, both direct-Vector-SDK, both reused not
invented:** the `rattle` note (song-specific binary) and the `brobots_ready_together`
note (the existing opening-chord binary, reused verbatim). See
`DECOUPLED_DIRECT_SDK_GOLDEN_PATH_001.md` for the shared connect→control→act→release
skeleton both are built from. Every other note in this song remains one robot active at a
time, strictly Wire-Pod HTTP. See `BINGO_RATTLE_ADDED_001.md` for the rattle build detail.

Seven note shapes:

- **`animation`** — added 2026-07-25 for the operator's own scored `kgSuccess`/
  `searching`/`answering` insertions (the knowledge-graph animation family, confirmed real
  and working this session): a
  bare animation dispatch, no spoken line at all, no pause/hold structure —
  assume → fire → release. `searching`/`answering` are real loops in `kgsim.go` (re-fired
  every ~0.333s for `hold_seconds`, matching Wire-Pod's own cadence); `kgSuccess`/
  `searchingGetout` are one-shots. Same `print-loudly-and-continue` failure handling as
  `rattle`/`arm_cue`/`nod` — an animation-only beat never stops the run. See
  `run_animation_only()` in `run_golden_song_001.py`.
- **`say_turn`** — plain back-and-forth banter/calls/reactions. Uses `Robots.say()`
  exactly as the interview does: assume → 0.25s settle → speak → release, per turn. Full
  normalization + pronunciation-safety pipeline applies, same as every interview line.
- **`emotion_beat`** — the golden paced shape, four-clean-runs-proven this cycle
  (`FOUR_CLEAN_RUNS_ANGRY_FIXED_001.md`): assume → speak the plain emotion line (bare,
  no embedded animation token) → 2.0s pause → a *separate* bare animation dispatch
  (`{{playAnimationWI||<token>}}`, async, fire-and-forget) → a stuck-animation check
  (read-only, non-blocking) → hold → release. Fully speaker-generic — confirmed before
  reuse that nothing in it is hardwired to Brobot 2; two of the beats run on Brobot 1.
- **`rattle`** — Brobot 1 only, fires ahead of each ball-draw call: Wire-Pod releases the
  robot → settle (past Wire-Pod's own release-polling window) → a standalone direct-SDK
  binary (`direct_sdk_bingo_rattle_001`) connects independently, requests behavior
  control, plays the Bingo sidecar's own `audio/rattle.wav`, releases, and disconnects.
- **`brobots_ready_together`** — both robots say "Brobots ready!" via the exact same
  golden path as `rattle` above, but reuses the *existing* compiled binary
  (`direct_sdk_brobots_ready_001`, the opening chord's own together-step tool) verbatim —
  no new Go file, since that binary already does exactly "two robots, one phrase,
  concurrent." **Fixed 2026-07-19:** the binary invocation used to be a hardcoded
  Python-side argument (`"Brobots ready!"`) regardless of this step's own `text` — flagged
  during the score rebuild, then fixed per the operator's go-ahead:
  `run_brobots_ready_together()` now takes a `phrase` parameter (default `"Brobots
  ready!"`, unchanged for any caller that doesn't pass one), and the step loop passes this
  step's own story.md text through. The binary itself already took the phrase as argv[1]
  — only the Python call site was hardcoded.
- **`arm_cue`** — one robot, N up/down lift cycles (`cycles` in `knobs.json`, per step),
  each direction its own full assume/`move_lift`/hold/stop/release call — ports
  `brobots-lift-up`/`-down`'s own HTTP shape (`brobots.sh`'s `_brobots_move_axis`) to a
  single serial instead of always both robots.
- **`nod`** — one robot, one down-then-up `move_head` cycle, ports
  `brobots-head-nod`'s own shape the same way.

A `wake_both` note fires `/api-sdk/conn_test` on both robots up front, then a 1.5s settle
— the same wake step every reaction-lane test this cycle has used before its first real
dispatch.

**Human pacing:** each step may carry an optional `buffer_after` (seconds) in
`knobs.json`, applied once that step's own work is done and before the next step
starts — a starting tune, meant to be adjusted live. The `reporter_gap_*` steps and
`exit` carry no `buffer_after` key (default 0), so their own pause (or lack of one) is
never doubled. `wake_both` keeps its own `settle_seconds` (1.5s) unchanged — its
`buffer_after` (3.0s) is an additional beat on top, not a replacement, giving ~4.5s from
wake-complete to the together-step.

**Segments:** every step carries an optional `section` (string) in `knobs.json`, pure
organization/metadata — never alters order, timing, or dispatch. The runner prints
`=== BINGO_CAPTURE SECTION: <name> === ` the moment a new section's first step begins, so
a live console/log reads as labeled blocks instead of one flat step list — plus, since
2026-07-19, a labeled `===== <tag>` divider before every individual step too (see
`.claude/skills/score-dividers/SKILL.md`). Sections in order: `OPENING SYNC`,
`INTERLEAVED BANTER`, `ROUND 1`, `ROUND 2`, `ROUND 3`, `CLOSING` — 69 steps total as of the
2026-08-11 notation remap (host_/brat_ prefixes, standard action suffixes, bare
`reporter_gap` separators added at every segment boundary, all three rounds now share one
skeleton, 8 steps cut - see `BINGO_NOTATION_REMAP_EXECUTED_001.md`).

All five `emotion_beat` steps (`beat_frustrated_1`, `beat_veryhappy`, `beat_1`, `beat_2`,
`beat_3`) carry a single comma as their bare spoken line — the operator's 2026-07-19
score names each beat's animation token and speaker but not a spoken line, so this reuses
the existing proven-safe placeholder convention (a prior pass found empty text
problematic; a bare comma was confirmed safe live) rather than inventing new dialogue for
lines the operator didn't write.

## Reporter gaps

Silent, deterministic pauses (`pause` note, reusing `brobots_bait_002`'s own mechanism —
`pause_seconds` in `knobs.json`, no hardware call, 5s each). As of the 2026-07-19 rebuild
there is no separate "please wait / ready" transition section between rounds — each
round now opens with its own `juggling`/`test_begins` line, `rattle`, and a
`reporter_gap_*`, then its own `ready` line before the ball-draw call. `OPENING SYNC` and
`ROUND 1` each carry one reporter gap (`reporter_gap_opening`, `reporter_gap_r1`); `ROUND
2` and `ROUND 3` each carry two (`reporter_gap_r2a`/`reporter_gap_r2b`,
`reporter_gap_r3a`/`reporter_gap_r3b`) — one after that round's rattle, one at the round's
own close (after the `attitude` volley in Round 2, after `round_3_line` in Round 3). Each
gap carries a `gap_label` in `knobs.json` so a video editor knows which gap is which —
these gaps are left open; no reporter voiceover is recorded or played by this song, that's
a separate, later recording lane.

## Stop condition

If an `emotion_beat` step's animation dispatch fails (non-200 HTTP) or the step raises,
the run stops at that beat rather than continuing into the next step — see
`run_golden_song_001.py`'s own stop-on-beat-failure logic. The other beats' results
still stand; this is not a silent failure, it prints plainly and the log records exactly
where it stopped. `rattle`/`brobots_ready_together`/`arm_cue`/`nod` failures do NOT stop
the run — print-loudly-and-continue, since none of these are the scored content itself.

**Corrected 2026-07-18:** this section previously stated the animation-wait-line check
(then called "stuck-animation check") also stopped the run. It never has — confirmed
against the actual code and against the golden run itself, which flagged all six
`emotion_beat` steps and still completed all 42 steps, `stopped_early: false`. Only a
non-200 HTTP result or a raised exception stops the run; see Troubleshooting below for
the full story on the animation-wait-line advisory.

`> TEXT:` is spoken verbatim (or is the note's own line, for `say_turn`/`emotion_beat`).
No `{robot_name}` substitution is used in this song — every line is authored for a
specific speaker already.

## Troubleshooting

### Animation-wait-line advisory — known-noisy by design, not a bug to fix

Every `emotion_beat` step's own log entry may carry `animation_wait_advisory: true`
(field renamed 2026-07-18, was `stuck_animation_suspected`). **This is advisory only,
never a confirmed failure, and it does not stop the run.** Resolved 2026-07-18.

**What it means:** Wire-Pod's own debug log printed the line
`"(waiting for animation to be done...)"` somewhere in its recent window after this
beat's animation dispatch. `ANIMATION_DISPATCH_ISOLATION_001.md` proved this line
*can* mean a stale `AnimationQueues` entry blocking the real animation call — but its
bare presence alone does not prove that happened here.

**Why it fires on clean runs:** this song fires an animation on nearly every step —
every `say_turn` line carries its own default `"thinking"` animation via
`normalize_robot_safe()`, plus each `emotion_beat` step fires its own bare dispatch. With
animations firing this often on both robots, brief queue contention between back-to-back
calls is normal and self-resolving, not evidence of a genuine stuck state.
`SILENT_ANGRY_SAY_ASYNC_FIX_001.md` independently observed the same log line appear "with
no ill effect" on an unrelated fire. On this song's own golden run (pre-rebuild), all six
`emotion_beat` steps flagged the advisory while the run played perfectly
(operator-confirmed, HTTP `200` throughout, all 42 steps reached).

**Why it isn't fixed to be more precise:** no HTTP-exposed way exists to query or reset
Wire-Pod's `AnimationQueues` state, and neither the success path nor the completion
path of that queue's own Go functions log anything — only the blocked path does. With
no corroborating signal available client-side, there's no way to distinguish
"genuinely stuck until a wire-pod restart" from "brief, harmless overlap between two
animations that resolved a moment later" from the log alone. Tightening the check
(e.g. requiring the line to be freshly-timestamped within this dispatch's own window)
would not fix this — the evidence above shows even a fresh occurrence isn't proof of
failure.

**What to actually do:** watch the robot. Operator's eyes are the health gate for this
beat, same as every other visual confirmation in this repo's own testing discipline.
If a beat's animation genuinely doesn't play, that's a live observation to act on
directly (see `robot_control_song_001/story.md`'s own Troubleshooting section for the
general Wire-Pod recovery order) — not something this log line by itself proves either
way.

## SECTION: OPENING SYNC


## STEP reporter_gap_0101

> TEXT:

## STEP brobots_wake_0101

> TEXT:


## STEP brobots_ready_0101

> TEXT: Brobots ready to Bingo!


## STEP reporter_gap_0102

> TEXT:

## STEP host_arm_cue_0102

> TEXT:


## STEP host_test_begins_0102

> TEXT: Bingo Test Begins. Juggling big shiny bingo balls. Please wait.


## STEP host_rattle_0102

> TEXT:


## STEP reporter_gap_0103

> TEXT:


## SECTION: INTERLEAVED BANTER


## STEP reporter_gap_0201

> TEXT:

## STEP brat_arm_cue_0201

> TEXT:


## STEP brat_say_0201

> TEXT: Wait! Do you know what my lucky big shiny Bingo ball number is?


## STEP reporter_gap_0202

> TEXT:

## STEP host_say_0202

> TEXT: Excuse me? We're running a test! I feel frustrated.


## STEP host_beat_frustrated_0202

> TEXT: ,


## STEP host_say_0202_2

> TEXT: What is your lucky big shiny Bingo ball number?


## STEP reporter_gap_0203

> TEXT:

## STEP brat_nod_0203

> TEXT:


## STEP brat_say_0203

> TEXT: It's "I, Won!" Lol.


## STEP brat_beat_veryHappy_0203

> TEXT: ,


## STEP host_say_0203

> TEXT: Cute. In all my nano-seconds doing big shiny bingo ball juggling, I never heard that one before. Let's begin the test, Mr funny pants!


## STEP reporter_gap_0204

> TEXT:

## SECTION: ROUND 1


## STEP reporter_gap_0301

> TEXT:

## STEP host_test_begins_0301

> TEXT: Bingo Test Begins. Round 1! Juggling big shiny bingo balls. Please wait.


## STEP host_rattle_0301

> TEXT:


## STEP reporter_gap_0302

> TEXT:


## STEP host_kg_success_0302

> TEXT:


## STEP host_say_0302

> TEXT: Big Shiny Bingo ball captured.


## STEP host_call_0302

> TEXT: Big Shiny Bingo Ball number, B-1!


## STEP reporter_gap_0303

> TEXT:

## STEP brat_searching_animation_0303

> TEXT:

## STEP brat_arm_0303

> TEXT:


## STEP brat_say_0303

> TEXT: B-1? That's true! I be, the one! Lol!


## STEP brat_beat_celebrate_0303

> TEXT: ,


## STEP reporter_gap_0304

> TEXT:

## SECTION: ROUND 2


## STEP reporter_gap_0401

> TEXT:

## STEP host_test_begins_0401

> TEXT: Round 2! Juggling big shiny bingo balls. Please wait.


## STEP host_rattle_0401

> TEXT:


## STEP reporter_gap_0402

> TEXT:


## STEP host_kg_success_0402

> TEXT:


## STEP host_say_0402

> TEXT: Big Shiny Bingo ball captured.


## STEP host_call_0402

> TEXT: Big Shiny Bingo Ball number, I-9!


## STEP reporter_gap_0403

> TEXT:

## STEP brat_searching_animation_0403

> TEXT:

## STEP brat_arm_0403

> TEXT:


## STEP brat_say_0403

> TEXT: Wait! I-9? No, you maybe 9, not me! Lol!


## STEP brat_beat_celebrate_0403

> TEXT: ,


## STEP reporter_gap_0404

> TEXT:

## STEP host_beat_angry_0404

> TEXT: ,


## STEP reporter_gap_0405

> TEXT:

## SECTION: ROUND 3


## STEP reporter_gap_0501

> TEXT:

## STEP host_test_begins_0501

> TEXT: Round 3! Final Round! Juggling big shiny bingo balls. Please wait.


## STEP host_rattle_0501

> TEXT:


## STEP reporter_gap_0502

> TEXT:


## STEP host_kg_success_0502

> TEXT:


## STEP host_say_0502

> TEXT: Big Shiny Bingo ball captured.


## STEP host_call_0502

> TEXT: Big Shiny Bingo Ball number, O-0!


## STEP reporter_gap_0503

> TEXT:

## STEP brat_say_0503

> TEXT: Wait! O-0? Verifying.


## STEP brat_searching_animation_0503

> TEXT:


## STEP brat_beat_angry_0503

> TEXT: ,


## STEP reporter_gap_0504

> TEXT:

## STEP brat_say_0504

> TEXT: O-zero? This election is rigged! I wanna speak to the manager!


## STEP reporter_gap_0505

> TEXT:


## SECTION: CLOSING


## STEP reporter_gap_0601

> TEXT:

## STEP host_say_0601

> TEXT: I'm the manager and I take all concerns seriously. Test run is over, Mister Big Shiny Bingo Balls!! Lol!


## STEP host_kg_success_0601

> TEXT:


## STEP brat_say_0601

> TEXT: Wait!


## STEP reporter_gap_0602

> TEXT:

## STEP exit

> TEXT:


## Running

Run via the `bingo-video-song` alias (dry) or `bingo-video-song-live` (live, no env var
needed). Both robots run within one continuous invocation — this song is a real
back-and-forth, not the same script played twice.

---
name: alias-mixer
description: Use when a task touches, extends, or asks about GOPOD's growing "mixer board" — the shared cockpit switches (live robots?, reporter gap?, apply phcal tweaks?) that pha0b/phcal ask across the song shelf, and the shared movement mechanism underneath them. Read this before widening a switch to a new song, adding a new switch, or explaining the mixer concept. Successor in spirit to the old ALIAS-MIXER/PIANO docs (absorbed into ALIAS-LIBRARY.md 2026-07-16, deleted 2026-07-23 once zero-caller) — this one documents a real, currently-growing mechanism, not a proposal.
---

# Alias mixer

## 1. What this is

The cockpit (`pha0b`/`pha0b_menu`/`phcal` in `~/.gopod_alias_lib/brobots.sh`) is
growing a small set of shared y/n switches that apply across more than one song,
instead of each song reinventing its own private prompt. This skill tracks which
switches exist, which songs each one currently reaches, and the shared code
underneath — so the next extension mirrors the existing pattern instead of
re-deriving it. Full origin survey: `REPORTER_GAP_SHARED_SWITCH_SURVEY_001.md`
(`gopod_notes/`) — read that for the full file:line evidence this skill
summarizes.

Operator framing: this is explicitly part of the niche-buzz build-in-public
story — a visibly-improving, coherent instrument panel behind the show is its
own kind of "rescued, local, mine" engineering narrative. See `niche-buzz` for
the campaign framing; this skill stays mechanical.

## 2. The five switches, current state

| Switch | Prompted where | Songs/aliases it reaches | Mechanism |
|---|---|---|---|
| **live robots?** | `live_robots_prompt()`, called from both `pha0b()` and `phcal()` (`brobots.sh`) | every song, every call — the one switch already fully shared | Sets `GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1` for the call, unconditionally asked |
| **apply reporter gaps to this range?** | `pha0b_menu()` only (not `pha0b()` itself — a direct `pha0b bingo <a> <b>` call skips it) | `bingo` and `bait` (awaken) — widened 2026-07-25 | `n` sets `GOPOD_BINGO_REPORTER_GAP_OVERRIDE=0`. `y` asks a second question, "reporter gaps in seconds? [default = 0]" — Enter/empty = `0`, any integer or decimal = that value, non-numeric = one re-ask then falls back to `0`. Read by `run_golden_song_001.py`'s own `"pause"` branch via `GOPOD_GOLDEN_REPORTER_GAP_OVERRIDE` (a different env var than the one this switch also still exports, `GOPOD_BINGO_REPORTER_GAP_OVERRIDE` — kept alongside it for a since-superseded fallback path; neither `run_golden_song_001.py` nor `run_robot_control_song_001.py` actually reads that older name today, confirmed by direct grep). Accepts any float via `float(gap_override_raw)`, `REPORTER_GAP_NUMERIC_ENTRY_001.md`. Neither the prompt nor its echo names a hardcoded seconds value anymore — the only number shown is the one the operator just typed |
| **apply phcal tweaks to this range?** | `pha0b()` | `bingo` and `bait` (awaken) — widened 2026-07-25 | Calls `phcal_apply_001.py <step_id> --yes --knobs <song's own knobs.json>` (renamed from `phcal_apply_bingo_001.py` 2026-08-07) per qualifying step (`arm_cue`/`nod`/`head_nod`), writing `phcal_last.json`'s confirmed `cycles`/`hold_seconds`/`speed` straight into that step |
| **apply phcal tweaks to this test cue?** | `test-arm-cue()`/`test-head-nod()` (`brobots.sh`) — NOT `pha0b`, neither alias is a `pha0b` song | the standalone bench-test aliases only, one primitive at a time — added 2026-07-25 | Calls `phcal_apply_control_song_001.py --yes --target test --primitive arm\|nod`, writing `phcal_last.json`'s confirmed `hold` into `ARM_TEST_LEG_HOLD_SECONDS`/`NOD_TEST_LEG_HOLD_SECONDS` only. Never touches `speed` (nod has a hard motor floor; arm speed isn't even a named constant). `--target`/`--primitive` also support `gesture`/`both`, but nothing calls it that way yet — see below |
| **rich display on console?** | `pha0b()`, right after `live_robots_prompt()` — unconditional, every song, same shape as the live-robots switch, not gated by `$song` like rows 2-4 | every song — universal, added 2026-07-25 (default flipped from off) | Sets/unsets `GOPOD_CONSOLE_RICH_DISPLAY`, read once in `Robots.__init__` (`run_section1_full_live_001.py`) — resolves to `True` unless explicitly `"0"`. An explicit `console_rich_display=True`/`False` from the caller (bingo's own construction still passes `True`) always wins over the env var. Console-only: never changes what the robot actually speaks (`robot_safe_line`), only which of `robot_safe_line`/`display_text` gets printed to the terminal |

`control`/`weather` still get none of the range-based switches (rows 2-3) —
their arm/head-nod still runs the fixed `gentle_arm_test_cue()`/
`head_nod_test_cue()` choreography.

**Studio rule: every reporter gap defaults to `pause_seconds: 0` in every song's
own `knobs.json`, no exceptions.** Not a suggestion, not a per-song judgment
call — silent, deterministic, no hardware call, left open for a later
edited-in reporter voiceover, exactly `brobots_bingo`'s own gaps and (as of
2026-07-25) `brobots_awaken`'s. A reporter gap is never a live dead-air pause
during capture; slowing a sequence's actual felt tempo is a
post-production/editing decision, not a live `pause_seconds` value. When
authoring a new song's own reporter-gap steps, start at `0` — the cockpit's
own reporter-gap switch (row 2 above) exists to *override away from* this
default for a specific test range, not the other way around.

**The TEST/GESTURE split is deliberate, not an oversight.** `phcal_apply_control_song_001.py`
can also write `ARM_GESTURE_LEG_HOLD_SECONDS`/`NOD_GESTURE_LEG_HOLD_SECONDS` —
the constants behind the interview's own live `arm_gesture`/`head_nod_gesture`
movement (`fire_scored_interview_movement()`) — but nothing wires that path yet.
A standing memory flags these exact GESTURE sequences as live-confirmed-good,
"don't retune preemptively" — wiring a GESTURE-facing switch is a separate,
explicitly-authorized ask, not a small follow-on to the TEST wiring above. See
`REPORTER_GAP_WIDENED_TO_BAIT_001.md`'s sibling report
(`PHCAL_APPLY_CONTROL_SONG_SPLIT_001.md`) for the full split rationale.

`brobots_interview_section_01` is on none of the song-shelf switches (rows 2-3)
— `pha0b`/`pha0b_menu` refuse it outright (no standalone step-loop runner; it's
driven inline by `generate_phase()`/`playback_phase()`). Getting the interview
onto the mixer at all is a separate, not-yet-started thread — see §5. Its own
GESTURE constants are reachable today only via a direct, by-hand
`phcal_apply_control_song_001.py --target gesture` call, not through any
cockpit prompt.

## 3. The shared movement mechanism

`run_move_axis()` / `run_arm_cue()` / `run_nod()` live in
`goverlord/runtime/songs/tools/run_robot_control_song_001.py` (moved there
2026-07-25 from the bingo runner of that era, which called them via
`control_mod` — the same dynamic-load pattern it already used for
`parse_control_story_md()`/`load_control_song()`; that runner has since been
retired in favor of `run_golden_song_001.py`, below). Both `brobots_bingo` and
`brobots_awaken` read `cycles`/`hold_seconds`/`speed` per step straight from
their own `knobs.json`.

**`manage_control` preserves each song's own control-holding behavior** — this
is the one detail that matters most if this mechanism ever gets ported further:
- `run_golden_song_001.py` calls with the default (`manage_control=True`)
  — assume/release control every single call, the same behavior the bingo
  runner it replaced already had, unchanged.
- `run_robot_control_song_001.py`'s own step loop calls with
  `manage_control=False` — this song's "connect" note assumes control once for
  the whole run, released only at the very end; per-call assume/release here
  would reintroduce the screen-flash bug `say_line()`'s own comment documents
  (`run_robot_control_song_001.py`, near line 476).
- A future song sharing this mechanism needs to know which control-holding
  model it uses before picking a `manage_control` value — don't assume either
  default.

`gentle_arm_test_cue()`/`head_nod_test_cue()` and their
`ARM_TEST_SEQUENCE`/`HEAD_NOD_TEST_SEQUENCE` are untouched, still live, still
back the standalone `test-arm-cue`/`test-head-nod` aliases — a second,
deliberately separate movement path, not part of the mixer.

**Isolated animation-token notes, added 2026-07-28, also not part of the
mixer.** `test-anim-searching`/`test-anim-answering`/`test-anim-kg-success`
(`brobots.sh`, `KG_ANIMATION_GOLDEN_NOTES_001.md`) are the same "one thing,
watch it, judge it with your own eyes" idea as `test-arm-cue`/`test-head-nod`,
but for a bare `playAnimationWI` token instead of a movement sequence — no
switch prompts them, no song's `knobs.json` drives them, they take a plain
`[robot] [hold]` argument pair. Worth naming here because the mechanism they
port from is the same family this section already tracks: the loop-vs-one-shot
call (`_brobots_anim_is_loop_token`) is unchanged, but the loop's own
dispatch-count math is ported straight from `run_songs_runner_001.py`'s
`run_animation_only()` accumulator, not delegated to this file's older
`_brobots_play_anim_single` — that helper's own `repeats=int(hold/0.333)`
formula was found, live, to undercount by one dispatch against bingo's real,
proven count. Full per-alias detail: `ALIAS-LIBRARY.md`.

## 4. How to extend

Widening an existing switch to a new song (the pattern just proven for phcal-
apply → `bait`):
1. Confirm the target song's runner reads the same field(s) the switch writes
   (e.g. does it read `cycles`/`hold_seconds`/`speed` per step, or is it still a
   fixed sequence?). If it's still fixed, that's a bigger change (§3's
   `manage_control` note applies) before the switch means anything.
   Read-first here — don't assume the shape.
2. Widen the bash-side gate (a `case` or added `||` condition on `$song`) at the
   exact spot the switch already fires.
3. If the switch's own script hardcodes a path (like `phcal_apply_bingo_001.py`
   did), generalize it with an explicit override argument, defaulting to the
   original song for backward compatibility — don't silently change existing
   callers' behavior.
4. Update `ALIAS-LIBRARY.md`'s own row for the alias — that file is the
   authoritative per-alias doc; this skill is the pattern-level summary, not a
   replacement for it.
5. Dry-verify both the widened song and the original (unchanged) song before
   calling it done — a shared-mechanism change is exactly the kind of edit
   that can silently break the song you didn't touch.

Adding a brand-new switch: same shape — one prompt, one env var or one apply
script, gated to the songs that actually support it, documented in the table
in §2 and in `ALIAS-LIBRARY.md`.

## 5. Known next candidates (not started, not decided here)

All three song-shelf switches now reach both `bingo` and `bait` — the
reporter-gap switch was the last one still bingo-only, widened 2026-07-25 the
same way phcal-apply was (`REPORTER_GAP_WIDENED_TO_BAIT_001.md`).

- **Wiring the GESTURE side of `phcal_apply_control_song_001.py` somewhere** —
  the tool already supports `--target gesture`, ready to reuse, but nothing
  calls it that way. This would retune the interview's own live
  `arm_gesture`/`head_nod_gesture` movement from a bench-tuning session — a
  standing memory flags these exact sequences as live-confirmed-good, so this
  needs its own explicit go, not an assumed follow-on to the TEST wiring.
- Wiring `brobots_interview_section_01` into `pha0b` at all — its knobs.json
  shape (line-based exchanges) is structurally different from every song
  currently on the mixer; this is a bigger lift than widening a switch, not a
  small mirror-in. Out of scope until asked for directly.

## Scope

- Documents the mixer pattern itself — which switches exist, which songs they
  reach, the shared code, how to extend it. Not a replacement for
  `ALIAS-LIBRARY.md`'s own per-alias rows, which stay authoritative for exact
  current call signatures.
- Procedure/reference only — no code of its own, no automation.
- Keep §2's table current whenever a switch is widened or added; that's the
  section expected to grow.

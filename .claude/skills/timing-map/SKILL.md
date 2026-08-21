---
name: timing-map
description: Use when the operator wants a timing map (cumulative durations, flagged gaps) built from a real net_video_timing_run_*.json self-log, produced by the timing instrumentation added 2026-07-14 (see NET_VIDEO_TIMING_INSTRUMENTATION_001.md). Reads one run's segments in performance order, sums non-overlapping top-level spans into a cumulative table, and flags genuine gaps/margins - never estimates what isn't measured. STARTING POINT ONLY: written from a single real run (interview5, run_id 20260714_231903) - expect to revise once more runs exist.
---

# Timing map

**Early/thin skill** - built from one real run so far. Treat the shape below as a
starting point, not a settled ritual the way `survey-then-commit`/`hardware-calibrate`
are. Revise freely once a second or third run surfaces a pattern this version doesn't
anticipate (multiple vamp rounds actually firing, a `voice_destination=monitor` run,
`GOPOD_RESTART_WIREPOD_BEFORE_RUN=1` making stage1 non-trivial, etc.).

## Input

`goverlord/runtime/songs/<song_dir>/runs/net_video_timing_run_<run_id>.json` - self-written
by `run_section1_full_live_001.py`'s `main()` (and `run_section1_preshow_generate_001.py`
for a generation-only stage) at the end of a run, live or dry. Private/untracked, same as
the rehearsal-run logs - never stage it (see `survey-then-commit`). If more than one
exists, ask which run, or default to the most recent `run_id` and say so.

Top-level shape:
```json
{
  "tool": "net_video_timing_instrumentation",
  "capture_quality_legend": {"true_completion": "...", "accept_only": "..."},
  "meta": {"voice_destination": "...", "live_enabled": true, "preshow_song_dir": "..."},
  "segment_count": 87,
  "segments": [ {"segment_id": "...", "label": "...", "capture_quality": "...",
                  "note": "...", "start_utc": "...", "end_utc": "...",
                  "elapsed_seconds": 0.0}, ... ]
}
```

## Segment families (confirmed 2026-07-14, from the actual instrumented code)

| Prefix | What it is | Nesting note |
|---|---|---|
| `stage1_kokoro_status_warmup` | Brobot 3 Kokoro model warm-up | Appears twice if run via the scored pre-show path (direct call + generate_phase's own no-op recheck) - use the first (earlier `start_utc`) |
| `stage1_kokoro_host4_warmup` | Brobot 4 subprocess spawn | |
| `stage1_wirepod_restart_preflight` | Wire-Pod restart→ready | Near-zero unless `GOPOD_RESTART_WIREPOD_BEFORE_RUN=1` |
| `stage1_robots_connect_esn_announce` | ESN routing announce | |
| `stage2_m{1-4}_{a-h}` | Fixed pre-show host beats | Top-level, no nesting |
| `stage2_m{2,3}_c_llm_colour` + `stage2_m{2,3}_c_speech` | Brobot 1/2 wake reactions | **Sum both** - one beat (m2_c / m3_c) split into two segments |
| `stage2_vamp_round{N}_*` | Individual vamp iterations | Contingent - only exist if generation ran long |
| `stage2_vamp_loop_summary` | Whole vamp-gate span + real `vamp_rounds` count | One synthetic entry, always present, `start_utc`/`end_utc` are `null` (see `note_summary()`) |
| `stage3_line{N}_pregen_total` | Per-line generation (background, concurrent with pre-show) | **Not on the playback critical path** - runs while pre-show plays, don't add to the show-length sum |
| `stage3_line{N}_brobot{1,2}_colour[_attemptN]` | Individual LLM colour-passes, incl. echo/self-repeat retries | Nested inside `pregen_total` |
| `stage3_line{N}_playback_total` | Whole-line playback (both speakers, movement, pre-pause) | **Use this for cumulative show length** - contains the four rows below, don't also add them |
| `stage3_line{N}_brobot{1,2}_movement` | Whole-gesture elapsed, `movement_steps` has per-waypoint detail | Nested inside `playback_total` |
| `stage3_line{N}_brobot{1,2}_speech` | Speech dispatch | Nested inside `playback_total` |
| `stage3_pause_after_line{N}` | Configured `between_exchange_pause_seconds` | Top-level, between lines |

## How to build the cumulative table

1. Pick the **top-level, non-overlapping** segments only, in `start_utc` order:
   `stage1_*` (once each) → `stage2_m1_a..m4_g` (summing the two `m2_c`/`m3_c` halves
   into one row each) → `stage3_line{N}_playback_total` interleaved with
   `stage3_pause_after_line{N}`. Do **not** also add the nested `_movement`/`_speech`/
   `_colour` rows into the same sum - that double-counts.
2. Running cumulative = running sum of `elapsed_seconds` in that order.
3. Sanity-check: sum should land within a fraction of a second of
   `(max end_utc across all segments) - (min start_utc across all segments)` - if it's off
   by more than that, a top-level segment was missed or double-counted somewhere. Compute
   both and say so if they disagree.
4. `stage3_line{N}_pregen_total` and its nested colour-pass children happen **concurrently**
   with stage 2 (a background thread) - report generation's own total span separately
   (earliest pregen `start_utc` to latest pregen `end_utc`) and compare it against the
   pre-show's own span, don't fold it into the same linear cumulative sum.

## Flagging gaps (only what's actually measured)

- **True silent/dead-air gaps**: segments with no spoken/gestured content and no
  configured-pause justification. In the one run so far, only the initial
  `stage1_kokoro_status_warmup` qualified (~10s before any word is spoken).
- **Vamp margin**: `stage2_vamp_loop_summary`'s `vamp_rounds` (0 = generation finished
  before the gate was ever checked) plus the gap between generation's own total span and
  the pre-show's own span - state the margin in seconds, not just "it fit."
- **Longest individual beats/lines** - worth naming since they're the highest-leverage
  spots to trim or accept, not because "long" is inherently a problem.
- **`accept_only` vs `true_completion`** - carry the tag through to the report. Note
  explicitly if an `accept_only` number (Wire-Pod `say_text`/movement calls) still looks
  text-length-correlated/plausible as real duration in practice (worth stating as an
  empirical observation) without ever promising it as guaranteed completion - that
  distinction is the entire point of the instrumentation's tagging.

## Output

A report to `~/crushn8r_git/gopod_notes/` (see `goreport`) with: a markdown table
(segment, duration, cumulative, capture quality), the flagged-gaps section above, and a
one-line statement of which `run_id` / file this was built from. See
`NET_VIDEO_TIMING_MAP_001.md` for the format this was first written in.

## Scope

- Read-only against the timing JSON and the song content needed to label segments
  (`story.md`/`knobs.json` for line/beat text) - never edit either.
- Don't build a map from more than one run's segments merged together unless the operator
  asks for a comparison - state which single run it's from.
- If the timing JSON doesn't exist yet for the run being asked about, that's a
  `dry-verify`/live-run question, not this skill's problem to solve by estimating.

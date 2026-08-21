# Brobots Awaken — merged capture video, out loud

**GOLDEN — 2026-08-10.** Two clean back-to-back live runs, Brobot 1 then Brobot 2, all 16
steps `ok=True` each run, zero failures, zero FAIL-line triggers, weather fetch correct both
times (location + date, per-robot unit format). `runs/golden_run_2026-08-10_16-15-16.log`
(Brobot 1) and `runs/golden_run_2026-08-10_16-16-46.log` (Brobot 2).

**Re-confirmed golden in the current songs format, 2026-08-16** — the operator's own live
`pha0b` run through today's 6-song menu (tools/ bypassed, `[1-6, 0 to exit]` prompt), Point
A→B range playback (`arm_test`..`weather`, steps 4-10 of 16 — pha0b's own "zoom in/out")
confirmed working end to end against `run_golden_song_001.py`: the reporter-gap override
(`0`) honored (every `pause_*` step slept `0.0s`), phcal tweaks applied cleanly to the
picked slice (`arm_test`/`head_nod` both diffed identical, confirmed written to
`zKnobs.json`), the real per-robot weather fetch spoke correctly (Windsor forecast, dry
mode). `runs/golden_run_2026-08-16_11-08-46.log`/`.json`. **This closes the previously
unrecorded "zoom in/out" check** (`GOPOLISHER_WAWN_SWEEP_001.md` flagged it as an orphaned
pin, not found anywhere on disk) — the range-slice mechanism (originally built for the
control-song family, `CONTROL_SONG_PLAYHEAD_LIFT_EXECUTED_001.md`) is proven live against
this song, in its current `run_golden_song_001.py`-driven format, not just the pre-golden
mechanism it started on.

Knobs: [knobs.json](knobs.json)

**Identity swap, 2026-07-24:** this song (formerly `brobots_bait_002`, "merged bait-video
capture song") now carries the `brobots_awaken` name and folder — the operator's own call:
"This is correct — it IS the awaken video." The original pure weather-only 3-step song that
used to live here was renamed `brobots_bait_000` and moved to `zzz_archives/`. **Since
decluttered, 2026-08-15** (`gopod_notes/ZZZ_ARCHIVES_DECLUTTER_EXECUTED_001.md`) — it was
pure scratch content, nothing left to revive, no longer on disk.

Same single-robot mechanism as
[robot_control_song_001](../zzz_archives/robot_control_song_001/story.md) and
`brobots_bait_000` (the weather note this song's own `weather` step reuses, now
decluttered — see above), reused as-is. One straight line of
notes, run once per robot (Brobot 1 first, then Brobot 2): connect, a silent reporter-gap
pause (so the robot has settled before the first word is spoken), a spoken connect lead-in
(restored from `robot_control_song_001`'s own shape), arm test (self-narrated), a spoken
gap cue then a reporter-gap pause, head nod test (same shape), another spoken gap cue
then a reporter-gap pause, the real per-robot weather fetch (including location and date),
another spoken gap cue then a reporter-gap pause, the party/self-ID payoff line, a spoken
show-outro cue then a final reporter-gap pause, then a clean exit.

Updated 2026-07-16 (second review pass) per operator timing/wording review: all five pauses
(the new post-connect warm-up included) are now 7s (was 10s, no warm-up pause existed
before this pass), and every spoken gap cue's wording was tightened - head_nod_done now
says "Loading weather check" (was "Loading next test"), weather_check_done now says "Tests
complete" (was "Loading next test"), matching how far along the run actually is at each
cue. Party line and weather content are unchanged from the prior pass.

Updated 2026-07-25 (first pass) per operator review of a live run ("perfect run, just
remove the please wait for 7s - the runner is rushing the performance"): the "Please wait
7 seconds" tag dropped from all three gap cues (arm_test_done, head_nod_done,
weather_check_done).

Updated 2026-07-25 (second pass), operator: "remove all stale 7s anything/everything from
the song. Then slow the sequence tempo down. No 7s gap = 0s." Show_outro's own "Exiting in
7 seconds" tag dropped too (now just "That's the show."). All five pauses (knobs.json)
dropped from 7s to 0s - same reporter-gap convention `brobots_bingo`'s own gaps already use
(silent, no hardware call, left open for a later edited-in reporter voiceover, not live
dead-air during the robot's own capture). Slowing the sequence's actual tempo is a
post-production/editing question, not a live-pause-duration one - matches how bingo's own
reporter gaps already work.
reverting to the pre-2026-07-16 value this song's own history above already names, not a
new invented number. No spoken text anywhere in this song mentions a pause duration anymore.

`> TEXT:` is spoken verbatim before the note plays (or is the note itself, for a plain
`say`). `> FAIL:` is spoken only if that note's own hardware call didn't come back clean.
`{robot_name}` in a `> TEXT:` line is substituted at runtime with this run's own robot name
("Brobot 1" or "Brobot 2") - the only per-robot difference in this song's text; everything
else is identical for both runs. The `pause` notes are silent, deterministic sleeps (no
hardware call) sized by `pause_seconds` in `knobs.json` - reporter-gap time, not test
content. The `weather` step's `weather_include_location_date: true` flag (in `knobs.json`,
this song only) is what turns on the location+date prefix in
`gopod_weather_fetch_001.py`'s `format_for_robot()` - every other song leaves that flag
unset and gets the old, shorter weather line unchanged.

## STEP connect
> TEXT:

## STEP pause_after_connect
> TEXT:

## STEP say_connected
> TEXT: I'm connected. Loading next test.

## STEP arm_test
> TEXT: Testing my arm.
> FAIL: My arm didn't respond.

## STEP arm_test_done
> TEXT: Arm test done. Loading next test.

## STEP pause_after_arm
> TEXT:

## STEP head_nod
> TEXT: Testing my head nods.
> FAIL: My head didn't respond.

## STEP head_nod_done
> TEXT: Nods test done. Loading weather check.

## STEP pause_after_head_nod
> TEXT:

## STEP weather
> TEXT:

## STEP weather_check_done
> TEXT: Weather check done. Tests complete.

## STEP pause_after_weather
> TEXT:

## STEP party
> TEXT: {robot_name} Ready to party! Did someone say GOPOD Yourself?

## STEP show_outro
> TEXT: That's the show.

## STEP pause_after_party
> TEXT:

## STEP exit
> TEXT:

## Pronunciation

The party line above is authored as literal "GOPOD Yourself" - that is the correct,
final text, not a placeholder. `say_line()`'s own `normalize_robot_safe()` step has no
pronunciation handling, but that's not the whole path: `say_line()` hands off to
`robots.say()`, and `robots` is `mod.Robots(...)` - the exact same shared class the
interview module instantiates, not a parallel implementation. That shared `Robots.say()`
unconditionally runs `flatten_for_robot_speech()` -> `apply_pronunciation_safety()`
on every line, from either lane, before building what a live call actually sends to
Wire-Pod. So the registry swap ("GOPOD" -> "Gowp-awd") applies here automatically, at
speech time only - this song's text, the terminal print, and story.md all keep the
literal "GOPOD" spelling; only the actual speech payload is corrected. Confirmed live,
not assumed - see `run_robot_control_song_001.py`'s own module docstring (PRONUNCIATION paragraph) and
`say_line()`'s `PRONUNCIATION_SWAP` proof print. Never author a phonetic respelling in
this song's own text - it isn't needed.

## Running both robots

Run via `pha0b` -> pick `00_brobots_awaken` (keyword `bait`) - the golden path
(`run_golden_song_001.py`, cut over 2026-08-07). Full-range picks Brobot 1's own whole-run
robot choice by default; `pha0b`'s own robot prompt selects which. Dry by default; the
live-robots prompt exports `GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1` for that run only.
`start-the-bait-song` (the old legacy-engine alias) was retired 2026-08-11 - stale, no
crucial use, easily rebuilt if ever needed again.

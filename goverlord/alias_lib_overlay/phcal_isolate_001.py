#!/usr/bin/env python3
"""phcal Rung 3 - isolate-and-watch, ONE mechanical primitive, one robot, then
stop, now with cycles/hold/speed adjustable per call AND a guided prompt flow
plus write-back memory on top. Sibling to pha0b: pha0b plays the SCORE (dry
navigation of a song slice); phcal tunes the INSTRUMENT (live motion in
isolation). Rung 1 isolated and played current behavior only, no value
changed. Rung 2 added --hold/--speed/--cycles flags (and nod's existing
positional [count]) to override the rung-1 defaults per call - the defaults
themselves unchanged. Rung 3 adds:

  - A guided flow, triggered by calling `phcal` with no primitive at all:
    asks arm-or-nod, then a robot, then walks each value (cycles/count,
    hold, speed) with the prompt pre-filled by the last-used value for that
    primitive - Enter keeps it, typing a new number overrides it. Fires the
    same way the direct-flag form already does (live if the gate's on, dry
    if it's off), then saves the values actually used.
  - Write-back memory at ~/.gopod_alias_lib/phcal_last.json (this tool's own
    memory file, outside the tracked GOPOD repo - NOT any song's knobs.json/
    story.md). Separate keys per primitive (arm: cycles/hold/speed; nod:
    count/hold/speed). Both entry paths - the new guided flow AND the
    existing direct-flag form (`phcal arm 1 --hold 0.5 ...`) - write to this
    same one file, so a value set either way becomes the next [last]. First
    run with no file on disk seeds [last] from the same rung-1 bingo-sourced
    defaults below, then saves after that first run.

Sourced from, not a copy of: run_songs_runner_001.py's own
run_move_axis()/run_arm_cue()/run_nod() (single-serial per-robot HTTP shape:
assume_behavior_control -> move_lift/move_head speed set -> hold -> speed=0
stop -> release_behavior_control). Reproduced with a printed, timestamped line
in front of every sub-instruction, so the operator can watch each mechanical
step land, whatever values (default or flag-overridden) are actually in play
- that source file's own functions return a result dict only, no line-by-line
print, which rung 1 added and rung 2 keeps.

2026-07-22 control-holding fix: assume/release_behavior_control now fire
ONCE per cmd_arm/cmd_nod call (before/after the whole cycle loop), not once
per direction/cycle - matching run_named_movement_sequence's own
(run_robot_control_song_001.py) assume-once/release-once shape, the same
fix that keeps the robot's screen from flashing to its home icon between
notes during a full song run. See run_move_axis_no_control's own comment
block, below, for the full root-cause writeup. Rung-1 (bingo-sourced)
defaults, unchanged, used whenever a flag is omitted:
  arm_cue: cycles=1, hold_seconds=1.2, speed=2 (move_lift, +speed up then
           -speed down)
  nod:     hold_seconds=0.35, speed=2          (move_head, -speed down then
           +speed up, in that order - the source code's own down-then-up
           shape; rung 1 confirmed this against the real bingo runner order,
           correcting an earlier spec's prose that wrongly said
           "up -> down" - kept as down-then-up here, do not revert).

Flags (rung 2, new):
  --hold S     hold_seconds override (float, > 0). Valid for arm and nod.
  --speed N    speed-magnitude override (int, > 0; applied as +N/-N per
               direction, same convention as the default). Valid for arm
               and nod.
  --cycles N   arm-cue repeat-count override (int, >= 1). arm-only - nod
               repeats via its existing positional [count] instead, not a
               flag, so there is exactly one way to say "how many" per
               primitive.

NOT a song runner: no song folder, no knobs.json, no story.md, no step loop,
nothing read from or written to any song's own files - tuning stays IN this
tool, it is never written back to a song's own files. One primitive, one
robot, then exit.

Live gate: same convention as every other live alias in
~/.gopod_alias_lib/ - GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 must be exported by the
caller for real HTTP calls to fire. Unset (or anything else) = dry: prints
the exact same planned sub-instructions (reflecting whatever values are in
play), sends nothing, holds nothing.

Hard limits kept from rungs 1/2, unchanged by rung 3: alias-lib only, no
repo/song/knobs.json/story.md touch; one robot per call, no two-robot sync;
nod stays down-then-up.

Rung 4 (2026-08-08, GOLDEN_BROBOTS_CONTROL_CATALOG_SURVEY_001.md /
GOPOD_WIDGET_TARGET_SHAPE_001.md's "instrument rack" step): the guided
flow's primitive pick is now a dynamic vertical numbered menu built from one
dict (_PRIMITIVE_MENU), widened from 4 entries (arm/nod/rattle/weather) to 7
by adding sleep/wake/brobots_ready - the catalog's other standalone-fireable
golden mechanisms, each firing through the exact same direct-SDK binary its
existing brobots.sh/core.sh alias already calls (no second implementation).
None of these three have a dry mode (confirmed against the Go source - no
--dry flag exists anywhere in that binary), so each refuses cleanly
(PHCAL_NO_DRY_MODE) rather than faking one when the live gate is off. Still
NOT a song editor: no knobs.json touched, no song wired to a catalog pick,
arm/nod/rattle/weather's own firing/tuning behavior is byte-unchanged - only
how a mechanism is SELECTED changed.
"""
import importlib.util
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
import termios
import time
import tty
import urllib.parse
import urllib.request
from pathlib import Path

# 2026-08-18, PHCAL_DEHARDCODE_001.md: pure config externalization, no
# robot-behavior change - only WHERE the base directories below come from.
# phcal_config.json (this file's own sibling, gitignored - real operator
# values) is read if present; phcal_config.example.json (committed,
# placeholder values) documents the shape for a new clone. Missing file ->
# every base dir below falls back to the exact literal it was hardcoded to
# before this config layer existed, so a fresh clone with no config yet
# behaves identically to pre-config phcal (just not portable until filled
# in). This is the multi-ESN + path foundation a future detect-first pass
# will read from, not a detect-first implementation itself.
_PHCAL_CONFIG_PATH = Path(__file__).resolve().parent / "phcal_config.json"


def _load_phcal_config():
    if not _PHCAL_CONFIG_PATH.is_file():
        return {}
    try:
        return json.loads(_PHCAL_CONFIG_PATH.read_text(encoding="utf-8"))
    except (json.JSONDecodeError, OSError):
        return {}


_PHCAL_CONFIG = _load_phcal_config()
_BASE_DIRS = _PHCAL_CONFIG.get("base_dirs", {})


def _base_dir(key, default):
    return Path(_BASE_DIRS.get(key, default))


# The two real base dirs every hardcoded path below used to spell out in
# full. A third "alias-lib home" base dir is deliberately NOT configured -
# phcal_config.json/tempo_set_001.py/phcal_last.json/phcal_run.log all live
# next to phcal_isolate_001.py itself, so Path(__file__).resolve().parent
# already finds them correctly on any machine, no config value needed.
# (sdk_home used to be a third base dir here, for rattle.wav's own old
# SDK-clone path - removed 2026-08-18, PHCAL_DETECT_FIRST_001.md, once
# audio consolidated to GOPOD_AUDIO_DIR below left it with zero callers.)
GOPOD_REPO_ROOT = _base_dir("gopod_repo_root", "/home/goverlord/crushn8r_git/GOPOD")
WIREPOD_HOME = _base_dir("wirepod_home", "/home/goverlord/wire-pod")


def _resolve_brobot_serials():
    """Brobot 1/2 ESN resolution, in priority order: phcal_config.json's
    robots.candidates (new, N-capable - the list a future detect-first pass
    will probe) if it has at least 2 entries; else the same
    GOPOD_BROBOT_1_SERIAL/GOPOD_BROBOT_2_SERIAL env vars
    run_section1_full_live_001.py has always read; else the same hardcoded
    defaults that module has always fallen back to. Every tier resolves to
    the identical two ESNs for this operator's own real setup today - only
    where they're read from changes."""
    candidates = _PHCAL_CONFIG.get("robots", {}).get("candidates", [])
    if len(candidates) >= 2:
        return candidates[0]["esn"], candidates[1]["esn"]
    return (
        os.getenv("GOPOD_BROBOT_1_SERIAL", "0dd1b9e9"),
        os.getenv("GOPOD_BROBOT_2_SERIAL", "0dd1d8bf"),
    )


BROBOT_1_SERIAL, BROBOT_2_SERIAL = _resolve_brobot_serials()

RUNNER_PATH = GOPOD_REPO_ROOT / "goverlord/runtime/songs/tools/run_section1_full_live_001.py"

# 2026-08-11, tempo calibration handover (operator direction: fold tempo-set
# into phcal's own menu). No new tempo logic here - hands off to the SAME
# tempo_set_001.py module brobots.sh's standalone `tempo-set` alias already
# uses (cmd_set_global/cmd_set_buffer), loaded the same way RUNNER_PATH is
# loaded above. This branch only owns the song/mode/step picking UI; every
# actual read/write stays in that one tool.
TEMPO_SET_PATH = Path(__file__).resolve().parent / "tempo_set_001.py"
# 2026-08-15, PHCAL_WEATHER_ROBOT_SPEAK_001.md: the weather primitive's own
# robot-pick + speak reuses run_robot_control_song_001.py's already-proven
# run_single_note("weather", ...) - the SAME function the standalone
# gopod-weather-say alias already calls (real Windsor fetch, per-robot
# unit/clock via gopod_weather_fetch_001.py's load_robot_format_from_jdocs,
# spoken once). Loaded the same way TEMPO_SET_PATH/KNOBS_ENVELOPE_PATH are
# loaded above - not a second implementation of any of it.
CONTROL_SONG_RUNNER_PATH = GOPOD_REPO_ROOT / "goverlord/runtime/songs/tools/run_robot_control_song_001.py"
KNOBS_ENVELOPE_PATH = GOPOD_REPO_ROOT / "goverlord/runtime/songs/tools/knobs_envelope_001.py"
SONGS_DIR = GOPOD_REPO_ROOT / "goverlord/runtime/songs"

# 2026-08-18, PHCAL_DEHARDCODE_COMMIT_001.md: canonical GOPOD audio home,
# INSIDE the repo tree - goverlord/wire_pod_overlay/chipper/sounds/ is the
# overlay source of truth (what apply_nongo_files.sh mirrors OUT to
# ~/wire-pod/chipper/sounds/), not the live mirrored copy. Every .wav asset
# phcal fires now reads directly from here, derived from GOPOD_REPO_ROOT
# (already config-driven above) - not a new hardcode. Killed the dependency
# on the separate SDK-clone tree for audio specifically (rattle.wav was
# copied in from there that pass) - SDK_HOME itself removed entirely
# 2026-08-18 (PHCAL_DETECT_FIRST_001.md) once this left it with zero callers.
GOPOD_AUDIO_DIR = GOPOD_REPO_ROOT / "goverlord/wire_pod_overlay/chipper/sounds"

# 2026-07-22, Part C (rattle primitive): same standalone direct-SDK rattle
# binary + outbound settle margin run_songs_runner_001.py's own run_rattle()
# uses - reproduced verbatim here, not imported. `mod` (this file's loaded
# module) is run_section1_full_live_001.py, which defines neither of these
# constants; the bingo runner is a different module phcal does not load.
# Same binary, same 1.0s margin (DECOUPLED_DIRECT_SDK_GOLDEN_PATH_001.md's
# own outbound-settle number) - not a fresh guess. The WAV asset itself
# used to read from a separate SDK-clone path (the now-removed SDK_HOME);
# repointed 2026-08-18 to the canonical GOPOD_AUDIO_DIR above.
RATTLE_BIN_PATH = WIREPOD_HOME / "chipper/gopod_probes/tools/direct_sdk_bingo_rattle_001"
RATTLE_WAV_PATH = GOPOD_AUDIO_DIR / "rattle.wav"

# 2026-08-12 (danger primitive): the direct-SDK binary above takes <serial>
# <wav_path> [volume] - it was never rattle-specific, just always called with
# rattle's own WAV path. Reused verbatim (RATTLE_BIN_PATH, not a redundant
# second constant) for a second WAV asset - GOPOD's own playSound sound
# (soundMap in kgsim_cmds.go). Same reasoning as rattle's own comment above:
# a song needs this fireable directly, not gated behind the LLM choosing to
# emit {{playSound||...}} in a chat response - see
# DOPLAYSOUND_REAL_IMPLEMENTATION_AND_RATTLE_FIT_001.md's own verdict on why
# the direct-SDK route beats DoPlaySound for scripted use. Repointed
# 2026-08-18 from the live wire-pod mirror (WIREPOD_HOME) to the canonical
# GOPOD_AUDIO_DIR overlay source above - phcal now reads the source file
# directly, not the copy apply_nongo_files.sh mirrors out.
DANGER_WAV_PATH = GOPOD_AUDIO_DIR / "danger-will-robinson.wav"
# 2026-08-15 (cube primitive): a separate binary, not RATTLE_BIN_PATH reused
# - direct_sdk_cube_blip_001 takes only <serial>, no wav/volume argument at
# all (connect -> all-corners red -> hold -> all-corners green -> release),
# so it doesn't fit the rattle/danger calling shape. See
# CUBE_BLIP_TOOL_BUILT_001.md / CUBE_BLIP_ALIAS_PHCAL_WIRED_001.md.
CUBE_BLIP_BIN_PATH = WIREPOD_HOME / "chipper/gopod_probes/tools/direct_sdk_cube_blip_001"
WIREPOD_HOME_PATH = str(WIREPOD_HOME)
DIRECT_SDK_RELEASE_SETTLE_SECONDS = 1.0

# Rung 4 (GOLDEN_BROBOTS_CONTROL_CATALOG_SURVEY_001.md / instrument-rack
# fattening pass): sleep/wake/brobots_ready are golden, live-confirmed,
# standalone-fireable mechanisms per the catalog - not song-embedded notes,
# so they widen phcal's own bench menu rather than the golden song engine.
# Same binaries `robot-sleep`/`robot-wake` (brobots.sh) and
# `_gopod_chord_direct_together_job` (core.sh) already call - reused
# verbatim here, not a second implementation of the sleep/together logic.
SLEEP_BIN_PATH = WIREPOD_HOME / "chipper/gopod_probes/tools/direct_sdk_robot_sleep_001"
TOGETHER_BIN_PATH = WIREPOD_HOME / "chipper/gopod_probes/tools/direct_sdk_brobots_ready_001"

# 2026-08-10 (BROBOT_2_INSTABILITY_EXTERNAL_AI_BRIEF_001.md's own concrete
# next step): read-only VersionState/ProtocolVersion/BatteryState snapshot,
# no BehaviorControl needed at all - simpler/lower-risk than every other
# direct-SDK primitive above. Same binary brobots.sh's own robot-info alias
# already calls. Timeout matches that alias's own external `timeout 30`
# wrapper - a real buffer over the binary's internal 20s-per-RPC deadline,
# not a guess.
ROBOT_INFO_BIN_PATH = WIREPOD_HOME / "chipper/gopod_probes/tools/direct_sdk_robot_info_001"
ROBOT_INFO_TIMEOUT_SECONDS = 30

# 2026-08-18, detect-first's own probe timeout - deliberately separate from
# ROBOT_INFO_TIMEOUT_SECONDS above, not a shortened copy of it.
# ROBOT_INFO_TIMEOUT_SECONDS's 30s is a real buffer over this same binary's
# internal ~20s-per-RPC deadline (see the comment above it) - correct for a
# DELIBERATE, occasional "1. info" pick where a human explicitly asked and
# expects to wait. Detect-first fires this same binary AUTOMATICALLY on
# every single phcal launch, silently, before the menu ever draws - live-
# tested by the operator, an absent robot at 30s read as a dead freeze
# (real Ctrl-C, twice). Every present robot in that same live test
# responded in ~0.5-0.6s. 8s here is a judgment call, not a proven number
# like the battery-volts threshold below is - generous enough (roughly
# 13-16x normal observed latency) to tolerate a genuinely slow-but-present
# robot without a false "not present," short enough that one absent
# candidate doesn't stall every startup for anywhere near 30s. Confirm/tune
# against real conditions if it ever mis-reads a slow-but-present robot as
# absent.
PHCAL_DETECT_PROBE_TIMEOUT_SECONDS = 8

# 2026-08-08, live-observed floor: the binary's legacy --hold path computes
# ONE shared release deadline up front, BEFORE either robot starts firing -
# releaseAt = start + getInSettleSeconds(2.7, fixed in the Go source) + hold
# (direct_sdk_robot_sleep_001.go:469). A robot whose own GoToSleepGetIn/
# GoToSleepSleeping round trips are slow (live-observed: 9.625s elapsed on
# one robot, vs 2.988s on the other, same call) can finish firing AFTER that
# deadline already passed - it then releases control with ZERO settle time,
# which live-confirmed wakes it immediately (the exact failure
# ROBOT_SLEEP_DIRECT_SDK_BUILT_001.md's own build history already names:
# "released too fast, robots woke immediately - fixed by adding --hold").
# 5.0 is that history's own proven default (5 live-confirmed rounds) - NOT
# 3.7/3.8: those numbers came from this session's own hold=1.0 postmortem
# (2.7 + 1.0 = 3.7) and are BELOW the already-proven-safe default, not a
# tighter floor - using them here would reintroduce more risk than today's
# default already carries, not less. Floored at the number actually proven
# golden, not a lower one that merely sounds close to the math just
# explained.
SLEEP_MIN_HOLD_SECONDS = 5.0

# Rung 5 (PHCAL_CANDIDATE_CONTROLS_SURVEY_001.md, Part 1/Part 2): two more
# standalone bench primitives, both confirmed liftable with zero song/knobs
# dependency stripped out (there never was one to strip - these dispatch
# functions never took a step/song object).
#
# Animation primitive - sourced from run_golden_song_001.py's own
# run_animation_only()/_LOOP_ANIMATION_TOKENS/_ANIMATION_LOOP_INTERVAL_SECONDS
# (not imported - phcal loads run_section1_full_live_001.py as `mod`, not the
# golden engine module those actually live in). Same loop-vs-single-fire
# split, same 0.333s loop interval, same {{playAnimationWI||TOKEN}} payload
# fired via /api-sdk/say_text - the survey confirmed this dispatch is a bare
# HTTP call, not welded to any live KG/LLM session, so it lifts verbatim.
_LOOP_ANIMATION_TOKENS = {"searching", "answering"}
_ANIMATION_LOOP_INTERVAL_SECONDS = 0.333

# The three catalog-confirmed, real animation tokens (read directly off
# 101_brobots_bingo_test/knobs.json, not invented) - this submenu's own numbering,
# separate from _PRIMITIVE_MENU's.
_ANIMATION_TOKEN_MENU = {
    "1": "kgSuccess",
    "2": "searching",
    "3": "answering",
}
# Same default run_golden_song_001.py's own animation note branch uses
# (step.get("hold_seconds", 2.5)) - not a fresh guess.
ANIMATION_DEFAULT_HOLD_SECONDS = 2.5
# 2026-08-15 (PHCAL_ANIMATION_ORDER_PAUSE_001.md): a fixed settle pause
# BETWEEN tokens in the "0 = all in sequence" run only - back-to-back
# firing (searching's getout ending the same instant answering starts)
# was overlapping sprites on the head screen. Single-token runs never
# reach this constant at all.
_ANIMATION_SEQUENCE_PAUSE_SECONDS = 1.0

# 2026-08-08, live finding: release_behavior_control releases the SDK's
# control PRIORITY - it does not cancel an animation already in flight on the
# robot. A loop-class token's last-fired instance plays out its own natural
# tail after the loop stops re-firing it, regardless of release. A live bench
# test of "searching" required a physical backpack-touch to actually stop it.
# Confirmed real fix, not a guess: ~/wire-pod/chipper/animation_vocab.json
# (Wire-Pod's own truth, read-only, never copied here) has a verified
# "searchingGetout" -> anim_knowledgegraph_searching_getout_01 entry - the
# exact same one-shot transition run_section1_full_live_001.py's own
# fire_interview_kg_search_sequence() already fires to end the interview's
# own searching loop cleanly (searching -> searchingGetout -> answering).
# Checked the same vocab file directly for an "answeringGetout" equivalent -
# none exists. "answering"'s own trailing tail is addressed differently, not
# by a getout clip - see _ANIMATION_NO_GETOUT_RELEASE_SETTLE_SECONDS below.
_ANIMATION_GETOUT_TOKENS = {
    "searching": "searchingGetout",
}

# 2026-08-16 (ANSWERING_NATIVE_STOP_DISCIPLINE_001.md): confirmed against
# Wire-Pod's own kgsim_stream.go that native Wire-Pod never needed a getout
# for "answering" because it never cuts the clip mid-play - it stops
# re-firing, waits for the in-flight cycle to actually finish (a channel
# confirmation from the gRPC PlayAnimation call), settles 100ms, THEN
# releases control. phcal's say_text dispatch is fire-and-forget over HTTP,
# not a blocking call like native's, so there's no cycle-completion signal
# to wait on - this settle has to cover the last-fired instance's own
# play-out tail, not just a network round-trip, hence longer than native's
# 100ms. 1.0s matches this bench's own existing settle convention
# (PHCAL_CYCLE_SETTLE_SECONDS) rather than inventing new math. Applies to
# ANY loop-class token with no confirmed getout (today, just "answering") -
# release is deferred by this settle instead of firing immediately.
_ANIMATION_NO_GETOUT_RELEASE_SETTLE_SECONDS = 1.0

# Hold-test primitive default - a plain, watch-worthy duration long enough to
# see whether the robot wanders while control is held, short enough not to
# waste bench time. Not tied to SLEEP_MIN_HOLD_SECONDS above (different
# mechanism, different reason for its own floor) - a fresh, undemanding
# default for a diagnostic-only primitive.
HOLD_DEFAULT_SECONDS = 5.0

# Rung 6 (2026-08-09, WHEEL_NUDGE_GOLDEN_PATH_SURVEY_001.md): the first wheel
# primitive - Stage 1 of that survey's 3-stage golden path only (bench-fire +
# confirm; no alias, no engine note type, no awaken step here). ON-CHARGER
# reverse pulse only - assuming behavior control disables the cliff sensors
# for as long as it's held, which is safe on the charger stand (nowhere to
# fall) and genuinely dangerous off it. This tool has no way to confirm the
# robot is actually on its charger - that's a physical/procedural guarantee
# the operator holds, not something phcal can verify, same gap the survey
# names for every other motion primitive here.
#
# lw/rw values sourced from Wire-Pod's own webroot control page
# (chipper/webroot/sdkapp/js/control.js, the S-key/backward binding), not
# invented - that's the same reverse magnitude the reference control surface
# itself sends. Fixed reverse only for this build; forward is a later,
# separate decision per the survey.
# RENAMED 2026-08-10 (stale-end sweep, per operator direction): these were
# named WHEEL_NUDGE_* back when wheel_nudge (the composed primitive) still
# existed as its own menu entry. wheel_nudge was removed outright this same
# day - these 5 names are used exclusively by move_reverse now (grep-
# confirmed before renaming), so the old prefix was live-misleading, not
# just historical - move_reverse's own log lines were printing tag
# "PHCAL_WHEEL_NUDGE" for a primitive with no wheel_nudge menu entry left.
# Does NOT touch the shared golden-flag mechanism's own WHEEL_NUDGE_WAKE_*
# names below (run_wheel_nudge_wake_check/run_wheel_nudge_wake_release_
# pulse) - those are genuinely shared (move_reverse's chain-wake AND
# brobots_wake both call them), a real naming-scope question, not a stale
# end from this session's own edit - left alone, out of scope here.
MOVE_REVERSE_LW = -150
MOVE_REVERSE_RW = -150
MOVE_REVERSE_DEFAULT_HOLD_SECONDS = 0.3
# Hard cap, refused not clamped (see _parse_float_flag's max_value below) -
# a fat-fingered long hold must never reach move_wheels.
MOVE_REVERSE_MAX_HOLD_SECONDS = 1.0
MOVE_REVERSE_CAUTION = (
    "MOVE_REVERSE CAUTION: cliff sensors disabled while control held - ON-CHARGER USE ONLY"
)

# WAKE-SETTLE, REPLACED 2026-08-10 (WHEEL_NUDGE_COLD_FIRE_ROOT_CAUSE_SURVEY_
# 001.md + its live-diagnostic follow-up sessions). The prior say_text-based
# "speak-ready" check below was proven wrong by direct measurement, not
# superseded by theory: 4/4 live runs against a freshly-released robot
# showed say_text returning in ~0.05-0.12s with NO audible speech - it was
# never actually confirming anything, on either robot. The real, measured
# mechanism: assume, then RELEASE behavior control right back (a "release
# pulse"), settle, then RE-ASSUME. Releasing control hands it back to the
# robot's own onboard behavior just long enough for it to visibly/physically
# wake on its own - this mirrors ROBOT_SLEEP_DIRECT_SDK_BUILT_001.md's own
# first live finding almost exactly (releasing BehaviorControl right after a
# forced sleep trigger wakes the robot on its own, no explicit wake trigger
# needed) - generalized here from "coming out of a forced sleep" to "any
# cold/just-woken control session." Live-proven 8/8 runs across BOTH robots
# (once Brobot 2's own separate charging-contact fault was found and fixed -
# a real, physical, unrelated defect this same investigation surfaced):
# head-pop-to-visible-wake measured at 1.03-1.82s every time, the first
# move_wheels call succeeding at just 0.5s after re-assume in every single
# run, no misses. Same replacement already made in
# run_robot_control_song_001.py's own run_brobots_wake(). Wheel-specific
# only - run_assume_control itself (shared by cmd_arm/cmd_nod/cmd_hold) is
# untouched.
WHEEL_NUDGE_WAKE_CHECK_TIMEOUT_SECONDS = 15  # CONNECTION_PASS_TIMEOUT_SECONDS's own value, matched - conn_test's own timeout, unchanged
WHEEL_NUDGE_WAKE_RELEASE_SETTLE_SECONDS = 2.5  # measured max observed 1.817s across 8 live runs, both robots - real margin above that ceiling, not the bare minimum; see WHEEL_NUDGE_COLD_FIRE_ROOT_CAUSE_SURVEY_001.md for the run-by-run numbers
# BUG FOUND AND FIXED 2026-08-10: nothing ever settled reassume->move. The
# probe that proved this mechanism never fired move_wheels at 0s after
# re-assume (its own sweep starts at 0.5s, and 8/8 clean runs succeeded at
# exactly that first attempt - WHEEL_NUDGE_COLD_FIRE_ROOT_CAUSE_SURVEY_001.md).
# That gap was missing here - live-confirmed 2026-08-10: wheel reversal
# silently no-opped on both robots even though every log line read ok=True.
# Same fix as run_robot_control_song_001.py's own
# BROBOTS_WAKE_POST_REASSUME_SETTLE_SECONDS.
WHEEL_NUDGE_WAKE_POST_REASSUME_SETTLE_SECONDS = 0.5

# 2026-08-08, live finding: separate sleep-then-wake phcal calls showed no
# visible wake effect. Root cause, sourced from real precedent, not guessed:
# ~/.gopod_alias_lib/brobots.sh's own gopod-song-open()/gopod-song-open-chord()
# already root-caused this EXACT symptom in their own build history - a
# release-then-fresh-reconnect gap between two separate BehaviorControl
# sessions lets the robot's own behavior resume on its own before the second
# call ever fires, so a later, freshly-connected wake has nothing left to
# reverse. Their fix: hold ONE continuous connection from GoToSleepSleeping
# straight through to GoToSleepOff via the binary's own --wait-signal mode,
# never releasing in between. Ported that same mechanism into phcal below -
# not a new design, the proven one.
SLEEP_WAKE_DEFAULT_WAIT_SECONDS = 5.0
# Outer safety-net timeout, same number gopod-song-open()'s own `timeout 70`
# uses - the binary's internal deadline already widens to 20+max-wait
# (default 60s) once --wait-signal is given; this is the process-level
# backstop on top of that, not a tighter limit.
SLEEP_WAKE_TIMEOUT_SECONDS = 70


# STANDARDIZED READINESS SIGNAL, added 2026-08-10 per operator direction:
# "standardize the signal, not the mechanism." Two genuinely different
# control channels in this file each answer "is the robot genuinely ready,
# hand off or gate?" - Wire-Pod REST's golden-flag release/settle/reassume
# pulse (cmd_brobots_wake) and direct-SDK's continuous-connection
# signal-file gate (_finish_sleep_wake). They stay different mechanisms on
# purpose (_BROBOTS_WAKE_CHAIN_ELIGIBLE's own comment already explains why
# mixing them is a real hazard, not just an inconsistency) - this only
# gives both the same SHAPE of answer, so a caller (today: the printed
# PHCAL_READY log line; later: any future code that wants the dict
# directly) reads one consistent signal regardless of which channel
# produced it.
def _readiness_signal(ready, reason, channel, detail=None):
    """`ready`/`ok` carry the same boolean (kept both so every existing
    `if not X["ok"]` call site in this file keeps working unchanged -
    standardizing the signal must not silently change any existing
    control-flow check). `reason` is always a short, human-readable string,
    success or failure. `channel` names which mechanism produced this -
    "rest_golden_flag" or "direct_sdk_continuous". `detail` carries
    whatever mechanism-specific sub-results already existed, untouched."""
    return {"ready": ready, "ok": ready, "reason": reason, "channel": channel, "detail": detail or {}}


def cmd_sleep_wake_set_time(mod, clock, which, wait_seconds):
    """Combined sleep->wake bench test, timed-release mode. One continuous
    BehaviorControl connection (direct_sdk_robot_sleep_001's own
    --wait-signal mode) holds both chosen robots asleep, then release fires
    in that SAME connection the instant the signal file appears - no
    release-then-regrant gap, the exact fix gopod-song-open() already proved
    live. Release trigger here: a plain background wait of `wait_seconds`."""
    specs = _sleep_wake_specs(mod, which)
    signal_file = tempfile.mktemp(prefix="phcal_sleep_wake_signal_")
    print(
        f"{clock.prefix()}PHCAL_SLEEP_WAKE which={which} mode=set_time "
        f"wait_seconds={wait_seconds} specs={' '.join(specs)}"
    )
    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    proc = subprocess.Popen(
        [str(SLEEP_BIN_PATH), "sleep", "--wait-signal", signal_file, *specs],
        stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, env=env,
    )
    print(f"{clock.prefix()}PHCAL_SLEEP_WAKE holding {wait_seconds}s - both robots should stay asleep on one continuous connection")
    remaining = wait_seconds
    while remaining > 0:
        step = min(1.0, remaining)
        time.sleep(step)
        remaining = round(remaining - step, 3)
        if remaining > 0:
            print(f"{clock.prefix()}PHCAL_SLEEP_WAKE {remaining:.1f}s remaining")
    print(f"{clock.prefix()}PHCAL_SLEEP_WAKE releasing now (signal file touched -> GoToSleepOff fires in the same connection)")
    Path(signal_file).touch()
    signal = _finish_sleep_wake(proc, signal_file, clock, "set_time")
    return 0 if signal["ok"] else 1


def cmd_sleep_wake_on_process(mod, clock, live, which):
    """Combined sleep->wake bench test, process-signaled release mode - same
    continuous-connection mechanism as cmd_sleep_wake_set_time, but the
    release trigger is a real completed process instead of a timer:
    run_restart_wirepod_preflight() (phcal's own wpr-equivalent check), the
    same real fill-work gopod-song-open() itself runs during its own sleep
    window - not a fake stand-in. Honest caveat, same one that file's own
    build history already names: if Wire-Pod is already healthy this check
    can complete in milliseconds, so the sleep window may be very short -
    not lengthened artificially here."""
    specs = _sleep_wake_specs(mod, which)
    signal_file = tempfile.mktemp(prefix="phcal_sleep_wake_signal_")
    print(f"{clock.prefix()}PHCAL_SLEEP_WAKE which={which} mode=on_process specs={' '.join(specs)}")
    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    proc = subprocess.Popen(
        [str(SLEEP_BIN_PATH), "sleep", "--wait-signal", signal_file, *specs],
        stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, env=env,
    )
    print(f"{clock.prefix()}PHCAL_SLEEP_WAKE both robots asleep - running a real restart_wirepod_preflight check as the fill-work (may be near-instant if Wire-Pod is already healthy)")
    run_restart_wirepod_preflight(mod, clock, live)
    print(f"{clock.prefix()}PHCAL_SLEEP_WAKE fill-work done, releasing now (signal file touched -> GoToSleepOff fires in the same connection)")
    Path(signal_file).touch()
    signal = _finish_sleep_wake(proc, signal_file, clock, "on_process")
    return 0 if signal["ok"] else 1


def _finish_sleep_wake(proc, signal_file, clock, mode):
    """Direct-SDK channel's own gate-if-not-ready endpoint - builds and
    returns the standardized readiness signal (_readiness_signal above),
    same shape cmd_brobots_wake's REST-channel gate returns."""
    try:
        stdout, _ = proc.communicate(timeout=SLEEP_WAKE_TIMEOUT_SECONDS)
    except subprocess.TimeoutExpired:
        proc.kill()
        reason = f"direct-SDK sleep/wake process timed out after {SLEEP_WAKE_TIMEOUT_SECONDS}s"
        print(f"{clock.prefix()}PHCAL_SLEEP_WAKE_TIMED_OUT")
        print(f"{clock.prefix()}PHCAL_READY ready=False reason={reason!r} channel=direct_sdk_continuous")
        Path(signal_file).unlink(missing_ok=True)
        return _readiness_signal(False, reason, "direct_sdk_continuous", detail={"mode": mode, "timed_out": True})
    Path(signal_file).unlink(missing_ok=True)
    combined = "\n".join(line for line in stdout.splitlines() if "guid" not in line.lower())
    ok = proc.returncode == 0
    reason = (
        f"continuous connection held sleep->wake cleanly (mode={mode})" if ok
        else f"direct-SDK sleep/wake binary exited {proc.returncode} (mode={mode})"
    )
    print(f"{clock.prefix()}{combined}")
    print(f"{clock.prefix()}PHCAL_SLEEP_WAKE_COMPLETE ok={ok} mode={mode}")
    print(f"{clock.prefix()}PHCAL_READY ready={ok} reason={reason!r} channel=direct_sdk_continuous")
    return _readiness_signal(ok, reason, "direct_sdk_continuous", detail={"mode": mode, "returncode": proc.returncode})


def _parse_sleep_hold_flag(raw):
    val = _parse_float_flag("hold", raw)
    if val < SLEEP_MIN_HOLD_SECONDS:
        raise ValueError(
            f"hold must be >= {SLEEP_MIN_HOLD_SECONDS}s - a shorter value has already been "
            "live-observed to release a slow-firing robot before it ever settles asleep "
            "(see SLEEP_MIN_HOLD_SECONDS's own comment above for the exact math)"
        )
    return val

# 2026-07-22 rattle-audibility fix: phcal's own rattle call was reusing the
# shared DIRECT_SDK_RELEASE_SETTLE_SECONDS (1.0) above for its
# release->settle->direct-connect margin. That 1.0s was only ever live-proven
# sufficient for a quick one-shot SayText call (the "Brobots ready!"
# together-step's own use of the shared constant) - never proven sufficient
# for rattle's own longer-held audio-streaming connection
# (BINGO_RATTLE_ADDED_001.md section 6 names this gap honestly at build time).
# This exact intermittent pattern (status=OK, not heard) has now happened
# twice: once in a full bingo run (round_1_rattle), once in phcal's isolated
# rattle test tonight. Does NOT touch the shared constant above (still correct
# for SayText/other uses).
#
# Value history, live-tested, one variable at a time:
#   1.0 (shared constant, reused) - CONFIRMED intermittently insufficient,
#       status=OK but not heard, at least twice.
#   2.0 (first separate rattle constant) - a full bingo run heard 3 of 4
#       rattles; opening_rattle still missed - improved, not fully fixed.
#   3.0 - CONFIRMED WORKING, live, by the operator, TWICE: once in a normal
#       run, once as a scratch test with Brobot 1 deliberately put to sleep
#       first (a harder condition than normal) - operator's own words both
#       times: "i heard the rattle" / "i heard it". Made permanent here.
# Do not lower below 3.0 without a fresh live re-confirmation.
PHCAL_RATTLE_SETTLE_SECONDS = 3.0

# 2026-07-22 bug fix: cmd_arm/cmd_nod's per-cycle loops had NO pause between
# one cycle's final release and the next cycle's assume - a live phcal run
# (nod, count=3) logged 3 genuinely complete down+up pairs but the operator
# only visually saw 2 distinct nods land: cycle N's return-to-neutral "up"
# and cycle N+1's "down" started back to back with no settle beat, so two
# adjacent cycles blended into one continuous motion. Not a loop/off-by-one
# bug - the loop already fires exactly `count`/`cycles` full pairs (verified:
# no early break, no skipped final iteration, both directions run every
# time). Fix is a real settle pause between cycles, not a count-logic
# change. Value sourced from this codebase's own existing settle-pause
# convention, not invented: DIRECT_SDK_RELEASE_SETTLE_SECONDS and
# BROBOTS_READY_TOGETHER_HANDBACK_SETTLE_SECONDS (run_songs_runner_001.py)
# both use 1.0s for "let the motion actually stop" pauses. Applied between
# cycles only (never after the last one - no dead air after the final
# motion); hold_seconds/speed untouched, nod's speed floor untouched.
PHCAL_CYCLE_SETTLE_SECONDS = 1.0

# 2026-07-22: pre-flight connection check, added after a live phcal run (arm,
# robot 1) logged ok=True on every HTTP call (assume_behavior_control/
# move_lift/release all status 200) but the robot did not physically move - a
# live curl of assume_behavior_control moments later succeeded fine, so
# Wire-Pod<->robot connectivity was not permanently broken. Working theory:
# phcal always fires cold/isolated with zero warm-up, unlike a full song run.
# Ported from run_songs_runner_001.py's own run_wake_both (conn_test,
# then a settle sleep) - the same pattern BINGO_PLAYHEAD_WAKE reuses whenever
# a slice starts mid-song instead of from that song's own opening, because the
# robots may not be woken/paired for it yet. phcal is ALWAYS that same kind of
# cold, isolated start (never a full song opening with its own wake_both step
# already run), so this fires on every call here, not just some. Single-serial
# only, unlike bingo's two-robot run_wake_both - phcal is one robot per call.
# Not run_connection_pass's stronger say_text-smoke variant: phcal has no
# say_text-equivalent action of its own to smoke-test with.
PHCAL_PREFLIGHT_SETTLE_SECONDS = 1.5  # run_wake_both's own fallback value
# (wake_step.get("settle_seconds", 1.5)) when it has no song-scored wake_both
# step to read a tuned value from - phcal has no song/knobs.json of its own
# either, so this fallback IS the number to reuse, not a new one.

# GOLDEN ALIAS NOTE - 2026-07-24 cold-first-cycle fix, live-reported by the
# operator: cycles=2 produced only ONE visible arm cue, cycles=1 produced
# NONE - not a loop/count bug (verified the loop already fires exactly N
# full up/down or down/up pairs, same shape as the 2026-07-22 intercycle-
# blend fix above). Root cause: run_assume_control() returned, and the very
# first move_lift/move_head call fired ~2ms later (from a live log:
# assume at +0.000s, first move at +0.002s) - the robot was not physically
# ready to execute yet, so cycle 1's motion was silently dropped while
# cycle 2+ landed because the robot was warm by then. Same class of bug as
# the already-fixed "cold first press" finding in test-silent-angry-say
# (HTTP success but no playback on the first live action after a wake,
# second press played) - that fix inserted a settle pause after the wake
# step, before the first live action. This is that same fix, applied here:
# a settle after assume_behavior_control, before the first move. Value
# reused, not invented - same PHCAL_PREFLIGHT_SETTLE_SECONDS constant
# already used for the analogous conn_test-then-settle wake pattern in this
# same file.
PHCAL_ASSUME_SETTLE_SECONDS = PHCAL_PREFLIGHT_SETTLE_SECONDS

# Same field-proven recovery text as run_songs_runner_001.py's own
# WIREPOD_PAIRING_RECOVERY_MESSAGE (robot_control_song_001/story.md's own
# Troubleshooting section, confirmed live 2026-07-15), adapted singular
# (phcal is one robot, not bingo's two) - reproduced here rather than
# imported, since phcal loads run_section1_full_live_001.py as `mod`, not the
# bingo runner module that constant actually lives in.
PHCAL_PREFLIGHT_RECOVERY_MESSAGE = (
    "sequence error - Wire-Pod could not reach the robot. Field-proven recovery "
    "order (robot_control_song_001/story.md's own Troubleshooting section, confirmed "
    "live - a partial version does not clear it): 1) power-cycle the robot, "
    "2) re-pair it with Wire-Pod, 3) restart Wire-Pod (wpr). Stopping here, not "
    "guessing a different fix."
)

# 2026-07-22: second, stronger preflight, added after conn_test's own preflight
# above still wasn't enough - a live phcal run logged ok=True on EVERY HTTP call
# (assume_behavior_control/move_lift/move_head/stop/release all status 200,
# conn_test preflight also OK) but the robot never physically moved, on both
# Brobot 1 and Brobot 2, across multiple attempts, robots confirmed powered on
# and unblocked. conn_test only proves Wire-Pod's cached gRPC channel answers a
# ping - it says nothing about whether the Wire-Pod service process itself is
# in a healthy, freshly-serving state. A stale/wedged Wire-Pod process can
# still answer every one of those calls with an HTTP 200 "success" string
# without ever forwarding real gRPC to the robot hardware. Found the real
# mechanism in the bingo song's own golden pathway rather than guessing again:
# run_section1_full_live_001.py's restart_wirepod_preflight(live) - the actual
# `wpr`-equivalent (sudo systemctl restart wire-pod, poll systemctl is-active,
# poll a bare HTTP GET, settle). Called here call-for-call via `mod`, not
# reimplemented - phcal already loads this exact module as `mod`.
#
# That function only fires when its OWN GOPOD_RESTART_WIREPOD_BEFORE_RUN=1 env
# var is set (else it's a no-op even when live=True) - it defaults off there
# because a full interview run doesn't want a service restart on every launch.
# phcal is different: it's a bench-tuning tool where reliability matters more
# than speed, and the operator does not want to have to remember a separate
# export just to get the real fix. So phcal sets
# os.environ.setdefault("GOPOD_RESTART_WIREPOD_BEFORE_RUN", "1") itself, once,
# before loading `mod` (module-level constants are read at import time) - the
# same "just works when phcal fires live" way phcal already reads its own
# GOPOD_ALLOW_LIVE_ROBOT_SPEECH live gate. setdefault, not a hard overwrite:
# an operator who already exported GOPOD_RESTART_WIREPOD_BEFORE_RUN=0 in that
# shell for their own reason is still honored, not silently overridden. This
# is scoped to phcal's own process only (os.environ here, no
# `export`/`.bashrc` touch) - not a global change to the operator's shell.
#
# Runs ONCE per phcal invocation, before conn_test's own preflight (the
# service needs to be up before a conn_test to a specific robot means
# anything) - never per-serial, never per-cycle, since it restarts the whole
# Wire-Pod service, not something scoped to one robot.
PHCAL_RESTART_WIREPOD_ENV = "GOPOD_RESTART_WIREPOD_BEFORE_RUN"


def run_restart_wirepod_preflight(mod, clock, live):
    """Calls run_section1_full_live_001.py's own restart_wirepod_preflight(live)
    exactly as written - see the comment block above PHCAL_RESTART_WIREPOD_ENV
    for why this exists and why phcal defaults its own env var on. Wrapped in
    try/except: restart_wirepod_preflight raises RuntimeError if the service
    never comes active or HTTP never responds, which is exactly what this
    handler exists to catch, same pattern as run_preflight's (conn_test)
    existing failure handling - print, stop, do not fire any move command."""
    if not live:
        print(f"{clock.prefix()}PHCAL_PREFLIGHT DRY: would restart_wirepod_preflight (wrp-equivalent) before conn_test")
        return {"ok": True, "mode": "dry"}
    print(f"{clock.prefix()}PHCAL_PREFLIGHT restart_wirepod_preflight starting (sudo systemctl restart wire-pod, then poll for ready)")
    try:
        result = mod.restart_wirepod_preflight(live)
    except RuntimeError as exc:
        print(f"{clock.prefix()}PHCAL_PREFLIGHT_FAILED restart_wirepod_preflight - {exc}")
        return {"ok": False, "error": str(exc)}
    if result.get("enabled"):
        print(
            f"{clock.prefix()}PHCAL_PREFLIGHT restart_wirepod_preflight enabled=True "
            f"service_state={result.get('service_state')} http_ready={result.get('http_ready')}"
        )
    else:
        print(
            f"{clock.prefix()}PHCAL_PREFLIGHT restart_wirepod_preflight enabled=False - "
            f"{PHCAL_RESTART_WIREPOD_ENV} was not '1' at the time `mod` loaded (phcal sets "
            "this itself by default via setdefault; only happens if something already set "
            "it to a different value first) - skipped, proceeding to conn_test preflight "
            "without a fresh service restart"
        )
    return {"ok": True, "result": result}


# phcal's OWN memory file - outside the tracked GOPOD repo, sibling to this
# script. NOT a song's knobs.json/story.md; this only remembers the
# operator's own bench settings for phcal's next run (rung 3's write-back).
LAST_PATH = Path(__file__).resolve().parent / "phcal_last.json"


def _load_module(path, name):
    spec = importlib.util.spec_from_file_location(name, str(path))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


class _Clock:
    """Same "[HH:MM:SS.mmm +X.XXXs] " shape as Robots.timing_prefix() in
    run_section1_full_live_001.py - reproduced standalone here rather than
    constructing a full Robots() (which needs load_interview_scaffold() and
    other interview-only state this isolated tool has no use for)."""

    def __init__(self):
        self._last = None

    def prefix(self):
        now = time.time()
        clock = time.strftime("%H:%M:%S", time.localtime(now)) + f".{int(now % 1 * 1000):03d}"
        if self._last is None:
            self._last = now
            return f"[{clock} +0.000s] "
        delta = now - self._last
        self._last = now
        return f"[{clock} +{delta:.3f}s] "


# 2026-07-22 bug fix: cmd_arm/cmd_nod previously called run_move_axis_isolated
# (below) once per DIRECTION, and that function did its own full
# assume_behavior_control -> move -> hold -> stop -> release_behavior_control
# cycle every single call - so one arm/nod cycle fired 2 separate
# assume/release pairs (up, then down), cycles=3 fired 6 pairs back to back.
# That is exactly the repeated assume/release churn
# run_named_movement_sequence()'s own comment (run_robot_control_song_001.py,
# ~line 261) names as the cause of the robot's screen flashing to its home
# icon between every note, confirmed live - and matches the operator's own
# live observation of phcal test runs "looking like robot pings, not expected
# robot mechanical control": HTTP ok=True on every call, but the robot never
# held control long enough to actually execute the commanded move. Root cause
# confirmed against that same function's own reference shape (assume ONCE
# before the whole sequence, release ONCE after, in a finally - never
# per-direction/per-cycle), not a new theory. Fixed by splitting the old
# run_move_axis_isolated into two pieces: run_move_axis_no_control (move ->
# hold -> stop only, no assume/release) fires per direction/cycle exactly as
# before; cmd_arm/cmd_nod now assume ONCE before their whole cycle loop and
# release ONCE after it, in a finally, matching run_named_movement_sequence's
# own try/finally shape.
def run_move_axis_no_control(mod, clock, serial, endpoint, speed, hold_seconds, live, direction_label):
    """Move-only half of the old run_move_axis_isolated: {endpoint}
    speed=speed -> sleep hold_seconds -> {endpoint} speed=0 -> stop. Does NOT
    assume or release behavior control - the caller (cmd_arm/cmd_nod) now
    holds control across the entire cycle loop, assumed once before this is
    ever called and released once after every direction/cycle has fired, so
    the robot never drops back to idle between individual moves. Prints each
    sub-instruction as it fires, live or dry, same style as before."""
    tag = f"PHCAL direction={direction_label}"

    if not live:
        print(f"{clock.prefix()}{tag} DRY: would {endpoint} serial={serial} speed={speed}")
        print(f"{clock.prefix()}{tag} DRY: would hold {hold_seconds}s")
        print(f"{clock.prefix()}{tag} DRY: would {endpoint} serial={serial} speed=0 (stop)")
        return {"ok": True, "mode": "dry"}

    print(f"{clock.prefix()}{tag} {endpoint} serial={serial} speed={speed}")
    move = mod.wirepod_web_send_form(f"/api-sdk/{endpoint}", {"serial": serial, "speed": speed}, timeout=10)

    print(f"{clock.prefix()}{tag} hold {hold_seconds}s")
    time.sleep(hold_seconds)

    print(f"{clock.prefix()}{tag} {endpoint} serial={serial} speed=0 (stop)")
    stop = mod.wirepod_web_send_form(f"/api-sdk/{endpoint}", {"serial": serial, "speed": 0}, timeout=10)

    ok = move.get("status") == 200 and stop.get("status") == 200
    print(f"{clock.prefix()}PHCAL_DONE direction={direction_label} ok={ok}")
    return {"ok": ok, "move": move, "stop": stop}


def run_move_wheels_no_control(mod, clock, serial, lw, rw, hold_seconds, live):
    """Wheel twin of run_move_axis_no_control above - same move -> hold ->
    stop shape, but move_wheels takes TWO independent speed params (lw/rw),
    not one, so this is a new function rather than a forced reuse of the
    single-speed helper (WHEEL_NUDGE_GOLDEN_PATH_SURVEY_001.md, section A).
    Does NOT assume or release behavior control - the caller
    (cmd_move_reverse) holds control across this one call, same convention as
    every other *_no_control helper here. The stop call is unconditional -
    fires even when the move call itself reports a non-200 - a wheel move
    must never be left without a matching stop."""
    tag = "PHCAL_MOVE_REVERSE"

    if not live:
        print(f"{clock.prefix()}{tag} DRY: would move_wheels serial={serial} lw={lw} rw={rw}")
        print(f"{clock.prefix()}{tag} DRY: would hold {hold_seconds}s")
        print(f"{clock.prefix()}{tag} DRY: would move_wheels serial={serial} lw=0 rw=0 (stop)")
        return {"ok": True, "mode": "dry"}

    # PRIME REMOVED 2026-08-10 (WHEEL_NUDGE_COLD_FIRE_ROOT_CAUSE_SURVEY_001.md
    # + its live-diagnostic follow-up): the old "fire twice, back to back"
    # duplicate below was a workaround for a still-waiting-on-the-grant
    # theory that was itself only ever tested inside a flat-sleep settle
    # window. The golden-flag mechanism (run_wheel_nudge_wake_release_pulse
    # below - a real release-then-reassume pulse, not a guessed delay) proved
    # reliable with a SINGLE move call across 8/8 live runs, both robots -
    # not kept as redundant insurance, removed outright, matching the same
    # replacement already made in run_robot_control_song_001.py's own
    # run_move_reverse().
    print(f"{clock.prefix()}{tag} move_wheels serial={serial} lw={lw} rw={rw}")
    try:
        move = mod.wirepod_web_send_form(
            "/api-sdk/move_wheels", {"serial": serial, "lw": lw, "rw": rw}, timeout=10
        )
    except Exception as exc:  # noqa: BLE001 - the stop below must still fire on a move error
        print(f"{clock.prefix()}{tag} move_wheels FAILED - {exc!r}")
        move = {"status": None, "error": repr(exc)}

    print(f"{clock.prefix()}{tag} hold {hold_seconds}s")
    time.sleep(hold_seconds)

    print(f"{clock.prefix()}{tag} move_wheels serial={serial} lw=0 rw=0 (stop)")
    stop = mod.wirepod_web_send_form("/api-sdk/move_wheels", {"serial": serial, "lw": 0, "rw": 0}, timeout=10)

    ok = move.get("status") == 200 and stop.get("status") == 200
    print(f"{clock.prefix()}PHCAL_DONE direction=reverse ok={ok}")
    return {"ok": ok, "move": move, "stop": stop}


def run_wheel_nudge_wake_check(mod, clock, serial, live):
    """conn_test only - a cheap, real pre-check ahead of assuming control,
    matching gopod-opening-chord's own separate WIREPOD_CHAIN_WAKE_START
    conn_test (fired before its own speak() call). Live field-testing
    (2026-08-09, 2nd round) showed this alone does NOT fix a cold-fire
    first-run miss - it proves connectivity, not that THIS run's own
    upcoming assume has settled. See run_wheel_nudge_wake_release_pulse
    below for the piece that actually does that."""
    if not live:
        print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_CHECK DRY: would conn_test serial={serial}")
        return {"ok": True, "mode": "dry"}
    print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_CHECK conn_test serial={serial}")
    try:
        conn = mod.wirepod_web_send_form(
            "/api-sdk/conn_test", {"serial": serial}, timeout=WHEEL_NUDGE_WAKE_CHECK_TIMEOUT_SECONDS
        )
        ok = conn.get("status") == 200
    except Exception as exc:  # noqa: BLE001 - a hang/timeout here is exactly what this check exists to catch fast
        print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_CHECK_FAILED serial={serial} - {exc!r}")
        return {"ok": False, "error": repr(exc)}
    print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_CHECK_DONE serial={serial} ok={ok}")
    return {"ok": ok, "conn_test": conn}


def run_wheel_nudge_wake_release_pulse(mod, clock, serial, live):
    """REPLACED 2026-08-10 (WHEEL_NUDGE_COLD_FIRE_ROOT_CAUSE_SURVEY_001.md):
    the real, measured fix - see WHEEL_NUDGE_WAKE_RELEASE_SETTLE_SECONDS's
    own comment above for the full why. Fired INSIDE the caller's
    assume/release window, immediately after assume and before the actual
    move: release_behavior_control (the "golden flag" pulse) -> settle ->
    re-assume_behavior_control. This is what proves the assume just fired
    has actually settled - not a guessed sleep, not a say_text call that
    turned out to prove nothing."""
    if not live:
        print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_RELEASE_PULSE DRY: would release_behavior_control -> settle {WHEEL_NUDGE_WAKE_RELEASE_SETTLE_SECONDS}s -> re-assume_behavior_control serial={serial}")
        return {"ok": True, "mode": "dry"}
    print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_RELEASE_PULSE release_behavior_control serial={serial}")
    try:
        release_pulse = mod.wirepod_web_send_form(
            "/api-sdk/release_behavior_control", {"serial": serial}, timeout=10
        )
        print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_RELEASE_PULSE settle {WHEEL_NUDGE_WAKE_RELEASE_SETTLE_SECONDS}s")
        time.sleep(WHEEL_NUDGE_WAKE_RELEASE_SETTLE_SECONDS)
        print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_RELEASE_PULSE re-assume_behavior_control serial={serial}")
        reassume = mod.wirepod_web_send_form(
            "/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10
        )
        time.sleep(WHEEL_NUDGE_WAKE_POST_REASSUME_SETTLE_SECONDS)
        ok = release_pulse.get("status") == 200 and reassume.get("status") == 200
    except Exception as exc:  # noqa: BLE001 - a hang/timeout here is exactly what this check exists to catch fast
        print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_RELEASE_PULSE_FAILED serial={serial} - {exc!r}")
        return {"ok": False, "error": repr(exc)}
    print(f"{clock.prefix()}PHCAL_WHEEL_NUDGE_WAKE_RELEASE_PULSE_DONE serial={serial} ok={ok}")
    return {"ok": ok, "release_pulse": release_pulse, "reassume": reassume}


def run_assume_control(mod, clock, serial, live):
    """Fires assume_behavior_control ONCE, before cmd_arm/cmd_nod's whole
    cycle loop starts - not per direction/per cycle. See the 2026-07-22 fix
    comment above run_move_axis_no_control for why."""
    if not live:
        print(f"{clock.prefix()}PHCAL DRY: would assume_behavior_control serial={serial} priority=high")
        return {"ok": True, "mode": "dry"}
    print(f"{clock.prefix()}PHCAL assume_behavior_control serial={serial} priority=high")
    assume = mod.wirepod_web_send_form(
        "/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10
    )
    # Cold-first-cycle fix (2026-07-24) - see PHCAL_ASSUME_SETTLE_SECONDS's
    # own comment above for the live finding this addresses. Without this,
    # the first move command after assume fired ~2ms later and was silently
    # dropped.
    print(f"{clock.prefix()}PHCAL assume settle {PHCAL_ASSUME_SETTLE_SECONDS}s")
    time.sleep(PHCAL_ASSUME_SETTLE_SECONDS)
    return {"ok": assume.get("status") == 200, "assume": assume}


def run_release_control(mod, clock, serial, live):
    """Fires release_behavior_control ONCE, after cmd_arm/cmd_nod's whole
    cycle loop has finished - not per direction/per cycle. Called from a
    finally block by the caller, so a mid-sequence failure still releases
    cleanly, matching run_named_movement_sequence's own try/finally shape."""
    if not live:
        print(f"{clock.prefix()}PHCAL DRY: would release_behavior_control serial={serial}")
        return {"ok": True, "mode": "dry"}
    print(f"{clock.prefix()}PHCAL release_behavior_control serial={serial}")
    release = mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)
    return {"ok": release.get("status") == 200, "release": release}


def run_preflight(mod, clock, serial, live):
    """conn_test + settle, before any move command fires - see
    PHCAL_PREFLIGHT_SETTLE_SECONDS above for why this exists and why it's
    ported call-for-call from run_wake_both rather than invented. Wrapped in
    try/except the same way run_wake_both is: a hang/timeout/refused
    connection here is exactly what this handler exists to catch, not let
    propagate as a raw traceback."""
    if not live:
        print(f"{clock.prefix()}PHCAL_PREFLIGHT DRY: would conn_test serial={serial}")
        return {"ok": True, "mode": "dry"}
    print(f"{clock.prefix()}PHCAL_PREFLIGHT conn_test serial={serial}")
    try:
        result = mod.wirepod_web_send_form("/api-sdk/conn_test", {"serial": serial}, timeout=10)
    except Exception as exc:  # noqa: BLE001 - same catch shape as run_wake_both
        print(f"{clock.prefix()}PHCAL_PREFLIGHT_FAILED serial={serial} - {exc!r}")
        return {"ok": False, "error": repr(exc)}
    time.sleep(PHCAL_PREFLIGHT_SETTLE_SECONDS)
    ok = result.get("status") == 200
    if ok:
        print(f"{clock.prefix()}PHCAL_PREFLIGHT_OK serial={serial}")
    else:
        print(f"{clock.prefix()}PHCAL_PREFLIGHT_FAILED serial={serial} - {PHCAL_PREFLIGHT_RECOVERY_MESSAGE}")
    return {"ok": ok, "conn_test": result}


def cmd_arm(mod, clock, serial, robot, live, cycles, hold_seconds, speed, pre_assumed=False):
    print(
        f"{clock.prefix()}PHCAL_ARM robot={robot} serial={serial} cycles={cycles} "
        f"hold_seconds={hold_seconds} speed={speed}"
    )
    # 2026-07-22 fix: assume control ONCE for the whole call (every
    # direction/cycle below), release ONCE at the end in a finally - not a
    # pair per direction. See run_move_axis_no_control's own comment block
    # above for the root cause this replaces.
    # pre_assumed, 2026-08-09 (brobots_wake chain toggle): skip this
    # function's own assume when a shared brobots_wake step already holds
    # control - release still fires unconditionally below either way, so
    # whichever step ran last (brobots_wake alone, or this) is the one that
    # actually hands control back.
    if pre_assumed:
        print(f"{clock.prefix()}PHCAL_ARM pre_assumed=True - skipping own assume (brobots_wake already holds control)")
        ok = True
    else:
        assume = run_assume_control(mod, clock, serial, live)
        ok = assume["ok"]
    try:
        for i in range(cycles):
            up = run_move_axis_no_control(mod, clock, serial, "move_lift", speed, hold_seconds, live, "up")
            down = run_move_axis_no_control(mod, clock, serial, "move_lift", -speed, hold_seconds, live, "down")
            ok = ok and up["ok"] and down["ok"]
            if i < cycles - 1:
                print(f"{clock.prefix()}PHCAL_ARM settle {PHCAL_CYCLE_SETTLE_SECONDS}s before next cycle")
                if live:
                    time.sleep(PHCAL_CYCLE_SETTLE_SECONDS)
    finally:
        release = run_release_control(mod, clock, serial, live)
        ok = ok and release["ok"]
    print(f"{clock.prefix()}PHCAL_ARM_COMPLETE ok={ok}")
    return 0 if ok else 1


def cmd_nod(mod, clock, serial, robot, live, count, hold_seconds, speed, pre_assumed=False):
    print(
        f"{clock.prefix()}PHCAL_NOD robot={robot} serial={serial} count={count} "
        f"hold_seconds={hold_seconds} speed={speed}"
    )
    # 2026-07-22 fix: assume control ONCE for the whole call (every
    # direction/cycle below), release ONCE at the end in a finally - not a
    # pair per direction. See run_move_axis_no_control's own comment block
    # above for the root cause this replaces.
    # pre_assumed, 2026-08-09 (brobots_wake chain toggle): see cmd_arm's own
    # comment above - same convention.
    if pre_assumed:
        print(f"{clock.prefix()}PHCAL_NOD pre_assumed=True - skipping own assume (brobots_wake already holds control)")
        ok = True
    else:
        assume = run_assume_control(mod, clock, serial, live)
        ok = assume["ok"]
    try:
        for i in range(count):
            print(f"{clock.prefix()}PHCAL_NOD cycle={i + 1}/{count}")
            # Source order (run_nod, run_songs_runner_001.py): down
            # first, then up - reproduced exactly, not reordered to match
            # this spec's own descriptive "up -> down" prose.
            down = run_move_axis_no_control(mod, clock, serial, "move_head", -speed, hold_seconds, live, "down")
            up = run_move_axis_no_control(mod, clock, serial, "move_head", speed, hold_seconds, live, "up")
            ok = ok and down["ok"] and up["ok"]
            if i < count - 1:
                print(f"{clock.prefix()}PHCAL_NOD settle {PHCAL_CYCLE_SETTLE_SECONDS}s before next cycle")
                if live:
                    time.sleep(PHCAL_CYCLE_SETTLE_SECONDS)
    finally:
        release = run_release_control(mod, clock, serial, live)
        ok = ok and release["ok"]
    print(f"{clock.prefix()}PHCAL_NOD_COMPLETE ok={ok}")
    return 0 if ok else 1


# 2026-07-22, Part C: phcal's own 1-5 bench scale, linearly mapped onto the
# real Vector SDK ExternalAudioStreamPrepare.AudioVolume range (1-100, that
# field's own documented max) - operator-specified mapping, not a guess.
_VOLUME_UI_TO_REAL = {1: 20, 2: 40, 3: 60, 4: 80, 5: 100}


def _map_volume_ui_to_real(volume_ui):
    return _VOLUME_UI_TO_REAL[volume_ui]


def cmd_rattle(mod, clock, serial, robot, live, volume_ui):
    """Fires the Bingo sidecar's rattle sound on one robot via the standalone
    direct-SDK binary (direct_sdk_bingo_rattle_001), ported call-for-call
    from run_songs_runner_001.py's own run_rattle(): Wire-Pod releases
    the robot -> settle past Wire-Pod's own release-polling window
    (DIRECT_SDK_RELEASE_SETTLE_SECONDS) -> the binary connects independently,
    requests BehaviorControl, plays the rattle WAV, releases, disconnects.
    Confirmed against the Go source: no separate hold/settle step exists
    anywhere in that flow - release fires immediately after playback
    finishes - so there is nothing to add here for "hold." volume_ui is
    phcal's own 1-5 scale; mapped to the real 1-100 AudioVolume range before
    being passed as the binary's optional 3rd argument."""
    volume = _map_volume_ui_to_real(volume_ui)
    print(
        f"{clock.prefix()}PHCAL_RATTLE robot={robot} serial={serial} "
        f"volume_ui={volume_ui} volume={volume}"
    )
    release = run_release_control(mod, clock, serial, live)
    ok = release["ok"]

    if not live:
        print(
            f"{clock.prefix()}PHCAL_RATTLE DRY: would settle "
            f"{PHCAL_RATTLE_SETTLE_SECONDS}s then fire rattle binary "
            f"serial={serial} wav={RATTLE_WAV_PATH} volume={volume}"
        )
        print(f"{clock.prefix()}PHCAL_RATTLE_COMPLETE ok={ok} mode=dry")
        return 0 if ok else 1

    print(
        f"{clock.prefix()}PHCAL_RATTLE settle {PHCAL_RATTLE_SETTLE_SECONDS}s "
        "(rattle's own audio-streaming settle margin, wider than the shared "
        "DIRECT_SDK_RELEASE_SETTLE_SECONDS SayText uses - see the constant's "
        "own comment above)"
    )
    time.sleep(PHCAL_RATTLE_SETTLE_SECONDS)

    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    try:
        proc = subprocess.run(
            [str(RATTLE_BIN_PATH), serial, str(RATTLE_WAV_PATH), str(volume)],
            capture_output=True, text=True, timeout=30, env=env,
        )
    except subprocess.TimeoutExpired:
        print(f"{clock.prefix()}PHCAL_RATTLE_DIRECT serial={serial} TIMED OUT (30s)")
        return 1

    # Same token-in-log discipline as run_rattle() itself: vector.NewWP()'s
    # own log.Println prints the robot's auth GUID in plaintext to stderr on
    # every call - filtered out before it's ever printed here.
    combined = "\n".join(
        line for line in (proc.stdout + proc.stderr).splitlines() if "guid" not in line.lower()
    )
    fired_ok = proc.returncode == 0 and "status=OK" in combined
    print(f"{clock.prefix()}{combined}")
    ok = ok and fired_ok
    print(f"{clock.prefix()}PHCAL_RATTLE_COMPLETE ok={ok}")
    return 0 if ok else 1


def cmd_danger(mod, clock, serial, robot, live, volume_ui):
    """Fires GOPOD's own playSound sound (danger-will-robinson.wav) on one
    robot via the SAME standalone direct-SDK binary cmd_rattle uses above
    (RATTLE_BIN_PATH - never rattle-specific, just always called with
    rattle's own WAV path until now) - identical mechanism, different WAV
    path. Added 2026-08-12 so a song can fire this sound directly rather
    than only through the LLM-gated {{playSound||...}} chat path (kgsim_cmds.go's
    DoPlaySound) - see DOPLAYSOUND_REAL_IMPLEMENTATION_AND_RATTLE_FIT_001.md.
    Same settle margin as rattle (PHCAL_RATTLE_SETTLE_SECONDS - the audio-
    streaming settle mechanism isn't asset-specific, just reused by name).
    volume_ui is phcal's own 1-5 scale, same mapping as rattle."""
    volume = _map_volume_ui_to_real(volume_ui)
    print(
        f"{clock.prefix()}PHCAL_DANGER robot={robot} serial={serial} "
        f"volume_ui={volume_ui} volume={volume}"
    )
    release = run_release_control(mod, clock, serial, live)
    ok = release["ok"]

    if not live:
        print(
            f"{clock.prefix()}PHCAL_DANGER DRY: would settle "
            f"{PHCAL_RATTLE_SETTLE_SECONDS}s then fire danger binary "
            f"serial={serial} wav={DANGER_WAV_PATH} volume={volume}"
        )
        print(f"{clock.prefix()}PHCAL_DANGER_COMPLETE ok={ok} mode=dry")
        return 0 if ok else 1

    print(
        f"{clock.prefix()}PHCAL_DANGER settle {PHCAL_RATTLE_SETTLE_SECONDS}s "
        "(same audio-streaming settle margin rattle uses)"
    )
    time.sleep(PHCAL_RATTLE_SETTLE_SECONDS)

    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    try:
        proc = subprocess.run(
            [str(RATTLE_BIN_PATH), serial, str(DANGER_WAV_PATH), str(volume)],
            capture_output=True, text=True, timeout=30, env=env,
        )
    except subprocess.TimeoutExpired:
        print(f"{clock.prefix()}PHCAL_DANGER_DIRECT serial={serial} TIMED OUT (30s)")
        return 1
    except FileNotFoundError:
        # 2026-08-15: the real cause of "picking danger does nothing" -
        # RATTLE_BIN_PATH is a pre-built binary, not built by anything at
        # import/dispatch time; if it's missing, subprocess.run raises this
        # (uncaught before this guard, a raw traceback, not a clean
        # message). Build it once from its own .go source:
        # `cd gopod_probes/tools && go build -o direct_sdk_bingo_rattle_001
        # direct_sdk_bingo_rattle_001.go` - danger and rattle share this one
        # binary, so building it once fixes both, matching the description
        # already on RATTLE_BIN_PATH's own comment ("never rattle-specific").
        print(
            f"{clock.prefix()}PHCAL_DANGER_BLOCKED binary not found at {RATTLE_BIN_PATH} - "
            "build it first: cd $(dirname " + str(RATTLE_BIN_PATH) + ") && go build -o "
            f"{RATTLE_BIN_PATH.name} {RATTLE_BIN_PATH.name}.go"
        )
        return 1

    # Same token-in-log discipline as cmd_rattle above.
    combined = "\n".join(
        line for line in (proc.stdout + proc.stderr).splitlines() if "guid" not in line.lower()
    )
    fired_ok = proc.returncode == 0 and "status=OK" in combined
    print(f"{clock.prefix()}{combined}")
    ok = ok and fired_ok
    print(f"{clock.prefix()}PHCAL_DANGER_COMPLETE ok={ok}")
    return 0 if ok else 1


def cmd_cube(mod, clock, serial, robot, live):
    """Fires the cube connect->all-corners-red->hold->all-corners-green->
    release blip via the standalone direct-SDK binary
    (direct_sdk_cube_blip_001) - connects independently, requests
    BehaviorControl, connects the cube (ConnectCube), sets all four corner
    LEDs red then green (2s hold between, baked into the binary's own
    blipHoldDuration constant), releases, disconnects. No volume/settle
    concept (unlike rattle/danger) - nothing audio-streamed here, so no
    audio-streaming settle margin applies. Net-new 2026-08-15, the first
    GOPOD code to touch the cube instrument - see
    CUBE_DOOR_SURVEY_001.md/CUBE_BLIP_TOOL_BUILT_001.md. Cube keeper is Brobot 2
    (robot=2, ESN 0dd1d8bf) - this primitive doesn't hardcode that (same as
    rattle/danger not hardcoding a robot), it just fails cleanly if fired at
    a robot with no cube paired (ConnectCube's own Success=false path).
    WIREPOD_HOME must be set so vector.NewWP() resolves
    chipper/jdocs/botSdkInfo.json as an absolute path rather than relative
    to whatever cwd this process happens to run from - the exact bug a raw
    unwired run hit (CUBE_BLIP_ALIAS_PHCAL_WIRED_001.md); same fix every
    other direct-SDK caller in this family already applies (rattle/danger's
    own env=dict(...) above, robot-sleep/-wake/-info's own inline
    WIREPOD_HOME= prefix in brobots.sh)."""
    print(f"{clock.prefix()}PHCAL_CUBE robot={robot} serial={serial}")
    release = run_release_control(mod, clock, serial, live)
    ok = release["ok"]

    if not live:
        print(f"{clock.prefix()}PHCAL_CUBE DRY: would fire cube blip binary serial={serial}")
        print(f"{clock.prefix()}PHCAL_CUBE_COMPLETE ok={ok} mode=dry")
        return 0 if ok else 1

    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    try:
        proc = subprocess.run(
            [str(CUBE_BLIP_BIN_PATH), serial],
            capture_output=True, text=True, timeout=30, env=env,
        )
    except subprocess.TimeoutExpired:
        print(f"{clock.prefix()}PHCAL_CUBE_DIRECT serial={serial} TIMED OUT (30s)")
        return 1

    # Same token-in-log discipline as cmd_rattle/cmd_danger above.
    combined = "\n".join(
        line for line in (proc.stdout + proc.stderr).splitlines() if "guid" not in line.lower()
    )
    fired_ok = proc.returncode == 0 and "status=OK" in combined
    print(f"{clock.prefix()}{combined}")
    ok = ok and fired_ok
    print(f"{clock.prefix()}PHCAL_CUBE_COMPLETE ok={ok}")
    return 0 if ok else 1


def _sleep_wake_specs(mod, which):
    """which is '1'/'2'/'both' - same which->serial mapping
    brobots.sh's own _robot_sleep_specs() shell function already applies
    (that one also accepts '0' as a synonym for 'both'; phcal's own guided
    prompt only ever offers '1'/'2'/'both', so '0' is not reproduced here -
    nothing is lost, just one spelling not exposed at this entry point)."""
    specs = []
    if which in ("1", "both"):
        specs.append(f"Brobot_1:{BROBOT_1_SERIAL}")
    if which in ("2", "both"):
        specs.append(f"Brobot_2:{BROBOT_2_SERIAL}")
    return specs


def cmd_brobots_ready(mod, clock, live, phrase):
    """Fires the synchronized dual-robot phrase via the same
    direct_sdk_brobots_ready_001 binary run_golden_song_001.py's own
    run_brobots_ready_together() and core.sh's
    _gopod_chord_direct_together_job() already call - always both robots,
    fixed pair, no per-call robot choice (matches the catalog's own
    ROBOT-ASSIGNABLE=NO finding for this mechanism). No dry mode, same
    reasoning as cmd_sleep/cmd_wake."""
    print(f"{clock.prefix()}PHCAL_BROBOTS_READY phrase={phrase!r}")
    if not live:
        print(f"{clock.prefix()}PHCAL_NO_DRY_MODE brobots_ready has no simulate path (no --dry exists in the binary) - re-run with GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 to fire it")
        return 1
    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    try:
        proc = subprocess.run(
            [str(TOGETHER_BIN_PATH), phrase, f"Brobot 1:{BROBOT_1_SERIAL}", f"Brobot 2:{BROBOT_2_SERIAL}"],
            capture_output=True, text=True, timeout=30, env=env,
        )
    except subprocess.TimeoutExpired:
        print(f"{clock.prefix()}PHCAL_BROBOTS_READY_DIRECT TIMED OUT")
        return 1
    combined = "\n".join(
        line for line in (proc.stdout + proc.stderr).splitlines() if "guid" not in line.lower()
    )
    ok = proc.returncode == 0 and combined.count("status=OK") == 2
    print(f"{clock.prefix()}{combined}")
    print(f"{clock.prefix()}PHCAL_BROBOTS_READY_COMPLETE ok={ok}")
    return 0 if ok else 1


def cmd_brobots_ready_single(clock, which, label, live, phrase):
    """2026-08-18, PHCAL_DETECT_FIRST_001.md: single-mode degrade of
    cmd_brobots_ready's own phrase. The two-robot direct-SDK binary above
    (direct_sdk_brobots_ready_001.go) structurally CANNOT run with one
    robot without a Go rebuild - its own main() hard-requires exactly two
    <label>:<serial> specs (`if len(os.Args) < 4`, `specs :=
    os.Args[2:4]`), confirmed by reading the source directly, not guessed -
    out of scope for this pass. Reuses run_robot_control_song_001.py's own
    run_single_note()'s new "say_phrase" note instead (added this same
    pass) - the exact same say_line()/Wire-Pod-REST-assume-release shape
    the "weather" primitive's own single-robot speak already uses, not a
    new speech channel invented here. The two-robot version already speaks
    the SAME single phrase on both robots at once (not two different
    half-lines), so one robot speaking that same phrase alone is a
    complete, faithful single-robot version - not a truncated half of a
    duet."""
    print(f"{clock.prefix()}PHCAL_BROBOTS_READY_SINGLE robot={which} ({label}) phrase={phrase!r}")
    control_mod = _load_module(CONTROL_SONG_RUNNER_PATH, "run_robot_control_song_001")
    result = control_mod.run_single_note("say_phrase", live, which, phrase=phrase)
    ok = bool(result.get("ok"))
    print(f"{clock.prefix()}PHCAL_BROBOTS_READY_SINGLE_COMPLETE ok={ok}")
    return 0 if ok else 1


def cmd_robot_info(mod, clock, which):
    """Read-only VersionState/ProtocolVersion/BatteryState snapshot, direct-
    SDK, no BehaviorControl needed at all - the diagnostic golden note built
    2026-08-10 for the Brobot 2 (Pip) intermittent-unresponsiveness
    investigation (BROBOT_2_INSTABILITY_EXTERNAL_AI_BRIEF_001.md /
    WHEEL_NUDGE_COLD_FIRE_ROOT_CAUSE_SURVEY_001.md's own follow-up). Same
    binary brobots.sh's own robot-info alias already calls
    (direct_sdk_robot_info_001 - ROBOT_SLEEP_DIRECT_SDK_BUILT_001.md's
    sibling tool, reused verbatim, not a second implementation). which is
    "1"/"2"/"both", same convention/resolver (_sleep_wake_specs) sleep_wake
    already uses. No dry mode - the binary itself has none (same shape as
    cmd_brobots_ready/sleep_wake above); this call is harmless regardless
    (never assumes BehaviorControl, never moves anything, read-only), so
    there is nothing a dry preview would meaningfully simulate. Live-gated
    the same as every other primitive that actually opens a connection to
    the real robot. (2026-08-20, PHCAL_NAV_FIXES_001.md: corrected - this
    used to claim "only weather (zero robot contact at all) skips this,"
    but weather DOES make robot contact, via run_single_note()'s own
    internal connect, not the explicit restart_preflight()/run_preflight()
    pair this function and most others call directly - see weather's own
    dispatch branch and _submenu_control_note() for the real
    classification.)"""
    specs = _sleep_wake_specs(mod, which)
    print(f"{clock.prefix()}PHCAL_ROBOT_INFO which={which} specs={' '.join(specs)}")
    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    try:
        proc = subprocess.run(
            [str(ROBOT_INFO_BIN_PATH), *specs],
            capture_output=True, text=True, timeout=ROBOT_INFO_TIMEOUT_SECONDS, env=env,
        )
    except subprocess.TimeoutExpired:
        print(f"{clock.prefix()}PHCAL_ROBOT_INFO_TIMED_OUT after {ROBOT_INFO_TIMEOUT_SECONDS}s")
        return 1
    combined = "\n".join(
        line for line in (proc.stdout + proc.stderr).splitlines() if "guid" not in line.lower()
    )
    ok = proc.returncode == 0
    print(f"{clock.prefix()}{combined}")
    print(f"{clock.prefix()}PHCAL_ROBOT_INFO_COMPLETE ok={ok}")
    return 0 if ok else 1


# BATTERY GATE, added 2026-08-10 per operator direction: a voltage check
# fired before every motor call (arm/nod/animation/move_reverse - anything
# that fires MoveLift/MoveHead/DriveWheels), not just an on-request
# diagnostic. Built directly from this session's own live data, not a
# vendor spec - none exists anywhere in this repo (robot_info's own
# BatteryState read confirmed no manufacturer-documented minimum ships in
# the SDK/proto). Observed-bad: Brobot 2's two real crash/reboot incidents
# this session both read ~3.61-3.62V (BROBOT_2_INSTABILITY_EXTERNAL_AI_
# BRIEF_001.md, then a same-day repeat). Observed-good: every healthy
# reading this session, both robots, sat at 3.99-4.08V. 3.7V splits that
# gap with real margin on both sides - adjust here if live experience says
# otherwise.
BATTERY_MIN_VOLTS = 3.7


def run_battery_check(mod, clock, serial, live):
    """Reuses the exact same direct_sdk_robot_info_001 binary robot_info's
    own menu item calls (cmd_robot_info above) - one serial at a time, no
    BehaviorControl, read-only. BLOCKS (returns ok=False, does not just
    warn) when battery_volts reads below BATTERY_MIN_VOLTS or can't be
    read at all - a failed read gets treated the same as a low reading,
    not waved through, since "couldn't confirm the robot is safe to drive"
    is exactly the same risk as "confirmed it isn't." Per operator
    direction 2026-08-10: block, don't just warn."""
    if not live:
        print(f"{clock.prefix()}PHCAL_BATTERY_CHECK DRY: would read battery_volts for serial={serial}")
        return {"ok": True, "mode": "dry"}
    label = "Brobot_1" if serial == BROBOT_1_SERIAL else "Brobot_2"
    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    try:
        proc = subprocess.run(
            [str(ROBOT_INFO_BIN_PATH), f"{label}:{serial}"],
            capture_output=True, text=True, timeout=ROBOT_INFO_TIMEOUT_SECONDS, env=env,
        )
    except subprocess.TimeoutExpired:
        print(f"{clock.prefix()}PHCAL_BATTERY_CHECK_TIMED_OUT after {ROBOT_INFO_TIMEOUT_SECONDS}s serial={serial}")
        return {"ok": False, "volts": None, "error": "timeout"}
    match = re.search(r"battery_volts=([\d.]+)", proc.stdout)
    if proc.returncode != 0 or not match:
        print(f"{clock.prefix()}PHCAL_BATTERY_CHECK_FAILED serial={serial} rc={proc.returncode} - could not read battery_volts")
        return {"ok": False, "volts": None, "error": "no_reading"}
    volts = float(match.group(1))
    ok = volts >= BATTERY_MIN_VOLTS
    tag = "PHCAL_BATTERY_CHECK_OK" if ok else "PHCAL_BATTERY_CHECK_BLOCKED"
    print(f"{clock.prefix()}{tag} serial={serial} volts={volts:.3f} min={BATTERY_MIN_VOLTS}")
    return {"ok": ok, "volts": volts}


# ==========================================================================
# DETECT-FIRST, 2026-08-18 (PHCAL_DETECT_FIRST_001.md). run_guided_flow()
# calls this ONCE, before the menu ever draws - replaces "1. info" (an
# optional menu pick) with a required first step that probes which
# configured robot(s) are actually present and shapes the session
# (session_mode: "none"/"single"/"multi") to that. No new capability: the
# probe reuses cmd_robot_info's own ROBOT_INFO_BIN_PATH call verbatim
# (read-only VersionState/BatteryState snapshot, no BehaviorControl, never
# moves anything), one candidate at a time, bounded by the same
# ROBOT_INFO_TIMEOUT_SECONDS that call already uses so a dead/absent
# candidate can't hang startup. Session-scoped state (_SESSION_MODE/
# _PRESENT_ROBOTS below) is module-level, set once here and read by
# _prompt_robot() and the robot_info/brobots_announce_in_sync dispatch
# branches further down - kept module-level (not threaded as an explicit
# param through every one of _prompt_robot's 8 call sites) specifically so
# a missed call site can't silently skip mode-awareness; every reader
# checks _SESSION_MODE == "single" explicitly and no-ops otherwise, so the
# direct-flag main() path (which never calls run_guided_flow() and never
# sets these) and MULTI/NONE guided-flow sessions are both completely
# unaffected - they behave byte-for-byte as before this pass.
_SESSION_MODE = None
_PRESENT_ROBOTS = []


def _candidate_list():
    """The list detect-first probes. Reads phcal_config.json's own
    robots.candidates directly (N-capable - see the demo config's own
    comment - but this repo has only ever tested exactly 2). If that list
    is empty (a fresh clone with no config yet, or an older-shape config
    missing the key entirely) falls back to synthesizing the same two
    candidates _resolve_brobot_serials() already falls back to
    (BROBOT_1_SERIAL/BROBOT_2_SERIAL, which themselves already resolve
    through env-var-then-hardcoded) - so a fresh clone with no config still
    probes the exact same two ESNs phcal has always assumed present, just
    genuinely probed now instead of assumed. `which` here (not present in
    phcal_config.json's own candidate dicts) is this file's own "1"/"2"
    identity, derived purely from list position - index 0 is Brobot 1's
    slot, index 1 is Brobot 2's slot (same convention the demo config's own
    comment states: "First two entries map to Brobot 1 / Brobot 2 in that
    order"). A 3rd+ candidate (untested) gets which=None - this file's
    _prompt_robot()/_sleep_wake_specs() contract is binary "1"/"2"/"both"
    only, unchanged by this pass; a genuine 3rd-robot identity would need
    deeper work, out of scope here."""
    candidates = _PHCAL_CONFIG.get("robots", {}).get("candidates", [])
    if not candidates:
        candidates = [
            {"label": "Brobot 1", "esn": BROBOT_1_SERIAL},
            {"label": "Brobot 2", "esn": BROBOT_2_SERIAL},
        ]
    out = []
    for i, c in enumerate(candidates):
        which = "1" if i == 0 else "2" if i == 1 else None
        out.append({"label": c["label"], "esn": c["esn"], "which": which})
    return out


def _probe_candidate(clock, candidate):
    """One candidate -> candidate dict + present/volts. Same binary call
    cmd_robot_info/run_battery_check already make - read-only, no
    BehaviorControl. A timed-out or unparseable read is treated as NOT
    present (same "couldn't confirm == not safe" posture run_battery_check
    already applies to its own low-battery gate, reused here for presence
    too) - never an ambiguous maybe.

    2026-08-18 fix, live-tested by the operator: this used to share
    ROBOT_INFO_TIMEOUT_SECONDS (30s) with the deliberate, occasional "1.
    info" pick - fine when a human explicitly asked and knows to wait, but
    this probe now fires automatically on EVERY phcal launch, and gave zero
    output while waiting - a genuinely absent robot read as a silent freeze
    for up to 30 real seconds (confirmed live: two Ctrl-C's, both while
    still waiting on an absent second candidate). Every PRESENT robot in
    the operator's own live testing responded in well under 1 second. Now:
    (1) prints a "probing..." line immediately, before the subprocess call,
    so waiting is visible instead of silent; (2) uses its own short,
    separate timeout (PHCAL_DETECT_PROBE_TIMEOUT_SECONDS, 5s - generous vs.
    the ~0.5-0.6s real observed latency, nowhere near the deliberate-check
    ceiling) so an absent robot resolves quickly on every startup instead
    of eating up to 30s per absent candidate. ROBOT_INFO_TIMEOUT_SECONDS
    itself is untouched - "1. info"/battery-check still get their original
    30s, unchanged."""
    label, esn = candidate["label"], candidate["esn"]
    print(f"{clock.prefix()}PHCAL_DETECT_PROBE label={label} esn={esn} probing...")
    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    try:
        proc = subprocess.run(
            [str(ROBOT_INFO_BIN_PATH), f"{label}:{esn}"],
            capture_output=True, text=True, timeout=PHCAL_DETECT_PROBE_TIMEOUT_SECONDS, env=env,
        )
    except subprocess.TimeoutExpired:
        print(f"{clock.prefix()}PHCAL_DETECT_PROBE label={label} esn={esn} TIMED_OUT after {PHCAL_DETECT_PROBE_TIMEOUT_SECONDS}s - treating as not present")
        return {**candidate, "present": False, "volts": None}
    match = re.search(r"battery_volts=([\d.]+)", proc.stdout)
    if proc.returncode != 0 or not match:
        print(f"{clock.prefix()}PHCAL_DETECT_PROBE label={label} esn={esn} NOT_PRESENT rc={proc.returncode}")
        return {**candidate, "present": False, "volts": None}
    volts = float(match.group(1))
    print(f"{clock.prefix()}PHCAL_DETECT_PROBE label={label} esn={esn} PRESENT volts={volts:.3f}")
    return {**candidate, "present": True, "volts": volts}


def _detect_present_robots(clock):
    """Runs once, at the very top of run_guided_flow(), before the menu
    draws. Returns (present_robots, session_mode): present_robots is the
    subset of _candidate_list() that responded, in candidate-list order;
    session_mode is "none" (0 present), "single" (1 present), or "multi"
    (2+ present - N-capable, tested at exactly 2)."""
    candidates = _candidate_list()
    print(f"{clock.prefix()}PHCAL_DETECT_FIRST probing {len(candidates)} configured candidate(s)...")
    results = [_probe_candidate(clock, c) for c in candidates]
    present = [r for r in results if r["present"]]
    if len(present) == 0:
        mode = "none"
    elif len(present) == 1:
        mode = "single"
    else:
        mode = "multi"
    print(f"{clock.prefix()}PHCAL_DETECT_FIRST present={[p['label'] for p in present]} mode={mode}")
    return present, mode


def _confirm_multi_mode(clock, present_robots):
    """2026-08-18, operator request after live-testing multi mode: an
    explicit confirm/override step, ONLY when detect-first resolves multi -
    single and none still proceed automatically as detected, unchanged.
    Returns (present_robots, session_mode) in the exact shape
    _detect_present_robots() returns, so run_guided_flow() and every
    mode-aware reader below treat an override identically to a real
    single/none/multi detection - no separate code path needed anywhere
    else in this file.

    2026-08-21, PHCAL_MULTI_MODE_SCREEN_RESHAPE_001.md (operator's own
    exact mockup, superseding the two prior passes at this same screen -
    PHCAL_NAV_CONSOLIDATION_001.md's "d" re-key and PHCAL_NAV_BASE_RULES_
    FIXES_001.md's "arrow down=none" hint-fold): rebuild to the locked
    base rules' shape -
      0. exit phcal from anywhere
      1..N. single mode on that one present robot (one row per robot)
      *. multi mode on all detected robots (replaces the old "y" row,
         DEFAULT highlight, matching "*" = the tool-wide all-flag)
      (last row) none mode on all robots (dry-run)
    "0" raises _PhcalEscExit (full clean exit, same as ESC) - distinct
    from the none-row, which continues the session dry.

    2026-08-21, PHCAL_LESS_FLAKY_SWEEP_001.md: the none-row's own key
    ("d") was removed outright as a first pass - reachable only by
    arrowing to it and pressing Enter, no key at all.

    2026-08-21, PHCAL_DOWN_ARROW_HOTKEY_001.md: "none" was made a
    down-arrow hotkey (no Enter needed) instead of a row in the
    arrow-navigable list.

    2026-08-21, PHCAL_SLASH_HOTKEY_001.md (operator direction, resolving
    PHCAL_FULL_ARROW_NAV_SURVEY_001.md §0's ↓ collision as Option A): the
    hotkey moved from the down-arrow to the "/" KEY - pressed from ANY
    highlight position, no Enter needed, resolves immediately, same
    mechanism as before, different trigger. This frees ↓ to mean ONLY
    "move down" here too, same as everywhere else in the file, ahead of
    full arrow-nav (finding 5) - no more screen-specific arrow meaning.
    Mnemonic, operator's own: "*" (multiply) = all, "/" (divide) = none.
    "/" is intercepted BEFORE the typed-buffer append (the "char" branch
    below), so it never becomes part of a typed sequence - pressing it
    always resolves immediately, it can never be "typed as part of a
    longer answer" by accident. Still a deliberate, LOCAL departure from
    arrow_column_pick()'s own contract (arrows-move/Enter-confirms) - not
    a change to that shared engine; this function still has its own small
    raw-key loop (same _read_key()/_raw_mode() primitives) because this
    one screen's semantics differ. UP and DOWN both now just cycle the
    highlight (wrapping) among the 4 real options (0/1/2.../*), same as
    any other screen. The none-row is shown as a static hint line under
    the real options with a leading "/" (not a real option in `choices`,
    not reachable by typing-then-Enter or by arrow+Enter on a highlight -
    the "/" keypress is its only path). "0"/digits/"*" typed-then-Enter
    still work exactly as before, including exit-on-invalid for anything
    unrecognized."""
    options = [("0", "exit phcal from anywhere")]
    options += [(p["which"], f"single mode on {p['label']}") for p in present_robots]
    options += [("*", "multi mode on all detected brobots")]
    choices = [k for k, _label in options]
    highlight = choices.index("*")
    typed = []

    def _redraw(first=False):
        if not first:
            sys.stdout.write(f"\x1b[{len(options) + 1}A")
        for i, (key, label) in enumerate(options):
            marker = "> " if i == highlight else "  "
            sys.stdout.write(f"\r\x1b[2K{marker}{key}. {label}\n")
        sys.stdout.write(f"\r\x1b[2K  /  none mode on all brobots (dry-run)\n")
        sys.stdout.write(f"\r\x1b[2K(typed: {''.join(typed)})")
        sys.stdout.flush()

    print("multi mode detected - continue:")
    _redraw(first=True)

    with _raw_mode(sys.stdin.fileno()):
        while True:
            kind, val = _read_key()
            if kind == "esc":
                print()
                raise _PhcalEscExit()
            if kind == "char" and val == "/":
                print()
                print(f"{clock.prefix()}PHCAL_MODE_OVERRIDE mode=none (operator chose dry-run over detected multi)")
                return [], "none"
            if kind == "arrow":
                if val == "up":
                    highlight = (highlight - 1) % len(options)
                else:
                    highlight = (highlight + 1) % len(options)
                typed = []
                _redraw()
                continue
            if kind == "backspace":
                if typed:
                    typed.pop()
                    _redraw()
                continue
            if kind == "enter":
                print()
                if typed:
                    raw = "".join(typed).strip().lower()
                else:
                    raw = choices[highlight]
                if raw not in choices:
                    print(f"PHCAL_GUIDED_INVALID_EXIT '{raw}' is not a control on this menu - exiting")
                    raise _PhcalEscExit()
                choice = raw
                break
            if kind == "char" and val:
                typed.append(val)
                sys.stdout.write(val)
                sys.stdout.flush()
                continue

    if choice == "0":
        raise _PhcalEscExit()
    if choice == "*":
        return present_robots, "multi"
    match = next(p for p in present_robots if p["which"] == choice)
    print(f"{clock.prefix()}PHCAL_MODE_OVERRIDE mode=single robot={match['which']} ({match['label']}) (operator chose single over detected multi)")
    return [match], "single"


def _low_voltage_gate(clock, present_robots):
    """After detect-first probes: if ANY present robot's battery reads
    below BATTERY_MIN_VOLTS (the same threshold run_battery_check() already
    gates individual motor calls on - reused, not a new number), warn
    plainly (which robot, its actual level) and ask before proceeding.
    Routed through the same _prompt_pick() single choke-point every other
    y/n CONFIRM in this file already uses (2026-08-21,
    PHCAL_NAV_CONSOLIDATION_001.md - was _prompt_choice(), same wrapper
    swap, unchanged args/behavior). default='n' - the operator must type
    something other than Enter/'n' to continue with a low robot present.
    volts=None on a "present" robot shouldn't happen in practice
    (_probe_candidate already treats an unparseable read as not-present),
    kept here as a defensive belt-and-suspenders low-battery treatment
    rather than assumed safe."""
    low = [r for r in present_robots if r["volts"] is None or r["volts"] < BATTERY_MIN_VOLTS]
    if not low:
        return True
    for r in low:
        volts_str = f"{r['volts']:.3f}" if r["volts"] is not None else "unknown"
        print(f"{clock.prefix()}PHCAL_LOW_BATTERY_WARNING label={r['label']} esn={r['esn']} volts={volts_str} min={BATTERY_MIN_VOLTS}")
    choice = _prompt_pick(
        "one or more present robots read low battery (see warning above) - proceed anyway? y/n [default n]: ",
        {"y", "n"}, default="n",
    )
    return choice == "y"


def _resolve_session_mode_once():
    """2026-08-19, PHCAL_NAV_POLISH_001.md addendum (operator review of the
    held nav-polish build): the detect-first probe, multi-mode confirm, and
    low-voltage gate used to live inside _run_guided_flow_once() itself, so
    every "continue" at the continue-or-exit prompt re-ran the whole ~1-2s
    probe (and could re-ask the multi-mode confirm question) on every loop
    pass. The robots present don't change between menu passes within one
    `phcal` invocation, so this now runs exactly ONCE, called by
    run_guided_flow() before its stay-in loop starts.

    Sets module-level _SESSION_MODE/_PRESENT_ROBOTS exactly as before -
    every downstream reader (_prompt_robot(), the robot_info/
    brobots_announce_in_sync dispatch branches) is unaffected by where the
    probe itself now runs. Returns True if OK to proceed, False if the
    low-voltage gate declined - the caller (run_guided_flow()) treats a
    decline as a blocked first pass so it still reaches the continue-or-exit
    prompt per operator decision 5, rather than dumping the operator
    straight out; it does not re-probe on a later "continue" since this
    function only ever runs once per invocation.

    2026-08-21, PHCAL_NAV_BASE_RULES_FIXES_001.md (PHCAL_NAV_BASE_RULES_
    SURVEY_001.md §5 step 1): GOPOD_ALLOW_LIVE_ROBOT_SPEECH is now decided
    HERE, once, from the resolved mode - not by a separate shell-side
    live_robots_prompt() gate in brobots.sh's phcal() (removed the same
    pass; that function still exists and is still used by phcal()'s
    direct-flag CLI form, untouched). multi/single (the operator chose to
    continue, or detect-first found exactly one robot) sets it to "1";
    none (detected, or explicitly chosen via _confirm_multi_mode's none/
    dry-run row) clears it - a hard set/pop either way, so a stray pre-exported value
    from the shell never silently overrides what this screen just
    resolved. Every dispatch branch already reads this fresh via
    `os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"` at fire time, so
    setting it once here, before the guided-flow loop starts, is
    sufficient - no per-primitive change needed."""
    global _SESSION_MODE, _PRESENT_ROBOTS
    _detect_clock = _Clock()
    _PRESENT_ROBOTS, _SESSION_MODE = _detect_present_robots(_detect_clock)
    if _SESSION_MODE == "multi":
        _PRESENT_ROBOTS, _SESSION_MODE = _confirm_multi_mode(_detect_clock, _PRESENT_ROBOTS)
    if not _low_voltage_gate(_detect_clock, _PRESENT_ROBOTS):
        print(f"{_detect_clock.prefix()}PHCAL_GUIDED_EXIT declined to proceed with a low-battery robot present")
        return False
    if _SESSION_MODE == "none":
        os.environ.pop("GOPOD_ALLOW_LIVE_ROBOT_SPEECH", None)
        print(f"{_detect_clock.prefix()}PHCAL_SESSION_MODE mode=none - no configured robots responded; session continues, live-firing primitives will report unreachable")
    elif _SESSION_MODE == "single":
        os.environ["GOPOD_ALLOW_LIVE_ROBOT_SPEECH"] = "1"
        print(f"{_detect_clock.prefix()}PHCAL_SESSION_MODE mode=single robot={_PRESENT_ROBOTS[0]['label']}")
    else:
        os.environ["GOPOD_ALLOW_LIVE_ROBOT_SPEECH"] = "1"
        print(f"{_detect_clock.prefix()}PHCAL_SESSION_MODE mode=multi robots={[r['label'] for r in _PRESENT_ROBOTS]}")
    return True


def cmd_animation(mod, clock, serial, robot, live, animation_token, hold_seconds, pre_assumed=False):
    """Fires one of the three catalog-confirmed animation tokens
    (kgSuccess/searching/answering) standalone, via the same
    {{playAnimationWI||TOKEN}}-over-say_text dispatch
    run_golden_song_001.py's own run_animation_only() uses - ported here
    verbatim (not imported - phcal loads run_section1_full_live_001.py as
    `mod`, not the golden engine module that function lives in).
    searching/answering loop the dispatch every
    _ANIMATION_LOOP_INTERVAL_SECONDS until hold_seconds elapses; kgSuccess
    (and anything else) fires once, then holds. Assume/release wrap the
    whole call once, same shape as every other bench primitive here."""
    loop = animation_token in _LOOP_ANIMATION_TOKENS
    print(
        f"{clock.prefix()}PHCAL_ANIMATION robot={robot} serial={serial} token={animation_token} "
        f"hold_seconds={hold_seconds} loop={loop}"
    )
    getout_token = _ANIMATION_GETOUT_TOKENS.get(animation_token)
    if not live:
        if pre_assumed:
            print(f"{clock.prefix()}PHCAL_ANIMATION DRY: pre_assumed=True - would skip own assume (brobots_wake already holds control)")
        else:
            print(f"{clock.prefix()}PHCAL DRY: would assume_behavior_control serial={serial} priority=high")
        if loop:
            print(
                f"{clock.prefix()}PHCAL DRY: would say_text {{{{playAnimationWI||{animation_token}}}}} "
                f"every {_ANIMATION_LOOP_INTERVAL_SECONDS}s until {hold_seconds}s elapsed"
            )
            if getout_token:
                print(
                    f"{clock.prefix()}PHCAL DRY: would then say_text {{{{playAnimationWI||{getout_token}}}}} "
                    f"once, to end the {animation_token} loop cleanly"
                )
            else:
                print(
                    f"{clock.prefix()}PHCAL DRY: no confirmed getout transition for '{animation_token}' - "
                    f"would settle {_ANIMATION_NO_GETOUT_RELEASE_SETTLE_SECONDS}s (native Wire-Pod stop "
                    "discipline, ANSWERING_NATIVE_STOP_DISCIPLINE_001.md) before releasing, instead of "
                    "releasing immediately"
                )
        else:
            print(
                f"{clock.prefix()}PHCAL DRY: would say_text {{{{playAnimationWI||{animation_token}}}}} "
                f"once, then hold {hold_seconds}s"
            )
        print(f"{clock.prefix()}PHCAL DRY: would release_behavior_control serial={serial}")
        print(f"{clock.prefix()}PHCAL_ANIMATION_COMPLETE ok=True mode=dry")
        return 0

    # 2026-08-08 fix: the first live spot-check (kgSuccess/searching on the
    # first two calls) got ok=True on every HTTP call but no visible robot
    # response - the exact "cold first press" pattern already root-caused and
    # fixed for arm/nod (see run_assume_control()'s own comment above): the
    # very first live action after a fresh assume fires ~2ms later, before
    # the robot is physically ready, and is silently dropped. This function
    # was ported from run_animation_only() (run_golden_song_001.py), which
    # has the same zero-settle gap - never hit there because emotion_beat's
    # own pre_animation_pause_seconds (2.0s default, spoken text first) masks
    # it, but a plain animation note with nothing said first has nothing to
    # mask it. Fixed here by routing through run_assume_control() instead of
    # a bespoke inline assume call - the same PHCAL_ASSUME_SETTLE_SECONDS
    # settle every other bench primitive already gets, not a new mechanism.
    # pre_assumed, 2026-08-09 (brobots_wake chain toggle): see cmd_arm's own
    # comment - same convention, skip own assume when already held.
    if pre_assumed:
        print(f"{clock.prefix()}PHCAL_ANIMATION pre_assumed=True - skipping own assume (brobots_wake already holds control)")
        ok = True
    else:
        assume = run_assume_control(mod, clock, serial, live)
        ok = assume["ok"]
    anim_payload = f"{{{{playAnimationWI||{animation_token}}}}}"
    dispatches = []
    try:
        if loop:
            elapsed = 0.0
            while elapsed < hold_seconds:
                dispatches.append(
                    mod.wirepod_web_send_form("/api-sdk/say_text", {"text": anim_payload, "serial": serial}, timeout=10)
                )
                time.sleep(_ANIMATION_LOOP_INTERVAL_SECONDS)
                elapsed += _ANIMATION_LOOP_INTERVAL_SECONDS
            if getout_token:
                getout_payload = f"{{{{playAnimationWI||{getout_token}}}}}"
                print(f"{clock.prefix()}PHCAL_ANIMATION firing getout token={getout_token} (ends the {animation_token} loop cleanly)")
                dispatches.append(
                    mod.wirepod_web_send_form("/api-sdk/say_text", {"text": getout_payload, "serial": serial}, timeout=10)
                )
            else:
                print(
                    f"{clock.prefix()}PHCAL_ANIMATION no confirmed getout for '{animation_token}' - "
                    f"settling {_ANIMATION_NO_GETOUT_RELEASE_SETTLE_SECONDS}s before release (native "
                    "Wire-Pod stop discipline) instead of releasing immediately"
                )
                time.sleep(_ANIMATION_NO_GETOUT_RELEASE_SETTLE_SECONDS)
        else:
            dispatches.append(
                mod.wirepod_web_send_form("/api-sdk/say_text", {"text": anim_payload, "serial": serial}, timeout=10)
            )
            time.sleep(hold_seconds)
        print(f"{clock.prefix()}PHCAL_ANIMATION token={animation_token} fired {len(dispatches)}x")
        ok = ok and bool(dispatches) and all(d.get("status") == 200 for d in dispatches)
    finally:
        release = run_release_control(mod, clock, serial, live)
        ok = ok and release["ok"]
    print(f"{clock.prefix()}PHCAL_ANIMATION_COMPLETE ok={ok}")
    return 0 if ok else 1


def cmd_hold(mod, clock, serial, robot, live, hold_seconds, pre_assumed=False):
    """Assume/release bench test (PHCAL_CANDIDATE_CONTROLS_SURVEY_001.md
    Part 2) - lets the operator watch a robot held at priority=high (the same
    OVERRIDE_BEHAVIORS level every golden-engine note already uses) genuinely
    stay put for hold_seconds, then resume normal behavior the instant it's
    released. Pure composition, no new HTTP shape: reuses
    run_assume_control()/run_release_control() verbatim, the same two
    functions cmd_arm/cmd_nod already call once per whole call."""
    print(f"{clock.prefix()}PHCAL_HOLD robot={robot} serial={serial} hold_seconds={hold_seconds}")
    # pre_assumed, 2026-08-09 (brobots_wake chain toggle): see cmd_arm's own
    # comment - same convention, skip own assume when already held.
    if pre_assumed:
        print(f"{clock.prefix()}PHCAL_HOLD pre_assumed=True - skipping own assume (brobots_wake already holds control)")
        ok = True
    else:
        assume = run_assume_control(mod, clock, serial, live)
        ok = assume["ok"]
    if not live:
        print(f"{clock.prefix()}PHCAL_HOLD DRY: would hold {hold_seconds}s (watch-for-wander is only meaningful live)")
    else:
        print(f"{clock.prefix()}PHCAL_HOLD holding {hold_seconds}s - watch the robot, it should NOT move/wander")
        remaining = hold_seconds
        while remaining > 0:
            step = min(1.0, remaining)
            time.sleep(step)
            remaining = round(remaining - step, 3)
            if remaining > 0:
                print(f"{clock.prefix()}PHCAL_HOLD {remaining:.1f}s remaining")
            else:
                print(f"{clock.prefix()}PHCAL_HOLD holding complete")
    release = run_release_control(mod, clock, serial, live)
    ok = ok and release["ok"]
    tail = " - release fired, robot should resume normal behavior now" if live else ""
    print(f"{clock.prefix()}PHCAL_HOLD_COMPLETE ok={ok}{tail}")
    return 0 if ok else 1


def cmd_brobots_wake(mod, clock, serial, robot, live):
    """2026-08-09, DECOMPOSED per operator direction. This is the wake
    half: conn_test -> assume -> golden-flag release/settle/re-assume pulse
    -> release. NOT wheel-specific - any future motion primitive that needs
    to prove a control session is genuinely live before acting on it can
    call the same core functions this composes (run_wheel_nudge_wake_check/
    run_assume_control/run_wheel_nudge_wake_release_pulse/
    run_release_control). Standalone-fireable and self-contained - releases
    at the end, doesn't leave the robot stuck holding control. Mechanism
    REPLACED 2026-08-10, see WHEEL_NUDGE_WAKE_RELEASE_SETTLE_SECONDS's own
    comment above for the full why. Builds and prints the standardized
    readiness signal (_readiness_signal above) before returning - REST
    channel's own gate-if-not-ready endpoint, same shape _finish_sleep_
    wake's direct-SDK channel gate returns."""
    print(f"{clock.prefix()}PHCAL_BROBOTS_WAKE robot={robot} serial={serial}")
    wake_check = run_wheel_nudge_wake_check(mod, clock, serial, live)
    assume = run_assume_control(mod, clock, serial, live)
    ok = wake_check["ok"] and assume["ok"]
    try:
        wake_confirm = run_wheel_nudge_wake_release_pulse(mod, clock, serial, live)
        ok = ok and wake_confirm["ok"]
    finally:
        release = run_release_control(mod, clock, serial, live)
        ok = ok and release["ok"]
    # reason resolved AFTER every sub-call has run (not inline as each one
    # fires) so a later failure (e.g. the final release) is never masked by
    # an earlier step's success - each check below is independent, in the
    # same order the calls themselves fire.
    if not wake_check["ok"]:
        reason = "conn_test failed - robot unreachable"
    elif not assume["ok"]:
        reason = "assume_behavior_control failed"
    elif not wake_confirm["ok"]:
        reason = "release/settle/reassume golden-flag pulse failed"
    elif not release["ok"]:
        reason = "final release_behavior_control failed"
    else:
        reason = "control session confirmed responsive (golden-flag pulse settled)"
    signal = _readiness_signal(
        ok, reason, "rest_golden_flag",
        detail={"wake_check": wake_check, "assume": assume, "wake_confirm": wake_confirm, "release": release},
    )
    print(f"{clock.prefix()}PHCAL_BROBOTS_WAKE_COMPLETE ok={ok}")
    print(f"{clock.prefix()}PHCAL_READY ready={signal['ready']} reason={signal['reason']!r} channel=rest_golden_flag")
    return 0 if signal["ok"] else 1


def cmd_move_reverse(mod, clock, serial, robot, live, hold_seconds, pre_assumed=False):
    """The move half of the same decomposition (see cmd_brobots_wake above)
    - assumes ITS OWN control (unless pre_assumed=True, see cmd_arm's own
    comment - same convention), fires run_move_wheels_no_control (prime +
    move + hold + stop), releases. Fixed ON-CHARGER reverse pulse only
    (MOVE_REVERSE_LW/RW, the webroot's own confirmed S-key/backward
    values) - forward is a later, separate decision, not built here.
    Standalone-fireable and self-contained, same shape as cmd_arm/cmd_nod."""
    print(f"{clock.prefix()}{MOVE_REVERSE_CAUTION}")
    print(
        f"{clock.prefix()}PHCAL_MOVE_REVERSE robot={robot} serial={serial} "
        f"lw={MOVE_REVERSE_LW} rw={MOVE_REVERSE_RW} hold_seconds={hold_seconds}"
    )
    if pre_assumed:
        print(f"{clock.prefix()}PHCAL_MOVE_REVERSE pre_assumed=True - skipping own assume (brobots_wake already holds control)")
        ok = True
    else:
        assume = run_assume_control(mod, clock, serial, live)
        ok = assume["ok"]
    try:
        move = run_move_wheels_no_control(
            mod, clock, serial, MOVE_REVERSE_LW, MOVE_REVERSE_RW, hold_seconds, live
        )
        ok = ok and move["ok"]
    finally:
        release = run_release_control(mod, clock, serial, live)
        ok = ok and release["ok"]
    print(f"{clock.prefix()}PHCAL_MOVE_REVERSE_COMPLETE ok={ok}")
    return 0 if ok else 1


def _weather_api_config():
    """Reads provider/key/unit from Wire-Pod's own apiConfig.json - read-only,
    same single source of truth CLAUDE.md already names for this file, never
    copied or duplicated here."""
    path = os.path.expanduser("~/wire-pod/chipper/apiConfig.json")
    with open(path) as f:
        cfg = json.load(f)
    return cfg.get("weather", {})


def _weather_geocode_candidates(location):
    """Same comma-grouping order as weather.go's geocodeOpenWeatherMap() -
    ported by hand for this API-only test, not imported (the Go source is
    the live-load-bearing copy; this mirrors it so the test proves the same
    fix the robot actually runs). Keep in sync if that Go function's
    candidate order ever changes."""
    fields = location.replace(",", " ").split()
    if not fields:
        return [location]
    seen = set()
    candidates = []

    def add(s):
        s = s.strip()
        if s and s not in seen:
            seen.add(s)
            candidates.append(s)

    add(location)
    add(" ".join(fields))
    if len(fields) >= 2:
        add(" ".join(fields[:-1]) + ", " + fields[-1])
    if len(fields) >= 3:
        add(", ".join(fields))
    return candidates


def cmd_weather_test(location):
    """phcal option 4 - API-only weather test, added 2026-08-02 to let the
    just-fixed geocode candidate logic be checked against any spoken-style
    phrase in seconds. Touches no robot at all: no behavior control, no
    direct SDK, nothing that can fire hardware - plain HTTP calls to
    OpenWeatherMap, same shape as the ad-hoc check that confirmed the
    original bug (a bare query like "windsor ontario canada" returning zero
    geocode matches) and the fix (trying comma-grouping candidates in
    order)."""
    cfg = _weather_api_config()
    if not cfg.get("enable") or not cfg.get("key"):
        print("PHCAL_WEATHER_BLOCKED weather API not enabled or key not set in apiConfig.json")
        return 1
    if cfg.get("provider") != "openweathermap.org":
        print(f"PHCAL_WEATHER_BLOCKED provider={cfg.get('provider')!r} not supported by this test (openweathermap.org only)")
        return 1
    key = cfg["key"]
    unit = cfg.get("unit") or "F"

    matched = None
    tried = _weather_geocode_candidates(location)
    for candidate in tried:
        geo_url = (
            "http://api.openweathermap.org/geo/1.0/direct?q="
            + urllib.parse.quote(candidate) + "&limit=1&appid=" + key
        )
        try:
            with urllib.request.urlopen(geo_url, timeout=10) as resp:
                result = json.loads(resp.read().decode())
        except Exception as exc:
            print(f"PHCAL_WEATHER_GEO_ERROR candidate={candidate!r} {exc}")
            continue
        if result:
            matched = (candidate, result[0])
            break

    if matched is None:
        print(f"PHCAL_WEATHER_NO_MATCH location={location!r} tried={tried}")
        return 1

    candidate, geo = matched
    lat, lon = geo["lat"], geo["lon"]
    print(
        f"PHCAL_WEATHER_GEOCODE matched_on={candidate!r} name={geo['name']} "
        f"state={geo.get('state')} country={geo['country']} lat={lat} lon={lon}"
    )

    units = "imperial" if unit == "F" else "metric"
    weather_url = (
        f"https://api.openweathermap.org/data/2.5/weather?lat={lat}&lon={lon}"
        f"&units={units}&appid={key}"
    )
    try:
        with urllib.request.urlopen(weather_url, timeout=10) as resp:
            weather = json.loads(resp.read().decode())
    except Exception as exc:
        print(f"PHCAL_WEATHER_FETCH_ERROR {exc}")
        return 1

    temp = weather.get("main", {}).get("temp")
    condition = weather.get("weather", [{}])[0].get("main")
    description = weather.get("weather", [{}])[0].get("description")
    print(
        f"PHCAL_WEATHER_RESULT location={geo['name']} temp={temp}{unit} "
        f"condition={condition} ({description})"
    )
    return 0


def cmd_tempo_calibration():
    """phcal option 12 - tempo calibration, added 2026-08-11 at operator
    direction (fold tempo-set into phcal's menu). This function only owns
    the song/mode/step picking UI, same shape as brobots.sh's own
    `tempo-set` alias (song list -> A/B mode -> value/step -> dry preview ->
    y/n confirm -> --yes write). Every actual read/write is handed off to
    tempo_set_001.py's own cmd_set_global/cmd_set_buffer - loaded and called
    directly, not reimplemented, so there is exactly one place tempo values
    are ever written. No robot, no hardware, no preflight - a song's
    knobs.json is the only thing this touches."""
    tempo_mod = _load_module(TEMPO_SET_PATH, "tempo_set_001")

    song_dirs = sorted(
        d.name for d in SONGS_DIR.iterdir() if d.is_dir() and d.name != "zzz_archives"
    )
    if not song_dirs:
        print(f"PHCAL_TEMPO_BLOCKED no songs found under {SONGS_DIR}")
        return 1
    print("PHCAL_TEMPO songs on disk:")
    for i, name in enumerate(song_dirs, start=1):
        print(f"  {i}. {name}")
    dir_choice = _prompt_pick(
        f"Pick a song [1-{len(song_dirs)}]: ", {str(i) for i in range(1, len(song_dirs) + 1)} | {"0"},
        exit_on_invalid=True,
    )
    if dir_choice == "0":
        raise _PhcalEscExit()
    chosen_dir = song_dirs[int(dir_choice) - 1]
    knobs_path = SONGS_DIR / chosen_dir / "knobs.json"
    if not knobs_path.exists():
        print(f"PHCAL_TEMPO_BLOCKED no knobs.json found for {chosen_dir} at {knobs_path}")
        return 1

    print("Mode:")
    print("  A. set the song's GLOBAL tempo (whole-song ease)")
    print("  B. set ONE step's tempo_factor (+ optional comment)")
    # _prompt_pick always lowercases raw input before matching/returning
    # (see arrow_column_pick()'s own docstring) - choices/comparisons here
    # follow that contract, not the bash tempo-set's own case-insensitive
    # if-chain. numbered/lettered PICK -> exit_on_invalid=True, standardized
    # by shape, PHCAL_NAV_CONSOLIDATION_001.md. "0" added as a deliberate
    # exit choice, PHCAL_NAV_BASE_RULES_FIXES_001.md.
    mode_choice = _prompt_pick("pick a mode [A/B]: ", {"a", "b", "0"}, exit_on_invalid=True)
    if mode_choice == "0":
        raise _PhcalEscExit()

    if mode_choice == "a":
        value = input("new global_tempo (0.0-9.9): ").strip()
        tempo_mod.cmd_set_global(["set-global", value, "--knobs", str(knobs_path)])
        apply_choice = _prompt_pick("apply the above? y/n [default n]: ", {"y", "n"}, default="n")
        if apply_choice == "y":
            return tempo_mod.cmd_set_global(["set-global", value, "--yes", "--knobs", str(knobs_path)])
        return 0

    # Mode B - numbered step menu, same reasoning as tempo-set's own bash
    # version: read the ACTIVE (dirty-first resolved) steps array through
    # the same shared resolver the write path uses, so the numbered list
    # always matches whatever file the write will actually land in.
    envelope_mod = _load_module(KNOBS_ENVELOPE_PATH, "knobs_envelope_001")
    _, envelope = envelope_mod.load_knobs_envelope(str(SONGS_DIR / chosen_dir))
    step_ids = [s.get("step_id") for s in envelope.get("steps", [])]
    if not step_ids:
        print(f"PHCAL_TEMPO_BLOCKED no steps found for {chosen_dir}")
        return 1
    print("steps:")
    for i, sid in enumerate(step_ids, start=1):
        print(f"  {i}. {sid}")
    step_choice = _prompt_pick(
        f"pick a step [1-{len(step_ids)}]: ", {str(i) for i in range(1, len(step_ids) + 1)} | {"0"},
        exit_on_invalid=True,
    )
    if step_choice == "0":
        raise _PhcalEscExit()
    step_id = step_ids[int(step_choice) - 1]
    factor = input("new tempo_factor (default 1.0): ").strip() or "1.0"
    comment = input("tempo_comment (optional, Enter to leave unchanged): ").strip()
    comment_flag = ["--comment", comment] if comment else []

    tempo_mod.cmd_set_buffer(["set-buffer", step_id, "--factor", factor, *comment_flag, "--knobs", str(knobs_path)])
    apply_choice = _prompt_pick("apply the above? y/n [default n]: ", {"y", "n"}, default="n")
    if apply_choice == "y":
        return tempo_mod.cmd_set_buffer(
            ["set-buffer", step_id, "--factor", factor, *comment_flag, "--yes", "--knobs", str(knobs_path)]
        )
    return 0


# Rung-1 (bingo-sourced) defaults - unchanged by rung 2/3, only now
# overridable per call via flags (rung 2) or the guided flow (rung 3).
# nod's "count" key is new in rung 3 - rungs 1/2 hardcoded the same value
# (1) as the positional default inline; added here too so the guided flow's
# first-run seed (see _load_last below) has exactly one source of truth for
# every default, not a second copy.
_DEFAULTS = {
    "arm": {"hold_seconds": 1.2, "speed": 2, "cycles": 1},
    "nod": {"hold_seconds": 0.35, "speed": 2, "count": 1},
    # 2026-07-22, Part C: rattle has no hold/speed of its own - see
    # cmd_rattle's own comment for why there is genuinely no "hold" step to
    # default here. volume is phcal's own 1-5 bench scale (mapped to the
    # real 1-100 AudioVolume range at fire time), default 5 = full volume.
    "rattle": {"volume": 5},
    # 2026-08-12: same 1-5 bench scale as rattle, same reasoning - see
    # DANGER_WAV_PATH's own comment above for what this primitive is.
    "danger": {"volume": 5},
}

# Which flags each primitive accepts. --cycles is arm-only: nod already has
# its own positional [count] for "how many," so there is exactly one way to
# say "how many" per primitive, not two competing ones.
_ALLOWED_FLAGS = {
    "arm": {"--hold", "--speed", "--cycles"},
    "nod": {"--hold", "--speed"},
    "rattle": {"--volume"},
    "danger": {"--volume"},
    "cube": set(),
}


def _load_last():
    """Reads phcal_last.json (this tool's own memory file). Missing file,
    unreadable JSON, or a missing primitive/key inside it all fall back to
    the rung-1 (bingo-sourced) defaults above for exactly the piece that's
    missing - never a crash, never a guessed number outside those defaults.
    Returns a dict keyed "arm"/"nod", each a plain {cycles|count, hold,
    speed} dict - the same key names phcal_last.json itself uses (note:
    "hold", not "hold_seconds" - matches the spec's own key naming)."""
    data = {}
    if LAST_PATH.exists():
        try:
            with open(LAST_PATH) as f:
                data = json.load(f)
        except (json.JSONDecodeError, OSError):
            data = {}

    seed = {
        "arm": {
            "cycles": _DEFAULTS["arm"]["cycles"],
            "hold": _DEFAULTS["arm"]["hold_seconds"],
            "speed": _DEFAULTS["arm"]["speed"],
        },
        "nod": {
            "count": _DEFAULTS["nod"]["count"],
            "hold": _DEFAULTS["nod"]["hold_seconds"],
            "speed": _DEFAULTS["nod"]["speed"],
        },
        "rattle": {
            "volume": _DEFAULTS["rattle"]["volume"],
        },
        "danger": {
            "volume": _DEFAULTS["danger"]["volume"],
        },
        # 2026-08-15 (MASTER_TWEAKS_STAGE1_SAVE_COVERAGE_001.md): 5 more
        # tunable primitives widened in - required, not optional, the
        # moment _save_last() starts writing these keys. _save_last()
        # calls _load_last() internally to get the base dict before
        # writing; if this loop below didn't know about these keys too,
        # the very next save from ANY primitive (old or new) would
        # silently drop them from the reloaded dict and erase them on
        # write. Same seed/merge shape as the original 4 - no new
        # mechanism.
        "animation": {"hold": ANIMATION_DEFAULT_HOLD_SECONDS},
        "brobots_stay_in_place": {"hold": HOLD_DEFAULT_SECONDS},
        "move_reverse": {"hold": MOVE_REVERSE_DEFAULT_HOLD_SECONDS},
        "brobots_announce_in_sync": {"phrase": "Brobots ready!"},
        "brobots_sleep_to_wake_direct_sdk": {"wait": SLEEP_WAKE_DEFAULT_WAIT_SECONDS},
    }
    merged = {}
    for primitive in (
        "arm", "nod", "rattle", "danger",
        "animation", "brobots_stay_in_place", "move_reverse",
        "brobots_announce_in_sync", "brobots_sleep_to_wake_direct_sdk",
    ):
        merged[primitive] = {**seed[primitive], **(data.get(primitive) or {})}
    return merged


def _save_last(primitive, values):
    """Writes {primitive: values} into phcal_last.json, preserving whatever
    is already saved for the OTHER primitive untouched. Both entry paths
    (guided flow and direct-flag form) call this after firing, so either one
    updates the same one memory file."""
    data = _load_last()
    data[primitive] = values
    LAST_PATH.parent.mkdir(parents=True, exist_ok=True)
    with open(LAST_PATH, "w") as f:
        json.dump(data, f, indent=2)
        f.write("\n")


# 2026-08-15, MASTER_TWEAKS_STAGE3_PROMOTE_001.md: phcal_last.json is
# phcal-global, not per-song, so there's no natural per-song promoted
# counterpart the way zKnobs.json/knobs.json have - this is the one
# tracked snapshot every promoted tuning value lands in.
PHCAL_PROMOTED_TWEAKS_PATH = GOPOD_REPO_ROOT / "tech/alias_play_studio/phcal_master_tweaks.json"


def _promote_tweaks():
    """Standalone promote step for phcal's own master tweaks file - mirrors
    run_golden_song_001.py's own _maybe_promote_knobs() shape exactly: same
    default-n y/n prompt, same shutil.copyfile byte-identical copy (no
    re-serialize), same only-if-source-exists guard. Not a new mechanism.
    Never auto-fired at the end of a phcal run - phcal gets invoked many
    times per tuning session (unlike a golden song's single end-of-run
    prompt), so a y/n on every fire would be noise; this is reachable on
    demand only (the phcal-promote alias), when the operator has actually
    landed on values worth banking."""
    if not LAST_PATH.exists():
        print(f"PHCAL_TWEAKS_PROMOTE_BLOCKED no {LAST_PATH} to promote from")
        return 1
    try:
        choice = _prompt_pick(
            f"update {PHCAL_PROMOTED_TWEAKS_PATH} from {LAST_PATH} for commit? y/n [default n]: ",
            {"y", "n"}, default="n",
        )
    except (EOFError, KeyboardInterrupt, _PhcalEscExit):
        choice = "n"
    if choice == "y":
        PHCAL_PROMOTED_TWEAKS_PATH.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(LAST_PATH, PHCAL_PROMOTED_TWEAKS_PATH)
        print(f"PHCAL_TWEAKS_PROMOTED {PHCAL_PROMOTED_TWEAKS_PATH} updated from {LAST_PATH}")
        return 0
    print(f"PHCAL_TWEAKS_PROMOTE_SKIPPED {PHCAL_PROMOTED_TWEAKS_PATH} left untouched")
    return 0


def _parse_rest(primitive, rest):
    """Splits the argv tail (everything after <primitive> <robot>) into a
    leading positional list and a flags dict. Recognizes --hold/--speed/
    --cycles (each takes the next token as its value); anything else
    starting with '--' is an unknown-flag error; anything not starting with
    '--' is a positional (nod's optional [count]; arm takes none). Raises
    ValueError with a plain message on any problem - never guesses."""
    positionals = []
    flags = {}
    i = 0
    while i < len(rest):
        tok = rest[i]
        if tok.startswith("--"):
            if tok == "--cycles" and primitive == "nod":
                raise ValueError("--cycles is arm-only; nod repeats via the positional [count] instead")
            if tok not in _ALLOWED_FLAGS[primitive]:
                raise ValueError(f"unknown flag '{tok}' for '{primitive}'")
            if i + 1 >= len(rest):
                raise ValueError(f"{tok} needs a value")
            flags[tok] = rest[i + 1]
            i += 2
        else:
            positionals.append(tok)
            i += 1
    return positionals, flags


def _parse_float_flag(name, raw, min_exclusive=0.0, max_value=None):
    try:
        val = float(raw)
    except ValueError:
        raise ValueError(f"{name} must be a number, got '{raw}'")
    if val <= min_exclusive:
        raise ValueError(f"{name} must be > {min_exclusive}, got {val}")
    # max_value - additive, defaults to None (every existing caller passes
    # none, so their behavior is unchanged). Refuse (raise, re-prompt),
    # never clamp - see MOVE_REVERSE_MAX_HOLD_SECONDS's own comment for why
    # this primitive needs it.
    if max_value is not None and val > max_value:
        raise ValueError(f"{name} must be <= {max_value} (BLOCKED - refusing to fire), got {val}")
    return val


def _parse_int_flag(name, raw, min_value=1, max_value=None):
    try:
        val = int(raw)
    except ValueError:
        raise ValueError(f"{name} must be an integer, got '{raw}'")
    if val < min_value:
        raise ValueError(f"{name} must be >= {min_value}, got {val}")
    if max_value is not None and val > max_value:
        raise ValueError(f"{name} must be <= {max_value}, got {val}")
    return val


def main():
    if len(sys.argv) < 3:
        print(
            "PHCAL_ISOLATE_USAGE phcal_isolate_001.py <arm|nod|rattle|danger|cube> <robot 1|2> "
            "[reps] [--hold S] [--speed N] [--cycles N] [--volume N]"
        )
        print("  (call with zero arguments for the guided prompt flow instead)")
        return 1

    primitive = sys.argv[1]
    robot = sys.argv[2]
    if robot not in ("1", "2"):
        print(f"PHCAL_ISOLATE_BLOCKED robot must be 1 or 2, got '{robot}'")
        return 1
    if primitive not in ("arm", "nod", "rattle", "danger", "cube"):
        print(f"PHCAL_ISOLATE_BLOCKED unknown primitive '{primitive}' - expected arm, nod, rattle, danger, or cube")
        return 1

    try:
        positionals, flags = _parse_rest(primitive, sys.argv[3:])
    except ValueError as exc:
        print(f"PHCAL_ISOLATE_BLOCKED {exc}")
        return 1

    live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
    # See PHCAL_RESTART_WIREPOD_ENV's own comment block above: setdefault, not
    # a hard overwrite, and must happen before _load_module (module-level
    # constant read at import time).
    os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
    mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
    serial = BROBOT_2_SERIAL if robot == "2" else BROBOT_1_SERIAL
    clock = _Clock()

    print(f"{clock.prefix()}PHCAL_ISOLATE primitive={primitive} robot={robot} serial={serial} live={live}")
    if not live:
        print(f"{clock.prefix()}DRY: live robot gate is off (GOPOD_ALLOW_LIVE_ROBOT_SPEECH != 1)")

    if primitive == "rattle":
        if positionals:
            print(
                f"PHCAL_ISOLATE_BLOCKED rattle takes no positional argument (got {positionals!r}) "
                "- use --volume N instead"
            )
            return 1
        try:
            volume_ui = (
                _parse_int_flag("--volume", flags["--volume"], min_value=1, max_value=5)
                if "--volume" in flags
                else _DEFAULTS["rattle"]["volume"]
            )
        except ValueError as exc:
            print(f"PHCAL_ISOLATE_BLOCKED {exc}")
            return 1
        restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
        if not restart_preflight["ok"]:
            return 1
        preflight = run_preflight(mod, clock, serial, live)
        if not preflight["ok"]:
            return 1
        rc = cmd_rattle(mod, clock, serial, robot, live, volume_ui)
        _save_last("rattle", {"volume": volume_ui})
        print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=rattle path={LAST_PATH}")
        return rc

    if primitive == "danger":
        if positionals:
            print(
                f"PHCAL_ISOLATE_BLOCKED danger takes no positional argument (got {positionals!r}) "
                "- use --volume N instead"
            )
            return 1
        try:
            volume_ui = (
                _parse_int_flag("--volume", flags["--volume"], min_value=1, max_value=5)
                if "--volume" in flags
                else _DEFAULTS["danger"]["volume"]
            )
        except ValueError as exc:
            print(f"PHCAL_ISOLATE_BLOCKED {exc}")
            return 1
        restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
        if not restart_preflight["ok"]:
            return 1
        preflight = run_preflight(mod, clock, serial, live)
        if not preflight["ok"]:
            return 1
        rc = cmd_danger(mod, clock, serial, robot, live, volume_ui)
        _save_last("danger", {"volume": volume_ui})
        print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=danger path={LAST_PATH}")
        return rc

    if primitive == "cube":
        if positionals:
            print(
                f"PHCAL_ISOLATE_BLOCKED cube takes no positional or flag arguments (got {positionals!r})"
            )
            return 1
        restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
        if not restart_preflight["ok"]:
            return 1
        preflight = run_preflight(mod, clock, serial, live)
        if not preflight["ok"]:
            return 1
        rc = cmd_cube(mod, clock, serial, robot, live)
        return rc

    try:
        hold_seconds = (
            _parse_float_flag("--hold", flags["--hold"]) if "--hold" in flags else _DEFAULTS[primitive]["hold_seconds"]
        )
        speed = _parse_int_flag("--speed", flags["--speed"], min_value=1) if "--speed" in flags else _DEFAULTS[primitive]["speed"]
    except ValueError as exc:
        print(f"PHCAL_ISOLATE_BLOCKED {exc}")
        return 1

    if primitive == "arm":
        if positionals:
            print(
                f"PHCAL_ISOLATE_BLOCKED arm takes no positional argument (got {positionals!r}) "
                "- use --cycles N instead"
            )
            return 1
        try:
            cycles = (
                _parse_int_flag("--cycles", flags["--cycles"], min_value=1)
                if "--cycles" in flags
                else _DEFAULTS["arm"]["cycles"]
            )
        except ValueError as exc:
            print(f"PHCAL_ISOLATE_BLOCKED {exc}")
            return 1
        restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
        if not restart_preflight["ok"]:
            return 1
        preflight = run_preflight(mod, clock, serial, live)
        if not preflight["ok"]:
            return 1
        battery = run_battery_check(mod, clock, serial, live)
        if not battery["ok"]:
            print(f"{clock.prefix()}PHCAL_BLOCKED_LOW_BATTERY not firing arm - see PHCAL_BATTERY_CHECK line above")
            return 1
        rc = cmd_arm(mod, clock, serial, robot, live, cycles, hold_seconds, speed)
        _save_last("arm", {"cycles": cycles, "hold": hold_seconds, "speed": speed})
        print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=arm path={LAST_PATH}")
        return rc

    # primitive == "nod"
    if len(positionals) > 1:
        print(f"PHCAL_ISOLATE_BLOCKED nod takes at most one positional [count] argument, got {positionals!r}")
        return 1
    count = 1
    if positionals:
        try:
            count = _parse_int_flag("count", positionals[0], min_value=1)
        except ValueError as exc:
            print(f"PHCAL_ISOLATE_BLOCKED {exc}")
            return 1
    restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
    if not restart_preflight["ok"]:
        return 1
    preflight = run_preflight(mod, clock, serial, live)
    if not preflight["ok"]:
        return 1
    battery = run_battery_check(mod, clock, serial, live)
    if not battery["ok"]:
        print(f"{clock.prefix()}PHCAL_BLOCKED_LOW_BATTERY not firing nod - see PHCAL_BATTERY_CHECK line above")
        return 1
    rc = cmd_nod(mod, clock, serial, robot, live, count, hold_seconds, speed)
    _save_last("nod", {"count": count, "hold": hold_seconds, "speed": speed})
    print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=nod path={LAST_PATH}")
    return rc


class _PhcalEscExit(Exception):
    """2026-08-18, PHCAL_NAV_POLISH_001.md - raised by the raw-key reader
    when the operator presses ESC. Caught near the top of whatever loop is
    reading input (_run_guided_flow_once, _promote_tweaks) for a clean exit
    from ANY nesting depth, without threading a sentinel value through
    every intermediate return path between here and there. Terminal raw
    mode is always restored before this can propagate past _raw_mode()'s
    own try/finally - see that context manager's own docstring."""


def _raw_mode(fd):
    """Context manager: puts fd (normally stdin) into cbreak raw-input mode
    for the duration of the `with` block, and ALWAYS restores the original
    (cooked) terminal settings on the way out - normal return, an
    exception, a KeyboardInterrupt, all of it, via try/finally. This is the
    one place raw mode is ever entered in this file; every caller of the
    key reader below goes through here, so "terminal restore is mandatory
    on every exit path" is enforced structurally, not by convention at each
    call site. cbreak (not full raw) keeps signal generation (Ctrl-C still
    raises KeyboardInterrupt normally) and output post-processing intact -
    only input line-buffering/echo are disabled, which is all arrow-key
    reading needs."""
    import contextlib

    @contextlib.contextmanager
    def _cm():
        old = termios.tcgetattr(fd)
        try:
            tty.setcbreak(fd)
            yield
        finally:
            termios.tcsetattr(fd, termios.TCSADRAIN, old)

    return _cm()


def _read_key():
    """One logical key event from stdin, in raw mode. Returns a
    (kind, value) tuple: ("char", <1-char str>) for a printable keystroke,
    ("enter", None) for Enter/Return, ("backspace", None), ("esc", None)
    for a bare ESC (no bytes follow), or ("arrow", "up"/"down") for the
    only two arrow directions this menu structure actually uses (left/
    right are read as raw bytes but not surfaced as a distinct kind - every
    call site here is a vertical list, not a left/right layout). Reads a
    single raw byte first; if it's ESC (0x1b), does one more short,
    non-blocking-ish read to see whether a `[`-prefixed escape sequence
    follows - if nothing follows quickly, treats it as a bare ESC rather
    than blocking forever waiting for a sequence that isn't coming."""
    fd = sys.stdin.fileno()
    b = os.read(fd, 1)
    if not b:
        raise EOFError("no interactive input available")
    if b == b"\x1b":
        import select

        ready, _, _ = select.select([fd], [], [], 0.05)
        if not ready:
            return ("esc", None)
        b2 = os.read(fd, 1)
        if b2 != b"[":
            return ("esc", None)
        b3 = os.read(fd, 1)
        if b3 == b"A":
            return ("arrow", "up")
        if b3 == b"B":
            return ("arrow", "down")
        if b3 == b"C":
            return ("arrow", "down")  # right treated as "down" (advance)
        if b3 == b"D":
            return ("arrow", "up")  # left treated as "up" (back)
        return ("esc", None)
    if b in (b"\r", b"\n"):
        return ("enter", None)
    if b in (b"\x7f", b"\x08"):
        return ("backspace", None)
    try:
        return ("char", b.decode("utf-8", errors="ignore"))
    except UnicodeDecodeError:
        return ("char", "")


def _prompt_pick(prompt, choices, default=None, exit_on_invalid=False):
    """2026-08-21, PHCAL_NAV_CONSOLIDATION_001.md (PHCAL_INPUT_TREE_SURVEY_
    001.md §5 steps 1+3): replaces the old `_prompt_choice()` - the second,
    independent raw-mode key reader this file used to carry alongside
    arrow_column_pick(). Deleted outright, not deprecated in place; this is
    a thin wrapper matching its old inline-mode signature exactly (every
    former caller changed only its function name, not its own
    choices/default/exit_on_invalid arguments) built entirely on
    arrow_column_pick() - the one input engine now. `_confirm_multi_mode()`
    (the file's one former `options=` caller) calls arrow_column_pick()
    directly instead of through this wrapper, since it already builds its
    own real (key, label) options list with descriptive labels; this
    wrapper's own options are always bare key==label pairs, which is all a
    short y/n or numbered/lettered pick ever needs.

    exit_on_invalid decided by prompt SHAPE, per the survey's own rule: a
    y/n CONFIRM passes False (retry-on-typo, matching every confirm's old
    behavior); a numbered/lettered PICK passes True (eject-with-one-warning,
    matching how the group/sub-menu tree already behaved) - one rule, set
    explicitly at each of this wrapper's own callers below, not left to
    drift per call site the way the old per-call boolean did.

    IMPORTANT, found live during this pass's own dry-verify (not part of
    the original survey): every former exit_on_invalid=True caller of the
    old `_prompt_choice()` (song pick, step pick, animation-token,
    _prompt_robot, release-mode) was written assuming a valid answer
    ALWAYS eventually comes back - most of those callers use the return
    value immediately with no None-check (`int(dir_choice) - 1`,
    `_ANIMATION_TOKEN_MENU[token_choice]`, a robot-serial lookup that
    silently falls back to Brobot 1 on anything not literally "2"). Under
    the OLD engine that was safe by accident, because exit_on_invalid=True
    callers there had never actually been exercised with real garbage
    input in a way that got audited - the old inline mode's own default
    (False) covered every OTHER call site's real-world typo path. Rather
    than audit and patch every one of those call sites individually, this
    wrapper converts an exit_on_invalid=True engine None straight into a
    `_PhcalEscExit` - the same clean-abort signal ESC already raises,
    already caught at every level that matters (run_guided_flow()'s own
    try/except, _promote_tweaks()'s own). So "eject on invalid" now
    actually means "abandon this whole operation, same as ESC" tool-wide,
    not "return None and hope the caller checked" - no silent wrong-robot
    fire, no crash, one consistent meaning.

    2026-08-21, PHCAL_NAV_BASE_RULES_FIXES_001.md (PHCAL_NAV_BASE_RULES_
    SURVEY_001.md §5 step 4): every non-tree PICK screen (_prompt_robot,
    _prompt_animation_token, tempo's song/mode/step picks, sleep_wake's
    release-mode) now lists "0" as its own DELIBERATE, recognized choice -
    not just caught by the None-eject backstop above. When "0" is present
    in `choices`, its row gets a real label ("exit") instead of this
    wrapper's usual bare key==label pairing, so it reads the same
    everywhere it appears rather than showing a bare "0. 0" row. Each
    caller below still does its own `if picked == "0": raise
    _PhcalEscExit()` immediately after this call - "0" isn't special-cased
    to auto-raise HERE, because a caller may need to do its own bookkeeping
    first (none needed one this pass, but the contract stays explicit,
    not implicit)."""
    options = [("0", "exit") if c == "0" else (c, c) for c in sorted(choices)]
    highlight = sorted(choices).index(default) if default in choices else 0
    print(prompt)
    choice = arrow_column_pick(options, highlight=highlight, exit_on_invalid=exit_on_invalid)
    if choice is None:
        raise _PhcalEscExit()
    return choice


def arrow_column_pick(options, highlight=0, window_height=None, exit_on_invalid=True):
    """2026-08-19, NAV_PATTERN_SURVEY_001.md / NAV_PRIMITIVE_BUILT_001.md.
    2026-08-21, PHCAL_NAV_CONSOLIDATION_001.md (PHCAL_INPUT_TREE_SURVEY_001.md
    §5 step 1): this is now THE one input engine in this file - the
    reusable single-level arrow-column picker, with real scrolling for long
    lists. `_prompt_choice()` (the old second, independent raw-mode key
    reader) is deleted outright as of this pass; every one of its 13
    former callers now reaches this function instead, most through the
    thin `_prompt_pick()` wrapper below, `_confirm_multi_mode()` directly
    (it already built a native (key, label) options list). One raw-mode
    key-reading loop left in the whole file, not two.

    options: ordered list of (key, label) tuples. highlight: initial
    highlighted index. window_height: max visible rows at once. None (or
    >= len(options)) means "show everything, no scrolling" - the size
    phcal's own group/sub-menus actually use today (<=9 items per level,
    NAV_PATTERN_SURVEY_001.md Part 2). When len(options) > window_height, a
    scroll offset tracks which window_height-sized slice is drawn; the
    highlighted row always stays visible - the window follows it past the
    top/bottom edge instead of reprinting the full list (up to 69 items,
    the future composition editor's own line-list size) on every key
    press.

    Arrow up/down moves+wraps across the FULL option list (not just the
    visible slice), Enter selects the highlighted item, typing an in-range
    key then Enter also resolves (typed fallback). ESC raises
    _PhcalEscExit immediately, uncaught here - propagates to whatever
    already catches it (run_guided_flow()'s own try/except), unchanged.

    "0" is whatever the caller put in `options` when it's a REAL listed
    choice (navigate()'s own injected "0" = back/exit entries, or any
    PICK screen that now lists "0" explicitly per
    PHCAL_NAV_BASE_RULES_FIXES_001.md) - returned normally like any other
    valid choice, the caller decides what it means. 2026-08-21,
    PHCAL_NAV_BASE_RULES_FIXES_001.md (PHCAL_NAV_BASE_RULES_SURVEY_001.md
    §5 step 3) added ONE exception: when exit_on_invalid=False (a plain
    y/n CONFIRM) and "0" is typed but is NOT itself a real choice (true of
    every CONFIRM in this file - they're all {"y","n"}), it raises
    _PhcalEscExit the same as ESC, instead of falling into the normal
    retry-and-reprompt loop below - "0" cleanly exits from a CONFIRM
    screen too, not just a PICK, with no per-call-site change needed.

    exit_on_invalid (new this pass - the old `_prompt_choice()`'s own
    per-call-site parameter, now decided by prompt SHAPE, not drift):
    True (the default, and navigate()'s own unconditional prior behavior)
    - a typed value that isn't a real option key prints ONE
    PHCAL_GUIDED_INVALID_EXIT warning and returns None immediately. False
    - the same bad value prints a plain PHCAL_GUIDED_INVALID warning, clears
    the typed buffer, redraws, and loops - the same retry-on-typo behavior
    `_prompt_choice()`'s own exit_on_invalid=False callers (every plain
    y/n confirm) always had. PHCAL_INPUT_TREE_SURVEY_001.md §5 step 3's
    rule: every numbered/lettered PICK passes True (unchanged from
    before), every y/n CONFIRM passes False - one rule by shape, not an
    unlabeled boolean set ad hoc per call site."""
    choices = [k for k, _label in options]
    n = len(options)
    if window_height is None or window_height >= n:
        window_height = n
    offset = 0
    typed = []

    def _redraw(first=False):
        nonlocal offset
        if highlight < offset:
            offset = highlight
        elif highlight >= offset + window_height:
            offset = highlight - window_height + 1
        if not first:
            sys.stdout.write(f"\x1b[{window_height}A")
        for i in range(window_height):
            real_index = offset + i
            key, label = options[real_index]
            marker = "> " if real_index == highlight else "  "
            sys.stdout.write(f"\r\x1b[2K{marker}{key}. {label}\n")
        sys.stdout.write(f"\r\x1b[2K(typed: {''.join(typed)})")
        sys.stdout.flush()

    _redraw(first=True)

    with _raw_mode(sys.stdin.fileno()):
        while True:
            kind, val = _read_key()
            if kind == "esc":
                print()
                raise _PhcalEscExit()
            if kind == "arrow":
                if val == "up":
                    highlight = (highlight - 1) % n
                else:
                    highlight = (highlight + 1) % n
                typed = []
                _redraw()
                continue
            if kind == "backspace":
                if typed:
                    typed.pop()
                    _redraw()
                continue
            if kind == "enter":
                print()
                if typed:
                    raw = "".join(typed).strip().lower()
                else:
                    raw = choices[highlight]
                if raw in choices:
                    return raw
                if not exit_on_invalid and raw == "0":
                    # 2026-08-21, PHCAL_NAV_BASE_RULES_FIXES_001.md
                    # (PHCAL_NAV_BASE_RULES_SURVEY_001.md §5 step 3): "0" =
                    # clean exit, everywhere, one meaning - including at a
                    # plain y/n CONFIRM (exit_on_invalid=False), which
                    # otherwise never has "0" as a real choice at all and
                    # would just retry-loop on it forever below. One check,
                    # here, fixes every CONFIRM call site in the file at
                    # once - none of them needed editing individually.
                    raise _PhcalEscExit()
                if exit_on_invalid:
                    print(f"PHCAL_GUIDED_INVALID_EXIT '{raw}' is not a control on this menu - exiting")
                    return None
                print(f"PHCAL_GUIDED_INVALID expected one of {sorted(choices)}, got '{raw}'")
                typed = []
                _redraw()
                continue
            if kind == "char" and val:
                typed.append(val)
                sys.stdout.write(val)
                sys.stdout.flush()
                continue


def navigate(tree, window_height=None):
    """2026-08-19, NAV_PATTERN_SURVEY_001.md / NAV_PRIMITIVE_BUILT_001.md:
    the reusable generic N-deep tree walker, built on arrow_column_pick()
    above - this is what lets an operator "arrow anywhere," one walker,
    any depth, instead of a hand-written while/continue loop copy-pasted
    per menu level (the shape _run_guided_flow_once()'s own group/sub menu
    used before this pass's one port - see that function for the sole
    real conversion this pass makes; every other prompt in this file is
    untouched).

    `tree`: a dict mapping pick_key -> (display_label, child), where
    child is either another such dict (descend one level) or a plain
    string (a real leaf - the value navigate() returns once picked, no
    further menu drawn for it). "0" is injected automatically at every
    level - "back to previous menu" if a parent level exists on the
    stack, "exit" at the root - the caller's own tree data never needs
    its own "0" entry. Picking "0" pops one level; popping past the root
    returns None (the whole navigate() call is over - same meaning
    _run_guided_flow_once()'s own top-level "0"/exit already had before
    this port). ESC propagates uncaught, same as every other prompt in
    this file - the caller already has one catch point for it
    (run_guided_flow()'s own try/except _PhcalEscExit).

    window_height: default None, meaning "size to this level's own real
    content, every level, dynamically" - NOT a fixed number (2026-08-20,
    PHCAL_NAV_LIVETEST_FINDINGS_001.md -> PHCAL_NAV_FIXES_001.md,
    operator decision: a static default re-breaks the instant any menu's
    real item count differs from whatever number it was tuned for - it did
    exactly that here, once, already). When None, EACH level computes its
    own window height as len(options) for THAT level - the "0" this
    function always injects is already inside `options` by the time that
    length is taken, so every level's window is exactly its own real-item
    count + 1, with no menu-wide constant anywhere to drift stale as menus
    grow, shrink, or differ level to level (root vs. a 2-3-item sub-menu).
    This also matches arrow_column_pick()'s own None contract one level up
    - "show everything, no scrolling" - cascaded per level instead of
    computed once for the whole tree. A caller that DOES want real
    scrolling (the future composition editor's up-to-69-line lists,
    arrow_column_pick()'s own docstring) passes an explicit integer here,
    same as before - that value then applies uniformly to every level of
    that one navigate() call, unchanged behavior for that case."""
    stack = [tree]
    while stack:
        node = stack[-1]
        back_label = "back to previous menu" if len(stack) > 1 else "exit"
        keys = sorted(node, key=lambda k: (len(k), k))
        options = [("0", back_label)] + [(k, node[k][0]) for k in keys]
        level_height = window_height if window_height is not None else len(options)
        choice = arrow_column_pick(options, highlight=0, window_height=level_height)
        if choice is None or choice == "0":
            stack.pop()
            if not stack:
                return None
            continue
        _label, child = node[choice]
        if isinstance(child, dict):
            stack.append(child)
            continue
        return child


def _prompt_robot(prompt_note="", default=None, allow_both=False):
    """Dedicated 'which robot?' prompt - its own process, single purpose.
    2026-08-15: replaces 7 separate hand-rolled copies of
    _prompt_choice("robot (1 or 2)? ", {"1","2"}) that had drifted across
    the guided flow (animation, brobots_stay_in_place,
    brobots_session_responsiveness, move_reverse, cube, the shared
    arm/nod/rattle/danger tail, plus sleep_wake's own "1, 2, or both"
    variant) - one place now, not seven. Deliberately NOT merged with
    _prompt_animation_token below or the primitive-menu's own print/pick
    logic - decoupled, each its own process, not a shared generic "pick
    from anything" utility. prompt_note is free text appended after the
    base phrasing, e.g. "[default 2 - Brobot 2, the cube keeper]".

    2026-08-21, PHCAL_NAV_CONSOLIDATION_001.md (PHCAL_INPUT_TREE_SURVEY_
    001.md §4 finding 1 / §5 step 2): "0" used to be a redundant extra
    synonym for "both" here, colliding with "0"'s ONE tool-wide meaning
    everywhere else (back/exit in a hierarchical pick) - dropped outright.

    2026-08-21, PHCAL_NAV_BASE_RULES_FIXES_001.md (PHCAL_NAV_BASE_RULES_
    SURVEY_001.md §5 step 4): "0" is RE-ADDED, deliberately, as its own
    choice meaning clean exit (raises _PhcalEscExit).

    2026-08-21, PHCAL_LESS_FLAKY_SWEEP_001.md: allow_both's own key for
    "both robots" was the literal typed WORD "both" - the one place in
    this file where a caller had to type more than a single character to
    resolve a pick, inconsistent with every other key in the tool and the
    same "flaky, arbitrary key" complaint that got "d" removed from
    _confirm_multi_mode's dry-run row this same pass. Re-keyed to "*" -
    the same tool-wide all-flag animation's "all" tokens and
    _confirm_multi_mode's own multi-continue row already use, so "pick
    everything in this set" now means the same single character
    everywhere it comes up. Built via a direct arrow_column_pick() call
    (not the _prompt_pick() wrapper) so the printed row can say "both"
    while the typed KEY is "*" - the function's own return-value contract
    to its callers is unchanged, still the literal string "both"
    (_sleep_wake_specs()'s own `if which in ("1", "both")` checks depend on
    this exact string, untouched by this pass).

    2026-08-18 (PHCAL_DETECT_FIRST_001.md): when detect-first has set
    session_mode="single" (module-level _SESSION_MODE, set once by
    run_guided_flow() before any of this file's dispatch branches run),
    this auto-resolves to the one detected-present robot's own "1"/"2"
    identity instead of prompting at all - true single-wake, and every
    other single-robot primitive, now just work with whichever robot is
    actually plugged in (Brobot 1 OR Brobot 2, not hardcoded to either).
    MULTI mode (_SESSION_MODE="multi") and the no-detection case
    (_SESSION_MODE is None - the direct-flag main() path never calls
    run_guided_flow() and never sets this) are both completely unaffected -
    this whole block is skipped, prompting exactly as before this pass."""
    if _SESSION_MODE == "single" and _PRESENT_ROBOTS and _PRESENT_ROBOTS[0]["which"]:
        resolved = _PRESENT_ROBOTS[0]["which"]
        note = f" {prompt_note}" if prompt_note else ""
        print(f"PHCAL_SINGLE_MODE_AUTO_RESOLVE robot={resolved} ({_PRESENT_ROBOTS[0]['label']}, the only one detected present){note}")
        return resolved
    suffix = f" {prompt_note}" if prompt_note else " "
    if allow_both:
        options = [("0", "exit"), ("1", "1"), ("2", "2"), ("*", "both (1 and 2)")]
        option_keys = [k for k, _label in options]
        default_key = "*" if default == "both" else default
        highlight = option_keys.index(default_key) if default_key in option_keys else 0
        print(f"robot (1, 2, or * for both)?{suffix}")
        robot = arrow_column_pick(options, highlight=highlight)
        if robot == "0" or robot is None:
            raise _PhcalEscExit()
        if robot == "*":
            return "both"
        return robot
    choices = {"1", "2", "0"}
    robot = _prompt_pick(f"robot (1 or 2)?{suffix}", choices, default=default, exit_on_invalid=True)
    if robot == "0":
        raise _PhcalEscExit()
    return robot


def _prompt_animation_token():
    """Dedicated animation-token picker - its own process, single purpose.
    2026-08-15: pulled out of the animation primitive's own inline block so
    it's named and testable on its own - NOT merged with the primitive
    menu's own print/pick logic above (run_guided_flow's own loop over
    _PRIMITIVE_MENU) even though the shape looks similar - kept decoupled
    on purpose, not a shared generic menu helper.

    "*" = all three in sequence (2026-08-15, PHCAL_ANIMATION_ALL_SEQUENCE_
    001.md; re-keyed off "0" to "a" 2026-08-21, PHCAL_NAV_CONSOLIDATION_
    001.md - "0" is reserved tool-wide for back/exit only - then "a" to
    "*" the same day, PHCAL_NAV_BASE_RULES_FIXES_001.md: "*" is the
    tool-wide all-flag now, PHCAL_NAV_BASE_RULES_SURVEY_001.md's own
    locked rule set). "0" is also now a deliberate, recognized choice here
    meaning clean exit. Returns the literal string "all"; the caller
    (run_guided_flow's "animation" branch) is what actually loops over the
    three tokens - this function stays a pure picker, no firing logic
    here."""
    print("Animation tokens:")
    print(f"  *. all in sequence ({', '.join(_ANIMATION_TOKEN_MENU[k] for k in sorted(_ANIMATION_TOKEN_MENU))})")
    for key in sorted(_ANIMATION_TOKEN_MENU):
        print(f"  {key}. {_ANIMATION_TOKEN_MENU[key]}")
    choices = set(_ANIMATION_TOKEN_MENU) | {"*", "0"}
    token_choice = _prompt_pick(f"pick a token [*,1-{len(_ANIMATION_TOKEN_MENU)}]: ", choices, exit_on_invalid=True)
    if token_choice == "0":
        raise _PhcalEscExit()
    if token_choice == "*":
        return "all"
    return _ANIMATION_TOKEN_MENU[token_choice]


def _prompt_value(label, last_value, parse_fn, display=None):
    """One guided-flow value prompt: shows the last-used value in brackets,
    Enter keeps it, typing a new value re-parses it with `parse_fn` (the
    same validators the direct-flag form already uses) - invalid input
    re-prompts rather than exiting, since this is a live back-and-forth, not
    a one-shot argv parse. `display`, if given, overrides the default
    "[last_value]" bracket text - used for rattle's volume prompt, which
    shows the valid range ("1-5") rather than the last-used number, per the
    operator's own spec wording."""
    shown = display if display is not None else last_value
    while True:
        raw = input(f"{label} [{shown}]: ").strip()
        if raw == "":
            return last_value
        try:
            return parse_fn(raw)
        except ValueError as exc:
            print(f"PHCAL_GUIDED_INVALID {exc}")


# 2026-07-22, Part A: guided flow's primitive pick is now a numbered menu
# (1/2/3), not a typed word - Part C adds rattle as the 3rd choice. Option 4
# (weather, 2026-08-02) is an API-only test, not a mechanical primitive - it
# never reaches the robot/hold/speed machinery below at all, see the early
# branch in run_guided_flow(). Rung 4 (2026-08-08,
# GOLDEN_BROBOTS_CONTROL_CATALOG_SURVEY_001.md) widens this from 4 to 7:
# sleep/wake/brobots_ready are the catalog's other standalone-fireable golden
# mechanisms (direct-SDK binaries, not song-embedded notes) - this dict is
# now the ONE source both the rendered menu and the dispatch below read from,
# so adding a future mechanism is one new entry here, not a reformatted
# prompt string. Deliberately NOT listed: the catalog's song-only note types
# (say_turn/emotion_beat/animation/pause/connect/exit/wake_both) - they have
# no standalone bench-fire path (see GOLDEN_BROBOTS_CONTROL_CATALOG_SURVEY_
# 001.md's own ASSIGNABILITY/Section-A-vs-fireable distinction) and phcal
# stays a bench tool, not a song editor, this pass.
# RENUMBERED 2026-08-10, per operator direction: wheel_nudge removed
# outright (it was a redundant third path to the same reversal move_reverse
# already reaches via its own chain-wake toggle - composing brobots_wake +
# move_reverse as one undifferentiated call added a menu entry and a
# maintained code path without adding any capability). 1-7 are the
# single-robot, single-purpose "core" tier. 8-11 were parked as "to sort
# out better" earlier the same day, then put back in - renamed for clarity
# per operator direction, same session: hold -> brobots_stay_in_place,
# sleep_wake -> brobots_sleep_to_wake_direct_sdk, brobots_wake ->
# brobots_session_responsiveness, brobots_ready -> brobots_announce_in_sync.
# Pure rename at the dispatch-matching layer only - every cmd_* function
# these map to (cmd_hold/cmd_sleep_wake_*/cmd_brobots_wake/cmd_brobots_ready)
# is untouched, same mechanism, reused verbatim per operator direction
# ("try to reuse existing working code").
_PRIMITIVE_MENU = {
    "1": "arm",
    "2": "nod",
    "3": "rattle",
    "4": "weather",
    "5": "animation",
    "6": "move_reverse",
    # 2026-08-10 (BROBOT_2_INSTABILITY_EXTERNAL_AI_BRIEF_001.md): read-only,
    # no BehaviorControl, no movement - a diagnostic snapshot, not a motion
    # primitive. which is always "both" now (prompt removed per operator
    # direction, 2026-08-10).
    "7": "robot_info",
    # 8-11 dispatch to cmd_hold/cmd_sleep_wake_set_time+on_process/
    # cmd_brobots_wake/cmd_brobots_ready respectively - see each dispatch
    # branch below for the mapping from this renamed primitive string to
    # the unchanged underlying function.
    "8": "brobots_stay_in_place",
    "9": "brobots_sleep_to_wake_direct_sdk",
    "10": "brobots_session_responsiveness",
    "11": "brobots_announce_in_sync",
    # 2026-08-11: hands off to tempo_set_001.py, the same tool brobots.sh's
    # standalone `tempo-set` alias already wraps - see TEMPO_SET_PATH above.
    # Song-scoped, not robot-scoped - no primitive/robot preflight below.
    "12": "tempo",
    # 2026-08-12: same direct-SDK mechanism as rattle (3), different WAV
    # asset (GOPOD's own playSound sound) - see DANGER_WAV_PATH's own
    # comment above.
    "13": "danger",
    # 2026-08-15: cube-blip, live-fire-confirmed 2026-08-14 against Brobot 2
    # (connect -> all corners red -> hold -> all corners green -> release).
    # No volume/hold/reps to configure - just robot -> fire, own self-
    # contained branch below (mirrors brobots_stay_in_place/move_reverse's
    # own restart_preflight -> preflight -> fire -> return shape), not the
    # shared arm/nod/rattle/danger volume-prompt tail.
    "14": "cube",
}
_MOVE_REVERSE_MENU_SUFFIX = " (ON-CHARGER reverse pulse)"

# 2026-08-09, SESSION-RESPONSIVENESS CHAIN TOGGLE (operator direction):
# brobots_session_responsiveness (renamed 2026-08-10, was brobots_wake) can
# front-run any primitive that shares its own Wire-Pod REST assume/release
# mechanism - arm/nod/brobots_stay_in_place/animation/move_reverse. NOT
# rattle/brobots_sleep_to_wake_direct_sdk/brobots_announce_in_sync
# (direct-SDK gRPC binaries - a whole different control channel; chaining a
# Wire-Pod assume+release right before one of those grabs its own session
# reintroduces the exact release-then-regrant race this codebase already
# hit and fixed once, see cmd_rattle's own release-before-connect shape
# above) and NOT weather (no robot control at all).
#
# Defaults: on for move_reverse (proven needed - WHEEL_NUDGE_WAKE_SETTLE_
# FIX_001.md's 3 rounds), off for arm/nod/brobots_stay_in_place/animation
# (already live-confirmed reliable without it - this isn't a fix for a
# proven problem there, just an available option).
_BROBOTS_WAKE_CHAIN_ELIGIBLE = {"arm", "nod", "brobots_stay_in_place", "animation", "move_reverse"}
_BROBOTS_WAKE_CHAIN_DEFAULT = {"move_reverse": "y"}  # everything else in the eligible set defaults "n"

# 2026-08-16 (PHCAL_MENU_REGROUP_SURVEY_001.md -> PHCAL_MENU_REGROUP_BUILT):
# pure-display grouping of the same 14 _PRIMITIVE_MENU identity strings into
# the operator's 8-group menu sketch - every value below is one of
# _PRIMITIVE_MENU's own strings, unchanged, unrenamed. A group with exactly
# one member routes straight through, same as the flat menu always did
# (info/cube/animations/tempo); a group with 2-3 members gets one new
# sub-menu prompt (movements/audio/brobots 1/brobots 2) before falling into
# the SAME `if primitive == "..."` dispatch chain below - no dispatch code
# changed, only how `primitive` gets its value. Confirmed by hand: every
# _PRIMITIVE_MENU value appears in exactly one group here, all 14 placed.
_PRIMITIVE_GROUPS = {
    "1": ("info (active brobots)", ["robot_info"]),
    "2": ("movements (arm/nod/reverse)", ["arm", "nod", "move_reverse"]),
    "3": ("vector's cube (colour control)", ["cube"]),
    "4": ("audio (rattle/danger)", ["rattle", "danger"]),
    "5": ("animations (kgsuccess/searching/answering)", ["animation"]),
    "6": ("brobots 1 (weather/announce/stay)", ["weather", "brobots_announce_in_sync", "brobots_stay_in_place"]),
    "7": ("brobots 2 (wake/responsiveness)", ["brobots_sleep_to_wake_direct_sdk", "brobots_session_responsiveness"]),
    "8": ("tempo (pacing)", ["tempo"]),
}


# 2026-08-20, PHCAL_NAV_FIXES_001.md (finding 2, PHCAL_NAV_LIVETEST_
# FINDINGS_001.md): single source of truth for the robot-vs-non-robot
# classification. Every _PRIMITIVE_MENU identity string NOT in this set
# needs a live robot to do anything; tempo is the sole confirmed exception
# - cmd_tempo_calibration()'s own docstring: "No robot, no hardware, no
# preflight - a song's knobs.json is the only thing this touches." weather
# is deliberately NOT in this set - it prompts for a robot and speaks
# through it (run_single_note("weather", ...)), despite older comments
# nearby once claiming otherwise; those are corrected alongside this fix,
# not left to contradict this set.
_NON_ROBOT_PRIMITIVES = {"tempo"}


def _none_mode_disabled_marker(primitive):
    """2026-08-20, PHCAL_NAV_FIXES_001.md (finding 2's proposed approach,
    PHCAL_NAV_LIVETEST_FINDINGS_001.md): the none-mode visible-disable cue,
    appended to a control's own menu label so an operator sees, before
    picking, which options can't do anything with no robot detected.
    Deliberately just a label string, not a block on the pick itself -
    arrow_column_pick()/navigate() stay robot-state-agnostic on purpose (a
    layering call: those are meant to be reused by non-phcal callers later,
    e.g. the future composition editor, NAV_PRIMITIVE_BUILT_001.md) - a
    disabled option is still selectable here; each dispatch branch's own
    none-mode guard (robot_info/brobots_announce_in_sync already had one;
    the rest gain the same pattern this same pass) is what actually stops a
    pick from attempting a doomed connection, not this function. Empty
    string outside none mode or for a non-robot primitive, so callers can
    always just append the result with no conditional at the call site."""
    if _SESSION_MODE == "none" and primitive not in _NON_ROBOT_PRIMITIVES:
        return " [disabled - no robot detected]"
    return ""


def _submenu_control_note(primitive):
    """Per-control note printed next to a control inside one of the 4 new
    multi-member group sub-menus (movements/audio/brobots 1/brobots 2) -
    carries forward the exact brobots_session_responsiveness chain-
    eligibility fact the old flat main-menu's items 2/3 used to spell out
    as two long sentences, now attached to the control's own line instead
    so the main menu can stay short (operator direction, PHCAL_MENU_
    REGROUP_BUILT). Derived from the same _BROBOTS_WAKE_CHAIN_ELIGIBLE set
    everywhere else already uses - PHCAL_CLEAN_EXIT_AND_NOTES_001.md's own
    self-updating convention, kept, so a future primitive still can't drift
    this stale. Only the genuine special cases are hardcoded:
    brobots_session_responsiveness IS the toggle, not a target of it;
    move_reverse keeps its own ON-CHARGER safety note. (2026-08-20,
    PHCAL_NAV_FIXES_001.md: weather used to be hardcoded here too, as "no
    robot control at all" - wrong, corrected; weather is robot-control, see
    _NON_ROBOT_PRIMITIVES above, and now falls through to the same
    chain-eligibility branches as any other robot-control primitive.) Only
    called for the 10 primitives that land in a multi-member group - the 4
    single-member direct groups (info/cube/animations/tempo) never reach
    this, so they get the same none-mode marker appended separately in
    _build_primitive_group_tree() below."""
    if primitive == "brobots_session_responsiveness":
        return " (the chain toggle itself - global wake note, no movement)" + _none_mode_disabled_marker(primitive)
    if primitive == "weather":
        return " (robot-control - speaks the forecast through a robot; not chain-eligible)" + _none_mode_disabled_marker(primitive)
    if primitive == "move_reverse":
        return _MOVE_REVERSE_MENU_SUFFIX + ", brobots_session_responsiveness chain-eligible" + _none_mode_disabled_marker(primitive)
    if primitive in _BROBOTS_WAKE_CHAIN_ELIGIBLE:
        return " (brobots_session_responsiveness chain-eligible)" + _none_mode_disabled_marker(primitive)
    return " (not chain-eligible - different control channel)" + _none_mode_disabled_marker(primitive)


def _build_primitive_group_tree():
    """2026-08-19, NAV_PRIMITIVE_BUILT_001.md: builds the navigate()-shaped
    tree from _PRIMITIVE_GROUPS - the ONE real navigate() call site this
    pass creates (_run_guided_flow_once()'s group/sub menu; every other
    prompt in this file is unchanged). A 1-member group's own tree value
    IS the primitive name directly (a leaf, no sub-menu ever drawn) -
    matching the original hand-rolled loop's own "1-member group routes
    straight through" behavior exactly. A multi-member group's tree value
    is a nested dict of numbered sub_keys -> (member label + its
    _submenu_control_note, member name) - same sub-menu content/ordering
    _PRIMITIVE_GROUPS + _submenu_control_note already produced before this
    port, just shaped for navigate() instead of built inline by hand each
    call."""
    tree = {}
    for group_key in sorted(_PRIMITIVE_GROUPS, key=int):
        group_label, members = _PRIMITIVE_GROUPS[group_key]
        if len(members) == 1:
            # 2026-08-20, PHCAL_NAV_FIXES_001.md: these 4 groups
            # (info/cube/animations/tempo) never reach _submenu_control_note()
            # (that's only called for multi-member groups, below) - the
            # none-mode marker gets appended here instead, same source
            # (_none_mode_disabled_marker) either way, so tempo stays
            # unmarked and the other 3 (all robot-control) get flagged too.
            tree[group_key] = (group_label + _none_mode_disabled_marker(members[0]), members[0])
        else:
            sub = {}
            for i, member in enumerate(members):
                sub_key = str(i + 1)
                sub[sub_key] = (f"{member}{_submenu_control_note(member)}", member)
            tree[group_key] = (group_label, sub)
    return tree


def _prompt_brobots_wake_chain(primitive):
    # Fallback flipped n->y 2026-08-15 (PHCAL_NOTES_WEATHER_WAKE_FIXES_001.md,
    # operator direction) - _BROBOTS_WAKE_CHAIN_DEFAULT's per-primitive "off
    # for arm/nod/brobots_stay_in_place/animation" reasoning above is now
    # superseded for the fallback case; move_reverse's own explicit "y" entry
    # is unaffected (already matched the new default, kept for clarity, not
    # functional anymore).
    default = _BROBOTS_WAKE_CHAIN_DEFAULT.get(primitive, "y")
    choice = _prompt_pick(
        f"run brobots_wake first? y/n [default {default}]: ", {"y", "n"}, default=default
    )
    return choice == "y"


def _run_brobots_wake_no_release(mod, clock, serial, live):
    """The chain toggle's own call: wake WITHOUT releasing, so the
    primitive that runs next picks up the SAME held-control window -
    reuses the exact same core functions cmd_brobots_wake composes
    (conn_test/assume/golden-flag release-pulse), just skips that command's
    own tail release. Mechanism REPLACED 2026-08-10, see
    WHEEL_NUDGE_WAKE_RELEASE_SETTLE_SECONDS's own comment above."""
    wake_check = run_wheel_nudge_wake_check(mod, clock, serial, live)
    assume = run_assume_control(mod, clock, serial, live)
    wake_confirm = run_wheel_nudge_wake_release_pulse(mod, clock, serial, live)
    ok = wake_check["ok"] and assume["ok"] and wake_confirm["ok"]
    return {"ok": ok, "wake_check": wake_check, "assume": assume, "wake_confirm": wake_confirm}


def _resolve_chain_wake(primitive, mod, clock, serial, live):
    """2026-08-21, PHCAL_NAV_CONSOLIDATION_001.md (PHCAL_INPUT_TREE_SURVEY_
    001.md §3 DUPLICATES / §5 step 5): the chain-wake prompt-then-fire
    sequence, pulled out of 4 near-identical inline copies inside
    _run_guided_flow_once() (animation, brobots_stay_in_place, move_reverse,
    the shared arm/nod tail) into one place - same prompt
    (_prompt_brobots_wake_chain), same no-release wake call
    (_run_brobots_wake_no_release), same failure message shape, just
    primitive-parameterized instead of copy-pasted per call site.

    Returns (chain_wake, failed). chain_wake is exactly what the caller
    should pass as its own cmd_*() call's pre_assumed= argument - False
    when the operator declined the chain, or when it was accepted but
    failed (failed=True covers that distinction instead). failed=True
    means the wake was accepted but did not actually succeed - the
    caller's own contract, unchanged from before this refactor, is to
    print nothing further and return 1 immediately in that case (the
    PHCAL_BROBOTS_WAKE_CHAIN_FAILED line below already covers the
    explanation)."""
    chain_wake = _prompt_brobots_wake_chain(primitive)
    if not chain_wake:
        return False, False
    wake = _run_brobots_wake_no_release(mod, clock, serial, live)
    if not wake["ok"]:
        print(f"{clock.prefix()}PHCAL_BROBOTS_WAKE_CHAIN_FAILED - stopping, not firing {primitive} without a confirmed wake")
        return False, True
    return True, False


def _run_guided_flow_once():
    """2026-08-18, PHCAL_NAV_POLISH_001.md: renamed from run_guided_flow()
    (now a thin stay-in-loop wrapper below, the real public entry point) -
    internally unchanged except the group/sub-menu preamble is now its own
    small while/break/continue loop (see its own comment above) and the
    two "operator explicitly chose to leave" return sites return None
    instead of 0, so the wrapper can tell that apart from "a primitive
    dispatch returned rc=0 (success)." Every other one of this function's
    ~15 dispatch-branch returns is untouched - same value, same meaning as
    before this pass.

    Rung 3's guided prompt flow, entered when `phcal` is called with no
    primitive at all. 2026-08-16 (PHCAL_MENU_REGROUP_BUILT): a two-level
    menu now picks `primitive` - a top-level group (8 human-labeled
    categories, _PRIMITIVE_GROUPS) then, for any group with 2-3 members, a
    sub-menu showing the real code identifiers (arm/nod/move_reverse under
    "movements", etc.). A 1-member group (info/cube/animations/tempo)
    routes straight through, same as the old flat menu always did for
    those. Pure display - every one of the 14 identity strings is
    unchanged, so everything below this point (robot -> reps/hold between
    reps/speed (arm, nod), volume (rattle) - each pre-filled from
    phcal_last.json, seeded from the rung-1 bingo defaults on a first run
    with no file yet -> fires exactly like the direct-flag form (live if
    the gate's on, dry if it's off) -> saves the values actually used back
    to the same one memory file) is untouched by this regroup. Weather has
    its own shape since 2026-08-15 (PHCAL_WEATHER_ROBOT_SPEAK_001.md):
    robot pick, then straight into run_robot_control_song_001.py's own
    run_single_note("weather", ...) - the same real fetch + per-robot
    unit/clock speak the awaken song's own weather step and the
    gopod-weather-say alias already use. No reps/hold/speed/volume value to
    remember (nothing about the forecast is tunable), so it never touches
    phcal_last.json.

    2026-08-18 (PHCAL_DETECT_FIRST_001.md): the detect-first probe
    (replaces the old optional "1. info" menu pick) sets this session's
    mode (none/single/multi, module-level _SESSION_MODE/_PRESENT_ROBOTS,
    read by _prompt_robot() and the robot_info/brobots_announce_in_sync
    branches below) and gates on any present robot reading low battery
    (default: do not proceed). The menu itself stays the exact same full
    8-group menu regardless of mode - nothing removed, behavior adapts
    underneath it.

    2026-08-19 refinement (PHCAL_NAV_POLISH_001.md addendum): that probe no
    longer runs HERE - it moved to _resolve_session_mode_once(), called
    exactly ONCE by run_guided_flow() before its stay-in loop starts, so a
    "continue" at the continue-or-exit prompt no longer re-triggers the
    ~1-2s probe or a repeated multi-mode confirm question. This function
    now assumes _SESSION_MODE/_PRESENT_ROBOTS are already resolved by the
    time it's called - it only reads them, same as _prompt_robot() and the
    dispatch branches below already did."""
    global _SESSION_MODE, _PRESENT_ROBOTS
    # 2026-08-19, NAV_PRIMITIVE_BUILT_001.md: ported onto the new generic
    # navigate() tree-walker (built standalone this pass,
    # NAV_PATTERN_SURVEY_001.md's design, alongside _prompt_choice() -
    # neither that function nor any of its OTHER callers were touched).
    # Was a hand-written two-level while/break/continue loop (see git
    # history / PHCAL_NAV_POLISH_001.md for that version); same content,
    # same effective choices, only the picking MECHANISM changed - a
    # generic walker instead of a loop written once for exactly two
    # levels. _build_primitive_group_tree() below reproduces
    # _PRIMITIVE_GROUPS's own group/sub-menu shape exactly, including the
    # 1-member-group-routes-straight-through behavior (no sub-menu ever
    # drawn for those). "0" backing up one level, "0" at the root exiting,
    # wrap, and warn-once-then-exit on genuinely invalid input are all
    # navigate()/arrow_column_pick()'s own built-in contract now, not
    # hand-rolled here - operator decision 1/4 from the original nav-polish
    # pass, still enforced, just by the shared primitive instead of local
    # bookkeeping. One deliberate simplification, noted explicitly: the
    # old loop reprinted "PLAYHEAD Calibrations (phcal):" + the notes lines
    # every time "0" returned to the group menu; navigate() redraws
    # in-place (the same \x1b[...A cursor-up technique the old sub-menu
    # already used), so those header lines now print once, not on every
    # back-up - a real difference in printed OUTPUT volume, not in what
    # gets picked or how "0"/exit/wrap behave.
    print("PLAYHEAD Calibrations (phcal):")
    print("Please note:")
    print(
        "  see SLEEP_BENCH_ALIASES_001.md & SLEEP_SEGMENT_ALIASES_001.md - "
        "not fireable from here"
    )
    print("** arrows to move, Enter to select, ESC to exit, 0 to go back **")
    primitive = navigate(_build_primitive_group_tree())
    if primitive is None:
        print("PHCAL_GUIDED_EXIT exiting")
        return None

    if primitive == "weather":
        # 2026-08-15 (PHCAL_WEATHER_ROBOT_SPEAK_001.md, widened same day in
        # PHCAL_NOTES_WEATHER_WAKE_FIXES_001.md): robot pick + reuse of
        # run_robot_control_song_001.py's proven run_single_note("weather",
        # ...) is unchanged from the first pass. The location prompt is
        # back, but bare Enter does NOT re-query "windsor ontario canada" -
        # that exact free-text phrase was the original geocode bug
        # (cmd_weather_test's own docstring: it used to return zero
        # matches). Bare Enter passes location=None, which
        # fetch_windsor_weather() resolves to its own proven LOCATION
        # constant ("Windsor,ON,CA") - the displayed default describes what
        # that means in plain words, it isn't the literal query string. A
        # typed location DOES take effect (single direct geocode query, no
        # candidate-fallback retry - that logic lives in phcal's own
        # now-unreachable cmd_weather_test/_weather_geocode_candidates, not
        # ported here to avoid reimplementing fetch_windsor_weather()).
        robot = _prompt_robot()
        location = input("weather test location [default windsor ontario canada]: ").strip()
        live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
        control_mod = _load_module(CONTROL_SONG_RUNNER_PATH, "run_robot_control_song_001")
        result = control_mod.run_single_note("weather", live, robot, location=location or None)
        return 0 if result.get("ok") else 1

    if primitive == "tempo":
        return cmd_tempo_calibration()

    if primitive == "brobots_announce_in_sync":
        phrase = _prompt_value("phrase", "Brobots ready!", lambda raw: raw)
        live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
        os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
        clock = _Clock()
        # 2026-08-18, PHCAL_DETECT_FIRST_001.md: single-mode degrades to
        # cmd_brobots_ready_single (one robot, same phrase, via the
        # Wire-Pod REST say channel - the direct-SDK binary can't run with
        # one robot, see that function's own comment). NONE mode has
        # nothing present to speak the phrase - skip cleanly rather than
        # attempt a call that can only fail. MULTI mode (or no detection,
        # _SESSION_MODE is None on the direct-flag path) is completely
        # unchanged - the exact same cmd_brobots_ready call as before this
        # pass.
        if _SESSION_MODE == "single" and _PRESENT_ROBOTS and _PRESENT_ROBOTS[0]["which"]:
            present = _PRESENT_ROBOTS[0]
            rc = cmd_brobots_ready_single(clock, present["which"], present["label"], live, phrase)
        elif _SESSION_MODE == "none":
            print(f"{clock.prefix()}PHCAL_BROBOTS_READY_SKIPPED no robots detected present this session")
            rc = 1
        else:
            mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
            rc = cmd_brobots_ready(mod, clock, live, phrase)
        _save_last("brobots_announce_in_sync", {"phrase": phrase})
        print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=brobots_announce_in_sync path={LAST_PATH}")
        return rc


    if primitive == "animation":
        animation_token = _prompt_animation_token()
        robot = _prompt_robot()
        hold_seconds = _prompt_value(
            "hold", ANIMATION_DEFAULT_HOLD_SECONDS, lambda raw: _parse_float_flag("hold", raw)
        )
        live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
        os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
        mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
        serial = BROBOT_2_SERIAL if robot == "2" else BROBOT_1_SERIAL
        clock = _Clock()
        restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
        if not restart_preflight["ok"]:
            return 1
        preflight = run_preflight(mod, clock, serial, live)
        if not preflight["ok"]:
            return 1
        battery = run_battery_check(mod, clock, serial, live)
        if not battery["ok"]:
            print(f"{clock.prefix()}PHCAL_BLOCKED_LOW_BATTERY not firing animation - see PHCAL_BATTERY_CHECK line above")
            return 1
        chain_wake, chain_wake_failed = _resolve_chain_wake("animation", mod, clock, serial, live)
        if chain_wake_failed:
            return 1
        if animation_token == "all":
            # 2026-08-15 (PHCAL_ANIMATION_ALL_SEQUENCE_001.md): fires the
            # existing single-token cmd_animation() three times, in order -
            # not a second implementation. kgSuccess/searching have a
            # confirmed getout (or fire-once-then-hold); answering has none
            # (_ANIMATION_GETOUT_TOKENS) and lingers after release, so it
            # fires last on purpose - its own tail never bleeds into the
            # other two. cmd_animation() always releases control at the end
            # of every call regardless of pre_assumed (see its own `finally`
            # block) - so pre_assumed only applies to the FIRST token here;
            # tokens 2/3 do their own normal assume, matching every other
            # multi-call sequence in this file.
            fire_order = ["kgSuccess", "searching", "answering"]
            ok = True
            for i, token in enumerate(fire_order):
                rc = cmd_animation(
                    mod, clock, serial, robot, live, token, hold_seconds,
                    pre_assumed=(chain_wake if i == 0 else False),
                )
                ok = ok and rc == 0
                # Settle pause BETWEEN tokens only - not after the last one
                # (nothing follows answering to settle for).
                if i < len(fire_order) - 1:
                    if live:
                        print(f"{clock.prefix()}PHCAL_ANIMATION_SEQUENCE_PAUSE settling {_ANIMATION_SEQUENCE_PAUSE_SECONDS}s before next token")
                        time.sleep(_ANIMATION_SEQUENCE_PAUSE_SECONDS)
                    else:
                        print(f"{clock.prefix()}PHCAL DRY: would pause {_ANIMATION_SEQUENCE_PAUSE_SECONDS}s before next token")
            _save_last("animation", {"hold": hold_seconds})
            print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=animation path={LAST_PATH}")
            return 0 if ok else 1
        rc = cmd_animation(mod, clock, serial, robot, live, animation_token, hold_seconds, pre_assumed=chain_wake)
        _save_last("animation", {"hold": hold_seconds})
        print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=animation path={LAST_PATH}")
        return rc

    if primitive == "brobots_stay_in_place":
        robot = _prompt_robot()
        hold_seconds = _prompt_value(
            "hold", HOLD_DEFAULT_SECONDS, lambda raw: _parse_float_flag("hold", raw)
        )
        live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
        os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
        mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
        serial = BROBOT_2_SERIAL if robot == "2" else BROBOT_1_SERIAL
        clock = _Clock()
        restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
        if not restart_preflight["ok"]:
            return 1
        preflight = run_preflight(mod, clock, serial, live)
        if not preflight["ok"]:
            return 1
        chain_wake, chain_wake_failed = _resolve_chain_wake("brobots_stay_in_place", mod, clock, serial, live)
        if chain_wake_failed:
            return 1
        rc = cmd_hold(mod, clock, serial, robot, live, hold_seconds, pre_assumed=chain_wake)
        _save_last("brobots_stay_in_place", {"hold": hold_seconds})
        print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=brobots_stay_in_place path={LAST_PATH}")
        return rc

    if primitive == "brobots_session_responsiveness":
        robot = _prompt_robot()
        live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
        os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
        mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
        serial = BROBOT_2_SERIAL if robot == "2" else BROBOT_1_SERIAL
        clock = _Clock()
        restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
        if not restart_preflight["ok"]:
            return 1
        preflight = run_preflight(mod, clock, serial, live)
        if not preflight["ok"]:
            return 1
        return cmd_brobots_wake(mod, clock, serial, robot, live)

    if primitive == "move_reverse":
        robot = _prompt_robot()
        hold_seconds = _prompt_value(
            "hold",
            MOVE_REVERSE_DEFAULT_HOLD_SECONDS,
            lambda raw: _parse_float_flag("hold", raw, max_value=MOVE_REVERSE_MAX_HOLD_SECONDS),
        )
        live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
        os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
        mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
        serial = BROBOT_2_SERIAL if robot == "2" else BROBOT_1_SERIAL
        clock = _Clock()
        restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
        if not restart_preflight["ok"]:
            return 1
        preflight = run_preflight(mod, clock, serial, live)
        if not preflight["ok"]:
            return 1
        battery = run_battery_check(mod, clock, serial, live)
        if not battery["ok"]:
            print(f"{clock.prefix()}PHCAL_BLOCKED_LOW_BATTERY not firing move_reverse - see PHCAL_BATTERY_CHECK line above")
            return 1
        chain_wake, chain_wake_failed = _resolve_chain_wake("move_reverse", mod, clock, serial, live)
        if chain_wake_failed:
            return 1
        rc = cmd_move_reverse(mod, clock, serial, robot, live, hold_seconds, pre_assumed=chain_wake)
        _save_last("move_reverse", {"hold": hold_seconds})
        print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=move_reverse path={LAST_PATH}")
        return rc

    if primitive == "robot_info":
        # 2026-08-18, PHCAL_DETECT_FIRST_001.md: mode-aware `which` -
        # reports whichever robot(s) detect-first actually found present,
        # not always the old hardcoded "both". NONE mode has nothing to
        # report on - skip cleanly. MULTI mode (or no detection at all,
        # _SESSION_MODE is None on the direct-flag path) keeps the exact
        # same which="both" this branch always used.
        if _SESSION_MODE == "single" and _PRESENT_ROBOTS and _PRESENT_ROBOTS[0]["which"]:
            which = _PRESENT_ROBOTS[0]["which"]
        elif _SESSION_MODE == "none":
            print("PHCAL_ROBOT_INFO_SKIPPED no robots detected present this session")
            return 1
        else:
            which = "both"
        live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
        if not live:
            print("PHCAL_NO_DRY_MODE robot_info has no simulate path (no --dry exists in the binary) - re-run with GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 to fire it")
            return 1
        os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
        mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
        clock = _Clock()
        return cmd_robot_info(mod, clock, which)

    if primitive == "cube":
        # 2026-08-15: nothing to configure (no volume/hold/reps - the
        # binary's own blipHoldDuration is fixed) - robot pick, defaulting
        # to Brobot 2 (2), the cube keeper, then the same restart_preflight ->
        # preflight -> fire -> return shape brobots_stay_in_place/
        # move_reverse above already use. No dry mode - same "this class of
        # tool always fires for real" precedent robot-sleep/-wake/-info and
        # cmd_cube's own CLI form already follow.
        robot = _prompt_robot("[default 2 - Brobot 2, the cube keeper]", default="2")
        live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
        os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
        mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
        serial = BROBOT_2_SERIAL if robot == "2" else BROBOT_1_SERIAL
        clock = _Clock()
        restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
        if not restart_preflight["ok"]:
            return 1
        preflight = run_preflight(mod, clock, serial, live)
        if not preflight["ok"]:
            return 1
        return cmd_cube(mod, clock, serial, robot, live)

    if primitive == "brobots_sleep_to_wake_direct_sdk":
        print("Release mode:")
        print("  1. after a set time")
        print("  2. signaled by a completed process (a real wpr/restart-wirepod check)")
        mode_choice = _prompt_pick("pick a release mode [1-2]: ", {"1", "2", "0"}, exit_on_invalid=True)
        if mode_choice == "0":
            raise _PhcalEscExit()
        which = _prompt_robot(default="both", allow_both=True)
        wait_seconds = None
        if mode_choice == "1":
            # Same floor as legacy --hold mode (_parse_sleep_hold_flag /
            # SLEEP_MIN_HOLD_SECONDS) - the signal file gets touched
            # `wait_seconds` after the sleep subprocess launches, running
            # concurrently with GoToSleepGetIn/getInSettleSeconds(2.7s)/
            # GoToSleepSleeping, not after confirming the robot actually
            # settled. A too-short wait risks the exact same
            # released-before-settled bug this floor was already built to
            # prevent, just in the --wait-signal path instead of --hold.
            wait_seconds = _prompt_value("wait", SLEEP_WAKE_DEFAULT_WAIT_SECONDS, _parse_sleep_hold_flag)
        live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
        if not live:
            print("PHCAL_NO_DRY_MODE sleep_wake has no simulate path (no --dry exists in the binary) - re-run with GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 to fire it")
            return 1
        os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
        mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
        clock = _Clock()
        if mode_choice == "1":
            rc = cmd_sleep_wake_set_time(mod, clock, which, wait_seconds)
            _save_last("brobots_sleep_to_wake_direct_sdk", {"wait": wait_seconds})
            print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive=brobots_sleep_to_wake_direct_sdk path={LAST_PATH}")
            return rc
        return cmd_sleep_wake_on_process(mod, clock, live, which)

    robot = _prompt_robot()

    last = _load_last()[primitive]

    count = None
    cycles = None
    hold_seconds = None
    speed = None
    volume_ui = None

    if primitive == "arm":
        cycles = _prompt_value("reps", last["cycles"], lambda raw: _parse_int_flag("reps", raw, min_value=1))
        hold_seconds = _prompt_value(
            "hold between reps", last["hold"], lambda raw: _parse_float_flag("hold between reps", raw)
        )
        speed = _prompt_value("speed", last["speed"], lambda raw: _parse_int_flag("speed", raw, min_value=1))
    elif primitive == "nod":
        count = _prompt_value("reps", last["count"], lambda raw: _parse_int_flag("reps", raw, min_value=1))
        hold_seconds = _prompt_value(
            "hold between reps", last["hold"], lambda raw: _parse_float_flag("hold between reps", raw)
        )
        speed = _prompt_value("speed", last["speed"], lambda raw: _parse_int_flag("speed", raw, min_value=1))
    else:  # rattle / danger - both volume-only, same 1-5 scale
        volume_ui = _prompt_value(
            "volume",
            last["volume"],
            lambda raw: _parse_int_flag("volume", raw, min_value=1, max_value=5),
            display="1-5",
        )

    live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
    # See PHCAL_RESTART_WIREPOD_ENV's own comment block above: setdefault, not
    # a hard overwrite, and must happen before _load_module (module-level
    # constant read at import time).
    os.environ.setdefault(PHCAL_RESTART_WIREPOD_ENV, "1")
    mod = _load_module(RUNNER_PATH, "run_section1_full_live_001")
    serial = BROBOT_2_SERIAL if robot == "2" else BROBOT_1_SERIAL
    clock = _Clock()

    print(f"{clock.prefix()}PHCAL_ISOLATE primitive={primitive} robot={robot} serial={serial} live={live}")
    if not live:
        print(f"{clock.prefix()}DRY: live robot gate is off (GOPOD_ALLOW_LIVE_ROBOT_SPEECH != 1)")

    restart_preflight = run_restart_wirepod_preflight(mod, clock, live)
    if not restart_preflight["ok"]:
        return 1

    preflight = run_preflight(mod, clock, serial, live)
    if not preflight["ok"]:
        return 1

    if primitive in ("arm", "nod"):
        battery = run_battery_check(mod, clock, serial, live)
        if not battery["ok"]:
            print(f"{clock.prefix()}PHCAL_BLOCKED_LOW_BATTERY not firing {primitive} - see PHCAL_BATTERY_CHECK line above")
            return 1
        chain_wake, chain_wake_failed = _resolve_chain_wake(primitive, mod, clock, serial, live)
        if chain_wake_failed:
            return 1
    else:
        chain_wake = False  # rattle/danger - not eligible, see _BROBOTS_WAKE_CHAIN_ELIGIBLE's own comment

    if primitive == "arm":
        rc = cmd_arm(mod, clock, serial, robot, live, cycles, hold_seconds, speed, pre_assumed=chain_wake)
        _save_last("arm", {"cycles": cycles, "hold": hold_seconds, "speed": speed})
    elif primitive == "nod":
        rc = cmd_nod(mod, clock, serial, robot, live, count, hold_seconds, speed, pre_assumed=chain_wake)
        _save_last("nod", {"count": count, "hold": hold_seconds, "speed": speed})
    elif primitive == "rattle":
        rc = cmd_rattle(mod, clock, serial, robot, live, volume_ui)
        _save_last("rattle", {"volume": volume_ui})
    else:  # danger
        rc = cmd_danger(mod, clock, serial, robot, live, volume_ui)
        _save_last("danger", {"volume": volume_ui})

    print(f"{clock.prefix()}PHCAL_LAST_SAVED primitive={primitive} path={LAST_PATH}")
    return rc


def _prompt_continue_or_exit():
    """2026-08-18, PHCAL_NAV_POLISH_001.md, operator decision 5: fires after
    EVERY _run_guided_flow_once() pass that wasn't an explicit exit -
    success or failure/blocked alike (e.g. the low-battery gate returning 1
    before firing) - so a blocked run never dumps the operator straight
    out. Routed through _prompt_pick(), so it's arrow-navigable like
    every other y/n in this file."""
    choice = _prompt_pick(
        "continue in phcal? y/n [default y]: ", {"y", "n"}, default="y"
    )
    return choice == "y"


def run_guided_flow():
    """2026-08-18, PHCAL_NAV_POLISH_001.md: the real public entry point now -
    a thin stay-in-loop wrapper around _run_guided_flow_once() (the actual
    guided flow, renamed but otherwise unchanged). _PhcalEscExit is caught
    HERE, not inside _run_guided_flow_once() itself, so that function's raw
    ~450-line dispatch body didn't need reindenting into a try block - the
    exception still propagates cleanly up through every intermediate return
    path since none of them catch it, and lands in this one handler.

    None from _run_guided_flow_once() means the operator explicitly chose
    to exit from the menu structure itself (group-menu "0", a sub-menu
    invalid-input exit, or an ESC caught here) - leave immediately, no
    continue-or-exit prompt. Any other return (0 or nonzero - a primitive
    dispatch attempted or completed, success or blocked) goes through
    _prompt_continue_or_exit() per operator decision 5 before looping back
    in or leaving for real.

    2026-08-19 addendum: calls _resolve_session_mode_once() exactly once,
    right here, before the loop starts - see that function's own docstring
    for why (this used to live inside _run_guided_flow_once() and re-ran on
    every "continue").

    2026-08-19 bugfix (live-test): the first cut of this addendum kept
    looping when the low-voltage gate declined - session_ok stayed False for
    the life of the invocation, so every pass synthesized a blocked result
    (1) and re-asked continue-or-exit, and answering "y" just looped back to
    ask the same question again with no real menu underneath it. A startup
    decline is an operator "don't run" signal, not a blocked mid-session
    primitive - it now exits cleanly outright, with no continue-or-exit
    prompt at all. Decision 5's own guarantee (a blocked run reaching
    continue-or-exit rather than being dumped straight out) still applies
    to what it always meant: a primitive dispatch attempted mid-session
    inside _run_guided_flow_once() and returning nonzero - untouched below.

    2026-08-19 (PHCAL_PROMPT_UNIFY_001.md): _resolve_session_mode_once()
    itself is now wrapped in the same try/except _PhcalEscExit as the loop
    below. Two reasons, one fix: the multi-mode-confirm prompt inside it
    just got converted to the ESC-raising _prompt_choice() this same pass,
    and the low-voltage gate's own _prompt_choice() call (already
    ESC-raising since the earlier nav-polish pass) was ALSO unprotected
    here - pressing ESC at "proceed anyway?" would have propagated
    _PhcalEscExit straight out of this function uncaught. Neither prompt
    had a live ESC test exercised against this exact call boundary before
    now; both are fixed by the same one-line move."""
    try:
        session_ok = _resolve_session_mode_once()
    except _PhcalEscExit:
        print("PHCAL_GUIDED_EXIT (ESC) exiting")
        return 0
    if not session_ok:
        return 1
    while True:
        try:
            result = _run_guided_flow_once()
        except _PhcalEscExit:
            print("PHCAL_GUIDED_EXIT (ESC) exiting")
            return 0
        if result is None:
            return 0
        if not _prompt_continue_or_exit():
            return result


# 2026-08-15, MASTER_TWEAKS_STAGE2_LOGS_001.md: mirrors every print() in
# this process (this file's own ~200 call sites, plus anything printed by
# a song-engine module loaded in-process via _load_module()) to both the
# real terminal and phcal_run.log - one wrap at the entry point below,
# not a change to any individual print() call. Lives in
# ~/.gopod_alias_lib/, the same untracked, no-git-repo-at-all folder
# phcal_last.json already lives in - provably outside git, never
# committable. Overwrites fresh each run (same "latest state, not an
# accumulating history" convention phcal_last.json already uses).
PHCAL_RUN_LOG_PATH = Path(__file__).resolve().parent / "phcal_run.log"


class _TeeStdout:
    def __init__(self, real_stdout, log_path):
        self._real = real_stdout
        self._log = open(log_path, "w", buffering=1)

    def write(self, data):
        self._real.write(data)
        self._log.write(data)

    def flush(self):
        self._real.flush()
        self._log.flush()

    def isatty(self):
        return self._real.isatty()


if __name__ == "__main__":
    sys.stdout = _TeeStdout(sys.stdout, PHCAL_RUN_LOG_PATH)
    if len(sys.argv) == 2 and sys.argv[1] == "--promote":
        sys.exit(_promote_tweaks())
    if len(sys.argv) == 1:
        try:
            sys.exit(run_guided_flow())
        except EOFError:
            # No interactive stdin to read from (piped/scripted call with no
            # TTY) - a plain blocked message, not a raw traceback.
            print("PHCAL_GUIDED_BLOCKED no interactive input available - run phcal from a real terminal")
            sys.exit(1)
        except KeyboardInterrupt:
            print("\nPHCAL_GUIDED_BLOCKED cancelled")
            sys.exit(1)
    sys.exit(main())

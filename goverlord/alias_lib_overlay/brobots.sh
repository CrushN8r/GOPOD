# brobots.sh - the GOPOD cockpit, cleaned out 2026-10-08 (operator: "clean out aliases
# and alias studio"; kept: Bingo test + phcal, Wire-Pod basics, LLM aliases). Kept here:
# pha0b (PLAYHEAD), phcal, score/score-save, tempo-set, gobingo, the audio-routing
# check. The full earlier library is in gopod_notes/older_archives/alias_studio.zip.
# GOPOD_WIREPOD_BASE_URL - single source of truth for every Wire-Pod-talking
# alias in this file. Never a literal IP here (this file is tracked,
# mirrored into the public repo's goverlord/alias_lib_overlay/) - the real
# value lives in a private, untracked local file instead, sourced below if
# present (2026-08-15, BROBOTS_SH_IP_SCRUB_AND_BACKUP_RESYNC_001.md -
# replaces 15 separate inline ${GOPOD_WIREPOD_BASE_URL:-<literal LAN IP>}
# fallbacks with this one non-IP-bearing source point). Every alias below now references
# bare "$GOPOD_WIREPOD_BASE_URL", no inline default - resolution is
# unchanged in practice, since this file is always sourced (via
# .bashrc/.bash_aliases) before any of those aliases can be called, same as
# it always was.
GOPOD_LOCAL_CONFIG_PATH="${GOPOD_LOCAL_CONFIG_PATH:-$HOME/.gopod_alias_lib/local_config.sh}"
[ -f "$GOPOD_LOCAL_CONFIG_PATH" ] && . "$GOPOD_LOCAL_CONFIG_PATH"

# === Single Python instrument for this file's own HTTP notes ===
# Registry polish, 2026-07-16 (see ALIAS_REGISTRY_POLISH_001.md /
# ALIAS_REGISTRY_TRUTH_SWEEP_001.md). Before this pass, this file's own
# assume/release/move_lift/move_head/say-with-animation notes each shelled
# out to curl directly - a second, separately hand-written HTTP client for
# the exact same /api-sdk/* endpoints run_section1_full_live_001.py's own
# wirepod_web_send_form() already plays for every song. Converted onto that
# same function instead, via the identical importlib load pattern
# gopod-opening-chord's kokoro/wirepod jobs already use in core.sh
# (spec_from_file_location + exec_module, confirmed this session to have
# zero top-level side effects - the file's module level is pure path/env
# constants, nothing network- or hardware-touching) - not a third parallel
# instrument, the same one already proven in production. Every alias name
# and its observable behavior (echoes, ordering, timing, silent-on-failure)
# is unchanged; only the HTTP client underneath changed.
_GOPOD_NOTE_RUNNER_PATH="/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools/run_section1_full_live_001.py"

# 2026-08-09, DECOMPOSED per operator direction into brobots-wake/move-
# reverse/nudge-reverse. WAKE MECHANISM REPLACED 2026-08-10
# (WHEEL_NUDGE_COLD_FIRE_ROOT_CAUSE_SURVEY_001.md + its live-diagnostic
# follow-up sessions): the say_text-based "speak-ready" check below was
# proven wrong by direct measurement - 4/4 live runs against a freshly-
# released robot showed say_text returning in ~0.05-0.12s with NO audible
# speech, on either robot. The real, measured mechanism: assume, then
# RELEASE behavior control right back (a "release pulse"), settle, then
# RE-ASSUME - hands control back to the robot's own onboard behavior just
# long enough for it to visibly/physically wake, mirroring
# ROBOT_SLEEP_DIRECT_SDK_BUILT_001.md's own first live finding (releasing
# BehaviorControl right after a forced sleep trigger wakes a robot on its
# own). Live-proven 8/8 runs across BOTH robots, head-pop-to-visible-wake
# measured at 1.03-1.82s every time, first move_wheels succeeding at just
# 0.5s after re-assume every run. Same replacement already made in
# run_robot_control_song_001.py's own run_brobots_wake() and
# phcal_isolate_001.py's own run_wheel_nudge_wake_release_pulse().
#
#   brobots-wake  - GLOBAL, reusable by any future motion primitive, not
#                   wheel-specific. conn_test (pre-check) -> assume ->
#                   golden-flag release/settle/re-assume pulse. Leaves
#                   control ASSUMED when composed - only releases when
#                   fired standalone.
#   move-reverse  - assumes control is ALREADY held. ONE real move_wheels
#                   call (the old "prime" duplicate removed 2026-08-10 -
#                   proven unnecessary once the golden flag lands), hold,
#                   stop. Only assumes its own control when fired
#                   standalone.
#   nudge-reverse - composes the two: brobots-wake's own core (no release)
#                   then move-reverse's own core (no assume), then ONE
#                   release at the end - the same continuous held-control
#                   window the whole fix has been about since round 2.
#
# _brobots_wake_core/_brobots_move_reverse_core are the shared, no-
# assume/no-release halves both the standalone wrappers and nudge-reverse
# itself call - one implementation, not three.
_BROBOTS_WHEEL_WAKE_RELEASE_SETTLE_SECONDS=2.5
# BUG FOUND AND FIXED 2026-08-10: nothing ever settled reassume->move. The
# probe that proved this mechanism never fired move_wheels at 0s after
# re-assume (its own sweep starts at 0.5s, and 8/8 clean runs succeeded at
# exactly that first attempt - WHEEL_NUDGE_COLD_FIRE_ROOT_CAUSE_SURVEY_001.md).
# That gap was missing here - live-confirmed 2026-08-10: wheel reversal
# silently no-opped on both robots even though every log line read ok=True.
# Same fix as run_robot_control_song_001.py's own
# BROBOTS_WAKE_POST_REASSUME_SETTLE_SECONDS.
_BROBOTS_WHEEL_WAKE_POST_REASSUME_SETTLE_SECONDS=0.5

# _gopod_check_audio_routing - startup audio-routing gate, added 2026-08-19
# (AUDIO_ROUTING_CHECK_001.md). Wire-Pod/robots log say=success even when
# audio is misrouted and nothing is actually heard (same "markers don't
# prove completion" trap CLAUDE.md's own gotchas already document for
# say_text/assume_behavior_control - this is the audio-hardware version of
# it) - the operator hit this on a real interview run: reporter voices never
# heard, robots logged success throughout. Real cause: PulseAudio's default
# sink/source drifting off the physical GOPOD devices, most often because a
# remote NoMachine session grabs them (this machine's own pactl list shows
# NoMachine's own nx_voice_out sink / nx_remapped_out source sitting right
# alongside the real hardware devices - an easy silent hijack, confirmed via
# AUDIO_ROUTING_CHECK_001.md's survey, not assumed).
#
# Expected devices are stable pactl NAMES (indexes drift on reboot/replug,
# names don't), config-driven so this portable to another Linux+PulseAudio
# machine (AUDIO_ROUTING_CHECK_002.md): set here first as THIS Jetson's own
# real fallback default (confirmed via `pactl list short sinks/sources`, not
# invented), then overridden if gopod_audio_config.sh - this file's own
# sibling, real per-machine values, never copied into the tracked overlay
# mirror - exists and is sourced. Same "config outside the repo tree, code
# fallback if missing" pattern phcal_config.json already established
# (PHCAL_DEHARDCODE_001.md) - just bash-native (sourced vars) instead of
# JSON, since brobots.sh has zero jq/JSON-parsing dependency today and two
# strings don't justify adding one. Missing config file -> both names stay
# exactly this Jetson's hardcoded values, so nothing changes here. Demo/
# template: gopod_audio_config.example.sh (committed, placeholder values,
# points at `pactl list short sources`/`sinks` to find your own).
GOPOD_EXPECTED_AUDIO_SINK="alsa_output.platform-sound.analog-stereo"     # Built-in Audio Analog Stereo
GOPOD_EXPECTED_AUDIO_SOURCE="alsa_input.usb-TTGK_Technology_USB_Audio_330212E9240828-00.mono-fallback"  # USB Audio Mono
_GOPOD_AUDIO_CONFIG_PATH="$(dirname "${BASH_SOURCE[0]}")/gopod_audio_config.sh"
if [ -f "$_GOPOD_AUDIO_CONFIG_PATH" ]; then
  source "$_GOPOD_AUDIO_CONFIG_PATH"
fi
#
# Warn-and-offer, never auto-force: on mismatch this prints exactly what's
# wrong and the manual fix command to stderr, then asks before touching
# system audio at all (default answer is NO on bare Enter - see
# _pha0b_prompt_yn's n-means-no mode below). Non-blocking either way - this
# is a fast "will sound work" gate, not a hard dependency; the caller
# proceeds regardless of the answer, now informed instead of finding out
# after a silent run.
_gopod_check_audio_routing() {
  if ! command -v pactl >/dev/null 2>&1; then
    echo "AUDIO_ROUTING_CHECK_SKIPPED pactl not found on this machine - cannot verify routing" >&2
    return 0
  fi
  local current_sink current_source mismatch=0
  current_sink="$(pactl get-default-sink 2>/dev/null)"
  current_source="$(pactl get-default-source 2>/dev/null)"

  if [ "$current_sink" != "$GOPOD_EXPECTED_AUDIO_SINK" ]; then
    echo "AUDIO_ROUTING_MISMATCH sink: got '$current_sink', expected '$GOPOD_EXPECTED_AUDIO_SINK' (Built-in Audio Analog Stereo)" >&2
    mismatch=1
  fi
  if [ "$current_source" != "$GOPOD_EXPECTED_AUDIO_SOURCE" ]; then
    echo "AUDIO_ROUTING_MISMATCH source: got '$current_source', expected '$GOPOD_EXPECTED_AUDIO_SOURCE' (USB Audio Mono)" >&2
    mismatch=1
  fi

  if [ "$mismatch" = "0" ]; then
    echo "AUDIO_ROUTING_OK sink=$current_sink source=$current_source" >&2
    return 0
  fi

  echo "AUDIO_ROUTING_WARNING say=success will still log even though nothing will be heard until this is fixed (NoMachine or a default-sink/source drift is the usual cause)." >&2
  echo "AUDIO_ROUTING_FIX_CMD pactl set-default-sink '$GOPOD_EXPECTED_AUDIO_SINK'; pactl set-default-source '$GOPOD_EXPECTED_AUDIO_SOURCE'" >&2
  local fix_choice
  fix_choice="$(_pha0b_prompt_yn "fix audio routing now? y/n [default n]: " n n-means-no)"
  if [ "$fix_choice" = "1" ]; then
    local sink_rc source_rc
    pactl set-default-sink "$GOPOD_EXPECTED_AUDIO_SINK" >&2 2>&1; sink_rc=$?
    pactl set-default-source "$GOPOD_EXPECTED_AUDIO_SOURCE" >&2 2>&1; source_rc=$?
    if [ "$sink_rc" = "0" ] && [ "$source_rc" = "0" ]; then
      echo "AUDIO_ROUTING_FIXED sink=$GOPOD_EXPECTED_AUDIO_SINK source=$GOPOD_EXPECTED_AUDIO_SOURCE" >&2
      return 0
    else
      echo "AUDIO_ROUTING_FIX_FAILED pactl rejected the set-default call(s) above (sink_rc=$sink_rc source_rc=$source_rc) - device name may be stale, re-check with 'pactl list short sinks/sources'" >&2
      return 1
    fi
  else
    echo "AUDIO_ROUTING_DECLINED proceeding without fixing - audio may be silent" >&2
  fi
  return 1
}

# _gopod_set_audio_routing - set, no question asked. Operator ask 2026-10-07:
# "Please set at run start" the speakers and the USB mic, after the reporter
# voices went silent because the default sink had drifted to the USB device.
# Same two names as the check above. Called by wpr, the operator's run start.
_gopod_set_audio_routing() {
  command -v pactl >/dev/null 2>&1 || return 0
  if pactl set-default-sink "$GOPOD_EXPECTED_AUDIO_SINK" 2>/dev/null \
    && pactl set-default-source "$GOPOD_EXPECTED_AUDIO_SOURCE" 2>/dev/null; then
    echo "AUDIO_ROUTING_SET speakers=Built-in Audio Analog Stereo mic=USB Audio Mono"
  else
    echo "AUDIO_ROUTING_SET_FAILED check 'pactl list short sinks' / 'pactl list short sources'" >&2
  fi
}

# bingo-video-song-pick-segment retired 2026-08-11 (bingo golden-lock pass,
# operator direction: "no crucial use, prune"). It picked one of bingo's
# sections via GOPOD_BINGO_SECTION on the legacy run_songs_runner_001.py -
# never cut over to the golden engine (which doesn't read that env var at
# all). Fully superseded by pha0b bingo's own numbered division picker
# (same section list, playhead A/B range, runs on run_golden_song_001.py).
# No other caller found. Use pha0b bingo instead.

# NOT phaob() (no zero, ~/.gopod_alias_lib/suits.sh) - that is a different,
# unrelated alias (a static gopod_alias/Suit Changer fitting recap, no
# song/robot involvement at all). One character apart, two different tools -
# if you meant the Open WebUI suit check, you want phaob, not this.
#
# pha0b - PlayHead A/0/B - the studio-wide Playhead Point A/0/B cockpit,
# first alias over the lifted playhead engine (STUDIO_SONG_TOOL_SURVEY_001.md,
# CONTROL_SONG_PLAYHEAD_LIFT_EXECUTED_001.md, PLAYHEAD_COCKPIT_ALIAS_001.md,
# PHA0B_RENAME_001.md). Renamed 2026-07-19 from this function's original
# name, song-playhead() - the operator claimed "pha0b" for this tool once
# the old, stale pha0b() alias in suits.sh (dead reference to a file that no
# longer exists) was confirmed gone and nothing else in the alias lib called
# it. Behavior unchanged by the rename - same routing, same validation, same
# dry-by-default.
#
# Picks a song off the shelf, takes Point A (start step_id) and Point B (end
# step_id), routes to the right runner and sets THAT runner's own playhead
# env vars - the operator never picks GOPOD_BINGO_PLAYHEAD_* vs.
# GOPOD_CONTROL_SONG_PLAYHEAD_* by hand, this alias hides the split. Point 0
# (the actual A->B advance during play) is the runner's own existing step
# loop, untouched - this alias only ever decides which slice plays.
#
# Deliberately does NOT re-derive the step_id/ordering check in bash - both
# runners already refuse a bad A/B with their own
# RuntimeError("BLOCKED: ...") before anything fires (dry-verified in
# CONTROL_SONG_PLAYHEAD_LIFT_EXECUTED_001.md). This alias lets that
# validation run for real, tees stderr live to the terminal (nothing
# hidden) and also greps the runner's own BLOCKED line out of it afterward,
# so a refusal reads as one clean line on top of the full trace, not a
# guess at the runner's own rules re-coded here.
#
# brobots_interview_section_01 and brobots_preshow have no standalone
# step-loop runner to attach a slice to (STUDIO_SONG_TOOL_SURVEY_001.md
# §3/§5, both driven inline by generate_phase()/playback_phase() inside
# run_section1_full_live_001.py itself) - refused here by name, never
# passed through to either runner.
#
# Playhead wake defaults ON (GOPOD_*_PLAYHEAD_WAKE=1) whenever A/B are set -
# a slice that doesn't start at the song's own first step needs that first
# step's own assume/connect fired before it, and re-firing it when A IS the
# first step is a harmless repeat of the exact call that step would have
# made itself. Dry by default, same convention as every other song alias in
# this file - export GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 yourself first for real
# hardware. Replay/in-place editing of buffer times or spoken text is a
# separate, later feature - not built here.
#
# live_robots_prompt() - shared golden y/n prompt, default y (added
# 2026-07-24 per operator request after a dry-by-silent-default bait run
# gave "no robot movements" with no warning; decoupled into one shared
# function the same day so pha0b() and phcal() never carry two copies of
# the same sequence - same "never write a second copy" precedent as
# restart_wirepod_preflight() in core.sh / the wirepod-restart-discipline
# skill). Bare Enter or "y"/"Y" = live; "n"/"N" = dry. Echoes "1" (live) or
# "0" (dry) to stdout and nothing else, so a caller captures it cleanly via
# command substitution: live_gate="$(live_robots_prompt)". Safe to capture
# this way because read -p's own prompt text is written to the terminal
# (not stdout) whenever stdin is an interactive tty, and is silently
# suppressed (not leaked into the captured value either way) when stdin
# isn't a tty - confirmed both ways before relying on it here.
#
# GOLDEN INVARIANT, root-caused 2026-08-15 (PHCAL_LIVE_GATE_GOLDEN_001.md): a
# `echo "** default to 'y' **"` visual hint line was added straight to stdout
# INSIDE this function at some point after the header comment above was
# written. Because both callers capture the ENTIRE function via
# `$(live_robots_prompt)`, that hint text rode along with the real "0"/"1"
# signal - `live_gate` became the two-line string `"** default to 'y' **\n1"`,
# which matches neither `= "1"` (phcal()) nor `-eq 1` (pha0b(), which threw a
# silent "integer expression expected" and fell through to dry) even on a
# live "y" answer. Both songs and single-primitive calibrations ran dry with
# no error the operator would necessarily connect to this. Root-fixed by
# sending any future human-facing hint text in this function to stderr
# (`>&2`), never plain stdout - stdout in this function is the "0"/"1" signal
# and ONLY the "0"/"1" signal, permanently. Do not add a bare `echo`/`print`
# to this function's stdout again; use `>&2` for anything meant for the
# terminal only.
# _live_robots_read_key() - one raw keypress for live_robots_prompt()'s own
# arrow-nav, 2026-08-18 PHCAL_NAV_POLISH_001.md. Deliberately NOT shared
# with phcal_isolate_001.py's Python _read_key() - PHCAL_NAV_POLISH_
# SURVEY_001.md found the pha0b/bash read-ahead pattern doesn't port
# cleanly across the Python/bash boundary, so this is its own small,
# bash-native implementation, same shape only. Echoes exactly one of:
# up down enter esc y n other. Does its own stty save/restore - none of
# it; the CALLER (live_robots_prompt) owns stty save/restore around the
# whole read loop, once, not per keypress, same as _raw_mode() on the
# Python side owns termios save/restore once per _prompt_choice() call.
_live_robots_read_key() {
  local k rest
  IFS= read -rsn1 k
  # 2026-08-19, PHCAL_PROMPT_UNIFY_001.md: `read -n1` returns an empty
  # string in $k when the single character it consumes is the newline
  # delimiter (a documented bash quirk - -n honors the delimiter and
  # strips it, same as a normal line read would, even though only 1 char
  # was requested). Pressing Enter therefore produced k="", which matched
  # none of the case patterns below (not $'\r', not $'\n', not y/n) and
  # fell through to "other" - silently ignored by every caller's `case`
  # (the `*) : ;;` no-op arm), so Enter appeared to do nothing. Verified
  # via a real pty-backed bash subprocess: raw CR and raw LF bytes both
  # produced k="" and classified as "other" before this check was added.
  # Checking for empty FIRST, before the ESC/case logic, fixes it - ESC and
  # every other real key still produce a non-empty $k, unaffected.
  if [ -z "$k" ]; then
    echo enter
    return
  fi
  if [ "$k" = $'\x1b' ]; then
    # Distinguish a bare ESC from the start of an arrow escape sequence
    # with a short timed read - same reasoning as the Python side's
    # select() 0.05s timeout in _read_key().
    if IFS= read -rsn2 -t 0.05 rest; then
      case "$rest" in
        '[A') echo up ;;
        '[B') echo down ;;
        '[C') echo down ;;  # right treated as "down" (advance), matches Python side
        '[D') echo up ;;    # left treated as "up" (back), matches Python side
        *) echo esc ;;
      esac
    else
      echo esc
    fi
    return
  fi
  case "$k" in
    y|Y) echo y ;;
    n|N) echo n ;;
    *) echo other ;;
  esac
}

live_robots_prompt() {
  local live_choice highlight old_stty key
  if [ ! -t 0 ]; then
    # No interactive tty (piped/scripted call) - keep the original plain
    # input() fallback exactly, arrow-nav needs a real terminal.
    read -r -p "live robots? y/n [default y]: " live_choice
    live_choice="${live_choice:-y}"
    if [ "$live_choice" = "n" ] || [ "$live_choice" = "N" ]; then
      echo 0
    else
      echo 1
    fi
    return
  fi

  highlight=0  # 0 = y (default), 1 = n
  old_stty="$(stty -g 2>/dev/null)"
  trap 'stty "$old_stty" 2>/dev/null' RETURN INT
  stty -echo -icanon min 1 time 0 2>/dev/null

  _live_robots_redraw() {
    if [ "$highlight" = 0 ]; then
      printf '\r\033[2Klive robots? [y] n  (arrows to move, Enter to select, ESC = abort): ' >&2
    else
      printf '\r\033[2Klive robots?  y [n]  (arrows to move, Enter to select, ESC = abort): ' >&2
    fi
  }
  _live_robots_redraw

  while true; do
    key="$(_live_robots_read_key)"
    case "$key" in
      up|down)
        highlight=$((1 - highlight))
        _live_robots_redraw
        ;;
      enter)
        printf '\n' >&2
        if [ "$highlight" = 0 ]; then live_choice=y; else live_choice=n; fi
        break
        ;;
      y) printf '\n' >&2; live_choice=y; break ;;
      n) printf '\n' >&2; live_choice=n; break ;;
      esc)
        # 2026-08-19 refinement (PHCAL_NAV_POLISH_001.md addendum, operator
        # review): ESC here has no menu level above it to back up to - this
        # prompt fires before phcal/pha0b's own flow even starts - so it no
        # longer degrades to a safe dry "n" (that let phcal/pha0b start
        # anyway, just dry). It now means a full abort: the whole
        # invocation stops right here, phcal/pha0b never starts at all.
        printf '\n' >&2
        printf 'live robots prompt: ESC - aborting, phcal/pha0b will not start\n' >&2
        live_choice=abort
        break
        ;;
      *) : ;;
    esac
  done

  stty "$old_stty" 2>/dev/null
  trap - RETURN INT
  unset -f _live_robots_redraw

  # Stdout signal set, GOLDEN INVARIANT-preserving (see this function's own
  # header comment - stdout carries ONLY this signal, nothing else): "0" =
  # dry, "1" = live, "2" = full abort (ESC, 2026-08-19 addendum) - every
  # caller must check for "2" immediately after capturing this function's
  # output and return before doing anything else, same as phcal()/pha0b()
  # now do.
  if [ "$live_choice" = "abort" ]; then
    echo 2
  elif [ "$live_choice" = "n" ]; then
    echo 0
  else
    echo 1
  fi
}

# _pha0b_prompt_range_choice / _pha0b_prompt_choice / _pha0b_prompt_yn -
# pha0b's own shared choke points (PHA0B_MENU_CONSOLIDATION_001.md), the
# bash equivalent of phcal's own _prompt_choice() (phcal_isolate_001.py).
# Collapse pha0b's repeated numbered/lettered pick-from-list and y/n
# prompts into one place each, same input behavior as before - a scatter-
# to-one-function refactor, not a nav/behavior redesign (no arrow keys
# here; that's a later, separate task). Each caller keeps its own hint
# `echo` line(s) and its own post-choice dispatch (case/if) untouched -
# only the read-and-validate boilerplate moved.
#
# live_robots_prompt() above is deliberately NOT rebuilt on top of these -
# it already is its own working choke point, and its GOLDEN INVARIANT
# comment (below its own body) explicitly warns against internal
# modification after a real historical bug there - left fully intact.
#
# Every existing call site's own prompt text/choice-set/default/block-
# message was mapped 1:1 before this refactor - see
# PHA0B_MENU_CONSOLIDATION_001.md for the full call-site-by-call-site
# comparison, including the shapes deliberately NOT collapsed here (Point
# A/B, the bingo-game run/grid picks, the interview v/p/g pick, robot-
# pick's own step_id/speaker free-text entry) because their real
# validation/default/dispatch shape diverges from these two functions'
# and forcing them through would have changed behavior, not just moved it.

# _pha0b_prompt_range_choice - numbered-range pick, invalid (non-numeric or
# out of [1, max]) prints "<block_prefix> invalid_choice choice=<input>" to
# STDERR (so a caller capturing this function's own stdout via command
# substitution still sees the message on the terminal) and returns 1.
# Matches pha0b_menu()'s song-pick prompt exactly, including its
# arithmetic (not string) range check - e.g. a zero-padded numeric answer
# like "007" is still accepted the same (quirky, pre-existing) way bash's
# own arithmetic array indexing already handled it before this refactor;
# not "fixed," faithfully preserved.
_pha0b_prompt_range_choice() {
  local prompt="$1" max="$2" block_prefix="$3"
  local input
  read -r -p "$prompt" input
  if ! [[ "$input" =~ ^[0-9]+$ ]] || [ "$input" -lt 1 ] || [ "$input" -gt "$max" ]; then
    echo "${block_prefix} invalid_choice choice=$input" >&2
    return 1
  fi
  echo "$input"
}

# _pha0b_prompt_choice - fixed-set pick (space-separated `choices`, exact
# string match, not a numeric range - e.g. robot-filter's {1,2,0}).  Empty
# input is filled with `default` before the membership check, matching
# every existing call site's own empty-means-default behavior. Invalid
# (not in the set after the default fill) prints the same
# "<block_prefix> invalid_choice choice=<input>" to STDERR and returns 1.
_pha0b_prompt_choice() {
  local prompt="$1" choices="$2" default="$3" block_prefix="$4"
  local input c
  read -r -p "$prompt" input
  input="${input:-$default}"
  for c in $choices; do
    if [ "$input" = "$c" ]; then
      echo "$input"
      return 0
    fi
  done
  echo "${block_prefix} invalid_choice choice=$input" >&2
  return 1
}

# _pha0b_prompt_yn - y/n confirm. Two genuinely different default shapes
# exist today across pha0b's own y/n prompts and both are preserved
# exactly via `mode`, not flattened into one:
#   mode=n-means-no: only "n"/"N" means no; anything else (including empty)
#     means yes - rich-display's and reporter-gap's own existing shape,
#     same shape live_robots_prompt() above already uses (that one stays
#     separate, see note above).
#   mode=y-means-yes: only "y"/"Y" means yes; anything else means no -
#     robot-pick's own existing shape.
# `fill_empty` (may be "") is substituted for a bare Enter BEFORE the
# mode check - phcal-apply's own prompt explicitly defaults empty input to
# "y" before its y/Y-only check (`apply_choice="${apply_choice:-y}"` in the
# pre-refactor code), which every other call site's original code did NOT
# do - preserved per-caller via this parameter, not assumed uniform.
# Echoes "1" (yes) or "0" (no) to stdout only.
_pha0b_prompt_yn() {
  local prompt="$1" fill_empty="$2" mode="$3"
  local choice
  read -r -p "$prompt" choice
  choice="${choice:-$fill_empty}"
  if [ "$mode" = "n-means-no" ]; then
    if [ "$choice" = "n" ] || [ "$choice" = "N" ]; then
      echo 0
    else
      echo 1
    fi
  else
    if [ "$choice" = "y" ] || [ "$choice" = "Y" ]; then
      echo 1
    else
      echo 0
    fi
  fi
}

# pha0b() calls live_robots_prompt() once per call, every song, before the
# range runs. "1" = live (exports GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 for this
# call only); "0" = dry, same as never exporting the gate at all. Replaces
# the old bingo-only auto-live special case - bingo now goes through the
# same prompt as every other song, default answer preserves its old
# always-live behavior for a bare Enter.
#
# Usage: pha0b <song> <point_a_step_id> <point_b_step_id> [robot]
#   song:  bingo | nap
#   robot: 1 or 2, control/bait only, default 1 - which robot's own
#     run to slice (bait is a single-robot-per-run score, same
#     GOPOD_CONTROL_SONG_ROBOT knob start-the-control-song already uses).
#     Bingo and vamp take no robot argument - bingo's score alternates
#     speakers, vamp has no physical robot at all (voice-only hosts).
# ==========================================================================
# === PHA0B COCKPIT - pha0b()/pha0b_menu(): the studio-wide song-slice   ===
# === cockpit. Bare `pha0b` opens pha0b_menu() (below); called with      ===
# === explicit args it skips straight to the runner.                    ===
# ==========================================================================
pha0b() {
  if [ $# -eq 0 ]; then
    pha0b_menu
    return $?
  fi

  local song="$1"
  local point_a="$2"
  local point_b="$3"
  local robot="${4:-1}"

  if [ -z "$song" ] || [ -z "$point_a" ] || [ -z "$point_b" ]; then
    echo "PHA0B_USAGE pha0b <song> <point_a_step_id> <point_b_step_id> [robot]"
    echo "  song: bingo | nap"
    return 1
  fi

  local tools_dir="/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools"
  local songs_dir="/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs"
  local runner=""
  local env_prefix=""
  local song_dir_export=""

  # --- Song keyword dispatch: resolves runner/env_prefix/song_dir_export ---
  # --- for the picked song keyword.                                     ---
  case "$song" in
    bingo)
      # goverlord/runtime/songs/01a_brobots_bingo_test/ - the flagship UPSELL 1 song.
      # CUT OVER 2026-08-07 (golden-song-runner Ladder-1 Phase 3, third
      # cutover, mirrors nap/mixup above -
      # GOLDEN_SONG_RUNNER_PHASE3_MIXUP_CUTOVER_001.md) from the legacy
      # run_songs_runner_001.py to the golden engine (run_golden_song_001.py)
      # - already golden-registered as "brobots_bingo" (SONG_REGISTRY, same
      # bingo-family shape as nap/mixup: fatal_notes={wake_both,emotion_beat},
      # manage_control=True, no note aliases). The phcal-apply and
      # reporter-gap override prompts below are both gated on $song, not
      # $runner, so they carry through this cutover unchanged - still fire
      # for song=bingo exactly as before. Plays DIRTY: bingo's own
      # zKnobs.json/knobs.json diverge today (arm_cue cycles 2 vs 1) - the
      # shared resolver (knobs_envelope_001.py, already backing this engine's
      # load_golden_song()) resolves zKnobs.json here, same file the
      # phcal-apply writer already targets (PHCAL_WRITER_DIRTY_FIX_001.md) -
      # tune and play now read/write the same file end to end. env_prefix
      # changed to the golden engine's own fixed playhead namespace
      # (GOPOD_GOLDEN_PLAYHEAD_FROM/_TO/_WAKE), same as nap/mixup. Bingo is
      # NOT the golden engine's own DEFAULT_SONG_ID (still
      # brobots_baby_robots_sleep) - so song_dir_export is now set (it wasn't
      # needed before, since run_songs_runner_001.py's own default song WAS
      # bingo) to reach the same GOPOD_GOLDEN_SONG_DIR gated-export line
      # mixup's cutover added to the shared dispatch block below.
      runner="run_golden_song_001.py"
      env_prefix="GOPOD_GOLDEN_PLAYHEAD"
      song_dir_export="$(_song_dir_by_id brobots_bingo)"
      ;;
    nap)
      # goverlord/runtime/songs/101_brobots_nap/ (RENAMED/RENUMBERED
      # 2026-08-18 from 104_brobots_baby_robots_sleep/, operator's own folder
      # surgery - song_id itself, "brobots_baby_robots_sleep", is UNCHANGED,
      # only the directory moved) - "Do Baby Robots Dream?" scored capture
      # song. Wired 2026-08-06 (Phase 0) onto the legacy
      # run_songs_runner_001.py; CUT OVER 2026-08-07 (golden-song-runner
      # Phase 3, first cutover) to the golden engine
      # (run_golden_song_001.py) - the first and only song moved so far.
      # First cutover picked as safest: smallest note vocabulary, shortest
      # range, already golden-engine-proven live and dry with matching
      # (step_id, note, ok) parity against this exact legacy baseline
      # (GOLDEN_SONG_RUNNER_PHASE1_ENGINE_PROVED_001.md). env_prefix changed
      # to the golden engine's own fixed playhead namespace
      # (GOPOD_GOLDEN_PLAYHEAD_FROM/_TO/_WAKE - run_golden_song_001.py's own
      # env vars are NOT song-derived, unlike the legacy per-song prefixes) -
      # the shared dispatch block below builds these generically off
      # env_prefix already, so no other code changed for this cutover.
      # song_dir_export is left as-is (harmless no-op for this runner: it
      # still exports GOPOD_CONTROL_SONG_DIR, which the golden engine never
      # reads, and GOPOD_BINGO_CAPTURE_SONG_DIR only fires when
      # runner="run_songs_runner_001.py", which is no longer true here) -
      # song_dir_export is still set below (2026-08-18 fix: an earlier draft
      # of this rename left it unset, reasoning the golden engine's own
      # DEFAULT_SONG_ID fallback covers it - true for the golden engine
      # itself, but song_dir_export is also read by the shared robot-pick
      # block further down this function (`robot_pick_song_dir=
      # "$song_dir_export"` for song=nap|mixup|bingo|itsyou-*), which has NO
      # such fallback and would have silently skipped its own prompt for
      # song=nap with this unset. Kept as a real path here, same as every
      # other song, not deliberately omitted). GOPOD_GOLDEN_SONG_DIR itself
      # is still NOT separately exported for this song in the block below -
      # this song IS the golden engine's own DEFAULT_SONG_ID
      # (brobots_baby_robots_sleep, song_dir repointed 2026-08-18 to
      # 101_brobots_nap in run_golden_song_001.py's own SONG_REGISTRY), so
      # the engine resolves the same directory either way; song_dir_export
      # here exists for this function's OTHER local consumers, not for the
      # engine. Plays DIRTY on purpose: the golden engine prefers
      # zKnobs.json over knobs.json when present - this song HAS a
      # zKnobs.json (currently byte-identical to knobs.json, so no visible
      # difference yet), and that preference is intentional and untouched -
      # COMPOSER_RECOMPOSE_MECHANISM_SNAPSHOT_001.md. See
      # GOLDEN_SONG_RUNNER_PHASE3_NAP_CUTOVER_001.md for the full cutover
      # record.
      runner="run_golden_song_001.py"
      env_prefix="GOPOD_GOLDEN_PLAYHEAD"
      song_dir_export="$(_song_dir_by_id brobots_baby_robots_sleep)"
      ;;
    *)
      echo "PHA0B_REFUSED song=$song unknown - must be one of: bingo nap"
      return 1
      ;;
  esac

  # --- Shared prompts: live-robots gate, rich-display, phcal-apply       ---
  # --- tuning, reporter-gap override, robot-pick, robot-filter.         ---
  local live_gate
  if [ "$song" = "bait" ] && [ "$PHA0B_BAIT_NAV_ACTIVE" = "1" ]; then
    # PHA0B_GOLDEN_NAV_RIPREPLACE_AND_SONG1_SCOPE_001.md Round 1: bait's own
    # nav bridge already resolved this (phcal's real presence-aware session
    # mode, not this flat y/n) - pha0b_menu()'s bait branch pre-exported it.
    # Every other song keeps calling live_robots_prompt() exactly as before.
    live_gate="$PHA0B_BAIT_LIVE_GATE"
  else
    live_gate="$(live_robots_prompt)"
  fi
  if [ "$live_gate" = "2" ]; then
    # 2026-08-19 refinement (PHCAL_NAV_POLISH_001.md addendum): ESC at the
    # live-robots prompt now means a full abort - pha0b never starts.
    echo "PHA0B_ABORTED (ESC at live-robots prompt) - exiting, pha0b did not start"
    return 1
  fi

  # rich-display switch, added 2026-07-25 (operator request, alias-mixer
  # widening) - universal, every song, same unconditional shape as
  # live_robots_prompt() above (not gated to specific songs like the
  # phcal-apply/reporter-gap switches below, since console_rich_display is a
  # Robots-class-level default every song's own Robots() construction reads
  # the same way). Default ON (Enter/y) - GOPOD_CONSOLE_RICH_DISPLAY is left
  # unexported, so Robots.__init__'s own default ("1" unless explicitly "0")
  # applies. "n" exports "0" for this run only, forcing the console back to
  # the robot-safe line even where a song has its own distinct display_text
  # authored.
  local rich_display_live
  if [ "$song" = "bait" ] && [ "$PHA0B_BAIT_NAV_ACTIVE" = "1" ]; then
    rich_display_live="$PHA0B_BAIT_RICH_DISPLAY"
  else
    rich_display_live="$(_pha0b_prompt_yn "rich display on console? y/n [default y]: " "" n-means-no)"
  fi
  local rich_display_override=""
  if [ "$rich_display_live" = "0" ]; then
    rich_display_override="0"
  fi

  # A native handleGetLogs LOGS-route toggle (webserver.go) rode this same
  # y/n from 2026-08-03 to 2026-08-10 - deployed once, found to regress
  # wire-pod's own native "Show all logs" checkbox distinction (both routes
  # served identical content whenever this flag was "1"), reverted the same
  # day. config-ws/webserver.go is back to pure native and out of the
  # overlay entirely - see tech/WIRED-POD.md's "added, deployed, found to
  # regress native, reverted" note. Nothing reads a rich-logs flag file
  # anymore; this y/n only drives the two mechanisms below now.

  # Second cooperating mechanism off the same y/n, added 2026-08-08 - drives
  # the webroot page's own big chat-bubble window (index.html, wire_pod_overlay
  # copy, served from the overlay folder). A plain flag file
  # placed in the overlay webroot/ so the browser can fetch it statically,
  # re-read fresh on every poll, no caching. "n" here means the window stays
  # collapsed (0px) for this run; it is never expanded from any other source.
  # 2026-10-17 (survey E15: "a state file nothing reads"): gopod_rich_display_ui_state.json is no longer written - the chat page opens and closes its own chat (the Log button), and nothing read it.

  # phcal->pha0b tweak-apply prompt (PHCAL_PHA0B_BINGO_APPLY_PROMPT_001.md,
  # widened 2026-07-25 past bingo-only, REPORTER_GAP_SHARED_SWITCH_SURVEY_001.md
  # / operator request, once brobots_awaken's arm_test/head_nod notes also
  # started reading cycles/hold_seconds/speed from their own knobs.json).
  # Fires for song=bingo or song=bait - control behavior is completely
  # unchanged, no prompt at all (its song's arm_cue/head_nod still runs the
  # fixed test-sequence choreography, not knobs-driven).
  # Asks once, before the range runs. On y: walks every step_id inside the
  # chosen point_a..point_b range whose own "note" is arm_cue, nod, or
  # head_nod (and ONLY those - step order/text/speaker and every other note
  # type are untouched) and calls the existing, already-verified
  # ~/.gopod_alias_lib/phcal_apply_001.py <step_id> --yes --knobs
  # <this song's own knobs.json> for each one, which writes phcal's
  # last-confirmed cycles/hold_seconds/speed (phcal_last.json) into that one
  # step's own line - same tool PHCAL_PHA0B_BINGO_LINK_001.md already built
  # and applied live, reused/generalized here, not rebuilt. On n (or
  # anything else): skip entirely, the range runs exactly as it always has.
  local phcal_song_dir=""
  case "$song" in
    bingo) phcal_song_dir="$(_song_dir_by_id brobots_bingo)" ;;
  esac
  # phcal_apply_knobs stays the CLEAN path, unchanged - it is only ever
  # handed to phcal_apply_001.py's own --knobs flag below, and that
  # tool already resolves dirty-over-clean internally (PHCAL_WRITER_DIRTY_FIX_001.md).
  # The scanner below (which step_ids qualify for the prompt) is the piece
  # that was clean-only until this pass (KNOBS_TOUCHERS_MAP_001.md #11) - it
  # now reads through the same shared resolver (knobs_envelope_001.py) so the
  # step_ids it offers reflect whichever file is actually live.
  local phcal_apply_knobs=""
  if [ -n "$phcal_song_dir" ]; then
    phcal_apply_knobs="$phcal_song_dir/knobs.json"
  fi
  if [ -n "$phcal_apply_knobs" ]; then
    local apply_live
    if [ "$song" = "bait" ] && [ "$PHA0B_BAIT_NAV_ACTIVE" = "1" ]; then
      apply_live="$PHA0B_BAIT_PHCAL_APPLY"
    else
      echo "* default to 'y'"
      apply_live="$(_pha0b_prompt_yn "apply phcal tweaks to this selected range? y/n " y y-means-yes)"
    fi
    if [ "$apply_live" = "1" ]; then
      local qualifying_ids
      qualifying_ids="$(python3 -c "
import sys
sys.path.insert(0, '$tools_dir')
from knobs_envelope_001 import load_knobs_envelope
_, d = load_knobs_envelope('$phcal_song_dir')
steps = d.get('steps', [])
ids = [s['step_id'] for s in steps]
try:
    ia = ids.index('$point_a')
    ib = ids.index('$point_b')
except ValueError:
    raise SystemExit(0)
lo, hi = (ia, ib) if ia <= ib else (ib, ia)
for s in steps[lo:hi + 1]:
    # Widened 2026-08-16 (MASTER_TWEAKS_STAGE4_APPLY_SEAM_001.md) to match
    # phcal_apply_001.py's own NOTE_TO_PRIMITIVE keys exactly - if that
    # map ever gains/loses an entry, this tuple must too, or the two
    # sides drift (this scanner decides what's OFFERED; the python tool
    # decides what's ACTUALLY appliable - they must agree).
    if s.get('note') in ('arm_cue', 'nod', 'head_nod', 'animation', 'wheel_nudge', 'brobots_ready_together'):
        print(s['step_id'])
")"
      if [ -n "$qualifying_ids" ]; then
        # 2026-10-17 (operator: "the PHCAL-APPLY spam > summary message
        # \"PLAYHEAD Calibrations Applied\""): each step's diff goes to a log in
        # the song's own runs/ (private); the screen gets one line.
        local apply_log apply_out apply_total=0 apply_changed=0 apply_blocked=0
        mkdir -p "$phcal_song_dir/runs"
        apply_log="$phcal_song_dir/runs/phcal_apply_$(date +%Y-%m-%d_%H-%M-%S).log"
        while IFS= read -r sid; do
          [ -z "$sid" ] && continue
          apply_total=$((apply_total + 1))
          apply_out="$(python3 /home/goverlord/.gopod_alias_lib/phcal_apply_001.py "$sid" --yes --knobs "$phcal_apply_knobs" 2>&1)"
          printf '%s\n' "$apply_out" >> "$apply_log"
          case "$apply_out" in
            *PHCAL_APPLY_BLOCKED*) apply_blocked=$((apply_blocked + 1)) ;;
            *"<-- CHANGES"*) apply_changed=$((apply_changed + 1)) ;;
          esac
        done <<< "$qualifying_ids"
        echo "PLAYHEAD Calibrations Applied: $apply_changed of $apply_total steps changed$([ "$apply_blocked" -gt 0 ] && echo ", $apply_blocked blocked") (details: $apply_log)"
      else
        echo "PHA0B_PHCAL_APPLY no arm_cue/nod/head_nod steps found in range $point_a..$point_b"
      fi
    fi
  fi

  # Robot-picker v1 (WIDGET_BOUNDARY_SURVEY_001.md Part D's recommended
  # first buildable step) - reassigns a single step's speaker between
  # brobot_1/brobot_2, via robot_pick_001.py, which reuses the same
  # knobs_envelope_001.py shared resolver + dirty-first, whole-file
  # load/mutate/write path phcal_apply_001.py already proved (Ladder-2
  # Phase 5). Scoped to song=nap/mixup/bingo only - the bingo-family songs
  # where SONG_REGISTRY (run_golden_song_001.py) confirms `speaker` is a
  # real, live-read per-step field. Deliberately NOT offered for song=bait
  # (00a_brobots_awaken): that song's SONG_REGISTRY entry sets
  # synthesize_speaker=True, meaning its own knobs.json "speaker" field is
  # unused dead data - the real speaker there comes from a whole-run robot
  # choice (GOPOD_GOLDEN_ROBOT/the `robot` argument), not per-step. Offering
  # this prompt for bait would silently write a field with zero playback
  # effect - excluded on that same honesty basis, not a missing feature.
  local robot_pick_song_dir=""
  case "$song" in
    nap|bingo) robot_pick_song_dir="$song_dir_export" ;;
  esac
  if [ -n "$robot_pick_song_dir" ]; then
    local pick_live
    echo "* default to 'n'"
    pick_live="$(_pha0b_prompt_yn "reassign a step's speaker (brobot_1/brobot_2)? y/n " "" y-means-yes)"
    if [ "$pick_live" = "1" ]; then
      local pick_step_id pick_speaker
      read -r -p "step_id to reassign: " pick_step_id
      read -r -p "new speaker (brobot_1/brobot_2): " pick_speaker
      if [ "$pick_speaker" = "brobot_1" ] || [ "$pick_speaker" = "brobot_2" ]; then
        python3 /home/goverlord/.gopod_alias_lib/robot_pick_001.py "$pick_step_id" "$pick_speaker" --yes --knobs "$robot_pick_song_dir/knobs.json"
      else
        echo "PHA0B_ROBOT_PICK_BLOCKED speaker must be brobot_1 or brobot_2"
      fi
    fi
  fi

  # Run-time robot playback filter (STAY_PUT survey/build lineage's own
  # sibling feature - PHA0B_ROBOT_PLAYBACK_FILTER_001.md) - which robots
  # actually fire THIS run, never a file write (robot_pick above is the
  # file-writing tool; this is the opposite: pick-and-don't-touch-disk).
  # Golden-engine songs only (GOPOD_GOLDEN_ROBOT_FILTER is a
  # run_golden_song_001.py-only env var - control/vamp still run on
  # the legacy runners, which never read it). Empty Enter takes the stated
  # default, same shape as tempo-set's own factor prompt and phcal's own
  # sleep_wake robot prompt.
  local robot_filter_choice="both"
  if [ "$runner" = "run_golden_song_001.py" ]; then
    if [ "$song" = "bait" ] && [ "$PHA0B_BAIT_NAV_ACTIVE" = "1" ]; then
      # bait's own bridge already speaks "1"/"2"/"both" natively
      # (_prompt_robot()'s own return contract) - no "0 means both"
      # translation needed, unlike the old prompt below.
      case "$PHA0B_BAIT_ROBOT_FILTER" in
        1|2|both) robot_filter_choice="$PHA0B_BAIT_ROBOT_FILTER" ;;
        *) echo "PHA0B_ROBOT_FILTER_BLOCKED invalid_choice choice=$PHA0B_BAIT_ROBOT_FILTER"; return 1 ;;
      esac
    else
      local rf_input
      echo "** anything entered outside selection range = exit **"
      echo "* default to '0'"
      rf_input="$(_pha0b_prompt_choice "run which robots? 1 / 2 / 0 for both [default]: " "1 2 0" "0" "PHA0B_ROBOT_FILTER_BLOCKED")" || return 1
      case "$rf_input" in
        1) robot_filter_choice="1" ;;
        2) robot_filter_choice="2" ;;
        0) robot_filter_choice="both" ;;
      esac
    fi
  fi

  # 2026-10-17 (operator: "make TOUCH_INTERRUPT optional y/n"): a song whose
  # knobs.json turns the Brobots' touch interrupt off for a live run asks first.
  local touch_off_choice=""
  if [ "$runner" = "run_golden_song_001.py" ] && [ "$live_gate" -eq 1 ] && [ -n "$song_dir_export" ] && \
     python3 -c "
import sys
sys.path.insert(0, '$tools_dir')
from knobs_envelope_001 import load_knobs_envelope
_, d = load_knobs_envelope('$song_dir_export')
sys.exit(0 if d.get('touch_interrupt_off') else 1)
" 2>/dev/null; then
    echo "* default to 'y'"
    if [ "$(_pha0b_prompt_yn "turn touch interrupt off for this run? y/n " y y-means-yes)" = "1" ]; then
      touch_off_choice="1"
    else
      touch_off_choice="0"
    fi
  fi

  # --- Assemble exports and dispatch to the picked runner. ---

  # AWAKEN_MULTI_MODE_TWO_PASS_SURVEY_PLAN_001.md: awaken (song=bait) is
  # synthesize_speaker=True in SONG_REGISTRY (run_golden_song_001.py) -
  # its own knobs.json "speaker" field is dead, the WHOLE run targets one
  # robot (GOPOD_GOLDEN_ROBOT, default "1"). "both" never got a second
  # meaning for this song family - it silently fell back to robot 1 only.
  # Real fix: when song=bait and robot_filter_choice=both, dispatch TWICE
  # - brobot 1's full pass, then brobot 2's full pass - instead of once.
  # Every other song, and every other robot_filter choice for bait itself
  # ("1"/"2"), takes the single empty-string pass below, byte-identical to
  # before this change.
  local -a _golden_robot_passes
  if [ "$song" = "bait" ] && [ "$robot_filter_choice" = "both" ]; then
    _golden_robot_passes=("1" "2")
  else
    _golden_robot_passes=("")
  fi

  local status=0
  local _pass_num=0
  for _golden_robot_pass in "${_golden_robot_passes[@]}"; do
    _pass_num=$((_pass_num + 1))
    if [ "$_pass_num" -gt 1 ]; then
      # Settle gap between the two sequential passes - each pass does its
      # own manage_control=False assume-once/release-once around its own
      # connect..exit range. Firing pass 2's assume immediately after pass
      # 1's release risks the exact no-settle no-op class of bug CLAUDE.md's
      # own TECHNICAL GOTCHAS section already names (a fresh
      # assume_behavior_control needs a settle before the first action) -
      # same 1.5s convention as PHCAL_ASSUME_SETTLE_SECONDS/
      # POST_ASSUME_SAY_SETTLE_SECONDS, reused here, not a new number.
      echo "PHA0B_BAIT_MULTI_PASS_SETTLE sleeping 1.5s before brobot 2's pass"
      sleep 1.5
    fi

    local stderr_capture
    stderr_capture="$(mktemp)"
    # pha0b_run.log (2026-08-15, MASTER_TWEAKS_STAGE2_LOGS_001.md) - same
    # tee pattern already proven twice in this file (test-reaction-in-the-
    # beat/test-anim-*), scoped to this return-free dispatch subshell only,
    # never wrapped around pha0b()'s own body (which has early `return 1`s
    # scattered through its own validation/prompt logic above - a `return`
    # inside a piped subshell only exits that subshell, never the real
    # function, so wrapping the whole function would silently break every
    # one of those). Lives in ~/.gopod_alias_lib/, no git repo at all in
    # that path - provably outside git, never committable. Overwritten
    # fresh each pass (matching phcal_run.log's own convention) - on the
    # two-pass awaken route, pass 2's own log overwrites pass 1's; each
    # pass's REAL run log/json (runs/golden_run_<timestamp>.*) is
    # unaffected, this is only the scratch tee copy.
    # PIPESTATUS[0], not $?, since $? after a pipe reflects tee's own exit
    # status, not the subshell's - the BLOCKED/FAILED detection below
    # depends on the real one.
    (
      cd "$tools_dir" && \
      export "${env_prefix}_FROM=$point_a" && \
      export "${env_prefix}_TO=$point_b" && \
      export "${env_prefix}_WAKE=1" && \
      if [ -n "$song_dir_export" ]; then export GOPOD_CONTROL_SONG_DIR="$song_dir_export"; fi && \
      if [ -n "$song_dir_export" ] && [ "$runner" = "run_golden_song_001.py" ]; then export GOPOD_GOLDEN_SONG_DIR="$song_dir_export"; fi && \
      if [ "$song" != "bingo" ]; then export GOPOD_CONTROL_SONG_ROBOT="$robot"; fi && \
      if [ -n "$_golden_robot_pass" ]; then
        export GOPOD_GOLDEN_ROBOT="$_golden_robot_pass"
      elif [ "$runner" = "run_golden_song_001.py" ] && [ "$robot_filter_choice" != "both" ]; then
        export GOPOD_GOLDEN_ROBOT="$robot_filter_choice"
      fi && \
      if [ "$runner" = "run_golden_song_001.py" ]; then export GOPOD_GOLDEN_ROBOT_FILTER="$robot_filter_choice"; fi && \
      if [ -n "$touch_off_choice" ]; then export GOPOD_GOLDEN_TOUCH_INTERRUPT_OFF="$touch_off_choice"; fi && \
      if [ "$_pass_num" -lt "${#_golden_robot_passes[@]}" ]; then export GOPOD_GOLDEN_SUPPRESS_PROMOTE=1; fi && \
      if [ "$live_gate" -eq 1 ]; then export GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1; fi && \
      if [ -n "$rich_display_override" ]; then export GOPOD_CONSOLE_RICH_DISPLAY="$rich_display_override"; fi && \
      python3 "$runner"
    ) 2> >(tee "$stderr_capture" >&2) | tee "/home/goverlord/.gopod_alias_lib/pha0b_run.log"
    status="${PIPESTATUS[0]}"

    local blocked_line=""
    if [ "$status" -ne 0 ]; then
      blocked_line="$(grep -o 'BLOCKED:.*' "$stderr_capture" | tail -1)"
    fi
    rm -f "$stderr_capture"
    # A failed/blocked first pass stops here - brobot 2's pass never fires
    # on top of a failure (matching the single-dispatch original: a failure
    # was always terminal for that one dispatch; running a second pass
    # after one already failed isn't assumed safe here). `break`, not
    # `return`, so the rich-display finalization below - which the
    # original always ran regardless of success/failure - still fires.
    if [ "$status" -ne 0 ]; then
      break
    fi
  done


  if [ "$status" -ne 0 ]; then
    if [ -n "$blocked_line" ]; then
      echo "PHA0B_BLOCKED $blocked_line"
    else
      echo "PHA0B_FAILED unexpected error (see trace above)"
    fi
    return 1
  fi
}

# pha0b_menu - read-only picker that sits IN FRONT of pha0b's own slice
# engine above (PHA0B MENU RUNG 1, gopod_notes/PHA0B_MENU_RUNG1_001.md).
# Fires automatically when pha0b is called with zero arguments. Lists
# every song directory that actually exists on disk
# (goverlord/runtime/songs/*/), lets the operator pick one, then reads
# THAT song's own knobs.json for existing "section" tags - bingo has
# them (OPENING SYNC / INTERLEAVED BANTER / ROUND 1-3 / CLOSING); most
# other songs don't, and this never invents sections for a song that
# doesn't declare them - shown honestly as "no divisions - step-id mode
# only," with the raw step list offered instead. Operator picks Point A
# and Point B off whichever list printed; this PRINTS the exact
# `pha0b <song> <a> <b>` command and STOPS - it never runs it, never
# writes anything, never touches a score/knobs file. A song picked here
# that pha0b's own case statement below doesn't wire up yet (bare interview/preshow, or
# brobots_bait_001 - all three share the same reason: they're built on the interview's
# own line-based exchange shape, not a robot_control_song_001-family score, so there's no
# standalone step-loop runner to attach a slice to; see start-the-net-song above for
# brobots_bait_001's own real, non-pha0b entry point) is refused here too, same as pha0b
# itself would refuse it - this menu reuses pha0b's own song-keyword mapping rather than
# re-deriving new capability.
# --- pha0b_menu(): disk-scan song picker + point A/B selection --------------
pha0b_menu() {
  local songs_dir="/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs"

  # PHA0B_PHCAL_NAV_CONSOLIDATION_SURVEY_001.md §6 / PHA0B_PHCAL_CODE_PATH_
  # AUDIT_001.md: this whole body now runs inside an outer loop so the
  # reserved "PLAYHEAD Calibrations" menu slot below can return here after
  # phcal exits, instead of falling all the way back to the shell prompt.
  # Every EXISTING exit path (0/invalid input/a completed song run) is a
  # plain `return` - bash `return` always exits the enclosing FUNCTION
  # regardless of loop nesting, so wrapping the body changes nothing about
  # any of those paths; only the new phcal branch uses `continue` instead
  # of `return`, deliberately, to get the "come back here" behavior.
  while true; do
  local -a dir_array
  local i=1
  echo "PHA0B_MENU songs on disk:"
  for d in "$songs_dir"/*/; do
    local name
    name="$(basename "$d")"
    # zzz_archives is a container folder, not a song - excluded 2026-07-24
    # (POST_MOVE_CLEANUP_AND_INTERVIEW_SURVEY_001.md flagged the same "archives
    # listed as if it were a song" issue under the old "archives" name).
    [ "$name" = "zzz_archives" ] && continue
    # tools/ is the song-tools folder (run_golden_song_001.py and friends),
    # not a song - same reasoning as zzz_archives above. Confirmed 2026-08-16
    # (TOOLS_FOLDER_SURVEY_001.md): picking it here always blocked
    # (PHA0B_MENU_REFUSED, no playhead wired for a non-song folder), so it's
    # excluded from the list entirely rather than listed-then-refused.
    # tools/ itself is untouched on disk - see songs/tools/README.md for how
    # to actually run each tool (its own dedicated alias, not from here).
    [ "$name" = "tools" ] && continue
    dir_array[$i]="$name"
    echo "  $i. $name"
    i=$((i + 1))
  done
  local dir_count=$((i - 1))
  if [ "$dir_count" -eq 0 ]; then
    echo "PHA0B_MENU_BLOCKED no_songs_found dir=$songs_dir"
    return 1
  fi

  # PLAYHEAD Calibrations - reserved, always-last menu slot (survey/audit
  # cited above). "$phcal_slot" is derived from dir_count every pass, so it
  # auto-shifts if a song is ever added/removed - never a hardcoded number.
  # Every existing song index 1..dir_count is untouched, nothing renumbered.
  local phcal_slot=$((dir_count + 1))
  echo "  $phcal_slot. PLAYHEAD Calibrations (PHCAL.md)"

  # Song Notation / Score - second reserved, always-last menu slot
  # (NOTATION_GOLDEN_GUTS_GAP_SURVEY_001.md), added right after phcal's own
  # slot so it stays the new last item. Same derived-not-hardcoded shape.
  local notation_slot=$((dir_count + 2))
  echo "  $notation_slot. Song Notation / Score"

  echo "Please note:"
  echo "  see songs/tools/README.md for the song-tools - run via their own aliases, not from here"

  echo "** anything entered outside selection range will exit **"
  local dir_choice
  dir_choice="$(_pha0b_prompt_range_choice "Pick a song [1-$notation_slot, 0 to exit]: " "$notation_slot" "PHA0B_MENU_BLOCKED")" || return 1

  if [ "$dir_choice" -eq "$phcal_slot" ]; then
    # Same subprocess call phcal()'s own guided-flow branch already makes
    # (brobots.sh's `phcal()`, bare-args case) - not a new code path, one
    # more door into it. `continue` (not `return`) is the one real behavior
    # change this build adds - see the outer `while true; do` comment above.
    ( cd "$(dirname "${BASH_SOURCE[0]}")" && python3 phcal_isolate_001.py )
    continue
  fi

  if [ "$dir_choice" -eq "$notation_slot" ]; then
    # Routes into the existing score()/score-save()/_score_song_dir()
    # plumbing (below) rather than reimplementing scoring - this menu only
    # picks song/mode/destination and calls them. Same "continue, don't
    # return" shape as the phcal slot above.
    _pha0b_notation_flow
    continue
  fi

  local chosen_dir="${dir_array[$dir_choice]}"

  # --- Resolve the chosen folder name to pha0b's own song keyword. ---
  local song=""
  # 2026-10-10: a song folder is known by its song_id, not its name
  # (song_folders_001.py), so renamed or renumbered folders still map.
  case "$(_song_id_of_dir "$songs_dir/$chosen_dir")" in
    brobots_bingo) song="bingo" ;;
    brobots_baby_robots_sleep) song="nap" ;;
  esac
  case "$chosen_dir" in
    # 01b_brobots_bingo_game - NOT a note-sequence song, no playhead A/B slice
    # applies (2026-08-10, operator direction). This is the live, voice/touch-
    # triggered gobingo GAME itself (goverlord/runtime/songs/01b_brobots_bingo_game/, a standalone Go
    # binary) - distinct from 01a_brobots_bingo_test, the scripted performance/
    # capture song. Smallest possible special-case: picking this entry calls
    # the existing, already-working gobingo() function (demo.sh) verbatim -
    # same launch (serial/locale/silent flags, reactor co-launch) every other
    # gobingo caller already uses, not a reimplementation. Bypasses the
    # song="..."/point-A-B flow entirely below since there is no step
    # sequence to slice. Unlike every other pha0b entry, gobingo has no
    # dry-by-default gate of its own - it always fires the real binary
    # against real hardware (--silent only suppresses speech, not action) -
    # left as-is, not newly gated, per "reuse the existing launch verbatim."
    01b_brobots_bingo_game)
      local bingo_game_run_choice
      echo "** default to '1' **"
      read -r -p "bingo game run 1 (continuous) or run 2 (pause for backpack rub) [1/2]: " bingo_game_run_choice
      local bingo_game_grid_choice
      echo "** default to '75' **"
      read -r -p "bingo game grid size, 75 or 90 [default 75]: " bingo_game_grid_choice
      bingo_game_grid_choice="${bingo_game_grid_choice:-75}"
      if [ "$bingo_game_grid_choice" != "75" ] && [ "$bingo_game_grid_choice" != "90" ]; then
        echo "PHA0B_BINGO_GAME_BAD_GRID value=$bingo_game_grid_choice - defaulting to 75"
        bingo_game_grid_choice="75"
      fi
      if [ "$bingo_game_run_choice" = "2" ]; then
        gobingo --pause-for-touch --grid-size "$bingo_game_grid_choice"
      else
        gobingo --grid-size "$bingo_game_grid_choice"
      fi
      return $?
      ;;
    *)
      # A song already matched by its song_id above goes on.
      if [ -z "$song" ]; then
        echo "PHA0B_MENU_REFUSED song_dir=$chosen_dir no playhead wired for this song yet under pha0b's own song keyword list (bingo|nap) - see pha0b()'s case statement above"
        return 1
      fi
      ;;
  esac

  # --- PHA0B_GOLDEN_NAV_RIPREPLACE_AND_SONG1_SCOPE_001.md Round 1: bait  ---
  # --- (00a_brobots_awaken) pilot. phcal's golden nav replaces EVERY one ---
  # --- of this song's own pre-run touchpoints (Point A/B, reporter-gap, ---
  # --- live/dry, phcal-apply y/n, robot filter, rich-display), all in   ---
  # --- ONE subprocess call - never touches the shared _pha0b_prompt_*/  ---
  # --- live_robots_prompt() functions below, which every other song    ---
  # --- still calls exactly as before. pha0b() itself gains matching    ---
  # --- bait-only bypass branches at each of its own 4 prompts (see that ---
  # --- function) so the SAME shared dispatch/run logic still fires -   ---
  # --- only the nav feeding it changed. Skips this function's own      ---
  # --- MODE=STEPS/SECTIONS + Point A/B + reporter-gap logic entirely -  ---
  # --- untouched below for every song that isn't bait.
  if [ "$song" = "bait" ]; then
    local bait_nav_out bait_nav_status
    bait_nav_out="$( cd "$(dirname "${BASH_SOURCE[0]}")" && python3 phcal_isolate_001.py --pha0b-bait-nav )"
    bait_nav_status=$?
    echo "$bait_nav_out"
    local bait_result_line
    bait_result_line="$(echo "$bait_nav_out" | grep '^PHA0B_BAIT_NAV_RESULT ')"
    if [ "$bait_nav_status" -ne 0 ] || [ -z "$bait_result_line" ]; then
      echo "PHA0B_MENU_BLOCKED bait nav did not complete"
      return 1
    fi
    local point_a point_b
    point_a="$(echo "$bait_result_line" | sed -n 's/.*point_a=\([^ ]*\).*/\1/p')"
    point_b="$(echo "$bait_result_line" | sed -n 's/.*point_b=\([^ ]*\).*/\1/p')"
    local bait_gap_seconds
    bait_gap_seconds="$(echo "$bait_result_line" | sed -n 's/.*reporter_gap_seconds=\([^ ]*\).*/\1/p')"
    echo "PHA0B_MENU_REPORTER_GAP override=$bait_gap_seconds - reporter gaps in $point_a..$point_b will pause ${bait_gap_seconds}s (this shell session only, until unset)"
    export GOPOD_BINGO_REPORTER_GAP_OVERRIDE="$bait_gap_seconds"
    export GOPOD_GOLDEN_REPORTER_GAP_OVERRIDE="$bait_gap_seconds"
    export PHA0B_BAIT_NAV_ACTIVE=1
    export PHA0B_BAIT_LIVE_GATE
    PHA0B_BAIT_LIVE_GATE="$(echo "$bait_result_line" | sed -n 's/.*live_gate=\([^ ]*\).*/\1/p')"
    export PHA0B_BAIT_PHCAL_APPLY
    PHA0B_BAIT_PHCAL_APPLY="$(echo "$bait_result_line" | sed -n 's/.*phcal_apply_choice=\([^ ]*\).*/\1/p')"
    export PHA0B_BAIT_ROBOT_FILTER
    PHA0B_BAIT_ROBOT_FILTER="$(echo "$bait_result_line" | sed -n 's/.*robot_filter=\([^ ]*\).*/\1/p')"
    export PHA0B_BAIT_RICH_DISPLAY
    PHA0B_BAIT_RICH_DISPLAY="$(echo "$bait_result_line" | sed -n 's/.*rich_display=\([^ ]*\).*/\1/p')"
    pha0b "$song" "$point_a" "$point_b"
    local pha0b_status=$?
    unset PHA0B_BAIT_NAV_ACTIVE PHA0B_BAIT_LIVE_GATE PHA0B_BAIT_PHCAL_APPLY PHA0B_BAIT_ROBOT_FILTER PHA0B_BAIT_RICH_DISPLAY
    return $pha0b_status
  fi

  # --- Load knobs.json, compute divisions/sections, pick Point A/B,     ---
  # --- reporter-gap override prompt, then hand off to pha0b() itself.   ---
  local tools_dir="/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools"
  local song_dir="$songs_dir/$chosen_dir"
  # Resolved through the shared resolver (knobs_envelope_001.py) instead of
  # a hardcoded knobs.json - this front door was clean-only
  # (KNOBS_TOUCHERS_MAP_001.md #13) until this pass, so the A/B and section
  # picker the operator sees now reflects whichever file the playback
  # engine would actually use for this song.
  local knobs_path
  knobs_path="$(python3 -c "
import sys
sys.path.insert(0, '$tools_dir')
from knobs_envelope_001 import resolve_active_knobs_path
print(resolve_active_knobs_path('$song_dir'))
")"
  if [ ! -s "$knobs_path" ]; then
    echo "PHA0B_MENU_BLOCKED knobs_missing path=$knobs_path"
    return 1
  fi

  local rows
  rows="$(python3 -c "
import sys
sys.path.insert(0, '$tools_dir')
from knobs_envelope_001 import load_knobs_envelope
_, d = load_knobs_envelope('$song_dir')
steps = d.get('steps', [])
has_sections = any(s.get('section') for s in steps)
if has_sections:
    print('MODE=SECTIONS')
    order = []
    groups = {}
    for s in steps:
        sec = s.get('section')
        if sec not in groups:
            groups[sec] = []
            order.append(sec)
        groups[sec].append(s['step_id'])
    for sec in order:
        ids = groups[sec]
        print(sec + '|' + ids[0] + '|' + ids[-1])
else:
    print('MODE=STEPS')
    for s in steps:
        sid = s['step_id']
        print(sid + '|' + sid + '|' + sid)
")"

  local mode
  mode="$(echo "$rows" | head -1)"
  local data
  data="$(echo "$rows" | tail -n +2)"
  if [ -z "$data" ]; then
    echo "PHA0B_MENU_BLOCKED no_steps_found dir=$chosen_dir"
    return 1
  fi

  if [ "$mode" = "MODE=STEPS" ]; then
    echo "PHA0B_MENU $chosen_dir: no divisions - step-id mode only"
  else
    echo "PHA0B_MENU $chosen_dir divisions:"
  fi
  echo "  0. Keyboard 0 (zero) for full song"

  local -a label_array first_array last_array
  i=1
  while IFS='|' read -r label first last; do
    label_array[$i]="$label"
    first_array[$i]="$first"
    last_array[$i]="$last"
    if [ "$first" = "$last" ]; then
      echo "  $i. $label"
    else
      echo "  $i. $label  (first=$first last=$last)"
    fi
    i=$((i + 1))
  done <<< "$data"
  local div_count=$((i - 1))

  # 2026-07-22: forgiving clamp instead of a hard block for numeric-but-
  # out-of-range input (0, negative, or > div_count) - garbage (non-numeric)
  # input still hard-blocks exactly as before. Point A clamps into
  # [1, div_count]; Point B clamps into [Point A's own chosen index,
  # div_count] - B can never precede A, but B = A (a single division) is
  # allowed. A one-line PHA0B_CLAMP note prints only when a clamp actually
  # changed the typed value.
  #
  # 2026-07-23: typing the literal "0" at Point A is a one-keystroke full-
  # song shortcut - resolves both Point A and Point B to the full range and
  # skips the Point B prompt entirely (no second prompt to answer). Empty
  # Enter, space, or an out-of-range number at Point A still fall through to
  # the Point B prompt as before - only a literal "0" skips it.
  local full_song=0
  local skip_point_b=0
  local a_choice
  read -r -p "Pick Point A [1-$div_count]: " a_choice
  if [ "$a_choice" = "0" ]; then
    full_song=1
    skip_point_b=1
    a_choice=1
  elif [ -z "$a_choice" ]; then
    full_song=1
    a_choice=1
  elif [[ "$a_choice" =~ ^[0-9]+$ ]] && [ "$a_choice" -gt "$div_count" ]; then
    full_song=1
    a_choice=1
  elif ! [[ "$a_choice" =~ ^[0-9]+$ ]]; then
    echo "PHA0B_MENU_BLOCKED invalid_choice choice=$a_choice"
    return 1
  else
    local a_typed="$a_choice"
    if [ "$a_choice" -lt 1 ]; then
      a_choice=1
    elif [ "$a_choice" -gt "$div_count" ]; then
      a_choice="$div_count"
    fi
    if [ "$a_choice" != "$a_typed" ]; then
      echo "PHA0B_CLAMP point_a: typed=$a_typed -> $a_choice"
    fi
  fi

  local b_choice
  if [ "$skip_point_b" -eq 1 ]; then
    b_choice="$div_count"
  else
    read -r -p "Pick Point B [$a_choice-$div_count]: " b_choice
    if [ -z "$b_choice" ]; then
      full_song=1
    elif [[ "$b_choice" =~ ^[0-9]+$ ]] && { [ "$b_choice" -eq 0 ] || [ "$b_choice" -gt "$div_count" ]; }; then
      full_song=1
    elif ! [[ "$b_choice" =~ ^[0-9]+$ ]]; then
      echo "PHA0B_MENU_BLOCKED invalid_choice choice=$b_choice"
      return 1
    else
      local b_typed="$b_choice"
      if [ "$b_choice" -lt "$a_choice" ]; then
        b_choice="$a_choice"
      elif [ "$b_choice" -gt "$div_count" ]; then
        b_choice="$div_count"
      fi
      if [ "$b_choice" != "$b_typed" ]; then
        echo "PHA0B_CLAMP point_b: typed=$b_typed -> $b_choice"
      fi
    fi
  fi

  local point_a point_b
  if [ "$full_song" -eq 1 ]; then
    point_a="${first_array[1]}"
    point_b="${last_array[$div_count]}"
  else
    point_a="${first_array[$a_choice]}"
    point_b="${last_array[$b_choice]}"
  fi

  # Widened 2026-07-25 past bingo-only to bingo|bait, mirroring the phcal-apply
  # gate above (REPORTER_GAP_SHARED_SWITCH_SURVEY_001.md /
  # REPORTER_GAP_WIDENED_TO_BAIT_001.md, operator request). A simple OR, not a
  # case statement like the phcal-apply gate above: that gate needed a
  # per-song knobs path attached to each branch, this one doesn't.
  # CORRECTED 2026-08-12 (studio tuning cut 2 step 4,
  # CONTROL_SONG_GOLDEN_CUTOVER_PREP_002.md) - was stale: this comment used
  # to claim bait's own pause steps read GOPOD_BINGO_REPORTER_GAP_OVERRIDE
  # via run_robot_control_song_001.py's "pause" branch. Untrue since bait's
  # own 2026-08-07 cutover - bait runs on the golden engine now, whose pause
  # branch reads a DIFFERENT env var (GOPOD_GOLDEN_REPORTER_GAP_OVERRIDE),
  # never exported here, so this switch was silently disconnected for bait
  # (and now for control too, added to the gate below - both just cut over
  # the same way). Fixed by exporting BOTH names below: the golden engine
  # gets the name it actually reads, and run_robot_control_song_001.py (the
  # intact fallback for every one of these 3 songs) still gets the name it's
  # always read, so a revert doesn't re-break this switch either way. (The
  # "weather" keyword itself was dropped 2026-08-16 - dead reference to a
  # decluttered folder, same purge as start-the-weather-song - so it never
  # reaches this gate anymore.)
  if [ "$song" = "bingo" ] || [ "$song" = "bait" ] || [ "$song" = "control" ]; then
    local gap_live
    echo "* default to 'y'"
    gap_live="$(_pha0b_prompt_yn "apply reporter gaps to this selected range? y/n " "" n-means-no)"
    if [ "$gap_live" = "0" ]; then
      export GOPOD_BINGO_REPORTER_GAP_OVERRIDE=0
      export GOPOD_GOLDEN_REPORTER_GAP_OVERRIDE=0
      echo "PHA0B_MENU_REPORTER_GAP override=0 - reporter gaps in $point_a..$point_b will pause 0s (this shell session only, until unset)"
    else
      local gap_seconds
      read -r -p "reporter gaps in seconds? [default = 0] " gap_seconds
      if [ -z "$gap_seconds" ]; then
        gap_seconds=0
      elif ! printf '%s' "$gap_seconds" | grep -Eq '^[0-9]+(\.[0-9]+)?$'; then
        echo "PHA0B_MENU_REPORTER_GAP invalid input '$gap_seconds' - please enter a number"
        read -r -p "reporter gaps in seconds? [default = 0] " gap_seconds
        if [ -z "$gap_seconds" ]; then
          gap_seconds=0
        elif ! printf '%s' "$gap_seconds" | grep -Eq '^[0-9]+(\.[0-9]+)?$'; then
          echo "PHA0B_MENU_REPORTER_GAP still invalid ('$gap_seconds') - falling back to 0"
          gap_seconds=0
        fi
      fi
      export GOPOD_BINGO_REPORTER_GAP_OVERRIDE="$gap_seconds"
      export GOPOD_GOLDEN_REPORTER_GAP_OVERRIDE="$gap_seconds"
      echo "PHA0B_MENU_REPORTER_GAP override=$gap_seconds - reporter gaps in $point_a..$point_b will pause ${gap_seconds}s (this shell session only, until unset)"
    fi
  fi

  pha0b "$song" "$point_a" "$point_b"
  return $?
  done
}

# phcal - calibration bench, Rung 3 - a sibling to pha0b: pha0b plays the
# SCORE (dry navigation of a song slice, never fires live by itself); phcal
# tunes the INSTRUMENT (one isolated mechanical primitive, live, on one
# robot). Rung 1 isolated and played current behavior only, no value
# changed. Rung 2 added cycles/hold/speed as adjustable per-call flags.
# Rung 3 adds two more things on top, both unchanged in rung 2's own
# direct-flag behavior:
#   - A guided prompt flow: bare `phcal` (no primitive at all) asks
#     arm-or-nod, then a robot, then walks each value pre-filled with the
#     last-used one (Enter keeps it, typing overrides), then fires exactly
#     like the direct-flag form and saves what was used.
#   - Write-back memory: both the guided flow and the direct-flag form
#     (phcal arm 1 --hold ... etc, still works unchanged) save the values
#     they actually used to phcal_last.json, phcal's OWN memory file, so
#     either entry path updates the same one [last] the other reads from
#     next time. See PHCAL_RUNG1_ISOLATE_AND_WATCH_001.md,
#     PHCAL_RUNG2_TUNING_001.md, PHCAL_RUNG3_GUIDED_FLOW_001.md.
#
# Fires exactly one of:
#   phcal                                          - guided prompt flow
#   phcal arm <robot> [--hold S] [--speed N] [--cycles N]
#       - one or more arm cues: rest -> established up -> rest, repeated
#         <cycles> times (default 1)
#   phcal nod <robot> [count] [--hold S] [--speed N]
#       - one or more nods: down -> up (source order, kept as rung 1
#         confirmed it - see the down/up note below), repeated [count]
#         times (default 1)
# robot: 1 or 2, same convention as pha0b's own control/bait robot
# argument. Flags may appear in any order after robot (and after count, for
# nod). One robot per call, no two-robot sync - unchanged from rung 1.
#
# Nod direction: the real source (run_nod(), run_songs_runner_001.py)
# fires down first, then up - rung 1 confirmed this against the actual
# bingo runner order, correcting an earlier spec's prose that wrongly said
# up-then-down. Kept as down-then-up here; do not revert to the wrong
# prose order.
#
# Sourced from run_songs_runner_001.py's own run_move_axis()/
# run_arm_cue()/run_nod() (the proven single-robot port of
# _brobots_move_axis already in this file) via
# ~/.gopod_alias_lib/phcal_isolate_001.py - same HTTP call shape, same
# rung-1 defaults (arm_cue: cycles=1, hold_seconds=1.2, speed=2; nod:
# hold_seconds=0.35, speed=2), now all three overridable per flag or via the
# guided flow, reproduced with a printed, timestamped line in front of every
# sub-instruction (assume/move/hold/stop/release) so the operator can watch
# each step land, whatever values (default, flag-overridden, or guided-flow
# entered) are actually in play for that call. Does not reuse
# _brobots_move_axis above directly - that helper fires BOTH robots every
# call and carries no live-fire gate at all; phcal is explicitly
# one-robot-at-a-time and gated (see below), so the Python single-serial
# port is the correct proven reference here, not that bash helper.
#
# _score_song_dir - shared song-keyword resolver backing score()/score-save()
# below, same one-helper-many-callers shape as _test_anim_isolated further
# down this file. Deliberately reuses pha0b()'s own case statement keywords
# verbatim (bingo/control/bait/vamp/mixup/nap/interview/preshow)
# rather than inventing a second vocabulary - read directly from pha0b()'s
# case block above. mixup (102_brobots_cross_persona) and nap
# (101_brobots_nap, formerly 104_brobots_baby_robots_sleep) added 2026-08-10
# (SHEET_MUSIC_GOLDEN_PROCESS_SURVEY_001.md finding #2 - both are on the
# current live shelf and both print correctly, but had no keyword here at
# all). Two keywords pha0b accepts are still refused here even though pha0b
# itself accepts them: interview (song_id brobots_interview_section_01, now
# housed in 02b_brobots_interview_run/) and preshow
# (02a_brobots_interview_vamp/, repointed 2026-08-19 - confirmed via gopod-vamp()'s own
# load_preshow_song() call) both tested clean (exit 0) against
# print_song_score_001.py, but their knobs.json steps carry no note/TEXT the
# way this printer's field-complete design expects - every step prints
# TEXT:(blank) even where the song plainly has real spoken lines, which is
# misleading rather than merely incomplete. bingo/control/bait/vamp/mixup/nap
# all confirmed to print real TEXT/note content.
# On success: echoes the resolved song directory to stdout only (safe to
# command-substitute). On failure: prints SCORE_REFUSED to stderr, returns 1,
# nothing on stdout.
# === SONG FOLDERS - found by song_id (2026-10-10) ===========================
# Operator: "even shuffling (song) folders without notice should be easy to
# remap". A song is known by the "song_id" in its knobs.json; its folder under
# goverlord/runtime/songs/ can be renamed or renumbered freely. Same rule as
# goverlord/runtime/songs/tools/song_folders_001.py (which this calls).
_song_dir_by_id() {
  python3 /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools/song_folders_001.py "$1"
}
_song_id_of_dir() {
  python3 -c 'import json,sys; print(json.load(open(sys.argv[1]+"/knobs.json")).get("song_id",""))' "$1" 2>/dev/null
}

# === SCORE - score/score-save: the bare notation-page printer =============
_score_song_dir() {
  local song="$1"
  local songs_dir="/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs"
  case "$song" in
    bingo)   _song_dir_by_id brobots_bingo ;;
    nap)     _song_dir_by_id brobots_baby_robots_sleep ;;
    *)
      echo "SCORE_REFUSED song=$song unknown - must be one of: bingo nap" >&2
      return 1
      ;;
  esac
}

# score - bare one-word notation-page printer, screen only. No paths, no
# flags. Bare call defaults to brobots_bingo (the song currently being
# worked). Read-only against song data - print_song_score_001.py itself is
# never modified or written to by this alias, only invoked.
# $2 (optional, default "field-dump"): "field-dump" (unchanged, existing
# behavior for every pre-existing caller that passes 0 or 1 args) or
# "narrative" (NOTATION_GOLDEN_GUTS_GAP_SURVEY_001.md's own script-read
# mode, print_song_score_001.py --narrative) - added 2026-08-30 so the new
# pha0b notation slot below can reuse this alias instead of reimplementing
# it, per that build's own instruction.
score() {
  local song="${1:-bingo}"
  local mode="${2:-field-dump}"
  local song_dir
  song_dir="$(_score_song_dir "$song")" || return 1
  local narrative_flag=""
  [ "$mode" = "narrative" ] && narrative_flag="--narrative"
  ( cd "/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools" && \
    python3 print_song_score_001.py "$song_dir" $narrative_flag )
}

# score-save - same song vocabulary as score() above, writes the page to
# that song's own notation/ directory instead of the screen (created if
# absent - sibling to that song's runs/ directory, same distinction
# NOTATION_PAGE_BUILD_001.md proposed: runs/ is machine-generated run
# output, notation/ is this operator-facing printed artifact). House dated
# filename convention. Prints the path it wrote, nothing else.
# $2 (optional, default "field-dump"): same mode switch as score() above,
# same reasoning - "narrative" saves a script-read instead of a field-dump,
# named accordingly in the output filename.
score-save() {
  local song="${1:-bingo}"
  local mode="${2:-field-dump}"
  local song_dir
  song_dir="$(_score_song_dir "$song")" || return 1
  local notation_dir="$song_dir/notation"
  mkdir -p "$notation_dir"
  local ts
  ts="$(date +%Y%m%d_%H%M%S)"
  local narrative_flag="" name_part="score"
  if [ "$mode" = "narrative" ]; then
    narrative_flag="--narrative"
    name_part="narrative"
  fi
  local out_file="$notation_dir/${song}_${name_part}_${ts}.txt"
  ( cd "/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools" && \
    python3 print_song_score_001.py "$song_dir" $narrative_flag ) > "$out_file"
  echo "SCORE_SAVED $out_file"
}

# _pha0b_notation_flow - the "Song Notation / Score" reserved menu slot's
# own sub-flow (pha0b_menu() above). Picks song/mode/destination, then
# calls score()/score-save() above exactly as if typed by hand - this
# function never touches knobs.json/story.md or print_song_score_001.py
# itself, it's pure menu glue over already-working plumbing
# (NOTATION_GOLDEN_GUTS_GAP_SURVEY_001.md's own instruction). Lists the
# SAME full vocabulary _score_song_dir() already accepts, interview/preshow
# included on purpose - picking either shows that function's own existing
# SCORE_REFUSED message (schema not supported) instead of a misleading
# blank sheet or a crash, exactly as asked; nothing new to build for that
# case, the refusal already exists.
_pha0b_notation_flow() {
  local notation_songs="bingo nap"
  local -a notation_song_array
  local i=1
  echo "PHA0B_NOTATION songs:"
  local kw
  for kw in $notation_songs; do
    notation_song_array[$i]="$kw"
    echo "  $i. $kw"
    i=$((i + 1))
  done
  local notation_song_count=$((i - 1))
  echo "** anything entered outside selection range = exit **"
  local song_choice
  song_choice="$(_pha0b_prompt_range_choice "Pick a song [1-$notation_song_count, 0 to exit]: " "$notation_song_count" "PHA0B_NOTATION_BLOCKED")" || return 1
  local notation_song="${notation_song_array[$song_choice]}"

  # _score_song_dir() itself prints SCORE_REFUSED (with the real reason) to
  # stderr and returns 1 for interview/preshow - surfaced here as-is, no
  # new message invented, no attempt to run the printer against a
  # song it already knows it can't read.
  _score_song_dir "$notation_song" > /dev/null || return 0

  echo "* default to 'field-dump'"
  local mode_choice
  mode_choice="$(_pha0b_prompt_choice "mode: field-dump / narrative [default field-dump]: " "field-dump narrative" "field-dump" "PHA0B_NOTATION_BLOCKED")" || return 1

  echo "* default to 'screen'"
  local dest_choice
  dest_choice="$(_pha0b_prompt_choice "screen or save: screen / save [default screen]: " "screen save" "screen" "PHA0B_NOTATION_BLOCKED")" || return 1

  if [ "$dest_choice" = "save" ]; then
    score-save "$notation_song" "$mode_choice"
  else
    score "$notation_song" "$mode_choice"
  fi
}

# phcal()'s guided flow (bare `phcal`, no primitive) no longer calls
# live_robots_prompt() - removed 2026-08-21, PHCAL_NAV_BASE_RULES_FIXES_
# 001.md (PHCAL_NAV_BASE_RULES_SURVEY_001.md §4 finding 1): it was a
# redundant shell-side gate asked BEFORE the operator even saw the tree,
# semantically overlapping with phcal_isolate_001.py's own mode-detect/
# confirm screen one step later. GOPOD_ALLOW_LIVE_ROBOT_SPEECH is now
# decided entirely inside that screen instead (_resolve_session_mode_once(),
# phcal_isolate_001.py) - multi/single (continue) = live, none/dry-run =
# dry. The direct-flag CLI form below (`phcal arm 1 ...`) is UNCHANGED -
# it never reaches that Python-side screen at all (main(), not
# run_guided_flow()), so it still needs its own gate and still calls
# live_robots_prompt() exactly as before.
# === PHCAL BENCH - phcal()/tempo-set(): the calibration bench, one ========
# === isolated mechanical primitive at a time, live, on one robot. =========
phcal() {
  local primitive="$1"

  if [ -z "$primitive" ]; then
    # No primitive at all -> rung 3's guided prompt flow. Handled entirely
    # inside phcal_isolate_001.py (arm-or-nod, robot, then each value,
    # pre-filled from phcal_last.json) so there is exactly one place that
    # reads/writes that memory file, not a bash-side copy of the same logic.
    # No live-gate export here - the Python side's own mode-confirm screen
    # decides GOPOD_ALLOW_LIVE_ROBOT_SPEECH now, see the header comment above.
    ( cd "$(dirname "${BASH_SOURCE[0]}")" && python3 phcal_isolate_001.py )
    return $?
  fi

  local -a live_env=()
  local live_gate
  live_gate="$(live_robots_prompt)"
  if [ "$live_gate" = "2" ]; then
    # 2026-08-19 refinement (PHCAL_NAV_POLISH_001.md addendum): ESC at the
    # live-robots prompt still means a full abort - phcal never starts.
    echo "PHCAL_ABORTED (ESC at live-robots prompt) - exiting, phcal did not start"
    return 1
  fi
  if [ "$live_gate" = "1" ]; then
    live_env=(GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1)
  fi

  local robot="$2"
  if [ -z "$robot" ]; then
    echo "PHCAL_USAGE phcal <arm|nod|rattle> <robot 1|2> [reps] [--hold S] [--speed N] [--cycles N] [--volume N]"
    echo "  arm:     phcal arm <robot> [--hold S] [--speed N] [--cycles N]  - one or more arm cues (rest->up->rest), reps=cycles, hold=hold between reps"
    echo "  nod:     phcal nod <robot> [reps] [--hold S] [--speed N]       - one or more nods (default 1 rep), hold=hold between reps"
    echo "  rattle:  phcal rattle <robot> [--volume N]                     - one rattle, volume 1-5 (default 5)"
    echo "  (--cycles is arm-only; nod repeats via the positional [reps] instead)"
    echo "  (bare 'phcal' with no arguments launches the guided prompt flow instead)"
    return 1
  fi

  shift 2
  ( cd "$(dirname "${BASH_SOURCE[0]}")" && \
    if [ "${#live_env[@]}" -gt 0 ]; then export "${live_env[@]}"; fi && \
    python3 phcal_isolate_001.py "$primitive" "$robot" "$@" )
}

# tempo-set - phcal-adjacent guided flow for Tempo Phase 2
# (TEMPO_BUFFER_KNOB_SURVEY_001.md/TEMPO_BUFFER_KNOB_PHASE1_EXECUTED_001.md/
# TEMPO_SET_TOOL_EXECUTED_001.md), thin wrapper around tempo_set_001.py -
# same shape as phcal()'s own delegation to phcal_isolate_001.py above: this
# function only picks the song directory and asks the two mode questions,
# every actual read/write happens inside the one Python tool. Reuses
# pha0b_menu()'s own disk-scan song list (goverlord/runtime/songs/*,
# zzz_archives excluded) rather than inventing a second song vocabulary.
# Name checked against every existing alias/function/pha0b keyword before
# adding - no collision (TEMPO_SET_TOOL_EXECUTED_001.md).
tempo-set() {
  local songs_dir="/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs"
  local -a dir_array
  local i=1
  echo "TEMPO_SET songs on disk:"
  for d in "$songs_dir"/*/; do
    local name
    name="$(basename "$d")"
    [ "$name" = "zzz_archives" ] && continue
    dir_array[$i]="$name"
    echo "  $i. $name"
    i=$((i + 1))
  done
  local dir_count=$((i - 1))
  local dir_choice
  read -r -p "Pick a song [1-$dir_count]: " dir_choice
  if ! [[ "$dir_choice" =~ ^[0-9]+$ ]] || [ "$dir_choice" -lt 1 ] || [ "$dir_choice" -gt "$dir_count" ]; then
    echo "TEMPO_SET_BLOCKED invalid_choice choice=$dir_choice"
    return 1
  fi
  local chosen_dir="${dir_array[$dir_choice]}"
  local knobs_path="$songs_dir/$chosen_dir/knobs.json"
  if [ ! -f "$knobs_path" ]; then
    echo "TEMPO_SET_BLOCKED no knobs.json found for $chosen_dir at $knobs_path"
    return 1
  fi

  echo "Mode:"
  echo "  A. set the song's GLOBAL tempo (whole-song ease)"
  echo "  B. set ONE step's tempo_factor (+ optional comment)"
  local mode_choice
  read -r -p "pick a mode [A/B]: " mode_choice

  ( cd "$(dirname "${BASH_SOURCE[0]}")" && \
    if [ "$mode_choice" = "A" ] || [ "$mode_choice" = "a" ]; then
      local value
      read -r -p "new global_tempo (0.0-9.9): " value
      python3 tempo_set_001.py set-global "$value" --knobs "$knobs_path"
      read -r -p "apply the above? y/n: " apply_choice
      if [ "$apply_choice" = "y" ] || [ "$apply_choice" = "Y" ]; then
        python3 tempo_set_001.py set-global "$value" --yes --knobs "$knobs_path"
      fi
    elif [ "$mode_choice" = "B" ] || [ "$mode_choice" = "b" ]; then
      # Numbered step menu, same shape as pha0b_menu()'s own song list and
      # Mode A's own song list above - reads the ACTIVE (dirty-first
      # resolved) steps array via the same shared knobs_envelope_001.py
      # resolver the write path itself uses, so the numbered list always
      # matches whatever file the write will actually land in. Fixes the old
      # blind "step_id:" prompt (operator typed "4", there is no step
      # numbered 4 - steps have names - TEMPO_SET_BLOCKED, TEMPO_SET_MODE_B_
      # NUMBERED_MENU_001.md).
      local step_ids
      step_ids="$(python3 -c "
import sys
sys.path.insert(0, '/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools')
from knobs_envelope_001 import load_knobs_envelope
_, envelope = load_knobs_envelope('$songs_dir/$chosen_dir')
for s in envelope.get('steps', []):
    print(s.get('step_id'))
")"
      if [ -z "$step_ids" ]; then
        echo "TEMPO_SET_BLOCKED no steps found for $chosen_dir"
        return 1
      fi
      local -a step_array
      local si=1
      echo "steps:"
      while IFS= read -r sid; do
        [ -z "$sid" ] && continue
        step_array[$si]="$sid"
        echo "  $si. $sid"
        si=$((si + 1))
      done <<< "$step_ids"
      local step_count=$((si - 1))
      local step_choice
      read -r -p "pick a step [1-$step_count]: " step_choice
      if ! [[ "$step_choice" =~ ^[0-9]+$ ]] || [ "$step_choice" -lt 1 ] || [ "$step_choice" -gt "$step_count" ]; then
        echo "TEMPO_SET_BLOCKED invalid_choice choice=$step_choice"
        return 1
      fi
      local step_id="${step_array[$step_choice]}"
      local factor comment
      read -r -p "new tempo_factor (default 1.0): " factor
      factor="${factor:-1.0}"
      read -r -p "tempo_comment (optional, Enter to leave unchanged): " comment
      local -a comment_flag=()
      if [ -n "$comment" ]; then comment_flag=(--comment "$comment"); fi
      python3 tempo_set_001.py set-buffer "$step_id" --factor "$factor" "${comment_flag[@]}" --knobs "$knobs_path"
      read -r -p "apply the above? y/n: " apply_choice
      if [ "$apply_choice" = "y" ] || [ "$apply_choice" = "Y" ]; then
        python3 tempo_set_001.py set-buffer "$step_id" --factor "$factor" "${comment_flag[@]}" --yes --knobs "$knobs_path"
      fi
    else
      echo "TEMPO_SET_BLOCKED mode must be A or B"
      return 1
    fi
  )
}

# gopod-conn-test - standalone version of the opening chord's own per-robot
# wake check (_gopod_chord_wirepod_job's WIREPOD_CHAIN_WAKE_* step in
# core.sh already fires this same /api-sdk/conn_test call - this just
# exposes it outside the chord, with no restart/wake/speech attached, for a
# plain "is this robot reachable right now" check). One of the three named
# gaps in ALIAS_REGISTRY_TRUTH_SWEEP_001.md. Argument: "1", "2", or "both"
# (default "both"). Uses the same single _gopod_note_send instrument as
# every note above.
# robot-sleep / robot-wake - fires the real GoToSleep* animation TRIGGERS
# directly over the vendored Go SDK, bypassing Wire-Pod's own dispatch
# entirely (gopod_probes/tools/direct_sdk_robot_sleep_001.go,
# ROBOT_SLEEP_DIRECT_SDK_BUILT_001.md, VECTOR_SLEEP_PETTING_ANIM_PINNED_REFERENCE_001.md).
# Wire-Pod itself can only fire raw one-shot clips (never triggers, and no
# gotosleep-family token exists in animation_vocab.json anyway), so this is
# the only real path to Vector's actual sleep behavior, not just an
# animation playing once. No dry mode - same as direct_sdk_brobots_ready_001,
# this class of tool always fires for real; there is nothing to simulate.
# First argument, both aliases: "0" = both robots (also the default, and
# "both" still works as a spelled-out alias), "1" or "2" for a single robot -
# "0" matches pha0b_menu's own "0 = full song" convention elsewhere in this
# codebase. robot-sleep's second argument: hold-seconds (default 5) - how
# long to keep behavior control open after firing GoToSleepSleeping before
# releasing it. Added 2026-07-25, live-observed need: a first live test with
# no hold at all had both robots wake right back up within moments -
# releasing BehaviorControl right after firing the trigger appears to hand
# control straight back to the robot's own onboard behavior, which does not
# stay asleep on its own. robot-wake fires GoToSleepOff - the nearest
# confirmed exit trigger (no GoToSleepWakeup trigger exists at all); whether
# it actually rouses a robot from GoToSleepSleeping is the live question
# these two aliases exist to let the operator answer by watching, not
# something either alias claims on its own.
_ROBOT_SLEEP_DIRECT_BIN="/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_probes/tools/direct_sdk_robot_sleep_001"

# robot-info - read-only VersionState/ProtocolVersion/BatteryState snapshot,
# direct-SDK, no BehaviorControl needed at all (unlike robot-sleep/robot-wake
# above). Built 2026-08-10 for the Brobot 2 (Pip) intermittent-unresponsiveness
# investigation (BROBOT_2_INSTABILITY_EXTERNAL_AI_BRIEF_001.md) - the concrete
# "settle the firmware version question" step that report's own external-AI
# research confirmed is safe to run. Same which=1|2|0|both convention as
# robot-sleep/robot-wake (_robot_sleep_specs reused verbatim, not a second
# copy). Timeout 30s - the binary's own internal per-RPC deadline is 20s x3
# sequential-worst-case-per-robot, but calls run concurrently across RPCs
# within a robot and across robots, so 30s is a real buffer, not a guess -
# same "belt and suspenders over the binary's own deadline" precedent as
# robot-sleep/robot-wake above, not a replacement for it.
_ROBOT_INFO_DIRECT_BIN="/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_probes/tools/direct_sdk_robot_info_001"

# cube-blip - net-new 2026-08-15, first GOPOD code to touch the cube
# instrument. Fires direct_sdk_cube_blip_001 (connect -> all four corner
# LEDs red -> 2s hold -> all four corners green -> release) via the
# standalone direct-SDK binary, gopod_probes/tools/direct_sdk_cube_blip_001.go
# - see CUBE_DOOR_SURVEY_001.md / CUBE_BLIP_TOOL_BUILT_001.md /
# CUBE_BLIP_ALIAS_PHCAL_WIRED_001.md. No dry mode - same as
# direct_sdk_brobots_ready_001/robot-sleep/robot-wake/robot-info above, this
# class of tool always fires for real; there is nothing to simulate.
# Defaults to Brobot 2 (0dd1d8bf), the cube keeper - override with a first
# argument for a different serial. Builds the binary first if it's missing
# or its .go source is newer (mtime check) - "builds if needed," so a fresh
# checkout or an edited tool still works with no separate manual build step.
# WIREPOD_HOME set inline, same fix robot-sleep/-wake/-info above already
# apply: vector.NewWP() reads this env var to build an absolute path to
# chipper/jdocs/botSdkInfo.json - unset, it falls back to a path relative to
# whatever cwd the binary happens to be launched from, which is exactly the
# bug a raw, unwired `./direct_sdk_cube_blip_001 0dd1d8bf` run from
# gopod_probes/tools/ hit ("open chipper/jdocs/botSdkInfo.json: no such file
# or directory" - it was looking for tools/chipper/jdocs/..., since nothing
# set WIREPOD_HOME). This alias exists specifically so that never happens
# again.
_CUBE_BLIP_DIRECT_BIN="/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_probes/tools/direct_sdk_cube_blip_001"
_CUBE_BLIP_DIRECT_SRC="/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_probes/tools/direct_sdk_cube_blip_001.go"

_GOPOD_SLEEP_FIRST_WAKE_SETTLE_SECONDS="${GOPOD_SLEEP_FIRST_WAKE_SETTLE_SECONDS:-3}"

# ==========================================================================
# === RETIRED - dead, kept for record. Every alias below is commented    ===
# === out in place; pressing any of these today does nothing. Gathered  ===
# === here 2026-08-13 (BROBOTS_SH_RESTRUCTURE_EXECUTED_001.md) from      ===
# === their original scattered locations (near the top of this file)    ===
# === without losing any of their own reasoning - nothing here was      ===
# === deleted, only relocated.                                          ===
# ==========================================================================

# RETIRED 2026-07-06: target archived, see alias_keyboard_survey_001.md
# alias brobots='gorepo && python3 goverlord/runtime/data_gomad/wire-pod/vector_backpack_brobot_default_fallback_001/validate_vector_backpack_brobot_default_fallback_001.py'
# RETIRED 2026-07-06: target archived, see alias_keyboard_survey_001.md
# alias brobots_audio='brobots pulse && brobots audio vector1 && brobots audio vector2'
# RETIRED 2026-07-06: target archived, see alias_keyboard_survey_001.md
# alias brobots_move='brobots pulse && brobots movement vector1 && brobots movement vector2'
# RETIRED 2026-07-06: target archived, see alias_keyboard_survey_001.md
# alias brobots_expr='brobots pulse && brobots expression vector1 && brobots expression vector2'
# RETIRED 2026-07-06: duplicate of brobots-happy/brobots-angry, see alias_keyboard_survey_001.md
# happy-brobots() { brobots-happy; }
# RETIRED 2026-07-06: duplicate of brobots-happy/brobots-angry, see alias_keyboard_survey_001.md
# angry-brobots() { brobots-angry; }
# RETIRED 2026-07-06: duplicate of brobots-happy/brobots-angry, see alias_keyboard_survey_001.md
# happy-robots() { brobots-happy; }
# RETIRED 2026-07-06: duplicate of brobots-happy/brobots-angry, see alias_keyboard_survey_001.md
# angry-robots() { brobots-angry; }

# gobingo: the live Bingo game (moved from demo.sh 2026-10-08, alias clean-out).
gobingo() {
    (
        cd /home/goverlord/wire-pod || exit 1

        # Brobot 2 angry-animation reactor: launched here only for operator
        # convenience (one terminal, one command instead of two). Still a fully
        # independent process — Bingo's own code has zero awareness of it, no
        # import, no call site. See goverlord/runtime/songs/01b_brobots_bingo_game/bingo_reactor/README.md.
        # Trap ensures it's killed when Bingo exits, however it exits.
        /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/01b_brobots_bingo_game/bingo_reactor/.venv/bin/python3 \
            /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/01b_brobots_bingo_game/bingo_reactor/bingo_reactor_001.py &
        local reactor_pid=$!
        trap 'kill "$reactor_pid" 2>/dev/null' EXIT INT TERM

        WIREPOD_EX_TMP_PATH=/home/goverlord/crushn8r_git/SDK/sources/vectorx/vectorfs/tmp \
        WIREPOD_EX_DATA_PATH=/home/goverlord/crushn8r_git/SDK/sources/vectorx/vectorfs/data \
        WIREPOD_EX_NVM_PATH=/home/goverlord/crushn8r_git/SDK/sources/vectorx/vectorfs/nvm \
        /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/01b_brobots_bingo_game/bin/vectorx-gobingo \
            --serial 0dd1b9e9 \
            --locale en-US \
            --speechText bingo \
            "$@"
    )
}

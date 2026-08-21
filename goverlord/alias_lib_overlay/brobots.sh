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

# === Wire-Pod ===

_brobots_cue_pair() {
  local doc_animation="$1"
  local pip_animation="$2"

  ssh gopod-laptop 'bash -s' -- "$doc_animation" "$pip_animation" <<'EOS'
set -u
BASE="http://127.0.0.1:8080"
DOC="0dd1b9e9"
PIP="0dd1d8bf"
DOC_ANIMATION="$1"
PIP_ANIMATION="$2"
START="$(date '+%Y-%m-%d %H:%M:%S')"

play_anim() {
  local serial="$1"
  local animation="$2"
  curl -fsS --max-time 25 --get "$BASE/api-sdk/play_animation" \
    --data-urlencode "serial=$serial" \
    --data-urlencode "animation=$animation" \
    --data-urlencode "interrupting=false" >/dev/null
}

echo "BROBOT_CUE_START doc=$DOC_ANIMATION pip=$PIP_ANIMATION"

if play_anim "$DOC" "$DOC_ANIMATION"; then
  echo "BROBOT_DOC_CUE_SENT serial=$DOC animation=$DOC_ANIMATION"
else
  echo "BROBOT_DOC_CUE_FAIL serial=$DOC animation=$DOC_ANIMATION"
fi

sleep 1

if play_anim "$PIP" "$PIP_ANIMATION"; then
  echo "BROBOT_PIP_CUE_SENT serial=$PIP animation=$PIP_ANIMATION"
else
  echo "BROBOT_PIP_CUE_FAIL serial=$PIP animation=$PIP_ANIMATION"
fi

echo "BROBOT_INTENT_RESULTS:"
journalctl -u wire-pod --since "$START" --no-pager | grep -E "API_SDK_PLAY_ANIMATION_(REQUEST|DONE)|Intent matched:" | tail -20 || true
echo "BROBOT_CUE_DONE"
EOS
}

brobots-happy() { _brobots_cue_pair "happy" "celebrate"; }
brobots-angry() { _brobots_cue_pair "angry" "frustrated"; }

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

_gopod_note_send() {
  local endpoint="$1"; shift
  GOPOD_WIREPOD_BASE_URL="${GOPOD_WIREPOD_BASE_URL}" \
  python3 - "$endpoint" "$@" <<PYEOF
import importlib.util
import sys

endpoint = sys.argv[1]
params = dict(arg.split("=", 1) for arg in sys.argv[2:])

spec = importlib.util.spec_from_file_location("run_section1_full_live_001", "$_GOPOD_NOTE_RUNNER_PATH")
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)

import time
ts = time.strftime("%H:%M:%S", time.localtime()) + f".{int(time.time() * 1000) % 1000:03d}"
result = mod.wirepod_web_send_form(endpoint, params, timeout=10)
print(f"[{ts}] NOTE_HTTP status={result['status']} body={result['body']!r}")
PYEOF
}

# === Movement notes: lift and head ===
# Real endpoints, traced read-only from the live source (no source touched):
# webroot/sdkapp/js/control.js's keyboard handler (R/F = lift up/down, T/G =
# head up/down) calls /api-sdk/move_lift?speed=<N> and
# /api-sdk/move_head?speed=<N> - both land on pkg/wirepod/sdkapp/server.go's
# MoveLift/MoveHead cases (robot.Conn.MoveLift/MoveHead, SpeedRadPerSec).
# These are continuous-speed commands, not move-to-position ones: the
# keyboard UI starts a speed on keydown and sends speed=0 on keyup, so a
# scripted note holds the speed for a fixed duration, then explicitly zeros
# it - the same shape as pressing then releasing the key by hand. Per
# server.go, the HTTP response returns right after the gRPC call is issued,
# before the physical motion finishes (fire-and-forget, same as the say_text
# path elsewhere in this stack) - the sleep below stands in for real motion
# time, it does not prove it.
#
# Behavior control: control.html's own "Assume" radio
# (onclick="sendForm('/api-sdk/assume_behavior_control?priority=high')")
# gates all movement in the web UI, and README.md documents the same
# assume -> action -> release shape core.sh's speak()/
# _gopod_chord_release_both already use. Followed here with the same
# base-URL/serial env vars core.sh uses (GOPOD_WIREPOD_BASE_URL,
# GOPOD_BROBOT_1_SERIAL, GOPOD_BROBOT_2_SERIAL) rather than the
# SSH-to-gopod-laptop shape brobots-happy/brobots-angry use above - that
# pair's own target, /api-sdk/play_animation, is not a real route in
# server.go's switch (confirmed by reading it; the real animation trigger is
# {{playAnimation||name}} / {{playAnimationWI||name}} parsed out of
# /api-sdk/say_text by wirepod_ttr.GetActionsFromString), so it was not
# trusted as the convention to copy here.
_brobots_move_axis() {
  local endpoint="$1" speed="$2" hold="$3"
  local base="${GOPOD_WIREPOD_BASE_URL}"
  local s1="${GOPOD_BROBOT_1_SERIAL:-0dd1b9e9}"
  local s2="${GOPOD_BROBOT_2_SERIAL:-0dd1d8bf}"
  local serial

  echo "BROBOT_MOVE_START endpoint=$endpoint speed=$speed hold=${hold}s"

  for serial in "$s1" "$s2"; do
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/assume_behavior_control" "priority=high" "serial=$serial" >/dev/null 2>&1
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/$endpoint" "serial=$serial" "speed=$speed" >/dev/null 2>&1
  done

  sleep "$hold"

  for serial in "$s1" "$s2"; do
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/$endpoint" "serial=$serial" "speed=0" >/dev/null 2>&1
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/release_behavior_control" "serial=$serial" >/dev/null 2>&1
  done

  echo "BROBOT_MOVE_DONE endpoint=$endpoint speed=$speed hold=${hold}s"
}

brobots-lift-up() { _brobots_move_axis "move_lift" 2 "${1:-1.2}"; }
brobots-lift-down() { _brobots_move_axis "move_lift" -2 "${1:-1.2}"; }

# Silent yes: one small down-up head sequence, self-contained as a single
# note - nod down first, then back up, matching the literal ask.
brobots-head-nod() {
  local hold="${1:-0.35}"
  _brobots_move_axis "move_head" -2 "$hold"
  _brobots_move_axis "move_head" 2 "$hold"
}

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

_brobots_wake_core() {
  local serial="$1"
  local base="${GOPOD_WIREPOD_BASE_URL}"
  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/conn_test" "serial=$serial" >/dev/null 2>&1
  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/assume_behavior_control" "priority=high" "serial=$serial" >/dev/null 2>&1
  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/release_behavior_control" "serial=$serial" >/dev/null 2>&1
  sleep "$_BROBOTS_WHEEL_WAKE_RELEASE_SETTLE_SECONDS"
  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/assume_behavior_control" "priority=high" "serial=$serial" >/dev/null 2>&1
  sleep "$_BROBOTS_WHEEL_WAKE_POST_REASSUME_SETTLE_SECONDS"
}

# Standalone-fireable: assume+golden-flag-pulse+release, self-contained -
# doesn't leave the robot stuck holding control if fired on its own.
# robot_target: "1"/"2"/"both" (default both).
brobots-wake() {
  local robot_target="${1:-both}"
  local base="${GOPOD_WIREPOD_BASE_URL}"
  local s1="${GOPOD_BROBOT_1_SERIAL:-0dd1b9e9}"
  local s2="${GOPOD_BROBOT_2_SERIAL:-0dd1d8bf}"
  local serial serials

  case "$robot_target" in
    1) serials="$s1" ;;
    2) serials="$s2" ;;
    both) serials="$s1 $s2" ;;
    *)
      echo "BROBOTS_WAKE_BLOCKED robot target must be 1, 2, or both - got '$robot_target'"
      return 1
      ;;
  esac

  echo "BROBOTS_WAKE_START robot=$robot_target"
  for serial in $serials; do
    _brobots_wake_core "$serial"
  done
  for serial in $serials; do
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/release_behavior_control" "serial=$serial" >/dev/null 2>&1
  done
  echo "BROBOTS_WAKE_DONE robot=$robot_target"
}

# PRIME REMOVED 2026-08-10 (WHEEL_NUDGE_COLD_FIRE_ROOT_CAUSE_SURVEY_001.md):
# the old "fire twice, back to back" duplicate was a workaround for a
# still-waiting-on-the-grant theory that was itself only ever tested
# inside a flat-sleep settle window. The golden-flag mechanism in
# _brobots_wake_core above (a real release-then-reassume pulse, not a
# guessed delay) proved reliable with a SINGLE move call across 8/8 live
# runs, both robots - not kept as redundant insurance, removed outright.
_brobots_move_reverse_core() {
  local serial="$1" lw="$2" rw="$3"
  local base="${GOPOD_WIREPOD_BASE_URL}"
  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/move_wheels" "serial=$serial" "lw=$lw" "rw=$rw" >/dev/null 2>&1
}

# Standalone-fireable: assumes its OWN control (no brobots-wake run
# first), moves, holds, stops, releases. Fixed reverse only (-150/-150,
# the webroot's own confirmed S-key/backward value) - forward is a later,
# separate decision. Default hold 0.5s (0.3s proved too short to see).
move-reverse() {
  local hold="${1:-0.5}"
  local robot_target="${2:-both}"
  local base="${GOPOD_WIREPOD_BASE_URL}"
  local s1="${GOPOD_BROBOT_1_SERIAL:-0dd1b9e9}"
  local s2="${GOPOD_BROBOT_2_SERIAL:-0dd1d8bf}"
  local serial serials

  case "$robot_target" in
    1) serials="$s1" ;;
    2) serials="$s2" ;;
    both) serials="$s1 $s2" ;;
    *)
      echo "BROBOT_MOVE_REVERSE_BLOCKED robot target must be 1, 2, or both - got '$robot_target'"
      return 1
      ;;
  esac

  echo "WHEEL_NUDGE CAUTION: cliff sensors disabled while control held - ON-CHARGER USE ONLY"
  echo "BROBOT_MOVE_REVERSE_START hold=${hold}s robot=$robot_target"

  for serial in $serials; do
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/assume_behavior_control" "priority=high" "serial=$serial" >/dev/null 2>&1
  done
  for serial in $serials; do
    _brobots_move_reverse_core "$serial" -150 -150
  done

  sleep "$hold"

  for serial in $serials; do
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/move_wheels" "serial=$serial" "lw=0" "rw=0" >/dev/null 2>&1
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/release_behavior_control" "serial=$serial" >/dev/null 2>&1
  done

  echo "BROBOT_MOVE_REVERSE_DONE hold=${hold}s robot=$robot_target"
}

# Composed: nudge-reverse = brobots-wake + move-reverse, sharing ONE held-
# control window (the wake's own core doesn't release, the move's own
# core doesn't assume) - never two separate windows, that's exactly what
# rounds 1/2 got wrong.
nudge-reverse() {
  local hold="${1:-0.5}"
  local robot_target="${2:-both}"
  local base="${GOPOD_WIREPOD_BASE_URL}"
  local s1="${GOPOD_BROBOT_1_SERIAL:-0dd1b9e9}"
  local s2="${GOPOD_BROBOT_2_SERIAL:-0dd1d8bf}"
  local serial serials

  case "$robot_target" in
    1) serials="$s1" ;;
    2) serials="$s2" ;;
    both) serials="$s1 $s2" ;;
    *)
      echo "BROBOT_MOVE_WHEELS_BLOCKED robot target must be 1, 2, or both - got '$robot_target'"
      return 1
      ;;
  esac

  echo "WHEEL_NUDGE CAUTION: cliff sensors disabled while control held - ON-CHARGER USE ONLY"
  echo "BROBOT_MOVE_WHEELS_START lw=-150 rw=-150 hold=${hold}s robot=$robot_target"

  for serial in $serials; do
    _brobots_wake_core "$serial"
  done
  for serial in $serials; do
    _brobots_move_reverse_core "$serial" -150 -150
  done

  sleep "$hold"

  for serial in $serials; do
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/move_wheels" "serial=$serial" "lw=0" "rw=0" >/dev/null 2>&1
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/release_behavior_control" "serial=$serial" >/dev/null 2>&1
  done

  echo "BROBOT_MOVE_WHEELS_DONE lw=-150 rw=-150 hold=${hold}s robot=$robot_target"
}

# === Animation notes: one per confirmed playAnimationWI token ===
# Confirmed read-only this session: a bare, speech-free /api-sdk/say_text
# call is possible and clean. Text "{{playAnimationWI||<token>}}" round-trips
# unchanged through NormalizeAnimationSyntax
# (pkg/wirepod/ttr/kgsim_cmds_animation_normalizer.go) and
# GetActionsFromString (pkg/wirepod/ttr/kgsim_cmds.go) parses it into a
# single ActionPlayAnimationWI with no accompanying ActionSayText (the text
# after "}}" is empty, and cleanRobotSpeechText("") is skipped) - so no words
# are ever spoken, only the animation fires. playAnimationWI/DoPlayAnimationWI
# is the async/fire-and-forget token confirmed this session (launches a
# goroutine, returns immediately, never interrupts speech) - used here per
# the ask, not the blocking playAnimation.
#
# Behavior control: say_text's own HTTP handler doesn't gate on BcAssumption
# in code, but the established, real convention for this exact endpoint is
# core.sh's speak() (assume -> say_text -> release) - reused here rather than
# guessed away. Because a bare animation-only say_text returns "success"
# almost immediately (DoPlayAnimationWI doesn't wait for the clip to finish -
# same fire-and-forget gotcha as move_lift/move_head above), releasing
# behavior control right away could hand control back to the robot's own
# behavior arbitration before the queued clip actually plays, so this note
# holds control for a short estimated duration before releasing - the sleep
# stands in for real motion time, it doesn't prove it.
#
# Only the original 10 tokens get a both-robots note here. The 4
# knowledge-graph tokens (answering/searching/searchingGetout/kgSuccess,
# added 2026-07-24, corrected to match Wire-Pod's own real
# anim_knowledgegraph_* names per kgsim.go) are now also verified:true -
# live-confirmed 2026-07-24, after a wire-pod.service restart picked up the
# vocab file (the running process caches animation_vocab.json at startup;
# every fire before the restart silently fell back to "thinking" instead of
# erroring, which is why the first two test rounds looked like a dead end) -
# but stay reachable only through brobots-anim-test/-test-all below, not
# promoted to their own both-robots aliases here.
_brobots_play_anim() {
  local token="$1" hold="${2:-2.5}"
  local base="${GOPOD_WIREPOD_BASE_URL}"
  local s1="${GOPOD_BROBOT_1_SERIAL:-0dd1b9e9}"
  local s2="${GOPOD_BROBOT_2_SERIAL:-0dd1d8bf}"
  local serial

  echo "BROBOT_ANIM_START token=$token hold=${hold}s"

  for serial in "$s1" "$s2"; do
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/assume_behavior_control" "priority=high" "serial=$serial" >/dev/null 2>&1
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/say_text" "serial=$serial" "text={{playAnimationWI||$token}}" >/dev/null 2>&1
  done

  sleep "$hold"

  for serial in "$s1" "$s2"; do
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/release_behavior_control" "serial=$serial" >/dev/null 2>&1
  done

  echo "BROBOT_ANIM_DONE token=$token hold=${hold}s"
}

brobots-anim-happy()      { _brobots_play_anim "happy" "${1:-2.5}"; }
brobots-anim-very-happy() { _brobots_play_anim "veryHappy" "${1:-2.5}"; }
brobots-anim-sad()        { _brobots_play_anim "sad" "${1:-2.5}"; }
brobots-anim-very-sad()   { _brobots_play_anim "verySad" "${1:-2.5}"; }
brobots-anim-angry()      { _brobots_play_anim "angry" "${1:-2.5}"; }
brobots-anim-frustrated() { _brobots_play_anim "frustrated" "${1:-2.5}"; }
brobots-anim-confused()   { _brobots_play_anim "confused" "${1:-2.5}"; }
brobots-anim-thinking()   { _brobots_play_anim "thinking" "${1:-2.5}"; }
brobots-anim-celebrate()  { _brobots_play_anim "celebrate" "${1:-2.5}"; }
brobots-anim-love()       { _brobots_play_anim "love" "${1:-2.5}"; }

# kgsim.go plays "searching" and "answering" as a real loop - an external Go
# loop re-firing the clip every ~1/3 second for as long as that phase lasts,
# NOT a single play held open. "searchingGetout"/"kgSuccess" are one-shots in
# kgsim.go. _brobots_play_anim_single below reproduces that same distinction
# instead of just firing every token once and sleeping - added 2026-07-24
# after a single-fire test made searching/answering look like a brief blip
# rather than the sustained "holding pattern" they actually are in production.
_brobots_anim_is_loop_token() {
  case "$1" in
    searching|answering) return 0 ;;
    *) return 1 ;;
  esac
}

# One-robot single-token fire, for judging an unverified/experimental token
# (e.g. the 4 knowledge-graph animations added 2026-07-24) without committing
# both robots to it. Same assume/say_text/release shape as _brobots_play_anim
# above, just one serial instead of a loop over both, and the token is a real
# argument instead of baked into the alias name. Loop-type tokens (see
# _brobots_anim_is_loop_token above) re-fire every ~0.333s for the hold
# duration, matching kgsim.go's own real cadence; one-shot tokens fire once
# and sleep, same as before.
_brobots_play_anim_single() {
  local token="$1" robot="${2:-1}" hold="${3:-2.5}"
  local base="${GOPOD_WIREPOD_BASE_URL}"
  local s1="${GOPOD_BROBOT_1_SERIAL:-0dd1b9e9}"
  local s2="${GOPOD_BROBOT_2_SERIAL:-0dd1d8bf}"
  local serial

  if [ -z "$token" ]; then
    echo "BROBOT_ANIM_TEST_USAGE: brobots-anim-test <token> [robot: 1|2, default 1] [hold seconds, default 2.5]"
    return 1
  fi
  case "$robot" in
    1) serial="$s1" ;;
    2) serial="$s2" ;;
    *) echo "BROBOT_ANIM_TEST_BAD_ROBOT robot=$robot (use 1 or 2)"; return 1 ;;
  esac

  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/assume_behavior_control" "priority=high" "serial=$serial" >/dev/null 2>&1

  if _brobots_anim_is_loop_token "$token"; then
    local repeats
    repeats=$(awk -v h="$hold" 'BEGIN{r=int(h/0.333); if (r<1) r=1; print r}')
    echo "BROBOT_ANIM_TEST_START robot=$robot serial=$serial token=$token hold=${hold}s mode=loop repeats=$repeats"
    local n=0
    while [ "$n" -lt "$repeats" ]; do
      GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/say_text" "serial=$serial" "text={{playAnimationWI||$token}}" >/dev/null 2>&1
      sleep 0.333
      n=$((n + 1))
    done
  else
    echo "BROBOT_ANIM_TEST_START robot=$robot serial=$serial token=$token hold=${hold}s mode=once"
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/say_text" "serial=$serial" "text={{playAnimationWI||$token}}" >/dev/null 2>&1
    sleep "$hold"
  fi

  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/release_behavior_control" "serial=$serial" >/dev/null 2>&1
  echo "BROBOT_ANIM_TEST_DONE robot=$robot serial=$serial token=$token hold=${hold}s"
}

brobots-anim-test() { _brobots_play_anim_single "$1" "${2:-1}" "${3:-2.5}"; }

# brobots-searching-out - "searching" loop token (kgsim.go's own real
# KG-search holding pattern - re-fired every ~0.333s, same cadence
# _brobots_play_anim_single/brobots-anim-test-all already use for this same
# token) held for `hold` seconds (default 1, per operator request - much
# shorter than a judgment-pass hold since this is meant as a quick loop, not
# an extended one), then "searchingGetout" fired once to close the loop
# cleanly - all under one continuously-held assume/release session, not two
# separate brobots-anim-test calls (which would each grab and release
# control on their own, reintroducing the per-call assume/release churn this
# repo's own say_line()/manage_control precedent already fixed elsewhere).
# Argument: robot (1 or 2, default 1), hold seconds (default 1). Same
# 0.333s-per-repeat math _brobots_play_anim_single already uses for loop
# tokens - not re-derived, just not factored into a shared helper (one line,
# low duplication risk, existing tested function left untouched).
brobots-searching-out() {
  local robot="${1:-1}" hold="${2:-1}"
  local base="${GOPOD_WIREPOD_BASE_URL}"
  local s1="${GOPOD_BROBOT_1_SERIAL:-0dd1b9e9}"
  local s2="${GOPOD_BROBOT_2_SERIAL:-0dd1d8bf}"
  local serial
  case "$robot" in
    1) serial="$s1" ;;
    2) serial="$s2" ;;
    *) echo "BROBOTS_SEARCHING_OUT_BAD_ROBOT robot=$robot (use 1 or 2)"; return 1 ;;
  esac

  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/assume_behavior_control" "priority=high" "serial=$serial" >/dev/null 2>&1

  local repeats
  repeats=$(awk -v h="$hold" 'BEGIN{r=int(h/0.333); if (r<1) r=1; print r}')
  echo "BROBOTS_SEARCHING_OUT_START robot=$robot serial=$serial hold=${hold}s repeats=$repeats"
  local n=0
  while [ "$n" -lt "$repeats" ]; do
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/say_text" "serial=$serial" "text={{playAnimationWI||searching}}" >/dev/null 2>&1
    sleep 0.333
    n=$((n + 1))
  done

  echo "BROBOTS_SEARCHING_OUT_GETOUT robot=$robot serial=$serial"
  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/say_text" "serial=$serial" "text={{playAnimationWI||searchingGetout}}" >/dev/null 2>&1
  sleep 0.5

  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/release_behavior_control" "serial=$serial" >/dev/null 2>&1
  echo "BROBOTS_SEARCHING_OUT_DONE robot=$robot serial=$serial hold=${hold}s"
}

# Runs all 4 knowledge-graph tokens (verified:true, live-confirmed 2026-07-24)
# back to back on one robot, a 1s gap between each so the operator can watch
# and judge each one on its own. Order matches kgsim.go's own real sequence:
# searching (loop while the KG/LLM searches) -> searchingGetout (transition
# out) -> answering (loop while speaking) -> kgSuccess (real name, currently
# unused/commented out in kgsim.go, included anyway since it's a confirmed
# real clip in the family). Reuses brobots-anim-test/_brobots_play_anim_single
# per token - no new dispatch mechanism, just a sequencer over the existing
# one.
brobots-anim-test-all() {
  local robot="${1:-1}" seed_hold="${2:-2.5}"
  local tokens=(searching searchingGetout answering kgSuccess)
  local default_hold typed hold i=0
  for token in "${tokens[@]}"; do
    if [ "$i" -eq 0 ]; then
      read -r -p "Hold seconds for $token (Enter for ${seed_hold}s): " typed
      default_hold="${typed:-$seed_hold}"
      hold="$default_hold"
    else
      read -r -p "Hold seconds for $token (Enter for ${default_hold}s): " typed
      hold="${typed:-$default_hold}"
    fi
    _brobots_play_anim_single "$token" "$robot" "$hold"
    sleep 1
    i=$((i + 1))
  done
}

brobots-check() {
  bash -n ~/.bashrc &&
  bash -n ~/.bash_aliases &&
  bash -n ~/.gopod_alias_lib/*.sh &&
  command -v brobots >/dev/null &&
  command -v brobots-happy >/dev/null &&
  command -v brobots-angry >/dev/null &&
  command -v happy-brobots >/dev/null &&
  command -v angry-brobots >/dev/null &&
  echo "BROBOTS_ALIAS_CHECK_PASS"
}

brobots-grep() {
  grep -nE '^(alias |[a-zA-Z0-9_-]+\(\)|_[a-zA-Z0-9_-]+\(\))' ~/.bash_aliases ~/.gopod_alias_lib/*.sh
}
# FIXED 2026-07-07: was wrapping the runner in the standalone announcer's
# --run mode, which duplicates the runner's own built-in Kokoro narration
# (_announce_status_kokoro, wired into every status() call via a persistent
# --speak-stdin worker) - two aplay playbacks per status line fought over
# the audio device and hung. Runner now handles its own narration; run it
# directly.
# RETIRED 2026-08-17 (PHA0B_INTERVIEW_CONSOLIDATION_EXECUTED_001.md, operator
# instruction: "reduce to one golden process"). Same call now lives behind
# pha0b's own door: `pha0b` -> pick Interview -> `g`, or `pha0b interview`
# directly. Commented out in place, not deleted - pressed, it does nothing.
# alias start-the-interview='cd /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools && export GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 && python3 run_section1_full_live_001.py'

# interview-json (RENAMED from start-the-preshow 2026-08-19,
# NAMING_APPLIED_001.md, operator naming decision - a survey pass first
# found this function already IS the "generate-only, no playback, no vamp"
# behavior the naming scheme wanted under the interview- prefix, so this is
# a pure rename onto existing, unchanged code, not new behavior). Runs only
# the Kokoro-narrated half (warm-up, scaffold/card load, ESN routing
# announce, LLM generating every exchange line) via
# run_section1_full_live_001.py's own generate_phase() - no robot speaks an
# interview line in this stage. It writes a JSON log and prints its path;
# feed that path to interview-replay via GOPOD_SECTION1_REPLAY_LOG (or just
# call interview-replay bare - it reads the newest file under demo_runs/,
# which this call just became) to play the already-generated script back as
# the actual interview, standalone, in its own separate run.
interview-json() {
  ( cd /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools && \
    GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 python3 run_section1_preshow_generate_001.py )
}

# RETIRED 2026-08-19 (NAMING_APPLIED_001.md, operator naming decision) -
# renamed to interview-json above. Commented out in place, not deleted -
# pressed, it does nothing.
# start-the-preshow() { interview-json "$@"; }

# RETIRED 2026-08-17 (PHA0B_INTERVIEW_CONSOLIDATION_EXECUTED_001.md, operator
# instruction: "reduce to one golden process"). One-shot generate+perform now
# lives behind pha0b's own door (`pha0b` -> pick Interview -> `g`, or
# `pha0b interview`). Commented out in place, not deleted - pressed, it does
# nothing.
# gopod-preshow-then-interview() {
#   local out log_path
#   out="$(start-the-preshow)" || { printf '%s\n' "$out"; return 1; }
#   printf '%s\n' "$out"
#   log_path="$(printf '%s\n' "$out" | sed -n 's/.*PRESHOW_GENERATE_DONE log_path=//p')"
#   if [ -z "$log_path" ]; then
#     echo "GOPOD_PRESHOW_THEN_INTERVIEW_FAIL no log_path in preshow output"
#     return 1
#   fi
#   ( cd /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools && \
#     GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 GOPOD_SECTION1_REPLAY_LOG="$log_path" python3 run_section1_full_live_001.py )
# }

# === Section 1 stage keys (rename + prerequisite wiring) ===
# gopod_warm_up below is the one remaining caller of _gopod_require_brobots,
# post-retirement of gopod_interview/gopod-preshow-then-interview (2026-08-17,
# PHA0B_INTERVIEW_CONSOLIDATION_EXECUTED_001.md) - the multi-frame call-stack
# scenario the guard's own dynamic-scoping design was originally built for
# (gopod_interview --full declaring it once, downstream calls in the same
# stack seeing it) no longer has a live caller. start-the-preshow itself
# still has zero prerequisite, unchanged, so muscle memory calling it
# directly keeps the old behavior.
#
# _gopod_require_brobots itself: a bash `local` variable is dynamically
# scoped to the whole call stack for as long as the declaring function is
# still running - a standalone call to gopod_warm_up (no ancestor frame
# declared the guard) always finds it unset and actually runs gopod_brobots
# fresh. The variable is never `export`-ed, so it never leaks into the
# interactive shell or lingers into a later, separate invocation.
_gopod_require_brobots() {
  if [ "${_GOPOD_BROBOTS_SATISFIED:-0}" = "1" ]; then
    echo "GOPOD_BROBOTS_PREREQ_SKIPPED already-ready-this-run"
    return 0
  fi
  gopod_brobots
  local rc=$?
  [ "$rc" -eq 0 ] && _GOPOD_BROBOTS_SATISFIED=1
  return "$rc"
}

# gopod_warm_up - gopod_brobots (required), then the Kokoro-narrated
# generate-only phase (interview-json's own unchanged body, renamed from
# start-the-preshow 2026-08-19: warm-up, routing, brain writes all
# interview lines, saves the log). No robot speaks an interview line here.
gopod_warm_up() {
  local _GOPOD_BROBOTS_SATISFIED="${_GOPOD_BROBOTS_SATISFIED:-0}"
  _gopod_require_brobots || { echo "GOPOD_WARM_UP_ABORTED brobots_prereq_failed"; return 1; }
  interview-json
}

# RETIRED 2026-08-17 (PHA0B_INTERVIEW_CONSOLIDATION_EXECUTED_001.md, operator
# instruction: "reduce to one golden process"). Both bodies this called
# (start-the-interview, gopod-preshow-then-interview) are retired alongside
# it. Chord prerequisite (gopod_brobots) still available standalone via
# gopod_warm_up or gopod-opening-chord if wanted before a pha0b interview run
# - not folded into pha0b itself, matching how every other song via pha0b
# already works (no auto-chord). Commented out in place, not deleted -
# pressed, it does nothing.
# gopod_interview() {
#   local _GOPOD_BROBOTS_SATISFIED="${_GOPOD_BROBOTS_SATISFIED:-0}"
#   if [ "$1" = "--full" ]; then
#     _gopod_require_brobots || { echo "GOPOD_INTERVIEW_FULL_ABORTED brobots_prereq_failed"; return 1; }
#     gopod-preshow-then-interview
#     return $?
#   fi
#   _gopod_require_brobots || { echo "GOPOD_INTERVIEW_ABORTED brobots_prereq_failed"; return 1; }
#   start-the-interview
# }

# === Vamp split: two clean halves, not yet joined ==========================
# INTERVIEW_VAMP_ALIASES_WIRED_001.md. Operator's design: this function "rolls
# a take" (concurrent pre-show vamp gate + JSON generation, run_scored_preshow_
# and_generate/run_preshow_song - the developing golden boilerplate for a
# reusable pre-show any song can get later); interview-run "plays the take"
# (part-2 playback_phase only, against the latest already-generated JSON, no
# regeneration). Deliberately NOT chained together yet - once the vamp is
# golden it gets inserted before the performance as one flow; these are the
# two clean halves that flow gets built from, not that flow itself.
#
# RENAMED 2026-08-19 (GOPOLISHER_FIXES_001.md, operator naming decision,
# scheme B): vamp-run -> interview-vamp. Old name retired in place, not
# deleted - see the one-line pointer stub below the real function.
#
# interview-vamp (formerly vamp-run) - calls run_scored_preshow_and_generate() directly (bypassing
# main()'s own unconditional playback_phase() call at the end) so this button
# stops after generation, never performs. Mirrors run_section1_preshow_
# generate_001.py's own shape (same importlib load, same explicit
# write_timing_log() call after - main() only calls that itself, so a caller
# that bypasses main() has to call it too, same as that script already does)
# plus main()'s own live/voice_destination resolution (lines ~4023-4026 of
# run_section1_full_live_001.py) so this reproduces exactly what main() would
# have resolved, not a fresh guess. Live by default, same convention as
# start-the-interview - no separate export step needed. The generated JSON
# itself lands wherever write_log()/LOG_DIR already puts it
# (gopod_probes/demo_runs/section1_full_live_<run_id>/..., confirmed by
# reading the runner directly - NOT the song's own runs/ folder, that folder
# only ever gets the timing log this alias also writes).
# Parameterized 2026-08-15 (VAMP_RUN_PARAMETERIZED_001.md): the vamp is a reusable
# module - the gate, the loader, and the reporter delivery are already song-agnostic
# (VAMP_MODULE_STANDARDIZE_SURVEY_001.md). Only the interview actually uses it today,
# because it's the only shelf song that generates its content live - ready for more the
# moment a second generating song exists.
# Repointed 2026-08-19 (INTERVIEW_VAMP_SPLIT_001.md): the old single folder
# split into 01_brobots_interview_vamp/ (own standalone song now) and
# 02_brobots_interview_run/ (this - the interview content). The old
# "$1/vamp subfolder" convention broke structurally once vamp stopped being
# nested - simplified rather than generalized, since only this one vamp
# exists today: the pre-show/vamp folder is now always the new standalone
# vamp path, unconditionally. $1, if given, still overrides
# GOPOD_SECTION_SONG_DIR (the generating song's own folder, default the new
# RUN path via DEFAULT_SECTION_SONG_DIR) - unchanged in spirit, just no
# longer derives the vamp path from it. Bare `interview-vamp` (no arg) behaves
# byte-equivalent to before: GOPOD_SECTION_SONG_DIR stays unset, falls
# through to DEFAULT_SECTION_SONG_DIR internally (now 02_brobots_interview_run).
interview-vamp() {
  local song_folder="${1:-/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/02_brobots_interview_run}"
  export GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1
  if [ -n "$1" ]; then
    export GOPOD_SECTION_SONG_DIR="$song_folder"
  fi
  python3 -c "
import importlib.util
import os

spec = importlib.util.spec_from_file_location('run_section1_full_live_001', '$_GOPOD_NOTE_RUNNER_PATH')
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)

voice_destination = mod.resolve_voice_destination()
live = (os.getenv(mod.LIVE_GATE) == '1') and (voice_destination == 'robot')
song_dir = '/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/01_brobots_interview_vamp'

log, interview_scaffold, robots = mod.run_scored_preshow_and_generate(song_dir, live, voice_destination, read_sheet=False)
timing_path = mod.write_timing_log(
    run_id=log.get('run_id'),
    song_dir=log.get('source_card'),
    meta={'tool': 'interview-vamp', 'log_path': log.get('log_path'), 'preshow_song_dir': song_dir},
)
print(f'VAMP_RUN_DONE log_path={log.get(\"log_path\")}')
print(f'TIMING_LOG_WRITTEN path={timing_path}')
"
}

# RETIRED 2026-08-19 (GOPOLISHER_FIXES_001.md, operator naming decision,
# scheme B) - renamed to interview-vamp above. Commented out in place, not
# deleted - pressed, it does nothing.
# vamp-run() { interview-vamp "$@"; }

# interview-replay (RENAMED from interview-run 2026-08-19, NAMING_APPLIED_001.md,
# operator naming decision) - plays video 2 (the part-2 playback_phase path)
# standalone, against the newest generated JSON under gopod_probes/demo_runs/
# (interview-vamp's own output home, confirmed by reading write_log()/LOG_DIR
# directly - NOT the song's own runs/ folder). Sets GOPOD_SECTION1_REPLAY_LOG
# so generate_phase() loads that JSON instead of calling the LLM (module-level
# os.getenv read at import time - must be exported before python3 starts,
# same as this alias already does), then runs the runner exactly like
# start-the-interview once did (unchanged call, no preshow dir set, so no
# vamp gate fires here - this button only ever performs). Refuses cleanly, no
# python call at all, if no generated JSON exists yet (run interview-vamp or
# interview-json first).
#
# Naming note: this function's BODY is unchanged from its old interview-run
# name - only the name moved. interview-run itself is NOT retired (see the
# new function below it) - the NAME is reused for a different, new meaning
# (the interview + optional full-run-with-vamp orchestrator), so there is no
# "old interview-run, commented out" stub here the way vamp-run/preshow-run
# got one; this rename freed the name for reuse rather than retiring it.
interview-replay() {
  local demo_runs_dir="/home/goverlord/wire-pod/chipper/gopod_probes/demo_runs"
  local latest_json
  latest_json=$(ls -t "$demo_runs_dir"/section1_full_live_*/section1_full_live_log_*.json 2>/dev/null | head -1)
  if [ -z "$latest_json" ]; then
    echo "INTERVIEW_REPLAY_REFUSED no generated JSON found under $demo_runs_dir - run interview-vamp or interview-json first"
    return 1
  fi
  ( cd /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools && \
    export GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 && \
    export GOPOD_SECTION1_REPLAY_LOG="$latest_json" && \
    echo "INTERVIEW_REPLAY_PLAYING $latest_json" && \
    python3 run_section1_full_live_001.py )
}

# interview-run (NEW meaning 2026-08-19, NAMING_APPLIED_001.md, operator
# naming decision) - the interview itself, video 2, with an optional
# full-run mode that plays the vamp (video 1) first. No CLI flags - two
# interactive prompts only, reusing existing mechanisms rather than
# duplicating any of their internals (interview-vamp/interview-vamp-play/
# interview-replay's own bodies are unchanged and untouched by this
# function - it only ever calls them):
#
#   y/n "include the vamp first?" - default NO, since "just perform the
#     interview" (interview-replay's own existing job) is the lighter,
#     more common ask; opting into the vamp is the heavier, deliberate
#     choice, so it should require an explicit yes, not an accidental
#     bare-Enter.
#   NO  -> interview-replay (replay whatever's already on disk, unchanged
#          existing behavior).
#   YES -> a/b "(a) fresh vamp + fresh generation, or (b) vamp for
#          atmosphere + replay the existing take?"
#     (a) -> interview-vamp (rolls a fresh take, vamp plays while
#            generation runs), then interview-replay - since interview-vamp
#            writes its output into the same demo_runs/ tree
#            interview-replay's own `ls -t` glob already reads, the take
#            interview-vamp just finished generating IS the newest file by
#            construction - no log-path plumbing needed between the two
#            calls, confirmed by reading write_log()/LOG_DIR directly (see
#            interview-vamp's own header comment above).
#     (b) -> interview-vamp-play (vamp only, zero generation triggered),
#            then interview-replay (replays whatever take was ALREADY on
#            disk before this call - interview-vamp-play never writes a
#            new one, so "newest" is unchanged from before the vamp played).
interview-run() {
  local include_vamp
  include_vamp="$(_pha0b_prompt_yn "include the vamp (interview-vamp) first? y/n [default n]: " n n-means-no)"
  if [ "$include_vamp" != "1" ]; then
    interview-replay
    return $?
  fi
  local mode_choice
  read -r -p "(a) fresh vamp + fresh generation, or (b) vamp for atmosphere + replay the existing take? [a/b, default a]: " mode_choice
  case "$mode_choice" in
    b|B)
      interview-vamp-play || return $?
      ;;
    *)
      interview-vamp || return $?
      ;;
  esac
  interview-replay
}

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

# interview-vamp-play (formerly preshow-run) - the pure "video 1" button.
# Added 2026-08-19 (INTERVIEW_VAMP_NO_GEN_PATH_001.md), once the operator
# called for a no-generation-side-effect fire path for the vamp/pre-show
# song (01_brobots_interview_vamp) - interview-vamp() itself stays untouched,
# since its own "roll a take" job genuinely needs generation running
# alongside it (that's the whole reason interview-vamp exists). This is a
# sibling, not an interview-vamp rewrite - a distinct new alias, not a
# repurposed gopod-vamp (that one's own contract is the vamp_1..vamp_4
# filler beats only, no robot/LLM involved at all - extending it to the
# full m1-m4 banter would add resource requirements gopod-vamp was
# deliberately built to avoid). Calls the runner's new run_preshow_only():
# plays M1/M2/M3/M4 once, zero vamp-filler looping (nothing to wait for),
# zero interview generation triggered - generate_phase() is never called.
# Still reads the shared interview_scaffold for Brobot 1/2's two
# llm_coloured wake-beat lines (persona/pronunciation consistency - a read,
# not generation). Live by default, same convention as interview-vamp/
# interview-run.
#
# RENAMED 2026-08-19 (GOPOLISHER_FIXES_001.md, operator naming decision,
# scheme B): preshow-run -> interview-vamp-play. Chosen over the original
# scheme-B suggestion (interview-video1-run) to avoid a real collision the
# operator caught live - "interview-run" already means "video 2, replay, no
# generation," so the video-1 no-generation player needed its own distinct
# name rather than a video-numbered one that reads too close to the
# existing video-2 name. interview-vamp-play pairs cleanly with
# interview-vamp (roll a take, WITH generation) - same "vamp" root, "-play"
# suffix marks the no-generation player, no collision with interview-run's
# "video 2" meaning. Old name retired in place, not deleted - see the
# one-line pointer stub below the real function.
#
# VAMP is this check's testbed (AUDIO_ROUTING_CHECK_001.md) - the natural
# place to verify "will this actually make sound" before a performance,
# since it's the standalone video-1 fire path. Runs before the robots ever
# get exported into live-speech mode.
interview-vamp-play() {
  local song_folder="${1:-/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/01_brobots_interview_vamp}"
  _gopod_check_audio_routing
  export GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1
  python3 -c "
import importlib.util
import os

spec = importlib.util.spec_from_file_location('run_section1_full_live_001', '$_GOPOD_NOTE_RUNNER_PATH')
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)

voice_destination = mod.resolve_voice_destination()
live = (os.getenv(mod.LIVE_GATE) == '1') and (voice_destination == 'robot')
song_dir = '$song_folder'

song_id = mod.run_preshow_only(song_dir, live, voice_destination, read_sheet=False)
timing_path = mod.write_timing_log(
    run_id=mod.utc_run_id(),
    song_dir=song_dir,
    meta={'tool': 'interview-vamp-play', 'song_id': song_id, 'preshow_song_dir': song_dir},
)
print(f'PRESHOW_RUN_DONE song_id={song_id}')
print(f'TIMING_LOG_WRITTEN path={timing_path}')
"
}

# RETIRED 2026-08-19 (GOPOLISHER_FIXES_001.md, operator naming decision,
# scheme B) - renamed to interview-vamp-play above. Commented out in place,
# not deleted - pressed, it does nothing.
# preshow-run() { interview-vamp-play "$@"; }

# Shared interview launch bypass - the one place pha0b (both call shapes) routes
# an Interview pick. Interview can't be step-sliced (line_type/exchange_type/
# brobot_1_mode fields, not note/speaker/pause_seconds - PHA0B_VAMP_INTEGRATION_
# SURVEY_001.md), so this is a bypass, not a slice-flow entry. Three real choices:
# roll a take (generate only, interview-vamp), perform the last take (replay
# only, interview-replay), or one-shot (generate then perform right now - the
# same call start-the-interview used to make before consolidation retired it
# in favor of this single shared door,
# PHA0B_INTERVIEW_CONSOLIDATION_EXECUTED_001.md).
#
# Reconciliation with the new interview-run (NAMING_APPLIED_001.md,
# 2026-08-19): this bypass's own v/p/g menu is kept exactly as it was,
# deliberately NOT folded into interview-run's own newer 2-question
# (y/n then a/b) flow - stacking interview-run's own prompts inside this
# already-interactive 3-way menu would mean answering questions about
# questions for the "p" (replay-only, no vamp at all) case, which never
# needed asking here. Only the `p` case's own callee was repointed, from
# the old interview-run (replay-only meaning) to interview-replay (the
# same behavior under its new name) - everything else about this bypass is
# byte-for-byte unchanged.
_pha0b_interview_bypass() {
  local interview_choice
  echo "** default to 'v' **"
  read -r -p "vamp this take (roll a new one), perform the last take, or go one-shot (generate + perform now)? [v/p/g, default v]: " interview_choice
  case "$interview_choice" in
    p|P)
      interview-replay
      ;;
    g|G)
      ( cd /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools && \
        export GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 && \
        python3 run_section1_full_live_001.py )
      ;;
    *)
      interview-vamp "/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/02_brobots_interview_run"
      ;;
  esac
}

# Robot control song - a talking self-check on one robot: connect, say
# "I'm connected", arm cue, head nod, fireworks, say "I'm good", exit. See
# ROBOT_CONTROL_SONG_001.md. Brobot 1 by default; pass "2" to target Brobot
# 2 instead. Dry by default - unlike start-the-interview/start-the-preshow
# above, this one does NOT export GOPOD_ALLOW_LIVE_ROBOT_SPEECH itself;
# export that gate yourself first if you want real hardware, same
# convention every other GOPOD_ALLOW_LIVE_ROBOT_SPEECH-gated call in this
# codebase already follows.
# CUTOVER PREP 2026-08-12 (studio tuning cut 2 step 2,
# CONTROL_SONG_GOLDEN_CUTOVER_PREP_001.md): repointed from the legacy
# run_robot_control_song_001.py to the golden engine (run_golden_song_001.py),
# already golden-registered as "robot_control_song_001" (SONG_REGISTRY,
# commit 16aa54e) - mirrors bait's own 2026-08-07 cutover
# (bingo-video-song's export pattern). Robot select now goes through
# GOPOD_GOLDEN_ROBOT (the golden engine's own env var), not
# GOPOD_CONTROL_SONG_ROBOT (which this runner never reads). PREPPED,
# DRY-OK - not live-fire confirmed yet; run_robot_control_song_001.py stays
# fully intact as fallback (see test-arm-cue/test-head-nod/test-fireworks/
# gopod-fireworks/gopod-weather-say below, all untouched, still calling it
# directly).
start-the-control-song() {
  ( cd /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools && \
    export GOPOD_GOLDEN_SONG_DIR=/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/zzz_archives/robot_control_song_001 && \
    export GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 && \
    GOPOD_GOLDEN_ROBOT="${1:-1}" python3 run_golden_song_001.py )
}

# Isolated single-note testers for the robot control song - fire exactly
# one motion, no song loop, no other notes' assume/release churn in the
# way. Same dry-by-default / GOPOD_ALLOW_LIVE_ROBOT_SPEECH-gated
# convention as start-the-control-song; same "2" argument to target
# Brobot 2 instead. Built for tuning timing live, one motion at a time.
# apply phcal tweaks prompt, added 2026-07-25 (operator request, alias-mixer
# widening pass) - TEST-only, one primitive at a time, matching these
# aliases' own "isolate one motion" philosophy. Calls the same
# phcal_apply_control_song_001.py the GESTURE side could reuse later, but
# --target test here always, never gesture - the interview's own live
# arm_gesture/head_nod_gesture motion is deliberately untouched by this
# prompt (see that script's own module docstring).
test-arm-cue() {
  local apply_choice
  read -r -p "apply phcal tweaks to this test cue? y/n " apply_choice
  if [ "$apply_choice" = "y" ] || [ "$apply_choice" = "Y" ]; then
    python3 /home/goverlord/.gopod_alias_lib/phcal_apply_control_song_001.py --yes --target test --primitive arm
  fi
  ( cd /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools && \
    GOPOD_CONTROL_SONG_NOTE=arm_test GOPOD_CONTROL_SONG_ROBOT="${1:-1}" python3 run_robot_control_song_001.py )
}

test-head-nod() {
  local apply_choice
  read -r -p "apply phcal tweaks to this test cue? y/n " apply_choice
  if [ "$apply_choice" = "y" ] || [ "$apply_choice" = "Y" ]; then
    python3 /home/goverlord/.gopod_alias_lib/phcal_apply_control_song_001.py --yes --target test --primitive nod
  fi
  ( cd /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools && \
    GOPOD_CONTROL_SONG_NOTE=head_nod GOPOD_CONTROL_SONG_ROBOT="${1:-1}" python3 run_robot_control_song_001.py )
}

test-fireworks() {
  ( cd /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools && \
    GOPOD_CONTROL_SONG_NOTE=fireworks GOPOD_CONTROL_SONG_ROBOT="${1:-1}" python3 run_robot_control_song_001.py )
}

# gopod-fireworks - human-named front door onto the exact same fireworks
# note test-fireworks already plays above (registry polish 2026-07-16, see
# ALIAS_REGISTRY_POLISH_001.md/ALIAS_REGISTRY_TRUTH_SWEEP_001.md - the
# sweep's own NOTES table had flagged cloud_intent/fireworks as alias
# coverage "No," which was only true at the raw-note level; test-fireworks
# already covered it at the tester level). Not a second implementation -
# a plain wrapper, same "2" argument, same dry-by-default convention.
gopod-fireworks() {
  test-fireworks "$@"
}

# gopod-weather-say - standalone version of the control song's own single-
# robot "weather" note (real Windsor fetch, formatted per that robot's own
# unit/clock, spoken once) - one of the three gaps
# ALIAS_REGISTRY_TRUTH_SWEEP_001.md found with zero alias at all.
# run_single_note()'s own "weather" branch (added same polish pass, in
# run_robot_control_song_001.py) is what this calls; not a second
# implementation, and it does not touch any song's own knobs/story. Same
# dry-by-default / "2" convention as every other single-note tester here.
gopod-weather-say() {
  ( cd /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools && \
    GOPOD_CONTROL_SONG_NOTE=weather GOPOD_CONTROL_SONG_ROBOT="${1:-1}" python3 run_robot_control_song_001.py )
}

# start-the-bait-song retired 2026-08-11 (operator direction: "stale, prune it,
# easily rebuilt"). It called run_robot_control_song_001.py directly - the
# legacy path - while pha0b's own "bait" case cut over to the golden engine
# (run_golden_song_001.py) back on 2026-08-07. No other caller found beyond
# documentation (checked brobots.sh, tech/, README.md, .claude/,
# gopod_probes/) - no live script/intent execs this name. Use `pha0b` ->
# pick 00_brobots_awaken (keyword "bait") instead - same song, golden engine.

# Interview's own short capture cut - brobots_bait_001 (the operator's own "net" name for
# it, distinct from brobots_bait_002's "bait" - both were archived together 2026-07-22,
# un-archived 2026-07-23). Structurally NOT a robot_control_song_001-family score - its
# knobs.json uses the interview's own exchange shape (line_type/exchange_type/
# brobot_1_movement, bare_capture: true), one single canned exchange: both robots wake
# (arm cue, then head nod, reused verbatim from the interview's own gestures), each
# speaks its own self-naming line, done - no LLM, no vamp, no handoff, under a minute.
# Reuses run_section1_full_live_001.py unchanged via GOPOD_SECTION_SONG_DIR (the same
# song-dir seam 02_brobots_interview_run itself uses), never a second
# implementation. Zero alias reached this before today. NOT pha0b-sliceable, same
# structural reason "interview"/"preshow" are refused there (no standalone step-loop
# runner exists for this line-based shape) - see pha0b()'s own refusal case. Live by
# default, same convention as start-the-interview - no separate export step needed.
# Repointed 2026-07-24: brobots_bait_001 moved into zzz_archives/ during the operator's
# manual song-folder cleanup - this is its only alias, repointed rather than retired.
start-the-net-song() {
  ( cd /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools && \
    export GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 && \
    export GOPOD_SECTION_SONG_DIR=/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/zzz_archives/brobots_bait_001 && \
    python3 run_section1_full_live_001.py )
}

# start-the-bingo-capture retired 2026-08-11 (bingo golden-lock pass, operator
# direction: "no crucial use, prune"). It was a pure duplicate of
# bingo-video-song below - same score, same golden-engine run, only the name
# differed ("build/test" vs "video-facing") - and was still pointed at the
# legacy run_songs_runner_001.py, never cut over. No other caller found
# (checked brobots.sh, tech/, README.md, .claude/, gopod_probes/). Use
# bingo-video-song / bingo-video-song-live instead.

# Golden alias for the Bingo video shoot - the full 57-step score. Video-
# facing name, kept beside the build/test alias rather than replacing it.
# Dry by default - export GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 first for real
# hardware. No robot-select argument - the score itself alternates speakers.
# REPOINTED 2026-08-10 (drift fix, BINGO_GOLDEN_ABSORB_AND_EVAL_001.md item
# 1): was still calling the legacy run_songs_runner_001.py directly, while
# pha0b's own "bingo" case cut over to the golden engine
# (run_golden_song_001.py) back on 2026-08-07 - this alias never followed,
# so it was quietly running a different, stale show than what pha0b/phcal
# had already been proving golden. Now mirrors pha0b's own bingo dispatch:
# same runner, same GOPOD_GOLDEN_SONG_DIR export, same shared knobs
# resolver (so it plays whichever of knobs.json/zKnobs.json the resolver
# picks, same as every other golden-engine song).
bingo-video-song() {
  ( cd /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools && \
    export GOPOD_GOLDEN_SONG_DIR=/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/101_brobots_bingo_test && \
    python3 run_golden_song_001.py )
}

# Same run as bingo-video-song above, live by default - no env var to type
# first. Same pattern start-the-interview already uses (exports the live
# flag inside the alias itself). REPOINTED 2026-08-10, same fix/reason as
# bingo-video-song above.
bingo-video-song-live() {
  ( cd /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools && \
    export GOPOD_GOLDEN_SONG_DIR=/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/101_brobots_bingo_test && \
    GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1 python3 run_golden_song_001.py )
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
#   song:  bingo | control | bait | vamp | mixup | nap | itsyou-single | itsyou-multi
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

  # Interview has no step-slice flow at all (see _pha0b_interview_bypass's own
  # comment) - point_a/point_b don't apply, so it's exempted from the usual
  # "both required" gate below rather than forcing the caller to pass dummy
  # values just to get past it. Checked before the gate on purpose.
  if [ "$song" = "interview" ]; then
    if [ -n "$point_a" ] || [ -n "$point_b" ]; then
      echo "PHA0B_INTERVIEW_RANGE_IGNORED interview has no step-slice flow - a/b arguments ignored"
    fi
    _pha0b_interview_bypass
    return $?
  fi

  if [ -z "$song" ] || [ -z "$point_a" ] || [ -z "$point_b" ]; then
    echo "PHA0B_USAGE pha0b <song> <point_a_step_id> <point_b_step_id> [robot]"
    echo "  song: bingo | control | bait | vamp | mixup | nap | itsyou-single | itsyou-multi | interview"
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
      # goverlord/runtime/songs/101_brobots_bingo_test/ - the flagship UPSELL 1 song.
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
      song_dir_export="$songs_dir/101_brobots_bingo_test"
      ;;
    control)
      # Repointed 2026-07-29: this song was archived to
      # zzz_archives/robot_control_song_001 at some point and this case arm
      # never followed - same drift _score_song_dir's own control case
      # already corrected (see that comment, below). Matching that precedent.
      # CUTOVER PREP 2026-08-12 (studio tuning cut 2 step 2,
      # CONTROL_SONG_GOLDEN_CUTOVER_PREP_001.md): mirrors bait's own
      # 2026-08-07 cutover below - same control-family shape, already
      # golden-registered as "robot_control_song_001" (SONG_REGISTRY,
      # commit 16aa54e). song_dir_export unchanged; only runner/env_prefix
      # move to the golden engine's own fixed playhead namespace. PREPPED,
      # DRY-OK - not live-fire confirmed yet.
      runner="run_golden_song_001.py"
      env_prefix="GOPOD_GOLDEN_PLAYHEAD"
      song_dir_export="$songs_dir/zzz_archives/robot_control_song_001"
      ;;
    bait)
      # Repointed 2026-07-24: this song's own identity moved TO brobots_awaken
      # (top-level, not archived) - see start-the-bait-song's own comment above.
      # CUT OVER 2026-08-07 (golden-song-runner Ladder-1 Phase 3, fourth and
      # final cutover, mirrors nap/mixup/bingo above -
      # GOLDEN_SONG_RUNNER_PHASE3_BINGO_CUTOVER_001.md) from the legacy
      # run_robot_control_song_001.py to the golden engine
      # (run_golden_song_001.py) - already golden-registered as
      # "brobots_awaken_golden" (SONG_REGISTRY), the control-family shim
      # built in an earlier phase: fatal_notes=empty (never stops early),
      # manage_control=False (assumes control once at "connect", holds for
      # the whole run), note_aliases={head_nod: nod}, synthesize_speaker=True
      # (every step's speaker comes from a whole-run robot choice, not a
      # per-step field - see the new GOPOD_GOLDEN_ROBOT export below),
      # unconditional_text_speak=True. phcal-apply and reporter-gap prompts
      # are gated on $song, not $runner (same as bingo) - both still fire for
      # song=bait unchanged. CORRECTED 2026-08-11 (was stale): a zKnobs.json
      # sibling DOES exist for this song now (the dirty working-file layer,
      # same shared-resolver preference every other song already has) - the
      # earlier "no zKnobs.json sibling exists" claim here no longer holds.
      # KNOWN GAP, not fixed by this cutover: the golden engine's generic
      # playhead-WAKE prep (GOPOD_GOLDEN_PLAYHEAD_WAKE=1, always set live by
      # the shared dispatch block below) calls a bare connectivity ping
      # (run_wake_both/api-sdk/conn_test) for every song regardless -
      # correct for bingo-family (that IS what their own wake_both step
      # does). For THIS song, the legacy runner's own equivalent prep
      # (CONTROL_SONG_PLAYHEAD_WAKE_ENV) instead calls
      # assume_behavior_control, specifically so a range that starts AFTER
      # "connect" still has control assumed. The golden engine has no such
      # per-song WAKE-prep branch yet - a full-range run (starting AT
      # "connect", as tested here) is unaffected (the ping is just a benign
      # extra step before "connect" runs normally), but a future PARTIAL
      # range that skips "connect" would not correctly assume control the
      # way the legacy runner did. Flagged, not patched - out of this
      # cutover's own scope (repointing the button, not engine behavior).
      runner="run_golden_song_001.py"
      env_prefix="GOPOD_GOLDEN_PLAYHEAD"
      song_dir_export="$songs_dir/00_brobots_awaken"
      ;;
    vamp)
      # goverlord/runtime/songs/brobots_vamp_gate/ - standalone, playhead-
      # sliceable home for the pre-show's own vamp_1..vamp_4 filler beats
      # (PRESHOW_SONG_SURVEY_001.md). No physical robot, no
      # GOPOD_CONTROL_SONG_ROBOT - run_vamp_gate_song_001.py's own default
      # song dir already points here, no song_dir_export override needed
      # (same as bingo).
      runner="run_vamp_gate_song_001.py"
      env_prefix="GOPOD_VAMP_GATE_PLAYHEAD"
      ;;
    mixup)
      # goverlord/runtime/songs/102_brobots_cross_persona/ - the "is that you?"
      # cross-persona demo reel, wired 2026-07-31. Renamed from
      # 103_gopod_is_that_you (collided with the live is-that-you PTT demo/alias
      # it was derived from) - see story.md. CUT OVER 2026-08-07 (golden-song-
      # runner Ladder-1 Phase 3, second cutover, mirrors nap's own cutover
      # above - GOLDEN_SONG_RUNNER_PHASE3_NAP_CUTOVER_001.md) from the legacy
      # run_songs_runner_001.py to the golden engine
      # (run_golden_song_001.py) - already golden-registered as
      # "brobots_cross_persona" (SONG_REGISTRY, same bingo-family shape:
      # fatal_notes={wake_both,emotion_beat}, manage_control=True, no note
      # aliases) and already live-proven clean through the golden engine
      # directly (GOLDEN_SONG_RUNNER_PHASE1_WIDENING_001.md). No phcal-apply
      # or reporter-gap prompt is wired for "mixup" (both fire only for
      # song=bingo/bait - confirmed by grep, same as nap) - nothing extra to
      # preserve at this cutover. env_prefix changed to the golden engine's
      # own fixed playhead namespace (GOPOD_GOLDEN_PLAYHEAD_FROM/_TO/_WAKE),
      # same as nap - the shared dispatch block builds these generically off
      # env_prefix already. UNLIKE nap, mixup is NOT the golden engine's own
      # DEFAULT_SONG_ID (that's still brobots_baby_robots_sleep) - so
      # song_dir_export must actually reach the engine this time; see the
      # shared dispatch block below for the new GOPOD_GOLDEN_SONG_DIR export
      # this requires (added alongside the existing GOPOD_BINGO_CAPTURE_
      # SONG_DIR one, gated the same way on $runner).
      # ARCHIVED 2026-08-12: moved to zzz_archives/ (operator call - the real
      # 103_gopod_is_that_you live PTT+LLM test made this scripted demo reel
      # redundant). Keyword still reachable, path repointed, not retired -
      # same pattern as control above.
      runner="run_golden_song_001.py"
      env_prefix="GOPOD_GOLDEN_PLAYHEAD"
      song_dir_export="$songs_dir/zzz_archives/102_brobots_cross_persona"
      ;;
    itsyou-single)
      # goverlord/runtime/songs/103_gopod_is_that_you_single/ - live-capture
      # sibling of mixup above, reshaped 2026-08-08, SPLIT 2026-08-18
      # (operator's own folder surgery) from the old singular
      # 103_gopod_is_that_you/ into this single-robot-scoped sibling plus
      # itsyou-multi below. Scopes the live is-that-you() PTT test's own
      # live-capture bookend to KP1 only (Brobot 1/Doc) - see this song's own
      # story.md for the split's full reasoning. Wired straight onto the
      # golden engine (SONG_REGISTRY key "gopod_is_that_you_single", same
      # bingo-family shape as mixup). Keyword is "itsyou-single", not
      # "is-that-you" or "isthatyou" - deliberately non-colliding with the
      # live is-that-you() PTT demo alias this song's own bookends wrap
      # around, same reason mixup itself isn't named "is-that-you".
      runner="run_golden_song_001.py"
      env_prefix="GOPOD_GOLDEN_PLAYHEAD"
      song_dir_export="$songs_dir/103_gopod_is_that_you_single"
      ;;
    itsyou-multi)
      # goverlord/runtime/songs/104_gopod_is_that_you_multi/ - live-capture
      # sibling of mixup above, reshaped 2026-08-08, SPLIT 2026-08-18 from
      # the old singular 103_gopod_is_that_you/ into this sibling (the
      # golden, unscoped KP1+KP2 version) plus itsyou-single above. Wired
      # straight onto the golden engine (SONG_REGISTRY key
      # "gopod_is_that_you_multi", same bingo-family shape as mixup). Keyword
      # is "itsyou-multi", same non-collision reasoning as itsyou-single.
      runner="run_golden_song_001.py"
      env_prefix="GOPOD_GOLDEN_PLAYHEAD"
      song_dir_export="$songs_dir/104_gopod_is_that_you_multi"
      ;;
    nap)
      # goverlord/runtime/songs/105_brobots_nap/ (RENAMED/RENUMBERED
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
      # 105_brobots_nap in run_golden_song_001.py's own SONG_REGISTRY), so
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
      song_dir_export="$songs_dir/105_brobots_nap"
      ;;
    preshow)
      echo "PHA0B_REFUSED song=$song no playhead for this song yet (no standalone step-loop runner - see STUDIO_SONG_TOOL_SURVEY_001.md)"
      return 1
      ;;
    *)
      echo "PHA0B_REFUSED song=$song unknown - must be one of: bingo control bait vamp mixup nap itsyou-single itsyou-multi"
      return 1
      ;;
  esac

  # --- Shared prompts: live-robots gate, rich-display, phcal-apply       ---
  # --- tuning, reporter-gap override, robot-pick, robot-filter.         ---
  local live_gate
  live_gate="$(live_robots_prompt)"
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
  rich_display_live="$(_pha0b_prompt_yn "rich display on console? y/n [default y]: " "" n-means-no)"
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
  # copy is canonical, applied via apply_nongo_files.sh). A plain flag file
  # placed directly in webroot/ so the browser can fetch it statically,
  # re-read fresh on every poll, no caching. "n" here means the window stays
  # collapsed (0px) for this run; it is never expanded from any other source.
  local rich_ui_state_file="/home/goverlord/wire-pod/chipper/webroot/gopod_rich_display_ui_state.json"
  if [ "$rich_display_live" = "0" ]; then
    echo '{"expanded":false,"run_active":false}' > "$rich_ui_state_file"
  fi

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
    bingo) phcal_song_dir="$songs_dir/101_brobots_bingo_test" ;;
    bait) phcal_song_dir="$songs_dir/00_brobots_awaken" ;;
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
    echo "* default to 'y'"
    apply_live="$(_pha0b_prompt_yn "apply phcal tweaks to this selected range? y/n " y y-means-yes)"
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
        echo "PHA0B_PHCAL_APPLY applying tuned values to: $(echo "$qualifying_ids" | tr '\n' ' ')"
        while IFS= read -r sid; do
          [ -z "$sid" ] && continue
          python3 /home/goverlord/.gopod_alias_lib/phcal_apply_001.py "$sid" --yes --knobs "$phcal_apply_knobs"
        done <<< "$qualifying_ids"
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
  # (00_brobots_awaken): that song's SONG_REGISTRY entry sets
  # synthesize_speaker=True, meaning its own knobs.json "speaker" field is
  # unused dead data - the real speaker there comes from a whole-run robot
  # choice (GOPOD_GOLDEN_ROBOT/the `robot` argument), not per-step. Offering
  # this prompt for bait would silently write a field with zero playback
  # effect - excluded on that same honesty basis, not a missing feature.
  local robot_pick_song_dir=""
  case "$song" in
    nap|mixup|bingo|itsyou-single|itsyou-multi) robot_pick_song_dir="$song_dir_export" ;;
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

  # --- Assemble exports and dispatch to the picked runner. ---
  if [ "$rich_display_live" = "1" ]; then
    echo '{"expanded":true,"run_active":true}' > "$rich_ui_state_file"
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
  # fresh each pha0b call, matching phcal_run.log's own convention.
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
    if [ "$runner" = "run_golden_song_001.py" ] && [ "$robot_filter_choice" != "both" ]; then export GOPOD_GOLDEN_ROBOT="$robot_filter_choice"; fi && \
    if [ "$runner" = "run_golden_song_001.py" ]; then export GOPOD_GOLDEN_ROBOT_FILTER="$robot_filter_choice"; fi && \
    if [ "$live_gate" -eq 1 ]; then export GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1; fi && \
    if [ -n "$rich_display_override" ]; then export GOPOD_CONSOLE_RICH_DISPLAY="$rich_display_override"; fi && \
    python3 "$runner"
  ) 2> >(tee "$stderr_capture" >&2) | tee "/home/goverlord/.gopod_alias_lib/pha0b_run.log"
  local status="${PIPESTATUS[0]}"

  if [ "$rich_display_live" = "1" ]; then
    # run ended (success or fail, doesn't matter) - stay expanded, swap to
    # the tiny close button instead of auto-collapsing, per operator design.
    # finished_at (epoch seconds) lets the browser tell "just finished, show
    # it" apart from "stale leftover from a much older run" on a fresh page
    # load/refresh - see index.html's STALE_AFTER_SECONDS.
    echo "{\"expanded\":true,\"run_active\":false,\"finished_at\":$(date +%s)}" > "$rich_ui_state_file"
  fi

  if [ "$status" -ne 0 ]; then
    local blocked_line
    blocked_line="$(grep -o 'BLOCKED:.*' "$stderr_capture" | tail -1)"
    if [ -n "$blocked_line" ]; then
      echo "PHA0B_BLOCKED $blocked_line"
    else
      echo "PHA0B_FAILED unexpected error (see trace above)"
    fi
    rm -f "$stderr_capture"
    return 1
  fi
  rm -f "$stderr_capture"
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

  echo "Please note:"
  echo "  see songs/tools/README.md for the song-tools - run via their own aliases, not from here"

  echo "** anything entered outside selection range will exit **"
  local dir_choice
  dir_choice="$(_pha0b_prompt_range_choice "Pick a song [1-$dir_count, 0 to exit]: " "$dir_count" "PHA0B_MENU_BLOCKED")" || return 1
  local chosen_dir="${dir_array[$dir_choice]}"

  # --- Resolve the chosen folder name to pha0b's own song keyword. ---
  local song=""
  case "$chosen_dir" in
    101_brobots_bingo_test) song="bingo" ;;
    robot_control_song_001) song="control" ;;
    # 00_brobots_awaken now holds the merged bait/capture video (formerly
    # brobots_bait_002, absorbed into this name 2026-07-24; folder renumbered
    # 2026-08-01) - maps to "bait", not "weather", to match its real content.
    # The pure weather song was archived to zzz_archives/brobots_bait_000,
    # then decluttered off disk entirely 2026-08-15 - the "weather" pha0b
    # keyword itself dropped 2026-08-16 (dead reference, no live equivalent,
    # same purge as start-the-weather-song, README_NOTE_AND_WEATHER_SONG_PURGE_001.md).
    00_brobots_awaken) song="bait" ;;
    brobots_vamp_gate) song="vamp" ;;
    # 102_brobots_cross_persona - the "is that you?" cross-persona demo reel, wired
    # 2026-07-31, folder renumbered 2026-08-01. Renamed from 103_gopod_is_that_you
    # (collided with the live is-that-you PTT demo/alias it was derived from) -
    # see story.md.
    102_brobots_cross_persona) song="mixup" ;;
    # 105_brobots_nap - "Do Baby Robots Dream?" scored capture song, wired
    # 2026-08-06 (Phase 0 of the golden-song-runner plan). Folder renamed/
    # renumbered 2026-08-18 from 104_brobots_baby_robots_sleep.
    105_brobots_nap) song="nap" ;;
    # 103_gopod_is_that_you_single / 104_gopod_is_that_you_multi -
    # live-capture siblings of mixup, wired 2026-08-08, SPLIT 2026-08-18 from
    # the old singular 103_gopod_is_that_you/ folder. Keywords "itsyou-single"/
    # "itsyou-multi", not "is-that-you"/"isthatyou" - same non-collision
    # reason mixup itself isn't named "is-that-you" (the live PTT demo alias
    # already owns that name).
    103_gopod_is_that_you_single) song="itsyou-single" ;;
    104_gopod_is_that_you_multi) song="itsyou-multi" ;;
    # 102_brobots_bingo_game - NOT a note-sequence song, no playhead A/B slice
    # applies (2026-08-10, operator direction). This is the live, voice/touch-
    # triggered gobingo GAME itself (goverlord/runtime/songs/102_brobots_bingo_game/, a standalone Go
    # binary) - distinct from 101_brobots_bingo_test, the scripted performance/
    # capture song. Smallest possible special-case: picking this entry calls
    # the existing, already-working gobingo() function (demo.sh) verbatim -
    # same launch (serial/locale/silent flags, reactor co-launch) every other
    # gobingo caller already uses, not a reimplementation. Bypasses the
    # song="..."/point-A-B flow entirely below since there is no step
    # sequence to slice. Unlike every other pha0b entry, gobingo has no
    # dry-by-default gate of its own - it always fires the real binary
    # against real hardware (--silent only suppresses speech, not action) -
    # left as-is, not newly gated, per "reuse the existing launch verbatim."
    102_brobots_bingo_game)
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
    02_brobots_interview_run)
      # Repointed 2026-08-19 (INTERVIEW_VAMP_SPLIT_001.md) from the old
      # combined 01_brobots_interview_section_01. Bypasses the step-slice
      # flow below, same reason 102_brobots_bingo_game does above: this
      # song is interview-shaped (line_type/exchange_type/brobot_1_mode
      # fields), not note-dispatch - no step sequence to slice
      # (PHA0B_VAMP_INTEGRATION_SURVEY_001.md). Same shared bypass pha0b()'s own
      # direct `pha0b interview` call routes to now too - one door, one prompt,
      # not duplicated here (PHA0B_INTERVIEW_CONSOLIDATION_EXECUTED_001.md).
      _pha0b_interview_bypass
      return $?
      ;;
    01_brobots_interview_vamp)
      # New 2026-08-19 (INTERVIEW_VAMP_SPLIT_001.md): the pre-show/vamp,
      # split out of the old combined folder into its own standalone song.
      # Same structural reason as the RUN arm above - interview-shaped
      # steps, not note-dispatch, no slice flow. Repointed 2026-08-19
      # (INTERVIEW_VAMP_NO_GEN_PATH_001.md) from interview-vamp to
      # interview-vamp-play (function/alias names renamed 2026-08-19,
      # GOPOLISHER_FIXES_001.md - was vamp-run/preshow-run at the time this
      # repoint first landed) - picking this song from the menu is "play
      # video 1," which should not also silently kick off a full interview
      # generation as a side effect. interview-vamp itself is untouched
      # (still reachable via the interview's own bypass "roll a take"
      # choice, where the generation side effect is the actual point) -
      # this menu entry now plays the pre-show ONLY, zero interview
      # generation triggered.
      interview-vamp-play
      return $?
      ;;
    *)
      echo "PHA0B_MENU_REFUSED song_dir=$chosen_dir no playhead wired for this song yet under pha0b's own song keyword list (bingo|control|bait|vamp|mixup|nap|itsyou-single|itsyou-multi) - see pha0b()'s case statement above"
      return 1
      ;;
  esac

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
# (105_brobots_nap, formerly 104_brobots_baby_robots_sleep) added 2026-08-10
# (SHEET_MUSIC_GOLDEN_PROCESS_SURVEY_001.md finding #2 - both are on the
# current live shelf and both print correctly, but had no keyword here at
# all). Two keywords pha0b accepts are still refused here even though pha0b
# itself accepts them: interview (song_id brobots_interview_section_01, now
# housed in 02_brobots_interview_run/) and preshow
# (01_brobots_interview_vamp/, repointed 2026-08-19 - confirmed via gopod-vamp()'s own
# load_preshow_song() call) both tested clean (exit 0) against
# print_song_score_001.py, but their knobs.json steps carry no note/TEXT the
# way this printer's field-complete design expects - every step prints
# TEXT:(blank) even where the song plainly has real spoken lines, which is
# misleading rather than merely incomplete. bingo/control/bait/vamp/mixup/nap
# all confirmed to print real TEXT/note content.
# On success: echoes the resolved song directory to stdout only (safe to
# command-substitute). On failure: prints SCORE_REFUSED to stderr, returns 1,
# nothing on stdout.
# === SCORE - score/score-save: the bare notation-page printer =============
_score_song_dir() {
  local song="$1"
  local songs_dir="/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs"
  case "$song" in
    bingo)   printf '%s\n' "$songs_dir/101_brobots_bingo_test" ;;
    control)
      # pha0b's own song_dir_export for "control" points at
      # $songs_dir/robot_control_song_001 (non-archived) - that directory no
      # longer exists (moved under zzz_archives at some point, same as
      # vamp already is in pha0b's own code). Using pha0b's literal
      # stale path here would make this keyword always fail; using the real
      # current directory instead. Flagged, not fixed - pha0b/the runner's
      # own DEFAULT_SONG_DIR were both left untouched, out of scope here.
      printf '%s\n' "$songs_dir/zzz_archives/robot_control_song_001"
      ;;
    bait)    printf '%s\n' "$songs_dir/00_brobots_awaken" ;;
    vamp)    printf '%s\n' "$songs_dir/zzz_archives/brobots_vamp_gate" ;;
    mixup)   printf '%s\n' "$songs_dir/zzz_archives/102_brobots_cross_persona" ;;
    nap)     printf '%s\n' "$songs_dir/105_brobots_nap" ;;
    interview|preshow)
      echo "SCORE_REFUSED song=$song not wired for score printing yet (its knobs.json steps don't carry note/TEXT the way print_song_score_001.py expects - the page would show misleading blanks, not real content); available: bingo control bait vamp mixup nap" >&2
      return 1
      ;;
    *)
      echo "SCORE_REFUSED song=$song unknown - must be one of: bingo control bait vamp mixup nap" >&2
      return 1
      ;;
  esac
}

# score - bare one-word notation-page printer, screen only. No paths, no
# flags. Bare call defaults to brobots_bingo (the song currently being
# worked). Read-only against song data - print_song_score_001.py itself is
# never modified or written to by this alias, only invoked.
score() {
  local song="${1:-bingo}"
  local song_dir
  song_dir="$(_score_song_dir "$song")" || return 1
  ( cd "/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools" && \
    python3 print_song_score_001.py "$song_dir" )
}

# score-save - same song vocabulary as score() above, writes the page to
# that song's own notation/ directory instead of the screen (created if
# absent - sibling to that song's runs/ directory, same distinction
# NOTATION_PAGE_BUILD_001.md proposed: runs/ is machine-generated run
# output, notation/ is this operator-facing printed artifact). House dated
# filename convention. Prints the path it wrote, nothing else.
score-save() {
  local song="${1:-bingo}"
  local song_dir
  song_dir="$(_score_song_dir "$song")" || return 1
  local notation_dir="$song_dir/notation"
  mkdir -p "$notation_dir"
  local ts
  ts="$(date +%Y%m%d_%H%M%S)"
  local out_file="$notation_dir/${song}_score_${ts}.txt"
  ( cd "/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools" && \
    python3 print_song_score_001.py "$song_dir" ) > "$out_file"
  echo "SCORE_SAVED $out_file"
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

# phcal-promote - deliberate promote step for phcal's own master tweaks
# file (MASTER_TWEAKS_STAGE3_PROMOTE_001.md), mirrors run_golden_song_001.py's
# own _maybe_promote_knobs() shape (default-n y/n prompt, byte-identical
# copy, no re-serialize) - not a new mechanism. phcal_last.json stays the
# untracked working file (outside git entirely, same as before); this
# promotes a deliberate snapshot into tech/alias_play_studio/
# phcal_master_tweaks.json, a real tracked file, only on explicit "y".
# Not fired automatically at the end of every phcal call - phcal gets
# invoked many times per tuning session, unlike a song's single
# end-of-run - reachable on demand only. The actual y/n + copy both
# happen inside phcal_isolate_001.py itself; this is a thin entry point.
phcal-promote() {
  ( cd "$(dirname "${BASH_SOURCE[0]}")" && python3 phcal_isolate_001.py --promote )
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

# Interview movement rehearsal - fires every scored movement in the real
# interview score (goverlord/runtime/songs/02_brobots_interview_run/
# knobs.json, repointed 2026-08-19), in real playback order, both robots, with placeholder
# speech ("Interviewer Line."/"Interviewee Line.") instead of generated
# lines - no LLM, no generation wait. See
# INTERVIEW_MOVEMENT_REHEARSAL_TOOL_001.md. Dry by default - same
# convention as start-the-control-song above, this one does NOT export
# GOPOD_ALLOW_LIVE_ROBOT_SPEECH itself; export that gate yourself first for
# real hardware. No robot-select argument - both robots always play.
# === ROBOT STATE - test-interview-movements()/robot-sleep()/-wake()/ ======
# === -info(): rehearsal + direct-SDK sleep/wake/info tools. ===============
test-interview-movements() {
  ( cd /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools && \
    python3 run_interview_movement_rehearsal_001.py )
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
_ROBOT_SLEEP_DIRECT_BIN="/home/goverlord/wire-pod/chipper/gopod_probes/tools/direct_sdk_robot_sleep_001"

_robot_sleep_specs() {
  local which="${1:-0}"
  local s1="${GOPOD_BROBOT_1_SERIAL:-0dd1b9e9}"
  local s2="${GOPOD_BROBOT_2_SERIAL:-0dd1d8bf}"
  case "$which" in
    1) echo "Brobot_1:$s1" ;;
    2) echo "Brobot_2:$s2" ;;
    0|both) echo "Brobot_1:$s1 Brobot_2:$s2" ;;
    *) echo "ROBOT_SLEEP_BLOCKED bad_arg which=$which - must be 0 (both), 1, or 2" >&2; return 1 ;;
  esac
}

robot-sleep() {
  local specs
  specs="$(_robot_sleep_specs "${1:-both}")" || return 1
  local hold="${2:-5}"
  # timeout = 30 + hold: 10s buffer over the binary's own internal
  # 20+hold-second deadline (added 2026-07-25 after a live hang needed
  # Ctrl-C - same external-timeout precedent core.sh already uses for
  # direct_sdk_brobots_ready_001, belt and suspenders over the binary's own
  # internal deadline, not a replacement for it).
  WIREPOD_HOME=/home/goverlord/wire-pod timeout "$((30 + ${hold%.*}))" "$_ROBOT_SLEEP_DIRECT_BIN" sleep --hold "$hold" $specs
}

robot-wake() {
  local specs
  specs="$(_robot_sleep_specs "${1:-both}")" || return 1
  WIREPOD_HOME=/home/goverlord/wire-pod timeout 35 "$_ROBOT_SLEEP_DIRECT_BIN" wake $specs
}

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
_ROBOT_INFO_DIRECT_BIN="/home/goverlord/wire-pod/chipper/gopod_probes/tools/direct_sdk_robot_info_001"

robot-info() {
  local specs
  specs="$(_robot_sleep_specs "${1:-both}")" || return 1
  WIREPOD_HOME=/home/goverlord/wire-pod timeout 30 "$_ROBOT_INFO_DIRECT_BIN" $specs
}

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
_CUBE_BLIP_DIRECT_BIN="/home/goverlord/wire-pod/chipper/gopod_probes/tools/direct_sdk_cube_blip_001"
_CUBE_BLIP_DIRECT_SRC="/home/goverlord/wire-pod/chipper/gopod_probes/tools/direct_sdk_cube_blip_001.go"

cube-blip() {
  local serial="${1:-0dd1d8bf}"
  local mode="${2:-blip}"
  if [ ! -x "$_CUBE_BLIP_DIRECT_BIN" ] || [ "$_CUBE_BLIP_DIRECT_SRC" -nt "$_CUBE_BLIP_DIRECT_BIN" ]; then
    echo "CUBE_BLIP_BUILDING source=$_CUBE_BLIP_DIRECT_SRC"
    ( cd "$(dirname "$_CUBE_BLIP_DIRECT_SRC")" && go build -o "$_CUBE_BLIP_DIRECT_BIN" "$(basename "$_CUBE_BLIP_DIRECT_SRC")" ) || {
      echo "CUBE_BLIP_BUILD_FAILED"
      return 1
    }
  fi
  WIREPOD_HOME=/home/goverlord/wire-pod timeout 30 "$_CUBE_BLIP_DIRECT_BIN" "$serial" "$mode"
}

# cube-flash / cube-off - 2026-08-15, operator request: two more ways to
# fire the same binary's newer modes without typing the mode positionally.
# Live-confirmed 2026-08-14 that a lit cube self-expires on its own
# firmware timeout eventually, but not instantly - cube-off exists so the
# operator doesn't have to wait on that. cube-flash is the "show the
# control feature" demo beat: one deliberate on-then-off, not the full
# red-then-green blip.
cube-flash() {
  cube-blip "${1:-0dd1d8bf}" flash
}

cube-off() {
  cube-blip "${1:-0dd1d8bf}" off
}

# ==========================================================================
# SLEEP-LANE TEST BENCH - one thin alias per candidate beat, isolated fire
# (SLEEP_BENCH_ALIASES_001.md). Every name below is source-verified against
# the vendored fforchino/vector-go-sdk CLAD source (animationTrigger.clad,
# via gopod_notes/SDK_SOURCES_BACKUP_001/) and sdk-wrapper-animations.go (the
# actual go.mod-resolved version, v0.0.0-20231108155304-62168f3595d6) - none
# invented. Fires the new `play` action on direct_sdk_robot_sleep_001 (one
# named trigger or clip, held open for `hold` seconds, then released). No dry
# mode - same precedent as robot-sleep/robot-wake, this whole tool class
# always fires for real; there is nothing to simulate. Same argument
# convention as robot-sleep: which (1|2|0|both, default both), hold seconds
# (default 5). The operator fires each one himself, one at a time.
_sleep_bench_play() {
  local which="${1:-both}" hold="${2:-5}" kind="$3" name="$4"
  local specs
  specs="$(_robot_sleep_specs "$which")" || return 1
  local clip_flag=""
  [ "$kind" = "clip" ] && clip_flag="--clip"
  WIREPOD_HOME=/home/goverlord/wire-pod timeout "$((30 + ${hold%.*}))" "$_ROBOT_SLEEP_DIRECT_BIN" play --hold "$hold" $clip_flag "$name" $specs
}

# ---- SLEEP family: GoToSleep* triggers + anim_gotosleep_*/eyepose_asleep/face_sleeping clips ----
sleep-beat-get-in() { _sleep_bench_play "$1" "$2" trigger "GoToSleepGetIn"; }
sleep-beat-off() { _sleep_bench_play "$1" "$2" trigger "GoToSleepOff"; }
sleep-beat-sleeping() { _sleep_bench_play "$1" "$2" trigger "GoToSleepSleeping"; }
sleep-beat-getin-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_gotosleep_getin_01"; }
sleep-beat-sleeping-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_gotosleep_sleeping_01"; }
sleep-beat-sleeping-02-clip() { _sleep_bench_play "$1" "$2" clip "anim_gotosleep_sleeping_02"; }
sleep-beat-sleeping-03-clip() { _sleep_bench_play "$1" "$2" clip "anim_gotosleep_sleeping_03"; }
sleep-beat-sleeping-04-clip() { _sleep_bench_play "$1" "$2" clip "anim_gotosleep_sleeping_04"; }
sleep-beat-sleeping-05-clip() { _sleep_bench_play "$1" "$2" clip "anim_gotosleep_sleeping_05"; }
sleep-beat-sleeploop-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_gotosleep_sleeploop_01"; }
sleep-beat-getout-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_gotosleep_getout_01"; }
sleep-beat-getout-02-clip() { _sleep_bench_play "$1" "$2" clip "anim_gotosleep_getout_02"; }
sleep-beat-getout-03-clip() { _sleep_bench_play "$1" "$2" clip "anim_gotosleep_getout_03"; }
sleep-beat-getout-04-clip() { _sleep_bench_play "$1" "$2" clip "anim_gotosleep_getout_04"; }
sleep-beat-off-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_gotosleep_off_01"; }
sleep-beat-wakeup-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_gotosleep_wakeup_01"; }
sleep-beat-eyepose-asleep-clip() { _sleep_bench_play "$1" "$2" clip "anim_eyepose_asleep"; }
sleep-beat-face-sleeping-clip() { _sleep_bench_play "$1" "$2" clip "anim_face_sleeping"; }

# ---- RTS-ASLEEP family: reaction-to-sound-while-asleep triggers + anim_rtsound_*_asleep_* clips ----
sleep-rts-off-charger-ambient() { _sleep_bench_play "$1" "$2" trigger "RTS_OffCharger_Sleep_Ambient"; }
sleep-rts-off-charger-back() { _sleep_bench_play "$1" "$2" trigger "RTS_OffCharger_Sleep_Back"; }
sleep-rts-off-charger-front() { _sleep_bench_play "$1" "$2" trigger "RTS_OffCharger_Sleep_Front"; }
sleep-rts-off-charger-left() { _sleep_bench_play "$1" "$2" trigger "RTS_OffCharger_Sleep_Left"; }
sleep-rts-off-charger-right() { _sleep_bench_play "$1" "$2" trigger "RTS_OffCharger_Sleep_Right"; }
sleep-rts-off-charger-30-left() { _sleep_bench_play "$1" "$2" trigger "RTS_OffCharger_Sleep_30Left"; }
sleep-rts-off-charger-30-right() { _sleep_bench_play "$1" "$2" trigger "RTS_OffCharger_Sleep_30Right"; }
sleep-rts-off-charger-60-left() { _sleep_bench_play "$1" "$2" trigger "RTS_OffCharger_Sleep_60Left"; }
sleep-rts-off-charger-60-right() { _sleep_bench_play "$1" "$2" trigger "RTS_OffCharger_Sleep_60Right"; }
sleep-rts-off-charger-120-left() { _sleep_bench_play "$1" "$2" trigger "RTS_OffCharger_Sleep_120Left"; }
sleep-rts-off-charger-120-right() { _sleep_bench_play "$1" "$2" trigger "RTS_OffCharger_Sleep_120Right"; }
sleep-rts-off-charger-150-left() { _sleep_bench_play "$1" "$2" trigger "RTS_OffCharger_Sleep_150Left"; }
sleep-rts-off-charger-150-right() { _sleep_bench_play "$1" "$2" trigger "RTS_OffCharger_Sleep_150Right"; }
sleep-rts-on-charger-ambient() { _sleep_bench_play "$1" "$2" trigger "RTS_OnCharger_Sleep_Ambient"; }
sleep-rts-on-charger-back() { _sleep_bench_play "$1" "$2" trigger "RTS_OnCharger_Sleep_Back"; }
sleep-rts-on-charger-front() { _sleep_bench_play "$1" "$2" trigger "RTS_OnCharger_Sleep_Front"; }
sleep-rts-on-charger-left() { _sleep_bench_play "$1" "$2" trigger "RTS_OnCharger_Sleep_Left"; }
sleep-rts-on-charger-right() { _sleep_bench_play "$1" "$2" trigger "RTS_OnCharger_Sleep_Right"; }
sleep-rts-on-charger-30-left() { _sleep_bench_play "$1" "$2" trigger "RTS_OnCharger_Sleep_30Left"; }
sleep-rts-on-charger-30-right() { _sleep_bench_play "$1" "$2" trigger "RTS_OnCharger_Sleep_30Right"; }
sleep-rts-on-charger-60-left() { _sleep_bench_play "$1" "$2" trigger "RTS_OnCharger_Sleep_60Left"; }
sleep-rts-on-charger-60-right() { _sleep_bench_play "$1" "$2" trigger "RTS_OnCharger_Sleep_60Right"; }
sleep-rts-on-charger-120-left() { _sleep_bench_play "$1" "$2" trigger "RTS_OnCharger_Sleep_120Left"; }
sleep-rts-on-charger-120-right() { _sleep_bench_play "$1" "$2" trigger "RTS_OnCharger_Sleep_120Right"; }
sleep-rts-on-charger-150-left() { _sleep_bench_play "$1" "$2" trigger "RTS_OnCharger_Sleep_150Left"; }
sleep-rts-on-charger-150-right() { _sleep_bench_play "$1" "$2" trigger "RTS_OnCharger_Sleep_150Right"; }
sleep-rts-offcharger-ambient-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_offcharger_asleep_ambient_01"; }
sleep-rts-offcharger-behind-left-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_offcharger_asleep_behind_left_01"; }
sleep-rts-offcharger-behind-right-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_offcharger_asleep_behind_right_01"; }
sleep-rts-offcharger-front-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_offcharger_asleep_front_01"; }
sleep-rts-offcharger-left-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_offcharger_asleep_left_01"; }
sleep-rts-offcharger-right-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_offcharger_asleep_right_01"; }
sleep-rts-offcharger-30-left-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_offcharger_asleep_30left_01"; }
sleep-rts-offcharger-30-right-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_offcharger_asleep_30right_01"; }
sleep-rts-offcharger-60-left-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_offcharger_asleep_60left_01"; }
sleep-rts-offcharger-60-right-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_offcharger_asleep_60right_01"; }
sleep-rts-offcharger-120-left-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_offcharger_asleep_120left_01"; }
sleep-rts-offcharger-120-right-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_offcharger_asleep_120right_01"; }
sleep-rts-offcharger-150-left-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_offcharger_asleep_150left_01"; }
sleep-rts-offcharger-150-right-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_offcharger_asleep_150right_01"; }
sleep-rts-oncharger-ambient-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_oncharger_asleep_ambient_01"; }
sleep-rts-oncharger-behind-left-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_oncharger_asleep_behind_left_01"; }
sleep-rts-oncharger-behind-right-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_oncharger_asleep_behind_right_01"; }
sleep-rts-oncharger-front-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_oncharger_asleep_front_01"; }
sleep-rts-oncharger-left-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_oncharger_asleep_left_01"; }
sleep-rts-oncharger-right-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_oncharger_asleep_right_01"; }
sleep-rts-oncharger-30-left-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_oncharger_asleep_30left_01"; }
sleep-rts-oncharger-30-right-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_oncharger_asleep_30right_01"; }
sleep-rts-oncharger-60-left-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_oncharger_asleep_60left_01"; }
sleep-rts-oncharger-60-right-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_oncharger_asleep_60right_01"; }
sleep-rts-oncharger-120-left-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_oncharger_asleep_120left_01"; }
sleep-rts-oncharger-120-right-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_oncharger_asleep_120right_01"; }
sleep-rts-oncharger-150-left-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_oncharger_asleep_150left_01"; }
sleep-rts-oncharger-150-right-clip() { _sleep_bench_play "$1" "$2" clip "anim_rtsound_oncharger_asleep_150right_01"; }

# ---- HELD-ON-PALM family: HeldOnPalm* triggers + anim_heldonpalm_* clips ----
sleep-palm-edge-nervous() { _sleep_bench_play "$1" "$2" trigger "HeldOnPalmEdgeNervous"; }
sleep-palm-edge-relaxed() { _sleep_bench_play "$1" "$2" trigger "HeldOnPalmEdgeRelaxed"; }
sleep-palm-get-in-nervous() { _sleep_bench_play "$1" "$2" trigger "HeldOnPalmGetInNervous"; }
sleep-palm-get-in-relaxed() { _sleep_bench_play "$1" "$2" trigger "HeldOnPalmGetInRelaxed"; }
sleep-palm-react-to-jolt() { _sleep_bench_play "$1" "$2" trigger "HeldOnPalmReactToJolt"; }
sleep-palm-looking-nervous() { _sleep_bench_play "$1" "$2" trigger "HeldOnPalmLookingNervous"; }
sleep-palm-nestling() { _sleep_bench_play "$1" "$2" trigger "HeldOnPalmNestling"; }
sleep-palm-pickup-nervous() { _sleep_bench_play "$1" "$2" trigger "HeldOnPalmPickupNervous"; }
sleep-palm-pickup-relaxed() { _sleep_bench_play "$1" "$2" trigger "HeldOnPalmPickupRelaxed"; }
sleep-palm-put-down-nervous() { _sleep_bench_play "$1" "$2" trigger "HeldOnPalmPutDownNervous"; }
sleep-palm-put-down-relaxed() { _sleep_bench_play "$1" "$2" trigger "HeldOnPalmPutDownRelaxed"; }
sleep-palm-roll-off() { _sleep_bench_play "$1" "$2" trigger "HeldOnPalmRollOff"; }
sleep-palm-transition-to-relaxed() { _sleep_bench_play "$1" "$2" trigger "HeldOnPalmTransitionToRelaxed"; }
sleep-palm-edge-nervous-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_heldonpalm_edge_nervous_01"; }
sleep-palm-edge-relaxed-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_heldonpalm_edge_relaxed_01"; }
sleep-palm-getin-nervous-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_heldonpalm_getin_nervous_01"; }
sleep-palm-getin-relaxed-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_heldonpalm_getin_relaxed_01"; }
sleep-palm-jolt-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_heldonpalm_jolt_01"; }
sleep-palm-looking-nervous-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_heldonpalm_looking_nervous_01"; }
sleep-palm-nestling-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_heldonpalm_nestling_01"; }
sleep-palm-pickup-nervous-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_heldonpalm_pickup_nervous_01"; }
sleep-palm-pickup-relaxed-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_heldonpalm_pickup_relaxed_01"; }
sleep-palm-putdown-nervous-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_heldonpalm_putdown_nervous_01"; }
sleep-palm-putdown-relaxed-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_heldonpalm_putdown_relaxed_01"; }
sleep-palm-relaxed-idle-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_heldonpalm_relaxed_idle_01"; }
sleep-palm-rolloff-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_heldonpalm_rolloff_01"; }
sleep-palm-transition-2-relaxed-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_heldonpalm_transition2relaxed_01"; }

# ---- PETTING lane (kept visibly separate from the sleep lanes above, per spec) ----
# ---- PettingLevel1-4(+Getout)/BlissLoop/BlissGetout triggers + anim_petting_*/eyepose_bliss clips ----
# ---- pet-purr-bliss-loop is the purr alias (PettingBlissLoop trigger) ----
pet-level-1() { _sleep_bench_play "$1" "$2" trigger "PettingLevel1"; }
pet-level-2() { _sleep_bench_play "$1" "$2" trigger "PettingLevel2"; }
pet-level-3() { _sleep_bench_play "$1" "$2" trigger "PettingLevel3"; }
pet-level-4() { _sleep_bench_play "$1" "$2" trigger "PettingLevel4"; }
pet-level-1-getout() { _sleep_bench_play "$1" "$2" trigger "PettingLevel1Getout"; }
pet-level-2-getout() { _sleep_bench_play "$1" "$2" trigger "PettingLevel2Getout"; }
pet-level-3-getout() { _sleep_bench_play "$1" "$2" trigger "PettingLevel3Getout"; }
pet-level-4-getout() { _sleep_bench_play "$1" "$2" trigger "PettingLevel4Getout"; }
pet-purr-bliss-loop() { _sleep_bench_play "$1" "$2" trigger "PettingBlissLoop"; }
pet-bliss-getout() { _sleep_bench_play "$1" "$2" trigger "PettingBlissGetout"; }
pet-getin-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_petting_getin_01"; }
pet-lvl-1-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_petting_lvl1_01"; }
pet-lvl-2-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_petting_lvl2_01"; }
pet-lvl-3-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_petting_lvl3_01"; }
pet-lvl-4-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_petting_lvl4_01"; }
pet-lvl-1-getout-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_petting_lvl1_getout_01"; }
pet-lvl-2-getout-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_petting_lvl2_getout_01"; }
pet-lvl-2-getout-02-clip() { _sleep_bench_play "$1" "$2" clip "anim_petting_lvl2_getout_02"; }
pet-lvl-3-getout-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_petting_lvl3_getout_01"; }
pet-lvl-3-getout-02-clip() { _sleep_bench_play "$1" "$2" clip "anim_petting_lvl3_getout_02"; }
pet-lvl-4-getout-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_petting_lvl4_getout_01"; }
pet-blissloop-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_petting_blissloop_01"; }
pet-blissloop-02-clip() { _sleep_bench_play "$1" "$2" clip "anim_petting_blissloop_02"; }
pet-blissloop-03-clip() { _sleep_bench_play "$1" "$2" clip "anim_petting_blissloop_03"; }
pet-bliss-getout-01-clip() { _sleep_bench_play "$1" "$2" clip "anim_petting_bliss_getout_01"; }
pet-bliss-getout-02-clip() { _sleep_bench_play "$1" "$2" clip "anim_petting_bliss_getout_02"; }
pet-eyepose-bliss-clip() { _sleep_bench_play "$1" "$2" clip "anim_eyepose_bliss"; }

# ---- SLEEP-LANE SEGMENT RUNNERS - 6 review batches over the 126 bench beats ----
# (grouped on request, so a batch fires in one command instead of 126 separate
# ones). Each segment fires its members in the same order they're defined
# above, one after another - no dry mode, same as every alias it calls (each
# one blocks for real, for --hold seconds, before the next starts). This
# supersedes the block comment above's "operator fires each one himself, one
# at a time" for THIS use only - a segment is still one deliberate command the
# operator types and watches; nothing fires unattended.
_sleep_bench_segment() {
  local which="${1:-both}" hold="${2:-5}"; shift 2
  local total=$# i=1 name
  for name in "$@"; do
    echo "[$i/$total] $name (which=$which hold=${hold}s)"
    "$name" "$which" "$hold"
    i=$((i + 1))
  done
}

# 1) SLEEP core - 18: GoToSleep* triggers + clips
sleep-segment-core() {
  _sleep_bench_segment "$1" "$2" \
    sleep-beat-get-in sleep-beat-off sleep-beat-sleeping \
    sleep-beat-getin-01-clip sleep-beat-sleeping-01-clip sleep-beat-sleeping-02-clip \
    sleep-beat-sleeping-03-clip sleep-beat-sleeping-04-clip sleep-beat-sleeping-05-clip \
    sleep-beat-sleeploop-01-clip sleep-beat-getout-01-clip sleep-beat-getout-02-clip \
    sleep-beat-getout-03-clip sleep-beat-getout-04-clip sleep-beat-off-01-clip \
    sleep-beat-wakeup-01-clip sleep-beat-eyepose-asleep-clip sleep-beat-face-sleeping-clip
}

# 2) RTS-ASLEEP, off-charger - 27: triggers + clips
sleep-segment-rts-off() {
  _sleep_bench_segment "$1" "$2" \
    sleep-rts-off-charger-ambient sleep-rts-off-charger-back sleep-rts-off-charger-front \
    sleep-rts-off-charger-left sleep-rts-off-charger-right sleep-rts-off-charger-30-left \
    sleep-rts-off-charger-30-right sleep-rts-off-charger-60-left sleep-rts-off-charger-60-right \
    sleep-rts-off-charger-120-left sleep-rts-off-charger-120-right sleep-rts-off-charger-150-left \
    sleep-rts-off-charger-150-right \
    sleep-rts-offcharger-ambient-clip sleep-rts-offcharger-behind-left-clip sleep-rts-offcharger-behind-right-clip \
    sleep-rts-offcharger-front-clip sleep-rts-offcharger-left-clip sleep-rts-offcharger-right-clip \
    sleep-rts-offcharger-30-left-clip sleep-rts-offcharger-30-right-clip sleep-rts-offcharger-60-left-clip \
    sleep-rts-offcharger-60-right-clip sleep-rts-offcharger-120-left-clip sleep-rts-offcharger-120-right-clip \
    sleep-rts-offcharger-150-left-clip sleep-rts-offcharger-150-right-clip
}

# 3) RTS-ASLEEP, on-charger - 27: triggers + clips
sleep-segment-rts-on() {
  _sleep_bench_segment "$1" "$2" \
    sleep-rts-on-charger-ambient sleep-rts-on-charger-back sleep-rts-on-charger-front \
    sleep-rts-on-charger-left sleep-rts-on-charger-right sleep-rts-on-charger-30-left \
    sleep-rts-on-charger-30-right sleep-rts-on-charger-60-left sleep-rts-on-charger-60-right \
    sleep-rts-on-charger-120-left sleep-rts-on-charger-120-right sleep-rts-on-charger-150-left \
    sleep-rts-on-charger-150-right \
    sleep-rts-oncharger-ambient-clip sleep-rts-oncharger-behind-left-clip sleep-rts-oncharger-behind-right-clip \
    sleep-rts-oncharger-front-clip sleep-rts-oncharger-left-clip sleep-rts-oncharger-right-clip \
    sleep-rts-oncharger-30-left-clip sleep-rts-oncharger-30-right-clip sleep-rts-oncharger-60-left-clip \
    sleep-rts-oncharger-60-right-clip sleep-rts-oncharger-120-left-clip sleep-rts-oncharger-120-right-clip \
    sleep-rts-oncharger-150-left-clip sleep-rts-oncharger-150-right-clip
}

# 4) HELD-ON-PALM - 27: triggers + clips
sleep-segment-palm() {
  _sleep_bench_segment "$1" "$2" \
    sleep-palm-edge-nervous sleep-palm-edge-relaxed sleep-palm-get-in-nervous \
    sleep-palm-get-in-relaxed sleep-palm-react-to-jolt sleep-palm-looking-nervous \
    sleep-palm-nestling sleep-palm-pickup-nervous sleep-palm-pickup-relaxed \
    sleep-palm-put-down-nervous sleep-palm-put-down-relaxed sleep-palm-roll-off \
    sleep-palm-transition-to-relaxed \
    sleep-palm-edge-nervous-01-clip sleep-palm-edge-relaxed-01-clip sleep-palm-getin-nervous-01-clip \
    sleep-palm-getin-relaxed-01-clip sleep-palm-jolt-01-clip sleep-palm-looking-nervous-01-clip \
    sleep-palm-nestling-01-clip sleep-palm-pickup-nervous-01-clip sleep-palm-pickup-relaxed-01-clip \
    sleep-palm-putdown-nervous-01-clip sleep-palm-putdown-relaxed-01-clip sleep-palm-relaxed-idle-01-clip \
    sleep-palm-rolloff-01-clip sleep-palm-transition-2-relaxed-01-clip
}

# 5) PETTING triggers - 10
sleep-segment-pet-triggers() {
  _sleep_bench_segment "$1" "$2" \
    pet-level-1 pet-level-2 pet-level-3 pet-level-4 \
    pet-level-1-getout pet-level-2-getout pet-level-3-getout pet-level-4-getout \
    pet-purr-bliss-loop pet-bliss-getout
}

# 6) PETTING clips - 17
sleep-segment-pet-clips() {
  _sleep_bench_segment "$1" "$2" \
    pet-getin-01-clip pet-lvl-1-01-clip pet-lvl-2-01-clip pet-lvl-3-01-clip pet-lvl-4-01-clip \
    pet-lvl-1-getout-01-clip pet-lvl-2-getout-01-clip pet-lvl-2-getout-02-clip \
    pet-lvl-3-getout-01-clip pet-lvl-3-getout-02-clip pet-lvl-4-getout-01-clip \
    pet-blissloop-01-clip pet-blissloop-02-clip pet-blissloop-03-clip \
    pet-bliss-getout-01-clip pet-bliss-getout-02-clip pet-eyepose-bliss-clip
}

# gopod-song-open - the "golden sequence" (operator-designed 2026-07-25):
# synced sleep -> warm-up (wpr) runs while both robots hold asleep -> synced
# wake the instant warm-up finishes -> caller starts the song. Event-driven,
# not a fixed guessed hold - built on direct_sdk_robot_sleep_001.go's
# --wait-signal mode (ROBOT_SLEEP_DIRECT_SDK_BUILT_001.md), which keeps both
# robots' BehaviorControl connections open continuously from asleep straight
# through to firing GoToSleepOff - no release-then-regrant gap in between
# (that gap is exactly what caused this same tool's very first live test to
# have both robots wake right back up on their own). No initial wake call
# here - operator's own call ("not important"): this starts straight into
# the synced sleep. Confirmed live 2026-07-25: sleep-fire sync at 104
# microseconds apart, GoToSleepOff genuinely rouses a held-asleep robot -
# this function is not yet itself live-tested, only compiled/syntax-checked.
# === GOLDEN-SEQUENCE OPENERS - gopod-song-open family: sleep both robots ==
# === in sync, do real fill-work while they hold asleep, wake together. ====
gopod-song-open() {
  local which="${1:-both}"
  local specs
  specs="$(_robot_sleep_specs "$which")" || return 1

  # mktemp -u only generates a name, does not create the file - the whole
  # point of --wait-signal is that direct_sdk_robot_sleep_001 must never see
  # this path exist until wpr below is actually done. A fresh, unique name
  # per invocation means a stale leftover from a prior run can never fool
  # this run into skipping the wait instantly.
  local signal_file
  signal_file="$(mktemp -u /tmp/gopod_song_open_signal.XXXXXX)"
  rm -f "$signal_file"

  echo "GOPOD_SONG_OPEN_START which=$which signal_file=$signal_file"

  # Backgrounded: fires the synced sleep, then blocks (event-driven) until
  # signal_file appears. Timeout=70: the binary's own internal deadline
  # widens to 20+max-wait (default 60s) once --wait-signal is given, same
  # 10s-buffer precedent robot-sleep/robot-wake already use above.
  WIREPOD_HOME=/home/goverlord/wire-pod timeout 70 "$_ROBOT_SLEEP_DIRECT_BIN" sleep --wait-signal "$signal_file" $specs &
  local sleep_pid=$!

  echo "GOPOD_SONG_OPEN_WARMUP_START"
  wpr check
  local wpr_status=$?
  echo "GOPOD_SONG_OPEN_WARMUP_DONE status=$wpr_status"

  # The touch itself IS the signal - direct_sdk_robot_sleep_001's
  # watchSignalFile polls for this path's existence, never its contents.
  touch "$signal_file"

  wait "$sleep_pid"
  local sleep_status=$?
  rm -f "$signal_file"

  echo "GOPOD_SONG_OPEN_DONE wpr_status=$wpr_status sleep_status=$sleep_status"
  [ "$sleep_status" -eq 0 ]
}

# gopod-song-open-chord - "chord inside the sleep window" (operator-requested
# 2026-07-25, same day as gopod-song-open): runs the chord's own
# robot-independent warm-up work (mic-set, LLM ping, Kokoro warm-up) as the
# real fill-work during the golden-sequence sleep hold, instead of a bare
# `wpr` that no-ops when Wire-Pod is already healthy.
#
# Deliberately does NOT call gopod_brobots() as a whole, and deliberately
# excludes its Wire-Pod restart+wake stage (_gopod_chord_wirepod_job) -
# LIVE-PROBED 2026-07-25 (see gopod_notes/SONG_OPEN_CHORD_INSIDE_001.md):
# firing Wire-Pod's own /api-sdk/assume_behavior_control (OVERRIDE_BEHAVIORS
# priority, bcassume.go) against a robot this tool already holds asleep via
# a separate direct-SDK BehaviorControl connection (same OVERRIDE_BEHAVIORS
# priority, direct_sdk_robot_sleep_001.go) made the sleep binary HANG
# indefinitely - it never returned, had to be force-killed by the outer
# `timeout` after the full window elapsed, and the robot was left in an
# unclean state until a plain `wake` call recovered it. Real, live-confirmed,
# not theoretical. Same collision would hit both robots' wake steps inside
# the full chord, so the whole Wire-Pod stage is excluded, not just
# softened.
#
# gopod_brobots()/core.sh is UNMODIFIED - this reuses its already-standalone
# sub-pieces (gopod-mic-set, _gopod_llm_ping, _gopod_chord_kokoro_job), the
# same three notes gopod_brobots() itself backgrounds, run the same way
# (parallel, `wait`ed, status read back from each note's own exit code) - no
# reimplementation, no new logic.
gopod-song-open-chord() {
  local which="${1:-both}"
  local specs
  specs="$(_robot_sleep_specs "$which")" || return 1

  local signal_file
  signal_file="$(mktemp -u /tmp/gopod_song_open_chord_signal.XXXXXX)"
  rm -f "$signal_file"

  echo "GOPOD_SONG_OPEN_CHORD_START which=$which signal_file=$signal_file"

  WIREPOD_HOME=/home/goverlord/wire-pod timeout 70 "$_ROBOT_SLEEP_DIRECT_BIN" sleep --wait-signal "$signal_file" $specs &
  local sleep_pid=$!

  export -f gopod-mic-detect gopod-mic-set _gopod_llm_ping >/dev/null 2>&1

  local workdir
  workdir="$(mktemp -d)"

  echo "GOPOD_SONG_OPEN_CHORD_WARMUP_START"
  ( timeout 40 bash -c 'gopod-mic-set'; echo $? > "$workdir/mic.status" ) > "$workdir/mic.log" 2>&1 &
  local mic_pid=$!

  ( timeout 60 bash -c '_gopod_llm_ping "$1" "$2"' _ "${GOPOD_BROBOT_LLM:-gemma2:2b}" "Say GOPOD ready in one short sentence."; echo $? > "$workdir/llm.status" ) > "$workdir/llm.log" 2>&1 &
  local llm_pid=$!

  ( _gopod_chord_kokoro_job; echo $? > "$workdir/kokoro.status" ) > "$workdir/kokoro.log" 2>&1 &
  local kokoro_pid=$!

  wait "$mic_pid" "$llm_pid" "$kokoro_pid"

  local mic_rc llm_rc kokoro_rc
  mic_rc="$(cat "$workdir/mic.status" 2>/dev/null || echo 1)"
  llm_rc="$(cat "$workdir/llm.status" 2>/dev/null || echo 1)"
  kokoro_rc="$(cat "$workdir/kokoro.status" 2>/dev/null || echo 1)"
  echo "GOPOD_SONG_OPEN_CHORD_WARMUP_DONE mic_rc=$mic_rc llm_rc=$llm_rc kokoro_rc=$kokoro_rc"

  # A failed warm-up note still releases the robots - they must never stay
  # asleep because mic/LLM/Kokoro errored.
  touch "$signal_file"

  wait "$sleep_pid"
  local sleep_status=$?
  rm -f "$signal_file"
  rm -rf "$workdir"

  echo "GOPOD_SONG_OPEN_CHORD_DONE mic_rc=$mic_rc llm_rc=$llm_rc kokoro_rc=$kokoro_rc sleep_status=$sleep_status"
  [ "$sleep_status" -eq 0 ]
}

# gopod-song-open-chord-sleep-first - "SLEEP-FIRST OPENING" (operator-specced
# 2026-07-26), three beats in order:
#   1) wake-call, then robots sleep in sync
#   2) mic/LLM/Kokoro/Wire-Pod-restart warm-up while robots hold the sleep
#   3) on release (the sleep binary's own GoToSleepOff is the wake), the
#      synced "Brobots ready!" finale - AFTER the release, never inside the
#      sleep window.
#
# Per SONG_OPEN_CHORD_INSIDE_001.md's live-confirmed hang: anything that
# speaks to or assumes control of a robot cannot run while
# direct_sdk_robot_sleep_001 holds its own BehaviorControl grant. core.sh's
# _gopod_chord_wirepod_job bundles a Wire-Pod restart (not robot-touching)
# and a wake+speak loop (robot-touching) in one indivisible python process -
# not separable at the bash-function level without editing core.sh. So
# core.sh/gopod_brobots() are UNMODIFIED; the job function below imports the
# same run_section1_full_live_001 module core.sh already uses and calls only
# the restart half, for beat 2. Beat 3's finale reuses core.sh's
# already-standalone _gopod_chord_release_both / _gopod_chord_direct_together_job
# unmodified - the operator's spec is one synced readiness message, not a
# per-robot ready line, so beat 3 carries nothing else.
_gopod_sleep_first_wirepod_restart_job() {
  timeout 150 python3 - <<PYEOF
import importlib.util
import sys

spec = importlib.util.spec_from_file_location("run_section1_full_live_001", "$_GOPOD_CHORD_RUNNER_PATH")
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)

print("STATUS: WIREPOD_CHAIN_RESTART_START", flush=True)
try:
    result = mod.restart_wirepod_preflight(True)
    print(f"STATUS: WIREPOD_CHAIN_RESTART_DONE {result}", flush=True)
except Exception as exc:  # noqa: BLE001
    print(f"STATUS: WIREPOD_CHAIN_RESTART_ERROR {type(exc).__name__}: {exc}", flush=True)
    sys.exit(1)
PYEOF
}

_GOPOD_SLEEP_FIRST_WAKE_SETTLE_SECONDS="${GOPOD_SLEEP_FIRST_WAKE_SETTLE_SECONDS:-3}"

gopod-song-open-chord-sleep-first() {
  local which="${1:-both}"
  local specs
  specs="$(_robot_sleep_specs "$which")" || return 1

  # Preflight sweep: kill any leftover instance of the sleep binary from a
  # prior invocation (crashed, backgrounded past a Ctrl-C, whatever) before
  # this run touches the robots at all. A surviving one holds its own EP
  # connection and collides with this run's wake/sleep calls - this is what
  # actually caused the "robots stayed awake during warm-up" symptom, not
  # anything about beat order. Runs every time, no operator action needed.
  local _leftover_pids
  _leftover_pids="$(pgrep -f "$_ROBOT_SLEEP_DIRECT_BIN sleep --wait-signal" 2>/dev/null)"
  if [ -n "$_leftover_pids" ]; then
    echo "GOPOD_SLEEP_FIRST_PREFLIGHT: killing leftover sleep-binary pid(s): $(echo "$_leftover_pids" | tr '\n' ' ')"
    kill $_leftover_pids 2>/dev/null
    sleep 0.5
    kill -9 $_leftover_pids 2>/dev/null
  fi
  rm -f /tmp/gopod_sleep_first_signal.* 2>/dev/null

  # Ctrl-C (or a kill) must not leave background jobs (the sleep binary,
  # mic/llm/kokoro/wirepod warm-up) running - an orphaned one collides with
  # the next invocation's own robot connections. Trap cleans up, then clears
  # itself so it doesn't linger for unrelated commands in this shell.
  local _gopod_sleep_first_bg_pids=()
  _gopod_sleep_first_kill_bg() {
    local pid
    for pid in "${_gopod_sleep_first_bg_pids[@]}"; do
      kill "$pid" 2>/dev/null
    done
  }
  trap '_gopod_sleep_first_kill_bg; trap - INT TERM; return 130' INT TERM

  # --- Beat 1: wake-call, wake-orientation settle, then sleep in sync ---
  echo "GOPOD_SLEEP_FIRST_START which=$which"
  robot-wake "$which"
  local wakecall_rc=$?
  sleep "$_GOPOD_SLEEP_FIRST_WAKE_SETTLE_SECONDS"

  local signal_file
  signal_file="$(mktemp -u /tmp/gopod_sleep_first_signal.XXXXXX)"
  rm -f "$signal_file"

  WIREPOD_HOME=/home/goverlord/wire-pod timeout 70 "$_ROBOT_SLEEP_DIRECT_BIN" sleep --wait-signal "$signal_file" $specs &
  local sleep_pid=$!
  _gopod_sleep_first_bg_pids+=("$sleep_pid")

  # --- Beat 2: mic/LLM/Kokoro/Wire-Pod-restart warm-up while asleep ---
  export -f gopod-mic-detect gopod-mic-set _gopod_llm_ping >/dev/null 2>&1

  local workdir
  workdir="$(mktemp -d)"

  ( timeout 40 bash -c 'gopod-mic-set'; echo $? > "$workdir/mic.status" ) > "$workdir/mic.log" 2>&1 &
  local mic_pid=$!
  _gopod_sleep_first_bg_pids+=("$mic_pid")

  ( timeout 60 bash -c '_gopod_llm_ping "$1" "$2"' _ "${GOPOD_BROBOT_LLM:-gemma2:2b}" "Say GOPOD ready in one short sentence."; echo $? > "$workdir/llm.status" ) > "$workdir/llm.log" 2>&1 &
  local llm_pid=$!
  _gopod_sleep_first_bg_pids+=("$llm_pid")

  ( _gopod_chord_kokoro_job; echo $? > "$workdir/kokoro.status" ) > "$workdir/kokoro.log" 2>&1 &
  local kokoro_pid=$!
  _gopod_sleep_first_bg_pids+=("$kokoro_pid")

  ( _gopod_sleep_first_wirepod_restart_job; echo $? > "$workdir/wirepod.status" ) > "$workdir/wirepod.log" 2>&1 &
  local wirepod_pid=$!
  _gopod_sleep_first_bg_pids+=("$wirepod_pid")

  wait "$mic_pid" "$llm_pid" "$kokoro_pid" "$wirepod_pid"

  local mic_rc llm_rc kokoro_rc wirepod_ready
  mic_rc="$(cat "$workdir/mic.status" 2>/dev/null || echo 1)"
  llm_rc="$(cat "$workdir/llm.status" 2>/dev/null || echo 1)"
  kokoro_rc="$(cat "$workdir/kokoro.status" 2>/dev/null || echo 1)"
  wirepod_ready=1
  grep -q "WIREPOD_CHAIN_RESTART_DONE" "$workdir/wirepod.log" 2>/dev/null && wirepod_ready=0

  # A failed warm-up note still releases the robots - never left asleep
  # because a beat-2 stage errored.
  touch "$signal_file"

  wait "$sleep_pid"
  local sleep_status=$?
  rm -f "$signal_file"

  # --- Beat 3: on release, the sleep binary's own GoToSleepOff (already
  # fired as part of the sleep action's release path) is the wake - no
  # separate wake+speak step. Just the synced "Brobots ready!" finale below.

  local note_line
  note_line() {
    local name="$1" rc="$2"
    if [ "$rc" = "0" ]; then
      echo "NOTE $name: done"
    else
      echo "NOTE $name: failed"
    fi
  }

  note_line "wake_call" "$wakecall_rc"
  note_line "mic_set" "$mic_rc"
  note_line "llm_warm" "$llm_rc"
  note_line "kokoro_warm" "$kokoro_rc"
  note_line "wirepod_ready" "$wirepod_ready"
  note_line "sleep_sync" "$sleep_status"

  local failed=()
  [ "$mic_rc" = "0" ] || failed+=("mic_set")
  [ "$llm_rc" = "0" ] || failed+=("llm_warm")
  [ "$kokoro_rc" = "0" ] || failed+=("kokoro_warm")
  [ "$wirepod_ready" = "0" ] || failed+=("wirepod_ready")
  [ "$sleep_status" = "0" ] || failed+=("sleep_sync")
  # wake_call is best-effort and its live effect is an open question (see
  # SLEEP_FIRST_OPENING_001.md item 3) - logged, never gates NOT-READY.

  # The finale, same shape gopod_brobots() itself uses - only for both
  # robots, only once every other note above is done.
  if [ "${#failed[@]}" -eq 0 ]; then
    case "$which" in
      both|0)
        local together_dir
        together_dir="$(mktemp -d)"
        _gopod_chord_release_both
        sleep "$_GOPOD_CHORD_RELEASE_SETTLE_SECONDS"
        local together_rc
        _gopod_chord_direct_together_job "$together_dir/direct.log"
        together_rc=$?
        cat "$together_dir/direct.log"

        if grep -q "TOGETHER_DIRECT label=Brobot 1" "$together_dir/direct.log" 2>/dev/null && \
           grep -q "TOGETHER_DIRECT label=Brobot 2" "$together_dir/direct.log" 2>/dev/null; then
          python3 -c "
import re
text = open('$together_dir/direct.log').read()
rows = {}
for m in re.finditer(r'TOGETHER_DIRECT label=(.+?) serial=(\S+) start=([\d.]+) end=([\d.]+) elapsed=([\d.]+)s status=(\S+)', text):
    label, serial, start, end, elapsed, status = m.groups()
    rows[label] = {'start': float(start), 'end': float(end)}
b1, b2 = rows.get('Brobot 1'), rows.get('Brobot 2')
if b1 and b2:
    overlap = min(b1['end'], b2['end']) - max(b1['start'], b2['start'])
    gap = abs(b1['start'] - b2['start'])
    print(f\"TOGETHER (direct SDK) start-gap={gap:.3f}s overlap={overlap:.3f}s\" + (\" (genuinely overlapped)\" if overlap > 0 else \" (did not overlap)\"))
"
        fi
        rm -rf "$together_dir"

        local brobots_together=1
        [ "$together_rc" = "0" ] && brobots_together=0
        note_line "brobots_together" "$brobots_together"
        [ "$brobots_together" = "0" ] || failed+=("brobots_together")
        ;;
    esac
  fi

  rm -rf "$workdir"
  trap - INT TERM

  if [ "${#failed[@]}" -eq 0 ]; then
    echo "READY - all notes played, stage is set."
    return 0
  else
    local joined
    joined="$(IFS=,; echo "${failed[*]}")"
    echo "NOT-READY-because-$joined"
    return 1
  fi
}

# === SESSION TOOLS - gopod-conn-test/-vamp/-pick-model: standalone =========
# === front doors onto pieces the opening chord/interview already do. ======
gopod-conn-test() {
  local which="${1:-both}"
  local s1="${GOPOD_BROBOT_1_SERIAL:-0dd1b9e9}"
  local s2="${GOPOD_BROBOT_2_SERIAL:-0dd1d8bf}"
  local failed=0
  local out

  if [ "$which" = "1" ] || [ "$which" = "both" ]; then
    echo "GOPOD_CONN_TEST_START robot=Brobot_1 serial=$s1"
    out="$(_gopod_note_send "/api-sdk/conn_test" "serial=$s1")"
    echo "$out"
    if echo "$out" | grep -q "body='success'"; then
      echo "GOPOD_CONN_TEST_PASS robot=Brobot_1"
    else
      echo "GOPOD_CONN_TEST_FAIL robot=Brobot_1"
      failed=1
    fi
  fi

  if [ "$which" = "2" ] || [ "$which" = "both" ]; then
    echo "GOPOD_CONN_TEST_START robot=Brobot_2 serial=$s2"
    out="$(_gopod_note_send "/api-sdk/conn_test" "serial=$s2")"
    echo "$out"
    if echo "$out" | grep -q "body='success'"; then
      echo "GOPOD_CONN_TEST_PASS robot=Brobot_2"
    else
      echo "GOPOD_CONN_TEST_FAIL robot=Brobot_2"
      failed=1
    fi
  fi

  return "$failed"
}

# gopod-vamp - standalone preview of the pre-show's own vamp gate (the
# scored vamp_1..vamp_4 filler beats run_preshow_song loops while interview
# generation is still in progress - see run_section1_full_live_001.py
# :3641-3667). Before this alias, the only way to hear these four lines at
# all was to run the whole pre-show song and hope generation was still
# running when the gate opened (a real, run-dependent timing race, not a
# rehearsal). All four vamp beats are brobot_3/brobot_4 (voice-only hosts,
# Kokoro, no physical robot, no LLM - confirmed from knobs.json this
# session), so this calls load_preshow_song() + _preshow_speak_host() -
# the exact same loader and speak function the real vamp loop calls -
# directly, with no fake generation_done_event and no robot/scaffold setup
# needed. Argument: number of full cycles through the 4 beats (default 1 =
# vamp_1..vamp_4 once, in order).
gopod-vamp() {
  local cycles="${1:-1}"
  python3 -c "
import importlib.util
import sys

cycles = int('$cycles')
spec = importlib.util.spec_from_file_location('run_section1_full_live_001', '$_GOPOD_NOTE_RUNNER_PATH')
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)

song = mod.load_preshow_song('/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/01_brobots_interview_vamp')
vamp_steps = [s for s in song['steps'] if s['movement'] == 'vamp']
if not vamp_steps:
    print('GOPOD_VAMP_NO_STEPS_FOUND')
    sys.exit(1)

total = cycles * len(vamp_steps)
print(f'GOPOD_VAMP_START cycles={cycles} steps={len(vamp_steps)} total_lines={total}')
for i in range(total):
    step = vamp_steps[i % len(vamp_steps)]
    mod._preshow_speak_host(step, read_sheet=False, segment_id=f'gopod_vamp_standalone_round{i+1}_{step[\"step_id\"]}')
print('GOPOD_VAMP_DONE')
"
}

# gopod-pick-model - standalone front door onto resolve_content_model(),
# per ALIAS_REGISTRY_TRUTH_SWEEP_001.md section 5's own finding: the model
# selector is already a golden, live-proven process (LLM_MODEL_SELECTOR_001.md,
# runs A-D), just not isolated as its own alias. This is that wrapper, no new
# selection logic - same env var (GOPOD_CONTENT_MODEL) skips the menu
# unattended, same remembered-model state file (repointed 2026-08-19,
# INTERVIEW_VAMP_SPLIT_001.md:
# goverlord/runtime/songs/02_brobots_interview_run/content_model_state.json).
# Uses `python3 -c`, not a
# heredoc piped over stdin - resolve_content_model() may call input() for
# the interactive menu, and a heredoc would already have consumed stdin as
# the script source, leaving input() nothing to read.
gopod-pick-model() {
  python3 -c "
import importlib.util
spec = importlib.util.spec_from_file_location('run_section1_full_live_001', '$_GOPOD_NOTE_RUNNER_PATH')
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)
chosen = mod.resolve_content_model()
print(f'GOPOD_CONTENT_MODEL_SELECTED model={chosen}')
"
}

# test-silent-angry-say - the "silent safe say" alt-path for a single angry
# animation, via say_text's {{playAnimationWI||angry}}... action-tag lane
# (async/fire-and-forget playAnimationWI, the same token every brobots-anim-*
# note in this file already uses) - built as an alternative to brobots-angry's
# own broken /api-sdk/play_animation route (see ALIAS_REGISTRY_POLISH_001.md
# Decision 3 - that route 404s). Same single instrument as every other note
# in this file (_gopod_note_send), no reimplementation.
#
# KNOWN LIVE FINDING, 2026-07-16 - the original blocking playAnimation token
# CRASHED Brobot 2 on its second live fire: animation played, but Wire-Pod's
# own log showed a blocking "(waiting for animation to be done...)" line
# followed by 15+ seconds of no conn-check response from the robot before it
# recovered on its own. See BROBOT2_SILENT_ANGRY_SAY_001.md in gopod_notes/
# for the full timeline.
#
# FIXED, same day - token swapped from playAnimation (blocking) to
# playAnimationWI (async) per that finding: the blocking variant is the
# prime suspect and is now retired from this alias entirely; the async
# variant has never crashed a robot across every brobots-anim-* note that
# already uses it. See SILENT_ANGRY_SAY_ASYNC_FIX_001.md for the live
# re-verification.
#
# HARDENED, same day - operator-observed finding: "HTTP success but no
# playback on cold first press; second press played." A wake step (the
# same /api-sdk/conn_test call gopod-conn-test already uses, reused here
# via _gopod_note_send - no new HTTP client) now fires before the
# animation send, followed by a 1.5s settle pause, then the existing
# assume -> say_text -> release sequence unchanged. Response is that a
# cold first press wakes the robot's own connection before asking it to
# play anything, rather than asking it to play on a connection that
# wasn't really live yet. Root cause of the cold-first-press drop itself
# was not investigated - this is a wake workaround, not a diagnosis. See
# SILENT_ANGRY_SAY_WAKE_STEP_001.md for the live re-verification.
#
# HOLD ALIGNED, same day - the animation hold here was still the original
# 2.5s guess from before any of the above was tuned. Aligned to 5.0s, the
# operator's own proven-safe value (see test-reaction-in-the-beat's own
# comment block for that history) - no more stale, un-reviewed hold values
# left sitting in this file.
#
# This is a TEST alias, not a production note, not wired into any song or
# the chord. Fire deliberately, one robot at a time, watching the hardware.
# === REACTION-LANE TESTS - the animation-reaction crash-diagnosis lineage,=
# === most-evolved shape is test-reaction-in-the-beat(), below. ============
test-silent-angry-say() {
  local serial="${1:-0dd1d8bf}"
  local base="${GOPOD_WIREPOD_BASE_URL}"
  local payload='{{playAnimationWI||angry}}...'

  echo "TEST_SILENT_ANGRY_START serial=$serial payload=$payload"
  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/conn_test" "serial=$serial"
  sleep 1.5
  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/assume_behavior_control" "priority=high" "serial=$serial"
  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/say_text" "serial=$serial" "text=$payload"
  sleep 5.0
  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/release_behavior_control" "serial=$serial"
  echo "TEST_SILENT_ANGRY_DONE serial=$serial"
}

# test-concurrent-reaction - the gate test for the future interview reaction
# lane: proves (or disproves) that Wire-Pod's HTTP lane can genuinely run two
# robots at once - Brobot 1 speaking a count while Brobot 2 fires the same
# proven silent-angry animation mid-speech. Nothing beyond this test is built
# here; the reaction lane itself is a later, separate piece of work.
#
# HOLD ALIGNED, same day - Brobot 2's animation hold here was still the
# original 2.5s guess from before any hold value was actually tuned. Aligned
# to 5.0s, the operator's own proven-safe value. Brobot 1's own 7s count
# hold is a separate category (untimed speech, not an animation) and was
# left as-is.
#
# Shape: wake both robots first (the same conn_test -> settle pattern
# test-silent-angry-say already uses, applied per robot) - then Brobot 1's
# full assume -> say_text (a spoken count) -> release fires backgrounded so
# the shell continues immediately, held for ~7s so it's genuinely still
# "speaking" when Brobot 2 fires. ~2s later, Brobot 2 runs the exact same
# assume -> say_text -> release sequence test-silent-angry-say already
# proved safe (async playAnimationWI token only - the blocking token's
# crash is already on record, see BROBOT2_SILENT_ANGRY_SAY_001.md - never
# revived here). Both sequences go through the same single instrument every
# other note in this file uses (_gopod_note_send) - no direct SDK, no new
# transport.
#
# Layer 0 discipline: roles only. Brobot 1 / Brobot 2 in every line of code,
# every echo, this comment block - no persona names anywhere here.
#
# This is a TEST alias, not a production note, not wired into any song or
# the chord. Fire deliberately, watching the hardware - see
# CONCURRENT_REACTION_GATE_TEST_001.md for the live verification.
test-concurrent-reaction() {
  local s1="${GOPOD_BROBOT_1_SERIAL:-0dd1b9e9}"
  local s2="${GOPOD_BROBOT_2_SERIAL:-0dd1d8bf}"
  local base="${GOPOD_WIREPOD_BASE_URL}"
  local count_text="1, 2, 3, 4, 5, 6, 7, 8, 9"
  local anim_payload='{{playAnimationWI||angry}}...'
  local workdir
  workdir="$(mktemp -d)"

  echo "TEST_CONCURRENT_REACTION_START robot_1=$s1 robot_2=$s2 workdir=$workdir"

  # Wake both robots first - proven pattern, applied per robot.
  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/conn_test" "serial=$s1"
  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/conn_test" "serial=$s2"
  sleep 1.5

  # Brobot 1: assume -> say_text (a 9-count) -> hold ~7s (so it's still
  # "speaking" when Brobot 2 fires) -> release. Backgrounded so the shell
  # continues immediately.
  (
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/assume_behavior_control" "priority=high" "serial=$s1"
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/say_text" "serial=$s1" "text=$count_text"
    sleep 7
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/release_behavior_control" "serial=$s1"
  ) > "$workdir/brobot1.log" 2>&1 &
  local b1_pid=$!

  # ~2s after Brobot 1's say fires (assume+say_text return in well under a
  # second at the HTTP layer, so this sleep is the real timing control) -
  # Brobot 2's own wake already happened above; this is exactly
  # test-silent-angry-say's proven sequence minus its own redundant wake step.
  sleep 2
  (
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/assume_behavior_control" "priority=high" "serial=$s2"
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/say_text" "serial=$s2" "text=$anim_payload"
    sleep 5.0
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/release_behavior_control" "serial=$s2"
  ) > "$workdir/brobot2.log" 2>&1 &
  local b2_pid=$!

  wait "$b1_pid" "$b2_pid"

  echo "--- Brobot 1 (robot_1=$s1) HTTP log ---"
  cat "$workdir/brobot1.log"
  echo "--- Brobot 2 (robot_2=$s2) HTTP log ---"
  cat "$workdir/brobot2.log"

  local b1_total b1_ok b2_total b2_ok
  b1_total=$(grep -c "NOTE_HTTP" "$workdir/brobot1.log")
  b1_ok=$(grep -c "NOTE_HTTP status=200 body='success'" "$workdir/brobot1.log")
  b2_total=$(grep -c "NOTE_HTTP" "$workdir/brobot2.log")
  b2_ok=$(grep -c "NOTE_HTTP status=200 body='success'" "$workdir/brobot2.log")

  if [ "$b1_total" -gt 0 ] && [ "$b1_ok" -eq "$b1_total" ]; then
    echo "TEST_CONCURRENT_REACTION_RESULT robot=Brobot_1 PASS calls_ok=$b1_ok/$b1_total"
  else
    echo "TEST_CONCURRENT_REACTION_RESULT robot=Brobot_1 FAIL calls_ok=$b1_ok/$b1_total"
  fi
  if [ "$b2_total" -gt 0 ] && [ "$b2_ok" -eq "$b2_total" ]; then
    echo "TEST_CONCURRENT_REACTION_RESULT robot=Brobot_2 PASS calls_ok=$b2_ok/$b2_total"
  else
    echo "TEST_CONCURRENT_REACTION_RESULT robot=Brobot_2 FAIL calls_ok=$b2_ok/$b2_total"
  fi

  echo "TEST_CONCURRENT_REACTION_DONE workdir=$workdir"
}

# test-reaction-in-the-beat - the sequenced (never overlapping) counterpart
# to test-concurrent-reaction: one robot holds assumed control at a time,
# always - release plus a settle gap fully completes before the next robot
# ever assumes. Tests two things at once: the "reaction in the beat" show
# pattern itself, and the leading crash hypothesis from
# CONCURRENCY_CRASH_DIAGNOSIS_001.md - that releasing behavior control before
# the animation has actually finished playing, not concurrency itself, may be
# the real cause.
#
# HOLD is the operator's own proven empirical value - a single calculated
# measurement was tried and explicitly discarded as unreliable (operator's
# order: do not use or reference it going forward). The number in use is the
# operator's own repeated, direct result: clean twice at 5.0s and twice at
# 4.5s, hand-tuned live via test-angry-hold. HOLD = 5.0s. See
# REACTION_SEQUENCE_CHECK_001.md for the retune and the sequenced
# run it backs, and REACTION_IN_THE_BEAT_TEST_001.md for the retired
# measurement's own history (kept for the record, not for reuse). The two
# spoken counts either side ("1,2,3,4,5"
# / "6,7,8,9") use a generous fixed 6s hold each, not separately measured -
# per the task's own "measure once, or use a generous fixed hold" allowance.
#
# CLIP SWAP, same day, isolation test: every variable tried across this whole
# series so far (blocking vs async token, wake step present or not, hold
# duration, sequential vs concurrent) has had both a clean run and a crash on
# record - none isolated on its own, including the 5.0s-hold sequential run
# above, which still crashed at the animation step. The one variable never
# varied across the whole series was the animation clip itself
# (anim_rtpickup_loop_10, "angry"). Swapped here to
# {{playAnimationWI||frustrated}} -> anim_feedback_shutup_01 (per
# animation_vocab.json), same 5.0s hold, everything else byte-identical. See
# CLIP_SWAP_ISOLATION_TEST_001.md for the result - if clean, the crash
# isolates to the angry clip specifically, not this lane in general.
#
# ROOT CAUSE FOUND, same day, for the "reports clean but plays nothing" gap
# specifically (separate from the earlier crash series): Wire-Pod's own
# AnimationQueues registry (kgsim_cmds.go) is a global, in-memory, per-ESN
# "currently playing" flag with no HTTP-exposed reset. If an earlier
# animation dispatch on a robot never reached its own StopAnim_Queue call
# (e.g. an interrupted/Ctrl-C'd prior run), every later WI-token animation
# on that same robot silently blocks forever inside StartAnim_Queue, never
# reaching the real PlayAnimation call - confirmed live, not inferred:
# Wire-Pod's debug log showed "(waiting for animation to be done...)" at the
# exact same second as a real dispatch that produced no visible motion. Step
# 2 below now checks the debug log for this exact signature right after
# dispatch and warns plainly if found - no client-side fix is possible
# (AnimationQueues can't be reset except by restarting the wire-pod
# service). See ANIMATION_DISPATCH_ISOLATION_001.md for the full trace.
#
# ANGRY CLIP RESTORED, same day: the frustrated swap ran clean (operator
# confirmed seeing the animation; no stuck-signature warning fired) - token
# swapped back to angry (anim_rtpickup_loop_10) to test it against the same
# stuck-check now in place, with Brobot 2's AnimationQueue entry presumably
# clean after that last successful frustrated run. See
# ANIMATION_TOKEN_RETEST_001.md for the result once run. Operator-fired, not
# run by Claude, per this clip's own crash history.
#
# Strictly sequential throughout - no backgrounding, no &, one robot holding
# assumed control at a time, always. Same single instrument every other note
# in this file uses (_gopod_note_send). Async token only. Roles only
# (Brobot 1/Brobot 2) - no persona names anywhere in code, output, or this
# comment block.
#
# This is a TEST alias, disposable, not wired into any song or the chord.
#
# TOKEN PARAMETERIZED, same day: optional first argument picks the Step 2
# animation token (default "frustrated", the current safe default) - built
# so test-reaction-pick-animation (below) can chain a menu-picked token
# straight into this same proven sequence, rather than hand-editing this
# function per token. Calling this with no argument is byte-identical to
# before.
#
# SHORTENED + SPOKEN LINES, same day, per operator request ("No counting.
# Too long"): the two counts are gone. Brobot 1 now says "Animation test
# run" / "Run Complete" (short, no reason to hold as long as a spoken
# count needed - hold_phrase dropped from 6s to 3s to match). Brobot 2 now
# speaks a short line naming the emotion (e.g. "I'm sad" for the sad token)
# before reacting. Deliberately fired as TWO SEPARATE say_text calls -
# speak the line, then a bare animation-only dispatch - not one call with
# the token embedded in the sentence. That embedded-in-one-call shape is
# exactly what caused two real say_text HTTP timeouts on angry earlier
# today (see ANGRY_CLIP_RETEST_AND_ANIMATION_PICKER_001.md) - same visible
# effect (robot speaks, then reacts), the proven-safer mechanism
# underneath.
test-reaction-in-the-beat() {
  local s1="${GOPOD_BROBOT_1_SERIAL:-0dd1b9e9}"
  local s2="${GOPOD_BROBOT_2_SERIAL:-0dd1d8bf}"
  local base="${GOPOD_WIREPOD_BASE_URL}"
  local anim_token="${1:-frustrated}"
  local anim_payload="{{playAnimationWI||${anim_token}}}..."
  local hold_anim=5.0
  local hold_phrase=3.0
  local workdir
  workdir="$(mktemp -d)"
  local logfile="$workdir/reaction_in_the_beat.log"

  local anim_phrase
  case "$anim_token" in
    happy) anim_phrase="I'm happy" ;;
    veryHappy) anim_phrase="I'm very happy" ;;
    sad) anim_phrase="I'm sad" ;;
    verySad) anim_phrase="I'm very sad" ;;
    angry) anim_phrase="I'm angry" ;;
    frustrated) anim_phrase="I'm frustrated" ;;
    confused) anim_phrase="I'm confused" ;;
    thinking) anim_phrase="I'm thinking" ;;
    celebrate) anim_phrase="I'm celebrating" ;;
    love) anim_phrase="I'm feeling the love" ;;
    *) anim_phrase="I'm $anim_token" ;;
  esac

  {
    echo "[$(date +%T.%3N)] TEST_REACTION_IN_THE_BEAT_START robot_1=$s1 robot_2=$s2 anim_token=$anim_token hold_anim=${hold_anim}s hold_phrase=${hold_phrase}s"

    # Wake both robots first, per robot.
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/conn_test" "serial=$s1"
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/conn_test" "serial=$s2"
    sleep 1.5

    # STEP 1: Brobot 1 speaks "Animation test run" - assume -> say -> hold ->
    # release -> settle. Fully sequential; nothing else fires until this
    # completes.
    echo "[$(date +%T.%3N)] STEP_1_BROBOT_1_INTRO_START"
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/assume_behavior_control" "priority=high" "serial=$s1"
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/say_text" "serial=$s1" "text=Animation test run"
    sleep "$hold_phrase"
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/release_behavior_control" "serial=$s1"
    # Release is async, up to ~500ms lag per CONCURRENCY_CRASH_DIAGNOSIS_001.md -
    # 1s settle here is deliberate margin past that, before the next robot
    # ever assumes.
    sleep 1
    echo "[$(date +%T.%3N)] STEP_1_BROBOT_1_INTRO_DONE"

    # STEP 2: Brobot 2 speaks the emotion line, THEN separately fires the
    # bare animation - two calls, not one embedded sentence (see comment
    # above this function for why).
    echo "[$(date +%T.%3N)] STEP_2_BROBOT_2_ANIMATION_START"
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/assume_behavior_control" "priority=high" "serial=$s2"
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/say_text" "serial=$s2" "text=$anim_phrase"
    sleep 2
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/say_text" "serial=$s2" "text=$anim_payload"
    sleep 0.5
    # STUCK-ANIMATION CHECK, evidence-backed 2026-07-16 (see
    # ANIMATION_DISPATCH_ISOLATION_001.md): Wire-Pod's own
    # "(waiting for animation to be done...)" debug-log line, appearing at/
    # right after this dispatch, means StartAnim_Queue found this robot's
    # AnimationQueue entry already marked "currently playing" from an
    # earlier, never-completed dispatch (e.g. an interrupted prior run) and
    # is blocking on a channel that will never fire - robot.Conn.PlayAnimation()
    # itself is never reached, so nothing visibly plays. HTTP "success" above
    # proves nothing either way (confirmed throughout this whole crash
    # series); this log line is the only client-visible signal that actually
    # distinguishes a real dispatch from a stuck one. No client-side fix
    # exists for this - AnimationQueues is in-memory Go state with no
    # HTTP-exposed reset; only a wire-pod service restart clears it.
    if curl -s --max-time 3 "${base}/api/get_debug_logs" 2>/dev/null | tail -5 | grep -q "waiting for animation to be done"; then
      echo "[$(date +%T.%3N)] STEP_2_WARNING animation likely STUCK on a stale AnimationQueue entry - Wire-Pod's log shows the blocking-wait signature. PlayAnimation was probably never reached; HTTP success above does not mean the animation played. A wire-pod restart is needed to clear this - no code fix is possible from this alias."
    fi
    sleep "$hold_anim"
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/release_behavior_control" "serial=$s2"
    sleep 1
    echo "[$(date +%T.%3N)] STEP_2_BROBOT_2_ANIMATION_DONE"

    # STEP 3: Brobot 1 says "Run Complete" - same shape as step 1.
    echo "[$(date +%T.%3N)] STEP_3_BROBOT_1_OUTRO_START"
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/assume_behavior_control" "priority=high" "serial=$s1"
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/say_text" "serial=$s1" "text=Run Complete"
    sleep "$hold_phrase"
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/release_behavior_control" "serial=$s1"
    sleep 1
    echo "[$(date +%T.%3N)] STEP_3_BROBOT_1_OUTRO_DONE"
  } | tee "$logfile"

  local total ok
  total=$(grep -c "NOTE_HTTP" "$logfile")
  ok=$(grep -c "NOTE_HTTP status=200 body='success'" "$logfile")
  if [ "$total" -gt 0 ] && [ "$ok" -eq "$total" ]; then
    echo "TEST_REACTION_IN_THE_BEAT_RESULT PASS calls_ok=$ok/$total"
  else
    echo "TEST_REACTION_IN_THE_BEAT_RESULT FAIL calls_ok=$ok/$total"
  fi
  echo "TEST_REACTION_IN_THE_BEAT_DONE workdir=$workdir"
}

# test-reaction-pick-animation - reads animation_vocab.json's own verified
# tokens (single source of truth, not hand-copied - same file every
# brobots-anim-* alias and this whole test series already reads from),
# shows a numbered menu, and chains the picked token straight into
# test-reaction-in-the-beat's own proven sequence (wake -> Brobot 1 count ->
# Brobot 2 animation -> Brobot 1 count). Only verified:true tokens are
# offered - GOPOD_ANIM_TODO (verified:false) is deliberately excluded, same
# rule _brobots_play_anim's own aliases already follow. As of 2026-07-16,
# animation_vocab.json has 11 total entries, 10 verified.
#
# angry is flagged plainly in the menu, not hidden or blocked: three real
# crashes on record today (BROBOT2_SILENT_ANGRY_SAY_001.md,
# CONCURRENT_REACTION_GATE_TEST_001.md, and this session's own embedded-line
# attempts) - frustrated has never crashed once. The operator can still pick
# it; the menu just says so first.
#
# Uses plain `read` for the interactive prompt (not a python heredoc, which
# would consume stdin the way gopod-pick-model's own comment already
# documents) - this runs as a normal shell function, so `read` sees the
# real terminal.
test-reaction-pick-animation() {
  local vocab_path="/home/goverlord/wire-pod/chipper/animation_vocab.json"
  if [ ! -s "$vocab_path" ]; then
    echo "TEST_REACTION_PICK_ANIMATION_BLOCKED vocab_missing path=$vocab_path"
    return 1
  fi

  local tokens
  tokens="$(python3 -c "
import json
d = json.load(open('$vocab_path'))
for t in d.get('tokens', []):
    if t.get('verified'):
        print(t['name'])
")"
  if [ -z "$tokens" ]; then
    echo "TEST_REACTION_PICK_ANIMATION_BLOCKED no_verified_tokens_found"
    return 1
  fi

  echo "Verified animation tokens (animation_vocab.json):"
  local i=1
  local -a token_array
  while IFS= read -r t; do
    token_array[$i]="$t"
    if [ "$t" = "angry" ]; then
      echo "  $i. $t   (CRASHED 3x today - see ANIMATION_DISPATCH_ISOLATION_001.md and gopod_notes/ for the record)"
    else
      echo "  $i. $t"
    fi
    i=$((i + 1))
  done <<< "$tokens"
  local count=$((i - 1))

  local choice
  read -r -p "Pick a number [1-$count]: " choice

  if ! [[ "$choice" =~ ^[0-9]+$ ]] || [ "$choice" -lt 1 ] || [ "$choice" -gt "$count" ]; then
    echo "TEST_REACTION_PICK_ANIMATION_BLOCKED invalid_choice choice=$choice"
    return 1
  fi

  local chosen="${token_array[$choice]}"
  echo "TEST_REACTION_PICK_ANIMATION_CHOSEN token=$chosen"
  test-reaction-in-the-beat "$chosen"
}

# test-angry-hold - quick manual tuner for the angry animation's hold
# duration, separate from the fixed 5.0s baked into test-reaction-in-the-beat.
# Same wake step, same single instrument (_gopod_note_send), same async token
# only. Pass the hold in seconds as the first argument, with or without a
# leading "--" (e.g. `test-angry-hold --3.5` or `test-angry-hold 3.5`);
# defaults to the operator's own proven 5.0s if nothing is given (a single
# calculated-measurement default was tried and discarded as unreliable - see
# test-reaction-in-the-beat's own comment block). Optional second argument
# overrides the target serial (defaults to Brobot 2, 0dd1d8bf). Disposable,
# not wired into any song or the chord - built for hand-tuning the hold value
# live, one fire at a time.
test-angry-hold() {
  local raw="${1:-5.0}"
  local hold="${raw#--}"
  local serial="${2:-0dd1d8bf}"
  local base="${GOPOD_WIREPOD_BASE_URL}"
  local payload='{{playAnimationWI||angry}}...'

  echo "TEST_ANGRY_HOLD_START serial=$serial hold=${hold}s payload=$payload"
  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/conn_test" "serial=$serial"
  sleep 1.5
  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/assume_behavior_control" "priority=high" "serial=$serial"
  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/say_text" "serial=$serial" "text=$payload"
  sleep "$hold"
  GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/release_behavior_control" "serial=$serial"
  echo "TEST_ANGRY_HOLD_DONE serial=$serial hold=${hold}s"
}

# === test-anim-searching / test-anim-answering / test-anim-kg-success ===
# Golden notes, KG_ANIMATION_GOLDEN_NOTES_001.md (gopod_notes/): one animation
# fireable in isolation, one alias each, so the operator can watch one thing
# at a time and judge it with his own eyes. animation_vocab.json confirms all
# three tokens present, verified:true - answering -> anim_knowledgegraph_answer_01,
# searching -> anim_knowledgegraph_searching_01, kgSuccess -> anim_knowledgegraph_success_01.
#
# Template note: test-arm-cue/test-head-nod were named as the shape to copy,
# but their own dispatch mechanism (run_single_note() in
# run_robot_control_song_001.py) only knows arm_test/head_nod/fireworks/weather -
# it has no code path for a playAnimationWI/animation_vocab.json token at all,
# and neither alias fires a conn_test wake step itself (their "assume" lives
# inside that python call, no pre-flight wake). The proven wake-step mechanism
# for firing a bare animation token (the actual real failure mode here) is
# test-silent-angry-say/test-angry-hold's own shape instead: conn_test -> 1.5s
# settle -> assume -> dispatch -> hold -> release (see brobots.sh's own
# comment above test-silent-angry-say, "HTTP success but no playback on cold
# first press"). So these three notes mirror test-arm-cue/test-head-nod at the
# ALIAS shape level (isolated single dispatch, robot argument, hold-override
# argument, PASS/BLOCKED reporting, no chaining) while reusing the
# already-proven wake step, and _brobots_anim_is_loop_token (unchanged,
# searching/answering=loop, kgSuccess=one-shot) for the loop-vs-one-shot call.
#
# Dispatch-count precision note, live-tested 2026-07-28: the loop count is
# ported directly from run_songs_runner_001.py's own run_animation_only()
# accumulator ("elapsed = 0.0; while elapsed < hold_seconds: dispatch;
# sleep(0.333); elapsed += 0.333"), NOT from this file's existing
# _brobots_play_anim_single (its own repeats=int(hold/0.333) formula is off by
# one dispatch vs. bingo's real one - confirmed live: hold=1.0s prints
# "repeats=3" there, but bingo's own accumulator produces 4 dispatches for
# searching at that same hold, which is the number the operator's own live run
# actually showed). _brobots_play_anim_single itself is untouched - this is a
# separate, narrower dispatch loop, ported straight from the song, not
# delegated to the pre-existing helper.
#
# bingo's own knobs.json uses hold=1.0s for searching (4 dispatches, matches
# this family's 1s default exactly) and hold=2.5s for answering (8 dispatches -
# this family's 1s default only gives 4; pass 2.5 as the hold argument to
# reproduce bingo's own answering count exactly). kgSuccess is a one-shot
# token - always 1 dispatch, any hold.
_test_anim_isolated() {
  local token="$1" robot="${2:-1}" hold="${3:-1}"
  local name_upper
  name_upper="$(printf '%s' "$token" | tr '[:lower:]' '[:upper:]')"
  local base="${GOPOD_WIREPOD_BASE_URL}"
  local s1="${GOPOD_BROBOT_1_SERIAL:-0dd1b9e9}"
  local s2="${GOPOD_BROBOT_2_SERIAL:-0dd1d8bf}"
  local serial

  case "$robot" in
    1) serial="$s1" ;;
    2) serial="$s2" ;;
    *) echo "TEST_ANIM_${name_upper}_BLOCKED bad_robot robot=$robot (use 1 or 2)"; return 1 ;;
  esac

  if ! [[ "$hold" =~ ^[0-9]+([.][0-9]+)?$ ]] || [ "$(awk -v h="$hold" 'BEGIN{print (h<=0)}')" = "1" ]; then
    echo "TEST_ANIM_${name_upper}_BLOCKED bad_hold hold=$hold (must be a positive number of seconds)"
    return 1
  fi

  local workdir logfile
  workdir="$(mktemp -d)"
  logfile="$workdir/test_anim_${token}.log"
  local anim_payload="{{playAnimationWI||${token}}}"

  {
    echo "[$(date +%T.%3N)] TEST_ANIM_${name_upper}_START robot=$robot serial=$serial token=$token hold=${hold}s"

    # Wake step - same conn_test -> 1.5s settle pattern test-silent-angry-say/
    # test-angry-hold already prove fixes the live-observed "HTTP success but
    # no playback on cold first press" defect for animation dispatch.
    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/conn_test" "serial=$serial"
    sleep 1.5

    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/assume_behavior_control" "priority=high" "serial=$serial"

    if _brobots_anim_is_loop_token "$token"; then
      # Ported verbatim from run_songs_runner_001.py's own
      # run_animation_only() accumulator - see comment block above this
      # function for why this isn't _brobots_play_anim_single's own
      # int(hold/0.333) formula.
      local elapsed=0.0
      echo "[$(date +%T.%3N)] TEST_ANIM_${name_upper}_MODE mode=loop"
      while awk -v e="$elapsed" -v h="$hold" 'BEGIN{exit !(e<h)}'; do
        GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/say_text" "serial=$serial" "text=$anim_payload"
        sleep 0.333
        elapsed=$(awk -v e="$elapsed" 'BEGIN{printf "%.6f", e+0.333}')
      done
    else
      echo "[$(date +%T.%3N)] TEST_ANIM_${name_upper}_MODE mode=once"
      GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/say_text" "serial=$serial" "text=$anim_payload"
      sleep "$hold"
    fi

    GOPOD_WIREPOD_BASE_URL="$base" _gopod_note_send "/api-sdk/release_behavior_control" "serial=$serial"
  } | tee "$logfile"

  local total ok
  total=$(grep -c "NOTE_HTTP" "$logfile")
  ok=$(grep -c "NOTE_HTTP status=200 body='success'" "$logfile")
  if [ "$total" -gt 0 ] && [ "$ok" -eq "$total" ]; then
    echo "TEST_ANIM_${name_upper}_RESULT PASS calls_ok=$ok/$total"
  else
    echo "TEST_ANIM_${name_upper}_RESULT FAIL calls_ok=$ok/$total"
  fi
  echo "TEST_ANIM_${name_upper}_DONE robot=$robot serial=$serial token=$token hold=${hold}s"
}

# One animation, one alias, 1-second default hold - `test-anim-searching [robot] [hold]`.
# searching is a loop token in kgsim.go (re-fired every ~0.333s for the hold
# duration) - default hold=1s reproduces bingo's own round_3_searching step
# (hold=1.0s, 4 dispatches) exactly.
test-anim-searching() { _test_anim_isolated "searching" "${1:-1}" "${2:-1}"; }

# answering is also a loop token - default hold=1s here (4 dispatches) does
# NOT reproduce bingo's own hold=2.5s/8-dispatch answering steps; pass 2.5 as
# the second argument (test-anim-answering 1 2.5) to match that exactly.
test-anim-answering() { _test_anim_isolated "answering" "${1:-1}" "${2:-1}"; }

# kgSuccess is a one-shot token (fires once, holds, releases) - dispatch
# count is always 1 regardless of hold, matching every kgSuccess step in
# bingo's own knobs.json (all hold=2.5s there; default hold here is 1s, only
# the hold duration differs, never the dispatch count for a one-shot token).
test-anim-kg-success() { _test_anim_isolated "kgSuccess" "${1:-1}" "${2:-1}"; }

# Generic counterpart to the three fixed-token aliases above - same
# _test_anim_isolated mechanism (conn_test wake step, dispatch-count-precise
# loop/one-shot via _brobots_anim_is_loop_token, PASS/BLOCKED HTTP
# verdict), but the token is a real argument instead of baked into the
# alias name - mirrors brobots-anim-test's own already-generic shape, just
# routed through this family's newer/more precise dispatch instead of
# _brobots_play_anim_single. Added 2026-08-12 so any animation_vocab.json
# token (e.g. dartingEyes) gets this family's proven dispatch precision
# without needing its own named alias first.
test-anim-token() {
  if [ -z "$1" ]; then
    echo "TEST_ANIM_TOKEN_USAGE: test-anim-token <token> [robot: 1|2, default 1] [hold seconds, default 1]"
    return 1
  fi
  _test_anim_isolated "$1" "${2:-1}" "${3:-1}"
}

# === rehearse-searching-1s / rehearse-searching-2s ===
# Two fixed-duration rehearsal siblings of test-anim-searching above - same
# "searching" token confirmed verbatim from bingo's own knobs.json
# (round_3_searching, animation_token="searching"), same
# _test_anim_isolated() call, same conn_test->assume->loop->release shape,
# same 1/2 robot argument (searching's own dispatch mechanism - Wire-Pod's
# HTTP say_text/playAnimationWI, not the direct-SDK sleep binary - has no
# "both" mode anywhere in this file; not invented here). Only the hold is
# fixed: exactly 1s / 2s, so the operator can rehearse one clean closed loop
# of a known length without having to remember or retype a hold argument.
# Isolated rehearsal tools only - no song, score, knobs.json, or story.md
# touched by either.
rehearse-searching-1s() { _test_anim_isolated "searching" "${1:-1}" 1; }
rehearse-searching-2s() { _test_anim_isolated "searching" "${1:-1}" 2; }

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

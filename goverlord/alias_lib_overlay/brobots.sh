# brobots.sh - GOPOD's terminal aliases. Each one's card: alias_play_studio/ in the GOPOD repo.

GOPOD_LOCAL_CONFIG_PATH="${GOPOD_LOCAL_CONFIG_PATH:-$HOME/.gopod_alias_lib/local_config.sh}"
[ -f "$GOPOD_LOCAL_CONFIG_PATH" ] && . "$GOPOD_LOCAL_CONFIG_PATH"

_GOPOD_NOTE_RUNNER_PATH="/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools/run_section1_full_live_001.py"

_BROBOTS_WHEEL_WAKE_RELEASE_SETTLE_SECONDS=2.5
_BROBOTS_WHEEL_WAKE_POST_REASSUME_SETTLE_SECONDS=0.5

GOPOD_EXPECTED_AUDIO_SINK=""
GOPOD_EXPECTED_AUDIO_SOURCE=""
_GOPOD_AUDIO_CONFIG_PATH="$(dirname "${BASH_SOURCE[0]}")/gopod_audio_config.sh"
if [ -f "$_GOPOD_AUDIO_CONFIG_PATH" ]; then
  source "$_GOPOD_AUDIO_CONFIG_PATH"
fi

_gopod_set_audio_routing() {
  command -v pactl >/dev/null 2>&1 || return 0
  if [ -z "$GOPOD_EXPECTED_AUDIO_SINK" ] || [ -z "$GOPOD_EXPECTED_AUDIO_SOURCE" ]; then
    echo "AUDIO_ROUTING_SET_SKIPPED no gopod_audio_config.sh (see gopod_audio_config.example.sh)" >&2
    return 0
  fi
  if pactl set-default-sink "$GOPOD_EXPECTED_AUDIO_SINK" 2>/dev/null \
    && pactl set-default-source "$GOPOD_EXPECTED_AUDIO_SOURCE" 2>/dev/null; then
    echo "AUDIO_ROUTING_SET speakers=Built-in Audio Analog Stereo mic=USB Audio Mono"
  else
    echo "AUDIO_ROUTING_SET_FAILED check 'pactl list short sinks' / 'pactl list short sources'" >&2
  fi
}

_live_robots_read_key() {
  local k rest
  IFS= read -rsn1 k
  if [ -z "$k" ]; then
    echo enter
    return
  fi
  if [ "$k" = $'\x1b' ]; then
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

# live_robots_prompt - asks: live robots? (y / n / Esc)
live_robots_prompt() {
  local live_choice highlight old_stty key
  if [ ! -t 0 ]; then
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

  if [ "$live_choice" = "abort" ]; then
    echo 2
  elif [ "$live_choice" = "n" ]; then
    echo 0
  else
    echo 1
  fi
}

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

# pha0b - play a song or a range of it, live or dry: alone opens the menu
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

  case "$song" in
    bingo)
      runner="run_golden_song_001.py"
      env_prefix="GOPOD_GOLDEN_PLAYHEAD"
      song_dir_export="$(_song_dir_by_id brobots_bingo)"
      ;;
    nap)
      runner="run_golden_song_001.py"
      env_prefix="GOPOD_GOLDEN_PLAYHEAD"
      song_dir_export="$(_song_dir_by_id brobots_baby_robots_sleep)"
      ;;
    *)
      echo "PHA0B_REFUSED song=$song unknown - must be one of: bingo nap"
      return 1
      ;;
  esac

  local live_gate
  if [ "$song" = "bait" ] && [ "$PHA0B_BAIT_NAV_ACTIVE" = "1" ]; then
    live_gate="$PHA0B_BAIT_LIVE_GATE"
  else
    live_gate="$(live_robots_prompt)"
  fi
  if [ "$live_gate" = "2" ]; then
    echo "PHA0B_ABORTED (ESC at live-robots prompt) - exiting, pha0b did not start"
    return 1
  fi

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

  local phcal_song_dir=""
  case "$song" in
    bingo) phcal_song_dir="$(_song_dir_by_id brobots_bingo)" ;;
  esac
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
    if s.get('note') in ('arm_cue', 'nod', 'head_nod', 'animation', 'wheel_nudge', 'brobots_ready_together'):
        print(s['step_id'])
")"
      if [ -n "$qualifying_ids" ]; then
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

  local robot_filter_choice="both"
  if [ "$runner" = "run_golden_song_001.py" ]; then
    if [ "$song" = "bait" ] && [ "$PHA0B_BAIT_NAV_ACTIVE" = "1" ]; then
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
      echo "PHA0B_BAIT_MULTI_PASS_SETTLE sleeping 1.5s before brobot 2's pass"
      sleep 1.5
    fi

    local stderr_capture
    stderr_capture="$(mktemp)"
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

# pha0b_menu - the song menu pha0b opens
pha0b_menu() {
  local songs_dir="/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs"

  while true; do
  local -a dir_array
  local i=1
  echo "PHA0B_MENU songs on disk:"
  for d in "$songs_dir"/*/; do
    local name
    name="$(basename "$d")"
    [ "$name" = "zzz_archives" ] && continue
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

  local phcal_slot=$((dir_count + 1))
  echo "  $phcal_slot. PLAYHEAD Calibrations (PHCAL.md)"

  local notation_slot=$((dir_count + 2))
  echo "  $notation_slot. Song Notation / Score"

  echo "Please note:"
  echo "  see songs/tools/README.md for the song-tools - run via their own aliases, not from here"

  echo "** anything entered outside selection range will exit **"
  local dir_choice
  dir_choice="$(_pha0b_prompt_range_choice "Pick a song [1-$notation_slot, 0 to exit]: " "$notation_slot" "PHA0B_MENU_BLOCKED")" || return 1

  if [ "$dir_choice" -eq "$phcal_slot" ]; then
    ( cd "$(dirname "${BASH_SOURCE[0]}")" && python3 phcal_isolate_001.py )
    continue
  fi

  if [ "$dir_choice" -eq "$notation_slot" ]; then
    _pha0b_notation_flow
    continue
  fi

  local chosen_dir="${dir_array[$dir_choice]}"

  local song=""
  case "$(_song_id_of_dir "$songs_dir/$chosen_dir")" in
    brobots_bingo) song="bingo" ;;
    brobots_baby_robots_sleep) song="nap" ;;
  esac
  case "$chosen_dir" in
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
      if [ -z "$song" ]; then
        echo "PHA0B_MENU_REFUSED song_dir=$chosen_dir no playhead wired for this song yet under pha0b's own song keyword list (bingo|nap) - see pha0b()'s case statement above"
        return 1
      fi
      ;;
  esac

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

  local tools_dir="/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools"
  local song_dir="$songs_dir/$chosen_dir"
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

_song_dir_by_id() {
  python3 /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools/song_folders_001.py "$1"
}
_song_id_of_dir() {
  python3 -c 'import json,sys; print(json.load(open(sys.argv[1]+"/knobs.json")).get("song_id",""))' "$1" 2>/dev/null
}

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

# score - a song's score on the screen: [song] [narrative]
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

# score-save - the score page saved in the song's notation/ folder: [song] [narrative]
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

# phcal - tune one move on one robot
phcal() {
  local primitive="$1"

  if [ -z "$primitive" ]; then
    ( cd "$(dirname "${BASH_SOURCE[0]}")" && python3 phcal_isolate_001.py )
    return $?
  fi

  local -a live_env=()
  local live_gate
  live_gate="$(live_robots_prompt)"
  if [ "$live_gate" = "2" ]; then
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

# tempo-set - a song's tempo, walked through
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

_ROBOT_SLEEP_DIRECT_BIN="/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_probes/tools/direct_sdk_robot_sleep_001"

_ROBOT_INFO_DIRECT_BIN="/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_probes/tools/direct_sdk_robot_info_001"

_CUBE_BLIP_DIRECT_BIN="/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_probes/tools/direct_sdk_cube_blip_001"
_CUBE_BLIP_DIRECT_SRC="/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_probes/tools/direct_sdk_cube_blip_001.go"

_GOPOD_SLEEP_FIRST_WAKE_SETTLE_SECONDS="${GOPOD_SLEEP_FIRST_WAKE_SETTLE_SECONDS:-3}"

_gopod_robot_serial() {
    python3 - "$1" <<'PYEOF'
import json, os, sys
label = sys.argv[1].upper()
key = {"ESN01": "brobot_1", "ESN02": "brobot_2"}.get(label)
out = sys.argv[1]
if key:
    for path, k in (("/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_brobots_now.json", key), (os.path.expanduser("~/.gopod_esns.json"), label)):
        try:
            v = json.load(open(path)).get(k)
            if v and v != "no-brobot-2":
                out = v
                break
        except (OSError, ValueError):
            pass
print(out)
PYEOF
}

# gobingo - the live Chocolate Bingo game, Brobot 2's angry reactor alongside
gobingo() {
    (
        cd /home/goverlord/wire-pod || exit 1

        /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/01b_brobots_bingo_game/bingo_reactor/.venv/bin/python3 \
            /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/01b_brobots_bingo_game/bingo_reactor/bingo_reactor_001.py &
        local reactor_pid=$!
        trap 'kill "$reactor_pid" 2>/dev/null' EXIT INT TERM

        WIREPOD_EX_TMP_PATH=/home/goverlord/crushn8r_git/SDK/sources/vectorx/vectorfs/tmp \
        WIREPOD_EX_DATA_PATH=/home/goverlord/crushn8r_git/SDK/sources/vectorx/vectorfs/data \
        WIREPOD_EX_NVM_PATH=/home/goverlord/crushn8r_git/SDK/sources/vectorx/vectorfs/nvm \
        /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/01b_brobots_bingo_game/bin/vectorx-gobingo \
            --serial "$(_gopod_robot_serial ESN01)" \
            --locale en-US \
            --speechText bingo \
            "$@"
    )
}

# gohockey - VR/AR air hockey: [2] Brobot 2 plays
gohockey() {
  local who="Brobot 1"
  [ "$1" = "2" ] && who="Brobot 2"
  ~/miniforge3/envs/gopod-coral39/bin/python ~/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_probes/tools/gopod_vr_air_hockey.py --player "$who"
}

_gopod_robot_ip() {
    local serial; serial="$(_gopod_robot_serial "ESN0$1")"
    python3 - "$serial" <<'PYEOF'
import json, sys
try:
    for r in json.load(open("/home/goverlord/wire-pod/chipper/jdocs/botSdkInfo.json")).get("robots", []):
        if r.get("esn") == sys.argv[1]:
            print(r.get("ip_address", ""))
except (OSError, ValueError):
    pass
PYEOF
}
_gopod_robot_ssh() {
    ssh -q -o BatchMode=yes -o ConnectTimeout=8 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
        -o PubkeyAcceptedAlgorithms=+ssh-rsa -o HostKeyAlgorithms=+ssh-rsa \
        -i "$HOME/.ssh/vector_unlock/ssh_root_key" "root@$1" "$2"
}

# gofps - the face's frame rate: 30|60 [1|2] (after a reboot)
gofps() {
    local want="$1" which="${2:-1 2}" n ip
    for n in $which; do
        ip="$(_gopod_robot_ip "$n")"
        if [ -z "$ip" ]; then echo "Brobot $n: not in Wire-Pod's robot list"; continue; fi
        case "$want" in
            30) _gopod_robot_ssh "$ip" 'touch /data/data/rebuild/using-30-fps' && echo "Brobot $n: set to 30 fps - reboot Brobot $n now for it to take effect" ;;
            60) _gopod_robot_ssh "$ip" 'rm -f /data/data/rebuild/using-30-fps' && echo "Brobot $n: set to 60 fps - reboot Brobot $n now for it to take effect" ;;
            "") if _gopod_robot_ssh "$ip" 'test -e /data/data/rebuild/using-30-fps'; then echo "Brobot $n: set to 30 fps"; else echo "Brobot $n: set to 60 fps"; fi ;;
            *) echo "usage: gofps 30|60 [1|2]"; return 1 ;;
        esac
    done
}

# gopower - the performance profile: normal|balanced|high [1|2]
gopower() {
    local want="$1" which="${2:-1 2}" n ip f
    case "$want" in normal) f=0 ;; balanced) f=1 ;; high) f=2 ;; "") f="" ;; *) echo "usage: gopower normal|balanced|high [1|2]"; return 1 ;; esac
    for n in $which; do
        ip="$(_gopod_robot_ip "$n")"
        if [ -z "$ip" ]; then echo "Brobot $n: not in Wire-Pod's robot list"; continue; fi
        if [ -n "$f" ]; then curl -s -m 20 -o /dev/null "http://$ip:8080/api/mods/FreqChange/set?freq=$f"; fi
        case "$(curl -s -m 8 "http://$ip:8080/api/mods/FreqChange/get")" in
            0) echo "Brobot $n: normal (533 MHz)" ;; 1) echo "Brobot $n: balanced (729 MHz)" ;; 2) echo "Brobot $n: high (1.26 GHz)" ;; *) echo "Brobot $n: no answer" ;;
        esac
    done
    if [ -n "$f" ]; then echo "No reboot needed - ok to continue."; fi
}

# goinit - a session's start: gorepo, gopower balanced, gofps, wpr 2
goinit() {
    cd ~/crushn8r_git/GOPOD || return 1
    gopower balanced
    gofps
    wpr 2
}

_GOPOD_HOCKEY_PY=~/miniforge3/envs/gopod-coral39/bin/python
_GOPOD_HOCKEY=~/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_probes/tools/gopod_vr_air_hockey.py
_GOPOD_STAGE_DIR=~/crushn8r_git/gopod_notes/stage

# gosnap - one overhead picture of the stage
gosnap() {
  mkdir -p "$_GOPOD_STAGE_DIR"
  local out="$_GOPOD_STAGE_DIR/snap_$(date +%Y%m%d_%H%M%S).jpg"
  "$_GOPOD_HOCKEY_PY" -I - "$_GOPOD_HOCKEY" "$out" 2>/dev/null <<'PYEOF'
import importlib.util, sys, cv2
spec = importlib.util.spec_from_file_location("hockey", sys.argv[1])
m = importlib.util.module_from_spec(spec); spec.loader.exec_module(m)
cap = m.open_camera()
ok, img = False, None
for _ in range(15):  # the first frames come before the exposure settles
    ok, img = cap.read()
cap.release()
print("stage snapshot: " + sys.argv[2] if ok and cv2.imwrite(sys.argv[2], img) else "no picture - is the RunCam plugged in?")
PYEOF
}

# gohockey-cal - find the ruler and the track again, then stop
gohockey-cal() {
  "$_GOPOD_HOCKEY_PY" "$_GOPOD_HOCKEY" --recalibrate --lock-only
}

# gohockey-part - one part of the game alone: 1|2|3 [2]
gohockey-part() {
  local part="$1" who="Brobot 1"
  [ "$2" = "2" ] && who="Brobot 2"
  case "$part" in 1|2|3) ;; *) echo "usage: gohockey-part 1|2|3 [2]"; return 1 ;; esac
  "$_GOPOD_HOCKEY_PY" "$_GOPOD_HOCKEY" --player "$who" --part "$part"
}

# gohockey-rec - the whole game saved to a video: [2]
gohockey-rec() {
  local who="Brobot 1"
  [ "$1" = "2" ] && who="Brobot 2"
  mkdir -p "$_GOPOD_STAGE_DIR"
  local out="$_GOPOD_STAGE_DIR/hockey_$(date +%Y%m%d_%H%M%S).mp4"
  "$_GOPOD_HOCKEY_PY" "$_GOPOD_HOCKEY" --player "$who" --out "$out" && echo "saved: $out"
}

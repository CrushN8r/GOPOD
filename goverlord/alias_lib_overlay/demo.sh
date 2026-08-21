# === GOPOD one-command feature launcher (formerly WDTM) ===

export GOPOD_REPO="${GOPOD_REPO:-$HOME/crushn8r_git/GOPOD}"
export GOPOD_PAGE_DIR="$GOPOD_REPO/goverlord/runtime/gopod_layer/web_display/gopod_demo_8011"

# STAGE PINNACLE: cockpit URL/port sourced from
# goverlord/runtime/data_gomad/configs/endpoints.json, not hardcoded here.
# Path corrected 2026-07-30 - gomads/ was removed, data_gomad/ moved to
# runtime/ (see GOMADS_TREE_REMOVED_001.md).
# python3 -c is used (not jq) since python3 is already a required
# dependency of the writer/cockpit, no new dependency added.
export GOPOD_COCKPIT_URL="$(python3 -c '
import json
with open("'"$GOPOD_REPO"'/goverlord/runtime/data_gomad/configs/endpoints.json") as f:
    print(json.load(f)["cockpit"]["url"])
')"
export GOPOD_CONTROLS="${GOPOD_CONTROLS:-KP1 Doc | KP2 Pip | 000 exit}"
export GOPOD_PAGE_PORT="$(python3 -c '
import json
with open("'"$GOPOD_REPO"'/goverlord/runtime/data_gomad/configs/endpoints.json") as f:
    print(json.load(f)["cockpit"]["port"])
')"

_gopod_open_cockpit() {
  echo "GOPOD_COCKPIT_OPEN $GOPOD_COCKPIT_URL"
  if [ "${GOPOD_NO_BROWSER:-0}" = "1" ]; then
    return 0
  fi
  if command -v xdg-open >/dev/null 2>&1; then
    nohup xdg-open "$GOPOD_COCKPIT_URL" >/tmp/gopod_demo_browser_open.log 2>&1 &
  elif command -v sensible-browser >/dev/null 2>&1; then
    nohup sensible-browser "$GOPOD_COCKPIT_URL" >/tmp/gopod_demo_browser_open.log 2>&1 &
  fi
}

gopod-demo1-validation-samples() {
  (
    set -u
    cd "$GOPOD_REPO" || {
      echo "GOPOD_VALIDATION_REPO_BLOCKED path=$GOPOD_REPO"
      return 1
    }
    python3 goverlord/runtime/gopod_layer/web_display/gopod_demo_8011/gopod_demo_8011.py --write-validation-samples
  )
}

_gopod_cockpit_healthy() {
  [ "$(curl -sS -o /dev/null -w '%{http_code}' --max-time 2 "$GOPOD_COCKPIT_URL" 2>/dev/null)" = "200" ]
}

# is-that-you: single golden-truth launcher for the Is-That-You cross-persona
# awareness bit (NumLock-gated KP1=Doc/KP2=Pip live push-to-talk, wrong-robot
# fast-reply). Consolidates and replaces the former gopod-demo1 (cockpit +
# writer) and gopod_film (writer only, no cockpit) - both retired 2026-07-23,
# this is their single successor: cockpit auto-launch/reuse from gopod-demo1,
# the accurate KP1(Doc)/KP2(Pip) controls line from gopod_film. GOPOD_NO_BROWSER=1
# still skips the browser tab (cockpit server still comes up, same as before).
#
# Cast, printed on every launch: Brobot 1 (Doc, KP1), Brobot 2 (Pip, KP2),
# and Brobot 0 - the Operator, voiced through Mic_1 (the one physical mic in
# use today; naming leaves room for a Mic_2 later, not built).
is-that-you() {
  (
    set -u

    cd "$GOPOD_REPO" || {
      echo "GOPOD_REPO_BLOCKED path=$GOPOD_REPO"
      return 1
    }

    if [ ! -s "$GOPOD_PAGE_DIR/gopod_demo.html" ]; then
      echo "GOPOD_COCKPIT_BLOCKED path=$GOPOD_PAGE_DIR/gopod_demo.html"
      return 1
    fi

    if _gopod_cockpit_healthy; then
      echo "GOPOD_COCKPIT_ALREADY_RUNNING"
    else
      nohup python3 "$GOPOD_PAGE_DIR/gopod_demo_8011.py" >/tmp/gopod_demo_8011_cockpit.log 2>&1 &
      disown
      cockpit_up=0
      for _attempt in 1 2 3 4 5 6 7 8 9 10; do
        sleep 0.5
        if _gopod_cockpit_healthy; then
          cockpit_up=1
          break
        fi
      done
      if [ "$cockpit_up" != "1" ]; then
        echo "GOPOD_COCKPIT_LAUNCH_FAILED see /tmp/gopod_demo_8011_cockpit.log"
        return 1
      fi
      echo "GOPOD_COCKPIT_LAUNCHED"
    fi

    device_path="$(python3 "$GOPOD_PAGE_DIR/gopod_ptt_chat_writer_013.py" --resolve-device 2>/tmp/gopod_ptt_resolve_device.err)"
    if [ -z "$device_path" ]; then
      echo "GOPOD_WRITER_DEVICE_NOT_FOUND"
      cat /tmp/gopod_ptt_resolve_device.err
      return 1
    fi

    _gopod_open_cockpit

    echo "GOPOD_STACK_READY"
    echo "Cockpit: $GOPOD_COCKPIT_URL"
    echo "Writer:  $GOPOD_PAGE_DIR/gopod_ptt_chat_writer_013.py --device $device_path"
    echo "GOPOD_CAST: Brobot 1=Doc (KP1) | Brobot 2=Pip (KP2) | Brobot 0=Operator (Mic_1)"
    echo "Controls: NumLock on, hold KP1 (Doc) or KP2 (Pip) to speak, KP0 x3 to exit."

    python3 "$GOPOD_PAGE_DIR/gopod_ptt_chat_writer_013.py" --device "$device_path" "$@"
  )
}

# gopod-ptt-display: spectator view for a second terminal/pane while
# is-that-you runs the writer elsewhere. Tails the writer's own
# session.log (~/gopod_tts/sessions/<date>/<time>_session/session.log,
# per paths.json's session_root) and highlights just numlock, PTT
# press/exit, mic/audio, STT transcript, and LLM lines out of the raw
# console spam. The writer has no distinct STT-result token of its own -
# transcript text ships as a bare JSON envelope (source_producer:
# operator_mic_kp1/kp2), so that's matched by key name, not a GOPOD_PTT_
# prefix like the others.
gopod-ptt-display() {
  local latest
  latest="$(ls -dt "$HOME"/gopod_tts/sessions/*/*_session 2>/dev/null | head -1)"
  if [ -z "$latest" ]; then
    echo "GOPOD_PTT_DISPLAY_NO_SESSION_YET - start the writer first (is-that-you)."
    return 1
  fi
  echo "GOPOD_PTT_DISPLAY watching: $latest/session.log"
  tail -n 0 -F "$latest/session.log" | grep --line-buffered -E --color=always \
    'GOPOD_PTT_NUMLOCK|GOPOD_PTT_LISTENING|GOPOD_PTT_EXIT_TAP|GOPOD_PTT_IGNORED_REPEAT|GOPOD_PTT_STDIN_KP|GOPOD_PTT_AUDIO_|GOPOD_PTT_NATIVE_RATE|GOPOD_PTT_MASTER_WAV|GOPOD_PTT_VOSK_|GOPOD_PTT_LLM_|source_producer'
}

# gopod-numpad-map: prints the golden numpad/NumLock persona mapping table -
# one source-of-truth file, numpad_persona_map_001.json (same dir as this
# script), edited directly to remap a key; this alias only ever reads and
# prints it, never writes. NumLock ON/OFF doubles the physical KP0-9 keys
# into 20 addressable persona slots - see
# gopod_notes/NUMPAD_NUMLOCK_GOLDEN_MAPPING_001.md for the design behind it.
# --json prints the raw file instead of the formatted table.
gopod-numpad-map() {
  python3 "$HOME/.gopod_alias_lib/print_numpad_map_001.py" "$@"
}

# RETIRED 2026-07-06: target archived, see alias_keyboard_survey_001.md
# gopod-pre-demo () {
#   (
#     set -u
#     gorepo || return 1
#     python3 goverlord/runtime/data_gomad/wire-pod/vector_backpack_brobot_default_fallback_001/pre_demo_wirepod_table_reset_001.py
#   )
# }

# goshot: GOPOD snapshot. Dumps a reviewable tree + full code/text snapshot of
# the repo to snapshots/<timestamp>/ at the repo root (gitignored/excluded,
# never tracked). Moved here from repo root on 2026-07-06 — it's a dev
# utility, not part of the shipped GOPOD product. Must run from the repo
# root: the script's git-recency scoping and its "snapshots/" output path
# are both relative to cwd. Pass args straight through, e.g.
# ACTIVE_DAYS=7 goshot.
goshot () {
  (
    cd "$GOPOD_REPO" || return 1
    ./goverlord/tools/goshot.sh "$@"
  )
}

GOPOD_JSON_VIEWER="${GOPOD_JSON_VIEWER:-$HOME/tools/chatgpt-json-tree-viewer/chatgpt-json-tree-viewer.html}"

# gopod-json-view: opens a GOPOD song's knobs.json in the existing tree
# viewer (chatgpt-json-tree-viewer.html) from the terminal. The viewer only
# ever loads a file via its own file-input/drag-drop (confirmed read-only:
# no URL/query-param/auto-load hook exists in the viewer's code) - this
# alias does not and cannot inject the file into the page without editing
# the viewer, which is out of scope. It resolves and prints the file's real
# path, then launches the viewer, so the one remaining manual step is a
# single click on the viewer's own file control (or a drag) using the path
# just printed, instead of hunting for the file in a file manager. Same
# xdg-open/sensible-browser launch shape as _gopod_open_cockpit above.
gopod-json-view() {
  (
    set -u
    local target="${1:-}"
    if [ -z "$target" ]; then
      # No argument: just open the viewer, nothing else - no path, no
      # auto-load. Every viewable GOPOD json lives in one plain folder,
      # ~/gopod_jsons/ - load one yourself from there with the viewer's
      # own Load control.
      if [ ! -s "$GOPOD_JSON_VIEWER" ]; then
        echo "GOPOD_JSON_VIEW_VIEWER_MISSING path=$GOPOD_JSON_VIEWER"
        return 1
      fi
      if command -v firefox >/dev/null 2>&1; then
        nohup firefox "$GOPOD_JSON_VIEWER" >/tmp/gopod_json_view_browser_open.log 2>&1 &
      elif command -v xdg-open >/dev/null 2>&1; then
        nohup xdg-open "$GOPOD_JSON_VIEWER" >/tmp/gopod_json_view_browser_open.log 2>&1 &
      elif command -v sensible-browser >/dev/null 2>&1; then
        nohup sensible-browser "$GOPOD_JSON_VIEWER" >/tmp/gopod_json_view_browser_open.log 2>&1 &
      else
        echo "GOPOD_JSON_VIEW_NO_BROWSER_LAUNCHER tried firefox,xdg-open,sensible-browser"
        return 1
      fi
      echo "GOPOD_JSON_VIEW_LAUNCHED"
      return 0
    fi

    local knobs_path
    case "$target" in
      */*|*.json)
        knobs_path="$target"
        ;;
      *)
        knobs_path="$GOPOD_REPO/goverlord/runtime/songs/$target/knobs.json"
        if [ ! -s "$knobs_path" ]; then
          # Folder-name lookup failed - fall back to a real song_id scan.
          # This resolver was folder-name-only; folder name and song_id
          # coincided for every song until the 2026-08-19 interview
          # vamp/run split (VAMP's folder is 01_brobots_interview_vamp but
          # its song_id is brobots_preshow; RUN's folder is
          # 02_brobots_interview_run but its song_id is still
          # brobots_interview_section_01) - the first time they diverged.
          # GOPOLISHER_FIXES_001.md. Every other song still resolves via
          # the direct folder-name path above, unchanged - this scan only
          # runs when that fast path already missed.
          local candidate found_id
          for candidate in "$GOPOD_REPO"/goverlord/runtime/songs/*/knobs.json; do
            [ -s "$candidate" ] || continue
            found_id="$(grep -o '"song_id"[[:space:]]*:[[:space:]]*"[^"]*"' "$candidate" | head -1 | sed 's/.*"song_id"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')"
            if [ "$found_id" = "$target" ]; then
              knobs_path="$candidate"
              break
            fi
          done
        fi
        ;;
    esac

    if [ ! -s "$knobs_path" ]; then
      echo "GOPOD_JSON_VIEW_BLOCKED path=$knobs_path"
      return 1
    fi
    knobs_path="$(cd "$(dirname "$knobs_path")" && pwd)/$(basename "$knobs_path")"

    if [ ! -s "$GOPOD_JSON_VIEWER" ]; then
      echo "GOPOD_JSON_VIEW_VIEWER_MISSING path=$GOPOD_JSON_VIEWER"
      return 1
    fi

    echo "GOPOD_JSON_VIEW_FILE $knobs_path"
    echo "GOPOD_JSON_VIEW_NOTE viewer has no auto-load hook - use its Load control or drag this file in"

    if [ "${GOPOD_NO_BROWSER:-0}" = "1" ]; then
      return 0
    fi
    # firefox first, not xdg-open: on this machine xdg-open's registered
    # default is the Chromium flatpak, and that flatpak's GPU process fails
    # EGL/GLX init here (glXQueryExtensionsString returned NULL) and never
    # actually presents a window - confirmed live, xdg-open reports success
    # (exit 0) while nothing visible ever opens. Firefox (snap) is already
    # the operator's real running browser on this machine and opens a new
    # tab in it cleanly - confirmed live. Falls back to xdg-open/
    # sensible-browser only if firefox itself isn't there.
    if command -v firefox >/dev/null 2>&1; then
      nohup firefox "$GOPOD_JSON_VIEWER" >/tmp/gopod_json_view_browser_open.log 2>&1 &
    elif command -v xdg-open >/dev/null 2>&1; then
      nohup xdg-open "$GOPOD_JSON_VIEWER" >/tmp/gopod_json_view_browser_open.log 2>&1 &
    elif command -v sensible-browser >/dev/null 2>&1; then
      nohup sensible-browser "$GOPOD_JSON_VIEWER" >/tmp/gopod_json_view_browser_open.log 2>&1 &
    else
      echo "GOPOD_JSON_VIEW_NO_BROWSER_LAUNCHER tried firefox,xdg-open,sensible-browser"
      return 1
    fi
    echo "GOPOD_JSON_VIEW_LAUNCHED"
  )
}

GOPOD_INDEX_PAGE="${GOPOD_INDEX_PAGE:-$HOME/gopod_index.html}"

# gopod-index: opens the one-page GOPOD song file index
# (~/gopod_index.html) in the browser - nothing else. Same
# firefox-first/xdg-open/sensible-browser launch shape as gopod-json-view.
gopod-index() {
  (
    set -u
    if [ ! -s "$GOPOD_INDEX_PAGE" ]; then
      echo "GOPOD_INDEX_PAGE_MISSING path=$GOPOD_INDEX_PAGE"
      return 1
    fi
    if command -v firefox >/dev/null 2>&1; then
      nohup firefox "$GOPOD_INDEX_PAGE" >/tmp/gopod_index_browser_open.log 2>&1 &
    elif command -v xdg-open >/dev/null 2>&1; then
      nohup xdg-open "$GOPOD_INDEX_PAGE" >/tmp/gopod_index_browser_open.log 2>&1 &
    elif command -v sensible-browser >/dev/null 2>&1; then
      nohup sensible-browser "$GOPOD_INDEX_PAGE" >/tmp/gopod_index_browser_open.log 2>&1 &
    else
      echo "GOPOD_INDEX_NO_BROWSER_LAUNCHER tried firefox,xdg-open,sensible-browser"
      return 1
    fi
    echo "GOPOD_INDEX_LAUNCHED"
  )
}

gobingo() {
    (
        cd /home/goverlord/wire-pod || exit 1

        # Brobot 2 angry-animation reactor: launched here only for operator
        # convenience (one terminal, one command instead of two). Still a fully
        # independent process — Bingo's own code has zero awareness of it, no
        # import, no call site. See goverlord/runtime/songs/102_brobots_bingo_game/bingo_reactor/README.md.
        # Trap ensures it's killed when Bingo exits, however it exits.
        /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/102_brobots_bingo_game/bingo_reactor/.venv/bin/python3 \
            /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/102_brobots_bingo_game/bingo_reactor/bingo_reactor_001.py &
        local reactor_pid=$!
        trap 'kill "$reactor_pid" 2>/dev/null' EXIT INT TERM

        WIREPOD_EX_TMP_PATH=/home/goverlord/crushn8r_git/SDK/sources/vectorx/vectorfs/tmp \
        WIREPOD_EX_DATA_PATH=/home/goverlord/crushn8r_git/SDK/sources/vectorx/vectorfs/data \
        WIREPOD_EX_NVM_PATH=/home/goverlord/crushn8r_git/SDK/sources/vectorx/vectorfs/nvm \
        /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/102_brobots_bingo_game/bin/vectorx-gobingo \
            --serial 0dd1b9e9 \
            --locale en-US \
            --speechText bingo \
            "$@"
    )
}

# gobingo-reactor: starts Brobot 2 angry-animation watcher standalone.
# No longer needed for normal use — gobingo now launches and cleans this up
# automatically. Kept for manual/standalone debugging in its own terminal.
# Ctrl-C to stop.
# Uses the reactor's own venv (not system python3) since anki_vector is
# installed there, not system-wide.
alias gobingo-reactor='/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/102_brobots_bingo_game/bingo_reactor/.venv/bin/python3 /home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/102_brobots_bingo_game/bingo_reactor/bingo_reactor_001.py'

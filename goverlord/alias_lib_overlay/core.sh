# GOPOD_WIREPOD_BASE_URL etc. - real values live in a private, untracked
# local file, never a literal IP in this tracked/mirrored file (2026-08-15,
# GOPOLISHER_CUBE_SESSION_SWEEP_001.md). Sourced here too (not just
# brobots.sh) so this file works correctly even if ever sourced on its own.
GOPOD_LOCAL_CONFIG_PATH="${GOPOD_LOCAL_CONFIG_PATH:-$HOME/.gopod_alias_lib/local_config.sh}"
[ -f "$GOPOD_LOCAL_CONFIG_PATH" ] && . "$GOPOD_LOCAL_CONFIG_PATH"

# === GOPOD core ===
alias gorepo='cd ~/crushn8r_git/GOPOD'
alias gp='git pull --rebase && git push'

# === Codex ===
alias codex-1='CODEX_HOME=~/.codex_1 command codex'
alias codex-2='CODEX_HOME=~/.codex_2 command codex'
alias codex-3='CODEX_HOME=~/.codex_3 command codex'
alias codex-4='CODEX_HOME=~/.codex_4 command codex'
alias codex-5='CODEX_HOME=~/.codex_5 command codex'
alias codex-6='CODEX_HOME=~/.codex_6 command codex'

# === Network ===
alias nmg='nmcli -t -f DEVICE,TYPE,STATE,CONNECTION dev'
alias nmc='nmcli con show --active'
alias gopod-devices='nmcli -t -f DEVICE,TYPE,STATE,CONNECTION dev'
alias gopod-net='sudo /usr/local/sbin/gopod-usb-wifi-5g-recover auto'
alias gopod-net-usb='sudo /usr/local/sbin/gopod-usb-wifi-5g-recover wlan-adapter-1'
alias gopod-net-fix='sudo /usr/local/sbin/gopod-net-fix'
alias gopod-scan-usb='nmcli -f IN-USE,SSID,CHAN,FREQ,SIGNAL,SECURITY dev wifi list ifname wlan-adapter-1'

# === Audio input ===
# Detects the USB mic among PulseAudio's input sources - never hardcodes a
# device name (USB serials in pactl source names, e.g. "...330212E9240828...",
# aren't stable across a device swap). Reuses the same truth config the PTT
# writer's own USB-mic match already uses:
# goverlord/runtime/data_gomad/configs/audio.json (usb_audio_match_string,
# usb_audio_exclude_terms) - one source of truth, not a second guess. Path
# corrected 2026-07-30, gomads/ removed (GOMADS_TREE_REMOVED_001.md).
gopod-mic-detect() {
  python3 - <<'PYEOF'
import json
import subprocess
import sys

cfg = json.load(open("/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/data_gomad/configs/audio.json"))
match = cfg["usb_audio_match_string"].lower()
excludes = [t.lower() for t in cfg["usb_audio_exclude_terms"]]

out = subprocess.run(["pactl", "list", "sources"], capture_output=True, text=True, check=True).stdout

sources, current = [], {}
for line in out.splitlines():
    line = line.strip()
    if line.startswith("Source #"):
        if current:
            sources.append(current)
        current = {}
    elif line.startswith("Name:"):
        current["name"] = line.split(":", 1)[1].strip()
    elif line.startswith("Description:"):
        current["description"] = line.split(":", 1)[1].strip()
if current:
    sources.append(current)

candidates = []
for s in sources:
    name = s.get("name", "")
    haystack = (name + " " + s.get("description", "")).lower()
    if ".monitor" in name.lower():
        continue  # a monitor mirrors an output sink, not a real input mic
    if match not in haystack:
        continue
    if any(term in haystack for term in excludes):
        continue
    candidates.append(name)

if not candidates:
    print("GOPOD_MIC_DETECT_FAIL no matching USB input source", file=sys.stderr)
    sys.exit(1)
print(candidates[0])
PYEOF
}

gopod-mic-set() {
  local mic active
  mic="$(gopod-mic-detect)" || return 1
  pactl set-default-source "$mic" || { echo "GOPOD_MIC_SET_FAIL could not set default source to $mic"; return 1; }
  active="$(pactl info | awk -F': ' '/Default Source/{print $2}')"
  if [ "$active" != "$mic" ]; then
    echo "GOPOD_MIC_SET_FAIL default source is '$active', expected '$mic'"
    return 1
  fi
  echo "GOPOD_MIC_SET source=$mic"
}

# Real end-to-end proof, not just a device-name check: records ~1.5s from
# whatever is currently the default source and reports the actual peak
# sample amplitude, so a silently-dead/suspended device shows 0 instead of
# a false PASS.
gopod-mic-test() {
  local mic wav pid
  mic="$(pactl info | awk -F': ' '/Default Source/{print $2}')"
  wav="$(mktemp --suffix=.wav)"
  parecord --channels=1 --rate=16000 --file-format=wav "$wav" >/dev/null 2>&1 &
  pid=$!
  sleep 3  # USB source starts SUSPENDED; needs a beat to resume before frames flow
  kill -INT "$pid" 2>/dev/null  # SIGTERM truncates the WAV to 0 frames; parecord only finalizes the header on SIGINT
  wait "$pid" 2>/dev/null
  python3 - "$wav" "$mic" <<'PYEOF'
import audioop
import sys
import wave

wav_path, mic = sys.argv[1], sys.argv[2]
with wave.open(wav_path, "rb") as w:
    frames = w.readframes(w.getnframes())
    width = w.getsampwidth()
peak = audioop.max(frames, width) if frames else 0
print(f"GOPOD_MIC_TEST source={mic} wav={wav_path} bytes={len(frames)} peak={peak}")
PYEOF
}

# === Opening chord ===
# One command to stage-prep before any GOPOD song (interview, bingo). Per
# OPENING_CHORD_DEPENDENCY_MAP_001.md: four independent notes fire together
# (mic, LLM warm, Kokoro warm, and the whole Wire-Pod chain as one unit),
# because within the Wire-Pod chain restart->ready->wake-robot-1->wake-robot-2
# is genuinely sequential - Wire-Pod's own newRobot() serializes robot
# connects via a global inhibitCreation flag (chipper/pkg/wirepod/sdkapp/robot.go),
# no chord design on the outside changes that. The conductor waits for all
# four background jobs, then reports each note true/false and one final
# READY / NOT-READY line - never a false READY.
#
# Reuses existing code, does not reimplement it: the Wire-Pod chain job
# imports run_section1_full_live_001.py and calls its own
# restart_wirepod_preflight() (the real systemctl-active + HTTP-poll +
# settle pattern - not wpr's flat sleep 2) and wirepod_web_send_form()
# against /api-sdk/conn_test (the endpoint whose own Go comment says
# "getRobot does connection check and will return error if failed" - the
# real per-robot wake/connectivity check). The Kokoro job imports the same
# file and calls its kokoro_status_warm_up() directly, unchanged.
_GOPOD_CHORD_RUNNER_PATH="/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/songs/tools/run_section1_full_live_001.py"

_gopod_chord_kokoro_job() {
  timeout 60 python3 - <<PYEOF
import contextlib
import importlib.util
import io
import sys

spec = importlib.util.spec_from_file_location("run_section1_full_live_001", "$_GOPOD_CHORD_RUNNER_PATH")
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)

buf = io.StringIO()
with contextlib.redirect_stdout(buf):
    mod.kokoro_status_warm_up()
output = buf.getvalue()
print(output, end="")
sys.exit(0 if "KOKORO_WARMUP_DONE" in output else 1)
PYEOF
}

_gopod_chord_wirepod_job() {
  timeout 150 python3 - <<PYEOF
import importlib.util
import os
import sys
import time

os.environ["GOPOD_RESTART_WIREPOD_BEFORE_RUN"] = "1"
spec = importlib.util.spec_from_file_location("run_section1_full_live_001", "$_GOPOD_CHORD_RUNNER_PATH")
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)

def speak(serial, text, display_role):
    """Same path Robots.say() uses: assume -> say_text (real text, not the
    empty-text log-mirror trick) -> release. say_text's HTTP response only
    returns after the robot's SayText behavior completes (matches the
    interview runner's own assumption - it never waits extra between calling
    say() and moving to the next turn), so the elapsed time reported here is
    real speech time, not just a request round-trip."""
    mod.wirepod_web_send_form("/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10)
    try:
        say = mod.wirepod_web_send_form(
            "/api-sdk/say_text",
            {"text": text, "serial": serial, "display_role": display_role, "display_text": text},
            timeout=90,
        )
        ok = say.get("status") == 200 and str(say.get("body", "")).strip() == "success"
        return ok, say
    finally:
        mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)

print("STATUS: WIREPOD_CHAIN_RESTART_START", flush=True)
try:
    result = mod.restart_wirepod_preflight(True)
    print(f"STATUS: WIREPOD_CHAIN_RESTART_DONE {result}", flush=True)
except Exception as exc:  # noqa: BLE001
    print(f"STATUS: WIREPOD_CHAIN_RESTART_ERROR {type(exc).__name__}: {exc}", flush=True)
    sys.exit(1)

READY_LINES = {
    "BROBOT_1": ("Brobot 1, ready for the interview.", mod.BROBOT_1_SERIAL),
    "BROBOT_2": ("Brobot 2, ready to interview.", mod.BROBOT_2_SERIAL),
}

for label, (line, serial) in READY_LINES.items():
    print(f"STATUS: WIREPOD_CHAIN_WAKE_START {label}", flush=True)
    try:
        resp = mod.wirepod_web_send_form("/api-sdk/conn_test", {"serial": serial}, timeout=20)
        body = str(resp.get("body", "")).strip()
        if resp.get("status") == 200 and body == "success":
            print(f"STATUS: WIREPOD_CHAIN_WAKE_DONE {label}", flush=True)
        else:
            print(f"STATUS: WIREPOD_CHAIN_WAKE_FAILED {label} body={body!r}", flush=True)
            sys.exit(1)
    except Exception as exc:  # noqa: BLE001
        print(f"STATUS: WIREPOD_CHAIN_WAKE_ERROR {label} {type(exc).__name__}: {exc}", flush=True)
        sys.exit(1)

    # The speech IS the note here, not decoration - the robot must fully
    # finish saying its line (say_text returns) before this label counts
    # as done, and before the loop moves on to the next robot.
    print(f"STATUS: WIREPOD_CHAIN_SPEAK_START {label} text={line!r}", flush=True)
    t0 = time.time()
    try:
        ok, say = speak(serial, line, label.replace("_", " ").title())
        elapsed = time.time() - t0
        if ok:
            print(f"STATUS: WIREPOD_CHAIN_SPEAK_DONE {label} elapsed={elapsed:.2f}s", flush=True)
        else:
            print(f"STATUS: WIREPOD_CHAIN_SPEAK_FAILED {label} elapsed={elapsed:.2f}s body={say.get('body','')!r}", flush=True)
            sys.exit(1)
    except Exception as exc:  # noqa: BLE001
        print(f"STATUS: WIREPOD_CHAIN_SPEAK_ERROR {label} {type(exc).__name__}: {exc}", flush=True)
        sys.exit(1)

print("STATUS: WIREPOD_CHAIN_DONE", flush=True)
PYEOF
}

# The "Brobots ready!" finale - fired only after every other note (including
# each robot's individual ready line) is confirmed done. Per
# DIRECT_SDK_BROBOTS_READY_HANDOFF_INVESTIGATION_001.md: this is now a
# decoupled side-road, not the Wire-Pod say path the rest of the chord uses.
#
# Shape: (1) Wire-Pod releases both robots via its own existing
# /api-sdk/release_behavior_control - no new endpoint. (2) Wait past
# Wire-Pod's own ~500ms internal release-polling window
# (chipper/pkg/wirepod/sdkapp/bcassume.go's actual polling constant) before
# the direct path asks for control - 1000ms used here, a 2x margin over that
# documented constant, not a guess. (3) direct_sdk_brobots_ready_001 (built
# from gopod_probes/tools/direct_sdk_brobots_ready_001.go, vendored
# fforchino/vector-go-sdk, chipper's own already-resolved go.mod/go.sum, no
# new dependency) connects to both robots directly via vector.NewWP() -
# which itself already reads chipper/jdocs/botSdkInfo.json, the exact file
# Wire-Pod's own newRobot() reads - requests BehaviorControl (same
# ControlRequest/priority shape bcassume.go uses), fires "Brobots ready!" on
# both concurrently via goroutines in one process (no separate-process
# import jitter between the two, unlike the old Python version), releases,
# and disconnects. Hand-back needs nothing else: Wire-Pod's own cached
# connection per robot was never torn down, only its behavior-control claim
# was released - the next real /api-sdk/assume_behavior_control call (the
# actual interview turns) works normally on the connection Wire-Pod already
# had.
_GOPOD_CHORD_DIRECT_TOGETHER_BIN="/home/goverlord/wire-pod/chipper/gopod_probes/tools/direct_sdk_brobots_ready_001"
_GOPOD_CHORD_RELEASE_SETTLE_SECONDS=1

_gopod_chord_release_both() {
  local base="${GOPOD_WIREPOD_BASE_URL}"
  local s1="${GOPOD_BROBOT_1_SERIAL:-0dd1b9e9}"
  local s2="${GOPOD_BROBOT_2_SERIAL:-0dd1d8bf}"
  curl -fsS --max-time 10 -X POST --data-urlencode "serial=$s1" "$base/api-sdk/release_behavior_control" >/dev/null 2>&1
  curl -fsS --max-time 10 -X POST --data-urlencode "serial=$s2" "$base/api-sdk/release_behavior_control" >/dev/null 2>&1
}

_gopod_chord_direct_together_job() {
  local outfile="$1"
  _gopod_chord_release_both
  sleep "$_GOPOD_CHORD_RELEASE_SETTLE_SECONDS"
  # The vendored SDK's own NewWP() (pkg/vector/vector.go) calls log.Println
  # with the robot's auth GUID in plaintext, to stderr - not our code, not
  # edited (scope: no changes to the vendored SDK). Filtered in-flight, in
  # the same pipeline that writes $outfile, so the token is never written to
  # disk at all - not captured then scrubbed, never captured in the first
  # place. PIPESTATUS[0] (not $?) is required here to still return the
  # binary's own exit code rather than grep's, since $? after a pipe
  # reflects the last command by default.
  WIREPOD_HOME=/home/goverlord/wire-pod timeout 30 "$_GOPOD_CHORD_DIRECT_TOGETHER_BIN" \
    "Brobots ready!" \
    "Brobot 1:${GOPOD_BROBOT_1_SERIAL:-0dd1b9e9}" \
    "Brobot 2:${GOPOD_BROBOT_2_SERIAL:-0dd1d8bf}" \
    2>&1 | grep -vi "guid" > "$outfile"
  return "${PIPESTATUS[0]}"
}

gopod_brobots() {
  local workdir
  workdir="$(mktemp -d)"

  # `timeout` execs a real binary - it can't invoke a bash function directly.
  # export -f + `timeout N bash -c 'func ...'` is the standard way to give it
  # one anyway. The kokoro/wirepod jobs are plain `python3` calls internally,
  # so their own `timeout N python3 ...` works natively, no export needed.
  export -f gopod-mic-detect gopod-mic-set _gopod_llm_ping >/dev/null 2>&1

  ( timeout 40 bash -c 'gopod-mic-set'; echo $? > "$workdir/mic.status" ) > "$workdir/mic.log" 2>&1 &
  local mic_pid=$!

  ( timeout 60 bash -c '_gopod_llm_ping "$1" "$2"' _ "${GOPOD_BROBOT_LLM:-gemma2:2b}" "Say GOPOD ready in one short sentence."; echo $? > "$workdir/llm.status" ) > "$workdir/llm.log" 2>&1 &
  local llm_pid=$!

  ( _gopod_chord_kokoro_job; echo $? > "$workdir/kokoro.status" ) > "$workdir/kokoro.log" 2>&1 &
  local kokoro_pid=$!

  ( _gopod_chord_wirepod_job; echo $? > "$workdir/wirepod.status" ) > "$workdir/wirepod.log" 2>&1 &
  local wirepod_pid=$!

  echo "GOPOD_OPENING_CHORD_START workdir=$workdir"
  wait "$mic_pid" "$llm_pid" "$kokoro_pid" "$wirepod_pid"

  local mic_rc llm_rc kokoro_rc wirepod_rc
  mic_rc="$(cat "$workdir/mic.status" 2>/dev/null || echo 1)"
  llm_rc="$(cat "$workdir/llm.status" 2>/dev/null || echo 1)"
  kokoro_rc="$(cat "$workdir/kokoro.status" 2>/dev/null || echo 1)"
  wirepod_rc="$(cat "$workdir/wirepod.status" 2>/dev/null || echo 1)"

  local wirepod_ready=1 brobot1_awake=1 brobot1_line=1 brobot2_awake=1 brobot2_line=1
  grep -q "WIREPOD_CHAIN_RESTART_DONE" "$workdir/wirepod.log" 2>/dev/null && wirepod_ready=0
  grep -q "WIREPOD_CHAIN_WAKE_DONE BROBOT_1" "$workdir/wirepod.log" 2>/dev/null && brobot1_awake=0
  grep -q "WIREPOD_CHAIN_SPEAK_DONE BROBOT_1" "$workdir/wirepod.log" 2>/dev/null && brobot1_line=0
  grep -q "WIREPOD_CHAIN_WAKE_DONE BROBOT_2" "$workdir/wirepod.log" 2>/dev/null && brobot2_awake=0
  grep -q "WIREPOD_CHAIN_SPEAK_DONE BROBOT_2" "$workdir/wirepod.log" 2>/dev/null && brobot2_line=0

  local note_line
  note_line() {
    local name="$1" rc="$2"
    if [ "$rc" = "0" ]; then
      echo "NOTE $name: done"
    else
      echo "NOTE $name: failed"
    fi
  }

  note_line "mic_set" "$mic_rc"
  note_line "llm_warm" "$llm_rc"
  note_line "kokoro_warm" "$kokoro_rc"
  note_line "wirepod_ready" "$wirepod_ready"
  note_line "brobot_1_awake" "$brobot1_awake"
  note_line "brobot_1_ready_line" "$brobot1_line"
  note_line "brobot_2_awake" "$brobot2_awake"
  note_line "brobot_2_ready_line" "$brobot2_line"

  local failed=()
  [ "$mic_rc" = "0" ] || failed+=("mic_set")
  [ "$llm_rc" = "0" ] || failed+=("llm_warm")
  [ "$kokoro_rc" = "0" ] || failed+=("kokoro_warm")
  [ "$wirepod_ready" = "0" ] || failed+=("wirepod_ready")
  [ "$brobot1_awake" = "0" ] || failed+=("brobot_1_awake")
  [ "$brobot1_line" = "0" ] || failed+=("brobot_1_ready_line")
  [ "$brobot2_awake" = "0" ] || failed+=("brobot_2_awake")
  [ "$brobot2_line" = "0" ] || failed+=("brobot_2_ready_line")

  # The baton does not drop yet even if every note above is done - the
  # simultaneous "Brobots ready!" finale is itself the last note, fired only
  # now, after both individual ready lines have already fully landed. Now a
  # decoupled direct-SDK side-road (see the function comment above) instead
  # of the Wire-Pod say path the rest of the chord uses.
  if [ "${#failed[@]}" -eq 0 ]; then
    local together_dir="$workdir/together"
    mkdir -p "$together_dir"
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
    rows[label] = {'serial': serial, 'start': float(start), 'end': float(end), 'elapsed': float(elapsed), 'status': status}
b1, b2 = rows.get('Brobot 1'), rows.get('Brobot 2')
if b1 and b2:
    overlap = min(b1['end'], b2['end']) - max(b1['start'], b2['start'])
    gap = abs(b1['start'] - b2['start'])
    print(f\"TOGETHER (direct SDK) start-gap={gap:.3f}s overlap={overlap:.3f}s\" + (\" (genuinely overlapped)\" if overlap > 0 else \" (did not overlap)\"))
"
    fi

    local brobots_together=1
    [ "$together_rc" = "0" ] && brobots_together=0
    note_line "brobots_together" "$brobots_together"
    [ "$brobots_together" = "0" ] || failed+=("brobots_together")
  fi

  if [ "${#failed[@]}" -eq 0 ]; then
    echo "READY - all notes played, stage is set."
    echo "logs: $workdir"
    return 0
  else
    local joined
    joined="$(IFS=,; echo "${failed[*]}")"
    echo "NOT-READY-because-$joined"
    echo "logs: $workdir"
    return 1
  fi
}

# Back-compat: pure rename, byte-identical body - see SECTION1_STAGE_RENAME
# report in gopod_notes/ for the rest of the Section 1 stage-key rename.
alias gopod-opening-chord='gopod_brobots'

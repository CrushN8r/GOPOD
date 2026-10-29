#!/usr/bin/env bash
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/gopod_full_audit.sh
# GOPOD's full audit (2026-10-25, operator: "a .sh file to run after a wpr 2/F5 *
# 100% everything audit * full web page run + phone app * just not with the actual
# robots being called * focus more on mechanical errors more than 'look'").
#
# Run it after wpr 2 and F5, with the chat page open:
#     ~/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_probes/tools/gopod_full_audit.sh
#
# Nothing here calls a robot: the robots check only reads the battery reports
# they already sent. Checks, in order:
#   1. Wire-Pod      running, its voices ready
#   2. Files         settings and descriptions are valid JSON, the chat page's code
#                    parses, every probe tool's Python compiles
#   3. Go tests      every test of the Brobots chat (no build, no swap)
#   4. Web page      the page, its code, styles, settings, feed and mirror answer
#   5. Phone app     the secure page, its manifest, worker, icons and certificate
#   6. System tests  system-test-all (the audit list) and system-test-eggvoice,
#                    run inside Wire-Pod - the game, the page replay with a
#                    visitor's page, voices, TPU, LLM, search, weather ...
#   7. Log           Wire-Pod's log since its start: panics, page errors, stuck
#                    lines, unheard voices, fallback answers, refused reports
# A report goes to ~/crushn8r_git/gopod_notes/AUDIT_REPORT_<date>_<time>.md.
set -u   # no pipefail: grep -q / -m1 ending a pipe early isn't a failure
here="$(cd "$(dirname "$0")" && pwd)"
chipper="$(cd "$here/../.." && pwd)"
overlay="$(cd "$chipper/.." && pwd)"
wirepod_chipper="$HOME/wire-pod/chipper"
notes="$HOME/crushn8r_git/gopod_notes"
base="http://localhost:8080"
stamp="$(date +%Y%m%d_%H%M)"
report="$notes/AUDIT_REPORT_${stamp}.md"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
pass=0; fail=0; lines=()

check() { # check <area> <name> <ok 0|1> <detail>
  local mark="PASS"
  if [ "$3" != "0" ]; then mark="FAIL"; fail=$((fail + 1)); else pass=$((pass + 1)); fi
  lines+=("| $1 | $2 | $mark | ${4//|//} |")
  printf '%-4s %-12s %s - %s\n' "$mark" "$1" "$2" "$4"
}
http_ok() { # http_ok <url> [curl args...] -> status code
  curl -s -o /dev/null -w '%{http_code}' --max-time 15 "$@"
}

echo "GOPOD full audit - $(date '+%Y-%m-%d %H:%M')"

# 1. Wire-Pod -----------------------------------------------------------------
started="$(systemctl show -p ActiveEnterTimestamp --value wire-pod 2>/dev/null)"
since="$(date -d "$started" '+%Y-%m-%d %H:%M:%S' 2>/dev/null || date -d '-1 hour' '+%Y-%m-%d %H:%M:%S')"
if systemctl is-active --quiet wire-pod; then check "Wire-Pod" "running" 0 "since $since"; else check "Wire-Pod" "running" 1 "wire-pod.service isn't active"; fi
for i in $(seq 1 60); do
  journalctl -a -u wire-pod --since "$since" --no-pager 2>/dev/null | grep -aq "GOPOD_REPORTER_VOICES all lines ready" && break
  sleep 5
done
if journalctl -a -u wire-pod --since "$since" --no-pager | grep -aq "GOPOD_REPORTER_VOICES all lines ready"; then
  check "Wire-Pod" "voices ready" 0 "the voice worker made every line"
else
  check "Wire-Pod" "voices ready" 1 "no 'all lines ready' in 5 minutes"
fi

# 2. Files --------------------------------------------------------------------
for f in "$chipper"/webroot/gopod_settings/*.json "$chipper/webroot/gopod_picture_descriptions.json" "$chipper/webroot/app/gopod_manifest.json" "$overlay/overlay.json"; do
  if err="$(python3 -I -c 'import json,sys; json.load(open(sys.argv[1]))' "$f" 2>&1)"; then check "Files" "$(basename "$f")" 0 "valid JSON"; else check "Files" "$(basename "$f")" 1 "${err##*: }"; fi
done
for f in "$chipper"/webroot/js/gopod_*.js "$chipper/webroot/gopod_sw.js" "$here/gopod_page_replay_001.js"; do
  [ -f "$f" ] || continue
  if err="$(node --check "$f" 2>&1)"; then check "Files" "$(basename "$f")" 0 "parses"; else check "Files" "$(basename "$f")" 1 "$(echo "$err" | grep -m1 -i error)"; fi
done
bad=""
for f in "$here"/*.py; do
  python3 -I -c 'import sys; compile(open(sys.argv[1]).read(), sys.argv[1], "exec")' "$f" 2>"$tmp/py.err" || bad="$bad $(basename "$f")"
done
if [ -z "$bad" ]; then check "Files" "probe tools (Python)" 0 "$(ls "$here"/*.py | wc -l) compile"; else check "Files" "probe tools (Python)" 1 "don't compile:$bad"; fi

# 3. Go tests -----------------------------------------------------------------
echo "... Go tests (a minute or two)"
cp "$overlay/overlay.json" "$tmp/ov.json"
# A clean build setting (run from inside Wire-Pod, its own point at root's Vosk folder).
if env -u CGO_CFLAGS -u CGO_LDFLAGS -u CGO_ENABLED -u LD_LIBRARY_PATH -u C_INCLUDE_PATH -u LIBRARY_PATH -u GOLDFLAGS -u GOTAGS \
  GOPOD_OVERLAY_CHIPPER="$chipper" go -C "$wirepod_chipper" test -count=1 -tags nolibopusfile -overlay="$tmp/ov.json" ./pkg/wirepod/ttr/ >"$tmp/go.log" 2>&1; then
  check "Go tests" "the Brobots chat" 0 "$(tail -1 "$tmp/go.log")"
else
  why="$(grep -m3 -E '^--- FAIL|_test.go:[0-9]+:|\.go:[0-9]+:[0-9]+:' "$tmp/go.log" | tr '\n' ' ')"
  [ -n "$why" ] || why="$(tail -3 "$tmp/go.log" | tr '\n' ' ')"
  check "Go tests" "the Brobots chat" 1 "$why"
fi

# 4. Web page -----------------------------------------------------------------
for p in "/" "/js/gopod_brobots_chat.js" "/css/gopod_chat_bubbles.css" "/gopod_brobots_settings.json" "/gopod_chat_feed.json" "/gopod_speaking_now.json" "/gopod_picture_descriptions.json"; do
  code="$(http_ok "$base$p")"
  if [ "$code" = "200" ]; then check "Web page" "$p" 0 "200"; else check "Web page" "$p" 1 "HTTP $code"; fi
done
code="$(http_ok "$base/api-sdk/gopod_display?since=0")"
if [ "$code" = "200" ] || [ "$code" = "204" ]; then check "Web page" "display mirror" 0 "HTTP $code"; else check "Web page" "display mirror" 1 "HTTP $code"; fi
feedok="$(curl -s --max-time 10 "$base/gopod_chat_feed.json" | python3 -I -c 'import json,sys; d=json.load(sys.stdin); ids=[e["id"] for e in d["events"]]; print("ok" if ids==sorted(set(ids)) else "out of order or repeated ids")' 2>&1)"
if [ "$feedok" = "ok" ]; then check "Web page" "feed in order" 0 "ids rising, none repeated"; else check "Web page" "feed in order" 1 "$feedok"; fi

# 5. Phone app ----------------------------------------------------------------
port="$(python3 -I -c 'import json,glob,sys; m={}; [m.update(json.load(open(f))) for f in sorted(glob.glob(sys.argv[1]+"/*.json"))]; print(m["app"]["https_port"])' "$chipper/webroot/gopod_settings" 2>/dev/null || echo 8443)"
sbase="https://localhost:$port"
ca="$HOME/.gopod_https/ca.crt"
for p in "/" "/app/gopod_manifest.json" "/gopod_sw.js" "/app/gopod_icon_192.png" "/app/gopod_icon_512.png" "/gopod_ca.crt" "/?app"; do
  code="$(http_ok "$sbase$p" -k)"
  if [ "$code" = "200" ]; then check "Phone app" "$p" 0 "200 on the secure page"; else check "Phone app" "$p" 1 "HTTP $code"; fi
done
if [ -f "$ca" ] && [ "$(http_ok "$sbase/" --cacert "$ca")" = "200" ]; then
  check "Phone app" "certificate" 0 "the page's certificate is trusted by GOPOD's own CA"
else
  check "Phone app" "certificate" 1 "the secure page isn't trusted by $ca"
fi
if [ -f "$HOME/.gopod_https/server.crt" ]; then
  end="$(openssl x509 -in "$HOME/.gopod_https/server.crt" -noout -enddate 2>/dev/null | cut -d= -f2)"
  if openssl x509 -in "$HOME/.gopod_https/server.crt" -noout -checkend 2592000 >/dev/null 2>&1; then check "Phone app" "certificate dates" 0 "good until $end"; else check "Phone app" "certificate dates" 1 "ends within 30 days ($end)"; fi
fi
if curl -sk --max-time 10 "$sbase/app/gopod_manifest.json" | python3 -I -c 'import json,sys; m=json.load(sys.stdin); assert m.get("name") and m.get("icons") and m.get("start_url")' 2>/dev/null; then
  check "Phone app" "manifest" 0 "name, icons and start page"
else
  check "Phone app" "manifest" 1 "missing name, icons or start_url"
fi

# 6. System tests (inside Wire-Pod) -------------------------------------------
run_system() { # run_system <command> -> the GOPOD_SYSTEM_SET result line
  local at; at="$(date '+%Y-%m-%d %H:%M:%S')"
  curl -s -X POST --max-time 15 --data-urlencode "serial=ESN01" --data-urlencode "text=$1" "$base/api-sdk/gopod_typed_input" >/dev/null
  for i in $(seq 1 120); do
    sleep 3
    out="$(journalctl -a -u wire-pod --since "$at" --no-pager | grep -a -m1 "GOPOD_SYSTEM_SET system-test")" && [ -n "$out" ] && { echo "${out#*GOPOD_SYSTEM_SET }"; return; }
  done
  echo "no result in 6 minutes"
}
echo "... system tests inside Wire-Pod (a few minutes)"
for cmd in "system-test-all" "system-test-eggvoice"; do
  res="$(run_system "$cmd")"
  # One row per test: "PASS name: detail" / "FAIL name: detail".
  echo "$res" | sed 's/ | /\n/g' | while IFS= read -r part; do
    case "$part" in
      PASS\ *) n="${part#PASS }"; echo "0|${n%%:*}|${n#*: }" ;;
      FAIL\ *) n="${part#FAIL }"; echo "1|${n%%:*}|${n#*: }" ;;
    esac
  done >"$tmp/sys.rows"
  if [ ! -s "$tmp/sys.rows" ]; then check "System" "$cmd" 1 "$res"; fi
  while IFS='|' read -r ok name detail; do check "System" "$name" "$ok" "$detail"; done <"$tmp/sys.rows"
done

# 7. Log ----------------------------------------------------------------------
journalctl -a -u wire-pod --since "$since" --no-pager >"$tmp/log" 2>/dev/null
logcheck() { # logcheck <name> <pattern> <why>
  local n; n="$(grep -a -c -E "$2" "$tmp/log")"
  if [ "$n" = "0" ]; then check "Log" "$1" 0 "none since the start"; else check "Log" "$1" 1 "${n}x - $3: $(grep -a -m1 -E "$2" "$tmp/log" | cut -d' ' -f6- | cut -c1-140)"; fi
}
logcheck "panics" "panic:|goroutine [0-9]+ \\[running\\]" "Wire-Pod crashed"
logcheck "page errors" "GOPOD_PAGE error" "the chat page broke on a line"
logcheck "stuck lines" "stuck [0-9]+ s on feed" "a line held the chat"
logcheck "missing bubbles" "GOPOD_FEED_CHECK" "the page played past a reply with no bubble"
logcheck "unheard voices" "not heard after" "a voice the page never played"
logcheck "fallback answers" "BROBOT_FALLBACK" "the fallback line instead of an answer"
logcheck "LLM errors" "LLM error" "the model failed a reply"
logcheck "refused reports" "not taken:" "the page's played report wasn't taken"
logcheck "hold gave up" "GOPOD_EXCHANGE gave up" "a follow-up never let its line go"
logcheck "voice worker" "GOPOD_REPORTER_VOICES worker gone|unable to make a line" "a voice couldn't be made"

# Report ----------------------------------------------------------------------
total=$((pass + fail))
bottom="all $total checks pass"
[ "$fail" -gt 0 ] && bottom="$fail of $total checks fail"
mkdir -p "$notes"
{
  echo "# GOPOD full audit $(date '+%Y-%m-%d %H:%M')"
  echo
  echo "**Bottom line:** $bottom. No robot was called."
  echo
  if [ "$fail" -gt 0 ]; then
    echo "## Failed"
    echo
    echo "| Area | Check | Result | Detail |"
    echo "|---|---|---|---|"
    printf '%s\n' "${lines[@]}" | grep '| FAIL |'
    echo
  fi
  echo "## Every check"
  echo
  echo "| Area | Check | Result | Detail |"
  echo "|---|---|---|---|"
  printf '%s\n' "${lines[@]}"
} >"$report"
echo
echo "$( [ "$fail" -gt 0 ] && echo "FAIL" || echo "PASS" ): $bottom - $report"
[ "$fail" -eq 0 ]

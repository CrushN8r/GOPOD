#!/usr/bin/env bash
# goverlord/wire_pod_overlay/gopod_autofix.sh
#
# GOPOD auto-fixer (2026-11-03, operator: "MAKE A DIAGNOSTIC AUTO-FIXER"; "Ask
# y/n each fix"; "Don't save logs"; "Yes, show why" Wire-Pod restarted). Checks
# from the root up - repo, the batch files, the deploy, the live binary, the
# service, the log - and for each problem shows the fix and asks y/n first.
# Saves nothing. Run: bash goverlord/wire_pod_overlay/gopod_autofix.sh
set -u
R="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
O="$R/goverlord/wire_pod_overlay"; C="$O/chipper"
W="${GOPOD_WIREPOD_DIR:-/home/goverlord/wire-pod}"; B="$W/chipper/chipper"
BR=claude/magical-fermat-2dbhdj
F=0; FIXED=0
ok(){ echo "OK   $*"; }; bad(){ echo "FAIL $*"; F=$((F+1)); }
ask(){ # ask "what the fix does" command...
  local what="$1"; shift
  read -r -p "  fix: $what? [y/n] " a </dev/tty
  if [ "$a" = y ]; then "$@" && { FIXED=$((FIXED+1)); return 0; }; echo "  the fix failed"; fi
  return 1
}
cd "$R" || exit 1

echo "== 1. Repo ($R)"
git fetch -q origin "$BR" 2>/dev/null || bad "git fetch failed (network?)"
echo "     branch $(git branch --show-current), here $(git log --oneline -1)"
[ "$(git branch --show-current)" = "$BR" ] || bad "not on $BR"
N=$(git rev-list --count HEAD..FETCH_HEAD 2>/dev/null || echo 0)
if [ "$N" != 0 ]; then
  bad "$N commit(s) behind"
  ask "pull them (your local edits kept)" git pull --no-rebase --autostash --no-edit origin "$BR"
else ok "up to date"; fi
[ -z "$(git status --porcelain -uno)" ] && ok "no local edits" || echo "NOTE local edits: $(git status --porcelain -uno | tr '\n' ' ')"

echo "== 2. Files"
for j in "$C"/webroot/gopod_settings/*.json "$O/overlay.json"; do python3 -m json.tool "$j" >/dev/null 2>&1 || bad "bad JSON: ${j#$R/}"; done
[ -f "$C/gopod_probes/locks/searching_golden.md" ] && ok "searching golden note" || bad "searching golden note missing"
MISS=$(cd "$O" && python3 - <<'PY'
import json,glob,os
o=json.load(open('overlay.json'))['Replace']; here=os.getcwd()+'/'
srcs={v.replace('/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/',here) for v in o.values()}
print(' '.join(f for f in glob.glob('chipper/pkg/**/*.go',recursive=True) if here+f not in srcs))
PY
)
[ -z "$MISS" ] && ok "every overlay .go file is in overlay.json" || bad "not in overlay.json (not built): $MISS"

echo "== 3. Deploy and live binary"
L=$(ls -t /tmp/chipper_overlay_tests_*.log 2>/dev/null | head -1)
LASTFAIL=0
if [ -n "$L" ] && grep -qE '^(--- FAIL|FAIL)' "$L"; then
  LASTFAIL=1; echo "NOTE last deploy's tests failed ($(date -r "$L" '+%F %T')):"; grep -E '^(--- FAIL|FAIL)|_test.go:' "$L" | head -10 | sed 's/^/     /'
fi
BT=$(stat -c %Y "$B" 2>/dev/null || echo 0); GT=$(git log -1 --format=%ct -- "$C/pkg")
if [ "$BT" -lt "$GT" ]; then
  bad "live binary older than the last Go change"
  if ask "run the deploy (tests, build, swap)" bash "$O/deploy_overlay_build.sh"; then
    BT=$(stat -c %Y "$B" 2>/dev/null || echo 0)
  else
    L=$(ls -t /tmp/chipper_overlay_tests_*.log 2>/dev/null | head -1)
    [ -n "$L" ] && grep -E '^(--- FAIL|FAIL)|_test.go:' "$L" | head -10 | sed 's/^/     /'
  fi
else ok "live binary is newer than the last Go change"; [ $LASTFAIL = 1 ] && echo "NOTE (the failed tests above are from an older deploy)"; fi

echo "== 4. Service"
if systemctl is-active -q wire-pod; then ok "wire-pod running"; else bad "wire-pod not running"; fi
ST=$(date -d "$(systemctl show wire-pod -p ActiveEnterTimestamp --value)" +%s 2>/dev/null || echo 0)
if [ "$ST" -lt "$BT" ] || ! systemctl is-active -q wire-pod; then
  bad "wire-pod not restarted since the deploy"
  ask "restart it with your own wpr 2" bash -ic 'wpr 2' && ST=$(date -d "$(systemctl show wire-pod -p ActiveEnterTimestamp --value)" +%s 2>/dev/null || echo 0)
else ok "restarted after the deploy"; fi

echo "== 5. Why Wire-Pod last restarted"
JL(){ journalctl "$@" --no-pager 2>/dev/null || sudo -n journalctl "$@" --no-pager 2>/dev/null; }
JL -u wire-pod --since "@$((ST-120))" --until "@$((ST+5))" \
  | grep -iE 'stopping|stopped|started|main process exited|code=|signal|panic|fatal error|out of memory|oom|killed' | tail -8 | sed 's/^/     /'
JL -k --since "@$((ST-300))" --until "@$((ST+5))" | grep -iE 'out of memory|oom-kill|killed process' | tail -3 | sed 's/^/     /'
echo "     (\"Stopping\" right before = a manual restart; panic / code=killed / out of memory = a crash)"

echo "== 6. Log since that start"
J=$(JL -u wire-pod --since "@$ST")
echo "     panics: $(grep -c 'panic' <<<"$J"), searching-note problems: $(grep -c GOPOD_SEARCHING_GOLDEN <<<"$J")"
grep 'panic' <<<"$J" | tail -3 | sed 's/^/     /'

echo "== RESULT: $F problem(s) found, $FIXED fixed"
[ "$F" = "$FIXED" ] && echo "Next: F5 the chat page, then system-comm-check" || echo "Run me again after the fixes you skipped."

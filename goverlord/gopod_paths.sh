#!/bin/bash
# goverlord/gopod_paths.sh - make a fresh GOPOD clone point at YOUR machine.
#
# GOPOD's files carry its author's paths (/home/goverlord/...). This swaps them
# for yours in one go:
#   /home/goverlord/crushn8r_git/GOPOD  ->  wherever you cloned GOPOD
#   /home/goverlord                     ->  your home folder
#
# Usage:
#   bash goverlord/gopod_paths.sh            look only: what would change, and
#                                            which outside paths you still need
#   bash goverlord/gopod_paths.sh --apply    make the change
#   --home /home/you                         use another home than $HOME
#                                            (e.g. when running with sudo)
#
# Docs (*.md) keep the author's paths as examples; the one .md GOPOD reads at
# runtime (under wire_pod_overlay/chipper/) is changed with the code.
# Run it once per clone. Your changed files are yours: don't push them back.

set -euo pipefail

OLD_HOME="/home/goverlord"
OLD_REPO="/home/goverlord/crushn8r_git/GOPOD"
REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NEW_HOME="${HOME}"
APPLY=0

while [ $# -gt 0 ]; do
  case "$1" in
    --apply) APPLY=1 ;;
    --home) NEW_HOME="${2:?--home needs a folder}"; shift ;;
    -h|--help) sed -n '2,20p' "${BASH_SOURCE[0]}"; exit 0 ;;
    *) echo "gopod_paths: unknown option $1 (try --help)" >&2; exit 2 ;;
  esac
  shift
done
NEW_HOME="${NEW_HOME%/}"

if [ "$NEW_HOME" = "$OLD_HOME" ] && [ "$REPO" = "$OLD_REPO" ]; then
  echo "gopod_paths: this is the author's own machine layout - nothing to change."
  exit 0
fi

cd "$REPO"

# The files to change: text files holding the author's home, minus the docs.
mapfile -t FILES < <(
  grep -rIl --exclude-dir=.git -F "$OLD_HOME" . \
    | sed 's|^\./||' \
    | grep -v -E '\.md$' \
    | grep -v -x -F "goverlord/gopod_paths.sh" || true
  grep -rIl --exclude-dir=.git -F "$OLD_HOME" goverlord/wire_pod_overlay/chipper --include='*.md' || true
)

if [ "${#FILES[@]}" -eq 0 ]; then
  echo "gopod_paths: no author paths left - already done."
  exit 0
fi

# sed needs the new paths with \ | & escaped.
esc() { printf '%s' "$1" | sed -e 's/[\\|&]/\\&/g'; }
SED_ARGS=(-e "s|${OLD_REPO}|$(esc "$REPO")|g" -e "s|${OLD_HOME}|$(esc "$NEW_HOME")|g")

echo "GOPOD folder : $REPO"
echo "Your home    : $NEW_HOME"
echo
echo "Files with the author's paths (${#FILES[@]}):"
for f in "${FILES[@]}"; do
  printf '  %4s  %s\n' "$(grep -cF "$OLD_HOME" "$f")" "$f"
done
echo

# The paths outside the GOPOD folder: things you install or make yourself.
echo "Paths outside the GOPOD folder, as they will read for you:"
grep -oh -E "${OLD_HOME}/[A-Za-z0-9._@+/-]+" "${FILES[@]}" \
  | grep -v -F "$OLD_REPO" \
  | sed "${SED_ARGS[@]}" \
  | sed -E 's|[/.]+$||' \
  | while read -r p; do
      # Wire-Pod's own tree: one line for the folder (the overlay's file names
      # in there are new files on purpose - they never exist on disk).
      case "$p" in "$NEW_HOME"/wire-pod/*) echo "$NEW_HOME/wire-pod" ;; *) echo "$p" ;; esac
    done \
  | sort -u \
  | while read -r p; do
      if [ -e "$p" ]; then echo "  ok       $p"; else echo "  MISSING  $p"; fi
    done
echo "  (MISSING is normal on a fresh machine. GOPOD_YOURSELF.md says which ones"
echo "   you need for what, and which are optional.)"
echo

if [ "$APPLY" -ne 1 ]; then
  echo "Nothing changed yet. To make the change:  bash goverlord/gopod_paths.sh --apply"
  exit 0
fi

for f in "${FILES[@]}"; do
  sed -i "${SED_ARGS[@]}" "$f"
done

LEFT_PAT="$OLD_HOME"
case "$NEW_HOME" in "$OLD_HOME"*) LEFT_PAT="$OLD_REPO" ;; esac  # your home starts like the author's
LEFT="$(grep -rIl --exclude-dir=.git -F "$LEFT_PAT" . | grep -v -E '\.md$' | grep -v -F "gopod_paths.sh" || true)"
echo "GOPOD_PATHS_DONE ${#FILES[@]} files now point at $REPO and $NEW_HOME"
if [ -n "$LEFT" ]; then
  echo "Still holding the author's paths (check by hand):"
  echo "$LEFT" | sed 's/^/  /'
fi
echo "Next: GOPOD_YOURSELF.md, step 4."

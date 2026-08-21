#!/usr/bin/env bash
set -euo pipefail

# -----------------------------
# Config
# -----------------------------

# Default recency window for the code dump (days). Override by exporting
# ACTIVE_DAYS before running, e.g. ACTIVE_DAYS=7 ./goshot.sh
ACTIVE_DAYS="${ACTIVE_DAYS:-30}"

# Optional allowlist override. If non-empty, the code dump walks only these
# directories (full contents) instead of the default "files touched in the
# last ${ACTIVE_DAYS} days" git-recency scoping. All other filters
# (PRUNE_DIRS, SKIP_FILE_REGEX, SKIP_SECRET_REGEX, size cap, binary check)
# still apply on top of this. Usage:
#   INCLUDE_DIRS=("goverlord/runtime/songs/02_brobots_interview_run/zmisc" "goverlord/runtime/data_gomad")
INCLUDE_DIRS=()

STAMP="$(date '+%Y-%m-%d_%H-%M')"
OUTDIR="snapshots/${STAMP}"
TREE_OUT="${OUTDIR}/GOPOD_TREE_${STAMP}.txt"
CODE_OUT="${OUTDIR}/GOPOD_FULL_CODE_DUMP_${STAMP}.txt"
SKIP_LOG="/tmp/goshot_skipped.txt"
MAX_FILE_SIZE_BYTES=102400

mkdir -p "$OUTDIR"
: > "$SKIP_LOG"

echo "SNAPSHOT: $STAMP"
echo "OUTPUT: $OUTDIR"

# -----------------------------
# Ignore rules
# -----------------------------

PRUNE_DIRS=(
  ".git"
  ".venv"
  "venv"
  "env"
  "__pycache__"
  ".pytest_cache"
  ".mypy_cache"
  ".ruff_cache"
  "node_modules"
  "dist"
  "build"
  ".tox"
  ".cache"
  "htmlcov"
  ".idea"
  ".vscode"
  "old_runtime_absorption_001"
  "working_guts_ingestion_001"
  "snapshots"
)

SKIP_FILE_REGEX='(\.pyc$|\.pyo$|\.zip$|\.tar$|\.tar\.gz$|\.tgz$|\.tar\.xz$|\.7z$|\.rar$|\.mp4$|\.mov$|\.avi$|\.mkv$|\.mp3$|\.wav$|\.flac$|\.ogg$|\.png$|\.jpg$|\.jpeg$|\.gif$|\.webp$|\.ico$|\.pdf$|\.sqlite$|\.db$|\.log$|\.lock$|\.bak$|\.tmp$|\.swp$|\.DS_Store$|\.csv$|\.tsv$|(^|/)migrations/|(^|/)data/)'

SKIP_SECRET_REGEX='(^|/)(\.env|\.env\..*|.*secret.*|.*password.*|.*credential.*|.*private.*key.*|id_rsa|id_ed25519)$'

build_prune_expr() {
  local expr=""
  for d in "${PRUNE_DIRS[@]}"; do
    expr="${expr} -name ${d} -o"
  done
  expr="${expr% -o}"
  echo "$expr"
}

# Returns success if the given relative path lies under one of PRUNE_DIRS.
is_pruned_path() {
  local path="$1"
  local d
  for d in "${PRUNE_DIRS[@]}"; do
    case "/$path/" in
      */"$d"/*) return 0 ;;
    esac
  done
  return 1
}

# -----------------------------
# Tree snapshot
# -----------------------------

{
  echo "GOPOD TREE SNAPSHOT"
  echo "created_at: $(date -Is)"
  echo "repo_root: $(pwd)"
  echo
  echo "Excluded dirs:"
  printf -- "- %s\n" "${PRUNE_DIRS[@]}"
  echo
  echo "Excluded file patterns:"
  echo "$SKIP_FILE_REGEX"
  echo
  echo "========================================"
  echo
} > "$TREE_OUT"

if command -v tree >/dev/null 2>&1; then
  tree -a \
    -I '.git|.venv|venv|env|__pycache__|node_modules|dist|build|.pytest_cache|.mypy_cache|.ruff_cache|.cache|htmlcov|*.pyc|*.zip|*.tar|*.tar.gz|*.tgz|*.tar.xz|*.7z|*.rar|*.mp4|*.mov|*.avi|*.mkv|*.mp3|*.wav|*.flac|*.ogg|*.png|*.jpg|*.jpeg|*.gif|*.webp|*.ico|*.pdf|*.sqlite|*.db|*.log' \
    . >> "$TREE_OUT"
else
  find . \
    \( -name .git -o -name .venv -o -name venv -o -name env -o -name __pycache__ -o -name node_modules -o -name dist -o -name build \) -prune \
    -o -print \
    | sed 's#^\./##' \
    | sort >> "$TREE_OUT"
fi

# -----------------------------
# Full repo code/text dump
# -----------------------------

{
  echo "GOPOD FULL CODE/TEXT DUMP"
  echo "created_at: $(date -Is)"
  echo "repo_root: $(pwd)"
  echo
  echo "Purpose: Snapshot current GOPOD repo source/text files into one reviewable dump."
  if [ ${#INCLUDE_DIRS[@]} -gt 0 ]; then
    echo "Scope: INCLUDE_DIRS override active - only walking: ${INCLUDE_DIRS[*]}"
  else
    echo "Scope: files touched in the last ${ACTIVE_DAYS} days (git log + git diff), repo-wide."
  fi
  echo "Skipped: venvs, git, caches, pyc, zips/archives, media, DBs, logs, csv/tsv, migrations/ and data/ contents, likely secrets."
  echo "Per-file size cap: ${MAX_FILE_SIZE_BYTES} bytes. Large-file skips logged to: $SKIP_LOG"
  echo
  echo "========================================"
} > "$CODE_OUT"

if [ ${#INCLUDE_DIRS[@]} -gt 0 ]; then
  file_list="$(find "${INCLUDE_DIRS[@]}" -type f -print 2>/dev/null | sed 's#^\./##' | sort -u)"
else
  if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    echo "WARNING: pwd is not inside a git working tree ($(pwd)) - git-recency scoping finds 0 files." >&2
    echo "WARNING: run goshot.sh from inside the git repo root, or set INCLUDE_DIRS to bypass git scoping." >&2
    {
      echo
      echo "WARNING: pwd was not inside a git working tree at run time ($(pwd))."
      echo "WARNING: git-recency scoping found 0 files as a result. Re-run from inside the repo, or set INCLUDE_DIRS."
    } >> "$CODE_OUT"
    file_list=""
  else
    file_list="$(
      {
        git log --since="${ACTIVE_DAYS}.days.ago" --name-only --pretty=format: 2>/dev/null || true
        git diff --name-only HEAD 2>/dev/null || true
      } | sed '/^$/d' | sort -u
    )"
  fi
fi

included_count=0
skipped_large_count=0

while IFS= read -r file; do
  [ -z "$file" ] && continue
  [ -f "$file" ] || continue
  is_pruned_path "$file" && continue
  [[ "$file" =~ $SKIP_FILE_REGEX ]] && continue
  [[ "$file" =~ $SKIP_SECRET_REGEX ]] && continue

  # Skip files over the size cap.
  size="$(stat -c%s "$file" 2>/dev/null || echo 0)"
  if [ "$size" -gt "$MAX_FILE_SIZE_BYTES" ]; then
    skipped_large_count=$((skipped_large_count + 1))
    echo "SKIPPED_LARGE: $file (${size} bytes)" >> "$SKIP_LOG"
    {
      echo
      echo "========================================"
      echo "FILE SKIPPED LARGE: $file"
      echo "SIZE_BYTES: $size"
      echo "========================================"
    } >> "$CODE_OUT"
    continue
  fi

  # Only include text-like files.
  if file --mime "$file" | grep -Eq 'charset=binary'; then
    continue
  fi

  included_count=$((included_count + 1))
  {
    echo
    echo "========================================"
    echo "FILE: $file"
    echo "SIZE_BYTES: $size"
    echo "========================================"
    sed -e 's/\r$//' "$file" || true
  } >> "$CODE_OUT"
done < <(printf '%s\n' "$file_list")

echo
echo "TREE SNAPSHOT:"
echo "$TREE_OUT"
echo
echo "FULL CODE DUMP:"
echo "$CODE_OUT"
echo

dump_size="$(stat -c%s "$CODE_OUT" 2>/dev/null || echo 0)"
echo "SUMMARY:"
echo "  files included:        $included_count"
echo "  files skipped (large): $skipped_large_count"
echo "  code dump size (bytes): $dump_size"
echo
echo "DONE"

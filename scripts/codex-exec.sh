#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

TASK=""
TARGET_PATH=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --task)
            TASK="${2:-}"
            shift 2
            ;;
        --path)
            TARGET_PATH="${2:-}"
            shift 2
            ;;
        *)
            echo "Unknown argument: $1" >&2
            exit 2
            ;;
    esac
done

if [[ -z "$TASK" ]]; then
    echo "Usage: $0 --task <task> [--path <path>]" >&2
    exit 2
fi

cd "$REPO_ROOT"

if [[ -n "$TARGET_PATH" ]]; then
    codex exec "$TASK in path: $TARGET_PATH"
else
    codex exec "$TASK"
fi

# Real GOPOD_WIREPOD_BASE_URL value lives in a private, untracked local
# file - never a literal IP here (2026-08-15,
# GOPOLISHER_CUBE_SESSION_SWEEP_001.md).
GOPOD_LOCAL_CONFIG_PATH="${GOPOD_LOCAL_CONFIG_PATH:-$HOME/.gopod_alias_lib/local_config.sh}"
[ -f "$GOPOD_LOCAL_CONFIG_PATH" ] && . "$GOPOD_LOCAL_CONFIG_PATH"

alias wirepodlogs='curl -s "$GOPOD_WIREPOD_BASE_URL/api/get_logs"'
alias wirepoddebuglogs='curl -s "$GOPOD_WIREPOD_BASE_URL/api/get_debug_logs"'
alias wirepodlogsave='mkdir -p ~/gopod_wirepod_logs && curl -s "$GOPOD_WIREPOD_BASE_URL/api/get_logs" > ~/gopod_wirepod_logs/wirepod_logs_$(date +%Y%m%d_%H%M%S).txt && curl -s "$GOPOD_WIREPOD_BASE_URL/api/get_debug_logs" > ~/gopod_wirepod_logs/wirepod_debug_logs_$(date +%Y%m%d_%H%M%S).txt'

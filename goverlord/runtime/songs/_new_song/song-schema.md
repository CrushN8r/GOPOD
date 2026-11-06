# Song schema
- **about:** 2026-11-06 (operator: "build schemas for the other places in GOPOD where a schema would earn its keep"): the master song - every song folder in `goverlord/runtime/songs/` (not `_new_song/`, `tools/`) has these. A song still being written keeps only `_comment` in its knobs.json (WIP) and is skipped. The system check (`gopodSchemaCheck`, every restart) lists a gap "to fix".

---

## Required files
1. knobs.json
2. story.md

## Required in knobs.json
1. song_id
2. song_type
3. steps

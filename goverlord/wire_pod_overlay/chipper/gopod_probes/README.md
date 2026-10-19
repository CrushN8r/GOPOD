# gopod_probes

GOPOD's content and tools for the Brobots chat. Wire-Pod reads them from here.

- `chapters/` — the interview's topics: one folder per chapter, one file per topic (its
  trigger words, facts, pictures and value points). The format is at the top of
  `gopod_chapters_folder.go`.
- `foundation_packs/gopod_core.txt` — GOPOD's facts, in every Brobot prompt.
- `interview_harvest/harvest_questions.json` — the questions a dry run asks.
- `section_packets/` — Template 1 (the pronunciation list) and the Section 1 card the song
  engine reads.
- `stage_directions.json`, `system_speech_content.json`, `actionParameter-failures-list.txt`
  — speech rules the Go code reads.
- `tools/` — the room mic service, the reporter voice worker, the draft and take-sheet
  renderers, the voice-command and song-stop scripts, and small direct-SDK helpers
  (`direct_sdk_*.go`, built with `build_direct_sdk_tool.sh`).

Run output (`demo_runs/`) and compiled binaries are not tracked. No secrets
live here.

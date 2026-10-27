# gopod_probes

GOPOD's content and tools for the Brobots chat. Wire-Pod reads them from here.

- `chapters/` — the interview's topics: one folder per chapter, one file per topic (its
  trigger words, facts, pictures and value points). The format is at the top of
  `gopod_chapters_folder.go`. `chapters/_always/` holds topics outside the chapters: the
  surprise eggs, and a song's own value points (`05_nap.txt`, `06_bingo_test.txt`).
- `_topic_cluster/` — the template for a new topic cluster (a chapter): copy it into
  `chapters/`. Its README explains every line.
- `foundation_packs/gopod_core.txt` — GOPOD's facts, in every Brobot prompt.
- `interview_harvest/harvest_questions.json` — the questions a dry run asks.
- `section_packets/` — Template 1 (the pronunciation list) and the Section 1 card the song
  engine reads.
- `stage_directions.json`, `system_speech_content.json`, `actionParameter-failures-list.txt`
  — speech rules the Go code reads.
- `locks/persona_soft_lock.json` — the soft-locked personas (the Brobots' and the Gomads'),
  flavour lines and temperatures; a change needs the operator's word, then `GOPOD_RELOCK=1`.
- `tools/` — the room mic service, the reporter voice worker (Kokoro, with a host's pitch via
  sox), the page replay check, the draft and take-sheet
  renderers, the voice-command and song-stop scripts, and small direct-SDK helpers
  (`direct_sdk_*.go`, built with `build_direct_sdk_tool.sh`).

Run output (`demo_runs/`) and compiled binaries are not tracked. No secrets
live here.

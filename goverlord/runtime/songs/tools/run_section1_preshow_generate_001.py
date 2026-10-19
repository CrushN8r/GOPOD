#!/usr/bin/env python3
"""Standalone pre-show stage - the middle of the three-stage split
(gopod-opening-chord -> this -> the actual interview).

Runs only the Kokoro-narrated half of a run: warm-up, scaffold/section-card
load, Wire-Pod ESN routing announce, then the LLM generating every exchange
line - written to a JSON log on disk - and stops there. No robot speaks an
interview line in this script; that's the next stage's job.

Reuses run_section1_full_live_001.py's own generate_phase() unchanged (same
importlib pattern gopod-opening-chord already uses to reuse that file's
warm-up helpers) - no interview logic duplicated here.

Wired to the bash alias `interview-json` (renamed from `start-the-preshow`
2026-08-19, NAMING_APPLIED_001.md - same script, same behavior, new name).
To play the generated script back as the actual interview, run
`interview-replay` - it reads the newest file under demo_runs/, which this
script's own run just became; no need to pass the log path by hand unless
replaying something other than the most recent take:
  GOPOD_SECTION1_REPLAY_LOG=<log_path printed below> interview-replay
"""
import importlib.util
from pathlib import Path

RUNNER_PATH = Path(__file__).resolve().parent / "run_section1_full_live_001.py"


def _load_runner_module():
    spec = importlib.util.spec_from_file_location("run_section1_full_live_001", str(RUNNER_PATH))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def main():
    mod = _load_runner_module()
    log, _interview_scaffold, _robots = mod.generate_phase()
    timing_path = mod.write_timing_log(
        run_id=log.get("run_id"),
        song_dir=log.get("source_card"),
        meta={"tool": "run_section1_preshow_generate_001", "log_path": log.get("log_path")},
    )
    print(f"PRESHOW_GENERATE_DONE log_path={log['log_path']}")
    print(f"TIMING_LOG_WRITTEN path={timing_path}")
    return log["log_path"]


if __name__ == "__main__":
    main()

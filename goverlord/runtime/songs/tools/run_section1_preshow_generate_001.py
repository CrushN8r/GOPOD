#!/usr/bin/env python3
"""Standalone pre-show stage - the middle of the three-stage split (gopod-opening-chord -> this -> the actual interview)."""
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

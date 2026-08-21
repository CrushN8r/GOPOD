#!/usr/bin/env python3
"""Vamp-gate song — standalone, playhead-sliceable home for the pre-show's own
vamp_1..vamp_4 filler beats (ported into goverlord/runtime/songs/zzz_archives/brobots_vamp_gate/,
same content, same speakers). brobots_preshow's own vamp movement and the real live
vamp gate (run_preshow_song(), gated on interview generation finishing) are UNCHANGED
by this file - this is a second, independent home for the same four lines, built so
the vamp gate can be run/sliced/evaluated exactly like any other GOPOD song
(bingo/control/weather/bait), not just previewed inline (gopod-vamp) or heard as a
side effect of a full pre-show run.

Same importlib-reuse pattern as run_section1_preshow_generate_001.py and
run_robot_control_song_001.py: loads run_section1_full_live_001.py's own
load_preshow_song()/_preshow_speak_host() unchanged, no speech logic duplicated here.

Unlike the real live vamp gate (threading.Event, loops until generation finishes,
length is run-dependent), this runner plays a FIXED step list start to finish, once -
no generation thread, no indefinite loop. Deliberate: a song built for evaluation
needs a finite, repeatable run, not one whose length depends on how long an LLM call
happened to take.

No physical robot, no assume/release, no serial - vamp_1..vamp_4 are all
brobot_3/brobot_4 (Kokoro voice-only hosts), same as the source movement. The
playhead WAKE env var is accepted for pha0b's own uniform wiring but is a deliberate
no-op here - there is no assume_behavior_control call for a voice-only host to wake.
"""
import importlib.util
import json
import os
from pathlib import Path

RUNNER_PATH = Path(__file__).resolve().parent / "run_section1_full_live_001.py"
# Repointed 2026-07-24: the operator moved brobots_vamp_gate into zzz_archives/ during
# a manual song-folder cleanup - repointed, not retired, since this is the song's only
# alias (pha0b vamp).
# Env-overridable absolute anchor, not location-derived (2026-08-15,
# INTERVIEW_TOOLS_MOVE_EXECUTED_001.md) - same fix as run_golden_song_001.py's
# own SONGS_DIR, same reason: parents[2] silently broke the moment this
# file's folder moved up a level.
DEFAULT_SONG_DIR = Path(os.getenv("GOPOD_REPO_ROOT", "/home/goverlord/crushn8r_git/GOPOD")).expanduser() / "goverlord" / "runtime" / "songs" / "zzz_archives" / "brobots_vamp_gate"

VAMP_GATE_SONG_DIR_ENV = "GOPOD_VAMP_GATE_SONG_DIR"
VAMP_GATE_READ_SHEET_ENV = "GOPOD_VAMP_GATE_READ_SHEET"
VAMP_GATE_PLAYHEAD_FROM_ENV = "GOPOD_VAMP_GATE_PLAYHEAD_FROM"
VAMP_GATE_PLAYHEAD_TO_ENV = "GOPOD_VAMP_GATE_PLAYHEAD_TO"
VAMP_GATE_PLAYHEAD_WAKE_ENV = "GOPOD_VAMP_GATE_PLAYHEAD_WAKE"


def _load_runner_module():
    spec = importlib.util.spec_from_file_location("run_section1_full_live_001", str(RUNNER_PATH))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def run_vamp_gate_song(song_dir, read_sheet=False):
    mod = _load_runner_module()
    song = mod.load_preshow_song(song_dir)

    # Playhead range (GOPOD_VAMP_GATE_PLAYHEAD_FROM/_TO) - same slice-by-step_id
    # pattern as run_robot_control_song_001.py's own run_control_song(). Computed
    # against the full, unfiltered step list before anything else touches it.
    full_steps = song["steps"]
    playhead_from_id = os.getenv(VAMP_GATE_PLAYHEAD_FROM_ENV)
    playhead_to_id = os.getenv(VAMP_GATE_PLAYHEAD_TO_ENV)
    if playhead_from_id or playhead_to_id:
        all_step_ids = [s["step_id"] for s in full_steps]
        if playhead_from_id and playhead_from_id not in all_step_ids:
            raise RuntimeError(
                f"BLOCKED: {VAMP_GATE_PLAYHEAD_FROM_ENV}={playhead_from_id!r} is not a known step_id."
            )
        if playhead_to_id and playhead_to_id not in all_step_ids:
            raise RuntimeError(
                f"BLOCKED: {VAMP_GATE_PLAYHEAD_TO_ENV}={playhead_to_id!r} is not a known step_id."
            )
        from_index = all_step_ids.index(playhead_from_id) if playhead_from_id else 0
        to_index = all_step_ids.index(playhead_to_id) if playhead_to_id else len(all_step_ids) - 1
        if to_index < from_index:
            raise RuntimeError(
                f"BLOCKED: {VAMP_GATE_PLAYHEAD_TO_ENV}={playhead_to_id!r} comes before "
                f"{VAMP_GATE_PLAYHEAD_FROM_ENV}={playhead_from_id!r} in the score's own step order."
            )
        playhead_steps = full_steps[from_index : to_index + 1]
        print(
            f"VAMP_GATE_PLAYHEAD range={all_step_ids[from_index]}..{all_step_ids[to_index]} "
            f"steps={len(playhead_steps)}",
            flush=True,
        )
        song = {**song, "steps": playhead_steps}

    if os.getenv(VAMP_GATE_PLAYHEAD_WAKE_ENV) == "1":
        print(
            "VAMP_GATE_PLAYHEAD_WAKE no-op: voice-only hosts have no assume_behavior_control to wake.",
            flush=True,
        )

    log = {"song_id": song["song_id"], "read_sheet": read_sheet, "steps": []}
    for step in song["steps"]:
        result = mod._preshow_speak_host(
            step, read_sheet=read_sheet, segment_id=f"vamp_gate_standalone_{step['step_id']}"
        )
        log["steps"].append(
            {
                "step_id": step["step_id"],
                "speaker": step["speaker"],
                "text": step["text"],
                "spoken": result.get("spoken"),
            }
        )
    print(f"VAMP_GATE_DONE steps_played={len(log['steps'])}", flush=True)
    return log


def main():
    song_dir = os.getenv(VAMP_GATE_SONG_DIR_ENV, str(DEFAULT_SONG_DIR))
    read_sheet = os.getenv(VAMP_GATE_READ_SHEET_ENV, "0") == "1"
    log = run_vamp_gate_song(song_dir, read_sheet)
    print(json.dumps(log, indent=2))
    return log


if __name__ == "__main__":
    main()

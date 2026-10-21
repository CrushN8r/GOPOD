#!/usr/bin/env python3
"""Golden song runner — Phase 1 of the golden-song-runner unification path
(gopod_notes/older_notes/GOLDEN_SONG_RUNNER_PLAN_SURVEY_001.md,
~/.claude/plans/squishy-snacking-falcon.md). A single fat-superset note-dispatch
engine meant to eventually replace BOTH run_songs_runner_001.py (the "bingo"
runner) and run_robot_control_song_001.py (the "control song" runner) — but
this file never edits either of those two, or run_section1_full_live_001.py
(the shared interview library both already import from and this file imports
from too, exactly the same pattern). This pass proves the engine on ONE
bingo-family song only (see SONG_REGISTRY below) — widening the registry to
the rest of the bingo family, and adding the control-song-family
compatibility shim (note_aliases/synthesize_speaker/unconditional_text_speak),
are later, separate passes.

Behavior is a straight port of run_songs_runner_001.py's own note vocabulary
and stop-condition logic — this file changes WHICH literals are hardcoded
(BINGO_* -> song_id-derived, per the survey's own schema table), not what
happens when a note fires. Every hardware-tuned timing constant below
(RATTLE_SETTLE_SECONDS, DIRECT_SDK_RELEASE_SETTLE_SECONDS,
BROBOTS_READY_TOGETHER_HANDBACK_SETTLE_SECONDS) is copied verbatim from that
file's own live-confirmed values — do not retune here; retune (if ever
needed) in the one file both engines temporarily share history with.

Eleven note types, one dispatch loop, config supplied per-song by SONG_REGISTRY:
wake_both, say_turn, rattle, brobots_ready_together, arm_cue, nod, wheel_nudge,
emotion_beat, animation, pause, exit. wheel_nudge added 2026-08-09
(WHEEL_NUDGE_GOLDEN_PATH_SURVEY_001.md Stage 2) - fixed reverse only, no
song has a step using it yet, so this is purely additive: a song with no
wheel_nudge step is unaffected.

Naming: print-line tags and the saved run log/json filename stem are derived
from the song's own song_id (read from its knobs.json) rather than a fixed
family literal — e.g. "BROBOTS_BABY_ROBOTS_SLEEP_CAPTURE" instead of
"BINGO_CAPTURE". Env var names are NOT song-derived — they are this engine's
own fixed configuration surface (one process = one song, selected via
GOPOD_GOLDEN_SONG_DIR), the same category as the already-global
GOPOD_ALLOW_LIVE_ROBOT_SPEECH/GOPOD_VOICE_DESTINATION vars every runner reads.
"""
import importlib.util
import base64
import json
import os
import queue
import re
import shutil
import subprocess
import wave
import sys
import threading
import time
import urllib.request
from pathlib import Path

from knobs_envelope_001 import load_knobs_envelope
from song_folders_001 import find_song_dir

RUNNER_PATH = Path(__file__).resolve().parent / "run_section1_full_live_001.py"
CONTROL_SONG_RUNNER_PATH = Path(__file__).resolve().parent / "run_robot_control_song_001.py"
# Env-overridable absolute anchor, not location-derived (2026-08-15,
# INTERVIEW_TOOLS_MOVE_EXECUTED_001.md) - the old Path(__file__).resolve().
# parents[2] silently pointed at the wrong directory the moment this file's
# own folder moved up a level (interview/tools/ -> tools/), same failure
# mode run_section1_full_live_001.py's own header comment already named and
# fixed for its REPO_ROOT/WIREPOD_CHIPPER_ROOT - this file just hadn't been
# fixed yet. Same env var, same hardcoded default, matching that precedent.
SONGS_DIR = Path(os.getenv("GOPOD_REPO_ROOT", "/home/goverlord/crushn8r_git/GOPOD")).expanduser() / "goverlord" / "runtime" / "songs"

GOLDEN_SONG_DIR_ENV = "GOPOD_GOLDEN_SONG_DIR"
GOLDEN_SONG_ID_ENV = "GOPOD_GOLDEN_SONG_ID"
GOLDEN_SECTION_ENV = "GOPOD_GOLDEN_SECTION"
GOLDEN_PLAYHEAD_FROM_ENV = "GOPOD_GOLDEN_PLAYHEAD_FROM"
GOLDEN_PLAYHEAD_TO_ENV = "GOPOD_GOLDEN_PLAYHEAD_TO"
GOLDEN_PLAYHEAD_WAKE_ENV = "GOPOD_GOLDEN_PLAYHEAD_WAKE"
GOLDEN_REPORTER_GAP_OVERRIDE_ENV = "GOPOD_GOLDEN_REPORTER_GAP_OVERRIDE"
GOLDEN_CONNECTION_PASS_ENV = "GOPOD_GOLDEN_CONNECTION_PASS"
CONNECTION_PASS_TIMEOUT_SECONDS = 15

# Phase 2 only: which single robot a synthesize_speaker=True song targets -
# mirrors run_robot_control_song_001.py's own GOPOD_CONTROL_SONG_ROBOT
# exactly (values "1"/"2", default "1"), new name to stay inside this
# engine's own GOPOD_GOLDEN_* namespace. No-op for every synthesize_speaker=
# False song (bingo-family) - never read.
GOLDEN_ROBOT_ENV = "GOPOD_GOLDEN_ROBOT"

# Run-time robot playback filter (pha0b's "run which robots?" prompt, per-
# run only - never writes a knobs.json, distinct from robot_pick_001.py
# which permanently reassigns a step's own `speaker` field on disk). Values
# "1"/"2"/"both" (default "both" = today's byte-identical behavior, every
# step fires exactly as it always has). "1"/"2" skip dispatching any step
# whose resolved `speaker` is the OTHER robot - the step is logged filtered,
# not errored, and its own buffer_after/tempo timing still fires normally so
# the run's overall pacing doesn't shift just because one robot's cues went
# silent for this take.
GOLDEN_ROBOT_FILTER_ENV = "GOPOD_GOLDEN_ROBOT_FILTER"
_ROBOT_FILTER_SPEAKER = {"1": "brobot_1", "2": "brobot_2"}
# Notes exempt from the filter regardless of setting: wake_both/
# brobots_ready_together always touch both physical robots no matter which
# nominal `speaker` their own knobs.json step carries (run_wake_both() pings
# both serials unconditionally; run_brobots_ready_together() fires the
# together binary against both serials unconditionally) - filtering "the
# other robot's" wake or ready-together would silently break connectivity
# prep/the together beat for a note that was never actually single-robot to
# begin with. pause/exit touch no robot at all - filtering them would only
# ever discard real timing content for no reason, breaking the "buffers/
# pauses intact" requirement.
_FILTER_EXEMPT_NOTES = {"wake_both", "sleep_both", "all_asleep", "brobots_ready_together", "system_line", "pause", "timer", "exit"}

# Same bare-presence-is-not-proof discriminator as run_songs_runner_001.py -
# see that file's own comment / BINGO_STUCK_ANIMATION_FALSE_POSITIVE_RESOLUTION_001.md.
ANIMATION_WAIT_LOG_SIGNATURE = "waiting for animation to be done"

_LOOP_ANIMATION_TOKENS = {"searching", "answering"}
_ANIMATION_LOOP_INTERVAL_SECONDS = 0.333

# GOLDEN, LOCKED 2026-09-30 - how a loop animation plays and ENDS in a song
# (SEARCHING_ANIMATION_SONG_LOCK_001.md, gopod_notes). "searching" is sent ONCE,
# held for the step's hold_seconds, then closed with one "searchingGetout"
# (brobots-searching-out's / phcal's own getout) and 0.5s. Never re-sent every
# 0.333s in a song: one searching clip runs ~3.9s (go search V7,
# gesture_clip_ms), so 4 sends in 1.3s left ~16s of clips waiting in Wire-Pod's
# animation queue, and the next line's own "thinking" animation set them
# playing again - the "searching sound runs on until the robot is called
# again" leak (bingo live runs 2026-09-16, 17:43 2026-09-30). The phcal bench
# and test-anim-* aliases still loop on purpose: they fire one thing in
# isolation with nothing after it. "answering" has no getout clip: it keeps the
# loop and settles 1.0s (ANSWERING_NATIVE_STOP_DISCIPLINE_001.md). Locked by
# test_golden_song_animation_lock_001.py - change only with a fresh live test.
_ANIMATION_GETOUT_TOKENS = {"searching": "searchingGetout"}
_ANIMATION_GETOUT_SETTLE_SECONDS = 0.5
_ANIMATION_NO_GETOUT_SETTLE_SECONDS = 1.0

# Direct-SDK golden path constants (DECOUPLED_DIRECT_SDK_GOLDEN_PATH_001.md),
# copied verbatim from run_songs_runner_001.py - live-confirmed values, not
# retuned here. See that file's own comments for the full tuning history
# (RATTLE_SETTLE_SECONDS in particular: confirmed working live TWICE at 3.0s,
# do not lower without a fresh live re-confirmation).
DIRECT_SDK_RELEASE_SETTLE_SECONDS = 1.0
BROBOTS_READY_TOGETHER_HANDBACK_SETTLE_SECONDS = 1.0
RATTLE_SETTLE_SECONDS = 3.0

# 2026-08-18, PHCAL_DETECT_FIRST_001.md live-test finding, same failure
# class as BROBOTS_WAKE_POST_REASSUME_SETTLE_SECONDS's own 2026-08-10 fix
# (run_robot_control_song_001.py): an action fired immediately after a
# fresh assume_behavior_control response, before the robot's real async
# control grant had actually landed - HTTP 200 confirms accepted, not
# landed. run_emotion_beat()/run_animation_only() below both fired their
# first say_text call with zero settle after a fresh (manage_control=True)
# assume. Matches PHCAL_ASSUME_SETTLE_SECONDS's own 1.5s "cold-first-cycle"
# value (phcal_isolate_001.py) - same kind of cold/first assume, not a
# reassume mid-hold, so the 1.5s number is used, not the smaller 0.5s
# reassume-specific one.
POST_ASSUME_SAY_SETTLE_SECONDS = 1.5

# A spoken line back faster than this was not spoken (see run_say_turn's
# self-heal). Silent lines measured 0.03-0.18s; the shortest real line
# ("Wait!") takes about 1s.
SAY_SILENT_SECONDS = 0.5

RATTLE_BIN_PATH = Path("/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_probes/tools/direct_sdk_bingo_rattle_001")
RATTLE_WAV_PATH = Path(
    "/home/goverlord/crushn8r_git/SDK/sources/vectorx/vectorfs/data/audio/rattle.wav"
)
TOGETHER_BIN_PATH = Path("/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_probes/tools/direct_sdk_brobots_ready_001")
# The cube blip (phcal's cube test): all corners red, 2 s, then green.
CUBE_BLIP_BIN_PATH = Path("/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_probes/tools/direct_sdk_cube_blip_001")
WIREPOD_HOME_PATH = "/home/goverlord/wire-pod"

# 2026-10-09, the nap (operator: "Start this song by putting the brobots to
# sleep * wake in time to act, then back to sleep? * like sleep walking"; "Since
# brobots are sleepy their actions don't require specific timing restrictions *
# they can lull out of and back into sleep * have it so they're sleeping by the
# end of the song"): phcal's golden sleep tool (brobots_sleep_to_wake_direct_sdk,
# direct_sdk_robot_sleep_001 sleep --wait-signal), one per Brobot. Banked rule
# (SLEEP_GOLDEN_NOTES_001.md): never assume control through Wire-Pod while the
# sleep tool holds a Brobot - it hung the sleep tool - so a move wakes the
# Brobot (the signal file: GoToSleepOff in the same connection, then its
# release), settles, takes control, moves, hands back, settles, and sleeps again.
SLEEP_BIN_PATH = TOGETHER_BIN_PATH.parent / "direct_sdk_robot_sleep_001"
# The sleep tool's own safety net (it wakes the Brobot after this long with no
# signal): longer than any nap in a song.
SLEEPY_MAX_WAIT_SECONDS = 300.0
SLEEPY_MOVE_NOTES = ("arm_cue", "nod", "animation")

WIREPOD_PAIRING_RECOVERY_MESSAGE = (
    "sequence error - Wire-Pod could not reach the robot(s). Field-proven recovery "
    "order (robot_control_song_001/story.md's own Troubleshooting section, confirmed "
    "live - a partial version does not clear it): 1) power-cycle both robots, "
    "2) re-pair them with Wire-Pod, 3) restart Wire-Pod (wpr). Stopping here, not "
    "guessing a different fix."
)

# Per-note-type stop message, only consulted for a note in that song's own
# fatal_notes (see SONG_REGISTRY). A note added to fatal_notes later without
# an entry here just gets the generic fallback - never a KeyError.
_FATAL_STOP_MESSAGES = {
    "wake_both": lambda step_id: (
        f"wake_both {WIREPOD_PAIRING_RECOVERY_MESSAGE}"
    ),
    "emotion_beat": lambda step_id: (
        "emotion_beat failed - stopping run per stop condition, not retrying into a "
        "possibly-wedged robot. release_behavior_control was still attempted above. "
        "Earlier beats' results stand as recorded."
    ),
}


def _generic_fatal_message(note, step_id):
    return f"{note} failed - stopping run per {note}'s own fatal_notes entry."


# Per-song config for the golden engine, keyed by song_id (the "song_id" in the
# song's own knobs.json). fatal_notes stop a run when they fail; every other
# note warns and goes on. manage_control=False + synthesize_speaker=False is the
# continuous dual-robot hold (STAY_PUT_FIX_SURVEY_001.md): both Brobots held for
# the whole run, each step addressed by its own speaker.
# 2026-10-10 (operator: "even shuffling (song) folders without notice should be
# easy to remap"): no folder names here - a song is found by its song_id
# (song_folders_001.py). Only songs on disk are listed; the archived ones went
# with zzz_archives.zip.
SONG_REGISTRY = {
    "brobots_baby_robots_sleep": {
        "runner": "run_golden_song_001.py",
        "golden_ready": True,
        "fatal_notes": {"wake_both", "emotion_beat"},
        "manage_control": False,
        "note_aliases": {},
        "synthesize_speaker": False,
        "unconditional_text_speak": False,
    },
    "brobots_bingo": {
        "runner": "run_golden_song_001.py",
        "golden_ready": True,
        "fatal_notes": {"wake_both", "emotion_beat"},
        "manage_control": False,
        "note_aliases": {},
        "synthesize_speaker": False,
        "unconditional_text_speak": False,
    },
}

# 2026-10-11 (operator: "new song, really fast"): the elevator fart joke.
SONG_REGISTRY["brobots_fart_joke"] = dict(SONG_REGISTRY["brobots_bingo"])

DEFAULT_SONG_ID = "brobots_baby_robots_sleep"


# ---- display-tag derivation (generic, step_id/text-driven, not song-specific) ----

_NOTE_TYPE_TAGS = {
    "wake_both": "wake",
    "brobots_ready_together": "brobots together",
    "arm_cue": "arm cue",
    "nod": "nod",
    "rattle": "rattle",
    "exit": "exit",
}

_SAY_TURN_FILLER_PHRASES = ("big shiny bingo balls", "big shiny bingo ball")

_SAY_TURN_STOPWORDS = {
    "is", "are", "am", "do", "you", "your", "my", "what", "in", "all", "of",
    "we're", "we", "one", "before", "let's", "us", "to", "for", "on",
    "it's", "its", "know", "did", "does", "was", "were", "big", "shiny",
    "bingo", "ball", "balls", "the",
}

_NUMBER_ECHO_RE = re.compile(r"\b([a-zA-Z]-\d+)\b")

_TAG_OVERRIDES = {
    "host_arm_cue_0102": "arm cue 2x",
}


def _derive_say_turn_tag(step_id, text):
    stripped = (text or "").strip()
    if not stripped:
        return "line"
    number_match = _NUMBER_ECHO_RE.search(stripped)
    if number_match:
        number = number_match.group(1).lower()
        if "call" in step_id:
            return f"call {number}"
        if stripped.rstrip("!.? ").lower() == number:
            return number
    lowered = stripped.lower()
    for phrase in _SAY_TURN_FILLER_PHRASES:
        lowered = lowered.replace(phrase, " ")
    first_sentence = re.split(r"[.!?]", lowered, maxsplit=1)[0]
    words = [w.strip(",\"'") for w in first_sentence.split() if w.strip(",\"'")]
    if not words:
        return "line"
    content = [w for w in words if w not in _SAY_TURN_STOPWORDS]
    chosen = content if content else words
    if len(chosen) > 3:
        chosen = chosen[:2]
    return " ".join(chosen)


def _derive_note_tag(step):
    note = step.get("note")
    step_id = step.get("step_id", "")
    if step_id in _TAG_OVERRIDES:
        return _TAG_OVERRIDES[step_id]
    if note == "pause":
        if step_id.startswith("reporter_gap"):
            return "reporter gap"
        if step_id.startswith("transition_gap"):
            return "transition gap"
        return "pause"
    if note == "emotion_beat":
        token = step.get("animation_token") or "beat"
        return f"{token} beat"
    if note == "animation":
        token = step.get("animation_token") or "animation"
        return f"{token} animation"
    if note == "say_turn":
        return _derive_say_turn_tag(step_id, step.get("text", ""))
    return _NOTE_TYPE_TAGS.get(note, note or "step")


# ---- module loading / song loading ----


class _TeeStdout:
    """Same save-to-disk shim as run_songs_runner_001.py's own _TeeStdout -
    duplicates every write to real stdout and an open log file, so
    Robots.say()'s own prints (run_section1_full_live_001.py, not touched
    here) land in the saved score too."""

    def __init__(self, real_stdout, log_file):
        self._real_stdout = real_stdout
        self._log_file = log_file
        self.capture_file = True

    def write(self, data):
        self._real_stdout.write(data)
        if self.capture_file:
            self._log_file.write(data)

    def flush(self):
        self._real_stdout.flush()
        self._log_file.flush()


def _load_module(path, name):
    spec = importlib.util.spec_from_file_location(name, str(path))
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def _load_runner_module():
    return _load_module(RUNNER_PATH, "run_section1_full_live_001")


def _load_control_song_module():
    return _load_module(CONTROL_SONG_RUNNER_PATH, "run_robot_control_song_001")


def _sequence_steps(step, song_dir, control_mod, seen=None):
    """2026-10-16 (operator: "extend that to include sequences"): a sequence
    step plays another song's steps, in place - its "song_id" names the song
    (found by name, song_folders_001.py). A song can't sequence itself, and an
    LLM warm song (played by Wire-Pod) isn't played here; either is skipped, said."""
    seen = set(seen or ()) | {Path(song_dir).resolve()}
    target_id = (step.get("song_id") or "").strip()
    folder = find_song_dir(target_id) if target_id else None
    if folder is None:
        print(f"GOLDEN_SEQUENCE {step.get('step_id')}: no song with song_id {target_id!r} - skipped", flush=True)
        return []
    if folder.resolve() in seen:
        print(f"GOLDEN_SEQUENCE {step.get('step_id')}: {target_id} is already playing - skipped", flush=True)
        return []
    _, knobs = load_knobs_envelope(folder)
    if knobs.get("song_type") == "llm_warm":
        print(f"GOLDEN_SEQUENCE {step.get('step_id')}: {target_id} is an LLM warm song, played by Wire-Pod - skipped", flush=True)
        return []
    inner = load_golden_song(folder, control_mod, seen)["steps"]
    for s in inner:
        s.setdefault("section", step.get("section", ""))
    return inner


def combo_parts(step):
    """2026-10-10 (operator: "see new steps with new combos, _nod, _beat_angry,
    _kg_success ... I combined steps that had Lol! at the end"): a combo step is
    one story.md step played as its "combo" notes, in order. A part with "text"
    says those words, cut from the step's TEXT ("Lol!", "Excuse Me?"); a part
    without "text" says what's left; "text": "" says nothing. Each part keeps
    the step's step_id, speaker and section. Each part's line is said to its
    end before the next part starts (say_text answers once it's spoken)."""
    rest = step.get("text", "")
    parts = []
    for raw in step["combo"]:
        part = {k: v for k, v in step.items() if k not in ("combo", "note", "text", "fail")}
        part.update(raw)
        if "text" in raw:
            words = raw["text"].strip()
            if words and words in rest:
                rest = rest.replace(words, "", 1)
            part["text"] = words
        parts.append(part)
    rest = re.sub(r"\s+", " ", rest).strip()
    for i, part in enumerate(parts):
        if "text" not in part:
            part["text"] = rest
        part["fail"] = step.get("fail", "")
        part["combo_texts"] = [p["text"] if "text" in p else rest for p in parts]
        part["combo_part"] = i + 1
        part["combo_last"] = i == len(parts) - 1
    return parts


def _llm_song_lines(song_dir, control_mod, fallback):
    """The song's _llm_song run (_llm_song/knobs.json + story.md, note "line",
    speaker a reporter host) as warm-up lines [{"who", "text"}]; with no such
    folder, fallback. An unreadable run says no lines (it never stops the song)."""
    run = Path(song_dir) / "_llm_song"
    if not (run / "knobs.json").is_file():
        return fallback
    try:
        knobs = json.loads((run / "knobs.json").read_text(encoding="utf-8"))
        texts = control_mod.parse_control_story_md((run / "story.md").read_text(encoding="utf-8")) if (run / "story.md").is_file() else {}
    except Exception as exc:  # noqa: BLE001 - a bad run never stops the song
        print(f"LLM_SONG {run}: unreadable ({type(exc).__name__}) - no lines", flush=True)
        return []
    lines = []
    for step in knobs.get("steps", []):
        text = texts.get(step.get("step_id", "")) if isinstance(texts, dict) else None
        if isinstance(text, dict):
            text = text.get("text", "")
        if step.get("note") == "line" and step.get("speaker") and text:
            lines.append({"who": step["speaker"], "text": str(text).strip()})
    return lines


def load_golden_song(song_dir, control_mod, _seen=None):
    """Same parsing shape as run_songs_runner_001.py's own
    load_bingo_capture_song(): reuses control_mod.parse_control_story_md() for
    STEP-heading parsing, merges every knobs.json field through (no
    whitelist, so animation_token/hold_seconds/
    gap_label all survive), plus story.md's own text/fail. Kept as this
    file's own copy (not imported from run_songs_runner_001.py) so the golden
    engine has zero runtime dependency on the file it is meant to replace.

    Ladder 2, Phase 4 (KNOBS_ROOT_UNTANGLE_SURVEY_001.md): file resolution +
    envelope load now come from the shared knobs_envelope_001.py
    (resolve_active_knobs_path/load_knobs_envelope) - the same piece
    print_song_score_001.py and brobots.sh's panel readers already call.
    Only the resolve+load changed; everything below (story merge, step
    shape) is untouched."""
    song_dir = Path(song_dir)
    knobs_path, knobs = load_knobs_envelope(song_dir)
    story = control_mod.parse_control_story_md((song_dir / "story.md").read_text(encoding="utf-8"))
    steps = []
    for raw_step in knobs.get("steps", []):
        step_id = raw_step.get("step_id")
        content = story.get(step_id, {"text": "", "fail": ""})
        merged = dict(raw_step)
        merged["text"] = content["text"]
        merged["fail"] = content["fail"]
        # 2026-10-09 (the nap, a lyric video): "robots_speak": false - story.md's
        # words are the narrator's captions, never said by a robot.
        # A step with no start_seconds (the ending's "Brobots Ready!") keeps its words.
        if knobs.get("robots_speak") is False and raw_step.get("start_seconds") is not None:
            merged["caption"] = merged["text"]
            merged["text"] = ""
        if merged.get("combo"):
            steps.extend(combo_parts(merged))
            continue
        if merged.get("note") == "sequence":
            steps.extend(_sequence_steps(merged, song_dir, control_mod, _seen))
            continue
        steps.append(merged)
    return {
        "song_id": knobs.get("song_id", ""),
        "global_tempo": knobs.get("global_tempo", 0.0),
        "chat_intro": knobs.get("chat_intro", ""),
        "sentence_split": knobs.get("sentence_split"),
        "rattle_on": knobs.get("rattle_on", "robot"),
        "touch_interrupt_off": bool(knobs.get("touch_interrupt_off")),
        # 2026-10-12 (operator: "all songs must have this same golden templated
        # process" - the opener's): warming, any warmup_lines, then the ready line;
        # every song unless its knobs.json says "llm_warmup": false.
        "llm_warmup": bool(knobs.get("llm_warmup", True)),
        "llm_ready_text": knobs.get("llm_ready_text", ""),
        # 2026-10-20 (operator: "all llm warmings load an expected run"): the
        # song's _llm_song run when it has one, else its knobs' warmup_lines.
        "warmup_lines": _llm_song_lines(song_dir, control_mod, knobs.get("warmup_lines", [])),
        # 2026-10-17: a song needing no voice files made / no language model.
        "voices_required": knobs.get("voices_required", True) is not False,
        "llm_required": knobs.get("llm_required", True) is not False,
        # 2026-10-12 (operator: "LLM warm-up start after “Go to sleep now please”"):
        # the warm-up runs right after this step instead of at the start.
        "warmup_after": knobs.get("warmup_after", ""),
        "timeline": knobs.get("timeline", ""),
        "audio": knobs.get("audio", ""),
        "sleepy": bool(knobs.get("sleepy")),
        # 2026-10-15: the sleepy Brobots' wake clip (SleepyBrobots), and a picture
        # for each narrator bubble, by section (_caption_bubbles).
        "wake_clip": knobs.get("wake_clip", ""),
        "section_pictures": knobs.get("section_pictures", {}),
        "narrator_single": bool(knobs.get("narrator_single")),
        # 2026-10-18 (forced alignment): the narrator's word times, by caption
        # (narrator_words.json, gopod_align_captions_001.py), when the song has them.
        "narrator_words": _narrator_words(song_dir),
        "steps": steps,
    }


def _narrator_words(song_dir):
    """The song's narrator_words.json captions ({start: [{t, s, e}]}), or {}."""
    try:
        return json.loads((Path(song_dir) / "narrator_words.json").read_text(encoding="utf-8")).get("captions", {})
    except (OSError, ValueError):
        return {}


def _serial_and_name(mod, speaker):
    if speaker == "brobot_2":
        return mod.BROBOT_2_SERIAL, "Brobot 2"
    return mod.BROBOT_1_SERIAL, "Brobot 1"


def _release_if_holding(mod, serials, live, manage_control):
    """manage_control=False songs assume control once (awaken's "connect" note,
    or the stay-put fix's own dual-robot assume off wake_both) and hold it
    continuously - Robots.close() never releases it (it's a no-op), so this
    must fire explicitly wherever the run ends, matching
    run_robot_control_song_001.py's own always-attempt `finally` release
    exactly. No-op for manage_control=True songs regardless of `serials`.

    Generalized from a single optional serial to an iterable
    (STAY_PUT_FIX_SURVEY_001.md, part 1) - awaken's own single-serial call site
    passes a one-element tuple; the stay-put fix's own bingo-family call site
    passes both robot serials. Empty/None `serials` is a clean no-op either
    way (nothing was ever synthesized/assumed to release)."""
    if not (live and not manage_control):
        return
    for serial in serials or ():
        if serial:
            mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)


def run_continuous_hold_assume(mod, live):
    """Stay-put fix (STAY_PUT_FIX_SURVEY_001.md, parts 2 and 4): assumes
    OVERRIDE_BEHAVIORS on BOTH robot serials at once. Two call sites use this -
    the initial dual-robot grab fired once, right after wake_both's own
    conn_test succeeds, and the re-assume fired after rattle/
    brobots_ready_together return, restoring the hold those two notes
    correctly (and necessarily) dropped for their own direct-SDK handoff."""
    if not live:
        return {"ok": True, "mode": "dry"}
    results = {}
    ok = True
    for label, serial in (("brobot_1", mod.BROBOT_1_SERIAL), ("brobot_2", mod.BROBOT_2_SERIAL)):
        result = mod.wirepod_web_send_form(
            "/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10
        )
        results[label] = result
        ok = ok and result.get("status") == 200
    # 2026-09-30: Wire-Pod answers the assume before the robot's control grant
    # lands (bcassume.go); an arm/nod fired straight after it silently did
    # nothing (bingo live run 2026-09-30 13:59, steps 5 and 10, right after
    # "Brobots ready" and the rattle, both of which hand control over fully).
    # Same 1.5s cold-assume settle as POST_ASSUME_SAY_SETTLE_SECONDS/phcal.
    time.sleep(POST_ASSUME_SAY_SETTLE_SECONDS)
    return {"ok": ok, "results": results, "settle_seconds": POST_ASSUME_SAY_SETTLE_SECONDS}


# 2026-10-01, operator: "take back a robot's control when it goes silent after a
# crash". A robot that crashes and reboots mid-song loses the song's held
# control, and stays silent for the rest of the run (Brobot 2, live 14:10: its
# line hung 90s, it was back at 14:12, and its later lines came back "success"
# but unheard - too slow for the silent-line heal in run_say_turn). So under
# the song's hold, a robot step that fails (a line that times out, an animation
# that errors) or a line still silent after the heal marks that robot; before
# its next step the song waits for Wire-Pod to reach it again (conn_test) and
# then retakes its control the heal's own way - release, settle, retake, settle.
CONTROL_RETAKE_NOTES = ("say_turn", "animation", "emotion_beat", "arm_cue", "nod")
CONTROL_RETAKE_WAIT_SECONDS = 30.0
CONTROL_RETAKE_POLL_SECONDS = 2.0


def retake_robot_control(mod, serial, name, live, robots, tag):
    """Wait (up to CONTROL_RETAKE_WAIT_SECONDS) until Wire-Pod reaches the
    robot, then release and retake its control. Never stops a run: a robot
    still out is printed and the song carries on."""
    if not live:
        return {"ok": True, "mode": "dry"}
    started = time.monotonic()
    reached = False
    while True:
        try:
            conn = mod.wirepod_web_send_form("/api-sdk/conn_test", {"serial": serial}, timeout=10)
            reached = conn.get("status") == 200 and conn.get("body", "").strip() == "success"
        except Exception:  # noqa: BLE001 - a robot still rebooting is what this waits out
            reached = False
        if reached or time.monotonic() - started >= CONTROL_RETAKE_WAIT_SECONDS:
            break
        time.sleep(CONTROL_RETAKE_POLL_SECONDS)
    waited = round(time.monotonic() - started, 1)
    if not reached:
        print(f"{robots.timing_prefix()}{tag}_CONTROL_RETAKE {name}: not reachable after {waited}s - carrying on", flush=True)
        return {"ok": False, "reached": False, "waited_seconds": waited}
    try:
        release = mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)
        time.sleep(DIRECT_SDK_RELEASE_SETTLE_SECONDS)
        assume = mod.wirepod_web_send_form("/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10)
        time.sleep(POST_ASSUME_SAY_SETTLE_SECONDS)
    except Exception as exc:  # noqa: BLE001 - never stops the song
        print(f"{robots.timing_prefix()}{tag}_CONTROL_RETAKE {name}: failed ({exc!r}) - carrying on", flush=True)
        return {"ok": False, "reached": True, "waited_seconds": waited, "error": repr(exc)}
    ok = assume.get("status") == 200
    print(f"{robots.timing_prefix()}{tag}_CONTROL_RETAKE {name}: reached after {waited}s, control retaken ok={ok}", flush=True)
    return {"ok": ok, "reached": True, "waited_seconds": waited, "release": release, "assume": assume}


class SleepyBrobots:
    """knobs.json "sleepy": true. sleep_both puts both Brobots to sleep; from
    then on a move step is handed to its Brobot's own worker, which lulls out
    (wakes), makes every move that's due, and drifts back to sleep - no timing
    limits, so the music and captions never wait on it. finish() waits for the
    moves; wake_all() wakes both (the ending's "Brobots Ready!")."""

    def __init__(self, mod, live, robots, tag, wake_clip=""):
        self.mod, self.live, self.robots, self.tag = mod, live, robots, tag
        # 2026-10-15 (operator: "switch to slower clip"): knobs.json "wake_clip" -
        # the Brobots wake with this clip (the sleep tool's --wake-clip) instead of
        # GoToSleepOff; bench 2026-10-14: anim_gotosleep_wakeup_01 looked softer.
        self.wake_clip = wake_clip
        self.names = {mod.BROBOT_1_SERIAL: "Brobot 1", mod.BROBOT_2_SERIAL: "Brobot 2"}
        self.procs = {}
        self.signals = {}
        self.queues = {}
        self.results = []
        self.lock = threading.Lock()

    def say(self, text):
        print(f"{self.robots.timing_prefix()}{self.tag}_SLEEPY {text}", flush=True)

    @property
    def asleep(self):
        return bool(self.queues)

    def _sleep(self, serial):
        name = self.names[serial]
        self.say(f"{name}: falls asleep")
        if not self.live:
            self.procs[serial] = None
            return
        signal = Path(f"/tmp/gopod_sleepy_{serial}.signal")
        signal.unlink(missing_ok=True)
        self.signals[serial] = signal
        env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
        self.procs[serial] = subprocess.Popen(
            [str(SLEEP_BIN_PATH), "sleep", "--wait-signal", str(signal), "--max-wait", str(SLEEPY_MAX_WAIT_SECONDS)]
            + (["--wake-clip", self.wake_clip] if self.wake_clip else [])
            + [f"{name}:{serial}"],
            stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, env=env,
        )

    def _wake(self, serial):
        name = self.names[serial]
        proc = self.procs.pop(serial, None)
        if proc is None:
            self.say(f"{name}: wakes")
            return {"ok": True, "mode": "dry"}
        self.signals[serial].touch()
        try:
            out, _ = proc.communicate(timeout=SLEEPY_MAX_WAIT_SECONDS + 30)
        except subprocess.TimeoutExpired:
            proc.kill()
            out = "timed out"
        self.signals[serial].unlink(missing_ok=True)
        out = "\n".join(line for line in (out or "").splitlines() if "guid" not in line.lower())
        ok = proc.returncode == 0 and "status=OK" in out
        self.say(f"{name}: wakes ok={ok}")
        # The sleep tool has released; Wire-Pod takes over after the settle.
        time.sleep(DIRECT_SDK_RELEASE_SETTLE_SECONDS)
        return {"ok": ok, "output": out}

    def _worker(self, serial):
        q = self.queues[serial]
        name = self.names[serial]
        stop = False
        while not stop:
            item = q.get()
            if item is None:
                q.task_done()
                return
            taken = 1
            wake = self._wake(serial)
            held = False
            while item is not None:
                step_id, move, awake = item
                self.say(f"{name}: {step_id}")
                if move is None:
                    # A stir: awake a moment, nothing else, then back to sleep.
                    if self.live:
                        time.sleep(awake)
                    result = {"ok": True, "stir_awake_seconds": awake}
                else:
                    if self.live and not held:
                        self.mod.wirepod_web_send_form("/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10)
                        time.sleep(POST_ASSUME_SAY_SETTLE_SECONDS)
                    held = True
                    try:
                        result = move()
                    except Exception as exc:  # noqa: BLE001 - a sleepy move never stops the song
                        result = {"ok": False, "error": repr(exc)}
                with self.lock:
                    self.results.append({"step_id": step_id, "wake": wake, "result": result})
                # Lull: a move that's already due is made before sleeping again.
                try:
                    item = q.get_nowait()
                except queue.Empty:
                    break
                taken += 1
                if item is None:
                    stop = True  # wake_all: stay awake
            if self.live and held:
                self.mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)
                time.sleep(DIRECT_SDK_RELEASE_SETTLE_SECONDS)
            if not stop:
                self._sleep(serial)
            for _ in range(taken):
                q.task_done()

    def sleep_both(self, look_seconds=0):
        for serial in self.names:
            self._sleep(serial)
            self.queues[serial] = queue.Queue()
            threading.Thread(target=self._worker, args=(serial,), daemon=True).start()
        if look_seconds > 0:
            self.watching = threading.Event()
            threading.Thread(target=self._watch, args=(look_seconds, self.watching), daemon=True).start()

    def _watch(self, look_seconds, stop):
        """2026-10-09 (operator: "during brobots audio play, have occasional looks
        for brobot awake state, if awake send back to sleep"): every
        look_seconds, a Brobot whose sleep has ended (the sleep tool let go - he
        woke) and who isn't mid-stir is sent back to sleep."""
        while not stop.wait(look_seconds):
            for serial in list(self.names):
                q = self.queues.get(serial)
                with self.lock:
                    proc = self.procs.get(serial)
                    awake = proc is not None and proc.poll() is not None
                if not awake or q is None or q.unfinished_tasks:
                    continue
                out = (proc.stdout.read() if proc.stdout else "") or ""
                self.say(f"{self.names[serial]}: awake ({proc.returncode}) - back to sleep")
                with self.lock:
                    self.results.append({"step_id": "sleep_watch", "serial": serial, "woke": proc.returncode,
                                         "output": "\n".join(l for l in out.splitlines() if "guid" not in l.lower())})
                    self._sleep(serial)

    def move(self, serial, step_id, fn, awake=1.0):
        """fn None: a stir - the Brobot wakes, stays awake `awake` seconds, sleeps again."""
        self.queues[serial].put((step_id, fn, awake))

    def finish(self):
        """Wait until every queued move is made and its Brobot is asleep again."""
        for q in self.queues.values():
            q.join()

    def wake_all(self):
        """Stop the workers and wake both Brobots together."""
        if getattr(self, "watching", None) is not None:
            self.watching.set()
        if not self.queues:
            return []
        for q in self.queues.values():
            q.put(None)
        self.finish()
        self.queues = {}
        results = []
        threads = [threading.Thread(target=lambda s=serial: results.append(self._wake(s))) for serial in list(self.procs)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()
        return results


# The run's sleepy Brobots, woken in run_golden_song's finally if a run ends early.
_SLEEPY = None


def _tag_for(song_id):
    """Song_id-derived print-tag stem, e.g. brobots_baby_robots_sleep ->
    BROBOTS_BABY_ROBOTS_SLEEP. Replaces run_songs_runner_001.py's hardcoded
    "BINGO" literal per the survey's own schema table."""
    return re.sub(r"[^A-Za-z0-9]+", "_", song_id or "golden").upper()


# ---- note handlers (behavior ported verbatim from run_songs_runner_001.py) ----


def run_wake_both(mod, live, settle_seconds):
    if not live:
        return {"ok": True, "mode": "dry"}
    try:
        b1 = mod.wirepod_web_send_form("/api-sdk/conn_test", {"serial": mod.BROBOT_1_SERIAL}, timeout=10)
        b2 = mod.wirepod_web_send_form("/api-sdk/conn_test", {"serial": mod.BROBOT_2_SERIAL}, timeout=10)
    except Exception as exc:  # noqa: BLE001 - a hang/timeout/refused-connection here is exactly what this handler exists to catch
        return {"ok": False, "error": repr(exc)}
    time.sleep(settle_seconds)
    return {
        "ok": b1.get("status") == 200 and b2.get("status") == 200,
        "brobot_1": b1,
        "brobot_2": b2,
        "settle_seconds": settle_seconds,
    }


def run_connection_pass(mod, live, robots, tag):
    if not live:
        return {"ok": True, "mode": "dry"}
    results = {}
    all_ok = True
    for serial, label in ((mod.BROBOT_1_SERIAL, "Brobot 1"), (mod.BROBOT_2_SERIAL, "Brobot 2")):
        entry = {}
        try:
            conn = mod.wirepod_web_send_form(
                "/api-sdk/conn_test", {"serial": serial}, timeout=CONNECTION_PASS_TIMEOUT_SECONDS
            )
            entry["conn_test"] = conn
            say = mod.wirepod_web_send_form(
                "/api-sdk/say_text", {"text": "Connection check.", "serial": serial},
                timeout=CONNECTION_PASS_TIMEOUT_SECONDS,
            )
            entry["say_smoke_test"] = say
            entry["ok"] = (
                conn.get("status") == 200
                and say.get("status") == 200
                and say.get("body", "").strip() == "success"
            )
        except Exception as exc:  # noqa: BLE001 - a hang/timeout here is exactly what this pass exists to catch fast
            entry["ok"] = False
            entry["error"] = repr(exc)
        results[label] = entry
        all_ok = all_ok and entry["ok"]
        print(f"{robots.timing_prefix()}{tag}_CONNECTION_PASS {label}: ok={entry['ok']}", flush=True)
    if not all_ok:
        print(
            f"{robots.timing_prefix()}{tag}_CONNECTION_PASS_FAILED - at least one robot did not pass cleanly. "
            f"{WIREPOD_PAIRING_RECOVERY_MESSAGE}",
            flush=True,
        )
    return {"ok": all_ok, "results": results}


def post_countdown(mod, live, seconds, robots=None):
    """2026-10-01 (operator: "after please wait, add a countdown timer in the
    brobots window"): one page-only chat-feed line from the countdown role
    (gopod_brobots_settings.json) with the whole seconds; the page counts it
    down and removes it. Never stops a run."""
    role = load_brobots_settings()["countdown"]["role"]
    return post_chat_line(mod, live, role, str(max(1, int(round(seconds)))), robots)


def post_bingo_ball(mod, live, ball, robots=None):
    """2026-10-02 (operator: "make a really big bingo ball number display for
    B-1, I-9 & O-0"): one page-only chat-feed line from the bingo_ball role
    (gopod_brobots_settings.json) with the called number; the page shows a big
    ball and keeps it in the chat window. Never stops a run."""
    role = load_brobots_settings()["bingo_ball"]["role"]
    return post_chat_line(mod, live, role, str(ball), robots)


def rattle_countdown_seconds(extra_before=0.0):
    """The rattle's settle wait plus its own sound (countdown.rattle_sound_seconds)."""
    return extra_before + RATTLE_SETTLE_SECONDS + load_brobots_settings()["countdown"]["rattle_sound_seconds"]


def rattle_wav_seconds():
    """The rattle sound's own length, read from the WAV."""
    with wave.open(str(RATTLE_WAV_PATH)) as w:
        return w.getnframes() / float(w.getframerate())


def play_rattle_on_speaker(robots=None, tag="GOLDEN"):
    """2026-10-03 (operator: "make the rattle WAV play through system audio rather
    than a brobot"): the rattle on this machine's speaker (paplay, the default
    sink - GOPOD's own routing check keeps that the real speaker). No hand-off of
    the robot's control, no direct-SDK connection, no settle wait. Never stops a
    run."""
    try:
        # A run started by voice ("bingo test live") comes from Wire-Pod's root
        # process via runuser, with no desktop session in its environment, so
        # paplay couldn't find the sound server (live 2026-10-03 01:00: "Connection
        # refused"). Point it at this user's own session folder.
        env = dict(os.environ)
        env.setdefault("XDG_RUNTIME_DIR", f"/run/user/{os.getuid()}")
        proc = subprocess.run(["paplay", str(RATTLE_WAV_PATH)], capture_output=True, text=True, timeout=15, env=env)
        ok = proc.returncode == 0
        detail = (proc.stdout + proc.stderr).strip()
    except Exception as exc:  # noqa: BLE001 - a sound, never stops the song
        ok, detail = False, repr(exc)
    prefix = robots.timing_prefix() if robots else ""
    print(f"{prefix}{tag}_RATTLE_SPEAKER ok={ok}{(' ' + detail) if detail else ''}", flush=True)
    return {"ok": ok, "mode": "speaker", "detail": detail}


def volume_share(volume):
    """2026-10-17 (operator: "other story.md knobs are either \"\" for default or
    \"0-100%\""): a volume knob as a share of the master (chat_audio.volume) -
    1 for "" or "default", 0.3 for "30%"; a bare number is a share (0.3)."""
    if isinstance(volume, (int, float)):
        return max(0.0, float(volume))
    text = str(volume or "").strip()
    if not text or text.lower() == "default":
        return 1.0
    try:
        n = float(text.rstrip("%").strip())
    except ValueError:
        return 1.0
    return max(0.0, n / 100 if text.endswith("%") else n)


def play_song_sound(mod, live, song_dir, name, volume, robots=None, tag="GOLDEN"):
    """A song's own sound file on the chat page, at volume (0-1, the page's
    audio volume, "#vol=" on its path). Never stops a run."""
    path = next(Path(song_dir).rglob(name), None) if name else None
    # 2026-10-13 (the Fart Joke's danger-will-robinson.wav): else Wire-Pod's
    # shared sounds/ (copied to the page at its start too).
    if name and path is None:
        shared = mod.GOPOD_CHIPPER_ROOT / "sounds" / name
        path = shared if shared.is_file() else None
    seconds = 0.0
    if path:
        try:
            with wave.open(str(path)) as w:
                seconds = w.getnframes() / float(w.getframerate())
        except Exception:  # noqa: BLE001 - a float wav (the fart sfx): sox reads it
            try:
                seconds = float(subprocess.run(["soxi", "-D", str(path)], capture_output=True, text=True, timeout=10).stdout.strip() or 0)
            except Exception:  # noqa: BLE001 - no length known; no wait
                seconds = 0.0
    prefix = robots.timing_prefix() if robots else ""
    if not live or not path:
        print(f"{prefix}{tag}_SOUND {name} volume={volume} seconds={seconds:.2f} {'missing' if not path else 'dry'}", flush=True)
        return {"ok": bool(path), "mode": "dry", "seconds": seconds}
    rv = load_brobots_settings()["reporter_voices"]
    share = volume_share(volume)
    posted = post_chat_line(mod, live, rv["role"], f"/gopod_reporter_audio/{path.name}" + (f"#vol={share:g}" if share != 1 else ""), robots)
    time.sleep(seconds + float(rv.get("page_lag_seconds", 0)))
    print(f"{prefix}{tag}_SOUND {name} volume={volume} seconds={seconds:.2f}", flush=True)
    return {"ok": posted.get("ok", False), "mode": "page", "seconds": seconds}


def show_song_picture(mod, live, picture, robots=None, tag="GOLDEN"):
    """A song's picture step: the files (one name, or a list) from the chat
    page's pictures folder, shown in one picture bubble. Never stops a run."""
    tp = load_brobots_settings()["topic_pictures"]
    names = [picture] if isinstance(picture, str) else list(picture or [])
    names = [n.strip() for n in names if n and n.strip()]
    folder = mod.GOPOD_CHIPPER_ROOT / "webroot" / tp.get("folder", "pictures")
    missing = [n for n in names if not (folder / n).is_file()]
    prefix = robots.timing_prefix() if robots else ""
    if not names or missing:
        print(f"{prefix}{tag}_PICTURE missing {missing or 'no picture named'}", flush=True)
        return {"ok": False, "missing": missing}
    line = "|".join(f"/{tp.get('folder', 'pictures')}/{n}" for n in names)
    if not live:
        print(f"{prefix}{tag}_PICTURE {line} dry", flush=True)
        return {"ok": True, "mode": "dry"}
    posted = post_chat_line(mod, live, tp.get("role", "Picture"), line, robots)
    print(f"{prefix}{tag}_PICTURE {line}", flush=True)
    return {"ok": posted.get("ok", False), "mode": "page"}


def play_rattle_on_page(mod, live, robots=None, tag="GOLDEN"):
    """2026-10-09 (operator: "keep but pause playing through the brobot ... instead
    make it play through the web page like the reporter voices"): the rattle is
    put in the chat feed as a reporter-voice line, so the Brobots chat page plays
    it where its other voices play, in its queue, with the countdown shown for
    exactly as long. Wire-Pod copies the overlay's sounds/rattle.wav into the page's
    voice folder at its start. The song waits the sound's
    length plus reporter_voices.page_lag_seconds. Never stops a run."""
    rv = load_brobots_settings()["reporter_voices"]
    try:
        # Wire-Pod puts the overlay's sounds/ in the page's voice folder at its
        # start (gopodSoundsToPage) - that folder is Wire-Pod's (root's), so the
        # song only names the file (live 16:23: "Permission denied" copying it).
        seconds = rattle_wav_seconds()
        post_countdown(mod, live, seconds, robots)
        posted = post_chat_line(mod, live, rv["role"], "/gopod_reporter_audio/rattle.wav", robots)
        time.sleep(seconds + float(rv.get("page_lag_seconds", 0)))
        ok, detail = posted.get("ok", False), ""
    except Exception as exc:  # noqa: BLE001 - a sound, never stops the song
        ok, detail = False, repr(exc)
    prefix = robots.timing_prefix() if robots else ""
    print(f"{prefix}{tag}_RATTLE_PAGE ok={ok}{(' ' + detail) if detail else ''}", flush=True)
    return {"ok": ok, "mode": "page", "detail": detail}


def post_countdown_end(mod, live, robots=None):
    """2026-10-04 (operator: "timer too fast to hear rattle audio - have visuals
    respond to what's there"): when the rattle has really finished, one more
    countdown-role line (countdown.end_text) tells the page to remove the timer,
    so it shows exactly as long as the sound plays. Never stops a run."""
    countdown = load_brobots_settings()["countdown"]
    return post_chat_line(mod, live, countdown["role"], countdown["end_text"], robots)


def run_rattle(mod, serial, live, robots, tag, on="robot"):
    """The page's timer stays until the rattle has finished (post_countdown_end);
    see _run_rattle_play."""
    if not live:
        return {"ok": True, "mode": "dry"}
    try:
        return _run_rattle_play(mod, serial, live, robots, tag, on)
    finally:
        post_countdown_end(mod, live, robots)


def _run_rattle_play(mod, serial, live, robots, tag, on="robot"):
    """on (rattle_on in the song's knobs.json): "page" - the Brobots chat page
    plays it like a reporter voice (2026-10-09); "robot" - Brobot 1 plays it over
    the direct SDK after handing back control; "speaker" - this machine's
    speaker plays it while the robots stay under the song's control."""
    if not live:
        return {"ok": True, "mode": "dry"}
    if on == "page":
        return play_rattle_on_page(mod, live, robots, tag)
    if on == "speaker":
        post_countdown(mod, live, rattle_wav_seconds(), robots)
        return play_rattle_on_speaker(robots, tag)
    post_countdown(mod, live, rattle_countdown_seconds(), robots)
    release = mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)
    time.sleep(RATTLE_SETTLE_SECONDS)
    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    try:
        proc = subprocess.run(
            [str(RATTLE_BIN_PATH), serial, str(RATTLE_WAV_PATH)],
            capture_output=True, text=True, timeout=30, env=env,
        )
    except subprocess.TimeoutExpired:
        print(f"{robots.timing_prefix()}{tag}_RATTLE_DIRECT serial={serial} TIMED OUT (30s)", flush=True)
        return {"ok": False, "error": "rattle binary timed out", "release": release}
    combined = "\n".join(
        line for line in (proc.stdout + proc.stderr).splitlines() if "guid" not in line.lower()
    )
    ok = proc.returncode == 0 and "status=OK" in combined
    print(f"{robots.timing_prefix()}{combined}", flush=True)
    return {"ok": ok, "release": release, "returncode": proc.returncode, "output": combined}


def run_wheel_hold(mod, serial, live):
    """2026-09-30: stop the wheels (move_wheels lw=0 rw=0 - the same stop the
    wheel-nudge alias ends with, brobots.sh). Never stops a run."""
    if not live:
        return {"ok": True, "mode": "dry"}
    try:
        result = mod.wirepod_web_send_form("/api-sdk/move_wheels", {"serial": serial, "lw": "0", "rw": "0"}, timeout=10)
    except Exception as exc:  # noqa: BLE001 - a hold, never stops the song
        result = {"error": repr(exc)}
    return {"ok": result.get("status") == 200, "result": result}


def post_chat_line(mod, live, name, text, robots=None):
    """2026-09-30: a chat-page-only line - nothing is said. Wire-Pod's say_text
    with empty text and a display_text (sdkapp/server.go) puts the line in its
    numbered chat feed (gopod_chat_feed.go) for the Brobots chat page. name is
    the speaker the page shows ("Operator", "Brobot 1", "Brobot 2"). Never
    stops a run: a failed post is printed and skipped."""
    if not live or not text:
        return {"ok": True, "mode": "dry"}
    try:
        result = mod.wirepod_web_send_form(
            "/api-sdk/say_text",
            {"text": "", "serial": mod.BROBOT_1_SERIAL, "display_text": text, "display_role": name},
            timeout=10,
        )
    except Exception as exc:  # noqa: BLE001 - page-only, never stops the song
        result = {"error": repr(exc)}
    ok = result.get("status") == 200
    if not ok:
        prefix = robots.timing_prefix() if robots else ""
        print(f"{prefix}CHAT_LINE_SKIPPED {name}: {text!r} ({result})", flush=True)
    return {"ok": ok, "result": result}


def run_cube_blip(mod, live, serial):
    """The cube blip on a Brobot's own cube (phcal's cmd_cube, as-is). It takes
    control itself. 2026-10-10 bench: the cube wakes a sleeping Brobot, so a
    sleepy song lights it only once both are awake (operator: "add the
    cube-call to the brobots ready part, just after the brobots are called to
    wake anyways")."""
    if not live:
        return {"ok": True, "mode": "dry", "serial": serial}
    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    try:
        proc = subprocess.run([str(CUBE_BLIP_BIN_PATH), serial], capture_output=True, text=True, timeout=30, env=env)
    except subprocess.TimeoutExpired:
        return {"ok": False, "serial": serial, "error": "cube blip timed out (30s)"}
    combined = "\n".join(line for line in (proc.stdout + proc.stderr).splitlines() if "guid" not in line.lower())
    return {"ok": proc.returncode == 0 and "status=OK" in combined, "serial": serial, "output": combined}


def start_cube_connect(live, serial):
    """2026-10-14 (operator: the nap's cube "takes forever to respond ... I need
    sync/async, not sequential based on time estimates"; "aim for cube option
    2"): the cube tool's connect mode in the background - the robot connects his
    cube over Bluetooth (about 18 s, ConnectCube) while the song goes on, so the
    later blip finds it connected. Returns the running job."""
    if not live:
        return {"mode": "dry", "serial": serial}
    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    proc = subprocess.Popen([str(CUBE_BLIP_BIN_PATH), serial, "connect"], stdout=subprocess.PIPE,
                            stderr=subprocess.STDOUT, text=True, env=env)
    return {"proc": proc, "serial": serial, "started": time.time()}


def start_cube_light(live, serial, most_seconds=60.0, after_seconds=0.0):
    """2026-10-15 (the nap, operator: "Cube always took way too long to light
    up"): the cube tool's light-on-signal mode in the background - the robot
    connects his cube now (about 12 s) and holds it; light_cube() then lights it
    at once. most_seconds: it lights anyway if no signal comes."""
    if not live:
        return {"mode": "dry", "serial": serial}
    signal = Path(f"/tmp/gopod_cube_light_{serial}.signal")
    signal.unlink(missing_ok=True)
    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    proc = subprocess.Popen([str(CUBE_BLIP_BIN_PATH), serial, "light-on-signal", str(signal), str(most_seconds + after_seconds), str(after_seconds)],
                            stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, env=env)
    return {"proc": proc, "serial": serial, "signal": signal, "started": time.time()}


def light_cube(job, most_seconds=30.0):
    """Lights a cube held by start_cube_light and waits for it (the connect, if
    it isn't done yet, then the lights) - up to most_seconds."""
    if job.get("mode") == "dry":
        return {"ok": True, "mode": "dry", "serial": job["serial"]}
    job["signal"].touch()
    proc = job["proc"]
    try:
        out, _ = proc.communicate(timeout=most_seconds)
    except subprocess.TimeoutExpired:
        proc.kill()
        out, _ = proc.communicate()
        job["signal"].unlink(missing_ok=True)
        return {"ok": False, "serial": job["serial"], "error": f"cube light timed out ({most_seconds:g}s)"}
    job["signal"].unlink(missing_ok=True)
    combined = "\n".join(line for line in (out or "").splitlines() if "guid" not in line.lower())
    return {"ok": proc.returncode == 0 and "status=OK" in combined, "serial": job["serial"],
            "since_wake_seconds": round(time.time() - job["started"], 3), "output": combined}


def finish_cube_connect(job, most_seconds=40.0):
    """Waits for a background cube connect (start_cube_connect) - up to
    most_seconds from when it started - and returns its result."""
    if job.get("mode") == "dry":
        return {"ok": True, "mode": "dry", "serial": job["serial"]}
    proc = job["proc"]
    left = max(0.0, most_seconds - (time.time() - job["started"]))
    try:
        out, _ = proc.communicate(timeout=left)
    except subprocess.TimeoutExpired:
        proc.kill()
        out, _ = proc.communicate()
        return {"ok": False, "serial": job["serial"], "error": f"cube connect timed out ({most_seconds:g}s)"}
    combined = "\n".join(line for line in (out or "").splitlines() if "guid" not in line.lower())
    return {"ok": proc.returncode == 0 and "status=OK" in combined, "serial": job["serial"],
            "waited_seconds": round(time.time() - job["started"], 3), "output": combined}


def run_brobots_ready_together(mod, live, phrase="Brobots ready!", before_speech=None, rattle_before=False, rattle_on="robot"):
    """before_speech (2026-09-30): called right before the together line is
    said, so its chat bubbles show just ahead of the voices.
    rattle_before (2026-09-30, operator: "play rattle audio before, brobots
    ready bubbles appear & brobots ready spoken"): Brobot 1's rattle, played
    while both robots are already handed over, after the rattle's own locked
    wait (RATTLE_SETTLE_SECONDS from the release). The first rattle of a run
    has been the weak one (BROBOTS_RUN_2147_FIXES_001.md). Never stops the step."""
    if not live:
        if before_speech:
            before_speech()
        return {"ok": True, "mode": "dry"}
    if rattle_before and rattle_on == "page":
        # rattle_on "page" (2026-10-09): the warm-up rattle on the chat page too.
        play_rattle_on_page(mod, live, None)
        post_countdown_end(mod, live, None)
        rattle_before = False
    if rattle_before and rattle_on == "speaker":
        # rattle_on "speaker" (2026-10-03): the warm-up on this machine's
        # speaker, before the hand-off - no settle waits around it.
        post_countdown(mod, live, rattle_wav_seconds(), None)
        play_rattle_on_speaker()
        post_countdown_end(mod, live, None)
        rattle_before = False
    if rattle_before:
        # The warm-up: its settle, the sound, then the 3.0 s before the line.
        post_countdown(mod, live, rattle_countdown_seconds(RATTLE_SETTLE_SECONDS), None)
    release_1 = mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": mod.BROBOT_1_SERIAL}, timeout=10)
    release_2 = mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": mod.BROBOT_2_SERIAL}, timeout=10)
    time.sleep(DIRECT_SDK_RELEASE_SETTLE_SECONDS)
    warmup = None
    if rattle_before:
        time.sleep(max(0.0, RATTLE_SETTLE_SECONDS - DIRECT_SDK_RELEASE_SETTLE_SECONDS))
        env_w = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
        try:
            proc_w = subprocess.run(
                [str(RATTLE_BIN_PATH), mod.BROBOT_1_SERIAL, str(RATTLE_WAV_PATH)],
                capture_output=True, text=True, timeout=30, env=env_w,
            )
            warmup = {"returncode": proc_w.returncode, "output": "\n".join(
                line for line in (proc_w.stdout + proc_w.stderr).splitlines() if "guid" not in line.lower())}
        except subprocess.TimeoutExpired:
            warmup = {"error": "rattle before timed out"}
        post_countdown_end(mod, live, None)
        # 2026-10-01 (operator: "yes, do it"): the rattle's own 3.0 s settle, not
        # 1.0 s, before "Brobots ready" - when the warm-up was heard, Brobot 1
        # started its line late (out of sync), likely still busy from the rattle.
        time.sleep(RATTLE_SETTLE_SECONDS)
    if before_speech:
        before_speech()
    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    try:
        proc = subprocess.run(
            [str(TOGETHER_BIN_PATH), phrase, f"Brobot 1:{mod.BROBOT_1_SERIAL}", f"Brobot 2:{mod.BROBOT_2_SERIAL}"],
            capture_output=True, text=True, timeout=30, env=env,
        )
    except subprocess.TimeoutExpired:
        time.sleep(BROBOTS_READY_TOGETHER_HANDBACK_SETTLE_SECONDS)
        return {"ok": False, "error": "together binary timed out", "release_1": release_1, "release_2": release_2}
    combined = "\n".join(
        line for line in (proc.stdout + proc.stderr).splitlines() if "guid" not in line.lower()
    )
    ok = proc.returncode == 0 and combined.count("status=OK") == 2
    time.sleep(BROBOTS_READY_TOGETHER_HANDBACK_SETTLE_SECONDS)
    return {"ok": ok, "release_1": release_1, "release_2": release_2, "returncode": proc.returncode, "output": combined, "rattle_before": warmup}


def run_say_turn(mod, robots, name, text, display_text=None, suppress_assume_release=False):
    """display_text/suppress_assume_release are additive (both default to the
    bingo-family's existing behavior - separate display text unused, a normal
    per-call assume/release) so every existing bingo-family caller is
    byte-identical. Control-family songs (manage_control=False) need
    suppress_assume_release=True - "connect" already assumed control once for
    the whole run (golden truth 2, run_robot_control_song_001.py) - and the
    weather note needs a separate spoken-vs-displayed string (no degree
    symbol spoken, one shown)."""
    if not text:
        return {"ok": True, "skipped": True}
    if display_text is None:
        display_text = text
    # A combo part: the step's one bubble, put up by its first speaking part.
    if getattr(robots, "piece_offset", None) is not None:
        display_text = getattr(robots, "combo_display", None) or ""
    robot_safe_line, _repair_used = mod.normalize_robot_safe(text, robots.scaffold)
    started = time.monotonic()
    result = robots.say(
        name, robot_safe_line, display_text=display_text, display_role=name,
        suppress_assume_release=suppress_assume_release,
    )
    elapsed = time.monotonic() - started
    heal = None
    # 2026-09-30 self-heal (BINGO_B1_SILENT_LINES_TIMING_COMPARE_001.md): under the
    # song's continuous hold, a robot can lose its control without Wire-Pod
    # noticing - its lines then come back in 0.03-0.18s, unspoken, while Wire-Pod
    # still answers "success", and it stays silent until the next rattle's
    # release/retake. A real line takes >= ~1s (the shortest, "Wait!"). So: a line
    # back in under SAY_SILENT_SECONDS gets the rattle's own recovery for that one
    # robot - release, settle, retake, settle - and is said once more.
    if suppress_assume_release and result.get("mode") == "wirepod" and elapsed < SAY_SILENT_SECONDS:
        serial = mod.BROBOT_1_SERIAL if name == "Brobot 1" else mod.BROBOT_2_SERIAL
        print(f"{robots.timing_prefix()}SAY_SILENT {name}: line came back in {elapsed:.2f}s - retaking control and saying it again", flush=True)
        release = mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)
        time.sleep(DIRECT_SDK_RELEASE_SETTLE_SECONDS)
        assume = mod.wirepod_web_send_form("/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10)
        time.sleep(POST_ASSUME_SAY_SETTLE_SECONDS)
        retry_started = time.monotonic()
        # The first try already put the line on the chat page (Wire-Pod adds it
        # before speaking), so the retry is spoken only - no second bubble.
        result = robots.say(
            name, robot_safe_line, display_text="", display_role=name,
            suppress_assume_release=suppress_assume_release,
        )
        retry_elapsed = time.monotonic() - retry_started
        heal = {"first_seconds": round(elapsed, 3), "retry_seconds": round(retry_elapsed, 3), "release": release, "assume": assume}
        print(f"{robots.timing_prefix()}SAY_SILENT {name}: retry took {retry_elapsed:.2f}s{' - STILL SILENT' if retry_elapsed < SAY_SILENT_SECONDS else ''}", flush=True)
    ok = result.get("spoken") is True or result.get("mode") in ("dry", "off")
    return {"ok": ok, "result": result, "silent_heal": heal}


# The Brobots settings file (single source of truth, shared with Wire-Pod and the
# chat page): the animation lead-in the emotion beats use.
BROBOTS_SETTINGS_PATH = (
    Path(__file__).resolve().parents[3] / "wire_pod_overlay" / "chipper" / "webroot" / "gopod_brobots_settings.json"
)


def load_brobots_settings():
    return json.loads(BROBOTS_SETTINGS_PATH.read_text(encoding="utf-8"))


def load_animation_lead_in():
    return load_brobots_settings()["animation_lead_in"]


def _llm_loaded(base, model):
    """True when Ollama already has the model loaded (its /api/ps list)."""
    root = base[:-3] if base.endswith("/v1") else base
    try:
        with urllib.request.urlopen(root.rstrip("/") + "/api/ps", timeout=5) as resp:
            loaded = json.loads(resp.read().decode("utf-8")).get("models", [])
    except Exception:  # noqa: BLE001 - unknown means warm it up
        return False
    return any(m.get("name") == model or m.get("model") == model for m in loaded)


def wait_for_page(mod, where, stop_request=None, prefix=""):
    """2026-10-17 (operator: "instead, system message, Goverlord awaits your
    return... + hourglass, and waits, nothing plays"; "both the terminal run &
    web page run must be the same run"): while the chat page is away - hidden,
    closed, or not heard from (Wire-Pod keeps session.page_hidden_file) - the
    run waits here, and says so in the terminal once, then when it goes on."""
    session = load_brobots_settings()["session"]
    flag = mod.GOPOD_CHIPPER_ROOT / (session.get("page_hidden_file") or "gopod_page_hidden")
    if not flag.exists():
        return
    started = time.time()
    print(f"{prefix}GOLDEN_AWAIT {where}: the chat page is away - {session.get('awaits_text') or 'waiting'}", flush=True)
    while flag.exists() and not (stop_request is not None and stop_request.exists()):
        time.sleep(0.5)
    print(f"{prefix}GOLDEN_AWAIT the chat page is back after {time.time() - started:.1f}s - going on", flush=True)


def song_llm_warmup(mod, ready_text="", lines=None, voices_required=True, llm_required=True):
    """2026-10-04 (operator: "upon bingo test live, force llm warm up timer
    visual, for minimum 5 seconds (or until actual warm up completes), with
    "Warming Up LLM""): a song with "llm_warmup": true in its knobs.json, on a
    live run, shows the chat page's warm-up bubble (the LLM power level, then
    "LLM Ready" - llm.warmup_text / llm.ready_text) while one tiny request to the chat's own model (Wire-Pod's
    knowledge endpoint and model, nothing else read) makes sure it's loaded. The
    bubble stays at least llm.song_warmup_min_seconds, or until the request is
    done if that's longer. Same page file Wire-Pod's own start warm-up uses."""
    llm = load_brobots_settings()["llm"]
    state = mod.GOPOD_CHIPPER_ROOT / "webroot" / "gopod_llm_warmup.json"
    # 2026-10-14 (operator: "part 1 standard * keep it in settings file once *
    # each song gives only part 2"): llm.run_ready_text, then the song's own
    # llm_ready_text.
    # 2026-10-16 (operator: "snapshot/lock/stamp this starting opener LLM as the
    # golden reference template * for song runs, apply this template as a
    # base"): the opener's message 3 shape - run_ready_text, the song's own
    # line, run_ready_end - one line each (gopod_llm_warmup.go gopodRunReadyLine).
    # 2026-10-17 (operator: "add: fallbacks to system messages: \"No Voice
    # Generation Required\" or voice generation check * \"No Language Model
    # Required\" of language model loading/loaded"): a song that needs no voice
    # files made or no language model (knobs.json "voices_required" /
    # "llm_required": false, the nap) says so in their place, and no model is called.
    loaded_text = (llm.get("run_ready_text") or "").strip() if llm_required else (llm.get("no_llm_text") or "").strip()
    ready_text = "\n".join(t for t in (loaded_text, (ready_text or "").strip(),
                                        (llm.get("run_ready_end") or "").strip()) if t)
    warming_text = llm.get("warmup_text", "")
    if not (voices_required and llm_required):
        voices_part, _, loading_part = warming_text.partition(". ")
        voices_part = (voices_part + "." if loading_part else voices_part) if voices_required else (llm.get("no_voices_text") or "").strip()
        loading_part = loading_part if llm_required else (llm.get("no_llm_text") or "").strip()
        if voices_part and loading_part and voices_part[-1] not in ".!?":
            voices_part += "."
        warming_text = " ".join(t for t in (voices_part, loading_part) if t)
    # 2026-10-17 (operator: "to skip in-between messages, a single in-between
    # message = \"\" = the indication to skip"): a line with no words isn't said.
    lines = [line for line in (lines or []) if (line.get("text") or "").strip()]

    started_ms = int(time.time() * 1000)

    def ready():
        # 2026-10-13 (operator: "chat bubbles must run in an orderly sequence so
        # Goverlord's orchestration is presented correctly"): the bubble turns
        # ready in the chat's order, after the lines before it, then Goverlord
        # says the ready line - the song goes on once it's said.
        try:
            answer = mod.wirepod_web_send_form("/api-sdk/gopod_system_say", {
                "warmup_ready": str(started_ms), "text": ready_text}, timeout=120)
            if isinstance(answer, dict) and str(answer.get("body", "")).strip() == "not_heard":
                print("LLM_WARMUP ready line not heard on the page", flush=True)
        except Exception as exc:  # noqa: BLE001 - the ready line never stops the song
            print(f"LLM_WARMUP ready line failed {type(exc).__name__}", flush=True)

    def message(text):
        # 2026-10-13 (operator: "Sequential Run Check"): the System llm bubble's
        # message, then llm.wait_text ("Please wait.").
        text, wait = (text or "").strip(), (llm.get("wait_text") or "").strip()
        if text and wait and text[-1] not in ".!?":
            text += "."
        if wait and text.endswith(wait):
            return text
        return (text + "\n" + wait).strip()

    def show(text, say=""):
        # GOLDEN, locked (2026-10-13, operator: "PLEASE FIX GOLDEN AND FUCKING
        # LOCK THE PROCESS IN"): the bubble's message as a chat-feed line, so the
        # page shows it in turn (gopod_llm_warmup.go GopodSongWarmupText), and
        # said by Goverlord when say is given.
        write(True, text)
        try:
            mod.wirepod_web_send_form("/api-sdk/gopod_system_say", {"warmup_text": str(started_ms), "text": text}, timeout=10)
            say_role = load_brobots_settings()["reporter_voices"].get("system_say_role", "")
            if say and say_role:
                # Voice only: Goverlord says it, no bubble of its own.
                post_chat_line(mod, True, say_role, say)
        except Exception as exc:  # noqa: BLE001 - a message never stops the song
            print(f"LLM_WARMUP message failed {type(exc).__name__}", flush=True)

    def write(active, text=""):
        # Same shape Wire-Pod writes (gopod_llm_warmup.go): the page fills the LLM
        # power level from started_ms, shows text while warming, and turns ready
        # at the feed's warmup_ready line.
        tmp = state.with_suffix(".json.tmp")
        # 2026-10-09 (operator: "maybe this time say LLM ready for Bingo?"): the
        # song's own ready text (knobs.json llm_ready_text), when it has one.
        tmp.write_text(json.dumps({"active": active, "started_ms": started_ms,
                                   "done_ms": 0 if active else int(time.time() * 1000),
                                   "ready_text": ready_text, "text": text}), encoding="utf-8")
        tmp.replace(state)

    knowledge = json.loads((mod.WIREPOD_CHIPPER_ROOT / "apiConfig.json").read_text(encoding="utf-8")).get("knowledge", {})
    base = (knowledge.get("endpoint") or "").rstrip("/")
    model = knowledge.get("model") or ""
    del knowledge  # only the endpoint and model are used; nothing else is kept or logged
    started = time.time()
    warm = False
    # 2026-10-11 (operator: "copy/add this reporters opener as-is to bingo *
    # include the llm warm-up"): the song's warmup_lines (knobs.json), each said
    # by its reporter through Wire-Pod, one after another.
    def say_lines():
        for line in lines or []:
            try:
                answer = mod.wirepod_web_send_form("/api-sdk/gopod_system_say", {
                    "who": line.get("who", ""), "text": line.get("text", ""), "wait_seconds": "30"}, timeout=40)
                # 2026-10-17 (survey E4): Wire-Pod says the truth now - heard or not.
                if isinstance(answer, dict) and str(answer.get("body", "")).strip() == "not_heard":
                    print(f"LLM_WARMUP line not heard on the page: {line.get('who', '')}", flush=True)
            except Exception as exc:  # noqa: BLE001 - a reporter line never stops the song
                print(f"LLM_WARMUP line failed {type(exc).__name__}", flush=True)
    # 2026-10-13 (operator: "\"System llm bubble\" displays first and always
    # stays at the bottom ... until LLM Ready"; "Sequential Run Check"): the
    # System llm bubble first with "LLM Warming Up. Please wait." (held a moment
    # when warm); then the lines with "Please wait."; then the ready line.
    try:
        # 2026-10-13 (operator: "\"LLM Warming up.\" not included in message
        # update"): the LLM Warming Up step always shows; held when warm.
        # 2026-10-14 (operator: "I did not hear the \"Please wait.\" at the end *
        # generate the full message displayed"): said as shown.
        show(message(warming_text), message(warming_text))
        if not llm_required:
            print("LLM_WARMUP no language model required", flush=True)
            left = float(llm["song_warmup_min_seconds"])
            if left > 0:
                time.sleep(left)
        elif base and model and _llm_loaded(base, model):
            print(f"LLM_WARMUP already warm model={model}", flush=True)
        else:
            print(f"LLM_WARMUP start model={model}", flush=True)
            if base and model:
                body = json.dumps({"model": model, "messages": [{"role": "user", "content": "Say ready."}],
                                   "stream": False, "max_tokens": 1}).encode()
                req = urllib.request.Request(base + "/chat/completions", data=body, headers={"Content-Type": "application/json"})
                try:
                    urllib.request.urlopen(req, timeout=600).read()
                    print(f"LLM_WARMUP done took={int((time.time() - started) * 1000)}", flush=True)
                    warm = True
                except Exception as exc:  # noqa: BLE001 - the warm-up never stops the song
                    print(f"LLM_WARMUP failed {type(exc).__name__}", flush=True)
            # 2026-10-12: the minimum only holds a warm-up that didn't answer.
            left = float(llm["song_warmup_min_seconds"]) - (time.time() - started)
            if left > 0 and not warm:
                time.sleep(left)
        say_lines()
    finally:
        write(False)
        # 2026-10-09 (operator: "no 'LLM Ready for Bingo' voice audio"): Goverlord
        # says it too.
        ready()


def _start_song_soundtrack(mod, live, song, first_at, robots, tag):
    """The song's soundtrack (knobs.json "audio", a file Wire-Pod puts in the page's
    voice folder - settings page_sound_files) posted as a song-audio line; the
    page plays it in the background. Returns the timeline's start: when the page
    starts playing it (reporter_voices.page_lag_seconds after the post), the zero
    every start_seconds counts from."""
    rv = load_brobots_settings()["reporter_voices"]
    role = rv.get("song_audio_role", "")
    posted = post_chat_line(mod, live, role, "/gopod_reporter_audio/" + song["audio"], robots) if role else {"ok": False}
    print(f"{robots.timing_prefix() if robots else ''}{tag}_SOUNDTRACK {song['audio']} ok={posted.get('ok')}", flush=True)
    return time.time() + float(rv.get("page_lag_seconds", 0))


def _caption_bubbles(song, marker):
    """2026-10-09 (operator: "instead of line by line, do several phrases in fewer
    bigger chat bubbles * use same highlighting per narration segment"): one
    narrator bubble per section, posted at its first caption's second. Each
    caption is one piece, "@start-end words", joined by the segment marker; the
    page lights each piece at its seconds on the soundtrack's own clock."""
    bubbles = []
    for s in song["steps"]:
        if s.get("start_seconds") is None or not s.get("caption"):
            continue
        start = float(s["start_seconds"])
        end = float(s.get("end_seconds", start))
        piece = f"@{start:.2f}-{end:.2f} {s['caption']}"
        # 2026-10-18 (operator: "grab timings from the audio file ... apply forced
        # alignment timings, following the way kokoro does it for highlighting"):
        # the caption's word times, from its own start, ride on its piece
        # ("@start-end ~<base64 json [[word, s, e], ...]> words"); the page
        # reveals each word at its time.
        timed = (song.get("narrator_words") or {}).get(f"{start:.2f}")
        if timed:
            rel = [[w["t"], round(w["s"] - start, 3), round(w["e"] - start, 3)] for w in timed]
            b64 = base64.b64encode(json.dumps(rel, separators=(",", ":")).encode("utf-8")).decode("ascii")
            piece = f"@{start:.2f}-{end:.2f} ~{b64} {s['caption']}"
        section = s.get("section", "")
        if bubbles and bubbles[-1][1] == section:
            bubbles[-1][2].append(piece)
        else:
            bubbles.append((start, section, [piece]))
    # 2026-10-15 (operator: "one image for each narrator bubble * image is inside
    # the narrator bubble under highlighted text"): knobs.json "section_pictures"
    # - the section's picture (in the page's pictures/ folder) as a last
    # "picture:" piece; the page shows it under the words, with the magnifier.
    pictures = song.get("section_pictures") or {}
    # 2026-10-15 (the nap): "narrator_single": true - one
    # narrator bubble, each section's words replacing the last over its picture.
    single = ["narrator:single"] if song.get("narrator_single") else []
    # 2026-10-15 (operator: "leave last image (5) at 50% solid (no fade out)"):
    # the last section is marked, so its picture stays.
    return [(at, marker.join(pieces + (["picture:/pictures/" + pictures[section]] if pictures.get(section) else []) + single
                             + (["narrator:last"] if single and i == len(bubbles) - 1 else [])))
            for i, (at, section, pieces) in enumerate(bubbles)]


def _start_captions(mod, live, song, timeline_start, stop, robots):
    """Each section's captions (a robots_speak false song's story.md words) posted
    as one bubble from reporter_voices.narrator_role at its first start_seconds,
    in the background."""
    settings = load_brobots_settings()
    role = settings["reporter_voices"].get("narrator_role", "")
    # GOLDEN chat bubble rule (reporter_voices.narrator_lead_seconds): sent
    # ahead, built invisible on the page, shown at its first line's second.
    lead = float(settings["reporter_voices"].get("narrator_lead_seconds", 0))
    captions = _caption_bubbles(song, settings["scripted_split"]["segment_marker"])
    if not role or not captions:
        return

    def run():
        for at, text in captions:
            wait = timeline_start + at - lead - time.time()
            if wait > 0 and stop.wait(wait if live else 0):
                return
            if stop.is_set():
                return
            post_chat_line(mod, live, role, text, robots)

    threading.Thread(target=run, daemon=True).start()


def session_run(mod, action, **form):
    """Checks a run in or out with Goverlord (/api-sdk/gopod_run,
    gopod_session.go). For "start": True (checked in), False (refused), or None
    when Wire-Pod can't be reached - the run goes on, as before."""
    try:
        result = mod.wirepod_web_send_form("/api-sdk/gopod_run", {"action": action, **form}, timeout=10)
        reply = json.loads(result.get("body") or "{}")
    except Exception as exc:  # noqa: BLE001 - a session hiccup never stops a run
        print(f"GOLDEN_SESSION {action} not reached: {exc!r}", flush=True)
        return None
    print(f"GOLDEN_SESSION {action} {form.get('name', '')!r} -> {reply}", flush=True)
    return bool(reply.get("ok")) if action == "start" else True


def song_stop_request_path(mod):
    """2026-10-04 (operator: "backpack the brobot 'not' speaking and give a verbal,
    'kill bingo test' to stop the current run at the next safe spot"): the voice
    command creates this file (song_stop.request_file, gopod_brobots_settings.json,
    in GOPOD's overlay chipper folder); the engine checks it before each step and
    ends the run there, the normal way - control handed back, touch interrupt on."""
    return mod.GOPOD_CHIPPER_ROOT / load_brobots_settings()["song_stop"]["request_file"]


def touch_interrupt_off_path(mod):
    """2026-10-01 (operator: "Touch interrupt ON/OFF to default ON again"):
    touch interrupt is on, as in native Wire-Pod, except while this file exists
    in GOPOD's overlay chipper folder (touch_interrupt.off_file in
    gopod_brobots_settings.json, read by kgsim_interrupt.go). A song with
    "touch_interrupt_off": true in its knobs.json creates it for its live run;
    run_golden_song deletes it when the run ends, and kill-bingo-test does too."""
    return mod.GOPOD_CHIPPER_ROOT / load_brobots_settings()["touch_interrupt"]["off_file"]


# ---- scripted split (2026-10-01) ----
# Operator: "scripted lane needs a standardized format to split sentences for
# highlighting" - "set up a starter standard", "Things like example, "Ah... I like
# to party", the "..." and other expressions are correctly grouped". A song opts
# in with "sentence_split": "scripted" in its knobs.json. Each line is split here,
# once; the pieces are joined with scripted_split.segment_marker (an invisible
# character, gopod_brobots_settings.json), and Wire-Pod (one say per piece) and
# the chat page (one highlight per piece) both split on that marker - so the
# robot and the bubble always agree. The live LLM chat keeps its own rule.
#
# The standard, v1:
#   1. A piece ends at . ! or ? (or a run of them, "!!") followed by a space.
#   2. An ellipsis (... or the single-character one) never ends a piece: the
#      words after it stay in the same piece ("Ah... I like to party").
#   3. A closing quote or bracket right after the end mark stays with its piece
#      ('It's "I, Won!" Lol.').
#   4. A trailing tag on scripted_split.join_trailing ("Lol", "Haha", "Ha-ha"),
#      with or without its own punctuation, joins the piece before it.
#   5. A line with no letters or digits (an emotion beat's ",") is not split.
def _scripted_end_re(piece_ends):
    """A piece ends at a run of piece_ends (scripted_split.piece_ends in the
    settings file; the comma since 2026-10-01) or an ellipsis, plus any closing
    quote, before a space."""
    ends = re.escape(piece_ends) + "\u2026"
    return re.compile(r"[" + ends + r"]+[\"'\)\]\u201d\u2019]*(?=\s)")


def scripted_pieces(text, join_trailing, piece_ends=".!?"):
    text = (text or "").strip()
    if not re.search(r"[A-Za-z0-9]", text):
        return [text] if text else []
    pieces, start = [], 0
    for match in _scripted_end_re(piece_ends).finditer(text):
        run = match.group()
        if "..." in run or "\u2026" in run:
            continue
        # A comma inside a quote ("I, Won!") is the line's own pause, not a
        # piece end.
        if run.startswith(",") and text[:match.start()].count('"') % 2 == 1:
            continue
        pieces.append(text[start:match.end()].strip())
        start = match.end()
    if text[start:].strip():
        pieces.append(text[start:].strip())
    tags = {t.lower() for t in join_trailing}
    joined = []
    for piece in pieces:
        if joined and re.sub(r"[^A-Za-z-]", "", piece).lower() in tags:
            joined[-1] = joined[-1] + " " + piece
        else:
            joined.append(piece)
    return joined


def scripted_line(text):
    """The line with its pieces joined by the segment marker. A line break in the
    text (a story.md TEXT that runs onto the next line) ends a piece and is kept,
    so the chat bubble shows it (2026-10-02, operator)."""
    split = load_brobots_settings()["scripted_split"]
    marker = split["segment_marker"]
    lines = [line for line in (text or "").split("\n") if line.strip()]
    return (marker + "\n").join(
        marker.join(scripted_pieces(line, split.get("join_trailing", []), split["piece_ends"])) for line in lines
    )


def run_move_with_line(move, say):
    """2026-10-01 (operator's own Bingo score: "Wait! Do you know..." moved onto
    brat_arm_cue_0201, "It's, "I, Won!"" onto brat_nod_0203, to shorten the
    song): a move step that carries a line does the move WHILE the line is said
    - the move on its own thread, the line here, the step ends when both have.
    Wire-Pod's own chat replies already lift the arm while speaking
    (GOPOD_VECTOR_ARM)."""
    box = {}
    def run():
        try:
            box["move"] = move()
        except Exception as exc:  # noqa: BLE001 - reported in the step's result
            box["move"] = {"ok": False, "error": repr(exc)}
    thread = threading.Thread(target=run, daemon=True)
    thread.start()
    line = say()
    thread.join(timeout=30)
    move_result = box.get("move", {"ok": False, "error": "move did not finish"})
    return {"ok": bool(move_result.get("ok")) and bool(line.get("ok")), "move": move_result, "line": line}


def run_emotion_beat(mod, serial, name, text, animation_token, hold_seconds, live, robots, tag, manage_control=True, scripted_split=False):
    """manage_control=False (STAY_PUT_FIX_SURVEY_001.md, part 3) skips this
    function's own inline assume/release, mirroring run_move_axis()'s already-
    proven bypass exactly - relies on the caller already holding continuous
    control. Default True is byte-identical to every existing call site
    (bingo-family's own manage_control=True songs, and any future caller that
    never passes this kwarg).

    2026-10-01 (operator: "Flip animation beats to animation, ... pause, then
    speech", "aim for native to wire-pod", "adding these extra seconds breaks the
    performance flow"): one say_text in native Wire-Pod's own shape,
    "{{playAnimationWI||angry}}... line" - the animation starts, the voice's own
    lead-in pause ("...", animation_lead_in in gopod_brobots_settings.json), then
    the line. The old say -> 2.0 s pause -> animation order and its
    pre_animation_pause_seconds knob are gone."""
    if not live:
        return {"ok": True, "mode": "dry", "animation_token": animation_token}
    entry = {"animation_token": animation_token}
    try:
        if manage_control:
            assume = mod.wirepod_web_send_form(
                "/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10
            )
            entry["assume"] = assume
            time.sleep(POST_ASSUME_SAY_SETTLE_SECONDS)
        # A beat that carries a real line (2026-10-01: "Excuse Me? ... Round 3!"
        # moved into host_beat_angry_0501) gets the same robot-safe text and chat
        # bubble as a say_turn line; a "," beat stays silent and bubble-free.
        spoken = text
        form = {"serial": serial}
        if re.search(r"[A-Za-z0-9]", text or ""):
            if scripted_split:
                text = scripted_line(text)
                form["sentence_highlight"] = "1"
            spoken, _repair_used = mod.normalize_robot_safe(text, robots.scaffold)
            # A combo part: the step's one bubble, put up by its first speaking part.
            if getattr(robots, "piece_offset", None) is None:
                form.update({"display_text": text, "display_role": name})
            else:
                form["piece_offset"] = str(robots.piece_offset)
                if robots.combo_display:
                    form.update({"display_text": robots.combo_display, "display_role": name})
        # 2026-10-10 (live: Brobot 2 froze on angry with "Wait!"): the clip and
        # the line are sent apart - the clip, chat_moves.speech_after_animation_seconds
        # (0.9), then the line with its lead-in. The banked clean shape
        # (ANGRY_PUN_CRASH_AND_THINKING_MOVES_001.md, operator 2026-10-09: "Use
        # playAnimationWI + 0.9s + start speech"); sent as one is the crash shape.
        anim = mod.wirepod_web_send_form(
            "/api-sdk/say_text", {"serial": serial, "text": f"{{{{playAnimationWI||{animation_token}}}}}"}, timeout=90
        )
        entry["animation_dispatch"] = anim
        time.sleep(float(load_brobots_settings()["chat_moves"].get("speech_after_animation_seconds", 0.9)))
        if re.search(r"[A-Za-z0-9]", text or ""):
            # The robot-safe form is "thinking... words": only the words are said.
            _cue, words = mod.split_robot_safe_line(spoken, robots.scaffold)
            form["text"] = f"{load_animation_lead_in()} {words}"
            said = mod.wirepod_web_send_form("/api-sdk/say_text", form, timeout=90)
            entry["line_dispatch"] = said
            anim = anim if said.get("status") == 200 else said
        print(f"{robots.timing_prefix()}{tag}_BEAT {name} token={animation_token}: animation, then the line {load_brobots_settings()['chat_moves'].get('speech_after_animation_seconds', 0.9)}s later", flush=True)
        time.sleep(0.5)
        # Advisory-only, non-blocking - see run_songs_runner_001.py's own long
        # comment / BINGO_STUCK_ANIMATION_FALSE_POSITIVE_RESOLUTION_001.md for
        # why this never flips entry["ok"].
        animation_wait_line_seen = False
        try:
            debug_log_body = mod.wirepod_get_text("/api/get_debug_logs", timeout=3)
            if ANIMATION_WAIT_LOG_SIGNATURE in debug_log_body[-4000:]:
                animation_wait_line_seen = True
        except Exception:  # noqa: BLE001 - best-effort diagnostic only
            pass
        entry["animation_wait_advisory"] = animation_wait_line_seen
        print(f"{robots.timing_prefix()}{tag}_BEAT {name} token={animation_token}: animation dispatched, holding {hold_seconds}s", flush=True)
        time.sleep(hold_seconds)
        entry["ok"] = anim.get("status") == 200
    except Exception as exc:  # noqa: BLE001 - caller decides whether to stop the run
        entry["ok"] = False
        entry["error"] = repr(exc)
    finally:
        if manage_control:
            release = mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)
            entry["release"] = release
            time.sleep(1.0)
    return entry


def run_animation_only(mod, serial, animation_token, hold_seconds, live, manage_control=True):
    """manage_control=False bypass, same shape/reason as run_emotion_beat()
    above (STAY_PUT_FIX_SURVEY_001.md, part 3)."""
    if not live:
        return {"ok": True, "mode": "dry", "animation_token": animation_token}
    entry = {"animation_token": animation_token}
    try:
        if manage_control:
            assume = mod.wirepod_web_send_form(
                "/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10
            )
            entry["assume"] = assume
            time.sleep(POST_ASSUME_SAY_SETTLE_SECONDS)
        anim_payload = f"{{{{playAnimationWI||{animation_token}}}}}"
        dispatches = []
        getout_token = _ANIMATION_GETOUT_TOKENS.get(animation_token)
        if getout_token:
            # GOLDEN, LOCKED - searching: one send, the step's hold, then the getout.
            dispatches.append(
                mod.wirepod_web_send_form("/api-sdk/say_text", {"text": anim_payload, "serial": serial}, timeout=10)
            )
            time.sleep(hold_seconds)
            dispatches.append(
                mod.wirepod_web_send_form(
                    "/api-sdk/say_text", {"text": f"{{{{playAnimationWI||{getout_token}}}}}", "serial": serial}, timeout=10
                )
            )
            time.sleep(_ANIMATION_GETOUT_SETTLE_SECONDS)
        elif animation_token in _LOOP_ANIMATION_TOKENS:
            elapsed = 0.0
            while elapsed < hold_seconds:
                dispatches.append(
                    mod.wirepod_web_send_form("/api-sdk/say_text", {"text": anim_payload, "serial": serial}, timeout=10)
                )
                time.sleep(_ANIMATION_LOOP_INTERVAL_SECONDS)
                elapsed += _ANIMATION_LOOP_INTERVAL_SECONDS
            # GOLDEN, LOCKED - a loop with no getout clip (answering) settles.
            time.sleep(_ANIMATION_NO_GETOUT_SETTLE_SECONDS)
        else:
            dispatches.append(
                mod.wirepod_web_send_form("/api-sdk/say_text", {"text": anim_payload, "serial": serial}, timeout=10)
            )
            time.sleep(hold_seconds)
        entry["dispatches"] = dispatches
        entry["ok"] = all(d.get("status") == 200 for d in dispatches)
    except Exception as exc:  # noqa: BLE001 - print-loudly-and-continue note, same as arm_cue/nod/rattle
        entry["ok"] = False
        entry["error"] = repr(exc)
    finally:
        if manage_control:
            release = mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)
            entry["release"] = release
    return entry


# Note types whose failure just warns-and-continues (generic message, note
# name substituted in) - mirrors run_songs_runner_001.py's own per-note
# warning text exactly, which was already the same sentence for all five.
_WARN_ON_FAILURE_NOTES = {"rattle", "brobots_ready_together", "arm_cue", "nod", "animation", "picture"}


# 2026-10-13 (operator: "add a check that warns when a step's name doesn't match
# its knobs parts * also, how to handle syntax errors, within the scope of the
# step-names list?"): a step name is a label (ALIAS-PLAY-STUDIO.md, "Step names,
# the master list") - this only warns, never stops a run. A name ending in
# _SSNN is read as <who>_<part>_<part>..._SSNN and its parts are compared, in
# order, to what its knobs.json entry plays; other names (the nap's line_30,
# soft_wake...) are free labels and aren't read.
STEP_NAME_WHO = {"host", "brat", "brobots", "reporter", "goverlord"}
STEP_NAME_NOTES = {"say": "say_turn", "nod": "nod", "rattle": "rattle", "gap": "pause",
                   "wake": "wake_both", "ready": "brobots_ready_together",
                   "sleep": "sleep_both", "exit": "exit",
                   # 2026-10-16 (operator: "extend that to include sequences,
                   # animations, timers, llm warm-up & in-between reporter
                   # lines"): timer_ (a timed wait), sequence_ (another song's
                   # steps, by song_id), and the LLM warm notes (lines_, sting_;
                   # warm_ / warm_wait_ / warm_ready_ below) - the warm notes are
                   # played by Wire-Pod (gopod_opener_song.go).
                   "timer": "timer", "sequence": "sequence", "lines": "reporter_lines",
                   "sting": "sting"}
STEP_NAME_THINKING = [("thinking", "")]
# 2026-10-13 (operator: "Add \"danger\" + new wavs to the master list * grunge &
# djent3b"): the shared sounds, by name (sounds/danger-will-robinson.wav,
# grunge.wav, dJent3b.wav).
STEP_NAME_SOUNDS = {"danger": "danger", "grunge": "grunge", "djent3b": "djent3b"}
_STEP_NAME_RE = re.compile(r"^(.*)_(\d{4})$")


def _step_plays(raw_step):
    """What a knobs.json step plays: [(note, detail)], detail the animation
    token, sound or picture."""
    parts = raw_step.get("combo") or [raw_step]
    out = []
    for part in parts:
        note = part.get("note", "")
        detail = part.get("animation_token") or part.get("sound") or part.get("picture") or ""
        if isinstance(detail, list):
            detail = ",".join(detail)
        out.append((note, str(detail)))
    return out


def step_name_problem(step_id, plays):
    """Why a step's name doesn't match what it plays, or ""."""
    m = _STEP_NAME_RE.match(step_id or "")
    if not m:
        return ""
    words = [w for w in m.group(1).split("_") if w]
    want = []  # (note, detail or "") per name part; "?word" = an old sound label
    i = 0
    while i < len(words):
        w, nxt = words[i].lower(), (words[i + 1].lower() if i + 1 < len(words) else "")
        if w in STEP_NAME_WHO:
            i += 1
        elif w == "arm" and nxt == "cue":
            want.append(("arm_cue", "")); i += 2
        elif w == "llm" and nxt == "song":
            # 2026-10-20: llm_song_ - the folder's _llm_song run (gopod_llm_song.go).
            want.append(("llm_song", "")); i += 2
        elif w == "warm" and nxt in ("wait", "ready"):
            want.append(("warm_" + nxt, "")); i += 2
        elif w == "warm":
            want.append(("warm_message", "")); i += 1
        elif w.startswith("anim-") and len(w) > 5:
            # 2026-10-16: anim-<token>, an animation by its token (anim-kgsuccess).
            want.append(("animation", w[5:])); i += 1
        elif w == "kg" and nxt == "success":
            want.append(("animation", "kgsuccess")); i += 2
        elif w == "beat" and nxt:
            want.append(("emotion_beat", nxt)); i += 2
        elif nxt == "animation":
            want.append(("animation", w)); i += 2
        elif w == "thinking":
            want.extend(STEP_NAME_THINKING); i += 1
        elif w.startswith("wav-") and len(w) > 4:
            want.append(("sound", w[4:])); i += 1
        elif w.startswith("img-") and len(w) > 4:
            want.append(("picture", w[4:])); i += 1
        elif w in STEP_NAME_SOUNDS:
            want.append(("sound", STEP_NAME_SOUNDS[w])); i += 1
        elif w in STEP_NAME_NOTES:
            want.append((STEP_NAME_NOTES[w], "")); i += 1
        else:
            want.append(("?", w)); i += 1
    # "say" names the step's line; when another part says it (a beat, a nod...)
    # and there's no say_turn of its own, the name's "say" is that line.
    if ("say_turn", "") in want and not any(n == "say_turn" for n, _ in plays):
        want.remove(("say_turn", ""))
    unknown = [d for n, d in want if n == "?"]
    match = len(want) == len(plays)
    if match:
        for (wn, wd), (pn, pd) in zip(want, plays):
            pd_low = pd.lower()
            if wn == "?":
                # The older form: a plain word for a sound (fart2 -> fart2.wav).
                ok = pn == "sound" and wd in pd_low
            elif wn in ("sound", "picture"):
                ok = pn == wn and wd in pd_low
            elif wd:
                ok = pn == wn and wd == pd_low.replace("_", "")
            else:
                ok = pn == wn
            if not ok:
                match = False
                break
    if match:
        return ""
    said = " + ".join(f"{n}:{d}" if d else n for n, d in plays)
    read = " + ".join((f"{d}?" if n == "?" else f"{n}:{d}" if d else n) for n, d in want)
    note = f" - {', '.join(repr(u) for u in unknown)} isn't on the master list" if unknown else ""
    return f"{step_id}: the name reads {read}, its knobs play {said}{note}"


def step_name_warnings(song_dir):
    """The step-name problems in a song's knobs.json (empty: all match)."""
    try:
        _, knobs = load_knobs_envelope(song_dir)
    except Exception:  # noqa: BLE001 - a broken knobs.json is reported elsewhere
        return []
    out = []
    for raw in knobs.get("steps", []):
        problem = step_name_problem(raw.get("step_id", ""), _step_plays(raw))
        if problem:
            out.append(problem)
    return out


def _maybe_promote_knobs(song_dir):
    """End-of-run promote prompt (operator direction, 2026-08-11): zKnobs.json
    is the workbench, knobs.json only updates on explicit operator say-so -
    divergence between the two is normal, never auto-reconciled. Only offered
    when a zKnobs.json actually exists (i.e. this run played a dirty file) -
    no zKnobs.json means knobs.json was already what played, nothing to
    promote. Copy is byte-consistent (shutil.copyfile, no json re-serialize)
    so the promoted knobs.json is identical to the file that was just run."""
    zknobs_path = Path(song_dir) / "zKnobs.json"
    knobs_path = Path(song_dir) / "knobs.json"
    if not zknobs_path.exists():
        return
    try:
        choice = input("update knobs.json from this run's zKnobs.json for commit? y/n [default n]: ").strip().lower()
    except (EOFError, KeyboardInterrupt):
        choice = "n"
    if choice == "y":
        shutil.copyfile(zknobs_path, knobs_path)
        print(f"KNOBS_PROMOTED knobs.json updated from {zknobs_path}")
    else:
        print("KNOBS_PROMOTE_SKIPPED knobs.json left untouched")


def run_golden_song(song_dir, live, voice_destination="robot"):
    run_stamp = time.strftime("%Y-%m-%d_%H-%M-%S")
    runs_dir = Path(song_dir) / "runs"
    runs_dir.mkdir(parents=True, exist_ok=True)
    log_path = runs_dir / f"golden_run_{run_stamp}.log"
    json_path = runs_dir / f"golden_run_{run_stamp}.json"
    log_file = open(log_path, "w", encoding="utf-8")
    real_stdout = sys.stdout
    tee = _TeeStdout(real_stdout, log_file)
    sys.stdout = tee

    def _finalize(log, tag):
        timed_steps = [s for s in log["steps"] if "step_wall_seconds" in s]
        if timed_steps:
            gap_report = sorted(timed_steps, key=lambda s: s["step_wall_seconds"], reverse=True)
            log["gap_report"] = [
                {"step_id": s["step_id"], "note": s["note"], "step_wall_seconds": s["step_wall_seconds"]}
                for s in gap_report
            ]
            print("")
            print(f"===== {tag}_GAP_REPORT (largest to smallest) =====")
            for rank, s in enumerate(gap_report, start=1):
                print(
                    f"{tag}_GAP_REPORT rank={rank} step_id={s['step_id']} "
                    f"note={s['note']} seconds={s['step_wall_seconds']:.3f}"
                )
        tee.capture_file = False
        print(json.dumps(log, indent=2))
        json_path.write_text(json.dumps(log, indent=2), encoding="utf-8")
        sys.stdout = real_stdout
        log_file.close()
        print(f"{tag}_SAVED log={log_path} json={json_path}")
        return log

    touch_off_file = None
    run_name, checked_in, log = None, None, None
    try:
        # 2026-10-17 (operator: "make TOUCH_INTERRUPT optional y/n"): pha0b asks;
        # GOPOD_GOLDEN_TOUCH_INTERRUPT_OFF=0 leaves it on for this run.
        if live and os.getenv("GOPOD_GOLDEN_TOUCH_INTERRUPT_OFF", "1") == "0":
            print("TOUCH_INTERRUPT left on for this run (pha0b: n)", flush=True)
        elif live and load_golden_song(song_dir, _load_control_song_module()).get("touch_interrupt_off"):
            touch_off_file = touch_interrupt_off_path(_load_runner_module())
            touch_off_file.touch()
            print(f"TOUCH_INTERRUPT off for this run ({touch_off_file})", flush=True)
        song = load_golden_song(song_dir, _load_control_song_module())
        for warning in step_name_warnings(song_dir):
            print(f"STEP_NAME_WARNING {warning}", flush=True)
        # 2026-10-10, GOLDEN (gopod_session.go, operator: "Goverlord is the
        # wpr-session orchestrator"): a live run checks in with Goverlord - one
        # run at a time - and out when it ends, however it ends.
        if live:
            run_name = song.get("chat_intro") or song.get("song_id") or Path(song_dir).name
            checked_in = session_run(_load_runner_module(), "start", name=run_name, kind="song", pid=str(os.getpid()))
            if checked_in is False:
                print(f"GOLDEN_REFUSED {run_name!r}: Goverlord says another run is going (or it isn't LLM Ready yet).", flush=True)
                tag = _tag_for(song.get("song_id"))
                return _finalize({"song_id": song.get("song_id"), "live": live, "voice_destination": voice_destination,
                                  "steps": [], "stopped_early": True, "refused": True}, tag)
        # 2026-10-17 (operator: "both the terminal run & web page run must be the
        # same run"): a live run starts only once the chat page is here.
        if live:
            wait_for_page(_load_runner_module(), "before the run")
        if live and song.get("llm_warmup") and not song.get("warmup_after"):
            song_llm_warmup(_load_runner_module(), song.get("llm_ready_text", ""), song.get("warmup_lines", []),
                            song.get("voices_required", True), song.get("llm_required", True))
        log = _run_golden_song_steps(song_dir, live, voice_destination, _finalize)
    finally:
        # A sleepy run that ended early: wake the Brobots, never leave them held asleep.
        global _SLEEPY
        if _SLEEPY is not None:
            _SLEEPY.wake_all()
            _SLEEPY = None
        # Checked out last, once the Brobots are handed back.
        if run_name is not None and checked_in:
            session_run(_load_runner_module(), "end", name=run_name,
                        reason="stopped" if (log or {}).get("stopped_early", True) else "done")
        if touch_off_file is not None:
            touch_off_file.unlink(missing_ok=True)
            print("TOUCH_INTERRUPT back on", flush=True)
        if sys.stdout is tee:
            sys.stdout = real_stdout
        if not log_file.closed:
            log_file.close()
    # AWAKEN_MULTI_MODE_TWO_PASS_SURVEY_PLAN_001.md follow-up (operator
    # request after a live two-pass run): pha0b's own bait "both" mode calls
    # this whole function twice, once per robot - without a suppress flag,
    # the promote prompt asked (and blocked on) the SAME pending zKnobs.json
    # diff twice, once after each pass, with brobot 2's pass sitting behind
    # that second interactive wait. GOPOD_GOLDEN_SUPPRESS_PROMOTE, exported
    # by brobots.sh for every pass except the last one in a multi-pass run,
    # skips this call entirely for those passes - the prompt now fires
    # exactly once, after the final pass, with nothing interactive between
    # brobot 1 finishing and brobot 2 starting.
    if live and not log.get("stopped_early", True) and not os.getenv("GOPOD_GOLDEN_SUPPRESS_PROMOTE"):
        _maybe_promote_knobs(song_dir)
    return log


def _run_golden_song_steps(song_dir, live, voice_destination, _finalize):
    mod = _load_runner_module()
    control_mod = _load_control_song_module()
    song = load_golden_song(song_dir, control_mod)
    song_id = song["song_id"]
    tag = _tag_for(song_id)
    # Tempo buffer knob (TEMPO_BUFFER_KNOB_SURVEY_001.md, Phase 1) - a per-song
    # global multiplier, 0.0-9.9, absent/0.0 means every buffer's added_space
    # below is always 0, i.e. no-op by construction until a knobs.json
    # actually sets this above 0.0.
    global_tempo = song.get("global_tempo", 0.0)

    config = SONG_REGISTRY.get(song_id)
    if config is None:
        print(f"GOLDEN_REFUSED song_id={song_id!r} not yet registered - registered: {', '.join(sorted(SONG_REGISTRY))}", flush=True)
        log = {"song_id": song_id, "live": live, "voice_destination": voice_destination, "steps": [], "stopped_early": True}
        return _finalize(log, tag)
    if not config.get("golden_ready", True):
        real_runner = config.get("runner", "its own runner")
        print(
            f"GOLDEN_REFUSED song_id={song_id!r} is on the roster but not golden-dispatch-ready yet "
            f"(different execution shape) - run it through {real_runner} directly.",
            flush=True,
        )
        log = {"song_id": song_id, "live": live, "voice_destination": voice_destination, "steps": [], "stopped_early": True}
        return _finalize(log, tag)
    fatal_notes = config["fatal_notes"]
    manage_control = config.get("manage_control", True)
    note_aliases = config.get("note_aliases", {})
    unconditional_text_speak = config.get("unconditional_text_speak", False)
    suppress_assume_release = not manage_control

    # synthesize_speaker=True songs (control-family): every step's own
    # knobs.json "speaker" field is unused/dead - the real target robot is a
    # whole-run choice (GOLDEN_ROBOT_ENV, "1"/"2", default "1"), matching
    # run_robot_control_song_001.py's own GOPOD_CONTROL_SONG_ROBOT exactly.
    # No-op for every other song - robot_param/synthesized_speaker stay None,
    # never read.
    robot_param = None
    synthesized_speaker = None
    if config.get("synthesize_speaker"):
        robot_param = os.getenv(GOLDEN_ROBOT_ENV, "1")
        if robot_param not in ("1", "2"):
            print(f"GOLDEN_REFUSED {GOLDEN_ROBOT_ENV}={robot_param!r} must be '1' or '2'", flush=True)
            log = {"song_id": song_id, "live": live, "voice_destination": voice_destination, "steps": [], "stopped_early": True}
            return _finalize(log, tag)
        synthesized_speaker = "brobot_2" if robot_param == "2" else "brobot_1"

    # Run-time robot playback filter - resolved once, up front. Validated the
    # same way GOLDEN_ROBOT_ENV is above (refuse cleanly, don't guess).
    # "both" is the default and is a complete no-op: robot_filter != "both"
    # gates every actual skip check in the per-step loop below, so a run
    # with this env var unset is byte-identical to before this feature
    # existed.
    robot_filter = os.getenv(GOLDEN_ROBOT_FILTER_ENV, "both")
    if robot_filter not in ("both", "1", "2"):
        print(f"GOLDEN_REFUSED {GOLDEN_ROBOT_FILTER_ENV}={robot_filter!r} must be 'both', '1', or '2'", flush=True)
        log = {"song_id": song_id, "live": live, "voice_destination": voice_destination, "steps": [], "stopped_early": True}
        return _finalize(log, tag)

    # Stay-put fix (STAY_PUT_FIX_SURVEY_001.md, part 1): a manage_control=False
    # song with synthesize_speaker=False is the NEW bingo-family continuous-
    # dual-robot-hold combination (brobots_bingo/cross_persona/baby_robots_sleep)
    # - distinct from awaken's existing manage_control=False + synthesize_speaker=
    # True (one whole-run robot) combination. This flag gates every stay-put
    # addition below (the dual-robot assume off wake_both, the emotion_beat/
    # animation bypass, the rattle/brobots_ready_together re-assume) - False for
    # every other combination, including awaken, which this fix never touches.
    continuous_hold_both = not manage_control and synthesized_speaker is None

    # manage_control=False's own release target(s) - resolved once, up front,
    # since a run can exit early (before the per-step loop ever assigns a local
    # `serial`) and still needs to hand control back. Robots.close() itself is
    # a no-op (never releases control) - see run_section1_full_live_001.py - so
    # this song's own final release, which the original always fires in its
    # own `finally` block, must be done explicitly here instead. No-op for
    # manage_control=True songs regardless of this value. Generalized from a
    # single serial to a tuple (part 1) - awaken's own combination still
    # resolves exactly one; the new continuous_hold_both combination resolves
    # both robot serials, since either can be addressed at any step.
    if continuous_hold_both:
        hold_release_serials = (mod.BROBOT_1_SERIAL, mod.BROBOT_2_SERIAL)
    elif synthesized_speaker:
        hold_release_serials = (mod.BROBOT_2_SERIAL if synthesized_speaker == "brobot_2" else mod.BROBOT_1_SERIAL,)
    else:
        hold_release_serials = ()

    full_steps = song["steps"]
    playhead_from_id = os.getenv(GOLDEN_PLAYHEAD_FROM_ENV)
    playhead_to_id = os.getenv(GOLDEN_PLAYHEAD_TO_ENV)
    if playhead_from_id or playhead_to_id:
        all_step_ids = [s["step_id"] for s in full_steps]
        playhead_error = None
        if playhead_from_id and playhead_from_id not in all_step_ids:
            playhead_error = f"{GOLDEN_PLAYHEAD_FROM_ENV}={playhead_from_id!r} is not a known step_id."
        elif playhead_to_id and playhead_to_id not in all_step_ids:
            playhead_error = f"{GOLDEN_PLAYHEAD_TO_ENV}={playhead_to_id!r} is not a known step_id."
        else:
            from_index = all_step_ids.index(playhead_from_id) if playhead_from_id else 0
            # A combo's parts share its step_id: a range ending on it plays them all.
            to_index = (len(all_step_ids) - 1 - all_step_ids[::-1].index(playhead_to_id)) if playhead_to_id else len(all_step_ids) - 1
            if to_index < from_index:
                playhead_error = (
                    f"{GOLDEN_PLAYHEAD_TO_ENV}={playhead_to_id!r} comes before "
                    f"{GOLDEN_PLAYHEAD_FROM_ENV}={playhead_from_id!r} in the score's own step order."
                )
        if playhead_error:
            print(f"GOLDEN_STOP playhead: {playhead_error}", flush=True)
            log = {"song_id": song_id, "live": live, "voice_destination": voice_destination, "steps": [], "stopped_early": True}
            return _finalize(log, tag)
        playhead_steps = full_steps[from_index:to_index + 1]
        print(
            f"{tag}_PLAYHEAD range={all_step_ids[from_index]}..{all_step_ids[to_index]} "
            f"steps={len(playhead_steps)}",
            flush=True,
        )
        song = {**song, "steps": playhead_steps}

    target_section = os.getenv(GOLDEN_SECTION_ENV)
    if target_section:
        filtered_steps = [s for s in song["steps"] if s.get("section") == target_section]
        print(f"{tag}_SECTION_FILTER active target={target_section!r} steps={len(filtered_steps)}", flush=True)
        if not filtered_steps:
            print(f"{tag}_SECTION_FILTER_EMPTY target={target_section!r} - no steps matched that section name.", flush=True)
        song = {**song, "steps": filtered_steps}

    scaffold = mod.load_interview_scaffold(mod.INTERVIEW_SCAFFOLD_PATH)
    robots = mod.Robots(live, scaffold, voice_destination, console_rich_display=True, console_timestamps=True)
    # 2026-09-30: each line is said one sentence at a time so the chat page
    # lights the sentence being said (sdkapp say_text sentence_highlight).
    robots.sentence_highlight = True
    # knobs.json "sleepy": true (the nap): the Brobots sleep, and lull out for their moves.
    global _SLEEPY
    sleepy = SleepyBrobots(mod, live, robots, tag, song.get("wake_clip", "")) if song.get("sleepy") else None
    _SLEEPY = sleepy

    startup_line = f"LIVE: Wire-Pod ESN routing enabled: {mod.WIREPOD_BASE_URL}" if live else "DRY: live speech gate is off."
    print(f"{robots.timing_prefix()}{startup_line}")

    log = {"song_id": song_id, "live": live, "voice_destination": voice_destination, "steps": [], "stopped_early": False}

    if os.getenv(GOLDEN_PLAYHEAD_WAKE_ENV) == "1":
        wake_step = next((s for s in full_steps if s["note"] == "wake_both"), None)
        wake_settle = wake_step.get("settle_seconds", 1.5) if wake_step else 1.5
        print(f"{robots.timing_prefix()}{tag}_PLAYHEAD_WAKE prep: running wake_both before the chosen range", flush=True)
        wake_result = run_wake_both(mod, live, wake_settle)
        log["playhead_wake_prep"] = wake_result
        if not wake_result["ok"]:
            print(f"{robots.timing_prefix()}{tag}_STOP playhead_wake: {WIREPOD_PAIRING_RECOVERY_MESSAGE}", flush=True)
            log["stopped_early"] = True
            _release_if_holding(mod, hold_release_serials, live, manage_control)
            robots.close()
            return _finalize(log, tag)

    if live and os.getenv(GOLDEN_CONNECTION_PASS_ENV) == "1":
        connection_pass = run_connection_pass(mod, live, robots, tag)
        log["connection_pass"] = connection_pass
        if not connection_pass["ok"]:
            log["stopped_early"] = True
            _release_if_holding(mod, hold_release_serials, live, manage_control)
            robots.close()
            return _finalize(log, tag)

    # 2026-09-30: a song can open the chat page with an operator bubble
    # (knobs.json "chat_intro", e.g. bingo's "Bingo Test Live") - on a full run
    # only, not a playhead slice that starts mid-song. Not when the run was
    # started by voice (GOPOD_CHAT_INTRO_SPOKEN=1, set by the "bingo test live"
    # intent): the operator's own spoken line is already that bubble, and only
    # one should show (operator, 2026-09-30).
    if (song.get("chat_intro") and song["steps"] and full_steps and song["steps"][0] is full_steps[0]
            and os.getenv("GOPOD_CHAT_INTRO_SPOKEN") != "1"):
        log["chat_intro"] = post_chat_line(mod, live, "Operator", song["chat_intro"], robots)

    def say_step_line(name, text):
        """A step's spoken line: the scripted split and the silent-line heal
        (run_say_turn) - the say_turn note's own path, also used by a move or
        rattle step that carries a line."""
        say_text = scripted_line(text) if song.get("sentence_split") == "scripted" else text
        return run_say_turn(mod, robots, name, say_text, suppress_assume_release=suppress_assume_release)

    current_section = None
    # Robots whose control is retaken before their next step (retake_robot_control).
    needs_control_retake = set()
    stop_request = song_stop_request_path(mod)
    if live:
        stop_request.unlink(missing_ok=True)  # an old request never stops a new run
    # 2026-10-09 (the nap, a lyric video): "timeline": "srt" - each step with a
    # start_seconds waits for that second, counted from the first timed step,
    # instead of following the step before it. A step that's late goes at once.
    timeline_start = None
    soundtrack_posted = False
    captions_stop = threading.Event()
    cube_job = None
    cube_light = None  # a wake_sleepy's "cube_ready" (start_cube_light)
    for step in song["steps"]:
        step_id = step["step_id"]
        note = note_aliases.get(step["note"], step["note"])
        # A background cube connect (a wake step's "cube_connect") finishes
        # before the next step that moves a robot; pauses and Goverlord's lines
        # go on beside it.
        if cube_job is not None and note not in ("pause", "system_line"):
            cube_result = finish_cube_connect(cube_job)
            log["cube_connect"] = cube_result
            print(f"{robots.timing_prefix()}{tag}_CUBE_CONNECT {cube_result.get('serial')} ok={cube_result['ok']} "
                  f"waited={cube_result.get('waited_seconds', 0)}s {cube_result.get('output', cube_result.get('error', ''))}", flush=True)
            cube_job = None
        if song.get("timeline") == "srt" and step.get("start_seconds") is not None:
            at = float(step["start_seconds"])
            if timeline_start is None:
                timeline_start = time.time() - at
                # 2026-10-09 (the nap, operator: "playing the video's audio ...
                # Go"; "having it run in the chat bubbles"): the soundtrack starts
                # on the chat page, and the captions post as the narrator's
                # bubbles on their own clock, so a slow move never holds them.
                if song.get("audio"):
                    timeline_start = _start_song_soundtrack(mod, live, song, at, robots, tag)
                    soundtrack_posted = True
                _start_captions(mod, live, song, timeline_start, captions_stop, robots)
            elif song.get("audio") and not soundtrack_posted:
                # The timeline was set by a "timeline_after" pause: its captions
                # are already on their way; the soundtrack starts now, on time.
                _start_song_soundtrack(mod, live, song, at, robots, tag)
                soundtrack_posted = True
            wait = timeline_start + at - time.time()
            if live and wait > 0:
                time.sleep(wait)
            print(f"{robots.timing_prefix()}{tag}_TIMELINE {step_id} at={at:.3f}s {'late ' + format(-wait, '.3f') + 's' if wait < 0 else 'on time'}", flush=True)
        # 2026-10-12 (operator: "Upon page unfocus = pause at next safe spot until
        # focus, then continue"): while the chat page is hidden, wait here (a timed
        # song plays on with its soundtrack); Goverlord says paused / continues.
        if live and song.get("timeline") != "srt":
            wait_for_page(mod, f"before {step_id}", stop_request, robots.timing_prefix())
        # The next safe spot: between steps, never in the middle of one.
        if live and stop_request.exists():
            stop_request.unlink(missing_ok=True)
            print(f"{robots.timing_prefix()}{tag}_STOP {step_id}: stopped by voice (kill bingo test) before this step", flush=True)
            log["stopped_early"] = True
            log["stopped_by_voice"] = step_id
            captions_stop.set()
            # 2026-10-15 (operator: "Stop run didn't work during song run * this
            # song is different"): a timed song's soundtrack and narration on the
            # chat page stop with it ("stop" on the song-audio line).
            if soundtrack_posted:
                role = load_brobots_settings()["reporter_voices"].get("song_audio_role", "")
                if role:
                    post_chat_line(mod, live, role, "stop", robots)
            _release_if_holding(mod, hold_release_serials, live, manage_control)
            robots.close()
            return _finalize(log, tag)
        speaker = synthesized_speaker if synthesized_speaker else (step.get("speaker") or "brobot_1")
        serial, name = _serial_and_name(mod, speaker)
        entry = {"step_id": step_id, "note": note, "speaker": speaker}
        # Run-time robot playback filter (pha0b's "run which robots?" prompt) -
        # never fires for note in _FILTER_EXEMPT_NOTES or when robot_filter is
        # "both" (today's default, byte-identical). Otherwise skips dispatch
        # for any step whose resolved speaker is the OTHER robot.
        step_filtered = (
            robot_filter != "both"
            and note not in _FILTER_EXEMPT_NOTES
            and speaker != _ROBOT_FILTER_SPEAKER[robot_filter]
        )

        section = step.get("section")
        if section and section != current_section:
            print("", flush=True)
            print("=====", flush=True)
            print(f"{robots.timing_prefix()}=== {tag} SECTION: {section} ===", flush=True)
            print("=====", flush=True)
            print("", flush=True)
            current_section = section

        display_tag = _derive_note_tag(step)
        # Standardized numbered step_id (section+beat) makes every step's
        # divider unique on its own - print one per step, not deduped by
        # tag, so the id is never silently dropped for a repeated tag.
        print(f"===== {step_id} Notes: {display_tag}", flush=True)
        step_start_wall = time.time()
        print(f"{robots.timing_prefix()}{tag} {step_id}: note={note} speaker={speaker}", flush=True)

        if serial in needs_control_retake and note in CONTROL_RETAKE_NOTES and not step_filtered:
            needs_control_retake.discard(serial)
            entry["control_retake"] = retake_robot_control(mod, serial, name, live, robots, tag)

        # 2026-10-10 (operator: "Should be combined!!"): a combo step is one chat
        # bubble. The first part that speaks puts the whole bubble up with its
        # voice; every part lights its own pieces in it.
        robots.piece_offset = None
        robots.combo_display = None
        if live and step.get("combo_part") and not step_filtered and song.get("sentence_split") == "scripted":
            marker = load_brobots_settings()["scripted_split"]["segment_marker"]
            texts = step["combo_texts"]
            pieces_of = [len(scripted_line(t).split(marker)) if t.strip() else 0 for t in texts]
            k = step["combo_part"] - 1
            if texts[k].strip() and not any(t.strip() for t in texts[:k]):
                robots.combo_display = marker.join(scripted_line(t) for t in texts if t.strip())
            robots.piece_offset = sum(pieces_of[:k])

        # A step with "show_ball" (the Bingo test's ball calls) puts the big ball
        # in the Brobots window as the step starts (post_bingo_ball).
        if step.get("show_ball") and not step_filtered:
            entry["show_ball"] = post_bingo_ball(mod, live, step["show_ball"], robots)

        # unconditional_text_speak (control-family only, no-op default False):
        # this song's own text is spoken BEFORE its note-specific action,
        # regardless of note type - not gated to a dedicated say_turn note the
        # way bingo-family works. "{robot_name}" is the one substitution
        # token this family supports (run_robot_control_song_001.py's own
        # say_line() caller). Result discarded, matching the original
        # exactly - a text-speak failure here never affects this step's own
        # entry["ok"], which is decided entirely by the note branch below.
        if unconditional_text_speak and step.get("text") and not step_filtered:
            run_say_turn(
                mod, robots, name, step["text"].replace("{robot_name}", name),
                suppress_assume_release=suppress_assume_release,
            )

        if step_filtered:
            entry["ok"] = True
            entry["filtered"] = True
            entry["filtered_robot_target"] = speaker
            print(f"{robots.timing_prefix()}{tag}_FILTERED {step_id}: note={note} speaker={speaker} skipped (robot_filter={robot_filter})", flush=True)

        elif note == "wake_both":
            result = run_wake_both(mod, live, step.get("settle_seconds", 1.5))
            entry["ok"] = result["ok"]
            entry["result"] = result
            # Stay-put fix, part 2: the run-level dual-robot grab for the new
            # continuous_hold_both combination, fired once, right after
            # wake_both's own conn_test succeeds (no "connect" note exists in
            # this song family's own vocabulary - this is the equivalent
            # hook). A failed grab fails this step too, so wake_both's own
            # fatal_notes entry stops the run the same way a failed conn_test
            # already does - nothing downstream would work correctly without
            # it anyway.
            # A sleepy song holds nothing here: sleep_both comes next. It takes
            # the Brobots from the chat's stay-in-place hold the Wire-Pod way
            # (an assume ends that hold), hands them back and settles, so the
            # sleep tool is the only control on them (live 2026-10-10 07:40:
            # the chat hold was still on through the whole nap).
            if sleepy and entry["ok"] and live:
                for hand_serial in (mod.BROBOT_1_SERIAL, mod.BROBOT_2_SERIAL):
                    mod.wirepod_web_send_form("/api-sdk/assume_behavior_control", {"priority": "high", "serial": hand_serial}, timeout=10)
                for hand_serial in (mod.BROBOT_1_SERIAL, mod.BROBOT_2_SERIAL):
                    mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": hand_serial}, timeout=10)
                time.sleep(DIRECT_SDK_RELEASE_SETTLE_SECONDS)
                entry["chat_hold_handed_off"] = True
            if continuous_hold_both and entry["ok"] and not sleepy:
                hold_assume = run_continuous_hold_assume(mod, live)
                entry["continuous_hold_assume"] = hold_assume
                entry["ok"] = entry["ok"] and hold_assume["ok"]
            # "cube_connect": "brobot_2" (2026-10-14, cube option 2): his cube
            # connects in the background while the robots are idle, so the
            # cube_after_wake blip later lights at once.
            if step.get("cube_connect") and entry["ok"]:
                cube_job = start_cube_connect(live, mod.BROBOT_2_SERIAL if step["cube_connect"] == "brobot_2" else mod.BROBOT_1_SERIAL)
                entry["cube_connect"] = "started"

        elif note == "sleep_both":
            # The nap (knobs.json "sleepy": true): both Brobots fall asleep, then
            # settle_seconds for GoToSleepGetIn and GoToSleepSleeping to play.
            if sleepy is None:
                entry["ok"] = False
                entry["error"] = "sleep_both needs \"sleepy\": true in knobs.json"
            else:
                sleepy.sleep_both(float(step.get("look_seconds", 0)) if live else 0)
                if live:
                    time.sleep(step.get("settle_seconds", 6.0))
                entry["ok"] = True

        elif sleepy and sleepy.asleep and note in SLEEPY_MOVE_NOTES and not step.get("text"):
            # Sleepy: the move goes to its Brobot's worker, which lulls out, moves
            # and drifts back to sleep on its own time (SleepyBrobots).
            if note == "arm_cue":
                move = lambda serial=serial, step=step: control_mod.run_arm_cue(
                    mod, serial, step.get("cycles", 1), step.get("hold_seconds", 1.2), live,
                    speed=step.get("speed", 2), manage_control=False)
            elif note == "nod":
                move = lambda serial=serial, step=step: control_mod.run_nod(
                    mod, serial, step.get("hold_seconds", 0.35), live, speed=step.get("speed", 2), manage_control=False)
            else:
                move = lambda serial=serial, step=step: run_animation_only(
                    mod, serial, step.get("animation_token"), step.get("hold_seconds", 2.5), live, manage_control=False)
            sleepy.move(serial, step_id, move)
            entry["ok"] = True
            entry["sleepy"] = "handed to " + name

        elif note == "stir":
            # 2026-10-09 (operator: "brobots stay asleep the whole song, only
            # waking up to go back to sleep, no other animations"): the Brobot
            # wakes, stays awake hold_seconds, and drifts back to sleep, on its
            # own time (SleepyBrobots).
            if sleepy is None or not sleepy.asleep:
                entry["ok"] = False
                entry["error"] = "stir needs sleep_both first"
            else:
                sleepy.move(serial, step_id, None, awake=step.get("hold_seconds", 1.0))
                entry["ok"] = True
                entry["sleepy"] = "stir " + name

        elif note == "wake_sleepy":
            # 2026-10-12 (the nap, operator: "start soft wake call at start of
            # \"I'm drawn to feel hopeful ...\""): both Brobots wake here, in time
            # with the captions; "Brobots Ready!" still comes after the last words.
            if sleepy and sleepy.asleep:
                sleepy.finish()
                entry["sleepy_wake"] = sleepy.wake_all()
            # 2026-10-15 (operator: "Cube always took way too long to light up"):
            # "cube_ready": "brobot_2" - right after the wake his cube connects in
            # the background and is held, so the ending's cube_after_wake lights
            # it at once (start_cube_light / light_cube).
            if step.get("cube_ready"):
                # "cube_ready_after_seconds": the cube connect waits this long
                # after the wake (live 2026-10-15: right after a wake it fails).
                cube_light = start_cube_light(live, mod.BROBOT_2_SERIAL if step["cube_ready"] == "brobot_2" else mod.BROBOT_1_SERIAL,
                                              after_seconds=float(step.get("cube_ready_after_seconds", 0)))
                entry["cube_ready"] = "connecting"
            entry["ok"] = True

        elif note == "all_asleep":
            # Every stir done and both Brobots asleep again before going on.
            if sleepy and sleepy.asleep:
                sleepy.finish()
                if live:
                    time.sleep(step.get("settle_seconds", 0))
            entry["ok"] = True

        elif note == "system_line":
            # 2026-10-09 (operator: "system message 'run complete'"): a Goverlord
            # bubble, said in the system voice; the bubble shows with the voice
            # (gopod_say_heard.go). "wait_spoken": true (operator: "Do not go
            # forward if an expected sleep is not fucking spoken!!") - the song
            # waits until the page has played it, up to spoken_wait_seconds, and
            # stops the run if it wasn't heard.
            wait_spoken = bool(step.get("wait_spoken"))
            most = float(step.get("spoken_wait_seconds", 30.0)) if wait_spoken else 0
            if not live:
                entry["said"] = {"ok": True, "mode": "dry"}
                entry["ok"] = True
            else:
                try:
                    said = mod.wirepod_web_send_form(
                        "/api-sdk/gopod_system_say", {"text": step["text"], "wait_seconds": str(most)},
                        timeout=most + 30,
                    )
                except Exception as exc:  # noqa: BLE001 - reported below
                    said = {"error": repr(exc)}
                answer = (said.get("body") or "").strip()
                entry["said"] = said
                entry["ok"] = answer == "heard" if wait_spoken else answer in ("heard", "queued")
                print(f"{robots.timing_prefix()}{tag}_SYSTEM_LINE {step_id}: {step['text']!r} -> {answer or said}", flush=True)
                if wait_spoken and not entry["ok"]:
                    entry["fatal"] = True
                    entry["error"] = f"{step['text']!r} was not heard on the chat page within {most:.0f}s - the run stops here"

        elif note == "say_turn":
            result = say_step_line(name, step["text"])
            entry["ok"] = result["ok"]
            entry["result"] = result

        elif note == "thinking":
            # GOLDEN, locked (2026-10-13, operator: "a single robust engine for live
            # & scripted runs * snapshot the full \"thinking please wait\" process
            # from live (+ arm, nod) * lock that snapshot * use locked \"thinking
            # please wait\" process for scripted runs"): the live chat's own
            # process, run by Wire-Pod (gopodThinkingProcess, gopod_thinking_line.go)
            # - arm, nod, "Thinking, please wait.", searching, its get-out -
            # answered once it has played.
            if live:
                result = mod.wirepod_web_send_form("/api-sdk/gopod_thinking", {"serial": serial}, timeout=180)
                entry["ok"] = result.get("status") == 200
            else:
                result = {"ok": True, "mode": "dry"}
                entry["ok"] = True
            print(f"{robots.timing_prefix()}{tag}_THINKING {name}: the live thinking process", flush=True)
            entry["result"] = result

        elif note == "rattle":
            # A rattle that carries a line ("Juggling big shiny bingo balls.
            # Please wait.", 2026-10-01): the line first, then the rattle - the
            # rattle's sound and the voice can't play at once.
            if step.get("text"):
                entry["line"] = say_step_line(name, step["text"])
            result = run_rattle(mod, serial, live, robots, tag, song.get("rattle_on", "robot"))
            entry["ok"] = result["ok"] and (entry.get("line") or {}).get("ok", True)
            entry["result"] = result
            # Stay-put fix, part 4: rattle's own release (inside run_rattle,
            # required for its direct-SDK handoff - untouched) drops whatever
            # hold was active. Re-assume both serials now to restore it before
            # the next note runs.
            # On the speaker nothing was handed off, so nothing to take back.
            if continuous_hold_both and result.get("mode") != "speaker":
                reassume = run_continuous_hold_assume(mod, live)
                entry["continuous_hold_reassume"] = reassume
                entry["ok"] = entry["ok"] and reassume["ok"]
            # 2026-09-30: the robot was free during the rattle hand-off and can
            # start driving off its charger (explore) - hold its wheels still
            # once control is back (the wheel-nudge alias's own stop).
            if result.get("mode") != "speaker":
                entry["wheel_hold"] = run_wheel_hold(mod, serial, live)

        elif note == "brobots_ready_together":
            ready_phrase = step.get("text") or "Brobots ready!"
            # A sleepy song: every move made, then both wake for the line.
            if sleepy and sleepy.asleep:
                sleepy.finish()
                entry["sleepy_wake"] = sleepy.wake_all()
            # "cube_after_wake": "brobot_2" - his cube lights after the wake,
            # before the line (never stops the step).
            if step.get("cube_after_wake"):
                cube_serial = mod.BROBOT_2_SERIAL if step["cube_after_wake"] == "brobot_2" else mod.BROBOT_1_SERIAL
                if cube_light is not None and cube_light["serial"] == cube_serial:
                    entry["cube"] = light_cube(cube_light)
                    cube_light = None
                else:
                    entry["cube"] = run_cube_blip(mod, live, cube_serial)
                print(f"{robots.timing_prefix()}{tag}_CUBE {entry['cube'].get('output', entry['cube'].get('error', ''))}", flush=True)
            # "pronounce": true - said with the robots' pronunciation list
            # ("GOPOD" as "Gowp-awd"); the bubbles keep the words as written.
            spoken_phrase = mod.apply_pronunciation_safety(ready_phrase) if step.get("pronounce") else ready_phrase
            # 2026-09-30: the together line is said over the direct SDK, so it
            # never reached the chat page - show it for each robot, Brobot 1
            # first, just before they say it (operator: "display just before the
            # brobots speak together").
            chat_lines = []
            def post_ready_lines():
                chat_lines.extend(post_chat_line(mod, live, who, ready_phrase, robots) for who in ("Brobot 1", "Brobot 2"))
            result = run_brobots_ready_together(
                mod, live, spoken_phrase, before_speech=post_ready_lines,
                rattle_before=step.get("rattle_before", False),
                rattle_on=song.get("rattle_on", "robot"),
            )
            entry["ok"] = result["ok"]
            entry["result"] = result
            entry["chat_lines"] = chat_lines
            # Stay-put fix, part 4 - same reasoning as rattle above.
            if continuous_hold_both:
                reassume = run_continuous_hold_assume(mod, live)
                entry["continuous_hold_reassume"] = reassume
                entry["ok"] = entry["ok"] and reassume["ok"]

        elif note == "arm_cue":
            def arm_move():
                return control_mod.run_arm_cue(
                    mod, serial, step.get("cycles", 1), step.get("hold_seconds", 1.2), live,
                    speed=step.get("speed", 2), manage_control=manage_control,
                )
            result = run_move_with_line(arm_move, lambda: say_step_line(name, step["text"])) if step.get("text") else arm_move()
            entry["ok"] = result["ok"]
            entry["result"] = result

        elif note == "nod":
            def nod_move():
                return control_mod.run_nod(
                    mod, serial, step.get("hold_seconds", 0.35), live,
                    speed=step.get("speed", 2), manage_control=manage_control,
                )
            result = run_move_with_line(nod_move, lambda: say_step_line(name, step["text"])) if step.get("text") else nod_move()
            entry["ok"] = result["ok"]
            entry["result"] = result

        elif note == "wheel_nudge":
            # Stage 2 of WHEEL_NUDGE_GOLDEN_PATH_SURVEY_001.md - fixed
            # reverse only, no song-authored lw/rw override yet (forward is
            # a later, separate decision, per scope).
            result = control_mod.run_wheel_nudge(
                mod, serial, control_mod.WHEEL_NUDGE_REVERSE_LW, control_mod.WHEEL_NUDGE_REVERSE_RW,
                step.get("hold_seconds", control_mod.WHEEL_NUDGE_DEFAULT_HOLD_SECONDS), live,
                manage_control=manage_control,
            )
            entry["ok"] = result["ok"]
            entry["result"] = result

        elif note == "sound":
            # 2026-10-11 (the fart joke: "play fart1.wav @ 80% volume"): a song
            # sound (knobs.json "sounds", copied into the page's voice folder by
            # Wire-Pod at its start) played on the chat page at the step's volume;
            # the song waits its length.
            entry["result"] = play_song_sound(mod, live, song_dir, step.get("sound", ""), step.get("volume", 1.0), robots, tag)
            entry["ok"] = entry["result"]["ok"]

        elif note == "picture":
            # 2026-10-13 (operator: "Build missing pieces for scripted songs * as
            # per live chocolate bingo pop-up"): a picture from the chat page's
            # pictures folder (topic_pictures.folder), in its own picture bubble
            # at this step, magnifier and all; "picture" is one file or a list.
            entry["result"] = show_song_picture(mod, live, step.get("picture", ""), robots, tag)
            entry["ok"] = entry["result"]["ok"]

        elif note == "emotion_beat":
            result = run_emotion_beat(
                mod, serial, name, step["text"], step.get("animation_token"),
                step.get("hold_seconds", 5.0), live, robots, tag,
                manage_control=manage_control,
                scripted_split=song.get("sentence_split") == "scripted",
            )
            entry["ok"] = result.get("ok", False)
            entry["result"] = result

        elif note == "animation":
            token = step.get("animation_token")
            if step.get("text") and token not in _ANIMATION_GETOUT_TOKENS and token not in _LOOP_ANIMATION_TOKENS:
                # A one-shot animation that carries a line ("Bingo ball captured.
                # Number B-1!" on host_kg_success_0302, 2026-10-01): the native
                # "{{playAnimationWI||kgSuccess}}... line", one say - the clip
                # plays as the line starts, no hold after it.
                result = run_emotion_beat(
                    mod, serial, name, step["text"], token, 0.0, live, robots, tag,
                    manage_control=manage_control,
                    scripted_split=song.get("sentence_split") == "scripted",
                )
            else:
                result = run_animation_only(
                    mod, serial, token, step.get("hold_seconds", 2.5), live,
                    manage_control=manage_control,
                )
                # A held clip (searching) that carries a line: the clip, then the line.
                if step.get("text"):
                    entry["line"] = say_step_line(name, step["text"])
                    result = {**result, "ok": result.get("ok", False) and entry["line"].get("ok", False)}
            entry["ok"] = result.get("ok", False)
            entry["result"] = result

        elif note in ("pause", "timer"):
            # 2026-10-16: a timer step is a pause; its "seconds" is the wait.
            pause_seconds = step.get("pause_seconds", step.get("seconds"))
            gap_label = step.get("gap_label", step_id)
            gap_override_raw = os.getenv(GOLDEN_REPORTER_GAP_OVERRIDE_ENV)
            if gap_override_raw is not None:
                try:
                    pause_seconds = float(gap_override_raw)
                except ValueError:
                    pass
            if pause_seconds is None:
                entry["ok"] = False
                entry["error"] = f"pause step {step_id!r} has no pause_seconds in knobs.json"
            else:
                # "timeline_after": true (the nap, operator 2026-10-09: "ensure
                # narrator bubbles are fully drawn when hidden/maybe start drawing
                # them sooner"): the timeline starts when this pause ends, so the
                # first narrator bubble is sent now, early, like the others.
                if step.get("timeline_after") and song.get("timeline") == "srt" and timeline_start is None:
                    lag = float(load_brobots_settings()["reporter_voices"].get("page_lag_seconds", 0)) if song.get("audio") else 0.0
                    timeline_start = time.time() + float(pause_seconds) + lag
                    _start_captions(mod, live, song, timeline_start, captions_stop, robots)
                print(f"{robots.timing_prefix()}{tag}_REPORTER_GAP {step_id} label={gap_label}: sleeping {pause_seconds}s", flush=True)
                time.sleep(pause_seconds)
                entry["ok"] = True
                entry["pause_seconds"] = pause_seconds
                entry["gap_label"] = gap_label

        elif note == "connect":
            # Control-family only: the one place manage_control=False's whole-
            # run hold begins - assumes control once, held until the run's
            # own robots.close()/finally-equivalent release at the end.
            if not live:
                entry["ok"] = True
                entry["mode"] = "dry"
            else:
                result = mod.wirepod_web_send_form(
                    "/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10
                )
                entry["ok"] = result.get("status") == 200
                entry["result"] = result
                # 2026-09-01 operator live-report ("timing/firing mismatch -
                # I do not hear brobots say..."): this branch had ZERO settle
                # after a completely fresh assume, unlike every manage_
                # control=True path in this file (run_emotion_beat/
                # run_animation_only both sleep POST_ASSUME_SAY_SETTLE_
                # SECONDS right after their own assume). say_connected's own
                # pause_after_connect step is 0s in knobs.json, so the very
                # first say_text after connect fired essentially instantly
                # after this assume - the exact "fresh assume_behavior_
                # control needs a settle before the first action, or it can
                # silently no-op" trap CLAUDE.md's own TECHNICAL GOTCHAS
                # already names (say=success gets logged regardless -
                # Wire-Pod's own /api-sdk/say_text handler discards the real
                # result). Same 1.5s constant, already defined in this file,
                # now actually used on this path too.
                time.sleep(POST_ASSUME_SAY_SETTLE_SECONDS)

        elif note == "say":
            # Control-family only: a pure placeholder note whose own text was
            # already spoken above via unconditional_text_speak - this note
            # itself does nothing else, matching run_robot_control_song_001.py's
            # own "say" branch exactly.
            entry["ok"] = True

        elif note == "weather":
            # Control-family only: real per-robot weather fetch, reusing
            # control_mod.fetch_and_format_weather() verbatim rather than
            # reimplementing it - no try/except, a fetch failure raises and
            # stops the run, same "no fallback line" stop condition the
            # original enforces. Takes `serial` (the already-resolved ESN
            # from _serial_and_name() above, line ~952), not robot_param
            # ("1"/"2") - fetch_and_format_weather() now reads that robot's
            # LIVE jdocs RobotSettings by serial, not a static file keyed by
            # "1"/"2" (WEATHER_LIVE_JDOCS_SOURCED_001.md). MISSED in that
            # pass - this call site was the live path for control/weather
            # post-cutover and still passed robot_param, causing a live
            # BLOCKED failure (vic:1 not found in jdocs.json) - fixed here.
            weather = control_mod.fetch_and_format_weather(
                serial, include_location_date=step.get("weather_include_location_date", False)
            )
            say_result = run_say_turn(
                mod, robots, name, weather["spoken"], display_text=weather["display_text"],
                suppress_assume_release=suppress_assume_release,
            )
            entry["ok"] = say_result["ok"]
            entry["result"] = weather

        elif note == "exit":
            entry["ok"] = True

        else:
            entry["ok"] = False
            entry["error"] = f"unknown note {note!r}"

        # Optional scored "fail" line (control-family only - step["fail"] is
        # always empty for bingo-family, since their own story.md never
        # carries a "> FAIL:" line, but gated on unconditional_text_speak
        # anyway so bingo-family behavior can never change here even by
        # accident). Result discarded, matching run_robot_control_song_001.py's
        # own say_line(mod, robots, name, step["fail"]) call exactly - this
        # never affects entry["ok"] itself.
        if unconditional_text_speak and not entry.get("ok") and step.get("fail"):
            run_say_turn(mod, robots, name, step["fail"], suppress_assume_release=suppress_assume_release)

        # A failed robot step, or a line still silent after the heal, under the
        # song's hold: retake that robot's control before its next step.
        if continuous_hold_both and note in CONTROL_RETAKE_NOTES and not step_filtered:
            step_result = entry.get("result") if isinstance(entry.get("result"), dict) else {}
            heal = step_result.get("silent_heal")
            still_silent = bool(heal) and heal.get("retry_seconds", 0) < SAY_SILENT_SECONDS
            if not entry.get("ok") or still_silent:
                needs_control_retake.add(serial)
                print(f"{robots.timing_prefix()}{tag}_CONTROL_RETAKE_PENDING {name}: {step_id} {'failed' if not entry.get('ok') else 'still silent'} - control is retaken before its next step", flush=True)
        # The rattle and the together line retake both robots already.
        if (entry.get("continuous_hold_reassume") or {}).get("ok"):
            needs_control_retake.clear()

        if not entry.get("ok") and (note in fatal_notes or entry.get("fatal")):
            message_fn = _FATAL_STOP_MESSAGES.get(note)
            message = entry["error"] if entry.get("fatal") and entry.get("error") else (
                message_fn(step_id) if message_fn else _generic_fatal_message(note, step_id))
            print(f"{robots.timing_prefix()}{tag}_STOP {step_id}: {message}", flush=True)
            log["steps"].append(entry)
            log["stopped_early"] = True
            _release_if_holding(mod, hold_release_serials, live, manage_control)
            robots.close()
            return _finalize(log, tag)
        elif not entry.get("ok") and note in _WARN_ON_FAILURE_NOTES:
            print(f"{robots.timing_prefix()}{tag}_WARNING {step_id}: {note} failed - continuing, not stopping the run.", flush=True)

        log["steps"].append(entry)
        print(f"{robots.timing_prefix()}{tag}_DONE {step_id}: ok={entry.get('ok')}", flush=True)
        if live and song.get("llm_warmup") and step_id == song.get("warmup_after"):
            song_llm_warmup(mod, song.get("llm_ready_text", ""), song.get("warmup_lines", []),
                            song.get("voices_required", True), song.get("llm_required", True))

        buffer_after = step.get("buffer_after", 0)
        # tempo_factor/tempo_comment ride the wholesale step merge in
        # load_golden_song() for free - tempo_comment is human-facing only,
        # never read here. added_space is additive on top of buffer_after,
        # never a replacement for it (TEMPO_BUFFER_KNOB_SURVEY_001.md, B).
        # Excluded on "exit" - nothing follows it to buffer against. Also
        # excluded on "pause" (2026-08-15, REPORTER_GAP_TEMPO_IMMUNE_001.md) -
        # reporter-gap/divider steps are tempo-immune BY DESIGN, not a
        # per-song data default that could be forgotten: proven live,
        # global_tempo=1.3 on awaken silently doubled every reporter-gap's
        # own authored pause_seconds (1.0s authored -> 2.3s actual), since a
        # "pause" step already sleeps its own gap a few lines above and then
        # fell through to this same unconditional block like any other note.
        # A reporter-gap's timing is deliberately authored to the second for
        # the video edit - tempo drifting it was never intended, invisible
        # only by accident (every song's global_tempo happened to be 0).
        tempo_factor = step.get("tempo_factor", 1.0)
        added_space = (global_tempo * tempo_factor) if note not in ("exit", "pause") else 0
        total_delay = buffer_after + added_space
        if buffer_after:
            print(f"{robots.timing_prefix()}{tag}_BUFFER step_id={step_id} sleeping {buffer_after}s", flush=True)
        if added_space:
            print(f"{robots.timing_prefix()}{tag}_TEMPO step_id={step_id} sleeping {added_space}s (global_tempo={global_tempo} x tempo_factor={tempo_factor})", flush=True)
        if total_delay:
            time.sleep(total_delay)

        entry["step_wall_seconds"] = round(time.time() - step_start_wall, 3)
    # A cube connect still going when the steps end (or stop) is finished too.
    if cube_job is not None:
        log["cube_connect"] = finish_cube_connect(cube_job)
    if cube_light is not None:
        log["cube_light"] = light_cube(cube_light)

    # 2026-08-23, operator request: reveal the webroot chat-bubble close
    # button off the LAST say landing, not off this whole process exiting.
    # brobots.sh's own post-subshell run_active:false write (pha0b(),
    # ~line 1988) only fires once this script returns - measured (not
    # guessed): on every live run that isn't stopped early, that return is
    # gated behind _maybe_promote_knobs()'s own blocking terminal y/n
    # prompt below, which fires whenever the song's own zKnobs.json exists
    # - true for every song directory in this repo today (checked all 8),
    # not a rare case. That prompt waits on the operator's own keyboard,
    # an indefinite, real, user-visible gap - not the sub-second network/
    # disk work release/close/finalize actually take. Writing
    # run_active:false HERE, the moment the step loop's own last buffer/
    # tempo delay finishes, means the close button appears when the
    # performance is actually done, independent of that unrelated prompt.
    # Mirrors Robots.__init__'s own start-of-run write (same file, same
    # gate on console_rich_display) - this is that same write's end-of-run
    # twin, not a new mechanism. brobots.sh's own later write still fires
    # once the process eventually exits - harmless, already-true content,
    # left as-is; not worth removing for this fix's own minimal scope.
    # AWAKEN_MULTI_MODE_TWO_PASS follow-up, 2026-09-01 operator live-report
    # ("chat bubbles div didn't flip on and display"): this write assumed
    # one Python process = one whole run. bait's "both" two-pass loop
    # launches a FRESH process per robot, so without the same suppress
    # gate _maybe_promote_knobs already uses, brobot 1's pass wrote
    # run_active:false (closeBtn.style.display flips to visible,
    # index.html:331) mid-run, well before brobot 2's pass even started -
    # the div's "run just finished" state fired a full pass early, not on
    # the actual end of the whole two-pass run. Same GOPOD_GOLDEN_SUPPRESS_
    # PROMOTE brobots.sh already exports for every pass except the last one
    # gates this write too - one flag, one meaning ("not the final pass
    # yet"), reused rather than duplicated. brobots.sh's own outer write
    # (pha0b(), after the whole loop) still fires unconditionally and is
    # the real backstop for every case, multi-pass included.
    # 2026-10-17 (survey E15: "a state file nothing reads"): gopod_rich_display_ui_state.json is no longer written - the chat page opens and closes its own chat (the Log button), and nothing read it.

    if sleepy:
        sleepy.wake_all()
        log["sleepy_moves"] = sleepy.results
    _release_if_holding(mod, hold_release_serials, live, manage_control)
    robots.close()
    return _finalize(log, tag)


def main():
    voice_destination = os.getenv("GOPOD_VOICE_DESTINATION", "robot")
    live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
    # The song: GOPOD_GOLDEN_SONG_ID (its song_id, wherever its folder is), else
    # GOPOD_GOLDEN_SONG_DIR, else the default song.
    song_id = os.getenv(GOLDEN_SONG_ID_ENV) or (None if os.getenv(GOLDEN_SONG_DIR_ENV) else DEFAULT_SONG_ID)
    if song_id:
        found = find_song_dir(song_id)
        if found is None:
            print(f"GOLDEN_REFUSED no song folder has song_id {song_id!r} (song_folders_001.py lists them).", flush=True)
            return None
        song_dir = str(found)
    else:
        song_dir = os.getenv(GOLDEN_SONG_DIR_ENV)
    # 2026-10-09: the interview's song folder is played by the Brobots chat
    # ("start interview", gopod_interview_song.go), not by this engine.
    try:
        song_type = json.loads((Path(song_dir) / "knobs.json").read_text(encoding="utf-8")).get("song_type")
    except (OSError, ValueError):
        song_type = None
    if song_type == "interview":
        print("This is the interview: type \"start interview\" in the Brobots chat to play it.", flush=True)
        return None
    return run_golden_song(song_dir, live, voice_destination)


if __name__ == "__main__":
    main()

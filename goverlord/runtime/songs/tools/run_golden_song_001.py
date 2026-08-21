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
import json
import os
import re
import shutil
import subprocess
import sys
import time
from pathlib import Path

from knobs_envelope_001 import load_knobs_envelope

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
_FILTER_EXEMPT_NOTES = {"wake_both", "brobots_ready_together", "pause", "exit"}

# Same bare-presence-is-not-proof discriminator as run_songs_runner_001.py -
# see that file's own comment / BINGO_STUCK_ANIMATION_FALSE_POSITIVE_RESOLUTION_001.md.
ANIMATION_WAIT_LOG_SIGNATURE = "waiting for animation to be done"

_LOOP_ANIMATION_TOKENS = {"searching", "answering"}
_ANIMATION_LOOP_INTERVAL_SECONDS = 0.333

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

RATTLE_BIN_PATH = Path("/home/goverlord/wire-pod/chipper/gopod_probes/tools/direct_sdk_bingo_rattle_001")
RATTLE_WAV_PATH = Path(
    "/home/goverlord/crushn8r_git/SDK/sources/vectorx/vectorfs/data/audio/rattle.wav"
)
TOGETHER_BIN_PATH = Path("/home/goverlord/wire-pod/chipper/gopod_probes/tools/direct_sdk_brobots_ready_001")
WIREPOD_HOME_PATH = "/home/goverlord/wire-pod"

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


# Per-song registry - Phase 1 prompt one seeded brobots_baby_robots_sleep
# alone; this pass (prompt two) widens to the rest of the bingo-family set
# actually driven by run_songs_runner_001.py today: brobots_bingo (via
# pha0b's "bingo" arm) and brobots_cross_persona (via "mixup"). Held out on
# purpose: the awakening/control-family songs (00_brobots_awaken and the two
# archived control-song-family songs, all on run_robot_control_song_001.py -
# Phase 2, a separate conversion) and vamp-gate (run_vamp_gate_song_001.py -
# flagged in the survey as an open "worth folding at all?" call, not folded
# in here). manage_control/note_aliases/synthesize_speaker/
# unconditional_text_speak are the Phase 2 (control-song-family shim) fields
# from the survey's schema table - present here as real, read fields (not
# dead placeholders) but every bingo-family song sets them to the no-op
# default, matching run_songs_runner_001.py's own current behavior exactly:
# that file's stop-condition logic (wake_both/emotion_beat fatal, everything
# else warn-and-continue) is global to the runner, not song-specific, so all
# three bingo-family entries below share the identical shape by design - no
# divergent quirk exists yet for this family to capture.
# Stay-put fix (STAY_PUT_FIX_SURVEY_001.md, SAFE-WITH-HANDLING, built 2026-08-08):
# manage_control=False on all three bingo-family songs below - continuous dual-robot
# hold instead of per-note assume/release, so neither robot goes autonomous
# mid-take. synthesize_speaker stays False (both robots are addressed per-step,
# real per-step `speaker` field, unlike awaken's single-run-robot shape) - this is
# a NEW manage_control=False + synthesize_speaker=False combination, handled by its
# own branch everywhere below (continuous_hold_both), never touching awaken's own
# existing manage_control=False + synthesize_speaker=True path. The other three
# pieces of handling this flip requires (two-serial hold, the initial dual-robot
# assume off wake_both, the emotion_beat/animation bypass, the rattle/
# brobots_ready_together re-assume-after) live in the dispatch loop and note
# handlers below, not here - this flag is what turns them on.
SONG_REGISTRY = {
    "brobots_baby_robots_sleep": {
        # song_dir repointed 2026-08-18: folder renamed/renumbered on disk
        # from 104_brobots_baby_robots_sleep/ to 105_brobots_nap/ (operator's
        # own folder surgery) - song_id itself (the SONG_REGISTRY key and the
        # knobs.json "song_id" field this key must match) is unchanged, only
        # the directory moved.
        "song_dir": SONGS_DIR / "105_brobots_nap",
        "runner": "run_golden_song_001.py",
        "golden_ready": True,
        "fatal_notes": {"wake_both", "emotion_beat"},
        "manage_control": False,
        "note_aliases": {},
        "synthesize_speaker": False,
        "unconditional_text_speak": False,
    },
    "brobots_bingo": {
        "song_dir": SONGS_DIR / "101_brobots_bingo_test",
        "runner": "run_golden_song_001.py",
        "golden_ready": True,
        "fatal_notes": {"wake_both", "emotion_beat"},
        "manage_control": False,
        "note_aliases": {},
        "synthesize_speaker": False,
        "unconditional_text_speak": False,
    },
    "brobots_cross_persona": {
        # song_dir corrected 2026-08-18 (ZZZ_ARCHIVES_PRUNE_001.md): this
        # entry never had "zzz_archives" in its path even after the folder
        # was archived there 2026-08-12 - masked in practice because pha0b's
        # own "mixup" case arm exports GOPOD_GOLDEN_SONG_DIR with the correct
        # path, which this registry default is only a fallback for.
        "song_dir": SONGS_DIR / "zzz_archives" / "102_brobots_cross_persona",
        "runner": "run_golden_song_001.py",
        "golden_ready": True,
        "fatal_notes": {"wake_both", "emotion_beat"},
        "manage_control": False,
        "note_aliases": {},
        "synthesize_speaker": False,
        "unconditional_text_speak": False,
    },
    # goverlord/runtime/songs/103_gopod_is_that_you_single/ and
    # 104_gopod_is_that_you_multi/ - split 2026-08-18 (operator's own folder
    # surgery) from the single prior 103_gopod_is_that_you/ folder, wired
    # straight onto the golden engine (2026-08-08, no legacy-runner stop -
    # mixup/nap/bingo were already all cut over). Same bingo-family shape
    # (wake_both/emotion_beat fatal, manage_control holds per-step assume/
    # release, no note aliases) even though neither song's own 3-step score
    # uses emotion_beat or say_turn - the shape is shared config, not a claim
    # every note fires. _single scopes the live is-that-you() PTT test's own
    # live-capture bookend to KP1 only (Brobot 1/Doc); _multi keeps the
    # original KP1+KP2 scope as the golden version - see each folder's own
    # story.md for the split's full reasoning. Two distinct song_ids (both
    # previously shared the literal string "gopod_is_that_you", which would
    # have collided once both were registered).
    "gopod_is_that_you_single": {
        "song_dir": SONGS_DIR / "103_gopod_is_that_you_single",
        "runner": "run_golden_song_001.py",
        "golden_ready": True,
        "fatal_notes": {"wake_both", "emotion_beat"},
        "manage_control": True,
        "note_aliases": {},
        "synthesize_speaker": False,
        "unconditional_text_speak": False,
    },
    "gopod_is_that_you_multi": {
        "song_dir": SONGS_DIR / "104_gopod_is_that_you_multi",
        "runner": "run_golden_song_001.py",
        "golden_ready": True,
        "fatal_notes": {"wake_both", "emotion_beat"},
        "manage_control": True,
        "note_aliases": {},
        "synthesize_speaker": False,
        "unconditional_text_speak": False,
    },
    # Phase 2 - control-family conversion, on a COPY only originally
    # (00_brobots_awaken_golden_copy/). The operator's own manual folder
    # surgery (2026-08-06) renamed that copy INTO 00_brobots_awaken (archiving
    # the old original to zzz_archives/00_brobots_awaken_old01) - song_dir
    # below updated to match current disk truth
    # (INTERVIEW_SONG_LOCATION_AND_00_WIRING_SURVEY_001.md). This field was
    # never actually load-bearing at runtime (song lookup keys on the song_id
    # found inside whatever GOPOD_GOLDEN_SONG_DIR points at, not this stored
    # path) but was stale/dishonest data and is fixed here regardless.
    # The 3 real differences from bingo-family, per the survey
    # (GOLDEN_SONG_RUNNER_PLAN_SURVEY_001.md), all captured as registry data:
    # zero fatal notes (a failure speaks its own optional scored "fail" line
    # and the run continues - run_robot_control_song_001.py never stops
    # early), manage_control=False (this song's own "connect" note assumes
    # control once for the whole run, held until the final release - never
    # per-step), and synthesize_speaker=True (every step's own knobs.json
    # "speaker" field is unused dead data in the original - the real speaker
    # comes from a whole-run robot choice, GOLDEN_ROBOT_ENV here). Plus
    # note_aliases (head_nod -> the golden engine's canonical "nod") and
    # unconditional_text_speak=True (this song speaks a step's own text
    # BEFORE its note-specific action fires, for every note type - not only
    # a dedicated say_turn note the way bingo-family works).
    "brobots_awaken_golden": {
        "song_dir": SONGS_DIR / "00_brobots_awaken",
        "runner": "run_golden_song_001.py",
        "golden_ready": True,
        "fatal_notes": set(),
        "manage_control": False,
        "note_aliases": {"head_nod": "nod"},
        "synthesize_speaker": True,
        "unconditional_text_speak": True,
    },
    # Studio tuning cut 2, Step 1 (CONTROL_SONG_FOLD_SURVEY_001 SS6.1) -
    # registration only. Same control-family shape as brobots_awaken_golden
    # above (zero fatal notes, manage_control=False, head_nod->nod alias,
    # synthesize_speaker/unconditional_text_speak both True) - both archived
    # songs use the identical note vocabulary awaken already proves live.
    # Registering makes golden CAPABLE of running these two; no alias
    # (start-the-control-song/pha0b control|weather) is repointed yet -
    # run_control_song() is still their live path.
    "robot_control_song_001": {
        "song_dir": SONGS_DIR / "zzz_archives" / "robot_control_song_001",
        "runner": "run_golden_song_001.py",
        "golden_ready": True,
        "fatal_notes": set(),
        "manage_control": False,
        "note_aliases": {"head_nod": "nod"},
        "synthesize_speaker": True,
        "unconditional_text_speak": True,
    },
    # Known to the composer's own single source of truth, NOT yet
    # golden-dispatch-ready - a real, different execution shape (a two-phase
    # generate-then-play architecture, LLM content pre-generated then
    # replayed, plus a concurrent pre-show/vamp filler loop), not a
    # note-dispatch score - confirmed this song's own story.md has zero
    # "## STEP" headings, the shape load_golden_song()/parse_control_story_md()
    # both require. Migration path, per the operator's own explicit
    # sequencing: vamp-gate goes first (Part 1 - a lower-risk proxy, no
    # physical robot, already fills the same pre-show-filler role this song's
    # own concurrent thread does today), the interview itself second (Part
    # 2), once vamp's own separate usefulness as a standalone filler
    # mechanism has ended. Neither part is started - this entry exists so the
    # roster is honest and complete, not to claim readiness. Attempting to
    # actually dispatch this song_id is refused cleanly (see golden_ready
    # check in _run_golden_song_steps), not attempted.
    "brobots_interview_section_01": {
        # Repointed 2026-08-19 (INTERVIEW_VAMP_SPLIT_001.md): the old
        # single folder split into 01_brobots_interview_vamp/ (its own
        # standalone song now) and 02_brobots_interview_run/ (this song_id's
        # actual content). song_id string kept unchanged - identity,
        # decoupled from folder path.
        "song_dir": SONGS_DIR / "02_brobots_interview_run",
        "runner": "run_section1_full_live_001.py",
        "golden_ready": False,
    },
}

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


def load_golden_song(song_dir, control_mod):
    """Same parsing shape as run_songs_runner_001.py's own
    load_bingo_capture_song(): reuses control_mod.parse_control_story_md() for
    STEP-heading parsing, merges every knobs.json field through (no
    whitelist, so animation_token/pre_animation_pause_seconds/hold_seconds/
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
        steps.append(merged)
    return {
        "song_id": knobs.get("song_id", ""),
        "global_tempo": knobs.get("global_tempo", 0.0),
        "steps": steps,
    }


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
    return {"ok": ok, "results": results}


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


def run_rattle(mod, serial, live, robots, tag):
    if not live:
        return {"ok": True, "mode": "dry"}
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


def run_brobots_ready_together(mod, live, phrase="Brobots ready!"):
    if not live:
        return {"ok": True, "mode": "dry"}
    release_1 = mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": mod.BROBOT_1_SERIAL}, timeout=10)
    release_2 = mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": mod.BROBOT_2_SERIAL}, timeout=10)
    time.sleep(DIRECT_SDK_RELEASE_SETTLE_SECONDS)
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
    return {"ok": ok, "release_1": release_1, "release_2": release_2, "returncode": proc.returncode, "output": combined}


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
    robot_safe_line, _repair_used = mod.normalize_robot_safe(text, robots.scaffold)
    result = robots.say(
        name, robot_safe_line, display_text=display_text, display_role=name,
        suppress_assume_release=suppress_assume_release,
    )
    ok = result.get("spoken") is True or result.get("mode") in ("dry", "off")
    return {"ok": ok, "result": result}


def run_emotion_beat(mod, serial, name, text, animation_token, pre_animation_pause_seconds, hold_seconds, live, robots, tag, manage_control=True):
    """manage_control=False (STAY_PUT_FIX_SURVEY_001.md, part 3) skips this
    function's own inline assume/release, mirroring run_move_axis()'s already-
    proven bypass exactly - relies on the caller already holding continuous
    control. Default True is byte-identical to every existing call site
    (bingo-family's own manage_control=True songs, and any future caller that
    never passes this kwarg)."""
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
        say = mod.wirepod_web_send_form("/api-sdk/say_text", {"text": text, "serial": serial}, timeout=90)
        entry["say"] = say
        print(f"{robots.timing_prefix()}{tag}_BEAT {name} token={animation_token}: said, pausing {pre_animation_pause_seconds}s before animation dispatch", flush=True)
        time.sleep(pre_animation_pause_seconds)
        anim_payload = f"{{{{playAnimationWI||{animation_token}}}}}..."
        anim = mod.wirepod_web_send_form("/api-sdk/say_text", {"text": anim_payload, "serial": serial}, timeout=90)
        entry["animation_dispatch"] = anim
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
        entry["ok"] = say.get("status") == 200 and anim.get("status") == 200
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
        if animation_token in _LOOP_ANIMATION_TOKENS:
            elapsed = 0.0
            while elapsed < hold_seconds:
                dispatches.append(
                    mod.wirepod_web_send_form("/api-sdk/say_text", {"text": anim_payload, "serial": serial}, timeout=10)
                )
                time.sleep(_ANIMATION_LOOP_INTERVAL_SECONDS)
                elapsed += _ANIMATION_LOOP_INTERVAL_SECONDS
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
_WARN_ON_FAILURE_NOTES = {"rattle", "brobots_ready_together", "arm_cue", "nod", "animation"}


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

    try:
        log = _run_golden_song_steps(song_dir, live, voice_destination, _finalize)
    finally:
        if sys.stdout is tee:
            sys.stdout = real_stdout
        if not log_file.closed:
            log_file.close()
    if live and not log.get("stopped_early", True):
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
            to_index = all_step_ids.index(playhead_to_id) if playhead_to_id else len(all_step_ids) - 1
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

    current_section = None
    for step in song["steps"]:
        step_id = step["step_id"]
        note = note_aliases.get(step["note"], step["note"])
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
            if continuous_hold_both and entry["ok"]:
                hold_assume = run_continuous_hold_assume(mod, live)
                entry["continuous_hold_assume"] = hold_assume
                entry["ok"] = entry["ok"] and hold_assume["ok"]

        elif note == "say_turn":
            result = run_say_turn(mod, robots, name, step["text"], suppress_assume_release=suppress_assume_release)
            entry["ok"] = result["ok"]
            entry["result"] = result

        elif note == "rattle":
            result = run_rattle(mod, serial, live, robots, tag)
            entry["ok"] = result["ok"]
            entry["result"] = result
            # Stay-put fix, part 4: rattle's own release (inside run_rattle,
            # required for its direct-SDK handoff - untouched) drops whatever
            # hold was active. Re-assume both serials now to restore it before
            # the next note runs.
            if continuous_hold_both:
                reassume = run_continuous_hold_assume(mod, live)
                entry["continuous_hold_reassume"] = reassume
                entry["ok"] = entry["ok"] and reassume["ok"]

        elif note == "brobots_ready_together":
            result = run_brobots_ready_together(mod, live, step.get("text") or "Brobots ready!")
            entry["ok"] = result["ok"]
            entry["result"] = result
            # Stay-put fix, part 4 - same reasoning as rattle above.
            if continuous_hold_both:
                reassume = run_continuous_hold_assume(mod, live)
                entry["continuous_hold_reassume"] = reassume
                entry["ok"] = entry["ok"] and reassume["ok"]

        elif note == "arm_cue":
            result = control_mod.run_arm_cue(
                mod, serial, step.get("cycles", 1), step.get("hold_seconds", 1.2), live,
                speed=step.get("speed", 2), manage_control=manage_control,
            )
            entry["ok"] = result["ok"]
            entry["result"] = result

        elif note == "nod":
            result = control_mod.run_nod(
                mod, serial, step.get("hold_seconds", 0.35), live,
                speed=step.get("speed", 2), manage_control=manage_control,
            )
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

        elif note == "emotion_beat":
            result = run_emotion_beat(
                mod, serial, name, step["text"], step.get("animation_token"),
                step.get("pre_animation_pause_seconds", 2.0),
                step.get("hold_seconds", 5.0), live, robots, tag,
                manage_control=manage_control,
            )
            entry["ok"] = result.get("ok", False)
            entry["result"] = result

        elif note == "animation":
            result = run_animation_only(
                mod, serial, step.get("animation_token"), step.get("hold_seconds", 2.5), live,
                manage_control=manage_control,
            )
            entry["ok"] = result.get("ok", False)
            entry["result"] = result

        elif note == "pause":
            pause_seconds = step.get("pause_seconds")
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

        if not entry.get("ok") and note in fatal_notes:
            message_fn = _FATAL_STOP_MESSAGES.get(note)
            message = message_fn(step_id) if message_fn else _generic_fatal_message(note, step_id)
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

    _release_if_holding(mod, hold_release_serials, live, manage_control)
    robots.close()
    return _finalize(log, tag)


def main():
    voice_destination = os.getenv("GOPOD_VOICE_DESTINATION", "robot")
    live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
    default_song_dir = SONG_REGISTRY[DEFAULT_SONG_ID]["song_dir"]
    song_dir = os.getenv(GOLDEN_SONG_DIR_ENV, str(default_song_dir))
    return run_golden_song(song_dir, live, voice_destination)


if __name__ == "__main__":
    main()

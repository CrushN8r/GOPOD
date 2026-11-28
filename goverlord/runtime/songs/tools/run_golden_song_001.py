#!/usr/bin/env python3
"""Golden song runner — Phase 1 of the golden-song-runner unification path."""
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

from knobs_envelope_001 import load_knobs_envelope, apply_knob_layers, global_runner_value
from song_folders_001 import find_song_dir

RUNNER_PATH = Path(__file__).resolve().parent / "run_section1_full_live_001.py"
CONTROL_SONG_RUNNER_PATH = Path(__file__).resolve().parent / "run_robot_control_song_001.py"
SONGS_DIR = Path(os.getenv("GOPOD_REPO_ROOT", "/home/goverlord/crushn8r_git/GOPOD")).expanduser() / "goverlord" / "runtime" / "songs"

GOLDEN_SONG_DIR_ENV = "GOPOD_GOLDEN_SONG_DIR"
GOLDEN_SONG_ID_ENV = "GOPOD_GOLDEN_SONG_ID"
GOLDEN_SECTION_ENV = "GOPOD_GOLDEN_SECTION"
GOLDEN_PLAYHEAD_FROM_ENV = "GOPOD_GOLDEN_PLAYHEAD_FROM"
GOLDEN_PLAYHEAD_TO_ENV = "GOPOD_GOLDEN_PLAYHEAD_TO"
GOLDEN_PLAYHEAD_WAKE_ENV = "GOPOD_GOLDEN_PLAYHEAD_WAKE"
GOLDEN_REPORTER_GAP_OVERRIDE_ENV = "GOPOD_GOLDEN_REPORTER_GAP_OVERRIDE"
GOLDEN_CONNECTION_PASS_ENV = "GOPOD_GOLDEN_CONNECTION_PASS"
CONNECTION_PASS_TIMEOUT_SECONDS = global_runner_value("connection_pass_timeout_seconds")  # songs/global_knobs.json

GOLDEN_ROBOT_ENV = "GOPOD_GOLDEN_ROBOT"

GOLDEN_ROBOT_FILTER_ENV = "GOPOD_GOLDEN_ROBOT_FILTER"
_ROBOT_FILTER_SPEAKER = {"1": "brobot_1", "2": "brobot_2"}
_FILTER_EXEMPT_NOTES = {"wake_both", "sleep_both", "all_asleep", "brobots_ready_together", "system_line", "pause", "timer", "exit"}

ANIMATION_WAIT_LOG_SIGNATURE = "waiting for animation to be done"

_LOOP_ANIMATION_TOKENS = {"searching", "answering"}
_ANIMATION_LOOP_INTERVAL_SECONDS = 0.333

_ANIMATION_GETOUT_TOKENS = {"searching": "searchingGetout"}
_ANIMATION_GETOUT_SETTLE_SECONDS = 0.5
_ANIMATION_NO_GETOUT_SETTLE_SECONDS = 1.0

DIRECT_SDK_RELEASE_SETTLE_SECONDS = 1.0
BROBOTS_READY_TOGETHER_HANDBACK_SETTLE_SECONDS = 1.0
RATTLE_SETTLE_SECONDS = 3.0

POST_ASSUME_SAY_SETTLE_SECONDS = 1.5

SAY_SILENT_SECONDS = 0.5

RATTLE_BIN_PATH = Path("/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_probes/tools/direct_sdk_bingo_rattle_001")
RATTLE_WAV_PATH = Path(
    "/home/goverlord/crushn8r_git/SDK/sources/vectorx/vectorfs/data/audio/rattle.wav"
)
TOGETHER_BIN_PATH = Path("/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_probes/tools/direct_sdk_brobots_ready_001")
CUBE_BLIP_BIN_PATH = Path("/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_probes/tools/direct_sdk_cube_blip_001")
WIREPOD_HOME_PATH = "/home/goverlord/wire-pod"

SLEEP_BIN_PATH = TOGETHER_BIN_PATH.parent / "direct_sdk_robot_sleep_001"
SLEEPY_MAX_WAIT_SECONDS = float(global_runner_value("sleepy_max_wait_seconds"))  # songs/global_knobs.json
SLEEPY_MOVE_NOTES = ("arm_cue", "nod", "animation")

WIREPOD_PAIRING_RECOVERY_MESSAGE = (
    "sequence error - Wire-Pod could not reach the robot(s). Field-proven recovery "
    "order (robot_control_song_001/story.md's own Troubleshooting section, confirmed "
    "live - a partial version does not clear it): 1) power-cycle both robots, "
    "2) re-pair them with Wire-Pod, 3) restart Wire-Pod (wpr). Stopping here, not "
    "guessing a different fix."
)

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

SONG_REGISTRY["brobots_fart_joke"] = dict(SONG_REGISTRY["brobots_bingo"])

SONG_REGISTRY["brobots_rapid_fire"] = dict(SONG_REGISTRY["brobots_bingo"])

SONG_REGISTRY["brobots_ready"] = dict(SONG_REGISTRY["brobots_bingo"])

DEFAULT_SONG_ID = "brobots_baby_robots_sleep"


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
    """a sequence step plays another song's steps, in place - its "song_id" names the song (found by name, song_folders_001.py)."""
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
    """a combo step is one story.md step played as its "combo" notes, in order."""
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
    """Same parsing shape as run_songs_runner_001.py's own load_bingo_capture_song(): reuses control_mod.parse_control_story_md() for STEP-heading parsing, merges every knobs.json field through (no whitelist, so animation_token/hold_seconds/ gap_label all survive), plus story.md's own text/fail."""
    song_dir = Path(song_dir)
    knobs_path, knobs = load_knobs_envelope(song_dir)
    apply_knob_layers(knobs)
    story_text = (song_dir / "story.md").read_text(encoding="utf-8")
    in_order = control_mod.parse_control_story_list(story_text) if hasattr(control_mod, "parse_control_story_list") else list(control_mod.parse_control_story_md(story_text).items())
    by_name = {}
    for sid, content in in_order:
        by_name.setdefault(sid, []).append(content)
    taken = {}
    steps = []
    for raw_step in knobs.get("steps", []):
        step_id = raw_step.get("step_id")
        n = taken.get(step_id, 0)
        taken[step_id] = n + 1
        found = by_name.get(step_id, [])
        content = found[n] if n < len(found) else (found[-1] if found else {"text": "", "fail": ""})
        merged = dict(raw_step)
        merged["text"] = content["text"]
        merged["fail"] = content["fail"]
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
    if knobs.get("compose_moves"):
        steps = compose_move_lines(steps)
    return {
        "song_id": knobs.get("song_id", ""),
        "global_tempo": knobs.get("global_tempo", 0.0),
        "chat_intro": knobs.get("chat_intro", ""),
        "sentence_split": knobs.get("sentence_split"),
        "rattle_on": knobs.get("rattle_on", "robot"),
        "touch_interrupt_off": bool(knobs.get("touch_interrupt_off")),
        "llm_warmup": bool(knobs.get("llm_warmup", True)),
        "llm_ready_text": knobs.get("llm_ready_text", ""),
        "warmup_lines": _llm_song_lines(song_dir, control_mod, knobs.get("warmup_lines", [])),
        "voices_required": knobs.get("voices_required", True) is not False,
        "llm_required": knobs.get("llm_required", True) is not False,
        "warmup_after": knobs.get("warmup_after", ""),
        "timeline": knobs.get("timeline", ""),
        "audio": knobs.get("audio", ""),
        "sleepy": bool(knobs.get("sleepy")),
        "wake_clip": knobs.get("wake_clip", ""),
        "section_pictures": knobs.get("section_pictures", {}),
        "narrator_single": bool(knobs.get("narrator_single")),
        "narrator_words": _narrator_words(song_dir),
        "steps": steps,
    }


COMPOSE_MOVE_NOTES = ("arm_cue", "nod", "animation", "emotion_beat")


def compose_move_lines(steps):
    out = []
    i = 0
    while i < len(steps):
        step = steps[i]
        nxt = steps[i + 1] if i + 1 < len(steps) else None
        if (step.get("note") in COMPOSE_MOVE_NOTES and not (step.get("text") or "").strip()
                and nxt is not None and nxt.get("note") == "say_turn" and (nxt.get("text") or "").strip()
                and nxt.get("speaker") == step.get("speaker")):
            joined = dict(step)
            joined["text"] = nxt["text"]
            joined["composed_from"] = [step.get("step_id"), nxt.get("step_id")]
            if "buffer_after" in nxt:
                joined["buffer_after"] = nxt["buffer_after"]
            if nxt.get("show_ball"):
                joined["show_ball"] = nxt["show_ball"]
            out.append(joined)
            i += 2
            continue
        out.append(step)
        i += 1
    return out


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
    """manage_control=False songs assume control once (awaken's "connect" note, or the stay-put fix's own dual-robot assume off wake_both) and hold it continuously - Robots.close() never releases it (it's a no-op), so this must fire explicitly wherever the run ends, matching run_robot_control_song_001.py's own always-attempt `finally` release exactly."""
    if not (live and not manage_control):
        return
    for serial in serials or ():
        if serial:
            mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)


def run_continuous_hold_assume(mod, live):
    """Stay-put fix: assumes OVERRIDE_BEHAVIORS on BOTH robot serials at once."""
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
    time.sleep(POST_ASSUME_SAY_SETTLE_SECONDS)
    return {"ok": ok, "results": results, "settle_seconds": POST_ASSUME_SAY_SETTLE_SECONDS}


CONTROL_RETAKE_NOTES = ("say_turn", "animation", "emotion_beat", "arm_cue", "nod")
CONTROL_RETAKE_WAIT_SECONDS = float(global_runner_value("control_retake_wait_seconds"))  # songs/global_knobs.json
CONTROL_RETAKE_POLL_SECONDS = float(global_runner_value("control_retake_poll_seconds"))


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
        """every look_seconds, a Brobot whose sleep has ended (the sleep tool let go - he woke) and who isn't mid-stir is sent back to sleep."""
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


_SLEEPY = None


def _tag_for(song_id):
    """Song_id-derived print-tag stem, e.g. brobots_baby_robots_sleep ->
    BROBOTS_BABY_ROBOTS_SLEEP. Replaces run_songs_runner_001.py's hardcoded
    "BINGO" literal per the survey's own schema table."""
    return re.sub(r"[^A-Za-z0-9]+", "_", song_id or "golden").upper()


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
    """one page-only chat-feed line from the countdown role (gopod_brobots_settings.json) with the whole seconds; the page counts it down and removes it."""
    role = load_brobots_settings()["countdown"]["role"]
    return post_chat_line(mod, live, role, str(max(1, int(round(seconds)))), robots)


def post_bingo_ball(mod, live, ball, robots=None):
    """one page-only chat-feed line from the bingo_ball role (gopod_brobots_settings.json) with the called number; the page shows a big ball and keeps it in the chat window."""
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
    """the rattle on this machine's speaker (paplay, the default sink - GOPOD's own routing check keeps that the real speaker)."""
    try:
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
    """a volume knob as a share of the master (chat_audio.volume) - 1 for "" or "default", 0.3 for "30%"; a bare number is a share (0.3)."""
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
    """the rattle is put in the chat feed as a reporter-voice line, so the Brobots chat page plays it where its other voices play, in its queue, with the countdown shown for exactly as long."""
    rv = load_brobots_settings()["reporter_voices"]
    try:
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
    """when the rattle has really finished, one more countdown-role line (countdown.end_text) tells the page to remove the timer, so it shows exactly as long as the sound plays."""
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
    """on (rattle_on in the song's knobs.json): "page" - the Brobots chat page plays it like a reporter voice; "robot" - Brobot 1 plays it over the direct SDK after handing back control; "speaker" - this machine's speaker plays it while the robots stay under the song's control."""
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
    """stop the wheels (move_wheels lw=0 rw=0 - the same stop the wheel-nudge alias ends with, brobots.sh)."""
    if not live:
        return {"ok": True, "mode": "dry"}
    try:
        result = mod.wirepod_web_send_form("/api-sdk/move_wheels", {"serial": serial, "lw": "0", "rw": "0"}, timeout=10)
    except Exception as exc:  # noqa: BLE001 - a hold, never stops the song
        result = {"error": repr(exc)}
    return {"ok": result.get("status") == 200, "result": result}


def post_chat_line(mod, live, name, text, robots=None):
    """a chat-page-only line - nothing is said."""
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
    """The cube blip on a Brobot's own cube (phcal's cmd_cube, as-is)."""
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
    """the cube tool's connect mode in the background - the robot connects his cube over Bluetooth (about 18 s, ConnectCube) while the song goes on, so the later blip finds it connected."""
    if not live:
        return {"mode": "dry", "serial": serial}
    env = dict(os.environ, WIREPOD_HOME=WIREPOD_HOME_PATH)
    proc = subprocess.Popen([str(CUBE_BLIP_BIN_PATH), serial, "connect"], stdout=subprocess.PIPE,
                            stderr=subprocess.STDOUT, text=True, env=env)
    return {"proc": proc, "serial": serial, "started": time.time()}


def start_cube_light(live, serial, most_seconds=60.0, after_seconds=0.0):
    """the cube tool's light-on-signal mode in the background - the robot connects his cube now (about 12 s) and holds it; light_cube() then lights it at once."""
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
    """before_speech: called right before the together line is said, so its chat bubbles show just ahead of the voices."""
    if not live:
        if before_speech:
            before_speech()
        return {"ok": True, "mode": "dry"}
    if rattle_before and rattle_on == "page":
        play_rattle_on_page(mod, live, None)
        post_countdown_end(mod, live, None)
        rattle_before = False
    if rattle_before and rattle_on == "speaker":
        post_countdown(mod, live, rattle_wav_seconds(), None)
        play_rattle_on_speaker()
        post_countdown_end(mod, live, None)
        rattle_before = False
    if rattle_before:
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
    if suppress_assume_release and result.get("mode") == "wirepod" and elapsed < SAY_SILENT_SECONDS:
        serial = mod.BROBOT_1_SERIAL if name == "Brobot 1" else mod.BROBOT_2_SERIAL
        print(f"{robots.timing_prefix()}SAY_SILENT {name}: line came back in {elapsed:.2f}s - retaking control and saying it again", flush=True)
        release = mod.wirepod_web_send_form("/api-sdk/release_behavior_control", {"serial": serial}, timeout=10)
        time.sleep(DIRECT_SDK_RELEASE_SETTLE_SECONDS)
        assume = mod.wirepod_web_send_form("/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10)
        time.sleep(POST_ASSUME_SAY_SETTLE_SECONDS)
        retry_started = time.monotonic()
        result = robots.say(
            name, robot_safe_line, display_text="", display_role=name,
            suppress_assume_release=suppress_assume_release,
        )
        retry_elapsed = time.monotonic() - retry_started
        heal = {"first_seconds": round(elapsed, 3), "retry_seconds": round(retry_elapsed, 3), "release": release, "assume": assume}
        print(f"{robots.timing_prefix()}SAY_SILENT {name}: retry took {retry_elapsed:.2f}s{' - STILL SILENT' if retry_elapsed < SAY_SILENT_SECONDS else ''}", flush=True)
    ok = result.get("spoken") is True or result.get("mode") in ("dry", "off")
    return {"ok": ok, "result": result, "silent_heal": heal}


BROBOTS_SETTINGS_DIR = (
    Path(__file__).resolve().parents[3] / "wire_pod_overlay" / "chipper" / "webroot" / "gopod_settings"
)


BROBOTS_SETTINGS_FULL = BROBOTS_SETTINGS_DIR.parent / "gopod_settings_full.json"


def load_brobots_settings():
    try:
        return json.loads(BROBOTS_SETTINGS_FULL.read_text(encoding="utf-8"))
    except (OSError, ValueError):
        print("GOLDEN_SETTINGS no gopod_settings_full.json yet (Wire-Pod writes it at its start) - "
              "the settings folder alone, without the topic-clusters", flush=True)
    merged = {}
    for f in sorted(BROBOTS_SETTINGS_DIR.glob("*.json")):
        merged.update(json.loads(f.read_text(encoding="utf-8")))
    return merged


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
    """while the chat page is away - hidden, closed, or not heard from (Wire-Pod keeps session.page_hidden_file) - the run waits here, and says so in the terminal once, then when it goes on."""
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
    """"upon bingo test live, force llm warm up timer visual, for minimum 5 seconds (or until actual warm up completes), with "Warming Up LLM""): a song with "llm_warmup": true in its knobs.json, on a live run, shows the chat page's warm-up bubble (the LLM power level, then "LLM Ready" - llm.warmup_text / llm.ready_text) while one tiny request to the chat's own model (Wire-Pod's knowledge endpoint and model, nothing else read) makes sure it's loaded."""
    llm = load_brobots_settings()["llm"]
    song_ready_line = (ready_text or "").strip()
    if not llm_required:
        line = "\n".join(t for t in ((llm.get("no_llm_text") or "").strip(), (ready_text or "").strip()) if t)
        if line:
            try:
                answer = mod.wirepod_web_send_form("/api-sdk/gopod_system_say", {"text": line, "wait_seconds": "30"}, timeout=40)
                if isinstance(answer, dict) and str(answer.get("body", "")).strip() == "not_heard":
                    print("LLM_WARMUP no-LLM line not heard on the page", flush=True)
            except Exception as exc:  # noqa: BLE001 - the line never stops the song
                print(f"LLM_WARMUP no-LLM line failed {type(exc).__name__}", flush=True)
        print("LLM_WARMUP no language model required - skipped", flush=True)
        return
    state = mod.GOPOD_CHIPPER_ROOT / "webroot" / "gopod_llm_warmup.json"
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
    lines = [line for line in (lines or []) if (line.get("text") or "").strip()]

    started_ms = int(time.time() * 1000)

    def ready():
        try:
            answer = mod.wirepod_web_send_form("/api-sdk/gopod_system_say", {
                "warmup_ready": str(started_ms), "text": ready_text}, timeout=120)
            if isinstance(answer, dict) and str(answer.get("body", "")).strip() == "not_heard":
                print("LLM_WARMUP ready line not heard on the page", flush=True)
        except Exception as exc:  # noqa: BLE001 - the ready line never stops the song
            print(f"LLM_WARMUP ready line failed {type(exc).__name__}", flush=True)

    def message(text):
        text, wait = (text or "").strip(), (llm.get("wait_text") or "").strip()
        if text and wait and text[-1] not in ".!?":
            text += "."
        if wait and text.endswith(wait):
            return text
        return (text + "\n" + wait).strip()

    def show(text, say=""):
        write(True, text)
        try:
            mod.wirepod_web_send_form("/api-sdk/gopod_system_say", {"warmup_text": str(started_ms), "text": text}, timeout=10)
            say_role = load_brobots_settings()["reporter_voices"].get("system_say_role", "")
            if say and say_role:
                post_chat_line(mod, True, say_role, say)
        except Exception as exc:  # noqa: BLE001 - a message never stops the song
            print(f"LLM_WARMUP message failed {type(exc).__name__}", flush=True)

    def write(active, text=""):
        tmp = state.with_suffix(".json.tmp")
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
    def say_lines():
        for line in lines or []:
            try:
                answer = mod.wirepod_web_send_form("/api-sdk/gopod_system_say", {
                    "who": line.get("who", ""), "text": line.get("text", ""), "wait_seconds": "30"}, timeout=40)
                if isinstance(answer, dict) and str(answer.get("body", "")).strip() == "not_heard":
                    print(f"LLM_WARMUP line not heard on the page: {line.get('who', '')}", flush=True)
            except Exception as exc:  # noqa: BLE001 - a reporter line never stops the song
                print(f"LLM_WARMUP line failed {type(exc).__name__}", flush=True)
    if llm_required and llm.get("smart_warm", True) and base and model and _llm_loaded(base, model):
        line = "\n".join(t for t in (song_ready_line, (llm.get("run_ready_end") or "").strip()) if t)
        print(f"LLM_WARMUP smart: already warm model={model} - loading lines bypassed", flush=True)
        if line:
            try:
                answer = mod.wirepod_web_send_form("/api-sdk/gopod_system_say", {"text": line, "wait_seconds": "30"}, timeout=40)
                if isinstance(answer, dict) and str(answer.get("body", "")).strip() == "not_heard":
                    print("LLM_WARMUP smart ready line not heard on the page", flush=True)
            except Exception as exc:  # noqa: BLE001 - the line never stops the song
                print(f"LLM_WARMUP smart ready line failed {type(exc).__name__}", flush=True)
        say_lines()
        return
    try:
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
            left = float(llm["song_warmup_min_seconds"]) - (time.time() - started)
            if left > 0 and not warm:
                time.sleep(left)
        say_lines()
    finally:
        write(False)
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
    """one narrator bubble per section, posted at its first caption's second."""
    bubbles = []
    for s in song["steps"]:
        if s.get("start_seconds") is None or not s.get("caption"):
            continue
        start = float(s["start_seconds"])
        end = float(s.get("end_seconds", start))
        piece = f"@{start:.2f}-{end:.2f} {s['caption']}"
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
    pictures = song.get("section_pictures") or {}
    single = ["narrator:single"] if song.get("narrator_single") else []
    return [(at, marker.join(pieces + (["picture:/pictures/" + pictures[section]] if pictures.get(section) else []) + single
                             + (["narrator:last"] if single and i == len(bubbles) - 1 else [])))
            for i, (at, section, pieces) in enumerate(bubbles)]


def _start_captions(mod, live, song, timeline_start, stop, robots):
    """Each section's captions (a robots_speak false song's story.md words) posted
    as one bubble from reporter_voices.narrator_role at its first start_seconds,
    in the background."""
    settings = load_brobots_settings()
    role = settings["reporter_voices"].get("narrator_role", "")
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
        result = mod.wirepod_web_send_form("/api-sdk/gopod_run", {"action": action, **form}, timeout=180 if action == "start" else 10)
        reply = json.loads(result.get("body") or "{}")
    except Exception as exc:  # noqa: BLE001 - a session hiccup never stops a run
        print(f"GOLDEN_SESSION {action} not reached: {exc!r}", flush=True)
        return None
    print(f"GOLDEN_SESSION {action} {form.get('name', '')!r} -> {reply}", flush=True)
    return bool(reply.get("ok")) if action == "start" else True


def song_stop_request_path(mod):
    """the voice command creates this file (song_stop.request_file, gopod_brobots_settings.json, in GOPOD's overlay chipper folder); the engine checks it before each step and ends the run there, the normal way - control handed back, touch interrupt on."""
    return mod.GOPOD_CHIPPER_ROOT / load_brobots_settings()["song_stop"]["request_file"]


def touch_interrupt_off_path(mod):
    """touch interrupt is on, as in native Wire-Pod, except while this file exists in GOPOD's overlay chipper folder (touch_interrupt.off_file in gopod_brobots_settings.json, read by kgsim_interrupt.go)."""
    return mod.GOPOD_CHIPPER_ROOT / load_brobots_settings()["touch_interrupt"]["off_file"]


def _scripted_end_re(piece_ends):
    """A piece ends at a run of piece_ends or an ellipsis, plus any closing quote, before a space."""
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
    """The line with its pieces joined by the segment marker."""
    split = load_brobots_settings()["scripted_split"]
    marker = split["segment_marker"]
    lines = [line for line in (text or "").split("\n") if line.strip()]
    return (marker + "\n").join(
        marker.join(scripted_pieces(line, split.get("join_trailing", []), split["piece_ends"])) for line in lines
    )


def run_move_with_line(move, say):
    """a move step that carries a line does the move WHILE the line is said - the move on its own thread, the line here, the step ends when both have."""
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
    """manage_control=False skips this function's own inline assume/release, mirroring run_move_axis()'s already- proven bypass exactly - relies on the caller already holding continuous control."""
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
        spoken = text
        form = {"serial": serial}
        if re.search(r"[A-Za-z0-9]", text or ""):
            if scripted_split:
                text = scripted_line(text)
                form["sentence_highlight"] = "1"
            spoken, _repair_used = mod.normalize_robot_safe(text, robots.scaffold)
            if getattr(robots, "piece_offset", None) is None:
                form.update({"display_text": text, "display_role": name})
            else:
                form["piece_offset"] = str(robots.piece_offset)
                if robots.combo_display:
                    form.update({"display_text": robots.combo_display, "display_role": name})
        _wait_last_animation(mod, serial, entry)
        anim = mod.wirepod_web_send_form(
            "/api-sdk/say_text", {"serial": serial, "text": f"{{{{playAnimationWI||{animation_token}}}}}"}, timeout=90
        )
        entry["animation_dispatch"] = anim
        time.sleep(float(load_brobots_settings()["chat_moves"].get("speech_after_animation_seconds", 0.9)))
        if re.search(r"[A-Za-z0-9]", text or ""):
            _cue, words = mod.split_robot_safe_line(spoken, robots.scaffold)
            form["text"] = f"{load_animation_lead_in()} {words}"
            said = mod.wirepod_web_send_form("/api-sdk/say_text", form, timeout=90)
            entry["line_dispatch"] = said
            anim = anim if said.get("status") == 200 else said
        print(f"{robots.timing_prefix()}{tag}_BEAT {name} token={animation_token}: animation, then the line {load_brobots_settings()['chat_moves'].get('speech_after_animation_seconds', 0.9)}s later", flush=True)
        time.sleep(0.3)
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


def _wait_last_animation(mod, serial, entry):
    """"Wait!"""
    try:
        waited = mod.wirepod_web_send_form("/api-sdk/gopod_run", {"action": "anim_wait", "serial": serial}, timeout=15)
        entry["waited_for_last_animation"] = waited
    except Exception as exc:  # noqa: BLE001 - best effort: an old Wire-Pod just goes on
        entry["waited_for_last_animation"] = repr(exc)


def run_animation_only(mod, serial, animation_token, hold_seconds, live, manage_control=True):
    """manage_control=False bypass, same shape/reason as run_emotion_beat() above."""
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
        _wait_last_animation(mod, serial, entry)
        anim_payload = f"{{{{playAnimationWI||{animation_token}}}}}"
        dispatches = []
        getout_token = _ANIMATION_GETOUT_TOKENS.get(animation_token)
        if getout_token:
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


_WARN_ON_FAILURE_NOTES = {"rattle", "brobots_ready_together", "arm_cue", "nod", "animation", "picture"}


STEP_NAME_WHO = {"host", "brat", "brobots", "reporter", "goverlord", "narrator"}
STEP_NAME_NOTES = {"say": "say_turn", "nod": "nod", "rattle": "rattle", "gap": "pause",
                   "wake": "wake_both", "ready": "brobots_ready_together",
                   "sleep": "sleep_both", "exit": "exit",
                   "timer": "timer", "sequence": "sequence", "lines": "reporter_lines",
                   "sting": "sting",
                   "line": "pause", "system": "system_line"}
STEP_NAME_THINKING = [("thinking", "")]
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
        elif w == "soft" and nxt == "wake":
            want.append(("wake_sleepy", "")); i += 2
        elif w == "llm" and nxt == "song":
            want.append(("llm_song", "")); i += 2
        elif w == "warm" and nxt in ("wait", "ready"):
            want.append(("warm_" + nxt, "")); i += 2
        elif w == "warm":
            want.append(("warm_message", "")); i += 1
        elif w.startswith("anim-") and len(w) > 5:
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
    if ("say_turn", "") in want and not any(n == "say_turn" for n, _ in plays):
        want.remove(("say_turn", ""))
    unknown = [d for n, d in want if n == "?"]
    match = len(want) == len(plays)
    if match:
        for (wn, wd), (pn, pd) in zip(want, plays):
            pd_low = pd.lower()
            if wn == "?":
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
    """End-of-run promote prompt: zKnobs.json is the workbench, knobs.json only updates on explicit operator say-so - divergence between the two is normal, never auto-reconciled."""
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
        if live and os.getenv("GOPOD_GOLDEN_TOUCH_INTERRUPT_OFF", "1") == "0":
            print("TOUCH_INTERRUPT left on for this run (pha0b: n)", flush=True)
        elif live and load_golden_song(song_dir, _load_control_song_module()).get("touch_interrupt_off"):
            touch_off_file = touch_interrupt_off_path(_load_runner_module())
            touch_off_file.touch()
            print(f"TOUCH_INTERRUPT off for this run ({touch_off_file})", flush=True)
        song = load_golden_song(song_dir, _load_control_song_module())
        for warning in step_name_warnings(song_dir):
            print(f"STEP_NAME_WARNING {warning}", flush=True)
        if live:
            run_name = song.get("chat_intro") or song.get("song_id") or Path(song_dir).name
            checked_in = session_run(_load_runner_module(), "start", name=run_name, kind="song", pid=str(os.getpid()))
            if checked_in is False:
                print(f"GOLDEN_REFUSED {run_name!r}: Goverlord says another run is going (or it isn't LLM Ready yet).", flush=True)
                tag = _tag_for(song.get("song_id"))
                return _finalize({"song_id": song.get("song_id"), "live": live, "voice_destination": voice_destination,
                                  "steps": [], "stopped_early": True, "refused": True}, tag)
        if live:
            wait_for_page(_load_runner_module(), "before the run")
        if live and song.get("llm_warmup") and not song.get("warmup_after"):
            song_llm_warmup(_load_runner_module(), song.get("llm_ready_text", ""), song.get("warmup_lines", []),
                            song.get("voices_required", True), song.get("llm_required", True))
        log = _run_golden_song_steps(song_dir, live, voice_destination, _finalize)
    finally:
        global _SLEEPY
        if _SLEEPY is not None:
            _SLEEPY.wake_all()
            _SLEEPY = None
        if run_name is not None and checked_in is not False:  # checked in, or not reached: always check out
            session_run(_load_runner_module(), "end", name=run_name,
                        reason="stopped" if (log or {}).get("stopped_early", True) else "done")
        if touch_off_file is not None:
            touch_off_file.unlink(missing_ok=True)
            print("TOUCH_INTERRUPT back on", flush=True)
        if sys.stdout is tee:
            sys.stdout = real_stdout
        if not log_file.closed:
            log_file.close()
    if live and not log.get("stopped_early", True) and not os.getenv("GOPOD_GOLDEN_SUPPRESS_PROMOTE"):
        _maybe_promote_knobs(song_dir)
    return log


def _run_golden_song_steps(song_dir, live, voice_destination, _finalize):
    mod = _load_runner_module()
    control_mod = _load_control_song_module()
    song = load_golden_song(song_dir, control_mod)
    song_id = song["song_id"]
    tag = _tag_for(song_id)
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

    robot_param = None
    synthesized_speaker = None
    if config.get("synthesize_speaker"):
        robot_param = os.getenv(GOLDEN_ROBOT_ENV, "1")
        if robot_param not in ("1", "2"):
            print(f"GOLDEN_REFUSED {GOLDEN_ROBOT_ENV}={robot_param!r} must be '1' or '2'", flush=True)
            log = {"song_id": song_id, "live": live, "voice_destination": voice_destination, "steps": [], "stopped_early": True}
            return _finalize(log, tag)
        synthesized_speaker = "brobot_2" if robot_param == "2" else "brobot_1"

    robot_filter = os.getenv(GOLDEN_ROBOT_FILTER_ENV, "both")
    if robot_filter not in ("both", "1", "2"):
        print(f"GOLDEN_REFUSED {GOLDEN_ROBOT_FILTER_ENV}={robot_filter!r} must be 'both', '1', or '2'", flush=True)
        log = {"song_id": song_id, "live": live, "voice_destination": voice_destination, "steps": [], "stopped_early": True}
        return _finalize(log, tag)

    continuous_hold_both = not manage_control and synthesized_speaker is None

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
    robots.sentence_highlight = True
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
    with_next_jobs = []  # steps started with the next one ("with_next")

    def start_with_next(run, entry):
        """a "with_next" step runs beside the step after it (both robots at once); it's waited for once that next step is done."""
        job = {}
        thread = threading.Thread(target=lambda: job.update(run()), daemon=True)
        thread.start()
        with_next_jobs.append((thread, job, entry))
        entry["ok"] = True
        entry["with_next"] = True
    needs_control_retake = set()
    stop_request = song_stop_request_path(mod)
    if live:
        stop_request.unlink(missing_ok=True)  # an old request never stops a new run
    timeline_start = None
    soundtrack_posted = False
    captions_stop = threading.Event()
    cube_job = None
    cube_light = None  # a wake_sleepy's "cube_ready" (start_cube_light)
    for step in song["steps"]:
        step_id = step["step_id"]
        note = note_aliases.get(step["note"], step["note"])
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
                if song.get("audio"):
                    timeline_start = _start_song_soundtrack(mod, live, song, at, robots, tag)
                    soundtrack_posted = True
                _start_captions(mod, live, song, timeline_start, captions_stop, robots)
            elif song.get("audio") and not soundtrack_posted:
                _start_song_soundtrack(mod, live, song, at, robots, tag)
                soundtrack_posted = True
            wait = timeline_start + at - time.time()
            if live and wait > 0:
                time.sleep(wait)
            print(f"{robots.timing_prefix()}{tag}_TIMELINE {step_id} at={at:.3f}s {'late ' + format(-wait, '.3f') + 's' if wait < 0 else 'on time'}", flush=True)
        if live and song.get("timeline") != "srt":
            wait_for_page(mod, f"before {step_id}", stop_request, robots.timing_prefix())
        if live and stop_request.exists():
            stop_request.unlink(missing_ok=True)
            print(f"{robots.timing_prefix()}{tag}_STOP {step_id}: stopped by voice (kill bingo test) before this step", flush=True)
            log["stopped_early"] = True
            log["stopped_by_voice"] = step_id
            captions_stop.set()
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
        print(f"===== {step_id} Notes: {display_tag}", flush=True)
        step_start_wall = time.time()
        print(f"{robots.timing_prefix()}{tag} {step_id}: note={note} speaker={speaker}", flush=True)
        if live and step.get("start_after"):
            time.sleep(float(step["start_after"]))

        if serial in needs_control_retake and note in CONTROL_RETAKE_NOTES and not step_filtered:
            needs_control_retake.discard(serial)
            entry["control_retake"] = retake_robot_control(mod, serial, name, live, robots, tag)

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

        if step.get("show_ball") and not step_filtered:
            entry["show_ball"] = post_bingo_ball(mod, live, step["show_ball"], robots)

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
            if step.get("cube_connect") and entry["ok"]:
                cube_job = start_cube_connect(live, mod.BROBOT_2_SERIAL if step["cube_connect"] == "brobot_2" else mod.BROBOT_1_SERIAL)
                entry["cube_connect"] = "started"

        elif note == "sleep_both":
            if sleepy is None:
                entry["ok"] = False
                entry["error"] = "sleep_both needs \"sleepy\": true in knobs.json"
            else:
                sleepy.sleep_both(float(step.get("look_seconds", 0)) if live else 0)
                if live:
                    time.sleep(step.get("settle_seconds", 6.0))
                entry["ok"] = True

        elif sleepy and sleepy.asleep and note in SLEEPY_MOVE_NOTES and not step.get("text"):
            if note == "arm_cue":
                move = lambda serial=serial, step=step: control_mod.run_arm_cue(
                    mod, serial, step["cycles"], step["hold_seconds"], live,
                    speed=step["speed"], manage_control=False)
            elif note == "nod":
                move = lambda serial=serial, step=step: control_mod.run_nod(
                    mod, serial, step["hold_seconds"], live, speed=step["speed"], manage_control=False)
            else:
                move = lambda serial=serial, step=step: run_animation_only(
                    mod, serial, step.get("animation_token"), step["hold_seconds"], live, manage_control=False)
            sleepy.move(serial, step_id, move)
            entry["ok"] = True
            entry["sleepy"] = "handed to " + name

        elif note == "stir":
            if sleepy is None or not sleepy.asleep:
                entry["ok"] = False
                entry["error"] = "stir needs sleep_both first"
            else:
                sleepy.move(serial, step_id, None, awake=step["hold_seconds"])
                entry["ok"] = True
                entry["sleepy"] = "stir " + name

        elif note == "wake_sleepy":
            if sleepy and sleepy.asleep:
                sleepy.finish()
                entry["sleepy_wake"] = sleepy.wake_all()
            if step.get("cube_ready"):
                cube_light = start_cube_light(live, mod.BROBOT_2_SERIAL if step["cube_ready"] == "brobot_2" else mod.BROBOT_1_SERIAL,
                                              after_seconds=float(step.get("cube_ready_after_seconds", 0)))
                entry["cube_ready"] = "connecting"
            entry["ok"] = True

        elif note == "all_asleep":
            if sleepy and sleepy.asleep:
                sleepy.finish()
                if live:
                    time.sleep(step.get("settle_seconds", 0))
            entry["ok"] = True

        elif note == "system_line":
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
            if step.get("with_next"):
                start_with_next(lambda n=name, t=step["text"]: say_step_line(n, t), entry)
            else:
                result = say_step_line(name, step["text"])
                entry["ok"] = result["ok"]
                entry["result"] = result

        elif note == "thinking":
            if live:
                result = mod.wirepod_web_send_form("/api-sdk/gopod_thinking", {"serial": serial}, timeout=180)
                entry["ok"] = result.get("status") == 200
            else:
                result = {"ok": True, "mode": "dry"}
                entry["ok"] = True
            print(f"{robots.timing_prefix()}{tag}_THINKING {name}: the live thinking process", flush=True)
            entry["result"] = result

        elif note == "rattle":
            if step.get("text"):
                entry["line"] = say_step_line(name, step["text"])
            result = run_rattle(mod, serial, live, robots, tag, song.get("rattle_on", "robot"))
            entry["ok"] = result["ok"] and (entry.get("line") or {}).get("ok", True)
            entry["result"] = result
            if continuous_hold_both and result.get("mode") != "speaker":
                reassume = run_continuous_hold_assume(mod, live)
                entry["continuous_hold_reassume"] = reassume
                entry["ok"] = entry["ok"] and reassume["ok"]
            if result.get("mode") != "speaker":
                entry["wheel_hold"] = run_wheel_hold(mod, serial, live)

        elif note == "brobots_ready_together":
            ready_phrase = step.get("text") or "Brobots ready!"
            if sleepy and sleepy.asleep:
                sleepy.finish()
                entry["sleepy_wake"] = sleepy.wake_all()
            if step.get("cube_after_wake"):
                cube_serial = mod.BROBOT_2_SERIAL if step["cube_after_wake"] == "brobot_2" else mod.BROBOT_1_SERIAL
                if cube_light is not None and cube_light["serial"] == cube_serial:
                    entry["cube"] = light_cube(cube_light)
                    cube_light = None
                else:
                    entry["cube"] = run_cube_blip(mod, live, cube_serial)
                print(f"{robots.timing_prefix()}{tag}_CUBE {entry['cube'].get('output', entry['cube'].get('error', ''))}", flush=True)
            spoken_phrase = mod.apply_pronunciation_safety(ready_phrase) if step.get("pronounce") else ready_phrase
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
            if continuous_hold_both:
                reassume = run_continuous_hold_assume(mod, live)
                entry["continuous_hold_reassume"] = reassume
                entry["ok"] = entry["ok"] and reassume["ok"]

        elif note == "arm_cue":
            def arm_move():
                return control_mod.run_arm_cue(
                    mod, serial, step["cycles"], step["hold_seconds"], live,
                    speed=step["speed"], manage_control=manage_control,
                )
            run = (lambda: run_move_with_line(arm_move, lambda: say_step_line(name, step["text"]))) if step.get("text") else arm_move
            if step.get("with_next"):
                start_with_next(run, entry)
            else:
                result = run()
                entry["ok"] = result["ok"]
                entry["result"] = result

        elif note == "nod":
            def nod_move():
                return control_mod.run_nod(
                    mod, serial, step["hold_seconds"], live,
                    speed=step["speed"], manage_control=manage_control,
                )
            run = (lambda: run_move_with_line(nod_move, lambda: say_step_line(name, step["text"]))) if step.get("text") else nod_move
            if step.get("with_next"):
                start_with_next(run, entry)
            else:
                result = run()
                entry["ok"] = result["ok"]
                entry["result"] = result

        elif note == "wheel_nudge":
            result = control_mod.run_wheel_nudge(
                mod, serial, control_mod.WHEEL_NUDGE_REVERSE_LW, control_mod.WHEEL_NUDGE_REVERSE_RW,
                step["hold_seconds"], live,
                manage_control=manage_control,
            )
            entry["ok"] = result["ok"]
            entry["result"] = result

        elif note == "sound":
            if step.get("with_next"):
                start_with_next(lambda s=step: play_song_sound(mod, live, song_dir, s.get("sound", ""), s.get("volume", 1.0), robots, tag), entry)
            else:
                entry["result"] = play_song_sound(mod, live, song_dir, step.get("sound", ""), step.get("volume", 1.0), robots, tag)
                entry["ok"] = entry["result"]["ok"]

        elif note == "picture":
            entry["result"] = show_song_picture(mod, live, step.get("picture", ""), robots, tag)
            entry["ok"] = entry["result"]["ok"]

        elif note == "emotion_beat":
            result = run_emotion_beat(
                mod, serial, name, step["text"], step.get("animation_token"),
                step["hold_seconds"], live, robots, tag,
                manage_control=manage_control,
                scripted_split=song.get("sentence_split") == "scripted",
            )
            entry["ok"] = result.get("ok", False)
            entry["result"] = result

        elif note == "animation":
            token = step.get("animation_token")
            if step.get("text") and token not in _ANIMATION_GETOUT_TOKENS and token not in _LOOP_ANIMATION_TOKENS:
                result = run_emotion_beat(
                    mod, serial, name, step["text"], token, 0.0, live, robots, tag,
                    manage_control=manage_control,
                    scripted_split=song.get("sentence_split") == "scripted",
                )
            else:
                result = run_animation_only(
                    mod, serial, token, step["hold_seconds"], live,
                    manage_control=manage_control,
                )
                if step.get("text"):
                    entry["line"] = say_step_line(name, step["text"])
                    result = {**result, "ok": result.get("ok", False) and entry["line"].get("ok", False)}
            entry["ok"] = result.get("ok", False)
            entry["result"] = result

        elif note in ("pause", "timer"):
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
            if not live:
                entry["ok"] = True
                entry["mode"] = "dry"
            else:
                result = mod.wirepod_web_send_form(
                    "/api-sdk/assume_behavior_control", {"priority": "high", "serial": serial}, timeout=10
                )
                entry["ok"] = result.get("status") == 200
                entry["result"] = result
                time.sleep(POST_ASSUME_SAY_SETTLE_SECONDS)

        elif note == "say":
            entry["ok"] = True

        elif note == "weather":
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

        if unconditional_text_speak and not entry.get("ok") and step.get("fail"):
            run_say_turn(mod, robots, name, step["fail"], suppress_assume_release=suppress_assume_release)

        if not step.get("with_next") and with_next_jobs:
            for thread, job, together_entry in with_next_jobs:
                thread.join()
                together_entry["ok"] = job.get("ok", False)
                together_entry["result"] = job
            with_next_jobs.clear()
        if continuous_hold_both and note in CONTROL_RETAKE_NOTES and not step_filtered:
            step_result = entry.get("result") if isinstance(entry.get("result"), dict) else {}
            heal = step_result.get("silent_heal")
            still_silent = bool(heal) and heal.get("retry_seconds", 0) < SAY_SILENT_SECONDS
            if not entry.get("ok") or still_silent:
                needs_control_retake.add(serial)
                print(f"{robots.timing_prefix()}{tag}_CONTROL_RETAKE_PENDING {name}: {step_id} {'failed' if not entry.get('ok') else 'still silent'} - control is retaken before its next step", flush=True)
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
    if cube_job is not None:
        log["cube_connect"] = finish_cube_connect(cube_job)
    if cube_light is not None:
        log["cube_light"] = light_cube(cube_light)


    if sleepy:
        sleepy.wake_all()
        log["sleepy_moves"] = sleepy.results
    _release_if_holding(mod, hold_release_serials, live, manage_control)
    robots.close()
    return _finalize(log, tag)


def main():
    voice_destination = os.getenv("GOPOD_VOICE_DESTINATION", "robot")
    live = os.getenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH") == "1"
    song_id = os.getenv(GOLDEN_SONG_ID_ENV) or (None if os.getenv(GOLDEN_SONG_DIR_ENV) else DEFAULT_SONG_ID)
    if song_id:
        found = find_song_dir(song_id)
        if found is None:
            print(f"GOLDEN_REFUSED no song folder has song_id {song_id!r} (song_folders_001.py lists them).", flush=True)
            return None
        song_dir = str(found)
    else:
        song_dir = os.getenv(GOLDEN_SONG_DIR_ENV)
    try:
        song_type = json.loads((Path(song_dir) / "knobs.json").read_text(encoding="utf-8")).get("song_type")
    except (OSError, ValueError):
        song_type = None
    if song_type == "interview":
        print("This is the interview: type \"start interview\" in the Brobots chat to play it.", flush=True)
        return None
    if song_type == "comm_check":
        print("This is the comm check: type \"system-comm-check\" in the Brobots chat to play it.", flush=True)
        return None
    return run_golden_song(song_dir, live, voice_destination)


if __name__ == "__main__":
    main()

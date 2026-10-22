#!/usr/bin/env python3
"""GOPOD's room-mic service for the Brobots chat (operator 2026-10-04): one
process owns the Jetson's USB mic and serves both

  - Record Session's continuous 48 kHz master (record_start / record_stop), and
  - the chat page's numpad push-to-talk (ptt_start / ptt_stop): "Does my
    recording via my mic from the web page work? ... please get it to work".

The USB mic can only be opened once, so the two share one stream: talking by
numpad while a session records works.

Golden guts, reused, not rewritten - the "Is That You" PTT writer
(goverlord/runtime/gopod_layer/web_display/gopod_demo_8011/
gopod_ptt_chat_writer_013.py):
  - the mic pick (resolve_audio_device_by_name, "USB Audio", webcam excluded);
    no fallback to the default input (it can be NoMachine's mic,
    AUDIO_ROUTING_CHECK_001.md);
  - MIN_HOLD_SECONDS (1.0), RELEASE_TAIL_SECONDS (0.9), the 48 kHz master /
    16 kHz Vosk targets, resample_pcm16 run ONCE on the whole capture (never
    per chunk - the writer's STAGE 5 lesson), the Vosk model and its loader,
    and the per-key master WAV archive (kp1_N.wav in the writer's session dir).

Wire-Pod (gopod_mic_service.go) starts it as the desktop user and talks to it
in JSON lines: one command per line on stdin, one reply per line on stdout
(with the command's "id"). Anything else it prints is a log line.
"""

from __future__ import annotations

import json
import queue
import subprocess
import sys
import threading
import time
import wave
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[5]
GOLDEN_WRITER_DIR = REPO_ROOT / "goverlord/runtime/gopod_layer/web_display/gopod_demo_8011"
for path in (REPO_ROOT, GOLDEN_WRITER_DIR):
    if str(path) not in sys.path:
        sys.path.insert(0, str(path))

import gopod_ptt_chat_writer_013 as golden  # noqa: E402

VOSK_MODEL_PATH = golden.DEFAULT_VOSK_MODEL_PATH  # the golden writer's own (its root fixed 2026-10-04)
TARGET_RATE = golden.ARCHIVE_TARGET_SAMPLE_RATE  # 48000
CHAT_RATE = golden.CHAT_TARGET_SAMPLE_RATE  # 16000


def log(text: str) -> None:
    print(f"GOPOD_MIC {text}", flush=True)


def reply(msg: dict) -> None:
    print(json.dumps(msg), flush=True)


class Mic:
    def __init__(self) -> None:
        self.lock = threading.Lock()
        self.stream = None
        self.rate = 0
        self.blocks: queue.Queue[bytes] = queue.Queue()
        self.writer = None
        self.record_wav: wave.Wave_write | None = None
        self.record_paths: tuple[Path, Path] | None = None  # (written, wanted)
        self.capturing = False
        self.capture: list[bytes] = []
        self.capture_key = ""
        self.capture_dir = ""
        self.press = 0.0

    # --- the stream, open only while something needs it ---

    def open(self) -> str:
        if self.stream is not None:
            return ""
        golden.ensure_audio_libs_loaded()
        index, name = golden.resolve_audio_device_by_name(
            golden.AUDIO_SOURCE_NAME_MATCH, exclude_terms=golden.AUDIO_SOURCE_EXCLUDE_TERMS
        )
        sd = golden._sounddevice
        if index is None:
            # 2026-10-21 (live: a mic plugged in after the start was never found -
            # PortAudio lists the devices once): look again, fresh, once.
            try:
                sd._terminate()
                sd._initialize()
            except Exception as exc:  # noqa: BLE001
                log(f"RESCAN_FAILED {exc}")
            index, name = golden.resolve_audio_device_by_name(
                golden.AUDIO_SOURCE_NAME_MATCH, exclude_terms=golden.AUDIO_SOURCE_EXCLUDE_TERMS
            )
            if index is None:
                return "no_mic"
            log(f"RESCAN found {name!r}")

        def callback(indata, frames, time_info, status) -> None:  # noqa: ARG001
            self.blocks.put(bytes(indata))

        kwargs = dict(blocksize=golden.CAPTURE_BLOCKSIZE, dtype="int16", channels=1, device=index, callback=callback)
        try:
            with golden.suppress_fd_stderr():
                stream = sd.RawInputStream(samplerate=TARGET_RATE, **kwargs)
        except Exception as exc:  # noqa: BLE001 - the mic won't run at 48 kHz: its own rate
            log(f"48K_REFUSED {exc} - using the mic's own rate")
            with golden.suppress_fd_stderr():
                stream = sd.RawInputStream(**kwargs)
        stream.start()
        self.stream, self.rate = stream, int(stream.samplerate)
        self.writer = threading.Thread(target=self.drain, daemon=True)
        self.writer.start()
        log(f"OPEN device={name!r} rate={self.rate}")
        return ""

    def close_if_idle(self) -> None:
        with self.lock:
            if self.stream is None or self.record_wav is not None or self.capturing:
                return
            stream, self.stream = self.stream, None
        stream.stop()
        stream.close()
        self.blocks.put(b"")  # wakes the drain thread so it ends
        log("CLOSED")

    def drain(self) -> None:
        while True:
            block = self.blocks.get()
            with self.lock:
                if self.stream is None and not block:
                    return
                if self.record_wav is not None and block:
                    self.record_wav.writeframes(block)
                if self.capturing and block:
                    self.capture.append(block)

    # --- Record Session ---

    def record_start(self, out: str) -> dict:
        err = self.open()
        if err:
            return {"ok": False, "error": err}
        wanted = Path(out)
        written = wanted if self.rate == TARGET_RATE else wanted.with_name(wanted.stem + ".native.wav")
        wav = wave.open(str(written), "wb")
        wav.setnchannels(1)
        wav.setsampwidth(2)
        wav.setframerate(self.rate)
        with self.lock:
            self.record_wav, self.record_paths = wav, (written, wanted)
        log(f"RECORD_START out={written} rate={self.rate}")
        return {"ok": True, "rate": self.rate}

    def record_stop(self) -> dict:
        time.sleep(0.2)  # the blocks already heard
        with self.lock:
            wav, paths = self.record_wav, self.record_paths
            self.record_wav, self.record_paths = None, None
        if wav is None:
            return {"ok": False, "error": "not_recording"}
        wav.close()
        self.close_if_idle()
        written, wanted = paths
        if written != wanted:
            done = subprocess.run(["sox", str(written), str(wanted), "rate", "-v", str(TARGET_RATE)], check=False)
            if done.returncode == 0:
                written.unlink()
            else:
                log(f"RESAMPLE_FAILED sox exit {done.returncode}; kept {written}")
        log(f"RECORD_STOP out={wanted}")
        return {"ok": True}

    # --- push-to-talk ---

    def ptt_start(self, key: str, session_dir: str = "") -> dict:
        err = self.open()
        if err:
            return {"ok": False, "error": err}
        with self.lock:
            self.capture, self.capturing, self.capture_key = [], True, key
            self.capture_dir = session_dir
            self.press = time.monotonic()
        log(f"PTT_START key={key}")
        return {"ok": True}

    def ptt_stop(self) -> dict:
        with self.lock:
            if not self.capturing:
                return {"ok": False, "error": "not_capturing"}
            held = time.monotonic() - self.press
        # The golden release tail: the last word often ends just after the key.
        time.sleep(golden.RELEASE_TAIL_SECONDS)
        with self.lock:
            self.capturing = False
            raw, key, rate, session_dir = b"".join(self.capture), self.capture_key, self.rate, self.capture_dir
            self.capture = []
        self.close_if_idle()
        archive = raw if rate == TARGET_RATE else golden.resample_pcm16(raw, rate, TARGET_RATE)
        counter = key.lower()
        golden._capture_counters[counter] = golden._capture_counters.get(counter, 0) + 1
        if session_dir:
            # 2026-10-11 (operator: "default to write ALL session files ... within
            # that session's folder"): Wire-Pod's session folder, named by the
            # time of the press, then the key and its count.
            master = Path(session_dir) / "numpad" / f"{time.strftime('%H%M%S')}_{counter}_{golden._capture_counters[counter]}.wav"
            master.parent.mkdir(parents=True, exist_ok=True)
        else:
            master = golden.ensure_session_dir() / "masters" / f"{counter}_{golden._capture_counters[counter]}.wav"
        golden.write_archive_wav(master, [archive], TARGET_RATE)
        log(f"PTT_STOP key={key} held={held:.2f} master={master}")
        if held < golden.MIN_HOLD_SECONDS:
            return {"ok": True, "status": "too_short", "held": round(held, 2), "master": str(master)}
        model = golden.ensure_vosk_model_loaded(VOSK_MODEL_PATH)
        recognizer = golden._vosk_KaldiRecognizer(model, CHAT_RATE)
        recognizer.AcceptWaveform(golden.resample_pcm16(raw, rate, CHAT_RATE))
        text = " ".join(str(json.loads(recognizer.FinalResult() or "{}").get("text", "")).split())
        log(f"PTT_TEXT key={key} text={text!r}")
        return {"ok": True, "status": "ok" if text else "nothing_heard", "text": text, "held": round(held, 2), "master": str(master)}


def main() -> int:
    mic = Mic()
    # Load the Vosk model now (about 2 s), so the first capture doesn't wait.
    threading.Thread(target=lambda: golden.ensure_vosk_model_loaded(VOSK_MODEL_PATH), daemon=True).start()
    log("READY")
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        try:
            msg = json.loads(line)
        except json.JSONDecodeError:
            log(f"BAD_COMMAND {line!r}")
            continue
        cmd = msg.get("cmd")
        try:
            if cmd == "record_start":
                out = mic.record_start(str(msg.get("out", "")))
            elif cmd == "record_stop":
                out = mic.record_stop()
            elif cmd == "ptt_start":
                out = mic.ptt_start(str(msg.get("key", "")), str(msg.get("dir", "")))
            elif cmd == "ptt_stop":
                out = mic.ptt_stop()
            else:
                out = {"ok": False, "error": f"unknown command {cmd!r}"}
        except Exception as exc:  # noqa: BLE001 - one bad command must not end the service
            out = {"ok": False, "error": f"{type(exc).__name__}: {exc}"}
        out["id"] = msg.get("id")
        reply(out)
    # Wire-Pod closed stdin: finish a recording cleanly.
    if mic.record_wav is not None:
        mic.record_stop()
    return 0


if __name__ == "__main__":
    sys.exit(main())

# ## EOF

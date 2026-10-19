#!/usr/bin/env python3
"""Reporter voices for the Brobots chat page (2026-10-06).

A persistent Kokoro worker, started once by Wire-Pod (gopod_reporter_voices.go):
the model loads once, then each stdin line - JSON {"voice", "text", "out"} - is
written to the "out" WAV file and acked on stdout with JSON {"ok", "seconds"}.
The chat page plays the file, so the voice comes out of the screen showing the
chat. Synthesis is the interview engine's own (synthesize() in
interview_status_kokoro_announcer_001.py), imported unchanged.
"""
from __future__ import annotations

import importlib.util
import json
import os
import sys
import tempfile
from pathlib import Path

ANNOUNCER = (
    Path(__file__).resolve().parents[5]
    / "goverlord/runtime/data_gomad/robot/kokoro_voice/local_tts_voice_registry_001/interview_status_kokoro_announcer_001.py"
)


def load_announcer():
    spec = importlib.util.spec_from_file_location("gopod_kokoro_announcer", ANNOUNCER)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def synthesize(announcer, pipeline, text, voice, out_wav, speed):
    """The announcer's synthesize(), with the talking speed from Wire-Pod
    (reporter_voices.speed; operator 2026-10-09: "decrease talking speed ever so
    slightly"). Same pipeline call, same sample rate."""
    import numpy as np
    import soundfile as sf

    # 2026-10-18 (operator: "Forced alignment * timing file per voice * the page
    # reading it"): Kokoro times every word and mark as it speaks them; each
    # part's times are moved on by the audio before it, and saved beside the
    # WAV as <name>.words.json - [{"t": text, "s": start, "e": end}, ...], in
    # seconds - for the page's word-by-word reveal.
    parts, words, offset = [], [], 0.0
    for result in pipeline(text, voice=voice, speed=speed, split_pattern=r"\n+"):
        audio = np.asarray(result.audio)
        for token in (getattr(result, "tokens", None) or []):
            if token.start_ts is None or token.end_ts is None:
                continue
            words.append({"t": token.text, "s": round(offset + token.start_ts, 3), "e": round(offset + token.end_ts, 3)})
        parts.append(audio)
        offset += len(audio) / announcer.SAMPLE_RATE
    if not parts:
        raise RuntimeError(f"Kokoro produced no audio for {text!r}")
    out_wav.parent.mkdir(parents=True, exist_ok=True)
    sf.write(str(out_wav), np.concatenate(parts), announcer.SAMPLE_RATE)
    if out_wav.stat().st_size <= 44:
        raise RuntimeError(f"empty voice WAV: {out_wav}")
    # 2026-10-18 (operator: "Your approach seems to look better than kokoro * can
    # GOPOD redo voices as per your timings?"): the words as heard in the made
    # voice - Vosk (Wire-Pod's speech model) listens to it, and each word takes
    # the heard word's times; a word Vosk didn't catch keeps Kokoro's; a mark
    # sits where the word before it ends.
    aligner = "kokoro"
    if VOSK.get("model") is not None:
        try:
            words = heard_times(out_wav, words)
            aligner = "vosk"
        except Exception:  # noqa: BLE001 - Kokoro's times then.
            pass
    out_wav.with_suffix(".words.json").write_text(json.dumps({"aligner": aligner, "words": words}))


# Vosk, loaded once (gopod_align_captions_001.py's way, the same model Wire-Pod hears with).
VOSK = {"model": None}
# GOPOD_HOME: the operator's home (Wire-Pod runs as root, whose home has no Vosk).
HOME = Path(os.environ.get("GOPOD_HOME") or Path.home())
VOSK_MODEL = HOME / "wire-pod/vosk/models/en-US/model"


def load_vosk():
    try:
        try:
            import vosk  # noqa: F401
        except ImportError:
            sys.path.append(str(HOME / ".local/lib/python3.10/site-packages"))
        from vosk import Model, SetLogLevel
        SetLogLevel(-1)
        VOSK["model"] = Model(str(VOSK_MODEL))
    except Exception:  # noqa: BLE001 - no Vosk: Kokoro's times.
        VOSK["model"] = None


def heard_times(wav, words):
    """Kokoro's word list with each word's times from what Vosk hears in the WAV."""
    import re
    import subprocess
    from vosk import KaldiRecognizer
    rec = KaldiRecognizer(VOSK["model"], 16000)
    rec.SetWords(True)
    pcm = subprocess.run(["ffmpeg", "-hide_banner", "-loglevel", "error", "-i", str(wav), "-ac", "1", "-ar", "16000",
                          "-f", "s16le", "-"], capture_output=True, check=True).stdout
    heard = []
    for i in range(0, len(pcm), 8000):
        if rec.AcceptWaveform(pcm[i:i + 8000]):
            heard += json.loads(rec.Result()).get("result", [])
    heard += json.loads(rec.FinalResult()).get("result", [])
    norm = lambda t: re.sub(r"[^a-z0-9]+", "", str(t).lower())
    out, at, last_end = [], 0, 0.0
    for w in words:
        n = norm(w["t"])
        if not n:
            out.append({"t": w["t"], "s": last_end, "e": last_end})
            continue
        hit = None
        for j in range(at, min(at + 4, len(heard))):
            h = norm(heard[j]["word"])
            if h and (h == n or h.startswith(n) or n.startswith(h)):
                hit = j
                break
        if hit is None:
            out.append(w)
            last_end = w["e"]
            continue
        s, e = round(float(heard[hit]["start"]), 3), round(float(heard[hit]["end"]), 3)
        out.append({"t": w["t"], "s": s, "e": e})
        last_end, at = e, hit + 1
    return out


def main() -> int:
    announcer = load_announcer()
    from kokoro import KPipeline

    pipeline = KPipeline(lang_code="a", repo_id="hexgrad/Kokoro-82M")
    load_vosk()
    # One throwaway line first: Kokoro's first line after loading is slow (live
    # 2026-10-06: 18 s for the first read-out, 1-3 s after), so it's spent here.
    with tempfile.TemporaryDirectory(prefix="gopod_reporter_warm_") as tmp:
        try:
            announcer.synthesize(pipeline, "Ready.", "af_bella", Path(tmp) / "warm.wav")
        except Exception:  # noqa: BLE001 - a failed warm-up only means a slower first line.
            pass
    print(json.dumps({"ready": True}), flush=True)
    for raw in sys.stdin:
        raw = raw.strip()
        if not raw:
            continue
        try:
            job = json.loads(raw)
            out = Path(job["out"])
            synthesize(announcer, pipeline, job["text"], job["voice"], out, float(job.get("speed") or 1.03))
            seconds = announcer.wav_duration_seconds(out)
            print(json.dumps({"ok": True, "seconds": seconds}), flush=True)
        except Exception as exc:  # noqa: BLE001 - never leave Wire-Pod waiting.
            print(json.dumps({"ok": False, "error": f"{type(exc).__name__}: {exc}"}), flush=True)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

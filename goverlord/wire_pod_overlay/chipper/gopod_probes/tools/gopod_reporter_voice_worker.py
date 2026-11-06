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
import subprocess
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
    # 2026-11-02: the words are timed later, on the finished file (finish_voice).
    return words


def finish_voice(out_wav, words):
    """2026-11-02 (operator: "timing is off for newly generated voice files * they
    should all go through a signal/volume leveling & vosk for timings"): every
    new voice, after its pitch and robot effect, is levelled (sox norm -3), then
    timed by ear on that finished file - Vosk (Wire-Pod's speech model) listens,
    each word takes the heard word's times (align_heard). Before, the timing ran
    on Kokoro's raw voice, before the effects and with no levelling."""
    try:
        levelled = out_wav.with_name(out_wav.stem + ".level.wav")
        subprocess.run(["sox", str(out_wav), str(levelled), "norm", "-3"], check=True, capture_output=True)
        levelled.replace(out_wav)
    except Exception:  # noqa: BLE001 - no sox: the voice as made.
        pass
    aligner = "kokoro"
    if VOSK.get("model") is not None:
        try:
            words = heard_times(out_wav, words)
            aligner = "vosk"
        except Exception:  # noqa: BLE001 - Kokoro's times then.
            pass
    out_wav.with_suffix(".words.json").write_text(json.dumps({"aligner": aligner, "levelled": True, "words": words}))


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
    return align_heard(words, heard)


def align_heard(words, heard):
    """GOLDEN, the one highlight gear (2026-11-06, operator: "can we please
    template/schema the highlighting to one gear we can look to to fix * no
    patching - remove all patching"): Kokoro's words with Vosk's heard times, one
    rule - the words said as written ("Intel", "battery") are matched to the heard
    words in order (the longest common run, anywhere in the line, no look-ahead
    window, no partial matches); every other word ("404" said four oh four,
    "P-N-F-T" said P N F T) and every mark shares the time between the matched
    words around it, by its length. Kokoro's own clock is used only after the
    last matched word. Replaces the look-ahead, "said another way" and gap rules
    (2026-11-02..04)."""
    import re
    norm = lambda t: re.sub(r"[^a-z0-9]+", "", str(t).lower())
    wi = [i for i, w in enumerate(words) if norm(w["t"])]
    a = [norm(words[i]["t"]) for i in wi]
    b = [norm(h["word"]) for h in heard]
    # The longest common run of words (LCS), in order.
    n, m = len(a), len(b)
    L = [[0] * (m + 1) for _ in range(n + 1)]
    for i in range(n - 1, -1, -1):
        for j in range(m - 1, -1, -1):
            L[i][j] = L[i + 1][j + 1] + 1 if a[i] and a[i] == b[j] else max(L[i + 1][j], L[i][j + 1])
    out = [{"t": w["t"], "s": None, "e": None, "k": max(float(w["e"]) - float(w["s"]), 0.05)} for w in words]
    i = j = 0
    while i < n and j < m:
        if a[i] and a[i] == b[j]:
            k = wi[i]
            out[k]["s"], out[k]["e"] = round(float(heard[j]["start"]), 3), round(float(heard[j]["end"]), 3)
            i, j = i + 1, j + 1
        elif L[i + 1][j] >= L[i][j + 1]:
            i += 1
        else:
            j += 1
    return _share_gaps(out, norm)


def _share_gaps(out, norm):
    """The words between two matched words share that time by their length
    (letters); a mark sits where the word before it ends. After the last matched
    word, each takes its own Kokoro length."""
    i = 0
    while i < len(out):
        if out[i]["s"] is not None:
            i += 1
            continue
        j = i
        while j < len(out) and out[j]["s"] is None:
            j += 1
        start = out[i - 1]["e"] if i > 0 else 0.0
        weights = [len(norm(out[k]["t"])) for k in range(i, j)]
        if j < len(out):
            span = max(0.0, out[j]["s"] - start)
            total = sum(weights) or 1
            per = [span * w / total for w in weights]
        else:
            per = [out[k]["k"] if weights[k - i] else 0.0 for k in range(i, j)]
        at = start
        for k in range(i, j):
            out[k]["s"], out[k]["e"] = round(at, 3), round(at + per[k - i], 3)
            at += per[k - i]
        i = j
    for w in out:
        w.pop("k", None)
    return out


def heard_words(wav):
    """2026-10-24 (the mic's own voice): every word Vosk hears in a recording,
    with its times - the words the line was transcribed from."""
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
    return [{"t": h["word"], "s": round(float(h["start"]), 3), "e": round(float(h["end"]), 3)} for h in heard]


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
            # 2026-10-24: "align" - time a recording's words (the mic's own voice).
            if job.get("align"):
                wav = Path(job["align"])
                words = heard_words(wav) if VOSK.get("model") is not None else []
                wav.with_suffix(".words.json").write_text(json.dumps({"aligner": "vosk", "words": words}))
                print(json.dumps({"ok": True, "seconds": announcer.wav_duration_seconds(wav)}), flush=True)
                continue
            out = Path(job["out"])
            words = synthesize(announcer, pipeline, job["text"], job["voice"], out, float(job.get("speed") or 1.03))
            # 2026-10-26 (operator: "use sox for slight but noticable pitch
            # shift"): a host's pitch_cents (voice_tweaks) - sox's pitch keeps
            # the length, so the word times still fit.
            pitch = float(job.get("pitch_cents") or 0)
            if pitch:
                shifted = out.with_name(out.stem + ".pitch.wav")
                subprocess.run(["sox", str(out), str(shifted), "pitch", f"{pitch:.0f}"], check=True, capture_output=True)
                shifted.replace(out)
            # 2026-10-28 (operator: "robotic voice flavouring, slightly offset"): a
            # host's robot effect (voice_tweaks robot_fx) - the sox chain on a copy,
            # mixed wet over the clean voice, then evened out.
            chain = str(job.get("fx_chain") or "").split()
            wet = float(job.get("fx_wet") or 0)
            if chain and 0 < wet:
                fx = out.with_name(out.stem + ".fx.wav")
                mixed = out.with_name(out.stem + ".mix.wav")
                subprocess.run(["sox", str(out), str(fx), *chain], check=True, capture_output=True)
                subprocess.run(["sox", "-m", "-v", f"{max(0.0, 1 - wet):.2f}", str(out), "-v", f"{min(wet, 1.0):.2f}", str(fx), str(mixed), "norm", "-3"], check=True, capture_output=True)
                fx.unlink(missing_ok=True)
                mixed.replace(out)
            finish_voice(out, words)
            seconds = announcer.wav_duration_seconds(out)
            print(json.dumps({"ok": True, "seconds": seconds}), flush=True)
        except Exception as exc:  # noqa: BLE001 - never leave Wire-Pod waiting.
            print(json.dumps({"ok": False, "error": f"{type(exc).__name__}: {exc}"}), flush=True)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

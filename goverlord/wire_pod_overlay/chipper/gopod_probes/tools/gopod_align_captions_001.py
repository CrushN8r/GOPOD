#!/usr/bin/env python3
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/gopod_align_captions_001.py
"""Forced alignment for a song's narrator (2026-10-18, operator: "Baby brobots song
now (premiere pro project mixdown track) * see the mp3 audio file for similar speech
patterns to highlight by * this mp3 is the narrator only * grab timings from the
audio file * apply forced alignment timings, following the way kokoro does it for
highlighting * discard the audio file * if doable add this to GOPOD").

The words are already known - the song's SRT captions. Vosk (Wire-Pod's own speech
model) listens to the narrator audio and times every word it hears; each caption's
words are then matched to those, in order, within the caption's own seconds (a word
not heard is placed between its neighbours). The result is the same shape as a
Kokoro voice's word times, saved in the song folder as narrator_words.json:

  {"source": "<audio file name>", "captions": {"<start seconds, 2 places>":
      [{"t": word, "s": start, "e": end}, ...]}}

The song engine passes each caption's word times with its narrator piece; the chat
page reveals each word at its time (the same reveal as every voice). The audio is
only listened to - nothing of it is kept.

Usage: gopod_align_captions_001.py <narrator audio> <song folder> [vosk model folder]
"""
import json
import re
import subprocess
import sys
from pathlib import Path

DEFAULT_MODEL = Path.home() / "wire-pod/vosk/models/en-US/model"


def srt_captions(path):
    """[(start, end, text)] from an SRT file."""
    out = []
    blocks = re.split(r"\n\s*\n", path.read_text(encoding="utf-8-sig").strip())
    for block in blocks:
        lines = [l.strip() for l in block.splitlines() if l.strip()]
        times = next((l for l in lines if "-->" in l), None)
        if not times:
            continue
        def sec(t):
            h, m, rest = t.strip().split(":")
            s, ms = rest.replace(".", ",").split(",")
            return int(h) * 3600 + int(m) * 60 + int(s) + int(ms) / 1000
        a, b = times.split("-->")
        text = " ".join(lines[lines.index(times) + 1:])
        out.append((sec(a), sec(b), text))
    return out


def heard_words(audio, model_dir):
    """[(word, start, end)] - everything Vosk hears in the audio, timed."""
    from vosk import Model, KaldiRecognizer, SetLogLevel
    SetLogLevel(-1)
    rec = KaldiRecognizer(Model(str(model_dir)), 16000)
    rec.SetWords(True)
    pcm = subprocess.run(["ffmpeg", "-hide_banner", "-loglevel", "error", "-i", str(audio),
                          "-ac", "1", "-ar", "16000", "-f", "s16le", "-"], capture_output=True, check=True).stdout
    words = []
    def take(res):
        for w in json.loads(res).get("result", []):
            words.append((w["word"], float(w["start"]), float(w["end"])))
    for i in range(0, len(pcm), 8000):
        if rec.AcceptWaveform(pcm[i:i + 8000]):
            take(rec.Result())
    take(rec.FinalResult())
    return words


def norm(t):
    return re.sub(r"[^a-z0-9]+", "", t.lower())


def align(caption_words, start, end, heard):
    """Each caption word's (start, end): the next heard word with the same letters,
    in order, inside the caption's seconds (half a second either side); a word not
    heard goes between its neighbours."""
    window = [h for h in heard if h[1] >= start - 0.5 and h[1] <= end + 0.5]
    times, at = [], 0
    for w in caption_words:
        n, hit = norm(w), None
        for j in range(at, min(at + 4, len(window))):
            if n and (norm(window[j][0]) == n or norm(window[j][0]).startswith(n) or n.startswith(norm(window[j][0]))):
                hit = j
                break
        if hit is None:
            times.append(None)
            continue
        times.append((window[hit][1], window[hit][2]))
        at = hit + 1
    # Not heard: between the times around it (inside the caption).
    for i, t in enumerate(times):
        if t is not None:
            continue
        k = i
        while k < len(times) and times[k] is None:
            k += 1
        a = times[i - 1][1] if i > 0 else start
        b = times[k][0] if k < len(times) else end
        for m in range(i, k):
            s = a + (b - a) * (m - i) / (k - i)
            times[m] = (s, a + (b - a) * (m - i + 1) / (k - i))
    return times


def main():
    if len(sys.argv) < 3:
        print(__doc__)
        return 2
    audio, song = Path(sys.argv[1]), Path(sys.argv[2])
    model = Path(sys.argv[3]) if len(sys.argv) > 3 else DEFAULT_MODEL
    knobs = json.loads((song / "knobs.json").read_text())
    srt = next(iter(sorted(song.rglob("*.srt"))), None)
    if not srt:
        print("no .srt in " + str(song))
        return 1
    heard = heard_words(audio, model)
    out, matched, total = {}, 0, 0
    for start, end, text in srt_captions(srt):
        words = [w for w in text.split() if norm(w)]
        times = align(words, start, end, heard)
        out[f"{start:.2f}"] = [{"t": w, "s": round(s, 3), "e": round(e, 3)} for w, (s, e) in zip(words, times)]
        total += len(words)
        matched += sum(1 for w in words if any(norm(h[0]) == norm(w) and start - 0.5 <= h[1] <= end + 0.5 for h in heard))
    (song / "narrator_words.json").write_text(json.dumps({"source": audio.name, "song": knobs.get("song_id", ""), "captions": out}, indent=1))
    print(f"ALIGNED {len(out)} captions, {matched} of {total} words heard ({len(heard)} heard in the audio) -> {song / 'narrator_words.json'}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

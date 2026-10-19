#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
import tempfile
import time
import wave
from pathlib import Path

OUT_ROOT = Path("/home/goverlord/gopod_tts/tests/interview_status_kokoro_announcer_001")
SAMPLE_RATE = 24000
DEFAULT_VOICE = os.getenv("GOPOD_STATUS_KOKORO_VOICE", "af_bella")
DEFAULT_PLAYER = os.getenv("GOPOD_STATUS_AUDIO_PLAYER", "aplay")

STATUS_LINES = [
    ("RUN_START", "STATUS: RUN_START"),
    ("SCAFFOLD_LOAD", "STATUS: SCAFFOLD_LOAD"),
    ("SCAFFOLD_READY", "STATUS: SCAFFOLD_READY BROBOTS_INTERVIEW_RUNTIME_SCAFFOLD_001"),
    ("SECTION_CARD_LOAD", "STATUS: SECTION_CARD_LOAD"),
    ("SECTION_CARD_READY", "STATUS: SECTION_CARD_READY"),
    ("WIREPOD_ROUTE", "LIVE: Wire-Pod ESN routing enabled"),
    ("BROBOT_1_ESN", "LIVE: Brobot 1 ESN: 0dd1b9e9"),
    ("BROBOT_2_ESN", "LIVE: Brobot 2 ESN: 0dd1d8bf"),
    ("GENERATION_START", "STATUS: GENERATION_START"),
    ("EXCHANGE_GENERATION_START", "STATUS: EXCHANGE_GENERATION_START"),
    ("LLM_START", "STATUS: LLM_START"),
    ("LLM_DONE", "STATUS: LLM_DONE"),
    ("EXCHANGE_GENERATION_DONE", "STATUS: EXCHANGE_GENERATION_DONE"),
    ("GENERATION_DONE", "STATUS: GENERATION_DONE"),
]

STOP_PREFIXES = (
    "STATUS: PLAYBACK_START",
    "STATUS: PLAYBACK_TURN_START",
    "Brobot 1:",
    "Brobot 2:",
)


def status(message: str) -> None:
    print(f"STATUS: {message}", flush=True)


def slugify(label: str) -> str:
    return re.sub(r"[^a-z0-9]+", "_", label.lower()).strip("_")


def out_dir(voice: str) -> Path:
    return OUT_ROOT / voice


def wav_path_for(label: str, voice: str) -> Path:
    return out_dir(voice) / f"{slugify(label)}.wav"


def synthesize(pipeline: KPipeline, text: str, voice: str, out_wav: Path) -> None:
    import numpy as np
    import soundfile as sf

    parts = []
    for _, _, audio in pipeline(text, voice=voice, speed=1.03, split_pattern=r"\n+"):
        parts.append(np.asarray(audio))
    if not parts:
        raise RuntimeError(f"Kokoro produced no audio for {text!r}")
    out_wav.parent.mkdir(parents=True, exist_ok=True)
    sf.write(str(out_wav), np.concatenate(parts), SAMPLE_RATE)
    if out_wav.stat().st_size <= 44:
        raise RuntimeError(f"empty status WAV: {out_wav}")


def generate_assets(voice: str) -> Path:
    from kokoro import KPipeline

    pipeline = KPipeline(lang_code="a", repo_id="hexgrad/Kokoro-82M")
    manifest = {
        "asset_id": "interview_status_kokoro_announcer_001",
        "voice": voice,
        "sample_rate": SAMPLE_RATE,
        "playback": "default system audio through aplay unless GOPOD_STATUS_AUDIO_PLAYER overrides it",
        "stop_before": list(STOP_PREFIXES),
        "items": [],
    }
    for label, spoken_text in STATUS_LINES:
        out_wav = wav_path_for(label, voice)
        synthesize(pipeline, spoken_text, voice, out_wav)
        manifest["items"].append(
            {
                "label": label,
                "match_text": spoken_text,
                "spoken_text": spoken_text,
                "wav": str(out_wav),
                "size_bytes": out_wav.stat().st_size,
            }
        )
        status(f"STATUS_WAV_READY label={label} wav={out_wav}")

    manifest_path = out_dir(voice) / "manifest.json"
    manifest_path.write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
    status(f"STATUS_MANIFEST_WRITTEN {manifest_path}")
    return manifest_path


def speak_stdin(voice: str, player: str) -> int:
    """Persistent synthesis worker: read one status line per stdin line,
    synthesize it live with Kokoro, play it blocking, ack with 'DONE' on
    stdout so the caller can stay in sync. Exits cleanly on EOF."""
    from kokoro import KPipeline

    pipeline = KPipeline(lang_code="a", repo_id="hexgrad/Kokoro-82M")
    scratch_dir = Path(tempfile.mkdtemp(prefix="gopod_status_speak_"))
    try:
        for raw_line in sys.stdin:
            text = raw_line.rstrip("\n")
            if not text:
                print("DONE", flush=True)
                continue
            try:
                out_wav = scratch_dir / f"{slugify(text)[:80] or 'line'}.wav"
                synthesize(pipeline, text, voice, out_wav)
                play_wav_blocking(out_wav, player)
            except Exception as exc:  # noqa: BLE001 - never hang the caller waiting for DONE.
                print(f"STATUS: KOKORO_SPEAK_STDIN_ERROR {type(exc).__name__}: {exc}", file=sys.stderr, flush=True)
            finally:
                print("DONE", flush=True)
    finally:
        import shutil

        shutil.rmtree(scratch_dir, ignore_errors=True)
    return 0


def load_manifest(voice: str) -> dict:
    manifest_path = out_dir(voice) / "manifest.json"
    if not manifest_path.is_file():
        raise RuntimeError(f"status WAV manifest missing: {manifest_path}. Run --generate first.")
    return json.loads(manifest_path.read_text(encoding="utf-8"))


def match_item(line: str, manifest: dict) -> dict | None:
    stripped = line.strip()
    for item in manifest.get("items", []):
        if stripped.startswith(item["match_text"]):
            return item
    return None


def wav_duration_seconds(path: Path) -> float:
    try:
        with wave.open(str(path), "rb") as handle:
            return handle.getnframes() / float(handle.getframerate())
    except wave.Error:
        import soundfile as sf

        info = sf.info(str(path))
        return info.frames / float(info.samplerate)


def aplay_ready_wav(source: Path, target: Path) -> None:
    import soundfile as sf

    audio, samplerate = sf.read(str(source), always_2d=False)
    sf.write(str(target), audio, samplerate, subtype="PCM_16")


def play_wav_blocking(path: Path, player: str) -> None:
    duration = wav_duration_seconds(path)
    started = time.monotonic()
    with tempfile.TemporaryDirectory(prefix="gopod_status_audio_") as tmpdir:
        playback_path = Path(tmpdir) / path.name
        aplay_ready_wav(path, playback_path)
        subprocess.run([player, str(playback_path)], check=False, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    remaining = duration - (time.monotonic() - started)
    if remaining > 0:
        time.sleep(remaining)


def play_all(voice: str, player: str, limit: int = 0) -> None:
    manifest = load_manifest(voice)
    items = manifest.get("items", [])
    if limit > 0:
        items = items[:limit]
    for item in items:
        status(f"STATUS_AUDIO_PLAY label={item['label']} wav={item['wav']}")
        play_wav_blocking(Path(item["wav"]), player)

def run_wrapped(command: list[str], voice: str, player: str) -> int:
    if not command:
        raise RuntimeError("--run requires a command after --")

    manifest = load_manifest(voice)
    announce_enabled = True
    generation_done = False

    proc = subprocess.Popen(
        command,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
        bufsize=1,
    )

    assert proc.stdout is not None

    for line in proc.stdout:
        print(line, end="", flush=True)
        stripped = line.strip()

        if "GENERATION_DONE" in stripped:
            generation_done = True
            continue

        if generation_done:
            continue

        if announce_enabled:
            item = match_item(stripped, manifest)
            if item is not None:
                play_wav_blocking(Path(item["wav"]), player)

    time.sleep(2.0)  # final status WAV wait
    return proc.wait()

def parse_args(argv: list[str] | None = None) -> tuple[argparse.Namespace, list[str]]:
    parser = argparse.ArgumentParser(description="Kokoro status announcer for the Brobots interview pre-chat phase.")
    parser.add_argument("--voice", default=DEFAULT_VOICE, help="Kokoro voice to use for status narration.")
    parser.add_argument("--player", default=DEFAULT_PLAYER, help="System audio player command. Default: aplay.")
    parser.add_argument("--generate", action="store_true", help="Generate status WAV assets.")
    parser.add_argument("--play-all", action="store_true", help="Play all generated status WAVs through default system audio.")
    parser.add_argument("--limit", type=int, default=0, help="Limit --play-all to the first N assets.")
    parser.add_argument("--run", action="store_true", help="Run an interview command and announce matching startup status lines.")
    parser.add_argument("--speak-stdin", action="store_true", help="Persistent worker: synthesize+speak each stdin line live, ack DONE per line.")
    args, command = parser.parse_known_args(argv)
    if command and command[0] == "--":
        command = command[1:]
    args.parser = parser
    return args, command


def main() -> int:
    args, command = parse_args()

    if args.speak_stdin:
        return speak_stdin(args.voice, args.player)
    if args.generate:
        generate_assets(args.voice)
    if args.play_all:
        play_all(args.voice, args.player, args.limit)
    if args.run:
        if not command:
            raise RuntimeError("--run requires a command after --")
        return run_wrapped(command, args.voice, args.player)
    if not args.generate and not args.play_all:
        args.parser.print_help()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

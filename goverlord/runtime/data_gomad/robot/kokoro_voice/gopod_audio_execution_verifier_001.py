#!/usr/bin/env python3
"""Minimum viable GOPOD audio loop: LLM text -> speech -> heard proof."""

from __future__ import annotations

import argparse
import json
import select
import shutil
import subprocess
import sys
import time
import urllib.parse
import urllib.request
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Callable


REPO_ROOT = Path(__file__).resolve().parents[5]
DEFAULT_LOG_PATH = REPO_ROOT / "gopod_audio_execution_log.jsonl"
PIPER_COMMAND = Path("/home/goverlord/.local/bin/piper")
FALLBACK_VOICE = Path("/home/goverlord/gopod_tts/piper_voices/en_US-lessac-medium.onnx")
FALLBACK_WAV = Path("/tmp/gopod_audio_execution_fallback.wav")

AudioRunner = Callable[[str, "AudioExecutionConfig"], dict[str, Any]]


@dataclass
class AudioExecutionConfig:
    log_path: Path = DEFAULT_LOG_PATH
    wirepod_url: str = ""
    wirepod_esn: str = ""
    heard: bool | None = None
    heard_timeout_seconds: float = 20.0
    wirepod_runner: AudioRunner | None = None
    fallback_runner: AudioRunner | None = None


def extract_speech(llm_output: Any) -> str:
    """Point B contract: speech_text equals the raw LLM output string."""
    return str(llm_output or "").strip()


def run_audio_execution(llm_output: Any, config: AudioExecutionConfig | None = None) -> dict[str, Any]:
    config = config or AudioExecutionConfig()
    speech = extract_speech(llm_output)
    if not speech:
        result = _result("BLOCKED", speech, route="", heard_audio=False, playback_status="not_attempted")
        _append_log(config.log_path, _audio_verification_event(speech, "", "not_attempted", False))
        return result

    wirepod_runner = config.wirepod_runner or wirepod_say_text
    fallback_runner = config.fallback_runner or fallback_tts_playback

    route_result: dict[str, Any]
    if config.wirepod_url:
        route_result = wirepod_runner(speech, config)
    else:
        route_result = fallback_runner(speech, config)

    route = str(route_result.get("route", "wire-pod" if config.wirepod_url else "fallback"))
    route_ok = route_result.get("status") == "OK"
    heard = _operator_heard(config) if route_ok else False
    playback_status = "success" if route_ok else "fail"
    status = "PASS" if route_ok and heard is True else "BLOCKED"
    result = _result(status, speech, route=route, heard_audio=heard is True, playback_status=playback_status)

    result["route_result"] = route_result
    _append_log(config.log_path, _audio_verification_event(speech, route, playback_status, heard is True))
    return result


def wirepod_say_text(speech: str, config: AudioExecutionConfig) -> dict[str, Any]:
    base = config.wirepod_url.rstrip("/")
    if not base:
        return {"status": "BLOCKED", "route": "wirepod", "detail": "wirepod_url_missing"}
    try:
        _wirepod_get(base, "/api-sdk/assume_behavior_control", config.wirepod_esn)
        _wirepod_get(base, "/api-sdk/say_text", config.wirepod_esn, {"text": speech})
        _wirepod_get(base, "/api-sdk/release_behavior_control", config.wirepod_esn)
    except Exception as exc:  # noqa: BLE001 - exact failure is logged for operator action.
        return {"status": "BLOCKED", "route": "wire-pod", "detail": f"{type(exc).__name__}: {exc}"}
    return {"status": "OK", "route": "wire-pod", "detail": "say_text_complete"}


def fallback_tts_playback(speech: str, config: AudioExecutionConfig) -> dict[str, Any]:
    if not PIPER_COMMAND.is_file() or not FALLBACK_VOICE.is_file():
        return {"status": "BLOCKED", "route": "fallback", "detail": "missing_piper_or_voice"}
    player = shutil.which("aplay")
    if not player:
        return {"status": "BLOCKED", "route": "fallback", "detail": "missing_aplay"}
    try:
        synth = subprocess.run(
            [str(PIPER_COMMAND), "--model", str(FALLBACK_VOICE), "--output_file", str(FALLBACK_WAV)],
            input=speech,
            text=True,
            capture_output=True,
            timeout=30.0,
            check=False,
        )
        if synth.returncode != 0:
            detail = " ".join((synth.stderr or synth.stdout or "").split())[:300]
            return {"status": "BLOCKED", "route": "fallback", "detail": f"piper_failed:{detail}"}
        playback = subprocess.run([player, str(FALLBACK_WAV)], capture_output=True, text=True, timeout=30.0, check=False)
        if playback.returncode != 0:
            detail = " ".join((playback.stderr or playback.stdout or "").split())[:300]
            return {"status": "BLOCKED", "route": "fallback", "detail": f"aplay_failed:{detail}"}
    except Exception as exc:  # noqa: BLE001
        return {"status": "BLOCKED", "route": "fallback", "detail": f"{type(exc).__name__}: {exc}"}
    return {"status": "OK", "route": "fallback", "detail": str(FALLBACK_WAV)}


def _wirepod_get(base: str, endpoint: str, esn: str, params: dict[str, str] | None = None) -> None:
    query = dict(params or {})
    if esn:
        query["serial"] = esn
    url = base + endpoint
    if query:
        url += "?" + urllib.parse.urlencode(query)
    with urllib.request.urlopen(url, timeout=10.0) as response:  # noqa: S310 - operator-provided local Wire-Pod URL.
        if response.status >= 400:
            raise RuntimeError(f"http_status:{response.status}")


def _operator_heard(config: AudioExecutionConfig) -> bool | None:
    if config.heard is not None:
        return bool(config.heard)
    print("HEARD_AUDIO? (true/false)", flush=True)
    readable, _, _ = select.select([sys.stdin], [], [], config.heard_timeout_seconds)
    if not readable:
        return None
    response = sys.stdin.readline().strip().lower()
    if response in {"heard=true", "true", "yes", "y"}:
        return True
    if response in {"heard=false", "false", "no", "n"}:
        return False
    return None


def _result(status: str, speech: str, *, route: str, heard_audio: bool, playback_status: str) -> dict[str, Any]:
    final_audio_gate = "PASS" if playback_status == "success" and heard_audio else "FAIL"
    return {
        "event": "gopod_audio_execution",
        "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "status": status,
        "speech": speech,
        "route": route,
        "heard_audio": heard_audio,
        "playback_status": playback_status,
        "final_audio_gate": final_audio_gate,
    }


def _audio_verification_event(speech: str, route: str, playback_status: str, heard_audio: bool) -> dict[str, Any]:
    return {
        "event": "audio_verification",
        "tts_engine": route,
        "playback_status": playback_status,
        "heard_audio": heard_audio,
        "final_audio_gate": "PASS" if playback_status == "success" and heard_audio else "FAIL",
        "speech": speech,
        "route": route,
        "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
    }


def _append_log(path: Path, result: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a", encoding="utf-8") as handle:
        handle.write(json.dumps(result, sort_keys=True) + "\n")


def _read_input(args: argparse.Namespace) -> str:
    if args.text:
        return args.text
    if args.input:
        return Path(args.input).read_text(encoding="utf-8")
    return sys.stdin.read()


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Route LLM output to audible speech and require operator-heard proof.")
    parser.add_argument("--text", default="", help="LLM output or SPEECH= packet text.")
    parser.add_argument("--input", default="", help="File containing LLM output.")
    parser.add_argument("--wirepod-url", default="", help="Wire-Pod base URL. If unavailable, fallback TTS is used.")
    parser.add_argument("--wirepod-esn", default="", help="Optional Wire-Pod robot serial/ESN.")
    parser.add_argument("--heard", choices=("true", "false"), help="Operator-heard hook for deterministic runs.")
    parser.add_argument("--heard-timeout", type=float, default=20.0)
    parser.add_argument("--log-path", type=Path, default=DEFAULT_LOG_PATH)
    args = parser.parse_args(argv)

    heard = None if args.heard is None else args.heard == "true"
    config = AudioExecutionConfig(
        log_path=args.log_path,
        wirepod_url=args.wirepod_url,
        wirepod_esn=args.wirepod_esn,
        heard=heard,
        heard_timeout_seconds=args.heard_timeout,
    )
    result = run_audio_execution(_read_input(args), config)
    print(json.dumps(result, indent=2, sort_keys=True))
    return 0 if result.get("status") == "PASS" else 2


if __name__ == "__main__":
    raise SystemExit(main())

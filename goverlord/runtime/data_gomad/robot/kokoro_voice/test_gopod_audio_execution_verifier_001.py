#!/usr/bin/env python3
"""Tests for the minimum viable GOPOD audio loop."""

from __future__ import annotations

import json
import tempfile
import unittest
from pathlib import Path

from goverlord.runtime.voice.gopod_audio_execution_verifier_001 import (
    AudioExecutionConfig,
    run_audio_execution,
)


class GOPODAudioExecutionVerifierTests(unittest.TestCase):
    def test_wirepod_success_requires_heard_true_and_logs_audio_verification_gate(self) -> None:
        calls: list[tuple[str, str]] = []

        def wirepod(text: str, config: AudioExecutionConfig) -> dict[str, object]:
            calls.append(("wirepod", text))
            return {"status": "OK", "route": "wire-pod"}

        with tempfile.TemporaryDirectory() as tmp:
            log_path = Path(tmp) / "gopod_audio_execution_log.jsonl"
            result = run_audio_execution(
                "SPEECH=Wire pod speaks.",
                AudioExecutionConfig(
                    log_path=log_path,
                    wirepod_url="http://wirepod.local:8080",
                    heard=True,
                    wirepod_runner=wirepod,
                ),
            )

            self.assertEqual(result["status"], "PASS")
            self.assertEqual(result["route"], "wire-pod")
            self.assertEqual(result["final_audio_gate"], "PASS")
            self.assertEqual(calls, [("wirepod", "SPEECH=Wire pod speaks.")])
            event = json.loads(log_path.read_text(encoding="utf-8").strip())
            self.assertEqual(
                set(event),
                {"event", "tts_engine", "playback_status", "heard_audio", "final_audio_gate", "speech", "route", "timestamp"},
            )
            self.assertEqual(event["event"], "audio_verification")
            self.assertEqual(event["speech"], "SPEECH=Wire pod speaks.")
            self.assertEqual(event["route"], "wire-pod")
            self.assertEqual(event["tts_engine"], "wire-pod")
            self.assertEqual(event["playback_status"], "success")
            self.assertTrue(event["heard_audio"])
            self.assertEqual(event["final_audio_gate"], "PASS")

    def test_no_secondary_routing_when_wirepod_route_is_selected(self) -> None:
        calls: list[tuple[str, str]] = []

        def wirepod(text: str, config: AudioExecutionConfig) -> dict[str, object]:
            calls.append(("wirepod", text))
            return {"status": "BLOCKED", "route": "wire-pod"}

        def fallback(text: str, config: AudioExecutionConfig) -> dict[str, object]:
            calls.append(("fallback", text))
            return {"status": "OK", "route": "fallback"}

        with tempfile.TemporaryDirectory() as tmp:
            result = run_audio_execution(
                "Wirepod only.",
                AudioExecutionConfig(
                    log_path=Path(tmp) / "gopod_audio_execution_log.jsonl",
                    wirepod_url="http://wirepod.local:8080",
                    heard=True,
                    wirepod_runner=wirepod,
                    fallback_runner=fallback,
                ),
            )

            self.assertEqual(result["status"], "BLOCKED")
            self.assertEqual(result["route"], "wire-pod")
            self.assertEqual(result["final_audio_gate"], "FAIL")
            self.assertEqual(calls, [("wirepod", "Wirepod only.")])

    def test_fallback_route_uses_heard_flag_as_only_success_condition(self) -> None:
        def fallback(text: str, config: AudioExecutionConfig) -> dict[str, object]:
            return {"status": "OK", "route": "fallback"}

        with tempfile.TemporaryDirectory() as tmp:
            pass_result = run_audio_execution(
                "Fallback speaks.",
                AudioExecutionConfig(
                    log_path=Path(tmp) / "pass.jsonl",
                    heard=True,
                    fallback_runner=fallback,
                ),
            )
            blocked_result = run_audio_execution(
                "Fallback speaks.",
                AudioExecutionConfig(
                    log_path=Path(tmp) / "blocked.jsonl",
                    heard=False,
                    fallback_runner=fallback,
                ),
            )

            self.assertEqual(pass_result["status"], "PASS")
            self.assertEqual(pass_result["final_audio_gate"], "PASS")
            self.assertEqual(blocked_result["status"], "BLOCKED")
            self.assertEqual(blocked_result["heard_audio"], False)
            self.assertEqual(blocked_result["final_audio_gate"], "FAIL")

    def test_empty_text_blocks_before_route(self) -> None:
        calls: list[str] = []

        def fallback(text: str, config: AudioExecutionConfig) -> dict[str, object]:
            calls.append(text)
            return {"status": "OK", "route": "fallback"}

        with tempfile.TemporaryDirectory() as tmp:
            result = run_audio_execution(
                "   ",
                AudioExecutionConfig(
                    log_path=Path(tmp) / "gopod_audio_execution_log.jsonl",
                    heard=True,
                    fallback_runner=fallback,
                ),
            )

            self.assertEqual(result["status"], "BLOCKED")
            self.assertEqual(calls, [])


if __name__ == "__main__":
    unittest.main()

import importlib.util
import unittest
from pathlib import Path

# Sibling-relative, same convention run_section1_preshow_generate_001.py
# already uses - portable regardless of where the interview/ engine folder
# itself is checked out.
RUNNER_PATH = Path(__file__).resolve().parent.parent / "tools" / "run_section1_full_live_001.py"


def _load_runner():
    spec = importlib.util.spec_from_file_location("run_section1_full_live_001_under_test", str(RUNNER_PATH))
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class EchoDetectionUnitTest(unittest.TestCase):
    """Tests the mechanical N-word overlap helper directly."""

    @classmethod
    def setUpClass(cls):
        cls.mod = _load_runner()

    def test_detect_echo_finds_verbatim_overlap_case_insensitive_punctuation_ignored(self):
        source = "That sounds like a Healthy Distraction!"
        candidate = "Yeah, THAT SOUNDS LIKE A healthy-distraction, for sure!!"
        snippet = self.mod.detect_echo(candidate, source, n=5)
        self.assertIsNotNone(snippet)
        self.assertEqual(snippet, "that sounds like a healthy")

    def test_detect_echo_returns_none_for_clean_response(self):
        source = "That sounds like a Healthy Distraction!"
        candidate = "Absolutely, this fits perfectly for community events and shelters."
        self.assertIsNone(self.mod.detect_echo(candidate, source, n=5))

    def test_detect_echo_respects_word_count_constant(self):
        # Only 4 words overlap ("sounds like a healthy") - below the default N=5, should not match.
        source = "That sounds like a Healthy Distraction!"
        candidate = "Honestly it sounds like a healthy pick for tonight."
        self.assertIsNone(self.mod.detect_echo(candidate, source, n=self.mod.ECHO_DETECTION_WORD_COUNT))
        # Same pair matches at a lower N.
        self.assertIsNotNone(self.mod.detect_echo(candidate, source, n=4))


class EchoDetectionExchangeRecordTest(unittest.TestCase):
    """Tests the guard wired into generate_exchange_record(), with
    llm_colour_line monkeypatched so no real Ollama call is made."""

    @classmethod
    def setUpClass(cls):
        cls.mod = _load_runner()
        cls.scaffold = cls.mod.load_interview_scaffold()

    def _run_with_fake_llm(self, line, brobot_2_reply, brobot_1_reply):
        mod = self.mod

        def fake_llm_colour_line(role, line_arg, value_target="", prior_context="", scaffold=None, content_model=None):
            return brobot_2_reply if role == "BROBOT 2" else brobot_1_reply

        original = mod.llm_colour_line
        mod.llm_colour_line = fake_llm_colour_line
        try:
            return mod.generate_exchange_record(line, prior_context="", scaffold=self.scaffold)
        finally:
            mod.llm_colour_line = original

    def test_echoing_brobot_1_response_is_flagged_against_visible_line(self):
        line = {
            "line_number": 4,
            "line_type": "STATEMENT",
            "exchange_type": "C&A crystallizer-flavoured",
            "visible_line": "That sounds like a Healthy Distraction!",
            "silent_value_cue": "GOPOD helps bars, shelters, and community spaces.",
        }
        record = self._run_with_fake_llm(
            line,
            brobot_2_reply="thinking... So this is basically a healthy distraction for the room, right?",
            brobot_1_reply="celebrate... That sounds like a Healthy Distraction, and it really is!",
        )
        self.assertTrue(record.get("echo_detected"))
        self.assertEqual(record.get("echo_source"), "visible_line")
        self.assertTrue(record.get("echo_snippet"))

    def test_clean_brobot_1_response_is_not_flagged(self):
        line = {
            "line_number": 4,
            "line_type": "STATEMENT",
            "exchange_type": "C&A crystallizer-flavoured",
            "visible_line": "That sounds like a Healthy Distraction!",
            "silent_value_cue": "GOPOD helps bars, shelters, and community spaces.",
        }
        record = self._run_with_fake_llm(
            line,
            brobot_2_reply="thinking... So this is basically a healthy distraction for the room, right?",
            brobot_1_reply="celebrate... Absolutely, this hits bars, shelters, and community spaces perfectly.",
        )
        self.assertFalse(record.get("echo_detected"))
        self.assertNotIn("echo_source", record)
        self.assertNotIn("echo_snippet", record)

    def test_canned_brobot_1_turn_skips_echo_check_entirely(self):
        # Closer exchange type: brobot_1_mode is "canned" - generated_turn()
        # never calls llm_colour_line for Brobot 1 on this line at all.
        line = {
            "line_number": 5,
            "line_type": "QUESTION",
            "exchange_type": "Closer",
            "visible_line": "Thank you. Interview over. What have we learned about Healthy Distractions?",
            "silent_value_cue": "Press to click my backpack, then say to me GOPOD Yourself. Your move.",
            "value_points": ["Press to click my backpack, then say to me GOPOD Yourself.", "Your move."],
        }
        record = self._run_with_fake_llm(
            line,
            brobot_2_reply="celebrate... Thank you. Interview over. What have we learned about Healthy Distractions?",
            brobot_1_reply="should not be called",
        )
        self.assertNotIn("echo_detected", record)
        self.assertNotIn("echo_source", record)
        self.assertNotIn("echo_snippet", record)


if __name__ == "__main__":
    unittest.main()

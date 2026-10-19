"""Lock for run_golden_song_001.py's live-tested animation shapes (2026-09-30).

Fails if a song's "searching" step is re-sent in a loop or stops ending with
searchingGetout (the run-on searching sound, SEARCHING_ANIMATION_SONG_LOCK_001.md), if "answering" stops settling instead, or if the continuous
dual-robot hold stops settling after it (re)takes control (arm/nod silently
dropped). Change these shapes only after a fresh live test.

Run: python3 test_golden_song_animation_lock_001.py
"""

import unittest
from unittest import mock

import run_golden_song_001 as engine


class FakeMod:
    BROBOT_1_SERIAL = "b1"
    BROBOT_2_SERIAL = "b2"

    def __init__(self):
        self.calls = []

    def wirepod_web_send_form(self, path, form, timeout=10):
        self.calls.append((path, dict(form)))
        return {"status": 200, "body": "success"}


def said(mod):
    return [form["text"] for path, form in mod.calls if path == "/api-sdk/say_text"]


class AnimationLock(unittest.TestCase):
    def run_anim(self, token, hold):
        mod = FakeMod()
        with mock.patch.object(engine.time, "sleep") as sleep:
            result = engine.run_animation_only(mod, "b2", token, hold, True, manage_control=False)
        return mod, result, [c.args[0] for c in sleep.call_args_list]

    def test_searching_sent_once_then_getout(self):
        # SEARCHING_ANIMATION_SONG_LOCK_001.md: re-sending searching every 0.333s
        # queues ~3.9s clips in Wire-Pod and leaks the searching sound.
        mod, result, sleeps = self.run_anim("searching", 1.3)
        self.assertTrue(result["ok"])
        self.assertEqual(said(mod), ["{{playAnimationWI||searching}}", "{{playAnimationWI||searchingGetout}}"])
        self.assertEqual(sleeps, [1.3, engine._ANIMATION_GETOUT_SETTLE_SECONDS])

    def test_answering_settles_without_getout(self):
        mod, result, sleeps = self.run_anim("answering", 1.0)
        texts = said(mod)
        self.assertTrue(result["ok"])
        self.assertTrue(all(t == "{{playAnimationWI||answering}}" for t in texts))
        self.assertEqual(sleeps[-1], engine._ANIMATION_NO_GETOUT_SETTLE_SECONDS)

    def test_kgsuccess_is_one_shot_held(self):
        mod, result, sleeps = self.run_anim("kgSuccess", 2.5)
        self.assertEqual(said(mod), ["{{playAnimationWI||kgSuccess}}"])
        self.assertEqual(sleeps, [2.5])

    def test_continuous_hold_settles_after_assume(self):
        mod = FakeMod()
        with mock.patch.object(engine.time, "sleep") as sleep:
            result = engine.run_continuous_hold_assume(mod, True)
        self.assertTrue(result["ok"])
        self.assertEqual([p for p, _ in mod.calls], ["/api-sdk/assume_behavior_control"] * 2)
        self.assertEqual(sleep.call_args_list[-1].args[0], engine.POST_ASSUME_SAY_SETTLE_SECONDS)


class FakeRobots:
    """say() is 'silent' (instant) for the first `silent` calls, then real."""

    scaffold = {}

    def __init__(self, silent, real_seconds=1.0):
        self.silent = silent
        self.real_seconds = real_seconds
        self.calls = 0

    def timing_prefix(self):
        return ""

    def say(self, name, line, **kwargs):
        self.calls += 1
        if self.calls > self.silent:
            engine.time.monotonic_offset += self.real_seconds
        return {"spoken": True, "mode": "wirepod"}


class SilentLineHeal(unittest.TestCase):
    def run_say(self, robots, suppress=True):
        mod = FakeMod()
        mod.normalize_robot_safe = lambda text, scaffold: (text, False)
        engine.time.monotonic_offset = 0.0
        with mock.patch.object(engine.time, "sleep"), \
                mock.patch.object(engine.time, "monotonic", lambda: engine.time.monotonic_offset):
            result = engine.run_say_turn(mod, robots, "Brobot 1", "Wait!", suppress_assume_release=suppress)
        return mod, result

    def test_silent_line_is_retaken_and_said_again(self):
        robots = FakeRobots(silent=1)
        mod, result = self.run_say(robots)
        self.assertEqual(robots.calls, 2)
        self.assertEqual([p for p, _ in mod.calls],
                         ["/api-sdk/release_behavior_control", "/api-sdk/assume_behavior_control"])
        self.assertEqual(mod.calls[0][1]["serial"], "b1")
        self.assertEqual(result["silent_heal"]["retry_seconds"], 1.0)

    def test_real_line_is_said_once(self):
        robots = FakeRobots(silent=0)
        mod, result = self.run_say(robots)
        self.assertEqual(robots.calls, 1)
        self.assertEqual(mod.calls, [])
        self.assertIsNone(result["silent_heal"])

    def test_threshold_sits_between_silent_and_real_wait(self):
        # measured: silent lines 0.03-0.18 s, real "Wait!" 0.84 s
        self.assertTrue(0.18 < engine.SAY_SILENT_SECONDS < 0.84)


class ChatPageLines(unittest.TestCase):
    def test_post_chat_line_is_display_only(self):
        mod = FakeMod()
        result = engine.post_chat_line(mod, True, "Operator", "Live Bingo Test")
        self.assertTrue(result["ok"])
        path, form = mod.calls[0]
        self.assertEqual(path, "/api-sdk/say_text")
        self.assertEqual(form["text"], "")
        self.assertEqual((form["display_role"], form["display_text"]), ("Operator", "Live Bingo Test"))

    def test_post_chat_line_dry_sends_nothing(self):
        mod = FakeMod()
        engine.post_chat_line(mod, False, "Operator", "Live Bingo Test")
        self.assertEqual(mod.calls, [])


class NarratorBubbles(unittest.TestCase):
    """2026-10-09 (operator: "do several phrases in fewer bigger chat bubbles *
    use same highlighting per narration segment"): one bubble per section, each
    caption one "@start-end" piece."""

    def test_one_bubble_per_section(self):
        steps = [
            {"section": "OPEN", "note": "wake_both"},
            {"section": "A", "start_seconds": 0, "end_seconds": 1.2, "caption": "Cupped in my hands"},
            {"section": "A", "start_seconds": 1.2, "end_seconds": 3, "caption": "are a pair"},
            {"section": "B", "start_seconds": 11, "end_seconds": 14, "caption": "They snuggle"},
        ]
        bubbles = engine._caption_bubbles({"steps": steps}, "|")
        self.assertEqual(bubbles, [
            (0.0, "@0.00-1.20 Cupped in my hands|@1.20-3.00 are a pair"),
            (11.0, "@11.00-14.00 They snuggle"),
        ])


class SleepyLull(unittest.TestCase):
    """2026-10-09 (operator: "they can lull out of and back into sleep * have
    it so they're sleeping by the end of the song"): every handed move is made,
    each Brobot is asleep again before finish() returns, then both wake."""

    def test_moves_lull_then_both_wake(self):
        class Clock:
            def timing_prefix(self):
                return ""
        sleepy = engine.SleepyBrobots(FakeMod(), False, Clock(), "T")
        sleepy.sleep_both()
        made = []
        for n in range(3):
            sleepy.move("b1", f"m{n}", lambda n=n: made.append(n) or {"ok": True})
        sleepy.move("b2", "m9", lambda: made.append(9) or {"ok": True})
        sleepy.move("b2", "stir", None)
        sleepy.finish()
        self.assertEqual(sorted(made), [0, 1, 2, 9])
        self.assertIn({"ok": True, "stir_awake_seconds": 1.0}, [r["result"] for r in sleepy.results])
        self.assertEqual(set(sleepy.procs), {"b1", "b2"})  # both asleep again
        self.assertEqual(len(sleepy.wake_all()), 2)
        self.assertFalse(sleepy.asleep)


if __name__ == "__main__":
    unittest.main()

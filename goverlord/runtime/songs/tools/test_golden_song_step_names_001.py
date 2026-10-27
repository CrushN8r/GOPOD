"""Step names are labels (ALIAS-PLAY-STUDIO.md, "Step names, the master list");
step_name_problem only warns when a name doesn't read as what its knobs play
(operator 2026-10-13: "add a check that warns when a step's name doesn't match
its knobs parts")."""
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import run_golden_song_001 as golden  # noqa: E402


class StepNameTests(unittest.TestCase):
    def test_names_that_match(self):
        for step_id, plays in [
            ("host_arm_cue_0102", [("arm_cue", "")]),
            ("brat_nod_kg_success_0203", [("nod", ""), ("animation", "kgSuccess")]),
            ("brat_searching_animation_0602", [("animation", "searching")]),
            ("host_beat_angry_host_rattle_0501", [("emotion_beat", "angry"), ("rattle", "")]),
            ("brat_say_fart2_0204", [("say_turn", ""), ("sound", "fart2.wav")]),
            ("host_wav-danger_beat_frustrated_0205", [("sound", "danger-will-robinson.wav"), ("emotion_beat", "frustrated")]),
            ("host_say_img-granny_0301", [("say_turn", ""), ("picture", "legacy/gangster_granny.jpg")]),
            ("host_thinking_say_0201", [("thinking", ""), ("say_turn", "")]),
            ("reporter_gap_0101", [("pause", "")]),
            ("host_say_danger_beat_frustrated_0205", [("sound", "danger-will-robinson.wav"), ("emotion_beat", "frustrated")]),
            ("host_grunge_say_0301", [("sound", "grunge.wav"), ("say_turn", "")]),
            ("brat_djent3b_0302", [("sound", "dJent3b.wav")]),
            ("soft_wake", [("wake_sleepy", "")]),  # no _SSNN: a free label
            ("narrator_line_0201", [("pause", "")]),
            ("goverlord_system_0103", [("system_line", "")]),
            ("brobots_soft_wake_0506", [("wake_sleepy", "")]),
            ("brobots_gap_0104", [("pause", "")]),
        ]:
            self.assertEqual(golden.step_name_problem(step_id, plays), "", step_id)

    def test_names_that_warn(self):
        self.assertIn("isn't on the master list", golden.step_name_problem("host_sya_0301", [("say_turn", "")]))
        self.assertIn("its knobs play nod", golden.step_name_problem("brat_beat_angry_0602", [("nod", "")]))
        self.assertIn("reads", golden.step_name_problem("host_say_0301", [("say_turn", ""), ("rattle", "")]))


if __name__ == "__main__":
    unittest.main()


class SongWarmupSequenceTests(unittest.TestCase):
    """GOLDEN, locked (operator 2026-10-13: "PLEASE FIX GOLDEN AND FUCKING LOCK
    THE PROCESS IN"): a song's warm-up, in the chat feed's order - "LLM Warming
    Up. Please wait." (shown and said), "Please wait." while its lines play,
    then the ready line."""

    def test_song_warmup_order(self):
        import json
        import tempfile
        from unittest import mock

        calls = []

        class FakeMod:
            pass

        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "webroot").mkdir()
            (root / "apiConfig.json").write_text(json.dumps({"knowledge": {}}))
            mod = FakeMod()
            mod.GOPOD_CHIPPER_ROOT = root
            mod.WIREPOD_CHIPPER_ROOT = root
            mod.BROBOT_1_SERIAL = "x"
            mod.wirepod_web_send_form = lambda path, form, timeout=10: calls.append((path, dict(form))) or {"status": 200}
            with mock.patch.object(golden.time, "sleep"):
                golden.song_llm_warmup(mod, "Song Ready", [{"who": "Brobot 3", "text": "a line"}])
        llm = golden.load_brobots_settings()["llm"]
        wait = llm["wait_text"]
        seen = []
        for path, form in calls:
            if "warmup_text" in form:
                seen.append("text: " + form["text"])
            elif "warmup_ready" in form:
                seen.append("ready: " + form["text"])
            elif form.get("who"):
                seen.append(form["who"] + ": " + form["text"])
            elif form.get("display_role"):
                seen.append("said: " + form["display_text"])
        self.assertEqual(seen, [
            # 2026-10-16 (the golden opener template): a line each.
            "text: " + (llm["warmup_text"] + "\n" + wait if wait else llm["warmup_text"]),
            # 2026-10-14 (operator: "generate the full message displayed"): said as shown.
            "said: " + (llm["warmup_text"] + "\n" + wait if wait else llm["warmup_text"]),
            "Brobot 3: a line",
            # 2026-10-14 (operator: "part 1 standard ... each song gives only part 2");
            # 2026-10-16: run_ready_text, the song's line, run_ready_end, a line each.
            "ready: " + llm["run_ready_text"] + "\nSong Ready\n" + llm["run_ready_end"],
        ])


class ThinkingIsTheLiveProcessTests(unittest.TestCase):
    """GOLDEN, locked (operator 2026-10-13: "use locked \"thinking please wait\"
    process for scripted runs"): a song's thinking step asks Wire-Pod for the
    live chat's own process, and no song carries its own copy of it."""

    def test_no_song_builds_its_own_thinking(self):
        import json
        songs = Path(__file__).resolve().parent.parent
        for knobs in songs.glob("*/knobs.json"):
            steps = json.loads(knobs.read_text()).get("steps", [])
            for step in steps:
                for part in step.get("combo") or [step]:
                    self.assertNotEqual(part.get("note"), "thinking_line", f"{knobs.parent.name} {step.get('step_id')}")
        self.assertIn('"/api-sdk/gopod_thinking"', Path(golden.__file__).read_text())


class NewCommandsTests(unittest.TestCase):
    """2026-10-16 (operator: "extend that to include sequences, animations,
    timers, llm warm-up & in-between reporter lines")."""

    def test_new_names(self):
        for step_id, plays in [
            ("goverlord_warm_0101", [("warm_message", "")]),
            ("goverlord_warm_wait_0103", [("warm_wait", "")]),
            ("goverlord_warm_ready_0105", [("warm_ready", "")]),
            ("reporter_lines_0104", [("reporter_lines", "")]),
            ("goverlord_sting_0106", [("sting", "")]),
            ("host_timer_0201", [("timer", "")]),
            ("brobots_sequence_0301", [("sequence", "")]),
            ("host_anim-kgsuccess_0202", [("animation", "kgSuccess")]),
        ]:
            self.assertEqual(golden.step_name_problem(step_id, plays), "", step_id)

    def test_warm_songs_names_match(self):
        songs = Path(__file__).resolve().parent.parent
        for name in ("_new_song/_llm_warm", "00_llm_warming"):
            self.assertEqual(golden.step_name_warnings(songs / name), [], name)

    def test_sequence_plays_the_other_song(self):
        import json
        import tempfile
        from unittest import mock

        control = golden._load_control_song_module()
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            for name, steps, text in [
                ("a", [{"step_id": "brobots_sequence_0101", "note": "sequence", "song_id": "b"},
                       {"step_id": "host_say_0102", "note": "say_turn"}], "## STEP host_say_0102\n\n> TEXT: After.\n"),
                ("b", [{"step_id": "brat_say_0101", "note": "say_turn"},
                       {"step_id": "brobots_sequence_0102", "note": "sequence", "song_id": "a"}], "## STEP brat_say_0101\n\n> TEXT: Inside.\n"),
            ]:
                (root / name).mkdir()
                (root / name / "knobs.json").write_text(json.dumps({"song_id": name, "steps": steps}))
                (root / name / "story.md").write_text(text)
            with mock.patch.object(golden, "find_song_dir", lambda sid: root / sid):
                song = golden.load_golden_song(root / "a", control)
        # b's line in place, b's sequence back to a skipped, then a's own line.
        self.assertEqual([(s["step_id"], s["text"]) for s in song["steps"]],
                         [("brat_say_0101", "Inside."), ("host_say_0102", "After.")])

    def test_sequence_skips_an_llm_warm_song(self):
        control = golden._load_control_song_module()
        songs = Path(__file__).resolve().parent.parent
        self.assertEqual(golden._sequence_steps({"step_id": "x", "song_id": "brobots_llm_warming"}, songs / "_new_song", control), [])


class VolumeKnobTests(unittest.TestCase):
    """2026-10-17 (operator: "other story.md knobs are either \"\" for default or
    \"0-100%\""): a volume knob is a share of the master."""

    def test_volume_share(self):
        for v, want in [("", 1.0), (None, 1.0), ("default", 1.0), ("30%", 0.3), ("150%", 1.5), (0.25, 0.25), ("0%", 0.0)]:
            self.assertAlmostEqual(golden.volume_share(v), want, msg=repr(v))

    def test_wait_for_page_goes_on_when_here(self):
        import tempfile

        class Mod:
            pass
        with tempfile.TemporaryDirectory() as tmp:
            mod = Mod()
            mod.GOPOD_CHIPPER_ROOT = Path(tmp)
            golden.wait_for_page(mod, "test")  # no flag: returns at once


class NoneRequiredWarmupTests(unittest.TestCase):
    """2026-10-17 (operator: "add: fallbacks to system messages: \"No Voice
    Generation Required\" ... \"No Language Model Required\" ... a single
    in-between message = \"\" = the indication to skip")."""

    def test_nap_style_warmup(self):
        import json
        import tempfile
        from unittest import mock

        calls = []

        class FakeMod:
            pass

        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "webroot").mkdir()
            (root / "apiConfig.json").write_text(json.dumps({"knowledge": {"endpoint": "http://x", "model": "m"}}))
            mod = FakeMod()
            mod.GOPOD_CHIPPER_ROOT = root
            mod.WIREPOD_CHIPPER_ROOT = root
            mod.wirepod_web_send_form = lambda path, form, timeout=10: calls.append(dict(form)) or {"status": 200}
            with mock.patch.object(golden.time, "sleep"), \
                    mock.patch.object(golden, "_llm_loaded", side_effect=AssertionError("no model call")):
                golden.song_llm_warmup(mod, "Nap Ready", [{"who": "Brobot 3", "text": ""}], False, False)
        llm = golden.load_brobots_settings()["llm"]
        # 2026-10-25 (operator: "No llm required, so skip the bullshit"): no
        # warm-up bubble, no wait - one Goverlord line, then the song.
        self.assertFalse([f for f in calls if "warmup_text" in f or "warmup_ready" in f], calls)
        self.assertEqual([f.get("text") for f in calls], [llm["no_llm_text"] + "\nNap Ready"])
        self.assertFalse([f for f in calls if f.get("who")], "an empty in-between line was said")


class NewStoryFormatTests(unittest.TestCase):
    """2026-10-25 (operator: "New format ... Translate/compose for native wire-pod
    compatibility"): one move or one line per step, a line's name repeating; a
    move followed by the same Brobot's line plays with it."""

    def test_order_and_compose(self):
        import json
        import tempfile
        story = ("## STEP brat_say_0203\n> TEXT: It's. Wait for it!\n"
                 "## STEP brat_kg_success_0203\n"
                 "## STEP brat_say_0203\n> TEXT: Lol!\n"
                 "## STEP host_say_0301\n> TEXT: Round 1! Please wait.\n"
                 "## STEP host_rattle_0301\n")
        knobs = {"compose_moves": True, "steps": [
            {"step_id": "brat_say_0203", "note": "say_turn", "speaker": "brobot_2", "buffer_after": 0.9},
            {"step_id": "brat_kg_success_0203", "note": "animation", "speaker": "brobot_2", "animation_token": "kgSuccess", "hold_seconds": 2.5, "buffer_after": 0.9},
            {"step_id": "brat_say_0203", "note": "say_turn", "speaker": "brobot_2", "buffer_after": 1.1},
            {"step_id": "host_say_0301", "note": "say_turn", "speaker": "brobot_1", "buffer_after": 0.0},
            {"step_id": "host_rattle_0301", "note": "rattle", "speaker": "brobot_1", "buffer_after": 0.0}]}
        with tempfile.TemporaryDirectory() as tmp:
            Path(tmp, "story.md").write_text(story)
            Path(tmp, "knobs.json").write_text(json.dumps(knobs))
            control = golden._load_control_module() if hasattr(golden, "_load_control_module") else None
            if control is None:
                import importlib.util
                spec = importlib.util.spec_from_file_location("ctl", Path(__file__).with_name("run_robot_control_song_001.py"))
                control = importlib.util.module_from_spec(spec)
                spec.loader.exec_module(control)
            song = golden.load_golden_song(tmp, control)
        got = [(s["step_id"], s.get("text", ""), s.get("buffer_after")) for s in song["steps"]]
        self.assertEqual(got, [
            ("brat_say_0203", "It's. Wait for it!", 0.9),
            ("brat_kg_success_0203", "Lol!", 1.1),
            ("host_say_0301", "Round 1! Please wait.", 0.0),
            ("host_rattle_0301", "", 0.0)])

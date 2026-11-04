"""zKnobs.md, the one source of a song's knobs (2026-11-04, step 9 of
GOPOD_CLUSTERS_PLAN_002.md; operator: "zKnobs.md is the one source"). Runs on
temporary copies - never a real song's files."""
import json
import shutil
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from knobs_envelope_001 import load_knobs_envelope, save_knobs_envelope  # noqa: E402
from knobs_md_001 import knobs_from_md, md_from_knobs  # noqa: E402

SONGS = Path(__file__).resolve().parent.parent


class KnobsMdTest(unittest.TestCase):
    def test_every_song_round_trips(self):
        """Every knobs.json, written as zKnobs.md and read back, is the same knobs."""
        files = sorted(SONGS.glob("*/knobs.json")) + sorted(SONGS.glob("*/*/knobs.json"))
        self.assertTrue(files)
        for f in files:
            knobs = json.loads(f.read_text(encoding="utf-8"))
            self.assertEqual(knobs_from_md(md_from_knobs(knobs, f.parent.name)), knobs, f)

    def test_values_keep_their_kind(self):
        knobs = {"a": " (me)", "b": "60", "c": 60, "d": 0.5, "e": True, "f": None, "g": [], "h": {},
                 "i": ["one"], "j": "x | y", "k": "line\nbreak",
                 "steps": [{"step_id": "s1", "moves": ["arm", "nod"], "hold": 1.2}, {"step_id": "s2"}]}
        self.assertEqual(knobs_from_md(md_from_knobs(knobs)), knobs)

    def test_an_edit_in_zknobs_md_plays_and_a_tool_write_lands_in_it(self):
        with tempfile.TemporaryDirectory() as tmp:
            song = Path(tmp) / "song"
            song.mkdir()
            knobs = {"song_id": "t", "global_tempo": 0.1, "steps": [{"step_id": "s1", "speed": 2}]}
            (song / "knobs.json").write_text(json.dumps(knobs, indent=2) + "\n", encoding="utf-8")
            (song / "zKnobs.md").write_text(md_from_knobs(knobs, "song"), encoding="utf-8")
            # The operator edits zKnobs.md; the next load plays the edit.
            md = (song / "zKnobs.md").read_text(encoding="utf-8").replace("speed: 2", "speed: 4")
            (song / "zKnobs.md").write_text(md, encoding="utf-8")
            path, loaded = load_knobs_envelope(song)
            self.assertEqual(loaded["steps"][0]["speed"], 4)
            self.assertEqual(json.loads(path.read_text(encoding="utf-8"))["steps"][0]["speed"], 4)
            # A tool (tempo-set) writes back: zKnobs.md holds it too.
            loaded["global_tempo"] = 0.3
            save_knobs_envelope(path, loaded)
            self.assertEqual(knobs_from_md((song / "zKnobs.md").read_text(encoding="utf-8"))["global_tempo"], 0.3)
            # An unchanged zKnobs.md leaves knobs.json as it is.
            before = (song / "knobs.json").stat().st_mtime_ns
            load_knobs_envelope(song)
            self.assertEqual((song / "knobs.json").stat().st_mtime_ns, before)

    def test_without_zknobs_md_nothing_changes(self):
        with tempfile.TemporaryDirectory() as tmp:
            song = Path(tmp) / "song"
            shutil.copytree(SONGS / "00_llm_warming", song, ignore=shutil.ignore_patterns("zKnobs.md", "runs", "notation"))
            path, loaded = load_knobs_envelope(song)
            self.assertEqual(path.name, "knobs.json")
            self.assertFalse((song / "zKnobs.md").exists())


if __name__ == "__main__":
    unittest.main()

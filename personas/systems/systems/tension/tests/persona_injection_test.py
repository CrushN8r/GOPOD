import tempfile
import unittest
from pathlib import Path

from orchestration.engine import run_pipeline


class PersonaInjectionTest(unittest.TestCase):
    def test_persona_constraints_override_operation(self) -> None:
        event = {
            "event_id": "event-override-1",
            "timestamp": "2026-04-18T12:00:00Z",
            "actors": [{"actor_id": "doc-1", "role": "doc"}],
            "expected_roles": [{"actor_id": "doc-1", "expected_role": "doc"}],
            "statements": [
                {
                    "actor_id": "doc-1",
                    "topic": "door_state",
                    "value": "open",
                    "state_key": "door_state",
                },
                {
                    "actor_id": "doc-1",
                    "topic": "door_state",
                    "value": "closed",
                    "state_key": "door_state",
                },
            ],
            "known_state": [{"key": "door_state", "value": "closed"}],
            "conditions": [],
        }

        with tempfile.TemporaryDirectory() as temp_dir:
            doc_notes_dir = Path(temp_dir) / "persona" / "notes" / "doc"
            doc_notes_dir.mkdir(parents=True)
            (doc_notes_dir / "notes.md").write_text("# Persona Notes\n", encoding="utf-8")
            (doc_notes_dir / "constraints.yaml").write_text(
                "\n".join(
                    [
                        "constraints:",
                        "  preferred_operations:",
                        "    contradiction: escalate",
                    ]
                ),
                encoding="utf-8",
            )

            result = run_pipeline(event, base_path=temp_dir)

        outputs = [
            output
            for output in result["outputs"]
            if output.persona_id == "doc" and output.field_id == "field-1"
        ]

        self.assertEqual(len(outputs), 1)
        self.assertEqual(outputs[0].operation, "escalate")
        self.assertEqual(outputs[0].payload["original_operation"], "question")
        self.assertEqual(outputs[0].payload["final_operation"], "escalate")
        self.assertTrue(outputs[0].payload["notes_applied"])


if __name__ == "__main__":
    unittest.main()

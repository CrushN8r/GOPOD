import tempfile
import unittest
from pathlib import Path

from orchestration.engine import assign_personas
from orchestration.engine import generate_outputs
from orchestration.engine import map_fields
from orchestration.engine import run_pipeline
from orchestration.engine import RuntimeContext
from persona_loader import load_persona_notes


class ConstraintsExpansionTest(unittest.TestCase):
    def _write_constraints(self, base_path: str, persona_id: str, content: str) -> None:
        notes_dir = Path(base_path) / "persona" / "notes" / persona_id
        notes_dir.mkdir(parents=True, exist_ok=True)
        (notes_dir / "notes.md").write_text("# Persona Notes\n", encoding="utf-8")
        (notes_dir / "constraints.yaml").write_text(content, encoding="utf-8")

    def test_operation_rules_override_preferred_operations(self) -> None:
        event = {
            "event_id": "event-1",
            "timestamp": "2026-04-18T12:00:00Z",
            "actors": [{"actor_id": "doc-1", "role": "doc"}],
            "expected_roles": [{"actor_id": "doc-1", "expected_role": "doc"}],
            "statements": [
                {"actor_id": "doc-1", "topic": "door_state", "value": "open", "state_key": "door_state"},
                {"actor_id": "doc-1", "topic": "door_state", "value": "closed", "state_key": "door_state"},
            ],
            "known_state": [{"key": "door_state", "value": "closed"}],
            "conditions": [],
        }

        with tempfile.TemporaryDirectory() as temp_dir:
            self._write_constraints(
                temp_dir,
                "doc",
                "\n".join(
                    [
                        "constraints:",
                        "  preferred_operations:",
                        "    contradiction: observe",
                        "  operation_rules:",
                        "    - rule_id: contradiction-doc",
                        "      signal_type: contradiction",
                        "      actor_id: '*'",
                        "      override: escalate",
                    ]
                ),
            )
            result = run_pipeline(event, base_path=temp_dir)

        output = next(
            item for item in result["outputs"] if item.persona_id == "doc" and item.field_id == "field-1"
        )
        self.assertEqual(output.operation, "escalate")
        self.assertEqual(output.payload["constraint_source"], "rule")
        self.assertEqual(output.payload["matched_rule"], "contradiction-doc")
        self.assertEqual(output.payload["original_operation"], "question")
        self.assertEqual(output.payload["final_operation"], "escalate")
        self.assertTrue(output.payload["notes_applied"])

    def test_field_filters_remove_field_for_one_persona_only(self) -> None:
        event = {
            "event_id": "event-2",
            "timestamp": "2026-04-18T12:00:00Z",
            "actors": [{"actor_id": "doc-1", "role": "doc"}],
            "expected_roles": [{"actor_id": "doc-1", "expected_role": "doc"}],
            "statements": [
                {"actor_id": "doc-1", "topic": "door_state", "value": "open", "state_key": "door_state"},
                {"actor_id": "doc-1", "topic": "door_state", "value": "closed", "state_key": "door_state"},
            ],
            "known_state": [{"key": "door_state", "value": "closed"}],
            "conditions": [],
        }

        with tempfile.TemporaryDirectory() as temp_dir:
            self._write_constraints(
                temp_dir,
                "doc",
                "\n".join(
                    [
                        "constraints:",
                        "  field_filters:",
                        "    exclude_signal_types:",
                        "      - contradiction",
                    ]
                ),
            )
            result = run_pipeline(event, base_path=temp_dir)

        doc_outputs = [item for item in result["outputs"] if item.persona_id == "doc"]
        pip_outputs = [item for item in result["outputs"] if item.persona_id == "pip"]
        self.assertFalse(any(item.field_id == "field-1" for item in doc_outputs))
        self.assertTrue(any(item.field_id == "field-1" for item in pip_outputs))

    def test_priority_rules_change_ordering_but_not_membership(self) -> None:
        fields = [
            type("FieldLike", (), {"field_id": "field-1", "field_type": "narrative_lag", "actor_id": "doc-1", "data": {"signal_type": "narrative_lag", "actor_id": "doc-1"}})(),
            type("FieldLike", (), {"field_id": "field-2", "field_type": "contradiction", "actor_id": "doc-1", "data": {"signal_type": "contradiction", "actor_id": "doc-1"}})(),
        ]
        default_assignments = assign_personas(fields)

        context = RuntimeContext(
            persona_notes={
                "doc": {
                    "notes": "",
                    "constraints": {
                        "constraints": {
                            "priority_rules": {
                                "contradiction": 10,
                                "narrative_lag": 1,
                            }
                        }
                    },
                }
            }
        )
        prioritized_assignments = assign_personas(fields, context)

        default_doc_fields = [item.field_id for item in default_assignments if item.persona_id == "doc"]
        prioritized_doc_fields = [item.field_id for item in prioritized_assignments if item.persona_id == "doc"]

        self.assertEqual(set(default_doc_fields), set(prioritized_doc_fields))
        self.assertEqual(default_doc_fields, ["field-1", "field-2"])
        self.assertEqual(prioritized_doc_fields, ["field-2", "field-1"])

    def test_no_constraints_keeps_existing_behavior(self) -> None:
        event = {
            "event_id": "event-3",
            "timestamp": "2026-04-18T12:00:00Z",
            "actors": [{"actor_id": "doc-1", "role": "doc"}],
            "expected_roles": [{"actor_id": "doc-1", "expected_role": "doc"}],
            "statements": [
                {"actor_id": "doc-1", "topic": "door_state", "value": "open", "state_key": "door_state"},
                {"actor_id": "doc-1", "topic": "door_state", "value": "closed", "state_key": "door_state"},
            ],
            "known_state": [{"key": "door_state", "value": "closed"}],
            "conditions": [],
        }

        with tempfile.TemporaryDirectory() as temp_dir:
            result = run_pipeline(event, base_path=temp_dir)

        output = next(
            item for item in result["outputs"] if item.persona_id == "doc" and item.field_id == "field-1"
        )
        self.assertEqual(output.operation, "question")
        self.assertEqual(output.payload["original_operation"], "question")
        self.assertEqual(output.payload["final_operation"], "question")
        self.assertFalse(output.payload["notes_applied"])
        self.assertEqual(output.payload["constraint_source"], "none")
        self.assertIsNone(output.payload["matched_rule"])


if __name__ == "__main__":
    unittest.main()

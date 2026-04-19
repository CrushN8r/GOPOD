import tempfile
import unittest
from pathlib import Path

from orchestration.engine import run_pipeline


class InteractionRulesTest(unittest.TestCase):
    def _write_constraints(self, base_path: str, persona_id: str, content: str) -> None:
        notes_dir = Path(base_path) / "persona" / "notes" / persona_id
        notes_dir.mkdir(parents=True, exist_ok=True)
        (notes_dir / "notes.md").write_text("# Persona Notes\n", encoding="utf-8")
        (notes_dir / "constraints.yaml").write_text(content, encoding="utf-8")

    def test_contradiction_and_narrative_lag_trigger_interaction_override(self) -> None:
        event = {
            "event_id": "event-i1",
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
                        "  interaction_rules:",
                        "    - id: rule_1",
                        "      match:",
                        "        signal_types: [contradiction, narrative_lag]",
                        "        same_actor: true",
                        "      action:",
                        "        operation: escalate",
                    ]
                ),
            )
            result = run_pipeline(event, base_path=temp_dir)

        outputs = [item for item in result["outputs"] if item.persona_id == "doc"]
        self.assertTrue(outputs)
        for output in outputs:
            self.assertEqual(output.operation, "escalate")
            self.assertTrue(output.payload["interaction_applied"])
            self.assertEqual(output.payload["interaction_rule_id"], "rule_1")
            self.assertEqual(output.payload["constraint_source"], "interaction")

    def test_rule_does_not_trigger_when_signals_incomplete(self) -> None:
        event = {
            "event_id": "event-i2",
            "timestamp": "2026-04-18T12:00:00Z",
            "actors": [{"actor_id": "doc-1", "role": "doc"}],
            "expected_roles": [{"actor_id": "doc-1", "expected_role": "doc"}],
            "statements": [
                {"actor_id": "doc-1", "topic": "door_state", "value": "open"},
                {"actor_id": "doc-1", "topic": "door_state", "value": "closed"},
            ],
            "known_state": [],
            "conditions": [],
        }

        with tempfile.TemporaryDirectory() as temp_dir:
            self._write_constraints(
                temp_dir,
                "doc",
                "\n".join(
                    [
                        "constraints:",
                        "  interaction_rules:",
                        "    - id: rule_1",
                        "      match:",
                        "        signal_types: [contradiction, narrative_lag]",
                        "        same_actor: true",
                        "      action:",
                        "        operation: escalate",
                    ]
                ),
            )
            result = run_pipeline(event, base_path=temp_dir)

        output = next(item for item in result["outputs"] if item.persona_id == "doc")
        self.assertEqual(output.operation, "question")
        self.assertFalse(output.payload["interaction_applied"])
        self.assertIsNone(output.payload["interaction_rule_id"])

    def test_multiple_groups_handled_independently(self) -> None:
        event = {
            "event_id": "event-i3",
            "timestamp": "2026-04-18T12:00:00Z",
            "actors": [
                {"actor_id": "doc-1", "role": "doc"},
                {"actor_id": "pip-1", "role": "pip"},
            ],
            "expected_roles": [
                {"actor_id": "doc-1", "expected_role": "doc"},
                {"actor_id": "pip-1", "expected_role": "pip"},
            ],
            "statements": [
                {"actor_id": "doc-1", "topic": "door_state", "value": "open", "state_key": "door_state"},
                {"actor_id": "doc-1", "topic": "door_state", "value": "closed", "state_key": "door_state"},
                {"actor_id": "pip-1", "topic": "gate_state", "value": "up"},
                {"actor_id": "pip-1", "topic": "gate_state", "value": "down"},
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
                        "  interaction_rules:",
                        "    - id: rule_1",
                        "      match:",
                        "        signal_types: [contradiction, narrative_lag]",
                        "        same_actor: true",
                        "      action:",
                        "        operation: escalate",
                    ]
                ),
            )
            result = run_pipeline(event, base_path=temp_dir)

        doc_outputs = [item for item in result["outputs"] if item.persona_id == "doc"]
        matching_group_outputs = [
            item for item in doc_outputs if item.payload["actor_id"] == "doc-1"
        ]
        non_matching_group_outputs = [
            item for item in doc_outputs if item.payload["actor_id"] == "pip-1"
        ]

        self.assertTrue(matching_group_outputs)
        self.assertTrue(all(item.payload["interaction_applied"] for item in matching_group_outputs))
        self.assertTrue(non_matching_group_outputs)
        self.assertTrue(all(not item.payload["interaction_applied"] for item in non_matching_group_outputs))

    def test_interaction_overrides_lower_layers(self) -> None:
        event = {
            "event_id": "event-i4",
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
                        "      override: constrain",
                        "  interaction_rules:",
                        "    - id: rule_1",
                        "      match:",
                        "        signal_types: [contradiction, narrative_lag]",
                        "        same_actor: true",
                        "      action:",
                        "        operation: escalate",
                    ]
                ),
            )
            result = run_pipeline(event, base_path=temp_dir)

        output = next(item for item in result["outputs"] if item.persona_id == "doc")
        self.assertEqual(output.operation, "escalate")
        self.assertEqual(output.payload["constraint_source"], "interaction")
        self.assertEqual(output.payload["interaction_rule_id"], "rule_1")

    def test_no_interaction_rules_keeps_existing_behavior(self) -> None:
        event = {
            "event_id": "event-i5",
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

        output = next(item for item in result["outputs"] if item.persona_id == "doc")
        self.assertEqual(output.operation, "question")
        self.assertFalse(output.payload["interaction_applied"])
        self.assertIsNone(output.payload["interaction_rule_id"])
        self.assertEqual(output.payload["constraint_source"], "none")


if __name__ == "__main__":
    unittest.main()

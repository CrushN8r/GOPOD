import tempfile
import unittest
from pathlib import Path

from context import ContextStore
from orchestration.engine import run_pipeline


class ContextRulesTest(unittest.TestCase):
    def _write_constraints(self, base_path: str, persona_id: str, content: str) -> None:
        notes_dir = Path(base_path) / "persona" / "notes" / persona_id
        notes_dir.mkdir(parents=True, exist_ok=True)
        (notes_dir / "notes.md").write_text("# Persona Notes\n", encoding="utf-8")
        (notes_dir / "constraints.yaml").write_text(content, encoding="utf-8")

    def _contradiction_event(self, actor_id: str = "doc-1") -> dict:
        return {
            "event_id": f"event-{actor_id}",
            "timestamp": "2026-04-18T12:00:00Z",
            "actors": [{"actor_id": actor_id, "role": "doc"}],
            "expected_roles": [{"actor_id": actor_id, "expected_role": "doc"}],
            "statements": [
                {"actor_id": actor_id, "topic": "door_state", "value": "open", "state_key": "door_state"},
                {"actor_id": actor_id, "topic": "door_state", "value": "closed", "state_key": "door_state"},
            ],
            "known_state": [{"key": "door_state", "value": "closed"}],
            "conditions": [],
        }

    def test_repeated_contradiction_triggers_override(self) -> None:
        store = ContextStore()
        constraints = "\n".join(
            [
                "constraints:",
                "  context_rules:",
                "    - id: repeat_contradiction",
                "      match:",
                "        signal_type: contradiction",
                "        same_actor: true",
                "        repeated: true",
                "      action:",
                "        operation: escalate",
            ]
        )

        with tempfile.TemporaryDirectory() as temp_dir:
            self._write_constraints(temp_dir, "doc", constraints)
            run_pipeline(self._contradiction_event(), base_path=temp_dir, context_store=store, tick_id=1)
            result = run_pipeline(self._contradiction_event(), base_path=temp_dir, context_store=store, tick_id=2)

        output = next(item for item in result["outputs"] if item.persona_id == "doc" and item.field_id == "field-1")
        self.assertEqual(output.operation, "escalate")
        self.assertTrue(output.payload["context_applied"])
        self.assertEqual(output.payload["context_rule_id"], "repeat_contradiction")
        self.assertTrue(output.payload["repeated_signal"])
        self.assertEqual(output.payload["constraint_source"], "context")

    def test_no_previous_snapshot_has_no_effect(self) -> None:
        store = ContextStore()
        constraints = "\n".join(
            [
                "constraints:",
                "  context_rules:",
                "    - id: repeat_contradiction",
                "      match:",
                "        signal_type: contradiction",
                "        same_actor: true",
                "        repeated: true",
                "      action:",
                "        operation: escalate",
            ]
        )

        with tempfile.TemporaryDirectory() as temp_dir:
            self._write_constraints(temp_dir, "doc", constraints)
            result = run_pipeline(self._contradiction_event(), base_path=temp_dir, context_store=store, tick_id=1)

        output = next(item for item in result["outputs"] if item.persona_id == "doc")
        self.assertEqual(output.operation, "question")
        self.assertFalse(output.payload["context_applied"])
        self.assertFalse(output.payload["repeated_signal"])

    def test_different_actor_does_not_match(self) -> None:
        store = ContextStore()
        constraints = "\n".join(
            [
                "constraints:",
                "  context_rules:",
                "    - id: repeat_contradiction",
                "      match:",
                "        signal_type: contradiction",
                "        same_actor: true",
                "        repeated: true",
                "      action:",
                "        operation: escalate",
            ]
        )

        with tempfile.TemporaryDirectory() as temp_dir:
            self._write_constraints(temp_dir, "doc", constraints)
            run_pipeline(self._contradiction_event("doc-1"), base_path=temp_dir, context_store=store, tick_id=1)
            result = run_pipeline(self._contradiction_event("doc-2"), base_path=temp_dir, context_store=store, tick_id=2)

        output = next(item for item in result["outputs"] if item.persona_id == "doc")
        self.assertEqual(output.operation, "question")
        self.assertFalse(output.payload["context_applied"])
        self.assertFalse(output.payload["repeated_signal"])

    def test_context_precedence_over_interaction_rules(self) -> None:
        store = ContextStore()
        constraints = "\n".join(
            [
                "constraints:",
                "  interaction_rules:",
                "    - id: rule_1",
                "      match:",
                "        signal_types: [contradiction, narrative_lag]",
                "        same_actor: true",
                "      action:",
                "        operation: constrain",
                "  context_rules:",
                "    - id: repeat_contradiction",
                "      match:",
                "        signal_type: contradiction",
                "        same_actor: true",
                "        repeated: true",
                "      action:",
                "        operation: escalate",
            ]
        )

        with tempfile.TemporaryDirectory() as temp_dir:
            self._write_constraints(temp_dir, "doc", constraints)
            run_pipeline(self._contradiction_event(), base_path=temp_dir, context_store=store, tick_id=1)
            result = run_pipeline(self._contradiction_event(), base_path=temp_dir, context_store=store, tick_id=2)

        contradiction_output = next(
            item for item in result["outputs"] if item.persona_id == "doc" and item.field_id == "field-1"
        )
        self.assertEqual(contradiction_output.operation, "escalate")
        self.assertEqual(contradiction_output.payload["constraint_source"], "context")

    def test_deterministic_across_runs(self) -> None:
        constraints = "\n".join(
            [
                "constraints:",
                "  context_rules:",
                "    - id: repeat_contradiction",
                "      match:",
                "        signal_type: contradiction",
                "        same_actor: true",
                "        repeated: true",
                "      action:",
                "        operation: escalate",
            ]
        )

        with tempfile.TemporaryDirectory() as temp_dir:
            self._write_constraints(temp_dir, "doc", constraints)

            first_store = ContextStore()
            run_pipeline(self._contradiction_event(), base_path=temp_dir, context_store=first_store, tick_id=1)
            first_result = run_pipeline(
                self._contradiction_event(), base_path=temp_dir, context_store=first_store, tick_id=2
            )

            second_store = ContextStore()
            run_pipeline(self._contradiction_event(), base_path=temp_dir, context_store=second_store, tick_id=1)
            second_result = run_pipeline(
                self._contradiction_event(), base_path=temp_dir, context_store=second_store, tick_id=2
            )

        first_output = next(item for item in first_result["outputs"] if item.persona_id == "doc")
        second_output = next(item for item in second_result["outputs"] if item.persona_id == "doc")
        self.assertEqual(first_output.operation, second_output.operation)
        self.assertEqual(first_output.payload, second_output.payload)


if __name__ == "__main__":
    unittest.main()

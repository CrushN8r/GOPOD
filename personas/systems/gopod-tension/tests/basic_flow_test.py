import unittest

from orchestration.engine import generate_outputs
from orchestration.engine import assign_personas
from orchestration.engine import extract_signals
from orchestration.engine import ingest_event
from orchestration.engine import map_fields


class BasicFlowTest(unittest.TestCase):
    def test_conflicting_statements_flow(self) -> None:
        event = {
            "event_id": "event-1",
            "timestamp": "2026-04-18T12:00:00Z",
            "actors": [
                {"actor_id": "doc-1", "role": "doc"},
                {"actor_id": "student-1", "role": "guards"},
                {"actor_id": "guard-1", "role": "guards"},
            ],
            "expected_roles": [
                {"actor_id": "doc-1", "expected_role": "doc"},
                {"actor_id": "student-1", "expected_role": "students"},
                {"actor_id": "guard-1", "expected_role": "guards"},
            ],
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
            "conditions": [{"name": "checkpoint", "status": "changing"}],
        }

        ingested = ingest_event(event)
        signals = extract_signals(ingested)
        fields = map_fields(signals)
        assignments = assign_personas(fields)
        outputs = generate_outputs(assignments)

        signal_types = {signal.signal_type for signal in signals}
        persona_ids = {assignment.persona_id for assignment in assignments}
        output_operations = {output.operation for output in outputs}

        self.assertIn("contradiction", signal_types)
        self.assertIn("narrative_lag", signal_types)
        self.assertIn("identity_collision", signal_types)
        self.assertIn("pressure", signal_types)
        self.assertGreater(len(fields), 0)
        self.assertIn("doc", persona_ids)
        self.assertIn("pip", persona_ids)
        self.assertIn("students", persona_ids)
        self.assertIn("guards", persona_ids)
        self.assertGreater(len(outputs), 0)
        self.assertIn("question", output_operations)
        self.assertIn("observe", output_operations)
        self.assertIn("constrain", output_operations)


if __name__ == "__main__":
    unittest.main()

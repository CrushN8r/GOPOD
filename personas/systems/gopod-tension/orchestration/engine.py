from __future__ import annotations

from dataclasses import asdict, dataclass
from typing import Any


ALLOWED_ROLES = {"doc", "pip", "students", "guards"}
ALLOWED_OPERATIONS = {"observe", "question", "escalate", "constrain"}
PERSONA_FIELD_MAP = {
    "doc": {"contradiction", "narrative_lag"},
    "pip": {"contradiction", "narrative_lag"},
    "students": {"identity_collision"},
    "guards": {"pressure"},
}
FIELD_OPERATION_MAP = {
    "contradiction": "question",
    "identity_collision": "observe",
    "narrative_lag": "question",
    "pressure": "constrain",
}


@dataclass(frozen=True)
class Signal:
    signal_id: str
    signal_type: str
    actor_id: str
    evidence: dict[str, Any]
    topic: str | None = None


@dataclass(frozen=True)
class Field:
    field_id: str
    field_type: str
    source_signal_id: str
    actor_id: str
    data: dict[str, Any]


@dataclass(frozen=True)
class PersonaAssignment:
    assignment_id: str
    persona_id: str
    field_id: str
    field_type: str
    operation: str
    payload: dict[str, Any]


@dataclass(frozen=True)
class PersonaOutput:
    persona_id: str
    field_id: str
    operation: str
    payload: dict[str, Any]


def ingest_event(input_event: dict[str, Any]) -> dict[str, Any]:
    required_fields = [
        "event_id",
        "timestamp",
        "actors",
        "expected_roles",
        "statements",
        "known_state",
        "conditions",
    ]
    for field_name in required_fields:
        if field_name not in input_event:
            raise ValueError(f"missing required field: {field_name}")
    return input_event


def extract_signals(event: dict[str, Any]) -> list[Signal]:
    signals: list[Signal] = []
    signal_index = 1

    actor_roles = {actor["actor_id"]: actor["role"] for actor in event["actors"]}
    expected_roles = {
        item["actor_id"]: item["expected_role"] for item in event["expected_roles"]
    }
    known_state = {item["key"]: item["value"] for item in event["known_state"]}

    statements_by_actor_topic: dict[tuple[str, str], list[dict[str, Any]]] = {}
    for statement in event["statements"]:
        key = (statement["actor_id"], statement["topic"])
        statements_by_actor_topic.setdefault(key, []).append(statement)

    for (actor_id, topic), statements in sorted(statements_by_actor_topic.items()):
        values = {statement["value"] for statement in statements}
        if len(values) > 1:
            signals.append(
                Signal(
                    signal_id=f"signal-{signal_index}",
                    signal_type="contradiction",
                    actor_id=actor_id,
                    topic=topic,
                    evidence={"statements": statements},
                )
            )
            signal_index += 1

    for actor_id in sorted(actor_roles):
        actual_role = actor_roles[actor_id]
        expected_role = expected_roles.get(actor_id)
        if expected_role and actual_role != expected_role:
            signals.append(
                Signal(
                    signal_id=f"signal-{signal_index}",
                    signal_type="identity_collision",
                    actor_id=actor_id,
                    evidence={
                        "actual_role": actual_role,
                        "expected_role": expected_role,
                    },
                )
            )
            signal_index += 1

    for statement in event["statements"]:
        state_key = statement.get("state_key")
        if not state_key:
            continue
        current_value = known_state.get(state_key)
        if current_value is not None and current_value != statement["value"]:
            signals.append(
                Signal(
                    signal_id=f"signal-{signal_index}",
                    signal_type="narrative_lag",
                    actor_id=statement["actor_id"],
                    topic=statement["topic"],
                    evidence={
                        "statement": statement,
                        "known_state": {"key": state_key, "value": current_value},
                    },
                )
            )
            signal_index += 1

    for condition in event["conditions"]:
        if condition["status"] not in {"unstable", "changing"}:
            continue
        signals.append(
            Signal(
                signal_id=f"signal-{signal_index}",
                signal_type="pressure",
                actor_id="environment",
                evidence={"condition": condition},
            )
        )
        signal_index += 1

    return signals


def map_fields(signals: list[Signal]) -> list[Field]:
    fields: list[Field] = []
    for index, signal in enumerate(signals, start=1):
        fields.append(
            Field(
                field_id=f"field-{index}",
                field_type=signal.signal_type,
                source_signal_id=signal.signal_id,
                actor_id=signal.actor_id,
                data=asdict(signal),
            )
        )
    return fields


def assign_personas(fields: list[Field]) -> list[PersonaAssignment]:
    assignments: list[PersonaAssignment] = []
    assignment_index = 1

    for field in fields:
        for persona_id in sorted(PERSONA_FIELD_MAP):
            if field.field_type not in PERSONA_FIELD_MAP[persona_id]:
                continue
            operation = FIELD_OPERATION_MAP[field.field_type]
            if operation not in ALLOWED_OPERATIONS:
                raise ValueError(f"invalid operation: {operation}")
            assignments.append(
                PersonaAssignment(
                    assignment_id=f"assignment-{assignment_index}",
                    persona_id=persona_id,
                    field_id=field.field_id,
                    field_type=field.field_type,
                    operation=operation,
                    payload=field.data,
                )
            )
            assignment_index += 1

    return assignments


def generate_outputs(assignments: list[PersonaAssignment]) -> list[PersonaOutput]:
    outputs: list[PersonaOutput] = []
    for assignment in assignments:
        if assignment.persona_id not in ALLOWED_ROLES:
            raise ValueError(f"invalid persona: {assignment.persona_id}")
        outputs.append(
            PersonaOutput(
                persona_id=assignment.persona_id,
                field_id=assignment.field_id,
                operation=assignment.operation,
                payload=assignment.payload,
            )
        )
    return outputs


def run_pipeline(input_event: dict[str, Any]) -> dict[str, Any]:
    event = ingest_event(input_event)
    signals = extract_signals(event)
    fields = map_fields(signals)
    assignments = assign_personas(fields)
    outputs = generate_outputs(assignments)
    return {
        "event": event,
        "signals": signals,
        "fields": fields,
        "assignments": assignments,
        "outputs": outputs,
    }


if __name__ == "__main__":
    sample_event = {
        "event_id": "event-1",
        "timestamp": "2026-04-18T12:00:00Z",
        "actors": [
            {"actor_id": "a1", "role": "doc"},
            {"actor_id": "a2", "role": "guards"},
        ],
        "expected_roles": [
            {"actor_id": "a1", "expected_role": "doc"},
            {"actor_id": "a2", "expected_role": "guards"},
        ],
        "statements": [
            {
                "actor_id": "a1",
                "topic": "gate_state",
                "value": "open",
                "state_key": "gate_state",
            },
            {
                "actor_id": "a1",
                "topic": "gate_state",
                "value": "closed",
                "state_key": "gate_state",
            },
        ],
        "known_state": [{"key": "gate_state", "value": "closed"}],
        "conditions": [{"name": "perimeter", "status": "changing"}],
    }
    result = run_pipeline(sample_event)
    print(asdict(result["outputs"][0]))

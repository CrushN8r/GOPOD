from __future__ import annotations

from dataclasses import asdict
from dataclasses import dataclass
from dataclasses import is_dataclass
from typing import Any

from context.schema import TickSnapshot
from extraction.schema import TensionSignal
from input.schema import InputEvent
from mapping.schema import FieldMap
from orchestration.engine import PIPELINE_ORDER
from orchestration.engine import build_runtime_context
from orchestration.engine import build_snapshot
from orchestration.engine import extract_signals
from orchestration.engine import generate_outputs
from orchestration.engine import ingest_event
from orchestration.engine import map_fields
from orchestration.engine import route_fields
from persona.schema import PersonaAssignment


@dataclass(frozen=True)
class HarnessResult:
    status: str
    stage_trace: tuple[dict[str, Any], ...]
    determinism: bool


def _validate_input_event(value: Any) -> bool:
    return isinstance(value, InputEvent)


def _validate_tension_signals(value: Any) -> bool:
    return isinstance(value, tuple) and all(isinstance(item, TensionSignal) for item in value)


def _validate_field_maps(value: Any) -> bool:
    return isinstance(value, tuple) and all(isinstance(item, FieldMap) for item in value)


def _validate_persona_assignments(value: Any) -> bool:
    return isinstance(value, tuple) and all(isinstance(item, PersonaAssignment) for item in value)


def _validate_tick_snapshot(value: Any) -> bool:
    return isinstance(value, TickSnapshot)


def _build_stage_record(stage: str, valid: bool, payload: Any) -> dict[str, Any]:
    if is_dataclass(payload):
        payload_type = type(payload).__name__
    else:
        payload_type = type(payload).__name__
    return {
        "stage": stage,
        "valid": valid,
        "payload_type": payload_type,
    }


def run_harness(
    input_event: InputEvent,
    base_path: str | None = None,
) -> HarnessResult:
    stage_trace: list[dict[str, Any]] = []
    runtime_context = build_runtime_context(base_path=base_path)

    ingested = ingest_event(input_event)
    ingested_valid = _validate_input_event(ingested)
    stage_trace.append(_build_stage_record("input", ingested_valid, ingested))

    signals = extract_signals(ingested)
    signals_valid = _validate_tension_signals(signals)
    stage_trace.append(_build_stage_record("extraction", signals_valid, signals))

    fields = map_fields(signals)
    fields_valid = _validate_field_maps(fields)
    stage_trace.append(_build_stage_record("mapping", fields_valid, fields))

    assignments = route_fields(fields, runtime_context=runtime_context)
    assignments_valid = _validate_persona_assignments(assignments)
    stage_trace.append(_build_stage_record("routing", assignments_valid, assignments))

    outputs = generate_outputs(assignments)
    outputs_valid = isinstance(outputs, tuple)
    stage_trace.append(_build_stage_record("output", outputs_valid, outputs))

    snapshot = build_snapshot(fields, tick_id=input_event.tick_id or 0)
    snapshot_valid = _validate_tick_snapshot(snapshot)
    stage_trace.append(_build_stage_record("snapshot", snapshot_valid, snapshot))

    status = "PASS" if all(item["valid"] for item in stage_trace) else "FAIL"
    determinism = tuple(item["stage"] for item in stage_trace[:5]) == PIPELINE_ORDER

    return HarnessResult(
        status=status,
        stage_trace=tuple(stage_trace),
        determinism=determinism,
    )


def result_as_dict(result: HarnessResult) -> dict[str, Any]:
    return asdict(result)

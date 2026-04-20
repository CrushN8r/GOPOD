from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Any

from context import ContextStore
from context import RuntimeContext
from context import TickFieldSnapshot
from context import TickSnapshot
from extraction.schema import TensionSignal
from input.schema import InputEvent
from mapping.schema import FieldMap
from persona.loader import PERSONA_NOTES_ROOT
from persona.loader import load_persona_registry
from persona.schema import PersonaAssignment
from robot.registry import load_robot_registry


PIPELINE_ORDER = ("input", "extraction", "mapping", "routing", "output")


@dataclass(frozen=True)
class PipelineState:
    input_event: Any
    signals: tuple[TensionSignal, ...]
    fields: tuple[FieldMap, ...]
    assignments: tuple[PersonaAssignment, ...]
    outputs: tuple[Any, ...]
    context: RuntimeContext
    snapshot: TickSnapshot


def _default_base_path() -> Path:
    return Path(__file__).resolve().parents[3]


def build_runtime_context(
    base_path: str | None = None, context_store: ContextStore | None = None
) -> RuntimeContext:
    root = Path(base_path) if base_path else _default_base_path()
    store = context_store if context_store is not None else ContextStore()
    return RuntimeContext(
        recent_snapshots=store.get_recent(),
        personas=load_persona_registry(str(root)),
        robots=load_robot_registry(str(root / "config" / "system.json")),
        notes_root=str(root / PERSONA_NOTES_ROOT),
    )


def ingest_event(input_event: InputEvent | dict[str, Any] | Any) -> InputEvent | dict[str, Any] | Any:
    return input_event


def extract_signals(event: InputEvent | dict[str, Any] | Any) -> tuple[TensionSignal, ...]:
    _ = event
    return ()


def map_fields(signals: tuple[TensionSignal, ...]) -> tuple[FieldMap, ...]:
    _ = signals
    return ()


def route_fields(
    fields: tuple[FieldMap, ...], runtime_context: RuntimeContext | None = None
) -> tuple[PersonaAssignment, ...]:
    _ = fields
    _ = runtime_context
    return ()


def generate_outputs(assignments: tuple[PersonaAssignment, ...]) -> tuple[Any, ...]:
    _ = assignments
    return ()


def build_snapshot(fields: tuple[FieldMap, ...], tick_id: int = 0) -> TickSnapshot:
    return TickSnapshot(
        tick_id=tick_id,
        fields=tuple(
            TickFieldSnapshot(
                field_id=field.field_id,
                signal_type=field.signal_type,
                actor_id=field.actor_id,
                topic=field.topic,
            )
            for field in fields
        ),
    )


def run_pipeline(
    input_event: InputEvent | dict[str, Any] | Any,
    base_path: str | None = None,
    context_store: ContextStore | None = None,
    tick_id: int = 0,
) -> PipelineState:
    store = context_store if context_store is not None else ContextStore()
    runtime_context = build_runtime_context(base_path=base_path, context_store=store)
    event = ingest_event(input_event)
    signals = extract_signals(event)
    fields = map_fields(signals)
    assignments = route_fields(fields, runtime_context=runtime_context)
    outputs = generate_outputs(assignments)
    snapshot = build_snapshot(fields, tick_id=tick_id)
    store.add_snapshot(snapshot)
    return PipelineState(
        input_event=event,
        signals=signals,
        fields=fields,
        assignments=assignments,
        outputs=outputs,
        context=runtime_context,
        snapshot=snapshot,
    )

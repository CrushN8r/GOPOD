from __future__ import annotations

from dataclasses import dataclass
from dataclasses import field

from persona.schema import PersonaDefinition
from robot.registry import Robot


@dataclass(frozen=True)
class TickFieldSnapshot:
    field_id: str | None = None
    signal_type: str | None = None
    actor_id: str | None = None
    topic: str | None = None


@dataclass(frozen=True)
class TickSnapshot:
    tick_id: int
    fields: tuple[TickFieldSnapshot, ...] = ()


@dataclass(frozen=True)
class RuntimeContext:
    recent_snapshots: tuple[TickSnapshot, ...] = ()
    personas: tuple[PersonaDefinition, ...] = ()
    robots: tuple[Robot, ...] = ()
    notes_root: str | None = None

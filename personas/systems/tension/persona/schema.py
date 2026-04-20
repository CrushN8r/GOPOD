from __future__ import annotations

from dataclasses import dataclass
from dataclasses import field
from typing import Any


@dataclass(frozen=True)
class PersonaDefinition:
    persona_id: str
    notes_path: str | None = None
    constraints_path: str | None = None
    metadata: dict[str, Any] = field(default_factory=dict)


@dataclass(frozen=True)
class PersonaAssignment:
    assignment_id: str | None = None
    persona_id: str | None = None
    field_id: str | None = None
    operation: str | None = None
    payload: dict[str, Any] = field(default_factory=dict)

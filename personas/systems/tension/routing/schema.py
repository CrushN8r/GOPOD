from __future__ import annotations

from dataclasses import dataclass
from dataclasses import field
from typing import Any


@dataclass(frozen=True)
class OperationRule:
    rule_id: str | None = None
    signal_type: str | None = None
    actor_id: str | None = None
    override: str | None = None
    metadata: dict[str, Any] = field(default_factory=dict)


@dataclass(frozen=True)
class InteractionRule:
    rule_id: str | None = None
    match: dict[str, Any] = field(default_factory=dict)
    action: dict[str, Any] = field(default_factory=dict)


@dataclass(frozen=True)
class PriorityRule:
    signal_type: str | None = None
    priority: int | None = None

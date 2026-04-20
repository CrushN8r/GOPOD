from __future__ import annotations

from dataclasses import dataclass
from dataclasses import field
from typing import Any


@dataclass(frozen=True)
class TensionSignal:
    signal_id: str | None = None
    signal_type: str | None = None
    actor_id: str | None = None
    topic: str | None = None
    payload: dict[str, Any] = field(default_factory=dict)

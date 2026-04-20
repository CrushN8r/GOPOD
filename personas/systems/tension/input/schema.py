from __future__ import annotations

from dataclasses import dataclass
from dataclasses import field
from typing import Any


@dataclass(frozen=True)
class InputEvent:
    event_id: str | None = None
    tick_id: int | None = None
    timestamp: str | None = None
    payload: dict[str, Any] = field(default_factory=dict)

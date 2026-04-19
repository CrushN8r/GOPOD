from __future__ import annotations

from context.schema import TickSnapshot


class ContextStore:
    def __init__(self, max_history: int = 1, history: tuple[TickSnapshot, ...] = ()) -> None:
        self.max_history = max_history
        self.history = list(history)

    def add_snapshot(self, snapshot: TickSnapshot) -> None:
        next_history = [*self.history, snapshot]
        if len(next_history) > self.max_history:
            next_history = next_history[-self.max_history :]
        self.history = next_history

    def get_recent(self, n: int = 1) -> tuple[TickSnapshot, ...]:
        if n <= 0:
            return ()
        return tuple(self.history[-n:])

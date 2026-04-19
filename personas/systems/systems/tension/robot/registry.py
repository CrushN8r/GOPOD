from __future__ import annotations

import json
from dataclasses import dataclass
from dataclasses import field
from pathlib import Path
from typing import Any


@dataclass(frozen=True)
class Robot:
    name: str
    id: str
    assigned_fields: tuple[str, ...] = ()
    filter_rules: tuple[dict[str, Any], ...] = ()
    priority: int | None = None
    metadata: dict[str, Any] = field(default_factory=dict)


def _default_system_config_path() -> Path:
    return Path(__file__).resolve().parents[3] / "config" / "system.json"


def load_robot_registry(system_config_path: str | None = None) -> tuple[Robot, ...]:
    config_path = Path(system_config_path) if system_config_path else _default_system_config_path()
    if not config_path.exists():
        return ()

    try:
        config_data = json.loads(config_path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return ()

    robots_block = config_data.get("robots")
    if not isinstance(robots_block, dict):
        return ()

    robots: list[Robot] = []
    for robot_id in sorted(robots_block):
        robot_data = robots_block[robot_id]
        if not isinstance(robot_data, dict):
            continue
        name = str(robot_data.get("name", robot_id))
        assigned_fields = robot_data.get("assigned_fields")
        if not isinstance(assigned_fields, list):
            assigned_fields = []
        filter_rules = robot_data.get("filter_rules")
        if not isinstance(filter_rules, list):
            filter_rules = []
        priority = robot_data.get("priority")
        if not isinstance(priority, int):
            priority = None
        robots.append(
            Robot(
                name=name,
                id=robot_id,
                assigned_fields=tuple(str(item) for item in assigned_fields),
                filter_rules=tuple(item for item in filter_rules if isinstance(item, dict)),
                priority=priority,
                metadata=robot_data,
            )
        )
    return tuple(robots)

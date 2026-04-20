#!/usr/bin/env python3
from __future__ import annotations

import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

import json
import sys
from pathlib import Path
from typing import Any

from core.router import route_code_action
from core.router import route_robot_action
from core.validator import validate_action


REPO_ROOT = Path(__file__).resolve().parents[1]
CONFIGS_DIR = REPO_ROOT / "configs"
SYSTEM_CONFIG_PATH = CONFIGS_DIR / "system.json"
SYSTEM_TEMPLATE_PATH = CONFIGS_DIR / "system.template.json"


def _load_config() -> dict[str, Any]:
    config_path = SYSTEM_CONFIG_PATH if SYSTEM_CONFIG_PATH.exists() else SYSTEM_TEMPLATE_PATH
    if not config_path.exists():
        raise RuntimeError(
            "Missing configs/system.json and configs/system.template.json."
        )
    return json.loads(config_path.read_text(encoding="utf-8"))


def goverlord_exec(action: dict[str, Any]) -> dict[str, Any]:
    """THE ONLY EXECUTION ENTRYPOINT — use this from every tool"""
    config = _load_config()
    robots = config.get("robots", {})
    normalized = validate_action(action, robots=robots)

    if normalized["type"] == "code":
        codex_exec_cmd = str(config.get("codex_exec_cmd", "bash scripts/codex-exec.sh"))
        return route_code_action(normalized, repo_root=REPO_ROOT, codex_exec_cmd=codex_exec_cmd)

    t560_base = str(config.get("t560_base", "")).strip()
    if not t560_base:
        raise RuntimeError("Missing required config value: t560_base")
    return route_robot_action(normalized, t560_base=t560_base, robot_config=robots[normalized["target"]])


def _build_legacy_action(argv: list[str]) -> dict[str, Any]:
    if len(argv) < 4:
        raise ValueError(
            "Usage: python3 core/goverlord_exec.py <json_action> or "
            "python3 core/goverlord_exec.py <type> <target> <action> [param_key param_value]"
        )
    params = {argv[3]: " ".join(argv[4:])} if len(argv) > 4 else {}
    return {
        "type": argv[0],
        "target": argv[1],
        "action": argv[2],
        "params": params,
    }


def main(argv: list[str] | None = None) -> int:
    args = list(sys.argv[1:] if argv is None else argv)
    if not args:
        print(
            "Usage: python3 core/goverlord_exec.py <json_action> or "
            "python3 core/goverlord_exec.py <type> <target> <action> [param_key param_value]",
            file=sys.stderr,
        )
        return 1

    action = json.loads(args[0]) if args[0].startswith("{") else _build_legacy_action(args)
    result = goverlord_exec(action)
    print(json.dumps(result, indent=2, sort_keys=True))
    return 0 if result.get("ok") else 1


if __name__ == "__main__":
    raise SystemExit(main())


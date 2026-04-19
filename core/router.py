from __future__ import annotations

import shlex
import subprocess
from pathlib import Path
from typing import Any
import requests


def route_code_action(
    action: dict[str, Any],
    repo_root: Path,
    codex_exec_cmd: str,
) -> dict[str, Any]:
    task = str(action["params"].get("task") or action["action"])
    target_path = str(action["params"].get("path") or "")

    cmd = shlex.split(codex_exec_cmd) + ["--task", task]
    if target_path:
        cmd.extend(["--path", target_path])

    result = subprocess.run(
        cmd,
        cwd=repo_root,
        capture_output=True,
        text=True,
        timeout=60,
    )
    return {
        "ok": result.returncode == 0,
        "route": "codex",
        "returncode": result.returncode,
        "stdout": result.stdout.strip(),
        "stderr": result.stderr.strip(),
    }


def route_robot_action(
    action: dict[str, Any],
    t560_base: str,
    robot_config: dict[str, Any],
) -> dict[str, Any]:
    prefix = str(robot_config.get("prefix", "/api-sdk")).rstrip("/")
    serial = str(action["params"].get("serial") or robot_config.get("serial") or "").strip()
    endpoint_base = f"{t560_base.rstrip('/')}{prefix}"
    # full robot logic follows the original validated pattern from PDF (T560 only)
    endpoint = f"{endpoint_base}/{action['action']}"
    try:
        if action["action"] == "say":
            payload = {"text": action["params"].get("text", "")}
            response = requests.post(f"{endpoint}_text", json=payload, timeout=5)
        else:
            response = requests.post(endpoint, json=action["params"], timeout=5)
        response.raise_for_status()
        return {"ok": True, "route": "robot", "response": response.json() if response.content else {}}
    except Exception as e:
        return {"ok": False, "route": "robot", "error": str(e)}

#!/usr/scripts/env python3
"""
goverlord_core.py — STRICT EXECUTION KERNEL (Option A + Codex unification)
Single entrypoint for the entire GOPOD system.
All execution paths now resolve here → T560 or Codex only.
"""

import json
import os
import shlex
import subprocess
import sys
from pathlib import Path
from typing import Any, Dict

import requests

# ====================== CONFIG ======================
CONFIG_PATH = Path(__file__).parent.parent / "config" / "system.json"

if not CONFIG_PATH.exists():
    raise RuntimeError(f"❌ Missing config: {CONFIG_PATH}\nCopy config/system.template.json → system.json and fill it.")

with open(CONFIG_PATH) as f:
    CONFIG = json.load(f)

T560_BASE = CONFIG.get("t560_base")
if not T560_BASE or "192.168.1.6" in T560_BASE and "TODO" in T560_BASE:
    raise RuntimeError("❌ T560_BASE still contains placeholder in config/system.json")

ROBOTS = CONFIG.get("robots", {})
CODEX_EXEC_CMD = CONFIG.get("codex_exec_cmd", "bash scripts/codex-exec.sh")

# ====================== SCHEMA + VALIDATION ======================
def validate_action(action: Dict[str, Any]) -> None:
    """Non-negotiable boundary enforcement"""
    if not isinstance(action, dict):
        raise ValueError("Action must be a dict")

    required = ["type", "target", "action"]
    for field in required:
        if field not in action:
            raise ValueError(f"Missing required field: {field}")

    if action["type"] not in ["robot", "code"]:
        raise ValueError(f"Invalid action type: {action['type']}. Must be 'robot' or 'code'")

    if action["type"] == "robot":
        if action["target"] not in ROBOTS:
            raise ValueError(f"Unknown robot target: {action['target']}. Known: {list(ROBOTS.keys())}")
    # Codex actions can target any repo path or task name

    if not isinstance(action.get("params"), dict):
        action["params"] = {}

    print(f"✅ [GOVERLORD] VALIDATED → {action['type']} | {action['target']} | {action['action']}")
# ====================== BACKEND ROUTERS ======================
def _execute_robot(action: Dict[str, Any]) -> bool:
    """Robot backend — 100% T560 only"""
    robot = ROBOTS[action["target"]]
    endpoint = f"{T560_BASE}{robot['prefix']}/{action['action']}"

    try:
        if action["action"] == "say":
            payload = {"text": action["params"].get("text", "")}
            response = requests.post(f"{endpoint}_text", json=payload, timeout=5)  # noqa: F821 (requests imported below)
        else:
            response = requests.post(endpoint, json=action["params"], timeout=5)

        response.raise_for_status()
        print(f"✅ [T560 EXEC] {action['target']} | {action['action']}")
        return True
    except Exception as e:
        print(f"❌ [T560 FAILED] {e}")
        return False
def _execute_code(action: Dict[str, Any]) -> bool:
    """Code-mutation backend — Codex CLI only"""
    task = action.get("params", {}).get("task", action["action"])
    target_path = action.get("params", {}).get("path", "")

    cmd = shlex.split(CODEX_EXEC_CMD) + ["--task", task]
    if target_path:
        cmd.extend(["--path", target_path])

    try:
        result = subprocess.run(
            cmd,
            cwd=Path(__file__).parent.parent,
            capture_output=True,
            text=True,
            timeout=30,
        )
        print(f"✅ [CODEX EXEC] {task} | returncode={result.returncode}")
        if result.stdout:
            print("   stdout:", result.stdout.strip()[:500])
        if result.stderr:
            print("   stderr:", result.stderr.strip()[:500])
        return result.returncode == 0
    except Exception as e:
        print(f"❌ [CODEX FAILED] {e}")
        return False
# ====================== SINGLE DISPATCHER (THE KERNEL) ======================
def goverlord_exec(action: Dict[str, Any]) -> bool:
    """THE ONLY EXECUTION ENTRYPOINT — use this from every tool"""
    validate_action(action)

    print(f"[GOVERLORD ACTION] {json.dumps(action, default=str)}")

    if action["type"] == "robot":
        return _execute_robot(action)
    elif action["type"] == "code":
        return _execute_code(action)

    raise ValueError("Unknown action type after validation")
# ====================== CLI FOR TESTING ======================
if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python3 goverlord_core.py <json_action>   OR")
        print("       python3 goverlord_core.py robot <target> <action> <param_key> <param_value>")
        sys.exit(1)

    if sys.argv[1].startswith("{"):  # JSON mode (Open WebUI friendly)
        action = json.loads(sys.argv[1])
    else:  # Legacy CLI mode for quick testing
        action = {
            "type": sys.argv[1],
            "target": sys.argv[2],
            "action": sys.argv[3],
            "params": {sys.argv[4]: " ".join(sys.argv[5:])} if len(sys.argv) > 5 else {}
        }

    success = goverlord_exec(action)
    sys.exit(0 if success else 1)

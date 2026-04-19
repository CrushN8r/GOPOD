from typing import Any, Dict


def validate_action(action: Dict[str, Any], robots: Dict = None) -> Dict[str, Any]:
    """Non-negotiable boundary enforcement (exact from PDF)"""
    if not isinstance(action, dict):
        raise ValueError("Action must be a dict")

    required = ["type", "target", "action"]
    for field in required:
        if field not in action:
            raise ValueError(f"Missing required field: {field}")

    if action["type"] not in ["robot", "code"]:
        raise ValueError(f"Invalid action type: {action['type']}. Must be 'robot' or 'code'")

    if action["type"] == "robot" and robots and action["target"] not in robots:
        raise ValueError(f"Unknown robot target: {action['target']}. Known: {list(robots.keys())}")

    if not isinstance(action.get("params"), dict):
        action["params"] = {}

    print(f"✅ [GOVERLORD] VALIDATED → {action['type']} | {action['target']} | {action['action']}")
    return action

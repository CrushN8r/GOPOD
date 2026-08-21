#!/usr/bin/env python3
"""Print the golden numpad/NumLock persona mapping table.

Reads numpad_persona_map_001.json (same directory) and prints it - read-only,
never writes. To remap a key, edit that JSON file directly, then rerun this.
"""
import json
import sys
from pathlib import Path

MAP_PATH = Path(__file__).parent / "numpad_persona_map_001.json"
KEYS = [f"KP{n}" for n in range(1, 10)]


def print_fixed_lanes(fixed_lanes):
    print("=== Fixed lanes (always-on, independent of NumLock) ===")
    print(f"{'Lane':<6} {'Action':<52} {'Gesture':<22} Status")
    print("-" * 110)
    for lane, entry in fixed_lanes.items():
        if lane.startswith("_"):
            continue
        action = entry.get("action", "-")
        gesture = entry.get("gesture", "-")
        status = entry.get("status", "-")
        print(f"{lane:<6} {action:<52} {gesture:<22} {status}")


def print_table(layer_name, layer):
    print(f"\n=== NumLock {layer_name} ===")
    print(f"{'Key':<5} {'Persona':<62} {'Body':<12} Status")
    print("-" * 110)
    for key in KEYS:
        entry = layer.get(key, {})
        persona = entry.get("persona", "-")
        body = entry.get("body", "-")
        status = entry.get("status", "-")
        print(f"{key:<5} {persona:<62} {body:<12} {status}")


def main():
    if not MAP_PATH.exists():
        print(f"GOPOD_NUMPAD_MAP_MISSING path={MAP_PATH}")
        sys.exit(1)

    with open(MAP_PATH) as f:
        data = json.load(f)

    if "--json" in sys.argv:
        print(json.dumps(data, indent=2))
        return

    print_fixed_lanes(data.get("fixed_lanes", {}))
    print_table("ON", data.get("numlock_on", {}))
    print_table("OFF", data.get("numlock_off", {}))
    print(f"\nEdit {MAP_PATH} directly to remap - this tool only reads and prints.")


if __name__ == "__main__":
    main()

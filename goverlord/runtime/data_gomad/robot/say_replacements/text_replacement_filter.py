"""Shared outbound text replacement filter for GOPOD-controlled speech."""

from __future__ import annotations

import argparse
import json
import os
import re
import sys
from pathlib import Path
from typing import Any


DEFAULT_RULES_PATH = Path(__file__).resolve().parent / "replacement_rules_001.json"
DEFAULT_PROFILE_ENV = "GOPOD_TEXT_REPLACEMENT_PROFILE"


def normalize_outbound_text(text: str, rules_path: str | None = None) -> str:
    """Apply configured outbound text replacements without risking speech failure."""
    if not isinstance(text, str):
        return text  # type: ignore[return-value]

    original = text
    try:
        payload = _load_rules(Path(rules_path).expanduser() if rules_path else DEFAULT_RULES_PATH)
        if not payload.get("default_enabled", True):
            return original
        enabled_rule_ids = _enabled_rule_ids(payload)
        rules = payload.get("rules", [])
        if not isinstance(rules, list):
            return original
        normalized = original
        indexed_rules = [(index, rule) for index, rule in enumerate(rules) if isinstance(rule, dict)]
        indexed_rules.sort(key=lambda item: (int(item[1].get("order", item[0])), item[0]))
        for _index, rule in indexed_rules:
            if not rule.get("enabled", False):
                continue
            rule_id = str(rule.get("rule_id", ""))
            if enabled_rule_ids is not None and rule_id not in enabled_rule_ids:
                continue
            normalized = _apply_rule(normalized, rule)
        return normalized
    except Exception:
        return original


def _load_rules(path: Path) -> dict[str, Any]:
    try:
        with path.open("r", encoding="utf-8") as handle:
            payload = json.load(handle)
    except (OSError, json.JSONDecodeError):
        return {}
    return payload if isinstance(payload, dict) else {}


def _enabled_rule_ids(payload: dict[str, Any]) -> set[str] | None:
    profiles = payload.get("session_profiles", {})
    if not isinstance(profiles, dict):
        return None
    profile_name = os.environ.get(DEFAULT_PROFILE_ENV, "default").strip() or "default"
    profile = profiles.get(profile_name, profiles.get("default", {}))
    if not isinstance(profile, dict):
        return None
    ids = profile.get("enabled_rule_ids")
    if not isinstance(ids, list):
        return None
    return {str(rule_id) for rule_id in ids}


def _apply_rule(text: str, rule: dict[str, Any]) -> str:
    match = rule.get("match")
    replacement = rule.get("replacement")
    if not isinstance(match, str) or not isinstance(replacement, str) or match == "":
        return text

    case_insensitive = bool(rule.get("case_insensitive", False))
    whole_word = bool(rule.get("whole_word", False))
    if not case_insensitive and not whole_word:
        return text.replace(match, replacement)

    pattern = re.escape(match)
    if whole_word:
        pattern = rf"(?<!\w){pattern}(?!\w)"
    flags = re.IGNORECASE if case_insensitive else 0
    return re.sub(pattern, lambda _match: replacement, text, flags=flags)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Dry-run GOPOD outbound text replacement.")
    parser.add_argument("text", nargs="*", help="Text to normalize.")
    parser.add_argument("--rules-path", default=None)
    args = parser.parse_args(argv)
    print(normalize_outbound_text(" ".join(args.text), args.rules_path))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

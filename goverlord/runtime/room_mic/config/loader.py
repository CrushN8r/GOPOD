"""The room mic's config files and their JSON loader."""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any

from jsonschema import Draft7Validator

CONFIG_DIR = Path(__file__).resolve().parent

ENDPOINTS_PATH = CONFIG_DIR / "endpoints.json"
PATHS_PATH = CONFIG_DIR / "paths.json"
AUDIO_PATH = CONFIG_DIR / "audio.json"
CHAT_ENVELOPE_SCHEMA_PATH = CONFIG_DIR / "chat_envelope_schema.json"
AUDIO_RESAMPLE_RULE_PATH = CONFIG_DIR / "audio_resample.json"


def load_json(path: Path) -> Any:
    """json.load wrapped with a clear error on a missing file - no silent
    fallback to a guessed default, matching the project's no-guess-filling
    rule."""
    try:
        raw = path.read_text(encoding="utf-8")
    except FileNotFoundError as exc:
        raise FileNotFoundError(f"GOPOD_DATA_CONFIG_MISSING path={path}") from exc
    return json.loads(raw)


_envelope_validator: Draft7Validator | None = None


def envelope_validator() -> Draft7Validator:
    """Lazily caches the chat envelope Draft7Validator - one schema load
    per process, shared by every write site that validates an envelope."""
    global _envelope_validator
    if _envelope_validator is None:
        schema = load_json(CHAT_ENVELOPE_SCHEMA_PATH)
        _envelope_validator = Draft7Validator(schema)
    return _envelope_validator


def validate_envelope(envelope: dict[str, Any]) -> None:
    """Raises ValueError with every violation joined into one message if
    envelope does not conform to chat_envelope_schema.json. Callers catch
    ValueError, log, and skip the write - never let this crash a caller."""
    errors = sorted(envelope_validator().iter_errors(envelope), key=lambda e: e.path)
    if errors:
        messages = [f"{'.'.join(str(p) for p in e.path) or '<root>'}: {e.message}" for e in errors]
        raise ValueError(f"Envelope schema violation: {'; '.join(messages)}")

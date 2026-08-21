"""Path constants + JSON loader for goverlord/runtime/data_gomad/.

Moved 2026-07-30: this module lives under runtime/data_gomad/configs/ now
that the gomads/ tree has been removed (only data_gomad ever had a live
caller; see gopod_notes/GOVERLORD_DEAD_WEIGHT_SURVEY_001.md).
"""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any

from jsonschema import Draft7Validator

REPO_ROOT = Path(__file__).resolve().parents[4]
DATA_DIR = REPO_ROOT / "goverlord" / "runtime" / "data_gomad"
RULES_DIR = REPO_ROOT / "goverlord" / "rules"

ENDPOINTS_PATH = DATA_DIR / "configs" / "endpoints.json"
PATHS_PATH = DATA_DIR / "configs" / "paths.json"
AUDIO_PATH = DATA_DIR / "configs" / "audio.json"
CHAT_ENVELOPE_SCHEMA_PATH = DATA_DIR / "configs" / "chat_envelope_schema.json"
AUDIO_RESAMPLE_RULE_PATH = RULES_DIR / "audio_resample.json"


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

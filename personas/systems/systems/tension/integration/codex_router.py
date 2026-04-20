from __future__ import annotations

from pathlib import Path
from typing import Any

try:
    import yaml
except ImportError:  # pragma: no cover
    yaml = None


def _default_profile_path() -> Path:
    return Path(__file__).resolve().parents[3] / "config" / "codex_profile.yaml"


def load_codex_profiles(profile_path: str | None = None) -> dict[str, Any]:
    resolved_path = Path(profile_path) if profile_path else _default_profile_path()
    if not resolved_path.exists() or yaml is None:
        return {}

    try:
        profile_data = yaml.safe_load(resolved_path.read_text(encoding="utf-8"))
    except Exception:
        return {}

    if not isinstance(profile_data, dict):
        return {}

    return profile_data


def get_active_codex_profile(profile_path: str | None = None) -> dict[str, Any]:
    profile_data = load_codex_profiles(profile_path)
    active_profile = profile_data.get("active_codex_profile")
    if not isinstance(active_profile, str):
        return {}

    selected_profile = profile_data.get(active_profile)
    if not isinstance(selected_profile, dict):
        return {}

    return {
        "active_codex_profile": active_profile,
        "profile": selected_profile,
    }

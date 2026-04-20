from __future__ import annotations

from pathlib import Path

from persona.schema import PersonaDefinition


PERSONA_NOTES_ROOT = "persona/notes"


def _default_base_path() -> Path:
    return Path(__file__).resolve().parents[3]


def load_persona_registry(base_path: str | None = None) -> tuple[PersonaDefinition, ...]:
    root = Path(base_path) if base_path else _default_base_path()
    notes_root = root / PERSONA_NOTES_ROOT
    if not notes_root.exists():
        return ()

    personas: list[PersonaDefinition] = []
    for persona_dir in sorted(path for path in notes_root.iterdir() if path.is_dir()):
        personas.append(
            PersonaDefinition(
                persona_id=persona_dir.name,
                notes_path=str(persona_dir / "notes.md"),
                constraints_path=str(persona_dir / "constraints.yaml"),
            )
        )
    return tuple(personas)

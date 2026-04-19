from pydantic import Field

class Tools:
    def __init__(self):
        pass

    def cdx1(self, task: str = Field(..., description="The coding / repo task to execute via Codex CLI inside the GOPOD repository.")) -> str:
        """
        Execute a coding task using Codex CLI inside the GOPOD repo.
        Routes through Goverlord → Codex execution layer only.
        """
        from tools.codex.cdx1 import cdx1

        return cdx1(task)

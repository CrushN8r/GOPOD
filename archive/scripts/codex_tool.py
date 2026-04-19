class Tools:
    @staticmethod
    def codex_exec(task: str) -> str:
        """
        Execute a coding task using Codex CLI inside GOPOD repo
        """
        from tools.codex.codex_exec import codex_exec

        return codex_exec(task)

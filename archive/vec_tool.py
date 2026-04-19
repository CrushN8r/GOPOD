from pydantic import Field

class Tools:
    def __init__(self):
        pass

    def vec1(self, text: str = Field(..., description="The text Vector-N7V3 will speak.")) -> str:
        """
        Make Vector-N7V3 speak (routed through unified vec tool).
        """
        from tools.vector.vec1 import vec1

        return vec1(text)

    def vec2(self, text: str = Field(..., description="The text Vector-T3P1 will speak.")) -> str:
        """
        Make Vector-T3P1 speak (routed through unified vec tool).
        """
        from tools.vector.vec2 import vec2

        return vec2(text)

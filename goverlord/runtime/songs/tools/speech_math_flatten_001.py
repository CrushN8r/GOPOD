# goverlord/runtime/songs/tools/speech_math_flatten_001.py
"""Math and LaTeX flatten for the robot speech lane.

One Doc reply, two outputs: the rich display keeps everything (LaTeX, symbols, bold,
emoji); the robot says a flattened copy. This module turns the math parts of that copy
into spoken words, so nothing unspeakable reaches a Vector:

  \\frac{a}{b} -> "a over b"         x^{4} / x⁴ -> "x to the fourth"
  \\neq / ≠   -> "does not equal"    \\times / × -> "times"
  \\infty / ∞ -> "infinity"          \\lambda / λ -> "lambda"
  \\text{belief} -> "belief"         \\sqrt{x} -> "the square root of x"
  \\overline{3} -> "3 repeating"     spacing commands, braces, backslashes -> gone

Display math (\\[ ... \\] and $$ ... $$) becomes its own spoken sentence. Decimals are
left intact (nothing here splits text into sentences). Every table is an ordered list,
applied in a fixed order, so the same input always gives the same output.

Called from run_section1_full_live_001.py's speech path (strip_rich_text_for_voice and
flatten_for_robot_speech) before any character stripping, so the LaTeX is still whole
when it is read. Ideas taken from the operator's old Go attempts (wirepod_ttr
llm_streaming_prepost_split.go, rational_number_utils*.go, special_characters_vars.go),
rebuilt here without their known failures. Report:
gopod_notes/DOC_RICH_DISPLAY_ROBOT_SAY_FLATTEN_001.md
"""
import re

# ---- Ordered tables -------------------------------------------------------------------
# LaTeX word commands, spoken form. Checked in this order.
LATEX_WORD_COMMANDS = [
    ("neq", "does not equal"), ("ne", "does not equal"),
    ("times", "times"), ("cdot", "times"), ("div", "divided by"),
    ("pm", "plus or minus"), ("mp", "minus or plus"),
    ("infty", "infinity"),
    ("approx", "approximately equals"), ("sim", "is about"), ("equiv", "is identical to"),
    ("leq", "is less than or equal to"), ("le", "is less than or equal to"),
    ("geq", "is greater than or equal to"), ("ge", "is greater than or equal to"),
    ("lt", "is less than"), ("gt", "is greater than"),
    ("propto", "is proportional to"),
    ("to", "goes to"), ("rightarrow", "goes to"), ("Rightarrow", "implies"),
    ("implies", "implies"), ("iff", "if and only if"),
    ("sum", "the sum of"), ("prod", "the product of"), ("int", "the integral of"),
    ("partial", "partial"), ("nabla", "del"),
    ("degree", "degrees"), ("circ", "degrees"), ("percent", "percent"),
    ("ldots", "and so on"), ("cdots", "and so on"), ("dots", "and so on"),
    ("in", "in"), ("cup", "union"), ("cap", "intersect"),
    ("log", "log"), ("ln", "natural log"), ("sin", "sine"), ("cos", "cosine"), ("tan", "tangent"),
    ("exp", "exp"), ("min", "min"), ("max", "max"), ("lim", "the limit of"),
]
GREEK_LETTERS = [
    ("alpha", "alpha"), ("beta", "beta"), ("gamma", "gamma"), ("delta", "delta"),
    ("epsilon", "epsilon"), ("varepsilon", "epsilon"), ("zeta", "zeta"), ("eta", "eta"),
    ("theta", "theta"), ("vartheta", "theta"), ("iota", "iota"), ("kappa", "kappa"),
    ("lambda", "lambda"), ("mu", "mu"), ("nu", "nu"), ("xi", "xi"), ("pi", "pi"),
    ("rho", "rho"), ("sigma", "sigma"), ("tau", "tau"), ("upsilon", "upsilon"),
    ("phi", "phi"), ("varphi", "phi"), ("chi", "chi"), ("psi", "psi"), ("omega", "omega"),
    ("Gamma", "gamma"), ("Delta", "delta"), ("Theta", "theta"), ("Lambda", "lambda"),
    ("Xi", "xi"), ("Pi", "pi"), ("Sigma", "sigma"), ("Phi", "phi"), ("Psi", "psi"), ("Omega", "omega"),
]
TEXT_COMMANDS = ["text", "textrm", "textbf", "textit", "mathrm", "mathbf", "mathit", "mathsf", "operatorname", "mbox"]
SPACING_COMMANDS = [";", ",", ":", "!", " ", "quad", "qquad", "enspace", "thinspace"]
DROPPED_COMMANDS = ["left", "right", "big", "Big", "bigg", "Bigg", "bigl", "bigr", "displaystyle", "limits"]

# Unicode math in plain prose (outside LaTeX), spoken form. Checked in this order.
UNICODE_SYMBOLS = [
    ("≠", "does not equal"), ("≈", "approximately equals"), ("≡", "is identical to"),
    ("≤", "is less than or equal to"), ("≥", "is greater than or equal to"),
    ("±", "plus or minus"), ("∓", "minus or plus"),
    ("×", "times"), ("÷", "divided by"),
    ("√", "the square root of"), ("∞", "infinity"), ("∝", "is proportional to"),
    ("°", "degrees"), ("∑", "the sum of"), ("∏", "the product of"), ("∫", "the integral of"),
    ("∆", "delta"), ("→", "goes to"), ("⇒", "implies"),
]
UNICODE_GREEK = [
    ("α", "alpha"), ("β", "beta"), ("γ", "gamma"), ("δ", "delta"), ("ε", "epsilon"),
    ("θ", "theta"), ("λ", "lambda"), ("μ", "mu"), ("π", "pi"), ("ρ", "rho"),
    ("σ", "sigma"), ("τ", "tau"), ("φ", "phi"), ("χ", "chi"), ("ψ", "psi"), ("ω", "omega"),
    ("Γ", "gamma"), ("Δ", "delta"), ("Θ", "theta"), ("Λ", "lambda"), ("Π", "pi"),
    ("Σ", "sigma"), ("Φ", "phi"), ("Ψ", "psi"), ("Ω", "omega"),
]
SUPERSCRIPT_DIGITS = {"⁰": "0", "¹": "1", "²": "2", "³": "3", "⁴": "4", "⁵": "5",
                      "⁶": "6", "⁷": "7", "⁸": "8", "⁹": "9", "⁻": "-"}
ORDINALS = {4: "fourth", 5: "fifth", 6: "sixth", 7: "seventh", 8: "eighth", 9: "ninth",
            10: "tenth", 11: "eleventh", 12: "twelfth", 13: "thirteenth", 14: "fourteenth",
            15: "fifteenth", 16: "sixteenth", 17: "seventeenth", 18: "eighteenth",
            19: "nineteenth", 20: "twentieth"}
ROOT_NAMES = {"2": "square", "3": "cube"}
# ----------------------------------------------------------------------------------------

_WORD_COMMANDS = dict(LATEX_WORD_COMMANDS + GREEK_LETTERS)
_GREEK_NAMES = "|".join(sorted({w for _, w in GREEK_LETTERS}, key=len, reverse=True))


def _ordinal(n):
    if n in ORDINALS:
        return ORDINALS[n]
    suffix = "th" if 10 <= n % 100 <= 20 else {1: "st", 2: "nd", 3: "rd"}.get(n % 10, "th")
    return f"{n}{suffix}"


def power_words(exponent):
    exp = " ".join(str(exponent).split())
    if exp == "2":
        return "squared"
    if exp == "3":
        return "cubed"
    if exp.isdigit():
        return "to the " + _ordinal(int(exp))
    return "to the power of " + exp


def _read_group(s, i):
    """Brace group starting at s[i] (nested braces allowed), else one token. Returns (text, next_i)."""
    while i < len(s) and s[i] == " ":
        i += 1
    if i >= len(s):
        return "", i
    if s[i] == "{":
        depth, j = 0, i
        while j < len(s):
            if s[j] == "{":
                depth += 1
            elif s[j] == "}":
                depth -= 1
                if depth == 0:
                    return s[i + 1:j], j + 1
            j += 1
        return s[i + 1:], len(s)
    if s[i] == "\\":
        m = re.match(r"\\([A-Za-z]+|.)", s[i:])
        return m.group(0), i + len(m.group(0))
    if s[i].isdigit():
        m = re.match(r"\d+(\.\d+)?", s[i:])
        return m.group(0), i + len(m.group(0))
    return s[i], i + 1


def speak_math(expr):
    """Turn one LaTeX math expression into spoken words."""
    s, out, i = str(expr), [], 0
    while i < len(s):
        c = s[i]
        if c == "\\":
            m = re.match(r"\\([A-Za-z]+|.)", s[i:])
            name = m.group(1)
            i += len(m.group(0))
            if name in ("frac", "dfrac", "tfrac"):
                a, i = _read_group(s, i)
                b, i = _read_group(s, i)
                out.append(f" {speak_math(a)} over {speak_math(b)} ")
            elif name in TEXT_COMMANDS:
                a, i = _read_group(s, i)
                out.append(" " + a.replace("{", " ").replace("}", " ") + " ")
            elif name == "sqrt":
                n = "2"
                if s[i:i + 1] == "[" and "]" in s[i:]:
                    close = s.index("]", i)
                    n, i = s[i + 1:close].strip(), close + 1
                a, i = _read_group(s, i)
                root = ROOT_NAMES.get(n) or (_ordinal(int(n)) if n.isdigit() else speak_math(n).strip())
                out.append(f" the {root} root of {speak_math(a)} ")
            elif name == "overline":
                a, i = _read_group(s, i)
                # 0.\overline{3} -> "0 point 3 repeating": a bare "0." here would later be
                # read as the end of a sentence and split the number in two.
                sofar = "".join(out).rstrip()
                if re.search(r"\d\.$", sofar):
                    out = [sofar[:-1] + " point"]
                out.append(f" {speak_math(a)} repeating ")
            elif name in _WORD_COMMANDS:
                out.append(" " + _WORD_COMMANDS[name] + " ")
            elif name in SPACING_COMMANDS:
                out.append(" ")
            elif name in DROPPED_COMMANDS or name in ("{", "}"):
                pass
            elif name == "%":
                out.append(" percent ")
            else:
                out.append(" " + name + " ")  # unknown command: say its name, never the backslash
        elif c == "^":
            exp, i = _read_group(s, i + 1)
            out.append(" " + power_words(speak_math(exp)) + " ")
        elif c == "_":
            sub, i = _read_group(s, i + 1)
            out.append(" sub " + speak_math(sub) + " ")
        elif c in "{}":
            i += 1
        elif c in "=+-/*<>":
            out.append({"=": " equals ", "+": " plus ", "-": " minus ", "/": " over ",
                        "*": " times ", "<": " is less than ", ">": " is greater than "}[c])
            i += 1
        else:
            out.append(c)
            i += 1
    return _prose_symbols(" ".join("".join(out).split()))


def _prose_symbols(text):
    # Superscripts first, while their base is still attached: λ⁴ -> λ to the fourth.
    def sup(m):
        digits = "".join(SUPERSCRIPT_DIGITS[ch] for ch in m.group(2))
        return f"{m.group(1)} {power_words(digits.replace('-', 'minus '))}"
    text = re.sub(r"(\S)([⁰¹²³⁴⁵⁶⁷⁸⁹⁻]+)", sup, text)
    for sym, words in UNICODE_SYMBOLS:
        text = text.replace(sym, f" {words} ")
    for sym, words in UNICODE_GREEK:
        text = text.replace(sym, f" {words} ")
    # 1/lambda -> 1 over lambda: a number over a single letter or a Greek name only.
    # Number/number (24/7, 9/28) and words (2/day) are left alone.
    text = re.sub(r"(\d)\s*/\s*(?=(?:[A-Za-z]|" + _GREEK_NAMES + r")\b)", r"\1 over ", text)
    return text


_DISPLAY_MATH = re.compile(r"\\\[(.+?)\\\]|\$\$(.+?)\$\$", re.S)
_INLINE_MATH = re.compile(r"\\\((.+?)\\\)", re.S)
_DOLLAR_MATH = re.compile(r"(?<!\$)\$(?!\$)([^$\n]*\\[^$\n]*)(?<!\$)\$(?!\$)")  # only $...$ holding a \command
_BARE_LATEX = re.compile(r"[\w.{}^_/=+\-]*\\[A-Za-z]+(?:\s*\{[^{}]*\})*[\w.{}^_/=+\-]*")
_BOLD_LINE = re.compile(r"^(\s*)(\*\*|__)(.+?)(\*\*|__)\s*$", re.M)


def flatten_math_for_speech(text):
    """Rich text in, speakable text out (math only - emoji, markdown and pronunciation are
    handled by the existing say path after this runs)."""
    s = str(text or "")
    if not s.strip():
        return s
    # A bold-only line (a header) gets a closing period so it doesn't run into the next sentence.
    s = _BOLD_LINE.sub(lambda m: m.group(1) + m.group(2) + m.group(3).rstrip()
                       + ("" if re.search(r"[.!?:]$", m.group(3).rstrip()) else ".") + m.group(4), s)
    s = _DISPLAY_MATH.sub(lambda m: " " + speak_math(m.group(1) or m.group(2)).strip() + ". ", s)
    s = _INLINE_MATH.sub(lambda m: " " + speak_math(m.group(1)).strip() + " ", s)
    s = _DOLLAR_MATH.sub(lambda m: " " + speak_math(m.group(1)).strip() + " ", s)
    s = _BARE_LATEX.sub(lambda m: " " + speak_math(m.group(0)).strip() + " ", s)
    s = _prose_symbols(s)
    # Tidy: no space before punctuation, no doubled sentence stops from the display blocks.
    s = re.sub(r"[ \t]+([.,!?;:])", r"\1", s)
    s = re.sub(r"([:;,])\.", r"\1", s)
    s = re.sub(r"\.(\s+\.)+", ".", s)  # ". ." only - "..." (ellipsis, action separator) is kept
    s = re.sub(r"[ \t]{2,}", " ", s)
    return s

# ## EOF

# Persona chains - every participant's thought chains

DocBS's template (`gopod_probes/docbs/`) for every chat participant: the Brobots, CHALK, PLAYHEAD,
the interns, Cameo (settings `persona_chains`, `gopod_persona_chains.go`).

- One folder per persona, its name in small letters with dashes: `brobot-1/`, `brobot-2/`, `chalk/`,
  `playhead/`, `git-repo/`, `cache-pyc/`, `intel-404-pnfh/`, `intel-404-pnft/`, `cameo/`.
- One file per beat, one line per entry: `<pool>_clean.txt` (no swears) and `<pool>_18.txt` (all,
  starred), or just `<pool>.txt`. A private, uncensored `z<pool>_18.txt` is used first when it's there.
- A chain (in the settings): `{"name": "...", "min_share": 0, "beats": ["openers", "reply", "closers"]}` -
  the beats before `reply` come before the persona's answer, the ones after it close it; `"a|b"` is
  one of two pools.
- `share` (0-100): how often the persona's answers are a chain.

Empty to start: the lines are the operator's.

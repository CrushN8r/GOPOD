# Persona profiles - self and chat awareness

One profile per chat participant (2026-10-30, operator: "fatten up all these personas for self & chat
awareness * offload to persona profile md files for look ups"; "These will be your first workings of a
brain bubble template for scaffolding for the other personas"). The file name is the persona's name in
small letters with dashes: `chalk.md`, `playhead.md`, `git-repo.md`, `cache-pyc.md`, `butler.md` (the persona ID, his job name), `cameo.md`, `intel-404-pnfh.md`, `intel-404-pnft.md`. A profile can use `{name}`, `{nickname}` and `{job}`, filled from the persona's labels (`host_personas`).

Every profile has the same sections, in this order:

- **Self** - who I am: my informal name (nickname), my name, my job title, my role, who helps me.
- **Voice** - how I talk.
- **Partner and banter** - my partner, and what we call each other.
- **The name** - what it means and how it's said (for reading, not said in answers).

The persona's prompt looks up the sections named in `persona_profiles.sections` (settings), one line
per bullet (`gopod_persona_profiles.go`). Edit a bullet, then `wpr 2`.

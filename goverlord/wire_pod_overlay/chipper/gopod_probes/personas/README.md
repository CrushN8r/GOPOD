# Persona profiles - self and chat awareness

One profile per chat participant (2026-10-30, operator: "fatten up all these personas for self & chat
awareness * offload to persona profile md files for look ups"; "These will be your first workings of a
brain bubble template for scaffolding for the other personas"). The file name is the persona's name in
small letters with dashes: `chalk.md`, `playhead.md`, `git-repo.md`, `cache-pyc.md`, `butler.md` (the persona ID, his job name), `cameo.md`, `intel-404-pnfh.md`, `intel-404-pnft.md`. A profile can use `{name}`, `{nickname}` and `{job}`, filled from the persona's labels (`host_personas`).

Every profile is a topic-cluster md, GOPOD's master format (`../_topic_cluster/README.md`):

- **Its head** - `# <full name>` (the persona's name, first middle last), `- **nickname:**`,
  `- **job:**` (its job name), `- **words:**` (the words that name it), `- **ask:**`. These are the
  persona's labels, written here and nowhere else: the settings, the prompts, the headers and the
  brain bubble fill from them (`{name}`, `{nickname}`, `{job}`).
- `---`
- **Self** - who I am: my informal name (nickname), my name, my job title, my role, who helps me.
- **Voice** - how I talk.
- **Partner and banter** - my partner, and what we call each other.
- **The name** - what it means and how it's said (for reading, not said in answers).

Each section's points are numbered `1.` `2.` ... A point ending "(to come)" is never said.
The persona's prompt looks up the sections named in `persona_profiles.sections` (settings)
(`gopod_persona_profiles.go`). Edit a line, then `wpr 2`.

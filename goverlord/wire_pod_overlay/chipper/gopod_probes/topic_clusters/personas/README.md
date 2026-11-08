# Persona profiles - self and chat awareness

One profile per chat participant (2026-10-30, operator: "fatten up all these personas for self & chat
awareness * offload to persona profile md files for look ups"; "These will be your first workings of a
brain bubble template for scaffolding for the other personas"). The file name is the persona's name in
small letters with dashes: `brobot-1.md` (Doc), `brobot-2.md` (Pip), `chalk.md`, `playhead.md`, `git-repo.md`, `cache-pyc.md`, `butler.md` (the persona ID, his job name), `cameo.md`, `intel-404-pnfh.md`, `intel-404-pnft.md`. A profile can use `{name}`, `{nickname}` and `{job}`, filled from the persona's labels (`host_personas`).

Every profile is a topic-cluster md, GOPOD's master format (`../_topic_cluster/README.md`):

- **Its head** - `# <full name>` (the persona's name, first middle last), `- **nickname:**`,
  `- **job:**` (its job name), `- **words:**` (the words that name it), `- **ask:**`. These are the
  persona's labels, written here and nowhere else: the settings, the prompts, the headers and the
  brain bubble fill from them (`{name}`, `{nickname}`, `{job}`).
- `---`
- **Self** - my card: my informal name (nickname, for banter only), my name (formal), my job title (one line, what I own).
- **Who I answer to** - one boss line.
- **Who works with me** - names, and what for.
- **Voice** - how I talk: my take.
- **Banter** - nicknames, mine and the others', and their stories. Company time is formal names; nicknames are for banter, anyone with anyone.
- **The name** - what it means and how it's said (for reading, not said in answers).
- **Lingo** - how my names are said: `word = said` (`CHALK = chock`). Every voice (`reporter_voices.say_words` for the reporter voices, the pronunciation list for the robots), mine winning, the longest word first. Names are written in caps; the lingo gives the voice normal case.

The brain bubble's "My relationships:" is my **Who I answer to** and **Who works with me** - the same lines my intro says (2026-11-06, the persona schema: `gopod_notes/PERSONA_SCHEMA_DRAFT_002.md`).

Each section's points are numbered `1.` `2.` ... A point ending "(to come)" is never said.
The persona's prompt looks up the sections named in `persona_profiles.sections` (settings)
(`gopod_persona_profiles.go`). Edit a line, then `wpr 2`.

**The Brobots' profiles** (`brobot-1.md`, `brobot-2.md`) have their own sections, the words their
prompt's "WHO I AM" is built from: **Who I am** (his persona), **Role**, **My way through an answer**
(his steps), **Signature** - and a **Self** like everyone's, for the intro recipe. The pair's banter and stereo dial stay in `host_pairs`; the shared
party-bro base stays in `../show/brobots.md` (the shared Brobots persona).

Each persona file names its group's shared root (`- **root:**`, in `groups/`), and every root sits under
`company/math-aftermath-news-channel.md` (GOPOD's company, filled from `company/company-schema.md`): the root's lines come first, then the persona's own.

**The schemas (2026-11-06, operator: "Consistency would then get checked on every restart instead of by you reading runs"):**
`personas/persona_chains/persona-schema.md` is the master persona card and `personas/company/company-schema.md` the master
company - sterile scaffolding, copied to make a new one. The system check (`gopodSchemaCheck`, every restart) measures every
persona file (each participant's, and the persona a robot hosts) and every company file (one with `## Our company`) against
them: a missing head line, section, Self line or company point is listed "to fix" in the journal.

**4 personas, 2 robots (2026-11-06, operator: "ensure DOC & PIP are split from brobots * 4 personas, DOC, PIP, 2 BROBOTS"):** `company/gominions/brobot-1.md` / `brobot-2.md` are the Brobots' own personas (party-bro GOPOD reps, BRUH!) at the dial's 0% - a line to both (KP*). `company/hosts/barfallonyou-burns.md` (DOC) and `pip-stall.md` (PIP) are the host and sidekick at 100% - KP1, KP2 - each with its own card, words and "Who I am"; the robot's file names it in `hosts:`.

**Names are variables, never quoted (2026-11-08):** a persona's name, nickname and job are written once, in its own file's head. Everywhere else - its own sections (`{name}`, `{nickname}`, `{job}`) and other files (`{Brobot 1.name}`, `{BUTLER.nickname}`, `{CHALK.job}`, `{Brobot 1.call}` for the call name) - they're variables, filled from that one source (`gopodPersonaFill`).

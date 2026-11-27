# Personas

One card per chat participant. The file name is the persona's full name in small letters with
dashes (`chalk-bored.md`, `keynote-vale.md`); a card is found by its file name wherever it sits.

## The folders are the chain of command
- `company/math-aftermath-news-channel.md` holds the chain line.
- Under `company/`, the folder a card sits in is its crew: the top folder the crew, the folders
  under it its part. KEYS' crew bubble reads them live (`{crews}`); move or rename a folder and it
  follows.
- Each card names its one boss in **Who I answer to**; each names its group's shared root
  (`- **root:**`, the `group-*.md` in `company/4_groups/`). The root's lines come first, then the
  persona's own.
- `behaviours/` holds how personas act (answers, dials, weather, photos ...); `brobots.md` there is
  the shared Brobots persona.

## A persona card
- **Head** - `# <full name>`, `- **nickname:**`, `- **job:**`, `- **words:**`, `- **ask:**`. Written
  here and nowhere else: settings, prompts, headers and the brain bubble fill from them
  (`{name}`, `{nickname}`, `{job}`; other cards use `{DOC.name}`, `{BUTLER.nickname}` ...).
- `---`
- **Self** - nickname (banter only), formal name, job title.
- **Who I answer to** - one boss line.
- **Who works with me** - names, and what for.
- **Voice** - how I talk.
- **Banter** - nicknames and their stories. Company time is formal names.
- **The name** - what it means and how it's said (not said in answers).
- **Lingo** - how my names are said: `word = said` (`CHALK = chock`).

Points are numbered `1.` `2.` ...; a point ending "(to come)" is never said. The brain bubble's
"My relationships:" is **Who I answer to** and **Who works with me**. Edit a line, then `wpr 2`.

## Four personas, two robots
- `brobot-1.md`, `brobot-2.md` - the Brobots themselves, party-bro GOPOD reps, at the dial's 0%.
- `barfallonyou-spewer.md` (DOC) and `pippen-stall.md` (PIP) - host and sidekick at the dial's 100%,
  answering from the same two robots. A robot's card names the persona it hosts (`hosts:`).
- The Brobots' cards add **Who I am**, **Role**, **My way through an answer** and **Signature**.

## The schemas
`persona-schema.md` is the master persona card, `company/company-schema.md` the master company
card - copy one to make a new one. Every restart checks each card against them; anything missing
is listed "to fix" in the journal.

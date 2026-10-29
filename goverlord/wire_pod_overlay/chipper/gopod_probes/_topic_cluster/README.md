# Topic cluster template

A topic cluster is one chapter of the GOPOD game: a folder with a `chapter.md` and one topic
file per topic - readable markdown files, the same format everywhere. The game's topics drive the repo: each topic tells one thing GOPOD presents,
names the repo file that tells it in full, and shows up on [BROBOT_ASKS.md](../../../../../BROBOT_ASKS.md).

**To start a new cluster:** copy this folder into `gopod_probes/chapters/` as
`NN_your_cluster/` (`NN` sets its place among the chapters), rename `01_example_topic.md`
(`NN_topic_name.md` - the name after the number is the topic's name), and fill in each line.
This folder itself is never loaded: only `chapters/` is.

## `chapter.md`

| Line | What it is |
|---|---|
| `# <title>` | The chapter's name - it's also the chapter egg's title. |
| `- **summary:**` | One line on what the chapter covers. |

## A topic file

Its label as the heading, each setting a list line (`- **words:** rig, gear`), a `---` line
(a blank line on both sides), then the facts the Brobots get. `system-topic` keeps the same format.

| Line | What it is |
|---|---|
| `# <label>` | How the Brobots name the topic in a reminder ("the rig"). No heading = never named in a reminder. |
| `- **words:**` | The words in a question that load the topic, comma separated. |
| `- **ask:**` | Its ask on BROBOT_ASKS.md. Left out = "Tell me about <label>". |
| `- **summary:**` | What a summary ask ("Tell me about the rig", "the rig?") shows in the notice bubble and the Brobot riffs on. Pieces split by `\|`. |
| `- **source:**` | The repo file that tells this topic in full. `system-check` flags the topic when that file changes after it. |
| `- **pictures:**` | Pictures shown with the topic, from `webroot/pictures/` (with their folder). |
| `- **reminder:**` | The line a Brobot uses to remind the room of this topic. |
| `- **pre:**` | The Brobot's note before the room has covered this topic - sell it with attitude, never make a fact up. Left out = `topic_asks.pre_default`. |
| `- **post:**` | His note once it's covered (a value point given) - build on it, don't repeat. Left out = `topic_asks.post_default`. |
| `- **closing:**` | A closing remark at the end of his answer once it's covered - like Bingo's "I Won!" Left out = none. |
| `- **say N:**` | Value point N's exact line, said word for word - for names and numbers the model gets wrong. |
| `- **notes N:**` | Value point N in note form: `title \| note \| note`. The model flavours the notes; a point with notes has no exact line. |

After `---`: an opening line, then the value points, numbered `1.` `2.` `3.`, revealed one at a
time.

A song's before and after is different: whether it has played - its `experiences` entry in the
settings (Bingo, the nap, the fart joke, the interview).

Optional lines: `- **sentences:**` (its own sentence budget), `- **start line:**` / `- **end line:**` (a scripted
line at the start or end of the answer).

## After adding a cluster

1. Add a game question per value point to `interview_harvest/harvest_questions.json` - the
   cross-check that each point can be won.
2. Run the build (the game sweep and `TestBrobotAsksPageMatches` check it), then write the asks
   page again: `GOPOD_WRITE_ASKS=1 deploy_overlay_build.sh`.
3. `system-value-points lock` in the chat, then `system-check`.

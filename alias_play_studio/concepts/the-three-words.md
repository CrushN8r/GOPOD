# The three words
- **words:** golden note, sequence, golden sequence, piano, mixer, sequencer
- **ask:** What are a golden note, a sequence and a golden sequence?
---

| Word | What it is | In a song folder |
|---|---|---|
| **Golden note** | The smallest move or line the robots can play, proven live on its own first. | One step's `note` type in `knobs.json` (`arm_cue`, `rattle`, `say_turn`, ...). |
| **Sequence** | A few notes in order that make one moment. | One `## SECTION` in `story.md`, the steps' `section` in `knobs.json`. |
| **Golden sequence** | A sequence, or a whole song, that's locked: it plays the same way every time and is ready to record. | A song whose status line says `GOLDEN`, or a locked opener (the Bingo test's `OPENING SYNC`). |

The **piano** is the notes. The **mixer** is how each note plays (its hold, speed, buffer,
tempo). The **sequencer** is their order. All three live in the song folder: no song's moves are
hidden in code.

[Concepts](README.md) · [ALIAS-PLAY-STUDIO](../../ALIAS-PLAY-STUDIO.md)

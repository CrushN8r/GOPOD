# The chat words
- **words:** operator bubble, system bubble, brobot bubble, reporter bubble, egg bubble, highlight, the baton, chat page
- **ask:** What are the chat page's bubbles called?
---

| Say | It means | Where it lives |
|---|---|---|
| **Operator bubble** | A line from you to the Brobots, white-eyed icon, labelled by its input: "Keyboard to Brobot 1", "KP1 to Brobot 1", "Numpad to <who>", "Backpack to Brobot 2" (heard by the robot himself: Hey Vector or the backpack button), "Mic to Brobot N", "Reporter 1 to Brobot 1" (a scripted run's question), "… to Brobots" (both icons). It goes up once the bubble on screen has finished (the baton). | the chat |
| **System bubble** | Goverlord's grey-eyed bubbles (warm-up, countdowns, chapters, system messages). Reporter and narrator bubbles are system bubbles with white eyes. | the chat |
| **System llm bubble** | The one warm-up bubble a run opens with ("Goverlord: GOPOD Brain", LEDs, hourglass). | `llm_warmup` |
| **System message** | A Goverlord line in a grey bubble, said in the system voice ("Brobots. Go to sleep now please.", "Run complete."). | note `system_line` |
| **Goverlord's terminal bubble** | Under Goverlord: where system commands answer - Goverlord's header, text at 50% size (`system_set.role`, the role's name inside only). | system- commands |
| **Brobot bubble** | A Brobot's line, in his eye colours, each piece lit as it's said. | note `say_turn`, a reply |
| **Reporter bubble** | CHALK or PLAYHEAD's scripted voice (Reporter 1 / 2 in scripts), headed by the header template (`display_names`: CHALK / PLAYHEAD), white-eyed icons. | the interview, `ask` |
| **Narrator bubble** | A timed song's captions, one bubble per section, each phrase lit as it's read. | `robots_speak: false` |
| **Manager bubble** | CHALK's chalkboard or PLAYHEAD's tape-brown bubble, or their part inside a Brobot's bubble. | `managers` |
| **Egg bubble** | A found easter egg, ringed by marquee bulbs. | [eggs](../06_CAST_AND_GAME.md#eggs) |
| **Chapter** | Goverlord's title between parts of the interview. | the interview |
| **Warm-up bubble** | "Goverlord: GOPOD Brain", the LEDs and hourglass, then the ready message. | `llm_warmup` |
| **Countdown** | The ⏳ bubble with juggled balls, under a rattle. | note `rattle` |
| **Big ball** | The Bingo ball bubble. | `show_ball` |
| **Highlight** | The lit piece in a bubble: the words being said or sung now. | [9.5](../04_CHAT.md#95-highlight-and-word-reveal) |
| **The baton** | The next bubble waits hidden until the one on screen finishes. | [9.3](../04_CHAT.md#93-the-chat-rules) |
| **Operator line** | `Brobots 1 2 All [Type to the Brobot, press Enter] \| Record? [] \| HH:MM:SS 🔇` under the chat. | the chat |
| **System menu** | The menu that opens from the speaker icon (🔇/🔊), listing every system command in tour order. | [the system commands](../07_SYSTEM.md#the-system-commands) |

[Concepts](README.md) · [ALIAS-PLAY-STUDIO](../../ALIAS-PLAY-STUDIO.md)

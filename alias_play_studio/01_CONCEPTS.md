# Concepts and terms

Part of [ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md) - [Concepts and terms](01_CONCEPTS.md) · [Songs](02_SONGS.md) · [The session and runs](03_SESSION.md) · [The Brobots chat](04_CHAT.md) · [Voices and sound](05_VOICES_AND_SOUND.md) · [The cast and the game](06_CAST_AND_GAME.md) · [The system menu, your stuff, the settings](07_SYSTEM.md) · [Golden rules, troubleshooting, tests](08_RULES_AND_TESTS.md) · [The gears](09_GEARS.md)

## 1. Concepts and terms

### 1.1 The cast

| Name | Who or what |
|---|---|
| **Brobot 1** | The first Vector robot, "Doc", the host (`brobot_1`). Says metric units and 24-hour time. |
| **Brobot 2** | The second Vector robot, "Pip", the brat (`brobot_2`). Says imperial units and 12-hour time. Keeps the cube. |
| **The Brobots** | Both robots together. |
| **Brobot 3 / Reporter 1** | A voice-only reporter played on the chat page (voice `af_heart`). |
| **Brobot 4 / Reporter 2** | The second voice-only reporter (voice `am_michael`). |
| **Brobot 0** | The CrushN8r topic: the owner's own topic, found by the word "crushn8r". |
| **Goverlord** | The system. It runs the session ("GOPOD Brain"), speaks the system lines in its own voice (`bm_george`, "the digital butler"), and owns the grey system bubbles. |
| **Operator** | You, the person at the keyboard, numpad or mic. Your lines show in **Operator bubbles**, read out in the Operator voice (`af_sky`). |
| **CHALK** | A manager: the display host (CHALK BORED, numpad key KP7). Reporter 1's voice, chalkboard bubbles. See [Managers](06_CAST_AND_GAME.md#12-managers-chalk-and-playhead). |
| **PLAYHEAD** | A manager: the play, time and place host (PLAYHEAD ZUMINTU, numpad key KP8). Reporter 2's voice, tape-brown bubbles. |
| **Gomad** | A manager persona (CHALK or PLAYHEAD) as a chat participant. |
| **Gominion** | An output bubble type named with Goverlord and Gomad. This manual has no further rules for it yet. |

Bubble types: **Operator** is input; **Goverlord**, **Gomad** and **Gominion** are output.

### 1.2 The three words

| Word | What it is | In a song folder |
|---|---|---|
| **Golden note** | The smallest move or line the robots can play, proven live on its own first. | One step's `note` type in `knobs.json` (`arm_cue`, `rattle`, `say_turn`, ...). |
| **Sequence** | A few notes in order that make one moment. | One `## SECTION` in `story.md`, the steps' `section` in `knobs.json`. |
| **Golden sequence** | A sequence, or a whole song, that's locked: it plays the same way every time and is ready to record. | A song whose status line says `GOLDEN`, or a locked opener (the Bingo test's `OPENING SYNC`). |

The **piano** is the notes. The **mixer** is how each note plays (its hold, speed, buffer,
tempo). The **sequencer** is their order. All three live in the song folder: no song's moves are
hidden in code.

### 1.3 On stage: the chat page

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
| **Egg bubble** | A found easter egg, ringed by marquee bulbs. | [Eggs](06_CAST_AND_GAME.md#13-eggs) |
| **Chapter** | Goverlord's title between parts of the interview. | the interview |
| **Warm-up bubble** | "Goverlord: GOPOD Brain", the LEDs and hourglass, then the ready message. | `llm_warmup` |
| **Countdown** | The ⏳ bubble with juggled balls, under a rattle. | note `rattle` |
| **Big ball** | The Bingo ball bubble. | `show_ball` |
| **Highlight** | The lit piece in a bubble: the words being said or sung now. | [9.5](04_CHAT.md#95-highlight-and-word-reveal) |
| **The baton** | The next bubble waits hidden until the one on screen finishes. | [9.3](04_CHAT.md#93-the-chat-rules) |
| **Operator line** | `Brobots 1 2 All [Type to the Brobot, press Enter] \| Record? [] \| HH:MM:SS 🔇` under the chat. | the chat |
| **System menu** | The menu that opens from the speaker icon (🔇/🔊), listing every system command in tour order. | [14](07_SYSTEM.md#14-the-system-menu-and-commands) |

### 1.4 The show

| Say | It means |
|---|---|
| **Session** | `wpr 2` until the next `wpr 2`. Goverlord runs it. |
| **Run** | One song, draft or phcal live, from its start to its check-out. One at a time. |
| **`start <song>`** | Starts a run: `start bingo test`, `start baby brobots`, `start interview`, `start fart joke`. |
| **`stop run`** | Stops the run at its next safe spot ("Run stopped."). |
| **Thanks** | The thanks ending: closes the run out ("thanks brobots", "thanks guys"). |
| **Sleep call / wake call** | Both Brobots put to sleep (`sleep_both`) / woken (`wake_both`, or the wake inside `brobots_ready_together`). |
| **Cube call** | The cube blip on a Brobot's own cube: red, 2 s, green. Brobot 2 keeps the cube. |
| **Brobots Ready** | Both Brobots say the line together (`brobots_ready_together`). |
| **Last words** | The song's last sung or spoken words; the ending is timed from them. |
| **Edit mark** | A 0-second `pause`: a place for the reporters' voiceover in the video edit. |
| **Live / dry** | With the robots / with no robots (nothing moves, everything is logged). |
| **Check run / check logs** | Read the last run's log and Wire-Pod's journal; report the bottom line. |
| **Bench test** | One move or tool tried on its own, on the bench (phcal or a one-off script). |
| **Egg** | An easter egg: a prize (pictures or a card) that pops up when a line's words find it. |
| **`_llm_song`** | The fixed-name folder inside every warm-up song that holds the lines said between "Voices ready" and "LLM ready". |

### 1.5 The workbench

| Say | It means |
|---|---|
| **Golden** | Proven live and locked. Changed only on the owner's go. |
| **WIP** | Being worked on. |
| **Lock it in** | Mark it golden and commit it. |
| **Commit / push** | Save to the repo / send it to GitHub. Two separate go's. |
| **`wpr 2`, F5** | Restart Wire-Pod (picks up a deploy), reload the chat page. |
| **Desk start** | `source ~/.bashrc; gorepo; lsusb; wpr 2`, then F5 (below). |
| **`pha0b`** | Plays a range of a song's steps, live or dry. |
| **`score`** | Prints a song's steps, on screen. |
| **phcal** | The bench: tune one move or sound on one robot. |

### 1.6 Desk start

Your terminal routine: `source ~/.bashrc; gorepo; lsusb; wpr 2`, then F5.

- `source` reads only the first file it's given (a second becomes an argument). `~/.bashrc`
  already loads `~/.bash_aliases`, so `source ~/.bashrc` alone does both.
- `gorepo`: `cd ~/crushn8r_git/GOPOD`.
- `lsusb` - the rig's parts, by name (bus/device numbers drift, names don't): **TTGK
  Technology USB Audio** (3302:12e9, the USB mic), **China Resource Semico ... USB Keyboard**
  (1a2c:4324, the numpad), **Novatel Wireless M3000** (the mobile hotspot), Realtek Bluetooth
  Radio and 802.11ac NIC (Wi-Fi), the VIA / Huasheng USB hubs. Mic or numpad missing here = the
  chat's mic or numpad won't work.
- `wpr 2`: sets the audio routing (`AUDIO_ROUTING_SET speakers=Built-in Audio Analog Stereo
  mic=USB Audio Mono`), then restarts Wire-Pod (asks for sudo; `WPR_RESULT status=FORCED`,
  "wire-pod restarted <time>"). A deploy made after that time needs another `wpr 2`.
- F5 on the chat page: the page clears the chat for the new start and reloads once by itself.
- Then click **Log** on Wire-Pod's web page. Nothing plays until you do (a browser plays sound
  only after a click).

---

---

[Back to ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md)

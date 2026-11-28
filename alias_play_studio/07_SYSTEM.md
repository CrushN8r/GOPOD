# The system menu, your stuff, the settings

Part of [ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md) - [Concepts](concepts/README.md) · [Songs](02_SONGS.md) · [The session and runs](03_SESSION.md) · [The Brobots chat](04_CHAT.md) · [Voices and sound](05_VOICES_AND_SOUND.md) · [The cast and the game](06_CAST_AND_GAME.md) · [The system menu, your stuff, the settings](07_SYSTEM.md) · [Golden rules, troubleshooting, tests](08_RULES_AND_TESTS.md) · [The gears](09_GEARS.md)

<details><summary>Game card: the GOPOD system</summary>
<!-- gopod topic: 05_the_system -->

# the GOPOD system
- **chapter:** Show and tell

- **words:** gopod system, the system, system menu, system commands, system command
- **ask:** Tell me about the GOPOD system menu
- **summary:** system commands, typed in the chat | the system menu, a tour guide | every write goes to system-undo
- **source:** alias_play_studio/07_SYSTEM.md
- **ask 1:** Brobot 2, tell me about the GOPOD system menu
- **ask 2:** Brobot 1, how does the system menu lead you, like a tour guide?
- **ask 3:** Brobot 2, what happens to every write in the GOPOD system - is there an undo?

---

The GOPOD system. Reveal these, in order, in your own words:
1. System commands are typed on the chat page's keyboard, or picked in the system menu - a fixed list of edits and reads, never a terminal command.
2. The speaker icon opens the system menu; it leads you like a tour guide, one group at a time.
3. Every write goes to system-undo - an edit Wire-Pod wouldn't start with is put back at once.

</details>

## The system commands
- **Every command, its group and its tour:** the Scrolls, [GOPOD_SYSTEM_SCROLLS.md](../GOPOD_SYSTEM_SCROLLS.md)
  and its folders in [`system_scrolls/`](../system_scrolls/). GOPOD writes them from the system
  menu card, [`system-menu.md`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/company/99_system/system-menu.md) (`system-scrolls`); `system-check` says when
  they're behind.
- In the chat: `system-help` lists them all, `system-tour` reads the Scrolls aloud, and the speaker
  icon opens the system menu.
- **Dial 0 for the system:** a `system-` command typed to anyone but 0 doesn't run.
- **A fixed list of edits and reads, never a terminal command.** Every write goes to
  `system-undo`; a settings edit Wire-Pod wouldn't start with is put back at once.
- **Every reply's words** are templates in `system_text`, `{names}` filled by the code.

## The rules the system keeps
- **The settings folder:** `webroot/gopod_settings/`, a short file per area, each block in exactly
  one file; Wire-Pod reads them as one and serves them to the page. `system-locate <word>` finds a
  setting's `file:line`; `system-set` edits one in place.
- **Cards fill the settings:** a card section that names a settings place in brackets is that
  place's one source.
- **Global knobs, then the song, then the step:** `goverlord/runtime/songs/global_knobs.json`,
  a song's `knobs.json`, a step's own value.
- **Show timings live in the settings** (`timings`); the page keeps no value of its own.
- **Look tokens:** every colour is a token in the `:root` block of `css/gopod_chat_bubbles.css`;
  change a colour once, there.
- **Every robot call goes through `gopodRobotCall`:** one call at a time per Brobot, control
  taken, settled, run, given back.
- **The Edge TPU:** its models in `goverlord/runtime/models/edge_tpu/`; turning it on after a
  replug: [models/README.md](../goverlord/runtime/models/README.md).

## Your stuff, no code

Everything below is data - no code is touched. After any of it, type **`system-check`** on the
chat keyboard: it looks for every file the settings, chapters and songs name, and every song step
(Wire-Pod also runs it at start - `GOPOD_CHECK` lines in the journal).

| To... | Do this | Then |
|---|---|---|
| **Change a setting** (volume, waits, texts) | `system-set <setting> <value>` on the chat keyboard, or edit `webroot/gopod_settings/` (`system-locate <word>` finds it) | page: F5; Wire-Pod: `wpr 2` |
| **Add a picture to an egg** | put it in `webroot/pictures/`, add its name to the topic file's `- **pictures:**` line (`gopod_probes/topic_clusters/personas/company/4_go_for_it/<chapter>/<topic>.md`) | `wpr 2` |
| **Add a topic** (new egg words, value points) | a new `NN_name.md` topic-cluster file in its chapter folder, from the template (`gopod_probes/topic_clusters/_topic_cluster/`): `# label`, then `- **words:**`, `- **pictures:**` (optional), `- **reminder:**`, `- **say 1:**` ..., then `---` and its reveal text | `wpr 2` |
| **Add a chapter** (a new egg) | a new folder `NN_name/` in `gopod_probes/topic_clusters/personas/company/4_go_for_it/` with `chapter.md` (`# title` = the egg's title, `- **summary:**`) and its topic files; the number sets its order | `wpr 2` |
| **Change a song line** | its `## STEP` / `> TEXT:` in the song's `story.md`, or `system-set <song> <step> text <words>` | next run |
| **Add a song step** | a `## STEP <id>` + `> TEXT:` in `story.md` **and** a step with the same `step_id` in `knobs.json` (`system-check` says if one is missing) | next run |
| **Add a sound** (sting, song sound, soundtrack) | put the `.wav` in `webroot/gopod_reporter_audio/`; name it in the settings (a sting) or a song's `knobs.json` | evened to the loudness rule by itself |
| **Change how a word is said** | the pronunciation registry in `brobots_wirepod_interview_section_card_template_1_001.md` (Brobots), `reporter_voices.say_words` (reporters), `system_say_words` (Goverlord), or `system-lingo` | `wpr 2` |
| **Add a QR code** | the picture in `webroot/qr/`, an entry in `qr_codes.codes` (words that show it) | F5 |
| **Video** | not supported yet in the chat | - |

[Back to ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md)

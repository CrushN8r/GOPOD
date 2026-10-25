# LLM Warming

**Status:** WIP (2026-10-16)

The opener: Wire-Pod plays it at every start (`wpr 2`), from the LLM warm template
(`../_new_song/_llm_warm/`). Edit a line's words here and restart Wire-Pod; the voice file is made
for the new words.

- **The lines** live here (`## STEP` / `> TEXT:`), **the steps** in [knobs.json](knobs.json).
- `(music)` in a line is shown for deaf viewers and never said.
- `llm_song_0104` plays the `_llm_song` folder's run (the lines between "Voices ready" and "LLM ready"); swap that folder to change them.
- `"volume"` on a step in knobs.json: `"default"` (chat_audio.volume) or a percentage of it
  (`"50%"`).

---

## SECTION: WARM UP

Note (operator 2026-10-18: "if I restart and don't put focus the wire-pod page (active tab) * this
is like, the equivalent of where the off-focus wait messages would happen"): yes - before each
step here, the opener waits for the chat page to be here (open, visible, its Log clicked). While
it's away, Goverlord's warm-up bubble says "Goverlord awaits your return..." with the hourglass
(`awaits_text`, gopod_brobots_settings.json) and nothing plays; on return the steps go on in order.

## STEP goverlord_warm_0101

> TEXT: Voice files generation check.

## STEP goverlord_warm_0102

> TEXT: Voices ready for playback.
Language Model Loading...

## STEP goverlord_warm_wait_0103

> TEXT:

## STEP llm_song_0104

> TEXT:

---

## SECTION: READY

## STEP goverlord_warm_ready_0105

> TEXT: CrushN8r is Ready. (music)

## STEP goverlord_sting_0106

> TEXT:

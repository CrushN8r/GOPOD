# LLM Warm Template

**Status:** WIP (2026-10-16)

The master opener LLM warm template. Copy this folder to `<number>_<name>/` and give its
knobs.json its own `song_id`. The standard is
[ALIAS-PLAY-STUDIO.md](../../../../../ALIAS-PLAY-STUDIO.md).

- **The lines** live here (`## STEP` / `> TEXT:`), **the steps** in [knobs.json](knobs.json).
  Both files name the same steps.
- `(music)` in a line is shown for deaf viewers and never said.
- `reporter_lines_0104`: `opening` plays the reporters' opening; an empty `> TEXT:` skips it.
- `"volume"` on a step in knobs.json: `"default"` (chat_audio.volume) or a percentage of it
  (`"50%"`).
- A step with no words keeps an empty `> TEXT:`.

---

## SECTION: WARM UP

## STEP goverlord_warm_0101

> TEXT: Voice files generation check.

## STEP goverlord_warm_0102

> TEXT: Voice files generated. Language Model Loading

## STEP goverlord_warm_wait_0103

> TEXT:

## STEP reporter_lines_0104

> TEXT: opening

---

## SECTION: READY

## STEP goverlord_warm_ready_0105

> TEXT: Language Model Loaded, Master CrushN8r. Your move. (music)

## STEP goverlord_sting_0106

> TEXT:

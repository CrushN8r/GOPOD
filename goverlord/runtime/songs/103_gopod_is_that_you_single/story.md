# GOPOD Is-That-You — live capture, not scripted (single-robot, KP1 only)

Knobs: [knobs.json](knobs.json)

**Split 2026-08-18** from the prior single `103_gopod_is_that_you/` folder into two
sibling versions — this one (`_single`) and
[`104_gopod_is_that_you_multi/`](../104_gopod_is_that_you_multi/story.md) (the golden,
unscoped original) — the same two-version pattern `101_brobots_bingo_test`/
`102_brobots_bingo_game` already use on the shelf. Operator direction, verbatim intent:
"take the gold from `_multi` and apply scoped to `_single`: like only a KP1 is that you
test." Everything below is the `_multi` gold, scoped down to a single robot; nothing about
the underlying mechanism changed to make that scoping possible — see "What's actually
scoped" below.

**Reshaped 2026-08-08** (pre-split history, same for both siblings). Same lineage as
`102_brobots_cross_persona/` (the "persona mixup") — both trace back to
`goverlord/runtime/gopod_layer/web_display/gopod_demo_8011/gopod_ptt_chat_writer_013.py`
(the live `is-that-you` PTT demo, aliased at `~/.gopod_alias_lib/demo.sh:62`,
`is-that-you()`) — but they answer the writer's interactivity in opposite directions.
Operator direction, verbatim intent (2026-08-08): this song is "intended to be a video
recording of a live test, not scripted. That's the diff with this song compared to the
persona mixup." `102_brobots_cross_persona`'s own `story.md` is the base scaffolding this
one borrows its shape from.

## What's actually scoped — and what isn't

The live `is-that-you()` writer (`gopod_ptt_chat_writer_013.py`) already handles KP1
(Doc) and KP2 (Pip) as fully independent, self-contained key handlers —
`handle_captured_key_press()`/`persona_awareness_reply()` never couple the two robots
together in code. Holding only one key already runs single-robot today, with zero
changes to the writer itself. So the "single" scope here isn't a code change at all — it's
a scope on the **live capture step**: press and hold only KP1 during the recording,
rather than alternating KP1/KP2. Everything else about the mechanism — real mic, real
Vosk transcription, the real canned-line-vs-live-Ollama decision — is exactly what
`104_gopod_is_that_you_multi/` describes, unchanged.

## The difference from the persona mixup

`102_brobots_cross_persona` scripts the one piece of the writer that IS fixed content —
`persona_awareness_reply()`'s 4 canned cross-persona lines — and explicitly leaves everything
interactive out (see that song's own LEFT OUT section: the keypress wait, live mic capture,
Vosk transcription, live Ollama fallback — none of it fits a fixed step list).

This song does the opposite: it doesn't script anything at all. There is no dialogue to author
here on purpose — the point of a video capture of this song is to show the writer actually
running live: a real KP1 press, real mic audio, a real decision between a canned line and a
live Ollama reply, exactly as `is-that-you()` behaves today when only one key is held.
Scripting fake dialogue over that would misrepresent it as pre-written when the whole
reason to shoot it is that it isn't.

## How this song is different from the scripted songs

Beyond tone, three concrete mechanical differences from every scripted song on the shelf
(bingo, the persona mixup, nap, awaken) — not just "less dialogue," but different plumbing:

1. **No reporter gaps.** Scripted songs place timed `reporter_gap_*`/`pause` steps that the
   engine inserts automatically, at a fixed point the author chose ahead of time. This song has
   none, because there's no fixed moment to gap around — what happens during `live_capture` is
   a live human at a keyboard, not a pre-known sequence of beats. The reporter beats for this
   song get added afterward instead, in post-production, as breaking-news freeze-frame VFX:
   freeze the live footage at a chosen moment, let the reporters (Brobots 3/4) react to what
   actually just happened in that take, unfreeze, and let the recording keep running until the
   next freeze. Those reactions are written after watching the footage, responding to what
   really happened — the reverse order from a scripted song's reporter gap, where the pause and
   its content both exist before the take is ever run.
2. **No tempo knobs either.** The tempo buffer knob (`global_tempo`/`tempo_factor`, see
   `TEMPO_BUFFER_KNOB_SURVEY_001.md` in `gopod_notes/`) spaces out engine-driven beats in a
   scripted song — it scales the gap between one authored line and the next. This song has no
   authored beats to space; its "content" is a live human's own pacing at the keyboard, which no
   tempo value could scale or replace even if one were set here.
3. **Live-recorded first, everything else follows.** A scripted song is authored, then run. This
   song runs backwards: record several live `is-that-you` (KP1-only) sessions first, keep the
   best take, and only then does anything else happen — choosing freeze points, writing the
   reporter reactions, editing. The knobs.json steps here don't carry any of that content;
   their only job is to bracket the live capture cleanly (`wake_both` → `live_capture` →
   `exit`), the same bookend role every other song's own open/close steps play.

Not yet live-run: this song is dry-verified through `run_golden_song_001.py` directly and
through `pha0b itsyou-single`, but no actual `is-that-you` (KP1-only) recording session has
been shot through it yet.

## Why this is still a "song" at all

Two bookend steps only, same note vocabulary as the persona mixup, so this folder is
`pha0b`/`phcal`-shape-compatible even though nothing plays between the bookends:

- **`wake_both`** — connective tissue, unchanged from `_multi`. It pings BOTH robots'
  connectivity, not just Doc's — that's deliberate, not an oversight: this note is a
  readiness check (`conn_test` on both serials), not a scope decision, and both robots are
  still physically present on camera even though only Doc/KP1 participates in the actual
  capture. No single-robot equivalent of this note exists in the golden engine's own
  vocabulary today (checked `run_golden_song_001.py`'s note-dispatch code directly before
  writing this — the closest alternative, the `connect` note, is a genuinely different
  mechanic used by the control-family songs: it assumes and holds behavior control for
  the whole run, which this bookend has never done and shouldn't start doing here).
  Not writer content — the writer itself has no wake step of its own, it assumes/releases
  control around every single reply instead of holding continuously.
- **`live_capture`** (`pause`, `pause_seconds: 0`) — a bookmark, not a real wait, same
  role as `_multi`'s own. Its `gap_label`/`section` are renamed
  (`live_is_that_you_capture_kp1_only` / "LIVE CAPTURE - NOT SCRIPTED - KP1 ONLY") to say
  plainly what's scoped, but the mechanism is identical: the runner passes through this
  step in zero seconds; the real live capture happens between running this song's
  `wake_both` and its `exit`, not inside any note this file defines.
- **`exit`** — same connective-tissue role as the persona mixup's own `exit`: marks the
  session ending, nothing more.

## Recording this for real

1. Run this song's `wake_both` (pings both robots' connectivity, gets them ready on
   camera).
2. Start recording, then run the real `is-that-you` alias
   (`~/.gopod_alias_lib/demo.sh`, `is-that-you()`) for the actual live, unscripted test —
   **hold only KP1** for this take (Doc only), real mic, real Vosk/Ollama decision loop,
   exactly as `is-that-you()` behaves standalone today when only one key is pressed.
   This is the content; nothing here scripts it.
3. Run this song's `exit` to close out.

No new mechanism was built for step 2 — it's the existing `is-that-you` alias, unchanged,
run directly. This song's only job is to bookend it consistently with every other song on
the shelf, scoped by which key the operator holds during step 2.

## STEP wake_both
> TEXT:

## STEP live_capture
> TEXT:

## STEP exit
> TEXT:

## Wired into pha0b — 2026-08-18

Registered in `run_golden_song_001.py`'s `SONG_REGISTRY` under `gopod_is_that_you_single`
(same bingo-family shape as `brobots_cross_persona`) and given its own `pha0b`/
`pha0b_menu()` keyword, `itsyou-single` — split from the old singular `itsyou` keyword
the same day this folder split from `103_gopod_is_that_you/`. Deliberately not
`is-that-you`/`isthatyou`, same non-collision reason the original keyword avoided that
name: the live `is-that-you()` PTT alias this song brackets already owns it. Dry-verified
clean both directly (`pha0b itsyou-single wake_both exit`) and via `pha0b_menu()`'s
disk-scan picker. Not yet live-run.

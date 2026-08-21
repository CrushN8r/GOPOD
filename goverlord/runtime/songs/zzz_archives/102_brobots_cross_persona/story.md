# Brobots Cross-Persona — "is that you?" demo reel

**Naming note, exception:** this song runs on brobots wire-pod layer hardware, but its own
content is deliberately framed as "leaked GOPOD-layer test footage" (per `niche-buzz`'s own
funnel role for this song, UPSELL 2). Because the song is *testing/previewing the GOPOD
layer* by design, "Doc"/"Pip" is the correct, in-character naming throughout this file —
including the descriptive prose, not just the spoken lines — unlike other wire-pod-layer
songs, which use Brobot 1/Brobot 2. Confirmed by the operator, 2026-08-12.

Knobs: [knobs.json](knobs.json)

**Derived 2026-07-25** from
`goverlord/runtime/gopod_layer/web_display/gopod_demo_8011/gopod_ptt_chat_writer_013.py` (the live
`is-that-you` PTT demo, aliased at `~/.gopod_alias_lib/demo.sh:60`), per the operator's own
scope instruction, verbatim: "gopod_ptt_chat_writer_013.py > song, as best you can. what may
not apply easy, leave out." The writer itself is untouched — this song is a new, separate
artifact, not a refactor of it.

**Renamed 2026-07-31** from `gopod_is_that_you` to `brobots_cross_persona` — the old name
collided with the live `is-that-you` alias/PTT demo it was derived from (two different
things sharing one name). The live PTT demo keeps the `is-that-you` name; this song, being
a separate scripted artifact, moved to a non-colliding name. Content, steps, and timing
below are unchanged from the original — only the song_id/folder name and this section
changed.

The writer is fundamentally interactive, not scripted (operator's own framing): it blocks on
a live physical KP1/KP2 keypress, records real mic audio, runs it through Vosk, and only THEN
decides what a robot says — either an instant canned "is that you?" line (if the operator's
own words name Doc or Pip while a key is held) or a live Ollama-generated reply. None of that
decision loop, the mic capture, or the live LLM call can become a fixed step list. See LEFT
OUT below for the complete, named account of what got dropped and why — nothing was
approximated or half-ported.

The one piece of the writer that IS fixed, author-able content is
`persona_awareness_reply()`'s 4 canned cross-persona lines — verbatim, unedited. This song
plays all 4 back to back, as a demo reel — not as how the real system ever actually chains
them. In the live writer exactly ONE of these fires per operator utterance, decided live; Doc
and Pip never actually address each other. Doc's own two lines run first (self-confirm, then
the "wrong robot" line), then Pip's own two lines — mirroring `PERSONA_AWARENESS_REPLIES`'s
own dict order in the source file, not an invented conversation.

Built on `run_songs_runner_001.py`'s note vocabulary (`wake_both`/`say_turn`/`pause`/
`exit`), not `run_robot_control_song_001.py`'s — chosen because `say_turn` resolves a real
per-step robot serial from each step's own `speaker` (`brobot_1`/`brobot_2`), letting Doc's
and Pip's genuinely different lines share one script. The control-song runner only supports
one robot per whole invocation against a shared `{robot_name}` template, which can't carry two
personas with different dialogue at all.

`wake_both` here is connective tissue, not writer content — the writer itself has no
"connect"/"wake" step of its own (it assumes/releases behavior control around every single
reply instead of holding continuously); see LEFT OUT below.

`> TEXT:` is spoken verbatim for a `say_turn` note. `> FAIL:` is spoken only if that note's own
hardware call didn't come back clean (unused here — no `arm_cue`/`nod`/`weather` steps in this
song). The `pause` note is a silent, deterministic sleep (no hardware call), `pause_seconds: 0`
per the studio's own standing reporter-gap rule (`.claude/skills/alias-mixer/SKILL.md` §2) —
marking the natural edit point between Doc's block and Pip's, left open for a later edited-in
transition, never a live dead-air pause.

## STEP wake_both
> TEXT:

## STEP doc_self_id
> TEXT: Doc here. Yes. Try to keep up.

## STEP doc_wrong_robot
> TEXT: That is Pip's lane. Wrong robot, right confusion.

## STEP reporter_gap_mid
> TEXT:

## STEP pip_self_id
> TEXT: Yep. Pip here. I think that was my cue.

## STEP pip_wrong_robot
> TEXT: Hey Doc, is this when I ask for emails?

## STEP exit
> TEXT:

## LEFT OUT — named, not approximated

Per the operator's own instruction, everything below was left out rather than half-ported or
invented:

- **CLI diagnostic flags** (`--list-devices`/`--resolve-device`/`--dry-validate`/`--stdin`) —
  operator tooling, no song-content equivalent.
- **Startup hardware bootstrap** (session dir + log tee, Vosk model load, mic device
  resolution, NumLock LED read/sync, exclusive `EVIOCGRAB`) — live PTT-rig hardware prep; no
  song beat is analogous.
- **The main event loop waiting on a physical keypress** — a song is a fixed step list; there
  is no "wait for a human" note type.
- **NumLock toggle gate** — a live software gate/state flip, not narrative content.
- **KP1/KP2 mic capture** (recorder thread, release-tail wait, resample, archive WAV write,
  signal-stats) — live audio hardware capture of unscripted human speech; there is no fixed
  transcript to author into a step.
- **Held-too-short discard** (`held_seconds < min_hold_seconds`) — a live timing branch on
  human behavior.
- **`append_operator_message`** (writes the Operator chat envelope) — writes the just-captured
  live transcript; nothing fixed exists to write.
- **`call_ollama_self_id`** (the live LLM self-ID fallback for whenever neither name is
  mentioned) — a live network call producing non-deterministic text; no fixed line exists to
  author.
- **`sanitize_for_robot_speech`** — a text-cleanup utility applied to whatever reply resulted;
  not content itself, and every existing song's own speech pipeline already normalizes
  outbound text at speak time.
- **The writer's own per-turn `assume_behavior_control` → `say_text` →
  `release_behavior_control` handshake with a one-time retry** (`dispatch_robot_speech`) — the
  underlying "speak this line" action is what `say_turn` already does; the writer's specific
  retry-once wrapper, and its assume/release-around-every-single-reply pattern (never held
  continuously, unlike this song or `brobots_awaken`), is a real, named difference, not
  reproduced here.
- **The KP0 triple-tap exit gesture and its 2-second window** — the physical tap-timing
  gesture itself has no equivalent; the fact a session ends is what this song's own `exit`
  step represents instead.
- **Cleanup** (`drain_fd`, releasing `EVIOCGRAB`, closing the fd) — OS/hardware teardown, no
  song equivalent.

## Wired into pha0b — 2026-07-31

Case-statement arms added: `pha0b_menu()`'s dir→song-keyword map now has
`brobots_cross_persona) song="mixup" ;;`, and `pha0b()`'s song-keyword→runner map has a
`mixup` arm (`run_songs_runner_001.py`, `env_prefix="GOPOD_BINGO_PLAYHEAD"` — matching
bingo's, since the runner's own env var names are fixed regardless of which song points at
it — `song_dir_export` pointed at this folder). Since this song isn't the runner's default
directory (bingo is), pha0b()'s shared dispatch block also picked up one new conditional
export line: `GOPOD_BINGO_CAPTURE_SONG_DIR` gets set from `song_dir_export` whenever the
chosen runner is `run_songs_runner_001.py` — a no-op for bingo (which never sets
`song_dir_export`) and a no-op for the `run_robot_control_song_001.py` songs (condition
excludes that runner). Dry-verified only — resolves to a real runner
(`run_songs_runner_001.py`) and a real directory (this one) — never fired live.

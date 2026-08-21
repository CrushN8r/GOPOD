# Bingo Reactor — Brobot 2 Angry Animation Watcher

## What this is

A small, standalone Python watcher (`bingo_reactor_001.py`) that tails the shared
`live_chat_messages.json` chat file for `bingo_draw` envelopes and fires an angry animation
(`anim_rtpickup_loop_10`) on Brobot 2 (`0dd1d8bf`) each time a number is drawn. It stops itself
shortly after seeing a `bingo_end` envelope.

## What this does NOT do

- **Never touches Bingo.** It only reads the chat file Bingo already writes to — no import, no
  call site, no shared process, no modification to the sidecar at
  `goverlord/runtime/songs/102_brobots_bingo_game/`.
- **Never touches Brobot 1.** Only ever connects to Brobot 2's serial (`0dd1d8bf` by default).
  Bingo itself continues to run entirely on Brobot 1, untouched.
- **Never blocks or crashes anything it doesn't own.** Every SDK connect, animation dispatch, and
  chat-file read is best-effort: failures are logged (`GOPOD_BINGO_REACTOR_CONNECT_FAIL`,
  `GOPOD_BINGO_REACTOR_ANIM_FAIL`, `GOPOD_BINGO_REACTOR_PARSE_FAIL`) and the watcher keeps polling.
  If this process isn't running at all, Bingo works exactly as it does today — just without a
  Brobot 2 reaction.

## How to run

Normal use: nothing to do — as of 2026-07-06, `gobingo` (in `~/.gopod_alias_lib/demo.sh`)
launches this watcher itself as a background job before starting Bingo, so one command in one
terminal drives both robots, with both processes' log lines interleaved in the same output. This
is still two independent OS processes started from the same shell wrapper — no import, no call
site, no shared process with Bingo itself.

Manual/standalone use (e.g. debugging the reactor on its own):

```
gobingo-reactor
```

Run it in its own terminal alongside `gobingo`. Two independent launches — starting Bingo does
not start the reactor, and vice versa, whether launched together by `gobingo` or run by hand.

## How to stop

When launched automatically by `gobingo`, it's killed for you when Bingo exits (normal
completion, error, or Ctrl-C on the `gobingo` terminal). When run standalone via
`gobingo-reactor`, `Ctrl-C` in its own terminal stops it. Either way it also stops on its own ~2
seconds after Bingo emits `bingo_end`, or if the chat file it's watching disappears.

## Brobot 2 serial / IP

Default serial: `0dd1d8bf` (Vector-T3P1). Override with `--brobot2-serial <serial>`. IP is
resolved via the `anki_vector` SDK's own `~/.anki_vector/sdk_config.ini`, keyed by serial — not
hardcoded here.

## Requirements

Needs the `anki_vector` package (the wire-pod fork — kercre123,
[github.com/kercre123/wirepod-vector-python-sdk](https://github.com/kercre123/wirepod-vector-python-sdk),
Apache-2.0, original Copyright (c) 2018 Anki Inc. preserved per license terms; see
`THIRD_PARTY_LICENSES.md` at the repo root — was vendored at
`goverlord/SDK/sources/wirepod-vector-python-sdk/` (replaced 2026-07-30 with a placeholder
note; the original is preserved in a private, byte-verified backup outside this repo)
— moved there from `~/crushn8r_git/SDK/sources/wirepod-vector-python-sdk/` during the 2026-07-05 SDK/Models
Gitignore Cleanup, which also broke this venv's editable install until it was reinstalled from
the new path on 2026-07-06), installed as an editable package into this directory's own `.venv/`
(not the system Python — `anki_vector` is not installed system-wide). Both the `gobingo-reactor`
alias and `gobingo`'s auto-launch use this venv's interpreter directly.

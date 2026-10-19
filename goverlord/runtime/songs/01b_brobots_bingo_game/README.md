# Brobots Bingo game (work in progress)

Chocolate Bingo, hosted live by Brobot 1. Not the scripted Bingo test song (that's
`01a_brobots_bingo_test/`).

- **Start it:** run `gobingo` (Take Out part 1). It runs the game silent.
- **How it plays:** the deck is shuffled and Brobot 1 says he's ready. The first touch draws a
  number and starts drawing one every 3 seconds. The backpack button ends early and reveals the
  rest of the deck.
- **Brobot 2 reacts:** `bingo_reactor/` watches the draws and plays an angry animation on
  Brobot 2. `gobingo` starts and stops it alongside the game.
- **What's here:** `cmd/main.go`, `pkg/intents/voicecommand_lottery.go` and
  `pkg/intents/gopod_chat_envelope.go` (GOPOD's edits, the source of truth), and the built
  binary `bin/vectorx-gobingo`.
- **Building:** these files have no `go.mod` of their own. Copy them over a
  [vectorx](https://github.com/fforchino/vectorx) clone, run `go build -o <out> cmd/main.go`
  there, and copy the binary back to `bin/`.
- **Known issue:** an Error 915 display/audio conflict in vectorx's own Bingo code. It doesn't
  show while the game runs silent.

Based on [vectorx](https://github.com/fforchino/vectorx) by Filippo Forchino, MIT licensed:
see `LICENSE` here and [THIRD_PARTY_LICENSES.md](../../../../THIRD_PARTY_LICENSES.md).

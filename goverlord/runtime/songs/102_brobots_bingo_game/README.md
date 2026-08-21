# Bingo Sidecar

**STATUS — 2026-08-10.** Stale, not broken. No game-logic edit since ~2026-07-06 (this
binary's build date) — every touch since has been doc/path housekeeping. Still live-wired
today: the `BROBOTS_BINGO` Wire-Pod custom intent ("go bingo") execs this binary, and
`gobingo` still auto-launches the Brobot 2 reactor alongside it. One known, pre-existing,
non-blocking issue: an Error 915 display/audio conflict in vectorx's own bingo code
(investigated, never fixed) — doesn't currently surface since both live paths run
`--silent`. Absorbed against this session's golden Python song-engine mechanics
(stay-put, tempo, playback filter, golden-flag wake) — none apply as real gaps: this
sidecar's own `BehaviorControl` grant already blocks on the real `ControlGrantedResponse`
(no race the golden-flag fix would need to repair), and Bingo does zero wheel/arm
movement, so there's no motor-race surface either way. See
`GOBINGO_GOLDEN_ABSORB_AND_EVAL_001.md` and `GOBINGO_REVIVAL_DIRECTION_001.md` in
`gopod_notes/` for the full absorption pass and the operator's own direction to bring this
game back as the actual game (distinct from `101_brobots_bingo_test`'s scripted demo-video
capture song).

## What this is

Vector-hosted spoken Bingo (single robot, Brobot 1 / `0dd1b9e9`), via the `vectorx-gobingo`
binary. This started (Stage Bingo Sidecar) as a read-only snapshot of the sibling vectorx repo's
Bingo code, but **as of Stage Bingo V1 this is no longer a snapshot** — the three Bingo-specific
Go files here are edited directly, in place, and this directory is their source of truth. Current
state on `main`: Stage Bingo V1 (`--grid-size`, `--silent`, chat envelope emission on
start/draw/end) plus Stage Bingo V2 (timed auto-draw after the first touch, full-deck-order
reveal on early exit; interval is 3 seconds between draws as of the 2026-07-06 timing pass, up
from V2's original 1 second — see "Status line" below).

**Attribution:** the sibling vectorx repo this sidecar started as a snapshot of is
[vectorx](https://github.com/fforchino/vectorx) by Filippo Forchino, MIT licensed — see
this directory's own `LICENSE` file and `THIRD_PARTY_LICENSES.md` at the repo root for the
full text.

## Source truth

**Edits to Bingo behavior happen here, in this sidecar**, not in the sibling repo. The three
Bingo-specific files —
`pkg/intents/voicecommand_lottery.go`, `cmd/main.go`, `pkg/intents/gopod_chat_envelope.go` — are
edited directly in `goverlord/runtime/songs/102_brobots_bingo_game/`. The build sandbox (a full vectorx git clone, same
module, still on the original commit these edits were snapshotted from) is **now at
`goverlord/SDK/sources/vectorx/`** — relocated there from its original path
(`~/crushn8r_git/SDK/sources/vectorx/`) during the 2026-07-05 SDK/Models Gitignore Cleanup pass,
which vendored SDK clones into GOPOD's own `SDK/` tree. That relocation moved the Go source but
left the **runtime `vectorfs/` directory behind at the old path** — `gobingo`'s
`WIREPOD_EX_*_PATH` env vars still (correctly) point there; see "Live status" below. This sandbox
remains sovereign for the rest of vectorx's ~15 other intents and its build tooling, but is
**not** where Bingo changes originate anymore.

Rebuilding the binary still requires the sandbox's Go module context (this sidecar has no
`go.mod` of its own), so the build process is an **overlay, not a fork**: the sidecar's 3 edited
files are temporarily copied into the sandbox, `go build -o <out> cmd/main.go` is run there (not
`./cmd/` — the package also has `setup.go`/`vimserver.go`/`webserver.go`, each with a conflicting
`func main()`), the resulting binary is copied back to `bin/vectorx-gobingo` here, and the
sandbox's working tree is restored to byte-identical pre-overlay state (verified via `sha256sum`
and `git status` diff, every stage). The sandbox is never a permanent home for these edits, only a
transient build sandbox.

`bin/` previously also held a rolling `.pre-v1/v2/v3.bak` backup lineage of the binary; these were
removed 2026-07-21 as superseded and no longer needed on disk (`dadb211f`), and are now gitignored
rather than tracked. `bin/` holds only the current `vectorx-gobingo` binary.

## Live status

The `gobingo` function lives in **`~/.gopod_alias_lib/demo.sh`** (moved out of `~/.bash_aliases`
during Stage Bingo V1's Alias Fix — see `STAGE_BINGO_V1_ALIAS_FIX_003.md`) and runs the binary
from this sidecar path: `goverlord/runtime/songs/102_brobots_bingo_game/bin/vectorx-gobingo`. A live Wire-Pod custom
intent, `BROBOTS_BINGO` (in `/home/goverlord/wire-pod/chipper/customIntents.json`, outside this
repo), also launches Bingo via a wrapper script,
`gopod_probes/tools/start_bingo_from_wirepod_intent_001.sh`. Both paths run silent
(`--silent`) by default. The binary's `WIREPOD_EX_TMP_PATH` / `WIREPOD_EX_DATA_PATH` /
`WIREPOD_EX_NVM_PATH` env vars still point at the sibling repo's own `vectorfs/` directory
(`~/crushn8r_git/SDK/sources/vectorx/vectorfs/`) — that runtime data/tmp directory is intentionally
not copied into this sidecar. Its `data/` subtree (audio, fonts, images — `rattle.wav` among them)
went missing during the 2026-07-05 SDK relocation described above and was restored on 2026-07-06
by copying it back from the still-intact copy at `goverlord/SDK/sources/vectorx/vectorfs/data/`;
if `gobingo` ever again errors with a `rattle: ... no such file or directory`, check that this
directory is still populated first.

**Second consumer of `rattle.wav`, added 2026-07-17:** the Bingo *capture song*
(`goverlord/runtime/songs/101_brobots_bingo_test/`, a separate, scored piece for the upsell
video — see `tech/alias_play_studio/SONG_101_BROBOTS_1_2_BINGO.md`) plays this same file through a standalone direct-SDK tool
(`~/wire-pod/chipper/gopod_probes/tools/direct_sdk_bingo_rattle_001.go`), not through this
sidecar's own code. It reads `rattle.wav` from `goverlord/SDK/sources/vectorx/vectorfs/data/audio/`
directly rather than through this sidecar's `WIREPOD_EX_*_PATH` env vars — if that file ever
moves again, both consumers need updating, not just `gobingo`'s own env vars.

## Integration status

**Chat envelopes: integrated.** `goverlord/runtime/data_gomad/configs/chat_envelope_schema.json` formalizes every
Bingo envelope shape this sidecar writes (`speaker_id=bingo`; `source_producer` values
`bingo_ready`, `bingo_draw`, `bingo_deck_reveal`, `bingo_end`, plus `bingo_start` reserved for
future use) and every sample envelope validates via `validate_envelope()`. The cockpit's chat pane
renders these live, same as any other envelope.

**Everything else: still not integrated.** The cockpit and PTT/chat writer don't read Bingo state
beyond the shared chat file. The swarm participant registry has no Bingo KP slot. No GOPOD crystal
code (Python) references Bingo at all — confirmed by zero-wiring greps every stage
(`grep -rn bingo` across cockpit `.py` and `data_gomad/config/*.json` + `loader.py` returns nothing).
Bingo is a silo that happens to write into the same chat file the rest of GOPOD reads.

## Firewall

The Layer 0 / Layer 1 firewall applies to this sidecar as if it were crystal code. Persona names
(Doc, Pip, CHALK) do not enter Layer 0 files inside this sidecar. Brobot role names (Brobot 1,
Brobot 2) do not enter Layer 1 rendering.

## Scope

Bingo itself is **Brobot 1 only**, unchanged since v1 (targets serial `0dd1b9e9`). This sidecar
has no Brobot 2 code and never will — see "Optional companion" below for how Brobot 2 gets
involved without Bingo itself becoming multi-robot.

## Optional companion: Brobot 2 angry reactor

Stage Bingo V3 added a **separate, standalone watcher process** — it lives outside this sidecar
at `goverlord/runtime/songs/102_brobots_bingo_game/bingo_reactor/`. It tails the shared chat file for `bingo_draw`
envelopes and fires an angry animation on Brobot 2 (`0dd1d8bf`) in reaction. This sidecar's own
code is completely unaware of it — no import, no call site, no shared process; it remains two
independent OS processes. Bingo runs exactly the same (single-robot, silent, timed auto-draw)
whether or not the reactor is running alongside it.

As of the 2026-07-06 timing pass, the `gobingo` shell function (in `~/.gopod_alias_lib/demo.sh`)
launches the reactor itself as a background job before starting Bingo, and kills it on exit —
purely an operator-convenience wrapper at the shell level, not a code coupling. One command, one
terminal, both robots' log lines interleaved in the same output. The standalone
`gobingo-reactor` alias still exists for manual/solo use if you want the reactor in its own
terminal. See `goverlord/runtime/songs/102_brobots_bingo_game/bingo_reactor/README.md` for the reactor's own details.

## Future integration

Any future integration requires an explicit, separate stage with its own scope, review, and
operator approval. This sidecar does **not** authorize any wiring. Prior planning material is
reference-only, not a pre-approved plan: a read-only architectural review of candidate
integration points (recommends a standalone bridge script, C7, as the smallest, most
reversible first step, if/when integration is separately approved), and a read-only
root-cause investigation into a display/audio conflict (Error 915) in vectorx's own
bingo code.

## Provenance

`bingo_sidecar_provenance_001.json` documented the **original** Stage Bingo Sidecar snapshot
only: source commit, sibling working-tree status at that snapshot time (dirty — uncommitted
`cmd/main.go`/`voicecommand_lottery.go` changes, untracked backups/binaries), and exactly what
was included/excluded from the sibling and why. It was a historical record of how this sidecar
was seeded, not a description of current state. Removed from this sidecar's root 2026-07-30,
preserved in a private, byte-verified backup outside this repo — for what's actually here
now, GOPOD's own `git log` on `goverlord/runtime/songs/102_brobots_bingo_game/` is authoritative.

## Status line

Stage Bingo V2 merged to `main` (`549092331bf78fbbad7732e11d83b262041b89d7`). Live behavior:
deck-prepared → `bingo_ready`; first touch draws and starts an auto-draw ticker (3 seconds between
numbers as of 2026-07-06, was 1 second at V2 merge); early exit (backpack button) reveals the
full deck order before ending; natural exhaustion ends without a reveal. Integration beyond chat
envelopes: deferred, no scheduled stage.

**2026-07-06 timing + operability pass** (not yet a numbered Stage/merge commit as of this
writing): auto-draw interval changed 1s → 3s (`voicecommand_lottery.go`'s ticker), giving Brobot
2's reactor room to connect and animate before the next number lands. `gobingo` now
auto-launches/cleans up the reactor in the background (see "Optional companion" above). Also
fixed as part of this pass: `vectorfs/data/` (rattle sound + other assets) had gone missing from
the runtime path after the 2026-07-05 SDK relocation, and the reactor's venv had a stale editable
install of `anki_vector` pointing at that relocation's old path — both restored/reinstalled from
the correct current locations.

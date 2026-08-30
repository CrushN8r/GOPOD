# WIRED-POD

> Wire-Pod gets one robot talking. This is what GOPOD did to the wire underneath it: a
> hardening pass and a new package bolted onto the same skeleton.

---

**A note on paths.** This document describes a live install on the author's own machine, not a
generic deployment guide. Absolute paths (`/home/goverlord/...`), the session log directory, the
Vosk library/include paths in `start.sh`, and similar hardcoded locations throughout are specific
to that machine. They're left as-is below because that's what the actual code says today — adapt
them before any of this runs anywhere else.

---

## The rule: stay thin

Native Wire-Pod, everywhere, except three touch points: **8 native files, current as of
2026-08-11** (down from 9 — `logger.go` was fully retired to native this pass, see
"Compiler-refereed overlay slim-down" below) — 5 native `.go` files that still carry a real
edit (`sdkapp/server.go`, `ttr/{kgsim, kgsim_cmds, kgsim_interrupt, weather}.go`) plus 3
non-Go files (`start.sh`, `webroot/index.html`, `.gitignore`) — the
`GOPOD_STREAM_MARKER_0`/`_1` debug-marker wrapper, and the marker-1-advance path that lets
touch/wake/keypad interrupts end a stream early. Nothing outside those three is expected to
behave differently from stock Wire-Pod. Before touching any native file under `~/wire-pod`, check
whether the change fits inside one of the three — if it doesn't, that's the layer growing past
"thin," not a normal edit.

**Rule, added 2026-08-10 after the rich-logs toggle incident below:** fitting inside one of the
three touch points is necessary but not sufficient. Before *deploying* any change to a native
touch-point file, confirm the file's actual current upstream behavior against real source
(`git show 11e7b22:<path>` in `~/wire-pod`'s own clone — that's the merge-base, `chipper/v1.5.10`
on `kercre123/wire-pod`, see "What this document is" below) if the change could alter anything
user-facing in the native UI — not just whether GOPOD's own new code works in isolation. A change
that's correctly coded can still be an unwanted deviation from native the moment it's live.

**As of 2026-08-02, "stay thin" is enforced structurally, not just by convention**, for 7 of
those 9: the live `~/wire-pod` tree holds genuine upstream code for `logger.go`, `sdkapp/
server.go`, and the `ttr/{kgsim, kgsim_cmds, kgsim_interrupt, weather}.go` files, and GOPOD's
actual edits exist only in this repo's `wire_pod_overlay/`, injected at build time. Editing the
live tree directly no longer does anything on the next overlay build — the overlay folder is the
only place left to make the change. See "Migration complete, 2026-08-02" further down for the
full mechanism and the two files (`start.sh`, plus config Go can't overlay) that remain a
deliberate, permanent exception. **Superseded in one respect as of 2026-08-11:** `logger.go` is
no longer one of these 7 at all — its GOPOD content moved to a standalone GOPOD-owned file, and
`logger.go` itself dropped out of the touch-point list entirely. See "Compiler-refereed overlay
slim-down, 2026-08-11" below.

Confirmed 2026-08-02: the touch-interrupt gap named below under "Where this stands today"
(wrapper hardcoded to veto every interrupt) was fixed by wiring the marker-1-advance machinery
that already existed for exactly this (`kgsim_marker_advance_control.go`'s `ShouldAdvance()` /
`MarkerOneEvents()` — built, unit-tested, zero production callers until this fix) into the one
native touch point `kgsim_interrupt.go` already exposes (`kbTouchInterruptWrapper`). No new
mechanism invented — the thin wrapper now calls the thin advance path that was always there.

---

## Open wire

Two engines already run: the song-runner / event bus (Bingo, two brobots, offline) and
the live loop (push-to-talk → speech → LLM → robot talks back). The missing piece is the
wire between them — one connected session.

**60-day plate:** a person pushes a key, says something, Doc or Pip answers live.

Smaller first cuts, one Vector, no Jetson: see
[SINGLE_BOT_QUICKSTART.md](SINGLE_BOT_QUICKSTART.md) — a hardcoded robot address that
needs to move into config, and a broken spoken wake trigger.

Keyboard-only taste, no commitment: [MY_NICHE_BUZZ_ASK.md](../MY_NICHE_BUZZ_ASK.md).
Front of house, ops/social/web: [MY_GOPOD_OPS_ASK.md](../MY_GOPOD_OPS_ASK.md).

---

## What this document is

`GOPOD_FEATURES.md` tells the product story — what a room sees. This one tells the code
story: the actual commit, the actual files, the actual lines that turned stock Wire-Pod into
something that survives a live demo.

**Source of truth note:** most of the code described here still lives in `~/wire-pod`, the
live runtime tree — not in this repo. Two-tree discipline holds: `~/crushn8r_git/GOPOD/` is
repo truth, `~/wire-pod/` is where the binary actually runs. This file is the showcase; it is
not a copy of the source, and no source files were pulled into this repo to write it. One
exception as of 2026-08-02, updated 2026-08-11: GOPOD's actual edits to native `.go` files now
live *only* in this repo's `wire_pod_overlay/` — the live tree holds pristine upstream code for
those files. The set shrank from 7 to 5 this pass (`logger.go` retired to fully native), and 4
new GOPOD-owned files joined the overlay holding the portable logic that used to sit inside
those touch points. See "Migration complete, 2026-08-02" and "Compiler-refereed overlay
slim-down, 2026-08-11" further down.

Everything below is read directly from commit `a0a35d4` in `~/wire-pod`
(`GOPOD Wire-Pod hardening: DAG/weather panic fixes + new source (local only)`) — 8 files
modified, 32 files new, 7,334 lines added, 355 removed. That commit exists only in the local
`~/wire-pod` clone; its remote is `kercre123/wire-pod`, an upstream fork this operator doesn't
control, so it hasn't been pushed anywhere and isn't meant to be.

**The fork point, stated once, plainly:** `~/wire-pod` forks from `kercre123/wire-pod` at
commit `11e7b22` (tag `chipper/v1.5.10`) — `git merge-base HEAD origin/main` resolves this
directly. Every "merge-base" reference elsewhere in this doc means this exact commit.

**How far a stranger could actually get repeating any of this on their own device** is a
separate, harder question than what this doc answers — see
`gopod_notes/WIRED_POD_PORTABILITY_SURVEY_001.md` for that survey specifically (what's
already env-var-portable, what's hardcoded-but-fixable, what's genuinely device-specific).
This doc stays the showcase; that one is the gap list.

---

## The hardening: two hard crashes converted to soft failures

Before this commit, two paths in wire-pod could take the whole process down:

- **`weather.go`** — three `panic(err)` call sites, one on every HTTP error from weatherapi.com
  and OpenWeatherMap. Any network hiccup during a live session killed the robot's process. Now
  all three log (`logger.Println("weather error:", err)`) and return a fixed sentinel
  (`"undefined"` / a placeholder temperature) instead of panicking.
- **The DAG anchor check**, `verifyGOPODDAGAnchor()` in `gopod_render_scaffold.go` — the
  panic-elimination story here happened in two steps. First it was stubbed: no hash comparison
  against `/gopod/ir/.dag_fingerprint.json`, just a one-time log line and an unconditional
  "verified" flag. That stub has since been pruned further, down to a plain no-op carrying the
  comment `// integrity check project deferred, see [note]` — no log line at all now, since
  there was nothing left worth logging once the check does nothing. What's deliberately still
  there: the fingerprint path, the anchor-hash constant, the violation-message string, and the
  struct to parse the fingerprint JSON — kept in place as the shape for a real future
  DAG-verification project, not as dead code to clean up. Read it as deferred, not as active
  integrity checking — nothing here verifies anything today.

Two supporting touch points changed alongside these: `kgsim_interrupt.go` gained an optional
callback (`kbTouchInterruptWrapper`) so a touch/wake interrupt can be vetoed rather than always
stopping mid-stream, and `logger.go` gained one new function, `LogDebugUI`, purely so the new
files below have somewhere to send debug-marker events.

---

## The new package: GOPOD's layer inside wire-pod's `ttr`

Everything else new lives in `chipper/pkg/wirepod/ttr/`, alongside wire-pod's own TTR
(text-to-response) code, prefixed `gopod_` or added to the existing `kgsim_` family. Grouped by
what each piece actually does:

### Animation vocabulary and normalization
- **`animation_vocab.go`** — loads `animation_vocab.json` into the one in-memory source of
  truth for every valid animation token and alias. Deliberately fails loud: a missing or
  corrupt vocab file panics on load, because a robot with no valid animation set is considered
  a startup-time problem, not a runtime one to degrade around.
- **`kgsim_cmds_animation_normalizer.go`** — rewrites every legacy animation syntax
  (`{{happy}}`, `[happy]`, `{{playAnimationWI||x}}` variants) into the one canonical form,
  falling back to `thinking` for anything it can't resolve.

### Emotional beats — the atomic unit of a robot's turn
- **`emotional_beat_actions.go`** — parses a single `actionParameter... speech` line into an
  animation call plus a say-text call, rejecting or repairing bad syntax (missing `...`, stray
  `{{`/`}}`, invalid or terminal action params like `getImage`).
- **`emotional_beat_action_plan.go`** — a dry-run compiler that turns a full multi-line LLM
  response into a sequence of beats without ever touching the robot, camera, or vision API — the
  thing that lets a scripted exchange get validated before it's allowed anywhere near hardware.

### Robot speech enforcement
- **`gopod_robot_speech_enforcement.go`** — the largest single addition. Validates every line
  of LLM output against the verified animation vocabulary, classifies exactly how a bad line
  failed (stage direction leaking through, broken command syntax, missing ellipsis, invalid
  action parameter), and normalizes what it can before falling back to requesting an LLM
  repair. One piece of it, `selectGOPODContactBeat`, is flagged in its own source comments as
  having no production call site — dead code, not a hidden feature.

### Session structure and memory
- **`gopod_section1_gominion.go`** — parses the plain-text Section Card format (`SECTION ID:`,
  `LINE N` blocks) into per-turn LLM requests for the two-Brobot scripted interview, injecting
  hidden role/personality/rules context per turn.
- **`gopod_session_memory.go`** — appends every completed exchange to a JSONL log, archives
  in chunks every 9 exchanges, and builds a byte-capped "recent memory" block injected into the
  system prompt. Worth noting plainly: its log directory is hardcoded
  (`/home/goverlord/wire-pod/chipper/sessions`) rather than routed through the path helper
  described next — an inconsistency, not a bug that's been hit yet.
- **`gopod_paths.go`** — the path helper that inconsistency refers to: a shared
  `runtime.Caller`-based resolver for paths relative to `chipper/`, overridable per-path via env
  vars, meant to replace scattered hardcoded absolute paths across the codebase.
- **`gopod_system_speech_content.go`** — loads canned fallback speech (the default prompt,
  the session-memory header, contact-beat templates) from `system_speech_content.json`, with
  hardcoded Go-literal text as the floor if that file can't be read — the one place hardcoding
  is the deliberate safety net rather than the thing being cleaned up.

### Flow control and diagnostics
- **`kgsim_marker_advance_control.go`** / **`kgsim_markers.go`** — a state machine mapping
  interrupt sources (backpack touch, wake word, keypad `3x000`) to a `GOPODMarkerAdvanceRequest`
  and emitting `GOPOD_STREAM_MARKER_0`/`_1` debug events. Its actual gating methods
  (`.ShouldAdvance()`, `.MarkerOneEvents()`) have zero live callers today — only unit tests
  exercise them; the one production call site reads just the request's `.Source` field for a
  debug log line. (Confirmed separately during a 2026-07-05 sanity sweep.)
- **`kgsim_cmds_diagnostics.go`** — env-gated (`SAVE_ROBOT_SAY_TEXT_DIR`) diagnostic dumps of
  raw/cleaned/final speech text and marker-annotated logs, for debugging a session after the
  fact without re-running it live.
- **`gopod_live_speech_gate.go`** — wraps `PerformActions` behind
  `GOPOD_ALLOW_LIVE_ROBOT_SPEECH`; if unset, the action plan computes but never dispatches to
  the robot. Read the deployment reality alongside the mechanism: `start.sh` now exports this
  var defaulting to `"1"`, so live speech is **on by default** in the normal startup path — the
  gate's real job is protecting test and offline runs, not blocking production.
- **`gopod_string_helpers.go`** — one function, `gopodNonEmptyStrings`, filtering blanks out
  of a variadic string list. The smallest file in the commit, included for completeness.

### A structural quirk worth naming
Three files — `kgsim_cmds_animation_normalizer.go`, `kgsim_cmds_diagnostics.go`, and
`kgsim_markers.go` — are production `.go` files (not `_test.go`) that directly `import
"testing"` and define `Test...` functions inline. That means the `testing` package compiles
into the production binary, while the tests inside don't get picked up by normal `go test`
discovery either, since that only scans `_test.go` files. Unconventional either way; noted here
rather than smoothed over.

---

## The plumbing that ties it together

- **`kgsim.go`** — the biggest diff in the commit: marker/regex infrastructure, phrase-unit
  parsing, the `|||0|||`/`|||C|||`/`|||1|||` canonical packet format, scenario-packet loading,
  and a rewritten three-file prompt assembly (response + identity + core, plus session memory
  and the matched scenario packet). One thing to flag rather than assume finished: an
  `interrupted` bool is declared in `StreamingKGSim` but never set `true` anywhere in this
  version, and the new touch-interrupt wrapper defaults to returning `false` — together this
  reads as a mid-stream stop path that's wired but not yet completed, not a confirmed live
  feature.
- **`kgsim_cmds.go`** — `PerformActions` now routes through the live-speech gate above instead
  of dispatching directly; that's the single choke point every animation/speech call passes
  through.
- **`sdkapp/server.go`** — the `/api-sdk/say_text` endpoint now runs any text containing
  `{{` through `GetActionsFromString` and dispatches say-text/animation actions individually,
  instead of handing the raw string straight to `robot.Conn.SayText`.
- **`start.sh`** — exports the live-speech gate default and switches the Vosk STT
  library/include paths from upstream's `/root/.vosk` convention to this deployment's actual
  paths.
- **`webroot/index.html`** — UI-only: a much taller log textarea (7 rows → 81) and a "Copy
  Logs to Clipboard" button.

---

## Outside the commit: intents, prompts, and the probe tree

Everything above lives inside commit `a0a35d4` — code committed to the local `~/wire-pod` clone.
Three more things ride alongside it in the same live tree, outside that commit's own diff:
`customIntents.json` and the two prompt `.txt` files stay untracked config, by design (confirmed
via a read-only survey of `git status --short --untracked-files=all` against that same clone);
`gopod_probes/` was untracked the same way until 2026-08-11, when it got tracked in its own
commit instead (see below). Read all three as equally real and equally load-bearing regardless
of tracked status — that status is about git history, not about whether the live system depends
on them.

### The custom intents
`customIntents.json` is the actual on/off switch for GOPOD's Wire-Pod features — three live
intents, confirmed cold-restart-surviving against the running install:
- `GOPOD_YOURSELF` — launches the core PTT demo wrapper on "GOPOD yourself" and its likely
  STT mis-hears ("go pod yourself," "go pattern yourself," "go pot your self," ...)
- `BROBOTS_INTERVIEW` — starts the two-Brobot scripted interview on "robot interview" and its
  mis-hears
- `BROBOTS_BINGO` — starts the two-brobot Bingo warm-up on "go bingo"

Each intent maps its utterances to an `exec` path — a shell script wire-pod runs when it hears
the phrase. None of this is upstream Wire-Pod behavior; the whole file is GOPOD's, and it isn't
tracked in `~/wire-pod`'s git history at all — it's a live config file, not a commit.

### The persona and response prompts
Two plain-text files carry the actual voice and format rules the LLM is held to:
- `wire-pod_brobots_prompt.txt` — the Brobots persona: high-energy, playful, cheeky party-bro
  characters whose whole bit is teasing the room toward the "GOPOD Yourself" trigger phrase.
- `wire-pod_response_prompt.txt` — the response contract: every output line must be shaped
  `actionParameter... spoken thought`, the exact format `gopod_robot_speech_enforcement.go`
  (above) validates and repairs against. The enforcement code is the guardrail; this file is the
  rule it's enforcing.

### Native vs. GOPOD's actual prompt path
Stock Wire-Pod's own knob for this is a single field: `apiConfig.json`'s
`knowledge.openai_prompt` — one string, hand-edited through the web config UI, sent to the LLM
as-is. That's the whole native mechanism: one file, one blob, no assembly step.

GOPOD doesn't run on that field. `loadCanonicalBrobotPrompt()` in `kgsim.go` reads the three
files above fresh on every single request — `wire-pod_response_prompt.txt`,
`wire-pod_brobots_prompt.txt`, and `gopod_probes/foundation_packs/gopod_core.txt` (GOPOD's own
explainer, not covered above) — concatenates them, then appends session memory and a matched
scenario packet if one applies to that message. Each of the three paths is individually
overridable by its own env var (`GOPOD_BROBOT_RESPONSE_PROMPT_PATH`,
`GOPOD_BROBOT_IDENTITY_PROMPT_PATH`, `GOPOD_CORE_PROMPT_PATH`) — external files loaded live, not
one inline value.

Worth stating plainly since it's easy to assume otherwise: `openai_prompt` still exists and
still holds real persona text, but per the actual code path it's read only as a fallback — only
reached if one or more of the three canonical files fails to load. As long as all three are
present (they are), edits to `openai_prompt` don't reach the live robot conversation at all. If
even the fallback is empty, a hardcoded floor (`defaultPrompt` / `gopodDefaultPromptFloor`) is
the last resort under that.

### animation_vocab.json itself
`animation_vocab.go` (above) is the loader; `animation_vocab.json` is the data it loads — the
actual list of verified animation tokens (`happy`, `veryHappy`, and the rest) and their underlying
Vector animation clip names. Worth naming separately: the loader panics without this file present,
so the data file is as load-bearing as the code that reads it, even though it's not Go.

As of 2026-08-10, this file is no longer live-tree-only — it rides `apply_nongo_files.sh`'s
diff-then-copy mirror alongside `start.sh`/`webroot/index.html`/`customIntents.json`/the two
prompt files (see "Migration complete, 2026-08-02" below, now a 7-file set). No loader code
change was needed: `animation_vocab.go` isn't one of the overlaid native `.go` files, so its
`runtime.Caller`-based default path always resolves to `chipper/animation_vocab.json` on the live
tree regardless of build method — the exact same path the mirror writes to. The repo copy at
`wire_pod_overlay/chipper/animation_vocab.json` and the live copy are diff-confirmed identical;
there's no separate live-only artifact left to remove, since the mirror's target path and the
loader's read path were always the same file.

### The interview/demo probe tree
`gopod_probes/` is the largest piece of this section — an entire parallel tree of content and
tooling that never got folded into the Go package above. **Tracked in `~/wire-pod`'s own git
history as of 2026-08-11** (commit `8085ced`) — cleaned first (run logs, backup binaries,
dead validator tooling, and every trace of an old single-organization pitch pilot swept out),
then committed, not left untracked indefinitely:
- **section_packets/** — the plain-text Section Cards (`section_01_brobots_gopod_card_001.txt`
  and friends) that `gopod_section1_gominion.go` parses into per-turn LLM requests
- **scenario_packs/** — `bingo.txt`, `nightclub.txt`, `crushn8r_websites.txt`, the scenario
  content `kgsim.go`'s scenario-packet loader matches against
- **tools/** — the actual runners: `run_section1_full_live_001.py` (the live interview driver)
  and the two scripts `customIntents.json` execs
  (`start_section1_interview_from_wirepod_intent_001.sh`,
  `start_bingo_from_wirepod_intent_001.sh`), plus a handful of general-purpose dry-audition
  harnesses
- **tests/** and **demo_runs/** — Python-side test coverage and a working (swept-clean, not
  archived) space for run logs and reports from real and dry-run auditions

None of it is part of commit `a0a35d4` — it's GOPOD content and tooling that lives beside the Go
changes, not inside them. It just isn't untracked anymore, the way the rest of this section still
is.

---

## The shell layer wire-pod runs inside

None of this is inside `~/wire-pod` at all — it's the operator's own shell environment, on the
same machine, and it's the thing that actually launches wire-pod's `customIntents.json` exec
scripts, restarts the service, and drives every calibration/demo alias this doc set refers to
elsewhere. Surveyed as of 2026-08-10.

**`~/.bashrc`** — standard Ubuntu interactive-shell boilerplate (history, prompt colors, PATH)
plus a few project-specific additions: `PYTHONDONTWRITEBYTECODE=1`, `OLLAMA_HOST="0.0.0.0"`
(binds Ollama to all interfaces, not just localhost — worth naming plainly rather than assuming
it's loopback-only), a CUDA 12.6 profile source, and a `bun` install PATH. It sources
`~/.bash_aliases` (below), then separately sources four GOPOD alias files directly:
`suits.sh`, `demo.sh`, `chat_capture.sh`, `wirepod_logs.sh` — `demo.sh` is also sourced by
`.bash_aliases`'s own loop, so it loads twice on a fresh shell. Harmless (bash redefines
functions/aliases idempotently) but real, and named here rather than smoothed over, same
"honest gaps" rule the rest of this doc holds to.

**`~/.bash_aliases`** — the GOPOD alias loader. Loops over `core.sh`, `brobots.sh`,
`openwebui.sh`, `llm.sh`, `goverlord.sh`, `demo.sh` (a `frame0.sh` entry was dropped from this
loop 2026-07-30, fully retired per its own comment). Then defines the two functions that most
directly govern the live wire-pod service: `wpr` — restart check (routes through
`restart_wirepod_preflight()` in the interview runner, skips the actual restart if wire-pod's
already healthy) by default, or a forced `sudo systemctl restart wire-pod` (drops any cached
sudo timestamp first via `sudo -k`, so it always re-prompts) as an explicit second option — and
`wpu` — wire-pod update: stops the service, runs wire-pod's own `update.sh` + `setup.sh
daemon-enable` from `~/wire-pod`, restarts the service.

**`~/.gopod_alias_lib/`** — the folder all of the above sources from; every GOPOD shell
function/alias lives here, one topic per file. Current contents, by role rather than by file
date: `core.sh` (stage-set/opening-chord aliases), `brobots.sh` (the Brobots motion/wake alias
family — largest file in the folder), `suits.sh`, `demo.sh`, `openwebui.sh`, `llm.sh`,
`goverlord.sh`, `chat_capture.sh`, `wirepod_logs.sh` — the sourced shell layer; `phcal_isolate_001.py`,
`phcal_apply_001.py`, `phcal_apply_control_song_001.py`, `phcal_last.json` — the phcal bench
calibration tool and its saved state; `tempo_set_001.py`, `robot_pick_001.py`,
`numpad_persona_map_001.json`, `print_numpad_map_001.py`, `gopod_json_capture.py` — smaller
standalone utilities; `__pycache__/` — compiled bytecode, not source. One file, `tools.sh`,
carries owner-only permissions (`rw-------`, unlike every other file here) — noted for
completeness, contents not inspected as part of this pass. This folder is outside git entirely
(no commit is possible for anything in it); `tech/alias_play_studio/ALIAS-LIBRARY.md` is the
canonical per-alias registry — read there for what each function actually does, not here.

---

## Since commit `a0a35d4`: two more native files touched, plus a way to stop touching them

Later work modified two more native files directly and built a mechanism meant to reduce how
often that keeps happening. Neither is part of `a0a35d4` — both are separate, later changes.

**`pkg/logger/logger.go` and `pkg/wirepod/sdkapp/server.go` (2026-07-23):** the
`/api-sdk/say_text` display-text feature printed lines as
`2026.07.23 11:33:39: BROBOT_RICH_DISPLAY Brobot 1: <text>`. Wanted: just `Brobot 1: <text>`,
every other wire-pod log line unchanged. `LogUI`/`LogDebugUI` in `logger.go` are shared by ~20
other call sites across the codebase and both unconditionally prepend a timestamp, so they
couldn't be changed directly without altering every other log line too. Fix: two new sibling
functions, `LogUIPlain`/`LogDebugUIPlain` (same underlying storage, no timestamp prepend),
added alongside the existing ones — nothing else in the logger package changed. `server.go`'s
display-text block now calls the plain versions and no longer builds the `BROBOT_RICH_DISPLAY`
tag at all.

**One correction to the record:** `ttr/gopod_render_scaffold.go` was previously counted here as
a "7th modified native file." It isn't — confirmed 2026-08-02 by checking it against the actual
upstream merge-base (`git cat-file -e 11e7b22:...` fails, the path never existed upstream
at all). It's a genuinely new GOPOD file, part of the same family as the other 32 additions from
`a0a35d4`, just added slightly later than that commit. The true native-file count was always
**9**, not 10. `gopod_render_scaffold.go` stays in the overlay folder below for build
convenience (it's still compiled as part of the `ttr` package), but it was never something to
"revert to native" — there's no native version of it to revert to.

---

## Migration complete, 2026-08-02: `~/wire-pod`'s Go source is genuinely native again

The 9 native files were reverted to their exact content at the upstream merge-base
(confirmed identical to `origin/main`'s current tip — upstream hasn't moved since GOPOD forked
from it, so there was nothing to reconcile). Split two ways, because Go's `-overlay` build flag
only helps one of them:

**7 native `.go` files — fully reverted, GOPOD's logic lives only in the overlay now:**
`logger.go`, `sdkapp/server.go`, `ttr/{kgsim, kgsim_cmds, kgsim_interrupt, weather}.go`. The
copies at `GOPOD/goverlord/wire_pod_overlay/chipper/...` (plus `gopod_render_scaffold.go`,
never-native but compiled alongside them) are the only place GOPOD's actual modifications to
these files exist now. **Any rebuild of `chipper` must go through
`wire_pod_overlay/build_with_overlay.sh` (or `deploy_overlay_build.sh` below) — a plain `go
build`/`go run` against the live tree now compiles genuine upstream code and silently drops
every GOPOD fix** (echo-defect suppression, the touch-interrupt fix, the log-order fix, all of
it). Proven, not asserted: built via the overlay against the now-pristine live tree
2026-08-02 and confirmed every GOPOD marker string still present in the result.

> **⚠ Do not "fix" these files if `git status` shows them modified.** Because the revert
> above was done as a live file-copy, never as a commit, `~/wire-pod`'s own git history
> still holds the *old, GOPOD-edited* version of these 6 files, last recorded at commit
> `a0a35d4` (HEAD has since advanced two commits past it — `34cf652`/`8085ced`, gitignore
> and `gopod_probes/` tracking housekeeping, neither touching these 6 files). That means
> `git status` inside `~/wire-pod` will show these 6 files as
> **permanently `M` (modified)**, forever, by design:
> `chipper/pkg/logger/logger.go`, `chipper/pkg/wirepod/sdkapp/server.go`,
> `chipper/pkg/wirepod/ttr/kgsim.go`, `chipper/pkg/wirepod/ttr/kgsim_cmds.go`,
> `chipper/pkg/wirepod/ttr/kgsim_interrupt.go`, `chipper/pkg/wirepod/ttr/weather.go`.
>
> Read that `M` correctly: it means the file on disk is **pristine native** and doesn't
> match wire-pod's own stale last commit — the *good* state, not a live GOPOD edit
> sitting uncommitted. **Never run `git checkout`/`git restore` on these 6 files to
> "clean up" that `M`** — doing so would silently overwrite the pristine native content
> with the old edited version from that stale commit, undoing this whole migration.
> GOPOD's actual, current logic for all 6 lives only in this repo's
> `goverlord/wire_pod_overlay/` — that's the file to edit, never the live tree copy.
> Confirmed live 2026-08-16, `WIREPOD_GOPOD_FOOTPRINT_SURVEY_001.md`.

**3 non-Go native files + 4 GOPOD-owned config/data files — stay edited in the live tree, by
design, operator-confirmed 2026-08-02 ("whatever needs to stay in native wire-pod is ok"):**
`start.sh`, `webroot/index.html`, `.gitignore`, plus `customIntents.json`, both
`wire-pod_*.txt` prompt files, and (added 2026-08-10) `animation_vocab.json`. None of these can
ride the `-overlay` flag — the first three aren't part of Go's compilation graph, the rest aren't
Go source at all (`animation_vocab.json` is data its loader reads, not code). Reverting
`start.sh` specifically has a real failure mode: native `start.sh` hardcodes upstream's
`/root/.vosk/libvosk` convention, which doesn't exist on this machine, so a reverted `start.sh`
sitting live at the moment of a crash-restart would fail to bring the service back up. All seven
now have permanent, current copies in `wire_pod_overlay/chipper/` (the first six added
2026-08-02, `animation_vocab.json` added 2026-08-10, alongside the already-existing 7 `.go`
files), applied via `apply_nongo_files.sh` — diff-then-copy, never a blind overwrite, run it any
time to confirm live still matches repo truth. This is the deliberate, documented exception to
"100% native": not a gap, a named and permanent seam. See "animation_vocab.json itself" above for
why that file's addition needed no loader code change.

**`deploy_overlay_build.sh`** is the new canonical path from "edited a file under
`wire_pod_overlay/`" to "live binary updated": builds via the overlay, sanity-checks the result,
backs up the current live binary with a timestamp, swaps the new one in. It deliberately stops
short of restarting `wire-pod.service` — firing that restart (`wpr`) stays the operator's own
action, same standing rule as every rebuild this project has done.

---

## `config-ws/webserver.go`: added, deployed, found to regress native, reverted — 2026-08-10

A short-lived 9th overlay entry, told straight so it doesn't need re-litigating. Commit `4b60bcf`
(2026-08-03) added a rich/flat `LOGS` toggle to `config-ws/webserver.go`'s `handleGetLogs` — a
flag file (`gopod_probes/gopod_rich_logs_flag.txt`, written by `pha0b`'s "rich display on
console?" prompt) could make that endpoint serve `logger.LogTrayList` (the full, unfiltered
firehose) instead of native's always-`logger.LogList` (a curated, event-driven log — intent
matches, LLM responses, nothing else). It sat undeployed in the overlay for a week; deployed
2026-08-10, then immediately found to regress the browser page's own native "Show all logs"
checkbox distinction: whenever the flag was `"1"` (a real prior answer, not a test artifact),
`/api/get_logs` and `/api/get_debug_logs` served identical content — the clean, native
"No logs yet, you must say a command to Vector" placeholder view was gone, replaced by the same
firehose regardless of the checkbox. Confirmed directly against `kercre123/wire-pod` upstream
source (`git show 11e7b22:.../webserver.go`) that native `handleGetLogs` has no such
conditional at all — this was a real, deliberate deviation from native, not a bug in testing it.

**Verdict: not worth the drift.** The toggle offered no capability the operator actually wanted
enough to keep once seen live — reverted the same day. `config-ws/webserver.go` is removed from
`overlay.json`'s `Replace` map entirely (not just content-reverted) and deleted from
`wire_pod_overlay/` — GOPOD has zero real modification to this file, so there's nothing left for
an overlay entry to do. Rebuilt via `deploy_overlay_build.sh`; the new binary has zero occurrences
of the removed feature's markers (`strings`-confirmed) while every other GOPOD marker
(`GOPOD_STREAM_MARKER`, `LogUIPlain`/`LogDebugUIPlain` — the *actually*-wanted mechanism the
chat-bubble feature's `display_text` writes ride, unaffected by this revert) is still present.
`handleGetLogs` is byte-for-byte native again. **Golden as of this line** — this file needs no
further "is it native yet" checking; if `config-ws/webserver.go` ever shows a diff against
upstream again, that's new work, not a resurfacing of this one.

---

## Compiler-refereed overlay slim-down, 2026-08-11

A cleanup pass on the same 6 files "Migration complete, 2026-08-02" reverted to native
almost went wrong in an instructive way, worth telling straight. A first attempt reasoned
"these 6 files are already byte-identical to upstream on disk, so there's nothing left to
overlay" and retired all 6 `Replace` entries. **The very next rebuild failed to compile** —
several of GOPOD's own newer files (`gopod_robot_speech_enforcement.go`,
`gopod_session_memory.go`, `kgsim_cmds_diagnostics.go`, `kgsim_markers.go`,
`emotional_beat_actions.go`, `gopod_render_scaffold.go`, `animation_vocab.go`) call
functions and reference types that only exist in the *old edited* versions of those 6
files, not the pristine native ones — the disk state and the last-successful-build state
had quietly decoupled (source reverted, but the overlay was still the only thing supplying
symbols other GOPOD files depend on). The wrong retirement was reverted immediately, never
committed, and the rule for the actual pass became: **let the compiler decide, not
eyeballing a diff.**

**The method:** for each of the 6 files, classify every top-level Go declaration (function,
type, var, const) by whether it exists in native at all. A declaration with zero native
counterpart is portable — move it verbatim to a new GOPOD-owned sibling file in the same
package. A declaration that modifies an *existing* native function's body or signature is
woven — it can't move without either behavioral risk or an import cycle, so it stays, and
the touch-point file keeps an overlay entry for exactly that piece. Verified by rebuilding
after every single move, not just at the end — 3 of those rebuilds caught real mistakes
(unused imports left behind by an extraction) immediately, fixed compiler-driven before
the next file.

| File | Result |
|---|---|
| `logger.go` | **100% portable — entry retired entirely.** All 4 GOPOD functions were pure additions; native `logger.go` is untouched upstream code again for the first time since 2026-08-02. |
| `kgsim.go` | 771→271 diff lines (65% smaller). ~35 new declarations (the marker/parsing/prompt-loading machinery) moved out to a new file. What stayed: two one-line tweaks and the two functions with a genuinely rewritten body (`CreateAIReq`, the `StreamingKGSim` streaming loop). |
| `kgsim_cmds.go` | 616→517 diff lines (16% smaller). 8 new declarations moved out. What stayed: this file is pervasively touched — nearly every existing function carries a real edit, including removing native's own `animationMap` table (superseded by `animation_vocab.json`) — a removal is woven by definition, nothing to move. |
| `weather.go` | 116→60 diff lines (48% smaller). The 2 geocoding helper functions moved out. What stayed: the 3 `panic()`→soft-fail swaps — the hardening this doc opens with — and the one-line call-site change to use the extracted helpers. |
| `sdkapp/server.go` | 59→43 diff lines. One purely-cosmetic whitespace hunk reverted to match native exactly. The `{{action}}` parser itself couldn't move — it calls `robot.Conn`/`ctx`, local to the handler, and moving it into `wirepod_ttr` would create an import cycle (`sdkapp` already imports that package). |
| `kgsim_interrupt.go` | No change possible — 0% portable. The entire diff is a variadic parameter added to the one existing function plus logic threaded through its body; nothing in it is a standalone addition. |

**Net result:** overlay diff footprint across these 6 files, 1620→910 lines (44% smaller)
overall. Net `Replace` entry count in `overlay.json`: **10** (was 7) — 5 native files still
carry an unavoidable edit, 5 are now GOPOD-owned files with no native counterpart at all
(`gopod_render_scaffold.go` plus 4 new: `gopod_logger_extensions.go`,
`gopod_kgsim_response_shaping.go`, `gopod_kgsim_cmds_extras.go`,
`gopod_weather_geocode.go`). A bigger entry count reads as regress at a glance — it isn't;
it's the honest shape of "migrate everything separable into the overlay, leave only what's
truly woven in native," which is a smaller total footprint even though it's more files.
Deployed via `deploy_overlay_build.sh` the same way every rebuild in this doc has been —
file swap only, `wire-pod.service` never restarted mid-pass, confirmed healthy before and
after. Full report: `gopod_notes/WIREPOD_OVERLAY_MIGRATION_EXECUTED_001.md`.

---

## Test coverage that shipped alongside it

Of the 32 new files, roughly half are `_test.go` counterparts to the production files above.
The rest are scenario-level audition harnesses, not narrow unit tests — they build real
prompts, optionally call a live LLM, and score the response against format/persona/coverage
rules, writing a JSON (and sometimes markdown) report:

- `gopod_demo_probe_test.go` — QA harness for the WDTM demo script
- `gopod_guided_reveal_audition_test.go` / `gopod_guided_reveal_neutral_llm_audition_test.go`
  — audition passes for the guided-reveal cue-card exchange
- `gopod_level0_interview_test.go` — the largest test file in the commit; adaptive scripted
  interview coverage, value-point/cluster scoring, standard and 25-line dry-run auditions
- `gopod_brobot2_llm_path_test.go` — the Brobot 2 ESN path through prompt build → dry compile,
  asserting the dry path never reaches robot/camera/vision calls
- `kgsim_pipeline_test.go` — the conventional unit layer: marked/filtered/sent packet
  separation, canonical prompt ordering, scenario-packet placement, session-memory bounding

---

## Where this stands today

Confirmed by reading the commit directly, not carried forward from memory:

- Weather panics: eliminated, converted to logged errors — real fix
- DAG panic: eliminated in two steps — first stubbed, then pruned further to an inert no-op
  carrying a deferred-project comment; the fingerprint/hash/violation scaffolding is kept on
  purpose as the shape for a future check, not as an active one today
- Live speech gate: present, but defaults to on via `start.sh` — a test safety net, not a
  production block
- Marker advance control: constructed and logged live; wired into the touch-interrupt wrapper
  2026-08-02 (see "The rule: stay thin" above) — first production caller
- Mid-stream touch-interrupt stop: fixed 2026-08-02, build-verified (`gofmt`/`go vet` clean),
  deployed to the live binary via `deploy_overlay_build.sh` — **not yet active** as of this
  line, since `wire-pod.service` hasn't been restarted onto it (that's the operator's own
  action). The `interrupted` bool itself is still never set, but that flag has no reader in
  this version either, so it isn't part of the stop path this fix restores
- One dead-code path acknowledged in its own source comments (`selectGOPODContactBeat`)
- One path inconsistency (`gopod_session_memory.go` hardcodes what `gopod_paths.go` exists to
  generalize)

This is the same "not a pitch deck" rule the rest of the doc set holds to — the honest gaps are
part of what makes the working parts credible.

---

> From Doctrine Barfallonyou
> Lesson! You don't showcase a system by hiding what's still a stub. You showcase it by naming the stub and shipping the rest anyway.
> Boom! Done! Class Dismissed!
> — Doc Squawkadoodle

---

## GOPOD YAHMM (You Are Here Mall Map)

Part of GOPOD — see [tech/README.md](README.md) for everything else in this folder, or [the root map](../README.md) for the rest of GOPOD.

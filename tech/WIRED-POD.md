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

Native Wire-Pod, everywhere, except eight touch points: five native `.go` files that carry a
real edit (`sdkapp/server.go`, `ttr/{kgsim, kgsim_cmds, kgsim_interrupt, weather}.go`) plus three
non-Go files (`start.sh`, `webroot/index.html`, `.gitignore`) — the `GOPOD_STREAM_MARKER_0`/`_1`
debug-marker wrapper, and the marker-1-advance path that lets touch/wake/keypad interrupts end a
stream early. Nothing outside those eight is expected to behave differently from stock Wire-Pod.
Before touching any native file under `~/wire-pod`, check whether the change fits inside one of
the eight — if it doesn't, that's the layer growing past "thin," not a normal edit.

Fitting inside one of the eight is necessary but not sufficient. Before *deploying* any change to
a native touch-point file, confirm the file's actual current upstream behavior against real
source (`git show 11e7b22:<path>` in `~/wire-pod`'s own clone — that's the merge-base, `chipper/
v1.5.10` on `kercre123/wire-pod`, see "What this document is" below) if the change could alter
anything user-facing in the native UI — not just whether GOPOD's own new code works in isolation.
A change that's correctly coded can still be an unwanted deviation from native the moment it's
live.

"Stay thin" is enforced structurally, not just by convention: the live `~/wire-pod` tree holds
genuine upstream code for all eight touch points, and GOPOD's actual edits exist only in this
repo's `wire_pod_overlay/`, injected at build time. Editing the live tree directly does nothing on
the next overlay build — the overlay folder is the only place left to make the change. See "How
the overlay build works" further down for the full mechanism.

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
story: the actual files, the actual lines that turned stock Wire-Pod into something that
survives a live demo.

**Source of truth note:** most of the code described here still lives in `~/wire-pod`, the
live runtime tree — not in this repo. Two-tree discipline holds: `~/crushn8r_git/GOPOD/` is
repo truth, `~/wire-pod/` is where the binary actually runs. This file is the showcase; it is
not a copy of the source, and no source files were pulled into this repo to write it. GOPOD's
actual edits to native `.go` files live *only* in this repo's `wire_pod_overlay/` — the live
tree holds pristine upstream code for those files.

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

- **`weather.go`** — used to `panic(err)` on every HTTP error from weatherapi.com and
  OpenWeatherMap, taking the whole robot process down on any network hiccup during a live
  session. Now every error path logs (`logger.Println("weather error:", err)`) and returns
  a fixed sentinel (`"undefined"` / a placeholder temperature) instead of panicking.
- **The DAG anchor check**, `verifyGOPODDAGAnchor()` in `gopod_render_scaffold.go` — a
  deliberate no-op today, carrying the comment `// integrity check project deferred, see
  [note]`. No hash comparison against `/gopod/ir/.dag_fingerprint.json`, no log line —
  nothing here verifies anything. What's deliberately still there: the fingerprint path,
  the anchor-hash constant, the violation-message string, and the struct to parse the
  fingerprint JSON — kept as the shape for a real future DAG-verification project, not as
  dead code to clean up.

Two supporting touch points changed alongside these: `kgsim_interrupt.go` carries an
optional callback (`kbTouchInterruptWrapper`) so a touch/wake interrupt can be vetoed
rather than always stopping mid-stream, and `logger.go`'s GOPOD content (now moved to a
standalone GOPOD-owned file, see "How the overlay build works" below) adds `LogDebugUI`,
purely so the new files below have somewhere to send debug-marker events.

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
- **`gopod_robot_speech_enforcement.go`** — the largest single file in the package. Validates
  every line of LLM output against the verified animation vocabulary, classifies exactly how a
  bad line failed (stage direction leaking through, broken command syntax, missing ellipsis,
  invalid action parameter), and normalizes what it can before falling back to requesting an
  LLM repair. One piece of it, `selectGOPODContactBeat`, is flagged in its own source comments
  as having no production call site — dead code, not a hidden feature.

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
  debug log line.
- **`kgsim_cmds_diagnostics.go`** — env-gated (`SAVE_ROBOT_SAY_TEXT_DIR`) diagnostic dumps of
  raw/cleaned/final speech text and marker-annotated logs, for debugging a session after the
  fact without re-running it live.
- **`gopod_live_speech_gate.go`** — wraps `PerformActions` behind
  `GOPOD_ALLOW_LIVE_ROBOT_SPEECH`; if unset, the action plan computes but never dispatches to
  the robot. Read the deployment reality alongside the mechanism: `start.sh` exports this
  var defaulting to `"1"`, so live speech is **on by default** in the normal startup path — the
  gate's real job is protecting test and offline runs, not blocking production.
- **`gopod_string_helpers.go`** — one function, `gopodNonEmptyStrings`, filtering blanks out
  of a variadic string list. The smallest file in the package, included for completeness.

### A structural quirk worth naming
Three files — `kgsim_cmds_animation_normalizer.go`, `kgsim_cmds_diagnostics.go`, and
`kgsim_markers.go` — are production `.go` files (not `_test.go`) that directly `import
"testing"` and define `Test...` functions inline. That means the `testing` package compiles
into the production binary, while the tests inside don't get picked up by normal `go test`
discovery either, since that only scans `_test.go` files. Unconventional either way; noted here
rather than smoothed over.

---

## The plumbing that ties it together

- **`kgsim.go`** — marker/regex infrastructure, phrase-unit parsing, the
  `|||0|||`/`|||C|||`/`|||1|||` canonical packet format, scenario-packet loading, and a
  three-file prompt assembly (response + identity + core, plus session memory and the matched
  scenario packet). One thing to flag rather than assume finished: an `interrupted` bool is
  declared in `StreamingKGSim` but never set `true` anywhere in this version, and the
  touch-interrupt wrapper defaults to returning `false` — together this reads as a mid-stream
  stop path that's wired but not yet completed, not a confirmed live feature.
- **`kgsim_cmds.go`** — `PerformActions` routes through the live-speech gate above instead
  of dispatching directly; that's the single choke point every animation/speech call passes
  through.
- **`sdkapp/server.go`** — the `/api-sdk/say_text` endpoint runs any text containing
  `{{` through `GetActionsFromString` and dispatches say-text/animation actions individually,
  instead of handing the raw string straight to `robot.Conn.SayText`.
- **`start.sh`** — exports the live-speech gate default and switches the Vosk STT
  library/include paths from upstream's `/root/.vosk` convention to this deployment's actual
  paths.
- **`webroot/index.html`** — the taller log textarea (81 rows) and "Copy Logs to Clipboard"
  button live here. This file, and the CSS beside it, also carry a much bigger feature — see
  "Brobots Chat Bubbles" below.

---

## Intents, prompts, and the probe tree

Three more things live in the same `~/wire-pod` tree, alongside the Go package above:
`customIntents.json`, the two prompt `.txt` files, and `gopod_probes/` — read all three as
equally real and equally load-bearing.

### The custom intents
`customIntents.json` is the actual on/off switch for GOPOD's Wire-Pod features — three live
intents:
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

GOPOD doesn't run on that field. `loadCanonicalBrobotPrompt()` in `kgsim.go` reads three
files fresh on every single request — `wire-pod_response_prompt.txt`,
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

### `animation_vocab.json` itself
`animation_vocab.go` (above) is the loader; `animation_vocab.json` is the data it loads — the
actual list of verified animation tokens (`happy`, `veryHappy`, and the rest) and their underlying
Vector animation clip names. Worth naming separately: the loader panics without this file present,
so the data file is as load-bearing as the code that reads it, even though it's not Go.

It rides `apply_nongo_files.sh`'s diff-then-copy mirror alongside `start.sh`/`webroot/
index.html`/`customIntents.json`/the two prompt files (see "How the overlay build works"
below). No loader code change was needed: `animation_vocab.go` isn't one of the overlaid native
`.go` files, so its `runtime.Caller`-based default path always resolves to
`chipper/animation_vocab.json` on the live tree regardless of build method — the exact same path
the mirror writes to.

### The interview/demo probe tree
`gopod_probes/` is the largest piece of this section — an entire parallel tree of content and
tooling that never got folded into the Go package above:
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

---

## The shell layer wire-pod runs inside

None of this is inside `~/wire-pod` at all — it's the operator's own shell environment, on the
same machine, and it's the thing that actually launches wire-pod's `customIntents.json` exec
scripts, restarts the service, and drives every calibration/demo alias this doc set refers to
elsewhere.

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
`openwebui.sh`, `llm.sh`, `goverlord.sh`, `demo.sh`. Then defines the two functions that most
directly govern the live wire-pod service: `wpr` — restart check (routes through
`restart_wirepod_preflight()` in the interview runner, skips the actual restart if wire-pod's
already healthy) by default, or a forced `sudo systemctl restart wire-pod` (drops any cached
sudo timestamp first via `sudo -k`, so it always re-prompts) as an explicit second option — and
`wpu` — wire-pod update: stops the service, runs wire-pod's own `update.sh` + `setup.sh
daemon-enable` from `~/wire-pod`, restarts the service.

**`~/.gopod_alias_lib/`** — the folder all of the above sources from; every GOPOD shell
function/alias lives here, one topic per file. Current contents, by role: `core.sh`
(stage-set/opening-chord aliases), `brobots.sh` (the Brobots motion/wake alias family —
largest file in the folder), `suits.sh`, `demo.sh`, `openwebui.sh`, `llm.sh`,
`goverlord.sh`, `chat_capture.sh`, `wirepod_logs.sh` — the sourced shell layer; `phcal_isolate_001.py`,
`phcal_apply_001.py`, `phcal_apply_control_song_001.py`, `phcal_last.json` — the phcal bench
calibration tool and its saved state; `tempo_set_001.py`, `robot_pick_001.py`,
`numpad_persona_map_001.json`, `print_numpad_map_001.py`, `gopod_json_capture.py` — smaller
standalone utilities; `__pycache__/` — compiled bytecode, not source. One file, `tools.sh`,
carries owner-only permissions (`rw-------`, unlike every other file here) — noted for
completeness, contents not inspected. This folder is outside git entirely
(no commit is possible for anything in it); `tech/alias_play_studio/ALIAS-LIBRARY.md` is the
canonical per-alias registry — read there for what each function actually does, not here.

---

## Songs, PLAYHEAD, and phcal: the performance layer built on top

Everything above is Wire-Pod itself — the engine one robot talks through. This layer is
GOPOD's own, running beside it on the same machine: the song format, the cockpit that
plays a song, and the bench that tunes one movement/audio primitive at a time. None of it
is Wire-Pod code; all of it drives Wire-Pod through the same `/api-sdk/*` endpoints
described above (`assume_behavior_control`, `say_text`, `release_behavior_control`).

### A song, as a file pair
Every song lives in its own folder under `goverlord/runtime/songs/` — two files hold the
whole thing:
- **`story.md`** — the text. `## STEP <step_id>` headings, each with a `> TEXT:` line (the
  spoken content) and an optional `> FAIL:` line. `{robot_name}` is the one substitution
  token, replaced with the actual speaker's name at runtime (`say_connected`'s TEXT reads
  `{robot_name} connected. Loading next test.` — spoken as "Brobot 1 connected..." or
  "Brobot 2 connected...", whichever robot is talking).
- **`knobs.json`** — the mechanical parameters, one object per step: `note` (what kind of
  action — `say`, `arm_cue`, `nod`, `weather`, `pause`, `connect`, `exit`, and a few more),
  `speaker`, plus whatever that note type needs (`cycles`/`hold_seconds`/`speed` for
  `arm_cue`, `pause_seconds`/`gap_label` for `pause`). Text and mechanics are deliberately
  separate files — editing a line's wording never touches its timing, and vice versa.

One runner reads both and plays them: `run_golden_song_001.py`'s `SONG_REGISTRY`
(`:193`) maps a song id to its folder, which runner function handles it, and a few
per-song flags (`manage_control` — does this song hold one continuous connection instead
of assume/release per step; `synthesize_speaker` — does the whole run target one env-var-
picked robot regardless of each step's own `speaker` field, the mechanism behind
`00_brobots_awaken`'s two-pass "run brobot 1's whole song, then brobot 2's" mode).
`story.md` is parsed by `parse_control_story_md()`; each step dispatches by its `note`
into the matching action. A run's full result — every step, every HTTP response, timing —
saves to `<song>/runs/golden_run_<timestamp>.json`.

### PLAYHEAD — the cockpit that plays a song
`pha0b` (`~/.gopod_alias_lib/brobots.sh:1519`) — "PlayHead A/0/B." Bare `pha0b` opens a
menu: pick a song off disk, pick a point A/point B range (or the whole song), and it runs
that slice live, no separate copy/paste step. A few prompts along the way, each with a
plain default: reporter-gap seconds (a dry, run-scoped override — never rewrites
`knobs.json`), whether to pull phcal's last-confirmed tuned values into this run, which
robot(s) to run, live speech on or dry. Called with explicit arguments
(`pha0b <song> <point_a> <point_b> [robot]`) skips the menu entirely.

### phcal — the bench that tunes one primitive
`phcal` (`~/.gopod_alias_lib/brobots.sh:2760`) — the sibling tool. Where `pha0b` plays
the score, `phcal` tunes the instrument: one isolated movement/audio/animation primitive
(arm, nod, cube light, rattle, an animation token), live, on one robot, independent of any
song. A startup probe checks which robots are present and their battery before the menu
draws. Every tunable primitive walks pre-filled with its last-used value (Enter keeps it,
typing overrides) and saves what fired to `phcal_last.json` — the same file `pha0b`'s
"pull phcal's tuned values into this run" prompt reads from, and the same file a direct
flag call (`phcal arm 1 --hold 1.3 --cycles 3`) reads and writes too. Movement/audio logic
is shared with the golden song runner (`run_robot_control_song_001.py`'s
`run_move_axis()`/`run_arm_cue()`/`run_nod()`) via `~/.gopod_alias_lib/
phcal_isolate_001.py` — tuning a primitive in phcal and running it inside a song are the
same underlying code path, not two implementations to keep in sync.

### Where to actually read more
This section is the map, not the manual — every alias, every menu option, every knob
`pha0b`/`phcal` expose is catalogued in one place: `tech/alias_play_studio/
ALIAS-LIBRARY.md`. Read that for what a specific alias does; read this section for how the
pieces fit together.

---

## Brobots Chat Bubbles: the webroot UI feature

A second, richer view of the same `/api/get_logs` feed the plain log textarea already
reads, sitting beside it as a sibling element inside the native `#section-log` block, not
layered on top of it — only one view is ever meant to be visible at a time. Wholly
GOPOD-owned front-end: one small CSS file (`webroot/css/gopod_chat_bubbles.css`) and one
inline `<script>` block inside `index.html` itself, both riding the same
`apply_nongo_files.sh` diff-then-copy mirror as every other non-Go overlay file (see "How
the overlay build works" below).

**The container, dormant by default.** `#brobot-chat-bubbles-container` (`index.html:166`)
sits beside `#botTranscriptedTextArea`, the native flat-log textarea. Its base CSS
(`gopod_chat_bubbles.css:16-20`) is `height: 0; overflow: hidden;` — the page looks
exactly like stock Wire-Pod unless a real run turns it on. State comes from a plain JSON
file the page polls every 500ms, `webroot/gopod_rich_display_ui_state.json` — not a
Wire-Pod endpoint, a flag file GOPOD's own Python runtime and shell layer write directly
into this same webroot directory.

**Turning it on, and clearing on a genuinely new run.** `Robots.__init__`
(`run_section1_full_live_001.py`) writes `{"expanded": true, "run_active": true}` once
per song run/pass, gated on `console_rich_display` (pha0b's own "rich display on
console?" y/n). `pollState()` (`index.html:323`) reads it, adds the `.expanded` class,
and the container animates open (`transition: height 0.25s ease`). `pollBubbles()`
(`index.html:307`) separately polls `/api/get_logs` and turns every `"Brobot 1: <text>"`
/ `"Brobot 2: <text>"` line into a bubble via `appendBubble()`. Clearing old bubbles is
keyed to `run_active` flipping false→true (`index.html:360`), not merely `expanded`
staying true — pha0b keeps `expanded` continuously true across an entire invocation
(including a two-pass, two-robot run), so keying the reset to `expanded`'s own
transition alone misses every run after the first following a manual close.
`run_active`'s own transition, written exactly once per process/pass by
`Robots.__init__`, is the real "a new run just began" signal.

**The close button.** `#brobot-chat-bubble-close` (`index.html:167`) is the first element
in the markup — bubbles append after it via JS. `position: sticky`/`position: absolute`
both fail here: sticky only offsets an element during scroll past its own normal-flow
position, it doesn't relocate by DOM order, so a sticky button renders wherever it sits
in the markup (the top, in this case) regardless of the offset given. Absolute
positioning inside a scrolling container (`overflow-y: auto`) measures its offset
against the full scrollable *content* height, not the visible viewport, so it scrolls
away with the content instead of staying fixed. What actually works: `order: 1`
(`gopod_chat_bubbles.css:51`) on the flex child — bubbles keep their implicit `order: 0`,
so the button always flex-sorts after every bubble, in normal flow, scrolling with the
rest of the content, however many bubbles exist. No DOM/JS change needed for that part.

Being real flex content means the button grows the container's own `scrollHeight` the
same way a bubble does — a view already scrolled to the bottom doesn't automatically
follow that growth, so the button can render below the visible line even though it's
correctly positioned. Fixed with a double `requestAnimationFrame` scroll
(`index.html:405-424`), fired only on the actual hidden→visible transition (a
`closeBtnVisible` tracking var), not on every 500ms poll tick — an unconditional
scroll-to-bottom every tick would fight a user who deliberately scrolled up to re-read
earlier bubbles while the run sits finished.

**Closing persists across a page refresh.** `manuallyClosed` (`index.html:236`) is a
plain in-memory flag — a page refresh reruns the whole script from scratch, which would
forget it on its own. Persisted via `localStorage`, keyed to that specific run's own
`finished_at` epoch timestamp (`getClosedMarker`/`setClosedMarker`, `index.html:257-268`),
not a bare boolean — a refresh re-checks "did I close *this* run," so a genuinely new run
(a different `finished_at`, or `run_active` flipping true again) is never suppressed by
an old close. Wrapped in try/catch; a storage failure (private browsing, storage
disabled) falls back to no persistence rather than breaking the feature. Honest limit,
same as any `localStorage` use: this is per-browser state, not shared across viewers and
never synced back to GOPOD.

**Survives a wire-pod restart too.** The state file above is a plain file in `webroot/`
— a service restart never touches it on its own, so a stale `expanded: true` written
before the restart can bleed through on the next page load, independent of the
per-browser close-persistence above. Both restart paths force it dormant on an actual
restart: `restart_wirepod_preflight()`'s own RESET_UI step
(`run_section1_full_live_001.py:2567` — the shared CHECK→CLEAR→MIRROR→START→CONFIRM
function `wpr check` and every song-runner restart route through, see "The shell layer
wire-pod runs inside" above), and `_wpr_force()` (`~/.bash_aliases`, the `wpr 2` path
that bypasses that shared function entirely) writing the identical
`{"expanded": false, "run_active": false}` literal directly.

**Background color, read live from the robot, not hardcoded.** Each bubble's color
(`.brobot-bubble.brobot-1` / `.brobot-2`, `gopod_chat_bubbles.css:80-89`) comes from a CSS
custom property (`--brobot-1-color` / `--brobot-2-color`) set by `pollEyeColors()`
(`index.html:440`), which polls a second small state file,
`webroot/gopod_brobot_eye_colors.json`. That file is written once per run start by
`_write_brobot_eye_colors_state()` (`run_section1_full_live_001.py:92`), which reads each
robot's real `custom_eye_color` hue/saturation straight from `jdocs.json`'s live
`vic.RobotSettings` — the same "one source of truth: the robot's own real settings"
principle `load_robot_format_from_jdocs()` already uses for weather units — and converts
it via `colorsys.hsv_to_rgb()` (Vector's stored hue is already the 0–1 fraction that
function expects, no manual degree math needed) into a plain hex string. A color changed
on the robot's own app is picked up by the *next* run, never mid-run — read once at
`Robots.__init__` time, not polled continuously on the Python side. The CSS carries a
static fallback (the two robots' actual colors as last read: `#FF001A` / `#0009FF`) for
before that JS runs, or if the color file's ever missing.

Full incident/build history for this feature, if it's ever needed: `gopod_notes/
WEATHER_PRONUNCIATION_AND_CHAT_BUBBLES_FIXES_001.md` through `_004.md`.

---

## How the overlay build works

Go's `-overlay` build flag is how GOPOD keeps `~/wire-pod`'s own Go source genuinely
native while still shipping real edits. The 8 touch points named in "The rule: stay
thin" above are split two ways, because the overlay flag only helps one of them:

**The `.go` touch points — reverted to exact upstream content on disk; GOPOD's logic
lives only in the overlay.** `sdkapp/server.go`, `ttr/{kgsim, kgsim_cmds, kgsim_interrupt,
weather}.go`, plus `logger.go`'s own small addition. The copies at `GOPOD/goverlord/
wire_pod_overlay/chipper/...` are the only place GOPOD's actual modifications to these
files exist. **Any rebuild of `chipper` must go through `wire_pod_overlay/
build_with_overlay.sh` (or `deploy_overlay_build.sh` below) — a plain `go build`/`go run`
against the live tree compiles genuine upstream code and silently drops every GOPOD fix.**

> **⚠ Do not "fix" these files if `git status` shows them modified.** The revert to
> native was done as a live file-copy, never as a commit, so `~/wire-pod`'s own git
> history still holds the *old, GOPOD-edited* version of these files. That means
> `git status` inside `~/wire-pod` shows them as **permanently `M` (modified)**, by
> design: `chipper/pkg/logger/logger.go`, `chipper/pkg/wirepod/sdkapp/server.go`,
> `chipper/pkg/wirepod/ttr/kgsim.go`, `chipper/pkg/wirepod/ttr/kgsim_cmds.go`,
> `chipper/pkg/wirepod/ttr/kgsim_interrupt.go`, `chipper/pkg/wirepod/ttr/weather.go`.
> Read that `M` correctly: it means the file on disk is **pristine native** and doesn't
> match wire-pod's own stale last commit — the *good* state, not a live GOPOD edit
> sitting uncommitted. **Never run `git checkout`/`git restore` on these files to "clean
> up" that `M`** — doing so silently overwrites the pristine native content with the old
> edited version from that stale commit. GOPOD's actual, current logic for all of them
> lives only in this repo's `goverlord/wire_pod_overlay/` — that's the file to edit,
> never the live tree copy.

**Portable vs. woven.** Every top-level Go declaration in a touch-point file is either
portable (no native counterpart — moved verbatim to a new GOPOD-owned sibling file in
the same package, compiled alongside but needing no `Replace` entry of its own) or woven
(modifies an existing native function's body or signature — can't move without
behavioral risk or an import cycle, stays in the touch-point file, which keeps an
overlay entry for exactly that piece):

| File | What's woven in (stays in the touch-point file) |
|---|---|
| `logger.go` | Nothing — 100% portable, native `logger.go` is untouched upstream code. |
| `kgsim.go` | Two one-line tweaks and the two functions with a genuinely rewritten body (`CreateAIReq`, the `StreamingKGSim` streaming loop). Everything else (marker/parsing/prompt-loading machinery, ~35 declarations) lives in a portable sibling file. |
| `kgsim_cmds.go` | Pervasively touched — nearly every existing function carries a real edit, including removing native's own `animationMap` table (superseded by `animation_vocab.json`). |
| `weather.go` | The 3 `panic()`→soft-fail swaps (the hardening this doc opens with) and the one-line call-site change to use the extracted geocoding helpers. |
| `sdkapp/server.go` | The `{{action}}` parser — it calls `robot.Conn`/`ctx`, local to the handler, and moving it into `wirepod_ttr` would create an import cycle (`sdkapp` already imports that package). |
| `kgsim_interrupt.go` | The whole diff — a variadic parameter added to the one existing function plus logic threaded through its body; nothing in it is a standalone addition. |

The portable declarations that moved out live in GOPOD-owned sibling files compiled
alongside the touch points: `gopod_render_scaffold.go`, `gopod_logger_extensions.go`,
`gopod_kgsim_response_shaping.go`, `gopod_kgsim_cmds_extras.go`,
`gopod_weather_geocode.go`.

**Non-Go files — stay edited in the live tree, by design, applied via a mirror
script.** `start.sh`, `webroot/index.html`, `.gitignore` (Go's `-overlay` flag doesn't
apply — the first two aren't part of the compilation graph, `.gitignore` isn't source at
all), plus four GOPOD-owned config/data files that aren't Go source either:
`customIntents.json`, both `wire-pod_*.txt` prompt files, and `animation_vocab.json`.
Reverting `start.sh` specifically has a real failure mode: native `start.sh` hardcodes
upstream's `/root/.vosk/libvosk` convention, which doesn't exist on this machine, so a
reverted `start.sh` sitting live at the moment of a crash-restart would fail to bring
the service back up. All of these have permanent, current copies in
`wire_pod_overlay/chipper/`, applied via `apply_nongo_files.sh` — diff-then-copy, never
a blind overwrite, run it any time to confirm live still matches repo truth. This is the
deliberate, documented exception to "100% native": not a gap, a named and permanent seam.

**`deploy_overlay_build.sh`** is the canonical path from "edited a file under
`wire_pod_overlay/`" to "live binary updated": builds via the overlay, sanity-checks the
result, backs up the current live binary with a timestamp, swaps the new one in. It
deliberately stops short of restarting `wire-pod.service` — firing that restart (`wpr`)
stays the operator's own action.

---

## `config-ws/webserver.go`: not part of the overlay

A rich/flat `LOGS` toggle was tried on `config-ws/webserver.go`'s `handleGetLogs` — a
flag file could make that endpoint serve `logger.LogTrayList` (the full, unfiltered
firehose) instead of native's always-`logger.LogList` (a curated, event-driven log —
intent matches, LLM responses, nothing else). It regressed the browser page's own native
"Show all logs" checkbox distinction: whenever the flag was on, `/api/get_logs` and
`/api/get_debug_logs` served identical content — the clean, native "No logs yet, you
must say a command to Vector" placeholder view was gone, replaced by the same firehose
regardless of the checkbox. Native `handleGetLogs` has no such conditional at all —
confirmed directly against `kercre123/wire-pod` upstream source.

**Not worth the drift.** `config-ws/webserver.go` carries zero GOPOD modification — it's
byte-for-byte native, not one of the touch points. If it ever shows a diff against
upstream, that's new work, not a resurfacing of this one.

---

## Test coverage that shipped alongside the new package

Roughly half the test files are `_test.go` counterparts to the production files above.
The rest are scenario-level audition harnesses, not narrow unit tests — they build real
prompts, optionally call a live LLM, and score the response against format/persona/coverage
rules, writing a JSON (and sometimes markdown) report:

- `gopod_demo_probe_test.go` — QA harness for the WDTM demo script
- `gopod_guided_reveal_audition_test.go` / `gopod_guided_reveal_neutral_llm_audition_test.go`
  — audition passes for the guided-reveal cue-card exchange
- `gopod_level0_interview_test.go` — the largest test file; adaptive scripted
  interview coverage, value-point/cluster scoring, standard and 25-line dry-run auditions
- `gopod_brobot2_llm_path_test.go` — the Brobot 2 ESN path through prompt build → dry compile,
  asserting the dry path never reaches robot/camera/vision calls
- `kgsim_pipeline_test.go` — the conventional unit layer: marked/filtered/sent packet
  separation, canonical prompt ordering, scenario-packet placement, session-memory bounding

---

## Where this stands today

- Weather panics: eliminated, converted to logged errors — real fix
- DAG panic: eliminated — the check itself is now an inert no-op carrying a
  deferred-project comment; the fingerprint/hash/violation scaffolding is kept on purpose
  as the shape for a future check, not as an active one today
- Live speech gate: present, but defaults to on via `start.sh` — a test safety net, not a
  production block
- Marker advance control: constructed and logged live; wired into the touch-interrupt
  wrapper (see "The rule: stay thin" above) — its production caller
- Mid-stream touch-interrupt stop: build-verified (`gofmt`/`go vet` clean), deployed to
  the live binary — becomes active only once `wire-pod.service` is restarted onto it,
  same as any overlay change (that's the operator's own action, never automatic). The
  `interrupted` bool itself is still never set, but that flag has no reader in this
  version either, so it isn't part of the stop path this fix restores
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

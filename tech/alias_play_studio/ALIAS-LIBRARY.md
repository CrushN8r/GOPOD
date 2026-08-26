# ALIAS-LIBRARY

> Every key here is a real alias or shell function loaded into this machine's shell right
> now. Press one, GOPOD does something — an interview starts, a robot cues, a log streams.
> This is the one registry: every alias/function actually loaded on this machine, what it
> resolves to, its file:line, the story behind it, and how it renders once played.

**Backup, 2026-08-13**: `~/.gopod_alias_lib/` — the live shell source every row below
describes — was a single point of failure (real, daily-exercised code, not git-tracked,
no backup anywhere). A scrubbed, showcase-only mirror of the performance/song/cockpit
core now lives in this repo at `goverlord/alias_lib_overlay/` (its own README explains
the split and what's excluded). That folder is never sourced live — this doc still
describes the real, running shell, not the mirror.

Surveyed fresh 2026-07-10, directly from the real files, not recalled from prose.
**Re-swept 2026-07-16**, and **consolidated 2026-07-16**:
`ALIAS-PIANO.md` and `ALIAS-MIXER.md` folded into this one doc — one fact, one home, no
duplicated prose across three files describing the same board. Both files were kept on as
tombstones (pointer-only stubs) rather than deleted outright at the time, per a
zero-caller-deletion rule — MIXER had zero remaining external callers by then, but PIANO
still had seven (`life/`/`learned/` docs linking to it with stale "every live alias" prose).
**Hygiene pass, 2026-07-23:** those seven callers were repointed here, and with both files
genuinely at zero callers, both tombstones were deleted outright — see this doc's own footer,
which no longer lists either. Same two-tree discipline as the rest of this set: none of
`~/.gopod_alias_lib/*.sh` or `~/.bash_aliases`/`~/.bashrc` lives in this repo — this doc is
the map, not the territory. Every key marked **RETIRED** below is still physically present in
its file, commented out in place rather than deleted — pressed, it does nothing.

---

## What this doc covers, and what moved here

- **The registry** (below) — every alias/function, grouped by source file, with what it
  does, its status (live/pinned/retired), and the history behind it. This absorbs
  everything `ALIAS-PIANO.md` used to carry as prose.
- **[Render Controls](#render-controls)** — how a note actually renders once played: async
  vs. sync, hold/timing, temperature-style dials. This absorbs everything
  `ALIAS-MIXER.md` used to carry.
- **Today's findings** (dated sections below, oldest ones archived to
  [ALIAS-LIBRARY-FINDINGS-ARCHIVE.md](ALIAS-LIBRARY-FINDINGS-ARCHIVE.md) once more than two
  accumulate) — crash-diagnosis, golden-shape, and drift results, canonized with citations.
- **[ALIAS-SEQUENCER.md](ALIAS-SEQUENCER.md)** stays separate — the arrangement layer
  (which notes, in what order, wired into a sequence or song) is a genuinely different
  concern from "what does this key do" and "how does it render," so it keeps its own doc.

`ALIAS-PIANO.md` and `ALIAS-MIXER.md` are gone outright as of 2026-07-23 (see above) — this
doc and `ALIAS-SEQUENCER.md` are the only two alias-tier docs left.

---

## Golden rules

Standing principles, distinct from the dated "Today's findings" below (those get archived once
stale — these don't). Read once, apply everywhere.

**Connect once, hold it — never reconnect per step.** Any tool that talks to a robot repeatedly
within one run (a step sequence, a poll loop, a reactor watching for events) should open ONE
connection at the start and reuse it for the whole run, not reconnect fresh for every single
action. A fresh connect is expensive — full handshake, plus (for animation-capable SDKs) a full
reload of the robot's own animation list every single time — and that overhead compounds fast
once actions come in close succession, causing real, live-confirmed failures (timeouts,
degraded response) under normal gameplay pace, not just in theory. Same principle as the golden
song engine's own "stay-put" mechanism (`run_continuous_hold_assume()`), proven there first, now
also applied to the bingo reactor after a live failure traced directly to reconnecting per draw
— see this doc's own 2026-08-12 findings entry and `SONG_102_BROBOTS_1_2_BINGO_GAME.md`'s
"Open" section for the incident. Reconnect-per-step is the default trap a first draft falls into;
stay-put is the fix, every time this shape of bug shows up again.

---

## Registry

Every alias/function in every file `.bashrc`/`.bash_aliases` actually source, grouped by
source file. Private helper functions (leading `_`, never called directly by the operator)
are named in a footnote under each table rather than given their own row — they're
plumbing, not keys on the board. **RETIRED** rows are still physically present
(commented out, not deleted) — pressing them today does nothing.

### `~/.bash_aliases` (entry point, two aliases defined directly)

| Alias | Resolves to | Line |
|---|---|---|
| `wpr` | **Two modes, added 2026-08-02; numeric shortcuts added 2026-08-23.** Bare `wpr` at an interactive terminal asks which one; `wpr check` / `wpr force` (the original word forms, still used internally — `gopod-song-open`'s own fill-work call passes `wpr check` explicitly) or `wpr 1` / `wpr 2` (matching the prompt's own numbering) skip the prompt. Any other argument prints a one-line `WPR_USAGE` note and does nothing, rather than silently falling through to a default — fixed 2026-08-23 after `wpr 2` was found live-broken: a non-empty `$1` already skipped the prompt, but the dispatch only ever matched the literal word `"force"`, so `2` silently fell through to the check path instead of forcing. **Restart check (default, `check`/`1`)** — routes through `restart_wirepod_preflight()`, the shared CHECK→CLEAR→START→CONFIRM sequence (`.claude/skills/wirepod-restart-discipline/SKILL.md`): skips the actual restart entirely if Wire-Pod is already healthy, which is what keeps repeated calls from ever tripping systemd's `start-limit-hit` guard. **Forced restart (`force`/`2`)** — the older unconditional behavior, `sudo systemctl restart wire-pod; sleep 2`, no health check, asks for the sudo password every time — as of 2026-08-04 the forced path also runs `sudo -k` first, so a still-valid cached sudo timestamp from an unrelated command can no longer silently skip the password prompt the menu promises; kept as an explicit second option, not the default. Not always sufficient alone: `robot_control_song_001`'s own story.md records `wpr` alone failing to clear a stuck connection; the proven recovery order is power-cycle both robots → re-pair with Wire-Pod → `wpr`. If the limiter has already tripped, recovery needs `sudo systemctl reset-failed wire-pod.service` first (needs an interactive terminal for `sudo` — the operator runs this, not Claude Code's sandboxed shell). Lives in `~/.bash_aliases` directly, not `~/.gopod_alias_lib/brobots.sh` — no repo overlay copy exists for `.bash_aliases`, so this alias has no mirrored backup file the way the `.gopod_alias_lib/*.sh` set does | `.bash_aliases:15-83` |
| `wpu` | **Wire-Pod update, added 2026-08-04.** Operator's own exact command sequence, wrapped as-is, no health check and no CHECK/CLEAR/CONFIRM discipline (not a restart-safety alias, a plain wrapper): `sudo systemctl stop wire-pod` → `cd ~/wire-pod && sudo ./update.sh && sudo ./setup.sh daemon-enable` → `sudo systemctl start wire-pod`. Pulls/builds whatever `update.sh` pulls from wire-pod's own upstream — distinct from `wpu`'s near-namesake `deploy_overlay_build.sh` (GOPOD-repo overlay → live binary, no service stop/start, no upstream pull), which stays a separate manual script, not aliased | `.bash_aliases:70-77` |

### `core.sh`

| Alias | Resolves to | Line |
|---|---|---|
| `gorepo` | `cd ~/crushn8r_git/GOPOD` — the most-used key on the board; several aliases below assume it's already been pressed | `core.sh:2` |
| `gp` | `git pull --rebase && git push` | `core.sh:3` |
| `codex-1`…`codex-6` | Launch Codex with an isolated `$CODEX_HOME` per number, for running several isolated Codex sessions side by side | `core.sh:6-11` |
| `nmg` / `gopod-devices` | `nmcli` device/type/state/connection summary (same alias, two names) | `core.sh:14,16` |
| `nmc` | `nmcli con show --active` | `core.sh:15` |
| `gopod-net` | Runs the auto Wi-Fi/USB network recovery script | `core.sh:17` |
| `gopod-net-usb` | Same recovery, pinned to the USB Wi-Fi adapter | `core.sh:18` |
| `gopod-net-fix` | Runs the general net-fix script | `core.sh:19` |
| `gopod-scan-usb` | Wi-Fi scan over the USB adapter | `core.sh:20` |
| `gopod-mic-detect` | Finds the USB mic via `audio.json`'s match string, prints it, changes nothing | `core.sh:29` |
| `gopod-mic-set` | Runs mic-detect, sets it as default input, confirms via `pactl info` | `core.sh:74` |
| `gopod-mic-test` | Real proof, not a device-name check: records ~3s, reports the actual peak sample amplitude, so a silently-dead/suspended device shows `peak=0` instead of a false pass | `core.sh:90` |
| `gopod_brobots` / `gopod-opening-chord` | The one-key stage-set: mic-set + LLM warm (`gemma2:2b`) + Kokoro warm-up + Wire-Pod restart/wake, concurrently, honest READY/NOT-READY. As each robot wakes it speaks its own ready line ("Brobot 1, ready for the interview." / "Brobot 2, ready to interview."), and the conductor holds — does not report ready — until every note has truly landed, spoken lines included. Only then do both robots attempt "Brobots ready!" at the same time, fired concurrently. **Honest caveat:** the Wire-Pod say path is fire-and-forget — a say call returns in well under a tenth of a second, before the robot is actually done talking — so the "together" attempt is proven simultaneous at the command-timing level, not guaranteed simultaneous in the air. **Confirmed 2026-07-16** (`CHORD_ABSORBS_PREDEMO_001.md`): its existing Wire-Pod-restart (service-active + HTTP-ready, retry-polled) and per-robot (`conn_test` + real spoken ready line) steps already fully cover every live-hardware check the retired `gopod-pre-demo` used to run — no code change made, none needed | `core.sh:288` (alias at `core.sh:406`) |

Private helpers in `core.sh`: `_gopod_chord_kokoro_job`, `_gopod_chord_wirepod_job`,
`_gopod_chord_release_both`, `_gopod_chord_direct_together_job`.

Kokoro voice output has no dedicated key of its own — it's wired straight into the
interview runner, warmed up automatically the moment `start-the-interview` is pressed,
before the first line ever plays.

### `brobots.sh`

| Alias | Resolves to | Line |
|---|---|---|
| `wirepodpulse` | SSH to `gopod-laptop`, restart `wire-pod.service` there — the remote/laptop-side twin of `wpr` | `brobots.sh:12` |
| `brobots-happy` | **PINNED WORK IN PROGRESS.** Doc → `happy`, Pip → `celebrate`, via SSH cue pair targeting `/api-sdk/play_animation` — confirmed 2026-07-16, read-only from `server.go`'s route switch plus a live 404 check, that this endpoint is not a real route (falls through to the `default:` 404 case). `brobots.sh`'s own comment beside `_brobots_move_axis` already named this exact finding, written to justify *not* copying `play_animation`'s convention elsewhere. `gopod-laptop` itself was unreachable this session (`No route to host`) so the exact remote instance wasn't independently re-checked — but the route's absence is a fact about the shared Wire-Pod fork's source, not the machine running it. Not fixed or deleted, per operator instruction — pinned as broken-as-shipped | `brobots.sh:58` |
| `brobots-angry` | **PINNED WORK IN PROGRESS**, same finding as `brobots-happy`. Doc → `angry`, Pip → `frustrated`, same shape | `brobots.sh:59` |
| `brobots-lift-up` / `brobots-lift-down` | Real movement notes, not animation cues: assumes behavior control on both robots, sends `/api-sdk/move_lift` at a fixed speed (the same endpoint the web control page's R/F keys drive — traced read-only from `control.js`/`server.go`), holds for a duration (default ~1.2s, overridable by a first argument), then zeros the speed and releases. `move_lift` is continuous-speed, not move-to-position, so the note has to start it, hold it, and explicitly stop it. Talks to Wire-Pod directly (`GOPOD_WIREPOD_BASE_URL`, default is the local Wire-Pod service address — see local machine config, no network identifiers in tracked files), not the SSH-to-`gopod-laptop` shape `brobots-happy`/`brobots-angry` use. **2026-07-16:** now via `_gopod_note_send` (shared Python instrument), not curl — same alias name/behavior, live-verified | `brobots.sh:158-159` |
| `brobots-head-nod` | The "silent yes": one small head-down-then-up sequence via `/api-sdk/move_head` (the web control page's T/G keys), same assume/hold/stop/release shape as the lift notes, packaged as a single self-contained note (default ~0.35s per half, overridable). Same 2026-07-16 instrument conversion, live-verified | `brobots.sh:163` |
| `brobots-anim-happy` / `-very-happy` / `-sad` / `-very-sad` / `-angry` / `-frustrated` / `-confused` / `-thinking` / `-celebrate` / `-love` | One note per confirmed `animation_vocab.json` token (`happy`/`veryHappy`/`sad`/`verySad`/`angry`/`frustrated`/`confused`/`thinking`/`celebrate`/`love` — the 4 knowledge-graph tokens — `answering`/`searching`/`searchingGetout`/`kgSuccess`, now also `verified:true`, below — are deliberately excluded, reachable only via `brobots-anim-test`/`-test-all`). Each fires a silent, speech-free `/api-sdk/say_text` call (`text={{playAnimationWI||<token>}}` — confirmed to parse into a bare animation action with no spoken text at all) through the shared `_brobots_play_anim` helper: assume control on both robots, fire, hold ~2.5s (overridable), release. Same 2026-07-16 instrument conversion, live-verified via `-thinking` | `brobots.sh:218-227` |
| `brobots-anim-test <token> [robot] [hold]` | One-robot single-token fire, for judging an unverified/experimental `animation_vocab.json` token without committing both robots to it — same assume/`say_text`/release shape as `_brobots_play_anim`, just one serial and a real token argument instead of one baked into the alias name. Added 2026-07-24 for the 4 knowledge-graph tokens; first pass guessed 2 real clip names (`answering`/`searching`) and 2 wrong ones (a `getin`/`listening` pair that don't exist anywhere in Wire-Pod's own source) — corrected same day by reading `kgsim.go` directly: `answering` (`anim_knowledgegraph_answer_01`), `searching` (`anim_knowledgegraph_searching_01`, loops while the KG/LLM is searching), `searchingGetout` (`anim_knowledgegraph_searching_getout_01`, the real transition-out clip, not a bare "getin"), `kgSuccess` (`anim_knowledgegraph_success_01`, a real name present in `kgsim.go` but currently commented out/unwired there) — all four fired as no-ops on the first two live rounds (`wire-pod.service` hadn't been restarted since before the vocab edits; the running process caches `animation_vocab.json` at startup, so every fire silently fell back to `thinking`), then confirmed real and working 2026-07-24 after a `wpr` restart — all four now `verified:true`. **Loop mode, added same day**: `searching`/`answering` are real loops in `kgsim.go` (re-fired every ~1/3s for as long as that phase lasts, not a single play held open) — `_brobots_anim_is_loop_token` re-fires those two every ~0.333s for the hold duration instead of firing once and sleeping; `searchingGetout`/`kgSuccess` stay one-shot, matching `kgsim.go`. `robot` is `1` or `2` (default `1`), `hold` defaults to `2.5`s. Example: `brobots-anim-test searching`. Backed by `_brobots_anim_is_loop_token`/`_brobots_play_anim_single` | `brobots.sh:239-298` |
| `brobots-anim-test-all [robot] [seed_hold]` | Runs all 4 knowledge-graph tokens back to back on one robot, 1s gap between each, in `kgsim.go`'s own real sequence order: `searching` → `searchingGetout` → `answering` → `kgSuccess`. **Interactive per-animation hold, added 2026-07-24** at the operator's own request: prompts before each fire ("Hold seconds for `<token>` (Enter for `<N>`s): "). The first prompt's typed value (or `seed_hold`/`2.5` if left blank) becomes the standing default for every later blank Enter; typing a number at any later prompt is a one-off override for that one fire only — it does not change the standing default. Example: typing `5, Enter, Enter, Enter` fires all four at 5s; `5, Enter, 3, Enter` fires `5s, 5s, 3s, 5s` (the `3` overrides only its own turn). `searching`/`answering` loop-fire (~every 0.333s) for the hold duration per `_brobots_anim_is_loop_token`, above; `searchingGetout`/`kgSuccess` fire once. No new dispatch mechanism — still a sequencer over `_brobots_play_anim_single`/`brobots-anim-test` | `brobots.sh:311-329` |
| `kgSuccess` — pinned standalone, for later use | Real clip (`anim_knowledgegraph_success_01`), `verified:true`, confirmed twice: once as the 4th beat in a full `brobots-anim-test-all` run ("Success. Longer time gaps.", operator's own report), once fired alone via `brobots-anim-test kgSuccess` (clean HTTP dispatch, same code path — not independently eyes-on-watched standalone, only as part of the sequence). Real name is present in `kgsim.go` but currently commented out/unwired there — Wire-Pod itself never fires it in production. No song wires it in today; noted here as a confirmed, ready-to-use "done!"/success reaction for whenever a future song or sequence wants one | `brobots.sh:239-298` (`_brobots_play_anim_single`) |
| `brobots-check` | Self-test: confirms `brobots`/`happy-brobots`/`angry-brobots` still exist — **known-broken**, those three are retired below, so this can never print `BROBOTS_ALIAS_CHECK_PASS`. A known-broken self-test sitting on the board, not a phantom regression if it ever comes up again | `brobots.sh:229` |
| `brobots-grep` | Greps every alias/function definition across `.bash_aliases` + `.gopod_alias_lib/*.sh` — a live, un-cached self-inspection of the whole board | `brobots.sh:242` |
| ~~`start-the-interview`~~ | **Retired 2026-08-17** (`PHA0B_INTERVIEW_CONSOLIDATION_EXECUTED_001.md`, operator: "reduce to one golden process"). Same one-shot generate+perform call now lives behind `pha0b` → pick Interview → `g`, or `pha0b interview` directly — see `pha0b`'s own row below | retired |
| `interview-json` | **Renamed from `start-the-preshow` 2026-08-19** (`NAMING_APPLIED_001.md`, operator naming decision — a survey pass first confirmed this function already IS "generate-only, no playback, no vamp," so this is a pure rename onto unchanged code, not new behavior; old name retired in place, commented out with a pointer, not deleted). Compose-only stage, still live standalone: `run_section1_preshow_generate_001.py`'s `generate_phase()` — writes the JSON run log, no robot speaks. `gopod_warm_up` (row below) also calls this directly | `brobots.sh:595` |
| ~~`gopod-preshow-then-interview`~~ | **Retired 2026-08-17**, same pass as `start-the-interview` above — its one-shot end result is what `pha0b`'s own interview `g` choice now covers | retired |
| `gopod_warm_up` | `gopod_brobots` (required) + `interview-json`'s body (renamed from `start-the-preshow` 2026-08-19). No robot speaks an interview line here either | `brobots.sh:654` |
| ~~`gopod_interview`~~ / ~~`--full`~~ | **Retired 2026-08-17**, same pass as `start-the-interview` above. `gopod_warm_up` (row above) keeps `_gopod_require_brobots`'s only remaining live caller — the multi-frame call-stack scenario this row's own guard behavior described no longer has a live caller | retired |
| `interview-vamp` | **Added, doc gap closed 2026-08-19; renamed from `vamp-run` same day** (`GOPOLISHER_FIXES_001.md`, operator naming decision — old name retired in place, commented out with a pointer, not deleted). Rolls a fresh interview take: fires `run_preshow_song()` (the backstage banter) WITH interview generation (`generate_phase()`) running alongside it in a background thread — the vamp fills the live generation wait. What `pha0b`'s own interview `v` choice calls under the hood, and one of `interview-run`'s own two reused mechanisms (row below). `$1`, if given, overrides `GOPOD_SECTION_SONG_DIR` (the generating song's own folder, default `02_brobots_interview_run` via `DEFAULT_SECTION_SONG_DIR`); the vamp/preshow path itself is now always the standalone `01_brobots_interview_vamp` folder unconditionally, no longer derived from `$1` (restructured 2026-08-19's vamp/run split, the old `"$1/vamp"` nesting convention broke once vamp stopped being a subfolder) | `brobots.sh:724` |
| `interview-replay` | **Renamed from `interview-run` 2026-08-19** (`NAMING_APPLIED_001.md`, operator naming decision — freed the `interview-run` name for a new meaning below; this function's own body is unchanged, only the name moved, no retirement stub since the old name is reused, not dropped). Replays the last generated interview take — performance only, no generation. What `pha0b`'s own interview `p` choice calls under the hood, and both of `interview-run`'s own three branches (row below) end with a call to this | `brobots.sh:777` |
| `interview-run` | **New meaning 2026-08-19** (`NAMING_APPLIED_001.md`, operator naming decision — the name freed up by the `interview-replay` rename above). The interview itself (video 2), with an optional full-run mode that plays the vamp (video 1) first — orchestrates existing functions, no new playback logic. Two interactive prompts, no CLI flags: y/n "include the vamp first?" (default **n** — replay-only is the lighter, more common ask, so opting into the heavier vamp path requires an explicit yes); on **no**, calls `interview-replay` directly. On **yes**, a/b "(a) fresh vamp + fresh generation, or (b) vamp for atmosphere + replay the existing take?" (default **a**) — (a) calls `interview-vamp` then `interview-replay` (the take `interview-vamp` just generated IS the newest file `interview-replay`'s own `ls -t` glob reads, by construction — no log-path plumbing needed between the two calls); (b) calls `interview-vamp-play` then `interview-replay` (the vamp plays for atmosphere only, triggers no generation, so `interview-replay` plays back whatever take was already on disk beforehand, unchanged by the vamp call) | `brobots.sh:821` |
| `interview-vamp-play` | **Added 2026-08-19** (`INTERVIEW_VAMP_NO_GEN_PATH_001.md`); **renamed from `preshow-run` same day** (`GOPOLISHER_FIXES_001.md`, operator naming decision — resolved to avoid a real collision with `interview-run`'s own then-existing "video 2, no generation" meaning; old name retired in place, commented out with a pointer, not deleted) — the pure "play video 1" button: fires `run_preshow_only()`, which reuses `run_preshow_song()` verbatim but pre-sets the generation-done event first, so **zero interview generation is ever triggered** (`generate_phase()` is never called) — a sibling to `interview-vamp`, not a replacement; `interview-vamp` still exists unchanged for when a fresh take actually needs rolling, and this function is also one of `interview-run`'s own two reused mechanisms (row above). Still reads the shared `interview_scaffold` for Brobot 1/2's two `llm_coloured` wake-beat lines (persona/pronunciation consistency — a read, not generation). Defaults to `01_brobots_interview_vamp`. `pha0b_menu()`'s `01_brobots_interview_vamp)` case-arm calls this, not `interview-vamp` — picking the vamp song from the bare `pha0b` menu means "play video 1," not "roll a new take" | `brobots.sh:960` |
| `_gopod_check_audio_routing` | **Added 2026-08-19** (`AUDIO_ROUTING_CHECK_001.md`/`_002.md`) — startup gate, called first thing inside `interview-vamp-play` (VAMP is its testbed). Confirms PulseAudio's default sink/source are the real GOPOD devices (mic = USB Audio Mono, speakers = Built-in Audio Analog Stereo) via `pactl get-default-sink`/`get-default-source`, since Wire-Pod logs `say=success` even when audio is misrouted and nothing is heard (a remote NoMachine session grabbing the defaults is the confirmed common cause). On mismatch: warns with the exact `pactl set-default-*` fix command, prompts before touching anything (`_pha0b_prompt_yn`, default no), never auto-forces. Non-blocking — the caller proceeds regardless, now informed instead of finding out after a silent run. The two expected device names are config-driven: hardcoded as this Jetson's own fallback default, overridden if `gopod_audio_config.sh` (this file's own sibling, real per-machine values, never committed — see `gopod_audio_config.example.sh` for the template) exists and is sourced | `brobots.sh:882` |
| `start-the-control-song` | **Cut over 2026-08-13** (studio tuning cut 2, `CONTROL_SONG_LOOP_RETIRED_001.md`) from the legacy `run_robot_control_song_001.py` full-song loop — now retired outright — to the golden song engine (`run_golden_song_001.py`), already golden-registered as `robot_control_song_001` (`SONG_REGISTRY`, commit `16aa54e`). **Live-confirmed on real hardware.** Fires LIVE by default now — exports `GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1` itself, no separate export step needed, same convention `start-the-interview` uses. Full song on one robot (default Brobot 1, pass `2` for Brobot 2, robot pick now via `GOPOD_GOLDEN_ROBOT`): connect, say connected, arm test, head nod, a real weather fetch (now sourced live from that robot's own Wire-Pod jdocs `RobotSettings`, not a static lookalike file — see `gopod-weather-say` below), say good, exit — every physical note still self-narrated. `run_robot_control_song_001.py` itself is not gone — it survives as a helper/tester library only (`test-arm-cue`/`test-head-nod`/`test-fireworks` below still route through its `run_single_note()`, and the interview runner still imports its gesture functions); only its own full-song dispatch loop was retired | `brobots.sh:671` |
| `test-arm-cue` / `test-head-nod` / `test-fireworks` | Isolated single-note testers: fire exactly one motion, no song loop, no other notes' assume/release churn in the way. Built for tuning one motion at a time. Same `2` argument to target Brobot 2. **Unaffected by the archive move, 2026-07-24** — all three route through `run_single_note()` → a bare note function (`gentle_arm_test_cue`/`head_nod_test_cue`/`fire_fireworks`) that never loads `robot_control_song_001`'s own folder at all (no `knobs.json`, no `DEFAULT_SONG_DIR`). `test-fireworks` confirmed directly (dry then live, HTTP `200 done`); `test-arm-cue`/`test-head-nod` confirmed by the same code path, not independently run. Only `start-the-control-song` (the full scored song) actually needs that folder and is genuinely still broken — the earlier "all four broken" note was a blanket statement that never checked this distinction. **`test-arm-cue`/`test-head-nod` gained their own "apply phcal tweaks to this test cue?" y/n prompt 2026-07-25** (operator request, alias-mixer widening) — on `y`, calls `phcal_apply_control_song_001.py --yes --target test --primitive arm\|nod`, writing phcal's confirmed `hold` into `ARM_TEST_LEG_HOLD_SECONDS`/`NOD_TEST_LEG_HOLD_SECONDS` only (never speed, never the GESTURE constants the interview's own live movement uses) — see `.claude/skills/alias-mixer/SKILL.md` §2 | `brobots.sh:481,486` |
| `gopod-fireworks` | **New 2026-07-16.** Human-named front door onto the exact same fireworks note `test-fireworks` already plays (`/api-sdk/cloud_intent`, `intent_seasonal_happynewyear`) — a plain wrapper, not a second implementation | `brobots.sh:381` |
| `gopod-weather-say` | **New 2026-07-16.** Standalone version of the control song's own single-robot `weather` note: a real Windsor fetch, formatted per that robot's own unit/clock (Brobot 1: Celsius/24-hour, Brobot 2: Fahrenheit/12-hour), spoken once. `run_single_note()` gained a `weather` branch to make it isolable, mirroring `arm_test`/`head_nod`/`fireworks`; no song's own knobs/story touched | `brobots.sh:394` |
| ~~`start-the-bait-song`~~ | **Retired 2026-08-11** (stale, no crucial use found beyond documentation, operator: "easily rebuilt"). Called the legacy `run_robot_control_song_001.py` directly while `pha0b`'s own `bait` case cut over to the golden engine 2026-08-07. Use `pha0b` → pick `00_brobots_awaken` instead — see that row above | retired |
| `start-the-net-song` | Interview's own short capture cut, `brobots_bait_001` — the operator's own "net" name for it, distinct from the bait/awaken song's "bait." Both robots wake (arm cue, head nod), each speaks its own self-naming line, done — no LLM, no vamp, no handoff, under a minute. Built on the interview engine (`GOPOD_SECTION_SONG_DIR`, the same song-dir seam `brobots_interview_section_01` uses), not the control-song family — structurally can't be `pha0b`-sliced, same reason `interview`/`preshow` can't (no standalone step-loop runner for this line-based shape). **Repointed 2026-07-24** to `zzz_archives/brobots_bait_001/` (moved there in the operator's own manual song-folder cleanup). Live by default | `brobots.sh:483` |
| `bingo-video-song` | Bingo capture song for the upsell video (`songs/101_brobots_bingo_test/`) — **not** the live 75/90-ball game (`gobingo`, below) or its reactor, both untouched. 69-step scored back-and-forth (`speaker` switches per step, was 57 before 2026-08-11's notation remap), three paced emotion beats plus a rattle+call+reaction round in front of each. Runs on the golden song engine (`run_golden_song_001.py`, cut over 2026-08-07). **WIP, 2026-08-11 — golden lock broken by the remap, re-locks only after a fresh live run confirms the new shape** — see `songs/101_brobots_bingo_test/story.md`. `start-the-bingo-capture` (a duplicate name for this same run) and `bingo-video-song-pick-segment` (superseded by `pha0b bingo`'s own segment picker) retired the same day, no crucial use found | `brobots.sh:822` |
| `bingo-video-song-live` | **Drift catch, 2026-07-23 (`.bashrc`/`.bash_aliases` cross-check).** Same run as `bingo-video-song` above, live by default — exports `GOPOD_ALLOW_LIVE_ROBOT_SPEECH=1` itself, no separate export step, same pattern `start-the-interview` already uses | `brobots.sh:507` |
| `test-interview-movements` | Rehearses every scored movement in the real interview score (`songs/02_brobots_interview_run/knobs.json`), real playback order, both robots, with placeholder speech instead of generated lines — no LLM, no generation wait | `brobots.sh:474` |
| `gopod-conn-test` | **New 2026-07-16.** Standalone version of the opening chord's own per-robot wake check (`/api-sdk/conn_test`, the same call `_gopod_chord_wirepod_job` fires) with no restart/wake/speech attached — just "is this robot reachable right now." Argument `1`, `2`, or `both` (default `both`) | `brobots.sh:458` |
| `gopod-vamp` | **New 2026-07-16.** Standalone preview of the pre-show's own vamp gate: the scored `vamp_1..vamp_4` filler beats `run_preshow_song` loops while interview generation is in progress. Before this alias, the only way to hear these lines was to run the whole pre-show song and hope generation was still in flight when the gate opened. Calls the exact same loader and speak function the real vamp loop calls (`load_preshow_song()` / `_preshow_speak_host()`) directly, no fake generation-done event, no robot/scaffold setup needed. Argument: number of full cycles through the four beats (default 1) | `brobots.sh:505` |
| `brobots_vamp_gate` (via `pha0b vamp`) | Standalone, playhead-sliceable *song* of the same four vamp beats `gopod-vamp` previews above — `run_vamp_gate_song_001.py`, reusing `load_preshow_song()`/`_preshow_speak_host()` unchanged. A fixed step list, once through (not the real generation-gated loop) — built so the vamp gate can be evaluated against every other GOPOD song on equal footing. No physical robot, so no live-speech gate: it always actually speaks (Kokoro audio, no hardware risk); `GOPOD_VAMP_GATE_READ_SHEET=1` gives a text-only dry preview instead. **Repointed 2026-07-24**: moved into `zzz_archives/brobots_vamp_gate/` in the operator's own manual song-folder cleanup — the runner's own `DEFAULT_SONG_DIR` follows it there; no longer `pha0b_menu`-visible (menu shows top-level song dirs only) but still reachable via direct `pha0b vamp <a> <b>`. Its own doc moved alongside it, off the grid on purpose — `goverlord/runtime/songs/zzz_archives/brobots_vamp_gate/BROBOTS_3_4_VAMP_GATE.md`, no longer cross-linked from the main doc set | `goverlord/runtime/songs/tools/run_vamp_gate_song_001.py` |
| `gopod-pick-model` | **New 2026-07-16.** Standalone front door onto `resolve_content_model()`, the LLM lane's live, remembered, no-filtering model picker (see `LLM_MODEL_SELECTOR_001.md`). Same `GOPOD_CONTENT_MODEL` env var skips the interactive menu unattended; same remembered-model state file (`goverlord/runtime/songs/02_brobots_interview_run/content_model_state.json` — moved here 2026-08-19's vamp/run split, no longer under a `vamp/` subfolder) | `brobots.sh:541` |
| `test-silent-angry-say` | Reaction-lane test alias, fixes the original `test-silent-angry-say` crash by swapping the payload's animation token from the blocking `{{playAnimation\|\|angry}}` to the async `{{playAnimationWI\|\|angry}}`. Full wake step (`conn_test` + 1.5s settle) → assume → say_text → `sleep 5.0` hold → release, against Brobot 2 by default. Extensive live-tuned history in-file (first fire produced no visible animation, fixed by adding the wake step; hold retuned to the operator's own proven 5.0s value after an unreliable 18.8s calculated figure was explicitly discarded) | `brobots.sh:596` |
| `test-concurrent-reaction` | Two-robot concurrency test: Brobot 1 backgrounded, count + 7s hold; ~2s later Brobot 2 fires `{{playAnimationWI\|\|angry}}` with a 5.0s hold — both via `_gopod_note_send`, PASS/FAIL summary via `grep -c "NOTE_HTTP"` (unanchored — the `^NOTE_HTTP` anchor broke silently once timestamps were added to the log line, fixed by dropping the `^`) | `brobots.sh:641` |
| `test-reaction-pick-animation` | The animation picker: reads `animation_vocab.json` live via `python3 -c`, filters to `verified:true` (10 of the 11 total entries — `GOPOD_ANIM_TODO` excluded), presents a numbered menu, flags `angry` inline with its crash history note, then calls `test-reaction-in-the-beat` with the chosen token. Built per the operator's own request: "reads the playAnimation vocab json file... provides a numbered list to choose from... chain that to this test" | `brobots.sh:911` |
| `test-reaction-in-the-beat` | The most-evolved reaction note, parameterized by animation token (default `frustrated`). Sequence: wake both (conn_test ×2 + 1.5s settle) → Brobot 1 speaks "Animation test run" (hold_phrase 3.0s) → Brobot 2 speaks its emotion phrase ("I'm sad", "I'm angry", etc., a per-token case map) → **2-second pause** → separate bare animation dispatch (`{{playAnimationWI\|\|<token>}}`) → stuck-animation check (curls `get_debug_logs`, greps `"waiting for animation to be done"`) → hold_anim (5.0s) → release → Brobot 1 speaks "Run Complete" (hold_phrase) → release. Redesigned per the operator's own explicit shape: "Robot 1 'Animation test run', robot 2 part, robot 1 'Run Complete'. No counting. Too long." **This is the golden dispatch shape** — see [Today's findings, 2026-07-16](ALIAS-LIBRARY-FINDINGS-ARCHIVE.md#todays-findings-2026-07-16) | `brobots.sh:791` |
| `test-angry-hold` | Standalone single-note tester: fires the `angry` token alone with a configurable hold (`${1:-5.0}`, `--`-prefix stripped) against a configurable serial (`${2:-0dd1d8bf}`) — wake, assume, say_text, hold, release. Built for isolating hold-duration effects on `angry` specifically, separate from the full picker/beat sequence | `brobots.sh:969` |
| `test-anim-searching` / `test-anim-answering` / `test-anim-kg-success` | **Golden notes, added 2026-07-28:** one KG animation token fireable in isolation, one alias each, `<alias> [robot: 1\|2, default 1] [hold seconds, default 1]` — so the operator can watch exactly one thing and judge it with his own eyes, no chaining, no sequence. Requested as a mirror of `test-arm-cue`/`test-head-nod`'s shape, but those two route through `run_single_note()` in `run_robot_control_song_001.py`, which has no code path for a `playAnimationWI` token at all and fires no `conn_test` wake step of its own — so the ALIAS shape (isolated dispatch, robot argument, hold-override argument, `_BLOCKED`/`PASS`/`FAIL` reporting) is mirrored, but the underlying mechanism is ported from this same file's own proven animation-dispatch precedent instead: `conn_test` → 1.5s settle (`test-silent-angry-say`/`test-angry-hold`'s own fix for the live "HTTP success, no playback on cold first press" defect) → assume → dispatch → release. `_brobots_anim_is_loop_token` (unchanged) still decides loop vs. one-shot (`searching`/`answering` loop every ~0.333s for the hold duration, `kgSuccess` fires once) — but the loop's own dispatch-count formula is ported straight from `run_songs_runner_001.py`'s `run_animation_only()` accumulator (`elapsed = 0.0; while elapsed < hold_seconds: dispatch; sleep(0.333); elapsed += 0.333`), **not** delegated to this file's existing `_brobots_play_anim_single`, whose own `repeats=int(hold/0.333)` formula is off by one dispatch from bingo's real, live-proven count (confirmed live 2026-07-28: at `hold=1.0s`, `_brobots_play_anim_single` computes `repeats=3`; bingo's own accumulator produces 4). Live-verified against a local mock listener (never real Wire-Pod): default `test-anim-searching` (hold=1s) → exactly 4 `say_text` dispatches of `{{playAnimationWI\|\|searching}}`, matching bingo's own `round_3_searching` step (`hold_seconds: 1.0`) exactly; `test-anim-answering 2 2.5` → exactly 8 dispatches of `{{playAnimationWI\|\|answering}}`, matching bingo's own `hold_seconds: 2.5` answering steps exactly (the family's uniform 1s *default* alone only gives 4 for `answering` — pass `2.5` as the hold argument to reproduce bingo's proven count); `test-anim-kg-success` → exactly 1 dispatch of `{{playAnimationWI\|\|kgSuccess}}` (one-shot, any hold). Blocks cleanly on a bad robot (`TEST_ANIM_<TOKEN>_BLOCKED bad_robot`) or bad hold (`..._BLOCKED bad_hold`) argument, firing nothing — confirmed live, no HTTP call ever attempted on either bad-input path | `brobots.sh:3599-3665` (private helper `_test_anim_isolated`) |
| `test-anim-token <token> [robot] [hold]` | Generic counterpart to the three fixed-token aliases above, added 2026-08-12 — same `_test_anim_isolated` mechanism (`conn_test` wake step, dispatch-count-precise loop/one-shot via `_brobots_anim_is_loop_token`, `PASS`/`BLOCKED` HTTP verdict), but the token is a real argument instead of baked into the alias name, mirroring `brobots-anim-test`'s own already-generic shape while routing through this family's newer, dispatch-count-precise mechanism instead of `_brobots_play_anim_single`. Built so any `animation_vocab.json` token (e.g. `dartingEyes`, restored to the vocab this same pass) gets this family's proven precision without needing its own named alias first. Blocks with `TEST_ANIM_TOKEN_USAGE` on a missing token argument, firing nothing | `brobots.sh:3693-3699` (calls the same private helper) |
| `pha0b` | **PlayHead A/0/B — the studio-wide song-slice cockpit.** Bare `pha0b` (no args) opens `pha0b_menu()`: pick a song off disk, pick Point A/Point B by division number (a forgiving clamp, not a hard block, on numeric-but-out-of-range input — `PHA0B_CLAMP` prints only when a clamp actually changed the typed value), then it runs the picked slice directly — no separate copy/paste step. The `0.` menu line (`0. Keyboard; 0 or number > song range, Enter, or Space Bar = Full song`) sits at the TOP of the divisions list; typing the literal `0` at the Point A prompt is a one-keystroke full-song shortcut that skips the Point B prompt entirely (empty Enter/space/out-of-range at Point A still fall through to Point B as before). An additional y/n prompt ("apply reporter gaps to this selected range?") fires for `song=bingo` or `song=bait` (widened from bingo-only 2026-07-25, same pass as the phcal-apply widening below, `REPORTER_GAP_WIDENED_TO_BAIT_001.md`) and sets a run-scoped `GOPOD_BINGO_REPORTER_GAP_OVERRIDE` env var (name kept as-is despite no longer being bingo-only) — `n` zeroes reporter-gap pauses for that run only, never rewriting `knobs.json`; `y` asks a second numeric question, "reporter gaps in seconds? [default = 0]" (Enter/empty = `0`, integer or decimal = that value, non-numeric = one re-ask then falls back to `0`), and the override reaches both runners as any float (`REPORTER_GAP_NUMERIC_ENTRY_001.md`, 2026-07-25 — replaced the old y/n-only prompt whose "5s"/"7s for bait, 5s for bingo" wording never matched disk truth) - **studio rule as of 2026-07-25: every reporter gap defaults to `pause_seconds: 0` in every song's own `knobs.json`, no exceptions** (both `bingo`'s and `bait`'s current values are `0`, matching), left open for a later edited-in reporter voiceover, never a live dead-air pause - see `.claude/skills/alias-mixer/SKILL.md` §2 for the full rule. A separate y/n prompt, "apply phcal tweaks to this selected range?", fires for `song=bingo` or `song=bait` (widened from bingo-only 2026-07-25, operator request, `REPORTER_GAP_SHARED_SWITCH_SURVEY_001.md`) — on `y`, walks every `arm_cue`/`nod`/`head_nod` step in the picked range and writes phcal's last-confirmed `cycles`/`hold_seconds`/`speed` (`phcal_last.json`) straight into that song's own `knobs.json` via `phcal_apply_001.py --knobs <song's knobs.json>` (renamed from `phcal_apply_bingo_001.py` 2026-08-07 once the old bingo-only name got confusing enough to fix — every caller updated the same pass). `control` still gets no phcal-apply auto-fill prompt at all (the `case "$song" in bingo|bait)` gate above simply doesn't include it) — **correction, re-verified 2026-08-13**: this does NOT mean its `arm_cue`/`head_nod` notes run fixed test-sequence choreography instead. It runs on the golden engine now (see `start-the-control-song` above) and dispatches `arm_cue`/`nod` through the exact same tunable `run_arm_cue()`/`run_nod()` every other golden-engine song uses, reading straight from its own `knobs.json` — only the *auto-apply-phcal_last.json-into-knobs.json* convenience prompt is unavailable for it, never the tunability itself; the fixed test-sequence functions (`gentle_arm_test_cue`/`ARM_TEST_SEQUENCE`) are reserved for `test-arm-cue`/`test-head-nod`'s own isolated testers alone. (The `weather` keyword itself was dropped 2026-08-16 — dead reference to `zzz_archives/brobots_bait_000`, decluttered off disk 2026-08-15, no live equivalent — same purge as `start-the-weather-song`, `README_NOTE_AND_WEATHER_SONG_PURGE_001.md`.) Called with explicit args (`pha0b <song> <point_a> <point_b> [robot]`) it skips the menu and routes straight to the right runner, hiding the split between `GOPOD_BINGO_PLAYHEAD_*`/`GOPOD_CONTROL_SONG_PLAYHEAD_*` env vars. **Interview consolidated onto this same door, 2026-08-17** (`PHA0B_INTERVIEW_CONSOLIDATION_EXECUTED_001.md`) — `pha0b interview` (no `<a> <b>` needed, ignored with a printed note if passed) and `pha0b`'s own menu pick both now route through one shared function, `_pha0b_interview_bypass()` (`brobots.sh:1012`): vamp a take, perform the last take, or go one-shot (generate + perform now, the retired `start-the-interview`'s old job). **Naming resolved 2026-08-19** (`NAMING_APPLIED_001.md`) — the bypass's `p` choice now calls `interview-replay` (renamed from `interview-run`, which itself now means something new — see `interview-run`'s own row above); a separate, heavier `interview-run` alias exists outside this bypass for when the vamp-then-interview full-run flow is wanted, deliberately not folded into this lighter v/p/g menu. **Split into two menu-visible song folders, 2026-08-19** (`INTERVIEW_VAMP_SPLIT_001.md`) — the interview's own two-video split means `pha0b`'s bare menu now shows `01_brobots_interview_vamp` and `02_brobots_interview_run` as two separate picks, not one combined folder; picking the RUN folder routes into `_pha0b_interview_bypass()` as before (`02_brobots_interview_run)` case-arm, renamed from the old `01_brobots_interview_section_01)`), picking the VAMP folder routes to `interview-vamp-play` directly (renamed from `preshow-run` 2026-08-19, `GOPOLISHER_FIXES_001.md` — its own new case-arm, `01_brobots_interview_vamp)` — the pure no-generation "play video 1" button, not the bypass's `v` choice). **Live-robots gate, added 2026-07-25** (`LIVE_ROBOTS_PROMPT_DECOUPLED_001.md`): every song now prompts once, shared with `phcal` below via one function, `live_robots_prompt()` (`brobots.sh`) — "live robots? y/n [default y]:" — replacing the old split behavior where `bingo` alone hardcoded always-live and every other song stayed silently dry unless `GOPOD_ALLOW_LIVE_ROBOT_SPEECH` was pre-exported by hand. Default (Enter/`y`) reproduces the prior live behavior for every song; `n` runs dry. `vamp` (`zzz_archives/brobots_vamp_gate/` as of 2026-07-24, was top-level) has no physical robot at all, so it always actually speaks (Kokoro audio, no hardware risk) regardless of the live gate. `pha0b_menu` excludes `zzz_archives` itself from the disk-scan list (2026-07-24 — it's a container folder, not a song). See `PLAYHEAD_COCKPIT_ALIAS_001.md`, `PHA0B_RENAME_001.md`, and this file's own 2026-07-23/2026-07-24 findings below. **Menus consolidated onto shared choke points, 2026-08-17** (`PHA0B_MENU_CONSOLIDATION_001.md`): the song pick and the robot-filter pick now route through `_pha0b_prompt_range_choice()`/`_pha0b_prompt_choice()`, and the rich-display/phcal-apply/reporter-gap/robot-pick y/n prompts now route through `_pha0b_prompt_yn()` (all three defined just above `pha0b()`, right after `live_robots_prompt()`) — same prompts, same defaults, same behavior, just one function per shape instead of scattered copies; no nav/arrow-key change | `brobots.sh:1450` (`pha0b`), `brobots.sh:1957` (`pha0b_menu`) |
| `sleep-beat-*` / `sleep-rts-*` / `sleep-palm-*` / `pet-*` | 126 bench-test aliases, one per real sleep, ambient-rest, held-on-palm, or petting animation Vector ships with. Same fire-and-hold shape as every other test alias, one robot at a time. Not in any song's score yet — broll catalogue for `105_brobots_nap` (formerly `104_brobots_baby_robots_sleep`); see that song's own doc, [SONG_105_BROBOTS_1_2_BABY_ROBOTS_SLEEP.md](SONG_105_BROBOTS_1_2_BABY_ROBOTS_SLEEP.md)'s "The golden goody" section, for the public family breakdown | `brobots.sh:2460-2593` |
| `sleep-segment-core` / `-rts-off` / `-rts-on` / `-palm` / `-pet-triggers` / `-pet-clips` | The 6 batch runners over the 126 aliases above, grouped by family (18/27/27/27/10/17) — fires and reviews a whole theme as one command instead of 126 separate ones | `brobots.sh:2614-2685` |
| `robot-sleep` / `robot-wake` | Direct-SDK synced sleep/wake, no dry mode — `which` is `1`/`2`/`0`\|`both` (default `both`), `robot-sleep`'s second argument is hold seconds (default 5). Backs the golden-sequence openers below | `brobots.sh:2348,2360` |
| `robot-info` | Read-only diagnostic snapshot via `direct_sdk_robot_info_001` — no `BehaviorControl`, no movement, just a status dump. `which` is `1`/`2`/`0`\|`both` (default `both`, same convention as `robot-sleep`/`robot-wake`), 30s timeout | `brobots.sh:2380` |
| `brobots-wake` | Standalone-fireable wake: `_brobots_wake_core`'s golden-flag pulse (`conn_test` → assume → release → settle 2.5s → re-assume → settle 0.5s — the live-confirmed 2026-08-10 fix for a "nothing ever settled reassume→move" cold-fire bug) fired per robot, then a final release so it's never left holding control if fired alone. `robot_target` is `1`/`2`/`both` (default `both`) | `brobots.sh:231` (private helper `_brobots_wake_core`, `brobots.sh:217`) |
| `move-reverse` | Standalone-fireable fixed reverse pulse: assumes its own control (no `brobots-wake` prerequisite), fires `_brobots_move_reverse_core` at -150/-150 (the webroot's own confirmed S-key/backward value) via `move_wheels`, holds, zeros the wheels, releases. `hold` (default 0.5s — 0.3s proved too short to see) and `robot_target` (`1`/`2`/`both`, default `both`). Prints a standing caution every fire: cliff sensors disable while control is held, on-charger use only | `brobots.sh:275` (private helper `_brobots_move_reverse_core`, `brobots.sh:265`) |
| `nudge-reverse` | Composes `brobots-wake` + `move-reverse` into ONE continuous held-control window instead of two separate assume/release cycles — the wake half's own core doesn't release, the move half's own core doesn't assume, fixing what two earlier rounds of this mechanism got wrong (two windows, not one). Same `hold`/`robot_target` arguments and the same on-charger-only caution as `move-reverse` | `brobots.sh:317` |
| `tempo-set` | Standalone phcal-adjacent guided flow for Tempo Phase 2 — thin wrapper around `tempo_set_001.py`, the same tool `phcal`'s own `tempo` menu item (12) hands off to. Prints every song on disk as a numbered pick list (same disk-scan `pha0b_menu` already uses, `zzz_archives` excluded), then asks the two mode questions; every actual `knobs.json` read/write happens inside that one Python tool, not duplicated here | `brobots.sh:2188` |
| `gopod-song-open` / `gopod-song-open-chord` / `gopod-song-open-chord-sleep-first` | The "golden sequence" openers (operator-designed, 2026-07-25/26): both robots sleep in sync, real fill-work runs while they hold asleep (bare `wpr`; then mic/LLM/Kokoro warm-up; then a full Wire-Pod restart in the sleep-first version), and the instant that work finishes the robots wake together — event-driven, not a guessed fixed hold. `-sleep-first` adds a preflight sweep for a leftover sleep-binary process and a synced "Brobots ready!" finale after release | `brobots.sh:1781,1846,1937` |
| `score` / `score-save` | Bare notation-page printer for a song's score — `score` prints to screen, `score-save` writes it to that song's own `notation/` folder. Song keyword (default `bingo`) reuses `pha0b`'s own vocabulary: `bingo`/`control`/`bait`/`vamp` (`interview`/`preshow` refused — their `knobs.json` doesn't carry the note/TEXT fields the printer needs) | `brobots.sh:2090,2104` |
| `brobots-searching-out` | Holds the `searching` KG-loop token (kgsim.go's real holding-pattern token) for `hold` seconds under one continuous assume/release session, then closes it cleanly with one `searchingGetout` fire — `robot` (`1`/`2`, default `1`), `hold` (default 1s) | `brobots.sh:315` |
| `rehearse-searching-1s` / `rehearse-searching-2s` | Isolated `searching` KG-token fire via `_test_anim_isolated`, at a fixed 1s/2s robot slot respectively, first argument overrides the hold (default 1) | `brobots.sh:3713-3714` |
| `phcal` | **The calibration bench — Rung 3.** A sibling to `pha0b`: `pha0b` plays the SCORE (dry navigation of a song slice); `phcal` tunes the INSTRUMENT (one isolated mechanical primitive, live, on one robot). **Regrouped 2026-08-16** (`PHCAL_MENU_REGROUP_BUILT_001.md`) — bare `phcal` (no primitive) now opens an 8-group menu, `0. exit` plus: `1` info (`robot_info`, direct — read-only battery/version/protocol snapshot, no sub-menu), `2` movements (`arm`/`nod`/`move_reverse`, sub-menu), `3` vector's cube (`cube`, direct), `4` audio (`rattle`/`danger`, sub-menu), `5` animations (`animation`, direct — fires a real `animation_vocab.json` token: `kgSuccess`/`searching`/`answering`, or `0` for all three in sequence, with a `searchingGetout` auto-fired to end `searching`'s loop cleanly), `6` brobots 1 (`weather`/`brobots_announce_in_sync`/`brobots_stay_in_place` — this last one the renamed former "hold" primitive, assume/release with a printed countdown watching for autonomous drift — sub-menu), `7` brobots 2 (`brobots_sleep_to_wake_direct_sdk` — the renamed former "sleep_wake," both robots over one continuous connection, released after a set time or a real completed process — and `brobots_session_responsiveness`, sub-menu), `8` tempo (`tempo`, direct). Pure display regroup — every primitive's own identity string, `phcal_last.json` key, and dispatch behavior is exactly what it was before; only which numbered group/sub-menu a control displays under changed. Full per-group table: `PHCAL.md`. Every primitive with a tunable value walks it pre-filled with the last-used one (Enter keeps it, typing overrides), fires, and saves what was used to `phcal_last.json` — the same memory file a direct-flag call (`phcal arm 1 --hold 1.3 --cycles 3 --speed 2`, `phcal nod 2 3`) also reads from and writes to; widened 2026-08-16 (`MASTER_TWEAKS_STAGE1_SAVE_COVERAGE_001.md`) to 9 saved primitives, up from the original 4. The `rattle` primitive (added 2026-07-22) has a real, settable volume (UI 1-5, mapped to the direct-SDK binary's real 20/40/60/80/100), and releases Wire-Pod's own behavior-control grant before firing over an independent direct-SDK connection, settling `PHCAL_RATTLE_SETTLE_SECONDS` (3.0s, live-confirmed twice including with the robot asleep — do not lower without a fresh live re-confirmation) first. **`weather`, added 2026-08-02**, is not a mechanical primitive and stands apart from the rest: it's a plain API-only test — prompts for a location, runs it through the same comma-grouping geocode-candidate logic as `weather.go`'s `geocodeOpenWeatherMap()` (ported by hand, kept in sync manually), then fetches and prints the actual condition/temp. No robot, no behavior control, no direct SDK, nothing hardware-facing at all — built to let the geocode fix (a bare phrase like "windsor ontario canada" was returning zero geocode matches; fixed by trying reformatted candidates in order) get re-checked against any phrase in seconds. **Live-robots gate, added 2026-07-25**: arm/nod/rattle route through the shared `live_robots_prompt()` now (see `pha0b`'s own row above) — default (Enter/`y`) reproduces the prior always-live behavior, `n` runs dry; `weather` never touches this gate at all, since it never reaches a robot. Sourced from `run_robot_control_song_001.py`'s own `run_move_axis()`/`run_arm_cue()`/`run_nod()` (moved there from `run_songs_runner_001.py` 2026-07-25 so `brobots_awaken` could share the same tunable movement, `REPORTER_GAP_SHARED_SWITCH_SURVEY_001.md`) plus rattle logic, via `~/.gopod_alias_lib/phcal_isolate_001.py` (outside this repo). See `PHCAL_RUNG1_ISOLATE_AND_WATCH_001.md`, `PHCAL_RUNG2_TUNING_001.md`, `PHCAL_RUNG3_GUIDED_FLOW_001.md`, and this file's own 2026-07-23 findings below. **Detect-first, added 2026-08-18** (`PHCAL_DETECT_FIRST_001.md`) — group 1's `robot_info` pick above is no longer the only way in: bare `phcal` now runs a required startup probe FIRST, before the menu ever draws, reading the configured robot candidates and shaping the session to `none`/`single`/`multi` (an explicit confirm prompt on `multi` — continue, downgrade to `none`-dry-run, or force `single` on one robot), gated by the same low-battery refusal the movement primitives already had, default not-proceed. In `single` mode the menu stays the full 8 groups, but robot-targeting auto-resolves to the one present robot (true single-robot wake, finally the default) instead of asking `1`/`2`/`both`. Full detail: `PHCAL.md`'s own "Detect-first" section. **v6 template rebuild complete, 2026-08-26** (`PHCAL_ARROW_NAV_BUILD_PLAN_006.md`, HEAD `c5fafd1`) — the group menu above is superseded: no more numbered `1`-`8` groups sorted alphabetically, no `0. exit` row (ESC quits). The menu is now this exact hand-ordered, `00.`-`07.`-numbered 8-row sequence (a divider rule between rows 05/06, not a 9th pick): `00` Active Brobots, `01` Audio Play (rattle/danger), `02` Audio Say (weather/announce), `03` Animations, `04` Moves (arm/nod/reverse/stay), `05` Vector's cube — divider — `06` Wake / Ready (wake/responsiveness), `07` Song Tempo. Every screen (detect probe, mode picker, main menu, sub-menus, prompts) now shares one `**`/divider/hint template, and `robot_info`/`brobots_announce_in_sync` gained a real `1`/`2`/`both` robot picker in multi mode too (previously hardcoded to fire both). `PHCAL.md` itself is still flagged WIP, not yet rewritten to match this shape. | `brobots.sh:982` |

**RESOLVED 2026-08-13, superseding the "known broken" note below** (kept for the historical
arc, not because it's still true): `start-the-control-song` no longer targets
`robot_control_song_001` as a plain archived folder at all — as of studio tuning cut 2
(`CONTROL_SONG_LOOP_RETIRED_001.md`) it's golden-registered
(`SONG_REGISTRY`, commit `16aa54e`) and runs on `run_golden_song_001.py`, live-confirmed on
real hardware. See its own row above for the current shape.

**Known broken as shipped (2026-07-22 archive move, historical):** `start-the-control-song`
(targeted `robot_control_song_001`) didn't resolve — that folder stayed archived (now
`zzz_archives/robot_control_song_001`, renamed from `archives/` 2026-07-24), out of scope,
untouched, same operator call as 2026-07-22 ("not a big deal now"). **Corrected 2026-07-24:** `test-arm-cue`/`test-head-nod`/`test-fireworks` were never
actually affected by this — all three route through `run_single_note()`, which never loads
the song folder at all. `test-fireworks` confirmed directly: dry-clean, then live-clean
(`"live": true`, HTTP `200 done`, via `/api-sdk/cloud_intent`) — HTTP-level confirmation, not
an eyes-on-the-robot watch. `test-arm-cue`/`test-head-nod` confirmed by the same code
structure (identical `run_single_note()` path, same no-song-folder dependency), not
independently run this pass. At the time, only `start-the-control-song` genuinely depended
on the folder — since 2026-08-13 it no longer does either.

**Song-folder history, 2026-07-23 → 2026-07-24, briefly:** `brobots_bait_001` and
`brobots_bait_002` were un-archived to top-level `songs/` on 2026-07-23 (fixing
`start-the-bait-song`/`pha0b bait` for one day). On 2026-07-24 the operator did a second,
manual swap: `brobots_awaken`'s own content became the merged bait/capture video (formerly
`brobots_bait_002` — "it IS the awaken video," the operator's own words), the original
3-step weather song was renamed `brobots_bait_000` and archived, and `brobots_bait_001`/
`brobots_preshow`/`brobots_vamp_gate` were archived too (into `zzz_archives`, renamed from
`archives`). Every alias/keyword touching any of these was repointed the same day — see
this file's own 2026-07-24 findings below for the full list. `brobots_bait_002` as a name
no longer exists anywhere; its content lives at `brobots_awaken` now.

RETIRED in `brobots.sh`: `brobots`, `brobots_audio`, `brobots_move`, `brobots_expr`,
`happy-brobots`, `angry-brobots`, `happy-robots`, `angry-robots`.
Private helpers: `_brobots_cue_pair`, `_brobots_move_axis`, `_brobots_play_anim`,
`_brobots_anim_is_loop_token`, `_brobots_play_anim_single`, `_test_anim_isolated`
(backs `test-anim-searching`/`-answering`/`-kg-success`, above),
`_gopod_require_brobots`, `_gopod_note_send` (the shared Python-instrument sender every
converted/new HTTP note above calls into — now prints a `[HH:MM:SS.mmm] NOTE_HTTP
status=... body=...` timestamp prefix), `_robot_sleep_specs`, `_sleep_bench_play`,
`_sleep_bench_segment`, `_score_song_dir`, `_gopod_sleep_first_wirepod_restart_job`.
(`play_anim()` at `brobots.sh:27` is *not* a local
shell function — it's defined inside the here-doc string sent over SSH to run on
`gopod-laptop`; it never exists in this machine's own shell.)

**`wirepodlogs`** / **`wirepoddebuglogs`** / **`wirepodlogsave`** (own file, below) pull
Wire-Pod's own logs — the debug log window is what the stuck-animation check above greps.

### `wirepod_logs.sh`

| Alias | Resolves to | Line |
|---|---|---|
| `wirepodlogs` | `curl` Wire-Pod's standard log window (`/api/get_logs`) | `wirepod_logs.sh:2` |
| `wirepoddebuglogs` | `curl` the debug log window (`/api/get_debug_logs`) | `wirepod_logs.sh:3` |
| `wirepodlogsave` | Pulls both, saves to timestamped files under `~/gopod_wirepod_logs/` | `wirepod_logs.sh:4` |

### `demo.sh`

| Alias | Resolves to | Line |
|---|---|---|
| `is-that-you` | **Single golden-truth launcher for the Is-That-You cross-persona bit** — brings up the cockpit webpage (`:8011`, reused if already serving) and the numpad/PTT chat writer in the foreground, device auto-resolved. NumLock-gated KP1=Doc/KP2=Pip live push-to-talk; the writer's `persona_awareness_reply()` fires an instant fast-reply (no LLM) when the operator's own words name the other robot — the "wrong robot, is that you?" moment, ported from the archived `ptt_gominion1.py`. Prints the full cast on launch: Brobot 1=Doc (KP1), Brobot 2=Pip (KP2), Brobot 0=Operator (Mic_1 — the one physical mic today, naming leaves room for a Mic_2 later). Consolidates and replaces `gopod-demo1` and `gopod_film`, both retired 2026-07-23 (see RETIRED note below) | `demo.sh:60` |
| `gopod-demo1-validation-samples` | Writes the cockpit's validation sample JSON, nothing live | `demo.sh:34` |
| `gopod-ptt-display` | Spectator view for a second terminal/pane while `is-that-you` runs the writer elsewhere. Tails the writer's own `session.log` and highlights just numlock, PTT press/exit, mic/audio, STT transcript, and LLM lines out of the raw console output | `demo.sh:122` |
| `gopod-numpad-map` | Prints the golden numpad/NumLock persona mapping table — fixed lanes (`000`=exit, `KP0`=guest-mic) plus the NumLock ON/OFF KP1-9 grid, read straight from `~/.gopod_alias_lib/numpad_persona_map_001.json` (single source of truth, edited directly to remap a key). `--json` dumps the raw file. Read-only — this alias only ever prints, never writes | `demo.sh:143` |
| `goshot` | Dumps a full reviewable tree/code snapshot to `snapshots/<timestamp>/` (gitignored) — a dev utility | `demo.sh:130` |
| `gopod-json-view` | Opens a GOPOD song's `knobs.json` in the existing tree viewer — resolves a bare `song_id` under `goverlord/runtime/songs/<id>/knobs.json` or a direct path, validates it exists, prints the resolved path, launches the viewer via `xdg-open`/`sensible-browser` | `demo.sh:138` |
| `gopod-index` | **Drift catch, 2026-07-23.** Opens the one-page GOPOD song file index (`~/gopod_index.html`, overridable via `GOPOD_INDEX_PAGE`) in the browser — nothing else. Same firefox-first/`xdg-open`/`sensible-browser` launch shape as `gopod-json-view` above | `demo.sh:250` |
| `gobingo` | **PINNED WORK IN PROGRESS.** Bingo is a genuine third instrument, not a variant of the shell-vs-Python split every other note on this board describes: it drives the vendored `vector-go-sdk` (`vectorx-gobingo`, a Go binary) and a Python `anki_vector` reactor directly, bypassing Wire-Pod's `/api-sdk/*` HTTP surface entirely — no `assume_behavior_control`, no `say_text`, none of the notes any other key plays. It also has no `story.md`/`knobs.json` score of its own — it doesn't follow the song-folder convention every other performable piece in this repo now does (see `102_brobots_bingo_game`, below, for the `pha0b`-side wrapper that gives it one). Pinned rather than folded into the note/sequence/song frame because it structurally sits outside that frame, not because anything about it is broken — both keys are built and working. Launches the Vector-native Bingo binary on Brobot 1 (serial `0dd1b9e9`, `"$@"` passed straight through to `vectorx-gobingo`), plus Brobot 2's angry-animation reactor in the background, automatically killed on exit. **Design confirmed 2026-08-12** (`SONG_102_BROBOTS_1_2_BINGO_GAME.md`'s own "Current state" section): fundamentally a single-robot host — Brobot 1 runs the whole game solo, Brobot 2's reactor no longer reacts per draw (that caused a stale-replay bug) but fires exactly one angry animation at the very end, for missing the chocolate prize specifically. **Also reachable via `pha0b` → pick `102_brobots_bingo_game`**, which adds a "run 1 (continuous) or run 2 (pause for backpack rub)" prompt plus a grid-size (75/90, default 75) prompt, then calls this same `gobingo` with `--pause-for-touch`/`--grid-size` flags (`brobots.sh`'s own `pha0b_menu()`, `102_brobots_bingo_game` case arm). **Resolved 2026-08-12: robot address mismatch** — `~/.anki_vector/sdk_config.ini` had Brobot 2 copied from Brobot 1's own IP/GUID, fixed from Wire-Pod's own `botSdkInfo.json`; Brobot 2 now connects successfully. **Reactor reshaped onto the golden connect-once/stay-put pattern** 2026-08-12 (was reconnecting fresh, full handshake, on every single draw — too slow once draws came close together) — compile-checked clean; per `SONG_102_BROBOTS_1_2_BINGO_GAME.md`'s own text, **not yet re-confirmed live** as of that doc's last update, flagged here rather than overstated. **Note:** `bingo-video-song` (`brobots.sh`, above) is a separate, scored piece for the upsell video — not this live game | `demo.sh:284` |
| `gobingo-reactor` | Brobot 2's angry-reactor alone, standalone — no longer needed for normal use (`gobingo` handles it and cleans it up automatically), kept for manual/standalone debugging in its own terminal, Ctrl-C to stop. **No longer a function** — now a plain `alias`, launched via the reactor's own venv (not system `python3`, since `anki_vector` is installed there) | `demo.sh:315` |

RETIRED in `demo.sh`: `gopod-pre-demo` — used to reset the Wire-Pod table before a demo; its
target script is archived/missing. **Absorbed 2026-07-16:** its archived source
(`pre_demo_wirepod_table_reset_001.py`) checked Wire-Pod service-active, HTTP-reachable,
per-robot presence in a now-nonexistent WDTM registry, a real per-robot speech test, and
Ollama warm — every one of those with a live-hardware payoff is already covered by
`gopod-opening-chord` today, confirmed live. Stays retired, not resurrected as a callable
alias — see `CHORD_ABSORBS_PREDEMO_001.md`. `wdtm-demo` (former name of `gopod-demo1`) is
fully gone, not even a stub. **`gopod-demo1` and `gopod_film` — retired 2026-07-23,
consolidated into `is-that-you`**, the single entry point for the Is-That-You bit (cockpit
auto-launch from `gopod-demo1`, the accurate KP1(Doc)/KP2(Pip) controls line from
`gopod_film`; `gopod_film` was never itself documented here — first and last mention of it
is this retirement note). Neither name is callable anymore.
Private helpers: `_gopod_open_cockpit`, `_gopod_cockpit_healthy`.

### `llm.sh`

| Alias | Resolves to | Line |
|---|---|---|
| `llm-status` | Every configured model (Brobot/Goverlord/deep/coder) + Ollama model list ping | `llm.sh:88` |
| `llm-models` | Raw Ollama model list | `llm.sh:87` |
| `llm-test-brobot` / `-goverlord` / `-deep` / `-coder` | One canned prompt to that specific model, prints the reply — a fast single-model smoke test | `llm.sh:89-92` |
| `llm-test-all` | All four smoke tests in sequence | `llm.sh:93` |

Private helpers: `_gopod_llm_models`, `_gopod_llm_status`, `_gopod_llm_ping`,
`_gopod_llm_test_all`.

### `openwebui.sh`

| Alias | Resolves to | Line |
|---|---|---|
| `owui-detect` / `owui-routes` / `owui-export-script` / `owui-tune-script` / `owui-import-script` / `owui-fix-port` / `owui-chain` | Thin wrappers around numbered setup/maintenance scripts in `$OWUI_TOOL_DIR` | `openwebui.sh:8-14` |
| `owui-local` / `owui-lan` / `ollama-lan` | Print known-good local/LAN URLs, no network call | `openwebui.sh:17-19` |
| `owui-url` | Probes localhost/LAN in turn, prints whichever answers | `openwebui.sh:127` |
| `owui-wait` | Same, retries up to two minutes | `openwebui.sh:128` |
| `owui-status` | URL, Docker state, listening ports, live Ollama model list | `openwebui.sh:129` |
| `owui-api-test` | Hits the model API with `$OWUI_TOKEN`, or reports it missing | `openwebui.sh:130` |
| `owui-export` | Snapshots the model list to a timestamped JSON file | `openwebui.sh:131` |
| `owui-debug` | Raw HTTP status/content-type for root/models(no token)/models(token) — for diagnosing a broken connection | `openwebui.sh:132` |
| `owui-fix` | Port-fix script, then `owui-wait` | `openwebui.sh:135` |
| `owui-ready` | `owui-status` then `owui-api-test` — the "is it actually usable" check | `openwebui.sh:136` |

Private helpers: `_owui_url`, `_owui_wait`, `_owui_status`, `_owui_api_test`,
`_owui_export`, `_owui_debug`.

### `suits.sh`

| Alias | Resolves to | Line |
|---|---|---|
| `owui-env` | Loads the saved Open WebUI API key from the secure vault into the shell | `suits.sh:1` |
| `owui-key-save` | Saves `$OWUI_API_KEY` to the secure vault | `suits.sh:7` |
| `owui-chat-bridge-json` | Writes the GOPOD Chat Bridge skill JSON into the repo's suits directory | `suits.sh:17` |
| `owui-suit-auto` | Pushes the Suit Changer model config (system prompt, tools, skills) to Open WebUI over its API | `suits.sh:37` |
| `owui-chat-bridge-close` | `owui-key-save` + `owui-chat-bridge-json` + `owui-suit-auto`, back to back — the full chat-bridge setup in one key | `suits.sh:85` |
| `phaob` | Prints the current Open WebUI alias-fitting playhead checklist (Frame 0 through Frame 1, PASS/expected states) inline | `suits.sh:91` |
| `pha0b` | Same checklist, cat'd from its file in the repo instead of the inline copy | `suits.sh:118` |

### `goverlord.sh`

| Alias | Resolves to | Line |
|---|---|---|
| `goverlord-suits` | Runs the script that creates and measures Goverlord's LLM suits | `goverlord.sh:4` |
| `suit-digest` | Runs the suit digest cycle script | `goverlord.sh:5` |
| `suit-latest` | Lists the 30 most recently touched files in the suit state dir | `goverlord.sh:6` |
| `suit-report` | Cats the newest suit digest report | `goverlord.sh:7` |
| `suit-measure` | Pretty-prints the newest suit digest measurement JSON | `goverlord.sh:8` |

### `chat_capture.sh`

| Alias | Resolves to | Line |
|---|---|---|
| `gopodcap` | Runs any command, captures its JSON output via `gopod_json_capture.py` | `chat_capture.sh:3` |
| `gopodcap-gopod` | Same, inside a shell pre-sourced with core/frame0/brobots/demo, `cd`'d into the repo | `chat_capture.sh:9` |
| `gopod-json-tail` | Tails the last 80 lines of the capture log | `chat_capture.sh:23` |
| `gopod-json-clear` | Clears the capture log | `chat_capture.sh:27` |

### `frame0.sh`

Every alias in this file is RETIRED (target scripts archived): `gopod_hdb`, `gopod_pots`,
`gopod_mots`, `gopod_cawg`, `gopod_paat`, `gopod_gpys`, `is-that-you`, `gopod_playhead`,
`gopod_playhead_next`, `gopod_playhead_prev`, `gopod_frame0_preintent`,
`gopod_frame0_live_all`. The file is still sourced (harmless — nothing left to load).

### `tools.sh` — off the board entirely

Exists on disk, **not sourced by `.bashrc` or `.bash_aliases`** — confirmed fresh 2026-07-10,
same finding as the original 2026-07-07 sweep. No key on this keyboard reaches it.

| Function | Resolves to | Line |
|---|---|---|
| `owui-tool-auto` | Open WebUI tool-JSON pusher | `tools.sh:1` |
| `owui-tool-cut-gopod-alias` | (unreachable — orphaned file) | `tools.sh:94` |
| `owui-tool-show-gopod-alias` | (unreachable — orphaned file) | `tools.sh:111` |

---

## Render Controls

*(Folded in from `ALIAS-MIXER.md`, 2026-07-16 consolidation — this section absorbs
everything that doc used to carry.)* How a note actually renders once played: async vs.
sync (does it fire-and-forget or block), hold/timing (how long a note is held before it's
released), and temperature-style dials (same family as the interview's thinking-window
weights — knobs that shape *how* a note plays, not *whether* or *when in a sequence* it
plays). This is about the knobs on a channel, not the order channels play in — arrangement
order lives in `ALIAS-SEQUENCER.md`.

**Async vs sync.** `playAnimationWI` (used by every `brobots-anim-*` note, and every
reaction-lane test below) is the async, fire-and-forget token — it launches a goroutine on
the Wire-Pod side and returns immediately, and it does not interrupt speech. `playAnimation`
(the token behind `brobots-happy`'s / `brobots-angry`'s style of cue, and the original
crash source before the reaction-lane fixes) is the blocking/sync counterpart — it runs
inline and interrupts speech. Every render-mode choice a note makes is this same
distinction: does it hand control back to the caller before or after the physical/robot
action actually finishes.

**Hold-before-release timing.** The movement notes (`brobots-lift-up`, `brobots-lift-down`,
`brobots-head-nod`) and the animation notes (`brobots-anim-*`) all share one render shape:
the underlying Wire-Pod call returns before the real motion or clip finishes (fire-and-forget
at the HTTP layer, regardless of the async/sync distinction above), so each note holds
behavior control for an estimated duration — a `hold` argument, overridable, defaulting to
~1.2s for lift, ~0.35s per half for the nod, ~2.5s for a standalone animation clip (`_brobots_play_anim`), and 5.0s for the reaction-lane tests' animation dispatch (the
operator's own proven value, retuned down from an unreliable 18.8s calculated figure that
was explicitly discarded) — before releasing. That hold is a stand-in for real completion
time, not proof of it. This is a render control (how long the channel stays "open" for a
given note); today it's hardcoded inside each note's own function rather than exposed as a
separate dial.

**Temperature-style dials.** Not yet cataloged in any note — flagged as the same family of
control as the interview's thinking-window `cycle_weights` (Template 1 / `THINKING WINDOW:`
/ `CYCLES:` fields): a dial that shades *how* something renders (tone, intensity, variation)
without changing *what* fires or *in what order*. No note currently exposes a dial like
this; recorded as the kind of control this section expects to grow into, not something
built yet.

**2026-07-16 registry-polish addendum — instrument change, render unchanged.**
`brobots-lift-up`/`-lift-down`/`-head-nod`/`-anim-*` send their HTTP calls through
`_gopod_note_send` instead of a separately hand-written curl client. This is an instrument
change, not a render change — every hold duration named above is untouched, the
async/fire-and-forget shape at the HTTP layer is untouched.

Render shape of the five notes added that same pass:

- **`gopod-conn-test`** — single request per robot, no hold, no async ambiguity: `conn_test`
  blocks until Wire-Pod's own connection check returns, pass/fail read straight off the
  HTTP body.
- **`gopod-vamp`** — sequential, blocking: each vamp beat's Kokoro call blocks on that beat's
  own speak-stdin worker returning its DONE ack (the same "true completion" render shape
  `_preshow_speak_host` already uses in the real song), one beat fully finishes speaking
  before the next one starts. No animation/movement note involved — voice only.
- **`gopod-weather-say`** — assume → one `say_text` call (blocking on Wire-Pod's own accept,
  not on real speech duration) → release. No new render shape; identical to every other
  spoken line in this codebase.
- **`gopod-fireworks`** — no new render shape of its own; delegates straight to
  `test-fireworks`'s existing one (`fire_fireworks()` → `/api-sdk/cloud_intent`).
- **`gopod-pick-model`** — not a hardware note at all, so no hold/async question applies.
  Interactive (blocks on a real terminal's `input()`) unless `GOPOD_CONTENT_MODEL` is set, in
  which case it returns immediately with no menu.

**Reaction-lane render shape, 2026-07-16 (`test-reaction-in-the-beat` and friends).** Two
separate `say_text` calls, not one sentence with the token embedded (a deliberate safety
change — see [Today's findings, 2026-07-16](ALIAS-LIBRARY-FINDINGS-ARCHIVE.md#todays-findings-2026-07-16)): a spoken emotion line, a real
2-second pause, then a bare animation-only dispatch through `playAnimationWI`, held 5.0s,
then released. A stuck-animation check (`get_debug_logs`, grep for `"waiting for animation
to be done"`) runs between dispatch and hold as a live health read, not a blocking gate.

---


**Older dated findings (2026-07-16, 2026-07-17 Bingo rattle, 2026-07-23) archived
2026-07-25** — see [ALIAS-LIBRARY-FINDINGS-ARCHIVE.md](ALIAS-LIBRARY-FINDINGS-ARCHIVE.md).
Only the most recent dated sections stay here going forward.

---

## Today's findings (2026-07-24)

**The operator swapped two song identities by hand.** `brobots_awaken`'s own folder now
holds the merged bait/capture video's content (formerly `brobots_bait_002`) — the
operator's own call: "This is correct — it IS the awaken video." The original pure 3-step
weather-only song was renamed `brobots_bait_000` and moved into `zzz_archives/` (renamed
from `archives/` the same pass). `brobots_bait_001`, `brobots_preshow`, and
`brobots_vamp_gate` moved into `zzz_archives/` alongside it. Both moved `knobs.json` files'
own `song_id` fields were fixed to match (`brobots_bait_002` → `brobots_awaken`;
`brobots_awaken` → `brobots_bait_000`) — a raw filesystem move doesn't touch file content,
so without this fix each song's internal identity would have kept naming the *other* one.
Both `story.md` titles and cross-links fixed the same way — see each song's own file.

**Every alias/keyword repointed, none retired.** All of it stayed genuinely reachable and
useful, so every fix below is a repoint, not a retirement:
- `start-the-bait-song` → `songs/00_brobots_awaken` (follows the capture-video content to
  *its* new name/home — the opposite direction, since this one moved to a top-level,
  non-archived folder).
- `start-the-net-song` → `zzz_archives/brobots_bait_001`.
- `gopod-vamp`'s own hardcoded `load_preshow_song()` call →
  `01_brobots_interview_section_01/vamp/` (repointed again 2026-08-08 — vamp is part 1 of
  the vamp+interview song, interview is part 2, both now live under the interview song's
  own folder; vamp keeps its own subfolder since the loader needs plain-named
  `knobs.json`/`story.md`/`zKnobs.json` and the interview's own top-level files already
  claim those names). `start-the-preshow-song`, the alias that once reached this same
  content end-to-end, was removed outright 2026-08-08 — there is no preshow alias anymore,
  only `gopod-vamp`'s own standalone preview of it. **Superseded 2026-08-19**: vamp and
  interview split into two fully separate top-level song folders
  (`01_brobots_interview_vamp`/`02_brobots_interview_run`) — `gopod-vamp`'s call target is
  now `01_brobots_interview_vamp/` directly, no `/vamp` subfolder, no shared folder with
  the interview's own content. This 2026-08-08 entry is kept as the historical record of
  that pass, not current truth.
- `run_vamp_gate_song_001.py`'s own `DEFAULT_SONG_DIR` (backs `pha0b vamp`) →
  `zzz_archives/brobots_vamp_gate`.
- `pha0b`'s own `weather`/`bait` case-statement entries repointed the same way as their
  matching aliases above. `pha0b_menu`'s `brobots_awaken` → keyword mapping changed from
  `weather` to `bait` (matching the folder's real content now); the now-permanently-dead
  `brobots_bait_002` mapping removed outright (that name will never exist again — contrast
  `brobots_vamp_gate`/`robot_control_song_001`'s own dormant-but-harmless mappings, kept,
  since those folder names still exist, just archived).
- `pha0b_menu`'s disk scan now explicitly excludes `zzz_archives` itself from the song
  list — the same "a container folder isn't a song" finding as before, recurring under the
  renamed folder; excluded by name this time instead of just falling through to a refusal.

**`robot_control_song_001` untouched, deliberately, again** — same out-of-scope call as
2026-07-22/2026-07-23; `start-the-control-song`/`test-arm-cue`/`test-head-nod`/
`test-fireworks` remain broken.

A read-only survey found two failure-mode findings — an orphaned LLM-calling thread surviving a
`start-the-preshow-song` crash, and `start-the-net-song` hanging silently instead of
failing fast — that this pass fixed the underlying paths for.

**GOLDEN ALIAS NOTE — `phcal`'s cold-first-cycle fix.** Live-reported by the operator:
`phcal arm 2 --cycles 2` produced only ONE visible arm cue; `--cycles 1` produced NONE.
Same symptom on `nod` ("I ask for 3 nods, I get 2"). Not a loop/count bug — the loop
already fires exactly N full pairs, verified. Root cause: `run_assume_control()` returned
and the first `move_lift`/`move_head` call fired ~2ms later (log: assume at +0.000s, first
move at +0.002s) — the robot wasn't physically ready yet, so cycle 1's motion was silently
dropped while cycle 2+ landed once warm. Same class of bug as the already-fixed "cold first
press" finding in `test-silent-angry-say` (HTTP success, no playback on the first live
action after a wake, second press played) — that fix's answer (a settle pause after the
wake, before the first live action) is the same fix applied here: a new
`PHCAL_ASSUME_SETTLE_SECONDS` (reuses `PHCAL_PREFLIGHT_SETTLE_SECONDS`'s own 1.5s, not
invented) now sleeps inside `run_assume_control()`, after `assume_behavior_control`, before
`cmd_arm`/`cmd_nod`'s first move — both callers get it for free from the one choke point.
Pending: live re-confirmation with `--cycles 1`/`nod 2 1` to confirm the first cue now
lands.

---

## Today's findings (2026-07-25)

**Live-robots gate, decoupled and shared.** `pha0b`'s bingo-only hardcoded-always-live
special case and every other song's silently-dry-unless-pre-exported default were two
different behaviors for the same underlying question, discovered live when the operator
hit both in one menu pass. Fixed with one shared function, `live_robots_prompt()`
(`brobots.sh`) — a single "live robots? y/n [default y]:" prompt, called from both `pha0b`
and `phcal` rather than duplicated per caller (same precedent as
`restart_wirepod_preflight()`). Default reproduces each caller's prior live behavior;
`n` runs dry. See both rows above.

---

## Today's findings (2026-08-12)

**Per-robot network identity — golden, verified source.** Brobot 2's reactor (and any other
direct-SDK Python `anki_vector.Robot()` caller) reads `~/.anki_vector/sdk_config.ini` for
IP/GUID/cert per serial — a separate file from what Wire-Pod itself uses (`gobingo`'s own
`sdk_wrapper.InitSDKForWirepod()` reads `chipper/jdocs/botSdkInfo.json` instead, Wire-Pod's
own authoritative robot registry). These two files can drift apart. Found live: `sdk_config.ini`
had Brobot 2's entry silently copied from Brobot 1's own IP and GUID — connections to "Brobot 2"
were actually hitting Brobot 1's device, failing every attempt with a certificate mismatch.
Fixed by reading the real values straight from `botSdkInfo.json` and writing them into
`sdk_config.ini`.

**Golden ID reference**, confirmed working 2026-08-12 — check
`~/wire-pod/chipper/jdocs/botSdkInfo.json` for the live IP if either of these ever needs
re-verifying. Only the serial (ESN) is a permanent hardware constant; IP is DHCP-derived and
can in principle change, but operator-confirmed steady in practice on this network — not
written here (no network identifiers in tracked files, same rule `.claude/skills/
gopod-layer/` enforces everywhere else), only in `botSdkInfo.json` itself and local machine
config. GUID is a re-pairing-derived credential, not banked here at all:

| Robot | Serial (ESN) | IP |
|---|---|---|
| Brobot 1 | `0dd1b9e9` | see `botSdkInfo.json` / local config, steady in practice, not guaranteed |
| Brobot 2 | `0dd1d8bf` | see `botSdkInfo.json` / local config, steady in practice, not guaranteed |

If a direct-SDK Python tool (the bingo reactor, or any future one) ever fails to connect with a
"self signed certificate" / "Unable to establish a connection" error again, check
`~/.anki_vector/sdk_config.ini` against this table (or the live `botSdkInfo.json`) before
assuming a code bug — this exact failure mode has already happened once.

---

## The json-view alias

`gopod-json-view <song_id|path>` opens a song's `knobs.json` in the existing viewer at
`~/tools/chatgpt-json-tree-viewer/chatgpt-json-tree-viewer.html`. **The viewer itself was
not touched** — confirmed by re-reading its source that it has no URL-parameter or
programmatic load hook (only its own file-input and drag-and-drop), so this alias cannot
inject the file into the page without editing the viewer, which was out of scope. What it
does instead: resolves the target (a bare `brobots_bingo`-style **folder name**, literally
`goverlord/runtime/songs/<target>/knobs.json`, or a direct path), validates the file is
real, prints its resolved absolute path to the terminal, and launches the viewer —
leaving one manual step (the viewer's own file control or a drag) instead of first having
to locate the file by hand. No clipboard tool (`xclip`/`xsel`/`wl-copy`) is installed on
this machine, confirmed, so the path is printed rather than silently assumed copyable.

**Real `song_id` resolution, fixed 2026-08-19** (`GOPOLISHER_FIXES_001.md`) — the folder-name
lookup above is tried first (the fast, unchanged path for every song where folder name and
`song_id` still coincide); if that path doesn't exist, it falls back to scanning every
song's `knobs.json` for a matching `"song_id"` field and resolves to that folder instead.
Fixes the real gap the interview vamp/run split exposed (`GOPOLISHER_FOCUSED_001.md`):
`gopod-json-view brobots_interview_section_01` (RUN's song_id, folder moved to
`02_brobots_interview_run/`) and `gopod-json-view brobots_preshow` (VAMP's own song_id, no
matching folder name at all) both now resolve correctly, verified directly against the
live function with `GOPOD_NO_BROWSER=1`, alongside two regression checks
(`00_brobots_awaken`, `101_brobots_bingo_test`, both still resolve via the fast
folder-name path, unaffected) and a still-clean-failure check on a nonexistent target.

Tested: usage error, missing-song error, valid bare song id, and a direct path — all four
verified against the live deployed function with `GOPOD_NO_BROWSER=1` (skips the actual
browser launch, exercises everything else).

---

## Historical drift record (closed 2026-07-16)

Cross-checking the 2026-07-10 survey against the old `ALIAS-PIANO.md`'s prose (last fully
swept 2026-07-07 at that point) turned up five real aliases in `brobots.sh` that PIANO's
tables didn't mention by name at the time: `brobots-grep`, `start-the-preshow`,
`gopod-preshow-then-interview`, `gopod_warm_up`, `gopod_interview`. By 2026-07-16, PIANO had
drifted further — six more launchers (`start-the-control-song`, `test-arm-cue`,
`test-head-nod`, `test-fireworks`, `start-the-bait-song`,
`test-interview-movements`) had been added to `brobots.sh` between 2026-07-12 and 2026-07-15/16
with no PIANO entry at all. The 2026-07-16 registry polish pass
(`ALIAS_REGISTRY_TRUTH_SWEEP_001.md` found the drift; `ALIAS_REGISTRY_POLISH_001.md`
executed the fix) closed the gap in both docs. This registry (this doc) is now the sole
source going forward — there is no second prose copy left to drift out of sync with it.

---

> From Doctrine Barfallonyou
> Lesson! A keyboard nobody can read is just a wall of switches.
> Boom! Done! Class Dismissed!
> — Doc Squawkadoodle

---

## GOPOD YAHMM (You Are Here Mall Map)

Two doors. Pick the one that fits how far in you want to go — nothing here needs a set order beyond that.

### Door 1 — New here? Explore GOPOD wide

Plain language, first look, no background needed — for a newcomer, human or AI.

**Start here**
- [README.md](../../README.md) — what GOPOD is and how it's built
- [AWAKEN.md](SONG_00_BROBOTS_1_2_AWAKEN.md) — watch first: a brobot wakes, checks itself, greets you
- [QUICKSTART.md](../SINGLE_BOT_QUICKSTART.md) — talk to your own Vector, one robot, no alias studio needed
- [MY_NICHE_BUZZ_ASK.md](../../MY_NICHE_BUZZ_ASK.md) — help test the keyboard grabber, no robot required
- [GOPOD_SONGS.md](GOPOD_SONGS.md) — all songs, explained in plain language, no background needed
- [PALM_TREE.md](../../life/01_PALM_TREE.md) — the whole thing, put together, no background needed
- [FUNNY_NAMINGS.md](../../web/FUNNY_NAMINGS.md) — every name, character, and phrase this project uses, explained once

**The songs**
- [INTERVIEW VAMP.md](SONG_01_BROBOTS_1_2_INTERVIEW_VAMP.md) — the flagship's video 1, the pre-show banter
- [INTERVIEW RUN.md](SONG_02_BROBOTS_1_2_INTERVIEW_RUN.md) — the flagship's video 2, the seven exchanges
- [BINGO.md](SONG_101_BROBOTS_1_2_BINGO.md) — the shareable upsell video
- [BINGO GAME.md](SONG_102_BROBOTS_1_2_BINGO_GAME.md) — the live two-brobot Bingo warm-up act
- [BABY ROBOTS SLEEP.md](SONG_105_BROBOTS_1_2_BABY_ROBOTS_SLEEP.md) — Doc's origin, told as a bedtime story
- [IS-THAT-YOU SINGLE.md](SONG_103_BROBOTS_1_2_IS-THAT-YOU_SINGLE.md) — the single-bot jewel, live and unscripted
- [IS-THAT-YOU MULTI.md](SONG_103_BROBOTS_1_2_IS-THAT-YOU_MULTI.md) — the two-brobot mix-up, live and unscripted

**The content engine**
- [NICHE_PILLARS.md](../../web/NICHE_PILLARS.md) — how the writing is split into nine kinds of everyday maths, and why
- [AI_WORDPLAY.md](../../web/AI_WORDPLAY.md) — the engine: AI Wordplay, the contests that feed it, and where the content lands
- [AHA_MOMENT.md](../../web/AHA_MOMENT.md) — a live demo the reader can feel work
- [FOODMATH_AHA_MOMENT.md](../../web/FOODMATH_AHA_MOMENT.md) — the foodmath cousin: a live subdomain built in hours, the gap between live and built is the demo
- [BIRTHDAY.md](../../web/BIRTHDAY.md) — the physical proof one — real food-car props, no rendering required
- [NEWSLETTER.md](../../web/NEWSLETTER.md) — subscribe to CRUSHN8R CREW'd — the live follow-along lane

**For venues, funders, and partners**
- [MOBILE_GEAR.md](../MOBILE_GEAR.md) — mobile deployment and field kit
- [OPS ASK.md](../../MY_GOPOD_OPS_ASK.md) — the operator's ops ask — social, sites, and content, a different role than the technical one
- [HEALTHY_DISTRACTIONS.md](../../life/02_HEALTHY_DISTRACTIONS.md) — GOPOD as healthy distraction
- [OUTREACH.md](../../life/02a_OUTREACH.md) — the community and paid outreach plan, two lanes side by side

**For teachers**
- [EDUCATION.md](../../life/03_EDUCATION.md) — GOPOD as a teaching tool
- [TEACHER_INSIGHT.md](../../life/04_TEACHER_INSIGHT.md) — what a session shows a teacher about their room

### Door 2 — More? Dive GOPOD deep

For readers who lean in, who know GitHub — the technical docs, the operator tooling, how it's made.

**Help wanted**
- [MY_GOPOD_ASK.md](../../MY_GOPOD_ASK.md) — the operator's own ask — what's built, where the line is, what kind of help this needs

**For Wire-Pod owners and builders**
- [WIRED-POD.md](../WIRED-POD.md) — what GOPOD changed and added on top of Wire-Pod
- [GOPOD_FEATURES.md](../GOPOD_FEATURES.md) — everything GOPOD built, feature by feature, stack included
- [BODIED_BROBOTS.md](../BODIED_BROBOTS.md) — the robot bodies — proof ladder, Vector L3, Cozmo/Scout staged below L3
- [ALIAS-SEQUENCER.md](ALIAS-SEQUENCER.md) — arrangement: notes into sequences into songs
- [PHA0B.md](PHA0B.md) — PLAYHEAD Part 2, the performance front door, whole songs at a time
- [PHCAL.md](PHCAL.md) — PLAYHEAD Part 1, the calibration bench, one primitive at a time

**Doc's Take — lessons learned**
- [DOCS_TAKE_LESSON_1.md](../../life/101_DOCS_TAKE_LESSON_1.md) — the first real mistake GOPOD survived
- [DOCS_TAKE_LESSON_2.md](../../life/102_DOCS_TAKE_LESSON_2.md) — fix the foundation, not the symptom
- [DOCS_TAKE_LESSON_3.md](../../life/103_DOCS_TAKE_LESSON_3.md) — a placeholder that looks finished is worse than an honest gap
- [DOCS_TAKE_LESSON_4.md](../../life/104_DOCS_TAKE_LESSON_4.md) — a living thing sheds, and shedding shows you what is left
- [DOCS_TAKE_LESSON_5.md](../../life/105_DOCS_TAKE_LESSON_5.md) — the scary answer is usually the true one, not the easy binary
- [DOCS_TAKE_LESSON_6.md](../../life/106_DOCS_TAKE_LESSON_6.md) — the honest edge isn't the ask — say it out loud

**The philosophy**
- [LEGACY.md](../../life/05_LEGACY.md) — the pedagogy behind what GOPOD propagates

**How this gets made**
- [AI_AHA_MOMENTS.md](../../web/AI_AHA_MOMENTS.md) — the aha moments from making GOPOD with AI, for the reader who wants to see how the thing thinks

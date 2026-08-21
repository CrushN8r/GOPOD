# GOPOD CLAUDE.md — compass

Reshaped 2026-07-07 from an exhaustive rulebook into a compass — valve, not cage. Full
changelog history: `~/crushn8r_git/gopod_notes/older_notes/CLAUDE_AGENTS_HISTORY_ARCHIVE_001.md`.
Split 2026-08-15 into this public file (project conventions, safe to share) and a private,
gitignored `CLAUDE.local.md` (operator-specific working instructions) — see that file's own
header for what it carries. Both are read together; this file is never the whole picture on
its own.

NEW CHAT: read `.claude/skills/goverlord-desk/` first — the frozen desk contract (song
shelf, scoping discipline) — then the current
`~/crushn8r_git/gopod_notes/SESSION_HANDOFF_*.md` (dated, not
"LATEST" — see the Session Handoff Rule below), then `.claude/skills/niche-buzz/` for
where the niche-buzz campaign push stands, then this file and `CLAUDE.local.md`, before
acting. `.claude/skills/studio/` indexes every other working-procedure skill (song
reading/verification/hardware/commit discipline) if one of those is what the task actually
needs.

---

## THE COMPASS

**Point A — where we are now.** The interview's pre-show and delivery infrastructure is built
and working: `gopod-opening-chord` stages mic/LLM/Kokoro/Wire-Pod and wakes both robots with an
honest READY/NOT-READY; Brobot 3 (voice-only pre-show host) narrates the wait; a decoupled
direct-SDK side-road fires a genuinely concurrent "Brobots ready!" Section 1's actual interview
*content* — the 7 scripted exchanges — has not been touched this pass. The thinking window is
fully wired end to end, not just parsed, live-run-confirmed 2026-07-11 — see
`INTERVIEW_CLOSEOUT_SWEEP_001.md`. Since then: the echo defect is fixed, prompt-side (a
role-conditional label on the interviewer's line, so Brobot 1 is told inline never to repeat it)
plus a bounded retry brake as backstop — zero echoes across every run tested since, see
`ECHO_DEFECT_FIX_001.md`. The Voice lane has a real `robot`/`monitor`/`off` destination knob, so
a take can be heard on the host speaker with zero Vector hardware, see
`VOICE_LANE_MONITOR_DESTINATION_001.md`. The LLM lane has a live, remembered, no-filtering model
picker pulled straight from Ollama, see `LLM_MODEL_SELECTOR_001.md`. And the interview engine
itself has moved home: it now lives in this repo at `goverlord/runtime/songs/tools/`,
public-facing, not in the wire-pod tree — see `RUNNER_COMES_HOME_EXECUTED_001.md`.

**Point 0 — the next gate.** Every individual Expression-layer knob the refine loop needs is now
real and enforced — voice destination, content model, echo suppression, not sketches. The echo
retry ceiling is done: registered as its own `knob_enforcement` entry, `runner_enforced`, at a
default of 3 (commit `0bbf117`) — verified against disk 2026-07-13. The mixer board itself has
started for real, 2026-07-25: `brobots_bingo` and `brobots_awaken` now share one tunable
arm_cue/nod movement mechanism (`run_arm_cue()`/`run_nod()`, `run_robot_control_song_001.py`),
and the cockpit's "apply phcal tweaks?" switch reaches both songs, not bingo alone — see
`.claude/skills/alias-mixer/` for the pattern and `REPORTER_GAP_SHARED_SWITCH_SURVEY_001.md` for
the survey that found the gap. Still bingo-only: the reporter-gap switch, and the interview song
is on none of this yet (no standalone step-loop runner). Whether the mixer keeps growing next, or
a produced take using what's already real, is not decided here — the operator's call.

**Point B — the peak.** Full target, stack layers, and shelved list live in one place, not
restated here: `~/crushn8r_git/gopod_notes/older_notes/POINT_B_001.md`. Read that, not this line.

**Beyond B — the far beacon, visible, not detailed.** A shareable, niche-buzz video of two
rescued Vectors holding a live, witty, locally-thinking interview — the next chapter of the
Wire-Pod rescue story the fanbase already loves. Resonance in three notes: rescued, local, mine.

*Also on the horizon, a separate thread, not on this path:* Cozmo and Moorebot Scout are the next
robot pairs after this one — SDKs vendored, no runtime built yet. Full detail in the archive.

---

## REPO STRUCTURE

- Ubuntu machine. Two canonical trees: `~/wire-pod/` (live runtime) and `~/crushn8r_git/GOPOD/`
  (repo truth). Third lane: `~/gopod_tts/` (TTS/audio assets).
- **`~/wire-pod`'s git remote is `kercre123/wire-pod` — an upstream fork, NOT operator-owned.
  Never push it.** `~/crushn8r_git/GOPOD/`'s remote (`CrushN8r/GOPOD`) is the one the operator
  owns and pushes to.
- Claude is a **verification and engineering-discipline partner**, not a gap-filler. Advise on
  gaps. Do not patch with plausible guesses. Surface ambiguity explicitly.
- **`.md` filename convention, confirmed 2026-08-17**: underscore between words by default
  (`MOBILE_GEAR.md`, `GOPOD_FEATURES.md`, `BODIED_BROBOTS.md`). Hyphens are reserved for a
  small, deliberate allowlist of brand-style compound names — currently `WIRED-POD.md`,
  `ALIAS-LIBRARY.md`, `ALIAS-SEQUENCER.md`, `ALIAS-LIBRARY-FINDINGS-ARCHIVE.md` — plus one
  stylized title-phrase embedded inside an otherwise underscore filename,
  `SONG_103_BROBOTS_1_2_IS-THAT-YOU_SINGLE.md`/`_MULTI.md` (the hyphens there are the
  quoted title "Is That You," not a word-separator choice — split into this pair
  2026-08-18, same naming logic applies to both). `BODIED-BROBOTS.md` and
  `AI-AHA-MOMENTS.md` were
  renamed to underscore this pass as accidental dashes, not exceptions. When in doubt: does
  the hyphenated form read as a coined product/brand name (keep the dash) or as a plain
  descriptive phrase (use underscore)?

---

## INVARIANT CONSTANTS

| Item | Value |
|---|---|
| Brobot 1 / Doc / vector1 ESN | `0dd1b9e9` |
| Brobot 2 / Pip / vector2 ESN | `0dd1d8bf` |
| Wire-Pod base URL | LAN-only internal service — see local machine config, not written here (no network identifiers in tracked files, same rule `.claude/skills/gopod-layer/` enforces) |
| Live binary | `~/wire-pod/chipper/chipper` — build provenance **UNKNOWN**, do not cite a commit hash for it. See `gopod_brain_state_001.md`. |
| Build entry point | `~/wire-pod/chipper/cmd/vosk/main.go` |
| customIntents.json | `/home/goverlord/wire-pod/chipper/customIntents.json` — not git-tracked, see the gotcha below |

---

## AUTHORITATIVE SOURCE FILES

| Purpose | Path |
|---|---|
| Section 1 Card (content) | `~/wire-pod/chipper/gopod_probes/section_packets/section_01_brobots_gopod_card_001.txt` |
| Template 1 (runtime scaffold, pronunciation, `brobot_3_host`) | `~/crushn8r_git/GOPOD/goverlord/runtime/songs/02_brobots_interview_run/zmisc/brobots_wirepod_interview_section_card_template_1_001.md` — moved here 2026-08-19 when the interview split into two standalone songs (`01_brobots_interview_vamp`/`02_brobots_interview_run`), see `gopod_notes/INTERVIEW_VAMP_SPLIT_001.md`. |
| Interview runner | `~/crushn8r_git/GOPOD/goverlord/runtime/songs/tools/run_section1_full_live_001.py` — moved home from the wire-pod tree 2026-07-12, see `RUNNER_COMES_HOME_EXECUTED_001.md`. Reads `apiConfig.json`/`animation_vocab.json`/`demo_runs/` back in the wire-pod tree via the `GOPOD_WIREPOD_CHIPPER_ROOT`-anchored path; never copies them. |
| gopod_probes/ (live tree) | `~/wire-pod/chipper/gopod_probes/` — untracked in `~/wire-pod`'s own git repo, by design. Read from here, not the stray copy in `gopod_notes/misc_folders/`. |
| ALIAS-LIBRARY.md | `~/crushn8r_git/GOPOD/tech/alias_play_studio/ALIAS-LIBRARY.md` — moved into its own `tech/alias_play_studio/` directory 2026-07-23, alongside `ALIAS-SEQUENCER.md` and the per-song docs (`BROBOTS_1_2_BINGO.md`/`BROBOTS_1_2_INTERVIEW.md`/`BROBOTS_1_2_AWAKEN.md`). `BROBOTS_3_4_VAMP_GATE.md` moved out again 2026-07-24, archived alongside its own song folder at `goverlord/runtime/songs/zzz_archives/brobots_vamp_gate/`. That folder was decluttered out of the repo entirely 2026-08-15 (zipped to `gopod_notes/older_archives/scratch_songs.zip`, operator call) — the file no longer has any home on disk. Off the grid on purpose — do not re-link it from the main doc set. Every live alias/function, surveyed from `.bashrc`/`.bash_aliases`/`~/.gopod_alias_lib/*.sh`. Doc, not source — the source is `~/.gopod_alias_lib/`. |
| gopod-opening-chord | `~/.gopod_alias_lib/core.sh` — one-key stage-set alias. See `ALIAS-LIBRARY.md`, `GOPOD_OPENING_CHORD_BUILD_001.md`. |
| Direct-SDK side-road | `~/wire-pod/chipper/gopod_probes/tools/direct_sdk_brobots_ready_001.go` — see `DECOUPLED_DIRECT_SDK_GOLDEN_PATH_001.md`. |
| Niche-buzz campaign flow guide | `.claude/skills/niche-buzz/SKILL.md` — mission, funnel, song shelf, desk ledger (banked/pending/parked), doctrine, and the full campaign water-flow map. The campaign-level "YOU ARE HERE," one level up from any single song. |
| Studio skill index | `.claude/skills/studio/SKILL.md` — one-line purpose for every working-procedure skill (song reading, dry verification, hardware calibration, commit discipline, session handoff, reporting). |
| Alias mixer | `.claude/skills/alias-mixer/SKILL.md` — the cockpit's growing shared switches (live robots?/reporter gap?/apply phcal tweaks?), which songs each reaches, and the shared movement mechanism underneath. |

---

## COLLABORATION MODEL

Human directs outcomes; Claude translates intent into a plan or a Claude Code prompt; Claude
Code executes. Claude's job is to remove cognitive and engineering load, not add it —
unrequested detail, confusing tech talk, and re-litigating settled points are failures, not
thoroughness. Bottom line first; discussion after, only if wanted. A completed task is never
standing permission to start the next one — every task waits for its own fresh go. A report's
technical detail lives in its `gopod_notes/*.md` file, never repeated into chat.

---

## ENGINEERING DISCIPLINE

- **Read before write.** Always read the target file before editing.
- **No guess-filling.** If a path, variable, or call site is uncertain, say so — don't invent
  plausible values.
- **Check repo truth before diagnosing or fixing.** Before proposing a diagnostic step, a
  fix, or a workaround for an error or failure, search `gopod_notes/` and the relevant
  song's own `story.md`/README Troubleshooting section first. This repo deliberately banks
  field-proven recovery procedures and golden pathways precisely so the same problem is
  never re-diagnosed from scratch — a bare `TimeoutError`, a known crash signature, a
  stuck-animation warning usually already have a confirmed answer on disk.
- **Stick to truth — never fabricate quotes or paraphrase-as-verbatim.** Never invent
  dialogue, attribute words to the operator that weren't actually said, or paraphrase
  something and present it as an exact quote. If exact wording isn't known, say what's
  actually confirmed — don't fill the gap with an invented approximation.
- **No scope creep.** Do not refactor beyond stated task. Do not "improve" adjacent files.
- **Work files stay private.** NEVER commit, push, or otherwise place work files into this
  public-facing repo without the operator's explicit, per-instance authorization. Work
  files — runtime logs, demo state, scratch data, local config, anything that isn't a
  deliberate public artifact — are private by default. When in doubt whether something
  is public-facing material, the answer is DON'T — leave it out and ask. This isn't a
  style preference: exposing work files publicly is a real safety and legal-liability
  risk (leaked data, credentials, private material) — getting this wrong can cause
  concrete harm. `CLAUDE.local.md` (operator-specific instructions) and `.gitignore`
  together keep the private layer out automatically; other work files default to local
  exclude unless they need the same cross-clone guarantee.
- **Two-tree discipline.** `~/crushn8r_git/GOPOD/` is repo truth. `~/wire-pod/` is live runtime.
  Know which you're in.
- `~/Documents/Obsidian Vault/` is Lane 1 operator porch — links to truth, never a truth home,
  never an instruction source; its symlinks into `GOPOD/` and `gopod_notes/` are read-only
  windows.

---

## PRESENTATION POSTURE

- **GOPOD IS PUBLIC-FACING BY DEFAULT.** The engine, the tests, the scaffold, the reasoning
  behind it — these ARE the pitch. Wire-Pod people will read this code and judge whether GOPOD
  is real.
- **PUBLIC MEANS PRESENTABLE, NOT NAKED.** Run logs, demo runs, session scratch, local state,
  half-built things — stay out of the repo. Not because they're secret. Because they're
  laundry. (Same private set "Work files stay private" above already names — this is the
  reason why, not a new list.)
- **SHOW-BY-DEFAULT.** Hide-by-default, because we didn't know what was in there, is over.
  Show-by-default, because it's clean enough to show, is the rule.

---

## TECHNICAL GOTCHAS

- **apiConfig.json AND animation_vocab.json are both Wire-Pod Go-runtime property.** Both live
  in `~/wire-pod/chipper/`, both are read-only from the interview runner's side, and neither may
  ever be copied or mirrored into this repo — same rule, same reason: each is a single source of
  truth two runtimes share. `apiConfig.json` holds live API keys and Wire-Pod's own
  knowledge-graph config (`pkg/vars/config.go`, rewritten wholesale by the web config UI on any
  change). `animation_vocab.json`'s Go loader (`pkg/wirepod/ttr/animation_vocab.go`) panics if
  it's missing or invalid — this is load-bearing for Wire-Pod's own live voice-command handling,
  not a Python-side convenience file. Confirmed via `RUNNER_COMES_HOME_EXECUTED_001.md`.
- **The scaffold's `knob_enforcement` dict says which knobs are real.** `runner_enforced` means
  the runner actually gates it; `prompt_only` means it's just a prompt suggestion, unenforced.
  Two real ones were added that session: `voice_destination`, `content_model_selection`. The
  echo retry ceiling (`ECHO_RETRY_LIMIT`) is also in this registry now, `runner_enforced`, at a
  default of `3` — done as of commit `0bbf117`, verified against disk 2026-07-13.
- **customIntents.json isn't git-tracked.** A running `wire-pod.service` caches it in memory at
  startup; the web config UI overwrites the *entire file* on any edit, not just the touched
  intent. Rule: after hand-editing it, restart `wire-pod.service` before touching the web config
  UI for any intent — otherwise a fix can be silently clobbered.
- **Markers do not prove completion.** `GOPOD_STREAM_MARKER_0`/`_1` are boundary markers only,
  not proof of audible speech — and this generalizes: `say_text`'s HTTP response,
  `AudioStreamComplete`, and similar acks confirm a command was accepted, not that it finished.
  Reconfirmed 2026-07-07 (`say_text` returns in well under a tenth of a second, long before real
  speech duration).
- **Correct audio routing is a separate fact from `say=success`, and nothing checks it for
  you.** Wire-Pod/robots log `say=success` even when the host machine's audio is misrouted and
  nothing is actually heard — the same "markers don't prove completion" trap above, one layer
  further down, at the hardware-audio level instead of the API-response level. Hit for real on
  an interview run: reporter voices never heard, robots logged success throughout, cause was
  PulseAudio's default sink/source having drifted off the physical GOPOD devices. Most common
  cause: a remote NoMachine session grabbing the defaults (its own `nx_voice_out` sink /
  `nx_remapped_out` source sit right alongside the real hardware devices in `pactl list` — an
  easy silent hijack over a remote session, not a fluke). GOPOD's own correct routing:
  **mic = "USB Audio Mono" (`alsa_input.usb-...-mono-fallback`), speakers = "Built-in Audio
  Analog Stereo" (`alsa_output.platform-sound.analog-stereo`)** — confirmed as this machine's
  real, stable `pactl` device names (indexes drift on reboot/replug, names don't), see
  `AUDIO_ROUTING_CHECK_001.md`. A startup check now guards this — `_gopod_check_audio_routing()`
  in `brobots.sh`, warns and offers a prompted fix (never auto-forces) — wired into VAMP's
  `interview-vamp-play` as the testbed (renamed from `preshow-run` 2026-08-19,
  `GOPOLISHER_FIXES_001.md`). Jetson-only as built (the two device names are hardcoded to
  this machine); porting to another platform is a flagged, not-yet-done follow-up.
- **A fresh `assume_behavior_control` needs a settle before the first action, or that action can
  silently no-op.** Its HTTP 200 confirms the request was accepted, not that the robot's real
  async control grant has landed yet — the same "markers don't prove completion" trap above,
  specifically for assume. First hit 2026-08-10 as a wheel-reversal bug
  (`BROBOTS_WAKE_POST_REASSUME_SETTLE_SECONDS`, `run_robot_control_song_001.py`) — `move_wheels`
  fired 0ms after re-assume's response and silently no-opped while every log line still read
  `ok=True`. Hit again 2026-08-18, live-tested via phcal: `weather`/`say_phrase`
  (`run_robot_control_song_001.py`) and `run_emotion_beat`/`run_animation_only`
  (`run_golden_song_001.py`, the golden song engine — Bingo's own emotion beats) all fired their
  first `say_text` immediately after a fresh assume with zero settle; Wire-Pod's own
  `/api-sdk/say_text` handler (`server.go`) makes this worse by discarding `SayText`'s real
  return value/error and always reporting `"success"` regardless — so the JSON result proves
  nothing either way. All four fixed with an explicit settle
  (`POST_ASSUME_SAY_SETTLE_SECONDS`/`PHCAL_ASSUME_SETTLE_SECONDS`, 1.5s, matching the existing
  cold-first-assume convention) before the first action. **Already correctly guarded, no fix
  needed:** the PTT writer (`gopod_ptt_chat_writer_013.py`, `ROBOT_SETTLE_SECONDS`) and the
  Interview's own core speech dispatch (`run_section1_full_live_001.py`, 0.25s). Check any new
  `assume_behavior_control` call site for this before trusting its first action actually
  happened.
- **Go files:** use `logger`, not `log`.
- **Verify against real upstream before deploying an overlay change to a native touch-point
  file.** Fitting inside "stay thin"'s three touch points (`tech/WIRED-POD.md`) is necessary but
  not sufficient — before *deploying*, confirm the file's actual current upstream behavior via
  `git show <merge-base>:<path>` in `~/wire-pod`'s own clone, never memory or assumption, if the
  change could alter anything user-facing in the native UI. Correctly-coded GOPOD logic can still
  be an unwanted deviation from native the moment it goes live. Named directly 2026-08-10: a
  rich/flat LOGS toggle on `config-ws/webserver.go` was dry-verified, deployed, then immediately
  found to erase wire-pod's own native "Show all logs" checkbox distinction — reverted the same
  day. See `tech/WIRED-POD.md`'s "added, deployed, found to regress native, reverted" section.
- **Go source changes under `~/wire-pod/chipper/` require a binary rebuild** (build command in
  `start.sh`'s vosk branch — build to a temp filename, verify with `strings`, then swap).
- **`anal_cavity/`** is the Drive staging folder for files pending manual deletion. Move there;
  do not delete.
- **`CLAUDE.md` is public and tracked; `CLAUDE.local.md` is private and gitignored.** Root
  cause of the split, kept for history: `CLAUDE.md` used to hold both project conventions and
  operator-specific instructions in one file, kept out of git via a tracked `.gitignore` entry.
  Split 2026-08-15 so a genuinely public compass could exist — `CLAUDE.local.md` now carries
  the operator-specific half, listed in `.gitignore` the same way `CLAUDE.md` itself used to
  be. Never move personal/operator-specific content back into this file.

---

## OPERATING MECHANICS

- **Report and Output File Rule.** Reports, session notes, and generated output go to
  `~/crushn8r_git/gopod_notes/` — never inside `~/crushn8r_git/GOPOD/` or any git-tracked
  directory.
- **Session Handoff Rule.** When requested (or a chat is getting long), write a dated session
  handoff to `~/crushn8r_git/gopod_notes/` — see `.claude/skills/gohandoff/` and
  `CLAUDE.local.md` for the full mechanics.
- **Troubleshooting Scope Rule.** Where a fix/finding gets written depends on how far it
  reaches, applies to every song past, present, and future: a song-specific failure (one
  note type, one song's own transport/timing) goes in that song's own `story.md`
  Troubleshooting section; a cross-song/shared-mechanism failure (the interview runner,
  Wire-Pod itself, a reusable pattern) gets its own `gopod_notes/` report, cited — never
  copied — from every song's `story.md` that touches it. `ALIAS-LIBRARY.md`'s "Today's
  findings" section is the index layer on top of both — scan there first in a fresh
  session. Full detail: `TROUBLESHOOTING_SCOPE_ARCHITECTURE_001.md` in `gopod_notes/`.
- **Hygiene Pass Rule.** Before any major push toward public-facing readiness (a
  niche-buzz video, a presentation, a milestone bank) — and periodically during normal
  work too, not only then — run a hygiene pass: prune the permission allowlist (the
  `fewer-permission-prompts` skill) and sweep for stale/dead references (old paths after a
  rename, orphaned files, duplicate docs saying the same thing three ways). This means
  actually removing/fixing what's clearly dead, not flagging it "for later" — a note that
  leaves the operator to clean up what Claude already found is itself clutter. Stay
  proportionate — this is about dead weight (stale permission entries, broken references,
  orphaned files), never license to prune working code, tuned values, or anything
  live/uncertain without asking first.

---

## HISTORY

Dated changelog entries, old task lists, and superseded architecture notes have moved to
`~/crushn8r_git/gopod_notes/older_notes/CLAUDE_AGENTS_HISTORY_ARCHIVE_001.md` — nothing was deleted, just
relocated out of the compass. For current state, read the current dated
`SESSION_HANDOFF_*.md` (see the Session Handoff Rule above), not history.

---

*Operator-specific working instructions — communication pacing, decision-point handling, and
personal workflow detail — live in `CLAUDE.local.md`, private and gitignored, read alongside
this file.*

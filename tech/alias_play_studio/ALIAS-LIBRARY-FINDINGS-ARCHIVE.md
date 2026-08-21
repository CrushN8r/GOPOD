# ALIAS-LIBRARY findings archive

> Older dated "Today's findings" sections from `ALIAS-LIBRARY.md`, moved here
> 2026-07-25 so that doc doesn't grow dated sections forever. Same
> archive-forward pattern the operator's own private session notes already use
> for handoffs — historical record, not current-truth reading. `ALIAS-LIBRARY.md`
> itself keeps only the most recent 1-2 dated sections; when a third accumulates,
> move the oldest here, same as this pass did.

---

## Today's findings (2026-07-16)

Canonized from the reaction-lane crash-diagnosis and golden-shape investigation, not
restated here as if freshly re-derived.

**The golden dispatch shape.** Speak an emotion-naming line ("I'm sad") as one `say_text`
call → a real 2-second pause → a separate bare animation-only `say_text` dispatch → hold
(5.0s) → release. Produced **four consecutive clean runs including `angry` twice** — the
first-ever clean `angry` results all session, after the identical clip crashed on every
prior attempt that day (a blocking-token crash, a stuck-queue-checked run that still
crashed, two embedded-in-one-sentence `say_text` HTTP timeouts of 10s/30s). Independently
verified against Wire-Pod's own debug log across the full run window — zero
`"(waiting for animation to be done...)"` lines on any of the four runs. Most defensible
read of the evidence: the explicit pause between speaking and animating is the most likely
single ingredient, since it's the most direct structural difference from every version that
crashed — but this is inference from a behavior change, not a diagnosed mechanism the way
the `AnimationQueues` finding below was proven from source. Not yet tested: whether this
generalizes to all 10 verified tokens (only `celebrate`, `love`, `angry` tried), or whether
concurrency (not just sequential pacing) is safe under this shape. See
`FOUR_CLEAN_RUNS_ANGRY_FIXED_001.md`.

**`angry` / `anim_rtpickup_loop_10` — handle with pacing, not DO-NOT-USE.** Crashed or hung
on every real attempt earlier in the day: the original blocking-token crash, a
stuck-queue-checked run that still crashed with nothing detected, and two embedded-line
`say_text` HTTP timeouts. Under the golden paced shape above, `angry` has now run clean
twice in a row, independently log-verified. Status: **handle with pacing** (use the golden
shape, don't fire it as the very first action after `assume`) — no longer a DO-NOT-USE
token, but not yet proven as reliable as the other nine across enough runs to drop the
caution entirely.

**The stuck-`AnimationQueues` mechanism.** Root cause, source-traced with high confidence:
Wire-Pod's `AnimationQueues []AnimationQueue` (`kgsim_cmds.go`) is a package-level global,
unmutexed, per-ESN "currently playing" flag with no HTTP-exposed reset. `StartAnim_Queue`
blocks forever on `for range AnimationQueues[i].AnimDone` if a prior dispatch on that robot
never called `StopAnim_Queue` (e.g. an interrupted/Ctrl-C'd run) — meaning
`robot.Conn.PlayAnimation()` is never reached, explaining "HTTP success, nothing plays."
Signature log line: `"(waiting for animation to be done...)"`. Only a wire-pod service
restart (`wpr`) clears it. **Correction on record:** this log line is not exclusively tied to
the blocking `playAnimation` token — it was also observed once for the async token with no
ill effect, so its presence alone isn't proof of the stuck-queue failure mode, only a
signal worth checking alongside actual play/no-play results. See
`ANIMATION_DISPATCH_ISOLATION_001.md`.

**The stale-shell trap.** Bash functions stay as originally defined for the life of a shell
process — re-running `source ~/.gopod_alias_lib/brobots.sh` (or a genuinely fresh terminal)
is required after any edit, or old code silently keeps running. Hit twice in one session:
the operator got a stale `hold_anim=18.8s` readout from a shell that hadn't re-sourced since
before two separate retunes. Check printed values against disk when anything looks off.
See `STALE_SHELL_CRASH_AND_ALIAS_CLEANUP_001.md`.

**systemd rate-limiting on `wpr`.** Space repeated `wpr` calls — firing it back-to-back too
fast trips systemd's `start-limit-hit` guard (confirmed not a wire-pod crash: chipper's own
log showed a clean startup, killed by systemd for another restart request mid-boot).
Recovery: `sudo systemctl reset-failed wire-pod.service && sudo systemctl restart
wire-pod.service`, run by the operator directly (the sandboxed Bash tool has no interactive
terminal for `sudo`'s password prompt).

---

## Bingo rattle — the reused golden pathway (2026-07-17)

Canonized from `BINGO_RATTLE_ADDED_001.md`, built while adding a ball-draw setup (rattle +
call + reaction) in front of each of `start-the-bingo-capture`/`bingo-video-song`'s three
emotion beats.

**The decoupled direct-SDK pattern now has two proven uses, not one.**
`DECOUPLED_DIRECT_SDK_GOLDEN_PATH_001.md`'s connect→`BehaviorControl`-wait→act→release
skeleton (built for the opening chord's "Brobots ready!" together-step, live-measured,
operator-confirmed) was copied verbatim for a second, different `vectorpb` call — the
Bingo sidecar's own rattle-audio stream (`ExternalAudioStreamPlayback`), not `SayText`.
New file: `direct_sdk_bingo_rattle_001.go`, same directory
(`~/wire-pod/chipper/gopod_probes/tools/`) as the original. That doc's own "Extending it"
section named exactly this move ("swap `SayText` for whatever other `vectorpb` call the
vendored SDK exposes") before it was ever needed — worth trusting the same template again
for the next one.

**Error 915 and 914 are not the same risk, and neither blocks this pattern.**
Re-confirmed while deciding whether the rattle shim was safe to build: Error 915's
best-evidenced cause (`VECTORX_BINGO_ERROR915_WRITECOLOREDTEXT_INVESTIGATION_001.md`,
hypothesis H5) is a **same-connection** race — rattle audio immediately followed by a
display call, zero drain delay. Error 914's dual-connection ("Wire-Pod contention")
hypothesis was **never actually tested** (`DIRECT_SDK_TOGETHER_SAY_HANDOFF_INVESTIGATION_001.md`)
— every observed 914 crash came from repeated arm/lift calls, no second connection
involved at all. Neither named fault is documented proof against a second, independent
SDK connection sharing a robot with Wire-Pod's own HTTP session — the release→wait→
direct-connect→release discipline is real caution, not a fix for a proven fault.

**A leading `"Word: "` in any scripted line gets silently eaten by
`normalize_robot_safe()`.** `run_section1_full_live_001.py`'s own generic label-stripper
(`re.sub(r"^[A-Za-z ]+:\s*", "", cleaned)`, built for the interview's echo-suppression
fix) also strips a scripted line's own leading label — found when "Ball: B-1" came out
of `say_turn` as spoken audio `"B-1"` (the on-screen `display_text` stayed correct).
Worth checking any new scripted `say_turn` line for a leading colon before assuming it'll
be spoken as written.

**Rattle's settle margin was too tight for audio, not just for the connection
(2026-07-22).** This section's own build day (2026-07-17) already flagged the open
question honestly: the release→settle→direct-connect margin
(`DIRECT_SDK_RELEASE_SETTLE_SECONDS = 1.0`) was live-proven for a one-shot `SayText`
call only, never for rattle's own longer-held audio stream. It surfaced live: a full
bingo run's `round_1_rattle` reported `status=OK` but wasn't heard (its sibling
`opening_rattle` in the same run WAS heard — intermittent, not a hard failure, consistent
with a too-tight margin). A separate constant, `RATTLE_SETTLE_SECONDS`, was added for
`run_rattle()` alone — the shared `DIRECT_SDK_RELEASE_SETTLE_SECONDS` stays untouched,
still correct for `run_brobots_ready_together()`. Value history, live-tested one step at a
time via phcal's own isolated rattle primitive: 2.0 improved but didn't fully fix it (a
full run still heard only 3 of 4 rattles); **3.0 is CONFIRMED WORKING, live, twice** —
once in a normal run, once with Brobot 1 deliberately put to sleep first (a harder
condition) — operator's own words both times: "i heard the rattle" / "i heard it". `3.0`
is now the permanent value in both phcal and `run_rattle()`. Do not lower it without a
fresh live re-confirmation.

---

## Today's findings (2026-07-23)

**Song rename executed:** `brobots_bingo_capture_001` → `brobots_bingo`,
`brobots_weather_001` → `brobots_awaken`, `brobots_preshow_001` → `brobots_preshow` (song
folders, `song_id` fields, and every alias/runner reference swept together, dry-verified
each still runs). `brobots_interview_section_01` left as-is. Same pass also archived
`brobots_bait_001`, `brobots_bait_002`, and `robot_control_song_001` into
`goverlord/runtime/songs/archives/` — see the "known broken as shipped" note on `pha0b`'s
own registry row above for the alias fallout.

**`pha0b_menu()` reworked twice more, correcting itself both times.** First pass added the
`0.` full-song shortcut and a reporter-gap y/n toggle exactly as asked. The `0.` line
initially printed at the *bottom* of the divisions list (a real placement bug, not what was
asked) and the menu still only *printed* the resulting command for manual copy/paste — both
corrected: `0.` now prints first, and the menu runs the picked slice directly. Second pass
made typing the literal `0` at Point A a genuine one-keystroke shortcut (skips Point B
entirely), where the first pass's version still asked Point B even when Point A alone
signaled full song.

**Reporter gaps zeroed permanently, not just via the run-scoped toggle.** All 6
`reporter_gap_*` steps in `brobots_bingo/knobs.json`: `pause_seconds` `5` → `0`, on the
operator's own explicit later instruction — supersedes an earlier same-session "do not
touch reporter_gap_* pause_seconds" ask. The `pha0b_menu()` y/n toggle above still exists
and still works, it just now toggles between "0s" (default) and "0s" (since the underlying
value itself is 0) — worth knowing if that prompt's continued relevance ever gets
revisited.

**One new step added:** `closing_wait` (brobot_2, "Wait!", `say_turn`) between `banter_6`
and `exit` — bingo is 46 steps now, not 45 or the long-stale "42" several docs still said
before this pass (see `WIRED-POD.md`/`GOPOD_FEATURES.md`/`BINGO.md`/`README.md`).

**A genuine, previously-unnoticed commit-integrity bug caught and fixed:** the song-rename
commit's `git add -A` call silently failed to stage 3 files (an invalid pathspec for an
already-renamed-away path aborted the whole add), so the first rename commit shipped with
the OLD `song_id` values still inside all 3 renamed `knobs.json` files, despite the working
tree being correct. Caught via `git show <commit>:<file>` rather than trusting a clean
`git status` after commit — worth a spot-check like that on any multi-file commit going
forward, not just here.

A real scope-creep incident occurred here (an unrequested new alias built and reverted) —
worth remembering before dispatching a "small add-on" task into `pha0b`/`brobots.sh` again.

**Later the same day: vamp gate, pre-show, and the bait songs all got golden-guts passes.**
- `brobots_vamp_gate/` + `run_vamp_gate_song_001.py` + `pha0b vamp` — a standalone,
  playhead-sliceable song for the pre-show's own vamp beats, evaluable against every other
  song on equal footing. `brobots_preshow` itself untouched.
- `start-the-preshow-song` — the pre-show's real scored song (`run_preshow_song()`) finally
  has an alias that plays it end to end; every prior pre-show alias only ever called
  `generate_phase()` directly.
- `brobots_bait_001` and `brobots_bait_002` un-archived back to top-level `songs/`, which
  incidentally fixed `start-the-bait-song`/`pha0b bait` (broken since the 2026-07-22 archive
  move, "not a big deal now" at the time — see the "Known broken as shipped" note above).
  `brobots_bait_001` also got its first-ever alias, `start-the-net-song` (built on the
  interview engine, not `pha0b`-sliceable — see its own registry row above).
- Two stale `story.md` cross-links fixed as a side effect (`brobots_awaken`,
  `archives/brobots_bait_002` at the time).
- `robot_control_song_001` deliberately left archived, untouched — `start-the-control-song`/
  `test-arm-cue`/`test-head-nod`/`test-fireworks` remain broken, out of scope for this pass.

**Full `.bashrc`/`.bash_aliases` cross-check.** Confirmed every file either loads (`core.sh`,
`brobots.sh`, `wirepod_logs.sh`, `demo.sh`, `llm.sh`, `openwebui.sh`, `suits.sh`,
`goverlord.sh`, `chat_capture.sh`, `frame0.sh`, all via `.bash_aliases`'s loader loop or a
direct `.bashrc` source line) matches this doc's own file coverage — no untracked source
file, no phantom/retired entry claimed live that isn't. Three real gaps found and fixed
here: `bingo-video-song-live`, `bingo-video-song-pick-segment` (both `brobots.sh`, both
undocumented since they were added), and `gopod-index` (`demo.sh`, undocumented since
built). All three were live and correct in the shell the whole time — this was a doc-drift
gap, not a functional one.

---


# PHCAL — PLAYHEAD, Part 1: the calibration bench

> Fire one thing, on one robot, and watch it happen. Not a performance — a bench test.

Two docs, one system: this is PLAYHEAD's tuning half. The performance half — whole
songs, start to finish — is [PHA0B.md](PHA0B.md), PLAYHEAD Part 2.

---

## What phcal is

PHCAL — PLAYHEAD Calibrations — is where a single robot behavior gets tuned in isolation,
away from any scored song. Pick one primitive off the menu, pick the robot, watch it fire,
adjust, fire again. Nothing here plays a song end to end; that's [PHA0B.md](PHA0B.md)'s job.
What gets confirmed here is exactly what pha0b's own "apply phcal tweaks?" step later drops
into a song's own file.

## The bench, eight groups, fourteen primitives

**Regrouped 2026-08-16** (`PHCAL_MENU_REGROUP_BUILT_001.md`) — pure display change, every
primitive's own identity string is exactly what it always was underneath. Bare `phcal`
opens a top-level menu of 8 human-labeled groups (plus `0. exit`); a group with one member
routes straight through, same as the old flat menu always did for those; a group with 2-3
members opens one more numbered sub-menu showing the real primitive names, then routes the
same way. `0` (or anything out of range) exits cleanly at both levels.

| Group | Label | Members | What each does | Target |
|---|---|---|---|---|
| 1 | info | `robot_info` | Read-only battery/version/protocol snapshot — no control assumed, nothing moves | both, always |
| 2 | movements | `arm`, `nod`, `move_reverse` | Arm-lift cue (rest→up→rest, N cycles) / head nod, one or more reps / a fixed on-charger reverse wheel pulse | one robot |
| 3 | vector's cube | `cube` | Connect → all four corner LEDs red → hold → all corners green → release. Net-new 2026-08-15, cube keeper is Brobot 2 (ESN `0dd1d8bf`) — fails cleanly, doesn't hardcode the robot, if fired at one with no cube paired | one robot |
| 4 | audio | `rattle`, `danger` | Fires the Bingo sidecar's own rattle sound / GOPOD's own `danger-will-robinson.wav` — same direct-SDK binary and settle margin, different WAV path, added 2026-08-12 so a song can fire it directly rather than only through the LLM-gated chat path | one robot |
| 5 | animations | `animation` | One of three catalog animations (a knowledge-graph success, a searching loop, an answering loop), or 0 to fire all three in sequence with a settle pause between each | one robot |
| 6 | brobots 1 | `weather`, `brobots_announce_in_sync`, `brobots_stay_in_place` | Picks a robot, fetches the real forecast, and has that robot speak it with its own per-robot unit/clock format / both robots saying a synced line together, over the same low-level path the pre-show sync uses / holds behavior control for a stretch and watches for drift | varies (weather: no robot control at all; the other two: one robot / both) |
| 7 | brobots 2 | `brobots_sleep_to_wake_direct_sdk`, `brobots_session_responsiveness` | Bench test of a full sleep-to-wake cycle over one held connection / the golden-flag wake pulse — assume, release, settle, reassume | 1, 2, or both / one robot |
| 8 | tempo | `tempo` | A song's pacing — global ease or one step's own factor — picked and confirmed here, written by the same tool `tempo-set` wraps | none |

`brobots_stay_in_place`/`brobots_sleep_to_wake_direct_sdk`/`brobots_session_responsiveness`/
`brobots_announce_in_sync` (now split across groups 6 and 7) are the four primitives that
were briefly set aside and later brought back under clearer names, well before this
regroup — the actual mechanism behind each one never changed, only what it's called and
which group it now displays under. `tempo` (group 8) is the one entry with no robot at all
behind it — it walks a song, a mode, and a value, then hands the actual write off to
`tempo_set_001.py`, the same tool the standalone `tempo-set` alias already wraps. Picking a
song is new territory for this bench (every other group works on a robot directly, never a
specific song's own file) — this is the one place phcal reaches into `knobs.json` itself.

## The battery gate

Before any of the four items that actually move a motor (`arm`, `nod`, `animation`,
`move_reverse`) fires, phcal checks the robot's own battery reading first. Below roughly
3.7 volts, it refuses — doesn't warn and continue, refuses outright — and nothing gets sent
to the robot. `rattle` is sound only, so it's deliberately left out of that gate; so are the
read-only and control-channel items, since none of them ask a motor to move.

Worth saying plainly: the gate's logic has been checked against real captured voltage
readings from both a healthy and a faulted robot, but the refusal itself has never actually
fired live against a robot that was genuinely low — there's no safe way to force that on
demand. It's proven on paper, not yet proven in the room.

## One shape, two different mechanisms underneath

Robots get put to sleep and woken back up two structurally different ways in this repo — one
over the same web path a song's own dialogue uses, one over a single held low-level
connection. They used to report success or failure in two different shapes, which made
"is this robot actually ready" a different question depending on which path answered it.
Now both report the same shape — ready or not, why, which path answered, and any extra
detail — so a caller never has to know or care which mechanism actually ran underneath.

## Writing a tuned value back

Once a value's been confirmed on the bench, it's remembered — cycles, hold time, speed —
separately from any song. Pointed at a specific step inside a specific song's own file, that
remembered value overwrites just that one step's own numbers and nothing else, leaving every
other step untouched. This is the exact same write pha0b's own "apply phcal tweaks?" prompt
triggers automatically across a whole picked range — one tool, two doors in.

## robot_info, the one read-only entry

Group 1 (direct, no sub-menu) never asks for control and never moves anything — it's a
snapshot: battery, firmware version, protocol version, for both robots at once, every
time. There's no dry version of it to fall back on; if the live gate is off, it says so
plainly and refuses rather than pretending to simulate an answer it doesn't have.

## Honest state of this bench

The current shape of this menu — the 8 groups, the battery gate, the renamed items — is
real and working code, checked by hand and by machine, but it hasn't been fired live end
to end since it was last reshaped. Treat this document as describing what's built, not a
claim that every one of the 14 primitives has been freshly proven live today. `tempo`
(group 8) has been dry-verified only — both modes, no write — same as the rest of this
document's own caveat. `brobots_sleep_to_wake_direct_sdk` (group 7), live-fired both
robots: they complete roughly 1-2 seconds apart, not simultaneously — both finished OK,
this is just observed timing, not a fault.

## Detect-first, live-tested 2026-08-18

phcal now probes every configured candidate ESN at startup (reusing `robot_info`'s
own read-only binary call) and shapes the session to what actually responds: `none`
(dry-only), `single` (one robot — the menu stays full, robot-targeting auto-resolves to
the one present robot, and single-robot wake finally becomes the default instead of
always "both"), or `multi` (2+ — an explicit confirm prompt now asks to continue in
multi, downgrade to a dry `none` run, or force `single` on one specific robot). Any
present robot reading below `BATTERY_MIN_VOLTS` (3.7V) warns and defaults to not
proceeding. See `gopod_notes/PHCAL_DETECT_FIRST_001.md` for the full build.

Two real findings from that live-testing pass, both fixed:

- **The startup probe itself could read as frozen.** It originally shared
  `ROBOT_INFO_TIMEOUT_SECONDS` (30s) with the deliberate, occasional "1. info" pick —
  fine for a check a human explicitly asked for, not fine for something that now fires
  silently on every launch. An absent robot gave zero output for up to 30 real seconds.
  Fixed: a separate, shorter probe timeout (`PHCAL_DETECT_PROBE_TIMEOUT_SECONDS`, 8s) plus
  an immediate "probing..." line, so waiting is visible and bounded.
- **`brobots_announce_in_sync`'s single-robot degrade spoke on-screen but produced no
  audio.** Root cause and the general rule it revealed are in `CLAUDE.md`'s own TECHNICAL
  GOTCHAS ("A fresh `assume_behavior_control` needs a settle before the first action") —
  this bench is where it was actually found and live-confirmed. Fixed here (the `weather`/
  `say_phrase` notes) and in the golden song engine's `run_emotion_beat`/
  `run_animation_only` (Bingo's own emotion beats), same explicit settle in both.

## Where these files actually live

Three homes, not one, and only one of them is backed up the normal way.

- **This repo, git-tracked, public.** `knobs_envelope_001.py` — the resolver every tool below
  reads a song's *current* values through (dirty `zKnobs.json` over clean `knobs.json`,
  whole-file, no partial reads) — lives in `goverlord/runtime/songs/tools/`, alongside
  the playback engines pha0b uses. `run_section1_full_live_001.py` also lives there; phcal
  loads it purely to reuse its `Robots`/HTTP-helper class, not to play anything.
- **`~/.gopod_alias_lib/`, a plain folder outside this repo, no git repo at all.** Everything
  that actually *is* phcal: `phcal()` itself (a bash function in `brobots.sh`),
  `phcal_isolate_001.py` (all 14 primitives, their 8-group menu, and their dispatch logic),
  `phcal_last.json` (the memory file — widened 2026-08-16, `MASTER_TWEAKS_STAGE1_SAVE_COVERAGE_001.md`,
  to 9 saved primitives: `arm`, `nod`, `rattle`, `danger`, `animation`,
  `brobots_stay_in_place`, `move_reverse`, `brobots_announce_in_sync`,
  `brobots_sleep_to_wake_direct_sdk` — each reads its last-confirmed values from and
  writes back to this same file; `robot_info`/`weather`/`tempo`/`cube`/
  `brobots_session_responsiveness` have no tunable value to remember, so they never touch
  it), `phcal_apply_001.py` (the write-back pha0b's own "apply phcal tweaks?" prompt
  calls), `tempo_set_001.py` (tempo's actual write, shared with the standalone
  `tempo-set` alias), and `robot_pick_001.py` (a small sibling writer for reassigning a
  step's speaker). All of it real, all of it exercised daily, none of it committed or pushed
  — a loss of this one folder would mean rebuilding the whole bench from scratch, not
  checking it out from history.
- **`~/wire-pod/chipper/gopod_probes/tools/`, the live Wire-Pod runtime tree — a third
  location, neither of the above.** The compiled direct-SDK binaries `rattle`, `danger`,
  `robot_info`, `brobots_sleep_to_wake_direct_sdk`, `brobots_announce_in_sync`, and `cube`
  actually shell out to live here, not in this repo and not in `~/.gopod_alias_lib`. Built
  from source kept in this repo's own `SDK/` tree, but the binary itself only exists on this
  one machine. `brobots_session_responsiveness` is the one control-channel primitive with
  no binary at all behind it — it's pure Wire-Pod REST (assume, release, settle,
  reassume), nothing direct-SDK.

## How phcal and pha0b work together

Two different jobs, sharing state through exactly two doors. Phcal tunes one primitive on one
robot, in isolation, away from any song at all; pha0b performs a song's whole score, start to
finish — see [PHA0B.md](PHA0B.md).

The first door is pha0b's own "apply phcal tweaks?" option — it reads whatever values were
last confirmed here and drops them into the picked range before the song plays. The second is
newer: this bench's own `tempo` item (12) reaches into a song's pacing directly, landing on
the exact same write the standalone `tempo-set` alias uses — tuning tempo from either door
writes the same place. Both boards, and every tool either one calls, resolve a song's
*current* values through the same `knobs_envelope_001.py` — neither one can ever see a
different "current" than the other.

## See also

Playing a whole scored song, start to finish, happens on a different board — see
[PHA0B.md](PHA0B.md). Slowing or quickening a song's own pacing permanently is reachable two
ways now — group 8 (`tempo`) on this bench, or the standalone `tempo-set` alias directly — both land
on the exact same `tempo_set_001.py` write, so neither door can drift out of sync with the
other.

---

> From Doctrine Barfallonyou
> Lesson! Confidence on stage comes from boredom on the bench.
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
- [ALIAS-LIBRARY.md](ALIAS-LIBRARY.md) — every shortcut command GOPOD can run
- [ALIAS-SEQUENCER.md](ALIAS-SEQUENCER.md) — arrangement: notes into sequences into songs
- [PHA0B.md](PHA0B.md) — PLAYHEAD Part 2, the performance front door, whole songs at a time

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

# PHA0B — PLAYHEAD, Part 2: the performance front door

> Pick a song. Pick where it starts and stops. Set a few options. Hit play. One control panel,
> every song on the shelf.

Two docs, one system: this is PLAYHEAD's performance half. The tuning half — one
primitive, one robot, in isolation — is [PHCAL.md](PHCAL.md), PLAYHEAD Part 1.

---

## What pha0b is

PHA0B — PlayHead A/0/B — is the one place an operator sits down to actually *run* a song.
Before it existed, every song had its own separate launcher with its own separate habits.
Now there's one board: pick the song off the shelf, pick a starting point and an ending
point, answer a short run of options, and it plays — live on the robots or dry on the
terminal, same choice, same board, every time.

It doesn't write the song and it doesn't tune a robot's arm. It performs what's already
written, exactly as much of it as you ask for.

## Picking the song

Bare `pha0b` opens the menu. The list isn't hand-typed anywhere — it's read straight off
disk, every folder under the songs shelf except the archive container. Today that's six
numbered entries: the Awaken bait/capture video, the Interview (playable through this
board too, its own way — see below), Bingo's capture song and its live-game
launcher sibling, Baby Robots Sleep, and the "is that you?" live-capture demo.

Four more songs exist behind the curtain, reachable only by typing their keyword
directly rather than picking a number: the archived control-song self-check, the archived
pure-weather song, the pre-show vamp filler, and the archived cross-persona mix-up reel
(`mixup` — superseded by the live "is that you?" demo above, which does the same bit for
real). They still work — they're just not on the menu, since their folders moved into
the archive.

Pick the Interview off the menu and pha0b doesn't slice it — it has no start/stop points,
generated and performed as one continuous piece by its own engine, not stepped through
note by note the way every other song here is. Instead it asks a different question:
vamp this take (roll a new one), perform the last take, or go one-shot — generate and
perform right now, the same single call the old standalone launcher used to make.
`pha0b interview` (the direct-keyword form, no `<a> <b>` needed) asks the same question —
one door, either way you open it.

## Picking the range

Point A, Point B. If the song is organized into named sections (Bingo's rounds, for
instance), you pick sections by number. If it isn't, you pick raw steps by number instead.
Either way, `0` — or just hitting Enter, or typing a number that doesn't exist — means the
whole song, start to finish. Pick a real Point A and it asks for Point B next, and won't let
B land before A. This "zoom in/out" (playing only a slice instead of the whole song)
re-confirmed live 2026-08-16 against Awaken's own `story.md` (`arm_test`..`weather`,
steps 4-10 of 16) — reporter-gap override honored, phcal tweaks applied cleanly to just
the picked slice, running on the current `run_golden_song_001.py` engine, not just the
mechanism's original control-song home.

## The options, in the order you actually see them

Every one of these is a yes/no or a short number, asked once per run, and every one has a
sane default so a bare Enter through all of them reproduces the old always-on behavior.

1. **Reporter gaps** *(bingo and the awaken/bait song only)* — apply this run's own silent
   pause length, or override it to something else just for this run? Saying no zeroes every
   gap for this run alone; nothing gets written to the song's own file either way.
2. **Live robots?** — default yes. Say no and every step still runs, on the terminal, with
   no robot ever contacted.
3. **Rich display?** — default yes. Turns the on-screen chat window on or off for this run.
4. **Apply phcal tweaks?** *(bingo and bait only)* — walks every arm/nod step inside your
   picked range and drops in whatever values were last confirmed on the calibration bench
   (see [PHCAL.md](PHCAL.md)). Say no and the song plays with whatever's already saved.
5. **Reassign a speaker?** *(most songs, not bait — bait's whole run is one robot's own
   choice, not a per-line pick)* — swap which robot says a given line, on the spot.
6. **Which robots?** *(the newer engine's own songs)* — 1, 2, or both. Skip a robot entirely
   and its lines get logged as filtered, but the pacing around them doesn't shift — the
   silence still takes the same amount of time the line would have.

## The playback filter and the exemptions

That last "which robots" answer is a real filter, not a suggestion — pick "1" and every step
whose speaker is Brobot 2 gets skipped outright, not spoken quietly. Four kinds of step are
exempt no matter what you pick: the shared wake, the "brobots ready together" sync line, any
silent pause, and the closing exit — all four either touch both robots by nature or touch
neither, so a single-robot filter would make no sense against them.

One separate wrinkle, only on songs built the older way (today, just the Awaken/bait song):
the *file's* own per-line speaker field is dead weight there. The whole run picks one robot,
once, for everything — that's a different knob than the filter above, and the two don't
overlap.

## Two different ways a run gets slower

A song can breathe two different ways, and it's easy to mix them up. **Tempo** lives inside
the song's own saved file — a global number times a per-step weight, read once when the run
starts, still there the next time you play it. **The reporter-gap override**, above, is the
opposite: a run-only choice that's gone the moment the terminal closes, never touching the
file at all. If a song feels slow and you don't remember setting anything, tempo is probably
why — check the song's own knobs before assuming a run-time option did it.

## The one song that isn't a song

Pick the Bingo *live game* entry off the menu and none of the above happens at all — no
range picker, no options, nothing read from a knobs file. That entry just launches the real
game binary directly, the same one voice-triggered "go bingo" launches. It's a genuinely
different kind of thing sharing a slot on the same menu: a live, running game, not a scored
performance. It has no dry mode of its own, either — picking it fires for real, every time.

## What a run leaves behind

Every played run — live or dry — writes two files: a plain running log of everything that
printed, and a structured record of every single step (what it was, who said it, whether it
worked, how long it took), closing with a ranked list of the biggest pauses in the whole run.
Nothing about a run vanishes once it's finished; it's all sitting in that song's own folder
afterward.

## Where these files actually live

Two homes, and they don't get the same safety net.

- **This repo, git-tracked, public.** The engines that actually step through a song and play
  it — `run_golden_song_001.py` (Bingo, the cross-persona mix-up, Baby Robots Sleep),
  `run_robot_control_song_001.py` (the Awaken/bait family), `run_vamp_gate_song_001.py` (the
  pre-show filler), and `run_section1_full_live_001.py` (the Interview's own two-phase
  engine — also the shared library every other engine above imports its HTTP/robot helpers
  from) — all four sit in `goverlord/runtime/songs/tools/`. Every song's own
  `knobs.json`/`story.md` lives in its own folder under `goverlord/runtime/songs/`. All of
  this is backed up, versioned, and pushed the normal way.
- **`~/.gopod_alias_lib/`, a plain folder outside this repo, no git repo at all.** `pha0b()`
  and `pha0b_menu()` themselves — the entire board this document describes — are one bash
  function living in `brobots.sh`. Real, working code, exercised every day, but nothing here
  is ever committed or pushed anywhere. If this one file were lost, pha0b would need to be
  rebuilt from scratch, not checked out from history.

## How pha0b and phcal work together

Two different jobs, sharing state through exactly two doors. Pha0b performs a song's whole
score, start to finish, on the robots it picks; phcal tunes one primitive on one robot, in
isolation, away from any song at all — see [PHCAL.md](PHCAL.md).

The first door is pha0b's own "apply phcal tweaks?" option, above — it reads whatever values
were last confirmed on phcal's bench and drops them into the picked range before the song
plays. The second is newer: phcal's own bench now has a `tempo` item that reaches into a
song's pacing directly, landing on the exact same write the standalone `tempo-set` alias
uses — so tuning tempo from either door writes the same place. Both boards, and every tool
either one calls, read a song's *current* values through the same shared resolver
(`knobs_envelope_001.py`, in this repo's own tools folder) — a dirty, in-progress
`zKnobs.json` always wins over the clean `knobs.json` underneath it, for pha0b's playback and
phcal's writes alike, so neither board can ever show a different "current" than the other.

## See also

Tuning one robot's own arm/nod/wake behavior happens on a different bench — see
[PHCAL.md](PHCAL.md). Slowing or quickening a whole song's pacing permanently is reachable
from either that bench (item 12) or the standalone `tempo-set` alias directly.

---

> From Doctrine Barfallonyou
> Lesson! A control panel isn't the show. It's just the thing that lets the show start on time.
> Boom! Done! Class Dismissed!
> — Doc Squawkadoodle

---

## GOPOD YAHMM (You Are Here Mall Map)

Part of GOPOD — see [tech/README.md](../README.md) for everything else in this folder, or [the root map](../../README.md) for the rest of GOPOD.

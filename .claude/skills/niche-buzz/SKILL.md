---
name: niche-buzz
description: Use when a fresh session (Claude, Claude Code, Claude Design, or a parallel chat) needs to know where GOPOD's niche-buzz campaign push stands — the mission, the funnel, the song shelf, what's banked vs pending — without the operator re-explaining it. A "YOU ARE HERE" flow guide for the CAMPAIGN, not the song work itself (see studio for that). A map, not a driver: it points, the operator steers, and the operator's live word always outranks this file.
---

# Niche buzz

The studio skills (see `studio`'s index) govern the song work — reading a score, dry
verification, hardware calibration, committing safely. This skill overlords the
CAMPAIGN one level up: the push to get GOPOD in front of the Wire-Pod/Vector
community first, as the spark for a wider social push. Read this to get oriented before
touching campaign work; read `studio` to get oriented before touching a song.

## 1. The mission

GOPOD's niche-buzz push targets the Wire-Pod / Vector community first — social proof
from the hardest crowd to impress is the spark for the wider social push. Positioning
(per operator direction — a community-standing claim, not a code dependency; confirmed
via `gopod_notes/older_notes/COMMUNITY_LINEAGE_SWEEP_001.md` that Cyb3rVector has zero
footprint in this repo, so this is peer/lineage framing, not an attributed dependency):
Wire-Pod gave Vector a voice; Cyb3rVector gave Vector a classroom; GOPOD gives Vectors a
stage. Tone: sharer, not founder — "I had the opportunity to do what I did, and that's
worth sharing, long term."

## 2. The water flow — the campaign's complete map

Per operator direction throughout — campaign intent, not a repo-verifiable fact; recorded
here so it isn't re-derived each session. The full stack, top to bottom, the operator's
own words for the shape: "the safe foundational lowest level water guided flow." Every
layer connects down through the trust layer at the bottom, not sideways:

1. **Brobots wire-pod layer** — niche buzz itself: the songs, the videos, this repo.
   Section 3 (the funnel) and section 4 (the song shelf) below are this layer's own
   detail.
2. **"GOPOD Yourself"** — the hinge: the wake phrase, Wire-Pod intent into the GOPOD
   layer, and the closing line of every video.
3. **GOPOD layer** — the house the videos tease: multi-chat with Doc & Pip, CHALK &
   PLAYHEAD (see §6, persona map). The songs carry a hidden marketing purpose — their
   titles are questions the GOPOD layer answers ("Is that you?", "Bingo?").
4. **Livestream content generator** — "AI WordPlay! Explain the Math!": yearly contest
   rounds, topic-driven for merch, Super Chat submissions fund the content directly.
5. **Math Aftermath News** — `mathaftermath.crushn8r.net`.
6. **The domain network** — `crushn8r.net` (home, a link-tree "YOU ARE HERE" mall map),
   `crushn8r.com` (landing pages, UTM/affiliate routing, weekly blog sourced from the
   newsletter), subdomains (niche pillars feeding Pinterest).
7. **Contact trust layer** — `crushn8r.ca`: emails, the CRUSHN8R CREW'd Newsletter.

## 3. The funnel

The campaign's one-page truth (layer 1 of the water flow above, in its own detail).
Canonical funnel-role map: `gopod_notes/FUNNEL_MAP_001.md` — reconciled here 2026-08-01,
see `gopod_notes/FUNNEL_MAP_RECONCILED_001.md` for what changed and why.

- **Hook/spine — "GOPOD Yourself."** The wake phrase, Wire-Pod intent into the GOPOD
  layer (confirmed live across `README.md`, `tech/WIRED-POD.md`,
  `tech/GOPOD_FEATURES.md`, `tech/alias_play_studio/SONG_02_BROBOTS_1_2_INTERVIEW_RUN.md`), and the closing line of every video.
- **BAIT — 00 brobots_awaken.** The GitHub niche-buzz intro video, one ~90-second news
  flash. `goverlord/runtime/songs/00_brobots_awaken/`, timed/scripted. Weather is a
  feature inside this song (the real per-robot weather fetch note), not a separate song
  or its own primary identity — resolved 2026-07-24, no more open question about it.
  Confirmed: its own `story.md` states it reuses `robot_control_song_001`'s mechanism,
  as-is — "the Awakening" is the operator's own campaign name for this piece, not a name
  found in the song's own files. Closes the core loop (Work order, below).
  - Historical note, kept for record, not the same slot: an older "Flash 2 =
    Is-That-You" BAIT concept describes a *different* thing — a genuinely live
    push-to-talk demo/alias, deliberately not a song, its liveness the hidden credential
    for the tech crowd. Not the `102 brobots_cross_persona` song below (UPSELL 2) —
    separate slot, separate mechanism.
  - The bait video also serves the repo's short-demo slot.
- **NET videos (two) — 01 brobots_interview_vamp + 02 brobots_interview_run.** Linked on
  GitHub to a posted x.com video; finished *after* 101. The flagship, split 2026-08-19
  into a 2-video playlist (`gopod_notes/INTERVIEW_VAMP_SPLIT_001.md`, operator's own
  framing call, `GOPOLISHER_FIXES_001.md`): video 1 is the pre-show banter
  (`goverlord/runtime/songs/01_brobots_interview_vamp/`), video 2 is the seven-exchange
  interview itself (`goverlord/runtime/songs/02_brobots_interview_run/`), Brobot 1 & Brobot 2.
- **UPSELL 1 — 101 brobots_bingo.** The proving ground. Newest golden, banked at commit
  `298e388`, `goverlord/runtime/songs/101_brobots_bingo_test/`. Opens the core loop (Work
  order, below).
- **UPSELL 2 — 102 brobots_cross_persona.** "Leaked GOPOD-layer test footage" framing.
  `goverlord/runtime/songs/102_brobots_cross_persona/`. Sits outside the core loop, its
  own standalone piece.
- **UPSELL 3 — 103 brobots_baby_robots_sleep.** An older song composed *before* GOPOD —
  a different kind of piece: revived After Effects / trippy-vibe lyric video, song
  track + possible ACE SFX + an `.srt` for timed lyrics (named intent only, not built).
  `goverlord/runtime/songs/105_brobots_nap/` (renamed/renumbered 2026-08-18 from
  `104_brobots_baby_robots_sleep/`). Sits outside the core loop,
  its own standalone piece — held for last, per operator direction, not a GOPOD song.
- **Work order (core loop).** 101 → 01 → 00 closes the loop. 102 and 103 sit outside
  that core, as their own pieces.
- Reporters (Brobots 3 & 4 voices) carry every news story — value points revealed with
  zero technical words.
- **Reporter framing — every song, runtime default (strengthened 2026-08-01).** Every
  GOPOD song now carries, by default, at least a 0-second intro/outro reporter wrapper
  in its own runtime score — the slot lives in the song's own `knobs.json`/`story.md`,
  not only at the edit layer. 103 is not an exception: its song predates GOPOD and
  carries no reporter content of its own, but the song-video wrapper still holds
  intro/outro reporter slots, set to 0 seconds — a zero-length instance of the same
  rule, not a carve-out (real on disk: `reporter_gap_intro`/`reporter_gap_outro`, both
  `pause_seconds: 0`). `00` and `01` carry no `reporter_gap_*` step yet —
  intended-later, per operator: `00` arrives with a planned full recompose, `01`
  arrives once 101's golden experience is carried over (its runner has no pause/gap-note
  mechanism at all today). `101` (6 gap steps, none literally intro/outro-positioned)
  and `102` (1 mid-only gap, no intro, no outro) don't cleanly present a literal
  intro+outro pair either — flagged, not resolved, timing firewalled. Full comb-through:
  `gopod_notes/REPORTER_WRAPPER_RULE_001.md`.
- **The standard closer** — every song-video ends the same two-beat way. Beat 1, Doc's
  Take: the one-line lesson verdict, then the hard stop — "Boom. Done. Class is
  dismissed!" Doc slams the book shut; the audience knows the show is over. Session
  end-marker and signature catchphrase in one (lineage: "Class over." — existing brand
  doctrine; "Doc's Take" formalizes it). Beat 2, Pip's door: the soft CTA. Class
  dismissed, but Pip lingers and leads — "...wanna know how to GOPOD Yourself?" Doc
  closes the show; Pip opens the door (lineage: Pied Piper — Pip's established subtle
  call-to-action role). Same ending every video: running gag, callback engine, binge
  recognition, clean edit point, deterministic end-marker for live sessions and the
  `playhead` skill alike.

## 4. The song shelf

Five songs, cockpit-confirmed (`pha0b_menu()` prints these off
`goverlord/runtime/songs/`, `zzz_archives` excluded — see `goverlord-desk/SKILL.md` §3
for full per-song status). Funnel role per `gopod_notes/FUNNEL_MAP_001.md`:

| Song | Directory | Funnel role | Note |
|---|---|---|---|
| Awakening | `00_brobots_awaken/` | BAIT | connect, arm test, head nod, a real per-robot weather fetch (a feature inside this song, never its own shelf entry), self-ID payoff, exit |
| Interview Vamp | `01_brobots_interview_vamp/` | NET video 1 | The pre-show banter, video 1 of 2 — fires standalone via `interview-vamp-play` (zero interview generation triggered) |
| Interview Run | `02_brobots_interview_run/` | NET video 2 | The flagship's seven exchanges, video 2 of 2 — fires standalone via `interview-replay` (`interview-run` adds an optional full-run mode that plays the vamp first) |
| Bingo capture | `101_brobots_bingo_test/` | UPSELL 1 | Newest golden, banked at `298e388` |
| Baby Robots Sleep | `105_brobots_nap/` (was `104_brobots_baby_robots_sleep/`) | UPSELL 3 | "Do Baby Robots Dream?" — Doc's origin, an older piece composed before GOPOD; held for last, not built into the campaign push yet |

`102_brobots_bingo_game/` ("Chocolate Bingo") sits alongside the shelf as the live game
itself, not a scored song — see `102_brobots_bingo_game/story.md`'s own "Two bingo songs,
compared" table for how it relates to the Bingo capture song above (also pointed to from
`tech/alias_play_studio/SONG_101_BROBOTS_1_2_BINGO.md`, which doesn't hold the table itself).

Two lineage entries, not shelf songs — kept for campaign history, not erased:
- `robot_control_song_001/` (now archived at `zzz_archives/robot_control_song_001/`) —
  the source mechanism Awakening was built from. Not its own shelf entry.
- `102_brobots_cross_persona/` ("Cross-Persona", UPSELL 2 — archived 2026-08-12, now at
  `zzz_archives/102_brobots_cross_persona/`) — a scripted "leaked GOPOD-layer test footage"
  demo reel, superseded once the real, live `103_gopod_is_that_you` PTT+LLM test proved the same
  bit for real. Reachable via the `mixup` pha0b keyword, repointed not retired. Full record:
  `gopod_notes/LAYER_NAMING_SWEEP_AND_102_ARCHIVED_001.md`.

**Split 2026-08-19** (`gopod_notes/INTERVIEW_VAMP_SPLIT_001.md`): the interview split
into two standalone, independently-fireable song folders for a 2-video playlist —
`01_brobots_interview_vamp/` (video 1, the pre-show banter) and
`02_brobots_interview_run/` (video 2, the seven-exchange interview itself). There is now
a dedicated end-to-end alias for the vamp alone: `interview-vamp-play` (pure "play video
1," zero interview generation triggered, renamed from `preshow-run` 2026-08-19) — a
sibling to `interview-vamp` (rolls a fresh take, WITH generation running alongside,
renamed from `vamp-run` same day). `gopod-vamp` still exists too, as the lighter
filler-beats preview it always was. **Framing decided 2026-08-19**
(`gopod_notes/GOPOLISHER_FIXES_001.md`, operator call): two NET videos, not one role
across two files — the song-shelf table above and the funnel bullet reflect this.

Bingo's timestamped terminal output (the `[HH:MM:SS.mmm +X.XXXs]` sheet-music view,
`timing_prefix()` in `run_section1_full_live_001.py`, scoped via `console_timestamps`)
is one proven piece of the growing golden-song boilerplate. The full, currently-known shape —
status markers, engine choice, the reporter-gap convention, the golden toolbox, per-song
status against all of it — now lives in one place:
`tech/alias_play_studio/SONG_SCAFFOLD.md`, a starting-point reference, not a locked spec.

## 5. Evergreen doctrine

Per operator direction. The brobots wire-pod layer (water-flow layer 1) is a
**finishable product** — once songs are golden and videos shot, it stands evergreen and
works unattended. Every new Vector owner, new Wire-Pod install, or Anki-relaunch wave
passes over the same standing content — it doesn't need re-shooting per wave. The only
exception: niche-buzz UPDATES, delivered as WAWN NEWS (see §7's Parked list).

**The ladder.** Dinner before dessert RECURSES — it's a ladder, not a line. Every
dessert, once begun, becomes the next dinner: brobots wire-pod layer (dinner) → GOPOD
layer (dessert); GOPOD layer (dinner) → livestream flywheel (dessert); livestream
(dinner) → domains/newsletter/shop (dessert) — down the water flow (§2) to the trust
layer. Same rule at every rung: prove the foundation, finish it properly once, let it
go evergreen, then build the flair above it. Each finished dinner is a plateau the next
course stands on. "GOPOD Yourself" is the door between each dinner and its dessert, at
every level.

## 6. Persona map

Per operator direction. **Naming note, correction 2026-08-12:** in the brobots wire-pod
layer, Brobots 1/2 are called Brobot 1 and Brobot 2 — plain. "Doc"/"Pip" are GOPOD-layer
persona names for the same hardware, a different, future layer. Expected overlap between the
two layers is exactly why this distinction is held deliberately, not blurred — this section
describes the GOPOD layer's own cast, where Doc/Pip is the correct name; wire-pod-layer song
docs use Brobot 1/Brobot 2 instead. Brobots 3/4 are
voice-only (`af_bella` female, `am_puck` male — confirmed in
`goverlord/runtime/songs/01_brobots_interview_vamp/story.md:5-6`), reporters at the video
layer. In the GOPOD layer (water-flow layer 3), the same two voices later carry
different scores: female voice → CHALK (content/display reporting; precedent — a
female piper voice already served as CHALK's voice before the display was ready), male
voice → PLAYHEAD (position/timing reporting; the committed `playhead` skill is that
persona's own conceptual seed, per its own text).

Principle, verbatim: **voices are instruments; personas are sheet music.** They remain
reporters in every layer — the beat changes, the reporting doesn't.

**GOPOD layer, staged truth** (snapshot, not an architecture doc). The GOPOD layer's
cast and senses are staged beyond the two Vectors: two Cozmos (intended bodies for the
Cache PYC and Git Repo personas — laptop-owned, SDK/Wi-Fi identities mapped, live
control NOT yet proven), Moorebot Scouts (Jetson-owned, source staged, live control NOT
yet proven), and a Coral USB Edge TPU vision lane (proven stable as an isolated
perception service: cameras rotate, the TPU consumes frames and emits recognition
packets to the 8011 cockpit — video is for humans, frames are for the TPU, packets are
for GOPOD). Nothing here is claimed live-capable without its own proof lane; golden
studio discipline (dry-first, guarded live, PASS/BLOCKED, operator eyes) applies to
every new body. Detail lives in the operator's project archives, not this skill — this
snapshot exists so no session re-derives the cast from scratch. The full layer map —
cast, routing, scenario chain — now lives in the `gopod-layer` skill.

## 7. The desk ledger

Two short lists, dated, meant to be **updated as things bank** — this is the one
section of this skill that changes over time.

**BANKED** (as of 2026-07-20):
- MIT license locked — `LICENSE` at repo root, `Copyright (c) 2026 CrushN8r`. See
  `gopod_notes/older_notes/LICENSE_LOCK_MIT_001.md`.
- Third-party attribution + `THIRD_PARTY_LICENSES.md`. See
  `gopod_notes/older_notes/LICENSE_ATTRIBUTION_FIX_001.md`.
- Studio skill suite (9 skills including the `studio` index). See
  `gopod_notes/older_notes/STUDIO_SKILLS_BANKED_001.md`.
- Bingo capture song golden, commit `298e388`. See
  `gopod_notes/older_notes/BINGO_VIDEO_SONG_GOLDEN_001.md`.
- Bingo sheet-music file capture + emotion-beat advisory fix, commit `f85c3a0`.
- Bingo targeted handback settle at `brobots_ready_together`, live-confirmed, commit
  `064ecda`. See `gopod_notes/older_notes/BINGO_TARGETED_HANDBACK_SETTLE_001.md`.
- Bingo score rebuilt to the operator's new 45-step version (+ `brobots_ready_together`
  phrase-passthrough fix + grouped-divider fix), live-confirmed 2026-07-19, commit
  `71fab7b`. See `gopod_notes/older_notes/BINGO_SCORE_REBUILD_001.md`.
- Bingo `wake_both` golden handler (clean stop + recovery message instead of a raw
  `TimeoutError` traceback) + playhead range-slice playback, dry-only, commit `5ff13be`.
  See `gopod_notes/older_notes/BINGO_WAKE_BOTH_PAIRING_GOLDEN_HANDLER_001.md`,
  `gopod_notes/older_notes/BINGO_PLAYHEAD_RANGE_PLAYBACK_001.md`.
- Playhead range-slice mechanism lifted into the shared control-song runner, dry-only,
  commit `f7485f8`. See `gopod_notes/older_notes/CONTROL_SONG_PLAYHEAD_LIFT_EXECUTED_001.md`.
- **Interview KG search sequence — a cool control-feature reveal.** Real
  `anim_knowledgegraph_*` clips (traced directly from Wire-Pod's own `kgsim.go`) scripted
  into the interview: Brobot 1 visibly "searches" (`searching` loop) → transitions
  (`searchingGetout`) → "answers" (`answering` loop) right before delivering an
  LLM-generated line — even though the interview's JSON is fully pre-generated before
  playback ever starts, so nothing is actually being searched live. The reveal doubles as
  the pitch: audience sees what looks like real-time thinking, then learns it's a fully
  offline, pre-generated performance — the "how'd they do that" moment reinforces the
  local-first story instead of contradicting it. Dry-verified, not yet live-fired on
  hardware, dry-only for now. See `gopod_notes/older_notes/INTERVIEW_KG_SEARCH_SEQUENCE_BUILT_001.md`.
- **Golden song scaffold + layer naming distinction, 2026-08-12.** Whole-shelf survey (active
  + `zzz_archives`) produced `tech/alias_play_studio/SONG_SCAFFOLD.md`, a starting reference
  for what a golden song looks like (status markers, engine choice, reporter-gap convention,
  the golden toolbox). Same pass: the brobots-wire-pod-layer-vs-GOPOD-layer naming distinction
  (Brobot 1/Brobot 2 here, Doc/Pip there) confirmed and swept repo-wide, with one real
  exception (`102_brobots_cross_persona`'s GOPOD-layer-test-footage framing) surfaced and
  handled correctly. `102_brobots_bingo_game` (Chocolate Bingo) reshaped onto proven golden
  patterns (connect-once/hold-it reactor fix, editable ball-call text, run 1/run 2 pacing
  choice, single end-of-game reaction). `102_brobots_cross_persona` archived as redundant once
  the real, live `103_gopod_is_that_you` proved the same bit for real. Full record:
  `gopod_notes/LAYER_NAMING_SWEEP_AND_102_ARCHIVED_001.md`.

**PENDING**, in standing order (reconciled 2026-07-20 against §8's launch checklist —
see that section's own reconciliation note; full sync behind this reconciliation:
`gopod_notes/older_notes/BINGO_GOVERLORD_BRAIN_SYNC_001.md`):
1. Bingo — runtime effectively done. **CLOSED**: the `stuck_animation` false
   positive → resolved as an honest advisory (`animation_wait_advisory`),
   evidence-based per `gopod_notes/older_notes/BINGO_STUCK_ANIMATION_FALSE_POSITIVE_RESOLUTION_001.md`,
   not a guess. **CLOSED**: the post-`brobots_ready_together` handback stall →
   targeted settle knob (`BROBOTS_READY_TOGETHER_HANDBACK_SETTLE_SECONDS`),
   live-confirmed, banked at commit `064ecda`. **CLOSED**: the sheet-music view,
   committed at `f85c3a0`. Remaining: the operator's own broader timing-tighten
   edits — no `pause_seconds`/buffer value has actually changed on disk yet as of
   2026-07-20, per `gopod_notes/older_notes/BINGO_GOVERLORD_BRAIN_SYNC_001.md` §1 (rehearsal
   tooling for this pass — the `wake_both` golden handler, playhead range-slice
   playback, and its lift to the shared control-song runner — is banked, but is not
   itself the tuning) — + reporter voiceover WAVs (still open, unchanged;
   post-production lane, the bait video's proven recipe, not a runtime change — see
   `gopod_notes/older_notes/NEXT_VIDEO_FOCUS_001.md`).
2. Launch conditions checklist (§8) — footage shooting (checklist step 1) begins only
   once 1 above is done. Absorbs what was tracked here as a separate
   "golden-guts sweep": per `gopod_notes/older_notes/BINGO_GOLDEN_ADVANCES_CARRY_FORWARD_001.md`'s
   own carry-forward assessment, one of five golden advances already carries free
   (the `[HH:MM:SS.mmm +X.XXXs]` timing prefix — a one-line `console_timestamps`
   opt-in per song, no lifting needed); the other four are Bingo-runner-local and
   need real lifting into shared code — that lifting IS what checklist step 3 (The
   Crystal Cut) actually is, the decouple proposal's own Stage D (shared song
   dispatch registry, `gopod_notes/older_notes/INTERVIEW_ENGINE_DECOUPLE_PROPOSAL_001.md`) —
   one effort, not two.
3. Community lineage piece — per
   `gopod_notes/older_notes/COMMUNITY_LINEAGE_SWEEP_001.md`'s own open findings (no
   root `LICENSE` was the blocking gap; now closed — the lineage piece itself is still
   the operator's own call on placement: README section vs standalone `tech/` doc).
4. Repo glamour polish — see §8 step 4 (The Glamour Skin) for the concrete checklist.
5. Launch — see §8 step 5 (Ignition).
6. **Reporter-wrapper cohesive shape** — pinned 2026-08-02. Status: the underlying
   pause/`gap_label`/`section` mechanism is proven (101/102/103's runner already
   supports it); the naming and placement across songs is not cohesive (101's six gaps
   don't sit at a literal start/end, 102 has only a mid gap, 00/01 carry no
   `reporter_gap_*` at all yet). Concrete plan already drafted, not yet actioned: adopt
   103's `intro`/`outro` naming as the floor for every song, add numbered interior gaps
   (`mid_1`, `mid_2`...) for multi-round songs like 101, and extend
   `run_robot_control_song_001.py`'s pause branch to read `gap_label`/`section` the way
   `run_golden_song_001.py` already does. Needs the operator's explicit go-ahead before
   any step gets touched — this reorders/relabels live song material. Full detail:
   `gopod_notes/older_notes/REPORTER_WRAPPER_RULE_001.md`.
7. **Pinned 2026-08-12: `life/03_EDUCATION.md` and `life/04_TEACHER_INSIGHT.md` naming
   cleanup.** Both are narrative pitch essays with Doc/Pip woven sentence-by-sentence into
   wire-pod-layer fact and GOPOD-layer aspiration together, not separated by section — the
   layer-naming sweep above deliberately left both untouched rather than chop up continuous
   creative copy. Needs the operator's own call: rename only the plain mechanical-fact lines,
   or leave the whole narrative voice alone (as was correctly done for the `DOCS_TAKE_LESSON`
   series). Not started.

**Parked** (ideas held, not built):
- Anki/DDL pitch — sequence is buzz first, leverage later, riding their relaunch wave
  as free traffic instead of approaching them first.
- WAWN reporter skill — formalizes the already-practiced WAWN-audit pattern as
  `playhead`'s Point-0 evidence engine. Precedent:
  `gopod_notes/older_notes/BINGO_WAWN_AUDIT_001.md`.
- WAWN NEWS broadcast segment — Brobots 3 & 4, the vehicle for niche-buzz updates (the
  evergreen doctrine's own exception, §5); full story gated behind the CRUSHN8R CREW'd
  Newsletter.
- Bodiless-persona visuals — phase 1 chat-bubble VFX, phase 2 AI-character face reveal;
  the reveal itself shot as a WAWN NEWS episode.
- "Why this song exists" purpose blocks for each song's `story.md` — landing survey
  already done in `gopod_notes/older_notes/REPO_FRONT_DOOR_SWEEP_001.md`; nothing written into any
  song yet.
- Lexicon glossary (the GOPOD shared vocabulary — WAWN, golden, scenario tags,
  doctrine words) — formalize at GOPOD-layer time, not before.
- PORTABLE_CONTEXT cards — paste-ready GOPOD context blurbs for outside AIs (short +
  full + hardware variant) — bank in `gopod_notes` when convenient.
- **pha0b (PLAYHEAD cockpit) aim** — settled direction, grows incrementally, no big
  build: pha0b is being grown into a universal starter-kit cockpit that understands ANY
  song's own exchange structure, within the robots' capabilities — not a bingo-only
  tool. Bingo is the first song it learns on; others can fork it for their own
  compositions. Four song-agnostic structural takeaways (from
  `gopod_notes/older_notes/BINGO_EXCHANGE_MAPPING_SURVEY_001.md` — that is the real filename on disk, not
  `BINGO_AS_EXCHANGES_MAPPING_SURVEY` as sometimes referenced verbally — plus the
  interview Section-Card model as reference): (1) a song is sections → exchanges →
  turns, a universal navigation tree; (2) every unit carries a one-line arc point ("why
  this exists"); (3) each song declares its OWN exchange-type vocabulary (interview:
  Opener/Q&A/C&A/Closer; bingo: Ball Call, Attitude Volley, Interrupt, Reaction Beat) —
  the tool reads what a song declares, hardcodes nothing; (4) stage mechanics (rattle,
  arm cue, gap, wake) are first-class but SEPARATE from exchanges. Growth rule: pha0b
  matures one lesson at a time WHILE serving real bingo-video polish — never scaffolding
  ahead of need; the section-vs-exchange-boundary question (bingo's boundaries don't
  always equal exchange boundaries) is deliberately left to emerge from real polish, not
  decided up front. Write-back to a score file remains the flagged one-pen crossing:
  explicit operator go, one pen, survey-then-commit.
- **Road-safety physics content pillar** — banked 2026-08-20, beyond-1k / vision-layer,
  not a get-to-1k task. Core hook: "the law vs. the law of nature" — the light says GO,
  the oncoming 18-wheeler's momentum doesn't read traffic lights. Same confusion→clarity
  fractal AI Wordplay already runs (green = safe, physics doesn't care about
  right-of-way), pointed at pedestrian/passenger/driver/road safety, delivered in
  GOPOD's own song/interview style — Police & EMS interviews, orientations, "day in the
  life" pieces, all possible future content, none scripted. Why future, not now: real
  accuracy stakes once actual emergency services are the subject, and a partnership
  realistically follows proof-of-audience, not precedes it. Candidate doors —
  **Windsor Police Service** and **Essex-Windsor EMS** — added to
  [OUTREACH.md](../../../life/02a_OUTREACH.md)'s Community lane targets, no arrangement
  claimed. Nothing scripted, no episodes, no pitch drafted.

## 8. Launch conditions

Per operator direction. Consequence of the evergreen doctrine (§5): this layer launches
once and then stands unattended for years — so it gets finished properly, once. Rockets
launch when weather allows optimum conditions, not on a fixed date. The checklist, in
order:

1. **Footage banked** — shoot every planned capture on the proven engine exactly as it
   runs today. The camera records the show, never the code; shooting before any
   refactor protects the footage completely.
2. **Dead weight jettisoned** — the dead binaries (`goverlord/runtime/songs/102_brobots_bingo_game/bin/*.pre-v{1,2,3}.bak`
   removed pre-session, `backups/` duplicate removed 2026-07-30) and the six zero-caller
   `gomads/` subdirectories (done 2026-07-30), per
   `gopod_notes/older_notes/INTERVIEW_ENGINE_DECOUPLE_PROPOSAL_001.md`'s own pruning list. May
   overlap with step 1 — pruning never touches the live engine.
3. **The Crystal Cut** — Perfect Crystal Chapter 2 (the interview/song engine decouple
   proposal, `gopod_notes/older_notes/INTERVIEW_ENGINE_DECOUPLE_PROPOSAL_001.md`), Stages B through
   E, dry-verified per stage, operator go/no-go per stage. Runs ONLY after footage is
   banked — the engine has no deadline once the captures are in the can.
4. **The Glamour Skin** — README synced to the four-song truth, the Claude Design look
   applied, credits/lineage proper, "Why this song exists" blocks on every score.
   Studio, songbook, and stage cohesive on one repo page.
5. **Ignition** — the net video posts to x.com; every click lands on a repo ready to be
   stared at.

**Reconciliation note** (recorded explicitly — this prevents double work): the
previously pending "golden-guts sweep" is ABSORBED by step 3 above — the decouple
proposal's Stage D (shared song dispatch registry) is that sweep's final form. One
effort, not two. No session should run a separate golden-guts sweep alongside the
decouple. The pre-shoot order still stands upstream of this checklist: bingo sheet
music → tighten (incl. the `stuck_animation` false-positive verdict) → then step 1
shooting begins.

## 9. Rules of the road

The doctrine this desk runs on:
- **Brand posture (the interview stance).** Calm is credible when the receipts are
  real. The repo stands in interview stance: license visible, credits proper, claims
  honest, boundaries without provocation. Launch-day rules: the bad-faith actor picks
  when it starts — posture is set before the moment, not during it; fundamentals beat
  tricks (golden discipline over gimmicks); disengaging is a trained skill, not
  weakness — not every comment is a fight, and there's no penalty for being too
  careful; expert in your thing, white belt in everything else.
- **Follow the operator's lead.** This file is a map, not a driver — it orients, the
  operator steers. A live correction from the operator outranks anything written here.
- **Translate intent, don't contribute.** Turn the operator's calls into concrete next
  steps; don't invent campaign direction of your own.
- **Survey-then-commit.** Any actual commit touching campaign material follows the
  `survey-then-commit` skill's own discipline — survey, classify, draft, stop for
  go-ahead.
- **One fact, one home.** This file points at the reports and skills that hold the real
  detail rather than restating them — when this ledger and a cited report disagree, the
  report (or the operator's live word) wins; update this file, don't argue from it.
- **Summaries are lossy.** The operator can't read everything and works from summaries
  — every relay hop (operator ↔ Claude ↔ Claude Code ↔ reports) is where mutation
  starts. This skill and the dated `gopod_notes/` reports are the referee: when any
  summary, recap, or chat memory conflicts with them, re-read the source and follow the
  written truth until the operator's live word says otherwise. Live operator word
  remains the highest truth of all, as already stated above.

## Scope

- Orientation/reference only — no logic, no automation, no code of its own.
- Keep §7 (the desk ledger) current as work actually bumps between PENDING and BANKED;
  the other sections are stable campaign structure, not expected to churn often.
- For song-production work itself (reading a score, dry-verifying a change, hardware
  calibration, committing), see `studio`'s index — this skill is one level up, the
  campaign, not the songs.

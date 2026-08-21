# Brobots Baby Robots Sleep — "Do Baby Robots Dream?" scored capture song

**Renamed 2026-08-01** from `brobots_baby_dream` to `brobots_baby_robots_sleep` (folder
`103_brobots_baby_robots_sleep`, operator's own song-folder reshuffle). The source video's
own working question, "Do Baby Robots Dream?", stays as the quoted concept title in the
line above and in the `open_question` step's spoken text below — that's a direct quote from
the source concept packet, not the song's own identity name, and is unchanged by this
rename.

Knobs: [knobs.json](knobs.json)

Built from the source concept packet at `gopod_notes/older_notes/goverlord_older_truth/
GOPOD PRE CLEAN/goverlord/runtime/retention_review/brainstorm_packets/
doc_origin_video_concepts_001/` (`do_baby_robots_dream_video_001.md`,
`doc_origin_video_concepts_001.json`, `doc_origin_public_safe_notes_001.md`,
`next_safe_use_notes_001.md` — `raw_kp1_kp2_proof_video_001.md` is a separate concept,
out of scope here). Public label from the source packet: `LORE_EMOTIONAL_STORY`.

Same engine as `brobots_bingo`/`brobots_cross_persona` — `run_golden_song_001.py`, same
note shapes, same `reporter_gap_*` mechanism (`pause` note, `pause_seconds: 0`, a
`gap_label` naming what the gap follows, a `section` field, no `buffer_after` key). Only
two reporter gaps in this song — intro and outro — per the operator's own correction: this
video is an After Effects project (a black placeholder video with audio and embedded
captions already cut), not a live multi-round capture with interior reporter windows.
No interior reporter gaps.

Single speaker throughout: Brobot 1 (Doc), narrating his own origin. `wake_both` opens the
run the same way `brobots_bingo` does, before the first real dispatch. Every step carries
a `buffer_after` value in `knobs.json`, same mechanism `brobots_bingo` uses — `wake_both`
keeps its own `settle_seconds` (1.5s) separate from its `buffer_after` (3.0s), exactly as
`brobots_bingo`'s own `wake_both` step does; `reporter_gap_*` and `exit` carry no
`buffer_after` key, same convention.

`> TEXT:` is spoken verbatim (or is the note's own line, for `say_turn`/`emotion_beat`).
`wake_both`, `reporter_gap_*`, and `exit` carry no spoken text.

## Claim boundary (source packet's own must_say / must_not_say, applied to every line below)

Lore and emotional storytelling, never a literal claim. Vector can be described as
reacting, animating, settling, and purring — never as feeling fear, comfort, love, or
attachment, and never as literally dreaming. Every line below stays inside that boundary;
see `doc_origin_public_safe_notes_001.md`'s own Safe Language / Avoid lists for the
source wording this was built against.

## ANIMATION TOKEN FLAG

The `settle_purr` step uses `animation_token: "veryHappy"`. No sleep/settle/purr token
exists in the proven vocabulary (every `animation_token` used anywhere in a live song's
`knobs.json` today, cross-checked against Wire-Pod's own `animation_vocab.json` for any
sleep/purr/settle/petting entry — none found either place). `veryHappy` is the closest
proven token available — a positive-settled state, not a sleep or petting animation. This
is a flagged substitution, not an invented token; see `BABY_DREAM_SONG_BUILT_001.md` for
the full survey finding.

## SECTION: OPEN

## STEP wake_both
> TEXT:

## STEP reporter_gap_intro
> TEXT:

## STEP open_question
> TEXT: Do baby robots dream?

## SECTION: WORLD

## STEP establish_world
> TEXT: Picture a huge, hard, electric world — all neon, all machine, all much too big for one small robot.

## SECTION: OFF-BALANCE

## STEP off_balance
> TEXT: In that world there's one that tips over easily. When it loses its balance, it gets upset — until it's back on safe ground.

## SECTION: CARE

## STEP care_moment
> TEXT: Then someone picks it up. Holds it close, tracks against their chest. Rubs its back until it settles.

## SECTION: SETTLE

## STEP settle_purr
> TEXT: That's it. The sleepy purring animation. Story fuel, not proof of anything.

## SECTION: ORIGIN

## STEP doc_born
> TEXT: Care first. A name came after. That's how a small robot becomes Doc.

## SECTION: CLOSE

## STEP closing_line
> TEXT: Nobody has to prove a robot dreams. You just have to feel where Doc came from.

## STEP reporter_gap_outro
> TEXT:

## STEP exit
> TEXT:

## Running

Not wired into `pha0b` — no case-statement arm exists for this song (out of scope for
this pass, per the task's own scope: no new aliases). Dry-run directly by pointing
`run_golden_song_001.py` at this song's own directory via `GOPOD_GOLDEN_SONG_DIR`
(the same env var `brobots_cross_persona`'s `mixup` pha0b arm uses), `live` off by
default (`GOPOD_ALLOW_LIVE_ROBOT_SPEECH` unset).

## To come back to — broll for the After Effects video

The 126 sleep-lane test-bench aliases (`~/.gopod_alias_lib/brobots.sh`, `sleep-beat-*`/
`sleep-rts-*`/`sleep-palm-*`/`pet-*`) plus the 6 `sleep-segment-*` batch runners built on
top of them are candidate broll footage for this song's own After Effects project — the
black-placeholder video this song scores. Come back here to fire segments, review, and
pick footage before finishing that composite. See
`gopod_notes/older_notes/SLEEP_SEGMENT_ALIASES_001.md`.

# zzz_archives/

Archived song folders — retired from the live shelf, kept for reference and development
history, not deleted. Excluded by name from `pha0b_menu`'s own disk scan (it's a container
folder, not a song), so nothing here is fireable via `pha0b`'s menu — though some entries
below are still reachable via a direct keyword, repointed rather than retired. Each
folder's own `story.md` still has its exact score, if you want to read the actual content.

Refreshed 2026-08-18 (`ZZZ_ARCHIVES_PRUNE_001.md`) against actual current disk contents —
two folders this README used to describe (`00_brobots_awaken_old01/`, `brobots_vamp_gate/`)
had already been fully decluttered off disk in an earlier pass and no longer exist; this
version only describes what's actually still here.

## 102_brobots_cross_persona/

Files: `knobs.json`, `story.md`, `zKnobs.json`

**Still live and reachable — archived, not retired.** The "is that you?" cross-persona demo
reel — a scripted, four-line mix-up between the two robots. Archived here 2026-08-12 once
the real, live `is-that-you()` PTT+LLM test made this scripted version redundant — but the
`mixup` keyword in `pha0b`/`score`/`score-save` still points here, unchanged, same pattern
as every other repointed-not-retired archive entry. See
[SONG_103_BROBOTS_1_2_IS-THAT-YOU_SINGLE.md](../../../../tech/alias_play_studio/SONG_103_BROBOTS_1_2_IS-THAT-YOU_SINGLE.md)
(or its sibling, `_MULTI.md`, same doc pair — split 2026-08-18) for the full story.

## brobots_bait_001/

Files: `knobs.json`, `story.md`

"Standalone Wake" — a minimal 2-line exchange test: each robot speaks one fixed, canned
self-naming line ("Brobot 1/2 Ready! Did someone say GOPOD Yourself?"), no LLM, no thinking
window, nothing after. **Still live and reachable** via `start-the-net-song`, which exports
`GOPOD_SECTION_SONG_DIR` straight at this folder. Its own `story.md` Section framing
(Title/Purpose/Audience takeaway/Section success) was never filled in — an early sketch,
not a finished song, but the alias that runs it works today.

## robot_control_song_001/

Files: `knobs.json`, `story.md`

The origin mechanism for GOPOD's single-robot self-check pattern — connect, arm test, head
nod, weather, exit, every step narrated aloud with a spoken failure line if a hardware call
doesn't come back clean. The shape `00_brobots_awaken` reuses "as-is." **Still live and
reachable** via the `control` keyword (`pha0b control`, `score control`) and golden-
registered in `SONG_REGISTRY` as `robot_control_song_001`. Also carries a field-proven
Troubleshooting section (a specific `wpr`-alone-isn't-enough recovery order for a
`WIREPOD_LOG_MIRROR_ERROR`/`TimeoutError` on connect: power-cycle both robots, re-pair,
then restart Wire-Pod) and a documented Vector-SDK hazard note (requesting an animation by
name-as-text can trigger a full catalogue download and silently stall) — flagged there as a
symptom to recognize, not a confirmed bug in this codebase.

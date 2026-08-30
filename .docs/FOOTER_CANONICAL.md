# FOOTER_CANONICAL

Records the GOPOD YAHMM (You Are Here Mall Map) architecture as of the 2026-08-27
folder restructure. Superseded the old single-block, copy-pasted-into-42-files scheme
(that version's own history: `gopod_notes/older_notes/` — see
`YAHMM_RESTRUCTURE_SURVEY_PLAN_001.md` for the survey that found the old scheme's drift
and proposed this one).

## The new architecture

Three real, hand-maintained menu pages replace the one repo-wide block:

- **`web/README.md`**, **`tech/README.md`**, **`life/README.md`** — each is the real
  map for that folder, grouped under the same reader-lane labels the old map used
  ("Start here," "The songs," "For teachers," "Doc's Take," etc.), filtered to only the
  docs that live in that folder. `tech/README.md` covers both `tech/`'s own docs and
  the 13 under `tech/alias_play_studio/`. These three files are the new source of
  truth for their own folders — not templated, not propagated, edited directly when a
  folder's doc set changes.

Three root-level docs (`README.md`, `MY_GOPOD_OPS_ASK.md`, `MY_NICHE_BUZZ_ASK.md`) carry
a slim footer instead of the full map: 3 pointer links to
the folder READMEs above, plus a **"Main docs"** group for root-level docs that don't
belong to any folder (each doc's own copy omits itself, same self-omission rule as
before):

```
## GOPOD YAHMM (You Are Here Mall Map)

Three folders, three maps — pick where you want to go:

- [web/README.md](web/README.md) — the content engine: pillars, wordplay, aha moments, newsletter
- [tech/README.md](tech/README.md) — the songs, the studio tooling, Wire-Pod integration
- [life/README.md](life/README.md) — the philosophy, teaching, and lessons learned

**Main docs**
- [README.md](README.md) — what GOPOD is and how it's built
- [tech/WIRED-POD.md](tech/WIRED-POD.md#open-wire) — the operator's own technical ask — what's built, where the line is, what kind of help this needs
- [MY_GOPOD_OPS_ASK.md](MY_GOPOD_OPS_ASK.md) — the operator's ops ask — social, sites, and content, a different role than the technical one
- [MY_NICHE_BUZZ_ASK.md](MY_NICHE_BUZZ_ASK.md) — help test the keyboard grabber, no robot required
- [TRAJECTORY.md](TRAJECTORY.md) — the planned arc, Point A to the pinnacle, honestly labeled built vs. aim
- [UNFAIR_ADVANTAGES.md](UNFAIR_ADVANTAGES.md) — the case for why this is worth your time — what GOPOD has that most projects don't
```

Every other doc that used to carry the full map (38 files across `web/`, `tech/`,
`tech/alias_play_studio/`, `life/`) now carries a two-line pointer instead, generated
the same way for every leaf doc — pointing at its own folder's README and the root:

```
## GOPOD YAHMM (You Are Here Mall Map)

Part of GOPOD — see [<folder>/README.md](<relative path>) for everything else in this
folder, or [the root map](<relative path>) for the rest of GOPOD.
```

`README.md` alone keeps a short epilogue (the Doc/Pip closer) after its own footer —
that's root-specific, not part of this pattern.

## Why this replaces the old scheme

The old single 42-row block was manually copy-pasted into every reader-facing doc and
had already drifted (a TRAJECTORY.md link landed in 3 of 4 root docs but not the 4th,
and never made it back into this file). A 42-file manual-propagation scheme has no
mechanism to catch that kind of drift — it just accumulates. The new scheme has exactly
3 real hand-maintained lists (the folder READMEs) instead of 42 copies of one list;
adding, removing, or renaming a doc means editing ONE folder README, not searching for
every stale copy.

## Maintenance

- A new doc added to `web/`, `tech/`, or `life/` gets one row in that folder's own
  README, plus a slim leaf-pointer footer copied from the pattern above (adjusted for
  its own depth — `tech/alias_play_studio/*.md` is one level deeper than `tech/*.md`
  and `web/`/`life/`'s own files, so its pointer paths get an extra `../`).
- A new root-level doc (not in any folder) gets a row in the "Main docs" group inside
  all 4 root/ASK docs' footers.
- Run `.claude/skills/gopolisher/gopod_consistency_check_001.py` after any of the
  above to catch a broken link before it ships.

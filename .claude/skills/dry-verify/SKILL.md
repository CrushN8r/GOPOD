---
name: dry-verify
description: Use when a GOPOD interview/pre-show code change (a new guard, an instrumentation pass, a text-cleanup fix, a display/data-plumbing change) needs to be proven correct before a live robot run, without touching real hardware, audio, or the network. Compile-check, import the module fresh (no side effects), reproduce the actual defect/behavior with a monkeypatched logic test, then run the existing test suite. Complements hardware-calibrate (which is for changes only a live robot's own motion can confirm) rather than overlapping it.
---

# Dry verify

Most GOPOD interview-runner changes are provable without a robot in the loop - a bug in
generation logic, a text-cleanup mapping, a new detection guard, or where a value gets
read from are all facts a standalone Python check can confirm. Reserve an actual live show
run for what genuinely needs one, and only after the operator says so.

## The four-step ritual (interview5 precedent: run this after timing instrumentation, the
self-repeat guard, say-cleaning alignment, and the display-lane fix - four real uses)

1. **Compile-check every touched file**: `python3 -m py_compile <file>`. Catches syntax
   errors before anything else does.
2. **Fresh module import, no side effects**: load the runner via the same
   `importlib.util.spec_from_file_location` pattern the codebase's own tools already use
   (see `_load_runner_module()` in `run_robot_control_song_001.py`/
   `run_interview_movement_rehearsal_001.py` for the precedent) and exec it. Module-level
   code runs; `main()` does not. Confirms imports resolve and no new top-level code throws.
3. **Reproduce the actual behavior with a monkeypatched logic test** - not a synthetic
   toy case, the real one. interview5 examples: fed the self-repeat guard the *exact*
   observed line-3-echoes-line-2 text and confirmed it resolved within the retry limit;
   fed `apply_universal_character_cleanup` every character in the Go-truth mapping and
   checked exact output; built a fake `generated` dict with curly quotes/emoji/markdown in
   `raw_llm_response` and confirmed `display_text` now equals the cleaned `speech_text`,
   not the raw value. Monkeypatch the one function that would otherwise hit a network/LLM
   call (`llm_colour_line`, `generated_turn`, etc.) - never mock the function under test
   itself.
4. **Run the existing test suite** (`goverlord/runtime/songs/02_brobots_interview_run/zmisc/*.py`, currently
   `test_section1_echo_detection_001.py`) - if a change touches a guard or generation path
   that suite covers, it must still pass. If the suite itself is broken in a way unrelated
   to your change (interview5 found two: a stale mock signature, a test fixture missing a
   field - both confirmed pre-existing by diffing against unmodified `HEAD` before
   touching them), fix the test too, since a guard test that can't run proves nothing - but
   confirm via `git stash`/`git show HEAD:<path>` that the breakage predates your session
   before touching a test file, and say so in the report either way.

## What this does NOT replace

- The task's own explicit stop conditions (e.g. "STOP before any live robot run") - dry
  verification is what earns the right to *ask* for the live-run go-ahead, never a
  substitute for asking.
- `hardware-calibrate`'s loop, for anything a robot's physical motion is the only proof of
  (motor speed floors, hold-duration tuning, amplitude). If in doubt which applies: can
  this be confirmed by reading code and running it against fakes, or does a human need to
  watch a robot's arm move? The former is this skill; the latter is `hardware-calibrate`.
- Live audio/TTS proof - Kokoro subprocess spawns, real Ollama LLM calls, and actual robot
  speech are not "dry" even with the Wire-Pod live gate off (CLAUDE.md's own gotcha:
  `say_text` and pre-show narration are decoupled from the live-robot gate). If a check
  would spawn real audio or hit the real LLM endpoint, it's not dry-verification - name
  that plainly rather than calling it dry.

## Scope

- Read-only against the actual interview song content (`story.md`/`knobs.json`) - dry
  verification proves the *code* is correct, never the *score*; don't touch either while
  doing this.
- No live robot run, ever, under this skill's own authority - that's always a separate,
  explicit ask.

# The searching sequence (GOLDEN, locked)
- **about:** What a Brobot's robot does while his answer is being made: the searching clip, one at a time, then one get-out. The one source - the code reads these values from here and nowhere else (gopod_thinking_line.go, gopodSearchingGolden). Locked by TestSearchingGoldenLock: a change here fails the deploy until the operator says so and runs the deploy with GOPOD_RELOCK=1.
- **locked:** 2026-11-03 (operator: "searching animation is breaking again. please isolate into a robust golden note and fucking lock it and refer only to it")
---
## The clips
1. searching: anim_knowledgegraph_searching_01
2. getout: anim_knowledgegraph_searching_getout_01

## The pace
1. min_seconds: 1
2. floor_seconds: 1

## The rules (the code, gopodThinkingProcess)
1. Before it: his moves (arm, nod) and "Got it!".
2. One searching clip at a time, back to back: the next is sent the moment the robot reports the last one done (2026-11-03, operator: "Back-to-back"; live 2026-11-02 the clip plays ~1.4 s, and the old 3.9 s re-send left ~2.5 s standing still). A clip that comes back sooner than floor_seconds (the robot not playing it) waits out the floor first, so he's never hit with commands.
3. It runs until his answer is ready, and for at least min_seconds.
4. A manager has the floor: one get-out, then he waits quietly; searching again once the manager is done.
5. Then one get-out. Nothing else sends the searching clip.

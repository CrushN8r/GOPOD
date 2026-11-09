# The numpad mapper
- **words:** numpad, numpad mapper, numpad map, keys, KP0, KP1, numlock, push to talk
- **ask:** Which key talks to whom?
- **chain:** Goverlord - the numpad assigner: each key, both NumLock layers, to a persona and a body
- **see also:** ../personas/ (each persona file), brobot-bubbles, chat-page-words
- **about:** 2026-11-04 (operator: "I want a system updated numpad-mapper topic-cluster * pull golden guts from GOPOD numpad mapping * like KP lines 0-9,* | [numlock on/off toggle], voices"; "build numpad-mapper.md as the one source"): the numpad map, moved here from ~/.gopod_alias_lib/numpad_persona_map_001.json. Two key layouts, one per keyboard (operator: "I have 2 keyboards * my PC keyboard which I use now * the connected USB numpad * can you add a keyboard-numpad mapping selector for the 2 numpad configs?"; "Two key layouts"): usb - the USB numpad, as it was; pc - the PC keyboard's numpad, a copy to start, yours to change. system-numpad-layout usb|pc picks one (numpad_ptt.layout in 05_voices_and_robots.json), then F5 the chat page. Each key is a section: persona, body (its robot or voice-only), status (live or reserved), and suit with NumLock off (reporter: the persona as its reporter; robot: the Brobot at dial zero). Edit a key here, then restart.
- **voices:** each key's voice is set in the chat page's Numpad voices panel, which saves it in 05_voices_and_robots.json (reporter_voices.hosts for the personas; robots_off.voices for Doc and Pip with the robots off) - it stays there so the panel keeps working.
- **the keys today, NumLock on:** KP0 Butler (Keys) - KP1 Doc - KP2 Pip - KP3 Git Repo - KP4 Cache PYC - KP5 Buzzer - KP6 Switch - KP7 CHALK - KP8 PLAYHEAD - KP9 CAMEO - * all the Brobots - / skip - 000 exit.
- **NumLock off:** KP7 CHALK and KP8 PLAYHEAD as their reporters (live); KP1 Brobot 1 and KP2 Brobot 2 at dial zero (reserved); KP3 GEMMA, the language model (live); KP4-KP6 and KP9 free (reserved).
---

## USB numpad - The map (numpad_maps.usb)
1. _note: Golden numpad/NumLock persona mapping - single source of truth: gopod_probes/topic_clusters/gopod-system/numpad-mapper.md (was ~/.gopod_alias_lib/numpad_persona_map_001.json until 2026-11-04). See gopod_notes/NUMPAD_NUMLOCK_GOLDEN_MAPPING_001.md for the design behind the NumLock ON/OFF doubling scheme.

## USB numpad - The fixed lanes - not part of the KP0-9 grid (numpad_maps.usb.fixed_lanes)
1. _note: Not part of the KP0-9 persona grid - always-on, independent of NumLock state. Two separate lanes on the same physical numpad, per operator: '000' and 'KP0' are NOT the same lane.

## USB numpad - Lane 000 - Exit (numpad_maps.usb.fixed_lanes.000)
1. action: Exit
2. gesture: 3x tap in quick succession
3. status: live

## USB numpad - Lane KP0 - the Butler (press/hold) - see numlock_on (numpad_maps.usb.fixed_lanes.KP0)
1. action: the Butler (press/hold) - see numlock_on.KP0; 3 quick taps still exit
2. gesture: press/hold
3. status: live
4. _was: Guest-mic chat participant (2nd operator / invited guest), decided_not_built

## USB numpad - Lane * - All Brobots - push-to-talk, in the order (numpad_maps.usb.fixed_lanes.*)
1. action: All Brobots - push-to-talk, in the order the All menu picks
2. gesture: press/hold
3. status: live

## USB numpad - Lane / - Skip (what * did until 2026-10-30) (numpad_maps.usb.fixed_lanes./)
1. action: Skip (what * did until 2026-10-30)
2. gesture: tap
3. status: live

## USB numpad - KP0, NumLock on - Butler (numpad_maps.usb.numlock_on.KP0)
1. persona: Butler
2. body: voice-only
3. status: live
4. _note: 2026-10-30 (operator: "KP0 hold -> Butlix (3 taps still exit)"; renamed: "Yes everywhere: KEYNOTE VALE (nickname "Keys")"): the Butler - persona ID BUTLER (his job name); his name and nickname are host_personas.BUTLER. Press and hold; he speaks with the butler's voice (managers.personas.BUTLER.voice_of)

## USB numpad - KP1, NumLock on - Doc (numpad_maps.usb.numlock_on.KP1)
1. persona: Doc
2. body: Vector 1
3. status: live

## USB numpad - KP2, NumLock on - Pip (numpad_maps.usb.numlock_on.KP2)
1. persona: Pip
2. body: Vector 2
3. status: live

## USB numpad - KP3, NumLock on - Git Repo (numpad_maps.usb.numlock_on.KP3)
1. persona: Git Repo
2. body: Cozmo
3. status: live
4. _note: 2026-10-26: live - Git Repo, a voice-assigned chat persona (managers.personas, reporter_voices.hosts); operator: "if a KPN is voice assigned, make it chat ready"

## USB numpad - KP4, NumLock on - Cache PYC (numpad_maps.usb.numlock_on.KP4)
1. persona: Cache PYC
2. body: Cozmo
3. status: live
4. _note: 2026-10-26: live - Cache PYC, a voice-assigned chat persona (managers.personas, reporter_voices.hosts); operator: "if a KPN is voice assigned, make it chat ready"

## USB numpad - KP5, NumLock on - Intel 404 PNFH (numpad_maps.usb.numlock_on.KP5)
1. persona: Intel 404 PNFH
2. body: Moorebot Scout
3. status: live
4. _note: 2026-10-26: live - Intel 404 PNFH, a voice-assigned chat persona (managers.personas, reporter_voices.hosts); operator: "if a KPN is voice assigned, make it chat ready"

## USB numpad - KP6, NumLock on - Intel 404 PNFT (numpad_maps.usb.numlock_on.KP6)
1. persona: Intel 404 PNFT
2. body: Moorebot Scout
3. status: live
4. _note: 2026-10-26: live - Intel 404 PNFT, a voice-assigned chat persona (managers.personas, reporter_voices.hosts); operator: "if a KPN is voice assigned, make it chat ready"

## USB numpad - KP7, NumLock on - CHALK (numpad_maps.usb.numlock_on.KP7)
1. persona: CHALK
2. body: voice-only
3. status: live
4. _note: 2026-10-18: live - CHALK BORED, Reporter 1's voice (managers.personas)

## USB numpad - KP8, NumLock on - PLAYHEAD (numpad_maps.usb.numlock_on.KP8)
1. persona: PLAYHEAD
2. body: voice-only
3. status: live
4. _note: 2026-10-18: live - PLAYHEAD ZUMINTU, Reporter 2's voice (managers.personas)

## USB numpad - KP9, NumLock on - CAMEO (numpad_maps.usb.numlock_on.KP9)
1. persona: CAMEO
2. body: voice-only
3. status: live
4. _note: 2026-10-26: live - Cameo, a voice-assigned chat persona (managers.personas, reporter_voices.hosts); operator: "if a KPN is voice assigned, make it chat ready"

## USB numpad - KP1, NumLock off - Brobot 1 (numpad_maps.usb.numlock_off.KP1)
1. persona: Brobot 1
2. body: Vector 1
3. suit: robot
4. status: reserved
5. _note: 2026-11-04 (operator: "reserve KP1/2 numlock OFF for brobots?"): Brobot 1 at the DocBS dial at zero - the shared Brobots persona, no Doc (personas/brobot-1.md); KP1 with NumLock on is Doc. Reserved, not built.

## USB numpad - KP2, NumLock off - Brobot 2 (numpad_maps.usb.numlock_off.KP2)
1. persona: Brobot 2
2. body: Vector 2
3. suit: robot
4. status: reserved
5. _note: 2026-11-04 (operator: "reserve KP1/2 numlock OFF for brobots?"): Brobot 2 at the PipTA dial at zero - the shared Brobots persona, no Pip (personas/brobot-2.md); KP2 with NumLock on is Pip. Reserved, not built.

## USB numpad - KP3, NumLock off - GEMMA (numpad_maps.usb.numlock_off.KP3)
1. persona: GEMMA
2. body: voice-only
3. status: live
4. _note: 2026-11-09 (operator: "Gemma should also be available in system numpad mapping"; "Do 1,3,4"): GEMMA ELLE-ELLE EHM, the language model as a chat persona, on the first free key

## USB numpad - KP4, NumLock off - - (numpad_maps.usb.numlock_off.KP4)
1. persona: -
2. body: -
3. status: reserved

## USB numpad - KP5, NumLock off - - (numpad_maps.usb.numlock_off.KP5)
1. persona: -
2. body: -
3. status: reserved

## USB numpad - KP6, NumLock off - - (numpad_maps.usb.numlock_off.KP6)
1. persona: -
2. body: -
3. status: reserved

## USB numpad - KP7, NumLock off - CHALK (numpad_maps.usb.numlock_off.KP7)
1. persona: CHALK
2. body: voice-only
3. suit: reporter
4. status: live
5. _note: 2026-10-30: CHALK in the outdoor reporter suit - same persona and voice, answers as its reporter (managers.personas.CHALK.reporter); operator: "CHALK/PLAYHEAD = same mapping, 2 suits (indoor/outdoor)"

## USB numpad - KP8, NumLock off - PLAYHEAD (numpad_maps.usb.numlock_off.KP8)
1. persona: PLAYHEAD
2. body: voice-only
3. suit: reporter
4. status: live
5. _note: 2026-10-30: PLAYHEAD in the outdoor reporter suit - same persona and voice, answers as its reporter (managers.personas.PLAYHEAD.reporter)

## USB numpad - KP9, NumLock off - - (numpad_maps.usb.numlock_off.KP9)
1. persona: -
2. body: -
3. status: reserved


## PC keyboard - The map (numpad_maps.pc)
1. _note: Golden numpad/NumLock persona mapping - the PC keyboard layout - a copy of the USB numpad layout to start (2026-11-04), yours to change; gopod_probes/topic_clusters/gopod-system/numpad-mapper.md (was ~/.gopod_alias_lib/numpad_persona_map_001.json until 2026-11-04). See gopod_notes/NUMPAD_NUMLOCK_GOLDEN_MAPPING_001.md for the design behind the NumLock ON/OFF doubling scheme.

## PC keyboard - The fixed lanes - not part of the KP0-9 grid (numpad_maps.pc.fixed_lanes)
1. _note: Not part of the KP0-9 persona grid - always-on, independent of NumLock state. Two separate lanes on the same physical numpad, per operator: '000' and 'KP0' are NOT the same lane.

## PC keyboard - Lane 000 - Exit (numpad_maps.pc.fixed_lanes.000)
1. action: Exit
2. gesture: 3x tap in quick succession
3. status: live

## PC keyboard - Lane KP0 - the Butler (press/hold) - see numlock_on (numpad_maps.pc.fixed_lanes.KP0)
1. action: the Butler (press/hold) - see numlock_on.KP0; 3 quick taps still exit
2. gesture: press/hold
3. status: live
4. _was: Guest-mic chat participant (2nd operator / invited guest), decided_not_built

## PC keyboard - Lane * - All Brobots - push-to-talk, in the order (numpad_maps.pc.fixed_lanes.*)
1. action: All Brobots - push-to-talk, in the order the All menu picks
2. gesture: press/hold
3. status: live

## PC keyboard - Lane / - Skip (what * did until 2026-10-30) (numpad_maps.pc.fixed_lanes./)
1. action: Skip (what * did until 2026-10-30)
2. gesture: tap
3. status: live

## PC keyboard - KP0, NumLock on - Butler (numpad_maps.pc.numlock_on.KP0)
1. persona: Butler
2. body: voice-only
3. status: live
4. _note: 2026-10-30 (operator: "KP0 hold -> Butlix (3 taps still exit)"; renamed: "Yes everywhere: KEYNOTE VALE (nickname "Keys")"): the Butler - persona ID BUTLER (his job name); his name and nickname are host_personas.BUTLER. Press and hold; he speaks with the butler's voice (managers.personas.BUTLER.voice_of)

## PC keyboard - KP1, NumLock on - Doc (numpad_maps.pc.numlock_on.KP1)
1. persona: Doc
2. body: Vector 1
3. status: live

## PC keyboard - KP2, NumLock on - Pip (numpad_maps.pc.numlock_on.KP2)
1. persona: Pip
2. body: Vector 2
3. status: live

## PC keyboard - KP3, NumLock on - Git Repo (numpad_maps.pc.numlock_on.KP3)
1. persona: Git Repo
2. body: Cozmo
3. status: live
4. _note: 2026-10-26: live - Git Repo, a voice-assigned chat persona (managers.personas, reporter_voices.hosts); operator: "if a KPN is voice assigned, make it chat ready"

## PC keyboard - KP4, NumLock on - Cache PYC (numpad_maps.pc.numlock_on.KP4)
1. persona: Cache PYC
2. body: Cozmo
3. status: live
4. _note: 2026-10-26: live - Cache PYC, a voice-assigned chat persona (managers.personas, reporter_voices.hosts); operator: "if a KPN is voice assigned, make it chat ready"

## PC keyboard - KP5, NumLock on - Intel 404 PNFH (numpad_maps.pc.numlock_on.KP5)
1. persona: Intel 404 PNFH
2. body: Moorebot Scout
3. status: live
4. _note: 2026-10-26: live - Intel 404 PNFH, a voice-assigned chat persona (managers.personas, reporter_voices.hosts); operator: "if a KPN is voice assigned, make it chat ready"

## PC keyboard - KP6, NumLock on - Intel 404 PNFT (numpad_maps.pc.numlock_on.KP6)
1. persona: Intel 404 PNFT
2. body: Moorebot Scout
3. status: live
4. _note: 2026-10-26: live - Intel 404 PNFT, a voice-assigned chat persona (managers.personas, reporter_voices.hosts); operator: "if a KPN is voice assigned, make it chat ready"

## PC keyboard - KP7, NumLock on - CHALK (numpad_maps.pc.numlock_on.KP7)
1. persona: CHALK
2. body: voice-only
3. status: live
4. _note: 2026-10-18: live - CHALK BORED, Reporter 1's voice (managers.personas)

## PC keyboard - KP8, NumLock on - PLAYHEAD (numpad_maps.pc.numlock_on.KP8)
1. persona: PLAYHEAD
2. body: voice-only
3. status: live
4. _note: 2026-10-18: live - PLAYHEAD ZUMINTU, Reporter 2's voice (managers.personas)

## PC keyboard - KP9, NumLock on - CAMEO (numpad_maps.pc.numlock_on.KP9)
1. persona: CAMEO
2. body: voice-only
3. status: live
4. _note: 2026-10-26: live - Cameo, a voice-assigned chat persona (managers.personas, reporter_voices.hosts); operator: "if a KPN is voice assigned, make it chat ready"

## PC keyboard - KP1, NumLock off - Brobot 1 (numpad_maps.pc.numlock_off.KP1)
1. persona: Brobot 1
2. body: Vector 1
3. suit: robot
4. status: reserved
5. _note: 2026-11-04 (operator: "reserve KP1/2 numlock OFF for brobots?"): Brobot 1 at the DocBS dial at zero - the shared Brobots persona, no Doc (personas/brobot-1.md); KP1 with NumLock on is Doc. Reserved, not built.

## PC keyboard - KP2, NumLock off - Brobot 2 (numpad_maps.pc.numlock_off.KP2)
1. persona: Brobot 2
2. body: Vector 2
3. suit: robot
4. status: reserved
5. _note: 2026-11-04 (operator: "reserve KP1/2 numlock OFF for brobots?"): Brobot 2 at the PipTA dial at zero - the shared Brobots persona, no Pip (personas/brobot-2.md); KP2 with NumLock on is Pip. Reserved, not built.

## PC keyboard - KP3, NumLock off - GEMMA (numpad_maps.pc.numlock_off.KP3)
1. persona: GEMMA
2. body: voice-only
3. status: live
4. _note: 2026-11-09 (operator: "Gemma should also be available in system numpad mapping"; "Do 1,3,4"): GEMMA ELLE-ELLE EHM, the language model as a chat persona, on the first free key

## PC keyboard - KP4, NumLock off - - (numpad_maps.pc.numlock_off.KP4)
1. persona: -
2. body: -
3. status: reserved

## PC keyboard - KP5, NumLock off - - (numpad_maps.pc.numlock_off.KP5)
1. persona: -
2. body: -
3. status: reserved

## PC keyboard - KP6, NumLock off - - (numpad_maps.pc.numlock_off.KP6)
1. persona: -
2. body: -
3. status: reserved

## PC keyboard - KP7, NumLock off - CHALK (numpad_maps.pc.numlock_off.KP7)
1. persona: CHALK
2. body: voice-only
3. suit: reporter
4. status: live
5. _note: 2026-10-30: CHALK in the outdoor reporter suit - same persona and voice, answers as its reporter (managers.personas.CHALK.reporter); operator: "CHALK/PLAYHEAD = same mapping, 2 suits (indoor/outdoor)"

## PC keyboard - KP8, NumLock off - PLAYHEAD (numpad_maps.pc.numlock_off.KP8)
1. persona: PLAYHEAD
2. body: voice-only
3. suit: reporter
4. status: live
5. _note: 2026-10-30: PLAYHEAD in the outdoor reporter suit - same persona and voice, answers as its reporter (managers.personas.PLAYHEAD.reporter)

## PC keyboard - KP9, NumLock off - - (numpad_maps.pc.numlock_off.KP9)
1. persona: -
2. body: -
3. status: reserved

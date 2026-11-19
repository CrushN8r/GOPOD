# Comm Check

**Status:** WIP (2026-11-04)

The comm check before the show: `system-comm-check` plays it. Goverlord hands it to his butler,
Keys, to get the party started; Keys introduces himself, then runs the round robin - he calls each
live key of the numpad map in: the Brobots first (Brobot 1 and 2, NumLock off - 2026-11-14,
operator: "I also asked for the brobots after Keys"), then every other live key in the map's order, then the Brobots together (2026-11-13: "keys": ["map"] - the map's
live keys, so a key made live or off in numpad-mapper.md joins or leaves the comm check itself),
and each introduces themselves: what they do, who they answer to and who works with them. Each line
goes to its keys in turn, as if the key were held and the words spoken - the model, the voices and
the robots, live - and the next waits for the answer to finish. It starts by waking both Brobots.
`{name}` is who the key reaches. 2026-11-17 (operator's review): Goverlord hands it to Keys -
Keys answers with his Got it and the crew brain bubble (comm-check.md, comm_check.crew) - then calls
Goverlord, introduces himself, and calls the crews in order: the System Crew (LLM), Security (Buzzer,
Switch), the Show Hosts (Doc, Pip), the Gomads (Chalkra, Zoomy, Chammy), their Interns (Repo, Mimi), the
Gominions (Bruh 1, Bruh 2), then the Brobots together. No reporter calls, no numpad status bubbles. The keys and who asks each step are in [knobs.json](knobs.json);
the waits are the settings' `comm_check` (quiet_seconds, line_max_seconds).

---

## SECTION: OPENING

## STEP brobots_wake_0101

---

## SECTION: ROUND ROBIN

## STEP keys_ask_0201
> TEXT: {name}! Get the party started: get everyone to introduce themselves to the chat - what they do, who they answer to and who works for who. Start with me, then yourself, our Hosts, Gomads and their Interns, Gominions

## STEP goverlord_ask_0202
> TEXT: Next up, {name}! Introduce yourself to the group chat.

## STEP keys_self_0203
> TEXT: Introduce yourself to the group chat.

## STEP crews_ask_0204
> TEXT: Next up, {name}! Introduce yourself to the group chat.

## STEP all_ask_0206
> TEXT: And now, {name}! You party animals! What's the weather like today?

# Comm Check

**Status:** WIP (2026-11-04)

The comm check before the show: `system-comm-check` plays it. Goverlord hands it to his butler,
Keys, to get the party started; Keys introduces himself, then runs the round robin - he calls each
persona in, top-down (the Gomads, the Brobots, the interns, the guards, then the Brobots together),
and each introduces themselves: what they do, who they answer to and who works with them. Each line
goes to its keys in turn, as if the key were held and the words spoken - the model, the voices and
the robots, live - and the next waits for the answer to finish. It starts by waking both Brobots.
`{name}` is who the key reaches. The keys and who asks each step are in [knobs.json](knobs.json);
the waits are the settings' `comm_check` (quiet_seconds, line_max_seconds).

---

## SECTION: OPENING

## STEP brobots_wake_0101

---

## SECTION: ROUND ROBIN

## STEP keys_ask_0201
> TEXT: Keys! Get the party started: introduce yourself to the chat - what you do, who you answer to and who works with you.

## STEP gomads_ask_0202
> TEXT: Next up, {name}! Introduce yourself to the chat - what you do, who you answer to and who works with you.

## STEP brobots_ask_0203
> TEXT: Next up, {name}! Introduce yourself to the chat - what you do, who you answer to and who works with you.

## STEP interns_ask_0204
> TEXT: Next up, {name}! Introduce yourself to the chat - what you do, who you answer to and who works with you.

## STEP guards_ask_0205
> TEXT: Next up, {name}! Introduce yourself to the chat - what you do, who you answer to and who works with you.

## STEP all_ask_0206
> TEXT: And now, {name}! You party animals! Introduce yourself to the chat, & what's the weather like today?

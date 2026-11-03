# The weather run (GOLDEN, locked)
- **words:** weather, weather run, what's the weather
- **ask:** What's the weather like in Tampa Florida right now?
- **about:** The golden weather run, Doc then Pip, as live 2026-11-02 20:15 (Tampa) - the one source for its order and every line the code says. The Brobots' own colour lines stay free. Locked by TestWeatherGoldenLock: a change here fails the deploy until the operator says so and deploys with GOPOD_RELOCK=1.
- **locked:** 2026-11-03 (operator: "Look at the tampa run. that is a near perfect example of a golden locked run * note the handoff sequencing and 2nd brobot following up * golden guts scaffolding = topic-cluster md"; "Lock order + code lines"; "Doc then Pip (Tampa)")
- **order:** Brobot A is the Brobot picked first, Brobot B his brother. A: "Got it!" - A's brain bubble, with PLAYHEAD: "Cache PYC will help you with that." then Cache PYC: "Weather info received from Git repo. Ready when you are Bro!" - A: "Thanks for your help, Cache PYC." - A's weather report - A's take - A's hand-over, "Brobot B, what say you on this, [funny nickname]?" - B: "Nice try, [funny nickname]!" - B's brain bubble, with Cache PYC's line (no PLAYHEAD: the intern is already there) - B's "Confirmed." - B's take - the Weather Kit link - B's "Hey, that reminds me..." and "Try asking: ..."
- **order_from:** 2026-11-03 (operator's flow: "PLAYHEAD in 1st brain bubble (PLAYHEAD take over required, for intern) * Cache PYC in both brain bubbles (PLAYHEAD take over not required, intern is there)"; Cache PYC's line "Said and shown")
---

## Step 3: PLAYHEAD hands it to Cache PYC (managers)
1. handoff_text: {intern} will help you with that.

## Step 4: Thanks to the intern (bubble_notices.value_points)
1. got_it_again: Thanks for your help, {helper}.

## Steps 5 and 9: Doc reports, Pip confirms (weather_lines)
1. report: Brobot Weather Report for {place}, {date}. Temperature is {temp} and {sky} as of {time}.
2. confirm: Confirmed. Temperature is {temp} at {time}.
3. intern_ready: Weather info received from Git repo. Ready when you are Bro!

## Steps 6 and 10: what the model is asked for, around the weather (chain_prompts)
1. weather_lead: First one short line about this weather. Then:
2. weather_lead_next: First one short line following up your brother's weather comment. Then:
3. weather_point_first: Write two short paragraphs: first one line about this weather; then, on a new line, party on your value point above, your way.
4. weather_point_next: Write two short paragraphs: first one line following up your brother's weather comment; then, on a new line, party on your value point above, your way.
5. weather_take: React to that weather report in one or two short lines, your way. No numbers, no dates, no names, no hand-over - the report and the hand-over are said for you.
6. weather_take_next: React to your brother's weather report and his take in one or two short lines, your way. No numbers, no dates, no names, no hand-over - the report is said for you.

## Step 7: the hand-over, Brobot 1 to Brobot 2 (host_personas.Brobot 1.handover)
1. line: {partner}, what say you on this, {title} {nickname}?

## The hand-over, Brobot 2 to Brobot 1 (host_personas.Brobot 2.handover)
1. line: {partner}, what say you on this, {title} {nickname}?

## Step 11: the Weather Kit link (live_weather)
1. follow_up_kit_line: Want to see every weather face Vector has? Check out the [GOPOD Weather Kit](/gopod_weather_kit.html).

## Doc's answer, steps 5-7, and Pip's, steps 9-12, in order (answer_shapes.shapes)
1. weather: weather_report | llm? | point? | llm? | reminder
2. weather_next: weather_confirm | llm? | point? | llm? | weather_kit | reminder

## How each step is known in an answer (answer_shapes.marks)
1. weather_report: Brobot Weather Report
2. weather_confirm: Confirmed. Temperature
3. weather_kit: Weather Kit
4. handover: what say you on this

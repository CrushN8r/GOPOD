# The weather run (GOLDEN, locked)
- **words:** weather, weather run, what's the weather
- **ask:** What's the weather like in Tampa Florida right now?
- **about:** The golden weather run, Doc then Pip, as live 2026-11-02 20:15 (Tampa) - the one source for its order and every line the code says. The Brobots' own colour lines stay free. Locked by TestWeatherGoldenLock: a change here fails the deploy until the operator says so and deploys with GOPOD_RELOCK=1.
- **locked:** 2026-11-03 (operator: "Look at the tampa run. that is a near perfect example of a golden locked run * note the handoff sequencing and 2nd brobot following up * golden guts scaffolding = topic-cluster md"; "Lock order + code lines"; "Doc then Pip (Tampa)")
- **order:** 1 the response chain, Doc then Pip - 2 Doc's brain bubble - 3 PLAYHEAD: "Cache PYC will help you with that." - 4 Doc: "Thanks for your help, Cache PYC." - 5 Doc's weather report - 6 Doc's colour - 7 Doc's hand-over, "what say you on this" - 8 Pip's brain bubble - 9 Pip's "Confirmed." - 10 Pip's colour - 11 the Weather Kit link - 12 Pip's "Hey, that reminds me..."
---

## Step 3: PLAYHEAD hands it to Cache PYC (managers)
1. handoff_text: {intern} will help you with that.

## Step 4: Thanks to the intern (bubble_notices.value_points)
1. got_it_again: Thanks for your help, {helper}.

## Steps 5 and 9: Doc reports, Pip confirms (weather_lines)
1. report: Brobot Weather Report for {place}, {date}. Temperature is {temp} and {sky} as of {time}.
2. confirm: Confirmed. Temperature is {temp} at {time}.

## Steps 6 and 10: what the model is asked for, around the weather (chain_prompts)
1. weather_lead: First one short line about this weather. Then:
2. weather_lead_next: First one short line following up your brother's weather comment. Then:
3. weather_point_first: Write two short paragraphs: first one line about this weather; then, on a new line, party on your value point above, your way.
4. weather_point_next: Write two short paragraphs: first one line following up your brother's weather comment; then, on a new line, party on your value point above, your way.

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

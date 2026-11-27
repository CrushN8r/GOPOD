# The weather run (GOLDEN, locked)
- **words:** weather, weather run, what's the weather
- **ask:** What's the weather like in Tampa Florida right now?
- **about:** The golden weather run, Doc then Pip, - the one source for its order and every line the code says. The Brobots' own colour lines stay free. Locked by TestWeatherGoldenLock: a change here fails the deploy until the operator says so and deploys with GOPOD_RELOCK=1.
- **locked:** golden - changed only on the owner's go.
---

## The run, as it plays - A is the Brobot asked first, B his brother; [model] is his own line, the rest the code says
1. Brobot A: Good [evening]!
2. Brobot A: Brobot Weather Report for [place], [date]. Temperature is [temp] and [sky] as of [time].
3. Brobot A: [model] his take on the weather.
4. Brobot A: Brobot B, what say you on this, [title] [nickname]?
5. Brobot B: Nice try, [nickname]!
6. Brobot B: Good [evening]!
7. Brobot B: Confirmed. Temperature is [temp] at [time].
8. Brobot B: [model] his follow-up on A's take.
9. Brobot B: Want to see every weather face Vector has? Check out the GOPOD Weather Kit.
10. Brobot B: Hey, that reminds me... we've still got [topics] to get to! To GOPOD Yourself, try asking: "[ask]"
11. Brain bubbles: A's - PLAYHEAD: "Cache PYC will help you with that.", then Cache PYC: "Weather info received from Git repo."; B's - Cache PYC's line only.

## Step 11: PLAYHEAD hands it to Cache PYC (managers)
1. handoff_text: {intern} will help you with that.

## Step 11: Thanks to the intern (bubble_notices.value_points)
1. got_it_again: Thanks for your help, {helper}.

## Steps 2 and 7: A reports, B confirms - each in his own units (weather_lines)
1. report: Brobot Weather Report for {place}, {date}. Temperature is {temp} and {sky} as of {time}.
2. confirm: Confirmed. Temperature is {temp} at {time}.
3. intern_ready: Weather info received from {from}. Ready when you are Bro!

## Steps 3 and 8: what the model is asked for, around the weather (chain_prompts)
1. weather_lead: First one short line about this weather. Then:
2. weather_lead_next: First one short line following up your brother's weather comment. Then:
3. weather_point_first: Write two short paragraphs: first one line about this weather; then, on a new line, party on your value point above, your way.
4. weather_point_next: Write two short paragraphs: first one line following up your brother's weather comment; then, on a new line, party on your value point above, your way.
5. weather_take: React to that weather report in one or two short lines, your way. No numbers, no dates, no names, no hand-over - the report and the hand-over are said for you.
6. weather_take_next: React to your brother's weather report and his take in one or two short lines, your way. No numbers, no dates, no names, no hand-over - the report is said for you.

## Step 4: the hand-over, Brobot 1 to Brobot 2 (host_personas.Brobot 1.handover)
1. line: {partner}, what say you on this, {title} {nickname}?

## Step 4: the hand-over, Brobot 2 to Brobot 1 (host_personas.Brobot 2.handover)
1. line: {partner}, what say you on this, {title} {nickname}?

## Step 2's sound: the weather sound plays as this lights up (weather_card)
1. sound_at: Brobot Weather Report

## Step 9: the Weather Kit link (live_weather)
1. follow_up_kit_line: Want to see every weather face Vector has? Check out the [GOPOD Weather Kit](/gopod_weather_kit.html).

## A's answer, steps 2-4, and B's, steps 7-10, in order (answer_shapes.shapes)
1. weather: weather_report | llm? | point? | llm? | reminder
2. weather_next: weather_confirm | llm? | point? | llm? | weather_kit | reminder

## How each step is known in an answer (answer_shapes.marks)
1. weather_report: Brobot Weather Report
2. weather_confirm: Confirmed. Temperature
3. weather_kit: Weather Kit
4. handover: what say you on this

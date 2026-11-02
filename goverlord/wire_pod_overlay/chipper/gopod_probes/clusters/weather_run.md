# The weather run
- **words:** weather, weather run, what's the weather
- **ask:** What's the weather like today?
- **about:** PLAYHEAD hands the weather to his intern Cache PYC; the first Brobot gives the weather report from his own Wire-Pod settings; the hand-over (what say you on this, with a nickname from nicknames.md); the second Brobot confirms it from his own settings; then thanks to the intern.

---

## PLAYHEAD hands it to Cache PYC (managers)
1. handoff_text: {intern} will help you with that.

## The first Brobot reports, the second confirms (weather_lines)
1. report: Brobot Weather Report for {place}, {date}. Temperature is {temp} and {sky} as of {time}.
2. confirm: Confirmed. Temperature is {temp} at {time}.

## The hand-over, Brobot 1 to Brobot 2 (host_personas.Brobot 1.handover)
1. line: {partner}, what say you on this, {title} {nickname}?

## The hand-over, Brobot 2 to Brobot 1 (host_personas.Brobot 2.handover)
1. line: {partner}, what say you on this, {title} {nickname}?

## Thanks to the intern (bubble_notices.value_points)
1. got_it_again: Thanks for your help, {helper}.

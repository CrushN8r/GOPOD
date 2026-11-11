# Bubble notices
- **words:** notice, brain bubble notice, got it, score line
- **ask:** What does a Brobot's notice bubble say?
- **chain:** Gominions - the notice inside a Brobot's bubble, built by the code
- **see also:** brobot-brain, game-words, weather_run (got_it_again)
- **about:** 2026-11-03 (step 6 of GOPOD_CLUSTERS_PLAN_002.md): from 03_chat_page.json bubble_notices. A value in backticks is that text exactly ("\n" a break). The numbers, switches, roles and nested groups stay in the settings for the format step.
---

## Notices: egg (bubble_notices.egg)
1. title: Easter Egg '{title}'
2. line: {summary}

## Notices: native intent (bubble_notices.native_intent)
1. title: Native Wire-Pod Intent Detected
2. line: {intent}

## Value points notice (bubble_notices.value_points)
1. title: {name} Brain
2. parts: plan | memory | score | points | summary | search | egg | combo | chapter
3. note: -{note}
4. notes: [{notes}]
5. notes_join: `|`
6. egg: Easter Egg '{title}': {summary}
7. combo: Combo: {topics}
8. chapter: Chapter cleared: {chapter}
9. recovery: {topics}
10. next: Next up: {topics}
11. summary: Summary: {topics}
12. score: `Topics {topics} of {topics_total}:	{topics_bar}\nValue Points {points} of {points_total}:	{points_bar}\nEaster Eggs {eggs} of {eggs_total}:	{eggs_bar}`
13. got_it: {thinking_text}
14. points: {topic} #{n} of {total}: {words}
15. weather: Weather: {line}
16. search: Web search: {search}
17. none: No value point - a quick take, then back on track:
18. helper_from: call_name | name | brobot
19. second_got_it: after_manager

## Value points notice: guide (bubble_notices.value_points.guide)
1. summary: {line} - give it as a short summary, your way.
2. search: {line} - the results show with your answer: use them, never read them out.
3. egg: {line} - react to the find.
4. combo: {line} - tie them together in one answer.
5. chapter: {line} - celebrate it.
6. none: {line} - one or two sentences, then bridge back to GOPOD.
7. recovery: Back on track with: {line} - pick one, don't reveal them all.
8. next: {line} - don't reveal these yet.

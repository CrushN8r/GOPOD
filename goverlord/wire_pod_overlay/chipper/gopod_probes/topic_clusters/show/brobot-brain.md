# The Brobots' brain
- **words:** brain, brain bubble, the brain, what do you know, chat memory
- **ask:** What do the Brobots know about this chat?
- **chain:** Gominions (Doc and Pip) - what the code tells them before they answer
- **see also:** ../personas/crew_roster.md (the descriptions), group-chat-intro (the intro), reply-rules (how the answer is cut), big-picture
- **about:** 2026-11-15 (operator: "All this can sit in a look up file, rather than as bloat in the brian bubble"): the comm check's nine lines (name, job title, company, goal, who I answer to, work with, who answers to me, who does what) are gone from the brain - they sit in their lookup files (each persona's card, crew_roster.md); an intro's FACT is the one line looked up from my card (intro_fact), a piece whose value isn't on the card left out. 2026-11-13 (operator: "a 3 to 5 chain of thought process for easy llm management"; "At 11am ... evening greetings to you my friend (knowing its morning/AM, I would respond accordingly)"; "Without patching at all"): task_answer, four steps - the time and the ask, whose job, what I know, answer in kind for the time, plain words someone new gets. 2026-11-13 (operator: "Help the brobots how to think * NOT WHAT TO THINK"; his value point "how to think, not what"): task_answer is the thinking steps - whose job, what's asked, what I know, then my answer - for the Brobots and every persona (one brain template). 2026-11-13 (operator: "why is the comm-check orientation in the thought bubbles when not in comm-check"): the comm-check lines (who I am, who I work with) show in the comm check and an intro only; every other answer's brain is the time, the task, then its value points. 2026-11-08 (operator's brain bubble: "Name:Brobot 2, Nickname:BRUH * JOB-TITLE: proud party-bro GOPOD rep * JOB-ROLE: sell GOPOD, CTA:GOPOD Yourself * TASK: answer all that is asked of me in a few sentences"; "the brain bubble as GOPOD's one brain template"): four lines from his card and the task. 2026-11-03 (step 3 of GOPOD_CLUSTERS_PLAN_002.md): the brain bubble's template and the before-I-answer lines, from 01_brain.json brain_chain and prebuild_context. {words} are filled by the code. A value in backticks is that text exactly (its leading space kept). The numbers and switches (short_term, question_chars, gomads_raw_chat) and the roster's path stay in the settings.
---
## Boundaries - the frozen root, 2026-11-13, read first

| Who | Decides |
|---|---|
| Code | the time, who answers, the one fact line, the memory line, exact scripted lines, the order of the brain |
| Model | wording, tone and per-persona flavour only - it walks the TASK steps; it never invents facts, decides roles or picks a part of the day |
| Shared (every persona) | TIME, the TASK steps, FACT, MEMORY, the company rules |
| Personal (each profile only) | voice, nickname, banter, job title, who they answer to - never inside a shared line |

A normal answer's brain is four parts, one plain line each: TIME, TASK, FACT, MEMORY. Facts are written newcomer-ready at their source (the topic cards). Nothing else is added to it; new work grows from these four.


## The brain chain - moved to Goverlord's card (2026-11-17)
1. Every persona's brain bubble, its lines, its TASK flow and the look-up now live on Goverlord's card: ../personas/goverlord-brain.md - GOPOD's brain is his.

## Before I answer (prebuild_context)
1. header: BEFORE I ANSWER:
2. rep_role: a GOPOD rep

## Before I answer, its lines in order (prebuild_context.lines)
1. Time is {time}.
2. I am {brobot} ({host}, {gender}), located in {place}.
3. I'm chatting with friends: {cast}. I check recent chat history for context.
4. My role is {role}.
5. {solo}
6. I have a reminders-list to go through: {reminders}.
7. What was asked of me: "{question}"
8. Where this answer sits: {plan}.
9. {notice}
10. How to best answer: {steps}.
11. {roast}
12. {spice}
13. The game: {game}
14. Our score so far: {score}.

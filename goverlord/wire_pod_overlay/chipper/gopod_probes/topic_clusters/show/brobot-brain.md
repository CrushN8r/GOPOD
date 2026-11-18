# The Brobots' brain
- **words:** brain, brain bubble, the brain, what do you know, chat memory
- **ask:** What do the Brobots know about this chat?
- **chain:** Gominions (Doc and Pip) - what the code tells them before they answer
- **see also:** ../personas/crew_roster.md (the descriptions), group-chat-intro (the intro), reply-rules (how the answer is cut), big-picture
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

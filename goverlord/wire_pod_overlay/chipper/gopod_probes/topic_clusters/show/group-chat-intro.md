# The group chat intro (GOLDEN, locked)
- **words:** intro, introduce yourself, tell me about yourself
- **ask:** Hey Keys, intro
- **about:** GOPOD's company persona orientation introduction group chat: GOPOD has called everyone in, and given them the chat details; each persona introduces themselves to the group. The one source for every intro - the code says the opening, the model a line or two of what they do and who they work with. An intro with the weather asked too ("Hey Doc, intro & weather") loads this file, then the weather run (weather_run.md): his intro first, then his weather. Locked by TestGroupChatIntroLock: a change here fails the deploy until the operator says so and deploys with GOPOD_RELOCK=1.
- **locked:** 2026-11-06 relocked (operator: "go with formal first names on company time"; "So fix all that" - PERSONA_SCHEMA_DRAFT_002.md; then "every intro should start similar to \"I'm [], my nickname is [] ...\""; "for nicknames, why the chosen nicknames too"); 2026-11-04 relocked (operator: "relationships aren't being talked about * remove who's here in the chat from every response, too repetitive"); 2026-11-03 (operator: "company-persona-orientation-introduction-group-chat * if GOPOD has given them chat details"; "Code + 1 model line"; "files stay separate. For intro-and-weather load intro topic cluster md, then weather topic cluster md"; "Intro, then weather")
- **runs:** a single intro (anyone, "Hey [name], intro") - the duo intro + weather (Brobot A, then Brobot B) - Brobot A: intro + weather - Brobot B: intro + weather follow-up
- **order:** a single intro: his opening (code) - his line (model: what he does and his working relationships). Brobot A, intro + weather: "Got it!" - his brain (PLAYHEAD, Cache PYC) - "Thanks for your help, Cache PYC." - his opening - his line - the weather report - his take - the hand-over. Brobot B: "Nice try, [funny nickname]!" - his brain (Cache PYC) - his opening - his line - "Confirmed." - his take - the Weather Kit - "Hey, that reminds me..."
---

## The intro, said to the group (intro_recipe)
1. words: intro | introduce yourself | tell me about yourself | tell me all about yourself
2. opening: I'm {name}{nickname_part}.
3. nickname_part: , my nickname is {nickname}
4. why_part: , because {why}
5. text: This is GOPOD's company orientation group chat - GOPOD gave everyone here the chat details, and each of us introduces ourselves to the group. Don't repeat the question. Your name, your nickname and its story are already said for you - never say or explain them again; start with your job. Then two or three sentences, in your own words, proud of your work, in plain words with no stage directions - never labels like "my job title" - of what you do, who you answer to and who works with you: {facts}. Only those. Don't list who's here.
6. weather_text: Then, on a new line, react to the weather report in one or two short lines, your way. No numbers, no dates, no names, no hand-over - the report and the hand-over are said for you.
7. weather_text_next: Then, on a new line, react to your brother's weather report in one or two short lines, your way. No numbers, no dates, no names, no hand-over - the report is said for you.
8. cast_line: ``
9. relations: Who I answer to| | Who works with me|

## Guided Task Flow (brain_chain.flow)
1. intro: a) Introduce yourself - who you answer to and who works with you. b) KEYNOTE opened the session.
2. intro_first: a) Introduce yourself - who you answer to and who works with you. b) You're the first to speak - open up the session.

## Who opens the session (brain_chain)
1. intro_opener: KEYNOTE

## A Brobot's intro + weather, in order (answer_chains.brobots_intro_weather)
1. parts: llm | intro_weather | weather_kit | reminder | redirect

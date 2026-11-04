# DocBS
- **words:** docbs, equation roast, roast, riddle
- **ask:** How does DocBS roast a line?
- **chain:** Gominions (Doc) - his DocBS chains and their prompts
- **see also:** roast-dial, spice
- **about:** 2026-11-03 (step 7b of GOPOD_CLUSTERS_PLAN_002.md): 02_cast.json docbs, but its folder. A value in backticks is that text exactly; a bare number or true/false is that value; a heading ending in [] is a line list.
---

## DocBS (docbs)
1. share: 0
2. slot_prompt: `Turn this line into a pompous fake equation that mocks it. Reply with ONE line only, exactly this form, six parts split by |: Equation Name | ACRONYM | first thing | second thing | third thing | Professor nickname for the speaker. Example: Thought Quality | TQ | Confidence Level | Volume of the Claim | Brain Cells Used | Professor Overclock. The line: {question}`
3. fallback: `Thought Quality | TQ | Your Confidence | Volume of the Claim | Brain Cells Used | Professor Know-It-All`
4. num_predict: 60
5. colour_prompt: In ONE short punchy sentence, roast this line with an over-the-top metaphor that makes it bigger and dumber than it is. No swearing, no emojis, no quotes, no labels. The line: {question}
6. colour_prompt_18: In ONE short punchy sentence, roast this line with an over-the-top, crude adult metaphor that makes it bigger and dumber than it is. No emojis, no quotes, no labels. The line: {question}

## equation roast (docbs.chains.0)
1. name: equation roast
2. min_share: 0
3. beats: openers | leadins | equations | plugins | breaks | hiddens | corrected | colour

## shorts: roast, riddle, reveal (docbs.chains.1)
1. name: shorts: roast, riddle, reveal
2. min_share: 0
3. beats: roasts | riddles | corrected

## roast, ratio, riddle (docbs.chains.2)
1. name: roast, ratio, riddle
2. min_share: 0
3. beats: roasts | equations | riddles

## chaos calculus loop (docbs.chains.3)
1. name: chaos calculus loop
2. min_share: 50
3. beats: insights | riddles|paradoxes | corrected | gutshots

## the response pipeline (docbs.chains.4)
1. name: the response pipeline
2. min_share: 80
3. beats: roasts | leadins | riddles|paradoxes | equations | plugins | corrected | gutshots

## 11/10 (docbs.chains.5)
1. name: 11/10
2. min_share: 110
3. beats: roasts | taunts | leadins | riddles|paradoxes | equations | plugins | breaks | hiddens | corrected | colour | gutshots | crew | signoffs

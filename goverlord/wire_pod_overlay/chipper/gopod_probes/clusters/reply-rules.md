# Reply rules
- **words:** reply rules, tighten, sentences, dropped sentences
- **ask:** What gets cut from a Brobot's answer?
- **chain:** Gominions (Doc and Pip) - how their answer is cleaned before it's said
- **see also:** brobot-brain, golden_answers
- **about:** 2026-11-03 (step 3): reply_tighten's word lists, from 01_brain.json. The sentence counts, the switches and brand_words stay in the settings for now.
---

## The word lists (reply_tighten)
1. more_words: tell me all | all about | explain | why | how does | how do | what happened | the story | everything | go on | tell me more
2. drop_sentences_with: the response reflects | this response | the response above | as an ai language model | would like me to continue | want me to continue | like to continue | use emojis in your responses | need more assistance | let me know
3. not_sentence_ends: Dr. | Mr. | Mrs. | Ms. | St. | vs. | etc. | e.g. | i.e.

## Brand words, "from: to" (reply_tighten.brand_words)
1. robots: Brobots
2. robot: Brobot
3. AI Wordplay! Explain the Math!: AI Wordplay! Explain the MAth!
4. Explain the Math!: Explain the MAth!

## Clean politics (clean_politics)
1. continents: africa | europe | asia | south america | north america | oceania | the middle east

## clean_politics words (clean_politics.words[])
1. zionist
2. zionism
3. hamas
4. hezbollah
5. intifada
6. free palestine
7. from the river to the sea
8. genocide
9. apartheid
10. occupation
11. occupied
12. maga
13. woke
14. antifa
15. trump
16. biden
17. kamala
18. netanyahu
19. putin
20. democrat
21. democrats
22. republican
23. republicans
24. liberals
25. conservatives
26. left-wing
27. right-wing
28. election
29. elections
30. abortion
31. pro-life
32. pro-choice
33. communist
34. fascist
35. nazi

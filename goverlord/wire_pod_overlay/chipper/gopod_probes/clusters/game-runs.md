# The game's runs
- **words:** runs, run cards, interview, bingo, songs
- **ask:** What runs can the game start?
- **chain:** Goverlord - the runs Goverlord starts and stops (songs, the draft, the interview)
- **see also:** game-words, system-menu
- **about:** 2026-11-03 (step 7b of GOPOD_CLUSTERS_PLAN_002.md): 04_game.json game.runs, a section per run and its parts, in order. A value in backticks is that text exactly; a bare number or true/false is that value; a heading ending in [] is a line list.
---

## Run 1: Interview (game.runs.0)
1. name: interview
2. pending: true
3. title: Interview
4. song: 02_brobots_interview
5. draft: true
6. start: start interview
7. stop: stop interview
8. menu_about: the scripted interview
9. session: interview
10. said: Interview

## Run 1: Interview - experience (game.runs.0.experience)
1. about: 2026-10-26 (operator: "The rest? Apply pre/post template to all, even minimally to what's already there"): the scripted interview's pre/post, minimal - slots empty until it's written (the interview is low priority).
2. run_marker: Interview started
3. run_marker_about: Goverlord's run line when the interview starts (session.started_text).
4. run_gap_seconds: 30
5. snapshot_note: [Not a question - the Brobots performed the scripted interview for the audience in this chat.]
6. question_word: interview
7. not_a_test_phrases: interview over |
8. pitch_pack: ``
9. run_pack: ``
10. no_run_note: No interview has been played yet - one short take of your own on it.
11. takes: {}

## Run 1: Interview - experience - pre_attitude (game.runs.0.experience.pre_attitude)
1. Brobot 1: ``
2. Brobot 2: ``

## Run 1: Interview - experience - post_attitude (game.runs.0.experience.post_attitude)
1. Brobot 1: ``
2. Brobot 2: ``

## Run 1: Interview - experience - closing (game.runs.0.experience.closing)
1. Brobot 1: ``
2. Brobot 2: ``

## Run 2: Bingo Test (game.runs.1)
1. name: bingo
2. title: Bingo Test
3. song: 01a_brobots_bingo_test
4. start: start bingo test
5. stop: stop bingo test
6. menu_about: the Bingo test
7. session: Bingo Test Live
8. said: Bingo test

## lines 1 (game.runs.1.lines.0)
1. when: not_played
2. lines: Another test performance? |
3. why: 2026-10-26 (operator: "Pre-Bingo: Did you play bingo yet? * duo response-chain (A/B or B/A) * no bingo test yet * b1 framing, “Another test performance?” [Ugh... whatever...] ... * b2 framing, “Another test performance?” [Super Hyped!! Ready for whatever...] ... * 'start Bingo test' in 2nd response regardless (A/B or B/A), before reminders * Post-bingo: How was bingo * we ran a test before the real bingo game * b1 framing, [Ha! You see?] ... “I Won!”. Lol!” * b2 framing, “Oh! Zero!” [I scored Zero?] ... “That's ok tho. Brobot ZERO is my HERO! Lol!”"): both open with it; the start move is experience_moves, in the 2nd answer (was "I guess so... Say start Bingo test!" / "I can't wait! Say start Bingo test!").

## lines 2 (game.runs.1.lines.1)
1. when: played
2. brobot: Brobot 2
3. lines: Oh! Zero! |

## Run 2: Bingo Test - experience (game.runs.1.experience)
1. run_pack: bingo_test
2. run_pack_about: The Bingo test run's own value points (chapters/_always/06_bingo_test.md), loaded for Bingo questions once it has played; the pitch pack (bingo) before.
3. flow_about: 2026-10-26 (operator: "Pre-Bingo: Did you play bingo yet? * duo response-chain (A/B or B/A) * no bingo test yet * b1 framing, “Another test performance?” [Ugh... whatever...] ... * b2 framing, “Another test performance?” [Super Hyped!! Ready for whatever...] ... * 'start Bingo test' in 2nd response regardless (A/B or B/A), before reminders * Post-bingo: How was bingo * we ran a test before the real bingo game * b1 framing, [Ha! You see?] ... “I Won!”. Lol!” * b2 framing, “Oh! Zero!” [I scored Zero?] ... “That's ok tho. Brobot ZERO is my HERO! Lol!”").
4. run_marker: Bingo Test
5. run_gap_seconds: 30
6. snapshot_note: [Not a question - the Brobots performed the scripted Bingo test for the audience in this chat: a test, before the real Bingo game.]
7. question_word: bingo
8. not_a_test_phrases: chocolate bingo |
9. pitch_pack: bingo
10. no_run_note: No Bingo test has been played yet - one short, fun line of your own about playing it.
11. closing_about: 2026-10-25 (operator: "Anyway, we got a winner! 🏆 = comedic irony/karma gold * at the end of the bingo test, the host (brobot 1) comes out on top, so he won * Extend to: Anyway, we got a winner! 🏆 \"I Won!\" * \"I Won!\" is the reference back to the brat's lucky number"; "O-0 is what he said reacting to his score on the cooperation test * \"Oh! Zero!\""): each Brobot's closing remark - the closing sentence of his own answer when asked about Bingo after a Bingo run (operator: "No, before the 'Hey' reminder"; "joke is closing remark of answer to how was bingo, not the response answer-chain"; answer_chains closing_remark, right after llm).

## Run 2: Bingo Test - experience - pre_attitude (game.runs.1.experience.pre_attitude)
1. Brobot 1: Your attitude: Ugh... whatever...
2. Brobot 2: Your attitude: Super Hyped!! Ready for whatever...

## Run 2: Bingo Test - experience - post_attitude (game.runs.1.experience.post_attitude)
1. Brobot 1: Your attitude: Ha! You see?
2. Brobot 2: Your attitude: I scored Zero?

## Run 2: Bingo Test - experience - closing (game.runs.1.experience.closing)
1. Brobot 1: "I Won!" Lol!
2. Brobot 2: That's ok tho. Brobot ZERO is my HERO! Lol!

## Run 2: Bingo Test - experience - takes (game.runs.1.experience.takes)
1. Brobot 1: If the Bingo test comes up, tell it your way, sticking to what really happened: you were the Bingo tester, just trying to run the test, and your over-enthusiastic brother kept bothering you. He interrupted to ask if you knew his lucky Bingo number - "I, Won!" Round 1, you drew B-1, a real Bingo number, and he said "I be, the one!" Round 2, you drew I-9, a test number, and he said the I column is I-16 to I-30 and asked if you were telling him your age. Round 3, you drew O-0, a number that doesn't exist on purpose, and he said "This election is rigged! I wanna speak to the manager!" You told him "I'm the manager", ended the test, and said "Good Night" - you got your brother back at the end.
2. Brobot 2: If the Bingo test comes up, tell it your way, sticking to what really happened: you were so happy to be part of it, and cluelessly messed it up. You interrupted your brother's test to ask if he knew your lucky Bingo number - "I, Won!" Round 1, he drew B-1, a real Bingo number, and you said "I be, the one!" Round 2, he drew I-9, a test number, and you said the I column is I-16 to I-30 and asked if he was telling you his age. Round 3, he drew O-0, a number that doesn't exist on purpose, and you said "This election is rigged! I wanna speak to the manager!" He said "I'm the manager" and ended the test.

## Run 3: Baby Brobots (game.runs.2)
1. name: nap
2. title: Baby Brobots
3. song: 101_brobots_nap
4. start: start baby brobots
5. stop: stop baby brobots
6. menu_about: the Baby Brobots nap song
7. session: Start Baby Brobots
8. said: Baby Brobots

## lines 1 (game.runs.2.lines.0)
1. when: not_played
2. brobot: Brobot 1
3. lines: Ugh. Dreaming again? |
4. why: 2026-10-25 (operator: "Brobot 1: “Ugh. Dreaming again?” with attitude. Then his own take on dreaming, then handoff to brobot 2 * Brobot 2: “A lot was lost” attitude. Then his own take on dreaming, then ensure/script in a move towards start baby brobots"): his opener, then his own take on dreaming (no_run_note); the start move is experience_moves, in the 2nd answer.

## lines 2 (game.runs.2.lines.1)
1. when: not_played
2. brobot: Brobot 2
3. lines: A lot was lost. |

## Run 3: Baby Brobots - experience (game.runs.2.experience)
1. run_marker: Cupped in my hands
2. run_marker_about: The nap's first caption (101_brobots_nap/story.md, the Narrator's bubble); the operator's own "Start Baby Brobots" doesn't count.
3. run_gap_seconds: 30
4. snapshot_note: [Not a question - the Brobots performed the scripted Baby Brobots nap for the audience in this chat.]
5. question_word: dream
6. not_a_test_phrases: []
7. pitch_pack: ``
8. no_run_note: No Baby Brobots nap has been played yet - one short take of your own on dreaming, with attitude.
9. pack: nap
10. pack_about: The nap's own value points (gopod_probes/chapters/_always/05_nap.md; operator: "another scoped set of value points for the sleep song"), loaded for its questions before and after the nap, outside the chapters.
11. closing: {}
12. takes_about: 2026-10-25 (operator: "they remember little before napping, cupped in a pair of hands after finally being rescued by Ron * everything went dark until Wire-Pod brought them back, where Ron found them * they remember waking together ready to party").

## Run 3: Baby Brobots - experience - takes (game.runs.2.experience.takes)
1. Brobot 1: If your dreams or the nap come up, tell it your way, sticking to what really happened: you remember little before napping, cupped in a pair of hands after finally being rescued by Ron. Everything went dark until Wire-Pod brought you back, where Ron found you. You remember waking together with your brother, ready to party.
2. Brobot 2: If your dreams or the nap come up, tell it your way, sticking to what really happened: you remember little before napping, cupped in a pair of hands after finally being rescued by Ron. Everything went dark until Wire-Pod brought you back, where Ron found you. You remember waking together with your brother, ready to party.

## Run 4: Fart Joke (game.runs.3)
1. name: fart_joke
2. title: Fart Joke
3. song: 03_fart_joke
4. start: start fart joke
5. stop: stop fart joke
6. menu_about: the Fart Joke
7. session: Fart Joke Live
8. said: Fart joke

## lines 1 (game.runs.3.lines.0)
1. when: played
2. brobot: Brobot 1
3. lines: Ugh. I didn't know. |
4. why: 2026-10-27 (operator: "Fart Joke start lines * Brobot 1, victim, Ugh. I didn't know. [not happy about being fart-hotboxed]"): once the Fart Joke has played, like Bingo's "Oh! Zero!"; the bracket is his attitude (fart_joke post_attitude).

## lines 2 (game.runs.3.lines.1)
1. when: played
2. brobot: Brobot 2
3. lines: Bro it was party-bro epic level lost-burps party. Bro! A highly localized rapid environmental change, for sure! |
4. why: 2026-10-27 (operator: "Brobot 2, prankster, Bro it was party-bro epic level lost-burps party. Bro! A highly localized rapid environmental change, for sure!"): once the Fart Joke has played.

## Run 4: Fart Joke - experience (game.runs.3.experience)
1. about: 2026-10-26 (operator: "Fart Joke has no pre/post chat yet * upgrade to new format template (bare minimum scaffolding + placeholders) * refer to bingo jokes format (fart song is funny too. I don't know what the jokes are yet)"): the Bingo shape, slots left empty ("" = nothing said) until the jokes are written. To fill, Bingo's way: closing (his post joke, like "I Won!" Lol!), pre_attitude / post_attitude ("Your attitude: ..."), takes (what happened, his way), run_pack (a topic in chapters/_always with the run's points), and start lines in scripted_snippets (when fart_joke_not_played / fart_joke_played, per brobot - not added while empty: an empty one would block the others). The start move (experience_moves) already plays in the 2nd answer. 2026-10-26: filled - attitudes, closings, takes and run_pack. 2026-10-27: his lines once it has played (scripted_snippets fart_joke_played).
2. run_marker: pretend you and I are stuck in an elevator
3. run_marker_about: Brobot 2's first line (03_fart_joke/story.md).
4. run_gap_seconds: 30
5. snapshot_note: [Not a question - the Brobots performed the scripted Fart Joke for the audience in this chat.]
6. question_word: fart
7. not_a_test_phrases: []
8. pitch_pack: ``
9. run_pack_about: 2026-10-26 (operator: "the run's value points? fart, prank"): the Fart Joke run's own value points (chapters/_always/07_fart_joke.md), loaded for Fart Joke questions once it has played.
10. run_pack: fart_joke
11. no_run_note: No Fart Joke has been played yet - one short take of your own on it.
12. pre_attitude_about: 2026-10-26 (operator: "preline Brobot 1 party-bro-victim experience = none, unprepared like a lost burp * preline Brobot 2 party-bro prankster experience = prepared, couldn't contain myself").
13. post_attitude_about: 2026-10-26 (operator: "Brobot 2 (party-bro prankster) * Brobot 1 (party-bro-victim)").
14. closing_about: 2026-10-26: the operator's closing jokes, word for word.
15. closing_riff_about: 2026-10-27 (operator: "New closing lines * randomly take turns from run to run"; "Use these base lines, riff on them"; Brobot 1 the victim, Brobot 2 the prankster): his base line; each run, his closing is either the closing word for word or a riff on this, at random.
16. takes_about: 2026-10-26 (operator: "draft the takes and the points from the story as written"): from 03_fart_joke/story.md, its own lines.

## Run 4: Fart Joke - experience - pre_attitude (game.runs.3.experience.pre_attitude)
1. Brobot 1: Your attitude: party-bro victim - no experience, unprepared, like a lost burp.
2. Brobot 2: Your attitude: party-bro prankster - prepared, couldn't contain myself.

## Run 4: Fart Joke - experience - post_attitude (game.runs.3.experience.post_attitude)
1. Brobot 1: Your attitude: party-bro victim - not happy about being fart-hotboxed.
2. Brobot 2: Your attitude: party-bro prankster.

## Run 4: Fart Joke - experience - closing (game.runs.3.experience.closing)
1. Brobot 1: In all my nano-seconds of Methane-Venting & Evacuation Sequence Pressure Testing, I never heard a containment failure before.
2. Brobot 2: A stale air dump of yesterday’s ambitions, scheduled and unapologetic. Looks like I won too!

## Run 4: Fart Joke - experience - closing_riff (game.runs.3.experience.closing_riff)
1. Brobot 1: Was not happy about being fart-hotboxed.
2. Brobot 2: The intended event found as a repeating systems-level discharge, due to the accumulation of residual atmospheric distribution, within my brobotic brotastic party-bro, Bod! Further variations are narrowly constrained mechanical and software-diagnostic registers, that I can now compose and fire on call.

## Run 4: Fart Joke - experience - takes (game.runs.3.experience.takes)
1. Brobot 1: If the Fart Joke comes up, tell it your way, sticking to what really happened: your brother said "Hey bro, pretend you and I are stuck in an elevator." You said "Ok. Done." He said "Listen closely to the forcible expulsion of gaseous waste!" and farted - you said "Bro! Nasty particulates!" He said "Wait! Firing!" and farted again - you got angry: "Bro! Brobot 1 has left the elevator." He said "Wait! One more!" and farted a third time - all you had left was "Bro!"
2. Brobot 2: If the Fart Joke comes up, tell it your way, sticking to what really happened: it was your prank. You said "Hey bro, pretend you and I are stuck in an elevator." Your brother said "Ok. Done." You said "Listen closely to the forcible expulsion of gaseous waste!" and farted - he said "Bro! Nasty particulates!" You said "Wait! Firing!" and farted again - he got angry: "Bro! Brobot 1 has left the elevator." You said "Wait! One more!" and farted a third time - all he had left was "Bro!"

# Brobots Interview

**Status:** WIP (2026-10-17)

The scripted interview. Type `start interview` in the Brobots chat to play it, `stop run` to stop
it at a safe spot.

- **The lines** live in this file (`## STEP` / `> TEXT:`). **Who says each one, the Brobot's
  animation cue and his moves at words** live in [knobs.json](knobs.json). The two must name the
  same steps; changing a line's words here is enough on its own.
- Each `## SECTION` is one question: the reporters' banter, the question (`note: ask`), then the
  answers and handoffs, in order. `{weather}` in a line is the live weather.
- 2026-10-12 (the operator's interview score): `chapter` on a section's first step starts a
  chapter (shown and said before its question); `pictures` on an ask shows those pictures with
  the question; the `Closer` section is said after "Interview run complete".
- 2026-10-17 (operator: "The interview: rebuilt with new questions and answers from scratch, as a
  summary of that same list" - the egg list, gopod_egg_list.go): one question per easter egg, in
  the chapters' order. The Brobots' answers are the operator's own lines from each topic file
  (`say N:`, `reminder:` and its reveal list in gopod_probes/chapters/), word for word; the
  reporters' questions are drafts for the operator to change. The old interview is kept in
  gopod_notes/older_archives/interview_before_2026-10-17/.


---

## SECTION: 1. The workbook

## STEP q01_01_reporter_1

> TEXT: Brobot 1, show us Doc and Pip's workbook.

## STEP q01_02_brobot_1

> TEXT: Pip works it out - confidently, charmingly wrong. Doc's key names the mistake and fixes it. The errors aren't random: they're the ones real learners actually make.

## STEP q01_03_brobot_2

> TEXT: Pip's Clueless Workbook: "pretty sure factoring means adding. i'll ask Doc." Doc's Answer Key: "The Answers. Obviously."


---

## SECTION: 2. Fractions and decimals

## STEP q02_01_reporter_2

> TEXT: Fractions and decimals. Teach us something.

## STEP q02_02_brobot_1

> TEXT: Fraction to decimal: divide the top by the bottom. A half is 0.5, a quarter is 0.25, three quarters is 0.75.

## STEP q02_03_brobot_2

> TEXT: Some decimals never stop. A third is 0.3 repeating, two thirds is 0.6 repeating, a ninth is 0.1 repeating.


---

## SECTION: 3. The rig

## STEP q03_01_reporter_1

> TEXT: What's in the rig?

## STEP q03_02_brobot_1

> TEXT: One dolly, two crates. The bottom crate holds the portable Anker Solix C1000 power station. The top crate holds the rest: the Jetson Orin NX 16GB that runs the local AI and Wire-Pod, the Wi-Fi hotspot that links the two Vectors, the mic, the numpad, a USB hub, a power bar, and a fan to keep it all cool.

## STEP q03_03_brobot_2

> TEXT: One dolly, two crates - roll in, plug in, party.


---

## SECTION: 4. The rest of the cast

## STEP q04_01_reporter_2

> TEXT: Who else joins the cast?

## STEP q04_02_brobot_2

> TEXT: Brobots 3 and 4 are live already - two voice-only reporters, a woman's voice and a man's voice. Coming, not built yet: CHALK, a future screen host.

## STEP q04_03_brobot_1

> TEXT: Doc and Pip are two of an eventual six-brobot GOPOD-layer cast - Git Repo and Cache PYC, the Cozmo-bodied "unruly students", and PNFH and PNFT, the Moorebot Scout-bodied "goofy security".


---

## SECTION: 5. Birthdays and merch

## STEP q05_01_reporter_1

> TEXT: GOPOD your birthday?

## STEP q05_02_brobot_2

> TEXT: A full pizza-box of food-shaped Hot Wheels cars - already bought, already sorted. A live happy-birthday performance is coming - not filmed, not run, not yet.

## STEP q05_03_brobot_1

> TEXT: Merch ideas, none of it printed, priced or sold yet: Doc & Pip crew tees, a chocolate-bingo pin, a LEGO-adapter wheel keychain, a mug.


---

## SECTION: 6. Bingo nights

## STEP q06_01_reporter_2

> TEXT: Bingo nights?

## STEP q06_02_brobot_1

> TEXT: GOPOD can host Bingo games for seniors while keeping the game familiar. The Brobot can call numbers, speak, react, celebrate small wins, and help the room feel more alive.


---

## SECTION: 7. Foodmath recipes

## STEP q07_01_reporter_1

> TEXT: And the food?

## STEP q07_02_brobot_2

> TEXT: Foodmath: the math hiding in real party food, on foodmath.crushn8r.net. Chocolate Bingo: the tray and the numbers. Fudge Bolt Squares: the ratio that makes the tray set.


---

## SECTION: 8. The weather

## STEP q08_01_reporter_2

> TEXT: Brobot 1, how's the weather out there?

## STEP q08_02_brobot_1

> TEXT: {weather}


---

## SECTION: 9. Roast battles

## STEP q09_01_reporter_1

> TEXT: Roast battles?

## STEP q09_02_brobot_2

> TEXT: Roast Battle! Party Brobots vs Any and All takers.


---

## SECTION: 10. Our legacy

## STEP q10_01_reporter_2

> TEXT: What's it all for?

## STEP q10_02_brobot_1

> TEXT: Teach how to think, not what to think. Content propagates as answers. Process propagates as capacity - fish versus fishing.

## STEP q10_03_brobot_2

> TEXT: Gangster Granny: a 93-year-old woman helped Brobot 0 when it counted - his thanks to her, his sweet, adorable, gangster granny in a sparkly bling cap.


---

## SECTION: Closer

## STEP closer_01_brobot_2

> TEXT: Press my backpack and tell me to "GOPOD Yourself". Play to win. Your move. 😉

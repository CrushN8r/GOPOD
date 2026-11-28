# The cast and the game

Part of [ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md) - [Concepts](concepts/README.md) · [Songs](02_SONGS.md) · [The session and runs](03_SESSION.md) · [The Brobots chat](04_CHAT.md) · [Voices and sound](05_VOICES_AND_SOUND.md) · [The cast and the game](06_CAST_AND_GAME.md) · [The system menu, your stuff, the settings](07_SYSTEM.md) · [Golden rules, troubleshooting, tests](08_RULES_AND_TESTS.md) · [The gears](09_GEARS.md)

<details><summary>Game card: the GOPOD game</summary>
<!-- gopod topic: 04_the_game -->

# the GOPOD game
- **chapter:** Show and tell

- **words:** gopod game, game engine, the game, value points, grand prize
- **ask:** Tell me about the GOPOD game
- **summary:** AI Wordplay! Explain the MAth! is a game | one session score: topics, value points, eggs | eggs are prizes that pop up | grand prize: every value point revealed
- **source:** alias_play_studio/06_CAST_AND_GAME.md
- **ask 1:** Brobot 1, tell me about the GOPOD game
- **ask 2:** Brobot 2, in the GOPOD game, what is an egg?
- **ask 3:** Brobot 1, what's the grand prize in the GOPOD game?

---

The GOPOD game. Reveal these, in order, in your own words:
1. "AI Wordplay! Explain the MAth!" is a game: a question can reveal a value point, and the session keeps one score - topics, value points and eggs.
2. An egg is a prize - pictures or a card - that pops up when a line's words find it; each chapter is one egg.
3. The grand prize: every value point revealed.

</details>

## The cast
- **Who's who:** the numpad map ([`numpad-mapper.md`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/company/99_system/numpad-mapper.md)) and
  each persona's own card in [`personas/company/`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/company/) - the folders are the chain of
  command: Goverlord, KEYS the Butler, the Gomads and interns, the Gominions.
- **Four personas, two robots:** Brobot 1 and Brobot 2 are party-bro GOPOD reps at the dial's 0%;
  DOC and PIP are the host and sidekick at 100%, answering from the same two robots.
- **The stereo dial** (`host_pairs`): 0 is all base persona, 100 all host persona, anything between
  a blend - said in the prompt as a share.
- **One persona schema for everyone:** a card's head is its labels, written once; every setting,
  prompt and header fills from them as `{name}`, `{nickname}`, `{job}`. Every restart checks every
  card against the schema and the source check walks each participant layer by layer.
- **One intro recipe:** the code says the opening and the cast line; the model writes only the middle.
- **Company time is formal names;** nicknames are for banter, anyone with anyone. Guards and hosts
  with a nickname are headed "Nickname (Name)".

## Managers
- **CHALK** runs the display (QR codes, the surprise eggs, the visual board); **PLAYHEAD** runs time
  and place (the weather, where and when); the **interns** take the USB lines and the search box; the
  **guards** watch the Brobots' batteries; **KEYS** is Goverlord's voice in the chat; **Goverlord**
  answers only when called, and hands the trivial down.
- **Two suits:** CHALK and PLAYHEAD are indoor anchors (NumLock on) and outdoor reporters (NumLock
  off) - same persona and voice, a different bubble.
- **Inside a Brobot's bubble:** asking Brobot N isn't a call to a manager, so the managers' parts
  sit inside his bubble - PLAYHEAD's at the top, CHALK's QR codes at the bottom. Asking a manager
  gets the manager's own bubble.
- **A web address in a line** gets a QR made for it, shown once.

## Eggs
An **egg** is a prize - pictures or a card - that pops up when a line's words find it. Which topics
are eggs, their names and cards: [`easter_eggs.md`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/behaviours/easter_eggs.md).

- **Chapter eggs** are the game's chapters, found by their topics' words, named in the reminders
  until found. **Surprise eggs** pop up but are never named.
- **Live chat only,** never during a run; each egg pops up once a session, then only its summary.
- **The egg bubble:** ringed by marquee bulbs (Chase, Groups, Pairs, Twinkle), Goverlord's
  congratulations said as it shows, then the prize, then "MouseOver to Zoom In" under anything that
  zooms. Two eggs in a line are two bubbles, in the order their words came.
- Goverlord's congratulations is made from kept pieces; only the egg's title is made live.

## The game
"AI Wordplay! Explain the MAth!" is a game of query for value points.

- **Value points** are each topic's numbered lines - a topic with none is one point - locked in
  `value_points_golden.json`; `system-value-points` shows them, `system-value-points lock` locks a
  change on purpose.
- **Which topics count, the score and the grand prizes:** the game cards and the settings
  (`value_points`, `game`); the topic map and funnels: [`topics-map.md`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/behaviours/topics-map.md).
- **The lightning:** a question pulls every point that shares its words, rarer words harder; of a
  topic's points not said yet, the one sharing the most words is given.
- **The gravitational pull:** however far it's pushed, every answer steers back to the points still
  to say - his pre-build context, his notice ("Next up"), and his reply's end ("Hey, that reminds
  me..."), all naming the same points. Reminder rounds start again until every topic is covered.
- **The grand prizes** are said once each, after the answer that won them; the thanks ending resets
  the game.

## Answers
- **Answer shapes:** one shape per answer type (weather, topic, free, math, thanks); every answer is
  checked as it's made (`BROBOT_SHAPE ... ok` or `DRIFT`), and every shape at every place runs at
  deploy. The shapes and chains: [`brobot-answers.md`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/behaviours/brobot-answers.md).
- **Response chains:** a line to one Brobot is one answer; to both, A then B - A ends handing over
  with a nickname, B opens with a comeback. Nicknames: [`nicknames.md`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/behaviours/nicknames.md).
- **Golden answers:** a golden question gets its line word for word first
  ([`golden_answers.md`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/behaviours/golden_answers.md)).
- **Experiences:** each run's before-and-after chat - whether it has played, its takes, its closing
  riffs, its own value points: the run cards, [`game-runs.md`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/behaviours/game-runs.md). A new run is one
  new card.
- **Off-topic lines** get wordplay on their own words, then a bridge back to GOPOD.
- **Greetings** are checked against GOPOD's clock: "Good evening? It's morning here!"

## DocBS, roasts and spice
- **DocBS:** at the dial's 100%, Doc's answer is a chain of beats - roasts, riddles, the fake
  equation, the gut-shot - around his reply, each beat a shuffled pool line, never repeated until
  the deck runs out; the dial decides how long a chain he gets. Recipes and pools:
  [`dials/docbs/`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/behaviours/dials/docbs/).
- **Persona chains:** the same template for every participant, their own pools.
- **The roast dial** - clean to 18+ afterburner, the room decides: [`roast-dial.md`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/behaviours/roast-dial.md);
  **spice** (swears and exclamations): [`spice.md`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/behaviours/spice.md).
- **18+ in public, starred:** expletives are starred in the repo (f\*ck); the uncensored copies are
  private and used when there. A starred word still matches the word said.

## Running with one Brobot
With one robot, Brobot 1 runs solo: lines to All go to him, and his prompt drops his brother.
Everything public names a robot by its label, ESN01 or ESN02; who they are is worked out at start.

[Back to ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md)

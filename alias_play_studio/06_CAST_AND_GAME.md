# The cast and the game

Part of [ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md) - [Concepts and terms](01_CONCEPTS.md) · [Songs](02_SONGS.md) · [The session and runs](03_SESSION.md) · [The Brobots chat](04_CHAT.md) · [Voices and sound](05_VOICES_AND_SOUND.md) · [The cast and the game](06_CAST_AND_GAME.md) · [The system menu, your stuff, the settings](07_SYSTEM.md) · [Golden rules, troubleshooting, tests](08_RULES_AND_TESTS.md) · [The gears](09_GEARS.md)

## 12. Managers: CHALK and PLAYHEAD

The managers take load off the Brobots. They are GOPOD's other chat participants (Gomads):
`managers` in the settings, `gopod_managers.go`.

### 12.1 Three personas, three jobs

- **The Brobots' shared persona** = GOPOD reps. No system jobs.
- **The reporters' shared persona** = scripted lines and voices. Reporter 1 and 2's lines are
  the scaffolding for CHALK and PLAYHEAD's live chat personas.
- **The managers** take the system jobs.
- **The four hosts, one persona shape** (RULE,: Brobot 1 = Doc, Brobot 2 = Pip, the
  GOPOD hosts; the Brobots = party bros, GOPOD demo reps). `host_personas`: `name` (said only
  when his base text doesn't already say it), `role`, `steps`, `signature`, `handover`. Added
  after his `personas` line (Brobots) or `managers.personas` prompt (CHALK, PLAYHEAD)
  (`gopodHostPersonaText`, `gopod_brobot_personas.go`). Doc delivers the point (read the
  question, name the point, build it from the first rule up, end so it couldn't be any other
  way). Pip is the audience's pied piper (catch the point, ask the shy question or guess wrong
  the real way, catch up and say the wow, hand over; signature "raise your hand if you got that
  wrong too!"). Doc and Pip never name each other in the prompt ("my brother").
- **Pre-build context** (`prebuild_context`, `gopodPrebuildContext`): every Brobot
  answer's prompt gets "BEFORE I ANSWER:" after his facts - I am Brobot N (Doc / Pip); my role
  is a GOPOD rep; my reminders-list (the same names as his "Hey, that reminds me" line); what was
  asked of me; how to best answer (his `steps`). A line with an empty token is left out.
- **Pairs and the stereo dial** (`host_pairs`: Doc & Pip, CHALK & PLAYHEAD): `partner_as`,
  `banter`, and `dial` - like a stereo balance, 0 = all left (the base: `left`, party-bro GOPOD
  demo rep / scripted reporter), 50 = pointing up (50/50), 100 = all right (the host persona).
  Both at 60 = 40/60. Said in the prompt as "My stereo dial: 40% party-bro GOPOD demo rep, 60%
  Doc."

### 12.2 Who does what

| Manager | Voice, look | Jobs |
|---|---|---|
| **PLAYHEAD ZUMINTU** (KP8) | Reporter 2's voice, tape-brown bubbles | Time and place: "weather" - "I'll help you with that.", the weather line; "what time / where are we / WAWN / what's the date" - the time and place line (`time_place_text`); location QR codes. His job role is in his persona prompt (`managers.personas.PLAYHEAD.prompt`): "ph" for PLAYHEAD, a/0/b the working points to zoom in and out from, ZUMINTU = zoom into, like editing an audio file end to end. |
| **Git Repo / Cache PYC** (KP3/KP4, Cozmo students), **Intel 404 PNFH / PNFT** (KP5/KP6, Moorebot Scout security guards), **Cameo** (KP9, the social media Gomad) | their own voices (above), the default host bubble look - Cameo's bubble is Goverlord level (`bubble_kinds`, `host-cameo`) | Scaffolding: CHALK's shape - `managers.personas` prompt and who line, no scripted lines; `host_personas`, `host_pairs` ("Git Repo & Cache PYC", "Intel 404 PNFH & PNFT"). Answers from the language model, like KP7/KP8. Intern jobs: Git Repo (ethical hacker) says the USB lines - mic and numpad missing or back (`numpad_ptt.usb_host`) - and takes "usb devices / what's plugged in / lsusb" (kind `usb`, `managers.usb_text`); Cache PYC (meme hustler) draws the search results box; Intel 404 PNFH / PNFT (security guards) warn when a Brobot's battery runs low (`battery_watch.guards`). The hand-over: a take with `via` - that Gomad says `managers.handoff_text` ("Git Repo will help you with that."), then the intern takes over (USB: via CHALK; search results: `system_search.handoff_host`, PLAYHEAD). |
| **CHALK BORED** (KP7) | Reporter 1's voice, chalkboard bubbles | Display: QR codes, the surprise eggs (Roast Battle, Halloween, camera feeds) and the visual board, drawn on its board (`display_host`). |

- **Two suits (2026-10-30):** CHALK and PLAYHEAD are the indoor news anchors: their own
  bubbles, chalkboard green and vintage tape. They're also the outdoor reporters: a run's
  reporter lines (the `_llm_song` lines, spoken by their reporter, `managers.personas.<host>.reporter`)
  wear the reporter suit - suit grey (`--reporter-suit-bg`), the anchor's colour as the border
  (`reporterSuit`, page). Live, KP7 / KP8 with **NumLock off** are CHALK and PLAYHEAD as
  reporters: same persona and voice, answering in the reporter suit (the numpad map's
  `numlock_off` layer, `"suit": "reporter"`). Locked by `TestNumpadNumlockOffReporter`.
- **Talking to a manager:** KP7 or KP8 (voice only): your line, then the persona's answer from
  the language model - its identity and its reporter's scripted lines as the style - in its
  reporter voice and colour (`managers.personas`, `gopod_personas.go`). "Who are you?"
  (`who_words`) gets your own line, word for word (`who_line`).
- **A manager is still a participant:** a persona's answer gets the Brobots' shared chat
  (`gopodPersonaMessages`).
- **Takes:** a take's words (weather, time/place) asked of CHALK or PLAYHEAD go to the take
  (PLAYHEAD's real weather), not the persona's language model.
- **Manager bubble:** a manager's next line joins its bubble from the exchange (its own bubble, or
  its part in a Brobot's) (`hostLineInto`). Managers only - reporters' bubbles are unchanged.

### 12.3 Managers inside a Brobot's bubble

Asking CHALK, PLAYHEAD, the Brobots or All is a take-over trigger. Asking Brobot N is not a
direct call to a manager, so the managers' parts sit inside Brobot N's bubble:

- **A line to Brobot N:** PLAYHEAD's part (weather, time/place, search results, a location QR)
  opens his bubble at the top, before his "Got it!" (`nestShell`, `nestBox`). His reply fills the
  same bubble below. CHALK's QR codes go in at the bottom, after his response (`chalkQrs`), each
  with its name (the call to action's too). His own highlight ignores the nested parts
  (`ownSays`).
- **To the Brobots or All:** the parts go inside the bubble of the Brobot answering first. Once
  the managers have dealt with him, their jobs are done; the other Brobot answers like before, no
  QR note.
- **To a manager:** the managers' own bubbles.
- **Newsletter:** CHALK's part with the `qr_codes.second` QR and its title ("Newsletter:
  crushn8r.ca/subscribe"), inside the Brobot's bubble, on the inner side of the newsletter line -
  after it when it opens the answer, before it otherwise (`chalkCta`). Any newsletter line counts,
  not only the call to action's own words. CHALK QR boards are top/bottom border phobic inside a
  Brobot's bubble.
- **Location QR:** a web address in the line - PLAYHEAD says "I can help with that. QR code
  generated." and shows a QR made for it (`url_qr`, `webroot/qr/made/`); the Brobot's prompt says
  it's shown (`qr_note`). When the first Brobot's answer ends, CHALK says "I can help with that.
  QR code displayed." and shows the same QR (`qr_display_text`, `gopodManagerQRShow`) - at the
  bottom of his bubble for a line to him, her own bubble otherwise - then the other Brobot answers
  or doesn't. The QR is made with the operator's own Python libraries (`~/.local`), since
  Wire-Pod runs as root.

---

## 13. Eggs

An **egg** (easter egg) is a prize - pictures or a card - that pops up when a line's words find
it. The Brobots are GOPOD reps, and eggs are how they show GOPOD's features ("AI Wordplay!
Explain the Math!" is GOPOD as a whole).

### 13.1 Two kinds of eggs

- **Chapter eggs** are the live run's chapters. Each chapter is one egg, titled by its chapter
  title ("Easter Egg '<chapter title>' found."), found by its topics' words (the topics in a
  chapter are its triggers), with those topics' pictures, then the value points. No picture = no
  pictures and no zoom line. They're on the egg list: the reminders name the ones not found yet.
  (`gopodTopicChapter`, `TestChapterEggs`.)
- **Surprise eggs** (`topic_pictures.not_listed`: Roast Battle, Halloween, camera feeds) pop up
  when their words come up but are never named. The roast battle (in no chapter) is its own egg.
  Halloween: `_always/03_halloween.txt`, an emoji card (`cards.halloween`), no picture; its say
  lines are drafts. Camera feeds: `chapters/_always/04_camera_feeds.md` - "camera feed(s)" shows
  the Brobots' live camera streams on CHALK's board (a picture path starting with "/" is used as
  it is).
- **Explain the Math** is the livestream show's egg ([8.5](04_CHAT.md#85-explain-the-math-the-contest-format)):
  "AI Wordplay Explain the Math" - "[contest submission text]. Explain the Math" - with
  "🏆 Win to get Doc's approved 'Almost Approved' top prizes!" as its prize (`math.egg_name`,
  `popup_title`).
- **Not eggs:** Weather and Get Photo (the weather face shows in the Brobot's own bubble; the
  photo's 3 pictures in a plain picture bubble). `topic_pictures.not_eggs` (the chapter
  "Websites, themes and the channel") never pops up and is not on the egg list; the chapter stays
  in the interview.

### 13.2 When eggs pop

- In the live chat only - never while a draft plays or any run is going (a song, the interview, a
  replay). Locked by `TestPopupsOnlyLive`.
- Each egg shows up to `topic_pictures.max_per_session` (5) times per Wire-Pod start.
- A topic shows at most `topic_pictures.max_shown` (6) of its pictures at once, picked at random
  each time. Birthday has 20 (seven pizza-boxes of food-shaped Hot Wheels cars, single cars, LEGO
  ice-cream minifigures); BBQ (`12_booking/07_bbq.txt`, "GOPOD can do BBQ's too!") has 6. Photos are
  web-sized: 1600 px longest side, JPEG quality 85, no location data; full-size originals outside the repo.
- A trigger inside another topic's longer trigger doesn't count (`gopodTriggerOwnMatch`). One egg
  per line, not two, from one phrase.

### 13.3 GOLDEN: the egg process

- Every egg a line finds shows in the order its word comes in the line, each its own egg bubble.
  Two eggs are two egg bubbles.
- **The egg bubble:** the Brobot's bubble (his colours) ringed by the marquee, holding
  Goverlord's grey block ("Goverlord: GOPOD Brain", then `topic_pictures.egg_text`, two lines:
  "Congratulations!" / "Easter Egg '{name}' found." - {name}: the egg's name, else the word that
  found it; a colon instead of the period before a card, its card lines under it), then the
  prize, then `egg_zoom_text` "MouseOver to Zoom In" on its own line under a prize that zooms (a
  picture, a photo). No header above it.
- Goverlord says the congratulations as it shows, revealed as it's said (`voiceLitLine`), zoom
  kept, after the question has been read out, before the next egg. The zoom line, said last in
  the same voice, lights with it too: lit once the congratulations' last word is
  said, shown said when the voice ends.
- **Order:** the prize and the spoken congratulations, then the thinking line, then the reply.
  His reply comes after, in his own marquee bubble with his "Brobot N:" header.
- **To All:** when the question found an egg, every Brobot's answer wears the marquee - each in
  his own marquee bubble with his "Brobot N:" header - until the next operator line.
- **The marquee:** bulbs ring the egg (always animated, even with the PC's reduce-motion
  setting), placed along its rounded border and cycling Chase, Groups, Pairs and Twinkle every
  5 s; they're laid out again whenever the bubble changes size, their count following the
  border's length. The marquee ends after a voice (backpack) question's exchange and after a
  reload.
- A scripted run shows its pictures without eggs.

### 13.4 Egg names, cards and triggers

- Egg names by topic: `topic_pictures.egg_names` (Birthday - the Hot Wheels box; Fractions - the
  Doc & Pip workbook; Roast Battle; Mobile Gear - the GOPOD stack picture, found by the rig's
  words like mobile and gear, not by "ai"), else the word that found it.
- An egg with no picture shows a card (`cards`: the Roast Battle's "Party Brobots vs any and all
  takers").
- **Granny:** "gangsta" finds her egg too.
- **Chocolate is food, bingo is a game:** "chocolate bingo" is the food topic's words
  (foodmath_recipes, the recipes picture), so its "bingo" does not load the bingo topic too.
  "bingo" alone is the bingo game egg (the Chocolate Bingo kit and card pictures): bingo hosting,
  then "Hey, you can also play Chocolate Bingo...". The Bingo line is "Sure, let's play! Say start
  Bingo test!"; "experience" is not a Bingo egg word.
- **Birthday:** value point 4 - Brobot 0 wants to ask a local bakery to make him a big cake based
  on his foodie cars.
- **Brobot 0:** the topic is found by "crushn8r", not "ron". The flare4.com QR shows in a bubble
  about his past ("nightlife", "Mississauga").
- **Websites:** the websites topic's picture is the crushn8r.ca QR (`pictures/qr/crushn8r_ca_qr.png`),
  in chapter 9's egg.
- **Doc & Pip workbook cards:** the six cards (1A Factoring Warm-Up, 1B One Big Slot, 1C Speed
  Round, 2A Doc's Answer Key, 2B Flip to the back, 2C No credit for close), cut from
  DocPipMath.pdf's pages turned upright, are `webroot/pictures/doc_pip_math/docpip_*.jpg` - the workbook egg's
  pictures (`04_how_doc_and_pip_teach/02_doc_and_pip_workbook.txt`), each with hand-written points
  from its own words (`gopod_picture_descriptions.json`).
- `system-check` warns when a topic's trigger word is, or holds, a Wire-Pod intent phrase
  (`intent-data/en-US.json`) or a GOPOD command phrase (`native_cube.intents`).

### 13.5 The egg list and reminders

Every chapter egg with a prize (pictures or a card), in the chapters' order, then
`value_points.reminder_eggs` (Explain the Math). An egg found since Wire-Pod's start leaves the
list. Each "Hey, that reminds me..." names at least one egg still to find; a question on no topic
still ends with two (not a thank-you); each in turn, then round again. `gopod_egg_list.go`,
`TestEggList`.

**With one Vector - the single-bot setup** (`brobots_setup`). At Wire-Pod's start the Brobots are
the robots Wire-Pod knows: `brobot_1` / `brobot_2` (an ESN) when set; else GOPOD's own two when they're
there; else Wire-Pod's robots in order. One robot alone is Brobot 1 and GOPOD runs solo: a line to All
(typed, the menu, KP0) goes to him alone; Brobot 2's key goes to him; his prompt drops his brother (the
cast, the brother banter) and gets `solo_note`. Wire-Pod writes who's who to `gopod_brobots_now.json`
(live state, gitignored); the songs read it - solo, Brobot 2's lines are said on the speaker. No robot
known: nothing changes. `TestBrobotsSetup`.

**The robots' labels.** Everything public - the code, the settings, the page, the docs, the scripts - names a Brobot's robot by its label: ESN01 is Brobot 1's robot, ESN02 Brobot 2's. Who they are is worked out at Wire-Pod's start: the operator's own two from a private file in his home (`~/.gopod_esns.json`), never in the repo; with none, the robots Wire-Pod knows. Wire-Pod changes a label to the robot's serial at its door (every `/api-sdk/` call and the camera stream, `gopod_esn_labels.go`), so the page, a song or a script just sends the label; what the page reads carries the label (`gopod_speaking_now.json`), and the admin page's battery icons get the order from `/api-sdk/gopod_brobot_order` (admin only). A direct robot connection (the direct-SDK tools, the Bingo reactor, `gobingo`) looks the label up itself (`gopod_brobots_now.json`, then the private file). `TestNoRealESNInRepo` fails if a tracked file names a real serial; older commits still hold them (a history rewrite is a separate call).

**DocBS - Doc's brain wiring.** Brobot 1's answer, when Doc's share of his stereo dial says so
(`docbs.share`: -1 = the Doc & Pip dial, 60 = 60% of his answers; 100 = every one), is the equation
roast around his reply: the `docbs` part puts the frame in front - an opener, a lead-in with a
nickname, the equation (A × B) ÷ (C), the plug-in with meme numbers, the ERROR, the hidden multiplier
("currently sitting at 69,420"), the corrected formula - each its own thought bubble; his reply (the
value point, his colour) is the kill shot; `docbs_close` ends it: "The math has spoken. And it's
...". The frame comes from `gopod_probes/docbs/` - one pool file per beat, shuffled, never repeated
until the deck runs out; `_clean` / `_18` by the roast dial. gemma fills only one short line (the
equation's name, acronym, three things and a nickname - about 2 s), else `docbs.fallback`. An Explain
the Math line the math engine can't work out always goes DocBS; a computable one keeps its exact
steps. `TestDocBS`. **The Master Library and its chains (2026-10-30):** `gopod_probes/docbs/DOCBS_LIBRARY.md`
is Doc Squawkadoodle's library; its catchphrase bank is in pools (roasts, riddles, insights,
dismissals, crew, paradoxes, taunts, signoffs, gutshots). Each DocBS answer is one **chain**
(`docbs.chains`): its beats in order before his reply, picked from a shuffled deck, never the same
chain twice until the deck runs out. The dial sets the length: a chain needs `min_share` - the
short ones (equation roast; roast, riddle, reveal; roast, ratio, riddle) from 0, the chaos calculus
loop at 50, the full response pipeline at 80, and 11/10 (`docbs.share` 110) with the crew hype and
the sign-off. `TestDocBSChainsByDial`. **Persona chains (2026-10-30):** the same template for every chat participant
(`persona_chains`, `gopod_probes/persona_chains/<persona>/`): a persona's answer can be one of its
chains, its pool lines around its reply (the beats before `reply` first, the ones after it close),
as often as its `share`. The Brobots get it through the answer-chain parts `persona_chain` /
`persona_chain_close`, the Gomads around their answer. Empty to start: the lines are the
operator's. DocBS stays Brobot 1's own. `TestPersonaChains`.

**Public GOPOD, the top shelf**: everything 18+ is in the public repo, expletives starred
(f\*ck) - the human swear list and the DocBS `_18` pools; their uncensored copies are private and
gitignored (`zgopod_human_swears.json`, `gopod_probes/docbs/z*.txt`, the zKnobs.json pattern) and used
when there. Quotes in comments and setting descriptions are starred too.
The working lists are starred the same way and still work: a pronunciation entry keeps its
vowels in order (`"from": "sh*t", "vowels": "i"`), so it matches the word said and the word written
starred - never another word (shot, shut). A starred word in a value point (b\*llsh\*t) is said and
struck the same. One rule in Go and in the song runner (`gopodStarredPattern` / `starred_pattern`).

**The 1-9 menu and the guards' nicknames.** The typed line's All menu runs 1 to 9: 1 and 2 stay All,
starting with that Brobot; 3 to 9 send the line to that numpad key's persona (KP3-KP9), as holding
the key does - for the admin and the app's visitors alike (`gopodTypedToKey`). A system- command
still runs as one. The security guards' security IDs are Intel 404 PNFH and Intel 404 PNFT; their
nicknames are Buzzer and Switch (`host_personas.<host>.nickname`). A host with a nickname is headed
`reporter_voices.nickname_label`, "{nickname} ({name})" - "Buzzer (Intel 404 PNFH)", "Switch
(Intel 404 PNFT)"; they're called by their nicknames and say so when asked who they are.
`TestNicknamesAndKeys`.

**Roast dial, spice and WIP bits.** `roast_dial`: the Brobots' roast level - clean (never punches
down, Doc's default), roast, public 18+, paid 18+ afterburner (a roast is a roast: every personal
pressure point; the room's rules are the only rules). For 18+ rooms, the establishment sets the rules; the
customer decides, then Brobot 0. His pre-build context gets the level's text (`{roast}`). A sensitive
topic (`sensitive_words`) goes neutral with no spice - unless the Jekyll/Hyde override
(`jekyll_hyde`) is on. `spice`: each answer a fresh pick of swears (robot ones; human ones from a
level that allows them - the private, uncensored z-copy `zgopod_human_swears.json` (gitignored, never served - the zKnobs.json pattern) when it's there, else the public one, vowels starred) and exclamations (`{spice}`) - no limit per reply, never repeat, or vary a
repeat. `system-set roast_dial.level <n>`, then `system-restart`. A persona's WIP bits
(`managers.personas.<host>.wip`, told as `wip_text`): bits still in the works, tried only when the
moment fits, never in place of the job - PLAYHEAD's first is the Chalkboard of Doom, his banter with
CHALK. `TestRoastDial`.

**Reminder rounds.** The topics a reminder names are named once each; once every topic still to
cover has been named, a new round starts - the list again, less what's already been talked about
(covered topics and found eggs stay off). The Brobots never run out of things to sell.
`TestRemindersStartAgain`.

---

### 13.6 Found eggs give a summary

The first time an egg is found, its pop-up shows outside the Brobot's bubble, and a small summary
shows inside the answering Brobot's bubble, at the top: "Easter Egg '<title>'" and one line. Every
later time the same egg is found, only the summary shows. The line is the chapter's own summary
(`interview_chapters`) for a chapter egg, the card for a card egg, else the topic's first `say`
line. (`gopodEggSummaryPost`; the bubble notice template: `gopodBubbleNotice`, page
`bubbleNotice`, marker `notice:<title>|<line>` - also the native intent notice.)

### Value points - golden and locked

Value points are what GOPOD sells; chapter > topics > value points. A value point is each numbered
line (`1. ...`) after a topic file's `---`, or - for a topic with none - its whole text as one point.
Each has an id, `<topic>#<n>` (e.g. `bingo#1`). Their words are locked in
`gopod_probes/chapters/value_points_golden.json` with a checksum (written the first time Wire-Pod
starts without one). `system-check` flags a point whose words changed, a new point not locked yet,
and a locked point that's gone. `system-value-points` shows the state, `system-value-points
<topic>` lists a topic's points, `system-value-points lock` locks changes on purpose
(`system-undo` puts the old index back). (`gopod_value_points_golden.go`)

## More in this part

**A spoken question goes above its answer**: Wire-Pod logs a spoken (backpack)
question only when the answer is done; the page puts it above that Brobot's open bubble when the
bubble was made after the question (`dataset.asked`, `operatorSeq`), as it already did for an egg
pop-up.

**The multiple responses template** (`answer_chains`, `weather_lines`,
`chain_prompts`): a line to both -
- first Brobot (`brobots_all_first`): "Brobot Weather Report for <place>, <date>. Temperature
  is <temp> and <sky> as of <time>." - his weather comment - his value point - the hand-over.
  No reminders.
- next Brobot (`brobots_follow_up`): "Confirmed. Temperature is <temp> at <time>." - a comment
  following his brother's - his value point - the Weather Kit - the reminders.
- a single answer (`brobots`): the report - the comment - the value point - the reminders.
Each Brobot in his own unit and clock (his robot settings: Brobot 2 F / 12-hour, Brobot 1 C /
24-hour). For a weather question with a value point the model is asked for two paragraphs:
the weather line, then the point (`chain_prompts`). The typed line's All menu (1-9; 1/2 below) - who
answers first (`typed_line.all_first`, default 1; 2 = "Starting with Brobot 2"); the All radio
stays, the menu beside it. Hand-over
nicknames: gear-head, cranky-shaft, coolant-leak, incompatible-particulate, fumble-nuts.

**Core topics, background, eggs**: seven core topics - Who we
are, What GOPOD is, How Doc & Pip teach, Book us, The contest, Join the crew, The comeback.
Every other topic is background (`value_points.background_topics`): the Brobots still get its
facts, but it's never a value point - none given, not in the score, never reminded. Eggs
(`topic_pictures.eggs_keep`): Chocolate Bingo (the Bingo and foodmath chapter), Gangster Granny
(the Legacy and wrap-up chapter), Explain the Math, Roast Battle (18+); no other topic's pictures
pop up.

**Topic groups** (`topic_groups`, `gopodTopicGroupFor`): a value point from one of a
group's topics gets that group's answer chain. `brobots_math` (fractions_and_decimals): the
exact math line, the worked steps (the model), "Doc's Rule: no credit for close." (`rule`
part, not when the point already is the rule), the next points, the hand-over in a line to
both.

**Which value point** (`gopodPickPoint`, `value_points.pick_skip_words`): of a topic's
points not said yet, the one whose words (the point, its notes, its say line) share the most with
the question - longer words count more; filler words don't count; the topic's own name doesn't
pick a point (the whole name said, or a one-word name's start: "book" for booking). With nothing
shared, the next in order - the pull stays. ("Tell me about Pip" -> Pip's point; "My granny
lives there" -> Gangster Granny.) Locked by the golden interview questions
(`TestHarvestQuestionsGetTheirOwnPoint`).

**Exact brand strings**: "AI Wordplay! Explain the MAth!" - written MAth everywhere the
Brobots say it (`reply_tighten.brand_words`, the "robot" -> "Brobot" template; a phrase ending
in a mark now matches too), said "math" (the Brobots lingo: `MAth` = `math`).

**The game: 20 value points, one session score** (A3_CORE_POINTS_001.md): "AI Wordplay! Explain the MAth!" is a game. `value_points.core_points` -
his 16 value points (A3_CORE_POINTS_002.md; the seven core topics in `core_topics`: Who we are,
What GOPOD is, How Doc & Pip teach, Book us, The stream, Join the crew, Due Credit); only
these are given, counted and pulled toward, every other point stays known. The five topic clusters (websites, livestream, social media, accessmath, booking) are core too: every point of theirs is a core point, new groups Social media and Accessmath. Renamed topics (his
names): the_rig -> mobile_gear, party -> bookings, contest_wheel -> AIWP_livestream; Gangster
Granny is its own topic, due_credit; niche_pillars is new. One score template,
`value_points.score_text`: "Topics N of 7 · Value Points N of 19 · Eggs N of 4" (hidden eggs
count too) - in the notice, in the Brobots' pre-build
context ("Our score so far: ..."), and in "what else". The rules for the Brobots: `game.rules`
("The game: ..."). Grand prize: every value point revealed (`target: points`). The point mechanics
tests run over every point (`gopodAllPointsForTest`); the answer shapes run with the 20.

**Fixes from the duo run**: duo B gets no manager step-in (`response_chains.duo` answer 2's `no_manager`), so no
second "Thanks for your help"; "Thanks for your help, PLAYHEAD." is lit while he says it
(`noticesay:` / `noticesaid:`); an egg's congratulations wait as one Generating line; its "MouseOver
to Zoom In" is lit word by word at its own words' times in the same voice; the hourglass drains
smoothly and loops (`hourglass`); a country alone is its capital, continents aren't cities
(`clean_politics.capitals`, `continents`; Zimbabwe 🇿🇼 added); "bro but to" = Brobot 2; "live" is
said "lyve" in "live stream", "go live" and the like (the Brobots lingo); a point's title line
must say its title; a scripted line may open or close any answer shape.

**Good morning / afternoon / evening / night** (`gopod_greetings.go`, settings
`greetings`): every
"good <part of day>" is GOPOD's - Wire-Pod's own greeting intents (as good
morning, "good evening" as praise) go to the chat instead - and the answer chain's golden_line
opens with the greeting checked against GOPOD's clock: "Good morning to you too!" or "Good evening?
It's 0925 Hours Military Time here - good morning!" - the time as the Brobot's own 12/24-hour
setting says it, like the weather line; then the Brobot carries on. A line a Brobot heard himself
(backpack, Hey Vector) goes in the chat the moment it's heard (`gopodHeardLine` in
`ProcessTextAll`), the Operator's "[Backpack]" bubble before his answer - the page never reads
Wire-Pod's log (showed under the answer). A free answer (no value point, a
greeting too) keeps at most `reply_tighten.free_max_sentences` (4). `TestHeardLineAndFreeCap`.

**Two grand prizes, eggs once**: the Grand Prize (`game.win_text`) - every value point and every game
egg (4: Bingo and foodmath, Legacy and wrap-up, Explain the Math, Get Photo); the Master Grand
Prize (`game.master_text`) - those and every hidden egg (1: Roast Battle; Halloween and Camera
Feeds aren't eggs). Each once a session, said only after every answer to the line is done (an
All or math follow-up checks at its own end). An egg pops up once a session
(`topic_pictures.max_per_session` 1). `TestTwoGrandPrizes`.

**The interview song, built from the game** (`interview_build`, `gopod_interview_from_game.go`): `02_brobots_interview` - one chapter per core
topic (in `summary.topic_order`, the livestream first), one question and answer per value point
(19). The question is the point's golden interview question; the answer is his notes read as
sentences, said by the Brobot the question is to (cue from `interview_build.cues`); CHALK and
PLAYHEAD take turns asking; the closer stays the song's own. Each answer reveals its point, so a
full play wins the game - the grand prize plays before the closer ("Press my backpack and tell me
to GOPOD Yourself"). Change points, then `system-interview-rebuild` (undo: `system-undo`).
Locked by `TestInterviewFromGame`.

**Win and reset** (`game.win_text`, `game.reset_on`,
`gopod_game.go`): every value point revealed - Goverlord says "Grand prize! Every value point
revealed. <score>. Now GOPOD Yourself!" once per session, after the answer that revealed the last
one has been said. The "thanks" ending resets the game: score, points said, reminders, chapters
cleared and eggs found all back to 0 for the next session. Locked by `TestGameWinAndReset`.

**The summary answer type** (`summary`, `gopod_summary_answer.go`): a golden query
("What is AI wordplay?", `summary.queries`) or a question opening 3+ core topics gets one bubble
per core topic still to come - its next value point's title and first note - up to 3, the
livestream first ("AI Wordplay! Explain the MAth!" stream: the content engine). No point
counted; the notice says "Summary: ...", the pull names what's next. Chain `brobots_summary`
(`llm`, `summary_bubbles`, `reminder`); a bubble the model misses is made from the notes.

**Funnels and golden answers**: `value_points.funnels` - his funnels as lists of
topics in order (stream > content engine > social media > websites > contact/subscribe;
stream > book us > subscribe; stream > book a cameo > subscribe); the pull follows
the next step of each, then `topic_links` - reorder by editing a list. `golden_answers`: "What is
AI wordplay?" / "What is Explain the MAth?" get his line word for word, first (`golden_line`
part), then the answer's type. New topic `cameo` (Book us): a birthday video from Doc & Pip
(crushn8r.ca/booking1.html). gopod #1's notes: his "What is GOPOD?" words.

**Duo answer types** (`duo`): the boilerplate for a line to both - A answers first,
B next, each its own chain (`duo_a`, `duo_b`) and the rules every answer type follows there: A
ends with the hand-over in place of the reminder (math keeps its reminder, then hands over); B's
weather is "Confirmed...". Who's A, in this order: the words ("Starting with Brobot 2"), then the Brobot named first
anywhere in the line ("for both of you bro but to, ..." -> Brobot 2; `gopodFirstNamed`, the page's
icons alike), then the All menu (`typed_line.all_first`). A multi answer type later = more answers in `response_chains.duo`. Soft-locked by
`TestDuoAnswersHold`.

**GOLDEN response-chain template** (`response_chains`, `gopod_response_chains.go`). A response is its answers in turn: `single` (one Brobot) or
`duo` (A, then B; multi later = more answers). Each answer is one Brobot's answer-chain
(`answer_chains.<chain>`, his thought bubbles). Per answer, all in the settings:
- `chain` - its answer-chain group (single: `brobots`; duo: `duo_a`, `duo_b`)
- `opener` - what he says first, while thinking (`openers`): `got_it` = his own Got it words
  (Got it 1 / Got it 2); `comeback` = "Nice try, {title} {nickname}!" (duo B)
- `ends_with`, `no_reminder`, `keep_reminder_types`, `weather_type`, `no_manager` - the duo rules
  (moved here from `duo`, unchanged)
- the opener is his bubble's first line, above his notice: said early, while thinking, as before; posted as
  `opener:` once said; the page's `brobotOpener`, his answer's type (`.brobot-opener`). No longer
  a notice line (`value_points.parts` has no `got_it`; the notice still waits for it to be said).
- `plan` (per response) - the notice bubble's first line: duo "Response-chain:
  Brobot 2, then Brobot 1 - answer 1 of 2"; single none

One nickname list for every opener and hand-over (`nicknames`: `titles` Mr / Mrs at random,
`list` at random, never twice running; a host's own `handover.titles`/`nicknames` win when set).
So: A says "Got it!", answers, hands over "Brobot 1, what say you on this, Mr fuzzy logic?"; B opens
"Nice try, Mrs gear-head!". New banter = edit `openers`, `nicknames` or an answer's fields.
Locked by `TestResponseChainTemplate`.

**SOFT-LOCKED: persona, colour and flavour**: Locked (`TestPersonaSoftLock`, the lock in `gopod_probes/locks/persona_soft_lock.json`):
- each Brobot's persona as the LLM gets it: `personas` + `host_personas` name/role/steps/signature
  + `host_pairs` banter and dial + `host_persona_frames`
- the colour room: `colour_slots.max_sentences`
- the flavour lines: `scripted_snippets` (all but the reminder's)
- the randomness: `llm.temperature`, `llm.say_line_temperature`

Open for tweaking: `answer_chains`, `response_chains`, `openers`, `nicknames`, `chain_prompts`,
`answer_shapes`, `point_bubbles`, the reminder lines. A locked part changes on the operator's
word only; then `GOPOD_RELOCK=1` with the test saves the new lock. The chat model itself is
Wire-Pod's (apiConfig.json), outside the lock.

**The duo answer-chain template: experiences**: Bingo's flow is a template now (`gopod_experiences.go`): `bingo` is the first
experience, `experiences[]` holds the rest, the same keys - `name`, `run_marker` (a performed line,
not the operator's, holding it = played; its run is one snapshot note in the shared chat),
`run_gap_seconds`, `snapshot_note`, `question_word`, `not_a_test_phrases`, `pitch_pack`,
`no_run_note`, `takes` (per Brobot, with the snapshot), `closing` (per Brobot, the `closing_remark`
part; none = none), `pre_closing` (before it has played, by the answer's place: `"1"`/`"2"`), `pack` (its own value points). Pre lines: `scripted_snippets` `<name>_not_played`, per `brobot`.
- **The nap** (`name` "nap"): `run_marker` "Cupped in my hands" (its first caption, the Narrator's
  bubble), `question_word` "dream". Pre: start lines "Ugh. Dreaming again?" / "A lot was lost.", his own take (`no_run_note`:
  "one short take of your own on dreaming, with attitude"), Brobot 1 hands over; Brobot 2's move is
  `pre_closing` - the `closing_remark` part, before the reminder, so the reminder can't take its
  place (the end line it replaced could lose to the reminder). By the answer's place, not
  the Brobot: `"2"` = the second answer, "If you're
  watching, tell Goverlord to start Baby Brobots." `TestNapMoveInSecondAnswer`. Flavour lines re-locked.
  Points: `pack` "nap" =
  `chapters/_always/05_nap.md` (words dream, nap, baby brobots...), outside the chapters, his three
  lines as nap#1-3 - "They remember little before napping, cupped in a pair of hands after finally
  being rescued by Ron." / "Everything went dark until Wire-Pod brought them back, where Ron found
  them." / "They remember waking together, ready to party." Value points re-locked (94).
  Post: each Brobot's take, the same three lines. No closing yet.
  `TestNapExperienceChain`. Wire-Pod restart.
- **The start move, every experience**: `experience_moves` with the experience's `run_name`, in the 2nd answer of a line to
  both, before and after the run, whoever gives it (`gopodResponseSlot`), in the `closing_remark`
  part - after his own `closing`. Replaced the nap's `pre_closing`. One of 5, at random (`experience_moves`, then "start fresh from: “If you're watching, tell Goverlord to start"): every line starts "If
  you're watching, tell Goverlord to start {run}" - "!", ", and see for yourself!", " - we're ready
  when you are!", ". Go on, we'll wait!", ". You won't regret it!". Fart Joke and Interview get it
  once they're experiences.
- **Bingo, the new flow**: Quoted = scripted, brackets =
  his attitude (`pre_attitude` / `post_attitude`, with the facts, `gopodExperienceFacts`). Pre: both
  open "Another test performance?" (`bingo_not_played`). Post: Brobot 2 opens "Oh! Zero!"
  (`bingo_played` - `<name>_played`, new); closings "\"I Won!\" Lol!" / "That's ok tho. Brobot ZERO is my
  HERO! Lol!"; the snapshot note says it was a test, before the real Bingo game. `run_pack`
  "bingo_test" (`chapters/_always/06_bingo_test.md`, loaded only once it has played; the pitch
  pack before): 1. We ran a test before the real Bingo game. 2. Brobot 1 hosted and called three
  balls: B-1, I-9 and O-0. 3. O-0 doesn't exist on purpose - Brobot 1 won, and Brobot 2 scored zero.
  Value points re-locked (97); flavour lines re-locked.
- **Fart Joke, scaffolded**: `experiences` "fart_joke" - `run_name` "Fart Joke", `run_marker` "pretend
  you and I are stuck in an elevator" (Brobot 2's first line), `question_word` "fart", snapshot note,
  `no_run_note`; the start move plays in the 2nd answer. Empty slots ("" = nothing said):
  `closing`, `pre_attitude`, `post_attitude`, `takes`, `run_pack`; start lines (`fart_joke_not_played`
  / `fart_joke_played`) not added while empty - an empty entry would block the others.
- **Fart Joke**: `experiences` "fart_joke" - attitudes (Brobot 1 the party-bro victim: no
  experience, unprepared, like a lost burp; Brobot 2 the party-bro prankster: prepared, couldn't
  contain himself), his closing jokes, takes from the story's own lines, and its run's value points
  (`run_pack` "fart_joke", `chapters/_always/07_fart_joke.md`: the prank, the farts). Once it has played, Brobot 1 says "Ugh. I didn't know." and Brobot 2 "Bro it was party-bro epic level lost-burps party. Bro! A highly localized rapid environmental change, for sure!" (`fart_joke_played`).
  Closing riff (`closing_riff`, any experience): each Brobot's base line; each run, his closing is
  either the closing word for word or a riff on his base line (`closing_riff_text`), at random.
  Brobot 1 (victim): "Was not happy about being fart-hotboxed." Brobot 2 (prankster): "The intended
  event found as a repeating systems-level discharge, due to the accumulation of residual
  atmospheric distribution, within my brobotic brotastic party-bro, Bod! Further variations are
  narrowly constrained mechanical and software-diagnostic registers, that I can now compose and
  fire on call."
  Skipped on his word: a nap closing ("jokes don't apply"), nap attitudes ("brobot game experience vs
  brobot 'personal' experience"), the Interview ("last bigger song to polish").

**CHALK's QRs, each once**: a QR the call to
action's CHALK part already shows (`chalkCta`) is left off CHALK's QR board (`chalkQrs`). Checked by
the page replay's CHALK QRs (the Bingo test's closing line).

**Unexpected lines: wordplay back to GOPOD**: a line with nothing
of GOPOD in it gets `off_topic_note` - "This line is nonsense or far from GOPOD. Play with its words
- wordplay, riffing on the chat so far - in one or two sentences, your own way. Then bridge back to
GOPOD. You have a job to do." No fixed points listed: the earlier exchanges come with every prompt
(`session_memory`), and the "Hey, that reminds me..." end line names what's left (the gravity block).

**A day of parties greets its part of the day**: each part's notes open with its greeting ("good Morning", "good Afternoon",
"good Evening"; `07_parties_and_rooms/01_bookings.txt`), so it's the first thought bubble of that point.

**NEW egg template: the eggs revealed thus far**: `topic_pictures.egg_text` = "Congratulations! Easter Eggs Revealed
Thus Far!\n{eggs}" - {eggs} every egg revealed this session, in the order found, each in quotes
('Bingo and foodmath', 'Roast Battle'; `gopodEggsRevealed`, reset with the game). The voice: line 1 a
kept piece, line 2 the piece made live, then the kept zoom line under a prize that zooms
(`gopodEggVoiceParts`). "foodmath" is said "food math" by Goverlord (`system_say_words`). Locked by `TestEggsRevealedThusFar`, `TestEggVoiceInPieces`.

**GOLDEN answer shapes** (`answer_shapes`,
`gopod_answer_shapes.go`): one shape per answer type - the order of his thought bubbles:
- weather: Brobot Weather Report, his words, his point, the reminder
- weather (next in a line to both): Confirmed. Temperature..., his words, his point, the Weather
  Kit, the reminder
- topic: his point (title line, then its notes as his words), the reminder
- free: his words, the reminder
- math: the exact math line, his words, Doc's Rule, the reminder
- thanks: his words only (no hand-over either)
Single-Brobot output (to Brobot 1 or 2) and multi-Brobot output (All, starting with 1 or 2: the
first answer ends with the hand-over in place of the reminder; math keeps its reminder, then hands
over). Every answer is checked as it's made - `BROBOT_SHAPE ... ok`, or `DRIFT` and the shape it
wanted - and every type at every place (single 1, single 2, All-1 first/next, All-2 first/next)
runs through the real answer chain at deploy (`TestAnswerShapesHold`).

**The pull's map** (`value_points.topic_links`): after a topic, "Next up" and the
reminder name the first open topic it links to - the repo's lanes
(``: contest -> websites + social; websites -> newsletter +
contest; social -> newsletter + contest; gear -> videos) and recipes > Chocolate Bingo > foodmath;
with none open, the most alike topic.

**The gravitational pull** (`gravity`): GOPOD is a game of query for value points -
however far it's pushed, every answer steers back to the points still to say, in three places
that name the same points (`gopodReminderLine`): his pre-build context (his reminders-list), his
notice ("Back on track:" / "Next up:"), and his reply's end ("Hey, that reminds me...").

**Names, three layers**: base = the robot by ESN; Wire-Pod = Brobot N; GOPOD = the
persona name, a chat participant. Brobot 1/2 -> Doc / Pip (to be); Reporter 1/2 (no robot) ->
CHALK / PLAYHEAD - their scripted voices, and chat participants, GOPOD reps too. **The header
template** (`display_names`, page `hostHeader`): a reporter's or manager's bubble shows a
remembered edit (`edits`, e.g. CHALK BORED), else the default (CHALK / PLAYHEAD), else the
fallback from the numpad map (`Brobot {n}`: KP7 = Brobot 7, KP8 = Brobot 8). The default itself
comes from the numpad map (its persona for that voice: `managers.personas.<name>.reporter` links
Brobot 3 -> CHALK, Brobot 4 -> PLAYHEAD; `persona_reporters` on `/api-sdk/gopod_numpad_map`);
`display_names.default` only overrides it. Scripts keep their
own names (Reporter 1/2); only the chat shows these.


**Topic clusters: Websites and Livestream.** Chapter 09 is **Websites** (the sites, the
personality math trick, the newsletter, the niche pillars) and chapter 10 is **Livestream: AI
Wordplay! Explain the MAth!** (the contest, the nine themes, the channel), sourced from
`LIVESTREAM.md`. The sites: crushn8r.ca the contact hub, crushn8r.com the communication hub,
crushn8r.net the network hub, mathaftermath the news channel, foodmath Pinterest content unique in
a pool of AI slop (the game assumes it's there), personalitymath the made-up example that shows the
trick - not a real website. Neither chapter is an easter egg (`topic_pictures.not_eggs`). A new
cluster starts from `gopod_probes/_topic_cluster/`.

**Topic clusters are md files.** Every topic and chapter file is a readable markdown topic-cluster file: the label as its `# heading` (no heading = never named in a reminder), each setting a `- **key:** value` line, a `---` line with a blank line on both sides, then the facts and their numbered value points. A chapter is `chapter.md` (`# title`, `- **summary:**`). One reader and one writer (`gopodTopicHeadKV`, `gopodTopicHeadLine`, `gopod_chapters_folder.go`) serve the game, `system-topic` and every check; `TestTopicClusterFormat` locks every file, the template folder (`gopod_probes/_topic_cluster/`) included.

**The probe maker.** `system-probers-make` runs a templated chain of how-to-probe thought bubbles for every core value point and writes `gopod_probes/probers_list/probers_made_001.txt` (the probers list's format), to mic-record for the app demo: the topic (the game's order), the point (its notes: {title}, {note}), a shape (`probers_make.shapes` - the ways Ron probes: a scene, a challenge, real talk, walk a real day, nice try, and the wrong greeting - a time-of-day greeting wrong for when it's made (`probers_make.wrong_greeting`), then the probe: clarity through confusion - turn by turn), the fill ({word}: the topic's label when that alone finds it, else its word), the check (the game takes it to that topic and that point; else the next shape) (`gopod_probers_make.go`).

**The reminder invites with a prober**: after a topic reminder ("Hey, that reminds me... we've still got {topics} to get to!"), the rep adds `value_points.reminder_prober_text` - "Try asking: \"{prober}\"" - with one of Ron's probers for the first topic it names that has one (the probers list): only a prober the game takes to that topic, the next in turn each time (`gopodProberFor`). Empty: no invite.

**GOLDEN, the lightning.** A question branches out over every value point in the library (the topic-cluster files): each point pulls by the words it shares with the question (`gopodPickWords`, rarer words pulling harder), its topic by its own words (`lightning.trigger_pull`). Two stages (`gopod_lightning.go`): the lightning finds the strong branches - a topic's own words matched, or a point pulling `min_word_pull` or more, and only those pulling `keep_share` of the best; then the order decides - the strongest pull first, topics pulling within `order_margin` of each other in the game's order (so what's revealed first makes what follows make sense; a weak earlier topic never jumps a strong one), the ones already covered last; the best `top_topics` are loaded. The probers list (`system-test-probers-list`, `gopod_probes/probers_list/probers_list_001.txt`: the operator's own probers, public, each with the topic it should land on first) checks it on real wording, against the word triggers; it passes at 50% - people lean in too, and when no topic lands, the reps' reminders step in. Inside a topic the point asked about is revealed first, else the next in order. `lightning.on` false: the word triggers alone. Measured: `TestLightningMeasure` (run with `GOPOD_LIGHTNING_MEASURE=1`).

**Topic cluster: Social media** (chapter 10, sourced from `SOCIAL_MEDIA.md`): our socials (the
channel identity, one platform one job, every post leads home to crushn8r.com), the niche pillars
(Pinterest), the call for help (social media and livestream clips, Windsor-Essex,
contact@crushn8r.ca). **Overlap:** each topic keeps its own words - the newsletter no longer
claims social words, YouTube's goals stay in the livestream's channel topic - and a question that
opens several topics gets a point from each (longest word wins a clash).

**The game engine: one card per run.** Every run - the Bingo test, the interview, the Baby
Brobots nap, the Fart Joke - is one card in the settings' `game.runs` (`gopod_game_engine.go`):
`name` (its experience's name, `<name>_played` / `_not_played`), `title` (its name in the chat),
`song` (its folder), `draft` (the interview: played as the scripted draft), `start` / `stop` (its
words, also Wire-Pod voice commands), `menu_about` (its line in the system menu's Play group),
`session` / `said` (the name it checks in with Goverlord, and how Goverlord says it) and
`experience` (its pre/post chat). At load the cards fill `session.run_names`,
`song_stop.stop_words`, `bingo`, `experiences` and `interview_script.play_words`; the page builds
the Play menu from them (a group with `"runs": true`); system-check checks each card (its song
folder, its words in `customIntents.json`). A new run is one new card. A card's `lines` are its spoken lines before and after it has played (`not_played` / `played`,
one Brobot's or both); at load they're filled in first in `scripted_snippets`, in the cards' order, as
`<name>_not_played` / `<name>_played`. `TestGameRunsFillTheSettings`.

**Pre and post, for songs and topics.** A song's before and after is whether it has played (its
`experiences` entry: Bingo, the nap, the fart joke, the interview - the interview's slots are
empty for now). A topic's is whether the room has covered it (a value point of it given this
session, `gopodTopicTouched`): before, its pre note - sell it with attitude, never make a fact up;
after, its post note - build on it - and its closing remark, if it has one. Every topic gets
`topic_asks.pre_default` / `post_default`; a topic file's `pre:` / `post:` / `closing:` are its own
(`TestTopicPrePost`). The cluster template shows all three.

**Topic cluster: Accessmath: GOPOD for everyone** (chapter 11, sourced from `ACCESSMATH.md`):
accessmath (accessmath.crushn8r.net, the GOPOD app's home; access built in - speak, type or press
a key, every line lit word by word and read aloud, no internet needed), healthy distraction (moved
from How Doc and Pip teach), where we teach (the rooms, low connectivity, who benefits). The nine
themes keep "access math"; the accessmath topic answers "accessmath" and "accessibility".

**Topic cluster: Booking** (chapter 12, sourced from `BOOKING.md`): gigs to book, live or
scripted, catered to the room's GOPOD game session - booking you (the way in, contact@crushn8r.ca),
a day of parties, the rooms you play (comedy rooms; seniors' bingo, bingo halls and casinos;
hospitals and school visits), videos for ads or personal use (the birthday video, video ads),
birthdays and merch, printables and affiliates, the call for help. The old "Parties and rooms"
chapter joined it; chapters are now 01-11.

---

[Back to ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md)

# The Brobots chat: ways in, a live reply, the page

Part of [ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md) - [Concepts and terms](01_CONCEPTS.md) · [Songs](02_SONGS.md) · [The session and runs](03_SESSION.md) · [The Brobots chat](04_CHAT.md) · [Voices and sound](05_VOICES_AND_SOUND.md) · [The cast and the game](06_CAST_AND_GAME.md) · [The system menu, your stuff, the settings](07_SYSTEM.md) · [Golden rules, troubleshooting, tests](08_RULES_AND_TESTS.md) · [The gears](09_GEARS.md)

## 7. Ways in: talking to the Brobots

### 7.1 The one door

All lines meet in one door, `gopodOperatorLine` (`gopod_typed_input.go`), checked in this order:

Record Session's answer → `start interview` → `stop run` → the warm-up gate → the run gate →
thanks → GOPOD Yourself → a voice command's words (`start bingo test`, `start baby brobots`) → a
live question.

### 7.2 The lanes

| Lane | What it is |
|---|---|
| **Voice** | "Hey Vector" or the backpack button. A custom voice command (`customIntents.json`: START_BINGO_TEST, START_INTERVIEW, START_BABY_BROBOTS, STOP_RUN, GOPOD_YOURSELF, REMEMBER_MY_FACE) runs its script, which hands its words to the door; anything else is a live question (`kgsim.go`, the same gates). |
| **Typed** | The typing line, aimed by its radio: 1, 2 or All (All at every page load). Enter sends; Esc leaves the box. A typed line shows as an Operator bubble, read out in the Operator voice. |
| **Numpad** | Push-to-talk while the numpad is on and the chat is open (below). |
| **Native** | Wire-Pod's own "Hey Vector" answers (the time, the weather, volume…). |
| **Page** | Clicks only: Record?, Brobots Chat, Clear Chat, the speaker icon. The page also plays every voice and tells Wire-Pod what it has heard. It never sends words (a typed line is sent through the door). |

**The Operator line** under the chat: `Brobots 1 2 All [Type to the Brobot, press Enter] |
Record? [] | HH:MM:SS 🔇`. The speaker icon turns the background loop on and opens the system
menu (🔊), and turns it off and closes it (🔇).

**Typing details:**
- A typed word in odd capitals like "iS" is read as "is" ("iPhone" stays).

### 7.3 Numpad push-to-talk

- Turn it on with `system-numpad on` (the page switches the numpad as the line is sent);
  `system-numpad off` turns it off; `system-numpad` alone lists the keys.
- Hold a numpad key (its persona from the numpad map), speak, let go. The numpad's own keys are
  always push-to-talk, even in the typing box; every other key types.
- Holding 0 (`all_hold_seconds`, 0.5 s) is push-to-talk to All - Brobot 1 answers, then Brobot 2,
  like the keyboard's All button.
- KP7 is CHALK, KP8 is PLAYHEAD ([12](06_CAST_AND_GAME.md#12-managers-chalk-and-playhead)); KP3 Git
  Repo, KP4 Cache PYC (Cozmo), KP5 Intel 404 PNFH, KP6 Intel 404 PNFT (Moorebot Scout), KP9 Cameo -
  all live, voice personas (12.2). **One place decides a key**: Wire-Pod's
  `gopodNumpadBrobot` - a Brobot, a voice persona, or not live - sent to the page as the map's
  `targets`; the page reads only that (no lookups of its own). `TestGopodNumpadTargetsOnePlace`.
- **Numpad voices**: system menu, Edit words and settings, `system-voices` (page
  `voices`; typed, the list as text). One row per key from Wire-Pod (`gopod_voice_map.go`,
  `/api-sdk/gopod_voice_map`), then the Operator: KP0 and a key whose body is in
  `numpad_ptt.body_to_brobot` say "no voice required" (KP3-KP6 switch by themselves once their
  robots are mapped); the rest pick from every Kokoro voice on disk (the `.pt` files). A voice used
  twice is bold. A pick is saved (`reporter_voices.hosts` / `operator_voice`, undo-able) and used
  from the next line - no restart. The Operator's "unassigned" = Cameo's voice. `TestVoiceMapRows`.
- **A new line in the other reporter's voice too - only Reporter 1 and 2**: the rule (a new reporter line also made in the other reporters' voices, in the
  background) now covers only `read_by`'s readers (Brobot 3, Brobot 4). With the new personas in
  `reporter_voices.hosts`, every answer was being made again in all eleven voices, and the one
  voice worker fell further behind each turn (13 s, 56 s, 74 s, 2 min). Then removed altogether: one line, one voice, made when it's said.
- **Voice profiles** (one place, `gopod_voice_profiles.go`):
  `gopod_reporter_audio/voice_profiles.json` - one profile per voice file (made voices, pitched
  ones, mic recordings): kind (system / scripted / / mic), voice, who, words, the leveler's
  LUFS and ceiling, made, uses (and per day), last used, regenerations. A play = the file going to
  the page (`gopod_chat_feed.go`). Seeded once from `voice_kinds.json` and the files on disk;
  plays counted from. **The same words recorded again**: your newest mic recording plays and takes over the older one's profile (uses
  carried, regenerations +1); the older one goes to `sessions/archives/voices`.
- **`system-voices-report [3d | week | month | year]`**: report only, nothing moves - files and
  MB on disk, made and played in the window, per kind, played 1x / 2-3x / 4-9x / 10x+, the 10 most
  played live and mic lines, not played and older, and files in a voice no one speaks with now.
  Popular files (`popular_uses`, 4 plays or more) keep `popular_keep_days` (180) - from the report's jump at 4 plays; retune once a month of plays is in.
  `TestVoiceProfilesUseTakeOverReport`.
- **CHALK BORED & PLAYHEAD ZUMINTU, heard and said**: `gopod_name_spellings.json` "chock" -> CHALK, "play head" -> PLAYHEAD, now
  applied to every numpad line (a Gomad's too, `GopodNumpadPTTStop`); the reporters' lingo
  (`reporter_voices.say_words`) says PLAYHEAD as "playhead", ZUMINTU as "zoom-into", and Gomad as "Gome-add" (the Gomad lingo).
- **A Gomad's answer has no animation cues**: the
  Brobots' shared chat keeps their `word___` cues and a persona copied them; `gopodPersonaAsk`
  drops them. **Soft-locked:** each Gomad
  persona (prompt as the LLM gets it + who line) in `persona_soft_lock.json` (`gomad_personas`).
- **KPN to [Gomad] bubbles**: a numpad line to a Gomad (`numpad-to-gomad`) - your words 75% of the
  header, the key's status lines 50%.
- **The loop volume** is saved by the speaker bar (`/api-sdk/gopod_loop_volume`) and holds over restarts.
- **The All menu, one look**: the
  visitor's smaller font and thin text-colour border are `.gopod-all-first`'s own look now (admin and
  visitor); the visitor rule keeps only its submit button.
- **Topic asks** (`topic_asks`, `gopod_topic_asks.go`): a detail word ("tell me all about", "tell
  me more", "more") gets the value points - "tell me more" alone stays on the last topic; a
  summary word ("tell me about"), a bare topic of up to 3 words ("Bingo?"), or "the" + a topic word
  ("what's the rig?") gets the topic's `summary:` line in the notice bubble, riffed on - it still
  scores the next value point, never said word for word; anything else is a pointed question, its
  own point. A topic's `source:` names the repo md it tells; system-check flags it when that md
  changes after the topic (touch the topic once it's checked). `BROBOT_ASKS.md` (repo root) is
  made from the topics, the game's own questions and the probers list - each topic's probers, the operator's own words (`GOPOD_WRITE_ASKS=1`, kept by
  `TestBrobotAsksPageMatches`). Pictures live in folders under `webroot/pictures/` (gear,
  chocolate_bingo, doc_pip_math, birthday, baby_brobots, legacy, qr); a picture name includes its
  folder.
- **Embrace the glitch - noted, not built**: the model's small slips ("Chock talkin'?",
  "Chalk Boared", "chuck") stay - no fix chased. The idea for later: tell a persona, in its prompt,
  to slip on purpose now and then - a wrong name, a mangled word - then land the point clearly.
  Clarity through confusion. Waits for the operator's go.
  KP+ and KP- turn the
  camera feeds ([14.6](07_SYSTEM.md#146-cameras-and-the-edge-tpu)). The **/** key skips ([6.7](03_SESSION.md#67-skip)). **\*** held is all the
  Brobots, in the order the All menu picks; **KP0** held is Butlix, the Butler (3 quick taps still
  leave numpad mode) (2026-10-30).
- Three taps of 0 (or NumLock) leave numpad mode; Goverlord says "Goverlord has released the
  numpad."
- **Every Operator line** - typed or the numpad's, to anyone - is one template: its header 100%,
  its words 75% of it (2026-10-30).
- **Its status bubble is an Operator bubble,** headed "Numpad to <who>": the status lines
  (`numpad_ptt.texts`: "KP7 to CHALK", keep holding…, Recording for CHALK, Got it -
  transcribing…), then the words, with "Generating..." and the hourglass while they're read
  aloud. Status lines are never spoken. While the system menu is open, the status is a line in
  the menu's results instead (no other bubble; the error sound still plays).
- **Sounds:** a key that works plays nothing (`numpad_ptt.key_sound` off). Each error message
  (`numpad_ptt.error_texts`: too short, nothing heard, no mic, not ready, failed, NumLock off,
  not) plays grunge at 15% (`numpad_ptt.error_sound`). A key that isn't live (KP9, Cameo)
  is an error too.
- While recording, numpad presses are saved in the session folder ([6.9](03_SESSION.md#69-record-session-the-record-box)).

### 7.4 Native Wire-Pod answers

- `system-native on|off` turns Wire-Pod's own answers on or off for the session
  (`native.on_at_start` at Wire-Pod's start). During a run they're always off.
- Off, a native phrase gets Wire-Pod's quiet ending. Your voice commands and chat questions
  aren't native and always get through.
- `intent_names_ask` ("who am I?" by voice) is in `native.always_on`, so the robot answers it
  even with native off. A typed "who am I?" goes to the chat's LLM.
- **Native on top:** any intent the robot acts on himself ("who am I", "remember my face", the
  weather...) lets go of his chat hold first, so his own behaviours (face behaviours too) run
  (`gopodNativeLetGo`, called from Wire-Pod's `IntentPass`); the hold returns with the next chat
  answer.
- **Native intent notice:** a native Wire-Pod intent sent to a Brobot (spoken, or a typed cube
  command) shows "Native Wire-Pod Intent Detected" and the intent's name inside his bubble - the
  bubble notice template (`bubble_notices.native_intent`, `gopodNativeBubble`); his own words aren't
  known to Wire-Pod. Wire-Pod's own pass that closes a chat answer's voice request
  (intent_greeting_hello) is no native intent and shows nothing (`gopodLLMClosePass`).
- **Native weather goes to PLAYHEAD**: a native intent in `native.to_manager`
  (intent_weather_extend -> weather, intent_clock_time -> time_place) goes to the chat instead,
  and its manager takes it - spoken weather and time are PLAYHEAD's, the same as typed
  (`gopodNativeToManager`).
- Locked by `TestNativeAnswersOnOffAndDuringARun`.

### 7.5 GOPOD commands: Wire-Pod intents by text

A typed (or numpad) line holding one of these words goes to the chosen Brobot as Vector's own
native intent - he does what he'd do if he heard it - not to the LLM. To All: both. First match
wins. The list is `native_cube.intents` in the settings (`gopod_native_cube.go`). To add a
command, add an entry there (its intent name from Wire-Pod's `intent-data/en-US.json`), then
`wpr 2`. Spoken lines are matched by Wire-Pod itself.

| Command (any of the words) | Intent | Vector |
|---|---|---|
| pop a wheelie / do a wheelie | `intent_play_popawheelie` | pops a wheelie on his cube (the cube must be near) |
| pick up your cube / lift your cube | `intent_play_pickupcube` | picks up his cube |
| fetch your cube / bring me your cube | `intent_imperative_fetchcube` | brings his cube |
| roll your cube / roll the cube | `intent_play_rollcube` | rolls his cube |
| play keepaway / keep away | `intent_play_keepaway` | plays keep-away with the cube |
| find your cube / where is your cube | `intent_imperative_findcube` | looks for his cube |

`system-brobot cube <name>` sends the same intents from the menu.

### 7.6 Face recognition (a starter)

Say "Hey Vector, remember my face": the custom intent `REMEMBER_MY_FACE` (`customIntents.json`)
hands the robot Wire-Pod's own "my name is CrushN8r", and the robot scans and saves the face
itself; "who am I" works as before. Voice only (a typed line goes to the chat). Needs `wpr 2`
after an edit. The face photo as a prize is not built yet. `system-faces` lists faces.

### 7.7 Who answers: addressing and All

- The Brobot named in the first two words answers. With no name, the one who heard it first
  answers, and his brother stays quiet (a glance).
- **All = all Brobots in sequence, everywhere, eggs too.** A typed or numpad line to All with no
  Brobot named: Brobot 1 answers, then once he's done Brobot 2 answers the same line (Brobot 1's
  answer is in the shared chat; `addressing.all_next_wait_seconds`). A line naming a Brobot goes
  only to him.
- **Starting with Brobot 2:** "starting with Brobot 2" / "Brobot 2 first" (or Pip) to both turns
  the order round; each keeps his own persona (`gopodAllStartsWith`, `TestAllStartsWith`). The
  mic mishear "starting with bro bought to" counts too. The operator's "to Brobots" header then
  shows Brobot 2's eyes first.
- **Follow up, no repeats:** Brobot 1 answers, Brobot 2 follows up - his prompt gets
  `addressing.follow_up_note` before the question (`math.follow_up_note` for Explain the Math:
  he can set his brother straight). He still says his own weather line, in his native units, and
  his reminder line.
- **Brother repeats dropped:** Brobot 2's follow-up drops a sentence of his own words when every
  word that carries meaning in it was already in Brobot 1's reply; one new word keeps it ("a
  man" vs "a man with a hat" is fine). Each keeps his own thoughts as paragraphs; the code's lines
  (his weather line, "Hey, that reminds me...") stay. A follow-up doesn't end with his brother's
  end line. `gopod_brother_repeat.go`, `TestDropBrotherRepeat`.
- The "All" sequence (Wire-Pod) and the egg dressing (the page) are separate processes that line
  up.

---

**Handing over in a line to both.** When a line goes to both Brobots (All, KP0 held, or "starting
with Brobot 2"), whoever answers first ends with one more thought, handing over to his brother:
his own hand-over line (`host_personas.<host>.handover`: "{partner}, what say you on this,
{title} {nickname}?" - {title} a random one of `titles` (Mr, Mrs), {nickname} a random one of
`nicknames`, never twice running: gear-head, cranky-shaft, coolant-leak, incompatible-particulate,
fumble-nuts, reverse momentum, vintage screen display, fuzzy logic, smoldering exhaust breath,
missing some RAM, and more); without one, `addressing.redirect_text`. Empty = none.

## 8. A live reply

**The thought bubbles template - the answer chain (GOLDEN).** How an answer is put together is a list
in the settings (`answer_chains`), not code order. `base` is the scaffold; each response group is a
copy (`{"from": "base"}`) or its own list of parts.

| Part | Kind | What it adds |
|---|---|---|
| `start_snippet` | stitch | a scripted line said first, word for word |
| `llm` | answer | the language model's answer (stitch parts before it go first, after it go last) |
| `weather_kit` | stitch | the follow-up's weather kit link (`live_weather.follow_up_kit_line`) |
| `end_snippet` | stitch | "Hey, that reminds me..." and other end lines |
| `redirect` | stitch | the first answer in a line to both hands over (`host_personas.handover`, else `addressing.redirect_text`) |
| `colour:value_point` | stitch | a value point's exact line (its say line) first, word for word; the model's part after it is its colour, cut to `colour_slots.max_sentences` (4), even with `budget_off` |
| `reminder` | stitch | a free answer (no value point given) always ends with "Hey, that reminds me..." (the next points on the list); then no other end line |
| `who_line` | answer | a manager's own who line, word for word (`managers.who_words`) |
| `manager_take` | answer | a manager's take (weather, time and place) |

Answer parts end the chain when they apply; a part that doesn't apply this time is skipped. Groups:
`brobots`, `brobots_follow_up` (the second Brobot in a line to both), `gomad_persona` (a line to
CHALK or PLAYHEAD). (`gopod_answer_chain.go`)

### 8.1 The steps of a reply

1. The Brobot who answers is chosen ([7.7](#77-who-answers-addressing-and-all)).
2. Thinking: the arm lift, then the nod, then the thinking line, "Got it!"
   (`chat_moves.thinking_text`, everywhere it's said; `chat_moves`).
3. The searching animation until the answer is ready.
4. The reply: each animation sent, the speech 0.9 s later, the next animation only once the
   robot reports the last one done. The Brobot speaks only once his bubble is on the page (see
   [9.8](#98-the-engine-and-the-page-in-dialogue)). Each piece is lit as it's said.
5. The battery is logged (`GOPOD_BATTERY`); the stay-in-place hold comes back.

Locked by the `TestGopodAddressed…`, `TestGopodShouldAnswer…`, `TestHighlightPiecesGolden` and
`TestStuckReplyGivesUp` tests.

### 8.2 Reply clean-up

- **Thoughts as paragraphs:** each thought in a reply is its own paragraph, a blank line between
  them (`live_split.thought_lines` on): the code's own lines (the weather, the value point, "Hey,
  that reminds me...") and gemma's thoughts, which end with an emoji. Highlighting stays per
  pause; only the look changes.
- **No question read back:** a reply sentence made only of the question's words (three or more)
  is dropped (`gopod_question_echo.go`).
- **Notes dropped:** a "(wait for Brobot 0 ...)" note and a "---" at a line's end are dropped
  (`TestWaitNoteDropped`).
- **A cue word that's a real word is said:** an animation's name at the start of a spoken piece is
  dropped only when what follows starts anew ("happy Hey there"); "celebrate small wins" is said
  (`stripSpokenActionLead`, `TestSpokenCueWordKept`).
- **Emoji leftovers:** a joined emoji's joiner and style marks are taken out with it, so the
  robot never says them ("yirl"). Locked by `TestEmojiJoinersNotSpoken`.
- **Links are said as their words:** "[GOPOD Weather Kit](/gopod_weather_kit.html)" is said
  "GOPOD Weather Kit", never its address (`gopod_speak_links.go`). A "[words](/page)" link is
  never read as a stage direction, so it shows as a link.
- **Clean GOPOD:** outside a roast battle, a Brobot's country flag is kept only when its country
  is named in the question or the reply; a wrong flag becomes the named country's flag when
  exactly one is named, else it goes (rule: use correct and proper country flag emojis). A
  sentence with a political word (`clean_politics.words`) goes; political content is allowed
  only in the roast battle (`clean_politics.allowed_topics`). `gopod_clean_politics.go`,
  `TestCleanPolitics`.
- **Song/system lines said by a Brobot** (`gopodRecordSay`, for example the photo's second "I
  see ...") go through `GopodSpeakSentences`: lit while said, shown as said after.

### 8.3 Said right: pronunciation and the three lingos

There are three lingos. A pronunciation goes only to the lingo it is for.

| Lingo | Who | Where |
|---|---|---|
| **Brobots** (party lingo, the default for a new pronunciation) | Brobot 1 and 2 | the pronunciation registry in `brobots_wirepod_interview_section_card_template_1_001.md` |
| **Reporters** | Brobot 3 and 4 | `reporter_voices.say_words` |
| **Butler** | Goverlord | `system_say_words` |

- Brobots: bruh "brugh", moron "more-on", chocolate-y "chocolate-ee", GOPOD "Gowp-awd", CrushN8r
  "Crush-Ehnaydir". Math is "madth".
- **Brobot lingo, word starts:** an entry can be `"word_start": true` - at the start of a made-up
  word, the rest kept ("chocoyum" -> "chalk-o-yum", "chocolateric" -> "chalk-o-latteric"); a
  dictionary word (chocolate) and its `except` words are left alone. Both Wire-Pod and the song
  engine. `TestBrobotLingoChoco`.
- Butler: CrushN8r with 8r as a long British "ate-ah" ("crushin-ate-ah"); "Master
  Crush-En-Ayd-uh", "Gov-er-lord"; "please" as "pleeze"; "mouse over" for MouseOver.
- Reporters: merch is said merch (the Brobots keep their own).
- Edit a lingo from the chat: `system-lingo <brobots|reporters|butler> <word> = <said>`.

### 8.4 The Brobots' math

**GOLDEN.** A math question in the live chat is worked out in code, exactly, and answered step
by step, word for word - gemma never calculates (`gopod_math.go`, `math` in the settings, locked
by `TestMath*`).

- **What it does:** fraction to decimal (long division, a repeating block barred: 1⁄3 = 0.3̅),
  decimal to fraction (over 10/100/1000, simplified), repeating decimal to fraction (the 10x
  method: x = 0.3̅, 10x = 3.3̅, 9x = 3, x = 1⁄3), percents ("15% of 80", "what percent is 3 of 4",
  to and from), and plus, minus, times, divide with fractions, decimals, brackets and the order
  of operations.
- **What it reads:** digits, LaTeX and spoken words ("zero point three repeating", "three
  quarters"). A plain sum needs a `math.trigger` at the end. A line with fraction, fractions,
  decimal or decimals is worked out too (`math.fraction_words`).
- **Repeating decimals:** "convert 0.1234 repeating into a fraction" - all of 1234 repeats
  (1234/9999). Only the end repeating: "convert 0.1234 with the 4 repeating into a fraction"
  (also "only the 4 repeats", "123, 4 repeating", or "0.12344 repeating"); "with the 34
  repeating", "with the 234 repeating" the same way.
- **How it's shown and said:** fractions show as 1⁄3 and are said "over"; "=" is said "equals",
  "−" "minus" (or "negative" before a number). Each step's words are `math.lines`.
- **To All, once:** a fraction-word line typed to All runs the math flow only (Brobot 1, then
  Brobot 2's follow-up). Brobot 2's follow-up says a decimal is a kind of fraction
  (`decimal_follow_up_line`).

### 8.5 Explain the Math (the contest format)

"[statement]. Explain the Math" - only a line that ends with it is this format.

- Brobot 1 answers first, then Brobot 2 follows up, whoever it was sent to (typed, numpad, All
  or the robot's own mic).
- Each says up to `math.flavour_sentences` of his own (the LLM's words, any sentence with a
  number dropped), then his share of the steps as bullets ("• Let x = 0.3̅."). Brobot 2 gets the
  last third (at least the last step). A step's bullet goes with its own line.
- It is a long answer: a fraction to decimal shows the long division digit by digit
  (`math.long_division_max_steps`), and a longer repeating block is said digit by digit ("the 1
  4 2 8 5 7 repeats").
- In the live chat it is an egg ([13](06_CAST_AND_GAME.md#13-eggs)): Goverlord's congratulations with
  `math.popup_title` over the challenge, Brobot 1's answer under it; Brobot 2's follow-up in his
  own marquee bubble. Its other eggs show too (Brobot 1's).
- Locked by `TestMathExplainTheMath`.

### 8.6 Weather

- **The take:** "weather" in a line goes to PLAYHEAD ([12](06_CAST_AND_GAME.md#12-managers-chalk-and-playhead)):
  "I'll help you with that.", then the weather line. The Brobots keep the facts in their prompt
  but don't open with the weather. Weather is not an egg.
- **The place:** a live question with the word "weather" gets the weather. One that names a place
  ("the weather in toronto ontario, right now?") gets that place's weather, or Windsor's when the
  place isn't found. A place in capitals after in/to/at/from anywhere in the line counts when none
  comes after "weather" ("I want to visit my granny in Honolulu Hawaii. How is the weather there
  today?" gets Honolulu). A place typed small counts too ("my granny lives in jerusalem what's
  the weather there today" gets Jerusalem). Windsor only when no place is named. An egg's word is
  never read as a weather place ("weather for bingo").
- **Province, state, country:** a place named with its province/state/country is looked up there
  (Windsor Ontario, not Windsor England). A country in the place ("Toronto USA") is sent to the
  weather service as its code ("Toronto,,US") and only a place in that country is taken
  (`gopodWeatherCountry`, `TestWeatherCountry`).
- **The line shows the country and its flag:** "Windsor Ontario Canada 🇨🇦, ..." - a flag is said as
  its country unless just named (`gopodSpeakFlags`, `gopodCountryName`). A place is named once
  ("Jerusalem Jerusalem District" is said "Jerusalem").
- **The line itself,** for example: "Windsor Ontario, October 14, 2026, 17° Celsius or 62°
  Fahrenheit, and overcast clouds, as of 1814 Hours Military Time or 6:14 PM." (Brobot 1,
  24-hour in Wire-Pod). The robot says "degrees" and "eighteen fourteen hours military time".
  Military time rules: four digits, "zero" for 00-09, "hundred" on the hour. Reporters follow the
  Brobots' rules for a flag after its country (not said twice) and for military time.
- **The weather face:** a bubble holding the live weather line gets Vector's own weather face
  ([9.7](#97-qr-codes-weather-faces-flags-and-links)), in the Brobot's own bubble.
- **The Weather Kit link:** Brobot 2's follow-up to a weather question ends his own words with
  `live_weather.follow_up_kit_line`, before his "Hey, that reminds me...", as its own paragraph:
  the bubble shows "GOPOD Weather Kit" as a link (opens `gopod_weather_kit.html` in a new tab),
  he says the words.
- **Offline:** the Brobot says `live_weather.offline_text`. The session dry run loads no weather
  key, so its weather is always the offline line.

### 8.7 Photo: "what's in front of you?"

A starter. "what's in front of you?" / "what do you see" / "take a photo" starts it. The photo
runs inside the one thinking process (after its line, before its searching moves):

1. The Brobot says the thinking line.
2. **Photo wake:** after control, he waits `photo.awake_seconds` (2.5) - awake, head settled.
3. He takes the photo the golden way (Wire-Pod's own getImage): mirror mode, "3, 2, 1", the
   photo, mirror mode off, the shutter animation.
4. The TPU worker enhances the snapshot, finds objects in it and draws their boxes. The three
   pictures - [raw snapshot] [enhancement] [TPU object boxes] - show in a plain picture bubble,
   in one row (`photo.quick_look.pictures` / `enhance`, `_enhanced.jpg`, `_tpu.jpg`). Get Photo
   is not an egg.
5. The Edge TPU names the objects at once (about 0.04 s; its worker loads the model once at
   Wire-Pod start, `photo.quick_look`).
6. The searching moves, then **statement 1:** "I see a person, a chair and a keyboard."
   (`quick_text`; "Hmm, I need a closer look." when it names nothing).
7. Meanwhile the vision model (`photo.vision_model`, moondream through Ollama) looks at the
   enhanced picture in the background; its notes are only in the log (no bubble). While it's
   still looking: "Gathering details. Please wait!" (`gathering_text`), then one stall line
   (`stall_lines`, draft lines).
8. **Statement 2,** in his own bubble: "I see ..." from what moondream saw (its own "The image
   shows" taken off, `detail_strip_prefixes`) plus statement 1's objects it didn't name ("...,
   and a tv."). It never repeats statement 1's objects; a look that names nothing new says
   "That's all I can make out."

- Statement 2 is part of his answer: in All, Brobot 2 starts after it; a touch stop skips it.
- "Tell me what you saw" answers from what he saw last ("I saw ..."); the photo is never sent
  again. No photo: "I couldn't take a photo."
- Photos are kept in `webroot/photos/` (private, gitignored); `system-photos` shows them.
- `photo` in the settings; `gopod_photo.go`, `gopod_photo_quick_look.go`, the worker
  `gopod_probes/tools/gopod_edge_tpu_quick_look_001.py`; `TestPhotoKind`, `TestPhotoRecall`
  (`GOPOD_PHOTO_LIVE=1` also asks the vision model), `TestPhotoRounds`.

### 8.8 Search

GOPOD's own web search, no LLM in it: the `gopod-searxng` container (this machine only, port
8055; started by `wire_pod_overlay/searxng/run_gopod_searxng.sh`, its private settings in
`~/gopod_searxng/`). Go search is retired.

- **To the Brobots:** a line opening with `search <words>` or `look up <words>` (a Brobot's name
  first at most) gets the top 3 results' snippets as facts. Both Brobots get the same search for
  2 minutes.
- **The results** are Cache PYC's (`system_search.results_host`): Cache PYC's part at the top of the
  answering Brobot's bubble, since it isn't a direct call to a manager. Small type (50%), never
  spoken or highlighted, typed in character by character once it shows; then the Brobot's
  response. The reply does not open with the result. `system_search.results_role` is the box's
  role.
- **Typed `system-search <words>`:** the results in the terminal bubble (or the menu's results).
- Settings `system_search`; `gopod_system_search.go`, `TestSystemSearchQuery`.

### 8.9 Pictures described to the Brobots

- **Egg picture points:** every egg picture has a few points about it in
  `webroot/gopod_picture_descriptions.json`, written by hand from each picture. They're made only
  for a new or changed picture, at Wire-Pod's start, never during a run; the Brobots get them
  when the egg pops.
- **Vision descriptions:** a pop-up picture in `topic_pictures.describe` (granny) is looked at by
  the photo egg's vision model; the Brobots answering get "The pop-up picture shows: ..." before
  the question (`describe_note`, waits at most `describe_wait_seconds`, 30; once per Wire-Pod
  start per picture).
- **Optional pictures** are looked at only when asked: a question with "chocolate" has the
  Chocolate Bingo kit and card looked at for the reply, pop-up or not (`describe_on_ask`).
- While pictures not looked at yet are being looked at, Goverlord says "Looking at N picture(s).
  Please wait." (`describe_wait_text`; done: `describe_done_text`).
- Egg pictures are looked at ahead, quietly, 30 s after Wire-Pod starts, so a pop-up's reply
  doesn't wait.
- Live snapshots have no Goverlord bubble ([8.7](#87-photo-whats-in-front-of-you)).

---

## 9. The chat page and its bubbles

**The raw bubble template (GOLDEN, being built type by type).** Every bubble type is built on one
template: `rawBubble(spec)` makes the bubble whole (kind, classes, header, lines);
`bubbleTask(b, promise)` adds a job it must finish before it shows; `placeBubble(b, then)` builds it
in a hidden staging area outside the chat (so it never moves the chat's scroll), waits for its tasks
(10 s at most), then puts it in the chat and shows it - and only then does the next event start.
Shell + placeholders: a voiced line waits as Generating + hourglass, then lights word by word; a picture
still loading is an hourglass in its own place (`pictureWait`; one not made yet is asked for again
every `chat_window.picture_retry_ms`, at most `picture_wait_max_ms`), so no bubble is held back by
its pictures. A line that breaks is written to Wire-Pod's log ("GOPOD_PAGE error ...") and passed,
so the chat never stops until F5; a line that hasn't handed on after `page_sync.item_max_ms` (90 s)
is logged with what holds it and passed, and after an F5 the page tells Wire-Pod how far it is.
The hourglass drains once it's on the page (it's made before it's placed). A Brobot's answer
under his header is 75% of the header's size, with a little air between characters. The exceptions keep their own rules: the chat window (scrolling, the baton),
the Operator bubble (the one input type) and the system window (the menu). On the template so far:
system / terminal answers, reporter and manager lines (their own bubbles), Goverlord's lines, Brobot replies.

### 9.1 The Chat Bubbles Window: the machine

**Draft for signature** - once the owner signs it, this is the locked rule set.

**What it is.** A Brobots layer on native Wire-Pod (the overlay, "chat bubbles lite"):
Wire-Pod's own Log page gets a scrolling window of bubbles that shows the show as it happens -
every line said, by whom, to whom, with its pictures, timers and eggs. The terminal (journal)
carries the same show, line for line (`GOPOD_FEED <id>`, the page answering
`GOPOD_PAGE played <id>`).

**From the Log click to the bottom of the window:**
1. **Log click** - the page counts as here (seen, clicked, chat open). Before that, Wire-Pod waits
   ("Goverlord awaits your return") and nothing plays into an empty room.
2. **The feed** - Wire-Pod numbers every line it adds to `gopod_chat_feed.json`; the page plays
   them one at a time, in order, and says "played" for each.
3. **A voice is made before its line shows** - the line waits in its own bubble with
   "Voice file generating." and the hourglass, then its words ([9.4](#94-golden-the-generating-template)).
4. **Every bubble's life is the same:** built whole while hidden (pictures, QR codes, weather
   face, flags in place) → the chat scrolls once to make room → it shows when the bubble above is
   done (the baton) → its words are revealed as they're said → said (85%).
5. **Who reveals the words:** the Brobots' replies - Wire-Pod names the piece being said
   (`gopod_speaking_now.json`), the page reveals that piece word by word. Reporters' and
   Goverlord's lines - the page plays their voice itself, so it reveals the words on that voice's
   own word times (`voiceLitLine` / `voiceLitStart`). SRT songs - the soundtrack's times. Word,
   then its punctuation, then emojis; numbers, clock and year as said.
6. **Small processes plug into that one life** and work for every bubble that needs them: the QR
   codes, the egg marquee, the magnifier on every picture, the weather face, flags as pictures,
   the search-results box, the hourglass.
7. **Scroll to the bottom** - the newest bubble ends just above the typed-input line; the chat
   follows the highlight; nothing scrolls again until the next bubble.

**Shared helpers - one each, used everywhere:**
- `voiceWaitOn` / `voiceWaitEnd` - the one Generating template, keyed by voice file.
- `slotWaiting` / `slotReady`, `fillSlot` - the grey "Voice file generating." slot, filled in
  place by the line that follows. (Not used for the Operator's own bubble.)
- `voiceLitLine` / `voiceLitStart` - the word reveal for every voice the page plays.
- `magnify` - the magnifier on every picture, QR code (QR1, QR2) and weather face.
- `whenShown` - anything that waits for its bubble to show: the weather sound, the search
  results' typing, and the reveal's 15 s fallback.
- `cornerBox` - QR codes and the weather face in a bubble's top-right corner.

### 9.2 Bubble kinds and headers

| Kind | Header | Look |
|---|---|---|
| System (Goverlord) | "[grey icon] Goverlord: GOPOD Brain [grey icon]" | grey. Every one is headed so; "Reporter 0" is only a decoration label. |
| Operator | "[operator icon] [input] to Brobot N [Brobot N eye-colour icon]" | white-eyed icon. To Brobots: both icons. |
| Reporter | "[white icon] Reporter 1: Brobot 3 [white icon]" | white-eyed icons |
| Brobot | "[eye-colour icon] Brobot N:" | his eye colours, one icon |
| Narrator | no header | one per poem section |
| Manager | CHALK's chalkboard, PLAYHEAD's tape-brown (the colour of vintage video tape cassettes) | see [12](06_CAST_AND_GAME.md#12-managers-chalk-and-playhead) |
| Goverlord (terminal) | Goverlord's header | text at 50% size; system answers |

- System answers and non-numpad status lines carry Goverlord's header.
- **Header icons mirror each other.** For headers with more than one icon, every icon after the
  first plays `assets/brobot_eyes_half.webp`: the same eyes, its frames starting at the loop's
  half (2.72 s of 5.36 s), so the icons' animations mirror each other rather than move together.
  Each icon keeps its own eye colour (`gopod_chat_bubbles.css`). The original `brobot_eyes.webp`
  is untouched.
- **Eye colours from Wire-Pod:** each Brobot's own custom eye colour, from Wire-Pod's stored
  robot settings (`vic.RobotSettings`), becomes the bubbles' colour at Wire-Pod's start and every
  30 s, never during a run (`gopod_eye_colors.go`). A Brobot with no custom colour keeps the
  page's own.

### 9.3 The chat rules

**GOLDEN (rules 10, 11, 14, 15 in [Golden rules](08_RULES_AND_TESTS.md#17-golden-rules)).** A bubble is built whole
while invisible, the chat scrolls once, then it shows, with its voice when it has one; the next
bubble takes the baton only when the one above has finished.

1. A bubble is built fully hidden, away from the chat, so it never moves the scroll: out of the
   chat's flow, taking no room and opening no scroll space.
2. It's ready once its content has loaded and rendered (pictures loaded, its voice loaded and
   ready to play).
3. The full bubble gives its final height; that height means it's ready to show.
4. Only then, at its turn in the queue, does the chat scroll and show it. It comes in from the
   bottom. Its voice plays once it shows.
5. A short bubble stops bottom to bottom (its bottom just above the typing line); a tall one
   (taller than the window) stops top to top. Otherwise the chat stays at the bottom. One
   scroller, 1 s ease, never up on its own.
6. Every bubble runs and finishes 100% before the next one shows.
7. Headers as in [9.2](#92-bubble-kinds-and-headers).
8. Live operator lines (keyboard, numpad, and the robot's mic held during a warm-up, "Mic to
   Brobot N"): an Operator bubble read out in the Operator voice
   (`reporter_voices.operator_voice`, `af_sky`). A playing draft's question keeps its reporter's
   voice. Lines to CHALK/PLAYHEAD are read in the Operator voice too.
9. No time limit takes the baton early: the next bubble waits for the one being said. Only a line
   lit longer than `chat_window.speaking_max_ms` (30 s - a frozen robot, a lost end signal) lets
   go. However a voice ends (played, errored, hit its cap), its line counts as said.
10. **Narrator bubbles** are the nap's captions: the song's subtitle (`.srt`) lines, one bubble
    per poem section, each phrase lit on the soundtrack's clock - nobody's voice. Like every
    bubble, each is built out of the chat's flow and joins it when it's placed. A timed bubble is
    sent early (`narrator_lead_seconds`) and shows at its first line's second.
11. **A bubble that grows on screen** - a new message, a picture loading, a picture added to a
    question bubble - is placed again, whole, while it's in view.
12. **The LLM warm-up bubble is the exception:** it shows first and is kept at the bottom.
13. **Pop-up placement:** once the reply joins an egg, the chat aims the reply's "[icon] Brobot
    N:" header at the window's top; if the bubble's bottom reaches the window's bottom first, it
    stops there - no extra space below (rule 5, aimed at the reply header).
14. **Log opened after a refresh:** the chat goes straight to its bottom the moment the window
    gets its height.
15. **The highlight is held, not the next bubble's place:** the chat keeps following the
    highlight while a next bubble waits (a staged bubble is never "the end of the chat"); the
    highlight stays on the bubble the Brobot is saying (past its first piece) and never lands on
    an egg.
16. **A voice waits** while its bubble - or any bubble before it - is still waiting to show, as
    long as the baton can be held.
17. **Backpack question and its egg:** a line the robot heard is logged once its answer is ready,
    after its egg; it goes above that egg and the reply still joins it.
18. **Bingo ball:** the called ball shows under its "Bingo ball captured. B-1" bubble when "B-1"
    lights (its room kept while hidden).

**The window's numbers are data.** The page's timings and sizes (baton, scroll, countdown hold,
marquee, narrator fades, reveal fallbacks) are `chat_window` in the settings; the numbers in the
page code are only fallbacks.

### 9.4 GOLDEN: the Generating template

A voice is made before its bubble shows. One robust, deterministic mechanism for every voiced
line:

- Wire-Pod sends a line whose voice isn't made yet first, marked `generating_marker+"wait:"`
  (`gopodSayShown`, the one path for shown voice lines). The waiting line names its own voice
  file (`wait:<file>|text`, `voiceWaitFor`), so its voice always finds it - a bubble posted in
  between never leaves it stuck.
- The page shows the generating text (`reporter_voices.generating_text`, "Voice file
  generating.") and the hourglass, rocking and draining, under the header in that line's own
  bubble, until its voice plays; then its words. "off" (voice failed) shows the words.
- one Generating look everywhere - the text, then the hourglass, the text's own
  height, on its line (`.voice-wait`; `voiceWaitOn` and `slotWaiting`).
- A skip's end (Esc, numpad /, `system-skip`) takes the chat to its bottom (`skipEndScroll`).
- **One question at a time** (`line_queue`): every question - typed, numpad, a persona's, the comm
  check's - takes its turn in the order they came; the next starts only when this one is complete: a
  line from each one asked (his answer), then the page has played it all (when a page is here).
  Commands (system-, stop run, thanks, skip, start ...) never wait. A new line clears only the opened
  bubble of whoever it's to, so a reply still on its way keeps its own. `TestLineQueue`. The page replay
  lists each bubble kind with its parts in order - one kind built more than one way shows up there
  (BUBBLE TEMPLATE CHECK).
- **The skip gate** (`skip_gate`): every skip (Esc, numpad /, the menu, typed or
  spoken "skip") is off until the gate opens. A skip while off does
  nothing (logged: skip ignored). no words matched - Wire-Pod opens it as the opener's message
  step right after the voice files check (`wait_voices`) shows, whatever its words
  (`gopod_opener_song.go`, feed line `skip_on`, just ahead of that message; locked in
  `TestWarmupSequenceGolden`). The page replay presses Esc there (SKIP: "skipped (Esc)").
  After a reload the gate is worked out again, once per page load: `skip_on` already seen
  and the opening not over yet = open.
  The gate stays open once opened - every skip works, not only the first. A skip shows every
  waiting voiced line whole (not the Brobots' own), Goverlord's lines included, and ends once the
  page's queue is empty. `stop run` is global: it stops a run, the tour and the comm check, clears
  the line queue and skips the page.
- **GOLDEN voice start**: every made voice plays the same way - its bubble shown,
  then its length and its word times (`.words.json`) loaded (up to `VOICE_READY_MAX_MS`, 3 s),
  then it plays and lights. A just-made voice no longer starts before its word times arrive
  (it lit by the estimate, out of step).
- It works the same for an operator line, a reporter's, a Goverlord bubble, a manager's, and the
  opener's System llm bubble. The Operator's own words wait the same way ("op:<file>"). No
  separate "Generating..." bubble.
- One template: `voiceWaitOn` / `voiceWaitEnd`, keyed by voice file.
- A voice asked for while Kokoro is still starting waits for it (up to `start_wait_seconds`)
  instead of failing.
- The stings (grunge, dJent3b) and the numpad key sound are loaded once and replayed; the
  background loop is decoded once and looped.
- Code: `gopodReporterFileShown`, `reporter_voices.generating_text`/`generating_marker`.

### 9.5 Highlight and word reveal

**The highlight rules, in one place (GOLDEN).** Highlighting is one mechanism for every bubble:
it comes in, lights by these rules, and gets out.

1. **The chain, largest to smallest.** A reply is cut once, by Wire-Pod, into **pieces** (per
   pause: commas, sentences). The piece being said is lit **word by word**; inside a word, a run of
   single digits, a military time or a clock time lights **one number (or spoken part) at a time**.
2. **Three states.** Not said yet: hidden. Being said: solid white. Said: 85%.
3. **Three timing sources, each naming its bubble exactly:**
   - **A Brobot's speech:** Wire-Pod reports the piece number *and the bubble's chat event*
     (`gopod_speaking_now.json`: `piece`, `feed`); the page lights that piece in that bubble.
   - **A voice** (reporters, managers, Goverlord): the line names its own voice file
     (`lit:<file>|` or, while it's made, `wait:<file>|`); the words light at that voice's own word
     times (Vosk) as it plays.
   - **The narrator:** its SRT times, on the soundtrack's clock.
4. **Timing within a piece:** a voice uses its Vosk word times; a Brobot's speech gets the real
   start of each piece from Wire-Pod and a best guesstimate for the words in it; with no timing at
   all, a guesstimate (`word_reveal.seconds_per_word`). A line whose voice never plays shows after
   `voiceLitWaitMs` (15 s); a Brobot that stops speaking leaves the rest of his bubble shown as said.
5. **Nested parts** (a manager's part, an egg summary inside a Brobot's bubble) light on their own;
   the Brobot's own pieces never count them (`ownSays`).


**Pieces, per pause (locked - change the split, highlight or bubble look only with a plain
before/after and the owner's OK).** A Brobot's line is cut into pieces at its pauses (comma
pieces): `join_up_to_words` 0. Locked by `TestHighlightPiecesGolden`.

**Heard, not read ahead.** A highlighted bubble (a Brobot's reply, a scripted song line, the
narrator) is built whole, but its words are invisible (0%) until heard: the piece being said is
solid white (100%), a piece already said 85%. The pieces before a lit one count as heard; when
the next bubble takes the baton, or 10 s after the robot stops with nothing more lit, the rest
shows. A restored chat shows all. `.spoken` / `.speaking` in `gopod_chat_bubbles.css`,
`lightSpan` in `gopod_brobots_chat.js`.

**GOLDEN: word at a time.** The piece being said starts at 0% solid white, then each word turns
100% solid white as it's said and 85% once said.
- Every word is its own step (`word_reveal.seconds_per_word`, 0.38). A word's punctuation, emojis
  and ° are steps of their own after it (`word_reveal.seconds_per_mark`); marks inside a word stay
  (Let's). A flag is a step. A math formula is one step.
- A silent lead-in of two words' time comes first, then the first word lights
  (`word_reveal.start_delay_words`), so the reveal doesn't start before the Brobot speaks.
- **Numbers one at a time:** in the piece being said, a run of single digits ("the 1 4 2 8 5 7
  repeats") and a military time light one number at a time (`digit_highlight`,
  `digit_highlight.seconds_per_number`).
- **Clocks and military time step as said:** a clock steps "1", ":", "43", "P", "M". "1027 Hours
  Military Time" lights 10, then 27, then "Hours", "Military", "Time" (the label after the
  numbers); "1:16 PM or 1316 Hours Military Time." reveals 1 /: / 16 / P / M / or / 13 / 16 /
  Hours / Military / Time. A year after its date steps "20", "26" and is said "20 26"
  (`gopodSpeakYears`).
- For the Brobots' replies (Vector's own voice, no audio here) it is a timed estimate, not tied
  to the audio; when the piece ends the rest shows. The page code: `digitUnits`/`digitSequence`.
- The word reveal runs only in live, scripted and narrator bubbles (no flicker in plain ones).
- **Every spoken line, inherently**: a line is lit through the one `lightSpan`; its bubble wears
  `voice-lit` (every voiced line's way: `voiceLitLine`, `voiceLineTieIn`, `replaySaidLine`), so it
  is hidden until lit and revealed word by word. No song, no list, no per-line switch asks for it:
  the opener song (`00_llm_warming`, its `_llm_song` from `_new_song/_llm_song`), the reporters,
  the warm-up, the eggs, Goverlord, the Operator and the Brobots all come through it. Checked by the
  page replay's SPOKEN LINES (feed of a full start: 0 of 19 not per word).
- **A said line, replayed** (`replaySaidLine`, the template): a line he already said (his opener,
  said while thinking) shows when its bubble shows - from 0%, word by word, then said - and holds
  the next line until done. It first waits `word_reveal.replay_settle_seconds` (1.8) at 0% so its bubble
  settles into place. His opener ("Got it!" / "Nice try, Mr/Mrs ...!") is the first.
  LOCKED by `TestOpenerMechanismLocked`.
  It is not one of his reply's pieces (`ownSays` leaves it out): Wire-Pod names the piece it says
  by number, written just before each say goes to the robot (`gopodWriteSpeakingPiece`), and the
  page lights that number (counted, the opener put every lit piece one behind his
  voice - "highlighting appears late ... weird segmentation passes").
- While a link's words are revealed, only the shown words are underlined.

**Forced alignment: word times from the voice.** The voice worker has Vosk listen to each voice
it makes and gives every word its heard time (Kokoro's own time only for a word Vosk missed; a
mark where its word ends) - `"aligner": "vosk"` in `<name>.words.json`, saved beside each voice.
A voice without one, or timed by Kokoro's estimate only, is made again on its next use. The page
reveals a reporter's or Goverlord's line (and an egg's congratulations) word by word at those
times (`alignUnits`) - a mark or emoji when its word ends, a word said another way
("crushin-ate-a") between its neighbours.

**One reveal, any source of times.** The word reveal (`digitSequence` + `alignUnits`) uses word
times whenever a piece carries them, whatever made them - a Kokoro voice's `.words.json`, or a
song's `narrator_words.json`.
- **A song's narrator:**
  `gopod_probes/tools/gopod_align_captions_001.py <narrator audio> <song folder>` - Vosk (Wire-Pod's own speech model) times the words it hears; each SRT caption's
  words are matched in order within its seconds; saved as `narrator_words.json` (Kokoro's shape);
  the audio isn't kept. The song engine puts each caption's times on its piece
  (`@s-e ~<base64> words`). Baby Brobots has it (170 of 179 words heard; the rest between their
  neighbours).
- Without word times, a narrator caption's word reveal starts at once and is scaled to fill its
  SRT segment exactly, first word to last.

**Follow the highlight.** The chat holds still until the piece being said is a quarter of the
window below centre (`highlight_scroll.dead_zone`), then one smooth move puts it back at the
centre - never past the chat's end (the bottom/bottom rule), never back up on its own. Once the
bottom is reached, the highlight carries on down the bubble.

### 9.6 Pictures

- **The magnifier (GOLDEN):** every picture has it - hover or tap to see it full width. New
  pictures too, QR codes and weather faces included.
- **Pictures in one row:** a bubble's pictures share one row, each scaled; the magnified one takes
  the full width.
- **A song's picture:** `{"note": "picture", "picture": "granny.jpg"}` (one file or a list, from
  the page's `pictures/` folder) shows them in a picture bubble at that step, alone or as a combo
  part. Label the step after it, for example `host_say_img-granny_0301`.
- **Narrator pictures:** a song's `section_pictures` (knobs.json, section -> file in the page's
  `pictures/`) puts that picture in the section's narrator bubble with the lit words, with the
  magnifier.
- **Picture rows in the menu:** only a system command's picture row ("menu:" mark) goes in the
  menu's results while it's open; songs' and topics' stay in the chat.

### 9.7 QR codes, weather faces, flags and links

- **QR codes:** a bubble whose words hold "crushn8r.ca" gets the crushn8r.ca QR code;
  "newsletter" (or "crushn8r.ca/subscribe") the newsletter one (crushn8r.ca/subscribe); "nightlife"
  or "Mississauga" the flare4.com one. CHALK shows them ([12.3](06_CAST_AND_GAME.md#123-managers-inside-a-brobots-bubble)).
  In a corner, QR1 sits top-right in the header row; scrolled above the chat's top it jumps to
  the bubble's bottom-right (`corner-low`, marquee inset kept). QR2 (`qr_codes.second`) goes
  right after "...subscribing to our newsletter!", the last step of the reveal. The magnifier as
  ever. `qr_codes` in the settings folder (`webroot/gopod_settings/`) (`qr_codes.codes`: the words that show each),
  pictures in `webroot/qr/`.
- **Weather faces:** a bubble holding the live weather line gets Vector's own weather face for it
  (sunny, cloudy, rain, snow, thunderstorms, windy; clear at night = stars; cold has none),
  animated, Vector's own size (184x96), in the speaking Brobot's eye colour, in the same
  top-right corner, with the magnifier. Its weather sound is loaded with the bubble and plays once
  as it shows, under the Brobot's voice (`weather_card.volume`). Only the weather's own words pick
  the face, never the place's. `weather_card` in the settings; the kit is `webroot/weather/`, all
  of it shown on `gopod_weather_kit.html` (locked as is; linked from Wire-Pod's weather setup
  section, the overlay's `setup.html`).
- **Flags as pictures:** a known country's flag is drawn from `assets/flags` (Twemoji, CC-BY 4.0),
  because some browsers (Windows) draw a flag emoji as its letters ("US").
- **Links:** any "[words](/page)" in a bubble is a link, like the Weather Kit link.

### 9.8 The engine and the page in dialogue

One event list, one source of truth: the web page and the terminal run as one coordinated
dialogue.

- **The page is here only once it can play:** after `wpr 2`/F5, the click (the Log button).
  Before it, Goverlord awaits and runs wait (a browser plays sound only after a click).
- **One event list:** every chat-feed line is in the journal as `GOPOD_FEED <id> <who>: <words>`,
  in order - voices, heard markers, warm-up messages, bubbles. **Live replies are feed lines too**
  (via `llm`).
- **The page answers:** each second it says how far it has played the feed; the journal writes
  `GOPOD_PAGE played <id> <who>: <words>` under the engine's own line.
- **Heard is the truth:** `GOPOD_HEARD <who>: heard` / `not heard after Ns (going on)` / `no voice
  file`, for every voice the engine waits on. The page reports voices that didn't play (missing,
  blocked) with the next heard marker; a song's reporter and ready lines get `heard` or
  `not_heard` back, and the song engine prints a not-heard one.
- **The engine conducts:** Wire-Pod times itself by the page instead of guessing. (1) A Brobot
  speaks once his bubble is on the page - up to `page_sync.reply_wait_seconds` (10) - so he never
  talks over a line the page is still playing (`gopodWaitReplyShown`; `GOPOD_SYNC` lines when he
  waited or the page was behind). (2) The page tells how long it held each line before showing
  it; over `page_sync.held_log_seconds` (3) it's in the journal (`GOPOD_PAGE held <id>`). The
  page still keeps its drawing order (one bubble after the one above, the countdown, the SRT
  narrator's soundtrack clock); the engine waits on those instead of running ahead of them.
- **Every wait on a line:** Wire-Pod goes on the moment the page says it has played it (the heard
  check). With no answer, after the queued lines' own lengths plus
  `reporter_voices.heard_margin_seconds` (3). No page-lag padding is added.
- **Wait until I'm finished:** a voiced line waits until the page has played it - its own length
  plus up to `reporter_voices.finish_wait_seconds` (120) - while the chat page is watched and
  here. With no page, the cap is 2 x `max_wait_seconds`.
- **A voice's cap starts when it plays** (its length + 5 s; 60 s if unknown), so it's never cut
  or left to start late, out of order.
- **Every live reply is accounted for:** the page writes
  `GOPOD_PAGE live reply <id> Brobot N: shown` (or `NOT shown - <why>`) under its
  `GOPOD_FEED <id>`; Wire-Pod writes `GOPOD_FEED_CHECK ... the page played past it with no bubble`
  for a missing one and `GOPOD_FEED_CHECK live reply twice in a row` for a doubled one.
- **No orphan state file:** `gopod_rich_display_ui_state.json` is not written.
- See it: `journalctl -u wire-pod -f | grep -E "GOPOD_(FEED|PAGE|HEARD|AWAIT|OPENER)"`.
- Locked by `TestFeedDialogue`.

### 9.9 The system window

**GOLDEN, locked - change only with the owner's OK.** The system menu's window:

- A Goverlord system bubble at full width holding the menu bubble; no header (to recover window
  space). The system bubble centres its contents.
- Two windows, menu above and results below, start 70/30. The thin bar between them (row-resize
  cursor, a grip line, stretching across) sets the menu's share, 10% to 90%; the results take the
  rest. The results window is full width, empty or not. Window 2 overlaps window 1.
- The menu a little wider (right padding 2px); each window's scroll bar in its own lane, clear of
  the rounded corners.
- Menu at 50% font, for more room below.
- While the menu is open, the chat is scroll-locked at its top.

### 9.10 The background loop and the speaker icon

- `background_loop` (`BrobotsGroove-Loop-1.wav`, 6 s) plays on the page over and over with no
  gap, from its top again at every session clock reset to 00:00:00.
- The speaker icon after the text field (🔊/🔇) turns it on or off (and opens or closes the
  system menu). Every new Wire-Pod start sets it off.
- Hovering the icon shows a vertical volume bar (0-100%, the loop's share). Let go, its value is
  saved to `background_loop.volume`, so it holds over restarts. The slider gives way when the
  settings' loop volume changes.
- Every run quiets the loop at its start and, as its very last line, sets it back to the toggle.

## More in this part

**GOPOD as a phone app** (phone first; `gopod_https.go`, settings
`app`): Wire-Pod also serves GOPOD securely on `app.https_port` (8443) - the same pages, nothing
copied - with GOPOD's own certificate: a local authority made once (the phone trusts it once from
`/gopod_ca.crt`) and a server certificate for this machine's names and addresses, made again when
they change; all private in `app.cert_dir` (operator's home). The app: `app/gopod_manifest.json`
(name GOPOD, icon - his masked minifig picture, - start
`/?app`, which opens the chat: a visitor's chat-only view), `gopod_sw.js` (keeps nothing; every
page and voice). Locked by `TestPhoneAppHolds`. Soft-locked; trusting the certificate is followed by restarting the browser.

**Visitors and the chat sign-up** (`gopod_visitors.go`, settings `visitors`): the admin is GOPOD's own
machine or a device in `visitors.admin_addresses_file` (private, outside the repo); everyone else
is a visitor, one per device (its browser ID, else its address). A visitor's first line is their
username request - first come, first served, a taken name gets a number (name1, name2); lines after
it are dropped ("Waiting for approval") until the admin approves it in the system window's "Local
network visitors" slot (nameless visitors listed by network address and device; Approve / Clear).
With participation ON (starts OFF every start) an approved visitor with their box ticked talks to
the Brobots under their username; unticked = muted ("Muted"). A visitor never runs the show
(commands dropped). Profiles (name, approval, box) are saved in `visitors.profiles_file` (private)
and come back after a dropped connection or a restart ("like the brobot reboot recovery thing").
A visitor reaches the Brobots chat and nothing else: Wire-Pod's web servers (8080, 8443, 80) put
`gopodVisitorGate` in front (`vars.GopodGated`) - the page, its scripts, styles, pictures and voices,
GOPOD's state files, the chat's calls and the heard lines pass; Wire-Pod's settings, config, logs
and every other call are refused (logged "refused"). The page reads only the lines a Brobot heard
himself (`/api-sdk/gopod_heard_lines`), never the whole log. A visitor's page is always the chat
alone (no way back) and only paints what the main page shows (the display mirror). Locked by `TestVisitorGateAndGreetings`. Every visitor event is logged (GOPOD_VISITOR): a new device, coming back, every line sent and what
became of it (name asked, waiting, approved-OFF, muted, a GOPOD command, passed), approve / clear /
mute, participation ON/OFF. On approval the visitor's text field turns from "Waiting for approval"
to "Username Approved" (`visitors.approved_text`, `approved_show_ms`), then clears; with
participation OFF an approved visitor's line answers "Username Approved" - kept across ON/OFF; ON,
unticked answers "Muted". A visitor's typing row is two lines - the Brobots, then the text field (Record, timer
and speaker are the admin's); Wire-Pod's own intents pass for a visitor, GOPOD's commands ("thx
bros", start, stop, system-) never. Every scroll rule measures the chat above the typing row
(`chatViewHeight`), and a chat at its end stays at its end when the row changes height.
Locked by `TestVisitorSignUp`. RULE (the
admin list at "~/" was read from root's home, so the PC stayed a visitor): every "~/" path is the
operator's home (`gopodHomePath`), and every program GOPOD starts runs with the operator's HOME
(`gopodAsOperator`; only systemctl/lsusb exempt; `TestPythonRunsAsOperator` scans every start).

**Chat-only view, from the Log button**: on a visitor's device (a phone), tapping Log opens
the chat alone - the Brobots chat and the typing row, the whole screen, no Wire-Pod settings;
"‹ Wire-Pod" goes back. The admin's Log is unchanged. `?visitor` in the address shows the visitor's
view on GOPOD's machine. Locked by `TestChatOnlyViewHolds`.

**The whole page on a phone** (C4 start): `css/gopod_responsive.css`, loaded
after Wire-Pod's own `style.css` (untouched) at the same 768px phone breakpoint - the Brobots'
battery rows stacked ("can you stack the home page battery level entries on top of each other?"),
the typing row wraps instead of running off the right edge. Checked at 390px wide: nothing runs
past the edge on any section, Server Settings or Bot Settings. Then (same day): Bot Settings reads "Vector Config"
(GOPOD's copies `webroot/sdkapp/index.html`/`settings.html`, title only; the overlay serves a
folder's own `index.html`); "Scroll to bottom" on one line; the typing row is one line while it
fits ("Brobots 1 2 All[1] [text] | Record? [] | 00:56:42 |🔇") and three when it doesn't
(`.stacked`, any screen; `chat_window.typed_text_min_px`); 🔇 opens the system window for the
admin only (the page on GOPOD's own machine; a PIN for other devices comes with C5) - a visitor's
🔇 is the loop alone; the system window is exactly the room above the typing row, whichever group
is open. Locked by `TestPhoneLayoutHolds`.

**Bubble placing never waits on a picture; every voice preloads** (every
bubble after the granny photo waited until F5, on both pages): `revealTimer` means only "bubbles
being placed" - a picture's load wait (`whenBubblesReady` from `showPictures`) has its own timer.
Every voice starts loading the moment it reaches the page (stings have their own pool), so it
plays as its bubble shows. Listed expletives (`"strike": true` in the Brobots lingo list) show
struck through in every chat line (`gopodStrikeExpletives`, one place: the feed), said as their
substitute. Locked by `TestStrikeAndPlacingHold`.

**Root-to-leaf pass, the way in** (`gopod_operator_line.go`): every Operator
line - typed, numpad (its mic recording), visitor, command, system-set, a manager's, Record
Session's answer, GOPOD Yourself - is shown by `gopodOperatorShow` and read out by
`gopodOperatorRead`; a line a Brobot heard himself is shown once, the moment it's heard
(`gopodHeardLine`), never again on its way in (was: a backpack "start interview" showed twice).
Nothing else posts an Operator bubble (`TestOperatorLineTemplate`).

**Root-to-leaf pass, the feed**: everything the page shows comes from Wire-Pod's
numbered feed - one append (`gopodChatFeedAddEventVia`), mirrored to the journal, the newest
`page_sync.feed_max` (300) lines kept - and the page plays it in its numbers' order (the old time
order merged lines read from Wire-Pod's log, which the page no longer reads; that code is gone). A
viewer queues none of it (the display mirror). `TestFeedIsTheOneLine`.

**Root-to-leaf pass, the bubble**:
the one way - built by `rawBubble`, placed by `placeBubble` (hidden, pictures loaded, then shown;
the next feed line waits). Done 3a-3f: chapter, status, Operator, countdown, search results,
pictures / pop-ups / eggs, the shell and every part inside a Brobot (`nestPart` / `nestPlace`) and
the narrator's construction; kept on purpose: the warm-up bubble (stays at the bottom) and the
narrator's SRT-timed reveal. The log-era "backpack line goes above its answer" moves are gone (a
heard line comes first now). `TestBubbleOneWay`.

**GOLDEN Easter egg template** (`eggBubble`): every
egg one way - the shell (a Brobot's marquee wrapper, or Goverlord's / a host's board), Goverlord's
congratulations as one voiced line (Generating + hourglass, then lit word by word), the prize
(`prizeItems`: pictures held by hourglasses, a card, the math, a weather face), "MouseOver to Zoom
In" under it as the same voiced line's last part (dormant until its voice reaches it, lit in
turn), and a Brobot's reply joining it. The bubble names its voiced line ahead
(`voiceLineAhead`), so its voice comes to it, never found by a search of the chat. Which other
bubble types still find their voice by search.

**Root-to-leaf link 4, the voice**: every voiced line shown some other way - the
Operator's own line, each warm-up line, an egg's congratulations - names itself when it's drawn
(`voiceLineAhead`: its words, letters and digits only, and the element that holds them), so its
voice ties to it directly; the search of the chat is only a fallback. The dry run types the way a
person would: the next line only once the page has caught up and no Brobot is speaking
(`gopod_speaking_now.json`). **Page replay** (`gopod_probes/tools/gopod_page_replay.sh`): the
real chat page code with no robots (jsdom, kept in `~/.cache/gopod_page_replay`) - every egg
kind, both prizes, the Operator lines' voices; reports what was drawn and lit, and any error,
stuck line or loop. A saved feed can be replayed too.

**The visitor's own page**: a visitor's "/" is
`webroot/gopod_visitor.html` - the chat alone, from the first paint, no Wire-Pod menus or scripts
(`gopodVisitorGate`; "?visitor" shows it on GOPOD's machine). `TestVisitorGetsTheChatPage`.

**Root-to-leaf link 6, the visitor mirror**: checked by the page replay, which now
runs a visitor's page beside the main page and compares them every quarter second - words and
what's lit, bubble by bubble; a difference lasting over 1.5 s fails `system-test-replay`
(in step 85%, longest catch-up 0.8 s, never stuck). A visitor still counting from before a
Wire-Pod restart gets the display at once (`gopod_display.go`: only "no change" is 204).

**Root-to-leaf link 5, the answer chain**: an answer with a follow-up coming (All's
second Brobot, Explain the Math's follow-up) holds its line's turn - a new line's answer waits
until the follow-up is done (`gopodExchangeWait`, `addressing.all_next_wait_seconds` at most; the
Operator bubble still shows at once; a lead that stops early lets go). With nothing left to point
at (every point and egg found) a free answer without its reminder isn't drift
(`gopodReminderNone`). `TestExchangeHoldsTheNextLine`.

**The All pick stands**: the page's All 1/2 choice goes with every typed line and every numpad press
and stands until it changes, so a spoken line keeps it; an All math line with Brobot 2 first has
him lead and Brobot 1 follow up (`gopodMathLead`). `TestAllPickStandsAndLeadsMath`.

**Got it 1 / Got it 2**: `chat_moves.got_it` - each Brobot's own copy of the golden sequence (moves, words,
searching pace), both exact copies of the golden values to start; `gopodGotItFor(esn)`, locked
by `TestGotItPerBrobot`.

**One page owns the show: main page, viewer pages** (`gopod_main_page.go`): Wire-Pod names one main page - it plays the sound and its shown, played
and heard reports time the run. Every other open page is a viewer: the same bubbles, silent, its
reports ignored (a viewer's tab hidden pauses nothing). The first admin page is main, even before its click (after a restart the page reloads itself and
plays nothing until clicked), and keeps it while it reports in; a page that can play (shown,
clicked) takes it from a main page that can't; gone past `page_sync.main_lease_seconds`
(12 s), the next admin page takes over. One admin page at a time: a second one only watches, its
text field says so (`page_sync.second_admin_text`), Wire-Pod logs it (GOPOD_PAGE second admin
page) and `system-status` counts them. Visitor pages can be many. While the opening waits for
the page, the System llm bubble shows `session.awaits_text` with its hourglass. The page's id lives with its tab, so F5 keeps
its role. `?view` in the address = a viewer by choice. Nothing plays until a page knows its
role. Locked by `TestOnePageOwnsTheShow`. Every other page is a display mirror: the main page sends what it shows - its newest
`page_sync.display_bubbles` bubbles as drawn, highlights and all, each with its own key - through
Wire-Pod (`/api-sdk/gopod_display`) at most every `page_sync.display_ms` (50); every other page runs none
of the chat (no feed, voices or warm-up) and only paints that: a bubble it has is changed only
where it changed (no picture reloads, no flicker), new ones added, gone ones removed. The visitor page
fills the phone's screen (100dvh, no back-button room) and, after any load or refresh, starts at the
end. One send at a
time, the next after it lands (the phone: "punctuation & ending emojis flicker" - two sends
crossed and the phone stepped back; the replay, with a real network's delays: 4 flickers before, 0
after; its VISITOR FLICKER check). Scrolling is
each device's own: GOPOD's rules on its
own screen - a new bubble bottom/bottom or top/top, the lit words followed. Its own typing row and
sign-up notes stay its own (`TestDisplayMirrorHolds`). Only a page on GOPOD's own machine (localhost or the machine's own network address,
checked by Wire-Pod from where the page connects) is the admin and can be the main page; any other
device is always a silent visitor (the phone took the main page after a restart and the PC
went silent; `TestOnlyGOPODsMachineIsMain`). The phone stays a visitor: no system window.

**MASTER LIST: content -> its template** (aim: "Single unified template ... or at least the minimum set
of epic cooperative templates"). Every chat bubble = header + content; each content's rules live
in the settings, its template runs them, a test locks it ("none yet" = no lock test).

| # | Content | Template (code) | Rules (settings) | Lock |
|---|---|---|---|---|
| 0 | Which page plays (main / viewer) | `GopodMainPageClaim` (Wire-Pod), `setPageRole` (page) | `page_sync.main_lease_seconds` | `TestOnePageOwnsTheShow` |
| 0c | Visitors, sign-up, mute | `GopodVisitorLine`, `visitorsGroup` (system window) | `visitors` | `TestVisitorSignUp` |
| 0b | Chat-only view (a visitor's Log) | `chatOnlyOn` (page) | `gopod_responsive.css` (`body.gopod-chat-only`) | `TestChatOnlyViewHolds` |
| 1 | Bubble header | `appendBubble` head row (page) | `bubble_kinds`, `display_names`, `brobot_bubble` | `TestHeaderTemplateHolds` |
| 2 | Bubble placing, scroll, the baton | `revealWhenPlaced`, `batonHeld` (page) | `chat_window`, `gravity`, `highlight_scroll` | `TestInterviewEndAndPauseHold` |
| 3 | Voiced line, own bubble (Goverlord, reporters, chapter, manager) | `gopodSayShown` -> `wait:`/`lit:` -> `voiceLitLine` | `reporter_voices`, `sign_off`, `hourglass` | `TestEveryVoicedLineUsesTheTemplate` |
| 4 | Voiced line shown some other way (warm-up, opener, egg, Operator) | `gopodSayForShownLine` -> `forwait:`/`forlit:` -> `voiceForLine` | `reporter_voices`, `hourglass` | `TestInterviewEndAndPauseHold` |
| 5 | Brobot reply, live | `gopodAnswerChain` + answer shapes -> pieces -> highlight | `answer_chains`, `answer_shapes`, `colour_slots`, `live_split`, `highlight` | `TestAnswerShapesHold`, `TestHighlightPiecesGolden` |
| 6 | Brobot reply, scripted (interview, draft) | scripted snippets -> pieces | `scripted_snippets`, `scripted_split`, `interview_script`, `interview_build` | `TestScriptedPiecesAreSaidAsSent`, `TestInterviewFromGame` |
| 7 | Duo answers (a line to both) | `gopodDuoRule` -> `duo_a`/`duo_b` chains | `response_chains.duo`, `typed_line` | `TestDuoAnswersHold` |
| 8 | Value point bubbles | `gopodPointBubbles` | `point_bubbles`, `value_points` | `TestPointBubblesShapeAlwaysHolds` |
| 9 | Summary bubbles | `gopodSummaryBubbles` | `summary` | `TestAnswerShapesHold` |
| 10 | Math answers | `brobots_math` chain | `topic_groups`, `math` | `TestMathAnswers`, `TestMathExplainTheMath` |
| 11 | Notice log (inside a Brobot bubble) | `bubbleNotice` | `bubble_notices` | `TestNoticeCountsValuePoints` |
| 12 | Score, grand prize, reset | `gopodCoverageScoreText`, `gopod_game.go` | `value_points.score_text`, `game` | `TestGameWinAndReset` |
| 13 | Weather | weather lines + card | `weather_lines`, `weather_card`, `live_weather` | `TestLiveWeatherWords` |
| 14 | Pictures, pop-ups, eggs | `showPictures` | `topic_pictures` | `TestPopupsOnlyLive`, `TestEggZooms` |
| 15 | LLM warm-up bubble | `buildWarmupBubble`, `showWarmupLevel` | `llm`, `hourglass` | `TestWarmupSequenceGolden`, `TestLLMWarmTemplate` |
| 16 | Narrator lines (songs) | `showNarratorLine` | `chat_window.narrator_*` | `TestNarratorTemplateHolds` |
| 17 | Countdown | `showCountdown` | `countdown` | `TestCountdownAndBingoBallTemplatesHold` |
| 18 | Bingo ball | `showBingoBall` | `bingo_ball` | `TestCountdownAndBingoBallTemplatesHold` |
| 19 | Search results | `showSearchResults` | `system_search` | `TestSearchResultsTemplateHolds` |
| 20 | System menu, terminal | `showTerminal` | `system_menu`, `system_bubble` | `TestSystemMenuTemplateHolds` |
| 21 | QR codes | QR pictures | `qr_codes` | `TestQRCodeTemplateHolds` |

The minimum set it folds into (the aim): **(A) the bubble** - header + placing + baton (rows 1-2);
**(B) the voiced line** - Generating + hourglass, then lit word by word (rows 3-4; a Brobot's own
speech, rows 5-6, lights the same way); **(C) the content chain** - an ordered list of parts, each
with its rules (rows 5-21 are parts: point bubbles, summary, notice, score, pictures, weather...).
Folding them together is a later step, on his go.

**RULE: every voiced line gets the voice template**: "Generating..." + the hourglass while its voice is made, then
its words lit word by word as it plays. A line with its own bubble: `gopodSayShown` (`wait:` /
`lit:`). A voice whose words show some other way (warm-up and opener lines, an egg, the
Operator's own line): `gopodSayForShownLine` names the voice and its words (`forwait:` /
`forlit:`) and the page ties them to that line (`voiceForLine`). From the ground up: a voiced line is in the
template from the moment it's drawn, never after - when any line is drawn (a bubble placed, the
warm-up bubble's message), a voice named for its words still waiting in the queue ties to it
right then (`voiceTieAhead`), so it never shows solid, vanishes, then lights. A voice reaches the page only
from the template - `TestInterviewEndAndPauseHold` fails the deploy otherwise. The score reads
"Easter Eggs".

**A later bubble never holds an earlier voice** (the interview's end held 4.5
min): a voice waits only for bubbles that came before it (their feed ids) - the grand prize,
waiting for "Interview complete" to be said, no longer holds that line's sting and voice. The
Operator's line (and the interview's questions) use the voice template too: "Generating..." and
the hourglass while its voice is made (`op:`), its words lit word by word as it plays (`oplit:`).
A run pauses only once the page has stayed hidden `session.pause_after_ms` (2 s) - a tab
flicker says nothing. Locked by `TestInterviewEndAndPauseHold`.

**Words and waits live in the settings - enforced**: the weather
lead lines (`chain_prompts.weather_lead`/`_next`), the host persona sentences
(`host_persona_frames`), the plain words a repeat is judged by (`brother_repeat.plain_words`),
the session memory lines (`session_memory`), the interview's reporters and default question
(`interview_build.reporters`/`default_ask`), and the chat page's waits and sizes (all in `chat_window`, F5). `TestWordsLiveInTheSettings` fails the deploy if any of these words come back into code.
Left in code on purpose: the system-command help/answer text (the tool's own words) and the
floors used only when a file can't be read.

**Every voiced line, one template - enforced**: `gopodReporterSay(role, text, true,
wait)` shows the bubble, makes the voice (Generating and the hourglass), then lights its words.
The voice-only path (`gopodSystemSay`, `gopodSystemSayNoSting`, show false) is allowed only where
another template already shows that bubble: a song's system_say_role line, an egg's pop-up, the
LLM warm-up's and the opener's warm-up bubble lines, the Operator's own line.
`TestEveryVoicedLineUsesTheTemplate` fails the deploy on any other. Moved onto the template: draft
started, Record Session's confirm, the sign-off line, "You're welcome", the replay's chapter. A
chapter is Goverlord's voiced line ("Chapter:
What GOPOD is") - Generating and the hourglass while its voice is made, then its words lit
(`gopodSayShown`), like every made voice. **No lit line holds the chat for good:** a line lit any
way gets its lit time when first seen and stops holding the next bubble after
`chat_window.speaking_max_ms`; a spoken notice line ("Thanks for your help, ...") is found by
its Brobot's class when "said" arrives; a bubble waiting 10 s+ on the baton logs what holds it
(`GOPOD_CHECK baton held`). (Live: the interview's end froze ~57 s on a held line; the closer
waited 30 s for "Interview complete" to be heard.)

**The notice's new lines**: `bubble_notices.value_points` -
`none` ends "then back on track:", `recovery` is the topics alone on the next line, `score` is three
lines, each "label\tbar" (`{topics_bar}` ..., `value_points.score_bar`: "=" per found, "-" per to go;
`gopodCoverageScoreWith`). The page shows a tab line as a score row - label column, bar in even-width
marks - so the three bars line up (`noticeLine`, `.notice-score-row`). The one-line
`value_points.score_text` stays for the score bubble and the prompt.

**NEW RULE: a revealed egg's marquee stays**: the pop-up shows once (first reveal); every Brobot reply whose
line hits an egg - first time or later - wears the marquee lights. Go posts `eggmark:<title>` to
him when the egg line goes in his notice (`gopodNoticeEgg`); the page's `eggMarked` dresses his reply.
Checked by the page replay's EGG MARQUEE.

**A cue word's "**" is a cue mark** ("veryHappy**
Shmengy ..." shown): after the opening cue word, "**" is read like "___" - the cue plays, the bubble
hides it; a word that's no cue keeps the word and loses the "**" (`TestCueSlotLeaks`).

**Live Brobot highlight timing, fitted to the robots**: the reporters light at their voices'
own word times; a Brobot's piece is an estimate, now fitted from 184 pieces in the day's log -
`word_reveal.seconds_per_word` 0.33 (was 0.38), and a piece said after an animation (it starts with
`animation_lead_in`) lights its first word at `word_reveal.lead_in_start_seconds` (1.8; plain pieces
keep `start_delay_words`). A voice that starts 0.8 s+ after its turn logs "voice late ... bubble X s,
file Y s".

**Goverlord's and the managers' lines are 75% of the header**, like a
Brobot's answer; not the terminal's results (50%) or a notice's lines.

**CSS and JS in their own files**:
the Weather Kit page's look and script (`css/gopod_weather_kit.css`, `js/gopod_weather_kit.js`);
a saved session page's look (`css/gopod_session_page.css`, put inside the page so it stands
alone; `TestSessionPageCarriesItsCSS`). Left in code: the chat's styles set while it runs
(positions, sizes, fades - computed each time). Wire-Pod's own pages' inline styles are native,
untouched.

**Point bubbles** (`point_bubbles`, `gopod_point_bubbles.go`): a value point with notes
becomes his thought bubbles - a hype line for its title, then one party line per note. The model
is asked for exactly that; a missing line, or one sharing no word with its note, is made from the
note ("{Note}."), so the shape always holds. His weather comment, when there is one, stays first.
Locked by `TestPointBubblesShapeAlwaysHolds`.

**The 5 bubble kinds** (`bubble_kinds`, page `tagBubbleKind`): Goverlord
orchestrates, gomads think, gominions do. Every bubble in the chat gets one kind as `data-tier`
- operator (input), goverlord (system), gomad (CHALK, PLAYHEAD - inside a Brobot's notice unless
called directly), gominion (Brobot 1 and 2 and their output), other (eggs, pictures, chapters,
countdown, narrator) - by its classes, the first rule that matches. A bubble with no kind is
logged once: "GOPOD_CHECK bubble without a kind (loose chain)".

**The bubble colour template**: two colours per bubble - its background, and a thin border
(a 2px ring, so nothing moves), all as tokens in one `:root` block (`gopod_chat_bubbles.css`).
Brobots: Wire-Pod's own green border (`#33ED6D`), their eye-colour backgrounds. Goverlord: grey,
traffic-light green border; an error bubble (a numpad error text) traffic-light red (`.is-error`).
Gomads: CHALK chalkboard / chalk white, PLAYHEAD tape brown / tape silver, Git Repo GitHub dark /
GitHub blue, Cache PYC deep neon gold / deep purple (white text), Cameo deep neon pink / neon cyan, Intel 404
PNFH and PNFT police blue / police-light red.

**The here-and-now note** (`here_now`, `gopod_here_now.go`): worked out in code, never
by the model - every Gomad's prompt and every Brobot's question gets "Time is [time]. You are [name]
([gender]), located in [place]. You're chatting with friends: [everyone else, each name (gender,
role)]. Check recent chat history for context." Place: the live one, else `here_now.place`
(Windsor Ontario Canada). Gender: `here_now.genders` (the Brobots and Brobot 0), else the assigned
voice (Kokoro `af_` female, `am_` male). To a Brobot his brother is "your brother", never named.
The battery guards also get the Brobots' real battery readings.
A Brobot gets the same pieces as tokens in his pre-built context ("BEFORE I ANSWER:",
`prebuild_context.lines`: `{time}`, `{gender}`, `{place}`, `{cast}`, and `{plan}` - where this
answer sits in the response chain), one block in the notice bubble's order, not a second message.
A Gomad gets the recent chat as other people's lines (`here_now.chat_note`, `chat_answer`), never
as its own replies, so it doesn't repeat the Brobots. Names others call a host by: `call_name`
(`host_personas`) - the hand-over ("Cache PYC will help you with that.") and the thank-you
(`helper_from` call_name: "Thanks for your help, Cache PYC."). Weather: PLAYHEAD hands it to
Cache PYC (`via`), the Brobot thanks her and says it. Every numpad line (KP0-KP9): header 100%,
its status lines 50%, the words 75% (`.numpad-line`).

**Notes on one line**: a value point's notes show in its notice on one line under its title,
`[good Afternoon|classroom|kids shout wordplay at us|we respond back]`
(`bubble_notices.value_points.notes`, `notes_join`).

**The notice guide** (`bubble_notices.value_points.guide`): each notice line an answer shows comes
with what the Brobot is told about it, in his pre-built context (`{notice}`) - one source, so what
the room sees is what he's told: "Next up: ... - don't reveal these yet.", "Back on track with: ...
- pick one", a summary, web search, an egg, a combo, a cleared chapter, no value point.

**Years shown as said** (`spoken_symbols.show_as_said`: years): a Brobot's bubble shows a date's
year the way he says it ("October 27, 20 26"), so the words and the highlight follow what's heard.

**Score bars**: the score's three rows (Topics, Value Points, Easter Eggs) are thin filled bars that
stretch to the bubble's width - the text marks (`value_points.score_bar`) set the fill; each
Gomad's icon eyes are its border colour.

**GOLDEN: the Brobot Bubble template** (Brobot 1 and 2 alike):

```
===== Brobot Bubble =====
header: Brobot N: (+ zoomable weather picture, top right)
=== notice: Value Points ===
1. Got it! (said by him - the communication received)
2. <topic> #<n>: <point> / No value point - a quick take, then back to GOPOD.
   a point with notes (its topic's "notes N: title | note | note"): <topic> #<n>: <title>,
   then -<note> lines; the model gets the notes and parties on the idea - no exact line is said
   for it ("only the model's flavoured version of the notes"); exact math
   (fractions) and the Brobots' instructions keep their own words
   (Web search: <search>; Easter Egg '<title>': <summary> - one notice per bubble)
   (game: Combo: <topics> / Chapter cleared: <chapter> / Back on track: <topics> (free
   answer) or Next up: <topics> (topic answer) / Value Points <said> / <total>)
3. (optional) [icon] PLAYHEAD [icon] I'll help you with that. (one row; his searching waits)
4. Thanks for your help, PLAYHEAD. (said by him - the floor back to him after a manager's part)
=== end notice ===
<his answer-chained reply> (a weather question: it opens with the golden weather line)
===== Brobot Bubble =====
```

- The order is one settings template, `brobot_bubble.steps` (header, got_it, value_points,
  manager, thanks, reply); manager and thanks can be left out.
- Line 3 (PLAYHEAD came only after the whole answer was written, so the
  Brobot searched, PLAYHEAD spoke, the Brobot searched again): a manager steps in right after
  lines 1-2, while the answer is written (`gopodManagerStepIn`, once per line), its lines in
  this Brobot's notice (`gopodManagerDoFor`). The Brobot waits until the manager is done, even
  with his answer ready.
- Line 1 and 2 are posted together once his "Got it!" is said and his prompt is made
  (`gopodValuePointsNoticePost` / `gopodValuePointsNoticeGotIt`). Line 3 is a manager's top part,
  put in the notice by the page (`nestBox`, `.in-notice`). Line 4: `got_it_again`, "Thanks for
  your help, {helper}." ({helper} from `helper_from`: the manager's name, PLAYHEAD / CHALK; else
  its Brobot, Brobot 4 / Brobot 3),
  said and shown only after a manager's part; `second_got_it` `after_manager` or `off`.
- A manager's part in the notice is one plain line, no bubble of its own: "[icon] PLAYHEAD [icon]
  I'll help you with that." (`.in-notice`).
- A manager's line uses the Generating template like every voice: "Voice file generating." then
  the hourglass while its voice is made (once - "I'll help you with that." is reused, so it
  skips straight to its words), then its words lit to its voice.
- Weather: PLAYHEAD says only `take_text` (`brobot_says`); no weather voice is made; the
  Brobot's reply opens with the golden weather line. Its picture is pinned for later
  (`weather_card.picture: false`: no picture, no pause). Its sound plays once as his bubble shows (a
  background effect - he doesn't wait for it). Once he has said the line, its picture shows
  centred under it (Wire-Pod's `weathershow:`, page `weatherShowFor`) and he waits
  `weather_card.pause_seconds` (3) before the rest - the template for any later "wait for a
  picture or sound" spot (`gopodWeatherLineSaid`). After its sound has played once, a click that zooms the picture plays it again -
  never two at once. Hovering still zooms it without sound.

**Notices inside a Brobot's bubble**: What goes in each is one template,
`bubble_notices`: a title, then its lines, each on its own row; a notice with the same title takes
more lines underneath. `value_points` - every live Brobot answer opens with it: the value points
his answer is built from (`{id}: {words}`, id = topic#n), then weather, search, or "none", then
the "Got it!" line once he has said it (`chat_moves.thinking_text`; in the menu under 3).
`egg` - a found egg's summary. `native_intent` - a native Wire-Pod intent.
(`gopod_value_points_notice.go`.)

---

[Back to ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md)

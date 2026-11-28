# The Brobots chat: ways in, a live reply, the page

Part of [ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md) - [Concepts](concepts/README.md) · [Songs](02_SONGS.md) · [The session and runs](03_SESSION.md) · [The Brobots chat](04_CHAT.md) · [Voices and sound](05_VOICES_AND_SOUND.md) · [The cast and the game](06_CAST_AND_GAME.md) · [The system menu, your stuff, the settings](07_SYSTEM.md) · [Golden rules, troubleshooting, tests](08_RULES_AND_TESTS.md) · [The gears](09_GEARS.md)

<details><summary>Game card: the one door</summary>
<!-- gopod topic: 09_the_one_door -->

# the one door
- **chapter:** Show and tell

- **words:** the one door, ways in, answer chain, thought bubbles, live reply
- **ask:** How does a line reach the Brobots?
- **summary:** every line meets in one door | a live reply is built from the answer chain | every bubble on one template
- **source:** alias_play_studio/04_CHAT.md
- **ask 1:** Brobot 1, tell me about the one door
- **ask 2:** Brobot 2, how is a live reply built - the answer chain?
- **ask 3:** Brobot 1, thought bubbles - is every bubble built on one template?

---

The one door. Reveal these, in order, in your own words:
1. Every line meets in one door - typed, numpad, mic or said to the robot himself - checked in one order.
2. A live reply is built from the answer chain - a list in the settings, not code order.
3. Every bubble is built on one template, made whole before it shows.

</details>

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
- Who each key reaches: the numpad map ([`numpad-mapper.md`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/company/99_system/numpad-mapper.md)). **One place decides a key**: Wire-Pod's
  `gopodNumpadBrobot` - a Brobot, a voice persona, or not live - sent to the page as the map's
  `targets`; the page reads only that (no lookups of its own). `TestGopodNumpadTargetsOnePlace`.
- **Numpad voices**: system menu, Edit words and settings, `system-voices` (page
  `voices`; typed, the list as text). One row per key from Wire-Pod (`gopod_voice_map.go`,
  `/api-sdk/gopod_voice_map`), then the Operator: KP0 and a key whose body is in
  `numpad_ptt.body_to_brobot` say "no voice required" (KP3-KP6 switch by themselves once their
  robots are mapped); the rest pick from every Kokoro voice on disk (the `.pt` files). A voice used
  twice is bold. A pick is saved (`reporter_voices.hosts` / `operator_voice`, undo-able) and used
  from the next line - no restart. The Operator's "unassigned" = Cameo's voice. `TestVoiceMapRows`.
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
  Popular files (`popular_uses`, 4 plays or more) keep `popular_keep_days` - from the report's jump at 4 plays; retune once a month of plays is in.
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
- KP+ and KP- turn the
  camera feeds ([the Edge TPU](07_SYSTEM.md#the-rules-the-system-keeps)). The **/** key skips ([skip](03_SESSION.md#stop-skip-replay)). **\*** held is all the
  Brobots, in the order the All menu picks; **KP0** held is the Butler, KEYNOTE VALE ("Keys") (3 quick taps still
  leave numpad mode).
- Three taps of 0 (or NumLock) leave numpad mode; Goverlord says "Goverlord has released the
  numpad."
- **Every Operator line** - typed or the numpad's, to anyone - is one template: its header 100%,
  its words 75% of it.
- **Its status bubble is an Operator bubble,** headed "Numpad to <who>": the status lines
  (`numpad_ptt.texts`: "KP7 to CHALK", keep holding…, Recording for CHALK, Got it -
  transcribing…), then the words, with "Generating..." and the hourglass while they're read
  aloud. Status lines are never spoken. While the system menu is open, the status is a line in
  the menu's results instead (no other bubble; the error sound still plays).
- **Sounds:** a key that works plays nothing (`numpad_ptt.key_sound` off). Each error message
  (`numpad_ptt.error_texts`: too short, nothing heard, no mic, not ready, failed, NumLock off,
  not) plays grunge at 15% (`numpad_ptt.error_sound`). A key that isn't live (KP9, Cameo)
  is an error too.
- While recording, numpad presses are saved in the session folder ([Record Session](03_SESSION.md#record-session-the-record-box)).

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

The commands and their words: [`native-cube.md`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/behaviours/native-cube.md).

`system-brobot cube <name>` sends the same intents from the menu.

### 7.6 Face recognition (a starter)

Say "Hey Vector, remember my face": the custom intent `REMEMBER_MY_FACE` (`customIntents.json`)
hands the robot Wire-Pod's own "my name is CrushN8r", and the robot scans and saves the face
itself; "who am I" works as before. Voice only (a typed line goes to the chat). Needs `wpr 2`
after an edit. The face photo as a prize is not built yet. `system-faces` lists faces.

### 7.7 Who answers: addressing and All

- **The picker, the key and the header match.** A command line ("start fart joke", "stop run")
  sent with KEYS picked shows "Keyboard to KEYS" - the header names whoever the numpad map puts on
  that key, never a hard-coded "Brobots" (`TestCommandLineNamesTheKey`). The * key's one name
  (`numpad_ptt.all_label`, the numpad map card) is what the picker says too: "All Brobots 1,2",
  "All Brobots 2,1".
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
the nicknames ([`nicknames.md`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/behaviours/nicknames.md)), never twice running); without one, `addressing.redirect_text`. Empty = none.

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
| `colour:value_point` | stitch | a value point's exact line (its say line) first, word for word; the model's part after it is its colour, cut to `colour_slots.max_sentences`, even with `budget_off` |
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
4. The reply: each animation sent, the speech `chat_moves.speech_after_animation_seconds` later, the next animation only once the
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
- **No scaffolding:** a tag like `<thinking>` or a working label ("emoji:", "note:") is the
  model talking to itself - it never reaches a bubble (`gopodStripLeaks`, `TestLeaksStripped`).
- **Nonsense gets nonsense:** a line with no real words and no card its words pick
  (a keyboard mash) isn't played along with. Whoever's asked says "What is '[the mash]'? I
  don't know - let's start at README.md." - the words are in
  Goverlord's card (`nonsense_line`, `real_words`); `TestNonsenseGetsReadme`.
- **Song/system lines said by a Brobot** (`gopodRecordSay`, for example the photo's second "I
  see ...") go through `GopodSpeakSentences`: lit while said, shown as said after.

### 8.3 Said right: pronunciation and the three lingos

There are three lingos. A pronunciation goes only to the lingo it is for.

| Lingo | Who | Where |
|---|---|---|
| **Brobots** (party lingo, the default for a new pronunciation) | Brobot 1 and 2 | the pronunciation registry in `brobots_wirepod_interview_section_card_template_1_001.md` |
| **Reporters** | CHALK, PLAYHEAD and the Gomads | `reporter_voices.say_words` |
| **Butler** | Goverlord | `system_say_words` |

- **Brobot lingo, word starts:** an entry can be `"word_start": true` - at the start of a made-up
  word, the rest kept ("chocoyum" -> "chalk-o-yum", "chocolateric" -> "chalk-o-latteric"); a
  dictionary word (chocolate) and its `except` words are left alone. Both Wire-Pod and the song
  engine. `TestBrobotLingoChoco`.
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
- In the live chat it is an egg ([eggs](06_CAST_AND_GAME.md#eggs)): Goverlord's congratulations with
  `math.popup_title` over the challenge, Brobot 1's answer under it; Brobot 2's follow-up in his
  own marquee bubble. Its other eggs show too (Brobot 1's).
- Locked by `TestMathExplainTheMath`.

### 8.6 Weather

- **Weather only:** a weather question gets the weather and nothing riding along - no booking
  point because the line said "party" or "today".
- **The take:** "weather" in a line goes to PLAYHEAD ([managers](06_CAST_AND_GAME.md#managers)):
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
- **The line itself:** the place, the date, both units and both clocks, each Brobot his own first;
  military time as four digits ("zero" for 00-09, "hundred" on the hour). Reporters follow the
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
2. **Photo wake:** after control, he waits `photo.awake_seconds` - awake, head settled.
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
`~/gopod_searxng/`).

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
so the chat never stops until F5; a line that hasn't handed on after `page_sync.item_max_ms`
is logged with what holds it and passed, and after an F5 the page tells Wire-Pod how far it is.
The hourglass drains once it's on the page (it's made before it's placed). A Brobot's answer
under his header is 75% of the header's size, with a little air between characters. The exceptions keep their own rules: the chat window (scrolling, the baton),
the Operator bubble (the one input type) and the system window (the menu). On the template so far:
system / terminal answers, reporter and manager lines (their own bubbles), Goverlord's lines, Brobot replies.

### 9.1 The Chat Bubbles Window: the machine


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
| Manager | CHALK's chalkboard, PLAYHEAD's tape-brown (the colour of vintage video tape cassettes) | see [managers](06_CAST_AND_GAME.md#managers) |
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
  then its length and its word times (`.words.json`) loaded (up to `VOICE_READY_MAX_MS`),
  then it plays and lights - never before its word times arrive.
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
   `voiceLitWaitMs`; a Brobot that stops speaking leaves the rest of his bubble shown as said.
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
- Every word is its own step (`word_reveal.seconds_per_word`). A word's punctuation, emojis
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
  the warm-up, the eggs, Goverlord, the Operator and the Brobots all come through it.
- **A said line, replayed** (`replaySaidLine`, the template): a line he already said (his opener,
  said while thinking) shows when its bubble shows - from 0%, word by word, then said - and holds
  the next line until done. It first waits `word_reveal.replay_settle_seconds` at 0% so its bubble
  settles into place. His opener ("Got it!" / "Nice try, Mr/Mrs ...!") is the first.
  LOCKED by `TestOpenerMechanismLocked`.
  It is not one of his reply's pieces (`ownSays` leaves it out): Wire-Pod names the piece it says
  by number, written just before each say goes to the robot (`gopodWriteSpeakingPiece`), and the
  page lights that number.
- While a link's words are revealed, only the shown words are underlined.

**Forced alignment: word times from the voice - GOLDEN, the one highlight gear.**
The voice worker has Vosk listen to each voice it makes (`align_heard`), and the page lines the
shown words up with those times (`alignUnits`) - both by the same one rule: the words said as
written are matched in order (the longest common run - no look-ahead window, no partial
matches); every other word ("404" said four oh four, "PNFT" said P N F T, "2026" in parts) and
every mark or emoji shares the time between the matched words around it, by its length. A highlight
glitch is looked for here, in this rule, never patched word by word. `"aligner": "vosk"` in
`<name>.words.json`, saved beside each voice; a voice without one is made again on its next use.

**One reveal, any source of times.** The word reveal (`digitSequence` + `alignUnits`) uses word
times whenever a piece carries them, whatever made them - a Kokoro voice's `.words.json`, or a
song's `narrator_words.json`.
- **A song's narrator:**
  `gopod_probes/tools/gopod_align_captions_001.py <narrator audio> <song folder>` - Vosk (Wire-Pod's own speech model) times the words it hears; each SRT caption's
  words are matched in order within its seconds; saved as `narrator_words.json` (Kokoro's shape);
  the audio isn't kept. The song engine puts each caption's times on its piece
  (`@s-e ~<base64> words`).
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

- **QR codes:** a bubble whose words name a QR's words gets that QR (`qr_codes.codes`); CHALK shows them ([managers](06_CAST_AND_GAME.md#managers)).
  In a corner, QR1 sits top-right in the header row; scrolled above the chat's top it jumps to
  the bubble's bottom-right (`corner-low`, marquee inset kept). QR2 (`qr_codes.second`) goes
  right after "...subscribing to our newsletter!", the last step of the reveal. The magnifier as
  ever. `qr_codes` in the settings folder (`webroot/gopod_settings/`) (`qr_codes.codes`: the words that show each),
  pictures in `webroot/qr/`.
- **Weather faces:** a bubble holding the live weather line gets Vector's own weather face for it
  (sunny, cloudy, rain, snow, thunderstorms, windy; clear at night = stars; cold has none),
  animated, Vector's own size, in the speaking Brobot's eye colour, in the same
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
  speaks once his bubble is on the page - up to `page_sync.reply_wait_seconds` - so he never
  talks over a line the page is still playing (`gopodWaitReplyShown`; `GOPOD_SYNC` lines when he
  waited or the page was behind). (2) The page tells how long it held each line before showing
  it; over `page_sync.held_log_seconds` it's in the journal (`GOPOD_PAGE held <id>`). The
  page still keeps its drawing order (one bubble after the one above, the countdown, the SRT
  narrator's soundtrack clock); the engine waits on those instead of running ahead of them.
- **Every wait on a line:** Wire-Pod goes on the moment the page says it has played it (the heard
  check). With no answer, after the queued lines' own lengths plus
  `reporter_voices.heard_margin_seconds`. No page-lag padding is added.
- **Wait until I'm finished:** a voiced line waits until the page has played it - its own length
  plus up to `reporter_voices.finish_wait_seconds` - while the chat page is watched and
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

- `background_loop` plays on the page over and over with no
  gap, from its top again at every session clock reset to 00:00:00.
- The speaker icon after the text field (🔊/🔇) turns it on or off (and opens or closes the
  system menu). Every new Wire-Pod start sets it off.
- Hovering the icon shows a vertical volume bar (0-100%, the loop's share). Let go, its value is
  saved to `background_loop.volume`, so it holds over restarts. The slider gives way when the
  settings' loop volume changes.
- Every run quiets the loop at its start and, as its very last line, sets it back to the toggle.

## More cool parts

**GOPOD as a phone app.** Wire-Pod serves GOPOD securely too, with GOPOD's own certificate (the
phone trusts it once); the app opens straight into the chat.

**Visitors.** Anyone on the local network can join: their first line asks for a username, the admin
approves it in the system window, and with participation on they talk to the Brobots. A visitor
reaches the chat and nothing else - Wire-Pod's settings and logs are refused - and never runs the show.

**One page owns the show.** GOPOD's own machine has the main page: it plays the sound and times the
run. Every other page is a silent mirror of it, bubble by bubble, highlights and all, with its own
scroll.

**One way in, one feed, one bubble.** Every Operator line goes through one door; everything the page
shows comes from Wire-Pod's one numbered feed, in order; every bubble is built one way, hidden
until whole.

**Every voiced line, one template.** "Generating..." and the hourglass while its voice is on its
way, then its words lit word by word as it plays - every page voice, a voice already made too. A
voice names its line ahead, so it never searches the chat for it. A later bubble never holds an
earlier voice, and no lit line holds the chat past `chat_window.speaking_max_ms`.

**The Brobot Bubble.** His header; a notice (Got it!, the value points his answer is built from, a
manager's part, "Thanks for your help"); then his answer-chained reply. The order is one settings
list, `brobot_bubble.steps`. A manager steps in while his answer is written, and he waits for her.

**Point bubbles.** A value point with notes becomes his thought bubbles - a hype line for its title,
a party line per note; a line the model misses is made from the note, so the shape always holds.

**The five bubble kinds.** Operator (input), Goverlord (system), Gomad, Gominion, other - every
bubble gets one; one with none is logged as a loose chain.

**The bubble colour template.** Two colours per bubble - background and a thin border - all tokens in
one `:root` block.

**The here-and-now note.** Worked out in code, never by the model: the time, who you are, where you
are, who you're chatting with.

**Score bars.** Topics, value points and eggs as three bars that fill as the game goes.

**Years shown as said** ("20 26"), so the words and the highlight follow what's heard.

**The page replay** (`gopod_page_replay.sh`): the real chat page with no robots - every egg, both
prizes, a visitor's mirror beside it - reports anything stuck, unlit or out of step.

[Back to ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md)

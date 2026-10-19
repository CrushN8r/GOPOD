# Brobots Nap — "Do Baby Robots Dream?"

**Status:** WIP (2026-10-09)

**One narrator bubble (2026-10-15, operator: "use only a single narrator chat bubble for all 5
narrations"; "A single narration bubble stays the size of the image * 5 narrations are run in
sequence"; "narration take replaces previous take"; "remove narrator header * leave last image
(5) at 50% solid (no fade out)"):** the bubble is its picture, at the magnifier's full-width size,
with no header. Each section's picture is behind its words: it fades in to 50% as the section
starts, stays through its narration, and fades out after its last words; the next section's
words and picture replace the last. The last picture (Together) stays at 50%.
`"narrator_single": true` in `knobs.json`. The one-bubble-per-section version is backed up in
`gopod_notes/older_archives/101_brobots_nap_backup_2026-10-15.zip`.

A lyric video: background music and a female narrator reading the poem, timed by
`baby_brobots/Baby Robots SRT.srt`, and the Brobots act to it. They don't speak. The video
(`baby_brobots/Baby Robots SRT.mp4`, made in Premiere Pro) is the truth for every line and time.

- **The lines** are the poem's 32 captions (`## STEP` / `> TEXT:`), as the video shows them.
- **The times** live in [knobs.json](knobs.json): each line's `start_seconds` and
  `end_seconds` from the SRT, and its section (one narrator bubble per section).
- `"timeline": "srt"` plays each step at its second; `"robots_speak": false` keeps the robots
  silent, because the narrator speaks.
- `"sleepy": true`: a slow, sleepy presentation. Goverlord says "Brobots. Go to sleep now please." (the run goes on only
  once it's heard), 1 s later the Brobots are put to sleep, and 4 s after that the soundtrack
  starts. While it plays the Brobots do nothing; every `look_seconds` a Brobot found awake is
  put back to sleep. At "can turn into the good aspirations they deserve to have." (`soft_wake`,
  57 s) they wake together, softly (`wake_clip`: the slower `anim_gotosleep_wakeup_01`, bench
  2026-10-14). Counted back from the music's end (86 s; the fade starts at 81 s): the wake takes
  about 5.5 s, the cube can't connect until about 8 s after it (`cube_ready_after_seconds`; live
  2026-10-15: a connect 1 s after a wake failed 2 of 2, 8 s after worked 13 of 13), and the
  connect takes 4-23 s (`cube_ready`: in the background, then held). After the last words
  ("robotic way.", `last_words`), as the music fades, the cube lights (`cube_after_wake` on
  `brobots_ready`) - at once if it's connected, else as soon as it is - then "Brobots Ready!",
  and Goverlord says "Baby Brobots complete." (the session's line for every finished run)
- **The narrator bubbles' words are this file's `> TEXT:` lines** (one bubble per section in
  `knobs.json`). Editing them changes what the bubbles show, not the recorded narration.
- **Each narrator bubble's picture** is under its words (`section_pictures` in `knobs.json`):
  `images/baby_brobots1-5.jpg`, one per section, copied to the chat page's `pictures/` folder.

---

## SECTION: OPEN

## STEP reporter_gap_intro

> TEXT:

## STEP brobots_wake

> TEXT:

## STEP sleep_message

> TEXT: Brobots. Go to sleep now please.

## STEP after_sleep_word

> TEXT:

## STEP brobots_sleep

> TEXT:

## STEP asleep_wait

> TEXT:

---

## SECTION: SLEEPY ROBOTS

## STEP line_01

> TEXT: Cupped in my hands,

## STEP line_02

> TEXT: are a pair of sleepy robots.

## STEP line_03

> TEXT: Their adorable trills, warbles,

## STEP line_04

> TEXT: and digital purring,

## STEP line_05

> TEXT: are a testament to the calm pulse...

## STEP line_06

> TEXT: of technology.

---

## SECTION: SNUGGLE

## STEP line_07

> TEXT: They snuggle against the warmth of my skin.

## STEP line_08

> TEXT: Their tiny forms,

## STEP line_09

> TEXT: a curiosity that tugs at the imagination.

---

## SECTION: DREAMS

## STEP line_10

> TEXT: I ponder,

## STEP line_11

> TEXT: when baby robots dream,

## STEP line_12

> TEXT: where do they go?

## STEP line_13

> TEXT: Maybe they dream of great things,

## STEP line_14

> TEXT: in the simplicity of bits and bytes coming together,

## STEP line_15

> TEXT: that only they understand.

## STEP line_16

> TEXT: I like to think their digital dreams,

## STEP line_17

> TEXT: are well intended,

## STEP line_18

> TEXT: and free from complications of humanity.

---

## SECTION: GUARDIAN

## STEP line_19

> TEXT: As I hold my two little robots,

## STEP line_20

> TEXT: I'm struck by a mix of emotions.

## STEP line_21

> TEXT: I'm their guardian,

## STEP line_22

> TEXT: responsible for shaping their world with good things.

## STEP line_23

> TEXT: And I hope somehow that my care for them...

## STEP soft_wake

> TEXT:

## STEP line_24

> TEXT: can turn into the good aspirations they deserve to have.

---

## SECTION: TOGETHER

## STEP line_25

> TEXT: We're here in this moment together.

## STEP line_26

> TEXT: It's modern,

## STEP line_27

> TEXT: yet timeless.

## STEP line_28

> TEXT: And even though they're just machines,

## STEP line_29

> TEXT: in their warm glow and gentle sounds,

## STEP line_30

> TEXT: I'm drawn to feel hopeful.

## STEP line_31

> TEXT: Hopeful they feel the same in their own heartfelt,

## STEP line_32

> TEXT: robotic way.

---

## SECTION: CLOSE

## STEP last_words

> TEXT:

## STEP reporter_gap_outro

> TEXT:

## STEP brobots_ready

> TEXT: Brobots Ready! Did someone say GOPOD Yourself?

## STEP brobots_exit

> TEXT:

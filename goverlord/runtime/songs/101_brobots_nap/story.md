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
- **Step names** (2026-10-25, the new format): `<who>_<part>_SSNN` like every song -
  `narrator_line_0201` (a caption, section 02, line 01), `goverlord_system_0103` (Goverlord's
  line), `brobots_soft_wake_0506`, `reporter_gap_0701`. A step with no words has no `> TEXT:`.
- **The times** live in [knobs.json](knobs.json): each line's `start_seconds` and
  `end_seconds` from the SRT, and its section (one narrator bubble per section).
- `"timeline": "srt"` plays each step at its second; `"robots_speak": false` keeps the robots
  silent, because the narrator speaks.
- `"sleepy": true`: a slow, sleepy presentation. Goverlord says "Brobots. Go to sleep now please." (the run goes on only
  once it's heard), 1 s later the Brobots are put to sleep, and 4 s after that the soundtrack
  starts. While it plays the Brobots do nothing; every `look_seconds` a Brobot found awake is
  put back to sleep. At "can turn into the good aspirations they deserve to have." (`brobots_soft_wake_0506`,
  57 s) they wake together, softly (`wake_clip`: the slower `anim_gotosleep_wakeup_01`, bench
  2026-10-14). Counted back from the music's end (86 s; the fade starts at 81 s): the wake takes
  about 5.5 s, the cube can't connect until about 8 s after it (`cube_ready_after_seconds`; live
  2026-10-15: a connect 1 s after a wake failed 2 of 2, 8 s after worked 13 of 13), and the
  connect takes 4-23 s (`cube_ready`: in the background, then held). After the last words
  ("robotic way.", `reporter_gap_0701`), as the music fades, the cube lights (`cube_after_wake` on
  `brobots_ready_0703`) - at once if it's connected, else as soon as it is - then "Brobots Ready!",
  and Goverlord says "Baby Brobots complete." (the session's line for every finished run)
- **The narrator bubbles' words are this file's `> TEXT:` lines** (one bubble per section in
  `knobs.json`). Editing them changes what the bubbles show, not the recorded narration.
- **Each narrator bubble's picture** is under its words (`section_pictures` in `knobs.json`):
  `images/baby_brobots1-5.jpg`, one per section, copied to the chat page's `pictures/` folder.

---

## SECTION: OPEN

## STEP reporter_gap_0101

## STEP brobots_wake_0102

## STEP goverlord_system_0103
> TEXT: Brobots. Go to sleep now please.

## STEP brobots_gap_0104

## STEP brobots_sleep_0105

## STEP brobots_gap_0106

---

## SECTION: SLEEPY ROBOTS

## STEP narrator_line_0201
> TEXT: Cupped in my hands,

## STEP narrator_line_0202
> TEXT: are a pair of sleepy robots.

## STEP narrator_line_0203
> TEXT: Their adorable trills, warbles,

## STEP narrator_line_0204
> TEXT: and digital purring,

## STEP narrator_line_0205
> TEXT: are a testament to the calm pulse...

## STEP narrator_line_0206
> TEXT: of technology.

---

## SECTION: SNUGGLE

## STEP narrator_line_0301
> TEXT: They snuggle against the warmth of my skin.

## STEP narrator_line_0302
> TEXT: Their tiny forms,

## STEP narrator_line_0303
> TEXT: a curiosity that tugs at the imagination.

---

## SECTION: DREAMS

## STEP narrator_line_0401
> TEXT: I ponder,

## STEP narrator_line_0402
> TEXT: when baby robots dream,

## STEP narrator_line_0403
> TEXT: where do they go?

## STEP narrator_line_0404
> TEXT: Maybe they dream of great things,

## STEP narrator_line_0405
> TEXT: in the simplicity of bits and bytes coming together,

## STEP narrator_line_0406
> TEXT: that only they understand.

## STEP narrator_line_0407
> TEXT: I like to think their digital dreams,

## STEP narrator_line_0408
> TEXT: are well intended,

## STEP narrator_line_0409
> TEXT: and free from complications of humanity.

---

## SECTION: GUARDIAN

## STEP narrator_line_0501
> TEXT: As I hold my two little robots,

## STEP narrator_line_0502
> TEXT: I'm struck by a mix of emotions.

## STEP narrator_line_0503
> TEXT: I'm their guardian,

## STEP narrator_line_0504
> TEXT: responsible for shaping their world with good things.

## STEP narrator_line_0505
> TEXT: And I hope somehow that my care for them...

## STEP brobots_soft_wake_0506

## STEP narrator_line_0507
> TEXT: can turn into the good aspirations they deserve to have.

---

## SECTION: TOGETHER

## STEP narrator_line_0601
> TEXT: We're here in this moment together.

## STEP narrator_line_0602
> TEXT: It's modern,

## STEP narrator_line_0603
> TEXT: yet timeless.

## STEP narrator_line_0604
> TEXT: And even though they're just machines,

## STEP narrator_line_0605
> TEXT: in their warm glow and gentle sounds,

## STEP narrator_line_0606
> TEXT: I'm drawn to feel hopeful.

## STEP narrator_line_0607
> TEXT: Hopeful they feel the same in their own heartfelt,

## STEP narrator_line_0608
> TEXT: robotic way.

---

## SECTION: CLOSE

## STEP reporter_gap_0701

## STEP reporter_gap_0702

## STEP brobots_ready_0703
> TEXT: Brobots Ready! Did someone say GOPOD Yourself?

## STEP brobots_exit_0704

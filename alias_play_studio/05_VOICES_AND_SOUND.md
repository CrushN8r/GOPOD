# Voices and sound

Part of [ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md) - [Concepts and terms](01_CONCEPTS.md) · [Songs](02_SONGS.md) · [The session and runs](03_SESSION.md) · [The Brobots chat](04_CHAT.md) · [Voices and sound](05_VOICES_AND_SOUND.md) · [The cast and the game](06_CAST_AND_GAME.md) · [The system menu, your stuff, the settings](07_SYSTEM.md) · [Golden rules, troubleshooting, tests](08_RULES_AND_TESTS.md)

## 10. Voices

### 10.1 Who speaks with which voice

| Voice | Kokoro voice | Says |
|---|---|---|
| Reporter 1 (Brobot 3) | `af_heart` | questions, banter, the opening and closing; CHALK's voice |
| Reporter 2 (Brobot 4) | `am_michael` | questions, banter; PLAYHEAD's voice |
| Goverlord (the digital butler) | `bm_george` | system lines |
| Cameo (KP9) | `af_sky` | Cameo's answers; and, with no `operator_voice`, your typed lines read out (`unassigned_voice_host` "CAMEO") -, . Numpad lines still play the mic's own recording. |
| Git Repo (KP3, Cozmo) | `am_liam` | the ethical hacker student |
| Cache PYC (KP4, Cozmo) | `af_nova` | the meme hustler student |
| Intel 404 PNFH (KP5, Moorebot Scout) | `am_puck`, +150 cents (sox) | security guard 1, higher |
| Intel 404 PNFT (KP6, Moorebot Scout) | `am_puck`, -150 cents (sox), speed 0.9 | security guard 2, lower, slower |

- All at `reporter_voices.speed`, unless `voice_tweaks` gives a host its own `speed` / `pitch_cents`:
  the worker runs `sox ...
  pitch` after Kokoro - the length stays, so the word times fit; a tweaked voice gets its own file
  name (`_p+150`). `am_liam` and `af_nova` come from Kokoro's Hugging Face page
 ; `am_michael` stayed PLAYHEAD's only.
- Goverlord's lines are said by the digital butler (`system_persona`, from Master_CrushN8r.pdf)
  with its own phonetics (`system_say_words`). Run lines keep the run name first.
- The page plays voices one after another; a line that must be heard before the run goes on is
  waited for (golden rule 11).
- The Brobots speak with Vector's own voice, on the robots.

### 10.2 Voice files: made once, play many

- At each Wire-Pod start every scripted line (system, reporters, the interview, each song's
  warm-up, the run commands) is made once, in the one voice that says it.
- Each line is made once and kept, named by its voice and speed: a new voice or speed never
  overwrites. Locked by `TestReporterFilesNamedByVoiceAndSpeed`.
- **Voice kinds:** each voice's kind is noted as it's made (`voice_kinds.json`): **system** =
  Goverlord's voice; **scripted** = the start's check or a playing script (reporter, narrator...);
  **live** = the Operator's lines and the managers' live answers.
- **Voice keeping, no auto-deletes:** at each start a voice not in use and not said for its kind's
  days (`reporter_voices.operator_keep_days` for system lines, `scripted_keep_days` for scripted lines,
  `_keep_days` for lines) moves to `sessions/archives/voices/` (the session chats'
  archive). Saying a line again resets its days. Logged as
  `GOPOD_REPORTER_VOICES cleaned up N old voice files`.
- **Rebuild all voices:** `system-voices-rebuild [go]` runs `system-check` first (`story.md`,
  `knobs.json`, `zKnobs.json`, every file named) - a problem pauses it ("go" rebuilds anyway);
  lists the scripted lines that need a voice; moves every generated voice to
  `webroot/gopod_reporter_audio_backup/` (sounds stay); says "Ready for a restart". Then `wpr 2`
  and F5: the start's generation check makes every line again; each voice's backup copy is
  deleted once its new playable voice is made. `gopod_system_voices.go`.

### 10.3 (music) places the sting

A Goverlord line that starts with `(music)` has the sting (`system_stings.default`, dJent3b)
before its words; one that ends with `(music)` has it after. `(music)` is shown (for deaf
viewers), never said. The warm-up's end sting doesn't play again when the ready line's own
`(music)` played it (`gopodMusicMarks`, `gopodMusicPlayed`).

### 10.4 System stings

A Goverlord line with no `(music)` gets a sting before it is spoken: `grunge.wav` for a run's
started, stopped, complete, paused or continues line (and the stop-run answer when nothing runs);
`dJent3b.wav` for every other system line. One sting per bubble (`reporter_voices.system_stings`).
Message 1 of the opener has no sting.

---

## 11. Volume and sound

- **One master:** `chat_audio.volume` (35%). Every sound the chat page plays - voices, stings, a
  song's sounds and soundtrack, weather sounds - plays at it.
- **Every other volume knob** - a sting's (`system_stings`), the numpad error sound, the weather
  sound, the background loop, a song step's `"volume"` (the opener's, the Fart Joke's sounds) - is
  `""` (the master) or a percentage of it (`"40%"`). Turning the master turns them all. A bare
  number (0.4) still reads as 40%.
- **Current shares:** the weather sound 29% and the background loop 14% of the master; `grunge.wav`
  at 40% to sound as loud as `dJent3b.wav`; the Fart Joke's danger sound at 25%; its farts at
  80%; numpad errors at 15%.
- **Evened sounds (loudness):** every sound is measured for average loudness (ffmpeg EBU R128),
  turned to `chat_audio.loudness_lufs` (-16), then limited under `chat_audio.normalize_db` (-3;
  the limiter at -3.5 so the true peak stays in). Made voices are done in place as they're made;
  stings, song sounds, soundtracks and the weather mp3s get done copies in
  `webroot/gopod_audio_norm/` that the page plays (`normSrc`, `webroot/gopod_audio_levels.json` "made",
  `gopod_audio_levels.go`); originals untouched. `normalize_db` 0 = off.
- **Change from the chat:** `system-set master-volume 35%` and the named volume settings
  ([14.3](07_SYSTEM.md#143-editing-from-the-chat)).

## More in this part

**Speech after an animation: 0.4 s**: `chat_moves.speech_after_animation_seconds` 0.9 -> 0.4.

**The egg's voice in pieces** (`gopod_egg_voice.go`): Goverlord's congratulations is said from its pieces -
`topic_pictures.egg_text`'s words around {name} (each line a piece, quotes not said) and
`egg_zoom_text` - each made once and kept like every voice; only the title is made live. They're
trimmed to their own words (`egg_voice_edge_ms` of air) and joined (`egg_voice_gap_ms` between),
word times moved along, so the page lights one line; Generating + hourglass until the joined
voice is ready. `system-test-eggvoice` makes one for real (5 pieces, 8.0 s on);
`system-test-replay` (system menu, and the audit) runs the page replay. only the quotes around a line go; an apostrophe inside it is said
("Granny's", "Don't"). Known catch, noted: an egg name ending in an apostrophe ("Brobots'")
loses it - none do today.

**The mic's own voice** (`gopod_mic_voice.go`): a numpad line
keeps its 48 kHz recording; it's evened like every made voice (`gopodLoudnessFile`), timed by the
voice worker's Vosk (an "align" job), and played through the voice template (`gopodVoicePlay`,
shared with made voices) in place of the Operator's made voice - Generating + hourglass while it's
processed, then lit word by word. Also: the bubble shows the speech's fallback line when that's
what's said (`gopodDisplayForSpeech`, used and by the highlight check); a typed answer to
Goverlord ("Keep the session recording") is shown and read out like every Operator line; a
visitor's chat opens without Wire-Pod's log polling (refused requests logged once a minute); the
Easter egg marquee is each page's own (sent empty, laid out for the visitor's screen); a visitor's
row has a smaller "All" menu with the text field's green and a Submit button. Lingo: doin' =
dew-in. A visitor's bubble text is 75% of each screen's size; on a phone the nap's narrator bubble grows with its words instead of cutting
them. A fallback logs why (BROBOT_FALLBACK: the failing line, the rule, the reply as written).
`TestMicVoiceAndVisitorRowHold`.

**Speaker and system window**: the speaker
starts as `background_loop.starts_on` says (off) on every page load - nothing kept in the browser;
its volume bar shows only while it's on. The system window opens with every slot closed, the menu
as tall as its slots, the results empty below and the bar between them dragging 5%-95%. The admin's
own bubbles (system window, system command answers: `gopod-admin-only`) never reach a visitor.
`TestSpeakerAndSystemWindowHold`.

**GOLDEN bubble sounds: the Operator & Manager bubble template**: `bubble_sounds` - each kind's sound when its bubble shows (`file`, else `default_file`
djent; `volume`; `mute`, all true now); the Operator's `error` label (grunge, 15%) plays even when
muted - the numpad's error messages wear it (`numpad_ptt.error_texts`: too short, not yet ...;
its old `error_sound` moved here). A file not found is skipped. `bubbleSound` in the page.
Goverlord's bubbles too: `bubble_sounds.Goverlord` (djent, muted) is a
Goverlord line's own sound (`gopodSystemSting`); the grunge run sting is gone. A "(music)" and a
warm-up's end still play djent (`reporter_voices.system_stings.default`, `gopodMusicSting`).
Locked by `TestGoverlordSoundMutedMusicStays`. A sting never waits for word times (the log showed
every one starting 3.0 s late).

---

[Back to ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md)

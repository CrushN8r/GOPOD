# Voices and sound

Part of [ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md) - [Concepts](concepts/README.md) · [Songs](02_SONGS.md) · [The session and runs](03_SESSION.md) · [The Brobots chat](04_CHAT.md) · [Voices and sound](05_VOICES_AND_SOUND.md) · [The cast and the game](06_CAST_AND_GAME.md) · [The system menu, your stuff, the settings](07_SYSTEM.md) · [Golden rules, troubleshooting, tests](08_RULES_AND_TESTS.md) · [The gears](09_GEARS.md)

<details><summary>Game card: voices and sound</summary>
<!-- gopod topic: 10_voices_and_sound -->

# voices and sound
- **chapter:** Show and tell

- **words:** voices and sound, the voices, who speaks, master volume
- **ask:** Tell me about the voices and sound
- **summary:** each persona its own voice | one master volume | speech 0.4 s after an animation
- **source:** alias_play_studio/05_VOICES_AND_SOUND.md
- **ask 1:** Brobot 2, tell me about the voices and sound
- **ask 2:** Brobot 1, is there one master volume for every sound?
- **ask 3:** Brobot 2, voices and sound - how soon does speech start after an animation?

---

Voices and sound. Reveal these, in order, in your own words:
1. Each persona speaks with its own voice; a host's robot voice effects sit on top of its pitch.
2. One master volume: every sound the chat page plays - voices, stings, songs, weather - plays at it.
3. Speech starts 0.4 s after an animation.

</details>

## Who speaks with which voice
- Each persona's voice, pitch and robot effect: the settings, [`05_voices_and_robots.json`](../goverlord/wire_pod_overlay/chipper/webroot/gopod_settings/05_voices_and_robots.json)
  (`reporter_voices.hosts`, `voice_tweaks`, `robot_fx_presets`); which key reaches which persona:
  the numpad map.
- The Brobots speak with Vector's own voice, on the robots. Every other voice is made by Kokoro on
  this machine and played by the chat page, one after another.
- A robot effect runs on a copy of the voice and keeps its length, so the word times still fit.

## Voice files: made once, played many times
- At each Wire-Pod start every scripted line is made once, in the voice that says it, and kept,
  named by its voice and speed - a new voice never overwrites.
- Voices not used for their kind's keep days move to `sessions/archives/voices/`; nothing is
  auto-deleted.
- `system-voices-rebuild` checks every file first, then sets every voice aside for a fresh make at
  the next `wpr 2`.
- A numpad line plays the mic's own recording, evened and timed like a made voice.

## (music) and stings
- `(music)` at a line's start plays the sting before the words, at its end after them; it's shown
  (for deaf viewers), never said.
- Each bubble kind's own sound: `bubble_sounds` in the settings.

## Volume
- **One master:** `chat_audio.volume`. Every sound the page plays follows it.
- Every other volume is the master (`""`) or a share of it (`"40%"`): turn the master, they all
  turn.
- Every sound is evened to `chat_audio.loudness_lufs` and kept under `chat_audio.normalize_db`;
  originals stay untouched.
- From the chat: `system-set master-volume 35%`.

[Back to ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md)

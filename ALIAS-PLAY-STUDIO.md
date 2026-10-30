# ALIAS-PLAY-STUDIO

> The reference manual for GOPOD's show: how two Vectors learn a song, how a session and its
> runs work, and how the Brobots chat page behaves, rule by rule. Take Out part 1 in
> [WIRED-POD.md](WIRED-POD.md).

**What this is.** GOPOD's reference for behaviour. Every song rule, every
chat-page rule, every command and setting is written here once, with the code that does it and
the test that locks it. If the code, a song folder and this manual disagree, fix one of them.

**Who it is for.** A Wire-Pod owner reading the repo, the owner running the show, and an AI
assistant working on the code. You do not need to know the code to use it: start with
[Concepts and terms](alias_play_studio/01_CONCEPTS.md).

**Where things live.** Values are in the settings folder (`webroot/gopod_settings/`) (each block's `about` says
what it tunes). Go files are in `goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/`. The chat
page is `webroot/js/gopod_brobots_chat.js` with `gopod_chat_bubbles.css`. Songs are in
`goverlord/runtime/songs/`.

**Tags used here.**
- **GOLDEN** - proven live and locked. Change it only with the owner's OK.
- **Locked by `TestName`** - a test fails the deploy if this behaviour changes.
- **WIP** - being worked on.

---

## Contents

- [Concepts and terms](alias_play_studio/01_CONCEPTS.md) - 1. Concepts and terms
- [Songs: the folder, the notes, making one](alias_play_studio/02_SONGS.md) - 2. A song folder; 3. The notes; 4. The toolbox and the songs; 5. Making and editing a song
- [The session and runs](alias_play_studio/03_SESSION.md) - 6. The session and runs
- [The Brobots chat: ways in, a live reply, the page](alias_play_studio/04_CHAT.md) - 7. Ways in: talking to the Brobots; 8. A live reply; 9. The chat page and its bubbles
- [Voices and sound](alias_play_studio/05_VOICES_AND_SOUND.md) - 10. Voices; 11. Volume and sound
- [The cast and the game](alias_play_studio/06_CAST_AND_GAME.md) - 12. Managers: CHALK and PLAYHEAD; 13. Eggs
- [The system menu, your stuff, the settings](alias_play_studio/07_SYSTEM.md) - 14. The system menu and commands; 15. Your stuff, no code; 16. Settings file reference
- [Golden rules, troubleshooting, tests](alias_play_studio/08_RULES_AND_TESTS.md) - 17. Golden rules; 18. Troubleshooting; 19. Tests that lock behaviour
- [The gears: the run chain](alias_play_studio/09_GEARS.md) - 20. Opener, run, closer; the switches; the master chat queue; values

---

GOPOD — [README](README.md) · [WIRED-POD](WIRED-POD.md) · [asks](BROBOT_ASKS.md) · [mobile gear](MOBILE_GEAR.md) · [livestream](LIVESTREAM.md)

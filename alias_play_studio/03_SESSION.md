# The session and runs

Part of [ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md) - [Concepts](concepts/README.md) · [Songs](02_SONGS.md) · [The session and runs](03_SESSION.md) · [The Brobots chat](04_CHAT.md) · [Voices and sound](05_VOICES_AND_SOUND.md) · [The cast and the game](06_CAST_AND_GAME.md) · [The system menu, your stuff, the settings](07_SYSTEM.md) · [Golden rules, troubleshooting, tests](08_RULES_AND_TESTS.md) · [The gears](09_GEARS.md)

<details><summary>Game card: the session</summary>
<!-- gopod topic: 08_the_session -->

# the session
- **chapter:** Show and tell

- **words:** the session, a session, replay a session, record session, the thanks ending, stop run
- **ask:** Tell me about the session
- **summary:** one Wire-Pod start to the next | runs start from the chat, stop run stops one | thanks guys ends it
- **source:** alias_play_studio/03_SESSION.md
- **ask 1:** Brobot 2, tell me about the session
- **ask 2:** Brobot 1, in a session, how do I stop run?
- **ask 3:** Brobot 2, how does the session end - the thanks ending?

---

The session. Reveal these, in order, in your own words:
1. A session is one Wire-Pod start to the next; Goverlord runs it and keeps its one record.
2. Runs - the interview, the songs - start from the chat, and "stop run" stops one at its next safe spot.
3. "Thanks guys" or "thanks brobots" ends the session at the next safe spot.

</details>

## The session
A **session** is one Wire-Pod start (`wpr 2`) to the next; Goverlord runs it (`gopod_session.go`).

| Phase | What gets through |
|---|---|
| **Warming** - start to LLM Ready | Nothing reaches the Brobots. |
| **Ready** - live chat | Every way in, to the Brobot it names. |
| **Run** - one run owns the Brobots | Only `stop run`, thanks, Record Session's answer and the run's own lines. |

The chat's memory is in memory only: `wpr 2` starts it empty; F5 only redraws the page. The
session log keeps everything.

## Runs
- The runs, their start words and their pre/post chat: the run cards,
  [`game-runs.md`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/behaviours/game-runs.md). Each song is its folder in
  [`goverlord/runtime/songs/`](../goverlord/runtime/songs/).
- **One at a time.** A run checks in with Goverlord when it starts and out when it ends; a second
  start is refused; a run whose process is gone is cleared.
- **The run owns the Brobots:** no live answers, no native Wire-Pod answers while it goes.
- **Before it starts,** the guards' security sweep: both Brobots found and powered, or KEYS stops
  it ([`dizzy-brobots.md`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/behaviours/dizzy-brobots.md)).
- **A dry run** touches nothing and doesn't check in.

## Wire-Pod start and the opener
At every start: a new chat feed (the page clears and reloads once), every scripted voice is made,
the speaker starts **off** until clicked, then the opener plays once the page is here. When its
ready line is said, the chat opens. Then the mic check.

**The opener is a song:** [`00_llm_warming/`](../goverlord/runtime/songs/00_llm_warming/) (from the
template `_new_song/_llm_warm/`, named by `llm.opener_song_id`). Its words are its `story.md`:
edit, then `wpr 2`. One System llm bubble, a message at a time, each line its own voice;
"(music)" shows where the sting plays and isn't said.

## Every run's warm-up
One template for every run: the warm-up bubble first, kept at the bottom until LLM Ready; then
whatever plays in between (the song's `_llm_song` run); then the ready line. A warm model skips the
wait. The run goes on only once the page has played the ready line and its sting. Lines typed
meanwhile wait in a queue and play in order. A song opts out with `"llm_warmup": false`.

## The page away
The page is **here** once it can play (after `wpr 2` / F5, click Log); hidden, collapsed or silent
for `session.page_present_seconds`, it's **away**. While away, the opener, songs, the interview and
replays wait before their next step ("Goverlord awaits your return..."); a timed song plays on.

## Stop, skip, replay
- `stop run` stops the run at its next safe spot and says "Run stopped."; `stop session` stops a
  replay.
- **Skip** the reporters' warm-up lines: Esc, the numpad's `/`, a typed or spoken "skip", or
  `system-skip`. Managers and Brobots are never skipped.
- `replay <name>` plays a saved session's score back through the live chat; a replay is never
  recorded.

## Record Session (the "Record?" box)
1. **Ticked:** recording starts at the next safe spot, in the session's folder
   (`~/gopod_tts/sessions/<date>/<time>_session/`): the mic, the chat, Wire-Pod's log.
2. **Unticked:** it stops and keeps the session.
3. **Thanks, or `max_minutes`:** the Butler asks keep or delete; no answer in
   `answer_wait_minutes` deletes. **Keep:** the takes go on the take sheet, the folder is zipped.

## The thanks ending
"thanks guys" or "thanks brobots" ends the session at the next safe spot, once, in order: a run
going is closed, the score, the reporters' closing, the brothers' sign-off, Record Session's
question, "You're welcome." (`gopodThanksEnding`).

## Mic, USB and battery watches
- **Mic and numpad:** after the warm-up and every `usb_watch_seconds`, Git Repo says when the mic
  or the numpad goes missing or comes back. A missing mic turns numpad keys into dry runs.
- **Battery and network:** every 15 s each Brobot's battery is read; low or a dip, his guard warns
  and hands over, and he says he's dizzy; dropping off the network and coming back is logged too.
  Lines and thresholds: [`dizzy-brobots.md`](../goverlord/wire_pod_overlay/chipper/gopod_probes/topic_clusters/personas/behaviours/dizzy-brobots.md).

[Back to ALIAS-PLAY-STUDIO](../ALIAS-PLAY-STUDIO.md)

# WIRED-POD

> Wire-Pod gets one robot talking. This is what GOPOD adds on top of it, and how to take it home.

**What this is.** A reference for Wire-Pod owners: what GOPOD changes in Wire-Pod, how it stays
out of Wire-Pod's own files, how to build and test it, and how to set it up on your own machine
in three parts. The full show standard (songs, the chat page, every rule and command) is
[ALIAS-PLAY-STUDIO.md](ALIAS-PLAY-STUDIO.md).

- **Native Wire-Pod:** [`kercre123/wire-pod`](https://github.com/kercre123/wire-pod), as its author ships it.
- **Brobots Wire-Pod (this document):** two Vectors, Brobot 1 and Brobot 2, talking through
  Wire-Pod with GOPOD on top. Everything runs locally: a small LLM, no cloud.

Paths like `/home/goverlord/...` are the author's machine. Change them to yours.

**GOPOD not tested on other devices or set ups. Back up your stuff and attempt at your own risk.**

---

## Contents

1. [What's new for Wire-Pod owners](#1-whats-new-for-wire-pod-owners)
2. [Stay thin: how GOPOD touches Wire-Pod](#2-stay-thin-how-gopod-touches-wire-pod)
3. [Build, deploy and test](#3-build-deploy-and-test)
4. [Voice commands](#4-voice-commands-customintentsjson)
5. [Take Out: set it up at home](#5-take-out-set-it-up-at-home)
6. [Where this stands](#6-where-this-stands)

---

## 1. What's new for Wire-Pod owners

- **Two Brobots, one conversation.** Ask either Vector. Both share one chat memory, and each
  knows which brother it is by its serial. Call one by name ("Hey Vector, Brobot 1, …") and
  only he answers.
- **No cold start.** The local model loads when Wire-Pod starts and stays loaded.
- **A chat window on Wire-Pod's Log page.** Each reply shows as a bubble, and the piece being
  spoken lights up as the robot says it. Type a line instead of speaking, or hold a numpad key
  and talk.
- **Animation, then voice, the native way.** The model picks its own animation cue; Wire-Pod
  plays it, then the voice comes in.
- **Said right.** A pronunciation list fixes names before they're spoken (GOPOD is "Gowp-awd").
- **Named runs.** Type `start interview` for the scripted interview, `start bingo test` for
  the Bingo test song, `start baby brobots` for the nap's lyric video, and `stop run` to stop
  any of them at its next safe spot.
- **Generate, then script.** A dry run asks every interview question through the real chat
  path with no robots. The best takes become a song folder set up like the Bingo test
  (`goverlord/runtime/songs/02_brobots_interview/`: `story.md` for the words, `knobs.json` for
  who says each line, cues and moves) that you edit and replay. Every run is the same, ready to
  record.
- **Page voices.** Two Kokoro reporters read the questions and host the show. A third voice,
  Goverlord, says the system lines (LLM Ready, chapters, "You're welcome."). Your own typed and
  numpad lines are read out in an Operator voice.
- **Record a session.** Tick one box: the chat, Wire-Pod's log and the room mic are saved
  together. Say "Thanks Brobots" and a Brobot asks whether to keep it.
- **One settings file.** Every value and text lives in
  `webroot/gopod_brobots_settings.json`, and Wire-Pod checks every key at start.

---

## 2. Stay thin: how GOPOD touches Wire-Pod

`~/wire-pod` stays byte-for-byte upstream (commit `dc6ae0c`). GOPOD reaches it only two ways:

1. **The overlay build.** Go's `-overlay` flag swaps in GOPOD's copies of nine native files
   and adds GOPOD's own Go files. A plain `go build` gives you stock Wire-Pod.
2. **One line** in Wire-Pod's `chipper/source.sh`: `source <overlay>/chipper/gopod_env.sh`.

### The nine native files GOPOD changes

| File | What GOPOD changes |
|---|---|
| `ttr/kgsim.go` | The Brobots' prompt, personas and shared chat, and the speaking loop. |
| `ttr/kgsim_cmds.go` | Animation and speech go through the live-speech gate; an animation call gives up after 10 s, so Wire-Pod's own animation queue is never held by a robot that dropped. |
| `ttr/kgsim_interrupt.go` | A touch stops a reply at the end of the current piece. |
| `ttr/weather.go` | A network error no longer crashes Wire-Pod. |
| `ttr/matchIntentSend.go` | Wire-Pod's own voice answers can be switched off (the chat's `system-native on\|off`), and are always off during a run. |
| `sdkapp/server.go` | The chat window's routes: typed input, Record Session, numpad push-to-talk, the LLM warm-up. |
| `config-ws/webserver.go` | The web page is served from GOPOD's folder first, native after. |
| `sdkapp/bcassume.go` | Crash guard: the robot is found by its serial. |
| `sdkapp/robot.go` | The same crash guard in the connection timer. |

**Following upstream.** GOPOD's copies follow upstream `dc6ae0c`, which changed 2 of the 9
files. Upstream's `kgsim.go` OpenAI model-setting fix is carried into GOPOD's copy. Upstream's
MaxCompletionTokens switch is not carried: GOPOD's copies keep their tuned MaxTokens 5120 for the
local LLM. `system-version` in the chat shows which of the nine files changed upstream.

Everything else (the `gopod_*` Go files, prompts, topic packs, tools, the chat page, the voice
commands) lives in `goverlord/wire_pod_overlay/`, and Wire-Pod reads it from there.

**Never in this repo:** `apiConfig.json` (keys), `jdocs/` and `certs/` (robot tokens),
`sessions/` (conversation logs and recordings).

---

## 3. Build, deploy and test

**Build and deploy:** `goverlord/wire_pod_overlay/deploy_overlay_build.sh` runs the tests,
builds, backs up the live binary and swaps the new one in. If a test fails, nothing deploys.
Then restart `wire-pod.service`.

**Tests:**

```
GOPOD_OVERLAY_CHIPPER=<overlay>/chipper go test -tags nolibopusfile -overlay=<overlay>/overlay.json ./pkg/wirepod/ttr/
```

---

## 4. Voice commands (`customIntents.json`)

| Say | Does |
|---|---|
| "start interview" | Plays the scripted interview, the same as typing it. Nothing happens before LLM Ready. |
| "start bingo test" | Runs the Bingo test song live. |
| "start baby brobots" | Runs the nap's lyric video: its soundtrack and captions in the chat while the Brobots sleep, then "Brobots Ready!". |
| "stop run" | Stops the running interview or song at its next safe spot. |
| "GOPOD Yourself" | The wake phrase. Brobot 1 answers with the newsletter, Brobot 2 with crushn8r.ca. |
| "remember my face" | The robot scans and saves your face (Wire-Pod's own "my name is CrushN8r"). Voice only. |

Typed in the chat, the same phrases do the same thing.

**Gotcha:** Wire-Pod caches this file at start, and its web config page rewrites the whole file
on any edit. After editing it by hand, restart Wire-Pod before touching that page.

---

## 5. Take Out: set it up at home

From the GOPOD kitchen: a thank-you back to the Wire-Pod community, packed to go in three
parts. Order one, two or all three. Each part says what you need, what to copy, what to
change, and how to know it works.

### Part 1: the alias play studio

The full standard (the notes, the song folder, the golden rules, how to make a new song) is
[ALIAS-PLAY-STUDIO.md](ALIAS-PLAY-STUDIO.md). The cockpit for scored songs and robot tuning:
`pha0b` (PLAYHEAD: pick a song, play a range of it, live or dry) and `phcal` (tune one arm, nod,
cube light or animation on one robot and save it). A song is a folder under
`goverlord/runtime/songs/`: `story.md` (the lines) and `knobs.json` (what each step does, and
when).

- **You need:** Part 2's Wire-Pod (the studio drives the robots through its `/api-sdk/*`
  endpoints), Python 3, bash.
- **Copy:** `goverlord/alias_lib_overlay/` to `~/.gopod_alias_lib/`. The songs folder stays in
  the repo, where the aliases find it.
- **Load:** in `~/.bash_aliases`, source `brobots.sh` (pha0b, phcal, score, gobingo),
  `llm.sh` and `wirepod_logs.sh` from `~/.gopod_alias_lib/`.
- **Change:** copy `phcal_config.example.json` to `phcal_config.json`,
  `gopod_audio_config.example.sh` to your own, and
  `goverlord/runtime/data_gomad/configs/endpoints.example.json` to `endpoints.json`; put in
  your robots' serials, your Wire-Pod address and your paths.
- **Works when:** in a new terminal, `pha0b` lists the songs (Bingo test, Bingo game, nap).

### Part 2: the chat page

Everything in [What's new for Wire-Pod owners](#1-whats-new-for-wire-pod-owners) above.

- **You need:** Wire-Pod working with your robots (`chipper/v1.5.10`, commit `dc6ae0c`), Go,
  Part 3's LLM. For the reporter and Goverlord voices: Kokoro-82M in its own Python venv (set
  `reporter_voices.python` and `worker` in the settings file); without it the chat runs
  without them. For numpad push-to-talk and Record Session: a USB mic.
- **Copy:** clone this repo. The overlay is `goverlord/wire_pod_overlay/`.
- **Change:** your robots' serials (`gopodBrobot1ESN` / `gopodBrobot2ESN` in
  `gopod_session_memory.go`); the absolute paths in `overlay.json`, `customIntents.json`,
  `build_with_overlay.sh` and `gopod_env.sh`. Add the `source` line to
  `~/wire-pod/chipper/source.sh`.
- **Build:** `goverlord/wire_pod_overlay/deploy_overlay_build.sh`, then restart
  `wire-pod.service`.
- **Works when:** on Wire-Pod's web page, click **Log**. The Brobots chat opens, the reporters
  play the opening, LLM Ready shows, and a typed "who are you?" to Brobot 1 comes back as a
  bubble that lights as he speaks.
- **Your own scripted run:** `harvest_dry_run.sh` asks every question with no robots and
  writes a draft. Turn it into a song folder (the `TestWriteInterviewSong` test, with
  `GOPOD_WRITE_INTERVIEW_SONG=<folder>`), point `interview_script.song_dir` at it, edit its
  `story.md`, then type `start interview` in the chat.

### Part 3: the offline LLM

The Brobots think with a small local model: no cloud, no monthly bill, no internet needed.

- **Tested on:** an NVIDIA Jetson Orin NX 16GB (Ubuntu) with Ollama and `gemma2:2b`. A cold
  load takes 35 seconds to 2 minutes; then the model stays loaded and a typical answer comes
  back in 3 to 6 seconds. Other hardware isn't tested: anything that runs Ollama and
  `gemma2:2b` comfortably should do.
- **Set up:** install Ollama, `ollama pull gemma2:2b`. In Wire-Pod's settings, set the
  knowledge provider to `custom` with your Ollama's OpenAI-style endpoint and the model
  `gemma2:2b`, with intent graph and **save chat** on.
- **What GOPOD adds:** the warm-up at start and the keep-loaded refresh (`llm` in the settings
  file), the Brobots' prompt and personas, and the reply clean-up.
- **Works when:** after a Wire-Pod restart, the log shows `GOPOD_LLM_WARMUP done`, and
  "Hey Vector, who are you?" gets a Brobot answer.

---

## 6. Where this stands

**Live-tested:** the Brobots chat, shared memory, personas, the piece-by-piece highlight,
the LLM warm-up, the scripted interview run, the Bingo test song, numpad push-to-talk with a
spoken line, and Record Session's untick path.

**Not yet, honestly:**
- Call-by-name is tested in code, not yet live.
- Record Session's "Thanks Brobots" keep-or-delete ending is tested in code, not yet end to
  end live.
- The highlight can light the wrong words on a few replies (28 of 838 checked), where a
  respelled word starts a piece. The fix is planned: cut the bubble once and respell each
  piece for speech.
- "GOPOD Yourself" is the wake phrase for the GOPOD layer, which isn't built yet. Until then,
  each Brobot answers it with his own call to action.
- Known stubs, kept on purpose: the DAG anchor check (`verifyGOPODDAGAnchor`) is a no-op;
  `selectGOPODContactBeat` has no caller.

---

> From Doctrine Barfallonyou
> Lesson! You don't showcase a system by hiding what's still a stub. You showcase it by naming the stub and shipping the rest anyway.
> Boom! Done! Class Dismissed!
> — Doc Squawkadoodle

---

GOPOD — [README](README.md) · [channel](CHANNEL.md) · [ask](ASK.md)

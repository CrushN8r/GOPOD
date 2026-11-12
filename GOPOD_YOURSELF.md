# GOPOD Yourself

> You saw two Vectors put on a show. Here's how to put yours on stage.
> One guide, start to finish. Hand it to your AI helper too - it's written for both of you.

**This is my gift back to the community** - the Wire-Pod people who kept Vector alive. Take what
you like. Make it yours. Then show us what you made.

**GOPOD is tested on one machine: mine.** This guide covers what I know. Otherwise: research,
back up your stuff, and test at your own risk. Everything below that hasn't been tried on another
set-up says so.

---

<details><summary>Game card: the GOPOD Yourself guide</summary>
<!-- gopod topic: 06_build_your_own -->

# the GOPOD Yourself guide
- **chapter:** The brain and the rig

- **words:** gopod yourself guide, build your own, your own robots, gift back, perform or compose
- **ask:** Tell me about the GOPOD Yourself guide
- **summary:** the guide to put your own Vectors on stage | perform or compose | ten steps, each with a done when
- **source:** GOPOD_YOURSELF.md

---

The GOPOD Yourself guide. Reveal these, in order, in your own words:
1. GOPOD Yourself is the guide to put your own two Vectors on stage - written for you and your AI helper, and a gift back to the Wire-Pod community.
2. Pick your lane: PERFORM, a rig that rolls into a room and runs offline; COMPOSE, the Alias Play Studio - your own show, your own way.
3. Ten set-up steps, each ending with "Done when" - from backing up Wire-Pod to the first chat.

</details>

## Pick your lane

GOPOD grows two ways. Pick one, or both - they share the same set-up (the steps below).

| | **PERFORM** | **COMPOSE** |
|---|---|---|
| **What it is** | How GOPOD is put together as a mobile performance station. | How GOPOD is played: the Alias Play Studio. |
| **You get** | A rig that rolls into a room, runs offline, and lets the crowd join the chat from their phones. | Songs, scripted runs, a cast of voices, and a system menu to change it all without code. |
| **Made for** | Paying local gigs. Making content on the move - a live, offline session in a (safely) travelling car, van or bus. | Making your own show, your own way, with your own robots. |
| **Start at** | [Perform: take it on the road](#perform-take-it-on-the-road) | [Compose: play it your way](#compose-play-it-your-way) |

Both lanes lead to the same place: **[Share the glow](#share-the-glow)** - your clips, your pages,
your crowd, and the GOPOD sites ([LIVESTREAM.md](LIVESTREAM.md#where-it-lands)).

---

## Before you start

### What can run it

GOPOD rides on top of Wire-Pod, so it runs where Wire-Pod is **built from source on Linux**.

| Machine | Status |
|---|---|
| NVIDIA Jetson Orin NX 16GB, Ubuntu | **Tested.** The author's rig. |
| A Linux PC or mini-PC (x86_64) | Not tested. Should work: everything GOPOD uses runs there. |
| Raspberry Pi 4/5, ARM TV boxes running Linux (a T95 with Armbian, say) | Not tested. Wire-Pod fits; the LLM may not. With under 8 GB of memory, run Ollama on a second machine on the same network and point Wire-Pod at it. |
| Windows, macOS, Wire-Pod's Android app | No. GOPOD's build and scripts need Linux (bash, systemd, Go). |

Rough sizes: Ollama with `gemma2:2b` needs about 3 GB of free memory; the Kokoro voices about 1 GB
more. On the Jetson, a Brobot starts talking about 8 to 10 seconds after a question; the numpad
voice cast takes 20 to 25 seconds.

### How many robots

| You have | What GOPOD does |
|---|---|
| **1 Vector** | He's Brobot 1, solo. In the live chat, anything meant for Brobot 2 (or for both) goes to him. In scripted runs, Brobot 2's lines play on the computer's speaker. |
| **2 Vectors** | Brobot 1 and Brobot 2: one shared chat, brother banter, the full show. |
| **3 or more** | **You pick two** to be the Brobots ([step 4](#step-4-name-your-brobots)). The rest stay ordinary Wire-Pod Vectors: they keep working the way they always have, just not in the Brobots' cast, shared chat or songs. Swap who's on stage by changing the pick and restarting. Not live-tested with more than two robots connected. |

The voice-only cast on the numpad (CHALK, PLAYHEAD, the Cozmo students, the Moorebot guards, Cameo)
has no robot bodies yet. Extra Vectors are the obvious future bodies - not built.

---

## The set-up, start to finish

Ten steps. Each one ends with **Done when**: don't move on until it's true.

### Step 0: Back up

If you already run Wire-Pod: copy `~/wire-pod/chipper/apiConfig.json` and the `jdocs/` and
`certs/` folders somewhere safe. That's your robots' pairing and your settings.

**Done when:** you could put your current Wire-Pod back if tonight goes sideways.

### Step 1: Wire-Pod, the version GOPOD is built on

GOPOD's copies of Wire-Pod's files follow Wire-Pod commit `dc6ae0c` (`chipper/v1.5.10`).
Clone it into your home folder, as yourself (not root), so GOPOD's build can reach it:

```bash
git clone https://github.com/kercre123/wire-pod ~/wire-pod
```

```bash
git -C ~/wire-pod checkout dc6ae0c
```

Then set it up and pair your robots the way Wire-Pod's own guide says
([the Wire-Pod wiki](https://github.com/kercre123/wire-pod/wiki)), as a service
(`wire-pod.service`). A newer Wire-Pod may work, but isn't tested; GOPOD's `system-version`
command shows which of its nine files Wire-Pod has changed since.

**Done when:** stock Wire-Pod works - "Hey Vector, what time is it?" gets an answer from each robot.

### Step 2: The brain (Ollama)

Install [Ollama](https://ollama.com), then:

```bash
ollama pull gemma2:2b
```

In Wire-Pod's web page, under the knowledge graph settings: provider `custom`, endpoint
`http://localhost:11434/v1` (or the address of the machine running Ollama), model `gemma2:2b`,
intent graph **on**, save chat **on**.

**Done when:** "Hey Vector, I have a question... who are you?" gets an answer from the LLM, still on
stock Wire-Pod.

### Step 3: GOPOD, pointed at your machine

```bash
git clone https://github.com/CrushN8r/GOPOD ~/GOPOD
```

GOPOD's files carry the author's paths (`/home/goverlord/...`). One script swaps them for yours.
Look first, then apply:

```bash
bash ~/GOPOD/goverlord/gopod_paths.sh
```

```bash
bash ~/GOPOD/goverlord/gopod_paths.sh --apply
```

The first run also lists every path **outside** GOPOD's folder that GOPOD looks for, marked `ok`
or `MISSING`. Step 5 tells you which ones you need.

**Done when:** a second run says `no author paths left - already done`.

### Step 4: Name your Brobots

Each robot's serial (ESN) is on Wire-Pod's web page, under the robot's settings. Put your two
picks in GOPOD's settings, `goverlord/wire_pod_overlay/chipper/webroot/gopod_settings/02_cast.json`:

```json
"brobots_setup": {
  "brobot_1": "<serial of your Brobot 1>",
  "brobot_2": "<serial of your Brobot 2>",
```

Leave both empty and GOPOD takes the first two robots Wire-Pod knows. Later, the chat's system
menu changes it without a text editor (`system-set brobots_setup.brobot_1`, then
`system-restart`).

**Done when:** your picks are saved (or you've chosen "first two").

### Step 5: Install the pieces

GOPOD's parts are optional except the first three rows. Skip a row and that feature stays off;
the rest still runs.

| Piece | What for | Required? |
|---|---|---|
| **Go** (at `/usr/local/go`, or set `GOPOD_GO`) | Building Wire-Pod with GOPOD. | Yes |
| **Vosk library and headers** | Building Wire-Pod (its speech-to-text). Wire-Pod's own setup usually puts them in `/root/.vosk/libvosk`. | Yes |
| **Python 3.10 to 3.12** (not 3.13: GOPOD uses `audioop`) | The room mic, the songs, the studio. | Yes |
| `pip install vosk sounddevice numpy scipy evdev` | Numpad push-to-talk and the room mic. | For the mic |
| **Vosk model** `vosk-model-small-en-us-0.15` from [alphacephei.com/vosk/models](https://alphacephei.com/vosk/models), unzipped into `goverlord/runtime/data_gomad/models/vosk/` | The room mic's speech-to-text. | For the mic |
| **Kokoro voices**: a Python venv at `~/gopod_tts/engines/kokoro_82m_001/venv` with `pip install kokoro soundfile`, plus `espeak-ng` from your package manager | The reporters, Goverlord and the numpad cast's voices. Voices download from Hugging Face on first use. Put the venv elsewhere and change `reporter_voices.python` in `05_voices_and_robots.json`. | For the voices |
| `sox` | Pitching two cast members on one voice apart. | Optional |
| `wamerican` (the `/usr/share/dict` word list) | Better splitting of made-up merged words. A missing list is skipped. | Optional |
| The Wire-Pod Vector Python SDK ([kercre123/wirepod-vector-python-sdk](https://github.com/kercre123/wirepod-vector-python-sdk)) | The studio's robot tuning (`phcal`) and the Bingo reactor. | For the studio |
| The vectorx SDK | The live Bingo game (`gobingo`). The tracked `vectorx-gobingo` binary is ARM64 (Jetson, Pi); on a PC, rebuild it from `cmd/main.go` - not tested. | Bingo game only |
| Coral USB Edge TPU, `libedgetpu1-max`, `pycoral` | Object boxes on the camera feeds. See [models/README.md](goverlord/runtime/data_gomad/models/README.md). | Optional |
| SearXNG in Docker (`goverlord/wire_pod_overlay/searxng/`) | GOPOD's own web search. | Optional |

**Done when:** step 3's `MISSING` list holds only the things you chose to skip.

### Step 6: Hook GOPOD into Wire-Pod

One line, at the end of `~/wire-pod/chipper/source.sh` (Wire-Pod's own local file):

```bash
source ~/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_env.sh
```

Write the full path, not `~`: Wire-Pod runs as root, and root's `~` is another folder.

In `gopod_env.sh`, check the one path it sets: the folder holding `libvosk.so`. If your Vosk lives
elsewhere (`/root/.vosk/libvosk`, say), change it there, and tell the build where it is:

```bash
export GOPOD_VOSK_LIB=/root/.vosk/libvosk GOPOD_VOSK_INC=/root/.vosk/libvosk
```

Also set `GOPOD_WIREPOD_DIR` if your Wire-Pod isn't at `~/wire-pod`.

**Done when:** `libvosk.so` and `vosk_api.h` are where the build is told to look.

### Step 7: Build, deploy, restart

```bash
bash ~/GOPOD/goverlord/wire_pod_overlay/deploy_overlay_build.sh
```

It runs GOPOD's tests first (a failed test deploys nothing), builds, keeps a dated backup of your
current Wire-Pod binary in `wire_pod_overlay/chipper_backups/`, and swaps the new one in. Then:

```bash
sudo systemctl restart wire-pod
```

**Done when:** the script ends with `DEPLOY_OK`, and Wire-Pod's log
(`journalctl -u wire-pod -f`) shows `GOPOD_BROBOTS Brobot 1 ...` and, a minute or two later,
`GOPOD_LLM_WARMUP done`.

### Step 8: The first chat

Open Wire-Pod's web page and click **Log**. The Brobots chat opens. Type "who are you?" to Brobot 1.

**Done when:** his answer comes back as a bubble that lights up word by word as he says it.
That's GOPOD. You're in.

### Step 9: The way back

Changed your mind? Copy the backup binary from `chipper_backups/` over `~/wire-pod/chipper/chipper`,
take the `source` line out of `source.sh`, and restart. A plain `go build` of `~/wire-pod` is
stock Wire-Pod again: GOPOD never edits Wire-Pod's own files.

---

## Perform: take it on the road

How GOPOD is put together as a mobile performance station.

- **The rig.** Two crates on a dolly: power station, Jetson, hotspot, mic, numpad. Roll in, plug
  in, party. Parts, power modes and photos: [MOBILE_GEAR.md](MOBILE_GEAR.md). About $1K to $3K CDN.
  Your own rig can be anything that runs the set-up above, plus a Wi-Fi network the robots and
  the crowd's phones share.
- **Offline first.** No internet needed: the brain, the voices and the speech-to-text run on the
  rig. That's what makes a gig in a basement, a park or a moving vehicle work. Strap the gear down
  and keep the driver out of it.
- **The crowd joins the chat.** The GOPOD app (its home is
  [accessmath.crushn8r.net](ACCESSMATH.md)) is how people at a live gig join in from their phones:
  1. They join the rig's Wi-Fi and open the chat (a QR code on the table helps). Quick way, in the
     browser: `http://<the rig's address>:8080`. App way: `https://<the rig's address>:8443/?app`,
     after trusting GOPOD's certificate once (*The certificate*, below).
  2. "Add to home screen" (Android: "Install app") makes it an app that opens straight into the chat.
  3. Their first line is the username they want. You approve it in the system window's
     "Local network visitors" slot, and switch participation **on** when you're ready.
  4. Approved, they talk to the Brobots under their name. Visitors only ever see the chat:
     Wire-Pod's settings, logs and GOPOD's commands are closed to them.
- **The certificate: why and how.**
  - *Why the secure page matters.* A phone only offers "install as app" on a secure (`https://`)
    page, and the secure page scrambles each visitor's lines on the shared Wi-Fi. Certificate
    companies don't issue certificates for a local network, so GOPOD makes its own: a local
    authority called **GOPOD local** (made once, kept in `~/.gopod_https`, never in the repo) and a
    certificate for the rig's own names and addresses (made again when they change).
  - *Why it's safe to trust.* GOPOD's authority is fenced (`app.fence_dns`, `app.fence_ips` in
    `06_system.json`): it can only vouch for local names (`localhost`, `*.local`, `*.lan`,
    `*.home.arpa`) and private addresses (`192.168.x.x`, `10.x.x.x` and the like) - never a public
    website. Its private key stays on the rig.
  - *Install it, once per phone.* On the rig's Wi-Fi, open
    `https://<the rig's address>:8443/gopod_ca.crt`.
    - Android: Settings > Security > Encryption & credentials > Install a certificate >
      CA certificate > `gopod_ca.crt`.
    - iPhone/iPad: allow the download, then Settings > General > VPN & Device Management >
      GOPOD local > Install; then Settings > General > About > Certificate Trust Settings >
      switch GOPOD local on.
    - Then close the browser completely and open it again. Menu names vary by phone and version.
  - *Remove it after.* Android: Encryption & credentials > User credentials. iPhone: VPN & Device
    Management > GOPOD local > Remove.
  - *Two things that make phones trust it again:* an authority made before the fence existed is
    set aside and a fenced one made at the next start, and a new `cert_dir` means a new authority.
    Either way, phones install `/gopod_ca.crt` once more.
- **Record the session.** Tick one box: the chat, Wire-Pod's log and the room mic are saved together
  - raw material for clips.
- **Gigs.** The rooms, birthdays, BBQs, videos: [BOOKING.md](BOOKING.md).

## Compose: play it your way

How GOPOD is played: the Alias Play Studio.

- **No terminal needed.** Click the speaker icon on the chat page: the system menu has every
  setting, song, voice and test, grouped in order. Its full map, written by GOPOD itself:
  [GOPOD_TORAH.md](GOPOD_TORAH.md).
- **Change the words.** A song is a folder under `goverlord/runtime/songs/`: `story.md` for the
  lines, `knobs.json` for who says what, the cues and the moves. Copy `_new_song/` to start one.
- **Script a run.** The LLM writes an interview, you keep the best takes, and it plays the same
  way every time - ready to record. WIRED-POD.md, ["Your own scripted run"](WIRED-POD.md#part-2-the-chat-page).
- **The cockpit.** `pha0b` (pick a song, play a range of it, live or dry) and `phcal` (tune one
  arm, nod, cube light or animation and save it). Set-up: WIRED-POD.md,
  [Part 1](WIRED-POD.md#part-1-the-alias-play-studio).
- **Make your own game.** The game in this repo is a test bench for live and scripted chat, not
  the game to play - an example of how to build one. Copy the topic template
  (`goverlord/wire_pod_overlay/chipper/gopod_probes/_topic_cluster/`) into `chapters/` and write
  your own: your topics, your asks, your scoring.
- **The cast.** Each persona is one file in
  `goverlord/wire_pod_overlay/chipper/gopod_probes/personas/`. Rewrite them; make your own.
- **Every rule, every note.** [ALIAS-PLAY-STUDIO.md](ALIAS-PLAY-STUDIO.md) - start with
  [Concepts and terms](alias_play_studio/01_CONCEPTS.md).

## Share the glow

This is the part that matters most. GOPOD grows when people see it.

- **Post it.** A clip of your Brobots, your rig, your crowd in the chat. Tag **@CrushN8r** and say
  the words: *GOPOD Yourself!*
- **Join the contest.** "AI Wordplay! Explain the Math!" - nine themes, one wheel:
  [LIVESTREAM.md](LIVESTREAM.md).
- **Say hi.** Show-and-tell, questions, a gig you played: contact@crushn8r.ca, or the CRUSHN8R
  CREW'd newsletter on [crushn8r.ca](https://crushn8r.ca). Tell your Wire-Pod groups what you built.
- **Help wanted.** Social media help, near Windsor-Essex: [SOCIAL_MEDIA.md](SOCIAL_MEDIA.md).

---

## Handing this to your AI helper

Paste this into the AI that's helping you (Claude, ChatGPT, anything that can read a repo):

```text
Help me set up GOPOD (https://github.com/CrushN8r/GOPOD) on my machine.
Read GOPOD_YOURSELF.md first and follow its steps in order, 0 to 8.
Don't skip a step's "Done when" check; stop and tell me if one fails.
My machine: <Jetson / Linux PC / Raspberry Pi / other>, <OS and version>.
My robots: <how many> Vectors, already on Wire-Pod: <yes/no>.
I want: <Perform / Compose / both>.
```

Rules for the AI:

- Don't edit files in `~/wire-pod`. GOPOD changes Wire-Pod only through its overlay build and one
  `source` line (WIRED-POD.md, ["Stay thin"](WIRED-POD.md#2-stay-thin-how-gopod-touches-wire-pod)).
- Run `goverlord/gopod_paths.sh` instead of hand-editing paths.
- Robot serials go in the settings file or `~/.gopod_esns.json`, never in code.
- Never commit or share `apiConfig.json`, `jdocs/`, `certs/` or `sessions/`: keys, robot tokens
  and conversations.
- Where this guide says "not tested", say so to the human; don't promise it works.
- For deeper questions: [WIRED-POD.md](WIRED-POD.md) (what GOPOD changes in Wire-Pod),
  [ALIAS-PLAY-STUDIO.md](ALIAS-PLAY-STUDIO.md) (every behaviour, with the code and the test that
  locks it), [alias_play_studio/08_RULES_AND_TESTS.md](alias_play_studio/08_RULES_AND_TESTS.md)
  (troubleshooting).

## Stuck?

| You see | Try |
|---|---|
| Build fails: `vosk_api.h` or `-lvosk` not found | Step 6: `GOPOD_VOSK_LIB` / `GOPOD_VOSK_INC`. |
| `DEPLOY_FAILED tests failed` | Nothing was deployed. The log path is in the message; most often a path step 3 didn't catch. |
| No `GOPOD_LLM_WARMUP done` | Ollama isn't reachable: `ollama list`, and the endpoint in step 2. |
| The chat has no reporter voices | Kokoro isn't set up (step 5). The chat runs without them. |
| The wrong robot is Brobot 1 | Step 4, then restart Wire-Pod. |
| The voice commands do nothing | Wire-Pod caches `customIntents.json` at start: restart it. Don't touch its web config page after a hand edit. |

More: [08_RULES_AND_TESTS.md, troubleshooting](alias_play_studio/08_RULES_AND_TESTS.md#18-troubleshooting).

---

> From Doctrine Barfallonyou
> Lesson! The show isn't finished when it works for you. It's finished when it works for the next person.
> Boom! Done! Class Dismissed!
> — Doc Squawkadoodle

---

GOPOD — [README](README.md) · [WIRED-POD](WIRED-POD.md) · [studio](ALIAS-PLAY-STUDIO.md) · [mobile gear](MOBILE_GEAR.md) · [booking](BOOKING.md) · [livestream](LIVESTREAM.md)

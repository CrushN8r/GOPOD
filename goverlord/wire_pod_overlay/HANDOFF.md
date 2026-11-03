# GOPOD handoff - for Claude Code on the Jetson

Read this first, every session. Written 2026-11-03 at the end of a long cloud
session; the branch is `claude/magical-fermat-2dbhdj`.

## 1. Who you're working with, and how

The operator is CrushN8r (Master CrushN8r in the chat). GOPOD is his gift to the
Wire-Pod community ("otherwise, research, back up your stuff & test at your own
risk").

- **Plain words, short answers.** The outcome first. No jargon, no AI lingo.
- **Ask before any design decision**: file names and places, wording, behaviour.
  Don't decide for him. A yes/no or a short pick list is fine; no numbered menus.
- **Do the fix first.** When he gives you a run or a log, that is the truth: fix
  from it. Don't explore 50 options.
- **One copyable command at a time**, in its own box. Don't make him read files.
- **Don't waste his tokens.** Small, exact reads (grep, sed -n), not whole files.
- **Never guess and never claim untested work works.** Say what you ran and what
  you didn't.
- He tests live on two Vector robots; a mistake costs him real time on deadlines.
- **Keep as-is (his calls):** "1 have spoken", the "Topics 0 of 9:" rows, the
  battery warnings, the Brobots' slightly-off colour lines ("that's their charm,
  as long as they don't get too off topic"), the comm check's tester lines (his
  benchmark - the intro output must not depend on the question's wording).

## 2. The big rule: big brains = code, the LLM only colours

Every answer is a template filled by code from topic-cluster md files - the one
source of truth. The model (gemma2:2b on the Jetson) only writes the colour
lines in each persona's flavour. If something is said by code, it's in an md
file, once, and nowhere else (no copies in the settings JSON).

## 3. Where things are

All under `goverlord/wire_pod_overlay/chipper/`:

| What | Where |
|---|---|
| Go code (Wire-Pod overlay) | `pkg/wirepod/ttr/*.go` - built into Wire-Pod with `-overlay` (`../overlay.json`) |
| Settings (JSON) | `webroot/gopod_settings/01_brain.json` ... `06_system.json` |
| **Topic-cluster md files** (laid over the settings at load) | `gopod_probes/clusters/` |
| Persona profiles (one md each) | `gopod_probes/personas/` |
| Locks (fingerprints) | `gopod_probes/locks/` |
| Chat page | `webroot/js/gopod_brobots_chat.js`, `webroot/css/` |

**Topic-cluster format:** `# title`, head lines (`- **words:** ...`), `---`, then
sections `## Name (settings.path)` with points `1. key: value`. A value with
` | ` is a list. Clusters can set text and lists only (not numbers). Every
section needs its settings path in brackets (TestTopicClusters).

**A new Go file must be added to `overlay.json`** or it isn't built.

## 4. The golden, locked pieces

Each is locked by a deploy test. Changing one is the operator's word only; then
the deploy runs with `GOPOD_RELOCK=1` to save the new fingerprint.

| Golden | File | Test |
|---|---|---|
| Searching animation: clips back to back, 1 s minimum, one get-out | `locks/searching_golden.md` | TestSearchingGoldenLock |
| Weather run, duo: A reports, hands over, B confirms | `clusters/weather_run.md` | TestWeatherGoldenLock, TestWeatherBrainLock, TestAnswerShapesHold |
| Group chat intro: single intro, intro + weather (A, B, one alone) | `clusters/group-chat-intro.md` | TestGroupChatIntroLock, TestIntroWeatherOrder |
| Personas (soft lock) | `locks/persona_soft_lock.json` | TestPersonaSoftLock |

**The duo weather run (Brobot A = picked first, B = his brother):**
- A: "Got it!" > A's brain: PLAYHEAD "Cache PYC will help you with that." then
  Cache PYC "Weather info received from Git repo. Ready when you are Bro!" >
  "Thanks for your help, Cache PYC." > the weather report > A's take > "Brobot B,
  what say you on this, [nickname]?"
- B: "Nice try, [nickname]!" > B's brain: Cache PYC's line only > "Confirmed.
  Temperature is ..." > B's take > the Weather Kit line > "Hey, that reminds
  me..." > "Try asking: ..." on its own line.
- Each Brobot uses his own robot's units and clock (Doc: °C, 24 h; Pip: °F, 12 h).

**The group chat intro:** GOPOD's company orientation group chat. Guided Task
Flow: a) introduce yourself, b) name who's in the chat, c) Keys opens the session
(Keys's own c: "You're the first to speak - open up the session"). An intro = the
code's opening ("I'm {name}, my nickname is ...") + one model line + the code's
"Here in the chat with us: ...". Intro + weather = the intro file, then the
weather file: intro first, then the weather parts.

## 5. Deploy and test

Everything is run from `~/crushn8r_git/GOPOD` on the Jetson.

- **Pull + check + fix (asks y/n before each fix):**
  `git pull --no-rebase --autostash --no-edit origin claude/magical-fermat-2dbhdj && bash goverlord/wire_pod_overlay/gopod_autofix.sh`
- The deploy alone: `bash goverlord/wire_pod_overlay/deploy_overlay_build.sh` -
  runs the ttr tests, builds, swaps the binary. It does NOT restart.
- Restart: the operator's own `wpr 2` (in his shell, not in the repo).
- Then F5 on the chat page, then type `system-comm-check`.
- Test log on failure: `/tmp/chipper_overlay_tests_<time>.log`.
- Wire-Pod's log survives a restart: `journalctl -u wire-pod --since "1 hour ago" --no-pager`.

**Useful log tags:** GOPOD_ALL (duo order), BROBOT_SHAPE (answer shape ok /
drift), GOPOD_LINE_QUEUE (turns, watchdogs), GOPOD_THINKING (searching),
GOPOD_ROBOT_CMD (every robot command), GOPOD_FEED (what the page got).

**How a line is routed:** the page's All box or the numpad `*` = both Brobots
(the 1/2 menu picks who's first). KP1 / KP2 = that Brobot alone. KP0 = Keys (the
system). KP3-KP9 = the Gomads.

## 6. Where things stand (2026-11-03)

Live-tested and good: the duo weather run both ways (Tampa, Topeka, Toronto),
searching, state capitals ("Kansas" -> Topeka), "Try asking" on its own line,
the Copy Window button.

**Pushed, not heard live yet** (the next deploy has them):
- the turn watchdog: at 180 s, a line not finished gets another 180 s, then the
  turn ends (a duo's turn ended while B was still speaking). Logged as
  GOPOD_LINE_QUEUE "180s, not finished - another 180s".
- `089ae0c` a split cue name ("darting Eyes___") joined back; never the
  brother's last nickname; Copy Window leaves out the typing row.
- `361f4d2` the group chat intro and the intro + weather runs - first heard in
  the next `system-comm-check` (KP0-KP9 single intros; `*` intro + weather).

**The next step:** deploy, then `system-comm-check`, then read the run with the
operator. Fix only what he points at.

## 7. Never

- Never push `~/crushn8r_git/gopod_notes`, serial numbers, IP addresses, keys,
  logs, `apiConfig.json`, `jdocs/`, `certs/`, `sessions/`, `~/.gopod_esns.json`.
- Never skip or weaken a test to get green; fix the cause.
- Never change a locked golden without the operator's word.
- Never commit or push unless he asks; he usually asks for "commit & push".

## ## EOF

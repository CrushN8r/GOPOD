// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_reporter_voices.go
package wirepod_ttr

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// The reporter voices on the Brobots chat page (2026-10-06, operator: "have the
// female/male voices speak operator bubble"; "since 1 operator, use female
// reporter voice (brobot 3)"; "can you channel the audio to play through the
// wire-pod web page?"). Brobots 3 and 4 - the interview vamp's two hosts, Kokoro
// voices - are heard through the page, not the robots:
//   - an operator line typed on the page is read out before the Brobot answers
//     it: to Brobot 1 by Brobot 3, to Brobot 2 by Brobot 4 (reporter_voices.read_by),
//     to both by reporter_voices.operator_host;
//   - when the LLM is ready, the hosts' opening lines; after "Thanks Brobots",
//     their closing lines - the vamp's own words (reporter_voices.opening/closing).
// One Kokoro worker (gopod_probes/tools/gopod_reporter_voice_worker.py) loads the
// model once and writes each line to a WAV in reporter_voices.audio_folder, named
// by its voice and words: a line said before (the bookends - pre-generated at
// start) is never made again, across restarts too. The page plays the file
// (a reporter_voices.role feed line) and shows the line in a host bubble.

var gopodReporter = struct {
	sync.Mutex
	in        *bufio.Writer
	out       *bufio.Scanner
	ready     bool
	closing   chan struct{} // this ending's closing; closed once it's queued
	signedOff bool          // this ending's sign_off.lines are said
	ended     bool          // this ending is running or done (gopodThanksEnding)
}{}

// gopodReporterBookendsMade is closed once the opening and closing lines' voice
// files are made (or the voices can't start), so the opening never waits on a
// line still being made (live 2026-10-06 14:42: 18 s between two lines).
var gopodReporterBookendsMade = make(chan struct{})
var gopodReporterBookendsOnce sync.Once

func gopodReporterBookendsDone() {
	gopodReporterBookendsOnce.Do(func() { close(gopodReporterBookendsMade) })
}

// gopodReporterAudioDir is where the voice files go, inside the page's webroot.
func gopodReporterAudioDir() string {
	return filepath.Join(gopodChipperRoot(), "webroot", gopodBrobotsSettings.ReporterVoices.AudioFolder)
}

// GopodReporterVoicesStart starts the Kokoro worker and makes the bookends' voice
// files, in the background (called once at Wire-Pod start).
func GopodReporterVoicesStart() {
	rv := gopodBrobotsSettings.ReporterVoices
	if rv.Python == "" || rv.Worker == "" || len(rv.Hosts) == 0 {
		gopodReporterBookendsDone()
		return
	}
	gopodReporterStarting.Store(true)
	go func() {
		defer gopodReporterBookendsDone()
		defer func() {
			gopodReporter.Lock()
			ready := gopodReporter.ready
			gopodReporter.Unlock()
			if !ready {
				gopodReporterStarting.Store(false)
			}
		}()
		worker := rv.Worker
		if !filepath.IsAbs(worker) {
			worker = filepath.Join(gopodChipperRoot(), worker)
		}
		cmd := gopodAsOperator(exec.Command(rv.Python, worker))
		// 2026-10-09: the Kokoro voices on this machine (am_michael was fetched there).
		if rv.HFHome != "" {
			cmd.Env = append(cmd.Env, "HF_HOME="+rv.HFHome)
		}
		stdin, err1 := cmd.StdinPipe()
		stdout, err2 := cmd.StdoutPipe()
		if err1 != nil || err2 != nil {
			logger.Println("GOPOD_REPORTER_VOICES unable to start the worker")
			return
		}
		if err := cmd.Start(); err != nil {
			logger.Println("GOPOD_REPORTER_VOICES unable to start the worker: " + err.Error())
			return
		}
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), `"ready"`) {
				break
			}
		}
		gopodReporter.Lock()
		gopodReporter.in, gopodReporter.out, gopodReporter.ready = bufio.NewWriter(stdin), scanner, true
		gopodReporter.Unlock()
		gopodVoicesAtStart.Store(true)
		logger.Println("GOPOD_REPORTER_VOICES ready")
		for _, l := range append(gopodBookends(true), gopodBookends(false)...) {
			gopodReporterFile(l.Host, gopodReporterSays(l.Text))
		}
		// 2026-10-08: the system voice's lines too (LLM status first, before the opening).
		for _, text := range gopodSystemLines() {
			if strings.TrimSpace(text) != "" {
				gopodReporterFile(gopodBrobotsSettings.SignOff.SystemRole, gopodReporterSays(text))
			}
		}
		gopodReporterBookendsDone()
		// 2026-10-08 (operator: "generate all reporter lines for both reporters"):
		// every line a reporter may read - the bookends and the newest draft's
		// lines and questions - made in both voices now, so none is made mid-run.
		// 2026-10-13 (operator: "I no longer need all voices to have all copies
		// of all exchanges * if there's a process to do all of all, prune it"):
		// each line is made once, in the one voice that says it - a reporter's
		// lines in his voice, a question in its reader's, the run commands in
		// the Operator voice that reads a typed command.
		made := 0
		for _, l := range gopodReporterAllLines() {
			if _, _, ok := gopodReporterCached(l.Host, gopodReporterSays(l.Text)); ok {
				continue
			}
			if _, _, ok := gopodReporterFile(l.Host, gopodReporterSays(l.Text)); ok {
				made++
			}
		}
		gopodVoicesAtStart.Store(false)
		logger.Println("GOPOD_REPORTER_VOICES all lines ready, made " + strconv.Itoa(made) + " new")
		gopodReporterPrune()
		go func() { _ = cmd.Wait() }()
	}()
}

// gopodReporterVoice: a reporter's Kokoro voice, or the system voice for the
// system role (operator 2026-10-08: "Add a third bm_george voice for system").
func gopodReporterVoice(host string) string {
	rv := gopodBrobotsSettings.ReporterVoices
	// 2026-10-29: robots off - a Brobot's temp voice (robots_off.voices, the
	// numpad's KP1/KP2); none picked: the unassigned voice (gopod_robots_off.go).
	if GopodRobotsOff() && gopodIsBodiedBrobot(host) {
		if voice := gopodBrobotsSettings.RobotsOff.Voices[host]; voice != "" {
			return voice
		}
		return rv.Hosts[rv.UnassignedVoiceHost]
	}
	if voice := rv.Hosts[host]; voice != "" {
		return voice
	}
	// 2026-10-28: a demo host ("Brobot 1 demo") speaks with its host's voice.
	if of := rv.DemoVoices[host]; of != "" && rv.Hosts[of] != "" {
		return rv.Hosts[of]
	}
	if sr := gopodBrobotsSettings.SignOff.SystemRole; sr != "" && host == sr {
		return rv.SystemVoice
	}
	if host == gopodOperatorVoiceHost {
		if rv.OperatorVoice != "" {
			return rv.OperatorVoice
		}
		return rv.Hosts[rv.UnassignedVoiceHost]
	}
	return ""
}

// GOLDEN global message (2026-10-18, operator: "no more "Generating..." as a
// separate bubble * decouple generating + hourglass into its own 'global
// message variable' available to all bubbles * all bubbles that generate voices
// use this * [message0]/[message1] * message0: "Generating..." + hourglass *
// message1: generated playable voice wav file + [transcription]"): a shown line
// whose voice isn't made yet goes to the page first, marked waiting
// (generating_marker+"wait:"+text) - its own bubble shows "Generating..." and
// the hourglass (message0) until its voice plays, then its words (message1).
// "off": the voice failed - its words show. Shown lines all come through here.
func gopodSayShown(host, to, text string) (string, float64, bool) {
	rv := gopodBrobotsSettings.ReporterVoices
	says := gopodReporterSays(text)
	said := says
	if sr := gopodBrobotsSettings.SignOff.SystemRole; sr != "" && host == sr {
		said = gopodSaySwap(said, rv.SystemSayWords)
	}
	_, _, made := gopodReporterCached(host, said)
	waiting := !made && rv.GeneratingMarker != ""
	if waiting {
		// 2026-10-21 (operator: "Voice bubble generating + hours glass, is
		// unstable"): the line names its voice file ("wait:<page>|text"), so the
		// page ties that voice to this line - not to whichever line came last.
		_, page := gopodReporterPaths(host, said)
		gopodChatFeedAddEventVia(host, to, "", "", rv.GeneratingMarker+"wait:"+page+"|"+text)
	}
	page, seconds, ok := gopodReporterFile(host, says)
	if !ok {
		if waiting {
			GopodChatFeedAdd(rv.Role, rv.GeneratingMarker+"off")
		}
		return page, seconds, ok
	}
	if !waiting {
		// 2026-10-21: a made voice's line names its file too ("lit:<page>|text"),
		// so its highlight is tied to that voice, never to the last line shown.
		if rv.GeneratingMarker != "" && page != "" {
			gopodChatFeedAddEventVia(host, to, "", "", rv.GeneratingMarker+"lit:"+page+"|"+text)
		} else {
			gopodChatFeedAddEventVia(host, to, "", "", text)
		}
	}
	return page, seconds, ok
}

// gopodReporterFile returns the page path of host's voice file for text and its
// length in seconds, making it first if it isn't there yet. ok is false when
// the worker isn't ready or the line couldn't be made.
func gopodReporterFile(host, text string) (string, float64, bool) {
	// Reporter 0 (2026-10-12): Goverlord's lines get its own phonetics too.
	if sr := gopodBrobotsSettings.SignOff.SystemRole; sr != "" && host == sr {
		text = gopodSaySwap(text, gopodBrobotsSettings.ReporterVoices.SystemSayWords)
	}
	if page, seconds, ok := gopodReporterCached(host, text); ok {
		return page, seconds, true
	}
	voice := gopodReporterVoice(host)
	file, page := gopodReporterPaths(host, text)
	if voice == "" || file == "" {
		return "", 0, false
	}
	// 2026-10-30 (operator: "account for voice generation times as well?"): a
	// voice being made counts as an answer going (the line queue's idle wait).
	gopodVoicesMaking.Add(1)
	defer gopodVoicesMaking.Add(-1)
	text = strings.TrimSpace(text)
	// 2026-10-17 (live 11:01: the opener's "Voice ready to play." and "Language
	// Model Loading..." were never said - asked for while Kokoro was still
	// starting, they failed at once): a line waits for the voice worker to be
	// ready (up to start_wait_seconds) and is made then.
	gopodReporterWaitReady()
	gopodReporter.Lock()
	defer gopodReporter.Unlock()
	if !gopodReporter.ready {
		return "", 0, false
	}
	_ = gopodMkdirAll(gopodReporterAudioDir(), 0755)
	fxChain, fxWet, _ := gopodReporterFX(host)
	job, _ := json.Marshal(map[string]any{"voice": voice, "text": text, "out": file, "speed": gopodReporterSpeedFor(host), "pitch_cents": gopodReporterPitch(host), "fx_chain": fxChain, "fx_wet": fxWet})
	if _, err := gopodReporter.in.Write(append(job, '\n')); err != nil || gopodReporter.in.Flush() != nil {
		gopodReporter.ready = false
		gopodReporterStarting.Store(false)
		logger.Println("GOPOD_REPORTER_VOICES worker gone")
		return "", 0, false
	}
	if !gopodReporter.out.Scan() {
		gopodReporter.ready = false
		gopodReporterStarting.Store(false)
		logger.Println("GOPOD_REPORTER_VOICES worker gone")
		return "", 0, false
	}
	var reply struct {
		OK      bool    `json:"ok"`
		Seconds float64 `json:"seconds"`
		Error   string  `json:"error"`
	}
	if json.Unmarshal(gopodReporter.out.Bytes(), &reply) != nil || !reply.OK {
		logger.Println("GOPOD_REPORTER_VOICES unable to make a line: " + reply.Error)
		return "", 0, false
	}
	// 2026-10-17 (operator: "Can you do what I do"): the new voice evened as
	// it's made - to chat_audio.loudness_lufs, held under normalize_db
	// (gopodLoudnessFile) - before anything plays it.
	if err := gopodLoudnessFile(file, file); err != nil {
		logger.Println("GOPOD_AUDIO_LEVELS " + filepath.Base(file) + ": " + err.Error())
	}
	gopodVoiceMade(host, file, text)
	// 2026-10-26 (operator: "Yes, bad rule now * clean out all stale files, rules *
	// generate a single voice file for only that instance, no dups"): the
	// 2026-10-09 rule that also made each new reporter line in the other
	// reporters' voices is gone - one line, one voice, made when it's said.
	return page, reply.Seconds, true
}

// gopodReporterWaitReady waits until the voice worker is ready, or
// start_wait_seconds have passed (it's started once at Wire-Pod's start).
func gopodReporterWaitReady() {
	limit := gopodSecondsToDuration(gopodBrobotsSettings.ReporterVoices.StartWaitSeconds)
	started := time.Now()
	for {
		gopodReporter.Lock()
		ready := gopodReporter.ready
		gopodReporter.Unlock()
		if ready || !gopodReporterStarting.Load() || time.Since(started) > limit {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// gopodReporterStarting: the voice worker is being started (not given up on).
var gopodReporterStarting atomic.Bool

// gopodReporterSpeed: reporter_voices.speed, Kokoro's talking speed (1.03 before 2026-10-09).
func gopodReporterSpeed() float64 {
	if sp := gopodBrobotsSettings.ReporterVoices.Speed; sp > 0 {
		return sp
	}
	return 1.03
}

// gopodReporterSpeedFor: host's own speed (voice_tweaks), else the shared one.
func gopodReporterSpeedFor(host string) float64 {
	if sp := gopodBrobotsSettings.ReporterVoices.VoiceTweaks[host].Speed; sp > 0 {
		return sp
	}
	return gopodReporterSpeed()
}

// gopodReporterPitch: host's pitch shift in cents (voice_tweaks), 0 for none.
func gopodReporterPitch(host string) float64 {
	return gopodBrobotsSettings.ReporterVoices.VoiceTweaks[host].PitchCents
}

// gopodReporterFX: host's robot effect (voice_tweaks robot_fx, its preset's
// chain with {carrier} filled), its wet mix, and the file-name tag; "" for none.
func gopodReporterFX(host string) (string, float64, string) {
	rv := gopodBrobotsSettings.ReporterVoices
	t := rv.VoiceTweaks[host]
	p, ok := rv.RobotFXPresets[t.RobotFX]
	if t.RobotFX == "" || !ok || strings.TrimSpace(p.Chain) == "" || p.Wet <= 0 {
		return "", 0, ""
	}
	carrier := fmt.Sprintf("%.0f", t.Carrier)
	return strings.ReplaceAll(p.Chain, "{carrier}", carrier), p.Wet, fmt.Sprintf("_fx%s%s_w%.0f", t.RobotFX, carrier, p.Wet*100)
}

// gopodReporterPaths: the voice file for host's line, on disk and on the page
// - named by its voice and words, so the same line is always the same file.
// gopodMusicSting (2026-10-13, then 2026-10-25): the sting a "(music)" plays, and a
// warm-up's end - reporter_voices.system_stings.default (djent), and its length.
// "" when there's none.
func gopodMusicSting() (string, float64) {
	s, ok := gopodBrobotsSettings.ReporterVoices.SystemStings["default"]
	if !ok {
		return "", 0
	}
	return gopodStingPage(s.File, s.Volume)
}

// gopodSystemSting (2026-10-25, operator: "[Fart joke started ...] & [Fart joke
// complete ...] * there's a grunge audio play in there somewhere * no error label *
// remove grunge audio * default should = djent, but muted"): a Goverlord bubble's
// own sound, from the bubble sounds template (bubble_sounds.Goverlord: its file,
// else default_file; none when muted) - before 2026-10-25 djent before every system
// line and grunge before a run's line. "" when there's none.
func gopodSystemSting(text string) (string, float64) {
	bs := gopodBrobotsSettings.BubbleSounds
	if bs.Goverlord.Mute {
		return "", 0
	}
	file := strings.TrimSpace(bs.Goverlord.File)
	if file == "" {
		file = strings.TrimSpace(bs.DefaultFile)
	}
	return gopodStingPage(filepath.Base(file), bs.Goverlord.Volume)
}

// gopodStingPage: a sting file (in the voice folder) as a page line, with its
// volume share, and its length; "" when it isn't there.
func gopodStingPage(file string, volume json.RawMessage) (string, float64) {
	rv := gopodBrobotsSettings.ReporterVoices
	if strings.TrimSpace(file) == "" {
		return "", 0
	}
	seconds, ok := gopodWavSeconds(filepath.Join(gopodReporterAudioDir(), file))
	if !ok {
		return "", 0
	}
	page := "/" + rv.AudioFolder + "/" + file
	if share, _ := gopodVolumePercent(volume); share != 1 {
		page += fmt.Sprintf("#vol=%g", share)
	}
	return page, seconds
}

// gopodIsRunLine: text is a run's started/stopped/complete/paused/continues
// line ("{run}" any name), or the stop-run answer when nothing is running.
func gopodIsRunLine(text string) bool {
	ss := gopodBrobotsSettings.Session
	text = strings.TrimSpace(text)
	if text != "" && text == strings.TrimSpace(ss.NoRunText) {
		return true
	}
	for _, tmpl := range []string{ss.StartedText, ss.StoppedText, ss.CompleteText, ss.PausedText, ss.ContinuesText} {
		tmpl = strings.TrimSpace(tmpl)
		if tmpl == "" {
			continue
		}
		pattern := "^" + strings.ReplaceAll(regexp.QuoteMeta(tmpl), regexp.QuoteMeta("{run}"), ".+?") + "$"
		if re, err := regexp.Compile(pattern); err == nil && re.MatchString(text) {
			return true
		}
	}
	return false
}

// gopodSystemStingNow: the sting before this system line - during a warm-up,
// only its first line's (the default sting), else gopodSystemSting.
func gopodSystemStingNow(text string) (string, float64) {
	if gopodWarmingStarted() {
		if gopodWarmStingStart.CompareAndSwap(true, false) {
			return gopodSystemSting("")
		}
		return "", 0
	}
	return gopodSystemSting(text)
}

// gopodNoSting: the system line being said follows one that already had its
// sting (one sting per bubble: "Warm-up complete. ..." after "LLM Ready").
var gopodNoSting atomic.Bool

// gopodSystemSayNoSting says a system line with no sting of its own.
func gopodSystemSayNoSting(text string, wait bool) {
	gopodNoSting.Store(true)
	defer gopodNoSting.Store(false)
	gopodSystemSay(text, wait)
}

// The voice files this Wire-Pod start uses (gopodReporterPaths), kept when the
// folder is cleaned (gopodReporterPrune).
var gopodReporterInUse = struct {
	sync.Mutex
	files map[string]bool
}{files: map[string]bool{}}

// gopodReporterPrune (2026-10-13, operator: "Please clean up old/stale voices
// files * probably getting cluttered * keep only most recent"; then "system,
// reporter, narrator, kept-sessions, etc..., (scripted) voices are typically
// generated once, play many * when I make edits, that new voice segment
// replaces the current one * Treat operator voice files a bit different ...
// Give these files a month to settle"): once every line is made at Wire-Pod's
// start, ... 2026-10-18 (operator: "no auto-deletes"): nothing is deleted any
// more - a voice not used for its kind's keep days is archived
// (gopodVoicesArchive, gopod_system_voices.go).
func gopodReporterPrune() {
	gopodVoicesArchive()
}

func gopodReporterPaths(host, text string) (string, string) {
	rv := gopodBrobotsSettings.ReporterVoices
	voice := gopodReporterVoice(host)
	text = strings.TrimSpace(text)
	if voice == "" || text == "" {
		return "", ""
	}
	// 2026-10-09 (operator: "upon new voice like now, ensure there's a
	// distinction so I will have wav of all generated voices without over
	// writing"): the name carries the voice and the speed, so a new voice or
	// speed makes new files and every earlier one stays. Speed 1.03 (the speed
	// before speed was a setting) keeps the old names, so those files still count.
	speed := gopodReporterSpeedFor(host)
	key, tag := voice+"\n"+text, voice
	if speed != 1.03 {
		key += fmt.Sprintf("\n%.2f", speed)
		tag += fmt.Sprintf("_s%03.0f", speed*100)
	}
	// 2026-10-26: a pitch-shifted voice (voice_tweaks) is its own file too.
	if pitch := gopodReporterPitch(host); pitch != 0 {
		key += fmt.Sprintf("\np%.0f", pitch)
		tag += fmt.Sprintf("_p%+.0f", pitch)
	}
	// 2026-10-28: a robot-effect voice is its own file too.
	if chain, _, fxTag := gopodReporterFX(host); chain != "" {
		key += "\n" + fxTag + "\n" + chain
		tag += fxTag
	}
	sum := sha1.Sum([]byte(key))
	name := tag + "_" + hex.EncodeToString(sum[:8]) + ".wav"
	gopodReporterInUse.Lock()
	gopodReporterInUse.files[name] = true
	gopodReporterInUse.Unlock()
	return filepath.Join(gopodReporterAudioDir(), name), "/" + rv.AudioFolder + "/" + name
}

// gopodReporterCached returns host's line's voice file if it's already made.
func gopodReporterCached(host, text string) (string, float64, bool) {
	file, page := gopodReporterPaths(host, text)
	if file == "" {
		return "", 0, false
	}
	seconds, ok := gopodWavSeconds(file)
	// 2026-10-18 (operator: "Forced alignment ... existing voices need remaking *
	// new ones get them automatically"): a voice without its word times
	// (<name>.words.json, the voice worker's) is made again.
	// 2026-10-18 later (operator: "can GOPOD redo voices as per your timings?"):
	// word times from before the heard-timing (no "aligner") are made again too.
	if ok {
		if b, err := os.ReadFile(strings.TrimSuffix(file, ".wav") + ".words.json"); err != nil || !strings.Contains(string(b), `"aligner"`) {
			ok = false
		} else if !strings.Contains(string(b), `"aligner": "vosk"`) && gopodFileExists(gopodHomePath("~/wire-pod/vosk/models/en-US/model")) {
			// 2026-10-18 (operator: "forced highlighting (CHALK & PLAYHEAD) ...
			// not in sync like the other highlightings"): timed by Kokoro's
			// estimate (the worker, as root, missed Vosk) - made again, heard.
			ok = false
		}
	}
	// A line said again is kept another keep_days (its kind's, gopodVoicesArchive).
	if ok {
		now := time.Now()
		_ = os.Chtimes(file, now, now)
	}
	return page, seconds, ok
}

// gopodWavSeconds reads a plain PCM WAV's length from its header.
func gopodWavSeconds(file string) (float64, bool) {
	f, err := os.Open(file)
	if err != nil {
		return 0, false
	}
	defer f.Close()
	head := make([]byte, 44)
	if n, _ := f.Read(head); n < 44 || string(head[0:4]) != "RIFF" {
		return 0, false
	}
	byteRate := int(head[28]) | int(head[29])<<8 | int(head[30])<<16 | int(head[31])<<24
	info, err := f.Stat()
	if err != nil || byteRate <= 0 {
		return 0, false
	}
	return float64(info.Size()-44) / float64(byteRate), true
}

// GopodReporterSay: a reporter's line, its bubble and voice; returns once played.
func GopodReporterSay(host, text string) bool { return gopodReporterSay(host, text, true, true) }

// gopodDryVoice: set only by the session dry run (gopod_session_dry_test.go) -
// a voice line there is looked up in the library and logged, never played.
var gopodDryVoice func(host, text string, show bool)

// gopodReporterSay plays host's line through the page; with wait, it returns
// once the line has been heard (its length, plus the page's poll lag, up to
// max_wait_seconds). show also puts the line in the chat as the host's bubble.
// The page plays the voice lines one after another, so the bookends don't wait.
func gopodReporterSay(host, text string, show, wait bool) bool {
	return gopodReporterSayTo(host, "", text, show, wait)
}

// gopodSayForShownLine (2026-10-23, operator: "RULE: Every voiced line gets
// generating+hourglass/voice per-word highlighting * when any voiced line is
// created, the f*ck*ng template applies!!"): a voice whose words are already on
// the page (a warm-up/opener line, an egg, the Operator's own line) - its line
// is named with its voice file, so the page ties them: "forwait:<page>|text"
// (Generating... and the hourglass while it's made) or "forlit:<page>|text"
// (made), then its words lit word by word as it plays. The same template as
// gopodSayShown, for a line shown some other way.
func gopodSayForShownLine(host, text string) (string, float64, bool) {
	rv := gopodBrobotsSettings.ReporterVoices
	says := gopodReporterSays(text)
	said := says
	if sr := gopodBrobotsSettings.SignOff.SystemRole; sr != "" && host == sr {
		said = gopodSaySwap(said, rv.SystemSayWords)
	}
	cachedPage, _, made := gopodReporterCached(host, said)
	shown := strings.TrimSpace(text)
	if rv.GeneratingMarker != "" && shown != "" {
		if made {
			GopodChatFeedAdd(rv.Role, rv.GeneratingMarker+"forlit:"+cachedPage+"|"+shown)
		} else {
			_, page := gopodReporterPaths(host, said)
			GopodChatFeedAdd(rv.Role, rv.GeneratingMarker+"forwait:"+page+"|"+shown)
		}
	}
	page, seconds, ok := gopodReporterFile(host, says)
	if !ok && !made && rv.GeneratingMarker != "" {
		GopodChatFeedAdd(rv.Role, rv.GeneratingMarker+"off")
	}
	return page, seconds, ok
}

// gopodReporterSayTo: gopodReporterSay with its bubble's "to" - a manager's line
// for a Brobot shows inside that Brobot's bubble (2026-10-18, gopod_managers.go).
func gopodReporterSayTo(host, to, text string, show, wait bool) bool {
	var page string
	var seconds float64
	var ok bool
	if show {
		page, seconds, ok = gopodSayShown(host, to, text)
	} else {
		page, seconds, ok = gopodSayForShownLine(host, text)
	}
	if !ok {
		if gopodDryVoice != nil {
			gopodDryVoice(host, text, show)
		}
		// 2026-10-17 (survey E5: "heard timeouts leave no trace"): said so.
		logger.Println("GOPOD_HEARD " + host + ": no voice file - not said: " + gopodHeardShort(text))
		return false
	}
	return gopodVoicePlay(host, text, page, seconds, wait)
}

// gopodVoicePlay (2026-10-24: shared, so a made voice and a recorded one - the
// mic's, gopod_mic_voice.go - play the one way): a voice file to the page, a
// system line's sting around it, in the voice queue, and - wait - on until the
// page has played it.
func gopodVoicePlay(host, text, page string, seconds float64, wait bool) bool {
	rv := gopodBrobotsSettings.ReporterVoices
	// 2026-10-13: a system line's sting first, then the speech (gopodSystemSting).
	// 2026-10-17 (operator: "new rule: if \"(music)\" is in the story.md *
	// display \"(music)\" visual for the deaf * play djent"): a line that starts
	// with (music) has the sting first, one that ends with it has the sting
	// after its words - wherever it's written, that's where it plays.
	system := host == gopodBrobotsSettings.SignOff.SystemRole && host != ""
	musicFirst, musicLast := gopodMusicMarks(text)
	if system && (musicFirst || (!musicLast && !gopodNoSting.Load())) {
		sting, stingSeconds := gopodSystemStingNow(text)
		if musicFirst {
			sting, stingSeconds = gopodMusicSting()
		}
		if sting != "" {
			GopodChatFeedAdd(rv.Role, gopodWithVolumeShare(sting))
			seconds += stingSeconds + rv.GapSeconds
		}
	}
	GopodChatFeedAdd(rv.Role, gopodWithVolumeShare(page))
	if system && musicLast {
		if sting, stingSeconds := gopodMusicSting(); sting != "" {
			GopodChatFeedAdd(rv.Role, gopodWithVolumeShare(sting))
			seconds += stingSeconds + rv.GapSeconds
			gopodMusicPlayed.Store(true)
		}
	}
	// The page plays the voice lines one after another, so this line ends
	// after every line queued before it (2026-10-08, live 23:09: Goverlord's
	// "Chapter: ..." played first, the reporter's question after it, and the
	// Brobot - who waited for the question's own length only - started while
	// the question was still being read).
	ends := gopodVoiceQueueAdd(seconds + rv.GapSeconds)
	if !wait {
		return true
	}
	// 2026-10-14 (operator: "make GOPOD respond the lines' own length plus a few
	// seconds * no more estimates of time padding please, or as minimal as
	// needed * more responsive"): it goes on the moment the page has played
	// the line (the heard check); with no answer, after the lines' own length
	// plus heard_margin_seconds.
	// 2026-10-17 (survey E5): heard or not, the journal says so.
	limit := gopodHeardLimit(ends)
	if gopodWaitHeard(limit) {
		logger.Println("GOPOD_HEARD " + host + ": heard: " + gopodHeardShort(text))
		return true
	}
	logger.Println("GOPOD_HEARD " + host + ": not heard after " + limit.Round(time.Second).String() + " (going on): " + gopodHeardShort(text))
	if rv.HeardMarker == "" {
		time.Sleep(limit)
	}
	return false
}

// gopodMusicMark: the deaf viewers' sign of the sting, as written in a line.
const gopodMusicMark = "(music)"

// gopodMusicMarks: the line starts / ends with (music).
func gopodMusicMarks(text string) (first, last bool) {
	t := strings.TrimSpace(text)
	return strings.HasPrefix(t, gopodMusicMark), len(t) > len(gopodMusicMark) && strings.HasSuffix(t, gopodMusicMark)
}

// gopodMusicPlayed: a line's own end (music) has played the sting, so the
// warm-up's end sting doesn't play it again (gopodWarmupEndSting).
var gopodMusicPlayed atomic.Bool

// gopodHeardShort: a line's first words, for the journal.
func gopodHeardShort(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	if len(text) > 60 {
		text = gopodCut(text, 60)
	}
	return text
}

// gopodHeardLimit: how long a heard check waits with no answer - the lines
// queued until ends (their own lengths), plus reporter_voices.heard_margin_seconds;
// never more than 2 x max_wait_seconds.
func gopodHeardLimit(ends time.Time) time.Duration {
	rv := gopodBrobotsSettings.ReporterVoices
	limit := time.Until(ends) + gopodSecondsToDuration(rv.HeardMarginSeconds)
	// 2026-10-20 (operator: "The long response voice file was played/heard, but
	// GOPOD continued before the file finished playing * add a 'wait until I'm
	// finished' flag"; live 05:56: a long Operator voice was still queued on the
	// page when the flat 30 s cap ran out, and the eggs and Brobot 1 went on):
	// with finish_wait_seconds, a line waits its own length plus that much for
	// the page to finish it - "heard" usually ends the wait long before.
	// Only while the chat page is here to play it - with no page, nothing is
	// playing, and the old cap holds.
	if fw := rv.FinishWaitSeconds; fw > 0 && gopodPageWatchOn.Load() && gopodPageHere() {
		if l := time.Until(ends) + gopodSecondsToDuration(fw); l > limit {
			limit = l
		}
	} else if most := gopodSecondsToDuration(2 * rv.MaxWaitSeconds); rv.MaxWaitSeconds > 0 && limit > most {
		limit = most
	}
	if limit < 0 {
		limit = 0
	}
	return limit
}

// gopodVoiceQueue: when the page's voice queue runs out - each line starts
// once the lines before it have played.
var gopodVoiceQueue = struct {
	sync.Mutex
	ends time.Time
}{}

// gopodVoiceQueueAdd adds a line this long to the queue; it returns when that
// line will have played.
func gopodVoiceQueueAdd(seconds float64) time.Time {
	gopodVoiceQueue.Lock()
	defer gopodVoiceQueue.Unlock()
	start := time.Now()
	if gopodVoiceQueue.ends.After(start) {
		start = gopodVoiceQueue.ends
	}
	gopodVoiceQueue.ends = start.Add(gopodSecondsToDuration(seconds))
	return gopodVoiceQueue.ends
}

// gopodReporterReadsOperator reads an operator line out, when how (e.g. "Typed
// text") is one of read_lines_from: a line to one Brobot in his reporter's voice
// (read_by; operator 2026-10-06: "assign to brobot 1, reporter 1, brobot 3,
// female voice / assign to brobot 2, reporter 2, brobot 4, male voice"), a line
// to both in operator_host's.
// gopodOperatorVoiceHost: the Operator voice's name among the voices.
const gopodOperatorVoiceHost = "Operator"

func gopodReporterReadsOperator(text, how, to string) {
	rv := gopodBrobotsSettings.ReporterVoices
	host := rv.ReadBy[to]
	// 2026-10-08: a playing draft can name the question's reporter (read_by).
	if gopodScriptPlaying.Load() {
		if h := gopodScriptTakeReader(); h != "" {
			host = h
		}
	}
	if host == "" {
		host = rv.OperatorHost
	}
	// 2026-10-13 (operator: "Currently, inputs to brobots are live generated
	// and reiterated/said as either reporter voice 1 or 2 * undo remove this
	// stale mapping * remap to simply Operator bubble & new Operator voice";
	// "Let's try af_sky"): a live line (keyboard, numpad) is read in the
	// Operator voice (reporter_voices.operator_voice). A playing draft's question
	// keeps its reporter.
	if !gopodScriptPlaying.Load() && gopodReporterVoice(gopodOperatorVoiceHost) != "" {
		host = gopodOperatorVoiceHost
	}
	for _, from := range rv.ReadLinesFrom {
		if from == how {
			// 2026-10-24: a numpad line plays its own mic recording, evened and
			// timed like a made voice (gopod_mic_voice.go) - the Operator's made
			// voice only when there's none.
			if gopodMicVoiceSay(host, text) {
				return
			}
			// 2026-10-16 (operator: "Operator bubble (as voice file is being
			// generated) * update message to \"Voice file generating.\" +
			// hourglass * update message upon voice wav play"); 2026-10-23: the
			// voice template does it (gopodSayForShownLine) - Generating and
			// the hourglass while it's made, then lit word by word.
			gopodReporterSay(host, text, false, true)
			return
		}
	}
}

// gopodReporterBookend plays the opening or closing lines, each host in turn.
func gopodReporterBookend(lines []gopodHostLine) {
	for _, l := range lines {
		gopodReporterSay(l.Host, l.Text, true, false)
	}
}

// gopodReporterOpening queues the opening for the page once its voice files are
// made (Kokoro loads in about a minute at Wire-Pod start; made at an earlier
// start, no wait), then a reporter_voices.opening_end_text line, and returns
// when the page says it has played it all (GopodOpeningHeard) - or after
// gate_max_seconds with no page. The page plays nothing until its Log section
// is open and has been clicked (operator 2026-10-06: "pause the chat bubbles
// from running until its clicked to and viewable ... then start the sequence
// running whatever's been queued then continue").
func gopodReporterOpening(warm bool) {
	rv := gopodBrobotsSettings.ReporterVoices
	opening := gopodBookends(true)
	if len(opening) == 0 {
		return
	}
	if !gopodReporterAllMade(opening) {
		select {
		case <-gopodReporterBookendsMade:
		case <-time.After(gopodSecondsToDuration(rv.StartWaitSeconds)):
			logger.Println("GOPOD_REPORTER_VOICES opening skipped: voices not ready")
			return
		}
	}
	// 2026-10-08: the system voice says the LLM status first - only while it
	// warms (2026-10-13: a warm start skips the warming part).
	if !warm {
		gopodSystemSay(gopodWarmupLine(gopodBrobotsSettings.LLM.WarmupText), false)
	}
	gopodReporterBookend(opening)
	GopodChatFeedAdd(rv.Role, rv.OpeningEndText)
	logger.Println("GOPOD_REPORTER_VOICES opening queued for the page")
	var timeout <-chan time.Time
	if rv.GateMaxSeconds > 0 {
		timeout = time.After(gopodSecondsToDuration(rv.GateMaxSeconds))
	}
	// While it waits, a page that's away (or not clicked yet, after its own
	// reload) gets the awaits line in the System llm bubble (gopodOpenerAwaits).
	defer gopodOpenerAwaits(false)
	for {
		select {
		case <-gopodOpeningHeard:
			logger.Println("GOPOD_REPORTER_VOICES opening heard on the page")
			return
		case <-timeout:
			logger.Println("GOPOD_REPORTER_VOICES opening not heard: no page after gate_max_seconds")
			return
		case <-time.After(time.Second):
			gopodOpenerAwaits(!gopodPageHere() && gopodLLMWarmupActive)
		}
	}
}

var gopodOpeningHeard = make(chan struct{})
var gopodOpeningHeardOnce sync.Once

// GopodOpeningHeard: the chat page has played the opening (POST
// /api-sdk/gopod_opening_heard).
func GopodOpeningHeard() {
	gopodOpeningHeardOnce.Do(func() { close(gopodOpeningHeard) })
}

// gopodReporterClosing plays the closing after "Thanks Brobots" - once, even
// when both Brobots answer the thanks, after both have finished their sign-off.
// Its place in the ending is gopodThanksEnding's (gopod_thanks_close_out.go,
// ALIAS-PLAY-STUDIO.md "The thanks ending"); every caller returns once the
// closing is queued.
func gopodReporterClosing() {
	gopodReporter.Lock()
	// Once per ending: a second "thanks" - the brother's, or one that only
	// finishes after a reboot (live 2026-10-06 15:58/15:59) - doesn't play it
	// again until a new question has been asked (gopodReporterNewQuestion).
	if gopodReporter.closing != nil {
		done := gopodReporter.closing
		gopodReporter.Unlock()
		<-done
		return
	}
	done := make(chan struct{})
	gopodReporter.closing = done
	gopodReporter.Unlock()
	defer close(done)
	for i := 0; i < 60 && gopodReplyBusy(); i++ {
		time.Sleep(time.Second)
	}
	// Each line waits until heard: the brothers' lines come after it.
	for _, l := range gopodBookends(false) {
		gopodReporterSay(l.Host, l.Text, true, true)
	}
}

// gopodReporterNewQuestion: a question that isn't a thanks - the next ending
// gets its own closing.
func gopodReporterNewQuestion() {
	gopodReporter.Lock()
	gopodReporter.closing = nil
	gopodReporter.signedOff = false
	gopodReporter.ended = false
	gopodReporter.Unlock()
}

// gopodDraftRunCompleted: after the ending of a session that played a draft,
// sign_off.system_text in a Goverlord system bubble - once.
func gopodDraftRunCompleted() {
	so := gopodBrobotsSettings.SignOff
	if so.SystemRole == "" || so.SystemText == "" || !gopodScriptRan.CompareAndSwap(true, false) {
		return
	}
	// 2026-10-23: the one voiced-line template (show and say: Generating, then lit).
	gopodReporterSay(so.SystemRole, so.SystemText, true, true)
}

// gopodSystemSay says a system line in the system voice (reporter_voices.
// system_voice) through the page, with no bubble of its own (operator
// 2026-10-08: "Third voice * can speak llm status messages (system) *
// chapters ... upon 'start draft' add in a system draft started message").
func gopodSystemSay(text string, wait bool) bool {
	if role := gopodBrobotsSettings.SignOff.SystemRole; role != "" && strings.TrimSpace(text) != "" {
		return gopodReporterSay(role, text, false, wait)
	}
	return false
}

// gopodSystemLines: every line the system voice says, made at start.
func gopodSystemLines() []string {
	s := gopodBrobotsSettings
	lines := []string{s.LLM.WarmupText, gopodWarmupLine(s.LLM.WarmupText), s.LLM.ReadyText, s.LLM.WarmupDoneText, s.SignOff.SystemText, s.SignOff.StartedText, s.SignOff.WelcomeText, s.TopicPictures.PopupText, gopodWarmupLine(s.LLM.VoicesText)}
	// 2026-10-16: the opener song's lines.
	lines = append(lines, gopodOpenerSaid()...)
	for _, ch := range s.InterviewChapters.Chapters {
		lines = append(lines, gopodChapterSaid(ch.Title))
	}
	// 2026-10-11: the session's run lines, each run by name, ready before a run starts.
	rs := s.RecordSession
	lines = append(lines, rs.KeptText, rs.DeletedText, rs.KeptText+" "+s.SignOff.WelcomeText, rs.DeletedText+" "+s.SignOff.WelcomeText, strings.TrimSpace(s.SignOff.NothingRecordedText+" "+s.SignOff.WelcomeText))
	// 2026-10-14: the numpad's spoken on/off lines (the page fills in {how}).
	// 2026-10-19: Native? and Numpad? are system commands now (answered in the
	// terminal bubble, not said); 3x NumLock / 000 still says the numpad's exit.
	lines = append(lines, s.NumpadPTT.Texts["exit"])
	sess := s.Session
	lines = append(lines, sess.BusyText, sess.RunGateText, sess.NoRunText, gopodRunText(sess.StoppedText, ""))
	for name := range sess.RunNames {
		lines = append(lines, gopodRunText(sess.StartedText, name), gopodRunText(sess.StoppedText, name), gopodRunText(sess.CompleteText, name))
	}
	return lines
}

// gopodChapterSaid: a chapter title as its bubble's header shows it ("Chapter: {title}").
func gopodChapterSaid(title string) string {
	label := gopodBrobotsSettings.InterviewChapters.Label
	if label == "" {
		return title
	}
	return strings.ReplaceAll(label, "{title}", title)
}

// gopodReporterAllMade: every line's voice file is already there.
func gopodReporterAllMade(lines []gopodHostLine) bool {
	for _, l := range lines {
		// 2026-10-16 (comb-through: the file is made from the line as said -
		// "live" -> "lyve" - so the check looks for that).
		if _, _, ok := gopodReporterCached(l.Host, gopodReporterSays(l.Text)); !ok {
			return false
		}
	}
	return len(lines) > 0
}

// gopodReporterSays: the line as the reporter says it - reporter_voices.say_words
// replaced, whole words, any case (operator 2026-10-06: "reporter word
// replacements ... thx = thanks"). The bubble keeps what was typed.
func gopodReporterSays(text string) string {
	// 2026-10-18 (operator, on CHALK's weather line: "canada spoken 2x"): a flag
	// right after its country isn't said again (gopodSpeakFlags).
	// The same rules as the Brobots' speech for a time: "1405 Hours Military
	// Time" said "fourteen oh five hours military time" (gopodSpeakMilitaryTime).
	return gopodSaySwap(gopodSpeakMilitaryTime(gopodSpeakFlags(text)), gopodBrobotsSettings.ReporterVoices.SayWords)
}

// gopodSaySwap: each whole word in words (any case) said as its value.
func gopodSaySwap(text string, words map[string]string) string {
	for from, to := range words {
		if strings.TrimSpace(from) == "" {
			continue
		}
		re := regexp.MustCompile(`(?i)(^|[^\pL\pN_])` + regexp.QuoteMeta(from) + `($|[^\pL\pN_])`)
		// Twice: back-to-back words ("thx thx") share the space between them.
		for i := 0; i < 2; i++ {
			text = re.ReplaceAllString(text, "${1}"+strings.ReplaceAll(to, "$", "$$")+"${2}")
		}
	}
	return text
}

// gopodReporterAllLines: every scripted line a voice may say, with the one
// voice that says it - reporter_voices' opening and closing, the interview's
// start lines, the interview's banter, handoffs and questions, and the run
// commands (Operator voice).
func gopodReporterAllLines() []gopodHostLine {
	rv := gopodBrobotsSettings.ReporterVoices
	var out []gopodHostLine
	seen := map[string]bool{}
	add := func(host, text string) {
		text = strings.TrimSpace(text)
		if host == "" || text == "" || seen[host+"\n"+text] {
			return
		}
		seen[host+"\n"+text] = true
		out = append(out, gopodHostLine{Host: host, Text: text})
	}
	for _, l := range append(append([]gopodHostLine{}, rv.Opening...), rv.Closing...) {
		add(l.Host, l.Text)
	}
	for _, l := range gopodBrobotsSettings.InterviewScript.StartLines {
		add(gopodReporterHost(l.Who), l.Text)
	}
	// The interview that "start interview" plays: its song folder, else the
	// newest draft.
	var draft gopodDraft
	var err error
	if dir, ferr := gopodFindSong(gopodBrobotsSettings.InterviewScript.SongID); ferr == nil {
		draft, err = gopodInterviewSongDraft(dir)
	} else {
		_, draft, err = gopodLoadDraft()
	}
	if err == nil {
		for _, l := range append(append([]gopodDraftLine{}, draft.Opening...), draft.Closing...) {
			add(gopodReporterHost(l.Who), l.Text)
		}
		for _, q := range draft.Questions {
			for _, l := range q.Before {
				add(gopodReporterHost(l.Who), l.Text)
			}
			for _, a := range q.Answers {
				add(gopodReporterHost(a.Brobot), a.Say)
			}
			reader := gopodReporterHost(q.ReadBy)
			if reader == "" {
				reader = rv.ReadBy[q.To]
			}
			add(reader, q.Ask)
		}
	}
	// Each song's own warm-up: its reporters' warmup_lines and its ready line
	// (Goverlord's), said at its start.
	if songs, err := gopodAllSongs(); err == nil {
		for _, dir := range songs {
			data, err := os.ReadFile(filepath.Join(dir, "knobs.json"))
			if err != nil {
				continue
			}
			var k struct {
				WarmupLines []gopodDraftLine `json:"warmup_lines"`
				ReadyText   string           `json:"llm_ready_text"`
			}
			if json.Unmarshal(data, &k) != nil {
				continue
			}
			// 2026-10-20: the song's _llm_song run, when it has one.
			if lines, found := gopodLLMSongLines(dir); found {
				for _, l := range lines {
					add(l.Host, l.Text)
				}
			} else {
				for _, l := range k.WarmupLines {
					add(gopodReporterHost(l.Who), l.Text)
				}
			}
			if k.ReadyText != "" {
				add(gopodBrobotsSettings.SignOff.SystemRole, gopodRunReadyLine(k.ReadyText))
			}
		}
	}
	// The run commands, read out in the Operator voice when typed.
	if rv.OperatorVoice != "" {
		for _, text := range gopodRunCommandLines() {
			add(gopodOperatorVoiceHost, text)
		}
	}
	return out
}

// gopodRunCommandLines: the start and stop commands - each run command's first
// phrase (customIntents.json) and song_stop.stop_words.
func gopodRunCommandLines() []string {
	var out []string
	if vars.CustomIntentsExist {
		for _, c := range vars.CustomIntents {
			if c.IsSystemIntent || c.LuaScript != "" || c.Exec == "" || len(c.Utterances) == 0 {
				continue
			}
			if u := strings.TrimSpace(c.Utterances[0]); u != "" {
				out = append(out, u)
			}
		}
	}
	for _, w := range gopodBrobotsSettings.SongStop.StopWords {
		if strings.TrimSpace(w) != "" {
			out = append(out, strings.TrimSpace(w))
		}
	}
	return out
}

// gopodSignOffLines: after "thanks Brobots", once the sign-off has been said,
// the Brobots' sign_off.lines, each in turn - once per ending (operator
// 2026-10-08: "after the very end where brobot 2 signed off, add in brobot 2
// says, this is where Pip asks for an email, right? brobot 1 says, bro!").
func gopodSignOffLines() {
	gopodReporter.Lock()
	if gopodReporter.signedOff {
		gopodReporter.Unlock()
		return
	}
	gopodReporter.signedOff = true
	gopodReporter.Unlock()
	for i := 0; i < 60 && gopodReplyBusy(); i++ {
		time.Sleep(time.Second)
	}
	for _, l := range gopodBrobotsSettings.SignOff.Lines {
		if host := gopodReporterHost(l.Who); host != "" {
			gopodReporterSay(host, l.Text, true, true)
		} else if gopodBrobotESN(l.Who) != "" {
			gopodRecordSay(l.Who, l.Text)
		}
	}
}

// ## EOF

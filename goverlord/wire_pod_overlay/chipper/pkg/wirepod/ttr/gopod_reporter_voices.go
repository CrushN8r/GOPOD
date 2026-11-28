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
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

var gopodReporter = struct {
	sync.Mutex
	in        *bufio.Writer
	out       *bufio.Scanner
	ready     bool
	closing   chan struct{} // this ending's closing; closed once it's queued
	signedOff bool          // this ending's sign_off.lines are said
	ended     bool          // this ending is running or done (gopodThanksEnding)
}{}

// gopodReporterBookendsMade is closed once the opening and closing lines' voice files are made (or
// the voices can't start), so the opening never waits on a line still being made.
var gopodReporterBookendsMade = make(chan struct{})
var gopodReporterBookendsOnce sync.Once

func gopodReporterBookendsDone() {
	gopodReporterBookendsOnce.Do(func() { close(gopodReporterBookendsMade) })
}

// gopodReporterAudioDir is where the voice files go, inside the page's webroot.
func gopodReporterAudioDir() string {
	return filepath.Join(gopodChipperRoot(), "webroot", gopodBrobotsSettings.ReporterVoices.AudioFolder)
}

// GopodReporterVoicesStart starts the Kokoro worker and makes the bookends' voice files, in the
// background (called once at Wire-Pod start).
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
		gopodReporterBookendsDone()
		for _, l := range gopodBookends(false) {
			gopodReporterFile(l.Host, gopodSaysFor(l.Host, l.Text))
		}
		for _, text := range gopodSystemLines() {
			if strings.TrimSpace(text) != "" {
				gopodReporterFile(gopodBrobotsSettings.SignOff.SystemRole, gopodSaysFor(gopodBrobotsSettings.SignOff.SystemRole, text))
			}
		}
		gopodReporterBookendsDone()
		made := 0
		for _, l := range gopodReporterAllLines() {
			if _, _, ok := gopodReporterCached(l.Host, gopodSaysFor(l.Host, l.Text)); ok {
				continue
			}
			if _, _, ok := gopodReporterFile(l.Host, gopodSaysFor(l.Host, l.Text)); ok {
				made++
			}
		}
		gopodVoicesAtStart.Store(false)
		logger.Println("GOPOD_REPORTER_VOICES all lines ready, made " + strconv.Itoa(made) + " new")
		gopodReporterPrune()
		go func() { _ = cmd.Wait() }()
	}()
}

// gopodReporterVoice: a reporter's Kokoro voice, or the system voice for the system role.
func gopodReporterVoice(host string) string {
	rv := gopodBrobotsSettings.ReporterVoices
	if of := strings.TrimSpace(gopodBrobotsSettings.Managers.Personas[host].VoiceOf); of != "" && of != host {
		return gopodReporterVoice(of)
	}
	if GopodRobotsOff() && gopodIsBodiedBrobot(host) {
		if voice := gopodBrobotsSettings.RobotsOff.Voices[host]; voice != "" {
			return voice
		}
		return rv.Hosts[rv.UnassignedVoiceHost]
	}
	if voice := rv.Hosts[host]; voice != "" {
		return voice
	}
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

// GOLDEN global message: a shown line whose voice isn't made yet goes to the page first, marked
// waiting (generating_marker+"wait:"+text) - its own bubble shows "Generating..." and the hourglass
// (message0) until its voice plays, then its words (message1).
func gopodSayShown(host, to, text string) (string, float64, bool) {
	rv := gopodBrobotsSettings.ReporterVoices
	says := gopodSaysFor(host, text)
	said := says
	_, _, made := gopodReporterCached(host, said)
	waiting := !made && rv.GeneratingMarker != ""
	if waiting {
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
		if rv.GeneratingMarker != "" && page != "" {
			gopodChatFeedAddEventVia(host, to, "", "", rv.GeneratingMarker+"lit:"+page+"|"+text)
		} else {
			gopodChatFeedAddEventVia(host, to, "", "", text)
		}
	}
	return page, seconds, ok
}

// gopodReporterFile returns the page path of host's voice file for text and its length in seconds,
// making it first if it isn't there yet.
func gopodReporterFile(host, text string) (string, float64, bool) {
	if page, seconds, ok := gopodReporterCached(host, text); ok {
		return page, seconds, true
	}
	voice := gopodReporterVoice(host)
	file, page := gopodReporterPaths(host, text)
	if voice == "" || file == "" {
		return "", 0, false
	}
	gopodVoicesMaking.Add(1)
	defer gopodVoicesMaking.Add(-1)
	text = strings.TrimSpace(text)
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
	if err := gopodLoudnessFile(file, file); err != nil {
		logger.Println("GOPOD_AUDIO_LEVELS " + filepath.Base(file) + ": " + err.Error())
	}
	gopodVoiceMade(host, file, text)
	return page, reply.Seconds, true
}

// gopodReporterWaitReady waits until the voice worker is ready, or start_wait_seconds have passed
// (it's started once at Wire-Pod's start).
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

// gopodReporterSpeed: reporter_voices.speed, Kokoro's talking speed.
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

// gopodReporterFX: host's robot effect (voice_tweaks robot_fx, its preset's chain with {carrier}
// filled), its wet mix, and the file-name tag; "" for none.
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

// gopodReporterPaths: the voice file for host's line, on disk and on the page - named by its voice
// and words, so the same line is always the same file.
func gopodMusicSting() (string, float64) {
	s, ok := gopodBrobotsSettings.ReporterVoices.SystemStings["default"]
	if !ok {
		return "", 0
	}
	return gopodStingPage(s.File, s.Volume)
}

// gopodSystemSting: a Goverlord bubble's own sound, from the bubble sounds template
// (bubble_sounds.Goverlord: its file, else default_file; none when muted)
// before every system line and grunge before a run's line.
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

// gopodStingPage: a sting file (in the voice folder) as a page line, with its volume share, and its
// length; "" when it isn't there.
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

// gopodIsRunLine: text is a run's started/stopped/complete/paused/continues line ("{run}" any
// name), or the stop-run answer when nothing is running.
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

// gopodSystemStingNow: the sting before this system line - during a warm-up, only its first line's
// (the default sting), else gopodSystemSting.
func gopodSystemStingNow(text string) (string, float64) {
	if gopodWarmingStarted() {
		if gopodWarmStingStart.CompareAndSwap(true, false) {
			return gopodSystemSting("")
		}
		return "", 0
	}
	return gopodSystemSting(text)
}

// gopodNoSting: the system line being said follows one that already had its sting (one sting per
// bubble: "Warm-up complete.
var gopodNoSting atomic.Bool

// gopodSystemSayNoSting says a system line with no sting of its own.
func gopodSystemSayNoSting(text string, wait bool) {
	gopodNoSting.Store(true)
	defer gopodNoSting.Store(false)
	gopodSystemSay(text, wait)
}

// The voice files this Wire-Pod start uses (gopodReporterPaths), kept when the folder is cleaned
// (gopodReporterPrune).
var gopodReporterInUse = struct {
	sync.Mutex
	files map[string]bool
}{files: map[string]bool{}}

// gopodReporterPrune: once every line is made at Wire-Pod's start, ...
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
	speed := gopodReporterSpeedFor(host)
	key, tag := voice+"\n"+text, voice
	if speed != 1.03 {
		key += fmt.Sprintf("\n%.2f", speed)
		tag += fmt.Sprintf("_s%03.0f", speed*100)
	}
	if pitch := gopodReporterPitch(host); pitch != 0 {
		key += fmt.Sprintf("\np%.0f", pitch)
		tag += fmt.Sprintf("_p%+.0f", pitch)
	}
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
	if ok {
		if b, err := os.ReadFile(strings.TrimSuffix(file, ".wav") + ".words.json"); err != nil || !strings.Contains(string(b), `"aligner"`) {
			ok = false
		} else if !strings.Contains(string(b), `"levelled": true`) {
			ok = false
		} else if !strings.Contains(string(b), `"aligner": "vosk"`) && gopodFileExists(gopodHomePath("~/wire-pod/vosk/models/en-US/model")) {
			ok = false
		}
	}
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

// gopodDryVoice: set only by the session dry run (gopod_session_dry_test.go) - a voice line there
// is looked up in the library and logged, never played.
var gopodDryVoice func(host, text string, show bool)

// gopodReporterSay plays host's line through the page; with wait, it returns once the line has been
// heard (its length, plus the page's poll lag, up to max_wait_seconds).
func gopodReporterSay(host, text string, show, wait bool) bool {
	return gopodReporterSayTo(host, "", text, show, wait)
}

// gopodSayForShownLine: a voice whose words are already on the page (a warm-up/opener line, an egg,
// the Operator's own line) - its line is named with its voice file, so the page ties them:
// "forwait:<page>|text" (Generating...
func gopodSayForShownLine(host, text string) (string, float64, bool) {
	rv := gopodBrobotsSettings.ReporterVoices
	says := gopodSaysFor(host, text)
	said := says
	cachedPage, _, made := gopodReporterCached(host, said)
	shown := strings.TrimSpace(text)
	if rv.GeneratingMarker != "" && shown != "" {
		// Every page voice shows Generating + the hourglass until its file is here - a voice
		// already made too (it arrives at once, so the hourglass only flashes).
		page := cachedPage
		if !made {
			_, page = gopodReporterPaths(host, said)
		}
		GopodChatFeedAdd(rv.Role, rv.GeneratingMarker+"forwait:"+page+"|"+shown)
	}
	page, seconds, ok := gopodReporterFile(host, says)
	if !ok && !made && rv.GeneratingMarker != "" {
		GopodChatFeedAdd(rv.Role, rv.GeneratingMarker+"off")
	}
	return page, seconds, ok
}

// gopodReporterSayTo: gopodReporterSay with its bubble's "to" - a manager's line for a Brobot shows
// inside that Brobot's bubble.
func gopodReporterSayTo(host, to, text string, show, wait bool) bool {
	if show && to == "" {
		if chunks := gopodSpokenChunks(text); len(chunks) > 1 {
			gopodSayLinesInBubble(host, chunks, wait, gopodChunksDisplay(text, chunks))
			return true
		}
	}
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
		logger.Println("GOPOD_HEARD " + host + ": no voice file - not said: " + gopodHeardShort(text))
		return false
	}
	return gopodVoicePlay(host, text, page, seconds, wait)
}

// gopodVoicePlay: a voice file to the page, a system line's sting around it, in the voice queue,
// and - wait - on until the page has played it.
func gopodVoicePlay(host, text, page string, seconds float64, wait bool) bool {
	rv := gopodBrobotsSettings.ReporterVoices
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
	ends := gopodVoiceQueueAdd(seconds + rv.GapSeconds)
	if !wait {
		return true
	}
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

// gopodMusicPlayed: a line's own end (music) has played the sting, so the warm-up's end sting
// doesn't play it again (gopodWarmupEndSting).
var gopodMusicPlayed atomic.Bool

// gopodHeardShort: a line's first words, for the journal.
func gopodHeardShort(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	if len(text) > 60 {
		text = gopodCut(text, 60)
	}
	return text
}

// gopodHeardLimit: how long a heard check waits with no answer - the lines queued until ends (their
// own lengths), plus reporter_voices.heard_margin_seconds; never more than 2 x max_wait_seconds.
func gopodHeardLimit(ends time.Time) time.Duration {
	rv := gopodBrobotsSettings.ReporterVoices
	limit := time.Until(ends) + gopodSecondsToDuration(rv.HeardMarginSeconds)
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

// gopodVoiceQueue: when the page's voice queue runs out - each line starts once the lines before it
// have played.
var gopodVoiceQueue = struct {
	sync.Mutex
	ends time.Time
}{}

// gopodVoiceQueueAdd adds a line this long to the queue; it returns when that line will have
// played.
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

// gopodReporterReadsOperator reads an operator line out, when how (e.g.
const gopodOperatorVoiceHost = "Operator"

func gopodReporterReadsOperator(text, how, to string) {
	rv := gopodBrobotsSettings.ReporterVoices
	host := rv.ReadBy[to]
	if gopodScriptPlaying.Load() {
		if h := gopodScriptTakeReader(); h != "" {
			host = h
		}
	}
	if host == "" {
		host = rv.OperatorHost
	}
	if !gopodScriptPlaying.Load() && gopodReporterVoice(gopodOperatorVoiceHost) != "" {
		host = gopodOperatorVoiceHost
	}
	asker := gopodCommCheckAsker(text)
	if asker != "" {
		host = asker
		if gopodReporterVoice(host) == "" {
			host = gopodBrobotsSettings.SignOff.SystemRole
		}
	}
	for _, from := range rv.ReadLinesFrom {
		if from == how {
			if asker == "" {
				if gopodMicVoiceSay(host, text) {
					return
				}
			}
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

func gopodReporterOpening(warm bool) {
	rv := gopodBrobotsSettings.ReporterVoices
	opening := gopodBookends(true)
	if len(opening) == 0 {
		return
	}
	select {
	case <-gopodReporterBookendsMade:
	default:
		if t := strings.TrimSpace(gopodBrobotsSettings.LLM.VoicesEngineText); t != "" && warm {
			logger.Println("GOPOD_REPORTER_VOICES waiting for the voice engine - shown: " + t)
			gopodWarmupMessage(t)
		}
	}
	select {
	case <-gopodReporterBookendsMade:
	case <-time.After(gopodSecondsToDuration(rv.StartWaitSeconds)):
		logger.Println("GOPOD_REPORTER_VOICES opening skipped: voices not ready")
		return
	}
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

// GopodOpeningHeard: the chat page has played the opening (POST /api-sdk/gopod_opening_heard).
func GopodOpeningHeard() {
	gopodOpeningHeardOnce.Do(func() { close(gopodOpeningHeard) })
}

// gopodReporterClosing plays the closing after "Thanks Brobots" - once, even when both Brobots
// answer the thanks, after both have finished their sign-off.
func gopodReporterClosing() {
	gopodReporter.Lock()
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
	for _, l := range gopodBookends(false) {
		gopodReporterSay(l.Host, l.Text, true, true)
	}
}

// gopodReporterNewQuestion: a question that isn't a thanks - the next ending gets its own closing.
func gopodReporterNewQuestion() {
	gopodReporter.Lock()
	gopodReporter.closing = nil
	gopodReporter.signedOff = false
	gopodReporter.ended = false
	gopodReporter.Unlock()
}

// gopodDraftRunCompleted: after the ending of a session that played a draft, sign_off.system_text
// in a Goverlord system bubble - once.
func gopodDraftRunCompleted() {
	so := gopodBrobotsSettings.SignOff
	if so.SystemRole == "" || so.SystemText == "" || !gopodScriptRan.CompareAndSwap(true, false) {
		return
	}
	gopodReporterSay(so.SystemRole, so.SystemText, true, true)
}

// gopodSystemSay says a system line in the system voice (reporter_voices.
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
	lines = append(lines, gopodOpenerSaid()...)
	for _, ch := range s.InterviewChapters.Chapters {
		lines = append(lines, gopodChapterSaid(ch.Title))
	}
	rs := s.RecordSession
	lines = append(lines, rs.KeptText, rs.DeletedText, rs.KeptText+" "+s.SignOff.WelcomeText, rs.DeletedText+" "+s.SignOff.WelcomeText, strings.TrimSpace(s.SignOff.NothingRecordedText+" "+s.SignOff.WelcomeText))
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
		if _, _, ok := gopodReporterCached(l.Host, gopodSaysFor(l.Host, l.Text)); !ok {
			return false
		}
	}
	return len(lines) > 0
}

// gopodReporterSays: the line as the reporter says it - reporter_voices.say_words replaced, whole
// words, any case.
func gopodSaysFor(host, text string) string {
	words := gopodLingoWords()
	if gopodSpeaksAsSystem(host) {
		for k, v := range gopodBrobotsSettings.ReporterVoices.SystemSayWords {
			for old := range words {
				if strings.EqualFold(old, k) {
					delete(words, old)
				}
			}
			words[k] = v
		}
	}
	return gopodSaySwap(gopodSpeakMilitaryTime(gopodSpeakNoEmojis(gopodSpeakFlags(text))), words)
}

// gopodSpeakNoEmojis: emojis are shown, not said - the voice skips them (a flag is said as its
// country first, gopodSpeakFlags).
func gopodSpeakNoEmojis(text string) string {
	return strings.Join(strings.Fields(gopodPersonaEmojiRE.ReplaceAllString(text, "")), " ")
}

// gopodReporterSays: gopodSaysFor a voice that isn't the system's.
func gopodReporterSays(text string) string {
	return gopodSaySwap(gopodSpeakMilitaryTime(gopodSpeakNoEmojis(gopodSpeakFlags(text))), gopodLingoWords())
}

// gopodSaySwap: each whole word in words (any case) said as its value.
func gopodSaySwap(text string, words map[string]string) string {
	keys := make([]string, 0, len(words))
	for from := range words {
		keys = append(keys, from)
	}
	sort.Slice(keys, func(i, j int) bool {
		if len(keys[i]) != len(keys[j]) {
			return len(keys[i]) > len(keys[j])
		}
		return keys[i] < keys[j]
	})
	for _, from := range keys {
		to := words[from]
		if strings.TrimSpace(from) == "" {
			continue
		}
		re := regexp.MustCompile(`(?i)(^|[^\pL\pN_])` + regexp.QuoteMeta(from) + `($|[^\pL\pN_])`)
		for i := 0; i < 2; i++ {
			text = re.ReplaceAllString(text, "${1}"+strings.ReplaceAll(to, "$", "$$")+"${2}")
		}
	}
	return text
}

// gopodReporterAllLines: every scripted line a voice may say, with the one voice that says it -
// reporter_voices' opening and closing, the interview's start lines, the interview's banter,
// handoffs and questions, and the run commands (Operator voice).
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
	// The interview that "start interview" plays: its song folder, else the newest draft.
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
	if rv.OperatorVoice != "" {
		for _, text := range gopodRunCommandLines() {
			add(gopodOperatorVoiceHost, text)
		}
	}
	return out
}

// gopodRunCommandLines: the start and stop commands - each run command's first phrase
// (customIntents.json) and song_stop.stop_words.
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

// gopodSignOffLines: after "thanks Brobots", once the sign-off has been said, the Brobots'
// sign_off.lines, each in turn - once per ending.
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
			gopodBodySay(l.Who, l.Text)
		}
	}
}

// gopodSpokenChunks: a line as spoken chunks: one sentence a chunk; a sentence under
// reporter_voices.chunk_min_words joins the next.
func gopodSpokenChunks(text string) []string {
	rv := gopodBrobotsSettings.ReporterVoices
	if rv.ChunkWords <= 0 || len(strings.Fields(text)) <= rv.ChunkWords {
		return []string{text}
	}
	var out []string
	cur := ""
	for _, s := range gopodWholeSentences(strings.TrimSpace(text)) {
		if cur != "" {
			cur += " "
		}
		cur += strings.TrimSpace(s)
		if len(strings.Fields(cur)) >= rv.ChunkMinWords {
			out = append(out, cur)
			cur = ""
		}
	}
	if cur != "" {
		if len(out) > 0 {
			out[len(out)-1] += " " + cur
		} else {
			out = append(out, cur)
		}
	}
	if len(out) == 0 {
		return []string{text}
	}
	return out
}

// ## EOF

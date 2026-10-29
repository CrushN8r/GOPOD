// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_llm_warmup.go
package wirepod_ttr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// The Brobots' LLM stays warm (operator 2026-10-03: "I prefer a pre-warm up and
// then have it stay loaded, so it doesn't go cold"). After a Wire-Pod restart the
// first chat question waited 2 min 37 s, about 2 min of it Ollama loading the model.
// At start, one silent request to the same model and endpoint the chat uses loads
// it (no num_ctx or other load options are sent, same as the chat, so the chat
// reuses that load; a question that comes in meanwhile waits on the same load).
// Then Ollama is told every llm.refresh_seconds to keep the model loaded for
// llm.keep_alive, since each chat request sets Ollama's own 5-minute default
// again. Nothing is spoken, shown as a chat line, or kept in the chat memory: the
// page shows the warm-up from its own small file (gopod_llm_warmup.json).

var gopodLLMWarmupStatePath = gopodPathWithEnvOverride("GOPOD_LLM_WARMUP_STATE_PATH", "webroot/gopod_llm_warmup.json")

// The warm-up's waits are timings in the settings file (gopod_timings.go):
// llm_config_wait_seconds, llm_retry_wait_seconds, llm_warm_request_seconds.
const gopodLLMWarmupRetries = 24

// GopodLLMWarmupStart runs the warm-up and then the keep-loaded refresh in the
// background. Called once at startup (sdkapp BeginServer).
func GopodLLMWarmupStart() {
	gopodWarmGateClosed.Store(true)
	// 2026-10-17: the chat page here or away (gopod_page_presence.go).
	gopodPageWatchStart()
	// 2026-10-17: every sound evened to chat_audio.normalize_db (gopod_audio_levels.go).
	gopodAudioLevelsStart()
	// 2026-10-18 (step 4): every file the data names, looked for (GOPOD_CHECK).
	go gopodSystemCheckAtStart()
	// 2026-10-17: the bubbles' eye colours from Wire-Pod's own robot settings.
	gopodEyeColorsStart()
	// GOLDEN, locked (2026-10-13, operator: "Always Use System llm bubble 2
	// with animation + system messages ... \"System llm bubble\" displays first
	// and always stays at the bottom of what ever happens until LLM Ready";
	// "Sequential Run Check"; "PLEASE FIX GOLDEN AND F*CK*NG LOCK THE PROCESS
	// IN"): the one System llm bubble from the start, then its messages in the
	// chat's own order (gopodWarmupStartSequence). Locked by
	// TestWarmupSequenceGolden.
	gopodWriteLLMWarmupState(true)
	go func() {
		base, model, ok := gopodWarmupOllama()
		if !ok {
			logger.Println("GOPOD_LLM_WARMUP failed no Ollama endpoint/model in the knowledge settings")
			gopodWarmGateClosed.Store(false)
			return
		}
		logger.Println(fmt.Sprintf("GOPOD_LLM_WARMUP model already loaded=%v", gopodWarmupLoaded(base, model)))
		warmed := make(chan struct{})
		go func() {
			defer close(warmed)
			gopodLLMWarmup(base, model)
		}()
		gopodWarmupStartSequence(warmed)
		gopodWarmGateClosed.Store(false)
		logger.Println("GOPOD_WARMUP_GATE open")
		every := gopodSecondsToDuration(gopodBrobotsSettings.LLM.RefreshSeconds)
		for {
			time.Sleep(every)
			if err := gopodLLMKeepLoaded(base, model); err != nil {
				logger.Println("GOPOD_LLM_KEEPALIVE failed " + err.Error())
			}
		}
	}()
}

// The warm-up's outside world, swapped in tests.
var (
	gopodWarmupOllama      = gopodWaitForOllamaConfig
	gopodWarmupLoaded      = gopodLLMLoaded
	gopodWarmupWarmRequest = gopodLLMWarmRequest
	gopodWarmupVoicesMade  = func() bool {
		opening := gopodBookends(true)
		return len(opening) == 0 || gopodReporterAllMade(opening)
	}
	gopodWarmupVoicesWait = func() {
		select {
		case <-gopodReporterBookendsMade:
		case <-time.After(gopodSecondsToDuration(gopodBrobotsSettings.ReporterVoices.StartWaitSeconds)):
		}
	}
	gopodWarmupOpening = func() { gopodReporterOpening(true) }
)

// gopodWarmupStartSequence: Wire-Pod start's sequential run check, in order -
//  1. "Voices are being generated. Please wait." while the voices are made
//     (skipped when they're made already);
//  2. "LLM Warming Up. Please wait.", always shown and said; a cold model is
//     waited out;
//  3. the lines in between (the reporters' opening), the message staying "LLM
//     Warming Up. Please wait." (operator 2026-10-13: "NOT \"LLM Warming Up.
//     Please Wait.\" then \"Please Wait.\" JUST \"LLM Warming Up. Please Wait.\"");
//  4. "LLM Ready. Warm-up complete. ...", shown and said.
//
// Each message is a line in the chat feed, so the page shows it in turn even
// when it's opened later (live 16:51: the page opened after the messages had
// passed and saw only the last one).
func gopodWarmupStartSequence(warmed <-chan struct{}) {
	// 2026-10-16 (operator: "make the opener LLM Warming song"): the opener
	// song's steps (gopod_opener_song.go), or the built-in opener.
	gopodPlayOpener(gopodOpenerSteps(), warmed)
	gopodWarmingEnd()
}

// gopodWarmupLine: the System llm bubble's message, then llm.wait_text
// ("" = wait_text alone).
func gopodWarmupLine(message string) string {
	text := strings.TrimSpace(message)
	// 2026-10-16: a message that already ends with it ("LLM warm up next. Please
	// wait.") doesn't get it twice.
	if wait := strings.TrimSpace(gopodBrobotsSettings.LLM.WaitText); wait != "" && !strings.HasSuffix(text, wait) {
		if text != "" && !strings.ContainsAny(text[len(text)-1:], ".!?") {
			text += "."
		}
		// 2026-10-16 (operator: "NEW system message ... lines"): each on its own line.
		text = strings.TrimSpace(text + "\n" + wait)
	}
	return text
}

// gopodWarmupMessage shows the System llm bubble (warming, LEDs filling) with
// its message, as a feed line ("warmup_text|<started_ms>|<text>") the page
// shows in the chat's order.
func gopodWarmupMessage(message string) {
	text := gopodWarmupLine(message)
	gopodWarmingBegin()
	gopodLLMWarmupText = text
	gopodWriteLLMWarmupState(true)
	GopodChatFeedAdd(gopodBrobotsSettings.ReporterVoices.Role, gopodWarmupTextMark+strconv.FormatInt(gopodLLMWarmupStartedMS, 10)+"|"+text)
}

// gopodWarmupTextMark: the feed line that sets the System llm bubble's message.
const gopodWarmupTextMark = "warmup_text|"

// GopodSongWarmupText: a song's warm-up (run_golden_song_001.py) sets its
// bubble's (started_ms) message, in the chat's order.
func GopodSongWarmupText(startedMS int64, text string) {
	gopodWarmingBegin()
	GopodChatFeedAdd(gopodBrobotsSettings.ReporterVoices.Role, gopodWarmupTextMark+strconv.FormatInt(startedMS, 10)+"|"+strings.TrimSpace(text))
}

// gopodLLMWarmupText: the System llm bubble's message while it warms.
var gopodLLMWarmupText string

// gopodRunWarmup (2026-10-12, operator: "Upon wpr & all start songs * look at
// the opener. that's your golden templated example * all songs must have this
// same golden templated process * reporter comments may or may not be present
// within warming up & ready"): a run's start like Wire-Pod's, GOLDEN and locked
// (TestWarmupSequenceGolden): the System llm bubble with "LLM Warming Up.
// Please wait." (shown and said; a cold model waited out), then "Please wait."
// while the lines play (may be none), then ready with readyText, said.
func gopodRunWarmup(readyText string, lines func()) {
	base, model, ok := gopodWarmupOllama()
	cold := !ok || !gopodWarmupLoaded(base, model)
	gopodLLMWarmupStartedMS = 0
	gopodLLMWarmupActive = false
	gopodWarmupMessage(gopodBrobotsSettings.LLM.WarmupText)
	// 2026-10-14 (operator: "For LLM warming up message, I did not hear the
	// \"Please wait.\" at the end * generate the full message displayed"): said as
	// shown, with llm.wait_text.
	gopodSystemSay(gopodWarmupLine(gopodBrobotsSettings.LLM.WarmupText), false)
	if cold && ok {
		warm := make(chan struct{})
		go func() {
			defer close(warm)
			if err := gopodWarmupWarmRequest(base, model); err != nil {
				logger.Println("GOPOD_LLM_WARMUP run warm-up failed " + err.Error())
			}
		}()
		select {
		case <-warm:
		case <-time.After(gopodTiming("llm_warm_request_seconds")):
		}
	}
	if lines != nil {
		lines()
	}
	gopodLLMWarmupReady(strings.TrimSpace(readyText), true)
	gopodSystemSay(readyText, true)
	gopodWarmupEndSting()
	gopodWarmingEnd()
}

// gopodLLMWarmupReady turns the warm-up bubble ready (warmedUp: the bubble
// that showed the warming), or, after a warm start, starts a ready bubble.
// The page shows it in the chat's order: at the warmup_ready line in the feed,
// after everything before it.
func gopodLLMWarmupReady(readyText string, warmedUp bool) {
	if !warmedUp {
		gopodLLMWarmupStartedMS = time.Now().UnixMilli()
	}
	gopodLLMWarmupReadyText = readyText
	gopodWriteLLMWarmupState(false)
	GopodChatFeedAdd(gopodBrobotsSettings.ReporterVoices.Role, gopodWarmupReadyMark+strconv.FormatInt(gopodLLMWarmupStartedMS, 10)+"|"+readyText)
}

// gopodSystemSayQueued: system_say_role lines being queued in the background.
var gopodSystemSayQueued sync.WaitGroup

// gopodRunReadyLine (2026-10-14, operator: "part 1 standard * keep it in
// settings file once * each song gives only part 2"): a run's LLM Ready
// message - llm.run_ready_text, then the run's own part 2.
//
// 2026-10-16 (operator: "for song runs, apply this template as a base * for each
// song, apply/add updates/changes accordingly"): the opener's message 3 shape -
// run_ready_text, the song's own line, run_ready_end - one line each.
func gopodRunReadyLine(part2 string) string {
	llm := gopodBrobotsSettings.LLM
	var lines []string
	for _, l := range []string{llm.RunReadyText, part2, llm.RunReadyEnd} {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, l)
		}
	}
	return strings.Join(lines, "\n")
}

// gopodWarmupReadyMark: the feed line that shows a warm-up bubble ready
// ("warmup_ready|<started_ms>|<ready text>", gopod_brobots_chat.js).
const gopodWarmupReadyMark = "warmup_ready|"

// GopodSongWarmupReady: a song's warm-up (run_golden_song_001.py) is ready -
// its bubble (started_ms) shows ready in the chat's order, then Goverlord says
// readyText; returns once said.
func GopodSongWarmupReady(startedMS int64, readyText string) bool {
	// 2026-10-14 (operator, the Fart Joke run: "audio didn't sync with bubbles
	// display, since no reporter lines in between * \"LLM Warming Up. Please
	// Wait.\" stays until fully spoken * if LLM Ready, then update to LLM Ready
	// message + speech * rule applies to all llm warming runs"): the song's
	// warming line is said in the background (system_say_role); the ready line
	// goes in the chat after it, so the page says it in full before turning
	// the bubble ready. Locked by TestSongWarmupReadyAfterWarming.
	gopodSystemSayQueued.Wait()
	GopodChatFeedAdd(gopodBrobotsSettings.ReporterVoices.Role, gopodWarmupReadyMark+strconv.FormatInt(startedMS, 10)+"|"+strings.TrimSpace(readyText))
	heard := gopodSystemSay(readyText, true)
	gopodWarmupEndSting()
	gopodWarmingEnd()
	return heard
}

// The LLM warm-up pause (2026-10-13, operator: "LLM warming = disabled/pause on
// all input * LLM must reach 100% LLM warmed status, then continue * so the
// stop run and response, should have waited in queue for the LLM Ready message
// to complete"): from a run's first warm-up message (Wire-Pod's start, the
// interview, a replay, a song) until its LLM Ready has been said, a typed or
// numpad line - "stop run" too - waits in a queue, then plays in the order it
// came. A warm-up left hanging (a song that died) lets go after
// gopodWarmingMax. Locked by TestWarmupQueuesInput.
var gopodWarming = struct {
	sync.Mutex
	since  time.Time
	queue  []gopodQueuedLine
	drains bool
}{}

// gopodWarmingMax: timings.llm_warming_max_seconds (gopod_timings.go).
func gopodWarmingMax() time.Duration { return gopodTiming("llm_warming_max_seconds") }

type gopodQueuedLine struct{ esn, text, how string }

// gopodMicHow: a line the robot's own microphone heard, held during a warm-up.
const gopodMicHow = "Mic text"

// gopodWarmupReplayLine plays a queued line (swapped in tests).
var gopodWarmupReplayLine func(esn, text, how string) error

func init() { gopodWarmupReplayLine = gopodOperatorLine }

func gopodWarmingBegin() {
	gopodWarming.Lock()
	if gopodWarming.since.IsZero() {
		gopodWarming.since = time.Now()
		gopodWarmStingStart.Store(true)
		gopodMusicPlayed.Store(false)
	}
	gopodWarming.Unlock()
}

// The warm-up's stings (2026-10-14, operator: "For all llm warm up runs * Rule:
// djent plays once at the very start and very end * [djent] [message] = audio
// first, pre-recorded ... * [message] [djent] = audio last, after LLM Ready
// message"): while a warm-up is going, its first spoken system line gets the
// default sting before it and every other line none (gopodSystemStingNow); once
// its LLM Ready message has been said, the sting plays once more
// (gopodWarmupEndSting). Locked by TestWarmupStings.
var gopodWarmStingStart atomic.Bool

// gopodWarmingStarted: a warm-up's first message has been shown and its ready
// not yet said.
func gopodWarmingStarted() bool {
	gopodWarming.Lock()
	defer gopodWarming.Unlock()
	return !gopodWarming.since.IsZero()
}

// gopodWarmupEndSting queues the default sting after everything queued before
// it (the LLM Ready message), then waits until the page has played them all
// (2026-10-14, operator: "Brobot run started before \"LLM Ready. Warm-up
// complete. We here, are Fart Joke Ready\" finished playing"): the run goes on
// only once they're heard, not after a guess of their length.
func gopodWarmupEndSting() {
	rv := gopodBrobotsSettings.ReporterVoices
	ends := gopodVoiceQueueAdd(0)
	// 2026-10-17: the ready line's own "(music)" already played it.
	if gopodMusicPlayed.Swap(false) {
		gopodWaitHeard(gopodHeardLimit(ends))
		return
	}
	if sting, seconds := gopodMusicSting(); sting != "" {
		GopodChatFeedAdd(rv.Role, gopodWithVolumeShare(sting))
		ends = gopodVoiceQueueAdd(seconds + rv.GapSeconds)
	}
	// 2026-10-17 (survey E5): heard or not, the journal says so.
	if gopodWaitHeard(gopodHeardLimit(ends)) {
		logger.Println("GOPOD_HEARD sting: heard")
	} else {
		logger.Println("GOPOD_HEARD sting: not heard (going on)")
	}
}

func gopodWarmingEnd() {
	gopodWarming.Lock()
	gopodWarming.since = time.Time{}
	gopodWarming.Unlock()
}

// gopodWarmingNow: a warm-up is going (or Wire-Pod's start gate is closed).
func gopodWarmingNow() bool {
	gopodWarming.Lock()
	defer gopodWarming.Unlock()
	return gopodWarmGateClosed.Load() || !gopodWarming.since.IsZero() && time.Since(gopodWarming.since) < gopodWarmingMax()
}

// gopodWarmupHold queues the line while a warm-up is going; true when queued.
func gopodWarmupHold(esn, text, how string) bool {
	if !gopodWarmingNow() {
		return false
	}
	gopodWarming.Lock()
	gopodWarming.queue = append(gopodWarming.queue, gopodQueuedLine{esn, text, how})
	start := !gopodWarming.drains
	gopodWarming.drains = true
	gopodWarming.Unlock()
	logger.Println("GOPOD_WARMUP_QUEUED " + how + " for " + esn + " until LLM Ready: " + text)
	if start {
		go gopodWarmupDrain()
	}
	return true
}

func gopodWarmupDrain() {
	for {
		for gopodWarmingNow() {
			time.Sleep(200 * time.Millisecond)
		}
		gopodWarming.Lock()
		if len(gopodWarming.queue) == 0 {
			gopodWarming.drains = false
			gopodWarming.Unlock()
			return
		}
		next := gopodWarming.queue[0]
		gopodWarming.queue = gopodWarming.queue[1:]
		gopodWarming.Unlock()
		if err := gopodWarmupReplayLine(next.esn, next.text, next.how); err != nil {
			logger.Println("GOPOD_WARMUP_QUEUED line: " + err.Error())
		}
	}
}

// The warm-up gate (operator 2026-10-06: "while warmup timer for 'brobots' and
// 'kokoro', brobots ... any activity attempt routes to null ... once timer gives
// the green light, routing to null stops"): from Wire-Pod start until LLM
// Ready - the LLM warm and the hosts' opening heard on the page - a typed,
// numpad or voice question to a Brobot is dropped (logged GOPOD_WARMUP_GATE).
// Native responses the robot makes on its own stay as they are.
var gopodWarmGateClosed atomic.Bool

// gopodWarmGateDrop reports (and logs) a question dropped while the gate is closed.
func gopodWarmGateDrop(esn, how, text string) bool {
	if !gopodWarmGateClosed.Load() || gopodBrobotName(esn) == "Brobot" && esn != gopodTypedAll {
		return false
	}
	logger.Println("GOPOD_WARMUP_GATE dropped " + how + " for " + esn + " (not ready yet): " + text)
	return true
}

// gopodWaitForOllamaConfig waits for Wire-Pod's knowledge settings (loaded at
// start) and returns Ollama's base address (the endpoint without "/v1") and the
// model. Only the "custom" provider (the local Ollama) is warmed.
func gopodWaitForOllamaConfig() (string, string, bool) {
	deadline := time.Now().Add(gopodTiming("llm_config_wait_seconds"))
	for {
		k := vars.APIConfig.Knowledge
		if k.Enable && k.Provider == "custom" && strings.TrimSpace(k.Endpoint) != "" && strings.TrimSpace(k.Model) != "" {
			return gopodOllamaBase(k.Endpoint), k.Model, true
		}
		if time.Now().After(deadline) {
			return "", "", false
		}
		time.Sleep(time.Second)
	}
}

// gopodOllamaBase turns the chat endpoint (".../v1") into Ollama's own base.
func gopodOllamaBase(endpoint string) string {
	base := strings.TrimRight(strings.TrimSpace(endpoint), "/")
	return strings.TrimRight(strings.TrimSuffix(base, "/v1"), "/")
}

// gopodLLMWarmup sends the warm request, retrying while Ollama is not up yet
// (Wire-Pod and Ollama start separately), then sets keep_alive.
func gopodLLMWarmup(base, model string) {
	logger.Println("GOPOD_LLM_WARMUP start model=" + model)
	start := time.Now()
	var err error
	for try := 0; try < gopodLLMWarmupRetries; try++ {
		if err = gopodLLMWarmRequest(base, model); err == nil || !gopodIsNotUpYet(err) {
			break
		}
		time.Sleep(gopodTiming("llm_retry_wait_seconds"))
	}
	if err == nil {
		err = gopodLLMKeepLoaded(base, model)
	}
	if err != nil {
		logger.Println("GOPOD_LLM_WARMUP failed " + err.Error())
		return
	}
	logger.Println(fmt.Sprintf("GOPOD_LLM_WARMUP done took=%d", time.Since(start).Milliseconds()))
}

// gopodLLMWarmRequest is one tiny chat request on the chat's own endpoint and
// model: it loads the model the way the chat will use it.
func gopodLLMWarmRequest(base, model string) error {
	return gopodOllamaPost(base+"/v1/chat/completions", map[string]any{
		"model":      model,
		"messages":   []map[string]string{{"role": "user", "content": "Say ready."}},
		"stream":     false,
		"max_tokens": 1,
	})
}

// gopodLLMLoaded: Ollama already holds the model (its /api/ps list); unknown is false.
func gopodLLMLoaded(base, model string) bool {
	client := &http.Client{Timeout: gopodTiming("llm_check_seconds")}
	resp, err := client.Get(base + "/api/ps")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	var ps struct {
		Models []struct {
			Name  string `json:"name"`
			Model string `json:"model"`
		} `json:"models"`
	}
	if resp.StatusCode != http.StatusOK || json.NewDecoder(resp.Body).Decode(&ps) != nil {
		return false
	}
	for _, m := range ps.Models {
		if m.Name == model || m.Model == model {
			return true
		}
	}
	return false
}

// gopodLLMKeepLoaded tells Ollama to keep the model loaded for llm.keep_alive.
// No prompt: Ollama only loads the model (already loaded = nothing to do) and
// sets the time.
func gopodLLMKeepLoaded(base, model string) error {
	return gopodOllamaPost(base+"/api/generate", map[string]any{
		"model":      model,
		"keep_alive": strings.TrimSpace(gopodBrobotsSettings.LLM.KeepAlive),
	})
}

func gopodOllamaPost(url string, body map[string]any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: gopodTiming("llm_warm_request_seconds")}
	resp, err := client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("HTTP " + resp.Status)
	}
	return nil
}

// gopodIsNotUpYet: Ollama isn't listening yet.
func gopodIsNotUpYet(err error) bool {
	return strings.Contains(err.Error(), "connection refused")
}

// gopodWriteLLMWarmupState writes the page's warm-up file. 2026-10-04
// (operator: "after restart until after fully warmed up llm, the timer stays;
// message changes, LLM Warming Up to LLM Ready"): started_ms when the warm-up
// starts, done_ms when the model is loaded - the page shows the LLM power level
// filling from started_ms, then the ready line (ready_text, or the start's
// "LLM Ready. Warm-up complete.") once done_ms is set; the bubble stays. The
// texts are in the settings file, which the page reads too.
var gopodLLMWarmupStartedMS int64

// gopodLLMWarmupReadyText, when set, is the bubble's ready line (ready_text in
// the file); the page falls back to llm.ready_text.
var gopodLLMWarmupReadyText string

// gopodLLMWarmupActive: the last state written was warming.
var gopodLLMWarmupActive bool

// gopodWarmupReadyLine joins LLM Ready and Warm-up complete into one line.
func gopodWarmupReadyLine(ready, done string) string {
	ready, done = strings.TrimSpace(ready), strings.TrimSpace(done)
	if done == "" || ready == "" {
		return ready + done
	}
	if !strings.ContainsAny(ready[len(ready)-1:], ".!?") {
		ready += "."
	}
	// 2026-10-16 (operator's message 3: "LLM Ready, Master CrushN8r." / "Standing
	// by for another crushing performance." / "Your move."): a line each.
	return ready + "\n" + done
}

func gopodWriteLLMWarmupState(active bool) {
	now := time.Now().UnixMilli()
	state := struct {
		Active    bool   `json:"active"`
		StartedMS int64  `json:"started_ms"`
		DoneMS    int64  `json:"done_ms"`
		ReadyText string `json:"ready_text,omitempty"`
		Text      string `json:"text,omitempty"` // the message while warming
		// 2026-10-17: the opener awaits the chat page (Text is awaits_text).
		Awaits bool `json:"awaits,omitempty"`
	}{Active: active, Awaits: active && gopodLLMWarmupAwaits}
	if active {
		// One bubble per warm-up: a new message keeps its start.
		if gopodLLMWarmupStartedMS == 0 || !gopodLLMWarmupActive {
			gopodLLMWarmupStartedMS = now
		}
		state.Text = gopodLLMWarmupText
	} else {
		state.DoneMS = now
		state.ReadyText = gopodLLMWarmupReadyText
	}
	state.StartedMS = gopodLLMWarmupStartedMS
	gopodLLMWarmupActive = active
	data, err := json.Marshal(state)
	if err != nil {
		return
	}
	tmp := gopodLLMWarmupStatePath + ".tmp"
	if err := gopodWriteFile(tmp, data, 0644); err != nil {
		logger.Println("GOPOD_LLM_WARMUP unable to write " + tmp + ": " + err.Error())
		return
	}
	if err := os.Rename(tmp, gopodLLMWarmupStatePath); err != nil {
		logger.Println("GOPOD_LLM_WARMUP unable to write " + gopodLLMWarmupStatePath + ": " + err.Error())
	}
}

// ## EOF

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_record_session.go
package wirepod_ttr

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// Record Session (operator 2026-10-04, spec gopod_notes/RECORD_SESSION_SPEC_003.md):
// a box on the chat window's typed-input line.
//   - Ticked: recording starts at the next safe point (no Brobot reply in
//     progress) into record_session.sessions_dir/<date>/<time><folder_suffix>/:
//     mic_48k.wav (the Jetson USB room mic, through the golden PTT writer's own mic
//     pick and 48 kHz path - the mic service, gopod_mic_service.go), chat.jsonl (every chat
//     line: the page's feed and every chat question and answer), wirepod.log
//     (every Wire-Pod log line), and chat.html at stop.
//   - Unticked: "you stopped it" - no question: zip, keep the zip, delete the
//     loose files.
//   - "Thanks Brobots" while recording, or max_minutes: stop, then ask_brobot asks
//     keep or delete. A typed or spoken line with every keep_words word keeps
//     (zip), with every delete_words word deletes, no answer within
//     answer_wait_minutes deletes. The answer shows as "Operator to <answer_to>";
//     confirm_brobot says kept_text / deleted_text; the box unticks.
//   - A Wire-Pod restart deletes a folder still waiting for an answer, and zips
//     one cut off mid-recording (never loses a recording nobody chose to delete).
// The page reads the box's state from webroot/gopod_record_session.json.

var gopodRecordStatePath = gopodPathWithEnvOverride("GOPOD_RECORD_SESSION_STATE_PATH", "webroot/gopod_record_session.json")

const (
	gopodRecordMarkerRecording = ".recording"
	gopodRecordMarkerPending   = ".pending"
	gopodRecordMicStartWait    = 15 * time.Second
	gopodRecordMicStopWait     = 3 * time.Minute // a resample at the end (sox) included
	gopodRecordReplyStale      = 5 * time.Minute
	// A typed question's answer comes this long after it at most.
	gopodRecordTypedSameWindow = 5 * time.Minute
)

var gopodRecord = struct {
	sync.Mutex
	wanted      bool // ticked, waiting for a safe point
	active      bool
	waiting     bool // asked keep or delete
	dir         string
	started     time.Time
	chat        *os.File
	log         *os.File
	micOn       bool
	maxTimer    *time.Timer
	answerTimer *time.Timer
	answeredAt  time.Time
	typedWords  []string // the last operator line from the feed (a typed line)
	typedAt     time.Time
	typedHow    string             // how the last operator line came in: "typed" or "numpad"
	decided     chan struct{}      // closed once the keep-or-delete answer is said
	takes       []gopodHarvestTake // the session's Brobot answers, for the take sheet on keep
	keptTakes   []gopodHarvestTake // the stopped session's, until kept or deleted
}{}

// gopodRecordWaitDecided waits until an open keep-or-delete question has been
// answered and the answer said (or decided with no answer), so what comes after
// it in the ending - "Draft run completed." and "You're welcome." - never talks
// over it (operator 2026-10-06: "So that's us. Mic's down. started without
// letting me answer").
func gopodRecordWaitDecided() {
	gopodRecord.Lock()
	decided := gopodRecord.decided
	gopodRecord.Unlock()
	if decided != nil {
		<-decided
	}
}

// GopodRecordSessionStart: once at Wire-Pod start (sdkapp BeginServer).
func GopodRecordSessionStart() {
	gopodRecordCleanupLeftovers()
	// The session's own account, always on (gopod_session_log.go).
	GopodSessionLogStart()
	gopodWriteRecordState()
	go gopodRecordLogTee()
}

// GopodRecordSessionSet is the box: on or off (sdkapp /api-sdk/gopod_record_session).
func GopodRecordSessionSet(on bool) {
	// 2026-10-12 (operator: "sessions cannot be recorded again"): never during a
	// session replay (gopod_session_replay.go).
	if on && gopodReplayPlaying.Load() {
		logger.Println("GOPOD_RECORD_SESSION not started: a session replay is playing")
		gopodWriteRecordState()
		return
	}
	gopodRecord.Lock()
	switch {
	case on && !gopodRecord.wanted && !gopodRecord.active && !gopodRecord.waiting:
		gopodRecord.wanted = true
		gopodRecord.Unlock()
		gopodWriteRecordState()
		go gopodRecordStartAtSafePoint()
		return
	case !on && gopodRecord.wanted && !gopodRecord.active:
		gopodRecord.wanted = false
		gopodRecord.Unlock()
		gopodWriteRecordState()
		return
	case !on && gopodRecord.waiting:
		// Unticked while the question is open: the operator chose - keep.
		gopodRecord.Unlock()
		gopodRecordDecide(true, "unticked")
		return
	case !on && gopodRecord.active:
		gopodRecord.Unlock()
		go func() {
			dir := gopodRecordStop("unticked")
			if dir != "" {
				gopodRecordKeep(dir)
			}
			gopodWriteRecordState()
		}()
		return
	}
	gopodRecord.Unlock()
}

// GopodRecordSessionThanks: "Thanks Brobots" was answered (the sign-off spoken).
func GopodRecordSessionThanks() {
	gopodRecordEndAndAsk("thanks")
}

// --- safe point: no Brobot chat reply in progress ---

var gopodReplying = struct {
	sync.Mutex
	at   map[string]time.Time
	done map[string]chan struct{}
}{at: map[string]time.Time{}, done: map[string]chan struct{}{}}

func gopodReplyStart(esn string) {
	gopodReplying.Lock()
	gopodReplying.at[esn] = time.Now()
	gopodReplying.Unlock()
}

func gopodReplyEnd(esn string) {
	gopodReplying.Lock()
	delete(gopodReplying.at, esn)
	done := gopodReplying.done[esn]
	delete(gopodReplying.done, esn)
	gopodReplying.Unlock()
	if done != nil {
		close(done)
	}
}

// gopodReplyDone: closed when this Brobot's next answer ends (gopodReplyEnd).
func gopodReplyDone(esn string) <-chan struct{} {
	gopodReplying.Lock()
	defer gopodReplying.Unlock()
	if gopodReplying.done == nil {
		gopodReplying.done = map[string]chan struct{}{}
	}
	done := make(chan struct{})
	gopodReplying.done[esn] = done
	return done
}

func gopodReplyBusy() bool {
	gopodReplying.Lock()
	defer gopodReplying.Unlock()
	for _, at := range gopodReplying.at {
		if time.Since(at) < gopodRecordReplyStale {
			return true
		}
	}
	return false
}

func gopodRecordStartAtSafePoint() {
	for gopodReplyBusy() {
		time.Sleep(250 * time.Millisecond)
	}
	gopodRecord.Lock()
	if !gopodRecord.wanted {
		gopodRecord.Unlock()
		return
	}
	gopodRecord.wanted = false
	gopodRecord.Unlock()
	if err := gopodRecordBegin(); err != nil {
		logger.Println("GOPOD_RECORD_SESSION failed to start: " + err.Error())
	}
	gopodWriteRecordState()
}

// --- recording ---

func gopodRecordBegin() error {
	root, uid, gid := gopodRecordRoot()
	now := time.Now()
	day := filepath.Join(root, now.Format("2006-01-02"))
	dir := filepath.Join(day, now.Format("150405")+gopodBrobotsSettings.RecordSession.FolderSuffix)
	// 2026-10-11 (operator: "write ALL session files ... within that session's
	// folder"): inside the open session's folder.
	if session := gopodSessionLogDir(); session != "" {
		day = session
		dir = filepath.Join(session, now.Format("150405")+gopodBrobotsSettings.RecordSession.FolderSuffix)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	gopodRecordChown(uid, gid, root, day, dir)
	marker := filepath.Join(dir, gopodRecordMarkerRecording)
	if err := os.WriteFile(marker, nil, 0o644); err != nil {
		return err
	}
	chat, err := os.OpenFile(filepath.Join(dir, "chat.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	wlog, err := os.OpenFile(filepath.Join(dir, "wirepod.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		chat.Close()
		return err
	}
	gopodRecordChown(uid, gid, marker, chat.Name(), wlog.Name())
	micOn := gopodRecordStartMic(dir)

	gopodRecord.Lock()
	gopodRecord.active, gopodRecord.dir, gopodRecord.started = true, dir, now
	gopodRecord.chat, gopodRecord.log = chat, wlog
	gopodRecord.micOn = micOn
	max := time.Duration(gopodBrobotsSettings.RecordSession.MaxMinutes * float64(time.Minute))
	gopodRecord.maxTimer = time.AfterFunc(max, func() { gopodRecordEndAndAsk("max_minutes") })
	gopodRecord.Unlock()
	logger.Println("GOPOD_RECORD_SESSION start dir=" + dir)
	gopodRecordWrite(gopodRecordLine{Source: "record", Who: "Record Session", Text: "started"})
	return nil
}

// gopodRecordStartMic starts the room-mic master through the mic service
// (gopod_mic_service.go), the one process that owns the USB mic - so numpad
// push-to-talk works during a recording. No USB mic: the session goes on with
// chat and logs only.
func gopodRecordStartMic(dir string) bool {
	r, err := gopodMicCall(map[string]any{"cmd": "record_start", "out": filepath.Join(dir, "mic_48k.wav")}, gopodRecordMicStartWait)
	if err == nil {
		if e := gopodMicReplyError(r); e != "" {
			err = errors.New(e)
		}
	}
	if err != nil {
		logger.Println("GOPOD_RECORD_SESSION room mic not recording: " + err.Error())
		return false
	}
	return true
}

// gopodRecordStop ends the recording and returns its folder ("" if none).
func gopodRecordStop(reason string) string {
	gopodRecord.Lock()
	if !gopodRecord.active {
		gopodRecord.Unlock()
		return ""
	}
	took := time.Since(gopodRecord.started)
	stopped := fmt.Sprintf("stopped (%s, %.1f min)", reason, took.Minutes())
	gopodRecordWriteLocked(gopodRecordLine{Source: "record", Who: "Record Session", Text: stopped})
	gopodRecord.log.WriteString(time.Now().Format("2006.01.02 15:04:05") + ": GOPOD_RECORD_SESSION " + stopped + "\n")
	dir, chat, wlog := gopodRecord.dir, gopodRecord.chat, gopodRecord.log
	micOn := gopodRecord.micOn
	if gopodRecord.maxTimer != nil {
		gopodRecord.maxTimer.Stop()
	}
	gopodRecord.active = false
	gopodRecord.chat, gopodRecord.log, gopodRecord.micOn = nil, nil, false
	gopodRecord.keptTakes, gopodRecord.takes = gopodRecord.takes, nil
	gopodRecord.Unlock()

	if micOn {
		if r, err := gopodMicCall(map[string]any{"cmd": "record_stop"}, gopodRecordMicStopWait); err != nil {
			logger.Println("GOPOD_RECORD_SESSION room mic stop: " + err.Error())
		} else if e := gopodMicReplyError(r); e != "" {
			logger.Println("GOPOD_RECORD_SESSION room mic stop: " + e)
		}
	}
	logger.Println(fmt.Sprintf("GOPOD_RECORD_SESSION stop reason=%s minutes=%.1f dir=%s", reason, took.Minutes(), dir))
	chat.Close()
	wlog.Close()
	gopodRecordWriteHTML(dir)
	os.Remove(filepath.Join(dir, gopodRecordMarkerRecording))
	return dir
}

// gopodRecordEndAndAsk: "Thanks Brobots" or max_minutes - stop, then ask.
func gopodRecordEndAndAsk(reason string) {
	dir := gopodRecordStop(reason)
	if dir == "" {
		return
	}
	_, uid, gid := gopodRecordRoot()
	pending := filepath.Join(dir, gopodRecordMarkerPending)
	os.WriteFile(pending, nil, 0o644)
	gopodRecordChown(uid, gid, pending)
	gopodRecord.Lock()
	gopodRecord.waiting, gopodRecord.dir = true, dir
	gopodRecord.decided = make(chan struct{})
	wait := time.Duration(gopodBrobotsSettings.RecordSession.AnswerWaitMinutes * float64(time.Minute))
	gopodRecord.answerTimer = time.AfterFunc(wait, func() { gopodRecordDecide(false, "no answer") })
	gopodRecord.Unlock()
	gopodWriteRecordState()
	s := gopodBrobotsSettings.RecordSession
	gopodRecordSay(s.AskBrobot, s.AskText)
}

// gopodRecordDecide keeps (zip) or deletes the folder waiting for an answer.
func gopodRecordDecide(keep bool, reason string) {
	gopodRecord.Lock()
	if !gopodRecord.waiting {
		gopodRecord.Unlock()
		return
	}
	gopodRecord.waiting = false
	if gopodRecord.answerTimer != nil {
		gopodRecord.answerTimer.Stop()
	}
	dir := gopodRecord.dir
	decided := gopodRecord.decided
	gopodRecord.decided = nil
	if decided != nil {
		defer close(decided)
	}
	gopodRecord.Unlock()
	os.Remove(filepath.Join(dir, gopodRecordMarkerPending))
	s := gopodBrobotsSettings.RecordSession
	if keep {
		gopodRecordKeep(dir)
	} else {
		os.RemoveAll(dir)
		gopodRecord.Lock()
		gopodRecord.keptTakes = nil
		gopodRecord.Unlock()
		logger.Println("GOPOD_RECORD_SESSION deleted reason=" + reason + " dir=" + dir)
	}
	gopodWriteRecordState()
	confirm := s.DeletedText
	if keep {
		confirm = s.KeptText
	}
	// 2026-10-11 (operator: "Goverlord: GOPOD Brain * Session recording
	// deleted/kept. You're welcome."): in the thanks ending, the answer is said
	// by Goverlord with the closing "You're welcome."; otherwise by Goverlord alone.
	if gopodRecordInEnding.Load() {
		gopodRecordConfirm.Store(confirm)
		return
	}
	if role := gopodBrobotsSettings.SignOff.SystemRole; role != "" {
		// 2026-10-23: the one voiced-line template (show and say: Generating, then lit).
		gopodReporterSay(role, confirm, true, true)
	}
}

// gopodRecordInEnding: the thanks ending is waiting on the keep-or-delete
// answer; gopodRecordConfirm: that answer's line, for "You're welcome.".
var gopodRecordInEnding atomic.Bool
var gopodRecordConfirm atomic.Value

// gopodRecordKeep zips the folder next to itself and deletes the loose files,
// and adds the session's answers to the take sheet as takes (operator
// 2026-10-06: "keep this session take?", gopod_take_sheet.go).
func gopodRecordKeep(dir string) {
	gopodRecord.Lock()
	takes := gopodRecord.keptTakes
	gopodRecord.keptTakes = nil
	gopodRecord.Unlock()
	if err := gopodRecordTakesToSheet(dir, takes); err != nil {
		logger.Println("GOPOD_TAKE_SHEET not added: " + err.Error())
	}
	zipPath, err := gopodRecordZip(dir)
	if err != nil {
		logger.Println("GOPOD_RECORD_SESSION zip failed, folder kept: " + err.Error())
		return
	}
	os.RemoveAll(dir)
	_, uid, gid := gopodRecordRoot()
	gopodRecordChown(uid, gid, zipPath)
	logger.Println("GOPOD_RECORD_SESSION kept zip=" + zipPath)
}

func gopodRecordZip(dir string) (string, error) {
	zipPath := dir + ".zip"
	tmp := zipPath + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	zw := zip.NewWriter(f)
	base := filepath.Base(dir)
	walkErr := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return err
		}
		head, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		head.Name = filepath.ToSlash(filepath.Join(base, rel))
		head.Method = zip.Deflate
		if strings.HasSuffix(info.Name(), ".wav") {
			head.Method = zip.Store // audio barely shrinks; storing is fast
		}
		w, err := zw.CreateHeader(head)
		if err != nil {
			return err
		}
		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()
		_, err = io.Copy(w, src)
		return err
	})
	closeErr := zw.Close()
	f.Close()
	if walkErr == nil {
		walkErr = closeErr
	}
	if walkErr != nil {
		os.Remove(tmp)
		return "", walkErr
	}
	return zipPath, os.Rename(tmp, zipPath)
}

// --- the answer ---

// gopodRecordAnswerKind: a line with every keep_words word is "keep", with every
// delete_words word "delete" (whole words, any order); both or neither: "".
func gopodRecordAnswerKind(text string) string {
	have := map[string]bool{}
	for _, w := range gopodAddressWords(text) {
		have[w] = true
	}
	all := func(words []string) bool {
		for _, w := range words {
			if !have[strings.ToLower(strings.TrimSpace(w))] {
				return false
			}
		}
		return len(words) > 0
	}
	keep, del := all(gopodBrobotsSettings.RecordSession.KeepWords), all(gopodBrobotsSettings.RecordSession.DeleteWords)
	switch {
	case keep && !del:
		return "keep"
	case del && !keep:
		return "delete"
	}
	return ""
}

// GopodRecordSessionAnswer handles a typed or heard line that answers the
// keep-or-delete question; true means the line is taken (no chat reply). Both
// Brobots hear a spoken answer: the second hearing, within
// addressing.first_heard_seconds, is taken quietly.
// how: the line's input ("Typed text", "Numpad text", a heard line) - a typed
// answer is shown and read out like every Operator line (2026-10-24, operator:
// "Operator to Goverlord * Keep the session recording * highlighting?"); a
// heard one is already shown (gopodHeardLine).
func GopodRecordSessionAnswer(text, how string) bool {
	kind := gopodRecordAnswerKind(text)
	if kind == "" {
		return false
	}
	gopodRecord.Lock()
	if !gopodRecord.waiting {
		recent := time.Since(gopodRecord.answeredAt) <= gopodSecondsToDuration(gopodBrobotsSettings.Addressing.FirstHeardSeconds)
		gopodRecord.Unlock()
		return recent
	}
	gopodRecord.answeredAt = time.Now()
	gopodRecord.Unlock()
	logger.Println("GOPOD_RECORD_SESSION answer=" + kind + " text=" + text)
	to := gopodBrobotsSettings.RecordSession.AnswerTo
	if how == gopodVoiceCommandHow {
		go gopodRecordDecide(kind == "keep", "answer")
		return true
	}
	gopodOperatorShow(to, text, how)
	go func() {
		gopodOperatorRead(to, text, how) // read out (and lit) first
		gopodRecordDecide(kind == "keep", "answer")
	}()
	return true
}

// --- what a session records ---

type gopodRecordLine struct {
	TimeMS int64  `json:"time_ms"`
	Time   string `json:"time"`
	Source string `json:"source"`
	Who    string `json:"who"`
	To     string `json:"to,omitempty"`
	Text   string `json:"text"`
	Spoken string `json:"spoken,omitempty"`
	FeedID int    `json:"feed_id,omitempty"`
}

func gopodRecordWrite(line gopodRecordLine) {
	gopodRecord.Lock()
	defer gopodRecord.Unlock()
	gopodRecordWriteLocked(line)
}

func gopodRecordWriteLocked(line gopodRecordLine) {
	if !gopodRecord.active || gopodRecord.chat == nil {
		return
	}
	if line.TimeMS == 0 {
		line.TimeMS = time.Now().UnixMilli()
	}
	line.Time = time.UnixMilli(line.TimeMS).Format("2006-01-02 15:04:05.000")
	data, err := json.Marshal(line)
	if err != nil {
		return
	}
	gopodRecord.chat.Write(append(data, '\n'))
}

// gopodRecordFeedLine: a line shown on the page's feed (songs, tools, typed
// lines, system lines).
func gopodRecordFeedLine(e gopodChatFeedEvent) {
	if e.Who == "Operator" {
		gopodRecord.Lock()
		gopodRecord.typedWords, gopodRecord.typedAt = gopodAddressWords(e.Text), time.Now()
		gopodRecord.Unlock()
	}
	gopodRecordWrite(gopodRecordLine{TimeMS: e.TimeMS, Source: "feed", Who: e.Who, To: e.To, Text: e.Text, FeedID: e.ID})
	gopodSessionLogFeed(e)
}

// gopodRecordExchange: a chat question and its answer (voice or typed).
func gopodRecordExchange(esn, question string, result gopodStreamResult) {
	name := gopodBrobotName(esn)
	// A typed question is already recorded, at the time it was typed (the feed
	// line); a spoken one is recorded here.
	gopodRecord.Lock()
	typed := time.Since(gopodRecord.typedAt) < gopodRecordTypedSameWindow && gopodSameQuestion(gopodRecord.typedWords, gopodAddressWords(question))
	input := "mic"
	if typed {
		input = "typed"
		if gopodRecord.typedHow != "" {
			input = gopodRecord.typedHow
		}
	}
	active := gopodRecord.active
	gopodRecord.Unlock()
	if !typed {
		gopodRecordWrite(gopodRecordLine{Source: "chat", Who: "Operator", To: name, Text: question})
		gopodSessionLogHeard(name, question)
	}
	gopodRecordWrite(gopodRecordLine{Source: "chat", Who: name, Text: result.Raw, Spoken: result.FilteredPacket})
	gopodSessionLogAnswer(esn, question, result)
	// The answer as a take, for the take sheet if the session is kept.
	if active && name != "Brobot" && !gopodIsThanks(question) {
		take := gopodLiveTake(esn, question, input, result)
		gopodRecord.Lock()
		if gopodRecord.active {
			gopodRecord.takes = append(gopodRecord.takes, take)
		}
		gopodRecord.Unlock()
	}
}

// gopodRecordOperatorInput notes how the next operator line came in (typed or
// numpad), for its take.
func gopodRecordOperatorInput(how string) {
	kind := "typed"
	if strings.Contains(strings.ToLower(how), "numpad") {
		kind = "numpad"
	}
	gopodRecord.Lock()
	gopodRecord.typedHow = kind
	gopodRecord.Unlock()
}

// gopodRecordLogTee writes every Wire-Pod log line to the session's wirepod.log.
// logger.Println offers each line on LogTrayChan without waiting; nothing else
// reads it, and the relay below takes each line at once so none is dropped
// while a line is written.
func gopodRecordLogTee() {
	ch := logger.GetLogTrayChan()
	if ch == nil {
		return
	}
	relay := make(chan string, 4096)
	go func() {
		for line := range relay {
			gopodSessionLogWire(line)
			gopodRecord.Lock()
			if gopodRecord.active && gopodRecord.log != nil {
				gopodRecord.log.WriteString(line)
			}
			gopodRecord.Unlock()
		}
	}()
	for line := range ch {
		select {
		case relay <- line:
		default:
		}
	}
}

func gopodRecordWriteHTML(dir string) {
	data, err := os.ReadFile(filepath.Join(dir, "chat.jsonl"))
	if err != nil {
		return
	}
	var b strings.Builder
	b.WriteString("<!doctype html>\n<html><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n<title>Brobots Session</title>\n<style>body{background:#1d1d1f;color:#eee;font:16px/1.4 sans-serif;margin:0;padding:16px}.l{margin:8px 0;padding:8px 12px;border-radius:8px;background:#2c2c2e;max-width:900px;white-space:pre-wrap}.t{color:#999;font-size:12px}.w{font-weight:bold}.Operator{background:#3a3a3a}.b1{background:#4a1018}.b2{background:#101a4a}</style></head><body>\n")
	b.WriteString("<h1>Brobots Session " + html.EscapeString(filepath.Base(dir)) + "</h1>\n")
	for _, raw := range strings.Split(string(data), "\n") {
		var l gopodRecordLine
		if json.Unmarshal([]byte(raw), &l) != nil || l.Text == "" {
			continue
		}
		cls := "l"
		switch l.Who {
		case "Operator":
			cls += " Operator"
		case "Brobot 1":
			cls += " b1"
		case "Brobot 2":
			cls += " b2"
		}
		who := l.Who
		if l.To != "" {
			who += " to " + l.To
		}
		b.WriteString("<div class=\"" + cls + "\"><div class=\"t\">" + html.EscapeString(l.Time) + "</div><span class=\"w\">" +
			html.EscapeString(who) + ":</span> " + html.EscapeString(l.Text) + "</div>\n")
	}
	b.WriteString("</body></html>\n")
	path := filepath.Join(dir, "chat.html")
	if os.WriteFile(path, []byte(b.String()), 0o644) == nil {
		_, uid, gid := gopodRecordRoot()
		gopodRecordChown(uid, gid, path)
	}
}

// --- the system lines (Brobot 1 asks, Brobot 2 confirms) ---

func gopodBrobotESN(name string) string {
	switch strings.TrimSpace(name) {
	case "Brobot 1":
		return gopodBrobot1ESN
	case "Brobot 2":
		return gopodBrobot2ESN
	}
	return ""
}

// gopodRecordSay shows the line on the page and has the Brobot say it.
func gopodRecordSay(brobot, text string) {
	esn := gopodBrobotESN(brobot)
	// 2026-10-24 (operator: "This is where Pip asks for an email, right? / Bro! *
	// not highlighted"): his bubble the way an answer's goes (GopodChatFeedAddLive),
	// and he speaks once it's on the page (gopodWaitReplyShown), so it lights as
	// he says it - "Bro!" was said before its bubble showed.
	GopodChatFeedAddLive(esn, text)
	var robot *vector.Vector
	for _, bot := range vars.BotInfo.Robots {
		if bot.Esn == esn {
			robot, _ = vector.New(vector.WithSerialNo(esn), vector.WithToken(bot.GUID), vector.WithTarget(bot.IPAddress+":443"))
			break
		}
	}
	if robot == nil {
		return
	}
	gopodChatHoldLetGo(esn)
	defer gopodChatHoldResume(esn)
	ctx := context.Background()
	granted := make(chan bool, 1)
	release := make(chan bool)
	gopodLogBControl(robot, ctx, granted, release)
	select {
	case <-granted:
		gopodWaitReplyShown(esn)
		// 2026-10-17 (operator, on the photo's "I see urns on the shelf ...":
		// "text still solid white, check for loose chaining"): said a sentence
		// at a time with the highlight's start and end, like a reply, so its
		// bubble is lit while said and shows as said after.
		GopodSpeakSentences(esn, text, -1, func(sentence string) {
			gopodLogSayText(robot, ctx, &vectorpb.SayTextRequest{
				// 2026-10-08: cleaned like a reply - emojis and symbols aren't
				// spoken (live 14:52: "Peace out ... 😎🎉🎤💥🔥" took 12.6 s).
				Text:           gopodApplyPronunciations(removeSpecialCharacters(sentence)),
				UseVectorVoice: true,
				DurationScalar: 1,
			})
		})
	case <-time.After(5 * time.Second):
		logger.Println("GOPOD_RECORD_SESSION " + brobot + " line not said: no behavior control")
	}
	select {
	case release <- true:
	case <-time.After(2 * time.Second):
	}
	// 2026-10-09: the battery after his line too (gopod_robot_cmd_log.go).
	gopodLogBattery(robot)
}

// --- state, folders, ownership ---

func gopodWriteRecordState() {
	gopodRecord.Lock()
	state := struct {
		On        bool  `json:"on"`
		Recording bool  `json:"recording"`
		Waiting   bool  `json:"waiting"`
		StartedMS int64 `json:"started_ms,omitempty"`
		// The session's start (gopod_session_log.go): the page's timer counts from it.
		SessionMS int64  `json:"session_started_ms,omitempty"`
		Folder    string `json:"folder,omitempty"`
	}{
		On:        gopodRecord.wanted || gopodRecord.active || gopodRecord.waiting,
		Recording: gopodRecord.active,
		Waiting:   gopodRecord.waiting,
	}
	if gopodRecord.active {
		state.StartedMS = gopodRecord.started.UnixMilli()
	}
	gopodRecord.Unlock()
	state.SessionMS = gopodSessionStartedMS()
	gopodRecord.Lock()
	if gopodRecord.active || gopodRecord.waiting {
		state.Folder = filepath.Base(gopodRecord.dir)
	}
	gopodRecord.Unlock()
	data, err := json.Marshal(state)
	if err != nil {
		return
	}
	tmp := gopodRecordStatePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err == nil {
		os.Rename(tmp, gopodRecordStatePath)
	}
}

// gopodRecordRoot: sessions_dir, with "~" as the home of the user who owns
// GOPOD's overlay folder (Wire-Pod runs as root), and that user's ids for the
// files written there.
func gopodRecordRoot() (string, int, int) {
	dir := strings.TrimSpace(os.Getenv("GOPOD_RECORD_SESSION_DIR"))
	uid, gid := -1, -1
	if info, err := os.Stat(gopodChipperRoot()); err == nil {
		if st, ok := info.Sys().(*syscall.Stat_t); ok {
			uid, gid = int(st.Uid), int(st.Gid)
		}
	}
	if dir == "" {
		dir = gopodBrobotsSettings.RecordSession.SessionsDir
		if strings.HasPrefix(dir, "~/") {
			home := ""
			if u, err := user.LookupId(strconv.Itoa(uid)); err == nil {
				home = u.HomeDir
			}
			dir = filepath.Join(home, dir[2:])
		}
	}
	return dir, uid, gid
}

func gopodRecordChown(uid, gid int, paths ...string) {
	if os.Geteuid() != 0 || uid < 0 {
		return
	}
	for _, p := range paths {
		os.Lchown(p, uid, gid)
	}
}

// gopodRecordCleanupLeftovers runs at Wire-Pod start: a folder still waiting for
// an answer is deleted (the default); one cut off mid-recording is zipped.
func gopodRecordCleanupLeftovers() {
	root, _, _ := gopodRecordRoot()
	dirs, _ := filepath.Glob(filepath.Join(root, "*", "*"+gopodBrobotsSettings.RecordSession.FolderSuffix))
	// A recording inside a session's folder (2026-10-11).
	nested, _ := filepath.Glob(filepath.Join(root, "*", "*", "*"+gopodBrobotsSettings.RecordSession.FolderSuffix))
	for _, dir := range append(dirs, nested...) {
		if _, err := os.Stat(filepath.Join(dir, gopodRecordMarkerPending)); err == nil {
			os.RemoveAll(dir)
			logger.Println("GOPOD_RECORD_SESSION deleted unanswered dir=" + dir)
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, gopodRecordMarkerRecording)); err == nil {
			gopodRecordWriteHTML(dir)
			os.Remove(filepath.Join(dir, gopodRecordMarkerRecording))
			gopodRecordKeep(dir)
		}
	}
}

// ## EOF

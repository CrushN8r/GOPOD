// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_session_log.go
package wirepod_ttr

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// GOLDEN - the session's own account (operator 2026-10-10: "I want a full run
// with a full account of all events of a session starting from timer 0 until
// ended, and timer reset to 0"; "extend/elevate runs/notations to a session *
// auto-log everything * recording/saving audio files requires Recording?
// checked"; "new session should be taken from same single source of truth,
// "_new_song" folder").
//
// A session starts at Wire-Pod start (the page's timer at 00:00:00) and ends
// after the thanks ending ("You're welcome."); the next one starts at 0. It is
// always logged, Record? or not, into record_session.sessions_dir/<date>/
// <time>_session/, laid out like a song folder (_new_song):
//   - story.md and knobs.json, composed as it happens: every chat line is a
//     step (its words in story.md, its kind, who and second in knobs.json,
//     "timeline": "srt"); each run is its own section;
//   - runs/: each run's own log and json, copied from its song's runs/ when it
//     ends; notation/: empty, as in a song;
//   - chat.jsonl (every line) and wirepod.log (every Wire-Pod log line).
// Record? adds the room audio (gopod_record_session.go) and turns the timer red.

type gopodSessionStep struct {
	StepID       string  `json:"step_id"`
	Note         string  `json:"note"`
	Speaker      string  `json:"speaker"`
	To           string  `json:"to,omitempty"`
	Via          string  `json:"via,omitempty"`
	Section      string  `json:"section"`
	StartSeconds float64 `json:"start_seconds"`
	// 2026-10-12 (operator: "should the cues and moves be logged into the score
	// so a replay includes them? * yes please"): an answer's words as the Brobot
	// got them, its cue marked ("happy___ words"), and a scripted answer's moves.
	Say   string           `json:"say,omitempty"`
	Moves []gopodDraftMove `json:"moves,omitempty"`
	text  string
}

var gopodSessionLog = struct {
	sync.Mutex
	dir     string
	started time.Time
	section string
	steps   []gopodSessionStep
	chat    *os.File
	log     *os.File
}{}

// gopodSessionLiveSection: the section for lines outside a run.
const gopodSessionLiveSection = "LIVE"

// GopodSessionLogStart opens a new session folder (and closes one still open).
func GopodSessionLogStart() {
	root, uid, gid := gopodRecordRoot()
	now := time.Now()
	dir := filepath.Join(root, now.Format("2006-01-02"), now.Format("150405")+"_session")
	for _, d := range []string{dir, filepath.Join(dir, "runs"), filepath.Join(dir, "notation")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			logger.Println("GOPOD_SESSION_LOG unable to make " + d + ": " + err.Error())
			return
		}
		gopodRecordChown(uid, gid, d)
	}
	chat, _ := os.OpenFile(filepath.Join(dir, "chat.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	wlog, _ := os.OpenFile(filepath.Join(dir, "wirepod.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	gopodRecordChown(uid, gid, filepath.Join(dir, "chat.jsonl"), filepath.Join(dir, "wirepod.log"))
	gopodSessionLog.Lock()
	gopodSessionLogCloseLocked()
	gopodSessionLog.dir, gopodSessionLog.started = dir, now
	gopodSessionLog.section, gopodSessionLog.steps = gopodSessionLiveSection, nil
	gopodSessionLog.chat, gopodSessionLog.log = chat, wlog
	gopodSessionLogWriteLocked()
	gopodSessionLog.Unlock()
	gopodWriteRecordState()
	logger.Println("GOPOD_SESSION_LOG start dir=" + dir)
}

// gopodSessionLogDir: the open session's folder ("" before the first one).
// 2026-10-11 (operator: "default to write ALL session files (logs, WAVs,
// etc...) to 'within that session's folder'"): numpad WAVs and a recording go here.
func gopodSessionLogDir() string {
	gopodSessionLog.Lock()
	defer gopodSessionLog.Unlock()
	return gopodSessionLog.dir
}

// GopodSessionLogEnd closes the session after the thanks ending and starts the
// next one (the timer back to 0).
func GopodSessionLogEnd(reason string) {
	gopodSessionLog.Lock()
	dir, took := gopodSessionLog.dir, time.Since(gopodSessionLog.started).Round(time.Second)
	gopodSessionLog.Unlock()
	logger.Println("GOPOD_SESSION_LOG end reason=" + reason + " took=" + took.String() + " dir=" + dir)
	GopodSessionLogStart()
}

func gopodSessionLogCloseLocked() {
	if gopodSessionLog.chat != nil {
		gopodSessionLog.chat.Close()
	}
	if gopodSessionLog.log != nil {
		gopodSessionLog.log.Close()
	}
	gopodSessionLog.chat, gopodSessionLog.log, gopodSessionLog.dir = nil, nil, ""
}

// gopodSessionStartedMS: the session's start, for the page's timer.
func gopodSessionStartedMS() int64 {
	gopodSessionLog.Lock()
	defer gopodSessionLog.Unlock()
	if gopodSessionLog.dir == "" {
		return 0
	}
	return gopodSessionLog.started.UnixMilli()
}

// gopodSessionLogLine: one chat line - written to chat.jsonl and added as a step.
func gopodSessionLogLine(line gopodRecordLine, note, via string, more ...func(*gopodSessionStep)) {
	if line.TimeMS == 0 {
		line.TimeMS = time.Now().UnixMilli()
	}
	line.Time = time.UnixMilli(line.TimeMS).Format("2006-01-02 15:04:05.000")
	gopodSessionLog.Lock()
	defer gopodSessionLog.Unlock()
	if gopodSessionLog.dir == "" {
		return
	}
	if data, err := json.Marshal(line); err == nil && gopodSessionLog.chat != nil {
		gopodSessionLog.chat.Write(append(data, '\n'))
	}
	if note == "" {
		return
	}
	at := float64(line.TimeMS-gopodSessionLog.started.UnixMilli()) / 1000
	n := len(gopodSessionLog.steps) + 1
	who := strings.ToLower(strings.NewReplacer(" ", "_", ",", "").Replace(line.Who))
	step := gopodSessionStep{
		StepID: fmt.Sprintf("s%04d_%s", n, who), Note: note, Speaker: line.Who, To: line.To, Via: via,
		Section: gopodSessionLog.section, StartSeconds: float64(int64(at*1000)) / 1000, text: line.Text,
	}
	for _, f := range more {
		f(&step)
	}
	gopodSessionLog.steps = append(gopodSessionLog.steps, step)
	gopodSessionLogWriteLocked()
}

// gopodSessionLogFeed: a chat feed line (every bubble Wire-Pod sends the page).
func gopodSessionLogFeed(e gopodChatFeedEvent) {
	note := map[string]string{"Operator": "operator_line", "Brobot 1": "say", "Brobot 2": "say", "Goverlord": "system_line",
		"Countdown": "countdown", "Bingo Ball": "show_ball", "Narrator": "narrator", "Chapter": "chapter", "Picture": "pictures",
		"Song Audio": "soundtrack", "Brobot 3": "reporter_line", "Brobot 4": "reporter_line"}[e.Who]
	if e.Who == gopodBrobotsSettings.ReporterVoices.Role {
		note = "" // a voice file playing: in chat.jsonl only
	}
	via := e.Via
	if e.By != "" {
		via = e.By
	}
	gopodSessionLogLine(gopodRecordLine{TimeMS: e.TimeMS, Source: "feed", Who: e.Who, To: e.To, Text: gopodShownText(e.Text), FeedID: e.ID}, note, via)
}

// gopodShownText: a line's words without its voice tag ("generating:wait:<file>|"
// or "generating:lit:<file>|", gopodSayShown) - what the session log keeps.
func gopodShownText(text string) string {
	m := gopodBrobotsSettings.ReporterVoices.GeneratingMarker
	for _, tag := range []string{"wait:", "lit:"} {
		if m != "" && strings.HasPrefix(text, m+tag) {
			rest := text[len(m+tag):]
			if bar := strings.Index(rest, "|"); bar > 0 && strings.HasPrefix(rest, "/") {
				return rest[bar+1:]
			}
			return rest
		}
	}
	return text
}

// gopodSessionLogAnswer: a Brobot's chat answer (live or scripted).
// The bubble's words, without the cue ("dartingEyes... GOPOD can" -> "GOPOD can").
func gopodSessionLogAnswer(esn, question string, result gopodStreamResult) {
	moves := gopodScriptTakeLoggedMoves(esn)
	gopodSessionLogLine(gopodRecordLine{Source: "chat", Who: gopodBrobotName(esn), Text: gopodDisplayReply(result.Raw), Spoken: result.FilteredPacket}, "answer", "",
		func(s *gopodSessionStep) { s.Say, s.Moves = strings.TrimSpace(result.Raw), moves })
}

// gopodSessionLogHeard: a line a Brobot heard himself (Hey Vector, the backpack
// button) - the page shows it from Wire-Pod's log, so it's logged here.
func gopodSessionLogHeard(to, text string) {
	gopodSessionLogLine(gopodRecordLine{Source: "chat", Who: "Operator", To: to, Text: text}, "operator_line", "Backpack")
}

// gopodSessionLogRunStart / End: a run is its own section; at its end, its
// song's run files are copied into the session's runs/.
func gopodSessionLogRunStart(name string) {
	shown := gopodRunText("{run}", name)
	if shown == "Run" {
		shown = name
	}
	gopodSessionLog.Lock()
	gopodSessionLog.section = strings.ToUpper(shown)
	gopodSessionLog.Unlock()
}

func gopodSessionLogRunEnd(started time.Time) {
	gopodSessionLog.Lock()
	gopodSessionLog.section = gopodSessionLiveSection
	dir := gopodSessionLog.dir
	gopodSessionLog.Unlock()
	if dir == "" {
		return
	}
	_, uid, gid := gopodRecordRoot()
	songs, _ := gopodAllSongs()
	for _, songDir := range songs {
		files, _ := filepath.Glob(filepath.Join(songDir, "runs", "*"))
		for _, f := range files {
			info, err := os.Stat(f)
			if err != nil || info.IsDir() || info.ModTime().Before(started) {
				continue
			}
			to := filepath.Join(dir, "runs", filepath.Base(songDir)+"_"+filepath.Base(f))
			if gopodCopyFile(f, to) == nil {
				gopodRecordChown(uid, gid, to)
			}
		}
	}
}

func gopodCopyFile(from, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(to)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// gopodSessionLogWire: one Wire-Pod log line into wirepod.log.
func gopodSessionLogWire(line string) {
	gopodSessionLog.Lock()
	if gopodSessionLog.log != nil {
		gopodSessionLog.log.WriteString(line)
	}
	gopodSessionLog.Unlock()
}

// gopodSessionLogWriteLocked rewrites story.md and knobs.json from the steps so
// far, on _new_song's knobs.json (its top keys) - the caller holds the lock.
func gopodSessionLogWriteLocked() {
	dir := gopodSessionLog.dir
	if dir == "" {
		return
	}
	started := gopodSessionLog.started
	top := map[string]any{}
	if data, err := os.ReadFile(filepath.Join(gopodSongsDir(), "_new_song", "knobs.json")); err == nil {
		_ = json.Unmarshal(data, &top)
	}
	top["song_id"] = "session_" + started.Format("2006-01-02_150405")
	top["song_type"] = "session"
	top["timeline"] = "srt"
	top["session_started"] = started.Format("2006-01-02 15:04:05")
	steps := gopodSessionLog.steps
	if steps == nil {
		steps = []gopodSessionStep{}
	}
	top["steps"] = steps
	if data, err := json.MarshalIndent(top, "", "  "); err == nil {
		gopodSessionLogWriteFile(filepath.Join(dir, "knobs.json"), append(data, '\n'))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Session %s\n\n**Status:** WIP (%s)\n\n", started.Format("2006-01-02 15:04:05"), started.Format("2006-01-02"))
	b.WriteString("Everything in this session from the timer's 00:00:00, as it happened (gopod_session_log.go).\n")
	b.WriteString("Each step's second from 00:00:00 is its `start_seconds` in knobs.json.\n")
	section := ""
	for _, s := range steps {
		if s.Section != section {
			section = s.Section
			fmt.Fprintf(&b, "\n---\n\n## SECTION: %s\n", section)
		}
		fmt.Fprintf(&b, "\n## STEP %s\n\n> TEXT: %s\n", s.StepID, strings.ReplaceAll(strings.TrimSpace(s.text), "\n", " "))
	}
	gopodSessionLogWriteFile(filepath.Join(dir, "story.md"), []byte(b.String()))
}

func gopodSessionLogWriteFile(path string, data []byte) {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return
	}
	_, uid, gid := gopodRecordRoot()
	gopodRecordChown(uid, gid, tmp)
	os.Rename(tmp, path)
}

// gopodSessionLogSteps: the steps so far, oldest first (tests).
func gopodSessionLogSteps() []gopodSessionStep {
	gopodSessionLog.Lock()
	defer gopodSessionLog.Unlock()
	out := append([]gopodSessionStep(nil), gopodSessionLog.steps...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].StartSeconds < out[j].StartSeconds })
	return out
}

// ## EOF

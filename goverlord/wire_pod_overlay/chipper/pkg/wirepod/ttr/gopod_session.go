// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_session.go
package wirepod_ttr

import (
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// GOLDEN - the session and its runs (operator 2026-10-10: "There is the
// wire-pod session (wpr 2 until next wpr 2) * Goverlord is the wpr-session
// orchestrator * What are the rules of orchestration (boundaries/single source
// of golden truth)? * What are runs? How do runs run within wpr-session run?";
// then "drop, yes, and one pass"). gopod_notes/GOPOD_SESSION_ORCHESTRATION_001.md.
//
//   - Rule A: Goverlord keeps the session's one record - its phase (warming,
//     ready, run) and the run going. Nothing else keeps its own copy.
//   - Rule B: one run at a time, and every run checks in and out here: the
//     draft (start interview), the song engine (start bingo test, start baby
//     brobots, pha0b live) and phcal live. A second start while one is going is
//     refused with session.busy_text.
//   - Rule C: a run owns the Brobots - no chat hold (gopod_chat_hold.go), no
//     live LLM answers (Rule D).
//   - Rule D: during a run only run words get through - stop run, thanks,
//     Record Session's answer, and the draft's own lines; anything else gets
//     session.run_gate_text once and is dropped.
//   - Rule F: a run that was stopped says session.stopped_text.
// A run that exits without checking out (a crash) is cleared the next time
// anyone looks: its process is gone.

type gopodRun struct {
	Name    string    `json:"name"`
	Kind    string    `json:"kind"` // "interview" (played by the chat), "song", "phcal"
	PID     int       `json:"pid,omitempty"`
	Started time.Time `json:"started"`
}

var gopodSession = struct {
	sync.Mutex
	run      *gopodRun
	gateSaid time.Time // when session.run_gate_text was last shown
}{}

// gopodRunNow: the run going, if any. A run whose process has gone is cleared.
func gopodRunNow() (gopodRun, bool) {
	gopodSession.Lock()
	defer gopodSession.Unlock()
	r := gopodSession.run
	if r == nil {
		return gopodRun{}, false
	}
	if r.PID > 0 && syscall.Kill(r.PID, 0) == syscall.ESRCH {
		logger.Println("GOPOD_SESSION run " + r.Name + " gone without checking out - cleared")
		gopodSession.run = nil
		return gopodRun{}, false
	}
	return *r, true
}

// gopodRunActive: any run is going.
func gopodRunActive() bool {
	_, ok := gopodRunNow()
	return ok
}

// gopodSongRunning: a song run (the golden song engine) is going.
func gopodSongRunning() bool {
	r, ok := gopodRunNow()
	return ok && r.Kind == "song"
}

// GopodSessionPhase: "warming", "run" or "ready".
func GopodSessionPhase() string {
	if gopodWarmGateClosed.Load() {
		return "warming"
	}
	if gopodRunActive() {
		return "run"
	}
	return "ready"
}

// GopodRunStart checks a run in. False (and the busy line said) when another
// run is going, or before LLM Ready; the second value names what's in the way.
func GopodRunStart(name, kind string, pid int) (bool, string) {
	if gopodWarmGateClosed.Load() && kind != "phcal" {
		logger.Println("GOPOD_SESSION run " + name + " refused: not ready yet")
		return false, "warming"
	}
	if r, ok := gopodRunNow(); ok {
		logger.Println("GOPOD_SESSION run " + name + " refused: " + r.Name + " is going")
		gopodSessionSay(gopodBrobotsSettings.Session.BusyText)
		return false, r.Name
	}
	gopodSession.Lock()
	gopodSession.run = &gopodRun{Name: name, Kind: kind, PID: pid, Started: time.Now()}
	gopodSession.Unlock()
	logger.Println("GOPOD_SESSION run " + name + " (" + kind + ") started")
	// 2026-10-13 (operator: "all songs (scripted runs) * remember speaker toggle
	// * toggle speaker (loop audio) off * after run, last thing is to set back to
	// remembered toggle"): the page's background loop goes quiet for the run.
	GopodChatFeedAdd(gopodBrobotsSettings.ReporterVoices.Role, gopodLoopPauseMark)
	gopodSessionLogRunStart(name)
	// Rule C: the run owns the Brobots.
	GopodChatHoldEnd("run " + name)
	// 2026-10-10 (operator: "[run-name] not-found/started/stopped").
	// Said before the run goes on (live 23:3x: "Interview started." came after
	// the reporters' first line).
	if _, named := gopodBrobotsSettings.Session.RunNames[name]; named {
		if text := gopodRunText(gopodBrobotsSettings.Session.StartedText, name); strings.TrimSpace(text) != "" {
			GopodSystemSayHeard(text, 6*time.Second) // inside the song's 10 s check-in
		}
	}
	return true, ""
}

// GopodRunEnd checks a run out. reason "stopped" says session.stopped_text.
func GopodRunEnd(name, reason string) {
	GopodRunEndThen(name, reason, nil)
}

// GopodRunEndThen is GopodRunEnd; a run that played to its end then runs then,
// after its complete_text is heard (2026-10-12: the interview's closer).
func GopodRunEndThen(name, reason string, then func()) {
	gopodSession.Lock()
	r := gopodSession.run
	if r == nil || (name != "" && r.Name != name) {
		gopodSession.Unlock()
		return
	}
	gopodSession.run = nil
	gopodSession.Unlock()
	// A run that ends mid warm-up lets the queued lines go.
	gopodWarmingEnd()
	logger.Println("GOPOD_SESSION run " + r.Name + " ended reason=" + reason + " took=" + time.Since(r.Started).Round(time.Second).String())
	go gopodSessionLogRunEnd(r.Started)
	// 2026-10-11 (operator: "check for "Run complete." after a successful run
	// completion * so us humans know the run is over * remember to replace
	// 'draft' running with []"): a run that played to its end says complete_text.
	// The last thing after the run's own end lines: the background loop back
	// to the page's speaker toggle.
	resume := func() { GopodChatFeedAdd(gopodBrobotsSettings.ReporterVoices.Role, gopodLoopResumeMark) }
	if _, named := gopodBrobotsSettings.Session.RunNames[r.Name]; named && reason == "done" {
		text := gopodRunText(gopodBrobotsSettings.Session.CompleteText, r.Name)
		go func() {
			if then == nil {
				GopodSystemSayHeard(text, 0)
			} else {
				GopodSystemSayHeard(text, 30*time.Second)
				then()
			}
			resume()
		}()
		return
	}
	// A thanks closing the run out goes on to its own ending instead.
	if reason == "stopped" && !gopodClosingOut.Load() {
		text := gopodRunText(gopodBrobotsSettings.Session.StoppedText, r.Name)
		go func() {
			GopodSystemSayHeard(text, 0)
			resume()
		}()
		return
	}
	resume()
}

// The page's background loop, quiet during a run (gopod_brobots_chat.js).
const (
	gopodLoopPauseMark  = "loop_pause"
	gopodLoopResumeMark = "loop_resume"
)

// gopodRunGateDrop: Rule D for a typed, numpad, page or voice line during a
// run - true when the line is dropped (session.run_gate_text shown, at most
// once every few seconds).
func gopodRunGateDrop(esn, how, text string) bool {
	r, ok := gopodRunNow()
	if !ok || gopodScriptAsking(text) {
		return false
	}
	logger.Println("GOPOD_SESSION dropped " + how + " for " + esn + " during run " + r.Name + ": " + text)
	gopodSession.Lock()
	say := time.Since(gopodSession.gateSaid) > 5*time.Second
	if say {
		gopodSession.gateSaid = time.Now()
	}
	gopodSession.Unlock()
	if say {
		gopodSessionSay(gopodBrobotsSettings.Session.RunGateText)
	}
	return true
}

// gopodRunText: a session line with {run} as the run's name (session.run_names),
// "Run" for a run not listed.
func gopodRunText(text, name string) string {
	shown := gopodBrobotsSettings.Session.RunNames[name]
	if shown == "" {
		shown = "Run"
	}
	return strings.ReplaceAll(text, "{run}", shown)
}

// gopodSessionSay: a Goverlord line, its bubble shown with its voice.
func gopodSessionSay(text string) {
	if strings.TrimSpace(text) != "" {
		go GopodSystemSayHeard(text, 0)
	}
}

// ## EOF

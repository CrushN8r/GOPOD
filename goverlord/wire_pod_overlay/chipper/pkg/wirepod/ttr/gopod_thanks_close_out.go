// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_thanks_close_out.go
package wirepod_ttr

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// gopodClosingOut: a thanks is waiting for a draft or song to stop.
var gopodClosingOut atomic.Bool

// gopodStopRuns asks a playing draft and a running song to stop at their next safe spot.
var gopodStopSeq atomic.Int64

func gopodStopRuns(why string) bool {
	gopodStopSeq.Add(1)
	stopped := false
	if gopodScriptPlaying.Load() {
		gopodScriptStop.Store(true)
		logger.Println("GOPOD_STOP_RUN " + why + ": the draft stops after the line being said")
		stopped = true
	}
	if gopodSongRunning() {
		ss := gopodBrobotsSettings.SongStop
		script := filepath.Join(gopodChipperRoot(), ss.SafeScript)
		if err := gopodAsOperator(exec.Command(script)).Run(); err != nil {
			logger.Println("GOPOD_STOP_RUN " + why + ": song stop failed: " + err.Error())
		} else {
			logger.Println("GOPOD_STOP_RUN " + why + ": the song stops at its next safe spot")
		}
		stopped = true
	}
	if gopodTourRunning.Load() {
		gopodTourStop.Store(true)
		logger.Println("GOPOD_STOP_RUN " + why + ": the tour stops")
		stopped = true
	}
	if gopodCommCheckRunning.Load() {
		gopodCommCheckStop.Store(true)
		logger.Println("GOPOD_STOP_RUN " + why + ": the comm check stops")
		stopped = true
	}
	if n := gopodLineQueueClear(); n > 0 {
		logger.Println("GOPOD_STOP_RUN " + why + ": " + strconv.Itoa(n) + " waiting questions dropped")
	}
	if stopped {
		gopodTurnDrop(why)
		GopodSkipped()
		gopodChatFeedAddTurn(0, gopodBrobotsSettings.ReporterVoices.Role, gopodSkipMarker+why)
	}
	return stopped
}

// gopodWaitRunsStopped waits for the draft and the song to have stopped.
func gopodWaitRunsStopped() {
	most := gopodSecondsToDuration(gopodBrobotsSettings.SongStop.WaitSeconds) + 8*time.Minute
	deadline := time.Now().Add(most)
	for (gopodScriptPlaying.Load() || gopodSongRunning()) && time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
	}
	logger.Println("GOPOD_STOP_RUN runs closed out")
}

// GopodEscape: the page's Esc - a run going (a song, a draft, the tour, the comm check) stops, as
// stop run does, and Goverlord says session.
func GopodEscape() bool {
	defer gopodTurnOwn(0)()
	songRun := gopodRunActive() || gopodScriptPlaying.Load()
	if !gopodStopRuns("Esc") {
		return false
	}
	if !songRun {
		go func() {
			defer gopodTurnOwn(0)()
			gopodStoppedSay()
		}()
	}
	return true
}

// gopodStoppedSay: Goverlord's session.stopped_text for a stop with no song run to check out (the
// tour, the comm check) - a song run says it at its check-out.
func gopodStoppedSay() {
	defer gopodTurnOwn(0)()
	GopodSystemSayHeard(gopodRunText(gopodBrobotsSettings.Session.StoppedText, ""), 0)
}

// gopodStopRunCommand: a typed or numpad song_stop.stop_words line.
func gopodStopRunMatch(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	for _, w := range gopodBrobotsSettings.SongStop.StopWords {
		if w != "" && gopodHasWholeWords(lower, strings.ToLower(w)) {
			return true
		}
	}
	return false
}

func gopodStopRunCommand(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	for _, w := range gopodBrobotsSettings.SongStop.StopWords {
		if w != "" && gopodHasWholeWords(lower, strings.ToLower(w)) {
			if !gopodStopRuns("stop run") {
				logger.Println("GOPOD_STOP_RUN nothing running")
				gopodChatFeedAddTurn(0, gopodBrobotsSettings.ReporterVoices.Role, gopodSkipMarker+"stop run")
				gopodSessionSay(gopodBrobotsSettings.Session.NoRunText)
			}
			return true
		}
	}
	return false
}

// gopodThanksCloseOut: a thanks while a draft or song is going closes them out first, then the
// thanks goes on (again) - true when it took the line.
func gopodThanksCloseOut(esn, text, how string) bool {
	if !gopodIsThanks(text) || gopodScriptAsking(text) {
		return false
	}
	if !gopodScriptPlaying.Load() && !gopodSongRunning() {
		return false
	}
	if !gopodClosingOut.CompareAndSwap(false, true) {
		logger.Println("GOPOD_STOP_RUN thanks already closing out: " + text)
		return true
	}
	gopodStopRuns("thanks")
	go func() {
		gopodWaitRunsStopped()
		gopodClosingOut.Store(false)
		if err := gopodOperatorLine(esn, text, how); err != nil {
			logger.Println("GOPOD_STOP_RUN thanks after the close-out: " + err.Error())
		}
	}()
	return true
}

// gopodScriptAsking: the playing draft is putting this line in itself.
func gopodScriptAsking(text string) bool {
	gopodScript.Lock()
	defer gopodScript.Unlock()
	return gopodScript.asking != "" && strings.TrimSpace(gopodScript.asking) == strings.TrimSpace(text)
}

// gopodThanksEnding: the ending after a thanks, once - the brother's thanks (to "All") doesn't run
// it again until a new question is asked.
func gopodThanksEnding() {
	gopodReporter.Lock()
	if gopodReporter.ended {
		gopodReporter.Unlock()
		return
	}
	gopodReporter.ended = true
	gopodReporter.Unlock()
	time.Sleep(gopodSecondsToDuration(gopodBrobotsSettings.SignOff.SafeSpotSeconds))
	for waited := time.Duration(0); gopodReplyBusy() && waited < gopodTiming("thanks_reply_wait_max_seconds"); waited += 250 * time.Millisecond {
		time.Sleep(250 * time.Millisecond)
	}
	gopodCoverageScore()
	gopodReporterClosing()
	gopodSignOffLines()
	gopodRecordInEnding.Store(true)
	GopodRecordSessionThanks()
	gopodRecordWaitDecided()
	gopodRecordInEnding.Store(false)
	gopodDraftRunCompleted()
	gopodYoureWelcome()
	GopodSessionLogEnd("thanks")
	if gopodBrobotsSettings.Game.ResetOn == "thanks" {
		gopodGameReset("thanks")
	}
}

// gopodYoureWelcome closes the thanks process: sign_off.welcome_text in a Goverlord bubble and
// voice, after everything before it has been heard.
func gopodYoureWelcome() {
	so := gopodBrobotsSettings.SignOff
	if so.SystemRole == "" || so.WelcomeText == "" {
		return
	}
	text := so.WelcomeText
	if confirm, _ := gopodRecordConfirm.Swap("").(string); confirm != "" {
		text = confirm + " " + text
	} else if nothing := strings.TrimSpace(so.NothingRecordedText); nothing != "" {
		text = nothing + " " + text
	}
	gopodReporterSay(so.SystemRole, text, true, true)
}

// ## EOF

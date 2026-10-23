// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_thanks_close_out.go
package wirepod_ttr

import (
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// "Thanks Brobots" is its own process, draft or not (operator 2026-10-09: "thx
// bros process looks for active draft run to close out * upon successful close
// out of draft run, thx bros can then end with your welcome"; "Build chat
// 'thanks' for songs as well"; "add 'stop run' as an input command for
// songs/runs"):
//  1. a playing draft stops after the line being said, and a running song (the
//     song_stop.engine) stops at its next safe spot (song_stop.safe_script);
//  2. only then does the thanks go to the Brobots - the sign-off and the ending
//     (kgsim.go), "Draft run completed." when a draft played;
//  3. Goverlord's sign_off.welcome_text closes it, always (gopodYoureWelcome).
// song_stop.stop_words does step 1 alone.

// gopodClosingOut: a thanks is waiting for a draft or song to stop.
var gopodClosingOut atomic.Bool

// gopodStopRuns asks a playing draft and a running song to stop at their next
// safe spot. False when neither was going.
func gopodStopRuns(why string) bool {
	stopped := false
	if gopodScriptPlaying.Load() {
		gopodScriptStop.Store(true)
		logger.Println("GOPOD_STOP_RUN " + why + ": the draft stops after the line being said")
		stopped = true
	}
	if gopodSongRunning() {
		ss := gopodBrobotsSettings.SongStop
		script := filepath.Join(gopodChipperRoot(), ss.SafeScript)
		if err := exec.Command(script).Run(); err != nil {
			logger.Println("GOPOD_STOP_RUN " + why + ": song stop failed: " + err.Error())
		} else {
			logger.Println("GOPOD_STOP_RUN " + why + ": the song stops at its next safe spot")
		}
		stopped = true
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

// gopodStopRunCommand: a typed or numpad song_stop.stop_words line. True when it was one.
// gopodStopRunMatch: text is a stop command (song_stop.stop_words).
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
				// 2026-10-10 (operator: "stop run was heard, no system message"):
				// a stop with nothing to stop still gets an answer.
				gopodSessionSay(gopodBrobotsSettings.Session.NoRunText)
			}
			return true
		}
	}
	return false
}

// gopodThanksCloseOut: a thanks while a draft or song is going closes them out
// first, then the thanks goes on (again) - true when it took the line. The
// draft's own question lines are never taken.
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

// gopodThanksEnding: the ending after a thanks, once - the brother's thanks
// (to "All") doesn't run it again until a new question is asked.
func gopodThanksEnding() {
	gopodReporter.Lock()
	if gopodReporter.ended {
		gopodReporter.Unlock()
		return
	}
	gopodReporter.ended = true
	gopodReporter.Unlock()
	// 2026-10-16 (operator: "\"thanks guys\" anywhere in a line ends the session,
	// at the next best safe spot"): the ending waits until no Brobot is answering -
	// to All, Brobot 2's sign-off first (he starts within sign_off.safe_spot_seconds
	// of Brobot 1's end); his own thanks then finds the ending already going.
	time.Sleep(gopodSecondsToDuration(gopodBrobotsSettings.SignOff.SafeSpotSeconds))
	for waited := time.Duration(0); gopodReplyBusy() && waited < 5*time.Minute; waited += 250 * time.Millisecond {
		time.Sleep(250 * time.Millisecond)
	}
	// The session's value-point score first (gopod_value_points.go).
	gopodCoverageScore()
	// 2026-10-08 (operator: "New ending: 1/5/2/3/4/Operator bubble"): the
	// reporters' closing, the brothers' extra lines, keep or delete, then
	// "Draft run completed." (when a draft played) and "You're welcome.".
	gopodReporterClosing()
	gopodSignOffLines()
	gopodRecordInEnding.Store(true)
	GopodRecordSessionThanks()
	gopodRecordWaitDecided()
	gopodRecordInEnding.Store(false)
	gopodDraftRunCompleted()
	gopodYoureWelcome()
	// The session ends here; the next one starts at 00:00:00 (gopod_session_log.go).
	GopodSessionLogEnd("thanks")
	// 2026-10-23 (A4): and the game starts at 0 (game.reset_on, gopod_game.go).
	if gopodBrobotsSettings.Game.ResetOn == "thanks" {
		gopodGameReset("thanks")
	}
}

// gopodYoureWelcome closes the thanks process: sign_off.welcome_text in a
// Goverlord bubble and voice, after everything before it has been heard.
func gopodYoureWelcome() {
	so := gopodBrobotsSettings.SignOff
	if so.SystemRole == "" || so.WelcomeText == "" {
		return
	}
	text := so.WelcomeText
	if confirm, _ := gopodRecordConfirm.Swap("").(string); confirm != "" {
		text = confirm + " " + text
	}
	// 2026-10-23: the one voiced-line template (show and say: Generating, then lit).
	gopodReporterSay(so.SystemRole, text, true, true)
}

// ## EOF

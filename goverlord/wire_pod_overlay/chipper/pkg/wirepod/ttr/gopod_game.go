// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_game.go
package wirepod_ttr

import (
	"strings"
	"sync"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// Win and reset (2026-10-23, FINISH_PLAN_001.md step A4; operator: "aim to reveal
// all value points as the grand prize * Go on A4"): the game's grand prize - every
// value point revealed - gets Goverlord's game.win_text (with the session score),
// once per session, after the answer that revealed the last one has been said.
// The session's end ("thanks", gopodThanksEnding) resets the game: the score,
// the points said, the reminders, the chapters cleared and the eggs found all go
// back to 0 for the next session.

var gopodGame = struct {
	sync.Mutex
	won bool
}{}

// gopodGameAllRevealed: every value point of the game has been revealed.
func gopodGameAllRevealed() bool {
	names, counts := gopodPackPointCounts()
	if len(names) == 0 {
		return false
	}
	gopodCoverage.Lock()
	defer gopodCoverage.Unlock()
	for _, n := range names {
		if gopodCoverage.revealed[n] < counts[n] {
			return false
		}
	}
	return true
}

// gopodGameWinText: the win line with the score filled in, "" when there's none.
func gopodGameWinText() string {
	text := strings.TrimSpace(gopodBrobotsSettings.Game.WinText)
	if text == "" {
		return ""
	}
	score, _ := gopodCoverageScoreText()
	return strings.ReplaceAll(text, "{score}", score)
}

// gopodGameWinCheck: after an answer has been said - the grand prize, once.
// True when it was announced now.
func gopodGameWinCheck() bool {
	if gopodScriptPlaying.Load() || gopodRunActive() || !gopodGameAllRevealed() {
		return false
	}
	gopodGame.Lock()
	if gopodGame.won {
		gopodGame.Unlock()
		return false
	}
	gopodGame.won = true
	gopodGame.Unlock()
	text := gopodGameWinText()
	if text == "" {
		return false
	}
	logger.Println("GOPOD_GAME grand prize: " + text)
	if role := gopodBrobotsSettings.SignOff.SystemRole; role != "" {
		gopodReporterSay(role, text, true, false)
	}
	return true
}

// gopodGameReset: a new session - the game starts at 0.
func gopodGameReset(why string) {
	gopodCoverage.Lock()
	gopodCoverage.revealed = map[string]int{}
	gopodCoverage.reminded = map[string]bool{}
	gopodCoverage.scored = false
	gopodCoverageSaid = map[string]map[int]bool{}
	gopodCoverage.Unlock()
	gopodChapterCleared.Lock()
	gopodChapterCleared.said = map[string]bool{}
	gopodChapterCleared.Unlock()
	gopodInterviewShown.Lock()
	gopodInterviewShown.pictures = map[string]int{}
	gopodInterviewShown.Unlock()
	gopodGame.Lock()
	gopodGame.won = false
	gopodGame.Unlock()
	logger.Println("GOPOD_GAME reset (" + why + "): the next session starts at 0")
}

// ## EOF

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_game.go
package wirepod_ttr

import (
	"strings"
	"sync"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

var gopodGame = struct {
	sync.Mutex
	won, masterWon bool
}{}

// gopodGameEggsFound: the game's own eggs all found (gopodEggList), and the hidden ones too
// (topic_pictures.not_listed that are eggs).
func gopodGameEggsFound() (game, hidden bool) {
	game, hidden = true, true
	listed := map[string]bool{}
	for _, e := range gopodEggList() {
		listed[e.Key] = true
		if !gopodEggFound(e.Key) {
			game = false
		}
	}
	for _, k := range gopodBrobotsSettings.TopicPictures.NotListed {
		if !listed[k] && !gopodNotEgg(k) && !gopodEggFound(k) {
			hidden = false
		}
	}
	return game, hidden
}

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

// gopodGameWinCheck: after every answer to a line has been said - the prizes, each once a session:
// the Grand Prize for every value point and every game egg; the Master Grand Prize for those and
// every hidden egg too.
func gopodGameWinCheck() bool {
	if gopodScriptPlaying.Load() || gopodRunActive() || !gopodGameAllRevealed() {
		return false
	}
	gameEggs, hiddenEggs := gopodGameEggsFound()
	if !gameEggs {
		return false
	}
	gopodGame.Lock()
	grand := !gopodGame.won
	master := hiddenEggs && !gopodGame.masterWon
	gopodGame.won = true
	if master {
		gopodGame.masterWon = true
	}
	gopodGame.Unlock()
	said := false
	if grand {
		said = gopodGameAnnounce("grand prize", gopodGameWinText()) || said
	}
	if master {
		said = gopodGameAnnounce("master grand prize", gopodGameFill(gopodBrobotsSettings.Game.MasterText)) || said
	}
	return said
}

// gopodGameAnnounce: Goverlord says a prize line.
func gopodGameAnnounce(what, text string) bool {
	if text == "" {
		return false
	}
	logger.Println("GOPOD_GAME " + what + ": " + text)
	if role := gopodBrobotsSettings.SignOff.SystemRole; role != "" {
		gopodReporterSay(role, text, true, false)
	}
	return true
}

// gopodGameFill: a prize line with the score filled in.
func gopodGameFill(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	score, _ := gopodCoverageScoreText()
	return strings.ReplaceAll(text, "{score}", score)
}

// gopodGameReset: a new session - the game starts at 0.
func gopodGameReset(why string) {
	gopodBrainReset(why != "thanks")
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
	gopodInterviewShown.revealed = nil
	gopodInterviewShown.Unlock()
	gopodGame.Lock()
	gopodGame.won, gopodGame.masterWon = false, false
	gopodGame.Unlock()
	logger.Println("GOPOD_GAME reset (" + why + "): the next session starts at 0")
}

// ## EOF

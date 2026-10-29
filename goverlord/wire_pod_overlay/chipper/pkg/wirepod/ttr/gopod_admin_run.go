// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_admin_run.go
package wirepod_ttr

// The 100% admin run, a standing self-check (2026-10-28, operator: "Yes!! Make
// the 100% admin run a standing self-check * each point not revealed * each egg
// not found * each prize not given * each reminder that went dry * add this
// 'admin run standing self-check' to GOPOD system item"). system-test-admin-run
// plays system_tests.admin_run.file - every line, in order - through the game's
// own code, as the room would: no robots, no model, nothing said. A song line
// must be a run card's start; a line to a Brobot gets his points (the same pick
// as live), the egg it would pop (its chapter's or topic's, Explain the Math,
// Get Photo), and the reminder after it. The live game is saved first and put
// back after, so a session's score is untouched.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// gopodAdminRunResult: what the run found missing.
type gopodAdminRunResult struct {
	Lines, Points, PointsWant, Eggs, EggsWant int
	Grand, Master                              bool
	Missed                                     []string // points not revealed
	EggsMissed, RunsMissed, Dry                []string
	Keys                                       int
	KeysMissed                                 []string
}

// gopodAdminRunEgg: the egg a line pops - its topics' chapter (or the topic
// itself), Explain the Math, Get Photo; "" for none.
func gopodAdminRunEggs(text string) []string {
	var out []string
	for _, p := range loadScenarioPacketsForMessage(text) {
		key, _ := gopodTopicChapter(p.Name)
		if key == "" {
			key = p.Name
		}
		out = append(out, key)
	}
	if _, ok := gopodMathExplain(text); ok {
		out = append(out, "math")
	}
	if gopodPhotoKind(text) == "ask" {
		out = append(out, "photo")
	}
	return out
}

// gopodAdminRunKey: what's wrong with a numpad key's line, "" for nothing.
func gopodAdminRunKey(key, text string) string {
	brobot, target, err := gopodNumpadBrobot(key)
	if err != nil {
		return err.Error()
	}
	if !strings.HasPrefix(target, gopodPersonaMark) {
		if gopodBrobotName(target) != brobot {
			return "goes to " + target + ", not " + brobot
		}
		return ""
	}
	host := strings.TrimPrefix(target, gopodPersonaMark)
	if at, _ := gopodEggWord(strings.ToLower(text), gopodBrobotsSettings.Managers.WhoWords); at >= 0 &&
		strings.TrimSpace(gopodBrobotsSettings.Managers.Personas[host].WhoLine) == "" {
		return host + " has no who line"
	}
	return ""
}

// gopodAdminRunPlay plays the run; the live game is put back after.
func gopodAdminRunPlay() (gopodAdminRunResult, error) {
	var res gopodAdminRunResult
	path := strings.TrimSpace(gopodBrobotsSettings.SystemTests.AdminRun.File)
	if path == "" {
		return res, fmt.Errorf("system_tests.admin_run.file is empty")
	}
	data, err := os.ReadFile(filepath.Join(gopodChipperRoot(), path))
	if err != nil {
		return res, err
	}
	if gopodRunActive() || gopodScriptPlaying.Load() {
		return res, fmt.Errorf("a run is going - try again after it")
	}
	// Save the live game, play from 0, put it back.
	gopodCoverage.Lock()
	keepRevealed, keepReminded, keepScored := gopodCoverage.revealed, gopodCoverage.reminded, gopodCoverage.scored
	keepSaid := gopodCoverageSaid
	gopodCoverage.Unlock()
	gopodChapterCleared.Lock()
	keepCleared := gopodChapterCleared.said
	gopodChapterCleared.Unlock()
	gopodInterviewShown.Lock()
	keepShown, keepShownOrder := gopodInterviewShown.pictures, gopodInterviewShown.revealed
	gopodInterviewShown.Unlock()
	gopodGiven.Lock()
	keepGiven := gopodGiven.points
	gopodGiven.points = map[string][]string{}
	gopodGiven.Unlock()
	gopodGame.Lock()
	keepWon, keepMaster := gopodGame.won, gopodGame.masterWon
	gopodGame.Unlock()
	gopodGameReset("admin run self-check")
	defer func() {
		gopodCoverage.Lock()
		gopodCoverage.revealed, gopodCoverage.reminded, gopodCoverage.scored = keepRevealed, keepReminded, keepScored
		gopodCoverageSaid = keepSaid
		gopodCoverage.Unlock()
		gopodChapterCleared.Lock()
		gopodChapterCleared.said = keepCleared
		gopodChapterCleared.Unlock()
		gopodInterviewShown.Lock()
		gopodInterviewShown.pictures, gopodInterviewShown.revealed = keepShown, keepShownOrder
		gopodInterviewShown.Unlock()
		gopodGiven.Lock()
		gopodGiven.points = keepGiven
		gopodGiven.Unlock()
		gopodGame.Lock()
		gopodGame.won, gopodGame.masterWon = keepWon, keepMaster
		gopodGame.Unlock()
	}()

	labels := gopodBrobotsSettings.InterviewChapters.TopicLabels
	starts := map[string]bool{}
	for _, r := range gopodBrobotsSettings.Game.Runs {
		starts[strings.ToLower(strings.TrimSpace(r.Start))] = true
	}
	eggKeys := map[string]bool{}
	for _, e := range gopodEggList() {
		eggKeys[e.Key] = true
	}
	for _, k := range gopodBrobotsSettings.TopicPictures.NotListed {
		if !gopodNotEgg(k) {
			eggKeys[k] = true
		}
	}
	got := map[string]bool{}
	for _, raw := range strings.Split(string(data), "\n") {
		raw = strings.TrimSpace(raw)
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		f := strings.SplitN(raw, "|", 3)
		if len(f) != 3 {
			continue
		}
		res.Lines++
		to, line := strings.TrimSpace(f[1]), strings.TrimSpace(f[2])
		if to == "" { // a run
			if !starts[strings.ToLower(line)] {
				res.RunsMissed = append(res.RunsMissed, line)
			}
			continue
		}
		// 2026-10-28 (operator: "Can you add a simulated KP1-9 to the test?"): a
		// numpad key, as holding it - a Brobot's key reaches that Brobot; a
		// persona's key reaches its persona, and "who are you" gets its who line.
		if strings.HasPrefix(strings.ToUpper(to), "KP") {
			res.Keys++
			if miss := gopodAdminRunKey(strings.ToUpper(to), gopodSpellNames(line)); miss != "" {
				res.KeysMissed = append(res.KeysMissed, to+": "+miss)
			}
			continue
		}
		esn := gopodBrobotESN(to)
		text := gopodSpellNames(line)
		gopodQuestionFactsFor(esn, text)
		// The reminder after his answer: dry when it names no topic while
		// points are left.
		rem := gopodReminderLine(text)
		for _, p := range gopodGivenPoints(esn) {
			got[p] = true
		}
		for _, k := range gopodAdminRunEggs(text) {
			if eggKeys[k] {
				gopodInterviewShown.Lock()
				gopodInterviewShown.pictures[k]++
				gopodInterviewShown.Unlock()
			}
		}
		// Dry (the live 2026-10-27 find: from question 8 on, only eggs were
		// named): a topic question whose reminder names no topic or chapter -
		// or any empty reminder - while points are left. A line on no topic
		// (a song's ask, the photo, a math entry) names eggs, as it should.
		named := false
		for _, l := range labels {
			if l != "" && strings.Contains(rem, l) {
				named = true
			}
		}
		for _, c := range gopodBrobotsSettings.InterviewChapters.Chapters {
			if c.Title != "" && strings.Contains(rem, c.Title) {
				named = true
			}
		}
		topical := len(loadScenarioPacketsForMessage(text)) > 0
		if !gopodGameAllRevealed() && !gopodIsThanks(text) && (strings.TrimSpace(rem) == "" || (topical && !named)) {
			res.Dry = append(res.Dry, fmt.Sprintf("line %d (%s)", res.Lines, line))
		}
	}
	for _, cp := range gopodBrobotsSettings.ValuePoints.CorePoints {
		res.PointsWant++
		parts := strings.SplitN(cp, "#", 2)
		found := false
		for g := range got {
			if len(parts) == 2 && strings.HasPrefix(g, parts[0]+" "+parts[1]+"/") {
				found = true
			}
		}
		if found {
			res.Points++
		} else {
			res.Missed = append(res.Missed, cp)
		}
	}
	for k := range eggKeys {
		res.EggsWant++
		if gopodEggFound(k) {
			res.Eggs++
		} else {
			res.EggsMissed = append(res.EggsMissed, k)
		}
	}
	gameEggs, hiddenEggs := gopodGameEggsFound()
	res.Grand = gopodGameAllRevealed() && gameEggs
	res.Master = res.Grand && hiddenEggs
	return res, nil
}

// gopodAdminRunLine: the test's answer - PASS when nothing is missing.
func gopodAdminRunLine() (bool, string) {
	r, err := gopodAdminRunPlay()
	if err != nil {
		return false, err.Error()
	}
	head := fmt.Sprintf("%d lines: value points %d/%d, eggs %d/%d, Grand Prize %s, Master Grand Prize %s, numpad keys %d/%d", r.Lines, r.Points, r.PointsWant, r.Eggs, r.EggsWant, gopodYesNo(r.Grand), gopodYesNo(r.Master), r.Keys-len(r.KeysMissed), r.Keys)
	var miss []string
	if len(r.Missed) > 0 {
		miss = append(miss, "points not revealed: "+strings.Join(r.Missed, ", "))
	}
	if len(r.EggsMissed) > 0 {
		miss = append(miss, "eggs not found: "+strings.Join(r.EggsMissed, ", "))
	}
	if !r.Grand {
		miss = append(miss, "Grand Prize not given")
	}
	if !r.Master {
		miss = append(miss, "Master Grand Prize not given")
	}
	if len(r.Dry) > 0 {
		miss = append(miss, "reminders dry: "+strings.Join(r.Dry, ", "))
	}
	if len(r.KeysMissed) > 0 {
		miss = append(miss, "numpad keys: "+strings.Join(r.KeysMissed, ", "))
	}
	if len(r.RunsMissed) > 0 {
		miss = append(miss, "not a run's start: "+strings.Join(r.RunsMissed, ", "))
	}
	if len(miss) == 0 {
		return true, head
	}
	return false, head + " | " + strings.Join(miss, " | ")
}

func gopodYesNo(b bool) string {
	if b {
		return "given"
	}
	return "NOT given"
}

// ## EOF

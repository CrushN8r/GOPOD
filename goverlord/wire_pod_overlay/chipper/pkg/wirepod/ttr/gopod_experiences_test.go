// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_experiences_test.go
package wirepod_ttr

import (
	"strings"
	"testing"
)

// 2026-10-25 (operator: "Make pre-[sleep experience] > [nap song] >
// post-[sleep experience]"): the nap, the Bingo way (gopod_experiences.go).
func TestNapExperienceChain(t *testing.T) {
	var nap *gopodExperience
	for _, e := range gopodExperienceList() {
		if e.Name == "nap" {
			nap = e
		}
	}
	if nap == nil {
		t.Fatal("no nap experience in settings")
	}
	flag := gopodExperiencePlayed(nap)
	saved := flag.Load()
	defer flag.Store(saved)
	flag.Store(false)
	q := "did you have any dreams?" // 2026-11-24: an experience ask (brain_chain.experience_asks)
	if gopodIsBingoQuestion(q) || !gopodIsExperienceQuestion(nap, q) {
		t.Fatal("a dreams question is the nap's, not Bingo's")
	}
	// Not played: its one line says so (brain_chain.experience_no).
	no := gopodExperienceFacts(gopodBrobot1ESN, q)
	if gopodExperienceInMemory(nap) == false && (len(no) != 1 || !strings.Contains(no[0], "not played")) {
		t.Errorf("before the nap: %q", no)
	}
	// The operator's own line doesn't count; the nap's first caption does.
	gopodExperienceMarkPlayed("Operator", nap.RunMarker)
	if flag.Load() {
		t.Error("the operator's line counted as the nap")
	}
	gopodExperienceMarkPlayed("Narrator", "Cupped in my hands,|are a pair of sleepy robots.")
	yes := gopodExperienceFacts(gopodBrobot1ESN, q)
	if !flag.Load() || len(yes) != 1 || !strings.Contains(yes[0], "you played it") {
		t.Errorf("after the nap: played, its one line %q", yes)
	}
	// Played: his closing remark is his take, word for word.
	if got := gopodExperienceClosing(gopodBrobot1ESN, q); got != strings.TrimSpace(nap.Closing["Brobot 1"]) {
		t.Errorf("after the nap, his take %q", got)
	}
}

// 2026-10-25 (operator: "' + start baby brobots' stays in 2nd brobot response
// regardless"): the nap's move goes with the second answer, whoever gives it.
func TestNapMoveInSecondAnswer(t *testing.T) {
	var nap *gopodExperience
	for _, e := range gopodExperienceList() {
		if e.Name == "nap" {
			nap = e
		}
	}
	flag := gopodExperiencePlayed(nap)
	saved := flag.Load()
	defer flag.Store(saved)
	flag.Store(false)
	if gopodExperienceInMemory(nap) {
		t.Skip("a nap run in this test's memory")
	}
	q := "how was dreaming?"
	for _, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN} {
		gopodAllFirstAnswer.Store(esn, true)
		if got := gopodExperienceClosing(esn, q); got != "" {
			t.Errorf("first answer (%s) got the move: %q", gopodBrobotName(esn), got)
		}
		gopodAllFirstAnswer.Delete(esn)
	}
	// The second answer (or a single one) redirects with a move.
	if got := gopodExperienceClosing(gopodBrobot2ESN, q); !strings.Contains(got, nap.RunName) {
		t.Errorf("second answer, not played: no move to %q: %q", nap.RunName, got)
	}
	if nap.RunName == "" || len(gopodBrobotsSettings.ExperienceMoves) != 5 {
		t.Error("no run_name, or not 5 experience_moves")
	}
	for _, m := range gopodBrobotsSettings.ExperienceMoves {
		if !strings.Contains(m, "{run}") || strings.HasPrefix(m, "Hey") {
			t.Errorf("move %q: needs {run}, no \"Hey! ...\"", m)
		}
	}
}

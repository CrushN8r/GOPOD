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
	q := "tell me about your dreams"
	if gopodIsBingoQuestion(q) || !gopodIsExperienceQuestion(nap, q) {
		t.Fatal("a dreams question is the nap's, not Bingo's")
	}
	// Pre: no run in memory here (the test feed is empty), so the pre lines match.
	if !gopodExperienceInMemory(nap) && !gopodExperienceNotPlayed("nap_not_played", q) {
		t.Error("before the nap: nap_not_played should match")
	}
	// The operator's own line doesn't count; the nap's first caption does.
	gopodExperienceMarkPlayed("Operator", nap.RunMarker)
	if flag.Load() {
		t.Error("the operator's line counted as the nap")
	}
	gopodExperienceMarkPlayed("Narrator", "Cupped in my hands,|are a pair of sleepy robots.")
	if !flag.Load() || gopodExperienceNotPlayed("nap_not_played", q) {
		t.Error("after the nap: played, and no pre lines")
	}
	// Post: his take goes with the nap's snapshot (it refers to being rescued).
	chat := gopodAddExperienceTakes(gopodBrobot1ESN, nil)
	found := false
	for _, m := range chat {
		if strings.HasPrefix(m.Content, nap.SnapshotNote) && strings.Contains(m.Content, nap.Takes["Brobot 1"]) && strings.Contains(m.Content, "rescued by Ron") {
			found = true
		}
	}
	if !found {
		t.Errorf("no nap snapshot with Brobot 1's take: %v", chat)
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
	q := "Is there anything more about dreaming you'd like to add?"
	for _, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN} {
		gopodAllFirstAnswer.Store(esn, true)
		if got := gopodExperienceClosing(esn, q); got != "" {
			t.Errorf("first answer (%s) got the move: %q", gopodBrobotName(esn), got)
		}
		gopodAllFirstAnswer.Delete(esn)
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

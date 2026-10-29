// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_docbs_test.go
package wirepod_ttr

import (
	"strings"
	"time"
	"testing"
)

// TestDocBS: at 100% Doc, Brobot 1's answer is the equation roast around his
// reply - opener, equation, error, hidden multiplier, corrected formula before
// it, "The math has spoken" after it; Brobot 2 never; a computable Explain the
// Math line keeps its steps; decks don't repeat a line before they run out.
func TestDocBS(t *testing.T) {
	saved := gopodBrobotsSettings.DocBS
	defer func() { gopodBrobotsSettings.DocBS = saved }()
	gopodGameReset("test")
	gopodGivenSet(gopodBrobot1ESN, nil)
	gopodGivenSet(gopodBrobot2ESN, nil)
	gopodBrobotsSettings.DocBS.Share = 100
	gopodBrobotsSettings.DocBS.SlotPrompt = "" // no model in the test: the fallback slots
	q := "Tell me about the rig"
	raw := "thinking___ One dolly, two crates, bro."
	t0 := time.Now()
	got := gopodShapeBrobotAnswer(gopodBrobot1ESN, q, raw)
	t.Logf("shaped in %v", time.Since(t0))
	for _, want := range []string{"Thought Quality (TQ) = (Your Confidence × Volume of the Claim) ÷ (Brain Cells Used)", "ERROR", "Corrected formula", "One dolly, two crates", "The math has spoken. And it's"} {
		if !strings.Contains(got, want) && !(want == "Corrected formula" && strings.Contains(got, "the real number")) {
			t.Errorf("DocBS answer has no %q:\n%s", want, got)
		}
	}
	if strings.Index(got, "ERROR") > strings.Index(got, "One dolly") || strings.Index(got, "One dolly") > strings.Index(got, "The math has spoken") {
		t.Errorf("DocBS beats out of order:\n%s", got)
	}
	t0 = time.Now()
	if b2 := gopodShapeBrobotAnswer(gopodBrobot2ESN, q, raw); strings.Contains(b2, "The math has spoken") {
		t.Errorf("Brobot 2 went DocBS")
	}
	t.Logf("brobot 2 shaped in %v", time.Since(t0))
	if gopodDocBSOn(gopodBrobot1ESN, "thanks Brobots") {
		t.Errorf("a thank-you went DocBS")
	}
	gopodBrobotsSettings.DocBS.Share = 0
	if !gopodDocBSOn(gopodBrobot1ESN, "Believe the sky is green all you want. It doesn't make it any less blue!! Explain the Math!!") {
		t.Errorf("an Explain the Math line the engine can't work out isn't DocBS")
	}
	if gopodDocBSOn(gopodBrobot1ESN, "What's 3/4 as a decimal? Explain the Math!") {
		t.Errorf("a computable Explain the Math line went DocBS")
	}
	t.Logf("sample:\n%s", got)
	gopodDocBSDecks.Lock()
	gopodDocBSDecks.left = map[string][]string{}
	gopodDocBSDecks.Unlock()
	seen := map[string]bool{}
	for i := 0; i < 5; i++ {
		l := gopodDocBSPick("errors")
		if seen[l] {
			t.Errorf("errors deck repeated %q before running out", l)
		}
		seen[l] = true
	}
}

// ## EOF

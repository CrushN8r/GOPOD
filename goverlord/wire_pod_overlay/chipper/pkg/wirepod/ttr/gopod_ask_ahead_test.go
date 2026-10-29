// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_ask_ahead_test.go
package wirepod_ttr

import (
	"strings"
	"testing"
	"time"
)

// TestAskAhead: Brobot 1's DocBS asks start when his prompt is built; by the
// time his answer is shaped they're back - the equation's slots used, the colour
// line just before his reply - with no ask made then.
func TestAskAhead(t *testing.T) {
	savedAsk, savedDocBS := gopodOllamaAsk, gopodBrobotsSettings.DocBS
	defer func() { gopodOllamaAsk, gopodBrobotsSettings.DocBS = savedAsk, savedDocBS }()
	gopodGameReset("test")
	gopodDocBSOnFor.Range(func(k, _ any) bool { gopodDocBSOnFor.Delete(k); return true }) // a clean slate
	gopodGivenSet(gopodBrobot1ESN, nil)
	gopodBrobotsSettings.DocBS.Share = 100
	asks := 0
	gopodOllamaAsk = func(prompt string, n int, temp float64) string {
		asks++
		time.Sleep(50 * time.Millisecond)
		if strings.Contains(prompt, "metaphor") {
			return "Your rig is a shopping cart with delusions of grandeur."
		}
		return "Cart Quotient | CQ | Wheels | Crates | Common Sense | Professor Dolly"
	}
	q := "Tell me about the rig"
	gopodPrebuildContext(gopodBrobot1ESN, q) // asks go out here
	time.Sleep(200 * time.Millisecond)
	before := asks
	got := gopodShapeBrobotAnswer(gopodBrobot1ESN, q, "thinking___ One dolly, two crates, bro.")
	if asks != before {
		t.Errorf("an ask was made while shaping (%d -> %d)", before, asks)
	}
	for _, want := range []string{"Cart Quotient (CQ) = (Wheels × Crates) ÷ (Common Sense)", "Professor Dolly", "shopping cart with delusions of grandeur"} {
		if !strings.Contains(got, want) {
			t.Errorf("no %q in:\n%s", want, got)
		}
	}
	if c, r := strings.Index(got, "shopping cart"), strings.Index(got, "One dolly"); c < 0 || r < 0 || c > r {
		t.Errorf("the colour line isn't just before his reply:\n%s", got)
	}
}

// ## EOF

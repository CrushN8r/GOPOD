// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_comm_check_test.go
package wirepod_ttr

import (
	"strings"
	"testing"
)

// TestCommCheck: the comm check asks the operator's tester once per key - KP0 to KP9 and * - each
// key reaches someone; the page draws a "held" key's bubbles.
func TestCommCheck(t *testing.T) {
	// Goverlord hands it to KEYS, KEYS calls Goverlord, answers himself, then the crews, then All.
	var keys []string
	for _, l := range gopodCommCheckLines() {
		keys = append(keys, l[0])
		if strings.HasPrefix(l[0], gopodCommCheckPersonaKey) {
			continue
		}
		if _, _, err := gopodNumpadBrobot(l[0]); err != nil {
			t.Errorf("%s: %v", l[0], err)
		}
	}
	want := "KP0 persona:GOVERLORD KP0 KP3off KP5 KP6 KP1 KP2 KP7 KP8 KP9 KP3 KP4 KP1off KP2off KPmul"
	if got := strings.Join(keys, " "); got != want {
		t.Errorf("comm check keys:\n %s\nwant\n %s", got, want)
	}
	lines := gopodCommCheckLines()
	if lines[0][2] != "GOVERLORD" || lines[1][2] != "BUTLER" || lines[2][3] != "self" || !strings.Contains(lines[0][1], "Start with me, then yourself") {
		t.Errorf("who asks: %v", lines[:3])
	}
	if crew := gopodCommCheckCrewLines(gopodBrobotsSettings.SystemBubble.VoicePersona); len(crew) < 8 || !strings.Contains(strings.Join(crew, "|"), "MATH AFTERMATH NEWS") {
		t.Errorf("crew bubble: %v", crew)
	}
	gopodCommCheckRunning.Store(true)
	defer gopodCommCheckRunning.Store(false)
	gopodCommCheckAsking.Store([2]string{"hand over", "GOVERLORD"})
	defer gopodCommCheckAsking.Store([2]string{})
	if !gopodCommCheckHandover(gopodBrobotsSettings.SystemBubble.VoicePersona, "hand over") || gopodCommCheckHandover("CHALK", "hand over") {
		t.Error("the handover is KEYS', from Goverlord")
	}
	if f := gopodBrainIntroFact(gopodBrainFacts("CHALK", gopodIntroFactAsk)); strings.Contains(f, "Math Aftermath") || !strings.Contains(f, "News Anchors") {
		t.Errorf("CHALK's comm check intro: %q", f)
	}
}

// the game welcome - only Brobot 2's, only once, the base or one of his flavours.
func TestGameWelcome(t *testing.T) {
	gw := gopodBrobotsSettings.CommCheck.GameWelcome
	if gw.By != "Brobot 2" || !strings.Contains(gw.Base, "Welcome to the GOPOD game") || len(gw.Flavours["Brobot 2"]) == 0 {
		t.Fatalf("game welcome: %+v", gw)
	}
	defer gopodGameWelcomePending.Store(false)
	gopodGameWelcomePending.Store(true)
	if got := gopodGameWelcomeLine(gopodBrobot1ESN); got != "" {
		t.Errorf("Brobot 1 got the game welcome: %q", got)
	}
	got := gopodGameWelcomeLine(gopodBrobot2ESN)
	if !strings.HasPrefix(got, "Hey, that reminds me...") || !strings.Contains(got, "GOPOD game") {
		t.Errorf("Brobot 2's game welcome: %q", got)
	}
	if again := gopodGameWelcomeLine(gopodBrobot2ESN); again != "" {
		t.Errorf("the game welcome twice: %q", again)
	}
}

// ## EOF

// after "hey all!! weather please" the comm check called KP1 "Brobot 1" - the dial left at 0% from
// the line to all): a key's name comes from the map alone (gopodNumpadKeyLabelIn), whatever the
// dial was left at.
func TestCommCheckNamesFromTheMap(t *testing.T) {
	gopodDialZero.Store(gopodBrobot1ESN, true)
	defer gopodDialZero.Delete(gopodBrobot1ESN)
	for key, want := range map[string]string{"KP1": "DOC", "KP2": "PIP", "KP1off": "BRUH-1", "KP2off": "BRUH-2"} {
		if got, ok := gopodNumpadKeyLabelIn(gopodNumpadLayout(), key); !ok || got != want {
			t.Errorf("%s: %q, want %q", key, got, want)
		}
	}
}

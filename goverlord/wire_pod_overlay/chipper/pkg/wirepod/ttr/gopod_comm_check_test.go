// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_comm_check_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCommCheck (2026-10-31, operator: "KP TEST WHY THE REDUNDANT QUESTIONS!!"):
// the comm check asks the operator's tester once per key - KP0 to KP9 and * -
// each key reaches someone; the page draws a "held" key's bubbles.
func TestCommCheck(t *testing.T) {
	lines := gopodCommCheckLines()
	keys := map[string]int{}
	for _, l := range lines {
		keys[l[0]]++
		if _, _, err := gopodNumpadBrobot(l[0]); err != nil {
			t.Errorf("%s: %v", l[0], err)
		}
	}
	for n := 0; n <= 9; n++ {
		if keys["KP"+string(rune('0'+n))] != 1 {
			t.Errorf("KP%d asked %d times, want once", n, keys["KP"+string(rune('0'+n))])
		}
	}
	// 2026-11-04 (operator: "Keys will now do the round robin"): Goverlord hands it to Keys first, then Keys asks.
	if keys[gopodNumpadAllKey] != 1 || lines[0][0] != "KP0" || lines[0][2] != "Goverlord" || lines[1][2] != "BUTLER" {
		t.Errorf("comm check lines: %v", keys)
	}
	js, _ := os.ReadFile(filepath.Join(gopodChipperRoot(), "webroot", "js", "gopod_brobots_chat.js"))
	if !strings.Contains(string(js), `indexOf("numpad_sim|")`) {
		t.Errorf("the page doesn't draw a held key's bubbles")
	}
}

// 2026-11-04 (operator: "GOPOD game = 1st reminder, however coloured/flavoured";
// "brobot 2 ends the song"): the game welcome - only Brobot 2's, only once, the
// base or one of his flavours.
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

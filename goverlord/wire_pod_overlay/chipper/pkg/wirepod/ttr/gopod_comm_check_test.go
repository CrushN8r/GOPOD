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

// 2026-11-13 (operator: "make it dynamic so it runs off the same core numpad
// mapping of all live/active keys * start with keys, then brobots"): the comm
// check asks Keys first, then the map's live keys in order, then all the Brobots.
func TestCommCheckRunsOffTheMap(t *testing.T) {
	var keys []string
	for _, l := range gopodCommCheckLines() {
		keys = append(keys, l[0])
	}
	got := strings.Join(keys, " ")
	if !strings.HasPrefix(got, "KP0 KP1 KP2 ") || !strings.HasSuffix(got, " KPmul") || !strings.Contains(got, "KP1off") || !strings.Contains(got, "KP7off") {
		t.Errorf("comm check keys: %s", got)
	}
	seen := map[string]bool{}
	for _, k := range keys {
		if seen[k] {
			t.Errorf("%s asked twice", k)
		}
		seen[k] = true
	}
}

// 2026-11-14 (live: after "hey all!! weather please" the comm check called KP1
// "Brobot 1" - the dial left at 0% from the line to all): a key's name comes
// from the map alone (gopodNumpadKeyLabelIn), whatever the dial was left at.
func TestCommCheckNamesFromTheMap(t *testing.T) {
	gopodDialZero.Store(gopodBrobot1ESN, true)
	defer gopodDialZero.Delete(gopodBrobot1ESN)
	for key, want := range map[string]string{"KP1": "DOC", "KP2": "PIP", "KP1off": "Brobot 1", "KP2off": "Brobot 2"} {
		if got, ok := gopodNumpadKeyLabelIn(gopodNumpadLayout(), key); !ok || got != want {
			t.Errorf("%s: %q, want %q", key, got, want)
		}
	}
}

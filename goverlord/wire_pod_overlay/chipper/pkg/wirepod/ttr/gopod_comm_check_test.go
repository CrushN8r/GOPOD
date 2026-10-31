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
	if keys[gopodNumpadAllKey] != 1 || !strings.Contains(lines[0][1], "party animal") {
		t.Errorf("comm check lines: %v", keys)
	}
	js, _ := os.ReadFile(filepath.Join(gopodChipperRoot(), "webroot", "js", "gopod_brobots_chat.js"))
	if !strings.Contains(string(js), `indexOf("numpad_sim|")`) {
		t.Errorf("the page doesn't draw a held key's bubbles")
	}
}

// ## EOF

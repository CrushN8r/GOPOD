// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_comm_check_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCommCheck: the comm check holds every key KP1-KP9 for each round, plus
// the interns question for CHALK and PLAYHEAD; each key reaches someone; the
// page draws a "held" key's bubbles.
func TestCommCheck(t *testing.T) {
	lines := gopodCommCheckLines()
	keys := map[string]int{}
	for _, l := range lines {
		keys[l[0]]++
		if _, _, err := gopodNumpadBrobot(l[0]); err != nil {
			t.Errorf("%s: %v", l[0], err)
		}
	}
	for n := 1; n <= 9; n++ {
		if keys["KP"+string(rune('0'+n))] < 4 {
			t.Errorf("KP%d asked %d times", n, keys["KP"+string(rune('0'+n))])
		}
	}
	if keys["KP7"] != 5 || keys["KP8"] != 5 || !strings.Contains(lines[0][1], "party animal") {
		t.Errorf("comm check lines: %v", keys)
	}
	js, _ := os.ReadFile(filepath.Join(gopodChipperRoot(), "webroot", "js", "gopod_brobots_chat.js"))
	if !strings.Contains(string(js), `indexOf("numpad_sim|")`) {
		t.Errorf("the page doesn't draw a held key's bubbles")
	}
}

// ## EOF

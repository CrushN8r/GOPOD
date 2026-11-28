// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_docbs_library_test.go
package wirepod_ttr

import (
	"strings"
	"testing"
)

// TestDocBSLibraryChainsPlay: every DocBS chain (chains.md) plays - each beat has a line, a chain's
// own lines ("library/NN_name#2") are there in order, and the library's 50 are on the menu at the
// dial's 100%.
func TestDocBSLibraryChainsPlay(t *testing.T) {
	chains := gopodBrobotsSettings.DocBS.Chains
	if len(chains) == 0 {
		t.Skip("no DocBS chains here")
	}
	at100 := 0
	for _, c := range chains {
		if c.MinShare <= 100 {
			at100++
		}
		for _, beat := range c.Beats {
			for _, pool := range strings.Split(beat, "|") {
				pool = strings.TrimSpace(pool)
				if pool == "colour" {
					continue
				}
				line := ""
				if at := strings.LastIndex(pool, "#"); at > 0 {
					line = gopodDocBSLineOf(pool[:at], pool[at+1:])
				} else {
					line = gopodDocBSPick(pool)
				}
				if strings.TrimSpace(line) == "" {
					t.Errorf("chain %q: beat %q has no line", c.Name, pool)
				}
			}
		}
	}
	if at100 < 55 {
		t.Errorf("at the dial's 100%%: %d chains on the menu, want the 5 below 11/10 plus the library's 50", at100)
	}
}

// ## EOF

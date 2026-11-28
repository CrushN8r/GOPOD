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

// TestDizzyAsk: asked if he's dizzy, a Brobot gives his power level and how he is - good, full, or
// his dizzy line with help getting home off the charger.
func TestDizzyAsk(t *testing.T) {
	bw := gopodBrobotsSettings.BatteryWatch
	if strings.TrimSpace(bw.AskLine) == "" {
		t.Skip("no ask_line here")
	}
	if !gopodDizzyAsk("Brobot 1, are you dizzy?") || gopodDizzyAsk("Brobot 1, tell me about the gears") {
		t.Fatal("ask words: 'are you dizzy' should ask, the gears shouldn't")
	}
	esn := gopodBrobot1ESN
	gopodBatteryLow.Lock()
	kept := [4]any{gopodBatteryLow.prev[esn], gopodBatteryLow.said[esn], gopodBatteryLow.full[esn], gopodBatteryLow.home[esn]}
	gopodBatteryLow.Unlock()
	defer func() {
		gopodBatteryLow.Lock()
		gopodBatteryLow.prev[esn], gopodBatteryLow.said[esn], gopodBatteryLow.full[esn], gopodBatteryLow.home[esn] = kept[0].(float64), kept[1].(bool), kept[2].(bool), kept[3].(bool)
		gopodBatteryLow.Unlock()
	}()
	set := func(volts float64, said, full, home bool) string {
		gopodBatteryLow.Lock()
		gopodBatteryLow.prev[esn], gopodBatteryLow.said[esn], gopodBatteryLow.full[esn], gopodBatteryLow.home[esn] = volts, said, full, home
		gopodBatteryLow.Unlock()
		return gopodDizzyLine(esn)
	}
	if l := set(3.95, false, false, false); !strings.Contains(l, "3.95") || !strings.Contains(l, strings.TrimSpace(bw.AskStates["good"])) {
		t.Errorf("good: %q", l)
	}
	if l := set(3.61, false, true, true); !strings.Contains(l, "3.61") || !strings.Contains(l, strings.TrimSpace(bw.AskStates["full"])) {
		t.Errorf("full on the charger: %q", l)
	}
	home := strings.TrimSpace(bw.HomeLines["Brobot 1"])
	if l := set(bw.LowVolts-0.05, false, false, false); !strings.Contains(strings.ToLower(l), "dizzy") || (home != "" && !strings.Contains(l, home)) {
		t.Errorf("low, off the charger: %q", l)
	}
	if l := set(bw.LowVolts-0.05, false, false, true); home != "" && strings.Contains(l, home) {
		t.Errorf("low on the charger needs no help home: %q", l)
	}
	if l := set(0, false, false, false); strings.Contains(l, "volts") {
		t.Errorf("not read yet: %q", l)
	}
}

// ## EOF

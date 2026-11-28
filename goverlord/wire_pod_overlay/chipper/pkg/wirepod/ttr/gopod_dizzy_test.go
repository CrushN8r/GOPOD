// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_dizzy_test.go
package wirepod_ttr

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// TestDizzyAsk: "Are you (feeling) dizzy?" - one call for everyone. A Brobot gives his power level
// and how he is (good, full, his dizzy line with help home off the charger); his guard gives his
// sweep words; anyone else gets no line, only the note to hand it to the guards.
func TestDizzyAsk(t *testing.T) {
	bw := gopodBrobotsSettings.BatteryWatch
	if strings.TrimSpace(bw.AskLine) == "" {
		t.Skip("no ask_line here")
	}
	if !gopodDizzyAsk("Brobot 1, are you dizzy?") || !gopodDizzyAsk("Are you feeling dizzy?") || gopodDizzyAsk("Brobot 1, tell me about the gears") {
		t.Fatal("the catchphrase: 'are you (feeling) dizzy' asks, the gears don't")
	}
	esn := gopodBrobot1ESN
	gopodBatteryLow.Lock()
	kept := [4]any{gopodBatteryLow.prev[esn], gopodBatteryLow.said[esn], gopodBatteryLow.full[esn], gopodBatteryLow.home[esn]}
	gopodBatteryLow.Unlock()
	gopodBatteryLast.Lock()
	keptText, hadText := gopodBatteryLast.text[esn]
	gopodBatteryLast.text[esn] = "3.90 V, NOMINAL"
	gopodBatteryLast.Unlock()
	defer func() {
		gopodBatteryLow.Lock()
		gopodBatteryLow.prev[esn], gopodBatteryLow.said[esn], gopodBatteryLow.full[esn], gopodBatteryLow.home[esn] = kept[0].(float64), kept[1].(bool), kept[2].(bool), kept[3].(bool)
		gopodBatteryLow.Unlock()
		gopodBatteryLast.Lock()
		if hadText {
			gopodBatteryLast.text[esn] = keptText
		} else {
			delete(gopodBatteryLast.text, esn)
		}
		gopodBatteryLast.Unlock()
	}()
	n := 0
	ask := func(who string, volts float64, said, full, home bool) (string, string) {
		gopodBatteryLow.Lock()
		gopodBatteryLow.prev[esn], gopodBatteryLow.said[esn], gopodBatteryLow.full[esn], gopodBatteryLow.home[esn] = volts, said, full, home
		gopodBatteryLow.Unlock()
		n++
		return gopodDizzy(who, strings.Repeat(" ", n)+"Are you dizzy?") // a fresh question each time
	}
	if l, _ := ask("Brobot 1", 3.95, false, false, false); !strings.Contains(l, "3.95") || !strings.Contains(l, strings.TrimSpace(bw.AskStates["good"])) {
		t.Errorf("good: %q", l)
	}
	if l, _ := ask("Brobot 1", 3.61, false, true, true); !strings.Contains(l, "3.61") || !strings.Contains(l, strings.TrimSpace(bw.AskStates["full"])) {
		t.Errorf("full on the charger: %q", l)
	}
	home := strings.TrimSpace(bw.HomeLines["Brobot 1"])
	if l, _ := ask("Brobot 1", bw.LowVolts-0.05, false, false, false); !strings.Contains(strings.ToLower(l), "dizzy") || (home != "" && !strings.Contains(l, home)) {
		t.Errorf("low, off the charger: %q", l)
	}
	if l, _ := ask("Brobot 1", bw.LowVolts-0.05, false, false, true); home != "" && strings.Contains(l, home) {
		t.Errorf("low on the charger needs no help home: %q", l)
	}
	if guard := bw.Guards["Brobot 1"]; guard != "" {
		if l, note := ask(guard, 3.95, false, false, false); !strings.Contains(l, "3.95") || !strings.Contains(note, l) {
			t.Errorf("his guard reports him in sweep words: %q / %q", l, note)
		}
	}
	if l, note := ask("GEMMA", 3.95, false, false, false); l != "" || note == "" {
		t.Errorf("anyone else: no line, the note to hand it to the guards: %q / %q", l, note)
	}
	q := "Brobot 1, are you feeling dizzy?"
	l1, _ := gopodDizzy("Brobot 1", q)
	gopodBatteryLow.Lock()
	gopodBatteryLow.prev[esn] = 3.10
	gopodBatteryLow.Unlock()
	if l2, _ := gopodDizzy("Brobot 1", q); l1 != l2 {
		t.Errorf("one answer, one reading: %q then %q", l1, l2)
	}
}

// TestDuoDropsAnUnreachableBrobot: a line to both with one Brobot off the network - the one who can
// be reached answers on his own.
func TestDuoDropsAnUnreachableBrobot(t *testing.T) {
	gopodNetwork.Lock()
	kept := gopodNetwork.downAt
	gopodNetwork.downAt = map[string]time.Time{}
	gopodNetwork.Unlock()
	defer func() { gopodNetwork.Lock(); gopodNetwork.downAt = kept; gopodNetwork.Unlock() }()
	if a, b, alone := gopodAllPair(gopodBrobot1ESN); a != gopodBrobot1ESN || b != gopodBrobot2ESN || alone {
		t.Errorf("both up: Brobot 1 then 2, a duo - got %s %s %v", a, b, alone)
	}
	gopodNetworkDown(gopodBrobot1ESN, errors.New("rpc error: code = Unavailable desc = connection error"))
	if a, _, alone := gopodAllPair(gopodBrobot1ESN); a != gopodBrobot2ESN || !alone {
		t.Errorf("Brobot 1 down: Brobot 2 alone - got %s %v", a, alone)
	}
	gopodNetwork.Lock()
	gopodNetwork.downAt = map[string]time.Time{}
	gopodNetwork.Unlock()
	gopodNetworkDown(gopodBrobot2ESN, errors.New("rpc error: connection error"))
	if a, _, alone := gopodAllPair(gopodBrobot1ESN); a != gopodBrobot1ESN || !alone {
		t.Errorf("Brobot 2 down: Brobot 1 alone - got %s %v", a, alone)
	}
	gopodNetworkDown(gopodBrobot1ESN, errors.New("the language model said nothing"))
	if !gopodNetworkUp(gopodBrobot1ESN) {
		t.Error("only a network error takes a Brobot down")
	}
}

// ## EOF

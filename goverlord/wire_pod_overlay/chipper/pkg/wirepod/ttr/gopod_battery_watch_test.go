// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_battery_watch_test.go
package wirepod_ttr

import (
	"strings"
	"testing"
	"time"
)

// The battery watchdog says its line once per drop - below low_volts, or a dip
// of dip_volts since the last reading - off the charger or on it not charging
// (a bad contact), and again only after charging or recover_volts.
func TestBatteryWatchOncePerDrop(t *testing.T) {
	bw := gopodBrobotsSettings.BatteryWatch
	low := bw.LowVolts - 0.05
	said, prev := false, 0.0
	steps := []struct {
		volts                   float64
		charging, charger, full bool
		want                    bool
	}{
		{4.05, false, false, false, false}, // full
		{low, false, false, false, true},   // dropped: says it
		{low - 0.02, false, false, false, false},
		{bw.LowVolts + 0.02, false, false, false, false}, // back over the line, under recover: no repeat
		{low, false, false, false, false},
		{low, true, true, false, false},  // on his charger, charging: re-armed
		{low, false, false, false, true}, // off again, still low: says it again
		{low, true, true, false, false},  // charging: re-armed
		{low, false, true, false, true},  // on the charger, not charging, low: a bad contact - says it
		{4.05, true, true, false, false}, // charged up
		{4.05, false, false, false, false},
		{4.05 - bw.DipVolts - 0.01, false, false, false, bw.DipVolts > 0}, // a dip, above low_volts
		{4.05, true, true, false, false},
		{3.58, false, true, true, false},       // 1.6-rebuild: resting full on his charger, not charging - fine
		{low - 0.1, false, false, false, true}, // off the charger and low: says it
	}
	for i, s := range steps {
		var say bool
		say, said = gopodBatteryDecide(said, prev, s.volts, s.charging, s.charger, s.full)
		prev = s.volts
		if say != s.want {
			t.Errorf("step %d (%.2f V charging=%v charger=%v): said=%v, want %v", i, s.volts, s.charging, s.charger, say, s.want)
		}
	}
}

// 2026-11-03 (operator: "after freeze/reboot response ... Then, Bam!";
// dizzy-brobots.md): he's back after back_after_misses failed readings in a row,
// only when he was read before; and the lines come from the cluster.
func TestBatteryBackAfterFreeze(t *testing.T) {
	bw := &gopodBrobotsSettings.BatteryWatch
	old := bw.BackAfterMisses
	defer func() { bw.BackAfterMisses = old }()
	bw.BackAfterMisses = 2
	esn := "ESN01"
	if gopodBatteryBack(esn, false) || gopodBatteryBack(esn, false) || gopodBatteryBack(esn, true) {
		t.Error("back before he was ever read (Wire-Pod start, robot asleep)")
	}
	if gopodBatteryBack(esn, false) || gopodBatteryBack(esn, true) {
		t.Error("one missed reading is not a freeze")
	}
	gopodBatteryBack(esn, false)
	gopodBatteryBack(esn, false)
	if !gopodBatteryBack(esn, true) {
		t.Error("not back after two missed readings")
	}
	if gopodBatteryBack(esn, true) {
		t.Error("back twice for one freeze")
	}
	if len(bw.Lines["Brobot 1"]) == 0 || bw.HomeLines["Brobot 1"] == "" || !strings.Contains(bw.BackLines["Brobot 1"], "Then, Bam!") {
		t.Errorf("dizzy-brobots.md not laid over the settings: %q %q %q", bw.Lines, bw.HomeLines, bw.BackLines)
	}
}

// 2026-11-19 (operator: "security sweep 1 = BUZZER sweeps brobot 1 detection &
// sufficient battery power to proceed * security sweep 2 = SWITCH"): a missing or
// low Brobot holds the run; both found and powered, it goes.
func TestSecuritySweepHoldsTheRun(t *testing.T) {
	low := gopodBrobotsSettings.BatteryWatch.LowVolts
	t.Cleanup(func() {
		gopodBatteryLast.Lock()
		gopodBatteryLast.text, gopodBatteryLast.at = map[string]string{}, map[string]time.Time{}
		gopodBatteryLast.Unlock()
	})
	set := func(esn, text string, volts float64) {
		gopodBatteryRemember(esn, text)
		gopodBatteryLow.Lock()
		gopodBatteryLow.prev[esn] = volts
		gopodBatteryLow.Unlock()
	}
	set(gopodBrobot1ESN, "unknown: off", 0)
	set(gopodBrobot2ESN, "4.0 V", low+0.3)
	if gopodSecuritySweep() {
		t.Error("Brobot 1 missing, the run went")
	}
	set(gopodBrobot1ESN, "4.0 V", low-0.1)
	if gopodSecuritySweep() {
		t.Error("Brobot 1 low, the run went")
	}
	set(gopodBrobot1ESN, "4.0 V", low+0.3)
	if !gopodSecuritySweep() {
		t.Error("both found and powered, the run was held")
	}
	// 2026-11-27: 3.58 V resting FULL on his charger (1.6-rebuild) is powered.
	set(gopodBrobot1ESN, "3.58 V, FULL, on the charger", 3.58)
	gopodBatteryLow.Lock()
	gopodBatteryLow.full[gopodBrobot1ESN] = true
	gopodBatteryLow.Unlock()
	t.Cleanup(func() {
		gopodBatteryLow.Lock()
		delete(gopodBatteryLow.full, gopodBrobot1ESN)
		gopodBatteryLow.Unlock()
	})
	if !gopodSecuritySweep() {
		t.Error("full on his charger at 3.58 V, the run was held")
	}
	bw := gopodBrobotsSettings.BatteryWatch
	if bw.SweepFull == "" || !strings.Contains(bw.SweepStop, "{who}") || !strings.Contains(bw.GuardWarning, "{volts}") {
		t.Errorf("dizzy-brobots.md sweep_full/sweep_stop/guard_warning not laid over: %q %q %q", bw.SweepFull, bw.SweepStop, bw.GuardWarning)
	}
}

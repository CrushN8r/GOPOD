// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_battery_watch_test.go
package wirepod_ttr

import "testing"

// The battery watchdog says its line once per drop - below low_volts, or a dip
// of dip_volts since the last reading - off the charger or on it not charging
// (a bad contact), and again only after charging or recover_volts.
func TestBatteryWatchOncePerDrop(t *testing.T) {
	bw := gopodBrobotsSettings.BatteryWatch
	low := bw.LowVolts - 0.05
	said, prev := false, 0.0
	steps := []struct {
		volts             float64
		charging, charger bool
		want              bool
	}{
		{4.05, false, false, false}, // full
		{low, false, false, true},   // dropped: says it
		{low - 0.02, false, false, false},
		{bw.LowVolts + 0.02, false, false, false}, // back over the line, under recover: no repeat
		{low, false, false, false},
		{low, true, true, false},  // on his charger, charging: re-armed
		{low, false, false, true}, // off again, still low: says it again
		{low, true, true, false},  // charging: re-armed
		{low, false, true, true},  // on the charger, not charging, low: a bad contact - says it
		{4.05, true, true, false}, // charged up
		{4.05, false, false, false},
		{4.05 - bw.DipVolts - 0.01, false, false, bw.DipVolts > 0}, // a dip, above low_volts
	}
	for i, s := range steps {
		var say bool
		say, said = gopodBatteryDecide(said, prev, s.volts, s.charging, s.charger)
		prev = s.volts
		if say != s.want {
			t.Errorf("step %d (%.2f V charging=%v charger=%v): said=%v, want %v", i, s.volts, s.charging, s.charger, say, s.want)
		}
	}
}

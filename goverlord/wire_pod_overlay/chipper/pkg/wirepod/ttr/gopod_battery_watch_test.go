// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_battery_watch_test.go
package wirepod_ttr

import "testing"

// The battery watchdog says its line once per drop below low_volts, off the
// charger, and again only after recover_volts or the charger.
func TestBatteryWatchOncePerDrop(t *testing.T) {
	bw := gopodBrobotsSettings.BatteryWatch
	low := bw.LowVolts - 0.05
	said := false
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
		{low, false, true, false}, // set on his charger: re-armed
		{low, false, false, true}, // off again, still low: says it again
	}
	for i, s := range steps {
		var say bool
		say, said = gopodBatteryDecide(said, s.volts, s.charging, s.charger)
		if say != s.want {
			t.Errorf("step %d (%.2f V charger=%v): said=%v, want %v", i, s.volts, s.charger, say, s.want)
		}
	}
}

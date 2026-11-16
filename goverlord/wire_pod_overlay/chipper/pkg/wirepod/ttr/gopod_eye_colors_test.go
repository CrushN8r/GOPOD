// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_eye_colors_test.go
package wirepod_ttr

import "testing"

// TestEyeColorHex: the same colour Python's colorsys gave the interview runner.
func TestEyeColorHex(t *testing.T) {
	for _, c := range []struct {
		h, s float64
		want string
	}{{0, 1, "#FF0000"}, {2.0 / 3, 1, "#0000FF"}, {0.5, 0.5, "#80FFFF"}, {0, 0, "#FFFFFF"}} {
		if got := gopodHSVHex(c.h, c.s, 1); got != c.want {
			t.Errorf("hsv(%v, %v) = %s, want %s", c.h, c.s, got, c.want)
		}
	}
}

// ## EOF

// 2026-11-15: a card's border is the robot's eye colour - hex to hue and
// saturation and back; DOC's red while he answers, Brobot 1's gold after.
func TestBrobotEyesFromCards(t *testing.T) {
	for _, hex := range []string{"#FF3B30", "#42A5F5", "#D4AF37"} {
		h, s, ok := gopodHexHS(hex)
		if !ok || gopodHSVHex(h, s, 1) == "" {
			t.Errorf("%s: %v %v %v", hex, h, s, ok)
		}
	}
	if h, s, _ := gopodHexHS("#0000FF"); gopodHSVHex(h, s, 1) != "#0000FF" {
		t.Errorf("blue round trip: %v %v", h, s)
	}
	own, hosted := gopodBrobotCardColours("Brobot 1")
	if gopodCardColour(own, "border") != "#d4af37" || gopodCardColour(hosted, "border") != "#ff3b30" {
		t.Errorf("Brobot 1's cards: %q / %q", own, hosted)
	}
	if !gopodAtDialZero("Brobot 1") && gopodBrobotEyeHex("Brobot 1", true) != "#ff3b30" {
		t.Errorf("DOC answering: %s", gopodBrobotEyeHex("Brobot 1", true))
	}
	if gopodBrobotEyeHex("Brobot 1", false) != "#d4af37" {
		t.Errorf("Brobot 1 at rest: %s", gopodBrobotEyeHex("Brobot 1", false))
	}
}

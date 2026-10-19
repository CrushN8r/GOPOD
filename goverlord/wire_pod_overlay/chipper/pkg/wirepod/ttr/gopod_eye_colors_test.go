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

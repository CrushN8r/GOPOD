// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_torah_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestTorahMatches: GOPOD_TORAH.md is the system as it is now (GOPOD_WRITE_TORAH=1
// writes it again); every system window command is in it.
func TestTorahMatches(t *testing.T) {
	if _, err := os.Stat(filepath.Join(gopodRepoRoot(), "README.md")); err != nil {
		t.Skip("not in the repo here")
	}
	page := strings.Join(func() []string {
		var all []string
		for _, v := range gopodTorahFiles() {
			all = append(all, v)
		}
		return all
	}(), "\n")
	for _, want := range []string{"# The GOPOD Torah", "system-test-admin-run", "system-comm-check", "start bingo test", "torah"} {
		if !strings.Contains(page, want) {
			t.Errorf("the Torah has no %q", want)
		}
	}
	// The tour: every scripture opens "Scripture N, ..."; every commandment it
	// runs is one of the Torah's, and never the restart, a robot move, a switch,
	// or the session's end.
	all := map[string]bool{}
	for i, sc := range gopodTorahScriptures() {
		if !strings.HasPrefix(sc.Title, "Scripture "+strconv.Itoa(i+1)+",") {
			t.Errorf("scripture %d: %q", i+1, sc.Title)
		}
		for _, c := range sc.Cmds {
			all[c] = true
		}
	}
	for _, c := range gopodBrobotsSettings.SystemTour.Run {
		if !all[c] {
			t.Errorf("the tour runs %q - not a commandment in the Torah", c)
		}
		for _, never := range []string{"system-restart", "system-brobot wake", "system-brobot head", "system-brobot lift", "system-brobot wheels", "system-calibrate", "stop session", "thanks brobots", "system-numpad", "system-native", "system-tweaks", "system-tpu restart", "system-comm-check"} {
			if strings.HasPrefix(c, never) {
				t.Errorf("the tour must not run %q", c)
			}
		}
	}
	if os.Getenv("GOPOD_WRITE_TORAH") == "1" {
		if err := gopodWriteTorah(); err != nil {
			t.Fatal(err)
		}
		return
	}
	if gopodTorahBehind() {
		t.Errorf("GOPOD_TORAH.md or a scripture (torah/) doesn't match the system - GOPOD_WRITE_TORAH=1 (or system-torah) writes it again")
	}
}

// ## EOF

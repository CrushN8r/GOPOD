// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_robot_fx_test.go
package wirepod_ttr

import (
	"strings"
	"testing"
)

// TestRobotFX: the guards, the interns and the Brobots' demo voices each get
// their robot effect with their own carrier (slightly offset) and their own
// files; the demo voices speak with PLAYHEAD's voice; the live Brobots get none.
func TestRobotFX(t *testing.T) {
	rv := gopodBrobotsSettings.ReporterVoices
	seen := map[string]string{}
	for _, h := range []string{"INTEL 404 PNFH", "INTEL 404 PNFT", "GIT REPO", "CACHE PYC", "Brobot 1 demo", "Brobot 2 demo"} {
		chain, wet, tag := gopodReporterFX(h)
		if chain == "" || wet <= 0 || strings.Contains(chain, "{carrier}") || tag == "" {
			t.Errorf("%s: no robot effect (%q %v %q)", h, chain, wet, tag)
		}
		if other, dup := seen[tag]; dup {
			t.Errorf("%s and %s share an effect - not offset", h, other)
		}
		seen[tag] = h
		if _, page := gopodReporterPaths(h, "Kernel panic!"); !strings.Contains(page, tag) {
			t.Errorf("%s: its file name has no effect tag: %s", h, page)
		}
	}
	for _, b := range []string{"Brobot 1 demo", "Brobot 2 demo"} {
		if gopodReporterVoice(b) != rv.Hosts["PLAYHEAD"] {
			t.Errorf("%s doesn't speak with PLAYHEAD's voice", b)
		}
	}
	if gopodReporterVoice("Brobot 1") != "" || gopodReporterHost("Brobot 1") != "" {
		t.Errorf("the live Brobot 1 took a reporter voice")
	}
}

// ## EOF

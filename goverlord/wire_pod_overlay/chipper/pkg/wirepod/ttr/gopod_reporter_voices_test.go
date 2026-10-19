// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_reporter_voices_test.go
package wirepod_ttr

import (
	"strings"
	"testing"
)

// Every voice and speed has its own files (operator 2026-10-09: "ensure there's
// a distinction so I will have wav of all generated voices without over writing").
func TestReporterFilesNamedByVoiceAndSpeed(t *testing.T) {
	saved := gopodBrobotsSettings.ReporterVoices
	defer func() { gopodBrobotsSettings.ReporterVoices = saved }()
	gopodBrobotsSettings.ReporterVoices.Hosts = map[string]string{"Brobot 3": "af_heart", "Brobot 4": "am_michael"}
	gopodBrobotsSettings.ReporterVoices.Speed = 1.03
	_, oldPage := gopodReporterPaths("Brobot 3", "Bro, can you hear me?")
	gopodBrobotsSettings.ReporterVoices.Speed = 0.98
	_, page3 := gopodReporterPaths("Brobot 3", "Bro, can you hear me?")
	_, page4 := gopodReporterPaths("Brobot 4", "Bro, can you hear me?")
	if !strings.Contains(page3, "/af_heart_s098_") || !strings.Contains(page4, "/am_michael_s098_") {
		t.Fatalf("names %q %q, want the voice and speed in them", page3, page4)
	}
	if page3 == oldPage || strings.Contains(oldPage, "_s") {
		t.Fatalf("speed 0.98 %q must differ from 1.03 %q, and 1.03 keeps the old name", page3, oldPage)
	}
}

// ## EOF

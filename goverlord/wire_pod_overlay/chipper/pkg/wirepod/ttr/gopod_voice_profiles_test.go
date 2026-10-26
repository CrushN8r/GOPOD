// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_voice_profiles_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 2026-10-26 (operator: "per-audio-file-profile-tag ... generation date, # of
// times used, time last used"; "popular repeating live questions, 'regenerate'
// ... takes over the per-audio-file-profile-tag, now with a # of
// regenerations"; "add a report-run ... past 3-days/week/month/year").
func TestVoiceProfilesUseTakeOverReport(t *testing.T) {
	root := t.TempDir()
	t.Setenv(gopodChipperRootEnvVar, root)
	t.Setenv("GOPOD_SESSION_DIR", "")
	dir := filepath.Join(root, "webroot", "audio")
	os.MkdirAll(dir, 0o755)
	old := gopodBrobotsSettings
	defer func() { gopodBrobotsSettings = old }()
	gopodBrobotsSettings.ReporterVoices.AudioFolder = "audio"
	p := &gopodVoiceProfiles
	p.Lock()
	saved := *p
	p.loaded, p.files, p.since = true, map[string]*gopodVoiceProfile{}, "2026-10-26T00:00:00Z"
	p.Unlock()
	defer func() { p.Lock(); p.loaded, p.files, p.since = saved.loaded, saved.files, saved.since; p.Unlock() }()
	if gopodReporterAudioDir() != dir || !strings.HasPrefix(gopodArchiveDir(), root) {
		t.Skip("the audio or archive folder can't be pointed at a test folder here")
	}
	first, second := "op_mic_00000000000000aa.wav", "op_mic_00000000000000bb.wav"
	os.WriteFile(filepath.Join(dir, first), []byte("RIFF"), 0o644)
	os.WriteFile(filepath.Join(dir, second), []byte("RIFF"), 0o644)
	gopodVoiceProfileMade(first, "mic", "CHALK", "", "Are we all in agreement")
	gopodVoiceUsed(first)
	gopodVoiceUsed(first)
	gopodVoiceProfileMade(second, "mic", "CHALK", "", "are we all in  agreement")
	gopodVoiceProfileTakeOver(second, "are we all in  agreement")
	p.Lock()
	prof, gone := p.files[second], p.files[first] == nil
	p.Unlock()
	if prof == nil || !gone || prof.Uses != 2 || prof.Regenerations != 1 || prof.LastUsed == "" {
		t.Fatalf("take over: %+v, old gone %v", prof, gone)
	}
	if _, err := os.Stat(filepath.Join(dir, first)); err == nil {
		t.Error("the older recording wasn't archived")
	}
	report := gopodSystemVoicesReport("system-voices-report 3d")
	if !strings.Contains(report, "past 3d") || !strings.Contains(report, "2x  are we all in  agreement") {
		t.Errorf("report:\n%s", report)
	}
	if !strings.Contains(gopodSystemVoicesReport("system-voices-report decade"), "[3d | week | month | year]") {
		t.Error("an unknown window should show the choices")
	}
}

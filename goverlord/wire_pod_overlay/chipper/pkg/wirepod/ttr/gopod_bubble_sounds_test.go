package wirepod_ttr

import (
	"strings"
	"testing"
)

// 2026-10-25 (operator: "remove grunge audio * default should = djent, but
// muted"): a Goverlord line - a run's started/complete line too - plays no sound
// of its own while bubble_sounds.Goverlord is muted; a "(music)" still plays djent.
func TestGoverlordSoundMutedMusicStays(t *testing.T) {
	if s, _ := gopodSystemSting("Fart joke started. Right away, sir."); s != "" {
		t.Fatalf("a muted Goverlord line plays %q", s)
	}
	m, _ := gopodMusicSting()
	if m != "" && !strings.Contains(m, "dJent3b.wav") {
		t.Fatalf("(music) plays %q", m)
	}
	if _, ok := gopodBrobotsSettings.ReporterVoices.SystemStings["runs"]; ok {
		t.Fatal("the grunge run sting is back")
	}
}

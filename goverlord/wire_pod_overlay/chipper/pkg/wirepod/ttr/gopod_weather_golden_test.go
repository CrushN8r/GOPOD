package wirepod_ttr

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// LOCKED 2026-11-03 (operator: "Look at the tampa run. that is a near perfect
// example of a golden locked run * golden guts scaffolding = topic-cluster md";
// "Lock order + code lines"; "Doc then Pip (Tampa)"): the golden weather run is
// gopod_probes/clusters/weather_run.md and only that. The deploy fails when it
// changed without the operator's relock (GOPOD_RELOCK=1 saves its fingerprint,
// gopod_probes/locks/weather_golden.lock), when its order or lines aren't the
// ones in use, or when a settings file holds a copy. The order itself - Doc's
// answer and Pip's - is run through the real answer chain by TestAnswerShapesHold.
func TestWeatherGoldenLock(t *testing.T) {
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("GOPOD_OVERLAY_CHIPPER not set (the deploy sets it)")
	}
	note, err := os.ReadFile(filepath.Join(root, "gopod_probes", "clusters", "weather_run.md"))
	if err != nil {
		t.Fatalf("LOCKED: the golden weather run is missing: %v", err)
	}
	sum := sha256.Sum256(note)
	got := hex.EncodeToString(sum[:])
	lock := filepath.Join(root, "gopod_probes", "locks", "weather_golden.lock")
	if os.Getenv("GOPOD_RELOCK") == "1" {
		if err := os.WriteFile(lock, []byte(got+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	} else if want, err := os.ReadFile(lock); err != nil {
		t.Errorf("LOCKED: the weather lock is missing: %v", err)
	} else if strings.TrimSpace(string(want)) != got {
		t.Errorf("LOCKED: the golden weather run changed (operator's word only, then GOPOD_RELOCK=1)")
	}
	s := gopodBrobotsSettings
	if strings.Join(s.AnswerShapes.Shapes["weather"], ",") != "weather_report,llm?,point?,llm?,reminder" ||
		strings.Join(s.AnswerShapes.Shapes["weather_next"], ",") != "weather_confirm,llm?,point?,llm?,weather_kit,reminder" {
		t.Errorf("LOCKED: the weather order in use isn't the golden one: %v / %v", s.AnswerShapes.Shapes["weather"], s.AnswerShapes.Shapes["weather_next"])
	}
	if !strings.Contains(s.LiveWeather.FollowUpKitLine, "[GOPOD Weather Kit](/gopod_weather_kit.html)") {
		t.Errorf("LOCKED: the Weather Kit line in use isn't the golden one: %q", s.LiveWeather.FollowUpKitLine)
	}
	// One source: the settings files hold no copy.
	data, _ := gopodSettingsMerged()
	for _, copy := range []string{"Brobot Weather Report", "Confirmed. Temperature", "\"follow_up_kit_line\":", "\"weather_lead\":", "\"weather_point_first\":", "\"weather_next\": ["} {
		if strings.Contains(string(data), copy) {
			t.Errorf("LOCKED: a settings file holds a copy of the golden weather run: %s", copy)
		}
	}
}

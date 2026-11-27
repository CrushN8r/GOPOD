package wirepod_ttr

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// LOCKED the golden weather run is gopod_probes/topic_clusters/personas/behaviours/weather_run.md
// and only that.
func TestWeatherGoldenLock(t *testing.T) {
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("GOPOD_OVERLAY_CHIPPER not set (the deploy sets it)")
	}
	note, err := os.ReadFile(gopodTopicTreeFind("weather_run.md"))
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
	for _, kind := range []string{"weather", "weather_next"} {
		want := ""
		for _, l := range strings.Split(string(note), "\n") {
			if _, rest, ok := strings.Cut(strings.TrimSpace(l), ". "+kind+": "); ok && strings.TrimLeft(strings.SplitN(strings.TrimSpace(l), ".", 2)[0], "0123456789") == "" {
				want = strings.ReplaceAll(rest, " | ", ",")
			}
		}
		if want == "" || strings.Join(s.AnswerShapes.Shapes[kind], ",") != want {
			t.Errorf("LOCKED: the %s order in use isn't the card's: %v, card %q", kind, s.AnswerShapes.Shapes[kind], want)
		}
	}
	if !strings.Contains(s.LiveWeather.FollowUpKitLine, "[GOPOD Weather Kit](/gopod_weather_kit.html)") {
		t.Errorf("LOCKED: the Weather Kit line in use isn't the golden one: %q", s.LiveWeather.FollowUpKitLine)
	}
	data, _ := gopodSettingsMerged()
	for _, copy := range []string{"Brobot Weather Report", "Confirmed. Temperature", "\"follow_up_kit_line\":", "\"weather_lead\":", "\"weather_point_first\":", "\"weather_next\": ["} {
		if strings.Contains(string(data), copy) {
			t.Errorf("LOCKED: a settings file holds a copy of the golden weather run: %s", copy)
		}
	}
}

// a state or province alone is its capital.
func TestStateCapital(t *testing.T) {
	for place, want := range map[string]string{
		"Kansas":           "Topeka",
		"Oz in Kansas":     "Topeka",
		"kansas":           "Topeka",
		"in West Virginia": "",
		"West Virginia":    "Charleston",
		"Ontario":          "Toronto",
		"toronto ontario":  "",
		"Honolulu Hawaii":  "",
		"Kansas City":      "",
		"New York":         "",
	} {
		if _, got := gopodStateCapital(place); got != want {
			t.Errorf("%q: %q, want %q", place, got, want)
		}
	}
}

// LOCKED Brobot A's brain gets PLAYHEAD's hand-off, then Cache PYC's ready line, then he thanks
// her; Brobot B's gets Cache PYC's ready line only.
func TestWeatherBrainLock(t *testing.T) {
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("GOPOD_OVERLAY_CHIPPER not set (the deploy sets it)")
	}
	s := gopodBrobotsSettings
	var take gopodManagerTake
	for _, tk := range s.Managers.Takes {
		if tk.Kind == "weather" {
			take = tk
		}
	}
	if take.Host != "CACHE PYC" || take.Via != "PLAYHEAD" || take.From != "GIT REPO" || !take.BrobotSays || gopodCallName(take.From) != "GIT REPO" {
		t.Errorf("LOCKED: the weather take isn't PLAYHEAD handing to Cache PYC, from Git Repo: %+v", take)
	}
	if s.Managers.HandoffText != "{intern} will help you with that." ||
		s.WeatherLines["intern_ready"] != "Weather info received from {from}. Ready when you are Bro!" ||
		s.BubbleNotices.ValuePoints.GotItAgain != "Thanks for your help, {helper}." {
		t.Errorf("LOCKED: a brain line of the weather run changed: %q / %q / %q", s.Managers.HandoffText, s.WeatherLines["intern_ready"], s.BubbleNotices.ValuePoints.GotItAgain)
	}
	if gopodWeatherInternFor("what's the weather in Tampa?") != "CACHE PYC" {
		t.Error("LOCKED: a weather line doesn't call Cache PYC in")
	}
	read := func(name string) string {
		b, err := os.ReadFile(filepath.Join(root, "pkg", "wirepod", "ttr", name))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	m := read("gopod_managers.go")
	do := m[strings.Index(m, "func gopodManagerDoFor("):]
	if h, r := strings.Index(do, "gopodHandoff("), strings.Index(do, "gopodWeatherInternReady("); h < 0 || r < h {
		t.Error("LOCKED: Brobot A's brain isn't PLAYHEAD, then Cache PYC")
	}
	n := read("gopod_value_points_notice.go")
	if b := strings.Index(n, "r.NoManager"); b < 0 || !strings.Contains(n[b:b+600], "gopodWeatherInternReady(") {
		t.Error("LOCKED: Brobot B's brain doesn't get Cache PYC's line")
	}
}

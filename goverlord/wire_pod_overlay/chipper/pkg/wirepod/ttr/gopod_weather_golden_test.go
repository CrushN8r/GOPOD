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
// gopod_probes/topic_clusters/show/weather_run.md and only that. The deploy fails when it
// changed without the operator's relock (GOPOD_RELOCK=1 saves its fingerprint,
// gopod_probes/locks/weather_golden.lock), when its order or lines aren't the
// ones in use, or when a settings file holds a copy. The order itself - Doc's
// answer and Pip's - is run through the real answer chain by TestAnswerShapesHold.
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

// 2026-11-03 (live: "the weather of Oz in Kansas, USA" got Kansas, Alabama;
// operator: "Its capital"): a state or province alone is its capital.
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

// LOCKED 2026-11-03 (operator: "Add the brain part of the duo weather run to
// the deploy lock (A: PLAYHEAD > Cache PYC > thanks; B: Cache PYC only)" - "Yes,
// lock it"): Brobot A's brain gets PLAYHEAD's hand-off, then Cache PYC's ready
// line, then he thanks her; Brobot B's gets Cache PYC's ready line only.
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
	// 2026-11-04 (operator: "Fix intern name: 'Git repo' * intern name should be
	// pulled from its topic-cluster"; "Git Repo does the api call ... handsoff to
	// Cache PYC"): the details come from GIT REPO - {from}, his call name.
	if take.Host != "CACHE PYC" || take.Via != "PLAYHEAD" || take.From != "GIT REPO" || !take.BrobotSays || gopodCallName(take.From) != "Git Repo" {
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
	// A: the hand-off, then the intern's line (the thanks follows, once the manager is done).
	m := read("gopod_managers.go")
	do := m[strings.Index(m, "func gopodManagerDoFor("):]
	if h, r := strings.Index(do, "gopodHandoff("), strings.Index(do, "gopodWeatherInternReady("); h < 0 || r < h {
		t.Error("LOCKED: Brobot A's brain isn't PLAYHEAD, then Cache PYC")
	}
	// B: no manager, the intern's line only.
	n := read("gopod_value_points_notice.go")
	if b := strings.Index(n, "r.NoManager"); b < 0 || !strings.Contains(n[b:b+600], "gopodWeatherInternReady(") {
		t.Error("LOCKED: Brobot B's brain doesn't get Cache PYC's line")
	}
}

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_game_engine_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGameRunsFillTheSettings: every run card fills its run name, stop word,
// experience (name and run_name from the card) and, for the draft, play_words;
// the page builds the Play menu from the cards; system-check finds nothing wrong.
func TestGameRunsFillTheSettings(t *testing.T) {
	s := gopodBrobotsSettings
	if len(s.Game.Runs) < 4 {
		t.Fatalf("game.runs: %d cards", len(s.Game.Runs))
	}
	names := map[string]bool{}
	for _, e := range gopodExperienceList() {
		names[e.Name] = true
	}
	for _, r := range s.Game.Runs {
		if s.Session.RunNames[r.Session] != r.Said {
			t.Errorf("%s: run_names[%q] = %q, want %q", r.Name, r.Session, s.Session.RunNames[r.Session], r.Said)
		}
		if !gopodHasString(s.SongStop.StopWords, r.Stop) {
			t.Errorf("%s: %q not in song_stop.stop_words", r.Name, r.Stop)
		}
		if r.Experience != nil && !names[r.Name] {
			t.Errorf("%s: its experience isn't in the list", r.Name)
		}
		if r.Draft && s.InterviewScript.PlayWords != r.Start {
			t.Errorf("%s: play_words %q, want %q", r.Name, s.InterviewScript.PlayWords, r.Start)
		}
	}
	if s.Bingo.Name != "bingo" || s.Bingo.RunName != "Bingo Test" || s.Bingo.NoRunNote == "" {
		t.Errorf("bingo from its card: %q %q", s.Bingo.Name, s.Bingo.RunName)
	}
	js, _ := os.ReadFile(filepath.Join(gopodChipperRoot(), "webroot", "js", "gopod_brobots_chat.js"))
	if !strings.Contains(string(js), "settings.game.runs") {
		t.Errorf("the page doesn't build the Play menu from game.runs")
	}
	if p := gopodGameRunsCheck(); len(p) > 0 {
		t.Errorf("system-check: %v", p)
	}
}

// ## EOF

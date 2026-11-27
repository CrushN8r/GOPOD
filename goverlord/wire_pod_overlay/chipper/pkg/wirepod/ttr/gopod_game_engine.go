// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_game_engine.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
)

// gopodGameRun is one run card (settings game.runs).
type gopodGameRun struct {
	Name       string           `json:"name"`
	Title      string           `json:"title"`
	Song       string           `json:"song"`
	Draft      bool             `json:"draft"`
	Pending    bool             `json:"pending"`
	Start      string           `json:"start"`
	Stop       string           `json:"stop"`
	MenuAbout  string           `json:"menu_about"`
	Session    string           `json:"session"`
	Said       string           `json:"said"`
	Experience *gopodExperience `json:"experience"`
	Lines      []struct {
		When   string   `json:"when"`
		Brobot string   `json:"brobot"`
		Lines  []string `json:"lines"`
	} `json:"lines"`
}

// gopodApplyGameRuns fills the settings from the run cards.
func gopodApplyGameRuns(s *gopodBrobotsSettingsFile) {
	if len(s.Game.Runs) == 0 {
		return
	}
	var experiences []gopodExperience
	var runLines []gopodSnippetLine
	for _, r := range s.Game.Runs {
		for _, l := range r.Lines {
			runLines = append(runLines, gopodSnippetLine{When: r.Name + "_" + l.When, Brobot: l.Brobot, Lines: l.Lines})
		}
		if r.Session != "" {
			if s.Session.RunNames == nil {
				s.Session.RunNames = map[string]string{}
			}
			s.Session.RunNames[r.Session] = r.Said
		}
		if r.Stop != "" && !gopodHasString(s.SongStop.StopWords, r.Stop) {
			s.SongStop.StopWords = append(s.SongStop.StopWords, r.Stop)
		}
		if r.Draft && r.Start != "" {
			s.InterviewScript.PlayWords = r.Start
		}
		if r.Experience == nil {
			continue
		}
		e := *r.Experience
		e.Name, e.RunName = r.Name, r.Title
		if r.Name == "bingo" {
			s.Bingo = e
		} else {
			experiences = append(experiences, e)
		}
	}
	s.Experiences = experiences
	if len(runLines) > 0 {
		s.ScriptedSnippets.Lines = append(runLines, s.ScriptedSnippets.Lines...)
	}
}

func gopodHasString(list []string, want string) bool {
	for _, v := range list {
		if strings.EqualFold(strings.TrimSpace(v), strings.TrimSpace(want)) {
			return true
		}
	}
	return false
}

// gopodGameRunsCheck (system-check): each card has its words, its song folder, and its start and
// stop in Wire-Pod's voice commands (customIntents.json).
func gopodGameRunsCheck() []string {
	var out []string
	songs := gopodSongsDir()
	intents, _ := os.ReadFile(filepath.Join(gopodChipperRoot(), "customIntents.json"))
	lowerIntents := strings.ToLower(string(intents))
	for _, r := range gopodBrobotsSettings.Game.Runs {
		who := "game run " + r.Name
		if r.Name == "" || r.Title == "" || r.Start == "" || r.Stop == "" || r.MenuAbout == "" {
			out = append(out, who+": needs name, title, start, stop and menu_about")
		}
		if r.Song != "" {
			if st, err := os.Stat(filepath.Join(songs, r.Song)); err != nil || !st.IsDir() {
				out = append(out, who+": no song folder "+r.Song)
			}
		}
		for _, w := range []string{r.Start, r.Stop} {
			if w != "" && len(intents) > 0 && !strings.Contains(lowerIntents, "\""+strings.ToLower(w)+"\"") {
				out = append(out, who+": \""+w+"\" isn't a Wire-Pod voice command (customIntents.json)")
			}
		}
	}
	return out
}

// ## EOF

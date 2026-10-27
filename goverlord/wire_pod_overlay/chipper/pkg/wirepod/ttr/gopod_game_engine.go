// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_game_engine.go
package wirepod_ttr

// The Unified Templated GOPOD Game Engine (2026-10-27, operator: "Unified
// Templated GOPOD Game Engine!! * clear?"; "Then the Engine: one game template
// that every run plugs into"). Every run - the Bingo test, the interview, the
// Baby Brobots nap, the Fart Joke - is one card in the settings' game.runs:
// its song folder, its start and stop words, its system-menu line, the name it
// checks in with Goverlord, and its pre/post chat (experience). At load, the
// cards fill the places that used to be kept by hand (session.run_names,
// song_stop.stop_words, bingo, experiences, interview_script.play_words); the
// page builds the Play menu from them; system-check checks each card. A new
// run is one new card.

import (
	"os"
	"path/filepath"
	"strings"
)

// gopodGameRun is one run card (settings game.runs).
type gopodGameRun struct {
	// Name: the experience's name (scripted_snippets <name>_played / _not_played).
	Name string `json:"name"`
	// Title: the run's name in the chat ("If you're watching, tell Goverlord to start {run}!").
	Title string `json:"title"`
	// Song: its folder in goverlord/runtime/songs.
	Song string `json:"song"`
	// Draft: played as the scripted draft (start interview), not by the song engine.
	Draft bool `json:"draft"`
	// Start / Stop: the words that start and stop it (typed, numpad or voice).
	Start string `json:"start"`
	Stop  string `json:"stop"`
	// MenuAbout: its line in the system menu's Play group.
	MenuAbout string `json:"menu_about"`
	// Session / Said: the name it checks in with Goverlord, and how Goverlord says it.
	Session string `json:"session"`
	Said    string `json:"said"`
	// Experience: its pre/post chat (gopod_experiences.go); name and run_name come from the card.
	Experience *gopodExperience `json:"experience"`
}

// gopodApplyGameRuns fills the settings from the run cards.
func gopodApplyGameRuns(s *gopodBrobotsSettingsFile) {
	if len(s.Game.Runs) == 0 {
		return
	}
	var experiences []gopodExperience
	for _, r := range s.Game.Runs {
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
}

func gopodHasString(list []string, want string) bool {
	for _, v := range list {
		if strings.EqualFold(strings.TrimSpace(v), strings.TrimSpace(want)) {
			return true
		}
	}
	return false
}

// gopodGameRunsCheck (system-check): each card has its words, its song folder,
// and its start and stop in Wire-Pod's voice commands (customIntents.json).
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

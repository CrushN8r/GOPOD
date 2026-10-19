// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_song_standard_test.go
package wirepod_ttr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The song standard (ALIAS-PLAY-STUDIO.md, operator 2026-10-09: "I'd like a
// standard format for new song development * and I don't want to present sloppy
// work"): every song folder has a status line, both files name the same steps,
// and every note is one the song's player knows.
var (
	gopodSongStatusLine = regexp.MustCompile(`^\*\*Status:\*\* (GOLDEN|WIP) \(\d{4}-\d{2}-\d{2}\)$`)
	gopodSongStepLine   = regexp.MustCompile(`^## STEP (\S+)\s*$`)
	gopodEngineNotes    = map[string]bool{"wake_both": true, "brobots_ready_together": true, "say_turn": true,
		"say": true, "animation": true, "emotion_beat": true, "arm_cue": true, "nod": true,
		"wheel_nudge": true, "rattle": true, "weather": true, "connect": true, "pause": true, "exit": true,
		"sleep_both": true, "stir": true, "all_asleep": true, "system_line": true, "combo": true, "sound": true, "wake_sleepy": true, "picture": true}
	gopodInterviewNotes = map[string]bool{"ask": true, "line": true, "closer": true}
)

func TestSongFoldersFollowTheStandard(t *testing.T) {
	songs := filepath.Join(gopodChipperRoot(), "..", "..", "runtime", "songs")
	dirs, err := os.ReadDir(songs)
	if err != nil {
		t.Skip("no songs folder: " + err.Error())
	}
	checked := 0
	for _, d := range dirs {
		if !d.IsDir() || d.Name() == "tools" {
			continue
		}
		dir := filepath.Join(songs, d.Name())
		story, err := os.ReadFile(filepath.Join(dir, "story.md"))
		if err != nil {
			t.Errorf("%s: no story.md", d.Name())
			continue
		}
		checked++
		lines := strings.Split(string(story), "\n")
		status := false
		for i, l := range lines {
			if i >= 6 {
				break
			}
			status = status || gopodSongStatusLine.MatchString(strings.TrimSpace(l))
		}
		if !status {
			t.Errorf("%s/story.md: no status line (**Status:** GOLDEN|WIP (YYYY-MM-DD)) under the title", d.Name())
		}
		storySteps := map[string]bool{}
		for _, l := range lines {
			if m := gopodSongStepLine.FindStringSubmatch(strings.TrimSpace(l)); m != nil {
				storySteps[m[1]] = true
			}
		}
		for _, knobsName := range []string{"knobs.json", "zKnobs.json"} {
			data, err := os.ReadFile(filepath.Join(dir, knobsName))
			if err != nil {
				if knobsName == "knobs.json" {
					t.Errorf("%s: no knobs.json", d.Name())
				}
				continue
			}
			var k struct {
				SongType string `json:"song_type"`
				Steps    []struct {
					StepID string `json:"step_id"`
					Note   string `json:"note"`
				} `json:"steps"`
			}
			if err := json.Unmarshal(data, &k); err != nil {
				t.Errorf("%s/%s: %v", d.Name(), knobsName, err)
				continue
			}
			if len(k.Steps) == 0 {
				continue // a launched program, not a step song (the Bingo game)
			}
			known := gopodEngineNotes
			if k.SongType == "interview" {
				known = gopodInterviewNotes
			}
			// 2026-10-16: an LLM warm song (the opener), played by gopod_opener_song.go.
			if k.SongType == "llm_warm" {
				known = gopodOpenerNotes
			}
			inKnobs := map[string]bool{}
			for _, s := range k.Steps {
				inKnobs[s.StepID] = true
				if !known[s.Note] {
					t.Errorf("%s/%s: step %s has an unknown note %q", d.Name(), knobsName, s.StepID, s.Note)
				}
				if !storySteps[s.StepID] {
					t.Errorf("%s: step %s is in %s but not in story.md", d.Name(), s.StepID, knobsName)
				}
			}
			for id := range storySteps {
				if !inKnobs[id] {
					t.Errorf("%s: step %s is in story.md but not in %s", d.Name(), id, knobsName)
				}
			}
		}
	}
	if checked == 0 {
		t.Error("no song folders checked")
	}
}

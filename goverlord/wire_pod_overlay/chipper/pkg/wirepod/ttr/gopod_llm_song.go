// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_llm_song.go
package wirepod_ttr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

const gopodLLMSongDir = "_llm_song"

// gopodLLMSongLines: the run's lines, and whether the folder was there to read.
func gopodLLMSongLines(songDir string) ([]gopodHostLine, bool) {
	dir := filepath.Join(songDir, gopodLLMSongDir)
	data, err := os.ReadFile(filepath.Join(dir, "knobs.json"))
	if err != nil {
		return nil, false
	}
	var k struct {
		Steps []struct {
			StepID  string `json:"step_id"`
			Note    string `json:"note"`
			Speaker string `json:"speaker"`
		} `json:"steps"`
	}
	if err := json.Unmarshal(data, &k); err != nil {
		logger.Println("GOPOD_CHECK " + dir + ": knobs.json doesn't read (" + err.Error() + ") - no lines")
		return nil, false
	}
	texts := map[string]string{}
	if story, err := os.ReadFile(filepath.Join(dir, "story.md")); err == nil {
		texts, _ = gopodInterviewStory(string(story))
	}
	var out []gopodHostLine
	for _, s := range k.Steps {
		host := gopodReporterHost(s.Speaker)
		text := strings.TrimSpace(texts[s.StepID])
		if s.Note != "line" || host == "" || text == "" {
			continue
		}
		out = append(out, gopodHostLine{Host: host, Text: text})
	}
	return out, true
}

var gopodOpenerLLMSong = struct {
	sync.Mutex
	loaded bool
	lines  []gopodHostLine
	ok     bool
}{}

// gopodOpenerLLMSongLines: the opener's run (read once), and whether it uses one.
func gopodOpenerLLMSongLines() ([]gopodHostLine, bool) {
	gopodOpenerLLMSong.Lock()
	defer gopodOpenerLLMSong.Unlock()
	if !gopodOpenerLLMSong.loaded {
		gopodOpenerLLMSong.loaded = true
		id := strings.TrimSpace(gopodBrobotsSettings.LLM.OpenerSongID)
		if id != "" {
			if dir, err := gopodFindSong(id); err == nil {
				uses := false
				if steps, err := gopodReadOpenerSong(dir); err == nil {
					for _, s := range steps {
						uses = uses || s.Note == "llm_song"
					}
				}
				if uses {
					lines, found := gopodLLMSongLines(dir)
					if !found {
						logger.Println("GOPOD_CHECK opener: no " + gopodLLMSongDir + " folder - its warm-up says no lines")
					}
					gopodOpenerLLMSong.lines, gopodOpenerLLMSong.ok = lines, true
				}
			}
		}
	}
	return gopodOpenerLLMSong.lines, gopodOpenerLLMSong.ok
}

// ## EOF

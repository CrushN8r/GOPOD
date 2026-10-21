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

// The _llm_song run (2026-10-20, operator: "load 'internal-run' from expected
// required external mini-run/song file, per llm * then all llm warmings load an
// expected run (required. the run file can be a non-run, 0 setting) * I can
// then mix/shuffle internal runs as I desire without breaking GOPOD"; "If
// _llm_song folder name is unchanged, that's how we make the folders
// interchangeable"). Every LLM warm-up - the opener (00_llm_warming) and each
// song's - says its folder's _llm_song run between "Voices ready" and "LLM
// ready": _llm_song/knobs.json (steps, note "line", speaker a reporter host)
// and _llm_song/story.md (each step's words). Swap the folder to change it;
// no steps = nothing said. The guards: system-check flags a warm-up with no
// _llm_song; a missing or unreadable one plays as no lines and logs
// GOPOD_CHECK; Esc / system-skip skips its lines on the page.

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

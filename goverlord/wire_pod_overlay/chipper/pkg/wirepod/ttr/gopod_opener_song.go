// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_opener_song.go
package wirepod_ttr

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

type gopodOpenerStep struct {
	StepID     string          `json:"step_id"`
	Note       string          `json:"note"`
	WaitVoices bool            `json:"wait_voices"`
	Volume     json.RawMessage `json:"volume"`
	Text       string          `json:"-"`
	Done       string          `json:"-"`
}

var gopodOpenerNotes = map[string]bool{"warm_message": true, "warm_wait": true, "reporter_lines": true, "llm_song": true, "warm_ready": true, "sting": true}

// gopodOpenerSong: the opener song's steps, with their story.md words; nil with no opener song set.
var gopodOpenerSong = func() ([]gopodOpenerStep, error) {
	id := strings.TrimSpace(gopodBrobotsSettings.LLM.OpenerSongID)
	if id == "" {
		return nil, nil
	}
	dir, err := gopodFindSong(id)
	if err != nil {
		return nil, err
	}
	return gopodReadOpenerSong(dir)
}

// gopodReadOpenerSong reads an LLM warm song folder (knobs.json + story.md).
func gopodReadOpenerSong(dir string) ([]gopodOpenerStep, error) {
	data, err := os.ReadFile(filepath.Join(dir, "knobs.json"))
	if err != nil {
		return nil, err
	}
	var knobs struct {
		Steps []gopodOpenerStep `json:"steps"`
	}
	if err := json.Unmarshal(data, &knobs); err != nil {
		return nil, errors.New("knobs.json: " + err.Error())
	}
	story, err := os.ReadFile(filepath.Join(dir, "story.md"))
	if err != nil {
		return nil, err
	}
	texts, err := gopodInterviewStory(string(story))
	if err != nil {
		return nil, err
	}
	if len(knobs.Steps) == 0 {
		return nil, errors.New("knobs.json has no steps")
	}
	for i, s := range knobs.Steps {
		if !gopodOpenerNotes[s.Note] {
			return nil, errors.New("step " + s.StepID + ": note " + s.Note + " isn't an LLM warm note")
		}
		if _, err := gopodVolumePercent(s.Volume); err != nil {
			return nil, errors.New("step " + s.StepID + ": " + err.Error())
		}
		text, ok := texts[s.StepID]
		if !ok {
			return nil, errors.New("step " + s.StepID + " isn't in story.md")
		}
		knobs.Steps[i].Text = text
	}
	return knobs.Steps, nil
}

// gopodBuiltInOpener: the opener from the settings (llm.voices_text, warmup_text, ready_text,
// warmup_done_text) - the same steps as the song.
func gopodBuiltInOpener() []gopodOpenerStep {
	llm := gopodBrobotsSettings.LLM
	return []gopodOpenerStep{
		{StepID: "voices", Note: "warm_message", WaitVoices: true, Text: llm.VoicesText},
		{StepID: "loading", Note: "warm_message", Text: llm.WarmupText},
		{StepID: "llm", Note: "warm_wait"},
		{StepID: "opening", Note: "reporter_lines", Text: "opening"},
		{StepID: "ready", Note: "warm_ready", Text: llm.ReadyText, Done: llm.WarmupDoneText},
		{StepID: "sting", Note: "sting"},
	}
}

// gopodOpenerSteps: the opener song's steps, or the built-in opener.
func gopodOpenerSteps() []gopodOpenerStep {
	steps, err := gopodOpenerSong()
	if err != nil {
		logger.Println("GOPOD_OPENER_SONG unable to read it, the built-in opener plays: " + err.Error())
	}
	if len(steps) == 0 {
		return gopodBuiltInOpener()
	}
	return steps
}

// gopodOpenerSaid: the lines the opener's system voice says (made at start).
func gopodOpenerSaid() []string {
	var lines []string
	for _, s := range gopodOpenerSteps() {
		switch s.Note {
		case "warm_message":
			lines = append(lines, gopodMessageLines(gopodWarmupLine(s.Text))...)
		case "warm_ready":
			lines = append(lines, gopodMessageLines(s.Text)...)
			lines = append(lines, s.Done)
		}
	}
	return lines
}

// gopodMessageLines: a message's lines, each one said on its own.
func gopodMessageLines(text string) []string {
	var lines []string
	for _, l := range strings.Split(text, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, l)
		}
	}
	if len(lines) == 0 {
		lines = []string{strings.TrimSpace(text)}
	}
	return lines
}

// gopodLinesMade: every line's system voice file is made already.
var gopodLinesMade = func(lines []string) bool {
	for _, line := range lines {
		if _, _, ok := gopodReporterCached(gopodBrobotsSettings.SignOff.SystemRole,
			gopodSaysFor(gopodBrobotsSettings.SignOff.SystemRole, line)); !ok {
			return false
		}
	}
	return true
}

// gopodMakeLines makes every line's system voice file (waiting for the voice worker).
func gopodMakeLines(lines []string) {
	for _, line := range lines {
		gopodReporterFile(gopodBrobotsSettings.SignOff.SystemRole, gopodSaysFor(gopodBrobotsSettings.SignOff.SystemRole, line))
	}
}

// gopodSkipOnMark: the feed line that opens the page's skip gate.
const gopodSkipOnMark = "skip_on"

// gopodPlayOpener plays the opener's steps; warmed closes once the model is loaded.
func gopodPlayOpener(steps []gopodOpenerStep, warmed <-chan struct{}) {
	defer gopodSetVolumeShare(1)
	voicesChecked, opened := false, false
	for _, s := range steps {
		gopodOpenerAwaitPage()
		share, _ := gopodVolumePercent(s.Volume)
		gopodSetVolumeShare(share)
		switch s.Note {
		case "warm_message":
			if strings.TrimSpace(s.Text) == "" {
				continue
			}
			lines := gopodMessageLines(gopodWarmupLine(s.Text))
			waiting := s.WaitVoices && !gopodWarmupVoicesMade()
			if waiting || !gopodLinesMade(lines) {
				gopodWarmupMessage(strings.TrimSpace(lines[0] + "\n" + gopodBrobotsSettings.ReporterVoices.GeneratingText))
				gopodWarmStingStart.Store(false)
			}
			if waiting {
				gopodWarmupVoicesWait()
			}
			if !gopodLinesMade(lines) {
				gopodMakeLines(lines)
			}
			skipOn := voicesChecked
			voicesChecked = true
			for i, line := range lines {
				if skipOn && !opened && i == 0 {
					opened = true
					GopodChatFeedAdd(gopodBrobotsSettings.ReporterVoices.Role, gopodSkipOnMark)
				}
				gopodWarmupMessage(strings.Join(lines[:i+1], "\n"))
				gopodWarmStingStart.Store(false)
				gopodSystemSay(line, false)
			}
		case "warm_wait":
			<-warmed
		case "llm_song":
			if lines, _ := gopodOpenerLLMSongLines(); len(lines) == 0 {
				continue
			}
			gopodWarmupOpening()
		case "reporter_lines":
			if strings.TrimSpace(s.Text) == "" {
				continue
			}
			gopodWarmupOpening()
		case "warm_ready":
			gopodLLMWarmupReady(gopodWarmupReadyLine(s.Text, s.Done), true)
			readyLines := gopodMessageLines(s.Text)
			gopodMakeLines(readyLines)
			for i, line := range readyLines {
				gopodSystemSay(line, s.Done == "" && i == len(readyLines)-1)
			}
			if s.Done != "" {
				gopodSystemSayNoSting(s.Done, true)
			}
		case "sting":
			gopodWarmupEndSting()
		}
	}
	gopodMicCheck()
}

// gopodVolumePercent: a volume knob as a share of the master (chat_audio.volume) - 1 for "" or
// "default", 0.4 for "40%"; a bare number is already a share (0.4, the older form).
func gopodVolumePercent(raw json.RawMessage) (float64, error) {
	text := strings.TrimSpace(string(raw))
	quoted := strings.HasPrefix(text, `"`)
	text = strings.TrimSpace(strings.Trim(text, `"`))
	if text == "" || text == "null" || strings.EqualFold(text, "default") {
		return 1, nil
	}
	percent := strings.HasSuffix(text, "%")
	n, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(text, "%")), 64)
	if err != nil || n < 0 || (quoted && !percent) {
		return 1, errors.New("volume " + text + " isn't \"\" (default) or a percentage (\"40%\")")
	}
	if percent {
		return n / 100, nil
	}
	return n, nil
}

// gopodVolumeShare: the share the voices and stings queued now carry (1: as ever).
var gopodVolumeShare = struct {
	sync.Mutex
	share float64
}{share: 1}

func gopodSetVolumeShare(share float64) {
	gopodVolumeShare.Lock()
	gopodVolumeShare.share = share
	gopodVolumeShare.Unlock()
}

// gopodWithVolumeShare adds the current share to a voice or sting path ("x.wav" -> "x.wav#vol=0.5";
// one with its own "#vol=" gets the two multiplied).
var gopodVolRe = regexp.MustCompile(`#vol=([0-9.]+)$`)

func gopodWithVolumeShare(page string) string {
	gopodVolumeShare.Lock()
	share := gopodVolumeShare.share
	gopodVolumeShare.Unlock()
	if share == 1 || page == "" {
		return page
	}
	if m := gopodVolRe.FindStringSubmatch(page); m != nil {
		if own, err := strconv.ParseFloat(m[1], 64); err == nil {
			share *= own
		}
		page = strings.TrimSuffix(page, m[0])
	}
	if share == 1 {
		return page
	}
	return page + "#vol=" + strconv.FormatFloat(share, 'f', -1, 64)
}

// ## EOF

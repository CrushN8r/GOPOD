// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_interview_song.go
package wirepod_ttr

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type gopodInterviewKnobs struct {
	SongID   string                    `json:"song_id"`
	SongType string                    `json:"song_type"`
	StoryRef string                    `json:"story_ref"`
	About    string                    `json:"about,omitempty"`
	Steps    []gopodInterviewKnobsStep `json:"steps"`
}

type gopodInterviewKnobsStep struct {
	StepID   string           `json:"step_id"`
	Note     string           `json:"note"` // "ask" or "line"
	Speaker  string           `json:"speaker"`
	To       string           `json:"to,omitempty"`
	Cue      string           `json:"cue,omitempty"`
	Moves    []gopodDraftMove `json:"moves,omitempty"`
	Section  string           `json:"section"`
	Chapter  string           `json:"chapter,omitempty"`
	Pictures []string         `json:"pictures,omitempty"`
}

// The knobs' speaker names and the draft's.
var gopodInterviewSpeakers = map[string]string{
	"brobot_1": "Brobot 1", "brobot_2": "Brobot 2", "brobots": "Brobots",
	"reporter_1": "CHALK reporter", "reporter_2": "PLAYHEAD reporter",
}

func gopodInterviewSpeaker(id string) (string, error) {
	name, ok := gopodInterviewSpeakers[id]
	if !ok {
		return "", errors.New("unknown speaker " + id)
	}
	return name, nil
}

func gopodInterviewSpeakerID(name string) string {
	for id, n := range gopodInterviewSpeakers {
		if n == name {
			return id
		}
	}
	return ""
}

// gopodInterviewStory reads story.md: the text of each step, and each step's section.
func gopodInterviewStory(text string) (map[string]string, error) {
	texts := map[string]string{}
	step := ""
	inText := false
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case strings.HasPrefix(line, "## STEP "):
			step = strings.TrimSpace(strings.TrimPrefix(line, "## STEP "))
			if _, dup := texts[step]; dup {
				return nil, errors.New("story.md: step " + step + " twice")
			}
			texts[step] = ""
			inText = false
		case strings.HasPrefix(line, "#"):
			step, inText = "", false
		case step != "" && strings.HasPrefix(line, "> TEXT:"):
			texts[step] = strings.TrimSpace(strings.TrimPrefix(line, "> TEXT:"))
			inText = true
		case step != "" && strings.HasPrefix(line, ">") && texts[step] != "":
			texts[step] += " " + strings.TrimSpace(strings.TrimPrefix(line, ">"))
		case step != "" && inText && line != "":
			if texts[step] == "" {
				texts[step] = line
			} else {
				texts[step] += "\n" + line
			}
		default:
			inText = false
		}
	}
	return texts, nil
}

// gopodInterviewSongDraft reads a song folder into the draft the player plays.
func gopodInterviewSongDraft(dir string) (gopodDraft, error) {
	var d gopodDraft
	knobsPath := filepath.Join(dir, "knobs.json")
	gopodKnobsMDBuild(dir) // zKnobs.md, the one source, first
	if _, err := os.Stat(knobsPath); err != nil {
		if data, rerr := os.ReadFile(filepath.Join(dir, "zKnobs.json")); rerr == nil {
			_ = gopodWriteFile(knobsPath, data, 0o644)
		}
	}
	data, err := os.ReadFile(knobsPath)
	if err != nil {
		return d, err
	}
	var k gopodInterviewKnobs
	if err := json.Unmarshal(data, &k); err != nil {
		return d, errors.New(knobsPath + ": " + err.Error())
	}
	story := k.StoryRef
	if story == "" {
		story = "story.md"
	}
	data, err = os.ReadFile(filepath.Join(dir, story))
	if err != nil {
		return d, err
	}
	texts, err := gopodInterviewStory(string(data))
	if err != nil {
		return d, err
	}
	d.Run = k.SongID
	var q *gopodDraftQuestion
	chapter := ""
	for _, s := range k.Steps {
		text, ok := texts[s.StepID]
		if !ok {
			return d, errors.New("step " + s.StepID + " is in the knobs but not in story.md")
		}
		who, err := gopodInterviewSpeaker(s.Speaker)
		if err != nil {
			return d, errors.New("step " + s.StepID + ": " + err.Error())
		}
		if s.Note == "closer" {
			if strings.TrimSpace(text) == "" {
				return d, errors.New("step " + s.StepID + " has no > TEXT: in story.md")
			}
			d.Closer = append(d.Closer, gopodDraftLine{Who: who, Text: text})
			continue
		}
		if s.Chapter != "" {
			chapter = s.Chapter
		}
		if q == nil || s.Section != q.Topic {
			d.Questions = append(d.Questions, gopodDraftQuestion{N: len(d.Questions) + 1, Topic: s.Section, Chapter: chapter})
			q = &d.Questions[len(d.Questions)-1]
		}
		if len(s.Pictures) > 0 {
			q.Pictures = append(q.Pictures, s.Pictures...)
		}
		switch s.Note {
		case "ask":
			to, err := gopodInterviewSpeaker(s.To)
			if err != nil || strings.HasPrefix(to, "Reporter") {
				return d, errors.New("step " + s.StepID + ": an ask is to brobot_1, brobot_2 or brobots")
			}
			if q.Ask != "" {
				return d, errors.New("step " + s.StepID + ": a second ask in section " + s.Section)
			}
			q.ReadBy, q.To, q.Ask = who, to, text
		case "line":
			if strings.TrimSpace(text) == "" {
				return d, errors.New("step " + s.StepID + " has no > TEXT: in story.md")
			}
			if q.Ask == "" {
				if !strings.HasPrefix(who, "Reporter") {
					return d, errors.New("step " + s.StepID + ": a Brobot line before the section's ask")
				}
				q.Before = append(q.Before, gopodDraftLine{Who: who, Text: text})
				continue
			}
			q.Answers = append(q.Answers, gopodDraftAnswer{Brobot: who, Cue: s.Cue, Say: text, Moves: s.Moves})
		default:
			return d, errors.New("step " + s.StepID + ": note is \"ask\", \"line\" or \"closer\"")
		}
	}
	return d, nil
}

// gopodInterviewSongFiles writes a draft as a song folder's story.md and knobs.json.
func gopodInterviewSongFiles(d gopodDraft) (story string, knobs []byte, err error) {
	var b strings.Builder
	b.WriteString("# Brobots Interview\n\n**Status:** WIP\n\n")
	b.WriteString("The scripted interview. Type `start interview` in the Brobots chat to play it, `stop run` to stop\n")
	b.WriteString("it at a safe spot.\n\n")
	b.WriteString("- **The lines** live in this file (`## STEP` / `> TEXT:`). **Who says each one, the Brobot's\n")
	b.WriteString("  animation cue and his moves at words** live in [knobs.json](knobs.json). The two must name the\n")
	b.WriteString("  same steps; changing a line's words here is enough on its own.\n")
	b.WriteString("- Each `## SECTION` is one question: the reporters' banter, the question (`note: ask`), then the\n")
	b.WriteString("  answers and handoffs, in order. `{weather}` in a line is the live weather.\n")
	k := gopodInterviewKnobs{SongID: "brobots_interview", SongType: "interview", StoryRef: "story.md",
		About: "The scripted interview (gopod_interview_song.go). note: ask (the question, read by speaker, to to) or line (said by speaker); cue: the Brobot's animation; moves: at their words (arm, nod, cube). Played by \"start interview\" in the Brobots chat, not by pha0b."}
	chapter := ""
	for _, q := range d.Questions {
		b.WriteString("\n---\n\n## SECTION: " + q.Topic + "\n")
		first := len(k.Steps)
		n := 0
		step := func(who, text string) string {
			n++
			id := fmt.Sprintf("q%02d_%02d_%s", q.N, n, gopodInterviewSpeakerID(who))
			b.WriteString("\n## STEP " + id + "\n\n> TEXT: " + text + "\n")
			return id
		}
		for _, l := range q.Before {
			k.Steps = append(k.Steps, gopodInterviewKnobsStep{StepID: step(l.Who, l.Text), Note: "line", Speaker: gopodInterviewSpeakerID(l.Who), Section: q.Topic})
		}
		k.Steps = append(k.Steps, gopodInterviewKnobsStep{StepID: step(q.ReadBy, q.Ask), Note: "ask", Speaker: gopodInterviewSpeakerID(q.ReadBy), To: gopodInterviewSpeakerID(q.To), Section: q.Topic})
		for _, a := range q.Answers {
			k.Steps = append(k.Steps, gopodInterviewKnobsStep{StepID: step(a.Brobot, a.Say), Note: "line", Speaker: gopodInterviewSpeakerID(a.Brobot), Cue: a.Cue, Moves: a.Moves, Section: q.Topic})
		}
		if q.Chapter != chapter {
			chapter = q.Chapter
			k.Steps[first].Chapter = chapter
		}
		for i := first; i < len(k.Steps); i++ {
			if k.Steps[i].Note == "ask" {
				k.Steps[i].Pictures = q.Pictures
			}
		}
	}
	if len(d.Closer) > 0 {
		b.WriteString("\n---\n\n## SECTION: Closer\n")
		for i, l := range d.Closer {
			id := fmt.Sprintf("closer_%02d_%s", i+1, gopodInterviewSpeakerID(l.Who))
			b.WriteString("\n## STEP " + id + "\n\n> TEXT: " + l.Text + "\n")
			k.Steps = append(k.Steps, gopodInterviewKnobsStep{StepID: id, Note: "closer", Speaker: gopodInterviewSpeakerID(l.Who), Section: "Closer"})
		}
	}
	knobs, err = json.MarshalIndent(k, "", "  ")
	return b.String(), append(knobs, '\n'), err
}

// ## EOF

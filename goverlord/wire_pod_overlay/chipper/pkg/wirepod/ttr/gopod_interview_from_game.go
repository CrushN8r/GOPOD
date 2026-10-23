// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_interview_from_game.go
package wirepod_ttr

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The interview song from the game (2026-10-23, FINISH_PLAN_001.md step A5;
// operator: "commit, then A5 please"): the scripted interview is built from the
// game's own template, so the two never drift apart - one chapter per core topic
// (value_points.core_topics, in order), one question and answer per value point
// (core_points). The question is that point's golden interview question
// (interview_harvest/harvest_questions.json); the answer is the operator's own
// words - the point's notes, read as sentences (else its say line, else its
// words) - said by the Brobot the question is to; the reporters take turns. The
// closer stays the song's own. Each answer step names its point, so playing the
// song scores the game (gopodScriptAnswer), and the grand prize plays before the
// closer. Written with the song's own writer (gopodInterviewSongFiles); rebuilt
// by system-interview-rebuild.

// gopodGameQuestion: one golden interview question (harvest_questions.json).
type gopodGameQuestion struct {
	Topic string `json:"topic"`
	To    string `json:"to"`
	Ask   string `json:"ask"`
}

// gopodNotesSentences: "Morning | seniors' bingo | Merch too!" -> "Morning.
// Seniors' bingo. Merch too!" - his words, each a sentence.
func gopodNotesSentences(title string, notes []string) string {
	var out []string
	for _, p := range append([]string{title}, notes...) {
		p = gopodCapFirst(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		if !strings.ContainsAny(p[len(p)-1:], ".!?") {
			p += "."
		}
		out = append(out, p)
	}
	return strings.Join(out, " ")
}

// gopodInterviewDraftFromGame: the interview, as a draft, from the game.
func gopodInterviewDraftFromGame(closer []gopodDraftLine) (gopodDraft, error) {
	d := gopodDraft{Run: "brobots_interview", Closer: closer}
	var qfile struct {
		Questions []gopodGameQuestion `json:"questions"`
	}
	data, err := os.ReadFile(filepath.Join(gopodChipperRoot(), "gopod_probes/interview_harvest/harvest_questions.json"))
	if err != nil || json.Unmarshal(data, &qfile) != nil {
		return d, err
	}
	asked := map[string][]gopodGameQuestion{} // topic -> its questions, in point order
	for _, q := range qfile.Questions {
		asked[q.Topic] = append(asked[q.Topic], q)
	}
	contents := map[string][]byte{}
	for _, p := range gopodTopicRegistry() {
		contents[p.Name], _ = gopodReadTopicFacts(p.File)
	}
	ib := gopodBrobotsSettings.InterviewBuild
	reporters := ib.Reporters
	if len(reporters) == 0 {
		return d, errors.New("interview_build.reporters is empty")
	}
	for _, core := range gopodBrobotsSettings.Summary.TopicOrder {
		files := gopodBrobotsSettings.ValuePoints.CoreTopics[core]
		for _, topic := range files {
			_, points := gopodSplitPoints(string(contents[topic]))
			total := len(points)
			if total == 0 {
				total = 1
			}
			for n := 1; n <= total; n++ {
				if !gopodCorePoint(topic, n) {
					continue
				}
				q := gopodDraftQuestion{N: len(d.Questions) + 1, Chapter: core}
				label := strings.TrimSpace(gopodBrobotsSettings.InterviewChapters.TopicLabels[topic])
				if label == "" {
					label = topic
				}
				q.Topic = gopodCapFirst(label)
				// Numbered among the topic's value points, when it has more than one.
				rank, count := 0, 0
				for i := 1; i <= total; i++ {
					if gopodCorePoint(topic, i) {
						count++
						if i <= n {
							rank = count
						}
					}
				}
				if count > 1 {
					q.Topic += " " + strconv.Itoa(rank)
				}
				q.ReadBy = reporters[len(d.Questions)%len(reporters)]
				q.To, q.Ask = "Brobot 1", strings.ReplaceAll(ib.DefaultAsk, "{topic}", label)
				if list := asked[topic]; n-1 < len(list) {
					q.To, q.Ask = list[n-1].To, list[n-1].Ask
				}
				if n == 1 || len(d.Questions) == 0 {
					q.Pictures = append([]string(nil), gopodBrobotsSettings.TopicPictures.Pictures[topic]...)
				}
				key := topic + " " + strconv.Itoa(n)
				say := ""
				if title, notes := gopodPointNotes(topic, n); title != "" {
					say = gopodNotesSentences(title, notes)
				} else if l := strings.TrimSpace(gopodBrobotsSettings.ValuePoints.SayLines[key]); l != "" {
					say = l
				} else if n-1 < len(points) {
					say = strings.TrimSpace(gopodPointLine.ReplaceAllString(points[n-1], ""))
				} else {
					say = strings.TrimSpace(string(contents[topic]))
				}
				q.Answers = []gopodDraftAnswer{{Brobot: q.To, Cue: ib.Cues[q.To], Say: say, Points: []string{topic + "#" + strconv.Itoa(n)}}}
				d.Questions = append(d.Questions, q)
			}
		}
	}
	return d, nil
}

// gopodInterviewRebuild writes the interview song from the game (undoable).
func gopodInterviewRebuild() string {
	dir, err := gopodFindSong(gopodBrobotsSettings.InterviewScript.SongID)
	if err != nil {
		return "system-interview-rebuild: " + err.Error()
	}
	var closer []gopodDraftLine
	if old, err := gopodInterviewSongDraft(dir); err == nil {
		closer = old.Closer
	}
	d, err := gopodInterviewDraftFromGame(closer)
	if err != nil {
		return "system-interview-rebuild: " + err.Error()
	}
	story, knobs, err := gopodInterviewSongFiles(d)
	if err != nil {
		return "system-interview-rebuild: " + err.Error()
	}
	if err := gopodWriteFileSafe(filepath.Join(dir, "story.md"), []byte(story)); err != nil {
		return "system-interview-rebuild: " + err.Error()
	}
	if err := gopodWriteFileSafe(filepath.Join(dir, "knobs.json"), knobs); err != nil {
		return "system-interview-rebuild: " + err.Error()
	}
	// The private working copy (zKnobs.json), when there, kept the same.
	if _, err := os.Stat(filepath.Join(dir, "zKnobs.json")); err == nil {
		if err := gopodWriteFileSafe(filepath.Join(dir, "zKnobs.json"), knobs); err != nil {
			return "system-interview-rebuild: " + err.Error()
		}
	}
	logger.Println("GOPOD_INTERVIEW rebuilt from the game: " + strconv.Itoa(len(d.Questions)) + " questions")
	return "Interview rebuilt from the game: " + strconv.Itoa(len(d.Questions)) + " questions, one per value point, in " + strconv.Itoa(len(gopodBrobotsSettings.ValuePoints.CoreTopics)) + " chapters (system-undo puts the old one back)."
}

// gopodCoverageMark (A5): a scripted answer reveals its point ("topic#n") - once.
func gopodCoverageMark(id string) {
	f := strings.SplitN(id, "#", 2)
	if len(f) != 2 {
		return
	}
	n, err := strconv.Atoi(f[1])
	if err != nil || !gopodCorePoint(f[0], n) {
		return
	}
	gopodCoverage.Lock()
	defer gopodCoverage.Unlock()
	said := gopodCoverageSaid[f[0]]
	if said == nil {
		said = map[int]bool{}
	}
	if said[n] {
		return
	}
	said[n] = true
	gopodCoverageSaid[f[0]] = said
	gopodCoverage.revealed[f[0]]++
	logger.Println("BROBOT_VALUE_POINT topic=" + f[0] + " point=" + f[1] + " (the interview's answer)")
}

// ## EOF

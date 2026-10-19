// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_draft_script.go
package wirepod_ttr

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// The draft as a plain script to read, edit and replay (operator 2026-10-09:
// "Ensure draft run is composed in a human friendly format, I can replay, and I
// can make edits"). draft_script.md sits beside draft_performance.json and,
// when it's there, is what "start interview" (interview_script.play_words) plays (gopodLoadDraft):
//
//	# Draft script: <run>
//	> A line starting with ">" is a note; blank lines don't count.
//	## 1. Who they are                      a question starts
//	Reporter 1: banter before the question  (optional)
//	Reporter 2 asks Brobot 1: the question  ("Brobots" asks both)
//	Brobot 1: his answer                    "Brobot 1 (happy):" picks the animation
//	moves: lift my arm = arm; nod = nod     moves at those words (arm, nod, cube)
//	Brobot 2: his brother's line            then reporters' and Brobots' lines, in order
//
// "{weather}" in a line is the live weather (gopod_script_moves.go).

const gopodDraftScriptName = "draft_script.md"

var (
	gopodDraftAskLine  = regexp.MustCompile(`^(Reporter [12])(?:, Brobot [34])? asks (Brobot [12]|Brobots):\s*(.+)$`)
	gopodDraftSayLine  = regexp.MustCompile(`^(Brobot [12]|Reporter [12])(?:, Brobot [34])?(?: \(([^)]*)\))?:\s*(.+)$`)
	gopodDraftMoveLine = regexp.MustCompile(`^moves:\s*(.+)$`)
)

// gopodDraftReporterLabel: "Reporter 1" -> its label in reporter_voices.labels
// ("Reporter 1, Brobot 3").
func gopodDraftReporterLabel(short string) string {
	for _, label := range gopodBrobotsSettings.ReporterVoices.Labels {
		if strings.HasPrefix(label, short+",") || strings.HasPrefix(label, short+":") || label == short {
			return label
		}
	}
	return short
}

// gopodDraftReporterShort: "Reporter 1: Brobot 3" (or "Reporter 1, Brobot 3") -> "Reporter 1".
func gopodDraftReporterShort(label string) string {
	if i := strings.IndexAny(label, ",:"); i > 0 && strings.HasPrefix(label, "Reporter") {
		return label[:i]
	}
	return label
}

// gopodParseDraftScript reads a draft_script.md.
func gopodParseDraftScript(text string) (gopodDraft, error) {
	var d gopodDraft
	var q *gopodDraftQuestion
	for i, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		at := "line " + strconv.Itoa(i+1) + ": "
		switch {
		case line == "" || strings.HasPrefix(line, ">"):
			continue
		case strings.HasPrefix(line, "## "):
			d.Questions = append(d.Questions, gopodDraftQuestion{N: len(d.Questions) + 1, Topic: strings.TrimSpace(line[3:])})
			q = &d.Questions[len(d.Questions)-1]
		case strings.HasPrefix(line, "# "):
			d.Run = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line[2:]), "Draft script:"))
		case q == nil:
			return d, errors.New(at + "a line before the first \"## \" question")
		case gopodDraftAskLine.MatchString(line):
			m := gopodDraftAskLine.FindStringSubmatch(line)
			if q.Ask != "" {
				return d, errors.New(at + "a second question in \"" + q.Topic + "\"")
			}
			q.ReadBy, q.To, q.Ask = gopodDraftReporterLabel(m[1]), m[2], strings.TrimSpace(m[3])
		case gopodDraftMoveLine.MatchString(line):
			if len(q.Answers) == 0 {
				return d, errors.New(at + "moves before an answer")
			}
			a := &q.Answers[len(q.Answers)-1]
			for _, part := range strings.Split(gopodDraftMoveLine.FindStringSubmatch(line)[1], ";") {
				phrase, move, ok := strings.Cut(part, "=")
				if !ok || strings.TrimSpace(phrase) == "" || strings.TrimSpace(move) == "" {
					return d, errors.New(at + "a move is written \"words = move\"")
				}
				a.Moves = append(a.Moves, gopodDraftMove{At: strings.TrimSpace(phrase), Move: strings.TrimSpace(move)})
			}
		case gopodDraftSayLine.MatchString(line):
			m := gopodDraftSayLine.FindStringSubmatch(line)
			who, cue, say := m[1], strings.TrimSpace(m[2]), strings.TrimSpace(m[3])
			if strings.HasPrefix(who, "Reporter") {
				who = gopodDraftReporterLabel(who)
				if q.Ask == "" {
					q.Before = append(q.Before, gopodDraftLine{Who: who, Text: say})
					continue
				}
			} else if q.Ask == "" {
				return d, errors.New(at + "a Brobot line before the question")
			}
			q.Answers = append(q.Answers, gopodDraftAnswer{Brobot: who, Cue: cue, Say: say})
		default:
			return d, errors.New(at + "not a script line: " + line)
		}
	}
	return d, nil
}

// gopodDraftScriptText writes a draft as a draft_script.md.
func gopodDraftScriptText(d gopodDraft) string {
	var b strings.Builder
	b.WriteString("# Draft script: " + d.Run + "\n\n")
	b.WriteString("> Edit the words after any colon. Type \"" + gopodBrobotsSettings.InterviewScript.PlayWords + "\" in the Brobots chat to play it, \"stop run\" to stop it at a safe spot.\n")
	b.WriteString("> \"Brobot 1 (happy):\" picks his animation. \"moves: lift my arm = arm\" does a move (arm, nod, cube) at those words. {weather} is the live weather.\n")
	b.WriteString("> Lines starting with \">\" are notes.\n")
	for _, q := range d.Questions {
		topic := q.Topic
		if topic == "" {
			topic = strconv.Itoa(q.N) + "."
		}
		b.WriteString("\n## " + topic + "\n")
		for _, l := range q.Before {
			b.WriteString(gopodDraftReporterShort(l.Who) + ": " + l.Text + "\n")
		}
		b.WriteString(gopodDraftReporterShort(q.ReadBy) + " asks " + q.To + ": " + q.Ask + "\n")
		for _, a := range q.Answers {
			who := gopodDraftReporterShort(a.Brobot)
			if a.Cue != "" {
				who += " (" + a.Cue + ")"
			}
			b.WriteString(who + ": " + a.Say + "\n")
			if len(a.Moves) > 0 {
				var parts []string
				for _, m := range a.Moves {
					parts = append(parts, m.At+" = "+m.Move)
				}
				b.WriteString("moves: " + strings.Join(parts, "; ") + "\n")
			}
		}
	}
	return b.String()
}

// ## EOF

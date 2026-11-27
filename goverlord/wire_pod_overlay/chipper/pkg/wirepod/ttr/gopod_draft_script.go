// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_draft_script.go
package wirepod_ttr

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

const gopodDraftScriptName = "draft_script.md"

var (
	gopodDraftAskLine  = regexp.MustCompile(`^([A-Z][A-Z0-9 ]* reporter) asks (Brobot [12]|Brobots):\s*(.+)$`)
	gopodDraftSayLine  = regexp.MustCompile(`^(Brobot [12]|[A-Z][A-Z0-9 ]* reporter)(?: \(([^)]*)\))?:\s*(.+)$`)
	gopodDraftMoveLine = regexp.MustCompile(`^moves:\s*(.+)$`)
)

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
			q.ReadBy, q.To, q.Ask = m[1], m[2], strings.TrimSpace(m[3])
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
			if strings.HasSuffix(who, " reporter") {
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
			b.WriteString(l.Who + ": " + l.Text + "\n")
		}
		b.WriteString(q.ReadBy + " asks " + q.To + ": " + q.Ask + "\n")
		for _, a := range q.Answers {
			who := a.Brobot
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

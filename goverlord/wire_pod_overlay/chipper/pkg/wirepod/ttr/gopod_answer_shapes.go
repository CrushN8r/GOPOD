// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_answer_shapes.go
package wirepod_ttr

import (
	"strings"
	"sync"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// GOLDEN answer shapes (2026-10-23, FINISH_PLAN_001.md step A2; operator: "the
// flow of output in the responses isn't consistent * I need a templated
// answer-chain mechanism"): one shape per answer type - the order of his thought
// bubbles - in answer_shapes (settings). Each answer is checked against its
// type's shape as it's made (BROBOT_SHAPE in the log: ok, or the drift), and a
// dry test runs every type through the real answer chain at deploy
// (TestAnswerShapesHold). A shape is a list of marks: a mark's text
// (answer_shapes.marks) labels a bubble; "llm" is his own words; "?" = may be
// missing; "a|b" = either. In a line to both, the first answer's "reminder" is the
// hand-over instead (math keeps its reminder, then hands over).

// gopodAnswerType: the answer's type, for its shape.
func gopodAnswerType(esn, question string) string {
	lower := strings.ToLower(question)
	switch {
	case gopodIsThanks(question):
		return "thanks"
	case len(gopodSummaryItemsFor(esn)) > 0:
		return "summary"
	case gopodTopicGroupFor(esn) == "brobots_math":
		return "math"
	case gopodHasWholeWords(lower, "weather") || gopodNativeTakeIs(gopodBrobotName(esn), "weather", question):
		if r, ok := gopodDuoRule(esn); ok && r.WeatherType != "" {
			return r.WeatherType
		}
		return "weather"
	case len(gopodGivenPoints(esn)) > 0:
		return "topic"
	}
	return "free"
}

// gopodAnswerShape: the shape this answer should have.
func gopodAnswerShape(esn, question string, first bool) []string {
	typ := gopodAnswerType(esn, question)
	shape := append([]string(nil), gopodBrobotsSettings.AnswerShapes.Shapes[typ]...)
	// A golden question's exact line may open any answer.
	if len(shape) > 0 {
		shape = append([]string{"golden?", "snippet?"}, shape...)
		shape = append(shape, "snippet?") // the code's scripted end line
	}
	// Duo A (duo.a): its reminder is the hand-over (ends_with), or the hand-over
	// comes after it for a type that keeps its reminder (keep_reminder_types).
	if a, ok := gopodBrobotsSettings.Duo["a"]; first && ok && a.EndsWith != "" {
		if gopodContains(a.KeepReminderTypes, typ) || !a.NoReminder {
			shape = append(shape, a.EndsWith)
		} else {
			for i, m := range shape {
				if m == "reminder" {
					shape[i] = a.EndsWith
				}
			}
		}
	}
	return shape
}

// gopodShapeLabels: each bubble of a made answer, labelled by its mark.
func gopodShapeLabels(esn, answer string) []string {
	title, _ := gopodPointBubblesNotes(esn)
	say := gopodSayLine(esn)
	body := answer
	if loc := gopodAnimationMarkerPattern.FindStringIndex(body); loc != nil && loc[0] == 0 {
		body = body[loc[1]:]
	}
	var labels []string
	for _, bubble := range strings.Split(body, "\n") {
		b := strings.TrimSpace(bubble)
		if b == "" {
			continue
		}
		label := "llm"
		// A scripted start line (scripted_snippets, the code's) - "snippet".
		if gopodIsSnippetLine(b) {
			labels = append(labels, "snippet")
			continue
		}
		// A point's exact line first (it may hold a mark's words: Doc's Rule).
		if say != "" && strings.Contains(b, strings.TrimSpace(say)) {
			labels = append(labels, "say_line")
			continue
		}
		for _, m := range gopodBrobotsSettings.AnswerShapes.MarkOrder {
			found := false
			for _, t := range strings.Split(gopodBrobotsSettings.AnswerShapes.Marks[m], "|") {
				if t = strings.TrimSpace(t); t != "" && strings.Contains(strings.ToLower(b), strings.ToLower(t)) {
					found = true
				}
			}
			if found {
				label = m
				break
			}
		}
		if label == "llm" && say != "" && strings.Contains(b, strings.TrimSpace(say)) {
			label = "say_line"
		}
		// A summary's bubbles name their core topics' titles: one "summary".
		if label == "llm" {
			for _, it := range gopodSummaryItemsFor(esn) {
				if strings.Contains(strings.ToLower(b), strings.ToLower(it.title)) || (it.note != "" && strings.Contains(strings.ToLower(b), strings.ToLower(it.note))) {
					label = "summary"
				}
			}
			if label == "summary" && len(labels) > 0 && labels[len(labels)-1] == "summary" {
				continue
			}
		}
		// The point's title line is the first that names it; its notes follow as his words.
		if label == "llm" && title != "" && !gopodHasLabel(labels, "point") && strings.Contains(strings.ToLower(b), strings.ToLower(title)) {
			label = "point"
		}
		// His own words run together: one "llm".
		if label == "llm" && len(labels) > 0 && labels[len(labels)-1] == "llm" {
			continue
		}
		labels = append(labels, label)
	}
	return labels
}

// gopodIsSnippetLine: the bubble is one of the code's scripted start lines.
func gopodIsSnippetLine(b string) bool {
	lower := strings.ToLower(b)
	for _, s := range gopodBrobotsSettings.ScriptedSnippets.Lines {
		if s.At != "" && s.At != "start" {
			continue
		}
		for _, l := range s.Lines {
			if l = strings.ToLower(strings.TrimSpace(l)); l != "" && strings.Contains(lower, l) {
				return true
			}
		}
	}
	if note := strings.ToLower(strings.TrimSpace(gopodBrobotsSettings.Bingo.NoRunNote)); note != "" && strings.Contains(note, strings.TrimSpace(lower)) {
		return true
	}
	return false
}

func gopodHasLabel(labels []string, l string) bool {
	for _, x := range labels {
		if x == l {
			return true
		}
	}
	return false
}

// gopodShapeMatches: labels follow shape ("?" may be missing, "a|b" either).
func gopodShapeMatches(shape, labels []string) bool {
	var match func(s, l int) bool
	match = func(s, l int) bool {
		if s == len(shape) {
			return l == len(labels)
		}
		mark, optional := shape[s], strings.HasSuffix(shape[s], "?")
		mark = strings.TrimSuffix(mark, "?")
		if optional && match(s+1, l) {
			return true
		}
		if l < len(labels) {
			for _, alt := range strings.Split(mark, "|") {
				if alt == labels[l] && match(s+1, l+1) {
					return true
				}
			}
		}
		return false
	}
	return match(0, 0)
}

// gopodAnswerShapeCheck logs whether a made answer has its type's shape.
func gopodAnswerShapeCheck(esn, question, answer string, first bool) bool {
	shape := gopodAnswerShape(esn, question, first)
	if len(shape) == 0 {
		return true
	}
	labels := gopodShapeLabels(esn, answer)
	// 2026-10-24 (link 5): with nothing left to point at, the reminder is optional.
	if _, none := gopodReminderNone.LoadAndDelete(esn); none {
		for i, s := range shape {
			if s == "reminder" {
				shape = append(append([]string{}, shape[:i]...), append([]string{"reminder?"}, shape[i+1:]...)...)
				break
			}
		}
	}
	ok := gopodShapeMatches(shape, labels)
	state := "ok"
	if !ok {
		state = "DRIFT want " + strings.Join(shape, ",")
	}
	logger.Println("BROBOT_SHAPE " + gopodBrobotName(esn) + " " + gopodAnswerType(esn, question) + ": " + strings.Join(labels, ",") + " - " + state)
	return ok
}

// gopodReminderNone: his answer's reminder part found nothing left to point at.
var gopodReminderNone sync.Map

// ## EOF

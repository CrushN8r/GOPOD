// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_answer_shapes.go
package wirepod_ttr

import (
	"strings"
	"sync"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// gopodAnswerType: the answer's type, for its shape.
func gopodAnswerType(esn, question string) string {
	lower := strings.ToLower(question)
	switch {
	case gopodIsThanks(question):
		return "thanks"
	case gopodIsIntro(question):
		return "intro"
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
	if len(shape) > 0 {
		shape = append([]string{"greeting?", "golden?", "snippet?"}, shape...)
		shape = append(shape, "snippet?") // the code's scripted end line
	}
	if rc := gopodBrobotsSettings.ResponseChains["duo"]; first && len(rc.Answers) > 0 && rc.Answers[0].EndsWith != "" {
		a := rc.Answers[0]
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
		if len(labels) == 0 && gopodGreetingLeadRE.MatchString(b) {
			labels = append(labels, "greeting")
			continue
		}
		if gopodIsSnippetLine(b) {
			labels = append(labels, "snippet")
			continue
		}
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
		if label == "llm" && len(labels) > 0 && labels[len(labels)-1] == "reminder" && gopodIsProberLine(b) {
			continue
		}
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
		if label == "llm" && title != "" && !gopodHasLabel(labels, "point") && strings.Contains(strings.ToLower(b), strings.ToLower(title)) {
			label = "point"
		}
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
	for _, e := range gopodExperienceList() {
		if note := strings.ToLower(strings.TrimSpace(e.NoRunNote)); note != "" && strings.Contains(note, strings.TrimSpace(lower)) {
			return true
		}
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

// gopodIsProberLine: the reminder's "Try asking: ..." line (value_points.reminder_prober_text).
func gopodIsProberLine(b string) bool {
	t := gopodBrobotsSettings.ValuePoints.ReminderProberText
	if at := strings.Index(t, "{prober}"); at >= 0 {
		t = t[:at]
	}
	t = strings.TrimSpace(t)
	b = strings.TrimSpace(b)
	if loc := gopodAnimationMarkerPattern.FindStringIndex(b); loc != nil && loc[0] < 20 {
		b = strings.TrimSpace(b[loc[1]:])
	}
	return t != "" && strings.HasPrefix(strings.TrimLeft(b, ". "), t)
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

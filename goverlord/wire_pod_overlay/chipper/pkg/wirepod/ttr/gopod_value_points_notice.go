// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_value_points_notice.go
package wirepod_ttr

import (
	"strconv"
	"strings"
	"sync"
)

// The value points notice (2026-10-22, operator: "make standard, the
// Notice-bubble inside a Brobot's bubble * new notice-bubble always shows the
// response's value points (what the llm is composing from, that the GOPOD reps
// (brobots) are hired to reveal) * in the same notice-bubble, if more info
// incoming add it in underneath, then \"Got it!\""): every live Brobot answer
// opens with a notice at the top of his bubble (the bubble notice template,
// gopodBubbleNotice): the value points his answer is built from, the extras
// (weather, search), then the thinking line once he has said it. To help
// prevent drift: what he was given is on screen, next to what he says.

// gopodNoticeWords: one notice kind's words (bubble_notices): title, then line.
type gopodNoticeWords struct {
	Title string `json:"title"`
	Line  string `json:"line"`
}

// gopodNoticeFill: a notice text with its {tokens} filled in.
func gopodNoticeFill(text string, fill map[string]string) string {
	for k, v := range fill {
		text = strings.ReplaceAll(text, "{"+k+"}", v)
	}
	return strings.TrimSpace(text)
}

// gopodNoticeShow posts one notice kind for brobot's bubble; an empty title = off.
func gopodNoticeShow(brobot string, w gopodNoticeWords, fill map[string]string) {
	if title := gopodNoticeFill(w.Title, fill); title != "" {
		gopodBubbleNotice(brobot, title, gopodNoticeFill(w.Line, fill))
	}
}

// The notice log, in order (2026-10-22): 1. "Got it!" as he says it, 2. the
// points or none, 3. a manager's part (the page puts it in the notice), 4.
// "Got it!" again when the floor comes back to him. The lines are made with the
// prompt (gopodNoticeSet), held until both they and his "Got it!" are in.
var gopodNotice = struct {
	sync.Mutex
	lines map[string][]string // esn -> lines, waiting to be posted
	asked map[string]string   // esn -> the line he's answering
	ready map[string]bool     // esn -> the lines are in (gopodValuePointsNoticePost)
	gotIt map[string]bool     // esn -> he has said "Got it!"
	open  map[string]bool     // esn -> posted, for "Got it!" again
}{lines: map[string][]string{}, asked: map[string]string{}, ready: map[string]bool{}, gotIt: map[string]bool{}, open: map[string]bool{}}

// gopodNoticeParts: is this part in value_points.parts?
func gopodNoticePart(part string) bool {
	for _, p := range gopodBrobotsSettings.BubbleNotices.ValuePoints.Parts {
		if p == part {
			return true
		}
	}
	return false
}

// gopodNoticePoint: one value point's line ({topic} its label, {n}, {id} topic#n, {words}).
func gopodNoticePoint(topic string, content []byte, given string) string {
	id, n, words := topic, "", ""
	if f := strings.Fields(given); len(f) == 2 {
		if k, err := strconv.Atoi(strings.SplitN(f[1], "/", 2)[0]); err == nil {
			n = strconv.Itoa(k)
			id = topic + "#" + n
			if _, points := gopodSplitPoints(string(content)); k >= 1 && k <= len(points) {
				words = gopodPointLine.ReplaceAllString(points[k-1], "")
			}
		}
	}
	if strings.TrimSpace(words) == "" {
		words = topic
	}
	label := strings.TrimSpace(gopodBrobotsSettings.InterviewChapters.TopicLabels[topic])
	if label == "" {
		label = topic
	}
	return gopodNoticeFill(gopodBrobotsSettings.BubbleNotices.ValuePoints.Points, map[string]string{"id": id, "topic": label, "n": n, "words": strings.TrimSpace(words)})
}

// gopodNoticeSet keeps this answer's lines, in value_points.parts order
// (gopodQuestionFactsFor); "got_it" is where his "Got it!" goes.
func gopodNoticeSet(esn, question string, parts map[string][]string) {
	vp := gopodBrobotsSettings.BubbleNotices.ValuePoints
	if len(parts["points"]) == 0 && len(parts["search"]) == 0 && !gopodIsThanks(question) {
		parts["none"] = []string{vp.None}
	}
	var keep []string
	content := false
	for _, part := range vp.Parts {
		if part == "got_it" {
			keep = append(keep, "\x00got_it")
			continue
		}
		for _, l := range parts[part] {
			if strings.TrimSpace(l) != "" {
				keep = append(keep, l)
				content = true
			}
		}
	}
	if !content {
		keep = nil
	}
	gopodNotice.Lock()
	gopodNotice.lines[esn] = keep
	gopodNotice.asked[esn] = question
	// A new answer: a "Got it!" from before (a song's) doesn't count.
	delete(gopodNotice.gotIt, esn)
	delete(gopodNotice.ready, esn)
	gopodNotice.Unlock()
}

// gopodNoticeTitle: the value points notice's title, "" when off or not a Brobot's.
func gopodNoticeTitle(esn string) string {
	brobot := gopodBrobotName(esn)
	if (brobot != "Brobot 1" && brobot != "Brobot 2") || gopodScriptPlaying.Load() {
		return ""
	}
	return strings.TrimSpace(gopodBrobotsSettings.BubbleNotices.ValuePoints.Title)
}

// gopodNoticeFlush posts the held lines once they and his "Got it!" are in.
// Call with gopodNotice locked; returns the lines to post.
func gopodNoticeFlushLocked(esn string) ([]string, string, bool) {
	if !gopodNotice.ready[esn] || !gopodNotice.gotIt[esn] {
		return nil, "", false
	}
	lines, asked := gopodNotice.lines[esn], gopodNotice.asked[esn]
	delete(gopodNotice.lines, esn)
	delete(gopodNotice.asked, esn)
	delete(gopodNotice.ready, esn)
	delete(gopodNotice.gotIt, esn)
	gopodNotice.open[esn] = len(lines) > 0
	return lines, asked, true
}

// gopodBubbleStep: this Brobot Bubble step is on (brobot_bubble.steps).
func gopodBubbleStep(step string) bool {
	for _, s := range gopodBrobotsSettings.BrobotBubble.Steps {
		if s == step {
			return true
		}
	}
	return false
}

// gopodNoticeFlushed: notice lines 1-2 posted, then line 3 - a manager steps in
// if his line asks for one (gopodManagerStepIn), while his answer is written.
func gopodNoticeFlushed(esn string, lines []string, asked string, ok bool) {
	if !ok {
		return
	}
	gopodNoticePostLines(esn, lines)
	if b := gopodBrobotName(esn); gopodBubbleStep("manager") && (b == "Brobot 1" || b == "Brobot 2") && !gopodScriptPlaying.Load() && strings.TrimSpace(asked) != "" {
		go gopodManagerStepIn(b, asked)
	}
}

func gopodNoticePostLines(esn string, lines []string) {
	title := gopodNoticeTitle(esn)
	if title == "" {
		return
	}
	vp := gopodBrobotsSettings.BubbleNotices.ValuePoints
	for _, l := range lines {
		if l == "\x00got_it" {
			l = gopodNoticeFill(vp.GotIt, map[string]string{"thinking_text": gopodThinkingLineSpoken()})
		}
		if l != "" {
			gopodBubbleNotice(gopodBrobotName(esn), title, l)
		}
	}
}

// gopodValuePointsNoticePost: this answer's lines are in (after its prompt).
func gopodValuePointsNoticePost(esn string) {
	gopodNotice.Lock()
	gopodNotice.ready[esn] = true
	lines, asked, ok := gopodNoticeFlushLocked(esn)
	gopodNotice.Unlock()
	gopodNoticeFlushed(esn, lines, asked, ok)
}

// gopodValuePointsNoticeGotIt: he has said "Got it!" (gopodSayThinkingLine).
func gopodValuePointsNoticeGotIt(esn string) {
	gopodNotice.Lock()
	gopodNotice.gotIt[esn] = true
	lines, asked, ok := gopodNoticeFlushLocked(esn)
	gopodNotice.Unlock()
	gopodNoticeFlushed(esn, lines, asked, ok)
}

// gopodNoticeGotItAgain: the floor is back to him after a manager's part
// (gopodThinkingProcess's end): got_it_again ("Thanks for your help, {helper}."),
// shown in his notice; returns what he says ("" = nothing).
func gopodNoticeGotItAgain(esn string, hadManager bool) string {
	vp := gopodBrobotsSettings.BubbleNotices.ValuePoints
	gopodNotice.Lock()
	open := gopodNotice.open[esn]
	delete(gopodNotice.open, esn)
	gopodNotice.Unlock()
	mode := strings.TrimSpace(vp.SecondGotIt)
	if !open || !hadManager || mode == "off" || mode == "" || !gopodBubbleStep("thanks") {
		return ""
	}
	host, _ := gopodManagerLastHost.Load().(string)
	helper := gopodNoticeHelper(host)
	l := gopodNoticeFill(vp.GotItAgain, map[string]string{"thinking_text": gopodThinkingLineSpoken(), "helper": helper})
	if l != "" {
		gopodBubbleNotice(gopodBrobotName(esn), gopodNoticeTitle(esn), l)
	}
	return l
}

// gopodNoticeHelper: {helper} for the manager who had the floor, from
// value_points.helper_from in order - name (his own, "PLAYHEAD"), brobot (his
// Brobot, "Brobot 4"); the first that has one.
func gopodNoticeHelper(host string) string {
	for _, from := range gopodBrobotsSettings.BubbleNotices.ValuePoints.HelperFrom {
		v := ""
		switch from {
		case "name":
			v = host
		case "brobot":
			v = gopodBrobotsSettings.Managers.Personas[host].Reporter
		}
		if v = strings.TrimSpace(v); v != "" {
			return v
		}
	}
	return strings.TrimSpace(host)
}

// ## EOF

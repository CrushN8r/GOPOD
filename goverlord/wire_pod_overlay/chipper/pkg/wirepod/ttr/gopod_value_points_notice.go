// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_value_points_notice.go
package wirepod_ttr

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

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

// The notice log, in order: 1.
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
	total := ""
	if f := strings.Fields(given); len(f) == 2 {
		if parts := strings.SplitN(f[1], "/", 2); len(parts) == 2 {
			total = parts[1]
		}
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
	shownN, shownTotal := n, total
	if k, err := strconv.Atoi(n); err == nil && len(gopodBrobotsSettings.ValuePoints.CorePoints) > 0 {
		if all, err := strconv.Atoi(total); err == nil {
			rank, count := 0, 0
			for i := 1; i <= all; i++ {
				if gopodCorePoint(topic, i) {
					count++
					if i <= k {
						rank = count
					}
				}
			}
			if count > 0 && rank > 0 {
				shownN, shownTotal = strconv.Itoa(rank), strconv.Itoa(count)
			}
		}
	}
	n, total = shownN, shownTotal
	if k, err := strconv.Atoi(n); err == nil {
		if title, notes := gopodPointNotes(topic, k); title != "" {
			vp := gopodBrobotsSettings.BubbleNotices.ValuePoints
			line := gopodNoticeFill(vp.Points, map[string]string{"id": id, "topic": label, "n": n, "total": total, "words": title})
			if vp.Notes != "" {
				return line + "\n" + gopodNoticeFill(vp.Notes, map[string]string{"notes": strings.Join(notes, vp.NotesJoin)})
			}
			for _, note := range notes {
				line += "\n" + gopodNoticeFill(vp.Note, map[string]string{"note": note})
			}
			return line
		}
	}
	return gopodNoticeFill(gopodBrobotsSettings.BubbleNotices.ValuePoints.Points, map[string]string{"id": id, "topic": label, "n": n, "total": total, "words": strings.TrimSpace(words)})
}

// One notice per Brobot bubble: an egg found for his line is a line of his Value Points notice, not
// a notice of its own.
var gopodNoticeEggs = struct {
	sync.Mutex
	lines map[string][]string // Brobot -> egg lines waiting
}{lines: map[string][]string{}}

// gopodNoticeEgg: an egg's line for this Brobot's notice; false when there's no live notice (the
// egg's own notice is used instead).
func gopodNoticeEgg(brobot, title, summary string) bool {
	vp := gopodBrobotsSettings.BubbleNotices.ValuePoints
	if strings.TrimSpace(vp.Title) == "" || strings.TrimSpace(vp.Egg) == "" || gopodScriptPlaying.Load() || !gopodNoticePart("egg") {
		return false
	}
	line := gopodNoticeFill(vp.Egg, map[string]string{"title": title, "summary": summary})
	GopodChatFeedAddTo(gopodBrobotsSettings.TopicPictures.Role, brobot, gopodEggMark+title)
	esn := gopodBrobotESN(brobot)
	gopodNotice.Lock()
	posted := gopodNotice.open[esn]
	gopodNotice.Unlock()
	if posted {
		gopodBubbleNotice(brobot, gopodNoticeTitle(esn), line)
		return true
	}
	gopodNoticeEggs.Lock()
	gopodNoticeEggs.lines[brobot] = append(gopodNoticeEggs.lines[brobot], line)
	gopodNoticeEggs.Unlock()
	return true
}

// The game lines: combo, chapter cleared, back on track, score.
var gopodChapterCleared = struct {
	sync.Mutex
	said map[string]bool
}{said: map[string]bool{}}

func gopodNoticeGame(esn, question string, parts map[string][]string) {
	vp := gopodBrobotsSettings.BubbleNotices.ValuePoints
	labelOf := func(topic string) string {
		if l := strings.TrimSpace(gopodBrobotsSettings.InterviewChapters.TopicLabels[topic]); l != "" {
			return l
		}
		return topic
	}
	var topics []string
	seen := map[string]bool{}
	for _, g := range gopodGivenPoints(esn) {
		if f := strings.Fields(g); len(f) == 2 && !seen[f[0]] {
			seen[f[0]] = true
			topics = append(topics, f[0])
		}
	}
	if len(topics) >= 2 && vp.Combo != "" {
		var names []string
		for _, t := range topics {
			names = append(names, labelOf(t))
		}
		parts["combo"] = []string{gopodNoticeFill(vp.Combo, map[string]string{"topics": strings.Join(names, " + ")})}
	}
	names, counts := gopodPackPointCounts()
	gopodCoverage.Lock()
	said, total := 0, 0
	byTopic := gopodBrobotsSettings.ValuePoints.Target == "topics"
	for _, n := range names {
		if byTopic {
			total++
			if gopodCoverage.revealed[n] > 0 {
				said++
			}
			continue
		}
		r := gopodCoverage.revealed[n]
		if r > counts[n] {
			r = counts[n]
		}
		said += r
		total += counts[n]
	}
	covered := func(t string) bool { return counts[t] > 0 && gopodTopicDoneLocked(t, counts[t]) }
	// 2.
	var cleared []string
	for _, c := range gopodBrobotsSettings.InterviewChapters.Chapters {
		touched, all := false, len(c.Topics) > 0
		for _, t := range c.Topics {
			if seen[t] {
				touched = true
			}
			if _, ok := counts[t]; ok && !covered(t) {
				all = false
			}
		}
		if touched && all {
			cleared = append(cleared, c.Title)
		}
	}
	gopodCoverage.Unlock()
	for _, title := range cleared {
		gopodChapterCleared.Lock()
		again := gopodChapterCleared.said[title]
		gopodChapterCleared.said[title] = true
		gopodChapterCleared.Unlock()
		if !again && vp.Chapter != "" {
			parts["chapter"] = append(parts["chapter"], gopodNoticeFill(vp.Chapter, map[string]string{"chapter": title}))
		}
	}
	if !gopodIsThanks(question) {
		if next := gopodReminderTopics(question); next != "" {
			if len(topics) == 0 && vp.Recovery != "" {
				parts["recovery"] = []string{gopodNoticeFill(vp.Recovery, map[string]string{"topics": next})}
			} else if len(topics) > 0 && vp.Next != "" {
				parts["next"] = []string{gopodNoticeFill(vp.Next, map[string]string{"topics": next})}
			}
		}
	}
	if vp.Score != "" && total > 0 {
		score, _ := gopodCoverageScoreText()
		filled, _ := gopodCoverageScoreWith(vp.Score)
		parts["score"] = []string{gopodNoticeFill(filled, map[string]string{"said": strconv.Itoa(said), "total": strconv.Itoa(total), "score": score})}
	}
}

// gopodBrainScoreRows: the game's score rows (value_points.score - topics, value points, eggs, each
// with its bar), for every brain bubble, a Brobot's or a Gomad's.
var gopodBrainScoreShown atomic.Value

func gopodBrainScoreRows() []string {
	vp := gopodBrobotsSettings.BubbleNotices.ValuePoints
	if strings.TrimSpace(vp.Score) == "" {
		return nil
	}
	filled, _ := gopodCoverageScoreWith(vp.Score)
	score, _ := gopodCoverageScoreText()
	if prev, _ := gopodBrainScoreShown.Swap(score).(string); prev == score {
		return nil
	}
	var out []string
	for _, row := range strings.Split(gopodNoticeFill(filled, map[string]string{"score": score}), "\n") {
		if strings.TrimSpace(row) != "" {
			out = append(out, row)
		}
	}
	return out
}

// gopodNoticeSet keeps this answer's lines, in value_points.parts order (gopodQuestionFactsFor);
// "got_it" is where his "Got it!" goes.
func gopodNoticeSet(esn, question string, parts map[string][]string) {
	vp := gopodBrobotsSettings.BubbleNotices.ValuePoints
	intro := gopodIsIntro(question)
	if len(parts["points"]) == 0 && len(parts["search"]) == 0 && len(parts["summary"]) == 0 && !gopodIsThanks(question) && !intro {
		parts["none"] = []string{vp.None}
	}
	gopodNoticeEggs.Lock()
	parts["egg"] = append(parts["egg"], gopodNoticeEggs.lines[gopodBrobotName(esn)]...)
	delete(gopodNoticeEggs.lines, gopodBrobotName(esn))
	gopodNoticeEggs.Unlock()
	if !intro {
		gopodNoticeGame(esn, question, parts)
	}
	gopodBrainReminders.Delete(gopodBrobotName(esn))
	if v, ok := gopodReminderNames.LoadAndDelete(question); ok && !intro && !gopodIsThanks(question) {
		gopodBrainReminders.Store(gopodBrobotName(esn), v)
	}
	parts["memory"] = gopodBrainBubbleLines(gopodBrobotName(esn), question)
	parts["score"] = gopodBrainScoreRows()
	var keep []string
	content := false
	for _, part := range vp.Parts {
		if part == "got_it" {
			keep = append(keep, "\x00got_it")
			continue
		}
		if part == "plan" {
			if l := gopodResponsePlan(esn); l != "" {
				keep = append(keep, l)
			}
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
	gopodNoticeGuideSet(esn, parts)
	gopodNotice.Lock()
	gopodNotice.lines[esn] = keep
	gopodNotice.asked[esn] = question
	delete(gopodNotice.gotIt, esn)
	delete(gopodNotice.ready, esn)
	gopodNotice.Unlock()
}

// The notice guide: each notice line this answer shows, with what he's to do about it
// (bubble_notices.value_points.guide.<part>, {line} the notice's own line), in value_points.parts
// order - his pre-build context's {notice} (one source: what the room sees is what he's told).
var gopodNoticeGuides = struct {
	sync.Mutex
	lines map[string][]string
}{lines: map[string][]string{}}

func gopodNoticeGuideSet(esn string, parts map[string][]string) {
	vp := gopodBrobotsSettings.BubbleNotices.ValuePoints
	var out []string
	for _, part := range vp.Parts {
		g := strings.TrimSpace(vp.Guide[part])
		if g == "" {
			continue
		}
		for _, l := range parts[part] {
			if l = strings.TrimSpace(strings.ReplaceAll(l, "\n", " ")); l != "" {
				out = append(out, strings.ReplaceAll(g, "{line}", strings.TrimRight(l, ":")))
			}
		}
	}
	gopodNoticeGuides.Lock()
	gopodNoticeGuides.lines[esn] = out
	gopodNoticeGuides.Unlock()
}

// gopodNoticeGuide: this answer's notice guide lines, one per line.
func gopodNoticeGuide(esn string) string {
	gopodNoticeGuides.Lock()
	defer gopodNoticeGuides.Unlock()
	return strings.Join(gopodNoticeGuides.lines[esn], "\n")
}

// gopodNoticeTitle: the value points notice's title, "" when off or not a Brobot's.
func gopodNoticeTitle(esn string) string {
	brobot := gopodBrobotName(esn)
	if (brobot != "Brobot 1" && brobot != "Brobot 2") || gopodScriptPlaying.Load() {
		return ""
	}
	return strings.TrimSpace(strings.ReplaceAll(gopodBrobotsSettings.BubbleNotices.ValuePoints.Title, "{name}", gopodPersonaShown(brobot)))
}

// gopodNoticeFlush posts the held lines once they and his "Got it!" are in.
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

const gopodNoticeSayMark = "noticesay:"
const gopodEggMark = "eggmark:"
const gopodNoticeSaidMark = "noticesaid:"

// gopodNoticeSaid: his spoken notice line has been said.
func gopodNoticeSaid(esn string) {
	GopodChatFeedAddTo(gopodBrobotsSettings.TopicPictures.Role, gopodBrobotName(esn), gopodNoticeSaidMark)
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

// gopodNoticeFlushed: notice lines 1-2 posted, then line 3 - a manager steps in if his line asks
// for one (gopodManagerStepIn), while his answer is written.
func gopodNoticeFlushed(esn string, lines []string, asked string, ok bool) {
	if !ok {
		return
	}
	gopodNoticePostLines(esn, lines)
	if r, ok := gopodDuoRule(esn); ok && r.NoManager {
		if intern := gopodWeatherInternFor(asked); intern != "" && !gopodScriptPlaying.Load() {
			go gopodWeatherInternReady(intern, gopodBrobotName(esn))
		}
		return
	}
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
			l = gopodNoticeFill(vp.GotIt, map[string]string{"thinking_text": gopodOpenerText(esn)})
		}
		for _, one := range strings.Split(l, "\n") {
			if one = strings.TrimSpace(one); one != "" {
				gopodBubbleNotice(gopodBrobotName(esn), title, one)
			}
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

// gopodNoticeGotItAgain: the floor is back to him after a manager's part (gopodThinkingProcess's
// end): got_it_again ("Thanks for your help, {helper}."), shown in his notice; returns what he says
// ("" = nothing).
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
	l := gopodNoticeFill(vp.GotItAgain, map[string]string{"thinking_text": gopodGotItFor(esn).Text, "helper": helper})
	if l != "" {
		GopodChatFeedAddTo(gopodBrobotsSettings.TopicPictures.Role, gopodBrobotName(esn),
			gopodNoticeSayMark+strings.ReplaceAll(gopodNoticeTitle(esn), "|", " ")+"|"+strings.ReplaceAll(l, "|", " "))
	}
	return l
}

// gopodNoticeHelper: {helper} for the manager who had the floor, from value_points.helper_from in
// order - name (his own, "PLAYHEAD"), brobot (his Brobot, "Brobot 4"); the first that has one.
func gopodNoticeHelper(host string) string {
	for _, from := range gopodBrobotsSettings.BubbleNotices.ValuePoints.HelperFrom {
		v := ""
		switch from {
		case "call_name":
			v = gopodCallName(host)
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

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
	// 2026-10-23 (live: "a day of parties #1 of 3" with only #1 a value point;
	// operator: "where's: a day of parties #2 of 3"): #N of N counts the topic's
	// value points (core_points) - its place among them, of how many.
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
	// 2026-10-22 (operator: note form): a point with notes shows its title, then
	// each note on its own line (value_points.note).
	if k, err := strconv.Atoi(n); err == nil {
		if title, notes := gopodPointNotes(topic, k); title != "" {
			vp := gopodBrobotsSettings.BubbleNotices.ValuePoints
			line := gopodNoticeFill(vp.Points, map[string]string{"id": id, "topic": label, "n": n, "total": total, "words": title})
			for _, note := range notes {
				line += "\n" + gopodNoticeFill(vp.Note, map[string]string{"note": note})
			}
			return line
		}
	}
	return gopodNoticeFill(gopodBrobotsSettings.BubbleNotices.ValuePoints.Points, map[string]string{"id": id, "topic": label, "n": n, "total": total, "words": strings.TrimSpace(words)})
}

// One notice per Brobot bubble (2026-10-22, operator: "brobot 2 response isn't
// same structure as brobot 1 response"): an egg found for his line is a line of
// his Value Points notice, not a notice of its own. Held until his notice is
// made (gopodNoticeSet); after it's posted, added to it at once.
var gopodNoticeEggs = struct {
	sync.Mutex
	lines map[string][]string // Brobot -> egg lines waiting
}{lines: map[string][]string{}}

// gopodNoticeEgg: an egg's line for this Brobot's notice; false when there's
// no live notice (the egg's own notice is used instead).
func gopodNoticeEgg(brobot, title, summary string) bool {
	vp := gopodBrobotsSettings.BubbleNotices.ValuePoints
	if strings.TrimSpace(vp.Title) == "" || strings.TrimSpace(vp.Egg) == "" || gopodScriptPlaying.Load() || !gopodNoticePart("egg") {
		return false
	}
	line := gopodNoticeFill(vp.Egg, map[string]string{"title": title, "summary": summary})
	// 2026-10-25 (operator: "NEW RULE: The brobot bubble indicates revealed easter
	// eggs upon 1st reveal * no more pop ups * however, the marquee lights stay, as
	// an indicator of a revealed easter egg"): his bubble wears the marquee, every
	// time a line of his hits an egg (the page's eggMarked).
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

// The game lines (2026-10-22, operator: "Game survey: 1-4";
// GAME_FEEL_SURVEY_001.md): combo, chapter cleared, back on track, score.
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
	// 3. combo: one answer, points from two topics or more.
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
			// 2026-10-23: the game counts topics - one point opens one.
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
	// 2. chapter cleared: every topic of a chapter this answer touched is covered.
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
	// 4. the gravitational pull (2026-10-22, operator: "remember the
	// gravitational pull to reveal all value points * please template this
	// somewhere in the flow, kinda like Back on track"): every answer names the
	// next points on the list - a free answer "Back on track", a topic answer
	// "Next up" (value_points.recovery / next; the gravity block).
	if !gopodIsThanks(question) {
		if next := gopodReminderTopics(question); next != "" {
			if len(topics) == 0 && vp.Recovery != "" {
				parts["recovery"] = []string{gopodNoticeFill(vp.Recovery, map[string]string{"topics": next})}
			} else if len(topics) > 0 && vp.Next != "" {
				parts["next"] = []string{gopodNoticeFill(vp.Next, map[string]string{"topics": next})}
			}
		}
	}
	// 1. score: the one score template (value_points.score_text, A3) - topics,
	// value points and eggs, N of N.
	if vp.Score != "" && total > 0 {
		score, _ := gopodCoverageScoreText()
		// 2026-10-25: the notice's own score lines (its bars) - the score's numbers
		// filled straight into value_points.score here (gopodCoverageScoreWith).
		filled, _ := gopodCoverageScoreWith(vp.Score)
		parts["score"] = []string{gopodNoticeFill(filled, map[string]string{"said": strconv.Itoa(said), "total": strconv.Itoa(total), "score": score})}
	}
}

// gopodNoticeSet keeps this answer's lines, in value_points.parts order
// (gopodQuestionFactsFor); "got_it" is where his "Got it!" goes.
func gopodNoticeSet(esn, question string, parts map[string][]string) {
	vp := gopodBrobotsSettings.BubbleNotices.ValuePoints
	if len(parts["points"]) == 0 && len(parts["search"]) == 0 && len(parts["summary"]) == 0 && !gopodIsThanks(question) {
		parts["none"] = []string{vp.None}
	}
	gopodNoticeEggs.Lock()
	parts["egg"] = append(parts["egg"], gopodNoticeEggs.lines[gopodBrobotName(esn)]...)
	delete(gopodNoticeEggs.lines, gopodBrobotName(esn))
	gopodNoticeEggs.Unlock()
	gopodNoticeGame(esn, question, parts)
	var keep []string
	content := false
	for _, part := range vp.Parts {
		if part == "got_it" {
			keep = append(keep, "\x00got_it")
			continue
		}
		// 2026-10-25: the response-chain's plan line (gopodResponsePlan).
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

// gopodNoticeFlushed: notice lines 1-2 posted, then line 3 - a manager steps in
// if his line asks for one (gopodManagerStepIn), while his answer is written.
func gopodNoticeFlushed(esn string, lines []string, asked string, ok bool) {
	if !ok {
		return
	}
	gopodNoticePostLines(esn, lines)
	if r, ok := gopodDuoRule(esn); ok && r.NoManager {
		return // duo B: his brother's manager already helped (duo.b.no_manager)
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
	l := gopodNoticeFill(vp.GotItAgain, map[string]string{"thinking_text": gopodGotItFor(esn).Text, "helper": helper})
	if l != "" {
		// 2026-10-23 (operator: "'Thanks for your help, PLAYHEAD.' not in
		// highlighting pipeline"): he says it - lit as he does (noticesay:), shown
		// said once he's done (gopodNoticeSaid).
		GopodChatFeedAddTo(gopodBrobotsSettings.TopicPictures.Role, gopodBrobotName(esn),
			gopodNoticeSayMark+strings.ReplaceAll(gopodNoticeTitle(esn), "|", " ")+"|"+strings.ReplaceAll(l, "|", " "))
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

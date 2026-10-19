// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_value_points.go
package wirepod_ttr

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// Value-point coverage (2026-10-06, operator: "My two brobots are party-bros as
// well as expert level GOPOD reps. For topics/chapters ... the brobots keep a
// list of, has this topic been previously discussed, what value points were
// revealed and do they remind me of other value points to cover. End goal ...
// cover all or as many value points in all topics, per interview session").
// The code keeps the list, not the LLM, and the LLM is given less:
//   - A topic pack's numbered lines ("1. ...") are its value points; a pack
//     with none is one point. Its other lines are its opening.
//   - When a topic loads, the LLM gets the opening and only the next point not
//     yet revealed; it's ticked off then (both Brobots share one list). Once all
//     are revealed the topic is covered, and asked again it gets the whole pack.
//   - "Hey, that reminds me..." (a scripted_snippets entry with when "reminder")
//     points to the next topic, in the interview's chapter order, that isn't
//     covered yet and has a line in value_points.reminders - the operator's
//     words.
//   - At "Thanks Brobots" the score shows in a Goverlord bubble (value_points.
//     score_text) and is logged (GOPOD_COVERAGE), with the list.
// One interview session is one Wire-Pod start: the list starts empty.

var gopodPointLine = regexp.MustCompile(`^\s*\d+\.\s+`)

// gopodSplitPoints splits a pack into its opening and its numbered points.
func gopodSplitPoints(content string) (string, []string) {
	var opening, points []string
	for _, line := range strings.Split(strings.TrimRight(content, "\n"), "\n") {
		if gopodPointLine.MatchString(line) {
			points = append(points, line)
		} else {
			opening = append(opening, line)
		}
	}
	return strings.Join(opening, "\n"), points
}

var gopodCoverage = struct {
	sync.Mutex
	revealed map[string]int  // topic -> points revealed
	scored   bool            // this ending's score has been shown
	reminded map[string]bool // topics and chapters a reminder has named (2026-10-07: not again)
}{revealed: map[string]int{}, reminded: map[string]bool{}}

func gopodReminded(name string) bool {
	gopodCoverage.Lock()
	defer gopodCoverage.Unlock()
	return gopodCoverage.reminded[name]
}

func gopodRemind(names ...string) {
	gopodCoverage.Lock()
	defer gopodCoverage.Unlock()
	if gopodCoverage.reminded == nil {
		gopodCoverage.reminded = map[string]bool{}
	}
	for _, n := range names {
		gopodCoverage.reminded[n] = true
	}
}

// gopodCoverageFacts is what the LLM gets for one loaded topic: its opening and
// the next point (marked revealed), or the whole pack once covered.
func gopodCoverageFacts(name string, content []byte) []byte {
	facts, _ := gopodCoverageReveal(name, content)
	return facts
}

// gopodCoverageReveal is gopodCoverageFacts, plus the point it gave out
// ("doc_and_pip 1/3"), "" when none.
func gopodCoverageReveal(name string, content []byte) ([]byte, string) {
	if name == gopodBrobotsSettings.ValuePoints.WhatElseTopic {
		return gopodWhatElseFacts(content), ""
	}
	opening, points := gopodSplitPoints(string(content))
	gopodCoverage.Lock()
	defer gopodCoverage.Unlock()
	total := len(points)
	if total == 0 {
		total = 1
	}
	n := gopodCoverage.revealed[name]
	if n >= total {
		logger.Println("BROBOT_VALUE_POINT topic=" + name + " covered")
		return content, ""
	}
	gopodCoverage.revealed[name] = n + 1
	given := fmt.Sprintf("%s %d/%d", name, n+1, total)
	logger.Println(fmt.Sprintf("BROBOT_VALUE_POINT topic=%s point=%d/%d", name, n+1, total))
	if len(points) == 0 {
		return content, given
	}
	return []byte(opening + "\n" + points[n] + "\n"), given
}

// The points given out for each Brobot's current question ("doc_and_pip 1/3"),
// for its say line and its take (gopod_take_sheet.go).
var gopodGiven = struct {
	sync.Mutex
	points map[string][]string // esn -> points
}{points: map[string][]string{}}

func gopodGivenSet(esn string, points []string) {
	gopodGiven.Lock()
	defer gopodGiven.Unlock()
	gopodGiven.points[esn] = points
}

func gopodGivenPoints(esn string) []string {
	gopodGiven.Lock()
	defer gopodGiven.Unlock()
	return append([]string(nil), gopodGiven.points[esn]...)
}

// gopodSayLine is the exact line for a point just given out to this Brobot
// (value_points.say_lines, keyed "topic N"; operator 2026-10-06: "From the take
// sheet? what is hard for the llm to comply with, make a scripted snippet" -
// names, numbers, gear and his own phrasing). Said word for word at the start of
// the reply, after the cue; the LLM adds the rest. "" when none.
func gopodSayLine(esn string) string {
	for _, given := range gopodGivenPoints(esn) {
		key := given
		if i := strings.LastIndex(key, "/"); i >= 0 {
			key = key[:i]
		}
		if line := gopodBrobotsSettings.ValuePoints.SayLines[key]; line != "" {
			return line
		}
	}
	return ""
}

// gopodTopicCovered: all of a topic's points have been revealed.
func gopodTopicCovered(name string, total int) bool {
	gopodCoverage.Lock()
	defer gopodCoverage.Unlock()
	return gopodCoverage.revealed[name] >= total
}

// gopodPackPointCounts returns every pack's number of points, in registry order.
func gopodPackPointCounts() ([]string, map[string]int) {
	var names []string
	counts := map[string]int{}
	for _, p := range gopodTopicRegistry() {
		data, err := gopodReadTopicFacts(p.File)
		if err != nil {
			continue
		}
		if p.Name == gopodBrobotsSettings.ValuePoints.WhatElseTopic {
			continue // the question itself, not a topic
		}
		_, points := gopodSplitPoints(string(data))
		counts[p.Name] = len(points)
		if counts[p.Name] == 0 {
			counts[p.Name] = 1
		}
		names = append(names, p.Name)
	}
	return names, counts
}

// gopodReminderLine is the "Hey, that reminds me..." line after a reply on the
// question's topics; "" when there's none. It never says a value point
// (operator 2026-10-07: "do not reveal any value points" - the old per-topic
// lines were points, and gave the next answer away). It points:
//   - after a chapter's last topic, to the next two chapters with topics left,
//     each with its summary (interview_chapters.summary, reminder_chapters_text);
//   - otherwise to the reminder_topic_count topics left that are most alike,
//     chained - the closest to this topic, then the closest to that one (the
//     "3d grid": how alike two topics' packs are, worked out by the code) -
//     named by interview_chapters.topic_labels (reminder_topics_text).
//
// Neither points at the next reminder_skip_next_topics topics: the interview is
// about to ask those.
func gopodReminderLine(userMessage string) string {
	vp, ic := gopodBrobotsSettings.ValuePoints, gopodBrobotsSettings.InterviewChapters
	var order []string
	chapterAt := map[string]int{}
	lastOfChapter := map[string]bool{}
	for ci, c := range ic.Chapters {
		order = append(order, c.Topics...)
		for _, t := range c.Topics {
			chapterAt[t] = ci
		}
		if len(c.Topics) > 0 {
			lastOfChapter[c.Topics[len(c.Topics)-1]] = true
		}
	}
	loaded := map[string]bool{}
	for _, p := range loadScenarioPacketsForMessage(userMessage) {
		loaded[p.Name] = true
	}
	last := -1
	for i, t := range order {
		if loaded[t] {
			last = i
		}
	}
	// 2026-10-17 (the egg list, gopod_egg_list.go): with no topic in the
	// question, the reply still ends by naming eggs still to find.
	eggWant := vp.ReminderEggCount
	if eggWant <= 0 {
		eggWant = 2
	}
	if last < 0 {
		if gopodIsThanks(userMessage) || vp.ReminderTopicsText == "" {
			return ""
		}
		var names []string
		for _, e := range gopodEggsToMention(eggWant, loaded) {
			names = append(names, e.Label)
		}
		if len(names) == 0 {
			return ""
		}
		return strings.ReplaceAll(vp.ReminderTopicsText, "{topics}", gopodNameList(names))
	}
	_, counts := gopodPackPointCounts()
	// A topic whose egg was found is off the list too.
	eggTopic := map[string]bool{}
	for _, e := range gopodEggList() {
		eggTopic[e.Key] = true
	}
	open := func(t string) bool {
		return !loaded[t] && !gopodTopicCovered(t, counts[t]) && !(eggTopic[t] && gopodEggFound(t))
	}
	skip := map[string]bool{}
	for k := 1; k <= vp.ReminderSkip && last+k < len(order); k++ {
		skip[order[last+k]] = true
	}
	current := order[last]
	if lastOfChapter[current] && vp.ReminderChaptersText != "" {
		// The next chapter with its summary, then the one after by title only
		// (2026-10-07: two summaries were long out loud); a chapter once named
		// isn't named again.
		var next, named []string
		for ci := chapterAt[current] + 1; ci < len(ic.Chapters) && len(next) < 2; ci++ {
			c := ic.Chapters[ci]
			if gopodReminded("chapter:" + c.Title) {
				continue
			}
			for _, t := range c.Topics {
				if open(t) {
					line := c.Title
					if len(next) == 0 && c.Summary != "" {
						line += " - " + strings.TrimRight(c.Summary, ".")
					}
					next = append(next, line)
					named = append(named, "chapter:"+c.Title)
					break
				}
			}
		}
		if len(next) > 0 {
			gopodRemind(named...)
			return strings.ReplaceAll(vp.ReminderChaptersText, "{chapters}", strings.Join(next, ". Then ")+".")
		}
	}
	if vp.ReminderTopicsText == "" {
		return ""
	}
	want := vp.ReminderTopicCount
	if want <= 0 {
		want = 3
	}
	used := map[string]bool{current: true}
	var chain []string
	for from := current; len(chain) < want; {
		best, bestSim := "", -1.0
		for _, t := range order {
			if used[t] || skip[t] || !open(t) || ic.TopicLabels[t] == "" || gopodReminded(t) {
				continue
			}
			if sim := gopodTopicLikeness(from, t); sim > bestSim {
				best, bestSim = t, sim
			}
		}
		if best == "" {
			break
		}
		used[best] = true
		chain = append(chain, ic.TopicLabels[best])
		from = best
	}
	// The egg list: the reminder names at least one egg still to find - in
	// place of the chain's last topic when none of them is one.
	hasEgg := false
	for t := range used {
		if t != current && eggTopic[t] {
			hasEgg = true
		}
	}
	if !hasEgg {
		if eggs := gopodEggsToMention(1, used); len(eggs) > 0 {
			if len(chain) >= want && len(chain) > 0 {
				chain = chain[:len(chain)-1]
			}
			chain = append(chain, eggs[0].Label)
		}
	}
	if len(chain) == 0 {
		return ""
	}
	for t := range used {
		if t != current {
			gopodRemind(t)
		}
	}
	return strings.ReplaceAll(vp.ReminderTopicsText, "{topics}", gopodNameList(chain))
}

// The 3d grid: how alike two topics are - the share of their packs' words
// they have in common. Worked out once, by the code, not by the LLM.
var gopodTopicWords = struct {
	sync.Once
	words map[string]map[string]bool
}{}

var gopodLikenessWord = regexp.MustCompile(`[a-z][a-z0-9']{3,}`)

func gopodTopicLikeness(a, b string) float64 {
	gopodTopicWords.Do(func() {
		gopodTopicWords.words = map[string]map[string]bool{}
		for _, p := range gopodTopicRegistry() {
			data, err := gopodReadTopicFacts(p.File)
			if err != nil {
				continue
			}
			set := map[string]bool{}
			for _, w := range gopodLikenessWord.FindAllString(strings.ToLower(string(data)), -1) {
				set[w] = true
			}
			gopodTopicWords.words[p.Name] = set
		}
	})
	wa, wb := gopodTopicWords.words[a], gopodTopicWords.words[b]
	if len(wa) == 0 || len(wb) == 0 {
		return 0
	}
	shared := 0
	for w := range wa {
		if wb[w] {
			shared++
		}
	}
	return float64(shared) / float64(len(wa)+len(wb)-shared)
}

// gopodCoverageScoreText is the session score and the list behind it.
func gopodCoverageScoreText() (string, string) {
	names, counts := gopodPackPointCounts()
	gopodCoverage.Lock()
	defer gopodCoverage.Unlock()
	points, total, topics := 0, 0, 0
	var detail []string
	for _, name := range names {
		n := gopodCoverage.revealed[name]
		if n > counts[name] {
			n = counts[name]
		}
		points += n
		total += counts[name]
		if n > 0 {
			topics++
		}
		detail = append(detail, fmt.Sprintf("%s=%d/%d", name, n, counts[name]))
	}
	sort.Strings(detail)
	text := strings.NewReplacer(
		"{points}", fmt.Sprint(points), "{points_total}", fmt.Sprint(total),
		"{topics}", fmt.Sprint(topics), "{topics_total}", fmt.Sprint(len(names)),
	).Replace(gopodBrobotsSettings.ValuePoints.ScoreText)
	return text, strings.Join(detail, " ")
}

// gopodCoverageScore shows the score once per ending (a new question that isn't
// a thanks allows the next one, gopodCoverageNewQuestion).
func gopodCoverageScore() {
	vp := gopodBrobotsSettings.ValuePoints
	gopodCoverage.Lock()
	if gopodCoverage.scored {
		gopodCoverage.Unlock()
		return
	}
	gopodCoverage.scored = true
	gopodCoverage.Unlock()
	text, detail := gopodCoverageScoreText()
	logger.Println("GOPOD_COVERAGE " + text + " | " + detail)
	// 2026-10-08: score_role "" keeps the score in the log only (operator: "remove:
	// 'Covered 0 of 78 value points in 0 of 38 topics' offset the scrolling").
	if vp.ScoreRole != "" {
		GopodChatFeedAdd(vp.ScoreRole, text)
	}
}

func gopodCoverageNewQuestion() {
	gopodCoverage.Lock()
	gopodCoverage.scored = false
	gopodCoverage.Unlock()
}

// gopodWhatElseFacts answers "what else can we talk about?" (operator
// 2026-10-06: "the brobot GOPOD pros know ... refer to the 'topics & value
// points covered' list"): the next chapters, in interview order, with points not
// covered yet, and the score so far.
func gopodWhatElseFacts(content []byte) []byte {
	_, counts := gopodPackPointCounts()
	want := gopodBrobotsSettings.ValuePoints.WhatElseCount
	if want <= 0 {
		want = 3
	}
	var left []string
	for _, c := range gopodBrobotsSettings.InterviewChapters.Chapters {
		for _, t := range c.Topics {
			if !gopodTopicCovered(t, counts[t]) {
				// 2026-10-07: with the chapter's summary - the Brobots' map of what's left.
				line := "- " + c.Title
				if c.Summary != "" {
					line += ": " + c.Summary
				}
				left = append(left, line)
				break
			}
		}
		if len(left) == want {
			break
		}
	}
	score, _ := gopodCoverageScoreText()
	logger.Println("BROBOT_WHAT_ELSE " + strings.Join(left, " "))
	return []byte(strings.TrimRight(string(content), "\n") + "\n" + strings.Join(left, "\n") + "\n(" + score + " so far.)\n")
}

// The scripted lines are the code's, never the LLM's (operator 2026-10-06: "Fix
// 'Hey, that reminds me...'" - in the harvest, 59 of 83 takes said a reminder:
// the code adds one at most half the time, the LLM copied the rest from its own
// earlier replies in the chat memory). So every flavour line and reminder line
// is dropped from the remembered replies, and a "Hey, that reminds me..." the
// LLM writes itself is cut from its reply, to the end - that phrase is the code's.
var gopodReminderPhrase = regexp.MustCompile(`(?is)\s*\bhey\W+that\s+reminds\s+me\b.*$`)

var gopodScriptedLinePatterns = func() []*regexp.Regexp {
	var lines []string
	for _, s := range gopodBrobotsSettings.ScriptedSnippets.Lines {
		lines = append(lines, s.Lines...)
	}
	for _, line := range gopodBrobotsSettings.ValuePoints.Reminders {
		lines = append(lines, line)
	}
	// 2026-10-06: and the hard points' say lines - one Brobot repeated the
	// other's (TIGHTEN_RESPONSES_FROM_TAKES_001.md).
	for _, line := range gopodBrobotsSettings.ValuePoints.SayLines {
		lines = append(lines, line)
	}
	words := regexp.MustCompile(`[\pL\pN']+`)
	var out []*regexp.Regexp
	for _, line := range lines {
		w := words.FindAllString(line, -1)
		if len(w) < 2 {
			continue
		}
		for i := range w {
			w[i] = regexp.QuoteMeta(w[i])
		}
		// The words in order, any punctuation or spacing between and after
		// (the reply's tidy-up turns "..." into ", ").
		out = append(out, regexp.MustCompile(`(?i)\b`+strings.Join(w, `[^\pL\pN']+`)+`[^\pL\pN\s]*`))
	}
	return out
}()

// gopodDropScriptedLines takes the scripted lines out of a remembered reply.
func gopodDropScriptedLines(text string) string {
	text = gopodReminderPhrase.ReplaceAllString(text, "")
	for _, re := range gopodScriptedLinePatterns {
		text = re.ReplaceAllString(text, "")
	}
	return strings.TrimSpace(regexp.MustCompile(`[ \t]{2,}`).ReplaceAllString(text, " "))
}

// gopodDropLLMReminder cuts a "Hey, that reminds me..." the LLM wrote itself.
func gopodDropLLMReminder(text string) string {
	return strings.TrimRight(gopodReminderPhrase.ReplaceAllString(text, ""), " \t")
}

// ## EOF

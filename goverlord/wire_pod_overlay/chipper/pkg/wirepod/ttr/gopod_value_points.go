// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_value_points.go
package wirepod_ttr

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

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
	reminded map[string]bool // topics and chapters a reminder has named (not again)
}{revealed: map[string]int{}, reminded: map[string]bool{}}

func gopodReminded(name string) bool {
	gopodCoverage.Lock()
	defer gopodCoverage.Unlock()
	return gopodCoverage.reminded[name]
}

// gopodRemindAgain: a new round of reminders - nothing counts as reminded.
func gopodRemindAgain() {
	gopodCoverage.Lock()
	defer gopodCoverage.Unlock()
	gopodCoverage.reminded = map[string]bool{}
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

// gopodCoverageFacts is what the LLM gets for one loaded topic: its opening and the next point
// (marked revealed), or the whole pack once covered.
func gopodCoverageFacts(name string, content []byte) []byte {
	facts, _ := gopodCoverageReveal(name, content)
	return facts
}

// gopodCoverageReveal is gopodCoverageFacts, plus the point it gave out ("doc_and_pip 1/3"), ""
// when none.
func gopodCoverageReveal(name string, content []byte) ([]byte, string) {
	return gopodCoverageRevealFor(name, content, "")
}

// gopodCoverageSaid: which of a topic's points have been given (topic -> point number -> true),
// beside revealed's count; cleared when the count is reset.
var gopodCoverageSaid = map[string]map[int]bool{}

var gopodPickWordRE = regexp.MustCompile(`[a-z0-9']+`)

// gopodTopicLinksFor: the topics this one leads to - the next step of every funnel it's in
// (value_points.funnels), then its topic_links.
func gopodTopicLinksFor(from string) []string {
	var out []string
	seen := map[string]bool{}
	for _, f := range gopodBrobotsSettings.ValuePoints.Funnels {
		for i := 0; i+1 < len(f); i++ {
			if f[i] == from && !seen[f[i+1]] {
				seen[f[i+1]] = true
				out = append(out, f[i+1])
			}
		}
	}
	for _, t := range gopodBrobotsSettings.ValuePoints.TopicLinks[from] {
		if !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	return out
}

// gopodCorePoint: with value_points.core_points set, only those are value points (topic#n); the
// rest of a topic stays known, never a target.
func gopodCorePoint(topic string, n int) bool {
	core := gopodBrobotsSettings.ValuePoints.CorePoints
	if len(core) == 0 {
		return true
	}
	id := topic + "#" + strconv.Itoa(n)
	for _, c := range core {
		if c == id {
			return true
		}
	}
	return false
}

// gopodPickWords: a text's words for matching - lower case, "pip's" -> "pip", three letters or
// more, not in value_points.pick_skip_words.
func gopodPickWords(text string) map[string]bool {
	skip := map[string]bool{}
	for _, w := range gopodBrobotsSettings.ValuePoints.PickSkipWords {
		skip[strings.ToLower(strings.TrimSpace(w))] = true
	}
	out := map[string]bool{}
	for _, w := range gopodPickWordRE.FindAllString(strings.ToLower(text), -1) {
		w = strings.TrimSuffix(strings.Trim(w, "'"), "'s")
		if len(w) > 3 && strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss") {
			w = strings.TrimSuffix(w, "s")
		}
		if len(w) >= 3 && !skip[w] {
			out[w] = true
		}
	}
	return out
}

// gopodPickPoint: of a topic's points not said yet, the one whose words (the point, its notes, its
// say line) share the most with the question; with none shared, the next in order - the pull stays.
func gopodPickPoint(name string, points []string, said map[int]bool, question string) int {
	asked := gopodPickWords(question)
	phrase := strings.ReplaceAll(name, "_", " ")
	nameWords := gopodPickWords(phrase)
	if strings.Contains(strings.ToLower(question), phrase) {
		for n := range nameWords {
			delete(asked, n)
		}
	} else if len(nameWords) == 1 {
		for n := range nameWords {
			for w := range asked {
				if len(w) >= 4 && strings.HasPrefix(n, w) {
					delete(asked, w)
				}
			}
		}
	}
	best, bestScore := -1, 0
	for i, p := range points {
		if said[i+1] || !gopodCorePoint(name, i+1) {
			continue
		}
		if best < 0 {
			best = i // the next in order
		}
		key := name + " " + strconv.Itoa(i+1)
		words := gopodPickWords(p + " " + gopodBrobotsSettings.ValuePoints.Notes[key] + " " + gopodBrobotsSettings.ValuePoints.SayLines[key])
		score := 0
		for w := range asked {
			if words[w] {
				score += len(w)
			}
		}
		if score > bestScore {
			best, bestScore = i, score
		}
	}
	return best
}

// gopodCoverageRevealFor: gopodCoverageReveal for this question - the point it asks about first
// (gopodPickPoint).
func gopodBackgroundTopic(name string) bool {
	for _, t := range gopodBrobotsSettings.ValuePoints.BackgroundTopics {
		if t == name {
			return true
		}
	}
	return false
}

func gopodCoverageRevealFor(name string, content []byte, question string) ([]byte, string) {
	if gopodBackgroundTopic(name) && name != gopodBrobotsSettings.ValuePoints.WhatElseTopic {
		logger.Println("BROBOT_VALUE_POINT topic=" + name + " background (known, not a target)")
		return content, ""
	}
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
	count := gopodCoverage.revealed[name]
	if count >= total {
		logger.Println("BROBOT_VALUE_POINT topic=" + name + " covered")
		return content, ""
	}
	said := gopodCoverageSaid[name]
	if said == nil || len(said) > count {
		said = map[int]bool{} // the count was reset (a new session, a test)
	}
	n := count
	if len(points) > 0 {
		n = gopodPickPoint(name, points, said, question)
		if n < 0 {
			return content, ""
		}
	}
	said[n+1] = true
	gopodCoverageSaid[name] = said
	gopodCoverage.revealed[name] = count + 1
	given := fmt.Sprintf("%s %d/%d", name, n+1, total)
	logger.Println(fmt.Sprintf("BROBOT_VALUE_POINT topic=%s point=%d/%d (picked by the question: %v)", name, n+1, total, n != count))
	if len(points) == 0 {
		return content, given
	}
	point := points[n]
	if title, notes := gopodPointNotes(name, n+1); title != "" {
		point = fmt.Sprintf("%d. %s: %s", n+1, title, strings.Join(notes, "; "))
	}
	return []byte(opening + "\n" + point + "\n"), given
}

// gopodPointNotes: a point's note form (its topic's "notes N:" line) - its title and notes; "" when
// it has none.
func gopodPointNotes(topic string, n int) (string, []string) {
	raw := strings.TrimSpace(gopodBrobotsSettings.ValuePoints.Notes[topic+" "+strconv.Itoa(n)])
	if raw == "" {
		return "", nil
	}
	parts := strings.Split(raw, "|")
	var notes []string
	for _, p := range parts[1:] {
		if p = strings.TrimSpace(p); p != "" {
			notes = append(notes, p)
		}
	}
	return strings.TrimSpace(parts[0]), notes
}

// The points given out for each Brobot's current question ("doc_and_pip 1/3"), for its say line and
// its take (gopod_take_sheet.go).
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

// gopodSayLine is the exact line for a point just given out to this Brobot.
func gopodSayLine(esn string) string {
	for _, given := range gopodGivenPoints(esn) {
		if gopodSummaryGivenHas(esn, given) {
			continue
		}
		key := given
		if i := strings.LastIndex(key, "/"); i >= 0 {
			key = key[:i]
		}
		if gopodBrobotsSettings.ValuePoints.Notes[key] != "" {
			continue
		}
		if line := gopodBrobotsSettings.ValuePoints.SayLines[key]; line != "" {
			return line
		}
	}
	return ""
}

// gopodGivenHasNotes: a point given to this Brobot has notes (the model flavours them).
func gopodGivenHasNotes(esn string) bool {
	for _, given := range gopodGivenPoints(esn) {
		key := given
		if i := strings.LastIndex(key, "/"); i >= 0 {
			key = key[:i]
		}
		if gopodBrobotsSettings.ValuePoints.Notes[key] != "" {
			return true
		}
	}
	return false
}

// gopodTopicCovered: the topic is done - one point said with value_points.target "topics", all its
// points with "points".
func gopodTopicCovered(name string, total int) bool {
	gopodCoverage.Lock()
	defer gopodCoverage.Unlock()
	return gopodTopicDoneLocked(name, total)
}

// gopodTopicDoneLocked: gopodTopicCovered with gopodCoverage held.
func gopodTopicDoneLocked(name string, total int) bool {
	if total <= 0 {
		return true
	}
	if gopodBrobotsSettings.ValuePoints.Target == "topics" {
		return gopodCoverage.revealed[name] >= 1
	}
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
		if p.Name == gopodBrobotsSettings.ValuePoints.WhatElseTopic || gopodBackgroundTopic(p.Name) {
			continue // the question itself, or a background topic - not a value point
		}
		_, points := gopodSplitPoints(string(data))
		counts[p.Name] = len(points)
		if counts[p.Name] == 0 {
			counts[p.Name] = 1
		}
		if len(gopodBrobotsSettings.ValuePoints.CorePoints) > 0 {
			n := 0
			for i := 1; i <= counts[p.Name]; i++ {
				if gopodCorePoint(p.Name, i) {
					n++
				}
			}
			if n == 0 {
				delete(counts, p.Name)
				continue
			}
			counts[p.Name] = n
		}
		names = append(names, p.Name)
	}
	return names, counts
}

// gopodReminderLine is the "Hey, that reminds me..." line after a reply on the question's topics;
// "" when there's none.
var gopodReminderNames sync.Map

func gopodReminderLine(userMessage string) string {
	return gopodReminderLineFor("", userMessage)
}

// gopodReminderLineFor: the reminder, its invite in the speaking Brobot's own shapes first.
func gopodReminderLineFor(esn, userMessage string) string {
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
	if last < 0 && (gopodIsThanks(userMessage) || vp.ReminderTopicsText == "") {
		return ""
	}
	_, counts := gopodPackPointCounts()
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
	fresh := false
	for _, t := range order {
		if open(t) && !skip[t] && ic.TopicLabels[t] != "" && !gopodReminded(t) {
			fresh = true
			break
		}
	}
	if !fresh {
		gopodRemindAgain()
	}
	current := ""
	if last >= 0 {
		current = order[last]
	}
	if current != "" && lastOfChapter[current] && vp.ReminderChaptersText != "" {
		// The next chapter with its summary, then the one after by title only; a chapter once named isn't
		// named again.
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
	var chain, chainTopics []string
	at := -1
	for i, t := range order {
		if t == current {
			at = i
		}
	}
	seat := func(t string) bool {
		return !used[t] && !skip[t] && open(t) && ic.TopicLabels[t] != "" && !gopodReminded(t)
	}
	lead := gopodReminderLead(userMessage, seat)
	for from := current; len(chain) < want; {
		best := ""
		if len(chain) == 0 && lead != "" && seat(lead) {
			best = lead
		}
		for _, t := range gopodTopicLinksFor(from) {
			if best != "" {
				break
			}
			if seat(t) {
				best = t
				break
			}
		}
		for k := 1; best == "" && k <= len(order); k++ {
			if t := order[(at+k+len(order))%len(order)]; seat(t) {
				best = t
			}
		}
		if best == "" {
			break
		}
		for i, t := range order {
			if t == best {
				at = i
			}
		}
		used[best] = true
		chain = append(chain, ic.TopicLabels[best])
		chainTopics = append(chainTopics, best)
		from = best
	}
	if len(chain) == 0 {
		return ""
	}
	for t := range used {
		if t != current {
			gopodRemind(t)
		}
	}
	prober := ""
	if strings.TrimSpace(vp.ReminderProberText) != "" {
		for i, topic := range chainTopics {
			if pr := gopodProberFor(topic, esn); pr != "" {
				prober = pr
				chain = append([]string{chain[i]}, append(append([]string(nil), chain[:i]...), chain[i+1:]...)...)
				break
			}
		}
	}
	gopodReminderNames.Store(userMessage, append([]string(nil), chain...))
	line := strings.ReplaceAll(vp.ReminderTopicsText, "{topics}", gopodNameList(chain))
	if prober != "" {
		line += gopodThoughtGap() + strings.ReplaceAll(strings.TrimSpace(vp.ReminderProberText), "{prober}", prober)
	}
	return line
}

// gopodCoverageScoreText is the session score and the list behind it.
func gopodCoverageScoreText() (string, string) {
	text, detail := gopodCoverageScoreWith(gopodBrobotsSettings.ValuePoints.ScoreText)
	return text, detail
}

// gopodCoverageScoreWith: tmpl filled with the session score - {topics}, {topics_total}, {points},
// {points_total}, {eggs}, {eggs_total}, and {topics_bar}, {points_bar}, {eggs_bar}: one
// value_points.score_bar "done" mark per one found, one "left" mark per one to go.
func gopodCoverageScoreWith(tmpl string) (string, string) {
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
	topicsTotal := len(names)
	if ct := gopodBrobotsSettings.ValuePoints.CoreTopics; len(ct) > 0 {
		topics, topicsTotal = 0, len(ct)
		for _, files := range ct {
			for _, f := range files {
				if gopodCoverage.revealed[f] > 0 {
					topics++
					break
				}
			}
		}
	}
	eggs, eggsTotal := 0, 0
	listed := map[string]bool{}
	for _, e := range gopodEggList() {
		listed[e.Key] = true
		eggsTotal++
		if gopodEggFound(e.Key) {
			eggs++
		}
	}
	for _, k := range gopodBrobotsSettings.TopicPictures.NotListed {
		if !listed[k] && !gopodNotEgg(k) {
			eggsTotal++
			if gopodEggFound(k) {
				eggs++
			}
		}
	}
	bar := func(n, of int) string {
		sb := gopodBrobotsSettings.ValuePoints.ScoreBar
		if n > of {
			n = of
		}
		if n < 0 || of <= 0 {
			return ""
		}
		return strings.Repeat(sb.Done, n) + strings.Repeat(sb.Left, of-n)
	}
	text := strings.NewReplacer(
		"{points_bar}", bar(points, total), "{topics_bar}", bar(topics, topicsTotal), "{eggs_bar}", bar(eggs, eggsTotal),
		"{points}", fmt.Sprint(points), "{points_total}", fmt.Sprint(total),
		"{topics}", fmt.Sprint(topics), "{topics_total}", fmt.Sprint(topicsTotal),
		"{eggs}", fmt.Sprint(eggs), "{eggs_total}", fmt.Sprint(eggsTotal),
	).Replace(tmpl)
	return text, strings.Join(detail, " ")
}

// gopodCoverageScore shows the score once per ending (a new question that isn't a thanks allows the
// next one, gopodCoverageNewQuestion).
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
	if vp.ScoreRole != "" {
		GopodChatFeedAdd(vp.ScoreRole, text)
	}
}

func gopodCoverageNewQuestion() {
	gopodCoverage.Lock()
	gopodCoverage.scored = false
	gopodCoverage.Unlock()
}

// gopodWhatElseFacts answers "what else can we talk about?": the next chapters, in interview order,
// with points not covered yet, and the score so far.
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

// The scripted lines are the code's, never the LLM's.
var gopodReminderPhrase = regexp.MustCompile(`(?is)\s*\bhey\W+that\s+reminds\s+me\b.*$`)

var gopodScriptedLinePatterns = func() []*regexp.Regexp {
	var lines []string
	for _, s := range gopodBrobotsSettings.ScriptedSnippets.Lines {
		lines = append(lines, s.Lines...)
	}
	for _, line := range gopodBrobotsSettings.ValuePoints.Reminders {
		lines = append(lines, line)
	}
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

func gopodReminderLead(question string, seat func(string) bool) string {
	bc := gopodBrobotsSettings.BrainChain
	list := func(s string) []string {
		var out []string
		for _, t := range strings.Split(s, ",") {
			if t = strings.TrimSpace(t); t != "" {
				out = append(out, t)
			}
		}
		return out
	}
	if after, err := strconv.Atoi(strings.TrimSpace(bc.UnknownPullAfter)); err == nil && after > 0 && int(gopodUnknownAsks.Load()) >= after {
		for _, t := range list(bc.UnknownPull) {
			if seat(t) {
				return t
			}
		}
	}
	if t := gopodValueGridBounce(question, seat); t != "" {
		return t
	}
	asked := gopodLookupMeaningful(strings.Join(gopodSameWordsExpand(gopodLookupTokens(question)), " "))
	for _, c := range gopodLookup(question) {
		for w := range gopodLookupMeaningful(strings.Join(gopodSameWordsExpand(c.Words), " ")) {
			asked[w] = true
		}
	}
	pull := map[string]bool{}
	for _, t := range list(bc.ReminderPull) {
		pull[t] = true
	}
	best, bestScore := "", 0
	for _, e := range gopodTopicRegistry() {
		if !seat(e.Name) {
			continue
		}
		score := 0
		for w := range gopodLookupMeaningful(strings.Join(e.Triggers, " ")) {
			if asked[w] {
				score++
			}
		}
		if score > bestScore || (score > 0 && score == bestScore && pull[e.Name] && !pull[best]) {
			best, bestScore = e.Name, score
		}
	}
	if best != "" {
		return best
	}
	for _, t := range list(bc.ReminderPull) {
		if seat(t) {
			return t
		}
	}
	return ""
}

// ## EOF

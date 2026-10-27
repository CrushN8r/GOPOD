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
	return gopodCoverageRevealFor(name, content, "")
}

// gopodCoverageSaid: which of a topic's points have been given (topic -> point
// number -> true), beside revealed's count; cleared when the count is reset.
var gopodCoverageSaid = map[string]map[int]bool{}

var gopodPickWordRE = regexp.MustCompile(`[a-z0-9']+`)

// gopodTopicLinksFor (2026-10-23, operator: "livestream > content engine > ...
// > subscribe * Where you can further decouple for easier reordering/shuffling,
// please do"): the topics this one leads to - the next step of every funnel it's
// in (value_points.funnels), then its topic_links.
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

// gopodCorePoint (2026-10-23, FINISH_PLAN_001.md A3; CORE_TOPICS_002.md): with
// value_points.core_points set, only those are value points (topic#n); the rest
// of a topic stays known, never a target.
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

// gopodPickWords: a text's words for matching - lower case, "pip's" -> "pip",
// three letters or more, not in value_points.pick_skip_words.
func gopodPickWords(text string) map[string]bool {
	skip := map[string]bool{}
	for _, w := range gopodBrobotsSettings.ValuePoints.PickSkipWords {
		skip[strings.ToLower(strings.TrimSpace(w))] = true
	}
	out := map[string]bool{}
	for _, w := range gopodPickWordRE.FindAllString(strings.ToLower(text), -1) {
		w = strings.TrimSuffix(strings.Trim(w, "'"), "'s")
		// "runs" and "run" count as one word.
		if len(w) > 3 && strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss") {
			w = strings.TrimSuffix(w, "s")
		}
		if len(w) >= 3 && !skip[w] {
			out[w] = true
		}
	}
	return out
}

// gopodPickPoint (2026-10-22, live: "Tell me about Pip" got Doc's point and "Who
// is Doc" got Pip's - points went out in order, not by what was asked;
// operator: "go on step 1"): of a topic's points not said yet, the one whose
// words (the point, its notes, its say line) share the most with the question;
// with none shared, the next in order - the pull stays. 0-based; -1 = none left.
func gopodPickPoint(name string, points []string, said map[int]bool, question string) int {
	asked := gopodPickWords(question)
	// Words that name the topic itself ask for the topic, not one of its points:
	// the whole name in the question ("what's the community path?"), or a
	// one-word name's start ("book" for booking) - four letters or more.
	// ("a decimal back into a fraction" still picks its point: the name isn't
	// said whole.)
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
		// A longer word is a more telling one ("granny" over "live").
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

// gopodCoverageRevealFor: gopodCoverageReveal for this question - the point it
// asks about first (gopodPickPoint).
// gopodBackgroundTopic (2026-10-22, operator: "Do 1 and 2 together first";
// CORE_TOPICS_002.md - seven core topics): a topic the Brobots still know, but
// never a value point - no point given, not in the score, never reminded.
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
	// 2026-10-22 (operator: "should be note form for the llm to flavour"): a
	// point with notes goes to the model as its notes.
	point := points[n]
	if title, notes := gopodPointNotes(name, n+1); title != "" {
		point = fmt.Sprintf("%d. %s: %s", n+1, title, strings.Join(notes, "; "))
	}
	return []byte(opening + "\n" + point + "\n"), given
}

// gopodPointNotes: a point's note form (its topic's "notes N:" line) - its
// title and notes; "" when it has none.
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
		// 2026-10-26: a summary ask's point is scored, not said (gopod_topic_asks.go).
		if gopodSummaryGivenHas(esn, given) {
			continue
		}
		key := given
		if i := strings.LastIndex(key, "/"); i >= 0 {
			key = key[:i]
		}
		// 2026-10-22 (operator: "only the model's flavoured version of the notes *
		// I'd rather the llm go brobots party on the idea * details don't matter as
		// much"): a point with notes has no exact line - the model flavours them.
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

// gopodTopicCovered: the topic is done - one point said with value_points.target
// "topics" (2026-10-23, operator: "Perhaps its the topics I'm after * value points
// ... too much to handle at this stage * Go"), all its points with "points".
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
		// 2026-10-23 (A3): only its value points count (core_points); a topic with
		// none isn't a target.
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
		// 2026-10-23 (operator: "Answer chain!! Start at what is GOPOD * GOPOD =
		// 'AI Wordplay! Explain the MAth!' = live stream > contest > ... = content
		// engine for websites * recipes > chocolate bingo > foodmath website"): the
		// map first - the first open topic linked from this one (topic_links).
		for _, t := range gopodTopicLinksFor(from) {
			if !used[t] && !skip[t] && open(t) && ic.TopicLabels[t] != "" && !gopodReminded(t) {
				best, bestSim = t, 2
				break
			}
		}
		for _, t := range order {
			if best != "" && bestSim == 2 {
				break
			}
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
	text, detail := gopodCoverageScoreWith(gopodBrobotsSettings.ValuePoints.ScoreText)
	return text, detail
}

// gopodCoverageScoreWith: tmpl filled with the session score - {topics},
// {topics_total}, {points}, {points_total}, {eggs}, {eggs_total}, and (2026-10-25,
// operator: "Topics 3 of 7: ===---- * 3 single lines with dynamic scores aligned
// vertically") {topics_bar}, {points_bar}, {eggs_bar}: one value_points.score_bar
// "done" mark per one found, one "left" mark per one to go.
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
	// 2026-10-23 (A3; operator: "keep list score for topics, easter eggs, value
	// points N of N so I can see my progress * and so can the brobots"): topics
	// are the seven core topics (core_topics), opened by any of their points.
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
	// 2026-10-23 (operator: "List hidden easter eggs as well"): the surprise eggs
	// (topic_pictures.not_listed) count too, when they're eggs.
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

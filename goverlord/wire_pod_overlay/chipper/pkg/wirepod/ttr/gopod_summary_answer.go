// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_summary_answer.go
package wirepod_ttr

import (
	"strconv"
	"strings"
	"sync"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The summary answer type (2026-10-23, operator: "'AI wordplay' is a good example
// of what I mean by overlapping * if I ask, 'What is AI wordplay?' * that would be
// like a golden query that unlocks it all * large overlapping value-points will
// demand more of a contextual summary * please build its own answer type"): a
// golden query (summary.queries) or a question that opens summary.min_topics
// core topics or more gets a short summary across the core topics - one bubble
// per core topic still to come (its next value point's title and first note), up
// to summary.max_topics - in place of one value point. No point is counted: it's
// a map of what's ahead, and the pull still names what's next.

type gopodSummaryItem struct{ topic, title, note string }

var gopodSummaryFor sync.Map // esn -> []gopodSummaryItem

// gopodSummaryQuery: the question asks for the overview.
func gopodSummaryQuery(question string) bool {
	s := gopodBrobotsSettings.Summary
	lower := strings.ToLower(strings.Join(strings.Fields(question), " "))
	for _, q := range s.Queries {
		if q = strings.ToLower(strings.TrimSpace(q)); q != "" && strings.Contains(lower, q) {
			return true
		}
	}
	if s.MinTopics <= 0 {
		return false
	}
	opened := map[string]bool{}
	for _, p := range loadScenarioPacketsForMessage(question) {
		for core, files := range gopodBrobotsSettings.ValuePoints.CoreTopics {
			for _, f := range files {
				if f == p.Name {
					opened[core] = true
				}
			}
		}
	}
	return len(opened) >= s.MinTopics
}

// gopodSummaryItems: the core topics still to come, each with its next value
// point's title and first note, in core_topics order (the settings' order).
func gopodSummaryItems() []gopodSummaryItem {
	s := gopodBrobotsSettings.Summary
	want := s.MaxTopics
	if want <= 0 {
		want = 3
	}
	_, counts := gopodPackPointCounts()
	var out []gopodSummaryItem
	for _, core := range s.TopicOrder {
		if len(out) >= want {
			break
		}
		for _, f := range gopodBrobotsSettings.ValuePoints.CoreTopics[core] {
			if gopodTopicCovered(f, counts[f]) {
				continue
			}
			for n := 1; n <= 20; n++ {
				if !gopodCorePoint(f, n) {
					continue
				}
				title, notes := gopodPointNotes(f, n)
				if title == "" {
					continue
				}
				note := ""
				if len(notes) > 0 {
					note = notes[0]
				}
				out = append(out, gopodSummaryItem{topic: core, title: title, note: note})
				break
			}
			break
		}
	}
	return out
}

// gopodSummaryFacts: the summary's facts for the prompt, and it's noted for the
// answer chain (summary_bubbles); "" when the question isn't one.
func gopodSummaryFacts(esn, question string) string {
	gopodSummaryFor.Delete(esn)
	if !gopodSummaryQuery(question) {
		return ""
	}
	items := gopodSummaryItems()
	if len(items) == 0 {
		return ""
	}
	gopodSummaryFor.Store(esn, items)
	var lines []string
	for i, it := range items {
		lines = append(lines, strconv.Itoa(i+1)+") "+it.topic+": "+it.title+" - "+it.note)
	}
	logger.Println("BROBOT_SUMMARY " + gopodBrobotName(esn) + ": " + strings.Join(lines, " "))
	return strings.ReplaceAll(gopodBrobotsSettings.Summary.FactsNote, "{topics}", strings.Join(lines, "\n"))
}

// gopodSummaryItemsFor: this answer's summary, if it is one.
func gopodSummaryItemsFor(esn string) []gopodSummaryItem {
	v, ok := gopodSummaryFor.Load(esn)
	if !ok {
		return nil
	}
	return v.([]gopodSummaryItem)
}

// gopodSummaryBubbles: the answer chain's summary_bubbles part - one bubble per
// core topic, the model's line if it names that topic's title or note, else
// made from them (summary.fallback). The shape always holds.
func gopodSummaryBubbles(esn, raw string) string {
	items := gopodSummaryItemsFor(esn)
	if len(items) == 0 {
		return raw
	}
	cue, body := "", raw
	if loc := gopodAnimationMarkerPattern.FindStringIndex(raw); loc != nil && loc[0] == 0 {
		cue, body = raw[:loc[1]], raw[loc[1]:]
	}
	var lines []string
	for _, l := range strings.Split(body, "\n") {
		for _, s := range gopodWholeSentences(l) {
			if s = strings.TrimSpace(s); s != "" {
				lines = append(lines, s)
			}
		}
	}
	used := make([]bool, len(lines))
	var out []string
	made := 0
	for _, it := range items {
		line := ""
		for i, l := range lines {
			if !used[i] && gopodShareWord(l, it.title+" "+it.note) {
				used[i], line = true, l
				break
			}
		}
		if line == "" {
			made++
			line = strings.NewReplacer("{topic}", it.topic, "{title}", it.title, "{Title}", gopodCapFirst(it.title), "{note}", it.note).Replace(gopodBrobotsSettings.Summary.Fallback)
		}
		out = append(out, strings.TrimSpace(line))
	}
	logger.Println("BROBOT_SUMMARY_BUBBLES " + gopodBrobotName(esn) + ": " + strconv.Itoa(len(items)-made) + " of " + strconv.Itoa(len(items)) + " from the model")
	result := strings.Join(out, gopodThoughtGap())
	if cue != "" {
		return strings.TrimRight(cue, " \t") + " " + result
	}
	return result
}

// ## EOF

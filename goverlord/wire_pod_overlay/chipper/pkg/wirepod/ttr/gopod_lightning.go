// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_lightning.go
package wirepod_ttr

// GOLDEN, the lightning (2026-10-29, operator: "I see Goverlord orchestrating like
// how lightning branches out gravitating to the path of least resistance: the
// library of value points * in whatever shortest trajectory it takes to grab the
// best relevant value points * pull out with the goodies the same way the
// lightning disappears"; "topic order still matters to GOPOD game * what needs to
// be revealed first, so whatever value reveals after make better sense * bigger
// aha moment * I would establish a dynamic topic order valuation list of
// topic-cluster topics found"). Measured first (TestLightningMeasure): asked out
// of order, the right topic first every time, against half the time before.
//
// A question branches out over every value point in the library (the
// topic-cluster files): each point pulls by the words it shares with the question
// (gopodPickWords - rarer words pull harder), its topic by its own words (today's
// trigger). Two stages - the lightning finds the strong branches (a trigger
// matched, or a point pulling lightning.min_word_pull or more, and at least
// lightning.keep_share of the best), then the dynamic order valuation decides:
// the topics found in the game's order, the ones already covered last. The best
// lightning.top_topics are loaded;
// inside a topic the point asked about is revealed first (gopodPickPoint), else
// the next in order. lightning.on false: today's word triggers, in registry order.

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

// gopodLightningTopic: one topic on the ranked list.
type gopodLightningTopic struct {
	entry   scenarioPacketEntry
	pull    float64 // its strongest point's pull, its own words included
	order   int     // its place in the game's order (the registry)
	covered bool    // every point already revealed
}

// gopodLightningTopics: the topics a question finds, ranked (the most valuable first).
func gopodLightningTopics(message string, registry []scenarioPacketEntry) []gopodLightningTopic {
	ls := gopodBrobotsSettings.Lightning
	lower := strings.ToLower(message)
	asked := gopodPickWords(message)
	type pointDoc struct {
		topic string
		n     int
		words map[string]bool
	}
	var docs []pointDoc
	files := map[string]gopodTopicFile{}
	chapters, always, _ := gopodReadChaptersFolder()
	for _, t := range always {
		files[t.Name] = t
	}
	for _, c := range chapters {
		for _, t := range c.Topics {
			files[t.Name] = t
		}
	}
	for _, e := range registry {
		facts, err := gopodReadTopicFacts(e.File)
		if err != nil {
			continue
		}
		t := files[e.Name]
		_, points := gopodSplitPoints(string(facts))
		for i, p := range points {
			key := e.Name + " " + strconv.Itoa(i+1)
			text := p + " " + gopodBrobotsSettings.ValuePoints.Notes[key] + " " + gopodBrobotsSettings.ValuePoints.SayLines[key] + " " + t.Label + " " + strings.Join(t.Words, " ")
			docs = append(docs, pointDoc{topic: e.Name, n: i + 1, words: gopodPickWords(text)})
		}
	}
	df := map[string]int{}
	for _, d := range docs {
		for w := range d.words {
			df[w]++
		}
	}
	idf := func(w string) float64 { return math.Log(float64(len(docs)+1) / float64(df[w]+1)) }
	best := map[string]float64{}
	for _, d := range docs {
		s := 0.0
		for w := range asked {
			if d.words[w] {
				s += idf(w)
			}
		}
		if s > best[d.topic] {
			best[d.topic] = s
		}
	}
	gopodCoverage.Lock()
	revealed := map[string]int{}
	for k, v := range gopodCoverage.revealed {
		revealed[k] = v
	}
	gopodCoverage.Unlock()
	pointsIn := map[string]int{}
	for _, d := range docs {
		pointsIn[d.topic]++
	}
	// 1. The lightning finds: the strong branches - a topic's own words (its
	// trigger), or a point pulling min_word_pull or more; only those pulling
	// keep_share or more of the best one.
	var out []gopodLightningTopic
	for i, e := range registry {
		trig := false
		for _, trigger := range e.Triggers {
			if gopodTriggerOwnMatch(lower, strings.ToLower(trigger), e.Name, registry) {
				trig = true
				break
			}
		}
		if !trig && best[e.Name] < ls.MinWordPull {
			continue
		}
		pull := best[e.Name]
		if trig {
			pull += ls.TriggerPull
		}
		out = append(out, gopodLightningTopic{entry: e, pull: pull, order: i,
			covered: pointsIn[e.Name] > 0 && revealed[e.Name] >= pointsIn[e.Name]})
	}
	strongest := 0.0
	for _, t := range out {
		strongest = math.Max(strongest, t.pull)
	}
	kept := out[:0]
	for _, t := range out {
		if ls.KeepShare <= 0 || t.pull >= ls.KeepShare*strongest {
			kept = append(kept, t)
		}
	}
	out = kept
	// 2. The order decides (operator: "what needs to be revealed first, so
	// whatever value reveals after make better sense"): the strongest pull first;
	// topics pulling within order_margin of each other go in the game's order; the
	// ones already covered last. (Live 2026-10-29, the real-questions run: by game
	// order alone, a weak earlier topic jumped ahead - "Who's Brobot 0?" went to
	// the brothers.)
	sort.SliceStable(out, func(a, b int) bool { return out[a].pull > out[b].pull })
	for i := 0; i < len(out); {
		j := i + 1
		for j < len(out) && out[i].pull-out[j].pull <= ls.OrderMargin {
			j++
		}
		group := out[i:j]
		sort.SliceStable(group, func(a, b int) bool { return group[a].order < group[b].order })
		i = j
	}
	sort.SliceStable(out, func(a, b int) bool { return !out[a].covered && out[b].covered })
	if ls.TopTopics > 0 && len(out) > ls.TopTopics {
		out = out[:ls.TopTopics]
	}
	return out
}

// ## EOF

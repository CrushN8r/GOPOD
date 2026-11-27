// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_lightning.go
package wirepod_ttr

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
	// 1.
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

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_lightning_measure_test.go
package wirepod_ttr

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestLightningMeasure (2026-10-29, operator: "Goverlord orchestrating like how
// lightning branches out gravitating to the path of least resistance: the library
// of value points ... whatever shortest trajectory it takes to grab the best
// relevant value points"; "go lightning for step 1, the measurement"): a
// measurement, not a gate - it runs only with GOPOD_LIGHTNING_MEASURE=1. The
// admin run's golden questions, in order: today (the question's words load whole
// topics) against lightning (every value point scored by its pull - words shared
// with the question, rarer words pulling harder, its topic's own words matching -
// the points already won out of the race).
func TestLightningMeasure(t *testing.T) {
	if os.Getenv("GOPOD_LIGHTNING_MEASURE") != "1" {
		t.Skip("a measurement - GOPOD_LIGHTNING_MEASURE=1 runs it")
	}
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	qdata, err := os.ReadFile(filepath.Join(root, "gopod_probes/interview_harvest/harvest_questions.json"))
	if err != nil {
		t.Fatal(err)
	}
	var qs struct {
		Questions []struct{ Topic, Ask string } `json:"questions"`
	}
	json.Unmarshal(qdata, &qs)
	gdata, _ := os.ReadFile(filepath.Join(root, "gopod_probes/chapters/value_points_golden.json"))
	var golden struct {
		Points map[string]struct{ Words string } `json:"points"`
	}
	json.Unmarshal(gdata, &golden)

	// The library: each point's words - its golden line, its notes and say line,
	// its topic's label and words.
	chapters, always, _ := gopodReadChaptersFolder()
	topics := map[string]gopodTopicFile{}
	for _, tf := range always {
		topics[tf.Name] = tf
	}
	for _, c := range chapters {
		for _, tf := range c.Topics {
			topics[tf.Name] = tf
		}
	}
	word := regexp.MustCompile(`[a-z0-9']+`)
	stop := map[string]bool{}
	for _, w := range strings.Fields("a an the and or of to in on at for with is are was be it its it's you your me my we our us i do does did can could would should will what who how why when where which that this these those there here about tell say so just get got have has had from by as if then than not no yes up out all any some more most very really one two brobot brobots bro hey please") {
		stop[w] = true
	}
	stem := func(w string) string {
		w = strings.Trim(w, "'")
		if len(w) > 4 && strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss") {
			w = w[:len(w)-1]
		}
		return w
	}
	tokens := func(s string) map[string]bool {
		out := map[string]bool{}
		for _, w := range word.FindAllString(strings.ToLower(s), -1) {
			if w = stem(w); !stop[w] && len(w) > 1 {
				out[w] = true
			}
		}
		return out
	}
	type point struct {
		id, topic string
		n         int
		toks      map[string]bool
	}
	var points []point
	for id, p := range golden.Points {
		i := strings.LastIndex(id, "#")
		name := id[:i]
		n, _ := strconv.Atoi(id[i+1:])
		tf := topics[name]
		text := p.Words + " " + tf.Notes[n] + " " + tf.SayLines[n] + " " + tf.Label + " " + strings.Join(tf.Words, " ")
		points = append(points, point{id: id, topic: name, n: n, toks: tokens(text)})
	}
	sort.Slice(points, func(a, b int) bool { return points[a].id < points[b].id })
	df := map[string]int{}
	for _, p := range points {
		for w := range p.toks {
			df[w]++
		}
	}
	idf := func(w string) float64 { return math.Log(float64(len(points)+1) / float64(df[w]+1)) }
	registry := gopodTopicRegistry()

	type row struct {
		ask, want             string
		todayHit, todayFirst  bool
		todayLoaded           int
		top1, top3, topicTop1 bool
		got                   string
	}
	var rows []row
	seen := map[string]int{}
	won := map[string]bool{}
	for _, q := range qs.Questions {
		seen[q.Topic]++
		want := q.Topic + "#" + strconv.Itoa(seen[q.Topic])
		if _, ok := golden.Points[want]; !ok {
			continue
		}
		r := row{ask: q.Ask, want: want}
		// Today: whole topics by their words.
		loaded := loadScenarioPacketsForMessage(q.Ask)
		r.todayLoaded = len(loaded)
		for i, p := range loaded {
			if p.Name == q.Topic {
				r.todayHit = true
				r.todayFirst = i == 0
			}
		}
		// Lightning: every point not yet won, scored by its pull.
		qt := tokens(q.Ask)
		lower := strings.ToLower(q.Ask)
		type scored struct {
			id, topic string
			s         float64
		}
		var race []scored
		for _, p := range points {
			if won[p.id] {
				continue
			}
			s := 0.0
			for w := range qt {
				if p.toks[w] {
					s += idf(w)
				}
			}
			// The topic's own words matching (today's trigger) pull the whole
			// topic - its next point not yet won most.
			for _, reg := range registry {
				if reg.Name != p.topic {
					continue
				}
				for _, trig := range reg.Triggers {
					if gopodTriggerOwnMatch(lower, strings.ToLower(trig), reg.Name, registry) {
						s += 2.0
						break
					}
				}
			}
			s -= 0.05 * float64(p.n) // the earlier point first, all else equal
			race = append(race, scored{p.id, p.topic, s})
		}
		sort.SliceStable(race, func(a, b int) bool { return race[a].s > race[b].s })
		if len(race) > 0 {
			r.got = race[0].id
			r.top1 = race[0].id == want
			r.topicTop1 = race[0].topic == q.Topic
			for i := 0; i < 3 && i < len(race); i++ {
				if race[i].id == want {
					r.top3 = true
				}
			}
		}
		won[want] = true // the game goes on: the question won its point
		rows = append(rows, r)
	}
	count := func(f func(row) bool) int {
		n := 0
		for _, r := range rows {
			if f(r) {
				n++
			}
		}
		return n
	}
	loadedSum := 0
	for _, r := range rows {
		loadedSum += r.todayLoaded
	}
	var b strings.Builder
	fmt.Fprintf(&b, "questions measured: %d (of %d; a question past its topic's points is left out)\n", len(rows), len(qs.Questions))
	fmt.Fprintf(&b, "TODAY  right topic loaded: %d · right topic first: %d · topics loaded per question: %.2f · none loaded: %d\n",
		count(func(r row) bool { return r.todayHit }), count(func(r row) bool { return r.todayFirst }), float64(loadedSum)/float64(len(rows)), count(func(r row) bool { return r.todayLoaded == 0 }))
	fmt.Fprintf(&b, "LIGHTNING  right point first: %d · right point in top 3: %d · right topic first: %d\n",
		count(func(r row) bool { return r.top1 }), count(func(r row) bool { return r.top3 }), count(func(r row) bool { return r.topicTop1 }))
	b.WriteString("\nMISSES (lightning's first pick not the right point):\n")
	for _, r := range rows {
		if !r.top1 {
			fmt.Fprintf(&b, "  %-28s got %-28s today:%v loaded:%d  %q\n", r.want, r.got, r.todayHit, r.todayLoaded, gopodCut(r.ask, 70))
		}
	}
	// The room asks out of order: the same questions shuffled (a fixed seed, five
	// rounds). Today the Brobot reveals the next point not yet said of the first
	// topic loaded; lightning picks by pull.
	type ask struct{ want, topic, text string }
	var asks []ask
	for _, r := range rows {
		asks = append(asks, ask{r.want, strings.SplitN(r.want, "#", 2)[0], r.ask})
	}
	rng := uint64(20261029)
	next := func(n int) int {
		rng = rng*6364136223846793005 + 1442695040888963407
		return int((rng >> 33) % uint64(n))
	}
	todayRight, lightRight, lightTop3, total := 0, 0, 0, 0
	for round := 0; round < 5; round++ {
		order := append([]ask{}, asks...)
		for i := len(order) - 1; i > 0; i-- {
			j := next(i + 1)
			order[i], order[j] = order[j], order[i]
		}
		said := map[string]bool{}
		for _, a := range order {
			total++
			// Today: the first topic loaded, its next point not yet said.
			if loaded := loadScenarioPacketsForMessage(a.text); len(loaded) > 0 {
				for n := 1; n < 20; n++ {
					id := loaded[0].Name + "#" + strconv.Itoa(n)
					if _, ok := golden.Points[id]; !ok {
						break
					}
					if !said[id] {
						if id == a.want {
							todayRight++
						}
						break
					}
				}
			}
			// Lightning: by pull, the points said out of the race.
			qt := tokens(a.text)
			lower := strings.ToLower(a.text)
			best, bestS := "", -1e9
			type sc struct {
				id string
				s  float64
			}
			var race []sc
			for _, p := range points {
				if said[p.id] {
					continue
				}
				s := 0.0
				for w := range qt {
					if p.toks[w] {
						s += idf(w)
					}
				}
				for _, reg := range registry {
					if reg.Name == p.topic {
						for _, trig := range reg.Triggers {
							if gopodTriggerOwnMatch(lower, strings.ToLower(trig), reg.Name, registry) {
								s += 2.0
								break
							}
						}
					}
				}
				s -= 0.05 * float64(p.n)
				race = append(race, sc{p.id, s})
				if s > bestS {
					best, bestS = p.id, s
				}
			}
			sort.SliceStable(race, func(x, y int) bool { return race[x].s > race[y].s })
			if best == a.want {
				lightRight++
			}
			for i := 0; i < 3 && i < len(race); i++ {
				if race[i].id == a.want {
					lightTop3++
				}
			}
			said[a.want] = true
		}
	}
	fmt.Fprintf(&b, "\nOUT OF ORDER (5 shuffled rounds, %d asks): today right point: %d (%.0f%%) · lightning right point first: %d (%.0f%%) · in top 3: %d (%.0f%%)\n",
		total, todayRight, 100*float64(todayRight)/float64(total), lightRight, 100*float64(lightRight)/float64(total), lightTop3, 100*float64(lightTop3)/float64(total))
	// Step 2 (operator: "dynamic topic order valuation list of topic-cluster
	// topics found"): the loader itself - lightning.on true against false - on the
	// same questions, in order and shuffled: the right topic first, topics loaded,
	// and chit-chat that names no topic.
	chat := []string{"how are you today?", "good evening bros!", "what's up?", "that was funny", "say something nice", "are you awake?", "lol", "thank you", "can you hear me?", "what time is it?"}
	loader := func(on bool) (inOrder, shuffled, loadedN, chatN int) {
		saved := gopodBrobotsSettings.Lightning.On
		gopodBrobotsSettings.Lightning.On = on
		defer func() { gopodBrobotsSettings.Lightning.On = saved }()
		for _, a := range asks {
			l := loadScenarioPacketsForMessage(a.text)
			loadedN += len(l)
			if len(l) > 0 && l[0].Name == a.topic {
				inOrder++
			}
		}
		rng2 := uint64(7)
		order := append([]ask{}, asks...)
		for i := len(order) - 1; i > 0; i-- {
			rng2 = rng2*6364136223846793005 + 1442695040888963407
			j := int((rng2 >> 33) % uint64(i+1))
			order[i], order[j] = order[j], order[i]
		}
		gopodCoverage.Lock()
		keep := gopodCoverage.revealed
		gopodCoverage.revealed = map[string]int{}
		gopodCoverage.Unlock()
		for _, a := range order {
			l := loadScenarioPacketsForMessage(a.text)
			if len(l) > 0 && l[0].Name == a.topic {
				shuffled++
			}
			gopodCoverage.Lock()
			gopodCoverage.revealed[a.topic]++
			gopodCoverage.Unlock()
		}
		gopodCoverage.Lock()
		gopodCoverage.revealed = keep
		gopodCoverage.Unlock()
		for _, c := range chat {
			chatN += len(loadScenarioPacketsForMessage(c))
		}
		return
	}
	o1, s1, n1, c1 := loader(false)
	o2, s2, n2, c2 := loader(true)
	fmt.Fprintf(&b, "\nSTEP 2 - THE LOADER (right topic first, of %d)\n", len(asks))
	fmt.Fprintf(&b, "  triggers (lightning off): in order %d · shuffled %d · topics loaded %d · chit-chat topics loaded %d of %d phrases\n", o1, s1, n1, c1, len(chat))
	fmt.Fprintf(&b, "  lightning (on):           in order %d · shuffled %d · topics loaded %d · chit-chat topics loaded %d of %d phrases\n", o2, s2, n2, c2, len(chat))
	t.Log("\n" + b.String())
	if out := os.Getenv("GOPOD_LIGHTNING_REPORT"); out != "" {
		os.WriteFile(out, []byte(b.String()), 0o644)
	}
}

// ## EOF

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_probers_list.go
package wirepod_ttr

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// gopodProber: a question and the topics it should land on first ("" = none).
type gopodProber struct {
	text string
	want []string
}

// gopodProbers: the probers list's questions.
func gopodProbers() ([]gopodProber, string, error) {
	file := filepath.Join(gopodChipperRoot(), gopodBrobotsSettings.SystemTests.ProbersList.File)
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, "", err
	}
	var out []gopodProber
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		q, want, ok := strings.Cut(line, " | ")
		if line == "" || strings.HasPrefix(line, "#") || !ok {
			continue
		}
		r := gopodProber{text: strings.TrimSpace(q)}
		if w := strings.TrimSpace(want); w != "-" {
			for _, t := range strings.Split(w, "/") {
				r.want = append(r.want, strings.TrimSpace(t))
			}
		}
		out = append(out, r)
	}
	return out, filepath.Base(file), nil
}

// gopodProbersScore: how many land right with the lightning on or off, and the misses.
func gopodProbersScore(qs []gopodProber, lightning bool) (int, []string) {
	saved := gopodBrobotsSettings.Lightning.On
	gopodBrobotsSettings.Lightning.On = lightning
	defer func() { gopodBrobotsSettings.Lightning.On = saved }()
	right := 0
	var misses []string
	for _, q := range qs {
		got := ""
		if l := loadScenarioPacketsForMessage(q.text); len(l) > 0 {
			got = l[0].Name
		}
		ok := got == "" && len(q.want) == 0
		for _, w := range q.want {
			ok = ok || w == got
		}
		if ok {
			right++
			continue
		}
		want := strings.Join(q.want, "/")
		if want == "" {
			want = "none"
		}
		if got == "" {
			got = "none"
		}
		misses = append(misses, fmt.Sprintf("%q -> %s (want %s)", gopodCut(q.text, 50), got, want))
	}
	return right, misses
}

// gopodProbersListLine: system-test-probers-list.
func gopodProbersListLine() (bool, string) {
	qs, file, err := gopodProbers()
	if err != nil {
		return false, err.Error()
	}
	if len(qs) == 0 {
		return false, file + ": no questions"
	}
	gopodCoverage.Lock()
	keep := gopodCoverage.revealed
	gopodCoverage.revealed = map[string]int{}
	gopodCoverage.Unlock()
	defer func() {
		gopodCoverage.Lock()
		gopodCoverage.revealed = keep
		gopodCoverage.Unlock()
	}()
	right, misses := gopodProbersScore(qs, true)
	old, _ := gopodProbersScore(qs, false)
	share := float64(right) / float64(len(qs))
	line := fmt.Sprintf("%s, %d questions: lightning right %d (%.0f%%) · word triggers right %d (%.0f%%)", file, len(qs), right, 100*share, old, 100*float64(old)/float64(len(qs)))
	if len(misses) > 0 {
		line += "\n  misses: " + strings.Join(misses, "\n  misses: ")
	}
	return share >= gopodBrobotsSettings.SystemTests.ProbersList.PassShare, line
}

// gopodProberLast: each topic's last invite, so the next one differs (gopodProberFor).
var gopodProberLast = struct {
	sync.Mutex
	was map[string]string
}{was: map[string]string{}}

// gopodProberFor: an invite to a topic, shuffled - one of Ron's probers for it, or one made live
// for a value point not said yet, in the speaking Brobot's own shapes first
// (probers_make.persona_shapes), then the shared ones; never the same twice running; "" with none.
func gopodProberFor(topic, esn string) string {
	var pool []string
	if probers, _, err := gopodProbers(); err == nil {
		for _, p := range probers {
			if len(p.want) == 0 || p.want[0] != topic {
				continue
			}
			if !gopodProbeLands(p.text, topic, nil, -1) {
				continue
			}
			pool = append(pool, p.text)
		}
	}
	pool = append(pool, gopodProbesLive(topic, esn)...)
	if len(pool) == 0 {
		return ""
	}
	gopodProberLast.Lock()
	defer gopodProberLast.Unlock()
	rand.Shuffle(len(pool), func(a, b int) { pool[a], pool[b] = pool[b], pool[a] })
	pick := pool[0]
	if pick == gopodProberLast.was[topic] && len(pool) > 1 {
		pick = pool[1]
	}
	gopodProberLast.was[topic] = pick
	return pick
}

// gopodProbesLive: an invite for each of a topic's core value points not said yet, its shapes shuffled.
func gopodProbesLive(topic, esn string) []string {
	pm := gopodBrobotsSettings.ProbersMake
	chapters, _, _ := gopodReadChaptersFolder()
	for _, c := range chapters {
		for _, t := range c.Topics {
			if t.Name != topic {
				continue
			}
			pts, word, ok := gopodProbeTopic(t)
			if !ok {
				return nil
			}
			gopodCoverage.Lock()
			said := map[int]bool{}
			for n, v := range gopodCoverageSaid[topic] {
				said[n] = v
			}
			gopodCoverage.Unlock()
			mine := append([]string(nil), pm.PersonaShapes[gopodBrobotName(esn)]...)
			shared := append([]string(nil), pm.Shapes...)
			rand.Shuffle(len(mine), func(a, b int) { mine[a], mine[b] = mine[b], mine[a] })
			rand.Shuffle(len(shared), func(a, b int) { shared[a], shared[b] = shared[b], shared[a] })
			var out []string
			for i, p := range pts {
				if said[i+1] || !gopodCorePoint(topic, i+1) {
					continue
				}
				if q := gopodProbeFor(t, pts, i, p, word, append(mine, shared...)); q != "" {
					out = append(out, q)
				}
			}
			return out
		}
	}
	return nil
}

// ## EOF

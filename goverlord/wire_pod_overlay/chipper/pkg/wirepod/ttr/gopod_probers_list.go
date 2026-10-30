// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_probers_list.go
package wirepod_ttr

// The probers list (2026-10-29, operator: "I like the name 'probers list'";, operator: "Please grab those real questions
// held in saved sessions, and create a scripted 'real questions' run to test with *
// then test if its working well enough"; "Just grab a bunch of my best probers").
// system-test-probers-list plays system_tests.probers_list.file - the
// operator's own probers from his saved sessions (public: "I consider my probing
// questions as examples of how I thought"), each with the topic it should land on first -
// through the game's topic loader, the lightning (gopod_lightning.go) and the word
// triggers side by side, nothing said. It passes when the lightning lands
// pass_share of them. The live game's coverage is put back after.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	// The live game's coverage is put back after (the covered-last order reads it).
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

// ## EOF

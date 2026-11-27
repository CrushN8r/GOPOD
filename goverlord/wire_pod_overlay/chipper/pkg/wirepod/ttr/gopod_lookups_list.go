// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_lookups_list.go
package wirepod_ttr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type gopodLookupCase struct {
	who, question string
	want          []string
}

func gopodLookupsList() ([]gopodLookupCase, string, error) {
	file := filepath.Join(gopodChipperRoot(), gopodBrobotsSettings.SystemTests.LookupsList.File)
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, "", err
	}
	var out []gopodLookupCase
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		f := strings.SplitN(line, " | ", 3)
		if line == "" || strings.HasPrefix(line, "#") || len(f) != 3 {
			continue
		}
		c := gopodLookupCase{who: strings.TrimSpace(f[0]), question: strings.TrimSpace(f[1])}
		for _, w := range strings.Split(f[2], ",") {
			if w = strings.TrimSpace(w); w != "" {
				c.want = append(c.want, w)
			}
		}
		out = append(out, c)
	}
	return out, filepath.Base(file), nil
}

// gopodLookupCardsGot: the cards a look-up took, by file name, in order.
func gopodLookupCardsGot(who, question string) (string, []string) {
	fact, sources := gopodLookupFact(who, question)
	var names []string
	for _, s := range sources {
		names = append(names, filepath.Base(strings.SplitN(s, "#", 2)[0]))
	}
	return fact, names
}

// gopodLookupsListLine: system-test-lookups.
func gopodLookupsListLine() (bool, string) {
	cases, file, err := gopodLookupsList()
	if err != nil {
		return false, err.Error()
	}
	if len(cases) == 0 {
		return false, file + ": no lines"
	}
	var drift []string
	for _, c := range cases {
		_, got := gopodLookupCardsGot(c.who, c.question)
		if strings.Join(got, ", ") != strings.Join(c.want, ", ") {
			drift = append(drift, fmt.Sprintf("%s %q: %s (was %s)", c.who, gopodCut(c.question, 40), strings.Join(got, ", "), strings.Join(c.want, ", ")))
		}
	}
	line := fmt.Sprintf("%s, %d look-ups: %d as listed", file, len(cases), len(cases)-len(drift))
	right, total, misses := gopodIndexQuestionsScore()
	line += fmt.Sprintf(" · index questions: %d of %d find their card", right, total)
	if len(misses) > 0 && len(misses) <= 10 {
		line += " (missed: " + strings.Join(misses, "; ") + ")"
	}
	if len(drift) > 0 {
		line += "\n  drift: " + strings.Join(drift, "\n  drift: ")
	}
	return len(drift) == 0, line
}

// ## EOF

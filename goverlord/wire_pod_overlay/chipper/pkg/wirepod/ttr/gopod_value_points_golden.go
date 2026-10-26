// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_value_points_golden.go
package wirepod_ttr

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// GOLDEN value points - stamped and locked (2026-10-21, operator: "value points
// are what will make me money * value points must be clearly defined and
// stamped/locked as golden reference"; "chapter > topics > value points").
// A value point is what the coverage tracker already counts (gopodSplitPoints):
// each numbered line ("1. ...") after a topic file's "---", or - a topic with
// none - its whole text as one point. Each gets an id, "<topic>#<n>" (bingo#1),
// and its words go in the golden index (chapters/value_points_golden.json) with
// a checksum. system-check flags a point whose words changed, a new point not
// locked yet, and a locked point that's gone - so a point only ever changes on
// purpose: "system-value-points lock" writes the index again (system-undo puts it
// back). The index is written the first time Wire-Pod starts without one.

type gopodGoldenPoint struct {
	Words string `json:"words"`
	Sum   string `json:"sum"`
}

func gopodValuePointsGoldenPath() string {
	return filepath.Join(gopodChipperRoot(), "gopod_probes", "chapters", "value_points_golden.json")
}

func gopodPointSum(words string) string {
	s := sha1.Sum([]byte(words))
	return hex.EncodeToString(s[:6])
}

// gopodValuePointsNow: every topic's value points now, by id.
func gopodValuePointsNow() map[string]string {
	out := map[string]string{}
	for _, p := range gopodTopicRegistry() {
		facts, err := gopodReadTopicFacts(p.File)
		if err != nil {
			continue
		}
		_, points := gopodSplitPoints(string(facts))
		if len(points) == 0 {
			if w := strings.TrimSpace(string(facts)); w != "" {
				out[p.Name+"#1"] = w
			}
			continue
		}
		for i, line := range points {
			out[fmt.Sprintf("%s#%d", p.Name, i+1)] = strings.TrimSpace(gopodPointLine.ReplaceAllString(line, ""))
		}
	}
	return out
}

// gopodValuePointsGolden: the locked index, and whether there is one.
func gopodValuePointsGolden() (map[string]gopodGoldenPoint, bool) {
	var f struct {
		Points map[string]gopodGoldenPoint `json:"points"`
	}
	data, err := os.ReadFile(gopodValuePointsGoldenPath())
	if err != nil || json.Unmarshal(data, &f) != nil {
		return nil, false
	}
	return f.Points, true
}

// gopodValuePointsLock writes the golden index from the topic files now.
func gopodValuePointsLock() (int, error) {
	now := gopodValuePointsNow()
	points := map[string]gopodGoldenPoint{}
	for id, w := range now {
		points[id] = gopodGoldenPoint{Words: w, Sum: gopodPointSum(w)}
	}
	data, err := json.MarshalIndent(map[string]any{
		"about":  "GOLDEN value points (chapter > topic > value point): each topic's numbered lines after ---, or its whole text as one point, by id <topic>#<n>. Locked - system-check flags any change; system-value-points lock writes this again on purpose. gopod_value_points_golden.go.",
		"points": points,
	}, "", " ")
	if err != nil {
		return 0, err
	}
	path := gopodValuePointsGoldenPath()
	if gopodFileExists(path) {
		err = gopodWriteFileSafe(path, append(data, '\n'))
	} else {
		err = gopodWriteFile(path, append(data, '\n'), 0o644)
	}
	if err == nil {
		if _, uid, gid := gopodRecordRoot(); uid >= 0 {
			os.Chown(path, uid, gid)
		}
	}
	return len(points), err
}

// gopodValuePointsProblems: changed, new (not locked) and gone points.
func gopodValuePointsProblems() []string {
	golden, ok := gopodValuePointsGolden()
	if !ok {
		return []string{"value points: not locked yet - system-value-points lock"}
	}
	now := gopodValuePointsNow()
	var out []string
	for id, w := range now {
		g, locked := golden[id]
		switch {
		case !locked:
			out = append(out, "value point "+id+": new, not locked - system-value-points lock")
		case g.Sum != gopodPointSum(w):
			out = append(out, "value point "+id+": words changed from the golden index - system-value-points lock if on purpose")
		}
	}
	for id := range golden {
		if _, there := now[id]; !there {
			out = append(out, "value point "+id+": locked but gone from its topic - system-value-points lock if on purpose")
		}
	}
	sort.Strings(out)
	return out
}

// gopodValuePointsFirstLock: the index written the first time there is none.
func gopodValuePointsFirstLock() {
	if _, ok := gopodValuePointsGolden(); ok {
		return
	}
	if n, err := gopodValuePointsLock(); err == nil {
		logger.Println(fmt.Sprintf("GOPOD_VALUE_POINTS locked %d points (first time) - %s", n, gopodValuePointsGoldenPath()))
	}
}

// gopodSystemValuePoints: "system-value-points [lock | <topic>]".
func gopodSystemValuePoints(text string) string {
	f := strings.Fields(text)
	if len(f) > 1 && strings.EqualFold(f[1], "lock") {
		n, err := gopodValuePointsLock()
		if err != nil {
			return "system-value-points lock: " + err.Error()
		}
		return fmt.Sprintf("value points locked: %d (system-undo puts the old index back)", n)
	}
	now := gopodValuePointsNow()
	if len(f) > 1 {
		var ids []string
		for id := range now {
			if strings.HasPrefix(id, f[1]+"#") {
				ids = append(ids, id)
			}
		}
		sort.Strings(ids)
		if len(ids) == 0 {
			return "no topic " + f[1] + " (system-topic lists them)"
		}
		out := []string{fmt.Sprintf("%s - %d value points:", f[1], len(ids))}
		for _, id := range ids {
			out = append(out, "  "+id+"  "+gopodShort(now[id], 140))
		}
		return gopodJoinLines(out)
	}
	problems := gopodValuePointsProblems()
	head := fmt.Sprintf("value points: %d in %d topics - %s", len(now), len(gopodTopicRegistry()),
		map[bool]string{true: "all locked, unchanged", false: fmt.Sprintf("%d to look at", len(problems))}[len(problems) == 0])
	if len(problems) > 8 {
		problems = append(problems[:8], fmt.Sprintf("... %d more (system-check)", len(problems)-8))
	}
	return gopodJoinLines(append([]string{head, "system-value-points <topic> lists a topic's; system-value-points lock locks changes"}, problems...))
}

// ## EOF

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_probers_make.go
package wirepod_ttr

import (
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// gopodProbersMake: a prober for every value point that can have one, and how many.
func gopodProbersMake() (lines []string, made, points int) {
	pm := gopodBrobotsSettings.ProbersMake
	chapters, _, _ := gopodReadChaptersFolder()
	turn := 0
	for _, c := range chapters {
		for _, t := range c.Topics {
			pts, word, ok := gopodProbeTopic(t)
			if !ok {
				continue
			}
			for i, p := range pts {
				if !gopodCorePoint(t.Name, i+1) {
					continue
				}
				points++
				shapes := append(append([]string(nil), pm.Shapes[turn%len(pm.Shapes):]...), pm.Shapes[:turn%len(pm.Shapes)]...)
				if q := gopodProbeFor(t, pts, i, p, word, shapes); q != "" {
					lines = append(lines, q+" | "+t.Name)
					made++
				}
				turn++
			}
		}
	}
	return lines, made, points
}

// gopodProbeTopic: a topic's value points and the word its probes name it by.
func gopodProbeTopic(t gopodTopicFile) ([]string, string, bool) {
	if len(t.Words) == 0 || len(gopodBrobotsSettings.ProbersMake.Shapes) == 0 {
		return nil, "", false
	}
	facts, err := gopodReadTopicFacts(t.Path)
	if err != nil {
		return nil, "", false
	}
	_, pts := gopodSplitPoints(string(facts))
	word := t.Words[0]
	if t.Label != "" && gopodProbeLands("tell me about "+t.Label, t.Name, nil, -1) {
		word = t.Label
	} else {
		for _, w := range t.Words {
			if len(w) >= 4 {
				word = w
				break
			}
		}
	}
	return pts, word, true
}

// gopodProbeFor: the first of shapes that, filled for point i, lands on that very point; "" with none.
func gopodProbeFor(t gopodTopicFile, pts []string, i int, p, word string, shapes []string) string {
	part := "evening"
	if h := time.Now().Hour(); h >= 5 && h < 12 {
		part = "morning"
	} else if h >= 12 && h < 17 {
		part = "afternoon"
	}
	wrong := gopodBrobotsSettings.ProbersMake.WrongGreeting[part]
	title, note := gopodProbeParts(t.Name, i+1, p)
	for _, shape := range shapes {
		if strings.EqualFold(title, word) && strings.Contains(shape, "{title}") && strings.Contains(shape, "{word}") {
			continue
		}
		q := strings.NewReplacer("{title}", title, "{note}", note, "{word}", word, "{label}", t.Label, "{wrong_greeting}", wrong).Replace(shape)
		if !gopodProbeLands(q, t.Name, pts, i) {
			continue
		}
		return q
	}
	return ""
}

// gopodProbeChecked: each question already checked, and whether it landed (gopodProbeLands).
var gopodProbeChecked sync.Map

// gopodProbeLands: q's first topic is topic and, with i >= 0, it picks point i; each question is
// checked once for each state of the game's revealed points (a covered topic pulls less), then
// remembered.
func gopodProbeLands(q, topic string, pts []string, i int) bool {
	gopodCoverage.Lock()
	var game []string
	for k, v := range gopodCoverage.revealed {
		game = append(game, k+"="+strconv.Itoa(v))
	}
	gopodCoverage.Unlock()
	sort.Strings(game)
	key := strings.Join(game, ",") + "\x00" + topic + "\x00" + strconv.Itoa(i) + "\x00" + q
	if v, ok := gopodProbeChecked.Load(key); ok {
		return v.(bool)
	}
	l := loadScenarioPacketsForMessage(q)
	lands := len(l) > 0 && l[0].Name == topic && (i < 0 || gopodPickPoint(topic, pts, map[int]bool{}, q) == i)
	gopodProbeChecked.Store(key, lands)
	return lands
}

// gopodProbeParts: a point's title and note (its notes line, else its own words).
func gopodProbeParts(topic string, n int, point string) (string, string) {
	notes := gopodBrobotsSettings.ValuePoints.Notes[topic+" "+strconv.Itoa(n)]
	parts := strings.Split(notes, "|")
	if strings.TrimSpace(notes) == "" {
		text := strings.TrimSpace(gopodPointLine.ReplaceAllString(point, ""))
		if i := strings.IndexAny(text, ":.,-"); i > 0 {
			text = text[:i]
		}
		return strings.TrimSpace(text), strings.TrimSpace(text)
	}
	title := strings.TrimSpace(parts[0])
	note := title
	if len(parts) > 1 {
		note = strings.TrimSpace(parts[1])
	}
	return title, note
}

// gopodSystemProbersMake: system-probers-make - the chain run, its file written.
func gopodSystemProbersMake() string {
	lines, made, points := gopodProbersMake()
	head := []string{
		"# Made by the probe maker (system-probers-make, gopod_probers_make.go): a prober",
		"# for each value point - topic, point, shape, fill, checked. Same format as the",
		"# probers list: the question, \" | \", the topic it lands on first.",
	}
	file := filepath.Join(gopodChipperRoot(), gopodBrobotsSettings.ProbersMake.File)
	if err := gopodWriteFileSafe(file, []byte(strings.Join(append(head, lines...), "\n")+"\n")); err != nil {
		return gopodSystemText("probers_make.not_written", "error", err.Error())
	}
	return gopodSystemText("probers_make.made", "made", strconv.Itoa(made), "points", strconv.Itoa(points), "file", gopodBrobotsSettings.ProbersMake.File)
}

// ## EOF

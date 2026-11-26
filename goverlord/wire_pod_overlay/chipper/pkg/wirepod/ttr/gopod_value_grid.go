// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_value_grid.go
package wirepod_ttr

// The value grid (2026-11-25, operator: "repo md file = chapter > topics >
// value-points (why lean into GOPOD) * chapter > chapter value-points *
// per-topic > topic value-points * per-value-point > value-point-special-notes";
// "locally layered value-points within the given boundaries, should also be
// interconnected/cross-linked for look-ups ... a tighter more accurate 2D/XY grid
// approach"): the merry-go-round as a grid, read live from the cards - never
// written by hand. X: the seats (chapters, then their topics) in the game's
// order - the wide sweep. Y: each topic's value points, 1 (wide) to n (deep) -
// the deep sweep. Each point carries its special notes, its telling words and
// its links: the points in other topics that share the most of those words.
// The reminders ride from it (gopodValueGridBounce: the operator's pick, 2026-11-25 -
// "In the reminder"), and the start-up check checks it (gopodValueGridCheck).

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// gopodGridPoint: one cell - a topic's value point.
type gopodGridPoint struct {
	Topic, Label string
	Depth, Of    int
	Say, Notes   string
	Words        map[string]bool
	Links        []string // "topic #n", strongest first
}

// gopodGridTopic: a seat's topic and its points.
type gopodGridTopic struct {
	Name, Label, Path string
	Points            []*gopodGridPoint
}

// gopodGridChapter: a seat - a chapter, its own value point (its summary), its topics.
type gopodGridChapter struct {
	Title, Summary string
	Topics         []gopodGridTopic
}

// gopodValueGridLinks: how many links a point keeps.
const gopodValueGridLinks = 3

// gopodValueGrid builds the grid from the cards, links included.
func gopodValueGrid() []gopodGridChapter {
	chapters, _, _ := gopodReadChaptersFolder()
	var out []gopodGridChapter
	var all []*gopodGridPoint
	for _, c := range chapters {
		gc := gopodGridChapter{Title: c.Title, Summary: c.Summary}
		for _, t := range c.Topics {
			label := t.Label
			if label == "" {
				label = t.Name
			}
			gt := gopodGridTopic{Name: t.Name, Label: label, Path: t.Path}
			var ns []int
			for n := range t.SayLines {
				ns = append(ns, n)
			}
			for n := range t.Notes {
				if _, ok := t.SayLines[n]; !ok {
					ns = append(ns, n)
				}
			}
			sort.Ints(ns)
			// 2026-11-26 (the grid flagged 18 topics as having no value points): the
			// game's rule (gopodCoverageRevealFor) - a card with no numbered points is
			// one point, its whole text after "---". The grid follows it.
			if len(ns) == 0 {
				if data, err := gopodTopicFileRead(t.Path); err == nil {
					if _, body, ok := strings.Cut(string(data), "\n---\n"); ok && strings.TrimSpace(body) != "" {
						p := &gopodGridPoint{Topic: t.Name, Label: label, Depth: 1, Of: 1, Notes: strings.Join(strings.Fields(body), " ")}
						p.Words = gopodLookupMeaningful(p.Notes)
						gt.Points = append(gt.Points, p)
						all = append(all, p)
					}
				}
			}
			for _, n := range ns {
				p := &gopodGridPoint{Topic: t.Name, Label: label, Depth: n, Of: len(ns), Say: t.SayLines[n], Notes: t.Notes[n]}
				p.Words = gopodLookupMeaningful(p.Say + " " + strings.ReplaceAll(p.Notes, "|", " "))
				gt.Points = append(gt.Points, p)
				all = append(all, p)
			}
			gc.Topics = append(gc.Topics, gt)
		}
		out = append(out, gc)
	}
	// Links: the points in other topics sharing the most telling words (two or more).
	for _, p := range all {
		type cand struct {
			key   string
			score int
		}
		var cs []cand
		for _, q := range all {
			if q.Topic == p.Topic {
				continue
			}
			score := 0
			for w := range p.Words {
				if q.Words[w] {
					score++
				}
			}
			if score >= 2 {
				cs = append(cs, cand{q.Topic + " #" + strconv.Itoa(q.Depth), score})
			}
		}
		sort.SliceStable(cs, func(i, j int) bool { return cs[i].score > cs[j].score })
		for i := 0; i < len(cs) && i < gopodValueGridLinks; i++ {
			p.Links = append(p.Links, cs[i].key)
		}
	}
	return out
}

// gopodValueGridMarkdown: the grid as a page to read - chapter, its value
// point, then each topic's points (depth, the point, its special notes, its links).
func gopodValueGridMarkdown(when string) string {
	var b strings.Builder
	grid := gopodValueGrid()
	points, links := 0, 0
	for _, c := range grid {
		for _, t := range c.Topics {
			points += len(t.Points)
			for _, p := range t.Points {
				if len(p.Links) > 0 {
					links++
				}
			}
		}
	}
	fmt.Fprintf(&b, "# GOPOD value grid - the merry-go-round, from the cards (%s)\n\n", when)
	fmt.Fprintf(&b, "Generated from the cards, never written by hand. %d chapters (seats) · %d value points · %d linked to another topic.\n", len(grid), points, links)
	b.WriteString("X: the seats in the game's order (wide). Y: each topic's points, 1 (wide) to n (deep). Links: points in other topics sharing the most telling words.\n")
	for ci, c := range grid {
		fmt.Fprintf(&b, "\n## Seat %d: %s\n\n", ci+1, c.Title)
		if strings.TrimSpace(c.Summary) != "" {
			fmt.Fprintf(&b, "**Chapter value point:** %s\n", strings.TrimSpace(c.Summary))
		}
		for _, t := range c.Topics {
			fmt.Fprintf(&b, "\n### %s (`%s`)\n\n", t.Label, t.Name)
			if len(t.Points) == 0 {
				b.WriteString("_No value points on this card._\n")
				continue
			}
			b.WriteString("| depth | value point | special notes | links |\n|---|---|---|---|\n")
			cell := func(s string) string {
				return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(s), "|", "·"), "\n", " ")
			}
			for _, p := range t.Points {
				fmt.Fprintf(&b, "| %d of %d | %s | %s | %s |\n", p.Depth, p.Of, cell(p.Say), cell(p.Notes), cell(strings.Join(p.Links, ", ")))
			}
		}
	}
	return b.String()
}

var gopodValueGridCache = struct {
	sync.Mutex
	at   time.Time
	grid []gopodGridChapter
}{}

// gopodValueGridNow: the grid, read again after a minute (cards change live).
func gopodValueGridNow() []gopodGridChapter {
	gopodValueGridCache.Lock()
	defer gopodValueGridCache.Unlock()
	if gopodValueGridCache.grid == nil || time.Since(gopodValueGridCache.at) > time.Minute {
		gopodValueGridCache.grid, gopodValueGridCache.at = gopodValueGrid(), time.Now()
	}
	return gopodValueGridCache.grid
}

// gopodValueGridBounce (2026-11-25, operator: "a loose happy glitch can lead to
// something humans can latch onto * the pinball can bounce around, but its
// contained"): where a question bounces to - the point it lands on (the one
// sharing the most telling words with it, in the topics it loaded when it loaded
// any), then that point's strongest link whose topic is a seat. "" for none.
func gopodValueGridBounce(question string, seat func(string) bool) string {
	words := gopodLookupMeaningful(question)
	if len(words) == 0 {
		return ""
	}
	loaded := map[string]bool{}
	for _, p := range loadScenarioPacketsForMessage(question) {
		loaded[p.Name] = true
	}
	var land *gopodGridPoint
	best := 0
	for _, c := range gopodValueGridNow() {
		for _, t := range c.Topics {
			if len(loaded) > 0 && !loaded[t.Name] {
				continue
			}
			for _, p := range t.Points {
				score := 0
				for w := range words {
					if p.Words[w] {
						score++
					}
				}
				if score > best {
					land, best = p, score
				}
			}
		}
	}
	if land == nil {
		return ""
	}
	for _, l := range land.Links {
		topic := strings.TrimSpace(strings.SplitN(l, " #", 2)[0])
		if seat(topic) {
			return topic
		}
	}
	return ""
}

// gopodValueGridCheck: what's wrong with a grid - breaks (a point with neither
// words nor notes, a link to a point that isn't there) and gaps (a topic with no
// value points on its card; one line naming the points that link to nothing -
// they can't be bounced to). Breaks are "to fix"; gaps are the cards' to-do, logged.
// Its last value is how many things it checked.
func gopodValueGridCheck(grid []gopodGridChapter) (breaks, gaps []string, checked int) {
	var out []string
	exists := map[string]bool{}
	for _, c := range grid {
		for _, t := range c.Topics {
			for _, p := range t.Points {
				exists[p.Topic+" #"+strconv.Itoa(p.Depth)] = true
			}
		}
	}
	var lonely []string
	for _, c := range grid {
		for _, t := range c.Topics {
			checked++
			if len(t.Points) == 0 {
				gaps = append(gaps, "grid: topic "+t.Name+" ("+c.Title+") has no value points on its card")
				continue
			}
			for _, p := range t.Points {
				checked++
				key := p.Topic + " #" + strconv.Itoa(p.Depth)
				if strings.TrimSpace(p.Say) == "" && strings.TrimSpace(p.Notes) == "" {
					out = append(out, "grid: "+key+" has no words")
				}
				for _, l := range p.Links {
					if !exists[l] {
						out = append(out, "grid: "+key+" links to "+l+", which isn't there")
					}
				}
				if len(p.Links) == 0 {
					lonely = append(lonely, key)
				}
			}
		}
	}
	if len(lonely) > 0 {
		gaps = append(gaps, fmt.Sprintf("grid: %d points link to nothing (no shared words with another topic): %s", len(lonely), strings.Join(lonely, ", ")))
	}
	return out, gaps, checked
}

// ## EOF

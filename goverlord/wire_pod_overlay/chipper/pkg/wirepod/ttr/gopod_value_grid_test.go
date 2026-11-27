package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestValueGrid: the grid holds every topic's value points from the cards; GOPOD_DUMP_GRID=<file>
// writes it as a page (private notes, never the repo).
func TestValueGrid(t *testing.T) {
	grid := gopodValueGrid()
	if len(grid) == 0 {
		t.Fatal("no chapters in the grid")
	}
	points := 0
	for _, c := range grid {
		for _, tp := range c.Topics {
			points += len(tp.Points)
		}
	}
	if points < 100 {
		t.Fatalf("only %d value points in the grid", points)
	}
	if out := os.Getenv("GOPOD_DUMP_GRID"); out != "" {
		page := gopodValueGridMarkdown(time.Now().Format("2006-01-02 15:04"))
		if err := os.WriteFile(filepath.Clean(out), []byte(page), 0644); err != nil {
			t.Fatal(err)
		}
		t.Logf("grid written to %s (%d lines)", out, strings.Count(page, "\n"))
	}
}

// TestValueGridCheckCatchesBreaks: the checker is fed a broken grid - a topic with no points, a
// point with no words, a dead link, a lonely point - and must name each one; a sound grid gets no
// break lines.
func TestValueGridCheckCatchesBreaks(t *testing.T) {
	good := &gopodGridPoint{Topic: "a", Depth: 1, Say: "alpha point", Links: []string{"b #1"}}
	other := &gopodGridPoint{Topic: "b", Depth: 1, Say: "beta point", Links: []string{"a #1"}}
	sound := []gopodGridChapter{{Title: "One", Topics: []gopodGridTopic{{Name: "a", Points: []*gopodGridPoint{good}}, {Name: "b", Points: []*gopodGridPoint{other}}}}}
	if b, g, n := gopodValueGridCheck(sound); len(b) != 0 || len(g) != 0 || n != 4 {
		t.Fatalf("a sound grid: %v %v (%d checked)", b, g, n)
	}
	empty := &gopodGridPoint{Topic: "c", Depth: 1}
	dead := &gopodGridPoint{Topic: "d", Depth: 1, Say: "delta", Links: []string{"zz #9"}}
	broken := []gopodGridChapter{{Title: "Two", Topics: []gopodGridTopic{
		{Name: "none"},
		{Name: "c", Points: []*gopodGridPoint{empty}},
		{Name: "d", Points: []*gopodGridPoint{dead}},
	}}}
	b, g, _ := gopodValueGridCheck(broken)
	all := strings.Join(append(b, g...), "\n")
	if len(b) != 2 || len(g) != 2 {
		t.Fatalf("breaks %v, gaps %v - want 2 of each", b, g)
	}
	for _, want := range []string{"topic none (Two) has no value points", "c #1 has no words", "d #1 links to zz #9, which isn't there", "link to nothing", "c #1"} {
		if !strings.Contains(all, want) {
			t.Fatalf("the checker missed %q:\n%s", want, all)
		}
	}
}

// TestValueGridReal: the grid the cards make has no dead links and no empty points; the bounce from
// a point lands on one of its links.
func TestValueGridReal(t *testing.T) {
	for _, line := range func() []string { b, _, _ := gopodValueGridCheck(gopodValueGrid()); return b }() {
		if strings.Contains(line, "isn't there") || strings.Contains(line, "has no words") {
			t.Errorf("real grid: %s", line)
		}
	}
	labels := gopodBrobotsSettings.InterviewChapters.TopicLabels
	seat := func(topic string) bool { return labels[topic] != "" && topic != "doc_and_pip" }
	got := gopodValueGridBounce("Doc explains too hard, Pip understands too late - the audience laughs, then learns", seat)
	if got == "" {
		t.Fatal("no bounce from Doc and Pip's third point")
	}
	var land *gopodGridPoint
	for _, c := range gopodValueGrid() {
		for _, tp := range c.Topics {
			for _, p := range tp.Points {
				if p.Topic == "doc_and_pip" && p.Depth == 3 {
					land = p
				}
			}
		}
	}
	if land == nil || !strings.Contains(strings.Join(land.Links, " "), got) {
		t.Fatalf("bounced to %q, not one of doc_and_pip #3's links %v", got, land.Links)
	}
}

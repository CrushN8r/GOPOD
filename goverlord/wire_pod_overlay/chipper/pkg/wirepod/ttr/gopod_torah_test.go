// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_torah_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestTorahMatches: GOPOD_SYSTEM_SCROLLS.md is the system as it is now (GOPOD_WRITE_TORAH=1 writes
// it again); every system window command is in it.
func TestTorahMatches(t *testing.T) {
	if _, err := os.Stat(filepath.Join(gopodRepoRoot(), "README.md")); err != nil {
		t.Skip("not in the repo here")
	}
	page := strings.Join(func() []string {
		var all []string
		for _, v := range gopodTorahFiles() {
			all = append(all, v)
		}
		return all
	}(), "\n")
	for _, want := range []string{"# GOPOD SYSTEM SCROLLS", "system-test-admin-run", "system-comm-check", "start bingo test", "torah"} {
		if !strings.Contains(page, want) {
			t.Errorf("the Scrolls has no %q", want)
		}
	}
	for i, sc := range gopodTorahScriptures() {
		if !strings.HasPrefix(sc.Title, "Scripture "+strconv.Itoa(i+1)+",") {
			t.Errorf("scripture %d: %q", i+1, sc.Title)
		}
	}
	if os.Getenv("GOPOD_WRITE_TORAH") == "1" {
		if err := gopodWriteTorah(); err != nil {
			t.Fatal(err)
		}
		return
	}
	if gopodTorahBehind() {
		t.Errorf("GOPOD_SYSTEM_SCROLLS.md or a scripture (system_scrolls/) doesn't match the system - GOPOD_WRITE_TORAH=1 (or system-scrolls) writes it again")
	}
}

// ## EOF

// the README's cast table is the numpad map's - GOPOD writes it (GOPOD_WRITE_TORAH=1 or
// system-scrolls); it fails when behind.
func TestReadmeCastMatches(t *testing.T) {
	if os.Getenv("GOPOD_WRITE_TORAH") == "1" {
		if err := gopodWriteReadmeCast(); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, ok := gopodReadmeCastWant(); !ok {
		t.Fatal("README.md has no cast table markers")
	}
	if gopodReadmeCastBehind() {
		t.Errorf("README.md's cast table doesn't match the numpad map - GOPOD_WRITE_TORAH=1 (or system-scrolls) writes it again")
	}
	table := gopodReadmeCastTable()
	for _, want := range []string{"**DOC** - DOCTRINE BARFALLONYOU SPEWER, on Brobot 1", "| KP1, NumLock off | **BRUH-1** - Brobot 1 |", "**All the Brobots**"} {
		if !strings.Contains(table, want) {
			t.Errorf("the cast table has no %q:\n%s", want, table)
		}
	}
}

// the tour is the scrolls' folders, in order - scripture 0 the intro, the lingo improver a stop.
func TestTourIsTheScrollsTree(t *testing.T) {
	if _, err := os.Stat(filepath.Join(gopodRepoRoot(), "README.md")); err != nil {
		t.Skip("not in the repo here")
	}
	stops := gopodTourScrolls()
	if len(stops) < 2 || !strings.HasPrefix(stops[0].Title, "Scripture 0,") || !strings.HasPrefix(stops[1].Title, "Scripture 1,") {
		t.Fatalf("tour stops: %d, first %q", len(stops), stops[0].Title)
	}
	lingo := false
	for _, s := range stops {
		if strings.Contains(strings.Join(s.Lines, " "), "system-lingo-improver") {
			lingo = true
		}
	}
	if !lingo {
		t.Error("the lingo improver isn't a stop of the tour")
	}
	if strings.TrimSpace(gopodSystemText("lingo_improver.open")) == "" || !strings.Contains(gopodSystemText("lingo_improver.open"), "gopod_lingo_improver.html") {
		t.Error("system-lingo-improver has no link")
	}
}

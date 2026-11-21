// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_lookups_list_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLookupsList (2026-11-20, operator: "a cheap system test run to catch
// look-up drift"): every line of the lookups list looks up the cards it lists.
// GOPOD_WRITE_LOOKUPS=1 writes the cards as they are now; GOPOD_DUMP_LOOKUPS=1
// writes each line's exact FACT and cards to gopod_notes (private, never the repo).
func TestLookupsList(t *testing.T) {
	cases, _, err := gopodLookupsList()
	if err != nil || len(cases) == 0 {
		t.Fatalf("lookups list: %v (%d lines)", err, len(cases))
	}
	file := filepath.Join(gopodChipperRoot(), gopodBrobotsSettings.SystemTests.LookupsList.File)
	if os.Getenv("GOPOD_WRITE_LOOKUPS") == "1" {
		data, _ := os.ReadFile(file)
		var out []string
		for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
			f := strings.SplitN(line, " | ", 3)
			if len(f) == 3 && !strings.HasPrefix(line, "#") {
				_, got := gopodLookupCardsGot(strings.TrimSpace(f[0]), strings.TrimSpace(f[1]))
				line = f[0] + " | " + f[1] + " | " + strings.Join(got, ", ")
			}
			out = append(out, line)
		}
		os.WriteFile(file, []byte(strings.Join(out, "\n")+"\n"), 0o644)
		return
	}
	if os.Getenv("GOPOD_DUMP_LOOKUPS") == "1" {
		var b strings.Builder
		b.WriteString("# Look-up dump (TestLookupsList, GOPOD_DUMP_LOOKUPS=1)\n")
		for _, c := range cases {
			fact, got := gopodLookupCardsGot(c.who, c.question)
			b.WriteString("\n## " + c.who + ": " + c.question + "\n- cards: " + strings.Join(got, ", ") + "\n- FACT:\n\n" + fact + "\n")
		}
		home, _ := os.UserHomeDir()
		os.WriteFile(filepath.Join(home, "crushn8r_git", "gopod_notes", "LOOKUP_DUMP_001.md"), []byte(b.String()), 0o644)
	}
	if ok, line := gopodLookupsListLine(); !ok {
		t.Error(line)
	}
}

// ## EOF

// 2026-11-20 (live: CAMEO said "Goverlord (aka ME)" - nine linked cards' lines
// fed in): the wide answer is the card asked about; its linked cards are listed, not said.
func TestLinkedCardsShareOneBudget(t *testing.T) {
	fact, src := gopodLookupFact("CAMEO", "who all works in your company?")
	// The company card answers it (and it's already in her prompt - no words again).
	if len(src) == 0 || src[0] != "math-aftermath-news-channel.md" || strings.Contains(fact, "I think alone") {
		t.Errorf("cards %v, fact %q", src, gopodCut(fact, 200))
	}
}

// ## EOF

// 2026-11-20 (operator: "The company has GOPOD reps who demo the GOPOD repo via a
// game engine - via self look-ups that let the GOPOD brain give a tour of
// itself"; "declare it and ensure its implemented that way"): a rep's question
// gets a game topic told from a repo file, and the brain's look-up of the repo.
func TestRepsTourTheRepo(t *testing.T) {
	q := "tell me about the rig"
	packets := gopodTopicPackets(q)
	if len(packets) == 0 || !strings.Contains(packets[0].File, "MOBILE_GEAR.md") {
		t.Errorf("game topic for %q: %v", q, packets)
	}
	if _, src := gopodLookupFact("Brobot 1", q); len(src) == 0 || src[0] != "MOBILE_GEAR.md" {
		t.Errorf("look-up for %q: %v", q, src)
	}
	if _, src := gopodLookupFact("Brobot 2", "so what's new with you?"); len(src) == 0 || src[0] != "README.md" {
		t.Errorf("no topic: the README, got %v", src)
	}
}

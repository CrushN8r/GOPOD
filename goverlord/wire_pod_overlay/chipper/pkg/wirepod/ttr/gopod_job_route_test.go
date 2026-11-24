package wirepod_ttr

import (
	"strings"
	"testing"
)

// TestJobRouteThreeStrikes (2026-11-23, the operator's loop): CAMEO asked about
// booking - the GOPOD reps' job - once names the rep, twice the rep and README.md,
// a third time hands the line to the rep, whose look-up is README.md's welcome.
func TestJobRouteThreeStrikes(t *testing.T) {
	bc := &gopodBrobotsSettings.BrainChain
	if bc.RouteRep == "" || bc.RouteOther == "" || bc.RouteAgain == "" || bc.RouteWelcome == "" {
		t.Fatal("brain_chain.route_* lines are needed (Goverlord's card)")
	}
	bc.RouteHandoff = "" // said on the third strike - not in a test
	gopodJobRouteReset()
	defer gopodJobRouteReset()
	q := "who do I talk to about booking you?"
	if owner := gopodAnyOwner(q); owner != "Brobot 1" {
		t.Fatalf("booking's owner = %q, want Brobot 1 (booking.md's owner line)", owner)
	}
	if who := gopodJobRoute("CAMEO", q); who != "CAMEO" {
		t.Fatalf("strike 1: %q answers, want CAMEO", who)
	}
	if n := gopodJobRouteNote("CAMEO"); !strings.Contains(n, "Brobot 1") || !strings.Contains(n, "Not my job") {
		t.Fatalf("strike 1 ROUTE line: %q", n)
	}
	gopodJobRoute("CAMEO", q)
	if n := gopodJobRouteNote("CAMEO"); !strings.Contains(n, "README.md") {
		t.Fatalf("strike 2 ROUTE line: %q", n)
	}
	if who := gopodJobRoute("CAMEO", q); who != "Brobot 1" {
		t.Fatalf("strike 3: %q answers, want Brobot 1", who)
	}
	fact, sources := gopodLookupFact("Brobot 1", q)
	if len(sources) == 0 || !strings.EqualFold(sources[0], "README.md") || !strings.Contains(fact, "welcome") {
		t.Fatalf("strike 3's look-up: %v / %q - want README.md's welcome", sources, fact)
	}
	// His own job: no ROUTE line.
	gopodJobRoute("PLAYHEAD", "Tell me about your interns")
	if n := gopodJobRouteNote("PLAYHEAD"); n != "" {
		t.Fatalf("PLAYHEAD on his interns got a ROUTE line: %q", n)
	}
}

package wirepod_ttr

import (
	"strings"
	"testing"
)

// TestJobRouteThreeStrikes (2026-11-23, the operator's loop): CAMEO asked about
// the weather - the GOPOD rep's job - once names the rep, twice the rep and
// README.md, a third time hands the line to the rep, whose look-up is README.md's
// welcome. (Booking and contact are every staffer's - 2026-11-24.)
func TestJobRouteThreeStrikes(t *testing.T) {
	bc := &gopodBrobotsSettings.BrainChain
	if bc.RouteRep == "" || bc.RouteOther == "" || bc.RouteAgain == "" || bc.RouteWelcome == "" {
		t.Fatal("brain_chain.route_* lines are needed (Goverlord's card)")
	}
	bc.RouteHandoff = "" // said on the third strike - not in a test
	gopodJobRouteReset()
	defer gopodJobRouteReset()
	q := "what's the weather like today?"
	if owner := gopodAnyOwner(q); owner != "Brobot 1" {
		t.Fatalf("the weather's owner = %q, want Brobot 1 (weather.md's owner line)", owner)
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
	// Contact info: any staffer gives it.
	if who := gopodJobRoute("CAMEO", "how do I contact you?"); who != "CAMEO" || gopodJobRouteNote("CAMEO") != "" {
		t.Fatalf("contact info to CAMEO: %q answers, ROUTE %q", who, gopodJobRouteNote("CAMEO"))
	}
	// His own job: no ROUTE line.
	gopodJobRoute("PLAYHEAD", "Tell me about your interns")
	if n := gopodJobRouteNote("PLAYHEAD"); n != "" {
		t.Fatalf("PLAYHEAD on his interns got a ROUTE line: %q", n)
	}
}

// TestContactLine (2026-11-24; 2026-11-25 the QR card): a contact or booking ask
// gets the contact line, a keeping-up ask the keep-up line, word for word, from
// any persona - after his own line, once.
func TestContactLine(t *testing.T) {
	gy := gopodBrobotsSettings.GopodYourselfQr
	line := gy.ContactLine
	if line == "" || gopodContactAddress(line) == "" || gy.KeepUpLine == "" {
		t.Fatal("gopod_yourself_qr.contact_line with an email and keep_up_line are needed (gopod_yourself_qr.md)")
	}
	for _, q := range []string{"how do I contact you?", "who do I talk to about booking GOPOD?", "can I hire the Brobots?"} {
		if got := gopodYourselfStitch(q, "Sure thing!"); !strings.HasPrefix(got, "Sure thing!") || !strings.HasSuffix(got, line) {
			t.Fatalf("%q: %q", q, got)
		}
	}
	if got := gopodYourselfStitch("how do I contact you?", "Mail "+gopodContactAddress(line)+" anytime."); strings.Contains(got, line) {
		t.Fatalf("contact line said twice: %q", got)
	}
	for _, q := range []string{"Brobot 2, how do I keep up with you guys?", "how do I subscribe?"} {
		if got := gopodYourselfStitch(q, "Oh yeah!"); !strings.HasPrefix(got, "Oh yeah!") || !strings.HasSuffix(got, gy.KeepUpLine) {
			t.Fatalf("%q: %q", q, got)
		}
	}
	if got := gopodYourselfStitch("what's the weather?", "Sunny."); got != "Sunny." {
		t.Fatalf("weather got a golden line: %q", got)
	}
}

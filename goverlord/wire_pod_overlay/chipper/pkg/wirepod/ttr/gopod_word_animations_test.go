package wirepod_ttr

import (
	"strings"
	"testing"
)

// The words a Brobot says pick its animation (operator 2026-10-02).
func TestGopodAnimationsFromWords(t *testing.T) {
	acts := []RobotAction{
		{Action: ActionPlayAnimationWI, Parameter: "thinking"},
		{Action: ActionSayText, Parameter: "party, time!! Ugh fine."},
		{Action: ActionSayText, Parameter: "No words of note here."},
		{Action: ActionSayText, Parameter: "I'm so grumpy right now."},
	}
	got := gopodAnimationsFromWords(acts)
	var seq []string
	for _, a := range got {
		if a.Action == ActionPlayAnimationWI {
			seq = append(seq, "anim:"+a.Parameter)
		} else {
			seq = append(seq, "say")
		}
	}
	if want := "anim:celebrate say say anim:frustrated say"; strings.Join(seq, " ") != want {
		t.Fatalf("got %q, want %q", strings.Join(seq, " "), want)
	}
	// "party... time!!": celebrate plays, "party" is said and shown.
	r := buildGOPODStreamResult(gopodAnimationMarkerToInternal("party... time!! Ugh fine."))
	if !strings.HasPrefix(r.FilteredPacket, "celebrate... party... time!!") {
		t.Fatalf("filtered = %q", r.FilteredPacket)
	}
	if got := gopodDisplayReply("party___ time!! Ugh fine."); got != "party... time!! Ugh fine." {
		t.Fatalf("bubble = %q", got)
	}
	// A token name stays a cue only.
	if got := gopodDisplayReply("happy___ Hi there."); got != "Hi there." {
		t.Fatalf("token bubble = %q", got)
	}
}

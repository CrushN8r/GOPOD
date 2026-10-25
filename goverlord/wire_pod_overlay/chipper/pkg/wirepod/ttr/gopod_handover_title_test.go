package wirepod_ttr

import (
	"strings"
	"testing"
)

// 2026-10-25 (operator: "I tend to use: Mr [funny-nickname] * randomly mix up
// Mr vs Mrs"): the hand-over line says Mr or Mrs, then a nickname.
func TestHandoverSaysMrOrMrs(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 60; i++ {
		l := gopodHandoverLine("Brobot 2", "Doc")
		if !strings.HasPrefix(l, "Doc, what say you on this, Mr ") && !strings.HasPrefix(l, "Doc, what say you on this, Mrs ") {
			t.Fatalf("hand-over line %q", l)
		}
		seen[strings.Fields(l)[6]] = true
	}
	if !seen["Mr"] || !seen["Mrs"] {
		t.Fatalf("both titles not used: %v", seen)
	}
}

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_brother_repeat_test.go
package wirepod_ttr

import "testing"

// TestDropBrotherRepeat: Brobot 2's follow-up drops what only repeats Brobot 1,
// keeps what adds something ("a man vs a man with a hat is fine").
func TestDropBrotherRepeat(t *testing.T) {
	b1 := "Bingo night at the seniors' residence is wild, bro.\nI call the numbers and the room goes nuts."
	b2 := "Bingo night at the seniors' residence is wild! I call numbers in a party hat.\nThe room goes nuts, bro. Bro!"
	want := "I call numbers in a party hat.\nBro!"
	if got := gopodDropBrotherRepeat(b2, b1); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := gopodDropBrotherRepeat("All new words here.", ""); got != "All new words here." {
		t.Errorf("no brother reply: %q", got)
	}
}

// ## EOF

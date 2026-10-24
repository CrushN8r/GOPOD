package wirepod_ttr

import "testing"

// Live 2026-09-30 openings: animation words plus a wide gap are dropped; an
// ordinary opening with a single space is left alone.
func TestGopodStripLeadingAnimationWords(t *testing.T) {
	tests := map[string]string{
		"celebrate  Oh man! Get ready.":           "Oh man! Get ready.",
		"🎉 celebrate So what's the plan, man? 😎":  "🎉 So what's the plan, man? 😎",
		"excited love   okay, so like, let's go!": "okay, so like, let's go!",
		"thinking  well, i guess it's a test-run": "well, i guess it's a test-run",
		"veryHappy   Man oh man! The Party robot": "Man oh man! The Party robot",
		"curious dartingEyes Like...  think":      "Like...  think",
		"happy  Bingo is awesome!!!":              "Bingo is awesome!!!",
		"Happy to help!  Really.":                 "Happy to help!  Really.",
		"Oh wow  okay":                            "Oh wow  okay",
		"Line one.\nhappy  Line two starts here.": "Line one.\nLine two starts here.",
		"thinking This is a bit much!  Ugh!":      "This is a bit much!  Ugh!",
		"happy to help, I am.":                    "happy to help, I am.",
		"Thinking This over":                      "Thinking This over",
	}
	for in, want := range tests {
		if got := gopodStripLeadingAnimationWords(in); got != want {
			t.Fatalf("gopodStripLeadingAnimationWords(%q) = %q, want %q", in, got, want)
		}
	}
}

// The one-dot opening cue (live 2026-09-30 21:47-21:55) is dropped from the
// chat display; a capitalised word or a non-animation word is kept.
func TestGopodStripAnimationCueForDisplay(t *testing.T) {
	tests := map[string]string{
		"happy. hey there!  so, uhh":         "hey there!  so, uhh",
		"thinking. double the fun, huh?":     "double the fun, huh?",
		"happy.  Man, that was a wild round": "Man, that was a wild round",
		"thinking... ok":                     "ok",
		"Happy. Me? Always.":                 "Happy. Me? Always.",
		"ok. so here goes":                   "ok. so here goes",
		"wow. the audience":                  "wow. the audience",
	}
	for in, want := range tests {
		if got := gopodStripAnimationCueForDisplay(in); got != want {
			t.Fatalf("gopodStripAnimationCueForDisplay(%q) = %q, want %q", in, got, want)
		}
	}
}

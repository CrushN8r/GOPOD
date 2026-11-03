package wirepod_ttr

import "testing"

// 2026-11-03 (live: Doc's bubble showed "darting Eyes... Here's my take"): a cue
// name split by a space is joined back before the bubble and the robot get it.
func TestJoinSplitCues(t *testing.T) {
	for in, want := range map[string]string{
		"darting Eyes___ Here's my take": "dartingEyes___ Here's my take",
		"darting Eyes... Here":           "dartingEyes... Here",
		"so darting eyes everywhere":     "so darting eyes everywhere",
	} {
		if got := gopodJoinSplitCues(in); got != want {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
}

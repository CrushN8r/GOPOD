package wirepod_ttr

import "testing"

// Doc's bubble showed "darting Eyes...
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

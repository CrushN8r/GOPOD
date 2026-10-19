package wirepod_ttr

import "testing"

// Extra animations go after the 2nd, 4th and 6th sentence of a part, never
// after its last sentence, and a short part gets none.
func TestSprinkleAnimationsUpToThreePerPart(t *testing.T) {
	part := func(n int) []RobotAction {
		var acts []RobotAction
		for i := 0; i < n; i++ {
			acts = append(acts, RobotAction{Action: ActionSayText, Parameter: "Sentence."})
		}
		return acts
	}
	extras := func(acts []RobotAction) (count int, last bool) {
		for i, a := range acts {
			if a.Action == ActionPlayAnimationWI {
				count++
				last = i == len(acts)-1
			}
		}
		return
	}
	for sentences, want := range map[int]int{2: 0, 3: 1, 4: 1, 5: 2, 6: 2, 7: 3, 10: 3} {
		got, last := extras(gopodSprinkleAnimations(part(sentences)))
		if got != want || last {
			t.Fatalf("%d sentences: %d extra animations (last=%v), want %d", sentences, got, last, want)
		}
	}
}

// The native lead-in after an animation survives GOPOD's speech cleanup, which
// says every other "..." as a comma pause (native Wire-Pod:
// "{{playAnimationWI||happy}}... text").
func TestAnimationLeadInReachesTheVoice(t *testing.T) {
	if got := gopodKeepLeadIn(gopodAnimationLeadIn+" Hello... there, bro.", cleanRobotSpeechText); got != gopodAnimationLeadIn+" Hello, there, bro." {
		t.Fatalf("lead-in = %q", got)
	}
	if got := gopodKeepLeadIn("Hello... there.", cleanRobotSpeechText); got != "Hello, there." {
		t.Fatalf("no lead-in = %q", got)
	}
	acts := GetActionsFromString("{{playAnimationWI||happy}}" + gopodAnimationLeadIn + " Hey there!")
	if len(acts) != 2 || acts[0].Action != ActionPlayAnimationWI || acts[1].Parameter != gopodAnimationLeadIn+" Hey there!" {
		t.Fatalf("song beat actions = %+v", acts)
	}
}

// ## EOF

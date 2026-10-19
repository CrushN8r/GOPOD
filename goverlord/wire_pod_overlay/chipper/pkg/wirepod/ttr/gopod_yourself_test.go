// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_yourself_test.go
package wirepod_ttr

import "testing"

// The wake phrase on its own (after "hey vector" or a name) is the command; a
// question about it is not.
func TestGOPODYourselfPhrase(t *testing.T) {
	for _, s := range []string{"GOPOD Yourself", "gopod yourself!", "go pod yourself", "Hey Vector, GOPOD Yourself", "Brobot 2, GOPOD Yourself."} {
		if !gopodIsYourself(s) {
			t.Errorf("%q should be the command", s)
		}
	}
	for _, s := range []string{"what is GOPOD Yourself?", "tell me how to GOPOD Yourself", "GOPOD", "yourself"} {
		if gopodIsYourself(s) {
			t.Errorf("%q should go to the chat", s)
		}
	}
	g := gopodBrobotsSettings.GOPODYourself
	if g.Lines["Brobot 1"] == "" || g.Lines["Brobot 2"] == "" || g.RepeatSeconds <= 0 {
		t.Error("gopod_yourself needs both lines and repeat_seconds")
	}
}

// The Gomads pun plays the joke's animation; other "mad" words stay angry.
func TestWordAnimationsPlayOnWords(t *testing.T) {
	acts := gopodAnimationsFromWords([]RobotAction{
		{Action: ActionSayText, Parameter: "Don't make the Gohm-adds go mad, man!"},
		{Action: ActionSayText, Parameter: "I'm so mad right now."},
	})
	var anims []string
	for _, a := range acts {
		if a.Action == ActionPlayAnimationWI {
			anims = append(anims, a.Parameter)
		}
	}
	angry := "angry"
	if swap, ok := gopodBrobotsSettings.WordAnimations.Swap[angry]; ok {
		angry = swap
	}
	want := []string{gopodBrobotsSettings.WordAnimations.PlayOnWordsAnimation, angry}
	if len(anims) != 2 || anims[0] != want[0] || anims[1] != want[1] {
		t.Errorf("animations %v, want %v", anims, want)
	}
}

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_chat_hold_test.go
package wirepod_ttr

import (
	"strings"
	"testing"
)

// The hold turns on after an answer and off on "thanks Brobots"; with no robot
// connection (a test) nothing is left held.
func TestGopodChatHoldOnAndOff(t *testing.T) {
	gopodChatHoldAfterAnswer()
	gopodChatHold.Lock()
	on := gopodChatHold.active
	gopodChatHold.Unlock()
	if !on {
		t.Fatalf("hold not on after an answer")
	}
	GopodChatHoldEnd("thanks")
	gopodChatHold.Lock()
	on, held := gopodChatHold.active, len(gopodChatHold.release)
	gopodChatHold.Unlock()
	if on || held != 0 {
		t.Fatalf("after thanks: active=%v held=%d", on, held)
	}
}

func TestGopodChatHoldIdleSetting(t *testing.T) {
	if gopodBrobotsSettings.ChatHold.IdleMinutes <= 0 {
		t.Fatalf("chat_hold.idle_minutes must be set")
	}
}

func TestGopodIsThanks(t *testing.T) {
	for text, want := range map[string]bool{
		"thanks bro bots":           true,
		"Thanks Brobots!":           true,
		"thanks bro but":            true,
		"thank you robots":          true,
		"thanks for the bingo":      false,
		"i wanna party with you":    false,
		"thanksgiving robots party": false,
	} {
		if got := gopodIsThanks(text); got != want {
			t.Errorf("gopodIsThanks(%q) = %v, want %v", text, got, want)
		}
	}
}

func TestGopodFixNearCues(t *testing.T) {
	in := "dartedEyes... Did someone say CHOCOLATE BINGO?!"
	if got := gopodFixNearCues(in); got != "dartingEyes... Did someone say CHOCOLATE BINGO?!" {
		t.Errorf("got %q", got)
	}
	keep := "Chocolate... Bingo, huh?"
	if got := gopodFixNearCues(keep); got != keep {
		t.Errorf("a real word changed: %q", got)
	}
}

// Every listed misspelling reads as its own token, a cue only.
func TestGopodMisspellingsAreCues(t *testing.T) {
	for target, words := range loadedAnimationVocab.Misspellings {
		for _, w := range words {
			if !gopodIsAnimationTokenName(w) {
				t.Errorf("%q is not read as a cue", w)
			}
			if got := animationAliases[normalizeAnimationToken(w)]; got != target {
				t.Errorf("%q resolves to %q, want %q", w, got, target)
			}
		}
	}
	if got := gopodFixNearCues("kgSucess... Nailed it!"); got != "kgSuccess... Nailed it!" {
		t.Errorf("got %q", got)
	}
}

func TestGopodRunFixes1454(t *testing.T) {
	if got := gopodESNToName("Okay, so Party Bro 0dd1d8bf."); got != "Okay, so Party Bro Brobot 2." {
		t.Errorf("ESN: %q", got)
	}
	if got := gopodFixNearCues("thinking__ Okay, so"); got != "thinking"+gopodAnimationMarker+" Okay, so" {
		t.Errorf("short marker: %q", got)
	}
	if got := gopodFixNearCues("my__ word"); got != "my__ word" {
		t.Errorf("non-token changed: %q", got)
	}
}

func TestGopodTidyLines(t *testing.T) {
	in := "Well this is embarrassing!\n\nBrotobot moment: \"Hey, hey, hey!!\"\n*  ... *\n**Happy:** Celebrate.\n* 🤖 **0)**:  Hear me clearly!\nOr a thing we're supposed to do,?"
	want := "Well this is embarrassing!\n\n\"Hey, hey, hey!!\"\n* 🤖 **0)**:  Hear me clearly!\nOr a thing we're supposed to do?"
	if got := gopodTidyLines(in); got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
	keep := "I'm *so* happy you're here!"
	if got := gopodTidyLines(keep); got != keep {
		t.Errorf("a normal line changed: %q", got)
	}
}

// A cue in markdown bold is still the cue (live 2026-10-04 19:26).
func TestGopodBoldCue(t *testing.T) {
	for in, want := range map[string]string{
		"**confused**___ Say the whole thing again?":          "confused... Say the whole thing again?",
		"thinking... **dartingEyes**___ Well, so like my bro": "dartingEyes... Well, so like my bro",
		"celebrate... **love**___ Thanks for joining in":      "love... Thanks for joining in",
	} {
		if got := gopodAnimationMarkerToInternal(in); !strings.HasPrefix(got, want) && !strings.Contains(got, want) {
			t.Errorf("gopodAnimationMarkerToInternal(%q) = %q, want it to hold %q", in, got, want)
		}
	}
	if got := gopodUnboldCues("**Bold** words___ stay"); got != "**Bold** words___ stay" {
		t.Errorf("a non-cue word changed: %q", got)
	}
}

// ## EOF

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_addressing_test.go
package wirepod_ttr

import (
	"testing"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

func TestGopodAddressedTo(t *testing.T) {
	for text, want := range map[string]string{
		"brobot one tell me about yourself":         "Brobot 1",
		"Brobot 1, tell me about yourself":          "Brobot 1",
		"bro bought one tell me about yourself":     "Brobot 1",
		"hey bro bot two what's up":                 "Brobot 2",
		"bro bought to tell me a joke":              "Brobot 2",
		"robot one tell me about yourself":          "Brobot 1",
		"robot two what is gopod":                   "Brobot 2",
		"tell me about your brother bro bought one": "", // a topic, not a call
		"did you get what bro bought one meant":     "", // a topic, not a call
		"tell me about yourself":                    "", // no name
		"brobots one more time":                     "", // whole words only
		"robots one by one":                         "", // whole words only
	} {
		if got := gopodAddressedTo(text); got != want {
			t.Errorf("gopodAddressedTo(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestGopodCanonicalCall(t *testing.T) {
	for text, want := range map[string]string{
		"bro bought two can you hear me":            "Brobot 2 can you hear me",
		"hey bro bot one, please answer":            "hey Brobot 1, please answer",
		"Robot 1 do you hear me":                    "Brobot 1 do you hear me",
		"tell me about your brother bro bought one": "tell me about your brother bro bought one",
	} {
		if got := gopodCanonicalCall(text); got != want {
			t.Errorf("gopodCanonicalCall(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestGopodDanglingCueLineDropped(t *testing.T) {
	in := "Gosh. I can hear ya!\n\nbrobot 1" + gopodAnimationMarker + "  **"
	if got := gopodPrepareReply(in); got != "Gosh. I can hear ya!\n" {
		t.Errorf("gopodPrepareReply(%q) = %q", in, got)
	}
	keep := "happy" + gopodAnimationMarker + " Yo!"
	if got := gopodPrepareReply(keep); got != keep {
		t.Errorf("a one-line cued reply changed: %q", got)
	}
}

func TestGopodCutReplyTail(t *testing.T) {
	in := "...you got it bro, we totally nailed that! 😎\n\nhappy \n\n### Explanation:\n\n1. **Emotional signal:** The user asks."
	if got := gopodCutReplyTail(in); got != "...you got it bro, we totally nailed that! 😎" {
		t.Errorf("got %q", got)
	}
	in = "Hey bro how's it going? 😉\nsad\n**Explanation of Robot's Thought Process:**\n1. x"
	if got := gopodCutReplyTail(in); got != "Hey bro how's it going? 😉" {
		t.Errorf("got %q", got)
	}
	keep := "happy" + gopodAnimationMarker + " An explanation is what you get, bro!"
	if got := gopodCutReplyTail(keep); got != keep {
		t.Errorf("a reply using the word changed: %q", got)
	}
}

func resetFirstHeard() {
	gopodFirstHeard.Lock()
	gopodFirstHeard.esn, gopodFirstHeard.words = "", nil
	gopodFirstHeard.Unlock()
}

// A named question: only the named Brobot answers, whichever is heard first.
func TestGopodShouldAnswerNamed(t *testing.T) {
	resetFirstHeard()
	now := time.Now()
	if answer, to := gopodShouldAnswer(gopodBrobot2ESN, "brobot one tell me about yourself", now); answer || to != "Brobot 1" {
		t.Errorf("Brobot 2 on a Brobot 1 question: answer=%v to=%q, want quiet", answer, to)
	}
	if answer, _ := gopodShouldAnswer(gopodBrobot1ESN, "bro bought one tell me about yourself", now); !answer {
		t.Errorf("Brobot 1 on his own name: quiet, want answer")
	}
}

// No name: the first Brobot heard answers, his brother hearing the same
// question stays quiet; a different question, or one after the window, is
// answered.
func TestGopodShouldAnswerNoName(t *testing.T) {
	resetFirstHeard()
	now := time.Now()
	if answer, _ := gopodShouldAnswer(gopodBrobot1ESN, "tell me about yourself", now); !answer {
		t.Fatalf("first heard: quiet, want answer")
	}
	if answer, _ := gopodShouldAnswer(gopodBrobot2ESN, "tell me about you're self", now.Add(time.Second)); answer {
		t.Errorf("brother on the same question: answer, want quiet")
	}
	if answer, _ := gopodShouldAnswer(gopodBrobot2ESN, "what is the weather like today", now.Add(2*time.Second)); !answer {
		t.Errorf("brother on a different question: quiet, want answer")
	}
	resetFirstHeard()
	gopodShouldAnswer(gopodBrobot1ESN, "tell me about yourself", now)
	window := gopodSecondsToDuration(gopodBrobotsSettings.Addressing.FirstHeardSeconds)
	if answer, _ := gopodShouldAnswer(gopodBrobot2ESN, "tell me about yourself", now.Add(window+time.Second)); !answer {
		t.Errorf("same question after the window: quiet, want answer")
	}
	// The same Brobot asked again is never quiet.
	if answer, _ := gopodShouldAnswer(gopodBrobot2ESN, "tell me about yourself", now.Add(window+2*time.Second)); !answer {
		t.Errorf("same Brobot again: quiet, want answer")
	}
}

func TestGopodQuietAnimationsAreTokens(t *testing.T) {
	a := gopodBrobotsSettings.Addressing
	for _, token := range []string{a.QuietAnimation} {
		if gopodAnimationClip(token) == "" {
			t.Errorf("%q is not an animation_vocab.json token", token)
		}
	}
}

// ## EOF

func TestGopodBrotherVoiceCommandTakesIt(t *testing.T) {
	resetFirstHeard()
	old := logger.LogTrayArray
	defer func() { logger.LogTrayArray = old }()
	now := time.Now()
	logger.LogTrayArray = []string{now.Format("2006.01.02 15:04:05") + ": Bot " + gopodBrobot2ESN + " Executing: /x/start_bingo_test_live.sh\n"}
	if answer, _ := gopodShouldAnswer(gopodBrobot1ESN, "the goat test live", now); answer {
		t.Errorf("brother ran the voice command; this one should stay quiet")
	}
	logger.LogTrayArray = []string{now.Add(-time.Minute).Format("2006.01.02 15:04:05") + ": Bot " + gopodBrobot2ESN + " Executing: /x/y.sh\n"}
	if answer, _ := gopodShouldAnswer(gopodBrobot1ESN, "what is gopod", now); !answer {
		t.Errorf("an old command must not silence a new question")
	}
}

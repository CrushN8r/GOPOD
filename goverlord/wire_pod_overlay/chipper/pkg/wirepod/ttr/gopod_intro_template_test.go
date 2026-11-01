// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_intro_template_test.go
package wirepod_ttr

import (
	"strings"
	"testing"
)

// 2026-10-31 (operator: "FIX IT CORRECT & PROPER"): every intro is a template -
// the code's opening, the model's middle, the code's cast line, once.
func TestIntroTemplate(t *testing.T) {
	for _, who := range []string{"Brobot 1", "Brobot 2", "BUTLER", "GIT REPO", "CHALK", "PLAYHEAD", "CAMEO"} {
		opening := gopodIntroOpening(who)
		cast := gopodIntroCastLine(who)
		if !strings.HasPrefix(opening, "I'm ") || cast == "" {
			t.Errorf("%s: opening %q, cast line %q", who, opening, cast)
			continue
		}
		names := strings.Split(gopodIntroCast(who), ", ")
		// The model said the opening anyway and listed the cast in a sentence.
		middle := opening + " I keep things running. Helping me out are " + strings.Join(names[:3], ", ") + ". That's my job!"
		got := gopodIntroWrap(who, middle)
		if !strings.HasPrefix(got, opening+" I keep things running.") || !strings.HasSuffix(got, cast) {
			t.Errorf("%s: %q", who, got)
		}
		if strings.Count(got, opening) != 1 || strings.Contains(got, "Helping me out") || !strings.Contains(got, "That's my job!") {
			t.Errorf("%s: opening or cast twice, or his words lost: %q", who, got)
		}
	}
	// An intro line: no value point, its own type, the brobots_intro chain wraps it.
	if gopodQuestionFactsFor("0dd1b9e9", "Hey Doc! You party animal! Introduce yourself fully to the chat") != "" {
		t.Error("an intro got value point facts")
	}
	if typ := gopodAnswerType("0dd1b9e9", "Hey Doc! Introduce yourself fully"); typ != "intro" {
		t.Errorf("intro answer type: %q", typ)
	}
	if p := gopodAnswerChain("brobots_intro"); len(p) != 2 || p[1] != "intro_wrap" {
		t.Errorf("brobots_intro chain: %v", p)
	}
}

// 2026-10-31 (operator: "Should the reminders name eggs? * no, but GOPOD easter
// eggs should still work"): a free answer's reminder names topics, never eggs.
func TestRemindersNameNoEggs(t *testing.T) {
	for i := 0; i < 6; i++ {
		line := gopodReminderLine("tpu")
		if line == "" {
			t.Fatal("a free answer has no reminder")
		}
		for _, e := range gopodBrobotsSettings.ValuePoints.ReminderEggs {
			if strings.Contains(line, e.Label) {
				t.Errorf("reminder names the egg %q: %q", e.Label, line)
			}
		}
	}
}

// 2026-10-31: the crew roster - everyone in the chat has his line; In the chat reads it.
func TestCrewRoster(t *testing.T) {
	roster := gopodCrewRoster()
	for _, c := range gopodBrobotsSettings.HereNow.Cast {
		if c.Who != gopodBrobotsSettings.BrainChain.OperatorWho && roster[strings.ToLower(gopodCallName(c.Who))] == "" {
			t.Errorf("%s isn't on the crew roster", gopodCallName(c.Who))
		}
	}
	l := gopodBrainLines("Brobot 1", "hey")
	if len(l) < 2 || !strings.Contains(l[1], "Keys [Butler - Goverlord's voice at GOPOD]") || !strings.Contains(l[1], "Doc [GOPOD host - a party-bro GOPOD rep] (me)") {
		t.Errorf("In the chat: %v", l)
	}
}

// 2026-11-01 (operator: "update weird [chunking to sentences to commas] splits *
// try to keep sentences generated whole, then play out the pauses"): a chunk is
// whole sentences - never cut at a comma.
func TestChunksAreWholeSentences(t *testing.T) {
	for _, line := range []string{
		"Correct. The system works. Two hosts, one mission: get you people hyped before two Brobots who used to be considered trash, have an actual conversation.",
		"Okay. Rescued. The point is, something is being prepared right now, live, as we speak. And we are going to be here, to make it feel like an epic event. Because it is, truly an event of a lifetime.",
		"I'm Git Repo. I take a moment to look around the room. My job is to help keep this place running smoothly. My manager, PLAYHEAD ZUMINTU, and my study buddy, Cache PYC, are helping me make sure everything stays on track.",
	} {
		chunks := gopodSpokenChunks(line)
		if strings.Join(chunks, " ") != line {
			t.Errorf("chunks lose words: %q", chunks)
		}
		for _, c := range chunks {
			if end := c[len(c)-1:]; !strings.ContainsAny(end, ".!?") {
				t.Errorf("chunk ends mid-sentence (%q): %q", end, c)
			}
		}
	}
}

// 2026-11-01 (live: the comm check stopped while Git Repo's answer was being
// made; it came 2 minutes later and played): a stop run while a Gomad's answer
// is made drops it - nothing is said.
func TestStopDropsGomadAnswerBeingMade(t *testing.T) {
	ask := gopodPersonaAsk
	defer func() { gopodPersonaAsk = ask }()
	gopodPersonaAsk = func(host, text string) (string, error) {
		gopodStopSeq.Add(1) // a stop run while it's made
		return "I'm Git Repo. Late answer.", nil
	}
	said := false
	dry := gopodDryVoice
	defer func() { gopodDryVoice = dry }()
	gopodDryVoice = func(host, text string, show bool) { said = true }
	if !gopodGomadAnswerAs("GIT REPO", "GIT REPO", "tell me a joke") {
		t.Fatal("the llm part didn't take the line")
	}
	if said {
		t.Error("a dropped answer was said")
	}
}

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_intro_template_test.go
package wirepod_ttr

import (
	"context"
	"strings"
	"testing"
)

// 2026-10-31 (operator: "FIX IT CORRECT & PROPER"): every intro is a template -
// the code's opening, the model's middle, the code's cast line, once.
func TestIntroTemplate(t *testing.T) {
	for _, who := range []string{"Brobot 1", "Brobot 2", "BUTLER", "GIT REPO", "CHALK", "PLAYHEAD", "CAMEO"} {
		opening := gopodIntroOpening(who)
		cast := gopodIntroCastLine(who)
		// 2026-11-04 (operator: "remove who's here in the chat from every response"): no cast line.
		if !strings.HasPrefix(opening, "I'm ") || cast != "" {
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
	if gopodQuestionFactsFor("ESN01", "Hey Doc! You party animal! Introduce yourself fully to the chat") != "" {
		t.Error("an intro got value point facts")
	}
	if typ := gopodAnswerType("ESN01", "Hey Doc! Introduce yourself fully"); typ != "intro" {
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
		if c.Who != gopodBrobotsSettings.BrainChain.OperatorWho && roster[strings.ToLower(gopodCallName(c.Who))] == "" && roster[strings.ToLower(c.Who)] == "" {
			t.Errorf("%s isn't on the crew roster", gopodCallName(c.Who))
		}
	}
	l := gopodBrainLines("Brobot 1", "hey")
	if len(l) < 2 || !strings.Contains(l[1], "GOPOD System: Keys [Butler, Goverlord's voice]") || !strings.Contains(l[1], "Brobot 1 [grumpy, impatient] (me)") {
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

// 2026-11-03 (operator: "there is too much time between exchanges"; his OK: the
// answer made while the question is read out): a Gomad's answer started early
// is asked for once, before the read-out ends, and that answer is the one said.
func TestGomadAnswerMadeEarly(t *testing.T) {
	ask := gopodPersonaAsk
	defer func() { gopodPersonaAsk = ask }()
	asks := 0
	gopodPersonaAsk = func(host, text string) (string, error) {
		asks++
		return "I'm Git Repo. Made early.", nil
	}
	var said []string
	dry := gopodDryVoice
	defer func() { gopodDryVoice = dry }()
	gopodDryVoice = func(host, text string, show bool) { said = append(said, text) }
	early := gopodGomadAnswerStart("GIT REPO", "GIT REPO", "tell me a joke")
	if early == nil {
		t.Fatal("the llm part's answer wasn't started early")
	}
	<-early.done // made while the line is read out
	if asks != 1 {
		t.Fatalf("asked %d times before the read-out ended, want 1", asks)
	}
	if !gopodGomadAnswerEarly("GIT REPO", "GIT REPO", "tell me a joke", early) {
		t.Fatal("the llm part didn't take the line")
	}
	if asks != 1 {
		t.Errorf("asked again after the read-out (%d asks)", asks)
	}
	if !strings.Contains(strings.Join(said, " "), "Made early.") {
		t.Errorf("the early answer wasn't said: %q", said)
	}
}

// 2026-11-02 (operator: "Add in the score in every brain bubble"): the score
// rows - topics, value points, eggs, each label then its bar.
func TestBrainScoreRows(t *testing.T) {
	rows := gopodBrainScoreRows()
	if len(rows) != 3 || !strings.HasPrefix(rows[0], "Topics ") || !strings.HasPrefix(rows[2], "Easter Eggs ") {
		t.Fatalf("score rows: %q", rows)
	}
	for _, r := range rows {
		if !strings.Contains(r, "\t") || strings.Contains(r, "{") {
			t.Errorf("score row not filled or not a row: %q", r)
		}
	}
}

// 2026-11-02 (live: Pip's KP2 intro said only "I'm Brobot 2, my nickname is
// Pip."; Doc's and Pip's All intros skipped their first thought): every part of a
// Brobot's templated intro is spoken - the opening, each of his thoughts, the cast.
func TestBrobotIntroAllSpoken(t *testing.T) {
	m := gopodAnimationMarker
	for _, middle := range []string{
		"thinking" + m + " Let me tell you a little something about myself! You know the party's on.",
		"Okay, so long story short, I'm the party guy! 😜\nhappy" + m + " Our gig? We help folks get it.",
	} {
		wrapped := gopodIntroWrap("Brobot 2", middle, true)
		// 2026-11-04 (operator: "remove who's here in the chat from every response"): no cast line.
		if strings.Contains(wrapped, "Here in the chat with us") {
			t.Errorf("still lists who's here: %q", wrapped)
		}
		r := buildGOPODStreamResultWithRepair(context.Background(), gopodAnimationMarkerToInternal(wrapped), "test", nil)
		said := strings.Join(gopodSpokenPieceTexts(r.AudibleChunks), " ")
		for _, want := range []string{"my nickname is Pip"} {
			if !strings.Contains(said, want) {
				t.Errorf("not said %q:\n%s\n(wrapped %q)", want, said, wrapped)
			}
		}
		for _, w := range strings.Fields(strings.NewReplacer("thinking"+m, "", "happy"+m, "", "😜", "").Replace(middle)) {
			if !strings.Contains(said, strings.Trim(w, ".,!?")) {
				t.Errorf("his word %q not said:\n%s", w, said)
				break
			}
		}
	}
}

// 2026-11-02 (live): the opening is never said twice, however the model
// repeats it, and the model's own "here in the chat" sentence goes.
func TestIntroOpeningOnce(t *testing.T) {
	m := gopodAnimationMarker
	cases := []struct{ who, middle, gone, kept string }{
		{"CHALK", "I'm CHALK BORED, my nickname is CHALK – because it's my name. I run the display.", "my nickname is CHALK –", "I run the display."},
		{"CACHE PYC", "I'm Cache PYC. I'm Cache PYC, and I'm hustling every day.", "I'm Cache PYC, and", "I'm hustling every day."},
		{"Brobot 2", "thinking" + m + " I'm Brobot 2, and hey! You know the party's on.", "I'm Brobot 2, and", "Hey!"},
		{"Brobot 1", "I'm Doc, Brobot 1, which means I get to be the party guy. Here in the chat: all our brothers are rocking it, including Pip.", "all our brothers", "Which means I get to be the party guy."},
	}
	for _, c := range cases {
		got := gopodIntroWrap(c.who, c.middle, c.who == "Brobot 1" || c.who == "Brobot 2")
		if strings.Count(gopodLooseWords(got), gopodLooseWords(gopodIntroOpening(c.who))) != 1 || strings.Contains(got, c.gone) || !strings.Contains(got, c.kept) {
			t.Errorf("%s: %q", c.who, got)
		}
	}
}

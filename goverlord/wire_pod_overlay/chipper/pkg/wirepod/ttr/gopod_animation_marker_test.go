package wirepod_ttr

import (
	"strings"
	"testing"
)

// The "___" animation marker: converted to Wire-Pod's internal "happy... " form
// for the reply pipeline, and removed from the chat display. A natural "Ah..."
// stays in the bubble; in speech its "..." is dropped, so it ends no sentence and
// is never taken for a cue ("Wow! Happy... birthday").
func TestGopodAnimationMarker(t *testing.T) {
	raw := "happy___ Ah... so, sorta. We're partners in crime. thinking___ Next bit here, bro."
	internal := gopodAnimationMarkerToInternal(raw)
	if want := "happy... Ah, so, sorta. We're partners in crime. thinking... Next bit here, bro."; internal != want {
		t.Fatalf("internal = %q, want %q", internal, want)
	}
	units := parseBrobotPhraseUnits(internal)
	if len(units) != 2 || units[0].ActionParameter != "happy" || units[1].ActionParameter != "thinking" {
		t.Fatalf("units = %+v, want happy then thinking", units)
	}
	if units[0].Text != "Ah, so, sorta. We're partners in crime." {
		t.Fatalf("first unit text = %q", units[0].Text)
	}
	tests := map[string]string{
		raw:                               "Ah... so, sorta. We're partners in crime. Next bit here, bro.",
		"Bro___ ok then":                  "Bro... ok then",
		"veryHappy ___  Man oh man!":      "Man oh man!",
		"happy. hey there! old habit":     "hey there! old habit",
		"Ah... so, sorta. No cue at all.": "Ah... so, sorta. No cue at all.",
		// The prompt's placeholder word copied in front of the cue (live 2026-10-01).
		"actionParameter_thinking___ Honestly, I was built for it.": "Honestly, I was built for it.",
		"actionParameter___ happy___ Dude, it was a blast!":         "Dude, it was a blast!",
	}
	for in, want := range tests {
		if got := gopodDisplayReply(in); got != want {
			t.Fatalf("gopodDisplayReply(%q) = %q, want %q", in, got, want)
		}
	}
	if got := gopodAnimationMarkerToInternal("happy___ Wow! Happy… birthday, bro..."); got != "happy... Wow! Happy, birthday, bro" {
		t.Fatalf("spoken ellipsis kept: %q", got)
	}
	if units := parseBrobotPhraseUnits(gopodAnimationMarkerToInternal("happy___ Wow! Happy... birthday, bro.")); len(units) != 1 {
		t.Fatalf("speech \"Happy...\" taken for a cue: %+v", units)
	}
	if got := gopodAnimationMarkerToInternal("actionParameter_thinking___ Honestly, bro."); got != "thinking... Honestly, bro." {
		t.Fatalf("placeholder word kept: %q", got)
	}
	if got := gopodInternalToAnimationMarker("happy... Ah... so, sorta. thinking... Next bit."); got != "happy___ Ah... so, sorta. thinking___ Next bit." {
		t.Fatalf("memory form = %q", got)
	}
	if got := gopodInternalToAnimationMarker("Ah... no cue here."); got != "thinking___ Ah... no cue here." {
		t.Fatalf("no-cue memory = %q", got)
	}
	// A bare leading lowercase cue word, no marker (live 2026-10-01 11:5x).
	if got := gopodAnimationMarkerToInternal("happy circuits buzzing! Hey there."); got != "happy... circuits buzzing! Hey there." {
		t.Fatalf("bare cue = %q", got)
	}
	if got := gopodDisplayReply("happy circuits buzzing! Hey there."); got != "circuits buzzing! Hey there." {
		t.Fatalf("bare cue bubble = %q", got)
	}
	if got := gopodAnimationMarkerToInternal("Happy to help, bro."); got != "Happy to help, bro." {
		t.Fatalf("capital Happy taken for a cue: %q", got)
	}
	gopodStageDirectionEmojiNext = 0
	if got := gopodDisplayReply("happy___ *circuits buzzing!* Hey dude! My *brain* switch. *thinking* *sarcastic* Wow."); got != "📣 circuits buzzing!\nHey dude! My *brain* switch.\n🎬 thinking\n📣 sarcastic\nWow." {
		t.Fatalf("stage directions = %q", got)
	}
	// The prompt's example text copied onto its own line (live 2026-10-01 15:25).
	if got := gopodAnimationMarkerToInternal("spoken thought  thinking \npartyBro I literally almost did."); got != "thinking... partyBro I literally almost did." {
		t.Fatalf("example-text line = %q", got)
	}
	if got := gopodDisplayReply("spoken thought  thinking \npartyBro I literally almost did."); got != "partyBro I literally almost did." {
		t.Fatalf("example-text bubble = %q", got)
	}
	if got := gopodAnimationMarkerToInternal("happy___ That spoken thought was mine."); got != "happy... That spoken thought was mine." {
		t.Fatalf("example words mid-line dropped: %q", got)
	}
	// A bracketed stage direction and a "---" line (live 2026-10-01 16:09).
	gopodStageDirectionEmojiNext = 0
	brackets := "thinking___ Let's play!\n\n[BROBOT turns heads, smiling]\n\n---   \n and if you don't get it, tell me.\n"
	if got := gopodDisplayReply(brackets); got != "Let's play!\n📣 BROBOT turns heads, smiling" {
		t.Fatalf("bracket/rule bubble = %q", got)
	}
	if r := buildGOPODStreamResult(gopodAnimationMarkerToInternal(brackets)); len(r.AudibleChunks) != 2 || !strings.Contains(r.AudibleChunks[1], "BROBOT turns heads, smiling") {
		t.Fatalf("bracket speech = %q", r.AudibleChunks)
	}
	if got := gopodAnimationMarkerToInternal("[happy] Hi."); got != "[happy] Hi." {
		t.Fatalf("one-word bracket cue changed: %q", got)
	}
	// "..." is said as a comma after a word, nothing elsewhere (operator 17:1x).
	for in, want := range map[string]string{
		"put something in... and it got all messed up!": "put something in, and it got all messed up!",
		"Wow!... ok then.":                                "Wow! ok then.",
		"... Hello.":                                      "Hello.",
		"So close...":                                     "So close.",
	} {
		if got := gopodEndWithPunctuation(cleanRobotSpeechText(in)); got != want {
			t.Fatalf("%q said as %q, want %q", in, got, want)
		}
	}
	// A reply that opens on a new line, no marker (live 2026-10-01 15:54).
	if got := gopodDisplayReply("\nI'm Brobot 1 - always ready."); got != "I'm Brobot 1 - always ready." {
		t.Fatalf("leading newline bubble = %q", got)
	}
	if !gopodRobotSpeechLinePattern.MatchString("happy___ Hello there, bro.") {
		t.Fatal("validator rejects the ___ marker")
	}
}

// A stray single quote is not read out as "single quote" (live 2026-10-02 22:12);
// apostrophes and elisions stay.
func TestGopodDropStrayQuotes(t *testing.T) {
	for in, want := range map[string]string{
		"less 'in-the-'bizzy business.'.":          "less 'in-the-bizzy business..",
		"blow ya' minds, 'cause we don't stop it.": "blow ya' minds, 'cause we don't stop it.",
	} {
		if got := gopodDropStrayQuotes(in); got != want {
			t.Fatalf("gopodDropStrayQuotes(%q) = %q, want %q", in, got, want)
		}
	}
}

// A "***" line is a markdown rule too: the bubble ends there (live 2026-10-03).
func TestGopodRuleLineStars(t *testing.T) {
	if got := gopodDisplayReply("happy___ Let's do this thing! 🚀\n\n***\n"); got != "Let's do this thing! 🚀" {
		t.Fatalf("bubble = %q", got)
	}
}

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_speak_links_test.go
package wirepod_ttr

import (
	"strings"
	"testing"
)

// TestLinkWords: a link is said as its words (live 2026-10-16: the weather kit's
// address was read out).
func TestLinkWords(t *testing.T) {
	for in, want := range map[string]string{
		"Check out the [GOPOD Weather Kit](/gopod_weather_kit.html).": "Check out the GOPOD Weather Kit.",
		"Check out the *GOPOD Weather Kit*(/gopod_weather_kit.html).": "Check out the GOPOD Weather Kit.",
		"Check out the GOPOD Weather Kit (/gopod_weather_kit.html).":  "Check out the GOPOD Weather Kit.",
		"Visit [crushn8r.ca](https://crushn8r.ca) today":              "Visit crushn8r.ca today",
		"Pip (the wrong one) is here":                                 "Pip (the wrong one) is here",
	} {
		if got := gopodLinkWords(in); got != want {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
}

// ## EOF

// TestLinkKeepsBrackets (live 2026-10-16): a link's brackets aren't read as a
// stage direction; a two-word bracket still is.
func TestLinkKeepsBrackets(t *testing.T) {
	if got := gopodBracketStageDirections("Check out the [GOPOD Weather Kit](/gopod_weather_kit.html). [waves both arms]"); got != "Check out the [GOPOD Weather Kit](/gopod_weather_kit.html). *waves both arms*" {
		t.Errorf("got %q", got)
	}
}

// TestThoughtParagraphs (2026-10-16, operator: "thought bubbles as (spaced)
// paragraphs"): a new thought's line break reaches the marked bubble, so the
// page can start a paragraph there; the pieces are the same.
func TestThoughtParagraphs(t *testing.T) {
	marker := gopodBrobotsSettings.ScriptedSplit.SegmentMarker
	if marker == "" {
		t.Skip("no segment marker")
	}
	got, _ := gopodSegmentDisplay("Let's get cheesy, bro! 😉\nYou wanna go first?", []string{"Let's get cheesy,", "bro!", "You wanna go first?"})
	if want := "Let's get cheesy, " + marker + "bro! 😉\n " + marker + "You wanna go first?"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestKitLinkSaidAndShown (live 2026-10-16 13:2x: "Check out the." was said):
// the bubble keeps the Weather Kit link, the robot says its words.
func TestKitLinkSaidAndShown(t *testing.T) {
	raw := "happy___ Granny is hot!\nWant to see every weather face Vector has? Check out the [GOPOD Weather Kit](/gopod_weather_kit.html)."
	display := gopodDisplayReply(raw)
	if !strings.Contains(display, "[GOPOD Weather Kit](/gopod_weather_kit.html)") {
		t.Errorf("the bubble lost the link: %q", display)
	}
	internal := gopodAnimationMarkerToInternal(gopodLinkWords(raw))
	said := strings.Join(gopodSpokenPieceTexts(buildGOPODStreamResult(internal).AudibleChunks), " ")
	if !strings.Contains(said, "Weather Kit") || strings.Contains(said, "html") {
		t.Errorf("said %q", said)
	}
}

// TestMusicShownNotSaid (2026-10-16, operator: "(music) ... show it in the bubble,
// to indicate to deaf humans"): the butler doesn't say it.
func TestMusicShownNotSaid(t *testing.T) {
	said := gopodSaySwap(gopodBrobotsSettings.LLM.ReadyText, gopodBrobotsSettings.ReporterVoices.SystemSayWords)
	if strings.Contains(said, "music") || !strings.Contains(gopodBrobotsSettings.LLM.ReadyText, "(music)") {
		t.Errorf("shown %q, said %q", gopodBrobotsSettings.LLM.ReadyText, said)
	}
}

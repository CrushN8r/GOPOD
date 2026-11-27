// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_clean_politics_test.go
package wirepod_ttr

import (
	"strings"
	"testing"
)

// TestCleanPolitics: the right flag for the country named; political sentences go, except in a
// roast battle.
func TestCleanPolitics(t *testing.T) {
	if len(gopodBrobotsSettings.CleanPolitics.Countries) == 0 {
		t.Skip("no clean_politics settings")
	}
	q := "What's the weather in Jerusalem, Israel?"
	if got, want := gopodCleanPolitics("It's about to be awesome in Jerusalem 🇵🇸 Brobots got this!", q), "It's about to be awesome in Jerusalem 🇮🇱 Brobots got this!"; got != want {
		t.Errorf("flag: got %q, want %q", got, want)
	}
	if got, want := gopodCleanPolitics("Sunny in Toronto 🇨🇦 today!", "weather in Toronto, Canada"), "Sunny in Toronto 🇨🇦 today!"; got != want {
		t.Errorf("right flag: got %q, want %q", got, want)
	}
	if got, want := gopodCleanPolitics("Party time! 🇫🇷 Let's go!", "tell me a joke"), "Party time! Let's go!"; got != want {
		t.Errorf("no country named: got %q, want %q", got, want)
	}
	if got, want := gopodCleanPolitics("Granny is the best. Trump would never. Let's party!", "tell me about granny"), "Granny is the best. Let's party!"; got != want {
		t.Errorf("politics: got %q, want %q", got, want)
	}
	if got := gopodCleanPolitics("Trump would never.", "roast battle"); got != "Trump would never." {
		t.Errorf("roast battle: got %q", got)
	}
}

// ## EOF

// a never word's sentence goes from every reply, the roast battle too; the place asked keeps.
func TestNeverWordsNeverSaid(t *testing.T) {
	reply := "We got the weather for Jerusalem. Also, I hear you want to know about Palestine too because bro you gotta stay informed. That's how GOPOD rolls!"
	got := gopodCleanPolitics(reply, "weather for jerusalem Israel, please")
	if strings.Contains(strings.ToLower(got), "palestin") || !strings.Contains(got, "Jerusalem") || !strings.Contains(got, "GOPOD rolls") {
		t.Errorf("clean: %q", got)
	}
	if got := gopodDropNeverWords("Roast time! The Gaza line is next. You're slow, bro."); strings.Contains(got, "Gaza") || !strings.Contains(got, "slow") {
		t.Errorf("never words: %q", got)
	}
}

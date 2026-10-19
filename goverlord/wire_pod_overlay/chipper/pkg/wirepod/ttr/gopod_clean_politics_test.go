// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_clean_politics_test.go
package wirepod_ttr

import "testing"

// TestCleanPolitics (live 2026-10-16: 🇵🇸 in a "Jerusalem, Israel" reply): the
// right flag for the country named; political sentences go, except in a roast
// battle.
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

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_eggs_test.go
package wirepod_ttr

import "testing"

// The pinnacle egg process: eggs come in the order their words come in the line,
// each congratulated by its own word; an egg's word is never a weather place.
func TestEggsInOrderMentioned(t *testing.T) {
	line := "is it granny weather for bingo? explain the math"
	granny, _ := gopodEggWord(line, []string{"legacy", "granny"})
	weather, _ := gopodEggWord(line, []string{"weather"})
	bingo, word := gopodEggWord(line, []string{"bingo"})
	math, _ := gopodEggWord(line, []string{"explain the math"})
	if !(granny < weather && weather < bingo && bingo < math) || word != "bingo" {
		t.Errorf("order granny %d weather %d bingo %d math %d (%q)", granny, weather, bingo, math, word)
	}
	if at, _ := gopodEggWord("bingos are fun", []string{"bingo"}); at != -1 {
		t.Errorf("part of a word found an egg at %d", at)
	}
	// 2026-10-25: the new template lists the eggs revealed so far (TestEggsRevealedThusFar).
	if got := gopodWeatherPlace("is it granny weather for bingo?"); got != "" {
		t.Errorf("weather place %q, want none", got)
	}
}

// TestEggZooms (2026-10-16): "MouseOver to Zoom View" only under a prize that zooms.
func TestEggZooms(t *testing.T) {
	if !gopodEggZooms("/pictures/gangster_granny.jpg") || !gopodEggZooms(gopodWeatherPopupMark+"clear sky@20") {
		t.Error("a picture or a weather face should zoom")
	}
	if gopodEggZooms(gopodEggCardMark+"Roast Battle") || gopodEggZooms(gopodMathPopupMark+"1/3") {
		t.Error("a card or the math trophy alone shouldn't")
	}
}

// TestChocolateIsFoodBingoIsGame (2026-10-16, operator: "chocolate = food, bingo =
// game"): "chocolate bingo" loads the food topic only; "bingo" the game.
func TestChocolateIsFoodBingoIsGame(t *testing.T) {
	names := func(line string) map[string]bool {
		out := map[string]bool{}
		for _, p := range loadScenarioPacketsForMessage(line) {
			out[p.Name] = true
		}
		return out
	}
	if got := names("tell me about chocolate bingo"); got["bingo"] || !got["foodmath_recipes"] {
		t.Errorf("chocolate bingo loaded %v", got)
	}
	if got := names("tell me about bingo hosting"); !got["bingo"] {
		t.Errorf("bingo loaded %v", got)
	}
}

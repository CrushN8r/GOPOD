// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_eye_colors_test.go
package wirepod_ttr

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestEyeColorHex: the same colour Python's colorsys gave the interview runner.
func TestEyeColorHex(t *testing.T) {
	for _, c := range []struct {
		h, s float64
		want string
	}{{0, 1, "#FF0000"}, {2.0 / 3, 1, "#0000FF"}, {0.5, 0.5, "#80FFFF"}, {0, 0, "#FFFFFF"}} {
		if got := gopodHSVHex(c.h, c.s, 1); got != c.want {
			t.Errorf("hsv(%v, %v) = %s, want %s", c.h, c.s, got, c.want)
		}
	}
}

// ## EOF

// 2026-11-15: a card's border is the robot's eye colour - hex to hue and
// saturation and back; DOC's red while he answers, Brobot 1's gold after.
func TestBrobotEyesFromCards(t *testing.T) {
	for _, hex := range []string{"#FF3B30", "#42A5F5", "#C9A227"} {
		h, s, ok := gopodHexHS(hex)
		if !ok || gopodHSVHex(h, s, 1) == "" {
			t.Errorf("%s: %v %v %v", hex, h, s, ok)
		}
	}
	if h, s, _ := gopodHexHS("#0000FF"); gopodHSVHex(h, s, 1) != "#0000FF" {
		t.Errorf("blue round trip: %v %v", h, s)
	}
	// Decoupled: each persona its own colours - Brobot 1's brass, DOC's red.
	if c := gopodCardColour(gopodBrobotsSettings.HostPersonas["Brobot 1"].Colours, "border"); c != "#c9a227" {
		t.Errorf("Brobot 1: %q", c)
	}
	if c := gopodCardColour(gopodBrobotsSettings.HostPersonas["DOC"].Colours, "border"); c != "#ff3b30" {
		t.Errorf("DOC: %q", c)
	}
	if !gopodAtDialZero("Brobot 1") && gopodBrobotEyeHex("Brobot 1", true) != "#ff3b30" {
		t.Errorf("DOC answering: %s", gopodBrobotEyeHex("Brobot 1", true))
	}
	if gopodBrobotEyeHex("Brobot 1", false) != "#c9a227" {
		t.Errorf("Brobot 1 at rest: %s", gopodBrobotEyeHex("Brobot 1", false))
	}
}

// 2026-11-15 (live: the page showed no persona colours): the page's settings
// carry each card's colours - KEYNOTE's, a Brobot's own and his hosted persona's.
func TestPageGetsCardColours(t *testing.T) {
	var out struct {
		HostPersonas map[string]map[string]any `json:"host_personas"`
	}
	full, _ := gopodSettingsFull()
	if err := json.Unmarshal(gopodPersonaProfileLabelsJSON(full), &out); err != nil {
		t.Fatal(err)
	}
	if c, _ := out.HostPersonas["BUTLER"]["colours"].(string); !strings.Contains(c, "#2E0A57") {
		t.Errorf("KEYNOTE's colours: %q", c)
	}
	b, _ := out.HostPersonas["Brobot 1"]["colours"].(string)
	h, _ := out.HostPersonas["DOC"]["colours"].(string)
	if !strings.Contains(b, "#c9a227") || !strings.Contains(h, "#ff3b30") {
		t.Errorf("Brobot 1's colours: own %q, hosted %q", b, h)
	}
}

// 2026-11-15 (live: Keys said CRUSHN8R like everyone): Keys' own way wins, whatever the case.
func TestKeysSaysCrushN8rHisWay(t *testing.T) {
	if got := gopodSaysFor("BUTLER", "CRUSHN8R Crew"); !strings.Contains(got, "crushin-ate-uh") {
		t.Errorf("Keys: %q", got)
	}
	if got := gopodSaysFor("CHALK", "CRUSHN8R Crew"); !strings.Contains(got, "crushin-ater") {
		t.Errorf("CHALK: %q", got)
	}
}

// 2026-11-15 (operator: "decouple brobots from DOC/PIP completely * They only
// share the same robot bodies"): four personas, four keys; a body is the
// persona answering from it only through gopodAnswerer.
func TestBrobotsDecoupledFromDocAndPip(t *testing.T) {
	hp := gopodBrobotsSettings.HostPersonas
	if hp["DOC"].Name != "DOCTRINE BARFALLONYOU SPEWER" || hp["PIP"].Name != "PIP N STALL" {
		t.Errorf("DOC and PIP's own cards: %q %q", hp["DOC"].Name, hp["PIP"].Name)
	}
	if hp["Brobot 1"].Nickname != "BRUH!" || hp["Brobot 1"].Name == hp["DOC"].Name || hp["Brobot 2"].Name == hp["PIP"].Name {
		t.Errorf("a Brobot's own card: %+v / %+v", hp["Brobot 1"], hp["Brobot 2"])
	}
	if gopodHostedPersona("Brobot 1") != "DOC" || gopodHostedPersona("Brobot 2") != "PIP" {
		t.Errorf("hosts: %q %q", gopodHostedPersona("Brobot 1"), gopodHostedPersona("Brobot 2"))
	}
	esn := gopodBrobot1ESN
	gopodDialZero.Store(esn, true)
	at0, call0 := gopodAnswerer("Brobot 1"), gopodCallName("Brobot 1")
	gopodDialZero.Delete(esn)
	at100, call100 := gopodAnswerer("Brobot 1"), gopodCallName("Brobot 1")
	if at0 != "Brobot 1" || call0 != "Brobot 1" || at100 != "DOC" || call100 != "BARFALLONYOU" {
		t.Errorf("Brobot 1 at 0%%: %q %q; at 100%%: %q %q", at0, call0, at100, call100)
	}
	// A Brobot's card has none of DOC's sections, and DOC's none of his.
	if s := gopodPersonaProfileSections("DOC")["Who I am"]; len(s) > 0 && strings.Contains(strings.Join(s, " "), "BRUH") {
		t.Errorf("DOC's card has the Brobot in it: %q", s)
	}
}

// 2026-11-15 (live: Brobot 1's own battery line headed "DOC:"): a line the body
// says about itself is headed by the Brobot, whatever persona he hosts.
func TestBodyLineIsTheBrobots(t *testing.T) {
	gopodDialZero.Delete(gopodBrobot1ESN)
	gopodChatFeedAddLiveBy(gopodBrobot1ESN, "Whoa, bro... I'm feeling dizzy.", "Brobot 1")
	gopodChatFeed.Lock()
	last := gopodChatFeed.events[len(gopodChatFeed.events)-1]
	gopodChatFeed.Unlock()
	if last.By != "Brobot 1" {
		t.Errorf("body line by: %+v", last)
	}
}

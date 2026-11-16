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
	if c := gopodBrobotsSettings.PersonaColours["Brobot 1"].Border; c != "#c9a227" {
		t.Errorf("Brobot 1: %q", c)
	}
	if c := gopodBrobotsSettings.PersonaColours["DOC"].Border; c != "#ff3b30" {
		t.Errorf("DOC: %q", c)
	}
	if !gopodAtDialZero("Brobot 1") && gopodBrobotEyeHex("Brobot 1", true) != "#ff3b30" {
		t.Errorf("DOC answering: %s", gopodBrobotEyeHex("Brobot 1", true))
	}
	if gopodBrobotEyeHex("Brobot 1", false) != "#c9a227" {
		t.Errorf("Brobot 1 at rest: %s", gopodBrobotEyeHex("Brobot 1", false))
	}
}

// 2026-11-15 (operator: "all the decoration stuff should be in external editable
// callable css/js/json files"): the page's settings carry persona_colours, the
// one place - and no persona card holds a colour.
func TestPageGetsPersonaColours(t *testing.T) {
	var out struct {
		PersonaColours map[string]map[string]string `json:"persona_colours"`
	}
	full, _ := gopodSettingsFull()
	if err := json.Unmarshal(gopodPersonaProfileLabelsJSON(full), &out); err != nil {
		t.Fatal(err)
	}
	if out.PersonaColours["BUTLER"]["background"] != "#2E0A57" || out.PersonaColours["DOC"]["border"] != "#ff3b30" || out.PersonaColours["Brobot 1"]["border"] != "#c9a227" {
		t.Errorf("persona colours: %v", out.PersonaColours)
	}
	for who := range gopodBrobotsSettings.HostPersonas {
		if head, _ := gopodPersonaProfileReadOne(gopodPersonaProfileFileIn(gopodBrobotsSettings.PersonaProfiles.Folder, who)); head["colours"] != "" {
			t.Errorf("%s's card still holds colours", who)
		}
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
	// 2026-11-15: each card names its body; the voices have none.
	if hp["DOC"].Body != "Brobot 1" || hp["Brobot 1"].Body != "Brobot 1" || hp["PIP"].Body != "Brobot 2" || hp["BUTLER"].Body != "" || hp["CHALK"].Body != "" {
		t.Errorf("bodies: DOC %q, Brobot 1 %q, PIP %q, KEYNOTE %q, CHALK %q", hp["DOC"].Body, hp["Brobot 1"].Body, hp["PIP"].Body, hp["BUTLER"].Body, hp["CHALK"].Body)
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

// 2026-11-16 (bubble log: "Numpad to DOC" over a KP1off line): a key names the
// persona it reaches.
func TestKeyReaches(t *testing.T) {
	for key, want := range map[string]string{"KP1": "DOC", "KP1off": "Brobot 1", "KP2": "PIP", "KP2off": "Brobot 2", "KP0": "BUTLER"} {
		_, target, err := gopodNumpadBrobot(key)
		if err != nil {
			t.Fatal(err)
		}
		if got := gopodKeyReaches(target); got != want {
			t.Errorf("%s (%s) reaches %q, want %q", key, target, got, want)
		}
	}
}

// 2026-11-16: the lingo improver - the robots, the system voice and every persona
// voice, each with its lingo; its page words are system texts.
func TestLingoImproverVoices(t *testing.T) {
	want := map[string]string{"Brobot 1": "brobots", "Brobot 2": "brobots", gopodBrobotsSettings.SignOff.SystemRole: "butler", "CHALK": "reporters", "CHALK reporter": "reporters"}
	got := map[string]string{}
	for _, v := range GopodLingoVoices() {
		got[v.Voice] = v.Lingo
	}
	for k, l := range want {
		if got[k] != l {
			t.Errorf("%s: lingo %q, want %q", k, got[k], l)
		}
	}
	tx := GopodSystemTexts("lingo_improver.")
	for _, k := range []string{"title", "intro", "voice", "say", "as_said", "keep", "pick_and_type"} {
		if tx[k] == "" {
			t.Errorf("no lingo_improver.%s text", k)
		}
	}
	if said, _, note := GopodLingoTry("nobody", "hi"); said != "" || note == "" {
		t.Errorf("unknown voice: %q %q", said, note)
	}
}

// 2026-11-16 (operator: "New voice please * for goverlord = Keys settings with
// lower/deeper voice"): Goverlord's lines - KEYS' voice, lower; KEYS unchanged.
func TestGoverlordVoiceDeeper(t *testing.T) {
	g := gopodBrobotsSettings.SignOff.SystemRole
	if gopodReporterPitch(g) >= 0 || gopodReporterPitch(gopodBrobotsSettings.SystemBubble.VoicePersona) != 0 {
		t.Errorf("Goverlord %v cents, KEYS %v cents", gopodReporterPitch(g), gopodReporterPitch(gopodBrobotsSettings.SystemBubble.VoicePersona))
	}
}

// 2026-11-16 (operator's picker: "-0 Numlock OFF: Goverlord", "+0 KEYS"): KP0 with
// NumLock off is Goverlord (the system), not asked in the comm check; a typed line
// to KP0 with NumLock on reaches KEYS.
func TestKP0OffIsGoverlord(t *testing.T) {
	label, target, err := gopodNumpadBrobot("KP0off")
	if err != nil || target != gopodTypedSystem || label != "Goverlord" {
		t.Fatalf("KP0off: %q %q %v", label, target, err)
	}
	if l, ok := gopodNumpadKeyLabelIn(gopodNumpadLayout(), "KP0off"); !ok || l != "Goverlord" {
		t.Errorf("KP0off's picker label: %q", l)
	}
	if k, ok := gopodTypedKey("kp:0"); !ok || k != "KP0" {
		t.Errorf("kp:0: %q %v", k, ok)
	}
	if gopodKeyReaches(target) != gopodBrobotsSettings.SignOff.SystemRole {
		t.Errorf("KP0off reaches %q", gopodKeyReaches(target))
	}
}

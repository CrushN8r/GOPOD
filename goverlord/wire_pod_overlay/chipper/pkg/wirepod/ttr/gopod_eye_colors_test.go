// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_eye_colors_test.go
package wirepod_ttr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
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

// a card's border is the robot's eye colour - hex to hue and saturation and back; DOC's red while
// he answers, Brobot 1's gold after.
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

// the page's settings carry persona_colours, the one place - and no persona card holds a colour.
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

// Keys said CRUSHN8R like everyone): Keys' own way wins, whatever the case.
func TestKeysSaysCrushN8rHisWay(t *testing.T) {
	if got := gopodSaysFor("BUTLER", "CRUSHN8R Crew"); !strings.Contains(got, "crushin-ate-uh") {
		t.Errorf("Keys: %q", got)
	}
	if got := gopodSaysFor("CHALK", "CRUSHN8R Crew"); !strings.Contains(got, "crushin-ater") {
		t.Errorf("CHALK: %q", got)
	}
}

// four personas, four keys; a body is the persona answering from it only through gopodAnswerer.
func TestBrobotsDecoupledFromDocAndPip(t *testing.T) {
	hp := gopodBrobotsSettings.HostPersonas
	if hp["DOC"].Name != "DOCTRINE BARFALLONYOU SPEWER" || hp["PIP"].Name != "PIPPEN STALL" {
		t.Errorf("DOC and PIP's own cards: %q %q", hp["DOC"].Name, hp["PIP"].Name)
	}
	if hp["Brobot 1"].Nickname != "BRUH-1" || hp["Brobot 1"].Name == hp["DOC"].Name || hp["Brobot 2"].Name == hp["PIP"].Name {
		t.Errorf("a Brobot's own card: %+v / %+v", hp["Brobot 1"], hp["Brobot 2"])
	}
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
	if s := gopodPersonaProfileSections("DOC")["Who I am"]; len(s) > 0 && strings.Contains(strings.Join(s, " "), "BRUH") {
		t.Errorf("DOC's card has the Brobot in it: %q", s)
	}
}

// Brobot 1's own battery line headed "DOC:"): a line the body says about itself is headed by the
// Brobot, whatever persona he hosts.
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

// "Numpad to DOC" over a KP1off line): a key names the persona it reaches.
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

// the lingo improver - the robots, the system voice and every persona voice, each with its lingo;
// its page words are system texts.
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

// KP0 with NumLock off is Goverlord (the system), not asked in the comm check; a typed line to KP0
// with NumLock on reaches KEYS.
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

// a system command typed to KEYS (+0) runs, as on -0.
func TestKeysTakesSystemCommands(t *testing.T) {
	if k, ok := gopodTypedKey("kp:0"); !ok || k != "KP0" || !gopodSystemCommandMatch("system-comm-check") {
		t.Fatalf("kp:0 -> %q %v", k, ok)
	}
}

// GOVERLORD has his own deeper voice; KEYS (the system's lines) keeps his; Goverlord says an answer
// only right after a -0 call.
func TestGoverlordOnlyWhenCalled(t *testing.T) {
	if gopodBrobotsSettings.ReporterVoices.Hosts[gopodGoverlordHost] == "" || gopodReporterPitch(gopodGoverlordHost) >= 0 {
		t.Errorf("GOVERLORD: voice %q, pitch %v", gopodBrobotsSettings.ReporterVoices.Hosts[gopodGoverlordHost], gopodReporterPitch(gopodGoverlordHost))
	}
	if gopodReporterPitch(gopodBrobotsSettings.SignOff.SystemRole) != 0 || gopodReporterPitch(gopodBrobotsSettings.SystemBubble.VoicePersona) != 0 {
		t.Error("KEYS' voice (the system's lines) was changed")
	}
	gopodGoverlordCalledAt.Store(0)
	gopodGoverlordAnswer("not called") // no call: nothing said, nothing stored
	if gopodGoverlordCalledAt.Load() != 0 {
		t.Error("an answer without a call")
	}
	if !gopodSpeaksAsSystem(gopodGoverlordHost) {
		t.Error("GOVERLORD should speak the system's lingo")
	}
}

// never cut - sentences, comma pieces or voice chunks.
func TestKeepWholePhrase(t *testing.T) {
	text := "GOPOD is AI Wordplay! Explain the MAth! It's a game."
	for _, pieces := range [][]string{gopodWholeSentences(text), gopodSentences(text)} {
		whole := false
		for _, p := range pieces {
			if strings.Contains(p, "AI Wordplay! Explain the MAth!") {
				whole = true
			}
		}
		if !whole {
			t.Errorf("cut: %q", pieces)
		}
	}
}

// the bubble names where its FACT came from; the prompt keeps the words; the score shows only when
// it changed.
func TestBrainBubbleLooksUp(t *testing.T) {
	q := "Next up, CHALKRA! Introduce yourself to the chat"
	bubble := strings.Join(gopodBrainBubbleLines("CHALK", q), "\n")
	if !strings.Contains(bubble, "LOOKED UP: chalk-bored.md") || strings.Contains(bubble, "FACT:") {
		t.Errorf("the bubble: %q", bubble)
	}
	if prompt := gopodBrainText("CHALK", q); !strings.Contains(prompt, "FACT: I'm the proud Display Gomad") {
		t.Errorf("the prompt lost its FACT: %q", prompt)
	}
	gopodBrainScoreShown.Store("")
	if first := gopodBrainScoreRows(); len(first) == 0 {
		t.Error("the score's first showing")
	}
	if again := gopodBrainScoreRows(); len(again) != 0 {
		t.Errorf("the same score again: %q", again)
	}
}

// a question picks the pool's cards by the words it says; the brain names them and says their
// lines.
func TestLookupPicksCards(t *testing.T) {
	if _, err := os.Stat(filepath.Join(gopodRepoRoot(), "README.md")); err != nil {
		t.Skip("not in the repo here")
	}
	if len(gopodLookupCards()) < 20 {
		t.Fatalf("the pool: %d cards", len(gopodLookupCards()))
	}
	for q, want := range map[string]string{
		"How do I book the Brobots for a birthday party?": "05_birthday_and_merch.md",
		"How much does booking cost?":                     "BOOKING.md",
		"When is the next livestream?":                    "LIVESTREAM.md",
	} {
		fact, src := gopodLookupFact("", q)
		if !strings.Contains(strings.Join(src, ","), want) || fact == "" {
			t.Errorf("%q: looked up %v, fact %q", q, src, fact)
		}
	}
	if fact, src := gopodLookupFact("", "ok"); len(src) != 1 || src[0] != "README.md" || fact == "" {
		t.Errorf("a line with no card's words: %v %q", src, fact)
	}
}

// asked to look, GEMMA uses her eyes; what she saw is every brain's ROOM while fresh; the safeguard
// is off.
func TestGemmaEyes(t *testing.T) {
	ge := gopodBrobotsSettings.GemmaEyes
	if ge.Safeguard || len(ge.Device) == 0 || ge.Prompt == "" || !strings.Contains(ge.SayText, "{saw}") || ge.KeepMinutes <= 0 {
		t.Fatalf("gemma_eyes: %+v", ge)
	}
	if !gopodGemmaEyesAsked("GEMMA", "GEMMA, what do you see?") || gopodGemmaEyesAsked("CHALK", "what do you see?") {
		t.Error("only GEMMA, asked to look, looks")
	}
	gopodGemmaRoom.Lock()
	gopodGemmaRoom.line, gopodGemmaRoom.at = "Two people sit at a table.", time.Now()
	gopodGemmaRoom.Unlock()
	if !strings.Contains(strings.Join(gopodBrainLines("Brobot 1", "hey"), "\n"), "ROOM: Two people sit at a table.") {
		t.Error("the brain has no ROOM line")
	}
	gopodGemmaRoom.Lock()
	gopodGemmaRoom.at = time.Now().Add(-time.Hour)
	gopodGemmaRoom.Unlock()
	if gopodGemmaRoomLine() != "" {
		t.Error("an old look is still in the brain")
	}
}

// "." names who's asked; no name, the topic's owner; the typed picker's "." is the same key.
func TestAnyPersonaKey(t *testing.T) {
	label, target, err := gopodNumpadBrobot(gopodNumpadAnyKey)
	if err != nil || target != gopodTypedAny || label == "" {
		t.Fatalf("KPdot: %q %q %v", label, target, err)
	}
	if k, ok := gopodTypedKey("kp:dot"); !ok || k != gopodNumpadAnyKey {
		t.Errorf("kp:dot -> %q %v", k, ok)
	}
	for line, want := range map[string]string{
		"CHALKRA, what's on the board?": "CHALK",
		"Hey Zoomy, how's the session?": "PLAYHEAD",
		"DOC, explain the math":         "DOC",
		"Brobot 2, you there?":          "Brobot 2",
		"Keys, how do I start a show?":  "BUTLER",
	} {
		if got := gopodAnyNamed(line); got != want {
			t.Errorf("%q names %q, want %q", line, got, want)
		}
	}
	if gopodAnyNamed("what's the weather like?") != "" {
		t.Error("no one named")
	}
	if o := gopodAnyOwner("how do I replay a session from the scrolls?"); o != "" && o != gopodBrobotsSettings.SystemBubble.VoicePersona && o != "Brobot 1" {
		t.Errorf("a system topic's owner: %q", o)
	}
	if gopodAnyTarget("CHALK") != gopodPersonaMark+"CHALK" || !strings.HasPrefix(gopodAnyTarget("Brobot 1"), gopodDialZeroMark) {
		t.Errorf("targets: %q %q", gopodAnyTarget("CHALK"), gopodAnyTarget("Brobot 1"))
	}
}

// a near "system-" is read as one.
func TestSystemTypo(t *testing.T) {
	for in, want := range map[string]string{"sysem-comm-check": "system-comm-check", "sytsem-tour": "system-tour", "system-tour": "system-tour", "hello-there": "hello-there"} {
		if got := gopodSystemTypo(in); got != want {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
}

// the look-up's cards are the whole public repo - every tracked md file, and only those.
func TestLookupTrackedOnly(t *testing.T) {
	tracked := gopodLookupTracked()
	cards := gopodLookupCards()
	if len(cards) < 120 {
		t.Errorf("only %d cards - the whole tracked repo has more", len(cards))
	}
	for _, c := range cards {
		if !tracked[filepath.Clean(c.Path)] {
			t.Errorf("an untracked file is a card: %s", c.Path)
		}
	}
}

// a card gives its lines that say the question's words, not only its first ones; a persona's own
// card is never looked up for him.
func TestLookupDepth(t *testing.T) {
	fact, src := gopodLookupFact("", "how do I book the brobots for a birthday party?")
	if len(fact) < 200 {
		t.Errorf("a shallow look-up (%d letters): %v %q", len(fact), src, fact)
	}
	_, src = gopodLookupFact("Brobot 1", "Brobot 1, tell me about brobot 1")
	for _, s := range src {
		if s == "brobot-1.md" {
			t.Error("his own card, looked up again")
		}
	}
}

// a hub brings its "see also" cards along, one step out.
func TestLookupHubs(t *testing.T) {
	for q, want := range map[string][]string{
		"who gets the weather?":                  {"weather.md", "git-repo.md", "cache-pyc.md"},
		"who hands off to whom?":                 {"duo-handoff.md", "chalk-bored.md"},
		"how do I book the brobots?":             {"booking.md", "BOOKING.md"},
		"who takes the photos?":                  {"photos.md", "elle-elle-ehm-gemma.md"},
		"which key talks to whom on the numpad?": {"numpad-map.md", "numpad-mapper.md"},
	} {
		fact, src := gopodLookupFact("", q)
		got := strings.Join(src, " ")
		for _, w := range want {
			if !strings.Contains(" "+got+" ", " "+w+" ") {
				t.Errorf("%q: no %s in %v", q, w, src)
			}
		}
		if regexp.MustCompile(`[^.]\.\.[^.]`).MatchString(fact) || strings.Contains(fact, "My informal name") {
			t.Errorf("%q: junk in %q", q, fact)
		}
	}
}

// a run command runs only when it came on 0.
func TestRunOnlyOnZero(t *testing.T) {
	gopodRunFromZero.Store(false)
	if gopodRunAllowed("start fart joke") {
		t.Error("a run command off 0 ran")
	}
	gopodRunFromZero.Store(true)
	if !gopodRunAllowed("start fart joke") || gopodRunFromZero.Load() {
		t.Error("a run command on 0 didn't run, or the flag stayed")
	}
	if gopodRunAllowed("what's the weather") {
		t.Error("talk counted as a run")
	}
	if _, target, _ := gopodNumpadBrobot("KP0"); target != gopodPersonaMark+gopodBrobotsSettings.SystemBubble.VoicePersona {
		t.Errorf("KP0's target %q isn't KEYS", target)
	}
}

// the weather look-up walks weather.md's chain - the station, the hand-offs, PLAYHEAD, GIT REPO,
// CACHE PYC, the Brobot - each once, all named.
func TestLookupFullChain(t *testing.T) {
	fact, src := gopodLookupFact("Brobot 1", "who gets the weather?")
	got := " " + strings.Join(src, " ") + " "
	for _, w := range []string{"weather.md", "crushn8r_crewd_news_station.md", "duo-handoff.md", "playhead-zuminto.md", "git-repo.md", "cache-pyc.md", "brobot-1.md"} {
		if !strings.Contains(got, " "+w+" ") {
			t.Errorf("no %s in %v", w, src)
		}
	}
	if strings.Contains(fact, ". .") || len(src) > gopodBrobotsSettings.BrainChain.LookupChainMax {
		t.Errorf("chain: %d cards, %q", len(src), fact)
	}
}

// "gov or lord what say you" - Goverlord misheard): the mishearing is spelled back, so the "." key
// finds him.
func TestGovOrLord(t *testing.T) {
	text := gopodSpellNames("gov or lord what say you owe grand master")
	if !strings.HasPrefix(text, "Goverlord") || gopodAnyNamed(text) != "GOVERLORD" {
		t.Errorf("%q -> %q", text, gopodAnyNamed(text))
	}
}

// every persona's card has his intro guide; an intro is the model's own, guided by it - no code
// opening around it; PIP is PIPPEN STALL.
func TestIntroGuides(t *testing.T) {
	for _, who := range []string{"GOVERLORD", "BUTLER", "GEMMA", "INTEL 404 PNFH", "INTEL 404 PNFT", "DOC", "PIP", "CHALK", "PLAYHEAD", "CAMEO", "GIT REPO", "CACHE PYC"} {
		if gopodIntroGuide(who) == "" {
			t.Errorf("%s has no intro guide", who)
		}
	}
	if !strings.Contains(gopodIntroGuide("INTEL 404 PNFT"), "My partner Buzzer") {
		t.Error("SWITCH's partner is Buzzer")
	}
	if got := gopodIntroWrap("CHALK", "Chalk dust everywhere! *waves*"); got != gopodIntroGuide("CHALK") {
		t.Errorf("guided intro: %q", got)
	}
	if !strings.Contains(gopodPersonaPromptFor("CHALK", true), "Zoomy gave me") {
		t.Error("the guide isn't in the intro prompt")
	}
	if gopodBrobotsSettings.HostPersonas["PIP"].Name != "PIPPEN STALL" {
		t.Errorf("PIP: %q", gopodBrobotsSettings.HostPersonas["PIP"].Name)
	}
}

// every persona has a head (persona_heads), coloured from persona_colours; the Brobots keep the
// eyes; the reporters' heads come from CHALK's and PLAYHEAD's.
func TestPersonaHeads(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(gopodChipperRoot(), "webroot", "gopod_settings", "03_chat_page.json"))
	if err != nil {
		t.Fatal(err)
	}
	var page struct {
		Colours map[string]json.RawMessage `json:"persona_colours"`
		Heads   map[string]struct {
			From  string   `json:"from"`
			Props []string `json:"props"`
		} `json:"persona_heads"`
	}
	if err := json.Unmarshal(data, &page); err != nil {
		t.Fatal(err)
	}
	for who := range page.Colours {
		_, has := page.Heads[who]
		if brobot := who == "Brobot 1" || who == "Brobot 2"; brobot == has {
			t.Errorf("%s: head %v (only the Brobots keep the eyes)", who, has)
		}
	}
	for who, h := range page.Heads {
		from := h.From
		if from == "" {
			from = who
		}
		if _, ok := page.Colours[from]; !ok || len(h.Props) == 0 {
			t.Errorf("%s: no colours (%s) or no props", who, from)
		}
	}
	if page.Heads["CHALK reporter"].From != "CHALK" || page.Heads["PLAYHEAD reporter"].From != "PLAYHEAD" {
		t.Error("the reporters' heads come from CHALK's and PLAYHEAD's")
	}
}

// an emoji after a sentence's end stays with it; a persona's name-and-colon label at a sentence's
// start goes.
func TestEmojiStaysAndLabelGoes(t *testing.T) {
	got := gopodCutAt("Think social media, for the internet itself. 🌎 I'll be taking care of it. 😎👀 Done.", gopodHardEnd)
	if len(got) != 3 || got[0] != "Think social media, for the internet itself. 🌎" || got[1] != "I'll be taking care of it. 😎👀" {
		t.Errorf("pieces: %q", got)
	}
	if r := gopodDropSpeakerLabels("I am GOVERLORD, GOPOD's brain. KEYNOTE: The chat began. Note: keep it."); r != "I am GOVERLORD, GOPOD's brain. The chat began. Note: keep it." {
		t.Errorf("labels: %q", r)
	}
}

// a question no card matches gets the repo's README and that line.
func TestLookupDefaultsToReadme(t *testing.T) {
	fact, src := gopodLookupFact("Brobot 1", "so what's new with you?")
	if len(src) == 0 || src[0] != "README.md" || !strings.Contains(fact, "I don't know. GOPOD Yourself") {
		t.Errorf("default: %v %q", src, fact)
	}
}

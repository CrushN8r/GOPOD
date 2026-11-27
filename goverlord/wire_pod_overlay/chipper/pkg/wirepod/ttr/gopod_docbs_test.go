// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_docbs_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDocBS(t *testing.T) {
	saved := gopodBrobotsSettings.DocBS
	defer func() { gopodBrobotsSettings.DocBS = saved }()
	gopodGameReset("test")
	gopodDocBSOnFor.Range(func(k, _ any) bool { gopodDocBSOnFor.Delete(k); return true }) // a clean slate
	gopodGivenSet(gopodBrobot1ESN, nil)
	gopodGivenSet(gopodBrobot2ESN, nil)
	gopodBrobotsSettings.DocBS.Share = 100
	gopodBrobotsSettings.DocBS.Chains = nil    // the equation roast, the chain this test checks
	gopodBrobotsSettings.DocBS.SlotPrompt = "" // no model in the test: the fallback slots
	q := "Tell me about the rig"
	raw := "thinking___ One dolly, two crates, bro."
	t0 := time.Now()
	got := gopodShapeBrobotAnswer(gopodBrobot1ESN, q, raw)
	t.Logf("shaped in %v", time.Since(t0))
	for _, want := range []string{"Thought Quality (TQ) = (Your Confidence × Volume of the Claim) ÷ (Brain Cells Used)", "ERROR", "Corrected formula", "One dolly, two crates", "The math has spoken. And it's"} {
		if !strings.Contains(got, want) && !(want == "Corrected formula" && strings.Contains(got, "the real number")) {
			t.Errorf("DocBS answer has no %q:\n%s", want, got)
		}
	}
	if strings.Index(got, "ERROR") > strings.Index(got, "One dolly") || strings.Index(got, "One dolly") > strings.Index(got, "The math has spoken") {
		t.Errorf("DocBS beats out of order:\n%s", got)
	}
	t0 = time.Now()
	if b2 := gopodShapeBrobotAnswer(gopodBrobot2ESN, q, raw); strings.Contains(b2, "The math has spoken") {
		t.Errorf("Brobot 2 went DocBS")
	}
	t.Logf("brobot 2 shaped in %v", time.Since(t0))
	if gopodDocBSOn(gopodBrobot1ESN, "thanks Brobots") {
		t.Errorf("a thank-you went DocBS")
	}
	gopodBrobotsSettings.DocBS.Share = 0
	if !gopodDocBSOn(gopodBrobot1ESN, "Believe the sky is green all you want. It doesn't make it any less blue!! Explain the Math!!") {
		t.Errorf("an Explain the Math line the engine can't work out isn't DocBS")
	}
	if gopodDocBSOn(gopodBrobot1ESN, "What's 3/4 as a decimal? Explain the Math!") {
		t.Errorf("a computable Explain the Math line went DocBS")
	}
	t.Logf("sample:\n%s", got)
	gopodDocBSDecks.Lock()
	gopodDocBSDecks.left = map[string][]string{}
	gopodDocBSDecks.Unlock()
	seen := map[string]bool{}
	for i := 0; i < 5; i++ {
		l := gopodDocBSPick("errors")
		if seen[l] {
			t.Errorf("errors deck repeated %q before running out", l)
		}
		seen[l] = true
	}
}

// ## EOF

// the dial sets which chains a DocBS answer can be - the short ones from 0, the full pipeline high,
// 11/10 at 110 - each picked from a deck, never the same twice until it runs out; with no chains
// set, the equation roast.
func TestDocBSChainsByDial(t *testing.T) {
	old := gopodBrobotsSettings
	defer func() { gopodBrobotsSettings = old }()
	gopodBrobotsSettings.DocBS.Chains = []gopodDocBSChain{
		{Name: "short", MinShare: 0, Beats: []string{"roasts"}},
		{Name: "full", MinShare: 80, Beats: []string{"roasts", "gutshots"}},
		{Name: "11/10", MinShare: 110, Beats: []string{"crew"}},
	}
	pick := func(share, n int) map[string]int {
		gopodBrobotsSettings.DocBS.Share = share
		seen := map[string]int{}
		for i := 0; i < n; i++ {
			seen[gopodDocBSChainPick().Name]++
		}
		return seen
	}
	if s := pick(20, 4); s["short"] != 4 {
		t.Errorf("share 20: only the short chain, got %v", s)
	}
	if s := pick(100, 4); s["short"] != 2 || s["full"] != 2 || s["11/10"] != 0 {
		t.Errorf("share 100: short and full in turn, not 11/10, got %v", s)
	}
	if s := pick(110, 3); s["short"] != 1 || s["full"] != 1 || s["11/10"] != 1 {
		t.Errorf("share 110: every chain once per deck, got %v", s)
	}
	gopodBrobotsSettings.DocBS.Chains = nil
	if c := gopodDocBSChainPick(); c.Name != gopodDocBSChainDefault.Name {
		t.Errorf("no chains: the equation roast, got %q", c.Name)
	}
}

// a persona's chain - its pools' lines around its reply, the beats before "reply" first; none at
// share 0.
func TestPersonaChains(t *testing.T) {
	old := gopodBrobotsSettings
	defer func() { gopodBrobotsSettings = old }()
	dir := t.TempDir()
	gopodBrobotsSettings.PersonaChains.Folder = dir
	root := gopodPersonaChainDir("CHALK")
	os.MkdirAll(root, 0o755)
	os.WriteFile(filepath.Join(root, "openers.txt"), []byte("Chalk opener line.\n"), 0o644)
	os.WriteFile(filepath.Join(root, "closers.txt"), []byte("Chalk closer line.\n"), 0o644)
	gopodBrobotsSettings.PersonaChains.Personas = map[string]gopodPersonaChains{
		"CHALK": {Share: 100, Chains: []gopodDocBSChain{{Name: "open-close", Beats: []string{"openers", "reply", "closers"}}}},
	}
	got := gopodPersonaChainWrap("CHALK", "The reply.")
	if !strings.HasPrefix(got, "Chalk opener line.") || !strings.Contains(got, "The reply.") || !strings.HasSuffix(got, "Chalk closer line.") {
		t.Fatalf("chain around the reply: %q", got)
	}
	gopodBrobotsSettings.PersonaChains.Personas = map[string]gopodPersonaChains{"CHALK": {Share: 0}}
	if got := gopodPersonaChainWrap("CHALK", "The reply."); got != "The reply." {
		t.Fatalf("share 0: the reply alone, got %q", got)
	}
}

// a persona's profile is looked up for its prompt - its Self and Voice lines, not The name; no
// profile, nothing.
func TestPersonaProfiles(t *testing.T) {
	old := gopodBrobotsSettings
	defer func() { gopodBrobotsSettings = old }()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "cache-pyc.md"), []byte("# Cache PYC\n- **nickname:** Cachey\n\n---\n\n## Self\n1. My name: {name}.\n2. Later (to come)\n\n## Voice\n1. Sharp.\n\n## The name\n1. Not said.\n"), 0o644)
	gopodBrobotsSettings.PersonaProfiles.Folder = dir
	gopodBrobotsSettings.PersonaProfiles.Sections = []string{"Self", "Voice"}
	gopodBrobotsSettings.PersonaProfiles.Text = "Profile: {lines}"
	gopodBrobotsSettings.HostPersonas = map[string]gopodHostPersona{"CACHE PYC": {Name: "Cache PYC"}}
	if got := gopodPersonaProfileText("CACHE PYC"); got != "Profile: My name: Cache PYC. Sharp." {
		t.Fatalf("profile: %q", got)
	}
	if got := gopodPersonaProfileText("NOBODY"); got != "" {
		t.Fatalf("no profile: %q", got)
	}
	head, _ := gopodPersonaProfileRead(gopodPersonaProfileFile("CACHE PYC"))
	if head["label"] != "Cache PYC" || head["nickname"] != "Cachey" {
		t.Fatalf("head: %v", head)
	}
}

// a Gomad's answer is the words said - stage notes, quote marks and "I say" tags come out.
func TestPersonaAnswerClean(t *testing.T) {
	if got := gopodPersonaStageRE.ReplaceAllString("Intel 404 (Page Not Found Here) on duty (laughs).", ""); got != "Intel 404 (Page Not Found Here) on duty ." {
		t.Errorf("stage notes: %q", got)
	}
	if got := gopodStripEcho("Hey PLAYHEAD! You party animal! Introduce yourself fully. I'm PLAYHEAD.", "Hey PLAYHEAD! You party animal! Introduce yourself fully to the chat"); got != "I'm PLAYHEAD." {
		t.Errorf("echo: %q", got)
	}
	in := `"Intel 404 PNFT here," I say flatly, "on patrol." [The operator greets me.] Done (chuckles).`
	got := gopodPersonaStageRE.ReplaceAllString(in, "")
	got = gopodPersonaSayTagRE.ReplaceAllString(got, "")
	got = strings.Join(strings.Fields(strings.NewReplacer("“", "", "”", "", "\"", "").Replace(got)), " ")
	got = gopodPersonaGapRE.ReplaceAllString(got, "$1")
	if got != "Intel 404 PNFT here, on patrol. Done." && got != "Intel 404 PNFT here on patrol. Done." {
		t.Fatalf("clean: %q", got)
	}
}

// the intro recipe - the operator's tester gets each participant's Self points and everyone else
// here; any other line gets none; a Brobot's intro answer is the brobots_intro chain.
func TestIntroRecipe(t *testing.T) {
	tester := "Hey Pip! You party animal! Introduce yourself fully to the chat and name who all else is here in the chat with us!!"
	for _, who := range []string{"Brobot 2", "CACHE PYC", "BUTLER"} {
		r := gopodIntroRecipe(who, tester)
		if r == "" || strings.Contains(r, "My informal name (nickname)") || strings.Contains(r, ": Brobot 0") || gopodIntroCastLine(who) != "" || strings.Contains(r, "I answer to ") || !strings.Contains(r, "intro guide") {
			t.Errorf("%s's intro recipe: %q", who, r)
		}
		if strings.Contains(r, "{") {
			t.Errorf("%s's intro recipe has an unfilled label: %q", who, r)
		}
	}
	if r := gopodIntroOpening("BUTLER"); !strings.Contains(r, `I'm KEYNOTE VALE, my nickname is KEYS, because I hold all the keys.`) {
		t.Errorf("the Butler's opening: %q", r)
	}
	if r := gopodIntroOpening("CAMEO"); !strings.Contains(r, `I'm CAMEO LOOKS, my nickname is CHAMMY, because it's Chamfer plus Chameleon - I cut the edge and change faces at the same time.`) {
		t.Errorf("CAMEO's opening: %q", r)
	}
	if r := gopodIntroOpening("GIT REPO"); !strings.Contains(r, `I'm GIT REPO, my nickname is REPO, because Git is code version control, and Repo is a repossessor, not a repository. Git is short for 'Git yer butt in gear!!'`) {
		t.Errorf("Git Repo's opening: %q", r)
	}
	if r := gopodIntroRecipe("Brobot 2", "what's the weather"); r != "" {
		t.Errorf("not an intro: %q", r)
	}
	gopodIntroFor.Store("zz-esn", true)
	defer gopodIntroFor.Delete("zz-esn")
	if g := gopodBrobotChainGroup("zz-esn"); g != "brobots_intro" {
		t.Errorf("an intro answer's chain: %q", g)
	}
}

// the source check finds nothing broken in GOPOD's own sources - only lines the operator still has
// to write ("to come").
func TestSourceCheck(t *testing.T) {
	problems, checked := gopodSourceCheck()
	if checked < 50 {
		t.Errorf("only %d things checked", checked)
	}
	for _, p := range problems {
		if !strings.Contains(p, "still to come") {
			t.Errorf("source check: %s", p)
		}
	}
	t.Logf("%d checked, %d to come", checked, len(problems))
}

// the brain chain - every answer a code-made record; the prompt's brain has short-term, long-term
// and the flow; the bubble shows them; a reset starts a new chain.
func TestBrainChain(t *testing.T) {
	gopodBrainReset()
	defer gopodBrainReset()
	if b := gopodBrainText("Brobot 1", "hey"); strings.Contains(b, "TASK:") || strings.Contains(b, "TIME:") || strings.Contains(b, "WHO DOES WHAT:") {
		t.Errorf("an answer's brain: %q", b)
	}
	gopodBrainAdd("Keys", "Hey Keys! Introduce yourself fully to the chat", "intro")
	gopodBrainAdd("Doc", "What is GOPOD?", "answer")
	b := gopodBrainText("Brobot 2", "Hey Pip! Introduce yourself fully")
	for _, want := range []string{"#1 Master CrushN8r to Keys", "#2 Master CrushN8r to Doc", "Spoken so far: Keys, Doc", "Introduced: Keys", "FACT: I'm the proud GOPOD sidekick", "I work for Math Aftermath News - we sell GOPOD: GOPOD Yourself!", "I answer to the Gomads, KEYNOTE, then Goverlord.", "Recent chat (newest last):"} {
		if !strings.Contains(b, want) {
			t.Errorf("brain has no %q:\n%s", want, b)
		}
	}
	gopodBrainReminders.Store("Brobot 2", []string{"Bingo", "Legacy", "wrap-up"})
	if a := gopodBrainLines("Brobot 2", "hey"); strings.Contains(strings.Join(a, "|"), "TIME:") || strings.Contains(strings.Join(a, "|"), "TASK:") {
		t.Errorf("an answer's brain lines: %q", a)
	}
	l := gopodBrainLines("Brobot 2", "Hey Pip! Introduce yourself fully")
	gopodBrainReminders.Delete("Brobot 2")
	if len(l) != 1 || !strings.HasPrefix(l[0], "FACT: ") || !strings.Contains(l[0], "I work with my host DOCTRINE BARFALLONYOU SPEWER") || strings.Contains(l[0], "Answering to me") || strings.Contains(l[0], "..") {
		t.Errorf("brain scaffold lines: %q", l)
	}
	for _, line := range append(l, b) {
		if strings.Contains(line, "Last:") || strings.Contains(line, "Asked by:") {
			t.Errorf("off the scaffold: %q", line)
		}
	}
	gopodBrainPerformed("Bingo Test Live", "done")
	if b := gopodBrainText("Doc", "hey"); !strings.Contains(b, "Performed by Doc and Pip: Bingo Test Live") {
		t.Errorf("no performed run in the brain:\n%s", b)
	}
	gopodBrainReset(true)
	if r := gopodBrainSnapshot(); len(r) != 1 || r[0].Kind != "performed" {
		t.Errorf("a game reset keeps only the performed runs: %v", r)
	}
	gopodBrainReset()
	if len(gopodBrainSnapshot()) != 0 {
		t.Error("the reset kept the chain")
	}
}

// an answer's note is kept with its question - in everyone's recent chat, and in the brain when the
// same question comes again.
func TestBrainAfterAnswerNotes(t *testing.T) {
	gopodBrainReset()
	defer gopodBrainReset()
	gopodBrainAdd("Pip", "What's the weather like in Milan Italy right now?", "answer", "Milan's sunny, 13 degrees,⁣ perfect pizza weather!")
	b := gopodBrainText("Brobot 1", "hey")
	if !strings.Contains(b, `Pip: "Milan's sunny, 13 degrees, perfect pizza weather!"`) {
		t.Errorf("the note isn't in the recent chat:\n%s", b)
	}
	l := gopodBrainLines("Brobot 1", "what's the weather in milan italy?")
	found := false
	for _, x := range l {
		if strings.HasPrefix(x, "MEMORY: asked before (#1) - Pip:") {
			found = true
		}
	}
	if !found {
		t.Errorf("the same question has no note: %q", l)
	}
	for _, x := range gopodBrainLines("Brobot 1", "tell me about bingo for seniors") {
		if strings.HasPrefix(x, "MEMORY:") {
			t.Errorf("a new question got a note: %q", x)
		}
	}
}

// TIME, TASK, FACT, MEMORY): FACT is the point given to him, as its topic card writes it.
func TestBrainFactLine(t *testing.T) {
	esn := gopodBrobotESN("Brobot 1")
	gopodGiven.Lock()
	saved := gopodGiven.points[esn]
	gopodGiven.points[esn] = []string{"booking 3/4"} // point 1 is now the wide line
	gopodGiven.Unlock()
	defer func() { gopodGiven.Lock(); gopodGiven.points[esn] = saved; gopodGiven.Unlock() }()
	found := ""
	for _, l := range gopodBrainLines("Brobot 1", "Nothing booked this week. What would it cost to book you two for a gig?") {
		if strings.HasPrefix(l, "FACT: ") {
			found = l
		}
	}
	if !strings.Contains(found, "we don't sit in a closet") {
		t.Errorf("FACT line: %q", found)
	}
}

// "this morning?" at 3:05 PM got "it's the mornin'"): the TIME line names the part of the day; the
// model's "<br>" is a line break.
func TestTimePartAndBreaks(t *testing.T) {
	for h, want := range map[int]string{2: "evening", 9: "morning", 15: "afternoon", 18: "evening", 22: "evening"} {
		if got := gopodDayPart(h); got != want {
			t.Errorf("%d o'clock: %q, want %q", h, got, want)
		}
	}
	if got := gopodDisplayReply("I'm DOC. <br>I'm here."); strings.Contains(got, "<br") {
		t.Errorf("br kept: %q", got)
	}
}

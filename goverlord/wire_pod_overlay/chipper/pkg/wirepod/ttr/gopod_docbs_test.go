// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_docbs_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"time"
	"testing"
)

// TestDocBS: at 100% Doc, Brobot 1's answer is the equation roast around his
// reply - opener, equation, error, hidden multiplier, corrected formula before
// it, "The math has spoken" after it; Brobot 2 never; a computable Explain the
// Math line keeps its steps; decks don't repeat a line before they run out.
func TestDocBS(t *testing.T) {
	saved := gopodBrobotsSettings.DocBS
	defer func() { gopodBrobotsSettings.DocBS = saved }()
	gopodGameReset("test")
	gopodDocBSOnFor.Range(func(k, _ any) bool { gopodDocBSOnFor.Delete(k); return true }) // a clean slate
	gopodGivenSet(gopodBrobot1ESN, nil)
	gopodGivenSet(gopodBrobot2ESN, nil)
	gopodBrobotsSettings.DocBS.Share = 100
	gopodBrobotsSettings.DocBS.Chains = nil // the equation roast, the chain this test checks
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

// 2026-10-30 (the Master Library): the dial sets which chains a DocBS answer can
// be - the short ones from 0, the full pipeline high, 11/10 at 110 - each picked
// from a deck, never the same twice until it runs out; with no chains set, the
// equation roast.
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

// 2026-10-30: a persona's chain - its pools' lines around its reply, the beats
// before "reply" first; none at share 0.
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

// 2026-10-30: a persona's profile is looked up for its prompt - its Self and
// Voice lines, not The name; no profile, nothing.
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
	// The head is the persona's labels: its full name (the label) and nickname.
	head, _ := gopodPersonaProfileRead(gopodPersonaProfileFile("CACHE PYC"))
	if head["label"] != "Cache PYC" || head["nickname"] != "Cachey" {
		t.Fatalf("head: %v", head)
	}
}

// 2026-10-31: a Gomad's answer is the words said - stage notes, quote marks and
// "I say" tags come out.
func TestPersonaAnswerClean(t *testing.T) {
	// A real parenthesis stays; only stage notes go.
	if got := gopodPersonaStageRE.ReplaceAllString("Intel 404 (Page Not Found Here) on duty (laughs).", ""); got != "Intel 404 (Page Not Found Here) on duty ." {
		t.Errorf("stage notes: %q", got)
	}
	// An answer that opens with the question starts after it.
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

// 2026-10-31: the intro recipe - the operator's tester gets each participant's
// Self points and everyone else here; any other line gets none; a Brobot's intro
// answer is the brobots_intro chain.
func TestIntroRecipe(t *testing.T) {
	tester := "Hey Pip! You party animal! Introduce yourself fully to the chat and name who all else is here in the chat with us!!"
	for _, who := range []string{"Brobot 2", "CACHE PYC", "BUTLER"} {
		r := gopodIntroRecipe(who, tester)
		// 2026-10-31: the facts without their labels, and no operator in the list.
		// 2026-11-04: no cast line (operator: "remove who's here in the chat from
		// every response"); his working relationships are in his facts.
		if r == "" || strings.Contains(r, "My informal name (nickname)") || strings.Contains(r, ": Brobot 0") || gopodIntroCastLine(who) != "" || !strings.Contains(r, "I answer to ") {
			t.Errorf("%s's intro recipe: %q", who, r)
		}
		if strings.Contains(r, "{") {
			t.Errorf("%s's intro recipe has an unfilled label: %q", who, r)
		}
	}
	// 2026-11-06 (operator: "every intro should start similar to \"I'm [], my
	// nickname is [] ...\""): formal name, then the nickname, once.
	if r := gopodIntroRecipe("BUTLER", tester); !strings.Contains(r, `"I'm KEYNOTE VALE, my nickname is KEYS, because I hold all the keys."`) {
		t.Errorf("the Butler's opening: %q", r)
	}
	// 2026-11-04 (operator: "NEW RULE: Gomad names in CAPS!!"; "nickname Chammy
	// (cammy) * why, short for Chameleon").
	if r := gopodIntroRecipe("CAMEO", tester); !strings.Contains(r, `"I'm CAMEO LOOKS, my nickname is CHAMMY, because it's short for Chameleon."`) {
		t.Errorf("CAMEO's opening: %q", r)
	}
	// 2026-11-04 (operator: "Git Repo nickname = 'Repo' * 'Repo' = repossesser
	// (wordplay, lol), not repository"): his why ends on its own "!!'" - no extra stop.
	if r := gopodIntroRecipe("GIT REPO", tester); !strings.Contains(r, `"I'm GIT REPO, my nickname is REPO, because Git is code version control, and Repo is a repossessor, not a repository. Git is short for 'Git yer butt in gear!!'"`) || !strings.Contains(r, "I answer to PLAYHEAD ZUMINTU") {
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

// 2026-10-31: the source check finds nothing broken in GOPOD's own sources -
// only lines the operator still has to write ("to come").
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

// 2026-10-31: the brain chain - every answer a code-made record; the prompt's
// brain has short-term, long-term and the flow; the bubble shows them; a reset
// starts a new chain.
func TestBrainChain(t *testing.T) {
	gopodBrainReset()
	defer gopodBrainReset()
	if b := gopodBrainText("Brobot 1", "hey"); !strings.Contains(b, "first to speak") {
		t.Errorf("an empty chain's flow: %q", b)
	}
	gopodBrainAdd("Keys", "Hey Keys! Introduce yourself fully to the chat", "intro")
	gopodBrainAdd("Doc", "What is GOPOD?", "answer")
	b := gopodBrainText("Brobot 2", "Hey Pip! Introduce yourself fully")
	for _, want := range []string{"#1 Master CrushN8r to Keys", "#2 Master CrushN8r to Doc", "Spoken so far: Keys, Doc", "Introduced: Keys", "a) Introduce yourself - who you answer to and who works with you. b) KEYNOTE opened the session.", "My full name is PIP N STALL, the GOPOD sidekick.", "In the chat: GOPOD hosts, on the two Brobots (the dial at 100%): BARFALLONYOU [the host, on Brobot 1 - explains too hard, on purpose], PIP [the sidekick, on Brobot 2 - the room's pied piper] (me). GOPOD party-brobot reps (the dial at 0%): Brobot 1 [grumpy, impatient], Brobot 2 [clueless, happy] (me). ", "So far: 2 have spoken, 1 introduced themselves.", "Recent chat (newest last):"} {
		if !strings.Contains(b, want) {
			t.Errorf("brain has no %q:\n%s", want, b)
		}
	}
	// The operator's brain scaffold, the same in the bubble and the prompt; no
	// "Last:", no "Asked by:"; his reminders when his answer named some.
	gopodBrainReminders.Store("Brobot 2", []string{"Bingo", "Legacy", "wrap-up"})
	l := gopodBrainLines("Brobot 2", "hey")
	gopodBrainReminders.Delete("Brobot 2")
	if len(l) != 4 || !strings.HasPrefix(l[0], "My full name is PIP N STALL") || !strings.HasSuffix(l[2], "Guided Task Flow: Does the recent chat help me answer the question asked. Reminders: Bingo, Legacy, & wrap-up") || l[3] != "My relationships: I answer to the Gomads, KEYNOTE, then Goverlord. I work with my host BARFALLONYOU SQUAWKADOODLE - he explains too hard, I get it too late, and the room roots for me." {
		t.Errorf("brain scaffold lines: %q", l)
	}
	for _, line := range append(l, b) {
		if strings.Contains(line, "Last:") || strings.Contains(line, "Asked by:") {
			t.Errorf("off the scaffold: %q", line)
		}
	}
	// A run the Brobots performed: in their long-term memory, kept by a game
	// reset (the comm check's), gone with the session (the thanks ending).
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

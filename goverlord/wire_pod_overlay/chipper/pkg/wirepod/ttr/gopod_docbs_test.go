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
	os.WriteFile(filepath.Join(dir, "cache-pyc.md"), []byte("# Cache PYC\n\n## Self\n- My name: Cache PYC.\n\n## Voice\n- Sharp.\n\n## The name\n- Not said.\n"), 0o644)
	gopodBrobotsSettings.PersonaProfiles.Folder = dir
	gopodBrobotsSettings.PersonaProfiles.Sections = []string{"Self", "Voice"}
	gopodBrobotsSettings.PersonaProfiles.Text = "Profile: {lines}"
	if got := gopodPersonaProfileText("CACHE PYC"); got != "Profile: My name: Cache PYC. Sharp." {
		t.Fatalf("profile: %q", got)
	}
	if got := gopodPersonaProfileText("NOBODY"); got != "" {
		t.Fatalf("no profile: %q", got)
	}
}

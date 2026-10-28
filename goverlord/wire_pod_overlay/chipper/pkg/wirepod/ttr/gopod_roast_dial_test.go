// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_roast_dial_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRoastDial: each level's text goes in his pre-build context; human swears
// only from a level that allows them; a sensitive topic goes neutral with no
// spice unless the Jekyll/Hyde override is on; PLAYHEAD's prompt has his WIP bit.
func TestRoastDial(t *testing.T) {
	saved := gopodBrobotsSettings.RoastDial
	defer func() { gopodBrobotsSettings.RoastDial = saved }()
	rd := &gopodBrobotsSettings.RoastDial
	if len(rd.Levels) < 4 || rd.Level != 0 {
		t.Fatalf("roast_dial: %d levels, level %d", len(rd.Levels), rd.Level)
	}
	human := func(line string) bool {
		for _, w := range append(gopodHumanSwears(), gopodBrobotsSettings.Spice.HumanSwears...) {
			if strings.Contains(line, w+",") || strings.HasSuffix(strings.Split(line, ". Exclamations")[0], w) {
				return true
			}
		}
		return false
	}
	q := "Tell me about the rig"
	for i, l := range rd.Levels {
		rd.Level = i
		if got := gopodRoastLine(q); got != l.Text {
			t.Errorf("level %d: %q", i, got)
		}
		ctx := gopodPrebuildContext(gopodBrobot1ESN, q)
		if !strings.Contains(ctx, l.Text) {
			t.Errorf("level %d not in the pre-build context", i)
		}
		for k := 0; k < 30 && !l.HumanSwears; k++ {
			if human(gopodSpiceLine(q)) {
				t.Fatalf("level %d gave a human swear: %q", i, gopodSpiceLine(q))
			}
		}
	}
	rd.Level = len(rd.Levels) - 1
	sad := "my dad died last week"
	if gopodRoastLine(sad) != rd.SensitiveText || gopodSpiceLine(sad) != "" {
		t.Errorf("sensitive topic not neutral: %q / %q", gopodRoastLine(sad), gopodSpiceLine(sad))
	}
	rd.JekyllHyde = true
	if gopodRoastLine(sad) != rd.Levels[rd.Level].Text || gopodSpiceLine(sad) == "" {
		t.Errorf("Jekyll/Hyde didn't override")
	}
	// The private z-copy, uncensored, wins when it's there; the public list is starred.
	for _, w := range gopodBrobotsSettings.Spice.HumanSwears {
		if strings.ContainsAny(w, "aeiouAEIOU") {
			t.Errorf("public human swear %q has a vowel", w)
		}
	}
	if _, err := os.Stat(filepath.Join(gopodChipperRoot(), gopodBrobotsSettings.Spice.HumanSwearsPrivate)); err == nil {
		if l := gopodHumanSwears(); len(l) == 0 || strings.Contains(l[0], "*") {
			t.Errorf("the private list isn't used: %v", l)
		}
	}
	if !strings.Contains(gopodPersonaPrompt("PLAYHEAD"), "Chalkboard of Doom") {
		t.Errorf("PLAYHEAD's prompt has no WIP bit")
	}
}

// ## EOF

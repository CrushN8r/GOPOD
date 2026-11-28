// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_alias_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSystemAliasReadsTheCards: system-alias lists the cards' aliases, marks the chat ones, and
// refuses one its card keeps in the terminal - all from the cards, no list in code.
func TestSystemAliasReadsTheCards(t *testing.T) {
	if _, err := os.Stat(filepath.Join(gopodRepoRoot(), "alias_play_studio")); err != nil {
		t.Skip("not in the repo here")
	}
	list := gopodSystemAlias("system-alias")
	for _, want := range []string{"gofps", "gopower", "wpr", "pha0b"} {
		if !strings.Contains(list, want) {
			t.Errorf("the list has no %s", want)
		}
	}
	chat := 0
	for _, c := range gopodAliasCards() {
		chat += len(c.FromChat)
		for name := range c.FromChat {
			found := false
			for _, a := range c.Aliases {
				found = found || a[0] == name
			}
			if !found {
				t.Errorf("%s: From the chat names %s, not in its commandments", c.Title, name)
			}
		}
	}
	if chat == 0 {
		t.Error("no card marks an alias for the chat")
	}
	if got := gopodSystemAlias("system-alias wpr 2"); !strings.Contains(got, "terminal only") {
		t.Errorf("wpr from the chat: %q", got)
	}
	if got := gopodSystemAlias("system-alias nosuch"); !strings.Contains(got, "no card names") {
		t.Errorf("unknown alias: %q", got)
	}
	if got := gopodSystemAlias("system-alias gofps 30;reboot"); !strings.Contains(got, "letters, digits") {
		t.Errorf("bad argument: %q", got)
	}
}

// ## EOF

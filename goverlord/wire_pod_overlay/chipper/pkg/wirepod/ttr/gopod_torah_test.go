// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_torah_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTorahMatches: GOPOD_TORAH.md is the system as it is now (GOPOD_WRITE_TORAH=1
// writes it again); every system window command is in it.
func TestTorahMatches(t *testing.T) {
	if _, err := os.Stat(filepath.Join(gopodRepoRoot(), "README.md")); err != nil {
		t.Skip("not in the repo here")
	}
	page := gopodTorahPage()
	for _, want := range []string{"# The GOPOD Torah", "system-test-admin-run", "system-comm-check", "start bingo test", "torah"} {
		if !strings.Contains(page, want) {
			t.Errorf("the Torah has no %q", want)
		}
	}
	if os.Getenv("GOPOD_WRITE_TORAH") == "1" {
		if err := os.WriteFile(gopodTorahPath(), []byte(page), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	if gopodTorahBehind() {
		t.Errorf("GOPOD_TORAH.md doesn't match the system - GOPOD_WRITE_TORAH=1 (or system-torah) writes it again")
	}
}

// ## EOF

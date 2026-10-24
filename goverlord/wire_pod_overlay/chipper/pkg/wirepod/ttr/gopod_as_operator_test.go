// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_as_operator_test.go
package wirepod_ttr

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestPythonRunsAsOperator: every Python GOPOD starts goes through
// gopodAsOperator (Wire-Pod runs as root; the libraries are the operator's).
func TestPythonRunsAsOperator(t *testing.T) {
	if cmd := gopodAsOperator(exec.Command("true")); !strings.Contains(strings.Join(cmd.Env, "\n"), "\nHOME="+strings.TrimRight(gopodHomePath("~/"), "/")) {
		t.Error("gopodAsOperator no longer gives the operator's home (HOME)")
	}
	start := regexp.MustCompile(`exec\.Command(Context)?\(`)
	files, _ := filepath.Glob(filepath.Join(gopodChipperRoot(), "pkg/wirepod/ttr", "gopod_*.go"))
	if len(files) < 20 {
		t.Fatalf("GOPOD's Go files not found (%d)", len(files))
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") || filepath.Base(f) == "gopod_as_operator.go" {
			continue
		}
		data, _ := os.ReadFile(f)
		for i, line := range strings.Split(string(data), "\n") {
			// 2026-10-24 (operator: "RULE!! ANYTHING UNDER ROOT, FIX TO MY HOME"):
			// every program GOPOD starts runs with the operator's home - only the
			// system's own tools that must stay root (systemctl, lsusb) don't.
			if start.MatchString(line) && !strings.Contains(line, "gopodAsOperator(") && !strings.Contains(line, `"systemctl"`) && !strings.Contains(line, `"lsusb"`) {
				// a start wrapped on its own line just after (cmd.Env set first)
				lines := strings.Split(string(data), "\n")
				wrapped := false
				for j := i + 1; j < len(lines) && j <= i+3; j++ {
					if strings.Contains(lines[j], "gopodAsOperator(cmd)") {
						wrapped = true
					}
				}
				if !wrapped {
					t.Errorf("%s:%d starts a program without gopodAsOperator (root's home): %s", f, i+1, strings.TrimSpace(line))
				}
			}
		}
	}
}

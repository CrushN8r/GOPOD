// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_as_operator_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestPythonRunsAsOperator: every Python GOPOD starts goes through
// gopodAsOperator (Wire-Pod runs as root; the libraries are the operator's).
func TestPythonRunsAsOperator(t *testing.T) {
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
			if start.MatchString(line) && strings.Contains(strings.ToLower(line), "python") && !strings.Contains(line, "gopodAsOperator(") {
				t.Errorf("%s:%d starts Python without gopodAsOperator: %s", f, i+1, strings.TrimSpace(line))
			}
		}
	}
}

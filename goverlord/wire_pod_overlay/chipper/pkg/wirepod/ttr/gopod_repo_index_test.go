// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_repo_index_test.go
package wirepod_ttr

import (
	"os"
	"testing"
)

// TestRepoIndexLocked: index 2 and index 1 are written from the repo; a change shows here until
// GOPOD_WRITE_INDEX=1 writes them again, on purpose.
func TestRepoIndexLocked(t *testing.T) {
	for _, f := range []struct {
		path string
		make func() string
		name string
	}{
		{gopodRepoIndexPath(), gopodRepoIndexText, "repo index 2"},
		{gopodWordsIndexPath(), gopodWordsIndexText, "index 1 (the words people say)"},
		{gopodIndexQuestionsPath(), gopodIndexQuestionsText, "the index questions"},
	} {
		f := struct{ path, text, name string }{f.path, f.make(), f.name}
		if os.Getenv("GOPOD_WRITE_INDEX") == "1" {
			if err := os.WriteFile(f.path, []byte(f.text), 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		data, err := os.ReadFile(f.path)
		if err != nil || string(data) != f.text {
			t.Errorf("LOCKED: %s doesn't match the repo - GOPOD_WRITE_INDEX=1 writes it again", f.name)
		}
	}
}

// ## EOF

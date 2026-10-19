// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_edit_test.go
package wirepod_ttr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestJSONFileDelete: a key comes out of a group (first, middle, last; one line
// or many), the rest kept, the file still JSON.
func TestJSONFileDelete(t *testing.T) {
	src := "{\n  \"a\": {\n    \"x\": \"1\",\n    \"y\": \"2\",\n    \"z\": \"3\"\n  },\n  \"b\": {\"p\": 1, \"q\": 2, \"r\": 3}\n}\n"
	for _, c := range [][]string{{"a", "x"}, {"a", "y"}, {"a", "z"}, {"b", "p"}, {"b", "q"}, {"b", "r"}} {
		f := filepath.Join(t.TempDir(), "s.json")
		os.WriteFile(f, []byte(src), 0o644)
		if _, err := gopodJSONFileDelete(f, c); err != nil {
			t.Fatalf("%v: %v", c, err)
		}
		data, _ := os.ReadFile(f)
		var m map[string]map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("%v: not JSON: %s", c, data)
		}
		if _, still := m[c[0]][c[1]]; still || len(m[c[0]]) != 2 {
			t.Fatalf("%v: %s", c, data)
		}
	}
}

// TestLingoBrobots: a word added, changed and taken out of the Brobots' list.
func TestLingoBrobots(t *testing.T) {
	orig := gopodPronunciationRegistryPath
	defer func() { gopodPronunciationRegistryPath = orig }()
	data, err := os.ReadFile(orig)
	if err != nil {
		t.Skip("no template card")
	}
	gopodPronunciationRegistryPath = filepath.Join(t.TempDir(), "card.md")
	os.WriteFile(gopodPronunciationRegistryPath, data, 0o644)
	for _, step := range []struct {
		word, said string
		remove     bool
		want       string
	}{{"zorbix", "zor-bix", false, `{"from": "zorbix", "to": "zor-bix"}`}, {"zorbix", "zor-bicks", false, `"to": "zor-bicks"`}, {"zorbix", "", true, ""}} {
		if got := gopodLingoBrobots(step.word, step.said, step.remove); strings.Contains(got, "not changed") {
			t.Fatal(got)
		}
		now, _ := os.ReadFile(gopodPronunciationRegistryPath)
		m := gopodPronunciationBlockPattern.FindSubmatch(now)
		if m == nil || !json.Valid(m[1]) {
			t.Fatal("the list doesn't read")
		}
		if step.remove && strings.Contains(string(now), `"zorbix"`) || !step.remove && !strings.Contains(string(now), step.want) {
			t.Fatalf("%+v not done", step)
		}
	}
}

// TestTopicPointsWrite: the say lines and the numbered points stay one list.
func TestTopicPointsWrite(t *testing.T) {
	src := "label: x\nwords: a, b\nsay 1: One.\nsay 2: Two.\n---\nIntro. Reveal these, in order:\n1. One.\n2. Two.\n"
	lines := strings.Split(src, "\n")
	out := strings.Join(gopodTopicPointsWrite(lines, 4, []string{"One.", "Two!", "Three."}), "\n")
	for _, want := range []string{"say 2: Two!", "say 3: Three.", "---", "Intro. Reveal these", "2. Two!", "3. Three.", "words: a, b"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "say 2: Two.") || strings.Count(out, "1. One.") != 1 {
		t.Fatalf("old points left:\n%s", out)
	}
}

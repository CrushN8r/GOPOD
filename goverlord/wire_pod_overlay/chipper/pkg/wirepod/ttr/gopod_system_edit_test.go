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
	// 2026-10-29: the topic-cluster md format (TOPIC_TEMPLATE.md).
	src := "# x\n\n- **words:** a, b\n- **say 1:** One.\n- **say 2:** Two.\n\n---\n\nIntro. Reveal these, in order:\n1. One.\n2. Two.\n"
	lines := strings.Split(src, "\n")
	out := strings.Join(gopodTopicPointsWrite(lines, 5, []string{"One.", "Two!", "Three."}), "\n")
	for _, want := range []string{"- **say 2:** Two!", "- **say 3:** Three.", "\n\n---\n\n", "Intro. Reveal these", "2. Two!", "3. Three.", "- **words:** a, b", "# x"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "say 2:** Two.") || strings.Count(out, "1. One.") != 1 {
		t.Fatalf("old points left:\n%s", out)
	}
}

// TestTopicClusterFormat (2026-10-29, operator: "template out a topic-cluster
// format for md files"): every topic and chapter file is a topic-cluster md file -
// its settings lines all read by the one format reader, a blank line around
// "---" - and the reader takes "# label", "- **key:** value" and a bare "key: value".
func TestTopicClusterFormat(t *testing.T) {
	for line, want := range map[string][2]string{"# BBQs": {"label", "BBQs"}, "- **say 2:** Two: yes": {"say 2", "Two: yes"}, "words: a, b": {"words", "a, b"}} {
		if k, v, ok := gopodTopicHeadKV(line); !ok || k != want[0] || v != want[1] {
			t.Errorf("%q read as %q %q", line, k, v)
		}
	}
	if gopodTopicHeadLine("label", "BBQs") != "# BBQs" || gopodTopicHeadLine("say 1", "Hi") != "- **say 1:** Hi" {
		t.Error("the format writer changed")
	}
	files, _ := filepath.Glob(filepath.Join(canonicalChaptersDir, "*", "*"))
	// The template folder too (gopod_probes/_topic_cluster/).
	tpl, _ := filepath.Glob(filepath.Join(filepath.Dir(canonicalChaptersDir), "_topic_cluster", "*.md"))
	for _, f := range tpl {
		if filepath.Base(f) != "README.md" {
			files = append(files, f)
		}
	}
	n := 0
	for _, f := range files {
		if strings.HasSuffix(f, ".txt") {
			t.Errorf("%s: a topic is a .md file now", f)
			continue
		}
		if !strings.HasSuffix(f, ".md") {
			continue
		}
		n++
		data, _ := os.ReadFile(f)
		lines := strings.Split(string(data), "\n")
		for i, l := range lines {
			if strings.TrimSpace(l) == "---" {
				if i == 0 || strings.TrimSpace(lines[i-1]) != "" || i+1 >= len(lines) || strings.TrimSpace(lines[i+1]) != "" {
					t.Errorf("%s: no blank line around ---", f)
				}
				break
			}
			if strings.TrimSpace(l) == "" {
				continue
			}
			if _, _, ok := gopodTopicHeadKV(l); !ok || !(strings.HasPrefix(l, "# ") || strings.HasPrefix(l, "- **")) {
				t.Errorf("%s: %q isn't a topic-cluster line", f, l)
			}
		}
	}
	if n == 0 {
		t.Skip("no chapters folder here")
	}
}

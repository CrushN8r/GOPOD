// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_set_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSystemSetEditsInPlace (2026-10-17, operator: "system-set this-volume
// 13%"): one value changes, everything else stays byte for byte.
func TestSystemSetEditsInPlace(t *testing.T) {
	dir := t.TempDir()
	knobs := filepath.Join(dir, "knobs.json")
	orig := "{\n  \"song_id\": \"x\",\n  \"steps\": [\n    {\"step_id\": \"a_0101\", \"note\": \"sting\", \"volume\": \"\"},\n    {\"step_id\": \"b_0102\", \"note\": \"pause\"}\n  ]\n}\n"
	os.WriteFile(knobs, []byte(orig), 0o644)
	old, err := gopodJSONFileSet(knobs, []string{"steps", "0", "volume"}, `"13%"`)
	got, _ := os.ReadFile(knobs)
	if err != nil || old != `""` || string(got) != strings.Replace(orig, `"volume": ""`, `"volume": "13%"`, 1) {
		t.Errorf("replace: %v %q\n%s", err, old, got)
	}
	// A missing field is added to its step.
	old, err = gopodJSONFileSet(knobs, []string{"steps", "1", "pause_seconds"}, `2`)
	got, _ = os.ReadFile(knobs)
	if err != nil || old != "(none)" || !strings.Contains(string(got), `{"step_id": "b_0102", "note": "pause", "pause_seconds": 2}`) {
		t.Errorf("add: %v %q\n%s", err, old, got)
	}
	if _, err := gopodJSONFileSet(knobs, []string{"nope", "x"}, `1`); err == nil {
		t.Error("wrote into a missing object")
	}
	for kind, bad := range map[string]string{"master": "135%", "share": "loud", "number": "-1", "db": "+6db"} {
		if _, err := gopodSystemSetValue(kind, bad); err == nil {
			t.Errorf("%s %q accepted", kind, bad)
		}
	}
	if v, _ := gopodSystemSetValue("master", "35%"); v != "0.35" {
		t.Errorf("master 35%% = %s", v)
	}
	if v, _ := gopodSystemSetValue("share", "default"); v != `""` {
		t.Errorf("share default = %s", v)
	}
}

// TestSystemSetStoryText: a step's words change in story.md, nothing else.
func TestSystemSetStoryText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "story.md")
	orig := "# S\n\n## STEP a_0101\n\n> TEXT: Old words\n> more old words\n\n## STEP b_0102\n\n> TEXT: Keep me.\n"
	os.WriteFile(path, []byte(orig), 0o644)
	old, err := gopodStorySetText(path, "a_0101", "New words.")
	got, _ := os.ReadFile(path)
	want := "# S\n\n## STEP a_0101\n\n> TEXT: New words.\n\n## STEP b_0102\n\n> TEXT: Keep me.\n"
	if err != nil || old != "Old words more old words" || string(got) != want {
		t.Errorf("%v %q\n%s", err, old, got)
	}
	if _, err := gopodStorySetText(path, "zz_0101", "x"); err == nil {
		t.Error("wrote a missing step")
	}
}

// TestStoryMultiLine (2026-10-17, operator: "this multi-line edit broke the
// code ... Can you upgrade globally to support multi-line edits?"): a line right
// under "> TEXT:" is the text's next line; a blank line ends it; system-set
// replaces all of it.
func TestStoryMultiLine(t *testing.T) {
	story := "## STEP a_0101\n\n> TEXT: Voice ready to play.\nLanguage Model Loading...\n\nNot words.\n\n## STEP b_0102\n\n> TEXT: One.\n"
	texts, err := gopodInterviewStory(story)
	if err != nil || texts["a_0101"] != "Voice ready to play.\nLanguage Model Loading..." || texts["b_0102"] != "One." {
		t.Errorf("%v %q", err, texts)
	}
	path := filepath.Join(t.TempDir(), "story.md")
	os.WriteFile(path, []byte(story), 0o644)
	old, err := gopodStorySetText(path, "a_0101", "New.")
	got, _ := os.ReadFile(path)
	if err != nil || old != "Voice ready to play.\nLanguage Model Loading..." || strings.Contains(string(got), "Loading") {
		t.Errorf("%v %q\n%s", err, old, got)
	}
}

// TestSystemSetReads: the list and a song's steps read the real files, and a
// voice or numpad line is turned away.
func TestSystemSetReads(t *testing.T) {
	if h := gopodSystemSet("system-set"); !strings.Contains(h, "master-volume") || !strings.Contains(h, "00_llm_warming") {
		t.Errorf("help:\n%s", h)
	}
	if s := gopodSystemSet("system-set 00_llm_warming"); !strings.Contains(s, "goverlord_sting_0106") {
		t.Errorf("steps:\n%s", s)
	}
	if s := gopodSystemSet("system-set 00_llm_warming goverlord_warm_0101"); !strings.Contains(s, "text: ") {
		t.Errorf("step:\n%s", s)
	}
	if !gopodSystemSetMatch("System-Set master-volume 13%") || gopodSystemSetMatch("system settings") {
		t.Error("match")
	}
}

// TestSystemLocate (2026-10-17, operator: "Please build system-locate"): a
// setting, a song step and a word are found with their file and line.
func TestSystemLocate(t *testing.T) {
	if s := gopodSystemLocate("system-locate master-volume"); !strings.Contains(s, "gopod_settings/03_chat_page.json:") {
		t.Errorf("setting: %s", s)
	}
	s := gopodSystemLocate("system-locate 00_llm_warming goverlord_sting_0106 volume")
	if !strings.Contains(s, "00_llm_warming/knobs.json:") || !strings.Contains(s, "story.md:") {
		t.Errorf("step: %s", s)
	}
	t.Log(gopodSystemLocate("system-locate master-volume"))
	t.Log(s)
	t.Log(gopodSystemLocate("system-locate volume"))
	if s := gopodSystemLocate("system-locate awaits"); !strings.Contains(s, "session.awaits_text") {
		t.Errorf("word: %s", s)
	}
	if l := gopodSystemLocateList(); !strings.Contains(l, "master-volume: ") || !strings.Contains(l, "00_llm_warming/story.md") {
		t.Errorf("locate list:\n%s", l)
	}
	if !gopodSystemSetMatch("system-set-list") || !gopodSystemSetMatch("system-locate-list") {
		t.Error("list match")
	}
	if !gopodSystemSetMatch("system-locate volume") {
		t.Error("match")
	}
}

// ## EOF

// TestSystemUndo (2026-10-17, operator: "Add: system-undo, system-status,
// system-help"): an edit is put back; one changed since is left alone.
func TestSystemUndo(t *testing.T) {
	saved := gopodUndoPath
	defer func() { gopodUndoPath = saved }()
	dir := t.TempDir()
	gopodUndoPath = filepath.Join(dir, "undo.json")
	file := filepath.Join(dir, "k.json")
	os.WriteFile(file, []byte(`{"a": 1}`), 0o644)
	gopodUndoBegin("system-set x a 2")
	gopodJSONFileSet(file, []string{"a"}, "2")
	gopodUndoEnd()
	if got := gopodSystemUndo(); !strings.Contains(got, "undid") {
		t.Errorf("undo: %s", got)
	}
	if data, _ := os.ReadFile(file); string(data) != `{"a": 1}` {
		t.Errorf("after undo: %s", data)
	}
	gopodUndoBegin("system-set x a 3")
	gopodJSONFileSet(file, []string{"a"}, "3")
	gopodUndoEnd()
	os.WriteFile(file, []byte(`{"a": 9}`), 0o644)
	if got := gopodSystemUndo(); !strings.Contains(got, "not undone") {
		t.Errorf("changed since: %s", got)
	}
	if got := gopodSystemHelp(); !strings.Contains(got, "system-undo") || !strings.Contains(got, "system-status") {
		t.Errorf("help: %s", got)
	}
	if !gopodSystemSetMatch("system-status") || !gopodSystemSetMatch("system-undo") || !gopodSystemSetMatch("system-help") {
		t.Error("match")
	}
}

// TestSystemFind (2026-10-17, operator: "system-find FEED_CHECK or system-find
// Toronto * yes, please"): the log's newest lines first, then the files.
func TestSystemFind(t *testing.T) {
	saved := gopodFindJournal
	defer func() { gopodFindJournal = saved }()
	gopodFindJournal = func() ([]byte, error) {
		return []byte("Oct 17 09:55:40 brain start.sh[1]: GOPOD_FEED 1 Toronto first\nOct 17 09:55:43 brain start.sh[1]: other\nOct 17 09:55:45 brain start.sh[1]: GOPOD_FEED 2 toronto second\n"), nil
	}
	got := gopodSystemFind("system-find Toronto")
	first, second := strings.Index(got, "second"), strings.Index(got, "first")
	if !strings.Contains(got, "2 line(s)") || first < 0 || second < 0 || first > second || strings.Contains(got, "other") {
		t.Errorf("%s", got)
	}
	if !strings.Contains(got, "09:55:45 GOPOD_FEED 2") {
		t.Errorf("log line form: %s", got)
	}
	if !gopodSystemSetMatch("system-find x") {
		t.Error("match")
	}
}

// TestSystemSearchQuery (2026-10-17, operator: "brobots use \"system-search\"
// for 'search [search-text]' like 'stat [run-name]'"): the line opens with it.
func TestSystemSearchQuery(t *testing.T) {
	for in, want := range map[string]string{
		"search Toronto Ohio population?":        "Toronto Ohio population",
		"Brobot 1, search for the tallest tower": "the tallest tower",
		"look up Hot Wheels Twin Mill":           "Hot Wheels Twin Mill",
		"I did some research today":              "",
		"can you search the room":                "",
	} {
		if got := gopodSearchQuery(in); got != want {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
}

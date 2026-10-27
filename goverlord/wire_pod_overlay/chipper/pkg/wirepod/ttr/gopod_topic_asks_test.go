// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_topic_asks_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 2026-10-26 (operator: "[topic]? = summary response * [more]? = detailed
// response * so always summarize first"; "go for \"tell me about\" vs \"tell me
// all about\""): the topic ask template.
func TestTopicAsksSummaryThenMore(t *testing.T) {
	for q, want := range map[string]string{
		"tell me about your mobile gear":     "summary",
		"mobile gear?":                       "summary",
		"tell me all about your mobile gear": "detail",
		"tell me more":                       "detail",
	} {
		if got := gopodTopicAskKind(q); got != want {
			t.Errorf("%q: %s, want %s", q, got, want)
		}
	}
	line, facts, ok := gopodTopicSummary("mobile_gear", "tell me about your gear")
	if !ok || !strings.Contains(line, "Lower Crate 1 - Power Station") || !strings.Contains(facts, "Upper Crate 2 - GOPOD Compute") {
		t.Fatalf("rig summary: %v %q %q", ok, line, facts)
	}
	if _, _, ok := gopodTopicSummary("mobile_gear", "tell me all about your gear"); ok {
		t.Error("a detail ask got the summary")
	}
	// "tell me more" names no topic: the last topic asked.
	if p := gopodTopicPackets("what's your gear like?"); len(p) == 0 {
		t.Skip("the gear question loads no topic here")
	}
	more := gopodTopicPackets("tell me more")
	found := false
	for _, p := range more {
		found = found || p.Name == "mobile_gear"
	}
	if !found {
		t.Errorf("tell me more didn't go back to the rig: %v", more)
	}
}

// 2026-10-26 (operator: "as I update the repo, I want the game to match the
// repo"): every topic's source: file is in the repo.
func TestTopicSourcesInRepo(t *testing.T) {
	if _, err := os.Stat(filepath.Join(gopodRepoRoot(), "README.md")); err != nil {
		t.Skip("not in the repo here")
	}
	for _, p := range gopodTopicSourceProblems() {
		if strings.Contains(p, "isn't in the repo") {
			t.Error(p)
		}
	}
}

// 2026-10-26: BROBOT_ASKS.md is the game's topics, as they are now
// (GOPOD_WRITE_ASKS=1 writes it again).
func TestBrobotAsksPageMatches(t *testing.T) {
	path := filepath.Join(gopodRepoRoot(), "BROBOT_ASKS.md")
	if _, err := os.Stat(filepath.Join(gopodRepoRoot(), "README.md")); err != nil {
		t.Skip("not in the repo here")
	}
	page := gopodAsksPage()
	if os.Getenv("GOPOD_WRITE_ASKS") == "1" {
		if err := os.WriteFile(path, []byte(page), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != page {
		t.Errorf("BROBOT_ASKS.md doesn't match the topics - GOPOD_WRITE_ASKS=1 writes it again")
	}
}

// 2026-10-26 (operator: "Apply pre/post template to all, even minimally"): every
// topic gets a pre note until the room has covered it, then a post note.
func TestTopicPrePost(t *testing.T) {
	gopodCoverage.Lock()
	saved := gopodCoverage.revealed
	gopodCoverage.revealed = map[string]int{}
	gopodCoverage.Unlock()
	defer func() { gopodCoverage.Lock(); gopodCoverage.revealed = saved; gopodCoverage.Unlock() }()
	if gopodTopicTouched("mobile_gear") {
		t.Fatal("covered before any point")
	}
	pre := gopodTopicPrePost("mobile_gear", false)
	if pre == "" || !strings.Contains(pre, "the rig") || strings.Contains(pre, "{topic}") {
		t.Errorf("pre: %q", pre)
	}
	gopodCoverage.Lock()
	gopodCoverage.revealed["mobile_gear"] = 1
	gopodCoverage.Unlock()
	if !gopodTopicTouched("mobile_gear") {
		t.Fatal("not covered after a point")
	}
	if post := gopodTopicPrePost("mobile_gear", true); post == "" || post == pre {
		t.Errorf("post: %q", post)
	}
	// Every chaptered topic gets one (the template's defaults at least).
	chapters, _, _ := gopodReadChaptersFolder()
	for _, c := range chapters {
		for _, tf := range c.Topics {
			if gopodTopicPrePost(tf.Name, false) == "" {
				t.Errorf("%s: no pre note", tf.Name)
			}
		}
	}
}

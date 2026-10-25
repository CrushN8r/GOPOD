package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// LOCKED 2026-10-25 (operator: "Mechanism works. Thank you. Lock it in."): his
// opener (Got it! / Nice try, Mr/Mrs ...!) - said while thinking, posted as
// "opener:" once said, the first line of his bubble above his notice, replayed
// word by word after the bubble settles (replaySaidLine,
// word_reveal.replay_settle_seconds), never counted among his reply's lit
// pieces (ownSays), never a notice line.
func TestOpenerMechanismLocked(t *testing.T) {
	raw, _ := gopodRawSettingsForTest(t)
	if wr, _ := raw["word_reveal"].(map[string]any); wr == nil || wr["replay_settle_seconds"] == nil {
		t.Fatal("word_reveal.replay_settle_seconds is missing")
	}
	for _, p := range gopodBrobotsSettings.BubbleNotices.ValuePoints.Parts {
		if p == "got_it" {
			t.Fatal("the opener is a notice line again")
		}
	}
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("GOPOD_OVERLAY_CHIPPER not set")
	}
	read := func(p string) string { b, _ := os.ReadFile(filepath.Join(root, p)); return string(b) }
	checks := map[string][]string{
		"pkg/wirepod/ttr/gopod_thinking_line.go":   {"gopodSayWords(robot, esn, gopodOpenerPick(esn))", "gopodOpenerShow(esn)"},
		"pkg/wirepod/ttr/gopod_response_chains.go": {"gopodOpenerMark+t", `const gopodOpenerMark = "opener:"`},
		"webroot/js/gopod_brobots_chat.js": {
			`indexOf("opener:") === 0) { brobotOpener(e.to, String(e.text).slice(7), done); return; }`,
			"replaySaidLine(say, done);",
			`bubble.classList.add("voice-lit");`,
			"wordReveal.replay_settle_seconds",
			`!x.closest(".nested-host, .brobot-opener")`,
		},
	}
	for file, wants := range checks {
		src := read(file)
		for _, w := range wants {
			if !strings.Contains(src, w) {
				t.Errorf("LOCKED opener mechanism: %s lacks %s", file, w)
			}
		}
	}
}

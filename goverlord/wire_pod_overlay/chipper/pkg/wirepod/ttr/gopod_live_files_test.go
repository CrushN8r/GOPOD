package wirepod_ttr

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMain points the files the chat page reads (its feed, the record-session
// and warm-up state, the speaking line) at a throwaway folder, so a test run
// never writes into the live page (2026-10-08: test lines - "Bingo Test Live",
// "Bingo Test Begins. Round 1!" - showed up in the operator's chat, and a test's
// feed looked like a Wire-Pod restart, which clears the chat).
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "gopod_ttr_test_")
	if err != nil {
		panic(err)
	}
	gopodChatFeedPath = filepath.Join(dir, "gopod_chat_feed.json")
	gopodRecordStatePath = filepath.Join(dir, "gopod_record_session.json")
	gopodLLMWarmupStatePath = filepath.Join(dir, "gopod_llm_warmup.json")
	gopodSpeakingNowPath = filepath.Join(dir, "gopod_speaking_now.json")
	// 2026-10-29: DocBS is off in the tests (its own test turns it on), so other
	// tests' answers stay the same each run.
	gopodBrobotsSettings.DocBS.Share = 0
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

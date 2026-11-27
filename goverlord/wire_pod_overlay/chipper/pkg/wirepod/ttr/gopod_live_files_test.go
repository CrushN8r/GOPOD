package wirepod_ttr

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMain points the files the chat page reads (its feed, the record-session and warm-up state,
// the speaking line) at a throwaway folder, so a test run never writes into the live page.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "gopod_ttr_test_")
	if err != nil {
		panic(err)
	}
	gopodChatFeedPath = filepath.Join(dir, "gopod_chat_feed.json")
	gopodRecordStatePath = filepath.Join(dir, "gopod_record_session.json")
	gopodLLMWarmupStatePath = filepath.Join(dir, "gopod_llm_warmup.json")
	gopodSpeakingNowPath = filepath.Join(dir, "gopod_speaking_now.json")
	gopodPrivateESNsPath = filepath.Join(dir, "gopod_esns.json")
	os.WriteFile(gopodPrivateESNsPath, []byte(`{"ESN01": "own00001", "ESN02": "own00002"}`), 0o644)
	gopodBrobotsSettings.DocBS.Share = 0
	gopodBrobotsSettings.LineQueue.On = false
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

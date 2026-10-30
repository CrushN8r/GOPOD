// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_record_session_test.go
package wirepod_ttr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGopodRecordAnswerKind(t *testing.T) {
	for text, want := range map[string]string{
		"yes keep the session recording":   "keep",
		"Keep session recording":           "keep",
		"recording session, keep it":       "keep",
		"no delete the session recording":  "delete",
		"delete session recording":         "delete",
		"yes":                              "",
		"keep the session":                 "",
		"delete the recording":             "",
		"keep or delete session recording": "",
	} {
		if got := gopodRecordAnswerKind(text); got != want {
			t.Errorf("gopodRecordAnswerKind(%q) = %q, want %q", text, got, want)
		}
	}
}

// gopodRecordTestSetup points every Record Session path at a temp folder; the
// room mic script is a missing file, so no real recording starts.
func gopodRecordTestSetup(t *testing.T) string {
	dir := t.TempDir()
	root := filepath.Join(dir, "sessions")
	t.Setenv("GOPOD_RECORD_SESSION_DIR", root)
	t.Setenv("GOPOD_TAKE_SHEET", filepath.Join(dir, "take_sheet.json"))
	oldState, oldFeed, oldSettings := gopodRecordStatePath, gopodChatFeedPath, gopodBrobotsSettings
	gopodRecordStatePath = filepath.Join(dir, "gopod_record_session.json")
	gopodChatFeedPath = filepath.Join(dir, "gopod_chat_feed.json")
	gopodBrobotsSettings.MicService.Script = "no_such_mic_service.py"
	t.Cleanup(func() {
		gopodRecordStatePath, gopodChatFeedPath, gopodBrobotsSettings = oldState, oldFeed, oldSettings
		gopodRecord.waiting, gopodRecord.wanted, gopodRecord.answeredAt = false, false, time.Time{}
	})
	GopodChatFeedStart()
	return root
}

func gopodRecordWaitFor(t *testing.T, what string, ok func() bool) {
	deadline := time.Now().Add(5 * time.Second)
	for !ok() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func gopodRecordIs(field func() bool) func() bool {
	return func() bool {
		gopodRecord.Lock()
		defer gopodRecord.Unlock()
		return field()
	}
}

func gopodRecordFolders(t *testing.T, root, pattern string) []string {
	found, _ := filepath.Glob(filepath.Join(root, "*", pattern))
	return found
}

// Thanks -> question -> "keep session recording": the session is zipped, the
// loose folder is gone, and the chat lines were recorded.
func TestGopodRecordSessionThanksKeep(t *testing.T) {
	root := gopodRecordTestSetup(t)
	GopodRecordSessionSet(true)
	gopodRecordWaitFor(t, "recording", gopodRecordIs(func() bool { return gopodRecord.active }))
	GopodChatFeedAddTo("Operator", "Brobot 1", "what is gopod")
	gopodRecordExchange(gopodBrobot1ESN, "what is gopod", gopodStreamResult{Raw: "GOPOD is a stage", FilteredPacket: "GOPOD is a stage"})

	GopodRecordSessionThanks()
	if !gopodRecordIs(func() bool { return gopodRecord.waiting })() {
		t.Fatal("after thanks the keep-or-delete question should be open")
	}
	if GopodRecordSessionAnswer("tell me a joke", "Typed text") {
		t.Fatal("a chat line is not an answer")
	}
	if !GopodRecordSessionAnswer("yes keep the session recording", "Typed text") {
		t.Fatal("the keep line should be taken")
	}
	if !GopodRecordSessionAnswer("yes keep the session recording", "Typed text") {
		t.Fatal("the brother's hearing of the same answer should be taken quietly")
	}
	gopodRecordWaitFor(t, "zip", func() bool { return len(gopodRecordFolders(t, root, "*.zip")) == 1 })
	// The zip is written before the loose folder is removed: wait for both.
	gopodRecordWaitFor(t, "loose folder removed", func() bool {
		return len(gopodRecordFolders(t, root, "*"+gopodBrobotsSettings.RecordSession.FolderSuffix)) == 0
	})
	if left := gopodRecordFolders(t, root, "*"+gopodBrobotsSettings.RecordSession.FolderSuffix); len(left) != 0 {
		t.Fatalf("loose folder left after keep: %v", left)
	}
	// "kept" is said just after the zip: wait for it.
	gopodRecordWaitFor(t, "kept said", func() bool {
		feed, _ := os.ReadFile(gopodChatFeedPath)
		return strings.Contains(string(feed), gopodBrobotsSettings.RecordSession.KeptText)
	})
	feed, _ := os.ReadFile(gopodChatFeedPath)
	// The question is a Brobot bubble, cut once into its pieces (gopodRecordSay).
	cut := func(text string) string {
		return strings.Join(gopodSentences(text), gopodBrobotsSettings.ScriptedSplit.SegmentMarker)
	}
	for _, want := range []string{cut(gopodBrobotsSettings.RecordSession.AskText), cut(gopodBrobotsSettings.RecordSession.KeptText), `"to":"Goverlord"`} {
		if !strings.Contains(string(feed), want) {
			t.Errorf("feed is missing %q", want)
		}
	}
	// 2026-10-06: keep also adds the session's answers to the take sheet.
	sheetPath, _, _ := gopodTakeSheetPath()
	var sheet gopodTakeSheet
	data, err := os.ReadFile(sheetPath)
	if err != nil || json.Unmarshal(data, &sheet) != nil {
		t.Fatalf("take sheet not written: %v", err)
	}
	found := false
	for _, p := range sheet.Points {
		for _, take := range p.Takes {
			if take.Ask == "what is gopod" && take.Input == "typed" && take.Say == "GOPOD is a stage" && take.To == "Brobot 1" {
				found = true
			}
		}
	}
	if !found || len(sheet.Runs) != 1 || !strings.HasSuffix(sheet.Runs[0], gopodBrobotsSettings.RecordSession.FolderSuffix) {
		t.Errorf("kept session's take not on the sheet: runs=%v", sheet.Runs)
	}
	if _, err := os.Stat(strings.TrimSuffix(sheetPath, ".json") + ".md"); err != nil {
		t.Errorf("take_sheet.md not drawn: %v", err)
	}
}

func TestGopodRecordSessionThanksDelete(t *testing.T) {
	root := gopodRecordTestSetup(t)
	GopodRecordSessionSet(true)
	gopodRecordWaitFor(t, "recording", gopodRecordIs(func() bool { return gopodRecord.active }))
	GopodRecordSessionThanks()
	if !GopodRecordSessionAnswer("no delete the session recording", "Typed text") {
		t.Fatal("the delete line should be taken")
	}
	gopodRecordWaitFor(t, "delete", func() bool {
		return len(gopodRecordFolders(t, root, "*")) == 0
	})
	// Delete adds nothing to the take sheet.
	if sheetPath, _, _ := gopodTakeSheetPath(); func() bool { _, err := os.Stat(sheetPath); return err == nil }() {
		t.Error("a deleted session added takes")
	}
}

// The record has its start and stop lines, a typed question once (when typed),
// and a spoken question with its answer.
func TestGopodRecordSessionLines(t *testing.T) {
	gopodRecordTestSetup(t)
	GopodRecordSessionSet(true)
	gopodRecordWaitFor(t, "recording", gopodRecordIs(func() bool { return gopodRecord.active }))
	GopodChatFeedAddTo("Operator", "Brobots", "Who all is here?")
	gopodRecordExchange(gopodBrobot2ESN, "Who all is here?", gopodStreamResult{Raw: "me and my bro", FilteredPacket: "me and my bro"})
	gopodRecordExchange(gopodBrobot1ESN, "tell me a joke", gopodStreamResult{Raw: "no", FilteredPacket: "no"})
	dir := gopodRecordStop("test")
	data, err := os.ReadFile(filepath.Join(dir, "chat.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if n := strings.Count(text, "Who all is here?"); n != 1 {
		t.Errorf("typed question recorded %d times, want 1:\n%s", n, text)
	}
	for _, want := range []string{`"text":"started"`, `"text":"stopped (test`, `"text":"tell me a joke"`, `"text":"me and my bro"`} {
		if !strings.Contains(text, want) {
			t.Errorf("record is missing %s:\n%s", want, text)
		}
	}
	log, _ := os.ReadFile(filepath.Join(dir, "wirepod.log"))
	if !strings.Contains(string(log), "GOPOD_RECORD_SESSION stopped (test") {
		t.Errorf("wirepod.log is missing the stop line")
	}
	os.RemoveAll(dir)
}

// Unticking stops with no question: zipped and kept.
func TestGopodRecordSessionUntickKeeps(t *testing.T) {
	root := gopodRecordTestSetup(t)
	GopodRecordSessionSet(true)
	gopodRecordWaitFor(t, "recording", gopodRecordIs(func() bool { return gopodRecord.active }))
	GopodChatFeedAdd("Brobot 2", "party on")
	GopodRecordSessionSet(false)
	gopodRecordWaitFor(t, "zip", func() bool { return len(gopodRecordFolders(t, root, "*.zip")) == 1 })
	if gopodRecordIs(func() bool { return gopodRecord.waiting })() {
		t.Fatal("unticking should not ask")
	}
}

// Recording starts only between replies.
func TestGopodRecordSessionWaitsForSafePoint(t *testing.T) {
	gopodRecordTestSetup(t)
	gopodReplyStart(gopodBrobot2ESN)
	GopodRecordSessionSet(true)
	time.Sleep(400 * time.Millisecond)
	if gopodRecordIs(func() bool { return gopodRecord.active })() {
		t.Fatal("recording started during a reply")
	}
	gopodReplyEnd(gopodBrobot2ESN)
	gopodRecordWaitFor(t, "recording", gopodRecordIs(func() bool { return gopodRecord.active }))
	GopodRecordSessionSet(false)
	gopodRecordWaitFor(t, "stopped", gopodRecordIs(func() bool { return !gopodRecord.active }))
}

// At Wire-Pod start: an unanswered folder is deleted, a cut-off one zipped.
func TestGopodRecordCleanupLeftovers(t *testing.T) {
	root := gopodRecordTestSetup(t)
	suffix := gopodBrobotsSettings.RecordSession.FolderSuffix
	pending := filepath.Join(root, "2026-10-04", "100000"+suffix)
	cut := filepath.Join(root, "2026-10-04", "110000"+suffix)
	for dir, marker := range map[string]string{pending: gopodRecordMarkerPending, cut: gopodRecordMarkerRecording} {
		os.MkdirAll(dir, 0o755)
		os.WriteFile(filepath.Join(dir, marker), nil, 0o644)
		os.WriteFile(filepath.Join(dir, "chat.jsonl"), []byte(`{"time_ms":1,"who":"Operator","text":"hi"}`+"\n"), 0o644)
	}
	gopodRecordCleanupLeftovers()
	if _, err := os.Stat(pending); err == nil {
		t.Error("unanswered folder should be deleted")
	}
	if _, err := os.Stat(cut + ".zip"); err != nil {
		t.Error("cut-off recording should be zipped")
	}
}

// ## EOF

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_llm_warmup_test.go
package wirepod_ttr

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestGopodOllamaBase(t *testing.T) {
	for in, want := range map[string]string{
		"http://host:11434/v1":  "http://host:11434",
		"http://host:11434/v1/": "http://host:11434",
		"http://host:11434":     "http://host:11434",
	} {
		if got := gopodOllamaBase(in); got != want {
			t.Errorf("gopodOllamaBase(%q) = %q, want %q", in, got, want)
		}
	}
}

// The warm-up loads the chat's own model on the chat's own endpoint, then sets
// keep_alive from the settings file.
func TestGopodLLMWarmupRequests(t *testing.T) {
	var mu sync.Mutex
	var paths []string
	var bodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(data, &body)
		mu.Lock()
		paths = append(paths, r.URL.Path)
		bodies = append(bodies, body)
		mu.Unlock()
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	gopodLLMWarmup(server.URL, "gemma2:2b")

	if len(paths) != 2 || paths[0] != "/v1/chat/completions" || paths[1] != "/api/generate" {
		t.Fatalf("requests = %v, want the chat warm request then the keep-loaded one", paths)
	}
	for i, body := range bodies {
		if body["model"] != "gemma2:2b" {
			t.Errorf("request %d model = %v", i, body["model"])
		}
		if _, ok := body["options"]; ok {
			t.Errorf("request %d sends options; the chat sends none, so the load would differ", i)
		}
	}
	if bodies[1]["keep_alive"] != gopodBrobotsSettings.LLM.KeepAlive {
		t.Errorf("keep_alive = %v, want %q", bodies[1]["keep_alive"], gopodBrobotsSettings.LLM.KeepAlive)
	}
	if _, ok := bodies[1]["prompt"]; ok {
		t.Errorf("the keep-loaded request must not hold a prompt (it would generate)")
	}
}

// The page's warm-up bubble comes from its own file, never from the chat feed,
// so the Brobots' chat memory never sees it.
func TestGopodLLMWarmupBubbleSkippedByMemory(t *testing.T) {
	dir := t.TempDir()
	oldState, oldFeed := gopodLLMWarmupStatePath, gopodChatFeedPath
	gopodLLMWarmupStatePath = filepath.Join(dir, "gopod_llm_warmup.json")
	gopodChatFeedPath = filepath.Join(dir, "gopod_chat_feed.json")
	defer func() { gopodLLMWarmupStatePath, gopodChatFeedPath = oldState, oldFeed }()
	GopodChatFeedStart()

	gopodWriteLLMWarmupState(true)
	data, err := os.ReadFile(gopodLLMWarmupStatePath)
	if err != nil {
		t.Fatal(err)
	}
	var state struct {
		Active    bool  `json:"active"`
		StartedMS int64 `json:"started_ms"`
		DoneMS    int64 `json:"done_ms"`
	}
	_ = json.Unmarshal(data, &state)
	if !state.Active || state.StartedMS == 0 || state.DoneMS != 0 {
		t.Fatalf("warm-up state = %+v, want active with a start time", state)
	}
	for _, m := range gopodSharedBrobotChat() {
		if strings.Contains(m.Content, gopodBrobotsSettings.LLM.WarmupText) {
			t.Fatalf("the warm-up text reached the chat memory: %q", m.Content)
		}
	}
	feed, _ := os.ReadFile(gopodChatFeedPath)
	if strings.Contains(string(feed), gopodBrobotsSettings.LLM.WarmupText) {
		t.Fatalf("the warm-up text reached the chat feed")
	}

	gopodWriteLLMWarmupState(false)
	data, _ = os.ReadFile(gopodLLMWarmupStatePath)
	_ = json.Unmarshal(data, &state)
	if state.Active || state.DoneMS == 0 || state.StartedMS == 0 {
		t.Fatalf("after the warm-up the state = %+v, want done with both times", state)
	}
}

// One bubble (operator 2026-10-09: "Combine into one bubble"), a line each
// (2026-10-16: "LLM Ready, Master CrushN8r." / "Standing by for another crushing
// performance." / "Your move.").
func TestWarmupReadyLineIsOneBubble(t *testing.T) {
	if got := gopodWarmupReadyLine("LLM Ready, Master CrushN8r.", "Standing by for another crushing performance.\nYour move."); got != "LLM Ready, Master CrushN8r.\nStanding by for another crushing performance.\nYour move." {
		t.Fatalf("got %q", got)
	}
	if got := gopodWarmupReadyLine("LLM Ready", ""); got != "LLM Ready" {
		t.Fatalf("got %q", got)
	}
}

// ## EOF

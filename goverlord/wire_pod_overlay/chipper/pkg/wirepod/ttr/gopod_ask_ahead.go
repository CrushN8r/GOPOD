// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_ask_ahead.go
package wirepod_ttr

// Ask ahead (2026-10-29, operator: "Maybe get gemma to help with guiding the flow
// with small background asks from the code in between steps? Hey gemma how would
// you colour this part?"; "re: each small background ask ... runs while the
// previous bubble is being spoken * yes, please build"). A small ask to the
// language model starts the moment its question is known - while the line is
// read out and the bubbles before it play - and its answer waits here until the
// answer chain needs it: speaking time is thinking time. The model's own queue
// keeps them in order with the main answer.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

type gopodAhead struct {
	done chan struct{}
	text string
}

var gopodAheads sync.Map // key -> *gopodAhead

// gopodAskAhead starts a small ask (once per key); its answer is kept for gopodAskAheadGet.
func gopodAskAhead(key, prompt string, numPredict int) {
	if strings.TrimSpace(prompt) == "" {
		return
	}
	a := &gopodAhead{done: make(chan struct{})}
	if _, loaded := gopodAheads.LoadOrStore(key, a); loaded {
		return
	}
	go func() {
		defer close(a.done)
		t0 := time.Now()
		a.text = gopodOllamaAsk(prompt, numPredict, 0.9)
		logger.Println("GOPOD_ASK_AHEAD " + key + " (" + time.Since(t0).Round(100*time.Millisecond).String() + "): " + gopodCut(strings.ReplaceAll(a.text, "\n", " / "), 120))
	}()
	// Unclaimed answers don't stay forever.
	time.AfterFunc(5*time.Minute, func() { gopodAheads.Delete(key) })
}

// gopodAskAheadGet: the ask's answer, waiting for it up to wait; "" when there's
// none (never asked, or not back in time).
func gopodAskAheadGet(key string, wait time.Duration) string {
	v, ok := gopodAheads.Load(key)
	if !ok {
		return ""
	}
	a := v.(*gopodAhead)
	select {
	case <-a.done:
		gopodAheads.Delete(key)
		return a.text
	case <-time.After(wait):
		logger.Println("GOPOD_ASK_AHEAD " + key + ": not back in time")
		return ""
	}
}

// gopodOllamaAsk: one short ask to the language model, "" on any failure (a
// variable, so a test can stand in for the model).
var gopodOllamaAsk = func(prompt string, numPredict int, temperature float64) string {
	base, model, ok := gopodWarmupOllama()
	if !ok || model == "" {
		return ""
	}
	if numPredict <= 0 {
		numPredict = 60
	}
	body, _ := json.Marshal(map[string]any{
		"model": model, "stream": false,
		"messages": []map[string]string{{"role": "user", "content": prompt}},
		"options":  map[string]any{"num_predict": numPredict, "temperature": temperature},
	})
	resp, err := (&http.Client{Timeout: gopodTiming("ask_ahead_seconds")}).Post(base+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var out struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if json.NewDecoder(resp.Body).Decode(&out) != nil {
		return ""
	}
	return strings.TrimSpace(out.Message.Content)
}

// ## EOF

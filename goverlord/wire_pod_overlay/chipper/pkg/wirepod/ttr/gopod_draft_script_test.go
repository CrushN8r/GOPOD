// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_draft_script_test.go
package wirepod_ttr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// A draft written as a script reads back the same.
func TestDraftScriptRoundTrip(t *testing.T) {
	d := gopodDraft{Run: "test", Questions: []gopodDraftQuestion{
		{N: 1, Topic: "1. Who they are", To: "Brobot 1", ReadBy: "Reporter 2: Brobot 4", Ask: "Hey party animal. Who are you?",
			Before: []gopodDraftLine{{Who: "Reporter 1: Brobot 3", Text: "Here we go."}},
			Answers: []gopodDraftAnswer{
				{Brobot: "Brobot 1", Cue: "happy", Say: "Brobot 1. I hear you: {weather}",
					Moves: []gopodDraftMove{{At: "lift my arm", Move: "arm"}, {At: "nod", Move: "nod"}}},
				{Brobot: "Reporter 1: Brobot 3", Say: "Did you see the arm?"},
				{Brobot: "Brobot 2", Say: "Rude. Accurate."},
			}},
		{N: 2, Topic: "2. Both", To: "Brobots", ReadBy: "Reporter 1: Brobot 3", Ask: "Thanks Brobots",
			Answers: []gopodDraftAnswer{{Brobot: "Brobot 2", Say: "Catch ya later."}}},
	}}
	text := gopodDraftScriptText(d)
	got, err := gopodParseDraftScript(text)
	if err != nil {
		t.Fatal(err, "\n", text)
	}
	if !reflect.DeepEqual(got, d) {
		t.Errorf("round trip:\n%#v\nwant\n%#v\n%s", got, d, text)
	}
	if _, err := gopodParseDraftScript("## 1.\nBrobot 1: before the question\n"); err == nil {
		t.Error("a Brobot line before the question should fail")
	}
}

// GOPOD_WRITE_DRAFT_SCRIPT=<draft_performance.json> writes its draft_script.md
// beside it (an existing one is left alone).
func TestWriteDraftScript(t *testing.T) {
	path := os.Getenv("GOPOD_WRITE_DRAFT_SCRIPT")
	if path == "" {
		t.Skip("GOPOD_WRITE_DRAFT_SCRIPT not set")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var d gopodDraft
	if err := json.Unmarshal(data, &d); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(filepath.Dir(path), gopodDraftScriptName)
	if _, err := os.Stat(out); err == nil {
		t.Fatal(out + " already exists")
	}
	if err := os.WriteFile(out, []byte(gopodDraftScriptText(d)), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Log("wrote " + out)
}

// Live replies say Brobot, not robot (reply_tighten.brand_words).
func TestReplyBrandWords(t *testing.T) {
	got := gopodTightenReply("happy___ Robots rock! A robot party, robot-hosted, for Brobots.", "hi", gopodBrobot2ESN)
	want := "happy___ Brobots rock! A Brobot party, Brobot-hosted, for Brobots."
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

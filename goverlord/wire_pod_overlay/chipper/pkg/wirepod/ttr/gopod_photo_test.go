// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_photo_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPhotoKind: the photo egg's questions, in the live chat only (operator
// 2026-10-14: "ensure live pop-ups are not triggered by a song run").
func TestPhotoKind(t *testing.T) {
	if len(gopodBrobotsSettings.Photo.AskPhrases) == 0 {
		t.Skip("no photo settings")
	}
	defer func() {
		gopodSession.Lock()
		gopodSession.run = nil
		gopodSession.Unlock()
	}()
	cases := map[string]string{
		"Brobot 2, what's in front of you?": "ask",
		"what do you see":                   "ask",
		"tell me what you saw":              "recall",
		"what did you see?":                 "recall",
		"what's the weather in Windsor":     "",
	}
	for line, want := range cases {
		if got := gopodPhotoKind(line); got != want {
			t.Errorf("gopodPhotoKind(%q) = %q, want %q", line, got, want)
		}
	}
	gopodSession.Lock()
	gopodSession.run = &gopodRun{Name: "Fart Joke", Kind: "song"}
	gopodSession.Unlock()
	if got := gopodPhotoKind("what's in front of you?"); got != "" {
		t.Errorf("during a run, gopodPhotoKind = %q, want none", got)
	}
}

// TestPhotoRecall: "tell me what you saw" is answered from what he saw last.
func TestPhotoRecall(t *testing.T) {
	p := gopodBrobotsSettings.Photo
	if p.RecallText == "" {
		t.Skip("no photo settings")
	}
	esn := "test-esn"
	defer func() {
		gopodPhotoLast.Lock()
		delete(gopodPhotoLast.saw, esn)
		gopodPhotoLast.Unlock()
	}()
	if got := gopodPhotoRecall(esn); got != p.NothingText {
		t.Errorf("before a photo: %q, want %q", got, p.NothingText)
	}
	gopodPhotoLast.Lock()
	gopodPhotoLast.saw[esn] = "A table, a chair and some book shelves."
	gopodPhotoLast.Unlock()
	if got := gopodPhotoRecall(esn); !strings.Contains(got, "a table, a chair and some book shelves") {
		t.Errorf("after a photo: %q", got)
	}
}

// TestPhotoRounds: round 1 names the Edge TPU's objects; round 2 builds up from
// it with what the vision model saw (operator 2026-10-16: "do this in 2 rounds").
func TestPhotoRounds(t *testing.T) {
	p := gopodBrobotsSettings.Photo
	if p.DetailText == "" {
		t.Skip("no photo settings")
	}
	if got := gopodQuickLookList([]string{"person", "orange", "chair"}); got != "a person, an orange and a chair" {
		t.Errorf("list: %q", got)
	}
	if got := gopodQuickLookList([]string{"keyboard"}); got != "a keyboard" {
		t.Errorf("list of one: %q", got)
	}
	if got := gopodPhotoDetail("The image shows a person with glasses at a desk.", []string{"person"}); got != "I see a person with glasses at a desk." {
		t.Errorf("statement 2: %q", got)
	}
	if got := gopodPhotoDetail("A person.", []string{"person"}); got != p.DetailUnclearText {
		t.Errorf("statement 2 repeats statement 1: %q", got)
	}
	if got := gopodPhotoDetail("Urn on the table next to the man", nil); got != "I see urn on the table next to the man." {
		t.Errorf("statement 2, no objects: %q", got)
	}
	if got := gopodPhotoDetail("", []string{"person"}); got != p.DetailUnclearText {
		t.Errorf("nothing seen: %q", got)
	}
}

// TestPhotoDescribeLive (GOPOD_PHOTO_LIVE=1): the vision model describes a
// picture from the chat's pictures folder, through this machine's Ollama.
func TestPhotoDescribeLive(t *testing.T) {
	if os.Getenv("GOPOD_PHOTO_LIVE") != "1" {
		t.Skip("set GOPOD_PHOTO_LIVE=1 to ask the vision model")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(gopodPhotoDir), "pictures", "gear/rig_brobots_in_harness.jpg"))
	if err != nil {
		t.Skip("no test picture: " + err.Error())
	}
	saved := gopodWarmupOllama
	defer func() { gopodWarmupOllama = saved }()
	gopodWarmupOllama = func() (string, string, bool) { return "http://localhost:11434", "", true }
	saw, err := gopodPhotoDescribe(data)
	if err != nil || saw == "" {
		t.Fatalf("describe: %q %v", saw, err)
	}
	t.Logf("saw: %s", saw)
}

// TestQuickLookLive (GOPOD_QUICK_LOOK_LIVE=1, Wire-Pod stopped - the Edge TPU
// runs one program at a time): the worker names the objects in a kept photo.
func TestQuickLookLive(t *testing.T) {
	if os.Getenv("GOPOD_QUICK_LOOK_LIVE") != "1" {
		t.Skip("set GOPOD_QUICK_LOOK_LIVE=1 to use the Edge TPU")
	}
	photos, _ := filepath.Glob(filepath.Join(gopodPhotoDir, "*.jpg"))
	if len(photos) == 0 {
		t.Skip("no kept photo")
	}
	gopodQuickLookStart()
	objects, err := gopodQuickLookObjects(photos[0])
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s: %v", filepath.Base(photos[0]), objects)
}

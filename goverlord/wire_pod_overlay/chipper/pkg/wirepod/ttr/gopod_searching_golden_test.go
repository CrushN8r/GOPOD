package wirepod_ttr

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// LOCKED 2026-11-03 (operator: "searching animation is breaking again. please
// isolate into a robust golden note and fucking lock it and refer only to it";
// "Fail the deploy"): the deploy fails when the golden note
// (gopod_probes/locks/searching_golden.md) is missing or broken, when it changed
// without the operator's relock, or when any other code sends the searching
// clip. Changing the note is the operator's word only; then GOPOD_RELOCK=1
// saves its new fingerprint (gopod_probes/locks/searching_golden.lock).
func TestSearchingGoldenLock(t *testing.T) {
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("GOPOD_OVERLAY_CHIPPER not set (the deploy sets it)")
	}
	note, err := os.ReadFile(gopodSearchingGoldenPath())
	if err != nil {
		t.Fatalf("LOCKED: the searching golden note is missing: %v", err)
	}
	if !gopodSearchingGolden().OK {
		t.Fatalf("LOCKED: the searching golden note is broken (clips or pace missing)")
	}
	sum := sha256.Sum256(note)
	got := hex.EncodeToString(sum[:])
	lock := filepath.Join(root, "gopod_probes", "locks", "searching_golden.lock")
	if os.Getenv("GOPOD_RELOCK") == "1" {
		if err := os.WriteFile(lock, []byte(got+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	} else if want, err := os.ReadFile(lock); err != nil {
		t.Errorf("LOCKED: the searching lock is missing: %v", err)
	} else if strings.TrimSpace(string(want)) != got {
		t.Errorf("LOCKED: the searching golden note changed (operator's word only, then GOPOD_RELOCK=1)")
	}
	// The note is the one source: no other code names the searching clip.
	files, _ := filepath.Glob(filepath.Join(root, "pkg", "wirepod", "*", "*.go"))
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, _ := os.ReadFile(f)
		if strings.Contains(string(src), "anim_knowledgegraph_searching_01") {
			t.Errorf("LOCKED: %s sends the searching clip itself - only the golden note names it", filepath.Base(f))
		}
	}
	src, _ := os.ReadFile(filepath.Join(root, "pkg", "wirepod", "ttr", "gopod_thinking_line.go"))
	if !strings.Contains(string(src), "gopodSearchingGolden()") {
		t.Error("LOCKED: gopodThinkingProcess doesn't read the golden note")
	}
}

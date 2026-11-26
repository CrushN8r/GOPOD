package wirepod_ttr

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// LOCKED 2026-11-25 (operator: "a locked card, laid out like weather_run.md,
// holding that closing sequence and its QRs in order" - "yes"): the GOPOD
// Yourself QR sequence is gopod_yourself_qr.md and only that. The deploy fails
// when it changed without the operator's relock (GOPOD_RELOCK=1 saves its
// fingerprint, gopod_probes/locks/gopod_yourself_golden.lock), when its golden
// lines aren't the ones in use, or when a QR its lines call is missing.
func TestGopodYourselfGoldenLock(t *testing.T) {
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("GOPOD_OVERLAY_CHIPPER not set (the deploy sets it)")
	}
	card, err := os.ReadFile(gopodTopicTreeFind("gopod_yourself_qr.md"))
	if err != nil {
		t.Fatalf("LOCKED: the GOPOD Yourself QR card is missing: %v", err)
	}
	sum := sha256.Sum256(card)
	got := hex.EncodeToString(sum[:])
	lock := filepath.Join(root, "gopod_probes", "locks", "gopod_yourself_golden.lock")
	if os.Getenv("GOPOD_RELOCK") == "1" {
		if err := os.WriteFile(lock, []byte(got+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	} else if want, err := os.ReadFile(lock); err != nil {
		t.Errorf("LOCKED: the GOPOD Yourself lock is missing: %v", err)
	} else if strings.TrimSpace(string(want)) != got {
		t.Errorf("LOCKED: the GOPOD Yourself QR card changed (operator's word only, then GOPOD_RELOCK=1)")
	}
	gy := gopodBrobotsSettings.GopodYourselfQr
	for _, l := range []string{gy.KeepUpLine, gy.ContactLine} {
		if l == "" || !strings.Contains(string(card), l) {
			t.Errorf("LOCKED: a golden line in use isn't the card's: %q", l)
		}
	}
	// Each QR the sequence calls is up for its line's words.
	data, _ := gopodSettingsMerged()
	var page struct {
		QrCodes struct {
			Codes []struct {
				Words []string `json:"words"`
				Title string   `json:"title"`
			} `json:"codes"`
		} `json:"qr_codes"`
	}
	_ = json.Unmarshal(data, &page)
	for _, want := range []struct{ line, word string }{{gy.KeepUpLine, "newsletter"}, {gy.KeepUpLine, "crushn8r.ca"}, {gy.ContactLine, "contact@crushn8r.ca"}} {
		found := false
		for _, c := range page.QrCodes.Codes {
			for _, w := range c.Words {
				if w == want.word && strings.Contains(strings.ToLower(want.line), w) {
					found = true
				}
			}
		}
		if !found {
			t.Errorf("LOCKED: no QR for %q in %q", want.word, want.line)
		}
	}
}

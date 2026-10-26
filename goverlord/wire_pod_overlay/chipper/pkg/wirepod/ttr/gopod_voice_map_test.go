// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_voice_map_test.go
package wirepod_ttr

import (
	"os"
	"testing"
)

// 2026-10-26 (operator: "KP0 (All), KP1 and KP2: \"no voice required\" ... KP3
// to KP9: KPN-TEMPLATE ... KP7 to KP9, Operator: KPN-TEMPLATE"): robot-bodied
// keys need no voice; the rest have a persona host and its voice; the Operator last.
func TestVoiceMapRows(t *testing.T) {
	if _, err := os.Stat(gopodNumpadMapPath()); err != nil {
		t.Skip("no golden map on this machine")
	}
	rows := gopodVoiceRows()
	if len(rows) != 11 || rows[10].Key != "Operator" {
		t.Fatalf("want KP0-KP9 then Operator, got %d rows", len(rows))
	}
	for i, r := range rows[:10] {
		bodied := i <= 2
		if r.Bodied != bodied {
			t.Errorf("%s bodied=%v, want %v", r.Key, r.Bodied, bodied)
		}
		if !bodied && (r.Host == "" || r.Voice == "") {
			t.Errorf("%s (%s): host %q voice %q", r.Key, r.Persona, r.Host, r.Voice)
		}
	}
	if _, err := GopodVoiceMapSet("KP1", "af_sky"); err == nil {
		t.Error("KP1 is robot-bodied: no voice to set")
	}
	if _, err := GopodVoiceMapSet("KP7", "no_such_voice"); err == nil {
		t.Error("an unknown voice was taken")
	}
}

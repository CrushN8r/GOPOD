// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_numpad_ptt_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"testing"
)

// The live NumLock-on keys of the golden persona map talk to their Brobot; an
// assigned key does not.
func TestGopodNumpadBrobot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "map.json")
	os.WriteFile(path, []byte(`{"numlock_on": {
		"KP1": {"persona": "Doc", "body": "Vector 1", "status": "live"},
		"KP2": {"persona": "Pip", "body": "Vector 2", "status": "live"},
		"KP3": {"persona": "Git Repo", "body": "Cozmo", "status": "assigned"}}}`), 0o644)
	old := gopodBrobotsSettings
	gopodBrobotsSettings.NumpadPTT.MapFile = path
	defer func() { gopodBrobotsSettings = old }()
	for key, want := range map[string]string{"KP1": gopodBrobot1ESN, "KP2": gopodBrobot2ESN} {
		if _, esn, err := gopodNumpadBrobot(key); err != nil || esn != want {
			t.Errorf("%s -> %q, %v; want %q", key, esn, err, want)
		}
	}
	if _, _, err := gopodNumpadBrobot("KP3"); err == nil {
		t.Error("KP3 (assigned, not live) should not talk")
	}
	if data, err := GopodNumpadMapJSON(); err != nil || len(data) == 0 {
		t.Errorf("map for the page: %v", err)
	}
}

// The real golden map file, as the page will get it.
func TestGopodNumpadRealMap(t *testing.T) {
	if _, err := os.Stat(gopodNumpadMapPath()); err != nil {
		t.Skip("no golden map on this machine")
	}
	if _, esn, err := gopodNumpadBrobot("KP1"); err != nil || esn != gopodBrobot1ESN {
		t.Fatalf("KP1 -> %q, %v", esn, err)
	}
}

// ## EOF

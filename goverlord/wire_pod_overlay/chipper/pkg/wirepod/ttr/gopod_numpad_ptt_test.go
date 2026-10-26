// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_numpad_ptt_test.go
package wirepod_ttr

import (
	"encoding/json"
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

// 2026-10-26 (operator: "\"1 robust key detection process\" = Template/unify so
// everything goes through 1 place"; "if a KPN is voice assigned, make it chat
// ready"): the page's targets come from gopodNumpadBrobot - every live key one.
func TestGopodNumpadTargetsOnePlace(t *testing.T) {
	if _, err := os.Stat(gopodNumpadMapPath()); err != nil {
		t.Skip("no golden map on this machine")
	}
	data, err := GopodNumpadMapJSON()
	if err != nil {
		t.Fatal(err)
	}
	var out struct {
		Targets map[string]string `json:"targets"`
	}
	if json.Unmarshal(data, &out) != nil {
		t.Fatal("no targets for the page")
	}
	for key, want := range map[string]string{"KP0": "All", "KP1": "Brobot 1", "KP2": "Brobot 2", "KP3": "GIT REPO", "KP4": "CACHE PYC",
		"KP5": "INTEL 404 PNFH", "KP6": "INTEL 404 PNFT", "KP7": "CHALK", "KP8": "PLAYHEAD", "KP9": "CAMEO"} {
		if got := out.Targets[key]; got != want {
			t.Errorf("%s -> %q, want %q", key, got, want)
		}
		if gopodReporterVoice(want) == "" && key != "KP0" && key != "KP1" && key != "KP2" {
			t.Errorf("%s (%s) has no voice", key, want)
		}
	}
}

// ## EOF

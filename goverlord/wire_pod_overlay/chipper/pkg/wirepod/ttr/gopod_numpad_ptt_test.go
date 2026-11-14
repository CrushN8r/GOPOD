// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_numpad_ptt_test.go
package wirepod_ttr

import (
	"encoding/json"
	"strings"
	"testing"
)

// The live NumLock-on keys of the golden persona map talk to their Brobot; an
// assigned key does not.
func TestGopodNumpadBrobot(t *testing.T) {
	old := gopodBrobotsSettings
	// 2026-11-04: a test map in the settings' numpad_map (numpad-mapper.md's place).
	gopodBrobotsSettings.NumpadMaps = map[string]json.RawMessage{"usb": json.RawMessage(`{"numlock_on": {
		"KP1": {"persona": "Doc", "body": "Vector 1", "status": "live"},
		"KP2": {"persona": "Pip", "body": "Vector 2", "status": "live"},
		"KP3": {"persona": "Git Repo", "body": "Cozmo", "status": "assigned"}}}`)}
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
	// 2026-11-04: the map is clusters/gopod-system/numpad-mapper.md (settings numpad_map).
	if len(gopodBrobotsSettings.NumpadMaps[gopodNumpadLayout()]) <= 2 {
		t.Fatal("no numpad_map in the settings (clusters/gopod-system/numpad-mapper.md)")
	}
	if _, esn, err := gopodNumpadBrobot("KP1"); err != nil || esn != gopodBrobot1ESN {
		t.Fatalf("KP1 -> %q, %v", esn, err)
	}
}

// 2026-10-26 (operator: "\"1 robust key detection process\" = Template/unify so
// everything goes through 1 place"; "if a KPN is voice assigned, make it chat
// ready"): the page's targets come from gopodNumpadBrobot - every live key one.
func TestGopodNumpadTargetsOnePlace(t *testing.T) {
	// 2026-11-04: the map is clusters/gopod-system/numpad-mapper.md (settings numpad_map).
	if len(gopodBrobotsSettings.NumpadMaps[gopodNumpadLayout()]) <= 2 {
		t.Fatal("no numpad_map in the settings (clusters/gopod-system/numpad-mapper.md)")
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
	// 2026-10-30 (the numpad remap): KP0 held is the Butler, * held is All.
	// 2026-11-13: a key to a robot shows the persona the map names (KP1 Doc).
	for key, want := range map[string]string{"KP0": "BUTLER", gopodNumpadAllKey: "All", "KP1": "DOC", "KP2": "PIP", "KP3": "GIT REPO", "KP4": "CACHE PYC",
		"KP5": "INTEL 404 PNFH", "KP6": "INTEL 404 PNFT", "KP7": "CHALK", "KP8": "PLAYHEAD", "KP9": "CAMEO"} {
		if got := out.Targets[key]; got != want {
			t.Errorf("%s -> %q, want %q", key, got, want)
		}
		if gopodReporterVoice(want) == "" && key != gopodNumpadAllKey && key != "KP1" && key != "KP2" {
			t.Errorf("%s (%s) has no voice", key, want)
		}
	}
}

// ## EOF

// 2026-10-30: KP7 with NumLock off is CHALK in its reporter suit - same persona,
// answering as its reporter; a reserved NumLock-off key isn't live.
func TestNumpadNumlockOffReporter(t *testing.T) {
	old := gopodBrobotsSettings
	defer func() { gopodBrobotsSettings = old }()
	gopodBrobotsSettings.NumpadMaps = map[string]json.RawMessage{"usb": json.RawMessage(`{"numlock_on":{"KP7":{"persona":"CHALK","body":"voice-only","status":"live"}},
"numlock_off":{"KP7":{"persona":"CHALK","body":"voice-only","suit":"reporter","status":"live"},"KP8":{"persona":"-","body":"-","status":"reserved"}}}`)}
	host := gopodPersonaFor("CHALK")
	if host == "" || gopodBrobotsSettings.Managers.Personas[host].Reporter == "" {
		t.Skip("no CHALK persona with a reporter in the settings")
	}
	label, esn, err := gopodNumpadBrobot("KP7off")
	if err != nil || label != host+", reporter" || esn != gopodPersonaReporterMark+host {
		t.Fatalf("KP7off: %q %q %v", label, esn, err)
	}
	if label, esn, err := gopodNumpadBrobot("KP7"); err != nil || esn != gopodPersonaMark+host {
		t.Fatalf("KP7 (NumLock on) stays the anchor: %q %q %v", label, esn, err)
	}
	if _, _, err := gopodNumpadBrobot("KP8off"); err == nil {
		t.Fatal("KP8off is reserved, not live")
	}
}

// 2026-11-13 (operator: "separate robot bodies from personas"; "numlock-off 0-9,
// numlock-on 0-9"): the keyboard picker is the map - its keys parse, and a key
// to a robot is named by its persona (KP1 Doc), never the body.
func TestPickerIsTheMap(t *testing.T) {
	for in, want := range map[string]string{"kp:1": "KP1", "kp:9": "KP9", "kp:1off": "KP1off", "kp:0off": "KP0off"} {
		if got, ok := gopodTypedKey(in); !ok || got != want {
			t.Errorf("%s: %q %v", in, got, ok)
		}
	}
	if _, ok := gopodTypedKey("kp:0"); ok {
		t.Error("kp:0 is the system's line, not a key")
	}
	data, err := GopodNumpadMapJSONFor("")
	if err != nil {
		t.Fatal(err)
	}
	var d struct {
		Targets map[string]string `json:"targets"`
	}
	_ = json.Unmarshal(data, &d)
	if d.Targets["KP1"] == "Brobot 1" || d.Targets["KP1"] == "" {
		t.Errorf("KP1 shows %q - the persona the map names, not the body", d.Targets["KP1"])
	}
}

// 2026-11-13 (operator: "Brobot 1 = KP1 numlock-off * Brobot 2 = KP2 numlock-off
// * DOC = KP1 numlock-on * PIP = KP2 numlock-on"): four personas, two bodies.
func TestFourPersonasTwoBodies(t *testing.T) {
	for key, want := range map[string][2]string{
		"KP1":    {"Brobot 1", gopodBrobot1ESN},
		"KP2":    {"Brobot 2", gopodBrobot2ESN},
		"KP1off": {"Brobot 1", gopodDialZeroMark + gopodBrobot1ESN},
		"KP2off": {"Brobot 2", gopodDialZeroMark + gopodBrobot2ESN},
	} {
		label, target, err := gopodNumpadBrobot(key)
		if err != nil || target != want[1] || (strings.HasSuffix(key, "off") && label != want[0]) {
			t.Errorf("%s: %q %q %v", key, label, target, err)
		}
	}
	data, _ := GopodNumpadMapJSONFor("")
	var d struct {
		Targets map[string]string `json:"targets"`
	}
	_ = json.Unmarshal(data, &d)
	if d.Targets["KP1"] != "DOC" || d.Targets["KP2"] != "PIP" || d.Targets["KP1off"] != "Brobot 1" || d.Targets["KP2off"] != "Brobot 2" {
		t.Errorf("the picker's names: %v", d.Targets)
	}
}

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_brobots_setup_test.go
package wirepod_ttr

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestBrobotsSetup: who's who from Wire-Pod's robots - GOPOD's own two, any
// two, one alone (solo), none; solo, All goes to Brobot 1 and his prompt drops
// his brother.
func TestBrobotsSetup(t *testing.T) {
	cases := []struct {
		robots     []string
		set1, set2 string
		b1, b2     string
	}{
		// The operator's own two (his private file - a test one here), in his order.
		{[]string{"own00002", "own00001"}, "", "", "own00001", "own00002"},
		{[]string{"aaaa1111", "bbbb2222"}, "", "", "aaaa1111", "bbbb2222"},
		{[]string{"aaaa1111"}, "", "", "aaaa1111", gopodNoBrobot2},
		{nil, "", "", "own00001", "own00002"},
		{[]string{"aaaa1111", "bbbb2222"}, "bbbb2222", "", "bbbb2222", "aaaa1111"},
		// A label in the settings is his own robot of that label.
		{[]string{"aaaa1111", "own00002", "own00001"}, "ESN02", "ESN01", "own00002", "own00001"},
	}
	for _, c := range cases {
		if b1, b2 := gopodBrobotsPick(c.robots, c.set1, c.set2); b1 != c.b1 || b2 != c.b2 {
			t.Errorf("%v %q %q: %s %s, want %s %s", c.robots, c.set1, c.set2, b1, b2, c.b1, c.b2)
		}
	}
	k1, k2 := gopodBrobot1ESN, gopodBrobot2ESN
	defer func() { gopodBrobot1ESN, gopodBrobot2ESN = k1, k2 }()
	gopodBrobot2ESN = gopodNoBrobot2
	if !gopodSolo() {
		t.Fatal("not solo with no Brobot 2")
	}
	cast := gopodHereNowCast("Brobot 1")
	if strings.Contains(cast, gopodBrobotsSettings.HereNow.BrotherLine) {
		t.Errorf("solo, his cast still has his brother: %s", cast)
	}
	ctx := gopodPrebuildContext(gopodBrobot1ESN, "Tell me about the rig")
	if !strings.Contains(ctx, gopodBrobotsSettings.BrobotsSetup.SoloNote) {
		t.Errorf("solo, no solo note in his context")
	}
	if pair, _, _ := gopodHostPairOf("Brobot 1"); pair.Banter != "" && strings.Contains(gopodHostPersonaText("Brobot 1", ""), strings.SplitN(pair.Banter, "{", 2)[0]) {
		t.Errorf("solo, his brother banter is still in")
	}
}

// TestESNLabels (2026-10-29, operator: "publically all people see is ESN01, ESN02
// labels"): a label is its Brobot's robot now, a Brobot's serial is its label,
// Wire-Pod changes a label to the serial at its door, and no public file names a
// real serial.
func TestESNLabels(t *testing.T) {
	k1, k2 := gopodBrobot1ESN, gopodBrobot2ESN
	defer func() { gopodBrobot1ESN, gopodBrobot2ESN = k1, k2 }()
	gopodBrobot1ESN, gopodBrobot2ESN = "own00001", "own00002"
	if GopodESNReal("ESN01") != "own00001" || GopodESNReal("esn02") != "own00002" || GopodESNReal("kp:3") != "kp:3" {
		t.Error("a label isn't its robot")
	}
	if GopodESNLabel("own00002") != "ESN02" || GopodESNLabel("zzzz") != "zzzz" {
		t.Error("a Brobot's serial isn't its label")
	}
	r := httptest.NewRequest("POST", "/api-sdk/say_text?serial=ESN01", strings.NewReader("serial=ESN02&text=hi"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	gopodESNLabelsIn(r)
	if r.URL.Query().Get("serial") != "own00001" || r.PostForm.Get("serial") != "own00002" || r.FormValue("text") != "hi" {
		t.Errorf("the door didn't change the labels: %q %q", r.URL.Query().Get("serial"), r.PostForm.Get("serial"))
	}
	if gopodESNToName("Party Bro ESN02.") != "Party Bro Brobot 2." {
		t.Error("a label said as a name stays")
	}
	own1, own2 := gopodPrivateESNs()
	if own1 != "own00001" || own2 != "own00002" {
		t.Errorf("the private file reads %q %q", own1, own2)
	}
}

// TestNoRealESNInRepo (2026-10-29): no file in the public repo names one of the
// operator's own robots by its serial - read from his private ~/.gopod_esns.json
// (skipped where there's none).
func TestNoRealESNInRepo(t *testing.T) {
	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(filepath.Join(home, ".gopod_esns.json"))
	if err != nil {
		t.Skip("no private robot file here")
	}
	var own map[string]string
	if json.Unmarshal(data, &own) != nil {
		t.Skip("the private robot file doesn't read")
	}
	out, err := exec.Command("git", "-C", gopodRepoRoot(), "ls-files", "-z").Output()
	if err != nil {
		t.Skip("no git here")
	}
	for _, f := range strings.Split(string(out), "\x00") {
		if f == "" {
			continue
		}
		body, err := os.ReadFile(filepath.Join(gopodRepoRoot(), f))
		if err != nil || len(body) > 4<<20 {
			continue
		}
		for label, esn := range own {
			if len(esn) >= 6 && strings.Contains(string(body), esn) {
				t.Errorf("%s names %s's real serial - use the label %s", f, label, label)
			}
		}
	}
}

// ## EOF

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_brobots_setup_test.go
package wirepod_ttr

import (
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
		{[]string{"0dd1d8bf", "0dd1b9e9"}, "", "", "0dd1b9e9", "0dd1d8bf"},
		{[]string{"aaaa1111", "bbbb2222"}, "", "", "aaaa1111", "bbbb2222"},
		{[]string{"aaaa1111"}, "", "", "aaaa1111", gopodNoBrobot2},
		{nil, "", "", "0dd1b9e9", "0dd1d8bf"},
		{[]string{"aaaa1111", "bbbb2222"}, "bbbb2222", "", "bbbb2222", "aaaa1111"},
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

// ## EOF

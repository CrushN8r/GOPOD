// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_nicknames_test.go
package wirepod_ttr

import (
	"strings"
	"testing"
)

// TestNicknamesAndKeys: the guards are Buzzer and Switch - headed
// "Buzzer (Intel 404 PNFH)", called by their nicknames, and saying so when
// asked who they are; the page's 1-9 menu sends 3 to 9 to that numpad key.
func TestNicknamesAndKeys(t *testing.T) {
	want := map[string]string{"INTEL 404 PNFH": "Buzzer", "INTEL 404 PNFT": "Switch"}
	for host, nick := range want {
		id := gopodBrobotsSettings.HostPersonas[host].Name
		if got := gopodBrobotsSettings.ReporterVoices.Labels[host]; got != nick+" ("+id+")" {
			t.Errorf("%s header %q", host, got)
		}
		if gopodCallName(host) != nick {
			t.Errorf("%s called %q", host, gopodCallName(host))
		}
		if !strings.Contains(gopodBrobotsSettings.Managers.Personas[host].WhoLine, "I'm "+nick) {
			t.Errorf("%s doesn't say he's %s", host, nick)
		}
	}
	for in, out := range map[string]string{"kp:5": "KP5", "kp:9": "KP9", "kp:3": "KP3"} {
		if k, ok := gopodTypedKey(in); !ok || k != out {
			t.Errorf("%s -> %q", in, k)
		}
	}
	for _, in := range []string{"all", "kp:0", "kp:10", "ESN01"} {
		if _, ok := gopodTypedKey(in); ok {
			t.Errorf("%s taken as a key", in)
		}
	}
}

// ## EOF

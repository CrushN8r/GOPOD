// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_nicknames_test.go
package wirepod_ttr

import (
	"testing"
)

// TestNicknamesAndKeys: on company time the guards go by their formal names
// (2026-11-06, operator: "go with formal first names on company time * let's
// keep nicknames for bantering") - headed, called and saying so by them; Buzzer
// and Switch stay their nicknames, for banter; the page's 1-9 menu sends 3 to 9
// to that numpad key.
func TestNicknamesAndKeys(t *testing.T) {
	want := map[string]string{"INTEL 404 PNFH": "BUZZER", "INTEL 404 PNFT": "SWITCH"}
	for host, nick := range want {
		if got := gopodBrobotsSettings.ReporterVoices.Labels[host]; got != host {
			t.Errorf("%s header %q", host, got)
		}
		if gopodCallName(host) != host {
			t.Errorf("%s called %q", host, gopodCallName(host))
		}
		if gopodBrobotsSettings.HostPersonas[host].Nickname != nick {
			t.Errorf("%s's banter nickname isn't %s", host, nick)
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

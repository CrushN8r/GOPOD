// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_staff_refer_test.go
package wirepod_ttr

import (
	"strings"
	"testing"
)

// TestStaffRefer: a staff answer ending on an open question ends on the refer line instead; one
// keep_once question stays, once a session.
func TestStaffRefer(t *testing.T) {
	sr := gopodBrobotsSettings.Managers.StaffRefer
	if strings.TrimSpace(sr.Line) == "" || len(sr.KeepOnce) == 0 {
		t.Skip("no staff_refer here")
	}
	gopodStaffReferKept.Store(false)
	defer gopodStaffReferKept.Store(false)
	line := strings.TrimSpace(sr.Line)
	if got := gopodStaffRefer("Let's get this party started! 🕺💥"); got != "Let's get this party started! 🕺💥" {
		t.Errorf("no ask, unchanged: %q", got)
	}
	if got := gopodStaffRefer("Let's get this party started! 🕺💥 Let me know what you guys were thinking!"); got != "Let's get this party started! 🕺💥 "+line {
		t.Errorf("an open ask out, the refer line in: %q", got)
	}
	if got := gopodStaffRefer("Like my circuits are all scrambled. Whatcha thinking, B? Anyone else feelin' it?"); got != "Like my circuits are all scrambled. "+line {
		t.Errorf("open questions out, the refer line in: %q", got)
	}
	keep := sr.KeepOnce[0]
	if got := gopodStaffRefer("Big night. " + keep + " 😉"); !strings.Contains(got, keep) || strings.Contains(got, line) {
		t.Errorf("%q kept the first time: %q", keep, got)
	}
	if got := gopodStaffRefer("Big night. " + keep + " 😉"); strings.Contains(got, keep) || !strings.HasSuffix(got, line) {
		t.Errorf("%q once only - the second time, the refer line: %q", keep, got)
	}
}

// ## EOF

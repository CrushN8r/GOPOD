// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_admin_run_test.go
package wirepod_ttr

import "testing"

// TestAdminRun: the 100% admin run reaches everything, and the live game it
// borrowed is put back untouched.
func TestAdminRun(t *testing.T) {
	gopodGameReset("test")
	gopodQuestionFactsFor(gopodBrobot1ESN, "Tell me about the rig")
	gopodCoverage.Lock()
	before := gopodCoverage.revealed["mobile_gear"]
	gopodCoverage.Unlock()
	ok, line := gopodAdminRunLine()
	t.Log(line)
	if !ok {
		t.Errorf("admin run: %s", line)
	}
	gopodCoverage.Lock()
	after, n := gopodCoverage.revealed["mobile_gear"], len(gopodCoverage.revealed)
	gopodCoverage.Unlock()
	if before != 1 || after != 1 || n != 1 {
		t.Errorf("the live game wasn't put back: rig %d -> %d, %d topics", before, after, n)
	}
}

// ## EOF

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_session_test.go
package wirepod_ttr

import (
	"os"
	"testing"
)

// GOLDEN (gopod_session.go): one run at a time, every run checks in and out, a
// run owns the Brobots, and during a run only run words get through.
func TestSessionOneRunAtATime(t *testing.T) {
	s := gopodBrobotsSettings.Session
	if s.BusyText == "" || s.RunGateText == "" || s.StoppedText == "" {
		t.Fatal("session.busy_text, run_gate_text and stopped_text are needed")
	}
	gopodWarmGateClosed.Store(false)
	gopodSession.Lock()
	gopodSession.run = nil
	gopodSession.Unlock()
	defer GopodRunEnd("", "done")

	if GopodSessionPhase() != "ready" {
		t.Fatalf("phase %q, want ready", GopodSessionPhase())
	}
	if ok, _ := GopodRunStart("baby brobots", "song", os.Getpid()); !ok {
		t.Fatal("the first run is refused")
	}
	if GopodSessionPhase() != "run" || !gopodSongRunning() {
		t.Fatal("a song run checked in is not seen")
	}
	if ok, busy := GopodRunStart("interview", "draft", 0); ok || busy != "baby brobots" {
		t.Fatalf("a second run started (ok=%v busy=%q)", ok, busy)
	}
	// Rule D: a live question is dropped; the thanks is let through by the door
	// itself (gopodOperatorLine checks gopodIsThanks first).
	if !gopodRunGateDrop(gopodBrobot1ESN, "Typed text", "party") {
		t.Error("a live question during a run reached the Brobots")
	}
	GopodRunEnd("baby brobots", "done")
	if gopodRunActive() || gopodRunGateDrop(gopodBrobot1ESN, "Typed text", "party") {
		t.Error("the run didn't check out")
	}
}

// A run whose process is gone without checking out is cleared.
func TestSessionClearsAGoneRun(t *testing.T) {
	gopodWarmGateClosed.Store(false)
	gopodSession.Lock()
	gopodSession.run = &gopodRun{Name: "crashed", Kind: "song", PID: 1 << 22}
	gopodSession.Unlock()
	if gopodRunActive() {
		t.Error("a run with no process is still counted")
	}
	// Nothing starts before LLM Ready.
	gopodWarmGateClosed.Store(true)
	defer gopodWarmGateClosed.Store(false)
	if ok, why := GopodRunStart("bingo test", "song", 0); ok || why != "warming" {
		t.Errorf("a run started while warming (ok=%v why=%q)", ok, why)
	}
}

// ## EOF

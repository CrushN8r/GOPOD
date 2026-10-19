// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_native_test.go
package wirepod_ttr

import "testing"

// Wire-Pod's own answers: the Native? box, and always off during a run (gopod_native.go).
func TestNativeAnswersOnOffAndDuringARun(t *testing.T) {
	gopodWarmGateClosed.Store(false)
	gopodSession.Lock()
	gopodSession.run = nil
	gopodSession.Unlock()
	saved := GopodNativeOn()
	defer GopodNativeSet(saved)

	GopodNativeSet(true)
	if gopodNativeBlocked(nil, gopodBrobot1ESN, "intent_clock_time", "what time is it") {
		t.Error("Native? on, no run: Wire-Pod's answer was blocked")
	}
	GopodNativeSet(false)
	if !gopodNativeBlocked(nil, gopodBrobot1ESN, "intent_clock_time", "what time is it") {
		t.Error("Native? off: Wire-Pod's answer got through")
	}
	GopodNativeSet(true)
	if ok, _ := GopodRunStart("bingo test", "song", 0); !ok {
		t.Fatal("run refused")
	}
	defer GopodRunEnd("bingo test", "done")
	if !gopodNativeBlocked(nil, gopodBrobot1ESN, "intent_clock_time", "what time is it") {
		t.Error("during a run Wire-Pod's answer got through")
	}
}

// ## EOF

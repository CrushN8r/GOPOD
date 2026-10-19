// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_thanks_close_out_test.go
package wirepod_ttr

import (
	"testing"
	"time"
)

// A thanks while a draft plays stops the draft first and takes the line; with
// nothing going it passes on. "stop run" is a command either way.
func TestThanksClosesOutTheDraft(t *testing.T) {
	gopodSession.Lock()
	gopodSession.run = nil // no run going in a test
	gopodSession.Unlock()
	if gopodBrobotsSettings.SignOff.WelcomeText == "" || len(gopodBrobotsSettings.SongStop.StopWords) == 0 {
		t.Fatal("sign_off.welcome_text and song_stop.stop_words are needed")
	}
	if !gopodStopRunCommand("Stop run") || gopodStopRunCommand("how are you") {
		t.Error("stop run: the command, and nothing else")
	}
	if gopodThanksCloseOut(gopodBrobot1ESN, "thanks Brobots", "Typed text") {
		t.Error("nothing going: the thanks goes on as usual")
	}
	// The re-sent thanks waits in the closed gate's queue (2026-10-13), which
	// this test empties, so no LLM is asked.
	gopodWarmGateClosed.Store(true)
	defer func() {
		gopodWarming.Lock()
		gopodWarming.queue = nil
		gopodWarming.Unlock()
		gopodWarmGateClosed.Store(false)
	}()
	gopodScriptPlaying.Store(true)
	gopodScriptStop.Store(false)
	if !gopodThanksCloseOut(gopodBrobot1ESN, "thanks Brobots", "Typed text") {
		t.Fatal("a playing draft: the thanks closes it out first")
	}
	if !gopodScriptStop.Load() {
		t.Error("the draft was not asked to stop")
	}
	if !gopodThanksCloseOut(gopodBrobot1ESN, "thanks Brobots", "Typed text") {
		t.Error("a second thanks while closing out is taken too")
	}
	gopodScriptPlaying.Store(false) // the player ends
	for i := 0; i < 20 && gopodClosingOut.Load(); i++ {
		time.Sleep(100 * time.Millisecond)
	}
	if gopodClosingOut.Load() {
		t.Error("the close-out never finished")
	}
	gopodScriptStop.Store(false)
}

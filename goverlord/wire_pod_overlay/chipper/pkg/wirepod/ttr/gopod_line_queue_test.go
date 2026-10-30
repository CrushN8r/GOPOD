// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_line_queue_test.go
package wirepod_ttr

import (
	"sync/atomic"
	"testing"
	"time"
)

// TestLineQueue: two questions sent together - the second starts only once the
// first has its answer on the feed.
func TestLineQueue(t *testing.T) {
	saved := gopodBrobotsSettings.LineQueue
	defer func() { gopodBrobotsSettings.LineQueue = saved }()
	gopodBrobotsSettings.LineQueue.On = true
	gopodBrobotsSettings.LineQueue.AnswerMaxSeconds = 10
	gopodBrobotsSettings.LineQueue.IdleSeconds = 10
	gopodBrobotsSettings.LineQueue.PageMaxSeconds = 1
	var answeredAt, secondAt time.Time
	gopodLineTurn("first", []string{"Brobot 1"}, func() error {
		go func() {
			time.Sleep(700 * time.Millisecond)
			answeredAt = time.Now()
			GopodChatFeedAdd("Brobot 1", "first answer")
		}()
		return nil
	})
	done := make(chan bool)
	gopodLineTurn("second", []string{"Brobot 2"}, func() error {
		secondAt = time.Now()
		done <- true
		return nil
	})
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("the second question never ran")
	}
	if answeredAt.IsZero() || secondAt.Before(answeredAt) {
		t.Errorf("the second question started before the first was answered (%v / %v)", secondAt, answeredAt)
	}
}

// ## EOF

// CLOCK_AUDIT_001.md A (2026-10-30): a question's idle wait doesn't count while an
// answer is going - a Brobot's reply from its start (his thinking too) to its end,
// or a Gomad's answer being made.
func TestLineQueueIdleWaitsForAnswer(t *testing.T) {
	if gopodAnswerGoing([]string{"Brobot 1"}) {
		t.Fatal("no reply going yet")
	}
	gopodReplyStart(gopodBrobot1ESN)
	if !gopodAnswerGoing([]string{"Brobot 1", "Brobot 2"}) {
		t.Error("Brobot 1 thinking: his answer is going")
	}
	gopodReplyEnd(gopodBrobot1ESN)
	if gopodAnswerGoing([]string{"Brobot 1"}) {
		t.Error("Brobot 1's reply ended")
	}
	n, _ := gopodGomadAnswers.LoadOrStore("CHALK", &atomic.Int32{})
	n.(*atomic.Int32).Add(1)
	if !gopodAnswerGoing([]string{"CHALK"}) {
		t.Error("CHALK's answer is being made")
	}
	n.(*atomic.Int32).Add(-1)
	if gopodAnswerGoing([]string{"CHALK"}) {
		t.Error("CHALK's answer is done")
	}
	gopodVoicesMaking.Add(1)
	if !gopodAnswerGoing([]string{"Brobot 1"}) {
		t.Error("a voice being made: an answer is going")
	}
	gopodVoicesMaking.Add(-1)
}

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_line_queue_test.go
package wirepod_ttr

import (
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

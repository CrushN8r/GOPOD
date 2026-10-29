// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_chat_queue_test.go
package wirepod_ttr

import (
	"strconv"
	"testing"
	"time"
)

// gopodFeedLast: the newest line on the feed.
func gopodFeedLast(t *testing.T) gopodChatFeedEvent {
	t.Helper()
	gopodChatFeed.Lock()
	defer gopodChatFeed.Unlock()
	if len(gopodChatFeed.events) == 0 {
		t.Fatal("no feed lines")
	}
	return gopodChatFeed.events[len(gopodChatFeed.events)-1]
}

// TestMasterChatQueue (2026-10-29, operator: "every event should go in a main
// chat-queue of nested chat-queues, so all processes have a chance to run to
// completion"): a turn's lines carry its number; a question asked while a turn
// that holds questions is going waits for its end; inside a turn that doesn't,
// a question is that turn's line; a question alone is a turn of its own; a drop
// is a control line naming the turn.
func TestMasterChatQueue(t *testing.T) {
	if gopodTurnNow.Load() != 0 {
		t.Fatal("a turn is going before the test")
	}
	saved := gopodBrobotsSettings.LineQueue
	defer func() { gopodBrobotsSettings.LineQueue = saved }()
	gopodBrobotsSettings.LineQueue.PageMaxSeconds = 0.1
	gopodBrobotsSettings.LineQueue.IdleSeconds = 0.1
	gopodBrobotsSettings.LineQueue.AnswerMaxSeconds = 1
	// A line with nothing going belongs to no turn.
	GopodChatFeedAdd("Goverlord", "queue test: no turn")
	if gopodFeedLast(t).Turn != 0 {
		t.Error("a line with nothing going has a turn")
	}
	// The tour: a turn that holds questions.
	tour := gopodTurnBegin("tour", true)
	GopodChatFeedAdd("Goverlord", "queue test: the tour's line")
	if gopodFeedLast(t).Turn != tour {
		t.Error("the tour's line doesn't carry its turn")
	}
	asked := make(chan int64, 1)
	go gopodLineRun(gopodLineJob{text: "a question", answerers: nil, run: func() error {
		asked <- gopodTurnNow.Load()
		return nil
	}})
	select {
	case <-asked:
		t.Fatal("a question ran during a turn that holds questions")
	case <-time.After(700 * time.Millisecond):
	}
	// A drop: a control line (no turn) naming the tour's turn.
	if !gopodTurnDrop("test") {
		t.Fatal("no turn to drop")
	}
	if l := gopodFeedLast(t); l.Turn != 0 || l.Text != gopodTurnDropMark+strconv.FormatInt(tour, 10) {
		t.Errorf("the drop line: turn %d, %q", l.Turn, l.Text)
	}
	// A control line during the dropped turn (the Esc reply) is no turn's.
	gopodChatFeedAddTurn(0, "Terminal", "queue test: Esc - the run stops.")
	if gopodFeedLast(t).Turn != 0 {
		t.Error("a control line carries the dropped turn")
	}
	gopodTurnEnd(tour)
	var own int64
	select {
	case own = <-asked:
	case <-time.After(3 * time.Second):
		t.Fatal("the question never ran after the tour")
	}
	if own == 0 || own == tour {
		t.Errorf("a question alone isn't a turn of its own: %d (tour %d)", own, tour)
	}
	time.Sleep(400 * time.Millisecond)
	// The comm check: a turn that doesn't hold - its question is its own line.
	cc := gopodTurnBegin("comm_check", false)
	got := make(chan int64, 1)
	gopodLineRun(gopodLineJob{text: "KP3 who are you", run: func() error { got <- gopodTurnNow.Load(); return nil }})
	if g := <-got; g != cc {
		t.Errorf("the comm check's question is turn %d, not the comm check's %d", g, cc)
	}
	// Each thread its own: the control lane's lines are no turn's while the comm
	// check goes on; a thread owning a turn stamps it whatever turn is going.
	func() {
		defer gopodTurnOwn(0)()
		GopodChatFeedAdd("Operator", "queue test: stop run")
		if gopodFeedLast(t).Turn != 0 {
			t.Error("a control line during a turn carries the turn")
		}
	}()
	done := make(chan int64)
	go func() {
		defer gopodTurnOwn(99)()
		GopodChatFeedAdd("Goverlord", "queue test: a thread's own turn")
		done <- 0
	}()
	<-done
	if gopodFeedLast(t).Turn != 99 {
		t.Error("a thread owning a turn doesn't stamp it")
	}
	gopodTurnEnd(cc)
	if gopodTurnNow.Load() != 0 {
		t.Error("a turn is still going after its end")
	}
}

// ## EOF

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_line_queue.go
package wirepod_ttr

// The line queue (2026-10-29, operator: "even if a few requests come in
// together, they still must each get processed until completion before another
// bubbles starts"; live 05:13: KP2's line came before Brobot 1's reply, and the
// replies landed among KP3's bubbles). Every question - typed, numpad, a
// persona's, the comm check's - takes its turn: one at a time, in the order they
// came; the next starts only when this one is complete - a line from each one
// asked after it (his answer), then the page has played it all. Commands (system-,
// stop run, thanks, skip, start ...) don't wait - they never reach the queue.
// line_queue: on, and how long each wait may take.

import (
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

type gopodLineJob struct {
	text      string
	answerers []string
	run       func() error
}

var gopodLineJobs = make(chan gopodLineJob, 64)

// gopodLineBusy: a question is being asked and answered now.
var gopodLineBusy atomic.Bool

func init() {
	go func() {
		for j := range gopodLineJobs {
			gopodLineRun(j)
		}
	}()
}

// gopodLineTurn queues a question; off (line_queue.on false), it runs now.
func gopodLineTurn(text string, answerers []string, run func() error) error {
	if !gopodBrobotsSettings.LineQueue.On {
		return run()
	}
	if n := len(gopodLineJobs); n > 0 {
		logger.Println("GOPOD_LINE_QUEUE waiting (" + strings.Join(answerers, ", ") + "), " + strconv.Itoa(n) + " ahead: " + gopodCut(text, 60))
	}
	gopodLineJobs <- gopodLineJob{text: text, answerers: answerers, run: run}
	return nil
}

// gopodLineRun: one question, then its answers, then the page caught up - its
// own turn of the master chat queue (gopod_chat_queue.go) when nothing else is
// going; a turn that holds questions (the tour) is waited out first; inside a
// turn that doesn't (the comm check's own questions), it's that turn's line.
func gopodLineRun(j gopodLineJob) {
	for gopodTurnHolding() {
		time.Sleep(300 * time.Millisecond)
	}
	gopodLineBusy.Store(true)
	defer gopodLineBusy.Store(false)
	if gopodTurnNow.Load() == 0 {
		id := gopodTurnSeq.Add(1)
		gopodTurn.Lock()
		gopodTurn.id, gopodTurn.kind, gopodTurn.holds = id, "question", false
		gopodTurn.Unlock()
		gopodTurnNow.Store(id)
		defer gopodTurnEnd(id)
		defer gopodTurnOwn(id)()
	}
	lq := gopodBrobotsSettings.LineQueue
	from := gopodFeedLastID()
	if err := j.run(); err != nil {
		logger.Println("GOPOD_LINE_QUEUE " + err.Error())
		return
	}
	idle := gopodSecondsToDuration(lq.IdleSeconds)
	end := time.Now().Add(gopodSecondsToDuration(lq.AnswerMaxSeconds))
	last, since := gopodFeedLastID(), time.Now()
	for time.Now().Before(end) {
		// CLOCK_AUDIT_001.md B (2026-10-30): answered once every answerer's
		// bubble is on the feed and his answer has ended - said, not just posted.
		if gopodLineAnswered(from, j.answerers) && !gopodAnswerGoing(j.answerers) {
			break
		}
		// 2026-10-30 (CLOCK_AUDIT_001.md A, operator: "go for A"): an answer still
		// coming - a Brobot thinking or speaking, a Gomad's answer being made - is
		// not "nothing more came"; idle counts only once no answer is going.
		if id := gopodFeedLastID(); id != last || gopodAnswerGoing(j.answerers) {
			last, since = id, time.Now()
		} else if time.Since(since) >= idle {
			logger.Println("GOPOD_LINE_QUEUE nothing more came for " + gopodCut(j.text, 60))
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
	// The page shows it all - when a page is here to show it (gopodPageHere).
	if gopodPageHere() {
		pend := time.Now().Add(gopodSecondsToDuration(lq.PageMaxSeconds))
		for time.Now().Before(pend) && gopodPagePlayedUpTo() < gopodFeedLastID() {
			time.Sleep(300 * time.Millisecond)
		}
	}
}

// gopodLineAnswered: every one asked has a line on the feed after from.
func gopodLineAnswered(from int, who []string) bool {
	gopodChatFeed.Lock()
	defer gopodChatFeed.Unlock()
	for _, w := range who {
		found := false
		for i := len(gopodChatFeed.events) - 1; i >= 0 && gopodChatFeed.events[i].ID > from; i-- {
			if strings.EqualFold(gopodChatFeed.events[i].Who, w) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// gopodLineQueueClear: the questions still waiting their turn dropped (stop run); how many.
func gopodLineQueueClear() int {
	n := 0
	for {
		select {
		case <-gopodLineJobs:
			n++
		default:
			return n
		}
	}
}

// gopodVoicesMaking: voice files being made now (gopodReporterFile).
var gopodVoicesMaking atomic.Int32

// gopodAnswerGoing: a voice is being made, or one of these answerers' answers is still going - a Brobot's
// reply from its start to its end (gopodReplyGoing), or a Gomad's answer being
// made (gopodGomadAnswering).
func gopodAnswerGoing(answerers []string) bool {
	if gopodVoicesMaking.Load() > 0 {
		return true
	}
	for _, who := range answerers {
		if esn := gopodBrobotESN(who); esn != "" && gopodReplyGoing(esn) {
			return true
		}
		if gopodGomadAnswering(who) {
			return true
		}
	}
	return false
}

// ## EOF

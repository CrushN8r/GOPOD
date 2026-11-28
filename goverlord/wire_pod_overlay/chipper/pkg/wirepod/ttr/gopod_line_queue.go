// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_line_queue.go
package wirepod_ttr

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
	if max := gopodBrobotsSettings.LineQueue.MaxWaiting; max > 0 && len(gopodLineJobs) >= max {
		logger.Println("GOPOD_LINE_QUEUE full (" + strconv.Itoa(max) + " waiting), dropped: " + gopodCut(text, 60))
		return nil // at most max_waiting lines wait; the next one is dropped
	}
	if n := len(gopodLineJobs); n > 0 {
		logger.Println("GOPOD_LINE_QUEUE waiting (" + strings.Join(answerers, ", ") + "), " + strconv.Itoa(n) + " ahead: " + gopodCut(text, 60))
	}
	gopodLineJobs <- gopodLineJob{text: text, answerers: answerers, run: run}
	return nil
}

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
	most := gopodSecondsToDuration(lq.AnswerMaxSeconds)
	end := time.Now().Add(most)
	extended := false
	last, since := gopodFeedLastID(), time.Now()
	for {
		if !time.Now().Before(end) {
			if extended {
				break
			}
			extended = true
			end = time.Now().Add(most)
			logger.Println("GOPOD_LINE_QUEUE " + most.String() + ", not finished - another " + most.String() + " for " + gopodCut(j.text, 60))
		}
		if gopodLineAnswered(from, j.answerers) && !gopodAnswerGoing(j.answerers) {
			break
		}
		if id := gopodFeedLastID(); id != last || gopodAnswerGoing(j.answerers) {
			last, since = id, time.Now()
		} else if time.Since(since) >= idle {
			logger.Println("GOPOD_LINE_QUEUE nothing more came for " + gopodCut(j.text, 60))
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
	if !time.Now().Before(end) {
		logger.Println("GOPOD_LINE_QUEUE answer time ran out (" + most.String() + ") for " + gopodCut(j.text, 60))
	}
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

// gopodAnswerGoing: a voice is being made, or one of these answerers' answers is still going - a
// Brobot's reply from its start to its end (gopodReplyGoing), or a Gomad's answer being made
// (gopodGomadAnswering).
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

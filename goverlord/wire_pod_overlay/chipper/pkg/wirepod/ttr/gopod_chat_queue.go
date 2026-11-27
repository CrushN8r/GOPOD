// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_chat_queue.go
package wirepod_ttr

import (
	"bytes"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// gopodTurnNow: the turn the chat's lines belong to now (0: none - the opener, a system answer with
// nothing going).
var gopodTurnNow atomic.Int64

var gopodTurnSeq atomic.Int64

var gopodTurn = struct {
	sync.Mutex
	id    int64
	kind  string
	holds bool // a question asked meanwhile waits for this turn's end
}{}

// gopodTurnBegin opens a turn (kind: "run", "tour", "comm_check"); a question still being answered
// finishes first (at most line_queue.answer_max_seconds).
func gopodTurnBegin(kind string, holds bool) int64 {
	end := time.Now().Add(gopodSecondsToDuration(gopodBrobotsSettings.LineQueue.AnswerMaxSeconds))
	for gopodLineBusy.Load() && time.Now().Before(end) {
		time.Sleep(200 * time.Millisecond)
	}
	return gopodTurnOpen(kind, holds)
}

// gopodTurnBeginAtOnce opens a run's turn without waiting: the line busy is the one that started
// the run - its native reply ("Got it!", "Ok.
func gopodTurnBeginAtOnce(kind string, holds bool) int64 {
	return gopodTurnOpen(kind, holds)
}

func gopodTurnOpen(kind string, holds bool) int64 {
	id := gopodTurnSeq.Add(1)
	gopodTurn.Lock()
	gopodTurn.id, gopodTurn.kind, gopodTurn.holds = id, kind, holds
	gopodTurn.Unlock()
	gopodTurnNow.Store(id)
	logger.Println("GOPOD_TURN " + strconv.FormatInt(id, 10) + " " + kind + " begins")
	return id
}

// gopodTurnBeginOwn: gopodTurnBegin, this thread owning the turn's lines (the tour's, the comm
// check's); the returned func ends the turn.
func gopodTurnBeginOwn(kind string, holds bool) func() {
	id := gopodTurnBegin(kind, holds)
	release := gopodTurnOwn(id)
	return func() {
		release()
		gopodTurnEnd(id)
	}
}

// gopodTurnEnd closes the turn (only that one - a later turn is left alone).
func gopodTurnEnd(id int64) {
	gopodTurn.Lock()
	if gopodTurn.id == id {
		gopodTurn.id, gopodTurn.kind, gopodTurn.holds = 0, "", false
	}
	gopodTurn.Unlock()
	if gopodTurnNow.CompareAndSwap(id, 0) {
		logger.Println("GOPOD_TURN " + strconv.FormatInt(id, 10) + " ends")
	}
}

// gopodTurnHolding: a turn going that makes questions wait.
func gopodTurnHolding() bool {
	gopodTurn.Lock()
	defer gopodTurn.Unlock()
	return gopodTurn.id != 0 && gopodTurn.holds
}

// gopodTurnDrop (stop run, Esc): the turn going is told to the page to drop - its lines still
// waiting and any still coming.
func gopodTurnDrop(why string) bool {
	id := gopodTurnNow.Load()
	if id == 0 {
		return false
	}
	logger.Println("GOPOD_TURN " + strconv.FormatInt(id, 10) + " dropped (" + why + ")")
	gopodChatFeedAddTurn(0, gopodBrobotsSettings.ReporterVoices.Role, gopodTurnDropMark+strconv.FormatInt(id, 10))
	return true
}

var gopodTurnOwners sync.Map // goroutine id -> turn

// gopodGoID: this goroutine's number.
func gopodGoID() uint64 {
	var b [64]byte
	n := runtime.Stack(b[:], false)
	f := bytes.Fields(b[:n])
	if len(f) < 2 {
		return 0
	}
	id, _ := strconv.ParseUint(string(f[1]), 10, 64)
	return id
}

// gopodTurnOwn: this thread's lines are turn's until the returned func is called.
func gopodTurnOwn(turn int64) func() {
	id := gopodGoID()
	prev, had := gopodTurnOwners.Load(id)
	gopodTurnOwners.Store(id, turn)
	return func() {
		if had {
			gopodTurnOwners.Store(id, prev)
		} else {
			gopodTurnOwners.Delete(id)
		}
	}
}

// gopodTurnForLine: the turn a line posted now belongs to - its thread's own, or the turn going.
func gopodTurnForLine() int64 {
	if v, ok := gopodTurnOwners.Load(gopodGoID()); ok {
		return v.(int64)
	}
	return gopodTurnNow.Load()
}

// gopodTurnDropMark: the feed line that drops a turn on the page.
const gopodTurnDropMark = "drop:"

// ## EOF

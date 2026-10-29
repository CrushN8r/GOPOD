// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_chat_queue.go
package wirepod_ttr

// GOLDEN, the master chat queue (2026-10-29, operator: "Like every event should
// go in a main chat-queue of nested chat-queues, so all processes have a chance
// to run to completion, no?"; "build the master robust bulletproof chat queue of
// nested chat queues"). Everything the chat shows belongs to a turn, and turns run
// one at a time, each to its end:
//
//   - A question (typed, numpad, a persona's) is a turn of its own when nothing
//     else is going - the line queue (gopod_line_queue.go): it, then its answers,
//     then the page caught up.
//   - A run is a turn from check-in to check-out (GopodRunStart / GopodRunEnd:
//     a song, the scripted draft); the tour and the comm check are turns too.
//     Their own lines - a song's from its runner, the comm check's KP1-9
//     questions, the tour's commandments - are the turn's nested queue, stamped
//     with its number.
//   - A turn that holds questions (the tour) makes a question asked meanwhile
//     wait its turn; a run's own rules keep chat questions out of a song
//     (Rule D, gopod_session.go), and the comm check's questions are its own.
//   - The control lane never waits: stop run, Esc, skip, system- commands.
//
// Every chat line carries its turn's number (the feed's "turn"). Stop run and Esc
// end the turn going and tell the page at once (gopodTurnDrop: a "drop:<turn>"
// line the page acts on the moment it arrives): the page drops that turn's lines
// still waiting and any still coming, and finishes the bubble on screen.

import (
	"bytes"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// gopodTurnNow: the turn the chat's lines belong to now (0: none - the opener,
// a system answer with nothing going).
var gopodTurnNow atomic.Int64

var gopodTurnSeq atomic.Int64

var gopodTurn = struct {
	sync.Mutex
	id    int64
	kind  string
	holds bool // a question asked meanwhile waits for this turn's end
}{}

// gopodTurnBegin opens a turn (kind: "run", "tour", "comm_check"); a question
// still being answered finishes first (at most line_queue.answer_max_seconds).
func gopodTurnBegin(kind string, holds bool) int64 {
	end := time.Now().Add(gopodSecondsToDuration(gopodBrobotsSettings.LineQueue.AnswerMaxSeconds))
	for gopodLineBusy.Load() && time.Now().Before(end) {
		time.Sleep(200 * time.Millisecond)
	}
	id := gopodTurnSeq.Add(1)
	gopodTurn.Lock()
	gopodTurn.id, gopodTurn.kind, gopodTurn.holds = id, kind, holds
	gopodTurn.Unlock()
	gopodTurnNow.Store(id)
	logger.Println("GOPOD_TURN " + strconv.FormatInt(id, 10) + " " + kind + " begins")
	return id
}

// gopodTurnBeginOwn: gopodTurnBegin, this thread owning the turn's lines (the
// tour's, the comm check's); the returned func ends the turn.
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

// gopodTurnDrop (stop run, Esc): the turn going is told to the page to drop -
// its lines still waiting and any still coming. True when there was one.
func gopodTurnDrop(why string) bool {
	id := gopodTurnNow.Load()
	if id == 0 {
		return false
	}
	logger.Println("GOPOD_TURN " + strconv.FormatInt(id, 10) + " dropped (" + why + ")")
	gopodChatFeedAddTurn(0, gopodBrobotsSettings.ReporterVoices.Role, gopodTurnDropMark+strconv.FormatInt(id, 10))
	return true
}

// Each process carries its own turn (2026-10-29, operator: "stop run * not heard *
// did templated bubble text mechanism go through a voice file generation templated
// check"): the tour's, the comm check's and a question's own thread, and the
// control lane's (a typed command, stop run, Esc - no turn), so a line is stamped
// by who posted it - the stop's own lines go through the whole template (shown,
// read out, heard) and stay, while the dropped run's last lines go with it. A line
// from any other thread (a song's, through Wire-Pod's API) is the turn going.
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

// gopodTurnForLine: the turn a line posted now belongs to - its thread's own, or
// the turn going.
func gopodTurnForLine() int64 {
	if v, ok := gopodTurnOwners.Load(gopodGoID()); ok {
		return v.(int64)
	}
	return gopodTurnNow.Load()
}

// gopodTurnDropMark: the feed line that drops a turn on the page.
const gopodTurnDropMark = "drop:"

// ## EOF

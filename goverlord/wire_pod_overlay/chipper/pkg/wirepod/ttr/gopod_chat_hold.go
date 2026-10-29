// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_chat_hold.go
package wirepod_ttr

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// The Brobots stay in place while chatting (operator 2026-10-04: "once I start
// brobot chatting, the brobots are instructed to stay in place and not
// take-off/wander/explore", "Hold both Brobots between answers", "I should have a
// thanks brobots cue to re-engage"). After a Brobot's chat answer, both Brobots
// are held under behavior control at DEFAULT priority - the same hold the songs
// use (sdkapp assume_behavior_control), with which voice commands still reach a
// held robot. A Brobot's hold is let go while he answers or defers (his own
// control takes over) and taken again after. The hold ends on "thanks Brobots"
// (GopodChatHoldEnd), when a run starts (gopod_session.go), after
// chat_hold.idle_minutes with no question, or for one robot when anything else
// assumes his control through Wire-Pod (a song, an alias, phcal).

var gopodChatHold = struct {
	sync.Mutex
	active   bool
	last     time.Time
	release  map[string]chan bool
	watching bool
}{release: map[string]chan bool{}}

// gopodChatHoldBeforeAnswer: this Brobot is about to answer or defer - his hold
// is let go so his own control takes over.
func gopodChatHoldBeforeAnswer(esn string) {
	gopodChatHold.Lock()
	gopodChatHold.last = time.Now()
	gopodChatHold.Unlock()
	gopodChatHoldLetGo(esn)
}

// gopodChatHoldAfterAnswer: a chat answer is done - the hold starts (or goes on)
// for both Brobots.
func gopodChatHoldAfterAnswer() {
	gopodChatHold.Lock()
	if !gopodChatHold.active {
		logger.Println("GOPOD_CHAT_HOLD start (both Brobots stay in place)")
	}
	gopodChatHold.active = true
	gopodChatHold.last = time.Now()
	if !gopodChatHold.watching {
		gopodChatHold.watching = true
		go gopodChatHoldWatchIdle()
	}
	gopodChatHold.Unlock()
	gopodChatHoldTakeBoth()
}

// gopodChatHoldResume: a Brobot that deferred is held again if the hold is on.
func gopodChatHoldResume(esn string) {
	gopodChatHold.Lock()
	active := gopodChatHold.active
	gopodChatHold.Unlock()
	if active {
		go func() {
			time.Sleep(gopodChatHoldRetakeDelay)
			gopodChatHoldTake(esn)
		}()
	}
}

// gopodIsThanks: the question is a "thanks Brobots" (chat_hold.thanks_phrases,
// whole words) - the Brobot signs off and the hold ends after his answer.
func gopodIsThanks(text string) bool {
	padded := " " + strings.Join(gopodAddressWords(text), " ") + " "
	for _, phrase := range gopodBrobotsSettings.ChatHold.ThanksPhrases {
		if p := strings.Join(gopodAddressWords(phrase), " "); p != "" && strings.Contains(padded, " "+p+" ") {
			return true
		}
	}
	return false
}

// GopodChatHoldEnd ends the hold for both Brobots - back to their own free play.
func GopodChatHoldEnd(reason string) {
	gopodChatHold.Lock()
	was := gopodChatHold.active
	gopodChatHold.active = false
	gopodChatHold.Unlock()
	gopodChatHoldLetGo(gopodBrobot1ESN)
	gopodChatHoldLetGo(gopodBrobot2ESN)
	if was {
		logger.Println("GOPOD_CHAT_HOLD end reason=" + reason)
	}
}

// GopodChatHoldRelease lets go of one robot's hold because something else is
// taking his control through Wire-Pod (sdkapp assume_behavior_control). The
// hold comes back with the next chat answer.
func GopodChatHoldRelease(esn string) {
	gopodChatHoldLetGo(esn)
}

// The answering Brobot's own control is released when his reply ends; the hold
// is taken just after.
const gopodChatHoldRetakeDelay = 500 * time.Millisecond

func gopodChatHoldTakeBoth() {
	go func() {
		time.Sleep(gopodChatHoldRetakeDelay)
		gopodChatHoldTake(gopodBrobot1ESN)
		gopodChatHoldTake(gopodBrobot2ESN)
	}()
}

func gopodChatHoldWatchIdle() {
	for {
		time.Sleep(gopodTiming("chat_hold_check_seconds"))
		idle := time.Duration(gopodBrobotsSettings.ChatHold.IdleMinutes * float64(time.Minute))
		gopodChatHold.Lock()
		expired := gopodChatHold.active && time.Since(gopodChatHold.last) > idle
		gopodChatHold.Unlock()
		if expired {
			GopodChatHoldEnd("idle")
		}
	}
}

// gopodChatHoldTake holds one robot (if the hold is on and he isn't held yet).
func gopodChatHoldTake(esn string) {
	// 2026-10-10 (live 07:40: the hold from a "Party" answer was still on when
	// "start baby brobots" put the Brobots to sleep over the direct SDK - two
	// controls on one robot): a run owns the Brobots; no hold is taken
	// (gopod_session.go, Rule C).
	if gopodRunActive() {
		return
	}
	gopodChatHold.Lock()
	if !gopodChatHold.active || gopodChatHold.release[esn] != nil {
		gopodChatHold.Unlock()
		return
	}
	release := make(chan bool)
	gopodChatHold.release[esn] = release
	gopodChatHold.Unlock()

	var robot *vector.Vector
	for _, bot := range vars.BotInfo.Robots {
		if bot.Esn == esn {
			robot, _ = vector.New(vector.WithSerialNo(esn), vector.WithToken(bot.GUID), vector.WithTarget(bot.IPAddress+":443"))
			break
		}
	}
	if robot == nil {
		gopodChatHoldForget(esn, release)
		return
	}
	go func() {
		defer gopodChatHoldForget(esn, release)
		r, err := robot.Conn.BehaviorControl(context.Background())
		if err != nil {
			logger.Println("GOPOD_CHAT_HOLD " + gopodBrobotName(esn) + " failed: " + err.Error())
			return
		}
		if err := r.Send(&vectorpb.BehaviorControlRequest{
			RequestType: &vectorpb.BehaviorControlRequest_ControlRequest{
				ControlRequest: &vectorpb.ControlRequest{Priority: vectorpb.ControlRequest_DEFAULT},
			},
		}); err != nil {
			logger.Println("GOPOD_CHAT_HOLD " + gopodBrobotName(esn) + " failed: " + err.Error())
			return
		}
		logger.Println("GOPOD_CHAT_HOLD " + gopodBrobotName(esn) + " held")
		<-release
		r.Send(&vectorpb.BehaviorControlRequest{
			RequestType: &vectorpb.BehaviorControlRequest_ControlRelease{ControlRelease: &vectorpb.ControlRelease{}},
		})
		logger.Println("GOPOD_CHAT_HOLD " + gopodBrobotName(esn) + " let go")
	}()
}

// gopodChatHoldLetGo releases one robot's hold, if he has one.
func gopodChatHoldLetGo(esn string) {
	gopodChatHold.Lock()
	release := gopodChatHold.release[esn]
	delete(gopodChatHold.release, esn)
	gopodChatHold.Unlock()
	if release == nil {
		return
	}
	select {
	case release <- true:
		// A moment for the release to reach the robot before new control.
		time.Sleep(200 * time.Millisecond)
	case <-time.After(2 * time.Second):
	}
}

func gopodChatHoldForget(esn string, release chan bool) {
	gopodChatHold.Lock()
	if gopodChatHold.release[esn] == release {
		delete(gopodChatHold.release, esn)
	}
	gopodChatHold.Unlock()
}

// ## EOF

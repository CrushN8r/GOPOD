// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_stuck_reply.go
package wirepod_ttr

import (
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/sashabaranov/go-openai"
)

func gopodStuckWait() time.Duration {
	return time.Duration(gopodBrobotsSettings.ChatHold.StuckSeconds * float64(time.Second))
}

func gopodLogStuck(esn, what string) {
	logger.Println("GOPOD_STUCK_REPLY " + gopodBrobotName(esn) + " " + what + " waited " + gopodStuckWait().String() + " - answer ended, robot dropped off?")
}

// gopodPerformOrGiveUp is PerformActions for one action, or true (disconnect, ends the answer) when
// the robot doesn't finish it in time.
func gopodPerformOrGiveUp(msgs []openai.ChatCompletionMessage, act RobotAction, robot *vector.Vector, stop chan bool) bool {
	done := make(chan bool, 1)
	go func() { done <- PerformActions(msgs, []RobotAction{act}, robot, stop) }()
	select {
	case disconnect := <-done:
		return disconnect
	case <-time.After(gopodStuckWait()):
		gopodLogStuck(robot.Cfg.SerialNo, "action")
		return true
	}
}

func gopodWaitOrGiveUp(ch chan struct{}, esn, what string) {
	select {
	case <-ch:
	case <-time.After(gopodStuckWait()):
		gopodLogStuck(esn, what)
	}
}

func gopodReceiveOrGiveUp(ch chan bool, esn, what string) {
	select {
	case <-ch:
	case <-time.After(gopodStuckWait()):
		gopodLogStuck(esn, what)
	}
}

func gopodSendOrGiveUp(ch chan bool, esn, what string) {
	select {
	case ch <- true:
	case <-time.After(gopodStuckWait()):
		gopodLogStuck(esn, what)
	}
}

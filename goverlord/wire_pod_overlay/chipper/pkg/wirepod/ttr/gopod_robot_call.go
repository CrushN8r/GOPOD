// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_robot_call.go
package wirepod_ttr

import (
	"context"
	"sync"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// gopodRobotCall (2026-10-19, operator, after "system-robot say ..." answered
// "success" and nothing was heard: "for all robot calls add in a robot wake
// beforehand then continue * once a call is made, ensure the call is robust
// enough to ignore distractions until that call is complete (no overlapping
// chains)"): every Brobots-menu call to a robot goes through here - one at a
// time per Brobot (a second call waits its turn), the chat's own hold let go,
// the robot woken (behavior control taken and granted, then the 1.5 s settle a
// fresh control needs before its first action - CLAUDE.md), the call run to its
// end while control is held (his own behaviours can't cut in), then control
// given back and the chat's hold resumed. keep=false gives control back before
// the call (a Wire-Pod intent - pop a wheelie - needs the robot's own behaviours).
var gopodRobotCallLocks = struct {
	sync.Mutex
	m map[string]*sync.Mutex
}{m: map[string]*sync.Mutex{}}

func gopodRobotCallLock(esn string) *sync.Mutex {
	gopodRobotCallLocks.Lock()
	defer gopodRobotCallLocks.Unlock()
	if gopodRobotCallLocks.m[esn] == nil {
		gopodRobotCallLocks.m[esn] = &sync.Mutex{}
	}
	return gopodRobotCallLocks.m[esn]
}

func gopodRobotFor(esn string) *vector.Vector {
	for _, bot := range vars.BotInfo.Robots {
		if bot.Esn == esn {
			robot, err := vector.New(vector.WithSerialNo(esn), vector.WithToken(bot.GUID), vector.WithTarget(bot.IPAddress+":443"))
			if err == nil {
				return robot
			}
		}
	}
	return nil
}

func gopodRobotCall(esn string, keep bool, act func(robot *vector.Vector, ctx context.Context) string) string {
	lock := gopodRobotCallLock(esn)
	lock.Lock()
	defer lock.Unlock()
	robot := gopodRobotFor(esn)
	if robot == nil {
		return "not connected to Wire-Pod"
	}
	gopodChatHoldLetGo(esn)
	defer gopodChatHoldResume(esn)
	ctx := context.Background()
	granted := make(chan bool, 1)
	release := make(chan bool)
	gopodLogBControl(robot, ctx, granted, release)
	giveBack := func() {
		select {
		case release <- true:
		case <-time.After(2 * time.Second):
		}
	}
	select {
	case <-granted:
	case <-time.After(8 * time.Second):
		giveBack()
		return "didn't wake (no behavior control in 8 s - asleep on the charger, or off?)"
	}
	time.Sleep(1500 * time.Millisecond)
	if !keep {
		giveBack()
		return act(robot, ctx)
	}
	out := act(robot, ctx)
	giveBack()
	return out
}

// ## EOF

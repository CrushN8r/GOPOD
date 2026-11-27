// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_robot_call.go
package wirepod_ttr

import (
	"context"
	"sync"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

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
				return GopodRobotsOffConn(robot) // robots off (gopod_robots_off.go)
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
		return gopodSystemText("robot_call.not_connected_to_wire_pod")
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
		return gopodSystemText("robot_call.didn_t_wake_no_behavior")
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

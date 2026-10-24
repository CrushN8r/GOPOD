// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_arm_cue.go
package wirepod_ttr

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// One small arm cue per chat reply - lift up, back to rest - fired with the reply's
// first mid-reply animation (operator 2026-09-29: "a minimal single arm cue
// (lift/rest) during a playAnimationWI animation"). Same shape as the songs' golden
// arm cue (run_arm_cue / run_move_axis, run_robot_control_song_001.py: lift at
// +speed, hold, lift at -speed, hold, stop), one cycle, with each robot's own
// phcal-tuned speed and hold (phcal_last.json "arm": Brobot 1 speed 2 hold 1.2 s,
// Brobot 2 speed 3 hold 0.5 s). It always ends with the lift stopped at rest.
type gopodArmCueSettings struct {
	speed float32
	hold  time.Duration
}

var gopodArmCueByBrobot = map[string]gopodArmCueSettings{
	"Brobot 1": {speed: 2, hold: 1200 * time.Millisecond},
	"Brobot 2": {speed: 3, hold: 500 * time.Millisecond},
}

// gopodStartArmCue starts the cue in the background and returns a channel that
// closes when the arm is back at rest (nil if this robot has no cue).
func gopodStartArmCue(robot *vector.Vector, esn string) chan struct{} {
	cue, ok := gopodArmCueByBrobot[gopodBrobotName(esn)]
	if !ok {
		return nil
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		ctx := context.Background()
		name := gopodBrobotName(esn)
		lift := func(step string, speed float32) {
			logger.LogDebugUI(fmt.Sprintf("GOPOD_VECTOR_ARM: %s %s speed=%g hold=%dms", name, step, speed, cue.hold.Milliseconds()))
			if _, err := gopodLogMoveLift(robot, ctx, &vectorpb.MoveLiftRequest{SpeedRadPerSec: speed}); err != nil {
				logger.Println("GOPOD_VECTOR_ARM: " + name + " " + step + " failed: " + err.Error())
			}
		}
		lift("up", cue.speed)
		time.Sleep(cue.hold)
		lift("down", -cue.speed)
		time.Sleep(cue.hold)
		lift("rest", 0)
	}()
	return done
}

// 2026-10-04 (operator: "I notice arm lifts throughout responses - add nods to
// replace some arm lifts?"): every chat_moves.nod_every-th reply of a Brobot
// nods instead of lifting the arm - the songs' own nod (run_nod /
// run_move_axis, run_robot_control_song_001.py: head down, hold, head up, hold,
// stop), at chat_moves.nod_speed / nod_hold_seconds (phcal_last.json nod).
var gopodReplyMoveCount = struct {
	sync.Mutex
	n map[string]int
}{n: map[string]int{}}

// gopodStartReplyMove starts this reply's one move - the arm cue or a nod - and
// returns a channel that closes when it's back at rest.
func gopodStartReplyMove(robot *vector.Vector, esn string) chan struct{} {
	gopodReplyMoveCount.Lock()
	gopodReplyMoveCount.n[esn]++
	nod := gopodReplyMoveCount.n[esn]%gopodBrobotsSettings.ChatMoves.NodEvery == 0
	gopodReplyMoveCount.Unlock()
	if !nod {
		return gopodStartArmCue(robot, esn)
	}
	return gopodStartNod(robot, esn)
}

func gopodStartNod(robot *vector.Vector, esn string) chan struct{} {
	m := gopodBrobotsSettings.ChatMoves
	speed := float32(m.NodSpeed)
	hold := gopodSecondsToDuration(m.NodHoldSeconds)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ctx := context.Background()
		name := gopodBrobotName(esn)
		head := func(step string, speed float32) {
			logger.LogDebugUI(fmt.Sprintf("GOPOD_VECTOR_HEAD: %s %s speed=%g hold=%dms", name, step, speed, hold.Milliseconds()))
			if _, err := gopodLogMoveHead(robot, ctx, &vectorpb.MoveHeadRequest{SpeedRadPerSec: speed}); err != nil {
				logger.Println("GOPOD_VECTOR_HEAD: " + name + " " + step + " failed: " + err.Error())
			}
		}
		head("down", -speed)
		time.Sleep(hold)
		head("up", speed)
		time.Sleep(hold)
		head("rest", 0)
	}()
	return done
}

// gopodThinkingMoves: chat_moves.thinking_moves before "Thinking, please wait." -
// live and scripted answers both - each waiting for the one before it to be back
// at rest, so they never overlap.
func gopodThinkingMoves(robot *vector.Vector, esn string) {
	for _, m := range gopodGotItFor(esn).Moves {
		var done chan struct{}
		switch m {
		case "arm":
			done = gopodStartArmCue(robot, esn)
		case "nod":
			done = gopodStartNod(robot, esn)
		}
		if done != nil {
			<-done
		}
	}
}

// ## EOF

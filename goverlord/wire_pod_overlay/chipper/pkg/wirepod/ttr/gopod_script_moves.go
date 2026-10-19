// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_script_moves.go
package wirepod_ttr

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// Moves and live parts in a draft answer (operator 2026-10-08, the Awaken
// line: "I wake up, lift my arm (arm cue), nod (nod cue), test my cube, tell
// you the weather (... a live real-time weather check ... if online is
// detected), say I'm ready"):
//   - "moves": [{"at": "lift my arm", "move": "arm"}, ...] - when the robot
//     starts the spoken piece holding "at", it does the move: "arm" (his
//     phcal arm cue), "nod" (his chat nod) or "cube" (phcal's cube blip: all
//     corners red, 2 s, then green - on this robot's own paired cube).
//   - "{weather}" in "say" - the live weather, his own unit and clock first
//     (gopod_live_weather.go). Offline, the sentence holding it is left out.

type gopodDraftMove struct {
	At   string `json:"at"`
	Move string `json:"move"`
}

// gopodScriptFill puts the live weather in a draft line. Offline it says
// live_weather.offline_text, or leaves the sentence out when that's empty.
func gopodScriptFill(esn, say string) string {
	if !strings.Contains(say, "{weather}") {
		return say
	}
	// The golden weather call (gopodWeatherSay, gopod_live_weather.go).
	weather, live := gopodWeatherSay(esn)
	if live {
		logger.Println("GOPOD_SCRIPT live weather for " + gopodBrobotName(esn) + ": " + weather)
		return strings.ReplaceAll(say, "{weather}", weather)
	}
	// 2026-10-09 (operator: "it appears as though we're offline! Only thing left
	// to do is step outside then GOPOD Yourself!!"): offline_text stands in.
	if weather != "" {
		logger.Println("GOPOD_SCRIPT no live weather (offline?) - offline_text said")
		return strings.ReplaceAll(say, "{weather}", weather)
	}
	logger.Println("GOPOD_SCRIPT no live weather (offline?) - its sentence is left out")
	var kept []string
	// The weather line ends its own sentence once filled; it ends it here too.
	for _, s := range gopodSplitPieces(strings.ReplaceAll(say, "{weather}", "{weather}."), false) {
		if !strings.Contains(s, "{weather}") {
			kept = append(kept, s)
		}
	}
	return strings.Join(kept, " ")
}

// gopodScriptMoveAt does the moves due at this spoken piece, once each.
func gopodScriptMoveAt(robot *vector.Vector, esn, spoken string) {
	gopodScript.Lock()
	moves := gopodScript.moves[esn]
	var due []gopodDraftMove
	var left []gopodDraftMove
	lower := strings.ToLower(spoken)
	for _, m := range moves {
		if m.At != "" && strings.Contains(lower, strings.ToLower(m.At)) {
			due = append(due, m)
		} else {
			left = append(left, m)
		}
	}
	gopodScript.moves[esn] = left
	gopodScript.Unlock()
	for _, m := range due {
		logger.Println("GOPOD_SCRIPT move " + m.Move + " for " + gopodBrobotName(esn) + " at \"" + m.At + "\"")
		switch m.Move {
		case "arm":
			gopodStartArmCue(robot, esn)
		case "nod":
			gopodStartNod(robot, esn)
		case "cube":
			go gopodCubeBlip(robot, esn)
		default:
			logger.Println("GOPOD_SCRIPT unknown move: " + m.Move)
		}
	}
}

// gopodCubeBlip: phcal's cube test (direct_sdk_cube_blip_001) on the robot's
// own connection, which holds control for the answer - connect the cube, all
// four corners red, 2 s, then green.
func gopodCubeBlip(robot *vector.Vector, esn string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// 2026-10-08 (operator: "can you add a log for cube activity?"): each cube
	// call is logged going out and answered (GOPOD_ROBOT_CMD ... cube).
	sent := gopodRobotCmdSend(robot, "cube connect")
	resp, err := robot.Conn.ConnectCube(ctx, &vectorpb.ConnectCubeRequest{})
	what := "cube connect"
	if err == nil {
		what += fmt.Sprintf(" success=%t object=%d factory=%s", resp.GetSuccess(), resp.GetObjectId(), resp.GetFactoryId())
	}
	gopodRobotCmdDone(robot, what, sent, err)
	if err != nil || !resp.GetSuccess() {
		logger.Println("GOPOD_SCRIPT cube: no cube connected for " + gopodBrobotName(esn))
		return
	}
	corners := func(name string, color uint32) error {
		c := []uint32{color, color, color, color}
		p := []uint32{3000, 3000, 3000, 3000}
		z := []uint32{0, 0, 0, 0}
		sent := gopodRobotCmdSend(robot, "cube lights "+name)
		_, err := robot.Conn.SetCubeLights(ctx, &vectorpb.SetCubeLightsRequest{
			ObjectId: resp.GetObjectId(), OnColor: c, OffColor: c,
			OnPeriodMs: p, OffPeriodMs: z, TransitionOnPeriodMs: z, TransitionOffPeriodMs: z,
		})
		gopodRobotCmdDone(robot, "cube lights "+name, sent, err)
		return err
	}
	if err := corners("red", 0xff0000ff); err != nil {
		logger.Println("GOPOD_SCRIPT cube red: " + err.Error())
		return
	}
	time.Sleep(2 * time.Second)
	if err := corners("green", 0x00ff00ff); err != nil {
		logger.Println("GOPOD_SCRIPT cube green: " + err.Error())
		return
	}
	logger.Println("GOPOD_SCRIPT cube blip done for " + gopodBrobotName(esn))
}

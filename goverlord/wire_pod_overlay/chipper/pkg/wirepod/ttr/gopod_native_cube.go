// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_native_cube.go
package wirepod_ttr

import (
	"context"
	"strings"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// Typed and numpad cube commands run the robot's own native cube behavior
// (operator 2026-10-08: "backpack vocal cube command should respond as native
// 'cube' command by default"). A spoken one already does: Wire-Pod matches it
// to the intent (intent-data/en-US.json) and Vector runs it. A typed or numpad
// line has no voice request, so it would go to the chat; a line holding one of
// native_cube.intents' phrases is sent as that intent instead, the way
// /api-sdk/cloud_intent does (AppIntent). The chat's stay-in-place hold lets
// go of that Brobot first so his own behavior can run.

// gopodNativeCubeIntent: the native intent for a line, or "".
func gopodNativeCubeIntent(text string) string {
	lower := " " + strings.Join(strings.Fields(strings.ToLower(text)), " ") + " "
	for _, in := range gopodBrobotsSettings.NativeCube.Intents {
		for _, phrase := range in.Phrases {
			phrase = strings.TrimSpace(strings.ToLower(phrase))
			if phrase != "" && strings.Contains(lower, " "+phrase) {
				return in.Intent
			}
		}
	}
	return ""
}

// gopodRunNativeCube sends the line's cube intent to the Brobot, if it has one.
func gopodRunNativeCube(esn, text string) bool {
	intent := gopodNativeCubeIntent(text)
	if intent == "" {
		return false
	}
	var robot *vector.Vector
	for _, bot := range vars.BotInfo.Robots {
		if bot.Esn == esn {
			robot, _ = vector.New(vector.WithSerialNo(esn), vector.WithToken(bot.GUID), vector.WithTarget(bot.IPAddress+":443"))
			break
		}
	}
	if robot == nil {
		return false
	}
	logger.Println("Bot " + esn + " Native cube intent (typed): " + intent)
	go func() {
		gopodChatHoldLetGo(esn)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		sent := gopodRobotCmdSend(robot, "cube intent "+intent)
		_, err := robot.Conn.AppIntent(ctx, &vectorpb.AppIntentRequest{Intent: intent})
		gopodRobotCmdDone(robot, "cube intent "+intent, sent, err)
	}()
	return true
}

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_thinking_line.go
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

// What a Brobot says while the model works on a chat reply, the way go search says
// "Let me look that up." Operator's words, 2026-09-29: "Thinking, please wait...".
// Spoken only, no chat bubble: on the chat path Wire-Pod logs the operator's own
// question ("Intent matched") only once the reply is ready, so a bubble for this
// line would show above the question it answers.
// 2026-10-10: the words live in the settings file (chat_moves.thinking_text).
func gopodThinkingLineSpoken() string {
	if t := strings.TrimSpace(gopodBrobotsSettings.ChatMoves.ThinkingText); t != "" {
		return t
	}
	return "Thinking, please wait."
}

// gopodSayThinkingLine speaks the line and returns when the robot has said it.
// The robot reports how each line went (SayTextResponse.State). Live 2026-09-29 the
// line was sometimes not heard, so the robot's own report is logged, and the line
// is said once more only if the robot reports it didn't finish.
func gopodSayThinkingLine(robot *vector.Vector, esn string) {
	for attempt := 1; attempt <= 2; attempt++ {
		started := time.Now()
		resp, err := gopodLogSayText(robot, context.Background(), &vectorpb.SayTextRequest{
			DurationScalar: 1,
			UseVectorVoice: true,
			Text:           gopodThinkingLineSpoken(),
		})
		state := "no reply"
		if resp != nil {
			state = resp.GetState().String()
		}
		if err != nil {
			state += " (" + err.Error() + ")"
		}
		logger.Println(fmt.Sprintf("GOPOD_THINKING_LINE: %s attempt=%d state=%s took=%dms", gopodBrobotName(esn), attempt, state, time.Since(started).Milliseconds()))
		if err == nil && resp != nil && resp.GetState() == vectorpb.SayTextResponse_FINISHED {
			return
		}
	}
}

// GOLDEN, locked (2026-10-13, operator: "a single robust engine for live &
// scripted runs * snapshot the full \"thinking please wait\" process from live
// (+ arm, nod) * lock that snapshot * use locked \"thinking please wait\"
// process for scripted runs"): the one "thinking please wait" process - the
// live chat's (kgsim), a scripted answer's and a song's (/api-sdk/gopod_thinking)
// alike. With control already granted: a Brobot's thinking moves (arm, then
// nod; chat_moves.thinking_moves) and "Thinking, please wait."; then the
// searching clip until ready() and for at least searching_min_seconds, at the
// golden pace; then one searching get-out. Locked by TestThinkingProcessIsOne.
//
// 2026-10-16: afterLine (a live photo, gopod_photo.go) runs after the thinking
// line, before the searching moves.
func gopodThinkingProcess(robot *vector.Vector, ctx context.Context, esn string, brobot bool, ready func() bool, afterLine ...func()) {
	if brobot {
		// GOPOD 2026-10-09 (operator: "drop the buffer so every reply starts
		// the same way"): the 2026-10-04 searching + kgSuccess buffer while the
		// brother defers is gone - the arm and nod below give him the time.
		// 2026-10-09 (operator: "start with an arm lift + nod (as visual cues
		// to the audience) then the rest"): one after the other, each back at
		// rest first (gopod_arm_cue.go).
		gopodThinkingMoves(robot, esn)
		// 2026-10-08 (operator: "stay with live 1s minimum thinking/searching"):
		// said for a draft answer too.
		gopodSayThinkingLine(robot, esn)
	}
	for _, step := range afterLine {
		if step != nil {
			step()
		}
	}
	// GOPOD 2026-10-08 (operator: "a required minimum run to act as a buffer
	// against an early trigger, whether composing a response, or not" - an
	// addition, not a change: "A live answer that takes longer continues until
	// the response is ready like before"): searching every
	// searching_interval_seconds until the answer is ready, and for at least
	// searching_min_seconds, then one searchingGetout. The pace is the golden
	// one, so the robot isn't hit with commands: in the first
	// searching_min_seconds a clip every searching_interval_seconds (go search
	// V4's 1 s burst), then never faster than one clip,
	// searching_resend_seconds (the song lock,
	// SEARCHING_ANIMATION_SONG_LOCK_001.md) - a clip that comes back at once
	// (Brobot 2, ~80 ms, live 2026-10-08 12:02) no longer means another send
	// every ~0.4 s.
	cm := gopodBrobotsSettings.ChatMoves
	searchStart := time.Now()
	minRun := gopodSecondsToDuration(cm.SearchingMinSeconds)
	done := func() bool { return ready() && time.Since(searchStart) >= minRun }
	for !done() {
		sent := time.Now()
		gopodLogPlayAnimation(robot, ctx, &vectorpb.PlayAnimationRequest{
			Animation: &vectorpb.Animation{Name: "anim_knowledgegraph_searching_01"},
			Loops:     1,
		})
		gap := gopodSecondsToDuration(cm.SearchingIntervalSeconds)
		if time.Since(searchStart) >= minRun {
			gap = gopodSecondsToDuration(cm.SearchingResendSeconds)
		}
		for !done() && time.Since(sent) < gap {
			time.Sleep(50 * time.Millisecond)
		}
	}
	gopodLogPlayAnimation(robot, ctx, &vectorpb.PlayAnimationRequest{
		Animation: &vectorpb.Animation{Name: "anim_knowledgegraph_searching_getout_01"},
		Loops:     1,
	})
}

// GopodThinkingProcess: the process for a song's Brobot (/api-sdk/gopod_thinking),
// on Wire-Pod's own connection the song already holds control through; returns
// once the get-out has played.
func GopodThinkingProcess(robot *vector.Vector, ctx context.Context, esn string) {
	logger.Println("GOPOD_THINKING_PROCESS " + gopodBrobotName(esn) + " (song)")
	gopodThinkingProcess(robot, ctx, esn, true, func() bool { return true })
}

// ## EOF

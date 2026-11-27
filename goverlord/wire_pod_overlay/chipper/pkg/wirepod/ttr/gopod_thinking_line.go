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

// What a Brobot says while the model works on a chat reply, the way go search says "Let me look
// that up." Operator's words, "Thinking, please wait...".
func gopodThinkingLineSpoken() string {
	if t := strings.TrimSpace(gopodBrobotsSettings.ChatMoves.ThinkingText); t != "" {
		return t
	}
	return gopodSystemText("thinking_line.thinking_please_wait")
}

// gopodGotIt: one Brobot's copy of the golden "Got it" sequence - his words and his moves before
// them.
type gopodGotIt struct {
	Name  string   `json:"name"`
	Text  string   `json:"text"`
	Moves []string `json:"moves"`
}

// gopodGotItFor: his own copy (chat_moves.got_it), each part not set taken from the shared golden
// values.
func gopodGotItFor(esn string) gopodGotIt { return gopodGotItNamed(gopodBrobotName(esn)) }

// gopodGotItNamed: anyone's Got it, by name - a Brobot's or a persona's (chat_moves.got_it.<name>,
// in its own file); the golden words when it has none.
func gopodGotItNamed(name string) gopodGotIt {
	cm := gopodBrobotsSettings.ChatMoves
	g := cm.GotIt[name]
	if strings.TrimSpace(g.Text) == "" {
		g.Text = gopodThinkingLineSpoken()
	}
	if g.Moves == nil {
		g.Moves = cm.ThinkingMoves
	}
	return g
}

// gopodSayThinkingLine speaks the line and returns when the robot has said it.
func gopodSayThinkingLine(robot *vector.Vector, esn string) {
	gopodSayThinkingWords(robot, esn)
	gopodOpenerShow(esn)
	gopodValuePointsNoticeGotIt(esn)
}

// gopodSayThinkingWords: the line said (once more if the robot says it didn't finish).
func gopodSayThinkingWords(robot *vector.Vector, esn string) {
	gopodSayWords(robot, esn, gopodOpenerPick(esn))
}

// gopodSayWords: a short line of his, said the thinking line's way.
func gopodSayWords(robot *vector.Vector, esn, text string) {
	for attempt := 1; attempt <= 2; attempt++ {
		started := time.Now()
		resp, err := gopodLogSayText(robot, context.Background(), &vectorpb.SayTextRequest{
			DurationScalar: 1,
			UseVectorVoice: true,
			Text:           gopodReplaceWholeWords(text, gopodBrobotPronunciations()),
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

// GOLDEN, locked: the one "thinking please wait" process - the live chat's (kgsim), a scripted
// answer's and a song's (/api-sdk/gopod_thinking) alike.
func gopodThinkingProcess(robot *vector.Vector, ctx context.Context, esn string, brobot bool, ready func() bool, afterLine ...func()) {
	gopodSetBrobotEyes(robot, esn, true)
	if brobot {
		gopodThinkingMoves(robot, esn)
		gopodSayThinkingLine(robot, esn)
	}
	for _, step := range afterLine {
		if step != nil {
			step()
		}
	}
	sg := gopodSearchingGolden()
	searchStart := time.Now()
	minRun := gopodSecondsToDuration(sg.MinSeconds)
	done := func() bool { return ready() && time.Since(searchStart) >= minRun }
	paused, hadManager := false, false
	for !done() || gopodManagerBusy.Load() {
		if gopodManagerBusy.Load() {
			if !paused {
				paused, hadManager = true, true
				logger.Println("GOPOD_THINKING " + gopodBrobotName(esn) + " searching paused (a manager has the floor)")
				gopodSearchingPlay(robot, ctx, sg, sg.Getout)
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if paused {
			paused = false
			logger.Println("GOPOD_THINKING " + gopodBrobotName(esn) + " searching again")
		}
		sent := time.Now()
		gopodSearchingPlay(robot, ctx, sg, sg.Clip)
		gap := gopodSecondsToDuration(sg.FloorSeconds)
		if !sg.OK {
			gap = 100 * time.Millisecond // no note: no clips, just the wait
		}
		for !done() && !gopodManagerBusy.Load() && time.Since(sent) < gap {
			time.Sleep(50 * time.Millisecond)
		}
	}
	gopodSearchingPlay(robot, ctx, sg, sg.Getout)
	if brobot {
		if line := gopodNoticeGotItAgain(esn, hadManager); line != "" {
			gopodSayWords(robot, esn, line)
			gopodNoticeSaid(esn)
		}
	}
}

// GopodThinkingProcess: the process for a song's Brobot (/api-sdk/gopod_thinking), on Wire-Pod's
// own connection the song already holds control through; returns once the get-out has played.
func GopodThinkingProcess(robot *vector.Vector, ctx context.Context, esn string) {
	logger.Println("GOPOD_THINKING_PROCESS " + gopodBrobotName(esn) + " (song)")
	gopodThinkingProcess(robot, ctx, esn, true, func() bool { return true })
}

// ## EOF

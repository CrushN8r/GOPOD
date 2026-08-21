package wirepod_ttr

import (
	"fmt"
	"strings"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/sashabaranov/go-openai"
)

func promptUsesCanonicalBrobotRawShape(prompt string) bool {
	if strings.Contains(prompt, "{{[emotion]}}") &&
		(strings.Contains(prompt, gopodCanonicalStartMarker) ||
			strings.Contains(prompt, gopodCanonicalChunkMarker) ||
			strings.Contains(prompt, gopodCanonicalEndMarker)) {
		return true
	}
	return strings.Contains(prompt, "actionParameter...") &&
		strings.Contains(prompt, "Valid action parameters:")
}

func promptIsBrobotPrompt(prompt string) bool {
	return strings.Contains(prompt, "Bro! Bots:") ||
		(strings.Contains(prompt, "Brobot") && strings.Contains(prompt, "GOPOD Yourself"))
}

func robotSpeechFromActions(actions []RobotAction) string {
	var speech []string
	for _, action := range actions {
		if action.Action != ActionSayText {
			continue
		}
		text := cleanRobotSpeechText(action.Parameter)
		if text != "" {
			speech = append(speech, text)
		}
	}
	return strings.Join(speech, " ")
}

func stripSpokenActionLead(input string) string {
	cleaned := strings.TrimSpace(input)
	for {
		before := cleaned
		lower := strings.ToLower(cleaned)
		for _, prefix := range []string{
			"actionparameter:",
			"actionparameter",
			"spoken thought:",
			"spoken thought",
		} {
			if strings.HasPrefix(lower, prefix) {
				cleaned = strings.TrimSpace(cleaned[len(prefix):])
				lower = strings.ToLower(cleaned)
			}
		}
		for _, animThing := range animationMap {
			action := animThing[0]
			if strings.HasPrefix(cleaned, action+" ") {
				cleaned = strings.TrimSpace(strings.TrimPrefix(cleaned, action))
				break
			}
			if strings.HasPrefix(cleaned, action+".") {
				cleaned = strings.TrimSpace(strings.TrimPrefix(cleaned, action+"."))
				break
			}
			if strings.HasPrefix(cleaned, action+":") {
				cleaned = strings.TrimSpace(strings.TrimPrefix(cleaned, action+":"))
				break
			}
		}
		if cleaned == before {
			return cleaned
		}
	}
}

func cleanRobotSpeechText(input string) string {
	cleaned := normalizeBrobotRobotCopy(input)
	cleaned = ansiEscapePattern.ReplaceAllString(cleaned, "")
	cleaned = stripCanonicalGOPODMarkers(cleaned)
	cleaned = strings.ReplaceAll(cleaned, "|||", " ")
	if strings.Contains(cleaned, "{{") {
		cleaned = robotSpeechFromActions(GetActionsFromString(cleaned))
	}
	cleaned = strings.ReplaceAll(cleaned, "playAnimationWI", "")
	cleaned = strings.TrimSpace(cleaned)
	for {
		next := strings.TrimSpace(strings.TrimPrefix(cleaned, "..."))
		next = strings.TrimSpace(strings.TrimPrefix(next, "."))
		if next == cleaned {
			break
		}
		cleaned = next
	}
	if units := parseBrobotPhraseUnits(cleaned); len(units) == 1 {
		cleaned = units[0].Text
	}
	cleaned = strings.TrimSpace(cleaned)
	cleaned = stripSpokenActionLead(cleaned)
	cleaned = strings.ReplaceAll(cleaned, "...", ".")
	for {
		next := strings.TrimSpace(strings.TrimPrefix(cleaned, "..."))
		next = strings.TrimSpace(strings.TrimPrefix(next, "."))
		if next == cleaned {
			break
		}
		cleaned = next
	}
	cleaned = stripSpokenActionLead(cleaned)
	return strings.Join(strings.Fields(cleaned), " ")
}

func connectVector(esn string) (*vector.Vector, error) {
	for _, bot := range vars.BotInfo.Robots {
		if esn == bot.Esn {
			return vector.New(
				vector.WithSerialNo(esn),
				vector.WithToken(bot.GUID),
				vector.WithTarget(bot.IPAddress+":443"),
			)
		}
	}
	return nil, fmt.Errorf("robot ESN not found: %s", esn)
}

type liveRobotActionDispatcher struct {
	robot *vector.Vector
}

func (d liveRobotActionDispatcher) PerformActions(msgs []openai.ChatCompletionMessage, actions []RobotAction, stopStop chan bool) bool {
	return performActionsLiveDispatchUnsafe(msgs, actions, d.robot, stopStop)
}

func performActionsLiveDispatchUnsafe(msgs []openai.ChatCompletionMessage, actions []RobotAction, robot *vector.Vector, stopStop chan bool) bool {
	stopPerforming := false
	go func() {
		for range stopStop {
			stopPerforming = true
		}
	}()
	for _, action := range actions {
		if stopPerforming {
			return false
		}
		switch {
		case action.Action == ActionSayText:
			DoSayText(action.Parameter, robot)
		case action.Action == ActionPlayAnimation:
			DoPlayAnimation(action.Parameter, robot)
		case action.Action == ActionPlayAnimationWI:
			DoPlayAnimationWI(action.Parameter, robot)
		case action.Action == ActionNewRequest:
			go DoNewRequest(robot)
			return true
		case action.Action == ActionGetImage:
			DoGetImage(msgs, action.Parameter, robot, stopStop)
			return true
		case action.Action == ActionPlaySound:
			DoPlaySound(action.Parameter, robot)
		}
	}
	WaitForAnim_Queue(robot.Cfg.SerialNo)
	return false
}

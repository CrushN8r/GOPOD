package wirepod_ttr

import (
	"os"

	"github.com/sashabaranov/go-openai"
)

const gopodAllowLiveRobotSpeechEnv = "GOPOD_ALLOW_LIVE_ROBOT_SPEECH"

type GOPODLiveSpeechGateResult struct {
	GateBlocked              bool `json:"gate_blocked"`
	GateEnabled              bool `json:"gate_enabled"`
	DispatchReachable        bool `json:"dispatch_reachable"`
	PerformActionsCalled     bool `json:"perform_actions_called"`
	DoSayTextCalled          bool `json:"do_say_text_called"`
	DoPlayAnimationWICalled  bool `json:"do_play_animation_wi_called"`
	RobotConnectionAttempted bool `json:"robot_connection_attempted"`
	DispatchReturn           bool `json:"dispatch_return"`
}

type gopodRobotActionDispatcher interface {
	PerformActions(msgs []openai.ChatCompletionMessage, actions []RobotAction, stopStop chan bool) bool
}

func gopodLiveRobotSpeechGateEnabled() bool {
	return os.Getenv(gopodAllowLiveRobotSpeechEnv) == "1"
}

func performActionsWithGOPODLiveSpeechGate(msgs []openai.ChatCompletionMessage, actions []RobotAction, stopStop chan bool, dispatcher gopodRobotActionDispatcher) GOPODLiveSpeechGateResult {
	result := GOPODLiveSpeechGateResult{}
	if !gopodLiveRobotSpeechGateEnabled() {
		result.GateBlocked = true
		return result
	}

	result.GateEnabled = true
	result.DispatchReachable = true
	result.PerformActionsCalled = true
	for _, action := range actions {
		switch action.Action {
		case ActionSayText:
			result.DoSayTextCalled = true
		case ActionPlayAnimationWI:
			result.DoPlayAnimationWICalled = true
		}
	}
	result.DispatchReturn = dispatcher.PerformActions(msgs, actions, stopStop)
	return result
}

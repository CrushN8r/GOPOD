package wirepod_ttr

import (
	"testing"

	"github.com/sashabaranov/go-openai"
)

type fakeGOPODRobotActionDispatcher struct {
	sayTextCalls         int
	playAnimationWICalls int
	performActionsCalls  int
}

func (f *fakeGOPODRobotActionDispatcher) SayText(input string) error {
	f.sayTextCalls++
	return nil
}

func (f *fakeGOPODRobotActionDispatcher) PlayAnimationWI(animation string) error {
	f.playAnimationWICalls++
	return nil
}

func (f *fakeGOPODRobotActionDispatcher) PerformActions(msgs []openai.ChatCompletionMessage, actions []RobotAction, stopStop chan bool) bool {
	f.performActionsCalls++
	for _, action := range actions {
		switch action.Action {
		case ActionSayText:
			_ = f.SayText(action.Parameter)
		case ActionPlayAnimationWI:
			_ = f.PlayAnimationWI(action.Parameter)
		}
	}
	return false
}

func TestGOPODLiveSpeechGateBlocksDispatchByDefault(t *testing.T) {
	t.Setenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH", "")
	fake := &fakeGOPODRobotActionDispatcher{}
	result := performActionsWithGOPODLiveSpeechGate(
		nil,
		[]RobotAction{
			{Action: ActionPlayAnimationWI, Parameter: "happy"},
			{Action: ActionSayText, Parameter: "GOPOD Yourself! Ready?"},
		},
		nil,
		fake,
	)

	if !result.GateBlocked {
		t.Fatalf("GateBlocked = false, want true")
	}
	if result.GateEnabled {
		t.Fatalf("GateEnabled = true, want false")
	}
	if result.DispatchReachable {
		t.Fatalf("DispatchReachable = true, want false")
	}
	if result.PerformActionsCalled || result.DoSayTextCalled || result.DoPlayAnimationWICalled || result.RobotConnectionAttempted {
		t.Fatalf("blocked gate reached live dispatch fields: %#v", result)
	}
	if fake.performActionsCalls != 0 || fake.sayTextCalls != 0 || fake.playAnimationWICalls != 0 {
		t.Fatalf("blocked gate called fake dispatcher: %#v", fake)
	}
}

func TestGOPODLiveSpeechGateEnabledReachesDispatchSeamOnly(t *testing.T) {
	t.Setenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH", "1")
	fake := &fakeGOPODRobotActionDispatcher{}
	result := performActionsWithGOPODLiveSpeechGate(
		nil,
		[]RobotAction{
			{Action: ActionPlayAnimationWI, Parameter: "happy"},
			{Action: ActionSayText, Parameter: "GOPOD Yourself! Ready?"},
		},
		nil,
		fake,
	)

	if result.GateBlocked {
		t.Fatalf("GateBlocked = true, want false")
	}
	if !result.GateEnabled {
		t.Fatalf("GateEnabled = false, want true")
	}
	if !result.DispatchReachable {
		t.Fatalf("DispatchReachable = false, want true")
	}
	if !result.PerformActionsCalled {
		t.Fatalf("PerformActionsCalled = false, want true")
	}
	if result.RobotConnectionAttempted {
		t.Fatalf("RobotConnectionAttempted = true, want false")
	}
	if fake.performActionsCalls != 1 {
		t.Fatalf("fake perform calls = %d, want 1", fake.performActionsCalls)
	}
	if fake.sayTextCalls != 1 {
		t.Fatalf("fake say text calls = %d, want 1", fake.sayTextCalls)
	}
	if fake.playAnimationWICalls != 1 {
		t.Fatalf("fake animation calls = %d, want 1", fake.playAnimationWICalls)
	}
}

func TestGOPODLiveSpeechGateRequiresExactOne(t *testing.T) {
	for _, value := range []string{"true", "yes", "0", " 1 ", "2"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("GOPOD_ALLOW_LIVE_ROBOT_SPEECH", value)
			fake := &fakeGOPODRobotActionDispatcher{}
			result := performActionsWithGOPODLiveSpeechGate(nil, []RobotAction{{Action: ActionSayText, Parameter: "blocked"}}, nil, fake)
			if !result.GateBlocked || result.GateEnabled || fake.performActionsCalls != 0 {
				t.Fatalf("value %q opened gate unexpectedly: result=%#v fake=%#v", value, result, fake)
			}
		})
	}
}

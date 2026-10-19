package wirepod_ttr

import (
	"strings"
	"testing"
)

func TestEmotionalBeatActionPlanOneBeatCreatesTwoActions(t *testing.T) {
	got := CompileRobotSafeEmotionalBeatsDry("happy... Hello there.")
	assertDryBoundaryFlags(t, got)
	assertDryActionPlan(t, got, []RobotAction{
		{Action: ActionPlayAnimationWI, Parameter: "happy"},
		{Action: ActionSayText, Parameter: "Hello there."},
	})
	if len(got.Beats) != 1 {
		t.Fatalf("len(beats) = %d, want 1", len(got.Beats))
	}
}

func TestEmotionalBeatActionPlanMultipleBeatsCreateOrderedPairs(t *testing.T) {
	got := CompileRobotSafeEmotionalBeatsDry(strings.Join([]string{
		"happy... First beat.",
		"thinking... Second beat. Still one payload.",
		"love... Third beat.",
	}, "\n"))
	assertDryBoundaryFlags(t, got)
	assertDryActionPlan(t, got, []RobotAction{
		{Action: ActionPlayAnimationWI, Parameter: "happy"},
		{Action: ActionSayText, Parameter: "First beat."},
		{Action: ActionPlayAnimationWI, Parameter: "thinking"},
		{Action: ActionSayText, Parameter: "Second beat. Still one payload."},
		{Action: ActionPlayAnimationWI, Parameter: "love"},
		{Action: ActionSayText, Parameter: "Third beat."},
	})
}

func TestEmotionalBeatActionPlanMiniPassagesRemainOneSayTextPerBeat(t *testing.T) {
	got := CompileEmotionalBeatActionPlanDry("confused... First sentence. Second sentence? Third sentence!")
	assertDryActionPlan(t, got, []RobotAction{
		{Action: ActionPlayAnimationWI, Parameter: "confused"},
		{Action: ActionSayText, Parameter: "First sentence. Second sentence? Third sentence!"},
	})
	if len(got.ActionPlan) != 2 {
		t.Fatalf("len(actionPlan) = %d, want 2", len(got.ActionPlan))
	}
}

func TestEmotionalBeatActionPlanEmptyOrInvalidLineHandledWithoutPanic(t *testing.T) {
	got := CompileRobotSafeEmotionalBeatsDry("\n   \nthis is not a beat\n")
	assertDryBoundaryFlags(t, got)
	assertDryActionPlan(t, got, []RobotAction{
		{Action: ActionPlayAnimationWI, Parameter: "thinking"},
		{Action: ActionSayText, Parameter: gopodRobotSpeechFallbackLine},
	})
	assertWarningContains(t, got.Warnings, "empty_emotional_beat_line_skipped")
	assertRejectReasonContains(t, got.RejectedReasons, "invalid_action_parameter")
}

func TestEmotionalBeatActionPlanInvalidParameterCarriesFallbackMetadata(t *testing.T) {
	got := CompileRobotSafeEmotionalBeatsDry("spin... Hello there.")
	assertDryActionPlan(t, got, []RobotAction{
		{Action: ActionPlayAnimationWI, Parameter: "thinking"},
		{Action: ActionSayText, Parameter: "Hello there."},
	})
	assertRejectReasonContains(t, got.RejectedReasons, "invalid_action_parameter")
	assertRepairPathContains(t, got.RepairPath, "invalid_action_parameter_fallback:spin->thinking")
}

func TestEmotionalBeatActionPlanBlocksTerminalBranches(t *testing.T) {
	for _, terminal := range []string{"getImage", "newVoiceRequest"} {
		t.Run(terminal, func(t *testing.T) {
			got := CompileRobotSafeEmotionalBeatsDry(terminal + "... stay dry")
			assertDryBoundaryFlags(t, got)
			assertRejectReasonContains(t, got.RejectedReasons, "terminal_action_parameter_rejected")
			assertNoTerminalActions(t, got.ActionPlan)
			assertDryActionPlan(t, got, []RobotAction{
				{Action: ActionPlayAnimationWI, Parameter: "thinking"},
				{Action: ActionSayText, Parameter: "stay dry"},
			})
		})
	}
}

func TestEmotionalBeatActionPlanBlocksRawCommandSyntax(t *testing.T) {
	got := CompileRobotSafeEmotionalBeatsDry("happy... {{playAnimationWI||love}} Hello {{getImage||front}}")
	assertDryBoundaryFlags(t, got)
	assertDryActionPlan(t, got, []RobotAction{
		{Action: ActionPlayAnimationWI, Parameter: "happy"},
		{Action: ActionSayText, Parameter: "Hello"},
	})
	assertRejectReasonContains(t, got.RejectedReasons, "raw_command_syntax_detected")
	assertRepairPathContains(t, got.RepairPath, "raw_command_syntax_removed")
	for _, action := range got.ActionPlan {
		if strings.Contains(action.Parameter, "{{") || strings.Contains(action.Parameter, "}}") {
			t.Fatalf("dry action plan leaked raw command syntax: %#v", got.ActionPlan)
		}
	}
}

func TestEmotionalBeatActionPlanDoesNotEmitTerminalActions(t *testing.T) {
	got := CompileRobotSafeEmotionalBeatsDry(strings.Join([]string{
		"getImage... camera stays blocked",
		"newVoiceRequest... new request stays blocked",
		"happy... normal line",
	}, "\n"))
	assertNoTerminalActions(t, got.ActionPlan)
}

func TestEmotionalBeatActionPlanOldCommandParserAndDispatcherStillAvailable(t *testing.T) {
	actions := GetActionsFromString("plain speech")
	if len(actions) != 1 || actions[0].Action != ActionSayText || actions[0].Parameter != "plain speech" {
		t.Fatalf("GetActionsFromString plain speech = %#v", actions)
	}
	if got := CompileRobotSafeEmotionalBeatsDry("happy... dry only"); got.PerformActionsCalled {
		t.Fatalf("PerformActionsCalled = true, want false")
	}
}

func TestEmotionalBeatActionPlanSilentCueMetadataDoesNotEnterActions(t *testing.T) {
	visibleBeat := "happy... GOPOD turns attention into a useful little moment."
	silentCue := "Bottom line to reveal: this silent cue must stay metadata only."
	got := CompileRobotSafeEmotionalBeatsDry(visibleBeat)

	assertDryBoundaryFlags(t, got)
	assertDryActionPlan(t, got, []RobotAction{
		{Action: ActionPlayAnimationWI, Parameter: "happy"},
		{Action: ActionSayText, Parameter: "GOPOD turns attention into a useful little moment."},
	})
	for _, action := range got.ActionPlan {
		if strings.Contains(action.Parameter, silentCue) || strings.Contains(action.Parameter, "Bottom line to reveal") || strings.Contains(action.Parameter, "silent cue") {
			t.Fatalf("silent cue leaked into action parameter: %#v", got.ActionPlan)
		}
	}
}

func TestEmotionalBeatActionPlanDeterministicIndependentOfExternalSilentCues(t *testing.T) {
	visibleBeat := "thinking... Same visible beat creates the same dry action plan."
	runWithMetadata := func(_ string) EmotionalBeatActionPlanResult {
		return CompileRobotSafeEmotionalBeatsDry(visibleBeat)
	}

	first := runWithMetadata("Bottom line to reveal: cue A")
	second := runWithMetadata("Bottom line to reveal: cue B")

	assertDryBoundaryFlags(t, first)
	assertDryBoundaryFlags(t, second)
	if len(first.ActionPlan) != len(second.ActionPlan) {
		t.Fatalf("action plan lengths differ: %#v vs %#v", first.ActionPlan, second.ActionPlan)
	}
	for i := range first.ActionPlan {
		if first.ActionPlan[i] != second.ActionPlan[i] {
			t.Fatalf("action plan changed with silent metadata: first=%#v second=%#v", first.ActionPlan, second.ActionPlan)
		}
	}
	if first.PerformActionsCalled || second.PerformActionsCalled {
		t.Fatalf("dry action plan crossed dispatch boundary: first=%#v second=%#v", first, second)
	}
}

func assertDryActionPlan(t *testing.T, got EmotionalBeatActionPlanResult, want []RobotAction) {
	t.Helper()
	if len(got.ActionPlan) != len(want) {
		t.Fatalf("len(actionPlan) = %d, want %d: %#v", len(got.ActionPlan), len(want), got.ActionPlan)
	}
	for i := range want {
		if got.ActionPlan[i] != want[i] {
			t.Fatalf("actionPlan[%d] = %#v, want %#v; full plan %#v", i, got.ActionPlan[i], want[i], got.ActionPlan)
		}
	}
	assertNoTerminalActions(t, got.ActionPlan)
}

func assertDryBoundaryFlags(t *testing.T, got EmotionalBeatActionPlanResult) {
	t.Helper()
	if got.LiveRobotSpeech {
		t.Fatalf("LiveRobotSpeech = true, want false")
	}
	if got.RobotCalls {
		t.Fatalf("RobotCalls = true, want false")
	}
	if got.CameraCalls {
		t.Fatalf("CameraCalls = true, want false")
	}
	if got.VisionAPICalls {
		t.Fatalf("VisionAPICalls = true, want false")
	}
	if got.PerformActionsCalled {
		t.Fatalf("PerformActionsCalled = true, want false")
	}
}

func assertNoTerminalActions(t *testing.T, actions []RobotAction) {
	t.Helper()
	for _, action := range actions {
		if action.Action == ActionGetImage {
			t.Fatalf("unexpected ActionGetImage in dry action plan: %#v", actions)
		}
		if action.Action == ActionNewRequest {
			t.Fatalf("unexpected ActionNewRequest in dry action plan: %#v", actions)
		}
	}
}

func assertWarningContains(t *testing.T, warnings []string, want string) {
	t.Helper()
	for _, got := range warnings {
		if got == want {
			return
		}
	}
	t.Fatalf("warnings %v missing %q", warnings, want)
}

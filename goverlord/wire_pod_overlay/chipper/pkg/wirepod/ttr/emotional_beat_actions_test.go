package wirepod_ttr

import (
	"strings"
	"testing"
)

func TestEmotionalBeatEveryValidActionParameterCompiles(t *testing.T) {
	validParams := []string{
		"happy",
		"veryHappy",
		"sad",
		"verySad",
		"angry",
		"frustrated",
		"confused",
		"thinking",
		"celebrate",
		"love",
	}

	for _, param := range validParams {
		t.Run(param, func(t *testing.T) {
			got := CompileEmotionalBeatActions(param + "... Hello there. This is still one spoken beat.")
			assertEmotionalBeatActions(t, got, param, "Hello there. This is still one spoken beat.")
			if got.Packet.FallbackUsed {
				t.Fatalf("FallbackUsed = true for valid param %q", param)
			}
		})
	}
}

func TestEmotionalBeatActiveAliasNormalizationCompiles(t *testing.T) {
	tests := []struct {
		line      string
		animation string
		speech    string
		repair    string
	}{
		{"excited... Hello there.", "veryHappy", "Hello there.", "animation_param_normalized:excited->veryHappy"},
		{"very happy... Hello there.", "veryHappy", "Hello there.", "animation_param_normalized:very happy->veryHappy"},
		{"curious... Hello there.", "thinking", "Hello there.", "animation_param_normalized:curious->thinking"},
		{"awkward... Hello there.", "confused", "Hello there.", "animation_param_normalized:awkward->confused"},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := CompileEmotionalBeatActions(tt.line)
			assertEmotionalBeatActions(t, got, tt.animation, tt.speech)
			assertRepairPathContains(t, got.Packet.RepairPath, tt.repair)
		})
	}
}

func TestEmotionalBeatInvalidParameterFallback(t *testing.T) {
	got := CompileEmotionalBeatActions("spin... Hello there.")
	assertEmotionalBeatActions(t, got, "thinking", "Hello there.")
	if !got.Packet.FallbackUsed {
		t.Fatalf("FallbackUsed = false, want true")
	}
	assertRejectReasonContains(t, got.Packet.RejectedReasons, "invalid_action_parameter")
	assertRepairPathContains(t, got.Packet.RepairPath, "invalid_action_parameter_fallback:spin->thinking")
}

func TestEmotionalBeatEmptySpokenTextFallback(t *testing.T) {
	got := CompileEmotionalBeatActions("thinking...")
	assertEmotionalBeatActions(t, got, "thinking", gopodRobotSpeechFallbackLine)
	if !got.Packet.FallbackUsed {
		t.Fatalf("FallbackUsed = false, want true")
	}
	assertRepairPathContains(t, got.Packet.RepairPath, "empty_spoken_text_fallback")
}

func TestEmotionalBeatSplitsOnlyOnFirstDelimiter(t *testing.T) {
	got := CompileEmotionalBeatActions("confused... First clause... second clause remains spoken.")
	assertEmotionalBeatActions(t, got, "confused", "First clause, second clause remains spoken.")
}

func TestEmotionalBeatMultiSentenceRemainsOneSayTextPayload(t *testing.T) {
	got := CompileEmotionalBeatActions("love... First sentence. Second sentence? Third sentence!")
	assertEmotionalBeatActions(t, got, "love", "First sentence. Second sentence? Third sentence!")
	if len(got.Actions) != 2 {
		t.Fatalf("len(actions) = %d, want 2", len(got.Actions))
	}
}

func TestEmotionalBeatNoCommandSyntaxLeakage(t *testing.T) {
	got := CompileEmotionalBeatActions("happy... {{playAnimationWI||love}} Hello {{getImage||front}}")
	assertEmotionalBeatActions(t, got, "happy", "Hello")
	if strings.Contains(got.Actions[1].Parameter, "{{") || strings.Contains(got.Actions[1].Parameter, "}}") {
		t.Fatalf("ActionSayText leaked command syntax: %q", got.Actions[1].Parameter)
	}
	assertRejectReasonContains(t, got.Packet.RejectedReasons, "raw_command_syntax_detected")
	assertRepairPathContains(t, got.Packet.RepairPath, "raw_command_syntax_removed")
}

func TestEmotionalBeatRawCommandSyntaxInActionParameterDetected(t *testing.T) {
	got := CompileEmotionalBeatActions("{{getImage||front}}... Hello there.")
	assertEmotionalBeatActions(t, got, "thinking", "Hello there.")
	assertRejectReasonContains(t, got.Packet.RejectedReasons, "raw_command_syntax_detected")
	assertRejectReasonContains(t, got.Packet.RejectedReasons, "invalid_action_parameter")
	for _, action := range got.Actions {
		if action.Action == ActionGetImage || action.Action == ActionNewRequest {
			t.Fatalf("terminal action emitted for command syntax param: %#v", got.Actions)
		}
	}
}

func TestEmotionalBeatCombinedFailureMetadataPreservesReasons(t *testing.T) {
	got := CompileEmotionalBeatActions("{{getImage||front}}... {{newVoiceRequest||now}}")
	assertEmotionalBeatActions(t, got, "thinking", gopodRobotSpeechFallbackLine)
	assertRejectReasonContains(t, got.Packet.RejectedReasons, "raw_command_syntax_detected")
	assertRejectReasonContains(t, got.Packet.RejectedReasons, "invalid_action_parameter")
	assertRepairPathContains(t, got.Packet.RepairPath, "raw_command_syntax_removed")
	assertRepairPathContains(t, got.Packet.RepairPath, "empty_spoken_text_fallback")
}

func TestEmotionalBeatTerminalNamesRejectedNotEmitted(t *testing.T) {
	for _, terminal := range []string{"getImage", "newVoiceRequest"} {
		t.Run(terminal, func(t *testing.T) {
			got := CompileEmotionalBeatActions(terminal + "... should not become terminal action")
			assertEmotionalBeatActions(t, got, "thinking", "should not become terminal action")
			assertRejectReasonContains(t, got.Packet.RejectedReasons, "terminal_action_parameter_rejected")
			for _, action := range got.Actions {
				if action.Action == ActionGetImage {
					t.Fatalf("emitted ActionGetImage for %q", terminal)
				}
				if action.Action == ActionNewRequest {
					t.Fatalf("emitted ActionNewRequest for %q", terminal)
				}
			}
		})
	}
}

func assertEmotionalBeatActions(t *testing.T, got EmotionalBeatCompileResult, animation string, speech string) {
	t.Helper()
	if !got.FormatPass {
		t.Fatalf("FormatPass = false, packet = %#v", got.Packet)
	}
	if got.Packet.Animation != animation {
		t.Fatalf("Animation = %q, want %q", got.Packet.Animation, animation)
	}
	if got.Packet.Speech != speech {
		t.Fatalf("Speech = %q, want %q", got.Packet.Speech, speech)
	}
	if len(got.Actions) != 2 {
		t.Fatalf("len(actions) = %d, want 2: %#v", len(got.Actions), got.Actions)
	}
	if got.Actions[0].Action != ActionPlayAnimationWI || got.Actions[0].Parameter != animation {
		t.Fatalf("first action = %#v, want ActionPlayAnimationWI(%q)", got.Actions[0], animation)
	}
	if got.Actions[1].Action != ActionSayText || got.Actions[1].Parameter != speech {
		t.Fatalf("second action = %#v, want ActionSayText(%q)", got.Actions[1], speech)
	}
	for _, action := range got.Actions {
		if action.Action == ActionGetImage {
			t.Fatalf("unexpected ActionGetImage emitted: %#v", got.Actions)
		}
		if action.Action == ActionNewRequest {
			t.Fatalf("unexpected ActionNewRequest emitted: %#v", got.Actions)
		}
	}
}

func assertRepairPathContains(t *testing.T, repairPath []string, want string) {
	t.Helper()
	for _, got := range repairPath {
		if got == want {
			return
		}
	}
	t.Fatalf("repair path %v missing %q", repairPath, want)
}

func assertRejectReasonContains(t *testing.T, reasons []string, want string) {
	t.Helper()
	for _, got := range reasons {
		if got == want {
			return
		}
	}
	t.Fatalf("reject reasons %v missing %q", reasons, want)
}

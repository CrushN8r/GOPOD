package wirepod_ttr

import "strings"

type EmotionalBeatActionPlanResult struct {
	RawText              string
	Beats                []EmotionalBeatCompileResult
	ActionPlan           []RobotAction
	Warnings             []string
	RejectedReasons      []string
	RepairPath           []string
	LiveRobotSpeech      bool
	RobotCalls           bool
	CameraCalls          bool
	VisionAPICalls       bool
	PerformActionsCalled bool
}

func CompileRobotSafeEmotionalBeatsDry(text string) EmotionalBeatActionPlanResult {
	result := EmotionalBeatActionPlanResult{
		RawText: text,
	}

	for _, line := range strings.Split(text, "\n") {
		beatLine := strings.TrimSpace(line)
		if beatLine == "" {
			result.Warnings = appendUniqueString(result.Warnings, "empty_emotional_beat_line_skipped")
			continue
		}

		beat := CompileEmotionalBeatActions(beatLine)
		result.Beats = append(result.Beats, beat)
		result.ActionPlan = append(result.ActionPlan, beat.Actions...)
		for _, reason := range beat.Packet.RejectedReasons {
			result.RejectedReasons = appendUniqueString(result.RejectedReasons, reason)
		}
		result.RepairPath = append(result.RepairPath, beat.Packet.RepairPath...)
	}

	if len(result.Beats) == 0 {
		beat := CompileEmotionalBeatActions("")
		result.Beats = append(result.Beats, beat)
		result.ActionPlan = append(result.ActionPlan, beat.Actions...)
		for _, reason := range beat.Packet.RejectedReasons {
			result.RejectedReasons = appendUniqueString(result.RejectedReasons, reason)
		}
		result.RepairPath = append(result.RepairPath, beat.Packet.RepairPath...)
	}

	return result
}

func CompileEmotionalBeatActionPlanDry(text string) EmotionalBeatActionPlanResult {
	return CompileRobotSafeEmotionalBeatsDry(text)
}

func appendUniqueString(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

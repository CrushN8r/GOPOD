package wirepod_ttr

import "strings"

type EmotionalBeatPacket struct {
	RawLine         string
	Animation       string
	Speech          string
	FallbackUsed    bool
	RepairPath      []string
	RejectedReasons []string
}

type EmotionalBeatCompileResult struct {
	Packet     EmotionalBeatPacket
	Actions    []RobotAction
	FormatPass bool
}

func CompileEmotionalBeatActions(line string) EmotionalBeatCompileResult {
	packet := ParseEmotionalBeatLine(line)
	actions := EmotionalBeatToRobotActions(packet)
	return EmotionalBeatCompileResult{
		Packet:     packet,
		Actions:    actions,
		FormatPass: len(actions) == 2 && strings.TrimSpace(packet.Speech) != "",
	}
}

func ParseEmotionalBeatLine(line string) EmotionalBeatPacket {
	packet := EmotionalBeatPacket{RawLine: strings.TrimSpace(line)}
	rawParam, rawSpeech, found := strings.Cut(packet.RawLine, "...")
	if !found {
		rawParam = packet.RawLine
		rawSpeech = ""
		packet.FallbackUsed = true
		packet.RepairPath = append(packet.RepairPath, "missing_emotional_delimiter")
	}

	if containsRawCommandSyntax(rawParam) || containsRawCommandSyntax(rawSpeech) {
		packet.RejectedReasons = appendRejectReason(packet.RejectedReasons, "raw_command_syntax_detected")
		packet.RepairPath = append(packet.RepairPath, "raw_command_syntax_removed")
	}

	animation, repair, ok, rejectReason := NormalizeEmotionalBeatParam(rawParam)
	packet.Animation = animation
	if repair != "" {
		packet.RepairPath = append(packet.RepairPath, repair)
	}
	if !ok {
		packet.FallbackUsed = true
		packet.RejectedReasons = appendRejectReason(packet.RejectedReasons, rejectReason)
		packet.RepairPath = append(packet.RepairPath, rejectReason+"_fallback:"+strings.TrimSpace(rawParam)+"->"+animation)
	}

	packet.Speech = cleanEmotionalBeatSpeech(rawSpeech)
	if packet.Speech == "" {
		packet.Speech = gopodRobotSpeechFallbackLine
		packet.FallbackUsed = true
		packet.RepairPath = append(packet.RepairPath, "empty_spoken_text_fallback")
	}
	return packet
}

func NormalizeEmotionalBeatParam(param string) (animation string, repair string, ok bool, rejectReason string) {
	cleanParam := strings.TrimSpace(param)
	if terminalEmotionalBeatParam(cleanParam) {
		return "thinking", "", false, "terminal_action_parameter_rejected"
	}
	if known, exists := knownAnimationParameter(cleanParam); exists {
		if cleanParam != known {
			return known, "animation_param_normalized:" + cleanParam + "->" + known, true, ""
		}
		return known, "", true, ""
	}
	if alias, exists := knownGOPODActionAlias(cleanParam); exists {
		return alias, "animation_param_normalized:" + cleanParam + "->" + alias, true, ""
	}
	normalized := normalizeActionParameter(cleanParam)
	if normalized != "" && normalized != "thinking" {
		return normalized, "animation_param_normalized:" + cleanParam + "->" + normalized, true, ""
	}
	return "thinking", "", false, "invalid_action_parameter"
}

func EmotionalBeatToRobotActions(packet EmotionalBeatPacket) []RobotAction {
	return []RobotAction{
		{
			Action:    ActionPlayAnimationWI,
			Parameter: packet.Animation,
		},
		{
			Action:    ActionSayText,
			Parameter: packet.Speech,
		},
	}
}

func terminalEmotionalBeatParam(param string) bool {
	switch strings.TrimSpace(param) {
	case "getImage", "newVoiceRequest":
		return true
	default:
		return false
	}
}

func cleanEmotionalBeatSpeech(rawSpeech string) string {
	return cleanRobotSpeechText(robotSafeFilter(rawSpeech))
}

func containsRawCommandSyntax(text string) bool {
	return strings.Contains(text, "{{") || strings.Contains(text, "}}")
}

func appendRejectReason(reasons []string, reason string) []string {
	for _, existing := range reasons {
		if existing == reason {
			return reasons
		}
	}
	return append(reasons, reason)
}

package wirepod_ttr

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// gopodSystemSpeechContent is the parsed shape of system_speech_content.json, the single source of
// truth for GOPOD's canned/fixed system speech text.
type gopodSystemSpeechContent struct {
	Description          string                     `json:"description"`
	ContactBeatTemplates []gopodContactBeatTemplate `json:"contact_beat_templates"`
	FallbackLine         string                     `json:"fallback_line"`
	DefaultPrompt        string                     `json:"default_prompt"`
	SessionMemoryHeader  string                     `json:"session_memory_header"`
}

// gopodSystemSpeechContentPath resolves relative to gopodChipperRoot() (see gopod_paths.go),
// overridable via GOPOD_SYSTEM_SPEECH_CONTENT_PATH.
var gopodSystemSpeechContentPath = gopodPathWithEnvOverride(
	"GOPOD_SYSTEM_SPEECH_CONTENT_PATH", "gopod_probes/system_speech_content.json",
)

// Floor values, used ONLY if system_speech_content.json itself cannot be read or parsed.
const gopodFallbackLineFloor = "thinking... My robot brain tripped over the format wire, so I am keeping this clean. Ask me again and I will GOPOD it better."
const gopodDefaultPromptFloor = "You are a helpful, animated robot called Vector. Keep the response concise yet informative."

func loadGOPODSystemSpeechContent() gopodSystemSpeechContent {
	floor := gopodSystemSpeechContent{
		FallbackLine:  gopodFallbackLineFloor,
		DefaultPrompt: gopodDefaultPromptFloor,
	}
	data, err := os.ReadFile(gopodSystemSpeechContentPath)
	if err != nil {
		logger.Println("Unable to load GOPOD system speech content at " + gopodSystemSpeechContentPath + ", using built-in floor values: " + err.Error())
		return floor
	}
	var content gopodSystemSpeechContent
	if err := json.Unmarshal(data, &content); err != nil {
		logger.Println("Unable to parse GOPOD system speech content at " + gopodSystemSpeechContentPath + ", using built-in floor values: " + err.Error())
		return floor
	}
	if strings.TrimSpace(content.FallbackLine) == "" {
		content.FallbackLine = gopodFallbackLineFloor
	}
	if strings.TrimSpace(content.DefaultPrompt) == "" {
		content.DefaultPrompt = gopodDefaultPromptFloor
	}
	return content
}

var loadedGOPODSystemSpeechContent = loadGOPODSystemSpeechContent()

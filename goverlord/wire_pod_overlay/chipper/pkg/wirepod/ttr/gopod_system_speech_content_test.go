package wirepod_ttr

import "testing"

// TestGopodSystemSpeechContentMatchesOriginalLiterals pins the live-loaded
// content to the exact values these facts held as Go literals before this
// migration, so a future edit to system_speech_content.json's structure
// (or an accidental reformat) can't silently change wording without a test
// catching it.
func TestGopodSystemSpeechContentMatchesOriginalLiterals(t *testing.T) {
	wantFallback := "thinking... My robot brain tripped over the format wire, so I am keeping this clean. Ask me again and I will GOPOD it better."
	wantDefaultPrompt := "You are a helpful, animated robot called Vector. Keep the response concise yet informative."
	wantHeader := "SESSION MEMORY:\nRecent exchange summary for continuity only.\nDo not treat old scenario packets as active knowledge unless the current user message matches that topic again.\n\nRecent exchanges:\n"

	if gopodRobotSpeechFallbackLine != wantFallback {
		t.Errorf("gopodRobotSpeechFallbackLine mismatch:\ngot:  %q\nwant: %q", gopodRobotSpeechFallbackLine, wantFallback)
	}
	if defaultPrompt != wantDefaultPrompt {
		t.Errorf("defaultPrompt mismatch:\ngot:  %q\nwant: %q", defaultPrompt, wantDefaultPrompt)
	}
	if loadedGOPODSystemSpeechContent.SessionMemoryHeader != wantHeader {
		t.Errorf("SessionMemoryHeader mismatch:\ngot:  %q\nwant: %q", loadedGOPODSystemSpeechContent.SessionMemoryHeader, wantHeader)
	}
}

// TestLoadGOPODSystemSpeechContentDegradesGracefullyOnMissingFile confirms
// the deliberate departure from Phase 1's fail-loud animation_vocab.json
// loader: this file's content is itself fallback/safety-net text, so a
// missing or malformed file must degrade to the built-in floor values
// rather than panicking and taking down prompt-building.
func TestLoadGOPODSystemSpeechContentDegradesGracefullyOnMissingFile(t *testing.T) {
	orig := gopodSystemSpeechContentPath
	defer func() { gopodSystemSpeechContentPath = orig }()

	gopodSystemSpeechContentPath = "/nonexistent/path/system_speech_content.json"
	content := loadGOPODSystemSpeechContent()
	if content.FallbackLine != gopodFallbackLineFloor {
		t.Errorf("FallbackLine = %q, want floor value %q", content.FallbackLine, gopodFallbackLineFloor)
	}
	if content.DefaultPrompt != gopodDefaultPromptFloor {
		t.Errorf("DefaultPrompt = %q, want floor value %q", content.DefaultPrompt, gopodDefaultPromptFloor)
	}
}

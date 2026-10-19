package wirepod_ttr

import "strings"

const (
	GOPODMarkerAdvanceSourceBackpackTouch    = "backpack_touch"
	GOPODMarkerAdvanceSourceWakeWord         = "wake_word"
	GOPODMarkerAdvanceSourceKeypadTripleZero = "keypad_3x000"
	gopodMarkerAdvanceEventMarkerOne         = "GOPOD_STREAM_MARKER_1"
	gopodMarkerAdvanceEventReleaseReady      = "BROBOT_RELEASE_READY=yes"
)

type GOPODMarkerAdvanceRequest struct {
	Source string
}

func NewGOPODMarkerAdvanceRequest(source string) GOPODMarkerAdvanceRequest {
	return GOPODMarkerAdvanceRequest{Source: normalizeGOPODMarkerAdvanceSource(source)}
}

func (request GOPODMarkerAdvanceRequest) ShouldAdvance() bool {
	return request.Source != ""
}

func (request GOPODMarkerAdvanceRequest) MarkerOneEvents() []string {
	if !request.ShouldAdvance() {
		return nil
	}
	return []string{gopodMarkerAdvanceEventMarkerOne, gopodMarkerAdvanceEventReleaseReady}
}

func emitGOPODMarkerAdvanceEvents(emit func(string), request GOPODMarkerAdvanceRequest) bool {
	if emit == nil || !request.ShouldAdvance() {
		return false
	}
	for _, event := range request.MarkerOneEvents() {
		emit(event)
	}
	return true
}

func normalizeGOPODMarkerAdvanceSource(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "touch", "touch sensor", "backpack_touch", "backpack touch":
		return GOPODMarkerAdvanceSourceBackpackTouch
	case "wake", "wake word", "wake_word":
		return GOPODMarkerAdvanceSourceWakeWord
	case "000", "kp0", "kp0_triple_exit", "keypad_3x000", "keypad triple zero":
		return GOPODMarkerAdvanceSourceKeypadTripleZero
	default:
		return strings.TrimSpace(source)
	}
}

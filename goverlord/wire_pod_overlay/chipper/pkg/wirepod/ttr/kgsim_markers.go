package wirepod_ttr

import "testing"

func TestGOPODMarkerEventsNormalMarkerOneAfterFinalSentChunk(t *testing.T) {
	result := gopodStreamResult{
		Raw:            "raw llm response stays untouched",
		MarkedPacket:   "|||0||| raw llm response stays untouched |||1|||",
		FilteredPacket: "happy... first robot chunk |||C||| thinking... second robot chunk",
		AudibleChunks:  []string{"happy... first robot chunk", "thinking... second robot chunk"},
	}

	events := simulateGOPODMarkerEvents(result, -1)

	assertEventOrder(t, events, []string{
		"GOPOD_STREAM_MARKER_0",
		"RAW_LLM_RESPONSE_MARKED:\nraw llm response stays untouched",
		"BROBOT_MARKED_PACKET:\n|||0||| raw llm response stays untouched |||1|||",
		"BROBOT_FILTER_PACKET:\nhappy... first robot chunk |||C||| thinking... second robot chunk",
		"FILTERED_SPLIT_READY:\nhappy... first robot chunk",
		"FILTERED_SPLIT_SENT:\n{{playAnimationWI||happy}}... first robot chunk",
		"FILTERED_SPLIT_READY:\nthinking... second robot chunk",
		"FILTERED_SPLIT_SENT:\n{{playAnimationWI||thinking}}... second robot chunk",
		"GOPOD_STREAM_MARKER_1",
		"BROBOT_RELEASE_READY=yes",
	})

	assertEventCount(t, events, "GOPOD_STREAM_MARKER_0", 1)
	assertEventCount(t, events, "GOPOD_STREAM_MARKER_1", 1)
	assertEventCount(t, events, "BROBOT_RELEASE_READY=yes", 1)
}

func TestGOPODMarkerEventsTouchMarkerOneStopsLaterChunks(t *testing.T) {
	result := gopodStreamResult{
		Raw:            "raw llm response stays untouched",
		MarkedPacket:   "|||0||| raw llm response stays untouched |||1|||",
		FilteredPacket: "happy... first robot chunk |||C||| thinking... second robot chunk |||C||| love... third robot chunk",
		AudibleChunks:  []string{"happy... first robot chunk", "thinking... second robot chunk", "love... third robot chunk"},
	}

	events := simulateGOPODMarkerEvents(result, 1)

	assertEventOrder(t, events, []string{
		"GOPOD_STREAM_MARKER_0",
		"RAW_LLM_RESPONSE_MARKED:\nraw llm response stays untouched",
		"BROBOT_MARKED_PACKET:\n|||0||| raw llm response stays untouched |||1|||",
		"BROBOT_FILTER_PACKET:\nhappy... first robot chunk |||C||| thinking... second robot chunk |||C||| love... third robot chunk",
		"FILTERED_SPLIT_READY:\nhappy... first robot chunk",
		"FILTERED_SPLIT_SENT:\n{{playAnimationWI||happy}}... first robot chunk",
		"GOPOD_STREAM_MARKER_1",
		"BROBOT_RELEASE_READY=yes",
	})

	assertEventCount(t, events, "GOPOD_STREAM_MARKER_0", 1)
	assertEventCount(t, events, "GOPOD_STREAM_MARKER_1", 1)
	assertEventCount(t, events, "BROBOT_RELEASE_READY=yes", 1)

	assertEventAbsent(t, events, "FILTERED_SPLIT_READY:\nthinking... second robot chunk")
	assertEventAbsent(t, events, "FILTERED_SPLIT_SENT:\n{{playAnimationWI||thinking}}... second robot chunk")
	assertEventAbsent(t, events, "FILTERED_SPLIT_READY:\nlove... third robot chunk")
	assertEventAbsent(t, events, "FILTERED_SPLIT_SENT:\n{{playAnimationWI||love}}... third robot chunk")
}

func simulateGOPODMarkerEvents(result gopodStreamResult, interruptBeforeChunk int) []string {
	var events []string

	emit := func(event string) {
		events = append(events, event)
	}

	emit("GOPOD_STREAM_MARKER_0")
	emit("RAW_LLM_RESPONSE_MARKED:\n" + result.Raw)
	emit("BROBOT_MARKED_PACKET:\n" + result.MarkedPacket)
	emit("BROBOT_FILTER_PACKET:\n" + result.FilteredPacket)

	for i, chunk := range result.AudibleChunks {
		if interruptBeforeChunk == i {
			emitGOPODMarkerAdvanceEvents(emit, NewGOPODMarkerAdvanceRequest(GOPODMarkerAdvanceSourceBackpackTouch))
			return events
		}

		emit("FILTERED_SPLIT_READY:\n" + chunk)
		emit("FILTERED_SPLIT_SENT:\n" + commandPacketFromFilteredChunk(chunk))
	}

	emitGOPODMarkerAdvanceEvents(emit, NewGOPODMarkerAdvanceRequest("normal"))
	return events
}

func assertEventOrder(t *testing.T, events []string, want []string) {
	t.Helper()

	next := 0
	for _, event := range events {
		if next < len(want) && event == want[next] {
			next++
		}
	}

	if next != len(want) {
		t.Fatalf("events missing ordered item %d %q\nall events: %#v", next, want[next], events)
	}
}

func assertEventCount(t *testing.T, events []string, want string, wantCount int) {
	t.Helper()

	got := 0
	for _, event := range events {
		if event == want {
			got++
		}
	}

	if got != wantCount {
		t.Fatalf("event count for %q = %d, want %d\nall events: %#v", want, got, wantCount, events)
	}
}

func assertEventAbsent(t *testing.T, events []string, forbidden string) {
	t.Helper()

	for _, event := range events {
		if event == forbidden {
			t.Fatalf("forbidden event found: %q\nall events: %#v", forbidden, events)
		}
	}
}

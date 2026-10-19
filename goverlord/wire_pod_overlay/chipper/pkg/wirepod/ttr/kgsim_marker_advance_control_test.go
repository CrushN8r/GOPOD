package wirepod_ttr

import (
	"reflect"
	"testing"
)

func TestGOPODMarkerAdvanceRequestUnifiesBackpackTouchAndKeypadTripleZero(t *testing.T) {
	touch := NewGOPODMarkerAdvanceRequest(GOPODMarkerAdvanceSourceBackpackTouch)
	keypad := NewGOPODMarkerAdvanceRequest(GOPODMarkerAdvanceSourceKeypadTripleZero)

	if !touch.ShouldAdvance() {
		t.Fatalf("backpack touch request should advance")
	}
	if !keypad.ShouldAdvance() {
		t.Fatalf("keypad triple-zero request should advance")
	}

	if !reflect.DeepEqual(touch.MarkerOneEvents(), keypad.MarkerOneEvents()) {
		t.Fatalf("touch and keypad marker events differ: touch=%#v keypad=%#v", touch.MarkerOneEvents(), keypad.MarkerOneEvents())
	}

	want := []string{gopodMarkerAdvanceEventMarkerOne, gopodMarkerAdvanceEventReleaseReady}
	if !reflect.DeepEqual(touch.MarkerOneEvents(), want) {
		t.Fatalf("marker events = %#v, want %#v", touch.MarkerOneEvents(), want)
	}
}

func TestGOPODMarkerAdvanceRequestNormalizesCurrentInputNames(t *testing.T) {
	cases := map[string]string{
		"touch sensor":    GOPODMarkerAdvanceSourceBackpackTouch,
		"wake word":       GOPODMarkerAdvanceSourceWakeWord,
		"KP0_TRIPLE_EXIT": GOPODMarkerAdvanceSourceKeypadTripleZero,
		"000":             GOPODMarkerAdvanceSourceKeypadTripleZero,
	}

	for input, want := range cases {
		got := NewGOPODMarkerAdvanceRequest(input)
		if got.Source != want {
			t.Fatalf("source for %q = %q, want %q", input, got.Source, want)
		}
		if !got.ShouldAdvance() {
			t.Fatalf("request for %q should advance", input)
		}
	}
}

func TestGOPODMarkerAdvanceRequestRejectsEmptySource(t *testing.T) {
	request := NewGOPODMarkerAdvanceRequest("   ")

	if request.ShouldAdvance() {
		t.Fatalf("empty source should not advance")
	}
	if events := request.MarkerOneEvents(); len(events) != 0 {
		t.Fatalf("empty source marker events = %#v, want none", events)
	}
}

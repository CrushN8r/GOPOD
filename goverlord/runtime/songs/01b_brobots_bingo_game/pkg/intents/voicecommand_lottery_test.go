package intents

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBingoEngineFirstDrawBounds(t *testing.T) {
	engine := newBingoEngine(1, MAX_NUMBERS)

	number, err := engine.Draw()
	if err != nil {
		t.Fatalf("first draw returned error: %v", err)
	}
	if number < 1 || number > MAX_NUMBERS {
		t.Fatalf("first draw = %d, want 1..%d", number, MAX_NUMBERS)
	}
}

func TestBingoEngineDrawsAllNumbersUniquely(t *testing.T) {
	engine := newBingoEngine(1, MAX_NUMBERS)
	seen := make(map[int]bool, MAX_NUMBERS)

	for i := 0; i < MAX_NUMBERS; i++ {
		number, err := engine.Draw()
		if err != nil {
			t.Fatalf("draw %d returned error: %v", i+1, err)
		}
		if number < 1 || number > MAX_NUMBERS {
			t.Fatalf("draw %d = %d, want 1..%d", i+1, number, MAX_NUMBERS)
		}
		if seen[number] {
			t.Fatalf("draw %d repeated number %d", i+1, number)
		}
		seen[number] = true
	}

	if len(seen) != MAX_NUMBERS {
		t.Fatalf("saw %d numbers, want %d", len(seen), MAX_NUMBERS)
	}
}

func TestBingoEngineReportsExhaustion(t *testing.T) {
	engine := newBingoEngine(1, MAX_NUMBERS)

	for i := 0; i < MAX_NUMBERS; i++ {
		if _, err := engine.Draw(); err != nil {
			t.Fatalf("draw %d returned error before exhaustion: %v", i+1, err)
		}
	}

	if _, err := engine.Draw(); err != ErrBingoExhausted {
		t.Fatalf("exhausted draw error = %v, want %v", err, ErrBingoExhausted)
	}
}

func TestBingoEngineNewEngineResetsState(t *testing.T) {
	engine := newBingoEngine(1, MAX_NUMBERS)

	for i := 0; i < MAX_NUMBERS; i++ {
		if _, err := engine.Draw(); err != nil {
			t.Fatalf("draw %d returned error before exhaustion: %v", i+1, err)
		}
	}

	fresh := newBingoEngine(1, MAX_NUMBERS)
	number, err := fresh.Draw()
	if err != nil {
		t.Fatalf("fresh engine first draw returned error: %v", err)
	}
	if number < 1 || number > MAX_NUMBERS {
		t.Fatalf("fresh engine first draw = %d, want 1..%d", number, MAX_NUMBERS)
	}
}

func TestBingoEngineFreshGamesAreIndependent(t *testing.T) {
	first := newBingoEngine(1, MAX_NUMBERS)
	second := newBingoEngine(2, MAX_NUMBERS)

	for i := 0; i < MAX_NUMBERS; i++ {
		if _, err := first.Draw(); err != nil {
			t.Fatalf("first engine draw %d returned error: %v", i+1, err)
		}
	}
	if _, err := first.Draw(); err != ErrBingoExhausted {
		t.Fatalf("first exhausted draw error = %v, want %v", err, ErrBingoExhausted)
	}
	if second.Remaining() != MAX_NUMBERS {
		t.Fatalf("second remaining = %d, want %d", second.Remaining(), MAX_NUMBERS)
	}
	if number, err := second.Draw(); err != nil || number < 1 || number > MAX_NUMBERS {
		t.Fatalf("second first draw = %d, err=%v; want valid number", number, err)
	}
}

func TestBingoInputControllerAcceptsTouchRisingEdge(t *testing.T) {
	controller := newBingoInputController(400 * time.Millisecond)
	now := time.Unix(1, 0)

	if action := controller.Handle(bingoInputEvent{touched: false, at: now}, bingoStateReady); action != bingoInputNone {
		t.Fatalf("released action = %v, want %v", action, bingoInputNone)
	}
	if action := controller.Handle(bingoInputEvent{touched: true, at: now.Add(time.Second)}, bingoStateReady); action != bingoInputDraw {
		t.Fatalf("rising edge action = %v, want %v", action, bingoInputDraw)
	}
}

func TestBingoInputControllerSuppressesHeldTouch(t *testing.T) {
	controller := newBingoInputController(400 * time.Millisecond)
	now := time.Unix(1, 0)

	controller.Handle(bingoInputEvent{touched: false, at: now}, bingoStateReady)
	if action := controller.Handle(bingoInputEvent{touched: true, at: now.Add(time.Second)}, bingoStateReady); action != bingoInputDraw {
		t.Fatalf("first touch action = %v, want %v", action, bingoInputDraw)
	}
	if action := controller.Handle(bingoInputEvent{touched: true, at: now.Add(2 * time.Second)}, bingoStateReady); action != bingoInputNone {
		t.Fatalf("held touch action = %v, want %v", action, bingoInputNone)
	}
}

func TestBingoInputControllerRequiresInitialReleaseBeforeFirstDraw(t *testing.T) {
	controller := newBingoInputController(400 * time.Millisecond)
	now := time.Unix(1, 0)

	if action := controller.Handle(bingoInputEvent{touched: true, at: now}, bingoStateReady); action != bingoInputNone {
		t.Fatalf("startup held touch action = %v, want %v", action, bingoInputNone)
	}
	controller.Handle(bingoInputEvent{touched: false, at: now.Add(time.Second)}, bingoStateReady)
	if action := controller.Handle(bingoInputEvent{touched: true, at: now.Add(2 * time.Second)}, bingoStateReady); action != bingoInputDraw {
		t.Fatalf("touch after initial release action = %v, want %v", action, bingoInputDraw)
	}
}

func TestBingoInputControllerRequiresReleaseAndDebouncesNoise(t *testing.T) {
	controller := newBingoInputController(400 * time.Millisecond)
	now := time.Unix(1, 0)

	controller.Handle(bingoInputEvent{touched: false, at: now}, bingoStateReady)
	if action := controller.Handle(bingoInputEvent{touched: true, at: now.Add(time.Second)}, bingoStateReady); action != bingoInputDraw {
		t.Fatalf("first touch action = %v, want %v", action, bingoInputDraw)
	}
	controller.Handle(bingoInputEvent{touched: false, at: now.Add(1050 * time.Millisecond)}, bingoStateReady)
	if action := controller.Handle(bingoInputEvent{touched: true, at: now.Add(1100 * time.Millisecond)}, bingoStateReady); action != bingoInputNone {
		t.Fatalf("debounced touch action = %v, want %v", action, bingoInputNone)
	}
	controller.Handle(bingoInputEvent{touched: false, at: now.Add(1450 * time.Millisecond)}, bingoStateReady)
	if action := controller.Handle(bingoInputEvent{touched: true, at: now.Add(1500 * time.Millisecond)}, bingoStateReady); action != bingoInputDraw {
		t.Fatalf("post-debounce touch action = %v, want %v", action, bingoInputDraw)
	}
}

func TestBingoInputControllerRejectsTouchWhileDrawing(t *testing.T) {
	controller := newBingoInputController(400 * time.Millisecond)
	now := time.Unix(1, 0)

	controller.Handle(bingoInputEvent{touched: false, at: now}, bingoStateReady)
	if action := controller.Handle(bingoInputEvent{touched: true, at: now.Add(time.Second)}, bingoStateDrawing); action != bingoInputIgnored {
		t.Fatalf("drawing touch action = %v, want %v", action, bingoInputIgnored)
	}
	if action := controller.Handle(bingoInputEvent{touched: true, at: now.Add(2 * time.Second)}, bingoStateReady); action != bingoInputNone {
		t.Fatalf("held after drawing action = %v, want %v", action, bingoInputNone)
	}
	controller.Handle(bingoInputEvent{touched: false, at: now.Add(3 * time.Second)}, bingoStateReady)
	if action := controller.Handle(bingoInputEvent{touched: true, at: now.Add(4 * time.Second)}, bingoStateReady); action != bingoInputDraw {
		t.Fatalf("released retouch action = %v, want %v", action, bingoInputDraw)
	}
}

func TestBingoInputControllerRejectsTouchWhileEnding(t *testing.T) {
	controller := newBingoInputController(400 * time.Millisecond)

	controller.Handle(bingoInputEvent{touched: false, at: time.Unix(1, 0)}, bingoStateReady)
	if action := controller.Handle(bingoInputEvent{touched: true, at: time.Unix(2, 0)}, bingoStateEnding); action != bingoInputIgnored {
		t.Fatalf("ending touch action = %v, want %v", action, bingoInputIgnored)
	}
	if action := controller.Handle(bingoInputEvent{buttonPressed: true, at: time.Unix(3, 0)}, bingoStateEnding); action != bingoInputNone {
		t.Fatalf("ending button action = %v, want %v", action, bingoInputNone)
	}
}

func TestBingoInputControllerBackpackButtonExits(t *testing.T) {
	controller := newBingoInputController(400 * time.Millisecond)

	if action := controller.Handle(bingoInputEvent{buttonPressed: true, touched: true, at: time.Unix(1, 0)}, bingoStateReady); action != bingoInputExit {
		t.Fatalf("button action = %v, want %v", action, bingoInputExit)
	}
}

type fakeBingoOutputOps struct {
	calls       []string
	rattleErr   error
	speechErr   error
	panicRattle bool
}

func (f *fakeBingoOutputOps) PlayRattle(context.Context) error {
	f.calls = append(f.calls, "rattle")
	if f.panicRattle {
		panic("rattle panic")
	}
	return f.rattleErr
}

func (f *fakeBingoOutputOps) SayText(_ context.Context, text string) error {
	f.calls = append(f.calls, "say:"+text)
	return f.speechErr
}

func TestBingoOutputControllerSerializesRattleThenNativeSpeech(t *testing.T) {
	ops := &fakeBingoOutputOps{}
	controller := newBingoOutputController(ops, time.Second)

	if err := controller.PlayDraw(context.Background(), 42); err != nil {
		t.Fatalf("PlayDraw returned error: %v", err)
	}

	want := []string{"rattle", "say:42"}
	if len(ops.calls) != len(want) {
		t.Fatalf("calls = %v, want %v", ops.calls, want)
	}
	for i := range want {
		if ops.calls[i] != want[i] {
			t.Fatalf("calls = %v, want %v", ops.calls, want)
		}
	}
}

func TestBingoOutputControllerStopsSpeechWhenRattleFails(t *testing.T) {
	rattleErr := errors.New("rattle failed")
	ops := &fakeBingoOutputOps{rattleErr: rattleErr}
	controller := newBingoOutputController(ops, time.Second)

	if err := controller.PlayDraw(context.Background(), 42); !errors.Is(err, rattleErr) {
		t.Fatalf("PlayDraw error = %v, want %v", err, rattleErr)
	}
	if len(ops.calls) != 1 || ops.calls[0] != "rattle" {
		t.Fatalf("calls = %v, want [rattle]", ops.calls)
	}
}

func TestBingoOutputControllerReportsSpeechFailure(t *testing.T) {
	speechErr := errors.New("speech failed")
	ops := &fakeBingoOutputOps{speechErr: speechErr}
	controller := newBingoOutputController(ops, time.Second)

	if err := controller.PlayDraw(context.Background(), 42); !errors.Is(err, speechErr) {
		t.Fatalf("PlayDraw error = %v, want %v", err, speechErr)
	}
}

func TestBingoOutputControllerRecoversPanic(t *testing.T) {
	ops := &fakeBingoOutputOps{panicRattle: true}
	controller := newBingoOutputController(ops, time.Second)

	if err := controller.PlayDraw(context.Background(), 42); err == nil {
		t.Fatal("PlayDraw returned nil error after panic")
	}
}

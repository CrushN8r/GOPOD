package intents

import (
	"context"
	"errors"
	"fmt"
	sdk_wrapper "github.com/fforchino/vector-go-sdk/pkg/sdk-wrapper"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"
)

/**********************************************************************************************************************/
/*                                          NEW YEAR'S EVE BINGO                                                      */
/**********************************************************************************************************************/

const MAX_NUMBERS = 90 // native maximum; unreferenced by production code after Task B's --grid-size flag (kept as documented ceiling, still used by tests)
const bingoTouchDebounce = 400 * time.Millisecond
const bingoOutputTimeout = 15 * time.Second
const bingoOpeningPhrase = "bingo time!"
const bingoExitPhrase = "game over!"

var ErrBingoExhausted = errors.New("bingo number pool exhausted")

// Set by main.go from CLI flags before the intent handler chain runs.
var BingoGridSize = 90
var BingoSilentMode = false
var BingoChatFilePath = ""
var BingoPauseForTouch = false

type bingoState int

const (
	bingoStateOpening bingoState = iota
	bingoStateReady
	bingoStateDrawing
	bingoStateEnding
	bingoStateShutdown
)

type bingoEngine struct {
	numbers []int
	next    int
}

type bingoInputAction int

const (
	bingoInputNone bingoInputAction = iota
	bingoInputDraw
	bingoInputExit
	bingoInputIgnored
)

type bingoInputEvent struct {
	touched       bool
	buttonPressed bool
	at            time.Time
}

type bingoDrawResult struct {
	number int
	err    error
}

type bingoInputController struct {
	debounce          time.Duration
	armed             bool
	wasTouched        bool
	lastAccepted      time.Time
	ignoredTouchBurst bool
}

type bingoOutputOps interface {
	PlayRattle(context.Context) error
	SayText(context.Context, string) error
}

type bingoOutputController struct {
	ops     bingoOutputOps
	timeout time.Duration
	mu      sync.Mutex
}

type vectorBingoOutputOps struct {
	rattlePath string
}

func newBingoEngine(seed int64, size int) *bingoEngine {
	numbers := make([]int, size)
	for i := range numbers {
		numbers[i] = i + 1
	}

	rnd := rand.New(rand.NewSource(seed))
	rnd.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})

	return &bingoEngine{numbers: numbers}
}

func (e *bingoEngine) Draw() (int, error) {
	if e.next >= len(e.numbers) {
		return 0, ErrBingoExhausted
	}

	number := e.numbers[e.next]
	e.next++
	return number, nil
}

func newBingoInputController(debounce time.Duration) *bingoInputController {
	return &bingoInputController{debounce: debounce}
}

func (c *bingoInputController) Handle(evt bingoInputEvent, state bingoState) bingoInputAction {
	if state == bingoStateReady {
		c.ignoredTouchBurst = false
	}

	if evt.buttonPressed {
		if state == bingoStateEnding || state == bingoStateShutdown {
			return bingoInputNone
		}
		return bingoInputExit
	}

	if !evt.touched {
		c.armed = true
		c.wasTouched = false
		c.ignoredTouchBurst = false
		return bingoInputNone
	}

	if c.wasTouched {
		return bingoInputNone
	}
	c.wasTouched = true

	if !c.armed {
		return bingoInputNone
	}

	if state != bingoStateReady {
		if c.ignoredTouchBurst {
			return bingoInputNone
		}
		c.ignoredTouchBurst = true
		return bingoInputIgnored
	}

	if !c.lastAccepted.IsZero() && evt.at.Sub(c.lastAccepted) < c.debounce {
		return bingoInputNone
	}

	c.lastAccepted = evt.at
	return bingoInputDraw
}

func newBingoOutputController(ops bingoOutputOps, timeout time.Duration) *bingoOutputController {
	return &bingoOutputController{ops: ops, timeout: timeout}
}

func (c *bingoOutputController) PlayDraw(ctx context.Context, number int) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	defer recoverBingoOutput(&err)

	logBingo("BINGO_DRAW_START number=%d", number)
	if err := c.ops.PlayRattle(ctx); err != nil {
		return fmt.Errorf("rattle: %w", err)
	}
	logBingo("BINGO_RATTLE_DONE number=%d", number)
	if err := c.ops.SayText(ctx, formatBingoNumber(number)); err != nil {
		return fmt.Errorf("speech: %w", err)
	}
	logBingo("BINGO_SPEECH_DONE number=%d", number)
	logBingo("BINGO_DRAW_PASS number=%d", number)
	return nil
}

func (c *bingoOutputController) SpeakText(ctx context.Context, text string) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	defer recoverBingoOutput(&err)

	if err := c.ops.SayText(ctx, text); err != nil {
		return err
	}
	return nil
}

func (c *bingoOutputController) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if c.timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, c.timeout)
}

func recoverBingoOutput(err *error) {
	if recovered := recover(); recovered != nil {
		*err = fmt.Errorf("bingo output panic: %v", recovered)
	}
}

func newVectorBingoOutputOps() *vectorBingoOutputOps {
	return &vectorBingoOutputOps{rattlePath: sdk_wrapper.GetDataPath("audio/rattle.wav")}
}

func (o *vectorBingoOutputOps) PlayRattle(ctx context.Context) error {
	return playBingoSound(ctx, o.rattlePath)
}

func (o *vectorBingoOutputOps) SayText(ctx context.Context, text string) error {
	if sdk_wrapper.Robot == nil {
		return errors.New("robot is not initialized")
	}
	_, err := sdk_wrapper.Robot.Conn.SayText(
		ctx,
		&vectorpb.SayTextRequest{
			Text:           text,
			UseVectorVoice: true,
			DurationScalar: 1.0,
		},
	)
	return err
}

type silentBingoOutputOps struct {
	inner bingoOutputOps
}

func newSilentBingoOutputOps(inner bingoOutputOps) *silentBingoOutputOps {
	return &silentBingoOutputOps{inner: inner}
}

func (o *silentBingoOutputOps) PlayRattle(ctx context.Context) error {
	return o.inner.PlayRattle(ctx)
}

func (o *silentBingoOutputOps) SayText(ctx context.Context, text string) error {
	logBingo("GOPOD_BINGO_SILENT_SAYTEXT text=%q", text)
	return nil
}

// BingoCallFormat is the spoken phrase for each ball call. %d is the drawn
// number, filled in at call time. Edit this line to change what Doc says on
// every draw (e.g. "Big Shiny Bingo Ball, number %d!" to match the scripted
// bingo song's own call style) - requires a rebuild (goverlord/runtime/songs/102_brobots_bingo_game/README.md's
// documented overlay-build process), not a live-editable knob like the golden
// song engine's knobs.json, since this is compiled Go, not a scored song.
var BingoCallFormat = "%d"

func formatBingoNumber(number int) string {
	return fmt.Sprintf(BingoCallFormat, number)
}

func logBingo(format string, args ...interface{}) {
	println(fmt.Sprintf(format, args...))
}

func playBingoSound(ctx context.Context, filename string) error {
	if sdk_wrapper.Robot == nil {
		return errors.New("robot is not initialized")
	}
	if _, err := os.Stat(filename); err != nil {
		return fmt.Errorf("rattle sound unavailable: %w", err)
	}

	pcm, cleanup, err := loadBingoPCM(ctx, filename)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		return err
	}

	audioClient, err := sdk_wrapper.Robot.Conn.ExternalAudioStreamPlayback(ctx)
	if err != nil {
		return fmt.Errorf("open audio stream: %w", err)
	}
	defer func() {
		if closeErr := audioClient.CloseSend(); closeErr != nil {
			println("BINGO audio close error:", closeErr.Error())
		}
	}()

	if err := audioClient.Send(&vectorpb.ExternalAudioStreamRequest{
		AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamPrepare{
			AudioStreamPrepare: &vectorpb.ExternalAudioStreamPrepare{
				AudioFrameRate: 16000,
				AudioVolume:    uint32(sdk_wrapper.GetAudioVolume()),
			},
		},
	}); err != nil {
		return fmt.Errorf("prepare audio stream: %w", err)
	}

	for offset := 0; offset < len(pcm); offset += 1024 {
		end := offset + 1024
		if end > len(pcm) {
			end = len(pcm)
		}
		chunk := pcm[offset:end]
		if err := audioClient.Send(&vectorpb.ExternalAudioStreamRequest{
			AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamChunk{
				AudioStreamChunk: &vectorpb.ExternalAudioStreamChunk{
					AudioChunkSizeBytes: uint32(len(chunk)),
					AudioChunkSamples:   chunk,
				},
			},
		}); err != nil {
			return fmt.Errorf("send audio chunk: %w", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(30 * time.Millisecond):
		}
	}

	if err := audioClient.Send(&vectorpb.ExternalAudioStreamRequest{
		AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamComplete{
			AudioStreamComplete: &vectorpb.ExternalAudioStreamComplete{},
		},
	}); err != nil {
		return fmt.Errorf("complete audio stream: %w", err)
	}
	return nil
}

func loadBingoPCM(ctx context.Context, filename string) ([]byte, func(), error) {
	if strings.HasSuffix(filename, ".pcm") || strings.HasSuffix(filename, ".raw") {
		pcm, err := os.ReadFile(filename)
		if err != nil {
			return nil, nil, fmt.Errorf("read pcm audio: %w", err)
		}
		return pcm, nil, nil
	}

	tmpFile, err := os.CreateTemp(sdk_wrapper.GetTempPath(), "bingo-sound-*.pcm")
	if err != nil {
		return nil, nil, fmt.Errorf("create temp pcm: %w", err)
	}
	tmpName := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		return nil, nil, fmt.Errorf("close temp pcm: %w", err)
	}
	cleanup := func() {
		_ = os.Remove(tmpName)
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-i", filename, "-f", "s16le", "-acodec", "pcm_s16le", "-ar", "16000", "-ac", "1", tmpName)
	if output, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("convert rattle audio: %w: %s", err, string(output))
	}

	pcm, err := os.ReadFile(tmpName)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("read converted audio: %w", err)
	}
	return pcm, cleanup, nil
}

func Lottery_Register(intentList *[]IntentDef) error {
	utterances := make(map[string][]string)
	utterances[LOCALE_ENGLISH] = []string{"bingo"}
	utterances[LOCALE_ITALIAN] = []string{"tombola"}
	utterances[LOCALE_SPANISH] = []string{"bingo"}
	utterances[LOCALE_FRENCH] = []string{"bingo"}
	utterances[LOCALE_GERMAN] = []string{"bingo"}

	var intent = IntentDef{
		IntentName:            "extended_intent_bingo",
		Utterances:            utterances,
		Parameters:            []string{},
		Handler:               bingo,
		OSKRTriggersUserInput: nil,
	}
	*intentList = append(*intentList, intent)
	addLocalizedString("STR_BINGO_BINGO", []string{"bingo time!", "tombola!", "bingo!", "bingo!", "bingo!"})
	addLocalizedString("STR_BINGO_GAME_OVER", []string{"game over!", "la partita è finita!", "juego terminado!", "jeu terminé!", "spiel vorbei!"})

	return nil
}

func bingo(intent IntentDef, speechText string, params IntentParams) string {
	returnIntent := STANDARD_INTENT_GREETING_HELLO
	doBingo(BingoGridSize, BingoSilentMode, BingoChatFilePath, BingoPauseForTouch)
	return returnIntent
}

func doBingo(gridSize int, silent bool, chatFilePath string, pauseForTouch bool) {
	ctx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stopSignals()

	events := make(chan bingoInputEvent, 8)
	eventErrors := make(chan error, 1)
	var readers sync.WaitGroup
	readers.Add(1)
	go func() {
		defer readers.Done()
		readVectorBingoInputEvents(ctx, events, eventErrors)
	}()

	engine := newBingoEngine(time.Now().UnixNano(), gridSize)
	input := newBingoInputController(bingoTouchDebounce)
	var outputOps bingoOutputOps = newVectorBingoOutputOps()
	if silent {
		outputOps = newSilentBingoOutputOps(outputOps)
	}
	output := newBingoOutputController(outputOps, bingoOutputTimeout)

	if err := runBingoCoordinator(ctx, engine, input, output, events, eventErrors, chatFilePath, pauseForTouch); err != nil {
		logBingo("BINGO_ERROR stage=fatal number=0 error=%q", err.Error())
	}
	stopSignals()
}

func readVectorBingoInputEvents(ctx context.Context, events chan<- bingoInputEvent, eventErrors chan<- error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		evt := sdk_wrapper.WaitForEvent()
		if evt == nil {
			select {
			case eventErrors <- errors.New("robot event stream closed"):
			default:
			}
			return
		}
		input, ok := vectorEventToBingoInput(evt)
		if !ok {
			continue
		}
		select {
		case events <- input:
		case <-ctx.Done():
			return
		default:
		}
	}
}

func vectorEventToBingoInput(evt *vectorpb.Event) (bingoInputEvent, bool) {
	robotState := evt.GetRobotState()
	if robotState == nil {
		return bingoInputEvent{}, false
	}
	touched := false
	if robotState.TouchData != nil {
		touched = robotState.TouchData.IsBeingTouched
	}
	return bingoInputEvent{
		touched:       touched,
		buttonPressed: robotState.Status&uint32(vectorpb.RobotStatus_ROBOT_STATUS_IS_BUTTON_PRESSED) > 0,
		at:            time.Now(),
	}, true
}

func runBingoCoordinator(
	ctx context.Context,
	engine *bingoEngine,
	input *bingoInputController,
	output *bingoOutputController,
	events <-chan bingoInputEvent,
	eventErrors <-chan error,
	chatFilePath string,
	pauseForTouch bool,
) error {
	state := bingoStateOpening
	// bingo_start is reserved for a future "game truly started" event; deck-prepared state now emits bingo_ready.
	AppendBingoEnvelope(chatFilePath, "bingo_ready",
		fmt.Sprintf("Bingo ready. Grid size: %d.", len(engine.numbers)))
	if err := output.SpeakText(ctx, bingoOpeningPhrase); err != nil {
		logBingo("BINGO_ERROR stage=opening number=0 error=%q", err.Error())
		return fmt.Errorf("opening speech failed: %w", err)
	}

	drainBingoInputEvents(events)
	state = bingoStateReady
	logBingo("BINGO_READY")

	drawRequests := make(chan int, 1)
	drawResults := make(chan bingoDrawResult, 1)
	defer close(drawRequests)
	go func() {
		for number := range drawRequests {
			drawResults <- bingoDrawResult{number: number, err: output.PlayDraw(ctx, number)}
		}
	}()

	pendingExit := false
	exitReason := ""
	autoDrawActive := false

	// Phase B's ticker is created once the first (touch-triggered) draw is dispatched.
	// It must be stopped on every exit path, hence the coordinator-scope defer.
	var ticker *time.Ticker
	var tickerC <-chan time.Time
	defer func() {
		if ticker != nil {
			ticker.Stop()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			state = bingoStateShutdown
			logBingo("BINGO_EXIT reason=signal")
			return nil
		case err := <-eventErrors:
			if err != nil {
				state = bingoStateShutdown
				logBingo("BINGO_EXIT reason=event_error")
				return err
			}
		case result := <-drawResults:
			if result.err != nil {
				logBingo("BINGO_ERROR stage=%s number=%d error=%q", bingoDrawErrorStage(result.err), result.number, result.err.Error())
				return result.err
			}
			if pendingExit {
				return finishBingo(ctx, output, exitReason, chatFilePath, engine)
			}
			if engine.Remaining() == 0 {
				logBingo("BINGO_GAME_COMPLETE draws=%d", len(engine.numbers))
				return finishBingo(ctx, output, "complete", chatFilePath, engine)
			}
			drainBingoInputEvents(events)
			state = bingoStateReady
			logBingo("BINGO_READY")
		case <-tickerC:
			// Only draw on a tick if the previous draw has fully resolved (state==bingoStateReady);
			// otherwise skip this tick and let it be picked up once the coordinator returns to ready.
			if state != bingoStateReady {
				continue
			}
			number, err := engine.Draw()
			if err != nil {
				if errors.Is(err, ErrBingoExhausted) {
					logBingo("BINGO_GAME_COMPLETE draws=%d", len(engine.numbers))
					return finishBingo(ctx, output, "complete", chatFilePath, engine)
				}
				return err
			}
			AppendBingoEnvelope(chatFilePath, "bingo_draw", fmt.Sprintf("%d", number))
			logBingo("GOPOD_BINGO_AUTO_DRAW number=%d", number)
			state = bingoStateDrawing
			drawRequests <- number
		case evt := <-events:
			action := input.Handle(evt, state)
			switch action {
			case bingoInputExit:
				if state == bingoStateDrawing {
					pendingExit = true
					exitReason = "button"
					state = bingoStateEnding
					continue
				}
				return finishBingo(ctx, output, "button", chatFilePath, engine)
			case bingoInputDraw:
				if autoDrawActive {
					logBingo("GOPOD_BINGO_TOUCH_IGNORED phase=auto_draw")
					continue
				}
				number, err := engine.Draw()
				if err != nil {
					if errors.Is(err, ErrBingoExhausted) {
						logBingo("BINGO_GAME_COMPLETE draws=%d", len(engine.numbers))
						return finishBingo(ctx, output, "complete", chatFilePath, engine)
					}
					return err
				}
				AppendBingoEnvelope(chatFilePath, "bingo_draw", fmt.Sprintf("%d", number))
				state = bingoStateDrawing
				drawRequests <- number
				if !pauseForTouch {
					autoDrawActive = true
					ticker = time.NewTicker(3 * time.Second)
					tickerC = ticker.C
				}
			case bingoInputNone:
			case bingoInputIgnored:
				logBingo("BINGO_INPUT_IGNORED reason=state_%s", state.String())
			}
		}
	}
}

func drainBingoInputEvents(events <-chan bingoInputEvent) {
	for {
		select {
		case <-events:
		default:
			return
		}
	}
}

func bingoDrawErrorStage(err error) string {
	if strings.HasPrefix(err.Error(), "rattle:") {
		return "rattle"
	}
	if strings.HasPrefix(err.Error(), "speech:") {
		return "speech"
	}
	return "draw"
}

func (e *bingoEngine) Remaining() int {
	if e.next >= len(e.numbers) {
		return 0
	}
	return len(e.numbers) - e.next
}

func (s bingoState) String() string {
	switch s {
	case bingoStateOpening:
		return "opening"
	case bingoStateReady:
		return "ready"
	case bingoStateDrawing:
		return "drawing"
	case bingoStateEnding:
		return "ending"
	case bingoStateShutdown:
		return "shutdown"
	default:
		return "unknown"
	}
}

func finishBingo(ctx context.Context, output *bingoOutputController, reason string, chatFilePath string, engine *bingoEngine) error {
	if err := output.SpeakText(ctx, bingoExitPhrase); err != nil {
		logBingo("BINGO_ERROR stage=ending number=0 error=%q", err.Error())
		return fmt.Errorf("exit speech failed: %w", err)
	}
	logBingo("BINGO_EXIT reason=%s", reason)
	if reason == "button" {
		AppendBingoEnvelope(chatFilePath, "bingo_deck_reveal",
			fmt.Sprintf("Deck (full order): %s", formatBingoDeck(engine.numbers)))
	}
	AppendBingoEnvelope(chatFilePath, "bingo_end", "Bingo game ended.")
	return nil
}

func formatBingoDeck(numbers []int) string {
	var deckStr strings.Builder
	for i, n := range numbers {
		if i > 0 {
			deckStr.WriteString(", ")
		}
		deckStr.WriteString(strconv.Itoa(n))
	}
	return deckStr.String()
}

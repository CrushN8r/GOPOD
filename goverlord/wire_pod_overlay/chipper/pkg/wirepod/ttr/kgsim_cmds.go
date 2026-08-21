package wirepod_ttr

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/sashabaranov/go-openai"
)

const (
	ActionSayText = iota
	ActionPlayAnimation
	ActionPlayAnimationWI
	ActionGetImage
	ActionNewRequest
	ActionPlaySound
)

// animationMap is derived from animation_vocab.json - see animation_vocab.go.

var soundMap = [][2]string{
	{"danger-will-robinson", "sounds/danger-will-robinson.wav"},
}

type RobotAction struct {
	Action    int
	Parameter string
}

type LLMCommand struct {
	Command         string
	Description     string
	ParamChoices    string
	Action          int
	SupportedModels []string
}

var ValidLLMCommands = []LLMCommand{
	{
		Command:         "playAnimationWI",
		Description:     "Plays an animation on the robot without interrupting speech. This should be used FAR more than the playAnimation command. This is great for storytelling and making any normal response animated. Don't put two of these right next to each other. Use this MANY times. The param choices are the only choices you have. You can't create any.",
		ParamChoices:    animationParamChoices,
		Action:          ActionPlayAnimationWI,
		SupportedModels: []string{"all"},
	},
	{
		Command:         "playAnimation",
		Description:     "Plays an animation on the robot. This will interrupt speech. Only use this if you are directed to play an animaion.",
		ParamChoices:    animationParamChoices,
		Action:          ActionPlayAnimation,
		SupportedModels: []string{"all"},
	},
	{
		Command:         "getImage",
		Description:     "Gets an image from the robot's camera and places it in the next message. If you want to do this, tell the user what you are about to do THEN use the command. This command should END a sentence. Your response will be stopped when this command is recognized. If a user says something like 'what do you see', you should assume that you need to take a new photo. Do NOT automatically assume that you are analyzing a previous photo.",
		ParamChoices:    "front, lookingUp",
		Action:          ActionGetImage,
		SupportedModels: []string{"all"},
	},
	{
		Command:         "newVoiceRequest",
		Description:     "Starts a new voice command from the robot. Use this if you want more input from the user after your response/if you want to carry out a conversation. Below this, there should be a NOTE telling you whether you are in conversation mode or not. If you are, DONT BE AFRAID TO USE THIS COMMAND! This goes at the end of your response, if you use it.",
		ParamChoices:    "now",
		Action:          ActionNewRequest,
		SupportedModels: []string{"all"},
	},
	{
		Command:         "playSound",
		Description:     "Plays a sound on the robot.",
		ParamChoices:    "danger-will-robinson",
		Action:          ActionPlaySound,
		SupportedModels: []string{"all"},
	},
}

func ModelIsSupported(cmd LLMCommand, model string) bool {
	for _, supported := range cmd.SupportedModels {
		if supported == "all" || supported == model {
			return true
		}
	}
	return false
}

func CreatePrompt(origPrompt string, model string, isKG bool) string {
	if promptIsBrobotPrompt(origPrompt) {
		return origPrompt
	}
	prompt := origPrompt + "\n\nKeep in mind, user input comes from speech-to-text software, so respond accordingly."
	if vars.APIConfig.Knowledge.CommandsEnable && !promptUsesCanonicalBrobotRawShape(origPrompt) {
		prompt += "\n\nUse only valid command action parameter choices listed below:"
		for _, cmd := range ValidLLMCommands {
			if ModelIsSupported(cmd, model) {
				prompt += "\n\nCommand Name: " + cmd.Command + "\nDescription: " + cmd.Description + "\nParameter choices: " + cmd.ParamChoices
			}
		}
		if isKG && vars.APIConfig.Knowledge.SaveChat {
			prompt += "\n\nNOTE: You are in 'conversation' mode. If you ask the user a question near the end of your response, you MUST use newVoiceRequest. If you decide you want to end the conversation, you should not use it."
		} else {
			prompt += "\n\nNOTE: You are NOT in 'conversation' mode. Refrain from asking the user any questions and from using newVoiceRequest."
		}
	}
	if os.Getenv("DEBUG_PRINT_PROMPT") == "true" {
		logger.Println(fmt.Sprintf("DEBUG_PROMPT_BYTES=%d", len([]byte(prompt))))
	}
	return prompt
}

func GetActionsFromString(input string) []RobotAction {
	input = NormalizeAnimationSyntax(input)
	splitInput := strings.Split(input, "{{")
	if len(splitInput) == 1 {
		return []RobotAction{{Action: ActionSayText, Parameter: input}}
	}
	var actions []RobotAction
	for _, spl := range splitInput {
		if strings.TrimSpace(spl) == "" {
			continue
		}
		if !strings.Contains(spl, "}}") {
			actions = append(actions, RobotAction{Action: ActionSayText, Parameter: strings.TrimSpace(spl)})
			continue
		}
		splitCommand := strings.SplitN(spl, "}}", 2)
		cmdPlusParam := strings.SplitN(strings.TrimSpace(splitCommand[0]), "||", 2)
		if len(cmdPlusParam) < 2 {
			continue
		}
		action := CmdParamToAction(strings.TrimSpace(cmdPlusParam[0]), strings.TrimSpace(cmdPlusParam[1]))
		if action.Action != -1 {
			actions = append(actions, action)
		}
		if len(splitCommand) == 2 {
			if speech := cleanRobotSpeechText(splitCommand[1]); speech != "" {
				actions = append(actions, RobotAction{Action: ActionSayText, Parameter: speech})
			}
		}
	}
	return actions
}

func CmdParamToAction(cmd, param string) RobotAction {
	for _, command := range ValidLLMCommands {
		if cmd == command.Command {
			return RobotAction{Action: command.Action, Parameter: param}
		}
	}
	logger.Println("LLM tried to do a command which doesn't exist: " + cmd + " (param: " + param + ")")
	return RobotAction{Action: -1}
}

func DoPlayAnimation(animation string, robot *vector.Vector) error {
	for _, animThing := range animationMap {
		if animation == animThing[0] {
			StartAnim_Queue(robot.Cfg.SerialNo)
			robot.Conn.PlayAnimation(context.Background(), &vectorpb.PlayAnimationRequest{
				Animation: &vectorpb.Animation{Name: animThing[1]},
				Loops:     1,
			})
			StopAnim_Queue(robot.Cfg.SerialNo)
			return nil
		}
	}
	logger.Println("Animation provided by LLM doesn't exist: " + animation)
	return nil
}

func DoPlayAnimationWI(animation string, robot *vector.Vector) error {
	for _, animThing := range animationMap {
		if animation == animThing[0] {
			go func() {
				StartAnim_Queue(robot.Cfg.SerialNo)
				robot.Conn.PlayAnimation(context.Background(), &vectorpb.PlayAnimationRequest{
					Animation: &vectorpb.Animation{Name: animThing[1]},
					Loops:     1,
				})
				StopAnim_Queue(robot.Cfg.SerialNo)
			}()
			return nil
		}
	}
	logger.Println("Animation provided by LLM doesn't exist: " + animation)
	return nil
}

func DoPlaySound(sound string, robot *vector.Vector) error {
	for _, soundThing := range soundMap {
		if sound == soundThing[0] {
			return playSoundFile(soundThing[1], robot)
		}
	}
	logger.Println("Sound provided by LLM doesn't exist: " + sound)
	return nil
}

// playSoundFile streams a mono 16-bit PCM WAV file to the robot via
// ExternalAudioStreamPlayback - the same chunked mechanism sdkapp's own
// /api-sdk/play_sound upload handler already proves works live, but parses
// the WAV header properly first instead of streaming it as if it were audio
// samples (that handler's own quirk - harmless there since the browser side
// always sends a fixed, known format, but not safe to copy blindly for an
// arbitrary file on disk).
func playSoundFile(path string, robot *vector.Vector) error {
	wavBytes, err := os.ReadFile(path)
	if err != nil {
		logger.Println("Sound file not found: " + path + " (" + err.Error() + ")")
		return nil
	}
	pcm, sampleRate, err := parseMonoPCMWav(wavBytes)
	if err != nil {
		logger.Println("Sound file could not be parsed: " + path + " (" + err.Error() + ")")
		return nil
	}

	var audioChunks [][]byte
	for len(pcm) >= 1024 {
		audioChunks = append(audioChunks, pcm[:1024])
		pcm = pcm[1024:]
	}
	if len(pcm) > 0 {
		audioChunks = append(audioChunks, pcm)
	}

	audioClient, err := robot.Conn.ExternalAudioStreamPlayback(context.Background())
	if err != nil {
		logger.Println("PlaySound stream open error: " + err.Error())
		return nil
	}
	if err := audioClient.Send(&vectorpb.ExternalAudioStreamRequest{
		AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamPrepare{
			AudioStreamPrepare: &vectorpb.ExternalAudioStreamPrepare{
				AudioFrameRate: sampleRate,
				AudioVolume:    100,
			},
		},
	}); err != nil {
		logger.Println("PlaySound prepare error: " + err.Error())
		return nil
	}
	for _, chunk := range audioChunks {
		if err := audioClient.Send(&vectorpb.ExternalAudioStreamRequest{
			AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamChunk{
				AudioStreamChunk: &vectorpb.ExternalAudioStreamChunk{
					AudioChunkSizeBytes: uint32(len(chunk)),
					AudioChunkSamples:   chunk,
				},
			},
		}); err != nil {
			logger.Println("PlaySound chunk send error: " + err.Error())
			return nil
		}
		time.Sleep(25 * time.Millisecond)
	}
	if err := audioClient.Send(&vectorpb.ExternalAudioStreamRequest{
		AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamComplete{
			AudioStreamComplete: &vectorpb.ExternalAudioStreamComplete{},
		},
	}); err != nil {
		logger.Println("PlaySound complete error: " + err.Error())
	}
	audioClient.CloseSend()
	logger.Println("Played sound: " + path)
	return nil
}

// parseMonoPCMWav extracts raw mono 16-bit PCM samples and the sample rate
// from a canonical 44-byte-header WAV file. Errors rather than guessing for
// anything else (stereo, non-16-bit, missing RIFF/WAVE/data markers).
func parseMonoPCMWav(b []byte) ([]byte, uint32, error) {
	if len(b) < 44 || string(b[0:4]) != "RIFF" || string(b[8:12]) != "WAVE" {
		return nil, 0, fmt.Errorf("not a RIFF/WAVE file")
	}
	channels := binary.LittleEndian.Uint16(b[22:24])
	sampleRate := binary.LittleEndian.Uint32(b[24:28])
	bitsPerSample := binary.LittleEndian.Uint16(b[34:36])
	if channels != 1 || bitsPerSample != 16 {
		return nil, 0, fmt.Errorf("expected mono 16-bit PCM, got %d channel(s) at %d bits", channels, bitsPerSample)
	}
	if string(b[36:40]) != "data" {
		return nil, 0, fmt.Errorf("expected canonical 44-byte header, no data chunk at offset 36")
	}
	dataLen := binary.LittleEndian.Uint32(b[40:44])
	if int(44+dataLen) > len(b) {
		return nil, 0, fmt.Errorf("declared data length exceeds file size")
	}
	return b[44 : 44+dataLen], sampleRate, nil
}

func DoSayText(input string, robot *vector.Vector) error {
	cleaned := removeSpecialCharacters(input)
	finalText := cleanRobotSpeechText(cleaned)
	if _, err := saveRobotSayTextDiagnostic(robot.Cfg.SerialNo, input, cleaned, finalText, time.Now().UTC()); err != nil {
		logger.Println("SAVE_ROBOT_SAY_TEXT_DIR write failed: " + err.Error())
	}
	if (vars.APIConfig.STT.Language != "en-US" && vars.APIConfig.Knowledge.Provider == "openai") || vars.APIConfig.Knowledge.OpenAIVoiceWithEnglish {
		return DoSayText_OpenAI(robot, finalText)
	}
	robot.Conn.SayText(context.Background(), &vectorpb.SayTextRequest{
		Text:           finalText,
		UseVectorVoice: true,
		DurationScalar: 0.95,
	})
	return nil
}

func pcmLength(data []byte) time.Duration {
	bytesPerSample := 2
	sampleRate := 16000
	numSamples := len(data) / bytesPerSample
	return time.Duration(numSamples*1000/sampleRate) * time.Millisecond
}

func getOpenAIVoice(voice string) openai.SpeechVoice {
	voiceMap := map[string]openai.SpeechVoice{
		"alloy":   openai.VoiceAlloy,
		"onyx":    openai.VoiceOnyx,
		"fable":   openai.VoiceFable,
		"shimmer": openai.VoiceShimmer,
		"nova":    openai.VoiceNova,
		"echo":    openai.VoiceEcho,
		"":        openai.VoiceFable,
	}
	return voiceMap[voice]
}

func DoSayText_OpenAI(robot *vector.Vector, input string) error {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	oc := openai.NewClient(vars.APIConfig.Knowledge.Key)
	resp, err := oc.CreateSpeech(context.Background(), openai.CreateSpeechRequest{
		Model:          openai.TTSModel1,
		Input:          input,
		Voice:          getOpenAIVoice(vars.APIConfig.Knowledge.OpenAIVoice),
		ResponseFormat: openai.SpeechResponseFormatPcm,
	})
	if err != nil {
		logger.Println(err)
		return err
	}
	speechBytes, _ := io.ReadAll(resp)
	vclient, err := robot.Conn.ExternalAudioStreamPlayback(context.Background())
	if err != nil {
		return err
	}
	vclient.Send(&vectorpb.ExternalAudioStreamRequest{
		AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamPrepare{
			AudioStreamPrepare: &vectorpb.ExternalAudioStreamPrepare{
				AudioFrameRate: 16000,
				AudioVolume:    100,
			},
		},
	})
	audioChunks := downsample24kTo16k(speechBytes)
	var chunksToDetermineLength []byte
	for _, chunk := range audioChunks {
		chunksToDetermineLength = append(chunksToDetermineLength, chunk...)
	}
	go func() {
		for _, chunk := range audioChunks {
			vclient.Send(&vectorpb.ExternalAudioStreamRequest{
				AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamChunk{
					AudioStreamChunk: &vectorpb.ExternalAudioStreamChunk{
						AudioChunkSizeBytes: 1024,
						AudioChunkSamples:   chunk,
					},
				},
			})
			time.Sleep(time.Millisecond * 25)
		}
		vclient.Send(&vectorpb.ExternalAudioStreamRequest{
			AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamComplete{
				AudioStreamComplete: &vectorpb.ExternalAudioStreamComplete{},
			},
		})
	}()
	time.Sleep(pcmLength(chunksToDetermineLength) + (time.Millisecond * 50))
	return nil
}

func DoGetImage(msgs []openai.ChatCompletionMessage, param string, robot *vector.Vector, stopStop chan bool) {
	stopImaging := false
	go func() {
		for range stopStop {
			stopImaging = true
			break
		}
	}()
	logger.Println("Get image here...")
	robot.Conn.EnableMirrorMode(context.Background(), &vectorpb.EnableMirrorModeRequest{Enable: true})
	for i := 3; i > 0; i-- {
		if stopImaging {
			return
		}
		time.Sleep(time.Millisecond * 300)
		robot.Conn.SayText(context.Background(), &vectorpb.SayTextRequest{
			Text:           fmt.Sprint(i),
			UseVectorVoice: true,
			DurationScalar: 1.05,
		})
	}
	resp, _ := robot.Conn.CaptureSingleImage(context.Background(), &vectorpb.CaptureSingleImageRequest{
		EnableHighResolution: true,
	})
	robot.Conn.EnableMirrorMode(context.Background(), &vectorpb.EnableMirrorModeRequest{Enable: false})
	go func() {
		robot.Conn.PlayAnimation(context.Background(), &vectorpb.PlayAnimationRequest{
			Animation: &vectorpb.Animation{Name: "anim_photo_shutter_01"},
			Loops:     1,
		})
	}()

	reqBase64 := base64.StdEncoding.EncodeToString(resp.Data)
	msgs = append(msgs, openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleUser,
		MultiContent: []openai.ChatMessagePart{{
			Type: openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{
				URL:    fmt.Sprintf("data:image/jpeg;base64,%s", reqBase64),
				Detail: openai.ImageURLDetailLow,
			},
		}},
	})

	var fullRespText string
	var fullfullRespText string
	var fullRespSlice []string
	var isDone bool
	var c *openai.Client
	switch vars.APIConfig.Knowledge.Provider {
	case "together":
		if vars.APIConfig.Knowledge.Model == "" {
			vars.APIConfig.Knowledge.Model = "meta-llama/Llama-2-70b-chat-hf"
			vars.WriteConfigToDisk()
		}
		conf := openai.DefaultConfig(vars.APIConfig.Knowledge.Key)
		conf.BaseURL = "https://api.together.xyz/v1"
		c = openai.NewClientWithConfig(conf)
	case "openai":
		c = openai.NewClient(vars.APIConfig.Knowledge.Key)
	case "custom":
		conf := openai.DefaultConfig(vars.APIConfig.Knowledge.Key)
		conf.BaseURL = vars.APIConfig.Knowledge.Endpoint
		c = openai.NewClientWithConfig(conf)
	}
	ctx := context.Background()
	speakReady := make(chan string)
	aireq := openai.ChatCompletionRequest{
		MaxTokens:        5120,
		Temperature:      1,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
		Messages:         msgs,
		Stream:           true,
	}
	if vars.APIConfig.Knowledge.Provider == "openai" {
		aireq.Model = openai.GPT4oMini
		logger.Println("Using " + aireq.Model)
	} else {
		aireq.Model = vars.APIConfig.Knowledge.Model
		logger.Println("Using " + aireq.Model)
	}
	if stopImaging {
		return
	}
	stream, err := c.CreateChatCompletionStream(ctx, aireq)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") && vars.APIConfig.Knowledge.Provider == "openai" {
			logger.Println("GPT model cannot be accessed with this API key.")
			logger.LogUI("GPT model cannot be accessed with this API key.")
			aireq.Model = openai.GPT3Dot5Turbo
			logger.Println("Falling back to " + aireq.Model)
			logger.LogUI("Falling back to " + aireq.Model)
			stream, err = c.CreateChatCompletionStream(ctx, aireq)
			if err != nil {
				logger.Println("OpenAI still not returning a response even after falling back. Erroring.")
				return
			}
		} else {
			logger.Println("LLM error: " + err.Error())
			return
		}
	}

	fmt.Println("LLM stream response: ")
	go func() {
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				isDone = true
				newStr := fullRespSlice[0]
				for i, str := range fullRespSlice {
					if i != 0 {
						newStr += " " + str
					}
				}
				if strings.TrimSpace(newStr) != strings.TrimSpace(fullfullRespText) {
					logger.Println("LLM debug: there is content after the last punctuation mark")
					fullRespSlice = append(fullRespSlice, strings.TrimPrefix(fullRespText, newStr))
				}
				if vars.APIConfig.Knowledge.SaveChat {
					Remember(msgs[len(msgs)-1], openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleAssistant,
						Content: newStr,
					}, robot.Cfg.SerialNo)
				}
				logger.LogDebugUI("LLM response for " + robot.Cfg.SerialNo + ": " + newStr)
				logger.Println("LLM stream finished")
				return
			}
			if err != nil {
				logger.Println("Stream error: " + err.Error())
				return
			}
			fullfullRespText += removeSpecialCharacters(response.Choices[0].Delta.Content)
			fullRespText += removeSpecialCharacters(response.Choices[0].Delta.Content)
			if strings.Contains(fullRespText, "...") || strings.Contains(fullRespText, ".'") || strings.Contains(fullRespText, ".\"") || strings.Contains(fullRespText, ".") || strings.Contains(fullRespText, "?") || strings.Contains(fullRespText, "!") {
				sepStr := "."
				for _, candidate := range []string{"...", ".'", ".\"", ".", "?", "!"} {
					if strings.Contains(fullRespText, candidate) {
						sepStr = candidate
						break
					}
				}
				splitResp := strings.Split(strings.TrimSpace(fullRespText), sepStr)
				fullRespSlice = append(fullRespSlice, strings.TrimSpace(splitResp[0])+sepStr)
				fullRespText = splitResp[1]
				select {
				case speakReady <- strings.TrimSpace(splitResp[0]) + sepStr:
				default:
				}
			}
		}
	}()
	numInResp := 0
	for {
		if stopImaging {
			return
		}
		respSlice := fullRespSlice
		if len(respSlice)-1 < numInResp {
			if !isDone {
				logger.Println("Waiting for more content from LLM...")
				for range speakReady {
					respSlice = fullRespSlice
					break
				}
			} else {
				break
			}
		}
		logger.Println(respSlice[numInResp])
		PerformActions(msgs, GetActionsFromString(respSlice[numInResp]), robot, stopStop)
		numInResp++
	}
}

func DoNewRequest(robot *vector.Vector) {
	time.Sleep(time.Second / 3)
	robot.Conn.AppIntent(context.Background(), &vectorpb.AppIntentRequest{Intent: "knowledge_question"})
}

func PerformActions(msgs []openai.ChatCompletionMessage, actions []RobotAction, robot *vector.Vector, stopStop chan bool) bool {
	result := performActionsWithGOPODLiveSpeechGate(msgs, actions, stopStop, liveRobotActionDispatcher{robot: robot})
	return result.DispatchReturn
}

func WaitForAnim_Queue(esn string) {
	for i, q := range AnimationQueues {
		if q.ESN == esn && q.AnimCurrentlyPlaying {
			for range AnimationQueues[i].AnimDone {
				break
			}
			return
		}
	}
}

func StartAnim_Queue(esn string) {
	for i, q := range AnimationQueues {
		if q.ESN == esn {
			if q.AnimCurrentlyPlaying {
				for range AnimationQueues[i].AnimDone {
					logger.Println("(waiting for animation to be done...)")
					break
				}
			} else {
				AnimationQueues[i].AnimCurrentlyPlaying = true
			}
			return
		}
	}
	AnimationQueues = append(AnimationQueues, AnimationQueue{
		ESN:                  esn,
		AnimDone:             make(chan bool),
		AnimCurrentlyPlaying: true,
	})
}

func StopAnim_Queue(esn string) {
	for i, q := range AnimationQueues {
		if q.ESN == esn {
			AnimationQueues[i].AnimCurrentlyPlaying = false
			select {
			case AnimationQueues[i].AnimDone <- true:
			default:
			}
		}
	}
}

type AnimationQueue struct {
	ESN                  string
	AnimDone             chan bool
	AnimCurrentlyPlaying bool
}

var AnimationQueues []AnimationQueue

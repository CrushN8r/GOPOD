package main

import (
	"context"
	"flag"
	"fmt"
	sdk_wrapper "github.com/fforchino/vector-go-sdk/pkg/sdk-wrapper"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"os"
	"strings"
	"time"
	"vectorx/pkg/intents"
	"vectorx/pkg/stats"
)

var Debug = true
var Ctx = context.Background()
var Start = make(chan bool)
var Stop = make(chan bool)

func main() {
	var serial = flag.String("serial", "", "Vector's Serial Number")
	var locale = flag.String("locale", "", "STT Locale in use")
	var speechText = flag.String("speechText", "", "Speech text")
	var gridSize int
	flag.IntVar(&gridSize, "grid-size", 75, "Deck size for Bingo. Must be 75 or 90.")
	var silent bool
	flag.BoolVar(&silent, "silent", false, "Suppress all robot speech (rattle and screen-display unaffected).")
	var chatFile string
	flag.StringVar(&chatFile, "chat-file",
		"/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/gopod_layer/web_display/gopod_demo_8011/live_chat_messages.json",
		"Absolute path to live_chat_messages.json. Default matches runtime/data_gomad/config/paths.json's live_chat_messages fact.")
	var pauseForTouch bool
	flag.BoolVar(&pauseForTouch, "pause-for-touch", false, "Wait for a fresh backpack touch before every draw instead of auto-drawing every 3 seconds after the first touch.")
	flag.Parse()

	if gridSize != 75 && gridSize != 90 {
		fmt.Fprintf(os.Stderr, "invalid --grid-size %d: must be 75 or 90\n", gridSize)
		os.Exit(2)
	}
	intents.BingoGridSize = gridSize
	intents.BingoSilentMode = silent
	intents.BingoChatFilePath = chatFile
	intents.BingoPauseForTouch = pauseForTouch

	if Debug {
		println("SERIAL: " + *serial)
		println("LOCALE: " + *locale)
		println("SPEECH TEXT: " + *speechText)
	}
	language := *locale
	if strings.Contains(language, "-") {
		language = strings.Split(language, "-")[0]
	}

	if len(*speechText) > 0 {
		// Remove "" if any
		if strings.HasPrefix(*speechText, "\"") && strings.HasSuffix(*speechText, "\"") {
			*speechText = (*speechText)[1 : len(*speechText)-1]
		}

		// Init SDK before intent match, so registration can use custom settings if needed
		err := sdk_wrapper.InitSDKForWirepod(*serial)
		if err != nil {
			println("FATAL: could not load Vector settings from JDOCS")
			return
		}

		// Register vectorx intents
		intents.RegisterIntents()

		// Find out whether the speech text matches any registered intent
		xIntent, err := intents.IntentMatch(*speechText, language)

		if err == nil {
			// Ok, we have a match. Then extract the parameters (if any) from the intent, but before that init the
			// SDK because we may need to get Vector's settings (e.g. for weather forecast we need the default location)

			stats.StatsIntentHandled(true)
			robotLocale := sdk_wrapper.GetLocale()
			if Debug {
				println("ROBOT LOCALE: " + robotLocale)
			}
			if robotLocale != *locale {
				if Debug {
					println("Different locales! Setting " + *locale)
				}
				sdk_wrapper.SetLocale(*locale)
			}
			if Debug {
				println("ROBOT LOCALE: " + robotLocale)
			}

			// Extract params
			intents.GetWirepodBotInfo(*serial)
			params := intents.ParseParams(*speechText, xIntent)

			engine, voice := sdk_wrapper.GetTTSConfiguration()
			sdk_wrapper.SetTTSEngine(engine)
			sdk_wrapper.SetTTSVoice(voice)

			behaviorErrors := make(chan error, 1)
			go func() {
				behaviorErrors <- runVectorXBehaviorControl(Ctx, Start, Stop)
			}()

			for {
				select {
				case <-Start:
					returnIntent := xIntent.Handler(xIntent, *speechText, params)
					// Seems that we have to force back en_US locale or "Hey Vector" won't work anymore
					sdk_wrapper.SetLocale("en-US")
					// Ok, intent handled. Return the intent that Wirepod has to send to the robot
					fmt.Println("{\"status\": \"ok\", \"returnIntent\": \"" + returnIntent + "\"}")
					if nil != xIntent.OSKRTriggersUserInput {
						if xIntent.OSKRTriggersUserInput() {
							time.Sleep(2 * time.Second)
							sdk_wrapper.TriggerWakeWord()
						}
					}
					Stop <- true
					if err := <-behaviorErrors; err != nil {
						println("ERROR: behavior control release failed: " + err.Error())
					} else {
						println("BINGO behavior control released")
					}
				case err := <-behaviorErrors:
					fmt.Println("{\"status\": \"ko\", \"returnIntent\": \"\"}")
					if err != nil {
						println("FATAL: behavior control failed: " + err.Error())
					}
				}
				return
			}
		} else {
			stats.StatsIntentHandled(false)
			// Intent cannot be handled by VectorX. Wirepod may continue its intent parsing chain
			fmt.Println("{\"status\": \"ko\", \"returnIntent\": \"\"}")
		}
	} else {
		// Intent cannot be handled by VectorX. Wirepod may continue its intent parsing chain
		sdk_wrapper.SetLocale("en-US")
		fmt.Println("{\"status\": \"ko\", \"returnIntent\": \"\"}")
		sdk_wrapper.SetLocale("en-US")
	}
}

func runVectorXBehaviorControl(ctx context.Context, start, stop chan bool) error {
	stream, err := sdk_wrapper.Robot.Conn.BehaviorControl(ctx)
	if err != nil {
		return err
	}

	if err := stream.Send(&vectorpb.BehaviorControlRequest{
		RequestType: &vectorpb.BehaviorControlRequest_ControlRequest{
			ControlRequest: &vectorpb.ControlRequest{
				Priority: vectorpb.ControlRequest_DEFAULT,
			},
		},
	}); err != nil {
		return err
	}

	for {
		response, err := stream.Recv()
		if err != nil {
			return err
		}
		if response.GetControlGrantedResponse() != nil {
			start <- true
			break
		}
	}

	<-stop
	return stream.Send(&vectorpb.BehaviorControlRequest{
		RequestType: &vectorpb.BehaviorControlRequest_ControlRelease{
			ControlRelease: &vectorpb.ControlRelease{},
		},
	})
}

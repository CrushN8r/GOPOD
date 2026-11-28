package wirepod_ttr

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/sashabaranov/go-openai"
)

func GetChat(esn string) vars.RememberedChat {
	for _, chat := range vars.RememberedChats {
		if chat.ESN == esn {
			return chat
		}
	}
	return vars.RememberedChat{
		ESN: esn,
	}
}

func PlaceChat(chat vars.RememberedChat) {
	for i, achat := range vars.RememberedChats {
		if achat.ESN == chat.ESN {
			vars.RememberedChats[i] = chat
			return
		}
	}
	vars.RememberedChats = append(vars.RememberedChats, chat)
}

// remember last 16 lines of chat
func Remember(user, ai openai.ChatCompletionMessage, esn string) {
	chatAppend := []openai.ChatCompletionMessage{
		user,
		ai,
	}
	currentChat := GetChat(esn)
	if len(currentChat.Chats) == 18 {
		var newChat vars.RememberedChat
		newChat.ESN = currentChat.ESN
		for i, chat := range currentChat.Chats {
			if i < 2 {
				continue
			}
			newChat.Chats = append(newChat.Chats, chat)
		}
		currentChat = newChat
	}
	currentChat.ESN = esn
	currentChat.Chats = append(currentChat.Chats, chatAppend...)
	PlaceChat(currentChat)
}

func isMn(r rune) bool {
	// Remove the characters that are not related to Vietnamese.
	// Retain the tonal marks and diacritics such as the circumflex, ơ, and ư in Vietnamese.
	keepMarks := []rune{
		'\u0300', // Dấu huyền
		'\u0301', // Dấu sắc
		'\u0303', // Dấu ngã
		'\u0309', // Dấu hỏi
		'\u0323', // Dấu nặng
		'\u0302', // Dấu mũ (â, ê, ô)
		'\u031B', // Dấu ơ và ư
		'\u0306', // Dấu trầm
	}
	if unicode.Is(unicode.Mn, r) {
		for _, mark := range keepMarks {
			if r == mark {
				return false
			}
		}
		return true
	}
	return false
}

func removeSpecialCharacters(str string) string {
	str = gopodSpeakSymbols(str)
	str = gopodLinkWords(str)

	// these two lines create a transformation that decomposes characters, removes non-spacing marks (like diacritics), and then recomposes the characters, effectively removing special characters
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	result, _, _ := transform.String(t, str)

	// Define the regular expression to match special characters
	result = gopodAmpersandRe.ReplaceAllString(result, " and ")
	result = gopodAsteriskWordGap(result)
	result = gopodAtSignRe.ReplaceAllString(result, "at ")
	result = gopodSayPlus(result)
	re := regexp.MustCompile(`[&^*#@]`)

	// Replace special characters with an empty string
	result = removeEmojis(gopodEmojiSentenceEnd(re.ReplaceAllString(result, "")))
	result = gopodEmojiJoinersRe.ReplaceAllString(result, "")

	// Replace special characters with ASCII
	// * COPY/PASTE TO ADD MORE CHARACTERS:
	//   result = strings.ReplaceAll(result, "", "")
	result = strings.ReplaceAll(result, "‘", "'")
	result = strings.ReplaceAll(result, "’", "'")
	result = strings.ReplaceAll(result, "“", "\"")
	result = strings.ReplaceAll(result, "”", "\"")
	result = strings.ReplaceAll(result, "—", ", ")
	result = gopodColonPauseRe.ReplaceAllString(result, "$1, ")
	result = gopodShortYearRe.ReplaceAllString(result, "$1$2")
	result = strings.ReplaceAll(result, "–", "-")
	result = strings.ReplaceAll(result, "…", ".")
	result = strings.ReplaceAll(result, "\u00A0", " ")
	result = strings.ReplaceAll(result, "•", "*")
	result = strings.ReplaceAll(result, "¼", "1/4")
	result = strings.ReplaceAll(result, "½", "1/2")
	result = strings.ReplaceAll(result, "¾", "3/4")
	result = strings.ReplaceAll(result, "ç", "c")
	result = strings.ReplaceAll(result, "©", "(c)")
	result = strings.ReplaceAll(result, "®", "(r)")
	result = strings.ReplaceAll(result, "™", "(tm)")
	result = strings.ReplaceAll(result, "@", "(a)")
	result = gopodSpokenDashes(result)
	result = gopodBrobotPossessiveRe.ReplaceAllStringFunc(result, func(m string) string {
		if strings.Contains(m, "1") {
			return "Brobot one's"
		}
		return "Brobot two's"
	})
	result = gopodAISpokenRe.ReplaceAllString(result, "AY-EYE")
	result = gopodSpaceBeforeCommaRe.ReplaceAllString(result, ",")
	result = gopodCommaAfterPunctRe.ReplaceAllString(result, "$1")
	result = gopodSpaceBeforeCloseRe.ReplaceAllString(result, "$1")
	result = gopodSpaceAfterOpenRe.ReplaceAllString(result, "$1")
	return result
}

var (
	gopodAISpokenRe         = regexp.MustCompile(`\bA\.I\.|\bAI\b`)
	gopodSpaceBeforeCommaRe = regexp.MustCompile(`[ \t]+,`)
	gopodCommaAfterPunctRe  = regexp.MustCompile(`([.!?;:]),`)
	gopodSpaceBeforeCloseRe = regexp.MustCompile(`[ \t]+([)\]])`)
	gopodAmpersandRe        = regexp.MustCompile(`[ \t]*&[ \t]*`)
	gopodBrobotPossessiveRe = regexp.MustCompile(`(?i)\bbrobot\s*[12]['\x{2019}]s\b`)
	gopodSpaceAfterOpenRe   = regexp.MustCompile(`([(\[])[ \t]+`)
)

func removeEmojis(input string) string {
	// a mess, but it works!
	re := regexp.MustCompile(`[\x{1FA00}-\x{1FAFF}]|[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E0}-\x{1F1FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]|[\x{1F900}-\x{1F9FF}]|[\x{1F004}]|[\x{1F0CF}]|[\x{1F18E}]|[\x{1F191}-\x{1F251}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]`)
	result := re.ReplaceAllString(input, "")
	return result
}

func CreateAIReq(transcribedText, esn string, gpt3tryagain, isKG bool) openai.ChatCompletionRequest {
	var nChat []openai.ChatCompletionMessage
	systemPrompt, promptBytes, usedOpenAIPrompt := loadCanonicalBrobotPrompt(transcribedText)
	esnLine := "Current robot ESN: " + esn
	if line, ok := gopodBrobotPersonaLines[gopodBrobotName(esn)]; ok {
		esnLine += ". " + line
	}

	smsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: systemPrompt,
	}

	var model string

	if gpt3tryagain {
		model = openai.GPT3Dot5Turbo
	} else if vars.APIConfig.Knowledge.Provider == "openai" {
		model = openai.GPT4oMini
		// Wire-Pod dc6ae0c (upstream #525): the configured knowledge model for OpenAI.
		if m := strings.TrimSpace(vars.APIConfig.Knowledge.Model); m != "" {
			model = m
		}
		logger.Println("Using " + model)
	} else {
		logger.Println("Using " + vars.APIConfig.Knowledge.Model)
		model = vars.APIConfig.Knowledge.Model
	}

	smsg.Content = CreatePrompt(smsg.Content, model, isKG)
	if !usedOpenAIPrompt {
		whoIAm := esnLine
		if line, ok := gopodBrobotPersonaLines[gopodAnswerer(gopodBrobotName(esn))]; ok {
			whoIAm = line
		}
		if !gopodAtDialZero(gopodBrobotName(esn)) {
			smsg.Content = gopodBrobotsPersonaFor(smsg.Content)
			if card := gopodPersonaProfileText(gopodAnswerer(gopodBrobotName(esn))); card != "" {
				whoIAm += "\n\n" + card
			}
		}
		smsg.Content += "\n\nWHO I AM:\n" + whoIAm
		intro := gopodIntroRecipe(gopodBrobotName(esn), transcribedText)
		gopodIntroFor.Store(esn, intro != "")
		gopodIntroWeatherFor.Store(esn, intro != "" && gopodIntroWeather(transcribedText))
		facts := strings.TrimSpace(gopodQuestionFactsFor(esn, transcribedText))
		if h := gopodHelpLookup(transcribedText); h != "" {
			facts = strings.TrimSpace(facts + "\n" + h)
		}
		gopodQuestionHasFacts.Store(esn, facts != "")
		if b := gopodBrainText(gopodBrobotName(esn), transcribedText); b != "" {
			smsg.Content += "\n\n" + b
		}
		if facts != "" {
			smsg.Content += "\n\n" + gopodBrobotsSettings.QuestionFactsHeader + "\n" + facts
		}
		if pre := gopodPrebuildContext(esn, transcribedText); pre != "" {
			smsg.Content += "\n\n" + pre
		}
		if cp := gopodChainPrompt(esn, transcribedText); cp != "" {
			smsg.Content += "\n\n" + cp
		}
		smsg.Content = string(appendPromptSection([]byte(smsg.Content), []byte(gopodReplyShapeSection(gopodBrobotName(esn)))))
		smsg.Content = string(appendPromptSection([]byte(smsg.Content), []byte(gopodFinalOutputGuard)))
	} else {
		smsg.Content = string(appendPromptSection([]byte(smsg.Content), []byte(gopodFinalOutputGuard)))
		smsg.Content += "\n\n" + esnLine
	}
	logger.Println(fmt.Sprintf("BROBOT_PROMPT_LOADED=%s,%s bytes=%d openai_prompt_used=%t esn_appended=%q", canonicalBrobotResponsePromptPath, canonicalBrobotIdentityPromptPath, promptBytes, usedOpenAIPrompt, esnLine))

	nChat = append(nChat, smsg)
	if vars.APIConfig.Knowledge.SaveChat {
		shared := gopodSharedBrobotChat(gopodAtDialZero(gopodBrobotName(esn)))
		logger.Println("Using shared Brobots chat, length of " + fmt.Sprint(len(shared)) + " messages")
		nChat = append(nChat, shared...)
	}
	if note := gopodAwarenessNote(esn); note != "" {
		nChat = append(nChat, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: note})
	}
	if gopodBrobotName(esn) != "Brobot" && gopodIsThanks(transcribedText) {
		nChat = append(nChat, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: gopodBrobotsSettings.ChatHold.ThanksNote})
	}
	question := gopodTagForBrobot(esn, transcribedText)
	if intro := gopodIntroRecipe(gopodBrobotName(esn), transcribedText); intro != "" {
		question += "\n\n" + intro
	}
	nChat = append(nChat, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
	})

	aireq := openai.ChatCompletionRequest{
		Model:            model,
		MaxTokens:        5120,
		Temperature:      1,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
		Messages:         nChat,
		Stream:           true,
	}
	if gopodBrobotName(esn) != "Brobot" {
		if t := gopodBrobotsSettings.LLM.Temperature; t > 0 {
			aireq.Temperature = t
		}
		if t := gopodBrobotsSettings.LLM.SayLineTemperature; t > 0 && gopodSayLine(esn) != "" {
			aireq.Temperature = t
			logger.Println(fmt.Sprintf("BROBOT_TEMPERATURE=%.2f say_line", t))
		}
	}
	return aireq
}

func StreamingKGSim(req interface{}, esn string, transcribedText string, isKG bool) (string, error) {
	if gopodNeverIn(transcribedText, gopodBrobotName(esn)+"'s ears") {
		return gopodBrobotsSettings.CleanPolitics.NeverInLine, nil
	}
	if _, following := gopodFollowingUp.Load(esn); !following {
		gopodExchangeWait(esn, transcribedText)
	}
	photoKind := ""
	var mathSteps []string
	mathOn, mathLead, mathFollowStarted := false, false, false
	_, allLead := gopodAllFirstAnswer.Load(esn)
	photoTaken := make(chan struct{})
	var photoAnswer string
	var photoRound2 func()
	// GOPOD: intent matching has already seen the raw transcript; from here on the
	// LLM and the shared chat get GOPOD's names spelled right ("go pod" -> GOPOD).
	transcribedText = gopodSpellNames(transcribedText)
	if req != nil && gopodBrobotName(esn) != "Brobot" && gopodWarmupHold(esn, transcribedText, gopodMicHow) {
		IntentPass(req, "intent_system_noaudio", transcribedText, map[string]string{}, false)
		return "", nil
	}
	if gopodWarmGateDrop(esn, "voice", transcribedText) {
		if req != nil {
			IntentPass(req, "intent_system_noaudio", transcribedText, map[string]string{}, false)
		}
		return "", nil
	}
	if req != nil && gopodBrobotName(esn) != "Brobot" && gopodThanksCloseOut(esn, transcribedText, "voice") {
		IntentPass(req, "intent_system_noaudio", transcribedText, map[string]string{}, false)
		return "", nil
	}
	if req != nil && gopodBrobotName(esn) != "Brobot" && !gopodIsThanks(transcribedText) && gopodRunGateDrop(esn, "voice", transcribedText) {
		IntentPass(req, "intent_system_noaudio", transcribedText, map[string]string{}, false)
		return "", nil
	}
	if gopodBrobotName(esn) != "Brobot" && !gopodIsThanks(transcribedText) {
		gopodReporterNewQuestion()
		gopodCoverageNewQuestion()
	}
	start := make(chan bool)
	stop := make(chan bool)
	stopStop := make(chan bool)
	kgReadyToAnswer := make(chan bool)
	kgStopLooping := false
	ctx := context.Background()
	matched := false
	var robot *vector.Vector
	var guid string
	var target string
	for _, bot := range vars.BotInfo.Robots {
		if esn == bot.Esn {
			guid = bot.GUID
			target = bot.IPAddress + ":443"
			matched = true
			break
		}
	}
	if matched {
		var err error
		robot, err = vector.New(vector.WithSerialNo(esn), vector.WithToken(guid), vector.WithTarget(target))
		if err != nil {
			return err.Error(), err
		}
		robot = GopodRobotsOffConn(robot)  // GOPOD: robots off (gopod_robots_off.go)
	}
	_, err := robot.Conn.BatteryState(context.Background(), &vectorpb.BatteryStateRequest{})
	if err != nil {
		return "", err
	}
	decidedAt := time.Now()
	if gopodBrobotName(esn) != "Brobot" {
		if GopodRecordSessionAnswer(transcribedText, gopodVoiceCommandHow) {
			if !isKG && req != nil {
				gopodLLMClosePass.Store(esn, true) // GOPOD: closes the voice request, not a native intent
				IntentPass(req, "intent_greeting_hello", "", map[string]string{}, false)
			}
			return "", nil
		}
		gopodChatHoldBeforeAnswer(esn)
		answer, to := gopodShouldAnswer(esn, transcribedText, decidedAt)
		gopodAddressingLog(esn, to, answer)
		if !answer {
			gopodQuietBrother(req, robot, esn, transcribedText, isKG)
			gopodChatHoldResume(esn)
			return "", nil
		}
		transcribedText = gopodCanonicalCall(transcribedText)
		if first, second, ok := gopodMathParts(transcribedText); ok {
			_, follow := gopodMathFollowUp.LoadAndDelete(esn)
			lead := gopodMathLead(transcribedText)
			if esn != lead && !follow {
				logger.Println("GOPOD_MATH " + gopodBrobotName(lead) + " answers first: " + transcribedText)
				gopodAnswerNext.Store(lead, true)
				go StreamingKGSim(nil, lead, transcribedText, false)
				gopodQuietBrother(req, robot, esn, transcribedText, isKG)
				gopodChatHoldResume(esn)
				return "", nil
			}
			mathOn = true
			if esn == lead {
				follower := gopodBrotherOf(lead)
				gopodExchangeOpen(transcribedText, follower) // let go when he has said his follow-up
				// An answer that stops before its follow-up starts lets the line go.
				defer func() {
					if !mathFollowStarted {
						gopodExchangeDone(follower)
					}
				}()
				mathSteps, mathLead = first, true
			} else {
				mathSteps = second
			}
		}
		// A Record Session starts only between replies.
		gopodReplyStart(esn)
		photoKind = gopodPhotoKind(transcribedText)
	}
	searchWhileWaiting := isKG || gopodBrobotName(esn) != "Brobot"
	// The searching gesture already ended with its get-out (no second one at the answer).
	var searchingClosed atomic.Bool
	if searchWhileWaiting {
		isBrobot := gopodBrobotName(esn) != "Brobot"
		controlGranted := start
		if isBrobot {
			controlGranted = make(chan bool)
		}
		gopodLogBControl(robot, ctx, controlGranted, stop)
		go func() {
			if isBrobot {
				<-controlGranted
				var photoStep func()
				if photoKind == "ask" {
					photoStep = func() {
						photoAnswer, photoRound2 = gopodPhotoTake(robot, ctx, esn)
						close(photoTaken)
					}
				} else {
					close(photoTaken)
				}
				go func() { start <- true }()
				gopodThinkingProcess(robot, ctx, esn, true, func() bool { return kgStopLooping }, photoStep)
			} else {
				gopodThinkingProcess(robot, ctx, esn, false, func() bool { return kgStopLooping })
			}
			searchingClosed.Store(true)
			kgReadyToAnswer <- true
		}()
	}
	var rawRespText string
	var fullRespSlice []string
	var streamResult gopodStreamResult
	var sentCommandPackets []string
	var isDone bool
	var streamMu sync.Mutex
	var markerZeroSeen bool
	var markerOneSeen bool
	var speechCursor int
	var pieceOfSpoken []int // GOLDEN highlight: bubble piece per SayText, in order
	var replyDisplay string
	var replyShown sync.Once
	showReply := func() {
		replyShown.Do(func() {
			streamMu.Lock()
			text := replyDisplay
			streamMu.Unlock()
			if text != "" {
				emitFinalLLMDisplay(esn, text)
			}
		})
	}
	spokenCount := 0
	weatherShown := false // GOPOD: his weather line's picture, once
	var interrupted bool
	stopAt := gopodStopSeq.Load() // GOPOD: a stop run since this answer began ends it
	markRawResponseLocked := func(result gopodStreamResult) {
		if !markerZeroSeen {
			markerZeroSeen = true
			logger.LogDebugUI("GOPOD_STREAM_MARKER_0")
			emitRawLLMDebug(result.Raw)
			emitMarkedPacketDebug(result.MarkedPacket)
			emitFilterPacketDebug(result.FilteredPacket)
		}
	}
	finalizeMarkerOneLocked := func(reason string) bool {
		if reason != "normal" {
			logger.LogDebugUI("GOPOD_STREAM_MARKER_1_DEFERRED_REASON=" + reason)
			return false
		}
		if markerOneSeen {
			return false
		}
		markerOneSeen = true
		markerPos := speechCursor
		if markerPos > len(fullRespSlice) {
			markerPos = len(fullRespSlice)
		}
		markerPos = len(fullRespSlice)
		speechCursor = markerPos
		logger.LogDebugUI("GOPOD_STREAM_MARKER_1")
		logger.LogDebugUI("BROBOT_FINALIZE_REASON=marker_1")
		return true
	}
	snapshotStream := func() ([]string, bool, bool) {
		streamMu.Lock()
		defer streamMu.Unlock()
		respSlice := append([]string(nil), fullRespSlice...)
		return respSlice, isDone, interrupted
	}
	var c *openai.Client
	switch vars.APIConfig.Knowledge.Provider {
	case "together":
		if vars.APIConfig.Knowledge.Model == "" {
			vars.APIConfig.Knowledge.Model = "meta-llama/Llama-3-70b-chat-hf"
			vars.WriteConfigToDisk()
		}
		conf := openai.DefaultConfig(vars.APIConfig.Knowledge.Key)
		conf.BaseURL = "https://api.together.xyz/v1"
		c = openai.NewClientWithConfig(conf)
	case "custom":
		conf := openai.DefaultConfig(vars.APIConfig.Knowledge.Key)
		conf.BaseURL = vars.APIConfig.Knowledge.Endpoint
		c = openai.NewClientWithConfig(conf)
	case "openai":
		c = openai.NewClient(vars.APIConfig.Knowledge.Key)
	}
	speakReady := make(chan string)
	successIntent := make(chan bool)

	aireq := CreateAIReq(transcribedText, esn, false, isKG)
	gopodValuePointsNoticePost(esn)
	if gopodBrobotName(esn) != "Brobot" && !gopodIsFollowingUp(esn) {
		gopodPictureDescribeOnAsk(transcribedText)
	}
	if note := gopodPictureNote(); note != "" && !gopodScriptPlaying.Load() && len(aireq.Messages) > 0 {
		last := &aireq.Messages[len(aireq.Messages)-1]
		last.Content = note + "\n" + last.Content
	}
	if gopodIsFollowingUp(esn) && len(aireq.Messages) > 0 {
		note := strings.TrimSpace(gopodBrobotsSettings.Addressing.FollowUpNote)
		if m := strings.TrimSpace(gopodBrobotsSettings.Math.FollowUpNote); mathOn && m != "" {
			note = m
		}
		if note != "" {
			last := &aireq.Messages[len(aireq.Messages)-1]
			last.Content = note + "\n" + last.Content
		}
	}

	var stream gopodChatStream
	scriptedAnswer, scripted := gopodScriptAnswer(esn, transcribedText)
	if !scripted && photoKind == "recall" {
		scriptedAnswer, scripted = gopodPhotoRecall(esn), true
	} else if !scripted && photoKind == "ask" && searchWhileWaiting {
		<-photoTaken
		scriptedAnswer, scripted = photoAnswer, true
	}
	if scripted {
		logger.Println("GOPOD_SCRIPT answer for " + esn)
		stream = &gopodScriptedStream{text: scriptedAnswer}
	} else {
		stream, err = c.CreateChatCompletionStream(ctx, aireq)
	}
	if err != nil {
		log.Printf("Error creating chat completion stream: %v", err)
		if strings.Contains(err.Error(), "does not exist") && vars.APIConfig.Knowledge.Provider == "openai" {
			logger.Println("GPT model cannot be accessed with this API key.")
			logger.LogUI("GPT model cannot be accessed with this API key.")
			aireq := CreateAIReq(transcribedText, esn, true, isKG)
			logger.Println("Falling back to " + aireq.Model)
			logger.LogUI("Falling back to " + aireq.Model)
			stream, err = c.CreateChatCompletionStream(ctx, aireq)
			if err != nil {
				logger.Println("OpenAI still not returning a response even after falling back. Erroring.")
				return "", err
			}
		} else {
			if searchWhileWaiting {
				kgStopLooping = true
				for range kgReadyToAnswer {
					break
				}
				stop <- true
				time.Sleep(time.Second / 3)
				if isKG {
					KGSim(esn, "There was an error getting data from the L. L. M.")
				}
			}
			return "", err
		}
	}
	nChat := aireq.Messages
	nChat = append(nChat, openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleAssistant,
	})
	fmt.Println("LLM stream response: ")
	go func() {
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				streamMu.Lock()
				// prevents a crash
				if strings.TrimSpace(rawRespText) == "" {
					streamMu.Unlock()
					logger.Println("LLM returned no response")
					successIntent <- false
					if searchWhileWaiting {
						kgStopLooping = true
						for range kgReadyToAnswer {
							break
						}
						stop <- true
						time.Sleep(time.Second / 3)
						if isKG {
							KGSim(esn, "There was an error getting data from the L. L. M.")
						}
					}
					break
				}
				if !scripted {
					rawRespText = gopodFixSelfName(esn, gopodCorrectNames(gopodStripLeadingAnimationWords(rawRespText)))
					rawRespText = gopodDropWebLinks(rawRespText)
					rawRespText = gopodDropLLMReminder(rawRespText)
					rawRespText = gopodDropQuestionEcho(rawRespText, transcribedText)
					rawRespText = gopodCleanPolitics(rawRespText, transcribedText)
					if gopodIsFollowingUp(esn) {
						gopodBrotherWords.Lock()
						brother := gopodBrotherWords.last[gopodBrotherOf(esn)]
						gopodBrotherWords.Unlock()
						rawRespText = gopodDropBrotherRepeat(rawRespText, brother)
					}
					rawRespText = gopodTightenReply(rawRespText, transcribedText, esn)
					if line := gopodNonsenseLine(transcribedText); line != "" {
						rawRespText = line
					}
					rawRespText = gopodThoughtBreaks(rawRespText)
					gopodRememberOwnWords(esn, rawRespText)
					if mathOn {
						extra := ""
						if !mathLead {
							extra = gopodMathDecimalLine(transcribedText)
						}
						rawRespText = gopodMathReply(rawRespText, mathSteps, extra)
						logger.Println("GOPOD_MATH " + gopodBrobotName(esn) + ": " + rawRespText)
					} else {
						rawRespText = gopodShapeBrobotAnswer(esn, transcribedText, rawRespText)
					}
				}
				displayRespText := gopodDisplayReply(rawRespText)
				var displayPieces []int // GOLDEN highlight: spoken piece -> bubble piece (gopod_highlight_pieces.go)
				rawRespText = gopodAnimationMarkerToInternal(gopodLinkWords(rawRespText))
				streamResult = buildGOPODStreamResultWithRepair(ctx, rawRespText, aireq.Model, openAIGOPODRobotSpeechRepairer(c))
				fullRespSlice = append([]string(nil), streamResult.AudibleChunks...)
				if streamResult.Fallback {
					v := validateGOPODRobotSpeech(strings.TrimSpace(rawRespText))
					logger.Println("BROBOT_FALLBACK " + gopodBrobotName(esn) + ": the fallback line said and shown - failed " + v.FailureType + " at \"" + v.FirstFailingLine + "\" - reply: " + strings.ReplaceAll(rawRespText, "\n", " / "))
				}
				displayRespText = gopodDisplayForSpeech(displayRespText, streamResult)
				if gopodBrobotName(esn) != "Brobot" {
					displayRespText, displayPieces = gopodSegmentDisplay(displayRespText, gopodSpokenPieceTexts(fullRespSlice))
					pieceOfSpoken = displayPieces
				}
				markRawResponseLocked(streamResult)
				isDone = true
				newStr := streamResult.Raw
				streamMu.Unlock()
				recordCompletedGOPODSessionExchange(esn, transcribedText, streamResult, time.Now().UTC())
				if vars.APIConfig.Knowledge.SaveChat {
					Remember(openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleUser,
						Content: transcribedText,
					},
						openai.ChatCompletionMessage{
							Role:    openai.ChatMessageRoleAssistant,
							Content: newStr,
						},
						esn)
				}
				if !isKG && req != nil {
					gopodLLMClosePass.Store(esn, true) // GOPOD: closes the voice request, not a native intent
					IntentPass(req, "intent_greeting_hello", transcribedText, map[string]string{}, false)
				}
				if gopodBrobotName(esn) != "Brobot" {
					gopodShowChapter(transcribedText)
					if req != nil {
						gopodShowTopicPictures(esn, transcribedText)
					}
				}
				if gopodBrobotName(esn) != "Brobot" {
					streamMu.Lock()
					replyDisplay = displayRespText
					streamMu.Unlock()
				} else {
					emitFinalLLMDisplay(esn, displayRespText)
				}
				logger.Println("LLM stream finished")
				select {
				case successIntent <- true:
				default:
				}
				select {
				case speakReady <- "":
				default:
				}
				return
			}

			if err != nil {
				logger.Println("Stream error: " + err.Error())
				return
			}

			if len(response.Choices) == 0 {
				logger.Println("Empty response")
				return
			}

			rawDelta := response.Choices[0].Delta.Content
			streamMu.Lock()
			rawRespText += rawDelta
			streamMu.Unlock()
		}
	}()
	for is := range successIntent {
		if is {
			// IntentPass itself now fires inside the streaming goroutine above,
			// right before emitFinalLLMDisplay - so "Intent matched" logs before
			// "LLM response for", matching real request order (transcribe -> intent
			// -> LLM reply) instead of the channel-handshake order. This loop still
			// exists purely to block until that goroutine signals done/failed.
			break
		} else {
			return "", errors.New("llm returned no response")
		}
	}
	time.Sleep(time.Millisecond * 200)
	if !searchWhileWaiting {
		gopodLogBControl(robot, ctx, start, stop)
	}
	go func() {
		InterruptKGSimWhenTouchedOrWaked(robot, stop, stopStop, func(source string) bool {
			markerAdvanceRequest := NewGOPODMarkerAdvanceRequest(source)
			streamMu.Lock()
			defer streamMu.Unlock()
			// Wires the marker-advance machinery (kgsim_marker_advance_control.go) that was
			// built for exactly this and had zero production callers (WIRED-POD.md,
			// "Marker advance control"). ShouldAdvance()/MarkerOneEvents() replace the old
			// hardcoded `return false`, which vetoed every touch/wake/keypad interrupt
			// permanently (InterruptKGSimWhenTouchedOrWaked returns without ever sending
			// `stop <- true` when this callback returns false) - not a deferral, a silent
			// no-op. True native touch-interrupt behavior, restored via the thin wrapper.
			emitGOPODMarkerAdvanceEvents(func(event string) { logger.LogDebugUI(event) }, markerAdvanceRequest)
			if markerAdvanceRequest.ShouldAdvance() {
				interrupted = true
				// The chat page's sentence outline goes solid (gopod_speaking_now.go).
				gopodMarkSpeakingInterrupted(robot.Cfg.SerialNo)
			}
			return markerAdvanceRequest.ShouldAdvance()
		})
	}()
	var TTSLoopAnimation string
	var TTSGetinAnimation string
	if searchWhileWaiting {
		TTSLoopAnimation = "anim_knowledgegraph_answer_01"
		TTSGetinAnimation = "anim_knowledgegraph_searching_getout_01"
	} else {
		TTSLoopAnimation = "anim_tts_loop_02"
		TTSGetinAnimation = "anim_getin_tts_01"
	}

	var stopTTSLoop bool
	TTSLoopStopped := make(chan bool)
	for range start {
		if searchWhileWaiting {
			kgStopLooping = true
			gopodReceiveOrGiveUp(kgReadyToAnswer, robot.Cfg.SerialNo, "searching loop")
		} else {
			time.Sleep(time.Millisecond * 300)
		}
		if !searchingClosed.Load() {
			gopodLogPlayAnimation(robot,
				ctx,
				&vectorpb.PlayAnimationRequest{
					Animation: &vectorpb.Animation{
						Name: TTSGetinAnimation,
					},
					Loops: 1,
				},
			)
		}
		if !vars.APIConfig.Knowledge.CommandsEnable {
			go func() {
				for {
					if stopTTSLoop {
						TTSLoopStopped <- true
						break
					}
					gopodLogPlayAnimation(robot,
						ctx,
						&vectorpb.PlayAnimationRequest{
							Animation: &vectorpb.Animation{
								Name: TTSLoopAnimation,
							},
							Loops: 1,
						},
					)
				}
			}()
		}
		var disconnect bool
		numInResp := 0
		var armCueDone chan struct{}
		for {
			respSlice, done, wasInterrupted := snapshotStream()
			if len(respSlice)-1 < numInResp {
				if !done {
					logger.Println("Waiting for more content from LLM...")
					for range speakReady {
						respSlice, done, wasInterrupted = snapshotStream()
						break
					}
				} else {
					break
				}
			}
			if wasInterrupted {
				break
			}
			streamMu.Lock()
			if numInResp >= len(fullRespSlice) || interrupted {
				streamMu.Unlock()
				continue
			}
			currentChunk := fullRespSlice[numInResp]
			speechCursor = numInResp + 1
			currentFullResponse := streamResult.Raw
			streamMu.Unlock()
			logger.Println(currentChunk)
			logger.LogDebugUI("FILTERED_SPLIT_READY:\n" + currentChunk)
			commandPacket := gopodPronounceCommandPacket(commandPacketFromFilteredChunk(currentChunk))
			commandPacket = gopodVaryDefaultAnimation(commandPacket, currentFullResponse)
			logger.LogDebugUI("FILTERED_SPLIT_SENT:\n" + commandPacket)
			streamMu.Lock()
			sentCommandPackets = append(sentCommandPackets, commandPacket)
			streamMu.Unlock()
			acts := GetActionsFromString(commandPacket)
			for i := range acts {
				if acts[i].Action == ActionSayText {
					acts[i].Parameter = gopodEndWithPunctuation(acts[i].Parameter)
				}
			}
			nChat[len(nChat)-1].Content = currentFullResponse
			prevWasSpeech := false
			prevWasAnimation := false
			showReply()
			if gopodBrobotName(esn) != "Brobot" {
				gopodWaitReplyShown(esn)
			}
			for _, act := range gopodAnimationsFromWords(gopodSprinkleAnimations(gopodSplitSpeechIntoSentences(acts))) {
				streamMu.Lock()
				stopNow := interrupted || gopodStopSeq.Load() != stopAt
				streamMu.Unlock()
				if stopNow {
					break
				}
				if act.Action == ActionPlayAnimationWI {
					gopodAnimationWaitDone(robot.Cfg.SerialNo)
					logger.LogDebugUI("GOPOD_SAY_ANIMATION: " + act.Parameter)
					if armCueDone == nil && prevWasSpeech {
						armCueDone = gopodStartReplyMove(robot, robot.Cfg.SerialNo)
					}
				}
				if act.Action == ActionSayText && prevWasAnimation {
					time.Sleep(gopodSecondsToDuration(gopodBrobotsSettings.ChatMoves.SpeechAfterAnimationSeconds))
				}
				if act.Action == ActionSayText && prevWasAnimation && gopodAnimationLeadIn != "" &&
					!strings.HasPrefix(strings.TrimSpace(act.Parameter), gopodAnimationLeadIn) {
					// Right after an animation: the native lead-in, so the robot's
					// voice pauses while the clip starts (gopodAnimationLeadIn).
					act.Parameter = gopodAnimationLeadIn + " " + act.Parameter
				}
				prevWasSpeech = act.Action == ActionSayText
				prevWasAnimation = act.Action == ActionPlayAnimationWI
				if act.Action == ActionSayText {
					logger.LogDebugUI("GOPOD_SAY_SENTENCE: " + act.Parameter)
					gopodScriptMoveAt(robot, robot.Cfg.SerialNo, act.Parameter)
					piece := -1
					streamMu.Lock()
					if spokenCount < len(pieceOfSpoken) {
						piece = pieceOfSpoken[spokenCount]
					}
					spokenCount++
					streamMu.Unlock()
					gopodWriteSpeakingPiece(robot.Cfg.SerialNo, act.Parameter, piece, true, false)
				}
				actionStop := make(chan bool)
				actStart := time.Now()
				disconnect = gopodPerformOrGiveUp(nChat, act, robot, actionStop)
				close(actionStop)
				if act.Action == ActionSayText {
					logger.LogDebugUI(fmt.Sprintf("GOPOD_VECTOR_PHRASE: %s done took=%dms", gopodBrobotName(robot.Cfg.SerialNo), time.Since(actStart).Milliseconds()))
					if !weatherShown && gopodWeatherLineEnd(act.Parameter) {
						weatherShown = true
						gopodWeatherLineSaid(robot.Cfg.SerialNo)
					}
				}
				if disconnect {
					break
				}
			}
			if disconnect {
				break
			}
			numInResp = numInResp + 1
		}
		if armCueDone != nil {
			gopodWaitOrGiveUp(armCueDone, robot.Cfg.SerialNo, "arm cue")
		}
		streamMu.Lock()
		stoppedByInterrupt := interrupted
		streamMu.Unlock()
		if stoppedByInterrupt {
			gopodMarkSpeakingInterrupted(robot.Cfg.SerialNo)
		} else {
			gopodWriteSpeakingNow(robot.Cfg.SerialNo, "", false, false)
		}
		if spokenCount > 0 {
			gopodSetBrobotEyes(robot, robot.Cfg.SerialNo, false)
		}
		if !mathLead && !allLead {
			gopodGameWinCheck()
		}
		// A reply that never got to speak still shows its bubble.
		showReply()
		if gopodBrobotName(robot.Cfg.SerialNo) != "Brobot" {
			gopodLogBattery(robot)
		}
		streamMu.Lock()
		if !interrupted {
			finalizeMarkerOneLocked("normal")
		}
		streamMu.Unlock()
		if !vars.APIConfig.Knowledge.CommandsEnable {
			stopTTSLoop = true
			gopodReceiveOrGiveUp(TTSLoopStopped, robot.Cfg.SerialNo, "tts loop")
		}
		time.Sleep(time.Millisecond * 100)
		// if isKG {
		// 	robot.Conn.PlayAnimation(
		// 		ctx,
		// 		&vectorpb.PlayAnimationRequest{
		// 			Animation: &vectorpb.Animation{
		// 				Name: "anim_knowledgegraph_success_01",
		// 			},
		// 			Loops: 1,
		// 		},
		// 	)
		// 	time.Sleep(time.Millisecond * 3300)
		// }
		streamMu.Lock()
		wasInterrupted := interrupted
		diagnosticResult := streamResult
		diagnosticSentPackets := append([]string(nil), sentCommandPackets...)
		streamMu.Unlock()
		logger.LogDebugUI("BROBOT_RELEASE_READY=yes")
		if _, err := saveBrobotResponseDiagnostics(robot.Cfg.SerialNo, diagnosticResult, diagnosticSentPackets, time.Now().UTC()); err != nil {
			logger.Println("SAVE_ROBOT_SAY_TEXT_DIR write failed: " + err.Error())
		}
		gopodSendOrGiveUp(stopStop, robot.Cfg.SerialNo, "interrupt listener")
		if !wasInterrupted {
			gopodSendOrGiveUp(stop, robot.Cfg.SerialNo, "behavior control release")
		}
		if gopodBrobotName(esn) != "Brobot" {
			if photoRound2 != nil && !wasInterrupted {
				photoRound2()
			}
			if !wasInterrupted {
				gopodManagerQRShow()
			}
			gopodReplyEnd(esn)
			gopodExchangeDone(esn)
			gopodFollowingUp.Delete(esn)
			if mathLead {
				mathFollowStarted = true
				go gopodMathFollowUpStart(transcribedText)
			}
			if gopodIsThanks(transcribedText) {
				GopodChatHoldEnd("thanks")
				// The ending, in order: the score, the reporters' closing, the brothers' extra lines,
				// keep or delete, "Draft run completed." and "You're welcome."
				// (gopod_thanks_close_out.go).
				go gopodThanksEnding()
			} else {
				gopodChatHoldAfterAnswer()
			}
		}
	}
	return "", nil
}

func KGSim(esn string, textToSay string) error {
	ctx := context.Background()
	matched := false
	var robot *vector.Vector
	var guid string
	var target string
	for _, bot := range vars.BotInfo.Robots {
		if esn == bot.Esn {
			guid = bot.GUID
			target = bot.IPAddress + ":443"
			matched = true
			break
		}
	}
	if matched {
		var err error
		robot, err = vector.New(vector.WithSerialNo(esn), vector.WithToken(guid), vector.WithTarget(target))
		if err != nil {
			return err
		}
		robot = GopodRobotsOffConn(robot)  // GOPOD: robots off (gopod_robots_off.go)
	}
	controlRequest := &vectorpb.BehaviorControlRequest{
		RequestType: &vectorpb.BehaviorControlRequest_ControlRequest{
			ControlRequest: &vectorpb.ControlRequest{
				Priority: vectorpb.ControlRequest_OVERRIDE_BEHAVIORS,
			},
		},
	}
	go func() {
		start := make(chan bool)
		stop := make(chan bool)

		go func() {
			// * begin - modified from official vector-go-sdk
			r, err := robot.Conn.BehaviorControl(
				ctx,
			)
			if err != nil {
				log.Println(err)
				return
			}

			if err := r.Send(controlRequest); err != nil {
				log.Println(err)
				return
			}

			for {
				ctrlresp, err := r.Recv()
				if err != nil {
					log.Println(err)
					return
				}
				if ctrlresp.GetControlGrantedResponse() != nil {
					start <- true
					break
				}
			}

			for {
				select {
				case <-stop:
					logger.Println("KGSim: releasing behavior control (interrupt)")
					if err := r.Send(
						&vectorpb.BehaviorControlRequest{
							RequestType: &vectorpb.BehaviorControlRequest_ControlRelease{
								ControlRelease: &vectorpb.ControlRelease{},
							},
						},
					); err != nil {
						log.Println(err)
						return
					}
					return
				default:
					continue
				}
			}
			// * end - modified from official vector-go-sdk
		}()

		var stopTTSLoop bool
		var TTSLoopStopped bool
		for range start {
			time.Sleep(time.Millisecond * 300)
			gopodLogPlayAnimation(robot,
				ctx,
				&vectorpb.PlayAnimationRequest{
					Animation: &vectorpb.Animation{
						Name: "anim_getin_tts_01",
					},
					Loops: 1,
				},
			)
			go func() {
				for {
					if stopTTSLoop {
						TTSLoopStopped = true
						break
					}
					gopodLogPlayAnimation(robot,
						ctx,
						&vectorpb.PlayAnimationRequest{
							Animation: &vectorpb.Animation{
								Name: "anim_tts_loop_02",
							},
							Loops: 1,
						},
					)
				}
			}()
			cleaned := strings.TrimSpace(removeSpecialCharacters(textToSay))
			if cleaned != "" {
				_, err := gopodLogSayText(robot,
					ctx,
					&vectorpb.SayTextRequest{
						Text:           cleaned,
						UseVectorVoice: true,
						DurationScalar: 1.0,
					},
				)
				if err != nil {
					logger.Println("KG SayText error: " + err.Error())
					stop <- true
				}
			}
			stopTTSLoop = true
			for {
				if TTSLoopStopped {
					break
				} else {
					time.Sleep(time.Millisecond * 10)
				}
			}
			time.Sleep(time.Millisecond * 100)
			//time.Sleep(time.Millisecond * 3300)
			stop <- true
		}
	}()
	return nil
}

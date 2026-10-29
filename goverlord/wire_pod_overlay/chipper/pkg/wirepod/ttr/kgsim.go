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
	// GOPOD 2026-10-08: math symbols, "$", "#" and LaTeX math are said in words
	// (spoken_symbols, gopod_spoken_symbols.go) - first, before the marks are stripped.
	str = gopodSpeakSymbols(str)
	// GOPOD 2026-10-16 (live: "Check out the GOPOD Weather Kit
	// (/gopod_weather_kit.html)" was said "gopod weather kit dot-h tee-emm-elle"):
	// a link is said as its words only.
	str = gopodLinkWords(str)

	// these two lines create a transformation that decomposes characters, removes non-spacing marks (like diacritics), and then recomposes the characters, effectively removing special characters
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	result, _, _ := transform.String(t, str)

	// Define the regular expression to match special characters
	// GOPOD 2026-09-29: "&" is said "and" ("Gomads & Goverlord" was said "Gomads Goverlord").
	result = gopodAmpersandRe.ReplaceAllString(result, " and ")
	// GOPOD 2026-10-05: an asterisk between two words leaves a space, not a join
	// ("fun!*Giggles merrily*" was said "fun!Giggles", live 2026-10-04 20:41).
	result = gopodAsteriskWordGap(result)
	// GOPOD 2026-10-07: "@" is said "at" ("Don't even @ me!" was said "Don't
	// even me").
	result = gopodAtSignRe.ReplaceAllString(result, "at ")
	// GOPOD 2026-10-08: "+" is said "plus" ("18+" is said "18 plus").
	result = gopodSayPlus(result)
	re := regexp.MustCompile(`[&^*#@]`)

	// Replace special characters with an empty string
	// GOPOD 2026-09-29: an emoji that ends a sentence becomes a period first.
	result = removeEmojis(gopodEmojiSentenceEnd(re.ReplaceAllString(result, "")))
	// GOPOD 2026-10-14: an emoji's joiner and style marks (U+200D, U+FE0F) go
	// with it - "\u200d I'll tell you" was sent once "🙅‍♂️" was taken out, and the
	// robot said "yirl" (operator: "I did hear a \"yirl\" (rhymes with whirl) sound").
	result = gopodEmojiJoinersRe.ReplaceAllString(result, "")

	// Replace special characters with ASCII
	// * COPY/PASTE TO ADD MORE CHARACTERS:
	//   result = strings.ReplaceAll(result, "", "")
	result = strings.ReplaceAll(result, "‘", "'")
	result = strings.ReplaceAll(result, "’", "'")
	result = strings.ReplaceAll(result, "“", "\"")
	result = strings.ReplaceAll(result, "”", "\"")
	// GOPOD 2026-09-29: em dash -> spoken pause (comma), same as the Python say path.
	result = strings.ReplaceAll(result, "—", ", ")
	// GOPOD 2026-10-07: a colon after a word is a pause, like a comma ("Doc's
	// Take: 1k subs"); times (4:20) are untouched. A year written '20 is said
	// "20", the apostrophe dropped.
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
	// GOPOD 2026-09-29: AI is spoken "AY-EYE" (the pronunciation registry's own
	// entry), not "A. I.", and the em-dash comma never leaves a stray " ," or ".,".
	// GOPOD 2026-09-29: a dash between words ("the hype – the verbal boom", "are- but",
	// "laugh--that") is a spoken pause, not "dash" (the robot said "dash"). A hyphen
	// inside a word (Gowp-awd, free-spirited) and a dash between two numbers stay.
	result = gopodSpokenDashes(result)
	// GOPOD 2026-09-29: "Brobot 1's" was said "Brobot one ess" - the number is
	// written as a word before 's ("Brobot one's").
	result = gopodBrobotPossessiveRe.ReplaceAllStringFunc(result, func(m string) string {
		if strings.Contains(m, "1") {
			return "Brobot one's"
		}
		return "Brobot two's"
	})
	result = gopodAISpokenRe.ReplaceAllString(result, "AY-EYE")
	result = gopodSpaceBeforeCommaRe.ReplaceAllString(result, ",")
	result = gopodCommaAfterPunctRe.ReplaceAllString(result, "$1")
	// GOPOD 2026-09-29: no space left inside brackets when an emoji is removed
	// ("too dramatic. 😬)" was said "dramatic. )").
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
	// GOPOD 2026-10-04: + the newer emoji block 1FA00-1FAFF (🪅 was said, live 14:57).
	re := regexp.MustCompile(`[\x{1FA00}-\x{1FAFF}]|[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E0}-\x{1F1FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]|[\x{1F900}-\x{1F9FF}]|[\x{1F004}]|[\x{1F0CF}]|[\x{1F18E}]|[\x{1F191}-\x{1F251}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]`)
	result := re.ReplaceAllString(input, "")
	return result
}

func CreateAIReq(transcribedText, esn string, gpt3tryagain, isKG bool) openai.ChatCompletionRequest {
	var nChat []openai.ChatCompletionMessage
	systemPrompt, promptBytes, usedOpenAIPrompt := loadCanonicalBrobotPrompt(transcribedText)
	esnLine := "Current robot ESN: " + esn
	// GOPOD 2026-09-29: say which brother this is, with his own persona line
	// (gopod_brobot_personas.go), so the reply is "I'm Brobot 1".
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
	// GOPOD 2026-10-05: general to specific - the shared persona and GOPOD's
	// facts above, then who this brother is, then his reply shape and the final
	// check last, nearest the question (BROBOTS_LAYERED_PROMPT_002.md).
	if !usedOpenAIPrompt {
		// 2026-10-05: the brother's persona line only - no ESN. The robot ID in
		// the prompt was said out loud ("I'm Brobot ESN <serial>"); the code tells
		// the brothers apart by ESN, the model only needs his name
		// (SOCIAL_DEMO_REPLAY_003.md).
		whoIAm := esnLine
		if line, ok := gopodBrobotPersonaLines[gopodBrobotName(esn)]; ok {
			whoIAm = line
		}
		smsg.Content += "\n\nWHO I AM:\n" + whoIAm
		if facts := strings.TrimSpace(gopodQuestionFactsFor(esn, transcribedText)); facts != "" {
			smsg.Content += "\n\n" + gopodBrobotsSettings.QuestionFactsHeader + "\n" + facts
		}
		// GOPOD 2026-10-22: his pre-build context, after the facts (gopodPrebuildContext).
		if pre := gopodPrebuildContext(esn, transcribedText); pre != "" {
			smsg.Content += "\n\n" + pre
		}
		// GOPOD 2026-10-22: the answer chain's shape for the model (gopodChainPrompt).
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
	// GOPOD 2026-09-29: both Brobots are one persona with one shared chat. Recent
	// exchanges from BOTH robots come back as that chat, each question tagged with
	// the brother it went to, and this question is tagged for this robot - so each
	// brother continues the conversation in his own voice instead of copying.
	// Replaces the per-robot remembered chat (GetChat) for the Brobots.
	if vars.APIConfig.Knowledge.SaveChat {
		shared := gopodAddExperienceTakes(esn, gopodSharedBrobotChat())
		logger.Println("Using shared Brobots chat, length of " + fmt.Sprint(len(shared)) + " messages")
		nChat = append(nChat, shared...)
	}
	// GOPOD 2026-10-02: chat awareness - whether the brother is asleep, right
	// before the question (gopod_chat_awareness.go).
	if note := gopodAwarenessNote(esn); note != "" {
		nChat = append(nChat, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: note})
	}
	// GOPOD 2026-10-04: "thanks Brobots" - a party-bro sign-off (gopod_chat_hold.go).
	if gopodBrobotName(esn) != "Brobot" && gopodIsThanks(transcribedText) {
		nChat = append(nChat, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: gopodBrobotsSettings.ChatHold.ThanksNote})
	}
	nChat = append(nChat, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: gopodTagForBrobot(esn, transcribedText),
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
	// GOPOD 2026-10-07: the Brobots' creativity (llm.temperature, Wire-Pod's 1
	// as-is), lower for a question whose point has a scripted line
	// (llm.say_line_temperature) - the code says the point flat, and the LLM's
	// added talk is where facts were invented.
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
	// GOPOD 2026-10-24 (link 5): a new line's answer waits for the line before's follow-up.
	if _, following := gopodFollowingUp.Load(esn); !following {
		gopodExchangeWait(esn, transcribedText)
	}
	// GOPOD 2026-10-14: the photo egg - "ask", "recall" or "" (gopod_photo.go).
	photoKind := ""
	// GOPOD 2026-10-14: "Explain the Math" - this Brobot's share of the steps,
	// and whether he answers first (gopod_math.go).
	var mathSteps []string
	mathOn, mathLead, mathFollowStarted := false, false, false
	// 2026-10-24 (operator: "Revealed ... after all responses are properly
	// completed"): an answer his brother follows up (All) - the prizes wait for it.
	_, allLead := gopodAllFirstAnswer.Load(esn)
	photoTaken := make(chan struct{})
	var photoAnswer string
	var photoRound2 func()
	// GOPOD: intent matching has already seen the raw transcript; from here on the
	// LLM and the shared chat get GOPOD's names spelled right ("go pod" -> GOPOD).
	transcribedText = gopodSpellNames(transcribedText)
	// GOPOD 2026-10-06: nothing reaches the Brobots before the green light
	// (gopod_llm_warmup.go); a voice request gets the quiet native ending.
	// GOPOD 2026-10-13 (operator: "Robot microphone * log & hold in queue same
	// as the others"): heard during a warm-up, the line waits for LLM Ready
	// like a typed one; the robot gets the quiet ending now.
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
	// GOPOD 2026-10-09: a spoken thanks while a draft or song is going closes
	// them out first, then comes back as the operator's line
	// (gopod_thanks_close_out.go).
	if req != nil && gopodBrobotName(esn) != "Brobot" && gopodThanksCloseOut(esn, transcribedText, "voice") {
		IntentPass(req, "intent_system_noaudio", transcribedText, map[string]string{}, false)
		return "", nil
	}
	// GOPOD 2026-10-10, Rule D (gopod_session.go): a spoken question during a
	// run gets no LLM answer - the quiet native ending, and Goverlord's line.
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
	}
	_, err := robot.Conn.BatteryState(context.Background(), &vectorpb.BatteryStateRequest{})
	if err != nil {
		return "", err
	}
	// GOPOD 2026-10-04: only the Brobot called by name answers; with no name, the
	// one heard first does. The other stays quiet (gopod_addressing.go).
	decidedAt := time.Now()
	if gopodBrobotName(esn) != "Brobot" {
		// GOPOD 2026-10-04: the answer to Record Session's keep-or-delete question
		// is taken here, by whichever Brobot hears it first; no chat reply
		// (gopod_record_session.go).
		if GopodRecordSessionAnswer(transcribedText, gopodVoiceCommandHow) {
			if !isKG && req != nil {
				gopodLLMClosePass.Store(esn, true) // GOPOD 2026-10-21: closes the voice request, not a native intent
				IntentPass(req, "intent_greeting_hello", "", map[string]string{}, false)
			}
			return "", nil
		}
		// GOPOD 2026-10-04: the stay-in-place hold lets go of this Brobot while he
		// answers or defers, and holds both again after (gopod_chat_hold.go).
		gopodChatHoldBeforeAnswer(esn)
		answer, to := gopodShouldAnswer(esn, transcribedText, decidedAt)
		gopodAddressingLog(esn, to, answer)
		if !answer {
			gopodQuietBrother(req, robot, esn, transcribedText, isKG)
			gopodChatHoldResume(esn)
			return "", nil
		}
		transcribedText = gopodCanonicalCall(transcribedText)
		// GOPOD 2026-10-14: "Explain the Math" - Brobot 1 answers first, then
		// Brobot 2 follows up (gopod_math.go). A line Brobot 2 got first goes to
		// Brobot 1; Brobot 2's turn comes after him.
		if first, second, ok := gopodMathParts(transcribedText); ok {
			_, follow := gopodMathFollowUp.LoadAndDelete(esn)
			// 2026-10-24: the lead is Brobot 1, or Brobot 2 when All was sent with him first.
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
		// GOPOD 2026-10-14: the photo egg (gopod_photo.go).
		photoKind = gopodPhotoKind(transcribedText)
	}
	// GOPOD 2026-09-29: the Brobots play Wire-Pod's own knowledge-graph wait on the
	// "Hey Vector" chat path too - behavior control at once, the searching clip on a
	// loop until the reply is ready, then the searching get-out - instead of standing
	// idle while the model thinks. Same gesture as go search (searching, then
	// searchingGetout). isKG itself is unchanged (prompt, intent pass).
	searchWhileWaiting := isKG || gopodBrobotName(esn) != "Brobot"
	// The searching gesture already ended with its get-out (no second one at the answer).
	var searchingClosed atomic.Bool
	if searchWhileWaiting {
		// GOPOD 2026-09-29: a Brobot first says "Thinking, please wait..." (like go
		// search's "Let me look that up."), as soon as behavior control is really
		// granted - BControl's own grant signal, not a timer - then loops searching.
		isBrobot := gopodBrobotName(esn) != "Brobot"
		controlGranted := start
		if isBrobot {
			controlGranted = make(chan bool)
		}
		gopodLogBControl(robot, ctx, controlGranted, stop)
		go func() {
			if isBrobot {
				<-controlGranted
				// GOPOD 2026-10-16 (operator: "thinking please wait > photo > TPU >
				// Brobot N: I see [TPU labels]"): a photo is taken inside his
				// thinking - after "Thinking, please wait.", before the searching
				// moves: the photo the golden way, its egg, the Edge TPU's quick look.
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
				// GOPOD 2026-10-13 (operator: "snapshot the full \"thinking please
				// wait\" process from live (+ arm, nod) * lock that snapshot * use
				// locked \"thinking please wait\" process for scripted runs"): one
				// process, live and scripted alike (gopod_thinking_line.go).
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
	// GOPOD 2026-10-09: a Brobot's reply bubble, published once when he starts
	// speaking (the operator's ask: after the thinking, not before it).
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
	weatherShown := false // GOPOD 2026-10-22: his weather line's picture, once
	var interrupted bool
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
	// GOPOD 2026-10-22: the value points notice in his bubble (gopodValuePointsNoticePost).
	gopodValuePointsNoticePost(esn)
	// GOPOD 2026-10-15: a pop-up's picture as the vision model saw it
	// (gopod_picture_describe.go), before the question; a question about
	// chocolate bingo has its pictures looked at now (describe_on_ask).
	if gopodBrobotName(esn) != "Brobot" && !gopodIsFollowingUp(esn) {
		gopodPictureDescribeOnAsk(transcribedText)
	}
	if note := gopodPictureNote(); note != "" && !gopodScriptPlaying.Load() && len(aireq.Messages) > 0 {
		last := &aireq.Messages[len(aireq.Messages)-1]
		last.Content = note + "\n" + last.Content
	}
	// GOPOD 2026-10-15: a follow-up (All, Explain the Math) - his brother has just
	// answered; the note goes before the question (gopodFollowingUp).
	if gopodIsFollowingUp(esn) && len(aireq.Messages) > 0 {
		note := strings.TrimSpace(gopodBrobotsSettings.Addressing.FollowUpNote)
		// 2026-10-15 (operator: "can the Brobot 2 follow-up response, also
		// follow-up on errors like the 'regular fraction'?"): Explain the Math's.
		if m := strings.TrimSpace(gopodBrobotsSettings.Math.FollowUpNote); mathOn && m != "" {
			note = m
		}
		if note != "" {
			last := &aireq.Messages[len(aireq.Messages)-1]
			last.Content = note + "\n" + last.Content
		}
	}

	// GOPOD 2026-10-07: while a draft performance plays, the Brobot says the
	// draft's answer word for word instead of asking the LLM
	// (gopod_interview_script.go); everything else runs as live.
	var stream gopodChatStream
	scriptedAnswer, scripted := gopodScriptAnswer(esn, transcribedText)
	// GOPOD 2026-10-14: a photo question is answered with what the Brobot sees
	// (or saw), word for word, not by the chat LLM (gopod_photo.go).
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
				// GOPOD 2026-10-07: a draft's answer is already shaped - it's said as written.
				if !scripted {
					// GOPOD 2026-09-29: names kept right (gopod_name_corrections.go).
					// 2026-09-30: and animation words opening a line are dropped
					// (gopod_leading_animation_words.go) - shown and spoken otherwise.
					rawRespText = gopodFixSelfName(esn, gopodCorrectNames(gopodStripLeadingAnimationWords(rawRespText)))
					// GOPOD 2026-10-06: no made-up web addresses (gopodDropWebLinks).
					rawRespText = gopodDropWebLinks(rawRespText)
					// GOPOD 2026-10-06: "Hey, that reminds me..." is the code's line only
					// (gopodDropLLMReminder, gopod_value_points.go).
					rawRespText = gopodDropLLMReminder(rawRespText)
					// GOPOD 2026-10-16: never the question read back (gopodDropQuestionEcho).
					rawRespText = gopodDropQuestionEcho(rawRespText, transcribedText)
					// GOPOD 2026-10-16: clean GOPOD - right flags, no politics (gopod_clean_politics.go).
					rawRespText = gopodCleanPolitics(rawRespText, transcribedText)
					// GOPOD 2026-10-16: a follow-up never just repeats his brother
					// (gopod_brother_repeat.go); each keeps his own words for the next.
					if gopodIsFollowingUp(esn) {
						gopodBrotherWords.Lock()
						brother := gopodBrotherWords.last[gopodBrotherOf(esn)]
						gopodBrotherWords.Unlock()
						rawRespText = gopodDropBrotherRepeat(rawRespText, brother)
					}
					// GOPOD 2026-10-06: tighter - no lists, no stage directions, and
					// 2026-10-07 the question's response budget (gopod_reply_tighten.go).
					rawRespText = gopodTightenReply(rawRespText, transcribedText, esn)
					// GOPOD 2026-10-15: each of gemma's emoji-closed thoughts on its own line.
					rawRespText = gopodThoughtBreaks(rawRespText)
					gopodRememberOwnWords(esn, rawRespText)
					// GOPOD 2026-10-14: "Explain the Math" - his own line or two,
					// then his steps, worked out in code (gopodMathReply).
					if mathOn {
						extra := ""
						if !mathLead {
							extra = gopodMathDecimalLine(transcribedText)
						}
						rawRespText = gopodMathReply(rawRespText, mathSteps, extra)
						logger.Println("GOPOD_MATH " + gopodBrobotName(esn) + ": " + rawRespText)
					} else {
						// GOPOD 2026-10-21: the thought bubbles template - the answer chain
						// (answer_chains, gopod_answer_chain.go): the start snippet, the
						// answer, the weather kit line, the end snippet, the hand-over.
						rawRespText = gopodShapeBrobotAnswer(esn, transcribedText, rawRespText)
					}
				}
				// 2026-09-30: the animation cue is marked "happy___" (gopod_animation_marker.go);
				// the bubble shows the reply as written, cue removed, and the rest of
				// Wire-Pod gets its internal "happy... " form.
				displayRespText := gopodDisplayReply(rawRespText)
				var displayPieces []int // GOLDEN highlight: spoken piece -> bubble piece (gopod_highlight_pieces.go)
				// GOPOD 2026-10-16 (live 13:2x: "Check out the [GOPOD Weather
				// Kit](...)." was said "Check out the."): the bubble keeps the link,
				// the robot says its words.
				rawRespText = gopodAnimationMarkerToInternal(gopodLinkWords(rawRespText))
				streamResult = buildGOPODStreamResultWithRepair(ctx, rawRespText, aireq.Model, openAIGOPODRobotSpeechRepairer(c))
				fullRespSlice = append([]string(nil), streamResult.AudibleChunks...)
				// GOPOD 2026-10-24: the bubble shows what's said - the fallback line
				// when the reply couldn't be said (gopodStreamResult.Fallback).
				if streamResult.Fallback {
					// 2026-10-24: why, so the next one can be fixed at its cause - the
					// line that failed, the rule it broke, and the reply as written.
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
				// GOPOD 2026-10-04: a typed question (gopod_typed_input.go) has no
				// voice request to close.
				if !isKG && req != nil {
					gopodLLMClosePass.Store(esn, true) // GOPOD 2026-10-21: closes the voice request, not a native intent
					IntentPass(req, "intent_greeting_hello", transcribedText, map[string]string{}, false)
				}
				// GOPOD 2026-10-06: the interview's chapter title before the reply.
				// 2026-10-08 (operator: "the granny pic should come after [the
				// question] bubble, before Brobot 1 response"): the topic's pictures
				// too - for a typed line they're already shown with it
				// (gopod_typed_input.go); each shows once.
				// 2026-10-14 (live: each egg popped up twice - once with the typed
				// line, again here): only a line the robot heard himself shows
				// them here.
				if gopodBrobotName(esn) != "Brobot" {
					gopodShowChapter(transcribedText)
					if req != nil {
						gopodShowTopicPictures(esn, transcribedText)
					}
				}
				// GOPOD 2026-10-09 (operator: "have bubble wait until brobot
				// finishes the scripted thinking please wait searching ... so brobot
				// bubble appears when brobot speaks, after thinking"): a Brobot's
				// bubble is published when he starts speaking (showReply below).
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
			// GOPOD 2026-09-29: an accepted interrupt also sets `interrupted`, so the
			// chunk loop below stops and drops the rest of the reply. Before this it
			// was never set (WIRED-POD.md, "the interrupted bool itself is still never
			// set"): behavior control was released, but the loop went on, and the
			// unspoken chunks played at the robot's next request.
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
		// GOPOD 2026-09-29: one arm cue per reply, with its first mid-reply animation
		// (gopod_arm_cue.go); waited for before behavior control is released.
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
			// GOPOD 2026-09-29: speak with the pronunciation registry (gopod_pronunciation.go).
			commandPacket := gopodPronounceCommandPacket(commandPacketFromFilteredChunk(currentChunk))
			// GOPOD 2026-09-29: vary the animation a little when the model gave none
			// (gopod_animation_variety.go).
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
			// GOPOD 2026-09-29: speak sentence by sentence and check for an interrupt
			// before each (gopod_speech_sentences.go), so a touch stops the robot at
			// the end of the current sentence instead of after the whole chunk. Each
			// call gets its own stop channel, closed right after, instead of stopStop:
			// every PerformActions call starts a listener on the channel it is given,
			// and on stopStop those listeners never ended and could take the
			// end-of-reply signal meant for the interrupt listener.
			prevWasSpeech := false
			prevWasAnimation := false
			showReply()
			// GOPOD 2026-10-18 (step 3, the engine conducts): he speaks once his
			// bubble is on the page (gopod_page_sync.go).
			if gopodBrobotName(esn) != "Brobot" {
				gopodWaitReplyShown(esn)
			}
			for _, act := range gopodAnimationsFromWords(gopodSprinkleAnimations(gopodSplitSpeechIntoSentences(acts))) {
				streamMu.Lock()
				stopNow := interrupted
				streamMu.Unlock()
				if stopNow {
					break
				}
				if act.Action == ActionPlayAnimationWI {
					// GOPOD 2026-10-09: no new animation until the one before it
					// is done (gopod_robot_cmd_log.go).
					gopodAnimationWaitDone(robot.Cfg.SerialNo)
					logger.LogDebugUI("GOPOD_SAY_ANIMATION: " + act.Parameter)
					if armCueDone == nil && prevWasSpeech {
						armCueDone = gopodStartReplyMove(robot, robot.Cfg.SerialNo)
					}
				}
				// GOPOD 2026-10-09 (operator: "Use playAnimationWI + 0.9s + start
				// speech"): the speech starts speech_after_animation_seconds after
				// the animation was sent.
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
					// 2026-10-08: a draft answer's moves at their words (gopod_script_moves.go).
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
				// GOPOD 2026-10-06: a Brobot that drops off mid-answer no longer
				// leaves this answer waiting forever (gopod_stuck_reply.go).
				disconnect = gopodPerformOrGiveUp(nChat, act, robot, actionStop)
				close(actionStop)
				if act.Action == ActionSayText {
					logger.LogDebugUI(fmt.Sprintf("GOPOD_VECTOR_PHRASE: %s done took=%dms", gopodBrobotName(robot.Cfg.SerialNo), time.Since(actStart).Milliseconds()))
					// GOPOD 2026-10-22: his weather line said - its picture and sound,
					// then the rest (gopodWeatherLineSaid).
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
		// GOPOD 2026-09-29: the reply is over. Finished: the chat page drops its
		// sentence outline. Touch-interrupted: the page shows the sentence the robot
		// stopped on with a solid outline (gopod_speaking_now.go).
		streamMu.Lock()
		stoppedByInterrupt := interrupted
		streamMu.Unlock()
		if stoppedByInterrupt {
			gopodMarkSpeakingInterrupted(robot.Cfg.SerialNo)
		} else {
			gopodWriteSpeakingNow(robot.Cfg.SerialNo, "", false, false)
		}
		// GOPOD 2026-10-23 (A4): the grand prize, once its last point has been said -
		// 2026-10-24: and once every answer to the line is done (a follow-up's own
		// end checks again).
		if !mathLead && !allLead {
			gopodGameWinCheck()
		}
		// A reply that never got to speak still shows its bubble.
		showReply()
		// GOPOD 2026-10-09: the battery after each Brobot reply (gopod_robot_cmd_log.go).
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
		// GOPOD 2026-10-04: the answer's own control is released - the chat's
		// stay-in-place hold takes both Brobots (gopod_chat_hold.go). Here, not
		// deferred: this function doesn't return after an answer (live 13:58).
		// A "thanks Brobots" sign-off ends the hold instead.
		if gopodBrobotName(esn) != "Brobot" {
			// GOPOD 2026-10-16: a photo's round 2 (gopod_photo.go) is part of
			// his answer - "Gathering details. Please wait!" and a stall line
			// while the vision model still looks, then statement 2.
			if photoRound2 != nil && !wasInterrupted {
				photoRound2()
			}
			// GOPOD 2026-10-18: CHALK's QR part after his answer (gopod_managers.go).
			if !wasInterrupted {
				gopodManagerQRShow()
			}
			gopodReplyEnd(esn)
			// GOPOD 2026-10-24 (link 5): his answer is said - a line held for him goes on.
			gopodExchangeDone(esn)
			gopodFollowingUp.Delete(esn)
			// GOPOD 2026-10-14: "Explain the Math" - Brobot 2 follows up.
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

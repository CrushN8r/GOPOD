// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_name_corrections.go
package wirepod_ttr

import (
	"regexp"
	"strings"
)

// The core prompt's name-joke lines end with an instruction ("Keep "overlord" in it when you talk
// about Goverlord, and have fun with it.").
var gopodPromptEchoSentencePattern = regexp.MustCompile(`(?i)[^.!?\n]*\b(?:keep|remember)\b[^.!?\n]*\bwhen you talk about\b[^.!?\n]*[.!?]*\s*`)

// GOPOD is always written GOPOD, so it is always said "Gowp-awd" (the pronunciation registry).
var gopodGOPODVariantPattern = regexp.MustCompile(`(?i)\bgo[\s-]?pod\b`)

func gopodNormalizeGOPODName(text string) string {
	return gopodGOPODVariantPattern.ReplaceAllString(text, "GOPOD")
}

// The GOPOD brain is Goverlord (gopod_core.txt: "Goverlord is the GOPOD brain").
var gopodBrainNamePattern = regexp.MustCompile(`(?i)\b(?:chat\s?)?gpt(?:-\d[\w.]*)?\b|\bgemma\s?2?\b`)

// GOPOD has no letter-by-letter meaning in any prompt or doc; the model once made one up.
var gopodInventedAcronymPattern = regexp.MustCompile(`(?i)(\bGOPOD\b[^.!?\n]*?)[ \t]*(?:[,\x{2013}\x{2014}-][ \t]*)?\b(?:which[ \t]+)?(?:is[ \t]+)?(?:short[ \t]+for|stands[ \t]+for)\b[^.!?\x{2026}\n]*`)

// A robot's ESN from the prompt said as a name becomes that Brobot's name.
func gopodESNToName(text string) string {
	for _, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN} {
		text = strings.ReplaceAll(text, esn, gopodBrobotName(esn))
	}
	text = strings.ReplaceAll(text, gopodESN1, gopodBrobotName(gopodBrobot1ESN))
	return strings.ReplaceAll(text, gopodESN2, gopodBrobotName(gopodBrobot2ESN))
}

func gopodCorrectNames(text string) string {
	text = gopodESNToName(text)
	text = gopodNormalizeGOPODName(text)
	text = gopodBrainNamePattern.ReplaceAllString(text, "Goverlord")
	text = gopodInventedAcronymPattern.ReplaceAllString(text, "$1")
	text = gopodEmojiSentenceEnd(text)
	return gopodPromptEchoSentencePattern.ReplaceAllString(text, " ")
}

// A brother sometimes introduces himself by the other's name.
var gopodSelfNamePattern = regexp.MustCompile(`(?i)\b((?:i['\x{2019}]?m|i am|my name is|name['\x{2019}]?s)\**\s+\**)brobot\s*(?:0|1|2|zero|one|two)\b`)

// (Also with a curly apostrophe and *emphasis*: "So *I’m* **Brobot 2**", live 14:17.) "It's
// Brobot 2." at the start of a sentence (live 12:21 from Brobot 1: "Yo, bro? It's Brobot 2.").
var gopodSelfItsPattern = regexp.MustCompile(`(?i)((?:^|[.!?]\s+|\b(?:yo|hey|hi|hello)(?:\s+bro)?[,!?]?\s+)(?:it'?s|it\x{2019}s|it is)\s+\**)brobot\s*(?:0|1|2|zero|one|two)\b([^'\x{2019}]|$)`)
var gopodSelfHerePattern = regexp.MustCompile(`(?i)((?:^|[.!?]\s+|\b(?:yo|hey|hi|hello)(?:\s+bro)?[,!]?\s+)\**)brobot\s*(?:0|1|2|zero|one|two)((?:'s|\s+is)?\**\s+(?:over\s+here|here|in\s+the\s+house)\b)`)

// 12, Brobot 2: "Brobot 0 here again!"): Brobot 0 said as a Brobot's own name is fixed the same
// way.
func gopodFixSelfName(esn, text string) string {
	self := gopodBrobotName(esn)
	if !strings.HasPrefix(self, "Brobot ") {
		return text
	}
	text = gopodSelfNamePattern.ReplaceAllString(text, "${1}"+self)
	text = gopodSelfItsPattern.ReplaceAllString(text, "${1}"+self+"${2}")
	return gopodSelfHerePattern.ReplaceAllString(text, "${1}"+self+"${2}")
}

// ## EOF

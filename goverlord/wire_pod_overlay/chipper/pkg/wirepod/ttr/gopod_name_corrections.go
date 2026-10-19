// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_name_corrections.go
package wirepod_ttr

import (
	"regexp"
	"strings"
)

// The core prompt's name-joke lines end with an instruction ("Keep "overlord" in
// it when you talk about Goverlord, and have fun with it."). The model sometimes
// reads that instruction out; any sentence that repeats it is dropped.
var gopodPromptEchoSentencePattern = regexp.MustCompile(`(?i)[^.!?\n]*\b(?:keep|remember)\b[^.!?\n]*\bwhen you talk about\b[^.!?\n]*[.!?]*\s*`)

// GOPOD is always written GOPOD, so it is always said "Gowp-awd" (the pronunciation
// registry). Speech-to-text hears it as "go pod", and the model copies that as
// "GO-pod" / "Go Pod" (live 2026-09-29) - every such spelling becomes GOPOD, in
// replies and in the question sent to the model. Operator, 2026-09-29: "Gowp-awd
// it is!! And lock it in."
var gopodGOPODVariantPattern = regexp.MustCompile(`(?i)\bgo[\s-]?pod\b`)

func gopodNormalizeGOPODName(text string) string {
	return gopodGOPODVariantPattern.ReplaceAllString(text, "GOPOD")
}

// The GOPOD brain is Goverlord (gopod_core.txt: "Goverlord is the GOPOD brain").
// The model calls its own brain "GPT" ("this brain of GPT", "brought together by
// the awesome GPT", live 2026-09-29) - operator: "Goverlord = GOPOD brain". GPT /
// ChatGPT / GPT-4 and the model's own name (gemma, gemma2) become Goverlord.
var gopodBrainNamePattern = regexp.MustCompile(`(?i)\b(?:chat\s?)?gpt(?:-\d[\w.]*)?\b|\bgemma\s?2?\b`)

// GOPOD has no letter-by-letter meaning in any prompt or doc; the model once made
// one up ("GOPOD bot - short for Government-Overlord-Party Dope", live 2026-09-29).
// A "short for ..." / "stands for ..." clause after GOPOD is cut off, up to the end
// of its sentence, keeping the rest of the sentence.
var gopodInventedAcronymPattern = regexp.MustCompile(`(?i)(\bGOPOD\b[^.!?\n]*?)[ \t]*(?:[,\x{2013}\x{2014}-][ \t]*)?\b(?:which[ \t]+)?(?:is[ \t]+)?(?:short[ \t]+for|stands[ \t]+for)\b[^.!?\x{2026}\n]*`)

// A robot's ESN from the prompt said as a name ("Okay, so Party Bro 0dd1d8bf",
// live 2026-10-04 14:54) becomes that Brobot's name.
func gopodESNToName(text string) string {
	for _, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN} {
		text = strings.ReplaceAll(text, esn, gopodBrobotName(esn))
	}
	return text
}

func gopodCorrectNames(text string) string {
	text = gopodESNToName(text)
	text = gopodNormalizeGOPODName(text)
	text = gopodBrainNamePattern.ReplaceAllString(text, "Goverlord")
	text = gopodInventedAcronymPattern.ReplaceAllString(text, "$1")
	text = gopodEmojiSentenceEnd(text)
	return gopodPromptEchoSentencePattern.ReplaceAllString(text, " ")
}

// A brother sometimes introduces himself by the other's name ("Yo, Brobot 2 over
// here!" from Brobot 1, live 2026-09-29; 1 in 18 replies in a gemma2:2b test).
// Only clear self-introductions are fixed to this robot's own name: "I'm / I am /
// my name is / name's Brobot N", "It's Brobot N" at a sentence start, and at the start of a sentence "Brobot N here",
// "Brobot N over here" or "Brobot N('s) in the house" (bold ** allowed - live
// 11:45: "**Brobot 2's in the house!**" from Brobot 1). Talk about the brother
// ("my bro Brobot 2", "and Brobot 2's in the house too") is left alone.
var gopodSelfNamePattern = regexp.MustCompile(`(?i)\b((?:i['\x{2019}]?m|i am|my name is|name['\x{2019}]?s)\**\s+\**)brobot\s*(?:0|1|2|zero|one|two)\b`)

// (Also with a curly apostrophe and *emphasis*: "So *I’m* **Brobot 2**", live 14:17.)
// "It's Brobot 2." at the start of a sentence (live 12:21 from Brobot 1: "Yo, bro?
// It's Brobot 2."). Not "It's Brobot 2's turn" - a following 's is left alone.
var gopodSelfItsPattern = regexp.MustCompile(`(?i)((?:^|[.!?]\s+|\b(?:yo|hey|hi|hello)(?:\s+bro)?[,!?]?\s+)(?:it'?s|it\x{2019}s|it is)\s+\**)brobot\s*(?:0|1|2|zero|one|two)\b([^'\x{2019}]|$)`)
var gopodSelfHerePattern = regexp.MustCompile(`(?i)((?:^|[.!?]\s+|\b(?:yo|hey|hi|hello)(?:\s+bro)?[,!]?\s+)\**)brobot\s*(?:0|1|2|zero|one|two)((?:'s|\s+is)?\**\s+(?:over\s+here|here|in\s+the\s+house)\b)`)

// 2026-10-06 (live 16:12, Brobot 2: "Brobot 0 here again!"): Brobot 0 (the
// human operator) said as a Brobot's own name is fixed the same way.
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

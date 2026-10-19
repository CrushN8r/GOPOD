// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_question_echo.go
package wirepod_ttr

import (
	"regexp"
	"strings"
)

// gopodEchoSentenceRe: one sentence of a reply, with its end mark.
var gopodEchoSentenceRe = regexp.MustCompile(`[^.!?\n]+[.!?]*`)

// gopodDropQuestionEcho (2026-10-16, live: Brobot 2's weather reply read the
// operator's line back - "I want to visit my granny today. She is gangsta!
// What's the weather in Jerusalem, Israel?"): a sentence of three words or more
// whose words are all the question's is the question read back, and goes.
func gopodDropQuestionEcho(reply, question string) string {
	asked := map[string]bool{}
	for _, w := range gopodHighlightWords(question) {
		asked[w.word] = true
	}
	if len(asked) < 4 {
		return reply
	}
	out := gopodEchoSentenceRe.ReplaceAllStringFunc(reply, func(sentence string) string {
		words := gopodHighlightWords(sentence)
		if len(words) < 3 {
			return sentence
		}
		for _, w := range words {
			if !asked[w.word] {
				return sentence
			}
		}
		return ""
	})
	return strings.TrimSpace(regexp.MustCompile(`[ \t]{2,}`).ReplaceAllString(out, " "))
}

// ## EOF

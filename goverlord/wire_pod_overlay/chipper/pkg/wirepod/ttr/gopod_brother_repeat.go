// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_brother_repeat.go
package wirepod_ttr

import (
	"regexp"
	"strings"
	"sync"
)

// The All follow-up, no repeats (2026-10-16, operator: "when 'All' is selected
// ... brobot 1 answers, brobot 2 follow up * handle similar thought bubbles from
// this 2-brobot response sequence, without blatant repeating 1st response
// material into the 2nd response * a man vs a man with a hat is fine * Each
// Brobot keeps his own thoughts as paragraphs"; "do it in code"). Each Brobot's
// own words (the LLM's part, before the code's lines) are kept; a follow-up drops
// a sentence of his that adds nothing new - every word that carries meaning in it
// was already in his brother's reply. One new word ("a man with a hat") keeps it.
// The code's own lines (the weather line, "Hey, that reminds me...") are added
// after, so they stay.

var gopodBrotherWords = struct {
	sync.Mutex
	last map[string]string
}{last: map[string]string{}}

// gopodRememberOwnWords keeps this Brobot's own words of his newest reply.
func gopodRememberOwnWords(esn, reply string) {
	gopodBrotherWords.Lock()
	gopodBrotherWords.last[esn] = reply
	gopodBrotherWords.Unlock()
}

// gopodBrotherSaid: a follow-up's brother ended his newest reply with this line.
func gopodBrotherSaid(esn, line string) bool {
	if !gopodIsFollowingUp(esn) {
		return false
	}
	gopodBrotherWords.Lock()
	last := gopodBrotherWords.last[gopodBrotherOf(esn)]
	gopodBrotherWords.Unlock()
	norm := func(s string) string { return strings.ToLower(strings.Join(strings.Fields(s), " ")) }
	return line != "" && strings.Contains(norm(last), norm(line))
}

func gopodBrotherOf(esn string) string {
	switch esn {
	case gopodBrobot1ESN:
		return gopodBrobot2ESN
	case gopodBrobot2ESN:
		return gopodBrobot1ESN
	}
	return ""
}

// gopodRepeatPlainWord: a word that carries no meaning of its own
// (brother_repeat.plain_words).
func gopodRepeatPlainWord(w string) bool {
	for _, p := range gopodBrobotsSettings.BrotherRepeat.PlainWords {
		if strings.EqualFold(p, w) {
			return true
		}
	}
	return false
}

// gopodDropBrotherRepeat: the follow-up's reply with every sentence that only
// repeats his brother's reply taken out (paragraphs kept).
func gopodDropBrotherRepeat(reply, brother string) string {
	said := map[string]bool{}
	for _, w := range gopodHighlightWords(brother) {
		said[w.word] = true
		said[strings.TrimSuffix(w.word, "s")] = true
	}
	if len(said) == 0 {
		return reply
	}
	out := gopodEchoSentenceRe.ReplaceAllStringFunc(reply, func(sentence string) string {
		meaning := 0
		for _, w := range gopodHighlightWords(sentence) {
			if gopodRepeatPlainWord(w.word) {
				continue
			}
			meaning++
			if !said[w.word] && !said[strings.TrimSuffix(w.word, "s")] {
				return sentence // something new
			}
		}
		if meaning < 2 {
			return sentence // too short to call a repeat ("Bro!")
		}
		return ""
	})
	out = regexp.MustCompile(`[ \t]{2,}`).ReplaceAllString(out, " ")
	out = regexp.MustCompile(`[ \t]*\n[ \t]*`).ReplaceAllString(out, "\n")
	return strings.TrimSpace(regexp.MustCompile(`\n{2,}`).ReplaceAllString(out, "\n"))
}

// ## EOF

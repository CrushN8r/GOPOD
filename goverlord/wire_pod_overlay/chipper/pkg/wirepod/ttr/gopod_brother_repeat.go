// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_brother_repeat.go
package wirepod_ttr

import (
	"regexp"
	"strings"
	"sync"
)

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

// gopodRepeatPlainWord: a word that carries no meaning of its own (brother_repeat.plain_words).
func gopodRepeatPlainWord(w string) bool {
	for _, p := range gopodBrobotsSettings.BrotherRepeat.PlainWords {
		if strings.EqualFold(p, w) {
			return true
		}
	}
	return false
}

// gopodDropBrotherRepeat: the follow-up's reply with every sentence that only repeats his brother's
// reply taken out (paragraphs kept).
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

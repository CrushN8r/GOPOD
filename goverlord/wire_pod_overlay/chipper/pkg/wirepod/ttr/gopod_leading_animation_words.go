// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_leading_animation_words.go
package wirepod_ttr

import (
	"regexp"
	"strings"
	"unicode"
)

// The model sometimes opens a reply (or a paragraph) with animation words and
// a wide gap, instead of the "happy... text" form: "celebrate  Oh man!",
// "curious dartingEyes Like...", "excited love   okay", "thinking  well",
// "veryHappy   Man oh man!" (live 2026-09-30 20:16-20:29). They showed in the
// chat bubble, and "excited love" was spoken. At the start of a line, a run of
// words that are all animation tokens or aliases (animation_vocab.json) is
// dropped when it is two or three words ("curious dartingEyes Like...") or one
// word followed by two or more spaces ("thinking  well"). One such word with a
// single space is left alone, so ordinary openings ("Happy to help!") are never
// touched - unless it is written lowercase and the next word starts a sentence
// ("thinking This is a bit much!", live 2026-09-30 21:07).
// A word at the start of a line, and the whitespace after it.
var gopodLeadingWordPattern = regexp.MustCompile(`^([A-Za-z]+)([ \t]+)`)

func gopodIsAnimationWord(word string) bool {
	token := normalizeAnimationToken(word)
	if token == "" {
		return false
	}
	for _, animThing := range animationMap {
		if token == strings.ToLower(animThing[0]) {
			return true
		}
	}
	_, ok := animationAliases[token]
	return ok
}

func gopodStripLeadingAnimationWords(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		// 2026-10-23 (live: "🎉 celebrate So what's the plan, man?"): emoji and
		// marks before the first word are kept, and the word after them is
		// checked like a line's first word.
		indent := line[:len(line)-len(strings.TrimLeftFunc(line, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '"' && r != '\'' && r != '*'
		}))]
		rest := line[len(indent):]
		count, firstGap, cut, firstLower := 0, 0, 0, false
		for count < 3 {
			m := gopodLeadingWordPattern.FindStringSubmatchIndex(rest[cut:])
			if m == nil || !gopodIsAnimationWord(rest[cut+m[2]:cut+m[3]]) {
				break
			}
			if count == 0 {
				firstGap = m[5] - m[4]
				firstLower = rest[0] >= 'a' && rest[0] <= 'z'
			}
			count++
			cut += m[1]
		}
		nextCapital := cut < len(rest) && rest[cut] >= 'A' && rest[cut] <= 'Z'
		if count >= 2 || (count == 1 && (firstGap >= 2 || (firstLower && nextCapital))) {
			lines[i] = indent + rest[cut:]
		}
	}
	return strings.Join(lines, "\n")
}

// The model now often opens with its animation cue as "happy." or "thinking."
// (one dot) instead of "happy...", live 2026-09-30 21:47-21:55. Wire-Pod already
// plays it as the animation and doesn't say it, but the chat bubble showed
// "happy. hey there!". gopodStripAnimationCueForDisplay drops that opening cue
// from the displayed reply only - a lowercase animation word with one or three
// dots at the very start. "Happy. Me? Always." (capitalised) is left alone.
var gopodOpeningCuePattern = regexp.MustCompile(`^\s*([a-z][A-Za-z]*)(?:\.\.\.|\.)\s+`)

func gopodStripAnimationCueForDisplay(text string) string {
	m := gopodOpeningCuePattern.FindStringSubmatchIndex(text)
	if m == nil || !gopodIsAnimationTokenName(text[m[2]:m[3]]) {
		return text
	}
	return text[m[1]:]
}

// ## EOF

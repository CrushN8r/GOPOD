// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_leading_animation_words.go
package wirepod_ttr

import (
	"regexp"
	"strings"
	"unicode"
)

// The model sometimes opens a reply (or a paragraph) with animation words and a wide gap, instead
// of the "happy...
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

// The model now often opens with its animation cue as "happy." or "thinking." (one dot) instead of
// "happy...", live 47-21:55.
var gopodOpeningCuePattern = regexp.MustCompile(`^\s*([a-z][A-Za-z]*)(?:\.\.\.|\.)\s+`)

func gopodStripAnimationCueForDisplay(text string) string {
	m := gopodOpeningCuePattern.FindStringSubmatchIndex(text)
	if m == nil || !gopodIsAnimationTokenName(text[m[2]:m[3]]) {
		return text
	}
	return text[m[1]:]
}

// ## EOF

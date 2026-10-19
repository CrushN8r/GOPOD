// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_spoken_dashes.go
package wirepod_ttr

import (
	"regexp"
	"unicode"
)

// A dash used as a pause between words is read out as "dash" by the robot (live
// 2026-09-29: "bringing the hype - the verbal boom"). It becomes a comma pause.
// Covered: a spaced dash ("a - b", "a -- b"), a dash hanging off a word ("are- but"),
// and a double dash inside words ("laugh--that"). Left alone: a hyphen inside a word
// (Gowp-awd, free-spirited, AY-EYE) and a dash between two numbers ("10 - 3").
var gopodSpacedDashPattern = regexp.MustCompile(`(\S)[ \t]+-{1,2}[ \t]+(\S)`)
var gopodHangingDashPattern = regexp.MustCompile(`(\pL)-{1,2}[ \t]+(\S)`)
var gopodDoubleDashPattern = regexp.MustCompile(`(\pL)--(\pL)`)

func gopodSpokenDashes(text string) string {
	fix := func(p *regexp.Regexp) {
		// Twice, so back-to-back dashes ("a - b - c") share their middle word.
		for i := 0; i < 2; i++ {
			text = p.ReplaceAllStringFunc(text, func(m string) string {
				parts := p.FindStringSubmatch(m)
				left, right := []rune(parts[1]), []rune(parts[2])
				if unicode.IsDigit(left[len(left)-1]) && unicode.IsDigit(right[0]) {
					return m
				}
				return parts[1] + ", " + parts[2]
			})
		}
	}
	fix(gopodSpacedDashPattern)
	fix(gopodHangingDashPattern)
	fix(gopodDoubleDashPattern)
	return text
}

// ## EOF

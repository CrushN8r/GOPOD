// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_spoken_dashes.go
package wirepod_ttr

import (
	"regexp"
	"unicode"
)

// A dash used as a pause between words is read out as "dash" by the robot.
var gopodSpacedDashPattern = regexp.MustCompile(`(\S)[ \t]+-{1,2}[ \t]+(\S)`)
var gopodHangingDashPattern = regexp.MustCompile(`(\pL)-{1,2}[ \t]+(\S)`)
var gopodDoubleDashPattern = regexp.MustCompile(`(\pL)--(\pL)`)

func gopodSpokenDashes(text string) string {
	fix := func(p *regexp.Regexp) {
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

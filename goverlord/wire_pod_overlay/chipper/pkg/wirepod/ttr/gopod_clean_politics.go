// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_clean_politics.go
package wirepod_ttr

import (
	"regexp"
	"sort"
	"strings"
)

// Clean GOPOD (2026-10-16, operator, on Brobot 2's 🇵🇸 flag in a "Jerusalem,
// Israel" weather reply: "Please account for obvious politically motivated
// bullshit in clean GOPOD * politically motivated bullshit only allowed in roast
// battle"; "RULE: use correct & proper country flag emojis"). Outside a roast
// battle (clean_politics.allowed_topics), a Brobot's reply keeps a country flag
// only when its country is named in the question or the reply; a wrong flag
// becomes the named country's flag when exactly one is named, else it goes. A
// sentence with one of clean_politics.words goes. Settings: clean_politics.

// gopodFlagRE: a country flag - two regional-indicator letters.
var gopodFlagRE = regexp.MustCompile(`[\x{1F1E6}-\x{1F1FF}]{2}`)

func gopodFlagCode(flag string) string {
	var b strings.Builder
	for _, r := range flag {
		b.WriteRune('A' + (r - 0x1F1E6))
	}
	return b.String()
}

func gopodFlagFor(code string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(code) {
		b.WriteRune(0x1F1E6 + (r - 'A'))
	}
	return b.String()
}

// gopodNamedCountries: the flag codes of the countries named in text, once each.
func gopodNamedCountries(text string) []string {
	lower := strings.ToLower(text)
	seen := map[string]bool{}
	var out []string
	names := make([]string, 0, len(gopodBrobotsSettings.CleanPolitics.Countries))
	for name := range gopodBrobotsSettings.CleanPolitics.Countries {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		code := strings.ToUpper(gopodBrobotsSettings.CleanPolitics.Countries[name])
		if code != "" && !seen[code] && gopodHasWholeWords(lower, strings.ToLower(name)) {
			seen[code] = true
			out = append(out, code)
		}
	}
	return out
}

// gopodCleanPoliticsAllowed: the question is in an allowed topic (a roast battle).
func gopodCleanPoliticsAllowed(question string) bool {
	for _, p := range loadScenarioPacketsForMessage(question) {
		for _, t := range gopodBrobotsSettings.CleanPolitics.AllowedTopics {
			if p.Name == t {
				return true
			}
		}
	}
	return false
}

// gopodCleanPolitics: the reply, clean (see above).
func gopodCleanPolitics(reply, question string) string {
	cp := gopodBrobotsSettings.CleanPolitics
	if len(cp.Countries) == 0 && len(cp.Words) == 0 {
		return reply
	}
	if gopodCleanPoliticsAllowed(question) {
		return reply
	}
	named := gopodNamedCountries(question + " " + gopodFlagRE.ReplaceAllString(reply, " "))
	// 2026-10-17: the weather line's own flag is its place's country.
	named = append(named, gopodWeatherCountriesSaid()...)
	reply = gopodFlagRE.ReplaceAllStringFunc(reply, func(flag string) string {
		code := gopodFlagCode(flag)
		for _, n := range named {
			if n == code {
				return flag
			}
		}
		if len(named) == 1 {
			return gopodFlagFor(named[0])
		}
		return ""
	})
	if len(cp.Words) > 0 {
		reply = gopodEchoSentenceRe.ReplaceAllStringFunc(reply, func(sentence string) string {
			lower := strings.ToLower(sentence)
			for _, w := range cp.Words {
				if w = strings.ToLower(strings.TrimSpace(w)); w != "" && gopodHasWholeWords(lower, w) {
					return ""
				}
			}
			return sentence
		})
	}
	return strings.TrimSpace(regexp.MustCompile(`[ \t]{2,}`).ReplaceAllString(reply, " "))
}

// ## EOF

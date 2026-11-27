// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_clean_politics.go
package wirepod_ttr

import (
	"regexp"
	"sort"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

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
	reply = gopodDropNeverWords(reply)
	if len(cp.Countries) == 0 && len(cp.Words) == 0 {
		return reply
	}
	if gopodCleanPoliticsAllowed(question) {
		return reply
	}
	named := gopodNamedCountries(question + " " + gopodFlagRE.ReplaceAllString(reply, " "))
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

// gopodNeverIn: a line coming in names a never word - it goes no further.
func gopodNeverIn(text, from string) bool {
	lower := strings.ToLower(text)
	for _, w := range gopodBrobotsSettings.CleanPolitics.NeverWords {
		if w = strings.ToLower(strings.TrimSpace(w)); w != "" && strings.Contains(lower, w) {
			logger.Println("GOPOD_CLEAN_IN a line from " + from + " named a never word - dropped before the page and the model")
			return true
		}
	}
	return false
}

// gopodDropNeverWords: every sentence saying a never word goes - any voice, any topic
// (clean_politics.never_words).
func gopodDropNeverWords(text string) string {
	words := gopodBrobotsSettings.CleanPolitics.NeverWords
	if len(words) == 0 {
		return text
	}
	out := gopodEchoSentenceRe.ReplaceAllStringFunc(text, func(sentence string) string {
		lower := strings.ToLower(sentence)
		for _, w := range words {
			if w = strings.ToLower(strings.TrimSpace(w)); w != "" && strings.Contains(lower, w) {
				return ""
			}
		}
		return sentence
	})
	return strings.TrimSpace(regexp.MustCompile(`[ \t]{2,}`).ReplaceAllString(out, " "))
}

// ## EOF

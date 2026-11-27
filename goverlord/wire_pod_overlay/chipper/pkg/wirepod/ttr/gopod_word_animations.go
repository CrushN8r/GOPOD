// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_word_animations.go
package wirepod_ttr

import (
	"regexp"
	"strings"
)

var gopodSpokenWordPattern = regexp.MustCompile(`[A-Za-z]+`)

// gopodIsAnimationTokenName: one of the animation names themselves, or a misspelling of one
// ("hapy", "happpy", "veryhappy" - the aliases within two letters of their own token), so a cue
// only, never said.
func gopodIsAnimationTokenName(word string) bool {
	token := normalizeAnimationToken(word)
	if token == "" {
		return false
	}
	for _, animThing := range animationMap {
		if token == strings.ToLower(animThing[0]) {
			return true
		}
	}
	if _, ok := animationMisspellings[token]; ok {
		return true
	}
	if target, ok := animationAliases[token]; ok {
		return gopodEditDistance(token, strings.ToLower(target)) <= 2
	}
	return false
}

func gopodEditDistance(a, b string) int {
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			best := prev[j] + 1
			if cur[j-1]+1 < best {
				best = cur[j-1] + 1
			}
			if prev[j-1]+cost < best {
				best = prev[j-1] + cost
			}
			cur[j] = best
		}
		prev = cur
	}
	return prev[len(b)]
}

// gopodAnimationFromWords: the animation named by the first token name or alias in the text, or ""
// when there is none.
func gopodAnimationFromWords(text string) string {
	for _, word := range gopodSpokenWordPattern.FindAllString(text, -1) {
		if known, ok := knownAnimationParameter(word); ok {
			return known
		}
		if alias, ok := knownGOPODActionAlias(word); ok {
			return alias
		}
	}
	return ""
}

// gopodPlayOnWords: the piece holds one of word_animations.play_on_words (the Gomads "go mad" pun),
// so its animation is the joke's, not the word's.
func gopodPlayOnWords(text string) bool {
	lower := " " + strings.Join(strings.Fields(strings.ToLower(text)), " ") + " "
	for _, p := range gopodBrobotsSettings.WordAnimations.PlayOnWords {
		if p != "" && strings.Contains(lower, strings.ToLower(p)) {
			return true
		}
	}
	return false
}

// gopodAnimationsFromWords sets the animation before every spoken piece whose words name one: the
// animation right before it is replaced, or one is added.
func gopodAnimationsFromWords(actions []RobotAction) []RobotAction {
	wa := gopodBrobotsSettings.WordAnimations
	var out []RobotAction
	for _, a := range actions {
		if a.Action == ActionSayText {
			anim := gopodAnimationFromWords(a.Parameter)
			if wa.PlayOnWordsAnimation != "" && gopodPlayOnWords(a.Parameter) {
				anim = wa.PlayOnWordsAnimation
			}
			if anim != "" {
				if n := len(out); n > 0 && out[n-1].Action == ActionPlayAnimationWI {
					out[n-1].Parameter = anim
				} else {
					out = append(out, RobotAction{Action: ActionPlayAnimationWI, Parameter: anim})
				}
			}
		}
		out = append(out, a)
	}
	for i := range out {
		if swap, ok := wa.Swap[out[i].Parameter]; ok && out[i].Action == ActionPlayAnimationWI {
			out[i].Parameter = swap
		}
	}
	return out
}

// ## EOF

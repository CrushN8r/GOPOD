// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_animation_variety.go
package wirepod_ttr

import (
	"regexp"
	"strings"
)

// Every part of a chat reply starts with an animation, {{playAnimationWI||x}}, taken from the
// emotion word the model is asked to begin with ("happy...
var gopodAnimationVariety = []string{"thinking", "happy", "veryHappy", "happy"}
var gopodAnimationVarietyNext int

var gopodReplyEmotionCuePattern = regexp.MustCompile(`(?im)^[\s|0-9C]*(happy|veryHappy|sad|verySad|angry|frustrated|dartingEyes|confused|thinking|celebrate|love)\s*\.\.\.`)

func gopodVaryDefaultAnimation(commandPacket, rawReply string) string {
	if gopodReplyEmotionCuePattern.MatchString(rawReply) {
		return commandPacket
	}
	defaultCue := formatAnimationCommand("thinking")
	if !strings.HasPrefix(commandPacket, defaultCue) {
		return commandPacket
	}
	pick := gopodAnimationVariety[gopodAnimationVarietyNext%len(gopodAnimationVariety)]
	gopodAnimationVarietyNext++
	return formatAnimationCommand(pick) + strings.TrimPrefix(commandPacket, defaultCue)
}

// gopodAnimationLeadIn opens the sentence right after an animation in a chat reply, the way native
// Wire-Pod does it: the model's "{{playAnimationWI||happy}}...
var gopodAnimationLeadIn = gopodBrobotsSettings.AnimationLeadIn

// gopodKeepLeadIn runs clean on the text after a leading lead-in and puts the lead-in back, so the
// cleanup (which drops "...") can't remove the native pause.
func gopodKeepLeadIn(text string, clean func(string) string) string {
	trimmed := strings.TrimLeft(text, " \t")
	if gopodAnimationLeadIn == "" || !strings.HasPrefix(trimmed, gopodAnimationLeadIn) {
		return clean(text)
	}
	rest := clean(strings.TrimPrefix(trimmed, gopodAnimationLeadIn))
	if rest == "" {
		return ""
	}
	return gopodAnimationLeadIn + " " + rest
}

// gopodSprinkleAnimations adds extra animations inside a longer spoken part: after its 2nd
// sentence, when the part has 3 or more, and since a token in the middle of a sentence hung
// say_text; speech first, then a separate animation step, ran clean.
var gopodSprinkleVariety = gopodBrobotsSettings.ExtraAnimations.Animations
var gopodSprinkleNext int

var gopodSprinkleMinSentences = gopodBrobotsSettings.ExtraAnimations.MinSentences

func gopodSprinkleAnimations(actions []RobotAction) []RobotAction {
	sentences := 0
	for _, a := range actions {
		if a.Action == ActionSayText {
			sentences++
		}
	}
	if sentences < gopodSprinkleMinSentences {
		return actions
	}
	var out []RobotAction
	said := 0
	for _, a := range actions {
		out = append(out, a)
		if a.Action != ActionSayText {
			continue
		}
		said++
		if gopodSprinkleAfter(said) && said < sentences {
			pick := gopodSprinkleVariety[gopodSprinkleNext%len(gopodSprinkleVariety)]
			gopodSprinkleNext++
			out = append(out, RobotAction{Action: ActionPlayAnimationWI, Parameter: pick})
		}
	}
	return out
}

// gopodSprinkleAfter: an extra animation goes after this sentence number
// (extra_animations.after_sentences).
func gopodSprinkleAfter(said int) bool {
	for _, n := range gopodBrobotsSettings.ExtraAnimations.AfterSentences {
		if n == said {
			return true
		}
	}
	return false
}

// ## EOF

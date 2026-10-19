// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_animation_variety.go
package wirepod_ttr

import (
	"regexp"
	"strings"
)

// Every part of a chat reply starts with an animation, {{playAnimationWI||x}},
// taken from the emotion word the model is asked to begin with ("happy... ").
// gemma2:2b mostly leaves it out, so every part played "thinking" (live
// 2026-09-29). Pushing the model to write it was tested and dropped: with emotion
// words in the chat history it wrote one 16/18 times, but then acted the word out
// ("confused... Dude, I'm totally lost!"). So when the reply has no emotion word
// at all, Wire-Pod varies the animation a little itself, over a small set of
// friendly ones, carrying on from reply to reply so a one-part reply varies too.
// A reply where the model did write one is left as it is.
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

// gopodAnimationLeadIn opens the sentence right after an animation in a chat
// reply, the way native Wire-Pod does it: the model's "{{playAnimationWI||happy}}...
// text" reaches the robot's voice as "... text", and the voice's own short pause
// lets the clip start before the words (operator 2026-10-01: "playAnimationWI +
// pause + robot speech (aim for native to wire-pod vosk)" - no fixed wait, which
// "breaks the performance flow"). The value is animation_lead_in in
// gopod_brobots_settings.json; GOPOD's speech cleanup keeps it
// (gopodKeepLeadIn).
var gopodAnimationLeadIn = gopodBrobotsSettings.AnimationLeadIn

// gopodKeepLeadIn runs clean on the text after a leading lead-in and puts the
// lead-in back, so the cleanup (which drops "...") can't remove the native pause.
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

// gopodSprinkleAnimations adds extra animations inside a longer spoken part:
// after its 2nd sentence, when the part has 3 or more (operator 2026-09-29: "Can
// any playAnimationWI animations be added within throughout sparingly?"), and
// since 2026-10-01 also after its 4th and 6th when more sentences follow
// (operator: "squeeze in another 1 or 2 playAnimationWI animations per
// response") - up to 3, never after the last sentence.
// Wire-Pod's own LLM command list (kgsim_cmds.go ValidLLMCommands) describes
// playAnimationWI as playing "without interrupting speech ... great for
// storytelling ... Don't put two of these right next to each other."
// GOPOD's own crash history decides the shape (gopod_notes/older_notes/
// ANGRY_CLIP_RETEST_AND_ANIMATION_PICKER_001.md, FOUR_CLEAN_RUNS_ANGRY_FIXED_001.md):
// a token in the middle of a sentence hung say_text; speech first, then a separate
// animation step, ran clean. So the extra animation is its own step between two
// sentences, never inside one, and only friendly clips are used (never angry).
// The clips, the positions and the minimum length are extra_animations in
// gopod_brobots_settings.json.
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

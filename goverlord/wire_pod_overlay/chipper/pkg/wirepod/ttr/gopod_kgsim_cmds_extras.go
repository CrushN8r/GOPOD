package wirepod_ttr

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/sashabaranov/go-openai"
)

func promptUsesCanonicalBrobotRawShape(prompt string) bool {
	if strings.Contains(prompt, "{{[emotion]}}") &&
		(strings.Contains(prompt, gopodCanonicalStartMarker) ||
			strings.Contains(prompt, gopodCanonicalChunkMarker) ||
			strings.Contains(prompt, gopodCanonicalEndMarker)) {
		return true
	}
	return strings.Contains(prompt, "actionParameter...") &&
		strings.Contains(prompt, "Valid action parameters:")
}

func promptIsBrobotPrompt(prompt string) bool {
	return strings.Contains(prompt, "Bro! Bots:") ||
		(strings.Contains(prompt, "Brobot") && strings.Contains(prompt, "GOPOD Yourself"))
}

func robotSpeechFromActions(actions []RobotAction) string {
	var speech []string
	for _, action := range actions {
		if action.Action != ActionSayText {
			continue
		}
		text := cleanRobotSpeechText(action.Parameter)
		if text != "" {
			speech = append(speech, text)
		}
	}
	return strings.Join(speech, " ")
}

func stripSpokenActionLead(input string) string {
	cleaned := strings.TrimSpace(input)
	for {
		before := cleaned
		lower := strings.ToLower(cleaned)
		for _, prefix := range []string{
			"actionparameter:",
			"actionparameter",
			strings.ToLower(gopodBrobotsSettings.PlaceholderThought) + ":",
			strings.ToLower(gopodBrobotsSettings.PlaceholderThought),
		} {
			if strings.HasPrefix(lower, prefix) {
				cleaned = strings.TrimSpace(cleaned[len(prefix):])
				lower = strings.ToLower(cleaned)
			}
		}
		for _, animThing := range animationMap {
			action := animThing[0]
			if strings.HasPrefix(cleaned, action+" ") {
				// 2026-10-17 (operator: "celebrate skipped? not heard" - "react,
				// celebrate small wins," was said "small wins,"): a cue word is
				// dropped only when what follows starts anew (a capital, or a wide
				// gap: "happy Hey there", "thinking  well"); a piece that just goes
				// on ("celebrate small wins,") keeps it as a word.
				rest := strings.TrimPrefix(cleaned, action+" ")
				if next := []rune(strings.TrimSpace(rest)); len(next) > 0 && !unicode.IsUpper(next[0]) && !strings.HasPrefix(rest, " ") {
					continue
				}
				cleaned = strings.TrimSpace(rest)
				break
			}
			if strings.HasPrefix(cleaned, action+".") {
				cleaned = strings.TrimSpace(strings.TrimPrefix(cleaned, action+"."))
				break
			}
			if strings.HasPrefix(cleaned, action+":") {
				cleaned = strings.TrimSpace(strings.TrimPrefix(cleaned, action+":"))
				break
			}
		}
		if cleaned == before {
			return cleaned
		}
	}
}

// ". ." left where two "..." met (a dropped stage direction between them) -> ".".
var gopodRepeatedPeriodPattern = regexp.MustCompile(`\.(?:\s*\.)+`)

func cleanRobotSpeechText(input string) string {
	cleaned := normalizeBrobotRobotCopy(input)
	cleaned = ansiEscapePattern.ReplaceAllString(cleaned, "")
	cleaned = stripCanonicalGOPODMarkers(cleaned)
	cleaned = strings.ReplaceAll(cleaned, "|||", " ")
	if strings.Contains(cleaned, "{{") {
		cleaned = robotSpeechFromActions(GetActionsFromString(cleaned))
	}
	cleaned = strings.ReplaceAll(cleaned, "playAnimationWI", "")
	cleaned = strings.TrimSpace(cleaned)
	for {
		next := strings.TrimSpace(strings.TrimPrefix(cleaned, "..."))
		next = strings.TrimSpace(strings.TrimPrefix(next, "."))
		if next == cleaned {
			break
		}
		cleaned = next
	}
	if units := parseBrobotPhraseUnits(cleaned); len(units) == 1 {
		cleaned = units[0].Text
	}
	cleaned = strings.TrimSpace(cleaned)
	cleaned = stripSpokenActionLead(cleaned)
	// GOPOD 2026-10-01: a "..." left in speech is no longer a sentence end
	// (operator: take "..." out as a sentence ender or phrase breaker); after a
	// word it is said as a comma pause (operator, 17:1x: '"..." = ","'),
	// gopodEllipsisToComma. "fingers...a hidden gem" still isn't glued together.
	// The Fields join below tidies spaces.
	cleaned = gopodEllipsisToComma(cleaned)
	cleaned = gopodDropStrayQuotes(cleaned)
	cleaned = gopodRepeatedPeriodPattern.ReplaceAllString(cleaned, ".")
	for {
		next := strings.TrimSpace(strings.TrimPrefix(cleaned, "..."))
		next = strings.TrimSpace(strings.TrimPrefix(next, "."))
		if next == cleaned {
			break
		}
		cleaned = next
	}
	cleaned = stripSpokenActionLead(cleaned)
	return strings.Join(strings.Fields(cleaned), " ")
}

// GOPOD 2026-09-29: every spoken part of a Wire-Pod chat reply ends with
// punctuation, so the robot pauses before the next part (used in StreamingKGSim's
// speaking loop only - song speech is left exactly as written). Bullet lines often had none ("...a little wink on that
// side" ran straight into "Gomads: ..."), and a reply could end on a dash ("with
// the right tools-"). A trailing dash or comma is dropped, then a period added.
var gopodTrailingDashPattern = regexp.MustCompile(`[\s,\-\x{2013}\x{2014}]+$`)

var gopodTrailingDashOnlyPattern = regexp.MustCompile(`[\s\-\x{2013}\x{2014}]+$`)

func gopodEndWithPunctuation(text string) string {
	original := text
	text = gopodTrailingDashPattern.ReplaceAllString(text, "")
	if text == "" {
		return text
	}
	switch text[len(text)-1] {
	case '.', '!', '?', ':', ';', '"', ')':
		return text
	}
	// A piece the comma split ended (live_split) keeps its comma - said as a
	// pause, not a full stop.
	if strings.HasSuffix(strings.TrimSpace(gopodTrailingDashOnlyPattern.ReplaceAllString(original, "")), ",") {
		return text + ","
	}
	return text + "."
}

func connectVector(esn string) (*vector.Vector, error) {
	for _, bot := range vars.BotInfo.Robots {
		if esn == bot.Esn {
			robot, err := vector.New(
				vector.WithSerialNo(esn),
				vector.WithToken(bot.GUID),
				vector.WithTarget(bot.IPAddress+":443"),
			)
			return GopodRobotsOffConn(robot), err // robots off (gopod_robots_off.go)
		}
	}
	return nil, fmt.Errorf("robot ESN not found: %s", esn)
}

type liveRobotActionDispatcher struct {
	robot *vector.Vector
}

func (d liveRobotActionDispatcher) PerformActions(msgs []openai.ChatCompletionMessage, actions []RobotAction, stopStop chan bool) bool {
	return performActionsLiveDispatchUnsafe(msgs, actions, d.robot, stopStop)
}

func performActionsLiveDispatchUnsafe(msgs []openai.ChatCompletionMessage, actions []RobotAction, robot *vector.Vector, stopStop chan bool) bool {
	stopPerforming := false
	go func() {
		for range stopStop {
			stopPerforming = true
		}
	}()
	for _, action := range actions {
		if stopPerforming {
			return false
		}
		switch {
		case action.Action == ActionSayText:
			DoSayText(action.Parameter, robot)
		case action.Action == ActionPlayAnimation:
			DoPlayAnimation(action.Parameter, robot)
		case action.Action == ActionPlayAnimationWI:
			DoPlayAnimationWI(action.Parameter, robot)
		case action.Action == ActionNewRequest:
			go DoNewRequest(robot)
			return true
		case action.Action == ActionGetImage:
			DoGetImage(msgs, action.Parameter, robot, stopStop)
			return true
		case action.Action == ActionPlaySound:
			DoPlaySound(action.Parameter, robot)
		}
	}
	WaitForAnim_Queue(robot.Cfg.SerialNo)
	return false
}

// gopodDropStrayQuotes: a single quote touching other punctuation ("-'bizzy",
// "business.'.") was read out by the robot as "single quote" (live 2026-10-02
// 22:12). It is dropped. One next to a letter and a space or a letter stays, so
// don't, ya', 'cause and nothin' are said as before.
func gopodDropStrayQuotes(text string) string {
	runes := []rune(text)
	var b strings.Builder
	isPunct := func(i int) bool {
		if i < 0 || i >= len(runes) {
			return false
		}
		r := runes[i]
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsSpace(r)
	}
	for i, r := range runes {
		if (r == '\'' || r == '‘' || r == '’') && (isPunct(i-1) || isPunct(i+1)) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

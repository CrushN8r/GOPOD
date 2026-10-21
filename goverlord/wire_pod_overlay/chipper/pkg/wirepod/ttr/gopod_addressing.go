// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_addressing.go
package wirepod_ttr

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// Only the Brobot you call answers (operator 2026-10-04: "The vocal rule is to
// call for "Brobot 1" or "Brobot 2"", "one short animation & default to first
// brobot heard, answers"). "Hey Vector" wakes both robots in their own firmware,
// and each sends its own transcript. Before the chat request:
//   - a question that opens with a Brobot's name (addressing.names, the "brobot"
//     forms first, "robot" only if none match) is answered by that Brobot only;
//   - a question with no name is answered by the Brobot heard first; the other
//     one, hearing the same question within addressing.first_heard_seconds,
//     stays quiet.
// The quiet Brobot plays addressing.quiet_animation once and goes back to idle:
// no chat request, no speech, nothing in the shared chat or the session memory.

// gopodAddressedTo returns the Brobot a question opens with ("Brobot 1",
// "Brobot 2"), or "" when it names none. The name counts only within the first
// addressing.start_within_words words ("hey brobot one, ..."), so "tell me about
// your brother bro bought one" names nobody.
func gopodAddressedTo(text string) string {
	to, _, _ := gopodFindCall(text)
	return to
}

// gopodFindCall finds the Brobot a question opens with and the byte span of the
// name as heard in text.
func gopodFindCall(text string) (string, int, int) {
	spans := gopodAddressWordSpans(text)
	words := make([]string, len(spans))
	for i, sp := range spans {
		words[i] = strings.ToLower(text[sp[0]:sp[1]])
	}
	a := gopodBrobotsSettings.Addressing
	for _, kind := range []string{"brobot", "robot"} {
		for _, brobot := range []string{"Brobot 1", "Brobot 2"} {
			for _, form := range a.Names[brobot][kind] {
				name := gopodAddressWords(form)
				for start := 0; start < a.StartWithinWords && start+len(name) <= len(words); start++ {
					if gopodWordsEqual(words[start:start+len(name)], name) {
						return brobot, spans[start][0], spans[start+len(name)-1][1]
					}
				}
			}
		}
	}
	return "", 0, 0
}

// gopodCanonicalCall writes the opening call as the Brobot's own name ("bro
// bought two, can you hear me" -> "Brobot 2, can you hear me"), so the chat
// model sees it's being called (operator 2026-10-04: "brobots acknowledge my
// heard call").
func gopodCanonicalCall(text string) string {
	to, from, until := gopodFindCall(text)
	if to == "" {
		return text
	}
	return text[:from] + to + text[until:]
}

// gopodAddressWords lowercases and splits on anything not a letter or digit.
func gopodAddressWords(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), gopodAddressSeparator)
}

func gopodAddressSeparator(r rune) bool {
	r = []rune(strings.ToLower(string(r)))[0]
	return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '\'')
}

// gopodAddressWordSpans returns each word's [start, end) byte span in text.
func gopodAddressWordSpans(text string) [][2]int {
	var spans [][2]int
	start := -1
	for i, r := range text {
		if gopodAddressSeparator(r) {
			if start >= 0 {
				spans = append(spans, [2]int{start, i})
				start = -1
			}
		} else if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		spans = append(spans, [2]int{start, len(text)})
	}
	return spans
}

func gopodWordsEqual(a, b []string) bool {
	if len(a) != len(b) || len(b) == 0 {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// The last no-name question a Brobot took, so his brother, hearing the same
// question, stays quiet.
var gopodFirstHeard = struct {
	sync.Mutex
	esn   string
	words []string
	at    time.Time
}{}

// gopodShouldAnswer decides whether this Brobot answers; to is the Brobot named
// ("" for none).
func gopodShouldAnswer(esn, text string, now time.Time) (answer bool, to string) {
	answer, to = gopodDecideAnswer(esn, text, now)
	if !answer {
	}
	return answer, to
}

// gopodAnswerNext: a Brobot who answers the next line even though his brother
// heard it first ("All", gopodTypedInputAll).
var gopodAnswerNext sync.Map

// gopodAllFirstAnswer: the Brobot answering first in a line to both - his answer
// ends with addressing.redirect_text (2026-10-21).
var gopodAllFirstAnswer sync.Map

// gopodFollowingUp: a Brobot answering after his brother on the same line (All,
// Explain the Math) - operator 2026-10-15: '"All" = Brobot 1 \'answers\',
// Brobot 2 \'follows up\', no repeats'. While it's set his reply follows up:
// addressing.follow_up_note before the question (math.follow_up_note for
// Explain the Math). His weather line (his own robot's units) and reminder stay
// (operator: "keep Brobot 2 repeats the weather line"; "keep the 'Hey, that
// reminds me...' lines as they were").
var gopodFollowingUp sync.Map

func gopodIsFollowingUp(esn string) bool {
	_, ok := gopodFollowingUp.Load(esn)
	return ok
}

func gopodDecideAnswer(esn, text string, now time.Time) (bool, string) {
	if _, ok := gopodAnswerNext.LoadAndDelete(esn); ok {
		return true, ""
	}
	self := gopodBrobotName(esn)
	if to := gopodAddressedTo(text); to != "" {
		return to == self, to
	}
	window := gopodSecondsToDuration(gopodBrobotsSettings.Addressing.FirstHeardSeconds)
	// The brother heard it as a voice command ("bingo test live" on one, "the
	// goat test live" on the other, live 2026-10-04 10:3x): he took it.
	if gopodBrotherRanCommand(esn, now, window) {
		return false, ""
	}
	words := gopodAddressWords(text)
	gopodFirstHeard.Lock()
	defer gopodFirstHeard.Unlock()
	if gopodFirstHeard.esn != "" && gopodFirstHeard.esn != esn && now.Sub(gopodFirstHeard.at) <= window &&
		gopodSameQuestion(gopodFirstHeard.words, words) {
		return false, ""
	}
	gopodFirstHeard.esn, gopodFirstHeard.words, gopodFirstHeard.at = esn, words, now
	return true, ""
}

// gopodBrotherRanCommand: the other Brobot ran a custom voice command (Wire-Pod's
// "Bot <ESN> Executing:" log line) within window before now.
func gopodBrotherRanCommand(esn string, now time.Time, window time.Duration) bool {
	brother := gopodBrobot1ESN
	if strings.TrimSpace(esn) == gopodBrobot1ESN {
		brother = gopodBrobot2ESN
	}
	lines := append([]string(nil), logger.LogTrayArray...)
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if len(line) < 21 {
			continue
		}
		at, err := time.ParseInLocation("2006.01.02 15:04:05", line[:19], time.Local)
		if err != nil {
			continue
		}
		// Log stamps are whole seconds.
		if now.Sub(at) > window+time.Second {
			return false
		}
		if strings.Contains(line, "Bot "+brother+" Executing: ") {
			return true
		}
	}
	return false
}

// gopodSameQuestion: the two robots' transcripts of one question - each has its
// own mics, so a word may differ - share at least half their words.
func gopodSameQuestion(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	seen := map[string]bool{}
	for _, w := range a {
		seen[w] = true
	}
	shared := 0
	for _, w := range b {
		if seen[w] {
			shared++
		}
	}
	longest := len(a)
	if len(b) > longest {
		longest = len(b)
	}
	return shared*2 >= longest
}

// gopodAddressingLog writes the BROBOT_ADDRESSED line.
func gopodAddressingLog(esn, to string, answer bool) {
	if to == "" {
		to = "none"
	}
	action := "answer"
	if !answer {
		action = "quiet"
	}
	logger.Println(fmt.Sprintf("BROBOT_ADDRESSED to=%s self=%s action=%s", to, gopodBrobotName(esn), action))
}

// gopodQuietBrother ends this Brobot's turn: one short animation under behavior
// control, the voice request closed the way a chat reply closes it, control back.
func gopodQuietBrother(req interface{}, robot *vector.Vector, esn, text string, isKG bool) {
	ctx := context.Background()
	granted := make(chan bool, 1)
	release := make(chan bool)
	gopodLogBControl(robot, ctx, granted, release)
	select {
	case <-granted:
		if clip := gopodAnimationClip(gopodBrobotsSettings.Addressing.QuietAnimation); clip != "" {
			gopodLogPlayAnimation(robot, ctx, &vectorpb.PlayAnimationRequest{
				Animation: &vectorpb.Animation{Name: clip},
				Loops:     1,
			})
		}
	case <-time.After(5 * time.Second):
		logger.Println("BROBOT_ADDRESSED quiet animation skipped: no behavior control")
	}
	if !isKG && req != nil {
		// No transcribed text: the chat page makes the operator's bubble from this
		// log line, and the answering brother already logs it (one bubble, not two).
		IntentPass(req, "intent_greeting_hello", "", map[string]string{}, false)
	}
	select {
	case release <- true:
	case <-time.After(2 * time.Second):
	}
}

// gopodAnimationClip: the Vector clip for an animation_vocab.json token, or "".
func gopodAnimationClip(token string) string {
	for _, anim := range animationMap {
		if anim[0] == token {
			return anim[1]
		}
	}
	return ""
}

// ## EOF

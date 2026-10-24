// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_typed_input.go
package wirepod_ttr

import (
	"sync"
	"bytes"
	"errors"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// Typed input from the chat page (operator 2026-10-04: "add a line for keyboard
// input at the very left ... radio buttons to select robot 1 or 2. And then a
// text input where I can type a command rather than speak"). The typed line goes
// the way a heard one does after speech-to-text: a GOPOD voice command
// (customIntents.json) runs its script; anything else is a Brobots chat question
// to the chosen Brobot. Native Wire-Pod intents ("sleep", ...) need the robot's
// own voice request, so typed they go to the chat.

// gopodTypedAll is the serial the page sends for its "All Brobots" button
// (operator 2026-10-04: "please add a 3rd All brobots () radio button"): both
// Brobots get the line, as when both hear a spoken one with no name - the
// addressing rule picks who answers (Brobot 1 first, his brother defers). A
// voice command runs once, on Brobot 1.
const gopodTypedAll = "all"

// GopodTypedInput handles one typed line for the robot with this ESN.
func GopodTypedInput(esn, text string) error {
	return gopodOperatorLine(esn, text, "Typed text")
}

// GopodVisitorTypedInput: an approved visitor's line, under their username.
func GopodVisitorTypedInput(esn, text, how string) error {
	return gopodOperatorLine(esn, text, how)
}

// GopodVoiceCommandInput: a line a voice command's script hands over (start
// interview, stop run, GOPOD Yourself) - the robot heard it, so the page already
// shows it as the operator's line and no second one is added.
func GopodVoiceCommandInput(esn, text string) error {
	return gopodOperatorLine(esn, text, gopodVoiceCommandHow)
}

const gopodVoiceCommandHow = "Voice command"

// gopodCapsTypoRE: a word typed small then capitals ("iS", "tHE").
var gopodCapsTypoRE = regexp.MustCompile(`\b\p{Ll}\p{Lu}+\b`)

// gopodCapsTypoFix (2026-10-16, live: "How iS the weather there today?" was
// read out as "i S"): a word typed small-then-capitals is made small ("is");
// "iPhone" and "eBay" stay as typed.
func gopodCapsTypoFix(text string) string {
	return gopodCapsTypoRE.ReplaceAllStringFunc(text, strings.ToLower)
}

// gopodOperatorLine: a typed line, or the words of a numpad push-to-talk capture
// (how names which in the log), for one Brobot or "all".
func gopodOperatorLine(esn, text, how string) error {
	text = gopodCapsTypoFix(strings.TrimSpace(text))
	if text == "" {
		return errors.New("empty text")
	}
	// A new question: an earlier pop-up's picture isn't its business.
	gopodPictureForget()
	// 2026-10-17 (operator: "system-set ... from web page keyboard as a
	// terminal-command"): a fixed list of edits, answered in a terminal bubble
	// (gopod_system_set.go); any time, warm-up or not.
	if gopodSystemSetMatch(text) {
		gopodSystemSetLine(esn, text, how)
		return nil
	}
	// 2026-10-13: during an LLM warm-up, every line waits for LLM Ready
	// (gopod_llm_warmup.go gopodWarmupHold).
	if gopodWarmupHold(esn, text, how) {
		return nil
	}
	// The answer to Record Session's keep-or-delete question, whichever Brobot is
	// chosen (gopod_record_session.go).
	if GopodRecordSessionAnswer(text, how) {
		return nil
	}
	// interview_script.play_words ("start interview") (gopod_interview_script.go).
	// 2026-10-10 (operator: "some system commands weren't displayed"): a command
	// line gets its Operator bubble and its read-out too.
	if gopodScriptCommandMatch(text) {
		gopodShowCommandLine(esn, text, how)
		gopodScriptCommand(text)
		return nil
	}
	// 2026-10-09: "stop run" stops a playing draft or a running song at its
	// next safe spot, any time (gopod_thanks_close_out.go).
	// 2026-10-11 (live: "No run is going." was heard before "stop run" - its voice
	// file was still being made): the line is shown and read out first, then acted on.
	if gopodStopRunMatch(text) {
		gopodShowCommandLine(esn, text, how)
		gopodStopRunCommand(text)
		return nil
	}
	// 2026-10-12 (operator: "replay [session-name]"): a session's score played
	// back (gopod_session_replay.go).
	if gopodStopSessionMatch(text) {
		gopodShowCommandLine(esn, text, how)
		gopodStopSessionCommand()
		return nil
	}
	if gopodReplayMatch(text) {
		gopodShowCommandLine(esn, text, how)
		gopodReplayCommand(text)
		return nil
	}
	// 2026-10-06: before the green light, typed and numpad lines go nowhere
	// (gopod_llm_warmup.go).
	if gopodWarmGateDrop(esn, how, text) {
		return nil
	}
	// 2026-10-10, Rule D (gopod_session.go): during a run only run words get
	// through - thanks (just below) closes the run out; the rest is dropped.
	if !gopodIsThanks(text) && gopodRunGateDrop(esn, how, text) {
		return nil
	}
	// 2026-10-09: a thanks while a draft or song is going closes them out
	// first, then comes back here (gopod_thanks_close_out.go).
	if gopodThanksCloseOut(esn, text, how) {
		return nil
	}
	// 2026-10-09: "GOPOD Yourself" gets each Brobot's call to action (gopod_yourself.go).
	if gopodYourselfCommand(esn, text, how) {
		return nil
	}
	// How it came in, for a recorded session's take (gopod_take_sheet.go).
	gopodRecordOperatorInput(how)
	if esn == gopodTypedAll {
		return gopodTypedInputAll(text, how)
	}
	if gopodBrobotName(esn) == "Brobot" {
		return errors.New("not a Brobot: " + esn)
	}
	logger.Println("Bot " + esn + " " + how + ": " + text)
	// 2026-10-07: a new chapter's title before the question, not after it
	// (gopod_interview_chapters.go; shown once, so the reply doesn't repeat it).
	gopodShowChapter(gopodSpellNames(text))
	gopodChatFeedAddOperator(gopodBrobotName(esn), text, gopodOperatorVia(how))
	// 2026-10-10 (operator: "\"start bingo test\" not spoken"): a typed run
	// command is read out first, then the run starts.
	if gopodTypedCommandMatch(text) {
		go func() {
			gopodReporterReadsOperator(text, how, gopodBrobotName(esn))
			gopodRunTypedCommand(esn, text)
		}()
		return nil
	}
	if gopodStartNotFound(text, func() { gopodReporterReadsOperator(text, how, gopodBrobotName(esn)) }) {
		return nil
	}
	if gopodRunNativeCube(esn, text) {
		return nil
	}
	go func() {
		// 2026-10-06: the reporter voice reads the line out first (gopod_reporter_voices.go).
		gopodReporterReadsOperator(text, how, gopodBrobotName(esn))
		// 2026-10-08: the topic's pictures after the question, before the reply.
		// 2026-10-13 (operator: "live pop up started before \"granny\" voice file
		// was generated"): once the question has been read out.
		gopodShowTopicPictures(esn, text)
		if _, err := StreamingKGSim(nil, esn, text, false); err != nil {
			logger.Println("LLM error (typed): " + err.Error())
		}
	}()
	return nil
}

// gopodTypedAll's gap between the two Brobots' requests, so Brobot 1 is the
// one heard first.
const gopodTypedAllGap = 100 * time.Millisecond

func gopodTypedInputAll(text, how string) error {
	logger.Println("Bot " + gopodBrobot1ESN + " " + how + " (all): " + text)
	gopodShowChapter(gopodSpellNames(text))
	gopodChatFeedAddOperator("Brobots", text, gopodOperatorVia(how))
	if gopodTypedCommandMatch(text) {
		go func() {
			gopodReporterReadsOperator(text, how, "Brobots")
			gopodRunTypedCommand(gopodBrobot1ESN, text)
		}()
		return nil
	}
	if gopodStartNotFound(text, func() { gopodReporterReadsOperator(text, how, "Brobots") }) {
		return nil
	}
	if gopodRunTypedCommand(gopodBrobot1ESN, text) {
		return nil
	}
	// A cube command to all: each Brobot runs it with his own cube.
	if gopodNativeCubeIntent(text) != "" {
		gopodRunNativeCube(gopodBrobot1ESN, text)
		gopodRunNativeCube(gopodBrobot2ESN, text)
		return nil
	}
	// 2026-10-06: the reporter voice reads the line out once, before both answer.
	read := make(chan struct{})
	go func() {
		gopodReporterReadsOperator(text, how, "Brobots")
		// 2026-10-13: the pictures once the question has been read out;
		// the Brobot answering first carries the pop-up.
		gopodShowTopicPictures(gopodAllFirst(text), text)
		close(read)
	}()
	// 2026-10-14: "Explain the Math" - Brobot 1 answers, and his follow-up by
	// Brobot 2 comes on its own (gopodMathFollowUpStart). 2026-10-16 (live: "convert
	// 0.1234 repeating into a fraction" ran the All sequence and the math
	// follow-up both - each Brobot answered twice): a fraction-word line too.
	if _, _, ok := gopodMathParts(text); ok {
		go func() {
			<-read
			if _, err := StreamingKGSim(nil, gopodBrobot1ESN, text, false); err != nil {
				logger.Println("LLM error (typed): " + err.Error())
			}
		}()
		return nil
	}
	// 2026-10-14 (operator: "if I ask a question to both (All), Brobot 1
	// responds, then Brobot 2 responds"): with no Brobot named, Brobot 1 answers,
	// then - once his answer is done - Brobot 2 answers the same line
	// (addressing.all_next_wait_seconds at most). A named Brobot answers alone.
	// 2026-10-17 (operator: "This question is for the both of you. Starting with
	// Brobot 2 ... this reverses, so brobot 2 answers, brobot [1] follows up * but
	// they maintain their slightly different coloured personas"): "starting with
	// Brobot 2" (or "Brobot 2 first") turns the order round.
	first := gopodAllFirst(text)
	if gopodAddressedTo(text) == "" || gopodAllStartsWith(text) != "" {
		second := gopodBrotherOf(first)
		go func() {
			<-read
			done := gopodReplyDone(first)
			failed := make(chan struct{})
			gopodAllFirstAnswer.Store(first, true) // his answer hands over to his brother (kgsim.go)
			if first != gopodBrobot1ESN {
				gopodAnswerNext.Store(first, true)
			}
			go func() {
				if _, err := StreamingKGSim(nil, first, text, false); err != nil {
					logger.Println("LLM error (typed): " + err.Error())
					close(failed)
				}
			}()
			select {
			case <-done:
			case <-failed:
			case <-time.After(gopodSecondsToDuration(gopodBrobotsSettings.Addressing.AllNextWaitSeconds)):
				logger.Println("GOPOD_ALL " + gopodBrobotName(first) + "'s answer didn't end in time")
			}
			logger.Println("GOPOD_ALL " + gopodBrobotName(second) + " answers next: " + text)
			gopodAnswerNext.Store(second, true)
			gopodFollowingUp.Store(second, true)
			if _, err := StreamingKGSim(nil, second, text, false); err != nil {
				logger.Println("LLM error (typed): " + err.Error())
			}
		}()
		return nil
	}
	for i, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN} {
		go func(esn string, wait time.Duration) {
			<-read
			time.Sleep(wait)
			if _, err := StreamingKGSim(nil, esn, text, false); err != nil {
				logger.Println("LLM error (typed): " + err.Error())
			}
		}(esn, time.Duration(i)*gopodTypedAllGap)
	}
	return nil
}

// gopodRunTypedCommand runs the first custom voice command whose phrase is in
// the typed text, the same match Wire-Pod uses for a heard one (system and Lua
// intents are left to voice). It logs "Bot <ESN> Executing:" like Wire-Pod, so
// the brother check in gopod_addressing.go sees it.
// gopodAllStartsWithRE: "starting with Brobot 2", "Brobot 2 first", "Pip first".
// 2026-10-17 (live: "starting with bro bought to"): the mic's spellings of
// "Brobot 2" too ("bro bought to", "brobot two").
var gopodAllStartsWithRE = regexp.MustCompile(`(?i)\b(?:start(?:ing)?\s+with|begin(?:ning)?\s+with)\s+(` + gopodAllNameRE + `)\b|\b(` + gopodAllNameRE + `)\s+(?:goes\s+|answers\s+)?first\b`)

const gopodAllNameRE = `(?:brobot|bro\s*bo(?:ugh)?t|bro\s*but)\s*(?:[12]|one|two|too|to)|doc|pip`

// gopodAllStartsWith: the Brobot a line to both says to start with ("Brobot 2"), or "".
func gopodAllStartsWith(text string) string {
	m := gopodAllStartsWithRE.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	who := strings.ToLower(strings.Join(strings.Fields(m[1]+m[2]), ""))
	switch {
	case who == "pip" || strings.HasSuffix(who, "2") || strings.HasSuffix(who, "two") || strings.HasSuffix(who, "to") || strings.HasSuffix(who, "too"):
		return "Brobot 2"
	case who == "doc" || strings.HasSuffix(who, "1") || strings.HasSuffix(who, "one"):
		return "Brobot 1"
	}
	return ""
}

// gopodAllFirst: the Brobot answering first in a line to both (Brobot 1 unless told).
func gopodAllFirst(text string) string {
	// The words first: "Starting with Brobot 2", then the Brobot named first in
	// the line (gopodFirstNamed), then the All menu (typed_line.all_first).
	switch gopodAllStartsWith(text) {
	case "Brobot 2":
		return gopodBrobot2ESN
	case "Brobot 1":
		return gopodBrobot1ESN
	}
	switch gopodFirstNamed(text) {
	case "Brobot 2":
		return gopodBrobot2ESN
	case "Brobot 1":
		return gopodBrobot1ESN
	}
	if gopodAllFirstPicked(text) == "2" {
		return gopodBrobot2ESN
	}
	return gopodBrobot1ESN
}

// The page's All menu (2026-10-22, operator: "replace 'All' radio button with a
// small drop down menu: 1/2 * default = '1' (brobot 1) * '2' = (brobot 2) *
// template drop down menu selection with default brobot 1/'Starting with Brobot
// 2'"): "2" is the same as "Starting with Brobot 2" for that line - kept by its
// words, so only that line is turned round (typed_line.all_first).
var gopodAllFirstPick = struct {
	sync.Mutex
	text, first string
}{}

// GopodAllFirstPick: the menu's choice for a typed line ("" = the default).
func GopodAllFirstPick(text, first string) {
	first = strings.TrimSpace(first)
	if first == "" {
		first = strings.TrimSpace(gopodBrobotsSettings.TypedLine.AllFirst)
	}
	gopodAllFirstPick.Lock()
	gopodAllFirstPick.text, gopodAllFirstPick.first = strings.TrimSpace(text), first
	gopodAllFirstPick.Unlock()
}

func gopodAllFirstPicked(text string) string {
	gopodAllFirstPick.Lock()
	defer gopodAllFirstPick.Unlock()
	if gopodAllFirstPick.text != "" && gopodAllFirstPick.text == strings.TrimSpace(text) {
		return gopodAllFirstPick.first
	}
	return ""
}

// gopodShowCommandLine: a command line's Operator bubble ("Keyboard to Brobot
// 1"), read out in the background.
func gopodShowCommandLine(esn, text, how string) {
	to := gopodBrobotName(esn)
	if esn == gopodTypedAll {
		to = "Brobots"
	}
	gopodChatFeedAddOperator(to, text, gopodOperatorVia(how))
	gopodReporterReadsOperator(text, how, to) // made if new, then heard
}

// gopodStartNotFound (2026-10-10, operator: "start [anything else] = not
// found"): a "start ..." line that is no run command gets session.not_found_text
// with its name. True when it took the line.
// 2026-10-11 (live: "the park not found." was heard before "start the park" -
// its voice file was still being made): readOut reads the line out first, then
// the not-found line is said, same order as "stop run".
func gopodStartNotFound(text string, readOut func()) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if !strings.HasPrefix(lower, "start ") || gopodTypedCommandMatch(text) {
		return false
	}
	name := strings.TrimSpace(strings.TrimRight(text[len("start "):], ".!?"))
	go func() {
		readOut()
		logger.Println("GOPOD_SESSION start not found: " + name)
		gopodSessionSay(strings.ReplaceAll(gopodBrobotsSettings.Session.NotFoundText, "{run}", name))
	}()
	return true
}

// gopodTypedCommandMatch: text is a run command gopodRunTypedCommand would run.
func gopodTypedCommandMatch(text string) bool {
	if !vars.CustomIntentsExist {
		return false
	}
	lower := strings.ToLower(text)
	for _, c := range vars.CustomIntents {
		if c.IsSystemIntent || c.LuaScript != "" || c.Exec == "" {
			continue
		}
		for _, u := range c.Utterances {
			if seek := strings.ToLower(strings.TrimSpace(u)); seek != "" && strings.Contains(lower, seek) {
				return true
			}
		}
	}
	return false
}

func gopodRunTypedCommand(esn, text string) bool {
	if !vars.CustomIntentsExist {
		return false
	}
	lower := strings.ToLower(text)
	for _, c := range vars.CustomIntents {
		if c.IsSystemIntent || c.LuaScript != "" || c.Exec == "" {
			continue
		}
		for _, u := range c.Utterances {
			seek := strings.ToLower(strings.TrimSpace(u))
			if seek == "" || !strings.Contains(lower, seek) {
				continue
			}
			var args []string
			for _, arg := range c.ExecArgs {
				switch arg {
				case "":
					continue
				case "!botSerial":
					arg = esn
				case "!speechText":
					arg = "\"" + lower + "\""
				case "!intentName":
					arg = c.Name
				case "!locale":
					arg = vars.APIConfig.STT.Language
				}
				args = append(args, arg)
			}
			logger.Println("Bot " + esn + " Custom Intent Matched (typed): " + c.Name)
			logger.Println("Bot " + esn + " Executing: " + c.Exec + " " + strings.Join(args, " "))
			cmd := gopodAsOperator(exec.Command(c.Exec, args...))
			var out bytes.Buffer
			cmd.Stdout, cmd.Stderr = &out, &out
			go func() {
				if err := cmd.Run(); err != nil {
					logger.Println("Bot " + esn + " Custom Intent Exec error: " + err.Error())
				}
				logger.Println("Bot " + esn + " Custom Intent Exec Output: " + strings.TrimSpace(out.String()))
			}()
			return true
		}
	}
	return false
}

// ## EOF

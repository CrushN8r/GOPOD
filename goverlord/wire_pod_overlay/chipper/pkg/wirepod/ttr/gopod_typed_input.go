// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_typed_input.go
package wirepod_ttr

import (
	"bytes"
	"errors"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// gopodTypedAll is the serial the page sends for its "All Brobots" button: both Brobots get the
// line, as when both hear a spoken one with no name - the addressing rule picks who answers (Brobot
// 1 first, his brother defers).
const gopodTypedAll = "all"

// GopodTypedInput handles one typed line for the robot with this ESN.
func GopodTypedInput(esn, text string) error {
	if gopodNeverIn(text, "the keyboard") {
		return nil
	}
	defer gopodTurnOwn(0)()
	if strings.EqualFold(strings.TrimSpace(esn), gopodTypedSystem) {
		return gopodGoverlordCall(text, "Typed text")
	}
	if key, ok := gopodTypedKey(esn); ok && key == "KP0" {
		if fixed := gopodSystemTypo(strings.TrimSpace(text)); fixed != "" && gopodSystemCommandMatch(fixed) {
			text = fixed
		}
	}
	if key, ok := gopodTypedKey(esn); ok && key == "KP0" && gopodSystemCommandMatch(strings.TrimSpace(text)) {
		if _, target, err := gopodNumpadBrobot(key); err == nil {
			defer gopodCommandFrom(target)()
		}
		return gopodTypedToSystem(text, "Typed text")
	}
	if gopodSystemSetMatch(strings.TrimSpace(text)) || strings.HasPrefix(strings.ToLower(strings.TrimSpace(text)), "system-") {
		gopodShowCommandLine(gopodTypedTo(esn), text, "Typed text")
		GopodSystemSayHeard(gopodSystemText("typed_system.dial_zero"), 0)
		return nil
	}
	if key, ok := gopodTypedKey(esn); ok {
		return gopodTypedToKey(key, text, "Typed text")
	}
	return gopodOperatorLine(esn, text, "Typed text")
}

// gopodTypedTo: who a typed line is shown to - a Brobot, or All (a persona key's line shows as to
// All).
func gopodTypedTo(esn string) string {
	if _, ok := gopodTypedKey(esn); ok {
		return gopodTypedAll
	}
	return esn
}

// gopodTypedSystem: the page's 0 - a line to the system itself.
const gopodTypedSystem = "system"

// gopodSystemCommandMatch: a line the system runs, whoever it's typed to - system-..., stop run,
// start <run>, stop session, replay <session>.
func gopodSystemCommandMatch(text string) bool {
	return gopodSystemSetMatch(text) || gopodScriptCommandMatch(text) || gopodStopRunMatch(text) ||
		gopodStopSessionMatch(text) || gopodReplayMatch(text)
}

// GOLDEN Goverlord on -0: a line to -0 (typed, or KP0 held with NumLock off) runs as a line to the
// system, and the system's answer to it is said by GOVERLORD (his own voice host,
// reporter_voices.hosts) - its first line, in his deep voice, his own bubble.
const gopodGoverlordHost = "GOVERLORD"

var gopodGoverlordCalledAt atomic.Int64 // unix ms of the last -0 line

// gopodGoverlordCall: a line to Goverlord (-0).
func gopodGoverlordCall(text, how string) error {
	gopodGoverlordCalledAt.Store(time.Now().UnixMilli())
	if t := gopodSystemTypo(strings.TrimSpace(text)); t != "" && !gopodSystemCommandMatch(t) {
		if _, ok := gopodBrobotsSettings.Managers.Personas[gopodGoverlordHost]; ok {
			gopodGoverlordCalledAt.Store(0)
			return gopodKeyLine(gopodPersonaMark+gopodGoverlordHost, text, how)
		}
	}
	return gopodTypedToSystem(text, how)
}

// gopodGoverlordAnswer: the system's answer to a -0 line, said by Goverlord - once, and only within
// a few seconds of the call.
func gopodGoverlordAnswer(answer string) bool {
	at := gopodGoverlordCalledAt.Load()
	if at == 0 || time.Since(time.UnixMilli(at)) > 15*time.Second || !gopodGoverlordCalledAt.CompareAndSwap(at, 0) {
		return false
	}
	first := strings.TrimSpace(strings.SplitN(strings.TrimSpace(answer), "\n", 2)[0])
	first = strings.Trim(first, "*# ")
	if first == "" {
		return false
	}
	gopodReporterSay(gopodGoverlordHost, first, true, true)
	return true
}

// gopodSystemTypo: "sysem-comm-check" read as "system-comm-check" - a first word within two letters
// of "system", before its "-", is "system".
func gopodSystemTypo(text string) string {
	head, rest, ok := strings.Cut(strings.TrimSpace(text), "-")
	if !ok || strings.EqualFold(head, "system") || len(head) < 4 || gopodEditDistance(strings.ToLower(head), "system") > 2 {
		return text
	}
	return "system-" + rest
}

// gopodTypedToSystem: a line typed to 0 - a system command runs; anything else gets the system's
// one-line pointer (system_text.typed_system.not_a_command).
func gopodTypedToSystem(text, how string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return errors.New("empty text")
	}
	if fixed := gopodSystemTypo(text); fixed != text && gopodSystemCommandMatch(fixed) {
		logger.Println("GOPOD_TYPED typo read as: " + text + " -> " + fixed)
		text = fixed
	}
	if gopodSystemCommandMatch(text) {
		return gopodOperatorLine(gopodTypedAll, text, how)
	}
	gopodOperatorShow(gopodBrobotsSettings.SignOff.SystemRole, text, how)
	GopodChatFeedAdd(gopodSystemSetRole(), gopodSystemText("typed_system.not_a_command"))
	return nil
}

// GopodVisitorTypedInput: an approved visitor's line, under their username.
func GopodVisitorTypedInput(esn, text, how string) error {
	if gopodNeverIn(text, "a visitor") {
		return nil
	}
	if strings.EqualFold(strings.TrimSpace(esn), gopodTypedSystem) {
		return errors.New("the system is the admin's")
	}
	if key, ok := gopodTypedKey(esn); ok {
		return gopodTypedToKey(key, text, how)
	}
	return gopodOperatorLine(esn, text, how)
}

// gopodTypedKey: the page's "kp:N" (its 1-9 menu, 3 to 9) as the numpad key "KPN".
func gopodTypedKey(esn string) (string, bool) {
	n := strings.TrimPrefix(strings.TrimSpace(esn), "kp:")
	if n == esn {
		return "", false
	}
	if n == "dot" {
		return gopodNumpadAnyKey, true
	}
	num, _ := strings.CutSuffix(n, gopodNumlockOffSuffix)
	if len(num) != 1 || num < "0" || num > "9" {
		return "", false
	}
	return "KP" + n, true
}

// gopodTypedToKey: a typed line to a numpad key's persona, as holding the key and speaking does
// (GopodNumpadPTTStop) - a persona answers in its own bubble; a Brobot's key goes to that Brobot.
func gopodTypedToKey(key, text, how string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return errors.New("empty text")
	}
	if gopodSystemCommandMatch(text) {
		if _, target, err := gopodNumpadBrobot(key); err == nil {
			defer gopodCommandFrom(target)()
		}
		return gopodOperatorLine(gopodTypedAll, text, how)
	}
	_, target, err := gopodNumpadBrobot(key)
	if err != nil {
		return err
	}
	return gopodKeyLine(target, gopodCapsTypoFix(gopodSpellNames(text)), how)
}

// GopodVoiceCommandInput: a line a voice command's script hands over (start interview, stop run,
// GOPOD Yourself) - the robot heard it, so the page already shows it as the operator's line and no
// second one is added.
func GopodVoiceCommandInput(esn, text string) error {
	return gopodOperatorLine(esn, text, gopodVoiceCommandHow)
}

const gopodVoiceCommandHow = "Voice command"

// gopodCapsTypoRE: a word typed small then capitals ("iS", "tHE").
var gopodCapsTypoRE = regexp.MustCompile(`\b\p{Ll}\p{Lu}+\b`)

// gopodCapsTypoFix: a word typed small-then-capitals is made small ("is"); "iPhone" and "eBay" stay
// as typed.
func gopodCapsTypoFix(text string) string {
	return gopodCapsTypoRE.ReplaceAllStringFunc(text, strings.ToLower)
}

// gopodOperatorLine: a typed line, or the words of a numpad push-to-talk capture (how names which
// in the log), for one Brobot or "all".
func gopodOperatorLine(esn, text, how string) error {
	text = gopodCapsTypoFix(strings.TrimSpace(text))
	if text == "" {
		return errors.New("empty text")
	}
	gopodPictureForget()
	if gopodSystemSetMatch(text) {
		gopodSystemSetLine(esn, text, how)
		return nil
	}
	if gopodStopRunMatch(text) {
		songRun := gopodRunActive() || gopodScriptPlaying.Load()
		if gopodStopRuns("stop run") {
			defer gopodTurnOwn(0)()
			gopodShowCommandLine(esn, text, how)
			if !songRun {
				gopodStoppedSay()
			}
			return nil
		}
		gopodShowCommandLine(esn, text, how)
		gopodStopRunCommand(text)
		return nil
	}
	if gopodWarmupHold(esn, text, how) {
		return nil
	}
	if GopodRecordSessionAnswer(text, how) {
		return nil
	}
	if gopodScriptCommandMatch(text) {
		gopodShowCommandLine(esn, text, how)
		gopodScriptCommand(text)
		return nil
	}
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
	if gopodWarmGateDrop(esn, how, text) {
		return nil
	}
	if !gopodIsThanks(text) && gopodRunGateDrop(esn, how, text) {
		return nil
	}
	if gopodThanksCloseOut(esn, text, how) {
		return nil
	}
	if gopodYourselfCommand(esn, text, how) {
		return nil
	}
	if esn == gopodTypedAll && gopodSolo() {
		esn = gopodBrobot1ESN
	}
	answerers := []string{gopodBrobotName(esn)}
	if esn == gopodTypedAll {
		answerers = []string{"Brobot 1", "Brobot 2"}
	}
	return gopodLineTurn(text, answerers, func() error { return gopodOperatorAsk(esn, text, how) })
}

// gopodOperatorAsk: a question to one Brobot or "all", in its turn.
func gopodOperatorAsk(esn, text, how string) error {
	gopodRecordOperatorInput(how)
	if esn == gopodTypedAll {
		return gopodTypedInputAll(text, how)
	}
	if gopodBrobotName(esn) == "Brobot" {
		return errors.New("not a Brobot: " + esn)
	}
	logger.Println("Bot " + esn + " " + how + ": " + text)
	gopodShowChapter(gopodSpellNames(text))
	gopodOperatorShow(gopodBrobotName(esn), text, how)
	run := gopodRunAllowed(text)
	if run && gopodTypedCommandMatch(text) {
		go func() {
			gopodOperatorRead(gopodBrobotName(esn), text, how)
			gopodRunTypedCommand(esn, text)
		}()
		return nil
	}
	if run && gopodStartNotFound(text, func() { gopodOperatorRead(gopodBrobotName(esn), text, how) }) {
		return nil
	}
	if gopodRunNativeCube(esn, text) {
		return nil
	}
	go func() {
		gopodOperatorRead(gopodBrobotName(esn), text, how)
		gopodShowTopicPictures(esn, text)
		if _, err := StreamingKGSim(nil, esn, text, false); err != nil {
			logger.Println("LLM error (typed): " + err.Error())
		}
	}()
	return nil
}

// gopodTypedAll's gap between the two Brobots' requests, so Brobot 1 is the one heard first.
const gopodTypedAllGap = 100 * time.Millisecond

func gopodTypedInputAll(text, how string) error {
	logger.Println("Bot " + gopodBrobot1ESN + " " + how + " (all): " + text)
	gopodShowChapter(gopodSpellNames(text))
	to := gopodLineTo("Brobots")
	gopodOperatorShow(to, text, how)
	run := gopodRunAllowed(text)
	if run && gopodTypedCommandMatch(text) {
		go func() {
			gopodOperatorRead(to, text, how)
			gopodRunTypedCommand(gopodBrobot1ESN, text)
		}()
		return nil
	}
	if run && gopodStartNotFound(text, func() { gopodOperatorRead(to, text, how) }) {
		return nil
	}
	if run && gopodRunTypedCommand(gopodBrobot1ESN, text) {
		return nil
	}
	if gopodNativeCubeIntent(text) != "" {
		gopodRunNativeCube(gopodBrobot1ESN, text)
		gopodRunNativeCube(gopodBrobot2ESN, text)
		return nil
	}
	read := make(chan struct{})
	go func() {
		gopodOperatorRead(to, text, how)
		gopodShowTopicPictures(gopodAllFirst(text), text)
		close(read)
	}()
	if _, _, ok := gopodMathParts(text); ok {
		lead := gopodAllFirst(text)
		if lead != gopodBrobot1ESN {
			gopodMathLeadPick.Store(strings.TrimSpace(text), lead)
			gopodAnswerNext.Store(lead, true)
		}
		go func() {
			<-read
			if _, err := StreamingKGSim(nil, lead, text, false); err != nil {
				logger.Println("LLM error (typed): " + err.Error())
			}
		}()
		return nil
	}
	first := gopodAllFirst(text)
	if gopodAddressedTo(text) == "" || gopodAllStartsWith(text) != "" {
		second := gopodBrotherOf(first)
		gopodExchangeOpen(text, second)
		go func() {
			<-read
			done := gopodReplyDone(first)
			failed := make(chan struct{})
			gopodAllFirstAnswer.Store(first, true) // his answer hands over to his brother (kgsim.go)
			gopodDialZero.Store(first, true)       // both asked: the Brobots at the dial's 0%
			if first != gopodBrobot1ESN {
				gopodAnswerNext.Store(first, true)
			}
			go func() {
				if _, err := StreamingKGSim(nil, first, text, false); err != nil {
					logger.Println("LLM error (typed): " + err.Error())
					close(failed)
				}
			}()
		wait:
			for {
				select {
				case <-done:
					break wait
				case <-failed:
					break wait
				case <-time.After(gopodSecondsToDuration(gopodBrobotsSettings.Addressing.AllNextWaitSeconds)):
					if gopodReplyGoing(first) {
						continue
					}
					logger.Println("GOPOD_ALL " + gopodBrobotName(first) + "'s answer didn't end in time")
					break wait
				}
			}
			gopodDialZero.Delete(first)
			logger.Println("GOPOD_ALL " + gopodBrobotName(second) + " answers next: " + text)
			gopodAnswerNext.Store(second, true)
			gopodFollowingUp.Store(second, true)
			gopodDialZero.Store(second, true)
			if _, err := StreamingKGSim(nil, second, text, false); err != nil {
				logger.Println("LLM error (typed): " + err.Error())
			}
			gopodDialZero.Delete(second)
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

// gopodRunTypedCommand runs the first custom voice command whose phrase is in the typed text, the
// same match Wire-Pod uses for a heard one (system and Lua intents are left to voice).
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

// The page's All menu: "2" is the same as "Starting with Brobot 2" for that line - kept by its
// words, so only that line is turned round (typed_line.all_first).
var gopodAllFirstPick = struct {
	sync.Mutex
	text, first, standing string
}{}

// GopodAllFirstPick: the menu's choice for a typed line ("" = the default).
func GopodAllFirstPick(text, first string) {
	first = strings.TrimSpace(first)
	from := "the page's All menu"
	if first == "" {
		first = strings.TrimSpace(gopodBrobotsSettings.TypedLine.AllFirst)
		from = "typed_line.all_first (no pick sent)"
	}
	line := "GOPOD_PICK all-first=" + first + " (" + from + ")"
	if t := strings.TrimSpace(text); t != "" {
		line += " for: " + gopodCut(t, 80)
	}
	logger.Println(line)
	gopodAllFirstPick.Lock()
	gopodAllFirstPick.text, gopodAllFirstPick.first = strings.TrimSpace(text), first
	gopodAllFirstPick.standing = first
	gopodAllFirstPick.Unlock()
}

func gopodAllFirstPicked(text string) string {
	gopodAllFirstPick.Lock()
	defer gopodAllFirstPick.Unlock()
	if gopodAllFirstPick.text != "" && gopodAllFirstPick.text == strings.TrimSpace(text) {
		return gopodAllFirstPick.first
	}
	return gopodAllFirstPick.standing
}

// gopodShowCommandLine: a command line's Operator bubble ("Keyboard to Brobot 1"), read out in the
// background.
func gopodShowCommandLine(esn, text, how string) {
	to := gopodBrobotName(esn)
	if esn == gopodTypedAll {
		to = gopodLineTo("Brobots")
	}
	gopodOperatorShow(to, text, how)
	gopodOperatorRead(to, text, how) // made if new, then heard
}

// gopodStartNotFound: a "start ..." line that is no run command gets session.not_found_text with
// its name.
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

// gopodRunFromZero: the next run command came on 0 (KEYS) - the only key that runs one; taken once
// by the line it came with.
var gopodRunFromZero atomic.Bool

// gopodRunCommandLine: a run command or a "start ..." line.
func gopodRunCommandLine(text string) bool {
	return gopodTypedCommandMatch(text) || strings.HasPrefix(strings.ToLower(strings.TrimSpace(text)), "start ")
}

// gopodRunAllowed: this line's run command came on 0; taken once.
func gopodRunAllowed(text string) bool {
	if !gopodRunCommandLine(text) {
		return false
	}
	if gopodRunFromZero.CompareAndSwap(true, false) {
		return true
	}
	logger.Println("GOPOD_TYPED run command off 0 - talk: " + text)
	return false
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

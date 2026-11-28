// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_dizzy.go
package wirepod_ttr

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// "Are you (feeling) dizzy?" - one mechanism, one call (gopodDizzy) for every persona asked. Its one
// source is the battery watch (gopodBatteryHow, gopodBatteryDizzyWords, gopodSweepLine); its words
// are in dizzy-brobots.md.

// gopodDizzyAsk: the line is the catchphrase (battery_watch.ask_words).
func gopodDizzyAsk(message string) bool {
	lower := strings.ToLower(message)
	for _, w := range gopodBrobotsSettings.BatteryWatch.AskWords {
		if gopodHasWholeWords(lower, strings.ToLower(strings.TrimSpace(w))) {
			return true
		}
	}
	return false
}

type gopodDizzyAnswer struct{ line, note string }

var gopodDizzyMade sync.Map // who + question -> gopodDizzyAnswer: made once an answer, so every part says the same

// gopodDizzy: the one call - asked the catchphrase, what the code says first (line) and what the
// model is told (note). A Brobot: his power level and how he is (ask_line, ask_note). His guard: the
// security sweep's words for him (ask_note_guard). Anyone else: no line - the batteries are the
// guards' job, so he hands it to them (ask_note_others). Not asked: "", "".
func gopodDizzy(who, question string) (line, note string) {
	if !gopodDizzyAsk(question) {
		return "", ""
	}
	key := who + "\x00" + question
	if v, ok := gopodDizzyMade.Load(key); ok {
		a := v.(gopodDizzyAnswer)
		return a.line, a.note
	}
	a := gopodDizzyMake(who)
	if v, loaded := gopodDizzyMade.LoadOrStore(key, a); loaded {
		a = v.(gopodDizzyAnswer)
	} else {
		time.AfterFunc(2*time.Minute, func() { gopodDizzyMade.Delete(key) })
	}
	return a.line, a.note
}

func gopodDizzyMake(who string) gopodDizzyAnswer {
	bw := gopodBrobotsSettings.BatteryWatch
	for _, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN} {
		name := gopodBrobotName(esn)
		if name == who {
			volts, state, onCharger := gopodBatteryHow(esn)
			how := bw.AskStates[state]
			if state == "low" || state == "dip" {
				how, state = gopodBatteryDizzyWords(name, onCharger), "dizzy"
			}
			line := strings.TrimSpace(how)
			if state != "missing" && strings.TrimSpace(bw.AskLine) != "" {
				line = strings.Join(strings.Fields(strings.NewReplacer("{volts}", fmt.Sprintf("%.2f", volts), "{state}", how).Replace(bw.AskLine)), " ")
			}
			return gopodDizzyAnswer{line, strings.ReplaceAll(bw.AskNote, "{state}", state)}
		}
		if g := strings.TrimSpace(bw.Guards[name]); g != "" && strings.EqualFold(g, strings.TrimSpace(who)) {
			report, _ := gopodSweepLine(esn)
			return gopodDizzyAnswer{report, strings.ReplaceAll(bw.AskNoteGuard, "{report}", report)}
		}
	}
	var guards []string
	for _, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN} {
		if g := bw.Guards[gopodBrobotName(esn)]; g != "" {
			guards = append(guards, gopodPersonaFill("", "{"+g+".call}")+" ("+gopodBrobotName(esn)+")")
		}
	}
	return gopodDizzyAnswer{"", strings.ReplaceAll(bw.AskNoteOthers, "{guards}", strings.Join(guards, ", "))}
}

// ## EOF

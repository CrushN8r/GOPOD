// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_comm_check.go
package wirepod_ttr

// The comm check (2026-10-28, operator: "Simulate the whole KPN 1-9 range to
// record those bubbles as well * same test question, Hey party animal. Who are
// you and what do you do?"; "make the 'who are you?' profile development part of
// this run as well * like a communication check with all chat participants,
// before the show starts * at bare minimum, do they all know who they are & each
// other? * who they are: do they all know the diff between tell me about you &
// tell me all about you * & each other: if I ask CHALK/PLAYHEAD tell me about
// their interns, if I ask any chat participant about any other chat participant >
// Hey [chat participant], who all else is here with us?"). system-comm-check plays
// comm_check's rounds - each line on each of its keys in turn, as if the key were
// held and the words spoken: the page draws the key's bubbles (numpad_sim), the
// line goes where the key goes (the same as GopodNumpadPTTStop), live - the
// model, the voices, the robots - and the next waits for the answer to finish
// (the feed quiet, the page caught up). The game goes back to 0 after: the show
// starts clean. Record Session keeps it all.

import (
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

var gopodCommCheckRunning atomic.Bool

// gopodFeedLastID: the newest feed line's id.
func gopodFeedLastID() int {
	gopodChatFeed.Lock()
	defer gopodChatFeed.Unlock()
	if n := len(gopodChatFeed.events); n > 0 {
		return gopodChatFeed.events[n-1].ID
	}
	return 0
}

// gopodPagePlayedUpTo: how far the page has played this boot's feed.
func gopodPagePlayedUpTo() int {
	gopodChatFeed.Lock()
	defer gopodChatFeed.Unlock()
	return gopodPagePlayedTo
}

// gopodCommCheckWaitPage: until the page has played id (or max).
func gopodCommCheckWaitPage(id int, max time.Duration) {
	for end := time.Now().Add(max); time.Now().Before(end); time.Sleep(250 * time.Millisecond) {
		if gopodPagePlayedUpTo() >= id {
			return
		}
	}
}

// gopodCommCheckWaitQuiet: until the feed has had no new line for quiet and the
// page has played it all (or max).
func gopodCommCheckWaitQuiet(quiet, max time.Duration) {
	last, since := gopodFeedLastID(), time.Now()
	for end := time.Now().Add(max); time.Now().Before(end); time.Sleep(500 * time.Millisecond) {
		if id := gopodFeedLastID(); id != last {
			last, since = id, time.Now()
			continue
		}
		if time.Since(since) >= quiet && gopodPagePlayedUpTo() >= last {
			return
		}
	}
}

// gopodCommCheckLines: every round's text on each of its keys, in order.
func gopodCommCheckLines() [][2]string {
	cc := gopodBrobotsSettings.CommCheck
	var lines [][2]string // key, text
	for _, r := range cc.Rounds {
		keys := r.Keys
		if len(keys) == 0 {
			keys = cc.Keys
		}
		for _, k := range keys {
			lines = append(lines, [2]string{k, r.Text})
		}
	}
	return lines
}

// gopodCommCheckWaitAnswer: until a feed line after from comes from who
// (his answer), or max.
func gopodCommCheckWaitAnswer(from int, who string, max time.Duration) {
	for end := time.Now().Add(max); time.Now().Before(end); time.Sleep(500 * time.Millisecond) {
		gopodChatFeed.Lock()
		found := false
		for i := len(gopodChatFeed.events) - 1; i >= 0 && gopodChatFeed.events[i].ID > from; i-- {
			if strings.EqualFold(gopodChatFeed.events[i].Who, who) {
				found = true
				break
			}
		}
		gopodChatFeed.Unlock()
		if found {
			return
		}
	}
	logger.Println("GOPOD_COMM_CHECK no answer from " + who + " in time")
}

// gopodSystemCommCheck starts the comm check; the answer says it started.
func gopodSystemCommCheck() string {
	cc := gopodBrobotsSettings.CommCheck
	if len(cc.Rounds) == 0 {
		return "system-comm-check: comm_check.rounds is empty"
	}
	if gopodRunActive() || gopodScriptPlaying.Load() {
		return "system-comm-check: a run is going - try again after it"
	}
	if !gopodCommCheckRunning.CompareAndSwap(false, true) {
		return "system-comm-check: already running"
	}
	lines := gopodCommCheckLines()
	go func() {
		defer gopodCommCheckRunning.Store(false)
		quiet := gopodSecondsToDuration(cc.QuietSeconds)
		max := gopodSecondsToDuration(cc.LineMaxSeconds)
		for i, l := range lines {
			if gopodRunActive() {
				logger.Println("GOPOD_COMM_CHECK stopped: a run started")
				return
			}
			key := strings.ToUpper(strings.TrimSpace(l[0]))
			label, target, err := gopodNumpadBrobot(key)
			if err != nil {
				logger.Println("GOPOD_COMM_CHECK " + key + ": " + err.Error())
				continue
			}
			name := label
			if strings.HasPrefix(target, gopodPersonaMark) {
				name = gopodCallName(strings.TrimPrefix(target, gopodPersonaMark))
			} else if h := strings.TrimSpace(gopodBrobotsSettings.HostPersonas[label].Name); h != "" {
				name = h
			}
			text := gopodSpellNames(strings.ReplaceAll(l[1], "{name}", name))
			logger.Println("GOPOD_COMM_CHECK " + strconv.Itoa(i+1) + "/" + strconv.Itoa(len(lines)) + " " + key + " -> " + label + ": " + text)
			// The page draws the key's bubbles, as if it were held.
			GopodChatFeedAdd(gopodBrobotsSettings.ReporterVoices.Role, "numpad_sim|"+key+"|"+label)
			gopodCommCheckWaitPage(gopodFeedLastID(), 30*time.Second)
			gopodNumpadLastKey.Store(key)
			// 2026-10-29 (live 05:13: KP2 went before Brobot 1 had answered, and
			// the replies landed among KP3's bubbles): the next key waits for
			// this one's own answer - a line from who was asked - then the quiet.
			from := gopodFeedLastID()
			answerer := label
			if strings.HasPrefix(target, gopodPersonaMark) {
				answerer = strings.TrimPrefix(target, gopodPersonaMark)
			}
			if strings.HasPrefix(target, gopodPersonaMark) {
				host := strings.TrimPrefix(target, gopodPersonaMark)
				gopodLineTurn(text, []string{host}, func() error { gopodPersonaReply(host, text, "Numpad text"); return nil })
			} else if err := gopodOperatorLine(target, text, "Numpad text"); err != nil {
				logger.Println("GOPOD_COMM_CHECK " + err.Error())
			}
			gopodCommCheckWaitAnswer(from, answerer, max)
			gopodCommCheckWaitQuiet(quiet, max)
		}
		gopodGameReset("comm check")
		logger.Println("GOPOD_COMM_CHECK done: " + strconv.Itoa(len(lines)) + " lines")
		if t := strings.TrimSpace(cc.DoneText); t != "" {
			if role := gopodBrobotsSettings.SignOff.SystemRole; role != "" {
				gopodReporterSay(role, strings.ReplaceAll(t, "{n}", strconv.Itoa(len(lines))), true, false)
			}
		}
	}()
	return strings.ReplaceAll(cc.StartText, "{n}", strconv.Itoa(len(lines)))
}

// ## EOF

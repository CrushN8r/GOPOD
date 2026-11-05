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
	"context"
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

var gopodCommCheckRunning, gopodCommCheckStop atomic.Bool

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
	for end := time.Now().Add(max); time.Now().Before(end) && !gopodCommCheckStop.Load(); time.Sleep(250 * time.Millisecond) {
		if gopodPagePlayedUpTo() >= id {
			return
		}
	}
}

// gopodCommCheckWaitQuiet: until the feed has had no new line for quiet and the
// page has played it all (or max).
func gopodCommCheckWaitQuiet(quiet, max time.Duration) {
	last, since := gopodFeedLastID(), time.Now()
	for end := time.Now().Add(max); time.Now().Before(end) && !gopodCommCheckStop.Load(); time.Sleep(500 * time.Millisecond) {
		if id := gopodFeedLastID(); id != last {
			last, since = id, time.Now()
			continue
		}
		if time.Since(since) >= quiet && gopodPagePlayedUpTo() >= last {
			return
		}
	}
}

// gopodCommCheckLines: every round's text on each of its keys, in order, with
// who asks it.
func gopodCommCheckLines() [][3]string {
	cc := gopodBrobotsSettings.CommCheck
	var lines [][3]string // key, text, asker
	for _, r := range cc.Rounds {
		keys := r.Keys
		if len(keys) == 0 {
			keys = cc.Keys
		}
		for _, k := range keys {
			lines = append(lines, [3]string{k, r.Text, r.Asker})
		}
	}
	return lines
}

// gopodCommCheckAsking (2026-11-04, operator: "Goverlord, to orchestrate,
// delegates butler Keys to get the party started * Keys will now do the round
// robin"): the comm check line going out and who asks it - its bubble is his
// (gopodScriptQuestionReader) and it's read in his voice
// (gopodReporterReadsOperator), not the Operator's.
var gopodCommCheckAsking atomic.Value // [2]string{text, asker}

// gopodCommCheckAsker: who asks this comm check line, "" when it isn't one.
func gopodCommCheckAsker(text string) string {
	a, _ := gopodCommCheckAsking.Load().([2]string)
	if a[1] == "" || strings.TrimSpace(a[0]) != strings.TrimSpace(text) {
		return ""
	}
	return a[1]
}

// gopodCommCheckWaitAnswer: until a feed line after from comes from who
// (his answer), or max.
func gopodCommCheckWaitAnswer(from int, who string, max time.Duration) {
	for end := time.Now().Add(max); time.Now().Before(end) && !gopodCommCheckStop.Load(); time.Sleep(500 * time.Millisecond) {
		gopodChatFeed.Lock()
		found := false
		for i := len(gopodChatFeed.events) - 1; i >= 0 && gopodChatFeed.events[i].ID > from; i-- {
			if strings.EqualFold(gopodChatFeed.events[i].Who, who) {
				found = true
				break
			}
		}
		gopodChatFeed.Unlock()
		// 2026-10-30 (live 15:48:04: KP8 started while CHALK's voice was still
		// being made - his line was on the feed, his answer wasn't over): his
		// line on the feed, and his answer ended - made, said, heard.
		if found && !gopodAnswerGoing([]string{who}) {
			return
		}
	}
	logger.Println("GOPOD_COMM_CHECK no answer from " + who + " in time")
}

// gopodSystemCommCheck starts the comm check; the answer says it started.
func gopodSystemCommCheck() string {
	cc := gopodBrobotsSettings.CommCheck
	if len(cc.Rounds) == 0 {
		return gopodSystemText("comm_check.system_comm_check_comm_check")
	}
	if gopodRunActive() || gopodScriptPlaying.Load() {
		return gopodSystemText("comm_check.system_comm_check_a_run")
	}
	if !gopodCommCheckRunning.CompareAndSwap(false, true) {
		return gopodSystemText("comm_check.system_comm_check_already_running")
	}
	lines := gopodCommCheckLines()
	go func() {
		defer gopodCommCheckRunning.Store(false)
		// 2026-10-29: the comm check is a turn of the master chat queue; its KP1-9
		// questions are its own nested lines (gopod_chat_queue.go).
		defer gopodTurnBeginOwn("comm_check", false)()
		quiet := gopodSecondsToDuration(cc.QuietSeconds)
		max := gopodSecondsToDuration(cc.LineMaxSeconds)
		gopodCommCheckStop.Store(false)
		// 2026-11-02 (live 16:38-16:41: KP0's answer was over, but the check sat
		// still - the voice library was still being made at start, and every
		// answer counts as going while any voice is being made): it starts once
		// the library is ready (at most ten minutes).
		if gopodVoicesAtStart.Load() {
			logger.Println("GOPOD_COMM_CHECK waiting for the voice library to be ready")
			for end := time.Now().Add(10 * time.Minute); gopodVoicesAtStart.Load() && time.Now().Before(end) && !gopodCommCheckStop.Load(); time.Sleep(time.Second) {
			}
			logger.Println("GOPOD_COMM_CHECK voice library ready - starting")
		}
		// 2026-11-04 (operator: "Add in wake calls to new songs"): the song's
		// wake_both - each Brobot woken (control taken, the settle, given back).
		if cc.Wake {
			for _, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN} {
				if strings.TrimSpace(esn) != "" {
					gopodRobotCall(esn, true, func(*vector.Vector, context.Context) string { return "" })
				}
			}
		}
		// 2026-11-04 (operator: "remove who's here in the chat from every
		// response, too repetitive"): no who's-here line before the first key -
		// Goverlord hands it to Keys, and Keys calls everyone in.
		defer gopodCommCheckAsking.Store([2]string{})
		defer gopodGameWelcomePending.Store(false)
		for i, l := range lines {
			if gopodCommCheckStop.Load() {
				logger.Println("GOPOD_COMM_CHECK stopped (stop run)")
				return
			}
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
			// 2026-11-02 (live: "Hey All! You party animal!", and Pip said "Hey All!"
			// back): the line to both is to the Brobots (comm_check.all_name).
			if n := strings.TrimSpace(cc.AllName); n != "" && strings.EqualFold(label, "All") {
				name = n
			}
			text := gopodSpellNames(strings.ReplaceAll(l[1], "{name}", name))
			logger.Println("GOPOD_COMM_CHECK " + strconv.Itoa(i+1) + "/" + strconv.Itoa(len(lines)) + " " + key + " -> " + label + ": " + text)
			// The page draws the key's bubbles, as if it were held.
			GopodChatFeedAdd(gopodBrobotsSettings.ReporterVoices.Role, "numpad_sim|"+key+"|"+label)
			gopodCommCheckWaitPage(gopodFeedLastID(), 30*time.Second)
			gopodNumpadLastKey.Store(key)
			gopodCommCheckAsking.Store([2]string{text, strings.TrimSpace(l[2])})
			// 2026-11-04: the last line's answer ends with the game welcome.
			if i == len(lines)-1 {
				gopodGameWelcomePending.Store(true)
			}
			// 2026-10-29 (live 05:13: KP2 went before Brobot 1 had answered, and
			// the replies landed among KP3's bubbles): the next key waits for
			// this one's own answer - a line from who was asked - then the quiet.
			from := gopodFeedLastID()
			answerer := label
			if strings.HasPrefix(target, gopodPersonaMark) {
				answerer = strings.TrimPrefix(target, gopodPersonaMark)
			}
			// 2026-10-30 (the full numpad: * held is All): All's line is answered
			// by both Brobots, Brobot 1 then Brobot 2 - it ends with Brobot 2's
			// (solo, Brobot 1's).
			if target == gopodTypedAll {
				answerer = "Brobot 2"
				if gopodSolo() {
					answerer = "Brobot 1"
				}
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

// gopodGameWelcomePending (2026-11-04, operator: "GOPOD game = 1st reminder,
// however coloured/flavoured"; "brobot 2 ends the song * point robust
// per-persona flavour mechanism at brobot 2"): the comm song's last line is
// out - its answer by comm_check.game_welcome.by ends with the game welcome.
var gopodGameWelcomePending atomic.Bool

// gopodGameWelcomeLine: the game welcome for this Brobot's reminder, once -
// the base word for word or one of his flavours, at random; "" otherwise.
func gopodGameWelcomeLine(esn string) string {
	gw := gopodBrobotsSettings.CommCheck.GameWelcome
	name := gopodBrobotName(esn)
	if !gopodGameWelcomePending.Load() || !strings.EqualFold(strings.TrimSpace(gw.By), name) {
		return ""
	}
	var pool []string
	for _, line := range append([]string{gw.Base}, gw.Flavours[name]...) {
		if line = strings.TrimSpace(line); line != "" {
			pool = append(pool, line)
		}
	}
	if len(pool) == 0 || !gopodGameWelcomePending.CompareAndSwap(true, false) {
		return ""
	}
	return pool[rand.Intn(len(pool))]
}

// gopodCommCheckNotes: the comm check song's one note - a line asked on its keys.
var gopodCommCheckNotes = map[string]bool{"ask": true, "wake_both": true}

// gopodApplyCommCheckSong (2026-11-04, operator: "Make system-comm-check a GOPOD
// song"): the comm check's rounds come from its song (comm_check.song, a song_id:
// runtime/songs/05_comm_check) - each "ask" step's line from story.md, asked on
// each of its keys (knobs.json) in order.
func gopodApplyCommCheckSong(s *gopodBrobotsSettingsFile) {
	id := strings.TrimSpace(s.CommCheck.Song)
	if id == "" {
		return
	}
	songs, _ := gopodAllSongs()
	dir := songs[id]
	if dir == "" {
		logger.Println("GOPOD_COMM_CHECK no song folder has song_id " + id)
		return
	}
	story, err := os.ReadFile(filepath.Join(dir, "story.md"))
	if err != nil {
		return
	}
	knobs, err := os.ReadFile(filepath.Join(dir, "knobs.json"))
	if err != nil {
		return
	}
	texts, err := gopodInterviewStory(string(story))
	var k struct {
		Steps []struct {
			StepID string   `json:"step_id"`
			Note   string   `json:"note"`
			Keys   []string `json:"keys"`
			Asker  string   `json:"asker"`
		} `json:"steps"`
	}
	if err != nil || json.Unmarshal(knobs, &k) != nil {
		logger.Println("GOPOD_COMM_CHECK song " + filepath.Base(dir) + " unreadable")
		return
	}
	s.CommCheck.Rounds = nil
	for _, st := range k.Steps {
		if st.Note == "wake_both" {
			s.CommCheck.Wake = true
		}
		if st.Note == "ask" && strings.TrimSpace(texts[st.StepID]) != "" {
			s.CommCheck.Rounds = append(s.CommCheck.Rounds, struct {
				Text  string   `json:"text"`
				Keys  []string `json:"keys"`
				Asker string   `json:"asker"`
			}{texts[st.StepID], st.Keys, st.Asker})
		}
	}
}

// ## EOF

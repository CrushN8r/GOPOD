// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_chat_feed.go
package wirepod_ttr

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// gopodChatFeedMax: the feed keeps its newest page_sync.feed_max lines (300 when unset).
func gopodChatFeedMax() int {
	if n := gopodBrobotsSettings.PageSync.FeedMax; n > 0 {
		return n
	}
	return 300
}

var gopodChatFeedPath = gopodPathWithEnvOverride("GOPOD_CHAT_FEED_PATH", "webroot/gopod_chat_feed.json")

type gopodChatFeedEvent struct {
	ID     int    `json:"id"`
	Who    string `json:"who"`
	To     string `json:"to,omitempty"`
	By     string `json:"by,omitempty"`  // a draft question: the reporter who reads it (the page's label)
	Via    string `json:"via,omitempty"` // an operator line's input: "Keyboard", "KP1", "Backpack" (the page's label)
	ToAs   string `json:"to_as,omitempty"`
	Text   string `json:"text"`
	TimeMS int64  `json:"time_ms"`
	Turn   int64  `json:"turn,omitempty"`
}

var gopodChatFeed = struct {
	sync.Mutex
	boot   int64
	next   int
	events []gopodChatFeedEvent
}{boot: time.Now().UnixMilli()}

// GopodChatFeedStart writes an empty feed with this Wire-Pod start's stamp.
func GopodChatFeedStart() {
	GopodBrobotsSetup()
	gopodSoundsToPage()
	gopodChatFeed.Lock()
	defer gopodChatFeed.Unlock()
	gopodChatFeed.next, gopodChatFeed.events = 0, nil
	gopodWriteChatFeedLocked()
}

// GopodChatFeedAdd adds one line to the page's feed.
func GopodChatFeedAdd(who, text string) {
	if who == gopodSystemSetRole() && gopodGoverlordAnswer(text) {
		_, rest, more := strings.Cut(strings.TrimSpace(text), "\n")
		if !more || strings.TrimSpace(rest) == "" {
			return
		}
		text = rest
	}
	if r := gopodBrobotsSettings.ReporterVoices.SystemSayRole; r != "" && who == r {
		gopodSystemSayQueued.Add(1)
		go func() {
			defer gopodSystemSayQueued.Done()
			gopodSystemSay(text, false)
		}()
		return
	}
	GopodChatFeedAddTo(who, "", text)
}

// gopodChatFeedAddTurn adds a line with its own turn (0: none - a control line).
func gopodChatFeedAddTurn(turn int64, who, text string) {
	gopodChatFeedAddEventTurn(who, "", "", "", text, turn)
}

// gopodChatFeedAddOperator adds an operator line to a Brobot ("Brobots" for both).
func gopodChatFeedAddOperator(to, text, via string) {
	gopodChatFeedAddEventVia("Operator", to, gopodScriptQuestionReader(to, text), via, text)
}

// gopodOperatorVia: the input an operator line came in by, for its bubble's label.
func gopodOperatorVia(how string) string {
	switch how {
	case "Typed text":
		return "Keyboard"
	case "Numpad text":
		if key, _ := gopodNumpadLastKey.Load().(string); key != "" {
			return key
		}
		return "Numpad"
	case gopodVoiceCommandHow:
		return "Backpack" // heard by the robot himself (operator: "use \"Backpack\"")
	case gopodMicHow:
		return "Mic"
	}
	if strings.HasPrefix(how, gopodVisitorHowPrefix) {
		return strings.TrimPrefix(how, gopodVisitorHowPrefix)
	}
	return ""
}

// gopodFeedLiveVia: a live (LLM) reply's feed line (GopodChatFeedAddLive).
const gopodFeedLiveVia = "llm"

// GopodChatFeedAddLive: a Brobot's live reply goes on the feed, in order with every other line; the
// page shows it as a live bubble.
func GopodChatFeedAddLive(esn, text string) {
	gopodChatFeedAddLiveBy(esn, text, gopodPersonaShown(gopodBrobotName(esn)))
}

// gopodChatFeedAddLiveBy: a Brobot's live bubble headed by - the persona answering, or the body
// itself for what the body says (gopodBodySay).
func gopodChatFeedAddLiveBy(esn, text, by string) {
	who := gopodBrobotName(esn)
	words := strings.Join(strings.Fields(text), " ")
	gopodLiveCheck.Lock()
	if last, ok := gopodLiveCheck.last[who]; ok && last == words && time.Since(gopodLiveCheck.at[who]) < 30*time.Second {
		logger.Println("GOPOD_FEED_CHECK live reply twice in a row from " + who + ": " + gopodHeardShort(words))
	}
	gopodLiveCheck.last[who], gopodLiveCheck.at[who] = words, time.Now()
	gopodLiveCheck.Unlock()
	gopodChatFeedAddEventVia(who, "", by, gopodFeedLiveVia, text)
	gopodChatFeed.Lock()
	id := gopodChatFeed.next
	gopodChatFeed.Unlock()
	gopodLiveCheck.Lock()
	gopodLiveCheck.open[id] = who
	gopodLiveCheck.lastID[esn] = id
	gopodLiveCheck.Unlock()
}

// gopodLiveCheck: each live reply line waits for the page's word on its bubble; one the page has
// played past with no word is logged as missing (GopodPagePlayed).
var gopodLiveCheck = struct {
	sync.Mutex
	last   map[string]string
	at     map[string]time.Time
	open   map[int]string
	lastID map[string]int
}{last: map[string]string{}, at: map[string]time.Time{}, open: map[int]string{}, lastID: map[string]int{}}

// GopodLiveReplyNoted: the page said what became of live reply id's bubble.
func GopodLiveReplyNoted(id int) {
	gopodLiveCheck.Lock()
	delete(gopodLiveCheck.open, id)
	gopodLiveCheck.Unlock()
}

// The page's side of the dialogue: the page says how far it has played the feed (each line shown,
// its voice ended), and the journal gets "GOPOD_PAGE played <id> <who>: <words>" under the engine's
// own "GOPOD_FEED <id> ..." - both sides of the same run, in one list.
var gopodPagePlayedTo int

// GopodPagePlayed: the page (this boot's feed) has played up to id.
func GopodPagePlayed(boot int64, id int) {
	gopodChatFeed.Lock()
	if boot != gopodChatFeed.boot || id <= gopodPagePlayedTo {
		if boot != gopodChatFeed.boot {
			logger.Println("GOPOD_PAGE played " + strconv.Itoa(id) + " not taken: the page has an old feed (boot " + strconv.FormatInt(boot, 10) + ")")
		}
		gopodChatFeed.Unlock()
		return
	}
	var played []gopodChatFeedEvent
	for _, e := range gopodChatFeed.events {
		if e.ID > gopodPagePlayedTo && e.ID <= id {
			played = append(played, e)
		}
	}
	gopodPagePlayedTo = id
	gopodChatFeed.Unlock()
	gopodLiveCheck.Lock()
	for liveID, who := range gopodLiveCheck.open {
		if liveID <= id {
			logger.Println("GOPOD_FEED_CHECK live reply " + strconv.Itoa(liveID) + " from " + who + ": the page played past it with no bubble")
			delete(gopodLiveCheck.open, liveID)
		}
	}
	gopodLiveCheck.Unlock()
	for _, e := range played {
		text := strings.Join(strings.Fields(e.Text), " ")
		if len(text) > 80 {
			text = gopodCut(text, 80)
		}
		logger.Println("GOPOD_PAGE played " + strconv.Itoa(e.ID) + " " + e.Who + ": " + text)
	}
}

// gopodCut: text cut to at most n bytes, never inside a character, then "...".
func gopodCut(text string, n int) string {
	if len(text) <= n {
		return text
	}
	for n > 0 && !utf8.RuneStart(text[n]) {
		n--
	}
	return text[:n] + "..."
}

// gopodFeedJournal prints a feed line: "GOPOD_FEED 42 Brobot 1 (to Brobot 2): words".
func gopodFeedJournal(e gopodChatFeedEvent) {
	head := e.Who
	if e.To != "" {
		head += " (to " + e.To + ")"
	}
	if e.Via != "" {
		head += " [" + e.Via + "]"
	}
	text := strings.Join(strings.Fields(e.Text), " ")
	if len(text) > 300 {
		text = gopodCut(text, 300)
	}
	logger.Println("GOPOD_FEED " + strconv.FormatInt(int64(e.ID), 10) + " " + head + ": " + text)
}

// GopodChatFeedAddTo adds one line said to one Brobot ("Brobot 1"), so the page's operator bubble
// reads "Operator to Brobot 1".
func GopodChatFeedAddTo(who, to, text string) {
	gopodChatFeedAddEvent(who, to, "", text)
}

func gopodChatFeedAddEvent(who, to, by, text string) {
	gopodChatFeedAddEventVia(who, to, by, "", text)
}

func gopodChatFeedAddEventVia(who, to, by, via, text string) {
	gopodChatFeedAddEventTurn(who, to, by, via, text, gopodTurnForLine())
}

// gopodChatFeedAddEventTurn: a line, with the turn it belongs to (gopod_chat_queue.go).
func gopodChatFeedAddEventTurn(who, to, by, via, text string, turn int64) {
	who, to, text = strings.TrimSpace(who), strings.TrimSpace(to), strings.TrimSpace(text)
	if who == "" || text == "" {
		return
	}
	text = gopodStrikeExpletives(text)
	gopodExperienceMarkPlayed(who, text)
	if strings.HasPrefix(text, "/gopod_reporter_audio/") && strings.HasSuffix(text, ".wav") {
		go gopodVoiceUsed(strings.TrimPrefix(text, "/gopod_reporter_audio/"))
	}
	gopodChatFeed.Lock()
	gopodChatFeed.next++
	event := gopodChatFeedEvent{ID: gopodChatFeed.next, Who: who, To: to, By: by, Via: via, Text: text, TimeMS: time.Now().UnixMilli(), Turn: turn}
	if as := gopodPersonaShown(to); as != to {
		event.ToAs = as
	}
	if who == "Brobot 1" || who == "Brobot 2" {
		gopodLastBrobotFeed.Store(who, event.ID)
	}
	gopodFeedJournal(event)
	if via != gopodFeedLiveVia {
		defer gopodRecordFeedLine(event)
	}
	defer gopodChatFeed.Unlock()
	gopodChatFeed.events = append(gopodChatFeed.events, event)
	if max := gopodChatFeedMax(); len(gopodChatFeed.events) > max {
		gopodChatFeed.events = gopodChatFeed.events[len(gopodChatFeed.events)-max:]
	}
	gopodWriteChatFeedLocked()
}

// gopodChatFeedSince returns a copy of the feed lines added after t, oldest first (used by the
// Brobots' shared chat memory, gopodSharedBrobotChat).
func gopodChatFeedSince(t time.Time) []gopodChatFeedEvent {
	gopodChatFeed.Lock()
	defer gopodChatFeed.Unlock()
	var out []gopodChatFeedEvent
	for _, e := range gopodChatFeed.events {
		if time.UnixMilli(e.TimeMS).After(t) {
			out = append(out, e)
		}
	}
	return out
}

// gopodWriteChatFeedLocked writes the feed file; the caller holds the lock.
func gopodWriteChatFeedLocked() {
	data, err := json.Marshal(struct {
		Boot   int64                `json:"boot"`
		Events []gopodChatFeedEvent `json:"events"`
	}{gopodChatFeed.boot, append([]gopodChatFeedEvent{}, gopodChatFeed.events...)})
	if err != nil {
		return
	}
	tmp := gopodChatFeedPath + ".tmp"
	if err := gopodWriteFile(tmp, data, 0644); err != nil {
		logger.Println("GOPOD chat feed: unable to write " + tmp + ": " + err.Error())
		return
	}
	if err := os.Rename(tmp, gopodChatFeedPath); err != nil {
		logger.Println("GOPOD chat feed: unable to write " + gopodChatFeedPath + ": " + err.Error())
	}
}

// ## EOF

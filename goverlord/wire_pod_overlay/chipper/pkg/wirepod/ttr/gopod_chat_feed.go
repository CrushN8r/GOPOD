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

// The Brobots chat page's feed of song/tool lines (2026-09-30). Every line a
// caller shows on the page (the display_text of /api-sdk/say_text - songs, the
// interview, go search, the shell aliases) used to reach the page only as an
// untimestamped "Brobot N: ..." line in Wire-Pod's 50-line UI log, so the page
// could only tell lines apart by their words: a line said twice, or the same
// song run twice, was taken as already shown. Here each line gets its own
// number, and the page (js/gopod_brobots_chat.js, pollFeed) shows each number
// once. A plain file next to the page, written whole each time (temp file, then
// rename), the same way gopod_speaking_now.json works. The UI log line itself
// is unchanged.

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
	Text   string `json:"text"`
	TimeMS int64  `json:"time_ms"`
}

var gopodChatFeed = struct {
	sync.Mutex
	boot   int64
	next   int
	events []gopodChatFeedEvent
}{boot: time.Now().UnixMilli()}

// GopodChatFeedStart writes an empty feed with this Wire-Pod start's stamp. Called
// once at startup (sdkapp BeginServer): the page clears its chat when the stamp
// changes, i.e. after every Wire-Pod restart (operator: "upon wpr 2, clear out
// brobots chat window contents").
func GopodChatFeedStart() {
	// 2026-10-09: the sounds the page plays (gopod_robot_cmd_log.go).
	gopodSoundsToPage()
	gopodChatFeed.Lock()
	defer gopodChatFeed.Unlock()
	gopodChatFeed.next, gopodChatFeed.events = 0, nil
	gopodWriteChatFeedLocked()
}

// GopodChatFeedAdd adds one line to the page's feed. who is the speaker as the
// page shows it ("Brobot 1", "Brobot 2", "Operator"); an empty who or text is
// ignored.
func GopodChatFeedAdd(who, text string) {
	// 2026-10-09 (operator: "no 'LLM Ready for Bingo' voice audio"): a line from
	// reporter_voices.system_say_role (a song's) is said in Goverlord's voice, with
	// no bubble of its own (the song's warm-up bubble already shows it).
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

// gopodChatFeedAddOperator adds an operator line to a Brobot ("Brobots" for
// both). A playing draft's own question is labelled on the page as the reporter
// who reads it, like the opening's bubbles (operator 2026-10-08: "relabel
// operator to Reporter ..., same as opening bubbles * keyboard & numpad/mic
// input stay as operator"); it stays an operator line everywhere else.
func gopodChatFeedAddOperator(to, text, via string) {
	gopodChatFeedAddEventVia("Operator", to, gopodScriptQuestionReader(to, text), via, text)
}

// gopodOperatorVia (2026-10-10, operator: "'Operator' now update message
// according to input * "[input] to Brobot [N]", "KP[N] to Brobot [N]" ... "Mic
// to Brobot [N]""): the input an operator line came in by, for its bubble's label.
// Numpad push-to-talk is its key ("KP1"); a line the robot heard himself (Hey
// Vector, the backpack button) is "Backpack".
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
	// 2026-10-24: a visitor's line is named by their approved username.
	if strings.HasPrefix(how, gopodVisitorHowPrefix) {
		return strings.TrimPrefix(how, gopodVisitorHowPrefix)
	}
	return ""
}

// gopodFeedLiveVia: a live (LLM) reply's feed line (GopodChatFeedAddLive).
const gopodFeedLiveVia = "llm"

// GopodChatFeedAddLive (2026-10-17, survey E6: live replies reached the page only
// through Wire-Pod's 50-line UI log): a Brobot's live reply goes on the feed, in
// order with every other line; the page shows it as a live bubble.
func GopodChatFeedAddLive(esn, text string) {
	who := gopodBrobotName(esn)
	// The same reply twice in a row from one Brobot: said in the journal.
	words := strings.Join(strings.Fields(text), " ")
	gopodLiveCheck.Lock()
	if last, ok := gopodLiveCheck.last[who]; ok && last == words && time.Since(gopodLiveCheck.at[who]) < 30*time.Second {
		logger.Println("GOPOD_FEED_CHECK live reply twice in a row from " + who + ": " + gopodHeardShort(words))
	}
	gopodLiveCheck.last[who], gopodLiveCheck.at[who] = words, time.Now()
	gopodLiveCheck.Unlock()
	gopodChatFeedAddEventVia(who, "", "", gopodFeedLiveVia, text)
	gopodChatFeed.Lock()
	id := gopodChatFeed.next
	gopodChatFeed.Unlock()
	gopodLiveCheck.Lock()
	gopodLiveCheck.open[id] = who
	gopodLiveCheck.lastID[esn] = id
	gopodLiveCheck.Unlock()
}

// gopodLiveCheck (2026-10-17, operator: "If a reply bubble goes missing or shows
// twice ... can you account for to reveal this possible glitch, in all-logs?"):
// each live reply line waits for the page's word on its bubble; one the page
// has played past with no word is logged as missing (GopodPagePlayed).
var gopodLiveCheck = struct {
	sync.Mutex
	last map[string]string
	at   map[string]time.Time
	open map[int]string
	// lastID: each Brobot's newest live reply line (his ESN -> its feed id).
	lastID map[string]int
}{last: map[string]string{}, at: map[string]time.Time{}, open: map[int]string{}, lastID: map[string]int{}}

// GopodLiveReplyNoted: the page said what became of live reply id's bubble.
func GopodLiveReplyNoted(id int) {
	gopodLiveCheck.Lock()
	delete(gopodLiveCheck.open, id)
	gopodLiveCheck.Unlock()
}

// The page's side of the dialogue (2026-10-17, operator: "the web page and
// terminal must have some sort of coordinated dialogue"): the page says how far
// it has played the feed (each line shown, its voice ended), and the journal
// gets "GOPOD_PAGE played <id> <who>: <words>" under the engine's own
// "GOPOD_FEED <id> ..." - both sides of the same run, in one list.
var gopodPagePlayedTo int

// GopodPagePlayed: the page (this boot's feed) has played up to id.
func GopodPagePlayed(boot int64, id int) {
	gopodChatFeed.Lock()
	if boot != gopodChatFeed.boot || id <= gopodPagePlayedTo {
		// 2026-10-24 (live: a reply shown, its "played" never logged): a report
		// that's not taken says why.
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
	// A live reply the page played past with no word on its bubble: missing.
	// (Its word comes the moment it shows; "played" a second later.)
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

// gopodCut (2026-10-24, live: a "played" line cut inside a 3-byte mark was kept by
// the system log as unreadable data, so the dry run never saw it): text cut to at
// most n bytes, never inside a character, then "...".
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

// GopodChatFeedAddTo adds one line said to one Brobot ("Brobot 1"), so the page's
// operator bubble reads "Operator to Brobot 1" (2026-10-04, a typed line).
func GopodChatFeedAddTo(who, to, text string) {
	gopodChatFeedAddEvent(who, to, "", text)
}

func gopodChatFeedAddEvent(who, to, by, text string) {
	gopodChatFeedAddEventVia(who, to, by, "", text)
}

func gopodChatFeedAddEventVia(who, to, by, via, text string) {
	who, to, text = strings.TrimSpace(who), strings.TrimSpace(to), strings.TrimSpace(text)
	if who == "" || text == "" {
		return
	}
	// 2026-10-23: listed expletives shown struck through (gopodStrikeExpletives).
	text = gopodStrikeExpletives(text)
	// A Brobot saying a Bingo test line means the Bingo test has been played
	// since this Wire-Pod start (gopodBingoPlayed, gopod_brobot_personas.go).
	// The operator typing "Bingo Test Live" doesn't count; the song's own lines do.
	if who != "Operator" && strings.Contains(text, gopodBingoRunMarker) {
		gopodBingoPlayed.Store(true)
	}
	gopodChatFeed.Lock()
	gopodChatFeed.next++
	event := gopodChatFeedEvent{ID: gopodChatFeed.next, Who: who, To: to, By: by, Via: via, Text: text, TimeMS: time.Now().UnixMilli()}
	// 2026-10-21 (operator: "highlighting ... should run as an independent agent on
	// any & all bubbles"): a Brobot's newest bubble, so his speech names it exactly.
	if who == "Brobot 1" || who == "Brobot 2" {
		gopodLastBrobotFeed.Store(who, event.ID)
	}
	// 2026-10-17 (operator: "one event list, single source of truth"; "both the
	// terminal run & web page run must be the same run"): every line the page
	// will show or play is in the journal too, numbered, in the same order.
	gopodFeedJournal(event)
	// A Record Session keeps every line (gopod_record_session.go), after the
	// feed's own lock is let go. A live reply is recorded as its exchange
	// (gopodRecordExchange), not again here.
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

// gopodChatFeedSince returns a copy of the feed lines added after t, oldest first
// (used by the Brobots' shared chat memory, gopodSharedBrobotChat).
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
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		logger.Println("GOPOD chat feed: unable to write " + tmp + ": " + err.Error())
		return
	}
	if err := os.Rename(tmp, gopodChatFeedPath); err != nil {
		logger.Println("GOPOD chat feed: unable to write " + gopodChatFeedPath + ": " + err.Error())
	}
}

// ## EOF

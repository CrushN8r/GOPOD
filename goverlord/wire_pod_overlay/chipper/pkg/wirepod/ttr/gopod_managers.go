// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_managers.go
package wirepod_ttr

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The managers (2026-10-18, operator: "CHALK & PLAYHEAD can take over system
// commands * KP1, what's the weather * 'weather' detected * CHALK says, I'll help
// you with that, and takes over the request * after I can still ask the brobots
// about the weather and other system results in the chat * they get the same
// thought bubbles * I see a lot of ways that reduce the load on the brobots";
// then: "PLAYHEAD * he now takes over 'time & location' stuff * WAWN, where are
// we now? * weather * his bubble background is tape ribbon color of vintage video
// tape cassettes"; "CHALK (display) now is to take ownership of: search-results
// box ... system-status and the other system bubbles * hidden/surprise egg
// cards"). In the live chat, a line with one of a take's words: its manager (a
// reporter voice, its own colour) says take_text, then does the job - the
// weather (its egg, then the weather line) or the time and place. The Brobots
// still get the facts in their prompt, but don't open their reply with the
// weather for release_seconds. CHALK's display bubbles are drawn by the page
// (managers.display_host).

// gopodManagerWeatherAt: when a manager last said the weather (unix ms).
var gopodManagerWeatherAt atomic.Int64

type gopodManagerTake struct {
	Host     string   `json:"host"`
	Kind     string   `json:"kind"`
	Words    []string `json:"words"`
	TakeText string   `json:"take_text"`
	// 2026-10-22: the Brobot says the result, the manager only take_text.
	BrobotSays bool `json:"brobot_says"`
}

// gopodManagerTakes: the take this line asks for (the first word in the line), live only.
func gopodManagerTakes(lower string) (gopodManagerTake, int, bool) {
	m := gopodBrobotsSettings.Managers
	if gopodScriptPlaying.Load() || gopodRunActive() {
		return gopodManagerTake{}, -1, false
	}
	best, bestAt := gopodManagerTake{}, -1
	for _, t := range m.Takes {
		if t.Host == "" {
			continue
		}
		at, _ := gopodEggWord(lower, t.Words)
		if t.Kind == "url_qr" {
			at = -1
			if loc := gopodManagerURLRE.FindStringIndex(lower); loc != nil {
				at = loc[0]
			}
		}
		if at >= 0 && (bestAt < 0 || at < bestAt) {
			best, bestAt = t, at
		}
	}
	return best, bestAt, bestAt >= 0
}

// gopodManagerTookWeather: a manager said the weather for this exchange.
func gopodManagerTookWeather() bool {
	t := gopodManagerWeatherAt.Load()
	return t != 0 && time.Since(time.UnixMilli(t)) < gopodSecondsToDuration(gopodBrobotsSettings.Managers.ReleaseSeconds)
}

// gopodManagerTo (2026-10-18, operator: "If I ask CHALK, PLAYHEAD, brobots or
// 'All' = take over trigger * CHALK or PLAYHEAD takes over * If I ask brobot N =
// weather trigger * PLAYHEAD takes over in PLAYHEAD bubble * PLAYHEAD bubble
// displays inside brobot N bubble * at top before brobot N 'Got it'"): the Brobot
// the newest operator line went to ("Brobot 1"), whose bubble the manager's lines
// go in; "" when it went to the Brobots, All or a manager (its own bubbles).
func gopodManagerTo() string {
	gopodChatFeed.Lock()
	defer gopodChatFeed.Unlock()
	for i := len(gopodChatFeed.events) - 1; i >= 0; i-- {
		if e := gopodChatFeed.events[i]; e.Who == "Operator" {
			if e.To == "Brobot 1" || e.To == "Brobot 2" {
				return e.To
			}
			// 2026-10-18 (operator: "In the run the line went to both Brobots * so
			// the managers will show in their own bubbles: incorrect * Once
			// managers have dealt with the brobot that asked. Their jobs are
			// done"): a line to both - inside the bubble of the Brobot answering
			// first. A line to a manager - its own bubbles.
			if _, ok := gopodBrobotsSettings.Managers.Personas[e.To]; ok || gopodHostIsManager(e.To) {
				return ""
			}
			return gopodBrobotName(gopodAllFirst(e.Text))
		}
	}
	return ""
}

// gopodManagerBusy: a manager has the floor (gopodManagerDo) - a Brobot's
// searching waits (gopodThinkingProcess).
var gopodManagerBusy atomic.Bool

// gopodManagerLastHost: the manager who last had the floor ("PLAYHEAD").
var gopodManagerLastHost atomic.Value

// gopodManagerStepIn (2026-10-22, the Brobot Bubble template's notice line 3):
// the take this Brobot's line asks for (its words, or a native intent's), done
// now - after his "Got it!" and value point line, while his answer is being
// written; its lines go in his notice. Once per line (the second Brobot in a line
// to both doesn't take it again).
var gopodManagerStepped = struct {
	sync.Mutex
	line string
	at   time.Time
}{}

func gopodManagerStepIn(brobot, question string) {
	take, _, ok := gopodManagerTakes(strings.ToLower(question))
	if !ok {
		take, ok = gopodNativeTakeFor(brobot, question)
	}
	if !ok {
		return
	}
	line := strings.TrimSpace(strings.ToLower(question))
	gopodManagerStepped.Lock()
	again := gopodManagerStepped.line == line && time.Since(gopodManagerStepped.at) < time.Minute
	if !again {
		gopodManagerStepped.line, gopodManagerStepped.at = line, time.Now()
	}
	gopodManagerStepped.Unlock()
	if again {
		return
	}
	gopodManagerDoFor(take, question, brobot)
}

// gopodWeatherLineEnd (2026-10-22, operator: "add weather image centered below,
// after weather report fully spoken * then the rest of the response"): this
// spoken piece ends the golden weather line ("... as of 4:45 AM or 0445 Hours
// Military Time.").
func gopodWeatherLineEnd(said string) bool {
	l := strings.ToLower(said)
	return strings.Contains(l, "as of ") && (strings.Contains(l, "military time") || strings.Contains(l, "ay-emm") || strings.Contains(l, "pee-emm") || strings.Contains(l, " am") || strings.Contains(l, " pm"))
}

// gopodWeatherShowMark: the page shows his weather picture and plays its sound.
const gopodWeatherShowMark = "weathershow:"

// gopodWeatherLineSaid: his weather line is said - the page shows its picture
// and plays its sound; he waits weather_card.pause_seconds, then goes on.
func gopodWeatherLineSaid(esn string) {
	brobot := gopodBrobotName(esn)
	if (brobot != "Brobot 1" && brobot != "Brobot 2") || !gopodBrobotsSettings.WeatherCard.Picture {
		return
	}
	GopodChatFeedAddTo(gopodBrobotsSettings.TopicPictures.Role, brobot, gopodWeatherShowMark)
	if p := gopodBrobotsSettings.WeatherCard.PauseSeconds; p > 0 {
		logger.Println(fmt.Sprintf("GOPOD_WEATHER %s: picture and sound, then %.1f s", brobot, p))
		time.Sleep(gopodSecondsToDuration(p))
	}
}

// gopodManagerDo: the manager takes the line and does its job.
func gopodManagerDo(t gopodManagerTake, userMessage string) {
	gopodManagerDoFor(t, userMessage, gopodManagerTo())
}

// gopodManagerDoFor: gopodManagerDo with its lines in this Brobot's bubble.
func gopodManagerDoFor(t gopodManagerTake, userMessage, to string) {
	gopodManagerBusy.Store(true)
	gopodManagerLastHost.Store(t.Host)
	defer gopodManagerBusy.Store(false)
	place := gopodWeatherPlace(userMessage)
	if t.Kind == "weather" && !t.BrobotSays {
		gopodManagerWeatherAt.Store(time.Now().UnixMilli())
	}
	if s := strings.TrimSpace(t.TakeText); s != "" {
		gopodReporterSayTo(t.Host, to, s, true, true)
	}
	switch t.Kind {
	case "weather":
		// 2026-10-22 (operator: "no weather voice generation or display, weather
		// spoken by brobot"): his reply opens with the golden weather line.
		if t.BrobotSays {
			return
		}
		line, _ := gopodWeatherSayAt(gopodBrobot1ESN, place)
		// 2026-10-18 (operator: "please remove easter egg status from weather * no
		// more weather pop up. too common"): no egg - the line only.
		if line != "" {
			gopodReporterSayTo(t.Host, to, line, true, true)
		}
		gopodManagerWeatherAt.Store(time.Now().UnixMilli())
	case "time_place":
		if line := gopodTimePlaceLine(place); line != "" {
			gopodReporterSayTo(t.Host, to, line, true, true)
		}
	case "url_qr":
		// 2026-10-18 (operator: "PLAYHEAD bubble displays, i can help message,
		// then generated QR of [url] location"): the QR, made now, in his bubble.
		if url := gopodManagerURL(userMessage); url != "" {
			if page, err := gopodMakeQR(url); err == nil {
				gopodChatFeedAddEventVia(t.Host, to, "", "", gopodManagerQRMark+page+"|"+url)
				gopodManagerQR.Store(url)
				gopodManagerQRAt.Store(time.Now().UnixMilli())
				gopodManagerQRShowing.Store(gopodManagerQRShown{to: to, line: gopodManagerQRMark + page + "|" + url})
			} else {
				logger.Println("GOPOD_MANAGER no QR for " + url + ": " + err.Error())
			}
		}
	}
}

// gopodManagerQRShown: the QR CHALK still has to display after the Brobot's answer.
type gopodManagerQRShown struct{ to, line string }

var gopodManagerQRShowing atomic.Value

// gopodManagerQRShow (2026-10-18, operator: "CHALK BUBBLE === I can help with
// that. QR code displayed. [live generated QR for url]" after "Brobot 1:
// [response]"; "Brobot 2 then responds or doesn't, accordingly"): when the first
// Brobot's answer ends, CHALK (display_host) says qr_display_text and shows the
// same QR - in his bubble, at the bottom, for a line to him; in her own bubble
// otherwise. Before his answer's end is told, so the other Brobot comes after.
func gopodManagerQRShow() {
	q, _ := gopodManagerQRShowing.Swap(gopodManagerQRShown{}).(gopodManagerQRShown)
	host := gopodBrobotsSettings.Managers.DisplayHost
	if q.line == "" || host == "" {
		return
	}
	if s := strings.TrimSpace(gopodBrobotsSettings.Managers.QRDisplayText); s != "" {
		gopodReporterSayTo(host, q.to, s, true, true)
	}
	gopodChatFeedAddEventVia(host, q.to, "", "", q.line)
	// Their jobs are done: the other Brobot answers like before, no QR note.
	gopodManagerQRAt.Store(0)
}

// gopodHostIsManager: a reporter host a take or the display belongs to (CHALK, PLAYHEAD).
func gopodHostIsManager(who string) bool {
	m := gopodBrobotsSettings.Managers
	if who == "" {
		return false
	}
	if who == m.DisplayHost {
		return true
	}
	for _, t := range m.Takes {
		if t.Host == who {
			return true
		}
	}
	for _, p := range m.Personas {
		if p.Host == who {
			return true
		}
	}
	return false
}

// gopodManagerQRMark: a manager's QR line ("qr:<picture>|<url>"), shown, not said.
const gopodManagerQRMark = "qr:"

// gopodManagerQR / gopodManagerQRAt: the address PLAYHEAD last made a QR for, and when.
var gopodManagerQR atomic.Value
var gopodManagerQRAt atomic.Int64

// gopodManagerURLRE: a web address in a line ("flare4.com", "https://crushn8r.ca/x").
var gopodManagerURLRE = regexp.MustCompile(`(?i)\b(?:https?://)?((?:[a-z0-9-]+\.)+(?:com|ca|net|org|io|co|tv|app|dev|info|me|us|uk)(?:/[^\s]*)?)`)

// gopodManagerURL: the first web address in the line, as https://..., or "".
func gopodManagerURL(text string) string {
	m := gopodManagerURLRE.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	return "https://" + strings.TrimRight(m[1], ".,!?)")
}

// gopodManagerQRNote: the Brobot's thought bubble (managers.qr_note) while a QR
// PLAYHEAD made is in his bubble (release_seconds), or "".
func gopodManagerQRNote() string {
	t := gopodManagerQRAt.Load()
	url, _ := gopodManagerQR.Load().(string)
	if t == 0 || url == "" || time.Since(time.UnixMilli(t)) > gopodSecondsToDuration(gopodBrobotsSettings.Managers.ReleaseSeconds) {
		return ""
	}
	return strings.ReplaceAll(gopodBrobotsSettings.Managers.QRNote, "{url}", strings.TrimPrefix(url, "https://"))
}

// gopodMakeQR: a QR picture for url (webroot/qr/made/, python3's qrcode), its page path.
var gopodMakeQR = func(url string) (string, error) {
	sum := sha1.Sum([]byte(url))
	name := hex.EncodeToString(sum[:6]) + ".png"
	dir := filepath.Join(gopodChipperRoot(), "webroot", "qr", "made")
	out := filepath.Join(dir, name)
	if !gopodFileExists(out) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
		cmd := gopodAsOperator(exec.Command("python3", "-c", "import qrcode,sys; qrcode.make(sys.argv[1], border=2).save(sys.argv[2])", url, out))
		if msg, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("qrcode: %v %s", err, strings.TrimSpace(string(msg)))
		}
	}
	return "/qr/made/" + name, nil
}

// gopodTimePlaceLine: "Windsor Ontario Canada 🇨🇦, Sunday, October 18, 2026, 1405
// Hours Military Time or 2:05 PM." (managers.time_place_text), or offline_text.
func gopodTimePlaceLine(place string) string {
	facts, err := gopodLiveWeatherFacts(place)
	if err != nil && place != "" {
		facts, err = gopodLiveWeatherFacts("")
	}
	if err != nil {
		return gopodBrobotsSettings.LiveWeather.OfflineText
	}
	where, clock := gopodPlaceAndClock(gopodBrobot1ESN, facts)
	return strings.NewReplacer("{place}", where, "{weekday}", facts.Local.Format("Monday"),
		"{date}", facts.Local.Format("January 2, 2006"), "{clock}", clock).Replace(gopodBrobotsSettings.Managers.TimePlaceText)
}

// ## EOF

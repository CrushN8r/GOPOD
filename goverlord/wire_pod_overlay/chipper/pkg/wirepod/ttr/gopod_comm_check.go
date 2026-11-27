// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_comm_check.go
package wirepod_ttr

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
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
	if gopodRunCheckSilent() {
		return
	}
	for end := time.Now().Add(max); time.Now().Before(end) && !gopodCommCheckStop.Load(); time.Sleep(250 * time.Millisecond) {
		if gopodPagePlayedUpTo() >= id {
			return
		}
	}
}

// gopodCommCheckWaitQuiet: until the feed has had no new line for quiet and the page has played it
// all (or max).
func gopodCommCheckWaitQuiet(quiet, max time.Duration) {
	last, since := gopodFeedLastID(), time.Now()
	for end := time.Now().Add(max); time.Now().Before(end) && !gopodCommCheckStop.Load(); time.Sleep(500 * time.Millisecond) {
		if id := gopodFeedLastID(); id != last {
			last, since = id, time.Now()
			continue
		}
		if time.Since(since) >= quiet && (gopodRunCheckSilent() || gopodPagePlayedUpTo() >= last) {
			return
		}
	}
}

// gopodCommCheckLines: every round's text on each of its keys, in order, with who asks it.
func gopodCommCheckLines() [][4]string {
	cc := gopodBrobotsSettings.CommCheck
	var lines [][4]string // key, text, asker, "self" for the asker's own turn
	asked := map[string]bool{}
	for _, r := range cc.Rounds {
		keys := r.Keys
		if len(keys) == 0 {
			keys = cc.Keys
		}
		for _, k := range keys {
			if strings.EqualFold(k, gopodCommCheckMapKeys) {
				for _, mk := range gopodNumpadLiveKeys() {
					if _, t, _ := gopodNumpadBrobot(mk); t == gopodTypedSystem {
						continue
					}
					if !asked[mk] && !strings.EqualFold(mk, gopodNumpadAllKey) {
						asked[mk] = true
						lines = append(lines, [4]string{mk, r.Text, r.Asker, ""})
					}
				}
				continue
			}
			asked[k] = true
			self := ""
			if r.Self {
				self = "self"
			}
			lines = append(lines, [4]string{k, r.Text, r.Asker, self})
		}
	}
	return lines
}

// gopodCommCheckMapKeys: a comm check step's keys word for every live map key.
const gopodCommCheckMapKeys = "map"

// gopodCommCheckPersonaKey: a comm check step's key for a persona by name ("persona:GOVERLORD").
const gopodCommCheckPersonaKey = "persona:"

// gopodCommCheckSelf: the asker's own turn is being answered (no Got it line).
var gopodCommCheckSelf atomic.Bool

// gopodPersonaNick: a persona's nickname (its card), else its call name.
func gopodPersonaNick(who string) string {
	if n := strings.TrimSpace(gopodBrobotsSettings.HostPersonas[who].Nickname); n != "" {
		return n
	}
	return gopodCallName(who)
}

// gopodCommCheckHandover: Goverlord's line handing the comm check to KEYS - KEYS answers it with
// his Got it and the crew brain bubble only (comm_check.crew), not an intro.
func gopodCommCheckHandover(host, text string) bool {
	return gopodCommCheckRunning.Load() && host == gopodBrobotsSettings.SystemBubble.VoicePersona &&
		strings.EqualFold(gopodCommCheckAsker(text), gopodGoverlordHost)
}

// gopodCommCheckCrew: the crew brain bubble, in KEYS' bubble.
func gopodCommCheckCrew(host, to string) {
	title := strings.ReplaceAll(strings.TrimSpace(gopodBrobotsSettings.BubbleNotices.ValuePoints.Title), "{name}", gopodCallName(host))
	for _, line := range gopodCommCheckCrewLines(host) {
		if line = strings.TrimSpace(line); line != "" && title != "" {
			gopodBubbleNotice(to, title, line)
		}
	}
}

// gopodCommCheckCrewLines: the crew bubble from the company's folders - the folder a persona card
// sits in is its crew.
func gopodCommCheckCrewLines(host string) []string {
	company := gopodTopicPath(gopodBrobotsSettings.BrainChain.Roster)
	dir := filepath.Dir(company)
	chain, roster := "", ""
	if data, err := os.ReadFile(company); err == nil {
		for _, l := range strings.Split(string(data), "\n") {
			if m := gopodProfilePointRE.FindStringSubmatch(strings.TrimSpace(l)); m != nil && strings.HasPrefix(m[1], "Chain of command:") {
				chain = strings.ToLower(gopodPersonaFill("", m[1]))
			}
		}
	}
	for _, g := range gopodRosterGroups() {
		roster += " " + strings.ToLower(strings.Join(g.Names, " | "))
	}
	at := func(names []string) (int, int) {
		c, r := 1<<30, 1<<30
		for _, n := range names {
			if n = strings.ToLower(strings.TrimSpace(n)); n == "" {
				continue
			}
			if i := strings.Index(chain, n); i >= 0 && i < c {
				c = i
			}
			if i := strings.Index(roster, n); i >= 0 && i < r {
				r = i
			}
		}
		return c, r
	}
	title := ""
	if data, err := os.ReadFile(company); err == nil {
		for _, l := range strings.Split(string(data), "\n") {
			if l = strings.TrimSpace(l); strings.HasPrefix(l, "# ") {
				title = strings.ToUpper(strings.TrimSpace(l[2:]))
				break
			}
		}
	}
	// Each persona card by the folder it sits in - any card in a folder under the company's, but a
	// group's ("group-") or a schema; named by its nickname.
	type card struct {
		path, nick string
		names      []string
	}
	hostFile := filepath.Clean(gopodPersonaProfileFileIn(gopodBrobotsSettings.PersonaProfiles.Folder, host))
	byDir := map[string][]card{}
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".md") || strings.HasPrefix(d.Name(), "group-") || strings.Contains(d.Name(), "schema") {
			return nil
		}
		rel, err := filepath.Rel(dir, filepath.Dir(p))
		if err != nil || rel == "." {
			return nil
		}
		head, _ := gopodPersonaProfileReadFile(p)
		nick := strings.TrimSpace(head["nickname"])
		if nick == "" {
			return nil
		}
		name := ""
		if data, err := os.ReadFile(p); err == nil {
			if first := strings.SplitN(string(data), "\n", 2)[0]; strings.HasPrefix(first, "# ") {
				name = strings.TrimSpace(first[2:])
			}
		}
		c := card{path: filepath.Clean(p), nick: nick, names: []string{nick, name}}
		for w := range gopodBrobotsSettings.HostPersonas {
			if f := gopodPersonaProfileFileIn(gopodBrobotsSettings.PersonaProfiles.Folder, w); f != "" && filepath.Clean(f) == c.path {
				c.names = append(c.names, w, gopodCallName(w))
			}
		}
		byDir[rel] = append(byDir[rel], c)
		return nil
	})
	var dirs []string
	for d := range byDir {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	words := func(seg string) string {
		seg = strings.TrimLeft(seg, "0123456789")
		seg = strings.TrimLeft(seg, "_- ")
		var out []string
		for _, w := range strings.FieldsFunc(seg, func(r rune) bool { return r == '_' || r == '-' || r == ' ' }) {
			out = append(out, gopodCapFirst(w))
		}
		return strings.Join(out, " ")
	}
	var crews []string
	for _, d := range dirs {
		segs := strings.Split(filepath.ToSlash(d), "/")
		label := words(segs[0]) + " Crew"
		if len(segs) > 1 {
			var part []string
			for _, s := range segs[1:] {
				part = append(part, words(s))
			}
			label += " (" + strings.Join(part, " ") + ")"
		}
		who := byDir[d]
		sort.SliceStable(who, func(i, j int) bool {
			ci, ri := at(who[i].names)
			cj, rj := at(who[j].names)
			if ci != cj {
				return ci < cj
			}
			if ri != rj {
				return ri < rj
			}
			return who[i].path < who[j].path
		})
		var names []string
		for _, c := range who {
			n := c.nick
			if c.path == hostFile {
				n += " (Myself)"
			}
			names = append(names, n)
		}
		crews = append(crews, "--- "+label+" --- "+strings.Join(names, ", "))
	}
	var out []string
	for _, line := range gopodBrobotsSettings.CommCheck.Crew {
		if strings.TrimSpace(line) == "{crews}" {
			out = append(out, crews...)
			continue
		}
		out = append(out, strings.ReplaceAll(line, "{company}", title))
	}
	return out
}

// gopodCommCheckAsking: the comm check line going out and who asks it - its bubble is his
// (gopodScriptQuestionReader) and it's read in his voice (gopodReporterReadsOperator), not the
// Operator's.
var gopodCommCheckAsking atomic.Value // [2]string{text, asker}

// gopodCommCheckAsker: who asks this comm check line, "" when it isn't one.
func gopodCommCheckAsker(text string) string {
	a, _ := gopodCommCheckAsking.Load().([2]string)
	if a[1] == "" || strings.TrimSpace(a[0]) != strings.TrimSpace(text) {
		return ""
	}
	return a[1]
}

// gopodCommCheckWaitAnswer: until a feed line after from comes from who (his answer), or max.
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
		defer gopodTurnBeginOwn("comm_check", false)()
		quiet := gopodSecondsToDuration(cc.QuietSeconds)
		max := gopodSecondsToDuration(cc.LineMaxSeconds)
		gopodCommCheckStop.Store(false)
		gopodIntroHeardReset()
		if gopodVoicesAtStart.Load() {
			logger.Println("GOPOD_COMM_CHECK waiting for the voice library to be ready")
			for end := time.Now().Add(10 * time.Minute); gopodVoicesAtStart.Load() && time.Now().Before(end) && !gopodCommCheckStop.Load(); time.Sleep(time.Second) {
			}
			logger.Println("GOPOD_COMM_CHECK voice library ready - starting")
		}
		if !gopodSecuritySweep() {
			logger.Println("GOPOD_COMM_CHECK held: security sweep")
			return
		}
		if cc.Wake {
			for _, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN} {
				if strings.TrimSpace(esn) != "" {
					gopodRobotCall(esn, true, func(*vector.Vector, context.Context) string { return "" })
				}
			}
		}
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
			var label, target, name string
			if p, ok := strings.CutPrefix(strings.TrimSpace(l[0]), gopodCommCheckPersonaKey); ok {
				label, target, name = p, gopodPersonaMark+p, gopodPersonaNick(p)
			} else {
				var err error
				label, target, err = gopodNumpadBrobot(key)
				if err != nil {
					logger.Println("GOPOD_COMM_CHECK " + key + ": " + err.Error())
					continue
				}
				name, _ = gopodNumpadKeyLabelIn(gopodNumpadLayout(), key)
			}
			if n := strings.TrimSpace(cc.AllName); n != "" && strings.EqualFold(label, "All") {
				name = n
			}
			text := gopodSpellNames(strings.ReplaceAll(l[1], "{name}", name))
			logger.Println("GOPOD_COMM_CHECK " + strconv.Itoa(i+1) + "/" + strconv.Itoa(len(lines)) + " " + key + " -> " + label + ": " + text)
			gopodNumpadLastKey.Store(key)
			gopodCommCheckAsking.Store([2]string{text, strings.TrimSpace(l[2])})
			if i == len(lines)-1 {
				gopodGameWelcomePending.Store(true)
			}
			from := gopodFeedLastID()
			answerer := label
			switch {
			case strings.HasPrefix(target, gopodPersonaMark):
				answerer = strings.TrimPrefix(target, gopodPersonaMark)
			case strings.HasPrefix(target, gopodPersonaReporterMark):
				answerer = gopodBrobotsSettings.Managers.Personas[strings.TrimPrefix(target, gopodPersonaReporterMark)].Reporter
			case strings.HasPrefix(target, gopodDialZeroMark):
				answerer = gopodBrobotName(strings.TrimPrefix(target, gopodDialZeroMark))
			}
			if target == gopodTypedAll {
				answerer = "Brobot 2"
				if gopodSolo() {
					answerer = "Brobot 1"
				}
			}
			if l[3] == "self" && strings.HasPrefix(target, gopodPersonaMark) {
				gopodCommCheckSelf.Store(true)
				gopodGomadAnswer(strings.TrimPrefix(target, gopodPersonaMark), text)
				gopodCommCheckSelf.Store(false)
			} else if err := gopodKeyLine(target, text, "Numpad text"); err != nil {
				logger.Println("GOPOD_COMM_CHECK " + err.Error())
			}
			gopodCommCheckWaitAnswer(from, answerer, max)
			gopodCommCheckWaitQuiet(quiet, max)
		}
		gopodGameReset("comm check")
		logger.Println("GOPOD_COMM_CHECK done: " + strconv.Itoa(len(lines)) + " lines")
		if t := strings.TrimSpace(cc.DoneText); t != "" {
			if role := gopodBrobotsSettings.SystemBubble.VoicePersona; role != "" {
				gopodReporterSay(role, strings.ReplaceAll(t, "{n}", strconv.Itoa(len(lines))), true, false)
			}
		}
	}()
	return strings.ReplaceAll(cc.StartText, "{n}", strconv.Itoa(len(lines)))
}

// gopodGameWelcomePending: the comm song's last line is out - its answer by
// comm_check.game_welcome.by ends with the game welcome.
var gopodGameWelcomePending atomic.Bool

// gopodGameWelcomeLine: the game welcome for this Brobot's reminder, once - the base word for word
// or one of his flavours, at random; "" otherwise.
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

// gopodApplyCommCheckSong: the comm check's rounds come from its song (comm_check.song, a song_id:
// runtime/songs/05_comm_check) - each "ask" step's line from story.md, asked on each of its keys
// (knobs.json) in order.
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
			Self   bool     `json:"self"`
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
				Self  bool     `json:"self"`
			}{texts[st.StepID], st.Keys, st.Asker, st.Self})
		}
	}
}

// ## EOF

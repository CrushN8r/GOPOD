// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_tour.go
package wirepod_ttr

import (
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

var gopodTourRunning, gopodTourStop atomic.Bool

// gopodSkippedAt: when the page last skipped (unix ms) - a stretch of lines being voiced stops
// there (gopodSayLinesInBubble).
var gopodSkippedAt atomic.Int64

// GopodSkipped: the page skipped (Esc, the numpad's *, a typed or spoken skip, the menu).
func GopodSkipped() { gopodSkippedAt.Store(time.Now().UnixMilli()) }

// gopodSystemTour starts the tour (or stops one going).
func gopodSystemTour() string {
	if gopodTourRunning.Load() {
		gopodTourStop.Store(true)
		return gopodSystemText("system_tour.system_tour_stopping_at_the")
	}
	if gopodRunActive() || gopodScriptPlaying.Load() {
		return gopodSystemText("system_tour.system_tour_a_run_is")
	}
	role := gopodBrobotsSettings.SignOff.SystemRole
	if role == "" {
		return gopodSystemText("system_tour.system_tour_no_system_voice")
	}
	scriptures := gopodTourScrolls()
	gopodTourRunning.Store(true)
	gopodTourStop.Store(false)
	go func() {
		defer gopodTourRunning.Store(false)
		defer gopodTurnBeginOwn("tour", true)()
		say := func(text string) bool {
			if gopodTourStop.Load() || gopodRunActive() {
				return false
			}
			here := gopodPageHere()
			gopodReporterSay(role, text, true, here)
			if here {
				gopodCommCheckWaitQuiet(gopodSecondsToDuration(gopodBrobotsSettings.SystemTour.QuietSeconds), 60*time.Second)
			}
			return true
		}
		for _, sc := range scriptures {
			if gopodTourStop.Load() || gopodRunActive() {
				return
			}
			lines := append([]string{sc.Title}, sc.Lines...)
			here := gopodPageHere()
			gopodSayLinesInBubble(role, lines, here)
			if here {
				gopodCommCheckWaitQuiet(gopodSecondsToDuration(gopodBrobotsSettings.SystemTour.QuietSeconds), 60*time.Second)
			}
		}
		say(gopodBrobotsSettings.SystemTour.EndText)
		logger.Println("GOPOD_TOUR done: " + strconv.Itoa(len(scriptures)) + " scriptures")
	}()
	return strings.ReplaceAll(gopodBrobotsSettings.SystemTour.StartText, "{n}", strconv.Itoa(len(scriptures)))
}

// gopodTourScroll: one stop of the tour - a scripture's title and what's said there.
type gopodTourScroll struct {
	Title string
	Lines []string
}

// gopodTourScrolls: the tour's stops, read from the scrolls' folder tree
// (system_scrolls/scripture-NN-*/README.md, in order): each its title, then its Tour points - or,
// with none, its Commandments and its Next.
func gopodTourScrolls() []gopodTourScroll {
	dirs, _ := filepath.Glob(filepath.Join(gopodRepoRoot(), gopodTorahDir, "scripture-*", "README.md"))
	sort.Strings(dirs)
	var out []gopodTourScroll
	for _, f := range dirs {
		if st, ok := gopodTourStopRead(gopodReadFileString(f)); ok {
			out = append(out, st)
		}
	}
	if len(out) > 0 {
		return out
	}
	files := gopodTorahFiles()
	var keys []string
	for k := range files {
		if strings.HasPrefix(k, gopodTorahDir+"/") {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		if st, ok := gopodTourStopRead(files[k]); ok {
			out = append(out, st)
		}
	}
	return out
}

// gopodTourStopRead: a scroll's title and what the tour says from it.
func gopodTourStopRead(text string) (gopodTourScroll, bool) {
	var st gopodTourScroll
	sections := map[string][]string{}
	section := ""
	for _, line := range strings.Split(text, "\n") {
		t := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(t, "# ") && st.Title == "":
			st.Title = strings.TrimSpace(t[2:])
		case strings.HasPrefix(t, "## "):
			section = strings.TrimSpace(t[3:])
		default:
			if m := gopodProfilePointRE.FindStringSubmatch(t); m != nil && section != "" {
				sections[section] = append(sections[section], m[1])
			}
		}
	}
	if st.Title == "" {
		return st, false
	}
	st.Lines = sections["Tour"]
	if len(st.Lines) == 0 {
		st.Lines = append(append([]string(nil), sections["Commandments"]...), sections["Next"]...)
	}
	return st, true
}

// gopodSameLineMark: in a "lines:" bubble, the join between two voice chunks of one written line
// (the page keeps them on that line).
const gopodSameLineMark = "\u2063"

// gopodChunksDisplay: chunks of text as the bubble shows them - joined by gopodSameLineMark, or by
// "\n" where the text itself starts a new line there.
func gopodChunksDisplay(text string, chunks []string) string {
	var b strings.Builder
	at := 0
	for i, c := range chunks {
		c = strings.TrimSpace(c)
		j := strings.Index(text[at:], c)
		if i > 0 {
			if j >= 0 && strings.Contains(text[at:at+j], "\n") {
				b.WriteString("\n")
			} else {
				b.WriteString(gopodSameLineMark)
			}
		}
		b.WriteString(c)
		if j >= 0 {
			at += j + len(c)
		}
	}
	return b.String()
}

// gopodSayLinesInBubble: lines in one bubble of host's ("lines:", the page's system bubble
// template), each its own voice - every line's voice named at once (so each starts dormant,
// Generating...
func gopodSayLinesInBubble(host string, lines []string, wait bool, display ...string) {
	rv := gopodBrobotsSettings.ReporterVoices
	var keep []string
	for _, l := range lines {
		if l = strings.TrimSpace(l); l != "" {
			keep = append(keep, l)
		}
	}
	if len(keep) == 0 {
		return
	}
	shown := strings.Join(keep, "\n")
	if len(display) > 0 && strings.TrimSpace(display[0]) != "" {
		shown = display[0]
	}
	GopodChatFeedAdd(host, "lines:"+shown)
	says := make([]string, len(keep))
	for i, l := range keep {
		says[i] = gopodSaysFor(host, l)
		said := says[i]
		if rv.GeneratingMarker == "" {
			continue
		}
		page, _, made := gopodReporterCached(host, said)
		if !made {
			_, page = gopodReporterPaths(host, said)
		}
		GopodChatFeedAdd(rv.Role, rv.GeneratingMarker+"forwait:"+page+"|"+l)
	}
	type made struct {
		page    string
		seconds float64
		ok      bool
	}
	ready := make([]chan made, len(keep))
	for i := range ready {
		ready[i] = make(chan made, 1)
	}
	makeOne := func(i int) {
		page, seconds, ok := gopodReporterFile(host, says[i])
		ready[i] <- made{page, seconds, ok}
	}
	go func() {
		for i := range keep {
			makeOne(i)
		}
	}()
	started := time.Now().UnixMilli()
	for i, l := range keep {
		m := <-ready[i]
		if gopodSkippedAt.Load() >= started {
			logger.Println("GOPOD_TOUR skipped: the rest of this bubble isn't voiced")
			return
		}
		if !m.ok {
			logger.Println("GOPOD_TOUR no voice for: " + gopodCut(l, 60))
			continue
		}
		gopodVoicePlay(host, l, m.page, m.seconds, wait)
	}
}

// ## EOF

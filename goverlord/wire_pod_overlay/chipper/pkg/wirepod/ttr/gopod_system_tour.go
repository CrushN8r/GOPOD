// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_tour.go
package wirepod_ttr

// The system window tour (2026-10-29, operator: "system menu demo = Go through
// it as the Torah * Goverlord's Torah: In the beginning there was a great
// restart. * Scripture 1, Restart and check * system-restart, a forced restart
// ... * All PASS? Go to scripture 2 to play, or 3 to edit."; "Yes to: a recorded
// tour of the system window: every menu item played in order, the way the admin
// run plays the game, with the restart left out"). system-tour goes through the
// GOPOD Torah in the chat, Goverlord saying it: the opening, then each scripture
// ("Scripture 1, Restart and check"), each commandment ("system-status, Wire-Pod,
// ..."), and the scripture's next line. A commandment on system_tour.run is run
// too and its answer shown; the rest - the restart, anything that moves a robot,
// changes a switch, ends the session, or waits for your words - is only said.
// system-tour again stops it. Record Session keeps it.

import (
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

var gopodTourRunning, gopodTourStop atomic.Bool

// gopodSkippedAt (2026-10-29, operator: "Please ensure Esc, stop run & skip all
// work globally as intended"): when the page last skipped (unix ms) - a stretch
// of lines being voiced stops there (gopodSayLinesInBubble).
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
	scriptures := gopodTorahScriptures()
	gopodTourRunning.Store(true)
	gopodTourStop.Store(false)
	run := map[string]bool{}
	for _, c := range gopodBrobotsSettings.SystemTour.Run {
		run[strings.ToLower(strings.TrimSpace(c))] = true
	}
	go func() {
		defer gopodTourRunning.Store(false)
		// 2026-10-29: the tour is a turn of the master chat queue; a question
		// asked meanwhile waits for its end (gopod_chat_queue.go).
		defer gopodTurnBeginOwn("tour", true)()
		say := func(text string) bool {
			if gopodTourStop.Load() || gopodRunActive() {
				return false
			}
			// The page plays each line before the next - when a page is here to play it.
			here := gopodPageHere()
			gopodReporterSay(role, text, true, here)
			if here {
				gopodCommCheckWaitQuiet(gopodSecondsToDuration(gopodBrobotsSettings.SystemTour.QuietSeconds), 60*time.Second)
			}
			return true
		}
		if !say(gopodBrobotsSettings.Torah.Opening) {
			return
		}
		for _, sc := range scriptures {
			if gopodTourStop.Load() || gopodRunActive() {
				return
			}
			// 2026-10-29 (operator: "each scripture in its own bubble * a brief
			// paragraph and examples"): the title, its paragraph (the group's
			// guide), its commandments, its next line - one bubble, a voice a line.
			// 2026-10-29 (operator: "Scripture tours >> human friendly easy to
			// read paragraph"): the title, then its plain paragraphs (the group's
			// tour), a voice a paragraph; a group with none falls back to the list.
			lines := []string{sc.Title}
			if len(sc.Tour) > 0 {
				lines = append(lines, sc.Tour...)
			} else {
				if sc.Guide != "" {
					lines = append(lines, sc.Guide)
				}
				lines = append(lines, sc.Commandments...)
				if sc.Next != "" {
					lines = append(lines, sc.Next)
				}
			}
			here := gopodPageHere()
			gopodSayLinesInBubble(role, lines, here)
			if here {
				gopodCommCheckWaitQuiet(gopodSecondsToDuration(gopodBrobotsSettings.SystemTour.QuietSeconds), 60*time.Second)
			}
			// Then the scripture's safe commandments run, each answer in its own bubble.
			for _, cmd := range sc.Cmds {
				if gopodTourStop.Load() {
					return
				}
				if !run[strings.ToLower(strings.TrimSpace(cmd))] {
					continue
				}
				// Sent as the keyboard sends it - system commands are the keyboard's.
				if err := gopodOperatorLine(gopodBrobot1ESN, cmd, "Typed text"); err != nil {
					logger.Println("GOPOD_TOUR " + cmd + ": " + err.Error())
				}
				if gopodPageHere() {
					gopodCommCheckWaitQuiet(gopodSecondsToDuration(gopodBrobotsSettings.SystemTour.QuietSeconds), 120*time.Second)
				}
			}
		}
		say(gopodBrobotsSettings.SystemTour.EndText)
		logger.Println("GOPOD_TOUR done: " + strconv.Itoa(len(scriptures)) + " scriptures")
	}()
	return strings.ReplaceAll(gopodBrobotsSettings.SystemTour.StartText, "{n}", strconv.Itoa(len(scriptures)))
}

// gopodSayLinesInBubble (2026-10-29, operator: "rather than a long wait for a
// single audio wav file, can brobot speech be sent in manageable chunks/segments
// * send segment for processing, while the brobot speaks send next segment,
// etc..."): lines in one bubble of host's ("lines:", the page's system bubble
// template), each its own voice - every line's voice named at once (so each
// starts dormant, Generating... until made), the first made, then each next one
// made while the one before it plays. wait: on until the page has played each.
func gopodSayLinesInBubble(host string, lines []string, wait bool) {
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
	GopodChatFeedAdd(host, "lines:"+strings.Join(keep, "\n"))
	says := make([]string, len(keep))
	for i, l := range keep {
		says[i] = gopodReporterSays(l)
		said := says[i]
		if sr := gopodBrobotsSettings.SignOff.SystemRole; sr != "" && host == sr {
			said = gopodSaySwap(said, rv.SystemSayWords)
		}
		if rv.GeneratingMarker == "" {
			continue
		}
		if page, _, made := gopodReporterCached(host, said); made {
			GopodChatFeedAdd(rv.Role, rv.GeneratingMarker+"forlit:"+page+"|"+l)
		} else {
			_, page := gopodReporterPaths(host, said)
			GopodChatFeedAdd(rv.Role, rv.GeneratingMarker+"forwait:"+page+"|"+l)
		}
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
	go makeOne(0)
	started := time.Now().UnixMilli()
	for i, l := range keep {
		m := <-ready[i]
		// A skip since this bubble began: the rest of it isn't voiced (the page shows it whole).
		if gopodSkippedAt.Load() >= started {
			logger.Println("GOPOD_TOUR skipped: the rest of this bubble isn't voiced")
			return
		}
		if i+1 < len(keep) {
			go makeOne(i + 1) // the next one is made while this one plays
		}
		if !m.ok {
			logger.Println("GOPOD_TOUR no voice for: " + gopodCut(l, 60))
			continue
		}
		gopodVoicePlay(host, l, m.page, m.seconds, wait)
	}
}

// ## EOF

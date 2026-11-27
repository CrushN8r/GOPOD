// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_page_presence.go
package wirepod_ttr

import (
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// gopodPageSeenMS: when the page last said it's showing (unix ms; 0 = never).
var gopodPageSeenMS atomic.Int64

// gopodPageWatchOn: the watch is running (Wire-Pod's start; never in tests).
var gopodPageWatchOn atomic.Bool

func gopodPagePresentFor() time.Duration {
	if d := gopodSecondsToDuration(gopodBrobotsSettings.Session.PagePresentSeconds); d > 0 {
		return d
	}
	return 20 * time.Second
}

// gopodPageHere: the chat page is showing and has been heard from lately.
func gopodPageHere() bool {
	if !gopodPageWatchOn.Load() {
		return true
	}
	seen := gopodPageSeenMS.Load()
	return seen != 0 && !gopodPageHidden.Load() && time.Since(time.UnixMilli(seen)) < gopodPagePresentFor()
}

// gopodPageWatchStart: once a second, a page not heard from is away.
func gopodPageWatchStart() {
	if gopodPageWatchOn.Swap(true) {
		return
	}
	gopodPageHidden.Store(false)
	go func() {
		for {
			if !gopodPageHidden.Load() && !gopodPageHere() {
				logger.Println("GOPOD_PAGE away: not heard from")
				GopodPageShown(false)
			}
			time.Sleep(time.Second)
		}
	}()
}

// gopodOpenerAwaitPage: the opener waits here while the page is away, the System llm bubble showing
// awaits_text with its hourglass; nothing plays.
func gopodOpenerAwaitPage() {
	if gopodPageHere() {
		return
	}
	gopodOpenerAwaits(true)
	logger.Println("GOPOD_OPENER awaits the chat page")
	started := time.Now()
	for !gopodPageHere() {
		time.Sleep(500 * time.Millisecond)
	}
	logger.Println("GOPOD_OPENER the chat page is back after " + time.Since(started).Round(time.Second).String())
	gopodOpenerAwaits(false)
}

// gopodOpenerAwaits: the System llm bubble shows awaits_text (on) or its own message again (off).
var gopodOpenerAwaitsSaved string

func gopodOpenerAwaits(on bool) {
	if on == gopodLLMWarmupAwaits {
		return
	}
	if on {
		gopodOpenerAwaitsSaved = gopodLLMWarmupText
		gopodLLMWarmupText = gopodBrobotsSettings.Session.AwaitsText
	} else {
		gopodLLMWarmupText = gopodOpenerAwaitsSaved
	}
	gopodLLMWarmupAwaits = on
	gopodWriteLLMWarmupState(true)
}

// gopodLLMWarmupAwaits: the System llm bubble shows awaits_text.
var gopodLLMWarmupAwaits bool

// ## EOF

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_page_sync.go
package wirepod_ttr

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The engine conducts, the page draws (2026-10-18, operator: "Finish Step 3";
// the plan: "Wire-Pod decides order and timing. The page shows what it's told
// and answers shown, played and heard ... terminal and page run the same show,
// line for line"). Two links on top of the feed's "played" (gopod_chat_feed.go):
//
//   - A Brobot speaks once his bubble is on the page: his reply's line goes on
//     the feed, then he waits - up to page_sync.reply_wait_seconds - for the
//     page to show it (or play past it), so he never talks over a voice or a
//     bubble the page is still playing (an egg's congratulations, his brother's
//     words). A page that's away isn't waited for.
//   - The page tells how long it held a line before showing it (the baton, a
//     countdown, the warm-up); one held longer than page_sync.held_log_seconds
//     is written to the journal, so a page running behind shows in the terminal.

// gopodWaitReplyShown: esn's newest reply bubble is on the page (or the wait is over).
var gopodWaitReplyShown = func(esn string) {
	most := gopodSecondsToDuration(gopodBrobotsSettings.PageSync.ReplyWaitSeconds)
	if most <= 0 || !gopodPageHere() {
		return
	}
	gopodLiveCheck.Lock()
	id := gopodLiveCheck.lastID[esn]
	gopodLiveCheck.Unlock()
	if id == 0 {
		return
	}
	started := time.Now()
	for {
		gopodLiveCheck.Lock()
		_, waiting := gopodLiveCheck.open[id]
		gopodLiveCheck.Unlock()
		gopodChatFeed.Lock()
		playedPast := gopodPagePlayedTo >= id
		gopodChatFeed.Unlock()
		if !waiting || playedPast {
			if waited := time.Since(started); waited > 500*time.Millisecond {
				logger.Println(fmt.Sprintf("GOPOD_SYNC %s waited %.1f s for his bubble (%d) before speaking", gopodBrobotName(esn), waited.Seconds(), id))
			}
			return
		}
		if time.Since(started) >= most || !gopodPageHere() {
			logger.Println(fmt.Sprintf("GOPOD_SYNC %s speaking: his bubble (%d) isn't on the page after %.1f s - the page is behind", gopodBrobotName(esn), id, time.Since(started).Seconds()))
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// GopodPageHeld: the page held feed line id ms milliseconds before showing it.
func GopodPageHeld(id int, ms int64) {
	limit := gopodBrobotsSettings.PageSync.HeldLogSeconds
	if id <= 0 || limit <= 0 || float64(ms) < limit*1000 {
		return
	}
	logger.Println("GOPOD_PAGE held " + strconv.Itoa(id) + " for " + fmt.Sprintf("%.1f", float64(ms)/1000) + " s before showing it")
}

// ## EOF

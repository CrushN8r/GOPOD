// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_main_page.go
package wirepod_ttr

import (
	"net"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The main page (2026-10-23, FINISH_PLAN_001.md C1; GOPOD_APP_REVIEW_001.md:
// "One page owns the show"): Wire-Pod times a run by the page - its shown,
// played and heard reports. With two pages open, both would play every voice
// and both would answer. So one page is the main page: it plays the sound and
// its reports time the run. Every other page is a viewer: the same bubbles,
// silent, its reports ignored. The first page that can play claims it; it
// keeps it while it reports in (every few seconds); gone longer than
// page_sync.main_lease_seconds, the next page that can play takes over. A page
// that asks to be a viewer (the app's ?view) never claims it.

var gopodMainPage = struct {
	sync.Mutex
	id   string
	seen time.Time
}{}

func gopodMainLease() time.Duration {
	s := gopodBrobotsSettings.PageSync.MainLeaseSeconds
	if s <= 0 {
		s = 12
	}
	return time.Duration(s * float64(time.Second))
}

// GopodMainPageClaim: the page id reports in (canPlay: it's shown and can play
// sound, and isn't a viewer by choice). True when it's the main page. A page
// with no id (an older page) is treated as the main page, as before.
func GopodMainPageClaim(id string, canPlay bool) bool {
	if id == "" {
		return true
	}
	gopodMainPage.Lock()
	defer gopodMainPage.Unlock()
	now := time.Now()
	if gopodMainPage.id == id {
		gopodMainPage.seen = now
		return true
	}
	if canPlay && (gopodMainPage.id == "" || now.Sub(gopodMainPage.seen) > gopodMainLease()) {
		logger.Println("GOPOD_PAGE main page: " + id + " (was " + gopodMainPageName(gopodMainPage.id) + ")")
		gopodMainPage.id, gopodMainPage.seen = id, now
		return true
	}
	return false
}

func gopodMainPageName(id string) string {
	if id == "" {
		return "none"
	}
	return id
}

// GopodIsMainPage: this page's reports time the run.
func GopodIsMainPage(id string) bool {
	if id == "" {
		return true
	}
	gopodMainPage.Lock()
	defer gopodMainPage.Unlock()
	return gopodMainPage.id == "" || gopodMainPage.id == id
}

// GopodMainPlayed: how far the main page has played the feed (this start's
// boot, the last id) - a viewer behind it catches up (2026-10-23, operator: "timing
// between the phone & pc appear to be in sync then fall out of sync").
func GopodMainPlayed() (int64, int) {
	gopodChatFeed.Lock()
	defer gopodChatFeed.Unlock()
	return gopodChatFeed.boot, gopodPagePlayedTo
}

// GopodIsLocalMachine (2026-10-23, live: the phone became the main page after a
// restart - it was clicked first - and the PC went silent; operator: "Stay with
// phone = visitor"): a page opened on GOPOD's own machine - by localhost or by
// the machine's own network address. Only it is the admin, and only it can be
// the main page; any other device is a visitor (silent, no system window).
func GopodIsLocalMachine(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() {
		return true
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false
	}
	for _, a := range addrs {
		if n, ok := a.(*net.IPNet); ok && n.IP.Equal(ip) {
			return true
		}
	}
	return false
}

// ## EOF

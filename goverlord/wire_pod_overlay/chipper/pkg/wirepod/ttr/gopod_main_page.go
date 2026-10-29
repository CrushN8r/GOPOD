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
// silent, its reports ignored. The first admin page claims it; it keeps it
// while it reports in (every few seconds); gone longer than
// page_sync.main_lease_seconds, the next admin page takes over. A page that
// asks to be a viewer (the app's ?view) never claims it.
// 2026-10-29 (operator: "llm warmup is flaking out on me"; live 10:43, 11:11:
// after a restart the page reloads itself, can't play until it's clicked, and
// so came back a viewer - no main page, nothing shown, the opening never
// heard): an admin page that can't play yet still owns an empty show, and a
// page that can play takes the show from a main page that can't.

var gopodMainPage = struct {
	sync.Mutex
	id    string
	seen  time.Time
	shown bool
}{}

func gopodMainLease() time.Duration {
	s := gopodBrobotsSettings.PageSync.MainLeaseSeconds
	if s <= 0 {
		s = 12
	}
	return time.Duration(s * float64(time.Second))
}

// GopodMainPageClaim: the page id reports in (admin: a page on GOPOD's own
// machine, not a viewer by choice; shown: it's shown and can play sound). True
// when it's the main page. A page with no id (an older page) is treated as the
// main page, as before.
func GopodMainPageClaim(id string, admin, shown bool) bool {
	if id == "" {
		return true
	}
	gopodMainPage.Lock()
	defer gopodMainPage.Unlock()
	now := time.Now()
	if admin {
		gopodAdminPageSeen(id, now)
	}
	if gopodMainPage.id == id {
		gopodMainPage.seen, gopodMainPage.shown = now, shown
		return true
	}
	if admin && (gopodMainPage.id == "" || now.Sub(gopodMainPage.seen) > gopodMainLease() || (shown && !gopodMainPage.shown)) {
		logger.Println("GOPOD_PAGE main page: " + id + " (was " + gopodMainPageName(gopodMainPage.id) + ")")
		gopodMainPage.id, gopodMainPage.seen, gopodMainPage.shown = id, now, shown
		return true
	}
	return false
}

// One admin page (2026-10-29, operator: "I should only have one admin page
// open, correct? * add a check for one admin controller page * visitor pages
// are multiple"): every admin page reporting in is counted (seen within the
// lease); a second one is logged once, its page told (page_sync.second_admin_text),
// and system-status says how many are open. Visitor pages aren't counted.
var gopodAdminPages = map[string]time.Time{}

// gopodAdminPageSeen (gopodMainPage held): an admin page reported in.
func gopodAdminPageSeen(id string, now time.Time) {
	if _, known := gopodAdminPages[id]; !known && gopodMainPage.id != "" && gopodMainPage.id != id && now.Sub(gopodMainPage.seen) <= gopodMainLease() {
		logger.Println("GOPOD_PAGE second admin page: " + id + " (main page " + gopodMainPage.id + ") - one admin page at a time")
	}
	gopodAdminPages[id] = now
	for k, t := range gopodAdminPages {
		if now.Sub(t) > gopodMainLease() {
			delete(gopodAdminPages, k)
		}
	}
}

// GopodAdminPagesOpen: the admin pages reporting in now.
func GopodAdminPagesOpen() int {
	gopodMainPage.Lock()
	defer gopodMainPage.Unlock()
	n := 0
	for _, t := range gopodAdminPages {
		if time.Since(t) <= gopodMainLease() {
			n++
		}
	}
	return n
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

// GopodSecondAdminText: what a second admin page's text field says.
func GopodSecondAdminText() string { return gopodBrobotsSettings.PageSync.SecondAdminText }

// ## EOF

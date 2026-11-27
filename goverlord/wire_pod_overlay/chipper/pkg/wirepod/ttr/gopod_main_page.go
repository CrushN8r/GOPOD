// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_main_page.go
package wirepod_ttr

import (
	"net"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

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

// GopodMainPageClaim: the page id reports in (admin: a page on GOPOD's own machine, not a viewer by
// choice; shown: it's shown and can play sound).
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

// One admin page: every admin page reporting in is counted (seen within the lease); a second one is
// logged once, its page told (page_sync.second_admin_text), and system-status says how many are
// open.
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

// GopodIsLocalMachine: a page opened on GOPOD's own machine - by localhost or by the machine's own
// network address.
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

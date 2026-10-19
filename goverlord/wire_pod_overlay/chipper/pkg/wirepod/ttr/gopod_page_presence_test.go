// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_page_presence_test.go
package wirepod_ttr

import (
	"testing"
	"time"
)

// TestPagePresence (2026-10-17, operator: "Goverlord awaits your return..."): a
// page that said it's showing lately is here; hidden, or not heard from, it's away.
func TestPagePresence(t *testing.T) {
	savedOn, savedSeen, savedHidden := gopodPageWatchOn.Load(), gopodPageSeenMS.Load(), gopodPageHidden.Load()
	defer func() {
		gopodPageWatchOn.Store(savedOn)
		gopodPageSeenMS.Store(savedSeen)
		gopodPageHidden.Store(savedHidden)
	}()
	gopodPageWatchOn.Store(true)
	gopodPageHidden.Store(false)
	gopodPageSeenMS.Store(0)
	if gopodPageHere() {
		t.Error("never heard from, but here")
	}
	gopodPageSeenMS.Store(time.Now().UnixMilli())
	if !gopodPageHere() {
		t.Error("just heard from, but away")
	}
	gopodPageHidden.Store(true)
	if gopodPageHere() {
		t.Error("hidden, but here")
	}
	gopodPageHidden.Store(false)
	gopodPageSeenMS.Store(time.Now().Add(-time.Hour).UnixMilli())
	if gopodPageHere() {
		t.Error("not heard from for an hour, but here")
	}
}

// ## EOF

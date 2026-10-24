// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_display.go
package wirepod_ttr

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
)

// The display mirror (2026-10-24, operator: "the visitor page does no thinking *
// visitor page zombie follows regardless of a nuclear war going on next door *
// visitor page zombie follows GOPOD admin feed"): the main page sends what it
// shows - its newest bubbles as drawn and where it's scrolled - and every other
// page paints exactly that. Wire-Pod only holds the latest one, numbered.

var gopodDisplay = struct {
	sync.Mutex
	v    int
	body []byte // {"html":..,"anchor":..,"offset":..,"at_end":..}
}{}

func init() {
	http.HandleFunc("/api-sdk/gopod_display", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Only GOPOD's own (admin) page sends the display.
			if !GopodIsAdmin(r.RemoteAddr) {
				http.Error(w, "admin only", http.StatusForbidden)
				return
			}
			data, err := io.ReadAll(io.LimitReader(r.Body, 8<<20))
			var probe map[string]any
			if err != nil || json.Unmarshal(data, &probe) != nil {
				http.Error(w, "bad display", http.StatusBadRequest)
				return
			}
			gopodDisplay.Lock()
			gopodDisplay.v++
			probe["v"] = gopodDisplay.v
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			enc.SetEscapeHTML(false) // the page's own HTML, kept as is (smaller)
			enc.Encode(probe)
			gopodDisplay.body = buf.Bytes()
			gopodDisplay.Unlock()
			w.Write([]byte("ok"))
			return
		}
		since, _ := strconv.Atoi(r.FormValue("since"))
		gopodDisplay.Lock()
		v, body := gopodDisplay.v, gopodDisplay.body
		gopodDisplay.Unlock()
		if v <= since || body == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	})
}

// ## EOF

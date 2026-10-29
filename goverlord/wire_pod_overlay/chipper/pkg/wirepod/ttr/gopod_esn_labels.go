// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_esn_labels.go
package wirepod_ttr

// The robots' labels (2026-10-29, operator: "Instead of the actual ESN, use ESN01,
// ESN02 labels * my internal GOPOD can sort it out privately * publically all
// people see is ESN01, ESN02 labels"). Everything public - the code, the
// settings, the page, the docs - names a Brobot's robot by its label: ESN01 is
// Brobot 1's robot, ESN02 Brobot 2's. Who they are is worked out at Wire-Pod's
// start (GopodBrobotsSetup): the operator's own two come from a private file in
// his home (~/.gopod_esns.json, {"ESN01": "...", "ESN02": "..."}), never in the
// repo; with no file, the robots Wire-Pod knows, in order. Wire-Pod changes a
// label to the robot's serial at its door (gopodESNLabelsIn: every /api-sdk/ call
// and the camera stream), and what the page reads carries the label
// (gopod_speaking_now.json).

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

// The labels.
const (
	gopodESN1 = "ESN01"
	gopodESN2 = "ESN02"
)

// gopodPrivateESNsPath: the operator's own robots, private, in his home.
var gopodPrivateESNsPath = "~/.gopod_esns.json"

// gopodPrivateESNs: the operator's own two robots' serials ("" when there's no file).
func gopodPrivateESNs() (string, string) {
	data, err := os.ReadFile(gopodHomePath(gopodPrivateESNsPath))
	if err != nil {
		return "", ""
	}
	var m map[string]string
	if json.Unmarshal(data, &m) != nil {
		return "", ""
	}
	return strings.TrimSpace(m[gopodESN1]), strings.TrimSpace(m[gopodESN2])
}

// GopodESNReal: a label as its robot's serial now; anything else as it is.
func GopodESNReal(s string) string {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case gopodESN1:
		return gopodBrobot1ESN
	case gopodESN2:
		return gopodBrobot2ESN
	}
	return s
}

// GopodESNLabel: a Brobot's serial as its label; any other robot's as it is.
func GopodESNLabel(esn string) string {
	switch {
	case esn != "" && strings.EqualFold(esn, gopodBrobot1ESN):
		return gopodESN1
	case esn != "" && strings.EqualFold(esn, gopodBrobot2ESN):
		return gopodESN2
	}
	return esn
}

// gopodESNLabelsIn: at Wire-Pod's door, a label in a call's serial becomes the
// robot's serial - every /api-sdk/ call and the camera stream, from the page, a
// song or a script.
func gopodESNLabelsIn(r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/api-sdk/") && r.URL.Path != "/cam-stream" {
		return
	}
	q := r.URL.Query()
	if v := q.Get("serial"); v != "" && GopodESNReal(v) != v {
		q.Set("serial", GopodESNReal(v))
		r.URL.RawQuery = q.Encode()
	}
	if r.Method == http.MethodPost && strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		if r.ParseForm() == nil {
			if v := r.PostForm.Get("serial"); v != "" && GopodESNReal(v) != v {
				r.PostForm.Set("serial", GopodESNReal(v))
				r.Form.Set("serial", GopodESNReal(v))
			}
		}
	}
}

// GopodBrobotOrderSerials: Brobot 1's and Brobot 2's serials, for the admin page's
// battery icons (Wire-Pod's own page names each by its serial).
func GopodBrobotOrderSerials() []string {
	return []string{gopodBrobot1ESN, gopodBrobot2ESN}
}

// ## EOF

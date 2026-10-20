// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_usb.go
package wirepod_ttr

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
)

// The USB panel (2026-10-20, operator: "add in a dynamic USB panel of
// (hot-swappable) devices"): lsusb's devices, root hubs left out - the menu's
// live panel polls /api-sdk/gopod_usb and marks what was plugged in or pulled
// out; "system-usb" lists them once.

type gopodUSBDevice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Bus  string `json:"bus"`
}

var gopodUSBLineRE = regexp.MustCompile(`^Bus (\d+) Device \d+: ID ([0-9a-f]{4}:[0-9a-f]{4}) ?(.*)$`)

func gopodUSBDevices() []gopodUSBDevice {
	list, _ := gopodUSBDevicesErr()
	return list
}

// gopodUSBDevicesErr: the devices, or lsusb's error (review: a failed lsusb
// isn't "the mic is missing").
func gopodUSBDevicesErr() ([]gopodUSBDevice, error) {
	out, err := exec.Command("lsusb").Output()
	if err != nil {
		return nil, err
	}
	var list []gopodUSBDevice
	for _, l := range strings.Split(string(out), "\n") {
		m := gopodUSBLineRE.FindStringSubmatch(strings.TrimSpace(l))
		if m == nil || strings.HasPrefix(m[2], "1d6b:") { // root hubs
			continue
		}
		name := strings.TrimSpace(m[3])
		if name == "" {
			name = "(no name)"
		}
		list = append(list, gopodUSBDevice{ID: m[2], Name: name, Bus: m[1]})
	}
	return list, nil
}

func init() {
	http.HandleFunc("/api-sdk/gopod_usb", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(gopodUSBDevices())
	})
}

func gopodSystemUSB() string {
	list := gopodUSBDevices()
	out := []string{"USB devices (" + strconvItoa(len(list)) + "):"}
	for _, d := range list {
		out = append(out, "  "+d.ID+"  "+d.Name)
	}
	return gopodJoinLines(out)
}

func strconvItoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

// ## EOF

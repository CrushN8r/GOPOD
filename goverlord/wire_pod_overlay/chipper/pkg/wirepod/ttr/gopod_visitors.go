// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_visitors.go
package wirepod_ttr

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// Visitors and the chat sign-up (2026-10-24, operator: "Is it easier set me
// apart by my local PC & Laptop IP addresses? ... everyone else is a visitor ...
// allow a chat participant gathering * initially visitor chat participation is
// OFF"): the admin is GOPOD's own machine, or a device in the private admin
// address list (visitors.admin_addresses_file, outside the repo); everyone else
// is a visitor, one per device (its network address). A visitor's first line is
// their username request (first come, first served; a taken name gets a number:
// name1, name2); every line after it is dropped until the admin approves it in
// the system window. With participation ON, an approved visitor with their box
// ticked talks to the Brobots like a typed line, under their username; unticked
// is muted. A dropped line tells the visitor why: "Waiting for approval" or
// "Muted". A visitor never runs the show: commands (start, stop, thanks,
// system-...) are dropped.

type gopodVisitor struct {
	Key      string `json:"key"`
	Addr     string `json:"addr"`
	Device   string `json:"device"`
	Name     string `json:"name"`
	Approved bool   `json:"approved"`
	Allowed  bool   `json:"allowed"`
	first    time.Time
	seen     time.Time
}

var gopodVisitors = struct {
	sync.Mutex
	byAddr map[string]*gopodVisitor // by the device's key (its browser ID, else its address)
	on     bool
	set    bool
	loaded bool
}{byAddr: map[string]*gopodVisitor{}}

// Saved profiles (2026-10-24, operator: "Save the visitor profile in case of a
// disconnect reconnect, like the brobot reboot recovery thing"): each visitor's
// username, approval and mute box are kept in visitors.profiles_file (private,
// outside the repo) and come back after a dropped connection or a Wire-Pod
// restart - by the phone's own browser ID, so a new network address after a
// reconnect is still the same visitor. Participation itself starts OFF.
func gopodVisitorsFile() string {
	return gopodHomePath(gopodBrobotsSettings.Visitors.ProfilesFile)
}

// gopodVisitorsLoad (locked): the saved profiles, once.
func gopodVisitorsLoad() {
	if gopodVisitors.loaded {
		return
	}
	gopodVisitors.loaded = true
	f := gopodVisitorsFile()
	if f == "" {
		return
	}
	data, err := os.ReadFile(f)
	if err != nil {
		return
	}
	var saved []gopodVisitor
	if json.Unmarshal(data, &saved) != nil {
		return
	}
	for i := range saved {
		v := saved[i]
		if v.Key == "" || v.Name == "" {
			continue
		}
		v.first, v.seen = time.Now(), time.Time{}
		gopodVisitors.byAddr[v.Key] = &v
	}
	logger.Println("GOPOD_VISITOR " + strconv.Itoa(len(gopodVisitors.byAddr)) + " saved visitor profiles back")
}

// gopodVisitorsSave (locked): every visitor with a name, to the profiles file.
func gopodVisitorsSave() {
	f := gopodVisitorsFile()
	if f == "" {
		return
	}
	var keep []gopodVisitor
	for _, v := range gopodVisitors.byAddr {
		if v.Name != "" {
			keep = append(keep, *v)
		}
	}
	if data, err := json.MarshalIndent(keep, "", "  "); err == nil {
		// Private: made readable by the operator only, owned by the operator.
		if _, err := os.Stat(f); err != nil {
			if os.WriteFile(f, nil, 0o600) == nil {
				if _, uid, gid := gopodRecordRoot(); uid > 0 {
					os.Chown(f, uid, gid)
				}
			}
		}
		gopodWriteFileSafe(f, data)
	}
}

const gopodVisitorHowPrefix = "Visitor: "

// gopodVisitorAddr: the device's network address (no port).
func gopodVisitorAddr(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

// gopodAdminAddresses: the private admin list (one address a line, # notes).
func gopodAdminAddresses() []string {
	// 2026-10-24 (live: the PC stayed a visitor - Wire-Pod runs as root, so "~"
	// was root's home): "~/" is the operator's home (gopodHomePath).
	f := gopodHomePath(gopodBrobotsSettings.Visitors.AdminAddressesFile)
	if f == "" {
		return nil
	}
	data, err := os.ReadFile(f)
	if err != nil {
		return nil
	}
	var out []string
	for _, l := range strings.Split(string(data), "\n") {
		if i := strings.Index(l, "#"); i >= 0 {
			l = l[:i]
		}
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

// GopodIsAdmin: GOPOD's own machine, or a device in the private admin list.
func GopodIsAdmin(remoteAddr string) bool {
	if GopodIsLocalMachine(remoteAddr) {
		return true
	}
	ip := net.ParseIP(gopodVisitorAddr(remoteAddr))
	if ip == nil {
		return false
	}
	for _, a := range gopodAdminAddresses() {
		if other := net.ParseIP(a); other != nil && other.Equal(ip) {
			return true
		}
	}
	return false
}

// gopodDeviceKind: "Android phone · Chrome" from the browser's own description.
func gopodDeviceKind(ua string) string {
	has := func(s string) bool { return strings.Contains(ua, s) }
	osName, kind := "", "computer"
	switch {
	case has("iPad"):
		osName, kind = "iPad", "tablet"
	case has("iPhone"):
		osName, kind = "iPhone", "phone"
	case has("Android"):
		osName, kind = "Android", "tablet"
		if has("Mobile") {
			kind = "phone"
		}
	case has("CrOS"):
		osName = "Chromebook"
	case has("Windows"):
		osName = "Windows"
	case has("Mac OS X"), has("Macintosh"):
		osName = "Mac"
	case has("Linux"):
		osName = "Linux"
	}
	browser := ""
	switch {
	case has("SamsungBrowser"):
		browser = "Samsung Internet"
	case has("Edg"):
		browser = "Edge"
	case has("Firefox"), has("FxiOS"):
		browser = "Firefox"
	case has("CriOS"), has("Chrome"):
		browser = "Chrome"
	case has("Safari"):
		browser = "Safari"
	}
	out := strings.TrimSpace(osName + " " + kind)
	if osName == "iPhone" || osName == "iPad" {
		out = osName
	}
	if browser != "" {
		out += " · " + browser
	}
	return out
}

// gopodVisitorOn: participation ON (starts as visitors.participation, OFF).
func gopodVisitorOn() bool {
	if !gopodVisitors.set {
		return gopodBrobotsSettings.Visitors.Participation
	}
	return gopodVisitors.on
}

// gopodVisitorKey: the device's key - its browser ID, else its address.
func gopodVisitorKey(remoteAddr, id string) string {
	if id = strings.TrimSpace(id); id != "" && len(id) <= 64 {
		return "id:" + id
	}
	return gopodVisitorAddr(remoteAddr)
}

// gopodVisitorGet (locked): the visitor for this device (made on first sight;
// a saved profile comes back).
func gopodVisitorGet(remoteAddr, id, ua string) *gopodVisitor {
	gopodVisitorsLoad()
	addr := gopodVisitorAddr(remoteAddr)
	key := gopodVisitorKey(remoteAddr, id)
	v := gopodVisitors.byAddr[key]
	if v == nil {
		v = &gopodVisitor{Key: key, Addr: addr, first: time.Now()}
		gopodVisitors.byAddr[key] = v
		logger.Println("GOPOD_VISITOR new device " + addr + " (" + gopodDeviceKind(ua) + ")")
	} else if v.seen.IsZero() && v.Name != "" {
		logger.Println("GOPOD_VISITOR " + v.Name + " is back (" + addr + ")")
	}
	v.Addr = addr
	if ua != "" {
		v.Device = gopodDeviceKind(ua)
	}
	v.seen = time.Now()
	return v
}

// GopodVisitorSeen: a visitor's page reports in.
func GopodVisitorSeen(remoteAddr, id, ua string) {
	if GopodIsAdmin(remoteAddr) {
		return
	}
	gopodVisitors.Lock()
	gopodVisitorGet(remoteAddr, id, ua)
	gopodVisitors.Unlock()
}

// gopodVisitorCleanName: a typed username, tidied - letters, digits, spaces,
// - _ . ' only, at most visitors.name_max letters.
func gopodVisitorCleanName(text string) string {
	var b strings.Builder
	for _, r := range strings.Join(strings.Fields(text), " ") {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune(" -_.'", r) {
			b.WriteRune(r)
		}
	}
	name := strings.TrimSpace(b.String())
	if max := gopodBrobotsSettings.Visitors.NameMax; max > 0 && len([]rune(name)) > max {
		name = strings.TrimSpace(string([]rune(name)[:max]))
	}
	return name
}

// gopodVisitorUniqueName: first come, first served - a taken name (another
// visitor's, or one of the show's own: Operator, Goverlord, Brobot 1...) gets a
// number: name1, name2.
func gopodVisitorUniqueName(name, self string) string {
	// self: the asking visitor's key.
	taken := map[string]bool{}
	for _, n := range []string{"Operator", "Goverlord", "Brobots", "Brobot 1", "Brobot 2", "Brobot 3", "Brobot 4", gopodBrobotsSettings.SignOff.SystemRole} {
		taken[strings.ToLower(n)] = true
	}
	for addr, v := range gopodVisitors.byAddr {
		if addr != self && v.Name != "" {
			taken[strings.ToLower(v.Name)] = true
		}
	}
	if !taken[strings.ToLower(name)] {
		return name
	}
	for i := 1; ; i++ {
		if c := name + strconv.Itoa(i); !taken[strings.ToLower(c)] {
			return c
		}
	}
}

// gopodVisitorCommand: a line that would run the show - never a visitor's.
func gopodVisitorCommand(text string) bool {
	return gopodSystemSetMatch(text) || gopodScriptCommandMatch(text) || gopodStopRunMatch(text) ||
		gopodStopSessionMatch(text) || gopodReplayMatch(text) || gopodIsThanks(text) || gopodTypedCommandMatch(text)
}

// GopodVisitorLine: a visitor's typed line. pass: it goes to the Brobots as
// their line (how names them); else note says why it was dropped.
func GopodVisitorLine(remoteAddr, id, ua, text string) (pass bool, how, note string) {
	vs := gopodBrobotsSettings.Visitors
	gopodVisitors.Lock()
	defer gopodVisitors.Unlock()
	v := gopodVisitorGet(remoteAddr, id, ua)
	if v.Name == "" {
		if name := gopodVisitorCleanName(text); name != "" {
			v.Name, v.Approved, v.Allowed = gopodVisitorUniqueName(name, v.Key), false, false
			logger.Println("GOPOD_VISITOR " + v.Addr + " asks for the name " + v.Name)
			gopodVisitorsSave()
		}
		return false, "", vs.WaitingText
	}
	if !v.Approved {
		return false, "", vs.WaitingText
	}
	if !gopodVisitorOn() || !v.Allowed || gopodVisitorCommand(text) {
		logger.Println("GOPOD_VISITOR " + v.Name + " (" + v.Addr + "): dropped (muted): " + text)
		return false, "", vs.MutedText
	}
	return true, gopodVisitorHowPrefix + v.Name, ""
}

// gopodVisitorList: the system window's list - every visitor seen lately, or
// with a name - first seen first.
func gopodVisitorList() []gopodVisitor {
	gopodVisitors.Lock()
	defer gopodVisitors.Unlock()
	gopodVisitorsLoad()
	seen := time.Duration(gopodBrobotsSettings.Visitors.SeenSeconds * float64(time.Second))
	var out []gopodVisitor
	var firsts []time.Time
	for _, v := range gopodVisitors.byAddr {
		if v.Name == "" && time.Since(v.seen) > seen {
			continue
		}
		out = append(out, *v)
		firsts = append(firsts, v.first)
	}
	idx := make([]int, len(out))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool { return firsts[idx[a]].Before(firsts[idx[b]]) })
	sorted := make([]gopodVisitor, len(out))
	for i, j := range idx {
		sorted[i] = out[j]
	}
	return sorted
}

// gopodVisitorAct: the admin's approve / clear / allow / participation; key is
// the visitor's (from the list).
func gopodVisitorAct(action, addr string, on bool) bool {
	gopodVisitors.Lock()
	defer gopodVisitors.Unlock()
	gopodVisitorsLoad()
	if action == "participation" {
		gopodVisitors.on, gopodVisitors.set = on, true
		logger.Println("GOPOD_VISITOR participation " + map[bool]string{true: "ON", false: "OFF"}[on])
		return true
	}
	v := gopodVisitors.byAddr[addr]
	if v == nil {
		return false
	}
	switch action {
	case "approve":
		if v.Name == "" {
			return false
		}
		v.Approved, v.Allowed = true, true
	case "clear":
		v.Name, v.Approved, v.Allowed = "", false, false
	case "allow":
		v.Allowed = on
	default:
		return false
	}
	logger.Println("GOPOD_VISITOR " + action + " " + v.Addr + " (" + v.Name + ")")
	gopodVisitorsSave()
	return true
}

func init() {
	// The system window's visitors slot (the admin's only).
	http.HandleFunc("/api-sdk/gopod_visitors", func(w http.ResponseWriter, r *http.Request) {
		if !GopodIsAdmin(r.RemoteAddr) {
			http.Error(w, "admin only", http.StatusForbidden)
			return
		}
		if r.Method == http.MethodPost {
			if !gopodVisitorAct(r.FormValue("action"), r.FormValue("key"), r.FormValue("on") == "1") {
				http.Error(w, "no such visitor or action", http.StatusBadRequest)
				return
			}
		}
		gopodVisitors.Lock()
		on := gopodVisitorOn()
		gopodVisitors.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"participation": on, "visitors": gopodVisitorList()})
	})
}

// ## EOF

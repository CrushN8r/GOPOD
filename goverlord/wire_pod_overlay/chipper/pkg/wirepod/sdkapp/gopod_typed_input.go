// goverlord/wire_pod_overlay/chipper/pkg/wirepod/sdkapp/gopod_typed_input.go
package sdkapp

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	wirepod_ttr "github.com/kercre123/wire-pod/chipper/pkg/wirepod/ttr"
)

// gopodTypedInputHandler: POST serial=<ESN>&text=<line> from the Brobots chat
// page's typed input line (wirepod_ttr.GopodTypedInput).
func gopodTypedInputHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	// 2026-10-09: voice=1 from a voice command's script (the robot heard it; the
	// page already shows what it heard).
	// 2026-10-22: the page's All 1/2 menu - who answers first.
	wirepod_ttr.GopodAllFirstPick(r.FormValue("text"), r.FormValue("first"))
	// 2026-10-24: a visitor's line goes through the chat sign-up (gopod_visitors.go):
	// passed under their username, or dropped with why ("Waiting for approval", "Muted").
	if !wirepod_ttr.GopodIsAdmin(r.RemoteAddr) {
		pass, how, note := wirepod_ttr.GopodVisitorLine(r.RemoteAddr, r.FormValue("visitor"), r.UserAgent(), r.FormValue("text"))
		w.Header().Set("Content-Type", "application/json")
		if !pass {
			json.NewEncoder(w).Encode(map[string]string{"note": note})
			return
		}
		if err := wirepod_ttr.GopodVisitorTypedInput(r.FormValue("serial"), r.FormValue("text"), how); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Write([]byte(`{"ok":true}`))
		return
	}
	input := wirepod_ttr.GopodTypedInput
	if r.FormValue("voice") == "1" {
		input = wirepod_ttr.GopodVoiceCommandInput
	}
	if err := input(r.FormValue("serial"), r.FormValue("text")); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write([]byte("ok"))
}

// gopodRecordSessionHandler: POST on=1|0 from the chat page's Record Session box
// (wirepod_ttr.GopodRecordSessionSet).
func gopodRecordSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	wirepod_ttr.GopodRecordSessionSet(r.FormValue("on") == "1")
	w.Write([]byte("ok"))
}

// gopodOpeningHeardHandler: POST - the chat page has played the reporter hosts'
// opening; Wire-Pod then says LLM Ready (wirepod_ttr.GopodOpeningHeard).
func gopodOpeningHeardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	// 2026-10-23 (C1): only the main page's reports time the run.
	if !wirepod_ttr.GopodIsMainPage(r.FormValue("page")) {
		w.Write([]byte("viewer"))
		return
	}
	wirepod_ttr.GopodOpeningHeard()
	w.Write([]byte("ok"))
}

// gopodSystemSayHandler: POST text, wait_seconds - a song's system line, its
// bubble shown with its voice; answers "heard" once the page has played it, or
// "not_heard" after wait_seconds (wirepod_ttr.GopodSystemSayHeard). With no
// wait_seconds it answers "queued" at once.
func gopodSystemSayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	wait, _ := strconv.ParseFloat(r.FormValue("wait_seconds"), 64)
	// 2026-10-13: a song's warm-up is ready - its bubble turns ready in the
	// chat's order, then Goverlord says the ready line (answered once said).
	// 2026-10-13: a song's warm-up message, in the chat's order.
	if ms := r.FormValue("warmup_text"); ms != "" {
		started, _ := strconv.ParseInt(ms, 10, 64)
		wirepod_ttr.GopodSongWarmupText(started, r.FormValue("text"))
		w.Write([]byte("queued"))
		return
	}
	if ms := r.FormValue("warmup_ready"); ms != "" {
		started, _ := strconv.ParseInt(ms, 10, 64)
		// 2026-10-17 (survey E4): the truth - heard, or not heard.
		if wirepod_ttr.GopodSongWarmupReady(started, r.FormValue("text")) {
			w.Write([]byte("heard"))
		} else {
			w.Write([]byte("not_heard"))
		}
		return
	}
	// 2026-10-11: who = a reporter (Brobot 3/4) - his line, bubble and voice,
	// answered once it has played (a song's warmup_lines).
	if who := r.FormValue("who"); who != "" && who != "Goverlord" {
		if wirepod_ttr.GopodReporterSay(who, r.FormValue("text")) {
			w.Write([]byte("heard"))
		} else {
			w.Write([]byte("not_heard"))
		}
		return
	}
	if wait <= 0 {
		go wirepod_ttr.GopodSystemSayHeard(r.FormValue("text"), 0)
		w.Write([]byte("queued"))
		return
	}
	if wirepod_ttr.GopodSystemSayHeard(r.FormValue("text"), time.Duration(wait*float64(time.Second))) {
		w.Write([]byte("heard"))
		return
	}
	w.Write([]byte("not_heard"))
}

// gopodRunHandler: POST action=start|end|status - a run checking in or out with
// Goverlord (gopod_session.go): start name kind pid -> {"ok", "busy"}; end name
// reason; status -> {"phase", "run"}.
func gopodRunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	reply := map[string]any{"ok": true}
	switch r.FormValue("action") {
	case "start":
		pid, _ := strconv.Atoi(r.FormValue("pid"))
		ok, busy := wirepod_ttr.GopodRunStart(r.FormValue("name"), r.FormValue("kind"), pid)
		reply["ok"], reply["busy"] = ok, busy
	case "end":
		wirepod_ttr.GopodRunEnd(r.FormValue("name"), r.FormValue("reason"))
	default:
		reply["phase"] = wirepod_ttr.GopodSessionPhase()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reply)
}

// gopodNativeHandler: GET -> {"on"}; POST on=1|0 - the chat page's Native? box
// (wirepod_ttr.GopodNativeSet, gopod_native.go).
func gopodNativeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		wirepod_ttr.GopodNativeSet(r.FormValue("on") == "1")
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"on": wirepod_ttr.GopodNativeOn()})
}

// gopodVoiceHeardHandler: POST token - the page reached that heard marker.
func gopodVoiceHeardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	// 2026-10-17 (survey E3: "heard means passed, not played"): skipped = the
	// voices before it that didn't play (missing, blocked).
	if !wirepod_ttr.GopodIsMainPage(r.FormValue("page")) {
		w.Write([]byte("viewer"))
		return
	}
	wirepod_ttr.GopodVoiceHeardSkipped(r.FormValue("token"), r.FormValue("skipped"))
	w.Write([]byte("ok"))
}

// gopodPagePlayedHandler: POST boot, id - the page has played the feed up to id
// (2026-10-17, operator: "the web page and terminal must have some sort of
// coordinated dialogue"; wirepod_ttr.GopodPagePlayed).
func gopodPagePlayedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	boot, _ := strconv.ParseInt(r.FormValue("boot"), 10, 64)
	id, _ := strconv.Atoi(r.FormValue("id"))
	if !wirepod_ttr.GopodIsMainPage(r.FormValue("page")) {
		logger.Println("GOPOD_PAGE played " + strconv.Itoa(id) + " not taken: page " + r.FormValue("page") + " isn't the main page")
		w.Write([]byte("viewer"))
		return
	}
	wirepod_ttr.GopodPagePlayed(boot, id)
	// 2026-10-18 (step 3): the longest the page held a line before showing it.
	if heldID, _ := strconv.Atoi(r.FormValue("held_id")); heldID > 0 {
		heldMS, _ := strconv.ParseInt(r.FormValue("held_ms"), 10, 64)
		wirepod_ttr.GopodPageHeld(heldID, heldMS)
	}
	w.Write([]byte("ok"))
}

// gopodPageShownHandler: POST shown=1|0 - the chat page is showing, or hidden
// (another tab, minimised); a playing draft waits while it's hidden
// (wirepod_ttr.GopodPageShown).
func gopodPageShownHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	// 2026-10-23 (C1): every page reports in; the main page's shown state is the
	// run's (a viewer hidden pauses nothing). The answer tells the page its role.
	shown := r.FormValue("shown") != "0"
	// Only a page on GOPOD's own machine is the admin and can be the main page;
	// a phone is a visitor - silent, no system window.
	admin := wirepod_ttr.GopodIsAdmin(r.RemoteAddr) && r.FormValue("viewer") != "1"
	if !admin {
		wirepod_ttr.GopodVisitorSeen(r.RemoteAddr, r.FormValue("visitor"), r.UserAgent())
	}
	main := wirepod_ttr.GopodMainPageClaim(r.FormValue("page"), shown && admin)
	if main {
		wirepod_ttr.GopodPageShown(shown)
	}
	w.Header().Set("Content-Type", "application/json")
	reply := map[string]any{"main": main, "admin": admin}
	// A visitor's page hears where it stands (waiting, approved, muted...).
	if !admin {
		reply["visitor"] = wirepod_ttr.GopodVisitorStatus(r.RemoteAddr, r.FormValue("visitor"), r.UserAgent())
	}
	json.NewEncoder(w).Encode(reply)
}

// gopodPageNoteHandler: POST note=... - a click on the chat page, written to
// Wire-Pod's log as GOPOD_PAGE (operator 2026-10-08: "add to log numpad
// checkbox clicks").
func gopodPageNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	note := strings.TrimSpace(r.FormValue("note"))
	if len(note) > 200 {
		note = note[:200]
	}
	if note != "" {
		logger.Println("GOPOD_PAGE " + note)
		// 2026-10-17: a live reply's bubble accounted for (wirepod_ttr.GopodLiveReplyNoted).
		if rest, ok := strings.CutPrefix(note, "live reply "); ok {
			if id, err := strconv.Atoi(strings.Fields(rest + " ")[0]); err == nil {
				wirepod_ttr.GopodLiveReplyNoted(id)
			}
		}
	}
	w.Write([]byte("ok"))
}

// gopodNumpadMapHandler: GET - the golden numpad persona map for the chat page's
// push-to-talk (wirepod_ttr.GopodNumpadMapJSON).
func gopodNumpadMapHandler(w http.ResponseWriter, r *http.Request) {
	data, err := wirepod_ttr.GopodNumpadMapJSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// gopodVoiceMapHandler (2026-10-26, the numpad-to-voices map,
// gopod_voice_map.go): GET the rows and the voices; POST key=KP7&voice=af_bella
// sets one.
func gopodVoiceMapHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		msg, err := wirepod_ttr.GopodVoiceMapSet(r.FormValue("key"), r.FormValue("voice"))
		out := map[string]any{"ok": err == nil, "message": msg}
		if err != nil {
			out["error"] = err.Error()
		}
		json.NewEncoder(w).Encode(out)
		return
	}
	json.NewEncoder(w).Encode(wirepod_ttr.GopodVoiceMap())
}

// gopodNumpadPTTHandler: POST action=start|stop key=KP1 from the chat page's
// numpad push-to-talk; the reply is the mic service's JSON (status, text, ...).
func gopodNumpadPTTHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var out map[string]any
	// 2026-10-24: the page's All 1/2 pick comes with every press - it stands for
	// the spoken line too (its words aren't known yet).
	if first := r.FormValue("first"); first != "" {
		wirepod_ttr.GopodAllFirstPick("", first)
	}
	switch r.FormValue("action") {
	case "start":
		out = wirepod_ttr.GopodNumpadPTTStart(r.FormValue("key"))
	case "stop":
		out = wirepod_ttr.GopodNumpadPTTStop(r.FormValue("key"))
	default:
		http.Error(w, "action must be start or stop", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// ## EOF

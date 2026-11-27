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

// gopodTypedInputHandler: POST serial=<ESN>&text=<line> from the Brobots chat page's typed input
// line (wirepod_ttr.GopodTypedInput).
func gopodTypedInputHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	wirepod_ttr.GopodAllFirstPick(r.FormValue("text"), r.FormValue("first"))
	logger.Println("GOPOD_PICK typed to=" + wirepod_ttr.GopodESNLabel(r.FormValue("serial")))
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

// gopodOpeningHeardHandler: POST - the chat page has played the reporter hosts' opening; Wire-Pod
// then says LLM Ready (wirepod_ttr.GopodOpeningHeard).
func gopodOpeningHeardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	if !wirepod_ttr.GopodIsMainPage(r.FormValue("page")) {
		w.Write([]byte("viewer"))
		return
	}
	wirepod_ttr.GopodOpeningHeard()
	w.Write([]byte("ok"))
}

// gopodSystemSayHandler: POST text, wait_seconds - a song's system line, its bubble shown with its
// voice; answers "heard" once the page has played it, or "not_heard" after wait_seconds
// (wirepod_ttr.GopodSystemSayHeard).
func gopodSystemSayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	wait, _ := strconv.ParseFloat(r.FormValue("wait_seconds"), 64)
	if ms := r.FormValue("warmup_text"); ms != "" {
		started, _ := strconv.ParseInt(ms, 10, 64)
		wirepod_ttr.GopodSongWarmupText(started, r.FormValue("text"))
		w.Write([]byte("queued"))
		return
	}
	if ms := r.FormValue("warmup_ready"); ms != "" {
		started, _ := strconv.ParseInt(ms, 10, 64)
		if wirepod_ttr.GopodSongWarmupReady(started, r.FormValue("text")) {
			w.Write([]byte("heard"))
		} else {
			w.Write([]byte("not_heard"))
		}
		return
	}
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

// gopodRunHandler: POST action=start|end|status - a run checking in or out with Goverlord
// (gopod_session.go): start name kind pid -> {"ok", "busy"}; end name reason; status -> {"phase",
// "run"}.
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
	case "anim_wait":
		reply["waited_ms"] = wirepod_ttr.GopodAnimationWaitDone(r.FormValue("serial"))
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
	if !wirepod_ttr.GopodIsMainPage(r.FormValue("page")) {
		w.Write([]byte("viewer"))
		return
	}
	wirepod_ttr.GopodVoiceHeardSkipped(r.FormValue("token"), r.FormValue("skipped"))
	w.Write([]byte("ok"))
}

// gopodPagePlayedHandler: POST boot, id - the page has played the feed up to id.
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
	if heldID, _ := strconv.Atoi(r.FormValue("held_id")); heldID > 0 {
		heldMS, _ := strconv.ParseInt(r.FormValue("held_ms"), 10, 64)
		wirepod_ttr.GopodPageHeld(heldID, heldMS)
	}
	w.Write([]byte("ok"))
}

// gopodPageShownHandler: POST shown=1|0 - the chat page is showing, or hidden (another tab,
// minimised); a playing draft waits while it's hidden (wirepod_ttr.GopodPageShown).
func gopodPageShownHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	shown := r.FormValue("shown") != "0"
	admin := wirepod_ttr.GopodIsAdmin(r.RemoteAddr) && r.FormValue("viewer") != "1"
	if !admin {
		wirepod_ttr.GopodVisitorSeen(r.RemoteAddr, r.FormValue("visitor"), r.UserAgent())
	}
	main := wirepod_ttr.GopodMainPageClaim(r.FormValue("page"), admin, shown)
	if main {
		wirepod_ttr.GopodPageShown(shown)
	}
	w.Header().Set("Content-Type", "application/json")
	reply := map[string]any{"main": main, "admin": admin}
	if admin && !main {
		reply["second_admin"] = wirepod_ttr.GopodSecondAdminText()
	}
	if !admin {
		reply["visitor"] = wirepod_ttr.GopodVisitorStatus(r.RemoteAddr, r.FormValue("visitor"), r.UserAgent())
	}
	json.NewEncoder(w).Encode(reply)
}

// gopodEscapeHandler: the page's Esc - whatever is running stops (wirepod_ttr.GopodEscape); with
// nothing running, the page skips instead.
func gopodEscapeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !wirepod_ttr.GopodIsAdmin(r.RemoteAddr) {
		http.Error(w, "admin POST only", http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"stopped": wirepod_ttr.GopodEscape()})
}

// gopodBrobotOrderHandler: Brobot 1's and Brobot 2's serials, in order, for the admin page's
// battery icons (Wire-Pod's own page names each by its serial).
func gopodBrobotOrderHandler(w http.ResponseWriter, r *http.Request) {
	if !wirepod_ttr.GopodIsAdmin(r.RemoteAddr) {
		http.Error(w, "admin only", http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wirepod_ttr.GopodBrobotOrderSerials())
}

// gopodPageNoteHandler: POST note=...
func gopodPageNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	note := strings.TrimSpace(r.FormValue("note"))
	if len(note) > 400 {
		note = note[:400]
	}
	if note != "" {
		logger.Println("GOPOD_PAGE " + note)
		if strings.HasPrefix(note, "skipped (") {
			wirepod_ttr.GopodSkipped()
		}
		if rest, ok := strings.CutPrefix(note, "live reply "); ok {
			if id, err := strconv.Atoi(strings.Fields(rest + " ")[0]); err == nil {
				wirepod_ttr.GopodLiveReplyNoted(id)
			}
		}
	}
	w.Write([]byte("ok"))
}

// gopodNumpadMapHandler: GET - the golden numpad persona map for the chat page's push-to-talk
// (wirepod_ttr.GopodNumpadMapJSON).
func gopodNumpadMapHandler(w http.ResponseWriter, r *http.Request) {
	data, err := wirepod_ttr.GopodNumpadMapJSONFor(r.FormValue("layout"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// gopodVoiceMapHandler: GET the rows and the voices; POST key=KP7&voice=af_bella sets one.
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

// gopodNumpadPTTHandler: POST action=start|stop key=KP1 from the chat page's numpad push-to-talk;
// the reply is the mic service's JSON (status, text, ...).
func gopodNumpadPTTHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var out map[string]any
	if first := r.FormValue("first"); first != "" {
		wirepod_ttr.GopodAllFirstPick("", first)
	}
	wirepod_ttr.GopodNumpadLayoutPick(r.FormValue("layout"))
	logger.Println("GOPOD_PICK numpad " + r.FormValue("action") + " key=" + r.FormValue("key") + " layout=" + r.FormValue("layout"))
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

// gopodLingoHandler: GET - the voices and the page's words; POST action=try voice text - said and
// heard; POST action=keep voice word said - kept in that voice's lingo (system-lingo).
func gopodLingoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		json.NewEncoder(w).Encode(map[string]any{"voices": wirepod_ttr.GopodLingoVoices(), "texts": wirepod_ttr.GopodSystemTexts("lingo_improver.")})
		return
	}
	voice, text := r.FormValue("voice"), r.FormValue("text")
	switch r.FormValue("action") {
	case "keep":
		note := wirepod_ttr.GopodLingoKeep(voice, r.FormValue("word"), r.FormValue("said"))
		logger.Println("GOPOD_LINGO keep " + voice + ": " + note)
		json.NewEncoder(w).Encode(map[string]string{"note": note})
	default:
		said, audio, note := wirepod_ttr.GopodLingoTry(voice, text)
		logger.Println("GOPOD_LINGO try " + voice + ": " + text + " -> " + said)
		json.NewEncoder(w).Encode(map[string]string{"said": said, "audio": audio, "note": note})
	}
}

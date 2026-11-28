// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_numpad_ptt.go
package wirepod_ttr

import (
	"encoding/json"
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

const gopodPTTStopWait = time.Minute // the release tail + Vosk on up to max_hold_seconds of audio

type gopodNumpadEntry struct {
	Persona string `json:"persona"`
	Body    string `json:"body"`
	Status  string `json:"status"`
	Suit    string `json:"suit"`
}

type gopodNumpadMapFile struct {
	FixedLanes map[string]json.RawMessage  `json:"fixed_lanes"`
	NumlockOn  map[string]gopodNumpadEntry `json:"numlock_on"`
	NumlockOff map[string]gopodNumpadEntry `json:"numlock_off"`
}

// gopodNumpadMapPath: map_file, with "~" as the home of the user who owns GOPOD's overlay folder
// (Wire-Pod runs as root).
func gopodNumpadMapPath() string {
	path := strings.TrimSpace(gopodBrobotsSettings.NumpadPTT.MapFile)
	if strings.HasPrefix(path, "~/") {
		_, uid, _ := gopodRecordRoot()
		if u, err := user.LookupId(strconv.Itoa(uid)); err == nil {
			path = filepath.Join(u.HomeDir, path[2:])
		}
	}
	return path
}

// gopodNumpadLayoutNow: the layout system-numpad-layout chose since this start ("" - the settings'
// numpad_ptt.layout).
var gopodNumpadLayoutNow atomic.Value

// gopodNumpadLayoutPicked: the layout the chat page sent with its newest key press or map read -
// each page picks its own keyboard's layout.
var gopodNumpadLayoutPicked atomic.Value

// GopodNumpadLayoutPick: the page's layout pick for this press ("" or unknown: none).
func GopodNumpadLayoutPick(layout string) {
	layout = strings.ToLower(strings.TrimSpace(layout))
	if _, ok := gopodBrobotsSettings.NumpadMaps[layout]; ok {
		gopodNumpadLayoutPicked.Store(layout)
	}
}

// gopodNumpadLayout: the numpad_maps layout in use - the page's pick, else system-numpad-layout's,
// else numpad_ptt.layout, else "usb".
func gopodNumpadLayout() string {
	if v, _ := gopodNumpadLayoutPicked.Load().(string); v != "" {
		return v
	}
	if v, _ := gopodNumpadLayoutNow.Load().(string); v != "" {
		return v
	}
	if l := strings.TrimSpace(gopodBrobotsSettings.NumpadPTT.Layout); l != "" {
		return l
	}
	return "usb"
}

// gopodSystemNumpadLayout: "system-numpad-layout [name]" - the layouts and the one in use, or
// switch to one (kept in numpad_ptt.layout; the chat page takes it on F5).
func gopodSystemNumpadLayout(text string) string {
	var names []string
	for k := range gopodBrobotsSettings.NumpadMaps {
		names = append(names, k)
	}
	sort.Strings(names)
	f := strings.Fields(strings.ToLower(text))
	if len(f) < 2 {
		return gopodSystemText("system_extras.numpad_layout_now", "layout", gopodNumpadLayout(), "names", strings.Join(names, "|"))
	}
	if _, ok := gopodBrobotsSettings.NumpadMaps[f[1]]; !ok {
		return gopodSystemText("system_extras.numpad_layout_names", "names", strings.Join(names, "|"))
	}
	if _, err := gopodSettingsWrite([]string{"numpad_ptt"}, func(file string) (string, error) {
		return gopodJSONFileSet(file, []string{"numpad_ptt", "layout"}, `"`+f[1]+`"`)
	}); err != nil {
		return gopodSystemText("system_edit.not_changed", "dotted", "numpad_ptt.layout", "error", err.Error())
	}
	gopodNumpadLayoutNow.Store(f[1])
	return gopodSystemText("system_extras.numpad_layout_set", "layout", f[1])
}

func gopodNumpadMap() (gopodNumpadMapFile, []byte, error) {
	return gopodNumpadMapIn(gopodNumpadLayout())
}

// gopodNumpadMapIn: one layout's map (numpad_maps.<layout>).
func gopodNumpadMapIn(layout string) (gopodNumpadMapFile, []byte, error) {
	var m gopodNumpadMapFile
	data := []byte(gopodBrobotsSettings.NumpadMaps[layout])
	if len(data) <= 2 {
		data = []byte(gopodBrobotsSettings.NumpadMap)
	}
	if len(data) <= 2 {
		var err error
		if data, err = os.ReadFile(gopodNumpadMapPath()); err != nil {
			return m, nil, err
		}
	}
	return m, data, json.Unmarshal(data, &m)
}

// GopodNumpadMapJSON is what the page reads: the golden map as it is, plus which body is which
// Brobot.
func GopodNumpadMapJSON() ([]byte, error) { return GopodNumpadMapJSONFor("") }

// GopodNumpadMapJSONFor: the map and targets in one layout (the page's pick; "" - the one in use),
// with the layout and the layouts there are.
func GopodNumpadMapJSONFor(layout string) ([]byte, error) {
	layout = strings.ToLower(strings.TrimSpace(layout))
	if _, ok := gopodBrobotsSettings.NumpadMaps[layout]; !ok {
		layout = gopodNumpadLayout()
	}
	_, raw, err := gopodNumpadMapIn(layout)
	if err != nil {
		return nil, err
	}
	var layouts []string
	for k := range gopodBrobotsSettings.NumpadMaps {
		layouts = append(layouts, k)
	}
	sort.Strings(layouts)
	reporters := map[string]string{}
	for name, p := range gopodBrobotsSettings.Managers.Personas {
		if r := strings.TrimSpace(p.Reporter); r != "" {
			reporters[r] = name
		}
	}
	targets := map[string]string{}
	for n := 0; n <= 9; n++ {
		key := "KP" + strconv.Itoa(n)
		if label, ok := gopodNumpadKeyLabelIn(layout, key); ok {
			targets[key] = label
		} else {
			targets[key] = ""
		}
		if n == 9 {
			if label, _, err := gopodNumpadBrobotIn(layout, gopodNumpadAllKey); err == nil {
				targets[gopodNumpadAllKey] = label
			}
			if label, _, err := gopodNumpadBrobotIn(layout, gopodNumpadAnyKey); err == nil {
				targets[gopodNumpadAnyKey] = label
			}
		}
		if label, ok := gopodNumpadKeyLabelIn(layout, key+gopodNumlockOffSuffix); ok {
			targets[key+gopodNumlockOffSuffix] = label
		}
	}
	return json.Marshal(map[string]any{
		"map":               json.RawMessage(raw),
		"targets":           targets,
		"persona_reporters": reporters,
		"layout":            layout,
		"layouts":           layouts,
	})
}

// gopodNumpadBrobot: the Brobot and ESN a live NumLock-on key talks to.
func gopodNumpadBrobot(key string) (string, string, error) {
	return gopodNumpadBrobotIn(gopodNumpadLayout(), key)
}

// gopodNumpadBrobotIn: gopodNumpadBrobot in one layout.
func gopodNumpadBrobotIn(layout, key string) (string, string, error) {
	if strings.EqualFold(key, gopodNumpadAnyKey) {
		label := strings.TrimSpace(gopodBrobotsSettings.NumpadPTT.Any.Label)
		if label == "" {
			label = "Any persona"
		}
		return label, gopodTypedAny, nil
	}
	if strings.EqualFold(key, gopodNumpadAllKey) {
		label := strings.TrimSpace(gopodBrobotsSettings.NumpadPTT.AllLabel)
		if label == "" {
			label = "All"
		}
		return label, gopodTypedAll, nil
	}
	m, _, err := gopodNumpadMapIn(layout)
	if err != nil {
		return "", "", err
	}
	if base, off := strings.CutSuffix(strings.ToLower(key), gopodNumlockOffSuffix); off {
		base = strings.ToUpper(base)
		entry, ok := m.NumlockOff[base]
		host := gopodPersonaFor(entry.Persona)
		if ok && entry.Status == "live" && entry.Suit == "system" {
			return strings.TrimSpace(entry.Persona), gopodTypedSystem, nil
		}
		if ok && entry.Status == "live" && entry.Suit == "" && host != "" {
			return host, gopodPersonaMark + host, nil
		}
		if ok && entry.Status == "live" && entry.Suit == "robot" {
			brobot := gopodBrobotsSettings.NumpadPTT.BodyToBrobot[entry.Body]
			if brobot == "Brobot 2" && gopodSolo() {
				brobot = "Brobot 1"
			}
			if esn := gopodBrobotESN(brobot); esn != "" {
				return strings.TrimSpace(entry.Persona), gopodDialZeroMark + esn, nil
			}
			return "", "", errors.New(key + "'s body " + entry.Body + " is no Brobot")
		}
		if !ok || entry.Status != "live" || entry.Suit != "reporter" || host == "" || gopodBrobotsSettings.Managers.Personas[host].Reporter == "" {
			return "", "", errors.New(key + " is not live")
		}
		return host + ", reporter", gopodPersonaReporterMark + host, nil
	}
	entry, ok := m.NumlockOn[key]
	if !ok || entry.Status != "live" {
		return "", "", errors.New(key + " is not live")
	}
	if host := gopodPersonaFor(entry.Persona); host != "" {
		return host, gopodPersonaMark + host, nil
	}
	brobot := gopodBrobotsSettings.NumpadPTT.BodyToBrobot[entry.Body]
	if brobot == "Brobot 2" && gopodSolo() {
		brobot = "Brobot 1"
	}
	esn := gopodBrobotESN(brobot)
	if esn == "" {
		return "", "", errors.New(key + "'s body " + entry.Body + " is no Brobot")
	}
	return brobot, esn, nil
}

// gopodPersonaMark: a numpad "ESN" that is a voice-only persona's host.
const gopodPersonaMark = "persona:"

// gopodDialZeroMark: a numpad "ESN" that is a robot as himself, at the dial's 0%.
const gopodDialZeroMark = "dial0:"

func gopodNumpadKeyLabelIn(layout, key string) (string, bool) {
	label, target, err := gopodNumpadBrobotIn(layout, key)
	if err != nil {
		return "", false
	}
	switch {
	case target == gopodTypedAll, target == gopodTypedSystem, target == gopodTypedAny:
		return label, true
	case strings.HasPrefix(target, gopodDialZeroMark):
		return gopodPersonaNickname(gopodBrobotName(strings.TrimPrefix(target, gopodDialZeroMark)), false), true
	case strings.HasPrefix(target, gopodPersonaReporterMark):
		return gopodPersonaNickname(strings.TrimPrefix(target, gopodPersonaReporterMark), false), true
	case strings.HasPrefix(target, gopodPersonaMark):
		return gopodPersonaNickname(strings.TrimPrefix(target, gopodPersonaMark), false), true
	}
	return gopodPersonaNickname(gopodBrobotName(target), true), true
}

// gopodPersonaNickname: the nickname on who's own card; hosted, the persona the body hosts (Brobot
// 1 -> DOC).
func gopodPersonaNickname(who string, hosted bool) string {
	if hosted {
		if n := strings.TrimSpace(gopodBrobotsSettings.HostPersonas[gopodHostedPersona(who)].Nickname); n != "" {
			return n
		}
		return gopodCallName(who)
	}
	if head, _ := gopodPersonaProfileReadOne(gopodPersonaProfileFileIn(gopodBrobotsSettings.PersonaProfiles.Folder, who)); head != nil {
		if n := strings.TrimSpace(head["nickname"]); n != "" {
			return n
		}
	}
	return who
}

// gopodNumpadLiveKeys: the map's live keys in its order - NumLock on KP0-KP9, then NumLock off
// KP0off-KP9off - each one that reaches someone.
func gopodNumpadLiveKeys() []string {
	m, _, err := gopodNumpadMap()
	if err != nil {
		return nil
	}
	var out []string
	for _, off := range []bool{false, true} {
		for n := 0; n <= 9; n++ {
			key := "KP" + strconv.Itoa(n)
			e, ok := m.NumlockOn[key]
			if off {
				e, ok = m.NumlockOff[key]
				key += gopodNumlockOffSuffix
			}
			if !ok || e.Status != "live" {
				continue
			}
			if _, _, err := gopodNumpadBrobot(key); err == nil {
				out = append(out, key)
			}
		}
	}
	return out
}

// gopodKeyReaches: the persona a map target reaches - a voice persona, its reporter suit, a Brobot
// as himself (the dial's 0%), or the persona on a body (DOC, PIP); "" for the line to all.
func gopodKeyReaches(target string) string {
	switch {
	case strings.HasPrefix(target, gopodPersonaReporterMark):
		return gopodBrobotsSettings.Managers.Personas[strings.TrimPrefix(target, gopodPersonaReporterMark)].Reporter
	case strings.HasPrefix(target, gopodPersonaMark):
		return strings.TrimPrefix(target, gopodPersonaMark)
	case strings.HasPrefix(target, gopodDialZeroMark):
		return gopodBrobotName(strings.TrimPrefix(target, gopodDialZeroMark))
	case target == gopodTypedSystem:
		return gopodBrobotsSettings.SignOff.SystemRole
	}
	if body := gopodBrobotName(target); gopodIsBrobotBody(body) {
		if p := gopodHostedPersona(body); p != "" {
			return p
		}
		return body
	}
	return ""
}

// gopodKeyLine: one path for a key's line, typed or held - to wherever the map's target says: a
// voice-only persona, a persona in its reporter suit, a robot as himself (the dial's 0%, for this
// line only), or a robot's hosted persona.
func gopodKeyLine(target, text, how string) error {
	if gopodNeverIn(text, how) {
		return nil
	}
	if fixed := gopodSystemTypo(text); fixed != text && gopodSystemCommandMatch(fixed) && target == gopodPersonaMark+gopodBrobotsSettings.SystemBubble.VoicePersona {
		// said to KEYS: "system comm check" runs as typed to him
		logger.Println("GOPOD_PTT said as a system command: " + text + " -> " + fixed)
		defer gopodCommandFrom(target)()
		return gopodTypedToSystem(fixed, how)
	}
	if gopodRunCommandLine(text) && (target == gopodTypedSystem || target == gopodPersonaMark+gopodBrobotsSettings.SystemBubble.VoicePersona) {
		gopodRunFromZero.Store(true)
		defer gopodCommandFrom(target)()
		return gopodOperatorLine(gopodTypedAll, text, how)
	}
	for _, mark := range []string{gopodPersonaReporterMark, gopodPersonaMark} {
		if host, ok := strings.CutPrefix(target, mark); ok {
			if host == gopodGoverlordHost && gopodGoverlordHandsDown(text, how) {
				return nil
			}
			if who := gopodJobRoute(host, text); who != host {
				target = gopodAnyTarget(who)
			}
			break
		}
	}
	switch {
	case target == gopodTypedSystem:
		return gopodGoverlordCall(text, how)
	case target == gopodTypedAny:
		return gopodAnyPersonaLine(text, how)
	case strings.HasPrefix(target, gopodPersonaReporterMark):
		host := strings.TrimPrefix(target, gopodPersonaReporterMark)
		suit := gopodBrobotsSettings.Managers.Personas[host].Reporter
		return gopodLineTurn(text, []string{suit}, func() error {
			gopodPersonaReplyAs(host, suit, text, how)
			return nil
		})
	case strings.HasPrefix(target, gopodPersonaMark):
		host := strings.TrimPrefix(target, gopodPersonaMark)
		return gopodLineTurn(text, []string{host}, func() error {
			gopodPersonaReply(host, text, how)
			return nil
		})
	case strings.HasPrefix(target, gopodDialZeroMark):
		esn := strings.TrimPrefix(target, gopodDialZeroMark)
		done := gopodReplyDone(esn)
		gopodDialZero.Store(esn, true)
		err := gopodOperatorLine(esn, text, how)
		go func() {
			select {
			case <-done:
			case <-time.After(2 * gopodSecondsToDuration(gopodBrobotsSettings.LineQueue.AnswerMaxSeconds)):
			}
			gopodDialZero.Delete(esn)
		}()
		return err
	}
	if target != gopodTypedAll {
		gopodDialZero.Delete(target)
	}
	return gopodOperatorLine(target, text, how)
}

// gopodPersonaReporterMark: a numpad "ESN" that is a persona in its reporter suit - it answers as
// its reporter (managers.personas.<host>.reporter).
const gopodPersonaReporterMark = "persona-reporter:"

// gopodNumpadAllKey: the numpad's * key held - all the Brobots, in the order selected.
const gopodNumpadAllKey = "KPmul"

// gopodNumlockOffSuffix: a key pressed with NumLock off ("KP7off").
const gopodNumlockOffSuffix = "off"

// gopodMicMissing: the USB mic wasn't found after the warm-up (gopodMicCheck).
var gopodMicMissing atomic.Bool

// gopodMicCheck: the mic (any of numpad_ptt.mic_usb_ids in lsusb) is there, or Goverlord says
// mic_missing_text (with its bubble) and the numpad works dry.
func gopodMicCheck() {
	if !gopodMicLook() {
		gopodSystemShowSay(gopodBrobotsSettings.NumpadPTT.MicMissingText)
	}
	gopodUSBWatchStart()
}

// gopodMicLook: true when the mic is there (or lsusb can't tell); sets gopodMicMissing.
func gopodMicLook() bool {
	found, err := gopodUSBAnyOf(gopodBrobotsSettings.NumpadPTT.MicUSBIDs)
	if err != nil {
		logger.Println("GOPOD_MIC lsusb failed (" + err.Error() + ") - not marking the mic missing")
		gopodMicMissing.Store(false)
		return true
	}
	gopodMicMissing.Store(!found)
	if found {
		logger.Println("GOPOD_MIC found")
	} else {
		logger.Println("GOPOD_MIC missing - numpad works dry")
	}
	return found
}

// GopodNumpadPTTStart: a live key went down - the mic starts keeping the audio.
var gopodNumpadLastKey atomic.Value

func GopodNumpadPTTStart(key string) map[string]any {
	brobot, _, err := gopodNumpadBrobot(key)
	if err != nil {
		return gopodNumpadPTTError(key, "start", err.Error())
	}
	if gopodWarmGateClosed.Load() {
		logger.Println("GOPOD_WARMUP_GATE dropped numpad " + key + " (not ready yet)")
		return map[string]any{"ok": false, "error": "not_ready"}
	}
	if gopodUSBSaying.Load() {
		logger.Println("GOPOD_NUMPAD_PTT dropped " + key + " (the mic message isn't done yet)")
		return map[string]any{"ok": false, "error": "wait_message"}
	}
	if gopodMicMissing.Load() && !gopodMicLook() {
		logger.Println("GOPOD_NUMPAD_PTT start key=" + key + " brobot=" + brobot + " (mic disabled)")
		return map[string]any{"ok": true, "brobot": brobot, "mic_disabled": true}
	}
	r, err := gopodMicCall(map[string]any{"cmd": "ptt_start", "key": key, "dir": gopodSessionLogDir()}, 15*time.Second)
	if err != nil {
		return gopodNumpadPTTError(key, "start", err.Error())
	}
	if ok, _ := r["ok"].(bool); !ok {
		e, _ := r["error"].(string)
		gopodNumpadPTTError(key, "start", e)
	}
	r["brobot"] = brobot
	delete(r, "id")
	logger.Println("GOPOD_NUMPAD_PTT start key=" + key + " brobot=" + brobot)
	return r
}

// GopodNumpadPTTStop: the key came up - the golden tail, then the words, which go to the key's
// Brobot like a typed line.
func GopodNumpadPTTStop(key string) map[string]any {
	brobot, esn, err := gopodNumpadBrobot(key)
	if err != nil {
		return gopodNumpadPTTError(key, "stop", err.Error())
	}
	if gopodMicMissing.Load() {
		time.Sleep(gopodSecondsToDuration(gopodBrobotsSettings.NumpadPTT.MicDisabledSeconds))
		logger.Println("GOPOD_NUMPAD_PTT stop key=" + key + " status=nothing_heard (mic disabled)")
		return map[string]any{"ok": true, "brobot": brobot, "status": "nothing_heard"}
	}
	r, err := gopodMicCall(map[string]any{"cmd": "ptt_stop"}, gopodPTTStopWait)
	if err != nil {
		return gopodNumpadPTTError(key, "stop", err.Error())
	}
	if ok, _ := r["ok"].(bool); !ok {
		e, _ := r["error"].(string)
		gopodNumpadPTTError(key, "stop", e)
	}
	delete(r, "id")
	r["brobot"] = brobot
	status, _ := r["status"].(string)
	text, _ := r["text"].(string)
	if status == "ok" {
		text = gopodSpellNames(text)
		r["text"] = text
	}
	logger.Println("GOPOD_NUMPAD_PTT stop key=" + key + " status=" + status + " text=" + text)
	if status == "ok" && gopodSkipLine(text, "mic") {
		return r
	}
	if status == "ok" && strings.TrimSpace(text) != "" {
		gopodNumpadLastKey.Store(key) // the bubble reads "KP1 to Brobot 1"
		if master, _ := r["master"].(string); master != "" {
			gopodMicVoiceRemember(text, master)
		}
		go func() {
			if err := gopodKeyLine(esn, text, "Numpad text"); err != nil {
				logger.Println("GOPOD_NUMPAD_PTT " + err.Error())
			}
		}()
	}
	return r
}

// gopodNumpadPTTError: the error the page shows in its numpad bubble, in Wire-Pod's log too (and so
// the session's wirepod.log).
func gopodNumpadPTTError(key, step, msg string) map[string]any {
	logger.Println("GOPOD_NUMPAD_PTT error key=" + key + " " + step + ": " + msg)
	return map[string]any{"ok": false, "error": msg}
}

// gopodCommandTo: a command line from a key is shown to the one the numpad map puts on that key -
// the picker, the key and the header all match.
var gopodCommandTo atomic.Value

// gopodCommandFrom: the header for the command line now being routed from target (a numpad map
// target); call the result when done.
func gopodCommandFrom(target string) func() {
	name := gopodTargetName(target)
	gopodCommandTo.Store(name)
	return func() { gopodCommandTo.Store("") }
}

// gopodTargetName: who a numpad map target is, as the chat's headers name them.
func gopodTargetName(target string) string {
	switch {
	case target == gopodTypedSystem:
		return gopodGoverlordHost
	case target == gopodTypedAll:
		return "Brobots"
	case strings.HasPrefix(target, gopodPersonaReporterMark):
		return gopodBrobotsSettings.Managers.Personas[strings.TrimPrefix(target, gopodPersonaReporterMark)].Reporter
	case strings.HasPrefix(target, gopodPersonaMark):
		return strings.TrimPrefix(target, gopodPersonaMark)
	case strings.HasPrefix(target, gopodDialZeroMark):
		return gopodBrobotName(strings.TrimPrefix(target, gopodDialZeroMark))
	}
	return gopodBrobotName(target)
}

// gopodLineTo: the header name for a command line - the key's, else def.
func gopodLineTo(def string) string {
	if v, _ := gopodCommandTo.Load().(string); v != "" {
		return v
	}
	return def
}

// ## EOF

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_numpad_ptt.go
package wirepod_ttr

import (
	"encoding/json"
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// Numpad push-to-talk from the Brobots chat page (operator 2026-10-04: "Does my
// recording via my mic from the web page work? ... please get it to work"). The
// page watches the numpad of whatever keyboard it sees (the Jetson's own, a PC's
// through NoMachine) with the golden numpad's rules, and calls start when a live
// key goes down and stop when it comes up. The mic service records (golden mic
// pick, 48 kHz master per capture, 16 kHz Vosk) and the words go to the key's
// Brobot like a typed line. Which key is which Brobot comes from the golden
// persona map (numpad_ptt.map_file, ~/.gopod_alias_lib/numpad_persona_map_001.json):
// NumLock-on keys with status "live", their body named in numpad_ptt.body_to_brobot.

const gopodPTTStopWait = time.Minute // the release tail + Vosk on up to max_hold_seconds of audio

type gopodNumpadEntry struct {
	Persona string `json:"persona"`
	Body    string `json:"body"`
	Status  string `json:"status"`
}

type gopodNumpadMapFile struct {
	FixedLanes map[string]json.RawMessage  `json:"fixed_lanes"`
	NumlockOn  map[string]gopodNumpadEntry `json:"numlock_on"`
	NumlockOff map[string]gopodNumpadEntry `json:"numlock_off"`
}

// gopodNumpadMapPath: map_file, with "~" as the home of the user who owns GOPOD's
// overlay folder (Wire-Pod runs as root).
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

func gopodNumpadMap() (gopodNumpadMapFile, []byte, error) {
	var m gopodNumpadMapFile
	data, err := os.ReadFile(gopodNumpadMapPath())
	if err != nil {
		return m, nil, err
	}
	return m, data, json.Unmarshal(data, &m)
}

// GopodNumpadMapJSON is what the page reads: the golden map as it is, plus
// which body is which Brobot.
func GopodNumpadMapJSON() ([]byte, error) {
	_, raw, err := gopodNumpadMap()
	if err != nil {
		return nil, err
	}
	// 2026-10-18: the voice-only personas' hosts (KP7 CHALK, KP8 PLAYHEAD).
	personas := map[string]string{}
	for name, p := range gopodBrobotsSettings.Managers.Personas {
		personas[name] = p.Host
	}
	return json.Marshal(map[string]any{
		"map":            json.RawMessage(raw),
		"body_to_brobot": gopodBrobotsSettings.NumpadPTT.BodyToBrobot,
		"persona_hosts":  personas,
	})
}

// gopodNumpadBrobot: the Brobot and ESN a live NumLock-on key talks to.
func gopodNumpadBrobot(key string) (string, string, error) {
	// 2026-10-16 (operator: "a long '0' press acts as the mic version of
	// 'keyboard's 'All' radio button option"): 0 held talks to all the Brobots.
	if key == "KP0" {
		label := strings.TrimSpace(gopodBrobotsSettings.NumpadPTT.AllLabel)
		if label == "" {
			label = "All"
		}
		return label, gopodTypedAll, nil
	}
	m, _, err := gopodNumpadMap()
	if err != nil {
		return "", "", err
	}
	entry, ok := m.NumlockOn[key]
	if !ok || entry.Status != "live" {
		return "", "", errors.New(key + " is not live")
	}
	// 2026-10-18 (operator: "KP7 live is Reporter 1 (female voice) is CHALK
	// persona * KP8 live is Reporter 2 (male voice) is PLAYHEAD persona"): a
	// voice-only persona talks to its host (gopodPersonaReply).
	if host := gopodPersonaFor(entry.Persona); host != "" {
		return host, gopodPersonaMark + host, nil
	}
	brobot := gopodBrobotsSettings.NumpadPTT.BodyToBrobot[entry.Body]
	esn := gopodBrobotESN(brobot)
	if esn == "" {
		return "", "", errors.New(key + "'s body " + entry.Body + " is no Brobot")
	}
	return brobot, esn, nil
}

// gopodPersonaMark: a numpad "ESN" that is a voice-only persona's host.
const gopodPersonaMark = "persona:"

// gopodMicMissing: the USB mic wasn't found after the warm-up (gopodMicCheck).
var gopodMicMissing atomic.Bool

// gopodMicCheck: the mic (numpad_ptt.mic_usb_id in lsusb) is there, or Goverlord
// says mic_missing_text and the numpad works dry. Checked after the warm-up.
func gopodMicCheck() {
	id := strings.TrimSpace(gopodBrobotsSettings.NumpadPTT.MicUSBID)
	if id == "" {
		return
	}
	found := false
	for _, d := range gopodUSBDevices() {
		if strings.EqualFold(d.ID, id) {
			found = true
		}
	}
	gopodMicMissing.Store(!found)
	if found {
		logger.Println("GOPOD_MIC found " + id)
		return
	}
	logger.Println("GOPOD_MIC missing " + id + " - numpad works dry")
	if t := strings.TrimSpace(gopodBrobotsSettings.NumpadPTT.MicMissingText); t != "" {
		gopodSystemSay(t, false)
	}
}

// GopodNumpadPTTStart: a live key went down - the mic starts keeping the audio.
// gopodNumpadLastKey: the key of the newest numpad capture ("KP1").
var gopodNumpadLastKey atomic.Value

func GopodNumpadPTTStart(key string) map[string]any {
	brobot, _, err := gopodNumpadBrobot(key)
	if err != nil {
		return gopodNumpadPTTError(key, "start", err.Error())
	}
	// 2026-10-09 (operator: "LLM not ready? Too bad!! Wait!!"): no recording
	// before LLM Ready (gopod_llm_warmup.go).
	if gopodWarmGateClosed.Load() {
		logger.Println("GOPOD_WARMUP_GATE dropped numpad " + key + " (not ready yet)")
		return map[string]any{"ok": false, "error": "not_ready"}
	}
	// Mic missing (gopodMicCheck): the key works dry - same messages, nothing kept.
	if gopodMicMissing.Load() {
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

// GopodNumpadPTTStop: the key came up - the golden tail, then the words, which
// go to the key's Brobot like a typed line.
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
	logger.Println("GOPOD_NUMPAD_PTT stop key=" + key + " status=" + status + " text=" + text)
	if status == "ok" && strings.TrimSpace(text) != "" {
		gopodNumpadLastKey.Store(key) // the bubble reads "KP1 to Brobot 1"
		go func() {
			if strings.HasPrefix(esn, gopodPersonaMark) {
				gopodPersonaReply(strings.TrimPrefix(esn, gopodPersonaMark), text, "Numpad text")
				return
			}
			if err := gopodOperatorLine(esn, text, "Numpad text"); err != nil {
				logger.Println("GOPOD_NUMPAD_PTT " + err.Error())
			}
		}()
	}
	return r
}

// gopodNumpadPTTError (2026-10-11, operator: "can those error messages be
// logged in all-logs, as well?"): the error the page shows in its numpad
// bubble, in Wire-Pod's log too (and so the session's wirepod.log).
func gopodNumpadPTTError(key, step, msg string) map[string]any {
	logger.Println("GOPOD_NUMPAD_PTT error key=" + key + " " + step + ": " + msg)
	return map[string]any{"ok": false, "error": msg}
}

// ## EOF

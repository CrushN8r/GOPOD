// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_voice_map.go
package wirepod_ttr

// The numpad-to-voices map (2026-10-26, operator: "In GOPOD system, can I have
// a Numpad-to-Voices dynamic-mapping? * TEMPLATE: KPN > [Persona Name] [drop
// down menu] * N = 0-9, Drop down Menu = dynamic list of all available voices *
// duplicate voice selections allowed, and are minimally highlighted (bolded?)";
// "KP0 (All), KP1 and KP2: \"no voice required\" talks to Brobots * these are
// 'robot-bodied' personas * KP3 to KP6: upon robot connectivity, \"no voice
// required\" * KP7 to KP9, Operator: KPN-TEMPLATE"). One row per key, then the
// Operator: a key whose persona has a robot body Wire-Pod knows
// (numpad_ptt.body_to_brobot), and KP0 (All), needs no voice; every other key
// picks its persona's voice from the Kokoro voices on disk. A pick is saved to
// the settings file (reporter_voices.hosts, operator_voice) and used at once.

import (
	"errors"
	"path/filepath"
	"sort"
	"strings"
)

type gopodVoiceRow struct {
	Key     string `json:"key"`
	Persona string `json:"persona"`
	Host    string `json:"host"`
	Voice   string `json:"voice"`
	Bodied  bool   `json:"bodied"`
}

// gopodVoiceList: every Kokoro voice on disk (the .pt files), by name.
func gopodVoiceList() []string {
	home := gopodBrobotsSettings.ReporterVoices.HFHome
	if home == "" {
		home = gopodHomePath("~/.cache/huggingface")
	}
	files, _ := filepath.Glob(filepath.Join(home, "hub", "models--hexgrad--Kokoro-82M", "snapshots", "*", "voices", "*.pt"))
	seen := map[string]bool{}
	var out []string
	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f), ".pt")
		if !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// gopodVoiceRows: KP0-KP9, then the Operator.
func gopodVoiceRows() []gopodVoiceRow {
	m, _, _ := gopodNumpadMap()
	rv := gopodBrobotsSettings.ReporterVoices
	var rows []gopodVoiceRow
	for n := 0; n <= 9; n++ {
		key := "KP" + string(rune('0'+n))
		if n == 0 {
			label := strings.TrimSpace(gopodBrobotsSettings.NumpadPTT.AllLabel)
			if label == "" {
				label = "All"
			}
			rows = append(rows, gopodVoiceRow{Key: key, Persona: label, Bodied: true})
			continue
		}
		e := m.NumlockOn[key]
		row := gopodVoiceRow{Key: key, Persona: e.Persona}
		// 2026-10-29 (operator: "follow numpad for voices"): with robots off, a
		// bodied Brobot's key picks the voice his says play in (gopod_robots_off.go).
		if brobot, robot := gopodBrobotsSettings.NumpadPTT.BodyToBrobot[e.Body]; robot && GopodRobotsOff() {
			row.Host, row.Voice = brobot, gopodBrobotsSettings.RobotsOff.Voices[brobot]
		} else if robot {
			row.Bodied = true
		} else if host := gopodPersonaFor(e.Persona); host != "" {
			row.Host, row.Voice = host, rv.Hosts[host]
		}
		rows = append(rows, row)
	}
	return append(rows, gopodVoiceRow{Key: "Operator", Persona: "Operator", Host: gopodOperatorVoiceHost, Voice: rv.OperatorVoice})
}

// GopodVoiceMap: what the system menu's voices panel shows.
func GopodVoiceMap() map[string]any {
	return map[string]any{"rows": gopodVoiceRows(), "voices": gopodVoiceList(), "unassigned_host": gopodBrobotsSettings.ReporterVoices.UnassignedVoiceHost}
}

// GopodVoiceMapSet: key's persona (or the Operator) takes voice - saved, and
// used from the next line on. The Operator's "" = the unassigned voice.
func GopodVoiceMapSet(key, voice string) (string, error) {
	voice = strings.TrimSpace(voice)
	var row *gopodVoiceRow
	rows := gopodVoiceRows()
	for i := range rows {
		if strings.EqualFold(rows[i].Key, strings.TrimSpace(key)) {
			row = &rows[i]
		}
	}
	if row == nil {
		return "", errors.New("no key " + key)
	}
	if row.Bodied || row.Host == "" {
		return "", errors.New(row.Key + " needs no voice")
	}
	known := voice == "" && row.Key == "Operator"
	for _, v := range gopodVoiceList() {
		known = known || v == voice
	}
	if !known {
		return "", errors.New("no voice " + voice)
	}
	settings := gopodSettingsFileFor([]string{"reporter_voices"})
	raw := `"` + voice + `"`
	rv := &gopodBrobotsSettings.ReporterVoices
	if gopodIsBodiedBrobot(row.Host) {
		// 2026-10-29: robots off - a Brobot's temp voice, in robots_off.voices.
		settings = gopodSettingsFileFor([]string{"robots_off"})
		if _, err := gopodJSONFileSet(settings, []string{"robots_off", "voices", row.Host}, raw); err != nil {
			return "", err
		}
		voices := map[string]string{}
		for k, v := range gopodBrobotsSettings.RobotsOff.Voices {
			voices[k] = v
		}
		voices[row.Host] = voice
		gopodBrobotsSettings.RobotsOff.Voices = voices
	} else if row.Key == "Operator" {
		if _, err := gopodJSONFileSet(settings, []string{"reporter_voices", "operator_voice"}, raw); err != nil {
			return "", err
		}
		rv.OperatorVoice = voice
	} else {
		if _, err := gopodJSONFileSet(settings, []string{"reporter_voices", "hosts", row.Host}, raw); err != nil {
			return "", err
		}
		// A new map, never the one being read.
		hosts := make(map[string]string, len(rv.Hosts)+1)
		for k, v := range rv.Hosts {
			hosts[k] = v
		}
		hosts[row.Host] = voice
		rv.Hosts = hosts
	}
	shown := voice
	if shown == "" {
		shown = "unassigned (" + rv.UnassignedVoiceHost + "'s voice)"
	}
	return row.Key + " > " + row.Persona + ": " + shown, nil
}

// gopodSystemVoices: "system-voices" - the map as text.
func gopodSystemVoices() string {
	var b strings.Builder
	b.WriteString("Numpad voices (system menu: Voices to change them):")
	for _, r := range gopodVoiceRows() {
		v := r.Voice
		switch {
		case r.Bodied:
			v = "no voice required"
		case v == "" && r.Key == "Operator":
			v = "unassigned"
		case v == "":
			v = "-"
		}
		b.WriteString("\n  " + r.Key + " > " + r.Persona + ": " + v)
	}
	return b.String()
}

// ## EOF

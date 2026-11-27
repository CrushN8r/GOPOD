// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_text.go
package wirepod_ttr

import "strings"

// gopodSystemText: a system command's reply, its words from system_text in the settings file, each
// {name} filled with the value given after it ("word", word).
func gopodSystemText(key string, kv ...string) string {
	t, ok := gopodBrobotsSettings.SystemText[key]
	if !ok {
		return "[system_text." + key + "]"
	}
	for i := 0; i+1 < len(kv); i += 2 {
		t = strings.ReplaceAll(t, "{"+kv[i]+"}", kv[i+1])
	}
	return t
}

// ## EOF

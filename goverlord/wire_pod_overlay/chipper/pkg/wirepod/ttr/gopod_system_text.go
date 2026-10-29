// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_text.go
package wirepod_ttr

import "strings"

// gopodSystemText (2026-10-29, operator: "decouple dumb code from flexible data &
// bridge the two by loading external already templated files ... then code
// blindly runs templates from loaded files"; pass 3 of FLEXIBLE_DATA_SURVEY_001.md):
// a system command's reply, its words from system_text in the settings file, each
// {name} filled with the value given after it ("word", word). A key the settings
// don't have shows as [system_text.<key>] - TestSystemTextKeys catches it first.
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

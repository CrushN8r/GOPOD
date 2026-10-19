// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_speak_links.go
package wirepod_ttr

import "regexp"

// gopodLinkRE: a link in a reply - "[words](/page.html)", "*words*(/page.html)"
// (the reply's internal form) or "words (/page.html)".
var gopodLinkRE = regexp.MustCompile(`(?:\[([^\]]+)\]|\*([^*]+)\*)?\s*\(((?:/|https?://)[^)\s]*)\)`)

// gopodLinkWords (2026-10-16, operator: "gopod_weather_kit.html = spoken as
// gopod weather kit dot-h tee-emm-elle"): a link is said as its words, never
// its address.
func gopodLinkWords(s string) string {
	return gopodLinkRE.ReplaceAllStringFunc(s, func(m string) string {
		sub := gopodLinkRE.FindStringSubmatch(m)
		if sub[1] != "" {
			return sub[1]
		}
		if sub[2] != "" {
			return sub[2]
		}
		return ""
	})
}

// ## EOF

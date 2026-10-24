// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_operator_line.go
package wirepod_ttr

import "strings"

// The way in (2026-10-24, the root-to-leaf pass, link 1; operator: "typed, mic,
// numpad, backpack and visitor lines all enter as one Operator line template"):
// every line the Operator (or an approved visitor) says to GOPOD is shown here,
// once - its bubble "[input] to <who>" - and read out here, through the voice
// template (the mic's own recording for a numpad line). A line a Brobot heard
// himself (backpack, Hey Vector) is shown the moment it's heard
// (gopodHeardLine); it's never shown again on its way in. Nothing else posts
// an Operator bubble (TestOperatorLineTemplate).

// gopodHeardLine (2026-10-24, live: "Backpack to Brobot N / good afternoon"
// showed under his answer - the page read the heard line from Wire-Pod's log,
// written only once the answer was ready): a line a Brobot heard himself
// (backpack, Hey Vector) goes in the chat the moment it's heard - the Operator's
// bubble, "[Backpack]", to the Brobot it names first, else to him - in the
// feed's order, before anything answers it. The one source, like a typed line.
func gopodHeardLine(esn, text string) {
	text = strings.TrimSpace(text)
	if esn == "" || text == "" {
		return
	}
	to := gopodFirstNamed(text)
	if to == "" {
		to = gopodBrobotName(esn)
	}
	gopodChatFeedAddOperator(to, text, gopodOperatorVia(gopodVoiceCommandHow))
}

// gopodOperatorShow: the line's Operator bubble - unless a Brobot heard it
// himself (shown already, gopodHeardLine).
func gopodOperatorShow(to, text, how string) {
	if how == gopodVoiceCommandHow || strings.TrimSpace(text) == "" {
		return
	}
	gopodChatFeedAddOperator(to, text, gopodOperatorVia(how))
}

// gopodOperatorRead: the line read out, when its input is read
// (reporter_voices.read_lines_from) - the voice template; returns once heard.
func gopodOperatorRead(to, text, how string) {
	gopodReporterReadsOperator(text, how, to)
}

// ## EOF

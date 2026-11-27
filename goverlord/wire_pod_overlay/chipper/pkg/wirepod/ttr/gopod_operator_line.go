// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_operator_line.go
package wirepod_ttr

import "strings"

// gopodHeardLine: a line a Brobot heard himself (backpack, Hey Vector) goes in the chat the moment
// it's heard - the Operator's bubble, "[Backpack]", to the Brobot it names first, else to him - in
// the feed's order, before anything answers it.
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

// gopodOperatorShow: the line's Operator bubble - unless a Brobot heard it himself (shown already,
// gopodHeardLine).
func gopodOperatorShow(to, text, how string) {
	if how == gopodVoiceCommandHow || strings.TrimSpace(text) == "" {
		return
	}
	gopodChatFeedAddOperator(to, text, gopodOperatorVia(how))
}

// gopodOperatorRead: the line read out, when its input is read (reporter_voices.read_lines_from) -
// the voice template; returns once heard.
func gopodOperatorRead(to, text, how string) {
	gopodReporterReadsOperator(text, how, to)
}

// ## EOF

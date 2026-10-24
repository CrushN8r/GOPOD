// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_greetings.go
package wirepod_ttr

import (
	"regexp"
	"strings"
	"time"
)

// Good morning / afternoon / evening / night (2026-10-24, operator: "good +
// morning/afternoon/evening ... bubble order messed up ... upon, good
// [time-of-day], is correct vs time? Respond accordingly = new compatible
// answer-chain"): Wire-Pod's own greeting intents got it wrong (live: "good
// afternoon" as good morning, "good evening" as praise) and answered outside the
// chat's order. Now GOPOD takes every "good <part of day>": it goes to the chat
// (gopodNativeToManager), and the answer chain's golden_line part opens with the
// greeting checked against the real time - greetings.right_text when it fits,
// greetings.wrong_text when it doesn't - then the Brobot carries on as ever.

var gopodGreetingRE = regexp.MustCompile(`(?i)\bgood\s+(morning|afternoon|evening|night)\b`)

// gopodGreetingSaid: the part of day a line greets ("" when it doesn't).
func gopodGreetingSaid(text string) string {
	if m := gopodGreetingRE.FindStringSubmatch(text); m != nil {
		return strings.ToLower(m[1])
	}
	return ""
}

// gopodPartOfDay: the part of day now (greetings.parts, by hour).
func gopodPartOfDay(now time.Time) string {
	h := now.Hour()
	for _, p := range gopodBrobotsSettings.Greetings.Parts {
		if (p.From <= p.To && h >= p.From && h < p.To) || (p.From > p.To && (h >= p.From || h < p.To)) {
			return p.Name
		}
	}
	return ""
}

// gopodGreetingLine: the golden greeting for a line - "" when it greets no part
// of day.
func gopodGreetingLine(esn, text string, now time.Time) string {
	said := gopodGreetingSaid(text)
	part := gopodPartOfDay(now)
	if said == "" || part == "" {
		return ""
	}
	g := gopodBrobotsSettings.Greetings
	tmpl := g.RightText
	if said != part {
		tmpl = g.WrongText
	}
	return strings.TrimSpace(strings.NewReplacer("{said}", said, "{part}", part, "{time}", gopodClockFor(esn, now)).Replace(tmpl))
}

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

// gopodClockFor: the time as this Brobot's own clock says it (his 12/24-hour
// setting, the weather line's way): "10:19 AM", or "1019 Hours Military Time".
func gopodClockFor(esn string, t time.Time) string {
	_, clock24 := gopodRobotWeatherFormat(esn)
	if !clock24 {
		return t.Format("3:04 PM")
	}
	if label := strings.TrimSpace(gopodBrobotsSettings.LiveWeather.MilitaryLabel); label != "" {
		return t.Format("1504") + " " + label
	}
	return t.Format("15:04")
}

// ## EOF

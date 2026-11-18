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
	// 2026-11-18 (operator: "use 'night' to close with, as in good night, see you
	// tomorrow * otherwise its 'evening'"): "good night" in the evening is a
	// closing - said back as one.
	if said == "night" && part == "evening" {
		part = said
	}
	if said != part {
		tmpl = g.WrongText
	}
	return strings.TrimSpace(strings.NewReplacer("{said}", said, "{part}", part, "{time}", gopodClockFor(esn, now)).Replace(tmpl))
}

// gopodGreetingLeadRE: a greeting of the day an answer opens with itself.
var gopodGreetingLeadRE = regexp.MustCompile(`(?i)^\s*good\s+(morning|afternoon|evening|night|day)\b[^\pL]*`)

// gopodGreetingOpen (2026-11-15, operator: "Have everyone start their response
// with a greeting of the day. Good morning, good afternoon, good evening, good
// night et. Cetera"): what every answer opens with - the reply to a line's
// greeting (gopodGreetingLine), else the greeting for the part of day it is now
// (greetings.day_text); "" when off.
func gopodGreetingOpen(esn, text string, now time.Time) string {
	if line := gopodGreetingLine(esn, text, now); line != "" {
		return line
	}
	part := gopodPartOfDay(now)
	if part == "" {
		return ""
	}
	return strings.TrimSpace(strings.ReplaceAll(gopodBrobotsSettings.Greetings.DayText, "{part}", part))
}

// gopodGreetingStitch: reply opening with its greeting - one the model wrote
// itself is dropped, so it's said once, and right for the time.
func gopodGreetingStitch(esn, text, reply string) string {
	g := gopodGreetingOpen(esn, text, time.Now())
	if g == "" || strings.TrimSpace(reply) == "" {
		return reply
	}
	return gopodStitchSnippet(g, gopodGreetingLeadRE.ReplaceAllString(reply, ""))
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

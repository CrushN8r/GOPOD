// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_here_now.go
package wirepod_ttr

import (
	"strings"
	"time"
)

// The here-and-now note (2026-10-26, operator: "all chat participants should
// be aware of themselves & others in the chat, in spacetime (time & location)";
// 2026-10-27: "rather than the llm asses all that * pre-generate all that for
// the llm, instead * Time is [time]. You are [name], located in [Windsor Ontario
// Canada]. You're chatting with friends, [list chat participants]. Check recent
// chat history for context"). All worked out in code, for every Gomad's prompt
// and before every Brobot's question - here_now.text with {time}, {self},
// {gender}, {place} (the live place, else here_now.place), {cast} (everyone but
// him, each "name (gender, role)"; to a Brobot his brother is brother_line,
// never named) and {batteries} (only for a battery guard, battery_watch.guards).
// Gender: here_now.genders, else his voice (reporter_voices.hosts: Kokoro's
// af_/am_ - the second letter, voice_genders).

// gopodHereNow: the note for self ("Brobot 1", "CHALK", ...); "" when off.
func gopodHereNow(self string) string {
	hn := gopodBrobotsSettings.HereNow
	if strings.TrimSpace(hn.Text) == "" {
		return ""
	}
	place, clock := gopodHereNowPlaceTime()
	batteries := ""
	for _, guard := range gopodBrobotsSettings.BatteryWatch.Guards {
		if guard == self {
			batteries = strings.ReplaceAll(hn.BatteryNote, "{batteries}", gopodHereNowBatteries())
		}
	}
	return strings.Join(strings.Fields(strings.NewReplacer(
		"{time}", clock, "{self}", gopodHereNowName(self), "{gender}", gopodHereNowGender(self),
		"{place}", place, "{cast}", gopodHereNowCast(self), "{batteries}", batteries).Replace(hn.Text)), " ")
}

func gopodHereNowName(who string) string {
	if n := gopodBrobotsSettings.HostPersonas[who].Name; n != "" {
		return n
	}
	return who
}

// gopodHereNowGender: here_now.genders, else from his voice.
func gopodHereNowGender(who string) string {
	hn := gopodBrobotsSettings.HereNow
	if g := hn.Genders[who]; g != "" {
		return g
	}
	if v := gopodBrobotsSettings.ReporterVoices.Hosts[who]; len(v) >= 2 {
		return hn.VoiceGenders[v[1:2]]
	}
	return ""
}

// gopodHereNowCast: everyone in the chat but self, each his line.
func gopodHereNowCast(self string) string {
	hn := gopodBrobotsSettings.HereNow
	brobot := self == "Brobot 1" || self == "Brobot 2"
	var out []string
	for _, c := range hn.Cast {
		switch {
		case c.Who == self:
			continue
		case brobot && (c.Who == "Brobot 1" || c.Who == "Brobot 2"):
			out = append(out, hn.BrotherLine)
		default:
			out = append(out, strings.NewReplacer("{gender}", gopodHereNowGender(c.Who)).Replace(c.Line))
		}
	}
	return strings.Join(out, hn.CastJoin)
}

// gopodHereNowPlaceTime: the live place and clock, else here_now.place and the clock.
func gopodHereNowPlaceTime() (string, string) {
	hn := gopodBrobotsSettings.HereNow
	facts, err := gopodLiveWeatherFacts("")
	if err != nil {
		return hn.Place, time.Now().Format(hn.ClockFormat)
	}
	where, _ := gopodPlaceAndClock(gopodBrobot1ESN, facts)
	if where == "" {
		where = hn.Place
	}
	return where, facts.Local.Format(hn.ClockFormat)
}

func gopodHereNowBatteries() string {
	var out []string
	gopodBatteryLast.Lock()
	defer gopodBatteryLast.Unlock()
	for _, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN} {
		text := gopodBatteryLast.text[esn]
		if text == "" {
			text = "not read yet"
		}
		out = append(out, strings.NewReplacer("{brobot}", gopodBrobotName(esn), "{reading}", text).Replace(gopodBrobotsSettings.HereNow.BatteryText))
	}
	return strings.Join(out, gopodBrobotsSettings.HereNow.CastJoin)
}

// ## EOF

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_here_now.go
package wirepod_ttr

import (
	"strings"
	"time"
)

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
	if n := gopodPersona(who).Name; n != "" {
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
			if !gopodSolo() {
				out = append(out, hn.BrotherLine)
			}
		default:
			out = append(out, gopodPersonaFill(c.Who, strings.NewReplacer("{gender}", gopodHereNowGender(c.Who)).Replace(c.Line)))
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

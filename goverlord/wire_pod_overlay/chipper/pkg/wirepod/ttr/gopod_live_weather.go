// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_live_weather.go
package wirepod_ttr

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

type gopodWeatherFacts struct {
	Name, State string
	Country     string
	Description string
	TempC       float64
	Local       time.Time
}

type gopodWeatherCached struct {
	facts gopodWeatherFacts
	at    time.Time
}

var gopodWeatherCache = struct {
	sync.Mutex
	places map[string]gopodWeatherCached
}{places: map[string]gopodWeatherCached{}}

// gopodWeatherPlaceRE: the place a weather question names - "the weather in toronto ontario, right
// now?", "weather for Toronto, ON, CA, please".
var gopodWeatherPlaceRE = regexp.MustCompile(`(?i)\bweather\b.*?\b(?:in|for|at)\s+([\p{L}][\p{L} .,'-]*?)\s*(?:,?\s*\b(?:right now|now|today|tonight|currently|please|bro)\b|[?!.]|$)`)

// gopodWeatherPlaceBeforeRE: "in Honolulu Hawaii", "to New York" - a place in capitals anywhere in
// the line.
var gopodWeatherPlaceBeforeRE = regexp.MustCompile(`\b(?:in|to|at|from)\s+(\p{Lu}[\p{L}'-]*(?:,?\s+\p{Lu}[\p{L}'-]*)*)`)

// gopodWeatherPlaceSmallRE / gopodWeatherNotPlace: a place typed small ("in jerusalem"), and the
// words that are never one (or end one).
var gopodWeatherPlaceSmallRE = regexp.MustCompile(`\b(?:in|to|from)\s+([a-z][a-z'-]*(?:\s+[a-z][a-z'-]*)?)`)
var gopodWeatherNotPlace = map[string]bool{
	"a": true, "an": true, "the": true, "my": true, "your": true, "our": true, "his": true, "her": true,
	"their": true, "this": true, "that": true, "these": true, "those": true, "there": true, "here": true,
	"what": true, "what's": true, "whats": true, "how": true, "how's": true, "hows": true, "is": true,
	"are": true, "and": true, "or": true, "but": true, "so": true, "today": true, "tomorrow": true,
	"tonight": true, "now": true, "right": true, "please": true, "bro": true, "it": true, "be": true,
	"go": true, "see": true, "visit": true, "get": true, "talk": true, "say": true, "know": true,
	"bed": true, "school": true, "town": true, "time": true, "case": true, "fact": true, "front": true,
	"general": true, "love": true, "mind": true, "order": true, "traffic": true, "line": true,
	"person": true, "car": true, "house": true, "home": true, "chat": true, "room": true, "building": true,
	"weather": true, "brobot": true, "brobots": true, "doc": true, "pip": true, "me": true, "you": true,
	"him": true, "them": true, "us": true, "do": true, "be.": true, "work": true, "bingo": true,
}

// gopodWeatherPlace: the place the question names, or "" for the usual live_weather.location.
func gopodWeatherPlace(message string) string {
	place := ""
	if m := gopodWeatherPlaceRE.FindStringSubmatch(message); m != nil {
		place = strings.Trim(strings.TrimSpace(m[1]), " ,.'-")
	}
	switch strings.ToLower(place) {
	case "", "here", "there", "the chat", "the room", "the building":
		place = ""
	}
	if place == "" {
		for _, m := range gopodWeatherPlaceBeforeRE.FindAllStringSubmatch(message, -1) {
			name := strings.Trim(strings.TrimSpace(m[1]), " ,.'-")
			first := strings.ToLower(strings.Fields(name + " x")[0])
			switch first {
			case "brobot", "brobots", "doc", "pip", "goverlord", "gopod", "vector", "x":
				continue
			}
			place = name
		}
	}
	if place == "" {
		for _, m := range gopodWeatherPlaceSmallRE.FindAllStringSubmatch(strings.ToLower(message), -1) {
			var words []string
			for _, w := range strings.Fields(m[1]) {
				if gopodWeatherNotPlace[w] {
					break
				}
				words = append(words, w)
			}
			if len(words) > 0 {
				place = strings.Join(words, " ")
			}
		}
	}
	if place == "" {
		return ""
	}
	small := place == strings.ToLower(place)
	for _, packet := range gopodTopicRegistry() {
		if !small && len(gopodBrobotsSettings.TopicPictures.Pictures[packet.Name]) == 0 {
			continue
		}
		for _, w := range packet.Triggers {
			if strings.EqualFold(strings.TrimSpace(w), place) {
				return ""
			}
		}
	}
	return place
}

// gopodLiveWeatherFacts fetches the weather at place ("" = live_weather.location), or returns that
// place's last fetch when it is newer than live_weather.cache_seconds (so Brobot 2 checks the same
// reading).
func gopodLiveWeatherFacts(place string) (gopodWeatherFacts, error) {
	lw := gopodBrobotsSettings.LiveWeather
	if strings.TrimSpace(place) == "" {
		place = lw.Location
	}
	cacheKey := strings.ToLower(strings.TrimSpace(place))
	gopodWeatherCache.Lock()
	defer gopodWeatherCache.Unlock()
	if c, ok := gopodWeatherCache.places[cacheKey]; ok && time.Since(c.at) < gopodSecondsToDuration(lw.CacheSeconds) {
		return c.facts, nil
	}
	key := vars.APIConfig.Weather.Key
	if key == "" || place == "" {
		return gopodWeatherFacts{}, errors.New("no weather key or location")
	}
	client := &http.Client{Timeout: gopodTiming("weather_seconds")}
	get := func(u string, into any) error {
		resp, err := client.Get(u)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("weather service answered %d", resp.StatusCode)
		}
		return json.NewDecoder(resp.Body).Decode(into)
	}
	tries := geocodeCandidates(place)
	if f := strings.Fields(strings.ReplaceAll(place, ",", " ")); len(f) > 1 {
		tries = append(tries, f[0])
	}
	city, country := gopodWeatherCountry(place)
	for _, c := range gopodBrobotsSettings.CleanPolitics.Continents {
		city = strings.TrimSpace(strings.ReplaceAll(" "+strings.ToLower(city)+" ", " "+c+" ", " "))
	}
	if country != "" && city == "" {
		city = gopodBrobotsSettings.CleanPolitics.Capitals[country]
	}
	if state, capital := gopodStateCapital(city); capital != "" {
		tries = append([]string{capital + ", " + state}, tries...)
		if country != "" {
			tries = append([]string{capital + ",," + country}, tries...)
		}
		city = ""
	}
	if country != "" && city != "" {
		tries = append([]string{city + ",," + country, city + "," + country}, tries...)
	}
	// 14, "Is it the same party weather in Windsor Ontario?" was answered with Windsor England): each
	// try asks for a few places, and the one in the province/state/country the question named wins
	// (gopodPickPlace); with none named, the first.
	var geo []gopodGeoPlace
	var found *gopodGeoPlace
	for _, q := range tries {
		var got []gopodGeoPlace
		if err := get("http://api.openweathermap.org/geo/1.0/direct?q="+url.QueryEscape(q)+"&limit=5&appid="+key, &got); err != nil {
			return gopodWeatherFacts{}, err
		}
		if len(got) > 0 && geo == nil {
			geo = got
		}
		if country != "" {
			for i := range got {
				if strings.EqualFold(got[i].Country, country) {
					found = &got[i]
					break
				}
			}
			if found != nil {
				break
			}
			continue
		}
		if i := gopodPickPlace(got, place); i >= 0 {
			found = &got[i]
			break
		}
	}
	if found == nil && len(geo) > 0 {
		found = &geo[0]
	}
	if found == nil {
		return gopodWeatherFacts{}, errors.New("no place found for " + place)
	}
	var now struct {
		Weather []struct {
			Description string `json:"description"`
		} `json:"weather"`
		Main struct {
			Temp float64 `json:"temp"`
		} `json:"main"`
		DT       int64 `json:"dt"`
		Timezone int64 `json:"timezone"`
	}
	if err := get(fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&units=metric&appid=%s", found.Lat, found.Lon, key), &now); err != nil {
		return gopodWeatherFacts{}, err
	}
	if len(now.Weather) == 0 {
		return gopodWeatherFacts{}, errors.New("no weather in the answer")
	}
	facts := gopodWeatherFacts{
		Name: found.Name, State: found.State, Country: strings.ToUpper(found.Country),
		Description: now.Weather[0].Description,
		TempC:       now.Main.Temp,
		Local:       time.Unix(now.DT+now.Timezone, 0).UTC(),
	}
	gopodWeatherCache.places[cacheKey] = gopodWeatherCached{facts: facts, at: time.Now()}
	return facts, nil
}

// gopodStateCapital: the state (or province) a place is, alone or as "<somewhere> in <state>", and
// its capital (clean_politics.state_capitals); "", "" for any other place.
func gopodStateCapital(place string) (string, string) {
	lc := strings.ToLower(strings.Join(strings.Fields(strings.ReplaceAll(place, ",", " ")), " "))
	if lc == "" {
		return "", ""
	}
	best, capital := "", ""
	for state, c := range gopodBrobotsSettings.CleanPolitics.StateCapitals {
		s := strings.ToLower(strings.TrimSpace(state))
		if s != "" && (lc == s || strings.HasSuffix(lc, " in "+s)) && len(s) > len(best) {
			best, capital = state, strings.TrimSpace(c)
		}
	}
	return best, capital
}

// gopodWeatherCountry: the place without a country named in it, and that country's code ("Toronto
// USA" -> "Toronto", "US"); "" when none is named.
func gopodWeatherCountry(place string) (string, string) {
	lower := " " + strings.ToLower(strings.Join(strings.Fields(strings.ReplaceAll(place, ",", " ")), " ")) + " "
	best, code := "", ""
	for name, c := range gopodBrobotsSettings.CleanPolitics.Countries {
		n := " " + strings.ToLower(name) + " "
		if c != "" && strings.Contains(lower, n) && len(name) > len(best) {
			best, code = name, strings.ToUpper(c)
		}
	}
	if code == "" {
		return place, ""
	}
	city := strings.TrimSpace(strings.Replace(lower, " "+strings.ToLower(best)+" ", " ", 1))
	if i := strings.Index(strings.ToLower(place), city); i >= 0 && city != "" {
		city = place[i : i+len(city)]
	}
	return strings.Trim(city, " ,"), code
}

// gopodWeatherCountriesSaid: the countries of the weather lines said in the last two minutes -
// their flags are right (gopodCleanPolitics keeps them).
func gopodWeatherCountriesSaid() []string {
	gopodWeatherCache.Lock()
	defer gopodWeatherCache.Unlock()
	var out []string
	for _, c := range gopodWeatherCache.places {
		if c.facts.Country != "" && time.Since(c.at) < 2*time.Minute {
			out = append(out, c.facts.Country)
		}
	}
	return out
}

// gopodGeoPlace is one place OpenWeatherMap's geocoding answers with.
type gopodGeoPlace struct {
	Name    string  `json:"name"`
	State   string  `json:"state"`
	Country string  `json:"country"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

// gopodPickPlace: the place among got whose state or country is one the question named after the
// city ("windsor ontario" -> State "Ontario"), or -1.
func gopodPickPlace(got []gopodGeoPlace, place string) int {
	if len(got) == 0 {
		return -1
	}
	words := strings.Fields(strings.ToLower(strings.ReplaceAll(place, ",", " ")))
	if len(words) < 2 {
		return 0
	}
	region := strings.Join(words[1:], " ")
	for i, g := range got {
		state, country := strings.ToLower(g.State), strings.ToLower(g.Country)
		if state != "" && (strings.Contains(region, state) || strings.Contains(state, region)) {
			return i
		}
		if country != "" && (region == country || strings.HasSuffix(region, " "+country)) {
			return i
		}
	}
	return -1
}

// gopodRobotWeatherFormat: this robot's own unit and clock, from his Wire-Pod RobotSettings
// (temp_is_fahrenheit, clock_24_hour); Celsius/24 h when unread.
func gopodRobotWeatherFormat(esn string) (fahrenheit, clock24 bool) {
	jdoc, ok := vars.GetJdoc("vic:"+esn, "vic.RobotSettings")
	if !ok {
		return false, true
	}
	var s struct {
		TempIsFahrenheit bool `json:"temp_is_fahrenheit"`
		Clock24Hour      bool `json:"clock_24_hour"`
	}
	if json.Unmarshal([]byte(jdoc.JsonDoc), &s) != nil {
		return false, true
	}
	return s.TempIsFahrenheit, s.Clock24Hour
}

// GOLDEN: the one weather call, live chat and scripted draft alike - the real weather line, or
// offline live_weather.offline_text, or "" when that's empty too.
func gopodWeatherSay(esn string) (string, bool) {
	return gopodWeatherSayAt(esn, "")
}

// gopodWeatherSayAt: the same, at a place the question named ("" = the usual).
func gopodWeatherSayAt(esn, place string) (string, bool) {
	if line := gopodLiveWeatherLine(esn, place); line != "" {
		return line, true
	}
	if place != "" {
		if line := gopodLiveWeatherLine(esn, ""); line != "" {
			return line, true
		}
	}
	return gopodBrobotsSettings.LiveWeather.OfflineText, false
}

// gopodLiveWeatherLine: the spoken weather for this robot at place, or "" offline.
func gopodLiveWeatherLine(esn, place string) string {
	facts, err := gopodLiveWeatherFacts(place)
	if err != nil {
		return ""
	}
	fahrenheit, _ := gopodRobotWeatherFormat(esn)
	c := int(math.Round(facts.TempC))
	f := int(math.Round(facts.TempC*9/5 + 32))
	units := fmt.Sprintf("%d° Celsius or %d° Fahrenheit", c, f)
	if fahrenheit {
		units = fmt.Sprintf("%d° Fahrenheit or %d° Celsius", f, c)
	}
	where, clock := gopodPlaceAndClock(esn, facts)
	return fmt.Sprintf("%s, %s, %d, %s, and %s, as of %s.",
		where, facts.Local.Format("January 2"), facts.Local.Year(), units, facts.Description, clock)
}

// gopodWeatherReportLine: the weather line in a weather_lines format - "report" (first or only) or
// "confirm" (next) - in this robot's own unit and clock; "" when that format is empty or there's no
// weather (the golden line then).
func gopodWeatherReportLine(esn, place, kind string) string {
	format := strings.TrimSpace(gopodBrobotsSettings.WeatherLines[kind])
	if format == "" {
		return ""
	}
	facts, err := gopodLiveWeatherFacts(place)
	if err != nil && place != "" {
		facts, err = gopodLiveWeatherFacts("")
	}
	if err != nil {
		return ""
	}
	fahrenheit, clock24 := gopodRobotWeatherFormat(esn)
	temp := fmt.Sprintf("%d° Celsius", int(math.Round(facts.TempC)))
	if fahrenheit {
		temp = fmt.Sprintf("%d° Fahrenheit", int(math.Round(facts.TempC*9/5+32)))
	}
	clock := facts.Local.Format("3:04 PM")
	if clock24 {
		clock = facts.Local.Format("15:04")
		if label := gopodBrobotsSettings.LiveWeather.MilitaryLabel; label != "" {
			clock = facts.Local.Format("1504") + " " + label
		}
	}
	where, _ := gopodPlaceAndClock(esn, facts)
	return strings.NewReplacer("{place}", where, "{date}", fmt.Sprintf("%s, %d", facts.Local.Format("January 2"), facts.Local.Year()),
		"{temp}", temp, "{sky}", facts.Description, "{time}", clock).Replace(format)
}

// gopodPlaceAndClock: the place ("Windsor Ontario Canada 🇨🇦") and the time ("1405 Hours
// Military Time or 2:05 PM") as the weather line says them - shared with PLAYHEAD's time and place
// line (gopod_managers.go).
func gopodPlaceAndClock(esn string, facts gopodWeatherFacts) (string, string) {
	_, clock24 := gopodRobotWeatherFormat(esn)
	h24, h12 := facts.Local.Format("15:04"), facts.Local.Format("3:04 PM")
	if label := gopodBrobotsSettings.LiveWeather.MilitaryLabel; label != "" {
		h24 = facts.Local.Format("1504") + " " + label
	}
	clock := h24 + " or " + h12
	if !clock24 {
		clock = h12 + " or " + h24
	}
	where := facts.Name
	if facts.State != "" && !strings.HasPrefix(strings.ToLower(facts.State), strings.ToLower(facts.Name)) {
		where += " " + facts.State
	}
	if facts.Country != "" {
		if name := gopodCountryName(facts.Country); name != "" {
			where += " " + name
		}
		where += " " + gopodFlagFor(facts.Country)
	}
	return where, clock
}

var gopodDigitWords = []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}
var gopodTeenWords = map[int]string{10: "ten", 11: "eleven", 12: "twelve", 13: "thirteen", 14: "fourteen",
	15: "fifteen", 16: "sixteen", 17: "seventeen", 18: "eighteen", 19: "nineteen"}
var gopodTensWords = map[int]string{2: "twenty", 3: "thirty", 4: "forty", 5: "fifty"}

func gopodTwoDigitWord(n int) string {
	if n < 20 {
		return gopodTeenWords[n]
	}
	word := gopodTensWords[n/10]
	if n%10 != 0 {
		word += "-" + gopodDigitWords[n%10]
	}
	return word
}

// gopodMilitaryTimeSpoken: "twenty twenty-five", "zero eight hundred".
func gopodMilitaryTimeSpoken(hour, minute int) string {
	hw := gopodTwoDigitWord(hour)
	if hour == 0 {
		hw = "zero"
	} else if hour < 10 {
		hw = "zero-" + gopodDigitWords[hour]
	}
	if minute == 0 {
		return hw + " hundred"
	}
	if minute < 10 {
		return hw + " zero-" + gopodDigitWords[minute]
	}
	return hw + " " + gopodTwoDigitWord(minute)
}

// gopodTwelveHourTimeSpoken: "eight twenty-five pm", "eight o'clock 'a' m".
func gopodTwelveHourTimeSpoken(hour24, minute int) string {
	hour := hour24 % 12
	if hour == 0 {
		hour = 12
	}
	hw := gopodTeenWords[hour]
	if hour < 10 {
		hw = gopodDigitWords[hour]
	}
	mw := gopodTwoDigitWord(minute)
	if minute == 0 {
		mw = "o'clock"
	} else if minute < 10 {
		mw = "oh " + gopodDigitWords[minute]
	}
	suffix := "pm"
	if hour24 < 12 {
		suffix = "'a' m"
	}
	return hw + " " + mw + " " + suffix
}

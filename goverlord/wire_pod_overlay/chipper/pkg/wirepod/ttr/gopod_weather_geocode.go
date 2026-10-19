package wirepod_ttr

import (
	"net/http"
	"net/url"
	"encoding/json"
	"io"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// geocodeOpenWeatherMap and geocodeCandidates: split out of weather.go so the
// only thing that file's own overlay edit still needs to carry is the panic-
// to-soft-fail swaps and the one-line call-site change to use these.
func geocodeOpenWeatherMap(location string, apiKey string) []openWeatherMapAPIGeoCodingStruct {
	for _, candidate := range geocodeCandidates(location) {
		geoURL := "http://api.openweathermap.org/geo/1.0/direct?q=" + url.QueryEscape(candidate) + "&limit=1&appid=" + apiKey
		resp, err := http.Get(geoURL)
		if err != nil {
			logger.Println(err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var result []openWeatherMapAPIGeoCodingStruct
		if err := json.Unmarshal(body, &result); err != nil {
			logger.Println(err)
			logger.Println("Geolocation API error: " + string(body))
			continue
		}
		if len(result) > 0 {
			if candidate != location {
				logger.Println("Geocode matched on reformatted candidate: `" + candidate + "`")
			}
			return result
		}
	}
	return nil
}

// geocodeCandidates builds ordered comma-grouping guesses for a spoken
// location phrase: as given, fully space-joined (no commas), comma before
// the final word only (2-part: city+region), then comma between every word
// (3+ part: city, state, country). Duplicates are dropped as word count
// shrinks the candidate set (e.g. a 1-word location only ever has one form).
func geocodeCandidates(location string) []string {
	fields := strings.Fields(strings.ReplaceAll(location, ",", " "))
	if len(fields) == 0 {
		return []string{location}
	}
	seen := make(map[string]bool)
	var candidates []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			return
		}
		seen[s] = true
		candidates = append(candidates, s)
	}
	add(location)
	add(strings.Join(fields, " "))
	if len(fields) >= 2 {
		add(strings.Join(fields[:len(fields)-1], " ") + ", " + fields[len(fields)-1])
	}
	if len(fields) >= 3 {
		add(strings.Join(fields, ", "))
	}
	return candidates
}

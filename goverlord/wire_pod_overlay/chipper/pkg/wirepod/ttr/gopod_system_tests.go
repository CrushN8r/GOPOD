// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_tests.go
package wirepod_ttr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// system-test-* (2026-10-18, operator: "build a full set of
// system-[test-commands] to test individually, like weather offline, etc... *
// and be able to chain system-[test-commands] into a comprehensive
// audit/checklist run"). Each test checks one thing in the running Wire-Pod -
// nothing is said, nothing moves, no bubble but the answer - and answers PASS
// or FAIL with what it saw. Typed on the chat page's keyboard:
//   system-test-list              the tests
//   system-test-<name>            one test
//   system-test-a system-test-b   several, in that order
//   system-test-all               the audit: every test, in system_tests.audit's order

type gopodSystemTest struct {
	name, about string
	run         func() (bool, string)
}

var gopodSystemTests = []gopodSystemTest{
	{"weather-live", "the weather service answers for the usual place", func() (bool, string) {
		f, err := gopodLiveWeatherFacts("")
		if err != nil {
			return false, err.Error()
		}
		return true, fmt.Sprintf("%s, %s", f.Description, f.Local.Format("15:04"))
	}},
	{"weather-offline", "with no weather service, the offline line is said", func() (bool, string) {
		line := strings.TrimSpace(gopodBrobotsSettings.LiveWeather.OfflineText)
		if line == "" {
			return false, "live_weather.offline_text is empty"
		}
		if vars.APIConfig.Weather.Key == "" {
			return true, "no weather key now - every weather line is: " + line
		}
		return true, "offline line: " + line
	}},
	{"time-place", "PLAYHEAD's time and place line", func() (bool, string) {
		line := gopodTimePlaceLine("")
		return line != "" && line != gopodBrobotsSettings.LiveWeather.OfflineText, line
	}},
	{"takes", "the managers take the right lines", func() (bool, string) {
		if gopodScriptPlaying.Load() || gopodRunActive() {
			return false, "a run is playing - takes are live only, try again after it"
		}
		cases := map[string]string{"what's the weather": "weather", "where are we now": "time_place", "show me flare4.com": "url_qr", "tell me a joke": ""}
		var bad []string
		for line, want := range cases {
			t, _, ok := gopodManagerTakes(line)
			if got := map[bool]string{true: t.Kind, false: ""}[ok]; got != want {
				bad = append(bad, fmt.Sprintf("%q took %q, not %q", line, got, want))
			}
		}
		if len(bad) > 0 {
			return false, strings.Join(bad, "; ")
		}
		return true, fmt.Sprintf("%d lines, each to the right take", len(cases))
	}},
	{"qr", "a QR code is made for a web address", func() (bool, string) {
		page, err := gopodMakeQR("https://gopod-system-test.example")
		if err != nil {
			return false, err.Error()
		}
		file := filepath.Join(gopodChipperRoot(), "webroot", page)
		ok := gopodFileExists(file)
		os.Remove(file)
		return ok, page
	}},
	{"python", "Python runs with the operator's libraries (qrcode, vosk)", func() (bool, string) {
		out, err := gopodAsOperator(exec.Command("python3", "-c", "import qrcode, vosk; print('ok')")).CombinedOutput()
		if err != nil {
			return false, strings.TrimSpace(string(out))
		}
		return true, "qrcode and vosk import"
	}},
	{"voices", "the voice worker is ready and voices are timed by Vosk", func() (bool, string) {
		gopodReporter.Lock()
		ready := gopodReporter.ready
		gopodReporter.Unlock()
		files, _ := filepath.Glob(filepath.Join(gopodReporterAudioDir(), "*.words.json"))
		vosk := 0
		for _, f := range files {
			if b, err := os.ReadFile(f); err == nil && strings.Contains(string(b), `"aligner": "vosk"`) {
				vosk++
			}
		}
		return ready && vosk == len(files), fmt.Sprintf("worker ready %v, %d of %d voices timed by Vosk", ready, vosk, len(files))
	}},
	{"personas", "CHALK and PLAYHEAD have their prompts and who lines", func() (bool, string) {
		var bad []string
		for name, p := range gopodBrobotsSettings.Managers.Personas {
			if strings.TrimSpace(p.WhoLine) == "" || strings.TrimSpace(gopodPersonaPrompt(p.Host)) == "" {
				bad = append(bad, name)
			}
		}
		if len(bad) > 0 || len(gopodBrobotsSettings.Managers.Personas) == 0 {
			return false, "missing: " + strings.Join(bad, ", ")
		}
		return true, fmt.Sprintf("%d personas", len(gopodBrobotsSettings.Managers.Personas))
	}},
	{"eggs", "the egg list: chapters (less not_eggs), no surprise eggs, no weather", func() (bool, string) {
		list := gopodEggList()
		for _, e := range list {
			if e.Key == "weather" || gopodNotEgg(e.Key) {
				return false, e.Key + " is on the list"
			}
			for _, n := range gopodBrobotsSettings.TopicPictures.NotListed {
				if e.Key == n {
					return false, n + " (a surprise egg) is on the list"
				}
			}
		}
		return true, fmt.Sprintf("%d eggs on the list", len(list))
	}},
	{"llm", "the language model answers (Ollama has the model)", func() (bool, string) {
		k := vars.APIConfig.Knowledge
		if k.Endpoint == "" || k.Model == "" {
			return false, "no knowledge endpoint or model in Wire-Pod's settings"
		}
		resp, err := (&http.Client{Timeout: 5 * time.Second}).Get(gopodOllamaBase(k.Endpoint) + "/api/tags")
		if err != nil {
			return false, err.Error()
		}
		defer resp.Body.Close()
		var tags struct {
			Models []struct {
				Name string `json:"name"`
			} `json:"models"`
		}
		json.NewDecoder(resp.Body).Decode(&tags)
		for _, m := range tags.Models {
			if m.Name == k.Model || strings.TrimSuffix(m.Name, ":latest") == k.Model {
				return true, k.Model
			}
		}
		return false, k.Model + " isn't in Ollama"
	}},
	{"search", "GOPOD's own web search answers", func() (bool, string) {
		got, err := gopodSearch("GOPOD Wire-Pod")
		if err != nil {
			return false, err.Error()
		}
		return len(got) > 0, fmt.Sprintf("%d results", len(got))
	}},
	{"files", "system-check: every file named is there, story.md / knobs.json / zKnobs.json match", func() (bool, string) {
		problems, checked := gopodSystemCheckProblems()
		if len(problems) > 0 {
			return false, fmt.Sprintf("%d to fix (system-check), first: %s", len(problems), problems[0])
		}
		return true, fmt.Sprintf("%d things checked", checked)
	}},
	{"page", "the chat page is open and here", func() (bool, string) {
		return gopodPageHere(), map[bool]string{true: "here", false: "not here (open it, click Log)"}[gopodPageHere()]
	}},
	// 2026-10-19 (operator: "start with the score and the audio test"; pha0b's
	// audio-routing check): the operator's sound defaults are GOPOD's speaker
	// and mic (system_tests.audio) - robots can log "say=success" with nothing heard.
	{"audio", "sound goes to GOPOD's speaker and comes from its mic", func() (bool, string) {
		a := gopodBrobotsSettings.SystemTests.Audio
		uid := os.Getenv("SUDO_UID")
		if _, u, _ := gopodRecordRoot(); u > 0 {
			uid = fmt.Sprint(u)
		}
		get := func(what string) string {
			cmd := exec.Command("pactl", "get-default-"+what)
			if os.Geteuid() == 0 && uid != "" {
				cmd = exec.Command("setpriv", "--reuid="+uid, "--regid="+uid, "--init-groups", "pactl", "get-default-"+what)
			}
			cmd.Env = append(os.Environ(), "XDG_RUNTIME_DIR=/run/user/"+uid)
			out, err := cmd.Output()
			if err != nil {
				return "(" + err.Error() + ")"
			}
			return strings.TrimSpace(string(out))
		}
		sink, source := get("sink"), get("source")
		ok := sink == a.Sink && strings.Contains(source, a.SourceHas)
		saw := "speaker " + sink + ", mic " + source
		if !ok {
			saw += " - fix: pactl set-default-sink " + a.Sink + " (and the USB mic as the source)"
		}
		return ok, saw
	}},
	{"robots", "both Brobots answered the battery watch in the last 10 minutes", func() (bool, string) {
		gopodBatteryLast.Lock()
		defer gopodBatteryLast.Unlock()
		var out []string
		ok := true
		for _, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN} {
			at, seen := gopodBatteryLast.at[esn]
			if !seen || time.Since(at) > 10*time.Minute {
				ok = false
				out = append(out, gopodBrobotName(esn)+" not heard")
			} else {
				out = append(out, gopodBrobotName(esn)+" "+gopodBatteryLast.text[esn])
			}
		}
		return ok, strings.Join(out, "; ")
	}},
}

// gopodSystemTestMatch: the line is system-test commands.
func gopodSystemTestMatch(text string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(text)), "system-test")
}

// gopodSystemTestRun: the answer to system-test lines.
func gopodSystemTestRun(text string) string {
	byName := map[string]gopodSystemTest{}
	for _, t := range gopodSystemTests {
		byName[t.name] = t
	}
	var names []string
	for _, w := range strings.Fields(strings.ToLower(text)) {
		w = strings.TrimPrefix(w, "system-test-")
		switch w {
		case "system-test", "list":
			out := []string{"system-test-<name>, several in a row, or system-test-all:"}
			for _, t := range gopodSystemTests {
				out = append(out, fmt.Sprintf("  %-16s %s", t.name, t.about))
			}
			return gopodJoinLines(out)
		case "all":
			names = append(names, gopodBrobotsSettings.SystemTests.Audit...)
			if len(gopodBrobotsSettings.SystemTests.Audit) == 0 {
				for _, t := range gopodSystemTests {
					names = append(names, t.name)
				}
			}
		default:
			names = append(names, w)
		}
	}
	out := []string{}
	pass := 0
	for _, n := range names {
		t, ok := byName[n]
		if !ok {
			out = append(out, "?    "+n+": no such test (system-test-list)")
			continue
		}
		good, saw := t.run()
		mark := "FAIL"
		if good {
			mark = "PASS"
			pass++
		}
		out = append(out, fmt.Sprintf("%s %s: %s", mark, n, saw))
	}
	head := fmt.Sprintf("system-test: %d of %d pass", pass, len(names))
	return gopodJoinLines(append([]string{head}, out...))
}

// ## EOF

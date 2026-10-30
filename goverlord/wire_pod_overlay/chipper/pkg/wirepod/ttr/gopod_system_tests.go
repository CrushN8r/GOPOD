// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_tests.go
package wirepod_ttr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
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
		// 2026-10-25 (audit: 12 untimed voices, all from 2026-10-18, before forced
		// alignment, unused since): the voices this start uses must be timed by
		// Vosk; an old one is made again, timed, the next time it's used.
		gopodReporterInUse.Lock()
		inUse := map[string]bool{}
		for name := range gopodReporterInUse.files {
			inUse[name] = true
		}
		gopodReporterInUse.Unlock()
		files, _ := filepath.Glob(filepath.Join(gopodReporterAudioDir(), "*.wav"))
		used, timed, old := 0, 0, 0
		for _, f := range files {
			var times struct {
				Aligner string `json:"aligner"`
			}
			b, err := os.ReadFile(strings.TrimSuffix(f, ".wav") + ".words.json")
			vosk := err == nil && json.Unmarshal(b, &times) == nil && times.Aligner == "vosk"
			if !inUse[filepath.Base(f)] {
				if !vosk {
					old++
				}
				continue
			}
			used++
			if vosk {
				timed++
			}
		}
		line := gopodSystemText("system_tests.voices_timed", "ready", fmt.Sprint(ready), "timed", strconv.Itoa(timed), "used", strconv.Itoa(used))
		if old > 0 {
			line += fmt.Sprintf(" (%d old ones not timed - made again when next used)", old)
		}
		return ready && timed == used, line
	}},
	{"personas", "CHALK and PLAYHEAD have their prompts and who lines", func() (bool, string) {
		var bad []string
		for name, p := range gopodBrobotsSettings.Managers.Personas {
			if strings.TrimSpace(p.WhoLine) == "" || strings.TrimSpace(gopodPersonaPrompt(p.Host)) == "" {
				bad = append(bad, name)
			}
		}
		if len(bad) > 0 || len(gopodBrobotsSettings.Managers.Personas) == 0 {
			return false, gopodSystemText("system_tests.missing", "missing", strings.Join(bad, ", "))
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
	// 2026-10-24 (operator: "Test every: Topic, Value Point, Easter Egg * fullest run
	// without live brobots"): gopod_game_sweep.go.
	{"game", "every topic, value point and Easter egg is found by its own words (no robots, no model)", func() (bool, string) {
		return gopodGameSweepLine(gopodGameSweep())
	}},
	// 2026-10-24 (operator: "Add into GOPOD system menu * The replay ... plays the
	// real chat page code with no robots"): gopod_probes/tools/gopod_page_replay.sh.
	// 2026-10-28 (operator: "Make the 100% admin run a standing self-check * each
	// point not revealed * each egg not found * each prize not given * each
	// reminder that went dry"): gopod_admin_run.go.
	{"admin-run", "the 100% admin run, every line in order through the game: every value point, egg and prize, and no dry reminder (no robots, no model; the live game is put back)", gopodAdminRunLine},
	// 2026-10-29 (operator: "create a scripted 'real questions' run to test with *
	// then test if its working well enough"; "I like the name 'probers list'"):
	// gopod_probers_list.go.
	{"probers-list", "the probers list - the operator's own probers from his saved sessions, through the topic loader: the right topic first, lightning against the word triggers (no robots, no model)", gopodProbersListLine},
	{"eggvoice", "an Easter egg voice is made from its kept pieces and a live title, joined", gopodEggVoiceCheck},
	{"replay", "the chat page plays every egg kind, both prizes and the Operator voices, and a visitor's page mirrors it, with no robots", func() (bool, string) {
		return gopodPageReplayLine(gopodPageReplayRun())
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
		get := func(what string) string {
			out, err := gopodPactl("get-default-" + what)
			if err != nil {
				return "(" + err.Error() + ")"
			}
			return out
		}
		sink, source := get("sink"), get("source")
		ok := sink == a.Sink && strings.Contains(source, a.SourceHas)
		saw := "speaker " + sink + ", mic " + source
		if sink != a.Sink {
			saw += " - fix: pactl set-default-sink " + a.Sink
		}
		if !strings.Contains(source, a.SourceHas) {
			saw += " - fix: pactl set-default-source <the USB mic's name> (pactl list short sources | grep " + a.SourceHas + ")"
		}
		return ok, saw
	}},
	// 2026-10-20 (operator: "Edge TPU: the Edge TPU worker isn't ready * how to
	// test or flip on in system menu?"): the worker is up and names what's in a
	// known picture (a copy in webroot/visual, so the boxes copy lands there).
	{"tpu", "the Edge TPU worker is ready and names what's in a picture", func() (bool, string) {
		gopodQuickLook.Lock()
		ready := gopodQuickLook.ready
		gopodQuickLook.Unlock()
		if !ready {
			return false, "the worker isn't ready - system-tpu restart"
		}
		// 2026-10-24 (operator: "PASS tpu: 1 objects: laptop * laptop?" - the
		// workbook photo, named a laptop): a known picture and what must be in it
		// (system_tests.tpu) - PASS only when the TPU names it.
		tt := gopodBrobotsSettings.SystemTests.Tpu
		src := filepath.Join(gopodChipperRoot(), "webroot", "pictures", tt.Picture)
		dst := filepath.Join(gopodVisualDir(), "tpu_test.jpg")
		gopodMkdirAll(gopodVisualDir(), 0o755)
		data, err := os.ReadFile(src)
		if err != nil || gopodWriteFile(dst, data, 0o644) != nil {
			return false, "no test picture"
		}
		objects, err := gopodQuickLookObjects(dst)
		if err != nil {
			return false, err.Error()
		}
		saw := fmt.Sprintf("%d objects: %s", len(objects), strings.Join(objects, ", "))
		for _, o := range objects {
			if strings.EqualFold(strings.TrimSpace(o), tt.Expect) {
				return true, saw + " (" + tt.Picture + ": " + tt.Expect + " expected)"
			}
		}
		return false, saw + " - " + tt.Picture + " should show a " + tt.Expect
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
	head := gopodSystemText("system_tests.head", "pass", strconv.Itoa(pass), "tests", strconv.Itoa(len(names)))
	return gopodJoinLines(append([]string{head}, out...))
}

// gopodPageReplayRun: the page replay's report (its output), or why it didn't run.
func gopodPageReplayRun() (string, error) {
	script := filepath.Join(gopodChipperRoot(), "gopod_probes", "tools", "gopod_page_replay.sh")
	out, err := gopodAsOperator(exec.Command("bash", script)).CombinedOutput()
	return string(out), err
}

// gopodPageReplayLine: the replay in one line - pass only when every line played,
// every egg drawn and lit, every Operator line tied to its voice, and no error,
// stuck line or loop.
func gopodPageReplayLine(out string, err error) (bool, string) {
	if err != nil && !strings.Contains(out, "PLAYED ") {
		return false, gopodSystemText("system_tests.didnt_run", "why", strings.TrimSpace(err.Error()+" "+gopodCut(strings.TrimSpace(out), 200)))
	}
	var problems []string
	played, eggs, ops, section, mirror := "", 0, 0, "", ""
	for _, l := range strings.Split(out, "\n") {
		t := strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(t, "PLAYED "):
			played = strings.TrimPrefix(t, "PLAYED ")
			if strings.Contains(t, "STOPPED") {
				problems = append(problems, "the page stopped at "+played)
			}
		case strings.HasPrefix(t, "LOOP"), strings.HasPrefix(t, "LOAD ERROR"):
			problems = append(problems, t)
		case t == "EGGS", t == "LIGHTING ORDER", t == "GOVERLORD LINES", t == "BUBBLES BY KIND", t == "VISITOR MIRROR", strings.HasPrefix(t, "OPERATOR LINES"), strings.HasPrefix(t, "PAGE NOTES"):
			section = t
		// 2026-10-24 (link 6): the visitor's page against the main page.
		case section == "VISITOR MIRROR":
			if mirror == "" {
				mirror = t
			}
			if strings.Contains(t, "OUT OF STEP") || strings.HasPrefix(t, "behind") {
				problems = append(problems, "mirror: "+gopodCut(t, 140))
			}
		case t == "":
		case section == "EGGS":
			eggs++
			if strings.Contains(t, "MISSING") || strings.Contains(t, "Generating left 1") || !strings.Contains(t, "congratulations spoken") {
				problems = append(problems, "egg: "+gopodCut(t, 120))
			}
		case strings.HasPrefix(section, "OPERATOR LINES"):
			ops++
			if !strings.HasSuffix(t, "spoken") {
				problems = append(problems, "Operator line: "+gopodCut(t, 100))
			}
		case strings.HasPrefix(section, "PAGE NOTES"):
			problems = append(problems, "page: "+gopodCut(t, 120))
		}
	}
	if played == "" {
		return false, gopodSystemText("system_tests.no_report", "out", gopodCut(strings.TrimSpace(out), 200))
	}
	line := fmt.Sprintf("%s lines played, %d eggs drawn and lit, %d Operator lines tied to their voice", played, eggs, ops)
	if mirror != "" {
		line += ", visitor mirror " + strings.TrimPrefix(mirror, " ")
	}
	if len(problems) > 0 {
		return false, line + " - " + strings.Join(problems, "; ")
	}
	return true, line + ", no errors"
}

// gopodPactl: the operator's sound system (pactl), run as the operator - Wire-Pod
// runs as root, the sound belongs to the desktop user.
func gopodPactl(args ...string) (string, error) {
	uid := os.Getenv("SUDO_UID")
	if _, u, _ := gopodRecordRoot(); u > 0 {
		uid = fmt.Sprint(u)
	}
	// As the operator (gopodAsOperator, one place since 2026-10-26).
	cmd := gopodAsOperator(exec.Command("pactl", args...))
	cmd.Env = append(cmd.Env, "XDG_RUNTIME_DIR=/run/user/"+uid)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// gopodSetAudioRouting (wpr's own first step, _gopod_set_audio_routing): the
// speaker and the USB mic as the defaults (system_tests.audio's names).
func gopodSetAudioRouting() string {
	a := gopodBrobotsSettings.SystemTests.Audio
	if a.Sink == "" {
		return gopodSystemText("system_tests.no_speaker_named_system_tests")
	}
	if _, err := gopodPactl("set-default-sink", a.Sink); err != nil {
		return "speaker not set: " + err.Error()
	}
	list, _ := gopodPactl("list", "short", "sources")
	for _, l := range strings.Split(list, "\n") {
		f := strings.Fields(l)
		if len(f) > 1 && strings.Contains(f[1], a.SourceHas) && !strings.Contains(f[1], ".monitor") {
			if _, err := gopodPactl("set-default-source", f[1]); err != nil {
				return "mic not set: " + err.Error()
			}
			return gopodSystemText("system_tests.speaker_and_mic_set")
		}
	}
	return gopodSystemText("system_tests.speaker_set_no_mic_with", "source_has", a.SourceHas)
}

// gopodAsOperatorUser: a program run as the operator himself (his user, not only
// his home) when Wire-Pod runs as root - for anything that writes in his home
// (the audit's Go test cache).
func gopodAsOperatorUser(name string, args ...string) *exec.Cmd {
	uid := os.Getenv("SUDO_UID")
	if _, u, _ := gopodRecordRoot(); u > 0 {
		uid = fmt.Sprint(u)
	}
	// As the operator (gopodAsOperator, one place since 2026-10-26).
	cmd := gopodAsOperator(exec.Command(name, args...))
	cmd.Env = append(cmd.Env, "XDG_RUNTIME_DIR=/run/user/"+uid, "PATH=/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin")
	return cmd
}

// gopodSystemAudit (2026-10-25, operator: "In GOPOD system please add ... new
// auditing upgrades"): system-audit runs the full audit
// (gopod_probes/tools/gopod_full_audit.sh, no robots called) and posts its
// bottom line and every failed check when it's done.
var gopodAuditRunning atomic.Bool

func gopodSystemAudit() string {
	if !gopodAuditRunning.CompareAndSwap(false, true) {
		return gopodSystemText("system_tests.system_audit_is_already_running")
	}
	script := filepath.Join(gopodChipperRoot(), "gopod_probes", "tools", "gopod_full_audit.sh")
	go func() {
		defer gopodAuditRunning.Store(false)
		out, _ := gopodAsOperatorUser("bash", script).CombinedOutput()
		var fails []string
		last := ""
		for _, l := range strings.Split(string(out), "\n") {
			t := strings.TrimSpace(l)
			if strings.HasPrefix(t, "FAIL ") {
				fails = append(fails, gopodCut(strings.Join(strings.Fields(strings.TrimPrefix(t, "FAIL ")), " "), 200))
			}
			if strings.HasPrefix(t, "PASS:") || strings.HasPrefix(t, "FAIL:") {
				last = t
			}
		}
		if last == "" {
			last = "FAIL: the audit didn't finish - " + gopodCut(strings.TrimSpace(string(out)), 200)
		}
		answer := gopodSystemText("system_tests.audit_last", "last", last)
		if len(fails) > 0 {
			answer += "\n" + strings.Join(fails, "\n")
		}
		logger.Println("GOPOD_SYSTEM_AUDIT " + strings.ReplaceAll(answer, "\n", " | "))
		GopodChatFeedAdd(gopodSystemSetRole(), answer)
	}()
	return gopodSystemText("system_tests.system_audit_started_everything_but")
}

// ## EOF

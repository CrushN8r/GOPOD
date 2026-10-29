// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_edit.go
package wirepod_ttr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The editing commands (2026-10-18, operator: "I want to be able to edit
// everything possible in GOPOD * aim for maximum automation, minimal editing";
// "Build all 5 system- commands please"). Typed on the chat page's keyboard;
// every write is kept for system-undo, and a settings edit that would stop
// Wire-Pod from starting is put back at once.
//   system-set <path.to.setting> <value>       any setting by its path
//   system-add <path> <item>                   an item into a list (or "key = value" into a map)
//   system-remove <path> <item>                an item out of a list (or a key out of a map)
//   system-lingo <brobots|reporters|butler> <word> = <said>   (remove <word>; or just the lingo: its list)
//   system-restart [go]                        Wire-Pod restarts; the page reloads itself
//   system-topic [<topic> [words|point|label|reminder|pictures ...]]   a chapter topic

func gopodSettingsFilePath() string {
	return filepath.Join(gopodChipperRoot(), "webroot", "gopod_brobots_settings.json")
}

// gopodSettingsStillLoad: the settings file as it now is would load at Wire-Pod's start.
func gopodSettingsStillLoad() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	loadGOPODBrobotsSettings()
	return nil
}

// gopodUndoForget: the open command's writes were put back - nothing to undo.
func gopodUndoForget() {
	gopodUndo.Lock()
	if gopodUndo.open != nil {
		gopodUndo.open.Files = nil
	}
	gopodUndo.Unlock()
}

// gopodSettingsWrite: one edit of the settings file, put back if it wouldn't load.
func gopodSettingsWrite(edit func(file string) (string, error)) (string, error) {
	file := gopodSettingsFilePath()
	before, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	old, err := edit(file)
	if err != nil {
		return "", err
	}
	if lerr := gopodSettingsStillLoad(); lerr != nil {
		gopodWriteRaw(file, before)
		gopodUndoForget()
		return "", errors.New("put back - Wire-Pod wouldn't start with it: " + lerr.Error())
	}
	return old, nil
}

const gopodAfterRestart = " (takes effect after system-restart)"

// gopodSettingsPath: "managers.takes.0.take_text" -> its keys.
func gopodSettingsPath(dotted string) []string {
	return strings.Split(strings.Trim(dotted, "."), ".")
}

// --- 1. system-set <path> <value> ---

// gopodSystemSetPath: any setting by its path; the value keeps the old one's kind.
func gopodSystemSetPath(dotted, value string) string {
	path := gopodSettingsPath(dotted)
	old, exists := gopodSettingsGet(path)
	if value == "" {
		if !exists {
			return "no setting " + dotted + " (system-locate <word> finds one)."
		}
		return dotted + " = " + old
	}
	raw := value
	switch {
	case exists && strings.HasPrefix(old, `"`):
		b, _ := json.Marshal(value)
		raw = string(b)
	case json.Valid([]byte(value)):
	default:
		b, _ := json.Marshal(value)
		raw = string(b)
	}
	if exists && (strings.HasPrefix(old, "{") || strings.HasPrefix(old, "[")) && !(strings.HasPrefix(raw, "{") || strings.HasPrefix(raw, "[")) {
		return dotted + " is a list or a group - system-add / system-remove change its items, or system-set " + dotted + ".<key> one value."
	}
	was, err := gopodSettingsWrite(func(file string) (string, error) { return gopodJSONFileSet(file, path, raw) })
	if err != nil {
		return dotted + ": not changed - " + err.Error()
	}
	return dotted + ": " + gopodShort(was, 80) + " -> " + gopodShort(raw, 80) + gopodAfterRestart
}

func gopodShort(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "..."
	}
	return s
}

// --- 2. system-add / system-remove ---

func gopodSystemAddRemove(text string) string {
	f := strings.Fields(text)
	add := strings.EqualFold(f[0], "system-add")
	if len(f) < 3 {
		return f[0] + " <path> <item> - e.g. " + f[0] + " managers.who_words what's your name"
	}
	dotted, item := f[1], strings.TrimSpace(strings.Join(f[2:], " "))
	path := gopodSettingsPath(dotted)
	old, ok := gopodSettingsGet(path)
	if !ok {
		return "no setting " + dotted + "."
	}
	var result string
	_, err := gopodSettingsWrite(func(file string) (string, error) {
		switch {
		case strings.HasPrefix(old, "["):
			var list []any
			if json.Unmarshal([]byte(old), &list) != nil {
				return "", errors.New("its list doesn't read")
			}
			for _, x := range list {
				if _, isText := x.(string); !isText {
					if _, isNum := x.(float64); !isNum {
						return "", errors.New("a list of groups - edit it with system-set " + dotted + ".<n>.<key>")
					}
				}
			}
			var out []any
			found := false
			for _, x := range list {
				if strings.EqualFold(fmt.Sprint(x), item) {
					found = true
					if !add {
						continue
					}
				}
				out = append(out, x)
			}
			if add && found {
				return "", errors.New("\"" + item + "\" is already in it")
			}
			if !add && !found {
				return "", errors.New("\"" + item + "\" isn't in it")
			}
			if add {
				var v any = item
				if n, err := strconv.ParseFloat(item, 64); err == nil && len(list) > 0 {
					if _, isNum := list[0].(float64); isNum {
						v = n
					}
				}
				out = append(out, v)
			}
			parts := make([]string, 0, len(out))
			for _, x := range out {
				b, _ := json.Marshal(x)
				parts = append(parts, string(b))
			}
			result = fmt.Sprintf("%d items", len(out))
			return gopodJSONFileSet(file, path, "["+strings.Join(parts, ", ")+"]")
		case strings.HasPrefix(old, "{"):
			key, value := item, ""
			if i := strings.IndexAny(item, "=:"); i > 0 {
				key, value = strings.TrimSpace(item[:i]), strings.TrimSpace(item[i+1:])
			}
			if add {
				if value == "" {
					return "", errors.New("give it as key = value")
				}
				b, _ := json.Marshal(value)
				result = key + " = " + value
				return gopodJSONFileSet(file, append(append([]string{}, path...), key), string(b))
			}
			result = key + " taken out"
			return gopodJSONFileDelete(file, append(append([]string{}, path...), key))
		}
		return "", errors.New("it's one value - system-set " + dotted + " <value>")
	})
	if err != nil {
		return f[0] + " " + dotted + ": not changed - " + err.Error()
	}
	verb := "added"
	if !add {
		verb = "removed"
	}
	return dotted + ": \"" + item + "\" " + verb + " (" + result + ")" + gopodAfterRestart
}

// gopodJSONFileDelete takes a key and its value out of an object, in place.
func gopodJSONFileDelete(file string, path []string) (string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	s, e, ok := gopodJSONSpan(data, path)
	if !ok {
		return "", errors.New("no " + path[len(path)-1] + " in it")
	}
	old := string(data[s:e])
	key, _ := json.Marshal(path[len(path)-1])
	ks := bytes.LastIndex(data[:s], key)
	if ks < 0 {
		return "", errors.New("its key isn't found")
	}
	end := e
	rest := bytes.TrimLeft(data[end:], " \t\r\n")
	if len(rest) > 0 && rest[0] == ',' {
		end = len(data) - len(rest) + 1
		// and the space up to the next key, so the next key takes its place
		next := bytes.TrimLeft(data[end:], " \t\r\n")
		end = len(data) - len(next)
		lineStart := bytes.LastIndexByte(data[:ks], '\n') + 1
		if len(bytes.TrimSpace(data[lineStart:ks])) == 0 {
			ks = lineStart
			end = bytes.LastIndexByte(data[:len(data)-len(next)], '\n') + 1
			if end <= ks {
				end = len(data) - len(next)
			}
		}
	} else {
		before := bytes.TrimRight(data[:ks], " \t\r\n")
		if len(before) > 0 && before[len(before)-1] == ',' {
			ks = len(before) - 1
		}
	}
	out := append(append([]byte{}, data[:ks]...), data[end:]...)
	if !json.Valid(out) {
		return "", errors.New("the edit didn't read back - left as it was")
	}
	return old, gopodWriteFileSafe(file, out)
}

// --- 3. system-lingo ---

var gopodLingoEntryRE = regexp.MustCompile(`^(\s*)\{"from":\s*("(?:[^"\\]|\\.)*"),\s*"to":\s*("(?:[^"\\]|\\.)*")(.*?)\}(,?)\s*$`)

func gopodLingoWhich(name string) (string, []string) {
	switch strings.ToLower(name) {
	case "brobots", "brobot", "party":
		return "brobots", nil
	case "reporters", "reporter":
		return "reporters", []string{"reporter_voices", "say_words"}
	case "butler", "system", "goverlord":
		return "butler", []string{"reporter_voices", "system_say_words"}
	}
	return "", nil
}

func gopodSystemLingo(text string) string {
	rest := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(text), strings.Fields(text)[0]))
	f := strings.Fields(rest)
	if len(f) == 0 {
		return "system-lingo <brobots|reporters|butler> <word> = <said> / remove <word> / (just the lingo: its list)"
	}
	lingo, path := gopodLingoWhich(f[0])
	if lingo == "" {
		return "system-lingo: brobots, reporters or butler."
	}
	rest = strings.TrimSpace(rest[len(f[0]):])
	remove := false
	if strings.HasPrefix(strings.ToLower(rest), "remove ") {
		remove, rest = true, strings.TrimSpace(rest[len("remove "):])
	}
	word, said := rest, ""
	if i := strings.Index(rest, "="); i >= 0 {
		word, said = strings.TrimSpace(rest[:i]), strings.TrimSpace(rest[i+1:])
	}
	if word == "" {
		return gopodLingoList(lingo, path)
	}
	if !remove && said == "" {
		return gopodLingoList(lingo, path) // a word with no "=": its entry shows in the list
	}
	if lingo == "brobots" {
		return gopodLingoBrobots(word, said, remove)
	}
	key := append(append([]string{}, path...), word)
	var err error
	if remove {
		_, err = gopodSettingsWrite(func(file string) (string, error) { return gopodJSONFileDelete(file, key) })
	} else {
		b, _ := json.Marshal(said)
		_, err = gopodSettingsWrite(func(file string) (string, error) { return gopodJSONFileSet(file, key, string(b)) })
	}
	if err != nil {
		return "system-lingo " + lingo + ": not changed - " + err.Error()
	}
	if remove {
		return lingo + " lingo: " + word + " taken out" + gopodAfterRestart
	}
	return lingo + " lingo: " + word + " = " + said + gopodAfterRestart
}

func gopodLingoList(lingo string, path []string) string {
	var lines []string
	if lingo == "brobots" {
		data, _ := os.ReadFile(gopodPronunciationRegistryPath)
		for _, l := range strings.Split(string(data), "\n") {
			if m := gopodLingoEntryRE.FindStringSubmatch(l); m != nil {
				var from, to string
				json.Unmarshal([]byte(m[2]), &from)
				json.Unmarshal([]byte(m[3]), &to)
				lines = append(lines, from+" = "+to)
			}
		}
	} else {
		raw, _ := gopodSettingsGet(path)
		var m map[string]string
		json.Unmarshal([]byte(raw), &m)
		for k, v := range m {
			lines = append(lines, k+" = "+v)
		}
		sort.Strings(lines)
	}
	out := []string{fmt.Sprintf("%s lingo - %d words:", lingo, len(lines))}
	for i, l := range lines {
		if i == 60 {
			out = append(out, fmt.Sprintf("  ... %d more", len(lines)-60))
			break
		}
		out = append(out, "  "+l)
	}
	return gopodJoinLines(out)
}

// gopodLingoBrobots: an entry in the Brobots' pronunciation list (the template card).
func gopodLingoBrobots(word, said string, remove bool) string {
	file := gopodPronunciationRegistryPath
	data, err := os.ReadFile(file)
	if err != nil {
		return "system-lingo brobots: " + err.Error()
	}
	lines := strings.Split(string(data), "\n")
	wordJSON, _ := json.Marshal(word)
	saidJSON, _ := json.Marshal(said)
	start, at := -1, -1
	for i, l := range lines {
		if start < 0 && strings.Contains(l, `"pronunciation_entries": [`) {
			start = i
		}
		if start >= 0 && at < 0 {
			if m := gopodLingoEntryRE.FindStringSubmatch(l); m != nil && m[2] == string(wordJSON) {
				at = i
			}
		}
	}
	if start < 0 {
		return "system-lingo brobots: the pronunciation list isn't found in " + filepath.Base(file)
	}
	switch {
	case remove && at < 0:
		return "brobots lingo: no " + word + " in it."
	case remove:
		m := gopodLingoEntryRE.FindStringSubmatch(lines[at])
		lines = append(lines[:at], lines[at+1:]...)
		if m[5] == "" && at > 0 { // it was the last: the one before ends the list now
			lines[at-1] = strings.TrimSuffix(strings.TrimRight(lines[at-1], " "), ",")
		}
	case at >= 0:
		m := gopodLingoEntryRE.FindStringSubmatch(lines[at])
		lines[at] = m[1] + `{"from": ` + m[2] + `, "to": ` + string(saidJSON) + m[4] + "}" + m[5]
	default:
		indent := "    "
		if start+1 < len(lines) {
			if m := gopodLingoEntryRE.FindStringSubmatch(lines[start+1]); m != nil {
				indent = m[1]
			}
		}
		entry := indent + `{"from": ` + string(wordJSON) + `, "to": ` + string(saidJSON) + `},`
		lines = append(lines[:start+1], append([]string{entry}, lines[start+1:]...)...)
	}
	out := []byte(strings.Join(lines, "\n"))
	if m := gopodPronunciationBlockPattern.FindSubmatch(out); m == nil || !json.Valid(m[1]) {
		return "system-lingo brobots: not changed - the list wouldn't read"
	}
	if err := gopodWriteFileSafe(file, out); err != nil {
		return "system-lingo brobots: " + err.Error()
	}
	if remove {
		return "brobots lingo: " + word + " taken out" + gopodAfterRestart
	}
	return "brobots lingo: " + word + " = " + said + gopodAfterRestart
}

// --- system-native / system-numpad (2026-10-19, operator: "make Native?[]
// Numpad?[] part of system- commands * prune Native?[] & Numpad?[] altogether") ---

func gopodSystemNative(text string) string {
	f := strings.Fields(strings.ToLower(text))
	if len(f) > 1 && (f[1] == "on" || f[1] == "off") {
		GopodNativeSet(f[1] == "on")
	}
	state := map[bool]string{true: "on", false: "off"}[GopodNativeOn()]
	return "Native voice answers (Wire-Pod's own: \"Hey Vector, what time is it\" and the rest): " + state + " - system-native on|off"
}

// gopodSystemNumpad: the numpad is the page's; the page switches it as the line is sent.
func gopodSystemNumpad(text string) string {
	f := strings.Fields(strings.ToLower(text))
	if len(f) < 2 {
		return gopodSystemNumpadList()
	}
	if f[1] != "on" && f[1] != "off" {
		return "system-numpad on|off - the numpad keys as push-to-talk; alone: each key."
	}
	if f[1] == "on" {
		return "Numpad on: " + gopodBrobotsSettings.NumpadPTT.Texts["on"]
	}
	return "Numpad off: " + gopodBrobotsSettings.NumpadPTT.Texts["exit"]
}

// --- system-score (2026-10-19, operator: "start with the score and the audio
// test"; pha0b's score) ---

// gopodSystemScore: a song's score - every step, its note and its full text.
func gopodSystemScore(text string) string {
	f := strings.Fields(text)
	if len(f) < 2 {
		songs, _ := gopodAllSongs()
		var names []string
		for _, dir := range songs {
			names = append(names, filepath.Base(dir))
		}
		sort.Strings(names)
		return "system-score <song> - " + strings.Join(names, ", ")
	}
	dir, err := gopodSystemSetSong(f[1])
	if err != nil {
		return err.Error()
	}
	steps, _, err := gopodSystemSetKnobsFile(dir)
	if err != nil {
		return err.Error()
	}
	texts := gopodSystemSetStory(dir)
	out := []string{fmt.Sprintf("%s - %d steps:", filepath.Base(dir), len(steps))}
	for i, s := range steps {
		line := fmt.Sprintf("%3d %s (%s)", i+1, s.StepID, s.Note)
		if t := strings.TrimSpace(texts[s.StepID]); t != "" {
			line += "\n      " + strings.ReplaceAll(t, "\n", "\n      ")
		}
		out = append(out, line)
	}
	return gopodJoinLines(out)
}

// --- 4. system-restart ---

// gopodSystemRestart: Wire-Pod restarts itself a moment after the answer shows;
// the page sees the new start and reloads itself (click Log once for sound).
// 2026-10-25 (operator: "forced wpr 2 restart, from what I just gave you"): what
// wpr 2 does - the speaker and USB mic set, the page's leftover run state
// cleared, then a forced restart (a playing run too). Only settings Wire-Pod
// couldn't start with stop it.
func gopodSystemRestart(text string) string {
	if lerr := gopodSettingsStillLoad(); lerr != nil {
		return "Not restarted - Wire-Pod wouldn't start with the settings as they are: " + lerr.Error()
	}
	audio := gopodSetAudioRouting()
	logger.Println("GOPOD_SYSTEM_RESTART audio: " + audio)
	// wpr 2's own reset: the rich display dormant on the next page load.
	_ = gopodWriteFile(filepath.Join(gopodChipperRoot(), "webroot", "gopod_rich_display_ui_state.json"), []byte(`{"expanded":false,"run_active":false}`+"\n"), 0o644)
	go func() {
		time.Sleep(2 * time.Second)
		logger.Println("GOPOD_SYSTEM_RESTART restarting wire-pod (forced, like wpr 2)")
		if out, err := exec.Command("systemctl", "restart", "--no-block", "wire-pod").CombinedOutput(); err != nil {
			logger.Println("GOPOD_SYSTEM_RESTART failed: " + err.Error() + " " + strings.TrimSpace(string(out)))
		}
	}()
	return "Restarting Wire-Pod (forced, like wpr 2; " + audio + ") - the page reloads itself when it's back; click Log once for sound."
}

// --- 5. system-topic ---

func gopodTopicFind(name string) (string, bool) {
	for _, p := range gopodTopicRegistry() {
		if strings.EqualFold(p.Name, name) {
			return p.File, true
		}
	}
	return "", false
}

var gopodTopicSayRE = regexp.MustCompile(`^say (\d+):\s*(.*)$`)
var gopodTopicPointRE = regexp.MustCompile(`^(\d+)\.\s+(.*)$`)

func gopodSystemTopic(text string) string {
	f := strings.Fields(text)
	if len(f) == 1 {
		var names []string
		for _, p := range gopodTopicRegistry() {
			names = append(names, p.Name)
		}
		return "system-topic <topic> [words add|remove <word> | words <a, b, c> | point <text> | point <n> <text> | point <n> remove | label|reminder|pictures <text>]\ntopics: " + strings.Join(names, ", ")
	}
	file, ok := gopodTopicFind(f[1])
	if !ok {
		return "no topic " + f[1] + " - system-topic lists them."
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return err.Error()
	}
	lines := strings.Split(string(data), "\n")
	sep := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == "---" {
			sep = i
			break
		}
	}
	if sep < 0 {
		sep = len(lines)
	}
	if len(f) == 2 {
		return f[1] + " (" + filepath.Base(filepath.Dir(file)) + "/" + filepath.Base(file) + "):\n" + strings.Join(lines[:sep], "\n")
	}
	what := strings.ToLower(f[2])
	rest := strings.TrimSpace(strings.Join(f[3:], " "))
	// The topic-cluster format's own reader and writer (gopod_chapters_folder.go).
	headSet := func(key, value string) {
		for i := 0; i < sep; i++ {
			if k, _, ok := gopodTopicHeadKV(lines[i]); ok && k == key {
				lines[i] = gopodTopicHeadLine(key, value)
				return
			}
		}
		at := 0
		if key != "label" {
			for at < sep && strings.TrimSpace(lines[at]) != "" && !strings.HasPrefix(strings.TrimSpace(lines[at]), "- ") {
				at++
			}
			for at < sep && strings.TrimSpace(lines[at]) == "" {
				at++
			}
		}
		lines = append(lines[:at], append([]string{gopodTopicHeadLine(key, value)}, lines[at:]...)...)
		sep++
	}
	headGet := func(key string) string {
		for i := 0; i < sep; i++ {
			if k, v, ok := gopodTopicHeadKV(lines[i]); ok && k == key {
				return v
			}
		}
		return ""
	}
	answer := ""
	switch what {
	case "label", "reminder", "pictures":
		if rest == "" {
			return f[1] + " " + what + ": " + headGet(what)
		}
		headSet(what, rest)
		answer = what + ": " + rest
	case "words":
		words := []string{}
		for _, w := range strings.Split(headGet("words"), ",") {
			if w = strings.TrimSpace(w); w != "" {
				words = append(words, w)
			}
		}
		g := strings.Fields(rest)
		switch {
		case len(g) > 1 && (strings.EqualFold(g[0], "add") || strings.EqualFold(g[0], "remove")):
			w := strings.TrimSpace(strings.Join(g[1:], " "))
			var out []string
			found := false
			for _, x := range words {
				if strings.EqualFold(x, w) {
					found = true
					if strings.EqualFold(g[0], "remove") {
						continue
					}
				}
				out = append(out, x)
			}
			if strings.EqualFold(g[0], "add") && !found {
				out = append(out, w)
			}
			words = out
		case rest != "":
			words = nil
			for _, w := range strings.Split(rest, ",") {
				if w = strings.TrimSpace(w); w != "" {
					words = append(words, w)
				}
			}
		default:
			return f[1] + " words: " + strings.Join(words, ", ")
		}
		headSet("words", strings.Join(words, ", "))
		answer = "words: " + strings.Join(words, ", ")
	case "point":
		// The say lines (top) and the numbered points (after ---) stay the same list.
		var says []string
		for i := 0; i < sep; i++ {
			if k, v, ok := gopodTopicHeadKV(lines[i]); ok && gopodTopicSayRE.MatchString(k+":") {
				says = append(says, v)
			}
		}
		g := strings.Fields(rest)
		n, numErr := 0, errors.New("")
		if len(g) > 0 {
			n, numErr = strconv.Atoi(g[0])
		}
		switch {
		case rest == "":
			out := []string{f[1] + " points:"}
			for i, s := range says {
				out = append(out, fmt.Sprintf("  %d. %s", i+1, s))
			}
			return gopodJoinLines(out)
		case numErr == nil && len(g) == 2 && strings.EqualFold(g[1], "remove"):
			if n < 1 || n > len(says) {
				return fmt.Sprintf("%s has %d points.", f[1], len(says))
			}
			says = append(says[:n-1], says[n:]...)
			answer = fmt.Sprintf("point %d taken out", n)
		case numErr == nil && len(g) > 1:
			if n < 1 || n > len(says) {
				return fmt.Sprintf("%s has %d points.", f[1], len(says))
			}
			says[n-1] = strings.TrimSpace(strings.Join(g[1:], " "))
			answer = fmt.Sprintf("point %d: %s", n, says[n-1])
		default:
			says = append(says, rest)
			answer = fmt.Sprintf("point %d: %s", len(says), rest)
		}
		lines = gopodTopicPointsWrite(lines, sep, says)
	default:
		return "system-topic " + f[1] + " words | point | label | reminder | pictures ..."
	}
	if err := gopodWriteFileSafe(file, []byte(strings.Join(lines, "\n"))); err != nil {
		return "system-topic: " + err.Error()
	}
	return f[1] + " " + answer + gopodAfterRestart
}

// gopodTopicPointsWrite: the say lines and the numbered points, both from says.
func gopodTopicPointsWrite(lines []string, sep int, says []string) []string {
	var head, body []string
	sayAt := -1
	for i := 0; i < sep; i++ {
		if k, _, ok := gopodTopicHeadKV(lines[i]); ok && gopodTopicSayRE.MatchString(k+":") {
			if sayAt < 0 {
				sayAt = len(head)
			}
			continue
		}
		head = append(head, lines[i])
	}
	if sayAt < 0 {
		sayAt = len(head)
	}
	var sayLines []string
	for i, s := range says {
		sayLines = append(sayLines, gopodTopicHeadLine(fmt.Sprintf("say %d", i+1), s))
	}
	head = append(head[:sayAt], append(sayLines, head[sayAt:]...)...)
	pointAt := -1
	for i := sep + 1; i < len(lines); i++ {
		if gopodTopicPointRE.MatchString(strings.TrimSpace(lines[i])) {
			if pointAt < 0 {
				pointAt = len(body)
			}
			continue
		}
		body = append(body, lines[i])
	}
	if sep >= len(lines) {
		body = []string{"Reveal these, in order, in your own words:"}
	}
	if pointAt < 0 {
		pointAt = len(body)
		for pointAt > 0 && strings.TrimSpace(body[pointAt-1]) == "" {
			pointAt--
		}
	}
	var points []string
	for i, s := range says {
		points = append(points, fmt.Sprintf("%d. %s", i+1, s))
	}
	body = append(body[:pointAt], append(points, body[pointAt:]...)...)
	// A blank line on both sides of "---" (markdown reads a line right above it as a heading).
	for len(head) > 0 && strings.TrimSpace(head[len(head)-1]) == "" {
		head = head[:len(head)-1]
	}
	for len(body) > 0 && strings.TrimSpace(body[0]) == "" {
		body = body[1:]
	}
	return append(append(head, "", "---", ""), body...)
}

// ## EOF

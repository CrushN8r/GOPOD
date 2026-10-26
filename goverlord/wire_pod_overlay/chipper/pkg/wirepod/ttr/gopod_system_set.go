// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_set.go
package wirepod_ttr

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// system-set (2026-10-17, operator: "automate the editing of all song pieces *
// like ask Goverlord to do it via the chat window * similar to start run-name,
// use \"system-set\" * \"system-set\" connects to alias play studio * I type:
// system-set this-volume 13%"; "send \"system-set\" from web page keyboard as
// a terminal-command * terminal-results are sent back to chat bubbles window *
// In terminal-results chat bubble, make terminal-bubble-text font-size 50% of
// current bubble-text font-size"). Typed on the chat page's keyboard only. It
// runs a fixed list of edits - never a terminal command (operator: "good
// catch"): a named setting (gopodSystemSetKnobs), or one field of one song
// step (gopodSystemSetFields), written into gopod_brobots_settings.json, the
// song's knobs.json (and zKnobs.json) or story.md. Each file is edited in
// place, only that value, and checked before it's saved. The answer - what
// changed, old and new, and when it takes effect - shows in a terminal bubble
// (system_set.role).
//
//   system-set                         the list
//   system-set master-volume 35%       a named setting
//   system-set <song>                  the song's steps
//   system-set <song> <step>           one step
//   system-set <song> <step> volume 13%
//   system-set <song> <step> text New words here.

type gopodSystemSetKnob struct {
	path  []string
	kind  string // "master" (0-100% -> 0-1), "share" ("" or "40%"), "db"
	when  string
	about string
}

var gopodSystemSetKnobs = map[string]gopodSystemSetKnob{
	"master-volume":  {[]string{"chat_audio", "volume"}, "master", "F5", "every sound's master volume"},
	"normalize":      {[]string{"chat_audio", "normalize_db"}, "db", "wpr 2", "the level every sound is evened to (-3db, off)"},
	"loop-volume":    {[]string{"background_loop", "volume"}, "share", "F5", "the background loop"},
	"weather-volume": {[]string{"weather_card", "volume"}, "share", "F5", "the weather sound"},
	"key-volume":     {[]string{"numpad_ptt", "key_sound", "volume"}, "share", "F5", "the numpad key sound"},
	"sting-volume":   {[]string{"reporter_voices", "system_stings", "default", "volume"}, "share", "wpr 2", "the system sting"},
}

// gopodSystemSetFields: the step fields system-set writes, and their kind.
var gopodSystemSetFields = map[string]string{
	"volume": "share", "text": "text",
	"seconds": "number", "pause_seconds": "number", "hold_seconds": "number", "buffer_after": "number",
	"settle_seconds": "number", "speed": "number", "cycles": "number",
	// 2026-10-19 (tempo-set): a step's tempo_factor (default 1.0).
	"tempo_factor": "number",
}

// gopodSystemSetMatch: the line is a system-set command.
func gopodSystemSetMatch(text string) bool {
	word := strings.ToLower(strings.TrimSpace(text))
	if gopodSystemTestMatch(text) {
		return true
	}
	for _, cmd := range []string{"system-set", "system-locate", "system-set-list", "system-locate-list", "system-undo", "system-status", "system-help", "system-find", "system-search", "system-check", "system-voices-rebuild", "system-add", "system-remove", "system-lingo", "system-restart", "system-topic", "system-native", "system-numpad", "system-score", "system-play", "system-gap", "system-tweaks", "system-calibrate", "system-robot", "system-faces", "system-photos", "system-version", "system-visual", "system-cams", "system-chats", "system-chat-save", "system-menu-theme", "system-tpu", "system-brobot", "system-usb", "system-anim", "system-value-points", "system-interview-rebuild", "system-audit", "system-voices", "system-voices-report"} {
		if word == cmd || strings.HasPrefix(word, cmd+" ") {
			return true
		}
	}
	return false
}

// gopodSystemSetLine shows the typed line and its answer in a terminal bubble.
func gopodSystemSetLine(esn, text, how string) {
	to := gopodBrobotName(esn)
	if esn == gopodTypedAll {
		to = "Brobots"
	}
	gopodOperatorShow(to, text, how)
	answer := ""
	if how != "Typed text" {
		answer = "system commands work from the chat page's keyboard only."
	} else if cmd := strings.ToLower(strings.Fields(text)[0]); cmd == "system-undo" {
		answer = gopodSystemUndo()
	} else if cmd == "system-status" {
		answer = gopodSystemStatus()
	} else if cmd == "system-search" {
		answer = gopodSystemSearchList(text)
	} else if cmd == "system-find" {
		answer = gopodSystemFind(text)
	} else if cmd == "system-help" {
		answer = gopodSystemHelp()
	} else if cmd == "system-check" {
		answer = gopodSystemCheck()
	} else if gopodSystemTestMatch(text) {
		answer = gopodSystemTestRun(text)
	} else if cmd == "system-add" || cmd == "system-remove" {
		gopodUndoBegin(strings.TrimSpace(text))
		answer = gopodSystemAddRemove(text)
		gopodUndoEnd()
	} else if cmd == "system-lingo" {
		gopodUndoBegin(strings.TrimSpace(text))
		answer = gopodSystemLingo(text)
		gopodUndoEnd()
	} else if cmd == "system-topic" {
		gopodUndoBegin(strings.TrimSpace(text))
		answer = gopodSystemTopic(text)
		gopodUndoEnd()
	} else if cmd == "system-voices-report" {
		answer = gopodSystemVoicesReport(text)
	} else if cmd == "system-voices" {
		answer = gopodSystemVoices()
	} else if cmd == "system-value-points" {
		gopodUndoBegin(strings.TrimSpace(text))
		answer = gopodSystemValuePoints(text)
		gopodUndoEnd()
	} else if cmd == "system-interview-rebuild" {
		// 2026-10-23 (A5): the interview song from the game (gopod_interview_from_game.go).
		gopodUndoBegin(strings.TrimSpace(text))
		answer = gopodInterviewRebuild()
		gopodUndoEnd()
	} else if cmd == "system-anim" {
		answer = gopodSystemAnim(text)
	} else if cmd == "system-usb" {
		answer = gopodSystemUSB()
	} else if cmd == "system-tpu" {
		answer = gopodSystemTPU(text)
	} else if cmd == "system-cams" {
		answer = gopodSystemCams(text)
	} else if cmd == "system-chats" {
		answer = gopodSystemChats(text)
	} else if cmd == "system-chat-save" {
		answer = gopodSystemChatSave(text)
	} else if cmd == "system-menu-theme" {
		gopodUndoBegin(strings.TrimSpace(text))
		answer = gopodSystemMenuTheme(text)
		gopodUndoEnd()
	} else if cmd == "system-visual" {
		answer = gopodSystemVisual(text)
	} else if cmd == "system-robot" || cmd == "system-brobot" {
		answer = gopodSystemRobot(text)
	} else if cmd == "system-faces" {
		answer = gopodSystemFaces(text)
	} else if cmd == "system-photos" {
		answer = gopodSystemPhotos(text)
	} else if cmd == "system-version" {
		answer = gopodSystemVersion()
	} else if cmd == "system-play" {
		answer = gopodSystemPlay(text)
	} else if cmd == "system-gap" {
		answer = gopodSystemGap(text)
	} else if cmd == "system-tweaks" {
		answer = gopodSystemTweaks(text)
	} else if cmd == "system-calibrate" {
		answer = gopodSystemCalibrate(text)
	} else if cmd == "system-score" {
		answer = gopodSystemScore(text)
	} else if cmd == "system-native" {
		answer = gopodSystemNative(text)
	} else if cmd == "system-numpad" {
		answer = gopodSystemNumpad(text)
	} else if cmd == "system-restart" {
		answer = gopodSystemRestart(text)
	} else if cmd == "system-audit" {
		answer = gopodSystemAudit()
	} else if cmd == "system-voices-rebuild" {
		answer = gopodSystemVoicesRebuild(text)
	} else if cmd == "system-set-list" {
		// 2026-10-17 (operator: "add these, to list the items in set & locate in
		// the response bubble * for fast reference"): the lists.
		answer = gopodSystemSetHelp()
	} else if cmd == "system-locate-list" {
		answer = gopodSystemLocateList()
	} else if strings.HasPrefix(strings.ToLower(strings.TrimSpace(text)), "system-locate") {
		answer = gopodSystemLocate(text)
	} else {
		// 2026-10-17: its writes are kept for system-undo (gopod_system_more.go).
		gopodUndoBegin(strings.TrimSpace(text))
		answer = gopodSystemSet(text)
		gopodUndoEnd()
	}
	logger.Println("GOPOD_SYSTEM_SET " + strings.ReplaceAll(answer, "\n", " | "))
	role := gopodSystemSetRole()
	// 2026-10-19 (operator: "system-search = location = PLAYHEAD, not CHALK"):
	// its results are a search results box (system_search.results_host draws it).
	if r := gopodBrobotsSettings.SystemSearch.ResultsRole; r != "" && strings.EqualFold(strings.Fields(text)[0], "system-search") && how == "Typed text" {
		role = r
	}
	GopodChatFeedAdd(role, answer)
}

// gopodSystemSetTempo (2026-10-19, tempo-set's mode A): the song's global_tempo
// (extra seconds per step, 0 = none; a step's tempo_factor weighs it).
func gopodSystemSetTempo(dir, value string) string {
	if v, err := strconv.ParseFloat(value, 64); err != nil || v < 0 {
		return "system-set <song> tempo <seconds, 0 or more>"
	}
	var out []string
	for _, name := range []string{"knobs.json", "zKnobs.json"} {
		path := filepath.Join(dir, name)
		if !gopodFileExists(path) {
			continue
		}
		old, err := gopodJSONFileSet(path, []string{"global_tempo"}, value)
		if err != nil {
			return name + ": not changed - " + err.Error()
		}
		out = append(out, name+" "+old+" -> "+value)
	}
	return filepath.Base(dir) + " tempo: " + strings.Join(out, ", ") + " (the next run of it)"
}

func gopodSystemSetRole() string {
	if r := strings.TrimSpace(gopodBrobotsSettings.SystemSet.Role); r != "" {
		return r
	}
	return "Terminal"
}

// gopodSystemSet runs one system-set line and says what it did.
func gopodSystemSet(text string) string {
	args := strings.Fields(strings.TrimSpace(text))[1:]
	if len(args) == 0 || strings.EqualFold(args[0], "help") || strings.EqualFold(args[0], "list") {
		return gopodSystemSetHelp()
	}
	name := strings.ToLower(args[0])
	// 2026-10-18 (operator: "edit everything possible"): any setting by its path.
	if _, knob := gopodSystemSetKnobs[name]; !knob && strings.Contains(args[0], ".") {
		return gopodSystemSetPath(args[0], strings.Join(args[1:], " "))
	}
	if k, ok := gopodSystemSetKnobs[name]; ok {
		if len(args) < 2 {
			old, _ := gopodSettingsGet(k.path)
			return name + " = " + old + " (" + k.about + ")"
		}
		return gopodSystemSetKnobWrite(name, k, strings.Join(args[1:], " "))
	}
	dir, err := gopodSystemSetSong(args[0])
	if err != nil {
		return err.Error() + "\n" + gopodSystemSetHelp()
	}
	switch len(args) {
	case 1:
		return gopodSystemSetSteps(dir)
	case 2:
		return gopodSystemSetStep(dir, args[1])
	case 3:
		// 2026-10-19 (tempo-set): the song's global_tempo, in knobs.json and zKnobs.json.
		if strings.EqualFold(args[1], "tempo") {
			return gopodSystemSetTempo(dir, args[2])
		}
		return "Give it a value: system-set " + args[0] + " " + args[1] + " " + args[2] + " <value>"
	}
	return gopodSystemSetStepWrite(dir, args[1], strings.ToLower(args[2]), strings.Join(args[3:], " "))
}

func gopodSystemSetHelp() string {
	var names []string
	for n := range gopodSystemSetKnobs {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	b.WriteString("system-set - settings:\n")
	for _, n := range names {
		old, _ := gopodSettingsGet(gopodSystemSetKnobs[n].path)
		b.WriteString("  " + n + " = " + old + "  (" + gopodSystemSetKnobs[n].about + ")\n")
	}
	songs, _ := gopodAllSongs()
	var ids []string
	for id, dir := range songs {
		ids = append(ids, filepath.Base(dir)+" ("+id+")")
	}
	sort.Strings(ids)
	b.WriteString("songs: " + strings.Join(ids, ", ") + "\n")
	var fields []string
	for f := range gopodSystemSetFields {
		fields = append(fields, f)
	}
	sort.Strings(fields)
	b.WriteString("step fields: " + strings.Join(fields, ", ") + "\n")
	b.WriteString("e.g. system-set master-volume 35% | system-set 00_llm_warming | system-set 00_llm_warming goverlord_sting_0106 volume 50%\n")
	b.WriteString("where a thing lives: system-locate, system-locate-list | undo: system-undo | all commands: system-help")
	return b.String()
}

// gopodSystemSetValue checks and writes a value as JSON for its kind.
func gopodSystemSetValue(kind, value string) (string, error) {
	v := strings.TrimSpace(value)
	switch kind {
	case "master":
		n, err := strconv.ParseFloat(strings.TrimSuffix(v, "%"), 64)
		if err != nil || !strings.HasSuffix(v, "%") || n < 0 || n > 100 {
			return "", errors.New("give a percentage, 0-100% (35%)")
		}
		return strconv.FormatFloat(n/100, 'f', -1, 64), nil
	case "share":
		if strings.EqualFold(v, "default") {
			v = ""
		}
		if v != "" {
			n, err := strconv.ParseFloat(strings.TrimSuffix(v, "%"), 64)
			if err != nil || !strings.HasSuffix(v, "%") || n < 0 {
				return "", errors.New("give \"default\" or a percentage of the master (40%)")
			}
		}
		b, _ := json.Marshal(v)
		return string(b), nil
	case "db":
		if strings.EqualFold(v, "off") {
			return "0", nil
		}
		n, err := strconv.ParseFloat(strings.TrimSuffix(strings.ToLower(v), "db"), 64)
		if err != nil || n > 0 || n < -40 {
			return "", errors.New("give a level from -40db to 0db (-3db), or off")
		}
		return strconv.FormatFloat(n, 'f', -1, 64), nil
	case "number":
		n, err := strconv.ParseFloat(v, 64)
		if err != nil || n < 0 {
			return "", errors.New("give a number (1.5)")
		}
		return strconv.FormatFloat(n, 'f', -1, 64), nil
	}
	return "", errors.New("not a value system-set writes")
}

func gopodSystemSetKnobWrite(name string, k gopodSystemSetKnob, value string) string {
	raw, err := gopodSystemSetValue(k.kind, value)
	if err != nil {
		return name + ": " + err.Error()
	}
	path := filepath.Join(gopodChipperRoot(), "webroot", "gopod_brobots_settings.json")
	old, err := gopodJSONFileSet(path, k.path, raw)
	if err != nil {
		return name + ": not changed - " + err.Error()
	}
	return name + ": " + old + " -> " + raw + " (" + k.about + "; takes effect after " + k.when + ")"
}

// gopodSettingsGet: a setting's value as written in the settings file.
func gopodSettingsGet(path []string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(gopodChipperRoot(), "webroot", "gopod_brobots_settings.json"))
	if err != nil {
		return "", false
	}
	s, e, ok := gopodJSONSpan(data, path)
	if !ok {
		return "(none)", false
	}
	return string(data[s:e]), true
}

// gopodSystemSetSong: a song folder by its folder name or its song_id.
func gopodSystemSetSong(name string) (string, error) {
	songs, err := gopodAllSongs()
	if err != nil {
		return "", err
	}
	if dir, ok := songs[name]; ok {
		return dir, nil
	}
	for _, dir := range songs {
		if strings.EqualFold(filepath.Base(dir), name) {
			return dir, nil
		}
	}
	return "", errors.New("no setting or song named " + name + ".")
}

type gopodSetStep struct {
	StepID string            `json:"step_id"`
	Note   string            `json:"note"`
	Combo  []json.RawMessage `json:"combo"`
}

func gopodSystemSetKnobsFile(dir string) ([]gopodSetStep, []byte, error) {
	data, err := os.ReadFile(filepath.Join(dir, "knobs.json"))
	if err != nil {
		return nil, nil, err
	}
	return gopodSystemSetKnobsFileData(data)
}

// gopodSystemSetKnobsFileData: a knobs file's steps.
func gopodSystemSetKnobsFileData(data []byte) ([]gopodSetStep, []byte, error) {
	var k struct {
		Steps []gopodSetStep `json:"steps"`
	}
	if err := json.Unmarshal(data, &k); err != nil {
		return nil, nil, errors.New("knobs.json: " + err.Error())
	}
	return k.Steps, data, nil
}

func gopodSystemSetSteps(dir string) string {
	steps, _, err := gopodSystemSetKnobsFile(dir)
	if err != nil {
		return err.Error()
	}
	texts := gopodSystemSetStory(dir)
	var b strings.Builder
	b.WriteString(filepath.Base(dir) + " - " + strconv.Itoa(len(steps)) + " steps:\n")
	for _, s := range steps {
		line := "  " + s.StepID + " (" + s.Note + ")"
		if t := texts[s.StepID]; t != "" {
			if len(t) > 50 {
				t = gopodCut(t, 50)
			}
			line += " " + t
		}
		b.WriteString(line + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func gopodSystemSetStory(dir string) map[string]string {
	data, err := os.ReadFile(filepath.Join(dir, "story.md"))
	if err != nil {
		return map[string]string{}
	}
	texts, _ := gopodInterviewStory(string(data))
	return texts
}

// gopodSystemSetStepIndex: the step's index in a knobs file, or -1.
func gopodSystemSetStepIndex(steps []gopodSetStep, id string) int {
	for i, s := range steps {
		if s.StepID == id {
			return i
		}
	}
	return -1
}

func gopodSystemSetStep(dir, id string) string {
	steps, data, err := gopodSystemSetKnobsFile(dir)
	if err != nil {
		return err.Error()
	}
	i := gopodSystemSetStepIndex(steps, id)
	if i < 0 {
		return "no step " + id + " in " + filepath.Base(dir) + " - system-set " + filepath.Base(dir) + " lists them."
	}
	s, e, _ := gopodJSONSpan(data, []string{"steps", strconv.Itoa(i)})
	out := id + ": " + string(bytes.TrimSpace(data[s:e]))
	if t, ok := gopodSystemSetStory(dir)[id]; ok {
		out += "\ntext: " + t
	}
	return out
}

func gopodSystemSetStepWrite(dir, id, field, value string) string {
	kind, ok := gopodSystemSetFields[field]
	if !ok {
		return field + " isn't a field system-set writes - system-set lists them."
	}
	where := filepath.Base(dir) + " " + id + " " + field
	if kind == "text" {
		old, err := gopodStorySetText(filepath.Join(dir, "story.md"), id, strings.TrimSpace(value))
		if err != nil {
			return where + ": not changed - " + err.Error()
		}
		return where + ": \"" + old + "\" -> \"" + strings.TrimSpace(value) + "\" (" + gopodSystemSetWhen(dir) + ")"
	}
	raw, err := gopodSystemSetValue(kind, value)
	if err != nil {
		return where + ": " + err.Error()
	}
	var changed []string
	old := ""
	for _, name := range []string{"knobs.json", "zKnobs.json"} {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var k struct {
			Steps []gopodSetStep `json:"steps"`
		}
		if json.Unmarshal(data, &k) != nil {
			return where + ": not changed - " + name + " doesn't read"
		}
		i := gopodSystemSetStepIndex(k.Steps, id)
		if i < 0 {
			if name == "knobs.json" {
				return "no step " + id + " in " + filepath.Base(dir) + " - system-set " + filepath.Base(dir) + " lists them."
			}
			continue
		}
		if len(k.Steps[i].Combo) > 0 {
			return where + ": a combo step - its parts are edited in knobs.json for now."
		}
		was, err := gopodJSONFileSet(path, []string{"steps", strconv.Itoa(i), field}, raw)
		if err != nil {
			return where + ": not changed in " + name + " - " + err.Error()
		}
		if old == "" {
			old = was
		}
		changed = append(changed, name)
	}
	return where + ": " + old + " -> " + raw + " (" + strings.Join(changed, ", ") + "; " + gopodSystemSetWhen(dir) + ")"
}

func gopodSystemSetWhen(dir string) string {
	if dir2, err := gopodFindSong(gopodBrobotsSettings.LLM.OpenerSongID); err == nil && dir2 == dir {
		return "takes effect after wpr 2"
	}
	return "takes effect at the next run"
}

// gopodStorySetText writes one step's > TEXT: line in a story.md; the old words back.
func gopodStorySetText(path, id, text string) (string, error) {
	if strings.ContainsAny(text, "\n\r") {
		return "", errors.New("one line of words only")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	old, _ := gopodInterviewStory(string(data))
	if _, ok := old[id]; !ok {
		return "", errors.New("no step " + id + " in story.md")
	}
	lines := strings.Split(string(data), "\n")
	var out []string
	in, done := false, false
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		switch {
		case strings.HasPrefix(line, "## STEP "):
			in = strings.TrimSpace(strings.TrimPrefix(line, "## STEP ")) == id
		case strings.HasPrefix(line, "#"):
			in = false
		case in && !done && strings.HasPrefix(line, "> TEXT:"):
			out = append(out, strings.TrimRight("> TEXT: "+text, " "))
			// The old words' further lines go with them (a "> ..." line, or a
			// line right under them with no blank line between).
			for i+1 < len(lines) {
				next := strings.TrimSpace(lines[i+1])
				if next == "" || strings.HasPrefix(next, "#") || gopodStoryFieldLine(lines[i+1]) {
					break
				}
				i++
			}
			done = true
			continue
		}
		out = append(out, lines[i])
	}
	if !done {
		return "", errors.New("step " + id + " has no > TEXT: line")
	}
	result := strings.Join(out, "\n")
	check, err := gopodInterviewStory(result)
	if err != nil || check[id] != text || len(check) != len(old) {
		return "", errors.New("the edit didn't read back - story.md left as it was")
	}
	if err := gopodWriteFileSafe(path, []byte(result)); err != nil {
		return "", err
	}
	return old[id], nil
}

// gopodStoryFieldLine: a "> FAIL: ..." style line (a field of its own, not more words).
func gopodStoryFieldLine(line string) bool {
	t := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), ">"))
	head, _, ok := strings.Cut(t, ":")
	return ok && head != "" && strings.ToUpper(head) == head && !strings.Contains(head, " ")
}

func gopodWriteFileSafe(path string, data []byte) error {
	info, err := os.Stat(path)
	mode := os.FileMode(0o644)
	if err == nil {
		mode = info.Mode().Perm()
	}
	before, _ := os.ReadFile(path)
	gopodUndoNote(path, before, data)
	tmp := path + ".system-set.tmp"
	if err := gopodWriteFile(tmp, data, mode); err != nil {
		return err
	}
	// 2026-10-21 (the speaker bar's save left the settings file root's, so the
	// operator could no longer edit it): the file keeps its owner.
	if err == nil {
		if st, ok := info.Sys().(*syscall.Stat_t); ok {
			gopodRecordChown(int(st.Uid), int(st.Gid), tmp)
		}
	}
	return os.Rename(tmp, path)
}

// gopodJSONFileSet writes raw (a JSON value) at path in a JSON file, keeping
// everything else byte for byte; a missing last key is added to its object.
// The old value back ("(none)" when added).
func gopodJSONFileSet(file string, path []string, raw string) (string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	if !json.Valid([]byte(raw)) {
		return "", errors.New("not a JSON value")
	}
	var out []byte
	old := "(none)"
	if s, e, ok := gopodJSONSpan(data, path); ok {
		old = string(data[s:e])
		out = append(append(append([]byte{}, data[:s]...), raw...), data[e:]...)
	} else {
		s, e, ok := gopodJSONSpan(data, path[:len(path)-1])
		if !ok || data[s] != '{' {
			return "", errors.New("no " + strings.Join(path[:len(path)-1], ".") + " in " + filepath.Base(file))
		}
		key, _ := json.Marshal(path[len(path)-1])
		closeAt := e - 1
		body := bytes.TrimSpace(data[s+1 : closeAt])
		insert := string(key) + ": " + raw
		if len(body) > 0 {
			// After the object's last value, before its closing space.
			at := s + 1 + bytes.LastIndex(data[s+1:closeAt], body[len(body)-1:]) + 1
			out = append(append(append([]byte{}, data[:at]...), ", "+insert...), data[at:]...)
		} else {
			out = append(append(append([]byte{}, data[:s+1]...), insert...), data[s+1:]...)
		}
	}
	if !json.Valid(out) {
		return "", errors.New("the edit didn't read back - left as it was")
	}
	if s, e, ok := gopodJSONSpan(out, path); !ok || string(out[s:e]) != raw {
		return "", errors.New("the edit didn't read back - left as it was")
	}
	return old, gopodWriteFileSafe(file, out)
}

// gopodJSONSpan: where the value at path (object keys, array indexes) is in data.
func gopodJSONSpan(data []byte, path []string) (int, int, bool) {
	var found [2]int
	ok := false
	same := func(at []string) bool {
		if len(at) != len(path) {
			return false
		}
		for k := range at {
			if at[k] != path[k] {
				return false
			}
		}
		return true
	}
	if gopodJSONWalk(data, func(at []string, start, end int) {
		if same(at) {
			found, ok = [2]int{start, end}, true
		}
	}) != nil {
		return 0, 0, false
	}
	return found[0], found[1], ok
}

// gopodJSONWalk calls visit with every value's path and span in data.
func gopodJSONWalk(data []byte, visit func(path []string, start, end int)) error {
	var walk func(i int, at []string) (int, error)
	skipSpace := func(i int) int {
		for i < len(data) && (data[i] == ' ' || data[i] == '\n' || data[i] == '\t' || data[i] == '\r') {
			i++
		}
		return i
	}
	str := func(i int) (int, string, error) {
		j := i + 1
		for j < len(data) && data[j] != '"' {
			if data[j] == '\\' {
				j++
			}
			j++
		}
		if j >= len(data) {
			return 0, "", errors.New("bad string")
		}
		var s string
		if err := json.Unmarshal(data[i:j+1], &s); err != nil {
			return 0, "", err
		}
		return j + 1, s, nil
	}
	walk = func(i int, at []string) (int, error) {
		i = skipSpace(i)
		if i >= len(data) {
			return 0, errors.New("end")
		}
		start := i
		var end int
		switch data[i] {
		case '{':
			i = skipSpace(i + 1)
			if i < len(data) && data[i] == '}' {
				end = i + 1
				break
			}
			for {
				i = skipSpace(i)
				if i >= len(data) || data[i] != '"' {
					return 0, errors.New("bad object")
				}
				var key string
				var err error
				if i, key, err = str(i); err != nil {
					return 0, err
				}
				i = skipSpace(i)
				if i >= len(data) || data[i] != ':' {
					return 0, errors.New("bad object")
				}
				if i, err = walk(i+1, append(append([]string{}, at...), key)); err != nil {
					return 0, err
				}
				i = skipSpace(i)
				if i < len(data) && data[i] == ',' {
					i++
					continue
				}
				if i < len(data) && data[i] == '}' {
					end = i + 1
					break
				}
				return 0, errors.New("bad object")
			}
		case '[':
			i = skipSpace(i + 1)
			if i < len(data) && data[i] == ']' {
				end = i + 1
				break
			}
			for n := 0; ; n++ {
				var err error
				if i, err = walk(i, append(append([]string{}, at...), strconv.Itoa(n))); err != nil {
					return 0, err
				}
				i = skipSpace(i)
				if i < len(data) && data[i] == ',' {
					i++
					continue
				}
				if i < len(data) && data[i] == ']' {
					end = i + 1
					break
				}
				return 0, errors.New("bad array")
			}
		case '"':
			j, _, err := str(i)
			if err != nil {
				return 0, err
			}
			end = j
		default:
			j := i
			for j < len(data) && !strings.ContainsRune(",}] \n\t\r", rune(data[j])) {
				j++
			}
			end = j
		}
		visit(at, start, end)
		return end, nil
	}
	_, err := walk(0, nil)
	return err
}

// ## EOF

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_extras.go
package wirepod_ttr

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// From the old pages and the JSON tree viewer (2026-10-20, operator: "do all 8
// please * and corrected tpu +/- rotation"; gopod_notes/PAGES_AND_VIEWER_ABSORB_SURVEY_001.md):
//   system-cams [tpu]                          Brobot 1, Brobot 2 and the USB webcam side by side;
//                                              tpu = the selected feed's Edge TPU result (system-visual)
//   system-chats <words> [operator|brobots|managers]   every saved session's chat searched
//   system-chat-save [session]                 a readable chat.md in the session's folder
//   system-menu-theme <dark|light|blue|green|purple>   system_menu.themes into system_menu.look
//   system-numpad (alone)                      each key: persona, body, live or not
// (camera feeds as a surprise egg: chapters/_always, gopod_interview_chapters.go)

func gopodSystemCams(text string) string {
	f := strings.Fields(strings.ToLower(text))
	if len(f) > 1 && f[1] == "tpu" {
		return gopodSystemVisual("system-visual")
	}
	// 2026-10-20 (operator: "Multi (more than 1) cam feeds * brobot 1 photo,
	// brobot 2 photo, TPU bounding boxes on enhanced version"): every feed's
	// photo now, then the selected one's TPU boxes.
	r := gopodVisualAll()
	gopodVisualPost(r)
	return gopodVisualLine(r)
}

// gopodSessionChats: every saved session's chat.jsonl, newest first.
func gopodSessionChats() []string {
	root, _, _ := gopodRecordRoot()
	files, _ := filepath.Glob(filepath.Join(root, "*", "*_session", "chat.jsonl"))
	sort.Sort(sort.Reverse(sort.StringSlice(files)))
	return files
}

type gopodChatLine struct {
	Time string `json:"time"`
	Who  string `json:"who"`
	To   string `json:"to"`
	Text string `json:"text"`
}

// gopodChatShown: a line worth reading (not a voice file, a marker or a sound).
func gopodChatShown(l gopodChatLine) bool {
	t := strings.TrimSpace(l.Text)
	return t != "" && !strings.HasPrefix(t, "/") && !strings.Contains(t, ".wav") && !strings.Contains(t, "|") && l.Who != "Reporter Voice"
}

func gopodSystemChats(text string) string {
	f := strings.Fields(text)
	if len(f) < 2 {
		return "system-chats <words> [operator|brobots|managers]"
	}
	who := ""
	words := f[1:]
	switch strings.ToLower(words[len(words)-1]) {
	case "operator", "brobots", "managers":
		who, words = strings.ToLower(words[len(words)-1]), words[:len(words)-1]
	}
	want := strings.ToLower(strings.Join(words, " "))
	if want == "" {
		return "system-chats <words> [operator|brobots|managers]"
	}
	var out []string
	found := 0
	for _, file := range gopodSessionChats() {
		fh, err := os.Open(file)
		if err != nil {
			continue
		}
		sc := bufio.NewScanner(fh)
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		for sc.Scan() {
			var l gopodChatLine
			if json.Unmarshal(sc.Bytes(), &l) != nil || !gopodChatShown(l) {
				continue
			}
			switch who {
			case "operator":
				if l.Who != "Operator" {
					continue
				}
			case "brobots":
				if l.Who != "Brobot 1" && l.Who != "Brobot 2" {
					continue
				}
			case "managers":
				if !gopodHostIsManager(l.Who) {
					continue
				}
			}
			low := strings.ToLower(l.Text)
			at := strings.Index(low, want)
			if at < 0 {
				continue
			}
			found++
			if len(out) < 30 {
				// The match marked >>like this<<, a little either side.
				from, to := at-60, at+len(want)+60
				if from < 0 {
					from = 0
				}
				if to > len(l.Text) {
					to = len(l.Text)
				}
				snip := l.Text[from:at] + ">>" + l.Text[at:at+len(want)] + "<<" + l.Text[at+len(want):to]
				out = append(out, fmt.Sprintf("%s %s: %s", l.Time, l.Who, strings.ReplaceAll(snip, "\n", " ")))
			}
		}
		fh.Close()
	}
	head := fmt.Sprintf("system-chats \"%s\": %d found in %d sessions", want, found, len(gopodSessionChats()))
	if found > len(out) {
		head += fmt.Sprintf(" (the newest %d shown)", len(out))
	}
	return gopodJoinLines(append([]string{head}, out...))
}

func gopodSystemChatSave(text string) string {
	f := strings.Fields(text)
	chats := gopodSessionChats()
	if len(chats) == 0 {
		return "system-chat-save: no saved sessions."
	}
	file := chats[0]
	if len(f) > 1 {
		file = ""
		for _, c := range chats {
			if strings.Contains(c, f[1]) {
				file = c
				break
			}
		}
		if file == "" {
			return "system-chat-save: no session " + f[1] + " (its folder name, e.g. 154136_session)"
		}
	}
	fh, err := os.Open(file)
	if err != nil {
		return err.Error()
	}
	defer fh.Close()
	out := []string{"# Chat - " + filepath.Base(filepath.Dir(filepath.Dir(file))) + " " + filepath.Base(filepath.Dir(file)), ""}
	sc := bufio.NewScanner(fh)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	n := 0
	for sc.Scan() {
		var l gopodChatLine
		if json.Unmarshal(sc.Bytes(), &l) != nil || !gopodChatShown(l) {
			continue
		}
		to := ""
		if l.To != "" {
			to = " to " + l.To
		}
		out = append(out, fmt.Sprintf("- %s **%s%s:** %s", l.Time, l.Who, to, strings.ReplaceAll(l.Text, "\n", " ")))
		n++
	}
	md := filepath.Join(filepath.Dir(file), "chat.md")
	if err := os.WriteFile(md, []byte(strings.Join(out, "\n")+"\n"), 0o644); err != nil {
		return "system-chat-save: " + err.Error()
	}
	if _, uid, gid := gopodRecordRoot(); uid >= 0 {
		os.Chown(md, uid, gid)
	}
	return fmt.Sprintf("system-chat-save: %d lines -> %s", n, md)
}

func gopodSystemMenuTheme(text string) string {
	f := strings.Fields(strings.ToLower(text))
	var names []string
	for k := range gopodBrobotsSettings.SystemMenu.Themes {
		names = append(names, k)
	}
	sort.Strings(names)
	if len(f) < 2 {
		return "system-menu-theme <" + strings.Join(names, "|") + "> (then F5)"
	}
	theme, ok := gopodBrobotsSettings.SystemMenu.Themes[f[1]]
	if !ok {
		return "system-menu-theme <" + strings.Join(names, "|") + ">"
	}
	keys := make([]string, 0, len(theme))
	for k := range theme {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b, _ := json.Marshal(theme[k])
		if _, err := gopodSettingsWrite(func(file string) (string, error) {
			return gopodJSONFileSet(file, []string{"system_menu", "look", k}, string(b))
		}); err != nil {
			return "system-menu-theme: not changed - " + err.Error()
		}
	}
	return "system-menu-theme: " + f[1] + " - F5 to see it (system-undo puts it back one value at a time)"
}

// gopodSystemTPU: "system-tpu [restart]" - the Edge TPU worker's state, or a fresh start.
func gopodSystemTPU(text string) string {
	f := strings.Fields(strings.ToLower(text))
	if len(f) > 1 && f[1] == "restart" {
		gopodQuickLook.Lock()
		gopodQuickLookRestartLocked()
		gopodQuickLook.Unlock()
		// 2026-10-20 (operator: "do I wait for a response? Or WTF?"): the answer
		// comes when the worker is up (or isn't, after 30 s).
		for i := 0; i < 60; i++ {
			time.Sleep(500 * time.Millisecond)
			gopodQuickLook.Lock()
			ready := gopodQuickLook.ready
			gopodQuickLook.Unlock()
			if ready {
				return "Edge TPU worker: ready."
			}
		}
		return "Edge TPU worker: not ready after 30 s."
	}
	gopodQuickLook.Lock()
	ready := gopodQuickLook.ready
	gopodQuickLook.Unlock()
	return "Edge TPU worker: " + map[bool]string{true: "ready", false: "not ready - system-tpu restart"}[ready]
}

func gopodSystemNumpadList() string {
	m, _, err := gopodNumpadMap()
	if err != nil {
		return "system-numpad: the key map can't be read - " + err.Error()
	}
	var keys []string
	for k := range m.NumlockOn {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) < len(keys[j]) || len(keys[i]) == len(keys[j]) && keys[i] < keys[j] })
	out := []string{"numpad keys (NumLock on) - system-numpad on|off switches them; KP+ / KP- turn the feeds:"}
	for _, k := range keys {
		e := m.NumlockOn[k]
		out = append(out, fmt.Sprintf("  %-5s %s (%s) - %s", k, e.Persona, e.Body, e.Status))
	}
	return gopodJoinLines(out)
}

// ## EOF

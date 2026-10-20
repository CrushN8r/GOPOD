// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_more.go
package wirepod_ttr

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// system-undo, system-status, system-help (2026-10-17, operator: "Add:
// system-undo, system-status, system-help"). Typed on the chat page's
// keyboard, answered in the terminal bubble (gopodSystemSetLine).
//
//   system-undo     puts back the last system-set edit (all its files), if
//                   nothing has changed them since; again for the one before
//   system-status   Wire-Pod, the language model, the warm-up gate, the chat
//                   page, the run going, both Brobots' battery
//   system-help     every system command, a line each

// gopodUndoFile: one file of an edit, before and after.
type gopodUndoFile struct {
	Path   string `json:"path"`
	Before []byte `json:"before"`
	After  []byte `json:"after"`
}

// gopodUndoEdit: one system-set command's files.
type gopodUndoEdit struct {
	What  string          `json:"what"`
	At    string          `json:"at"`
	Files []gopodUndoFile `json:"files"`
}

const gopodUndoMost = 20

var gopodUndoPath = gopodPathWithEnvOverride("GOPOD_SYSTEM_UNDO_PATH", "gopod_system_undo.json")

var gopodUndo = struct {
	sync.Mutex
	open *gopodUndoEdit // the command being run
}{}

// gopodUndoBegin / gopodUndoEnd wrap one system-set command; its writes
// (gopodWriteFileSafe) are kept for system-undo.
func gopodUndoBegin(what string) {
	gopodUndo.Lock()
	gopodUndo.open = &gopodUndoEdit{What: what, At: time.Now().Format("15:04:05")}
	gopodUndo.Unlock()
}

func gopodUndoEnd() {
	gopodUndo.Lock()
	edit := gopodUndo.open
	gopodUndo.open = nil
	gopodUndo.Unlock()
	if edit == nil || len(edit.Files) == 0 {
		return
	}
	edits := gopodUndoLoad()
	edits = append(edits, *edit)
	if len(edits) > gopodUndoMost {
		edits = edits[len(edits)-gopodUndoMost:]
	}
	gopodUndoSave(edits)
}

// gopodUndoNote: a file about to be written (called by gopodWriteFileSafe).
func gopodUndoNote(path string, before, after []byte) {
	gopodUndo.Lock()
	defer gopodUndo.Unlock()
	if gopodUndo.open != nil {
		gopodUndo.open.Files = append(gopodUndo.open.Files, gopodUndoFile{path, before, after})
	}
}

func gopodUndoLoad() []gopodUndoEdit {
	var edits []gopodUndoEdit
	if data, err := os.ReadFile(gopodUndoPath); err == nil {
		json.Unmarshal(data, &edits)
	}
	return edits
}

func gopodUndoSave(edits []gopodUndoEdit) {
	data, _ := json.Marshal(edits)
	tmp := gopodUndoPath + ".tmp"
	if os.WriteFile(tmp, data, 0o600) == nil {
		os.Rename(tmp, gopodUndoPath)
	}
}

// gopodSystemUndo puts back the last edit.
func gopodSystemUndo() string {
	edits := gopodUndoLoad()
	if len(edits) == 0 {
		return "system-undo: nothing to undo."
	}
	last := edits[len(edits)-1]
	for _, f := range last.Files {
		now, err := os.ReadFile(f.Path)
		if err != nil || !bytes.Equal(now, f.After) {
			return "system-undo: not undone - " + filepath.Base(f.Path) + " has changed since \"" + last.What + "\" (" + last.At + "). Edit it by hand: system-locate."
		}
	}
	for i := len(last.Files) - 1; i >= 0; i-- {
		f := last.Files[i]
		if err := gopodWriteRaw(f.Path, f.Before); err != nil {
			return "system-undo: " + filepath.Base(f.Path) + " - " + err.Error()
		}
	}
	gopodUndoSave(edits[:len(edits)-1])
	left := ""
	if n := len(edits) - 1; n > 0 {
		left = " (system-undo again for \"" + edits[n-1].What + "\")"
	}
	return "system-undo: undid \"" + last.What + "\" (" + last.At + ")" + left
}

// gopodWriteRaw writes a file through a temp file, keeping its mode (no undo note).
func gopodWriteRaw(path string, data []byte) error {
	info, err := os.Stat(path)
	if err != nil {
		return errors.New("is gone")
	}
	tmp := path + ".system-set.tmp"
	if err := os.WriteFile(tmp, data, info.Mode().Perm()); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// gopodSystemStatus: how things stand, a line each.
func gopodSystemStatus() string {
	lines := []string{"system-status, " + time.Now().Format("15:04:05") + ":", "  Wire-Pod: up"}
	if base, model, ok := gopodWarmupOllama(); !ok {
		lines = append(lines, "  language model: no endpoint/model in the knowledge settings")
	} else if gopodWarmupLoaded(base, model) {
		lines = append(lines, "  language model: "+model+" loaded (warm)")
	} else {
		lines = append(lines, "  language model: "+model+" not loaded (cold)")
	}
	if gopodWarmingNow() {
		lines = append(lines, "  warm-up: going - lines wait for LLM Ready")
	} else {
		lines = append(lines, "  warm-up: done")
	}
	if gopodPageHere() {
		lines = append(lines, "  chat page: here")
	} else {
		lines = append(lines, "  chat page: away")
	}
	if r, ok := gopodRunNow(); ok {
		lines = append(lines, "  run: "+r.Name+" going")
	} else {
		lines = append(lines, "  run: none")
	}
	gopodBatteryLast.Lock()
	for _, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN} {
		text, at := gopodBatteryLast.text[esn], gopodBatteryLast.at[esn]
		if text == "" {
			text = "no reading yet"
		} else {
			text += " (" + time.Since(at).Round(time.Second).String() + " ago)"
		}
		lines = append(lines, "  "+gopodBrobotName(esn)+": "+text)
	}
	gopodBatteryLast.Unlock()
	return gopodJoinLines(lines)
}

func gopodJoinLines(lines []string) string {
	var b bytes.Buffer
	for i, l := range lines {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(l)
	}
	return b.String()
}

// gopodSystemHelp: every system command.
func gopodSystemHelp() string {
	return gopodJoinLines([]string{
		"system commands (typed on this keyboard):",
		"  system-set <setting> <value>      change a setting (system-set master-volume 35%)",
		"  system-set <path.to.setting> <value>   any setting by its path (system-set managers.qr_note Scan it!)",
		"  system-add / system-remove <path> <item>   an item in a list, or key = value in a group",
		"  system-lingo <brobots|reporters|butler> <word> = <said>   (remove <word>; the lingo alone: its list)",
		"  system-topic [<topic> ...]        a chapter topic: words, points (say lines), label, reminder, pictures",
		"  system-restart [go]               Wire-Pod restarts; the page reloads itself (click Log for sound)",
		"  system-score <song>               a song's score: every step and its words",
		"  system-play <song> [from <step>] [to <step>] [robot 1|2] [dry]   a song or part of it (pha0b)",
		"  system-gap <seconds>|off|default  reporter gap for system-play",
		"  system-tweaks on|off              apply phcal tweaks before a live system-play",
		"  system-calibrate <move> <1|2> [values] [dry]   one move alone (phcal)",
		"  system-set <song> tempo <seconds> the song's tempo (tempo-set); a step's tempo_factor too",
		"  system-brobot <stats|volume|eyes|locale|units|location|timezone|say|wake|head|lift|wheels|camera> [1|2|both]",
		"  system-faces [list|rename|delete|add] [1|2]   the faces a Brobot knows",
		"  system-usb                        the USB devices plugged in (live in the menu)",
		"  system-tpu [restart]              the Edge TPU worker's state, or start it again",
		"  system-photos [1|2]               a Brobot's photos",
		"  system-cams [tpu]                 the Brobots' cameras and the webcam side by side; tpu = the Edge TPU on the selected one",
		"  system-chats <words> [operator|brobots|managers]   search every saved session's chat",
		"  system-chat-save [session]        a readable chat.md in the session's folder",
		"  system-menu-theme <preset>        the menu's look: dark, light, blue, green, purple",
		"  system-visual [next|prev|<id>]    the TPU +/- rotation: a feed's frame and its Edge TPU objects (KP+ / KP-)",
		"  system-version                    Wire-Pod's commit, the newest, what an update touches",
		"  system-native [on|off]            Wire-Pod's own voice answers (was the Native? box)",
		"  system-numpad on|off              the numpad keys as push-to-talk (was the Numpad? box)",
		"  system-set <song> <step> <field> <value>   change one step (volume, text, timing)",
		"  system-set <song> [step]          a song's steps, or one step",
		"  system-set-list                   every setting, song and step field",
		"  system-locate <thing>             where it lives, file:line (a setting, a song step, a word)",
		"  system-locate-list                every setting's and song's file:line",
		"  system-undo                       put back the last system-set",
		"  system-status                     Wire-Pod, language model, page, run, Brobots' battery",
		"  system-find <word>                where a word shows up: the log (6 hours), settings, songs",
		"  system-search <words>             GOPOD's own web search, the results (to the Brobots: \"search <words>\")",
		"  system-check                      every file the settings, chapters and songs name - all there?",
		"  system-voices-rebuild [go]        check, back up every voice, then wpr 2 + F5 makes them all again",
		"  system-test-list                  the tests: system-test-<name>, several in a row, system-test-all (the audit)",
		"  system-help                       this list",
	})
}

// ## EOF

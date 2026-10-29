// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_more.go
package wirepod_ttr

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
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
	if gopodWriteFile(tmp, data, 0o600) == nil {
		os.Rename(tmp, gopodUndoPath)
	}
}

// gopodSystemUndo puts back the last edit.
func gopodSystemUndo() string {
	edits := gopodUndoLoad()
	if len(edits) == 0 {
		return gopodSystemText("system_more.system_undo_nothing_to_undo")
	}
	last := edits[len(edits)-1]
	for _, f := range last.Files {
		now, err := os.ReadFile(f.Path)
		if err != nil || !bytes.Equal(now, f.After) {
			return gopodSystemText("system_more.system_undo_not_undone_has", "path", filepath.Base(f.Path), "what", last.What, "at", last.At)
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
	return gopodSystemText("system_more.system_undo_undid", "what", last.What, "at", last.At, "left", left)
}

// gopodWriteRaw writes a file through a temp file, keeping its mode (no undo note).
func gopodWriteRaw(path string, data []byte) error {
	info, err := os.Stat(path)
	if err != nil {
		return errors.New("is gone")
	}
	tmp := path + ".system-set.tmp"
	if err := gopodWriteFile(tmp, data, info.Mode().Perm()); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// gopodSystemStatus: how things stand, a line each.
func gopodSystemStatus() string {
	lines := []string{"system-status, " + time.Now().Format("15:04:05") + ":", "  Wire-Pod: up"}
	if base, model, ok := gopodWarmupOllama(); !ok {
		lines = append(lines, gopodSystemText("system_status.model_none"))
	} else if gopodWarmupLoaded(base, model) {
		lines = append(lines, gopodSystemText("system_status.model_warm", "model", model))
	} else {
		lines = append(lines, gopodSystemText("system_status.model_cold", "model", model))
	}
	if gopodWarmingNow() {
		lines = append(lines, gopodSystemText("system_status.warmup_going"))
	} else {
		lines = append(lines, gopodSystemText("system_status.warmup_done"))
	}
	if gopodPageHere() {
		lines = append(lines, gopodSystemText("system_status.page_here"))
	} else {
		lines = append(lines, gopodSystemText("system_status.page_away"))
	}
	if n := GopodAdminPagesOpen(); n > 1 {
		lines = append(lines, gopodSystemText("system_status.admin_pages", "n", strconv.Itoa(n)))
	}
	if r, ok := gopodRunNow(); ok {
		lines = append(lines, gopodSystemText("system_status.run_going", "run", r.Name))
	} else {
		lines = append(lines, gopodSystemText("system_status.run_none"))
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

// gopodSystemHelp: every system command - its list is system_help.lines in the
// settings file (2026-10-29, pass 3 of FLEXIBLE_DATA_SURVEY_001.md: the code
// prints it, the words live with the rest).
func gopodSystemHelp() string {
	return gopodJoinLines(gopodBrobotsSettings.SystemHelp.Lines)
}

// ## EOF

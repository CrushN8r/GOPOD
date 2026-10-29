// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_settings_folder.go
package wirepod_ttr

// The settings folder (2026-10-29, operator: "a named folder with correct &
// organized contents"; pass 6 of FLEXIBLE_DATA_SURVEY_001.md). GOPOD's settings
// are one set of blocks kept in a folder, a short file per area
// (webroot/gopod_settings/: 00_about, 01_brain, 02_cast, 03_chat_page, 04_game,
// 05_voices_and_robots, 06_system), each block in exactly one file. Wire-Pod
// loads them as one (gopodSettingsMerged) and serves that whole to the chat page
// at /gopod_brobots_settings.json; an edit goes to the file holding its block
// (gopodSettingsFileFor) - system-set, system-add/remove, the lingo, the voices
// menu, the menu look; system-locate and system-find answer file:line in it.

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// gopodSettingsDir: the settings folder.
var gopodSettingsDir = gopodPathWithEnvOverride("GOPOD_SETTINGS_DIR", "webroot/gopod_settings")

// gopodSettingsFiles: the folder's files, in order.
func gopodSettingsFiles() []string {
	files, _ := filepath.Glob(filepath.Join(gopodSettingsDir, "*.json"))
	sort.Strings(files)
	return files
}

// gopodSettingsMerged: every file's blocks as one JSON object, in file order.
// A block in two files is an error.
func gopodSettingsMerged() ([]byte, error) {
	files := gopodSettingsFiles()
	if len(files) == 0 {
		return nil, errors.New("no settings files in " + gopodSettingsDir)
	}
	seen := map[string]string{}
	var parts [][]byte
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		var top map[string]json.RawMessage
		if err := json.Unmarshal(data, &top); err != nil {
			return nil, errors.New(filepath.Base(f) + ": " + err.Error())
		}
		for k := range top {
			if other, ok := seen[k]; ok {
				return nil, errors.New(k + " is in both " + other + " and " + filepath.Base(f))
			}
			seen[k] = filepath.Base(f)
		}
		inner := bytes.TrimSpace(data)
		inner = bytes.TrimSpace(inner[1 : len(inner)-1])
		if len(inner) > 0 {
			parts = append(parts, inner)
		}
	}
	return append(append([]byte("{\n  "), bytes.Join(parts, []byte(",\n  "))...), []byte("\n}\n")...), nil
}

// gopodSettingsFileFor: the file holding a setting's block (path[0]); a new
// block goes into the first file.
func gopodSettingsFileFor(path []string) string {
	files := gopodSettingsFiles()
	if len(path) > 0 {
		for _, f := range files {
			data, err := os.ReadFile(f)
			if err != nil {
				continue
			}
			var top map[string]json.RawMessage
			if json.Unmarshal(data, &top) == nil {
				if _, ok := top[path[0]]; ok {
					return f
				}
			}
		}
	}
	if len(files) > 0 {
		return files[0]
	}
	return filepath.Join(gopodSettingsDir, "00_about.json")
}

// GopodSettingsHandler: the settings as one, for the chat page
// (/gopod_brobots_settings.json - the address it has always read).
func GopodSettingsHandler(w http.ResponseWriter, r *http.Request) {
	data, err := gopodSettingsMerged()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(data)
}

// gopodSettingsShow: a settings file as system-locate shows it.
func gopodSettingsShow(f string) string {
	return strings.TrimPrefix(f, gopodRepoRoot()+"/")
}

// ## EOF

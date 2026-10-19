// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_song_folders.go
package wirepod_ttr

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GOLDEN - a song is found by its name, never its folder (operator 2026-10-10:
// "even shuffling (song) folders without notice should be easy to remap"). The
// "song_id" in a song's own knobs.json names it; its folder under
// goverlord/runtime/songs/ can be renamed or renumbered with nothing else to
// update. The Python side is runtime/songs/tools/song_folders_001.py, the same rule.

// gopodSongsDir: goverlord/runtime/songs, beside the overlay.
func gopodSongsDir() string {
	return filepath.Join(gopodChipperRoot(), "..", "..", "runtime", "songs")
}

type gopodSongKnobsTop struct {
	SongID string `json:"song_id"`
	Audio  string `json:"audio"`
	// 2026-10-11: a song's sound files ("sound" steps, the fart joke).
	Sounds []string `json:"sounds"`
}

// gopodAllSongs: song_id -> folder, for every song folder with a knobs.json.
func gopodAllSongs() (map[string]string, error) {
	found := map[string]string{}
	knobs, _ := filepath.Glob(filepath.Join(gopodSongsDir(), "*", "knobs.json"))
	sort.Strings(knobs)
	for _, k := range knobs {
		folder := filepath.Dir(k)
		name := filepath.Base(folder)
		// 2026-10-16: a folder starting with "_" is a template (_new_song, with _llm_warm inside it).
		if name == "tools" || strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".") {
			continue
		}
		data, err := os.ReadFile(k)
		if err != nil {
			continue
		}
		var top gopodSongKnobsTop
		if json.Unmarshal(data, &top) != nil || top.SongID == "" {
			continue
		}
		if other, dup := found[top.SongID]; dup {
			return found, errors.New("song_id " + top.SongID + " is in both " + filepath.Base(other) + " and " + name)
		}
		found[top.SongID] = folder
	}
	return found, nil
}

// gopodFindSong: the folder of the song with this song_id.
func gopodFindSong(songID string) (string, error) {
	songs, err := gopodAllSongs()
	if err != nil {
		return "", err
	}
	if dir, ok := songs[songID]; ok {
		return dir, nil
	}
	return "", errors.New("no song folder has song_id " + songID)
}

// gopodSongAudioFiles: each song's own soundtrack (its knobs.json "audio"),
// found anywhere inside its folder (its private media folder) - copied to the
// page at Wire-Pod's start, so a song's audio needs no path in the settings.
func gopodSongAudioFiles() []string {
	songs, _ := gopodAllSongs()
	var files []string
	for _, folder := range songs {
		data, err := os.ReadFile(filepath.Join(folder, "knobs.json"))
		if err != nil {
			continue
		}
		var top gopodSongKnobsTop
		if json.Unmarshal(data, &top) != nil {
			continue
		}
		want := map[string]bool{}
		for _, name := range append([]string{top.Audio}, top.Sounds...) {
			if name != "" {
				want[name] = true
			}
		}
		if len(want) == 0 {
			continue
		}
		filepath.WalkDir(folder, func(p string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && want[d.Name()] {
				files = append(files, p)
				delete(want, d.Name())
			}
			return nil
		})
	}
	return files
}

// ## EOF

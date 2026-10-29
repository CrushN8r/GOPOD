// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_locate.go
package wirepod_ttr

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// system-locate (2026-10-17, operator: "system-set (set a setting) *
// system-locate (file to edit setting)"; "Please build system-locate"): where a
// setting, a song step or a word lives - the file and its line - so it can be
// opened and edited by hand. Read only; typed on the chat page's keyboard,
// answered in system-set's terminal bubble (gopodSystemSetLine).
//
//   system-locate                     how to use it
//   system-locate master-volume       a system-set setting
//   system-locate <song>              the song's files
//   system-locate <song> <step>       the step in knobs.json, zKnobs.json, story.md
//   system-locate <song> <step> <field>
//   system-locate <word>              settings keys holding the word (volume, sting...)

const gopodLocateMost = 15

func gopodSystemLocate(text string) string {
	args := strings.Fields(strings.TrimSpace(text))[1:]
	if len(args) == 0 || strings.EqualFold(args[0], "help") {
		return "system-locate - where a thing lives (file:line):\n" +
			"  system-locate master-volume          a system-set setting\n" +
			"  system-locate <song>                 a song's files\n" +
			"  system-locate <song> <step> [field]  a step (knobs.json, zKnobs.json, story.md)\n" +
			"  system-locate <word>                 settings holding the word (volume, sting, awaits...)"
	}
	if k, ok := gopodSystemSetKnobs[strings.ToLower(args[0])]; ok {
		return args[0] + ": " + gopodLocateJSON(gopodSettingsFileFor(k.path), k.path) + "  (" + strings.Join(k.path, ".") + ")"
	}
	if dir, err := gopodSystemSetSong(args[0]); err == nil {
		return gopodLocateSong(dir, args[1:])
	}
	// A word: every settings file that holds it (the settings folder).
	word := strings.ToLower(strings.Join(args, " "))
	var found []string
	for _, f := range gopodSettingsFiles() {
		if out := gopodLocateWord(f, word); out != gopodSystemText("system_locate.nothing_named_in_the_settings", "word", word) {
			found = append(found, out)
		}
	}
	if len(found) == 0 {
		return gopodSystemText("system_locate.nothing_named_in_the_settings", "word", word)
	}
	return strings.Join(found, "\n")
}

// gopodLocateShow: a file as the repo names it (goverlord/...).
func gopodLocateShow(file string) string {
	repo := filepath.Join(gopodChipperRoot(), "..", "..", "..")
	if rel, err := filepath.Rel(repo, file); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return file
}

// gopodLineAt: the line number (from 1) of a byte offset.
func gopodLineAt(data []byte, at int) int {
	if at > len(data) {
		at = len(data)
	}
	return bytes.Count(data[:at], []byte("\n")) + 1
}

// gopodLocateJSON: "file:line" of the value at path, or where its object is.
func gopodLocateJSON(file string, path []string) string {
	data, err := os.ReadFile(file)
	if err != nil {
		return gopodSystemText("system_locate.can_t_read_it", "value", gopodLocateShow(file))
	}
	for n := len(path); n > 0; n-- {
		if s, _, ok := gopodJSONSpan(data, path[:n]); ok {
			out := gopodLocateShow(file) + ":" + strconv.Itoa(gopodLineAt(data, s))
			if n < len(path) {
				out += " (no " + path[n] + " yet - its place)"
			}
			return out
		}
	}
	return gopodSystemText("system_locate.not_in_it", "value", gopodLocateShow(file))
}

func gopodLocateSong(dir string, args []string) string {
	files := []string{"knobs.json", "zKnobs.json", "story.md"}
	if len(args) == 0 {
		var out []string
		for _, f := range files {
			if p := filepath.Join(dir, f); gopodFileExists(p) {
				out = append(out, gopodLocateShow(p))
			}
		}
		return filepath.Base(dir) + ":\n  " + strings.Join(out, "\n  ")
	}
	id := args[0]
	var out []string
	for _, f := range files[:2] {
		p := filepath.Join(dir, f)
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		steps, _, _ := gopodSystemSetKnobsFileData(data)
		i := gopodSystemSetStepIndex(steps, id)
		if i < 0 {
			continue
		}
		path := []string{"steps", strconv.Itoa(i)}
		if len(args) > 1 {
			path = append(path, args[1])
		}
		out = append(out, gopodLocateJSON(p, path))
	}
	if p := filepath.Join(dir, "story.md"); gopodFileExists(p) {
		data, _ := os.ReadFile(p)
		for n, line := range strings.Split(string(data), "\n") {
			if strings.TrimSpace(line) == "## STEP "+id {
				out = append(out, gopodLocateShow(p)+":"+strconv.Itoa(n+1)+" (its words)")
				break
			}
		}
	}
	if len(out) == 0 {
		return gopodSystemText("system_locate.no_step_in_system_set", "id", id, "dir", filepath.Base(dir))
	}
	return id + ":\n  " + strings.Join(out, "\n  ")
}

// gopodLocateWord: the settings whose key holds the word, with their lines.
func gopodLocateWord(file, word string) string {
	data, err := os.ReadFile(file)
	if err != nil {
		return gopodSystemText("system_locate.can_t_read_it", "value", gopodLocateShow(file))
	}
	type hit struct {
		line int
		text string
	}
	var hits []hit
	gopodJSONWalk(data, func(path []string, start, end int) {
		if len(path) == 0 {
			return
		}
		key := strings.ToLower(path[len(path)-1])
		if !strings.Contains(key, word) || strings.HasSuffix(key, "about") {
			return
		}
		value := string(data[start:end])
		if len(value) > 40 || strings.ContainsAny(value, "{[\n") {
			value = ""
		} else {
			value = " = " + value
		}
		hits = append(hits, hit{gopodLineAt(data, start), strings.Join(path, ".") + value})
	})
	if len(hits) == 0 {
		return gopodSystemText("system_locate.nothing_named_in_the_settings", "word", word)
	}
	sort.Slice(hits, func(a, b int) bool { return hits[a].line < hits[b].line })
	var b strings.Builder
	b.WriteString(gopodLocateShow(file) + ", \"" + word + "\":\n")
	for i, h := range hits {
		if i == gopodLocateMost {
			b.WriteString("  ... " + strconv.Itoa(len(hits)-gopodLocateMost) + " more\n")
			break
		}
		b.WriteString("  " + strconv.Itoa(h.line) + ": " + h.text + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// gopodSystemLocateList (2026-10-17, operator: "system-locate-list ... for fast
// reference"): every system-set setting and every song's files, with their lines.
func gopodSystemLocateList() string {
	var names []string
	for n := range gopodSystemSetKnobs {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	b.WriteString("system-locate - settings:\n")
	for _, n := range names {
		b.WriteString("  " + n + ": " + gopodLocateJSON(gopodSettingsFileFor(gopodSystemSetKnobs[n].path), gopodSystemSetKnobs[n].path) + "\n")
	}
	songs, _ := gopodAllSongs()
	var dirs []string
	for _, dir := range songs {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	b.WriteString("songs:\n")
	for _, dir := range dirs {
		b.WriteString("  " + strings.ReplaceAll(gopodLocateSong(dir, nil), "\n  ", "\n    ") + "\n")
	}
	b.WriteString("one step: system-locate <song> <step> [field] | a word: system-locate <word>")
	return b.String()
}

// ## EOF

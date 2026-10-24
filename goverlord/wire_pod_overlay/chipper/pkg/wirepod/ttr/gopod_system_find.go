// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_find.go
package wirepod_ttr

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// system-find (2026-10-17, operator: "I want Goverlord to know how to pull lists
// like you do"; "system-find FEED_CHECK or system-find Toronto * yes, please"):
// every place a word shows up - Wire-Pod's log of the last few hours (the
// journal, read with fixed arguments; the word only filters the lines, it's
// never part of a command), the settings, and the songs' story.md / knobs.json -
// the newest log lines first, a few of each. Typed on the chat page's keyboard,
// answered in the terminal bubble.
//
//   system-find FEED_CHECK      log lines about a missing / doubled reply
//   system-find Toronto         where Toronto came up, in the log and the files

const (
	gopodFindLogMost  = 15
	gopodFindFileMost = 10
	gopodFindHours    = "-6h"
)

// gopodFindJournal reads the journal (swapped in tests).
var gopodFindJournal = func() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	return gopodAsOperator(exec.CommandContext(ctx, "journalctl", "-u", "wire-pod", "--since", gopodFindHours, "--no-pager", "-o", "short")).Output()
}

func gopodSystemFind(text string) string {
	word := strings.TrimSpace(strings.Join(strings.Fields(text)[1:], " "))
	if word == "" {
		return "system-find <word> - where it shows up: the log (last 6 hours), the settings, the songs.\n" +
			"  e.g. system-find FEED_CHECK | system-find Toronto | system-find not heard"
	}
	lower := strings.ToLower(word)
	var b strings.Builder
	b.WriteString("system-find \"" + word + "\":\n")

	// The log, newest first.
	if data, err := gopodFindJournal(); err != nil {
		b.WriteString("log: can't read it (" + err.Error() + ")\n")
	} else {
		var hits []string
		sc := bufio.NewScanner(bytes.NewReader(data))
		sc.Buffer(make([]byte, 1024*1024), 1024*1024)
		for sc.Scan() {
			if line := sc.Text(); strings.Contains(strings.ToLower(line), lower) {
				hits = append(hits, gopodFindLogLine(line))
			}
		}
		b.WriteString("log, last 6 hours: " + strconv.Itoa(len(hits)) + " line(s)")
		if len(hits) > gopodFindLogMost {
			b.WriteString(", the newest " + strconv.Itoa(gopodFindLogMost))
		}
		b.WriteString("\n")
		for i := len(hits) - 1; i >= 0 && i >= len(hits)-gopodFindLogMost; i-- {
			b.WriteString("  " + hits[i] + "\n")
		}
	}

	// The settings.
	settings := filepath.Join(gopodChipperRoot(), "webroot", "gopod_brobots_settings.json")
	b.WriteString(gopodFindInFile(settings, lower))

	// The songs.
	songs, _ := gopodAllSongs()
	var dirs []string
	for _, dir := range songs {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	for _, dir := range dirs {
		for _, name := range []string{"story.md", "knobs.json"} {
			b.WriteString(gopodFindInFile(filepath.Join(dir, name), lower))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// gopodFindLogLine: a journal line without its host and process ("Oct 17 09:55:43 GOPOD_...").
func gopodFindLogLine(line string) string {
	if i := strings.Index(line, "]: "); i > 0 && len(line) > 15 {
		line = line[7:15] + " " + line[i+3:]
	}
	if len(line) > 220 {
		line = line[:220] + "..."
	}
	return line
}

// gopodFindInFile: "file: N line(s)" and the first few, or "" when none.
func gopodFindInFile(file, lower string) string {
	data, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	var hits []string
	for n, line := range strings.Split(string(data), "\n") {
		if strings.Contains(strings.ToLower(line), lower) {
			t := strings.TrimSpace(line)
			if len(t) > 160 {
				i := strings.Index(strings.ToLower(t), lower)
				from := i - 60
				if from < 0 {
					from = 0
				}
				to := from + 160
				if to > len(t) {
					to = len(t)
				}
				t = "..." + t[from:to] + "..."
			}
			hits = append(hits, strconv.Itoa(n+1)+": "+t)
		}
	}
	if len(hits) == 0 {
		return ""
	}
	out := gopodLocateShow(file) + ": " + strconv.Itoa(len(hits)) + " line(s)\n"
	for i, h := range hits {
		if i == gopodFindFileMost {
			out += "  ... " + strconv.Itoa(len(hits)-gopodFindFileMost) + " more\n"
			break
		}
		out += "  " + h + "\n"
	}
	return out
}

// ## EOF

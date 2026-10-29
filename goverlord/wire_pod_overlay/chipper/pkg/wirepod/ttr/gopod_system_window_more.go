// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_window_more.go
package wirepod_ttr

// Run everything from the system window (2026-10-29, operator: "Aim for a point
// where I can run everything currently GOPOD from the system window * all
// possible edits to calibrations, songs, scripted handling, live response chains
// and the chains within, etc... * most of what I'm asking for is done, just check
// during your sweep"). The sweep's two missing commandments:
//   system-record on|off                      Record Session (the page's box, from the keyboard)
//   system-docbs [<pool> [add|remove <line>]] DocBS's pool files (gopod_probes/docbs): list, show, edit

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// gopodSystemRecord: system-record on|off.
func gopodSystemRecord(text string) string {
	f := strings.Fields(strings.ToLower(text))
	if len(f) < 2 || (f[1] != "on" && f[1] != "off") {
		return gopodSystemText("system_window_more.system_record_on_off_record")
	}
	GopodRecordSessionSet(f[1] == "on")
	return gopodSystemText("system_window_more.record_session", "arg1", f[1])
}

var gopodDocBSVowelRE = regexp.MustCompile(`[aeiouAEIOU]`)

// gopodSystemDocBS: system-docbs - the pools and their lines; add or remove one.
// An _18 pool's line goes uncensored into its private z-copy (when there) and
// starred into the public file.
func gopodSystemDocBS(text string) string {
	dir := filepath.Join(gopodChipperRoot(), gopodBrobotsSettings.DocBS.Folder)
	f := strings.Fields(text)
	if len(f) < 2 {
		names, _ := filepath.Glob(filepath.Join(dir, "*.txt"))
		var out []string
		for _, n := range names {
			base := strings.TrimSuffix(filepath.Base(n), ".txt")
			if strings.HasPrefix(base, "z") {
				continue
			}
			out = append(out, base+" ("+strconv.Itoa(len(gopodDocBSLines(n)))+")")
		}
		sort.Strings(out)
		return gopodSystemText("system_window_more.docbs_pools_system_docbs_pool", "out", strings.Join(out, ", "))
	}
	pool := f[1]
	if strings.ContainsAny(pool, "/\\.") || strings.HasPrefix(pool, "z") {
		return gopodSystemText("system_window_more.system_docbs_no_pool", "pool", pool)
	}
	public := filepath.Join(dir, pool+".txt")
	if _, err := os.Stat(public); err != nil {
		return gopodSystemText("system_window_more.system_docbs_no_pool_system", "pool", pool)
	}
	if len(f) < 4 || (f[2] != "add" && f[2] != "remove") {
		return pool + ": " + strings.Join(gopodDocBSLines(public), " | ")
	}
	line := strings.TrimSpace(strings.Join(f[3:], " "))
	files := []string{public}
	private := filepath.Join(dir, "z"+pool+".txt")
	if _, err := os.Stat(private); err == nil {
		files = append(files, private)
	}
	for _, file := range files {
		want := line
		if file == public && strings.HasSuffix(pool, "_18") {
			want = gopodStarExpletives(line)
		}
		lines := gopodDocBSLines(file)
		var out []string
		found := false
		for _, l := range lines {
			if strings.EqualFold(l, want) || (file == public && strings.EqualFold(l, gopodStarExpletives(line))) {
				found = true
				if f[2] == "remove" {
					continue
				}
			}
			out = append(out, l)
		}
		if f[2] == "add" && !found {
			out = append(out, want)
		}
		if err := gopodWriteFileSafe(file, []byte(strings.Join(out, "\n")+"\n")); err != nil {
			return "system-docbs: not changed - " + err.Error()
		}
	}
	// The next pick reads it fresh.
	gopodDocBSDecks.Lock()
	gopodDocBSDecks.left = map[string][]string{}
	gopodDocBSDecks.Unlock()
	return gopodSystemText("system_window_more.lines_now", "pool", pool, "arg2", f[2], "line", line, "value", strconv.Itoa(len(gopodDocBSLines(public))))
}

func gopodDocBSLines(file string) []string {
	data, _ := os.ReadFile(file)
	var out []string
	for _, l := range strings.Split(string(data), "\n") {
		if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "#") {
			out = append(out, l)
		}
	}
	return out
}

// gopodStarExpletives: a line's expletives starred (f*ck) - the public files' rule.
var gopodExpletiveRE = regexp.MustCompile(`(?i)\b(?:fuck|shit|cunt|bitch|faggot|bastard|motherf|asshole|pussy|piss|jizz|whore|slut|dumbass|goddamn|dick|bullshit)[a-z]*\b`)

func gopodStarExpletives(line string) string {
	return gopodExpletiveRE.ReplaceAllStringFunc(line, func(w string) string { return gopodDocBSVowelRE.ReplaceAllString(w, "*") })
}

// ## EOF

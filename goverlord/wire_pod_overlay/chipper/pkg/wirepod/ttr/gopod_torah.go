// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_torah.go
package wirepod_ttr

// The GOPOD Torah (2026-10-29, operator: "New Take: GOPOD system Torah (the
// ancient scriptures that Ron used to bring the brobots to life), is now the
// matching master-GOPOD-system document to refer to"; "everybody says a master
// bible. I want a master Torah, just because, lol * so for me that means a fun
// take, not religious * 1) a master GOPOD system reference document, that always
// matches growing GOPOD system * 2) A matching a light touch of a Torah look").
// GOPOD_TORAH.md, at the repo's root, is written from the running settings -
// the system window's groups and items (system_menu, with the run cards for a
// "runs" group), every system command (system-help), every test (system-test-*)
// and the audit's order, the run cards, the looks. 2026-10-29 (operator: "all
// system commands are \"commandments\" (as in the 10 commandments, lol)"). system-check says when it's
// behind; system-torah writes it again; TestTorahMatches holds it at every build.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// gopodTorahTests: the system tests (set at start - the test list itself runs
// system-check, which reads the Torah, so it's handed over after).
var gopodTorahTests = func() []gopodSystemTest { return nil }

func init() { gopodTorahTests = func() []gopodSystemTest { return gopodSystemTests } }

// gopodTorahDir: the scriptures' own scrolls, one file each.
const gopodTorahDir = "torah"

// gopodTorahFiles: every scroll GOPOD writes - the main Torah and each
// scripture's own file - by its path from the repo's root.
func gopodTorahFiles() map[string]string {
	files := map[string]string{"GOPOD_TORAH.md": gopodTorahPage()}
	all := gopodTorahScriptures()
	for i, sc := range all {
		files[sc.File] = gopodTorahScripturePage(all, i)
	}
	return files
}

// gopodTorahScripturePage: one scripture's scroll - its title, its plain
// paragraphs (or its guide), its commandments, and the way on.
func gopodTorahScripturePage(all []gopodTorahScripture, i int) string {
	sc := all[i]
	var b strings.Builder
	w := func(lines ...string) {
		for _, l := range lines {
			b.WriteString(l + "\n")
		}
	}
	w("# "+sc.Title, "")
	for _, p := range sc.Tour {
		w(p, "")
	}
	if len(sc.Tour) == 0 && sc.Guide != "" {
		w(sc.Guide, "")
	}
	w("## Commandments", "")
	for _, c := range sc.Commandments {
		w("- " + c)
	}
	w("")
	if sc.Next != "" {
		w("*"+sc.Next+"*", "")
	}
	nav := []string{"[The GOPOD Torah](../GOPOD_TORAH.md)"}
	if i > 0 {
		nav = append([]string{"[" + all[i-1].Title + "](" + filepath.Base(all[i-1].File) + ")"}, nav...)
	}
	if i+1 < len(all) {
		nav = append(nav, "["+all[i+1].Title+"]("+filepath.Base(all[i+1].File)+")")
	}
	w("---", "", strings.Join(nav, " · "))
	return b.String()
}

// gopodTorahFileName: "Scripture 1, Restart and check" -> torah/SCRIPTURE_01_RESTART_AND_CHECK.md.
func gopodTorahFileName(n int, title string) string {
	var words []string
	for _, f := range strings.FieldsFunc(strings.ToUpper(title), func(r rune) bool { return !(r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') }) {
		words = append(words, f)
	}
	num := strconv.Itoa(n)
	if n < 10 {
		num = "0" + num
	}
	return gopodTorahDir + "/SCRIPTURE_" + num + "_" + strings.Join(words, "_") + ".md"
}

// gopodTorahFirstSentence: a paragraph's first sentence, for the main Torah's list.
func gopodTorahFirstSentence(p string) string {
	for i, r := range p {
		if (r == '.' || r == '?' || r == '!') && (i+1 == len(p) || p[i+1] == ' ') {
			return strings.TrimSpace(p[:i+1])
		}
	}
	return strings.TrimSpace(p)
}

// gopodTorahPage: the scrolls, from the settings as they are now.
func gopodTorahPage() string {
	var raw struct {
		SystemMenu struct {
			Intro  string `json:"intro"`
			Groups []struct {
				Title string `json:"title"`
				Guide string `json:"guide"`
				Next  string `json:"next"`
				Runs  bool   `json:"runs"`
				Items []struct {
					Cmd   string `json:"cmd"`
					About string `json:"about"`
				} `json:"items"`
			} `json:"groups"`
		} `json:"system_menu"`
	}
	data, _ := gopodSettingsMerged()
	_ = json.Unmarshal(data, &raw)
	s := gopodBrobotsSettings
	var b strings.Builder
	w := func(lines ...string) {
		for _, l := range lines {
			b.WriteString(l + "\n")
		}
	}
	tw := s.Torah
	w("# The GOPOD Torah", "", "> "+tw.Opening, "",
		"The master reference for GOPOD's system window: every scripture (the window's groups), every commandment (its",
		"commands), every test, every run. A fun take, not a religious one. GOPOD writes these scrolls itself from its",
		"running settings, so they always match the system as it grows - `system-check` says when they've fallen behind,",
		"`system-torah` writes them again; `system-tour` goes through them in the chat.", "")
	if t := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(raw.SystemMenu.Intro), tw.Opening)); t != "" {
		w("*"+t+"*", "")
	}
	// 2026-10-29 (operator: "create a main torah md & pull scriptures to
	// separate md files?"): each scripture is its own scroll (torah/), linked here.
	w("## The Scriptures", "")
	for _, sc := range gopodTorahScriptures() {
		about := ""
		if len(sc.Tour) > 0 {
			about = " - " + gopodTorahFirstSentence(sc.Tour[0])
		}
		w("- [" + sc.Title + "](" + sc.File + ")" + about)
	}
	w("")
	w("## The Commandments", "", "Every system commandment, as `system-help` gives it:", "", "```")
	for _, l := range strings.Split(strings.TrimSpace(gopodSystemHelp()), "\n") {
		w(strings.TrimRight(l, " "))
	}
	w("```", "", "## The Tests", "", "Each checks one thing, nothing said or moved (`system-test-<name>`; `system-test-all` runs the audit).", "",
		"| Test | What it checks |", "|---|---|")
	for _, t := range gopodTorahTests() {
		w("| `" + t.name + "` | " + strings.ReplaceAll(t.about, "|", "/") + " |")
	}
	w("", "The audit's order: "+strings.Join(s.SystemTests.Audit, ", ")+".", "", "## The Runs", "",
		"One card per run (`game.runs`) - start it, stop it, its song.", "",
		"| Run | Start | Stop | Song |", "|---|---|---|---|")
	for _, r := range s.Game.Runs {
		song := r.Song
		if r.Draft {
			song += " (the scripted draft)"
		}
		if r.Pending {
			song += " - pending, not on the menu yet"
		}
		w("| " + r.Title + " | `" + r.Start + "` | `" + r.Stop + "` | " + song + " |")
	}
	var themes []string
	for k := range s.SystemMenu.Themes {
		themes = append(themes, k)
	}
	sort.Strings(themes)
	w("", "## The Looks", "", "The system window's looks (`system-menu-theme <name>`, then F5): "+strings.Join(themes, ", ")+".", "",
		"---", "", "GOPOD - [README](README.md) · [asks](BROBOT_ASKS.md) · [the studio](ALIAS-PLAY-STUDIO.md)")
	return b.String()
}

func itoaTorah(n int) string { return strconv.Itoa(n) }

// gopodTorahScripture: one group of the system window, in the Torah's words.
type gopodTorahScripture struct {
	Title, Guide, Next string
	Tour               []string // the tour's plain paragraphs (system_menu.groups[].tour)
	File               string   // its own scroll, from the repo's root (torah/SCRIPTURE_01_....md)
	Commandments       []string // "{cmd}, {about}"
	Cmds               []string // the commandments' own words
}

// gopodTorahScriptures: the system window's groups as scriptures (torah's words):
// "Scripture 1, Restart and check", each item "system-status, Wire-Pod, ...",
// its next line "Go to scripture 2".
func gopodTorahScriptures() []gopodTorahScripture {
	var raw struct {
		SystemMenu struct {
			Groups []struct {
				Title string   `json:"title"`
				Guide string   `json:"guide"`
				Next  string   `json:"next"`
				Runs  bool     `json:"runs"`
				Tour  []string `json:"tour"`
				Items []struct {
					Cmd   string `json:"cmd"`
					About string `json:"about"`
				} `json:"items"`
			} `json:"groups"`
		} `json:"system_menu"`
	}
	data, _ := gopodSettingsMerged()
	_ = json.Unmarshal(data, &raw)
	tw := gopodBrobotsSettings.Torah
	var out []gopodTorahScripture
	for i, g := range raw.SystemMenu.Groups {
		title := strings.TrimSpace(g.Title)
		if at := strings.Index(title, ". "); at >= 0 && at < 4 {
			title = strings.TrimSpace(title[at+2:])
		}
		sc := gopodTorahScripture{Title: strings.NewReplacer("{n}", strconv.Itoa(i+1), "{title}", title).Replace(tw.Scripture), Guide: g.Guide, Tour: g.Tour, File: gopodTorahFileName(i+1, title)}
		add := func(cmd, about string) {
			sc.Cmds = append(sc.Cmds, cmd)
			sc.Commandments = append(sc.Commandments, strings.NewReplacer("{cmd}", cmd, "{about}", about).Replace(tw.Commandment))
		}
		if g.Runs {
			for _, r := range gopodBrobotsSettings.Game.Runs {
				if !r.Pending {
					add(r.Start, r.MenuAbout)
				}
			}
		}
		for _, it := range g.Items {
			add(it.Cmd, it.About)
		}
		sc.Next = g.Next
		if tw.NextFrom != "" {
			sc.Next = strings.ReplaceAll(sc.Next, tw.NextFrom, tw.NextTo)
		}
		out = append(out, sc)
	}
	return out
}

// gopodTorahBehind: a scroll doesn't match the system now (or an old one is left).
func gopodTorahBehind() bool {
	files := gopodTorahFiles()
	for rel, want := range files {
		data, err := os.ReadFile(filepath.Join(gopodRepoRoot(), rel))
		if err != nil || string(data) != want {
			return true
		}
	}
	return len(gopodTorahStale(files)) > 0
}

// gopodTorahStale: scripture scrolls left in torah/ that the system no longer has.
func gopodTorahStale(files map[string]string) []string {
	var out []string
	old, _ := filepath.Glob(filepath.Join(gopodRepoRoot(), gopodTorahDir, "SCRIPTURE_*.md"))
	for _, f := range old {
		if _, ok := files[gopodTorahDir+"/"+filepath.Base(f)]; !ok {
			out = append(out, f)
		}
	}
	return out
}

// gopodWriteTorah writes every scroll again (and drops the stale ones).
func gopodWriteTorah() error {
	files := gopodTorahFiles()
	if err := gopodMkdirAll(filepath.Join(gopodRepoRoot(), gopodTorahDir), 0o755); err != nil {
		return err
	}
	for rel, data := range files {
		path := filepath.Join(gopodRepoRoot(), rel)
		if old, err := os.ReadFile(path); err == nil && string(old) == data {
			continue
		}
		if err := gopodWriteFileSafe(path, []byte(data)); err != nil {
			return err
		}
	}
	for _, f := range gopodTorahStale(files) {
		os.Remove(f)
	}
	return nil
}

// gopodSystemTorah: system-torah - the scrolls written again.
func gopodSystemTorah() string {
	if err := gopodWriteTorah(); err != nil {
		return "system-torah: " + err.Error()
	}
	return gopodSystemText("torah.the_gopod_torah_is_written")
}

// ## EOF

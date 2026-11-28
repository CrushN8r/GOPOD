// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_torah.go
package wirepod_ttr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// gopodTorahTests: the system tests (set at start - the test list itself runs system-check, which
// reads the Scrolls, so it's handed over after).
var gopodTorahTests = func() []gopodSystemTest { return nil }

func init() { gopodTorahTests = func() []gopodSystemTest { return gopodSystemTests } }

// gopodTorahDir: the scriptures' own scrolls, one file each.
const gopodTorahDir = "system_scrolls"

// gopodTorahFiles: every scroll GOPOD writes - the main scroll and each scripture's own file - by
// its path from the repo's root.
func gopodTorahFiles() map[string]string {
	files := map[string]string{"GOPOD_SYSTEM_SCROLLS.md": gopodTorahPage()}
	all := append([]gopodTorahScripture{gopodTorahIntro()}, gopodTorahScriptures()...)
	for i, sc := range all {
		files[sc.File] = gopodTorahScripturePage(all, i)
	}
	return files
}

// gopodTorahScripturePage: one scripture's scroll - its title, its plain paragraphs (or its guide),
// its commandments, and the way on.
func gopodTorahScripturePage(all []gopodTorahScripture, i int) string {
	sc := all[i]
	var b strings.Builder
	w := func(lines ...string) {
		for _, l := range lines {
			b.WriteString(l + "\n")
		}
	}
	words := []string{strings.ToLower(sc.Title)}
	for _, c := range sc.Commandments {
		if cmd, _, ok := strings.Cut(c, ","); ok {
			words = append(words, strings.TrimSpace(cmd))
		}
	}
	w("# "+sc.Title,
		"- **words:** "+strings.Join(words, ", "),
		"- **ask:** What's in "+sc.Title+"?",
		"- **about:** written by GOPOD from the system window (system-menu.md) - system-scrolls writes it again; system-tour reads it from here",
		"---", "")
	numbered := func(title string, lines []string) {
		if len(lines) == 0 {
			return
		}
		w("## " + title)
		for n, l := range lines {
			w(strconv.Itoa(n+1) + ". " + l)
		}
		w("")
	}
	tour := sc.Tour
	if len(tour) == 0 && sc.Guide != "" {
		tour = []string{sc.Guide}
	}
	numbered("Tour", tour)
	numbered("Commandments", sc.Commandments)
	if sc.Next != "" {
		numbered("Next", []string{sc.Next})
	}
	nav := []string{"[GOPOD SYSTEM SCROLLS](../../GOPOD_SYSTEM_SCROLLS.md)"}
	if i > 0 {
		nav = append([]string{"[" + all[i-1].Title + "](../" + filepath.Base(filepath.Dir(all[i-1].File)) + "/README.md)"}, nav...)
	}
	if i+1 < len(all) {
		nav = append(nav, "["+all[i+1].Title+"](../"+filepath.Base(filepath.Dir(all[i+1].File))+"/README.md)")
	}
	w(strings.Join(nav, " · "))
	return b.String()
}

// gopodTorahIntro: scripture 0, the intro - the opening and the window's intro (torah.intro_title;
// system_menu.intro).
func gopodTorahIntro() gopodTorahScripture {
	tw := gopodBrobotsSettings.Torah
	title := strings.NewReplacer("{n}", "0", "{title}", tw.IntroTitle).Replace(tw.Scripture)
	tour := []string{tw.Opening}
	if t := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(gopodBrobotsSettings.SystemMenu.Intro), tw.Opening)); t != "" {
		tour = append(tour, t)
	}
	return gopodTorahScripture{Title: title, Tour: tour, File: gopodTorahFileName(0, tw.IntroTitle)}
}

// gopodTorahFileName: each scripture is its own folder, its scroll the folder's README.md -
// "Scripture 1, Restart and check" -> system_scrolls/scripture-01-restart-and-check/README.md.
func gopodTorahFileName(n int, title string) string {
	var words []string
	for _, f := range strings.FieldsFunc(strings.ToLower(title), func(r rune) bool { return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9') }) {
		words = append(words, f)
	}
	num := strconv.Itoa(n)
	if n < 10 {
		num = "0" + num
	}
	return gopodTorahDir + "/scripture-" + num + "-" + strings.Join(words, "-") + "/README.md"
}

// gopodTorahFirstSentence: a paragraph's first sentence, for the main scroll's list.
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
	data, _ := gopodSettingsFull()
	_ = json.Unmarshal(data, &raw)
	s := gopodBrobotsSettings
	var b strings.Builder
	w := func(lines ...string) {
		for _, l := range lines {
			b.WriteString(l + "\n")
		}
	}
	tw := s.Torah
	w("# GOPOD SYSTEM SCROLLS", "", "> "+tw.Opening, "",
		"The master reference for GOPOD's system window: every scripture (the window's groups), every commandment (its",
		"commands), every test, every run. A fun take, not a religious one. GOPOD writes these scrolls itself from its",
		"running settings, so they always match the system as it grows - `system-check` says when they've fallen behind,",
		"`system-scrolls` writes them again; `system-tour` goes through them in the chat.", "")
	if t := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(raw.SystemMenu.Intro), tw.Opening)); t != "" {
		w("*"+t+"*", "")
	}
	w("## The Scriptures", "")
	for _, sc := range append([]gopodTorahScripture{gopodTorahIntro()}, gopodTorahScriptures()...) {
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

// gopodTorahScripture: one group of the system window, in the Scrolls' words.
type gopodTorahScripture struct {
	Title, Guide, Next string
	Tour               []string // the tour's plain paragraphs (system_menu.groups[].tour)
	File               string   // its own scroll, from the repo's root (system_scrolls/scripture-01-.../README.md)
	Commandments       []string // "{cmd}, {about}"
}

// gopodMenuGroup: one of the system window's groups, as the settings hold it.
type gopodMenuGroup struct {
	Title string   `json:"title"`
	Guide string   `json:"guide"`
	Next  string   `json:"next"`
	Runs  bool     `json:"runs"`
	Tour  []string `json:"tour"`
	Items []struct {
		Cmd   string `json:"cmd"`
		About string `json:"about"`
	} `json:"items"`
}

// gopodSystemMenuGroups: the system window's groups (system_menu.groups, from system-menu.md).
func gopodSystemMenuGroups() []gopodMenuGroup {
	var raw struct {
		SystemMenu struct {
			Groups []gopodMenuGroup `json:"groups"`
		} `json:"system_menu"`
	}
	data, _ := gopodSettingsFull()
	_ = json.Unmarshal(data, &raw)
	return raw.SystemMenu.Groups
}

// gopodTorahScriptures: the system window's groups as scriptures (torah's words): "Scripture 1,
// Restart and check", each item "system-status, Wire-Pod, ...", its next line "Go to scripture 2".
func gopodTorahScriptures() []gopodTorahScripture {
	tw := gopodBrobotsSettings.Torah
	var out []gopodTorahScripture
	for i, g := range gopodSystemMenuGroups() {
		title := strings.TrimSpace(g.Title)
		if at := strings.Index(title, ". "); at >= 0 && at < 4 {
			title = strings.TrimSpace(title[at+2:])
		}
		sc := gopodTorahScripture{Title: strings.NewReplacer("{n}", strconv.Itoa(i+1), "{title}", title).Replace(tw.Scripture), Guide: g.Guide, Tour: g.Tour, File: gopodTorahFileName(i+1, title)}
		add := func(cmd, about string) {
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

// gopodTorahStale: scripture scrolls left in system_scrolls/ that the system no longer has.
func gopodTorahStale(files map[string]string) []string {
	var out []string
	flat, _ := filepath.Glob(filepath.Join(gopodRepoRoot(), gopodTorahDir, "SCRIPTURE_*.md"))
	out = append(out, flat...) // the one-file scrolls
	dirs, _ := filepath.Glob(filepath.Join(gopodRepoRoot(), gopodTorahDir, "scripture-*"))
	for _, d := range dirs {
		if _, ok := files[gopodTorahDir+"/"+filepath.Base(d)+"/README.md"]; !ok {
			out = append(out, d)
		}
	}
	return out
}

// gopodWriteTorah writes every scroll again (and drops the stale ones).
func gopodWriteTorah() error {
	files := gopodTorahFiles()
	for rel, data := range files {
		path := filepath.Join(gopodRepoRoot(), rel)
		if err := gopodMkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if old, err := os.ReadFile(path); err == nil && string(old) == data {
			continue
		}
		if err := gopodWriteFileSafe(path, []byte(data)); err != nil {
			return err
		}
	}
	for _, f := range gopodTorahStale(files) {
		os.RemoveAll(f)
	}
	return nil
}

// gopodSystemTorah: system-scrolls - the scrolls written again.
func gopodSystemTorah() string {
	if err := gopodWriteTorah(); err != nil {
		return "system-scrolls: " + err.Error()
	}
	if err := gopodWriteReadmeCast(); err != nil {
		return "system-scrolls: " + err.Error()
	}
	return gopodSystemText("torah.the_gopod_torah_is_written")
}

// ## EOF

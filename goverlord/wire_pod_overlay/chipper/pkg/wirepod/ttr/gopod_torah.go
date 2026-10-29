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

func gopodTorahPath() string { return filepath.Join(gopodRepoRoot(), "GOPOD_TORAH.md") }

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
	data, _ := os.ReadFile(gopodBrobotsSettingsPath)
	_ = json.Unmarshal(data, &raw)
	s := gopodBrobotsSettings
	var b strings.Builder
	w := func(lines ...string) {
		for _, l := range lines {
			b.WriteString(l + "\n")
		}
	}
	w("# The GOPOD Torah", "",
		"> The ancient scrolls Ron used to bring the Brobots to life.", "",
		"The master reference for GOPOD's system window: every book (the window's groups), every commandment (its",
		"commands), every test, every run. A fun take, not a religious one. GOPOD writes these scrolls itself from its running",
		"settings, so they always match the system as it grows - `system-check` says when they've fallen behind,",
		"`system-torah` writes them again.", "")
	if t := strings.TrimSpace(raw.SystemMenu.Intro); t != "" {
		w("*" + t + "*", "")
	}
	w("## The Books of the System Window", "")
	for i, g := range raw.SystemMenu.Groups {
		title := strings.TrimSpace(g.Title)
		if at := strings.Index(title, ". "); at >= 0 && at < 4 {
			title = strings.TrimSpace(title[at+2:])
		}
		w("### Book " + itoaTorah(i+1) + " - " + title, "")
		if g.Guide != "" {
			w(g.Guide, "")
		}
		w("| Commandment | What it does |", "|---|---|")
		if g.Runs {
			for _, r := range s.Game.Runs {
				w("| `" + r.Start + "` | " + r.MenuAbout + " |")
			}
		}
		for _, it := range g.Items {
			w("| `" + it.Cmd + "` | " + strings.ReplaceAll(it.About, "|", "/") + " |")
		}
		w("")
		if g.Next != "" {
			w("*" + g.Next + "*", "")
		}
	}
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

// gopodTorahBehind: the file doesn't match the system now.
func gopodTorahBehind() bool {
	data, err := os.ReadFile(gopodTorahPath())
	return err != nil || string(data) != gopodTorahPage()
}

// gopodSystemTorah: system-torah - the scrolls written again.
func gopodSystemTorah() string {
	if err := gopodWriteFileSafe(gopodTorahPath(), []byte(gopodTorahPage())); err != nil {
		return "system-torah: " + err.Error()
	}
	return "The GOPOD Torah is written again: GOPOD_TORAH.md matches the system as it is now."
}

// ## EOF

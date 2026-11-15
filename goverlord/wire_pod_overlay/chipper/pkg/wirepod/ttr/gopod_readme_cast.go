// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_readme_cast.go
package wirepod_ttr

// The README's cast table (2026-11-15, operator: "lists/stuff like that should be
// written to files that are either auto-updated on the spot/per-session/daily?
// Like if/when persona(s) join/leave/replace cast members already listed"):
// GOPOD writes the README's "Meet the cast" table itself, between its two
// markers, from the one source - the numpad map's live keys, in its order - with
// each persona's names from its card and what it does from the crew roster. A
// key made live or off in numpad-mapper.md, or a renamed persona, changes the
// table; nothing is written by hand. system-scrolls writes it (with the Scrolls);
// system-check and TestReadmeCastMatches fail when it's behind. Its few words:
// readme_cast (numpad-mapper.md).

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	gopodReadmeCastStart = "<!-- gopod cast table: written by GOPOD from the numpad map (numpad-mapper.md) - system-scrolls writes it again -->"
	gopodReadmeCastEnd   = "<!-- /gopod cast table -->"
)

// gopodReadmeCastTable: the table, a row a live key, then the * key.
func gopodReadmeCastTable() string {
	rc := gopodBrobotsSettings.ReadmeCast
	roster := gopodCrewRoster()
	what := func(names ...string) string {
		for _, n := range names {
			if d := strings.TrimSpace(roster[strings.ToLower(strings.TrimSpace(n))]); d != "" {
				return d
			}
		}
		return ""
	}
	named := func(call, full string) string {
		who := "**" + call + "**"
		if f := strings.TrimSpace(full); f != "" && !strings.EqualFold(f, call) {
			who += " - " + f
		}
		return who
	}
	fill := func(tpl string, kv ...string) string {
		return strings.NewReplacer(kv...).Replace(tpl)
	}
	lines := []string{rc.Head, rc.Rule}
	row := func(key, who, does string) {
		lines = append(lines, fill(rc.Row, "{key}", key, "{who}", who, "{what}", does))
	}
	for _, key := range gopodNumpadLiveKeys() {
		label, target, err := gopodNumpadBrobot(key)
		if err != nil {
			continue
		}
		nick, _ := gopodNumpadKeyLabelIn(gopodNumpadLayout(), key)
		shown := key
		if base, off := strings.CutSuffix(key, gopodNumlockOffSuffix); off {
			shown = fill(rc.NumlockOff, "{key}", base)
		}
		switch {
		case strings.HasPrefix(target, gopodDialZeroMark):
			row(shown, "**"+nick+"** - "+label, what(label))
		case strings.HasPrefix(target, gopodPersonaReporterMark):
			host := strings.TrimPrefix(target, gopodPersonaReporterMark)
			row(shown, fill(rc.OnLocation, "{who}", named(nick, gopodBrobotsSettings.HostPersonas[host].Name)), what(gopodCallName(host), host))
		case strings.HasPrefix(target, gopodPersonaMark):
			host := strings.TrimPrefix(target, gopodPersonaMark)
			row(shown, named(nick, gopodBrobotsSettings.HostPersonas[host].Name), what(gopodCallName(host), host))
		default:
			body := gopodBrobotName(target)
			row(shown, fill(rc.OnBody, "{who}", named(nick, gopodBrobotsSettings.HostPersonas[body].Name), "{body}", body), what(gopodCallName(body), body))
		}
	}
	row("*", rc.AllWho, rc.AllWhat)
	return strings.Join(lines, "\n")
}

func gopodReadmePath() string { return filepath.Join(gopodRepoRoot(), "README.md") }

// gopodReadmeCastWant: the README with its table as it should be; false when the
// README has no markers.
func gopodReadmeCastWant() (string, string, bool) {
	data, err := os.ReadFile(gopodReadmePath())
	if err != nil {
		return "", "", false
	}
	text := string(data)
	a := strings.Index(text, gopodReadmeCastStart)
	b := strings.Index(text, gopodReadmeCastEnd)
	if a < 0 || b < a {
		return text, text, false
	}
	want := text[:a+len(gopodReadmeCastStart)] + "\n" + gopodReadmeCastTable() + "\n" + text[b:]
	return text, want, true
}

// gopodReadmeCastBehind: the README's table doesn't match the map.
func gopodReadmeCastBehind() bool {
	have, want, ok := gopodReadmeCastWant()
	return ok && have != want
}

// gopodWriteReadmeCast writes the README's table again.
func gopodWriteReadmeCast() error {
	have, want, ok := gopodReadmeCastWant()
	if !ok || have == want {
		return nil
	}
	return os.WriteFile(gopodReadmePath(), []byte(want), 0o644)
}

// ## EOF

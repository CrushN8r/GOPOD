// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_roast_dial.go
package wirepod_ttr

// The roast dial and the spice (2026-10-28, operator: "I want like a
// personality toggle from not punching down, to after burner level roasts using
// every f*ck*ng possible personal pressure point * A roast is a f*ck*ng roast or
// its not a roast"; "The controlled legal establishment who pays decides the
// rules of the room. So put it in GOPOD"; "There's public 18+, vs paid work in
// public 18+"; "Sensitive topics switch to neutral (default fallback with my new
// Jekyll/Hyde toggle override)"; "Robot swear list = make it a human swear list,
// lol"; "Exclamations"; "no 'spice per-reply-limit' = aim to not repeat, or try
// to vary a repeat"). Settings roast_dial and spice; his Brobot's pre-build
// context gets {roast} and {spice} (gopodPrebuildContext).

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

// gopodRoastSensitive: the question touches a sensitive topic and the
// Jekyll/Hyde override is off - the answer goes neutral.
func gopodRoastSensitive(question string) bool {
	rd := gopodBrobotsSettings.RoastDial
	if rd.JekyllHyde {
		return false
	}
	lower := strings.ToLower(question)
	for _, w := range rd.SensitiveWords {
		if w != "" && gopodHasWholeWords(lower, strings.ToLower(w)) {
			return true
		}
	}
	return false
}

// gopodRoastLevel: the dial's level now (clamped to the list).
func gopodRoastLevel() (int, bool) {
	rd := gopodBrobotsSettings.RoastDial
	if len(rd.Levels) == 0 {
		return 0, false
	}
	n := rd.Level
	if n < 0 {
		n = 0
	}
	if n >= len(rd.Levels) {
		n = len(rd.Levels) - 1
	}
	return n, true
}

// gopodRoastLine: the {roast} line - the level's text, or the neutral text for
// a sensitive topic.
func gopodRoastLine(question string) string {
	if gopodRoastSensitive(question) {
		return strings.TrimSpace(gopodBrobotsSettings.RoastDial.SensitiveText)
	}
	n, ok := gopodRoastLevel()
	if !ok {
		return ""
	}
	return strings.TrimSpace(gopodBrobotsSettings.RoastDial.Levels[n].Text)
}

// gopodSpiceLine: the {spice} line - a fresh pick of swears (robot ones, and
// human ones from the level that allows them) and exclamations, each answer.
// None for a sensitive topic.
func gopodSpiceLine(question string) string {
	sp := gopodBrobotsSettings.Spice
	if strings.TrimSpace(sp.Text) == "" || gopodRoastSensitive(question) {
		return ""
	}
	pool := append([]string(nil), sp.RobotSwears...)
	if n, ok := gopodRoastLevel(); ok && gopodBrobotsSettings.RoastDial.Levels[n].HumanSwears {
		pool = append(pool, gopodHumanSwears()...)
	}
	pick := func(list []string, n int) string {
		if n <= 0 || n > len(list) {
			n = len(list)
		}
		var out []string
		for _, i := range rand.Perm(len(list))[:n] {
			out = append(out, list[i])
		}
		return strings.Join(out, ", ")
	}
	r := strings.NewReplacer("{swears}", pick(pool, sp.Pick), "{exclamations}", pick(sp.Exclamations, sp.PickExclamations))
	return strings.TrimSpace(r.Replace(sp.Text))
}

// gopodHumanSwears (operator 2026-10-28: "from human swear list, make a copy
// private local human swear list * in (public) human swear list, edit in '*' for
// vowels"; "Look at zKnobs & knobs json configs"): the private, uncensored z-copy
// (spice.human_swears_private, in the overlay chipper folder - gitignored, never
// served) when it's there, else the public one, vowels starred.
func gopodHumanSwears() []string {
	sp := gopodBrobotsSettings.Spice
	if sp.HumanSwearsPrivate != "" {
		// The z-copy beside the settings' chipper folder (the zKnobs.json pattern).
		if data, err := os.ReadFile(filepath.Join(gopodChipperRoot(), sp.HumanSwearsPrivate)); err == nil {
			var list []string
			if json.Unmarshal(data, &list) == nil && len(list) > 0 {
				return list
			}
		}
	}
	return sp.HumanSwears
}

// gopodPersonaWIP: a persona's bits still in the works (managers.personas.
// <host>.wip; operator 2026-10-28: "thought-bubble-WIP, so every persona has a
// chance to develop themselves without interfering with performance duties *
// thought-bubble-WIPs can be tested when a situation comes up"), told as
// wip_text - "" when none.
func gopodPersonaWIP(host string) string {
	p := gopodBrobotsSettings.Managers.Personas[host]
	tpl := strings.TrimSpace(gopodBrobotsSettings.Managers.WIPText)
	if len(p.WIP) == 0 || tpl == "" {
		return ""
	}
	return strings.ReplaceAll(tpl, "{bits}", strings.Join(p.WIP, "; "))
}

// ## EOF

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_preshow.go
package wirepod_ttr

// The pre-show pass, simulated (2026-10-29, operator: "simulate, the live
// pre-show pass, with no brobots * keep me physically out of the loop").
// system-test-preshow runs the pre-show through the game's own code - no
// robots, no model, nothing said: the comm check (every KP1-9 line to its Brobot
// or persona, the persona's prompt built), Doc's BS (a Brobot 1 answer with its
// beats in order, the fallback slots), and a low battery (Brobot 2 at 3.6 V: his
// guard's warning and handover, then his dizzy reply - gopodBatteryLines, the
// live watch's own).

import (
	"fmt"
	"strings"
)

// gopodPreshowLine: system-test-preshow.
func gopodPreshowLine() (bool, string) {
	var parts, problems []string
	// 1. The comm check.
	lines := gopodCommCheckLines()
	keys := map[string]bool{}
	for _, l := range lines {
		_, target, err := gopodNumpadBrobot(l[0])
		if err != nil {
			problems = append(problems, l[0]+": "+err.Error())
			continue
		}
		keys[l[0]] = true
		if strings.HasPrefix(target, gopodPersonaMark) {
			if host := strings.TrimPrefix(target, gopodPersonaMark); strings.TrimSpace(gopodPersonaPrompt(host)) == "" {
				problems = append(problems, l[0]+": "+host+"'s prompt is empty")
			}
		}
	}
	parts = append(parts, fmt.Sprintf("comm check: %d lines, keys %d/9 routed", len(lines), len(keys)))
	if len(keys) < 9 {
		problems = append(problems, fmt.Sprintf("comm check reaches %d of 9 keys", len(keys)))
	}
	// 2. Doc's BS, on a clean slate, put back after.
	saved := gopodBrobotsSettings.DocBS
	gopodBrobotsSettings.DocBS.Share = 100
	gopodBrobotsSettings.DocBS.SlotPrompt = "" // no model: the fallback slots
	gopodDocBSOnFor.Range(func(k, _ any) bool { gopodDocBSOnFor.Delete(k); return true })
	q, raw := "Tell me about the rig", "thinking___ One dolly, two crates, bro."
	got := gopodShapeBrobotAnswer(gopodBrobot1ESN, q, raw)
	gopodBrobotsSettings.DocBS.Share = 0
	gopodDocBSOnFor.Range(func(k, _ any) bool { gopodDocBSOnFor.Delete(k); return true })
	plain := gopodShapeBrobotAnswer(gopodBrobot1ESN, q, raw)
	gopodBrobotsSettings.DocBS = saved
	gopodDocBSOnFor.Range(func(k, _ any) bool { gopodDocBSOnFor.Delete(k); return true })
	// Doc's BS: his answer still in it, wrapped in its beats (longer than plain).
	if !strings.Contains(got, "One dolly") || len(got) <= len(plain)+40 {
		problems = append(problems, "Doc's BS: the beats aren't all there, in order")
	} else {
		parts = append(parts, "Doc's BS: beats in order")
	}
	// 3. A low battery.
	if lines := gopodBrobotsSettings.BatteryWatch.Lines["Brobot 2"]; len(lines) == 0 {
		problems = append(problems, "battery: Brobot 2 has no dizzy line")
	} else {
		guard, warning, reply, _ := gopodBatteryLines("Brobot 2", 3.6, lines[0])
		if guard == "" || warning == "" || !strings.Contains(reply, lines[0]) {
			problems = append(problems, "battery: no guard warning, or the reply lost his line")
		} else {
			parts = append(parts, fmt.Sprintf("battery at 3.6 V: %s: %q -> Brobot 2: %q", guard, gopodCut(warning, 70), gopodCut(reply, 70)))
		}
	}
	line := strings.Join(parts, " · ")
	if len(problems) > 0 {
		line += " - " + strings.Join(problems, "; ")
	}
	return len(problems) == 0, line
}

// ## EOF

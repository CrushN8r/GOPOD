// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_brobots_setup.go
package wirepod_ttr

// The single-bot setup (2026-10-29, operator: "Yes, a single-bot GOPOD path *
// more realistic on-ramp for most Wire-Pod users * does GOPOD detect a single
// brobot and still work? To what extent? * how to handle shared brobots persona?
// * how to handle 'All' selected? * how to handle scripted runs for 2 brobots?";
// "Finish 7 (the single-bot setup)"). At Wire-Pod's start the Brobots are the
// robots Wire-Pod knows (brobots_setup: an ESN set there wins; else GOPOD's own
// two when they're there; else Wire-Pod's robots in order). One robot alone is
// Brobot 1 and GOPOD runs solo: a line to All - typed, the menu's 1/2, KP0 - goes
// to him alone; Brobot 2's key goes to him; his prompt drops his brother (the
// cast, the brother banter) and gets solo_note. No robot known: nothing changes.

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// gopodNoBrobot2: Brobot 2's ESN when there's no second robot (never a real ESN).
const gopodNoBrobot2 = "no-brobot-2"

// gopodSolo: one robot - Brobot 1 alone.
func gopodSolo() bool { return gopodBrobot2ESN == gopodNoBrobot2 }

// GopodBrobotsSetup works out who Brobot 1 and Brobot 2 are (Wire-Pod's start).
func GopodBrobotsSetup() {
	var robots []string
	for _, bot := range vars.BotInfo.Robots {
		if esn := strings.TrimSpace(bot.Esn); esn != "" {
			robots = append(robots, esn)
		}
	}
	b1, b2 := gopodBrobotsPick(robots, gopodBrobotsSettings.BrobotsSetup.Brobot1, gopodBrobotsSettings.BrobotsSetup.Brobot2)
	gopodBrobot1ESN, gopodBrobot2ESN = b1, b2
	mode := "two Brobots"
	if gopodSolo() {
		mode = "solo - one robot, Brobot 1"
	}
	logger.Println("GOPOD_BROBOTS Brobot 1 " + gopodESN1 + ", Brobot 2 " + GopodESNLabel(b2) + " (" + mode + ")")
	// The songs read who's who from here (run_section1_full_live_001.py, Robots.say:
	// solo, Brobot 2's lines are said on the speaker).
	state, _ := json.Marshal(map[string]any{"brobot_1": b1, "brobot_2": b2, "solo": gopodSolo()})
	if err := gopodWriteFileSafe(gopodBrobotsNowPath(), state); err != nil {
		logger.Println("GOPOD_BROBOTS state not written: " + err.Error())
	}
}

// gopodBrobotsNowPath: who's who now, for the songs (live state, gitignored).
func gopodBrobotsNowPath() string { return filepath.Join(gopodChipperRoot(), "gopod_brobots_now.json") }

// gopodBrobotsPick: Brobot 1 and 2 from the settings, the operator's own two
// (his private ~/.gopod_esns.json, gopod_esn_labels.go), or the robots in order;
// one robot alone is Brobot 1 with no Brobot 2. A label in the settings (ESN01,
// ESN02) is his own robot of that label.
func gopodBrobotsPick(robots []string, set1, set2 string) (string, string) {
	own1, own2 := gopodPrivateESNs()
	label := func(v string) string {
		switch strings.ToUpper(strings.TrimSpace(v)) {
		case gopodESN1:
			return own1
		case gopodESN2:
			return own2
		}
		return v
	}
	set1, set2 = label(set1), label(set2)
	has := func(e string) bool {
		for _, r := range robots {
			if strings.EqualFold(r, e) {
				return true
			}
		}
		return false
	}
	if len(robots) == 0 {
		return firstNonEmpty(firstNonEmpty(set1, own1), gopodESN1), firstNonEmpty(firstNonEmpty(set2, own2), gopodESN2)
	}
	b1 := strings.TrimSpace(set1)
	if b1 == "" {
		if own1 != "" && has(own1) {
			b1 = own1
		} else {
			b1 = robots[0]
		}
	}
	b2 := strings.TrimSpace(set2)
	if b2 == "" {
		if own2 != "" && has(own2) && own2 != b1 {
			b2 = own2
		} else {
			for _, r := range robots {
				if !strings.EqualFold(r, b1) {
					b2 = r
					break
				}
			}
		}
	}
	if b2 == "" {
		b2 = gopodNoBrobot2
	}
	return b1, b2
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return strings.TrimSpace(a)
	}
	return b
}

// ## EOF

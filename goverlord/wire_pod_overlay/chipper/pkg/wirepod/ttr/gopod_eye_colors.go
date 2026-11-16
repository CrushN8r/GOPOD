// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_eye_colors.go
package wirepod_ttr

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The Brobots' eye colours - one colour value, the border on a card (2026-11-15,
// operator: "ensure the robots' eye colours match their respective bubble's
// border colours from a single colour value setting * when DOC/PIP respond,
// their respective eye colours change as well, to match DOC/PIP bubbles * upon
// DOC/PIP response completion, respective eye colours default back to brobot
// bubble colours"). A Brobot's own card (brobot-1.md) holds his colours; the
// persona he hosts (its "hosts" card, DOC or PIP) holds its own. The page's
// bubbles and the robot's eyes read the same border. (Was, 2026-10-17: the
// robot's own app colour read into gopod_brobot_eye_colors.json for the page -
// two sources; gone.)

var gopodCardColourRE = regexp.MustCompile(`(?i)\b(background|border|text)\s+(#[0-9a-f]{6})\b`)

// gopodCardColour: one colour ("border") from a card's colours line, "" for none.
func gopodCardColour(colours, which string) string {
	for _, m := range gopodCardColourRE.FindAllStringSubmatch(colours, -1) {
		if strings.EqualFold(m[1], which) {
			return m[2]
		}
	}
	return ""
}

// gopodBrobotCardColours: a Brobot body's colours (his own card) and the hosted
// persona's (the card his "hosts" names), whatever the dial says now.
func gopodBrobotCardColours(body string) (own, hosted string) {
	return gopodBrobotCardColoursIn(gopodBrobotsSettings.PersonaProfiles.Folder, body)
}

func gopodBrobotCardColoursIn(folder, body string) (own, hosted string) {
	path := gopodPersonaProfileFileIn(folder, body)
	head, _ := gopodPersonaProfileReadOne(path)
	if head == nil {
		return "", ""
	}
	own = strings.TrimSpace(head["colours"])
	if host := strings.TrimSpace(head["hosts"]); host != "" {
		hp := filepath.Join(filepath.Dir(path), host)
		if f := gopodTopicTreeFind(host); f != "" {
			hp = f
		}
		if ph, _ := gopodPersonaProfileReadOne(hp); ph != nil {
			hosted = strings.TrimSpace(ph["colours"])
		}
	}
	return own, hosted
}

// gopodBrobotEyeHex: the eye colour for body now - the hosted persona's border
// while he answers as it (the dial past 0%), else his own.
func gopodBrobotEyeHex(body string, answering bool) string {
	own, hosted := gopodBrobotCardColours(body)
	if answering && !gopodAtDialZero(body) {
		if c := gopodCardColour(hosted, "border"); c != "" {
			return c
		}
	}
	return gopodCardColour(own, "border")
}

// gopodSetBrobotEyes sets the robot's eyes to his colour now (answering: the
// persona he answers as); nothing when the card has none.
func gopodSetBrobotEyes(robot *vector.Vector, esn string, answering bool) {
	hex := gopodBrobotEyeHex(gopodBrobotName(esn), answering)
	h, s, ok := gopodHexHS(hex)
	if robot == nil || !ok {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := robot.Conn.SetEyeColor(ctx, &vectorpb.SetEyeColorRequest{Hue: float32(h), Saturation: float32(s)}); err != nil {
		logger.Println("GOPOD_EYES " + gopodBrobotName(esn) + " " + hex + ": " + err.Error())
		return
	}
	logger.Println("GOPOD_EYES " + gopodBrobotName(esn) + " " + hex)
}

// gopodHexHS: "#RRGGBB" -> hue and saturation (0-1), the eyes' two numbers.
func gopodHexHS(hex string) (h, s float64, ok bool) {
	hex = strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(hex) != 6 {
		return 0, 0, false
	}
	n, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return 0, 0, false
	}
	r, g, b := float64(n>>16&255)/255, float64(n>>8&255)/255, float64(n&255)/255
	mx, mn := math.Max(r, math.Max(g, b)), math.Min(r, math.Min(g, b))
	d := mx - mn
	if mx > 0 {
		s = d / mx
	}
	switch {
	case d == 0:
		h = 0
	case mx == r:
		h = math.Mod((g-b)/d, 6) / 6
	case mx == g:
		h = ((b-r)/d + 2) / 6
	default:
		h = ((r-g)/d + 4) / 6
	}
	if h < 0 {
		h++
	}
	return h, s, true
}

// gopodHSVHex: hue, saturation, value (0-1) -> "#RRGGBB" (Python's colorsys).
func gopodHSVHex(h, s, v float64) string {
	h = math.Mod(h, 1)
	if h < 0 {
		h++
	}
	i := math.Floor(h * 6)
	f := h*6 - i
	p, q, t := v*(1-s), v*(1-s*f), v*(1-s*(1-f))
	var r, g, b float64
	switch int(i) % 6 {
	case 0:
		r, g, b = v, t, p
	case 1:
		r, g, b = q, v, p
	case 2:
		r, g, b = p, v, t
	case 3:
		r, g, b = p, q, v
	case 4:
		r, g, b = t, p, v
	default:
		r, g, b = v, p, q
	}
	return fmt.Sprintf("#%02X%02X%02X", int(math.Round(r*255)), int(math.Round(g*255)), int(math.Round(b*255)))
}

// ## EOF

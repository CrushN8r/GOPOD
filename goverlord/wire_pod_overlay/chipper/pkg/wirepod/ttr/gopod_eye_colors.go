// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_eye_colors.go
package wirepod_ttr

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// The Brobots' eye colours - one colour value, the border on a card (2026-11-15,
// operator: "ensure the robots' eye colours match their respective bubble's
// border colours from a single colour value setting * when DOC/PIP respond,
// their respective eye colours change as well, to match DOC/PIP bubbles * upon
// DOC/PIP response completion, respective eye colours default back to brobot
// bubble colours"). Each persona's colours are in one place, persona_colours
// (03_chat_page.json): Brobot N's own, and DOC's or PIP's while they answer from
// his body. The page's bubbles and the robot's eyes read the same border. (Was, 2026-10-17: the
// robot's own app colour read into gopod_brobot_eye_colors.json for the page -
// two sources; gone.)

// gopodBrobotEyeHex: the eye colour for body now - answering, the border of the
// persona answering from him (gopodAnswerer: DOC, PIP, or the Brobot himself).
func gopodBrobotEyeHex(body string, answering bool) string {
	who := body
	if answering {
		who = gopodAnswerer(body)
	}
	return gopodBrobotsSettings.PersonaColours[who].Border
}

// gopodSetBrobotEyes sets the robot's eyes: answering, the persona he answers
// as (its card's border); done, his default - his eye colour in Wire-Pod's
// settings (2026-11-15, operator: "set default eye colour in wire-pod settings").
func gopodSetBrobotEyes(robot *vector.Vector, esn string, answering bool) {
	var h, s float64
	var ok bool
	hex := "default"
	if answering {
		if gopodAtDialZero(gopodBrobotName(esn)) {
			return
		}
		hex = gopodBrobotEyeHex(gopodBrobotName(esn), true)
		h, s, ok = gopodHexHS(hex)
	} else {
		h, s, ok = gopodWirePodEyes(esn)
	}
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

// gopodWirePodEyes: the robot's eye colour in Wire-Pod's settings
// (vic.RobotSettings custom_eye_color), false when none is set.
func gopodWirePodEyes(esn string) (h, s float64, ok bool) {
	doc, found := vars.GetJdoc("vic:"+esn, "vic.RobotSettings")
	if !found {
		return 0, 0, false
	}
	var settings struct {
		Custom struct {
			Enabled    bool    `json:"enabled"`
			Hue        float64 `json:"hue"`
			Saturation float64 `json:"saturation"`
		} `json:"custom_eye_color"`
	}
	if json.Unmarshal([]byte(doc.JsonDoc), &settings) != nil || !settings.Custom.Enabled {
		return 0, 0, false
	}
	return settings.Custom.Hue, settings.Custom.Saturation, true
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

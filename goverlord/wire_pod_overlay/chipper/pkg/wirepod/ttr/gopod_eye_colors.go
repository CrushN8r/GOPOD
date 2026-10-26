// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_eye_colors.go
package wirepod_ttr

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// The Brobots' eye colours from Wire-Pod itself (2026-10-17, operator: "eye
// colours from native wire-pod"; survey E16: only the interview runner wrote
// them, at its start). Each Brobot's own settings as Wire-Pod keeps them
// (vic.RobotSettings, custom_eye_color hue/saturation, when it's on) become
// webroot/gopod_brobot_eye_colors.json, the bubbles' colours - at Wire-Pod's
// start and every 30 s, never while a run is going (2026-09-01, operator: "not
// mid runs though"). A Brobot with no custom colour keeps the page's own.

var gopodEyeColorsPath = gopodPathWithEnvOverride("GOPOD_EYE_COLORS_PATH", "webroot/gopod_brobot_eye_colors.json")

func gopodEyeColorsStart() {
	go func() {
		last := ""
		for {
			if _, running := gopodRunNow(); !running {
				if colors := gopodEyeColors(); len(colors) > 0 {
					data, _ := json.Marshal(colors)
					if string(data) != last {
						tmp := gopodEyeColorsPath + ".tmp"
						if gopodWriteFile(tmp, data, 0o644) == nil && os.Rename(tmp, gopodEyeColorsPath) == nil {
							last = string(data)
							logger.Println("GOPOD_EYE_COLORS " + last)
						}
					}
				}
			}
			time.Sleep(30 * time.Second)
		}
	}()
}

// gopodEyeColors: {"brobot_1": "#RRGGBB", ...} for each Brobot with a custom colour on.
func gopodEyeColors() map[string]string {
	colors := map[string]string{}
	for label, esn := range map[string]string{"brobot_1": gopodBrobot1ESN, "brobot_2": gopodBrobot2ESN} {
		doc, ok := vars.GetJdoc("vic:"+esn, "vic.RobotSettings")
		if !ok {
			continue
		}
		var settings struct {
			Custom struct {
				Enabled    bool    `json:"enabled"`
				Hue        float64 `json:"hue"`
				Saturation float64 `json:"saturation"`
			} `json:"custom_eye_color"`
		}
		if json.Unmarshal([]byte(doc.JsonDoc), &settings) != nil || !settings.Custom.Enabled {
			continue
		}
		colors[label] = gopodHSVHex(settings.Custom.Hue, settings.Custom.Saturation, 1)
	}
	return colors
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

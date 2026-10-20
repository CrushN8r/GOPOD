// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_visual.go
package wirepod_ttr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
)

// GOPOD VISUAL CONTROL - the TPU +/- rotation (2026-10-20, operator: "Focus on
// adding the tpu +/- rotation"; golden guts from the retired 8011 demo,
// wdtm_demo.html: VISUAL_SOURCE_RING ["doc_robot", "pip_robot", "usb_c920"],
// KP+ next / KP- previous, wrapping; the page shows the selected feed, the
// Edge TPU's result on it, live or cached). The ring is visual_sources in the
// settings. A turn grabs one frame from the selected source - a Brobot's camera
// (woken first, gopodRobotCall) or the USB webcam (ffmpeg, v4l2) - runs the
// Edge TPU quick look on it (its objects and the copy with their boxes), and
// keeps it as webroot/visual/<id>.jpg. A source that can't give a frame now
// shows its last frame, marked cached.
//   KP+ / KP- on the chat page (numpad on)   next / previous source
//   system-visual [next|prev|<id>]           the same, typed

type gopodVisualResult struct {
	Index   int      `json:"index"`
	Count   int      `json:"count"`
	ID      string   `json:"id"`
	Label   string   `json:"label"`
	Live    bool     `json:"live"`
	Frame   string   `json:"frame"`
	// 2026-10-20 (operator: "single cam has 3 image display: brobot photo,
	// enhancement, TPU bounding boxes * the "I see..." part"; "Multi cam feeds *
	// brobot 1 photo, brobot 2 photo, TPU bounding boxes on enhanced version"):
	// the row of pictures, and the vision model's "I see ...".
	Frames []string `json:"frames"`
	// Only one feed there now: KP+ / KP- have nothing to turn (the page does nothing).
	Single bool `json:"single"`
	Objects []string `json:"objects"`
	Status  string   `json:"status"`
	At      string   `json:"at"`
}

var gopodVisual = struct {
	sync.Mutex
	index int
}{}

func init() {
	http.HandleFunc("/api-sdk/gopod_visual", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.FormValue("dir") == "all" {
			json.NewEncoder(w).Encode(gopodVisualAll())
			return
		}
		json.NewEncoder(w).Encode(gopodVisualTurn(r.FormValue("dir")))
	})
}

func gopodVisualDir() string { return filepath.Join(gopodChipperRoot(), "webroot", "visual") }

// gopodVisualTurn: dir "next", "prev", "" (the current one) or a source id.
func gopodVisualTurn(dir string) gopodVisualResult {
	ring := gopodBrobotsSettings.VisualSources.Ring
	if len(ring) == 0 {
		return gopodVisualResult{Status: "no visual_sources in the settings"}
	}
	// 2026-10-20 (operator: "no multiple feeds to rotate = not applicable"):
	// KP+ / KP- turn only through the feeds that are there now (a webcam not
	// plugged in is skipped); with one, nothing turns.
	there := func(i int) bool {
		d := ring[i].Device
		if d == "" {
			return true
		}
		_, e := os.Stat(d)
		return e == nil
	}
	count := 0
	for i := range ring {
		if there(i) {
			count++
		}
	}
	gopodVisual.Lock()
	switch dir {
	case "next", "+", "prev", "-":
		if count < 2 {
			gopodVisual.Unlock()
			return gopodVisualResult{Single: true, Count: count}
		}
		step := 1
		if dir == "prev" || dir == "-" {
			step = len(ring) - 1
		}
		for n := 0; n < len(ring); n++ {
			gopodVisual.index = (gopodVisual.index + step) % len(ring)
			if there(gopodVisual.index) {
				break
			}
		}
	case "":
	default:
		for i, s := range ring {
			if strings.EqualFold(s.ID, dir) {
				gopodVisual.index = i
			}
		}
	}
	i := gopodVisual.index % len(ring)
	gopodVisual.Unlock()
	src := ring[i]
	pos := 0
	for j := 0; j <= i; j++ {
		if there(j) {
			pos++
		}
	}
	res := gopodVisualResult{Index: pos, Count: count, ID: src.ID, Label: src.Label}
	os.MkdirAll(gopodVisualDir(), 0o755)
	file := filepath.Join(gopodVisualDir(), src.ID+".jpg")
	var err error
	switch {
	case src.Robot != "":
		esn := gopodBrobotESN(src.Robot)
		err = fmt.Errorf("not reached")
		gopodRobotCall(esn, true, func(robot *vector.Vector, ctx context.Context) string {
			resp, e := robot.Conn.CaptureSingleImage(ctx, &vectorpb.CaptureSingleImageRequest{EnableHighResolution: true})
			if e == nil && resp != nil && len(resp.Data) > 0 {
				e = os.WriteFile(file, resp.Data, 0o644)
			} else if e == nil {
				e = fmt.Errorf("empty frame")
			}
			err = e
			return ""
		})
	case src.Device != "":
		if _, e := os.Stat(src.Device); e != nil {
			err = fmt.Errorf("%s not plugged in", src.Device)
		} else {
			out, e := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-f", "v4l2", "-i", src.Device, "-frames:v", "1", file).CombinedOutput()
			if e != nil {
				err = fmt.Errorf("%v %s", e, strings.TrimSpace(string(out)))
			}
		}
	default:
		err = fmt.Errorf("no robot or device for %s", src.ID)
	}
	res.Live = err == nil
	info, statErr := os.Stat(file)
	if statErr != nil {
		res.Status = "no frame - " + err.Error()
		return res
	}
	res.At = info.ModTime().Format("15:04:05")
	if err != nil {
		res.Status = "cached frame (" + err.Error() + ")"
	}
	frame := file
	if res.Live {
		objects, terr := gopodQuickLookObjects(file)
		if terr != nil {
			res.Status = "Edge TPU: " + terr.Error()
		} else {
			res.Objects = objects
			if res.Status == "" {
				res.Status = "Edge TPU: ok"
			}
		}
	}
	stamp := fmt.Sprintf("?t=%d", time.Now().UnixNano())
	res.Frames = []string{"/visual/" + filepath.Base(file) + stamp}
	if enhanced, boxes := gopodQuickLookPictures(file); gopodFileExists(boxes) {
		if b, e := os.Stat(boxes); e == nil && !b.ModTime().Before(info.ModTime()) {
			frame = boxes
			if gopodFileExists(enhanced) {
				res.Frames = append(res.Frames, "/visual/"+filepath.Base(enhanced)+stamp)
			}
			res.Frames = append(res.Frames, "/visual/"+filepath.Base(boxes)+stamp)
		}
	}
	res.Frame = "/visual/" + filepath.Base(frame) + stamp
	return res
}

// gopodVisualAll (multi-cam): every source's photo now, in a row, then the
// selected source's Edge TPU boxes on its enhanced copy - one more picture for
// each feed found (a webcam plugged in makes it 4).
func gopodVisualAll() gopodVisualResult {
	ring := gopodBrobotsSettings.VisualSources.Ring
	gopodVisual.Lock()
	sel := gopodVisual.index
	gopodVisual.Unlock()
	// 2026-10-20 (operator: "if only a single brobot is detected/called and
	// sends photo, only the photo, enhancement, tpu + boxes on enhanced ... if
	// more than a single brobot ... brobots 1 photo, brobots 2 photo, tpu + boxes
	// on 'selected' enhanced"): only feeds that sent a photo now count.
	var live []gopodVisualResult
	main := gopodVisualResult{}
	for i, src := range ring {
		if src.Device != "" {
			if _, e := os.Stat(src.Device); e != nil {
				continue // not plugged in
			}
		}
		r := gopodVisualTurn(src.ID)
		if !r.Live || len(r.Frames) == 0 {
			continue
		}
		live = append(live, r)
		if i == sel || main.ID == "" {
			main = r
		}
	}
	gopodVisual.Lock()
	gopodVisual.index = sel
	gopodVisual.Unlock()
	if len(live) < 2 {
		return main // one feed: its photo, enhancement, TPU boxes
	}
	var raws []string
	for _, r := range live {
		raws = append(raws, r.Frames[0])
	}
	if len(main.Frames) > 1 {
		raws = append(raws, main.Frames[len(main.Frames)-1])
	}
	main.Frames = raws
	main.Label = fmt.Sprintf("%d feeds (TPU on %s)", len(live), main.Label)
	return main
}

// gopodSystemVisual: the answer to "system-visual [next|prev|<id>]" - its line,
// and the frame in the chat (the menu's results while it's open).
func gopodSystemVisual(text string) string {
	f := strings.Fields(text)
	dir := ""
	if len(f) > 1 {
		dir = strings.ToLower(f[1])
	}
	r := gopodVisualTurn(dir)
	gopodVisualPost(r)
	return gopodVisualLine(r)
}

// gopodVisualPost: the row of pictures in the chat (the menu's results while it's open).
func gopodVisualPost(r gopodVisualResult) {
	var paths []string
	for _, f := range r.Frames {
		paths = append(paths, strings.SplitN(f, "?", 2)[0])
	}
	if len(paths) > 0 {
		GopodChatFeedAdd(gopodBrobotsSettings.TopicPictures.Role, strings.Join(paths, "|"))
	}
}

func gopodVisualLine(r gopodVisualResult) string {
	live := map[bool]string{true: "LIVE", false: "CACHED"}[r.Live]
	objects := "-"
	if len(r.Objects) > 0 {
		objects = strings.Join(r.Objects, ", ")
	}
	line := fmt.Sprintf("FEED %d/%d %s %s - TPU: %s", r.Index, r.Count, r.Label, live, objects)
	if r.Status != "" && r.Status != "Edge TPU: ok" {
		line += " - " + r.Status
	}
	return line
}

// ## EOF

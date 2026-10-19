// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_photo_quick_look.go
package wirepod_ttr

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The photo egg's quick look (2026-10-16, operator: "Edge TPU list shown first
// while moondream is still looking (both, not a swap) * so do this in 2 rounds *
// round 1 is from TPU quick look (while moondream looks) * round 2 is from
// moondream"). One Edge TPU worker (gopod_probes/tools/
// gopod_edge_tpu_quick_look_001.py, settings photo.quick_look) loads the object
// model once at Wire-Pod start; each photo's objects come back in well under a
// second ("person", "chair", ...). A worker that can't start, or doesn't answer
// in time, just gives no objects (photo.quick_none_text).

var gopodQuickLook = struct {
	sync.Mutex
	in    *bufio.Writer
	out   *bufio.Scanner
	cmd   *exec.Cmd
	ready bool
}{}

func init() {
	// Not in the tests (no Edge TPU there).
	if strings.HasSuffix(os.Args[0], ".test") {
		return
	}
	go gopodQuickLookStart()
}

// gopodQuickLookPath: a settings path - "~/" is the GOPOD owner's home, a
// relative one is from the overlay chipper folder.
func gopodQuickLookPath(path string) string {
	path = gopodHomePath(path)
	if path != "" && !filepath.IsAbs(path) {
		path = filepath.Join(gopodChipperRoot(), path)
	}
	return path
}

// gopodQuickLookStart starts the worker and waits for it to say it's ready.
func gopodQuickLookStart() {
	q := gopodBrobotsSettings.Photo.QuickLook
	if !q.Enabled || q.Python == "" || q.Worker == "" {
		return
	}
	cmd := gopodAsOperator(exec.Command(gopodQuickLookPath(q.Python), gopodQuickLookPath(q.Worker),
		gopodQuickLookPath(q.Model), gopodQuickLookPath(q.Labels), strings.TrimSpace(jsonNumber(q.MinScore))))
	stdin, err1 := cmd.StdinPipe()
	stdout, err2 := cmd.StdoutPipe()
	if err1 != nil || err2 != nil {
		logger.Println("GOPOD_QUICK_LOOK unable to start the worker")
		return
	}
	if err := cmd.Start(); err != nil {
		logger.Println("GOPOD_QUICK_LOOK unable to start the worker: " + err.Error())
		return
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), `"ready"`) {
			gopodQuickLook.Lock()
			gopodQuickLook.in, gopodQuickLook.out, gopodQuickLook.cmd, gopodQuickLook.ready = bufio.NewWriter(stdin), scanner, cmd, true
			gopodQuickLook.Unlock()
			logger.Println("GOPOD_QUICK_LOOK ready")
			return
		}
	}
	logger.Println("GOPOD_QUICK_LOOK the worker stopped before it was ready")
	cmd.Wait()
}

func jsonNumber(f float64) string {
	b, _ := json.Marshal(f)
	return string(b)
}

// gopodQuickLookObjects: the objects in the photo at this path, most sure first.
// A worker that doesn't answer in time is stopped and started again, for the
// next photo.
var gopodQuickLookObjects = func(photo string) ([]string, error) {
	q := gopodBrobotsSettings.Photo.QuickLook
	gopodQuickLook.Lock()
	defer gopodQuickLook.Unlock()
	if !gopodQuickLook.ready {
		return nil, errors.New("the Edge TPU worker isn't ready")
	}
	type result struct {
		Objects []string `json:"objects"`
		Seconds float64  `json:"seconds"`
		Error   string   `json:"error"`
	}
	got := make(chan result, 1)
	in, out := gopodQuickLook.in, gopodQuickLook.out
	line := photo
	if q.Pictures {
		enhanced, boxes := gopodQuickLookPictures(photo)
		b, _ := json.Marshal(map[string]any{"photo": photo, "enhanced": enhanced, "boxes": boxes, "enhance": q.Enhance})
		line = string(b)
	}
	go func() {
		var r result
		if _, err := in.WriteString(line + "\n"); err != nil || in.Flush() != nil {
			r.Error = "unable to send the photo"
		} else if !out.Scan() {
			r.Error = "the worker stopped"
		} else if err := json.Unmarshal(out.Bytes(), &r); err != nil {
			r.Error = err.Error()
		}
		got <- r
	}()
	wait := gopodSecondsToDuration(q.TimeoutSeconds)
	if wait <= 0 {
		wait = 3 * time.Second
	}
	select {
	case r := <-got:
		if r.Error != "" {
			if r.Error == "the worker stopped" || r.Error == "unable to send the photo" {
				gopodQuickLookRestartLocked()
			}
			return nil, errors.New(r.Error)
		}
		logger.Println("GOPOD_QUICK_LOOK " + filepath.Base(photo) + ": " + strings.Join(r.Objects, ", ") + " (" + jsonNumber(r.Seconds) + " s)")
		return r.Objects, nil
	case <-time.After(wait):
		gopodQuickLookRestartLocked()
		return nil, errors.New("the Edge TPU worker didn't answer in time")
	}
}

// gopodQuickLookPictures (2026-10-17, operator: "3 images displayed * [raw
// snapshot][snapshot enhancement attempt][TPU object recognition on
// enhancement]"; "TPU with bounding boxes"): where the worker keeps the
// enhanced snapshot and the copy with the objects' boxes, beside the photo.
func gopodQuickLookPictures(photo string) (string, string) {
	base := strings.TrimSuffix(photo, filepath.Ext(photo))
	return base + "_enhanced.jpg", base + "_tpu.jpg"
}

// gopodQuickLookRestartLocked stops the worker and starts a new one.
func gopodQuickLookRestartLocked() {
	if gopodQuickLook.cmd != nil && gopodQuickLook.cmd.Process != nil {
		gopodQuickLook.cmd.Process.Kill()
		go gopodQuickLook.cmd.Wait()
	}
	gopodQuickLook.ready, gopodQuickLook.cmd = false, nil
	go gopodQuickLookStart()
}

// gopodQuickLookList: ["person", "chair", "keyboard"] -> "a person, a chair and
// a keyboard".
func gopodQuickLookList(objects []string) string {
	var named []string
	for _, o := range objects {
		if o = strings.TrimSpace(o); o == "" {
			continue
		}
		article := "a "
		if strings.ContainsRune("aeiou", rune(o[0])) {
			article = "an "
		}
		named = append(named, article+o)
	}
	switch len(named) {
	case 0:
		return ""
	case 1:
		return named[0]
	}
	return strings.Join(named[:len(named)-1], ", ") + " and " + named[len(named)-1]
}

// ## EOF

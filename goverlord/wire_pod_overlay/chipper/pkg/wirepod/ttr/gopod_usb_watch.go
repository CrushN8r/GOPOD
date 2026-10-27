// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_usb_watch.go
package wirepod_ttr

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The USB watch (2026-10-21, operator: "no 'mic deteced and wired in', no
// message bubble * same applies to the original numpad-keyboard"): after the
// warm-up, lsusb every numpad_ptt.usb_watch_seconds. When the mic or the
// numpad goes missing or comes back, Goverlord says its line, with its bubble.

var gopodUSBWatchOnce sync.Once

// gopodUSBAnyOf: one of ids is plugged in (lsusb).
func gopodUSBAnyOf(ids []string) (bool, error) {
	list, err := gopodUSBDevicesErr()
	if err != nil {
		return false, err
	}
	for _, d := range list {
		for _, id := range ids {
			if strings.EqualFold(d.ID, strings.TrimSpace(id)) {
				return true, nil
			}
		}
	}
	return false, nil
}

// gopodUSBSaying (2026-10-22, operator: "after llm ready, mic may not be ready *
// so mic not ready message * if voice not generated yet, 'wait until the message
// bubble completes' * same with mic detected"): Goverlord's mic or numpad
// message is being made or said - a numpad key waits for it (wait_message).
var gopodUSBSaying atomic.Bool

// gopodSystemShowSay: the USB line, said and shown as its host's bubble -
// numpad_ptt.usb_host (Git Repo, the ethical hacker intern, 2026-10-26), else
// Goverlord; returns once it's been heard (its voice made first, if need be).
func gopodSystemShowSay(text string) {
	role := gopodBrobotsSettings.NumpadPTT.USBHost
	if role == "" {
		role = gopodBrobotsSettings.SignOff.SystemRole
	}
	if role != "" && strings.TrimSpace(text) != "" {
		gopodUSBSaying.Store(true)
		defer gopodUSBSaying.Store(false)
		gopodReporterSay(role, strings.TrimSpace(text), true, true)
	}
}

type gopodUSBWatched struct {
	name                   string
	ids                    []string
	missingText, foundText string
	onChange               func(found bool)
}

func gopodUSBWatchStart() {
	np := gopodBrobotsSettings.NumpadPTT
	if np.USBWatchSeconds <= 0 {
		return
	}
	gopodUSBWatchOnce.Do(func() {
		watched := []gopodUSBWatched{
			{"mic", np.MicUSBIDs, np.MicMissingText, np.MicFoundText, func(found bool) { gopodMicMissing.Store(!found) }},
			{"numpad", np.NumpadUSBIDs, np.NumpadMissingText, np.NumpadFoundText, nil},
		}
		go gopodUSBWatch(watched, gopodSecondsToDuration(np.USBWatchSeconds))
	})
}

func gopodUSBWatch(watched []gopodUSBWatched, every time.Duration) {
	last := map[string]bool{}
	first := true
	for {
		for _, w := range watched {
			if len(w.ids) == 0 {
				continue
			}
			found, err := gopodUSBAnyOf(w.ids)
			if err != nil {
				continue
			}
			was, seen := last[w.name]
			last[w.name] = found
			if w.onChange != nil {
				w.onChange(found)
			}
			// The mic's start state was already said by gopodMicCheck.
			if first && (w.name == "mic" || found) {
				continue
			}
			if seen && was == found {
				continue
			}
			if found {
				logger.Println("GOPOD_USB " + w.name + " plugged in")
				gopodSystemShowSay(w.foundText)
			} else {
				logger.Println("GOPOD_USB " + w.name + " missing")
				gopodSystemShowSay(w.missingText)
			}
		}
		first = false
		time.Sleep(every)
	}
}

// ## EOF

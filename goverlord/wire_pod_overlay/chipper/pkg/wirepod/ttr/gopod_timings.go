// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_timings.go
package wirepod_ttr

import "time"

// gopodTiming: a show timing from the settings file's timings block, in seconds - how long a
// warm-up, a recording, an answer or a call may take, how often a watch looks again.
func gopodTiming(name string) time.Duration {
	return gopodSecondsToDuration(gopodBrobotsSettings.Timings[name])
}

// ## EOF

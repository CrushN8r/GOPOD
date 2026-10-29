// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_timings.go
package wirepod_ttr

import "time"

// gopodTiming (2026-10-29, pass 4 of FLEXIBLE_DATA_SURVEY_001.md; operator:
// "decouple dumb code from flexible data"): a show timing from the settings
// file's timings block, in seconds - how long a warm-up, a recording, an answer
// or a call may take, how often a watch looks again. The short polling ticks and
// the hardware-tuned movement values stay in the code. TestTimingsInSettings
// fails if a timing the code asks for is missing or isn't more than 0.
func gopodTiming(name string) time.Duration {
	return gopodSecondsToDuration(gopodBrobotsSettings.Timings[name])
}

// ## EOF

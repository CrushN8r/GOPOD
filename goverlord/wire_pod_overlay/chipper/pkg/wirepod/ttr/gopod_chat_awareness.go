// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_chat_awareness.go
package wirepod_ttr

import (
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The brothers' awareness of each other (2026-10-02/03, operator: "chat history
// awareness (look for opposing 'zzz' to answer if bro is asleep)"). Right before
// the question, a Brobot whose brother was told to sleep gets brothers.asleep_note
// (gopod_brobots_settings.json). What the Brobots know by topic is in the
// scenario packs, loaded by the question's own words
// (loadScenarioPacketsForMessage). The brother is never named (see
// gopod_brobot_personas.go).

// gopodBrotherAsleep: the other brother's newest line in Wire-Pod's log is his
// sleep intent (nothing from him since).
func gopodBrotherAsleep(esn string) bool {
	brother := gopodBrobot1ESN
	if strings.TrimSpace(esn) == gopodBrobot1ESN {
		brother = gopodBrobot2ESN
	}
	lines := append([]string(nil), logger.LogTrayArray...)
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], "Bot "+brother+" ") {
			return strings.Contains(lines[i], "Intent Sent: intent_system_sleep")
		}
	}
	return false
}

// gopodAwarenessNote: the note for this question, or "" when there is nothing
// to add.
func gopodAwarenessNote(esn string) string {
	if gopodBrotherAsleep(esn) {
		logger.Println("BROBOT_AWARENESS brother_asleep=yes")
		return gopodBrobotsSettings.Brothers.AsleepNote
	}
	return ""
}

// ## EOF

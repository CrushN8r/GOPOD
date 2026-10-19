package wirepod_ttr

import (
	"strings"
	"testing"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The brother is asleep while his newest log line is the sleep intent; topic
// packs load by whole-word triggers, more than one per question.
func TestGopodChatAwareness(t *testing.T) {
	saved := logger.LogTrayArray
	defer func() { logger.LogTrayArray = saved }()
	logger.LogTrayArray = []string{"x: Bot " + gopodBrobot2ESN + " Intent Sent: intent_system_sleep\n", "x: Bot " + gopodBrobot1ESN + " Stream type: OPUS\n"}
	if !gopodBrotherAsleep(gopodBrobot1ESN) || gopodBrotherAsleep(gopodBrobot2ESN) {
		t.Fatal("brother 2 should be asleep, brother 1 awake")
	}
	if note := gopodAwarenessNote(gopodBrobot1ESN); !strings.Contains(note, "asleep") {
		t.Fatalf("note = %q", note)
	}
	names := func(msg string) string {
		var n []string
		for _, p := range loadScenarioPacketsForMessage(msg) {
			n = append(n, p.Name)
		}
		return strings.Join(n, ",")
	}
	for msg, want := range map[string]string{
		"introduce yourself and is your brother asleep": "who_we_are,brothers",
		"i want to party where can we go":                "party",
		"tell me all about yourself":                     "who_we_are,background",
		"what is gopod":                                  "gopod",
		"the whole thing was barely funny":               "",
	} {
		if got := names(msg); got != want {
			t.Fatalf("%q loaded %q, want %q", msg, got, want)
		}
	}
}

// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_reminder_rounds_test.go
package wirepod_ttr

import (
	"strings"
	"testing"
)

// TestRemindersStartAgain (operator 2026-10-27: "once all reminders have been
// suggested, brobots go through the reminder's list again less what's already
// talked about"): asked over and over, the reminders keep naming topics - a new
// round once every topic has been named - never a topic already covered.
func TestRemindersStartAgain(t *testing.T) {
	gopodGameReset("test")
	defer gopodGameReset("test")
	labels := gopodBrobotsSettings.InterviewChapters.TopicLabels
	// "the rig" covered: never named again.
	gopodQuestionFactsFor(gopodBrobot1ESN, "Tell me about the rig")
	rig := labels["mobile_gear"]
	for i := 0; i < 60; i++ {
		line := gopodReminderLine("Tell me about the contest")
		named := false
		for topic, label := range labels {
			if label != "" && topic != "AIWP_livestream" && strings.Contains(line, label) {
				named = true
			}
		}
		if !named {
			t.Fatalf("reminder %d names no topic: %q", i+1, line)
		}
		if rig != "" && strings.Contains(line, rig) {
			t.Fatalf("reminder %d names %q, already covered: %q", i+1, rig, line)
		}
	}
}

// ## EOF

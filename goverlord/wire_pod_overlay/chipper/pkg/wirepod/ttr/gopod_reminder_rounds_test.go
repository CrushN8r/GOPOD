// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_reminder_rounds_test.go
package wirepod_ttr

import (
	"strings"
	"testing"
)

// TestRemindersStartAgain: asked over and over, the reminders keep naming topics - a new round once
// every topic has been named - never a topic already covered.
func TestRemindersStartAgain(t *testing.T) {
	gopodGameReset("test")
	defer gopodGameReset("test")
	labels := gopodBrobotsSettings.InterviewChapters.TopicLabels
	gopodQuestionFactsFor(gopodBrobot1ESN, "Tell me about the rig")
	gopodQuestionFactsFor(gopodBrobot1ESN, "Tell me about the rig") // wide, then deep
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

// the question to try is the first topic reminded of.
func TestTryAskingMatchesTheReminder(t *testing.T) {
	gopodGameReset("test")
	defer gopodGameReset("test")
	labels := gopodBrobotsSettings.InterviewChapters.TopicLabels
	tpl := strings.TrimSpace(gopodBrobotsSettings.ValuePoints.ReminderProberText)
	pre, post, _ := strings.Cut(tpl, "{prober}")
	checked := 0
	for i := 0; i < 40; i++ {
		line := gopodReminderLine("Tell me about the contest")
		at := strings.Index(line, pre)
		if at < 0 {
			continue
		}
		prober := strings.TrimSuffix(strings.TrimSpace(line[at+len(pre):]), strings.TrimSpace(post))
		names, _ := gopodReminderNames.Load("Tell me about the contest")
		list, _ := names.([]string)
		goes := loadScenarioPacketsForMessage(prober)
		if len(list) == 0 || len(goes) == 0 || labels[goes[0].Name] != list[0] {
			t.Errorf("reminder %d: try asking %q isn't about the first topic named (%v): %q", i+1, prober, list, line)
		}
		checked++
	}
	if checked == 0 {
		t.Error("no reminder with a try-asking line")
	}
}

// ## EOF

// the topics named next are the ones after the topic asked about, in the chapters' order (or its
// own links first).
func TestRemindersRideInOrder(t *testing.T) {
	gopodGameReset("test")
	defer gopodGameReset("test")
	ic := gopodBrobotsSettings.InterviewChapters
	var order []string
	for _, c := range ic.Chapters {
		order = append(order, c.Topics...)
	}
	pos := map[string]int{}
	for i, tp := range order {
		pos[ic.TopicLabels[tp]] = i
	}
	q := "Tell me about the contest"
	gopodReminderLine(q)
	names, _ := gopodReminderNames.Load(q)
	list, _ := names.([]string)
	if len(list) < 2 {
		t.Fatalf("reminder names: %v", list)
	}
	seen := 0
	for i := 1; i < len(list); i++ {
		a, b := pos[list[i-1]], pos[list[i]]
		if b <= a {
			seen++
		}
	}
	if seen > 1 {
		t.Errorf("more than one turn back in a short ride: %v", list)
	}
}

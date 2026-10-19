// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_session_replay_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSessionReplayPlan: a session score becomes the replay's acts - reporters,
// questions with their logged answers (cue kept), songs started again (their
// own lines skipped), the interview's section from the score, and the thanks
// ending live (2026-10-12, gopod_session_replay.go).
func TestSessionReplayPlan(t *testing.T) {
	live, iv := gopodSessionLiveSection, strings.ToUpper(gopodRunText("{run}", "interview"))
	st := func(note, who, to, section, text, say string) gopodSessionStep {
		return gopodSessionStep{Note: note, Speaker: who, To: to, Section: section, Say: say, text: text}
	}
	steps := []gopodSessionStep{
		st("reporter_line", "Brobot 3", "", live, "Grab the mic.", ""),
		st("operator_line", "Operator", "Brobot 1", live, "what is gopod", ""),
		st("answer", "Brobot 1", "", live, "It's a stage.", "happy___ It's a stage."),
		st("operator_line", "Operator", "Brobot 1", live, "start fart joke", ""),
		st("system_line", "Goverlord", "", "FART JOKE", "Fart joke started.", ""),
		st("say", "Brobot 2", "", "FART JOKE", "Hey bro", ""),
		st("operator_line", "Operator", "Brobot 1", live, "start interview", ""),
		st("system_line", "Goverlord", "", iv, "Interview started.", ""),
		st("chapter", "Chapter", "", iv, "Who we are", ""),
		st("operator_line", "Operator", "Brobot 1", iv, "Who are you?", ""),
		st("answer", "Brobot 1", "", iv, "Brobot 1.", "thinking___ Brobot 1."),
		st("reporter_line", "Brobot 4", "", iv, "Did you see the arm?", ""),
		st("answer", "Brobot 2", "", iv, "I'm here.", ""),
		st("say", "Brobot 2", "", iv, "Verified.", ""),
		st("operator_line", "Operator", "Brobots", live, "GOPOD yourself", ""),
		st("say", "Brobot 1", "", live, "Join the crew!", ""),
		st("operator_line", "Operator", "Brobot 2", live, "thanks brobots", ""),
		st("answer", "Brobot 2", "", live, "Bye!", ""),
	}
	var got []string
	for _, a := range gopodReplayPlan(steps) {
		line := a.Kind + ":" + a.Text
		for _, ans := range a.Answers {
			line += "|" + ans.Brobot + "=" + ans.Say
		}
		got = append(got, line)
	}
	want := []string{
		"reporter:Grab the mic.",
		"question:what is gopod|Brobot 1=happy___ It's a stage.",
		"song:start fart joke",
		"show:start interview",
		"system:Interview started.",
		"chapter:Who we are",
		"question:Who are you?|Brobot 1=thinking___ Brobot 1.",
		"reporter:Did you see the arm?",
		"handoff:|Brobot 2=I'm here.",
		"handoff:|Brobot 2=Verified.",
		"command:GOPOD yourself",
		"thanks:thanks brobots",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("plan:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if n, ok := gopodReplayName("replay 095819_session"); !ok || n != "095819_session" {
		t.Fatalf("replay name = %q %v", n, ok)
	}
	// A kept session on this machine, if there is one: its plan is printed.
	if home, err := os.UserHomeDir(); err == nil {
		dir := filepath.Join(home, "gopod_tts", "sessions", "2026-10-12", "095819_session")
		if s, err := gopodReplayLoad(dir); err == nil {
			for _, a := range gopodReplayPlan(s) {
				t.Logf("%s %s %q answers=%d", a.Kind, a.To, a.Text, len(a.Answers))
			}
		}
	}
}

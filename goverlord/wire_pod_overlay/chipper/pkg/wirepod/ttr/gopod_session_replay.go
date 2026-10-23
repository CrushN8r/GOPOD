// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_session_replay.go
package wirepod_ttr

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// Replay a session (operator 2026-10-12: "Can you expand/build a session to run
// a previous session run? Like if I wanted to replay the kept session I just
// made * replay [session-name] ... run the entire session, then run complete";
// "Should I build it like that? * yes please"; "should the cues and moves be
// logged into the score so a replay includes them? * yes please").
//
// "replay <name>" plays a session folder's score (story.md + knobs.json,
// gopod_session_log.go) back through the live chat, in order:
//   - the reporters' lines in their voices;
//   - each operator line as a typed line to its Brobot, and his logged answer
//     said through the chat-reply path in place of the LLM (the draft's
//     scripted answer, gopod_interview_script.go): its cue, moves, bubble,
//     highlight and speech; an answer nobody asked for is said as a handoff;
//   - a song's start line starts the song again (its own score plays it, and
//     the replay waits for it to end); the interview's section plays from the
//     score, with Goverlord's lines in it;
//   - a command line ("stop run", "GOPOD yourself") runs again, so its answer
//     comes from Wire-Pod as before; a thanks plays the ending live.
// Goverlord says session.started_text / complete_text / stopped_text with the
// session's name. "stop run" stops it after the line being said. Not a run in
// the one-run-at-a-time sense: the songs in it check in and out as before.

var gopodReplayPlaying, gopodReplayStop atomic.Bool

// gopodReplayNow: the playing replay's shown name ("Session 095819").
var gopodReplayNow atomic.Value

// gopodReplayMatch: "replay <name>" (session.replay_word).
func gopodReplayName(text string) (string, bool) {
	word := strings.ToLower(strings.TrimSpace(gopodBrobotsSettings.Session.ReplayWord))
	lower := strings.ToLower(strings.TrimSpace(text))
	if word == "" || !strings.HasPrefix(lower, word+" ") {
		return "", false
	}
	return strings.TrimSpace(strings.TrimRight(strings.TrimSpace(text)[len(word):], ".!?")), true
}

func gopodReplayMatch(text string) bool {
	_, ok := gopodReplayName(text)
	return ok
}

// gopodStopSessionMatch: a session.stop_session_words line ("stop session").
func gopodStopSessionMatch(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	for _, w := range gopodBrobotsSettings.Session.StopSession {
		if w != "" && gopodHasWholeWords(lower, strings.ToLower(w)) {
			return true
		}
	}
	return false
}

// gopodStopSessionCommand: the replay stops after the line being said, with the
// song going in it; nothing to stop gets session.no_run_text.
func gopodStopSessionCommand() {
	if !gopodReplayPlaying.Load() {
		logger.Println("GOPOD_REPLAY stop session: nothing playing")
		gopodSessionSay(gopodBrobotsSettings.Session.NoRunText)
		return
	}
	gopodReplayStop.Store(true)
	gopodStopRuns("stop session")
	logger.Println("GOPOD_REPLAY stop session: stops after the line being said")
}

// gopodReplayShownName: the replay's name in its lines ("Session 095819").
func gopodReplayShownName(dir string) string {
	name := strings.TrimSuffix(filepath.Base(dir), "_session")
	if t := gopodBrobotsSettings.Session.ReplayName; t != "" {
		return strings.ReplaceAll(t, "{name}", name)
	}
	return name
}

// gopodReplayFind: the session folder named name ("095819_session", "095819",
// or "2026-10-12/095819_session"), the newest first when several match.
func gopodReplayFind(name string) (string, error) {
	root, _, _ := gopodRecordRoot()
	want := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), " ", "_"))
	dirs, _ := filepath.Glob(filepath.Join(root, "*", "*_session"))
	for i := len(dirs) - 1; i >= 0; i-- {
		d := dirs[i]
		base := strings.ToLower(filepath.Base(d))
		rel := strings.ToLower(filepath.Base(filepath.Dir(d)) + "/" + base)
		if base == want || strings.TrimSuffix(base, "_session") == want || rel == want {
			if _, err := os.Stat(filepath.Join(d, "knobs.json")); err == nil {
				return d, nil
			}
		}
	}
	return "", errors.New("no session folder named " + name)
}

var gopodReplayStoryStep = regexp.MustCompile(`(?m)^## STEP (\S+)\s*\n\s*\n> TEXT: ?(.*)$`)

// gopodReplayLoad: the score's steps, each with its words from story.md (the
// words the operator may have edited there).
func gopodReplayLoad(dir string) ([]gopodSessionStep, error) {
	data, err := os.ReadFile(filepath.Join(dir, "knobs.json"))
	if err != nil {
		return nil, err
	}
	var knobs struct {
		Steps []gopodSessionStep `json:"steps"`
	}
	if err := json.Unmarshal(data, &knobs); err != nil {
		return nil, err
	}
	story, _ := os.ReadFile(filepath.Join(dir, "story.md"))
	words := map[string]string{}
	for _, m := range gopodReplayStoryStep.FindAllStringSubmatch(string(story), -1) {
		words[m[1]] = strings.TrimSpace(m[2])
	}
	for i := range knobs.Steps {
		knobs.Steps[i].text = words[knobs.Steps[i].StepID]
	}
	return knobs.Steps, nil
}

// gopodReplayAct: one thing the replay does.
type gopodReplayAct struct {
	Kind    string // "reporter", "question", "command", "song", "show", "system", "handoff", "thanks"
	Who     string
	To      string
	How     string
	Text    string
	Answers []gopodDraftAnswer // a question's own answers, said in place of the LLM
}

// gopodReplayPlan turns the score into what the replay does, in order.
func gopodReplayPlan(steps []gopodSessionStep) []gopodReplayAct {
	interview := strings.ToUpper(gopodRunText("{run}", "interview"))
	var out []gopodReplayAct
	used := map[int]bool{}
	answerOf := func(st gopodSessionStep) gopodDraftAnswer {
		say := st.Say
		if say == "" {
			say = st.text
		}
		return gopodDraftAnswer{Brobot: st.Speaker, Say: say, Moves: st.Moves}
	}
	for i, st := range steps {
		if used[i] {
			continue
		}
		live := st.Section == gopodSessionLiveSection
		if !live && st.Section != interview {
			continue // a song's own lines: the song plays them again
		}
		switch st.Note {
		case "reporter_line":
			out = append(out, gopodReplayAct{Kind: "reporter", Who: st.Speaker, Text: st.text})
		case "chapter":
			// A chapter title shows live only while a draft plays (gopod_interview_chapters.go).
			if !live {
				out = append(out, gopodReplayAct{Kind: "chapter", Who: st.Speaker, Text: st.text})
			}
		case "system_line":
			// Outside the interview, Goverlord's lines answer commands that run again.
			if !live {
				out = append(out, gopodReplayAct{Kind: "system", Text: st.text})
			}
		case "answer":
			out = append(out, gopodReplayAct{Kind: "handoff", Who: st.Speaker, Answers: []gopodDraftAnswer{answerOf(st)}})
		case "say":
			// In the interview, a Brobot's handoff line; outside it, a command's
			// answer ("GOPOD Yourself"), said again by the command.
			if !live {
				out = append(out, gopodReplayAct{Kind: "handoff", Who: st.Speaker, Answers: []gopodDraftAnswer{answerOf(st)}})
			}
		case "operator_line":
			how := "Typed text"
			if strings.HasPrefix(st.Via, "KP") || st.Via == "Numpad" {
				how = "Numpad text"
			}
			act := gopodReplayAct{Kind: "question", To: st.To, How: how, Text: st.text}
			switch {
			case gopodScriptCommandMatch(st.text):
				act.Kind = "show" // the interview: its section plays from the score
			case gopodIsThanks(st.text):
				act.Kind = "thanks"
			case gopodTypedCommandMatch(st.text) || (i+1 < len(steps) && steps[i+1].Section != gopodSessionLiveSection && steps[i+1].Section != interview):
				// A run command, or a line the score shows starting a song's section.
				act.Kind = "song"
			case gopodReplayMatch(st.text) || gopodStopSessionMatch(st.text):
				continue // a replay in the session isn't played again inside this one
			case gopodStopRunMatch(st.text) || strings.HasPrefix(strings.ToLower(st.text), "start "):
				act.Kind = "command"
			}
			if act.Kind == "question" {
				// His first answer (each Brobot's, for "Brobots") before the next operator line.
				taken := map[string]bool{}
				for j := i + 1; j < len(steps) && steps[j].Note != "operator_line"; j++ {
					a := steps[j]
					if a.Note != "answer" || used[j] || taken[a.Speaker] || (st.To != "Brobots" && a.Speaker != st.To) {
						continue
					}
					taken[a.Speaker] = true
					used[j] = true
					act.Answers = append(act.Answers, answerOf(a))
				}
				if len(act.Answers) == 0 {
					act.Kind = "command" // answered by Wire-Pod itself ("GOPOD Yourself")
				}
			}
			out = append(out, act)
			if act.Kind == "thanks" {
				return out // the ending plays live from here
			}
		}
	}
	return out
}

// gopodReplayCommand: "replay <name>" - plays that session. True when it was one.
func gopodReplayCommand(text string) bool {
	name, ok := gopodReplayName(text)
	if !ok {
		return false
	}
	ss := gopodBrobotsSettings.Session
	if gopodReplayPlaying.Load() || gopodRunActive() {
		gopodSessionSay(ss.BusyText)
		return true
	}
	dir, err := gopodReplayFind(name)
	if err != nil {
		logger.Println("GOPOD_REPLAY " + err.Error())
		gopodSessionSay(strings.ReplaceAll(ss.NotFoundText, "{run}", name))
		return true
	}
	steps, err := gopodReplayLoad(dir)
	if err != nil {
		logger.Println("GOPOD_REPLAY unable to read " + dir + ": " + err.Error())
		gopodSessionSay(strings.ReplaceAll(ss.NotFoundText, "{run}", name))
		return true
	}
	gopodReplayPlaying.Store(true)
	gopodReplayStop.Store(false)
	// 2026-10-12 (operator: "sessions cannot be recorded again"): a recording
	// going stops here and is kept; Record? can't start during the replay.
	GopodRecordSessionSet(false)
	gopodReplayNow.Store(gopodReplayShownName(dir))
	go gopodReplayPlay(gopodReplayShownName(dir), dir, gopodReplayPlan(steps))
	return true
}

func gopodReplayPlay(name, dir string, plan []gopodReplayAct) {
	ss := gopodBrobotsSettings.Session
	gap := time.Duration(gopodBrobotsSettings.InterviewScript.GapSeconds * float64(time.Second))
	defer gopodReplayPlaying.Store(false)
	logger.Println("GOPOD_REPLAY start dir=" + dir + " acts=" + strconv.Itoa(len(plan)))
	GopodSystemSayHeard(strings.ReplaceAll(ss.StartedText, "{run}", name), 30*time.Second)
	var at []string
	for _, act := range plan {
		for _, a := range act.Answers {
			for _, m := range a.Moves {
				if m.At != "" {
					at = append(at, strings.ToLower(m.At))
				}
			}
		}
	}
	gopodScript.Lock()
	gopodScript.movesAt = at
	gopodScript.Unlock()
	defer func() {
		gopodScript.Lock()
		gopodScript.pending, gopodScript.ask, gopodScript.movesAt = map[string]gopodDraftAnswer{}, map[string]string{}, nil
		gopodScript.Unlock()
	}()
	// 2026-10-12 (operator: the LLM warm-up at the start of any run, kept
	// sessions too): the opener's warm-up, with the session's first reporter
	// lines inside it - warming while they play, or ready first when warm.
	n := 0
	for n < len(plan) && plan[n].Kind == "reporter" {
		n++
	}
	gopodRunWarmup(gopodWarmupReadyLine(gopodBrobotsSettings.LLM.ReadyText, gopodBrobotsSettings.LLM.WarmupDoneText), func() {
		for _, act := range plan[:n] {
			if host := gopodReporterHost(act.Who); host != "" && !gopodReplayStop.Load() {
				gopodReporterSay(host, act.Text, true, true)
			}
		}
	})
	for _, act := range plan[n:] {
		// 2026-10-17 (survey E8): the pause is in the journal too.
		if gopodPageHidden.Load() {
			logger.Println("GOPOD_REPLAY paused: the chat page is away")
			for gopodPageHidden.Load() && !gopodReplayStop.Load() {
				time.Sleep(500 * time.Millisecond)
			}
			logger.Println("GOPOD_REPLAY the chat page is back")
		}
		if gopodReplayStop.Load() {
			logger.Println("GOPOD_REPLAY stopped")
			GopodSystemSayHeard(strings.ReplaceAll(ss.StoppedText, "{run}", name), 30*time.Second)
			return
		}
		to := gopodBrobotESN(act.To)
		if act.To == "Brobots" || to == "" {
			to = gopodTypedAll
		}
		switch act.Kind {
		case "reporter":
			if host := gopodReporterHost(act.Who); host != "" {
				gopodReporterSay(host, act.Text, true, true)
			}
		case "system":
			GopodSystemSayHeard(act.Text, 30*time.Second)
		case "chapter":
			// Shown and said in the system voice, as in the draft (operator
			// 2026-10-12: "Chapter bubbles aren't spoken").
			// 2026-10-23: the chapter as Goverlord's voiced line, like the live run's.
			if sr := gopodBrobotsSettings.SignOff.SystemRole; sr != "" {
				gopodReporterSay(sr, gopodChapterSaid(act.Text), true, true)
			} else {
				GopodChatFeedAdd(act.Who, act.Text)
			}
		case "handoff":
			for _, a := range act.Answers {
				if say := strings.TrimSpace(gopodDisplayReply(a.Say)); say != "" {
					gopodRecordSay(a.Brobot, say)
				}
			}
		case "show":
			gopodShowCommandLine(to, act.Text, act.How)
		case "question":
			gopodScript.Lock()
			for _, a := range act.Answers {
				esn := gopodBrobotESN(a.Brobot)
				gopodScript.pending[esn], gopodScript.ask[esn] = a, act.Text
			}
			gopodScript.Unlock()
			logger.Println("GOPOD_REPLAY question to " + act.To + ": " + act.Text)
			if err := gopodOperatorLine(to, act.Text, act.How); err != nil {
				logger.Println("GOPOD_REPLAY " + err.Error())
			}
			gopodScriptWaitAnswered(len(act.Answers))
			gopodScript.Lock()
			gopodScript.pending, gopodScript.ask = map[string]gopodDraftAnswer{}, map[string]string{}
			gopodScript.Unlock()
		case "song":
			if err := gopodOperatorLine(to, act.Text, act.How); err != nil {
				logger.Println("GOPOD_REPLAY " + err.Error())
			}
			// The song checks in, plays its own score and checks out.
			for wait := time.Now().Add(20 * time.Second); !gopodRunActive() && time.Now().Before(wait); {
				time.Sleep(250 * time.Millisecond)
			}
			for gopodRunActive() {
				time.Sleep(500 * time.Millisecond)
			}
		case "command":
			if err := gopodOperatorLine(to, act.Text, act.How); err != nil {
				logger.Println("GOPOD_REPLAY " + err.Error())
			}
			gopodScriptWaitAnswered(0)
			time.Sleep(3 * time.Second)
			for gopodReplyBusy() {
				time.Sleep(200 * time.Millisecond)
			}
		case "thanks":
			// The complete line first: the ending closes the session.
			GopodSystemSayHeard(strings.ReplaceAll(ss.CompleteText, "{run}", name), 30*time.Second)
			gopodReplayPlaying.Store(false)
			logger.Println("GOPOD_REPLAY done, the ending plays live")
			if err := gopodOperatorLine(to, act.Text, act.How); err != nil {
				logger.Println("GOPOD_REPLAY " + err.Error())
			}
			return
		}
		time.Sleep(gap)
	}
	logger.Println("GOPOD_REPLAY done")
	GopodSystemSayHeard(strings.ReplaceAll(ss.CompleteText, "{run}", name), 30*time.Second)
}

// ## EOF

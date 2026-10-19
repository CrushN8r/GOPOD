// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_session_dry_test.go
package wirepod_ttr

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/sashabaranov/go-openai"
)

// The session dry run (operator 2026-10-11: "Full run done by you, no robots,
// nothing spoken"; "I want a full run with a full account of all events of a
// session starting from timer 0 until ended, and timer reset to 0"). One
// session through Wire-Pod's own paths - the session log (gopod_session_log.go),
// the chat feed, the Brobots' real chat path with the real LLM, the session's
// run lines, the songs from their own story.md and knobs.json, and the song
// engine's own dry run for their runs/ - with no robot and no sound: a voice
// line is looked up in the voice library and logged (gopodDryVoice), never
// played. session_dry_run.sh runs it.
//
//   live:  "Do you like to party?" and "What's the weather like?" to both
//   runs:  start interview, start bingo test, start baby brobots
//   live:  "How did bingo go for you?" to both, then "thx bros" (the ending)

var gopodDryStoryStep = regexp.MustCompile(`^## STEP (\S+)\s*$`)
var gopodDryStorySection = regexp.MustCompile(`^## SECTION:\s*(.+?)\s*$`)

type gopodDrySongStep struct {
	StepID   string `json:"step_id"`
	Note     string `json:"note"`
	Speaker  string `json:"speaker"`
	ShowBall string `json:"show_ball"`
	text     string
}

func TestGopodSessionDryRun(t *testing.T) {
	out := os.Getenv("GOPOD_SESSION_DRY_OUT")
	if out == "" {
		t.Skip("session dry run: set GOPOD_SESSION_DRY_OUT (session_dry_run.sh)")
	}
	endpoint, model := os.Getenv("GOPOD_HARVEST_ENDPOINT"), os.Getenv("GOPOD_HARVEST_MODEL")
	if endpoint == "" || model == "" {
		t.Fatal("GOPOD_HARVEST_ENDPOINT and GOPOD_HARVEST_MODEL are needed")
	}
	// Nothing live: chat memory, feed and state files in the run folder; the
	// session folder under out/sessions.
	t.Setenv("GOPOD_SESSION_DIR", filepath.Join(out, "chat_memory"))
	t.Setenv("GOPOD_RECORD_SESSION_DIR", filepath.Join(out, "sessions"))
	gopodChatFeedPath = filepath.Join(out, "chat_feed.json")
	gopodRecordStatePath = filepath.Join(out, "record_state.json")
	gopodSpeakingNowPath = filepath.Join(out, "speaking_now.json")
	gopodSharedBrobotChatStart = time.Now()
	vars.APIConfig.Knowledge.SaveChat = true
	vars.APIConfig.Knowledge.Model = model
	gopodBingoPlayed.Store(false)
	gopodWarmGateClosed.Store(false)
	gopodSession.Lock()
	gopodSession.run = nil
	gopodSession.Unlock()

	// A voice line: looked up in the library, logged, never played.
	gopodDryVoice = func(host, text string, show bool) {
		_, _, cached := gopodReporterCached(host, gopodReporterSays(text))
		state := "new (would be made, then played)"
		if cached {
			state = "in the voice library"
		}
		if show {
			GopodChatFeedAdd(host, text)
		}
		gopodSessionLogLine(gopodRecordLine{Source: "voice", Who: host, Text: text + " [" + state + "]"}, "", "")
		t.Logf("VOICE %s: %s [%s]", host, text, state)
	}
	defer func() { gopodDryVoice = nil }()

	// Every log line into the session's wirepod.log, as inside Wire-Pod.
	logger.Init()
	go gopodRecordLogTee()
	GopodSessionLogStart()
	gopodSessionLog.Lock()
	dir := gopodSessionLog.dir
	gopodSessionLog.Unlock()
	if dir == "" {
		t.Fatal("no session folder")
	}
	t.Logf("SESSION %s", dir)

	cfg := openai.DefaultConfig("ollama")
	cfg.BaseURL = endpoint
	client := openai.NewClientWithConfig(cfg)
	esnOf := map[string]string{"Brobot 1": gopodBrobot1ESN, "Brobot 2": gopodBrobot2ESN}

	// operator: a typed line to a Brobot (or both), shown and read out.
	operator := func(to, text string) {
		gopodChatFeedAddOperator(to, text, "Keyboard")
		gopodReporterReadsOperator(text, "Typed text", to)
		t.Logf("OPERATOR Keyboard to %s: %s", to, text)
	}
	// answer: one Brobot's answer through the live chat's path (the harvest's).
	answer := func(b, text string) {
		esn := esnOf[b]
		req := CreateAIReq(text, esn, false, false)
		req.Stream = false
		resp, err := client.CreateChatCompletion(context.Background(), req)
		if err != nil || len(resp.Choices) == 0 {
			t.Logf("ANSWER %s: none: %v", b, err)
			return
		}
		raw := resp.Choices[0].Message.Content
		raw = gopodFixSelfName(esn, gopodCorrectNames(gopodStripLeadingAnimationWords(raw)))
		raw = gopodDropWebLinks(raw)
		raw = gopodDropLLMReminder(raw)
		raw = gopodTightenReply(raw, text, esn)
		if snippet := gopodScriptedSnippet(esn, text, "start", ""); snippet != "" {
			raw = gopodStitchSnippet(snippet, raw)
		}
		if snippet := gopodScriptedSnippet(esn, text, "end", raw); snippet != "" {
			raw = gopodStitchEndSnippet(snippet, raw)
		}
		result := buildGOPODStreamResult(gopodAnimationMarkerToInternal(raw))
		// What would be sent to the Brobot, as live logs it (kgsim.go).
		for _, chunk := range result.AudibleChunks {
			packet := gopodPronounceCommandPacket(commandPacketFromFilteredChunk(chunk))
			logger.LogDebugUI("FILTERED_SPLIT_SENT:\n" + gopodVaryDefaultAnimation(packet, result.Raw))
		}
		result.Raw = gopodDisplayReply(raw)
		recordCompletedGOPODSessionExchange(esn, text, result, time.Now().UTC())
		t.Logf("ANSWER %s: %s", b, result.Raw)
	}
	ask := func(to, text string) {
		operator(to, text)
		brobots := []string{to}
		if to == "Brobots" {
			brobots = []string{"Brobot 1", "Brobot 2"}
		}
		for _, b := range brobots {
			answer(b, gopodSpellNames(text))
		}
	}

	// --- live ---
	ask("Brobots", "Do you like to party?")
	ask("Brobots", "What's the weather like?")

	// --- start interview: the interview song, played as the chat plays it ---
	operator("Brobots", "start interview")
	if _, draft, err := gopodLoadDraft(); err != nil {
		t.Errorf("interview: %v", err)
	} else if ok, _ := GopodRunStart("interview", "interview", 0); !ok {
		t.Error("interview refused")
	} else {
		for _, l := range draft.Opening {
			GopodChatFeedAdd(gopodDryHost(l.Who), l.Text)
		}
		for _, q := range draft.Questions {
			for _, l := range q.Before {
				GopodChatFeedAdd(l.Who, l.Text)
			}
			to := q.To
			if to == "" {
				to = "Brobot 1"
			}
			gopodChatFeedAddEventVia("Operator", to, gopodDryHost(q.ReadBy), "Keyboard", q.Ask)
			for _, a := range q.Answers {
				say := gopodScriptFill(esnOf[a.Brobot], a.Say)
				cue := a.Cue
				if cue == "" {
					cue = "thinking"
				}
				result := buildGOPODStreamResult(gopodAnimationMarkerToInternal(cue + gopodAnimationMarker + " " + say))
				for _, chunk := range result.AudibleChunks {
					logger.LogDebugUI("FILTERED_SPLIT_SENT:\n" + gopodPronounceCommandPacket(commandPacketFromFilteredChunk(chunk)))
				}
				gopodSessionLogAnswer(esnOf[a.Brobot], q.Ask, gopodStreamResult{Raw: say, FilteredPacket: result.FilteredPacket})
			}
			for _, l := range q.After {
				GopodChatFeedAdd(l.Who, l.Text)
			}
		}
		GopodRunEnd("interview", "done")
	}

	// --- the songs: their own story.md and knobs.json, and the engine's dry run ---
	song := func(command, runName, songID string) {
		operator("Brobots", command)
		songDir, err := gopodFindSong(songID)
		if err != nil {
			t.Errorf("%s: %v", songID, err)
			return
		}
		if ok, _ := GopodRunStart(runName, "song", 0); !ok {
			t.Errorf("%s refused", runName)
			return
		}
		for _, s := range gopodDrySongSteps(t, songDir) {
			who := map[string]string{"brobot_1": "Brobot 1", "brobot_2": "Brobot 2"}[s.Speaker]
			if s.ShowBall != "" {
				GopodChatFeedAdd("Bingo Ball", s.ShowBall)
			}
			switch {
			case s.text == "":
			case s.Note == "system_line":
				GopodSystemSayHeard(s.text, 0)
			case s.Note == "brobots_ready_together":
				GopodChatFeedAdd("Brobot 1", s.text)
				GopodChatFeedAdd("Brobot 2", s.text)
			case s.Note == "pause": // a timed song's caption (the nap)
				GopodChatFeedAdd("Narrator", s.text)
			default:
				GopodChatFeedAdd(who, s.text)
			}
			if s.Note == "rattle" {
				GopodChatFeedAdd("Countdown", "2")
				GopodChatFeedAdd("Countdown", "done")
			}
		}
		// The song engine's own dry run, for its runs/ files.
		cmd := exec.Command("python3", "run_golden_song_001.py")
		cmd.Dir = filepath.Join(gopodSongsDir(), "tools")
		cmd.Env = append(os.Environ(), "GOPOD_GOLDEN_SONG_ID="+songID)
		cmd.Stdin = nil
		if data, err := cmd.CombinedOutput(); err != nil {
			t.Errorf("%s engine dry run: %v\n%s", songID, err, gopodDryTail(string(data), 600))
		}
		if songID == "brobots_bingo" {
			gopodBingoPlayed.Store(true)
		}
		GopodRunEnd(runName, "done")
		time.Sleep(time.Second) // its runs/ copied into the session
	}
	song("start bingo test", "Bingo Test Live", "brobots_bingo")
	song("start baby brobots", "Start Baby Brobots", "brobots_baby_robots_sleep")

	// --- live again, then the thanks ending ---
	ask("Brobots", "How did bingo go for you?")
	operator("Brobots", "thx bros")
	for _, b := range []string{"Brobot 1", "Brobot 2"} {
		answer(b, "thx bros")
	}
	rv := gopodBrobotsSettings.ReporterVoices
	for _, l := range rv.Closing {
		GopodChatFeedAdd(l.Host, l.Text)
	}
	for _, l := range gopodBrobotsSettings.SignOff.Lines {
		GopodChatFeedAdd(l.Who, l.Text)
	}
	gopodYoureWelcome()
	steps := gopodSessionLogSteps()
	GopodSessionLogEnd("thanks")
	t.Logf("SESSION_DONE %s steps=%d", dir, len(steps))
	if len(steps) < 40 {
		t.Errorf("only %d steps in the session", len(steps))
	}
}

// gopodDryHost: a reporter label ("Reporter 1, Brobot 3") back to its host ("Brobot 3").
func gopodDryHost(label string) string {
	for host, l := range gopodBrobotsSettings.ReporterVoices.Labels {
		if l == label || host == label {
			return host
		}
	}
	return label
}

// gopodDrySongSteps: a song's steps in order (knobs.json), each with its words
// (story.md), a combo's words on its first part.
func gopodDrySongSteps(t *testing.T, dir string) []gopodDrySongStep {
	var knobs struct {
		Steps []gopodDrySongStep `json:"steps"`
	}
	data, err := os.ReadFile(filepath.Join(dir, "knobs.json"))
	if err != nil || json.Unmarshal(data, &knobs) != nil {
		t.Errorf("%s knobs.json: %v", dir, err)
		return nil
	}
	words := map[string]string{}
	f, err := os.Open(filepath.Join(dir, "story.md"))
	if err == nil {
		defer f.Close()
		sc := bufio.NewScanner(f)
		step := ""
		for sc.Scan() {
			line := sc.Text()
			if m := gopodDryStoryStep.FindStringSubmatch(line); m != nil {
				step = m[1]
			} else if gopodDryStorySection.MatchString(line) {
				step = ""
			} else if step != "" && strings.HasPrefix(line, "> TEXT:") {
				words[step] = strings.TrimSpace(strings.TrimPrefix(line, "> TEXT:"))
			}
		}
	}
	for i := range knobs.Steps {
		knobs.Steps[i].text = words[knobs.Steps[i].StepID]
	}
	return knobs.Steps
}

func gopodDryTail(s string, n int) string {
	if len(s) > n {
		return s[len(s)-n:]
	}
	return s
}

// ## EOF

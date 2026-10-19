// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_harvest_test.go
package wirepod_ttr

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/sashabaranov/go-openai"
)

// The harvest dry run (operator 2026-10-06: "start with a single full harvest
// dry run (no actual robots) and set up the take sheet ... one good end to end
// of all topics/value points"). Not a unit test: it runs only when
// GOPOD_HARVEST_OUT is set (run with `goverlord/wire_pod_overlay/harvest_dry_run.sh`).
//
// Every question in gopod_probes/interview_harvest/harvest_questions.json goes
// through the Brobots' real path - the same prompt (CreateAIReq: persona, the
// value-point list's next point, shared chat memory), the same LLM, the same
// reply shaping (names, links, flavour lines, the reminder, display and speech
// split) - with no robot: nothing is said, moved or shown, and no live file is
// written (the chat memory, chat feed and Record Session go to the run folder).
// Each answer is a take. The run writes takes.json to GOPOD_HARVEST_OUT and adds
// its takes to the take sheet (GOPOD_HARVEST_SHEET, gopod_take_sheet.go): one
// entry per value point, in interview order, each with its takes - mark the
// keeper, edit its words.

type gopodHarvestQuestion struct {
	Chapter string `json:"chapter"`
	Topic   string `json:"topic"`
	To      string `json:"to"`
	Ask     string `json:"ask"`
}

var gopodHarvestCue = regexp.MustCompile(`^\s*\**([A-Za-z][A-Za-z0-9-]*)\**\s*___`)

func TestGopodHarvestRun(t *testing.T) {
	out := os.Getenv("GOPOD_HARVEST_OUT")
	if out == "" {
		t.Skip("harvest dry run: set GOPOD_HARVEST_OUT (harvest_dry_run.sh)")
	}
	endpoint, model := os.Getenv("GOPOD_HARVEST_ENDPOINT"), os.Getenv("GOPOD_HARVEST_MODEL")
	if endpoint == "" || model == "" {
		t.Fatal("GOPOD_HARVEST_ENDPOINT and GOPOD_HARVEST_MODEL are needed")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	// Nothing live: the chat memory, feed and Record Session live in the run folder.
	t.Setenv("GOPOD_SESSION_DIR", filepath.Join(out, "chat_memory"))
	t.Setenv("GOPOD_RECORD_SESSION_DIR", filepath.Join(out, "record"))
	gopodChatFeedPath = filepath.Join(out, "chat_feed.json")
	gopodRecordStatePath = filepath.Join(out, "record_state.json")
	gopodSharedBrobotChatStart = time.Now()
	vars.APIConfig.Knowledge.SaveChat = true
	vars.APIConfig.Knowledge.Model = model
	gopodBingoPlayed.Store(false)
	gopodCoverage.Lock()
	gopodCoverage.revealed = map[string]int{}
	gopodCoverage.reminded = map[string]bool{}
	gopodCoverage.Unlock()

	var qfile struct {
		Input     string                 `json:"input"`
		EndText   string                 `json:"end_text"`
		Questions []gopodHarvestQuestion `json:"questions"`
	}
	qpath := filepath.Join(gopodChipperRoot(), "gopod_probes/interview_harvest/harvest_questions.json")
	// 2026-10-08: another questions file (harvest_dry_run.sh questions <file>).
	if alt := os.Getenv("GOPOD_HARVEST_QUESTIONS"); alt != "" {
		qpath = alt
	}
	data, err := os.ReadFile(qpath)
	if err != nil || json.Unmarshal(data, &qfile) != nil {
		t.Fatalf("questions: %v", err)
	}
	cfg := openai.DefaultConfig("ollama")
	cfg.BaseURL = endpoint
	client := openai.NewClientWithConfig(cfg)
	esnOf := map[string]string{"Brobot 1": gopodBrobot1ESN, "Brobot 2": gopodBrobot2ESN}
	run := filepath.Base(out)
	_, counts := gopodPackPointCounts()

	started := time.Now()
	var takes []gopodHarvestTake
	// The draft performance (operator 2026-10-07: "Do a full dry run ... fully
	// featured, fully logged ... I need to replay a 'draft performance run' to
	// review & edit"): the whole show in order - the reporters' opening, each
	// question as its reporter reads it, the chapter title, the Brobot's answer,
	// the pictures, the sign-off, the score, keep or delete, the closing.
	// Wire-Pod plays its questions and answers back word for word
	// (gopod_interview_script.go).
	rv := gopodBrobotsSettings.ReporterVoices
	label := func(host string) string {
		if l := rv.Labels[host]; l != "" {
			return l
		}
		return host
	}
	draft := gopodDraft{About: gopodDraftAbout, Run: run, Model: model}
	for _, l := range rv.Opening {
		draft.Opening = append(draft.Opening, gopodDraftLine{Who: label(l.Host), Text: l.Text})
	}
	feedSeen := 0
	feedNew := func() []gopodDraftLine {
		gopodChatFeed.Lock()
		defer gopodChatFeed.Unlock()
		var lines []gopodDraftLine
		for _, e := range gopodChatFeed.events {
			if e.ID > feedSeen {
				feedSeen = e.ID
				lines = append(lines, gopodDraftLine{Who: e.Who, Text: e.Text})
			}
		}
		return lines
	}
	// answer is one Brobot's answer through the live chat's path.
	answer := func(n int, esn, text string) (gopodHarvestTake, bool) {
		before := gopodHarvestRevealed()
		req := CreateAIReq(text, esn, false, false)
		req.Stream = false
		after := gopodHarvestRevealed()
		var points []string
		for topic, k := range after {
			for p := before[topic] + 1; p <= k && p <= counts[topic]; p++ {
				points = append(points, fmt.Sprintf("%s %d/%d", topic, p, counts[topic]))
			}
		}
		sort.Strings(points)
		askStart := time.Now()
		resp, err := client.CreateChatCompletion(context.Background(), req)
		if err != nil || len(resp.Choices) == 0 {
			t.Logf("take %d: no answer: %v", n, err)
			return gopodHarvestTake{}, false
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
		cue := ""
		if m := gopodHarvestCue.FindStringSubmatch(raw); m != nil {
			cue = m[1]
		}
		display := gopodDisplayReply(raw)
		result := buildGOPODStreamResult(gopodAnimationMarkerToInternal(raw))
		recordCompletedGOPODSessionExchange(esn, text, result, time.Now().UTC())
		return gopodHarvestTake{
			Take: fmt.Sprintf("%s#%02d", run, n), Points: points, To: gopodBrobotName(esn),
			Input: qfile.Input, Cue: cue, Say: strings.TrimSpace(display),
			Spoken: gopodSpokenPieceTexts(result.AudibleChunks), LLMSeconds: time.Since(askStart).Seconds(),
		}, true
	}
	asks := append([]gopodHarvestQuestion{}, qfile.Questions...)
	if qfile.EndText != "" {
		asks = append(asks, gopodHarvestQuestion{To: "Brobots", Ask: qfile.EndText})
	}
	for i, q := range asks {
		host := rv.ReadBy[q.To]
		if host == "" {
			host = rv.OperatorHost
		}
		dq := gopodDraftQuestion{N: i + 1, Chapter: q.Chapter, Topic: q.Topic, To: q.To, Ask: q.Ask, ReadBy: label(host)}
		text := gopodSpellNames(q.Ask)
		brobots := []string{q.To}
		if q.To == "Brobots" {
			brobots = []string{"Brobot 1", "Brobot 2"}
		}
		feedNew()
		for k, b := range brobots {
			take, ok := answer(i+1, esnOf[b], text)
			if !ok {
				continue
			}
			if k == 0 {
				gopodShowChapter(text)
				dq.Before = feedNew()
			}
			dq.Answers = append(dq.Answers, gopodDraftAnswer{Brobot: b, Points: take.Points, Cue: take.Cue, Say: take.Say, Spoken: take.Spoken, LLMSeconds: take.LLMSeconds})
			if q.Topic != "" {
				take.Chapter, take.Topic, take.Ask = q.Chapter, q.Topic, q.Ask
				takes = append(takes, take)
			}
			t.Logf("take %d/%d %.1fs %s %s %v", i+1, len(asks), take.LLMSeconds, b, q.Topic, take.Points)
		}
		gopodShowTopicPictures(gopodBrobot1ESN, text)
		dq.After = feedNew()
		draft.Questions = append(draft.Questions, dq)
	}
	score, detail := gopodCoverageScoreText()
	elapsed := time.Since(started)
	// The ending as it goes live after "Thanks Brobots": the score, Record
	// Session's keep-or-delete (answered keep here), the reporters' closing.
	rs := gopodBrobotsSettings.RecordSession
	draft.Ending = []gopodDraftLine{
		{Who: gopodBrobotsSettings.ValuePoints.ScoreRole, Text: score},
		{Who: rs.AskBrobot, Text: rs.AskText},
		{Who: "Operator to " + rs.AnswerTo, Text: "Keep the session recording."},
		{Who: rs.ConfirmBrobot, Text: rs.KeptText},
	}
	for _, l := range rv.Closing {
		draft.Ending = append(draft.Ending, gopodDraftLine{Who: label(l.Host), Text: l.Text})
	}
	draft.Score, draft.Seconds = score, elapsed.Seconds()
	gopodHarvestWriteJSON(t, filepath.Join(out, "draft_performance.json"), draft)
	summary := map[string]interface{}{
		"run": run, "questions": len(qfile.Questions), "takes": len(takes),
		"seconds": elapsed.Seconds(), "score": score, "coverage": detail, "model": model,
	}
	gopodHarvestWriteJSON(t, filepath.Join(out, "takes.json"), map[string]interface{}{"summary": summary, "takes": takes})
	if sheet := os.Getenv("GOPOD_HARVEST_SHEET"); sheet != "" {
		if err := gopodTakeSheetAdd(sheet, run, takes); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("HARVEST_DONE %s in %s - %s", run, elapsed.Round(time.Second), score)
}

func gopodHarvestRevealed() map[string]int {
	gopodCoverage.Lock()
	defer gopodCoverage.Unlock()
	out := map[string]int{}
	for k, v := range gopodCoverage.revealed {
		out[k] = v
	}
	return out
}

func gopodHarvestWriteJSON(t *testing.T, path string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

// 2026-10-07 (operator: "decouple assess recouple - best questions & best
// responses ... use leading questions"): every harvest question gets its own
// value point - the next one in chapter order - and no other topic's.
func TestHarvestQuestionsGetTheirOwnPoint(t *testing.T) {
	// As at a Wire-Pod start: no Bingo test played, an empty chat memory.
	t.Setenv("GOPOD_SESSION_DIR", t.TempDir())
	played := gopodBingoPlayed.Load()
	gopodBingoPlayed.Store(false)
	defer gopodBingoPlayed.Store(played)
	gopodCoverage.Lock()
	saved := gopodCoverage.revealed
	gopodCoverage.revealed = map[string]int{}
	gopodCoverage.Unlock()
	defer func() { gopodCoverage.Lock(); gopodCoverage.revealed = saved; gopodCoverage.Unlock() }()
	var qfile struct {
		Questions []gopodHarvestQuestion `json:"questions"`
	}
	data, err := os.ReadFile(filepath.Join(gopodChipperRoot(), "gopod_probes/interview_harvest/harvest_questions.json"))
	if err != nil || json.Unmarshal(data, &qfile) != nil {
		t.Fatalf("questions: %v", err)
	}
	_, counts := gopodPackPointCounts()
	var want []string
	for _, c := range gopodBrobotsSettings.InterviewChapters.Chapters {
		for _, topic := range c.Topics {
			for p := 1; p <= counts[topic]; p++ {
				want = append(want, fmt.Sprintf("%s %d/%d", topic, p, counts[topic]))
			}
		}
	}
	if len(qfile.Questions) != len(want) {
		t.Fatalf("%d questions for %d value points", len(qfile.Questions), len(want))
	}
	esnOf := map[string]string{"Brobot 1": gopodBrobot1ESN, "Brobot 2": gopodBrobot2ESN}
	for i, q := range qfile.Questions {
		text := gopodSpellNames(q.Ask)
		gopodQuestionFactsFor(esnOf[q.To], text)
		got := gopodGivenPoints(esnOf[q.To])
		if len(got) != 1 || got[0] != want[i] {
			var loaded []string
			for _, p := range loadScenarioPacketsForMessage(text) {
				loaded = append(loaded, p.Name)
			}
			t.Errorf("question %d %q: given %v, want [%s] (loads %v)", i+1, q.Ask, got, want[i], loaded)
		}
	}
}

// ## EOF

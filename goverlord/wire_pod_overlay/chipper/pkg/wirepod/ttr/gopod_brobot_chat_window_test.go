package wirepod_ttr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeBrobotChatSession(t *testing.T, exchanges []gopodSessionExchange) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("GOPOD_SESSION_DIR", dir)
	var lines []string
	for _, ex := range exchanges {
		line, err := json.Marshal(ex)
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, string(line))
	}
	if err := os.WriteFile(filepath.Join(dir, "current_session.jsonl"), []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func brobotChatText(t *testing.T) string {
	t.Helper()
	var parts []string
	for _, m := range gopodSharedBrobotChat() {
		parts = append(parts, m.Content)
	}
	return strings.Join(parts, "\n")
}

// The shared Brobots chat leaves out exchanges older than 120 minutes and
// anything from before the current Wire-Pod start.
func TestSharedBrobotChatWindowDropsOldAndPreStartExchanges(t *testing.T) {
	savedStart := gopodSharedBrobotChatStart
	defer func() { gopodSharedBrobotChatStart = savedStart }()
	now := time.Now()
	writeBrobotChatSession(t, []gopodSessionExchange{
		{Timestamp: now.Add(-3 * time.Hour), ESN: "0dd1b9e9", UserMessage: "three hours ago", RawResponse: "old answer"},
		{Timestamp: now.Add(-30 * time.Minute), ESN: "0dd1b9e9", UserMessage: "thirty minutes ago", RawResponse: "middle answer"},
		{Timestamp: now.Add(-1 * time.Minute), ESN: "0dd1d8bf", UserMessage: "one minute ago", RawResponse: "new answer"},
	})

	// Started long ago: only the 120-minute limit applies.
	gopodSharedBrobotChatStart = now.Add(-5 * time.Hour)
	chat := brobotChatText(t)
	if strings.Contains(chat, "three hours ago") || strings.Contains(chat, "old answer") {
		t.Fatalf("exchange older than 120 minutes was kept:\n%s", chat)
	}
	if !strings.Contains(chat, "thirty minutes ago") || !strings.Contains(chat, "one minute ago") {
		t.Fatalf("recent exchanges missing:\n%s", chat)
	}

	// Started 10 minutes ago: the 30-minute-old exchange is from before the start.
	gopodSharedBrobotChatStart = now.Add(-10 * time.Minute)
	chat = brobotChatText(t)
	if strings.Contains(chat, "thirty minutes ago") {
		t.Fatalf("exchange from before the service start was kept:\n%s", chat)
	}
	if !strings.Contains(chat, "one minute ago") || !strings.Contains(chat, "new answer") {
		t.Fatalf("exchange after the service start missing:\n%s", chat)
	}
}

// A line that is only a stage direction is skipped; the lines after it are
// still spoken.
func TestNormalizeGOPODRobotSpeechSkipsStageDirectionOnlyLine(t *testing.T) {
	input := "happy... It was an awesome time, you know?\n*giggles and winks*\nAlright folks, that's the gist. Go team GOPOD!"
	got, _ := normalizeGOPODRobotSpeech(input)
	if !strings.Contains(got, "that's the gist") || !strings.Contains(got, "Go team GOPOD") {
		t.Fatalf("reply was cut off at the stage-direction-only line: %q", got)
	}
	if strings.Contains(strings.ToLower(got), "giggles") {
		t.Fatalf("stage direction was kept: %q", got)
	}
}

// Performed lines from the chat feed: a Bingo run becomes one snapshot note (the
// song is scripted), placed in time order; other lines (go search) come back as
// they were performed.
func TestSharedBrobotChatSnapshotsBingoRuns(t *testing.T) {
	savedStart := gopodSharedBrobotChatStart
	defer func() { gopodSharedBrobotChatStart = savedStart }()
	now := time.Now()
	gopodSharedBrobotChatStart = now.Add(-time.Hour)
	writeBrobotChatSession(t, []gopodSessionExchange{
		{Timestamp: now.Add(-10 * time.Minute), ESN: "0dd1b9e9", UserMessage: "ready for a test run", RawResponse: "Ready!"},
		{Timestamp: now.Add(-1 * time.Minute), ESN: "0dd1b9e9", UserMessage: "so was that fun", RawResponse: "It was a blast."},
	})
	gopodChatFeed.Lock()
	savedEvents := gopodChatFeed.events
	gopodChatFeed.events = []gopodChatFeedEvent{
		{ID: 1, Who: "Brobot 1", Text: "go search: GOPOD is a stage", TimeMS: now.Add(-6 * time.Minute).UnixMilli()},
		{ID: 2, Who: "Operator", Text: "Bingo Test Live", TimeMS: now.Add(-5 * time.Minute).UnixMilli()},
		{ID: 3, Who: "Brobot 1", Text: "Bingo ball captured. Number B-1!", TimeMS: now.Add(-5*time.Minute + 10*time.Second).UnixMilli()},
		{ID: 4, Who: "Brobot 2", Text: "Wait!", TimeMS: now.Add(-5*time.Minute + 20*time.Second).UnixMilli()},
	}
	gopodChatFeed.Unlock()
	defer func() { gopodChatFeed.Lock(); gopodChatFeed.events = savedEvents; gopodChatFeed.Unlock() }()

	chat := gopodSharedBrobotChat()
	var got []string
	for _, m := range chat {
		got = append(got, m.Role+": "+m.Content)
	}
	all := strings.Join(got, "\n")
	if len(chat) != 6 {
		t.Fatalf("want question, answer, go search note, Bingo snapshot, question, answer; got %d:\n%s", len(chat), all)
	}
	if !strings.Contains(chat[2].Content, "go search: GOPOD is a stage") || strings.Contains(chat[2].Content, "B-1") {
		t.Fatalf("go search note wrong: %q", chat[2].Content)
	}
	if !strings.HasPrefix(chat[3].Content, gopodBingoSnapshotNote) || strings.Contains(all, "Bingo ball captured") || strings.Contains(all, "Wait!") {
		t.Fatalf("Bingo run not snapshotted:\n%s", all)
	}
	if strings.Contains(all, gopodForgottenNote) {
		t.Fatalf("forgotten note without anything forgotten:\n%s", all)
	}
}

// Past 13 questions the oldest are left out and the Brobots are told they don't
// remember them; Bingo runs from before them are still remembered, as one run.
func TestSharedBrobotChatKeeps13QuestionsAndTheBingoSnapshot(t *testing.T) {
	savedStart := gopodSharedBrobotChatStart
	defer func() { gopodSharedBrobotChatStart = savedStart }()
	now := time.Now()
	gopodSharedBrobotChatStart = now.Add(-2 * time.Hour)
	var exchanges []gopodSessionExchange
	for i := 0; i < 16; i++ {
		exchanges = append(exchanges, gopodSessionExchange{Timestamp: now.Add(time.Duration(i-60) * time.Minute), ESN: "0dd1b9e9", UserMessage: fmt.Sprintf("question %02d", i), RawResponse: "ok"})
	}
	writeBrobotChatSession(t, exchanges)
	gopodChatFeed.Lock()
	savedEvents := gopodChatFeed.events
	gopodChatFeed.events = []gopodChatFeedEvent{
		{ID: 1, Who: "Operator", Text: "Bingo Test Live", TimeMS: now.Add(-90 * time.Minute).UnixMilli()},
		{ID: 2, Who: "Operator", Text: "Bingo Test Live", TimeMS: now.Add(-70 * time.Minute).UnixMilli()},
	}
	gopodChatFeed.Unlock()
	defer func() { gopodChatFeed.Lock(); gopodChatFeed.events = savedEvents; gopodChatFeed.Unlock() }()

	chat := gopodSharedBrobotChat()
	if len(chat) != 2+13*2 || chat[0].Content != gopodForgottenNote || chat[1].Content != gopodBingoSnapshotNote {
		var got []string
		for _, m := range chat {
			got = append(got, m.Content)
		}
		t.Fatalf("got %d messages:\n%s", len(chat), strings.Join(got, "\n"))
	}
	if !strings.Contains(chat[2].Content, "question 03") {
		t.Fatalf("oldest kept question wrong: %q", chat[2].Content)
	}
}

// The Bingo hosting pitch stays apart from a Bingo test run (operator
// 2026-10-01): with a run in memory the pitch pack is left out; with none, a
// Bingo question gets the "still working on it" note.
func TestBingoPitchPackAndNoRunNote(t *testing.T) {
	gopodBingoPlayed.Store(false)
	t.Cleanup(func() { gopodBingoPlayed.Store(false) })
	savedStart := gopodSharedBrobotChatStart
	defer func() { gopodSharedBrobotChatStart = savedStart }()
	now := time.Now()
	gopodSharedBrobotChatStart = now.Add(-time.Hour)
	writeBrobotChatSession(t, nil)
	gopodChatFeed.Lock()
	savedEvents := gopodChatFeed.events
	gopodChatFeed.events = nil
	gopodChatFeed.Unlock()
	defer func() { gopodChatFeed.Lock(); gopodChatFeed.events = savedEvents; gopodChatFeed.Unlock() }()

	prompt := gopodQuestionFacts("did you have fun with bingo")
	if !strings.Contains(prompt, gopodBrobotsSettings.Bingo.NoRunNote) {
		t.Fatal("no run in memory: the no-run note is missing")
	}
	pack, packOK := loadScenarioPacketForMessage("did you have fun with bingo")

	gopodChatFeed.Lock()
	gopodChatFeed.events = []gopodChatFeedEvent{
		{ID: 1, Who: "Operator", Text: "Bingo Test Live", TimeMS: now.Add(-5 * time.Minute).UnixMilli()},
		{ID: 2, Who: "Brobot 1", Text: "Bingo ball captured. Number B-1!", TimeMS: now.Add(-5*time.Minute + 10*time.Second).UnixMilli()},
	}
	gopodChatFeed.Unlock()
	prompt = gopodQuestionFacts("did you have fun with bingo")
	if strings.Contains(prompt, gopodBrobotsSettings.Bingo.NoRunNote) {
		t.Fatal("run in memory: the no-run note was added")
	}
	if packOK && pack.Name == gopodBrobotsSettings.Bingo.PitchPack && strings.Contains(prompt, string(pack.Content)) {
		t.Fatal("run in memory: the Bingo pitch pack was loaded")
	}
}

// A line labelled with a Brobot's name (the model writing for someone else) is
// not chat content: the shared chat skips it. Live 2026-10-04 18:48.
func TestDropNameLabelLines(t *testing.T) {
	raw := "confused... \"Uh, hmm, sleep?\" 🤔 But you know what I say about power naps. 😴\n\n🚀  Brobot 2:  Dude,  \"Sleep\"? Brobots, NO!! 😏 |||1|||"
	got := strings.TrimSpace(gopodDropNameLabelLines(raw))
	if strings.Contains(got, "Brobot 2:") || strings.Contains(got, "Dude") {
		t.Fatalf("labelled line kept: %q", got)
	}
	if !strings.Contains(got, "power naps") {
		t.Fatalf("the answer itself was dropped: %q", got)
	}
	if got := gopodDropNameLabelLines("Brobot 0:  Aaaand that's all she wrote!"); strings.TrimSpace(got) != "" {
		t.Fatalf("a reply written as Brobot 0 should be skipped: %q", got)
	}
	if got := gopodDropNameLabelLines("happy___ Me and Brobot 1: party bros!"); !strings.Contains(got, "party bros") {
		t.Fatalf("a name mid-line is chat content: %q", got)
	}
}

// The Bingo-played flag: a Brobot's Bingo test line sets it, the operator typing
// "Bingo Test Live" doesn't; once set, a Bingo question gets the snapshot and
// the brother's take even with no run in the chat memory.
func TestBingoPlayedFlag(t *testing.T) {
	gopodBingoPlayed.Store(false)
	t.Cleanup(func() { gopodBingoPlayed.Store(false) })
	GopodChatFeedAddTo("Operator", "Brobots", "Bingo Test Live")
	if gopodBingoPlayed.Load() {
		t.Fatal("the operator's own words set the flag")
	}
	GopodChatFeedAdd("Brobot 1", "Bingo Test Begins. Round 1!")
	if !gopodBingoPlayed.Load() {
		t.Fatal("the song's line did not set the flag")
	}
	chat := gopodAddBingoTake(gopodBrobot2ESN, nil)
	if len(chat) != 1 || !strings.Contains(chat[0].Content, gopodBingoSnapshotNote) || !strings.Contains(chat[0].Content, gopodBrobotBingoTakes["Brobot 2"]) {
		t.Fatalf("played but forgotten: want the snapshot and Brobot 2's take, got %+v", chat)
	}
}

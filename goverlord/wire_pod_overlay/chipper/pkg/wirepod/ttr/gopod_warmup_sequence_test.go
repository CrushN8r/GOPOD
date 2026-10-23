// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_warmup_sequence_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// GOLDEN, locked (operator 2026-10-13: "PLEASE FIX GOLDEN AND FUCKING LOCK
// THE PROCESS IN"): the System llm bubble's sequential run check, in the chat
// feed's own order, so a page opened late still shows every step in turn.
func TestWarmupSequenceGolden(t *testing.T) {
	savedFeed, savedState := gopodChatFeedPath, gopodLLMWarmupStatePath
	savedOllama, savedLoaded, savedWarm := gopodWarmupOllama, gopodWarmupLoaded, gopodWarmupWarmRequest
	savedMade, savedWait, savedOpening := gopodWarmupVoicesMade, gopodWarmupVoicesWait, gopodWarmupOpening
	// The built-in opener (the opener song is proved the same: TestOpenerSongPlaysLikeBuiltIn).
	savedSong := gopodOpenerSong
	gopodOpenerSong = func() ([]gopodOpenerStep, error) { return nil, nil }
	defer func() { gopodOpenerSong = savedSong }()
	// 2026-10-17: the lines' voice files count as made (the real voice folder
	// may be empty after a fresh start).
	savedLinesMade := gopodLinesMade
	gopodLinesMade = func([]string) bool { return true }
	defer func() { gopodLinesMade = savedLinesMade }()
	defer func() {
		gopodChatFeedPath, gopodLLMWarmupStatePath = savedFeed, savedState
		gopodWarmupOllama, gopodWarmupLoaded, gopodWarmupWarmRequest = savedOllama, savedLoaded, savedWarm
		gopodWarmupVoicesMade, gopodWarmupVoicesWait, gopodWarmupOpening = savedMade, savedWait, savedOpening
	}()
	// Voices aren't waited out in a test.
	savedMax, savedLag := gopodBrobotsSettings.ReporterVoices.MaxWaitSeconds, gopodBrobotsSettings.ReporterVoices.PageLagSeconds
	defer func() {
		gopodBrobotsSettings.ReporterVoices.MaxWaitSeconds, gopodBrobotsSettings.ReporterVoices.PageLagSeconds = savedMax, savedLag
	}()
	gopodBrobotsSettings.ReporterVoices.MaxWaitSeconds, gopodBrobotsSettings.ReporterVoices.PageLagSeconds = 0.01, 0
	dir := t.TempDir()
	gopodChatFeedPath = filepath.Join(dir, "feed.json")
	gopodLLMWarmupStatePath = filepath.Join(dir, "warmup.json")
	llm := gopodBrobotsSettings.LLM
	role := gopodBrobotsSettings.ReporterVoices.Role
	feed := func() []string {
		gopodChatFeed.Lock()
		defer gopodChatFeed.Unlock()
		var out []string
		for _, e := range gopodChatFeed.events {
			text := e.Text
			if e.Who == role {
				// "warmup_text|<ms>|<text>" -> "warmup_text|<text>"
				if parts := strings.SplitN(text, "|", 3); len(parts) == 3 {
					text = parts[0] + "|" + parts[2]
				} else if strings.Contains(text, ".wav") {
					continue // a voice or a sting
				} else if hm := gopodBrobotsSettings.ReporterVoices.HeardMarker; hm != "" && strings.HasPrefix(text, hm) {
					continue // the page's heard check
				} else if gm := gopodBrobotsSettings.ReporterVoices.GeneratingMarker; gm != "" && strings.HasPrefix(text, gm) {
					continue // the voice template's own marks (generating, lit, off)
				}
			}
			out = append(out, e.Who+": "+text)
		}
		return out
	}
	reset := func() {
		gopodChatFeed.Lock()
		gopodChatFeed.next, gopodChatFeed.events = 0, nil
		gopodChatFeed.Unlock()
	}
	check := func(name string, got, want []string) {
		t.Helper()
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("%s:\ngot\n  %s\nwant\n  %s", name, strings.Join(got, "\n  "), strings.Join(want, "\n  "))
		}
	}
	warming := role + ": warmup_text|" + gopodWarmupLine(llm.WarmupText)
	gopodWarmupOpening = func() { GopodChatFeedAdd("Brobot 3", "opening line") }

	// Wire-Pod start, voices still being made.
	reset()
	gopodWarmupVoicesMade = func() bool { return false }
	gopodWarmupVoicesWait = func() {}
	warmed := make(chan struct{})
	close(warmed)
	gopodWarmupStartSequence(warmed)
	// 2026-10-18 (operator: "I don't see/hear: Voice files generation check while
	// new files generating * I see old message lingering"): while voices are
	// made, message 1 shows with generating_text under it, then on its own.
	check("start, voices being made", feed(), []string{
		role + ": warmup_text|" + gopodWarmupLine(llm.VoicesText) + "\n" + gopodBrobotsSettings.ReporterVoices.GeneratingText,
		role + ": warmup_text|" + gopodWarmupLine(llm.VoicesText),
		warming,
		"Brobot 3: opening line",
		role + ": warmup_ready|" + gopodWarmupReadyLine(llm.ReadyText, llm.WarmupDoneText),
	})
	// Wire-Pod start, voices made already: 2026-10-16 (operator: "(music) Voice
	// files generation check."), message 1 shows at every start - it's the check.
	reset()
	gopodWarmupVoicesMade = func() bool { return true }
	gopodWarmupStartSequence(warmed)
	check("start, voices made", feed(), []string{
		role + ": warmup_text|" + gopodWarmupLine(llm.VoicesText),
		warming,
		"Brobot 3: opening line",
		role + ": warmup_ready|" + gopodWarmupReadyLine(llm.ReadyText, llm.WarmupDoneText),
	})
	// A run (the interview, a replay), warm and cold alike.
	gopodWarmupOllama = func() (string, string, bool) { return "http://ollama", "model", true }
	gopodWarmupWarmRequest = func(string, string) error { return nil }
	for _, loaded := range []bool{true, false} {
		reset()
		gopodWarmupLoaded = func(string, string) bool { return loaded }
		gopodRunWarmup("Run Ready", func() { GopodChatFeedAdd("Brobot 4", "a line") })
		check("run", feed(), []string{warming, "Brobot 4: a line", role + ": warmup_ready|Run Ready"})
	}
}

// GOLDEN, locked (operator 2026-10-13: "LLM warming = disabled/pause on all
// input ... the stop run and response, should have waited in queue for the
// LLM Ready message to complete"): lines typed during a warm-up wait, then
// play in order once LLM Ready has been said.
func TestWarmupQueuesInput(t *testing.T) {
	saved := gopodWarmupReplayLine
	defer func() { gopodWarmupReplayLine = saved; gopodWarmingEnd() }()
	var mu sync.Mutex
	var played []string
	gopodWarmupReplayLine = func(esn, text, how string) error {
		mu.Lock()
		played = append(played, text)
		mu.Unlock()
		return nil
	}
	gopodWarmGateClosed.Store(false)
	// Lines an earlier test left in the queue play out first.
	for i := 0; i < 50; i++ {
		gopodWarming.Lock()
		busy := gopodWarming.drains
		gopodWarming.Unlock()
		if !busy {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	mu.Lock()
	played = nil
	mu.Unlock()
	if gopodWarmupHold("esn", "not warming", "Typed text") {
		t.Fatal("a line with no warm-up going was queued")
	}
	gopodWarmingBegin()
	for _, line := range []string{"stop run", "hello", "thanks"} {
		if !gopodWarmupHold("esn", line, "Typed text") {
			t.Fatalf("%q wasn't queued during the warm-up", line)
		}
	}
	time.Sleep(500 * time.Millisecond)
	mu.Lock()
	if len(played) != 0 {
		t.Fatalf("played during the warm-up: %v", played)
	}
	mu.Unlock()
	gopodWarmingEnd()
	deadline := time.Now().Add(3 * time.Second)
	for {
		mu.Lock()
		n := len(played)
		mu.Unlock()
		if n == 3 || time.Now().After(deadline) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	mu.Lock()
	defer mu.Unlock()
	if strings.Join(played, "|") != "stop run|hello|thanks" {
		t.Errorf("after LLM Ready: %v", played)
	}
}

// GOLDEN, locked (operator 2026-10-13: "snapshot the full \"thinking please
// wait\" process from live (+ arm, nod) * lock that snapshot * use locked
// \"thinking please wait\" process for scripted runs"): one process - the live
// chat (kgsim) and a song's thinking step (/api-sdk/gopod_thinking) both run
// gopodThinkingProcess, and nothing else plays its pieces on its own.
func TestThinkingProcessIsOne(t *testing.T) {
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("GOPOD_OVERLAY_CHIPPER not set (the deploy sets it)")
	}
	read := func(rel string) string {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	kgsim := read("pkg/wirepod/ttr/kgsim.go")
	if !strings.Contains(kgsim, "gopodThinkingProcess(robot, ctx, esn, true,") {
		t.Error("the live chat doesn't run gopodThinkingProcess")
	}
	for _, piece := range []string{"gopodThinkingMoves(", "gopodSayThinkingLine(", "anim_knowledgegraph_searching_01"} {
		if strings.Contains(kgsim, piece) {
			t.Errorf("kgsim.go plays %s on its own, outside gopodThinkingProcess", piece)
		}
	}
	if !strings.Contains(read("pkg/wirepod/sdkapp/server.go"), "wirepod_ttr.GopodThinkingProcess(") {
		t.Error("/api-sdk/gopod_thinking doesn't run the same process")
	}
}

// GOLDEN, locked (operator 2026-10-14: "For all llm warm up runs * Rule: djent
// plays once at the very start and very end * [djent] [message] = audio first,
// pre-recorded ... * [message] [djent] = audio last, after LLM Ready message"):
// the first sound of a warm-up is the default sting, the last is the sting, and
// none plays in between - Wire-Pod's start and a run alike.
func TestWarmupStings(t *testing.T) {
	sting, _ := gopodSystemSting("")
	if sting == "" {
		t.Skip("no default sting file on this machine")
	}
	savedFeed, savedState := gopodChatFeedPath, gopodLLMWarmupStatePath
	savedOllama, savedLoaded, savedWarm := gopodWarmupOllama, gopodWarmupLoaded, gopodWarmupWarmRequest
	savedMade, savedWait, savedOpening := gopodWarmupVoicesMade, gopodWarmupVoicesWait, gopodWarmupOpening
	savedSong := gopodOpenerSong
	gopodOpenerSong = func() ([]gopodOpenerStep, error) { return nil, nil }
	defer func() { gopodOpenerSong = savedSong }()
	savedMax, savedLag := gopodBrobotsSettings.ReporterVoices.MaxWaitSeconds, gopodBrobotsSettings.ReporterVoices.PageLagSeconds
	defer func() {
		gopodChatFeedPath, gopodLLMWarmupStatePath = savedFeed, savedState
		gopodWarmupOllama, gopodWarmupLoaded, gopodWarmupWarmRequest = savedOllama, savedLoaded, savedWarm
		gopodWarmupVoicesMade, gopodWarmupVoicesWait, gopodWarmupOpening = savedMade, savedWait, savedOpening
		gopodBrobotsSettings.ReporterVoices.MaxWaitSeconds, gopodBrobotsSettings.ReporterVoices.PageLagSeconds = savedMax, savedLag
	}()
	gopodBrobotsSettings.ReporterVoices.MaxWaitSeconds, gopodBrobotsSettings.ReporterVoices.PageLagSeconds = 0.01, 0
	dir := t.TempDir()
	gopodChatFeedPath = filepath.Join(dir, "feed.json")
	gopodLLMWarmupStatePath = filepath.Join(dir, "warmup.json")
	role := gopodBrobotsSettings.ReporterVoices.Role
	sounds := func() []string {
		gopodChatFeed.Lock()
		defer gopodChatFeed.Unlock()
		var out []string
		for _, e := range gopodChatFeed.events {
			gm := gopodBrobotsSettings.ReporterVoices.GeneratingMarker
			if e.Who == role && strings.Contains(e.Text, ".wav") && (gm == "" || !strings.HasPrefix(e.Text, gm)) {
				out = append(out, e.Text)
			}
		}
		gopodChatFeed.next, gopodChatFeed.events = 0, nil
		return out
	}
	check := func(name string, got []string) {
		t.Helper()
		if len(got) < 3 {
			t.Skipf("%s: the system voice files aren't made on this machine", name)
		}
		n := 0
		for _, s := range got {
			if s == sting {
				n++
			}
		}
		if got[0] != sting || got[len(got)-1] != sting || n != 2 {
			t.Errorf("%s: want the sting first and last only, got\n  %s", name, strings.Join(got, "\n  "))
		}
	}
	gopodWarmupOpening = func() {}
	gopodWarmupVoicesMade = func() bool { return true }
	gopodWarmupVoicesWait = func() {}
	sounds()
	warmed := make(chan struct{})
	close(warmed)
	gopodWarmupStartSequence(warmed)
	// 2026-10-16 (operator: "remove 1st \"(music)\" & remove its audio play"):
	// Wire-Pod's start has the end sting only.
	if got := sounds(); len(got) >= 2 {
		n := 0
		for _, s := range got {
			if s == sting {
				n++
			}
		}
		if got[len(got)-1] != sting || n != 1 {
			t.Errorf("Wire-Pod start: want the sting last only, got\n  %s", strings.Join(got, "\n  "))
		}
	}
	gopodWarmupOllama = func() (string, string, bool) { return "http://ollama", "model", true }
	gopodWarmupLoaded = func(string, string) bool { return true }
	gopodRunWarmup(gopodBrobotsSettings.LLM.ReadyText, func() { gopodSystemSay(gopodBrobotsSettings.LLM.WarmupDoneText, true) })
	check("a run", sounds())
	// After the warm-up, a system line has its sting again.
	if s, _ := gopodSystemStingNow("hello"); s == "" {
		t.Error("no sting for a system line after the warm-up")
	}
}

// GOLDEN, locked (operator 2026-10-14: "\"LLM Warming Up. Please Wait.\" stays
// until fully spoken * if LLM Ready, then update to LLM Ready message + speech *
// rule applies to all llm warming runs"): a song's warming line, said in the
// background, is in the chat before its ready line, even with no lines between.
func TestSongWarmupReadyAfterWarming(t *testing.T) {
	sayRole := gopodBrobotsSettings.ReporterVoices.SystemSayRole
	warming := gopodWarmupLine(gopodBrobotsSettings.LLM.WarmupText)
	page, _, ok := gopodReporterFile(gopodBrobotsSettings.SignOff.SystemRole, gopodReporterSays(warming))
	if sayRole == "" || !ok {
		t.Skip("no system_say_role, or the warming line's voice isn't made on this machine")
	}
	savedFeed := gopodChatFeedPath
	savedMax, savedLag := gopodBrobotsSettings.ReporterVoices.MaxWaitSeconds, gopodBrobotsSettings.ReporterVoices.PageLagSeconds
	defer func() {
		gopodChatFeedPath = savedFeed
		gopodBrobotsSettings.ReporterVoices.MaxWaitSeconds, gopodBrobotsSettings.ReporterVoices.PageLagSeconds = savedMax, savedLag
		gopodWarmingEnd()
	}()
	gopodBrobotsSettings.ReporterVoices.MaxWaitSeconds, gopodBrobotsSettings.ReporterVoices.PageLagSeconds = 0.01, 0
	gopodChatFeedPath = filepath.Join(t.TempDir(), "feed.json")
	for i := 0; i < 20; i++ {
		gopodChatFeed.Lock()
		gopodChatFeed.next, gopodChatFeed.events = 0, nil
		gopodChatFeed.Unlock()
		GopodSongWarmupText(1, warming)
		GopodChatFeedAdd(sayRole, warming)
		GopodSongWarmupReady(1, gopodBrobotsSettings.LLM.ReadyText)
		gopodChatFeed.Lock()
		voiceAt, readyAt := -1, -1
		for j, e := range gopodChatFeed.events {
			if voiceAt < 0 && e.Text == page {
				voiceAt = j
			}
			if readyAt < 0 && strings.HasPrefix(e.Text, gopodWarmupReadyMark) {
				readyAt = j
			}
		}
		gopodChatFeed.Unlock()
		if voiceAt < 0 || readyAt < 0 || voiceAt > readyAt {
			t.Fatalf("try %d: the warming line's voice (at %d) isn't before the ready line (at %d)", i, voiceAt, readyAt)
		}
	}
}

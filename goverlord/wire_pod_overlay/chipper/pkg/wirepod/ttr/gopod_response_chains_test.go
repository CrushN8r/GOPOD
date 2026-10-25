package wirepod_ttr

import (
	"path/filepath"
	"strings"
	"testing"
)

// 2026-10-25 (operator: "Brobot A: 'What say you, [funny-nickname]?' * Brobot B
// replaces his 'Got it!' with 'Nice try [funny-nickname]'"): the response-chain
// template - single says his Got it; duo A his Got it and hands over; duo B opens
// with the comeback; the notice shows the plan.
func TestResponseChainTemplate(t *testing.T) {
	a, b := gopodBrobot2ESN, gopodBrobot1ESN
	defer gopodAllFirstAnswer.Delete(a)
	defer gopodFollowingUp.Delete(b)

	if got := gopodOpenerPick(a); got != gopodGotItFor(a).Text {
		t.Fatalf("single opener %q", got)
	}
	if p := gopodResponsePlan(a); p != "" {
		t.Fatalf("single plan %q", p)
	}

	gopodAllFirstAnswer.Store(a, true)
	if got := gopodOpenerPick(a); got != gopodGotItFor(a).Text {
		t.Fatalf("duo A opener %q", got)
	}
	if p := gopodResponsePlan(a); p != "Response-chain: Brobot 2, then Brobot 1 - answer 1 of 2" {
		t.Fatalf("duo A plan %q", p)
	}
	if h := gopodHandoverLine("Brobot 2", "Brobot 1"); !strings.HasPrefix(h, "Brobot 1, what say you on this, Mr") {
		t.Fatalf("duo A hand-over %q", h)
	}

	gopodFollowingUp.Store(b, true)
	got := gopodOpenerPick(b)
	if !strings.HasPrefix(got, "Nice try, Mr ") && !strings.HasPrefix(got, "Nice try, Mrs ") || !strings.HasSuffix(got, "!") {
		t.Fatalf("duo B opener %q", got)
	}
	if gopodOpenerText(b) != got {
		t.Fatalf("the notice shows %q, he said %q", gopodOpenerText(b), got)
	}
	if p := gopodResponsePlan(b); p != "Response-chain: Brobot 2, then Brobot 1 - answer 2 of 2" {
		t.Fatalf("duo B plan %q", p)
	}
	if r, ok := gopodDuoRule(b); !ok || !r.NoManager || r.WeatherType != "weather_next" {
		t.Fatalf("duo B rules %+v", r)
	}
}

// 2026-10-25 (operator: "just add the nice try before the notice bubble"): the
// opener he said goes to the page for the top of his bubble; not a notice line.
func TestOpenerAboveTheNotice(t *testing.T) {
	esn := gopodBrobot1ESN
	gopodChainOpenerSaid.Store(esn, "Nice try, Mr fuzzy logic!")
	defer gopodChainOpenerSaid.Delete(esn)
	oldFeed := gopodChatFeedPath
	gopodChatFeedPath = filepath.Join(t.TempDir(), "gopod_chat_feed.json")
	defer func() { gopodChatFeedPath = oldFeed }()
	gopodChatFeed.Lock()
	before := len(gopodChatFeed.events)
	gopodChatFeed.Unlock()
	gopodOpenerShow(esn)
	gopodChatFeed.Lock()
	feed := append([]gopodChatFeedEvent(nil), gopodChatFeed.events[before:]...)
	gopodChatFeed.Unlock()
	if len(feed) != 1 || feed[0].Text != "opener:Nice try, Mr fuzzy logic!" || feed[0].To != "Brobot 1" {
		t.Fatalf("feed %+v", feed)
	}
	for _, p := range gopodBrobotsSettings.BubbleNotices.ValuePoints.Parts {
		if p == "got_it" {
			t.Fatal("the opener is still a notice line")
		}
	}
}

// 2026-10-25 (operator: "post-bingo, brobot 1, \"I Won!\", brobot 2, \"Oh!
// Zero!\""; "No, before the 'Hey' reminder"): after a Bingo run, a Bingo
// question's answer has his closing remark just before the reminder.
func TestBingoClosingBeforeReminder(t *testing.T) {
	saved := gopodBingoPlayed.Load()
	gopodBingoPlayed.Store(true)
	defer gopodBingoPlayed.Store(saved)
	q := "how was bingo?"
	if !gopodIsBingoQuestion(q) {
		t.Skip("not a bingo question here")
	}
	if got := gopodBingoClosing(gopodBrobot1ESN, q); got != "Anyway, we got a winner! 🏆 \"I Won!\"" {
		t.Fatalf("Brobot 1 closing %q", got)
	}
	if got := gopodBingoClosing(gopodBrobot2ESN, q); got != "Oh! Zero!" {
		t.Fatalf("Brobot 2 closing %q", got)
	}
	parts := gopodAnswerChain("brobots")
	ci, ri := -1, -1
	for i, p := range parts {
		if p == "closing_remark" {
			ci = i
		}
		if p == "reminder" {
			ri = i
		}
	}
	li := -1
	for i, p := range parts {
		if p == "llm" {
			li = i
		}
	}
	if ci != li+1 || ri < ci {
		t.Fatalf("closing_remark isn't right after the answer: %v", parts)
	}
	// Joined to his answer, its closing sentence - not a thought of its own.
	if got := gopodChainStitch("closing_remark", gopodBrobot2ESN, q, "Bingo was wild, bro!", false); got != "Bingo was wild, bro! Oh! Zero!" {
		t.Fatalf("stitched %q", got)
	}
}

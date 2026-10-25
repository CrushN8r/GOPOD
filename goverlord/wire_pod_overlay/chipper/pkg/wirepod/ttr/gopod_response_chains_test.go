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

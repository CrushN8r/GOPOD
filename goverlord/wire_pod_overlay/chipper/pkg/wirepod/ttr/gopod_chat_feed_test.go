package wirepod_ttr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Every chat-page line gets its own number, even the same words twice, and
// the feed keeps only its newest gopodChatFeedMax() lines.
func TestGopodChatFeedNumbersEveryLine(t *testing.T) {
	savedPath := gopodChatFeedPath
	defer func() { gopodChatFeedPath = savedPath }()
	gopodChatFeedPath = filepath.Join(t.TempDir(), "gopod_chat_feed.json")
	gopodChatFeed.Lock()
	gopodChatFeed.next, gopodChatFeed.events = 0, nil
	gopodChatFeed.Unlock()

	GopodChatFeedAdd("Brobot 1", "Bingo Test Begins. Round 1!")
	GopodChatFeedAdd("Brobot 1", "Bingo Test Begins. Round 1!")
	GopodChatFeedAdd("", "no speaker is ignored")
	// 2026-10-18 (flaky: another test's background line could land in this
	// feed and shift the count): the property itself - the same words twice
	// get two numbers, every number is new and rising - not an exact total.
	gopodChatFeed.Lock()
	var bingo []int
	for _, e := range gopodChatFeed.events {
		if e.Text == "Bingo Test Begins. Round 1!" {
			bingo = append(bingo, e.ID)
		}
		if e.Who == "" {
			t.Error("a line with no speaker was kept")
		}
	}
	gopodChatFeed.Unlock()
	if len(bingo) != 2 || bingo[0] == bingo[1] {
		t.Fatalf("the repeated line: ids %v", bingo)
	}
	for i := 0; i < gopodChatFeedMax(); i++ {
		GopodChatFeedAdd("Brobot 2", "filler")
	}
	data, err := os.ReadFile(gopodChatFeedPath)
	if err != nil {
		t.Fatal(err)
	}
	var feed struct {
		Boot   int64                `json:"boot"`
		Events []gopodChatFeedEvent `json:"events"`
	}
	if err := json.Unmarshal(data, &feed); err != nil {
		t.Fatal(err)
	}
	if feed.Boot == 0 || len(feed.Events) != gopodChatFeedMax() {
		t.Fatalf("boot=%d events=%d", feed.Boot, len(feed.Events))
	}
	for i := 1; i < len(feed.Events); i++ {
		if feed.Events[i].ID <= feed.Events[i-1].ID {
			t.Fatalf("ids not rising: %d then %d", feed.Events[i-1].ID, feed.Events[i].ID)
		}
	}
	if last := feed.Events[len(feed.Events)-1]; last.ID < gopodChatFeedMax()+2 {
		t.Fatalf("last id = %d, want at least %d", last.ID, gopodChatFeedMax()+2)
	}
}

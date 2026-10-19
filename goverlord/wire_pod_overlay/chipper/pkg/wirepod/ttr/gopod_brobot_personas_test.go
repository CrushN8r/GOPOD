package wirepod_ttr

import (
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
)

// The Bingo snapshot in the shared chat gets the asking brother's own take; Brobot 1's
// take never names Brobot 2 (see gopodBrobotPersonaLines).
func TestBingoTakeIsAddedForTheAskingBrother(t *testing.T) {
	bingo := gopodBingoSnapshotNote
	chat := func() []openai.ChatCompletionMessage {
		return []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: bingo}}
	}
	one := gopodAddBingoTake(gopodBrobot1ESN, chat())[0].Content
	if !strings.Contains(one, "you were the Bingo tester") || strings.Contains(one, "Brobot 2") {
		t.Fatalf("Brobot 1 take wrong: %q", one)
	}
	two := gopodAddBingoTake(gopodBrobot2ESN, chat())[0].Content
	if !strings.Contains(two, "cluelessly messed it up") {
		t.Fatalf("Brobot 2 take missing: %q", two)
	}
	other := gopodSharedFeedNote([]gopodChatFeedEvent{{Who: "Brobot 1", Text: "go search result about Bingo"}})
	plain := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: other}}
	if got := gopodAddBingoTake(gopodBrobot1ESN, plain)[0].Content; got != other {
		t.Fatalf("take added without a Bingo run: %q", got)
	}
}

// ## EOF

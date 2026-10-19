package wirepod_ttr

import (
	"strings"
	"testing"
	"time"
)

func TestGOPODNameSpellingWritesNamesAsTheRepoDoes(t *testing.T) {
	tests := map[string]string{
		"explain how go pod works under the hood":                "explain how GOPOD works under the hood",
		"what's go-pod and gopod":                                "what's GOPOD and GOPOD",
		"who is governed lord and the go mads and to go minions": "who is Goverlord and the Gomads and to Gominions",
		"tell me about govern lord":                              "tell me about Goverlord",
		"go minion, gomad, gominions":                            "Gominion, Gomad, Gominions",
		"GOPOD and Goverlord stay as they are":                   "GOPOD and Goverlord stay as they are",
		// Operator decisions 2026-09-30: "go mad" stays as heard (a real phrase).
		"explain go part and go put under the hood":  "explain GOPOD and GOPOD under the hood",
		"so how discover lord and were lord operate": "so how Goverlord and Goverlord operate",
		"discover lowered":                           "Goverlord",
		"use go mad or go marion's":                  "use go mad or Gominions",
		// Whole words only.
		"gopodcast and ago pod": "gopodcast and ago pod",
	}
	for in, want := range tests {
		if got := gopodSpellNames(in); got != want {
			t.Fatalf("gopodSpellNames(%q) = %q, want %q", in, got, want)
		}
	}
}

// Robot speech is unchanged: once spelled GOPOD, the pronunciation registry
// still says it Gowp-awd.
func TestGOPODNameSpellingKeepsRobotPronunciation(t *testing.T) {
	got := gopodApplyPronunciations(gopodSpellNames("what is go pod and who are the gomads"))
	if !strings.Contains(got, "Gowp-awd") || !strings.Contains(got, "Gohm-adds") {
		t.Fatalf("robot say lost the pronunciation: %q", got)
	}
}

// The shared chat memory hands the LLM the spelled names, even for exchanges
// saved before this fix.
func TestSharedBrobotChatSpellsNames(t *testing.T) {
	savedStart := gopodSharedBrobotChatStart
	defer func() { gopodSharedBrobotChatStart = savedStart }()
	gopodSharedBrobotChatStart = time.Now().Add(-time.Hour)
	writeBrobotChatSession(t, []gopodSessionExchange{
		{Timestamp: time.Now().Add(-time.Minute), ESN: "0dd1b9e9", UserMessage: "explain go pod under the hood", RawResponse: "Go minions do the work."},
	})
	chat := brobotChatText(t)
	if !strings.Contains(chat, "explain GOPOD under the hood") || !strings.Contains(chat, "Gominions do the work") {
		t.Fatalf("memory not spelled: %q", chat)
	}
}

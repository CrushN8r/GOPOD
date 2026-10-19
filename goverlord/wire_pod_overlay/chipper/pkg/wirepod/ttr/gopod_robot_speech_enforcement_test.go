package wirepod_ttr

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestValidateGOPODRobotSpeechLines(t *testing.T) {
	valid := []string{
		"happy... I am ready.",
		"thinking... Robots sense, process, and act.",
	}
	for _, input := range valid {
		result := validateGOPODRobotSpeech(input)
		if !result.Valid {
			t.Fatalf("validateGOPODRobotSpeech(%q) valid = false, failure = %#v", input, result)
		}
	}

	failures := map[string]string{
		"Bro! I am ready.":                     "plain prose",
		"happpy... I am ready.":                "actionParameter typo",
		"happy ... I am ready.":                "bad spacing around dots",
		"angry  ... I am ready.":               "bad spacing around dots",
		"happy...I am ready.":                  "missing space after dots",
		"Note: I am ready.":                    "note/label/header",
		"*waves arms* I am ready.":             "stage direction",
		"[happy] I am ready.":                  "command/parser syntax",
		"{actionParameter: happy} I am ready.": "command/parser syntax",
	}
	for input, wantType := range failures {
		result := validateGOPODRobotSpeech(input)
		if result.Valid {
			t.Fatalf("validateGOPODRobotSpeech(%q) valid = true, want false", input)
		}
		if result.FailureType != wantType {
			t.Fatalf("validateGOPODRobotSpeech(%q) failure type = %q, want %q", input, result.FailureType, wantType)
		}
		if result.FirstFailingLine == "" {
			t.Fatalf("validateGOPODRobotSpeech(%q) missing first failing line", input)
		}
	}
}

func TestNormalizeGOPODRobotSpeechDeterministically(t *testing.T) {
	tests := map[string]string{
		"happpy... I am ready.":     "happy... I am ready.",
		"angry  ...  I am ready.":   "angry... I am ready.",
		"happy ... I am ready.":     "happy... I am ready.",
		"happy...I am ready.":       "happy... I am ready.",
		"excited... I am ready.":    "veryHappy... excited... I am ready.", // an alias is said (2026-10-02)
		"curious... I am checking.": "thinking... curious... I am checking.",
		"Bro! I am ready.":          "thinking... Bro! I am ready.",
		"[happy] I am ready.":       "happy... I am ready.",
	}
	for input, want := range tests {
		got, result := normalizeGOPODRobotSpeech(input)
		if !result.Valid {
			t.Fatalf("normalizeGOPODRobotSpeech(%q) invalid result %#v output %q", input, result, got)
		}
		if got != want {
			t.Fatalf("normalizeGOPODRobotSpeech(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestActionParameterFailuresListLoadsAndFeedsNormalizer(t *testing.T) {
	path := filepath.Join(gopodChipperRoot(), "gopod_probes", "actionParameter-failures-list.txt")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%q) error = %v", path, err)
	}
	if info.Size() == 0 {
		t.Fatalf("%s is empty", path)
	}
	entries, err := loadActionParameterFailureEntries(path)
	if err != nil {
		t.Fatalf("loadActionParameterFailureEntries error = %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("loaded no actionParameter failure entries")
	}
	if _, ok := entries["happpy..."]; !ok {
		t.Fatal("known failure happpy... was not loaded")
	}
	got, result := normalizeGOPODRobotSpeech("hapy... I am ready.")
	if !result.Valid {
		t.Fatalf("normalizer did not use failures list: %#v output %q", result, got)
	}
	if got != "happy... I am ready." {
		t.Fatalf("normalized hapy alias = %q", got)
	}
}

func TestGOPODFallbackRunsAfterRepairFailureAndPassesRegex(t *testing.T) {
	repairer := func(ctx context.Context, req openai.ChatCompletionRequest) (string, error) {
		return "Note: still malformed", nil
	}
	result := buildGOPODStreamResultWithRepair(context.Background(), "!!!", "repair-model", repairer)
	if result.FilteredPacket != gopodRobotSpeechFallbackLine {
		t.Fatalf("fallback filtered = %q, want %q", result.FilteredPacket, gopodRobotSpeechFallbackLine)
	}
	if validation := validateGOPODRobotSpeech(result.FilteredPacket); !validation.Valid {
		t.Fatalf("fallback failed validator: %#v", validation)
	}
}

func TestGOPODContactBeatsPassRegexAndSelectByStage(t *testing.T) {
	stages := []gopodContactBeatStage{
		gopodContactBeatMemoryLoading,
		gopodContactBeatScenarioLoading,
		gopodContactBeatFormatting,
		gopodContactBeatRepair,
		gopodContactBeatFallback,
	}
	for _, beat := range gopodContactBeatTemplates {
		if strings.Contains(strings.ToLower(beat.Line), "scratchpad") || strings.Contains(strings.ToLower(beat.Line), "reasoning") {
			t.Fatalf("contact beat exposes hidden reasoning terms: %q", beat.Line)
		}
		if validation := validateGOPODRobotSpeech(beat.Line); !validation.Valid {
			t.Fatalf("contact beat failed validator: %#v", validation)
		}
	}
	for _, stage := range stages {
		beat, ok := selectGOPODContactBeat(stage, "")
		if !ok {
			t.Fatalf("no contact beat for stage %q", stage)
		}
		if validation := validateGOPODRobotSpeech(beat); !validation.Valid {
			t.Fatalf("selected beat failed validator: %#v", validation)
		}
	}
	if beat, ok := selectGOPODContactBeat(gopodContactBeatScenarioLoading, "bingo"); !ok || !strings.Contains(beat, "Bingo context") {
		t.Fatalf("bingo scenario beat = %q ok=%t", beat, ok)
	}
}

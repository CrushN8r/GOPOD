package wirepod_ttr

import (
	"strings"
	"testing"
)

const validAnimationVocabJSONForTest = `{
  "description": "test fixture",
  "tokens": [
    {"name": "happy", "vector_clip": "anim_happy_01", "verified": true},
    {"name": "thinking", "vector_clip": "anim_thinking_01", "verified": true},
    {"name": "GOPOD_ANIM_TODO", "vector_clip": "anim_todo_01", "verified": false}
  ],
  "aliases": {
    "glad": "happy",
    "hmm": "thinking"
  }
}`

func TestParseAnimationVocabValidInput(t *testing.T) {
	vocab, err := parseAnimationVocab([]byte(validAnimationVocabJSONForTest))
	if err != nil {
		t.Fatalf("parseAnimationVocab error = %v", err)
	}
	if len(vocab.Tokens) != 3 {
		t.Fatalf("len(Tokens) = %d, want 3", len(vocab.Tokens))
	}
	if got, ok := vocab.Aliases["glad"]; !ok || got != "happy" {
		t.Fatalf("Aliases[glad] = %q, %v, want happy, true", got, ok)
	}
}

func TestParseAnimationVocabRejectsMalformedJSON(t *testing.T) {
	_, err := parseAnimationVocab([]byte(`{not valid json`))
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestParseAnimationVocabRejectsDuplicateTokenName(t *testing.T) {
	input := `{
		"tokens": [
			{"name": "happy", "vector_clip": "anim_a", "verified": true},
			{"name": "happy", "vector_clip": "anim_b", "verified": true}
		],
		"aliases": {}
	}`
	_, err := parseAnimationVocab([]byte(input))
	if err == nil {
		t.Fatal("expected error for duplicate token name, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate token name") {
		t.Fatalf("error = %v, want mention of duplicate token name", err)
	}
}

func TestParseAnimationVocabRejectsDuplicateAliasKeyAfterNormalization(t *testing.T) {
	input := `{
		"tokens": [{"name": "happy", "vector_clip": "anim_a", "verified": true}],
		"aliases": {"Glad": "happy", "glad": "happy"}
	}`
	_, err := parseAnimationVocab([]byte(input))
	if err == nil {
		t.Fatal("expected error for duplicate alias key after normalization, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate alias key") {
		t.Fatalf("error = %v, want mention of duplicate alias key", err)
	}
}

func TestParseAnimationVocabRejectsAliasPointingAtUnknownToken(t *testing.T) {
	input := `{
		"tokens": [{"name": "happy", "vector_clip": "anim_a", "verified": true}],
		"aliases": {"glad": "notARealToken"}
	}`
	_, err := parseAnimationVocab([]byte(input))
	if err == nil {
		t.Fatal("expected error for alias pointing at unknown token, got nil")
	}
	if !strings.Contains(err.Error(), "unknown token") {
		t.Fatalf("error = %v, want mention of unknown token", err)
	}
}

// TestAnimationAliasNervousResolvesToConfused pins a cross-table conflict
// found while consolidating the vocabulary: knownGOPODActionAlias used to
// map "nervous" -> "confused" while resolveAnimationParameter/
// resolveGOPODLikelyAnimationAlias mapped it (via the cheeky/suspicious/
// sneaky/awkward/nervous/shifty cluster) to the now-retired "dartingEyes".
// Repointing that whole cluster to "confused" makes every path agree. This
// test exists so a future hand-edit of animation_vocab.json can't silently
// reintroduce the disagreement.
func TestAnimationAliasNervousResolvesToConfused(t *testing.T) {
	got, ok := animationAliases["nervous"]
	if !ok {
		t.Fatal(`animationAliases["nervous"] not found`)
	}
	if got != "confused" {
		t.Fatalf(`animationAliases["nervous"] = %q, want "confused"`, got)
	}
}

// TestResolveAnimationParameterAfterConsolidation exercises
// resolveAnimationParameter (kgsim_cmds_animation_normalizer.go) directly.
// That file's own embedded "Test..." functions are NOT named *_test.go, so
// `go test` never discovers them (a pre-existing build-hygiene gap, flagged
// separately, not fixed here) - this gives the post-consolidation behavior
// real automated coverage instead of relying on manual inspection.
func TestResolveAnimationParameterAfterConsolidation(t *testing.T) {
	tests := map[string]string{
		"happy":                    "happy",     // exact animationMap token
		"HAPPY":                    "happy",     // case-insensitive exact match
		"excited":                  "veryHappy", // alias
		"awkward":                  "confused",  // repointed alias (was dartingEyes)
		"nervous":                  "confused",  // repointed alias (was dartingEyes)
		"totally_unresolvable_xyz": "thinking",  // generic fallback
	}
	for input, want := range tests {
		if got := resolveAnimationParameter(input); got != want {
			t.Errorf("resolveAnimationParameter(%q) = %q, want %q", input, got, want)
		}
	}
}

// TestNormalizeAnimationSyntaxAfterConsolidation exercises
// NormalizeAnimationSyntax (kgsim_cmds_animation_normalizer.go) directly,
// for the same reason as the test above.
func TestNormalizeAnimationSyntaxAfterConsolidation(t *testing.T) {
	tests := map[string]string{
		"{{happy}} Hi!": "{{playAnimationWI||happy}} Hi!",
		"[awkward] Hi!": "{{playAnimationWI||confused}} Hi!",
	}
	for input, want := range tests {
		if got := NormalizeAnimationSyntax(input); got != want {
			t.Errorf("NormalizeAnimationSyntax(%q) = %q, want %q", input, got, want)
		}
	}
}

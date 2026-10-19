package wirepod_ttr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AnimationToken is one entry in animation_vocab.json's "tokens" array.
type AnimationToken struct {
	Name       string `json:"name"`
	VectorClip string `json:"vector_clip"`
	Verified   bool   `json:"verified"`
}

// AnimationVocab is the parsed shape of animation_vocab.json, the single
// source of truth for the GOPOD animation vocabulary. See that file's
// "description" field.
type AnimationVocab struct {
	Description string            `json:"description"`
	Tokens      []AnimationToken  `json:"tokens"`
	Aliases     map[string]string `json:"aliases"`
	// GOPOD 2026-10-04: misspelled cue words per token (operator: "a short,
	// focused list of these types of misspellings for each playAnimationWI") -
	// read as that token, a cue only, never said.
	Misspellings map[string][]string `json:"misspellings"`
}

const animationVocabPathEnvVar = "GOPOD_ANIMATION_VOCAB_PATH"

// resolveAnimationVocabPath locates animation_vocab.json in GOPOD's chipper
// folder (gopodChipperRoot - the overlay, 2026-10-02), overridable via
// GOPOD_ANIMATION_VOCAB_PATH for tests/alternate deployments.
func resolveAnimationVocabPath() string {
	if override := strings.TrimSpace(os.Getenv(animationVocabPathEnvVar)); override != "" {
		return override
	}
	return filepath.Join(gopodChipperRoot(), "animation_vocab.json")
}

// parseAnimationVocab is pure (no file I/O, never panics) so it can be unit
// tested directly against malformed/adversarial input.
func parseAnimationVocab(data []byte) (AnimationVocab, error) {
	var vocab AnimationVocab
	if err := json.Unmarshal(data, &vocab); err != nil {
		return AnimationVocab{}, fmt.Errorf("animation_vocab: invalid JSON: %w", err)
	}
	if len(vocab.Tokens) == 0 {
		return AnimationVocab{}, fmt.Errorf("animation_vocab: no tokens defined")
	}

	seenTokenNames := make(map[string]bool, len(vocab.Tokens))
	knownTokenNames := make(map[string]bool, len(vocab.Tokens))
	for _, tok := range vocab.Tokens {
		name := strings.TrimSpace(tok.Name)
		if name == "" {
			return AnimationVocab{}, fmt.Errorf("animation_vocab: token with empty name")
		}
		if seenTokenNames[name] {
			return AnimationVocab{}, fmt.Errorf("animation_vocab: duplicate token name %q", name)
		}
		seenTokenNames[name] = true
		knownTokenNames[name] = true
		if strings.TrimSpace(tok.VectorClip) == "" {
			return AnimationVocab{}, fmt.Errorf("animation_vocab: token %q has empty vector_clip", name)
		}
	}

	seenNormalizedAliasKeys := make(map[string]bool, len(vocab.Aliases))
	for rawKey, target := range vocab.Aliases {
		normKey := normalizeAnimationToken(rawKey)
		if normKey == "" {
			return AnimationVocab{}, fmt.Errorf("animation_vocab: alias key %q normalizes to empty string", rawKey)
		}
		if seenNormalizedAliasKeys[normKey] {
			return AnimationVocab{}, fmt.Errorf("animation_vocab: duplicate alias key %q after normalization", normKey)
		}
		seenNormalizedAliasKeys[normKey] = true
		if !knownTokenNames[target] {
			return AnimationVocab{}, fmt.Errorf("animation_vocab: alias %q points at unknown token %q", rawKey, target)
		}
	}

	for target, words := range vocab.Misspellings {
		if !knownTokenNames[target] {
			return AnimationVocab{}, fmt.Errorf("animation_vocab: misspellings for unknown token %q", target)
		}
		for _, w := range words {
			if normalizeAnimationToken(w) == "" {
				return AnimationVocab{}, fmt.Errorf("animation_vocab: empty misspelling for %q", target)
			}
		}
	}

	return vocab, nil
}

// mustLoadAnimationVocab reads and parses animation_vocab.json, panicking on
// any failure. This is the one place allowed to fail loudly: silently
// falling back to a baked-in list would recreate the exact duplication this
// file exists to eliminate.
func mustLoadAnimationVocab() AnimationVocab {
	path := resolveAnimationVocabPath()
	data, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("animation_vocab: failed to read %s: %v", path, err))
	}
	vocab, err := parseAnimationVocab(data)
	if err != nil {
		panic(fmt.Sprintf("animation_vocab: failed to load %s: %v", path, err))
	}
	return vocab
}

func buildAnimationMap(vocab AnimationVocab) [][2]string {
	m := make([][2]string, 0, len(vocab.Tokens))
	for _, tok := range vocab.Tokens {
		m = append(m, [2]string{tok.Name, tok.VectorClip})
	}
	return m
}

func buildAnimationAliases(vocab AnimationVocab) map[string]string {
	aliases := make(map[string]string, len(vocab.Aliases))
	for rawKey, target := range vocab.Aliases {
		aliases[normalizeAnimationToken(rawKey)] = target
	}
	// The misspellings resolve the same way, wherever a cue word is read.
	for target, words := range vocab.Misspellings {
		for _, w := range words {
			if key := normalizeAnimationToken(w); aliases[key] == "" {
				aliases[key] = target
			}
		}
	}
	return aliases
}

// verifiedTokenNames returns token names in file order, excluding any
// unverified/dormant tokens (e.g. GOPOD_ANIM_TODO) - these are still
// dispatchable directly but are not offered to the LLM as a choice.
func verifiedTokenNames(vocab AnimationVocab) []string {
	names := make([]string, 0, len(vocab.Tokens))
	for _, tok := range vocab.Tokens {
		if tok.Verified {
			names = append(names, tok.Name)
		}
	}
	return names
}

func buildAnimationParamChoices(vocab AnimationVocab) string {
	return strings.Join(verifiedTokenNames(vocab), ", ")
}

// buildAnimationRegexAlternation returns just the parenthesized-alternation
// body (e.g. "happy|veryHappy|...|love"), shared by both the compiled regex
// (escaped dots) and the prose prompt-guard text (unescaped dots) so the
// token list itself is assembled in exactly one place.
func buildAnimationRegexAlternation(vocab AnimationVocab) string {
	return strings.Join(verifiedTokenNames(vocab), "|")
}

var loadedAnimationVocab = mustLoadAnimationVocab()

var animationMap = buildAnimationMap(loadedAnimationVocab)

var animationAliases = buildAnimationAliases(loadedAnimationVocab)

// animationMisspellings: normalized misspelled cue word -> its token.
var animationMisspellings = buildAnimationMisspellings(loadedAnimationVocab)

func buildAnimationMisspellings(vocab AnimationVocab) map[string]string {
	out := map[string]string{}
	for target, words := range vocab.Misspellings {
		for _, w := range words {
			out[normalizeAnimationToken(w)] = target
		}
	}
	return out
}

var animationParamChoices = buildAnimationParamChoices(loadedAnimationVocab)

var gopodAnimationRegexAlternation = buildAnimationRegexAlternation(loadedAnimationVocab)

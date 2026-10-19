package wirepod_ttr

import (
	"strings"
	"testing"
)

func NormalizeAnimationSyntax(input string) string {
	var b strings.Builder
	for i := 0; i < len(input); {
		switch {
		case strings.HasPrefix(input[i:], "{{"):
			end := strings.Index(input[i+2:], "}}")
			if end == -1 {
				b.WriteString(input[i:])
				i = len(input)
				continue
			}
			rawCommand := input[i+2 : i+2+end]
			if normalized, ok := normalizeBracedAnimationCommand(rawCommand); ok {
				b.WriteString(normalized)
			} else {
				b.WriteString(input[i : i+2+end+2])
			}
			i += 2 + end + 2
		case input[i] == '[':
			end := strings.IndexByte(input[i+1:], ']')
			if end == -1 {
				b.WriteByte(input[i])
				i++
				continue
			}
			rawEmotion := input[i+1 : i+1+end]
			if isLikelyAnimationToken(rawEmotion) {
				b.WriteString(formatAnimationCommand(resolveAnimationParameter(rawEmotion)))
				i += 1 + end + 1
			} else {
				b.WriteString(input[i : i+1+end+1])
				i += 1 + end + 1
			}
		default:
			b.WriteByte(input[i])
			i++
		}
	}
	return b.String()
}

func normalizeBracedAnimationCommand(rawCommand string) (string, bool) {
	trimmed := strings.TrimSpace(rawCommand)
	if trimmed == "" {
		return "", false
	}
	if strings.Contains(trimmed, "||") {
		parts := strings.SplitN(trimmed, "||", 2)
		cmd := normalizeAnimationToken(parts[0])
		if cmd != "playanimationwi" && cmd != "playanimation" {
			return "", false
		}
		return formatAnimationCommand(resolveAnimationParameter(parts[1])), true
	}

	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return "", false
	}
	if len(fields) == 1 && strings.HasPrefix(fields[0], "[") && strings.HasSuffix(fields[0], "]") {
		return formatAnimationCommand(resolveAnimationParameter(fields[0])), true
	}
	cmd := normalizeAnimationToken(fields[0])
	if cmd == "playanimationwi" || cmd == "playanimation" {
		if len(fields) < 2 {
			return formatAnimationCommand("thinking"), true
		}
		return formatAnimationCommand(resolveAnimationParameter(fields[1])), true
	}
	if len(fields) == 1 && isLikelyAnimationToken(fields[0]) {
		return formatAnimationCommand(resolveAnimationParameter(fields[0])), true
	}
	return "", false
}

func formatAnimationCommand(param string) string {
	return "{{playAnimationWI||" + param + "}}"
}

// resolveAnimationParameter reads its alias vocabulary from animationAliases
// (derived from animation_vocab.json - see animation_vocab.go). The
// "default: thinking" fallback below is real control-flow logic (genuinely
// unresolved input), not vocabulary data, so it stays as Go code.
func resolveAnimationParameter(rawParam string) string {
	token := normalizeAnimationToken(rawParam)
	for _, animThing := range animationMap {
		if token == strings.ToLower(animThing[0]) {
			return animThing[0]
		}
	}
	if canonical, ok := animationAliases[token]; ok {
		return canonical
	}
	return "thinking"
}

func normalizeAnimationToken(token string) string {
	token = strings.TrimSpace(token)
	token = strings.Trim(token, "\"'`.,!?;:(){}[]<>")
	token = strings.Join(strings.Fields(token), "")
	token = strings.ReplaceAll(token, "_", "")
	token = strings.ReplaceAll(token, "-", "")
	return strings.ToLower(token)
}

func isLikelyAnimationToken(token string) bool {
	token = normalizeAnimationToken(token)
	return token != "" && !strings.ContainsAny(token, " \t\r\n")
}

func TestNormalizeAnimationSyntax(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "legacy braced happy",
			input: "{{happy}} We made it!",
			want:  "{{playAnimationWI||happy}} We made it!",
		},
		{
			name:  "legacy braced veryHappy",
			input: "{{veryHappy}} We made it!",
			want:  "{{playAnimationWI||veryHappy}} We made it!",
		},
		{
			name:  "legacy braced sad",
			input: "{{sad}} We made it!",
			want:  "{{playAnimationWI||sad}} We made it!",
		},
		{
			name:  "legacy braced verySad",
			input: "{{verySad}} We made it!",
			want:  "{{playAnimationWI||verySad}} We made it!",
		},
		{
			name:  "legacy braced angry",
			input: "{{angry}} We made it!",
			want:  "{{playAnimationWI||angry}} We made it!",
		},
		{
			name:  "legacy braced frustrated",
			input: "{{frustrated}} We made it!",
			want:  "{{playAnimationWI||frustrated}} We made it!",
		},
		{
			name:  "legacy braced retired token falls back to thinking",
			input: "{{unknownRetiredToken}} We made it!",
			want:  "{{playAnimationWI||thinking}} We made it!",
		},
		{
			name:  "legacy braced confused",
			input: "{{confused}} We made it!",
			want:  "{{playAnimationWI||confused}} We made it!",
		},
		{
			name:  "legacy braced thinking",
			input: "{{thinking}} We made it!",
			want:  "{{playAnimationWI||thinking}} We made it!",
		},
		{
			name:  "legacy braced celebrate",
			input: "{{celebrate}} We made it!",
			want:  "{{playAnimationWI||celebrate}} We made it!",
		},
		{
			name:  "legacy braced love",
			input: "{{love}} We made it!",
			want:  "{{playAnimationWI||love}} We made it!",
		},
		{
			name:  "canonical bracketed happy",
			input: "{{[happy]}} We made it!",
			want:  "{{playAnimationWI||happy}} We made it!",
		},
		{
			name:  "canonical bracketed celebrate",
			input: "{{[celebrate]}} We made it!",
			want:  "{{playAnimationWI||celebrate}} We made it!",
		},
		{
			name:  "canonical bracketed giggles alias",
			input: "{{[giggles]}} We made it!",
			want:  "{{playAnimationWI||happy}} We made it!",
		},
		{
			name:  "canonical bracketed cheers alias",
			input: "{{[cheers]}} We made it!",
			want:  "{{playAnimationWI||celebrate}} We made it!",
		},
		{
			name:  "canonical bracketed cranky alias",
			input: "{{[cranky]}} We made it!",
			want:  "{{playAnimationWI||frustrated}} We made it!",
		},
		{
			name:  "canonical bracketed grumpy alias",
			input: "{{[grumpy]}} We made it!",
			want:  "{{playAnimationWI||frustrated}} We made it!",
		},
		{
			name:  "canonical bracketed unknown fallback",
			input: "{{[totalNonsenseEmotion]}} We made it!",
			want:  "{{playAnimationWI||thinking}} We made it!",
		},
		{
			name:  "glad alias",
			input: "[glad] We made it!",
			want:  "{{playAnimationWI||happy}} We made it!",
		},
		{
			name:  "malformed thinking braces",
			input: "{{thinking}} Brobot for life!",
			want:  "{{playAnimationWI||thinking}} Brobot for life!",
		},
		{
			name:  "valid command passthrough",
			input: "{{playAnimationWI||happy}} We made it!",
			want:  "{{playAnimationWI||happy}} We made it!",
		},
		{
			name:  "missing separator",
			input: "{{playAnimationWI happy}} Keep going.",
			want:  "{{playAnimationWI||happy}} Keep going.",
		},
		{
			name:  "empty parameter",
			input: "{{playAnimationWI||}} Keep going.",
			want:  "{{playAnimationWI||thinking}} Keep going.",
		},
		{
			name:  "extra whitespace",
			input: "  {{  playAnimationWI  ||  thinking  }}   Keep going.  ",
			want:  "  {{playAnimationWI||thinking}}   Keep going.  ",
		},
		{
			name:  "mixed casing",
			input: "{{PLAYANIMATIONWI||VERYHAPPY}} Keep going.",
			want:  "{{playAnimationWI||veryHappy}} Keep going.",
		},
		{
			name:  "punctuation around emotion",
			input: "[thinking!] Keep going.",
			want:  "{{playAnimationWI||thinking}} Keep going.",
		},
		{
			name:  "unresolved animation attempt",
			input: "[unresolvedEmotion] Keep going.",
			want:  "{{playAnimationWI||thinking}} Keep going.",
		},
		{
			name:  "plain speech unchanged",
			input: "Normal spoken sentence.",
			want:  "Normal spoken sentence.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("NormalizeAnimationSyntax panicked: %v", r)
				}
			}()
			got := NormalizeAnimationSyntax(tt.input)
			if got != tt.want {
				t.Fatalf("NormalizeAnimationSyntax(%q) = %q, want %q", tt.input, got, tt.want)
			}
			gotAgain := NormalizeAnimationSyntax(got)
			if gotAgain != got {
				t.Fatalf("NormalizeAnimationSyntax is not idempotent: first %q, second %q", got, gotAgain)
			}
		})
	}
}

func TestGetActionsFromStringNormalizesAnimationSyntax(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantParam string
		wantText  string
	}{
		{
			name:      "glad alias becomes happy",
			input:     "[glad] We made it!",
			wantParam: "happy",
			wantText:  "We made it!",
		},
		{
			name:      "malformed braces become thinking",
			input:     "{{thinking}} Brobot for life!",
			wantParam: "thinking",
			wantText:  "Brobot for life!",
		},
		{
			name:      "unresolved attempt becomes thinking",
			input:     "[unresolvedEmotion] Keep going.",
			wantParam: "thinking",
			wantText:  "Keep going.",
		},
		{
			name:      "canonical bracketed celebrate keeps pause text",
			input:     "{{[celebrate]}}... We made it!",
			wantParam: "celebrate",
			wantText:  "We made it!",
		},
		{
			name:      "canonical bracketed giggles keeps pause text",
			input:     "{{[giggles]}}... We made it!",
			wantParam: "happy",
			wantText:  "We made it!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("GetActionsFromString panicked: %v", r)
				}
			}()
			actions := GetActionsFromString(tt.input)
			if len(actions) != 2 {
				t.Fatalf("len(actions) = %d, want 2: %#v", len(actions), actions)
			}
			if actions[0].Action != ActionPlayAnimationWI || actions[0].Parameter != tt.wantParam {
				t.Fatalf("animation action = %#v, want ActionPlayAnimationWI %q", actions[0], tt.wantParam)
			}
			if actions[1].Action != ActionSayText || actions[1].Parameter != tt.wantText {
				t.Fatalf("text action = %#v, want ActionSayText %q", actions[1], tt.wantText)
			}
		})
	}
}

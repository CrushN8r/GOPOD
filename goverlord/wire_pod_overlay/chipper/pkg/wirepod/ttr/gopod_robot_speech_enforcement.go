package wirepod_ttr

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// gopodRobotSpeechRegex is derived from animation_vocab.json (verified
// tokens only) - see animation_vocab.go.
// 2026-09-30: "___" is the animation marker the model writes (gopod_animation_marker.go);
// "..." is Wire-Pod's internal form after conversion. Both pass.
var gopodRobotSpeechRegex = `^(` + gopodAnimationRegexAlternation + `)(?:\.\.\.|` + gopodAnimationMarkerRegex() + `) .+`

// gopodRobotSpeechFallbackLine is derived from system_speech_content.json -
// see gopod_system_speech_content.go. Confirmed live: read by
// emotional_beat_actions.go and gopod_render_scaffold.go.
var gopodRobotSpeechFallbackLine = loadedGOPODSystemSpeechContent.FallbackLine

// canonicalActionParameterFailuresPath resolves relative to gopodChipperRoot()
// (see gopod_paths.go), overridable via GOPOD_ACTION_PARAMETER_FAILURES_PATH.
var canonicalActionParameterFailuresPath = gopodPathWithEnvOverride("GOPOD_ACTION_PARAMETER_FAILURES_PATH", "gopod_probes/actionParameter-failures-list.txt")

// The last thing the model reads before the chat: the reply shape once more, as
// final_output_check in gopod_brobots_settings.json (2026-10-05). It used to show
// a template ("actionParameter___ spoken thought") and a regex; gemma2:2b copied
// the template into about 45% of replies (THINKING_ORDER_REPLAY_001.md), so it now
// points at the one real example line (reply_example) instead.
var gopodFinalOutputGuard = gopodFillPromptTokens(gopodBrobotsSettings.FinalOutputCheck)

var (
	gopodRobotSpeechLinePattern      = regexp.MustCompile(gopodRobotSpeechRegex)
	gopodRobotSpeechLooseDotsPattern = regexp.MustCompile(`(?i)^([A-Za-z][A-Za-z0-9_-]*(?:\s+[A-Za-z][A-Za-z0-9_-]*)?)\s*\.\s*\.\s*\.\s*(.*)$`)
	gopodBracketActionPattern        = regexp.MustCompile(`^\[([A-Za-z][A-Za-z0-9_-]*)\]\s*(?:\.\.\.)?\s*(.*)$`)
	gopodObjectActionPattern         = regexp.MustCompile(`(?i)^\{\s*actionParameter\s*:\s*([A-Za-z][A-Za-z0-9_-]*)\s*\}\s*(.*)$`)
	gopodStageDirectionPattern       = regexp.MustCompile(`\*[^*\n]+\*`)
)

type gopodRobotSpeechValidation struct {
	Valid            bool
	FirstFailingLine string
	FailureType      string
}

type actionParameterFailureEntry struct {
	BadPattern string
	RepairType string
	Target     string
}

type gopodRobotSpeechRepairer func(context.Context, openai.ChatCompletionRequest) (string, error)

type gopodContactBeatStage string

const (
	gopodContactBeatMemoryLoading   gopodContactBeatStage = "memory_loading"
	gopodContactBeatScenarioLoading gopodContactBeatStage = "scenario_loading"
	gopodContactBeatFormatting      gopodContactBeatStage = "formatting"
	gopodContactBeatRepair          gopodContactBeatStage = "repair"
	gopodContactBeatFallback        gopodContactBeatStage = "fallback"
)

type gopodContactBeatTemplate struct {
	Stage    gopodContactBeatStage `json:"stage"`
	Scenario string                `json:"scenario"`
	Line     string                `json:"line"`
}

// gopodContactBeatTemplates is derived from system_speech_content.json -
// see gopod_system_speech_content.go. NOTE: as of this pass, no production
// call site invokes selectGOPODContactBeat (confirmed by grep - only this
// file's own tests reference it), so this content is not currently spoken
// by the live GOPOD Brobot path. Migrated anyway since it's still a
// hand-typed Go literal that would need a rebuild to edit if/when it is
// wired up; flagged here rather than left as a silent surprise.
var gopodContactBeatTemplates = loadedGOPODSystemSpeechContent.ContactBeatTemplates

func validateGOPODRobotSpeech(raw string) gopodRobotSpeechValidation {
	lines := nonEmptyRobotSpeechLines(raw)
	if len(lines) == 0 {
		return gopodRobotSpeechValidation{FirstFailingLine: strings.TrimSpace(raw), FailureType: "empty/garbled output"}
	}
	for _, line := range lines {
		if gopodRobotSpeechLinePattern.MatchString(line) {
			continue
		}
		return gopodRobotSpeechValidation{
			FirstFailingLine: line,
			FailureType:      classifyGOPODRobotSpeechFailure(line),
		}
	}
	return gopodRobotSpeechValidation{Valid: true}
}

func classifyGOPODRobotSpeechFailure(line string) string {
	trimmed := strings.TrimSpace(line)
	lower := strings.ToLower(trimmed)
	switch {
	case trimmed == "":
		return "empty/garbled output"
	case strings.HasPrefix(lower, "note:") || strings.HasPrefix(lower, "response:") || strings.HasSuffix(trimmed, ":"):
		return "note/label/header"
	case strings.HasPrefix(trimmed, "*") || gopodStageDirectionPattern.MatchString(trimmed):
		return "stage direction"
	case strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{"):
		return "command/parser syntax"
	}
	if match := gopodRobotSpeechLooseDotsPattern.FindStringSubmatch(trimmed); match != nil {
		action := match[1]
		rest := match[2]
		if _, ok := knownAnimationParameter(action); !ok {
			if _, aliasOK := knownGOPODActionAlias(action); aliasOK {
				return "actionParameter typo"
			}
			return "invalid actionParameter"
		}
		if strings.Contains(trimmed[:strings.Index(trimmed, "...")+3], " ") || strings.Contains(trimmed[:strings.Index(trimmed, "...")+3], "\t") {
			return "bad spacing around dots"
		}
		if rest != "" && !strings.HasPrefix(rest, " ") {
			return "missing space after dots"
		}
		return "missing literal ... "
	}
	if strings.Contains(trimmed, "...") {
		head := strings.TrimSpace(strings.SplitN(trimmed, "...", 2)[0])
		if _, aliasOK := knownGOPODActionAlias(head); aliasOK {
			return "actionParameter typo"
		}
		if head != "" {
			return "invalid actionParameter"
		}
		return "missing actionParameter"
	}
	if !containsLetterOrDigit(trimmed) {
		return "empty/garbled output"
	}
	return "plain prose"
}

func normalizeGOPODRobotSpeech(raw string) (string, gopodRobotSpeechValidation) {
	entries, _ := loadActionParameterFailureEntries(canonicalActionParameterFailuresPath)
	var normalized []string
	for _, line := range robotSpeechSegmentsForNormalization(raw) {
		// GOPOD 2026-09-29: a line that is only a stage direction ("*giggles and
		// winks*") is skipped, not treated as the end of the speech - it used to
		// drop every line after it.
		if gopodIsStageDirectionOnlyLine(line) {
			continue
		}
		fixed, ok := normalizeGOPODRobotSpeechLine(line, entries)
		if !ok {
			return strings.Join(normalized, "\n"), validateGOPODRobotSpeech(strings.Join(normalized, "\n"))
		}
		normalized = append(normalized, fixed)
	}
	output := strings.Join(normalized, "\n")
	return output, validateGOPODRobotSpeech(output)
}

// gopodIsStageDirectionOnlyLine reports a non-empty line that the stage
// direction strip empties completely.
func gopodIsStageDirectionOnlyLine(line string) bool {
	line = strings.TrimSpace(ansiEscapePattern.ReplaceAllString(strings.TrimSpace(normalizeBrobotRobotCopy(line)), ""))
	return line != "" && strings.TrimSpace(gopodStripStageDirections(line)) == ""
}

func normalizeGOPODRobotSpeechLine(line string, entries map[string]actionParameterFailureEntry) (string, bool) {
	line = strings.TrimSpace(normalizeBrobotRobotCopy(line))
	line = strings.TrimSpace(ansiEscapePattern.ReplaceAllString(line, ""))
	// GOPOD 2026-09-29: only listed stage directions (*laugh*) are dropped; any
	// other *emphasis* is spoken without its asterisks - gopod_stage_directions.go.
	line = strings.TrimSpace(gopodStripStageDirections(line))
	if line == "" {
		return "", false
	}
	if gopodRobotSpeechLinePattern.MatchString(line) {
		return line, true
	}
	if action, speech, ok := normalizeGOPODCommandSyntaxLine(line); ok {
		return action + "... " + speech, true
	}
	if match := gopodBracketActionPattern.FindStringSubmatch(line); match != nil {
		action := normalizeGOPODActionAlias(match[1], entries)
		if action == "" {
			return "", false
		}
		return action + "... " + strings.TrimSpace(match[2]), true
	}
	if match := gopodObjectActionPattern.FindStringSubmatch(line); match != nil {
		action := normalizeGOPODActionAlias(match[1], entries)
		if action == "" {
			return "", false
		}
		return action + "... " + strings.TrimSpace(match[2]), true
	}
	if match := gopodRobotSpeechLooseDotsPattern.FindStringSubmatch(line); match != nil {
		action := normalizeGOPODActionAlias(match[1], entries)
		text := strings.TrimSpace(match[2])
		if text == "" {
			return "", false
		}
		text = normalizeGOPODSpeechEllipsisSpacing(text)
		if action == "" {
			return "thinking... " + normalizeGOPODSpeechEllipsisSpacing(strings.TrimSpace(match[1])) + "... " + text, true
		}
		// 2026-10-02: an alias is a real word - said, and it picks the animation
		// (gopod_word_animations.go). A token name stays a cue only.
		if !gopodIsAnimationTokenName(match[1]) {
			return action + "... " + strings.TrimSpace(match[1]) + "... " + text, true
		}
		return action + "... " + text, true
	}
	if !containsLetterOrDigit(line) {
		return "", false
	}
	return "thinking... " + line, true
}

func normalizeGOPODActionAlias(raw string, entries map[string]actionParameterFailureEntry) string {
	action := strings.TrimSpace(raw)
	if known, ok := knownAnimationParameter(action); ok {
		return known
	}
	if known, ok := knownGOPODActionAlias(action); ok {
		return known
	}
	if likely := resolveGOPODLikelyAnimationAlias(action); likely != "" {
		return likely
	}
	if entry, ok := entries[strings.ToLower(action+"...")]; ok {
		target := strings.TrimSuffix(strings.TrimSpace(entry.Target), "...")
		if known, ok := knownAnimationParameter(target); ok {
			return known
		}
	}
	return ""
}

// knownGOPODActionAlias and resolveGOPODLikelyAnimationAlias both read from
// animationAliases (derived from animation_vocab.json - see
// animation_vocab.go). Normalization is now shared via normalizeAnimationToken
// for both, fixing a prior inconsistency where knownGOPODActionAlias used a
// narrower ad hoc normalization.
func knownGOPODActionAlias(raw string) (string, bool) {
	if canonical, ok := animationAliases[normalizeAnimationToken(raw)]; ok {
		return canonical, true
	}
	return "", false
}

func resolveGOPODLikelyAnimationAlias(raw string) string {
	token := normalizeAnimationToken(raw)
	if canonical, ok := animationAliases[token]; ok {
		return canonical
	}
	return ""
}

func normalizeGOPODCommandSyntaxLine(line string) (string, string, bool) {
	normalizedSyntax := NormalizeAnimationSyntax(line)
	if normalizedSyntax == line && !strings.Contains(line, "{{") {
		return "", "", false
	}
	actions := GetActionsFromString(normalizedSyntax)
	actionParameter := ""
	for _, action := range actions {
		if action.Action == ActionPlayAnimationWI || action.Action == ActionPlayAnimation {
			actionParameter = normalizeActionParameter(action.Parameter)
			break
		}
	}
	if actionParameter == "" {
		return "", "", false
	}
	speech := cleanRobotSpeechText(robotSafeFilter(robotSpeechFromActions(actions)))
	if speech == "" {
		speech = cleanRobotSpeechText(robotSafeFilter(normalizedSyntax))
	}
	if speech == "" {
		return "", "", false
	}
	return actionParameter, speech, true
}

func loadActionParameterFailureEntries(path string) (map[string]actionParameterFailureEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	entries := make(map[string]actionParameterFailureEntry)
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) != 3 {
			continue
		}
		entry := actionParameterFailureEntry{
			BadPattern: strings.TrimSpace(parts[0]),
			RepairType: strings.TrimSpace(parts[1]),
			Target:     strings.TrimSpace(parts[2]),
		}
		entries[strings.ToLower(entry.BadPattern)] = entry
	}
	return entries, nil
}

func buildGOPODStreamResultWithRepair(ctx context.Context, raw, model string, repairer gopodRobotSpeechRepairer) gopodStreamResult {
	ir := NewGOPODSemanticIR(gopodStripCueWordNotes(raw))
	rich := RenderGOPODResponse(ir, GOPODRenderModeRich)
	flat := RenderGOPODResponse(ir, GOPODRenderModeFlat)
	return gopodStreamResult{
		Raw:            ir.Content,
		MarkedPacket:   rich.MarkedPacket,
		FilteredPacket: flat.FilteredPacket,
		AudibleChunks:  flat.AudibleChunks,
		Fallback:       strings.TrimSpace(flat.Content) == strings.TrimSpace(gopodRobotSpeechFallbackLine),
	}
}

// gopodDisplayForSpeech: the bubble's words for what's said - the reply as
// written, or the fallback line when that's what's said (2026-10-24, live: the
// fallback said while the bubble showed the reply). The live reply and the
// highlight check both take it.
func gopodDisplayForSpeech(display string, sr gopodStreamResult) string {
	if sr.Fallback {
		return gopodStripAnimationCueForDisplay(gopodRobotSpeechFallbackLine)
	}
	// 2026-10-27 (operator: "2026 * should be observed as 20 26 * spoken & heard,
	// highlights follow 20 26 spoken/heard pattern"): a spoken form in
	// spoken_symbols.show_as_said is shown as it's said, so the bubble's words
	// and the highlight follow what's heard.
	for _, form := range gopodBrobotsSettings.SpokenSymbols.ShowAsSaid {
		switch form {
		case "years":
			display = gopodSpeakYears(display)
		}
	}
	return display
}

func createGOPODRobotSpeechRepairRequest(model, raw string) openai.ChatCompletionRequest {
	instruction := "Rewrite malformed output into valid " + gopodBrobotsSettings.PlaceholderWord + gopodAnimationMarker + " " + gopodBrobotsSettings.PlaceholderThought + " lines.\n" +
		"Do not add new facts.\n" +
		"Do not add headings, notes, explanations, labels, markdown, brackets, braces, or stage directions.\n" +
		"Preserve meaning as much as possible.\n" +
		"Output only valid lines.\n" +
		"Every non-empty line must pass:\n" +
		gopodRobotSpeechRegex
	return openai.ChatCompletionRequest{
		Model:       model,
		MaxTokens:   5120,
		Temperature: 0,
		TopP:        1,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: instruction},
			{Role: openai.ChatMessageRoleUser, Content: raw},
		},
		Stream: false,
	}
}

func openAIGOPODRobotSpeechRepairer(client *openai.Client) gopodRobotSpeechRepairer {
	return func(ctx context.Context, req openai.ChatCompletionRequest) (string, error) {
		response, err := client.CreateChatCompletion(ctx, req)
		if err != nil {
			return "", err
		}
		if len(response.Choices) == 0 {
			return "", nil
		}
		return response.Choices[0].Message.Content, nil
	}
}

func selectGOPODContactBeat(stage gopodContactBeatStage, scenario string) (string, bool) {
	scenario = strings.ToLower(strings.TrimSpace(scenario))
	for _, beat := range gopodContactBeatTemplates {
		if beat.Stage == stage && beat.Scenario != "" && strings.ToLower(beat.Scenario) == scenario {
			return beat.Line, true
		}
	}
	for _, beat := range gopodContactBeatTemplates {
		if beat.Stage == stage && beat.Scenario == "" {
			return beat.Line, true
		}
	}
	return "", false
}

func nonEmptyRobotSpeechLines(raw string) []string {
	var lines []string
	for _, line := range strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func normalizeGOPODSpeechEllipsisSpacing(raw string) string {
	return regexp.MustCompile(`\s+\.\.\.`).ReplaceAllString(raw, "...")
}

func robotSpeechSegmentsForNormalization(raw string) []string {
	if chunks, ok := splitCanonicalGOPODChunks(raw); ok {
		return chunks
	}
	return nonEmptyRobotSpeechLines(raw)
}

func containsLetterOrDigit(raw string) bool {
	for _, r := range raw {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return true
		}
	}
	return false
}

func actionParameterFailureListPathForRepoRoot(root string) string {
	return filepath.Join(root, "gopod_probes", "actionParameter-failures-list.txt")
}

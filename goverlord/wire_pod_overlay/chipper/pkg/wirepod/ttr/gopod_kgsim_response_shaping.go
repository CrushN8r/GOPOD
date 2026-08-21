package wirepod_ttr

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// These four resolve relative to gopodChipperRoot() (see gopod_paths.go),
// each individually overridable by its own env var. Single-owner: moving
// the chipper/ tree, or relocating just one of these files, no longer
// requires a hand-edit here.
var canonicalBrobotResponsePromptPath = gopodPathWithEnvOverride("GOPOD_BROBOT_RESPONSE_PROMPT_PATH", "wire-pod_response_prompt.txt")
var canonicalBrobotIdentityPromptPath = gopodPathWithEnvOverride("GOPOD_BROBOT_IDENTITY_PROMPT_PATH", "wire-pod_brobots_prompt.txt")
var canonicalGOPODCorePromptPath = gopodPathWithEnvOverride("GOPOD_CORE_PROMPT_PATH", "gopod_probes/foundation_packs/gopod_core.txt")
var canonicalScenarioRegistryPath = gopodPathWithEnvOverride("GOPOD_SCENARIO_REGISTRY_PATH", "gopod_probes/scenario_packs/registry.json")

// defaultPrompt is derived from system_speech_content.json - see
// gopod_system_speech_content.go. This is the last-resort fallback used
// only when the 3 canonical prompt files AND vars.APIConfig.Knowledge.OpenAIPrompt
// have already failed/are empty, so its own loader degrades to a built-in
// floor value (gopodDefaultPromptFloor) rather than failing loudly - the
// fallback's own failure must not cascade into breaking every request.
var defaultPrompt = loadedGOPODSystemSpeechContent.DefaultPrompt

const (
	gopodCanonicalStartMarker = "|||0|||"
	gopodCanonicalChunkMarker = "|||C|||"
	gopodCanonicalEndMarker   = "|||1|||"
)

var (
	malformedGOPODPipeMarkerPattern  = regexp.MustCompile(`\|{2,3}(?:[A-Za-z][A-Za-z0-9_, -]*|[01C])\|{2,4}`)
	malformedGOPODChunkMarkerPattern = regexp.MustCompile(`\|{2,3}C\|{2,4}`)
	ansiEscapePattern                = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)
	htmlTagPattern                   = regexp.MustCompile(`<[^>\n]+>`)
	leakedNewVoiceRequestPattern     = regexp.MustCompile(`(?i)\bnewVoiceRequest\b`)
	markdownHeadingPattern           = regexp.MustCompile(`(?m)^\s{0,3}#{1,6}\s+`)
	markdownListPattern              = regexp.MustCompile(`(?m)^\s*[*+-]\s+`)
)

type gopodStreamResult struct {
	Raw            string
	MarkedPacket   string
	FilteredPacket string
	AudibleChunks  []string
}

type brobotPhraseUnit struct {
	ActionParameter string
	Text            string
}

type scenarioPacketRegistry struct {
	Packets []scenarioPacketEntry `json:"packets"`
}

type scenarioPacketEntry struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Triggers []string `json:"triggers"`
}

type loadedScenarioPacket struct {
	Name    string
	File    string
	Content []byte
}

var brobotPhraseUnitPattern = regexp.MustCompile(`(?i)(^|\s)([A-Za-z][A-Za-z0-9_-]*)\s*\.\.\.`)

func parseBrobotPhraseUnits(raw string) []brobotPhraseUnit {
	trimmed := strings.TrimSpace(normalizeBrobotRobotCopy(raw))
	if trimmed == "" {
		return nil
	}
	matches := brobotPhraseUnitPattern.FindAllStringSubmatchIndex(trimmed, -1)
	if len(matches) == 0 {
		return nil
	}
	// Phrase-chunks start only at allowlisted actionParameter + normalized ellipsis.
	// Other ellipses belong to speech; emotion words may influence parameter choice elsewhere.
	type phraseBoundary struct {
		match               []int
		actionParameter     string
		consumedFirstPrefix bool
	}
	boundaries := make([]phraseBoundary, 0, len(matches))
	var units []brobotPhraseUnit
	for i, match := range matches {
		actionParameter := strings.TrimSpace(trimmed[match[4]:match[5]])
		if i == 0 {
			prefix := strings.TrimSpace(trimmed[:match[0]])
			if prefix != "" {
				if known, ok := knownAnimationParameter(prefix + " " + actionParameter); ok {
					boundaries = append(boundaries, phraseBoundary{
						match:               match,
						actionParameter:     known,
						consumedFirstPrefix: true,
					})
					continue
				}
			}
		}
		if known, ok := knownAnimationParameter(actionParameter); ok {
			boundaries = append(boundaries, phraseBoundary{
				match:           match,
				actionParameter: known,
			})
			continue
		}
	}
	if len(boundaries) == 0 {
		return nil
	}
	if prefix := strings.TrimSpace(trimmed[:boundaries[0].match[0]]); prefix != "" && !boundaries[0].consumedFirstPrefix {
		units = append(units, brobotPhraseUnit{
			ActionParameter: "thinking",
			Text:            prefix,
		})
	}
	for i, boundary := range boundaries {
		match := boundary.match
		textStart := match[1]
		if boundary.consumedFirstPrefix {
			textStart = match[1]
		}
		textEnd := len(trimmed)
		if i+1 < len(boundaries) {
			textEnd = boundaries[i+1].match[0]
		}
		text := strings.TrimSpace(trimmed[textStart:textEnd])
		if text == "" {
			continue
		}
		units = append(units, brobotPhraseUnit{
			ActionParameter: boundary.actionParameter,
			Text:            text,
		})
	}
	return units
}

func normalizeActionParameter(rawAction string) string {
	if known, ok := knownAnimationParameter(rawAction); ok {
		return known
	}
	resolved := resolveAnimationParameter(rawAction)
	if resolved == "" {
		return "thinking"
	}
	return resolved
}

func knownAnimationParameter(rawAction string) (string, bool) {
	token := normalizeAnimationToken(rawAction)
	for _, animThing := range animationMap {
		if token == strings.ToLower(animThing[0]) {
			return animThing[0], true
		}
	}
	return "", false
}

func composeBrobotMarkedPacket(units []brobotPhraseUnit) string {
	if len(units) == 0 {
		return ""
	}
	chunks := make([]string, 0, len(units))
	for _, unit := range units {
		text := strings.TrimSpace(unit.Text)
		if text == "" {
			continue
		}
		chunks = append(chunks, formatAnimationCommand(normalizeActionParameter(unit.ActionParameter))+"... "+text)
	}
	if len(chunks) == 0 {
		return ""
	}
	return gopodCanonicalStartMarker + " " + strings.Join(chunks, " "+gopodCanonicalChunkMarker+" ") + " " + gopodCanonicalEndMarker
}

func composeBrobotMarkedPacketFromChunks(chunks []string) string {
	filteredChunks := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		if chunk = strings.TrimSpace(chunk); chunk != "" {
			filteredChunks = append(filteredChunks, chunk)
		}
	}
	if len(filteredChunks) == 0 {
		return ""
	}
	return gopodCanonicalStartMarker + " " + strings.Join(filteredChunks, " "+gopodCanonicalChunkMarker+" ") + " " + gopodCanonicalEndMarker
}

func composeBrobotRawMarkedPacket(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	return gopodCanonicalStartMarker + " " + raw + " " + gopodCanonicalEndMarker
}

func composeBrobotFilterPacket(units []brobotPhraseUnit) string {
	if len(units) == 0 {
		return ""
	}
	chunks := make([]string, 0, len(units))
	for _, unit := range units {
		text := strings.TrimSpace(unit.Text)
		if text == "" {
			continue
		}
		chunks = append(chunks, normalizeActionParameter(unit.ActionParameter)+"... "+text)
	}
	return strings.Join(chunks, " "+gopodCanonicalChunkMarker+" ")
}

func splitBrobotFilterPacket(packet string) []string {
	filteredChunks := make([]string, 0)
	for _, chunk := range strings.Split(packet, gopodCanonicalChunkMarker) {
		if chunk = strings.TrimSpace(chunk); chunk != "" {
			filteredChunks = append(filteredChunks, chunk)
		}
	}
	return filteredChunks
}

func hasCanonicalGOPODMarker(raw string) bool {
	return strings.Contains(raw, gopodCanonicalStartMarker) ||
		strings.Contains(raw, gopodCanonicalChunkMarker) ||
		strings.Contains(raw, gopodCanonicalEndMarker) ||
		malformedGOPODPipeMarkerPattern.MatchString(raw)
}

func stripCanonicalGOPODMarkers(input string) string {
	input = strings.ReplaceAll(input, gopodCanonicalStartMarker, "")
	input = strings.ReplaceAll(input, gopodCanonicalChunkMarker, "")
	input = strings.ReplaceAll(input, gopodCanonicalEndMarker, "")
	input = malformedGOPODPipeMarkerPattern.ReplaceAllString(input, "")
	return strings.TrimSpace(input)
}

func canonicalGOPODBody(raw string) string {
	body := raw
	if start := strings.Index(body, gopodCanonicalStartMarker); start >= 0 {
		body = body[start+len(gopodCanonicalStartMarker):]
	}
	if end := strings.Index(body, gopodCanonicalEndMarker); end >= 0 {
		body = body[:end]
	}
	return body
}

func splitCanonicalGOPODChunks(raw string) ([]string, bool) {
	if !hasCanonicalGOPODMarker(raw) {
		return nil, false
	}
	var chunks []string
	body := malformedGOPODChunkMarkerPattern.ReplaceAllString(canonicalGOPODBody(raw), gopodCanonicalChunkMarker)
	for _, rawChunk := range strings.Split(body, gopodCanonicalChunkMarker) {
		chunk := stripCanonicalGOPODMarkers(rawChunk)
		if chunk == "" {
			continue
		}
		chunk = robotSafeFilter(chunk)
		if chunk == "" {
			continue
		}
		chunks = append(chunks, chunk)
	}
	return chunks, true
}

func normalizeBrobotRobotCopy(raw string) string {
	return strings.ReplaceAll(raw, "…", "...")
}

func robotSafeFilter(raw string) string {
	filtered := normalizeBrobotRobotCopy(raw)
	filtered = ansiEscapePattern.ReplaceAllString(filtered, "")
	filtered = htmlTagPattern.ReplaceAllString(filtered, "")
	filtered = leakedNewVoiceRequestPattern.ReplaceAllString(filtered, "")
	filtered = markdownHeadingPattern.ReplaceAllString(filtered, "")
	filtered = markdownListPattern.ReplaceAllString(filtered, "")
	filtered = strings.ReplaceAll(filtered, "**", "")
	filtered = strings.ReplaceAll(filtered, "__", "")
	filtered = strings.ReplaceAll(filtered, "`", "")
	filtered = strings.ReplaceAll(filtered, "|", " ")
	filtered = strings.Join(strings.Fields(removeSpecialCharacters(filtered)), " ")
	for _, replacement := range []struct {
		old string
		new string
	}{
		{" .", "."},
		{" ?", "?"},
		{" !", "!"},
		{" ,", ","},
		{" ;", ";"},
		{" :", ":"},
	} {
		filtered = strings.ReplaceAll(filtered, replacement.old, replacement.new)
	}
	return filtered
}

func filteredPhraseUnitsFromRaw(raw string) []brobotPhraseUnit {
	var rawChunks []string
	if hasCanonicalGOPODMarker(raw) {
		body := malformedGOPODChunkMarkerPattern.ReplaceAllString(canonicalGOPODBody(raw), gopodCanonicalChunkMarker)
		rawChunks = strings.Split(body, gopodCanonicalChunkMarker)
	} else {
		rawChunks = []string{raw}
	}
	var units []brobotPhraseUnit
	for _, rawChunk := range rawChunks {
		chunk := stripCanonicalGOPODMarkers(rawChunk)
		filtered := robotSafeFilter(chunk)
		parsed := parseBrobotPhraseUnits(filtered)
		if len(parsed) > 0 {
			units = append(units, parsed...)
			continue
		}
		actionInput := ansiEscapePattern.ReplaceAllString(chunk, "")
		actions := GetActionsFromString(actionInput)
		actionParameter := "thinking"
		hasCommandAction := false
		for _, action := range actions {
			if action.Action == ActionPlayAnimationWI || action.Action == ActionPlayAnimation {
				actionParameter = action.Parameter
				hasCommandAction = true
				break
			}
		}
		speech := ""
		if hasCommandAction {
			speech = robotSafeFilter(robotSpeechFromActions(actions))
		}
		if speech == "" {
			speech = filtered
		}
		if hasCommandAction {
			speech = cleanRobotSpeechText(speech)
		}
		if speech == "" {
			continue
		}
		units = append(units, brobotPhraseUnit{
			ActionParameter: actionParameter,
			Text:            speech,
		})
	}
	return units
}

func commandPacketFromFilteredChunk(chunk string) string {
	units := parseBrobotPhraseUnits(chunk)
	if len(units) == 0 {
		text := cleanRobotSpeechText(robotSafeFilter(chunk))
		if text == "" {
			return ""
		}
		return formatAnimationCommand("thinking") + "... " + text
	}
	unit := units[0]
	text := cleanRobotSpeechText(robotSafeFilter(unit.Text))
	if text == "" {
		return ""
	}
	return formatAnimationCommand(normalizeActionParameter(unit.ActionParameter)) + "... " + text
}

func buildGOPODStreamResult(raw string) gopodStreamResult {
	return buildGOPODStreamResultWithRepair(context.Background(), raw, "", nil)
}

func joinResponseChunks(chunks []string) string {
	return strings.TrimSpace(strings.Join(chunks, " "))
}

// displayNewlineRunLimit caps how many consecutive newlines a log/debug display
// line can show. Set to 2 on operator request (2026-08-02) to make repeated
// LLM trailing-newline behavior visible at a readable size instead of 5-6 in a
// row; display-only, never applied to the values actually parsed into speech.
const displayNewlineRunLimit = 2

func collapseNewlineRuns(s string, maxRun int) string {
	var b strings.Builder
	run := 0
	for _, r := range s {
		if r == '\n' {
			run++
			if run > maxRun {
				continue
			}
		} else {
			run = 0
		}
		b.WriteRune(r)
	}
	return b.String()
}

func emitRawLLMDebug(raw string) {
	logger.LogDebugUI("RAW_LLM_RESPONSE_MARKED:\n" + collapseNewlineRuns(raw, displayNewlineRunLimit))
}

func emitMarkedPacketDebug(markedPacket string) {
	logger.LogDebugUI("BROBOT_MARKED_PACKET:\n" + collapseNewlineRuns(markedPacket, displayNewlineRunLimit))
}

func emitFilterPacketDebug(filteredPacket string) {
	logger.LogDebugUI("BROBOT_FILTER_PACKET:\n" + collapseNewlineRuns(filteredPacket, displayNewlineRunLimit))
}

func emitFinalLLMDisplay(esn, raw string) {
	logger.LogUI("LLM response for " + esn + ": " + collapseNewlineRuns(raw, displayNewlineRunLimit))
}

func appendPromptSection(promptBytes []byte, sectionBytes []byte) []byte {
	if len(sectionBytes) == 0 {
		return promptBytes
	}
	if len(promptBytes) > 0 && promptBytes[len(promptBytes)-1] != '\n' {
		promptBytes = append(promptBytes, '\n')
	}
	return append(promptBytes, sectionBytes...)
}

func loadScenarioPacketForMessage(message string) (loadedScenarioPacket, bool) {
	registryBytes, err := os.ReadFile(canonicalScenarioRegistryPath)
	if err != nil {
		logger.Println("Unable to load scenario packet registry: " + err.Error())
		return loadedScenarioPacket{}, false
	}
	var registry scenarioPacketRegistry
	if err := json.Unmarshal(registryBytes, &registry); err != nil {
		logger.Println("Unable to parse scenario packet registry: " + err.Error())
		return loadedScenarioPacket{}, false
	}
	normalizedMessage := strings.ToLower(message)
	for _, packet := range registry.Packets {
		for _, trigger := range packet.Triggers {
			if strings.Contains(normalizedMessage, strings.ToLower(trigger)) {
				packetPath := packet.File
				if !filepath.IsAbs(packetPath) {
					packetPath = filepath.Join(filepath.Dir(filepath.Dir(canonicalScenarioRegistryPath)), packetPath)
				}
				packetBytes, err := os.ReadFile(packetPath)
				if err != nil {
					logger.Println("Unable to load scenario packet file: " + err.Error())
					return loadedScenarioPacket{}, false
				}
				return loadedScenarioPacket{
					Name:    packet.Name,
					File:    packet.File,
					Content: packetBytes,
				}, true
			}
		}
	}
	return loadedScenarioPacket{}, false
}

func loadCanonicalBrobotPrompt(userMessage string) (string, int, bool) {
	responsePromptBytes, responseErr := os.ReadFile(canonicalBrobotResponsePromptPath)
	identityPromptBytes, identityErr := os.ReadFile(canonicalBrobotIdentityPromptPath)
	corePromptBytes, coreErr := os.ReadFile(canonicalGOPODCorePromptPath)
	if responseErr == nil && identityErr == nil && coreErr == nil {
		promptBytes := make([]byte, 0, len(responsePromptBytes)+1+len(identityPromptBytes)+1+len(corePromptBytes))
		promptBytes = appendPromptSection(promptBytes, responsePromptBytes)
		promptBytes = appendPromptSection(promptBytes, identityPromptBytes)
		promptBytes = appendPromptSection(promptBytes, corePromptBytes)
		if memoryPrompt := loadGOPODSessionMemoryPrompt(); strings.TrimSpace(memoryPrompt) != "" {
			promptBytes = appendPromptSection(promptBytes, []byte(memoryPrompt))
		}
		if packet, ok := loadScenarioPacketForMessage(userMessage); ok {
			promptBytes = appendPromptSection(promptBytes, packet.Content)
			logger.Println(fmt.Sprintf("BROBOT_SCENARIO_PACKET_LOADED=%s file=%s bytes=%d", packet.Name, packet.File, len(packet.Content)))
		}
		return string(promptBytes), len(promptBytes), false
	}
	if responseErr != nil {
		logger.Println("Unable to load canonical Brobot response prompt file: " + responseErr.Error())
	}
	if identityErr != nil {
		logger.Println("Unable to load canonical Brobot identity prompt file: " + identityErr.Error())
	}
	if coreErr != nil {
		logger.Println("Unable to load GOPOD core prompt file: " + coreErr.Error())
	}
	if strings.TrimSpace(vars.APIConfig.Knowledge.OpenAIPrompt) != "" {
		prompt := strings.TrimSpace(vars.APIConfig.Knowledge.OpenAIPrompt)
		return prompt, len([]byte(prompt)), true
	}
	return defaultPrompt, len([]byte(defaultPrompt)), false
}

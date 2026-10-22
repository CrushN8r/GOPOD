package wirepod_ttr

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"sync"
	"unicode"

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

// gopodIsActionCueSpot: an animation word followed by "..." ("thinking... ") starts a
// new part only where a cue can really be - at the start of the text or right after a
// sentence ends - and only with words after it. Mid-sentence it's speech: "that takes
// some thinking..." (live 2026-09-29) was cut to "that takes some", because "thinking"
// is also an animation name and the empty part after it was dropped.
func gopodIsActionCueSpot(text string, match []int) bool {
	if strings.TrimSpace(text[match[1]:]) == "" {
		return false
	}
	before := strings.TrimRight(text[:match[4]], " \t\r\n")
	if before == "" {
		return true
	}
	switch before[len(before)-1] {
	case '.', '!', '?', ':':
		return true
	}
	return false
}

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
		if known, ok := knownAnimationParameter(actionParameter); ok && gopodIsActionCueSpot(trimmed, match) {
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
	// 2026-10-17 (operator: "one event list, single source of truth"): the reply's
	// bubble comes from the feed now, not from the page reading this log line.
	GopodChatFeedAddLive(esn, collapseNewlineRuns(raw, displayNewlineRunLimit))
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
	packets := loadScenarioPacketsForMessage(message)
	if len(packets) == 0 {
		return loadedScenarioPacket{}, false
	}
	return packets[0], true
}

// gopodMaxScenarioPackets: how many topic packs one question can load.
const gopodMaxScenarioPackets = 3

// loadScenarioPacketsForMessage returns every pack whose trigger is in the
// message, in registry order, up to gopodMaxScenarioPackets. 2026-10-03 (operator:
// group the value points into topics - who we are, the brothers, party,
// background, GOPOD): a question can touch two topics ("introduce yourself and is
// your brother asleep"), and a trigger matches whole words only, so "who" doesn't
// fire on "whole" or "bar" on "barely".
func loadScenarioPacketsForMessage(message string) []loadedScenarioPacket {
	normalizedMessage := strings.ToLower(message)
	registry := gopodTopicRegistry()
	var out []loadedScenarioPacket
	for _, packet := range registry {
		if len(out) >= gopodMaxScenarioPackets {
			break
		}
		for _, trigger := range packet.Triggers {
			if !gopodTriggerOwnMatch(normalizedMessage, strings.ToLower(trigger), packet.Name, registry) {
				continue
			}
			packetBytes, err := gopodReadTopicFacts(packet.File)
			if err != nil {
				logger.Println("Unable to load scenario packet file: " + err.Error())
				break
			}
			out = append(out, loadedScenarioPacket{Name: packet.Name, File: packet.File, Content: packetBytes})
			break
		}
	}
	return out
}

// gopodTriggerOwnMatch (2026-10-16, operator: "remove chocolate from bingo *
// chocolate = food, bingo = game"): the trigger is in the text as whole words,
// somewhere not inside another topic's longer trigger - "chocolate bingo" is the
// food topic's, so its "bingo" doesn't also load the bingo topic; "bingo" alone
// still does.
func gopodTriggerOwnMatch(text, trigger, owner string, registry []scenarioPacketEntry) bool {
	for start := 0; trigger != ""; {
		i := strings.Index(text[start:], trigger)
		if i < 0 {
			return false
		}
		i += start
		end := i + len(trigger)
		start = i + 1
		if (i > 0 && gopodIsWordByte(text[i-1])) || (end < len(text) && gopodIsWordByte(text[end])) {
			continue
		}
		if !gopodInsideLongerTrigger(text, i, end, owner, registry) {
			return true
		}
	}
	return false
}

// gopodInsideLongerTrigger: text[i:end] sits inside a longer trigger of another
// topic, found there as whole words.
func gopodInsideLongerTrigger(text string, i, end int, owner string, registry []scenarioPacketEntry) bool {
	for _, packet := range registry {
		if packet.Name == owner {
			continue
		}
		for _, other := range packet.Triggers {
			other = strings.ToLower(strings.TrimSpace(other))
			if len(other) <= end-i {
				continue
			}
			for from := i - len(other) + (end - i); from <= i; from++ {
				if from < 0 || from+len(other) > len(text) || text[from:from+len(other)] != other {
					continue
				}
				oEnd := from + len(other)
				if (from == 0 || !gopodIsWordByte(text[from-1])) && (oEnd == len(text) || !gopodIsWordByte(text[oEnd])) {
					return true
				}
			}
		}
	}
	return false
}

// gopodHasWholeWords: phrase appears in text with no letter or digit right
// before or after it.
func gopodHasWholeWords(text, phrase string) bool {
	if phrase == "" {
		return false
	}
	for start := 0; ; {
		i := strings.Index(text[start:], phrase)
		if i < 0 {
			return false
		}
		i += start
		end := i + len(phrase)
		before := i == 0 || !gopodIsWordByte(text[i-1])
		after := end == len(text) || !gopodIsWordByte(text[end])
		if before && after {
			return true
		}
		start = i + 1
	}
}

func gopodIsWordByte(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
}

// gopodFillPromptTokens puts the settings' marker and placeholder word into the
// reply rules ($ANIMATION_MARKER; $PLACEHOLDER_WORD/_THOUGHT are still filled for
// an older prompt that names them). $REPLY_EXAMPLE is per brother: gopodReplyShapeSection.
func gopodFillPromptTokens(prompt string) string {
	return strings.NewReplacer(
		"$ANIMATION_MARKER", gopodAnimationMarker,
		"$PLACEHOLDER_WORD", gopodBrobotsSettings.PlaceholderWord,
		"$PLACEHOLDER_THOUGHT", gopodBrobotsSettings.PlaceholderThought,
	).Replace(prompt)
}

// gopodReplyShapeSection is the reply shape (wire-pod_response_prompt.txt) for
// one brother: the marker filled in, and $REPLY_EXAMPLE his own example lines
// (reply_examples in gopod_brobots_settings.json). It goes after his persona
// and before the final check, nearest the question (2026-10-05: with one shared
// example, gemma2:2b copied its cue word into every reply; with each brother's
// own two, the cue words follow his character - BROBOTS_LAYERED_PROMPT_002.md).
func gopodReplyShapeSection(brobot string) string {
	data, err := os.ReadFile(canonicalBrobotResponsePromptPath)
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(gopodFillPromptTokens(string(data)), "$REPLY_EXAMPLE", gopodBrobotsSettings.ReplyExamples[brobot])
}

// gopodQuestionFacts is what this question needs to know: the topic packs its
// words trigger (gopod_probes/chapters/) and the Bingo note. 2026-10-05:
// it sits nearest the question, after "who I am" and before the reply shape
// (CreateAIReq); placed earlier, gemma2:2b often answered past it
// (SOCIAL_DEMO_REPLAY_001.md: 44% of answers landed their value point, 66% with
// the facts moved and the examples fixed).
func gopodQuestionFacts(userMessage string) string {
	return gopodQuestionFactsFor("", userMessage)
}

// gopodQuestionFactsFor is gopodQuestionFacts for one Brobot: the points it
// gives out are kept for his say line and his take (gopodGivenPoints).
func gopodQuestionFactsFor(esn, userMessage string) string {
	var facts []byte
	var given []string
	notice := map[string][]string{} // 2026-10-22: the value points notice's parts (gopodNoticeSet)
	defer func() { gopodGivenSet(esn, given); gopodNoticeSet(esn, userMessage, notice) }()
	// 2026-10-01 (operator): the Bingo hosting pitch stays apart from a Bingo
	// test run. With a run in memory the pitch pack is left out, so the answer
	// comes from the run; with none (the Bingo-played flag at 0), a Bingo question
	// gets the "sure, let's play" note (bingo.no_run_note), and the pitch pack loads.
	bingoRun := gopodBingoPlayed.Load() || gopodBingoRunInMemory()
	for _, packet := range loadScenarioPacketsForMessage(userMessage) {
		if bingoRun && packet.Name == gopodBrobotsSettings.Bingo.PitchPack {
			logger.Println("BROBOT_SCENARIO_PACKET_SKIPPED=" + packet.Name + " reason=bingo_run_in_memory")
		} else {
			// 2026-10-06: the next value point only (gopod_value_points.go).
			// 2026-10-22: the point this question asks about first (gopodPickPoint).
			packFacts, point := gopodCoverageRevealFor(packet.Name, packet.Content, userMessage)
			facts = appendPromptSection(facts, packFacts)
			if point != "" {
				given = append(given, point)
				notice["points"] = append(notice["points"], gopodNoticePoint(packet.Name, packet.Content, point))
			}
			logger.Println(fmt.Sprintf("BROBOT_SCENARIO_PACKET_LOADED=%s file=%s bytes=%d", packet.Name, packet.File, len(packet.Content)))
		}
	}
	// 2026-10-10 (operator: "brobot did not check actual weather ... like in
	// scripted interview"): a weather question gets the real weather.
	// 2026-10-22: a native weather intent PLAYHEAD takes counts too ("whether").
	if note := gopodBrobotsSettings.LiveWeatherNote; note != "" && (gopodHasWholeWords(strings.ToLower(userMessage), "weather") || gopodNativeTakeIs(gopodBrobotName(esn), "weather", userMessage)) {
		// The golden weather call, the same as the scripted draft's {weather}.
		if line, live := gopodWeatherSayAt(esn, gopodWeatherPlace(userMessage)); line != "" {
			facts = appendPromptSection(facts, []byte(note+" "+line))
			notice["weather"] = append(notice["weather"], gopodNoticeFill(gopodBrobotsSettings.BubbleNotices.ValuePoints.Weather, map[string]string{"line": line}))
			logger.Println(fmt.Sprintf("BROBOT_LIVE_WEATHER=added live=%v", live))
		}
	}
	// 2026-10-17 (operator: "brobots use \"system-search\" for 'search
	// [search-text]'"): a "search ..." line gets the web's snippets as facts.
	if searched := gopodSearchFacts(userMessage); searched != "" {
		facts = appendPromptSection(facts, []byte(searched))
		notice["search"] = append(notice["search"], gopodNoticeFill(gopodBrobotsSettings.BubbleNotices.ValuePoints.Search, map[string]string{"search": gopodSearchQuery(userMessage)}))
		logger.Println("BROBOT_SYSTEM_SEARCH=added")
	}
	// 2026-10-18 (operator: "QR codes are brobot N's bubble * thought bubble *
	// brobot can say refer to QR codes below"): PLAYHEAD's QR in his bubble.
	if note := gopodManagerQRNote(); note != "" {
		facts = appendPromptSection(facts, []byte(note))
		logger.Println("BROBOT_QR_NOTE=added")
	}
	if !bingoRun && gopodIsBingoQuestion(userMessage) {
		facts = appendPromptSection(facts, []byte(gopodBrobotsSettings.Bingo.NoRunNote))
		logger.Println("BROBOT_BINGO_NO_RUN_NOTE=added")
	}
	// No topic pack and no note: the question is far from GOPOD (operator
	// 2026-10-05: "Tell me about the wilderness in the Amazon jungle ... If the
	// request drifts, the Brobots know - I have a job to do"). A quick take,
	// then a rejoinder back to one GOPOD value point (off_topic_note). A
	// thank-you is not off topic: it gets chat_hold.thanks_note instead.
	if len(facts) == 0 && strings.TrimSpace(userMessage) != "" && !gopodIsThanks(userMessage) {
		facts = appendPromptSection(facts, []byte(gopodBrobotsSettings.OffTopicNote))
		logger.Println("BROBOT_OFF_TOPIC_NOTE=added")
	}
	return string(facts)
}

// Flavour lists (scripted_snippets in gopod_brobots_settings.json; operator
// 2026-10-05: "scripted snippets to stitch together", 2026-10-06: "each build
// up a flavoured personality from a list of shared and specific party-brobot
// responses"). Exact lines the code says for a Brobot, word for word: gemma2:2b
// paraphrased these when asked to say them word for word, so they don't go
// through it. One start line (after the cue word) and one end line at most.
var gopodSnippetLast = struct {
	sync.Mutex
	line map[int]string
}{line: map[int]string{}}

// gopodScriptedSnippet picks this Brobot's line for one place in the reply
// ("start" or "end"): the first entry that matches - its topic pack loaded,
// bingo_not_played, thanks (a thank-you), reply_mentions one of its words in
// reply, or any - and is his
// own or shared. "" = none.
func gopodScriptedSnippet(esn, userMessage, at, reply string) string {
	// 2026-10-11 (live 23:2x: the real weather was given to gemma2:2b, and it
	// still said "a high of 63°F" and "a scorcher"): a weather question's reply
	// starts with the golden weather line itself, said by the code word for word.
	// 2026-10-18: not when a manager said it (gopod_managers.go) - the Brobots are released from it.
	if at == "start" && (gopodHasWholeWords(strings.ToLower(userMessage), "weather") || gopodNativeTakeIs(gopodBrobotName(esn), "weather", userMessage)) && !gopodManagerTookWeather() {
		// 2026-10-22: the multiple responses template - "report" first or alone,
		// "confirm" next (weather_lines); the golden line when they're empty.
		kind := "report"
		if gopodIsFollowingUp(esn) {
			kind = "confirm"
		}
		if line := gopodWeatherReportLine(esn, gopodWeatherPlace(userMessage), kind); line != "" {
			return line
		}
		if line, _ := gopodWeatherSayAt(esn, gopodWeatherPlace(userMessage)); line != "" {
			return line
		}
	}
	// 2026-10-17 (operator: "Don't speak or highlight search results"): what
	// the search found shows in its own bubble (gopodSearchFacts), not in the
	// reply; system_search.line only when there's no results_role.
	if at == "start" && gopodBrobotsSettings.SystemSearch.ResultsRole == "" {
		if line := gopodSearchLine(userMessage); line != "" {
			return line
		}
	}
	// 2026-10-06: a point's exact line comes first (gopodSayLine, gopod_value_points.go).
	// 2026-10-22: it's the answer chain's colour:value_point part now; with
	// one, no flavour start line (still one start line at most).
	if at == "start" && gopodSayLine(esn) != "" {
		return ""
	}
	loaded := map[string]bool{}
	for _, packet := range loadScenarioPacketsForMessage(userMessage) {
		loaded[packet.Name] = true
	}
	bingoQuestion := !(gopodBingoPlayed.Load() || gopodBingoRunInMemory()) && gopodIsBingoQuestion(userMessage)
	lowerReply := strings.ToLower(reply)
	for n, s := range gopodBrobotsSettings.ScriptedSnippets.Lines {
		where := s.At
		if where == "" {
			where = "start"
		}
		if where != at || (s.Brobot != "" && s.Brobot != gopodBrobotName(esn)) {
			continue
		}
		// 2026-10-06: "reminder" - the next uncovered topic's line, worked out
		// by the value-point list (gopod_value_points.go).
		if s.When == "reminder" {
			// The chance first: a reminder not said isn't counted as named.
			if s.Chance > 0 && rand.Float64() >= s.Chance {
				continue
			}
			if line := gopodReminderLine(userMessage); line != "" {
				return line
			}
			continue
		}
		match := loaded[s.When] || (s.When == "bingo_not_played" && bingoQuestion) ||
			(s.When == "thanks" && gopodIsThanks(userMessage)) || s.When == "any"
		if s.When == "reply_mentions" {
			for _, w := range s.Words {
				match = match || gopodHasWholeWords(lowerReply, strings.ToLower(w))
			}
		}
		// chance: how often a matching entry is used (0 or left out = always).
		if match && (s.Chance <= 0 || rand.Float64() < s.Chance) {
			return gopodPickSnippetLine(n, s.Lines)
		}
	}
	return ""
}

// gopodPickSnippetLine picks one line of an entry, never the same one twice in a row.
func gopodPickSnippetLine(entry int, lines []string) string {
	gopodSnippetLast.Lock()
	defer gopodSnippetLast.Unlock()
	pick := lines[rand.Intn(len(lines))]
	for len(lines) > 1 && pick == gopodSnippetLast.line[entry] {
		pick = lines[rand.Intn(len(lines))]
	}
	gopodSnippetLast.line[entry] = pick
	return pick
}

// gopodDropSnippetCopy drops the LLM's own copy of a snippet, so it isn't said twice.
func gopodDropSnippetCopy(snippet, reply string) string {
	for _, s := range []string{snippet, strings.TrimRight(snippet, "!.?")} {
		reply = regexp.MustCompile(`(?i)`+regexp.QuoteMeta(s)+`[!.?]*`).ReplaceAllString(reply, "")
	}
	return strings.TrimSpace(regexp.MustCompile(`[ \t]{2,}`).ReplaceAllString(reply, " "))
}

// gopodStitchSnippet puts the snippet at the front of the reply, right after
// its cue word ("confused___ No - I'm on GOPOD! Ugh, TikTok..."), so the cue
// still plays first and the snippet is shown, spoken and remembered as his.
func gopodStitchSnippet(snippet, reply string) string {
	if snippet == "" {
		return reply
	}
	// The reply's own cue fixes first ("thinking__" -> "thinking___"), so the
	// snippet lands after the cue, not in front of it.
	reply = gopodDropSnippetCopy(snippet, gopodPrepareReply(reply))
	if word, rest := gopodBareLeadingCue(reply); word != "" {
		reply = word + gopodAnimationMarker + " " + rest
	}
	if loc := gopodAnimationMarkerPattern.FindStringIndex(reply); loc != nil &&
		strings.TrimFunc(reply[:loc[0]], func(r rune) bool { return !unicode.IsLetter(r) }) == "" {
		return reply[:loc[1]] + snippet + gopodThoughtGap() + reply[loc[1]:]
	}
	return snippet + gopodThoughtGap() + reply
}

// gopodStitchEndSnippet puts the snippet at the end of the reply.
func gopodStitchEndSnippet(snippet, reply string) string {
	if snippet == "" {
		return reply
	}
	return gopodDropSnippetCopy(snippet, reply) + gopodThoughtGap() + snippet
}

// Thoughts as paragraphs (operator 2026-10-15: "Go with: reply still has
// natural thoughts * the weather line * the value-point line * the math steps *
// 'Hey, that reminds me...' * splits/closes a thought with an emoji"; "Go with
// optional paragraph look * thought bubble = paragraph"). With
// live_split.thought_lines, each thought starts its own line: the code's own
// lines (the start and end snippets) and gemma's thoughts, which she closes with
// an emoji. A line is never joined with the next (gopodSplitPieces works line by
// line), so a run of pieces stays inside its thought.
func gopodThoughtGap() string {
	if gopodBrobotsSettings.LiveSplit.ThoughtLines {
		return "\n"
	}
	return " "
}

// gopodThoughtBreaks: a new line after an emoji that closes a thought - one
// followed by a new sentence (a capital or a quote).
var gopodThoughtEmojiRe = regexp.MustCompile(`((?:[\x{1FA00}-\x{1FAFF}\x{1F600}-\x{1F64F}\x{1F300}-\x{1F5FF}\x{1F680}-\x{1F6FF}\x{1F900}-\x{1F9FF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}\x{2B50}\x{FE0F}\x{200D}])+)[ \t]+(\p{Lu}|["'\x{201C}])`)

func gopodThoughtBreaks(text string) string {
	if !gopodBrobotsSettings.LiveSplit.ThoughtLines {
		return text
	}
	for i := 0; i < 2; i++ {
		text = gopodThoughtEmojiRe.ReplaceAllString(text, "${1}\n${2}")
	}
	return text
}

// gopodIsBingoQuestion: the question is about the Bingo test (bingo.question_word),
// not one of bingo.not_a_test_phrases (operator 2026-10-06: "Chocolate Bingo"
// is the party game, not the test - it no longer gets "Say Bingo Test Live!").
func gopodIsBingoQuestion(userMessage string) bool {
	lower := strings.ToLower(userMessage)
	for _, phrase := range gopodBrobotsSettings.Bingo.NotATest {
		if gopodHasWholeWords(lower, strings.ToLower(phrase)) {
			return false
		}
	}
	return strings.Contains(lower, strings.ToLower(gopodBrobotsSettings.Bingo.QuestionWord))
}

// No web addresses or links in a reply (web_links in gopod_brobots_settings.json;
// operator 2026-10-06: "no urls"). gemma2:2b made up addresses and handles
// ("gc.crrshn8r.com"), and through the shared chat the other Brobot repeated
// them. A link or bare address is taken out; one holding keep_with keeps only
// its bare address ("https://www.crushn8r.ca/news" -> "crushn8r.ca"), which the
// pronunciation registry says.
var gopodWebLinkPattern = regexp.MustCompile(`(?i)(?:https?://|www\.)[^\s]+|\b[a-z0-9-]+(?:\.[a-z0-9-]+)*\.(?:com|net|org|ca|io|co|tv|app|ly|me|gg|us|uk|info|biz)\b(?:/[^\s]*)?`)
var gopodLinkTailPattern = regexp.MustCompile(`[.,!?;:)\]]+$`)

func gopodDropWebLinks(text string) string {
	keep := strings.ToLower(gopodBrobotsSettings.WebLinks.KeepWith)
	out := gopodWebLinkPattern.ReplaceAllStringFunc(text, func(link string) string {
		tail := gopodLinkTailPattern.FindString(link)
		host := regexp.MustCompile(`(?i)^(?:https?://)?(?:www\.)?`).ReplaceAllString(strings.TrimSuffix(link, tail), "")
		if i := strings.IndexAny(host, "/?#"); i >= 0 {
			host = host[:i]
		}
		// Only the address itself counts: "tiktok.com/@crushn8r" is not ours.
		if keep == "" || !strings.Contains(strings.ToLower(host), keep) {
			return tail
		}
		return host + tail
	})
	return regexp.MustCompile(`[ \t]{2,}`).ReplaceAllString(out, " ")
}

func loadCanonicalBrobotPrompt(userMessage string) (string, int, bool) {
	responsePromptBytes, responseErr := os.ReadFile(canonicalBrobotResponsePromptPath)
	identityPromptBytes, identityErr := os.ReadFile(canonicalBrobotIdentityPromptPath)
	corePromptBytes, coreErr := os.ReadFile(canonicalGOPODCorePromptPath)
	if responseErr == nil && identityErr == nil && coreErr == nil {
		promptBytes := make([]byte, 0, len(responsePromptBytes)+1+len(identityPromptBytes)+1+len(corePromptBytes))
		// The reply rules name the marker and the placeholder word as
		// $ANIMATION_MARKER / $PLACEHOLDER_WORD; their values come from
		// gopod_brobots_settings.json, so the prompt can't drift from the code.
		// 2026-10-05: the reply shape is no longer first. It goes last, after
		// "who I am" (gopodReplyShapeSection, CreateAIReq), with that brother's own
		// examples - general to specific, the most specific nearest the question.
		promptBytes = appendPromptSection(promptBytes, identityPromptBytes)
		promptBytes = appendPromptSection(promptBytes, corePromptBytes)
		// GOPOD 2026-09-29: session memory now reaches the model as one shared
		// Brobots chat (gopodSharedBrobotChat, used by CreateAIReq), not as a
		// summary block here - the model copied summary lines word for word.
		// 2026-10-05: this question's topic packs and notes are no longer here;
		// they go right after "who I am" (gopodQuestionFacts, CreateAIReq).
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

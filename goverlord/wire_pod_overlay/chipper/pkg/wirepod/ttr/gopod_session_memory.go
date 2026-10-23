package wirepod_ttr

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

const maxGOPODRecentSessionExchanges = 18
const gopodSessionArchiveCadence = 9
const maxGOPODSessionMemoryPromptBytes = 1800

type gopodSessionExchange struct {
	ExchangeID   int       `json:"exchange_id"`
	Timestamp    time.Time `json:"timestamp"`
	ESN          string    `json:"esn"`
	UserMessage  string    `json:"user_message"`
	RawResponse  string    `json:"raw_response"`
	SpokenOutput string    `json:"spoken_output"`
	Scenario     string    `json:"scenario"`
	ScenarioFile string    `json:"scenario_file"`
}

type gopodSessionSummary struct {
	TotalExchanges int `json:"total_exchanges"`
	ArchivePart    int `json:"archive_part"`
}

func gopodSessionDir() string {
	if dir := strings.TrimSpace(os.Getenv("GOPOD_SESSION_DIR")); dir != "" {
		return dir
	}
	return filepath.Join(gopodChipperRoot(), "sessions")
}

func gopodCurrentSessionPath() string {
	return filepath.Join(gopodSessionDir(), "current_session.jsonl")
}

func gopodCurrentSummaryPath() string {
	return filepath.Join(gopodSessionDir(), "current_summary.txt")
}

func gopodArchiveDir() string {
	return filepath.Join(gopodSessionDir(), "archives")
}

func appendGOPODSessionExchange(exchange gopodSessionExchange) error {
	if err := os.MkdirAll(gopodArchiveDir(), 0o755); err != nil {
		return err
	}
	summary, err := loadGOPODSessionSummary()
	if err != nil {
		return err
	}
	summary.TotalExchanges++
	exchange.ExchangeID = summary.TotalExchanges
	if exchange.Timestamp.IsZero() {
		exchange.Timestamp = time.Now().UTC()
	}

	recent, err := loadRecentGOPODSessionExchanges()
	if err != nil {
		return err
	}
	recent = append(recent, exchange)
	if summary.TotalExchanges%gopodSessionArchiveCadence == 0 {
		summary.ArchivePart++
		if err := archiveGOPODSessionChunk(recent, summary.ArchivePart, exchange.Timestamp); err != nil {
			return err
		}
	}
	if len(recent) > maxGOPODRecentSessionExchanges {
		recent = recent[len(recent)-maxGOPODRecentSessionExchanges:]
	}
	if err := writeGOPODSessionExchanges(gopodCurrentSessionPath(), recent); err != nil {
		return err
	}
	return writeGOPODSessionSummary(summary)
}

func loadGOPODSessionSummary() (gopodSessionSummary, error) {
	data, err := os.ReadFile(gopodCurrentSummaryPath())
	if os.IsNotExist(err) {
		return gopodSessionSummary{}, nil
	}
	if err != nil {
		return gopodSessionSummary{}, err
	}
	var summary gopodSessionSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return gopodSessionSummary{}, err
	}
	return summary, nil
}

func writeGOPODSessionSummary(summary gopodSessionSummary) error {
	if err := os.MkdirAll(gopodSessionDir(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(gopodCurrentSummaryPath(), data, 0o644)
}

func loadRecentGOPODSessionExchanges() ([]gopodSessionExchange, error) {
	file, err := os.Open(gopodCurrentSessionPath())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var exchanges []gopodSessionExchange
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var exchange gopodSessionExchange
		if err := json.Unmarshal([]byte(line), &exchange); err != nil {
			return nil, err
		}
		exchanges = append(exchanges, exchange)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(exchanges) > maxGOPODRecentSessionExchanges {
		exchanges = exchanges[len(exchanges)-maxGOPODRecentSessionExchanges:]
	}
	return exchanges, nil
}

func writeGOPODSessionExchanges(path string, exchanges []gopodSessionExchange) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var builder strings.Builder
	for _, exchange := range exchanges {
		line, err := json.Marshal(exchange)
		if err != nil {
			return err
		}
		builder.Write(line)
		builder.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(builder.String()), 0o644)
}

func archiveGOPODSessionChunk(exchanges []gopodSessionExchange, part int, timestamp time.Time) error {
	if len(exchanges) > gopodSessionArchiveCadence {
		exchanges = exchanges[len(exchanges)-gopodSessionArchiveCadence:]
	}
	if err := os.MkdirAll(gopodArchiveDir(), 0o755); err != nil {
		return err
	}
	name := fmt.Sprintf("session_%s_part%03d.jsonl", timestamp.UTC().Format("20060102_150405"), part)
	return writeGOPODSessionExchanges(filepath.Join(gopodArchiveDir(), name), exchanges)
}

func loadGOPODSessionMemoryPrompt() string {
	recent, err := loadRecentGOPODSessionExchanges()
	if err != nil {
		return ""
	}
	return buildGOPODSessionMemoryPrompt(recent)
}

func buildGOPODSessionMemoryPrompt(exchanges []gopodSessionExchange) string {
	if len(exchanges) == 0 {
		return ""
	}
	if len(exchanges) > 6 {
		exchanges = exchanges[len(exchanges)-6:]
	}
	var builder strings.Builder
	// Header text is derived from system_speech_content.json - see
	// gopod_system_speech_content.go. The per-exchange lines below stay in
	// Go: they're generated from live session data (loop + byte-budget
	// truncation), not fixed content, so externalizing them would move
	// control flow into data rather than remove it.
	builder.WriteString(loadedGOPODSystemSpeechContent.SessionMemoryHeader)
	for _, exchange := range exchanges {
		// GOPOD 2026-09-29: name which brother answered (by ESN), so each Brobot
		// can pick up the chat instead of repeating its brother's answer.
		who := gopodBrobotName(exchange.ESN)
		sm := gopodBrobotsSettings.SessionMemory
		line := strings.NewReplacer("{who}", who, "{asked}", compactGOPODMemoryText(exchange.UserMessage), "{answered}", compactGOPODMemoryText(exchange.SpokenOutput)).Replace(sm.Line)
		if strings.TrimSpace(exchange.Scenario) != "" {
			line += strings.ReplaceAll(sm.Scenario, "{scenario}", compactGOPODMemoryText(exchange.Scenario))
		}
		builder.WriteString(line)
		builder.WriteByte('\n')
		if builder.Len() >= maxGOPODSessionMemoryPromptBytes {
			break
		}
	}
	packet := builder.String()
	if len([]byte(packet)) <= maxGOPODSessionMemoryPromptBytes {
		return packet
	}
	return string([]byte(packet)[:maxGOPODSessionMemoryPromptBytes])
}

func compactGOPODMemoryText(text string) string {
	// GOPOD 2026-09-29: memory is read back by the model as an example of how to
	// talk, so it holds plain words only - no "thinking..." cues, no |||C|||
	// markers, no emoji, and no "..." (the model copies all of these into new
	// replies, and a copied "word..." splits the spoken reply).
	text = gopodMemoryMarkerPattern.ReplaceAllString(text, " ")
	text = gopodMemoryActionCuePattern.ReplaceAllStringFunc(text, func(cue string) string {
		word := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(cue), "..."))
		if strings.EqualFold(word, "actionParameter") {
			return " "
		}
		if _, ok := knownAnimationParameter(word); ok {
			return " "
		}
		if _, ok := knownGOPODActionAlias(word); ok {
			return " "
		}
		return cue
	})
	text = removeEmojis(text)
	text = strings.ReplaceAll(text, "…", " ")
	text = gopodMemoryDotsPattern.ReplaceAllString(text, " ")
	text = strings.Join(strings.Fields(text), " ")
	text = strings.TrimSpace(text)
	text = strings.Trim(text, ".")
	if text == "" {
		return gopodBrobotsSettings.SessionMemory.EmptyText
	}
	const maxRunes = 120
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	// Cut at a word boundary, with no trailing "..." for the model to copy.
	cut := string(runes[:maxRunes])
	if i := strings.LastIndex(cut, " "); i > 0 {
		cut = cut[:i]
	}
	return strings.TrimRight(cut, " ,;:-")
}

var (
	gopodMemoryMarkerPattern    = regexp.MustCompile(`\|\|\|[01C]\|\|\|`)
	gopodMemoryActionCuePattern = regexp.MustCompile(`(?:^|\s)[A-Za-z]+\s*\.\.\.`)
	gopodMemoryDotsPattern      = regexp.MustCompile(`\.{2,}`)
)

// gopodMemoryNameLabelLinePattern: a line of a saved answer that opens with a
// Brobot's name and a colon ("🚀  Brobot 2: What's up?"). That's the model
// writing a line for someone else, not chat content, so the shared chat skips it
// (operator 2026-10-05: "when brobots review chat history for answering, skip
// those types of lines"). One such line in memory was copied into most of the
// replies after it (replay on gemma2:2b: 7 of 12, vs 2 of 24 with none).
var gopodMemoryNameLabelLinePattern = regexp.MustCompile(`(?m)^[^\p{L}\p{N}\n]*Brobot\s*\d\s*:.*$`)

// gopodDropNameLabelLines removes those lines from a saved answer.
func gopodDropNameLabelLines(text string) string {
	return gopodMemoryNameLabelLinePattern.ReplaceAllString(text, "")
}

func recordCompletedGOPODSessionExchange(esn, userMessage string, result gopodStreamResult, timestamp time.Time) {
	packet, _ := loadScenarioPacketForMessage(userMessage)
	exchange := gopodSessionExchange{
		Timestamp:    timestamp,
		ESN:          esn,
		UserMessage:  userMessage,
		RawResponse:  result.Raw,
		SpokenOutput: result.FilteredPacket,
		Scenario:     packet.Name,
		ScenarioFile: packet.File,
	}
	if err := appendGOPODSessionExchange(exchange); err != nil {
		logger.Println("GOPOD session memory write failed: " + err.Error())
	}
	gopodRecordExchange(esn, userMessage, result)
}

// The Brobots' robot serials (ESNs).
const (
	gopodBrobot1ESN = "0dd1b9e9"
	gopodBrobot2ESN = "0dd1d8bf"
)

// gopodBrobotName maps a robot ESN to its Brobot name. Unknown ESNs stay "Brobot".
// GopodBrobotOrder (2026-10-10, operator: "upon detection of more than 1 robot,
// list Brobot 1 first, then Brobot 2"): a robot's place in a list - Brobot 1,
// Brobot 2, then any other robot.
func GopodBrobotOrder(esn string) int {
	switch gopodBrobotName(strings.ToLower(esn)) {
	case "Brobot 1":
		return 1
	case "Brobot 2":
		return 2
	}
	return 3
}

// GopodBrobotName: "Brobot 1", "Brobot 2", or "Brobot" for another serial.
func GopodBrobotName(esn string) string { return gopodBrobotName(esn) }

func gopodBrobotName(esn string) string {
	switch strings.TrimSpace(esn) {
	case gopodBrobot1ESN:
		return "Brobot 1"
	case gopodBrobot2ESN:
		return "Brobot 2"
	}
	return "Brobot"
}

// The shared chat's limits and notes below are shared_chat (and bingo) in
// gopod_brobots_settings.json.
//
// gopodSharedBrobotChatExchanges is how many recent exchanges (both robots) come
// back as the shared chat. Operator 2026-10-01: "cap at 13 most recent
// questions+ snapshot bingo experiences, after that brobots say they don't
// remember".
var gopodSharedBrobotChatExchanges = gopodBrobotsSettings.SharedChat.MaxQuestions

// gopodSharedBrobotChatMaxAge: exchanges older than this are left out of the
// shared chat, and so is anything from before this Wire-Pod start - so a restart
// or a break gives a fresh chat, and an old made-up detail (an invented name)
// isn't fed back forever. The session log itself keeps everything.
var gopodSharedBrobotChatMaxAge = time.Duration(gopodBrobotsSettings.SharedChat.MaxMinutes) * time.Minute

var gopodSharedBrobotChatStart = time.Now()

// gopodSharedBrobotAnswerMaxRunes bounds each remembered answer.
var gopodSharedBrobotAnswerMaxRunes = gopodBrobotsSettings.SharedChat.MaxAnswerChars

// gopodTagForBrobot marks which brother a question went to: "[to Brobot 1] ...".
func gopodTagForBrobot(esn, text string) string {
	return "[to " + gopodBrobotName(esn) + "] " + gopodNormalizeGOPODName(text)
}

// gopodForgottenNote opens the shared chat when older exchanges since this
// Wire-Pod start were left out (past the 13 newest, or older than 120 minutes).
var gopodForgottenNote = gopodBrobotsSettings.SharedChat.ForgottenNote

// gopodBingoRunMarker: a performed line holding this starts a Bingo test run
// ("Bingo Test Live", "Bingo Test Begins. Round 1!"); the run is every performed
// line from there until the next question, or until the feed goes quiet for
// gopodBingoRunGap (other lines after it, such as go search, are kept).
var gopodBingoRunMarker = gopodBrobotsSettings.Bingo.RunMarker

var gopodBingoRunGap = gopodSecondsToDuration(gopodBrobotsSettings.Bingo.RunGapSeconds)

// gopodSharedBrobotChat returns the recent exchanges of both robots as one chat:
// the tagged question, then the answer as the model wrote it (split markers
// removed). It also carries what the brothers performed in between, from
// Wire-Pod's chat feed (gopod_chat_feed.go):
//   - The Bingo test is scripted, so every run is the same. Its lines are not
//     carried; one snapshot note stands for them - however many runs, remembered
//     as the same single run - since this Wire-Pod start and within
//     gopodSharedBrobotChatMaxAge, however many questions came after
//     (gopodBingoSnapshotNote, plus the asking brother's take, gopodAddBingoTake).
//     Operator 2026-10-01: "can the brobots just respond as if the same single
//     run, since its a scripted performance?"
//   - Other performed lines (go search answers) come as one note per run of
//     lines (gopodSharedFeedNote), from the oldest remembered exchange on.
func gopodSharedBrobotChat() []openai.ChatCompletionMessage {
	cutoff := time.Now().Add(-gopodSharedBrobotChatMaxAge)
	if gopodSharedBrobotChatStart.After(cutoff) {
		cutoff = gopodSharedBrobotChatStart
	}
	loaded, _ := loadRecentGOPODSessionExchanges()
	var fresh []gopodSessionExchange
	forgotten := false
	for _, ex := range loaded {
		if ex.Timestamp.After(cutoff) {
			fresh = append(fresh, ex)
		} else if ex.Timestamp.After(gopodSharedBrobotChatStart) {
			forgotten = true
		}
	}
	recent := fresh
	if len(recent) > gopodSharedBrobotChatExchanges {
		recent = recent[len(recent)-gopodSharedBrobotChatExchanges:]
		forgotten = true
	}

	// Split the performed lines into Bingo runs and the rest.
	sawBingo := false
	var lastBingo time.Time
	var other []gopodChatFeedEvent
	inRun := false
	ei := 0
	var prev time.Time
	for _, e := range gopodChatFeedSince(cutoff) {
		at := time.UnixMilli(e.TimeMS)
		for ei < len(fresh) && fresh[ei].Timestamp.Before(at) {
			inRun = false
			ei++
		}
		if at.Sub(prev) >= gopodBingoRunGap {
			inRun = false
		}
		prev = at
		if !inRun && strings.Contains(e.Text, gopodBingoRunMarker) {
			inRun = true
			sawBingo = true
			lastBingo = at
		}
		if !inRun {
			other = append(other, e)
		}
	}
	feedFrom := cutoff
	if len(recent) > 0 && recent[0].Timestamp.After(feedFrom) {
		feedFrom = recent[0].Timestamp
	}
	var feed []gopodChatFeedEvent
	for _, e := range other {
		// The song's countdown and Bingo ball lines are page-only, never remembered;
		// so are the chapter titles, topic pictures, reporter voice files and the
		// reporter hosts' lines (2026-10-06).
		if gopodPageOnlyRole(e.Who) {
			continue
		}
		if time.UnixMilli(e.TimeMS).After(feedFrom) {
			feed = append(feed, e)
		}
	}
	if len(feed) > gopodSharedFeedMaxLines {
		feed = feed[len(feed)-gopodSharedFeedMaxLines:]
	}

	var chat []openai.ChatCompletionMessage
	if forgotten {
		chat = append(chat, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: gopodForgottenNote})
	}
	bingoPlaced := !sawBingo
	placeBingo := func() {
		if !bingoPlaced {
			chat = append(chat, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: gopodBingoSnapshotNote})
			bingoPlaced = true
		}
	}
	if len(recent) == 0 || lastBingo.Before(recent[0].Timestamp) {
		placeBingo()
	}
	var block []gopodChatFeedEvent
	flush := func() {
		if len(block) > 0 {
			chat = append(chat, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: gopodSharedFeedNote(block)})
			block = nil
		}
	}
	fi := 0
	for _, ex := range recent {
		for fi < len(feed) && time.UnixMilli(feed[fi].TimeMS).Before(ex.Timestamp) {
			block = append(block, feed[fi])
			fi++
		}
		flush()
		if lastBingo.Before(ex.Timestamp) {
			placeBingo()
		}
		// Lines labelled with a Brobot's name are skipped (gopodDropNameLabelLines),
		// and the code's scripted lines are taken out (gopodDropScriptedLines,
		// 2026-10-06) so the LLM never learns them to copy.
		// Each saved answer also gets its owner's name back (gopodFixSelfName), so an
		// old "It's Brobot 2" from Brobot 1 isn't copied into the next reply.
		// The cue goes back to the "happy___" form the prompt asks for
		// (gopodInternalToAnimationMarker), so memory doesn't teach the old "...".
		answer := strings.TrimSpace(gopodInternalToAnimationMarker(gopodSpellNames(gopodFixSelfName(ex.ESN, gopodCorrectNames(gopodMemoryMarkerPattern.ReplaceAllString(gopodPlaceholderWordPattern.ReplaceAllString(gopodDropScriptedLines(gopodDropNameLabelLines(ex.RawResponse)), ""), " "))))))
		if answer == "" {
			continue
		}
		if runes := []rune(answer); len(runes) > gopodSharedBrobotAnswerMaxRunes {
			answer = string(runes[:gopodSharedBrobotAnswerMaxRunes])
		}
		chat = append(chat,
			openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: gopodTagForBrobot(ex.ESN, gopodSpellNames(ex.UserMessage))},
			openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: answer},
		)
	}
	block = append(block, feed[fi:]...)
	flush()
	placeBingo()
	return chat
}

// gopodBingoSnapshotNote stands for the Bingo test in the chat window, however
// many times it ran; the asking brother's own take is added to it
// (gopodAddBingoTake). The text is bingo.snapshot_note in gopod_brobots_settings.json.
var gopodBingoSnapshotNote = gopodBrobotsSettings.Bingo.SnapshotNote

// gopodSharedFeedMaxLines bounds the other performed lines carried in the shared
// chat.
var gopodSharedFeedMaxLines = gopodBrobotsSettings.SharedChat.MaxPerformedLines

// gopodSharedFeedLineMaxRunes bounds each performed line.
var gopodSharedFeedLineMaxRunes = gopodBrobotsSettings.SharedChat.MaxPerformedLineChars

// gopodSharedFeedNotePrefix opens every performed-lines note.
var gopodSharedFeedNotePrefix = gopodBrobotsSettings.SharedChat.PerformedNote

// gopodSharedFeedNote is one note for a run of performed lines - scripted, not a
// chat, so the model reads it as something that happened, not a question to answer.
func gopodSharedFeedNote(lines []gopodChatFeedEvent) string {
	var b strings.Builder
	b.WriteString(gopodSharedFeedNotePrefix)
	for _, e := range lines {
		text := strings.Join(strings.Fields(strings.ReplaceAll(e.Text, gopodBrobotsSettings.ScriptedSplit.SegmentMarker, " ")), " ")
		if runes := []rune(text); len(runes) > gopodSharedFeedLineMaxRunes {
			text = string(runes[:gopodSharedFeedLineMaxRunes])
		}
		b.WriteString("\n" + e.Who + ": " + text)
	}
	return b.String()
}

// gopodPageOnlyRole: a feed line the page shows (or plays) that the Brobots
// never remember.
func gopodPageOnlyRole(who string) bool {
	s := gopodBrobotsSettings
	for _, role := range []string{s.Countdown.Role, s.BingoBall.Role, s.InterviewChapters.Role, s.TopicPictures.Role, s.ReporterVoices.Role, s.ValuePoints.ScoreRole, s.SignOff.SystemRole, s.ReporterVoices.SongAudioRole, s.ReporterVoices.NarratorRole} {
		if role != "" && who == role {
			return true
		}
	}
	_, host := s.ReporterVoices.Hosts[who]
	return host
}

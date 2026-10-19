package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGOPODDiagnosticsPipelineSeparatesMarkedFilteredAndSentPackets(t *testing.T) {
	raw := "\x1b[31m[giggles]\x1b[0m... We made it! |||C||| {{[thinking]}}... Next proof."

	result := buildGOPODStreamResult(raw)

	wantMarked := "|||0||| " + raw + " |||1|||"
	if result.Raw != raw {
		t.Fatalf("Raw = %q, want untouched %q", result.Raw, raw)
	}
	if result.MarkedPacket != wantMarked {
		t.Fatalf("MarkedPacket = %q, want %q", result.MarkedPacket, wantMarked)
	}
	wantFiltered := "happy... We made it! |||C||| thinking... Next proof."
	if result.FilteredPacket != wantFiltered {
		t.Fatalf("FilteredPacket = %q, want %q", result.FilteredPacket, wantFiltered)
	}
	wantChunks := []string{"happy... We made it!", "thinking... Next proof."}
	if len(result.AudibleChunks) != len(wantChunks) {
		t.Fatalf("AudibleChunks = %#v, want %#v", result.AudibleChunks, wantChunks)
	}
	for i := range wantChunks {
		if result.AudibleChunks[i] != wantChunks[i] {
			t.Fatalf("AudibleChunks[%d] = %q, want %q", i, result.AudibleChunks[i], wantChunks[i])
		}
	}
	if got := commandPacketFromFilteredChunk(result.AudibleChunks[0]); got != "{{playAnimationWI||happy}}... We made it!" {
		t.Fatalf("commandPacketFromFilteredChunk = %q", got)
	}
}

func TestLoadCanonicalBrobotPromptUsesSplitPromptFilesInOrder(t *testing.T) {
	responseBytes, err := os.ReadFile(canonicalBrobotResponsePromptPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", canonicalBrobotResponsePromptPath, err)
	}
	// The loader fills the settings' tokens ($PLACEHOLDER_THOUGHT, ...) in.
	responseBytes = []byte(gopodFillPromptTokens(string(responseBytes)))
	brobotBytes, err := os.ReadFile(canonicalBrobotIdentityPromptPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", canonicalBrobotIdentityPromptPath, err)
	}
	coreBytes, err := os.ReadFile(canonicalGOPODCorePromptPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", canonicalGOPODCorePromptPath, err)
	}

	got, gotBytes, usedOpenAIPrompt := loadCanonicalBrobotPrompt("")
	// 2026-10-05: the reply shape is no longer in the foundation; it comes last,
	// per brother (gopodReplyShapeSection, TestCreateAIReqPromptOrderGeneralToSpecific).
	_ = responseBytes
	wantBytesRaw := append([]byte{}, brobotBytes...)
	if len(wantBytesRaw) > 0 && wantBytesRaw[len(wantBytesRaw)-1] != '\n' {
		wantBytesRaw = append(wantBytesRaw, '\n')
	}
	wantBytesRaw = append(wantBytesRaw, coreBytes...)
	want := string(wantBytesRaw)

	if got != want {
		t.Fatalf("loadCanonicalBrobotPrompt content did not match ordered foundation prompt")
	}
	if gotBytes != len(wantBytesRaw) {
		t.Fatalf("loadCanonicalBrobotPrompt bytes = %d, want %d", gotBytes, len(wantBytesRaw))
	}
	if usedOpenAIPrompt {
		t.Fatal("loadCanonicalBrobotPrompt used OpenAIPrompt fallback")
	}
}

func TestCreateAIReqLoadsOnlyCurrentMatchingScenarioPacket(t *testing.T) {
	gopodBingoPlayed.Store(false) // no Bingo test played yet: the Bingo pitch pack loads
	packets := map[string]string{
		"bingo":     mustReadScenarioPacket(t, "bingo.txt"),
		"nightclub": mustReadScenarioPacket(t, "nightclub.txt"),
	}

	tests := []struct {
		name       string
		userText   string
		wantPacket string
	}{
		{
			name:       "bingo trigger",
			userText:   "Tell me about bingo",
			wantPacket: "bingo",
		},
		{
			name:       "nightclub trigger",
			userText:   "What about comedy clubs?",
			wantPacket: "nightclub",
		},
		{
			name:     "no trigger",
			userText: "What are robots?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateAIReq(tt.userText, "test-esn", false, true)
			if len(req.Messages) < 2 {
				t.Fatalf("CreateAIReq messages length = %d, want system and user messages", len(req.Messages))
			}
			systemPrompt := req.Messages[0].Content
			for packetName, packetText := range packets {
				hasPacket := strings.Contains(systemPrompt, packetText)
				if packetName == tt.wantPacket {
					if !hasPacket {
						t.Fatalf("system prompt missing %s scenario packet", packetName)
					}
					continue
				}
				if hasPacket {
					t.Fatalf("system prompt included non-matching %s scenario packet", packetName)
				}
			}
		})
	}
}

// 2026-10-05: general to specific - shared persona, GOPOD core, who this
// brother is, this question's facts (scenario packet), then his reply shape, then
// the final check
// (BROBOTS_LAYERED_PROMPT_002.md).
func TestCreateAIReqPromptOrderGeneralToSpecific(t *testing.T) {
	gopodBingoPlayed.Store(false) // no Bingo test played yet: the Bingo pitch pack loads
	brobotBytes, err := os.ReadFile(canonicalBrobotIdentityPromptPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", canonicalBrobotIdentityPromptPath, err)
	}
	coreBytes, err := os.ReadFile(canonicalGOPODCorePromptPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", canonicalGOPODCorePromptPath, err)
	}
	bingoPacket := mustReadScenarioPacket(t, "bingo.txt")
	shape := gopodReplyShapeSection("Brobot 1")

	req := CreateAIReq("Tell me about bingo", gopodBrobot1ESN, false, true)
	systemPrompt := req.Messages[0].Content
	brobotIndex := strings.Index(systemPrompt, string(brobotBytes))
	coreIndex := strings.Index(systemPrompt, string(coreBytes))
	scenarioIndex := strings.Index(systemPrompt, bingoPacket)
	esnIndex := strings.Index(systemPrompt, "WHO I AM:\n"+gopodBrobotsSettings.Personas["Brobot 1"])
	if strings.Contains(systemPrompt, gopodBrobot1ESN) || strings.Contains(systemPrompt, gopodBrobot2ESN) {
		t.Fatal("a robot ID (ESN) is in the prompt; it gets said out loud")
	}
	shapeIndex := strings.Index(systemPrompt, shape)
	guardIndex := strings.Index(systemPrompt, gopodFinalOutputGuard)
	if shape == "" || brobotIndex == -1 || coreIndex == -1 || scenarioIndex == -1 || esnIndex == -1 || shapeIndex == -1 || guardIndex == -1 {
		t.Fatalf("system prompt is missing a section")
	}
	if !(brobotIndex < coreIndex && coreIndex < esnIndex && esnIndex < scenarioIndex && scenarioIndex < shapeIndex && shapeIndex < guardIndex) {
		t.Fatalf("prompt order brobot=%d core=%d esn=%d scenario=%d shape=%d guard=%d", brobotIndex, coreIndex, esnIndex, scenarioIndex, shapeIndex, guardIndex)
	}
	if !strings.Contains(shape, gopodBrobotsSettings.ReplyExamples["Brobot 1"]) || strings.Contains(systemPrompt, gopodBrobotsSettings.ReplyExamples["Brobot 2"]) {
		t.Fatal("Brobot 1's prompt should carry his own examples, not his brother's")
	}
}

func TestCreateAIReqScenarioPacketDoesNotStickAcrossTurns(t *testing.T) {
	gopodBingoPlayed.Store(false) // no Bingo test played yet: the Bingo pitch pack loads
	bingoPacket := mustReadScenarioPacket(t, "bingo.txt")
	nightclubPacket := mustReadScenarioPacket(t, "nightclub.txt")

	turns := []struct {
		name      string
		userText  string
		want      string
		forbidden []string
	}{
		{
			name:      "bingo only",
			userText:  "Tell me about bingo",
			want:      bingoPacket,
			forbidden: []string{nightclubPacket},
		},
		{
			name:      "nightclub only after bingo",
			userText:  "What about comedy clubs?",
			want:      nightclubPacket,
			forbidden: []string{bingoPacket},
		},
		{
			name:      "generic has no previous scenario",
			userText:  "What are robots?",
			forbidden: []string{bingoPacket, nightclubPacket},
		},
	}

	for _, turn := range turns {
		t.Run(turn.name, func(t *testing.T) {
			req := CreateAIReq(turn.userText, "test-esn", false, true)
			systemPrompt := req.Messages[0].Content
			if turn.want != "" && !strings.Contains(systemPrompt, turn.want) {
				t.Fatalf("system prompt missing expected scenario packet")
			}
			for _, forbidden := range turn.forbidden {
				if strings.Contains(systemPrompt, forbidden) {
					t.Fatalf("system prompt included sticky or wrong scenario packet")
				}
			}
		})
	}
}

func TestGOPODSessionMemoryAppendsArchivesAndBoundsPromptPacket(t *testing.T) {
	sessionDir := t.TempDir()
	t.Setenv("GOPOD_SESSION_DIR", sessionDir)
	for i := 1; i <= 20; i++ {
		if err := appendGOPODSessionExchange(gopodSessionExchange{
			Timestamp:    time.Date(2026, 6, 24, 15, i, 0, 0, time.UTC),
			ESN:          "test-esn",
			UserMessage:  "User message number " + string(rune('A'+i)),
			RawResponse:  "happy... Raw response number " + string(rune('A'+i)),
			SpokenOutput: "happy... Spoken response number " + string(rune('A'+i)),
		}); err != nil {
			t.Fatalf("appendGOPODSessionExchange(%d) error = %v", i, err)
		}
	}

	recent, err := loadRecentGOPODSessionExchanges()
	if err != nil {
		t.Fatalf("loadRecentGOPODSessionExchanges error = %v", err)
	}
	if len(recent) != 18 {
		t.Fatalf("recent exchanges = %d, want 18", len(recent))
	}
	packet := buildGOPODSessionMemoryPrompt(recent)
	if !strings.Contains(packet, "SESSION MEMORY:") {
		t.Fatalf("memory packet missing header: %q", packet)
	}
	if strings.Contains(packet, "User message number B") {
		t.Fatal("memory packet included exchange older than bounded recent window")
	}
	if len([]byte(packet)) > maxGOPODSessionMemoryPromptBytes {
		t.Fatalf("memory packet bytes = %d, want <= %d", len([]byte(packet)), maxGOPODSessionMemoryPromptBytes)
	}

	archiveMatches, err := filepath.Glob(filepath.Join(sessionDir, "archives", "session_*_part*.jsonl"))
	if err != nil {
		t.Fatalf("Glob archives error = %v", err)
	}
	if len(archiveMatches) < 2 {
		t.Fatalf("archive files = %d, want at least 2 after 20 exchanges", len(archiveMatches))
	}
	for _, archivePath := range archiveMatches {
		info, err := os.Stat(archivePath)
		if err != nil {
			t.Fatalf("Stat(%q) error = %v", archivePath, err)
		}
		if info.Size() == 0 {
			t.Fatalf("archive %q is empty", archivePath)
		}
	}
}

func mustReadScenarioPacket(t *testing.T, name string) string {
	t.Helper()
	for _, p := range gopodTopicRegistry() {
		if p.Name+".txt" == name {
			packetBytes, err := gopodReadTopicFacts(p.File)
			if err != nil {
				t.Fatalf("read %q error = %v", name, err)
			}
			return string(packetBytes)
		}
	}
	t.Fatalf("no topic %q in the chapters folder", name)
	return ""
}

func TestBrobotFilterPacketPreservesMultiWordPhraseHead(t *testing.T) {
	raw := "very happy... Oh yeah, you could call me a sleek one"

	result := buildGOPODStreamResult(raw)

	wantFiltered := "veryHappy... Oh yeah, you could call me a sleek one"
	if result.FilteredPacket != wantFiltered {
		t.Fatalf("FilteredPacket = %q, want %q", result.FilteredPacket, wantFiltered)
	}
	wantChunks := []string{wantFiltered}
	if len(result.AudibleChunks) != len(wantChunks) {
		t.Fatalf("AudibleChunks = %#v, want %#v", result.AudibleChunks, wantChunks)
	}
	for i := range wantChunks {
		if result.AudibleChunks[i] != wantChunks[i] {
			t.Fatalf("AudibleChunks[%d] = %q, want %q", i, result.AudibleChunks[i], wantChunks[i])
		}
	}
	if got := commandPacketFromFilteredChunk(result.AudibleChunks[0]); got != "{{playAnimationWI||veryHappy}}... Oh yeah, you could call me a sleek one" {
		t.Fatalf("commandPacketFromFilteredChunk = %q", got)
	}
}

func TestBrobotFilterPacketPhraseChunkBoundariesUseOnlyAllowlistedHeads(t *testing.T) {
	tests := []struct {
		name         string
		raw          string
		wantFiltered string
		wantChunks   []string
	}{
		{
			name:         "unicode ellipsis normalizes in robot copy path",
			raw:          "happy … unicode delimiter",
			wantFiltered: "happy... unicode delimiter",
			wantChunks:   []string{"happy... unicode delimiter"},
		},
		{
			name:         "happy spaced ellipsis starts phrase chunk",
			raw:          "happy ... speech payload",
			wantFiltered: "happy... speech payload",
			wantChunks:   []string{"happy... speech payload"},
		},
		{
			name:         "confused tight ellipsis starts phrase chunk",
			raw:          "confused... speech payload",
			wantFiltered: "confused... speech payload",
			wantChunks:   []string{"confused... speech payload"},
		},
		{
			name:         "celebrate spaced ellipsis starts phrase chunk",
			raw:          "celebrate ... speech payload",
			wantFiltered: "celebrate... speech payload",
			wantChunks:   []string{"celebrate... speech payload"},
		},
		{
			name:         "bright ellipsis stays inside current speech",
			raw:          "happy... bright... we programmed the cue",
			wantFiltered: "happy... bright... we programmed the cue",
			wantChunks:   []string{"happy... bright... we programmed the cue"},
		},
		{
			name:         "bare spaced ellipsis falls back once",
			raw:          "bare ... payload continues",
			wantFiltered: "thinking... bare... payload continues",
			wantChunks:   []string{"thinking... bare... payload continues"},
		},
		{
			name:         "bare unicode ellipsis falls back once",
			raw:          "bare … payload continues",
			wantFiltered: "thinking... bare... payload continues",
			wantChunks:   []string{"thinking... bare... payload continues"},
		},
		{
			name:         "quoted ellipsis falls back once",
			raw:          "\"...\" payload continues",
			wantFiltered: "thinking... \"...\" payload continues",
			wantChunks:   []string{"thinking... \"...\" payload continues"},
		},
		{
			name:         "orphan payload with internal ellipses creates one fallback chunk",
			raw:          "bare ... one … two \"...\" three",
			wantFiltered: "thinking... bare... one... two \"...\" three",
			wantChunks:   []string{"thinking... bare... one... two \"...\" three"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildGOPODStreamResult(tt.raw)
			if result.Raw != tt.raw {
				t.Fatalf("Raw = %q, want untouched %q", result.Raw, tt.raw)
			}
			if result.FilteredPacket != tt.wantFiltered {
				t.Fatalf("FilteredPacket = %q, want %q", result.FilteredPacket, tt.wantFiltered)
			}
			if len(result.AudibleChunks) != len(tt.wantChunks) {
				t.Fatalf("AudibleChunks = %#v, want %#v", result.AudibleChunks, tt.wantChunks)
			}
			for i := range tt.wantChunks {
				if result.AudibleChunks[i] != tt.wantChunks[i] {
					t.Fatalf("AudibleChunks[%d] = %q, want %q", i, result.AudibleChunks[i], tt.wantChunks[i])
				}
			}
		})
	}
}

func TestCleanRobotSpeechTextRemovesCommandMarkersAndActionPrefixesFromFinalSpeech(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: "{{playAnimationWI||happy}}... happy... Mobility = marketable asset... Car setup travels. |||C|||",
			want:  "Mobility = marketable asset, Car setup travels.",
		},
		{
			input: "{{playAnimationWI||dartingEyes}}... actionParameter: happy Brobots are playful reps.",
			want:  "Brobots are playful reps.",
		},
		{
			input: "{{playAnimationWI||dartingEyes}}... spoken thought Well put, party animal.",
			want:  "Well put, party animal.",
		},
		{
			input: "{{playAnimationWI||dartingEyes}}... thinking Hmm, gotta explain this.",
			want:  "Hmm, gotta explain this.",
		},
	}
	for _, tt := range tests {
		got := cleanRobotSpeechText(tt.input)
		if got != tt.want {
			t.Fatalf("cleanRobotSpeechText(%q) = %q, want %q", tt.input, got, tt.want)
		}
		for _, forbidden := range []string{"{{playAnimationWI", "playAnimationWI", "{{", "}}", "|||", gopodCanonicalChunkMarker, "happy...", "...", "actionParameter", "spoken thought"} {
			if strings.Contains(got, forbidden) {
				t.Fatalf("cleanRobotSpeechText output contains %q: %q", forbidden, got)
			}
		}
	}
}

func TestSaveBrobotResponseDiagnosticsWritesMatchingLogsFiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SAVE_ROBOT_SAY_TEXT_DIR", dir)

	result := gopodStreamResult{
		Raw:            "raw response",
		MarkedPacket:   "|||0||| raw response |||1|||",
		FilteredPacket: "happy... speech",
		AudibleChunks:  []string{"happy... speech"},
	}
	paths, err := saveBrobotResponseDiagnostics("00e20100", result, []string{"{{playAnimationWI||happy}}... speech"}, time.Date(2026, 6, 22, 12, 34, 56, 789000000, time.UTC))
	if err != nil {
		t.Fatalf("saveBrobotResponseDiagnostics error = %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("paths = %#v, want 2 paths", paths)
	}
	for _, path := range paths {
		if filepath.Dir(path) != dir {
			t.Fatalf("diagnostic dir = %q, want %q", filepath.Dir(path), dir)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", path, err)
		}
		content := string(data)
		for _, want := range []string{
			"RAW_LLM_RESPONSE_MARKED:",
			"BROBOT_MARKED_PACKET:",
			"BROBOT_FILTER_PACKET:",
			"FILTERED_SPLIT_READY[0]:",
			"FILTERED_SPLIT_SENT[0]:",
			"{{playAnimationWI||happy}}... speech",
		} {
			if !strings.Contains(content, want) {
				t.Fatalf("%s missing %q:\n%s", filepath.Base(path), want, content)
			}
		}
	}
	if !strings.HasPrefix(filepath.Base(paths[0]), "logs_00e20100_20260622T123456.789Z") {
		t.Fatalf("logs file name = %q", filepath.Base(paths[0]))
	}
	if !strings.HasPrefix(filepath.Base(paths[1]), "all_logs_00e20100_20260622T123456.789Z") {
		t.Fatalf("all_logs file name = %q", filepath.Base(paths[1]))
	}
}

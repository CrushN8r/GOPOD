package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func robotDiagnosticsDir() string {
	return strings.TrimSpace(os.Getenv("SAVE_ROBOT_SAY_TEXT_DIR"))
}

func safeDiagnosticESN(esn string) string {
	safeESN := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, esn)
	if safeESN == "" {
		return "unknown"
	}
	return safeESN
}

func diagnosticStamp(timestamp time.Time) string {
	return timestamp.UTC().Format("20060102T150405.000Z")
}

func saveRobotSayTextDiagnostic(esn, rawStr, cleanedText, finalText string, timestamp time.Time) (string, error) {
	dir := robotDiagnosticsDir()
	if dir == "" {
		return "", nil
	}
	if err := gopodMkdirAll(dir, 0755); err != nil {
		return "", err
	}
	safeESN := safeDiagnosticESN(esn)
	stamp := diagnosticStamp(timestamp)
	path := filepath.Join(dir, "saytext_"+safeESN+"_"+stamp+".txt")
	body := strings.Join([]string{
		"timestamp=" + timestamp.UTC().Format(time.RFC3339Nano),
		"esn=" + esn,
		"",
		"rawStr passed into DoSayText:",
		rawStr,
		"",
		"cleaned text after existing cleaner:",
		cleanedText,
		"",
		"final text sent to robot.Conn.SayText:",
		finalText,
		"",
	}, "\n")
	if err := gopodWriteFile(path, []byte(body), 0644); err != nil {
		return "", err
	}
	return path, nil
}

func saveBrobotResponseDiagnostics(esn string, result gopodStreamResult, sentPackets []string, timestamp time.Time) ([]string, error) {
	dir := robotDiagnosticsDir()
	if dir == "" {
		return nil, nil
	}
	if err := gopodMkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	safeESN := safeDiagnosticESN(esn)
	stamp := diagnosticStamp(timestamp)
	readyLines := make([]string, 0, len(result.AudibleChunks)*4)
	for i, chunk := range result.AudibleChunks {
		readyLines = append(readyLines,
			"FILTERED_SPLIT_READY["+strconv.Itoa(i)+"]:",
			chunk,
		)
		if i < len(sentPackets) {
			readyLines = append(readyLines,
				"FILTERED_SPLIT_SENT["+strconv.Itoa(i)+"]:",
				sentPackets[i],
			)
		}
	}
	logBody := strings.Join(append([]string{
		"timestamp=" + timestamp.UTC().Format(time.RFC3339Nano),
		"esn=" + esn,
		"",
		"RAW_LLM_RESPONSE_MARKED:",
		result.Raw,
		"",
		"BROBOT_MARKED_PACKET:",
		result.MarkedPacket,
		"",
		"BROBOT_FILTER_PACKET:",
		result.FilteredPacket,
		"",
	}, readyLines...), "\n") + "\n"
	allBody := logBody + strings.Join([]string{
		"GOPOD_STREAM_MARKER_0",
		"GOPOD_STREAM_MARKER_1",
		"BROBOT_FINALIZE_REASON=marker_1",
		"BROBOT_RELEASE_READY=yes",
		"",
	}, "\n")
	logPath := filepath.Join(dir, "logs_"+safeESN+"_"+stamp+".txt")
	allPath := filepath.Join(dir, "all_logs_"+safeESN+"_"+stamp+".txt")
	if err := gopodWriteFile(logPath, []byte(logBody), 0644); err != nil {
		return nil, err
	}
	if err := gopodWriteFile(allPath, []byte(allBody), 0644); err != nil {
		return nil, err
	}
	return []string{logPath, allPath}, nil
}

func TestSaveRobotSayTextDiagnosticWritesEnvGatedChunkFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SAVE_ROBOT_SAY_TEXT_DIR", dir)

	path, err := saveRobotSayTextDiagnostic("00e20100", "**raw**", "raw", "raw", time.Date(2026, 6, 22, 12, 34, 56, 789000000, time.UTC))
	if err != nil {
		t.Fatalf("saveRobotSayTextDiagnostic error = %v", err)
	}
	if path == "" {
		t.Fatal("diagnostic path is empty")
	}
	if filepath.Dir(path) != dir {
		t.Fatalf("diagnostic dir = %q, want %q", filepath.Dir(path), dir)
	}
	if !strings.HasPrefix(filepath.Base(path), "saytext_00e20100_20260622T123456.789Z") {
		t.Fatalf("diagnostic file name = %q", filepath.Base(path))
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	content := string(data)
	for _, want := range []string{
		"timestamp=2026-06-22T12:34:56.789Z",
		"esn=00e20100",
		"rawStr passed into DoSayText:",
		"**raw**",
		"cleaned text after existing cleaner:",
		"raw",
		"final text sent to robot.Conn.SayText:",
		"raw",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("diagnostic file missing %q:\n%s", want, content)
		}
	}
}

func TestSaveRobotSayTextDiagnosticNoopsWhenUnset(t *testing.T) {
	t.Setenv("SAVE_ROBOT_SAY_TEXT_DIR", "")

	path, err := saveRobotSayTextDiagnostic("00e20100", "raw", "raw", "raw", time.Date(2026, 6, 22, 12, 34, 56, 0, time.UTC))
	if err != nil {
		t.Fatalf("saveRobotSayTextDiagnostic error = %v", err)
	}
	if path != "" {
		t.Fatalf("diagnostic path = %q, want empty", path)
	}
}
